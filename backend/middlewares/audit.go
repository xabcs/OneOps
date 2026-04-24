package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"oneops/backend/services"
	"oneops/backend/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计中间件
type AuditMiddleware struct {
	auditService *services.AuditService
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware() *AuditMiddleware {
	return &AuditMiddleware{
		auditService: services.NewAuditService(),
	}
}

// OperationLog 操作日志记录中间件
func (m *AuditMiddleware) OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过静态文件和健康检查等不需要审计的请求
		if m.shouldSkipAudit(c) {
			c.Next()
			return
		}

		// 记录请求开始时间
		startTime := time.Now()

		// 复制请求体以便后续读取
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器来捕获响应
		writer := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 计算请求耗时
		duration := int(time.Since(startTime).Milliseconds())

		// 获取用户信息
		userID, username, nickname := m.getUserInfo(c)

		// 记录操作日志
		m.recordOperationLog(c, userID, username, nickname, requestBody, writer.body.Bytes(), duration)
	}
}

// shouldSkipAudit 判断是否跳过审计
func (m *AuditMiddleware) shouldSkipAudit(c *gin.Context) bool {
	path := c.Request.URL.Path

	// 跳过静态文件请求
	if strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/favicon.ico") {
		return true
	}

	// 跳过健康检查和其他不需要审计的API
	if path == "/api/health" || path == "/api/metrics" {
		return true
	}

	// 跳过登录请求（有专门的登录日志记录）
	if path == "/api/login" {
		return true
	}

	return false
}

// getUserInfo 从上下文中获取用户信息
func (m *AuditMiddleware) getUserInfo(c *gin.Context) (uint, string, string) {
	// 优先从JWT claims中获取用户信息
	if claims, exists := c.Get("claims"); exists {
		if userClaims, ok := claims.(*utils.Claims); ok {
			return userClaims.UserID, userClaims.Username, ""
		}
	}

	// 备用方案：从单独的user_id和username字段获取
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint); ok {
			if username, exists := c.Get("username"); exists {
				if uname, ok := username.(string); ok {
					return uid, uname, ""
				}
			}
			return uid, "", ""
		}
	}

	return 0, "", ""
}

// recordOperationLog 记录操作日志
func (m *AuditMiddleware) recordOperationLog(c *gin.Context, userID uint, username, nickname string, requestBody, responseBody []byte, duration int) {
	// 解析模块名称
	module := m.getModuleFromPath(c.Request.URL.Path)

	// 解析操作动作（使用智能识别方法）
	action := m.getActionFromMethodAndPath(c.Request.Method, c.Request.URL.Path)

	// 生成操作描述
	description := m.generateDescription(module, action, c.Request.URL.Path)

	// 解析请求参数
	var params interface{}
	if len(requestBody) > 0 {
		json.Unmarshal(requestBody, &params)
	}

	// 解析响应数据
	var response interface{}
	if len(responseBody) > 0 && c.Writer.Status() < 400 {
		json.Unmarshal(responseBody, &response)
	}

	// 获取状态和错误信息
	status := "success"
	errorMsg := ""
	if c.Writer.Status() >= 400 {
		status = "failed"
		// 尝试从响应中提取错误信息
		var errorResponse struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(responseBody, &errorResponse) == nil && errorResponse.Message != "" {
			errorMsg = errorResponse.Message
		}
	}

	// 记录日志
	m.auditService.LogOperation(
		userID,
		username,
		nickname,
		module,
		action,
		description,
		c.Request.Method,
		c.Request.URL.Path,
		params,
		response,
		c.Writer.Status(),
		c.ClientIP(),
		c.Request.UserAgent(),
		duration,
		status,
		errorMsg,
	)
}

// getModuleFromPath 从路径解析模块名称（智能识别）
func (m *AuditMiddleware) getModuleFromPath(path string) string {
	// 如果不是API路径，直接返回
	if !strings.HasPrefix(path, "/api/") {
		return "其他"
	}

	// 去掉/api前缀
	apiPath := strings.TrimPrefix(path, "/api")

	// 基于 API 路径前缀智能识别模块
	if strings.HasPrefix(apiPath, "/login") || strings.HasPrefix(apiPath, "/logout") || strings.HasPrefix(apiPath, "/user") {
		return "认证管理"
	}
	if strings.HasPrefix(apiPath, "/system/menus") {
		return "菜单管理"
	}
	if strings.HasPrefix(apiPath, "/system/roles") {
		return "角色管理"
	}
	if strings.HasPrefix(apiPath, "/system/users") {
		return "用户管理"
	}
	if strings.HasPrefix(apiPath, "/audit") {
		return "审计管理"
	}
	if strings.HasPrefix(apiPath, "/monitoring") {
		return "监控中心"
	}
	if strings.HasPrefix(apiPath, "/tasks") {
		return "任务管理"
	}
	if strings.HasPrefix(apiPath, "/servers") {
		return "服务器管理"
	}
	if strings.HasPrefix(apiPath, "/containers") {
		return "容器管理"
	}
	if strings.HasPrefix(apiPath, "/certificates") {
		return "证书管理"
	}
	if strings.HasPrefix(apiPath, "/system") {
		return "系统管理"
	}

	// 从路径中提取第一级作为模块名
	parts := strings.Split(strings.Trim(apiPath, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return "其他"
}

// getActionFromMethod 从HTTP方法解析操作动作
func (m *AuditMiddleware) getActionFromMethod(method string) string {
	switch method {
	case "GET":
		return "查询"
	case "POST":
		return "新增"
	case "PUT":
		return "更新"
	case "DELETE":
		return "删除"
	case "PATCH":
		return "修改"
	default:
		return method
	}
}

// getActionFromMethodAndPath 从方法和路径智能识别操作
func (m *AuditMiddleware) getActionFromMethodAndPath(method, path string) string {
	// 特殊路径优先处理
	if strings.Contains(path, "/refresh") {
		return "刷新"
	}
	if strings.Contains(path, "/handle") {
		return "处理"
	}
	if strings.Contains(path, "/ignore") {
		return "忽略"
	}
	if strings.Contains(path, "/export") {
		return "导出"
	}
	if strings.Contains(path, "/import") {
		return "导入"
	}
	if strings.Contains(path, "/visit") {
		return "访问"
	}
	if strings.Contains(path, "/statistics") || strings.Contains(path, "/stats") {
		return "统计"
	}

	// 基于 HTTP 方法
	switch method {
	case "GET":
		// 判断是查询还是访问
		if strings.Contains(path, "/detail") || strings.Contains(path, "/info") {
			return "查询"
		}
		return "查询"
	case "POST":
		// 判断是创建还是其他操作
		if strings.HasSuffix(path, "/login") {
			return "登录"
		}
		if strings.HasSuffix(path, "/logout") {
			return "登出"
		}
		return "新增"
	case "PUT":
		return "更新"
	case "DELETE":
		return "删除"
	case "PATCH":
		return "修改"
	default:
		return method
	}
}

// generateDescription 生成操作描述
func (m *AuditMiddleware) generateDescription(module, action, path string) string {
	// 特殊处理
	if action == "刷新" {
		return fmt.Sprintf("刷新%s数据", module)
	}
	if action == "访问" {
		return fmt.Sprintf("访问%s", module)
	}
	if action == "处理" && strings.Contains(path, "alert") {
		return "处理告警"
	}
	if action == "统计" {
		return fmt.Sprintf("查询%s统计数据", module)
	}

	// 根据模块和动作生成描述
	switch module {
	case "监控中心":
		if action == "查询" {
			return "查询监控数据"
		}
		return fmt.Sprintf("%s监控", action)
	case "用户管理":
		return fmt.Sprintf("%s用户", action)
	case "角色管理":
		return fmt.Sprintf("%s角色", action)
	case "菜单管理":
		return fmt.Sprintf("%s菜单", action)
	case "审计管理":
		return fmt.Sprintf("%s审计日志", action)
	case "任务管理":
		return fmt.Sprintf("%s任务", action)
	case "服务器管理":
		return fmt.Sprintf("%s服务器", action)
	case "容器管理":
		return fmt.Sprintf("%s容器", action)
	case "证书管理":
		return fmt.Sprintf("%s证书", action)
	case "认证管理":
		if action == "登录" {
			return "用户登录"
		}
		if action == "登出" {
			return "用户登出"
		}
		return fmt.Sprintf("%s认证", action)
	default:
		// 通用格式
		return fmt.Sprintf("%s%s", action, module)
	}
}

// getDescriptionFromPathAndMethod 从路径和方法生成操作描述（已废弃，保留用于兼容）
func (m *AuditMiddleware) getDescriptionFromPathAndMethod(path, method string) string {
	// 委托给新的 generateDescription 方法
	module := m.getModuleFromPath(path)
	action := m.getActionFromMethodAndPath(method, path)
	return m.generateDescription(module, action, path)
}

// responseWriter 自定义响应写入器用于捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
