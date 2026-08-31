package modelticket

import "time"

// TicketMessage 工单站内消息（通知兜底渠道，保留 90 天）
//
// 只要事件未停用（ticket_notify_policies.enabled=1），每次通知都会为每个收件人
// 写一条站内消息——站内待办/消息是通知的第一公民，不依赖外部 IM 可达性；
// 事件矩阵（渠道绑定）仅控制外部推送渠道。
type TicketMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"index;not null"`                 // 接收人
	TicketID  uint      `json:"ticketId" gorm:"index"`                        // 关联工单
	TicketNo  string    `json:"ticketNo" gorm:"size:50"`                      // 工单号
	Event     string    `json:"event" gorm:"size:50"`                         // pending/result/reassign/cancel/urge
	Title     string    `json:"title" gorm:"size:200"`                        // 消息标题
	Content   string    `json:"content" gorm:"type:text"`                     // 消息正文摘要
	IsRead    int       `json:"isRead" gorm:"column:is_read;default:0;index"` // 0未读 1已读（read 是 MySQL 保留字，列名用 is_read）
	CreatedAt time.Time `json:"createdAt" gorm:"index"`
}

// TableName 表名
func (TicketMessage) TableName() string { return "ticket_messages" }
