package cmdb

import (
	"net/http"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 业务系统管理 ==========

// GetBusinessUnits godoc
// @Summary      获取业务系统列表
// @Description  获取业务系统列表（树形结构）
// @Tags         CMDB-业务单元
// @Produce      json
// @Success      200  {object}  utils.Response  "业务系统树形列表"
// @Failure      200  {object}  utils.Response  "获取业务系统列表失败"
// @Router       /cmdb/business-units [get]
// @Security     BearerAuth
func (c *CMDBController) GetBusinessUnits(ctx *gin.Context) {
	units, err := c.svc.GetBusinessUnits()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(units))
}

// CreateBusinessUnit godoc
// @Summary      创建业务系统
// @Description  创建一个新的业务系统
// @Tags         CMDB-业务单元
// @Accept       json
// @Produce      json
// @Param        unit  body      modelcmdb.BusinessUnit  true  "业务系统信息"
// @Success      200   {object}  utils.Response  "业务系统创建成功"
// @Failure      200   {object}  utils.Response  "请求参数错误 / 创建失败"
// @Router       /cmdb/business-units [post]
// @Security     BearerAuth
func (c *CMDBController) CreateBusinessUnit(ctx *gin.Context) {
	var unit modelcmdb.BusinessUnit
	if err := ctx.ShouldBindJSON(&unit); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 获取当前用户
	operator := ctx.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := c.svc.CreateBusinessUnit(&unit, operator); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("业务系统创建成功"))
}

// UpdateBusinessUnit godoc
// @Summary      更新业务系统
// @Description  根据业务系统 ID 更新业务系统信息（部分字段更新）
// @Tags         CMDB-业务单元
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "业务系统 ID"
// @Param        unit  body      modelcmdb.BusinessUnit  true  "需要更新的字段"
// @Success      200   {object}  utils.Response  "业务系统更新成功"
// @Failure      200   {object}  utils.Response  "无效的 ID / 请求参数错误 / 更新失败"
// @Router       /cmdb/business-units/{id} [put]
// @Security     BearerAuth
func (c *CMDBController) UpdateBusinessUnit(ctx *gin.Context) {
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

	if err := c.svc.UpdateBusinessUnit(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("业务系统更新成功"))
}

// DeleteBusinessUnit godoc
// @Summary      删除业务系统
// @Description  根据业务系统 ID 删除业务系统
// @Tags         CMDB-业务单元
// @Produce      json
// @Param        id  path  int  true  "业务系统 ID"
// @Success      200  {object}  utils.Response  "业务系统删除成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 删除失败"
// @Router       /cmdb/business-units/{id} [delete]
// @Security     BearerAuth
func (c *CMDBController) DeleteBusinessUnit(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	if err := c.svc.DeleteBusinessUnit(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("业务系统删除成功"))
}
