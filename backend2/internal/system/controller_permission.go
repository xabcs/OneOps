package system

import (
	"net/http"
	"oneops/backend2/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	permissionService *PermissionService
}

// NewPermissionController 创建权限控制器实例
func NewPermissionController() (*PermissionController, error) {
	permService, err := NewPermissionService()
	if err != nil {
		return nil, err
	}

	return &PermissionController{
		permissionService: permService,
	}, nil
}

// GetPermissionTree 获取权限树
func (c *PermissionController) GetPermissionTree(ctx *gin.Context) {
	tree, err := c.permissionService.GetPermissionTree()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("获取权限树失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(tree))
}

// GetPermissionList 获取权限列表（分页）
func (c *PermissionController) GetPermissionList(ctx *gin.Context) {
	current, _ := strconv.Atoi(ctx.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	name := ctx.Query("name")
	code := ctx.Query("code")
	module := ctx.Query("module")
	status := ctx.Query("status")

	query := map[string]interface{}{}
	if name != "" {
		query["name"] = name
	}
	if code != "" {
		query["code"] = code
	}
	if module != "" {
		query["module"] = module
	}
	if status != "" {
		query["status"] = status
	}

	permissions, total, err := c.permissionService.GetPermissionList(current, size, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("获取权限列表失败"))
		return
	}

	responseData := gin.H{
		"records": permissions,
		"total":   total,
		"current": current,
		"size":    size,
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(responseData))
}

// CreatePermission 创建权限
func (c *PermissionController) CreatePermission(ctx *gin.Context) {
	var req CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	permission := &Permission{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Module:      req.Module,
		Resource:    req.Resource,
		Action:      req.Action,
		Level:       req.Level,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		Status:      1,
	}

	if err := c.permissionService.CreatePermission(permission); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("创建权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permission))
}

// UpdatePermission 更新权限
func (c *PermissionController) UpdatePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的权限ID"))
		return
	}

	var req UpdatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	permission := &Permission{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Level:       req.Level,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}

	if err := c.permissionService.UpdatePermission(permission); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("更新权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeletePermission 删除权限
func (c *PermissionController) DeletePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的权限ID"))
		return
	}

	if err := c.permissionService.DeletePermission(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("删除权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// GetRolePermissions 获取角色权限
func (c *PermissionController) GetRolePermissions(ctx *gin.Context) {
	roleIDStr := ctx.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	permissions, err := c.permissionService.GetRolePermissions(uint(roleID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("获取角色权限失败"))
		return
	}

	permissionIDs := make([]uint, len(permissions))
	for i, perm := range permissions {
		permissionIDs[i] = perm.ID
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permissionIDs))
}

// AssignRolePermissions 分配角色权限
func (c *PermissionController) AssignRolePermissions(ctx *gin.Context) {
	roleIDStr := ctx.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	var req AssignRolePermissionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	if err := c.permissionService.BatchAssignPermissions(uint(roleID), req.PermissionIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("分配权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("权限分配成功"))
}

// GetUserPermissions 获取当前用户权限
func (c *PermissionController) GetUserPermissions(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.ErrorUnauthorized("未授权"))
		return
	}

	permissions, err := c.permissionService.GetUserPermissions(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("获取用户权限失败"))
		return
	}

	responseData := gin.H{
		"permissions": permissions,
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(responseData))
}

// CheckPermission 检查权限（内部使用）
func (c *PermissionController) CheckPermission(ctx *gin.Context) {
	var req CheckPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误"))
		return
	}

	allowed, err := c.permissionService.HasPermission(req.UserID, req.Permission)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("权限检查失败"))
		return
	}

	response := CheckPermissionResponse{
		Allowed: allowed,
	}

	if !allowed {
		response.Reason = "权限不足"
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(response))
}
