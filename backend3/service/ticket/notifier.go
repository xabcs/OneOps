package serviceticket

import (
	"encoding/json"
	"fmt"
	"regexp"
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
// 渠道来自平台级渠道池（mon_notification_channels，email/wechat/dingtalk 可并存），
// 按事件矩阵（ticket_notify_policies）路由：每个事件可绑定一组渠道；
// 未配置策略的事件沿用默认行为（全部启用渠道、每类型取最先创建的一条）。
//   - email：向收件人逐一发送邮件（需用户配置邮箱）；
//   - wechat：向企业微信群机器人发送文本消息（正文点名相关人，有手机号者 @精准提醒）；
//   - dingtalk：向钉钉群机器人发送文本消息（atMobiles @有手机号的相关人）。
//
// 在审批事件（待审批/终态/改派/撤销/催办）发生后向相关人员发送提醒：
//   - 站内消息（ticket_messages）是兜底的第一公民——事件启用即为每个收件人落一条，
//     不依赖外部渠道可达性，外部渠道仅是推送增量；
//   - 外部渠道投递结果落发送记录（ticket_notify_logs），供"没收到通知"排障。
//
// 扩展新渠道：resolveChannels 的 switch 增加对应 channel_type 分支 + sendLoop 增加发送分支，
// 手机号（sys_users.phone）是各群机器人 @人的通用数据来源。
// 设计约束：
//   - 所有发送均为异步（goroutine），失败仅记日志，绝不阻塞或回滚审批事务；
//   - 未配置渠道、事件被停用或用户缺邮箱时静默跳过（记 info 日志），通知是尽力而为的增强能力；
//   - 事件 → 渠道解析结果缓存 5 分钟，避免每次发送都查库。
type TicketNotifier struct {
	db *gorm.DB

	mu        sync.Mutex
	cachedAt  time.Time
	cache     map[string]eventResolve // event -> 启用状态与渠道列表
	cacheInit bool                    // 是否已尝试加载过（区分"无渠道"与"未加载"，避免反复查库）

	// 防轰炸（蓝图⑥）
	quietMu     sync.Mutex
	quietCached time.Time
	quietCfg    modelticket.NotifyGlobal // 夜间静默期配置（5 分钟缓存）

	dedupeMu sync.Mutex
	dedupe   map[string]time.Time // "event:ticketID:userID" → 上次发送时间（同工单同事件去重窗口）

	rateMu    sync.Mutex
	rateHour  int          // 频控桶所属小时（整点，翻转变更）
	rateCount map[uint]int // userID → 该小时外部渠道已推送条数
}

// 防轰炸固定参数（不配置化，调整改这里）
const (
	dedupeWindow = 10 * time.Minute // 同工单同事件同收件人去重窗口（站内与外部均去重）
	hourlyLimit  = 30               // 单人外部渠道每小时推送上限（超出仅记站内消息）
	dedupeMaxKey = 50000            // 去重表容量上限（超过整体重置，防内存膨胀）
)

// eventResolve 单个事件的矩阵解析结果
type eventResolve struct {
	channels []*resolvedChannel // 该事件应走的外部渠道
	enabled  bool               // 事件是否启用（false=站内与外部渠道均不发送）
	titleTpl string             // 标题模板（空=代码内置默认文案，蓝图⑤）
	bodyTpl  string             // 正文模板（空=代码内置默认文案）
}

// resolvedChannel 解析后的可发送渠道实例（渠道池一行 → 一个 service）
type resolvedChannel struct {
	ChannelType string // email / wechat / dingtalk
	ChannelName string // 日志定位用

	Email    *notification.EmailService
	WeChat   *notification.WeChatService
	DingTalk *notification.DingTalkService
}

// NewTicketNotifier 创建通知器（db 为 nil 时通知功能自动降级为空操作）
func NewTicketNotifier(db *gorm.DB) *TicketNotifier {
	return &TicketNotifier{
		db:        db,
		cache:     make(map[string]eventResolve),
		dedupe:    make(map[string]time.Time),
		rateCount: make(map[uint]int),
	}
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
	ctx := baseTplCtx(ticket)
	ctx["node"] = nodeName
	n.sendAsync(modelticket.NotifyEventPending, ticket, pending, subject, body, ctx)
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
	ctx := baseTplCtx(ticket)
	ctx["status"] = resultText
	ctx["operator"] = operator
	ctx["comment"] = defaultText(comment, "（无）")
	n.sendAsync(modelticket.NotifyEventResult, ticket, []uint{ticket.CreatorID}, subject, body, ctx)
}

// NotifyReassigned 通知被改派的新审批人
func (n *TicketNotifier) NotifyReassigned(ticket *modelticket.Ticket, approverIDs []uint, nodeName, operator string) {
	if n == nil || n.db == nil || len(approverIDs) == 0 {
		return
	}
	subject := fmt.Sprintf("【OneOps 待审批】%s（%s）", ticket.Title, ticket.TicketNo)
	body := fmt.Sprintf("工单「%s」（%s）的节点「%s」已由 %s 改派给您审批，请登录 OneOps 工单中心处理。",
		ticket.Title, ticket.TicketNo, nodeName, operator)
	ctx := baseTplCtx(ticket)
	ctx["node"] = nodeName
	ctx["operator"] = operator
	n.sendAsync(modelticket.NotifyEventReassign, ticket, approverIDs, subject, body, ctx)
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
	ctx := baseTplCtx(ticket)
	ctx["node"] = nodeName
	ctx["operator"] = operator
	n.sendAsync(modelticket.NotifyEventCancel, ticket, pending, subject, body, ctx)
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
	ctx := baseTplCtx(ticket)
	ctx["node"] = nodeName
	ctx["operator"] = operator
	n.sendAsync(modelticket.NotifyEventUrge, ticket, pending, subject, body, ctx)
}

// NotifyTimeout 审批超时提醒：超时升级链触发，通知待审批人（蓝图④）
func (n *TicketNotifier) NotifyTimeout(ticket *modelticket.Ticket, approverIDs []uint, subject, body string, ctx map[string]string) {
	if n == nil || n.db == nil || len(approverIDs) == 0 {
		return
	}
	n.sendAsync(modelticket.NotifyEventTimeout, ticket, approverIDs, subject, body, ctx)
}

// NotifyEscalation 超时升级：超过 2 倍阈值仍无人处理，通知管理员与发起人（蓝图④）
func (n *TicketNotifier) NotifyEscalation(ticket *modelticket.Ticket, receiverIDs []uint, subject, body string, ctx map[string]string) {
	if n == nil || n.db == nil || len(receiverIDs) == 0 {
		return
	}
	n.sendAsync(modelticket.NotifyEventEscalation, ticket, receiverIDs, subject, body, ctx)
}

// ────────────────────────── 防轰炸（蓝图⑥） ──────────────────────────

// quietActive 夜间静默期是否生效（escalation 升级通知穿透静默；配置 5 分钟缓存）
func (n *TicketNotifier) quietActive(event string) bool {
	if event == modelticket.NotifyEventEscalation {
		return false
	}
	n.quietMu.Lock()
	defer n.quietMu.Unlock()
	if n.quietCached.IsZero() || time.Since(n.quietCached) > 5*time.Minute {
		var cfg modelticket.NotifyGlobal
		if err := n.db.Where("id = 1").First(&cfg).Error; err == nil {
			n.quietCfg = cfg
		}
		n.quietCached = time.Now()
	}
	if n.quietCfg.QuietEnabled != 1 {
		return false
	}
	h := time.Now().Hour()
	start, end := n.quietCfg.QuietStartHour, n.quietCfg.QuietEndHour
	if start == end {
		return false
	}
	if start < end { // 常规区间，如 0-8
		return h >= start && h < end
	}
	// 跨夜区间，如 22-8
	return h >= start || h < end
}

// dedupeUsers 同工单同事件去重：剔除窗口期内已收过的用户并登记本次时间（站内与外部均生效）
func (n *TicketNotifier) dedupeUsers(event string, ticketID uint, users []mentionUser) []mentionUser {
	now := time.Now()
	n.dedupeMu.Lock()
	defer n.dedupeMu.Unlock()
	if len(n.dedupe) > dedupeMaxKey {
		n.dedupe = make(map[string]time.Time) // 容量兜底整体重置（尽力而为的防抖，无需精确）
	}
	out := make([]mentionUser, 0, len(users))
	for _, u := range users {
		key := fmt.Sprintf("%s:%d:%d", event, ticketID, u.ID)
		if t, ok := n.dedupe[key]; ok && now.Sub(t) < dedupeWindow {
			continue
		}
		n.dedupe[key] = now
		out = append(out, u)
	}
	return out
}

// allowExternal 频控判定：该用户本小时外部渠道推送是否未超上限（超限仅记站内消息）
func (n *TicketNotifier) allowExternal(userID uint, count int) bool {
	hour := time.Now().Hour()
	n.rateMu.Lock()
	defer n.rateMu.Unlock()
	if n.rateHour != hour { // 整点翻桶，重置全部计数
		n.rateHour = hour
		n.rateCount = make(map[uint]int)
	}
	if n.rateCount[userID]+count > hourlyLimit {
		return false
	}
	n.rateCount[userID] += count
	return true
}

// ────────────────────────── 内部实现 ──────────────────────────

// sendAsync 异步发送：按事件解析启用状态与渠道 → 应用用户偏好（蓝图③）→
// 写站内消息（兜底，第一公民）→ 外部渠道（email/wechat/dingtalk）投递并记录结果
// （ticket_notify_logs），全程失败仅记日志。
//
// 用户偏好（ticket_user_notify_settings）叠加在事件矩阵之上：
//   - 屏蔽事件（mutedEvents）→ 该用户完全跳过（站内与外部渠道均不发）；
//   - 停用渠道（offChannels）→ 该用户不参与此渠道推送，站内消息照常落；
//   - IM userid（dingtalkId/wechatId）→ 群机器人 @人优先 userid，回退手机号。
func (n *TicketNotifier) sendAsync(event string, ticket *modelticket.Ticket, userIDs []uint, subject, body string, ctx map[string]string) {
	ec := n.resolveChannels(event)
	if !ec.enabled {
		logger.Info("工单通知：事件已停用（通知设置），跳过",
			zap.String("event", event), zap.String("subject", subject))
		return
	}

	// 自定义模板优先（蓝图⑤）：空模板沿用代码内置默认文案
	subject = renderTpl(ec.titleTpl, ctx, subject)
	body = renderTpl(ec.bodyTpl, ctx, body)

	// 昵称供群消息点名，邮箱供邮件渠道，手机号供企微/钉钉 @人兜底，
	// IM userid 供精准 @人（atUserIds / mentioned_list，来自用户偏好设置）
	var users []mentionUser
	err := n.db.Table("sys_users").
		Select("id, nickname, email, phone").
		Where("id IN ? AND status = 'active'", userIDs).
		Scan(&users).Error
	if err != nil {
		logger.Warn("工单通知：查询收件人信息失败", zap.Uint64s("user_ids", u64(userIDs)), zap.Error(err))
		return
	}

	pref := n.loadUserPrefs(userIDs, event)

	// 屏蔽事件的用户完全剔除（站内与外部渠道均不发）
	active := make([]mentionUser, 0, len(users))
	for _, u := range users {
		if s := pref[u.ID]; s != nil && s.MutesEvent(event) {
			continue
		}
		if s := pref[u.ID]; s != nil {
			u.DingtalkID, u.WechatID = s.DingtalkID, s.WechatID
		}
		active = append(active, u)
	}
	if len(active) == 0 {
		logger.Info("工单通知：收件人已全部屏蔽该事件，跳过",
			zap.String("event", event), zap.String("subject", subject))
		return
	}

	// 防轰炸①：同工单同事件同收件人去重窗口（站内与外部渠道均生效）
	active = n.dedupeUsers(event, ticket.ID, active)
	if len(active) == 0 {
		logger.Info("工单通知：去重窗口内重复通知，跳过",
			zap.String("event", event), zap.Uint("ticket_id", ticket.ID))
		return
	}

	// 站内消息兜底：不依赖外部渠道可达性，事件启用即为每个收件人落一条
	msgs := make([]modelticket.TicketMessage, 0, len(active))
	for _, u := range active {
		msgs = append(msgs, modelticket.TicketMessage{
			UserID: u.ID, TicketID: ticket.ID, TicketNo: ticket.TicketNo,
			Event: event, Title: subject, Content: body,
		})
	}
	if err := n.db.Create(&msgs).Error; err != nil {
		logger.Warn("工单通知：站内消息写入失败", zap.String("event", event), zap.Error(err))
	}

	if len(ec.channels) == 0 {
		logger.Info("工单通知：该事件未绑定可用渠道，仅记录站内消息",
			zap.String("event", event), zap.String("subject", subject))
		return
	}

	// 防轰炸②：夜间静默期，外部渠道暂停（站内消息不受影响；超时升级穿透静默）
	if n.quietActive(event) {
		logger.Info("工单通知：夜间静默期，外部渠道暂停发送（仅站内消息）",
			zap.String("event", event), zap.String("subject", subject))
		return
	}

	for _, ch := range ec.channels {
		// 用户停用该渠道 → 不参与推送；防轰炸③：单人小时频控超限 → 该用户本轮不推外部渠道
		chUsers := make([]mentionUser, 0, len(active))
		for _, u := range active {
			if s := pref[u.ID]; s != nil && s.OffChannel(ch.ChannelType) {
				continue
			}
			if !n.allowExternal(u.ID, 1) {
				logger.Info("工单通知：单人小时频控超限，暂停该收件人外部推送",
					zap.Uint("user_id", u.ID), zap.String("event", event))
				continue
			}
			chUsers = append(chUsers, u)
		}
		if len(chUsers) == 0 {
			logger.Info("工单通知：收件人已全部停用该渠道或频控超限，跳过",
				zap.String("event", event), zap.String("channel", ch.ChannelType))
			continue
		}

		switch ch.ChannelType {
		case "email":
			to := make([]string, 0, len(chUsers))
			for _, u := range chUsers {
				if u.Email != "" {
					to = append(to, u.Email)
				}
			}
			if len(to) == 0 {
				logger.Info("工单通知：收件人未配置邮箱，跳过邮件", zap.Uint64s("user_ids", u64(userIDs)))
				continue
			}
			svc, rcpts, subj, text := ch.Email, to, subject, body
			channelType, channelName, recipient := ch.ChannelType, ch.ChannelName, strings.Join(to, ",")
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("工单通知：邮件发送 panic", zap.Any("recover", r))
						n.logDelivery(event, ticket, channelType, channelName, recipient, false, fmt.Sprintf("panic: %v", r))
					}
				}()
				if err := svc.Send(rcpts, subj, text, false); err != nil {
					logger.Warn("工单通知：邮件发送失败", zap.Strings("to", rcpts), zap.Error(err))
					n.logDelivery(event, ticket, channelType, channelName, recipient, false, err.Error())
					return
				}
				n.logDelivery(event, ticket, channelType, channelName, recipient, true, "")
			}()
		case "wechat":
			// @人优先 userid（mentioned_list），未配置者回退手机号（mentioned_mobile_list）
			text, userids, mobiles := mentionText(body, chUsers, "wechat")
			svc, msg, atIDs, atMobiles := ch.WeChat, text, userids, mobiles
			channelType, channelName := ch.ChannelType, ch.ChannelName
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("工单通知：企业微信发送 panic", zap.Any("recover", r))
						n.logDelivery(event, ticket, channelType, channelName, "群机器人", false, fmt.Sprintf("panic: %v", r))
					}
				}()
				if err := svc.Send(&notification.WeChatMessage{
					MsgType: "text",
					Text: &notification.WeChatText{
						Content:             msg,
						MentionedList:       atIDs,
						MentionedMobileList: atMobiles,
					},
				}); err != nil {
					logger.Warn("工单通知：企业微信发送失败", zap.Error(err))
					n.logDelivery(event, ticket, channelType, channelName, "群机器人", false, err.Error())
					return
				}
				n.logDelivery(event, ticket, channelType, channelName, "群机器人", true, "")
			}()
		case "dingtalk":
			// @人优先 userid（atUserIds），未配置者回退手机号（atMobiles）
			text, userids, mobiles := mentionText(body, chUsers, "dingtalk")
			svc, msg, atIDs, atMobiles := ch.DingTalk, text, userids, mobiles
			channelType, channelName := ch.ChannelType, ch.ChannelName
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("工单通知：钉钉发送 panic", zap.Any("recover", r))
						n.logDelivery(event, ticket, channelType, channelName, "群机器人", false, fmt.Sprintf("panic: %v", r))
					}
				}()
				if err := svc.SendTextAt(msg, atIDs, atMobiles); err != nil {
					logger.Warn("工单通知：钉钉发送失败", zap.Error(err))
					n.logDelivery(event, ticket, channelType, channelName, "群机器人", false, err.Error())
					return
				}
				n.logDelivery(event, ticket, channelType, channelName, "群机器人", true, "")
			}()
		}
	}
}

// loadUserPrefs 批量加载收件人的通知偏好（user_id → 设置行，无记录=无偏好）
func (n *TicketNotifier) loadUserPrefs(userIDs []uint, event string) map[uint]*modelticket.UserNotifySetting {
	var rows []modelticket.UserNotifySetting
	if err := n.db.Where("user_id IN ?", userIDs).Find(&rows).Error; err != nil {
		logger.Warn("工单通知：加载用户偏好失败，按无偏好处理",
			zap.String("event", event), zap.Error(err))
		return map[uint]*modelticket.UserNotifySetting{}
	}
	pref := make(map[uint]*modelticket.UserNotifySetting, len(rows))
	for i := range rows {
		pref[rows[i].UserID] = &rows[i]
	}
	return pref
}

// logDelivery 落一条外部渠道投递记录（成功/失败均记，字段超长截断）
func (n *TicketNotifier) logDelivery(event string, ticket *modelticket.Ticket, channelType, channelName, recipient string, ok bool, errMsg string) {
	if ticket == nil {
		ticket = &modelticket.Ticket{}
	}
	if len(recipient) > 500 {
		recipient = recipient[:500]
	}
	if len(errMsg) > 500 {
		errMsg = errMsg[:500]
	}
	status := 0
	if ok {
		status = 1
	}
	if err := n.db.Create(&modelticket.NotifyLog{
		Event: event, TicketID: ticket.ID, TicketNo: ticket.TicketNo,
		ChannelType: channelType, ChannelName: channelName,
		Recipient: recipient, Status: status, Error: errMsg,
	}).Error; err != nil {
		logger.Warn("工单通知：写入发送记录失败", zap.Error(err))
	}
}

// mentionUser 通知收件人（站内消息按 ID 落库，邮箱/手机号供外部渠道，
// IM userid 来自用户偏好设置，供群机器人精准 @人）
type mentionUser = struct {
	ID         uint
	Nickname   string
	Email      string
	Phone      string
	DingtalkID string
	WechatID   string
}

// mentionText 构造群消息正文（正文追加相关人名单）与 @ 人列表。
// channel 决定 userid 取值：wechat→WechatID，dingtalk→DingtalkID；
// 配了 userid 的用 userid @，未配置者回退手机号，两者皆无仅正文点名。
func mentionText(body string, users []mentionUser, channel string) (text string, userids, mobiles []string) {
	names := make([]string, 0, len(users))
	userids = make([]string, 0, len(users))
	mobiles = make([]string, 0, len(users))
	for _, u := range users {
		if u.Nickname != "" {
			names = append(names, u.Nickname)
		}
		id := ""
		if channel == "wechat" {
			id = u.WechatID
		} else if channel == "dingtalk" {
			id = u.DingtalkID
		}
		if id != "" {
			userids = append(userids, id)
		} else if u.Phone != "" {
			mobiles = append(mobiles, u.Phone)
		}
	}
	if len(names) > 0 {
		text = fmt.Sprintf("%s\n\n相关人：%s", body, strings.Join(names, "、"))
	} else {
		text = body
	}
	return text, userids, mobiles
}

// resolveChannels 按事件解析启用状态与渠道列表（缓存 5 分钟）。
// 策略语义见 modelticket.NotifyPolicy：无记录=默认（启用 + 全部启用渠道，每类型一条）；
// enabled=0=事件停用（站内与外部渠道均不发送）；enabled=1 且 channels 空=仅站内消息。
func (n *TicketNotifier) resolveChannels(event string) eventResolve {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.cacheInit && time.Since(n.cachedAt) < 5*time.Minute {
		return n.cache[event]
	}
	n.cacheInit = true
	n.cachedAt = time.Now()
	n.cache = make(map[string]eventResolve)

	// 渠道池：id → 渠道行（一次性加载，供各事件按 ID 解析）
	pool := n.loadChannelPool()

	var policies []modelticket.NotifyPolicy
	if err := n.db.Find(&policies).Error; err != nil {
		logger.Warn("工单通知：加载事件矩阵失败，全部事件走默认渠道", zap.Error(err))
	}

	for _, meta := range notifyEventMetas {
		n.cache[meta.Event] = n.channelsForEvent(meta.Event, policies, pool)
	}
	return n.cache[event]
}

// channelsForEvent 计算单个事件的渠道与启用状态
func (n *TicketNotifier) channelsForEvent(event string, policies []modelticket.NotifyPolicy, pool map[uint]*resolvedChannel) eventResolve {
	var policy *modelticket.NotifyPolicy
	for i := range policies {
		if policies[i].Event == event {
			policy = &policies[i]
			break
		}
	}

	// 未配置策略：默认行为——启用 + 全部启用渠道，每类型取一条（渠道池已按 id 升序）
	if policy == nil {
		out := make([]*resolvedChannel, 0, 3)
		seen := make(map[string]bool)
		for id := uint(1); id <= maxChannelID(pool); id++ {
			ch, ok := pool[id]
			if !ok || seen[ch.ChannelType] {
				continue
			}
			seen[ch.ChannelType] = true
			out = append(out, ch)
		}
		return eventResolve{channels: out, enabled: true}
	}
	if policy.Enabled == 0 {
		return eventResolve{enabled: false} // 事件已停用
	}

	var ids []uint
	if err := json.Unmarshal([]byte(policy.Channels), &ids); err != nil {
		logger.Warn("工单通知：事件渠道绑定格式错误，按未配置处理",
			zap.String("event", event), zap.Error(err))
		return eventResolve{enabled: true}
	}
	out := make([]*resolvedChannel, 0, len(ids))
	for _, id := range ids {
		if ch, ok := pool[id]; ok {
			out = append(out, ch)
		}
	}
	return eventResolve{channels: out, enabled: true, titleTpl: policy.TitleTpl, bodyTpl: policy.BodyTpl}
}

// ────────────────────────── 通知模板（蓝图⑤） ──────────────────────────

// tplVarRe 模板变量占位：{{name}}（仅字母数字，避免误伤 JSON/HTML 花括号）
var tplVarRe = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_]+)\s*\}\}`)

// renderTpl 渲染通知模板：替换 {{name}} 变量（未定义变量原样保留，便于发现拼写错误），
// 渲染结果为空串时回退默认文案。模板保存在事件矩阵（ticket_notify_policies），空=用默认。
func renderTpl(tpl string, ctx map[string]string, fallback string) string {
	if tpl == "" {
		return fallback
	}
	out := tplVarRe.ReplaceAllStringFunc(tpl, func(m string) string {
		name := tplVarRe.FindStringSubmatch(m)[1]
		if v, ok := ctx[name]; ok && v != "" {
			return v
		}
		return m // 未定义变量原样保留
	})
	if strings.TrimSpace(out) == "" {
		return fallback
	}
	return out
}

// baseTplCtx 工单通用模板变量（各事件通用，正文/标题均可引用）
func baseTplCtx(t *modelticket.Ticket) map[string]string {
	return map[string]string{
		"ticketNo": t.TicketNo,
		"title":    t.Title,
		"typeName": defaultText(t.TypeName, "-"),
		"creator":  t.CreatorName,
		"priority": priorityText(t.Priority),
		"time":     time.Now().Format("2006-01-02 15:04:05"),
	}
}

// maxChannelID 池内最大渠道 ID（用于按 id 升序遍历；池空返回 0）
func maxChannelID(pool map[uint]*resolvedChannel) uint {
	var max uint
	for id := range pool {
		if id > max {
			max = id
		}
	}
	return max
}

// loadChannelPool 加载渠道池（enabled=1 的 email/wechat/dingtalk；配置无效的行记日志跳过）
func (n *TicketNotifier) loadChannelPool() map[uint]*resolvedChannel {
	var rows []struct {
		ID          uint
		ChannelType string
		ChannelName string
		Config      string
	}
	err := n.db.Raw(`SELECT id, channel_type, IFNULL(channel_name, '') AS channel_name, IFNULL(config, '') AS config
		FROM mon_notification_channels
		WHERE enabled = 1 AND channel_type IN ('email', 'wechat', 'dingtalk')
		ORDER BY id`).Scan(&rows).Error
	if err != nil {
		logger.Warn("工单通知：加载渠道池失败", zap.Error(err))
		return map[uint]*resolvedChannel{}
	}

	pool := make(map[uint]*resolvedChannel, len(rows))
	for _, row := range rows {
		switch row.ChannelType {
		case "email":
			var cfg notification.EmailConfig
			if err := json.Unmarshal([]byte(row.Config), &cfg); err != nil {
				logger.Warn("工单通知：邮件渠道配置格式错误，跳过", zap.Uint("channel_id", row.ID), zap.Error(err))
				continue
			}
			pool[row.ID] = &resolvedChannel{
				ChannelType: "email", ChannelName: row.ChannelName,
				Email: notification.NewEmailService(cfg),
			}
		case "wechat":
			var cfg notification.WeChatConfig
			if err := json.Unmarshal([]byte(row.Config), &cfg); err != nil || strings.TrimSpace(cfg.WebhookURL) == "" {
				logger.Warn("工单通知：企业微信渠道配置无效，跳过", zap.Uint("channel_id", row.ID), zap.Error(err))
				continue
			}
			pool[row.ID] = &resolvedChannel{
				ChannelType: "wechat", ChannelName: row.ChannelName,
				WeChat: notification.NewWeChatService(cfg),
			}
		case "dingtalk":
			var cfg notification.DingTalkConfig
			if err := json.Unmarshal([]byte(row.Config), &cfg); err != nil || strings.TrimSpace(cfg.WebhookURL) == "" {
				logger.Warn("工单通知：钉钉渠道配置无效，跳过", zap.Uint("channel_id", row.ID), zap.Error(err))
				continue
			}
			pool[row.ID] = &resolvedChannel{
				ChannelType: "dingtalk", ChannelName: row.ChannelName,
				DingTalk: notification.NewDingTalkService(cfg),
			}
		}
	}
	return pool
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
