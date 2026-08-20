package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Workloads - CronJobs ==========

// ListCronJobs godoc
// @Summary      获取 CronJob 列表
// @Description  分页获取指定集群下指定命名空间的 CronJob 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/cronjobs [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListCronJobs(c *gin.Context) {
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

	cronJobs, total, err := ctrl.svc.ForUser(userID).ListCronJobs(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 CronJob 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  cronJobs,
		"total": total,
	}))
}

// GetCronJob godoc
// @Summary      获取 CronJob 详情
// @Description  按集群、命名空间、名称获取 CronJob 详情
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "CronJob 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/cronjobs/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetCronJob(c *gin.Context) {
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

	cronJob, err := ctrl.svc.ForUser(userID).GetCronJob(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 CronJob 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(cronJob))
}

// DeleteCronJobRequest 删除 CronJob 请求
type DeleteCronJobRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteCronJob godoc
// @Summary      删除 CronJob
// @Description  删除指定集群中的 CronJob
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "集群 ID"
// @Param        body  body      DeleteCronJobRequest  true  "删除请求(命名空间+名称)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 删除失败"
// @Router       /k8s/clusters/{id}/cronjobs [delete]
// @Security     BearerAuth
func (ctrl *K8sResourceController) DeleteCronJob(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.delete")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.delete）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).DeleteCronJob(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 CronJob 失败: "+err.Error()))
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

// SuspendCronJob godoc
// @Summary      暂停/恢复 CronJob
// @Description  暂停或恢复指定 CronJob 的调度
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      SuspendCronJobRequest  true  "暂停请求(suspend=true暂停,false恢复)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 操作失败"
// @Router       /k8s/clusters/{id}/cronjobs/suspend [put]
// @Security     BearerAuth
func (ctrl *K8sResourceController) SuspendCronJob(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req SuspendCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.update")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.update）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).SuspendCronJob(uint(clusterID), req.Namespace, req.Name, req.Suspend); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("暂停 CronJob 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("操作 CronJob 成功"))
}
