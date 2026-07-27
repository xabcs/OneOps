package controllers

import (
	"net/http"
	"oneops/backend/logger"
	"oneops/backend/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PermissionMappingController 权限映射控制器
type PermissionMappingController struct {
	permissionService *services.PermissionMappingService
}

// NewPermissionMappingController 创建权限映射控制器实例
func NewPermissionMappingController() *PermissionMappingController {
	return &PermissionMappingController{
		permissionService: services.NewPermissionMappingService(),
	}
}

// AssignPermissionRequest 权限分配请求
type AssignPermissionRequest struct {
	AppID            uint                   `json:"appId" binding:"required"`
	MappingType      string                 `json:"mappingType" binding:"required"`
	ExternalID       string                 `json:"externalId" binding:"required"`
	ExternalName     string                 `json:"externalName" binding:"required"`
	PermissionDetail map[string]interface{} `json:"permissionDetail" binding:"required"`
}

// AssignPermissionToAuthGroup 为授权中心用户组分配外部应用权限
// @Summary 权限分配
// @Description 为授权中心用户组分配外部应用权限
// @Tags 权限映射
// @Accept json
// @Produce json
// @Param authGroupId path int true "授权中心用户组ID"
// @Param request body AssignPermissionRequest true "权限分配请求"
// @Success 200 {object} Response{data=models.PermissionMappingResult}
// @Router /api/auth-groups/{authGroupId}/permissions [post]
func (c *PermissionMappingController) AssignPermissionToAuthGroup(ctx *gin.Context) {
	authGroupID, err := strconv.ParseUint(ctx.Param("authGroupId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户组ID",
		})
		return
	}

	var req AssignPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前用户（从JWT中）
	// currentUser := getCurrentUser(ctx)
	// grantedBy := currentUser.Username
	grantedBy := "admin" // 临时使用

	result, err := c.permissionService.AssignPermissionToAuthGroup(
		uint(authGroupID),
		req.AppID,
		req.MappingType,
		req.ExternalID,
		req.ExternalName,
		req.PermissionDetail,
		grantedBy,
	)

	if err != nil {
		logger.Error("分配权限失败",
			zap.Uint("authGroupId", uint(authGroupID)),
			zap.Uint("appId", req.AppID),
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "分配权限失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "操作成功",
		"data":    result,
	})
}

// GetAuthGroupPermissions 获取授权中心用户组的所有权限映射
// @Summary 获取用户组权限列表
// @Description 获取授权中心用户组的所有权限映射
// @Tags 权限映射
// @Accept json
// @Produce json
// @Param authGroupId path int true "授权中心用户组ID"
// @Success 200 {object} Response{data=[]models.AuthGroupEffectivePermission}
// @Router /api/auth-groups/{authGroupId}/permissions [get]
func (c *PermissionMappingController) GetAuthGroupPermissions(ctx *gin.Context) {
	authGroupID, err := strconv.ParseUint(ctx.Param("authGroupId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户组ID",
		})
		return
	}

	permissions, err := c.permissionService.GetAuthGroupPermissions(uint(authGroupID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限列表失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permissions,
	})
}

// GetUserEffectivePermissions 获取用户的有效权限
// @Summary 获取用户有效权限
// @Description 获取用户通过用户组继承的所有有效权限
// @Tags 权限映射
// @Accept json
// @Produce json
// @Param userId path int true "用户ID"
// @Success 200 {object} Response{data=[]models.AuthGroupEffectivePermission}
// @Router /api/users/{userId}/effective-permissions [get]
func (c *PermissionMappingController) GetUserEffectivePermissions(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("userId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	permissions, err := c.permissionService.GetUserEffectivePermissions(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户权限失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    permissions,
	})
}

// RevokePermissionFromAuthGroup 撤销授权中心用户组的外部应用权限
// @Summary 撤销权限
// @Description 撤销授权中心用户组的外部应用权限
// @Tags 权限映射
// @Accept json
// @Produce json
// @Param mappingId path int true "权限映射ID"
// @Success 200 {object} Response
// @Router /api/permission-mappings/{mappingId} [delete]
func (c *PermissionMappingController) RevokePermissionFromAuthGroup(ctx *gin.Context) {
	mappingID, err := strconv.ParseUint(ctx.Param("mappingId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的权限映射ID",
		})
		return
	}

	// 获取当前用户
	// currentUser := getCurrentUser(ctx)
	// revokedBy := currentUser.Username
	revokedBy := "admin" // 临时使用

	if err := c.permissionService.RevokePermissionFromAuthGroup(uint(mappingID), revokedBy); err != nil {
		logger.Error("撤销权限失败",
			zap.Uint("mappingId", uint(mappingID)),
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "撤销权限失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "撤销权限成功",
	})
}

// CreatePermissionTemplateRequest 创建权限模板请求
type CreatePermissionTemplateRequest struct {
	AppID           uint                   `json:"appId" binding:"required"`
	PermissionCode  string                 `json:"permissionCode" binding:"required"`
	PermissionName  string                 `json:"permissionName" binding:"required"`
	PermissionType  string                 `json:"permissionType" binding:"required"`
	Description     string                 `json:"description"`
	Template        map[string]interface{} `json:"template" binding:"required"`
}

// CreatePermissionTemplate 创建权限模板
// @Summary 创建权限模板
// @Description 为外部应用创建权限模板
// @Tags 权限模板
// @Accept json
// @Produce json
// @Param request body CreatePermissionTemplateRequest true "权限模板请求"
// @Success 200 {object} Response{data=models.ApplicationPermissionTemplate}
// @Router /api/permission-templates [post]
func (c *PermissionMappingController) CreatePermissionTemplate(ctx *gin.Context) {
	var req CreatePermissionTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	template, err := c.permissionService.CreatePermissionTemplate(
		req.AppID,
		req.PermissionCode,
		req.PermissionName,
		req.PermissionType,
		req.Description,
		req.Template,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建权限模板失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    template,
	})
}

// GetPermissionTemplates 获取应用的权限模板列表
// @Summary 获取权限模板列表
// @Description 获取外部应用的权限模板列表
// @Tags 权限模板
// @Accept json
// @Produce json
// @Param appId path int true "应用ID"
// @Success 200 {object} Response{data=[]models.ApplicationPermissionTemplate}
// @Router /api/applications/{appId}/permission-templates [get]
func (c *PermissionMappingController) GetPermissionTemplates(ctx *gin.Context) {
	appID, err := strconv.ParseUint(ctx.Param("appId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的应用ID",
		})
		return
	}

	templates, err := c.permissionService.GetPermissionTemplates(uint(appID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取权限模板失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    templates,
	})
}

// SyncAllPermissions 同步所有待处理的权限映射
// @Summary 同步权限映射
// @Description 同步所有待处理的权限映射到外部应用
// @Tags 权限映射
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /api/permission-mappings/sync [post]
func (c *PermissionMappingController) SyncAllPermissions(ctx *gin.Context) {
	if err := c.permissionService.SyncAllPendingPermissions(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "同步权限失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "同步成功",
	})
}

// RegisterPermissionMappingRoutes 注册权限映射相关路由
func RegisterPermissionMappingRoutes(router *gin.Engine) {
	controller := NewPermissionMappingController()

	// 权限映射路由
	permissionGroup := router.Group("/api")
	{
		// 用户组权限管理
		permissionGroup.POST("/auth-groups/:authGroupId/permissions", controller.AssignPermissionToAuthGroup)
		permissionGroup.GET("/auth-groups/:authGroupId/permissions", controller.GetAuthGroupPermissions)

		// 用户权限查询
		permissionGroup.GET("/users/:userId/effective-permissions", controller.GetUserEffectivePermissions)

		// 权限映射管理
		permissionGroup.DELETE("/permission-mappings/:mappingId", controller.RevokePermissionFromAuthGroup)
		permissionGroup.POST("/permission-mappings/sync", controller.SyncAllPermissions)

		// 权限模板管理
		permissionGroup.POST("/permission-templates", controller.CreatePermissionTemplate)
		permissionGroup.GET("/applications/:appId/permission-templates", controller.GetPermissionTemplates)
	}
}
