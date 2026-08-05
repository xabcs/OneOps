package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/handlers"
	"oneops/backend/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupCMDBRoutes 设置CMDB资产管理相关路由
func SetupCMDBRoutes(
	r *gin.Engine,
	cmdbController *controllers.CMDBController,
	bastionController *controllers.BastionController,
	attributeController *controllers.AttributeController,
	sshHandler *handlers.SSHWebSocketHandler,
) {
	api := r.Group("/api")
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

		// Agent 监控增强 API (功能待实现)
		// cmdb.GET("/servers/:id/extended-metrics", cmdbController.GetServerExtendedMetrics)
		// cmdb.GET("/servers/:id/metrics/history", cmdbController.GetServerMetricsHistory)
		// cmdb.GET("/servers/:id/hardware", cmdbController.GetServerHardware)
		// cmdb.GET("/servers/:id/processes", cmdbController.GetServerProcesses)
		// cmdb.GET("/servers/:id/services", cmdbController.GetServerServices)
		// cmdb.GET("/servers/:id/network", cmdbController.GetServerNetwork)
		// cmdb.GET("/servers/:id/security", cmdbController.GetServerSecurity)

		// 主机分组管理
		cmdb.GET("/groups", cmdbController.GetServerGroups)
		cmdb.GET("/groups/:id", cmdbController.GetServerGroupByID)
		cmdb.GET("/asset-tree", cmdbController.GetAssetTree)
		cmdb.POST("/groups", cmdbController.CreateServerGroup)
		cmdb.PUT("/groups/:id", cmdbController.UpdateServerGroup)
		cmdb.DELETE("/groups/:id", cmdbController.DeleteServerGroup)
		cmdb.POST("/groups/assign", cmdbController.AssignServerToGroup)
		// cmdb.POST("/groups/assign-multi", cmdbController.AssignServerToServers) // 功能待实现
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

		// 堡垒机功能
		// setupBastionRoutes(cmdb, bastionController, sshHandler) // 函数未定义，注释掉
	}

	// WebSocket SSH 连接（不经过 Auth 中间件）
	api.GET("/cmdb/sessions/:id/ws", func(ctx *gin.Context) {
		sshHandler.HandleWebSocket(ctx)
	})

	// Agent 心跳（不经过 Auth 中间件）
	api.POST("/cmdb/agent/heartbeat", cmdbController.ReceiveAgentHeartbeat)
}
