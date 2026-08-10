package middleware

import (
	"time"

	"oneops/backend3/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger 自动记录每个请求的方法、路径、耗时、状态码、用户ID
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		query := c.Request.URL.RawQuery

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
