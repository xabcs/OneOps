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
			system.GET("/menus/tree", menuController.GetMenuTree)
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
			system.PUT("/users/:id/password", userController.ResetPassword)

			// 属性管理
			system.GET("/attributes", attributeController.GetAttributeDefinitions)
			system.POST("/attributes", attributeController.CreateAttributeDefinition)
			system.PUT("/attributes/:id", attributeController.UpdateAttributeDefinition)
			system.DELETE("/attributes/:id", attributeController.DeleteAttributeDefinition)
			system.GET("/attributes/:id", attributeController.GetAttributeDefinitionByID)

			// 主机属性管理
			system.GET("/server-attributes/:serverId", attributeController.GetServerAttributes)
			system.POST("/server-attributes/:serverId", attributeController.SaveServerAttributes)
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
			monitoringAuth.GET("/grafana/url", monitoringController.GetGrafanaUrl)
			// 监控数据
			monitoringAuth.GET("/stats", monitoringController.GetMonitoringStats)
			// 刷新监控数据
			monitoringAuth.POST("/refresh", monitoringController.RefreshMonitoring)
			monitoringAuth.POST("/alert/handle", monitoringController.HandleAlert)

			// Agent 监控增强 API (P0)
			monitoring.GET("/overview", monitoringController.GetOverview)
			monitoring.GET("/alerts", monitoringController.GetAlerts)
			monitoring.POST("/alerts/:id/acknowledge", monitoringController.AcknowledgeAlert)
			monitoring.GET("/alerts/stats", monitoringController.GetAlertStats)

				// 告警规则管理 API (P0 - 动态配置)
			monitoring.GET("/alerts/rules", monitoringController.GetAlertRules)
			monitoring.POST("/alerts/rules", monitoringController.CreateAlertRule)
			monitoring.PUT("/alerts/rules/:id", monitoringController.UpdateAlertRule)
			monitoring.DELETE("/alerts/rules/:id", monitoringController.DeleteAlertRule)
			monitoring.PUT("/alerts/rules/:id/status", monitoringController.UpdateAlertRuleStatus)

			// 通知渠道管理 API (P2)
			monitoring.GET("/notifications/channels", monitoringController.GetNotificationChannels)
			monitoring.POST("/notifications/channels", monitoringController.CreateNotificationChannel)
			monitoring.PUT("/notifications/channels/:id", monitoringController.UpdateNotificationChannel)
			monitoring.DELETE("/notifications/channels/:id", monitoringController.DeleteNotificationChannel)
			monitoring.POST("/notifications/channels/:id/test", monitoringController.TestNotificationChannel)

				// 巡检报告 API (P3)
				monitoring.GET("/reports", monitoringController.GetReports)
				monitoring.POST("/reports", monitoringController.CreateReport)
				monitoring.GET("/reports/:id", monitoringController.GetReportDetail)
				monitoring.GET("/reports/:id/export", monitoringController.ExportReport)
				monitoring.DELETE("/reports/:id", monitoringController.DeleteReport)
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
			cmdb.GET("/servers/:id/connect", cmdbController.GetServerForConnect) // 新增：轻量级连接接口
			cmdb.POST("/servers/:id/connect", bastionController.ConnectServer)
			cmdb.GET("/servers/:id/permission", bastionController.CheckConnectPermission)
			cmdb.POST("/servers/:id/sync-metrics", cmdbController.SyncServerMetrics)

			// Agent 管理
			cmdb.POST("/servers/:id/agent/deploy", cmdbController.DeployAgent)
			cmdb.POST("/servers/:id/agent/restart", cmdbController.RestartAgent)
			cmdb.POST("/servers/:id/agent/uninstall", cmdbController.UninstallAgent)
			cmdb.GET("/servers/:id/agent/status", cmdbController.GetAgentStatus)
			cmdb.POST("/servers/:id/test-connection", cmdbController.TestSSHConnection)

			// Agent 管理页面专用接口
			cmdb.GET("/agents", cmdbController.GetAgentList)
			cmdb.POST("/agents/batch-deploy", cmdbController.BatchDeployAgent)
			cmdb.POST("/agents/batch-uninstall", cmdbController.BatchUninstallAgent)
			cmdb.DELETE("/agents/:id", cmdbController.DeleteAgentRecord)

			// Agent 监控增强 API (P0) - 主机相关监控接口
			cmdb.GET("/servers/:id/extended-metrics", monitoringController.GetServerExtendedMetrics)
			cmdb.GET("/servers/:id/metrics/history", monitoringController.GetServerMetricsHistory)
			cmdb.GET("/servers/:id/hardware", monitoringController.GetServerHardware)
			cmdb.GET("/servers/:id/processes", monitoringController.GetServerProcesses)
			cmdb.GET("/servers/:id/services", monitoringController.GetServerServices)
			cmdb.GET("/servers/:id/network", monitoringController.GetServerNetwork)
			cmdb.GET("/servers/:id/security", monitoringController.GetServerSecurity)

			// 主机分组管理
			cmdb.GET("/groups", cmdbController.GetServerGroups)
			cmdb.GET("/groups/:id", cmdbController.GetServerGroupByID)
			cmdb.GET("/asset-tree", cmdbController.GetAssetTree)
			cmdb.POST("/groups", cmdbController.CreateServerGroup)
			cmdb.PUT("/groups/:id", cmdbController.UpdateServerGroup)
			cmdb.DELETE("/groups/:id", cmdbController.DeleteServerGroup)
			cmdb.POST("/groups/assign", cmdbController.AssignServerToGroup)
			cmdb.POST("/groups/assign-multi", cmdbController.AssignServerToGroups)
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
			cmdb.GET("/sessions/list", bastionController.GetSessionsList) // 轻量级列表接口，只返回展示字段
			cmdb.GET("/sessions/active", bastionController.GetActiveSessions)
			cmdb.GET("/sessions/active-memory", bastionController.GetActiveSessionsFromMemory) // 从内存获取真正活跃的会话
			cmdb.GET("/sessions/stats", bastionController.GetSessionStats)
			cmdb.GET("/sessions/:id", bastionController.GetSessionByID)
			cmdb.POST("/sessions/:id/terminate", bastionController.TerminateSession)
			cmdb.GET("/sessions/:id/commands", bastionController.GetSessionCommands)
			cmdb.GET("/sessions/:id/file-transfers", bastionController.GetSessionFileTransfers)
			cmdb.POST("/sessions/:id/resize", sshHandler.ResizePTY)

			// WebSocket SSH 连接（不经过 Auth 中间件，由 handler 自行从 query param 验证 token）
			// 注意：此路由注册在 cmdb 组外，见下方

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
		}

			// K8s 集群管理路由（需要认证）
			k8s := api.Group("/k8s")
			k8s.Use(middlewares.Auth())
			{
				// 集群管理
				k8s.GET("/clusters", k8sClusterController.GetClusters)
				k8s.POST("/clusters", k8sClusterController.CreateCluster)
				k8s.GET("/clusters/:id", k8sClusterController.GetClusterByID)
				k8s.PUT("/clusters/:id", k8sClusterController.UpdateCluster)
				k8s.DELETE("/clusters/:id", k8sClusterController.DeleteCluster)

				// 连接测试
				k8s.POST("/clusters/:id/test", k8sClusterController.TestConnection)

				// 集群节点和命名空间
				k8s.GET("/clusters/:id/nodes", k8sClusterController.GetClusterNodes)
				k8s.GET("/clusters/:id/namespaces", k8sClusterController.GetClusterNamespaces)

				// 集群用户管理
				k8s.GET("/clusters/:id/users", k8sClusterController.GetClusterUsers)

				// 权限管理
				k8s.POST("/clusters/:id/permissions", k8sPermissionController.AssignClusterRole)
				k8s.DELETE("/clusters/:id/permissions/:userId", k8sPermissionController.RevokeClusterRole)
				k8s.GET("/users/clusters", k8sPermissionController.GetUserClusters)
				k8s.GET("/clusters/:id/users/:userId/role", k8sPermissionController.GetUserRoleInCluster)
				k8s.POST("/permissions/batch-assign", k8sPermissionController.BatchAssignClusterRoles)
					// Workloads - Deployments
					k8s.GET("/clusters/:id/deployments", k8sResourceController.ListDeployments)
					k8s.GET("/clusters/:id/deployments/:namespace/:name", k8sResourceController.GetDeployment)
					k8s.POST("/clusters/:id/deployments", k8sResourceController.CreateDeployment)
					k8s.PUT("/clusters/:id/deployments", k8sResourceController.UpdateDeployment)
					k8s.DELETE("/clusters/:id/deployments", k8sResourceController.DeleteDeployment)
					k8s.POST("/clusters/:id/deployments/scale", k8sResourceController.ScaleDeployment)
					k8s.POST("/clusters/:id/deployments/restart", k8sResourceController.RestartDeployment)

					// Workloads - StatefulSets
					k8s.GET("/clusters/:id/statefulsets", k8sResourceController.ListStatefulSets)
					k8s.GET("/clusters/:id/statefulsets/:namespace/:name", k8sResourceController.GetStatefulSet)

					// Workloads - DaemonSets
					k8s.GET("/clusters/:id/daemonsets", k8sResourceController.ListDaemonSets)
					k8s.GET("/clusters/:id/daemonsets/:namespace/:name", k8sResourceController.GetDaemonSet)

					// Services
					k8s.GET("/clusters/:id/services", k8sResourceController.ListServices)
					k8s.GET("/clusters/:id/services/:namespace/:name", k8sResourceController.GetService)
					k8s.POST("/clusters/:id/services", k8sResourceController.CreateService)
					k8s.PUT("/clusters/:id/services", k8sResourceController.UpdateService)
					k8s.DELETE("/clusters/:id/services", k8sResourceController.DeleteService)

					// Pods
					k8s.GET("/clusters/:id/pods", k8sResourceController.ListPods)
					k8s.GET("/clusters/:id/pods/:namespace/:name", k8sResourceController.GetPod)
					k8s.GET("/clusters/:id/pods/:namespace/:name/logs", k8sResourceController.GetPodLogs)
					k8s.DELETE("/clusters/:id/pods", k8sResourceController.DeletePod)

					// ConfigMaps
					k8s.GET("/clusters/:id/configmaps", k8sResourceController.ListConfigMaps)
					k8s.GET("/clusters/:id/configmaps/:namespace/:name", k8sResourceController.GetConfigMap)
					k8s.POST("/clusters/:id/configmaps", k8sResourceController.CreateConfigMap)
					k8s.PUT("/clusters/:id/configmaps", k8sResourceController.UpdateConfigMap)
					k8s.DELETE("/clusters/:id/configmaps", k8sResourceController.DeleteConfigMap)

					// Secrets
					k8s.GET("/clusters/:id/secrets", k8sResourceController.ListSecrets)
					k8s.GET("/clusters/:id/secrets/:namespace/:name", k8sResourceController.GetSecret)
					k8s.POST("/clusters/:id/secrets", k8sResourceController.CreateSecret)
					k8s.PUT("/clusters/:id/secrets", k8sResourceController.UpdateSecret)
					k8s.DELETE("/clusters/:id/secrets", k8sResourceController.DeleteSecret)

					// Events
					k8s.GET("/clusters/:id/events", k8sResourceController.ListEvents)

					// K8s 终端管理（需要认证）
					k8s.GET("/terminal/active", k8sTerminalHandler.GetActiveSessions)
					k8s.POST("/terminal/sessions/:sessionId/terminate", k8sTerminalHandler.TerminateSession)

			}

	}
}
