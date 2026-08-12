package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupSwagger 注册 Swagger UI 路由。
// 仅应在非生产环境调用（由 main.go 依据配置决定是否注册），
// 避免在生产环境暴露接口文档。
func SetupSwagger(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
