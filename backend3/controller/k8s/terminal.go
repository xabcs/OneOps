package k8s

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/k8s"
	syssvc "oneops/backend3/service/system"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// TerminalController K8s Pod 终端 WebSocket 处理器
type TerminalController struct {
	svc            *K8sClusterService
	db             *gorm.DB
	upgrader       websocket.Upgrader
	activeSessions map[uint]*k8sSession
	sessionsMutex  sync.RWMutex
}

// NewTerminalController 创建 K8s 终端控制器
func NewTerminalController(svc *K8sClusterService) *TerminalController {
	return &TerminalController{
		svc: svc,
		db:  database.GetDB(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		activeSessions: make(map[uint]*k8sSession),
	}
}

// k8sSession K8s 终端会话（内部）
type k8sSession struct {
	sessionID     uint
	clusterID     uint
	namespace     string
	podName       string
	containerName string
	wsConn        *websocket.Conn
	k8sExecutor   remotecommand.Executor
	ctx           context.Context
	cancel        context.CancelFunc
	userID        uint
	lastActivity  time.Time
	mu            sync.Mutex
}

// K8sTerminalMessage 终端消息
type K8sTerminalMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
	Data string `json:"data"`
}

// k8sStreamSize 终端大小
type k8sStreamSize struct {
	Height uint16
	Width  uint16
}

// k8sTerminalSizeQueue 实现 TerminalSizeQueue 接口
type k8sTerminalSizeQueue struct {
	mu     sync.Mutex
	closed bool
	sizes  chan remotecommand.TerminalSize
}

func newK8sTerminalSizeQueue() *k8sTerminalSizeQueue {
	return &k8sTerminalSizeQueue{
		sizes: make(chan remotecommand.TerminalSize, 10),
	}
}

func (q *k8sTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.sizes
	if !ok {
		return nil
	}
	return &size
}

// TrySend 非阻塞推送终端尺寸；Stop 后静默丢弃（避免向已关闭 channel 发送 panic）
func (q *k8sTerminalSizeQueue) TrySend(size remotecommand.TerminalSize) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	select {
	case q.sizes <- size:
	default:
	}
}

func (q *k8sTerminalSizeQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	close(q.sizes)
}

// HandleWebSocket godoc
// @Summary      K8s Pod 终端 WebSocket
// @Description  通过 WebSocket 连接到指定 Pod 的容器终端，实现交互式 shell。认证通过 query 参数 token 完成，不经过 Auth 中间件。
// @Tags         K8s-终端
// @Produce      json
// @Param        token         query     string  true   "JWT token"
// @Param        clusterId     query     string  true   "集群 ID"
// @Param        namespace     query     string  true   "命名空间"
// @Param        podName       query     string  true   "Pod 名称"
// @Param        containerName query     string  false  "容器名称(留空则取第一个容器)"
// @Success      101  {string}  string  "升级为 WebSocket 连接"
// @Failure      200  {object}  utils.Response  "缺少参数 / token验证失败 / 无权访问"
// @Router       /k8s/terminal/ws [get]
func (ctrl *TerminalController) HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "缺少认证token"})
		return
	}

	userID, err := ctrl.validateToken(token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "token验证失败"})
		return
	}

	// 用户状态校验（H4）：WS 不走 Auth 中间件，此处自查
	var userStatus string
	if err := ctrl.db.Table("sys_users").Select("status").
		Where("id = ?", userID).Scan(&userStatus).Error; err != nil || userStatus != "active" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "用户已被禁用或不存在"})
		return
	}

	// 系统权限码校验（H8）：该端点在 Auth/权限中间件组外，必须自查 k8s.terminal.connect
	permSvc, err := syssvc.GetPermissionService()
	if err == nil {
		hasPerm, permErr := permSvc.HasPermission(userID, "k8s.terminal.connect")
		if permErr != nil || !hasPerm {
			c.JSON(http.StatusOK, gin.H{"code": 403, "success": false, "message": "无终端连接权限（k8s.terminal.connect）"})
			return
		}
	} else {
		// 权限服务不可用时 fail-closed
		c.JSON(http.StatusOK, gin.H{"code": 500, "success": false, "message": "权限服务不可用"})
		return
	}

	c.Set("user_id", userID)

	clusterIDStr := c.Query("clusterId")
	namespace := c.Query("namespace")
	podName := c.Query("podName")
	containerName := c.Query("containerName")

	if clusterIDStr == "" || namespace == "" || podName == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "缺少必要参数"})
		return
	}

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "无效的集群ID"})
		return
	}

	allowed, err := ctrl.svc.CheckClusterOperation(userID, uint(clusterID), "k8s.terminal.connect")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, gin.H{"code": 403, "success": false, "message": "无权执行该操作（需要集群角色操作集包含 k8s.terminal.connect）"})
		return
	}

	clientset, _, err := ctrl.svc.GetScopedClientWithConfig(uint(clusterID), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "success": false, "message": "获取集群连接失败"})
		return
	}

	ctx := context.Background()
	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{"code": 404, "success": false, "message": "Pod 不存在"})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 500, "success": false, "message": fmt.Sprintf("获取 Pod 失败: %v", err)})
		}
		return
	}

	if containerName == "" {
		if len(pod.Spec.Containers) > 0 {
			containerName = pod.Spec.Containers[0].Name
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "Pod 没有可连接的容器"})
			return
		}
	}

	wsConn, err := ctrl.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("升级 WebSocket 连接失败", zap.Error(err))
		return
	}
	defer func() {
		// 先发 Close 帧再关 TCP，使前端 onclose.wasClean=true（正常断开而非"连接异常关闭"）
		_ = wsConn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		wsConn.Close()
	}()

	sessionCtx, cancel := context.WithCancel(context.Background())

	session := &modelk8s.K8sSession{
		UserID:        userID,
		ClusterID:     uint(clusterID),
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		Status:        "active",
		StartedAt:     time.Now(),
	}
	if err := ctrl.db.Create(session).Error; err != nil {
		logger.Error("创建 K8s 会话记录失败", zap.Error(err))
		cancel()
		return
	}

	terminalSession := &k8sSession{
		sessionID:     session.ID,
		clusterID:     uint(clusterID),
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,
		wsConn:        wsConn,
		ctx:           sessionCtx,
		cancel:        cancel,
		userID:        userID,
		lastActivity:  time.Now(),
	}

	ctrl.sessionsMutex.Lock()
	ctrl.activeSessions[session.ID] = terminalSession
	ctrl.sessionsMutex.Unlock()

	defer func() {
		ctrl.sessionsMutex.Lock()
		delete(ctrl.activeSessions, session.ID)
		ctrl.sessionsMutex.Unlock()
	}()

	size := &k8sStreamSize{Width: 80, Height: 24}
	shellCmd, probeErr := ctrl.probeShellCommand(terminalSession)
	if probeErr != nil {
		logger.Error("探测容器 shell 失败", zap.Error(probeErr))
		ctrl.sendTerminalMessage(terminalSession.wsConn, "\r\n\x1b[31m✗ 终端初始化失败: "+probeErr.Error()+"\x1b[0m\r\n")
		ctrl.closeSession(session, "初始化失败: "+probeErr.Error())
		return
	}
	executor, err := ctrl.createExecutor(terminalSession, shellCmd)
	if err != nil {
		logger.Error("初始化 K8s executor 失败", zap.Error(err))
		ctrl.sendTerminalMessage(terminalSession.wsConn, "\r\n\x1b[31m✗ 终端初始化失败: "+err.Error()+"\x1b[0m\r\n")
		ctrl.closeSession(session, "初始化失败")
		return
	}
	terminalSession.k8sExecutor = executor

	logger.Info("K8s 终端会话已建立",
		zap.Uint("session_id", session.ID),
		zap.Uint("user_id", userID),
		zap.Uint("cluster_id", uint(clusterID)),
		zap.String("namespace", namespace),
		zap.String("pod", podName),
		zap.String("container", containerName))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ctrl.handleSession(terminalSession, session, size)
	}()

	wg.Wait()

	ctrl.closeSession(session, "会话结束")
}

// probeShellCommand 探测容器内可用的交互 shell：依次尝试 /bin/bash、/bin/sh，
// 用一次性非交互 exec（exit 0）验证二进制存在；均不可用时返回错误。
// 注意：必须用 scoped 凭据（H2）——与预检/正式 exec 同一身份，
// 否则只读绑定（无 pods/exec）的用户也会因平台凭据探测成功而获得 shell
func (ctrl *TerminalController) probeShellCommand(session *k8sSession) ([]string, error) {
	clientset, config, err := ctrl.svc.GetScopedClientWithConfig(session.clusterID, session.userID)
	if err != nil {
		return nil, fmt.Errorf("获取集群连接失败: %w", err)
	}

	var lastErr error
	for _, shell := range []string{"/bin/bash", "/bin/sh"} {
		req := clientset.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(session.podName).
			Namespace(session.namespace).
			SubResource("exec").
			VersionedParams(&v1.PodExecOptions{
				Container: session.containerName,
				Command:   []string{shell, "-c", "exit 0"},
				Stdout:    true,
				Stderr:    true,
			}, scheme.ParameterCodec)

		executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			lastErr = err
			continue
		}
		if err := executor.StreamWithContext(session.ctx, remotecommand.StreamOptions{
			Stdout: io.Discard,
			Stderr: io.Discard,
		}); err != nil {
			lastErr = err
			continue
		}
		if shell == "/bin/bash" {
			return []string{shell, "-l"}, nil
		}
		return []string{shell}, nil
	}
	return nil, fmt.Errorf("容器内无可用 shell（/bin/bash、/bin/sh 均执行失败）: %v", lastErr)
}

// sendTerminalMessage 向终端 WebSocket 推送一条终端内可见的文本消息
func (ctrl *TerminalController) sendTerminalMessage(ws *websocket.Conn, text string) {
	if ws == nil {
		return
	}
	_ = ws.WriteJSON(K8sTerminalMessage{Type: "output", Data: text})
}

// createExecutor 创建 K8s executor
// 注意：必须用 scoped 凭据（H2）——exec 是否放行由集群原生 RBAC（pods/exec）终判，
// 不能用平台管理员凭据，否则集群 RBAC 对 exec 的限制被架空
func (ctrl *TerminalController) createExecutor(session *k8sSession, command []string) (remotecommand.Executor, error) {
	clientset, config, err := ctrl.svc.GetScopedClientWithConfig(session.clusterID, session.userID)
	if err != nil {
		return nil, fmt.Errorf("获取集群连接失败: %w", err)
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(session.podName).
		Namespace(session.namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: session.containerName,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	return remotecommand.NewSPDYExecutor(config, "POST", req.URL())
}

// handleSession 处理会话
func (ctrl *TerminalController) handleSession(session *k8sSession, dbSession *modelk8s.K8sSession, size *k8sStreamSize) {
	streamer := &k8sTerminalStreamer{
		session:    session,
		dbSession:  dbSession,
		controller: ctrl,
		stdin:      make(chan []byte, 10),
		stdinDone:  make(chan struct{}),
		stdoutDone: make(chan struct{}),
		sizeQueue:  newK8sTerminalSizeQueue(),
	}

	streamer.sizeQueue.sizes <- remotecommand.TerminalSize{
		Width:  size.Width,
		Height: size.Height,
	}

	go streamer.handleStdout()
	go streamer.handleWebSocketMessages()

	err := session.k8sExecutor.StreamWithContext(session.ctx, remotecommand.StreamOptions{
		Stdin:             streamer,
		Stdout:            streamer,
		Stderr:            streamer,
		Tty:               true,
		TerminalSizeQueue: streamer.sizeQueue,
	})

	if err != nil {
		logger.Error("执行远程命令失败", zap.Error(err))
		// 透传断开原因到前端终端，避免用户只看到"连接异常关闭"
		ctrl.sendTerminalMessage(session.wsConn, "\r\n\x1b[31m✗ 终端已断开: "+err.Error()+"\x1b[0m\r\n")
	}

	close(streamer.stdinDone)
	<-streamer.stdoutDone
	streamer.sizeQueue.Stop()
}

// closeSession 关闭会话
func (ctrl *TerminalController) closeSession(session *modelk8s.K8sSession, reason string) {
	now := time.Now()
	session.Status = "closed"
	session.EndedAt = &now
	session.Duration = int(now.Sub(session.StartedAt).Seconds())
	session.CloseReason = reason

	if err := ctrl.db.Save(session).Error; err != nil {
		logger.Error("更新 K8s 会话状态失败", zap.Error(err))
	}

	logger.Info("K8s 终端会话已关闭",
		zap.Uint("session_id", session.ID),
		zap.String("reason", reason),
		zap.Int("duration", session.Duration))
}

// validateToken 验证token并返回用户ID
func (ctrl *TerminalController) validateToken(token string) (uint, error) {
	claims, err := utils.ParseToken(token)
	if err != nil {
		return 0, fmt.Errorf("token验证失败: %w", err)
	}

	return claims.UserID, nil
}

// GetActiveSessions godoc
// @Summary      获取活跃终端会话
// @Description  返回当前用户的活跃 K8s 终端会话列表
// @Tags         K8s-终端
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/terminal/active [get]
// @Security     BearerAuth
func (ctrl *TerminalController) GetActiveSessions(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	ctrl.sessionsMutex.RLock()
	defer ctrl.sessionsMutex.RUnlock()

	sessions := make([]map[string]interface{}, 0)
	for _, session := range ctrl.activeSessions {
		if session.userID == userID {
			sessions = append(sessions, map[string]interface{}{
				"session_id":    session.sessionID,
				"cluster_id":    session.clusterID,
				"namespace":     session.namespace,
				"pod_name":      session.podName,
				"container":     session.containerName,
				"last_activity": session.lastActivity.Format("2006-01-02 15:04:05"),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "success": true, "data": sessions})
}

// TerminateSession godoc
// @Summary      终止终端会话
// @Description  终止指定的 K8s 终端 WebSocket 会话
// @Tags         K8s-终端
// @Produce      json
// @Param        sessionId  path      int  true  "会话 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "无效的会话ID / 会话不存在 / 无权操作"
// @Router       /k8s/terminal/sessions/{sessionId}/terminate [post]
// @Security     BearerAuth
func (ctrl *TerminalController) TerminateSession(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	sessionID, err := strconv.ParseUint(c.Param("sessionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "无效的会话ID"})
		return
	}

	ctrl.sessionsMutex.RLock()
	session, exists := ctrl.activeSessions[uint(sessionID)]
	ctrl.sessionsMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": 404, "success": false, "message": "会话不存在或已关闭"})
		return
	}

	if session.userID != userID {
		c.JSON(http.StatusOK, gin.H{"code": 403, "success": false, "message": "无权操作该会话"})
		return
	}

	session.cancel()
	session.wsConn.Close()

	c.JSON(http.StatusOK, gin.H{"code": 200, "success": true, "message": "会话已终止"})
}

// k8sTerminalStreamer 实现 remotecommand.Streamer 接口
type k8sTerminalStreamer struct {
	session    *k8sSession
	dbSession  *modelk8s.K8sSession
	controller *TerminalController
	stdin      chan []byte
	stdinDone  chan struct{}
	stdoutDone chan struct{}
	sizeQueue  *k8sTerminalSizeQueue
}

// Write 实现 io.Writer 接口 (stdout/stderr)
func (s *k8sTerminalStreamer) Write(p []byte) (n int, err error) {
	msg := K8sTerminalMessage{
		Type: "output",
		Data: string(p),
	}
	if err := s.session.wsConn.WriteJSON(msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Read 实现 io.Reader 接口 (stdin)
func (s *k8sTerminalStreamer) Read(p []byte) (n int, err error) {
	select {
	case data, ok := <-s.stdin:
		if !ok {
			return 0, fmt.Errorf("stdin closed")
		}
		copy(p, data)
		return len(data), nil
	case <-s.session.ctx.Done():
		return 0, fmt.Errorf("context canceled")
	case <-s.stdinDone:
		return 0, fmt.Errorf("stdin done")
	}
}

// handleStdout 处理 stdout 输出
func (s *k8sTerminalStreamer) handleStdout() {
	defer close(s.stdoutDone)
	// stdout 由 Write 方法处理
}

// handleWebSocketMessages 处理 WebSocket 消息
func (s *k8sTerminalStreamer) handleWebSocketMessages() {
	for {
		select {
		case <-s.session.ctx.Done():
			return
		default:
			var msg K8sTerminalMessage
			err := s.session.wsConn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logger.Warn("WebSocket 读取错误", zap.Error(err))
				}
				return
			}

			s.session.mu.Lock()
			s.session.lastActivity = time.Now()
			s.session.mu.Unlock()

			switch msg.Type {
			case "resize":
				// 调整终端大小（非阻塞推送，SPDY 流实时生效）
				if msg.Cols > 0 && msg.Rows > 0 {
					s.sizeQueue.TrySend(remotecommand.TerminalSize{
						Width:  uint16(msg.Cols),
						Height: uint16(msg.Rows),
					})
				}
			case "close":
				s.session.cancel()
				return
			case "stdin":
				select {
				case s.stdin <- []byte(msg.Data):
				case <-s.session.ctx.Done():
					return
				}
			}
		}
	}
}
