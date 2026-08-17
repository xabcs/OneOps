package cmdb

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	. "oneops/backend3/service/cmdb"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

// BastionController 堡垒机控制器
type BastionController struct {
	svc *BastionService
}

// NewBastionController 创建堡垒机控制器
func NewBastionController(svc *BastionService) *BastionController {
	return &BastionController{
		svc: svc,
	}
}

// ========== 连接相关 ==========

// ConnectServer godoc
// @Summary      连接服务器
// @Description  创建到指定服务器的 SSH 会话，返回会话 ID 和 WebSocket 地址
// @Tags         CMDB-堡垒机
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "服务器 ID"
// @Param        body  body      modelcmdb.ConnectRequest   true  "连接请求（凭证 ID、协议等）"
// @Success      200   {object}  utils.Response{data=modelcmdb.ConnectResponse}
// @Failure      200   {object}  utils.Response  "无效的服务器 ID / 请求参数错误 / 未认证 / 没有连接权限 / 创建会话失败"
// @Router       /cmdb/servers/{id}/connect [post]
// @Security     BearerAuth
func (c *BastionController) ConnectServer(ctx *gin.Context) {
	// 获取服务器ID
	serverIDStr := ctx.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	// 解析连接请求
	var req modelcmdb.ConnectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 获取用户信息
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	// 获取客户端IP
	clientIP := ctx.ClientIP()

	// 检查权限并创建会话
	session, err := c.svc.CreateSSHSession(userID, uint(serverID), req.CredentialID, clientIP, req.Protocol)
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
	response := modelcmdb.ConnectResponse{
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

	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	serverIDUint := uint(serverID)
	filter := modelcmdb.SessionFilter{
		ServerID: &serverIDUint,
	}

	sessions, total, err := c.svc.GetSessions(filter, params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(sessions, total, params)))
}

// ========== 会话管理 ==========

// GetSessions godoc
// @Summary      获取会话列表
// @Description  分页获取堡垒机会话列表，支持多维筛选（服务器、用户、状态、协议、客户端 IP、登录账号、时间范围等）
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        page         query     int     false  "页码"          default(1)
// @Param        pageSize     query     int     false  "每页数量"      default(10)
// @Param        serverId     query     int     false  "服务器 ID"
// @Param        userId       query     int     false  "用户 ID"
// @Param        status       query     string  false  "状态"
// @Param        protocol     query     string  false  "协议"
// @Param        clientIp     query     string  false  "客户端 IP"
// @Param        loginAccount query     string  false  "登录账号"
// @Param        startDate    query     string  false  "开始日期"
// @Param        endDate      query     string  false  "结束日期"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取会话列表失败"
// @Router       /cmdb/sessions [get]
// @Security     BearerAuth
func (c *BastionController) GetSessions(ctx *gin.Context) {
	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 构建筛选条件
	filter := modelcmdb.SessionFilter{}

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

	sessions, total, err := c.svc.GetSessions(filter, params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(sessions, total, params)))
}

// GetSessionsList godoc
// @Summary      获取会话列表（轻量级）
// @Description  分页获取会话列表，只返回列表展示所需字段，避免加载敏感信息和冗余数据，性能更优
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        page         query     int     false  "页码"          default(1)
// @Param        pageSize     query     int     false  "每页数量"      default(10)
// @Param        serverId     query     int     false  "服务器 ID"
// @Param        userId       query     int     false  "用户 ID"
// @Param        status       query     string  false  "状态"
// @Param        protocol     query     string  false  "协议"
// @Param        clientIp     query     string  false  "客户端 IP"
// @Param        loginAccount query     string  false  "登录账号"
// @Param        startDate    query     string  false  "开始日期"
// @Param        endDate      query     string  false  "结束日期"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取会话列表失败"
// @Router       /cmdb/sessions/list [get]
// @Security     BearerAuth
func (c *BastionController) GetSessionsList(ctx *gin.Context) {
	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 构建筛选条件
	filter := modelcmdb.SessionFilter{}

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

	sessions, total, err := c.svc.GetSessionsList(filter, params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(sessions, total, params)))
}

// GetSessionByID godoc
// @Summary      获取会话详情
// @Description  根据会话 ID 获取会话详细信息
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "会话 ID"
// @Success      200  {object}  utils.Response  "会话详情"
// @Failure      200  {object}  utils.Response  "无效的会话 ID / 会话不存在"
// @Router       /cmdb/sessions/{id} [get]
// @Security     BearerAuth
func (c *BastionController) GetSessionByID(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	session, err := c.svc.GetSessionByID(uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("会话不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(session))
}

// TerminateSession godoc
// @Summary      强制断开会话
// @Description  强制断开指定的堡垒机会话
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "会话 ID"
// @Success      200  {object}  utils.Response  "会话已断开"
// @Failure      200  {object}  utils.Response  "无效的会话 ID / 未认证 / 断开失败"
// @Router       /cmdb/sessions/{id}/terminate [post]
// @Security     BearerAuth
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

	// 先终止内存中的 WebSocket + SSH 连接（如果存在）
	if sm := GetSessionManager(); sm != nil {
		sm.TerminateSession(uint(sessionID))
	}

	// 更新数据库状态
	if err := c.svc.TerminateSession(uint(sessionID), operatorID); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("会话已断开"))
}

// GetActiveSessions godoc
// @Summary      获取活跃会话列表
// @Description  获取数据库中状态为活跃的会话列表
// @Tags         CMDB-堡垒机
// @Produce      json
// @Success      200  {object}  utils.Response  "活跃会话列表"
// @Failure      200  {object}  utils.Response  "获取活跃会话失败"
// @Router       /cmdb/sessions/active [get]
// @Security     BearerAuth
func (c *BastionController) GetActiveSessions(ctx *gin.Context) {
	sessions, err := c.svc.GetActiveSessions()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(sessions))
}

// GetActiveSessionsFromMemory godoc
// @Summary      获取内存中的活跃会话
// @Description  返回当前 WebSocket 连接仍然存在的会话列表（比查询数据库更准确），并实时计算会话持续时长
// @Tags         CMDB-堡垒机
// @Produce      json
// @Success      200  {object}  utils.Response  "活跃会话列表（含实时持续时长）"
// @Failure      200  {object}  utils.Response  "获取活跃会话失败"
// @Router       /cmdb/sessions/active-memory [get]
// @Security     BearerAuth
func (c *BastionController) GetActiveSessionsFromMemory(ctx *gin.Context) {
	sm := GetSessionManager()
	if sm == nil {
		ctx.JSON(http.StatusOK, utils.SuccessWithData([]interface{}{}))
		return
	}

	activeStates := sm.GetAllActiveSessions()
	now := time.Now()

	result := make([]gin.H, 0, len(activeStates))
	for _, s := range activeStates {
		result = append(result, gin.H{
			"sessionId":    s.SessionID,
			"userId":       s.UserID,
			"username":     s.Username,
			"serverName":   s.ServerName,
			"serverIp":     s.ServerIP,
			"loginAccount": s.LoginAccount,
			"createdAt":    s.CreatedAt,
			"lastActiveAt": s.LastActiveAt,
			"duration":     int(now.Sub(s.CreatedAt).Seconds()),
		})
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(result))
}

// GetSessionCommands godoc
// @Summary      获取会话命令列表
// @Description  获取指定会话的命令审计记录列表
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "会话 ID"
// @Success      200  {object}  utils.Response  "命令列表"
// @Failure      200  {object}  utils.Response  "无效的会话 ID / 获取命令列表失败"
// @Router       /cmdb/sessions/{id}/commands [get]
// @Security     BearerAuth
func (c *BastionController) GetSessionCommands(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	commands, err := c.svc.GetSessionCommands(uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(commands))
}

// GetCommands godoc
// @Summary      获取命令审计列表
// @Description  分页获取命令审计记录，支持按会话、风险等级、是否拦截、命令内容、时间范围筛选
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        page       query     int     false  "页码"            default(1)
// @Param        pageSize   query     int     false  "每页数量"        default(10)
// @Param        sessionId  query     int     false  "会话 ID"
// @Param        riskLevel  query     string  false  "风险等级"
// @Param        blocked    query     boolean false  "是否被拦截"
// @Param        command    query     string  false  "命令内容（模糊匹配）"
// @Param        startDate  query     string  false  "开始日期"
// @Param        endDate    query     string  false  "结束日期"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取命令列表失败"
// @Router       /cmdb/commands [get]
// @Security     BearerAuth
func (c *BastionController) GetCommands(ctx *gin.Context) {
	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 构建筛选条件
	filter := modelcmdb.CommandFilter{}

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

	commands, total, err := c.svc.GetCommands(filter, params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(commands, total, params)))
}

// GetSessionFileTransfers godoc
// @Summary      获取会话文件传输记录
// @Description  获取指定会话的文件传输记录列表
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "会话 ID"
// @Success      200  {object}  utils.Response  "文件传输记录列表"
// @Failure      200  {object}  utils.Response  "无效的会话 ID / 获取文件传输记录失败"
// @Router       /cmdb/sessions/{id}/file-transfers [get]
// @Security     BearerAuth
func (c *BastionController) GetSessionFileTransfers(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	transfers, err := c.svc.GetSessionFileTransfers(uint(sessionID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(transfers))
}

// GetFileTransfers godoc
// @Summary      获取文件传输列表
// @Description  分页获取文件传输审计记录，支持按会话、方向、状态、时间范围筛选
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        page       query     int     false  "页码"        default(1)
// @Param        pageSize   query     int     false  "每页数量"    default(10)
// @Param        sessionId  query     int     false  "会话 ID"
// @Param        direction  query     string  false  "传输方向（upload/download）"
// @Param        status     query     string  false  "状态"
// @Param        startDate  query     string  false  "开始日期"
// @Param        endDate    query     string  false  "结束日期"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取文件传输列表失败"
// @Router       /cmdb/file-transfers [get]
// @Security     BearerAuth
func (c *BastionController) GetFileTransfers(ctx *gin.Context) {
	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	// 构建筛选条件
	filter := modelcmdb.FileTransferFilter{}

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

	transfers, total, err := c.svc.GetFileTransfers(filter, params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(transfers, total, params)))
}

// ResizeTerminalPTY godoc
// @Summary      调整终端大小
// @Description  调整指定会话的 SSH PTY 窗口大小（通过 SessionManager 获取活跃 SSH 会话发送 window-change 请求）
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "会话 ID"
// @Success      200  {object}  utils.Response  "提示信息"
// @Router       /cmdb/sessions/{id}/resize [post]
// @Security     BearerAuth
func (c *BastionController) ResizeTerminalPTY(ctx *gin.Context) {
	sessionIDStr := ctx.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	var req struct {
		Rows uint `json:"rows" binding:"required"`
		Cols uint `json:"cols" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	sm := GetSessionManager()
	if sm == nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("SessionManager 未初始化"))
		return
	}

	sshSession := sm.GetSSHSession(uint(sessionID))
	if sshSession == nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("会话不存在或已断开"))
		return
	}

	type windowChangeMsg struct {
		Columns uint32
		Rows    uint32
		Width   uint32
		Height  uint32
	}

	msg := windowChangeMsg{
		Columns: uint32(req.Cols),
		Rows:    uint32(req.Rows),
		Width:   uint32(req.Cols * 8),
		Height:  uint32(req.Rows * 16),
	}

	ok, err := sshSession.SendRequest("window-change", false, ssh.Marshal(&msg))
	if err != nil || !ok {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("调整终端大小失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("终端大小已调整"))
}

// ========== 访问策略管理 ==========

// GetAccessPolicies godoc
// @Summary      获取访问策略列表
// @Description  分页获取访问策略列表
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        page      query  int  false  "页码"      default(1)
// @Param        pageSize  query  int  false  "每页数量"  default(10)
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取访问策略列表失败"
// @Router       /cmdb/access-policies [get]
// @Security     BearerAuth
func (c *BastionController) GetAccessPolicies(ctx *gin.Context) {
	var params dto.BasePageQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	policies, total, err := c.svc.GetAccessPolicies(params.GetPage(), params.GetPageSize())
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(policies, total, params)))
}

// CreateAccessPolicy godoc
// @Summary      创建访问策略
// @Description  创建一个新的访问策略
// @Tags         CMDB-堡垒机
// @Accept       json
// @Produce      json
// @Param        policy  body      modelcmdb.AssetAccessPolicy  true  "访问策略信息"
// @Success      200     {object}  utils.Response  "访问策略创建成功"
// @Failure      200     {object}  utils.Response  "请求参数错误 / 创建失败"
// @Router       /cmdb/access-policies [post]
// @Security     BearerAuth
func (c *BastionController) CreateAccessPolicy(ctx *gin.Context) {
	var policy modelcmdb.AssetAccessPolicy
	if err := ctx.ShouldBindJSON(&policy); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.CreateAccessPolicy(&policy); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("访问策略创建成功"))
}

// UpdateAccessPolicy godoc
// @Summary      更新访问策略
// @Description  根据策略 ID 更新访问策略信息（部分字段更新）
// @Tags         CMDB-堡垒机
// @Accept       json
// @Produce      json
// @Param        id      path      int                          true  "策略 ID"
// @Param        policy  body      modelcmdb.AssetAccessPolicy  true  "需要更新的字段"
// @Success      200     {object}  utils.Response  "访问策略更新成功"
// @Failure      200     {object}  utils.Response  "无效的策略 ID / 请求参数错误 / 更新失败"
// @Router       /cmdb/access-policies/{id} [put]
// @Security     BearerAuth
func (c *BastionController) UpdateAccessPolicy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的策略ID"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	if err := c.svc.UpdateAccessPolicy(uint(id), updates); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("访问策略更新成功"))
}

// DeleteAccessPolicy godoc
// @Summary      删除访问策略
// @Description  根据策略 ID 删除访问策略
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "策略 ID"
// @Success      200  {object}  utils.Response  "访问策略删除成功"
// @Failure      200  {object}  utils.Response  "无效的策略 ID / 删除失败"
// @Router       /cmdb/access-policies/{id} [delete]
// @Security     BearerAuth
func (c *BastionController) DeleteAccessPolicy(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的策略ID"))
		return
	}

	if err := c.svc.DeleteAccessPolicy(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("访问策略删除成功"))
}

// GetAccessPolicyByID godoc
// @Summary      获取访问策略详情
// @Description  根据策略 ID 获取访问策略详情
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "策略 ID"
// @Success      200  {object}  utils.Response{data=modelcmdb.AssetAccessPolicy}
// @Failure      200  {object}  utils.Response  "无效的策略 ID / 策略不存在"
// @Router       /cmdb/access-policies/{id} [get]
// @Security     BearerAuth
func (c *BastionController) GetAccessPolicyByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的策略ID"))
		return
	}

	policy, err := c.svc.GetAccessPolicyByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("策略不存在"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(policy))
}

// ========== 统计信息 ==========

// GetSessionStats godoc
// @Summary      获取会话统计信息
// @Description  获取当前活跃会话数和今日会话数等统计信息
// @Tags         CMDB-堡垒机
// @Produce      json
// @Success      200  {object}  utils.Response  "会话统计信息"
// @Router       /cmdb/sessions/stats [get]
// @Security     BearerAuth
func (c *BastionController) GetSessionStats(ctx *gin.Context) {
	stats := make(map[string]interface{})

	// 内存中真实活跃的 WebSocket 会话数
	if sm := GetSessionManager(); sm != nil {
		stats["active"] = sm.GetActiveSessionCount()
	}

	// 今日会话数（用足够大的 pageSize 确保 total 准确）
	today := time.Now().Format("2006-01-02")
	todayFilter := modelcmdb.SessionFilter{
		StartDate: &today,
	}
	_, todayTotal, _ := c.svc.GetSessions(todayFilter, 1, 1000)
	stats["today"] = todayTotal

	ctx.JSON(http.StatusOK, utils.SuccessWithData(stats))
}

// CheckConnectPermission godoc
// @Summary      检查连接权限
// @Description  检查当前用户是否有权限连接指定服务器，并返回可用凭证列表
// @Tags         CMDB-堡垒机
// @Produce      json
// @Param        id  path  int  true  "服务器 ID"
// @Success      200  {object}  utils.Response  "权限检查结果（含可用凭证列表）"
// @Failure      200  {object}  utils.Response  "无效的服务器 ID / 未认证 / 检查失败"
// @Router       /cmdb/servers/{id}/permission [get]
// @Security     BearerAuth
func (c *BastionController) CheckConnectPermission(ctx *gin.Context) {
	serverIDStr := ctx.Param("id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的服务器ID"))
		return
	}

	// 获取用户信息
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未认证"))
		return
	}

	// 检查权限，返回可用凭证列表
	hasPermission, credentials, err := c.svc.CheckConnectPermission(userID, uint(serverID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
		"hasPermission": hasPermission,
		"credentials":   credentials,
	}))
}
