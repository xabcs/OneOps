package controllers

import (
	"net/http"
	"oneops/backend/logger"
	"strconv"
	"time"

	"oneops/backend/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MonitoringController 监控控制器
type MonitoringController struct {
	monitoringService *services.MonitoringService
	alertRuleService  *services.AlertRuleService
}

// NewMonitoringController 创建监控控制器
func NewMonitoringController() *MonitoringController {
	return &MonitoringController{
		monitoringService: services.NewMonitoringService(),
		alertRuleService:  services.NewAlertRuleService(),
	}
}

// HandleAlertRequest 处理/忽略告警请求
type HandleAlertRequest struct {
	AlertID uint   `json:"alertId" binding:"required"`
	Action  string `json:"action" binding:"required,oneof=handle ignore"`
	Reason  string `json:"reason"`
}

// GrafanaConfig Grafana配置
type GrafanaConfig struct {
	BaseURL string
	APIKey  string
	OrgID   string
}

// GetGrafanaUrl 获取Grafana面板URL（带认证）
func (c *MonitoringController) GetGrafanaUrl(ctx *gin.Context) {
	// Grafana配置
	config := GrafanaConfig{
		BaseURL: "http://192.168.4.168:3000",
		APIKey:  "", // 如果有API Token，在这里配置
		OrgID:   "1",
	}

	// 面板配置
	dashboardID := "fdwecevaqo7wge"
	dashboardName := "cloud-dns-record-info"

	// 构建Grafana URL（使用kiosk模式）
	grafanaURL := c.buildKioskURL(config, dashboardID, dashboardName)

	logger.Info("使用kiosk模式构建Grafana URL", zap.String("url", grafanaURL))

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"url":  grafanaURL,
			"type": "kiosk",
		},
	})
}

// buildKioskURL 构建kiosk模式URL（全屏显示，无需登录）
func (c *MonitoringController) buildKioskURL(config GrafanaConfig, dashboardID, dashboardName string) string {
	baseURL := config.BaseURL
	dashboardPath := "/d/" + dashboardID + "/" + dashboardName

	return baseURL + dashboardPath + "?orgId=" + config.OrgID + "&kiosk"
}

// ProxyGrafana 代理Grafana请求（处理认证和CORS）
func (c *MonitoringController) ProxyGrafana(ctx *gin.Context) {
	logger.Info("代理Grafana请求",
		zap.String("path", ctx.Request.URL.Path),
		zap.String("method", ctx.Request.Method))

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"message": "代理功能开发中",
		"data": gin.H{
			"note": "建议使用kiosk模式或配置Grafana匿名访问",
		},
	})
}

// RefreshMonitoring 刷新监控数据
func (c *MonitoringController) RefreshMonitoring(ctx *gin.Context) {
	logger.Info("刷新监控数据")

	// 这里可以调用实际的数据刷新逻辑
	// 比如重新获取系统指标、检查告警状态等

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "刷新成功",
		"data": gin.H{
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// HandleAlert 处理告警
func (c *MonitoringController) HandleAlert(ctx *gin.Context) {
	var req HandleAlertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "参数错误",
			"data":    nil,
		})
		return
	}

	logger.Info("处理告警",
		zap.Uint("alertId", req.AlertID),
		zap.String("action", req.Action),
		zap.String("reason", req.Reason))

	// 这里可以调用实际的告警处理逻辑
	// 比如更新数据库中的告警状态、发送通知等

	actionText := "处理"
	if req.Action == "ignore" {
		actionText = "忽略"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "告警" + actionText + "成功",
		"data": gin.H{
			"alertId": req.AlertID,
			"action":  req.Action,
		},
	})
}

// GetMonitoringStats 获取监控统计数据
func (c *MonitoringController) GetMonitoringStats(ctx *gin.Context) {
	logger.Info("获取监控统计数据")

	// 模拟监控数据
	data := gin.H{
		"cpu":     "45.2",
		"memory":  "62.8",
		"network": "128",
		"alerts": []gin.H{
			{"id": 1, "time": "2024-04-10 10:30:00", "level": "warning", "source": "Web-Server-01", "message": "CPU 使用率持续超过 80% (当前 85.4%)"},
			{"id": 2, "time": "2024-04-10 10:15:22", "level": "critical", "source": "DB-Master-01", "message": "检测到数据库连接数异常增长，触发限流策略"},
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    data,
	})
}

// ============================================
// Agent 监控增强 API (P0)
// ============================================

// GetOverview 获取监控概览
// GET /api/monitoring/overview
func (c *MonitoringController) GetOverview(ctx *gin.Context) {
	overview, err := c.monitoringService.GetOverview()
	if err != nil {
		logger.Error("获取监控概览失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    overview,
	})
}

// GetServerExtendedMetrics 获取主机扩展指标
// GET /api/servers/:id/extended-metrics
func (c *MonitoringController) GetServerExtendedMetrics(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的主机ID",
		})
		return
	}

	metrics, err := c.monitoringService.GetServerExtendedMetrics(uint(id))
	if err != nil {
		logger.Error("获取主机扩展指标失败", zap.Uint("serverID", uint(id)), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    metrics,
	})
}

// GetServerMetricsHistory 查询主机历史指标
// GET /api/servers/:id/metrics/history
func (c *MonitoringController) GetServerMetricsHistory(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的主机ID",
		})
		return
	}

	metricType := ctx.Query("metricType")
	startTimeStr := ctx.Query("startTime")
	endTimeStr := ctx.Query("endTime")
	interval := ctx.DefaultQuery("interval", "5m")

	if metricType == "" || startTimeStr == "" || endTimeStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "缺少必要参数",
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "开始时间格式错误",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "结束时间格式错误",
		})
		return
	}

	datapoints, err := c.monitoringService.GetMetricsHistory(uint(id), metricType, startTime, endTime, interval)
	if err != nil {
		logger.Error("查询历史指标失败", zap.Uint("serverID", uint(id)), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"metricType": metricType,
			"interval":   interval,
			"datapoints": datapoints,
		},
	})
}

// GetServerHardware 获取主机硬件信息
// GET /api/servers/:id/hardware
func (c *MonitoringController) GetServerHardware(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的主机ID",
		})
		return
	}

	hardware, err := c.monitoringService.GetServerHardware(uint(id))
	if err != nil {
		logger.Error("获取主机硬件信息失败", zap.Uint("serverID", uint(id)), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    hardware,
	})
}

// GetServerProcesses 获取主机进程信息
// GET /api/servers/:id/processes
func (c *MonitoringController) GetServerProcesses(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的主机ID",
		})
		return
	}

	processes, err := c.monitoringService.GetServerProcesses(uint(id))
	if err != nil {
		logger.Error("获取主机进程信息失败", zap.Uint("serverID", uint(id)), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    processes,
	})
}

// GetServerServices 获取主机服务状态
// GET /api/servers/:id/services
func (c *MonitoringController) GetServerServices(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的主机ID",
		})
		return
	}

	services, err := c.monitoringService.GetServerServices(uint(id))
	if err != nil {
		logger.Error("获取主机服务状态失败", zap.Uint("serverID", uint(id)), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    services,
	})
}

// GetServerNetwork 获取主机网络配置
// GET /api/servers/:id/network
func (c *MonitoringController) GetServerNetwork(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的主机ID",
		})
		return
	}

	network, err := c.monitoringService.GetServerNetwork(uint(id))
	if err != nil {
		logger.Error("获取主机网络配置失败", zap.Uint("serverID", uint(id)), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    network,
	})
}

// GetServerSecurity 获取主机安全信息
// GET /api/servers/:id/security
func (c *MonitoringController) GetServerSecurity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "无效的主机ID",
		})
		return
	}

	security, err := c.monitoringService.GetServerSecurity(uint(id))
	if err != nil {
		logger.Error("获取主机安全信息失败", zap.Uint("serverID", uint(id)), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    security,
	})
}

// GetAlerts 获取告警列表
// GET /api/monitoring/alerts
func (c *MonitoringController) GetAlerts(ctx *gin.Context) {
	params := services.AlertQueryParams{
		ServerID:     ctx.Query("serverId"),
		Level:        ctx.Query("level"),
		Acknowledged: ctx.Query("acknowledged"),
	}
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	params.Page = page
	params.PageSize = pageSize

	result, err := c.monitoringService.GetAlerts(params)
	if err != nil {
		logger.Error("获取告警列表失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}

// AcknowledgeAlert 确认告警
// POST /api/monitoring/alerts/:id/acknowledge
func (c *MonitoringController) AcknowledgeAlert(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "无效的告警ID"})
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	_ = ctx.ShouldBindJSON(&req)

	// 从 JWT 中获取当前用户名
	acknowledgedBy := "admin"
	if user, exists := ctx.Get("username"); exists {
		acknowledgedBy = user.(string)
	}

	if err := c.monitoringService.AcknowledgeAlert(uint64(id), acknowledgedBy, req.Comment); err != nil {
		logger.Error("确认告警失败", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

// GetAlertStats 获取告警统计
// GET /api/monitoring/alerts/stats
func (c *MonitoringController) GetAlertStats(ctx *gin.Context) {
	stats, err := c.monitoringService.GetAlertStats()
	if err != nil {
		logger.Error("获取告警统计失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": stats})
}

// ============================================
// 告警规则管理 API (P0 - 动态配置)
// ============================================

// GetAlertRules 获取告警规则列表
// GET /api/monitoring/alerts/rules
func (c *MonitoringController) GetAlertRules(ctx *gin.Context) {
	rules, err := c.alertRuleService.GetAlertRules()
	if err != nil {
		logger.Error("获取告警规则失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    rules,
	})
}

// CreateAlertRule 创建告警规则
// POST /api/monitoring/alerts/rules
func (c *MonitoringController) CreateAlertRule(ctx *gin.Context) {
	var rule services.AlertRule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		logger.Error("参数解析失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := c.alertRuleService.CreateAlertRule(&rule); err != nil {
		logger.Error("创建告警规则失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    rule,
	})
}

// UpdateAlertRule 更新告警规则
// PUT /api/monitoring/alerts/rules/:id
func (c *MonitoringController) UpdateAlertRule(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "规则ID不能为空",
		})
		return
	}

	var rule services.AlertRule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		logger.Error("参数解析失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := c.alertRuleService.UpdateAlertRule(id, &rule); err != nil {
		logger.Error("更新告警规则失败", zap.String("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// DeleteAlertRule 删除告警规则
// DELETE /api/monitoring/alerts/rules/:id
func (c *MonitoringController) DeleteAlertRule(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "规则ID不能为空",
		})
		return
	}

	if err := c.alertRuleService.DeleteAlertRule(id); err != nil {
		logger.Error("删除告警规则失败", zap.String("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// UpdateAlertRuleStatus 更新告警规则状态（启用/禁用）
// PUT /api/monitoring/alerts/rules/:id/status
func (c *MonitoringController) UpdateAlertRuleStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "规则ID不能为空",
		})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("参数解析失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := c.alertRuleService.UpdateAlertRuleStatus(id, req.Enabled); err != nil {
		logger.Error("更新告警规则状态失败", zap.String("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	status := "禁用"
	if req.Enabled {
		status = "启用"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": status + "成功",
	})
}

