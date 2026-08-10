package routes

import (
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/middleware"

	cmdbsvc "oneops/backend3/service/cmdb"
	syssvc "oneops/backend3/service/system"

	cmdbctrl "oneops/backend3/controller/cmdb"
	sysctrl "oneops/backend3/controller/system"
	repocmdb "oneops/backend3/repository/cmdb"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes 设置系统管理相关路由（含权限管理路由）
func SetupSystemRoutes(
	r *gin.Engine,
	menuController *sysctrl.MenuController,
	roleController *sysctrl.RoleController,
	userController *sysctrl.UserController,
) {
	// 创建属性控制器（属性定义属于 CMDB 模块，路由挂载在 system 下）
	attributeRepo := repocmdb.NewAttributeRepository(database.GetDB())
	attributeSvc := cmdbsvc.NewAttributeService(attributeRepo)
	attributeController := cmdbctrl.NewAttributeController(attributeSvc)

	// 创建权限控制器
	permSvc, err := syssvc.NewPermissionService()
	if err != nil {
		panic("failed to create permission service: " + err.Error())
	}
	permController := sysctrl.NewPermissionController(permSvc)

	api := r.Group("/api")
	system := api.Group("/system")
	system.Use(middleware.Auth())
	system.Use(middleware.RequirePermissionFromDB())
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

		// 权限管理路由
		registerPermissionRoutes(system, permController)
	}
}

// registerPermissionRoutes 注册权限相关路由
func registerPermissionRoutes(router *gin.RouterGroup, permController *sysctrl.PermissionController) {
	// 权限管理路由
	permGroup := router.Group("/permissions")
	permGroup.Use(middleware.RequirePermissionFromDB())
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
	rolePermGroup.Use(middleware.RequirePermissionFromDB())
	{
		rolePermGroup.GET("", permController.GetRolePermissions)
		rolePermGroup.POST("", permController.AssignRolePermissions)
	}

	// 用户权限路由
	userPermGroup := router.Group("/user")
	userPermGroup.Use(middleware.RequirePermissionFromDB())
	{
		userPermGroup.GET("/permissions", permController.GetUserPermissions)
	}
}
