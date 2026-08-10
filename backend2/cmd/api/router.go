package main

import (
	"oneops/backend2/internal/audit"
	"oneops/backend2/internal/authorization"
	"oneops/backend2/internal/cmdb"
	"oneops/backend2/internal/k8s"
	"oneops/backend2/internal/monitoring"
	"oneops/backend2/internal/system"
	"oneops/backend2/pkg/database"
	"oneops/backend2/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 应用中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Response())
	r.Use(middleware.ErrorHandler())
	r.Use(cors.New(middleware.CORS()))

	// 创建审计中间件
	auditMiddleware := audit.NewAuditMiddleware()
	r.Use(auditMiddleware.OperationLog())

	// 创建 K8s 相关服务
	k8sClientPool := k8s.NewK8sClientPool()
	k8sIdentitySvc := k8s.NewK8sIdentityService(k8sClientPool)
	k8sClusterSvc := k8s.NewK8sClusterService(k8sClientPool, k8sIdentitySvc)
	k8sResourceSvc := k8s.NewK8sResourceService(k8sClientPool)
	db := database.GetDB()

	// 创建控制器
	authController := system.NewAuthController(system.NewAuthService(), audit.NewAuditService())
	menuController := system.NewMenuController()
	roleController := system.NewRoleController()
	userController := system.NewUserController()
	auditController := audit.NewAuditController()
	monitoringController := monitoring.NewMonitoringController()
	wsMonitoringController := monitoring.NewMonitoringWebSocketController()
	routeController := system.NewRouteController()
	cmdbController := cmdb.NewCMDBController(cmdb.NewCMDBService())
	bastionController := cmdb.NewBastionController()
	attributeController := system.NewAttributeController()
	sshHandler := cmdb.NewSSHWebSocketHandler()
	k8sClusterController := k8s.NewK8sClusterController(k8sClusterSvc)
	k8sPermissionController := k8s.NewK8sPermissionController(k8sClusterSvc, db)
	k8sResourceController := k8s.NewK8sResourceController(k8sClusterSvc, k8sResourceSvc)
	k8sTerminalHandler := k8s.NewK8sTerminalHandler(k8sClusterSvc, k8sClientPool, db)
	diagnosticController := k8s.NewDiagnosticController(k8sClusterSvc, k8sClientPool, db)
	applicationPermissionController := authorization.NewApplicationPermissionController()

	// API 路由组
	api := r.Group("/api")
	{
		api.POST("/login", authController.Login)
		api.GET("/route/getConstantRoutes", routeController.GetConstantRoutes)

		api.GET("/user/info", middleware.Auth(), authController.GetUserInfo)
		api.POST("/logout", middleware.Auth(), authController.Logout)

		routeGroup := api.Group("/route")
		routeGroup.Use(middleware.Auth())
		routeGroup.Use(system.RequirePermissionFromDB())
		{
			routeGroup.GET("/getUserRoutes", routeController.GetUserRoutes)
			routeGroup.GET("/isRouteExist", routeController.IsRouteExist)
			routeGroup.POST("/invalidateCache", routeController.InvalidateCache)
			routeGroup.GET("/debugCache", routeController.DebugCache)
		}
	}

	// 模块化路由注册
	SetupSystemRoutes(r, menuController, roleController, userController, attributeController)
	SetupAuthRoutes(r, applicationPermissionController)
	SetupAuditRoutes(r, auditController)
	SetupMonitoringRoutes(r, monitoringController, wsMonitoringController)
	SetupCMDBRoutes(r, cmdbController, bastionController, attributeController, sshHandler)
	SetupK8sRoutes(r, k8sClusterController, k8sPermissionController, k8sResourceController, k8sTerminalHandler, diagnosticController)
}
