package serviceticket

import (
	"errors"
	"fmt"

	modelticket "oneops/backend3/model/ticket"
	repoticket "oneops/backend3/repository/ticket"

	"gorm.io/gorm"
)

// WorkflowService 工单流程定义/工单类型管理
type WorkflowService struct {
	repo *repoticket.WorkflowRepository
}

func NewWorkflowService(repo *repoticket.WorkflowRepository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

// WorkflowSaveRequest 流程保存请求（含节点全量）
type WorkflowSaveRequest struct {
	TypeID      uint               `json:"typeId" binding:"required"`
	Name        string             `json:"name" binding:"required"`
	Code        string             `json:"code" binding:"required"`
	Description string             `json:"description"`
	Status      int                `json:"status"`
	Nodes       []WorkflowNodeSave `json:"nodes"`
}

// WorkflowNodeSave 节点保存请求
type WorkflowNodeSave struct {
	Name         string          `json:"name" binding:"required"`
	ApproverType string          `json:"approverType" binding:"required,oneof=user role initiator"`
	ApproverIDs  []uint          `json:"approverIds"`
	MultiType    string          `json:"multiType"`
	Condition    []ConditionItem `json:"condition"`
	TimeoutHours int             `json:"timeoutHours" binding:"gte=0,lte=8760"` // 审批超时阈值（小时），0=不启用
}

// ConditionItem 节点激活条件：工单表单字段比较
type ConditionItem struct {
	Field string `json:"field"`
	Op    string `json:"op"` // eq/ne/in/gt/lt
	Value string `json:"value"`
}

// GetWorkflows 分页查询流程（typeID>0 时按场景过滤）
func (s *WorkflowService) GetWorkflows(keyword string, status int, typeID uint, offset, limit int) ([]modelticket.Workflow, int64, error) {
	return s.repo.FindWorkflows(keyword, status, typeID, offset, limit)
}

// GetEnabledWorkflowsByType 某场景下启用的流程（发起工单选择用，仅需登录）
func (s *WorkflowService) GetEnabledWorkflowsByType(typeID uint) ([]modelticket.Workflow, error) {
	return s.repo.FindEnabledWorkflowsByTypeID(typeID)
}

// GetWorkflow 流程详情（含节点）
func (s *WorkflowService) GetWorkflow(id uint) (*modelticket.Workflow, error) {
	return s.repo.FindWorkflowByID(id)
}

// CreateWorkflow 创建流程
func (s *WorkflowService) CreateWorkflow(req WorkflowSaveRequest) error {
	if _, err := s.repo.FindWorkflowByCode(req.Code); err == nil {
		return fmt.Errorf("流程编码 %s 已存在", req.Code)
	}
	if _, err := s.repo.FindTypeByID(req.TypeID); err != nil {
		return fmt.Errorf("所属工单场景不存在")
	}
	wf, err := s.buildWorkflow(0, req)
	if err != nil {
		return err
	}
	return s.repo.CreateWorkflow(wf)
}

// UpdateWorkflow 更新流程（全量替换节点；已发起工单走快照不受影响）
func (s *WorkflowService) UpdateWorkflow(id uint, req WorkflowSaveRequest) error {
	if _, err := s.repo.FindWorkflowByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("流程不存在")
		}
		return err
	}
	if _, err := s.repo.FindTypeByID(req.TypeID); err != nil {
		return fmt.Errorf("所属工单场景不存在")
	}
	wf, err := s.buildWorkflow(id, req)
	if err != nil {
		return err
	}
	return s.repo.UpdateWorkflow(wf)
}

// DeleteWorkflow 删除流程（被工单引用时禁止）
func (s *WorkflowService) DeleteWorkflow(id uint) error {
	count, err := s.repo.CountTicketsByWorkflowID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该流程已有 %d 个工单记录，无法删除", count)
	}
	return s.repo.DeleteWorkflow(id)
}

// buildWorkflow 组装流程实体（校验节点合法性）
func (s *WorkflowService) buildWorkflow(id uint, req WorkflowSaveRequest) (*modelticket.Workflow, error) {
	if len(req.Nodes) == 0 {
		return nil, fmt.Errorf("流程至少需要一个审批节点")
	}
	wf := &modelticket.Workflow{
		ID:          id,
		TypeID:      req.TypeID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}
	for i, n := range req.Nodes {
		if n.MultiType == "" {
			n.MultiType = "any"
		}
		if n.ApproverType == "user" && len(n.ApproverIDs) == 0 {
			return nil, fmt.Errorf("节点「%s」未指定审批用户", n.Name)
		}
		if n.ApproverType == "role" && len(n.ApproverIDs) == 0 {
			return nil, fmt.Errorf("节点「%s」未指定审批角色", n.Name)
		}
		node := modelticket.WorkflowNode{
			WorkflowID:   id,
			Name:         n.Name,
			ApproverType: n.ApproverType,
			ApproverIDs:  joinUintIDs(n.ApproverIDs),
			MultiType:    n.MultiType,
			TimeoutHours: n.TimeoutHours,
			SortOrder:    i + 1,
		}
		wf.Nodes = append(wf.Nodes, node)
	}
	return wf, nil
}

// ────────────────────────── 工单类型 ──────────────────────────

// TypeSaveRequest 类型保存请求
type TypeSaveRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	FormSchema  string `json:"formSchema"`
	Status      int    `json:"status"`
}

// GetTypes 分页查询工单类型
func (s *WorkflowService) GetTypes(keyword string, status int, offset, limit int) ([]modelticket.TicketType, int64, error) {
	return s.repo.FindTypes(keyword, status, offset, limit)
}

// GetEnabledTypes 启用的类型选项
func (s *WorkflowService) GetEnabledTypes() ([]modelticket.TicketType, error) {
	return s.repo.FindEnabledTypes()
}

// GetType 类型详情
func (s *WorkflowService) GetType(id uint) (*modelticket.TicketType, error) {
	return s.repo.FindTypeByID(id)
}

// CreateType 创建工单类型
func (s *WorkflowService) CreateType(req TypeSaveRequest) error {
	t := &modelticket.TicketType{
		Name:        req.Name,
		Code:        req.Code,
		Icon:        req.Icon,
		Description: req.Description,
		FormSchema:  req.FormSchema,
		Status:      req.Status,
	}
	return s.repo.CreateType(t)
}

// UpdateType 更新工单类型
func (s *WorkflowService) UpdateType(id uint, req TypeSaveRequest) error {
	if _, err := s.repo.FindTypeByID(id); err != nil {
		return fmt.Errorf("工单类型不存在")
	}
	t := &modelticket.TicketType{
		ID:          id,
		Name:        req.Name,
		Code:        req.Code,
		Icon:        req.Icon,
		Description: req.Description,
		FormSchema:  req.FormSchema,
		Status:      req.Status,
	}
	return s.repo.UpdateType(t)
}

// DeleteType 删除工单类型（有工单或流程归属时禁止）
func (s *WorkflowService) DeleteType(id uint) error {
	count, err := s.repo.CountTicketsByTypeID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该类型下已有 %d 个工单，无法删除", count)
	}
	wfCount, err := s.repo.CountWorkflowsByTypeID(id)
	if err != nil {
		return err
	}
	if wfCount > 0 {
		return fmt.Errorf("该场景下还有 %d 个审批流程，请先删除或迁移流程", wfCount)
	}
	return s.repo.DeleteType(id)
}

// joinUintIDs 将 uint 列表转为逗号分隔字符串
func joinUintIDs(ids []uint) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("%d", id)
	}
	return out
}
