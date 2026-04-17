package middlewares

import (
	"github.com/gin-contrib/cors"
)

// CORS 跨域中间件
func CORS() cors.Config {
	config := cors.DefaultConfig()

	// 允许所有来源，方便开发和测试
	config.AllowAllOrigins = true

	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"Content-Length", "Content-Type"}
	return config
}
