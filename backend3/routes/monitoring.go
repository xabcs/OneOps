package routes

import (
	"oneops/backend3/pkg/middleware"

	monsvc "oneops/backend3/service/monitoring"

	monctrl "oneops/backend3/controller/monitoring"

	"github.com/gin-gonic/gin"
)

// SetupMonitoringRoutes 设置监控中心相关路由
func SetupMonitoringRoutes(r *gin.Engine) {
	// 创建 services 和 controller
	monitoringSvc := monsvc.NewMonitoringService()
	monitoringController := monctrl.NewMonitoringController(monitoringSvc)

	api := r.Group("/api")

	// WebSocket 实时推送（由 controller 自行验证 token，不经过 Auth 中间件）
	monitoring := api.Group("/monitoring")
	{
		monitoring.GET("/ws", func(ctx *gin.Context) {
			monitoringController.HandleWebSocket(ctx)
		})
	}

	// 监控管理路由（需要认证）
	monitoringAuth := api.Group("/monitoring")
	monitoringAuth.Use(middleware.Auth())
	monitoringAuth.Use(middleware.RequirePermissionFromDB())
	{
		// 监控概览
		monitoringAuth.GET("/overview", monitoringController.GetOverview)

		// 告警处理
		monitoringAuth.GET("/alerts", monitoringController.GetAlerts)
		monitoringAuth.POST("/alerts/:id/acknowledge", monitoringController.AcknowledgeAlert)
		monitoringAuth.GET("/alerts/stats", monitoringController.GetAlertStats)

		// 告警规则管理 API
		monitoringAuth.GET("/alerts/rules", monitoringController.GetAlertRules)
		monitoringAuth.POST("/alerts/rules", monitoringController.CreateAlertRule)
		monitoringAuth.PUT("/alerts/rules/:id", monitoringController.UpdateAlertRule)
		monitoringAuth.DELETE("/alerts/rules/:id", monitoringController.DeleteAlertRule)
		monitoringAuth.PUT("/alerts/rules/:id/status", monitoringController.UpdateAlertRuleStatus)

		// 通知渠道管理 API
		monitoringAuth.GET("/notifications/channels", monitoringController.GetNotificationChannels)
		monitoringAuth.POST("/notifications/channels", monitoringController.CreateNotificationChannel)
		monitoringAuth.PUT("/notifications/channels/:id", monitoringController.UpdateNotificationChannel)
		monitoringAuth.DELETE("/notifications/channels/:id", monitoringController.DeleteNotificationChannel)
		monitoringAuth.POST("/notifications/channels/:id/test", monitoringController.TestNotificationChannel)

		// 巡检报告 API
		monitoringAuth.GET("/reports", monitoringController.GetReports)
		monitoringAuth.POST("/reports", monitoringController.CreateReport)
		monitoringAuth.GET("/reports/:id", monitoringController.GetReportDetail)
		monitoringAuth.GET("/reports/:id/export", monitoringController.ExportReport)
		monitoringAuth.DELETE("/reports/:id", monitoringController.DeleteReport)
	}
}
