package routes

import (
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/middleware"

	syssvc "oneops/backend3/service/system"

	sysctrl "oneops/backend3/controller/system"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes 设置系统管理相关路由（含权限管理路由）
func SetupSystemRoutes(
	r *gin.Engine,
	menuController *sysctrl.MenuController,
	roleController *sysctrl.RoleController,
	userController *sysctrl.UserController,
) {
	// 创建 repositories
	// 必须使用单例：中间件鉴权用的是 GetPermissionService()，
	// 若此处新建实例，运行期分配/撤销权限只会写入新实例内存，鉴权侧不生效（H1）
	permSvc, err := syssvc.GetPermissionService()
	if err != nil {
		panic("failed to get permission service: " + err.Error())
	}
	permController := sysctrl.NewPermissionController(permSvc)
	userGroupController := sysctrl.NewUserGroupController(syssvc.NewUserGroupService(database.GetDB()))

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
		system.GET("/roles/options", roleController.GetRoleOptions)
		system.GET("/roles/menu-paths", roleController.GetRoleMenuPaths)
		system.GET("/roles/:id/users", roleController.GetRoleUsers)
		system.POST("/roles", roleController.CreateRole)
		system.PUT("/roles/:id", roleController.UpdateRole)
		system.DELETE("/roles/:id", roleController.DeleteRole)

		// 用户管理
		system.GET("/users", userController.GetUsers)
		system.GET("/users/options", userController.GetUserOptions)
		system.POST("/users", userController.CreateUser)
		system.PUT("/users/:id", userController.UpdateUser)
		system.DELETE("/users/:id", userController.DeleteUser)
		system.PUT("/users/:id/password", userController.ResetPassword)

		// 用户组管理（组成员为平台用户，组用于集群等资源的批量授权）
		system.GET("/user-groups", userGroupController.GetUserGroups)
		system.GET("/user-groups/options", userGroupController.GetUserGroupOptions)
		system.POST("/user-groups", userGroupController.CreateUserGroup)
		system.PUT("/user-groups/:id", userGroupController.UpdateUserGroup)
		system.DELETE("/user-groups/:id", userGroupController.DeleteUserGroup)
		system.GET("/user-groups/:id/members", userGroupController.GetGroupMembers)
		system.POST("/user-groups/:id/members", userGroupController.AddGroupMembers)
		system.DELETE("/user-groups/:id/members/:userId", userGroupController.RemoveGroupMember)

		// 权限管理路由
		registerPermissionRoutes(system, permController)
	}
}

// registerPermissionRoutes 注册权限相关路由
// 注意：三个子组均挂在 system 组下，父组已统一应用 Auth + RequirePermissionFromDB，
// 子组不再重复挂权限中间件（重复挂载会导致同一请求执行两次权限校验）
func registerPermissionRoutes(router *gin.RouterGroup, permController *sysctrl.PermissionController) {
	// 权限管理路由
	permGroup := router.Group("/permissions")
	{
		permGroup.GET("", permController.GetPermissionList)
		permGroup.GET("/tree", permController.GetPermissionTree)
		permGroup.GET("/options", permController.GetPermissionOptions)
		permGroup.POST("", permController.CreatePermission)
		permGroup.PUT("/:id", permController.UpdatePermission)
		permGroup.DELETE("/:id", permController.DeletePermission)
		permGroup.POST("/check", permController.CheckPermission)

		// 权限路由映射管理（运行时权限校验的数据来源，页面维护入口）
		permGroup.GET("/routes", permController.GetPermissionRoutes)
		permGroup.POST("/routes", permController.CreatePermissionRoute)
		permGroup.PUT("/routes/:id", permController.UpdatePermissionRoute)
		permGroup.DELETE("/routes/:id", permController.DeletePermissionRoute)
	}

	// 角色权限路由（段名统一 :id，与 /roles/:id 保持一致，避免 gin 通配符冲突）
	rolePermGroup := router.Group("/roles/:id/permissions")
	{
		rolePermGroup.GET("", permController.GetRolePermissions)
		rolePermGroup.POST("", permController.AssignRolePermissions)
	}

	// 用户权限路由
	userPermGroup := router.Group("/user")
	{
		userPermGroup.GET("/permissions", permController.GetUserPermissions)
	}
}
