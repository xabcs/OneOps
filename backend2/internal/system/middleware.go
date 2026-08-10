package system

import (
	"fmt"
	"oneops/backend2/pkg/database"
	"oneops/backend2/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RequirePermissionFromDB 权限检查中间件（数据库驱动，无缓存）
func RequirePermissionFromDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, utils.ErrorUnauthorized("未授权访问"))
			c.Abort()
			return
		}

		method := c.Request.Method
		path := c.FullPath()

		db := database.GetDB()
		var perm Permission

		err := db.Where("route_method = ? AND route_path = ? AND status = 1",
			method, path).First(&perm).Error

		if err != nil {
			c.Next()
			return
		}

		permService, err := GetPermissionService()
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限服务初始化失败"))
			c.Abort()
			return
		}

		hasPermission, err := permService.HasPermission(userID.(uint), perm.Code)
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限检查失败"))
			c.Abort()
			return
		}

		if !hasPermission {
			username, _ := c.Get("username")
			fmt.Printf("[权限不足] user_id=%v, username=%v, route=%s %s, required_permission=%s\n",
				userID, username, method, path, perm.Code)

			c.JSON(403, utils.ErrorForbidden("权限不足: 需要 "+perm.Code+" 权限"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 统一的权限检查中间件
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, utils.ErrorUnauthorized("未授权访问"))
			c.Abort()
			return
		}

		permService, err := GetPermissionService()
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限服务初始化失败: "+err.Error()))
			c.Abort()
			return
		}

		hasPermission, err := permService.HasPermission(userID.(uint), permissionCode)
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限检查失败: "+err.Error()))
			c.Abort()
			return
		}

		if !hasPermission {
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
