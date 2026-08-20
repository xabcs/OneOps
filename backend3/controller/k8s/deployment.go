package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Workloads - Deployment ==========

// ListDeployments godoc
// @Summary      获取 Deployment 列表
// @Description  分页获取指定集群下指定命名空间的 Deployment 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int  true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/deployments [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListDeployments(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var params dto.K8sResourceQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	deployments, total, err := ctrl.svc.ForUser(userID).ListDeployments(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Deployment 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  deployments,
		"total": total,
	}))
}

// GetDeployment godoc
// @Summary      获取 Deployment 详情
// @Description  按集群、命名空间、名称获取 Deployment 详情
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Deployment 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/deployments/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetDeployment(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	deployment, err := ctrl.svc.ForUser(userID).GetDeployment(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Deployment 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(deployment))
}

// GetDeploymentPods godoc
// @Summary      获取 Deployment 的 Pods
// @Description  返回指定 Deployment 管理的 Pod 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Deployment 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/deployments/{namespace}/{name}/pods [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetDeploymentPods(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pods, err := ctrl.svc.ForUser(userID).GetDeploymentPods(uint(clusterID), namespace, name)
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

// CreateDeployment godoc
// @Summary      创建 Deployment
// @Description  在指定集群的命名空间中创建 Deployment
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "集群 ID"
// @Param        body  body      CreateDeploymentRequest  true  "创建请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 创建失败"
// @Router       /k8s/clusters/{id}/deployments [post]
// @Security     BearerAuth
func (ctrl *K8sResourceController) CreateDeployment(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.create")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.create）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).CreateDeployment(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Deployment 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Deployment 成功"))
}

// UpdateDeploymentRequest 更新 Deployment 请求
type UpdateDeploymentRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateDeployment godoc
// @Summary      更新 Deployment
// @Description  更新指定集群中的 Deployment
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "集群 ID"
// @Param        body  body      UpdateDeploymentRequest  true  "更新请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 更新失败"
// @Router       /k8s/clusters/{id}/deployments [put]
// @Security     BearerAuth
func (ctrl *K8sResourceController) UpdateDeployment(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.update")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.update）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).UpdateDeployment(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Deployment 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Deployment 成功"))
}

// DeleteDeploymentRequest 删除 Deployment 请求
type DeleteDeploymentRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteDeployment godoc
// @Summary      删除 Deployment
// @Description  删除指定集群中的 Deployment
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "集群 ID"
// @Param        body  body      DeleteDeploymentRequest  true  "删除请求(命名空间+名称)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 删除失败"
// @Router       /k8s/clusters/{id}/deployments [delete]
// @Security     BearerAuth
func (ctrl *K8sResourceController) DeleteDeployment(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.delete")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.delete）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).DeleteDeployment(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Deployment 失败: "+err.Error()))
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

// ScaleDeployment godoc
// @Summary      扩缩容 Deployment
// @Description  调整指定 Deployment 的副本数
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      ScaleDeploymentRequest  true  "扩缩容请求"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 扩缩容失败"
// @Router       /k8s/clusters/{id}/deployments/scale [post]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ScaleDeployment(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req ScaleDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.update")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.update）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).ScaleDeployment(uint(clusterID), req.Namespace, req.Name, req.Replicas); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("扩缩容 Deployment 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("扩缩容 Deployment 成功"))
}

// RestartDeploymentRequest 重启 Deployment 请求
type RestartDeploymentRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// RestartDeployment godoc
// @Summary      重启 Deployment
// @Description  重启指定 Deployment 的所有 Pod
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int                     true  "集群 ID"
// @Param        body  body      RestartDeploymentRequest  true  "重启请求"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 重启失败"
// @Router       /k8s/clusters/{id}/deployments/restart [post]
// @Security     BearerAuth
func (ctrl *K8sResourceController) RestartDeployment(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req RestartDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.update")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.update）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).RestartDeployment(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("重启 Deployment 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("重启 Deployment 成功"))
}
