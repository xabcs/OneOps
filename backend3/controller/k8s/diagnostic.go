package k8s

import (
	"net/http"
	"strconv"
	"time"

	k8ssvc "oneops/backend3/service/k8s"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// DiagnosticController 诊断控制器
type DiagnosticController struct {
	svc        *k8ssvc.DiagnosticService
	clusterSvc *k8ssvc.K8sClusterService
}

// NewDiagnosticController 创建诊断控制器
func NewDiagnosticController(svc *k8ssvc.DiagnosticService, clusterSvc *k8ssvc.K8sClusterService) *DiagnosticController {
	return &DiagnosticController{
		svc:        svc,
		clusterSvc: clusterSvc,
	}
}

// GetDiagnosticCommands godoc
// @Summary      获取诊断命令列表
// @Description  返回所有可用的 Java 诊断命令及其参数定义
// @Tags         K8s-诊断
// @Produce      json
// @Success      200  {object}  utils.Response{data=object{commands=object}}
// @Router       /k8s/diagnostic/commands [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetDiagnosticCommands(c *gin.Context) {
	commands := ctrl.svc.GetDiagnosticCommands()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    gin.H{"commands": commands},
	})
}

// GetJavaPods godoc
// @Summary      获取 Java 应用 Pod 列表
// @Description  获取指定集群和命名空间下的 Java 应用 Pod 列表
// @Tags         K8s-诊断
// @Produce      json
// @Param        clusterId  path      int     true  "集群 ID"
// @Param        namespace  path      string  true  "命名空间"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 内部错误"
// @Router       /k8s/diagnostic/pods/{clusterId}/{namespace} [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetJavaPods(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterIDStr := c.Param("clusterId")
	namespace := c.Param("namespace")

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的集群ID"})
		return
	}

	javaPods, err := ctrl.svc.GetJavaPods(uint(clusterID), namespace, userID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    javaPods,
	})
}

// GetNamespaces godoc
// @Summary      获取命名空间列表
// @Description  获取指定集群下所有可访问的命名空间列表
// @Tags         K8s-诊断
// @Produce      json
// @Param        clusterId  path      int  true  "集群 ID"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 内部错误"
// @Router       /k8s/diagnostic/namespaces/{clusterId} [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetNamespaces(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterIDStr := c.Param("clusterId")

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的集群ID"})
		return
	}

	namespaces, err := ctrl.svc.GetNamespaces(uint(clusterID), userID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    namespaces,
	})
}

// ExecuteDiagnostic godoc
// @Summary      执行诊断命令
// @Description  在指定 Pod 上执行 Java 诊断命令(通过 DaemonSet 侧车方式执行)
// @Tags         K8s-诊断
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "诊断执行请求"  example({"clusterId":"1","namespace":"default","podName":"app-pod","command":"jstack","args":{"pid":"1"},"timeout":30})
// @Success      200  {object}  utils.Response{data=object{status=string,output=string,timestamp=int,duration=int,method=string}}
// @Failure      200  {object}  utils.Response  "参数错误 / 无效的集群ID / 诊断执行失败"
// @Router       /k8s/diagnostic/execute [post]
// @Security     BearerAuth
func (ctrl *DiagnosticController) ExecuteDiagnostic(c *gin.Context) {
	var request struct {
		ClusterID string            `json:"clusterId"`
		Namespace string            `json:"namespace"`
		PodName   string            `json:"podName"`
		Command   string            `json:"command"`
		Args      map[string]string `json:"args"`
		Timeout   int               `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	username := c.GetString("username")
	if username == "" {
		username = "system"
	}

	clusterID, err := strconv.ParseUint(request.ClusterID, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的集群ID"})
		return
	}

	allowed, err := ctrl.clusterSvc.CheckClusterOperation(userID, uint(clusterID), "k8s.diagnostic.execute")
	if err != nil || !allowed {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权执行该操作（需要集群角色操作集包含 k8s.diagnostic.execute）"))
		return
	}

	result, err := ctrl.svc.ExecuteDiagnostic(&k8ssvc.DiagnosticExecRequest{
		ClusterID:    uint(clusterID),
		ClusterIDStr: request.ClusterID,
		Namespace:    request.Namespace,
		PodName:      request.PodName,
		Command:      request.Command,
		Args:         request.Args,
		Timeout:      request.Timeout,
		UserID:       userID,
		Username:     username,
	}, userID)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	if result.Status == "error" {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "诊断执行失败: " + result.ErrMsg,
			"data": map[string]interface{}{
				"status":    "error",
				"error":     result.ErrMsg,
				"timestamp": time.Now().Unix(),
				"duration":  result.Duration,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "诊断执行成功",
		"data": map[string]interface{}{
			"status":    "success",
			"output":    result.Output,
			"timestamp": time.Now().Unix(),
			"duration":  result.Duration,
			"method":    "daemonset",
		},
	})
}

// GetDiagnosticHistory godoc
// @Summary      获取诊断历史
// @Description  分页获取诊断执行历史记录,支持按集群/命名空间/Pod 名称筛选
// @Tags         K8s-诊断
// @Produce      json
// @Param        page        query     int     false  "页码"     default(1)
// @Param        pageSize    query     int     false  "每页数量"  default(20)
// @Param        clusterId   query     string  false  "集群 ID"
// @Param        namespace   query     string  false  "命名空间"
// @Param        podName     query     string  false  "Pod 名称"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "参数错误 / 内部错误"
// @Router       /k8s/diagnostic/history [get]
// @Security     BearerAuth
func (ctrl *DiagnosticController) GetDiagnosticHistory(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	clusterID := c.Query("clusterId")
	namespace := c.Query("namespace")
	podName := c.Query("podName")

	histories, total, err := ctrl.svc.GetDiagnosticHistory(clusterID, namespace, podName, params.GetPage(), params.GetPageSize())
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(histories, total, params)))
}
