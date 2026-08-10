package system

import (
	"github.com/gin-gonic/gin"
)

// RegisterPermissionRoutes 注册权限相关路由
func RegisterPermissionRoutes(router *gin.RouterGroup) {
	// 创建权限控制器实例
	permController, err := NewPermissionController()
	if err != nil {
		panic("Failed to create permission controller: " + err.Error())
	}

	// 权限管理路由（需要认证）
	permGroup := router.Group("/permissions")
	permGroup.Use(RequirePermissionFromDB())
	{
		permGroup.GET("", permController.GetPermissionList)
		permGroup.GET("/tree", permController.GetPermissionTree)
		permGroup.POST("", permController.CreatePermission)
		permGroup.PUT("/:id", permController.UpdatePermission)
		permGroup.DELETE("/:id", permController.DeletePermission)
		permGroup.POST("/check", permController.CheckPermission)
	}

	// 角色权限路由
	rolePermGroup := router.Group("/roles/:roleId/permissions")
	rolePermGroup.Use(RequirePermissionFromDB())
	{
		rolePermGroup.GET("", permController.GetRolePermissions)
		rolePermGroup.POST("", permController.AssignRolePermissions)
	}

	// 用户权限路由
	userPermGroup := router.Group("/user")
	userPermGroup.Use(RequirePermissionFromDB())
	{
		userPermGroup.GET("/permissions", permController.GetUserPermissions)
	}
}
