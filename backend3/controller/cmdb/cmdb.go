package cmdb

import (
	"net/http"
	"strconv"
	"strings"

	. "oneops/backend3/service/cmdb"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/middleware"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
)

// CMDBController CMDB控制器
type CMDBController struct {
	svc      *CMDBService
	agentSvc *AgentService
}

// NewCMDBController 创建CMDB控制器
func NewCMDBController(svc *CMDBService, agentSvc *AgentService) *CMDBController {
	return &CMDBController{
		svc:      svc,
		agentSvc: agentSvc,
	}
}

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
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
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
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
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

// ========== 机房机柜管理 ==========

// GetServerRooms 获取机房列表
func (c *CMDBController) GetServerRooms(ctx *gin.Context) {
	rooms, err := c.svc.GetServerRooms()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(rooms))
}

// CreateServerRoom 创建机房
func (c *CMDBController) CreateServerRoom(ctx *gin.Context) {
	var room modelcmdb.ServerRoom
	if err := ctx.ShouldBindJSON(&room); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.CreateServerRoom(&room); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("机房创建成功"))
}

// UpdateServerRoom 更新机房
func (c *CMDBController) UpdateServerRoom(ctx *gin.Context) {
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

	if err := c.svc.UpdateServerRoom(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("机房更新成功"))
}

// DeleteServerRoom 删除机房
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

// GetCabinets 获取机柜列表
func (c *CMDBController) GetCabinets(ctx *gin.Context) {
	roomID, _ := strconv.ParseUint(ctx.Query("roomId"), 10, 32)

	cabinets, err := c.svc.GetCabinets(uint(roomID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(cabinets))
}

// ========== 标签管理 ==========

// GetServerTags 获取服务器标签列表
func (c *CMDBController) GetServerTags(ctx *gin.Context) {
	tags, err := c.svc.GetServerTags()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(tags))
}

// CreateServerTag 创建服务器标签
func (c *CMDBController) CreateServerTag(ctx *gin.Context) {
	var tag modelcmdb.ServerTag
	if err := ctx.ShouldBindJSON(&tag); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.CreateServerTag(&tag); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签创建成功"))
}

// UpdateServerTag 更新服务器标签
func (c *CMDBController) UpdateServerTag(ctx *gin.Context) {
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

	if err := c.svc.UpdateServerTag(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签更新成功"))
}

// DeleteServerTag 删除服务器标签
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

// AssignServerTag 为服务器分配标签
func (c *CMDBController) AssignServerTag(ctx *gin.Context) {
	var req struct {
		ServerID uint `json:"serverId" binding:"required"`
		TagID    uint `json:"tagId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.AssignServerTag(req.ServerID, req.TagID); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签分配成功"))
}

// RemoveServerTag 移除服务器标签
func (c *CMDBController) RemoveServerTag(ctx *gin.Context) {
	tagID, _ := strconv.ParseUint(ctx.Param("tagId"), 10, 32)
	serverID, _ := strconv.ParseUint(ctx.Param("serverId"), 10, 32)

	if err := c.svc.RemoveServerTag(uint(serverID), uint(tagID)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("标签移除成功"))
}

// ========== 资产变更记录 ==========

// GetAssetChanges 获取资产变更记录
func (c *CMDBController) GetAssetChanges(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	assetType := ctx.Query("assetType")
	assetID, _ := strconv.ParseUint(ctx.Query("assetId"), 10, 32)

	changes, total, err := c.svc.GetAssetChanges(assetType, uint(assetID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  changes,
		"total": total,
	}))
}

// SyncServerMetrics 手动触发单台主机指标采集（仅通过 Agent HTTP 拉取）
func (c *CMDBController) SyncServerMetrics(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	server, err := c.svc.GetServerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("服务器不存在"))
		return
	}

	// 检查 Agent 是否运行
	if server.AgentStatus != "running" {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("Agent 未运行，无法采集指标"))
		return
	}

	// 异步拉取 Agent 指标
	go c.agentSvc.PullMetrics(uint(id))

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("Agent 指标拉取任务已提交"))
}

// ========== 统计与配置 ==========

// GetServerStats 获取服务器统计信息
func (c *CMDBController) GetServerStats(ctx *gin.Context) {
	stats, err := c.svc.GetServerStats()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// GetServerConfig 获取服务器配置信息（通过SSH）
func (c *CMDBController) GetServerConfig(ctx *gin.Context) {
	var req struct {
		Hostname string `json:"hostname" binding:"required"`
		IP       string `json:"ip" binding:"required"`
		SSHUser  string `json:"sshUser"`
		SSHPort  int    `json:"sshPort"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	// 设置默认值
	if req.SSHUser == "" {
		req.SSHUser = "root"
	}
	if req.SSHPort == 0 {
		req.SSHPort = 22
	}

	config, err := c.svc.GetServerConfig(req.Hostname, req.IP, req.SSHUser, req.SSHPort)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取服务器配置失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(config))
}

// ========== 主机分组管理 ==========

// GetServerGroups 获取主机分组列表（树形结构）
func (c *CMDBController) GetServerGroups(ctx *gin.Context) {
	groups, err := c.svc.GetServerGroups()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(groups))
}

// GetAssetTree 获取完整的资产树（分组+服务器），一次性返回所有数据
func (c *CMDBController) GetAssetTree(ctx *gin.Context) {
	data, err := c.svc.GetAssetTree()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(data))
}

// GetServerGroupByID 获取主机分组详情
func (c *CMDBController) GetServerGroupByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	group, err := c.svc.GetServerGroupByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("分组不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(group))
}

// CreateServerGroup 创建主机分组
func (c *CMDBController) CreateServerGroup(ctx *gin.Context) {
	var group modelcmdb.ServerGroup
	if err := ctx.ShouldBindJSON(&group); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.CreateServerGroup(&group); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组创建成功"))
}

// UpdateServerGroup 更新主机分组
func (c *CMDBController) UpdateServerGroup(ctx *gin.Context) {
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

	if err := c.svc.UpdateServerGroup(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组更新成功"))
}

// DeleteServerGroup 删除主机分组
func (c *CMDBController) DeleteServerGroup(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的ID"))
		return
	}

	if err := c.svc.DeleteServerGroup(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("分组删除成功"))
}

// AssignServerToGroup 将服务器分配到分组
func (c *CMDBController) AssignServerToGroup(ctx *gin.Context) {
	var req struct {
		ServerID uint `json:"serverId" binding:"required"`
		GroupID  uint `json:"groupId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.AssignServerToGroup(req.ServerID, req.GroupID); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器分配成功"))
}

// AssignServerToGroups 将服务器分配到多个分组
func (c *CMDBController) AssignServerToGroups(ctx *gin.Context) {
	var req struct {
		ServerID uint   `json:"serverId" binding:"required"`
		GroupIDs []uint `json:"groupIds" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.svc.AssignServerToGroups(req.ServerID, req.GroupIDs); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("服务器分配成功"))
}

// GetServersByGroup 获取指定分组下的服务器列表
func (c *CMDBController) GetServersByGroup(ctx *gin.Context) {
	groupIDStr := ctx.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的分组ID"))
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	servers, err := c.svc.GetServersByGroup(uint(groupID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	// 手动分页
	total := int64(len(servers))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(servers) {
		servers = []modelcmdb.Server{}
	} else if end > len(servers) {
		servers = servers[start:]
	} else {
		servers = servers[start:end]
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  servers,
		"total": total,
	}))
}

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
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
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
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
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
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	result, err := c.svc.TestSSHCredential(uint(id), req.TestIP, req.TestPort)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(result))
}
