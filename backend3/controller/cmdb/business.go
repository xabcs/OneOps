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

// GetBusinessUnits 获取业务系统列表（树形结构）
func (c *CMDBController) GetBusinessUnits(ctx *gin.Context) {
	units, err := c.svc.GetBusinessUnits()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(units))
}

// CreateBusinessUnit 创建业务系统
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

// UpdateBusinessUnit 更新业务系统
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

// DeleteBusinessUnit 删除业务系统
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
