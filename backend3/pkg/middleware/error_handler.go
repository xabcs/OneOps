package middleware

import (
	"net/http"

	apperrors "oneops/backend3/pkg/errors"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			response := utils.HandleError(err)

			if _, ok := err.(*apperrors.AppError); !ok {
				logger.Error("未知错误", zap.Error(err), zap.String("path", c.Request.URL.Path))
			}

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
	c.JSON(http.StatusOK, utils.HandleError(err))
}
