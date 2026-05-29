package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"oneops/backend/logger"
	"oneops/backend/services"
	"oneops/backend/utils"

	"go.uber.org/zap"
)

// MonitoringWebSocketController 监控 WebSocket 控制器
type MonitoringWebSocketController struct{}

// NewMonitoringWebSocketController 创建 WebSocket 控制器
func NewMonitoringWebSocketController() *MonitoringWebSocketController {
	return &MonitoringWebSocketController{}
}

// WebSocketUpgrader WebSocket 升级器
var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应验证）
	},
}

// HandleWebSocket WebSocket 连接处理
func (c *MonitoringWebSocketController) HandleWebSocket(ctx *gin.Context) {
	// 验证 Token - 从原始URL手动解析query参数
	// 注意：WebSocket升级时，ctx.Query可能无法正确获取参数
	token := ctx.Query("token")
	if token == "" {
		// 从原始URL手动解析
		rawURL := ctx.Request.URL.String()
		if idx := strings.Index(rawURL, "?token="); idx > 0 {
			token = rawURL[idx+7:]
			if endIdx := strings.Index(token, "&"); endIdx > 0 {
				token = token[:endIdx]
			}
		}
	}

	// 如果还是没有，尝试从 Authorization header 获取
	if token == "" {
		token = ctx.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	logger.Info("WebSocket 连接请求", zap.String("token", token[:20]+"..."), zap.String("rawURL", ctx.Request.URL.String()))

	// 验证 Token 有效性
	if token == "" {
		logger.Warn("WebSocket Token 为空")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证 Token"})
		return
	}

	_, err := utils.ParseToken(token)
	if err != nil {
		logger.Warn("WebSocket Token 验证失败", zap.Error(err))
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token"})
		return
	}

	// 升级 HTTP 连接到 WebSocket
	conn, err := WebSocketUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Warn("WebSocket 升级失败", zap.Error(err))
		return
	}

	// 创建客户端
	clientID := uuid.New().String()
	hub := services.GetMonitoringHub()

	client := &services.WebSocketClient{
		ID:   clientID,
		Conn: conn,
		Send: make(chan *services.WebSocketMessage, 256),
		Hub:  hub,
		Topics: map[string]bool{
			"overview": true, // 默认订阅概览数据
			"alerts":   true, // 默认订阅告警
		},
	}

	// 注册客户端
	hub.RegisterClient(client)

	// 启动读写循环
	go client.WritePump()
	go client.ReadPump()

	// 发送欢迎消息
	welcomeMsg := &services.WebSocketMessage{
		Type:      "connected",
		Data:      map[string]string{"clientID": clientID},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	client.Send <- welcomeMsg

	logger.Info("WebSocket 连接已建立",
		zap.String("clientID", clientID),
		zap.String("remoteAddr", ctx.Request.RemoteAddr))
}

// BroadcastOverview 广播监控概览数据
func (c *MonitoringWebSocketController) BroadcastOverview() {
	hub := services.GetMonitoringHub()
	monitoringService := services.NewMonitoringService()

	overview, err := monitoringService.GetOverview()
	if err != nil {
		logger.Error("获取监控概览失败", zap.Error(err))
		return
	}

	message := &services.WebSocketMessage{
		Type:      "overview",
		Data:      overview,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	hub.Broadcast(message)
	logger.Debug("已广播监控概览数据",
		zap.Int("clientCount", hub.GetClientCount()))
}

// BroadcastAlert 广播告警事件
func (c *MonitoringWebSocketController) BroadcastAlert(alert interface{}) {
	hub := services.GetMonitoringHub()

	message := &services.WebSocketMessage{
		Type:      "alert",
		Data:      alert,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	hub.Broadcast(message)
	logger.Debug("已广播告警事件",
		zap.Int("clientCount", hub.GetClientCount()))
}

// GetConnectedClients 获取已连接的客户端列表（用于管理）
func (c *MonitoringWebSocketController) GetConnectedClients(ctx *gin.Context) {
	hub := services.GetMonitoringHub()

	type ClientInfo struct {
		ID        string            `json:"id"`
		Topics    map[string]bool   `json:"topics"`
		Connected string            `json:"connected"`
	}

	clients := make([]ClientInfo, 0, hub.GetClientCount())
	// 这里需要从 Hub 暴露客户端列表，或者维护一个客户端信息映射

	ctx.JSON(http.StatusOK, gin.H{
		"code":   200,
		"message": "success",
		"data":   clients,
		"total":  hub.GetClientCount(),
	})
}
