package middleware

import (
	"fmt"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"

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

		// 获取当前路由信息
		method := c.Request.Method   // GET/POST/PUT/DELETE
		path := c.FullPath()         // /api/audit/login-logs

		// 从数据库查询该路由需要的权限
		db := services.GetDB()
		var perm models.Permission

		err := db.Where("route_method = ? AND route_path = ? AND status = 1",
			method, path).First(&perm).Error

		if err != nil {
			// 该路由没有配置权限，默认放行
			// 也可以选择拒绝访问，根据业务需求调整
			c.Next()
			return
		}

		// 获取权限服务
		permService, err := services.GetPermissionService()
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限服务初始化失败"))
			c.Abort()
			return
		}

		// 检查用户权限（通过 Casbin）
		hasPermission, err := permService.HasPermission(userID.(uint), perm.Code)
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限检查失败"))
			c.Abort()
			return
		}

		if !hasPermission {
			// 记录权限不足的详细信息
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
