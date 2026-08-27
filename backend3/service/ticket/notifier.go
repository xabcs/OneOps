package serviceticket

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	modelticket "oneops/backend3/model/ticket"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/service/monitoring/notification"

	"go.uber.org/zap"
)

// TicketNotifier 工单事件通知器
//
// 复用 monitoring 模块的邮件通道（mon_notification_channels 表中 channel_type=email 且启用的渠道），
// 在审批事件（节点激活/终态/改派）发生后向相关人员发送提醒。
// 设计约束：
//   - 所有发送均为异步（goroutine），失败仅记日志，绝不阻塞或回滚审批事务；
//   - 未配置邮件渠道或用户缺邮箱时静默跳过（记 info 日志），通知是尽力而为的增强能力；
//   - 邮件配置在首次使用后缓存 5 分钟，避免每次发送都查库。
type TicketNotifier struct {
	db *gorm.DB

	mu        sync.Mutex
	cachedAt  time.Time
	emailSvc  *notification.EmailService
	cacheInit bool // 是否已尝试加载过（区分"未配置"与"未加载"，避免反复查库）
}

// NewTicketNotifier 创建通知器（db 为 nil 时通知功能自动降级为空操作）
func NewTicketNotifier(db *gorm.DB) *TicketNotifier {
	return &TicketNotifier{db: db}
}

// NotifyApprovers 通知节点待审批人（排除已审过的人）
func (n *TicketNotifier) NotifyApprovers(ticket *modelticket.Ticket, approverIDs []uint, approvedIDs []uint, nodeName string) {
	if n == nil || n.db == nil || len(approverIDs) == 0 {
		return
	}
	pending := filterNotApproved(approverIDs, approvedIDs)
	if len(pending) == 0 {
		return
	}
	subject := fmt.Sprintf("【OneOps 待审批】%s（%s）", ticket.Title, ticket.TicketNo)
	body := fmt.Sprintf("您有一张工单待审批：\n\n"+
		"工单号：%s\n标题：%s\n类型：%s\n当前节点：%s\n发起人：%s\n发起时间：%s\n\n"+
		"请登录 OneOps 工单中心处理。",
		ticket.TicketNo, ticket.Title, ticket.TypeName, nodeName,
		ticket.CreatorName, time.Now().Format("2006-01-02 15:04:05"))
	n.sendAsync(pending, subject, body)
}

// NotifyCreator 通知发起人工单终态（通过/驳回）
func (n *TicketNotifier) NotifyCreator(ticket *modelticket.Ticket, result string, operator, comment string) {
	if n == nil || n.db == nil || ticket.CreatorID == 0 {
		return
	}
	resultText := map[string]string{
		modelticket.TicketStatusApproved: "已审批通过",
		modelticket.TicketStatusRejected: "已被驳回",
	}[result]
	if resultText == "" {
		resultText = result
	}
	subject := fmt.Sprintf("【OneOps 工单%s】%s（%s）", resultText, ticket.Title, ticket.TicketNo)
	body := fmt.Sprintf("您发起的工单%s：\n\n"+
		"工单号：%s\n标题：%s\n类型：%s\n处理人：%s\n\n审批意见：%s\n\n请登录 OneOps 工单中心查看详情。",
		resultText, ticket.TicketNo, ticket.Title, ticket.TypeName, operator,
		defaultText(comment, "（无）"))
	n.sendAsync([]uint{ticket.CreatorID}, subject, body)
}

// NotifyReassigned 通知被改派的新审批人
func (n *TicketNotifier) NotifyReassigned(ticket *modelticket.Ticket, approverIDs []uint, nodeName, operator string) {
	if n == nil || n.db == nil || len(approverIDs) == 0 {
		return
	}
	subject := fmt.Sprintf("【OneOps 待审批】%s（%s）", ticket.Title, ticket.TicketNo)
	body := fmt.Sprintf("工单「%s」（%s）的节点「%s」已由 %s 改派给您审批，请登录 OneOps 工单中心处理。",
		ticket.Title, ticket.TicketNo, nodeName, operator)
	n.sendAsync(approverIDs, subject, body)
}

// NotifyApproversCanceled 工单被发起人撤销时通知当前节点待审人（无需再处理）
func (n *TicketNotifier) NotifyApproversCanceled(ticket *modelticket.Ticket, approverIDs, approvedIDs []uint, nodeName, operator string) {
	if n == nil || n.db == nil {
		return
	}
	pending := filterNotApproved(approverIDs, approvedIDs)
	if len(pending) == 0 {
		return
	}
	subject := fmt.Sprintf("【OneOps 工单已撤销】%s（%s）", ticket.Title, ticket.TicketNo)
	body := fmt.Sprintf("工单「%s」（%s）的节点「%s」已被发起人 %s 撤销，您无需再处理。",
		ticket.Title, ticket.TicketNo, nodeName, operator)
	n.sendAsync(pending, subject, body)
}

// NotifyUrge 催办提醒：通知当前节点待审批人（排除已审过的人）
func (n *TicketNotifier) NotifyUrge(ticket *modelticket.Ticket, approverIDs, approvedIDs []uint, nodeName, operator string) {
	if n == nil || n.db == nil {
		return
	}
	pending := filterNotApproved(approverIDs, approvedIDs)
	if len(pending) == 0 {
		return
	}
	subject := fmt.Sprintf("【OneOps 催办提醒】%s（%s）", ticket.Title, ticket.TicketNo)
	body := fmt.Sprintf("发起人 %s 催办：工单「%s」（%s）的节点「%s」等待您的审批，请尽快登录 OneOps 工单中心处理。",
		operator, ticket.Title, ticket.TicketNo, nodeName)
	n.sendAsync(pending, subject, body)
}

// ────────────────────────── 内部实现 ──────────────────────────

// sendAsync 异步发送：查收件人邮箱 → 加载邮件通道 → 发送，全程失败仅记日志
func (n *TicketNotifier) sendAsync(userIDs []uint, subject, body string) {
	recipients, err := n.findEmails(userIDs)
	if err != nil {
		logger.Warn("工单通知：查询收件人邮箱失败", zap.Error(err))
		return
	}
	if len(recipients) == 0 {
		logger.Info("工单通知：收件人未配置邮箱，跳过", zap.Uint64s("user_ids", u64(userIDs)))
		return
	}

	emailSvc, ok := n.loadEmailService()
	if !ok {
		logger.Info("工单通知：未配置启用的邮件渠道，跳过", zap.String("subject", subject))
		return
	}

	// 拷贝值捕获，避免外层数据被后续修改影响
	svc, to := emailSvc, recipients
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("工单通知：发送 panic", zap.Any("recover", r))
			}
		}()
		if err := svc.Send(to, subject, body, false); err != nil {
			logger.Warn("工单通知：邮件发送失败", zap.Strings("to", to), zap.Error(err))
		}
	}()
}

// findEmails 批量查用户邮箱（状态正常且邮箱非空）
func (n *TicketNotifier) findEmails(userIDs []uint) ([]string, error) {
	var emails []string
	err := n.db.Table("sys_users").
		Select("email").
		Where("id IN ? AND status = 'active' AND email != ''", userIDs).
		Scan(&emails).Error
	return emails, err
}

// loadEmailService 加载邮件渠道配置（缓存 5 分钟；未配置返回 false）
func (n *TicketNotifier) loadEmailService() (*notification.EmailService, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.cacheInit && time.Since(n.cachedAt) < 5*time.Minute {
		return n.emailSvc, n.emailSvc != nil
	}
	n.cacheInit = true
	n.cachedAt = time.Now()

	// 取第一个启用的邮件渠道
	var raw string
	err := n.db.Raw("SELECT config FROM mon_notification_channels WHERE channel_type = 'email' AND enabled = 1 LIMIT 1").Scan(&raw).Error
	if err != nil || raw == "" {
		return nil, false
	}
	var cfg notification.EmailConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		logger.Warn("工单通知：邮件渠道配置格式错误", zap.Error(err))
		return nil, false
	}
	svc := notification.NewEmailService(cfg)
	n.emailSvc = svc
	return svc, true
}

func filterNotApproved(ids, approved []uint) []uint {
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if !containsUint(approved, id) {
			out = append(out, id)
		}
	}
	return out
}

func defaultText(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func u64(ids []uint) []uint64 {
	out := make([]uint64, len(ids))
	for i, id := range ids {
		out[i] = uint64(id)
	}
	return out
}
