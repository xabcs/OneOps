package k8s

import (
	"net/http"
	"strconv"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/utils"
	. "oneops/backend3/service/k8s"

	"github.com/gin-gonic/gin"
)

// K8sPermissionController K8s权限控制器
type K8sPermissionController struct {
	svc *K8sClusterService
}

// NewK8sPermissionController 创建K8s权限控制器
func NewK8sPermissionController(svc *K8sClusterService) *K8sPermissionController {
	return &K8sPermissionController{
		svc: svc,
	}
}

// GetUserClusters godoc
// @Summary      获取用户可访问的集群
// @Description  分页获取当前登录用户有权限访问的集群列表
// @Tags         K8s-集群权限
// @Produce      json
// @Param        page      query     int  false  "页码"    default(1)
// @Param        pageSize  query     int  false  "每页数量" default(10)
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "获取集群列表失败"
// @Router       /k8s/users/clusters [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetUserClusters(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	clusters, total, err := ctrl.svc.GetClusters(userID, page, pageSize, nil)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群列表失败: "+err.Error()))
		return
	}

	result := make([]map[string]interface{}, len(clusters))
	for i, cluster := range clusters {
		result[i] = map[string]interface{}{
			"id":          cluster.ID,
			"name":        cluster.Name,
			"description": cluster.Description,
			"endpoint":    cluster.Endpoint,
			"clusterType": cluster.ClusterType,
			"region":      cluster.Region,
			"version":     cluster.Version,
			"nodeCount":   cluster.NodeCount,
			"status":      cluster.Status,
			"createdAt":   cluster.CreatedAt.Format("2006-01-02 15:04:05"),
			"updatedAt":   cluster.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"success":  true,
		"data":     result,
		"message":  "success",
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ========== 原生 RBAC 绑定（授权与执行层） ==========

// GetNativeBindings godoc
// @Summary      获取集群的原生 RBAC 绑定列表
// @Description  列出 OneOps 主体（用户/用户组）与 K8s 原生角色的绑定记录（B 模式）
// @Tags         K8s-集群权限
// @Produce      json
// @Param        id  path  int  true  "集群 ID"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/clusters/{id}/native-bindings [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetNativeBindings(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	bindings, err := ctrl.svc.GetNativeBindings(uint(clusterID))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取原生绑定失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(bindings))
}

// AssignNativeBinding godoc
// @Summary      创建原生 RBAC 绑定
// @Description  为 OneOps 用户/用户组绑定 K8s 原生角色（ClusterRole/Role）：落库并在集群内创建真实 Binding；此后该主体在此集群的操作经 impersonation 由原生 RBAC 判定
// @Tags         K8s-集群权限
// @Accept       json
// @Produce      json
// @Param        id    path      int                                         true  "集群 ID"
// @Param        body  body      modelk8s.AssignNativeRoleBindingRequest     true  "原生绑定请求"
// @Success      200   {object}  utils.Response  "绑定成功"
// @Router       /k8s/clusters/{id}/native-bindings [post]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) AssignNativeBinding(c *gin.Context) {
	operatorID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	var req modelk8s.AssignNativeRoleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	if err := ctrl.svc.AssignNativeBinding(operatorID, uint(clusterID), &req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建原生绑定失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("绑定成功"))
}

// RevokeNativeBinding godoc
// @Summary      撤销原生 RBAC 绑定
// @Description  删除集群内对应 Binding 与平台记录，主体失去经原生 RBAC 获得的权限
// @Tags         K8s-集群权限
// @Produce      json
// @Param        id          path  int  true  "集群 ID"
// @Param        bindingId   path  int  true  "原生绑定 ID"
// @Success      200  {object}  utils.Response  "撤销成功"
// @Router       /k8s/clusters/{id}/native-bindings/{bindingId} [delete]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) RevokeNativeBinding(c *gin.Context) {
	operatorID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	bindingID, err := strconv.ParseUint(c.Param("bindingId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的绑定ID"))
		return
	}

	if err := ctrl.svc.RevokeNativeBinding(operatorID, uint(clusterID), uint(bindingID)); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("撤销原生绑定失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("撤销成功"))
}

// GetAllNativeBindings godoc
// @Summary      获取全部集群的原生 RBAC 绑定列表（授权管理页）
// @Description  列出 OneOps 主体（用户/用户组）与 K8s 原生角色的绑定记录，可按集群筛选；支持集群/用户双视角审计
// @Tags         K8s-集群权限
// @Produce      json
// @Param        clusterId  query  int  false  "集群 ID（不传 = 全部集群）"
// @Success      200  {object}  utils.Response{data=object}
// @Router       /k8s/native-bindings [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetAllNativeBindings(c *gin.Context) {
	clusterID := uint(0)
	if v := c.Query("clusterId"); v != "" {
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
			return
		}
		clusterID = uint(id)
	}

	bindings, err := ctrl.svc.GetAllNativeBindings(clusterID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取授权记录失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(bindings))
}

// GetClusterOptions godoc
// @Summary      获取集群候选（授权管理页）
// @Description  返回全部集群的 id/名称/状态，供授权管理页筛选与表单选择；权限归属 k8s.permission.list（不借用 k8s.cluster.list）
// @Tags         K8s-集群权限
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]object}
// @Router       /k8s/clusters/options [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetClusterOptions(c *gin.Context) {
	clusters, err := ctrl.svc.GetClusterOptions()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取集群候选失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(clusters))
}

// GetSubjectOptions godoc
// @Summary      获取授权主体候选（用户+用户组）
// @Description  原生授权表单专用轻量接口，仅返回 id/用户名/昵称；权限归属 k8s.permission.list，与绑定查看对齐（不借用 system.user.list）
// @Tags         K8s-集群权限
// @Produce      json
// @Success      200  {object}  utils.Response{data=object{users=[]object,groups=[]object}}
// @Failure      200  {object}  utils.Response  "获取授权主体候选失败"
// @Router       /k8s/clusters/{id}/permission/subject-options [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetSubjectOptions(c *gin.Context) {
	data, err := ctrl.svc.GetSubjectOptions()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取授权主体候选失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(data))
}

// GetRoleOptions godoc
// @Summary      获取集群内角色候选
// @Description  原生授权表单专用：返回目标集群的 ClusterRole/Role 名称列表；权限归属 k8s.permission.list（不借用 k8s.rbac.view）
// @Tags         K8s-集群权限
// @Produce      json
// @Param        id    path     int    true   "集群 ID"
// @Param        kind  query    string false  "ClusterRole（默认）| Role"
// @Success      200  {object}  utils.Response{data=[]string}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 获取角色候选失败"
// @Router       /k8s/clusters/{id}/permission/role-options [get]
// @Security     BearerAuth
func (ctrl *K8sPermissionController) GetRoleOptions(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	names, err := ctrl.svc.GetRoleOptions(uint(clusterID), c.DefaultQuery("kind", "ClusterRole"))
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色候选失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(names))
}
