package routes

import (
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/middleware"

	controllerticket "oneops/backend3/controller/ticket"
	repoticket "oneops/backend3/repository/ticket"
	serviceticket "oneops/backend3/service/ticket"

	"github.com/gin-gonic/gin"
)

// SetupTicketRoutes 工单系统路由
//
// 权限设计说明：
//   - 工单发起/审批/详情/评论/类型选项：仅需登录（Auth）。
//     审批权来自流程节点配置（业务层校验"是否当前节点审批人"），
//     所有登录用户均可发起工单与查看自己相关的工单。
//   - 流程定义/工单类型的管理 CRUD、查看全部工单（scope=all）：
//     走 RBAC（Auth + RequirePermissionFromDB），权限码见 seed_permission.go。
func SetupTicketRoutes(r *gin.Engine) {
	db := database.GetDB()

	wfRepo := repoticket.NewWorkflowRepository(db)
	ticketRepo := repoticket.NewTicketRepository(db)

	wfSvc := serviceticket.NewWorkflowService(wfRepo)
	ticketSvc := serviceticket.NewTicketService(ticketRepo, wfRepo)

	wfController := controllerticket.NewWorkflowController(wfSvc)
	ticketController := controllerticket.NewTicketController(ticketSvc)

	api := r.Group("/api")
	ticket := api.Group("/ticket")
	ticket.Use(middleware.Auth())
	{
		// ── 工单中心（登录即可，可见性/审批权由业务层校验） ──
		ticket.GET("/tickets", ticketController.GetTickets)
		ticket.POST("/tickets", ticketController.CreateTicket)
		ticket.GET("/tickets/:id", ticketController.GetTicketDetail)
		ticket.POST("/tickets/:id/approve", ticketController.ApproveTicket)
		ticket.POST("/tickets/:id/reject", ticketController.RejectTicket)
		ticket.POST("/tickets/:id/cancel", ticketController.CancelTicket)
		ticket.POST("/tickets/:id/comment", ticketController.CommentTicket)

		// 发起工单时的类型选项与场景下流程选项（仅需登录）
		ticket.GET("/types/options", wfController.GetTypeOptions)
		ticket.GET("/workflows/by-type/:typeId", wfController.GetWorkflowsByType)
	}

	// ── 管理端（认证 + RBAC 权限校验） ──
	manage := api.Group("/ticket")
	manage.Use(middleware.Auth())
	manage.Use(middleware.RequirePermissionFromDB())
	{
		// 流程定义
		manage.GET("/workflows", wfController.GetWorkflows)
		manage.GET("/workflows/:id", wfController.GetWorkflow)
		manage.POST("/workflows", wfController.CreateWorkflow)
		manage.PUT("/workflows/:id", wfController.UpdateWorkflow)
		manage.DELETE("/workflows/:id", wfController.DeleteWorkflow)

		// 工单类型
		manage.GET("/types", wfController.GetTypes)
		manage.POST("/types", wfController.CreateType)
		manage.PUT("/types/:id", wfController.UpdateType)
		manage.DELETE("/types/:id", wfController.DeleteType)
	}
}
