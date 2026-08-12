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

// GetSSHCredentials godoc
// @Summary      获取 SSH 凭证列表
// @Description  获取 SSH 凭证列表，支持按类型（user/system）筛选
// @Tags         CMDB-SSH凭证
// @Produce      json
// @Param        type  query  string  false  "凭证类型（user/system）"
// @Success      200  {object}  utils.Response  "凭证列表"
// @Failure      200  {object}  utils.Response  "获取凭证列表失败"
// @Router       /cmdb/ssh-credentials [get]
// @Security     BearerAuth
func (c *CMDBController) GetSSHCredentials(ctx *gin.Context) {
	credentialType := ctx.Query("type") // "" | "user" | "system"
	credentials, err := c.svc.GetSSHCredentials(credentialType)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(credentials))
}

// GetSSHCredentialByID godoc
// @Summary      获取 SSH 凭证详情
// @Description  根据凭证 ID 获取 SSH 凭证详情
// @Tags         CMDB-SSH凭证
// @Produce      json
// @Param        id  path  int  true  "凭证 ID"
// @Success      200  {object}  utils.Response{data=modelcmdb.SSHCredential}
// @Failure      200  {object}  utils.Response  "无效的 ID / 凭证不存在"
// @Router       /cmdb/ssh-credentials/{id} [get]
// @Security     BearerAuth
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

// CreateSSHCredential godoc
// @Summary      创建 SSH 凭证
// @Description  创建一个新的 SSH 凭证
// @Tags         CMDB-SSH凭证
// @Accept       json
// @Produce      json
// @Param        credential  body      modelcmdb.SSHCredential  true  "凭证信息"
// @Success      200         {object}  utils.Response  "SSH 凭证创建成功"
// @Failure      200         {object}  utils.Response  "请求参数错误 / 创建失败"
// @Router       /cmdb/ssh-credentials [post]
// @Security     BearerAuth
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

// UpdateSSHCredential godoc
// @Summary      更新 SSH 凭证
// @Description  根据凭证 ID 更新 SSH 凭证信息（部分字段更新）
// @Tags         CMDB-SSH凭证
// @Accept       json
// @Produce      json
// @Param        id          path      int                    true  "凭证 ID"
// @Param        credential  body      modelcmdb.SSHCredential  true  "需要更新的字段"
// @Success      200         {object}  utils.Response  "SSH 凭证更新成功"
// @Failure      200         {object}  utils.Response  "无效的 ID / 请求参数错误 / 更新失败"
// @Router       /cmdb/ssh-credentials/{id} [put]
// @Security     BearerAuth
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

// DeleteSSHCredential godoc
// @Summary      删除 SSH 凭证
// @Description  根据凭证 ID 删除 SSH 凭证
// @Tags         CMDB-SSH凭证
// @Produce      json
// @Param        id  path  int  true  "凭证 ID"
// @Success      200  {object}  utils.Response  "SSH 凭证删除成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 删除失败"
// @Router       /cmdb/ssh-credentials/{id} [delete]
// @Security     BearerAuth
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

// TestSSHCredential godoc
// @Summary      测试 SSH 凭证连接
// @Description  使用指定凭证测试 SSH 连接是否可用
// @Tags         CMDB-SSH凭证
// @Accept       json
// @Produce      json
// @Param        id    path      int     true  "凭证 ID"
// @Param        body  body      object  true  "测试目标"  examples({\"testIp\":\"10.0.0.1\",\"testPort\":22})
// @Success      200   {object}  utils.Response  "测试结果"
// @Failure      200   {object}  utils.Response  "无效的 ID / 请求参数错误 / 测试失败"
// @Router       /cmdb/ssh-credentials/{id}/test [post]
// @Security     BearerAuth
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
