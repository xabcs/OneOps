package modelticket

import "time"

// NotifyLog 工单通知发送记录（外部渠道投递结果，保留 30 天）
//
// 每次向外部渠道（email/wechat/dingtalk）投递后落一条记录，成功/失败均记，
// 用于"没收到通知"的排障与审计。站内消息（inbox）不走此表。
type NotifyLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Event       string    `json:"event" gorm:"size:50;index"`  // pending/result/reassign/cancel/urge
	TicketID    uint      `json:"ticketId" gorm:"index"`       // 关联工单（0 表示无单条工单维度）
	TicketNo    string    `json:"ticketNo" gorm:"size:50"`     // 工单号（冗余，展示用）
	ChannelType string    `json:"channelType" gorm:"size:20"`  // email/wechat/dingtalk
	ChannelName string    `json:"channelName" gorm:"size:100"` // 渠道名（群名/邮箱主题）
	Recipient   string    `json:"recipient" gorm:"size:500"`   // 接收者（邮箱列表/群标识，超长截断）
	Status      int       `json:"status"`                      // 1成功 0失败
	Error       string    `json:"error" gorm:"size:500"`       // 失败原因（成功为空）
	CreatedAt   time.Time `json:"createdAt" gorm:"index"`
}

// TableName 表名
func (NotifyLog) TableName() string { return "ticket_notify_logs" }
