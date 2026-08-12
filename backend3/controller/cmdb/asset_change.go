package cmdb

import (
	"net/http"
	"strconv"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 资产变更记录 ==========

// GetAssetChanges godoc
// @Summary      获取资产变更记录
// @Description  分页获取资产变更记录，支持按资产类型和资产 ID 筛选
// @Tags         CMDB-资产变更
// @Produce      json
// @Param        page      query     int     false  "页码"          default(1)
// @Param        pageSize  query     int     false  "每页数量"      default(10)
// @Param        assetType query     string  false  "资产类型"
// @Param        assetId   query     int     false  "资产 ID"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取变更记录失败"
// @Router       /cmdb/asset-changes [get]
// @Security     BearerAuth
func (c *CMDBController) GetAssetChanges(ctx *gin.Context) {
	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	assetType := ctx.Query("assetType")
	assetID, _ := strconv.ParseUint(ctx.Query("assetId"), 10, 32)

	changes, total, err := c.svc.GetAssetChanges(assetType, uint(assetID), params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(changes, total, params)))
}

// SyncServerMetrics godoc
// @Summary      同步服务器指标
// @Description  手动触发单台主机的指标采集（仅通过 Agent HTTP 拉取）
// @Tags         CMDB-资产变更
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "Agent 指标拉取任务已提交"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID / 服务器不存在 / Agent 未运行"
// @Router       /cmdb/servers/{id}/sync-metrics [post]
// @Security     BearerAuth
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
