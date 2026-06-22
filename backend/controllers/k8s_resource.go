package controllers

import (
	"net/http"
	"oneops/backend/container"
	"oneops/backend/dto"
	"oneops/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// K8sResourceController K8s资源管理控制器
type K8sResourceController struct {
	container *container.ServiceContainer
}

// NewK8sResourceController 创建K8s资源管理控制器
func NewK8sResourceController(cnt *container.ServiceContainer) *K8sResourceController {
	return &K8sResourceController{container: cnt}
}

// ========== Workloads - Deployment ==========

// ListDeployments 获取 Deployment 列表
func (ctrl *K8sResourceController) ListDeployments(c *gin.Context) {
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

	// 绑定查询参数（复用项目标准方式）
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取分页数据
	deployments, total, err := ctrl.container.K8sResourceService().ListDeployments(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Deployment 列表失败: "+err.Error()))
		return
	}

	// 返回分页数据（使用项目标准格式）
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  deployments,
		"total": total,
	}))
}

// GetDeployment 获取 Deployment 详情
func (ctrl *K8sResourceController) GetDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	deployment, err := ctrl.container.K8sResourceService().GetDeployment(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Deployment 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(deployment))
}

// GetDeploymentPods 获取 Deployment 管理的 Pods
func (ctrl *K8sResourceController) GetDeploymentPods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pods, err := ctrl.container.K8sResourceService().GetDeploymentPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Deployment Pods 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// CreateDeploymentRequest 创建 Deployment 请求
type CreateDeploymentRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateDeployment 创建 Deployment
func (ctrl *K8sResourceController) CreateDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().CreateDeployment(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Deployment 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Deployment 成功"))
}

// UpdateDeploymentRequest 更新 Deployment 请求
type UpdateDeploymentRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateDeployment 更新 Deployment
func (ctrl *K8sResourceController) UpdateDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().UpdateDeployment(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Deployment 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Deployment 成功"))
}

// DeleteDeploymentRequest 删除 Deployment 请求
type DeleteDeploymentRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteDeployment 删除 Deployment（需要二次确认）
func (ctrl *K8sResourceController) DeleteDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeleteDeployment(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Deployment 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Deployment 成功"))
}

// ScaleDeploymentRequest 扩缩容 Deployment 请求
type ScaleDeploymentRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Replicas  int32  `json:"replicas" binding:"required,min=0"`
}

// ScaleDeployment 扩缩容 Deployment
func (ctrl *K8sResourceController) ScaleDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req ScaleDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().ScaleDeployment(uint(clusterID), req.Namespace, req.Name, req.Replicas); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("扩缩容 Deployment 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("扩缩容 Deployment 成功"))
}

// RestartDeploymentRequest 重启 Deployment 请求
type RestartDeploymentRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// RestartDeployment 重启 Deployment
func (ctrl *K8sResourceController) RestartDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req RestartDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().RestartDeployment(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("重启 Deployment 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("重启 Deployment 成功"))
}

// ========== Workloads - StatefulSet ==========

// ListStatefulSets 获取 StatefulSet 列表
func (ctrl *K8sResourceController) ListStatefulSets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	statefulSets, total, err := ctrl.container.K8sResourceService().ListStatefulSets(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 StatefulSet 列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  statefulSets,
		"total": total,
	}))
}

// GetStatefulSet 获取 StatefulSet 详情
func (ctrl *K8sResourceController) GetStatefulSet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	statefulSet, err := ctrl.container.K8sResourceService().GetStatefulSet(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 StatefulSet 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(statefulSet))
}

// GetStatefulSetPods 获取 StatefulSet 管理的 Pods
func (ctrl *K8sResourceController) GetStatefulSetPods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pods, err := ctrl.container.K8sResourceService().GetStatefulSetPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 StatefulSet Pods 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// ========== Workloads - DaemonSet ==========

// ListDaemonSets 获取 DaemonSet 列表
func (ctrl *K8sResourceController) ListDaemonSets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	daemonSets, total, err := ctrl.container.K8sResourceService().ListDaemonSets(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 DaemonSet 列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  daemonSets,
		"total": total,
	}))
}

// GetDaemonSet 获取 DaemonSet 详情
func (ctrl *K8sResourceController) GetDaemonSet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	daemonSet, err := ctrl.container.K8sResourceService().GetDaemonSet(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 DaemonSet 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(daemonSet))
}

// GetDaemonSetPods 获取 DaemonSet 管理的 Pods
func (ctrl *K8sResourceController) GetDaemonSetPods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pods, err := ctrl.container.K8sResourceService().GetDaemonSetPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 DaemonSet Pods 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// GetJobPods 获取 Job 关联的 Pods
func (ctrl *K8sResourceController) GetJobPods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pods, err := ctrl.container.K8sResourceService().GetJobPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Job Pods 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// GetCronJobPods 获取 CronJob 关联的 Pods
func (ctrl *K8sResourceController) GetCronJobPods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pods, err := ctrl.container.K8sResourceService().GetCronJobPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 CronJob Pods 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// ========== Services ==========

// ListServices 获取 Service 列表
func (ctrl *K8sResourceController) ListServices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数（复用项目标准方式）
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取分页数据
	services, total, err := ctrl.container.K8sResourceService().ListServices(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Service 列表失败: "+err.Error()))
		return
	}

	// 返回分页数据（使用项目标准格式）
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  services,
		"total": total,
	}))
}

// GetService 获取 Service 详情
func (ctrl *K8sResourceController) GetService(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	service, err := ctrl.container.K8sResourceService().GetService(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Service 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(service))
}

// CreateServiceRequest 创建 Service 请求
type CreateServiceRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateService 创建 Service
func (ctrl *K8sResourceController) CreateService(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().CreateService(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Service 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Service 成功"))
}

// UpdateServiceRequest 更新 Service 请求
type UpdateServiceRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateService 更新 Service
func (ctrl *K8sResourceController) UpdateService(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().UpdateService(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Service 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Service 成功"))
}

// DeleteServiceRequest 删除 Service 请求
type DeleteServiceRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteService 删除 Service（需要二次确认）
func (ctrl *K8sResourceController) DeleteService(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeleteService(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Service 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Service 成功"))
}

// ========== Ingresses ==========

// ListIngress 获取 Ingress 列表
func (ctrl *K8sResourceController) ListIngress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取分页数据
	ingresses, total, err := ctrl.container.K8sResourceService().ListIngress(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Ingress 列表失败: "+err.Error()))
		return
	}

	// 返回分页数据
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  ingresses,
		"total": total,
	}))
}

// GetIngress 获取 Ingress 详情
func (ctrl *K8sResourceController) GetIngress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	ingress, err := ctrl.container.K8sResourceService().GetIngress(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Ingress 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(ingress))
}

// CreateIngressRequest 创建 Ingress 请求
type CreateIngressRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateIngress 创建 Ingress
func (ctrl *K8sResourceController) CreateIngress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateIngressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().CreateIngress(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Ingress 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Ingress 成功"))
}

// UpdateIngressRequest 更新 Ingress 请求
type UpdateIngressRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateIngress 更新 Ingress
func (ctrl *K8sResourceController) UpdateIngress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateIngressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().UpdateIngress(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Ingress 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Ingress 成功"))
}

// DeleteIngressRequest 删除 Ingress 请求
type DeleteIngressRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteIngress 删除 Ingress
func (ctrl *K8sResourceController) DeleteIngress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteIngressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeleteIngress(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Ingress 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Ingress 成功"))
}

// ========== Pods ==========

// ListPods 获取 Pod 列表
func (ctrl *K8sResourceController) ListPods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数（复用项目标准方式）
	var params dto.K8sPodQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取分页数据
	pods, total, err := ctrl.container.K8sResourceService().ListPods(uint(clusterID), params.Namespace, params.LabelSelector, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Pod 列表失败: "+err.Error()))
		return
	}

	// 返回分页数据（使用项目标准格式）
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  pods,
		"total": total,
	}))
}

// GetPod 获取 Pod 详情
func (ctrl *K8sResourceController) GetPod(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pod, err := ctrl.container.K8sResourceService().GetPod(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Pod 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pod))
}

// GetPodLogs 获取 Pod 日志
func (ctrl *K8sResourceController) GetPodLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")
	container := c.DefaultQuery("container", "")

	// 解析行数
	tailLines := int64(100)
	if lines := c.Query("tailLines"); lines != "" {
		if l, err := strconv.ParseInt(lines, 10, 64); err == nil {
			tailLines = l
		}
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	logs, err := ctrl.container.K8sResourceService().GetPodLogs(uint(clusterID), namespace, name, container, tailLines)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Pod 日志失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(map[string]interface{}{
		"logs": logs,
	}))
}

// UpdatePodRequest 更新 Pod 请求
type UpdatePodRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdatePod 更新 Pod
func (ctrl *K8sResourceController) UpdatePod(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdatePodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().UpdatePod(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Pod 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Pod 成功"))
}

// DeletePodRequest 删除 Pod 请求
type DeletePodRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeletePod 删除 Pod
func (ctrl *K8sResourceController) DeletePod(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeletePodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeletePod(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Pod 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Pod 成功"))
}

// ========== Workloads - Jobs ==========

// ListJobs 获取 Job 列表
func (ctrl *K8sResourceController) ListJobs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数（复用项目标准方式）
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取分页数据
	jobs, total, err := ctrl.container.K8sResourceService().ListJobs(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Job 列表失败: "+err.Error()))
		return
	}

	// 返回分页数据（使用项目标准格式）
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  jobs,
		"total": total,
	}))
}

// GetJob 获取 Job 详情
func (ctrl *K8sResourceController) GetJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	job, err := ctrl.container.K8sResourceService().GetJob(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Job 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(job))
}

// DeleteJobRequest 删除 Job 请求
type DeleteJobRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteJob 删除 Job（需要二次确认）
func (ctrl *K8sResourceController) DeleteJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeleteJob(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Job 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Job 成功"))
}

// ========== Workloads - CronJobs ==========

// ListCronJobs 获取 CronJob 列表
func (ctrl *K8sResourceController) ListCronJobs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	cronJobs, total, err := ctrl.container.K8sResourceService().ListCronJobs(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 CronJob 列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  cronJobs,
		"total": total,
	}))
}

// GetCronJob 获取 CronJob 详情
func (ctrl *K8sResourceController) GetCronJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	cronJob, err := ctrl.container.K8sResourceService().GetCronJob(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 CronJob 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(cronJob))
}

// DeleteCronJobRequest 删除 CronJob 请求
type DeleteCronJobRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteCronJob 删除 CronJob
func (ctrl *K8sResourceController) DeleteCronJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeleteCronJob(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 CronJob 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 CronJob 成功"))
}

// SuspendCronJobRequest 暂停 CronJob 请求
type SuspendCronJobRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Suspend   bool   `json:"suspend"`
}

// SuspendCronJob 暂停/恢复 CronJob
func (ctrl *K8sResourceController) SuspendCronJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req SuspendCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().SuspendCronJob(uint(clusterID), req.Namespace, req.Name, req.Suspend); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("暂停 CronJob 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("操作 CronJob 成功"))
}

// ========== ConfigMaps ==========

// ListConfigMaps 获取 ConfigMap 列表
func (ctrl *K8sResourceController) ListConfigMaps(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数（复用项目标准方式）
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取分页数据
	configMaps, total, err := ctrl.container.K8sResourceService().ListConfigMaps(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 ConfigMap 列表失败: "+err.Error()))
		return
	}

	// 返回分页数据（使用项目标准格式）
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  configMaps,
		"total": total,
	}))
}

// GetConfigMap 获取 ConfigMap 详情
func (ctrl *K8sResourceController) GetConfigMap(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	configMap, err := ctrl.container.K8sResourceService().GetConfigMap(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 ConfigMap 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(configMap))
}

// CreateConfigMapRequest 创建 ConfigMap 请求
type CreateConfigMapRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateConfigMap 创建 ConfigMap
func (ctrl *K8sResourceController) CreateConfigMap(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateConfigMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().CreateConfigMap(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 ConfigMap 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 ConfigMap 成功"))
}

// UpdateConfigMapRequest 更新 ConfigMap 请求
type UpdateConfigMapRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateConfigMap 更新 ConfigMap
func (ctrl *K8sResourceController) UpdateConfigMap(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateConfigMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().UpdateConfigMap(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 ConfigMap 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 ConfigMap 成功"))
}

// DeleteConfigMapRequest 删除 ConfigMap 请求
type DeleteConfigMapRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteConfigMap 删除 ConfigMap（需要二次确认）
func (ctrl *K8sResourceController) DeleteConfigMap(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteConfigMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeleteConfigMap(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 ConfigMap 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 ConfigMap 成功"))
}

// ========== Secrets ==========

// ListSecrets 获取 Secret 列表
func (ctrl *K8sResourceController) ListSecrets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	// 绑定查询参数（复用项目标准方式）
	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	// 获取分页数据
	secrets, total, err := ctrl.container.K8sResourceService().ListSecrets(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Secret 列表失败: "+err.Error()))
		return
	}

	// 返回分页数据（使用项目标准格式）
	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  secrets,
		"total": total,
	}))
}

// GetSecret 获取 Secret 详情
func (ctrl *K8sResourceController) GetSecret(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	secret, err := ctrl.container.K8sResourceService().GetSecret(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Secret 详情失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(secret))
}

// CreateSecretRequest 创建 Secret 请求
type CreateSecretRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateSecret 创建 Secret
func (ctrl *K8sResourceController) CreateSecret(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().CreateSecret(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Secret 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Secret 成功"))
}

// UpdateSecretRequest 更新 Secret 请求
type UpdateSecretRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateSecret 更新 Secret
func (ctrl *K8sResourceController) UpdateSecret(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().UpdateSecret(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Secret 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Secret 成功"))
}

// DeleteSecretRequest 删除 Secret 请求
type DeleteSecretRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteSecret 删除 Secret（需要二次确认）
func (ctrl *K8sResourceController) DeleteSecret(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: " + err.Error()))
		return
	}

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.container.K8sResourceService().DeleteSecret(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Secret 失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Secret 成功"))
}

// ========== Events ==========

// ListEvents 获取 Event 列表
func (ctrl *K8sResourceController) ListEvents(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.DefaultQuery("namespace", "default")
	fieldSelector := c.Query("fieldSelector")

	// 检查权限
	hasAccess, err := ctrl.container.K8sClusterService().CheckUserClusterAccess(userID.(uint), uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	events, err := ctrl.container.K8sResourceService().ListEvents(uint(clusterID), namespace, fieldSelector)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Event 列表失败: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(events))
}
