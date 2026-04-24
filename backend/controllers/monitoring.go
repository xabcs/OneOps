package controllers

import (
	"net/http"
	"oneops/backend/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MonitoringController 监控控制器
type MonitoringController struct{}

// NewMonitoringController 创建监控控制器
func NewMonitoringController() *MonitoringController {
	return &MonitoringController{}
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

