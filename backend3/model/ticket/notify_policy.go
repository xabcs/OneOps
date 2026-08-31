package modelticket

import "time"

// 工单通知事件类型（事件矩阵的固定事件集）
const (
	NotifyEventPending    = "pending"    // 待审批：节点激活，通知待审批人
	NotifyEventResult     = "result"     // 审批结果：工单通过/驳回，通知发起人
	NotifyEventReassign   = "reassign"   // 改派：审批人被改派，通知新审批人
	NotifyEventCancel     = "cancel"     // 撤销：发起人撤销工单，通知当前待审人
	NotifyEventUrge       = "urge"       // 催办：通知待审批人
	NotifyEventTimeout    = "timeout"    // 审批超时：节点停留超阈值，自动提醒待审批人（蓝图④）
	NotifyEventEscalation = "escalation" // 超时升级：超过 2 倍阈值仍无人处理，通知管理员与发起人（蓝图④）
)

// NotifyPolicy 工单通知策略（事件矩阵），一行一个事件
//
// channels 为 mon_notification_channels.id 的 JSON 数组（如 [1,2]），语义：
//   - 无记录：默认行为——该事件走所有启用渠道（兼容未配置时的历史行为）
//   - enabled=0：该事件不发任何通知
//   - enabled=1 且 channels 非空：仅走所选渠道（渠道自身也须启用）
//   - enabled=1 且 channels 为空：显式不选渠道，不发送
//
// TitleTpl/BodyTpl 通知模板（蓝图⑤）：支持 {{ticketNo}}/{{title}}/{{node}} 等变量占位，
// 空值 = 使用代码内置默认文案；未定义变量原样保留（便于发现拼写错误）
type NotifyPolicy struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Event     string    `json:"event" gorm:"uniqueIndex;size:50;not null"`
	Channels  string    `json:"channels" gorm:"size:500"` // JSON 数组 [1,2]
	Enabled   int       `json:"enabled" gorm:"default:1"`
	TitleTpl  string    `json:"titleTpl" gorm:"size:500"`
	BodyTpl   string    `json:"bodyTpl" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 表名
func (NotifyPolicy) TableName() string { return "ticket_notify_policies" }
