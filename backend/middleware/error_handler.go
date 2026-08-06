package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"oneops/backend/errors"
	"oneops/backend/utils"
	"oneops/backend/logger"

	"go.uber.org/zap"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// 处理不同类型的错误
			var response utils.Response

			if appErr, ok := err.(*errors.AppError); ok {
				// 应用错误
				response = utils.ErrorResponseFromAppError(appErr)
				logger.Debug("应用错误",
					zap.Int("code", appErr.Code),
					zap.String("message", appErr.Message),
					zap.String("path", c.Request.URL.Path),
				)
			} else {
				// 未知错误
				response = utils.ErrorInternal("服务器内部错误")
				logger.Error("未知错误",
					zap.Error(err),
					zap.String("path", c.Request.URL.Path),
				)
			}

			// 如果还没有响应，则返回错误响应
			if !c.Writer.Written() {
				c.JSON(http.StatusOK, response)
			}
		}
	}
}

// HandleControllerError 在 Controller 中处理错误
func HandleControllerError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	c.Error(err)

	// 如果是 AppError，直接返回响应
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(http.StatusOK, utils.ErrorResponseFromAppError(appErr))
		return
	}

	// 其他错误，交给错误中间件处理
}
