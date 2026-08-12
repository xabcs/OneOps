package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Secrets ==========

// ListSecrets godoc
// @Summary      获取 Secret 列表
// @Description  分页获取指定集群下指定命名空间的 Secret 列表
// @Tags         K8s-配置
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  query     string  false  "命名空间"
// @Param        page       query     int     false  "页码"    default(1)
// @Param        pageSize   query     int     false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object{list=object,total=int}}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/secrets [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) ListSecrets(c *gin.Context) {
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

	secrets, total, err := ctrl.svc.ListSecrets(uint(clusterID), params.Namespace, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Secret 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  secrets,
		"total": total,
	}))
}

// GetSecret godoc
// @Summary      获取 Secret 详情
// @Description  按集群、命名空间、名称获取 Secret 详情
// @Tags         K8s-配置
// @Produce      json
// @Param        id         path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Param        name       path      string  true  "Secret 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/secrets/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sResourceController) GetSecret(c *gin.Context) {
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

	secret, err := ctrl.svc.GetSecret(uint(clusterID), namespace, name)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Secret 详情失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(secret))
}

// CreateSecretRequest 创建 Secret 请求
type CreateSecretRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// CreateSecret godoc
// @Summary      创建 Secret
// @Description  在指定集群的命名空间中创建 Secret
// @Tags         K8s-配置
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "集群 ID"
// @Param        body  body      CreateSecretRequest  true  "创建请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 创建失败"
// @Router       /k8s/clusters/{id}/secrets [post]
// @Security     BearerAuth
func (ctrl *K8sResourceController) CreateSecret(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req CreateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.CreateSecret(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建 Secret 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("创建 Secret 成功"))
}

// UpdateSecretRequest 更新 Secret 请求
type UpdateSecretRequest struct {
	Namespace string                 `json:"namespace" binding:"required"`
	Manifest  map[string]interface{} `json:"manifest" binding:"required"`
}

// UpdateSecret godoc
// @Summary      更新 Secret
// @Description  更新指定集群中的 Secret
// @Tags         K8s-配置
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "集群 ID"
// @Param        body  body      UpdateSecretRequest  true  "更新请求(含 manifest)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 更新失败"
// @Router       /k8s/clusters/{id}/secrets [put]
// @Security     BearerAuth
func (ctrl *K8sResourceController) UpdateSecret(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req UpdateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.UpdateSecret(uint(clusterID), req.Namespace, req.Manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新 Secret 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新 Secret 成功"))
}

// DeleteSecretRequest 删除 Secret 请求
type DeleteSecretRequest struct {
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// DeleteSecret godoc
// @Summary      删除 Secret
// @Description  删除指定集群中的 Secret
// @Tags         K8s-配置
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "集群 ID"
// @Param        body  body      DeleteSecretRequest  true  "删除请求(命名空间+名称)"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的集群ID / 无权访问 / 删除失败"
// @Router       /k8s/clusters/{id}/secrets [delete]
// @Security     BearerAuth
func (ctrl *K8sResourceController) DeleteSecret(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req DeleteSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	if err := ctrl.svc.DeleteSecret(uint(clusterID), req.Namespace, req.Name); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除 Secret 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除 Secret 成功"))
}
