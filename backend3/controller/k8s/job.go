package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Workloads - Jobs ==========

// ListJobs godoc
// @Summary      获取 Job 列表
// @Description  分页获取指定集群下指定命名空间的 Job 列表
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/jobs [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListJobs(c *gin.Context) {
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

	jobs, total, err := ctrl.svc.ListJobs(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Job 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  jobs,
		"total": total,
	}))
}

// GetJob godoc
// @Summary      获取 Job 详情
// @Description  按集群、命名空间、名称获取 Job 详情
// @Tags         K8s-工作负载
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Job 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/jobs/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetJob(c *gin.Context) {
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

	job, err := ctrl.svc.GetJob(uint(clusterID), namespace, name)
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

// DeleteJob godoc
// @Summary      删除 Job
// @Description  删除指定集群中的 Job
// @Tags         K8s-工作负载
// @Accept       json
// @Produce      json
// @Param        id    path      int               true  "集群 ID"
// @Param        body  body      DeleteJobRequest  true  "删除请求(命名空间+名称)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 删除失败"
// @Router       /k8s/clusters/{id}/jobs [delete]
// @Security     BearerAuth
func (ctrl *K8sResourceController) DeleteJob(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
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

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.DeleteJob(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Job 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Job 成功"))
}
