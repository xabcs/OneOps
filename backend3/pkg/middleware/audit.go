package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	repoaudit "oneops/backend3/repository/audit"
	"oneops/backend3/service/audit"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// maxAuditBodySize 审计捕获请求/响应体的最大字节数，超过则截断，避免大响应成倍内存分配
const maxAuditBodySize = 8 * 1024

// truncatedFlag 截断标记
var truncatedFlag = []byte(`...[TRUNCATED]`)

// sensitiveFieldKeywords 敏感字段关键词（小写，子串匹配）：字段名命中即打码，避免请求体明文落库。
// 子串匹配可覆盖 old_password、access_token、client_secret 等衍生字段名
var sensitiveFieldKeywords = []string{
	"password",
	"passwd",
	"passphrase",
	"token",
	"secret",
	"credential",
	"privatekey",
	"private_key",
	"authorization",
	"apikey",
}

// isSensitiveField 判断字段名是否命中敏感关键词（忽略大小写，子串匹配）
func isSensitiveField(key string) bool {
	lowerKey := strings.ToLower(strings.TrimSpace(key))
	for _, keyword := range sensitiveFieldKeywords {
		if strings.Contains(lowerKey, keyword) {
			return true
		}
	}
	return false
}

// maskSensitiveFields 递归打码嵌套结构（map/slice）中的敏感字段，值替换为 "***"
func maskSensitiveFields(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, item := range val {
			if isSensitiveField(k) {
				val[k] = "***"
			} else {
				val[k] = maskSensitiveFields(item)
			}
		}
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = maskSensitiveFields(item)
		}
		return val
	default:
		return v
	}
}

// sanitizeRequestBody 审计请求体脱敏：JSON 对象递归打码敏感字段；解析失败（非 JSON）按原始文本记录；超过上限截断并追加标记
func sanitizeRequestBody(body []byte) []byte {
	if len(body) == 0 {
		return body
	}
	var parsed map[string]interface{}
	if json.Unmarshal(body, &parsed) == nil {
		if masked, err := json.Marshal(maskSensitiveFields(parsed)); err == nil {
			body = masked
		}
	}
	if len(body) > maxAuditBodySize {
		body = append(body[:maxAuditBodySize:maxAuditBodySize], truncatedFlag...)
	}
	return body
}

// determineAuditStatus 判定审计成败：项目统一封装为业务错误返回 HTTP 200 + {"code":>=400,"success":false}，
// 故优先解析响应体业务码（code>=400 或 success==false 判 failed，errorMsg 取 message），解析失败再回退 HTTP 状态码判定
func determineAuditStatus(statusCode int, responseBody []byte) (status, errorMsg string) {
	status = "success"
	var resp struct {
		Code    *int   `json:"code"`
		Success *bool  `json:"success"`
		Message string `json:"message"`
	}
	if len(responseBody) > 0 && json.Unmarshal(responseBody, &resp) == nil && resp.Code != nil {
		if *resp.Code >= 400 || (resp.Success != nil && !*resp.Success) {
			status = "failed"
			errorMsg = resp.Message
		}
		return
	}
	// 回退：按 HTTP 状态码判定
	if statusCode >= 400 {
		status = "failed"
		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(responseBody, &errResp) == nil && errResp.Message != "" {
			errorMsg = errResp.Message
		}
	}
	return
}

// AuditMiddleware 审计中间件
type AuditMiddleware struct {
	auditService *audit.AuditService
}

// auditLogQueueSize 审计异步写入队列容量：满时直接丢弃并计数告警，不阻塞请求路径
const auditLogQueueSize = 1024

// auditLogChan 审计日志异步写入队列（有界，避免高峰期无限堆积内存）
var auditLogChan = make(chan *auditLogEntry, auditLogQueueSize)

// droppedAuditLogs 队列满时累计丢弃的审计日志条数
var droppedAuditLogs int64

// auditWorkerOnce 保证审计消费协程全局只启动一次
var auditWorkerOnce sync.Once

// auditLogEntry 异步审计日志条目（不持有 gin.Context，可在请求结束后安全使用）
type auditLogEntry struct {
	userID       uint
	username     string
	nickname     string
	method       string
	path         string
	statusCode   int
	ip           string
	userAgent    string
	requestBody  []byte
	responseBody []byte
	duration     int
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware() *AuditMiddleware {
	m := &AuditMiddleware{
		auditService: audit.NewAuditService(repoaudit.NewAuditRepository(database.GetDB())),
	}
	// 启动审计日志消费协程（全局仅一个），经 SafeGo 防 panic 打崩进程
	auditWorkerOnce.Do(func() {
		utils.SafeGo("audit-log-worker", m.consumeAuditLogQueue)
	})
	return m
}

// consumeAuditLogQueue 持续消费审计队列并落库；单条写入 panic 由 SafeRun 兜底，不影响后续消费
func (m *AuditMiddleware) consumeAuditLogQueue() {
	for entry := range auditLogChan {
		utils.SafeRun("audit-log-worker", func() {
			m.recordOperationLog2(entry.userID, entry.username, entry.nickname,
				entry.method, entry.path, entry.statusCode, entry.ip, entry.userAgent,
				entry.requestBody, entry.responseBody, entry.duration)
		})
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

		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()
		respBody := append([]byte(nil), writer.body.Bytes()...)
		// 请求体先脱敏（递归打码 password/token/secret 等敏感字段）再入队，避免明文落库
		reqBody := sanitizeRequestBody(requestBody)

		// 非阻塞写入有界队列，由后台消费协程异步落库，不阻塞响应；
		// 队列满时丢弃并计数告警，避免高峰期阻塞请求或内存无限增长
		entry := &auditLogEntry{
			userID:       userID,
			username:     username,
			nickname:     nickname,
			method:       method,
			path:         path,
			statusCode:   statusCode,
			ip:           ip,
			userAgent:    userAgent,
			requestBody:  reqBody,
			responseBody: respBody,
			duration:     duration,
		}
		select {
		case auditLogChan <- entry:
		default:
			dropped := atomic.AddInt64(&droppedAuditLogs, 1)
			logger.Warn("审计日志队列已满，丢弃本次审计记录",
				zap.Int64("dropped_total", dropped),
				zap.String("method", method),
				zap.String("path", path),
			)
		}
	}
}

// shouldSkipAudit 判断是否跳过审计
func (m *AuditMiddleware) shouldSkipAudit(c *gin.Context) bool {
	path := c.Request.URL.Path

	// 跳过静态文件请求（前缀自带 "/" 边界，不会误伤 /staticXXX 之类路径）
	if strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/assets/") ||
		path == "/favicon.ico" {
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

	// 跳过所有 WebSocket 请求：固定路径精确匹配；
	// cmdb 会话 WebSocket 实际为 /api/cmdb/sessions/{id}/ws，用前缀+后缀联合匹配，
	// 避免误伤 /api/cmdb/sessions/active 等 REST 接口和 /api/monitoring/ws/clients
	if path == "/api/k8s/terminal/ws" ||
		path == "/api/monitoring/ws" ||
		(strings.HasPrefix(path, "/api/cmdb/sessions/") && strings.HasSuffix(path, "/ws")) {
		return true
	}

	// 也跳过带有 Upgrade: websocket 头的请求
	if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
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

// recordOperationLog2 不依赖 gin.Context 的审计写入（用于 goroutine 异步调用）
func (m *AuditMiddleware) recordOperationLog2(userID uint, username, nickname, method, path string,
	statusCode int, ip, userAgent string, requestBody, responseBody []byte, duration int) {

	module := m.getModuleFromPath(path)
	action := m.getActionFromMethodAndPath(method, path)
	description := m.generateDescription(module, action, path)

	var params interface{}
	if len(requestBody) > 0 {
		json.Unmarshal(requestBody, &params)
	}

	var response interface{}
	if len(responseBody) > 0 && statusCode < 400 {
		json.Unmarshal(responseBody, &response)
	}

	// 按响应体业务码判定成败（HTTP 200 + code>=400/success=false 视为失败），
	// body 非 JSON 或解析失败时回退按 HTTP 状态码判定
	status, errorMsg := determineAuditStatus(statusCode, responseBody)

	m.auditService.LogOperation(
		userID, username, nickname,
		module, action, description,
		method, path,
		params, response,
		statusCode, ip, userAgent,
		duration, status, errorMsg,
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

	// K8s 集群管理模块
	if strings.HasPrefix(apiPath, "/k8s") {
		if strings.Contains(path, "/clusters") {
			return "K8s集群管理"
		}
		if strings.Contains(path, "/deployments") {
			return "K8s Deployment管理"
		}
		if strings.Contains(path, "/statefulsets") {
			return "K8s StatefulSet管理"
		}
		if strings.Contains(path, "/daemonsets") {
			return "K8s DaemonSet管理"
		}
		if strings.Contains(path, "/services") {
			return "K8s Service管理"
		}
		if strings.Contains(path, "/pods") {
			return "K8s Pod管理"
		}
		if strings.Contains(path, "/configmaps") {
			return "K8s ConfigMap管理"
		}
		if strings.Contains(path, "/secrets") {
			return "K8s Secret管理"
		}
		if strings.Contains(path, "/terminal") {
			return "K8s终端管理"
		}
		if strings.Contains(path, "/permissions") {
			return "K8s权限管理"
		}
		return "K8s资源管理"
	}

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
		if strings.Contains(path, "/detail") || strings.Contains(path, "/info") {
			return "查询"
		}
		return "查询"
	case "POST":
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

	// K8s 操作的特殊处理
	if strings.HasPrefix(module, "K8s") {
		resourceType := ""
		if strings.Contains(path, "/clusters") {
			resourceType = "集群"
		} else if strings.Contains(path, "/deployments") {
			resourceType = "Deployment"
		} else if strings.Contains(path, "/statefulsets") {
			resourceType = "StatefulSet"
		} else if strings.Contains(path, "/daemonsets") {
			resourceType = "DaemonSet"
		} else if strings.Contains(path, "/services") {
			resourceType = "Service"
		} else if strings.Contains(path, "/pods") {
			resourceType = "Pod"
		} else if strings.Contains(path, "/configmaps") {
			resourceType = "ConfigMap"
		} else if strings.Contains(path, "/secrets") {
			resourceType = "Secret"
		} else if strings.Contains(path, "/terminal") {
			resourceType = "终端"
		}

		operationType := ""
		if strings.Contains(path, "/scale") {
			operationType = "扩缩容"
		} else if strings.Contains(path, "/restart") {
			operationType = "重启"
		} else if strings.Contains(path, "/logs") {
			operationType = "查看日志"
		} else if strings.Contains(path, "/test") {
			operationType = "测试连接"
		} else if strings.Contains(path, "/permissions") {
			operationType = "权限"
		}

		if operationType != "" {
			return fmt.Sprintf("%s%s", operationType, resourceType)
		}

		return fmt.Sprintf("%s%s", action, resourceType)
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
		return fmt.Sprintf("%s%s", action, module)
	}
}

// getDescriptionFromPathAndMethod 从路径和方法生成操作描述（已废弃，保留用于兼容）
func (m *AuditMiddleware) getDescriptionFromPathAndMethod(path, method string) string {
	module := m.getModuleFromPath(path)
	action := m.getActionFromMethodAndPath(method, path)
	return m.generateDescription(module, action, path)
}

// responseWriter 自定义响应写入器用于捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// appendBody 捕获响应体到缓冲，超过 maxAuditBodySize 上限后不再捕获，避免大响应内存放大
func (w *responseWriter) appendBody(b []byte) {
	if w.body.Len() >= maxAuditBodySize {
		return
	}
	if remain := maxAuditBodySize - w.body.Len(); len(b) > remain {
		b = b[:remain]
	}
	w.body.Write(b)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.appendBody(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	if w.body.Len() < maxAuditBodySize {
		captured := s
		if remain := maxAuditBodySize - w.body.Len(); len(captured) > remain {
			captured = captured[:remain]
		}
		w.body.WriteString(captured)
	}
	return w.ResponseWriter.WriteString(s)
}
