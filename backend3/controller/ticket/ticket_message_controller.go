package controllerticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	serviceticket "oneops/backend3/service/ticket"
)

// TicketMessageController 工单站内消息（登录用户本人）
type TicketMessageController struct {
	svc *serviceticket.TicketMessageService
}

// NewTicketMessageController 创建控制器
func NewTicketMessageController(svc *serviceticket.TicketMessageService) *TicketMessageController {
	return &TicketMessageController{svc: svc}
}

// currentUserID 从上下文取当前登录用户 ID（Auth 中间件注入）
func currentUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	switch id := v.(type) {
	case uint:
		return id, true
	case float64:
		return uint(id), true
	case string:
		n, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(n), true
	}
	return 0, false
}

// ListMessages godoc
// @Summary      分页查询本人站内消息
// @Tags         工单-站内消息
// @Produce      json
// @Param        page       query  int  false  "页码"
// @Param        pageSize   query  int  false  "每页条数"
// @Param        unreadOnly query  int  false  "1=仅未读"
// @Success      200  {object}  utils.Response
// @Router       /ticket/messages [get]
// @Security     BearerAuth
func (ctrl *TicketMessageController) ListMessages(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("未登录"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	unreadOnly := c.DefaultQuery("unreadOnly", "0") == "1"

	rows, total, err := ctrl.svc.ListMessages(uid, page, pageSize, unreadOnly)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("查询消息失败"))
		return
	}
	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(rows, total, dto.BasePageQuery{Page: page, PageSize: pageSize})))
}

// UnreadCount godoc
// @Summary      本人未读消息数
// @Tags         工单-站内消息
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/messages/unread-count [get]
// @Security     BearerAuth
func (ctrl *TicketMessageController) UnreadCount(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("未登录"))
		return
	}
	n, err := ctrl.svc.UnreadCount(uid)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("查询未读数失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(n))
}

// MarkReadRequest 标记已读请求体
type MarkReadRequest struct {
	IDs []uint `json:"ids"` // 为空表示全部已读
}

// MarkRead godoc
// @Summary      标记消息已读（ids 为空=全部已读）
// @Tags         工单-站内消息
// @Accept       json
// @Produce      json
// @Param        body  body  MarkReadRequest  true  "消息 ID 列表"
// @Success      200  {object}  utils.Response
// @Router       /ticket/messages/read [put]
// @Security     BearerAuth
func (ctrl *TicketMessageController) MarkRead(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("未登录"))
		return
	}
	var req MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}
	if err := ctrl.svc.MarkRead(uid, req.IDs); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("操作失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("已标记已读"))
}
