package routes

import (
	"oneops/backend/controller"
	"oneops/backend/handler"
	"oneops/backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupCMDBRoutes 设置CMDB资产管理相关路由
func SetupCMDBRoutes(
	r *gin.Engine,
	cmdbController *controller.CMDBController,
	bastionController *controller.BastionController,
	attributeController *controller.AttributeController,
	sshHandler *handler.SSHWebSocketHandler,
) {
	api := r.Group("/api")

	// WebSocket SSH 连接（不经过 Auth 中间件，handler 自行从 query param 验证 token）
	api.GET("/cmdb/sessions/:id/ws", func(ctx *gin.Context) {
		sshHandler.HandleWebSocket(ctx)
	})

	// Agent 心跳（不经过 Auth 中间件，由 Agent 直接上报）
	api.POST("/cmdb/agent/heartbeat", cmdbController.ReceiveAgentHeartbeat)

	// Agent 版本管理接口（需要认证）
	agentVersionGroup := api.Group("/cmdb/agent-versions")
	agentVersionGroup.Use(middleware.Auth())
	agentVersionGroup.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		agentVersionGroup.GET("", cmdbController.GetAgentVersions)
		agentVersionGroup.GET("/latest", cmdbController.GetLatestAgentVersion)
		agentVersionGroup.GET("/:id", cmdbController.GetAgentVersionByID)
		agentVersionGroup.POST("", cmdbController.CreateAgentVersion)
		agentVersionGroup.PUT("/:id", cmdbController.UpdateAgentVersion)
		agentVersionGroup.DELETE("/:id", cmdbController.DeleteAgentVersion)
	}

	// Agent 升级管理接口（需要认证）
	agentUpgradeGroup := api.Group("/cmdb")
	agentUpgradeGroup.Use(middleware.Auth())
	agentUpgradeGroup.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		agentUpgradeGroup.POST("/servers/:id/agent/upgrade", cmdbController.UpgradeAgent)
		agentUpgradeGroup.GET("/agent-upgrade-tasks", cmdbController.GetUpgradeTasks)
		agentUpgradeGroup.GET("/agent-upgrade-tasks/:id", cmdbController.GetUpgradeTaskByID)
	}

	// CMDB 主路由组（需要认证）
	cmdb := api.Group("/cmdb")
	cmdb.Use(middleware.Auth())
	cmdb.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		// 服务器管理
		cmdb.GET("/servers", cmdbController.GetServers)
		cmdb.POST("/servers", cmdbController.CreateServer)
		cmdb.PUT("/servers/:id", cmdbController.UpdateServer)
		cmdb.DELETE("/servers/:id", cmdbController.DeleteServer)
		cmdb.GET("/servers/stats", cmdbController.GetServerStats)
		cmdb.POST("/servers/config", cmdbController.GetServerConfig)
		cmdb.GET("/servers/:id", cmdbController.GetServerByID)
		cmdb.GET("/servers/:id/connect", cmdbController.GetServerForConnect)
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

		// 堡垒机功能 - 会话管理
		cmdb.GET("/sessions", bastionController.GetSessions)
		cmdb.GET("/sessions/list", bastionController.GetSessionsList)
		cmdb.GET("/sessions/active", bastionController.GetActiveSessions)
		cmdb.GET("/sessions/active-memory", bastionController.GetActiveSessionsFromMemory)
		cmdb.GET("/sessions/stats", bastionController.GetSessionStats)
		cmdb.GET("/sessions/:id", bastionController.GetSessionByID)
		cmdb.POST("/sessions/:id/terminate", bastionController.TerminateSession)
		cmdb.GET("/sessions/:id/commands", bastionController.GetSessionCommands)
		cmdb.GET("/sessions/:id/file-transfers", bastionController.GetSessionFileTransfers)
		cmdb.POST("/sessions/:id/resize", sshHandler.ResizePTY)

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
}
