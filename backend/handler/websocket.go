package handler

import (
	"context"
	"fmt"
	"io"
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
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域，生产环境应该限制
		},
		// 设置握手超时
		HandshakeTimeout: 10 * time.Second,
	}

	// 全局 SessionManager 单例
	globalSessionManager *SessionManager
)

// GetSessionManager 获取全局 SessionManager 实例
func GetSessionManager() *SessionManager {
	return globalSessionManager
}

// InitSessionManager 初始化 SessionManager（在服务启动时调用）
func InitSessionManager() {
	if globalSessionManager == nil {
		globalSessionManager = NewSessionManager(services.NewBastionService())
	}
}

// SSHWebSocketHandler SSH WebSocket 处理器
type SSHWebSocketHandler struct {
	bastionService *services.BastionService
	sessionManager *SessionManager
}

// NewSSHWebSocketHandler 创建 SSH WebSocket 处理器
func NewSSHWebSocketHandler() *SSHWebSocketHandler {
	return &SSHWebSocketHandler{
		bastionService: services.NewBastionService(),
		sessionManager: globalSessionManager,
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
	// 添加调试日志
	println("[DEBUG] HandleWebSocket called")
	println("[DEBUG] Path:", ctx.Request.URL.Path)
	println("[DEBUG] Query:", ctx.Request.URL.RawQuery)
	println("[DEBUG] Method:", ctx.Request.Method)

	// 获取会话ID
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		println("[DEBUG] 无效的会话ID:", sessionIDStr)
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	println("[DEBUG] SessionID:", sessionID)

	// 获取用户信息 - 支持两种方式：header或query参数
	userID := ctx.GetUint("user_id")
	token := ctx.Query("token")

	if userID == 0 && token != "" {
		// 从token解析用户ID
		claims, err := utils.ParseToken(token)
		if err == nil {
			userID = claims.UserID
			println("[DEBUG] 从token解析userID:", userID)
		}
	}

	if userID == 0 {
		println("[DEBUG] userID为0，未认证")
		// 对于WebSocket，返回401但不升级为WebSocket
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	println("[DEBUG] UserID:", userID)

	// 获取会话信息 - 需要预加载关联数据
	session, err := h.bastionService.GetSessionByID(uint(sessionID))
	if err != nil {
		println("[DEBUG] 获取会话失败:", err)
		ctx.JSON(http.StatusOK, utils.ErrorInternal("会话不存在"))
		return
	}

	println("[DEBUG] Session loaded:", session.ID, "ServerID:", session.ServerID, "Status:", session.Status)
	if session.Server != nil {
		println("[DEBUG] Server IP:", session.Server.IP, "Port:", session.Server.SSHPort)
	} else {
		println("[DEBUG] Server is nil!")
	}
	if session.SSHCredential != nil {
		println("[DEBUG] SSH Credential found:", session.SSHCredential.Name)
	} else {
		println("[DEBUG] SSH Credential is nil!")
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
		h.closeWithError(conn, fmt.Sprintf("SSH 连接失败: %v", err), websocket.CloseInternalServerErr)
		h.bastionService.CloseSession(session.ID, "SSH 连接失败")
		return
	}
	defer sshClient.Close()

	// 创建 SSH 会话
	sshSession, err := sshClient.NewSession()
	if err != nil {
		log.Printf("创建 SSH 会话失败: %v", err)
		h.closeWithError(conn, fmt.Sprintf("创建 SSH 会话失败: %v", err), websocket.CloseInternalServerErr)
		h.bastionService.CloseSession(session.ID, "创建 SSH 会话失败")
		return
	}
	defer sshSession.Close()

	// 必须在 Shell() 之前获取 stdin/stdout pipe，否则 ssh 包会报 "already set"
	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		h.closeWithError(conn, fmt.Sprintf("获取 stdin 管道失败: %v", err), websocket.CloseInternalServerErr)
		h.bastionService.CloseSession(session.ID, "获取 stdin 管道失败")
		return
	}
	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		h.closeWithError(conn, fmt.Sprintf("获取 stdout 管道失败: %v", err), websocket.CloseInternalServerErr)
		h.bastionService.CloseSession(session.ID, "获取 stdout 管道失败")
		return
	}

	// 设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// 设置伪终端
	if err := sshSession.RequestPty("xterm-256color", 40, 80, modes); err != nil {
		log.Printf("设置伪终端失败: %v", err)
		h.closeWithError(conn, fmt.Sprintf("设置伪终端失败: %v", err), websocket.CloseInternalServerErr)
		h.bastionService.CloseSession(session.ID, "设置伪终端失败")
		return
	}

	// 启动远程 shell
	if err := sshSession.Shell(); err != nil {
		log.Printf("启动 shell 失败: %v", err)
		h.closeWithError(conn, fmt.Sprintf("启动 shell 失败: %v", err), websocket.CloseInternalServerErr)
		h.bastionService.CloseSession(session.ID, "启动 shell 失败")
		return
	}

	// 注册会话到管理器
	log.Printf("[DEBUG] 注册会话到管理器，会话ID: %d", session.ID)
	h.sessionManager.Add(session.ID, conn, sshSession)
	log.Printf("[DEBUG] 设置 defer Remove，会话ID: %d", session.ID)
	defer func() {
		log.Printf("[DEBUG] defer Remove 开始执行，会话ID: %d", session.ID)
		h.sessionManager.Remove(session.ID)
		log.Printf("[DEBUG] defer Remove 执行完成，会话ID: %d", session.ID)
	}()

	// 创建可取消的 context，用于控制会话生命周期
	sshCtx, cancel := context.WithCancel(context.Background())
	defer func() {
		log.Printf("[DEBUG] 取消 context，会话ID: %d", session.ID)
		cancel()  // 确保在函数退出时取消 context
	}()

	// 设置 Pong 超时处理（如果60秒没有收到 Pong 响应，认为连接断开）
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动心跳检测
	stopHeartbeat := make(chan struct{})
	go h.heartbeat(session.ID, conn, stopHeartbeat)
	defer close(stopHeartbeat)

	// 记录连接成功
	log.Printf("SSH 会话 %d 已建立", session.ID)

	// 启动双向数据转发（pipe 已在 Shell() 前获取，直接传入）
	var wg sync.WaitGroup
	wg.Add(2)

	log.Printf("[DEBUG] 即将启动双向转发 goroutines，会话ID: %d", session.ID)
	// WebSocket -> SSH
	go func() {
		log.Printf("[DEBUG] WebSocket->SSH goroutine 启动，会话ID: %d", session.ID)
		defer wg.Done()
		h.forwardWebSocketToSSH(conn, stdinPipe, session, sshCtx)
		log.Printf("[DEBUG] WebSocket->SSH goroutine 退出，会话ID: %d", session.ID)

		// WebSocket->SSH 退出后：
		// 1. 立即关闭 SSH 会话
		// 2. 取消 context 通知另一个 goroutine
		log.Printf("[DEBUG] 关闭 SSH 会话，会话ID: %d", session.ID)
		if sshSession != nil {
			sshSession.Close()
		}
		log.Printf("[DEBUG] 取消 context，会话ID: %d", session.ID)
		cancel()
	}()

	// SSH -> WebSocket
	go func() {
		log.Printf("[DEBUG] SSH->WebSocket goroutine 启动，会话ID: %d", session.ID)
		defer wg.Done()
		h.forwardSSHToWebSocket(stdoutPipe, conn, session, sshCtx)
		log.Printf("[DEBUG] SSH->WebSocket goroutine 退出，会话ID: %d", session.ID)
	}()

	// 等待转发结束
	log.Printf("[DEBUG] 等待转发结束，会话ID: %d", session.ID)
	wg.Wait()
	log.Printf("[DEBUG] 转发已结束，会话ID: %d", session.ID)

	// 记录会话关闭（defer sessionManager.Remove 会处理数据库更新）
	log.Printf("SSH 会话 %d 已关闭", session.ID)
}

// connectToServer 建立到服务器的 SSH 连接
func (h *SSHWebSocketHandler) connectToServer(session *models.BastionSession) (*ssh.Client, error) {
	// 获取服务器信息
	server := session.Server
	log.Printf("[SSH CONNECT DEBUG] Connecting to %s (SSHPort=%d)", server.IP, server.SSHPort)

	// 获取 SSH 凭证
	credential := session.SSHCredential
	if credential == nil {
		return nil, fmt.Errorf("未找到 SSH 凭证")
	}
	log.Printf("[SSH CONNECT] Using credential: %s (credential username: %s)", credential.Name, credential.Username)
	log.Printf("[SSH CONNECT] Session LoginAccount: %s", session.LoginAccount)

	// 解密密码（如果加密存储）
	if credential.Password != "" {
		decryptedPwd, err := utils.DecryptString(credential.Password)
		if err != nil {
			log.Printf("[SSH CONNECT] Failed to decrypt password: %v", err)
			// 如果解密失败，尝试使用原值（可能是旧数据）
			log.Printf("[SSH CONNECT] Using original password (maybe old data)")
		} else {
			credential.Password = decryptedPwd
			log.Printf("[SSH CONNECT] Password decrypted successfully, length: %d", len(decryptedPwd))
		}
	}

	// 解密私钥（如果加密存储）
	if credential.PrivateKey != "" {
		decryptedKey, err := utils.DecryptString(credential.PrivateKey)
		if err != nil {
			log.Printf("[SSH CONNECT] Failed to decrypt private key: %v", err)
		} else {
			credential.PrivateKey = decryptedKey
			log.Printf("[SSH CONNECT] Private key decrypted successfully, length: %d", len(decryptedKey))
		}
	}

	// 确定连接端口
	port := server.SSHPort
	if port == 0 {
		port = 22
		log.Printf("[SSH CONNECT WARNING] SSHPort is 0, using default port 22")
	}
	log.Printf("[SSH CONNECT] Final address: %s:%d (original SSHPort: %d)", server.IP, port, server.SSHPort)

	// 构建认证方法
	var authMethods []ssh.AuthMethod

	log.Printf("[SSH CONNECT] Credential check - Password: %t, PrivateKey length: %d",
		credential.Password != "", len(credential.PrivateKey))

	if credential.Password != "" {
		authMethods = append(authMethods, ssh.Password(credential.Password))
		log.Printf("[SSH CONNECT] Added password auth method")
	}

	if credential.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(credential.PrivateKey))
		if err != nil {
			log.Printf("[SSH CONNECT] Failed to parse private key: %v", err)
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
		log.Printf("[SSH CONNECT] Added publickey auth method")
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("没有可用的认证方法")
	}
	log.Printf("[SSH CONNECT] Total auth methods: %d", len(authMethods))

	// 构建客户端配置
	config := &ssh.ClientConfig{
		User:            session.LoginAccount,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
		Timeout:         30 * time.Second,
	}
	log.Printf("[SSH CONNECT] SSH Client Config - User: %s, AuthMethods: %d", session.LoginAccount, len(authMethods))

	// 连接到服务器
	address := fmt.Sprintf("%s:%d", server.IP, port)
	log.Printf("[SSH CONNECT] Attempting to connect to: %s (user: %s)", address, session.LoginAccount)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	return client, nil
}

// forwardWebSocketToSSH 从 WebSocket 转发数据到 SSH
func (h *SSHWebSocketHandler) forwardWebSocketToSSH(conn *websocket.Conn, stdinPipe io.WriteCloser, session *models.BastionSession, ctx context.Context) {
	log.Printf("[DEBUG] forwardWebSocketToSSH 开始执行，会话ID: %d", session.ID)
	// 关键：函数退出时关闭 stdinPipe，向远端 shell 发送 EOF
	// 这样 SSH 会话会结束，forwardSSHToWebSocket 的 Read 会返回 error，wg.Wait() 才能解除阻塞
	defer func() {
		log.Printf("[DEBUG] forwardWebSocketToSSH 即将关闭 stdinPipe，会话ID: %d", session.ID)
		stdinPipe.Close()
		log.Printf("[DEBUG] forwardWebSocketToSSH 已关闭 stdinPipe，会话ID: %d", session.ID)
	}()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket -> SSH 转发异常: %v", r)
		}
	}()

	// 记录命令缓冲区
	var commandBuffer []byte

	for {
		log.Printf("[DEBUG] forwardWebSocketToSSH 等待读取 WebSocket 消息，会话ID: %d", session.ID)

		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			log.Printf("[DEBUG] forwardWebSocketToSSH context 已取消，退出，会话ID: %d", session.ID)
			return
		default:
			// 继续执行
		}

		// 读取 WebSocket 消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[DEBUG] forwardWebSocketToSSH WebSocket 读取错误，会话ID: %d，错误: %v", session.ID, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			log.Printf("[DEBUG] forwardWebSocketToSSH 即将 return，会话ID: %d", session.ID)
			return
		}

		// 更新最后活动时间（每次收到用户输入）
		h.sessionManager.UpdateLastActiveAt(session.ID)

		// 转发到 SSH stdin
		if _, err := stdinPipe.Write(message); err != nil {
			log.Printf("SSH 写入错误: %v", err)
			return
		}

		// 累积命令
		commandBuffer = append(commandBuffer, message...)

		// 检测命令结束（换行符）
		if len(message) > 0 && (message[len(message)-1] == '\n' || message[len(message)-1] == '\r') {
			command := string(cleanCommand(commandBuffer))
			if command != "" {
				h.bastionService.RecordCommand(session.ID, command, 0, "")
			}
			commandBuffer = nil
		}
	}
}

// forwardSSHToWebSocket 从 SSH 转发数据到 WebSocket
func (h *SSHWebSocketHandler) forwardSSHToWebSocket(stdoutPipe io.Reader, conn *websocket.Conn, session *models.BastionSession, ctx context.Context) {
	log.Printf("[DEBUG] forwardSSHToWebSocket 开始执行，会话ID: %d", session.ID)
	defer func() {
		log.Printf("[DEBUG] forwardSSHToWebSocket 即将退出，会话ID: %d", session.ID)
		if r := recover(); r != nil {
			log.Printf("SSH -> WebSocket 转发异常: %v", r)
		}
	}()

	// 读取 SSH 输出并转发到 WebSocket
	output := make([]byte, 32*1024)

	// 使用 goroutine + channel 模式，同时监听 context 和数据
	type readResult struct {
		n   int
		err error
	}

	for {
		log.Printf("[DEBUG] forwardSSHToWebSocket 等待读取 SSH 输出，会话ID: %d", session.ID)

		resultChan := make(chan readResult, 1)
		go func() {
			n, err := stdoutPipe.Read(output)
			resultChan <- readResult{n: n, err: err}
		}()

		select {
		case <-ctx.Done():
			log.Printf("[DEBUG] forwardSSHToWebSocket context 已取消，停止读取，会话ID: %d", session.ID)
			// context 取消，立即返回（SSH 会话已被另一个 goroutine 关闭）
			return
		case result := <-resultChan:
			// 读取完成，处理数据
			n := result.n
			err := result.err
			log.Printf("[DEBUG] forwardSSHToWebSocket 读取到 %d 字节，错误: %v，会话ID: %d", n, err, session.ID)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.TextMessage, output[:n]); werr != nil {
					log.Printf("WebSocket 写入错误: %v", werr)
					return
				}
			}
			if err != nil {
				log.Printf("SSH 读取结束: %v", err)
				return
			}
		}
	}
}

// closeWithError 发送错误消息并等待 WS 关闭握手完成，确保客户端能收到 close 帧
func (h *SSHWebSocketHandler) closeWithError(conn *websocket.Conn, msg string, closeCode int) {
	// 1. 先把错误消息写到终端
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31m"+msg+"\x1b[0m\r\n"))
	// 2. 发送 close 帧
	closeMsg := websocket.FormatCloseMessage(closeCode, msg)
	conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(2*time.Second))
	// 3. 等待客户端回复 close 帧（最多 2 秒），否则 defer conn.Close() 直接关 TCP 会导致浏览器看到 1006
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
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

// SessionManager 会话管理器（导出类型，供其他包使用）
type SessionManager struct {
	mu             sync.RWMutex
	sessions       map[uint]*SessionState
	bastionService *services.BastionService
}

// SessionState 会话状态
type SessionState struct {
	Conn         *websocket.Conn
	SSHSession   *ssh.Session
	CreatedAt    time.Time
	LastActiveAt time.Time // 最后活动时间（用于审计和展示）
}

// NewSessionManager 创建会话管理器
func NewSessionManager(bastionService *services.BastionService) *SessionManager {
	return &SessionManager{
		sessions:       make(map[uint]*SessionState),
		bastionService: bastionService,
	}
}

// Add 添加会话
func (m *SessionManager) Add(sessionID uint, conn *websocket.Conn, sshSession *ssh.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.sessions[sessionID] = &SessionState{
		Conn:         conn,
		SSHSession:   sshSession,
		CreatedAt:    now,
		LastActiveAt: now,
	}
}

// Remove 移除会话（总是更新数据库状态以防止僵尸会话）
func (m *SessionManager) Remove(sessionID uint) {
	m.mu.Lock()
	_, exists := m.sessions[sessionID]
	if exists {
		delete(m.sessions, sessionID)
		log.Printf("会话 %d 从内存中移除", sessionID)
	}
	m.mu.Unlock()

	// 在锁外调用 CloseSession，避免死锁
	// CloseSession 是幂等的，多次调用安全
	// 这确保即使因错误提前退出，数据库状态也会被更新
	if exists {
		log.Printf("会话 %d 正在更新数据库状态", sessionID)
		if err := m.bastionService.CloseSession(sessionID, "用户断开连接"); err != nil {
			log.Printf("会话 %d 更新数据库失败: %v", sessionID, err)
		} else {
			log.Printf("会话 %d 数据库状态已更新为 closed", sessionID)
		}
	} else {
		log.Printf("会话 %d 不在内存中（可能已被移除）", sessionID)
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

// GetAllActiveSessions 获取所有活跃会话的 ID 列表
func (m *SessionManager) GetAllActiveSessions() []uint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionIDs := make([]uint, 0, len(m.sessions))
	for sessionID := range m.sessions {
		sessionIDs = append(sessionIDs, sessionID)
	}
	return sessionIDs
}

// UpdateLastActiveAt 更新会话的最后活动时间
func (m *SessionManager) UpdateLastActiveAt(sessionID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[sessionID]; ok {
		session.LastActiveAt = time.Now()
	}
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
