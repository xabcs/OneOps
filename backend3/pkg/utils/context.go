package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext 安全地从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, ErrorUnauthorized("用户未登录"))
		return 0, false
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusOK, ErrorBadRequest("无效的用户ID格式"))
		return 0, false
	}

	return uid, true
}

// MustGetUserIDFromContext 从上下文中获取用户ID，失败时panic
// 仅在确保用户已认证的中间件之后使用
func MustGetUserIDFromContext(c *gin.Context) uint {
	userID, exists := c.Get("user_id")
	if !exists {
		panic("user_id not found in context")
	}

	uid, ok := userID.(uint)
	if !ok {
		panic("user_id is not uint type")
	}

	return uid
}
