package authorization

import (
	"net/http"
	"strconv"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/authorization"

	"github.com/gin-gonic/gin"
)

// ApplicationPermissionController 应用权限控制器
type ApplicationPermissionController struct {
	svc *ApplicationPermissionService
}

// NewApplicationPermissionController 创建控制器
func NewApplicationPermissionController(svc *ApplicationPermissionService) *ApplicationPermissionController {
	return &ApplicationPermissionController{svc: svc}
}

// === 应用管理 ===

// GetApplications godoc
// @Summary      获取应用列表
// @Description  分页获取授权应用列表，支持按名称搜索
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(20)
// @Param        name      query     string  false  "应用名称"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "获取应用列表失败"
// @Router       /system/applications [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetApplications(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	name := c.Query("name")

	apps, total, err := ctrl.svc.GetApplications(params.GetPage(), params.GetPageSize(), name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取应用列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(apps, total, params)))
}

// GetApplicationOptions godoc
// @Summary      获取应用选项列表
// @Description  返回所有应用的精简信息（不分页，用于选择器）
// @Tags         应用授权-应用管理
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "获取应用选项失败"
// @Router       /system/applications/options [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetApplicationOptions(c *gin.Context) {
	apps, err := ctrl.svc.GetAllApplicationOptions()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取应用选项失败"))
		return
	}

	options := make([]gin.H, len(apps))
	for i, a := range apps {
		options[i] = gin.H{
			"id":   a.ID,
			"name": a.Name,
			"code": a.Code,
			"type": a.Type,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(options))
}

// CreateApplication godoc
// @Summary      创建应用
// @Description  新增一个授权应用
// @Tags         应用授权-应用管理
// @Accept       json
// @Produce      json
// @Param        body  body      modelauth.Application  true  "应用信息"
// @Success      200   {object}  utils.Response{data=modelauth.Application}
// @Failure      200   {object}  utils.Response  "参数错误 / 创建应用失败"
// @Router       /system/applications [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) CreateApplication(c *gin.Context) {
	var app modelauth.Application
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	if err := ctrl.svc.CreateApplication(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建应用失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(app))
}

// UpdateApplication godoc
// @Summary      更新应用
// @Description  按应用 ID 更新应用信息
// @Tags         应用授权-应用管理
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "应用 ID"
// @Param        body  body      modelauth.Application  true  "应用信息"
// @Success      200   {object}  utils.Response{data=modelauth.Application}
// @Failure      200   {object}  utils.Response  "参数错误 / 更新应用失败"
// @Router       /system/applications/{id} [put]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) UpdateApplication(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var app modelauth.Application
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	app.ID = uint(id)
	if err := ctrl.svc.UpdateApplication(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新应用失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(app))
}

// DeleteApplication godoc
// @Summary      删除应用
// @Description  按应用 ID 删除授权应用
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "删除应用失败"
// @Router       /system/applications/{id} [delete]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) DeleteApplication(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteApplication(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除应用失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 角色同步 ===

// SyncRoles godoc
// @Summary      同步应用角色
// @Description  从外部应用同步角色信息到本地
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "同步失败"
// @Router       /system/applications/{id}/sync-roles [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) SyncRoles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := ctrl.svc.SyncRoles(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationRoles godoc
// @Summary      获取应用角色列表
// @Description  返回指定应用的角色列表
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response{data=[]modelauth.ApplicationRole}
// @Failure      200  {object}  utils.Response  "获取角色列表失败"
// @Router       /system/applications/{id}/roles [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetApplicationRoles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	roles, err := ctrl.svc.GetApplicationRoles(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(roles))
}

// === 用户同步 ===

// SyncUsers godoc
// @Summary      同步应用用户
// @Description  从外部应用同步用户信息到本地
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "同步失败"
// @Router       /system/applications/{id}/sync-users [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) SyncUsers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := ctrl.svc.SyncUsers(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationUsers godoc
// @Summary      获取应用用户列表
// @Description  返回指定应用的用户列表
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response{data=[]modelauth.ApplicationUser}
// @Failure      200  {object}  utils.Response  "获取用户列表失败"
// @Router       /system/applications/{id}/users [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetApplicationUsers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	users, err := ctrl.svc.GetApplicationUsers(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}

// === 角色绑定 ===

// GetGroupBindings godoc
// @Summary      获取用户组绑定列表
// @Description  返回指定用户组的权限绑定列表
// @Tags         应用授权-用户组绑定
// @Produce      json
// @Param        id  path      int  true  "用户组 ID"
// @Success      200  {object}  utils.Response{data=[]modelauth.GroupBinding}
// @Failure      200  {object}  utils.Response  "获取用户组绑定失败"
// @Router       /system/groups/{id}/bindings [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetGroupBindings(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	bindings, err := ctrl.svc.GetGroupBindings(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(bindings))
}

// CreateGroupBinding godoc
// @Summary      创建用户组绑定
// @Description  为指定用户组创建权限绑定
// @Tags         应用授权-用户组绑定
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "用户组 ID"
// @Param        body  body      modelauth.GroupBinding  true  "绑定信息"
// @Success      200   {object}  utils.Response{data=modelauth.GroupBinding}
// @Failure      200   {object}  utils.Response  "参数错误 / 创建用户组绑定失败"
// @Router       /system/groups/{id}/bindings [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) CreateGroupBinding(c *gin.Context) {
	var binding modelauth.GroupBinding
	if err := c.ShouldBindJSON(&binding); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	if err := ctrl.svc.CreateGroupBinding(&binding); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(binding))
}

// DeleteGroupBinding godoc
// @Summary      删除用户组绑定
// @Description  按绑定 ID 删除用户组权限绑定
// @Tags         应用授权-用户组绑定
// @Produce      json
// @Param        id  path      int  true  "绑定 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "删除用户组绑定失败"
// @Router       /system/groups/bindings/{id} [delete]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) DeleteGroupBinding(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteGroupBinding(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 用户组成员分配 ===

// GetUserGroups godoc
// @Summary      获取用户所属用户组
// @Description  返回指定用户加入的用户组列表
// @Tags         应用授权-用户组成员
// @Produce      json
// @Param        id  path      int  true  "用户 ID"
// @Success      200  {object}  utils.Response{data=[]modelauth.AuthGroup}
// @Failure      200  {object}  utils.Response  "获取用户组失败"
// @Router       /system/users/{id}/groups [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetUserGroups(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	groups, err := ctrl.svc.GetUserGroups(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// AssignUserToGroup godoc
// @Summary      分配用户到用户组
// @Description  将指定用户加入用户组，并触发权限下发
// @Tags         应用授权-用户组成员
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "分配请求"  examples({\"userId\":1,\"groupId\":2})
// @Success      200   {object}  utils.Response{data=object{message=string,results=object}}
// @Failure      200   {object}  utils.Response  "参数错误 / 授权失败"
// @Router       /system/users/assign-group [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) AssignUserToGroup(c *gin.Context) {
	var req struct {
		UserID  uint `json:"userId" binding:"required"`
		GroupID uint `json:"groupId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	results, err := ctrl.svc.AssignUserToGroup(req.UserID, req.GroupID, operator)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"message": "授权成功",
		"results": results,
	}))
}

// DeleteUserGroup godoc
// @Summary      删除用户组成员
// @Description  将指定用户从用户组中移除
// @Tags         应用授权-用户组成员
// @Produce      json
// @Param        id       path      int  true  "用户 ID"
// @Param        groupId  path      int  true  "用户组 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "删除用户组成员失败"
// @Router       /system/users/{id}/groups/{groupId} [delete]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) DeleteUserGroup(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	groupID, _ := strconv.ParseUint(c.Param("groupId"), 10, 32)

	if err := ctrl.svc.DeleteUserGroup(uint(userID), uint(groupID)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组成员失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 操作日志 ===

// GetOperationLogs godoc
// @Summary      获取应用操作日志
// @Description  分页获取指定授权应用的操作日志
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id        path      int  true  "应用 ID"
// @Param        page      query     int  false  "页码"    default(1)
// @Param        pageSize  query     int  false  "每页数量" default(20)
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "获取操作日志失败"
// @Router       /system/applications/{id}/operation-logs [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetOperationLogs(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	logs, total, err := ctrl.svc.GetOperationLogs(uint(id), params.GetPage(), params.GetPageSize())
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取操作日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(logs, total, params)))
}

// ========== 授权中心用户管理 ==========

// GetAuthUsers godoc
// @Summary      获取授权中心用户列表
// @Description  分页获取授权中心用户，支持按用户名搜索
// @Tags         应用授权-授权用户
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(20)
// @Param        username  query     string  false  "用户名"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "获取用户列表失败"
// @Router       /system/auth-users/list [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetAuthUsers(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	username := c.Query("username")

	users, total, err := ctrl.svc.GetAuthUsers(params.GetPage(), params.GetPageSize(), username)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(users, total, params)))
}

// GetAuthUserPassword godoc
// @Summary      获取用户初始密码
// @Description  返回指定授权用户的初始密码
// @Tags         应用授权-授权用户
// @Produce      json
// @Param        id  path      int  true  "用户 ID"
// @Success      200  {object}  utils.Response{data=object{username=string,password=string}}
// @Failure      200  {object}  utils.Response  "获取用户失败 / 用户未设置初始密码"
// @Router       /system/auth-users/{id}/password [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetAuthUserPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	user, err := ctrl.svc.GetAuthUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户失败"))
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusOK, utils.ErrorInternal("用户未设置初始密码"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"username": user.Username,
		"password": user.Password,
	}))
}

// CreateAuthUser godoc
// @Summary      创建授权中心用户
// @Description  新建授权中心用户并生成初始密码
// @Tags         应用授权-授权用户
// @Accept       json
// @Produce      json
// @Param        body  body      modelauth.AuthUser  true  "用户信息"
// @Success      200   {object}  utils.Response{data=object}
// @Failure      200   {object}  utils.Response  "参数错误 / 创建用户失败"
// @Router       /system/auth-users [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) CreateAuthUser(c *gin.Context) {
	var req modelauth.AuthUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	password, err := ctrl.svc.CreateAuthUser(&req)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"id":          req.ID,
		"username":    req.Username,
		"nickname":    req.Nickname,
		"email":       req.Email,
		"phone":       req.Phone,
		"description": req.Description,
		"status":      req.Status,
		"password":    password,
		"message":     "用户创建成功，请妥善保管初始密码",
	}))
}

// UpdateAuthUser godoc
// @Summary      更新授权中心用户
// @Description  按用户 ID 更新授权中心用户信息
// @Tags         应用授权-授权用户
// @Accept       json
// @Produce      json
// @Param        id    path      int                 true  "用户 ID"
// @Param        body  body      modelauth.AuthUser  true  "用户信息"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "参数错误 / 更新用户失败"
// @Router       /system/auth-users/{id} [put]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) UpdateAuthUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req modelauth.AuthUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	req.ID = uint(id)

	if err := ctrl.svc.UpdateAuthUser(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// DeleteAuthUser godoc
// @Summary      删除授权中心用户
// @Description  按用户 ID 删除授权中心用户
// @Tags         应用授权-授权用户
// @Produce      json
// @Param        id  path      int  true  "用户 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "删除用户失败"
// @Router       /system/auth-users/{id} [delete]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) DeleteAuthUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteAuthUser(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetAllAuthUsers godoc
// @Summary      获取全部授权用户
// @Description  返回全部授权中心用户（不分页，用于下拉选择等）
// @Tags         应用授权-授权用户
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]modelauth.AuthUser}
// @Failure      200  {object}  utils.Response  "获取用户列表失败"
// @Router       /system/auth-users [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetAllAuthUsers(c *gin.Context) {
	users, err := ctrl.svc.GetAllAuthUsers()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}

// ========== 授权中心角色管理 ==========

// GetAuthGroups godoc
// @Summary      获取授权中心用户组列表
// @Description  分页获取授权中心用户组，支持按名称搜索
// @Tags         应用授权-授权用户组
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(20)
// @Param        name      query     string  false  "用户组名称"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "获取用户组列表失败"
// @Router       /system/auth-groups/list [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetAuthGroups(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	name := c.Query("name")

	groups, total, err := ctrl.svc.GetAuthGroups(params.GetPage(), params.GetPageSize(), name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(groups, total, params)))
}

// CreateAuthGroup godoc
// @Summary      创建授权中心用户组
// @Description  新建授权中心用户组
// @Tags         应用授权-授权用户组
// @Accept       json
// @Produce      json
// @Param        body  body      modelauth.AuthGroup  true  "用户组信息"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "参数错误 / 创建用户组失败"
// @Router       /system/auth-groups [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) CreateAuthGroup(c *gin.Context) {
	var req modelauth.AuthGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	if err := ctrl.svc.CreateAuthGroup(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// UpdateAuthGroup godoc
// @Summary      更新授权中心用户组
// @Description  按用户组 ID 更新授权中心用户组信息
// @Tags         应用授权-授权用户组
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "用户组 ID"
// @Param        body  body      modelauth.AuthGroup  true  "用户组信息"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "参数错误 / 更新用户组失败"
// @Router       /system/auth-groups/{id} [put]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) UpdateAuthGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req modelauth.AuthGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	req.ID = uint(id)

	if err := ctrl.svc.UpdateAuthGroup(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// DeleteAuthGroup godoc
// @Summary      删除授权中心用户组
// @Description  按用户组 ID 删除授权中心用户组
// @Tags         应用授权-授权用户组
// @Produce      json
// @Param        id  path      int  true  "用户组 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "删除用户组失败"
// @Router       /system/auth-groups/{id} [delete]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) DeleteAuthGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteAuthGroup(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetAllAuthGroups godoc
// @Summary      获取全部授权用户组
// @Description  返回全部授权中心用户组（不分页，用于下拉选择等）
// @Tags         应用授权-授权用户组
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]modelauth.AuthGroup}
// @Failure      200  {object}  utils.Response  "获取用户组列表失败"
// @Router       /system/auth-groups [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetAllAuthGroups(c *gin.Context) {
	groups, err := ctrl.svc.GetAllAuthGroups()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// GetAuthGroupOptions godoc
// @Summary      获取用户组选项列表
// @Description  返回所有用户组的精简信息（不分页，用于选择器）
// @Tags         应用授权-授权用户组
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "获取用户组选项失败"
// @Router       /system/auth-groups/options [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetAuthGroupOptions(c *gin.Context) {
	groups, err := ctrl.svc.GetAllAuthGroups()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组选项失败"))
		return
	}

	options := make([]gin.H, len(groups))
	for i, g := range groups {
		options[i] = gin.H{
			"id":   g.ID,
			"name": g.Name,
			"code": g.Code,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(options))
}

// === 应用类型配置 ===

// GetSupportedAppTypes godoc
// @Summary      获取支持的应用类型
// @Description  返回系统支持的授权应用类型列表
// @Tags         应用授权-应用类型
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}
// @Router       /system/applications/types [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetSupportedAppTypes(c *gin.Context) {
	types := ctrl.svc.GetSupportedAppTypes()
	c.JSON(http.StatusOK, utils.SuccessWithData(types))
}

// GetAppTypeConfigTemplate godoc
// @Summary      获取应用类型配置模板
// @Description  返回指定应用类型的默认配置模板
// @Tags         应用授权-应用类型
// @Produce      json
// @Param        type  path      string  true  "应用类型"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "应用类型参数缺失 / 获取配置模板失败"
// @Router       /system/applications/types/{type}/config [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetAppTypeConfigTemplate(c *gin.Context) {
	appType := c.Param("type")
	if appType == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("应用类型参数缺失"))
		return
	}

	template, err := ctrl.svc.GetAppTypeConfigTemplate(appType)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取配置模板失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(template))
}

// SyncApplicationGroups godoc
// @Summary      同步应用用户组
// @Description  从外部应用同步用户组信息到本地
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "无效的应用ID / 同步失败"
// @Router       /system/applications/{id}/sync-groups [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) SyncApplicationGroups(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的应用ID"))
		return
	}

	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := ctrl.svc.SyncGroups(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationGroups godoc
// @Summary      获取应用用户组列表
// @Description  返回指定应用已同步的用户组列表
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response{data=[]modelauth.ApplicationGroup}
// @Failure      200  {object}  utils.Response  "无效的应用ID / 获取失败"
// @Router       /system/applications/{id}/groups [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetApplicationGroups(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的应用ID"))
		return
	}

	groups, err := ctrl.svc.GetApplicationGroups(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// === 授权规则管理 ===

// SyncAuthorizationRules godoc
// @Summary      同步应用授权规则
// @Description  从外部应用同步授权规则到本地
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "同步失败"
// @Router       /system/applications/{id}/sync-rules [post]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) SyncAuthorizationRules(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := ctrl.svc.SyncAuthorizationRules(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationAuthorizationRules godoc
// @Summary      获取应用授权规则
// @Description  返回指定应用的授权规则列表
// @Tags         应用授权-应用管理
// @Produce      json
// @Param        id  path      int  true  "应用 ID"
// @Success      200  {object}  utils.Response{data=[]modelauth.ApplicationAuthorizationRule}
// @Failure      200  {object}  utils.Response  "无效的应用ID / 获取失败"
// @Router       /system/applications/{id}/rules [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetApplicationAuthorizationRules(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的应用ID"))
		return
	}

	rules, err := ctrl.svc.GetApplicationAuthorizationRules(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(rules))
}

// === 用户身份映射管理 ===

// GetUserIdentityMappings godoc
// @Summary      获取用户身份映射列表
// @Description  分页获取用户身份映射，支持按用户名、应用 ID、状态筛选
// @Tags         应用授权-身份映射
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(20)
// @Param        username  query     string  false  "用户名"
// @Param        appId     query     int     false  "应用 ID"
// @Param        status    query     string  false  "状态"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "获取用户身份映射列表失败"
// @Router       /system/user-identity-mappings [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetUserIdentityMappings(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	username := c.Query("username")
	appIDStr := c.Query("appId")
	status := c.Query("status")

	var appID uint
	if appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		appID = uint(id)
	}

	mappings, total, err := ctrl.svc.GetUserIdentityMappings(params.GetPage(), params.GetPageSize(), username, appID, status)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户身份映射列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(mappings, total, params)))
}

// DeleteUserIdentityMapping godoc
// @Summary      删除用户身份映射
// @Description  按映射 ID 删除用户身份映射
// @Tags         应用授权-身份映射
// @Produce      json
// @Param        id  path      int  true  "映射 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "无效的身份映射ID / 删除失败"
// @Router       /system/user-identity-mappings/{id} [delete]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) DeleteUserIdentityMapping(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的身份映射ID"))
		return
	}

	if err := ctrl.svc.DeleteUserIdentityMapping(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户身份映射失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 用户有效权限查询 ===

// GetUserEffectivePermissions godoc
// @Summary      获取用户有效权限
// @Description  分页获取用户在应用上的有效权限，支持按用户名、应用 ID 筛选
// @Tags         应用授权-权限查询
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(20)
// @Param        username  query     string  false  "用户名"
// @Param        appId     query     int     false  "应用 ID"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "获取用户有效权限列表失败"
// @Router       /system/user-permissions [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetUserEffectivePermissions(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	username := c.Query("username")
	appIDStr := c.Query("appId")

	var appID uint
	if appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		appID = uint(id)
	}

	permissions, total, err := ctrl.svc.GetUserEffectivePermissions(params.GetPage(), params.GetPageSize(), username, appID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户有效权限列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(permissions, total, params)))
}

// GetUserEffectivePermissionsMatrix godoc
// @Summary      获取用户有效权限矩阵
// @Description  返回指定应用下所有用户的有效权限矩阵视图
// @Tags         应用授权-权限查询
// @Produce      json
// @Param        appId  query     int  true  "应用 ID"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "缺少应用ID参数 / 获取权限矩阵失败"
// @Router       /system/user-permissions/matrix [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetUserEffectivePermissionsMatrix(c *gin.Context) {
	appIDStr := c.Query("appId")
	if appIDStr == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("缺少应用ID参数"))
		return
	}

	appID, err := strconv.ParseUint(appIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的应用ID"))
		return
	}

	result, err := ctrl.svc.GetUserEffectivePermissionsMatrix(uint(appID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取权限矩阵失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(result))
}

// === 权限执行记录 ===

// GetGroupBindingExecutions godoc
// @Summary      获取权限绑定执行记录
// @Description  返回指定用户组绑定的权限下发执行记录
// @Tags         应用授权-用户组绑定
// @Produce      json
// @Param        bindingId  path      int  true  "绑定 ID"
// @Success      200  {object}  utils.Response{data=[]modelauth.GroupBindingExecution}
// @Failure      200  {object}  utils.Response  "无效的绑定ID / 获取执行记录失败"
// @Router       /system/group-bindings/{bindingId}/executions [get]
// @Security     BearerAuth
func (ctrl *ApplicationPermissionController) GetGroupBindingExecutions(c *gin.Context) {
	bindingID, err := strconv.ParseUint(c.Param("bindingId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的绑定ID"))
		return
	}

	executions, err := ctrl.svc.GetGroupBindingExecutions(uint(bindingID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取执行记录失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(executions))
}
