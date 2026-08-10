package cmdb

import (
	"net/http"

	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 统计与配置 ==========

// GetServerStats 获取服务器统计信息
func (c *CMDBController) GetServerStats(ctx *gin.Context) {
	stats, err := c.svc.GetServerStats()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// GetServerConfig 获取服务器配置信息（通过SSH）
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
