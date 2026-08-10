package system

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/utils"
	"oneops/backend3/service/system"
)

// PermissionController 权限控制器
type PermissionController struct {
	svc *system.PermissionService
}

// NewPermissionController 创建权限控制器实例
func NewPermissionController(svc *system.PermissionService) *PermissionController {
	return &PermissionController{svc: svc}
}

// GetPermissionTree 获取权限树
// GET /api/v1/permissions/tree
func (ctrl *PermissionController) GetPermissionTree(ctx *gin.Context) {
	tree, err := ctrl.svc.GetPermissionTree()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("获取权限树失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(tree))
}

// GetPermissionList 获取权限列表（分页）
// GET /api/v1/permissions
func (ctrl *PermissionController) GetPermissionList(ctx *gin.Context) {
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
	permissions, total, err := ctrl.svc.GetPermissionList(current, size, query)
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
func (ctrl *PermissionController) CreatePermission(ctx *gin.Context) {
	var req modelsystem.CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	permission := &modelsystem.Permission{
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

	if err := ctrl.svc.CreatePermission(permission); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("创建权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permission))
}

// UpdatePermission 更新权限
// PUT /api/v1/permissions/:id
func (ctrl *PermissionController) UpdatePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的权限ID"))
		return
	}

	var req modelsystem.UpdatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	permission := &modelsystem.Permission{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Level:       req.Level,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}

	if err := ctrl.svc.UpdatePermission(permission); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("更新权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeletePermission 删除权限
// DELETE /api/v1/permissions/:id
func (ctrl *PermissionController) DeletePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的权限ID"))
		return
	}

	if err := ctrl.svc.DeletePermission(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("删除权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// GetRolePermissions 获取角色权限
// GET /api/v1/roles/:roleId/permissions
func (ctrl *PermissionController) GetRolePermissions(ctx *gin.Context) {
	roleIDStr := ctx.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	permissions, err := ctrl.svc.GetRolePermissions(uint(roleID))
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
func (ctrl *PermissionController) AssignRolePermissions(ctx *gin.Context) {
	// 从URL路径参数获取角色ID（符合项目规范）
	roleIDStr := ctx.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	// 从请求体绑定参数（符合项目规范）
	var req modelsystem.AssignRolePermissionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	// 调用服务层
	if err := ctrl.svc.BatchAssignPermissions(uint(roleID), req.PermissionIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("分配权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("权限分配成功"))
}

// GetUserPermissions 获取当前用户权限
// GET /api/v1/user/permissions
func (ctrl *PermissionController) GetUserPermissions(ctx *gin.Context) {
	// 从上下文获取用户ID（假设由认证中间件设置）
	userID, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return
	}

	permissions, err := ctrl.svc.GetUserPermissions(userID)
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
func (ctrl *PermissionController) CheckPermission(ctx *gin.Context) {
	var req modelsystem.CheckPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ErrorBadRequest("参数错误"))
		return
	}

	allowed, err := ctrl.svc.HasPermission(req.UserID, req.Permission)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.ErrorInternal("权限检查失败"))
		return
	}

	response := modelsystem.CheckPermissionResponse{
		Allowed: allowed,
	}

	if !allowed {
		response.Reason = "权限不足"
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(response))
}
