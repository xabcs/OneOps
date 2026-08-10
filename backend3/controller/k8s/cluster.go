package k8s

import (
	"net/http"
	"strconv"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/k8s"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// K8sClusterController K8s集群控制器
type K8sClusterController struct {
	svc *K8sClusterService
}

// NewK8sClusterController 创建K8s集群控制器
func NewK8sClusterController(svc *K8sClusterService) *K8sClusterController {
	return &K8sClusterController{svc: svc}
}

// GetClusters 获取集群列表
func (ctrl *K8sClusterController) GetClusters(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	name := c.Query("name")
	status := c.Query("status")
	clusterType := c.Query("clusterType")

	filter := &modelk8s.K8sClusterFilter{}
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

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusters, total, err := ctrl.svc.GetClusters(userID, params.GetPage(), params.GetPageSize(), filter)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群列表失败: "+err.Error()))
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

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(result, total, params)))
}

// GetClusterByID 获取集群详情
func (ctrl *K8sClusterController) GetClusterByID(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	cluster, err := ctrl.svc.GetClusterByID(uint(clusterID), userID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群详情失败: "+err.Error()))
		return
	}

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
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint" binding:"required"`
	Kubeconfig  string `json:"kubeconfig" binding:"required"`
	ClusterType string `json:"clusterType"`
	Region      string `json:"region"`
	NodeCount   int    `json:"nodeCount"`
}

// CreateCluster 创建集群
func (ctrl *K8sClusterController) CreateCluster(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	cluster := &modelk8s.K8sCluster{
		Name:        req.Name,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		ClusterType: req.ClusterType,
		Region:      req.Region,
		Version:     "",
		NodeCount:   req.NodeCount,
		Status:      1,
	}

	if err := ctrl.svc.CreateCluster(cluster, req.Kubeconfig, userID); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建集群失败: "+err.Error()))
		return
	}

	logger.Info("创建K8s集群成功",
		zap.Uint("user_id", userID),
		zap.Uint("cluster_id", cluster.ID),
		zap.String("cluster_name", cluster.Name))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建集群成功"))
}

// UpdateClusterRequest 更新集群请求
type UpdateClusterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Kubeconfig  string `json:"kubeconfig"`
	ClusterType string `json:"clusterType"`
	Region      string `json:"region"`
	NodeCount   int    `json:"nodeCount"`
	Status      int    `json:"status"`
}

// UpdateCluster 更新集群
func (ctrl *K8sClusterController) UpdateCluster(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

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

	if err := ctrl.svc.UpdateCluster(uint(clusterID), updates, userID); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新集群失败: "+err.Error()))
		return
	}

	logger.Info("更新K8s集群成功",
		zap.Uint("user_id", userID),
		zap.Uint("cluster_id", uint(clusterID)))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新集群成功"))
}

// DeleteClusterRequest 删除集群请求
type DeleteClusterRequest struct {
	ConfirmName string `json:"confirmName" binding:"required"`
}

// DeleteCluster 删除集群（需要二次确认）
func (ctrl *K8sClusterController) DeleteCluster(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := ctrl.svc.DeleteCluster(uint(clusterID), req.ConfirmName, userID); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除集群失败: "+err.Error()))
		return
	}

	logger.Warn("删除K8s集群成功",
		zap.Uint("user_id", userID),
		zap.Uint("cluster_id", uint(clusterID)),
		zap.String("confirm_name", req.ConfirmName))

	c.JSON(http.StatusOK, utils.SuccessWithMessage("集群已删除"))
}

// TestConnection 测试集群连接
func (ctrl *K8sClusterController) TestConnection(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	hasAccess, err := ctrl.svc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.TestConnection(uint(clusterID)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("连接测试失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("连接测试成功"))
}

// GetClusterNodes 获取集群节点列表
func (ctrl *K8sClusterController) GetClusterNodes(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	hasAccess, err := ctrl.svc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	nodes, err := ctrl.svc.GetClusterNodes(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取节点列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(nodes))
}

// GetClusterNamespaces 获取集群命名空间列表
func (ctrl *K8sClusterController) GetClusterNamespaces(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	hasAccess, err := ctrl.svc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	namespaces, err := ctrl.svc.GetClusterNamespaces(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取命名空间列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(namespaces))
}

// GetClusterUsers 获取集群用户列表
func (ctrl *K8sClusterController) GetClusterUsers(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
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
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群用户列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(users))
}
