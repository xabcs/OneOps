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

// GetLoginLogs godoc
// @Summary      获取登录日志
// @Description  分页获取登录日志，支持按用户名、状态、位置、时间范围筛选
// @Tags         审计日志
// @Produce      json
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(20)
// @Param        username   query     string  false  "用户名"
// @Param        status     query     string  false  "状态(success/failed)"
// @Param        location   query     string  false  "登录位置"
// @Param        startTime  query     string  false  "起始时间"
// @Param        endTime    query     string  false  "结束时间"
// @Success      200  {object}  utils.Response{data=dto.PageResult}  "登录日志列表"
// @Failure      200  {object}  utils.Response  "获取登录日志失败"
// @Router       /audit/login-logs [get]
// @Security     BearerAuth
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

// GetOperationLogs godoc
// @Summary      获取操作日志
// @Description  分页获取操作日志，支持按用户名、模块、状态、动作、HTTP 方法、路径、耗时区间、时间范围筛选
// @Tags         审计日志
// @Produce      json
// @Param           page           query     int     false  "页码"    default(1)
// @Param           pageSize       query     int     false  "每页数量" default(20)
// @Param           username       query     string  false  "用户名"
// @Param           module         query     string  false  "模块"
// @Param           status         query     string  false  "状态"
// @Param           action         query     string  false  "操作动作"
// @Param           method         query     string  false  "HTTP 方法"
// @Param           statusCode     query     string  false  "状态码"
// @Param           path           query     string  false  "请求路径"
// @Param           durationRange  query     string  false  "耗时区间"
// @Param           startTime      query     string  false  "起始时间"
// @Param           endTime        query     string  false  "结束时间"
// @Success      200  {object}  utils.Response{data=dto.PageResult}  "操作日志列表"
// @Failure      200  {object}  utils.Response  "获取操作日志失败"
// @Router       /audit/operation-logs [get]
// @Security     BearerAuth
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

// GetSystemEventLogs godoc
// @Summary      获取系统事件日志
// @Description  分页获取系统事件日志，支持按级别、来源、分类、时间范围筛选
// @Tags         审计日志
// @Produce      json
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(20)
// @Param        level      query     string  false  "日志级别"
// @Param        source     query     string  false  "来源"
// @Param        category   query     string  false  "分类"
// @Param        startTime  query     string  false  "起始时间"
// @Param        endTime    query     string  false  "结束时间"
// @Success      200  {object}  utils.Response{data=dto.PageResult}  "系统事件日志列表"
// @Failure      200  {object}  utils.Response  "获取系统事件日志失败"
// @Router       /audit/system-event-logs [get]
// @Security     BearerAuth
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

// GetAuditStats godoc
// @Summary      获取审计统计
// @Description  返回审计模块的汇总统计信息（登录/操作/事件计数等）
// @Tags         审计日志
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "获取审计统计信息失败"
// @Router       /audit/stats [get]
// @Security     BearerAuth
func (ctrl *AuditController) GetAuditStats(c *gin.Context) {
	stats, err := ctrl.svc.GetAuditStats()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取审计统计信息失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// GetModules godoc
// @Summary      获取审计模块列表
// @Description  返回操作日志筛选可用的业务模块列表
// @Tags         审计日志
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]string}
// @Router       /audit/modules [get]
// @Security     BearerAuth
func (ctrl *AuditController) GetModules(c *gin.Context) {
	modules := ctrl.svc.GetModules()
	c.JSON(http.StatusOK, utils.SuccessWithData(modules))
}

// ExportLoginLogs godoc
// @Summary      导出登录日志
// @Description  按筛选条件导出登录日志为 CSV 文件下载
// @Tags         审计日志
// @Produce      plain
// @Param        username   query     string  false  "用户名"
// @Param        status     query     string  false  "状态"
// @Param        startTime  query     string  false  "起始时间"
// @Param        endTime    query     string  false  "结束时间"
// @Success      200  {file}   binary  "login_logs.csv"
// @Failure      200  {object}  utils.Response  "导出登录日志失败"
// @Router       /audit/login-logs/export [get]
// @Security     BearerAuth
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

// ExportOperationLogs godoc
// @Summary      导出操作日志
// @Description  按筛选条件导出操作日志为 CSV 文件下载
// @Tags         审计日志
// @Produce      plain
// @Param        username   query     string  false  "用户名"
// @Param        module     query     string  false  "模块"
// @Param        status     query     string  false  "状态"
// @Param        startTime  query     string  false  "起始时间"
// @Param        endTime    query     string  false  "结束时间"
// @Success      200  {file}   binary  "operation_logs.csv"
// @Failure      200  {object}  utils.Response  "导出操作日志失败"
// @Router       /audit/operation-logs/export [get]
// @Security     BearerAuth
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
