package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/middlewares"

	"github.com/gin-gonic/gin"
)

// RegisterPermissionRoutes 注册权限相关路由
func RegisterPermissionRoutes(router *gin.RouterGroup) {
	// 创建权限控制器实例
	permController, err := controllers.NewPermissionController()
	if err != nil {
		panic("Failed to create permission controller: " + err.Error())
	}

	// 权限管理路由（需要认证）
	permGroup := router.Group("/permissions")
	permGroup.Use(middlewares.Auth())
	{
		permGroup.GET("",
			middlewares.RequirePermission("system.permission.list"),
			permController.GetPermissionList) // 获取权限列表（分页）
		permGroup.GET("/tree",
			middlewares.RequirePermission("system.permission.list"),
			permController.GetPermissionTree) // 获取权限树

		permGroup.POST("",
			middlewares.RequirePermission("system.permission.create"),
			permController.CreatePermission) // 创建权限

		permGroup.PUT("/:id",
			middlewares.RequirePermission("system.permission.update"),
			permController.UpdatePermission) // 更新权限

		permGroup.DELETE("/:id",
			middlewares.RequirePermission("system.permission.delete"),
			permController.DeletePermission) // 删除权限

		permGroup.POST("/check", permController.CheckPermission) // 检查权限（内部使用）
	}

	// 角色权限路由（需要认证）
	rolePermGroup := router.Group("/roles/:roleId/permissions")
	rolePermGroup.Use(middlewares.Auth())
	{
		rolePermGroup.GET("",
			middlewares.RequirePermission("system.role.view"),
			permController.GetRolePermissions) // 获取角色权限

		rolePermGroup.POST("",
			middlewares.RequirePermission("system.role.assign_permissions"),
			permController.AssignRolePermissions) // 分配权限
	}

	// 用户权限路由（需要认证）
	userPermGroup := router.Group("/user")
	userPermGroup.Use(middlewares.Auth())
	{
		userPermGroup.GET("/permissions",
			middlewares.RequirePermission("auth.permission.query"),
			permController.GetUserPermissions) // 获取当前用户权限
	}
}
