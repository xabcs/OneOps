package authorization

import (
	"net/http"
	"strconv"

	modelauth "oneops/backend3/model/authorization"
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

// GetApplications 获取应用列表
func (ctrl *ApplicationPermissionController) GetApplications(c *gin.Context) {
	name := c.Query("name")
	pagination := utils.ParsePagination(c)

	apps, total, err := ctrl.svc.GetApplications(pagination.Page, pagination.PageSize, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取应用列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(utils.BuildPaginatedResponse(apps, total, pagination)))
}

// CreateApplication 创建应用
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

// UpdateApplication 更新应用
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

// DeleteApplication 删除应用
func (ctrl *ApplicationPermissionController) DeleteApplication(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteApplication(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除应用失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 角色同步 ===

// SyncRoles 同步应用角色
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

// GetApplicationRoles 获取应用角色列表
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

// SyncUsers 同步应用用户
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

// GetApplicationUsers 获取应用用户列表
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

// GetGroupBindings 获取角色绑定列表
func (ctrl *ApplicationPermissionController) GetGroupBindings(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	bindings, err := ctrl.svc.GetGroupBindings(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(bindings))
}

// CreateGroupBinding 创建用户组绑定
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

// DeleteGroupBinding 删除用户组绑定
func (ctrl *ApplicationPermissionController) DeleteGroupBinding(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteGroupBinding(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组绑定失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// === 用户组成员分配 ===

// GetUserGroups 获取用户所属用户组列表
func (ctrl *ApplicationPermissionController) GetUserGroups(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	groups, err := ctrl.svc.GetUserGroups(uint(id))
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

// DeleteUserGroup 删除用户组成员
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

// GetOperationLogs 获取操作日志
func (ctrl *ApplicationPermissionController) GetOperationLogs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	pagination := utils.ParsePagination(c)

	logs, total, err := ctrl.svc.GetOperationLogs(uint(id), pagination.Page, pagination.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取操作日志失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(utils.BuildPaginatedResponse(logs, total, pagination)))
}

// ========== 授权中心用户管理 ==========

// GetAuthUsers 获取授权中心用户列表
func (ctrl *ApplicationPermissionController) GetAuthUsers(c *gin.Context) {
	username := c.Query("username")
	pagination := utils.ParsePagination(c)

	users, total, err := ctrl.svc.GetAuthUsers(pagination.Page, pagination.PageSize, username)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(utils.BuildPaginatedResponse(users, total, pagination)))
}

// GetAuthUserPassword 获取用户初始密码
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

// CreateAuthUser 创建授权中心用户
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

// UpdateAuthUser 更新授权中心用户
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

// DeleteAuthUser 删除授权中心用户
func (ctrl *ApplicationPermissionController) DeleteAuthUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteAuthUser(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetAllAuthUsers 获取所有授权中心用户
func (ctrl *ApplicationPermissionController) GetAllAuthUsers(c *gin.Context) {
	users, err := ctrl.svc.GetAllAuthUsers()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}

// ========== 授权中心角色管理 ==========

// GetAuthGroups 获取授权中心用户组列表
func (ctrl *ApplicationPermissionController) GetAuthGroups(c *gin.Context) {
	name := c.Query("name")
	pagination := utils.ParsePagination(c)

	groups, total, err := ctrl.svc.GetAuthGroups(pagination.Page, pagination.PageSize, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(utils.BuildPaginatedResponse(groups, total, pagination)))
}

// CreateAuthGroup 创建授权中心用户组
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

// UpdateAuthGroup 更新授权中心用户组
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

// DeleteAuthGroup 删除授权中心用户组
func (ctrl *ApplicationPermissionController) DeleteAuthGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := ctrl.svc.DeleteAuthGroup(uint(id)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户组失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nil))
}

// GetAllAuthGroups 获取所有授权中心用户组
func (ctrl *ApplicationPermissionController) GetAllAuthGroups(c *gin.Context) {
	groups, err := ctrl.svc.GetAllAuthGroups()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户组列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// === 应用类型配置 ===

// GetSupportedAppTypes 获取支持的应用类型列表
func (ctrl *ApplicationPermissionController) GetSupportedAppTypes(c *gin.Context) {
	types := ctrl.svc.GetSupportedAppTypes()
	c.JSON(http.StatusOK, utils.SuccessWithData(types))
}

// GetAppTypeConfigTemplate 获取应用类型的配置模板
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

// SyncApplicationGroups 同步应用用户组
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

// GetApplicationGroups 获取应用用户组列表
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

// SyncAuthorizationRules 同步应用授权规则
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

// GetApplicationAuthorizationRules 获取应用授权规则列表
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

// GetUserIdentityMappings 获取用户身份映射列表
func (ctrl *ApplicationPermissionController) GetUserIdentityMappings(c *gin.Context) {
	username := c.Query("username")
	appIDStr := c.Query("appId")
	status := c.Query("status")
	pagination := utils.ParsePagination(c)

	var appID uint
	if appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		appID = uint(id)
	}

	mappings, total, err := ctrl.svc.GetUserIdentityMappings(pagination.Page, pagination.PageSize, username, appID, status)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户身份映射列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(utils.BuildPaginatedResponse(mappings, total, pagination)))
}

// DeleteUserIdentityMapping 删除用户身份映射
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

// GetUserEffectivePermissions 获取用户有效权限列表
func (ctrl *ApplicationPermissionController) GetUserEffectivePermissions(c *gin.Context) {
	username := c.Query("username")
	appIDStr := c.Query("appId")
	pagination := utils.ParsePagination(c)

	var appID uint
	if appIDStr != "" {
		id, _ := strconv.ParseUint(appIDStr, 10, 32)
		appID = uint(id)
	}

	permissions, total, err := ctrl.svc.GetUserEffectivePermissions(pagination.Page, pagination.PageSize, username, appID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户有效权限列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(utils.BuildPaginatedResponse(permissions, total, pagination)))
}

// GetUserEffectivePermissionsMatrix 获取用户有效权限矩阵视图
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

// GetGroupBindingExecutions 获取权限绑定执行记录
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
