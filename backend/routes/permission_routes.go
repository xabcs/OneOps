package routes

import (
	"oneops/backend/controller"
	"oneops/backend/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPermissionRoutes 注册权限相关路由
func RegisterPermissionRoutes(router *gin.RouterGroup) {
	// 创建权限控制器实例
	permController, err := controller.NewPermissionController()
	if err != nil {
		panic("Failed to create permission controller: " + err.Error())
	}

	// 权限管理路由（需要认证）
	permGroup := router.Group("/permissions")
	permGroup.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		permGroup.GET("", permController.GetPermissionList) // 获取权限列表（分页）
		permGroup.GET("/tree", permController.GetPermissionTree) // 获取权限树

		permGroup.POST("", permController.CreatePermission) // 创建权限

		permGroup.PUT("/:id", permController.UpdatePermission) // 更新权限

		permGroup.DELETE("/:id", permController.DeletePermission) // 删除权限

		permGroup.POST("/check", permController.CheckPermission) // 检查权限（内部使用）
	}

	// 角色权限路由（需要认证）
	rolePermGroup := router.Group("/roles/:roleId/permissions")
	rolePermGroup.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		rolePermGroup.GET("", permController.GetRolePermissions) // 获取角色权限

		rolePermGroup.POST("", permController.AssignRolePermissions) // 分配权限
	}

	// 用户权限路由（需要认证）
	userPermGroup := router.Group("/user")
	userPermGroup.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		userPermGroup.GET("/permissions", permController.GetUserPermissions) // 获取当前用户权限
	}
}
