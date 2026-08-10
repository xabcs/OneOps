package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Ingresses ==========

// ListIngress 获取 Ingress 列表
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

	ingresses, total, err := ctrl.svc.ListIngress(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Ingress 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  ingresses,
		"total": total,
	}))
}

// GetIngress 获取 Ingress 详情
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

	ingress, err := ctrl.svc.GetIngress(uint(clusterID), namespace, name)
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

// CreateIngress 创建 Ingress
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

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.CreateIngress(uint(clusterID), req.Namespace, req.Manifest); err != nil {
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

// UpdateIngress 更新 Ingress
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

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.UpdateIngress(uint(clusterID), req.Namespace, req.Manifest); err != nil {
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

// DeleteIngress 删除 Ingress
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

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.DeleteIngress(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Ingress 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Ingress 成功"))
}
