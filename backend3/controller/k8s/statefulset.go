package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Workloads - StatefulSet ==========

// ListStatefulSets godoc
// @Summary      获取 StatefulSet 列表
// @Description  分页获取指定集群下指定命名空间的 StatefulSet 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/statefulsets [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListStatefulSets(c *gin.Context) {
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

	statefulSets, total, err := ctrl.svc.ListStatefulSets(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 StatefulSet 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  statefulSets,
		"total": total,
	}))
}

// GetStatefulSet godoc
// @Summary      获取 StatefulSet 详情
// @Description  按集群、命名空间、名称获取 StatefulSet 详情
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "StatefulSet 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/statefulsets/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetStatefulSet(c *gin.Context) {
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

	statefulSet, err := ctrl.svc.GetStatefulSet(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 StatefulSet 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(statefulSet))
}

// GetStatefulSetPods godoc
// @Summary      获取 StatefulSet 的 Pods
// @Description  返回指定 StatefulSet 管理的 Pod 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "StatefulSet 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/statefulsets/{namespace}/{name}/pods [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetStatefulSetPods(c *gin.Context) {
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

	pods, err := ctrl.svc.GetStatefulSetPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 StatefulSet Pods 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// ========== Workloads - DaemonSet ==========

// ListDaemonSets godoc
// @Summary      获取 DaemonSet 列表
// @Description  分页获取指定集群下指定命名空间的 DaemonSet 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/daemonsets [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListDaemonSets(c *gin.Context) {
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

	daemonSets, total, err := ctrl.svc.ListDaemonSets(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 DaemonSet 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  daemonSets,
		"total": total,
	}))
}

// GetDaemonSet godoc
// @Summary      获取 DaemonSet 详情
// @Description  按集群、命名空间、名称获取 DaemonSet 详情
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "DaemonSet 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/daemonsets/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetDaemonSet(c *gin.Context) {
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

	daemonSet, err := ctrl.svc.GetDaemonSet(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 DaemonSet 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(daemonSet))
}

// GetDaemonSetPods godoc
// @Summary      获取 DaemonSet 的 Pods
// @Description  返回指定 DaemonSet 管理的 Pod 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "DaemonSet 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/daemonsets/{namespace}/{name}/pods [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetDaemonSetPods(c *gin.Context) {
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

	pods, err := ctrl.svc.GetDaemonSetPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 DaemonSet Pods 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// GetJobPods godoc
// @Summary      获取 Job 的 Pods
// @Description  返回指定 Job 关联的 Pod 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Job 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/jobs/{namespace}/{name}/pods [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetJobPods(c *gin.Context) {
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

	pods, err := ctrl.svc.GetJobPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Job Pods 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}

// GetCronJobPods godoc
// @Summary      获取 CronJob 的 Pods
// @Description  返回指定 CronJob 关联的 Pod 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "CronJob 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/cronjobs/{namespace}/{name}/pods [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetCronJobPods(c *gin.Context) {
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

	pods, err := ctrl.svc.GetCronJobPods(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 CronJob Pods 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pods))
}
