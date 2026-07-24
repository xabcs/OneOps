package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
)

// ApplicationPermissionController 应用权限控制器
type ApplicationPermissionController struct {
	service *services.ApplicationPermissionService
}

// NewApplicationPermissionController 创建控制器
func NewApplicationPermissionController() *ApplicationPermissionController {
	return &ApplicationPermissionController{
		service: services.NewApplicationPermissionService(),
	}
}

// === 应用管理 ===

// GetApplications 获取应用列表
func (ctrl *ApplicationPermissionController) GetApplications(c *gin.Context) {
	name := c.Query("name")
	currentStr := c.DefaultQuery("current", "1")
	sizeStr := c.DefaultQuery("size", "10")

	current, _ := strconv.Atoi(currentStr)
	size, _ := strconv.Atoi(sizeStr)

	apps, total, err := ctrl.service.GetApplications(current, size, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取应用列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"records": apps,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

// CreateApplication 创建应用
func (ctrl *ApplicationPermissionController) CreateApplication(c *gin.Context) {
	var app models.Application
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	if err := ctrl.service.CreateApplication(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建应用失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(app))
}

// UpdateApplication 更新应用
func (ctrl *ApplicationPermissionController) UpdateApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var app models.Application
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	app.ID = uint(id)
	if err := ctrl.service.UpdateApplication(&app); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新应用失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(app))
}

// DeleteApplication 删除应用
func (ctrl *ApplicationPermissionController) DeleteApplication(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ctrl.service.DeleteApplication(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除应用失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 角色同步 ===

// SyncRoles 同步应用角色
func (ctrl *ApplicationPermissionController) SyncRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// 获取操作人
	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := ctrl.service.SyncRoles(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationRoles 获取应用角色列表
func (ctrl *ApplicationPermissionController) GetApplicationRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	roles, err := ctrl.service.GetApplicationRoles(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(roles))
}

// === 用户同步 ===

// SyncUsers 同步应用用户
func (ctrl *ApplicationPermissionController) SyncUsers(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// 获取操作人
	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := ctrl.service.SyncUsers(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationUsers 获取应用用户列表
func (ctrl *ApplicationPermissionController) GetApplicationUsers(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	users, err := ctrl.service.GetApplicationUsers(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}

// === 角色绑定 ===

// GetRoleBindings 获取角色绑定列表
func (ctrl *ApplicationPermissionController) GetGroupBindings(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	bindings, err := ctrl.service.GetGroupBindings(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(bindings))
}

// CreateGroupBinding 创建用户组绑定
func (ctrl *ApplicationPermissionController) CreateGroupBinding(c *gin.Context) {
	var binding models.GroupBinding
	if err := c.ShouldBindJSON(&binding); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	if err := ctrl.service.CreateGroupBinding(&binding); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(binding))
}

// DeleteGroupBinding 删除用户组绑定
func (ctrl *ApplicationPermissionController) DeleteGroupBinding(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ctrl.service.DeleteGroupBinding(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 用户组成员分配 ===

// GetUserGroups 获取用户所属用户组列表
func (ctrl *ApplicationPermissionController) GetUserGroups(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	groups, err := ctrl.service.GetUserGroups(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// AssignUserToGroup 为用户分配用户组成员
func (ctrl *ApplicationPermissionController) AssignUserToGroup(c *gin.Context) {
	var req struct {
		UserID  uint `json:"userId" binding:"required"`
		GroupID uint `json:"groupId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	// 获取操作人
	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	results, err := ctrl.service.AssignUserToGroup(req.UserID, req.GroupID, operator)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	// 返回授权结果，包含密码信息
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"message": "授权成功",
		"results": results,
	}))
}

// DeleteUserGroup 删除用户组成员
func (ctrl *ApplicationPermissionController) DeleteUserGroup(c *gin.Context) {
	userIDStr := c.Param("id")
	groupIDStr := c.Param("groupId")

	userID, _ := strconv.ParseUint(userIDStr, 10, 32)
	groupID, _ := strconv.ParseUint(groupIDStr, 10, 32)

	if err := ctrl.service.DeleteUserGroup(uint(userID), uint(groupID)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组成员失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 操作日志 ===

// GetOperationLogs 获取操作日志
func (ctrl *ApplicationPermissionController) GetOperationLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	currentStr := c.DefaultQuery("current", "1")
	sizeStr := c.DefaultQuery("size", "10")

	current, _ := strconv.Atoi(currentStr)
	size, _ := strconv.Atoi(sizeStr)

	logs, total, err := ctrl.service.GetOperationLogs(uint(id), current, size)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取操作日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"records": logs,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

// ========== 授权中心用户管理 ==========

// GetAuthUsers 获取授权中心用户列表
func (ctrl *ApplicationPermissionController) GetAuthUsers(c *gin.Context) {
	username := c.Query("username")
	currentStr := c.DefaultQuery("current", "1")
	sizeStr := c.DefaultQuery("size", "10")

	current, _ := strconv.Atoi(currentStr)
	size, _ := strconv.Atoi(sizeStr)

	users, total, err := ctrl.service.GetAuthUsers(current, size, username)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"records": users,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

// GetAuthUserPassword 获取用户初始密码
func (ctrl *ApplicationPermissionController) GetAuthUserPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	user, err := ctrl.service.GetAuthUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户失败"))
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusOK, utils.ErrorInternal("用户未设置初始密码"))
		return
	}

	// 返回明文密码
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"username": user.Username,
		"password": user.Password,
	}))
}

// CreateAuthUser 创建授权中心用户
func (ctrl *ApplicationPermissionController) CreateAuthUser(c *gin.Context) {
	var req models.AuthUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	password, err := ctrl.service.CreateAuthUser(&req)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户失败"))
		return
	}

	// 返回用户信息和初始密码
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"id":          req.ID,
		"username":    req.Username,
		"nickname":    req.Nickname,
		"email":       req.Email,
		"phone":       req.Phone,
		"description": req.Description,
		"status":      req.Status,
		"password":    password, // 返回明文密码，仅在创建时显示一次
		"message":     "用户创建成功，请妥善保管初始密码",
	}))
}

// UpdateAuthUser 更新授权中心用户
func (ctrl *ApplicationPermissionController) UpdateAuthUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req models.AuthUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	req.ID = uint(id)

	if err := ctrl.service.UpdateAuthUser(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// DeleteAuthUser 删除授权中心用户
func (ctrl *ApplicationPermissionController) DeleteAuthUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ctrl.service.DeleteAuthUser(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetAllAuthUsers 获取所有授权中心用户
func (ctrl *ApplicationPermissionController) GetAllAuthUsers(c *gin.Context) {
	users, err := ctrl.service.GetAllAuthUsers()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}

// ========== 授权中心角色管理 ==========

// GetAuthRoles 获取授权中心角色列表
func (ctrl *ApplicationPermissionController) GetAuthGroups(c *gin.Context) {
	name := c.Query("name")
	currentStr := c.DefaultQuery("current", "1")
	sizeStr := c.DefaultQuery("size", "10")

	current, _ := strconv.Atoi(currentStr)
	size, _ := strconv.Atoi(sizeStr)

	groups, total, err := ctrl.service.GetAuthGroups(current, size, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"records": groups,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

// CreateAuthGroup 创建授权中心用户组
func (ctrl *ApplicationPermissionController) CreateAuthGroup(c *gin.Context) {
	var req models.AuthGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	if err := ctrl.service.CreateAuthGroup(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// UpdateAuthGroup 更新授权中心用户组
func (ctrl *ApplicationPermissionController) UpdateAuthGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req models.AuthGroup
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("参数错误"))
		return
	}

	req.ID = uint(id)

	if err := ctrl.service.UpdateAuthGroup(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// DeleteAuthGroup 删除授权中心用户组
func (ctrl *ApplicationPermissionController) DeleteAuthGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := ctrl.service.DeleteAuthGroup(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetAllAuthGroups 获取所有授权中心用户组
func (ctrl *ApplicationPermissionController) GetAllAuthGroups(c *gin.Context) {
	groups, err := ctrl.service.GetAllAuthGroups()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// === 应用类型配置 ===

// GetSupportedAppTypes 获取支持的应用类型列表
func (ctrl *ApplicationPermissionController) GetSupportedAppTypes(c *gin.Context) {
	types := ctrl.service.GetSupportedAppTypes()
	c.JSON(http.StatusOK, utils.SuccessWithData(types))
}

// GetAppTypeConfigTemplate 获取应用类型的配置模板
func (ctrl *ApplicationPermissionController) GetAppTypeConfigTemplate(c *gin.Context) {
	appType := c.Param("type")
	if appType == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("应用类型参数缺失"))
		return
	}

	template, err := ctrl.service.GetAppTypeConfigTemplate(appType)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取配置模板失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(template))
}

// SyncApplicationGroups 同步应用用户组
// @Summary 同步应用用户组
// @Description 从外部应用同步用户组列表
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /applications/{id}/groups/sync [post]
func (ctrl *ApplicationPermissionController) SyncApplicationGroups(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的应用ID"))
		return
	}

	// 获取操作人
	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	// 同步用户组
	if err := ctrl.service.SyncGroups(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationGroups 获取应用用户组列表
// @Summary 获取应用用户组列表
// @Description 获取指定应用的用户组列表
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} Response{data=[]models.ApplicationGroup}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /applications/{id}/groups [get]
func (ctrl *ApplicationPermissionController) GetApplicationGroups(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的应用ID"))
		return
	}

	groups, err := ctrl.service.GetApplicationGroups(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// === 授权规则管理 ===

// SyncAuthorizationRules 同步应用授权规则
func (ctrl *ApplicationPermissionController) SyncAuthorizationRules(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	// 获取操作人
	operator := c.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := ctrl.service.SyncAuthorizationRules(uint(id), operator); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetApplicationAuthorizationRules 获取应用授权规则列表
func (ctrl *ApplicationPermissionController) GetApplicationAuthorizationRules(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的应用ID"))
		return
	}

	rules, err := ctrl.service.GetApplicationAuthorizationRules(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(rules))
}
