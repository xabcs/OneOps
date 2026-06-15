package controllers

import (
	"fmt"
	"net/http"
	"oneops/backend/container"
	"oneops/backend/logger"
	"oneops/backend/models"
	"oneops/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// K8sPermissionController K8s权限控制器
type K8sPermissionController struct {
	container *container.ServiceContainer
}

// NewK8sPermissionController 创建K8s权限控制器
func NewK8sPermissionController(cnt *container.ServiceContainer) *K8sPermissionController {
	return &K8sPermissionController{container: cnt}
}

// AssignClusterRoleRequest 分配集群角色请求
type AssignClusterRoleRequest struct {
	UserID  uint `json:"userId" binding:"required"`
	RoleID  uint `json:"roleId" binding:"required"`
}

// AssignClusterRole 为用户分配集群角色
func (ctrl *K8sPermissionController) AssignClusterRole(c *gin.Context) {
	// 获取当前用户信息
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 解析请求
	var req AssignClusterRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 分配角色
	if err := ctrl.container.K8sClusterService().AssignClusterRole(req.UserID, uint(clusterID), req.RoleID, operatorID.(uint)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("分配角色失败: " + err.Error()))
		return
	}

	logger.Info("分配K8s集群角色",
		zap.Uint("operator_id", operatorID.(uint)),
		zap.Uint("cluster_id", uint(clusterID)),
		zap.Uint("user_id", req.UserID),
		zap.Uint("role_id", req.RoleID))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("分配角色成功"))
}

// RevokeClusterRole 撤销用户的集群角色
func (ctrl *K8sPermissionController) RevokeClusterRole(c *gin.Context) {
	// 获取当前用户信息
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID和用户ID
	clusterID, err := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户ID"))
		return
	}

	// 撤销角色
	if err := ctrl.container.K8sClusterService().RevokeClusterRole(uint(userID), uint(clusterID), operatorID.(uint)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("撤销角色失败: " + err.Error()))
		return
	}

	logger.Info("撤销K8s集群角色",
		zap.Uint("operator_id", operatorID.(uint)),
		zap.Uint("cluster_id", uint(clusterID)),
		zap.Uint("user_id", uint(userID)))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("撤销角色成功"))
}

// GetUserClusters 获取用户有权限的集群列表
func (ctrl *K8sPermissionController) GetUserClusters(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 获取集群列表（已包含权限过滤）
	clusters, total, err := ctrl.container.K8sClusterService().GetClusters(userID.(uint), page, pageSize, nil)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群列表失败: " + err.Error()))
		return
	}

	// 清理敏感信息
	result := make([]map[string]interface{}, len(clusters))
	for i, cluster := range clusters {
		result[i] = map[string]interface{}{
			"id":          cluster.ID,
			"name":        cluster.Name,
			"description": cluster.Description,
			"endpoint":    cluster.Endpoint,
			"clusterType": cluster.ClusterType,
			"region":      cluster.Region,
			"version":     cluster.Version,
			"nodeCount":   cluster.NodeCount,
			"status":      cluster.Status,
			"createdAt":   cluster.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":   cluster.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"success":  true,
		"data":     result,
		"message":  "success",
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetClusterUsers 获取集群用户列表（权限详情）
func (ctrl *K8sPermissionController) GetClusterUsers(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取用户列表
	users, err := ctrl.container.K8sClusterService().GetClusterUsers(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群用户列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}

// GetUserRoleInCluster 获取用户在集群中的角色
func (ctrl *K8sPermissionController) GetUserRoleInCluster(c *gin.Context) {
	// 获取当前用户信息
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析参数
	clusterID, err := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户ID"))
		return
	}

	// 获取用户在集群中的角色绑定
	var binding models.ClusterRoleBinding
	err = ctrl.container.GetDB().Where("user_id = ? AND cluster_id = ?", userID, clusterID).
		Preload("Role").
		First(&binding).Error

	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户角色失败: " + err.Error()))
		return
	}

	if binding.ID == 0 {
		c.JSON(http.StatusOK, utils.SuccessWithData(map[string]interface{}{
			"hasAccess": false,
			"role":      nil,
		}))
		return
	}

	result := map[string]interface{}{
		"hasAccess": true,
		"role": map[string]interface{}{
			"id":          binding.Role.ID,
			"name":        binding.Role.Name,
			"code":        binding.Role.Code,
			"description": binding.Role.Description,
		},
		"createdAt": binding.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(result))
}

// BatchAssignClusterRoles 批量分配集群角色
func (ctrl *K8sPermissionController) BatchAssignClusterRoles(c *gin.Context) {
	// 获取当前用户信息
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析请求
	var req struct {
		ClusterID uint     `json:"clusterId" binding:"required"`
		UserIDs   []uint  `json:"userIds" binding:"required"`
		RoleID    uint     `json:"roleId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 批量分配角色
	successCount := 0
	for _, userID := range req.UserIDs {
		if err := ctrl.container.K8sClusterService().AssignClusterRole(userID, req.ClusterID, req.RoleID, operatorID.(uint)); err != nil {
			logger.Warn("批量分配角色失败",
				zap.Uint("user_id", userID),
				zap.Uint("cluster_id", req.ClusterID),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	logger.Info("批量分配K8s集群角色",
		zap.Uint("operator_id", operatorID.(uint)),
		zap.Uint("cluster_id", req.ClusterID),
		zap.Int("total", len(req.UserIDs)),
		zap.Int("success", successCount))

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"success":  true,
		"data": gin.H{
			"total":   len(req.UserIDs),
			"success": successCount,
			"failed":  len(req.UserIDs) - successCount,
		},
		"message":  fmt.Sprintf("批量分配完成，成功 %d/%d", successCount, len(req.UserIDs)),
	})
}
