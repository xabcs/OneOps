package k8s

import (
	"time"

	modelk8s "oneops/backend3/model/k8s"

	"gorm.io/gorm"
)

// DiagnosticRepository 诊断数据访问层
type DiagnosticRepository struct {
	db *gorm.DB
}

// NewDiagnosticRepository 创建诊断仓库
func NewDiagnosticRepository(db *gorm.DB) *DiagnosticRepository {
	return &DiagnosticRepository{db: db}
}

// DiagnosticHistoryQuery 诊断历史分页查询条件
type DiagnosticHistoryQuery struct {
	ClusterID string
	Namespace string
	PodName   string
	Page      int
	PageSize  int
}

// Create 创建诊断历史记录
func (r *DiagnosticRepository) Create(history *modelk8s.DiagnosticHistory) error {
	return r.db.Create(history).Error
}

// FindHistoryWithPagination 分页查询诊断历史
func (r *DiagnosticRepository) FindHistoryWithPagination(q DiagnosticHistoryQuery) ([]modelk8s.DiagnosticHistory, int64, error) {
	var histories []modelk8s.DiagnosticHistory
	var total int64

	query := r.db.Model(&modelk8s.DiagnosticHistory{})

	if q.ClusterID != "" {
		query = query.Where("cluster_id = ?", q.ClusterID)
	}
	if q.Namespace != "" {
		query = query.Where("namespace = ?", q.Namespace)
	}
	if q.PodName != "" {
		query = query.Where("pod_name = ?", q.PodName)
	}

	query.Count(&total)

	offset := (q.Page - 1) * q.PageSize
	if err := query.Offset(offset).Limit(q.PageSize).Order("timestamp DESC").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// ========== Agent 注册表 ==========

// UpsertAgent 按 agentId 插入或更新 agent 记录
func (r *DiagnosticRepository) UpsertAgent(agent *modelk8s.DiagnosticAgent) error {
	return r.db.Save(agent).Error
}

// FindAgentByAgentID 按 agentId 查询
func (r *DiagnosticRepository) FindAgentByAgentID(agentID string) (*modelk8s.DiagnosticAgent, error) {
	var agent modelk8s.DiagnosticAgent
	if err := r.db.Where("agent_id = ?", agentID).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

// ListAgents 查询 agent 列表，appName 为空返回全部
func (r *DiagnosticRepository) ListAgents(appName string, onlineOnly bool) ([]modelk8s.DiagnosticAgent, error) {
	var agents []modelk8s.DiagnosticAgent
	query := r.db.Model(&modelk8s.DiagnosticAgent{})
	if appName != "" {
		query = query.Where("app_name = ?", appName)
	}
	if onlineOnly {
		query = query.Where("online = ?", true)
	}
	err := query.Order("app_name ASC, pod_name ASC").Find(&agents).Error
	return agents, err
}

// MarkStaleAgentsOffline 将超时未上报的 agent 置为离线
func (r *DiagnosticRepository) MarkStaleAgentsOffline(before time.Time) error {
	return r.db.Model(&modelk8s.DiagnosticAgent{}).
		Where("online = ? AND last_seen < ?", true, before).
		Update("online", false).Error
}

// CountAgents 统计总数与在线数
func (r *DiagnosticRepository) CountAgents() (total int64, online int64, err error) {
	if err = r.db.Model(&modelk8s.DiagnosticAgent{}).Count(&total).Error; err != nil {
		return
	}
	err = r.db.Model(&modelk8s.DiagnosticAgent{}).Where("online = ?", true).Count(&online).Error
	return
}

// ========== 会话 ==========

// CreateSession 创建会话
func (r *DiagnosticRepository) CreateSession(s *modelk8s.DiagnosticSession) error {
	return r.db.Create(s).Error
}

// UpdateSession 更新会话
func (r *DiagnosticRepository) UpdateSession(s *modelk8s.DiagnosticSession) error {
	return r.db.Save(s).Error
}

// FindSessionByKey 按互斥键查询（agentId）
func (r *DiagnosticRepository) FindSessionByKey(sessionKey string) (*modelk8s.DiagnosticSession, error) {
	var s modelk8s.DiagnosticSession
	if err := r.db.Where("session_key = ? AND status = ?", sessionKey, "active").First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// FindSessionByID 按 ID 查询
func (r *DiagnosticRepository) FindSessionByID(id uint) (*modelk8s.DiagnosticSession, error) {
	var s modelk8s.DiagnosticSession
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ListActiveSessions 活跃会话列表
func (r *DiagnosticRepository) ListActiveSessions() ([]modelk8s.DiagnosticSession, error) {
	var sessions []modelk8s.DiagnosticSession
	err := r.db.Where("status = ?", "active").Order("start_at DESC").Find(&sessions).Error
	return sessions, err
}

// ListSessionsWithPagination 分页查询全部会话
func (r *DiagnosticRepository) ListSessionsWithPagination(page, pageSize int) ([]modelk8s.DiagnosticSession, int64, error) {
	var sessions []modelk8s.DiagnosticSession
	var total int64
	r.db.Model(&modelk8s.DiagnosticSession{}).Count(&total)
	offset := (page - 1) * pageSize
	err := r.db.Order("start_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error
	return sessions, total, err
}

// ========== 执行记录 ==========

// CreateExecution 创建执行记录
func (r *DiagnosticRepository) CreateExecution(e *modelk8s.DiagnosticExecution) error {
	return r.db.Create(e).Error
}

// ExecutionQuery 执行记录查询条件
type ExecutionQuery struct {
	ClusterID string
	AppName   string
	AgentID   string
	Username  string
	Command   string
	RiskLevel string
	Page      int
	PageSize  int
}

// FindExecutionsWithPagination 分页查询执行记录
func (r *DiagnosticRepository) FindExecutionsWithPagination(q ExecutionQuery) ([]modelk8s.DiagnosticExecution, int64, error) {
	var list []modelk8s.DiagnosticExecution
	var total int64

	query := r.db.Model(&modelk8s.DiagnosticExecution{})
	if q.ClusterID != "" {
		query = query.Where("cluster_id = ?", q.ClusterID)
	}
	if q.AppName != "" {
		query = query.Where("app_name = ?", q.AppName)
	}
	if q.AgentID != "" {
		query = query.Where("agent_id = ?", q.AgentID)
	}
	if q.Username != "" {
		query = query.Where("username = ?", q.Username)
	}
	if q.Command != "" {
		query = query.Where("command LIKE ?", "%"+q.Command+"%")
	}
	if q.RiskLevel != "" {
		query = query.Where("risk_level = ?", q.RiskLevel)
	}

	query.Count(&total)
	offset := (q.Page - 1) * q.PageSize
	err := query.Order("timestamp DESC").Offset(offset).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

// ExecutionStats 总览统计
type ExecutionStats struct {
	Total      int64 `json:"total"`
	HighRisk   int64 `json:"highRisk"`   // L4+L5
	ErrorCount int64 `json:"errorCount"` // 失败数
	ActiveSess int64 `json:"activeSess"` // 活跃会话
}

// GetExecutionStats 总览统计
func (r *DiagnosticRepository) GetExecutionStats() (*ExecutionStats, error) {
	stats := &ExecutionStats{}
	db := r.db.Model(&modelk8s.DiagnosticExecution{})
	if err := db.Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	if err := db.Where("risk_level IN ?", []string{"L4", "L5"}).Count(&stats.HighRisk).Error; err != nil {
		return nil, err
	}
	if err := db.Where("result_status = ?", "error").Count(&stats.ErrorCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&modelk8s.DiagnosticSession{}).Where("status = ?", "active").Count(&stats.ActiveSess).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// ========== 命令风险覆盖 ==========

// ListCommandOverrides 查询全部命令风险覆盖
func (r *DiagnosticRepository) ListCommandOverrides() ([]modelk8s.DiagnosticCommandOverride, error) {
	var list []modelk8s.DiagnosticCommandOverride
	err := r.db.Order("command ASC").Find(&list).Error
	return list, err
}

// UpsertCommandOverride 插入或更新命令风险覆盖
func (r *DiagnosticRepository) UpsertCommandOverride(o *modelk8s.DiagnosticCommandOverride) error {
	var existing modelk8s.DiagnosticCommandOverride
	err := r.db.Where("command = ?", o.Command).First(&existing).Error
	if err != nil {
		return r.db.Create(o).Error
	}
	existing.RiskLevel = o.RiskLevel
	existing.Description = o.Description
	existing.UpdatedBy = o.UpdatedBy
	return r.db.Save(&existing).Error
}

// DeleteCommandOverride 删除命令风险覆盖（恢复默认分级）
func (r *DiagnosticRepository) DeleteCommandOverride(id uint) error {
	return r.db.Delete(&modelk8s.DiagnosticCommandOverride{}, id).Error
}

// ========== 诊断配置 ==========

// GetConfigValue 读取诊断配置值（clusterID 可为 "default"）
func (r *DiagnosticRepository) GetConfigValue(clusterID, key, fallback string) string {
	var cfg modelk8s.DiagnosticConfig
	if err := r.db.Where("cluster_id = ? AND config_key = ?", clusterID, key).First(&cfg).Error; err == nil {
		return cfg.ConfigValue
	}
	return fallback
}

// SetConfigValue 写入诊断配置
func (r *DiagnosticRepository) SetConfigValue(clusterID, key, value, description string) error {
	var cfg modelk8s.DiagnosticConfig
	err := r.db.Where("cluster_id = ? AND config_key = ?", clusterID, key).First(&cfg).Error
	if err != nil {
		return r.db.Create(&modelk8s.DiagnosticConfig{
			ClusterID: clusterID, Namespace: "*", ConfigKey: key, ConfigValue: value, Description: description,
		}).Error
	}
	cfg.ConfigValue = value
	return r.db.Save(&cfg).Error
}

// DeleteExecutionsBefore 清理保留期外的执行记录（返回删除行数）
func (r *DiagnosticRepository) DeleteExecutionsBefore(before time.Time) (int64, error) {
	res := r.db.Where("timestamp < ?", before).Delete(&modelk8s.DiagnosticExecution{})
	return res.RowsAffected, res.Error
}
