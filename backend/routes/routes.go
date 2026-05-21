package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/handlers"
	"oneops/backend/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 应用中间件
	r.Use(gin.Recovery())
	r.Use(middlewares.Response())
	r.Use(cors.New(middlewares.CORS()))

	// 创建审计中间件
	auditMiddleware := middlewares.NewAuditMiddleware()
	// 应用操作日志审计中间件
	r.Use(auditMiddleware.OperationLog())

	// 创建控制器
	authController := controllers.NewAuthController()
	menuController := controllers.NewMenuController()
	roleController := controllers.NewRoleController()
	userController := controllers.NewUserController()
	auditController := controllers.NewAuditController()
	monitoringController := controllers.NewMonitoringController()
	routeController := controllers.NewRouteController()
	cmdbController := controllers.NewCMDBController()
	bastionController := controllers.NewBastionController()
	sshHandler := handlers.NewSSHWebSocketHandler()

	// API 路由组
	api := r.Group("/api")
	{
		// 认证路由（无需认证）
		api.POST("/login", authController.Login)
		// 常量路由接口（无需认证，因为常量路由本身就是公开的）
		api.GET("/route/getConstantRoutes", routeController.GetConstantRoutes)

		// 用户信息路由（需要认证）
		api.GET("/user/info", middlewares.Auth(), authController.GetUserInfo)
		api.POST("/logout", middlewares.Auth(), authController.Logout)

		// 系统管理路由（需要认证）
		system := api.Group("/system")
		system.Use(middlewares.Auth())
		{
			// 菜单管理
			system.GET("/menus", menuController.GetMenus)
			system.POST("/menus", menuController.CreateMenu)
			system.PUT("/menus/:id", menuController.UpdateMenu)
			system.DELETE("/menus/:id", menuController.DeleteMenu)

			// 角色管理
			system.GET("/roles", roleController.GetRoles)
			system.POST("/roles", roleController.CreateRole)
			system.PUT("/roles/:id", roleController.UpdateRole)
			system.DELETE("/roles/:id", roleController.DeleteRole)

			// 用户管理
			system.GET("/users", userController.GetUsers)
			system.POST("/users", userController.CreateUser)
			system.PUT("/users/:id", userController.UpdateUser)
			system.DELETE("/users/:id", userController.DeleteUser)
		}

		// 审计管理路由（需要认证）
		audit := api.Group("/audit")
		audit.Use(middlewares.Auth())
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

		// 监控管理路由（需要认证）
		monitoring := api.Group("/monitoring")
		monitoring.Use(middlewares.Auth())
		{
			// Grafana面板URL
			monitoring.GET("/grafana/url", monitoringController.GetGrafanaUrl)
			// 监控数据
			monitoring.GET("/stats", monitoringController.GetMonitoringStats)
			// 刷新监控数据
			monitoring.POST("/refresh", monitoringController.RefreshMonitoring)
			// 处理告警
			monitoring.POST("/alert/handle", monitoringController.HandleAlert)
		}

		// CMDB资产管理路由（需要认证）
		cmdb := api.Group("/cmdb")
		cmdb.Use(middlewares.Auth())
		{
			// 服务器管理
			cmdb.GET("/servers", cmdbController.GetServers)
			cmdb.POST("/servers", cmdbController.CreateServer)
			cmdb.PUT("/servers/:id", cmdbController.UpdateServer)
			cmdb.DELETE("/servers/:id", cmdbController.DeleteServer)
			cmdb.GET("/servers/stats", cmdbController.GetServerStats)
			cmdb.POST("/servers/config", cmdbController.GetServerConfig)
			cmdb.GET("/servers/:id", cmdbController.GetServerByID)
			cmdb.POST("/servers/:id/connect", bastionController.ConnectServer)
			cmdb.GET("/servers/:id/permission", bastionController.CheckConnectPermission)

			// 主机分组管理
			cmdb.GET("/groups", cmdbController.GetServerGroups)
			cmdb.GET("/groups/:id", cmdbController.GetServerGroupByID)
			cmdb.POST("/groups", cmdbController.CreateServerGroup)
			cmdb.PUT("/groups/:id", cmdbController.UpdateServerGroup)
			cmdb.DELETE("/groups/:id", cmdbController.DeleteServerGroup)
			cmdb.POST("/groups/assign", cmdbController.AssignServerToGroup)
			cmdb.GET("/group-servers/:groupId", cmdbController.GetServersByGroup)

			// 业务系统管理
			cmdb.GET("/business-units", cmdbController.GetBusinessUnits)
			cmdb.POST("/business-units", cmdbController.CreateBusinessUnit)
			cmdb.PUT("/business-units/:id", cmdbController.UpdateBusinessUnit)
			cmdb.DELETE("/business-units/:id", cmdbController.DeleteBusinessUnit)

			// 机房机柜管理
			cmdb.GET("/rooms", cmdbController.GetServerRooms)
			cmdb.POST("/rooms", cmdbController.CreateServerRoom)
			cmdb.PUT("/rooms/:id", cmdbController.UpdateServerRoom)
			cmdb.DELETE("/rooms/:id", cmdbController.DeleteServerRoom)
			cmdb.GET("/cabinets", cmdbController.GetCabinets)

			// 标签管理
			cmdb.GET("/tags", cmdbController.GetServerTags)
			cmdb.POST("/tags", cmdbController.CreateServerTag)
			cmdb.PUT("/tags/:id", cmdbController.UpdateServerTag)
			cmdb.DELETE("/tags/:id", cmdbController.DeleteServerTag)
			cmdb.POST("/tags/assign", cmdbController.AssignServerTag)
			cmdb.DELETE("/server-tags/:serverId/:tagId", cmdbController.RemoveServerTag)

			// SSH凭证管理
			cmdb.GET("/ssh-credentials", cmdbController.GetSSHCredentials)
			cmdb.GET("/ssh-credentials/:id", cmdbController.GetSSHCredentialByID)
			cmdb.POST("/ssh-credentials", cmdbController.CreateSSHCredential)
			cmdb.PUT("/ssh-credentials/:id", cmdbController.UpdateSSHCredential)
			cmdb.DELETE("/ssh-credentials/:id", cmdbController.DeleteSSHCredential)
			cmdb.POST("/ssh-credentials/:id/test", cmdbController.TestSSHCredential)

			// 资产变更记录
			cmdb.GET("/asset-changes", cmdbController.GetAssetChanges)

			// ========== 堡垒机功能 ==========

			// 会话管理
			cmdb.GET("/sessions", bastionController.GetSessions)
			cmdb.GET("/sessions/active", bastionController.GetActiveSessions)
			cmdb.GET("/sessions/stats", bastionController.GetSessionStats)
			cmdb.GET("/sessions/:id", bastionController.GetSessionByID)
			cmdb.POST("/sessions/:id/terminate", bastionController.TerminateSession)
			cmdb.GET("/sessions/:id/commands", bastionController.GetSessionCommands)
			cmdb.GET("/sessions/:id/file-transfers", bastionController.GetSessionFileTransfers)
			cmdb.POST("/sessions/:id/resize", sshHandler.ResizePTY)

			// WebSocket SSH 连接
			cmdb.GET("/sessions/:id/ws", func(ctx *gin.Context) {
				sshHandler.HandleWebSocket(ctx)
			})

			// 命令审计
			cmdb.GET("/commands", bastionController.GetCommands)

			// 文件传输审计
			cmdb.GET("/file-transfers", bastionController.GetFileTransfers)

			// 访问策略管理
			cmdb.GET("/access-policies", bastionController.GetAccessPolicies)
			cmdb.GET("/access-policies/:id", bastionController.GetAccessPolicyByID)
			cmdb.POST("/access-policies", bastionController.CreateAccessPolicy)
			cmdb.PUT("/access-policies/:id", bastionController.UpdateAccessPolicy)
			cmdb.DELETE("/access-policies/:id", bastionController.DeleteAccessPolicy)
		}

		// 动态路由接口（需要认证）
		routeGroup := api.Group("/route")
		routeGroup.Use(middlewares.Auth())
		{
			routeGroup.GET("/getUserRoutes", routeController.GetUserRoutes)
			routeGroup.GET("/isRouteExist", routeController.IsRouteExist)
		}
	}
}
