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

// GetServers godoc
// @Summary      获取服务器列表
// @Description  分页获取服务器列表，支持多条件筛选
// @Tags         CMDB-服务器
// @Produce      json
// @Param        page           query     int     false  "页码"                  default(1)
// @Param        pageSize       query     int     false  "每页数量"              default(10)
// @Param        hostname       query     string  false  "主机名"
// @Param        ip             query     string  false  "IP 地址"
// @Param        innerIp        query     string  false  "内网 IP"
// @Param        env            query     string  false  "环境"
// @Param        status         query     string  false  "状态"
// @Param        provider       query     string  false  "云供应商"
// @Param        agentStatus    query     string  false  "Agent 状态"
// @Param        groupId        query     int     false  "分组 ID"
// @Param        businessUnitId query     int     false  "业务单元 ID"
// @Param        tags           query     string  false  "标签（逗号分隔）"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取服务器列表失败"
// @Router       /cmdb/servers [get]
// @Security     BearerAuth
func (c *CMDBController) GetServers(ctx *gin.Context) {
	var params dto.ServerQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
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

	servers, total, err := c.svc.GetServers(query, params.GetPage(), params.GetPageSize())
	if err != nil {
		middleware.HandleControllerError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(servers, total, params.BasePageQuery)))
}

// GetServerOptions godoc
// @Summary      获取服务器选项列表
// @Description  返回所有服务器的基本信息（不分页，用于选择器）
// @Tags         CMDB-服务器
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "获取服务器选项失败"
// @Router       /cmdb/servers/options [get]
// @Security     BearerAuth
func (c *CMDBController) GetServerOptions(ctx *gin.Context) {
	options, err := c.svc.GetServerOptions()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取服务器选项失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.SuccessWithData(options))
}

// GetServerByID godoc
// @Summary      获取服务器详情
// @Description  根据服务器 ID 获取服务器详细信息
// @Tags         CMDB-服务器
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response{data=modelcmdb.Server}
// @Failure      200  {object}  utils.Response  "无效的 ID / 服务器不存在"
// @Router       /cmdb/servers/{id} [get]
// @Security     BearerAuth
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

// GetServerForConnect godoc
// @Summary      获取连接所需的服务器信息
// @Description  返回连接服务器所需的轻量级信息（用于堡垒机连接前置校验）
// @Tags         CMDB-服务器
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "连接信息"
// @Failure      200  {object}  utils.Response  "无效的 ID / 服务器不存在"
// @Router       /cmdb/servers/{id}/connect [get]
// @Security     BearerAuth
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

// CreateServer godoc
// @Summary      创建服务器
// @Description  创建一台新的服务器记录
// @Tags         CMDB-服务器
// @Accept       json
// @Produce      json
// @Param        server  body      modelcmdb.Server  true  "服务器信息"
// @Success      200     {object}  utils.Response  "服务器创建成功"
// @Failure      200     {object}  utils.Response  "请求参数错误 / 主机名或 IP 重复 / 创建失败"
// @Router       /cmdb/servers [post]
// @Security     BearerAuth
func (c *CMDBController) CreateServer(ctx *gin.Context) {
	var server modelcmdb.Server
	if err := ctx.ShouldBindJSON(&server); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
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

// UpdateServer godoc
// @Summary      更新服务器
// @Description  根据服务器 ID 更新服务器信息（部分字段更新）
// @Tags         CMDB-服务器
// @Accept       json
// @Produce      json
// @Param        id      path      int               true  "服务器 ID"
// @Param        server  body      modelcmdb.Server  true  "需要更新的字段"
// @Success      200     {object}  utils.Response  "服务器更新成功"
// @Failure      200     {object}  utils.Response  "无效的 ID / 请求参数错误 / 主机名或 IP 重复 / 更新失败"
// @Router       /cmdb/servers/{id} [put]
// @Security     BearerAuth
func (c *CMDBController) UpdateServer(ctx *gin.Context) {
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

// DeleteServer godoc
// @Summary      删除服务器
// @Description  根据服务器 ID 删除服务器记录
// @Tags         CMDB-服务器
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "服务器删除成功"
// @Failure      200  {object}  utils.Response  "无效的 ID / 删除失败"
// @Router       /cmdb/servers/{id} [delete]
// @Security     BearerAuth
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
