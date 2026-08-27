package serviceticket

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	modelticket "oneops/backend3/model/ticket"
	repoticket "oneops/backend3/repository/ticket"
)

// TicketService 工单流程引擎
type TicketService struct {
	repo   *repoticket.TicketRepository
	wfRepo *repoticket.WorkflowRepository
}

func NewTicketService(repo *repoticket.TicketRepository, wfRepo *repoticket.WorkflowRepository) *TicketService {
	return &TicketService{repo: repo, wfRepo: wfRepo}
}

// WorkflowSnapshotNode 流程节点快照（工单发起时固化，流程后续修改不影响已有工单）
type WorkflowSnapshotNode struct {
	Key           string          `json:"key"`
	Name          string          `json:"name"`
	ApproverType  string          `json:"approverType"`
	ApproverIDs   []uint          `json:"approverIds"`
	ApproverNames []string        `json:"approverNames"`
	MultiType     string          `json:"multiType"`
	Condition     []ConditionItem `json:"condition"`
	SortOrder     int             `json:"sortOrder"`
}

// FormField 工单类型动态表单字段定义
type FormField struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"` // input/textarea/number/select/date
	Required    bool     `json:"required"`
	Options     []string `json:"options"`
	Placeholder string   `json:"placeholder"`
}

// TicketCreateRequest 发起工单请求
type TicketCreateRequest struct {
	TypeID     uint                   `json:"typeId" binding:"required"`
	WorkflowID uint                   `json:"workflowId" binding:"required"`
	Title      string                 `json:"title" binding:"required"`
	Priority   string                 `json:"priority"`
	FormData   map[string]interface{} `json:"formData"`
}

// TicketActionRequest 审批/撤销/评论请求
type TicketActionRequest struct {
	Comment string `json:"comment"`
}

// ────────────────────────── 发起工单 ──────────────────────────

// CreateTicket 发起工单并初始化流程
func (s *TicketService) CreateTicket(userID uint, username string, req TicketCreateRequest) (*modelticket.Ticket, error) {
	// 1. 校验类型与流程（流程必须归属所选场景）
	tp, err := s.wfRepo.FindTypeByID(req.TypeID)
	if err != nil {
		return nil, fmt.Errorf("工单类型不存在")
	}
	if tp.Status != 1 {
		return nil, fmt.Errorf("工单类型已停用")
	}
	wf, err := s.wfRepo.FindWorkflowByID(req.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("审批流程不存在")
	}
	if wf.TypeID != tp.ID {
		return nil, fmt.Errorf("审批流程不属于所选工单场景")
	}
	if wf.Status != 1 {
		return nil, fmt.Errorf("审批流程已停用")
	}
	if len(wf.Nodes) == 0 {
		return nil, fmt.Errorf("审批流程未配置审批节点")
	}

	// 2. 表单必填校验
	if err := s.validateFormData(tp.FormSchema, req.FormData); err != nil {
		return nil, err
	}

	// 3. 生成流程快照（解析审批人）
	snapshot, err := s.buildSnapshot(wf, userID, username)
	if err != nil {
		return nil, err
	}
	snapshotJSON, _ := json.Marshal(snapshot)

	formJSON, _ := json.Marshal(req.FormData)
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// 4. 生成单号
	ticketNo, err := s.generateTicketNo()
	if err != nil {
		return nil, err
	}

	ticket := &modelticket.Ticket{
		TicketNo:         ticketNo,
		Title:            req.Title,
		TypeID:           tp.ID,
		TypeName:         tp.Name,
		WorkflowID:       wf.ID,
		WorkflowName:     wf.Name,
		WorkflowSnapshot: string(snapshotJSON),
		Status:           "pending",
		Priority:         req.Priority,
		FormData:         string(formJSON),
		CreatorID:        userID,
		CreatorName:      username,
	}

	// 5. 落库 + 初始化节点记录 + 推进到首个节点（事务）
	err = s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		if err := tx.CreateTicket(ticket); err != nil {
			return err
		}
		records := make([]modelticket.TicketNodeRecord, 0, len(snapshot))
		for _, n := range snapshot {
			records = append(records, modelticket.TicketNodeRecord{
				TicketID:      ticket.ID,
				NodeKey:       n.Key,
				NodeName:      n.Name,
				Status:        "waiting",
				ApproverIDs:   joinUintIDs(n.ApproverIDs),
				ApproverNames: strings.Join(n.ApproverNames, ","),
				MultiType:     n.MultiType,
			})
		}
		if err := tx.CreateNodeRecords(records); err != nil {
			return err
		}
		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: "submit",
			OperatorID: userID, OperatorName: username,
			Comment: "发起工单",
		}); err != nil {
			return err
		}
		return s.advanceFlow(tx, ticket, snapshot, records)
	})
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

// buildSnapshot 解析每个节点的实际审批人（角色展开为用户）
func (s *TicketService) buildSnapshot(wf *modelticket.Workflow, creatorID uint, creatorName string) ([]WorkflowSnapshotNode, error) {
	snapshot := make([]WorkflowSnapshotNode, 0, len(wf.Nodes))
	for i, node := range wf.Nodes {
		snap := WorkflowSnapshotNode{
			Key:          fmt.Sprintf("n%d", i+1),
			Name:         node.Name,
			ApproverType: node.ApproverType,
			MultiType:    node.MultiType,
			Condition:    parseCondition(node.Condition),
			SortOrder:    i + 1,
		}
		switch node.ApproverType {
		case "initiator":
			snap.ApproverIDs = []uint{creatorID}
			snap.ApproverNames = []string{creatorName}
		case "role":
			roleIDs := parseUintIDs(node.ApproverIDs)
			users, err := s.repo.FindUsersByRoleIDs(roleIDs)
			if err != nil {
				return nil, fmt.Errorf("解析节点「%s」审批角色失败: %v", node.Name, err)
			}
			if len(users) == 0 {
				return nil, fmt.Errorf("节点「%s」指定的角色下没有可用的审批用户", node.Name)
			}
			for _, u := range users {
				snap.ApproverIDs = append(snap.ApproverIDs, u.ID)
				snap.ApproverNames = append(snap.ApproverNames, displayName(u))
			}
		default: // user
			ids := parseUintIDs(node.ApproverIDs)
			users, err := s.repo.FindUsersByIDs(ids)
			if err != nil {
				return nil, fmt.Errorf("解析节点「%s」审批人失败: %v", node.Name, err)
			}
			if len(users) == 0 {
				return nil, fmt.Errorf("节点「%s」指定的审批用户不存在或已禁用", node.Name)
			}
			for _, u := range users {
				snap.ApproverIDs = append(snap.ApproverIDs, u.ID)
				snap.ApproverNames = append(snap.ApproverNames, displayName(u))
			}
		}
		snapshot = append(snapshot, snap)
	}
	return snapshot, nil
}

// advanceFlow 从当前节点之后寻找下一个满足条件的节点并激活；
// 不满足条件的节点标记 skipped；全部走完则工单通过。
func (s *TicketService) advanceFlow(tx *repoticket.TicketRepository, ticket *modelticket.Ticket,
	snapshot []WorkflowSnapshotNode, records []modelticket.TicketNodeRecord) error {

	var formData map[string]interface{}
	_ = json.Unmarshal([]byte(ticket.FormData), &formData)

	now := time.Now()
	// 定位当前节点在快照中的下标（currentNodeKey 为空表示从头开始）
	pos := -1
	if ticket.CurrentNodeKey != "" {
		for i, n := range snapshot {
			if n.Key == ticket.CurrentNodeKey {
				pos = i
				break
			}
		}
	}

	for i := pos + 1; i < len(snapshot); i++ {
		node := snapshot[i]
		if !conditionMatch(formData, node.Condition) {
			// 条件不满足：跳过该节点
			for j := range records {
				if records[j].NodeKey == node.Key {
					records[j].Status = "skipped"
					records[j].FinishedAt = &now
					if err := tx.UpdateNodeRecord(&records[j]); err != nil {
						return err
					}
				}
			}
			_ = tx.CreateFlowLog(modelticket.TicketFlowLog{
				TicketID: ticket.ID, Action: "skip", NodeName: node.Name,
				OperatorID: 0, OperatorName: "系统",
				Comment: "条件不满足，自动跳过",
			})
			continue
		}
		// 激活该节点
		nowPtr := now
		for j := range records {
			if records[j].NodeKey == node.Key {
				records[j].Status = "pending"
				records[j].StartedAt = &nowPtr
				if err := tx.UpdateNodeRecord(&records[j]); err != nil {
					return err
				}
			}
		}
		return tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
			"status":            "pending",
			"current_node_key":  node.Key,
			"current_node_name": node.Name,
		})
	}

	// 没有更多节点：流程结束
	finished := time.Now()
	return tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
		"status":            "approved",
		"current_node_key":  "",
		"current_node_name": "",
		"finished_at":       finished,
	})
}

// ────────────────────────── 审批操作 ──────────────────────────

// Approve 审批通过
func (s *TicketService) Approve(ticketID, userID uint, username, comment string) error {
	return s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		ticket, snapshot, records, err := s.loadTicketForAction(tx, ticketID)
		if err != nil {
			return err
		}

		rec := findCurrentRecord(records, ticket.CurrentNodeKey)
		if rec == nil || rec.Status != "pending" {
			return fmt.Errorf("工单当前节点状态异常")
		}
		approverIDs := parseUintIDs(rec.ApproverIDs)
		approvedIDs := parseUintIDs(rec.ApprovedIDs)
		if !containsUint(approverIDs, userID) {
			return fmt.Errorf("您不是当前节点的审批人")
		}
		if containsUint(approvedIDs, userID) {
			return fmt.Errorf("您已审批过该节点")
		}
		approvedIDs = append(approvedIDs, userID)

		// 判断节点是否完成：或签任一通过即完成；会签需全部通过
		completed := rec.MultiType == "all" && len(approvedIDs) >= len(approverIDs) ||
			rec.MultiType != "all"

		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: "approve", NodeName: rec.NodeName,
			OperatorID: userID, OperatorName: username, Comment: comment,
		}); err != nil {
			return err
		}

		if !completed {
			rec.ApprovedIDs = joinUintIDs(approvedIDs)
			rec.Comment = comment
			return tx.UpdateNodeRecord(rec)
		}

		now := time.Now()
		rec.ApprovedIDs = joinUintIDs(approvedIDs)
		rec.Comment = comment
		rec.Status = "approved"
		rec.FinishedAt = &now
		if err := tx.UpdateNodeRecord(rec); err != nil {
			return err
		}
		return s.advanceFlow(tx, ticket, snapshot, records)
	})
}

// Reject 驳回（终止工单）
func (s *TicketService) Reject(ticketID, userID uint, username, comment string) error {
	return s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		ticket, _, records, err := s.loadTicketForAction(tx, ticketID)
		if err != nil {
			return err
		}

		rec := findCurrentRecord(records, ticket.CurrentNodeKey)
		if rec == nil || rec.Status != "pending" {
			return fmt.Errorf("工单当前节点状态异常")
		}
		if !containsUint(parseUintIDs(rec.ApproverIDs), userID) {
			return fmt.Errorf("您不是当前节点的审批人")
		}

		now := time.Now()
		rec.Status = "rejected"
		rec.Comment = comment
		rec.FinishedAt = &now
		if err := tx.UpdateNodeRecord(rec); err != nil {
			return err
		}
		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: "reject", NodeName: rec.NodeName,
			OperatorID: userID, OperatorName: username, Comment: comment,
		}); err != nil {
			return err
		}
		return tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
			"status":            "rejected",
			"current_node_key":  "",
			"current_node_name": "",
			"finished_at":       now,
		})
	})
}

// Cancel 发起人撤销工单
func (s *TicketService) Cancel(ticketID, userID uint, username, comment string) error {
	return s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		ticket, _, _, err := s.loadTicketForAction(tx, ticketID)
		if err != nil {
			return err
		}
		if ticket.CreatorID != userID {
			return fmt.Errorf("只有发起人可以撤销工单")
		}
		now := time.Now()
		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: "cancel",
			OperatorID: userID, OperatorName: username, Comment: comment,
		}); err != nil {
			return err
		}
		return tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
			"status":            "canceled",
			"current_node_key":  "",
			"current_node_name": "",
			"finished_at":       now,
		})
	})
}

// Comment 添加评论
func (s *TicketService) Comment(ticketID, userID uint, username, comment string) error {
	ticket, err := s.repo.FindTicketByID(ticketID)
	if err != nil {
		return fmt.Errorf("工单不存在")
	}
	_ = ticket
	return s.repo.CreateFlowLog(modelticket.TicketFlowLog{
		TicketID: ticketID, Action: "comment",
		OperatorID: userID, OperatorName: username, Comment: comment,
	})
}

// ────────────────────────── 查询 ──────────────────────────

// TicketListItem 列表项（附带当前审批人信息）
type TicketListItem struct {
	modelticket.Ticket
	CurrentApprovers string `json:"currentApprovers"` // 当前节点待审批人（逗号分隔）
	CanApprove       bool   `json:"canApprove"`       // 当前用户是否可审批
}

// GetTickets 分页查询工单
func (s *TicketService) GetTickets(q repoticket.TicketQuery) ([]TicketListItem, int64, error) {
	if q.Scope == "" {
		q.Scope = "created"
	}
	list, total, err := s.repo.FindTickets(q)
	if err != nil {
		return nil, 0, err
	}

	items := make([]TicketListItem, 0, len(list))
	// 批量取当前待审批节点信息
	pendingIDs := make([]uint, 0, len(list))
	for _, t := range list {
		if t.Status == "pending" {
			pendingIDs = append(pendingIDs, t.ID)
		}
	}
	pendingRecords, _ := s.repo.FindPendingRecordsByTicketIDs(pendingIDs)
	pendingMap := make(map[uint]modelticket.TicketNodeRecord, len(pendingRecords))
	for _, rec := range pendingRecords {
		pendingMap[rec.TicketID] = rec
	}
	for _, t := range list {
		item := TicketListItem{Ticket: t}
		if rec, ok := pendingMap[t.ID]; ok {
			item.CurrentApprovers = rec.ApproverNames
			ids := parseUintIDs(rec.ApproverIDs)
			approved := parseUintIDs(rec.ApprovedIDs)
			item.CanApprove = containsUint(ids, q.UserID) && !containsUint(approved, q.UserID)
		}
		items = append(items, item)
	}
	return items, total, nil
}

// TicketDetail 工单详情
type TicketDetail struct {
	Ticket     modelticket.Ticket             `json:"ticket"`
	FormSchema []FormField                    `json:"formSchema"`
	FormData   map[string]interface{}         `json:"formData"`
	Nodes      []modelticket.TicketNodeRecord `json:"nodes"`
	Logs       []modelticket.TicketFlowLog    `json:"logs"`
	CanApprove bool                           `json:"canApprove"`
	CanCancel  bool                           `json:"canCancel"`
}

// GetTicketDetail 工单详情（含节点进度与流转日志）
func (s *TicketService) GetTicketDetail(ticketID, userID uint) (*TicketDetail, error) {
	ticket, err := s.repo.FindTicketByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("工单不存在")
	}

	detail := &TicketDetail{Ticket: *ticket}

	// 表单 schema（类型可能已删除，容错）
	if tp, err := s.wfRepo.FindTypeByID(ticket.TypeID); err == nil {
		detail.FormSchema = parseFormSchema(tp.FormSchema)
	}
	_ = json.Unmarshal([]byte(ticket.FormData), &detail.FormData)

	if detail.Nodes, err = s.repo.FindNodeRecords(ticketID); err != nil {
		return nil, err
	}
	if detail.Logs, err = s.repo.FindFlowLogs(ticketID); err != nil {
		return nil, err
	}

	detail.CanCancel = ticket.Status == "pending" && ticket.CreatorID == userID
	for _, rec := range detail.Nodes {
		if rec.NodeKey == ticket.CurrentNodeKey && rec.Status == "pending" {
			ids := parseUintIDs(rec.ApproverIDs)
			approved := parseUintIDs(rec.ApprovedIDs)
			detail.CanApprove = containsUint(ids, userID) && !containsUint(approved, userID)
			break
		}
	}
	return detail, nil
}

// CanView 工单可见性：发起人 / 任一节点审批人 / 参与过评论审批 / 拥有 ticket.ticket.list 权限
func (s *TicketService) CanView(ticketID, userID uint) bool {
	ticket, err := s.repo.FindTicketByID(ticketID)
	if err != nil {
		return false
	}
	if ticket.CreatorID == userID {
		return true
	}
	if ok, _ := s.repo.UserIsNodeApprover(ticketID, userID); ok {
		return true
	}
	if ok, _ := s.repo.UserTouchedTicket(ticketID, userID); ok {
		return true
	}
	return false
}

// ────────────────────────── 内部工具 ──────────────────────────

// loadTicketForAction 加载工单 + 快照 + 节点记录（审批前置）
func (s *TicketService) loadTicketForAction(tx *repoticket.TicketRepository, ticketID uint) (
	*modelticket.Ticket, []WorkflowSnapshotNode, []modelticket.TicketNodeRecord, error) {

	ticket, err := tx.FindTicketByID(ticketID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("工单不存在")
	}
	if ticket.Status != "pending" {
		return nil, nil, nil, fmt.Errorf("工单已结束，无法操作")
	}
	var snapshot []WorkflowSnapshotNode
	if err := json.Unmarshal([]byte(ticket.WorkflowSnapshot), &snapshot); err != nil {
		return nil, nil, nil, fmt.Errorf("流程快照解析失败")
	}
	records, err := tx.FindNodeRecords(ticketID)
	if err != nil {
		return nil, nil, nil, err
	}
	return ticket, snapshot, records, nil
}

// generateTicketNo 生成工单单号：T + yyyymmdd + 4位序号
func (s *TicketService) generateTicketNo() (string, error) {
	prefix := "T" + time.Now().Format("20060102")
	count, err := s.repo.CountTodayTickets(prefix)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

// validateFormData 按类型表单 schema 校验必填项
func (s *TicketService) validateFormData(schemaJSON string, data map[string]interface{}) error {
	fields := parseFormSchema(schemaJSON)
	for _, f := range fields {
		if !f.Required {
			continue
		}
		v, ok := data[f.Key]
		if !ok || v == nil {
			return fmt.Errorf("表单项「%s」为必填", f.Label)
		}
		if str, ok := v.(string); ok && strings.TrimSpace(str) == "" {
			return fmt.Errorf("表单项「%s」为必填", f.Label)
		}
	}
	return nil
}

// conditionMatch 条件求值：全部满足才激活（空条件恒真）
func conditionMatch(data map[string]interface{}, conds []ConditionItem) bool {
	if len(conds) == 0 {
		return true
	}
	for _, c := range conds {
		actual := fmt.Sprintf("%v", data[c.Field])
		switch c.Op {
		case "eq":
			if actual != c.Value {
				return false
			}
		case "ne":
			if actual == c.Value {
				return false
			}
		case "in":
			found := false
			for _, v := range strings.Split(c.Value, ",") {
				if actual == strings.TrimSpace(v) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "gt":
			if toFloat(actual) <= toFloat(c.Value) {
				return false
			}
		case "lt":
			if toFloat(actual) >= toFloat(c.Value) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func toFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func findCurrentRecord(records []modelticket.TicketNodeRecord, nodeKey string) *modelticket.TicketNodeRecord {
	for i := range records {
		if records[i].NodeKey == nodeKey {
			return &records[i]
		}
	}
	return nil
}

func parseUintIDs(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.ParseUint(p, 10, 64); err == nil {
			ids = append(ids, uint(v))
		}
	}
	return ids
}

func containsUint(ids []uint, target uint) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func parseCondition(s string) []ConditionItem {
	if s == "" {
		return nil
	}
	var conds []ConditionItem
	if err := json.Unmarshal([]byte(s), &conds); err != nil {
		return nil
	}
	return conds
}

func parseFormSchema(s string) []FormField {
	if s == "" {
		return nil
	}
	var fields []FormField
	if err := json.Unmarshal([]byte(s), &fields); err != nil {
		return nil
	}
	return fields
}

func displayName(u repoticket.UserBrief) string {
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Username
}
