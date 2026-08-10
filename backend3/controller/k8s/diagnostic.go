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
	svc *k8ssvc.DiagnosticService
}

// NewDiagnosticController 创建诊断控制器
func NewDiagnosticController(svc *k8ssvc.DiagnosticService) *DiagnosticController {
	return &DiagnosticController{
		svc: svc,
	}
}

// GetDiagnosticCommands 获取诊断命令列表
func (ctrl *DiagnosticController) GetDiagnosticCommands(c *gin.Context) {
	commands := ctrl.svc.GetDiagnosticCommands()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    gin.H{"commands": commands},
	})
}

// GetJavaPods 获取Java应用Pod列表
func (ctrl *DiagnosticController) GetJavaPods(c *gin.Context) {
	clusterIDStr := c.Param("clusterId")
	namespace := c.Param("namespace")

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的集群ID"})
		return
	}

	javaPods, err := ctrl.svc.GetJavaPods(uint(clusterID), namespace)
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

// GetNamespaces 获取命名空间列表
func (ctrl *DiagnosticController) GetNamespaces(c *gin.Context) {
	clusterIDStr := c.Param("clusterId")

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的集群ID"})
		return
	}

	namespaces, err := ctrl.svc.GetNamespaces(uint(clusterID))
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

// ExecuteDiagnostic 执行诊断
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
	})
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

// GetDiagnosticHistory 获取诊断历史
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
