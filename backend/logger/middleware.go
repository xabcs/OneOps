package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLogger 返回Gin日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", ip),
			zap.String("user-agent", userAgent),
		}

		// 添加用户信息（如果有）
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}
		if rawUsername, exists := c.Get("username"); exists {
			if username, ok := rawUsername.(string); ok {
				fields = append(fields, zap.String("username", username))
			}
		}

		// 根据状态码选择日志级别
		switch {
		case status >= 500:
			Error("HTTP请求", fields...)
		case status >= 400:
			Warn("HTTP请求", fields...)
		default:
			Info("HTTP请求", fields...)
		}

		// 如果请求有错误，记录错误信息
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				Error("请求错误", zap.String("error", e.Error()))
			}
		}
	}
}

// GinRecovery Gin恢复中间件，记录panic
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic
				Error("Panic捕获",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)

				// 返回500错误
				c.JSON(500, gin.H{
					"code":    500,
					"success": false,
					"message": "服务器内部错误",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
