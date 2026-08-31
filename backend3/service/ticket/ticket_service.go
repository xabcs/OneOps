package serviceticket

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
	notify *TicketNotifier // 事件通知（可为 nil，表示禁用通知）
}

func NewTicketService(repo *repoticket.TicketRepository, wfRepo *repoticket.WorkflowRepository, notify *TicketNotifier) *TicketService {
	return &TicketService{repo: repo, wfRepo: wfRepo, notify: notify}
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
	TimeoutHours  int             `json:"timeoutHours"` // 审批超时阈值（小时），0=不启用（超时升级链读取）
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
	Priority   string                 `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
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

	// 2. 表单必填校验（schema 格式错误时报错阻断，而非静默跳过校验）
	if err := s.validateFormData(tp.FormSchema, req.FormData); err != nil {
		return nil, err
	}

	// 3. 生成流程快照（解析审批人）
	snapshot, err := s.buildSnapshot(wf, userID, username)
	if err != nil {
		return nil, err
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("生成流程快照失败: %v", err)
	}
	formJSON, err := json.Marshal(req.FormData)
	if err != nil {
		return nil, fmt.Errorf("序列化表单数据失败: %v", err)
	}
	if req.Priority == "" {
		req.Priority = modelticket.PriorityNormal
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
		Status:           modelticket.TicketStatusPending,
		Priority:         req.Priority,
		FormData:         string(formJSON),
		CreatorID:        userID,
		CreatorName:      username,
	}

	// 5. 落库 + 初始化节点记录 + 推进到首个节点（事务）
	var outcome flowOutcome
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
				Status:        modelticket.NodeStatusWaiting,
				ApproverIDs:   joinUintIDs(n.ApproverIDs),
				ApproverNames: strings.Join(n.ApproverNames, ","),
				MultiType:     n.MultiType,
			})
		}
		if err := tx.CreateNodeRecords(records); err != nil {
			return err
		}
		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: modelticket.FlowActionSubmit,
			OperatorID: userID, OperatorName: username,
			Comment: "发起工单",
		}); err != nil {
			return err
		}
		var err error
		outcome, err = s.advanceFlow(tx, ticket, snapshot, records)
		return err
	})
	if err != nil {
		return nil, err
	}
	// 事务提交后通知首个节点审批人（异步，不影响结果）
	if rec := outcome.activated; rec != nil {
		s.notify.NotifyApprovers(ticket, parseUintIDs(rec.ApproverIDs), nil, rec.NodeName)
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
			TimeoutHours: node.TimeoutHours,
			SortOrder:    i + 1,
		}
		// 条件解析失败必须报错：静默忽略会导致节点退化为无条件激活，改变审批路径
		conds, err := parseConditionStrict(node.Condition)
		if err != nil {
			return nil, fmt.Errorf("节点「%s」的激活条件配置格式错误: %v", node.Name, err)
		}
		snap.Condition = conds
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

// flowOutcome advanceFlow 的流转结果（事务提交后用于触发通知）
type flowOutcome struct {
	activated *modelticket.TicketNodeRecord // 本次激活的节点（nil 表示流程结束）
	finished  bool                          // 流程是否走完（工单通过）
}

// advanceFlow 从当前节点之后寻找下一个满足条件的节点并激活；
// 不满足条件的节点标记 skipped；全部走完则工单通过。
func (s *TicketService) advanceFlow(tx *repoticket.TicketRepository, ticket *modelticket.Ticket,
	snapshot []WorkflowSnapshotNode, records []modelticket.TicketNodeRecord) (flowOutcome, error) {

	outcome := flowOutcome{}
	var formData map[string]interface{}
	if err := json.Unmarshal([]byte(ticket.FormData), &formData); err != nil {
		// 表单数据损坏会导致条件求值失真（缺失字段按 nil 比较），必须中止推进而非继续流转
		return outcome, fmt.Errorf("工单表单数据解析失败: %v", err)
	}

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
					records[j].Status = modelticket.NodeStatusSkipped
					records[j].FinishedAt = &now
					if err := tx.UpdateNodeRecord(&records[j]); err != nil {
						return outcome, err
					}
				}
			}
			if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
				TicketID: ticket.ID, Action: modelticket.FlowActionSkip, NodeName: node.Name,
				OperatorID: 0, OperatorName: "系统",
				Comment: "条件不满足，自动跳过",
			}); err != nil {
				return outcome, err
			}
			continue
		}
		// 激活该节点
		nowPtr := now
		for j := range records {
			if records[j].NodeKey == node.Key {
				records[j].Status = modelticket.NodeStatusPending
				records[j].StartedAt = &nowPtr
				if err := tx.UpdateNodeRecord(&records[j]); err != nil {
					return outcome, err
				}
				outcome.activated = &records[j]
			}
		}
		if err := tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
			"status":            modelticket.TicketStatusPending,
			"current_node_key":  node.Key,
			"current_node_name": node.Name,
		}); err != nil {
			return outcome, err
		}
		return outcome, nil
	}

	// 没有更多节点：流程结束
	finished := time.Now()
	outcome.finished = true
	if err := tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
		"status":            modelticket.TicketStatusApproved,
		"current_node_key":  "",
		"current_node_name": "",
		"finished_at":       finished,
	}); err != nil {
		return outcome, err
	}
	return outcome, nil
}

// ────────────────────────── 审批操作 ──────────────────────────

// Approve 审批通过
func (s *TicketService) Approve(ticketID, userID uint, username, comment string) error {
	var (
		ticket  *modelticket.Ticket
		outcome flowOutcome
	)
	err := s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		var snapshot []WorkflowSnapshotNode
		var records []modelticket.TicketNodeRecord
		var err error
		ticket, snapshot, records, err = s.loadTicketForAction(tx, ticketID)
		if err != nil {
			return err
		}

		rec := findCurrentRecord(records, ticket.CurrentNodeKey)
		if rec == nil || rec.Status != modelticket.NodeStatusPending {
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
		completed := (rec.MultiType == modelticket.MultiTypeAll && len(approvedIDs) >= len(approverIDs)) ||
			rec.MultiType != modelticket.MultiTypeAll

		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: modelticket.FlowActionApprove, NodeName: rec.NodeName,
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
		rec.Status = modelticket.NodeStatusApproved
		rec.FinishedAt = &now
		if err := tx.UpdateNodeRecord(rec); err != nil {
			return err
		}
		outcome, err = s.advanceFlow(tx, ticket, snapshot, records)
		return err
	})
	if err != nil {
		return err
	}

	// 事务提交后触发通知：下一节点激活 → 通知审批人；流程结束 → 通知发起人
	if outcome.finished {
		s.notify.NotifyCreator(ticket, modelticket.TicketStatusApproved, username, comment)
	} else if rec := outcome.activated; rec != nil {
		s.notify.NotifyApprovers(ticket, parseUintIDs(rec.ApproverIDs), parseUintIDs(rec.ApprovedIDs), rec.NodeName)
	}
	return nil
}

// Reject 驳回（终止工单，发起人可修改后重新提交）
func (s *TicketService) Reject(ticketID, userID uint, username, comment string) error {
	var ticket *modelticket.Ticket
	err := s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		var records []modelticket.TicketNodeRecord
		var err error
		ticket, _, records, err = s.loadTicketForAction(tx, ticketID)
		if err != nil {
			return err
		}

		rec := findCurrentRecord(records, ticket.CurrentNodeKey)
		if rec == nil || rec.Status != modelticket.NodeStatusPending {
			return fmt.Errorf("工单当前节点状态异常")
		}
		if !containsUint(parseUintIDs(rec.ApproverIDs), userID) {
			return fmt.Errorf("您不是当前节点的审批人")
		}

		now := time.Now()
		rec.Status = modelticket.NodeStatusRejected
		rec.Comment = comment
		rec.FinishedAt = &now
		if err := tx.UpdateNodeRecord(rec); err != nil {
			return err
		}
		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: modelticket.FlowActionReject, NodeName: rec.NodeName,
			OperatorID: userID, OperatorName: username, Comment: comment,
		}); err != nil {
			return err
		}
		// 关闭尚未走到的 waiting 节点，避免记录与工单终态不一致
		if err := tx.CloseNodeRecords(ticket.ID, []string{modelticket.NodeStatusWaiting}, now); err != nil {
			return err
		}
		return tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
			"status":            modelticket.TicketStatusRejected,
			"current_node_key":  "",
			"current_node_name": "",
			"finished_at":       now,
		})
	})
	if err != nil {
		return err
	}
	// 事务提交后通知发起人被驳回（附驳回人与意见，便于修改重提）
	s.notify.NotifyCreator(ticket, modelticket.TicketStatusRejected, username, comment)
	return nil
}

// Cancel 发起人撤销工单
func (s *TicketService) Cancel(ticketID, userID uint, username, comment string) error {
	var (
		ticket *modelticket.Ticket
		curRec *modelticket.TicketNodeRecord
	)
	err := s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		var records []modelticket.TicketNodeRecord
		var err error
		ticket, _, records, err = s.loadTicketForAction(tx, ticketID)
		if err != nil {
			return err
		}
		if ticket.CreatorID != userID {
			return fmt.Errorf("只有发起人可以撤销工单")
		}
		now := time.Now()
		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: modelticket.FlowActionCancel,
			OperatorID: userID, OperatorName: username, Comment: comment,
		}); err != nil {
			return err
		}
		// 记录撤销前的当前待审节点（事务外通知其审批人）
		curRec = findCurrentRecord(records, ticket.CurrentNodeKey)
		// 同步关闭审批中与未到达的节点记录，避免残留 pending 与工单终态不一致
		if err := tx.CloseNodeRecords(ticket.ID, []string{modelticket.NodeStatusPending, modelticket.NodeStatusWaiting}, now); err != nil {
			return err
		}
		return tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
			"status":            modelticket.TicketStatusCanceled,
			"current_node_key":  "",
			"current_node_name": "",
			"finished_at":       now,
		})
	})
	if err != nil {
		return err
	}
	// 事务提交后通知原当前节点审批人：工单已被发起人撤销
	if curRec != nil && curRec.Status == modelticket.NodeStatusPending {
		s.notify.NotifyApproversCanceled(ticket, parseUintIDs(curRec.ApproverIDs), parseUintIDs(curRec.ApprovedIDs), curRec.NodeName, username)
	}
	return nil
}

// TicketResubmitRequest 重新提交请求（驳回后发起人修改重提）
type TicketResubmitRequest struct {
	Title    string                 `json:"title" binding:"required"`
	Priority string                 `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	FormData map[string]interface{} `json:"formData"`
}

// Resubmit 驳回后重新提交：发起人可修改标题/优先级/表单，流程从首个节点重新流转。
// 已有的审批记录全部重置（保留流转日志作为历史），条件节点按新表单数据重新求值。
func (s *TicketService) Resubmit(ticketID, userID uint, username string, req TicketResubmitRequest) error {
	var (
		ticket    *modelticket.Ticket
		activated *modelticket.TicketNodeRecord
	)
	err := s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		var err error
		ticket, err = tx.FindTicketByIDForUpdate(ticketID)
		if err != nil {
			return fmt.Errorf("工单不存在")
		}
		if ticket.Status != modelticket.TicketStatusRejected {
			return fmt.Errorf("仅被驳回的工单可以重新提交")
		}
		if ticket.CreatorID != userID {
			return fmt.Errorf("只有发起人可以重新提交工单")
		}

		// 按类型 schema 校验新表单（schema 损坏时报错阻断）
		tp, err := s.wfRepo.FindTypeByID(ticket.TypeID)
		if err != nil {
			return fmt.Errorf("工单类型已不存在，无法重新提交")
		}
		if err := s.validateFormData(tp.FormSchema, req.FormData); err != nil {
			return err
		}
		formJSON, err := json.Marshal(req.FormData)
		if err != nil {
			return fmt.Errorf("序列化表单数据失败: %v", err)
		}
		if req.Priority == "" {
			req.Priority = ticket.Priority
		}

		// 重置工单为待审批
		if err := tx.UpdateTicketFields(ticket.ID, map[string]interface{}{
			"status":            modelticket.TicketStatusPending,
			"title":             req.Title,
			"priority":          req.Priority,
			"form_data":         string(formJSON),
			"current_node_key":  "",
			"current_node_name": "",
			"finished_at":       nil, // 清空结束时间
		}); err != nil {
			return err
		}

		// 全部节点记录重置为 waiting，从快照重新流转
		records, err := tx.FindNodeRecords(ticketID)
		if err != nil {
			return err
		}
		if err := tx.ResetNodeRecords(ticketID); err != nil {
			return err
		}
		for i := range records {
			records[i].Status = modelticket.NodeStatusWaiting
			records[i].StartedAt = nil
			records[i].FinishedAt = nil
			records[i].ApprovedIDs = ""
			records[i].Comment = ""
		}

		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: modelticket.FlowActionResubmit,
			OperatorID: userID, OperatorName: username,
			Comment: "驳回后修改重新提交",
		}); err != nil {
			return err
		}

		var snapshot []WorkflowSnapshotNode
		if err := json.Unmarshal([]byte(ticket.WorkflowSnapshot), &snapshot); err != nil {
			return fmt.Errorf("流程快照解析失败")
		}
		ticket.FormData = string(formJSON)
		ticket.CurrentNodeKey = ""
		outcome, err := s.advanceFlow(tx, ticket, snapshot, records)
		activated = outcome.activated
		return err
	})
	if err != nil {
		return err
	}
	// 事务提交后通知重新流转到的首个节点审批人
	if activated != nil {
		s.notify.NotifyApprovers(ticket, parseUintIDs(activated.ApproverIDs), nil, activated.NodeName)
	}
	return nil
}

// Reassign 管理员改派当前节点审批人（解决审批人离职/请假导致节点卡死）
// 已审批记录（ApprovedIDs）保留：仍在新审批人列表中的已审人数计入会签进度
func (s *TicketService) Reassign(ticketID, operatorID uint, operatorName string, nodeKey string, approverIDs []uint) error {
	var (
		ticket  *modelticket.Ticket
		newRec  modelticket.TicketNodeRecord
		newName string
	)
	err := s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		var err error
		// 注意：不可用 := 声明，否则遮蔽外层 ticket，事务提交后通知时为 nil 导致 panic
		t, _, records, err := s.loadTicketForAction(tx, ticketID)
		if err != nil {
			return err
		}
		ticket = t
		if nodeKey == "" {
			nodeKey = ticket.CurrentNodeKey
		}
		rec := findCurrentRecord(records, nodeKey)
		if rec == nil {
			return fmt.Errorf("节点不存在")
		}
		if rec.Status != modelticket.NodeStatusPending {
			return fmt.Errorf("仅审批中的节点可以改派")
		}
		if nodeKey != ticket.CurrentNodeKey {
			return fmt.Errorf("仅当前审批节点可以改派")
		}
		if len(approverIDs) == 0 {
			return fmt.Errorf("请选择新的审批人")
		}
		// 校验新审批人存在且启用
		users, err := s.repo.FindUsersByIDs(approverIDs)
		if err != nil {
			return fmt.Errorf("查询审批人失败: %v", err)
		}
		if len(users) != len(uniqueUint(approverIDs)) {
			return fmt.Errorf("部分审批人不存在或已禁用")
		}
		names := make([]string, 0, len(users))
		for _, u := range users {
			names = append(names, displayName(u))
		}

		rec.ApproverIDs = joinUintIDs(uniqueUint(approverIDs))
		rec.ApproverNames = strings.Join(names, ",")
		if err := tx.UpdateNodeRecord(rec); err != nil {
			return err
		}
		newRec = *rec
		newName = strings.Join(names, ",")

		return tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: modelticket.FlowActionReassign, NodeName: rec.NodeName,
			OperatorID: operatorID, OperatorName: operatorName,
			Comment: "改派审批人：" + newName,
		})
	})
	if err != nil {
		return err
	}
	// 事务提交后通知新审批人
	s.notify.NotifyReassigned(ticket, parseUintIDs(newRec.ApproverIDs), newRec.NodeName, operatorName)
	return nil
}

// Urge 发起人催办：向当前节点待审批人发送提醒（带冷却频控）
func (s *TicketService) Urge(ticketID, userID uint, username string) error {
	var (
		ticket *modelticket.Ticket
		curRec *modelticket.TicketNodeRecord
	)
	err := s.repo.Transaction(func(tx *repoticket.TicketRepository) error {
		var err error
		ticket, err = tx.FindTicketByIDForUpdate(ticketID)
		if err != nil {
			return fmt.Errorf("工单不存在")
		}
		if ticket.CreatorID != userID {
			return fmt.Errorf("只有发起人可以催办")
		}
		if ticket.Status != modelticket.TicketStatusPending {
			return fmt.Errorf("工单已结束，无需催办")
		}
		// 冷却频控：行锁保证并发催办下读到的 LastUrgeAt 一定是最新值
		if ticket.LastUrgeAt != nil {
			if remain := modelticket.UrgeCooldown - time.Since(*ticket.LastUrgeAt); remain > 0 {
				return fmt.Errorf("催办太频繁，请 %d 分钟后再试", int(remain.Minutes())+1)
			}
		}
		records, err := tx.FindNodeRecords(ticketID)
		if err != nil {
			return err
		}
		curRec = findCurrentRecord(records, ticket.CurrentNodeKey)
		if curRec == nil || curRec.Status != modelticket.NodeStatusPending {
			return fmt.Errorf("工单当前节点状态异常")
		}
		now := time.Now()
		if err := tx.CreateFlowLog(modelticket.TicketFlowLog{
			TicketID: ticket.ID, Action: modelticket.FlowActionUrge, NodeName: curRec.NodeName,
			OperatorID: userID, OperatorName: username,
			Comment: "发起人催办提醒",
		}); err != nil {
			return err
		}
		return tx.UpdateTicketFields(ticket.ID, map[string]interface{}{"last_urge_at": now})
	})
	if err != nil {
		return err
	}
	// 事务提交后异步通知当前节点待审批人（排除已审过的人）
	s.notify.NotifyUrge(ticket, parseUintIDs(curRec.ApproverIDs), parseUintIDs(curRec.ApprovedIDs), curRec.NodeName, username)
	return nil
}

// Comment 添加评论
func (s *TicketService) Comment(ticketID, userID uint, username, comment string) error {
	if _, err := s.repo.FindTicketByID(ticketID); err != nil {
		return fmt.Errorf("工单不存在")
	}
	return s.repo.CreateFlowLog(modelticket.TicketFlowLog{
		TicketID: ticketID, Action: modelticket.FlowActionComment,
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
		if t.Status == modelticket.TicketStatusPending {
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
	Ticket      modelticket.Ticket             `json:"ticket"`
	FormSchema  []FormField                    `json:"formSchema"`
	FormData    map[string]interface{}         `json:"formData"`
	Nodes       []modelticket.TicketNodeRecord `json:"nodes"`
	Logs        []modelticket.TicketFlowLog    `json:"logs"`
	CanApprove  bool                           `json:"canApprove"`
	CanCancel   bool                           `json:"canCancel"`
	CanResubmit bool                           `json:"canResubmit"` // 驳回后发起人可修改重提
	CanUrge     bool                           `json:"canUrge"`     // 审批中且发起人可催办（冷却结束）
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

	detail.CanCancel = ticket.Status == modelticket.TicketStatusPending && ticket.CreatorID == userID
	detail.CanResubmit = ticket.Status == modelticket.TicketStatusRejected && ticket.CreatorID == userID
	urgeReady := ticket.LastUrgeAt == nil || time.Since(*ticket.LastUrgeAt) >= modelticket.UrgeCooldown
	detail.CanUrge = ticket.Status == modelticket.TicketStatusPending && ticket.CreatorID == userID && urgeReady
	for _, rec := range detail.Nodes {
		if rec.NodeKey == ticket.CurrentNodeKey && rec.Status == modelticket.NodeStatusPending {
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
// 使用 SELECT ... FOR UPDATE 锁定工单行：同一工单的并发审批在事务层面串行化，
// 后到的事务读到的一定是先到事务提交后的最新状态
func (s *TicketService) loadTicketForAction(tx *repoticket.TicketRepository, ticketID uint) (
	*modelticket.Ticket, []WorkflowSnapshotNode, []modelticket.TicketNodeRecord, error) {

	ticket, err := tx.FindTicketByIDForUpdate(ticketID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("工单不存在")
	}
	if ticket.Status != modelticket.TicketStatusPending {
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

// generateTicketNo 生成工单单号：T + yyyymmdd + 4位随机后缀
// 不使用 count+1（并发发起会因唯一索引冲突必现失败），随机后缀冲突概率极低，
// 唯一索引兜底 + 最多重试 5 次
func (s *TicketService) generateTicketNo() (string, error) {
	prefix := "T" + time.Now().Format("20060102")
	for i := 0; i < 5; i++ {
		no := fmt.Sprintf("%s%04d", prefix, rand.Intn(10000))
		exists, err := s.repo.TicketNoExists(no)
		if err != nil {
			return "", err
		}
		if !exists {
			return no, nil
		}
	}
	return "", fmt.Errorf("生成工单单号失败，请重试")
}

// validateFormData 按类型表单 schema 校验必填项
func (s *TicketService) validateFormData(schemaJSON string, data map[string]interface{}) error {
	fields, err := parseFormSchemaStrict(schemaJSON)
	if err != nil {
		return fmt.Errorf("工单类型的表单配置格式错误，请联系管理员修正: %v", err)
	}
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

// uniqueUint 去重并保持原有顺序
func uniqueUint(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// parseConditionStrict 严格解析节点激活条件（空串返回空切片）
// 写路径（发起工单快照固化）必须使用：格式错误时报错而非静默忽略，
// 否则节点会退化为无条件激活，静默改变审批路径
func parseConditionStrict(s string) ([]ConditionItem, error) {
	if s == "" {
		return nil, nil
	}
	var conds []ConditionItem
	if err := json.Unmarshal([]byte(s), &conds); err != nil {
		return nil, err
	}
	return conds, nil
}

// parseFormSchemaStrict 严格解析表单 schema（空串返回空切片）
// 写路径（发起工单校验）使用：schema 损坏时报错，避免必填校验被静默跳过
func parseFormSchemaStrict(s string) ([]FormField, error) {
	if s == "" {
		return nil, nil
	}
	var fields []FormField
	if err := json.Unmarshal([]byte(s), &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

// parseFormSchema 容错解析表单 schema（读路径使用：详情页展示，类型可能已损坏/删除）
func parseFormSchema(s string) []FormField {
	fields, _ := parseFormSchemaStrict(s)
	return fields
}

func displayName(u repoticket.UserBrief) string {
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Username
}
