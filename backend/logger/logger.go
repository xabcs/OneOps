package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"oneops/backend/config"
)

var Logger *zap.Logger

// 请求追踪上下文Key
type contextKey string

const (
	// TraceIDKey 请求追踪ID
	TraceIDKey contextKey = "trace_id"
	// UserIDKey 用户ID
	UserIDKey contextKey = "user_id"
	// RequestIDKey 请求ID
	RequestIDKey contextKey = "request_id"
)

// InitLogger 初始化日志系统
func InitLogger(cfg config.LogConfig) error {
	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 日志级别
	level := getLogLevel(cfg.Level)

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 打开日志文件
	fileWriter, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 文件输出（JSON格式）
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		level,
	)

	// 控制台输出（彩色、开发友好）
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// 组合核心（同时输出到文件和控制台）
	core := zapcore.NewTee(fileCore, consoleCore)

	// 创建Logger
	Logger = zap.New(core,
		zap.AddCaller(),                       // 添加调用者信息
		zap.AddCallerSkip(0),                  // 跳过的调用栈层数
		zap.AddStacktrace(zapcore.ErrorLevel), // Error级别及以上记录堆栈
	)

	return nil
}

// getLogLevel 将字符串转换为zap日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Sync 刷新日志缓冲区
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Debug 记录Debug级别日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 记录Info级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 记录Warn级别日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 记录Error级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 记录Fatal级别日志并退出
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// With 创建带有预设置字段的子logger
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

// ========== 日志增强功能 ==========

// WithTraceID 添加请求追踪ID
func WithTraceID(traceID string) *zap.Logger {
	return Logger.With(zap.String("trace_id", traceID))
}

// WithRequestContext 从请求上下文创建带追踪信息的logger
func WithRequestContext(ctx context.Context) *zap.Logger {
	fields := []zap.Field{}

	// 添加追踪ID
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		fields = append(fields, zap.String("trace_id", traceID.(string)))
	}

	// 添加用户ID
	if userID := ctx.Value(UserIDKey); userID != nil {
		fields = append(fields, zap.Any("user_id", userID))
	}

	// 添加请求ID
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	return Logger.With(fields...)
}

// LogRequest 记录HTTP请求
func LogRequest(method, path, clientIP, userAgent string, duration time.Duration, statusCode int) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("client_ip", clientIP),
		zap.String("user_agent", userAgent),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
	}

	// 根据状态码和时长选择日志级别
	if statusCode >= 500 {
		Logger.Error("HTTP请求失败", fields...)
	} else if statusCode >= 400 || duration > time.Second {
		Logger.Warn("HTTP请求（慢或失败）", fields...)
	} else {
		Logger.Debug("HTTP请求", fields...)
	}
}

// LogDBQuery 记录数据库查询
func LogDBQuery(sql string, duration time.Duration, rowsAffected int64) {
	fields := []zap.Field{
		zap.String("sql", maskSensitiveData(sql)),
		zap.Duration("duration", duration),
		zap.Int64("rows_affected", rowsAffected),
	}

	if duration > time.Second {
		Logger.Warn("数据库慢查询", fields...)
	} else {
		Logger.Debug("数据库查询", fields...)
	}
}

// LogCacheOperation 记录缓存操作
func LogCacheOperation(operation, key string, hit bool, duration time.Duration) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", maskSensitiveData(key)),
		zap.Bool("hit", hit),
		zap.Duration("duration", duration),
	}

	Logger.Debug("缓存操作", fields...)
}

// LogExternalAPI 调用外部API
func LogExternalAPI(url, method string, duration time.Duration, statusCode int, err error) {
	fields := []zap.Field{
		zap.String("url", maskURL(url)),
		zap.String("method", method),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Logger.Error("外部API调用失败", fields...)
	} else if duration > time.Second {
		Logger.Warn("外部API慢调用", fields...)
	} else {
		Logger.Debug("外部API调用", fields...)
	}
}

// LogBusinessEvent 记录业务事件
func LogBusinessEvent(event string, data map[string]interface{}) {
	fields := []zap.Field{zap.String("event", event)}

	for k, v := range data {
		fields = append(fields, zap.Any(k, maskSensitiveValue(k, v)))
	}

	Logger.Info("业务事件", fields...)
}

// LogErrorWithStack 记录错误和堆栈信息
func LogErrorWithStack(msg string, err error, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.Error(err),
		zap.StackSkip("stack", 1), // 跳过当前函数
	}, fields...)

	Logger.Error(msg, allFields...)
}

// ========== 敏感信息脱敏 ==========

// 敏感字段列表
var sensitiveFields = map[string]bool{
	"password":       true,
	"passwd":         true,
	"pwd":            true,
	"secret":         true,
	"token":          true,
	"access_token":   true,
	"refresh_token":  true,
	"api_key":        true,
	"apikey":         true,
	"authorization":  true,
	"cookie":         true,
	"session_id":     true,
	"sessionid":      true,
	"credit_card":    true,
	"ssn":            true,
	"private_key":    true,
}

// maskSensitiveData 脱敏SQL语句中的敏感数据
func maskSensitiveData(sql string) string {
	// 简单实现：移除常见的敏感值
	// TODO: 实现更完善的SQL解析和脱敏
	if len(sql) > 500 {
		return sql[:500] + "...(truncated)"
	}
	return sql
}

// maskURL 脱敏URL中的敏感参数
func maskURL(url string) string {
	// TODO: 实现URL参数脱敏
	// 例如：移除token、password等参数
	return url
}

// maskSensitiveValue 脱敏敏感字段值
func maskSensitiveValue(key string, value interface{}) interface{} {
	// 检查是否为敏感字段
	if sensitiveFields[key] {
		return "***"
	}

	// 检查值中是否包含敏感信息
	if str, ok := value.(string); ok {
		// 检查是否为密码模式（至少8位，包含字母和数字）
		if len(str) >= 8 && hasPasswordPattern(str) {
			return "***"
		}
	}

	return value
}

// hasPasswordPattern 检查字符串是否像密码
func hasPasswordPattern(s string) bool {
	hasLetter := false
	hasDigit := false

	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}

// GetLogLevel 获取当前日志级别
func GetLogLevel() string {
	if Logger == nil {
		return "info"
	}
	// TODO: 从Logger获取当前日志级别
	return "info"
}

// SetLogLevel 动态设置日志级别（开发时有用）
func SetLogLevel(level string) error {
	// TODO: 实现动态日志级别切换
	return fmt.Errorf("动态日志级别切换尚未实现")
}

// Flush 刷新日志缓冲区
func Flush() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// ========== 结构化日志辅助方法 ==========

// LogWithFields 记录带自定义字段的日志
func LogWithFields(level string, msg string, fields ...zap.Field) {
	switch level {
	case "debug":
		Logger.Debug(msg, fields...)
	case "info":
		Logger.Info(msg, fields...)
	case "warn":
		Logger.Warn(msg, fields...)
	case "error":
		Logger.Error(msg, fields...)
	default:
		Logger.Info(msg, fields...)
	}
}
