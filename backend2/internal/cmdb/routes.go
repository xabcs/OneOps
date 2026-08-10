package cmdb

import (
	"oneops/backend2/internal/system"
	"oneops/backend2/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupCMDBRoutes 设置CMDB资产管理相关路由
func SetupCMDBRoutes(
	r *gin.Engine,
	cmdbController *CMDBController,
	bastionController *BastionController,
	attributeController *system.AttributeController,
	sshHandler *SSHWebSocketHandler,
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
	agentVersionGroup.Use(system.RequirePermissionFromDB()) // ✅ 统一权限检查
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
	agentUpgradeGroup.Use(system.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		agentUpgradeGroup.POST("/servers/:id/agent/upgrade", cmdbController.UpgradeAgent)
		agentUpgradeGroup.GET("/agent-upgrade-tasks", cmdbController.GetUpgradeTasks)
		agentUpgradeGroup.GET("/agent-upgrade-tasks/:id", cmdbController.GetUpgradeTaskByID)
	}

	// CMDB 主路由组（需要认证）
	cmdb := api.Group("/cmdb")
	Use(middleware.Auth())
	Use(system.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		// 服务器管理
		GET("/servers", cmdbController.GetServers)
		POST("/servers", cmdbController.CreateServer)
		PUT("/servers/:id", cmdbController.UpdateServer)
		DELETE("/servers/:id", cmdbController.DeleteServer)
		GET("/servers/stats", cmdbController.GetServerStats)
		POST("/servers/config", cmdbController.GetServerConfig)
		GET("/servers/:id", cmdbController.GetServerByID)
		GET("/servers/:id/connect", cmdbController.GetServerForConnect)
		POST("/servers/:id/connect", bastionController.ConnectServer)
		GET("/servers/:id/permission", bastionController.CheckConnectPermission)
		POST("/servers/:id/sync-metrics", cmdbController.SyncServerMetrics)

		// Agent 管理
		POST("/servers/:id/agent/deploy", cmdbController.DeployAgent)
		POST("/servers/:id/agent/restart", cmdbController.RestartAgent)
		POST("/servers/:id/agent/uninstall", cmdbController.UninstallAgent)
		GET("/servers/:id/agent/status", cmdbController.GetAgentStatus)
		POST("/servers/:id/test-connection", cmdbController.TestSSHConnection)

		// Agent 管理页面专用接口
		GET("/agents", cmdbController.GetAgentList)
		POST("/agents/batch-deploy", cmdbController.BatchDeployAgent)
		POST("/agents/batch-uninstall", cmdbController.BatchUninstallAgent)
		DELETE("/agents/:id", cmdbController.DeleteAgentRecord)

		// 主机分组管理
		GET("/groups", cmdbController.GetServerGroups)
		GET("/groups/:id", cmdbController.GetServerGroupByID)
		GET("/asset-tree", cmdbController.GetAssetTree)
		POST("/groups", cmdbController.CreateServerGroup)
		PUT("/groups/:id", cmdbController.UpdateServerGroup)
		DELETE("/groups/:id", cmdbController.DeleteServerGroup)
		POST("/groups/assign", cmdbController.AssignServerToGroup)
		POST("/groups/assign-multi", cmdbController.AssignServerToGroups)
		GET("/group-servers/:groupId", cmdbController.GetServersByGroup)

		// 业务系统管理
		GET("/business-units", cmdbController.GetBusinessUnits)
		POST("/business-units", cmdbController.CreateBusinessUnit)
		PUT("/business-units/:id", cmdbController.UpdateBusinessUnit)
		DELETE("/business-units/:id", cmdbController.DeleteBusinessUnit)

		// 机房机柜管理
		GET("/rooms", cmdbController.GetServerRooms)
		POST("/rooms", cmdbController.CreateServerRoom)
		PUT("/rooms/:id", cmdbController.UpdateServerRoom)
		DELETE("/rooms/:id", cmdbController.DeleteServerRoom)
		GET("/cabinets", cmdbController.GetCabinets)

		// 标签管理
		GET("/tags", cmdbController.GetServerTags)
		POST("/tags", cmdbController.CreateServerTag)
		PUT("/tags/:id", cmdbController.UpdateServerTag)
		DELETE("/tags/:id", cmdbController.DeleteServerTag)
		POST("/tags/assign", cmdbController.AssignServerTag)
		DELETE("/server-tags/:serverId/:tagId", cmdbController.RemoveServerTag)

		// SSH凭证管理
		GET("/ssh-credentials", cmdbController.GetSSHCredentials)
		GET("/ssh-credentials/:id", cmdbController.GetSSHCredentialByID)
		POST("/ssh-credentials", cmdbController.CreateSSHCredential)
		PUT("/ssh-credentials/:id", cmdbController.UpdateSSHCredential)
		DELETE("/ssh-credentials/:id", cmdbController.DeleteSSHCredential)
		POST("/ssh-credentials/:id/test", cmdbController.TestSSHCredential)

		// 资产变更记录
		GET("/asset-changes", cmdbController.GetAssetChanges)

		// 堡垒机功能 - 会话管理
		GET("/sessions", bastionController.GetSessions)
		GET("/sessions/list", bastionController.GetSessionsList)
		GET("/sessions/active", bastionController.GetActiveSessions)
		GET("/sessions/active-memory", bastionController.GetActiveSessionsFromMemory)
		GET("/sessions/stats", bastionController.GetSessionStats)
		GET("/sessions/:id", bastionController.GetSessionByID)
		POST("/sessions/:id/terminate", bastionController.TerminateSession)
		GET("/sessions/:id/commands", bastionController.GetSessionCommands)
		GET("/sessions/:id/file-transfers", bastionController.GetSessionFileTransfers)
		POST("/sessions/:id/resize", sshHandler.ResizePTY)

		// 命令审计
		GET("/commands", bastionController.GetCommands)

		// 文件传输审计
		GET("/file-transfers", bastionController.GetFileTransfers)

		// 访问策略管理
		GET("/access-policies", bastionController.GetAccessPolicies)
		GET("/access-policies/:id", bastionController.GetAccessPolicyByID)
		POST("/access-policies", bastionController.CreateAccessPolicy)
		PUT("/access-policies/:id", bastionController.UpdateAccessPolicy)
		DELETE("/access-policies/:id", bastionController.DeleteAccessPolicy)
	}
}
