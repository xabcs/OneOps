package controllers

import (
	"net/http"
	"oneops/backend/services"
	"oneops/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AuditController 审计控制器
type AuditController struct {
	auditService *services.AuditService
}

// NewAuditController 创建审计控制器
func NewAuditController() *AuditController {
	return &AuditController{
		auditService: services.NewAuditService(),
	}
}

// GetLoginLogs 获取登录日志列表
func (ctrl *AuditController) GetLoginLogs(c *gin.Context) {
	// 解析查询参数
	query := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		query["username"] = username
	}
	if status := c.Query("status"); status != "" {
		query["status"] = status
	}
	if location := c.Query("location"); location != "" {
		query["location"] = location
	}
	if startTime := c.Query("startTime"); startTime != "" {
		query["startTime"] = startTime
	}
	if endTime := c.Query("endTime"); endTime != "" {
		query["endTime"] = endTime
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取日志数据
	logs, total, err := ctrl.auditService.GetLoginLogs(query, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取登录日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"pages":    (int(total) + pageSize - 1) / pageSize,
	}))
}

// GetOperationLogs 获取操作日志列表
func (ctrl *AuditController) GetOperationLogs(c *gin.Context) {
	// 解析查询参数
	query := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		query["username"] = username
	}
	if module := c.Query("module"); module != "" {
		query["module"] = module
	}
	if status := c.Query("status"); status != "" {
		query["status"] = status
	}
	if action := c.Query("action"); action != "" {
		query["action"] = action
	}
	if startTime := c.Query("startTime"); startTime != "" {
		query["startTime"] = startTime
	}
	if endTime := c.Query("endTime"); endTime != "" {
		query["endTime"] = endTime
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取日志数据
	logs, total, err := ctrl.auditService.GetOperationLogs(query, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取操作日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"pages":    (int(total) + pageSize - 1) / pageSize,
	}))
}

// GetSystemEventLogs 获取系统事件日志列表
func (ctrl *AuditController) GetSystemEventLogs(c *gin.Context) {
	// 解析查询参数
	query := make(map[string]interface{})
	if level := c.Query("level"); level != "" {
		query["level"] = level
	}
	if source := c.Query("source"); source != "" {
		query["source"] = source
	}
	if category := c.Query("category"); category != "" {
		query["category"] = category
	}
	if startTime := c.Query("startTime"); startTime != "" {
		query["startTime"] = startTime
	}
	if endTime := c.Query("endTime"); endTime != "" {
		query["endTime"] = endTime
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取日志数据
	logs, total, err := ctrl.auditService.GetSystemEventLogs(query, page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取系统事件日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"pages":    (int(total) + pageSize - 1) / pageSize,
	}))
}

// GetAuditStats 获取审计统计信息
func (ctrl *AuditController) GetAuditStats(c *gin.Context) {
	stats, err := ctrl.auditService.GetAuditStats()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取审计统计信息失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// GetModules 获取可用的审计模块列表
func (ctrl *AuditController) GetModules(c *gin.Context) {
	modules := ctrl.auditService.GetModules()
	c.JSON(http.StatusOK, utils.SuccessWithData(modules))
}

// ExportLoginLogs 导出登录日志
func (ctrl *AuditController) ExportLoginLogs(c *gin.Context) {
	// 获取所有符合条件的日志
	query := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		query["username"] = username
	}
	if status := c.Query("status"); status != "" {
		query["status"] = status
	}
	if startTime := c.Query("startTime"); startTime != "" {
		query["startTime"] = startTime
	}
	if endTime := c.Query("endTime"); endTime != "" {
		query["endTime"] = endTime
	}

	logs, _, err := ctrl.auditService.GetLoginLogsForExport(query, 1, 10000) // 最大导出10000条
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("导出登录日志失败"))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=login_logs.csv")

	// 生成CSV内容
	csvContent := "ID,用户名,昵称,IP地址,用户代理,位置,状态,失败原因,登录时间,登出时间,会话时长(秒)\n"
	for _, log := range logs {
		csvContent += strconv.FormatUint(uint64(log.ID), 10) + ","
		csvContent += log.Username + ","
		csvContent += log.Nickname + ","
		csvContent += log.IP + ","
		csvContent += "\"" + log.UserAgent + "\","
		csvContent += log.Location + ","
		csvContent += log.Status + ","
		csvContent += log.FailReason + ","
		csvContent += log.LoginTime.Format("2006-01-02 15:04:05") + ","
		if log.LogoutTime != nil {
			csvContent += log.LogoutTime.Format("2006-01-02 15:04:05") + ","
		} else {
			csvContent += ","
		}
		csvContent += strconv.Itoa(log.Duration) + "\n"
	}

	c.String(http.StatusOK, csvContent)
}

// ExportOperationLogs 导出操作日志
func (ctrl *AuditController) ExportOperationLogs(c *gin.Context) {
	// 获取所有符合条件的日志
	query := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		query["username"] = username
	}
	if module := c.Query("module"); module != "" {
		query["module"] = module
	}
	if status := c.Query("status"); status != "" {
		query["status"] = status
	}
	if startTime := c.Query("startTime"); startTime != "" {
		query["startTime"] = startTime
	}
	if endTime := c.Query("endTime"); endTime != "" {
		query["endTime"] = endTime
	}

	logs, _, err := ctrl.auditService.GetOperationLogs(query, 1, 10000) // 最大导出10000条
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("导出操作日志失败"))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=operation_logs.csv")

	// 生成CSV内容
	csvContent := "ID,用户名,昵称,模块,操作,描述,HTTP方法,路径,状态码,IP地址,用户代理,耗时(ms),状态,错误信息,操作时间\n"
	for _, log := range logs {
		csvContent += strconv.FormatUint(uint64(log.ID), 10) + ","
		csvContent += log.Username + ","
		csvContent += log.Nickname + ","
		csvContent += log.Module + ","
		csvContent += log.Action + ","
		csvContent += "\"" + log.Description + "\","
		csvContent += log.Method + ","
		csvContent += log.Path + ","
		csvContent += strconv.Itoa(log.StatusCode) + ","
		csvContent += log.IP + ","
		csvContent += "\"" + log.UserAgent + "\","
		csvContent += strconv.Itoa(log.Duration) + ","
		csvContent += log.Status + ","
		csvContent += "\"" + log.ErrorMsg + "\","
		csvContent += log.OperateTime.Format("2006-01-02 15:04:05") + "\n"
	}

	c.String(http.StatusOK, csvContent)
}
