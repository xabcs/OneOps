package cmdb

import (
	"net/http"
	"oneops/backend2/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// BastionController 堡垒机控制器
type BastionController struct {
	bastionService *cmdb.BastionService
	sessionManager *cmdb.SessionManager
}

// NewBastionController 创建堡垒机控制器
func NewBastionController() *BastionController {
	return &BastionController{
		bastionService: cmdb.NewBastionService(),
		sessionManager: handler.GetSessionManager(),
	}
}

// ========== 连接相关 ==========

// ConnectServer 连接服务器
func (c *BastionController) ConnectServer(ctx *gin.Context) {
	// 添加调试日志

	// 获取服务器ID
	serverIDStr := ctx.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	// 解析连接请求
	var req cmdb.ConnectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	// 获取用户信息
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		println("[DEBUG] userID is 0, not authenticated")
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	println("[DEBUG] userID:", userID, "serverID:", serverID, "credentialID:", req.CredentialID)

	// 获取客户端IP
	clientIP := ctx.ClientIP()

	// 检查权限并创建会话
	session, err := c.bastionService.CreateSSHSession(userID, uint(serverID), req.CredentialID, clientIP, req.Protocol)
	if err != nil {
		// 检查是否是权限错误
		if strings.Contains(err.Error(), "没有访问权限") || strings.Contains(err.Error(), "没有连接权限") {
			ctx.JSON(http.StatusOK, utils.ErrorForbidden("没有连接权限"))
			return
		}
		if strings.Contains(err.Error(), "不允许使用该凭证") {
			ctx.JSON(http.StatusOK, utils.ErrorForbidden(err.Error()))
			return
		}
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	// 返回会话信息
	response := cmdb.ConnectResponse{
		SessionID:    session.ID,
		WebSocketURL: "/api/cmdb/sessions/" + strconv.FormatUint(uint64(session.ID), 10) + "/ws",
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(response))
}

// GetServerSessions 获取服务器的会话列表
func (c *BastionController) GetServerSessions(ctx *gin.Context) {
	serverIDStr := ctx.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	serverIDUint := uint(serverID)
	filter := cmdb.SessionFilter{
		ServerID: &serverIDUint,
	}

	sessions, total, err := c.bastionService.GetSessions(filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  sessions,
		"total": total,
	}))
}

// ========== 会话管理 ==========

// GetSessions 获取会话列表
func (c *BastionController) GetSessions(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 构建筛选条件
	filter := cmdb.SessionFilter{}

	if serverID := ctx.Query("serverId"); serverID != "" {
		if id, err := strconv.ParseUint(serverID, 10, 32); err == nil {
			uid := uint(id)
			filter.ServerID = &uid
		}
	}

	if userID := ctx.Query("userId"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			uid := uint(id)
			filter.UserID = &uid
		}
	}

	if status := ctx.Query("status"); status != "" {
		filter.Status = &status
	}

	if protocol := ctx.Query("protocol"); protocol != "" {
		filter.Protocol = &protocol
	}

	if clientIP := ctx.Query("clientIp"); clientIP != "" {
		filter.ClientIP = &clientIP
	}

	if loginAccount := ctx.Query("loginAccount"); loginAccount != "" {
		filter.LoginAccount = &loginAccount
	}

	if startDate := ctx.Query("startDate"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := ctx.Query("endDate"); endDate != "" {
		filter.EndDate = &endDate
	}

	sessions, total, err := c.bastionService.GetSessions(filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  sessions,
		"total": total,
	}))
}

// GetSessionsList 获取会话列表（轻量级，只返回列表展示需要的字段）
// 优化性能，避免加载敏感信息和冗余数据
func (c *BastionController) GetSessionsList(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 构建筛选条件
	filter := cmdb.SessionFilter{}

	if serverID := ctx.Query("serverId"); serverID != "" {
		if id, err := strconv.ParseUint(serverID, 10, 32); err == nil {
			uid := uint(id)
			filter.ServerID = &uid
		}
	}

	if userID := ctx.Query("userId"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			uid := uint(id)
			filter.UserID = &uid
		}
	}

	if status := ctx.Query("status"); status != "" {
		filter.Status = &status
	}

	if protocol := ctx.Query("protocol"); protocol != "" {
		filter.Protocol = &protocol
	}

	if clientIP := ctx.Query("clientIp"); clientIP != "" {
		filter.ClientIP = &clientIP
	}

	if loginAccount := ctx.Query("loginAccount"); loginAccount != "" {
		filter.LoginAccount = &loginAccount
	}

	if startDate := ctx.Query("startDate"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := ctx.Query("endDate"); endDate != "" {
		filter.EndDate = &endDate
	}

	sessions, total, err := c.bastionService.GetSessionsList(filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  sessions,
		"total": total,
	}))
}

// GetSessionByID 获取会话详情
func (c *BastionController) GetSessionByID(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	session, err := c.bastionService.GetSessionByID(uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("会话不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(session))
}

// TerminateSession 强制断开会话
func (c *BastionController) TerminateSession(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	// 获取操作人ID
	operatorID := ctx.GetUint("user_id")
	if operatorID == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	// 1. 先断开 WebSocket 连接（从内存中移除会话）
	GetSessionManager().TerminateSession(uint(sessionID))

	// 2. 更新数据库状态
	if err := c.bastionService.TerminateSession(uint(sessionID), operatorID); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("会话已断开"))
}

// GetActiveSessions 获取活跃会话列表
func (c *BastionController) GetActiveSessions(ctx *gin.Context) {
	sessions, err := c.bastionService.GetActiveSessions()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(sessions))
}

// GetActiveSessionsFromMemory 从内存中获取真正的活跃会话
// 返回当前 WebSocket 连接仍然存在的会话列表（比查询数据库更准确）
func (c *BastionController) GetActiveSessionsFromMemory(ctx *gin.Context) {
	// 1. 从 SessionManager 获取所有活跃会话的 ID
	activeSessionIDs := c.sessionManager.GetAllActiveSessions()

	// 2. 从数据库获取这些会话的完整信息
	var sessions []cmdb.BastionSession
	err := c.bastionService.GetDB().
		Preload("Server").
		Preload("User").
		Preload("SSHCredential").
		Where("id IN ?", activeSessionIDs).
		Order("started_at DESC").
		Find(&sessions).Error

	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	// 3. 实时计算会话持续时长
	now := time.Now()
	for i := range sessions {
		if sessions[i].StartedAt != nil {
			sessions[i].Duration = int(now.Sub(*sessions[i].StartedAt).Seconds())
		}
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(sessions))
}

// GetSessionCommands 获取会话的命令列表
func (c *BastionController) GetSessionCommands(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	commands, err := c.bastionService.GetSessionCommands(uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(commands))
}

// GetCommands 获取命令列表（分页）
func (c *BastionController) GetCommands(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 构建筛选条件
	filter := cmdb.CommandFilter{}

	if sessionID := ctx.Query("sessionId"); sessionID != "" {
		if id, err := strconv.ParseUint(sessionID, 10, 32); err == nil {
			uid := uint(id)
			filter.SessionID = &uid
		}
	}

	if riskLevel := ctx.Query("riskLevel"); riskLevel != "" {
		filter.RiskLevel = &riskLevel
	}

	blocked := ctx.Query("blocked") == "true"
	filter.Blocked = &blocked

	if commandLike := ctx.Query("command"); commandLike != "" {
		filter.CommandLike = &commandLike
	}

	if startDate := ctx.Query("startDate"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := ctx.Query("endDate"); endDate != "" {
		filter.EndDate = &endDate
	}

	commands, total, err := c.bastionService.GetCommands(filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  commands,
		"total": total,
	}))
}

// GetSessionFileTransfers 获取会话的文件传输记录
func (c *BastionController) GetSessionFileTransfers(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	transfers, err := c.bastionService.GetSessionFileTransfers(uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(transfers))
}

// GetFileTransfers 获取文件传输列表（分页）
func (c *BastionController) GetFileTransfers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 构建筛选条件
	filter := cmdb.FileTransferFilter{}

	if sessionID := ctx.Query("sessionId"); sessionID != "" {
		if id, err := strconv.ParseUint(sessionID, 10, 32); err == nil {
			uid := uint(id)
			filter.SessionID = &uid
		}
	}

	if direction := ctx.Query("direction"); direction != "" {
		filter.Direction = &direction
	}

	if status := ctx.Query("status"); status != "" {
		filter.Status = &status
	}

	if startDate := ctx.Query("startDate"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := ctx.Query("endDate"); endDate != "" {
		filter.EndDate = &endDate
	}

	transfers, total, err := c.bastionService.GetFileTransfers(filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  transfers,
		"total": total,
	}))
}

// ResizeTerminalPTY 调整终端大小
func (c *BastionController) ResizeTerminalPTY(ctx *gin.Context) {
	// 这个接口由 WebSocket handler 直接处理
	ctx.JSON(http.StatusOK, utils.ErrorBadRequest("请使用 WebSocket 连接"))
}

// ========== 访问策略管理 ==========

// GetAccessPolicies 获取访问策略列表
func (c *BastionController) GetAccessPolicies(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	policies, total, err := c.bastionService.GetAccessPolicies(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"list":  policies,
		"total": total,
	}))
}

// CreateAccessPolicy 创建访问策略
func (c *BastionController) CreateAccessPolicy(ctx *gin.Context) {
	var policy cmdb.AssetAccessPolicy
	if err := ctx.ShouldBindJSON(&policy); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.bastionService.CreateAccessPolicy(&policy); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("访问策略创建成功"))
}

// UpdateAccessPolicy 更新访问策略
func (c *BastionController) UpdateAccessPolicy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的策略ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	if err := c.bastionService.UpdateAccessPolicy(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("访问策略更新成功"))
}

// DeleteAccessPolicy 删除访问策略
func (c *BastionController) DeleteAccessPolicy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的策略ID"))
		return
	}

	if err := c.bastionService.DeleteAccessPolicy(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("访问策略删除成功"))
}

// GetAccessPolicyByID 获取访问策略详情
func (c *BastionController) GetAccessPolicyByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的策略ID"))
		return
	}

	policy, err := c.bastionService.GetAccessPolicyByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("策略不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(policy))
}

// ========== 统计信息 ==========

// GetSessionStats 获取会话统计信息
func (c *BastionController) GetSessionStats(ctx *gin.Context) {
	stats := make(map[string]interface{})

	// 活跃会话数
	activeSessions, err := c.bastionService.GetActiveSessions()
	if err == nil {
		stats["active"] = len(activeSessions)
	}

	// 今日会话数
	today := time.Now().Format("2006-01-02")
	todayFilter := cmdb.SessionFilter{
		StartDate: &today,
	}
	todaySessions, _, _ := c.bastionService.GetSessions(todayFilter, 1, 1)
	stats["today"] = len(todaySessions)

	ctx.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// CheckConnectPermission 检查连接权限
func (c *BastionController) CheckConnectPermission(ctx *gin.Context) {
	// 添加调试日志
	println("[DEBUG] CheckConnectPermission called")
	println("[DEBUG] Path:", ctx.Request.URL.Path)
	println("[DEBUG] Method:", ctx.Request.Method)
	println("[DEBUG] Authorization:", ctx.GetHeader("Authorization"))

	serverIDStr := ctx.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	// 获取用户信息
	userID := ctx.GetUint("user_id")
	println("[DEBUG] userID:", userID, "serverID:", serverID)
	if userID == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	// 检查权限，返回可用凭证列表
	hasPermission, credentials, err := c.bastionService.CheckConnectPermission(userID, uint(serverID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"hasPermission": hasPermission,
		"credentials":   credentials,
	}))
}
