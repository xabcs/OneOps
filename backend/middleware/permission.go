package middleware

import (
	"fmt"
	"oneops/backend/services"
	"oneops/backend/utils"

	"github.com/gin-gonic/gin"
)

// RequirePermission 统一的权限检查中间件
// 参数：permissionCode - 层级权限代码，格式为 "模块.资源.操作"
// 例如："system.user.list", "cmdb.asset.create", "monitor.task.delete"
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, utils.ErrorUnauthorized("未授权访问"))
			c.Abort()
			return
		}

		// 获取权限服务
		permService, err := services.GetPermissionService()
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限服务初始化失败: "+err.Error()))
			c.Abort()
			return
		}

		// 检查权限（HasPermission 内部会检查 admin 用户和超级管理员角色）
		hasPermission, err := permService.HasPermission(userID.(uint), permissionCode)
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限检查失败: "+err.Error()))
			c.Abort()
			return
		}

		if !hasPermission {
			// 记录权限不足的详细信息
			username, _ := c.Get("username")
			fmt.Printf("[权限不足] user_id=%v, username=%v, required_permission=%s\n",
				userID, username, permissionCode)

			c.JSON(403, utils.ErrorForbidden("权限不足: 需要 "+permissionCode+" 权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// 资源操作权限快捷方式
// RequireResourceAccess 为指定资源和操作生成权限检查中间件
func RequireResourceAccess(module, resource, action string) gin.HandlerFunc {
	permissionCode := fmt.Sprintf("%s.%s.%s", module, resource, action)
	return RequirePermission(permissionCode)
}

// 常用权限快捷方式
func RequireView(module, resource string) gin.HandlerFunc {
	return RequireResourceAccess(module, resource, "view")
}

func RequireCreate(module, resource string) gin.HandlerFunc {
	return RequireResourceAccess(module, resource, "create")
}

func RequireUpdate(module, resource string) gin.HandlerFunc {
	return RequireResourceAccess(module, resource, "update")
}

func RequireDelete(module, resource string) gin.HandlerFunc {
	return RequireResourceAccess(module, resource, "delete")
}

func RequireList(module, resource string) gin.HandlerFunc {
	return RequireResourceAccess(module, resource, "list")
}