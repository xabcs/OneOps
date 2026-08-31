package modelticket

import "time"

// TicketType 工单类型（工单场景）→ 表 ticket_types
// 定义一类工单场景（如通用申请、SQL 审核、应用发布）：
// 一份动态表单 Schema（JSON 字段数组）+ 场景下可用的多个审批流程（Workflow.TypeID 反向归属），
// FormSchema 格式: [{"key":"sql","label":"SQL内容","type":"textarea","required":true,"options":[...]}]
// WorkflowCount: 该场景下启用的审批流程数量（列表查询时批量填充，非表列；0=发起工单时无流程可选）
type TicketType struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"size:100;not null"`
	Code          string    `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Icon          string    `json:"icon" gorm:"size:100"`
	Description   string    `json:"description" gorm:"size:500"`
	FormSchema    string    `json:"formSchema" gorm:"type:longtext"`
	Status        int       `json:"status" gorm:"default:1"` // 1启用 0禁用
	WorkflowCount int       `json:"workflowCount" gorm:"-"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (TicketType) TableName() string { return "ticket_types" }

// Ticket 工单 → 表 tickets
// Status     : pending=审批中 / approved=已通过 / rejected=已驳回 / canceled=已撤销
// FormData   : 发起时填写的动态表单数据（JSON object）
// WorkflowSnapshot: 发起时的流程节点定义快照（JSON array，见 modelticket.WorkflowSnapshotNode）
// LastUrgeAt : 发起人最近一次催办时间（用于催办频控，见 UrgeCooldown）
type Ticket struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	TicketNo         string     `json:"ticketNo" gorm:"uniqueIndex;size:50;not null"`
	Title            string     `json:"title" gorm:"size:200;not null"`
	TypeID           uint       `json:"typeId" gorm:"index"`
	TypeName         string     `json:"typeName" gorm:"size:100"`
	WorkflowID       uint       `json:"workflowId"`
	WorkflowName     string     `json:"workflowName" gorm:"size:100"`
	WorkflowSnapshot string     `json:"workflowSnapshot" gorm:"type:longtext"`
	CurrentNodeKey   string     `json:"currentNodeKey" gorm:"size:50;index"` // 当前节点 key，空=已结束
	CurrentNodeName  string     `json:"currentNodeName" gorm:"size:100"`
	Status           string     `json:"status" gorm:"size:20;default:pending;index"`
	Priority         string     `json:"priority" gorm:"size:20;default:normal"` // low/normal/high/urgent
	FormData         string     `json:"formData" gorm:"type:longtext"`
	CreatorID        uint       `json:"creatorId" gorm:"index"`
	CreatorName      string     `json:"creatorName" gorm:"size:50"`
	FinishedAt       *time.Time `json:"finishedAt"`
	LastUrgeAt       *time.Time `json:"lastUrgeAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (Ticket) TableName() string { return "tickets" }

// TicketNodeRecord 工单节点审批记录 → 表 ticket_node_records
// 每个工单每经过一个节点生成一条记录（跳过的节点记为 skipped）。
// Status      : waiting=未到达 / pending=审批中 / approved=通过 / rejected=驳回 / skipped=条件跳过 / canceled=工单终止未走完
// ApproverIDs : 该节点实际可审批的用户 ID（逗号分隔，发起时已按角色解析展开）
// ApprovedIDs : 已通过该节点的用户 ID（逗号分隔，会签进度）
// TimeoutNotifiedAt / EscalationNotifiedAt: 超时升级链的防重发时间戳（蓝图④），
// 超时提醒每 24h 重复一次，升级通知只发一次
type TicketNodeRecord struct {
	ID                   uint       `json:"id" gorm:"primaryKey"`
	TicketID             uint       `json:"ticketId" gorm:"index;not null"`
	NodeKey              string     `json:"nodeKey" gorm:"size:50;index"`
	NodeName             string     `json:"nodeName" gorm:"size:100"`
	Status               string     `json:"status" gorm:"size:20;default:waiting"`
	ApproverIDs          string     `json:"approverIds" gorm:"size:1000"`
	ApproverNames        string     `json:"approverNames" gorm:"size:2000"`
	ApprovedIDs          string     `json:"approvedIds" gorm:"size:1000"`
	MultiType            string     `json:"multiType" gorm:"size:20;default:any"`
	Comment              string     `json:"comment" gorm:"size:1000"`
	StartedAt            *time.Time `json:"startedAt"`
	FinishedAt           *time.Time `json:"finishedAt"`
	TimeoutNotifiedAt    *time.Time `json:"timeoutNotifiedAt"`
	EscalationNotifiedAt *time.Time `json:"escalationNotifiedAt"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func (TicketNodeRecord) TableName() string { return "ticket_node_records" }

// TicketFlowLog 工单流转日志 → 表 ticket_flow_logs
// Action: submit=发起 / approve=通过 / reject=驳回 / cancel=撤销 / comment=评论 / skip=节点跳过 / auto=流转
type TicketFlowLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	TicketID     uint      `json:"ticketId" gorm:"index;not null"`
	Action       string    `json:"action" gorm:"size:20"`
	NodeName     string    `json:"nodeName" gorm:"size:100"`
	OperatorID   uint      `json:"operatorId"`
	OperatorName string    `json:"operatorName" gorm:"size:50"`
	Comment      string    `json:"comment" gorm:"type:text"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (TicketFlowLog) TableName() string { return "ticket_flow_logs" }
