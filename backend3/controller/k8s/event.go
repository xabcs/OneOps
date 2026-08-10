package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Events ==========

// ListEvents 获取 Event 列表
func (ctrl *K8sResourceController) ListEvents(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return
	}

	clusterID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的集群ID"))
		return
	}

	namespace := c.DefaultQuery("namespace", "default")
	fieldSelector := c.Query("fieldSelector")

	hasAccess, err := ctrl.clusterSvc.CheckUserClusterAccess(userID, uint(clusterID))
	if err != nil || !hasAccess {
		c.JSON(http.StatusOK, utils.ErrorForbidden("无权访问该集群"))
		return
	}

	events, err := ctrl.svc.ListEvents(uint(clusterID), namespace, fieldSelector)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取 Event 列表失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(events))
}
