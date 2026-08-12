package k8s

import (
	"fmt"
	"net/http"
	"strconv"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/k8s"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// K8sPermissionController K8s权限控制器
type K8sPermissionController struct {
	svc *K8sClusterService
	db  *gorm.DB
}

// NewK8sPermissionController 创建K8s权限控制器
func NewK8sPermissionController(svc *K8sClusterService) *K8sPermissionController {
	return &K8sPermissionController{
		svc: svc,
		db:  database.GetDB(),
	}
}

// AssignClusterRoleRequest 分配集群角色请求
type AssignClusterRoleRequest struct {
	UserID uint `json:"userId" binding:"required"`
	RoleID uint `json:"roleId" binding:"required"`
}

// AssignClusterRole godoc
// @Summary      分配集群角色
// @Description  为指定用户在指定集群中分配角色
// @Tags         K8s-集群权限
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "集群 ID"
// @Param        body  body      AssignClusterRoleRequest  true  "分配集群角色请求"
// @Success      200  {object}  utils.Response  "分配角色成功"
// @Failure      200  {object}  utils.Response  "分配角色失败"
// @Router       /k8s/clusters/{id}/permissions [post]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) AssignClusterRole(c *gin.Context) {
	operatorID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req AssignClusterRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	if err := ctrl.svc.AssignClusterRole(req.UserID, uint(clusterID), req.RoleID, operatorID); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("分配角色失败: " + err.Error()))
		return
	}

	logger.Info("分配K8s集群角色",
		zap.Uint("operator_id", operatorID),
		zap.Uint("cluster_id", uint(clusterID)),
		zap.Uint("user_id", req.UserID),
		zap.Uint("role_id", req.RoleID))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("分配角色成功"))
}

// RevokeClusterRole godoc
// @Summary      撤销集群角色
// @Description  撤销指定用户在指定集群中的角色授权
// @Tags         K8s-集群权限
// @Produce      json
// @Param        id      path      int  true  "集群 ID"
// @Param        userId  path      int  true  "用户 ID"
// @Success      200  {object}  utils.Response  "撤销角色成功"
// @Failure      200  {object}  utils.Response  "撤销角色失败"
// @Router       /k8s/clusters/{id}/permissions/{userId} [delete]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) RevokeClusterRole(c *gin.Context) {
	operatorID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

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

	if err := ctrl.svc.RevokeClusterRole(uint(userID), uint(clusterID), operatorID); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("撤销角色失败: " + err.Error()))
		return
	}

	logger.Info("撤销K8s集群角色",
		zap.Uint("operator_id", operatorID),
		zap.Uint("cluster_id", uint(clusterID)),
		zap.Uint("user_id", uint(userID)))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("撤销角色成功"))
}

// GetUserClusters godoc
// @Summary      获取用户可访问的集群
// @Description  分页获取当前登录用户有权限访问的集群列表
// @Tags         K8s-集群权限
// @Produce      json
// @Param        page      query     int  false  "页码"    default(1)
// @Param        pageSize  query     int  false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "获取集群列表失败"
// @Router       /k8s/users/clusters [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetUserClusters(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	clusters, total, err := ctrl.svc.GetClusters(userID, page, pageSize, nil)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群列表失败: " + err.Error()))
		return
	}

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
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	hasAccess, err := ctrl.svc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	users, err := ctrl.svc.GetClusterUsers(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群用户列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}

// GetUserRoleInCluster godoc
// @Summary      获取用户在集群中的角色
// @Description  查询指定用户在指定集群中的角色绑定信息
// @Tags         K8s-集群权限
// @Produce      json
// @Param        id       path      int  true  "集群 ID"
// @Param        userId   path      int  true  "用户 ID"
// @Success      200  {object}  utils.Response{data=object{hasAccess=bool,role=object}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无效的用户ID / 获取失败"
// @Router       /k8s/clusters/{id}/users/{userId}/role [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetUserRoleInCluster(c *gin.Context) {
	_, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

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

	var binding modelk8s.ClusterRoleBinding
	err = ctrl.db.Where("user_id = ? AND cluster_id = ?", userID, clusterID).
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

// BatchAssignClusterRoles godoc
// @Summary      批量分配集群角色
// @Description  为多个用户批量分配同一集群的同一角色
// @Tags         K8s-集群权限
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "批量分配请求"  examples({\"clusterId\":1,\"userIds\":[1,2],\"roleId\":3})
// @Success      200   {object}  utils.Response{data=object{total=int,success=int,failed=int}}
// @Failure      200   {object}  utils.Response  "参数错误 / 分配失败"
// @Router       /k8s/permissions/batch-assign [post]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) BatchAssignClusterRoles(c *gin.Context) {
	operatorID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	var req struct {
		ClusterID uint    `json:"clusterId" binding:"required"`
		UserIDs   []uint  `json:"userIds" binding:"required"`
		RoleID    uint    `json:"roleId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	successCount := 0
	for _, userID := range req.UserIDs {
		if err := ctrl.svc.AssignClusterRole(userID, req.ClusterID, req.RoleID, operatorID); err != nil {
			logger.Warn("批量分配角色失败",
				zap.Uint("user_id", userID),
				zap.Uint("cluster_id", req.ClusterID),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	logger.Info("批量分配K8s集群角色",
		zap.Uint("operator_id", operatorID),
		zap.Uint("cluster_id", req.ClusterID),
		zap.Int("total", len(req.UserIDs)),
		zap.Int("success", successCount))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"data": gin.H{
			"total":   len(req.UserIDs),
			"success": successCount,
			"failed":  len(req.UserIDs) - successCount,
		},
		"message": fmt.Sprintf("批量分配完成，成功 %d/%d", successCount, len(req.UserIDs)),
	})
}
