package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"oneops/backend/models"
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

	// 解析操作动作
	action := m.getActionFromMethod(c.Request.Method)

	// 解析操作描述
	description := m.getDescriptionFromPathAndMethod(c.Request.URL.Path, c.Request.Method)

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

// getModuleFromPath 从路径解析模块名称（从menu表查询）
func (m *AuditMiddleware) getModuleFromPath(path string) string {
	// 如果不是API路径，直接返回
	if !strings.HasPrefix(path, "/api/") {
		return "其他"
	}

	// 去掉/api前缀，得到菜单路径
	menuPath := strings.TrimPrefix(path, "/api")

	// 查询该路径对应的菜单项
	var menu models.Menu
	err := services.GetDB().Where("path = ?", menuPath).First(&menu).Error
	if err != nil {
		// 如果找不到对应菜单，尝试从operation_logs学习
		var moduleName string
		learnErr := services.GetDB().Table("operation_logs").
			Select("DISTINCT module").
			Where("path = ?", path).
			Where("module != ?", "").
			Order("id DESC").
			Limit(1).
			Pluck("module", &moduleName).Error

		if learnErr == nil && moduleName != "" {
			return moduleName
		}

		// 都找不到，返回默认值
		return "未知模块"
	}

	// 如果是顶级菜单（没有父菜单），直接返回菜单名称作为模块名
	if menu.ParentID == 0 {
		return menu.Name
	}

	// 如果有父菜单，查询父菜单获取模块名
	var parentMenu models.Menu
	err = services.GetDB().Where("id = ?", menu.ParentID).First(&parentMenu).Error
	if err != nil {
		return "未知模块"
	}

	return parentMenu.Name
}

// getActionFromMethod 从HTTP方法解析操作动作
func (m *AuditMiddleware) getActionFromMethod(method string) string {
	switch method {
	case "GET":
		return "查询"
	case "POST":
		return "新增"
	case "PUT":
		return "修改"
	case "DELETE":
		return "删除"
	default:
		return method
	}
}

// getDescriptionFromPathAndMethod 从路径和方法生成操作描述
func (m *AuditMiddleware) getDescriptionFromPathAndMethod(path, method string) string {
	// 根据路径和方法生成更详细的描述
	if strings.Contains(path, "/users") {
		return m.getActionFromMethod(method) + "用户"
	} else if strings.Contains(path, "/roles") {
		return m.getActionFromMethod(method) + "角色"
	} else if strings.Contains(path, "/menus") {
		return m.getActionFromMethod(method) + "菜单"
	} else if strings.Contains(path, "/servers") {
		return m.getActionFromMethod(method) + "服务器"
	} else if strings.Contains(path, "/tasks") {
		return m.getActionFromMethod(method) + "任务"
	}
	return m.getActionFromMethod(method) + "操作"
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
