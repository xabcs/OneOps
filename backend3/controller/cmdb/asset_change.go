package cmdb

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 资产变更记录 ==========

// GetAssetChanges 获取资产变更记录
func (c *CMDBController) GetAssetChanges(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	assetType := ctx.Query("assetType")
	assetID, _ := strconv.ParseUint(ctx.Query("assetId"), 10, 32)

	changes, total, err := c.svc.GetAssetChanges(assetType, uint(assetID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  changes,
		"total": total,
	}))
}

// SyncServerMetrics 手动触发单台主机指标采集（仅通过 Agent HTTP 拉取）
func (c *CMDBController) SyncServerMetrics(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	server, err := c.svc.GetServerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("服务器不存在"))
		return
	}

	// 检查 Agent 是否运行
	if server.AgentStatus != "running" {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("Agent 未运行，无法采集指标"))
		return
	}

	// 异步拉取 Agent 指标
	go c.agentSvc.PullMetrics(uint(id))

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 指标拉取任务已提交"))
}
