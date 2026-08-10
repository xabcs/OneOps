package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Services ==========

// ListServices 获取 Service 列表
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

// GetService 获取 Service 详情
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

// CreateService 创建 Service
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

// UpdateService 更新 Service
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

// DeleteService 删除 Service
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
