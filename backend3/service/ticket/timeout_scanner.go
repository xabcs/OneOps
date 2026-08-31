package serviceticket

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	modelticket "oneops/backend3/model/ticket"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
)

// 超时升级链参数
const (
	timeoutScanInterval = 10 * time.Minute // 扫描周期
	timeoutReNotifyGap  = 24 * time.Hour   // 超时提醒重复间隔（升级后不再重复）
	escalationFactor    = 2                // 超过阈值 2 倍触发升级
)

// TimeoutScanner 审批超时升级链（蓝图④）
//
// 定时扫描审批中的工单当前节点，按节点快照里的 TimeoutHours 阈值触发：
//   - 停留 ≥ 阈值        → timeout 事件提醒待审批人（每 24h 重复一次）
//   - 停留 ≥ 2 倍阈值    → escalation 事件升级通知管理员与发起人（只发一次）
//
// 防重发状态记录在 ticket_node_records（timeout_notified_at / escalation_notified_at），
// 节点流转到下一节点后记录随之废弃，无需清理。
type TimeoutScanner struct {
	db     *gorm.DB
	notify *TicketNotifier
}

// NewTimeoutScanner 创建扫描器
func NewTimeoutScanner(db *gorm.DB, notify *TicketNotifier) *TimeoutScanner {
	return &TimeoutScanner{db: db, notify: notify}
}

// scanOnce 执行一轮扫描（工单量级小，两步查简单可控）
func (s *TimeoutScanner) scanOnce() {
	// 1. 审批中的工单（current_node_key 非空）
	var tickets []modelticket.Ticket
	if err := s.db.Where("status = ? AND current_node_key != ''", modelticket.TicketStatusPending).
		Find(&tickets).Error; err != nil {
		logger.Warn("超时升级链：查询待审批工单失败", zap.Error(err))
		return
	}
	if len(tickets) == 0 {
		return
	}
	ids := make([]uint, 0, len(tickets))
	keys := make([]string, 0, len(tickets))
	for _, t := range tickets {
		ids = append(ids, t.ID)
		keys = append(keys, t.CurrentNodeKey)
	}

	// 2. 当前节点审批中记录（含节点开始时间与防重发时间戳）
	var records []modelticket.TicketNodeRecord
	if err := s.db.Where("ticket_id IN ? AND node_key IN ? AND status = ?",
		ids, keys, modelticket.NodeStatusPending).Find(&records).Error; err != nil {
		logger.Warn("超时升级链：查询节点记录失败", zap.Error(err))
		return
	}
	recByTicket := make(map[uint]*modelticket.TicketNodeRecord, len(records))
	for i := range records {
		// 一个工单同一时刻只有一个 pending 节点，按当前节点 key 对齐（防止异常数据串节点）
		if records[i].NodeKey != "" {
			recByTicket[records[i].TicketID] = &records[i]
		}
	}

	now := time.Now()
	for i := range tickets {
		t := &tickets[i]
		rec := recByTicket[t.ID]
		if rec == nil || rec.NodeKey != t.CurrentNodeKey || rec.StartedAt == nil {
			continue
		}
		s.checkTicket(t, rec, now)
	}
}

// checkTicket 单工单超时判定与通知
func (s *TimeoutScanner) checkTicket(t *modelticket.Ticket, rec *modelticket.TicketNodeRecord, now time.Time) {
	// 阈值取自发起时的流程快照（流程后续修改不影响在途工单）
	th := snapshotTimeoutHours(t.WorkflowSnapshot, t.CurrentNodeKey)
	if th <= 0 || rec.StartedAt == nil {
		return
	}
	elapsed := now.Sub(*rec.StartedAt)
	if elapsed < time.Duration(th)*time.Hour {
		return
	}

	// 升级阶段：超过 2 倍阈值且未升级过 → 通知管理员与发起人（一次性）
	if elapsed >= time.Duration(th*escalationFactor)*time.Hour && rec.EscalationNotifiedAt == nil {
		if s.notifyEscalation(t, rec, th, elapsed) {
			logger.Info("超时升级链：已升级通知",
				zap.String("ticket_no", t.TicketNo), zap.Int("timeout_hours", th))
		}
		return // 升级后本轮不再重复发超时提醒
	}

	// 超时阶段：首次超时或距上次提醒超过 24h
	if rec.TimeoutNotifiedAt == nil || now.Sub(*rec.TimeoutNotifiedAt) >= timeoutReNotifyGap {
		if s.notifyTimeout(t, rec, th, elapsed) {
			logger.Info("超时升级链：已发送超时提醒",
				zap.String("ticket_no", t.TicketNo), zap.Int("timeout_hours", th))
		}
	}
}

// notifyTimeout 超时提醒待审批人（排除已审过的人）
func (s *TimeoutScanner) notifyTimeout(t *modelticket.Ticket, rec *modelticket.TicketNodeRecord, th int, elapsed time.Duration) bool {
	pending := filterNotApproved(parseUintIDs(rec.ApproverIDs), parseUintIDs(rec.ApprovedIDs))
	if len(pending) == 0 {
		return false
	}
	stay := formatDuration(elapsed)
	subject := fmt.Sprintf("【OneOps 审批超时】%s（%s）", t.Title, t.TicketNo)
	body := fmt.Sprintf("工单「%s」（%s）的节点「%s」已停留 %s，超过 %d 小时阈值：\n\n"+
		"发起人：%s\n优先级：%s\n\n请尽快登录 OneOps 工单中心处理。",
		t.Title, t.TicketNo, rec.NodeName, stay, th, t.CreatorName, priorityText(t.Priority))
	ctx := baseTplCtx(t)
	ctx["node"] = rec.NodeName
	ctx["stay"] = stay
	ctx["threshold"] = fmt.Sprintf("%d", th)
	s.notify.NotifyTimeout(t, pending, subject, body, ctx)
	return s.markNotified(rec, "timeout_notified_at")
}

// notifyEscalation 升级通知：管理员角色用户 + 工单发起人
func (s *TimeoutScanner) notifyEscalation(t *modelticket.Ticket, rec *modelticket.TicketNodeRecord, th int, elapsed time.Duration) bool {
	receivers := s.adminUserIDs()
	if !containsUint(receivers, t.CreatorID) {
		receivers = append(receivers, t.CreatorID)
	}
	if len(receivers) == 0 {
		return false
	}
	stay := formatDuration(elapsed)
	pendingNames := defaultText(rec.ApproverNames, "（无）")
	subject := fmt.Sprintf("【OneOps 审批超时升级】%s（%s）", t.Title, t.TicketNo)
	body := fmt.Sprintf("工单「%s」（%s）的节点「%s」已停留 %s（阈值的 %d 倍），仍未处理，请介入跟进：\n\n"+
		"发起人：%s\n待审批人：%s\n优先级：%s\n\n请登录 OneOps 工单中心处理或改派。",
		t.Title, t.TicketNo, rec.NodeName, stay, escalationFactor,
		t.CreatorName, pendingNames, priorityText(t.Priority))
	s.notify.NotifyEscalation(t, receivers, subject, body, func() map[string]string {
		c := baseTplCtx(t)
		c["node"] = rec.NodeName
		c["stay"] = stay
		c["threshold"] = fmt.Sprintf("%d", th)
		c["approvers"] = pendingNames
		return c
	}())
	ok := s.markNotified(rec, "escalation_notified_at")
	if rec.TimeoutNotifiedAt == nil {
		s.markNotified(rec, "timeout_notified_at") // 升级同时视为已提醒，避免之后再补发超时提醒
	}
	return ok
}

// markNotified 写防重发时间戳（轻量 UPDATE，失败下轮重试）
func (s *TimeoutScanner) markNotified(rec *modelticket.TicketNodeRecord, column string) bool {
	now := time.Now()
	if err := s.db.Model(&modelticket.TicketNodeRecord{}).
		Where("id = ?", rec.ID).
		Update(column, now).Error; err != nil {
		logger.Warn("超时升级链：写入防重发时间戳失败", zap.Uint("record_id", rec.ID), zap.Error(err))
		return false
	}
	// 同步内存值，供同轮后续判断
	if column == "timeout_notified_at" {
		rec.TimeoutNotifiedAt = &now
	} else {
		rec.EscalationNotifiedAt = &now
	}
	return true
}

// adminUserIDs 管理员角色（code=admin 且启用）下的用户
func (s *TimeoutScanner) adminUserIDs() []uint {
	var ids []uint
	if err := s.db.Table("sys_users u").
		Joins("JOIN sys_user_roles ur ON ur.user_id = u.id").
		Joins("JOIN sys_roles r ON r.id = ur.role_id").
		Where("r.code = ? AND r.status = 1 AND u.status = 'active'", "admin").
		Pluck("u.id", &ids).Error; err != nil {
		logger.Warn("超时升级链：查询管理员用户失败", zap.Error(err))
		return nil
	}
	return ids
}

// snapshotTimeoutHours 从工单流程快照解析指定节点的超时阈值
func snapshotTimeoutHours(snapshotJSON, nodeKey string) int {
	if snapshotJSON == "" || nodeKey == "" {
		return 0
	}
	var nodes []WorkflowSnapshotNode
	if err := json.Unmarshal([]byte(snapshotJSON), &nodes); err != nil {
		return 0
	}
	for _, n := range nodes {
		if n.Key == nodeKey {
			return n.TimeoutHours
		}
	}
	return 0
}

// formatDuration 停留时长人话化（如 "26 小时 10 分钟"）
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h == 0 {
		return fmt.Sprintf("%d 分钟", m)
	}
	return fmt.Sprintf("%d 小时 %d 分钟", h, m)
}

// priorityText 优先级文案
func priorityText(p string) string {
	text := map[string]string{
		"low": "低", "normal": "普通", "high": "高", "urgent": "紧急",
	}[p]
	return defaultText(text, p)
}

// StartTicketTimeoutScanner 启动超时升级链扫描（启动延迟 1 分钟首扫，避免与迁移抢库）
func StartTicketTimeoutScanner(notify *TicketNotifier) {
	db := database.GetDB()
	if db == nil || notify == nil {
		return
	}
	sc := NewTimeoutScanner(db, notify)
	go func() {
		time.Sleep(time.Minute)
		sc.scanOnce()
		ticker := time.NewTicker(timeoutScanInterval)
		defer ticker.Stop()
		for range ticker.C {
			sc.scanOnce()
		}
	}()
	logger.Info("超时升级链扫描器已启动", zap.String("interval", timeoutScanInterval.String()))
}
