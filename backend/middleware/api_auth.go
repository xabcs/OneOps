package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"oneops/backend/services"
	"oneops/backend/utils"
)

// APIAuthMiddleware API权限校验中间件
func APIAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, utils.ErrorUnauthorized("未登录"))
			c.Abort()
			return
		}

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 跳过登录接口等公开接口
		if isPublicAPI(path) {
			c.Next()
			return
		}

		// 检查API权限
		permissionService, err := services.GetPermissionService()
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorInternal("权限服务异常"))
			c.Abort()
			return
		}

		allowed, err := permissionService.HasAPIPermission(userID.(uint), path, method)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorInternal("权限检查失败"))
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, utils.ErrorForbidden("无权限访问该接口"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPublicAPI 判断是否为公开接口（不需要权限验证）
func isPublicAPI(path string) bool {
	publicAPIs := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/api/public/",
		"/api/docs/",
		"/api/health",
	}

	for _, publicAPI := range publicAPIs {
		if strings.HasPrefix(path, strings.TrimSuffix(publicAPI, "/")) {
			return true
		}
	}

	return false
}

// OptionalAPIAuthMiddleware 可选的API权限校验（用于需要检查权限但不强制要求的情况）
func OptionalAPIAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		permissionService, err := services.GetPermissionService()
		if err != nil {
			c.Next()
			return
		}

		// 将权限检查结果存储在上下文中
		allowed, _ := permissionService.HasAPIPermission(userID.(uint), path, method)
		c.Set("hasAPIPermission", allowed)

		c.Next()
	}
}