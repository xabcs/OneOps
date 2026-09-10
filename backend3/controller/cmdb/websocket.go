package cmdb

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/utils"
	cmdbsvc "oneops/backend3/service/cmdb"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// ========== 线程安全的 WebSocket 连接 ==========
// gorilla/websocket 不支持并发写，所有写操作（WriteMessage / WriteControl）
// 必须通过 wsConn 的互斥锁串行化。

type wsConn struct {
	*websocket.Conn
	writeMu sync.Mutex
}

func (w *wsConn) safeWrite(messageType int, data []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	return w.Conn.WriteMessage(messageType, data)
}

func (w *wsConn) safeWriteControl(messageType int, data []byte, deadline time.Time) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	return w.Conn.WriteControl(messageType, data, deadline)
}

// ========== WebSocket Upgrader ==========

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
	HandshakeTimeout: 10 * time.Second,
}

// ========== SessionManager ==========

var globalSessionManager *SessionManager

// PTYSize PTY 窗口尺寸（像素按字符 8x16 估算）
type PTYSize struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

// SessionManager 管理 SSH WebSocket 会话
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uint]*SessionState
	// pendingResizes SSH 握手期间到达的 resize 暂存（sessionID → 尺寸），
	// 会话注册进 sessions 时补发 window-change，防止前端初始化尺寸丢失
	pendingResizes map[uint]PTYSize
	svc            *cmdbsvc.BastionService
}

// SessionState 会话状态
type SessionState struct {
	SessionID    uint
	Conn         *wsConn
	SSHSession   *ssh.Session
	UserID       uint
	Username     string
	ServerName   string
	ServerIP     string
	LoginAccount string
	Protocol     string
	ClientIP     string
	CreatedAt    time.Time
	LastActiveAt time.Time
}

// InitSessionManager 初始化全局 SessionManager（在路由注册时调用）
func InitSessionManager(svc *cmdbsvc.BastionService) {
	if globalSessionManager == nil {
		globalSessionManager = &SessionManager{
			sessions:       make(map[uint]*SessionState),
			pendingResizes: make(map[uint]PTYSize),
			svc:            svc,
		}
	}
}

// GetSessionManager 获取全局 SessionManager
func GetSessionManager() *SessionManager {
	return globalSessionManager
}

// SetPendingResize 暂存 SSH 握手期间到达的终端尺寸
func (m *SessionManager) SetPendingResize(sessionID uint, cols, rows uint) {
	m.mu.Lock()
	m.pendingResizes[sessionID] = PTYSize{
		Columns: uint32(cols),
		Rows:    uint32(rows),
		Width:   uint32(cols * 8),
		Height:  uint32(rows * 16),
	}
	m.mu.Unlock()
}

// applyPendingResize 会话注册后补发暂存的终端尺寸（若有）
func (m *SessionManager) applyPendingResize(sessionID uint, sshSession *ssh.Session) {
	m.mu.Lock()
	size, ok := m.pendingResizes[sessionID]
	if ok {
		delete(m.pendingResizes, sessionID)
	}
	m.mu.Unlock()
	if ok && sshSession != nil {
		_, _ = sshSession.SendRequest("window-change", false, ssh.Marshal(&size))
	}
}

// Add 添加会话
func (m *SessionManager) Add(sessionID uint, conn *wsConn, sshSession *ssh.Session, userID uint, serverName, serverIP, loginAccount, username, protocol, clientIP string) {
	m.mu.Lock()
	now := time.Now()
	m.sessions[sessionID] = &SessionState{
		SessionID:    sessionID,
		Conn:         conn,
		SSHSession:   sshSession,
		UserID:       userID,
		Username:     username,
		ServerName:   serverName,
		ServerIP:     serverIP,
		LoginAccount: loginAccount,
		Protocol:     protocol,
		ClientIP:     clientIP,
		CreatedAt:    now,
		LastActiveAt: now,
	}
	m.mu.Unlock()

	// 补发 SSH 握手期间暂存的终端尺寸（锁外发送避免死锁）
	m.applyPendingResize(sessionID, sshSession)
}

// Remove 移除会话并更新数据库状态
func (m *SessionManager) Remove(sessionID uint) {
	m.mu.Lock()
	_, exists := m.sessions[sessionID]
	if exists {
		delete(m.sessions, sessionID)
	}
	delete(m.pendingResizes, sessionID) // 清理暂存尺寸，防泄漏
	m.mu.Unlock()

	// 在锁外更新数据库（幂等，多次调用安全）
	if exists && m.svc != nil {
		if err := m.svc.CloseSession(sessionID, "用户断开连接"); err != nil {
			log.Printf("会话 %d 更新数据库失败: %v", sessionID, err)
		}
	}
}

// GetSSHSession 获取 SSH 会话
func (m *SessionManager) GetSSHSession(sessionID uint) *ssh.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.sessions[sessionID]; ok {
		return s.SSHSession
	}
	return nil
}

// GetSession 获取会话状态
func (m *SessionManager) GetSession(sessionID uint) *SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

// GetAllActiveSessions 获取所有活跃会话
func (m *SessionManager) GetAllActiveSessions() []*SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*SessionState, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}

// GetActiveSessionCount 获取活跃会话数
func (m *SessionManager) GetActiveSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// UpdateLastActiveAt 更新最后活动时间
func (m *SessionManager) UpdateLastActiveAt(sessionID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		s.LastActiveAt = time.Now()
	}
}

// TerminateSession 终止会话（关闭连接，数据库由 defer Remove 更新）
func (m *SessionManager) TerminateSession(sessionID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		if s.Conn != nil {
			s.Conn.safeWriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "会话被强制终止"), time.Now().Add(2*time.Second))
		}
		if s.SSHSession != nil {
			s.SSHSession.Close()
		}
		delete(m.sessions, sessionID)
	}
}

// ========== SSH WebSocket Handler ==========

// SSHWebSocketHandler SSH WebSocket 处理器
type SSHWebSocketHandler struct{}

// NewSSHWebSocketHandler 创建 SSH WebSocket 处理器
func NewSSHWebSocketHandler() *SSHWebSocketHandler {
	return &SSHWebSocketHandler{}
}

// HandleWebSocket godoc
// @Summary      SSH 终端 WebSocket
// @Description  通过 WebSocket 连接到堡垒机会话，实现交互式 SSH 终端。认证通过 query 参数 token 完成，不经过 Auth 中间件；会话需先经创建接口建立
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id    path     string  true  "会话 ID"
// @Param        token query    string  true  "JWT token"
// @Success      101   {string} string   "升级为 WebSocket 连接"
// @Failure      200   {object} utils.Response  "无效的会话ID / 未认证 / SessionManager 未初始化"
// @Router       /cmdb/sessions/{id}/ws [get]
func (h *SSHWebSocketHandler) HandleWebSocket(ctx *gin.Context) {
	// 获取会话 ID
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	// 从 token query 参数认证（WebSocket 无法发送 Authorization header）
	userID := ctx.GetUint("user_id")
	token := ctx.Query("token")
	if userID == 0 && token != "" {
		claims, err := utils.ParseToken(token)
		if err == nil {
			userID = claims.UserID
		}
	}
	if userID == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	// 用户状态校验（H4）：WS 不走 Auth 中间件，此处自查，禁用用户不能接入会话
	var userStatus string
	if err := database.GetDB().Table("sys_users").
		Select("status").Where("id = ?", userID).Scan(&userStatus).Error; err != nil || userStatus != "active" {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("用户已被禁用或不存在"))
		return
	}

	sm := GetSessionManager()
	if sm == nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("SessionManager 未初始化"))
		return
	}

	// 获取会话信息（预加载 Server 和 SSHCredential）
	session, err := sm.svc.GetSessionByID(uint(sessionID))
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

	// 升级到 WebSocket
	rawConn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}
	defer rawConn.Close()
	conn := &wsConn{Conn: rawConn}

	// 建立 SSH 连接
	sshClient, err := h.connectToServer(session)
	if err != nil {
		log.Printf("SSH 连接失败: %v", err)
		h.closeWithError(conn, fmt.Sprintf("SSH 连接失败: %v", err), websocket.CloseInternalServerErr)
		_ = sm.svc.CloseSession(session.ID, "SSH 连接失败", "error")
		return
	}
	defer sshClient.Close()

	// 创建 SSH 会话
	sshSession, err := sshClient.NewSession()
	if err != nil {
		log.Printf("创建 SSH 会话失败: %v", err)
		h.closeWithError(conn, fmt.Sprintf("创建 SSH 会话失败: %v", err), websocket.CloseInternalServerErr)
		_ = sm.svc.CloseSession(session.ID, "创建 SSH 会话失败", "error")
		return
	}
	defer sshSession.Close()

	// 必须在 Shell() 之前获取管道
	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		h.closeWithError(conn, fmt.Sprintf("获取 stdin 管道失败: %v", err), websocket.CloseInternalServerErr)
		_ = sm.svc.CloseSession(session.ID, "获取 stdin 管道失败", "error")
		return
	}
	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		h.closeWithError(conn, fmt.Sprintf("获取 stdout 管道失败: %v", err), websocket.CloseInternalServerErr)
		_ = sm.svc.CloseSession(session.ID, "获取 stdout 管道失败", "error")
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
		_ = sm.svc.CloseSession(session.ID, "设置伪终端失败", "error")
		return
	}

	// 启动远程 shell
	if err := sshSession.Shell(); err != nil {
		log.Printf("启动 shell 失败: %v", err)
		h.closeWithError(conn, fmt.Sprintf("启动 shell 失败: %v", err), websocket.CloseInternalServerErr)
		_ = sm.svc.CloseSession(session.ID, "启动 shell 失败", "error")
		return
	}

	// 注册会话到管理器
	serverName, serverIP := "", ""
	if session.Server != nil {
		serverName = session.Server.Hostname
		serverIP = session.Server.IP
	}
	sm.Add(session.ID, conn, sshSession, userID, serverName, serverIP, session.LoginAccount, session.Username, session.Protocol, session.ClientIP)
	defer sm.Remove(session.ID)

	// 创建可取消的 context
	sshCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置 Pong 超时处理
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 心跳检测
	stopHeartbeat := make(chan struct{})
	go h.heartbeat(session.ID, conn, stopHeartbeat)
	defer close(stopHeartbeat)

	log.Printf("SSH 会话 %d 已建立", session.ID)

	// 双向数据转发
	var wg sync.WaitGroup
	wg.Add(2)

	// WebSocket -> SSH
	go func() {
		defer wg.Done()
		h.forwardWebSocketToSSH(conn, stdinPipe, session, sshCtx)
		// WebSocket->SSH 退出后关闭 SSH 会话，触发 SSH->WebSocket 退出
		if sshSession != nil {
			sshSession.Close()
		}
		cancel()
	}()

	// SSH -> WebSocket
	go func() {
		defer wg.Done()
		h.forwardSSHToWebSocket(stdoutPipe, conn, session, sshCtx)
	}()

	wg.Wait()
	log.Printf("SSH 会话 %d 已关闭", session.ID)
}

// connectToServer 建立到服务器的 SSH 连接（不修改 model 中的敏感字段）
func (h *SSHWebSocketHandler) connectToServer(session *modelcmdb.BastionSession) (*ssh.Client, error) {
	server := session.Server
	if server == nil {
		return nil, fmt.Errorf("服务器信息不存在")
	}

	credential := session.SSHCredential
	if credential == nil {
		return nil, fmt.Errorf("未找到 SSH 凭证")
	}

	// 解密凭证到局部变量（不修改 model 指针，避免明文密码驻留在 model 中）。
	// 解密失败必须立即中止连接：继续用密文/坏密钥认证只会导致下游诡异失败
	password := credential.Password
	if password != "" {
		decrypted, err := utils.DecryptString(password)
		if err != nil {
			return nil, fmt.Errorf("凭证密码解密失败: %w", err)
		}
		password = decrypted
	}

	privateKey := credential.PrivateKey
	if privateKey != "" {
		decrypted, err := utils.DecryptString(privateKey)
		if err != nil {
			return nil, fmt.Errorf("凭证私钥解密失败: %w", err)
		}
		privateKey = decrypted
	}

	// 确定连接端口
	port := server.SSHPort
	if port == 0 {
		port = 22
	}

	// 构建认证方法
	var authMethods []ssh.AuthMethod

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("没有可用的认证方法")
	}

	config := &ssh.ClientConfig{
		User:    session.LoginAccount,
		Auth:    authMethods,
		Timeout: 30 * time.Second,
		// TODO 安全风险：跳过主机指纹校验，后续应接入 known_hosts
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", server.IP, port)
	log.Printf("[SSH CONNECT] 连接 %s (用户: %s)", address, session.LoginAccount)

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	return client, nil
}

// forwardWebSocketToSSH 从 WebSocket 转发数据到 SSH
func (h *SSHWebSocketHandler) forwardWebSocketToSSH(conn *wsConn, stdinPipe io.WriteCloser, session *modelcmdb.BastionSession, ctx context.Context) {
	defer func() {
		stdinPipe.Close()
		if r := recover(); r != nil {
			log.Printf("WebSocket -> SSH 转发异常: %v", r)
		}
	}()

	sm := GetSessionManager()
	var commandBuffer []byte

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			return
		}

		if sm != nil {
			sm.UpdateLastActiveAt(session.ID)
		}

		// 命令拦截（H7）：消息中任何位置出现回车/换行即视为命令提交，逐行送检。
		// 历史实现仅检查"消息末字节是回车"，帧内回车（原生 WS 客户端单帧 "rm -rf /\rls"）
		// 与整块粘贴（bracketed paste，末字节非回车）会被整帧透传直接执行，属授权绕过，已修复。
		// 送检文本经行编辑回放（Ctrl-C/Ctrl-U/退格/ANSI 剥离）后与远端实际行内容一致，
		// 避免"输入一半又删除"造成的误拦/漏拦错位
		combined := append(commandBuffer, message...)
		lines, pending, hasSubmit := utils.ExtractSubmittedLines(combined)
		if hasSubmit {
			commandBuffer = pending

			blockedCmd := ""
			if sm != nil {
				for _, raw := range lines {
					cmd := strings.TrimSpace(string(raw))
					if cmd == "" {
						continue
					}
					if sm.svc.IsCommandBlocked(cmd, session.ID) {
						blockedCmd = cmd
						break
					}
				}
			}

			if blockedCmd != "" {
				// 整帧拦截不透传（安全优先）：远端未收到本帧任何字节，
				// 本地缓冲必须同步清空，否则残留远端不存在的字符导致后续失同步
				commandBuffer = nil
				_, _ = stdinPipe.Write([]byte("\x15")) // Ctrl-U：清除远端未执行行
				_ = conn.safeWrite(websocket.TextMessage,
					[]byte("\r\n\x1b[31m[OneOps] 命令已被安全策略拦截，未执行: "+blockedCmd+"\x1b[0m\r\n"))
				log.Printf("[SSH 拦截] 会话 %d 用户 %d 命令被拦截: %s", session.ID, session.UserID, blockedCmd)
				_ = sm.svc.RecordCommand(session.ID, blockedCmd, -1, "[已拦截，未执行]", true)
				continue
			}

			if _, err := stdinPipe.Write(message); err != nil {
				log.Printf("SSH 写入错误: %v", err)
				return
			}
			if sm != nil {
				for _, raw := range lines {
					cmd := strings.TrimSpace(string(raw))
					if cmd != "" {
						_ = sm.svc.RecordCommand(session.ID, cmd, 0, "", false)
					}
				}
			}
			continue
		}

		if _, err := stdinPipe.Write(message); err != nil {
			log.Printf("SSH 写入错误: %v", err)
			return
		}
		// pending 已经过行编辑回放，直接作为本地行缓冲
		commandBuffer = pending
	}
}

// forwardSSHToWebSocket 从 SSH 转发数据到 WebSocket
func (h *SSHWebSocketHandler) forwardSSHToWebSocket(stdoutPipe io.Reader, conn *wsConn, session *modelcmdb.BastionSession, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SSH -> WebSocket 转发异常: %v", r)
		}
	}()

	output := make([]byte, 32*1024)

	type readResult struct {
		n   int
		err error
	}

	for {
		resultChan := make(chan readResult, 1)
		go func() {
			n, err := stdoutPipe.Read(output)
			resultChan <- readResult{n: n, err: err}
		}()

		select {
		case <-ctx.Done():
			return
		case result := <-resultChan:
			if result.n > 0 {
				if err := conn.safeWrite(websocket.TextMessage, output[:result.n]); err != nil {
					log.Printf("WebSocket 写入错误: %v", err)
					return
				}
			}
			if result.err != nil {
				log.Printf("SSH 读取结束: %v", result.err)
				return
			}
		}
	}
}

// closeWithError 发送错误消息并关闭 WebSocket
func (h *SSHWebSocketHandler) closeWithError(conn *wsConn, msg string, closeCode int) {
	conn.safeWrite(websocket.TextMessage, []byte("\r\n\x1b[31m"+msg+"\x1b[0m\r\n"))
	closeMsg := websocket.FormatCloseMessage(closeCode, msg)
	conn.safeWriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(2*time.Second))
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// heartbeat 心跳检测
func (h *SSHWebSocketHandler) heartbeat(sessionID uint, conn *wsConn, stop chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.safeWrite(websocket.PingMessage, nil); err != nil {
				log.Printf("心跳检测失败: %v", err)
				return
			}
		case <-stop:
			return
		}
	}
}
