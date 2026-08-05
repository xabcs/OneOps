package controllers

import (
	"net/http"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	permissionService *services.PermissionService
}

// NewPermissionController 创建权限控制器实例
func NewPermissionController() (*PermissionController, error) {
	permService, err := services.NewPermissionService()
	if err != nil {
		return nil, err
	}

	return &PermissionController{
		permissionService: permService,
	}, nil
}

// GetPermissionTree 获取权限树
// GET /api/v1/permissions/tree
func (c *PermissionController) GetPermissionTree(ctx *gin.Context) {
	tree, err := c.permissionService.GetPermissionTree()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("获取权限树失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(tree))
}

// GetPermissionList 获取权限列表（分页）
// GET /api/v1/permissions
func (c *PermissionController) GetPermissionList(ctx *gin.Context) {
	// 获取分页参数
	current, _ := strconv.Atoi(ctx.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	// 获取搜索参数
	name := ctx.Query("name")
	code := ctx.Query("code")
	module := ctx.Query("module")
	status := ctx.Query("status")

	// 构建查询条件
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

	// 获取权限列表
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
// POST /api/v1/permissions
func (c *PermissionController) CreatePermission(ctx *gin.Context) {
	var req models.CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误: " + err.Error()))
		return
	}

	permission := &models.Permission{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Module:      req.Module,
		Resource:    req.Resource,
		Action:      req.Action,
		Level:       req.Level,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		Status:      1, // 默认启用
	}

	if err := c.permissionService.CreatePermission(permission); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("创建权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permission))
}

// UpdatePermission 更新权限
// PUT /api/v1/permissions/:id
func (c *PermissionController) UpdatePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的权限ID"))
		return
	}

	var req models.UpdatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误: " + err.Error()))
		return
	}

	permission := &models.Permission{
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
// DELETE /api/v1/permissions/:id
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
// GET /api/v1/roles/:roleId/permissions
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

	// 提取权限ID数组
	permissionIDs := make([]uint, len(permissions))
	for i, perm := range permissions {
		permissionIDs[i] = perm.ID
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permissionIDs))
}

// AssignRolePermissions 分配角色权限
// POST /api/v1/roles/:roleId/permissions
// AssignRolePermissions 分配角色权限
// POST /api/v1/roles/:roleId/permissions
func (c *PermissionController) AssignRolePermissions(ctx *gin.Context) {
	// 从URL路径参数获取角色ID（符合项目规范）
	roleIDStr := ctx.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	// 从请求体绑定参数（符合项目规范）
	var req models.AssignRolePermissionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	// 调用服务层
	if err := c.permissionService.BatchAssignPermissions(uint(roleID), req.PermissionIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("分配权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("权限分配成功"))
}

// GetUserPermissions 获取当前用户权限
// GET /api/v1/user/permissions
func (c *PermissionController) GetUserPermissions(ctx *gin.Context) {
	// 从上下文获取用户ID（假设由认证中间件设置）
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
// POST /api/v1/permissions/check
func (c *PermissionController) CheckPermission(ctx *gin.Context) {
	var req models.CheckPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误"))
		return
	}

	allowed, err := c.permissionService.HasPermission(req.UserID, req.Permission)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("权限检查失败"))
		return
	}

	response := models.CheckPermissionResponse{
		Allowed: allowed,
	}

	if !allowed {
		response.Reason = "权限不足"
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(response))
}
