package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/container"
	"oneops/backend/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes 设置系统管理相关路由
func SetupSystemRoutes(
	r *gin.Engine,
	cnt *container.ServiceContainer,
	menuController *controllers.MenuController,
	roleController *controllers.RoleController,
	userController *controllers.UserController,
	attributeController *controllers.AttributeController,
	applicationPermissionController *controllers.ApplicationPermissionController,
) {
	// 获取服务容器
	api := r.Group("/api")
	system := api.Group("/system")
	system.Use(middlewares.Auth())
	{
		// 菜单管理
		system.GET("/menus",
			middlewares.APILevelPermissionMiddleware("/api/system/menus", "GET"),
			menuController.GetMenus)
		system.GET("/menus/tree",
			middlewares.APILevelPermissionMiddleware("/api/system/menus/tree", "GET"),
			menuController.GetMenuTree)
		system.POST("/menus",
			middlewares.APILevelPermissionMiddleware("/api/system/menus", "POST"),
			menuController.CreateMenu)
		system.PUT("/menus/:id",
			middlewares.APILevelPermissionMiddleware("/api/system/menus/:id", "PUT"),
			menuController.UpdateMenu)
		system.DELETE("/menus/:id",
			middlewares.APILevelPermissionMiddleware("/api/system/menus/:id", "DELETE"),
			menuController.DeleteMenu)

		// 角色管理
		system.GET("/roles",
			middlewares.APILevelPermissionMiddleware("/api/system/roles", "GET"),
			roleController.GetRoles)
		system.POST("/roles",
			middlewares.APILevelPermissionMiddleware("/api/system/roles", "POST"),
			roleController.CreateRole)
		system.PUT("/roles/:id",
			middlewares.APILevelPermissionMiddleware("/api/system/roles/:id", "PUT"),
			roleController.UpdateRole)
		system.DELETE("/roles/:id",
			middlewares.APILevelPermissionMiddleware("/api/system/roles/:id", "DELETE"),
			roleController.DeleteRole)

		// 用户管理
		system.GET("/users",
			middlewares.APILevelPermissionMiddleware("/api/system/users", "GET"),
			userController.GetUsers)
		system.POST("/users",
			middlewares.APILevelPermissionMiddleware("/api/system/users", "POST"),
			userController.CreateUser)
		system.PUT("/users/:id",
			middlewares.APILevelPermissionMiddleware("/api/system/users/:id", "PUT"),
			userController.UpdateUser)
		system.DELETE("/users/:id",
			middlewares.APILevelPermissionMiddleware("/api/system/users/:id", "DELETE"),
			userController.DeleteUser)
		system.PUT("/users/:id/password",
			middlewares.APILevelPermissionMiddleware("/api/system/users/:id/password", "PUT"),
			userController.ResetPassword)

		// 属性管理
		system.GET("/attributes", attributeController.GetAttributeDefinitions)
		system.POST("/attributes", attributeController.CreateAttributeDefinition)
		system.PUT("/attributes/:id", attributeController.UpdateAttributeDefinition)
		system.DELETE("/attributes/:id", attributeController.DeleteAttributeDefinition)
		system.GET("/attributes/:id", attributeController.GetAttributeDefinitionByID)

		// 主机属性管理
		system.GET("/server-attributes/:serverId", attributeController.GetServerAttributes)
		system.POST("/server-attributes/:serverId", attributeController.SaveServerAttributes)

		// 应用权限管理
		// setupApplicationPermissionRoutes(system, applicationPermissionController) // 函数未定义，注释掉
	}
}
