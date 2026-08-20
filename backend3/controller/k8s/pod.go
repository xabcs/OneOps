package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Pods ==========

// ListPods godoc
// @Summary      获取 Pod 列表
// @Description  分页获取指定集群下指定命名空间的 Pod 列表，支持标签选择器
// @Tags         K8s-Pod
// @Produce      json
// @Param        id              path      int     true  "集群 ID"
// @Param        namespace       query     string  false  "命名空间"
// @Param        labelSelector   query     string  false  "标签选择器"
// @Param        page            query     int     false  "页码"    default(1)
// @Param        pageSize        query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/pods [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListPods(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var params dto.K8sPodQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	pods, total, err := ctrl.svc.ForUser(userID).ListPods(uint(clusterID), params.Namespace, params.LabelSelector, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Pod 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  pods,
		"total": total,
	}))
}

// GetPod godoc
// @Summary      获取 Pod 详情
// @Description  按集群、命名空间、名称获取 Pod 详情
// @Tags         K8s-Pod
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Pod 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/pods/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetPod(c *gin.Context) {
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

	pod, err := ctrl.svc.ForUser(userID).GetPod(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Pod 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(pod))
}

// GetPodLogs godoc
// @Summary      获取 Pod 日志
// @Description  获取指定 Pod 的容器日志（一次性返回，非流式）
// @Tags         K8s-Pod
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Pod 名称"
// @Param        container  query     string  false  "容器名称"
// @Param        tailLines  query     int     false  "尾部行数" default(100)
// @Success      200  {object}  utils.Response{data=object{logs=string}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/pods/{namespace}/{name}/logs [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetPodLogs(c *gin.Context) {
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
	container := c.DefaultQuery("container", "")

	tailLines := int64(100)
	if lines := c.Query("tailLines"); lines != "" {
		if l, err := strconv.ParseInt(lines, 10, 64); err == nil {
			tailLines = l
		}
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	logs, err := ctrl.svc.ForUser(userID).GetPodLogs(uint(clusterID), namespace, name, container, tailLines)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Pod 日志失败: "+err.Error()))
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

// UpdatePod godoc
// @Summary      更新 Pod
// @Description  更新指定集群中的 Pod
// @Tags         K8s-Pod
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "集群 ID"
// @Param        body  body      UpdatePodRequest  true  "更新请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 更新失败"
// @Router       /k8s/clusters/{id}/pods [put]
// @Security     BearerAuth
func (ctrl *K8sResourceController) UpdatePod(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdatePodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.update")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.update）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).UpdatePod(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Pod 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Pod 成功"))
}

// DeletePodRequest 删除 Pod 请求
type DeletePodRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeletePod godoc
// @Summary      删除 Pod
// @Description  删除指定集群中的 Pod
// @Tags         K8s-Pod
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "集群 ID"
// @Param        body  body      DeletePodRequest  true  "删除请求(命名空间+名称)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 删除失败"
// @Router       /k8s/clusters/{id}/pods [delete]
// @Security     BearerAuth
func (ctrl *K8sResourceController) DeletePod(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeletePodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.delete")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.delete）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).DeletePod(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Pod 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Pod 成功"))
}
