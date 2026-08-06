package routes

import (
	"oneops/backend/controller"
	"oneops/backend/container"
	"oneops/backend/handler"
	"oneops/backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 应用中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Response())
	r.Use(middleware.ErrorHandler())  // 统一错误处理
	r.Use(cors.New(middleware.CORS()))

	// 创建审计中间件
	auditMiddleware := middleware.NewAuditMiddleware()
	// 应用操作日志审计中间件
	r.Use(auditMiddleware.OperationLog())

	// 获取服务容器
	cnt := container.GetContainer()

	// 创建控制器（使用服务容器）
	authController := controller.NewAuthController(cnt)
	menuController := controller.NewMenuController()
	roleController := controller.NewRoleController()
	userController := controller.NewUserController()
	auditController := controller.NewAuditController()
	monitoringController := controller.NewMonitoringController()
	wsMonitoringController := controller.NewMonitoringWebSocketController()
	routeController := controller.NewRouteController()
	cmdbController := controller.NewCMDBController(cnt)
	bastionController := controller.NewBastionController()
	attributeController := controller.NewAttributeController()
	sshHandler := handler.NewSSHWebSocketHandler()
	k8sClusterController := controller.NewK8sClusterController(cnt)
	k8sPermissionController := controller.NewK8sPermissionController(cnt)
	k8sResourceController := controller.NewK8sResourceController(cnt)
	k8sTerminalHandler := handler.NewK8sTerminalHandler(cnt)
	diagnosticController := controller.NewDiagnosticController(cnt)
	applicationPermissionController := controller.NewApplicationPermissionController()

	// API 路由组
	api := r.Group("/api")
	{
		// 认证路由（无需认证）
		api.POST("/login", authController.Login)
		// 常量路由接口（无需认证，因为常量路由本身就是公开的）
		api.GET("/route/getConstantRoutes", routeController.GetConstantRoutes)

		// 用户信息路由（需要认证）
		api.GET("/user/info", middleware.Auth(), authController.GetUserInfo)
		api.POST("/logout", middleware.Auth(), authController.Logout)

		// 系统管理路由（需要认证）
		system := api.Group("/system")
		system.Use(middleware.Auth())
		{
			// 菜单管理
			system.GET("/menus",
				middleware.RequirePermission("system.menu.list"),
				menuController.GetMenus)
			system.GET("/menus/tree",
				middleware.RequirePermission("system.menu.list"),
				menuController.GetMenuTree)
			system.POST("/menus",
				middleware.RequirePermission("system.menu.create"),
				menuController.CreateMenu)
			system.PUT("/menus/:id",
				middleware.RequirePermission("system.menu.update"),
				menuController.UpdateMenu)
			system.DELETE("/menus/:id",
				middleware.RequirePermission("system.menu.delete"),
				menuController.DeleteMenu)

			// 角色管理
			system.GET("/roles",
				middleware.RequirePermission("system.role.list"),
				roleController.GetRoles)
			system.POST("/roles",
				middleware.RequirePermission("system.role.create"),
				roleController.CreateRole)
			system.PUT("/roles/:id",
				middleware.RequirePermission("system.role.update"),
				roleController.UpdateRole)
			system.DELETE("/roles/:id",
				middleware.RequirePermission("system.role.delete"),
				roleController.DeleteRole)

			// 用户管理
			system.GET("/users",
				middleware.RequirePermission("system.user.list"),
				userController.GetUsers)
			system.POST("/users",
				middleware.RequirePermission("system.user.create"),
				userController.CreateUser)
			system.PUT("/users/:id",
				middleware.RequirePermission("system.user.update"),
				userController.UpdateUser)
			system.DELETE("/users/:id",
				middleware.RequirePermission("system.user.delete"),
				userController.DeleteUser)
			system.PUT("/users/:id/password",
				middleware.RequirePermission("system.user.reset_password"),
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
		audit.Use(middleware.Auth())
		{
			// 登录日志
			audit.GET("/login-logs",
					middleware.RequirePermission("audit.login_log.list"),
					auditController.GetLoginLogs)
			audit.GET("/login-logs/export",
					middleware.RequirePermission("audit.login_log.export"),
					auditController.ExportLoginLogs)

			// 操作日志
			audit.GET("/operation-logs",
					middleware.RequirePermission("audit.operation_log.list"),
					auditController.GetOperationLogs)
			audit.GET("/operation-logs/export",
					middleware.RequirePermission("audit.operation_log.export"),
					auditController.ExportOperationLogs)

			// 系统事件日志
			audit.GET("/system-event-logs",
					middleware.RequirePermission("audit.system_event.list"),
					auditController.GetSystemEventLogs)

			// 审计统计
			audit.GET("/stats",
					middleware.RequirePermission("audit.stats.view"),
					auditController.GetAuditStats)

			// 可用模块列表
			audit.GET("/modules",
					middleware.RequirePermission("audit.system_event.list"),
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
		monitoringAuth.Use(middleware.Auth())
		{
			// Grafana面板URL
			monitoringAuth.GET("/grafana/url",
					middleware.RequirePermission("monitor.data.view"),
					monitoringController.GetGrafanaUrl)
			// 监控数据
			monitoringAuth.GET("/stats",
					middleware.RequirePermission("monitor.data.view"),
					monitoringController.GetMonitoringStats)
			// 刷新监控数据
			monitoringAuth.POST("/refresh",
					middleware.RequirePermission("monitor.data.view"),
					monitoringController.RefreshMonitoring)
			monitoringAuth.POST("/alert/handle",
					middleware.RequirePermission("monitor.alert.handle"),
					monitoringController.HandleAlert)

			// Agent 监控增强 API (P0)
			monitoring.GET("/overview",
					middleware.RequirePermission("monitor.data.view"),
					monitoringController.GetOverview)
			monitoring.GET("/alerts",
					middleware.RequirePermission("monitor.alert.list"),
					monitoringController.GetAlerts)
			monitoring.POST("/alerts/:id/acknowledge",
					middleware.RequirePermission("monitor.alert.ack"),
					monitoringController.AcknowledgeAlert)
			monitoring.GET("/alerts/stats",
					middleware.RequirePermission("monitor.alert.list"),
					monitoringController.GetAlertStats)

				// 告警规则管理 API (P0 - 动态配置)
			monitoring.GET("/alerts/rules",
					middleware.RequirePermission("monitor.task.list"),
					monitoringController.GetAlertRules)
			monitoring.POST("/alerts/rules",
					middleware.RequirePermission("monitor.task.create"),
					monitoringController.CreateAlertRule)
			monitoring.PUT("/alerts/rules/:id",
					middleware.RequirePermission("monitor.task.update"),
					monitoringController.UpdateAlertRule)
			monitoring.DELETE("/alerts/rules/:id",
					middleware.RequirePermission("monitor.task.delete"),
					monitoringController.DeleteAlertRule)
			monitoring.PUT("/alerts/rules/:id/status",
					middleware.RequirePermission("monitor.task.update"),
					monitoringController.UpdateAlertRuleStatus)

			// 通知渠道管理 API (P2)
			monitoring.GET("/notifications/channels",
					middleware.RequirePermission("monitor.task.list"),
					monitoringController.GetNotificationChannels)
			monitoring.POST("/notifications/channels",
					middleware.RequirePermission("monitor.task.create"),
					monitoringController.CreateNotificationChannel)
			monitoring.PUT("/notifications/channels/:id",
					middleware.RequirePermission("monitor.task.update"),
					monitoringController.UpdateNotificationChannel)
			monitoring.DELETE("/notifications/channels/:id",
					middleware.RequirePermission("monitor.task.delete"),
					monitoringController.DeleteNotificationChannel)
			monitoring.POST("/notifications/channels/:id/test",
					middleware.RequirePermission("monitor.task.execute"),
					monitoringController.TestNotificationChannel)

				// 巡检报告 API (P3)
				monitoring.GET("/reports",
					middleware.RequirePermission("monitor.data.view"),
					monitoringController.GetReports)
				monitoring.POST("/reports",
					middleware.RequirePermission("monitor.task.create"),
					monitoringController.CreateReport)
				monitoring.GET("/reports/:id",
					middleware.RequirePermission("monitor.data.view"),
					monitoringController.GetReportDetail)
				monitoring.GET("/reports/:id/export",
					middleware.RequirePermission("monitor.data.export"),
					monitoringController.ExportReport)
				monitoring.DELETE("/reports/:id",
					middleware.RequirePermission("monitor.task.delete"),
					monitoringController.DeleteReport)
		}

		// CMDB资产管理路由（需要认证）
		cmdb := api.Group("/cmdb")
		cmdb.Use(middleware.Auth())
		{
			// 服务器管理
			cmdb.GET("/servers",
					middleware.RequirePermission("cmdb.server.list"),
					cmdbController.GetServers)
			cmdb.POST("/servers",
					middleware.RequirePermission("cmdb.server.create"),
					cmdbController.CreateServer)
			cmdb.PUT("/servers/:id",
					middleware.RequirePermission("cmdb.server.update"),
					cmdbController.UpdateServer)
			cmdb.DELETE("/servers/:id",
					middleware.RequirePermission("cmdb.server.delete"),
					cmdbController.DeleteServer)
			cmdb.GET("/servers/stats",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerStats)
			cmdb.POST("/servers/config",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerConfig)
			cmdb.GET("/servers/:id",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerByID)
			cmdb.GET("/servers/:id/connect",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.GetServerForConnect) // 新增：轻量级连接接口
			cmdb.POST("/servers/:id/connect",
					middleware.RequirePermission("cmdb.server.connect"),
					bastionController.ConnectServer)
			cmdb.GET("/servers/:id/permission",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.CheckConnectPermission)
			cmdb.POST("/servers/:id/sync-metrics",
					middleware.RequirePermission("cmdb.server.update"),
					cmdbController.SyncServerMetrics)

			// Agent 管理
			cmdb.POST("/servers/:id/agent/deploy",
					middleware.RequirePermission("cmdb.agents.deploy"),
					cmdbController.DeployAgent)
			cmdb.POST("/servers/:id/agent/restart",
					middleware.RequirePermission("cmdb.agents.restart"),
					cmdbController.RestartAgent)
			cmdb.POST("/servers/:id/agent/uninstall",
					middleware.RequirePermission("cmdb.agents.uninstall"),
					cmdbController.UninstallAgent)
			cmdb.GET("/servers/:id/agent/status",
					middleware.RequirePermission("cmdb.agents.list"),
					cmdbController.GetAgentStatus)
			cmdb.POST("/servers/:id/test-connection",
					middleware.RequirePermission("cmdb.server.update"),
					cmdbController.TestSSHConnection)

			// Agent 管理页面专用接口
			cmdb.GET("/agents",
					middleware.RequirePermission("cmdb.agents.list"),
					cmdbController.GetAgentList)
			cmdb.POST("/agents/batch-deploy",
					middleware.RequirePermission("cmdb.agents.deploy"),
					cmdbController.BatchDeployAgent)
			cmdb.POST("/agents/batch-uninstall",
					middleware.RequirePermission("cmdb.agents.uninstall"),
					cmdbController.BatchUninstallAgent)
			cmdb.DELETE("/agents/:id",
					middleware.RequirePermission("cmdb.agents.uninstall"),
					cmdbController.DeleteAgentRecord)

			// Agent 监控增强 API (P0) - 主机相关监控接口
			cmdb.GET("/servers/:id/extended-metrics",
					middleware.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerExtendedMetrics)
			cmdb.GET("/servers/:id/metrics/history",
					middleware.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerMetricsHistory)
			cmdb.GET("/servers/:id/hardware",
					middleware.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerHardware)
			cmdb.GET("/servers/:id/processes",
					middleware.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerProcesses)
			cmdb.GET("/servers/:id/services",
					middleware.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerServices)
			cmdb.GET("/servers/:id/network",
					middleware.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerNetwork)
			cmdb.GET("/servers/:id/security",
					middleware.RequirePermission("cmdb.server.view"),
					monitoringController.GetServerSecurity)

			// 主机分组管理
			cmdb.GET("/groups",
					middleware.RequirePermission("cmdb.group.list"),
					cmdbController.GetServerGroups)
			cmdb.GET("/groups/:id",
					middleware.RequirePermission("cmdb.group.view"),
					cmdbController.GetServerGroupByID)
			cmdb.GET("/asset-tree",
					middleware.RequirePermission("cmdb.group.list"),
					cmdbController.GetAssetTree)
			cmdb.POST("/groups",
					middleware.RequirePermission("cmdb.group.create"),
					cmdbController.CreateServerGroup)
			cmdb.PUT("/groups/:id",
					middleware.RequirePermission("cmdb.group.update"),
					cmdbController.UpdateServerGroup)
			cmdb.DELETE("/groups/:id",
					middleware.RequirePermission("cmdb.group.delete"),
					cmdbController.DeleteServerGroup)
			cmdb.POST("/groups/assign",
					middleware.RequirePermission("cmdb.group.assign"),
					cmdbController.AssignServerToGroup)
			cmdb.POST("/groups/assign-multi",
					middleware.RequirePermission("cmdb.group.assign"),
					cmdbController.AssignServerToGroups)
			cmdb.GET("/group-servers/:groupId",
					middleware.RequirePermission("cmdb.group.view"),
					cmdbController.GetServersByGroup)

			// 业务系统管理
			cmdb.GET("/business-units",
					middleware.RequirePermission("cmdb.business.list"),
					cmdbController.GetBusinessUnits)
			cmdb.POST("/business-units",
					middleware.RequirePermission("cmdb.business.create"),
					cmdbController.CreateBusinessUnit)
			cmdb.PUT("/business-units/:id",
					middleware.RequirePermission("cmdb.business.update"),
					cmdbController.UpdateBusinessUnit)
			cmdb.DELETE("/business-units/:id",
					middleware.RequirePermission("cmdb.business.delete"),
					cmdbController.DeleteBusinessUnit)

			// 机房机柜管理
			cmdb.GET("/rooms",
					middleware.RequirePermission("cmdb.rooms.list"),
					cmdbController.GetServerRooms)
			cmdb.POST("/rooms",
					middleware.RequirePermission("cmdb.rooms.create"),
					cmdbController.CreateServerRoom)
			cmdb.PUT("/rooms/:id",
					middleware.RequirePermission("cmdb.rooms.update"),
					cmdbController.UpdateServerRoom)
			cmdb.DELETE("/rooms/:id",
					middleware.RequirePermission("cmdb.rooms.delete"),
					cmdbController.DeleteServerRoom)
			cmdb.GET("/cabinets",
					middleware.RequirePermission("cmdb.rooms.list"),
					cmdbController.GetCabinets)

			// 标签管理
			cmdb.GET("/tags",
					middleware.RequirePermission("cmdb.tags.list"),
					cmdbController.GetServerTags)
			cmdb.POST("/tags",
					middleware.RequirePermission("cmdb.tags.create"),
					cmdbController.CreateServerTag)
			cmdb.PUT("/tags/:id",
					middleware.RequirePermission("cmdb.tags.update"),
					cmdbController.UpdateServerTag)
			cmdb.DELETE("/tags/:id",
					middleware.RequirePermission("cmdb.tags.delete"),
					cmdbController.DeleteServerTag)
			cmdb.POST("/tags/assign",
					middleware.RequirePermission("cmdb.tags.update"),
					cmdbController.AssignServerTag)
			cmdb.DELETE("/server-tags/:serverId/:tagId",
					middleware.RequirePermission("cmdb.tags.update"),
					cmdbController.RemoveServerTag)

			// SSH凭证管理
			cmdb.GET("/ssh-credentials",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.GetSSHCredentials)
			cmdb.GET("/ssh-credentials/:id",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.GetSSHCredentialByID)
			cmdb.POST("/ssh-credentials",
					middleware.RequirePermission("cmdb.server.update"),
					cmdbController.CreateSSHCredential)
			cmdb.PUT("/ssh-credentials/:id",
					middleware.RequirePermission("cmdb.server.update"),
					cmdbController.UpdateSSHCredential)
			cmdb.DELETE("/ssh-credentials/:id",
					middleware.RequirePermission("cmdb.server.delete"),
					cmdbController.DeleteSSHCredential)
			cmdb.POST("/ssh-credentials/:id/test",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.TestSSHCredential)

			// 资产变更记录
			cmdb.GET("/asset-changes",
					middleware.RequirePermission("cmdb.server.view"),
					cmdbController.GetAssetChanges)

			// ========== 堡垒机功能 ==========

			// 会话管理
			cmdb.GET("/sessions",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetSessions)
			cmdb.GET("/sessions/list",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionsList) // 轻量级列表接口，只返回展示字段
			cmdb.GET("/sessions/active",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetActiveSessions)
			cmdb.GET("/sessions/active-memory",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetActiveSessionsFromMemory) // 从内存获取真正活跃的会话
			cmdb.GET("/sessions/stats",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionStats)
			cmdb.GET("/sessions/:id",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionByID)
			cmdb.POST("/sessions/:id/terminate",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.TerminateSession)
			cmdb.GET("/sessions/:id/commands",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionCommands)
			cmdb.GET("/sessions/:id/file-transfers",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetSessionFileTransfers)
			cmdb.POST("/sessions/:id/resize",
					middleware.RequirePermission("cmdb.server.view"),
					sshHandler.ResizePTY)

			// WebSocket SSH 连接（不经过 Auth 中间件，由 handler 自行从 query param 验证 token）
			// 注意：此路由注册在 cmdb 组外，见下方

			// 命令审计
			cmdb.GET("/commands",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetCommands)

			// 文件传输审计
			cmdb.GET("/file-transfers",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetFileTransfers)

			// 访问策略管理
			cmdb.GET("/access-policies",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetAccessPolicies)
			cmdb.GET("/access-policies/:id",
					middleware.RequirePermission("cmdb.server.view"),
					bastionController.GetAccessPolicyByID)
			cmdb.POST("/access-policies",
					middleware.RequirePermission("cmdb.server.update"),
					bastionController.CreateAccessPolicy)
			cmdb.PUT("/access-policies/:id",
					middleware.RequirePermission("cmdb.server.update"),
					bastionController.UpdateAccessPolicy)
			cmdb.DELETE("/access-policies/:id",
					middleware.RequirePermission("cmdb.server.delete"),
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
		routeGroup.Use(middleware.Auth())
		{
			routeGroup.GET("/getUserRoutes", routeController.GetUserRoutes)
			routeGroup.GET("/isRouteExist", routeController.IsRouteExist)
			routeGroup.POST("/invalidateCache", routeController.InvalidateCache)
			routeGroup.GET("/debugCache", middleware.Auth(), routeController.DebugCache)
		}

			// K8s 集群管理路由（需要认证）
			k8s := api.Group("/k8s")
			k8s.Use(middleware.Auth())
			{
				// 集群管理
				k8s.GET("/clusters",
					middleware.RequirePermission("k8s.cluster.list"),
					k8sClusterController.GetClusters)
				k8s.POST("/clusters",
					middleware.RequirePermission("k8s.cluster.create"),
					k8sClusterController.CreateCluster)
				k8s.GET("/clusters/:id",
					middleware.RequirePermission("k8s.cluster.view"),
					k8sClusterController.GetClusterByID)
				k8s.PUT("/clusters/:id",
					middleware.RequirePermission("k8s.cluster.update"),
					k8sClusterController.UpdateCluster)
				k8s.DELETE("/clusters/:id",
					middleware.RequirePermission("k8s.cluster.delete"),
					k8sClusterController.DeleteCluster)

				// 连接测试
				k8s.POST("/clusters/:id/test",
					middleware.RequirePermission("k8s.cluster.connect"),
					k8sClusterController.TestConnection)

				// 集群节点和命名空间
				k8s.GET("/clusters/:id/nodes",
					middleware.RequirePermission("k8s.cluster.view"),
					k8sClusterController.GetClusterNodes)
				k8s.GET("/clusters/:id/namespaces",
					middleware.RequirePermission("k8s.cluster.view"),
					k8sClusterController.GetClusterNamespaces)

				// 集群用户管理
				k8s.GET("/clusters/:id/users",
					middleware.RequirePermission("k8s.permission.list"),
					k8sClusterController.GetClusterUsers)

				// 权限管理
				k8s.POST("/clusters/:id/permissions",
					middleware.RequirePermission("k8s.permission.assign"),
					k8sPermissionController.AssignClusterRole)
				k8s.DELETE("/clusters/:id/permissions/:userId",
					middleware.RequirePermission("k8s.permission.revoke"),
					k8sPermissionController.RevokeClusterRole)
				k8s.GET("/users/clusters",
					middleware.RequirePermission("k8s.permission.list"),
					k8sPermissionController.GetUserClusters)
				k8s.GET("/clusters/:id/users/:userId/role",
					middleware.RequirePermission("k8s.permission.list"),
					k8sPermissionController.GetUserRoleInCluster)
				k8s.POST("/permissions/batch-assign",
					middleware.RequirePermission("k8s.permission.assign"),
					k8sPermissionController.BatchAssignClusterRoles)
					// Workloads - Deployments
					k8s.GET("/clusters/:id/deployments",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListDeployments)
					k8s.GET("/clusters/:id/deployments/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDeployment)
					k8s.GET("/clusters/:id/deployments/:namespace/:name/pods",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDeploymentPods)
					k8s.POST("/clusters/:id/deployments",
					middleware.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateDeployment)
					k8s.PUT("/clusters/:id/deployments",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateDeployment)
					k8s.DELETE("/clusters/:id/deployments",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteDeployment)
					k8s.POST("/clusters/:id/deployments/scale",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.ScaleDeployment)
					k8s.POST("/clusters/:id/deployments/restart",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.RestartDeployment)

					// Workloads - StatefulSets
					k8s.GET("/clusters/:id/statefulsets",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListStatefulSets)
					k8s.GET("/clusters/:id/statefulsets/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetStatefulSet)
					k8s.GET("/clusters/:id/statefulsets/:namespace/:name/pods",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetStatefulSetPods)
					// k8s.POST("/clusters/:id/statefulsets", k8sResourceController.CreateStatefulSet)
					// k8s.PUT("/clusters/:id/statefulsets", k8sResourceController.UpdateStatefulSet)
					// k8s.DELETE("/clusters/:id/statefulsets", k8sResourceController.DeleteStatefulSet)
					// k8s.POST("/clusters/:id/statefulsets/restart", k8sResourceController.RestartStatefulSet)

					// Workloads - DaemonSets
					k8s.GET("/clusters/:id/daemonsets",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListDaemonSets)
					k8s.GET("/clusters/:id/daemonsets/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDaemonSet)
					k8s.GET("/clusters/:id/daemonsets/:namespace/:name/pods",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetDaemonSetPods)
					// k8s.POST("/clusters/:id/daemonsets", k8sResourceController.CreateDaemonSet)
					// k8s.PUT("/clusters/:id/daemonsets", k8sResourceController.UpdateDaemonSet)
					// k8s.DELETE("/clusters/:id/daemonsets", k8sResourceController.DeleteDaemonSet)
					// k8s.POST("/clusters/:id/daemonsets/restart", k8sResourceController.RestartDaemonSet)

					// Workloads - Jobs
					k8s.GET("/clusters/:id/jobs",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListJobs)
					k8s.GET("/clusters/:id/jobs/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetJob)
					k8s.GET("/clusters/:id/jobs/:namespace/:name/pods",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetJobPods)
					k8s.DELETE("/clusters/:id/jobs",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteJob)

					// Workloads - CronJobs
					k8s.GET("/clusters/:id/cronjobs",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListCronJobs)
					k8s.GET("/clusters/:id/cronjobs/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetCronJob)
					k8s.GET("/clusters/:id/cronjobs/:namespace/:name/pods",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetCronJobPods)
					k8s.DELETE("/clusters/:id/cronjobs",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteCronJob)
					k8s.PUT("/clusters/:id/cronjobs/suspend",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.SuspendCronJob)

					// Services
					k8s.GET("/clusters/:id/services",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListServices)
					k8s.GET("/clusters/:id/services/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetService)
					k8s.POST("/clusters/:id/services",
					middleware.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateService)
					k8s.PUT("/clusters/:id/services",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateService)
					k8s.DELETE("/clusters/:id/services",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteService)

					// Ingresses
					k8s.GET("/clusters/:id/ingresses",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListIngress)
					k8s.GET("/clusters/:id/ingresses/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetIngress)
					k8s.POST("/clusters/:id/ingresses",
					middleware.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateIngress)
					k8s.PUT("/clusters/:id/ingresses",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateIngress)
					k8s.DELETE("/clusters/:id/ingresses",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteIngress)

					// Pods
					k8s.GET("/clusters/:id/pods",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListPods)
					k8s.GET("/clusters/:id/pods/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetPod)
					k8s.PUT("/clusters/:id/pods",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdatePod)
					k8s.GET("/clusters/:id/pods/:namespace/:name/logs",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetPodLogs)
					k8s.DELETE("/clusters/:id/pods",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeletePod)

					// ConfigMaps
					k8s.GET("/clusters/:id/configmaps",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListConfigMaps)
					k8s.GET("/clusters/:id/configmaps/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetConfigMap)
					k8s.POST("/clusters/:id/configmaps",
					middleware.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateConfigMap)
					k8s.PUT("/clusters/:id/configmaps",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateConfigMap)
					k8s.DELETE("/clusters/:id/configmaps",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteConfigMap)

					// Secrets
					k8s.GET("/clusters/:id/secrets",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListSecrets)
					k8s.GET("/clusters/:id/secrets/:namespace/:name",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.GetSecret)
					k8s.POST("/clusters/:id/secrets",
					middleware.RequirePermission("k8s.resource.create"),
					k8sResourceController.CreateSecret)
					k8s.PUT("/clusters/:id/secrets",
					middleware.RequirePermission("k8s.resource.update"),
					k8sResourceController.UpdateSecret)
					k8s.DELETE("/clusters/:id/secrets",
					middleware.RequirePermission("k8s.resource.delete"),
					k8sResourceController.DeleteSecret)

					// Events
					k8s.GET("/clusters/:id/events",
					middleware.RequirePermission("k8s.resource.view"),
					k8sResourceController.ListEvents)

					// K8s 终端管理（需要认证）
					k8s.GET("/terminal/active",
					middleware.RequirePermission("k8s.cluster.view"),
					k8sTerminalHandler.GetActiveSessions)
					k8s.POST("/terminal/sessions/:sessionId/terminate",
					middleware.RequirePermission("k8s.cluster.view"),
					k8sTerminalHandler.TerminateSession)

					// K8s 诊断功能（需要认证）
					k8s.GET("/diagnostic/commands",
					middleware.RequirePermission("k8s.cluster.view"),
					diagnosticController.GetDiagnosticCommands)
					k8s.GET("/diagnostic/pods/:clusterId/:namespace",
					middleware.RequirePermission("k8s.resource.view"),
					diagnosticController.GetJavaPods)
					k8s.GET("/diagnostic/namespaces/:clusterId",
					middleware.RequirePermission("k8s.cluster.view"),
					diagnosticController.GetNamespaces)
					k8s.POST("/diagnostic/execute",
					middleware.RequirePermission("k8s.resource.view"),
					diagnosticController.ExecuteDiagnostic)
					k8s.GET("/diagnostic/history",
					middleware.RequirePermission("k8s.resource.view"),
					diagnosticController.GetDiagnosticHistory)

			}

	}
}
