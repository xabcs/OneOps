package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhookUrl"` // 机器人 Webhook URL（含 access_token）
	Secret     string `json:"secret"`     // 加签密钥（SEC 开头，可选；配置了才加签）
}

// DingTalkService 钉钉通知服务
type DingTalkService struct {
	config     DingTalkConfig
	httpClient *http.Client
}

// NewDingTalkService 创建钉钉通知服务
func NewDingTalkService(config DingTalkConfig) *DingTalkService {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &DingTalkService{
		config:     config,
		httpClient: httpClient,
	}
}

// DingTalkMessage 钉钉消息结构
type DingTalkMessage struct {
	MsgType string        `json:"msgtype"` // 消息类型：text, markdown
	Text    *DingTalkText `json:"text,omitempty"`
	At      *DingTalkAt   `json:"at,omitempty"`
}

// DingTalkText 文本消息
type DingTalkText struct {
	Content string `json:"content"`
}

// DingTalkAt @人设置
type DingTalkAt struct {
	AtMobiles []string `json:"atMobiles,omitempty"` // 被@人的手机号
	AtUserIds []string `json:"atUserIds,omitempty"` // 被@人的 userid（企业内员工标识，优先于手机号）
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// DingTalkResponse 钉钉 API 响应
type DingTalkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendText 发送文本消息（atMobiles 为被@人的手机号列表）
func (s *DingTalkService) SendText(content string, atMobiles ...string) error {
	message := &DingTalkMessage{
		MsgType: "text",
		Text: &DingTalkText{
			Content: content,
		},
	}
	if len(atMobiles) > 0 {
		message.At = &DingTalkAt{AtMobiles: atMobiles}
	}

	return s.Send(message)
}

// SendTextAt 发送文本消息（@人优先 userid，无 userid 者回退手机号）
func (s *DingTalkService) SendTextAt(content string, atUserIDs, atMobiles []string) error {
	message := &DingTalkMessage{
		MsgType: "text",
		Text:    &DingTalkText{Content: content},
	}
	if len(atUserIDs) > 0 || len(atMobiles) > 0 {
		message.At = &DingTalkAt{AtUserIds: atUserIDs, AtMobiles: atMobiles}
	}
	return s.Send(message)
}

// Send 发送消息（配置了 Secret 时按加签方式请求）
func (s *DingTalkService) Send(message *DingTalkMessage) error {
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("钉钉配置无效: %w", err)
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	logger.Debug("发送钉钉消息",
		zap.String("url", s.config.WebhookURL),
		zap.String("message", string(jsonData)))

	var lastErr error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.Warn("钉钉消息发送失败，重试中",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
				zap.Error(lastErr))
			time.Sleep(time.Duration(i) * time.Second)
		}

		req, err := http.NewRequest("POST", s.signedURL(), bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("创建 HTTP 请求失败: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("发送 HTTP 请求失败: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %w", err)
			continue
		}

		var dingtalkResp DingTalkResponse
		if err := json.Unmarshal(body, &dingtalkResp); err != nil {
			lastErr = fmt.Errorf("解析响应失败: %w", err)
			continue
		}

		if dingtalkResp.ErrCode == 0 {
			logger.Info("钉钉消息发送成功")
			return nil
		}

		lastErr = fmt.Errorf("钉钉 API 返回错误: [%d] %s", dingtalkResp.ErrCode, dingtalkResp.ErrMsg)
	}

	return fmt.Errorf("钉钉消息发送失败（已重试 %d 次）: %w", maxRetries, lastErr)
}

// signedURL 计算加签后的请求 URL：timestamp+"\n"+secret 作 HMAC-SHA256，
// base64 后 URL encode，拼到 webhook 上（未配置 secret 时直接返回原 webhook）
func (s *DingTalkService) signedURL() string {
	if s.config.Secret == "" {
		return s.config.WebhookURL
	}

	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, s.config.Secret)

	mac := hmac.New(sha256.New, []byte(s.config.Secret))
	mac.Write([]byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	return fmt.Sprintf("%s&timestamp=%d&sign=%s", s.config.WebhookURL, timestamp, sign)
}

// validateConfig 验证钉钉配置
func (s *DingTalkService) validateConfig() error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("Webhook URL 不能为空")
	}

	if !strings.HasPrefix(s.config.WebhookURL, "https://") &&
		!strings.HasPrefix(s.config.WebhookURL, "http://") {
		return fmt.Errorf("Webhook URL 格式无效，必须以 http:// 或 https:// 开头")
	}

	if !strings.Contains(s.config.WebhookURL, "access_token=") {
		return fmt.Errorf("Webhook URL 不包含 access_token 参数")
	}

	return nil
}
