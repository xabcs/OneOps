package cmdb

import (
	"net/http"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== SSH凭证管理 ==========

// GetSSHCredentials 获取SSH凭证列表，支持 ?type=user|system 筛选
func (c *CMDBController) GetSSHCredentials(ctx *gin.Context) {
	credentialType := ctx.Query("type") // "" | "user" | "system"
	credentials, err := c.svc.GetSSHCredentials(credentialType)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(credentials))
}

// GetSSHCredentialByID 获取SSH凭证详情
func (c *CMDBController) GetSSHCredentialByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	credential, err := c.svc.GetSSHCredentialByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("凭证不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(credential))
}

// CreateSSHCredential 创建SSH凭证
func (c *CMDBController) CreateSSHCredential(ctx *gin.Context) {
	var credential modelcmdb.SSHCredential
	if err := ctx.ShouldBindJSON(&credential); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.CreateSSHCredential(&credential); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("SSH凭证创建成功"))
}

// UpdateSSHCredential 更新SSH凭证
func (c *CMDBController) UpdateSSHCredential(ctx *gin.Context) {
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

	if err := c.svc.UpdateSSHCredential(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("SSH凭证更新成功"))
}

// DeleteSSHCredential 删除SSH凭证
func (c *CMDBController) DeleteSSHCredential(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	if err := c.svc.DeleteSSHCredential(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("SSH凭证删除成功"))
}

// TestSSHCredential 测试SSH凭证连接
func (c *CMDBController) TestSSHCredential(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	var req struct {
		TestIP   string `json:"testIp"`
		TestPort int    `json:"testPort"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	result, err := c.svc.TestSSHCredential(uint(id), req.TestIP, req.TestPort)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(result))
}
