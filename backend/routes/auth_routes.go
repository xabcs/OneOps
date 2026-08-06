package routes

import (
	"oneops/backend/controller"
	"oneops/backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes 设置授权中心相关路由
func SetupAuthRoutes(
	r *gin.Engine,
	applicationPermissionController *controller.ApplicationPermissionController,
) {
	api := r.Group("/api")
	system := api.Group("/system")
	system.Use(middleware.Auth())
	system.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		// 应用管理
		system.GET("/applications", applicationPermissionController.GetApplications)
		system.POST("/applications", applicationPermissionController.CreateApplication)
		system.PUT("/applications/:id", applicationPermissionController.UpdateApplication)
		system.DELETE("/applications/:id", applicationPermissionController.DeleteApplication)
		system.POST("/applications/:id/sync-roles", applicationPermissionController.SyncRoles)
		system.GET("/applications/:id/roles", applicationPermissionController.GetApplicationRoles)
		system.POST("/applications/:id/sync-users", applicationPermissionController.SyncUsers)
		system.GET("/applications/:id/users", applicationPermissionController.GetApplicationUsers)
		system.POST("/applications/:id/sync-groups", applicationPermissionController.SyncApplicationGroups)
		system.GET("/applications/:id/groups", applicationPermissionController.GetApplicationGroups)
		system.POST("/applications/:id/sync-rules", applicationPermissionController.SyncAuthorizationRules)
		system.GET("/applications/:id/rules", applicationPermissionController.GetApplicationAuthorizationRules)
		system.GET("/applications/:id/operation-logs", applicationPermissionController.GetOperationLogs)

		// 应用类型配置
		system.GET("/applications/types", applicationPermissionController.GetSupportedAppTypes)
		system.GET("/applications/types/:type/config", applicationPermissionController.GetAppTypeConfigTemplate)

		// 授权中心用户管理
		system.GET("/auth-users", applicationPermissionController.GetAllAuthUsers)
		system.GET("/auth-users/list", applicationPermissionController.GetAuthUsers)
		system.GET("/auth-users/:id/password", applicationPermissionController.GetAuthUserPassword)
		system.POST("/auth-users", applicationPermissionController.CreateAuthUser)
		system.PUT("/auth-users/:id", applicationPermissionController.UpdateAuthUser)
		system.DELETE("/auth-users/:id", applicationPermissionController.DeleteAuthUser)

		// 授权中心用户组管理
		system.GET("/auth-groups", applicationPermissionController.GetAllAuthGroups)
		system.GET("/auth-groups/list", applicationPermissionController.GetAuthGroups)
		system.POST("/auth-groups", applicationPermissionController.CreateAuthGroup)
		system.PUT("/auth-groups/:id", applicationPermissionController.UpdateAuthGroup)
		system.DELETE("/auth-groups/:id", applicationPermissionController.DeleteAuthGroup)

		// 用户组权限绑定管理
		system.GET("/groups/:id/bindings", applicationPermissionController.GetGroupBindings)
		system.POST("/groups/:id/bindings", applicationPermissionController.CreateGroupBinding)
		system.DELETE("/groups/bindings/:id", applicationPermissionController.DeleteGroupBinding)

		// 用户组成员管理
		system.GET("/users/:id/groups", applicationPermissionController.GetUserGroups)
		system.POST("/users/assign-group", applicationPermissionController.AssignUserToGroup)
		system.DELETE("/users/:id/groups/:groupId", applicationPermissionController.DeleteUserGroup)

		// 用户身份映射管理
		system.GET("/user-identity-mappings", applicationPermissionController.GetUserIdentityMappings)
		system.DELETE("/user-identity-mappings/:id", applicationPermissionController.DeleteUserIdentityMapping)

		// 用户有效权限查询
		system.GET("/user-permissions", applicationPermissionController.GetUserEffectivePermissions)
		system.GET("/user-permissions/matrix", applicationPermissionController.GetUserEffectivePermissionsMatrix)

		// 权限绑定执行记录
		system.GET("/group-bindings/:bindingId/executions", applicationPermissionController.GetGroupBindingExecutions)
	}
}
