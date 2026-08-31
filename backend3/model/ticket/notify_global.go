package modelticket

import "time"

// NotifyGlobal 工单通知全局设置（单行表，id 恒为 1）—— 防轰炸配置（蓝图⑥）
//
// QuietEnabled : 夜间静默期开关。静默时段内外部渠道（邮件/企微/钉钉）暂停发送，
//
//	站内消息不受影响（第一公民）；超时升级（escalation）穿透静默。
//
// QuietStartHour/QuietEndHour : 静默时段起止小时（0-23，支持跨夜，如 22 → 8）。
//
// 另有两项固定防护（代码常量，不配置化）：
//   - 同工单同事件同收件人 10 分钟去重窗口（站内与外部均去重）；
//   - 单人外部渠道每小时推送上限 30 条（超出仅记站内消息）。
type NotifyGlobal struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	QuietEnabled   int       `json:"quietEnabled" gorm:"default:0"`
	QuietStartHour int       `json:"quietStartHour" gorm:"default:22"`
	QuietEndHour   int       `json:"quietEndHour" gorm:"default:8"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// TableName 表名
func (NotifyGlobal) TableName() string { return "ticket_notify_globals" }
