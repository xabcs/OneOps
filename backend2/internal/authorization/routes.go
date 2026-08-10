package authorization

import (
	"oneops/backend2/internal/system"
	"oneops/backend2/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes 设置授权中心相关路由
func SetupAuthRoutes(
	r *gin.Engine,
	applicationPermissionController *ApplicationPermissionController,
) {
	api := r.Group("/api")
	systemGroup := api.Group("/system")
	systemGroup.Use(middleware.Auth())
	systemGroup.Use(system.RequirePermissionFromDB()) // ✅ 统一权限检查
	{
		// 应用管理
		systemGroup.GET("/applications", applicationPermissionController.GetApplications)
		systemGroup.POST("/applications", applicationPermissionController.CreateApplication)
		systemGroup.PUT("/applications/:id", applicationPermissionController.UpdateApplication)
		systemGroup.DELETE("/applications/:id", applicationPermissionController.DeleteApplication)
		systemGroup.POST("/applications/:id/sync-roles", applicationPermissionController.SyncRoles)
		systemGroup.GET("/applications/:id/roles", applicationPermissionController.GetApplicationRoles)
		systemGroup.POST("/applications/:id/sync-users", applicationPermissionController.SyncUsers)
		systemGroup.GET("/applications/:id/users", applicationPermissionController.GetApplicationUsers)
		systemGroup.POST("/applications/:id/sync-groups", applicationPermissionController.SyncApplicationGroups)
		systemGroup.GET("/applications/:id/groups", applicationPermissionController.GetApplicationGroups)
		systemGroup.POST("/applications/:id/sync-rules", applicationPermissionController.SyncAuthorizationRules)
		systemGroup.GET("/applications/:id/rules", applicationPermissionController.GetApplicationAuthorizationRules)
		systemGroup.GET("/applications/:id/operation-logs", applicationPermissionController.GetOperationLogs)

		// 应用类型配置
		systemGroup.GET("/applications/types", applicationPermissionController.GetSupportedAppTypes)
		systemGroup.GET("/applications/types/:type/config", applicationPermissionController.GetAppTypeConfigTemplate)

		// 授权中心用户管理
		systemGroup.GET("/auth-users", applicationPermissionController.GetAllAuthUsers)
		systemGroup.GET("/auth-users/list", applicationPermissionController.GetAuthUsers)
		systemGroup.GET("/auth-users/:id/password", applicationPermissionController.GetAuthUserPassword)
		systemGroup.POST("/auth-users", applicationPermissionController.CreateAuthUser)
		systemGroup.PUT("/auth-users/:id", applicationPermissionController.UpdateAuthUser)
		systemGroup.DELETE("/auth-users/:id", applicationPermissionController.DeleteAuthUser)

		// 授权中心用户组管理
		systemGroup.GET("/auth-groups", applicationPermissionController.GetAllAuthGroups)
		systemGroup.GET("/auth-groups/list", applicationPermissionController.GetAuthGroups)
		systemGroup.POST("/auth-groups", applicationPermissionController.CreateAuthGroup)
		systemGroup.PUT("/auth-groups/:id", applicationPermissionController.UpdateAuthGroup)
		systemGroup.DELETE("/auth-groups/:id", applicationPermissionController.DeleteAuthGroup)

		// 用户组权限绑定管理
		systemGroup.GET("/groups/:id/bindings", applicationPermissionController.GetGroupBindings)
		systemGroup.POST("/groups/:id/bindings", applicationPermissionController.CreateGroupBinding)
		systemGroup.DELETE("/groups/bindings/:id", applicationPermissionController.DeleteGroupBinding)

		// 用户组成员管理
		systemGroup.GET("/users/:id/groups", applicationPermissionController.GetUserGroups)
		systemGroup.POST("/users/assign-group", applicationPermissionController.AssignUserToGroup)
		systemGroup.DELETE("/users/:id/groups/:groupId", applicationPermissionController.DeleteUserGroup)

		// 用户身份映射管理
		systemGroup.GET("/user-identity-mappings", applicationPermissionController.GetUserIdentityMappings)
		systemGroup.DELETE("/user-identity-mappings/:id", applicationPermissionController.DeleteUserIdentityMapping)

		// 用户有效权限查询
		systemGroup.GET("/user-permissions", applicationPermissionController.GetUserEffectivePermissions)
		systemGroup.GET("/user-permissions/matrix", applicationPermissionController.GetUserEffectivePermissionsMatrix)

		// 权限绑定执行记录
		systemGroup.GET("/group-bindings/:bindingId/executions", applicationPermissionController.GetGroupBindingExecutions)
	}
}
