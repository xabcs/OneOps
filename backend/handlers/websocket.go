package handlers

import (
	"fmt"
	"log"
	"net/http"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域，生产环境应该限制
		},
	}
)

// SSHWebSocketHandler SSH WebSocket 处理器
type SSHWebSocketHandler struct {
	bastionService *services.BastionService
	sessionManager *SessionManager
}

// NewSSHWebSocketHandler 创建 SSH WebSocket 处理器
func NewSSHWebSocketHandler() *SSHWebSocketHandler {
	return &SSHWebSocketHandler{
		bastionService: services.NewBastionService(),
		sessionManager: NewSessionManager(),
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *SSHWebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建一个 Gin 上下文
	// 注意：这里只是临时解决方案，实际上应该重构为直接使用 HTTP
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = r

	// 调用 HandleWebSocket
	h.HandleWebSocket(ctx)
}

// HandleWebSocket 处理 WebSocket 连接
func (h *SSHWebSocketHandler) HandleWebSocket(ctx *gin.Context) {
	// 获取会话ID
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	// 获取用户信息
	userID := ctx.GetUint("userId")
	if userID == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	// 获取会话信息
	session, err := h.bastionService.GetSessionByID(uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("会话不存在"))
		return
	}

	// 验证会话归属
	if session.UserID != userID {
		ctx.JSON(http.StatusOK, utils.ErrorForbidden("无权访问此会话"))
		return
	}

	// 检查会话状态
	if session.Status != "active" {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("会话已关闭"))
		return
	}

	// 升级到 WebSocket 连接
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 建立 SSH 连接
	sshClient, err := h.connectToServer(session)
	if err != nil {
		log.Printf("SSH 连接失败: %v", err)
		h.sendErrorMessage(conn, fmt.Sprintf("SSH 连接失败: %v", err))
		h.bastionService.CloseSession(session.ID, "SSH 连接失败")
		return
	}
	defer sshClient.Close()

	// 创建 SSH 会话
	sshSession, err := sshClient.NewSession()
	if err != nil {
		log.Printf("创建 SSH 会话失败: %v", err)
		h.sendErrorMessage(conn, fmt.Sprintf("创建 SSH 会话失败: %v", err))
		h.bastionService.CloseSession(session.ID, "创建 SSH 会话失败")
		return
	}
	defer sshSession.Close()

	// 设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 启用回显
		ssh.TTY_OP_ISPEED: 14400, // 输入速度 = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // 输出速度 = 14.4kbaud
	}

	// 设置伪终端
	if err := sshSession.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Printf("设置伪终端失败: %v", err)
		h.sendErrorMessage(conn, fmt.Sprintf("设置伪终端失败: %v", err))
		h.bastionService.CloseSession(session.ID, "设置伪终端失败")
		return
	}

	// 启动远程 shell
	if err := sshSession.Shell(); err != nil {
		log.Printf("启动 shell 失败: %v", err)
		h.sendErrorMessage(conn, fmt.Sprintf("启动 shell 失败: %v", err))
		h.bastionService.CloseSession(session.ID, "启动 shell 失败")
		return
	}

	// 注册会话到管理器
	h.sessionManager.Add(session.ID, conn, sshSession)
	defer h.sessionManager.Remove(session.ID)

	// 启动心跳检测
	stopHeartbeat := make(chan struct{})
	go h.heartbeat(session.ID, conn, stopHeartbeat)
	defer close(stopHeartbeat)

	// 记录连接成功
	log.Printf("SSH 会话 %d 已建立", session.ID)

	// 启动双向数据转发
	var wg sync.WaitGroup
	wg.Add(2)

	// WebSocket -> SSH
	go func() {
		defer wg.Done()
		h.forwardWebSocketToSSH(conn, sshSession, session)
	}()

	// SSH -> WebSocket
	go func() {
		defer wg.Done()
		h.forwardSSHToWebSocket(sshSession, conn, session)
	}()

	// 等待转发结束
	wg.Wait()

	// 关闭会话
	h.bastionService.CloseSession(session.ID, "用户断开连接")
	log.Printf("SSH 会话 %d 已关闭", session.ID)
}

// connectToServer 建立到服务器的 SSH 连接
func (h *SSHWebSocketHandler) connectToServer(session *models.BastionSession) (*ssh.Client, error) {
	// 获取服务器信息
	server := session.Server

	// 获取 SSH 凭证
	credential := session.SSHCredential
	if credential == nil {
		return nil, fmt.Errorf("未找到 SSH 凭证")
	}

	// 构建认证方法
	var authMethods []ssh.AuthMethod

	if credential.Password != "" {
		authMethods = append(authMethods, ssh.Password(credential.Password))
	}

	if credential.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(credential.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("没有可用的认证方法")
	}

	// 构建客户端配置
	config := &ssh.ClientConfig{
		User:            session.LoginAccount,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
		Timeout:         30 * time.Second,
	}

	// 连接到服务器
	address := fmt.Sprintf("%s:%d", server.IP, server.SSHPort)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	return client, nil
}

// forwardWebSocketToSSH 从 WebSocket 转发数据到 SSH
func (h *SSHWebSocketHandler) forwardWebSocketToSSH(conn *websocket.Conn, sshSession *ssh.Session, session *models.BastionSession) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket -> SSH 转发异常: %v", r)
		}
	}()

	// 获取 SSH 会话的 stdin 管道
	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		log.Printf("获取 stdin 管道失败: %v", err)
		return
	}

	// 记录命令缓冲区
	var commandBuffer []byte

	for {
		// 读取 WebSocket 消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			return
		}

		// 转发到 SSH stdin
		if _, err := stdinPipe.Write(message); err != nil {
			log.Printf("SSH 写入错误: %v", err)
			return
		}

		// 累积命令
		commandBuffer = append(commandBuffer, message...)

		// 检测命令结束（换行符）
		if len(message) > 0 && (message[len(message)-1] == '\n' || message[len(message)-1] == '\r') {
			// 记录命令（去掉控制字符）
			command := string(cleanCommand(commandBuffer))
			if command != "" {
				h.bastionService.RecordCommand(session.ID, command, 0, "")
			}
			commandBuffer = nil
		}
	}
}

// forwardSSHToWebSocket 从 SSH 转发数据到 WebSocket
func (h *SSHWebSocketHandler) forwardSSHToWebSocket(sshSession *ssh.Session, conn *websocket.Conn, session *models.BastionSession) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SSH -> WebSocket 转发异常: %v", r)
		}
	}()

	// 获取 SSH 会话的 stdout 管道
	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		log.Printf("获取 stdout 管道失败: %v", err)
		return
	}

	// 读取 SSH 输出
	output := make([]byte, 1024)
	for {
		n, err := stdoutPipe.Read(output)
		if err != nil {
			log.Printf("SSH 读取错误: %v", err)
			return
		}

		// 发送到 WebSocket
		if err := conn.WriteMessage(websocket.TextMessage, output[:n]); err != nil {
			log.Printf("WebSocket 写入错误: %v", err)
			return
		}
	}
}

// sendErrorMessage 发送错误消息
func (h *SSHWebSocketHandler) sendErrorMessage(conn *websocket.Conn, message string) {
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31m"+message+"\x1b[0m\r\n"))
}

// heartbeat 心跳检测
func (h *SSHWebSocketHandler) heartbeat(sessionID uint, conn *websocket.Conn, stop chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送 ping
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("心跳检测失败: %v", err)
				return
			}
		case <-stop:
			return
		}
	}
}

// cleanCommand 清理命令中的控制字符
func cleanCommand(cmd []byte) []byte {
	// 去除 ANSI 转义序列
	result := make([]byte, 0, len(cmd))
	inEscape := false

	for _, b := range cmd {
		if b == 0x1b { // ESC
			inEscape = true
			continue
		}
		if inEscape {
			if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' {
				inEscape = false
			}
			continue
		}
		// 保留可打印字符
		if b >= 32 && b <= 126 || b == '\n' || b == '\r' || b == '\t' {
			result = append(result, b)
		}
	}

	return result
}

// ResizePTY 调整终端大小
func (h *SSHWebSocketHandler) ResizePTY(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	// 获取新尺寸
	var req struct {
		Rows uint `json:"rows" binding:"required"`
		Cols uint `json:"cols" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	// 获取会话的 SSH session
	sshSession := h.sessionManager.GetSSHSession(uint(sessionID))
	if sshSession == nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("会话不存在"))
		return
	}

	// 发送窗口大小变化请求
	// RFC 4254 Section 6.7: pty-req request format
	type windowChangeMsg struct {
		Columns uint32
		Rows    uint32
		Width   uint32
		Height  uint32
	}

	msg := windowChangeMsg{
		Columns: uint32(req.Cols),
		Rows:    uint32(req.Rows),
		Width:   uint32(req.Cols * 8), // 像素宽度
		Height:  uint32(req.Rows * 16), // 像素高度
	}

	// 使用 SendRequest 发送窗口变化请求
	ok, err := sshSession.SendRequest("window-change", false, ssh.Marshal(&msg))
	if err != nil || !ok {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("调整终端大小失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("终端大小已调整"))
}

// SessionManager 会话管理器
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uint]*SessionState
}

// SessionState 会话状态
type SessionState struct {
	Conn       *websocket.Conn
	SSHSession *ssh.Session
	CreatedAt  time.Time
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[uint]*SessionState),
	}
}

// Add 添加会话
func (m *SessionManager) Add(sessionID uint, conn *websocket.Conn, sshSession *ssh.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[sessionID] = &SessionState{
		Conn:       conn,
		SSHSession: sshSession,
		CreatedAt:  time.Now(),
	}
}

// Remove 移除会话
func (m *SessionManager) Remove(sessionID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, ok := m.sessions[sessionID]; ok {
		if session.Conn != nil {
			session.Conn.Close()
		}
		if session.SSHSession != nil {
			session.SSHSession.Close()
		}
		delete(m.sessions, sessionID)
	}
}

// GetSSHSession 获取 SSH 会话
func (m *SessionManager) GetSSHSession(sessionID uint) *ssh.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if session, ok := m.sessions[sessionID]; ok {
		return session.SSHSession
	}
	return nil
}

// GetSession 获取会话状态
func (m *SessionManager) GetSession(sessionID uint) *SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

// GetActiveSessionCount 获取活跃会话数
func (m *SessionManager) GetActiveSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// TerminateSession 终止会话
func (m *SessionManager) TerminateSession(sessionID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, ok := m.sessions[sessionID]; ok {
		if session.Conn != nil {
			session.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "会话被强制终止"))
		}
		if session.SSHSession != nil {
			session.SSHSession.Close()
		}
		delete(m.sessions, sessionID)
	}
}
