package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"oneops/backend/logger"
)

// PerformanceConfig 性能监控配置
type PerformanceConfig struct {
	// SlowThreshold 慢请求阈值（毫秒）
	SlowThreshold int64
	// EnableDetailedLog 是否启用详细日志（记录请求和响应体）
	EnableDetailedLog bool
	// EnableMetrics 是否启用性能指标收集
	EnableMetrics bool
}

// DefaultPerformanceConfig 默认配置
var DefaultPerformanceConfig = PerformanceConfig{
	SlowThreshold:     1000, // 1秒
	EnableDetailedLog: false,
	EnableMetrics:     true,
}

// PerformanceMiddleware 性能监控中间件
type PerformanceMiddleware struct {
	config PerformanceConfig
	logger *zap.Logger
}

// NewPerformanceMiddleware 创建性能监控中间件
func NewPerformanceMiddleware(config ...PerformanceConfig) *PerformanceMiddleware {
	cfg := DefaultPerformanceConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return &PerformanceMiddleware{
		config: cfg,
		logger: logger.Logger,
	}
}

// Monitor 性能监控中间件函数
func (pm *PerformanceMiddleware) Monitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算处理时长
		duration := time.Since(startTime)
		durationMs := duration.Milliseconds()

		// 获取请求信息
		path := c.Request.URL.Path
		method := c.Request.Method
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 判断是否为慢请求
		isSlowRequest := durationMs > pm.config.SlowThreshold

		// 构建日志字段
		logFields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int64("duration_ms", durationMs),
			zap.Bool("is_slow", isSlowRequest),
		}

		// 添加用户信息（如果存在）
		if userID, exists := c.Get("user_id"); exists {
			logFields = append(logFields, zap.Any("user_id", userID))
		}
		if username, exists := c.Get("username"); exists {
			logFields = append(logFields, zap.String("username", username.(string)))
		}

		// 根据请求状态和时长记录不同级别的日志
		if statusCode >= 500 {
			// 服务器错误
			pm.logger.Error("API请求失败", logFields...)
		} else if statusCode >= 400 {
			// 客户端错误
			if isSlowRequest {
				pm.logger.Warn("API慢请求（客户端错误）", logFields...)
			} else {
				pm.logger.Info("API请求（客户端错误）", logFields...)
			}
		} else {
			// 成功请求
			if isSlowRequest {
				pm.logger.Warn("API慢请求", logFields...)
			} else {
				pm.logger.Debug("API请求成功", logFields...)
			}
		}

		// 如果启用详细日志，记录请求和响应体（注意：不要记录敏感信息）
		if pm.config.EnableDetailedLog && isSlowRequest {
			pm.logger.Debug("慢请求详情",
				append(logFields,
					zap.String("query", c.Request.URL.RawQuery),
					zap.Int64("request_size", c.Request.ContentLength),
					zap.Int("response_size", c.Writer.Size()),
				)...)
		}
	}
}

// PrometheusMetrics Prometheus指标中间件
func (pm *PerformanceMiddleware) PrometheusMetrics() gin.HandlerFunc {
	// TODO: 集成Prometheus客户端库
	// 目前先记录日志，后续可以集成Prometheus
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		path := c.Request.URL.Path
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// 记录性能指标
		pm.logger.Debug("性能指标",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
		)

		// TODO: 使用Prometheus客户端记录指标
		// httpDuration.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Observe(duration.Seconds())
	}
}

// SlowQueryLog 慢查询日志记录（用于数据库操作）
type SlowQueryLog struct {
	SQL       string
	Duration  time.Duration
	Threshold time.Duration
}

// LogSlowQuery 记录慢查询
func LogSlowQuery(sql string, duration time.Duration, threshold time.Duration) {
	if duration > threshold {
		logger.Logger.Warn("数据库慢查询",
			zap.String("sql", sql),
			zap.Duration("duration", duration),
			zap.Duration("threshold", threshold),
			zap.Int64("duration_ms", duration.Milliseconds()),
		)
	}
}

// GetSlowRequestWarning 获取慢请求警告信息
func GetSlowRequestWarning(path string, method string, durationMs int64, threshold int64) string {
	return fmt.Sprintf("慢请求检测: %s %s 耗时 %dms，超过阈值 %dms",
		method, path, durationMs, threshold)
}

// GetPerformanceStats 获取性能统计信息（可用于监控面板）
func GetPerformanceStats() map[string]interface{} {
	// TODO: 实现性能统计收集
	// 可以使用Redis或内存存储性能统计数据
	return map[string]interface{}{
		"total_requests":     0,
		"slow_requests":      0,
		"avg_duration_ms":    0,
		"p95_duration_ms":    0,
		"p99_duration_ms":    0,
		"error_rate":         0.0,
		"last_updated_time":  time.Now(),
	}
}
