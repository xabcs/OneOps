package k8s

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== Events ==========

// ListEvents godoc
// @Summary      获取 Event 列表
// @Description  获取指定集群下指定命名空间的 K8s 事件列表，支持字段选择器
// @Tags         K8s-Pod
// @Produce      json
// @Param        id              path      int     true  "集群 ID"
// @Param        namespace       query     string  false  "命名空间" default(default)
// @Param        fieldSelector   query     string  false  "字段选择器"
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "无效的集群ID / 无权访问 / 获取失败"
// @Router       /k8s/clusters/{id}/events [get]
// @Security     BearerAuth
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
