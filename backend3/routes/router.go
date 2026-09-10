package routes

import (
	"oneops/backend3/config"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/middleware"

	auditsvc "oneops/backend3/service/audit"
	syssvc "oneops/backend3/service/system"

	auditctrl "oneops/backend3/controller/audit"
	sysctrl "oneops/backend3/controller/system"
	repoaudit "oneops/backend3/repository/audit"
	reposystem "oneops/backend3/repository/system"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// 应用中间件（Recovery 已在 main.go 全局注册，此处不再重复挂）
	r.Use(middleware.RequestLogger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORSMiddleware(cfg.CORS))

	// 创建审计中间件并应用操作日志审计
	auditMiddleware := middleware.NewAuditMiddleware()
	r.Use(auditMiddleware.OperationLog())

	db := database.GetDB()

	// 创建 repositories
	menuRepo := reposystem.NewMenuRepository(db)
	roleRepo := reposystem.NewRoleRepository(db)
	userRepo := reposystem.NewUserRepository(db)
	auditRepo := repoaudit.NewAuditRepository(db)

	// 创建 services
	authSvc := syssvc.NewAuthService()
	auditSvc := auditsvc.NewAuditService(auditRepo)
	routeGenSvc := syssvc.NewRouteGenService(db)
	menuSvc := syssvc.NewMenuService(menuRepo)
	roleSvc := syssvc.NewRoleService(roleRepo)
	userSvc := syssvc.NewUserService(userRepo)

	// 创建 controllers
	authController := sysctrl.NewAuthController(authSvc, auditSvc)
	menuController := sysctrl.NewMenuController(menuSvc)
	roleController := sysctrl.NewRoleController(roleSvc)
	userController := sysctrl.NewUserController(userSvc)
	routeController := sysctrl.NewRouteController(routeGenSvc)
	auditController := auditctrl.NewAuditController(auditSvc)

	// API 路由组
	api := r.Group("/api")
	{
		// 认证路由（无需认证）
		api.POST("/login", authController.Login)
		api.GET("/route/getConstantRoutes", routeController.GetConstantRoutes)

		// 用户信息路由（需要认证）
		api.GET("/user/info", middleware.Auth(), authController.GetUserInfo)
		api.POST("/logout", middleware.Auth(), authController.Logout)

		// 动态路由接口（登录后构建菜单必需，仅需认证——返回的本来就是用户自己的路由）
		api.GET("/route/getUserRoutes", middleware.Auth(), routeController.GetUserRoutes)
		// 路由存在性检查（前端路由守卫 404 兜底用，仅需认证——全局存在性布尔值无敏感性）
		api.GET("/route/isRouteExist", middleware.Auth(), routeController.IsRouteExist)

		// 路由管理接口（认证 + 权限检查）
		routeGroup := api.Group("/route")
		routeGroup.Use(middleware.Auth())
		routeGroup.Use(middleware.RequirePermissionFromDB())
		{
			routeGroup.POST("/invalidateCache", routeController.InvalidateCache)
			routeGroup.GET("/debugCache", routeController.DebugCache)
		}
	}

	// ========== 模块化路由注册 ==========
	SetupSystemRoutes(r, menuController, roleController, userController)
	SetupCMDBRoutes(r)
	SetupK8sRoutes(r)
	SetupAuditRoutes(r, auditController)
	SetupMonitoringRoutes(r)
	SetupAuthorizationRoutes(r)
	SetupTicketRoutes(r)
}
