package k8s

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	k8ssvc "oneops/backend3/service/k8s"
	syssvc "oneops/backend3/service/system"

	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"

	modelk8s "oneops/backend3/model/k8s"
	"oneops/backend3/pkg/database"
)

// 诊断终端空闲超时：10 分钟无输入自动结束
const diagnosticIdleTimeout = 10 * time.Minute

// 会话 I/O 录制上限（超出后滚动丢弃最早内容）
const diagnosticIOLogLimit = 2 << 20 // 2MB

// DiagnosticTerminalMessage 诊断终端消息协议（前端 ⇄ OneOps）
type DiagnosticTerminalMessage struct {
	Type string `json:"type"` // input / output / resize / closed
	Data string `json:"data"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// DiagnosticWSController 诊断专家终端 WebSocket 处理器
//
// 双桥架构：浏览器 WS ⇄ OneOps ⇄ tunnel-server WS（arthas telnet 协议透传）
// 职责：会话互斥（每 agent 一个活跃会话）、空闲超时、I/O 录制、断开自动 stop（清理字节码增强）
type DiagnosticWSController struct {
	svc         *k8ssvc.DiagnosticService
	db          *gorm.DB
	upgrader    websocket.Upgrader
	activeConns map[uint]*diagnosticConn // sessionID -> 连接
	connsMutex  sync.RWMutex
}

// diagnosticConn 内存态诊断会话
type diagnosticConn struct {
	sessionID uint
	cancel    chan struct{}
	once      sync.Once
}

// NewDiagnosticWSController 创建诊断终端控制器
func NewDiagnosticWSController(svc *k8ssvc.DiagnosticService) *DiagnosticWSController {
	return &DiagnosticWSController{
		svc: svc,
		db:  database.GetDB(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		activeConns: make(map[uint]*diagnosticConn),
	}
}

// HandleSessionWS godoc
// @Summary      诊断专家终端 WebSocket
// @Description  与指定 agent 建立 Arthas 交互会话（经 tunnel-server 透传）。同一 agent 同时仅允许一个活跃会话。
// @Tags         K8s-诊断
// @Produce      json
// @Param        token     query  string  true  "JWT token"
// @Param        agentId   query  string  true  "agent 标识（POD_NAME-NAMESPACE）"
// @Param        clusterId query string  false "集群 ID（用于集群级权限校验）"
// @Success      101  {string} string "升级为 WebSocket 连接"
// @Failure      200  {object} utils.Response "缺少参数 / 无权限 / agent 离线 / 会话被占用"
// @Router       /k8s/diagnostic/session/ws [get]
func (ctrl *DiagnosticWSController) HandleSessionWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "缺少认证token"})
		return
	}
	claims, err := utils.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "token验证失败"})
		return
	}
	userID := claims.UserID

	// 用户状态校验（WS 不走 Auth 中间件，自查）
	var userStatus string
	if err := ctrl.db.Table("sys_users").Select("status").
		Where("id = ?", userID).Scan(&userStatus).Error; err != nil || userStatus != "active" {
		c.JSON(http.StatusOK, gin.H{"code": 401, "success": false, "message": "用户已被禁用或不存在"})
		return
	}

	// 系统权限码校验（fail-closed）
	permSvc, err := syssvc.GetPermissionService()
	if err == nil {
		hasPerm, permErr := permSvc.HasPermission(userID, "k8s.diagnostic.execute")
		if permErr != nil || !hasPerm {
			c.JSON(http.StatusOK, gin.H{"code": 403, "success": false, "message": "无诊断执行权限（k8s.diagnostic.execute）"})
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 500, "success": false, "message": "权限服务不可用"})
		return
	}

	agentID := c.Query("agentId")
	clusterIDStr := c.Query("clusterId")
	if agentID == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "success": false, "message": "缺少 agentId"})
		return
	}

	agent, err := ctrl.svc.GetAgent(agentID)
	if err != nil || agent == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "success": false, "message": "agent 不存在或未注册"})
		return
	}
	if !agent.Online {
		c.JSON(http.StatusOK, gin.H{"code": 409, "success": false,
			"message": "agent 离线，无法建立会话（若此前执行过 stop：Arthas 已被关闭且不会自愈，需重启该 Pod 恢复诊断）"})
		return
	}

	clusterID := agent.ClusterID
	if clusterIDStr != "" {
		if _, parseErr := strconv.ParseUint(clusterIDStr, 10, 32); parseErr == nil {
			clusterID = clusterIDStr
		}
	}

	username := ctrl.lookupUsername(userID)

	// 会话互斥：同一 agent 仅一个活跃会话
	session, err := ctrl.svc.EnsureSessionActive(clusterID, agent.AppName, agentID, userID, username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 409, "success": false, "message": err.Error()})
		return
	}

	// 连接 tunnel（会话 WS 走 7777 端口，非 HTTP 地址）
	sessionURL := ctrl.svc.GetTunnelSessionURL(clusterID)
	tunnelConn, err := ctrl.svc.DialTunnelSession(sessionURL, agentID)
	if err != nil {
		_ = ctrl.svc.CloseSession(session.ID, "closed")
		c.JSON(http.StatusOK, gin.H{"code": 502, "success": false, "message": "连接诊断通道失败: " + err.Error()})
		return
	}

	// 升级浏览器 WS
	browserConn, err := ctrl.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = tunnelConn.Close()
		_ = ctrl.svc.CloseSession(session.ID, "closed")
		logger.Error("升级诊断终端 WebSocket 失败", zap.Error(err))
		return
	}

	conn := &diagnosticConn{sessionID: session.ID, cancel: make(chan struct{})}
	ctrl.connsMutex.Lock()
	ctrl.activeConns[session.ID] = conn
	ctrl.connsMutex.Unlock()

	logger.Info("诊断终端会话已建立",
		zap.Uint("sessionId", session.ID),
		zap.Uint("userId", userID),
		zap.String("app", agent.AppName),
		zap.String("agentId", agentID))

	ctrl.bridge(browserConn, tunnelConn, conn, session, agent)

	// 清理：互斥解除、I/O 落库
	ctrl.connsMutex.Lock()
	delete(ctrl.activeConns, session.ID)
	ctrl.connsMutex.Unlock()
}

// bridge 双向转发与生命周期管理
// 空闲超时由两端 ReadDeadline（diagnosticIdleTimeout）实现：任一端超时即结束会话
func (ctrl *DiagnosticWSController) bridge(browser, tunnel *websocket.Conn, conn *diagnosticConn, session *modelk8s.DiagnosticSession, agent *modelk8s.DiagnosticAgent) {
	var ioMu sync.Mutex
	ioLog := &strings.Builder{}

	recordIO := func(s string) {
		ioMu.Lock()
		defer ioMu.Unlock()
		ioLog.WriteString(s)
		if ioLog.Len() > diagnosticIOLogLimit {
			// 滚动截断：保留后半段
			trimmed := ioLog.String()
			ioLog.Reset()
			ioLog.WriteString(trimmed[len(trimmed)-diagnosticIOLogLimit/2:])
		}
	}

	defer func() {
		conn.once.Do(func() { close(conn.cancel) })

		// 断开自动 stop：结束 arthas 会话并还原字节码增强（方案要求）
		_ = tunnel.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(time.Second))
		_ = browser.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session closed"), time.Now().Add(time.Second))
		_ = tunnel.Close()
		_ = browser.Close()

		_ = ctrl.svc.CloseSession(session.ID, "closed")

		ioMu.Lock()
		logCopy := ioLog.String()
		ioMu.Unlock()
		ctrl.svc.UpdateSessionIOLog(session.ID, logCopy)

		logger.Info("诊断终端会话已结束", zap.Uint("sessionId", session.ID), zap.String("agentId", agent.AgentID))
	}()

	// 输入行检测：累积字符，遇到回车成行 → 命令审计入库
	var lineBuf strings.Builder
	recordCommand := func(line string) {
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}
		_, _, disabled := ctrl.svc.ClassifyCommand(line)
		if disabled {
			return
		}
		ctrl.svc.RecordTerminalCommand(session, line, "", 0, true)
	}

	// 下行：tunnel → browser
	go func() {
		for {
			select {
			case <-conn.cancel:
				return
			default:
			}
			_ = tunnel.SetReadDeadline(time.Now().Add(diagnosticIdleTimeout))
			msgType, data, err := tunnel.ReadMessage()
			if err != nil {
				conn.once.Do(func() { close(conn.cancel) })
				return
			}
			if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
				recordIO(string(data))
				_ = browser.WriteMessage(websocket.TextMessage, data)
			}
		}
	}()

	// 上行：browser → tunnel
	for {
		select {
		case <-conn.cancel:
			return
		default:
		}
		_ = browser.SetReadDeadline(time.Now().Add(diagnosticIdleTimeout))
		var msg DiagnosticTerminalMessage
		if err := browser.ReadJSON(&msg); err != nil {
			return
		}

		switch msg.Type {
		case "input":
			recordIO("\n> " + msg.Data)
			// 行检测（xterm onData 逐字符，含 \r 结尾）
			for _, r := range msg.Data {
				if r == '\r' || r == '\n' {
					recordCommand(lineBuf.String())
					lineBuf.Reset()
				} else {
					lineBuf.WriteRune(r)
				}
			}
			// Arthas WS 输入协议是 JSON（termd HttpTtyConnection.writeToDecoder）：
			// {"action":"read","data":"<输入>"}；纯文本会被 agent 端 JSON 解析失败
			// 静默丢弃，表现为终端不能键入
			inputPayload, _ := json.Marshal(map[string]string{"action": "read", "data": msg.Data})
			if err := tunnel.WriteMessage(websocket.TextMessage, inputPayload); err != nil {
				return
			}
		case "resize":
			// termd 协议支持 resize：{"action":"resize","cols":..,"rows":..}
			// 让 thread/dashboard 等表格按终端宽度渲染，避免固定 80 列折行
			if msg.Cols > 0 && msg.Rows > 0 {
				resizePayload, _ := json.Marshal(map[string]interface{}{"action": "resize", "cols": msg.Cols, "rows": msg.Rows})
				_ = tunnel.WriteMessage(websocket.TextMessage, resizePayload)
			}
		case "ping":
			_ = browser.WriteJSON(DiagnosticTerminalMessage{Type: "pong"})
		}
	}
}

// lookupUsername 查询用户名
func (ctrl *DiagnosticWSController) lookupUsername(userID uint) string {
	var username string
	if err := ctrl.db.Table("sys_users").Select("username").
		Where("id = ?", userID).Scan(&username).Error; err != nil || username == "" {
		username = fmt.Sprintf("user-%d", userID)
	}
	return username
}

// GetSessions godoc
// @Summary      获取诊断活跃会话列表
// @Tags         K8s-诊断
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/sessions [get]
func (ctrl *DiagnosticWSController) GetSessions(c *gin.Context) {
	sessions, err := ctrl.svc.GetActiveSessions()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": sessions})
}

// TerminateSession godoc
// @Summary      强制断开诊断会话（管理员熔断）
// @Tags         K8s-诊断
// @Param        sessionId path int true "会话 ID"
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /k8s/diagnostic/sessions/{sessionId}/terminate [post]
func (ctrl *DiagnosticWSController) TerminateSession(c *gin.Context) {
	sessionIDStr := c.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的会话ID"))
		return
	}

	ctrl.connsMutex.RLock()
	conn, ok := ctrl.activeConns[uint(sessionID)]
	ctrl.connsMutex.RUnlock()

	if !ok {
		// 内存无连接（进程重启后残留 DB 状态）：直接标记 terminated
		if err := ctrl.svc.CloseSession(uint(sessionID), "terminated"); err != nil {
			c.JSON(http.StatusOK, utils.ErrorInternal(err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "会话已标记终止"})
		return
	}

	conn.once.Do(func() { close(conn.cancel) })
	_ = ctrl.svc.CloseSession(uint(sessionID), "terminated")

	username := c.GetString("username")
	logger.Info("诊断会话被强制终止",
		zap.Uint("sessionId", uint(sessionID)),
		zap.String("operator", username))

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "会话已断开"})
}
