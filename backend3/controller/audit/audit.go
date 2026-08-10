package audit

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/audit"

	"github.com/gin-gonic/gin"
)

// AuditController 审计控制器
type AuditController struct {
	svc *AuditService
}

// NewAuditController 创建审计控制器
func NewAuditController(svc *AuditService) *AuditController {
	return &AuditController{svc: svc}
}

// GetLoginLogs 获取登录日志列表
func (ctrl *AuditController) GetLoginLogs(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	query := make(map[string]interface{})
	if v := c.Query("username"); v != "" {
		query["username"] = v
	}
	if v := c.Query("status"); v != "" {
		query["status"] = v
	}
	if v := c.Query("location"); v != "" {
		query["location"] = v
	}
	if v := c.Query("startTime"); v != "" {
		query["startTime"] = v
	}
	if v := c.Query("endTime"); v != "" {
		query["endTime"] = v
	}

	logs, total, err := ctrl.svc.GetLoginLogs(query, params.GetPage(), params.GetPageSize())
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取登录日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(logs, total, params)))
}

// GetOperationLogs 获取操作日志列表
func (ctrl *AuditController) GetOperationLogs(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	query := make(map[string]interface{})
	if v := c.Query("username"); v != "" {
		query["username"] = v
	}
	if v := c.Query("module"); v != "" {
		query["module"] = v
	}
	if v := c.Query("status"); v != "" {
		query["status"] = v
	}
	if v := c.Query("action"); v != "" {
		query["action"] = v
	}
	if v := c.Query("method"); v != "" {
		query["method"] = v
	}
	if v := c.Query("statusCode"); v != "" {
		query["statusCode"] = v
	}
	if v := c.Query("path"); v != "" {
		query["path"] = v
	}
	if v := c.Query("durationRange"); v != "" {
		query["durationRange"] = v
	}
	if v := c.Query("startTime"); v != "" {
		query["startTime"] = v
	}
	if v := c.Query("endTime"); v != "" {
		query["endTime"] = v
	}

	logs, total, err := ctrl.svc.GetOperationLogs(query, params.GetPage(), params.GetPageSize())
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取操作日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(logs, total, params)))
}

// GetSystemEventLogs 获取系统事件日志列表
func (ctrl *AuditController) GetSystemEventLogs(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	query := make(map[string]interface{})
	if v := c.Query("level"); v != "" {
		query["level"] = v
	}
	if v := c.Query("source"); v != "" {
		query["source"] = v
	}
	if v := c.Query("category"); v != "" {
		query["category"] = v
	}
	if v := c.Query("startTime"); v != "" {
		query["startTime"] = v
	}
	if v := c.Query("endTime"); v != "" {
		query["endTime"] = v
	}

	logs, total, err := ctrl.svc.GetSystemEventLogs(query, params.GetPage(), params.GetPageSize())
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取系统事件日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(logs, total, params)))
}

// GetAuditStats 获取审计统计信息
func (ctrl *AuditController) GetAuditStats(c *gin.Context) {
	stats, err := ctrl.svc.GetAuditStats()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取审计统计信息失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// GetModules 获取可用的审计模块列表
func (ctrl *AuditController) GetModules(c *gin.Context) {
	modules := ctrl.svc.GetModules()
	c.JSON(http.StatusOK, utils.SuccessWithData(modules))
}

// ExportLoginLogs 导出登录日志
func (ctrl *AuditController) ExportLoginLogs(c *gin.Context) {
	query := make(map[string]interface{})
	if v := c.Query("username"); v != "" {
		query["username"] = v
	}
	if v := c.Query("status"); v != "" {
		query["status"] = v
	}
	if v := c.Query("startTime"); v != "" {
		query["startTime"] = v
	}
	if v := c.Query("endTime"); v != "" {
		query["endTime"] = v
	}

	logs, _, err := ctrl.svc.GetLoginLogsForExport(query, 1, 10000)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("导出登录日志失败"))
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=login_logs.csv")

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
	query := make(map[string]interface{})
	if v := c.Query("username"); v != "" {
		query["username"] = v
	}
	if v := c.Query("module"); v != "" {
		query["module"] = v
	}
	if v := c.Query("status"); v != "" {
		query["status"] = v
	}
	if v := c.Query("startTime"); v != "" {
		query["startTime"] = v
	}
	if v := c.Query("endTime"); v != "" {
		query["endTime"] = v
	}

	logs, _, err := ctrl.svc.GetOperationLogs(query, 1, 10000)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("导出操作日志失败"))
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=operation_logs.csv")

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
