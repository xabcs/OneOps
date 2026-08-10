package notification

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"oneops/backend2/pkg/logger"

	"go.uber.org/zap"
)

// WeChatConfig 企业微信配置
type WeChatConfig struct {
	WebhookURL string `json:"webhookUrl"` // 机器人 Webhook URL
}

// WeChatService 企业微信通知服务
type WeChatService struct {
	config     WeChatConfig
	httpClient *http.Client
}

// NewWeChatService 创建企业微信通知服务
func NewWeChatService(config WeChatConfig) *WeChatService {
	// 创建 HTTP 客户端，跳过 TLS 验证（内网环境）
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &WeChatService{
		config:     config,
		httpClient: httpClient,
	}
}

// WeChatMessage 企业微信消息结构
type WeChatMessage struct {
	MsgType  string          `json:"msgtype"` // 消息类型：text, markdown
	Text     *WeChatText     `json:"text,omitempty"`
	Markdown *WeChatMarkdown `json:"markdown,omitempty"`
}

// WeChatText 文本消息
type WeChatText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`        // @的用户
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"` // @的手机号
}

// WeChatMarkdown Markdown 消息
type WeChatMarkdown struct {
	Content string `json:"content"`
}

// WeChatResponse 企业微信 API 响应
type WeChatResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendAlert 发送告警通知（使用 Markdown 格式）
func (s *WeChatService) SendAlert(alert *AlertNotification) error {
	// 构建 Markdown 消息
	content := s.buildMarkdownContent(alert)

	message := &WeChatMessage{
		MsgType: "markdown",
		Markdown: &WeChatMarkdown{
			Content: content,
		},
	}

	return s.Send(message)
}

// SendText 发送文本消息
func (s *WeChatService) SendText(content string, mentionedList ...string) error {
	message := &WeChatMessage{
		MsgType: "text",
		Text: &WeChatText{
			Content:             content,
			MentionedList:       mentionedList,
			MentionedMobileList: []string{},
		},
	}

	return s.Send(message)
}

// Send 发送消息
func (s *WeChatService) Send(message *WeChatMessage) error {
	// 验证配置
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("企业微信配置无效: %w", err)
	}

	// 序列化消息
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	logger.Debug("发送企业微信消息",
		zap.String("url", s.config.WebhookURL),
		zap.String("message", string(jsonData)))

	// 发送消息（带重试机制）
	var lastErr error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.Warn("企业微信消息发送失败，重试中",
				zap.Int("attempt", i+1),
				zap.Int("maxRetries", maxRetries),
				zap.Error(lastErr))
			time.Sleep(time.Duration(i) * time.Second)
		}

		// 创建 HTTP 请求
		req, err := http.NewRequest("POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("创建 HTTP 请求失败: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		// 发送请求
		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("发送 HTTP 请求失败: %w", err)
			continue
		}

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %w", err)
			continue
		}

		// 解析响应
		var wechatResp WeChatResponse
		if err := json.Unmarshal(body, &wechatResp); err != nil {
			lastErr = fmt.Errorf("解析响应失败: %w", err)
			continue
		}

		// 检查响应
		if wechatResp.ErrCode == 0 {
			logger.Info("企业微信消息发送成功")
			return nil
		}

		lastErr = fmt.Errorf("企业微信 API 返回错误: [%d] %s", wechatResp.ErrCode, wechatResp.ErrMsg)
	}

	return fmt.Errorf("企业微信消息发送失败（已重试 %d 次）: %w", maxRetries, lastErr)
}

// buildMarkdownContent 构建 Markdown 消息内容
func (s *WeChatService) buildMarkdownContent(alert *AlertNotification) string {
	levelEmoji := s.getLevelEmoji(alert.Level)
	levelText := s.getLevelText(alert.Level)

	// 企业微信 Markdown 颜色不支持，使用 emoji 和格式化增强可读性
	content := fmt.Sprintf("%s **%s告警**\n", levelEmoji, levelText)
	content += fmt.Sprintf("> **主机**: %s (%s)\n", alert.Hostname, alert.IP)
	content += fmt.Sprintf("> **级别**: %s\n", levelText)
	content += fmt.Sprintf("> **消息**: %s\n", alert.Message)
	content += fmt.Sprintf("> **当前值**: %.2f%%\n", alert.Value)
	content += fmt.Sprintf("> **阈值**: %.2f%%\n", alert.Threshold)
	content += fmt.Sprintf("> **时间**: %s\n", alert.Timestamp)
	content += fmt.Sprintf("\n请及时处理此告警，避免影响业务运行。")

	return content
}

// getLevelEmoji 获取告警级别 Emoji
func (s *WeChatService) getLevelEmoji(level string) string {
	switch level {
	case "critical":
		return "🚨"
	case "high":
		return "⚠️"
	case "medium":
		return "🔔"
	case "low":
		return "ℹ️"
	case "info":
		return "✅"
	default:
		return "📊"
	}
}

// getLevelText 获取告警级别文本
func (s *WeChatService) getLevelText(level string) string {
	switch level {
	case "critical":
		return "严重"
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	case "info":
		return "信息"
	default:
		return level
	}
}

// validateConfig 验证企业微信配置
func (s *WeChatService) validateConfig() error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("Webhook URL 不能为空")
	}

	// 验证 URL 格式
	if !strings.HasPrefix(s.config.WebhookURL, "https://") &&
		!strings.HasPrefix(s.config.WebhookURL, "http://") {
		return fmt.Errorf("Webhook URL 格式无效，必须以 http:// 或 https:// 开头")
	}

	// 验证 URL 包含 qpye_key
	if !strings.Contains(s.config.WebhookURL, "key=") {
		return fmt.Errorf("Webhook URL 不包含 key 参数")
	}

	return nil
}

// TestConnection 测试企业微信 Webhook 连接
func (s *WeChatService) TestConnection() error {
	// 验证配置
	if err := s.validateConfig(); err != nil {
		return err
	}

	// 发送测试消息
	testMessage := &WeChatMessage{
		MsgType: "text",
		Text: &WeChatText{
			Content: "✅ OneOps 监控系统连接测试成功！",
		},
	}

	// 发送消息（不重试）
	jsonData, err := json.Marshal(testMessage)
	if err != nil {
		return fmt.Errorf("序列化测试消息失败: %w", err)
	}

	req, err := http.NewRequest("POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建测试请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送测试请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取测试响应失败: %w", err)
	}

	var wechatResp WeChatResponse
	if err := json.Unmarshal(body, &wechatResp); err != nil {
		return fmt.Errorf("解析测试响应失败: %w", err)
	}

	if wechatResp.ErrCode != 0 {
		return fmt.Errorf("企业微信 API 返回错误: [%d] %s", wechatResp.ErrCode, wechatResp.ErrMsg)
	}

	logger.Info("企业微信 Webhook 测试成功")

	return nil
}
