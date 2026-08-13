package routes

import (
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/middleware"

	cmdbsvc "oneops/backend3/service/cmdb"

	cmdbctrl "oneops/backend3/controller/cmdb"
	repocmdb "oneops/backend3/repository/cmdb"

	"github.com/gin-gonic/gin"
)

// SetupCMDBRoutes 设置CMDB资产管理相关路由
func SetupCMDBRoutes(r *gin.Engine) {
	db := database.GetDB()

	// 创建 repositories
	serverRepo := repocmdb.NewServerRepository(db)
	bastionRepo := repocmdb.NewBastionRepository(db)
	agentRepo := repocmdb.NewAgentRepository(db)
	attributeRepo := repocmdb.NewAttributeRepository(db)

	// 创建 services
	cmdbSvc := cmdbsvc.NewCMDBService(serverRepo)
	bastionSvc := cmdbsvc.NewBastionService(bastionRepo)
	agentSvc := cmdbsvc.NewAgentService(agentRepo)
	attributeSvc := cmdbsvc.NewAttributeService(attributeRepo)

	// 创建 controllers
	cmdbController := cmdbctrl.NewCMDBController(cmdbSvc, agentSvc)
	bastionController := cmdbctrl.NewBastionController(bastionSvc)
	agentController := cmdbctrl.NewAgentController(agentSvc, cmdbSvc)
	attributeController := cmdbctrl.NewAttributeController(attributeSvc)

	api := r.Group("/api")

	// Agent 心跳（不经过 Auth 中间件，由 Agent 直接上报）
	api.POST("/cmdb/agent/heartbeat", agentController.ReceiveAgentHeartbeat)

	// Agent 版本管理接口（需要认证）
	agentVersionGroup := api.Group("/cmdb/agent-versions")
	agentVersionGroup.Use(middleware.Auth())
	agentVersionGroup.Use(middleware.RequirePermissionFromDB())
	{
		agentVersionGroup.GET("", agentController.GetAgentVersions)
		agentVersionGroup.GET("/latest", agentController.GetLatestAgentVersion)
		agentVersionGroup.GET("/:id", agentController.GetAgentVersionByID)
		agentVersionGroup.POST("", agentController.CreateAgentVersion)
		agentVersionGroup.PUT("/:id", agentController.UpdateAgentVersion)
		agentVersionGroup.DELETE("/:id", agentController.DeleteAgentVersion)
	}

	// Agent 升级管理接口（需要认证）
	agentUpgradeGroup := api.Group("/cmdb")
	agentUpgradeGroup.Use(middleware.Auth())
	agentUpgradeGroup.Use(middleware.RequirePermissionFromDB())
	{
		agentUpgradeGroup.POST("/servers/:id/agent/upgrade", agentController.UpgradeAgent)
		agentUpgradeGroup.GET("/agent-upgrade-tasks", agentController.GetUpgradeTasks)
		agentUpgradeGroup.GET("/agent-upgrade-tasks/:id", agentController.GetUpgradeTaskByID)
	}

	// CMDB 主路由组（需要认证）
	cmdb := api.Group("/cmdb")
	cmdb.Use(middleware.Auth())
	cmdb.Use(middleware.RequirePermissionFromDB())
	{
		// 服务器管理
		cmdb.GET("/servers", cmdbController.GetServers)
		cmdb.POST("/servers", cmdbController.CreateServer)
		cmdb.PUT("/servers/:id", cmdbController.UpdateServer)
		cmdb.DELETE("/servers/:id", cmdbController.DeleteServer)
		cmdb.GET("/servers/stats", cmdbController.GetServerStats)
		cmdb.GET("/servers/options", cmdbController.GetServerOptions)
		cmdb.POST("/servers/config", cmdbController.GetServerConfig)
		cmdb.GET("/servers/:id", cmdbController.GetServerByID)
		cmdb.GET("/servers/:id/connect", cmdbController.GetServerForConnect)
		cmdb.POST("/servers/:id/connect", bastionController.ConnectServer)
		cmdb.GET("/servers/:id/permission", bastionController.CheckConnectPermission)
		cmdb.POST("/servers/:id/sync-metrics", cmdbController.SyncServerMetrics)

		// Agent 管理
		cmdb.POST("/servers/:id/agent/deploy", agentController.DeployAgent)
		cmdb.POST("/servers/:id/agent/restart", agentController.RestartAgent)
		cmdb.POST("/servers/:id/agent/uninstall", agentController.UninstallAgent)
		cmdb.GET("/servers/:id/agent/status", agentController.GetAgentStatus)
		cmdb.POST("/servers/:id/test-connection", agentController.TestSSHConnection)

		// Agent 管理页面专用接口
		cmdb.GET("/agents", agentController.GetAgentList)
		cmdb.POST("/agents/batch-deploy", agentController.BatchDeployAgent)
		cmdb.POST("/agents/batch-uninstall", agentController.BatchUninstallAgent)
		cmdb.DELETE("/agents/:id", agentController.DeleteAgentRecord)

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
		cmdb.POST("/sessions/:id/resize", bastionController.ResizeTerminalPTY)

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

		// 属性定义管理
		cmdb.GET("/attributes", attributeController.GetAttributeDefinitions)
		cmdb.POST("/attributes", attributeController.CreateAttributeDefinition)
		cmdb.PUT("/attributes/:id", attributeController.UpdateAttributeDefinition)
		cmdb.DELETE("/attributes/:id", attributeController.DeleteAttributeDefinition)
		cmdb.GET("/attributes/:id", attributeController.GetAttributeDefinitionByID)
		cmdb.POST("/attributes/validate", attributeController.ValidateServerAttribute)

		// 主机属性值管理
		cmdb.GET("/server-attributes/:serverId", attributeController.GetServerAttributes)
		cmdb.POST("/server-attributes/:serverId", attributeController.SaveServerAttributes)

		// 按属性筛选主机
		cmdb.GET("/servers-by-attributes", attributeController.GetServersByAttributes)
	}
}
