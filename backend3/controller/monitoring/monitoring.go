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

// GetOverview godoc
// @Summary      获取监控概览
// @Description  获取监控中心概览数据，包括主机、告警、资源使用等汇总信息
// @Tags         监控-概览
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}  "监控概览数据"
// @Failure      200  {object}  utils.Response  "获取监控概览失败"
// @Router       /monitoring/overview [get]
// @Security     BearerAuth
func (ctrl *MonitoringController) GetOverview(c *gin.Context) {
	overview, err := ctrl.svc.GetOverview()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取监控概览失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(overview))
}

// GetServerExtendedMetrics godoc
// @Summary      获取主机扩展指标
// @Description  返回指定主机的扩展监控指标（注意：该 handler 尚未注册路由）
// @Tags         监控-主机
// @Produce      json
// @Param        id  path  int  true  "主机 ID"
// @Success      200  {object}  utils.Response{data=object}  "扩展指标数据"
// @Failure      200  {object}  utils.Response  "无效的主机ID / 获取失败"
// @Router       /monitoring/servers/{id}/extended-metrics [get]
// @Security     BearerAuth
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

// GetServerProcesses godoc
// @Summary      获取主机进程信息
// @Description  返回指定主机运行中的进程列表（注意：该 handler 尚未注册路由）
// @Tags         监控-主机
// @Produce      json
// @Param        id  path  int  true  "主机 ID"
// @Success      200  {object}  utils.Response{data=[]object}  "进程列表"
// @Failure      200  {object}  utils.Response  "无效的主机ID / 获取失败"
// @Router       /monitoring/servers/{id}/processes [get]
// @Security     BearerAuth
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

// GetServerServices godoc
// @Summary      获取主机服务状态
// @Description  返回指定主机的系统服务运行状态（注意：该 handler 尚未注册路由）
// @Tags         监控-主机
// @Produce      json
// @Param        id  path  int  true  "主机 ID"
// @Success      200  {object}  utils.Response{data=[]object}  "服务状态列表"
// @Failure      200  {object}  utils.Response  "无效的主机ID / 获取失败"
// @Router       /monitoring/servers/{id}/services [get]
// @Security     BearerAuth
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

// GetServerHardware godoc
// @Summary      获取主机硬件信息
// @Description  返回指定主机的硬件资产信息（CPU/内存/磁盘等）（注意：该 handler 尚未注册路由）
// @Tags         监控-主机
// @Produce      json
// @Param        id  path  int  true  "主机 ID"
// @Success      200  {object}  utils.Response{data=object}  "硬件信息"
// @Failure      200  {object}  utils.Response  "无效的主机ID / 获取失败"
// @Router       /monitoring/servers/{id}/hardware [get]
// @Security     BearerAuth
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

// GetServerNetwork godoc
// @Summary      获取主机网络配置
// @Description  返回指定主机的网络接口与配置信息（注意：该 handler 尚未注册路由）
// @Tags         监控-主机
// @Produce      json
// @Param        id  path  int  true  "主机 ID"
// @Success      200  {object}  utils.Response{data=object}  "网络配置"
// @Failure      200  {object}  utils.Response  "无效的主机ID / 获取失败"
// @Router       /monitoring/servers/{id}/network [get]
// @Security     BearerAuth
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

// GetServerSecurity godoc
// @Summary      获取主机安全信息
// @Description  返回指定主机的安全状况汇总（防火墙/登录记录/补丁等）（注意：该 handler 尚未注册路由）
// @Tags         监控-主机
// @Produce      json
// @Param        id  path  int  true  "主机 ID"
// @Success      200  {object}  utils.Response{data=object}  "安全信息"
// @Failure      200  {object}  utils.Response  "无效的主机ID / 获取失败"
// @Router       /monitoring/servers/{id}/security [get]
// @Security     BearerAuth
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

// GetServerMetricsHistory godoc
// @Summary      查询主机历史指标
// @Description  按指标类型与时间范围查询指定主机的历史指标序列（注意：该 handler 尚未注册路由）
// @Tags         监控-主机
// @Produce      json
// @Param        id          path      string  true   "主机 ID"
// @Param        metricType  query     string  true   "指标类型"
// @Param        startTime   query     string  true   "开始时间（RFC3339）"
// @Param        endTime     query     string  true   "结束时间（RFC3339）"
// @Param        interval    query     string  false  "聚合间隔"  default(5m)
// @Success      200  {object}  utils.Response{data=object}  "历史指标序列"
// @Failure      200  {object}  utils.Response  "无效的主机ID / 缺少必要参数 / 时间格式错误 / 查询失败"
// @Router       /monitoring/servers/{id}/metrics-history [get]
// @Security     BearerAuth
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

// GetAlerts godoc
// @Summary      获取告警列表
// @Description  分页获取告警列表，支持按服务器、级别、是否确认筛选
// @Tags         监控-告警
// @Produce      json
// @Param        page         query     int     false  "页码"               default(1)
// @Param        pageSize     query     int     false  "每页数量"            default(20)
// @Param        serverId     query     string  false  "服务器 ID"
// @Param        level        query     string  false  "告警级别(critical/warning/info)"
// @Param        acknowledged query     string  false  "是否已确认(true/false)"
// @Success      200  {object}  utils.Response{data=dto.PageResult}  "告警列表"
// @Failure      200  {object}  utils.Response  "获取告警列表失败"
// @Router       /monitoring/alerts [get]
// @Security     BearerAuth
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

// AcknowledgeAlert godoc
// @Summary      确认告警
// @Description  根据告警 ID 确认指定告警，可附带确认备注
// @Tags         监控-告警
// @Accept       json
// @Produce      json
// @Param        id     path      int     true  "告警 ID"
// @Param        body   body      object  true  "确认信息"  examples({"comment":"已处理"})
// @Success      200  {object}  utils.Response  "告警确认成功"
// @Failure      200  {object}  utils.Response  "无效的告警ID 或 确认失败"
// @Router       /monitoring/alerts/{id}/acknowledge [post]
// @Security     BearerAuth
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

// GetAlertStats godoc
// @Summary      获取告警统计
// @Description  获取告警按级别、状态等维度的汇总统计信息
// @Tags         监控-告警
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}  "告警统计信息"
// @Failure      200  {object}  utils.Response  "获取告警统计失败"
// @Router       /monitoring/alerts/stats [get]
// @Security     BearerAuth
func (ctrl *MonitoringController) GetAlertStats(c *gin.Context) {
	stats, err := ctrl.svc.GetAlertStats()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// === 告警规则管理 ===

// GetAlertRules godoc
// @Summary      获取告警规则列表
// @Description  获取全部告警规则列表
// @Tags         监控-告警规则
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]AlertRule}  "告警规则列表"
// @Failure      200  {object}  utils.Response  "获取告警规则失败"
// @Router       /monitoring/alerts/rules [get]
// @Security     BearerAuth
func (ctrl *MonitoringController) GetAlertRules(c *gin.Context) {
	ruleSvc := NewAlertRuleService()
	rules, err := ruleSvc.GetAlertRules()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(rules))
}

// CreateAlertRule godoc
// @Summary      创建告警规则
// @Description  创建一条新的告警规则
// @Tags         监控-告警规则
// @Accept       json
// @Produce      json
// @Param        body  body      AlertRule  true  "告警规则信息"
// @Success      200  {object}  utils.Response{data=AlertRule}  "创建成功的告警规则"
// @Failure      200  {object}  utils.Response  "参数错误 或 创建失败"
// @Router       /monitoring/alerts/rules [post]
// @Security     BearerAuth
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

// UpdateAlertRule godoc
// @Summary      更新告警规则
// @Description  根据规则 ID 更新指定告警规则
// @Tags         监控-告警规则
// @Accept       json
// @Produce      json
// @Param        id    path      string     true  "规则 ID"
// @Param        body  body      AlertRule  true  "告警规则信息"
// @Success      200  {object}  utils.Response  "更新成功"
// @Failure      200  {object}  utils.Response  "缺少规则ID 或 参数错误 或 更新失败"
// @Router       /monitoring/alerts/rules/{id} [put]
// @Security     BearerAuth
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

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteAlertRule godoc
// @Summary      删除告警规则
// @Description  根据规则 ID 删除指定告警规则
// @Tags         监控-告警规则
// @Produce      json
// @Param        id  path      string  true  "规则 ID"
// @Success      200  {object}  utils.Response  "删除成功"
// @Failure      200  {object}  utils.Response  "缺少规则ID 或 删除失败"
// @Router       /monitoring/alerts/rules/{id} [delete]
// @Security     BearerAuth
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

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// UpdateAlertRuleStatus godoc
// @Summary      更新告警规则状态
// @Description  根据规则 ID 启用或禁用指定告警规则
// @Tags         监控-告警规则
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "规则 ID"
// @Param        body  body      object  true  "启用状态"  examples({"enabled":true})
// @Success      200  {object}  utils.Response  "状态更新成功"
// @Failure      200  {object}  utils.Response  "缺少规则ID 或 参数错误 或 更新失败"
// @Router       /monitoring/alerts/rules/{id}/status [put]
// @Security     BearerAuth
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

	c.JSON(http.StatusOK, utils.SuccessWithMessage("状态更新成功"))
}

// === 通知渠道管理 ===

// GetNotificationChannels godoc
// @Summary      获取通知渠道列表
// @Description  获取全部通知渠道列表
// @Tags         监控-通知渠道
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]NotificationChannel}  "通知渠道列表"
// @Failure      200  {object}  utils.Response  "获取通知渠道失败"
// @Router       /monitoring/notifications/channels [get]
// @Security     BearerAuth
func (ctrl *MonitoringController) GetNotificationChannels(c *gin.Context) {
	channels, err := ctrl.svc.GetNotificationChannels()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(channels))
}

// CreateNotificationChannel godoc
// @Summary      创建通知渠道
// @Description  创建一条新的通知渠道
// @Tags         监控-通知渠道
// @Accept       json
// @Produce      json
// @Param        body  body      NotificationChannel  true  "通知渠道信息"
// @Success      200  {object}  utils.Response{data=object}  "创建成功，返回新渠道 ID"
// @Failure      200  {object}  utils.Response  "参数错误 或 创建失败"
// @Router       /monitoring/notifications/channels [post]
// @Security     BearerAuth
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

// UpdateNotificationChannel godoc
// @Summary      更新通知渠道
// @Description  根据渠道 ID 更新指定通知渠道
// @Tags         监控-通知渠道
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "渠道 ID"
// @Param        body  body      NotificationChannel  true  "通知渠道信息"
// @Success      200  {object}  utils.Response  "更新成功"
// @Failure      200  {object}  utils.Response  "无效的渠道ID 或 参数错误 或 更新失败"
// @Router       /monitoring/notifications/channels/{id} [put]
// @Security     BearerAuth
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

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteNotificationChannel godoc
// @Summary      删除通知渠道
// @Description  根据渠道 ID 删除指定通知渠道
// @Tags         监控-通知渠道
// @Produce      json
// @Param        id  path      int  true  "渠道 ID"
// @Success      200  {object}  utils.Response  "删除成功"
// @Failure      200  {object}  utils.Response  "无效的渠道ID 或 删除失败"
// @Router       /monitoring/notifications/channels/{id} [delete]
// @Security     BearerAuth
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

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// TestNotificationChannel godoc
// @Summary      测试通知渠道
// @Description  根据渠道 ID 向指定通知渠道发送一条测试通知
// @Tags         监控-通知渠道
// @Produce      json
// @Param        id  path      int  true  "渠道 ID"
// @Success      200  {object}  utils.Response  "测试通知已发送"
// @Failure      200  {object}  utils.Response  "无效的渠道ID 或 测试失败"
// @Router       /monitoring/notifications/channels/{id}/test [post]
// @Security     BearerAuth
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

// GetReports godoc
// @Summary      获取巡检报告列表
// @Description  分页获取巡检报告列表，支持按报告类型、状态筛选
// @Tags         监控-报表
// @Produce      json
// @Param        page        query     int     false  "页码"    default(1)
// @Param        pageSize    query     int     false  "每页数量" default(20)
// @Param        reportType  query     string  false  "报告类型"
// @Param        status      query     string  false  "报告状态"
// @Success      200  {object}  utils.Response{data=dto.PageResult}  "巡检报告列表"
// @Failure      200  {object}  utils.Response  "获取巡检报告列表失败"
// @Router       /monitoring/reports [get]
// @Security     BearerAuth
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

// CreateReport godoc
// @Summary      创建巡检报告
// @Description  创建一条新的巡检报告，根据请求参数生成对应类型的报告
// @Tags         监控-报表
// @Accept       json
// @Produce      json
// @Param        body  body      CreateReportRequest  true  "巡检报告请求信息"
// @Success      200  {object}  utils.Response{data=object}  "创建成功，返回新报告 ID"
// @Failure      200  {object}  utils.Response  "参数错误 或 创建失败"
// @Router       /monitoring/reports [post]
// @Security     BearerAuth
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

// GetReportDetail godoc
// @Summary      获取巡检报告详情
// @Description  根据报告 ID 获取巡检报告详情
// @Tags         监控-报表
// @Produce      json
// @Param        id  path      int  true  "报告 ID"
// @Success      200  {object}  utils.Response{data=ReportDetail}  "巡检报告详情"
// @Failure      200  {object}  utils.Response  "无效的报告ID 或 获取详情失败"
// @Router       /monitoring/reports/{id} [get]
// @Security     BearerAuth
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

// ExportReport godoc
// @Summary      导出巡检报告
// @Description  根据报告 ID 导出指定格式的巡检报告，返回生成的文件路径
// @Tags         监控-报表
// @Produce      json
// @Param        id      path      int     true  "报告 ID"
// @Param        format  query     string  false  "导出格式(pdf/html/csv)"  default(pdf)
// @Success      200  {object}  utils.Response{data=object}  "导出成功，返回文件路径"
// @Failure      200  {object}  utils.Response  "无效的报告ID 或 导出失败"
// @Router       /monitoring/reports/{id}/export [get]
// @Security     BearerAuth
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

// DeleteReport godoc
// @Summary      删除巡检报告
// @Description  根据报告 ID 删除指定巡检报告
// @Tags         监控-报表
// @Produce      json
// @Param        id  path      int  true  "报告 ID"
// @Success      200  {object}  utils.Response  "删除成功"
// @Failure      200  {object}  utils.Response  "无效的报告ID 或 删除失败"
// @Router       /monitoring/reports/{id} [delete]
// @Security     BearerAuth
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

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
