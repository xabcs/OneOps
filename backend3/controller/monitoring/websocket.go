package monitoring

import (
	"net/http"
	"strings"
	"time"

	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"go.uber.org/zap"
)

// monitoringUpgrader WebSocket 升级器
var monitoringUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket godoc
// @Summary      监控 WebSocket 连接
// @Description  升级 HTTP 连接为 WebSocket，用于实时推送监控概览与告警事件。鉴权通过 query 参数 token（或 Authorization 头）完成，不经过 Auth 中间件
// @Tags         监控-WebSocket
// @Param        token  query     string  true  "JWT token"
// @Success      101  {string}  string  "升级为 WebSocket 连接"
// @Failure      401  {object}  utils.Response  "缺少认证 Token 或 无效的 Token"
// @Router       /monitoring/ws [get]
func (ctrl *MonitoringController) HandleWebSocket(c *gin.Context) {
	// 验证 Token - 优先从 query 参数获取
	token := c.Query("token")
	if token == "" {
		rawURL := c.Request.URL.String()
		if idx := strings.Index(rawURL, "?token="); idx > 0 {
			token = rawURL[idx+7:]
			if endIdx := strings.Index(token, "&"); endIdx > 0 {
				token = token[:endIdx]
			}
		}
	}

	// 尝试从 Authorization header 获取
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	logger.Info("WebSocket 连接请求",
		zap.String("token", tokenMask(token)),
		zap.String("raw_url", c.Request.URL.String()))

	if token == "" {
		logger.Warn("WebSocket Token 为空")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证 Token"})
		return
	}

	if _, err := utils.ParseToken(token); err != nil {
		logger.Warn("WebSocket Token 验证失败", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token"})
		return
	}

	// 升级 HTTP 连接到 WebSocket
	conn, err := monitoringUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warn("WebSocket 升级失败", zap.Error(err))
		return
	}

	// 创建客户端
	clientID := uuid.New().String()
	hub := GetMonitoringHub()

	client := &WebSocketClient{
		ID:   clientID,
		Conn: conn,
		Send: make(chan *WebSocketMessage, 256),
		Hub:  hub,
		Topics: map[string]bool{
			"overview": true,
			"alerts":   true,
		},
	}

	// 注册客户端
	hub.RegisterClient(client)

	// 启动读写循环
	go client.WritePump()
	go client.ReadPump()

	// 发送欢迎消息
	welcomeMsg := &WebSocketMessage{
		Type:      "connected",
		Data:      map[string]string{"clientID": clientID},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	client.Send <- welcomeMsg

	logger.Info("WebSocket 连接已建立",
		zap.String("client_id", clientID),
		zap.String("remote_addr", c.Request.RemoteAddr))
}

// BroadcastOverview 广播监控概览数据
func (ctrl *MonitoringController) BroadcastOverview() {
	hub := GetMonitoringHub()

	overview, err := ctrl.svc.GetOverview()
	if err != nil {
		logger.Error("获取监控概览失败", zap.Error(err))
		return
	}

	message := &WebSocketMessage{
		Type:      "overview",
		Data:      overview,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	hub.Broadcast(message)
	logger.Debug("已广播监控概览数据",
		zap.Int("client_count", hub.GetClientCount()))
}

// BroadcastAlert 广播告警事件
func (ctrl *MonitoringController) BroadcastAlert(alert interface{}) {
	hub := GetMonitoringHub()

	message := &WebSocketMessage{
		Type:      "alert",
		Data:      alert,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	hub.Broadcast(message)
	logger.Debug("已广播告警事件",
		zap.Int("client_count", hub.GetClientCount()))
}

// GetConnectedClients godoc
// @Summary      获取已连接的 WebSocket 客户端列表
// @Description  返回监控推送 Hub 当前在线客户端（注意：该 handler 尚未注册路由）
// @Tags         监控-实时推送
// @Produce      json
// @Success      200  {object}  object  "客户端列表与总数"
// @Router       /monitoring/ws/clients [get]
// @Security     BearerAuth
func (ctrl *MonitoringController) GetConnectedClients(c *gin.Context) {
	hub := GetMonitoringHub()

	type ClientInfo struct {
		ID        string          `json:"id"`
		Topics    map[string]bool `json:"topics"`
		Connected string          `json:"connected"`
	}

	clients := make([]ClientInfo, 0, hub.GetClientCount())

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    clients,
		"total":   hub.GetClientCount(),
	})
}

// tokenMask 遮掩 token 用于日志输出
func tokenMask(token string) string {
	if len(token) > 20 {
		return token[:20] + "..."
	}
	return token
}
