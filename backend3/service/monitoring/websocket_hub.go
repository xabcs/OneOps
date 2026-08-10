package monitoring

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// WebSocketMessage WebSocket 消息结构
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

// WebSocketClient WebSocket 客户端
type WebSocketClient struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan *WebSocketMessage
	Hub    *MonitoringHub
	Topics map[string]bool
	mu     sync.Mutex
}

// MonitoringHub WebSocket 连接管理中心
type MonitoringHub struct {
	clients    map[*WebSocketClient]bool
	clientIDs  map[string]*WebSocketClient
	broadcast  chan *WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	topics     map[string]map[*WebSocketClient]bool
	mu         sync.RWMutex
}

// NewMonitoringHub 创建 WebSocket Hub
func NewMonitoringHub() *MonitoringHub {
	hub := &MonitoringHub{
		clients:    make(map[*WebSocketClient]bool),
		clientIDs:  make(map[string]*WebSocketClient),
		broadcast:  make(chan *WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		topics:     make(map[string]map[*WebSocketClient]bool),
	}

	go hub.run()
	go hub.subscribeRedisEvents()

	return hub
}

// run 运行 Hub 主循环
func (h *MonitoringHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.clientIDs[client.ID] = client
			h.mu.Unlock()
			logger.Info("WebSocket 客户端已连接",
				zap.String("client_id", client.ID),
				zap.Int("total_clients", len(h.clients)))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.clientIDs, client.ID)
				close(client.Send)
			}
			for _, clients := range h.topics {
				delete(clients, client)
			}
			h.mu.Unlock()
			logger.Info("WebSocket 客户端已断开",
				zap.String("client_id", client.ID),
				zap.Int("total_clients", len(h.clients)))

		case message := <-h.broadcast:
			h.mu.RLock()
			topic := message.Type
			if clients, ok := h.topics[topic]; ok {
				for client := range clients {
					select {
					case client.Send <- message:
					default:
						h.unregister <- client
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// subscribeRedisEvents 订阅 Redis 事件
func (h *MonitoringHub) subscribeRedisEvents() {
	if !database.IsRedisEnabled() {
		return
	}

	cache := database.NewRedisCache()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := cache.SubscribeAlerts(ctx)
	if ch == nil {
		return
	}

	for msg := range ch {
		var alertData map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Payload), &alertData); err != nil {
			logger.Warn("解析 Redis 告警事件失败", zap.Error(err))
			continue
		}

		wsMessage := &WebSocketMessage{
			Type:      "alert",
			Data:      alertData,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		h.Broadcast(wsMessage)
	}
}

// RegisterClient 注册客户端
func (h *MonitoringHub) RegisterClient(client *WebSocketClient) {
	h.register <- client
}

// UnregisterClient 注销客户端
func (h *MonitoringHub) UnregisterClient(client *WebSocketClient) {
	h.unregister <- client
}

// Subscribe 订阅主题
func (h *MonitoringHub) Subscribe(client *WebSocketClient, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.topics[topic] == nil {
		h.topics[topic] = make(map[*WebSocketClient]bool)
	}
	h.topics[topic][client] = true

	client.mu.Lock()
	if client.Topics == nil {
		client.Topics = make(map[string]bool)
	}
	client.Topics[topic] = true
	client.mu.Unlock()

	logger.Debug("客户端订阅主题",
		zap.String("client_id", client.ID),
		zap.String("topic", topic))
}

// Unsubscribe 取消订阅主题
func (h *MonitoringHub) Unsubscribe(client *WebSocketClient, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.topics[topic]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.topics, topic)
		}
	}

	client.mu.Lock()
	delete(client.Topics, topic)
	client.mu.Unlock()

	logger.Debug("客户端取消订阅主题",
		zap.String("client_id", client.ID),
		zap.String("topic", topic))
}

// Broadcast 广播消息到所有订阅该主题的客户端
func (h *MonitoringHub) Broadcast(message *WebSocketMessage) {
	h.broadcast <- message
}

// SendToClient 发送消息给指定客户端
func (h *MonitoringHub) SendToClient(clientID string, message *WebSocketMessage) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clientIDs[clientID]
	if !ok {
		return false
	}

	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

// GetClientCount 获取客户端数量
func (h *MonitoringHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// WritePump 写入循环
func (c *WebSocketClient) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteJSON(message); err != nil {
				logger.Warn("WebSocket 写入消息失败",
					zap.String("client_id", c.ID),
					zap.Error(err))
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump 读取循环
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Warn("WebSocket 读取错误",
					zap.String("client_id", c.ID),
					zap.Error(err))
			}
			break
		}

		c.handleClientMessage(message)
	}
}

// handleClientMessage 处理客户端消息
func (c *WebSocketClient) handleClientMessage(data []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Warn("解析客户端消息失败", zap.Error(err))
		return
	}

	action, ok := msg["action"].(string)
	if !ok {
		return
	}

	switch action {
	case "subscribe":
		if topic, ok := msg["topic"].(string); ok {
			c.Hub.Subscribe(c, topic)
		}
	case "unsubscribe":
		if topic, ok := msg["topic"].(string); ok {
			c.Hub.Unsubscribe(c, topic)
		}
	}
}

var monitoringHub *MonitoringHub

// GetMonitoringHub 获取监控 WebSocket Hub
func GetMonitoringHub() *MonitoringHub {
	if monitoringHub == nil {
		monitoringHub = NewMonitoringHub()
	}
	return monitoringHub
}
