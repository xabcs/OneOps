package middleware

import (
	"net/url"
	"strings"
	"time"

	"oneops/backend3/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// sensitiveQueryKeys 查询参数敏感 key 关键词（小写，子串匹配），命中则对参数值脱敏，避免凭证进日志
var sensitiveQueryKeys = []string{
	"password",
	"passwd",
	"token",
	"secret",
	"credential",
	"privatekey",
	"private_key",
	"authorization",
	"apikey",
}

// maskQueryValue 脱敏查询参数值：保留前 8 个字符，其余以 ... 代替；长度不足时直接打码
func maskQueryValue(v string) string {
	runes := []rune(v)
	if len(runes) <= 8 {
		return "***"
	}
	return string(runes[:8]) + "..."
}

// sanitizeQuery 重组查询串并对敏感参数值脱敏（如 ?token=xxx），避免 Authorization/token 明文进日志
func sanitizeQuery(rawQuery string) string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		// 解析失败时整体打码，宁可日志不可读也不泄露参数
		return "***"
	}
	for key, vals := range values {
		lowerKey := strings.ToLower(key)
		for _, keyword := range sensitiveQueryKeys {
			if strings.Contains(lowerKey, keyword) {
				for i := range vals {
					vals[i] = maskQueryValue(vals[i])
				}
				break
			}
		}
	}
	return values.Encode()
}

// RequestLogger 自动记录每个请求的方法、路径、耗时、状态码、用户ID
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		// 查询串脱敏后再入日志，避免 token 等敏感参数明文输出
		query := sanitizeQuery(c.Request.URL.RawQuery)

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("duration", duration),
			zap.Int("status_code", statusCode),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}

		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok && uid > 0 {
				fields = append(fields, zap.Uint("user_id", uid))
			}
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		if statusCode >= 500 {
			logger.Error("HTTP请求", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP请求", fields...)
		} else {
			logger.Info("HTTP请求", fields...)
		}
	}
}
