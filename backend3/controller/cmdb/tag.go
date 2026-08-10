package cmdb

import (
	"net/http"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 标签管理 ==========

// GetServerTags 获取服务器标签列表
func (c *CMDBController) GetServerTags(ctx *gin.Context) {
	tags, err := c.svc.GetServerTags()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(tags))
}

// CreateServerTag 创建服务器标签
func (c *CMDBController) CreateServerTag(ctx *gin.Context) {
	var tag modelcmdb.ServerTag
	if err := ctx.ShouldBindJSON(&tag); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.CreateServerTag(&tag); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签创建成功"))
}

// UpdateServerTag 更新服务器标签
func (c *CMDBController) UpdateServerTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.UpdateServerTag(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签更新成功"))
}

// DeleteServerTag 删除服务器标签
func (c *CMDBController) DeleteServerTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	if err := c.svc.DeleteServerTag(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签删除成功"))
}

// AssignServerTag 为服务器分配标签
func (c *CMDBController) AssignServerTag(ctx *gin.Context) {
	var req struct {
		ServerID uint `json:"serverId" binding:"required"`
		TagID    uint `json:"tagId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.AssignServerTag(req.ServerID, req.TagID); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签分配成功"))
}

// RemoveServerTag 移除服务器标签
func (c *CMDBController) RemoveServerTag(ctx *gin.Context) {
	tagID, _ := strconv.ParseUint(ctx.Param("tagId"), 10, 32)
	serverID, _ := strconv.ParseUint(ctx.Param("serverId"), 10, 32)

	if err := c.svc.RemoveServerTag(uint(serverID), uint(tagID)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签移除成功"))
}
