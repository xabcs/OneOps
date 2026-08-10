package cmdb

import (
	"net/http"
	"strconv"
	"strings"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/middleware"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ========== 服务器管理 ==========

// GetServers 获取服务器列表
func (c *CMDBController) GetServers(ctx *gin.Context) {
	var params dto.ServerQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误: "+err.Error()))
		return
	}

	// 转换为 Service 层使用的查询参数
	query := make(map[string]interface{})
	if params.Hostname != "" {
		query["hostname"] = params.Hostname
	}
	if params.IP != "" {
		query["ip"] = params.IP
	}
	if params.InnerIP != "" {
		query["innerIp"] = params.InnerIP
	}
	if params.Env != "" {
		query["env"] = params.Env
	}
	if params.Status != "" {
		query["status"] = params.Status
	}
	if params.Provider != "" {
		query["provider"] = params.Provider
	}
	if params.AgentStatus != "" {
		query["agentStatus"] = params.AgentStatus
	}
	if params.GroupID != nil {
		query["groupId"] = *params.GroupID
	}
	if params.BusinessUnitID != nil {
		query["businessUnitId"] = *params.BusinessUnitID
	}
	if params.Tags != "" {
		query["tags"] = params.Tags
	}

	servers, total, err := c.svc.GetServers(query, params.Page, params.PageSize)
	if err != nil {
		middleware.HandleControllerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  servers,
		"total": total,
	}))
}

// GetServerByID 获取服务器详情
func (c *CMDBController) GetServerByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	server, err := c.svc.GetServerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("服务器不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(server))
}

// GetServerForConnect 获取连接所需的服务器信息（轻量级）
func (c *CMDBController) GetServerForConnect(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	server, err := c.svc.GetServerForConnect(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("服务器不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(server))
}

// CreateServer 创建服务器
func (c *CMDBController) CreateServer(ctx *gin.Context) {
	var server modelcmdb.Server
	if err := ctx.ShouldBindJSON(&server); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	// 获取当前用户
	operator := ctx.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := c.svc.CreateServer(&server, operator); err != nil {
		// 检查是否是重复键错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "Duplicate entry") && strings.Contains(errMsg, "hostname") {
			ctx.JSON(http.StatusOK, utils.ErrorDuplicateHostname())
			return
		}
		if strings.Contains(errMsg, "Duplicate entry") && strings.Contains(errMsg, "ip") {
			ctx.JSON(http.StatusOK, utils.ErrorDuplicateIP())
			return
		}
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器创建成功"))
}

// UpdateServer 更新服务器
func (c *CMDBController) UpdateServer(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	// 获取当前用户
	operator := ctx.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := c.svc.UpdateServer(uint(id), updates, operator); err != nil {
		// 检查是否是重复键错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "Duplicate entry") && strings.Contains(errMsg, "hostname") {
			ctx.JSON(http.StatusOK, utils.ErrorDuplicateHostname())
			return
		}
		if strings.Contains(errMsg, "Duplicate entry") && strings.Contains(errMsg, "ip") {
			ctx.JSON(http.StatusOK, utils.ErrorDuplicateIP())
			return
		}
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器更新成功"))
}

// DeleteServer 删除服务器
func (c *CMDBController) DeleteServer(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	// 获取当前用户
	operator := ctx.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := c.svc.DeleteServer(uint(id), operator); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器删除成功"))
}
