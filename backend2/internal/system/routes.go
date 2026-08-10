package system

import (
	"oneops/backend2/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes 设置系统管理相关路由
func SetupSystemRoutes(
	r *gin.Engine,
	menuController *MenuController,
	roleController *RoleController,
	userController *UserController,
	attributeController *AttributeController,
) {
	api := r.Group("/api")
	system := api.Group("/system")
	system.Use(middleware.Auth())
	system.Use(RequirePermissionFromDB())
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

		// 注册权限管理路由
		RegisterPermissionRoutes(system)
	}
}
