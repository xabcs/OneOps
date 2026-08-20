package system

import (
	"net/http"
	"strconv"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/utils"

	syssvc "oneops/backend3/service/system"

	"github.com/gin-gonic/gin"
)

// UserGroupController 平台用户组控制器
type UserGroupController struct {
	svc *syssvc.UserGroupService
}

// NewUserGroupController 创建用户组控制器
func NewUserGroupController(svc *syssvc.UserGroupService) *UserGroupController {
	return &UserGroupController{svc: svc}
}

// GetUserGroups godoc
// @Summary      获取用户组列表
// @Description  分页获取平台用户组，支持按名称/编码过滤，含成员数
// @Tags         系统管理-用户组
// @Produce      json
// @Param        page      query  int    false  "页码"      default(1)
// @Param        pageSize  query  int    false  "每页数量"  default(10)
// @Param        keyword   query  string false  "名称/编码模糊过滤"
// @Success      200  {object}  utils.Response{data=object{list=[]object,total=int}}
// @Router       /system/user-groups [get]
// @Security     BearerAuth
func (ctrl *UserGroupController) GetUserGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	groups, total, err := ctrl.svc.GetUserGroups(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"success":  true,
		"data":     gin.H{"list": groups, "total": total},
		"message":  "success",
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetUserGroupOptions godoc
// @Summary      获取用户组选项
// @Description  获取全部启用用户组的 id/code/name（供选择器）
// @Tags         系统管理-用户组
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]object}
// @Router       /system/user-groups/options [get]
// @Security     BearerAuth
func (ctrl *UserGroupController) GetUserGroupOptions(c *gin.Context) {
	options, err := ctrl.svc.GetUserGroupOptions()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组选项失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(options))
}

// CreateUserGroup godoc
// @Summary      创建用户组
// @Tags         系统管理-用户组
// @Accept       json
// @Produce      json
// @Param        body  body  modelsystem.CreateUserGroupRequest  true  "用户组信息"
// @Success      200   {object}  utils.Response  "创建成功"
// @Router       /system/user-groups [post]
// @Security     BearerAuth
func (ctrl *UserGroupController) CreateUserGroup(c *gin.Context) {
	var req modelsystem.CreateUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	if err := ctrl.svc.CreateUserGroup(req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户组失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建成功"))
}

// UpdateUserGroup godoc
// @Summary      更新用户组
// @Tags         系统管理-用户组
// @Accept       json
// @Produce      json
// @Param        id    path  int                                   true  "用户组 ID"
// @Param        body  body  modelsystem.UpdateUserGroupRequest    true  "用户组信息"
// @Success      200   {object}  utils.Response  "更新成功"
// @Router       /system/user-groups/{id} [put]
// @Security     BearerAuth
func (ctrl *UserGroupController) UpdateUserGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户组ID"))
		return
	}

	var req modelsystem.UpdateUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	if err := ctrl.svc.UpdateUserGroup(uint(id), req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新用户组失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteUserGroup godoc
// @Summary      删除用户组
// @Description  级联删除组成员与集群组绑定
// @Tags         系统管理-用户组
// @Produce      json
// @Param        id  path  int  true  "用户组 ID"
// @Success      200 {object}  utils.Response  "删除成功"
// @Router       /system/user-groups/{id} [delete]
// @Security     BearerAuth
func (ctrl *UserGroupController) DeleteUserGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户组ID"))
		return
	}

	if err := ctrl.svc.DeleteUserGroup(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// GetGroupMembers godoc
// @Summary      获取组成员
// @Tags         系统管理-用户组
// @Produce      json
// @Param        id  path  int  true  "用户组 ID"
// @Success      200 {object}  utils.Response{data=[]object}
// @Router       /system/user-groups/{id}/members [get]
// @Security     BearerAuth
func (ctrl *UserGroupController) GetGroupMembers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户组ID"))
		return
	}

	members, err := ctrl.svc.GetGroupMembers(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取组成员失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(members))
}

// AddGroupMembers godoc
// @Summary      批量添加组成员
// @Tags         系统管理-用户组
// @Accept       json
// @Produce      json
// @Param        id    path  int     true  "用户组 ID"
// @Param        body  body  object  true  "用户ID集合"  examples({\"userIds\":[1,2]})
// @Success      200   {object}  utils.Response  "添加成功"
// @Router       /system/user-groups/{id}/members [post]
// @Security     BearerAuth
func (ctrl *UserGroupController) AddGroupMembers(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户组ID"))
		return
	}

	var req struct {
		UserIDs []uint `json:"userIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	added, err := ctrl.svc.AddGroupMembers(uint(id), req.UserIDs)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("添加组成员失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("添加成功"))
	_ = added
}

// RemoveGroupMember godoc
// @Summary      移除组成员
// @Tags         系统管理-用户组
// @Produce      json
// @Param        id      path  int  true  "用户组 ID"
// @Param        userId  path  int  true  "用户 ID"
// @Success      200     {object}  utils.Response  "移除成功"
// @Router       /system/user-groups/{id}/members/{userId} [delete]
// @Security     BearerAuth
func (ctrl *UserGroupController) RemoveGroupMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户组ID"))
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户ID"))
		return
	}

	if err := ctrl.svc.RemoveGroupMember(uint(id), uint(userID)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("移除组成员失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("移除成功"))
}
