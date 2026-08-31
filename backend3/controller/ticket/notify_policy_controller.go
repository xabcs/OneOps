package controllerticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	serviceticket "oneops/backend3/service/ticket"
)

// NotifyPolicyController 工单通知设置（事件矩阵）控制器
type NotifyPolicyController struct {
	svc *serviceticket.NotifyPolicyService
}

// NewNotifyPolicyController 创建控制器
func NewNotifyPolicyController(svc *serviceticket.NotifyPolicyService) *NotifyPolicyController {
	return &NotifyPolicyController{svc: svc}
}

// GetPolicies godoc
// @Summary      获取工单通知事件矩阵
// @Tags         工单-通知设置
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/notify-policies [get]
// @Security     BearerAuth
func (ctrl *NotifyPolicyController) GetPolicies(c *gin.Context) {
	views, err := ctrl.svc.GetPolicies()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取通知设置失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(views))
}

// SavePolicyRequest 保存通知策略请求体
type SavePolicyRequest struct {
	Channels []uint `json:"channels"`
	Enabled  int    `json:"enabled"`
	TitleTpl string `json:"titleTpl"` // 标题模板（空=内置默认）
	BodyTpl  string `json:"bodyTpl"`  // 正文模板（空=内置默认）
}

// SavePolicy godoc
// @Summary      保存某通知事件的渠道绑定、启用状态与通知模板
// @Tags         工单-通知设置
// @Accept       json
// @Produce      json
// @Param        event  path  string  true  "事件: pending/result/reassign/cancel/urge/timeout/escalation"
// @Param        body   body  SavePolicyRequest  true  "渠道ID列表、启用状态与模板"
// @Success      200  {object}  utils.Response
// @Router       /ticket/notify-policies/{event} [put]
// @Security     BearerAuth
func (ctrl *NotifyPolicyController) SavePolicy(c *gin.Context) {
	event := c.Param("event")
	var req SavePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}
	if err := ctrl.svc.SavePolicy(event, req.Channels, req.Enabled, req.TitleTpl, req.BodyTpl); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("保存成功"))
}

// GetLogs godoc
// @Summary      分页查询通知发送记录（排障：谁、哪个渠道、成功/失败）
// @Tags         工单-通知设置
// @Produce      json
// @Param        page     query  int  false  "页码"
// @Param        pageSize query  int  false  "每页条数"
// @Param        event    query  string  false  "事件过滤"
// @Param        status   query  int  false  "状态过滤: 0失败 1成功"
// @Success      200  {object}  utils.Response
// @Router       /ticket/notify-logs [get]
// @Security     BearerAuth
func (ctrl *NotifyPolicyController) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	event := c.Query("event")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	rows, total, err := ctrl.svc.ListLogs(page, pageSize, event, status)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("查询发送记录失败"))
		return
	}
	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(rows, total, dto.BasePageQuery{Page: page, PageSize: pageSize})))
}

// SaveQuietRequest 保存夜间静默期配置请求体
type SaveQuietRequest struct {
	QuietEnabled   int `json:"quietEnabled"`
	QuietStartHour int `json:"quietStartHour"`
	QuietEndHour   int `json:"quietEndHour"`
}

// GetQuietConfig godoc
// @Summary      读取夜间静默期配置（防轰炸）
// @Tags         工单-通知设置
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/notify-quiet [get]
// @Security     BearerAuth
func (ctrl *NotifyPolicyController) GetQuietConfig(c *gin.Context) {
	cfg, err := ctrl.svc.GetQuietConfig()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("读取静默期配置失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(cfg))
}

// SaveQuietConfig godoc
// @Summary      保存夜间静默期配置（防轰炸）
// @Tags         工单-通知设置
// @Accept       json
// @Produce      json
// @Param        body  body  SaveQuietRequest  true  "启用状态与起止小时"
// @Success      200  {object}  utils.Response
// @Router       /ticket/notify-quiet [put]
// @Security     BearerAuth
func (ctrl *NotifyPolicyController) SaveQuietConfig(c *gin.Context) {
	var req SaveQuietRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}
	if err := ctrl.svc.SaveQuietConfig(req.QuietEnabled, req.QuietStartHour, req.QuietEndHour); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("保存成功（5 分钟内生效）"))
}
