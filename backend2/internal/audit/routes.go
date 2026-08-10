package audit

import (
	"oneops/backend2/internal/system"
	"oneops/backend2/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuditRoutes 设置审计管理相关路由
func SetupAuditRoutes(r *gin.Engine, auditController *AuditController) {
	api := r.Group("/api")
	auditGroup := api.Group("/audit")
	auditGroup.Use(middleware.Auth())
	auditGroup.Use(system.RequirePermissionFromDB()) // 统一权限检查
	{
		// 登录日志
		auditGroup.GET("/login-logs", auditController.GetLoginLogs)
		auditGroup.GET("/login-logs/export", auditController.ExportLoginLogs)

		// 操作日志
		auditGroup.GET("/operation-logs", auditController.GetOperationLogs)
		auditGroup.GET("/operation-logs/export", auditController.ExportOperationLogs)

		// 系统事件日志
		auditGroup.GET("/system-event-logs", auditController.GetSystemEventLogs)

		// 审计统计
		auditGroup.GET("/stats", auditController.GetAuditStats)

		// 可用模块列表
		auditGroup.GET("/modules", auditController.GetModules)
	}
}
