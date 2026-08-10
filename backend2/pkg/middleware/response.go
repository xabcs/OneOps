package middleware

import (
	"oneops/backend2/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Response 统一响应中间件
func Response() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在响应之前处理
		c.Next()

		// 如果有错误，返回错误响应
		if len(c.Errors) > 0 {
			c.JSON(500, utils.ErrorInternal(c.Errors.String()))
		}
	}
}
