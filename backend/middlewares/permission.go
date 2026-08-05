package middlewares

import (
	"fmt"
	"strings"
	"oneops/backend/services"
	"oneops/backend/utils"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(permissionCode string) gin.HandlerFunc {
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

		// 检查权限（HasPermission内部会检查admin用户）
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

// APILevelPermissionMiddleware Level 4 API级别权限检查（使用API路径和HTTP方法）
// 直接检查 API 端点权限，而不是操作级权限码
// 示例：p, ops, /api/system/users, GET
func APILevelPermissionMiddleware(apiPath string, httpMethod string) gin.HandlerFunc {
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

		// 构造完整的API路径（如果有路径参数，保留参数模板）
		fullPath := apiPath
		if fullPath == "" {
			// 如果没有指定路径，使用当前请求的路径
			fullPath = c.FullPath()
		}

		// 标准化HTTP方法
		method := httpMethod
		if method == "" {
			method = c.Request.Method
		}

		// 检查API级权限
		hasPermission, err := permService.HasAPIPermission(userID.(uint), fullPath, method)
		if err != nil {
			c.JSON(500, utils.ErrorInternal("权限检查失败: "+err.Error()))
			c.Abort()
			return
		}

		if !hasPermission {
			// 记录权限不足的详细信息
			username, _ := c.Get("username")
			fmt.Printf("[API权限不足] user_id=%v, username=%v, api=%s, method=%s\n",
				userID, username, fullPath, method)

			c.JSON(403, utils.ErrorForbidden(fmt.Sprintf("权限不足: 需要 %s %s 权限", method, fullPath)))
			c.Abort()
			return
		}

		c.Next()
	}
}

// API权限映射：METHOD:PATH -> 权限编码
var apiPermissionMap = map[string]string{
	// 用户管理API
	"GET:/api/system/users":                    "system.user.view",
	"GET:/api/system/users/:id":                "system.user.view",
	"GET:/api/system/users/:id/roles":          "system.user.view_roles",
	"POST:/api/system/users":                   "system.user.create",
	"PUT:/api/system/users/:id":                "system.user.update",
	"DELETE:/api/system/users/:id":             "system.user.delete",
	"PUT:/api/system/users/:id/password":        "system.user.reset_password",

	// 角色管理API
	"GET:/api/system/roles":                     "system.role.view",
	"GET:/api/system/roles/:id":                 "system.role.view",
	"POST:/api/system/roles":                    "system.role.create",
	"PUT:/api/system/roles/:id":                 "system.role.update",
	"DELETE:/api/system/roles/:id":              "system.role.delete",

	// 权限管理API
	"GET:/api/system/permissions":              "system.permission.view",
	"GET:/api/system/permissions/:id":          "system.permission.view",
	"POST:/api/system/permissions":              "system.permission.create",
	"PUT:/api/system/permissions/:id":            "system.permission.update",
	"DELETE:/api/system/permissions/:id":        "system.permission.delete",
}

// getPermissionFromAPI 从API路径获取权限编码
func getPermissionFromAPI(method, path string) string {
	// 标准化路径
	normalizedPath := path
	if idx := strings.Index(path, "?"); idx != -1 {
		normalizedPath = path[:idx]
	}

	key := fmt.Sprintf("%s:%s", method, normalizedPath)

	// 精确匹配
	if permission, ok := apiPermissionMap[key]; ok {
		return permission
	}

	// 模式匹配（处理路径参数）
	for apiPath, permission := range apiPermissionMap {
		if matchPathPattern(normalizedPath, apiPath) {
			return permission
		}
	}

	return ""
}

// matchPathPattern 匹配路径模式
func matchPathPattern(actualPath, patternPath string) bool {
	actualParts := strings.Split(actualPath, "/")
	patternParts := strings.Split(patternPath, "/")

	if len(actualParts) != len(patternParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		patternPart := patternParts[i]
		actualPart := actualParts[i]

		// 路径参数匹配
		if strings.HasPrefix(patternPart, ":") {
			continue
		}

		// 精确匹配
		if patternPart != actualPart {
			return false
		}
	}

	return true
}

// RequirePermission 权限辅助函数
func RequirePermission(permission string) gin.HandlerFunc {
	return PermissionMiddleware(permission)
}

// 资源操作权限快捷方式
func RequireResourceAccess(resource, action string) gin.HandlerFunc {
	permissionCode := resource + "." + action
	return PermissionMiddleware(permissionCode)
}

// 常用权限快捷方式
func RequireView(resource string) gin.HandlerFunc {
	return RequireResourceAccess(resource, "view")
}

func RequireCreate(resource string) gin.HandlerFunc {
	return RequireResourceAccess(resource, "create")
}

func RequireUpdate(resource string) gin.HandlerFunc {
	return RequireResourceAccess(resource, "update")
}

func RequireDelete(resource string) gin.HandlerFunc {
	return RequireResourceAccess(resource, "delete")
}