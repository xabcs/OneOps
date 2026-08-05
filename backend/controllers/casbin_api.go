package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"oneops/backend/services"
	"oneops/backend/utils"
)

// CasbinAPIController 基于Casbin的API管理控制器
type CasbinAPIController struct{}

// NewCasbinAPIController 创建Casbin API控制器
func NewCasbinAPIController() *CasbinAPIController {
	return &CasbinAPIController{}
}

// GetAllAPIResources 获取所有API资源（从casbin_rule中提取）
func (ctrl *CasbinAPIController) GetAllAPIResources(c *gin.Context) {
	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	resources, err := manager.GetUniqueAPIResources()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取API资源失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(resources))
}

// GetAllPolicies 获取所有策略
func (ctrl *CasbinAPIController) GetAllPolicies(c *gin.Context) {
	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	policies, err := manager.GetAllPolicies()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取策略失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(policies))
}

// AddAPIResourceRequest 添加API资源请求
type AddAPIResourceRequest struct {
	RoleCode    string `json:"roleCode" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Module      string `json:"module"`
}

// AddAPIResource 添加API资源到casbin_rule
func (ctrl *CasbinAPIController) AddAPIResource(c *gin.Context) {
	var req AddAPIResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	if err := manager.AddAPIWithMetadata(req.RoleCode, req.Path, req.Method, req.Name, req.Description, req.Module); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("添加成功"))
}

// GetRolePermissions 获取角色的API权限
func (ctrl *CasbinAPIController) GetRolePermissions(c *gin.Context) {
	roleCode := c.Query("roleCode")
	if roleCode == "" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("角色代码不能为空"))
		return
	}

	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	permissions, err := manager.GetRolePermissions(roleCode)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色权限失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(permissions))
}

// CasbinAssignPermissionRequest 分配权限请求
type CasbinAssignPermissionRequest struct {
	RoleCode    string `json:"roleCode" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Module      string `json:"module"`
}

// AssignPermissionToRole 为角色分配权限
func (ctrl *CasbinAPIController) AssignPermissionToRole(c *gin.Context) {
	var req CasbinAssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	if err := manager.AssignPermissionToRole(req.RoleCode, req.Path, req.Method, req.Name, req.Description, req.Module); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("分配成功"))
}

// RevokePermissionRequest 撤销权限请求
type RevokePermissionRequest struct {
	RoleCode string `json:"roleCode" binding:"required"`
	Path     string `json:"path" binding:"required"`
	Method   string `json:"method" binding:"required"`
}

// RevokePermissionFromRole 撤销角色权限
func (ctrl *CasbinAPIController) RevokePermissionFromRole(c *gin.Context) {
	var req RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	if err := manager.RevokePermissionFromRole(req.RoleCode, req.Path, req.Method); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("撤销成功"))
}

// BatchAssignPermissionsRequest 批量分配权限请求
type BatchAssignPermissionsRequest struct {
	RoleCode string                          `json:"roleCode" binding:"required"`
	APIs     []CasbinAssignPermissionRequest `json:"apis" binding:"required"`
}

// BatchAssignPermissions 批量分配权限
func (ctrl *CasbinAPIController) BatchAssignPermissions(c *gin.Context) {
	var req BatchAssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	// 转换为APIResource格式
	apis := make([]services.APIResource, 0, len(req.APIs))
	for _, api := range req.APIs {
		apis = append(apis, services.APIResource{
			Path:        api.Path,
			Method:      api.Method,
			Name:        api.Name,
			Description: api.Description,
			Module:      api.Module,
		})
	}

	if err := manager.BatchAssignPermissions(req.RoleCode, apis); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("批量分配成功"))
}

// SyncCommonAPIs 同步常用API到casbin_rule
func (ctrl *CasbinAPIController) SyncCommonAPIs(c *gin.Context) {
	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	if err := manager.SyncCommonAPIs(); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("同步成功"))
}

// CheckPermission 检查权限
func (ctrl *CasbinAPIController) CheckPermission(c *gin.Context) {
	var req struct {
		RoleCode string `json:"roleCode" binding:"required"`
		Path     string `json:"path" binding:"required"`
		Method   string `json:"method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	manager, err := services.NewCasbinAPIManager()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("API管理器初始化失败"))
		return
	}

	allowed, err := manager.CheckPermission(req.RoleCode, req.Path, req.Method)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(map[string]bool{
		"allowed": allowed,
	}))
}