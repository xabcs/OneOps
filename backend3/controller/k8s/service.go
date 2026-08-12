package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Services ==========

// ListServices godoc
// @Summary      获取 Service 列表
// @Description  分页获取指定集群下指定命名空间的 Service 列表
// @Tags         K8s-网络
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/services [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListServices(c *gin.Context) {
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

	services, total, err := ctrl.svc.ListServices(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Service 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  services,
		"total": total,
	}))
}

// GetService godoc
// @Summary      获取 Service 详情
// @Description  按集群、命名空间、名称获取 Service 详情
// @Tags         K8s-网络
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Service 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/services/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetService(c *gin.Context) {
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

	service, err := ctrl.svc.GetService(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Service 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(service))
}

// CreateServiceRequest 创建 Service 请求
type CreateServiceRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateService godoc
// @Summary      创建 Service
// @Description  在指定集群的命名空间中创建 Service
// @Tags         K8s-网络
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      CreateServiceRequest  true  "创建请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 创建失败"
// @Router       /k8s/clusters/{id}/services [post]
// @Security     BearerAuth
func (ctrl *K8sResourceController) CreateService(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.CreateService(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Service 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Service 成功"))
}

// UpdateServiceRequest 更新 Service 请求
type UpdateServiceRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateService godoc
// @Summary      更新 Service
// @Description  更新指定集群中的 Service
// @Tags         K8s-网络
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      UpdateServiceRequest  true  "更新请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 更新失败"
// @Router       /k8s/clusters/{id}/services [put]
// @Security     BearerAuth
func (ctrl *K8sResourceController) UpdateService(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.UpdateService(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Service 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Service 成功"))
}

// DeleteServiceRequest 删除 Service 请求
type DeleteServiceRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteService godoc
// @Summary      删除 Service
// @Description  删除指定集群中的 Service
// @Tags         K8s-网络
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      DeleteServiceRequest  true  "删除请求(命名空间+名称)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 删除失败"
// @Router       /k8s/clusters/{id}/services [delete]
// @Security     BearerAuth
func (ctrl *K8sResourceController) DeleteService(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.DeleteService(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Service 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Service 成功"))
}
