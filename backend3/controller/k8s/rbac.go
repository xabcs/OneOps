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
	svc        *k8ssvc.K8sRbacService
	clusterSvc *k8ssvc.K8sClusterService
}

// NewK8sRbacController 创建 RBAC 代管控制器
func NewK8sRbacController(svc *k8ssvc.K8sRbacService, clusterSvc *k8ssvc.K8sClusterService) *K8sRbacController {
	return &K8sRbacController{svc: svc, clusterSvc: clusterSvc}
}

func clusterIDFromPath(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return 0, false
	}
	return uint(id), true
}

// checkRbacAccess 集群级授权校验（H3）：此前 handler 只解析 clusterID 不查归属，
// 任何持有 k8s.rbac.* 系统码的用户可以平台管理员身份读写任意集群的 RBAC，构成提权链。
// 现：读类要求目标集群可见（CheckUserClusterAccess），写类要求集群内具备操作绑定
// （CheckClusterOperation），实际客户端再走 impersonation，由集群原生 RBAC 终判。
func (ctrl *K8sRbacController) checkRbacAccess(c *gin.Context, clusterID uint, write bool) bool {
	userID := c.GetUint("user_id")
	var (
		allowed bool
		err     error
	)
	if write {
		allowed, err = ctrl.clusterSvc.CheckClusterOperation(userID, clusterID, "k8s.rbac.manage")
	} else {
		allowed, err = ctrl.clusterSvc.CheckUserClusterAccess(userID, clusterID)
	}
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("集群授权校验失败"))
		return false
	}
	if !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群的 RBAC 资源"))
		return false
	}
	return true
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	list, err := ctrl.svc.ListClusterRoles(clusterID, userID, c.Query("search"))
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	role, err := ctrl.svc.GetClusterRole(clusterID, userID, c.Param("name"))
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, true) {
		return
	}
	var manifest map[string]interface{}
	if err := c.ShouldBindJSON(&manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}
	if err := ctrl.svc.UpdateClusterRole(clusterID, userID, manifest); err != nil {
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	list, err := ctrl.svc.ListRoles(clusterID, userID, c.Query("namespace"), c.Query("search"))
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	role, err := ctrl.svc.GetRole(clusterID, userID, c.Param("namespace"), c.Param("name"))
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, true) {
		return
	}
	var manifest map[string]interface{}
	if err := c.ShouldBindJSON(&manifest); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}
	if err := ctrl.svc.UpdateRole(clusterID, userID, c.Param("namespace"), manifest); err != nil {
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	list, err := ctrl.svc.ListClusterRoleBindings(clusterID, userID, c.Query("search"))
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	binding, err := ctrl.svc.GetClusterRoleBinding(clusterID, userID, c.Param("name"))
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	list, err := ctrl.svc.ListRoleBindings(clusterID, userID, c.Query("namespace"), c.Query("search"))
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
	userID := c.GetUint("user_id")
	if !ctrl.checkRbacAccess(c, clusterID, false) {
		return
	}
	binding, err := ctrl.svc.GetRoleBinding(clusterID, userID, c.Param("namespace"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessWithData(binding))
}
