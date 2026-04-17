package routes

import (
	"oneops/backend/controllers"
	"oneops/backend/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 应用中间件
	r.Use(gin.Recovery())
	r.Use(middlewares.Response())
	r.Use(cors.New(middlewares.CORS()))

	// 创建审计中间件
	auditMiddleware := middlewares.NewAuditMiddleware()
	// 应用操作日志审计中间件
	r.Use(auditMiddleware.OperationLog())

	// 创建控制器
	authController := controllers.NewAuthController()
	menuController := controllers.NewMenuController()
	roleController := controllers.NewRoleController()
	userController := controllers.NewUserController()
	auditController := controllers.NewAuditController()

	// API 路由组
	api := r.Group("/api")
	{
		// 认证路由（无需认证）
		api.POST("/login", authController.Login)

		// 用户信息路由（需要认证）
		api.GET("/user/info", middlewares.Auth(), authController.GetUserInfo)
		api.POST("/logout", middlewares.Auth(), authController.Logout)

		// 系统管理路由（需要认证）
		system := api.Group("/system")
		system.Use(middlewares.Auth())
		{
			// 菜单管理
			system.GET("/menus", menuController.GetMenus)
			system.POST("/menus", menuController.CreateMenu)
			system.PUT("/menus/:id", menuController.UpdateMenu)
			system.DELETE("/menus/:id", menuController.DeleteMenu)

			// 角色管理
			system.GET("/roles", roleController.GetRoles)
			system.POST("/roles", roleController.CreateRole)
			system.PUT("/roles/:id", roleController.UpdateRole)
			system.DELETE("/roles/:id", roleController.DeleteRole)

			// 用户管理
			system.GET("/users", userController.GetUsers)
			system.POST("/users", userController.CreateUser)
			system.PUT("/users/:id", userController.UpdateUser)
			system.DELETE("/users/:id", userController.DeleteUser)
		}

		// 审计管理路由（需要认证）
		audit := api.Group("/audit")
		audit.Use(middlewares.Auth())
		{
			// 登录日志
			audit.GET("/login-logs", auditController.GetLoginLogs)
			audit.GET("/login-logs/export", auditController.ExportLoginLogs)

			// 操作日志
			audit.GET("/operation-logs", auditController.GetOperationLogs)
			audit.GET("/operation-logs/export", auditController.ExportOperationLogs)

			// 系统事件日志
			audit.GET("/system-event-logs", auditController.GetSystemEventLogs)

			// 审计统计
			audit.GET("/stats", auditController.GetAuditStats)
		}
	}
}
