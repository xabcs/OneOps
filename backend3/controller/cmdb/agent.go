package cmdb

import (
	"net/http"
	"strconv"
	"sync"

	. "oneops/backend3/service/cmdb"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentController Agent控制器
type AgentController struct {
	svc     *AgentService
	cmdbSvc *CMDBService
}

// NewAgentController 创建Agent控制器
func NewAgentController(svc *AgentService, cmdbSvc *CMDBService) *AgentController {
	return &AgentController{
		svc:     svc,
		cmdbSvc: cmdbSvc,
	}
}

// ========== Agent 部署/重启/卸载 ==========

// DeployAgent godoc
// @Summary      部署 Agent
// @Description  异步部署 Agent 到目标主机
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "Agent 部署任务已提交"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID"
// @Router       /cmdb/servers/{id}/agent/deploy [post]
// @Security     BearerAuth
func (c *AgentController) DeployAgent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	go func() {
		if err := c.svc.DeployAgent(uint(id)); err != nil {
			c.svc.MarkAgentFailed(uint(id), err.Error())
		}
	}()

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 部署任务已提交"))
}

// RestartAgent godoc
// @Summary      重启 Agent
// @Description  异步重启目标主机的 Agent 服务
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "Agent 重启任务已提交"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID"
// @Router       /cmdb/servers/{id}/agent/restart [post]
// @Security     BearerAuth
func (c *AgentController) RestartAgent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	go c.svc.RestartAgent(uint(id))

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 重启任务已提交"))
}

// UninstallAgent godoc
// @Summary      卸载 Agent
// @Description  异步卸载目标主机的 Agent
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "Agent 卸载任务已提交"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID"
// @Router       /cmdb/servers/{id}/agent/uninstall [post]
// @Security     BearerAuth
func (c *AgentController) UninstallAgent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	go func() {
		if err := c.svc.UninstallAgent(uint(id)); err != nil {
			c.svc.MarkAgentFailed(uint(id), err.Error())
		}
	}()

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 卸载任务已提交"))
}

// GetAgentStatus godoc
// @Summary      查询 Agent 状态
// @Description  查询指定服务器上 Agent 的当前状态、端口、版本和最近心跳时间
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "Agent 状态信息"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID / 服务器不存在"
// @Router       /cmdb/servers/{id}/agent/status [get]
// @Security     BearerAuth
func (c *AgentController) GetAgentStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	cmdbSvc := c.cmdbSvc
	server, err := cmdbSvc.GetServerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("服务器不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"agentStatus":     server.AgentStatus,
		"agentPort":       server.AgentPort,
		"agentVersion":    server.AgentVersion,
		"lastHeartbeatAt": server.LastHeartbeatAt,
	}))
}

// ReceiveAgentHeartbeat godoc
// @Summary      接收 Agent 心跳
// @Description  接收 Agent 上报的心跳数据（不经过 Auth 中间件，由 Agent 直接上报）
// @Tags         CMDB-Agent
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "心跳数据"
// @Success      200   {object}  utils.Response  "ok"
// @Failure      200   {object}  utils.Response  "参数错误 / 处理失败"
// @Router       /cmdb/agent/heartbeat [post]
func (c *AgentController) ReceiveAgentHeartbeat(ctx *gin.Context) {
	var data AgentHeartbeatData
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	if err := c.svc.ReceiveHeartbeat(data); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("ok"))
}

// ========== Agent 管理页面专用接口 ==========

// GetAgentList godoc
// @Summary      获取 Agent 管理列表
// @Description  分页获取 Agent 管理列表，支持按主机名、IP、Agent 状态筛选
// @Tags         CMDB-Agent
// @Produce      json
// @Param        page        query     int     false  "页码"           default(1)
// @Param        pageSize    query     int     false  "每页数量"       default(10)
// @Param        hostname    query     string  false  "主机名"
// @Param        ip          query     string  false  "IP 地址"
// @Param        agentStatus query     string  false  "Agent 状态"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取列表失败"
// @Router       /cmdb/agents [get]
// @Security     BearerAuth
func (c *AgentController) GetAgentList(ctx *gin.Context) {
	var params dto.ServerQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	query := make(map[string]interface{})
	if params.Hostname != "" {
		query["hostname"] = params.Hostname
	}
	if params.IP != "" {
		query["ip"] = params.IP
	}
	if params.AgentStatus != "" {
		query["agentStatus"] = params.AgentStatus
	}

	cmdbSvc := c.cmdbSvc
	servers, total, err := cmdbSvc.GetServers(query, params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(servers, total, params.BasePageQuery)))
}

// BatchDeployAgent godoc
// @Summary      批量部署 Agent
// @Description  批量异步部署 Agent 到多台主机
// @Tags         CMDB-Agent
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "服务器 ID 列表"  examples({\"serverIds\":[1,2,3]})
// @Success      200   {object}  utils.Response  "批量 Agent 部署任务已提交"
// @Failure      200   {object}  utils.Response  "参数错误 / serverIds 不能为空"
// @Router       /cmdb/agents/batch-deploy [post]
// @Security     BearerAuth
func (c *AgentController) BatchDeployAgent(ctx *gin.Context) {
	var req struct {
		ServerIDs []uint `json:"serverIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}
	if len(req.ServerIDs) == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("serverIds 不能为空"))
		return
	}

	for _, id := range req.ServerIDs {
		serverID := id
		go func() {
			if err := c.svc.DeployAgent(serverID); err != nil {
				_ = err
			}
		}()
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("批量 Agent 部署任务已提交"))
}

// BatchUninstallAgent godoc
// @Summary      批量卸载 Agent
// @Description  批量异步卸载多台主机的 Agent
// @Tags         CMDB-Agent
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "服务器 ID 列表"  examples({\"serverIds\":[1,2,3]})
// @Success      200   {object}  utils.Response  "批量 Agent 卸载任务已提交"
// @Failure      200   {object}  utils.Response  "参数错误 / serverIds 不能为空"
// @Router       /cmdb/agents/batch-uninstall [post]
// @Security     BearerAuth
func (c *AgentController) BatchUninstallAgent(ctx *gin.Context) {
	var req struct {
		ServerIDs []uint `json:"serverIds"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}
	if len(req.ServerIDs) == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("serverIds 不能为空"))
		return
	}

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	for _, id := range req.ServerIDs {
		serverID := id
		wg.Add(1)
		sem <- struct{}{}
		utils.SafeGo("agent-batch-uninstall", func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := c.svc.UninstallAgent(serverID); err != nil {
				logger.Error("批量卸载 Agent 失败", zap.Uint("server_id", serverID), zap.Error(err))
			}
		})
	}
	utils.SafeGo("agent-batch-uninstall-wait", wg.Wait)

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("批量 Agent 卸载任务已提交"))
}

// DeleteAgentRecord godoc
// @Summary      删除 Agent 记录
// @Description  仅清空服务器上的 Agent 相关字段，不执行 SSH 操作
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "Agent 记录已清除"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID / 清除失败"
// @Router       /cmdb/agents/{id} [delete]
// @Security     BearerAuth
func (c *AgentController) DeleteAgentRecord(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	cmdbSvc := c.cmdbSvc
	if err := cmdbSvc.ClearAgentRecord(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 记录已清除"))
}

// TestSSHConnection godoc
// @Summary      测试 SSH 连接
// @Description  测试到指定服务器的 SSH 连接是否可用
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "测试结果"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID / 测试失败"
// @Router       /cmdb/servers/{id}/test-connection [post]
// @Security     BearerAuth
func (c *AgentController) TestSSHConnection(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	result, err := c.svc.TestSSHConnection(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	if result.Success {
		ctx.JSON(http.StatusOK, utils.SuccessWithData(result))
	} else {
		// 连接失败，返回200状态码但success为false，前端可以显示详细信息
		ctx.JSON(http.StatusOK, gin.H{
			"code":    200,
			"success": false,
			"message": result.Message,
			"data":    result,
		})
	}
}

// ========================================
// Agent 版本管理
// ========================================

// GetAgentVersions godoc
// @Summary      获取 Agent 版本列表
// @Description  获取所有 Agent 版本列表
// @Tags         CMDB-Agent
// @Produce      json
// @Success      200  {object}  utils.Response  "版本列表"
// @Failure      200  {object}  utils.Response  "获取版本列表失败"
// @Router       /cmdb/agent-versions [get]
// @Security     BearerAuth
func (c *AgentController) GetAgentVersions(ctx *gin.Context) {
	versions, err := c.svc.GetAllVersions()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(versions))
}

// GetLatestAgentVersion godoc
// @Summary      获取最新 Agent 版本
// @Description  获取最新的 Agent 版本信息
// @Tags         CMDB-Agent
// @Produce      json
// @Success      200  {object}  utils.Response  "最新版本信息"
// @Failure      200  {object}  utils.Response  "获取最新版本失败"
// @Router       /cmdb/agent-versions/latest [get]
// @Security     BearerAuth
func (c *AgentController) GetLatestAgentVersion(ctx *gin.Context) {
	version, err := c.svc.GetLatestVersion()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(version))
}

// GetAgentVersionByID godoc
// @Summary      获取 Agent 版本详情
// @Description  根据版本 ID 获取 Agent 版本详细信息
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "版本 ID"
// @Success      200  {object}  utils.Response{data=modelcmdb.AgentVersion}
// @Failure      200  {object}  utils.Response  "无效的版本 ID / 获取失败"
// @Router       /cmdb/agent-versions/{id} [get]
// @Security     BearerAuth
func (c *AgentController) GetAgentVersionByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的版本ID"))
		return
	}

	version, err := c.svc.GetVersionByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(version))
}

// CreateAgentVersion godoc
// @Summary      创建 Agent 版本
// @Description  创建一个新的 Agent 版本
// @Tags         CMDB-Agent
// @Accept       json
// @Produce      json
// @Param        version  body      modelcmdb.AgentVersion  true  "版本信息"
// @Success      200      {object}  utils.Response  "Agent 版本创建成功"
// @Failure      200      {object}  utils.Response  "请求参数错误 / 创建失败"
// @Router       /cmdb/agent-versions [post]
// @Security     BearerAuth
func (c *AgentController) CreateAgentVersion(ctx *gin.Context) {
	var version modelcmdb.AgentVersion
	if err := ctx.ShouldBindJSON(&version); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	operator := ctx.GetString("username")
	if operator == "" {
		operator = "system"
	}

	if err := c.svc.CreateVersion(&version, operator); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 版本创建成功"))
}

// UpdateAgentVersion godoc
// @Summary      更新 Agent 版本
// @Description  根据版本 ID 更新 Agent 版本信息（部分字段更新）
// @Tags         CMDB-Agent
// @Accept       json
// @Produce      json
// @Param        id       path      int                   true  "版本 ID"
// @Param        version  body      modelcmdb.AgentVersion  true  "需要更新的字段"
// @Success      200      {object}  utils.Response  "Agent 版本更新成功"
// @Failure      200      {object}  utils.Response  "无效的版本 ID / 请求参数错误 / 更新失败"
// @Router       /cmdb/agent-versions/{id} [put]
// @Security     BearerAuth
func (c *AgentController) UpdateAgentVersion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的版本ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.UpdateVersion(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 版本更新成功"))
}

// DeleteAgentVersion godoc
// @Summary      删除 Agent 版本
// @Description  根据版本 ID 删除 Agent 版本
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "版本 ID"
// @Success      200  {object}  utils.Response  "Agent 版本删除成功"
// @Failure      200  {object}  utils.Response  "无效的版本 ID / 删除失败"
// @Router       /cmdb/agent-versions/{id} [delete]
// @Security     BearerAuth
func (c *AgentController) DeleteAgentVersion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的版本ID"))
		return
	}

	if err := c.svc.DeleteVersion(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 版本删除成功"))
}

// ========================================
// Agent 升级管理
// ========================================

// UpgradeAgent godoc
// @Summary      升级 Agent
// @Description  将指定服务器的 Agent 升级到目标版本
// @Tags         CMDB-Agent
// @Accept       json
// @Produce      json
// @Param        id    path      int     true  "服务器 ID"
// @Param        body  body      object  true  "升级目标版本"  examples({\"targetVersion\":\"v1.2.0\"})
// @Success      200   {object}  utils.Response  "Agent 升级任务已提交"
// @Failure      200   {object}  utils.Response  "无效的服务器 ID / 请求参数错误 / 升级失败"
// @Router       /cmdb/servers/{id}/agent/upgrade [post]
// @Security     BearerAuth
func (c *AgentController) UpgradeAgent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	var req struct {
		TargetVersion string `json:"targetVersion" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.UpgradeAgent(uint(id), req.TargetVersion); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 升级任务已提交"))
}

// GetUpgradeTasks godoc
// @Summary      获取升级任务列表
// @Description  分页获取 Agent 升级任务列表，支持按状态筛选
// @Tags         CMDB-Agent
// @Produce      json
// @Param        page      query     int     false  "页码"      default(1)
// @Param        pageSize  query     int     false  "每页数量"  default(10)
// @Param        status    query     string  false  "任务状态"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取任务列表失败"
// @Router       /cmdb/agent-upgrade-tasks [get]
// @Security     BearerAuth
func (c *AgentController) GetUpgradeTasks(ctx *gin.Context) {
	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}
	status := ctx.Query("status")

	tasks, total, err := c.svc.GetUpgradeTasks(params.GetPage(), params.GetPageSize(), status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(tasks, total, params)))
}

// GetUpgradeTaskByID godoc
// @Summary      获取升级任务详情
// @Description  根据任务 ID 获取 Agent 升级任务详情
// @Tags         CMDB-Agent
// @Produce      json
// @Param        id  path  int  true  "任务 ID"
// @Success      200  {object}  utils.Response  "任务详情"
// @Failure      200  {object}  utils.Response  "无效的任务 ID / 获取失败"
// @Router       /cmdb/agent-upgrade-tasks/{id} [get]
// @Security     BearerAuth
func (c *AgentController) GetUpgradeTaskByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的任务ID"))
		return
	}

	task, err := c.svc.GetUpgradeTaskByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(task))
}
