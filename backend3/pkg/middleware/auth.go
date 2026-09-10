package middleware

import (
	"net/http"
	"strings"

	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/utils"
	"oneops/backend3/service/system"

	"github.com/gin-gonic/gin"
)

// Auth JWT 认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.ErrorUnauthorized("缺少认证令牌"))
			c.Abort()
			return
		}

		// Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, utils.ErrorUnauthorized("认证令牌格式错误"))
			c.Abort()
			return
		}

		// 解析 token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.ErrorUnauthorized("认证令牌无效或已过期"))
			c.Abort()
			return
		}

		// 用户状态校验（H4）：禁用用户即使 token 未过期也立即失效，避免引入 token 黑名单。
		// 一条 join 同时取状态与启用角色，角色存入上下文供权限中间件复用（每请求再省一条查询）；
		// 查询失败或用户不存在按禁用处理（fail-closed）
		status, roles, err := system.FetchUserStatusAndRoles(database.GetDB(), claims.UserID)
		if err != nil || status != "active" {
			c.JSON(http.StatusUnauthorized, utils.ErrorUnauthorized("用户已被禁用或不存在"))
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("claims", claims) // 设置完整的claims对象
		c.Set("user_roles", roles)
		c.Next()
	}
}

// OptionalAuth 可选的 JWT 认证中间件（如果有 token 则验证，没有则跳过）
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("claims", claims) // 设置完整的claims对象
		c.Next()
	}
}
