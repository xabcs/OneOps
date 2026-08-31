package serviceticket

import (
	"gorm.io/gorm"

	modelticket "oneops/backend3/model/ticket"
)

// TicketMessageService 工单站内消息（登录用户本人的消息收件箱）
type TicketMessageService struct {
	db *gorm.DB
}

// NewTicketMessageService 创建服务
func NewTicketMessageService(db *gorm.DB) *TicketMessageService {
	return &TicketMessageService{db: db}
}

// ListMessages 分页查询本人消息（unreadOnly=1 时只看未读，前端铃铛默认读最近消息）
func (s *TicketMessageService) ListMessages(userID uint, page, pageSize int, unreadOnly bool) ([]modelticket.TicketMessage, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := s.db.Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("is_read = 0")
	}
	var total int64
	if err := q.Model(&modelticket.TicketMessage{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []modelticket.TicketMessage
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UnreadCount 本人未读消息数
func (s *TicketMessageService) UnreadCount(userID uint) (int64, error) {
	var n int64
	err := s.db.Model(&modelticket.TicketMessage{}).
		Where("user_id = ? AND is_read = 0", userID).Count(&n).Error
	return n, err
}

// MarkRead 标记已读（ids 为空表示全部已读）
func (s *TicketMessageService) MarkRead(userID uint, ids []uint) error {
	q := s.db.Model(&modelticket.TicketMessage{}).Where("user_id = ? AND is_read = 0", userID)
	if len(ids) > 0 {
		q = q.Where("id IN ?", ids)
	}
	return q.Update("is_read", 1).Error
}
