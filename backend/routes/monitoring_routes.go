package routes

import (
	"oneops/backend/controller"
	"oneops/backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupMonitoringRoutes 设置监控中心相关路由
func SetupMonitoringRoutes(
	r *gin.Engine,
	monitoringController *controller.MonitoringController,
	wsMonitoringController *controller.MonitoringWebSocketController,
) {
	api := r.Group("/api")

	// WebSocket 实时推送（由 handler 自行验证 token，不经过 Auth 中间件）
	monitoring := api.Group("/monitoring")
	{
		monitoring.GET("/ws", func(ctx *gin.Context) {
			wsMonitoringController.HandleWebSocket(ctx)
		})
	}

	// 监控管理路由（需要认证）
	monitoringAuth := api.Group("/monitoring")
	monitoringAuth.Use(middleware.Auth())
	monitoringAuth.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
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
		monitoring.GET("/overview", monitoringController.GetOverview)
		monitoring.GET("/alerts", monitoringController.GetAlerts)
		monitoring.POST("/alerts/:id/acknowledge", monitoringController.AcknowledgeAlert)
		monitoring.GET("/alerts/stats", monitoringController.GetAlertStats)

		// 告警规则管理 API
		monitoring.GET("/alerts/rules", monitoringController.GetAlertRules)
		monitoring.POST("/alerts/rules", monitoringController.CreateAlertRule)
		monitoring.PUT("/alerts/rules/:id", monitoringController.UpdateAlertRule)
		monitoring.DELETE("/alerts/rules/:id", monitoringController.DeleteAlertRule)
		monitoring.PUT("/alerts/rules/:id/status", monitoringController.UpdateAlertRuleStatus)

		// 通知渠道管理 API
		monitoring.GET("/notifications/channels", monitoringController.GetNotificationChannels)
		monitoring.POST("/notifications/channels", monitoringController.CreateNotificationChannel)
		monitoring.PUT("/notifications/channels/:id", monitoringController.UpdateNotificationChannel)
		monitoring.DELETE("/notifications/channels/:id", monitoringController.DeleteNotificationChannel)
		monitoring.POST("/notifications/channels/:id/test", monitoringController.TestNotificationChannel)

		// 巡检报告 API
		monitoring.GET("/reports", monitoringController.GetReports)
		monitoring.POST("/reports", monitoringController.CreateReport)
		monitoring.GET("/reports/:id", monitoringController.GetReportDetail)
		monitoring.GET("/reports/:id/export", monitoringController.ExportReport)
		monitoring.DELETE("/reports/:id", monitoringController.DeleteReport)
	}
}
