package k8s

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/k8s"

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
	sizes chan remotecommand.TerminalSize
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

func (q *k8sTerminalSizeQueue) Stop() {
	close(q.sizes)
}

// HandleWebSocket 处理 WebSocket 连接
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

	hasAccess, err := ctrl.svc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, gin.H{"code": 403, "success": false, "message": "无权访问该集群"})
		return
	}

	clientset, _, err := ctrl.svc.GetClientWithConfig(uint(clusterID))
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
	defer wsConn.Close()

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
	executor, err := ctrl.createExecutor(terminalSession, size)
	if err != nil {
		logger.Error("初始化 K8s executor 失败", zap.Error(err))
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

// createExecutor 创建 K8s executor
func (ctrl *TerminalController) createExecutor(session *k8sSession, size *k8sStreamSize) (remotecommand.Executor, error) {
	clientset, config, err := ctrl.svc.GetClientWithConfig(session.clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取集群连接失败: %w", err)
	}

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
	go streamer.handleResize()
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

// GetActiveSessions 获取活跃的 K8s 终端会话
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

// TerminateSession 终止 K8s 终端会话
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
				// 调整终端大小
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

// handleResize 处理终端大小调整
func (s *k8sTerminalStreamer) handleResize() {
	// 实现终端大小调整逻辑
}
