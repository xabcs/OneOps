package routes

import (
	"oneops/backend3/pkg/middleware"

	auditctrl "oneops/backend3/controller/audit"

	"github.com/gin-gonic/gin"
)

// SetupAuditRoutes 设置审计管理相关路由
func SetupAuditRoutes(r *gin.Engine, auditController *auditctrl.AuditController) {
	api := r.Group("/api")
	audit := api.Group("/audit")
	audit.Use(middleware.Auth())
	audit.Use(middleware.RequirePermissionFromDB())
	{
		// 登录日志
		audit.GET("/login-logs", auditController.GetLoginLogs)
		audit.GET("/login-logs/export", auditController.ExportLoginLogs)

		// 操作日志
		audit.GET("/operation-logs", auditController.GetOperationLogs)
		audit.GET("/operation-logs/export", auditController.ExportOperationLogs)

		// 系统事件日志
		audit.GET("/system-event-logs", auditController.GetSystemEventLogs)

		// 审计统计
		audit.GET("/stats", auditController.GetAuditStats)

		// 可用模块列表
		audit.GET("/modules", auditController.GetModules)
	}
}
