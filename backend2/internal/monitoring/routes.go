package monitoring

import (
	"oneops/backend2/internal/system"
	"oneops/backend2/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMonitoringRoutes 设置监控中心相关路由
func SetupMonitoringRoutes(
	r *gin.Engine,
	monitoringController *MonitoringController,
	wsMonitoringController *MonitoringWebSocketController,
) {
	api := r.Group("/api")

	// WebSocket 实时推送（由 handler 自行验证 token，不经过 Auth 中间件）
	monitoring := api.Group("/monitoring")
	{
		GET("/ws", func(ctx *gin.Context) {
			wsMonitoringController.HandleWebSocket(ctx)
		})
	}

	// 监控管理路由（需要认证）
	monitoringAuth := api.Group("/monitoring")
	monitoringAuth.Use(middleware.Auth())
	monitoringAuth.Use(system.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		// Grafana面板URL
		monitoringAuth.GET("/grafana/url", monitoringController.GetGrafanaUrl)

		// 监控数据
		monitoringAuth.GET("/stats", monitoringController.GetMonitoringStats)

		// 刷新监控数据
		monitoringAuth.POST("/refresh", monitoringController.RefreshMonitoring)

		// 告警处理
		monitoringAuth.POST("/alert/handle", monitoringController.HandleAlert)

		// Agent 监控增强 API
		GET("/overview", monitoringController.GetOverview)
		GET("/alerts", monitoringController.GetAlerts)
		POST("/alerts/:id/acknowledge", monitoringController.AcknowledgeAlert)
		GET("/alerts/stats", monitoringController.GetAlertStats)

		// 告警规则管理 API
		GET("/alerts/rules", monitoringController.GetAlertRules)
		POST("/alerts/rules", monitoringController.CreateAlertRule)
		PUT("/alerts/rules/:id", monitoringController.UpdateAlertRule)
		DELETE("/alerts/rules/:id", monitoringController.DeleteAlertRule)
		PUT("/alerts/rules/:id/status", monitoringController.UpdateAlertRuleStatus)

		// 通知渠道管理 API
		GET("/notifications/channels", monitoringController.GetNotificationChannels)
		POST("/notifications/channels", monitoringController.CreateNotificationChannel)
		PUT("/notifications/channels/:id", monitoringController.UpdateNotificationChannel)
		DELETE("/notifications/channels/:id", monitoringController.DeleteNotificationChannel)
		POST("/notifications/channels/:id/test", monitoringController.TestNotificationChannel)

		// 巡检报告 API
		GET("/reports", monitoringController.GetReports)
		POST("/reports", monitoringController.CreateReport)
		GET("/reports/:id", monitoringController.GetReportDetail)
		GET("/reports/:id/export", monitoringController.ExportReport)
		DELETE("/reports/:id", monitoringController.DeleteReport)
	}
}
