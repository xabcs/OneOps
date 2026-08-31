package modelticket

import (
	"encoding/json"
	"time"
)

// UserNotifySetting 用户级通知偏好（每人最多一行，无记录=跟随全局默认）
//
// 与管理员侧事件矩阵（NotifyPolicy）的分工：
//   - 事件矩阵：全局路由——这个事件通过哪些渠道推送给谁（角色维度，管理员配置）；
//   - 用户偏好：个人收件箱——我不想被打扰（蓝图③，摆脱手机号依赖的第一步）。
//
// 语义：
//   - MutedEvents 含某事件 → 该用户完全不接收此事件（站内消息与外部渠道均不发）；
//   - OffChannels 含某渠道类型（email/wechat/dingtalk）→ 该用户不参与此渠道推送，
//     站内消息照常落（站内是第一公民，外部推送是增量）；
//   - DingtalkID / WechatID：IM 用户标识，群机器人 @人优先用 userid（atUserIds /
//     mentioned_list），未配置时回退手机号。
type UserNotifySetting struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"userId" gorm:"uniqueIndex;not null"`
	DingtalkID  string    `json:"dingtalkId" gorm:"size:100"`  // 钉钉 userid（@人用）
	WechatID    string    `json:"wechatId" gorm:"size:100"`    // 企业微信 userid（@人用）
	MutedEvents string    `json:"mutedEvents" gorm:"size:200"` // JSON 数组，如 ["urge","result"]
	OffChannels string    `json:"offChannels" gorm:"size:200"` // JSON 数组，如 ["email"]
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName 表名
func (UserNotifySetting) TableName() string { return "ticket_user_notify_settings" }

// MutesEvent 该用户是否完全屏蔽此事件（站内与外部渠道均不发）
func (r *UserNotifySetting) MutesEvent(event string) bool { return listContains(r.MutedEvents, event) }

// OffChannel 该用户是否不参与此渠道推送（站内消息照常落）
func (r *UserNotifySetting) OffChannel(channelType string) bool {
	return listContains(r.OffChannels, channelType)
}

// listContains JSON 字符串数组包含判断（空/格式错返回 false）
func listContains(s, item string) bool {
	if s == "" {
		return false
	}
	var list []string
	if err := json.Unmarshal([]byte(s), &list); err != nil {
		return false
	}
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
