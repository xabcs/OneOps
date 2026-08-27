package modelticket

import "time"

// ────────────────────────── 工单公共常量 ──────────────────────────
// 状态/动作字面量统一在此定义，service / repository / controller 共用，
// 避免散落的魔法字符串拼写不一致导致状态机判断失效。

// TicketStatus 工单状态
const (
	TicketStatusPending  = "pending"  // 审批中
	TicketStatusApproved = "approved" // 已通过
	TicketStatusRejected = "rejected" // 已驳回
	TicketStatusCanceled = "canceled" // 已撤销
)

// NodeStatus 节点审批记录状态
const (
	NodeStatusWaiting  = "waiting"  // 未到达
	NodeStatusPending  = "pending"  // 审批中
	NodeStatusApproved = "approved" // 已通过
	NodeStatusRejected = "rejected" // 已驳回
	NodeStatusSkipped  = "skipped"  // 条件跳过
	NodeStatusCanceled = "canceled" // 工单终止未走完
)

// FlowAction 流转日志动作
const (
	FlowActionSubmit   = "submit"   // 发起
	FlowActionApprove  = "approve"  // 通过
	FlowActionReject   = "reject"   // 驳回
	FlowActionCancel   = "cancel"   // 撤销
	FlowActionComment  = "comment"  // 评论
	FlowActionSkip     = "skip"     // 节点跳过
	FlowActionResubmit = "resubmit" // 驳回后重新提交
	FlowActionReassign = "reassign" // 改派审批人
	FlowActionUrge     = "urge"     // 发起人催办
)

// Priority 工单优先级
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// MultiType 多人审批方式
const (
	MultiTypeAll = "all" // 会签：全部通过
)

// UrgeCooldown 催办冷却时长（两次催办最小间隔）
const UrgeCooldown = 30 * time.Minute
