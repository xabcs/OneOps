package serviceticket

import (
	"encoding/json"

	"gorm.io/gorm"

	modelticket "oneops/backend3/model/ticket"
)

// 用户可操作的渠道类型（与通知渠道池一致）
var userChannelTypes = map[string]bool{
	"email": true, "wechat": true, "dingtalk": true,
}

// UserNotifySettingDTO 用户偏好的读写视图（数组已解析，前端直接用）
type UserNotifySettingDTO struct {
	DingtalkID  string   `json:"dingtalkId"`
	WechatID    string   `json:"wechatId"`
	MutedEvents []string `json:"mutedEvents"`
	OffChannels []string `json:"offChannels"`
}

// SaveUserNotifySettingRequest 保存请求
type SaveUserNotifySettingRequest struct {
	DingtalkID  string   `json:"dingtalkId" binding:"max=100"`
	WechatID    string   `json:"wechatId" binding:"max=100"`
	MutedEvents []string `json:"mutedEvents"`
	OffChannels []string `json:"offChannels"`
}

// UserNotifySettingService 用户通知偏好（本人自助，登录即可读写）
type UserNotifySettingService struct {
	db *gorm.DB
}

// NewUserNotifySettingService 创建服务
func NewUserNotifySettingService(db *gorm.DB) *UserNotifySettingService {
	return &UserNotifySettingService{db: db}
}

// GetSetting 读取本人偏好（无记录返回空默认，跟随全局）
func (s *UserNotifySettingService) GetSetting(userID uint) (*UserNotifySettingDTO, error) {
	var row modelticket.UserNotifySetting
	err := s.db.Where("user_id = ?", userID).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &UserNotifySettingDTO{MutedEvents: []string{}, OffChannels: []string{}}, nil
		}
		return nil, err
	}
	return &UserNotifySettingDTO{
		DingtalkID:  row.DingtalkID,
		WechatID:    row.WechatID,
		MutedEvents: parseJSONList(row.MutedEvents),
		OffChannels: parseJSONList(row.OffChannels),
	}, nil
}

// SaveSetting 保存本人偏好（upsert；事件/渠道值非法时静默忽略该项）
func (s *UserNotifySettingService) SaveSetting(userID uint, req *SaveUserNotifySettingRequest) error {
	setting := modelticket.UserNotifySetting{
		UserID:     userID,
		DingtalkID: trimSpace(req.DingtalkID),
		WechatID:   trimSpace(req.WechatID),
	}
	if b, err := json.Marshal(filterValidEvents(req.MutedEvents)); err == nil {
		setting.MutedEvents = string(b)
	}
	if b, err := json.Marshal(filterValidChannels(req.OffChannels)); err == nil {
		setting.OffChannels = string(b)
	}
	return s.db.Where("user_id = ?", userID).
		Assign(map[string]interface{}{
			"dingtalk_id":  setting.DingtalkID,
			"wechat_id":    setting.WechatID,
			"muted_events": setting.MutedEvents,
			"off_channels": setting.OffChannels,
		}).
		FirstOrCreate(&setting).Error
}

// ────────────────────────── 内部工具 ──────────────────────────

// parseJSONList 解析 JSON 字符串数组（空/格式错返回空数组）
func parseJSONList(s string) []string {
	out := parseJSONListValue(s)
	if out == nil {
		return []string{}
	}
	return out
}

func parseJSONListValue(s string) []string {
	if s == "" {
		return nil
	}
	var list []string
	if err := json.Unmarshal([]byte(s), &list); err != nil {
		return nil
	}
	return list
}

func jsonListContains(s, item string) bool {
	for _, v := range parseJSONListValue(s) {
		if v == item {
			return true
		}
	}
	return false
}

func filterValidEvents(in []string) []string {
	valid := map[string]bool{}
	for _, meta := range notifyEventMetas {
		valid[meta.Event] = true
	}
	return filterList(in, valid)
}

func filterValidChannels(in []string) []string {
	return filterList(in, userChannelTypes)
}

func filterList(in []string, valid map[string]bool) []string {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, v := range in {
		if valid[v] && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
