package notification

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"oneops/backend/logger"

	"go.uber.org/zap"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtpHost"`     // SMTP 服务器地址
	SMTPPort     int    `json:"smtpPort"`     // SMTP 端口
	Username     string `json:"username"`     // 用户名
	Password     string `json:"password"`     // 密码
	From         string `json:"from"`         // 发件人邮箱
	FromName     string `json:"fromName"`     // 发件人名称
	UseTLS      bool   `json:"useTLS"`       // 是否使用 TLS
}

// EmailService 邮件通知服务
type EmailService struct {
	config EmailConfig
}

// NewEmailService 创建邮件通知服务
func NewEmailService(config EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// AlertNotification 告警通知内容
type AlertNotification struct {
	Level     string `json:"level"`     // 告警级别
	Hostname  string `json:"hostname"`  // 主机名
	IP        string `json:"ip"`        // IP 地址
	Message   string `json:"message"`   // 告警消息
	Value     float64 `json:"value"`    // 当前值
	Threshold float64 `json:"threshold"` // 阈值
	Timestamp string `json:"timestamp"` // 时间戳
}

// SendAlert 发送告警邮件
func (s *EmailService) SendAlert(to []string, alert *AlertNotification) error {
	// 构建邮件内容
	subject := s.buildSubject(alert)
	body := s.buildHTMLBody(alert)

	// 发送邮件
	return s.Send(to, subject, body, true)
}

// Send 发送邮件
// to: 收件人列表
// subject: 邮件主题
// body: 邮件内容
// isHTML: 是否为 HTML 格式
func (s *EmailService) Send(to []string, subject, body string, isHTML bool) error {
	if len(to) == 0 {
		return fmt.Errorf("收件人列表不能为空")
	}

	// 验证配置
	if err := s.validateConfig(); err != nil {
		return fmt.Errorf("邮件配置无效: %w", err)
	}

	// 构建邮件地址
	fromAddr := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)

	// 构建 SMTP 地址
	smtpAddr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// 构建邮件内容
	var contentType string
	if isHTML {
		contentType = "text/html; charset=UTF-8"
	} else {
		contentType = "text/plain; charset=UTF-8"
	}

	msg := fmt.Sprintf("From: %s\r\n", fromAddr)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(to, ", "))
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += fmt.Sprintf("Content-Type: %s\r\n", contentType)
	msg += "\r\n" // 空行分隔头部和正文
	msg += body

	// 发送邮件（带重试机制）
	var lastErr error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.Warn("邮件发送失败，重试中",
				zap.Int("attempt", i+1),
				zap.Int("maxRetries", maxRetries),
				zap.Error(lastErr))
			time.Sleep(time.Duration(i) * time.Second) // 指数退避
		}

		if s.config.UseTLS {
			lastErr = s.sendWithTLS(smtpAddr, to, msg)
		} else {
			lastErr = s.sendWithoutTLS(smtpAddr, to, msg)
		}

		if lastErr == nil {
			logger.Info("邮件发送成功",
				zap.Strings("to", to),
				zap.String("subject", subject))
			return nil
		}
	}

	return fmt.Errorf("邮件发送失败（已重试 %d 次）: %w", maxRetries, lastErr)
}

// sendWithTLS 使用 TLS 发送邮件
func (s *EmailService) sendWithTLS(addr string, to []string, msg string) error {
	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.config.SMTPHost,
	}

	// 连接 SMTP 服务器
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %w", err)
	}
	defer conn.Close()

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	// 认证
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}

	// 设置发件人
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取邮件写入器失败: %w", err)
	}
	defer writer.Close()

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	return nil
}

// sendWithoutTLS 不使用 TLS 发送邮件
func (s *EmailService) sendWithoutTLS(addr string, to []string, msg string) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	err := smtp.SendMail(addr, auth, s.config.From, to, []byte(msg))
	if err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// buildSubject 构建邮件主题
func (s *EmailService) buildSubject(alert *AlertNotification) string {
	levelText := s.getLevelText(alert.Level)
	return fmt.Sprintf("[%s] %s - %s", levelText, alert.Hostname, alert.Message)
}

// buildHTMLBody 构建 HTML 邮件正文
func (s *EmailService) buildHTMLBody(alert *AlertNotification) string {
	levelColor := s.getLevelColor(alert.Level)
	levelText := s.getLevelText(alert.Level)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: %s; color: white; padding: 15px; border-radius: 5px 5px 0 0; }
		.content { background-color: #f9f9f9; padding: 20px; border: 1px solid #ddd; border-top: none; border-radius: 0 0 5px 5px; }
		.alert-info { margin: 10px 0; padding: 10px; background-color: white; border-left: 4px solid %s; }
		.label { font-weight: bold; color: #555; }
		.footer { margin-top: 20px; font-size: 12px; color: #999; text-align: center; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2 style="margin: 0;">%s 告警通知</h2>
		</div>
		<div class="content">
			<div class="alert-info">
				<p><span class="label">主机:</span> %s (%s)</p>
				<p><span class="label">级别:</span> %s</p>
				<p><span class="label">消息:</span> %s</p>
				<p><span class="label">当前值:</span> %.2f%%</p>
				<p><span class="label">阈值:</span> %.2f%%</p>
				<p><span class="label">时间:</span> %s</p>
			</div>
			<p style="color: #666; font-size: 14px;">请及时处理此告警，避免影响业务运行。</p>
		</div>
		<div class="footer">
			<p>本邮件由 OneOps 监控系统自动发送，请勿回复。</p>
		</div>
	</div>
</body>
</html>
`, levelColor, levelColor, levelText, alert.Hostname, alert.IP, levelText, alert.Message,
		alert.Value, alert.Threshold, alert.Timestamp)
}

// getLevelText 获取告警级别文本
func (s *EmailService) getLevelText(level string) string {
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

// getLevelColor 获取告警级别颜色
func (s *EmailService) getLevelColor(level string) string {
	switch level {
	case "critical":
		return "#d32f2f" // 红色
	case "high":
		return "#f57c00" // 橙色
	case "medium":
		return "#fbc02d" // 黄色
	case "low":
		return "#1976d2" // 蓝色
	case "info":
		return "#388e3c" // 绿色
	default:
		return "#666666" // 灰色
	}
}

// validateConfig 验证邮件配置
func (s *EmailService) validateConfig() error {
	if s.config.SMTPHost == "" {
		return fmt.Errorf("SMTP 服务器地址不能为空")
	}
	if s.config.SMTPPort <= 0 || s.config.SMTPPort > 65535 {
		return fmt.Errorf("SMTP 端口无效")
	}
	if s.config.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if s.config.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if s.config.From == "" {
		return fmt.Errorf("发件人邮箱不能为空")
	}
	if s.config.FromName == "" {
		s.config.FromName = "OneOps 监控系统"
	}

	// 验证邮箱格式
	if err := s.validateEmail(s.config.From); err != nil {
		return fmt.Errorf("发件人邮箱格式无效: %w", err)
	}

	return nil
}

// validateEmail 验证邮箱格式
func (s *EmailService) validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("邮箱地址不能为空")
	}

	// 简单验证邮箱格式
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("邮箱地址格式无效")
	}

	// 验证域名
	_, domain, found := strings.Cut(email, "@")
	if !found || domain == "" {
		return fmt.Errorf("邮箱地址域名无效")
	}

	// 验证域名有 MX 记录（可选，这里跳过 DNS 查询以提高性能）
	// mxRecords, err := net.LookupMX(domain)
	// if err != nil || len(mxRecords) == 0 {
	// 	return fmt.Errorf("邮箱域名不存在 MX 记录")
	// }

	return nil
}

// TestConnection 测试邮件服务器连接
func (s *EmailService) TestConnection() error {
	// 验证配置
	if err := s.validateConfig(); err != nil {
		return err
	}

	// 构建 SMTP 地址
	smtpAddr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// 测试连接
	timeout := 10 * time.Second
	conn, err := net.DialTimeout("tcp", smtpAddr, timeout)
	if err != nil {
		return fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}
	conn.Close()

	logger.Info("SMTP 连接测试成功",
		zap.String("host", s.config.SMTPHost),
		zap.Int("port", s.config.SMTPPort))

	return nil
}
