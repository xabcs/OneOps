package controllers

import (
	"net/http"
	"oneops/backend/container"
	"oneops/backend/logger"
	"oneops/backend/models"
	"oneops/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// K8sClusterController K8s集群控制器
type K8sClusterController struct {
	container *container.ServiceContainer
}

// NewK8sClusterController 创建K8s集群控制器
func NewK8sClusterController(cnt *container.ServiceContainer) *K8sClusterController {
	return &K8sClusterController{container: cnt}
}

// GetClusters 获取集群列表
func (ctrl *K8sClusterController) GetClusters(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	name := c.Query("name")
	status := c.Query("status")
	clusterType := c.Query("clusterType")

	// 构建过滤条件
	filter := &models.K8sClusterFilter{}
	if name != "" {
		filter.Name = &name
	}
	if status != "" {
		statusInt := 0
		if status == "1" {
			statusInt = 1
		}
		filter.Status = &statusInt
	}
	if clusterType != "" {
		filter.ClusterType = &clusterType
	}

	// 获取集群列表
	clusters, total, err := ctrl.container.K8sClusterService().GetClusters(userID.(uint), page, pageSize, filter)
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
		"code":    200,
		"success": true,
		"data":    result,
		"message": "success",
		"total":   total,
		"page":    page,
		"pageSize": pageSize,
	})
}

// GetClusterByID 获取集群详情
func (ctrl *K8sClusterController) GetClusterByID(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 获取集群详情
	cluster, err := ctrl.container.K8sClusterService().GetClusterByID(uint(clusterID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群详情失败: " + err.Error()))
		return
	}

	// 清理敏感信息
	result := map[string]interface{}{
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

	c.JSON(http.StatusOK, utils.SuccessWithData(result))
}

// CreateClusterRequest 创建集群请求
type CreateClusterRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	Endpoint      string `json:"endpoint" binding:"required"`
	Kubeconfig    string `json:"kubeconfig" binding:"required"`
	ClusterType   string `json:"clusterType"`
	Region        string `json:"region"`
	NodeCount     int    `json:"nodeCount"`
}

// CreateCluster 创建集群
func (ctrl *K8sClusterController) CreateCluster(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析请求
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 创建集群对象
	cluster := &models.K8sCluster{
		Name:        req.Name,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		ClusterType: req.ClusterType,
		Region:      req.Region,
		Version:     "", // 后续通过连接测试获取
		NodeCount:   req.NodeCount,
		Status:      1,
	}

	// 创建集群
	if err := ctrl.container.K8sClusterService().CreateCluster(cluster, req.Kubeconfig, userID.(uint)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建集群失败: " + err.Error()))
		return
	}

	logger.Info("创建K8s集群成功",
		zap.Uint("user_id", userID.(uint)),
		zap.Uint("cluster_id", cluster.ID),
		zap.String("cluster_name", cluster.Name))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建集群成功"))
}

// UpdateClusterRequest 更新集群请求
type UpdateClusterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Kubeconfig string `json:"kubeconfig"`
	ClusterType string `json:"clusterType"`
	Region      string `json:"region"`
	NodeCount   int    `json:"nodeCount"`
	Status      int    `json:"status"`
}

// UpdateCluster 更新集群
func (ctrl *K8sClusterController) UpdateCluster(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 解析请求
	var req UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Endpoint != "" {
		updates["endpoint"] = req.Endpoint
	}
	if req.Kubeconfig != "" {
		updates["kubeconfig"] = req.Kubeconfig
	}
	if req.ClusterType != "" {
		updates["cluster_type"] = req.ClusterType
	}
	if req.Region != "" {
		updates["region"] = req.Region
	}
	if req.NodeCount > 0 {
		updates["node_count"] = req.NodeCount
	}
	if req.Status >= 0 {
		updates["status"] = req.Status
	}

	// 更新集群
	if err := ctrl.container.K8sClusterService().UpdateCluster(uint(clusterID), updates, userID.(uint)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新集群失败: " + err.Error()))
		return
	}

	logger.Info("更新K8s集群成功",
		zap.Uint("user_id", userID.(uint)),
		zap.Uint("cluster_id", uint(clusterID)))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新集群成功"))
}

// DeleteClusterRequest 删除集群请求
type DeleteClusterRequest struct {
	ConfirmName string `json:"confirmName" binding:"required"`
}

// DeleteCluster 删除集群（需要二次确认）
func (ctrl *K8sClusterController) DeleteCluster(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 解析二次确认请求
	var req DeleteClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 删除集群（带二次确认）
	if err := ctrl.container.K8sClusterService().DeleteCluster(uint(clusterID), req.ConfirmName, userID.(uint)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除集群失败: " + err.Error()))
		return
	}

	logger.Warn("删除K8s集群成功",
		zap.Uint("user_id", userID.(uint)),
		zap.Uint("cluster_id", uint(clusterID)),
		zap.String("confirm_name", req.ConfirmName))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("集群已删除"))
}

// TestConnection 测试集群连接
func (ctrl *K8sClusterController) TestConnection(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 检查用户是否有权限访问该集群
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 测试连接
	if err := ctrl.container.K8sClusterService().TestConnection(uint(clusterID)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("连接测试失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("连接测试成功"))
}

// GetClusterNodes 获取集群节点列表
func (ctrl *K8sClusterController) GetClusterNodes(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
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

	// 获取节点列表
	nodes, err := ctrl.container.K8sClusterService().GetClusterNodes(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取节点列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nodes))
}

// GetClusterNamespaces 获取集群命名空间列表
func (ctrl *K8sClusterController) GetClusterNamespaces(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
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

	// 获取命名空间列表
	namespaces, err := ctrl.container.K8sClusterService().GetClusterNamespaces(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取命名空间列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(namespaces))
}

// GetClusterUsers 获取集群用户列表
func (ctrl *K8sClusterController) GetClusterUsers(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 解析集群ID
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
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
