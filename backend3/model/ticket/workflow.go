package modelticket

import "time"

// Workflow 流程定义 → 表 ticket_workflows
// 一个流程定义由有序的审批节点组成，工单发起时对节点做快照，
// 后续流程修改不影响已发起的工单。
// TypeID: 归属的工单场景（TicketType）；同一场景可配置多个流程，发起时选用。
type Workflow struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	TypeID      uint           `json:"typeId" gorm:"index;not null;default:0"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Code        string         `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Description string         `json:"description" gorm:"size:500"`
	Status      int            `json:"status" gorm:"default:1"` // 1启用 0禁用
	Version     int            `json:"version" gorm:"default:1"`
	NodeCount   int            `json:"nodeCount" gorm:"-"` // 审批节点数量（列表查询时批量填充，非表列）
	Nodes       []WorkflowNode `json:"nodes,omitempty" gorm:"foreignKey:WorkflowID"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

func (Workflow) TableName() string { return "ticket_workflows" }

// WorkflowNode 流程审批节点 → 表 ticket_workflow_nodes
//
// ApproverType: user=指定用户 / role=指定角色（该角色下所有用户）/ initiator=发起人（用于自确认节点）
// ApproverIDs : 逗号分隔的 ID 列表（"1,2,3"）；role 类型时为角色 ID 列表
// MultiType   : any=或签（任一人通过即过） / all=会签（全部通过才过）
// Condition   : 可选 JSON 条件数组（工单表单字段满足才激活该节点），如：
//
//	[{"field":"priority","op":"eq","value":"urgent"}]
//
// TimeoutHours: 审批超时阈值（小时）。0=不启用；>0 时节点停留超过阈值自动提醒
// 待审批人（timeout 事件），超过 2 倍阈值升级通知管理员与发起人（escalation 事件），
// 之后每 24 小时重复提醒一次（蓝图④超时升级链）。
type WorkflowNode struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	WorkflowID    uint      `json:"workflowId" gorm:"index;not null"`
	Name          string    `json:"name" gorm:"size:100;not null"`
	ApproverType  string    `json:"approverType" gorm:"size:20;default:user"`
	ApproverIDs   string    `json:"approverIds" gorm:"size:500"`
	ApproverNames string    `json:"approverNames" gorm:"size:1000"` // 展示用名称快照（逗号分隔）
	MultiType     string    `json:"multiType" gorm:"size:20;default:any"`
	Condition     string    `json:"condition" gorm:"type:text"`
	TimeoutHours  int       `json:"timeoutHours" gorm:"default:0"`
	SortOrder     int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (WorkflowNode) TableName() string { return "ticket_workflow_nodes" }
