package controllers

import (
	"net/http"
	"oneops/backend/services"
	"oneops/backend/utils"

	"github.com/gin-gonic/gin"
)

// APIPermissionController API权限管理控制器
type APIPermissionController struct {
	permService *services.PermissionService
}

// NewAPIPermissionController 创建API权限管理控制器
func NewAPIPermissionController() (*APIPermissionController, error) {
	permService, err := services.GetPermissionService()
	if err != nil {
		return nil, err
	}

	return &APIPermissionController{
		permService: permService,
	}, nil
}

// GetSystemAPIEndpoints 获取系统所有API端点定义
func (c *APIPermissionController) GetSystemAPIEndpoints(ctx *gin.Context) {
	endpoints := services.GetSystemAPIEndpoints()

	// 按分类组织
	categoryMap := make(map[string][]services.SystemAPIEndpoint)
	for _, endpoint := range endpoints {
		categoryMap[endpoint.Category] = append(categoryMap[endpoint.Category], endpoint)
	}

	// 转换为响应格式
	type CategoryResponse struct {
		Category  string                             `json:"category"`
		Endpoints []services.SystemAPIEndpoint       `json:"endpoints"`
	}

	response := make([]CategoryResponse, 0)
	categories := services.GetAllCategories()
	for _, category := range categories {
		response = append(response, CategoryResponse{
			Category:  category,
			Endpoints: categoryMap[category],
		})
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(response))
}

// GetRoleAPIPermissions 获取角色的API权限
func (c *APIPermissionController) GetRoleAPIPermissions(ctx *gin.Context) {
	roleCode := ctx.Param("roleCode")
	if roleCode == "" {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("角色代码不能为空"))
		return
	}

	permissions, err := c.permService.GetRoleAPIPermissions(roleCode)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取角色API权限失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permissions))
}

// AssignAPIPermission 为角色分配API权限
func (c *APIPermissionController) AssignAPIPermission(ctx *gin.Context) {
	var req struct {
		RoleCode  string `json:"roleCode" binding:"required"`
		APIPath   string `json:"apiPath" binding:"required"`
		HTTPMethod string `json:"httpMethod" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	if err := c.permService.AssignAPIPermission(req.RoleCode, req.APIPath, req.HTTPMethod); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("分配API权限失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("API权限分配成功"))
}

// RevokeAPIPermission 撤销角色的API权限
func (c *APIPermissionController) RevokeAPIPermission(ctx *gin.Context) {
	var req struct {
		RoleCode   string `json:"roleCode" binding:"required"`
		APIPath    string `json:"apiPath" binding:"required"`
		HTTPMethod string `json:"httpMethod" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	if err := c.permService.RevokeAPIPermission(req.RoleCode, req.APIPath, req.HTTPMethod); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("撤销API权限失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("API权限撤销成功"))
}

// BatchAssignAPIPermissions 批量为角色分配API权限
func (c *APIPermissionController) BatchAssignAPIPermissions(ctx *gin.Context) {
	var req struct {
		RoleCode    string `json:"roleCode" binding:"required"`
		Permissions []struct {
			Path   string `json:"path" binding:"required"`
			Method string `json:"method" binding:"required"`
		} `json:"permissions" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	// 转换格式
	permissions := make([]struct {
		Path   string
		Method string
	}, len(req.Permissions))
	for i, perm := range req.Permissions {
		permissions[i] = struct {
			Path   string
			Method string
		}{
			Path:   perm.Path,
			Method: perm.Method,
		}
	}

	if err := c.permService.BatchAssignAPIPermissions(req.RoleCode, permissions); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("批量分配API权限失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("批量分配API权限成功"))
}

// GetAPIPermissionStats 获取API权限统计信息
func (c *APIPermissionController) GetAPIPermissionStats(ctx *gin.Context) {
	allEndpoints := services.GetSystemAPIEndpoints()

	// 获取所有角色的权限
	policies := c.permService.GetAllPolicies()

	// 统计每个角色的权限数量
	roleStats := make(map[string]int)
	rolePermissions := make(map[string][]map[string]string) // roleCode -> []{path, method}

	for _, policy := range policies {
		if len(policy) >= 3 {
			roleCode := policy[0]
			apiPath := policy[1]
			method := policy[2]

			roleStats[roleCode]++
			rolePermissions[roleCode] = append(rolePermissions[roleCode], map[string]string{
				"path":   apiPath,
				"method": method,
			})
		}
	}

	type RoleStats struct {
		RoleCode    string   `json:"roleCode"`
		PermissionCount int  `json:"permissionCount"`
		HasAllAPIs  bool     `json:"hasAllAPIs"`
	}

	stats := make([]RoleStats, 0)
	for roleCode, count := range roleStats {
		stats = append(stats, RoleStats{
			RoleCode:    roleCode,
			PermissionCount: count,
			HasAllAPIs:  count == len(allEndpoints),
		})
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(map[string]interface{}{
		"totalAPIs":    len(allEndpoints),
		"roleStats":    stats,
		"totalPolicies": len(policies),
	}))
}

// CheckAPIPermission 检查当前用户是否有特定API权限
func (c *APIPermissionController) CheckAPIPermission(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"success": false,
			"message": "未授权访问",
			"data":    nil,
		})
		return
	}

	var req struct {
		APIPath    string `json:"apiPath" binding:"required"`
		HTTPMethod string `json:"httpMethod" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	hasPermission, err := c.permService.HasAPIPermission(userID.(uint), req.APIPath, req.HTTPMethod)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("检查API权限失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(map[string]bool{
		"hasPermission": hasPermission,
	}))
}
