package cmdb

import (
	"net/http"
	"strconv"

	. "oneops/backend3/service/cmdb"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
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

// DeployAgent 部署 Agent 到目标主机
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

// RestartAgent 重启目标主机 Agent 服务
func (c *AgentController) RestartAgent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	go c.svc.RestartAgent(uint(id))

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 重启任务已提交"))
}

// UninstallAgent 卸载目标主机 Agent
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

// GetAgentStatus 查询 Agent 当前状态
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

// ReceiveAgentHeartbeat 接收 Agent 心跳（不经过 Auth 中间件）
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

// GetAgentList 获取 Agent 管理列表（带筛选分页）
func (c *AgentController) GetAgentList(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	query := make(map[string]interface{})
	if hostname := ctx.Query("hostname"); hostname != "" {
		query["hostname"] = hostname
	}
	if ip := ctx.Query("ip"); ip != "" {
		query["ip"] = ip
	}
	if status := ctx.Query("agentStatus"); status != "" {
		query["agentStatus"] = status
	}

	cmdbSvc := c.cmdbSvc
	servers, total, err := cmdbSvc.GetServers(query, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  servers,
		"total": total,
	}))
}

// BatchDeployAgent 批量部署 Agent
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

// BatchUninstallAgent 批量卸载 Agent
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

	for _, id := range req.ServerIDs {
		serverID := id
		go c.svc.UninstallAgent(serverID)
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("批量 Agent 卸载任务已提交"))
}

// DeleteAgentRecord 删除 Agent 记录（仅清空字段，不 SSH）
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

// TestSSHConnection 测试SSH连接
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
		ctx.JSON(http.StatusOK, utils.Response{
			Code:    200,
			Success: false,
			Message: result.Message,
			Data:    result,
		})
	}
}

// ========================================
// Agent 版本管理
// ========================================

// GetAgentVersions 获取 Agent 版本列表
func (c *AgentController) GetAgentVersions(ctx *gin.Context) {
	versions, err := c.svc.GetAllVersions()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(versions))
}

// GetLatestAgentVersion 获取最新 Agent 版本
func (c *AgentController) GetLatestAgentVersion(ctx *gin.Context) {
	version, err := c.svc.GetLatestVersion()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(version))
}

// GetAgentVersionByID 获取指定版本的详细信息
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

// CreateAgentVersion 创建新的 Agent 版本
func (c *AgentController) CreateAgentVersion(ctx *gin.Context) {
	var version modelcmdb.AgentVersion
	if err := ctx.ShouldBindJSON(&version); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
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

// UpdateAgentVersion 更新 Agent 版本信息
func (c *AgentController) UpdateAgentVersion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的版本ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.UpdateVersion(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 版本更新成功"))
}

// DeleteAgentVersion 删除 Agent 版本
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

// UpgradeAgent 升级单台主机的 Agent
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
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.UpgradeAgent(uint(id), req.TargetVersion); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 升级任务已提交"))
}

// GetUpgradeTasks 获取升级任务列表
func (c *AgentController) GetUpgradeTasks(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	status := ctx.Query("status")

	tasks, total, err := c.svc.GetUpgradeTasks(page, pageSize, status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  tasks,
		"total": total,
	}))
}

// GetUpgradeTaskByID 获取升级任务详情
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
