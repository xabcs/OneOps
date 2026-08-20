package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/utils"
	k8ssvc "oneops/backend3/service/k8s"

	"github.com/gin-gonic/gin"
)

// K8sRbacController K8s 原生 RBAC 对象代管（A 模式）
type K8sRbacController struct {
	svc *k8ssvc.K8sRbacService
}

// NewK8sRbacController 创建 RBAC 代管控制器
func NewK8sRbacController(svc *k8ssvc.K8sRbacService) *K8sRbacController {
	return &K8sRbacController{svc: svc}
}

func clusterIDFromPath(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return 0, false
	}
	return uint(id), true
}

// ListClusterRoles godoc
// @Summary      获取 ClusterRole 列表
// @Description  列出集群内原生 RBAC ClusterRole（精简视图，含 Helm 管理标记）
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id      path   int    true  "集群 ID"
// @Param        search  query  string false "按名称搜索"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/clusterroles [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) ListClusterRoles(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	list, err := ctrl.svc.ListClusterRoles(clusterID, c.Query("search"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(list))
}

// GetClusterRole godoc
// @Summary      获取 ClusterRole 详情
// @Description  获取集群内原生 ClusterRole 的完整 rules
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id    path   int    true  "集群 ID"
// @Param        name  path   string true  "ClusterRole 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/clusterroles/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) GetClusterRole(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	role, err := ctrl.svc.GetClusterRole(clusterID, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(role))
}

// UpdateClusterRole godoc
// @Summary      更新 ClusterRole
// @Description  以 manifest 更新集群内原生 ClusterRole（rules 整体替换）
// @Tags         K8s-RBAC管理
// @Accept       json
// @Produce      json
// @Param        id    path   int    true  "集群 ID"
// @Param        body  body   object true  "ClusterRole manifest"
// @Success      200  {object}  utils.Response
// @Router       /k8s/clusters/{id}/rbac/clusterroles [put]
// @Security     BearerAuth
func (ctrl *K8sRbacController) UpdateClusterRole(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	var manifest map[string]interface{}
	if err := c.ShouldBindJSON(&manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}
	if err := ctrl.svc.UpdateClusterRole(clusterID, manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// ListRoles godoc
// @Summary      获取 Role 列表
// @Description  列出命名空间（为空则全集群）内原生 RBAC Role
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id         path   int    true  "集群 ID"
// @Param        namespace  query  string false "命名空间（空=全部）"
// @Param        search     query  string false "按名称搜索"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/roles [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) ListRoles(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	list, err := ctrl.svc.ListRoles(clusterID, c.Query("namespace"), c.Query("search"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(list))
}

// GetRole godoc
// @Summary      获取 Role 详情
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id         path   int    true  "集群 ID"
// @Param        namespace  path   string true  "命名空间"
// @Param        name       path   string true  "Role 名称"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/roles/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) GetRole(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	role, err := ctrl.svc.GetRole(clusterID, c.Param("namespace"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(role))
}

// UpdateRole godoc
// @Summary      更新 Role
// @Description  以 manifest 更新命名空间内原生 Role
// @Tags         K8s-RBAC管理
// @Accept       json
// @Produce      json
// @Param        id         path   int    true  "集群 ID"
// @Param        namespace  path   string true  "命名空间"
// @Param        body       body   object true  "Role manifest"
// @Success      200  {object}  utils.Response
// @Router       /k8s/clusters/{id}/rbac/roles/{namespace} [put]
// @Security     BearerAuth
func (ctrl *K8sRbacController) UpdateRole(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	var manifest map[string]interface{}
	if err := c.ShouldBindJSON(&manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}
	if err := ctrl.svc.UpdateRole(clusterID, c.Param("namespace"), manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// ListClusterRoleBindings godoc
// @Summary      获取 ClusterRoleBinding 列表
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id      path   int    true  "集群 ID"
// @Param        search  query  string false "按名称/主体搜索"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/clusterrolebindings [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) ListClusterRoleBindings(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	list, err := ctrl.svc.ListClusterRoleBindings(clusterID, c.Query("search"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(list))
}

// GetClusterRoleBinding godoc
// @Summary      获取 ClusterRoleBinding 详情
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id    path   int    true  "集群 ID"
// @Param        name  path   string true  "名称"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/clusterrolebindings/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) GetClusterRoleBinding(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	binding, err := ctrl.svc.GetClusterRoleBinding(clusterID, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(binding))
}

// ListRoleBindings godoc
// @Summary      获取 RoleBinding 列表
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id         path   int    true  "集群 ID"
// @Param        namespace  query  string false "命名空间（空=全部）"
// @Param        search     query  string false "按名称/主体搜索"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/rolebindings [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) ListRoleBindings(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	list, err := ctrl.svc.ListRoleBindings(clusterID, c.Query("namespace"), c.Query("search"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(list))
}

// GetRoleBinding godoc
// @Summary      获取 RoleBinding 详情
// @Tags         K8s-RBAC管理
// @Produce      json
// @Param        id         path   int    true  "集群 ID"
// @Param        namespace  path   string true  "命名空间"
// @Param        name       path   string true  "名称"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/rbac/rolebindings/{namespace}/{name} [get]
// @Security     BearerAuth
func (ctrl *K8sRbacController) GetRoleBinding(c *gin.Context) {
	clusterID, ok := clusterIDFromPath(c)
	if !ok {
		return
	}
	binding, err := ctrl.svc.GetRoleBinding(clusterID, c.Param("namespace"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(binding))
}
