package system

import (
	modelsystem "oneops/backend3/model/system"
	modelticket "oneops/backend3/model/ticket"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// syncTicketSeeds 同步工单示例数据（工单场景 + 场景下的审批流程）
// 幂等：按 code 判定，已存在（含页面修改过）则跳过，不覆盖用户数据；
// 已存在但未归属场景（type_id=0）的种子流程会回填归属。
func (i *Initializer) syncTicketSeeds() error {
	logger.Info("开始同步工单示例数据...")

	db := database.GetDB()

	// 解析内置 admin 角色 ID 作为示例审批人（按角色审批，角色下用户变化自动生效）
	var adminRole modelsystem.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		logger.Warn("未找到 admin 角色，跳过工单示例数据", zap.Error(err))
		return nil
	}
	roleJSON := uintIDsToCSV([]uint{adminRole.ID})

	// ── 1. 工单场景（先建场景，流程归属场景） ──
	typeSeeds := []modelticket.TicketType{
		{
			Name: "通用申请", Code: "general", Icon: "mdi:clipboard-text-outline",
			Description: "通用事务申请工单", Status: 1,
			FormSchema: `[{"key":"reason","label":"申请事由","type":"textarea","required":true,"placeholder":"请描述申请事由"}]`,
		},
		{
			Name: "SQL 审核", Code: "sql", Icon: "mdi:database-search",
			Description: "SQL 变更审核（高风险变更自动追加主管复核）", Status: 1,
			FormSchema: `[` +
				`{"key":"db_name","label":"目标数据库","type":"input","required":true,"placeholder":"如：order_db"},` +
				`{"key":"risk_level","label":"风险等级","type":"select","required":true,"options":["low","medium","high"]},` +
				`{"key":"sql_content","label":"SQL 内容","type":"textarea","required":true,"placeholder":"待执行的 SQL 语句"},` +
				`{"key":"rollback_plan","label":"回滚方案","type":"textarea","required":false,"placeholder":"出问题时的回滚步骤"}` +
				`]`,
		},
		{
			Name: "应用发布", Code: "release", Icon: "mdi:rocket-launch-outline",
			Description: "应用版本发布审批（生产环境自动追加复核）", Status: 1,
			FormSchema: `[` +
				`{"key":"app_name","label":"应用名称","type":"input","required":true},` +
				`{"key":"env","label":"发布环境","type":"select","required":true,"options":["dev","test","prod"]},` +
				`{"key":"version","label":"版本号","type":"input","required":true},` +
				`{"key":"release_note","label":"发布说明","type":"textarea","required":false}` +
				`]`,
		},
	}
	typeID := func(code string) uint {
		var t modelticket.TicketType
		if err := db.Where("code = ?", code).First(&t).Error; err != nil {
			return 0
		}
		return t.ID
	}
	for _, t := range typeSeeds {
		var existing modelticket.TicketType
		if err := db.Where("code = ?", t.Code).First(&existing).Error; err != nil {
			if err := db.Create(&t).Error; err != nil {
				logger.Warn("创建工单场景失败", zap.String("code", t.Code), zap.Error(err))
			}
		}
	}

	// ── 2. 场景下的审批流程（同一场景可配多个，此处每场景一个示例） ──
	workflowSeeds := []modelticket.Workflow{
		{
			TypeID:      typeID("general"),
			Name:        "通用审批流程",
			Code:        "general-approval",
			Description: "两级审批示例：一级审批全员必经；紧急工单追加二级审批（条件节点示例）",
			Status:      1,
			Nodes: []modelticket.WorkflowNode{
				{Name: "一级审批", ApproverType: "role", ApproverIDs: roleJSON, MultiType: "any", SortOrder: 1},
				{Name: "二级审批（紧急工单）", ApproverType: "role", ApproverIDs: roleJSON, MultiType: "all", SortOrder: 2,
					Condition: `[{"field":"priority","op":"eq","value":"urgent"}]`},
			},
		},
		{
			TypeID:      typeID("general"),
			Name:        "通用快速通道",
			Code:        "general-fast",
			Description: "单节点审批示例：适用于低风险通用申请（发起时与通用审批流程二选一）",
			Status:      1,
			Nodes: []modelticket.WorkflowNode{
				{Name: "快速审批", ApproverType: "role", ApproverIDs: roleJSON, MultiType: "any", SortOrder: 1},
			},
		},
		{
			TypeID:      typeID("sql"),
			Name:        "SQL 审核流程",
			Code:        "sql-audit",
			Description: "DBA 审核所有工单；高风险变更追加主管会签复核",
			Status:      1,
			Nodes: []modelticket.WorkflowNode{
				{Name: "DBA 审核", ApproverType: "role", ApproverIDs: roleJSON, MultiType: "any", SortOrder: 1},
				{Name: "主管复核（高风险）", ApproverType: "role", ApproverIDs: roleJSON, MultiType: "all", SortOrder: 2,
					Condition: `[{"field":"risk_level","op":"eq","value":"high"}]`},
			},
		},
		{
			TypeID:      typeID("release"),
			Name:        "应用发布流程",
			Code:        "app-release",
			Description: "普通环境一级审批；生产环境追加二级会签",
			Status:      1,
			Nodes: []modelticket.WorkflowNode{
				{Name: "发布审批", ApproverType: "role", ApproverIDs: roleJSON, MultiType: "any", SortOrder: 1},
				{Name: "生产发布复核", ApproverType: "role", ApproverIDs: roleJSON, MultiType: "all", SortOrder: 2,
					Condition: `[{"field":"env","op":"eq","value":"prod"}]`},
			},
		},
	}
	for _, wf := range workflowSeeds {
		var existing modelticket.Workflow
		if err := db.Where("code = ?", wf.Code).First(&existing).Error; err != nil {
			if wf.TypeID == 0 {
				logger.Warn("场景未就绪，跳过流程种子", zap.String("code", wf.Code))
				continue
			}
			if err := db.Create(&wf).Error; err != nil {
				logger.Warn("创建审批流程失败", zap.String("code", wf.Code), zap.Error(err))
			}
		} else if existing.TypeID == 0 && wf.TypeID > 0 {
			// 旧版本种子流程未归属场景，回填
			if err := db.Model(&existing).Update("type_id", wf.TypeID).Error; err != nil {
				logger.Warn("回填流程场景归属失败", zap.String("code", wf.Code), zap.Error(err))
			}
		}
	}

	logger.Info("工单示例数据同步完成")
	return nil
}

// uintIDsToCSV uint 列表转逗号分隔字符串
func uintIDsToCSV(ids []uint) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ","
		}
		out += itoa(id)
	}
	return out
}

func itoa(v uint) string {
	if v == 0 {
		return "0"
	}
	digits := ""
	for v > 0 {
		digits = string(rune('0'+v%10)) + digits
		v /= 10
	}
	return digits
}
