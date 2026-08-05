package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/container"
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
	r.Use(middlewares.ErrorHandler())  // 统一错误处理
	r.Use(cors.New(middlewares.CORS()))

	// 创建审计中间件
	auditMiddleware := middlewares.NewAuditMiddleware()
	// 应用操作日志审计中间件
	r.Use(auditMiddleware.OperationLog())

	// 获取服务容器
	cnt := container.GetContainer()

	// 创建控制器（使用服务容器）
	authController := controllers.NewAuthController(cnt)
	menuController := controllers.NewMenuController()
	roleController := controllers.NewRoleController()
	userController := controllers.NewUserController()
	auditController := controllers.NewAuditController()
	monitoringController := controllers.NewMonitoringController()
	wsMonitoringController := controllers.NewMonitoringWebSocketController()
	routeController := controllers.NewRouteController()
	cmdbController := controllers.NewCMDBController(cnt)
	bastionController := controllers.NewBastionController()
	attributeController := controllers.NewAttributeController()
	sshHandler := handlers.NewSSHWebSocketHandler()
	k8sClusterController := controllers.NewK8sClusterController(cnt)
	k8sPermissionController := controllers.NewK8sPermissionController(cnt)
	k8sResourceController := controllers.NewK8sResourceController(cnt)
	k8sTerminalHandler := handlers.NewK8sTerminalHandler(cnt)
	diagnosticController := controllers.NewDiagnosticController(cnt)
	applicationPermissionController := controllers.NewApplicationPermissionController()

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
			system.GET("/menus",
				middlewares.RequirePermission("system.menu.list"),
				menuController.GetMenus)
			system.GET("/menus/tree",
				middlewares.RequirePermission("system.menu.list"),
				menuController.GetMenuTree)
			system.POST("/menus",
				middlewares.RequirePermission("system.menu.create"),
				menuController.CreateMenu)
			system.PUT("/menus/:id",
				middlewares.RequirePermission("system.menu.update"),
				menuController.UpdateMenu)
			system.DELETE("/menus/:id",
				middlewares.RequirePermission("system.menu.delete"),
				menuController.DeleteMenu)

			// 角色管理
			system.GET("/roles",
				middlewares.RequirePermission("system.role.list"),
				roleController.GetRoles)
			system.POST("/roles",
				middlewares.RequirePermission("system.role.create"),
				roleController.CreateRole)
			system.PUT("/roles/:id",
				middlewares.RequirePermission("system.role.update"),
				roleController.UpdateRole)
			system.DELETE("/roles/:id",
				middlewares.RequirePermission("system.role.delete"),
				roleController.DeleteRole)

			// 用户管理
			system.GET("/users",
				middlewares.RequirePermission("system.user.list"),
				userController.GetUsers)
			system.POST("/users",
				middlewares.RequirePermission("system.user.create"),
				userController.CreateUser)
			system.PUT("/users/:id",
				middlewares.RequirePermission("system.user.update"),
				userController.UpdateUser)
			system.DELETE("/users/:id",
				middlewares.RequirePermission("system.user.delete"),
				userController.DeleteUser)
			system.PUT("/users/:id/password",
				middlewares.RequirePermission("system.user.reset_password"),
				userController.ResetPassword)


			// 属性管理
			system.GET("/attributes", attributeController.GetAttributeDefinitions)
			system.POST("/attributes", attributeController.CreateAttributeDefinition)
			system.PUT("/attributes/:id", attributeController.UpdateAttributeDefinition)
			system.DELETE("/attributes/:id", attributeController.DeleteAttributeDefinition)
			system.GET("/attributes/:id", attributeController.GetAttributeDefinitionByID)

			// 主机属性管理
			system.GET("/server-attributes/:serverId", attributeController.GetServerAttributes)
			system.POST("/server-attributes/:serverId", attributeController.SaveServerAttributes)

			// 应用权限管理
			system.GET("/applications", applicationPermissionController.GetApplications)
			system.POST("/applications", applicationPermissionController.CreateApplication)
			system.PUT("/applications/:id", applicationPermissionController.UpdateApplication)
			system.DELETE("/applications/:id", applicationPermissionController.DeleteApplication)
			system.POST("/applications/:id/sync-roles", applicationPermissionController.SyncRoles)
			system.GET("/applications/:id/roles", applicationPermissionController.GetApplicationRoles)
			system.POST("/applications/:id/sync-users", applicationPermissionController.SyncUsers)
			system.GET("/applications/:id/users", applicationPermissionController.GetApplicationUsers)
			system.POST("/applications/:id/sync-groups", applicationPermissionController.SyncApplicationGroups)
			system.GET("/applications/:id/groups", applicationPermissionController.GetApplicationGroups)
			system.POST("/applications/:id/sync-rules", applicationPermissionController.SyncAuthorizationRules)
			system.GET("/applications/:id/rules", applicationPermissionController.GetApplicationAuthorizationRules)
			system.GET("/applications/:id/operation-logs", applicationPermissionController.GetOperationLogs)

			// 应用类型配置
			system.GET("/applications/types", applicationPermissionController.GetSupportedAppTypes)
			system.GET("/applications/types/:type/config", applicationPermissionController.GetAppTypeConfigTemplate)

			// 授权中心用户管理（使用 /auth-users 路径避免与系统管理的 /users 冲突）
			system.GET("/auth-users", applicationPermissionController.GetAllAuthUsers)
			system.GET("/auth-users/list", applicationPermissionController.GetAuthUsers)
			system.GET("/auth-users/:id/password", applicationPermissionController.GetAuthUserPassword)
			system.POST("/auth-users", applicationPermissionController.CreateAuthUser)
			system.PUT("/auth-users/:id", applicationPermissionController.UpdateAuthUser)
			system.DELETE("/auth-users/:id", applicationPermissionController.DeleteAuthUser)

			// 授权中心用户组管理（使用 /auth-groups 路径避免与系统管理的 /roles 冲突）
			system.GET("/auth-groups", applicationPermissionController.GetAllAuthGroups)
			system.GET("/auth-groups/list", applicationPermissionController.GetAuthGroups)
			system.POST("/auth-groups", applicationPermissionController.CreateAuthGroup)
			system.PUT("/auth-groups/:id", applicationPermissionController.UpdateAuthGroup)
			system.DELETE("/auth-groups/:id", applicationPermissionController.DeleteAuthGroup)

			// 用户组权限绑定管理
			system.GET("/groups/:id/bindings", applicationPermissionController.GetGroupBindings)
			system.POST("/groups/:id/bindings", applicationPermissionController.CreateGroupBinding)
			system.DELETE("/groups/bindings/:id", applicationPermissionController.DeleteGroupBinding)

			// 用户组成员管理
			system.GET("/users/:id/groups", applicationPermissionController.GetUserGroups)
			system.POST("/users/assign-group", applicationPermissionController.AssignUserToGroup)
			system.DELETE("/users/:id/groups/:groupId", applicationPermissionController.DeleteUserGroup)

			// 用户身份映射管理
			system.GET("/user-identity-mappings", applicationPermissionController.GetUserIdentityMappings)
			system.DELETE("/user-identity-mappings/:id", applicationPermissionController.DeleteUserIdentityMapping)

			// 用户有效权限查询
			system.GET("/user-permissions", applicationPermissionController.GetUserEffectivePermissions)
			system.GET("/user-permissions/matrix", applicationPermissionController.GetUserEffectivePermissionsMatrix)

			// 权限绑定执行记录
			system.GET("/group-bindings/:bindingId/executions", applicationPermissionController.GetGroupBindingExecutions)


			// 注册权限管理路由
			RegisterPermissionRoutes(system)
		}

		// 审计管理路由（需要认证）
		audit := api.Group("/audit")
		audit.Use(middlewares.Auth())
		{
			// 登录日志
			audit.GET("/login-logs",
					middlewares.RequirePermission("audit.login_log.list"),
					auditController.GetLoginLogs)
			audit.GET("/login-logs/export",
					middlewares.RequirePermission("audit.login_log.export"),
					auditController.ExportLoginLogs)

			// 操作日志
			audit.GET("/operation-logs",
					middlewares.RequirePermission("audit.operation_log.list"),
					auditController.GetOperationLogs)
			audit.GET("/operation-logs/export",
					middlewares.RequirePermission("audit.operation_log.export"),
					auditController.ExportOperationLogs)

			// 系统事件日志
			audit.GET("/system-event-logs",
					middlewares.RequirePermission("audit.system_event.list"),
					auditController.GetSystemEventLogs)

			// 审计统计
			audit.GET("/stats",
					middlewares.RequirePermission("audit.stats.view"),
					auditController.GetAuditStats)

			// 可用模块列表
			audit.GET("/modules",
					middlewares.RequirePermission("audit.system_event.list"),
					auditController.GetModules)
		}

		// 监控管理路由（需要认证）
		monitoring := api.Group("/monitoring")
		{
			// WebSocket 实时推送（由 handler 自行验证 token，不经过 Auth 中间件）
			monitoring.GET("/ws", func(ctx *gin.Context) {
				wsMonitoringController.HandleWebSocket(ctx)
			})
		}

		// 监控管理路由（需要认证）
		monitoringAuth := api.Group("/monitoring")
		monitoringAuth.Use(middlewares.Auth())
		{
			// Grafana面板URL
			monitoringAuth.GET("/grafana/url",
					middlewares.RequirePermission("monitor.data.view"),
					monitoringController.GetGrafanaUrl)
			// 监控数据
			monitoringAuth.GET("/stats",
					middlewares.RequirePermission("monitor.data.view"),
					monitoringController.GetMonitoringStats)
			// 刷新监控数据
			monitoringAuth.POST("/refresh",
					middlewares.RequirePermission("monitor.data.view"),
					monitoringController.RefreshMonitoring)
			monitoringAuth.POST("/alert/handle",
					middlewares.RequirePermission("monitor.alert.handle"),
					monitoringController.HandleAlert)

			// Agent 监控增强 API (P0)
			monitoring.GET("/overview",
					middlewares.RequirePermission("monitor.data.view"),
					monitoringController.GetOverview)
			monitoring.GET("/alerts",
					middlewares.RequirePermission("monitor.alert.list"),
					monitoringController.GetAlerts)
			monitoring.POST("/alerts/:id/acknowledge",
					middlewares.RequirePermission("monitor.alert.ack"),
					monitoringController.AcknowledgeAlert)
			monitoring.GET("/alerts/stats",
					middlewares.RequirePermission("monitor.alert.list"),
					monitoringController.GetAlertStats)

				// 告警规则管理 API (P0 - 动态配置)
			monitoring.GET("/alerts/rules",
					middlewares.RequirePermission("monitor.task.list"),
					monitoringController.GetAlertRules)
			monitoring.POST("/alerts/rules",
					middlewares.RequirePermission("monitor.task.create"),
					monitoringController.CreateAlertRule)
			monitoring.PUT("/alerts/rules/:id",
					middlewares.RequirePermission("monitor.task.update"),
					monitoringController.UpdateAlertRule)
			monitoring.DELETE("/alerts/rules/:id",
					middlewares.RequirePermission("monitor.task.delete"),
					monitoringController.DeleteAlertRule)
			monitoring.PUT("/alerts/rules/:id/status",
					middlewares.RequirePermission("monitor.task.update"),
					monitoringController.UpdateAlertRuleStatus)

			// 通知渠道管理 API (P2)
			monitoring.GET("/notifications/channels",
					middlewares.RequirePermission("monitor.task.list"),
					monitoringController.GetNotificationChannels)
			monitoring.POST("/notifications/channels",
					middlewares.RequirePermission("monitor.task.create"),
					monitoringController.CreateNotificationChannel)
			monitoring.PUT("/notifications/channels/:id",
					middlewares.RequirePermission("monitor.task.update"),
					monitoringController.UpdateNotificationChannel)
			monitoring.DELETE("/notifications/channels/:id",
					middlewares.RequirePermission("monitor.task.delete"),
					monitoringController.DeleteNotificationChannel)
			monitoring.POST("/notifications/channels/:id/test",
					middlewares.RequirePermission("monitor.task.execute"),
					monitoringController.TestNotificationChannel)

				// 巡检报告 API (P3)
				monitoring.GET("/reports",
					middlewares.RequirePermission("monitor.data.view"),
					monitoringController.GetReports)
				monitoring.POST("/reports",
					middlewares.RequirePermission("monitor.task.create"),
					monitoringController.CreateReport)
				monitoring.GET("/reports/:id",
					middlewares.RequirePermission("monitor.data.view"),
					monitoringController.GetReportDetail)
				monitoring.GET("/reports/:id/export",
					middlewares.RequirePermission("monitor.data.export"),
					monitoringController.ExportReport)
				monitoring.DELETE("/reports/:id",
					middlewares.RequirePermission("monitor.task.delete"),
					monitoringController.DeleteReport)
		}

		// CMDB资产管理路由（需要认证）
		cmdb := api.Group("/cmdb")
		cmdb.Use(middlewares.Auth())
		{
			// 服务器管理
			cmdb.GET("/servers",
					middlewares.RequirePermission("cmdb.server.list"),
					cmdbController.GetServers)
			cmdb.POST("/servers",
					middlewares.RequirePermission("cmdb.server.create"),
					cmdbController.CreateServer)
			cmdb.PUT("/servers/:id",
					middlewares.RequirePermission("cmdb.server.update"),
					cmdbController.UpdateServer)
			cmdb.DELETE("/servers/:id",
					middlewares.RequirePermission("cmdb.server.delete"),
					cmdbController.DeleteServer)
			cmdb.GET("/servers/stats",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerStats)
			cmdb.POST("/servers/config",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerConfig)
			cmdb.GET("/servers/:id",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerByID)
			cmdb.GET("/servers/:id/connect",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerForConnect) // 新增：轻量级连接接口
			cmdb.POST("/servers/:id/connect",
					middlewares.RequirePermission("cmdb.server.connect"),
					bastionController.ConnectServer)
			cmdb.GET("/servers/:id/permission",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.CheckConnectPermission)
			cmdb.POST("/servers/:id/sync-metrics",
					middlewares.RequirePermission("cmdb.server.update"),
					cmdbController.SyncServerMetrics)

			// Agent 管理
			cmdb.POST("/servers/:id/agent/deploy",
					middlewares.RequirePermission("cmdb.agents.deploy"),
					cmdbController.DeployAgent)
			cmdb.POST("/servers/:id/agent/restart",
					middlewares.RequirePermission("cmdb.agents.restart"),
					cmdbController.RestartAgent)
			cmdb.POST("/servers/:id/agent/uninstall",
					middlewares.RequirePermission("cmdb.agents.uninstall"),
					cmdbController.UninstallAgent)
			cmdb.GET("/servers/:id/agent/status",
					middlewares.RequirePermission("cmdb.agents.list"),
					cmdbController.GetAgentStatus)
			cmdb.POST("/servers/:id/test-connection",
					middlewares.RequirePermission("cmdb.server.update"),
					cmdbController.TestSSHConnection)

			// Agent 管理页面专用接口
			cmdb.GET("/agents",
					middlewares.RequirePermission("cmdb.agents.list"),
					cmdbController.GetAgentList)
			cmdb.POST("/agents/batch-deploy",
					middlewares.RequirePermission("cmdb.agents.deploy"),
					cmdbController.BatchDeployAgent)
			cmdb.POST("/agents/batch-uninstall",
					middlewares.RequirePermission("cmdb.agents.uninstall"),
					cmdbController.BatchUninstallAgent)
			cmdb.DELETE("/agents/:id",
					middlewares.RequirePermission("cmdb.agents.uninstall"),
					cmdbController.DeleteAgentRecord)

			// Agent 监控增强 API (P0) - 主机相关监控接口
			cmdb.GET("/servers/:id/extended-metrics",
					middlewares.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerExtendedMetrics)
			cmdb.GET("/servers/:id/metrics/history",
					middlewares.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerMetricsHistory)
			cmdb.GET("/servers/:id/hardware",
					middlewares.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerHardware)
			cmdb.GET("/servers/:id/processes",
					middlewares.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerProcesses)
			cmdb.GET("/servers/:id/services",
					middlewares.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerServices)
			cmdb.GET("/servers/:id/network",
					middlewares.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerNetwork)
			cmdb.GET("/servers/:id/security",
					middlewares.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerSecurity)

			// 主机分组管理
			cmdb.GET("/groups",
					middlewares.RequirePermission("cmdb.group.list"),
					cmdbController.GetServerGroups)
			cmdb.GET("/groups/:id",
					middlewares.RequirePermission("cmdb.group.view"),
					cmdbController.GetServerGroupByID)
			cmdb.GET("/asset-tree",
					middlewares.RequirePermission("cmdb.group.list"),
					cmdbController.GetAssetTree)
			cmdb.POST("/groups",
					middlewares.RequirePermission("cmdb.group.create"),
					cmdbController.CreateServerGroup)
			cmdb.PUT("/groups/:id",
					middlewares.RequirePermission("cmdb.group.update"),
					cmdbController.UpdateServerGroup)
			cmdb.DELETE("/groups/:id",
					middlewares.RequirePermission("cmdb.group.delete"),
					cmdbController.DeleteServerGroup)
			cmdb.POST("/groups/assign",
					middlewares.RequirePermission("cmdb.group.assign"),
					cmdbController.AssignServerToGroup)
			cmdb.POST("/groups/assign-multi",
					middlewares.RequirePermission("cmdb.group.assign"),
					cmdbController.AssignServerToGroups)
			cmdb.GET("/group-servers/:groupId",
					middlewares.RequirePermission("cmdb.group.view"),
					cmdbController.GetServersByGroup)

			// 业务系统管理
			cmdb.GET("/business-units",
					middlewares.RequirePermission("cmdb.business.list"),
					cmdbController.GetBusinessUnits)
			cmdb.POST("/business-units",
					middlewares.RequirePermission("cmdb.business.create"),
					cmdbController.CreateBusinessUnit)
			cmdb.PUT("/business-units/:id",
					middlewares.RequirePermission("cmdb.business.update"),
					cmdbController.UpdateBusinessUnit)
			cmdb.DELETE("/business-units/:id",
					middlewares.RequirePermission("cmdb.business.delete"),
					cmdbController.DeleteBusinessUnit)

			// 机房机柜管理
			cmdb.GET("/rooms",
					middlewares.RequirePermission("cmdb.rooms.list"),
					cmdbController.GetServerRooms)
			cmdb.POST("/rooms",
					middlewares.RequirePermission("cmdb.rooms.create"),
					cmdbController.CreateServerRoom)
			cmdb.PUT("/rooms/:id",
					middlewares.RequirePermission("cmdb.rooms.update"),
					cmdbController.UpdateServerRoom)
			cmdb.DELETE("/rooms/:id",
					middlewares.RequirePermission("cmdb.rooms.delete"),
					cmdbController.DeleteServerRoom)
			cmdb.GET("/cabinets",
					middlewares.RequirePermission("cmdb.rooms.list"),
					cmdbController.GetCabinets)

			// 标签管理
			cmdb.GET("/tags",
					middlewares.RequirePermission("cmdb.tags.list"),
					cmdbController.GetServerTags)
			cmdb.POST("/tags",
					middlewares.RequirePermission("cmdb.tags.create"),
					cmdbController.CreateServerTag)
			cmdb.PUT("/tags/:id",
					middlewares.RequirePermission("cmdb.tags.update"),
					cmdbController.UpdateServerTag)
			cmdb.DELETE("/tags/:id",
					middlewares.RequirePermission("cmdb.tags.delete"),
					cmdbController.DeleteServerTag)
			cmdb.POST("/tags/assign",
					middlewares.RequirePermission("cmdb.tags.update"),
					cmdbController.AssignServerTag)
			cmdb.DELETE("/server-tags/:serverId/:tagId",
					middlewares.RequirePermission("cmdb.tags.update"),
					cmdbController.RemoveServerTag)

			// SSH凭证管理
			cmdb.GET("/ssh-credentials",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.GetSSHCredentials)
			cmdb.GET("/ssh-credentials/:id",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.GetSSHCredentialByID)
			cmdb.POST("/ssh-credentials",
					middlewares.RequirePermission("cmdb.server.update"),
					cmdbController.CreateSSHCredential)
			cmdb.PUT("/ssh-credentials/:id",
					middlewares.RequirePermission("cmdb.server.update"),
					cmdbController.UpdateSSHCredential)
			cmdb.DELETE("/ssh-credentials/:id",
					middlewares.RequirePermission("cmdb.server.delete"),
					cmdbController.DeleteSSHCredential)
			cmdb.POST("/ssh-credentials/:id/test",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.TestSSHCredential)

			// 资产变更记录
			cmdb.GET("/asset-changes",
					middlewares.RequirePermission("cmdb.server.view"),
					cmdbController.GetAssetChanges)

			// ========== 堡垒机功能 ==========

			// 会话管理
			cmdb.GET("/sessions",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetSessions)
			cmdb.GET("/sessions/list",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionsList) // 轻量级列表接口，只返回展示字段
			cmdb.GET("/sessions/active",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetActiveSessions)
			cmdb.GET("/sessions/active-memory",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetActiveSessionsFromMemory) // 从内存获取真正活跃的会话
			cmdb.GET("/sessions/stats",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionStats)
			cmdb.GET("/sessions/:id",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionByID)
			cmdb.POST("/sessions/:id/terminate",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.TerminateSession)
			cmdb.GET("/sessions/:id/commands",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionCommands)
			cmdb.GET("/sessions/:id/file-transfers",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionFileTransfers)
			cmdb.POST("/sessions/:id/resize",
					middlewares.RequirePermission("cmdb.server.view"),
					sshHandler.ResizePTY)

			// WebSocket SSH 连接（不经过 Auth 中间件，由 handler 自行从 query param 验证 token）
			// 注意：此路由注册在 cmdb 组外，见下方

			// 命令审计
			cmdb.GET("/commands",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetCommands)

			// 文件传输审计
			cmdb.GET("/file-transfers",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetFileTransfers)

			// 访问策略管理
			cmdb.GET("/access-policies",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetAccessPolicies)
			cmdb.GET("/access-policies/:id",
					middlewares.RequirePermission("cmdb.server.view"),
					bastionController.GetAccessPolicyByID)
			cmdb.POST("/access-policies",
					middlewares.RequirePermission("cmdb.server.update"),
					bastionController.CreateAccessPolicy)
			cmdb.PUT("/access-policies/:id",
					middlewares.RequirePermission("cmdb.server.update"),
					bastionController.UpdateAccessPolicy)
			cmdb.DELETE("/access-policies/:id",
					middlewares.RequirePermission("cmdb.server.delete"),
					bastionController.DeleteAccessPolicy)
		}

		// WebSocket SSH 连接（不经过 Auth 中间件，handler 自行从 query param 验证 token）
		api.GET("/cmdb/sessions/:id/ws", func(ctx *gin.Context) {
			sshHandler.HandleWebSocket(ctx)
		})

		// K8s Pod 终端 WebSocket（不经过 Auth 中间件，由 handler 自行验证）
		api.GET("/k8s/terminal/ws", func(ctx *gin.Context) {
			k8sTerminalHandler.HandleWebSocket(ctx)
		})

		// Agent 心跳（不经过 Auth 中间件，由 Agent 直接上报）
		api.POST("/cmdb/agent/heartbeat", cmdbController.ReceiveAgentHeartbeat)

		// Agent 版本管理接口（需要认证）
		api.GET("/cmdb/agent-versions", cmdbController.GetAgentVersions)
		api.GET("/cmdb/agent-versions/latest", cmdbController.GetLatestAgentVersion)
		api.GET("/cmdb/agent-versions/:id", cmdbController.GetAgentVersionByID)
		api.POST("/cmdb/agent-versions", cmdbController.CreateAgentVersion)
		api.PUT("/cmdb/agent-versions/:id", cmdbController.UpdateAgentVersion)
		api.DELETE("/cmdb/agent-versions/:id", cmdbController.DeleteAgentVersion)

		// Agent 升级管理接口（需要认证）
		api.POST("/cmdb/servers/:id/agent/upgrade", cmdbController.UpgradeAgent)
		api.GET("/cmdb/agent-upgrade-tasks", cmdbController.GetUpgradeTasks)
		api.GET("/cmdb/agent-upgrade-tasks/:id", cmdbController.GetUpgradeTaskByID)

		// 动态路由接口（需要认证）
		routeGroup := api.Group("/route")
		routeGroup.Use(middlewares.Auth())
		{
			routeGroup.GET("/getUserRoutes", routeController.GetUserRoutes)
			routeGroup.GET("/isRouteExist", routeController.IsRouteExist)
			routeGroup.POST("/invalidateCache", routeController.InvalidateCache)
			routeGroup.GET("/debugCache", middlewares.Auth(), routeController.DebugCache)
		}

			// K8s 集群管理路由（需要认证）
			k8s := api.Group("/k8s")
			k8s.Use(middlewares.Auth())
			{
				// 集群管理
				k8s.GET("/clusters",
					middlewares.RequirePermission("k8s.cluster.list"),
					k8sClusterController.GetClusters)
				k8s.POST("/clusters",
					middlewares.RequirePermission("k8s.cluster.create"),
					k8sClusterController.CreateCluster)
				k8s.GET("/clusters/:id",
					middlewares.RequirePermission("k8s.cluster.view"),
					k8sClusterController.GetClusterByID)
				k8s.PUT("/clusters/:id",
					middlewares.RequirePermission("k8s.cluster.update"),
					k8sClusterController.UpdateCluster)
				k8s.DELETE("/clusters/:id",
					middlewares.RequirePermission("k8s.cluster.delete"),
					k8sClusterController.DeleteCluster)

				// 连接测试
				k8s.POST("/clusters/:id/test",
					middlewares.RequirePermission("k8s.cluster.connect"),
					k8sClusterController.TestConnection)

				// 集群节点和命名空间
				k8s.GET("/clusters/:id/nodes",
					middlewares.RequirePermission("k8s.cluster.view"),
					k8sClusterController.GetClusterNodes)
				k8s.GET("/clusters/:id/namespaces",
					middlewares.RequirePermission("k8s.cluster.view"),
					k8sClusterController.GetClusterNamespaces)

				// 集群用户管理
				k8s.GET("/clusters/:id/users",
					middlewares.RequirePermission("k8s.permission.list"),
					k8sClusterController.GetClusterUsers)

				// 权限管理
				k8s.POST("/clusters/:id/permissions",
					middlewares.RequirePermission("k8s.permission.assign"),
					k8sPermissionController.AssignClusterRole)
				k8s.DELETE("/clusters/:id/permissions/:userId",
					middlewares.RequirePermission("k8s.permission.revoke"),
					k8sPermissionController.RevokeClusterRole)
				k8s.GET("/users/clusters",
					middlewares.RequirePermission("k8s.permission.list"),
					k8sPermissionController.GetUserClusters)
				k8s.GET("/clusters/:id/users/:userId/role",
					middlewares.RequirePermission("k8s.permission.list"),
					k8sPermissionController.GetUserRoleInCluster)
				k8s.POST("/permissions/batch-assign",
					middlewares.RequirePermission("k8s.permission.assign"),
					k8sPermissionController.BatchAssignClusterRoles)
					// Workloads - Deployments
					k8s.GET("/clusters/:id/deployments",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListDeployments)
					k8s.GET("/clusters/:id/deployments/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDeployment)
					k8s.GET("/clusters/:id/deployments/:namespace/:name/pods",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDeploymentPods)
					k8s.POST("/clusters/:id/deployments",
					middlewares.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateDeployment)
					k8s.PUT("/clusters/:id/deployments",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateDeployment)
					k8s.DELETE("/clusters/:id/deployments",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteDeployment)
					k8s.POST("/clusters/:id/deployments/scale",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.ScaleDeployment)
					k8s.POST("/clusters/:id/deployments/restart",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.RestartDeployment)

					// Workloads - StatefulSets
					k8s.GET("/clusters/:id/statefulsets",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListStatefulSets)
					k8s.GET("/clusters/:id/statefulsets/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetStatefulSet)
					k8s.GET("/clusters/:id/statefulsets/:namespace/:name/pods",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetStatefulSetPods)
					// k8s.POST("/clusters/:id/statefulsets", k8sResourceController.CreateStatefulSet)
					// k8s.PUT("/clusters/:id/statefulsets", k8sResourceController.UpdateStatefulSet)
					// k8s.DELETE("/clusters/:id/statefulsets", k8sResourceController.DeleteStatefulSet)
					// k8s.POST("/clusters/:id/statefulsets/restart", k8sResourceController.RestartStatefulSet)

					// Workloads - DaemonSets
					k8s.GET("/clusters/:id/daemonsets",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListDaemonSets)
					k8s.GET("/clusters/:id/daemonsets/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDaemonSet)
					k8s.GET("/clusters/:id/daemonsets/:namespace/:name/pods",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDaemonSetPods)
					// k8s.POST("/clusters/:id/daemonsets", k8sResourceController.CreateDaemonSet)
					// k8s.PUT("/clusters/:id/daemonsets", k8sResourceController.UpdateDaemonSet)
					// k8s.DELETE("/clusters/:id/daemonsets", k8sResourceController.DeleteDaemonSet)
					// k8s.POST("/clusters/:id/daemonsets/restart", k8sResourceController.RestartDaemonSet)

					// Workloads - Jobs
					k8s.GET("/clusters/:id/jobs",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListJobs)
					k8s.GET("/clusters/:id/jobs/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetJob)
					k8s.GET("/clusters/:id/jobs/:namespace/:name/pods",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetJobPods)
					k8s.DELETE("/clusters/:id/jobs",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteJob)

					// Workloads - CronJobs
					k8s.GET("/clusters/:id/cronjobs",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListCronJobs)
					k8s.GET("/clusters/:id/cronjobs/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetCronJob)
					k8s.GET("/clusters/:id/cronjobs/:namespace/:name/pods",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetCronJobPods)
					k8s.DELETE("/clusters/:id/cronjobs",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteCronJob)
					k8s.PUT("/clusters/:id/cronjobs/suspend",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.SuspendCronJob)

					// Services
					k8s.GET("/clusters/:id/services",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListServices)
					k8s.GET("/clusters/:id/services/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetService)
					k8s.POST("/clusters/:id/services",
					middlewares.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateService)
					k8s.PUT("/clusters/:id/services",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateService)
					k8s.DELETE("/clusters/:id/services",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteService)

					// Ingresses
					k8s.GET("/clusters/:id/ingresses",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListIngress)
					k8s.GET("/clusters/:id/ingresses/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetIngress)
					k8s.POST("/clusters/:id/ingresses",
					middlewares.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateIngress)
					k8s.PUT("/clusters/:id/ingresses",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateIngress)
					k8s.DELETE("/clusters/:id/ingresses",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteIngress)

					// Pods
					k8s.GET("/clusters/:id/pods",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListPods)
					k8s.GET("/clusters/:id/pods/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetPod)
					k8s.PUT("/clusters/:id/pods",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdatePod)
					k8s.GET("/clusters/:id/pods/:namespace/:name/logs",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetPodLogs)
					k8s.DELETE("/clusters/:id/pods",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeletePod)

					// ConfigMaps
					k8s.GET("/clusters/:id/configmaps",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListConfigMaps)
					k8s.GET("/clusters/:id/configmaps/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetConfigMap)
					k8s.POST("/clusters/:id/configmaps",
					middlewares.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateConfigMap)
					k8s.PUT("/clusters/:id/configmaps",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateConfigMap)
					k8s.DELETE("/clusters/:id/configmaps",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteConfigMap)

					// Secrets
					k8s.GET("/clusters/:id/secrets",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListSecrets)
					k8s.GET("/clusters/:id/secrets/:namespace/:name",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetSecret)
					k8s.POST("/clusters/:id/secrets",
					middlewares.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateSecret)
					k8s.PUT("/clusters/:id/secrets",
					middlewares.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateSecret)
					k8s.DELETE("/clusters/:id/secrets",
					middlewares.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteSecret)

					// Events
					k8s.GET("/clusters/:id/events",
					middlewares.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListEvents)

					// K8s 终端管理（需要认证）
					k8s.GET("/terminal/active",
					middlewares.RequirePermission("k8s.cluster.view"),
					k8sTerminalHandler.GetActiveSessions)
					k8s.POST("/terminal/sessions/:sessionId/terminate",
					middlewares.RequirePermission("k8s.cluster.view"),
					k8sTerminalHandler.TerminateSession)

					// K8s 诊断功能（需要认证）
					k8s.GET("/diagnostic/commands",
					middlewares.RequirePermission("k8s.cluster.view"),
					diagnosticController.GetDiagnosticCommands)
					k8s.GET("/diagnostic/pods/:clusterId/:namespace",
					middlewares.RequirePermission("k8s.resource.view"),
					diagnosticController.GetJavaPods)
					k8s.GET("/diagnostic/namespaces/:clusterId",
					middlewares.RequirePermission("k8s.cluster.view"),
					diagnosticController.GetNamespaces)
					k8s.POST("/diagnostic/execute",
					middlewares.RequirePermission("k8s.resource.view"),
					diagnosticController.ExecuteDiagnostic)
					k8s.GET("/diagnostic/history",
					middlewares.RequirePermission("k8s.resource.view"),
					diagnosticController.GetDiagnosticHistory)

			}

	}
}
