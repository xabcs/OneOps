package cmdb

import (
	"net/http"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 机房机柜管理 ==========

// GetServerRooms godoc
// @Summary      获取机房列表
// @Description  获取所有机房列表
// @Tags         CMDB-机房
// @Produce      json
// @Success      200  {object}  utils.Response  "机房列表"
// @Failure      200  {object}  utils.Response  "获取机房列表失败"
// @Router       /cmdb/rooms [get]
// @Security     BearerAuth
func (c *CMDBController) GetServerRooms(ctx *gin.Context) {
	rooms, err := c.svc.GetServerRooms()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(rooms))
}

// CreateServerRoom godoc
// @Summary      创建机房
// @Description  创建一个新的机房
// @Tags         CMDB-机房
// @Accept       json
// @Produce      json
// @Param        room  body      modelcmdb.ServerRoom  true  "机房信息"
// @Success      200   {object}  utils.Response  "机房创建成功"
// @Failure      200   {object}  utils.Response  "请求参数错误 / 创建失败"
// @Router       /cmdb/rooms [post]
// @Security     BearerAuth
func (c *CMDBController) CreateServerRoom(ctx *gin.Context) {
	var room modelcmdb.ServerRoom
	if err := ctx.ShouldBindJSON(&room); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.CreateServerRoom(&room); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("机房创建成功"))
}

// UpdateServerRoom godoc
// @Summary      更新机房
// @Description  根据机房 ID 更新机房信息（部分字段更新）
// @Tags         CMDB-机房
// @Accept       json
// @Produce      json
// @Param        id    path      int                 true  "机房 ID"
// @Param        room  body      modelcmdb.ServerRoom  true  "需要更新的字段"
// @Success      200   {object}  utils.Response  "机房更新成功"
// @Failure      200   {object}  utils.Response  "无效的 ID / 请求参数错误 / 更新失败"
// @Router       /cmdb/rooms/{id} [put]
// @Security     BearerAuth
func (c *CMDBController) UpdateServerRoom(ctx *gin.Context) {
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

	if err := c.svc.UpdateServerRoom(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("机房更新成功"))
}

// DeleteServerRoom godoc
// @Summary      删除机房
// @Description  根据机房 ID 删除机房
// @Tags         CMDB-机房
// @Produce      json
// @Param        id  path  int  true  "机房 ID"
// @Success      200  {object}  utils.Response  "机房删除成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 删除失败"
// @Router       /cmdb/rooms/{id} [delete]
// @Security     BearerAuth
func (c *CMDBController) DeleteServerRoom(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	if err := c.svc.DeleteServerRoom(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("机房删除成功"))
}

// GetCabinets godoc
// @Summary      获取机柜列表
// @Description  根据机房 ID 获取机柜列表
// @Tags         CMDB-机房
// @Produce      json
// @Param        roomId  query  int  false  "机房 ID"
// @Success      200  {object}  utils.Response  "机柜列表"
// @Failure      200  {object}  utils.Response  "获取机柜列表失败"
// @Router       /cmdb/cabinets [get]
// @Security     BearerAuth
func (c *CMDBController) GetCabinets(ctx *gin.Context) {
	roomID, _ := strconv.ParseUint(ctx.Query("roomId"), 10, 32)

	cabinets, err := c.svc.GetCabinets(uint(roomID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(cabinets))
}
