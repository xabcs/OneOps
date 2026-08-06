package routes

import (
	"oneops/backend/controller"
	"oneops/backend/container"
	"oneops/backend/handler"
	"oneops/backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 应用中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Response())
	r.Use(middleware.ErrorHandler()) // 统一错误处理
	r.Use(cors.New(middleware.CORS()))

	// 创建审计中间件
	auditMiddleware := middleware.NewAuditMiddleware()
	// 应用操作日志审计中间件
	r.Use(auditMiddleware.OperationLog())

	// 获取服务容器
	cnt := container.GetContainer()

	// 创建控制器（使用服务容器）
	authController := controller.NewAuthController(cnt)
	menuController := controller.NewMenuController()
	roleController := controller.NewRoleController()
	userController := controller.NewUserController()
	auditController := controller.NewAuditController()
	monitoringController := controller.NewMonitoringController()
	wsMonitoringController := controller.NewMonitoringWebSocketController()
	routeController := controller.NewRouteController()
	cmdbController := controller.NewCMDBController(cnt)
	bastionController := controller.NewBastionController()
	attributeController := controller.NewAttributeController()
	sshHandler := handler.NewSSHWebSocketHandler()
	k8sClusterController := controller.NewK8sClusterController(cnt)
	k8sPermissionController := controller.NewK8sPermissionController(cnt)
	k8sResourceController := controller.NewK8sResourceController(cnt)
	k8sTerminalHandler := handler.NewK8sTerminalHandler(cnt)
	diagnosticController := controller.NewDiagnosticController(cnt)
	applicationPermissionController := controller.NewApplicationPermissionController()

	// API 路由组
	api := r.Group("/api")
	{
		// ========== 认证路由（无需认证） ==========
		api.POST("/login", authController.Login)
		api.GET("/route/getConstantRoutes", routeController.GetConstantRoutes)

		// ========== 用户信息路由（需要认证） ==========
		api.GET("/user/info", middleware.Auth(), authController.GetUserInfo)
		api.POST("/logout", middleware.Auth(), authController.Logout)

		// ========== 动态路由接口（需要认证） ==========
		routeGroup := api.Group("/route")
		routeGroup.Use(middleware.Auth())
		routeGroup.Use(middleware.RequirePermissionFromDB()) // ✅ 统一权限检查
		{
			routeGroup.GET("/getUserRoutes", routeController.GetUserRoutes)
			routeGroup.GET("/isRouteExist", routeController.IsRouteExist)
			routeGroup.POST("/invalidateCache", routeController.InvalidateCache)
			routeGroup.GET("/debugCache", routeController.DebugCache)
		}
	}

	// ========== 模块化路由注册 ==========
	// 系统管理模块
	SetupSystemRoutes(r, menuController, roleController, userController, attributeController)

	// 授权中心模块
	SetupAuthRoutes(r, applicationPermissionController)

	// 审计中心模块
	SetupAuditRoutes(r, auditController)

	// 监控中心模块
	SetupMonitoringRoutes(r, monitoringController, wsMonitoringController)

	// 资产管理模块 (CMDB)
	SetupCMDBRoutes(r, cmdbController, bastionController, attributeController, sshHandler)

	// K8s管理模块
	SetupK8sRoutes(r, k8sClusterController, k8sPermissionController, k8sResourceController, k8sTerminalHandler, diagnosticController)
}
