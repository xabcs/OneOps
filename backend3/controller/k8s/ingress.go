package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Ingresses ==========

// ListIngress godoc
// @Summary      获取 Ingress 列表
// @Description  分页获取指定集群下指定命名空间的 Ingress 列表
// @Tags         K8s-网络
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/ingresses [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListIngress(c *gin.Context) {
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

	ingresses, total, err := ctrl.svc.ForUser(userID).ListIngress(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Ingress 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  ingresses,
		"total": total,
	}))
}

// GetIngress godoc
// @Summary      获取 Ingress 详情
// @Description  按集群、命名空间、名称获取 Ingress 详情
// @Tags         K8s-网络
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Ingress 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/ingresses/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetIngress(c *gin.Context) {
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

	ingress, err := ctrl.svc.ForUser(userID).GetIngress(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Ingress 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(ingress))
}

// CreateIngressRequest 创建 Ingress 请求
type CreateIngressRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateIngress godoc
// @Summary      创建 Ingress
// @Description  在指定集群的命名空间中创建 Ingress
// @Tags         K8s-网络
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      CreateIngressRequest  true  "创建请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 创建失败"
// @Router       /k8s/clusters/{id}/ingresses [post]
// @Security     BearerAuth
func (ctrl *K8sResourceController) CreateIngress(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateIngressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.create")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.create）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).CreateIngress(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Ingress 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Ingress 成功"))
}

// UpdateIngressRequest 更新 Ingress 请求
type UpdateIngressRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateIngress godoc
// @Summary      更新 Ingress
// @Description  更新指定集群中的 Ingress
// @Tags         K8s-网络
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      UpdateIngressRequest  true  "更新请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 更新失败"
// @Router       /k8s/clusters/{id}/ingresses [put]
// @Security     BearerAuth
func (ctrl *K8sResourceController) UpdateIngress(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateIngressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.update")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.update）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).UpdateIngress(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Ingress 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Ingress 成功"))
}

// DeleteIngressRequest 删除 Ingress 请求
type DeleteIngressRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteIngress godoc
// @Summary      删除 Ingress
// @Description  删除指定集群中的 Ingress
// @Tags         K8s-网络
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "集群 ID"
// @Param        body  body      DeleteIngressRequest  true  "删除请求(命名空间+名称)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 删除失败"
// @Router       /k8s/clusters/{id}/ingresses [delete]
// @Security     BearerAuth
func (ctrl *K8sResourceController) DeleteIngress(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteIngressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.resource.delete")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.resource.delete）"))
		return
	}

	if err := ctrl.svc.ForUser(userID).DeleteIngress(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Ingress 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Ingress 成功"))
}
