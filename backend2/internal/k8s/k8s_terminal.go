package k8s

import (
	"context"
	"fmt"
	"net/http"
	"oneops/backend2/pkg/logger"
	"oneops/backend2/pkg/utils"
	"strconv"
	"sync"
	"time"

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

// K8sTerminalHandler K8s Pod 终端 WebSocket 处理器
type K8sTerminalHandler struct {
	clusterService *K8sClusterService
	clientPool     *K8sClientPool
	db             *gorm.DB
	upgrader       websocket.Upgrader
	activeSessions map[uint]*k8sTerminalSession
	sessionsMutex  sync.RWMutex
}

// K8sSession K8s 终端会话
type K8sSession struct {
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

// NewK8sTerminalHandler 创建 K8s 终端处理器
func NewK8sTerminalHandler(clusterSvc *K8sClusterService, clientPool *K8sClientPool, db *gorm.DB) *K8sTerminalHandler {
	return &K8sTerminalHandler{
		clusterService: clusterSvc,
		clientPool:     clientPool,
		db:             db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		activeSessions: make(map[uint]*K8sSession),
	}
}

// K8sTerminalMessage 终端消息
type K8sTerminalMessage struct {
	Type string `json:"type"` // resize, close
	Cols int    `json:"cols"` // 终端列数
	Rows int    `json:"rows"` // 终端行数
	Data string `json:"data"` // 数据内容（用于 stdin）
}

// K8sStreamSize 终端大小
type K8sStreamSize struct {
	Height uint16
	Width  uint16
}

// K8sTerminalSizeQueue 实现 TerminalSizeQueue 接口
type K8sTerminalSizeQueue struct {
	sizes chan remotecommand.TerminalSize
}

func NewK8sTerminalSizeQueue() *K8sTerminalSizeQueue {
	return &K8sTerminalSizeQueue{
		sizes: make(chan remotecommand.TerminalSize, 10),
	}
}

func (q *K8sTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-q.sizes
	if !ok {
		return nil
	}
	return &size
}

func (q *K8sTerminalSizeQueue) Stop() {
	close(q.sizes)
}

// HandleWebSocket 处理 WebSocket 连接
func (h *K8sTerminalHandler) HandleWebSocket(c *gin.Context) {
	// 从查询参数获取 token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "缺少认证token"})
		return
	}

	// 验证 token 并获取用户ID
	userID, err := h.validateToken(token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "token验证失败"})
		return
	}

	// 将用户ID设置到context中
	c.Set("user_id", userID)

	// 解析参数
	clusterIDStr := c.Query("clusterId")
	namespace := c.Query("namespace")
	podName := c.Query("podName")
	containerName := c.Query("containerName")

	// 参数验证
	if clusterIDStr == "" || namespace == "" || podName == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "缺少必要参数"})
		return
	}

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "无效的集群ID"})
		return
	}

	// 检查集群访问权限
	hasAccess, err := h.clusterService.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, gin.H{"code": 403, "success": false, "message": "无权访问该集群"})
		return
	}

	// 检查 Pod 是否存在
	clientset, _, err := h.clientPool.GetClient(uint(clusterID))
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

	// 如果没有指定容器，使用第一个容器
	if containerName == "" {
		if len(pod.Spec.Containers) > 0 {
			containerName = pod.Spec.Containers[0].Name
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "Pod 没有可连接的容器"})
			return
		}
	}

	// 升级到 WebSocket 连接
	wsConn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("升级 WebSocket 连接失败", zap.Error(err))
		return
	}
	defer wsConn.Close()

	// 创建会话上下文
	sessionCtx, cancel := context.WithCancel(context.Background())

	// 创建 K8s 会话记录
	session := &K8sSession{
		UserID:        userID,
		ClusterID:     uint(clusterID),
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		Status:        "active",
		StartedAt:     time.Now(),
	}
	if err := h.db.Create(session).Error; err != nil {
		logger.Error("创建 K8s 会话记录失败", zap.Error(err))
		cancel()
		return
	}

	// 创建终端会话
	terminalSession := &k8sTerminalSession{
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

	// 注册到活跃会话
	h.sessionsMutex.Lock()
	h.activeSessions[session.ID] = terminalSession
	h.sessionsMutex.Unlock()

	defer func() {
		h.sessionsMutex.Lock()
		delete(h.activeSessions, session.ID)
		h.sessionsMutex.Unlock()
	}()

	// 初始化 K8s executor
	size := &K8sStreamSize{Width: 80, Height: 24}
	executor, err := h.createExecutor(terminalSession, size)
	if err != nil {
		logger.Error("初始化 K8s executor 失败", zap.Error(err))
		h.closeSession(session, "初始化失败")
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

	// 启动会话
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		h.handleSession(terminalSession, session, size)
	}()

	wg.Wait()

	// 关闭会话
	h.closeSession(session, "会话结束")
}

// createExecutor 创建 K8s executor
func (h *K8sTerminalHandler) createExecutor(session *K8sSession, size *K8sStreamSize) (remotecommand.Executor, error) {
	clientset, config, err := h.clientPool.GetClient(session.clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取集群连接失败: %w", err)
	}

	// 直接使用 bash 作为登录 shell
	// 如果某些容器确实没有 bash，可以改为 []string{"/bin/sh"}
	command := []string{"/bin/bash", "-l"}

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
func (h *K8sTerminalHandler) handleSession(session *K8sSession, dbSession *K8sSession, size *K8sStreamSize) {
	// 创建 Streamer 接口实现
	streamer := &k8sTerminalStreamer{
		session:    session,
		dbSession:  dbSession,
		handler:    h,
		stdin:      make(chan []byte, 10),
		stdinDone:  make(chan struct{}),
		stdoutDone: make(chan struct{}),
		sizeQueue:  NewK8sTerminalSizeQueue(),
	}

	// 初始大小
	streamer.sizeQueue.sizes <- remotecommand.TerminalSize{
		Width:  size.Width,
		Height: size.Height,
	}

	// 启动 stdout 处理
	go streamer.handleStdout()

	// 启动 resize 处理
	go streamer.handleResize()

	// 启动 WebSocket 消息读取
	go streamer.handleWebSocketMessages()

	// 执行远程命令
	err := session.k8sExecutor.StreamWithContext(session.ctx, remotecommand.StreamOptions{
		Stdin:             streamer,
		Stdout:            streamer,
		Stderr:            streamer,
		Tty:               true,
		TerminalSizeQueue: streamer.sizeQueue,
	})

	if err != nil {
		logger.Error("执行远程命令失败", zap.Error(err))
	}

	close(streamer.stdinDone)
	<-streamer.stdoutDone
	streamer.sizeQueue.Stop()
}

// closeSession 关闭会话
func (h *K8sTerminalHandler) closeSession(session *K8sSession, reason string) {
	now := time.Now()
	session.Status = "closed"
	session.EndedAt = &now
	session.Duration = int(now.Sub(session.StartedAt).Seconds())
	session.CloseReason = reason

	if err := h.db.Save(session).Error; err != nil {
		logger.Error("更新 K8s 会话状态失败", zap.Error(err))
	}

	logger.Info("K8s 终端会话已关闭",
		zap.Uint("session_id", session.ID),
		zap.String("reason", reason),
		zap.Int("duration", session.Duration))
}

// validateToken 验证token并返回用户ID
func (h *K8sTerminalHandler) validateToken(token string) (uint, error) {
	// 使用JWT工具验证token
	claims, err := utils.ParseToken(token)
	if err != nil {
		return 0, fmt.Errorf("token验证失败: %w", err)
	}

	return claims.UserID, nil
}

// GetActiveSessions 获取活跃的 K8s 终端会话
func (h *K8sTerminalHandler) GetActiveSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "用户未登录"})
		return
	}

	h.sessionsMutex.RLock()
	defer h.sessionsMutex.RUnlock()

	sessions := make([]map[string]interface{}, 0)
	for _, session := range h.activeSessions {
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

// TerminateSession 终止 K8s 终端会话
func (h *K8sTerminalHandler) TerminateSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "用户未登录"})
		return
	}

	sessionID, err := strconv.ParseUint(c.Param("sessionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "无效的会话ID"})
		return
	}

	h.sessionsMutex.RLock()
	session, ok := h.activeSessions[uint(sessionID)]
	h.sessionsMutex.RUnlock()

	if !ok {
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
	session    *k8sTerminalSession
	dbSession  *K8sSession
	handler    *K8sTerminalHandler
	stdin      chan []byte
	stdinDone  chan struct{}
	stdoutDone chan struct{}
	sizeQueue  *K8sTerminalSizeQueue
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
				// 调整终端大小 - 在实际实现中需要通过队列发送
			case "close":
				s.session.cancel()
				return
			case "stdin":
				// 发送 stdin 数据
				select {
				case s.stdin <- []byte(msg.Data):
				case <-s.session.ctx.Done():
					return
				}
			}
		}
	}
}

// handleResize 处理终端大小调整
func (s *k8sTerminalStreamer) handleResize() {
	// 实现终端大小调整逻辑
}
