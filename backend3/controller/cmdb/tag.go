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

// GetServerTags godoc
// @Summary      获取服务器标签列表
// @Description  获取所有服务器标签列表
// @Tags         CMDB-标签
// @Produce      json
// @Success      200  {object}  utils.Response  "标签列表"
// @Failure      200  {object}  utils.Response  "获取标签列表失败"
// @Router       /cmdb/tags [get]
// @Security     BearerAuth
func (c *CMDBController) GetServerTags(ctx *gin.Context) {
	tags, err := c.svc.GetServerTags()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(tags))
}

// CreateServerTag godoc
// @Summary      创建服务器标签
// @Description  创建一个新的服务器标签
// @Tags         CMDB-标签
// @Accept       json
// @Produce      json
// @Param        tag  body      modelcmdb.ServerTag  true  "标签信息"
// @Success      200  {object}  utils.Response  "标签创建成功"
// @Failure      200  {object}  utils.Response  "请求参数错误 / 创建失败"
// @Router       /cmdb/tags [post]
// @Security     BearerAuth
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

// UpdateServerTag godoc
// @Summary      更新服务器标签
// @Description  根据标签 ID 更新服务器标签信息（部分字段更新）
// @Tags         CMDB-标签
// @Accept       json
// @Produce      json
// @Param        id   path      int                 true  "标签 ID"
// @Param        tag  body      modelcmdb.ServerTag  true  "需要更新的字段"
// @Success      200  {object}  utils.Response  "标签更新成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 请求参数错误 / 更新失败"
// @Router       /cmdb/tags/{id} [put]
// @Security     BearerAuth
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

// DeleteServerTag godoc
// @Summary      删除服务器标签
// @Description  根据标签 ID 删除服务器标签
// @Tags         CMDB-标签
// @Produce      json
// @Param        id  path  int  true  "标签 ID"
// @Success      200  {object}  utils.Response  "标签删除成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 删除失败"
// @Router       /cmdb/tags/{id} [delete]
// @Security     BearerAuth
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

// AssignServerTag godoc
// @Summary      为服务器分配标签
// @Description  为指定服务器分配单个标签
// @Tags         CMDB-标签
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "分配请求"  examples({\"serverId\":1,\"tagId\":2})
// @Success      200   {object}  utils.Response  "标签分配成功"
// @Failure      200   {object}  utils.Response  "请求参数错误 / 分配失败"
// @Router       /cmdb/tags/assign [post]
// @Security     BearerAuth
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

// RemoveServerTag godoc
// @Summary      移除服务器标签
// @Description  移除指定服务器的指定标签
// @Tags         CMDB-标签
// @Produce      json
// @Param        serverId  path  int  true  "服务器 ID"
// @Param        tagId     path  int  true  "标签 ID"
// @Success      200  {object}  utils.Response  "标签移除成功"
// @Failure      200  {object}  utils.Response  "移除失败"
// @Router       /cmdb/server-tags/{serverId}/{tagId} [delete]
// @Security     BearerAuth
func (c *CMDBController) RemoveServerTag(ctx *gin.Context) {
	tagID, _ := strconv.ParseUint(ctx.Param("tagId"), 10, 32)
	serverID, _ := strconv.ParseUint(ctx.Param("serverId"), 10, 32)

	if err := c.svc.RemoveServerTag(uint(serverID), uint(tagID)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签移除成功"))
}
