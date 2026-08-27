package controllerticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	repoticket "oneops/backend3/repository/ticket"
	syssvc "oneops/backend3/service/system"
	serviceticket "oneops/backend3/service/ticket"
)

// TicketController 工单控制器
type TicketController struct {
	svc *serviceticket.TicketService
}

func NewTicketController(svc *serviceticket.TicketService) *TicketController {
	return &TicketController{svc: svc}
}

// currentUser 从上下文取当前用户
func currentUser(c *gin.Context) (uint, string) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	uid, _ := userID.(uint)
	uname, _ := username.(string)
	return uid, uname
}

// hasTicketListPermission 是否拥有查看全部工单的权限
func hasTicketListPermission(userID uint) bool {
	permSvc, err := syssvc.GetPermissionService()
	if err != nil {
		return false
	}
	ok, err := permSvc.HasAnyPermission(userID, []string{"ticket.ticket.list"})
	if err != nil {
		return false
	}
	return ok
}

// GetTickets godoc
// @Summary      获取工单列表
// @Description  scope: created=我发起的 / todo=待我审批 / done=我已审批 / all=全部工单(需 ticket.ticket.list 权限)
// @Tags         工单-工单中心
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets [get]
// @Security     BearerAuth
func (ctrl *TicketController) GetTickets(c *gin.Context) {
	params, ok := parsePage(c)
	if !ok {
		return
	}
	userID, _ := currentUser(c)

	scope := c.DefaultQuery("scope", "created")
	if scope == "all" && !hasTicketListPermission(userID) {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权查看全部工单"))
		return
	}

	typeID := uint(0)
	if v := c.Query("typeId"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			typeID = uint(id)
		}
	}

	q := repoticket.TicketQuery{
		Scope:   scope,
		Status:  c.Query("status"),
		TypeID:  typeID,
		Keyword: c.Query("keyword"),
		UserID:  userID,
		Offset:  params.GetOffset(),
		Limit:   params.GetPageSize(),
	}
	items, total, err := ctrl.svc.GetTickets(q)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取工单列表失败"))
		return
	}
	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(items, total, params)))
}

// CreateTicket godoc
// @Summary      发起工单
// @Tags         工单-工单中心
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets [post]
// @Security     BearerAuth
func (ctrl *TicketController) CreateTicket(c *gin.Context) {
	var req serviceticket.TicketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	userID, username := currentUser(c)
	ticket, err := ctrl.svc.CreateTicket(userID, username, req)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"id": ticket.ID, "ticketNo": ticket.TicketNo,
	}))
}

// GetTicketDetail godoc
// @Summary      获取工单详情（发起人/审批人/有权限者可见）
// @Tags         工单-工单中心
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id} [get]
// @Security     BearerAuth
func (ctrl *TicketController) GetTicketDetail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, _ := currentUser(c)
	if !ctrl.svc.CanView(id, userID) && !hasTicketListPermission(userID) {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权查看该工单"))
		return
	}
	detail, err := ctrl.svc.GetTicketDetail(id, userID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(detail))
}

// ApproveTicket godoc
// @Summary      审批通过
// @Tags         工单-工单中心
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id}/approve [post]
// @Security     BearerAuth
func (ctrl *TicketController) ApproveTicket(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req serviceticket.TicketActionRequest
	_ = c.ShouldBindJSON(&req)
	userID, username := currentUser(c)
	if err := ctrl.svc.Approve(id, userID, username, req.Comment); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("审批通过"))
}

// RejectTicket godoc
// @Summary      驳回工单
// @Tags         工单-工单中心
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id}/reject [post]
// @Security     BearerAuth
func (ctrl *TicketController) RejectTicket(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req serviceticket.TicketActionRequest
	_ = c.ShouldBindJSON(&req)
	if req.Comment == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("驳回时必须填写审批意见"))
		return
	}
	userID, username := currentUser(c)
	if err := ctrl.svc.Reject(id, userID, username, req.Comment); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("已驳回"))
}

// CancelTicket godoc
// @Summary      撤销工单（仅发起人）
// @Tags         工单-工单中心
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id}/cancel [post]
// @Security     BearerAuth
func (ctrl *TicketController) CancelTicket(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req serviceticket.TicketActionRequest
	_ = c.ShouldBindJSON(&req)
	userID, username := currentUser(c)
	if err := ctrl.svc.Cancel(id, userID, username, req.Comment); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("已撤销"))
}

// ResubmitTicket godoc
// @Summary      驳回后重新提交（仅发起人，可修改标题/优先级/表单）
// @Tags         工单-工单中心
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id}/resubmit [post]
// @Security     BearerAuth
func (ctrl *TicketController) ResubmitTicket(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req serviceticket.TicketResubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	userID, username := currentUser(c)
	if err := ctrl.svc.Resubmit(id, userID, username, req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("已重新提交"))
}

// UrgeTicket godoc
// @Summary      催办工单（发起人）
// @Description  向当前节点待审批人发送催办提醒；30 分钟冷却频控
// @Tags         工单-工单中心
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id}/urge [post]
func (ctrl *TicketController) UrgeTicket(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, username := currentUser(c)
	if err := ctrl.svc.Urge(id, userID, username); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("已向当前审批人发送催办提醒"))
}

// ReassignTicket godoc
// @Summary      改派当前节点审批人（需 ticket.ticket.reassign 权限）
// @Description  解决审批人离职/请假导致节点卡死；已审批记录保留
// @Tags         工单-工单中心
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id}/reassign [post]
// @Security     BearerAuth
func (ctrl *TicketController) ReassignTicket(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		NodeKey     string `json:"nodeKey"`
		ApproverIDs []uint `json:"approverIds" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	userID, username := currentUser(c)
	if err := ctrl.svc.Reassign(id, userID, username, req.NodeKey, req.ApproverIDs); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("已改派"))
}

// CommentTicket godoc
// @Summary      工单评论
// @Tags         工单-工单中心
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/tickets/{id}/comment [post]
// @Security     BearerAuth
func (ctrl *TicketController) CommentTicket(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req serviceticket.TicketActionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Comment == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("评论内容不能为空"))
		return
	}
	userID, username := currentUser(c)
	if !ctrl.svc.CanView(id, userID) && !hasTicketListPermission(userID) {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权评论该工单"))
		return
	}
	if err := ctrl.svc.Comment(id, userID, username, req.Comment); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("评论成功"))
}
