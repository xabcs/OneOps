package repositoryticket

import (
	modelticket "oneops/backend3/model/ticket"

	"gorm.io/gorm"
)

// TicketRepository 工单数据访问
type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// Transaction 事务执行（引擎内多步流转操作使用）
func (r *TicketRepository) Transaction(fn func(tx *TicketRepository) error) error {
	return r.db.Transaction(func(db *gorm.DB) error {
		return fn(NewTicketRepository(db))
	})
}

// TicketQuery 工单列表查询条件
type TicketQuery struct {
	Scope   string // created=我发起的 / todo=待我审批 / done=我已审批 / all=全部
	Status  string
	TypeID  uint
	Keyword string
	UserID  uint
	Offset  int
	Limit   int
}

// FindTickets 分页查询工单列表
func (r *TicketRepository) FindTickets(q TicketQuery) ([]modelticket.Ticket, int64, error) {
	query := r.db.Model(&modelticket.Ticket{})

	switch q.Scope {
	case "created":
		query = query.Where("creator_id = ?", q.UserID)
	case "todo":
		// 当前节点审批中，且我是审批人，且我还没审过（会签场景）
		query = query.Where(`status = 'pending' AND id IN (
			SELECT ticket_id FROM ticket_node_records
			WHERE status = 'pending'
			  AND FIND_IN_SET(?, approver_ids)
			  AND NOT FIND_IN_SET(?, approved_ids))`, q.UserID, q.UserID)
	case "done":
		query = query.Where(`id IN (
			SELECT ticket_id FROM ticket_flow_logs
			WHERE operator_id = ? AND action IN ('approve','reject'))`, q.UserID)
	}

	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}
	if q.TypeID > 0 {
		query = query.Where("type_id = ?", q.TypeID)
	}
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		query = query.Where("title LIKE ? OR ticket_no LIKE ? OR creator_name LIKE ?", kw, kw, kw)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []modelticket.Ticket
	if err := query.Order("id DESC").Offset(q.Offset).Limit(q.Limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindTicketByID 按 ID 查工单
func (r *TicketRepository) FindTicketByID(id uint) (*modelticket.Ticket, error) {
	var t modelticket.Ticket
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTicket 创建工单
func (r *TicketRepository) CreateTicket(t *modelticket.Ticket) error {
	return r.db.Create(t).Error
}

// UpdateTicketFields 更新工单指定字段
func (r *TicketRepository) UpdateTicketFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&modelticket.Ticket{}).Where("id = ?", id).Updates(fields).Error
}

// CountTodayTickets 当日工单数（生成单号用）
func (r *TicketRepository) CountTodayTickets(prefix string) (int64, error) {
	var count int64
	err := r.db.Model(&modelticket.Ticket{}).Where("ticket_no LIKE ?", prefix+"%").Count(&count).Error
	return count, err
}

// ────────────────────────── 节点记录 ──────────────────────────

// FindNodeRecords 查询工单的全部节点记录（按 ID 升序，与快照顺序一致）
func (r *TicketRepository) FindNodeRecords(ticketID uint) ([]modelticket.TicketNodeRecord, error) {
	var list []modelticket.TicketNodeRecord
	err := r.db.Where("ticket_id = ?", ticketID).Order("id ASC").Find(&list).Error
	return list, err
}

// CreateNodeRecords 批量创建节点记录
func (r *TicketRepository) CreateNodeRecords(records []modelticket.TicketNodeRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.Create(&records).Error
}

// FindNodeRecordByID 按ID查节点记录
func (r *TicketRepository) FindNodeRecordByID(id uint) (*modelticket.TicketNodeRecord, error) {
	var rec modelticket.TicketNodeRecord
	if err := r.db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindPendingRecordsByTicketIDs 批量查询工单的当前待审批节点（每单一条）
func (r *TicketRepository) FindPendingRecordsByTicketIDs(ticketIDs []uint) ([]modelticket.TicketNodeRecord, error) {
	if len(ticketIDs) == 0 {
		return nil, nil
	}
	var list []modelticket.TicketNodeRecord
	err := r.db.Where("ticket_id IN ? AND status = 'pending'", ticketIDs).Find(&list).Error
	return list, err
}

// UpdateNodeRecord 更新节点记录
func (r *TicketRepository) UpdateNodeRecord(rec *modelticket.TicketNodeRecord) error {
	return r.db.Save(rec).Error
}

// ────────────────────────── 流转日志 ──────────────────────────

// CreateFlowLog 写入流转日志
func (r *TicketRepository) CreateFlowLog(log modelticket.TicketFlowLog) error {
	return r.db.Create(&log).Error
}

// FindFlowLogs 查询工单流转日志
func (r *TicketRepository) FindFlowLogs(ticketID uint) ([]modelticket.TicketFlowLog, error) {
	var list []modelticket.TicketFlowLog
	err := r.db.Where("ticket_id = ?", ticketID).Order("id ASC").Find(&list).Error
	return list, err
}

// UserTouchedTicket 用户是否参与过该工单（发起/审批/评论）
func (r *TicketRepository) UserTouchedTicket(ticketID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&modelticket.TicketFlowLog{}).
		Where("ticket_id = ? AND operator_id = ?", ticketID, userID).Count(&count).Error
	return count > 0, err
}

// UserIsNodeApprover 用户是否是该工单任一节点的审批人
func (r *TicketRepository) UserIsNodeApprover(ticketID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&modelticket.TicketNodeRecord{}).
		Where("ticket_id = ? AND FIND_IN_SET(?, approver_ids)", ticketID, userID).Count(&count).Error
	return count > 0, err
}

// ────────────────────────── 用户解析（角色展开） ──────────────────────────

// UserBrief 用户摘要
type UserBrief struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// FindUsersByIDs 按用户 ID 列表查用户（仅激活用户）
func (r *TicketRepository) FindUsersByIDs(ids []uint) ([]UserBrief, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []UserBrief
	err := r.db.Table("sys_users").
		Select("id, username, nickname").
		Where("id IN ? AND status = 'active'", ids).
		Find(&users).Error
	return users, err
}

// FindUsersByRoleIDs 按角色 ID 列表查拥有该角色的用户（仅激活用户）
func (r *TicketRepository) FindUsersByRoleIDs(roleIDs []uint) ([]UserBrief, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var users []UserBrief
	err := r.db.Table("sys_users u").
		Joins("JOIN sys_user_roles ur ON ur.user_id = u.id").
		Where("ur.role_id IN ? AND u.status = 'active'", roleIDs).
		Distinct().
		Select("u.id, u.username, u.nickname").
		Find(&users).Error
	return users, err
}
