package controllerticket

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"oneops/backend3/pkg/utils"
	serviceticket "oneops/backend3/service/ticket"
)

// UserNotifySettingController 用户通知偏好（登录用户本人自助读写）
type UserNotifySettingController struct {
	svc *serviceticket.UserNotifySettingService
}

// NewUserNotifySettingController 创建控制器
func NewUserNotifySettingController(svc *serviceticket.UserNotifySettingService) *UserNotifySettingController {
	return &UserNotifySettingController{svc: svc}
}

// GetMySetting godoc
// @Summary      读取本人通知偏好
// @Tags         工单-通知偏好
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /ticket/my-notify-setting [get]
// @Security     BearerAuth
func (ctrl *UserNotifySettingController) GetMySetting(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("未登录"))
		return
	}
	data, err := ctrl.svc.GetSetting(uid)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("查询通知偏好失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(data))
}

// SaveMySetting godoc
// @Summary      保存本人通知偏好
// @Tags         工单-通知偏好
// @Accept       json
// @Produce      json
// @Param        body  body  serviceticket.SaveUserNotifySettingRequest  true  "偏好内容"
// @Success      200  {object}  utils.Response
// @Router       /ticket/my-notify-setting [put]
// @Security     BearerAuth
func (ctrl *UserNotifySettingController) SaveMySetting(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("未登录"))
		return
	}
	var req serviceticket.SaveUserNotifySettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}
	if err := ctrl.svc.SaveSetting(uid, &req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("保存通知偏好失败"))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("通知偏好已保存"))
}
