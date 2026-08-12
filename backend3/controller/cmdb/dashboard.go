package cmdb

import (
	"net/http"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 统计与配置 ==========

// GetServerStats godoc
// @Summary      获取服务器统计信息
// @Description  获取服务器总数、状态分布、Agent 状态等统计信息
// @Tags         CMDB-服务器
// @Produce      json
// @Success      200  {object}  utils.Response  "统计信息"
// @Failure      200  {object}  utils.Response  "获取统计信息失败"
// @Router       /cmdb/servers/stats [get]
// @Security     BearerAuth
func (c *CMDBController) GetServerStats(ctx *gin.Context) {
	stats, err := c.svc.GetServerStats()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// GetServerConfig godoc
// @Summary      获取服务器配置信息
// @Description  通过 SSH 远程获取指定服务器的硬件/操作系统配置信息
// @Tags         CMDB-服务器
// @Accept       json
// @Produce      json
// @Param        body     body      object  true  "服务器连接信息"  examples({\"hostname\":\"web-01\",\"ip\":\"10.0.0.1\",\"sshUser\":\"root\",\"sshPort\":22})
// @Success      200      {object}  utils.Response  "服务器配置信息"
// @Failure      200      {object}  utils.Response  "请求参数错误 / 获取服务器配置失败"
// @Router       /cmdb/servers/config [post]
// @Security     BearerAuth
func (c *CMDBController) GetServerConfig(ctx *gin.Context) {
	var req struct {
		Hostname string `json:"hostname" binding:"required"`
		IP       string `json:"ip" binding:"required"`
		SSHUser  string `json:"sshUser"`
		SSHPort  int    `json:"sshPort"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 设置默认值
	if req.SSHUser == "" {
		req.SSHUser = "root"
	}
	if req.SSHPort == 0 {
		req.SSHPort = 22
	}

	config, err := c.svc.GetServerConfig(req.Hostname, req.IP, req.SSHUser, req.SSHPort)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取服务器配置失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(config))
}
