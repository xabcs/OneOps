package repositoryticket

import (
	modelticket "oneops/backend3/model/ticket"

	"gorm.io/gorm"
)

// WorkflowRepository 工单流程/类型数据访问
type WorkflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// ────────────────────────── 流程定义 ──────────────────────────

// FindWorkflows 分页查询流程定义（typeID>0 时按场景过滤）
func (r *WorkflowRepository) FindWorkflows(keyword string, status int, typeID uint, offset, limit int) ([]modelticket.Workflow, int64, error) {
	query := r.db.Model(&modelticket.Workflow{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if typeID > 0 {
		query = query.Where("type_id = ?", typeID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []modelticket.Workflow
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	r.fillNodeCounts(list)
	return list, total, nil
}

// fillNodeCounts 批量填充流程的审批节点数量（单条 COUNT 查询，避免逐行子查询）
func (r *WorkflowRepository) fillNodeCounts(list []modelticket.Workflow) {
	if len(list) == 0 {
		return
	}
	ids := make([]uint, len(list))
	for i, wf := range list {
		ids[i] = wf.ID
	}
	var rows []struct {
		WorkflowID uint
		Cnt        int
	}
	if err := r.db.Table("ticket_workflow_nodes").
		Select("workflow_id, COUNT(*) AS cnt").
		Where("workflow_id IN ?", ids).
		Group("workflow_id").Scan(&rows).Error; err != nil {
		return
	}
	counts := make(map[uint]int, len(rows))
	for _, row := range rows {
		counts[row.WorkflowID] = row.Cnt
	}
	for i := range list {
		list[i].NodeCount = counts[list[i].ID]
	}
}

// FindEnabledWorkflowsByTypeID 某场景下启用的流程（发起工单选择用，含节点数）
func (r *WorkflowRepository) FindEnabledWorkflowsByTypeID(typeID uint) ([]modelticket.Workflow, error) {
	var list []modelticket.Workflow
	err := r.db.Preload("Nodes").Where("type_id = ? AND status = 1", typeID).Order("id ASC").Find(&list).Error
	return list, err
}

// FindWorkflowByID 按 ID 查流程（含节点）
func (r *WorkflowRepository) FindWorkflowByID(id uint) (*modelticket.Workflow, error) {
	var wf modelticket.Workflow
	if err := r.db.Preload("Nodes", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).First(&wf, id).Error; err != nil {
		return nil, err
	}
	return &wf, nil
}

// FindWorkflowByCode 按 code 查流程
func (r *WorkflowRepository) FindWorkflowByCode(code string) (*modelticket.Workflow, error) {
	var wf modelticket.Workflow
	if err := r.db.Where("code = ?", code).First(&wf).Error; err != nil {
		return nil, err
	}
	return &wf, nil
}

// CreateWorkflow 创建流程（含节点）
func (r *WorkflowRepository) CreateWorkflow(wf *modelticket.Workflow) error {
	return r.db.Create(wf).Error
}

// UpdateWorkflow 更新流程（事务内全量替换节点）
func (r *WorkflowRepository) UpdateWorkflow(wf *modelticket.Workflow) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&modelticket.Workflow{}).Where("id = ?", wf.ID).Updates(map[string]interface{}{
			"type_id":     wf.TypeID,
			"name":        wf.Name,
			"description": wf.Description,
			"status":      wf.Status,
			"version":     gorm.Expr("version + 1"),
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("workflow_id = ?", wf.ID).Delete(&modelticket.WorkflowNode{}).Error; err != nil {
			return err
		}
		if len(wf.Nodes) > 0 {
			for i := range wf.Nodes {
				wf.Nodes[i].ID = 0
				wf.Nodes[i].WorkflowID = wf.ID
			}
			return tx.Create(&wf.Nodes).Error
		}
		return nil
	})
}

// DeleteWorkflow 删除流程（节点级联删除）
func (r *WorkflowRepository) DeleteWorkflow(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", id).Delete(&modelticket.WorkflowNode{}).Error; err != nil {
			return err
		}
		return tx.Delete(&modelticket.Workflow{}, id).Error
	})
}

// CountTicketsByWorkflowID 统计引用某流程的工单数（在途/历史工单走快照，删除流程前需确认无引用）
func (r *WorkflowRepository) CountTicketsByWorkflowID(workflowID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelticket.Ticket{}).Where("workflow_id = ?", workflowID).Count(&count).Error
	return count, err
}

// CountWorkflowsByTypeID 统计某场景下的流程数
func (r *WorkflowRepository) CountWorkflowsByTypeID(typeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelticket.Workflow{}).Where("type_id = ?", typeID).Count(&count).Error
	return count, err
}

// ────────────────────────── 工单类型 ──────────────────────────

// FindTypes 分页查询工单类型
func (r *WorkflowRepository) FindTypes(keyword string, status int, offset, limit int) ([]modelticket.TicketType, int64, error) {
	query := r.db.Model(&modelticket.TicketType{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []modelticket.TicketType
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	r.fillWorkflowCounts(list)
	return list, total, nil
}

// fillWorkflowCounts 批量填充场景下启用的审批流程数量（单条 COUNT 查询；0=发起工单时无流程可选）
func (r *WorkflowRepository) fillWorkflowCounts(list []modelticket.TicketType) {
	if len(list) == 0 {
		return
	}
	ids := make([]uint, len(list))
	for i, t := range list {
		ids[i] = t.ID
	}
	var rows []struct {
		TypeID uint
		Cnt    int
	}
	if err := r.db.Table("ticket_workflows").
		Select("type_id, COUNT(*) AS cnt").
		Where("type_id IN ? AND status = 1", ids).
		Group("type_id").Scan(&rows).Error; err != nil {
		return
	}
	counts := make(map[uint]int, len(rows))
	for _, row := range rows {
		counts[row.TypeID] = row.Cnt
	}
	for i := range list {
		list[i].WorkflowCount = counts[list[i].ID]
	}
}

// FindEnabledTypes 启用的类型（发起工单下拉用）
func (r *WorkflowRepository) FindEnabledTypes() ([]modelticket.TicketType, error) {
	var list []modelticket.TicketType
	err := r.db.Where("status = 1").Order("id ASC").Find(&list).Error
	return list, err
}

// FindTypeByID 按 ID 查类型
func (r *WorkflowRepository) FindTypeByID(id uint) (*modelticket.TicketType, error) {
	var t modelticket.TicketType
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateType 创建类型
func (r *WorkflowRepository) CreateType(t *modelticket.TicketType) error {
	return r.db.Create(t).Error
}

// UpdateType 更新类型
func (r *WorkflowRepository) UpdateType(t *modelticket.TicketType) error {
	return r.db.Model(t).Updates(map[string]interface{}{
		"name":        t.Name,
		"icon":        t.Icon,
		"description": t.Description,
		"form_schema": t.FormSchema,
		"status":      t.Status,
	}).Error
}

// DeleteType 删除类型
func (r *WorkflowRepository) DeleteType(id uint) error {
	return r.db.Delete(&modelticket.TicketType{}, id).Error
}

// CountTicketsByTypeID 统计某类型下工单数
func (r *WorkflowRepository) CountTicketsByTypeID(typeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelticket.Ticket{}).Where("type_id = ?", typeID).Count(&count).Error
	return count, err
}
