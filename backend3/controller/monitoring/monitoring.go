package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/monitoring"

	"github.com/gin-gonic/gin"
)

// MonitoringController 监控控制器
type MonitoringController struct {
	svc *MonitoringService
}

// NewMonitoringController 创建监控控制器
func NewMonitoringController(svc *MonitoringService) *MonitoringController {
	return &MonitoringController{svc: svc}
}

// GetOverview 获取监控概览
func (ctrl *MonitoringController) GetOverview(c *gin.Context) {
	overview, err := ctrl.svc.GetOverview()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取监控概览失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(overview))
}

// GetServerExtendedMetrics 获取主机扩展指标
func (ctrl *MonitoringController) GetServerExtendedMetrics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	metrics, err := ctrl.svc.GetServerExtendedMetrics(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(metrics))
}

// GetServerProcesses 获取主机进程信息
func (ctrl *MonitoringController) GetServerProcesses(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	processes, err := ctrl.svc.GetServerProcesses(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(processes))
}

// GetServerServices 获取主机服务状态
func (ctrl *MonitoringController) GetServerServices(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	services, err := ctrl.svc.GetServerServices(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(services))
}

// GetServerHardware 获取主机硬件信息
func (ctrl *MonitoringController) GetServerHardware(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	hardware, err := ctrl.svc.GetServerHardware(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(hardware))
}

// GetServerNetwork 获取主机网络配置
func (ctrl *MonitoringController) GetServerNetwork(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	network, err := ctrl.svc.GetServerNetwork(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(network))
}

// GetServerSecurity 获取主机安全信息
func (ctrl *MonitoringController) GetServerSecurity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	security, err := ctrl.svc.GetServerSecurity(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(security))
}

// GetServerMetricsHistory 查询主机历史指标
func (ctrl *MonitoringController) GetServerMetricsHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的主机ID"))
		return
	}

	metricType := c.Query("metricType")
	startTimeStr := c.Query("startTime")
	endTimeStr := c.Query("endTime")
	interval := c.DefaultQuery("interval", "5m")

	if metricType == "" || startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("缺少必要参数"))
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("开始时间格式错误"))
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("结束时间格式错误"))
		return
	}

	history, err := ctrl.svc.GetMetricsHistory(uint(id), metricType, startTime, endTime, interval)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(history))
}

// === 告警管理 ===

// GetAlerts 查询告警列表
func (ctrl *MonitoringController) GetAlerts(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	queryParams := AlertQueryParams{
		ServerID:     c.Query("serverId"),
		Level:        c.Query("level"),
		Acknowledged: c.Query("acknowledged"),
		Page:         params.GetPage(),
		PageSize:     params.GetPageSize(),
	}

	result, err := ctrl.svc.GetAlerts(queryParams)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(result.Items, result.Total, params)))
}

// AcknowledgeAlert 确认告警
func (ctrl *MonitoringController) AcknowledgeAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的告警ID"))
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&req)

	username := c.GetString("username")
	if username == "" {
		username = "system"
	}

	if err := ctrl.svc.AcknowledgeAlert(uint64(id), username, req.Comment); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("告警确认成功"))
}

// GetAlertStats 获取告警统计
func (ctrl *MonitoringController) GetAlertStats(c *gin.Context) {
	stats, err := ctrl.svc.GetAlertStats()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// === 告警规则管理 ===

// GetAlertRules 获取告警规则列表
func (ctrl *MonitoringController) GetAlertRules(c *gin.Context) {
	ruleSvc := NewAlertRuleService()
	rules, err := ruleSvc.GetAlertRules()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(rules))
}

// CreateAlertRule 创建告警规则
func (ctrl *MonitoringController) CreateAlertRule(c *gin.Context) {
	var rule AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	ruleSvc := NewAlertRuleService()
	if err := ruleSvc.CreateAlertRule(&rule); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(rule))
}

// UpdateAlertRule 更新告警规则
func (ctrl *MonitoringController) UpdateAlertRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("缺少规则ID"))
		return
	}

	var rule AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	ruleSvc := NewAlertRuleService()
	if err := ruleSvc.UpdateAlertRule(id, &rule); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// DeleteAlertRule 删除告警规则
func (ctrl *MonitoringController) DeleteAlertRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("缺少规则ID"))
		return
	}

	ruleSvc := NewAlertRuleService()
	if err := ruleSvc.DeleteAlertRule(id); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// UpdateAlertRuleStatus 更新告警规则状态
func (ctrl *MonitoringController) UpdateAlertRuleStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("缺少规则ID"))
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	ruleSvc := NewAlertRuleService()
	if err := ruleSvc.UpdateAlertRuleStatus(id, req.Enabled); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 通知渠道管理 ===

// GetNotificationChannels 获取通知渠道列表
func (ctrl *MonitoringController) GetNotificationChannels(c *gin.Context) {
	channels, err := ctrl.svc.GetNotificationChannels()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(channels))
}

// CreateNotificationChannel 创建通知渠道
func (ctrl *MonitoringController) CreateNotificationChannel(c *gin.Context) {
	var channel NotificationChannel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	id, err := ctrl.svc.CreateNotificationChannel(&channel)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{"id": id}))
}

// UpdateNotificationChannel 更新通知渠道
func (ctrl *MonitoringController) UpdateNotificationChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的渠道ID"))
		return
	}

	var channel NotificationChannel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	channel.ID = uint(id)

	if err := ctrl.svc.UpdateNotificationChannel(&channel); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// DeleteNotificationChannel 删除通知渠道
func (ctrl *MonitoringController) DeleteNotificationChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的渠道ID"))
		return
	}

	if err := ctrl.svc.DeleteNotificationChannel(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// TestNotificationChannel 测试通知渠道
func (ctrl *MonitoringController) TestNotificationChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的渠道ID"))
		return
	}

	if err := ctrl.svc.TestNotificationChannel(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("测试通知已发送"))
}

// === 巡检报告管理 ===

// GetReports 获取巡检报告列表
func (ctrl *MonitoringController) GetReports(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	queryParams := ReportQueryParams{
		ReportType: c.Query("reportType"),
		Status:     c.Query("status"),
		Page:       params.GetPage(),
		PageSize:   params.GetPageSize(),
	}

	result, err := ctrl.svc.GetReports(queryParams)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(result.Items, result.Total, params)))
}

// CreateReport 创建巡检报告
func (ctrl *MonitoringController) CreateReport(c *gin.Context) {
	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	createdBy := c.GetString("username")
	if createdBy == "" {
		createdBy = "system"
	}

	id, err := ctrl.svc.CreateReport(&req, createdBy)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{"id": id}))
}

// GetReportDetail 获取巡检报告详情
func (ctrl *MonitoringController) GetReportDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的报告ID"))
		return
	}

	detail, err := ctrl.svc.GetReportDetail(uint64(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(detail))
}

// ExportReport 导出巡检报告
func (ctrl *MonitoringController) ExportReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的报告ID"))
		return
	}

	format := c.DefaultQuery("format", "pdf")

	filePath, err := ctrl.svc.ExportReport(uint64(id), format)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{"filePath": filePath}))
}

// DeleteReport 删除巡检报告
func (ctrl *MonitoringController) DeleteReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的报告ID"))
		return
	}

	if err := ctrl.svc.DeleteReport(uint64(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}
