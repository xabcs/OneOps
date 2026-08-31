package serviceticket

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	modelticket "oneops/backend3/model/ticket"
)

// NotifyEventMeta 工单通知事件元数据（固定事件集，顺序即前端展示序）
type NotifyEventMeta struct {
	Event string `json:"event"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
}

// notifyEventMetas 事件矩阵的固定事件定义
var notifyEventMetas = []NotifyEventMeta{
	{modelticket.NotifyEventPending, "待审批提醒", "审批节点激活时通知待审批人"},
	{modelticket.NotifyEventResult, "审批结果通知", "工单通过/驳回时通知发起人"},
	{modelticket.NotifyEventReassign, "改派通知", "审批人被改派时通知新审批人"},
	{modelticket.NotifyEventCancel, "撤销通知", "发起人撤销工单时通知当前待审人"},
	{modelticket.NotifyEventUrge, "催办提醒", "发起人催办时通知待审批人"},
	{modelticket.NotifyEventTimeout, "审批超时提醒", "节点停留超阈值时自动提醒待审批人"},
	{modelticket.NotifyEventEscalation, "超时升级", "超过2倍阈值仍无人处理，通知管理员与发起人"},
}

// NotifyPolicyService 工单通知策略（事件矩阵）管理
type NotifyPolicyService struct {
	db *gorm.DB
}

// NewNotifyPolicyService 创建服务
func NewNotifyPolicyService(db *gorm.DB) *NotifyPolicyService {
	return &NotifyPolicyService{db: db}
}

// NotifyPolicyView 事件矩阵视图：固定事件 × 已有配置
type NotifyPolicyView struct {
	Event      string `json:"event"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Configured bool   `json:"configured"` // 是否已定制（false=默认：全部启用渠道）
	Enabled    int    `json:"enabled"`
	Channels   []uint `json:"channels"` // 选中的渠道 ID 列表
	TitleTpl   string `json:"titleTpl"` // 标题模板（空=内置默认文案）
	BodyTpl    string `json:"bodyTpl"`  // 正文模板（空=内置默认文案）
	HasTpl     bool   `json:"hasTpl"`   // 是否已自定义模板
}

// GetPolicies 返回完整事件矩阵（未配置的事件给默认值）
func (s *NotifyPolicyService) GetPolicies() ([]NotifyPolicyView, error) {
	var rows []modelticket.NotifyPolicy
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	byEvent := make(map[string]modelticket.NotifyPolicy, len(rows))
	for _, r := range rows {
		byEvent[r.Event] = r
	}

	views := make([]NotifyPolicyView, 0, len(notifyEventMetas))
	for _, m := range notifyEventMetas {
		v := NotifyPolicyView{
			Event: m.Event, Name: m.Name, Desc: m.Desc,
			Enabled: 1, Channels: []uint{},
		}
		if p, ok := byEvent[m.Event]; ok {
			v.Configured = true
			v.Enabled = p.Enabled
			v.TitleTpl = p.TitleTpl
			v.BodyTpl = p.BodyTpl
			v.HasTpl = p.TitleTpl != "" || p.BodyTpl != ""
			var ch []uint
			if json.Unmarshal([]byte(p.Channels), &ch) == nil {
				v.Channels = ch
			}
		}
		views = append(views, v)
	}
	return views, nil
}

// SavePolicy 保存某事件的渠道绑定、启用状态与通知模板（upsert；模板空=恢复内置默认）
func (s *NotifyPolicyService) SavePolicy(event string, channels []uint, enabled int, titleTpl, bodyTpl string) error {
	if !isValidNotifyEvent(event) {
		return fmt.Errorf("未知通知事件: %s", event)
	}
	if enabled != 0 && enabled != 1 {
		return fmt.Errorf("enabled 取值须为 0 或 1")
	}
	if len(channels) > 20 {
		return fmt.Errorf("渠道数量超出限制")
	}
	if len(titleTpl) > 500 || len(bodyTpl) > 5000 {
		return fmt.Errorf("模板长度超出限制（标题 500、正文 5000 字符）")
	}
	chJSON, _ := json.Marshal(channels)

	var existing modelticket.NotifyPolicy
	err := s.db.Where("event = ?", event).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.db.Create(&modelticket.NotifyPolicy{
			Event: event, Channels: string(chJSON), Enabled: enabled,
			TitleTpl: titleTpl, BodyTpl: bodyTpl,
		}).Error
	}
	if err != nil {
		return err
	}
	return s.db.Model(&existing).Updates(map[string]interface{}{
		"channels":  string(chJSON),
		"enabled":   enabled,
		"title_tpl": titleTpl,
		"body_tpl":  bodyTpl,
	}).Error
}

// isValidNotifyEvent 校验事件是否属于固定事件集
func isValidNotifyEvent(event string) bool {
	for _, m := range notifyEventMetas {
		if m.Event == event {
			return true
		}
	}
	return false
}

// GetQuietConfig 读取夜间静默期配置（无记录时返回默认：未启用 22-8）
func (s *NotifyPolicyService) GetQuietConfig() (*modelticket.NotifyGlobal, error) {
	var row modelticket.NotifyGlobal
	err := s.db.Where("id = 1").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &modelticket.NotifyGlobal{ID: 1, QuietEnabled: 0, QuietStartHour: 22, QuietEndHour: 8}, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// SaveQuietConfig 保存夜间静默期配置（单行 upsert）
func (s *NotifyPolicyService) SaveQuietConfig(enabled, startHour, endHour int) error {
	if enabled != 0 && enabled != 1 {
		return fmt.Errorf("enabled 取值须为 0 或 1")
	}
	if startHour < 0 || startHour > 23 || endHour < 0 || endHour > 23 {
		return fmt.Errorf("静默时段小时取值须为 0-23")
	}
	row := modelticket.NotifyGlobal{ID: 1, QuietEnabled: enabled, QuietStartHour: startHour, QuietEndHour: endHour}
	return s.db.Save(&row).Error
}

// ListLogs 分页查询通知发送记录（event/status 可选过滤，倒序）
func (s *NotifyPolicyService) ListLogs(page, pageSize int, event string, status int) ([]modelticket.NotifyLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := s.db.Model(&modelticket.NotifyLog{})
	if event != "" {
		q = q.Where("event = ?", event)
	}
	if status == 0 || status == 1 {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []modelticket.NotifyLog
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
