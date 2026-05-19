package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupCMDBRoutes 设置CMDB路由
func SetupCMDBRoutes(r *gin.Engine) {
	cmdbController := controllers.NewCMDBController()

	// API 路由组
	api := r.Group("/api")
	{
		// CMDB资产管理路由（需要认证）
		cmdb := api.Group("/cmdb")
		cmdb.Use(middlewares.Auth())
		{
			// 服务器管理 - 先注册复杂路由避免冲突
			cmdb.DELETE("/server-tags/:serverId/:tagId", func(ctx *gin.Context) {
				// 重写参数以匹配现有控制器
				ctx.Params = append(ctx.Params, gin.Param{Key: "serverId", Value: ctx.Param("serverId")})
				ctx.Params = append(ctx.Params, gin.Param{Key: "tagId", Value: ctx.Param("tagId")})
				cmdbController.RemoveServerTag(ctx)
			})

			cmdb.GET("/servers", cmdbController.GetServers)
			cmdb.GET("/servers/:id", cmdbController.GetServerByID)
			cmdb.POST("/servers", cmdbController.CreateServer)
			cmdb.PUT("/servers/:id", cmdbController.UpdateServer)
			cmdb.DELETE("/servers/:id", cmdbController.DeleteServer)
			cmdb.GET("/servers/stats", cmdbController.GetServerStats)

			// 业务系统管理
			cmdb.GET("/business-units", cmdbController.GetBusinessUnits)
			cmdb.POST("/business-units", cmdbController.CreateBusinessUnit)
			cmdb.PUT("/business-units/:id", cmdbController.UpdateBusinessUnit)
			cmdb.DELETE("/business-units/:id", cmdbController.DeleteBusinessUnit)

			// 机房机柜管理
			cmdb.GET("/rooms", cmdbController.GetServerRooms)
			cmdb.GET("/cabinets", cmdbController.GetCabinets)

			// 标签管理
			cmdb.GET("/tags", cmdbController.GetServerTags)
			cmdb.POST("/tags", cmdbController.CreateServerTag)
			cmdb.DELETE("/tags/:id", cmdbController.DeleteServerTag)
			cmdb.POST("/tags/assign", cmdbController.AssignServerTag)

			// 资产变更记录
			cmdb.GET("/asset-changes", cmdbController.GetAssetChanges)
		}
	}
}
