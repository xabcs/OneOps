package cmdb

import (
	"fmt"
	"time"

	modelcmdb "oneops/backend3/model/cmdb"
	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// BastionRepository 堡垒机数据访问层
type BastionRepository struct {
	db *gorm.DB
}

// NewBastionRepository 创建堡垒机仓库
func NewBastionRepository(db *gorm.DB) *BastionRepository {
	return &BastionRepository{db: db}
}

// DB 返回底层 *gorm.DB
func (r *BastionRepository) DB() *gorm.DB {
	return r.db
}

// ========== 连接权限检查 ==========

// FindUserByID 根据ID获取用户
func (r *BastionRepository) FindUserByID(userID uint) (*modelsystem.User, error) {
	var user modelsystem.User
	// 只预加载启用角色：禁用角色的授权策略不应再对跳板机生效
	if err := r.db.Preload("Roles", "status = 1").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindServerWithUserCredentials 获取服务器及其用户凭证
func (r *BastionRepository) FindServerWithUserCredentials(serverID uint) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	if err := r.db.Preload("Credentials", "credential_type = ?", modelcmdb.CredentialTypeUser).First(&server, serverID).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// FindAccessPoliciesBySubject 获取命中的启用策略（角色策略 ∨ 用户策略）
func (r *BastionRepository) FindAccessPoliciesBySubject(userID uint, roleIDs []uint) ([]modelcmdb.AssetAccessPolicy, error) {
	var policies []modelcmdb.AssetAccessPolicy
	err := r.db.Where(
		"status = 1 AND ((subject_type = 'role' AND subject_id IN (?)) OR (subject_type = 'user' AND subject_id = ?))",
		roleIDs, userID,
	).Find(&policies).Error
	return policies, err
}

// FindAccessPoliciesWithHighRisk 获取含高危命令的启用策略（角色策略 ∨ 用户策略）
func (r *BastionRepository) FindAccessPoliciesWithHighRisk(userID uint, roleIDs []uint) ([]modelcmdb.AssetAccessPolicy, error) {
	var policies []modelcmdb.AssetAccessPolicy
	err := r.db.Where(
		"status = 1 AND ((subject_type = 'role' AND subject_id IN (?)) OR (subject_type = 'user' AND subject_id = ?)) "+
			"AND JSON_LENGTH(high_risk_commands) > 0",
		roleIDs, userID,
	).Find(&policies).Error
	return policies, err
}

// FindBusinessUnitAncestorIDs 获取业务系统自身及全部祖先 ID（含自身），用于子树匹配
func (r *BastionRepository) FindBusinessUnitAncestorIDs(businessID uint) ([]uint, error) {
	ids := make([]uint, 0, 4)
	current := businessID
	// 上限防御脏数据造成的父级循环
	for depth := 0; current != 0 && depth < 32; depth++ {
		ids = append(ids, current)
		var parentID uint
		if err := r.db.Model(&modelcmdb.BusinessUnit{}).
			Select("parent_id").
			Where("id = ?", current).
			Scan(&parentID).Error; err != nil {
			return nil, err
		}
		current = parentID
	}
	return ids, nil
}

// CountServerGroupRelation 统计服务器分组关联数量
func (r *BastionRepository) CountServerGroupRelation(serverID, groupID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelcmdb.ServerGroupRelation{}).
		Where("server_id = ? AND group_id = ?", serverID, groupID).
		Count(&count).Error
	return count, err
}

// CountServerTagRelation 统计服务器标签关联数量
func (r *BastionRepository) CountServerTagRelation(serverID, tagID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelcmdb.ServerTagRelation{}).
		Where("server_id = ? AND tag_id = ?", serverID, tagID).
		Count(&count).Error
	return count, err
}

// ========== 会话管理 ==========

// CreateSession 创建SSH会话
func (r *BastionRepository) CreateSession(session *modelcmdb.BastionSession) error {
	return r.db.Create(session).Error
}

// UpdateServerLastConnectTime 更新服务器最后连接时间
func (r *BastionRepository) UpdateServerLastConnectTime(serverID uint, t time.Time) error {
	return r.db.Model(&modelcmdb.Server{}).Where("id = ?", serverID).Update("last_connect_time", t).Error
}

// FindSessionByID 根据ID获取会话
func (r *BastionRepository) FindSessionByID(sessionID uint) (*modelcmdb.BastionSession, error) {
	var session modelcmdb.BastionSession
	if err := r.db.First(&session, sessionID).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// FindSessionDetailByID 获取会话详情（含关联）
func (r *BastionRepository) FindSessionDetailByID(sessionID uint) (*modelcmdb.BastionSession, error) {
	var session modelcmdb.BastionSession
	err := r.db.Preload("Server").
		Preload("Server.Cabinet").
		Preload("User").
		Preload("User.Roles").
		Preload("SSHCredential").
		First(&session, sessionID).Error
	return &session, err
}

// UpdateSession 更新会话
func (r *BastionRepository) UpdateSession(sessionID uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.BastionSession{}).Where("id = ?", sessionID).Updates(updates).Error
}

// CleanupOrphanedSessions 清理孤儿会话
func (r *BastionRepository) CleanupOrphanedSessions(now time.Time) (int64, error) {
	result := r.db.Model(&modelcmdb.BastionSession{}).
		Where("status = ?", "active").
		Updates(map[string]interface{}{
			"status":       "interrupted",
			"ended_at":     now,
			"close_reason": "服务器重启，会话已中断",
		})
	return result.RowsAffected, result.Error
}

// FindActiveSessions 获取活跃会话列表
func (r *BastionRepository) FindActiveSessions() ([]modelcmdb.BastionSession, error) {
	var sessions []modelcmdb.BastionSession
	err := r.db.Preload("Server").
		Preload("User").
		Preload("SSHCredential").
		Where("status = ?", "active").
		Order("started_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// FindSessions 获取会话列表（完整信息，分页）
func (r *BastionRepository) FindSessions(filter modelcmdb.SessionFilter, page, pageSize int) ([]modelcmdb.BastionSession, int64, error) {
	var sessions []modelcmdb.BastionSession
	var total int64

	tx := r.db.Model(&modelcmdb.BastionSession{})

	if filter.ServerID != nil {
		tx = tx.Where("server_id = ?", *filter.ServerID)
	}
	if filter.UserID != nil {
		tx = tx.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != nil {
		tx = tx.Where("status = ?", *filter.Status)
	}
	if filter.Protocol != nil {
		tx = tx.Where("protocol = ?", *filter.Protocol)
	}
	if filter.ClientIP != nil {
		tx = tx.Where("client_ip LIKE ?", "%"+*filter.ClientIP+"%")
	}
	if filter.LoginAccount != nil {
		tx = tx.Where("login_account = ?", *filter.LoginAccount)
	}
	if filter.StartDate != nil {
		tx = tx.Where("started_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		tx = tx.Where("started_at <= ?", *filter.EndDate)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Preload("Server").
		Preload("User").
		Preload("SSHCredential").
		Order("started_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&sessions).Error

	return sessions, total, err
}

// SessionListItem 会话列表项
type SessionListItem struct {
	ID           uint            `json:"id"`
	ServerID     uint            `json:"serverId"`
	Server       ServerBasicInfo `json:"server"`
	UserID       uint            `json:"userId"`
	User         UserBasicInfo   `json:"user"`
	LoginAccount string          `json:"loginAccount"`
	Protocol     string          `json:"protocol"`
	ClientIP     string          `json:"clientIp"`
	Status       string          `json:"status"`
	StartedAt    time.Time       `json:"startedAt"`
	EndedAt      *time.Time      `json:"endedAt"`
	Duration     int             `json:"duration"`
	CloseReason  string          `json:"closeReason"`
}

// ServerBasicInfo 服务器基本信息
type ServerBasicInfo struct {
	ID       uint   `json:"id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

// UserBasicInfo 用户基本信息
type UserBasicInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// FindSessionsList 获取会话列表（轻量级，分页）
func (r *BastionRepository) FindSessionsList(filter modelcmdb.SessionFilter, page, pageSize int) ([]SessionListItem, int64, error) {
	var sessions []SessionListItem
	var total int64

	tx := r.db.Table("cmdb_bastion_sessions").
		Select(`
			cmdb_bastion_sessions.id,
			cmdb_bastion_sessions.server_id,
			cmdb_bastion_sessions.user_id,
			cmdb_bastion_sessions.login_account,
			cmdb_bastion_sessions.protocol,
			cmdb_bastion_sessions.client_ip,
			cmdb_bastion_sessions.status,
			cmdb_bastion_sessions.started_at,
			cmdb_bastion_sessions.ended_at,
			cmdb_bastion_sessions.duration,
			cmdb_bastion_sessions.close_reason,
			cmdb_servers.id as server_id,
			cmdb_servers.hostname,
			cmdb_servers.ip,
			sys_users.id as user_id,
			sys_users.username
		`).
		Joins("LEFT JOIN cmdb_servers ON cmdb_bastion_sessions.server_id = cmdb_servers.id").
		Joins("LEFT JOIN sys_users ON cmdb_bastion_sessions.user_id = sys_users.id")

	if filter.ServerID != nil {
		tx = tx.Where("cmdb_bastion_sessions.server_id = ?", *filter.ServerID)
	}
	if filter.UserID != nil {
		tx = tx.Where("cmdb_bastion_sessions.user_id = ?", *filter.UserID)
	}
	if filter.Status != nil {
		tx = tx.Where("cmdb_bastion_sessions.status = ?", *filter.Status)
	}
	if filter.Protocol != nil {
		tx = tx.Where("cmdb_bastion_sessions.protocol = ?", *filter.Protocol)
	}
	if filter.ClientIP != nil {
		tx = tx.Where("cmdb_bastion_sessions.client_ip LIKE ?", "%"+*filter.ClientIP+"%")
	}
	if filter.LoginAccount != nil {
		tx = tx.Where("cmdb_bastion_sessions.login_account = ?", *filter.LoginAccount)
	}
	if filter.StartDate != nil {
		tx = tx.Where("cmdb_bastion_sessions.started_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		tx = tx.Where("cmdb_bastion_sessions.started_at <= ?", *filter.EndDate)
	}

	countTx := r.db.Model(&modelcmdb.BastionSession{})
	if filter.ServerID != nil {
		countTx = countTx.Where("server_id = ?", *filter.ServerID)
	}
	if filter.UserID != nil {
		countTx = countTx.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != nil {
		countTx = countTx.Where("status = ?", *filter.Status)
	}
	if filter.Protocol != nil {
		countTx = countTx.Where("protocol = ?", *filter.Protocol)
	}
	if filter.ClientIP != nil {
		countTx = countTx.Where("client_ip LIKE ?", "%"+*filter.ClientIP+"%")
	}
	if filter.LoginAccount != nil {
		countTx = countTx.Where("login_account = ?", *filter.LoginAccount)
	}
	if filter.StartDate != nil {
		countTx = countTx.Where("started_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		countTx = countTx.Where("started_at <= ?", *filter.EndDate)
	}

	if err := countTx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.
		Order("cmdb_bastion_sessions.started_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&sessions).Error

	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	for i := range sessions {
		if sessions[i].Status == "active" && sessions[i].StartedAt.After(time.Time{}) {
			sessions[i].Duration = int(now.Sub(sessions[i].StartedAt).Seconds())
		} else if sessions[i].Duration == 0 && sessions[i].EndedAt != nil {
			sessions[i].Duration = int(sessions[i].EndedAt.Sub(sessions[i].StartedAt).Seconds())
		}
	}

	return sessions, total, nil
}

// ========== 命令审计 ==========

// CreateCommand 创建命令记录
func (r *BastionRepository) CreateCommand(cmd *modelcmdb.BastionCommand) error {
	return r.db.Create(cmd).Error
}

// FindSessionWithServer 获取会话及其服务器（用于命令拦截检查）
func (r *BastionRepository) FindSessionWithServer(sessionID uint) (*modelcmdb.BastionSession, error) {
	var session modelcmdb.BastionSession
	if err := r.db.Preload("Server").Preload("User.Roles").First(&session, sessionID).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// FindSessionCommands 获取会话的命令列表
func (r *BastionRepository) FindSessionCommands(sessionID uint) ([]modelcmdb.BastionCommand, error) {
	var commands []modelcmdb.BastionCommand
	err := r.db.Where("session_id = ?", sessionID).
		Order("executed_at ASC").
		Find(&commands).Error
	return commands, err
}

// FindCommands 获取命令列表（分页）
func (r *BastionRepository) FindCommands(filter modelcmdb.CommandFilter, page, pageSize int) ([]modelcmdb.BastionCommand, int64, error) {
	var commands []modelcmdb.BastionCommand
	var total int64

	tx := r.db.Model(&modelcmdb.BastionCommand{})

	if filter.SessionID != nil {
		tx = tx.Where("session_id = ?", *filter.SessionID)
	}
	if filter.RiskLevel != nil {
		tx = tx.Where("risk_level = ?", *filter.RiskLevel)
	}
	if filter.Blocked != nil {
		tx = tx.Where("blocked = ?", *filter.Blocked)
	}
	if filter.CommandLike != nil {
		tx = tx.Where("command LIKE ?", "%"+*filter.CommandLike+"%")
	}
	if filter.StartDate != nil {
		tx = tx.Where("executed_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		tx = tx.Where("executed_at <= ?", *filter.EndDate)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Preload("Session").
		Preload("Session.Server").
		Preload("Session.User").
		Order("executed_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&commands).Error

	return commands, total, err
}

// ========== 文件传输审计 ==========

// CreateFileTransfer 创建文件传输记录
func (r *BastionRepository) CreateFileTransfer(transfer *modelcmdb.BastionFileTransfer) error {
	return r.db.Create(transfer).Error
}

// UpdateFileTransferStatus 更新文件传输状态
func (r *BastionRepository) UpdateFileTransferStatus(transferID uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.BastionFileTransfer{}).
		Where("id = ?", transferID).
		Updates(updates).Error
}

// FindSessionFileTransfers 获取会话的文件传输记录
func (r *BastionRepository) FindSessionFileTransfers(sessionID uint) ([]modelcmdb.BastionFileTransfer, error) {
	var transfers []modelcmdb.BastionFileTransfer
	err := r.db.Where("session_id = ?", sessionID).
		Order("started_at DESC").
		Find(&transfers).Error
	return transfers, err
}

// FindFileTransfers 获取文件传输列表（分页）
func (r *BastionRepository) FindFileTransfers(filter modelcmdb.FileTransferFilter, page, pageSize int) ([]modelcmdb.BastionFileTransfer, int64, error) {
	var transfers []modelcmdb.BastionFileTransfer
	var total int64

	tx := r.db.Model(&modelcmdb.BastionFileTransfer{})

	if filter.SessionID != nil {
		tx = tx.Where("session_id = ?", *filter.SessionID)
	}
	if filter.Direction != nil {
		tx = tx.Where("direction = ?", *filter.Direction)
	}
	if filter.Status != nil {
		tx = tx.Where("status = ?", *filter.Status)
	}
	if filter.StartDate != nil {
		tx = tx.Where("started_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		tx = tx.Where("started_at <= ?", *filter.EndDate)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Preload("Session").
		Preload("Session.Server").
		Preload("Session.User").
		Order("started_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&transfers).Error

	return transfers, total, err
}

// ========== 访问策略管理 ==========

// FindAccessPolicies 获取访问策略列表（分页）
func (r *BastionRepository) FindAccessPolicies(page, pageSize int) ([]modelcmdb.AssetAccessPolicy, int64, error) {
	var policies []modelcmdb.AssetAccessPolicy
	var total int64

	tx := r.db.Model(&modelcmdb.AssetAccessPolicy{})

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&policies).Error

	return policies, total, err
}

// CreateAccessPolicy 创建访问策略
func (r *BastionRepository) CreateAccessPolicy(policy *modelcmdb.AssetAccessPolicy) error {
	return r.db.Create(policy).Error
}

// UpdateAccessPolicy 更新访问策略
func (r *BastionRepository) UpdateAccessPolicy(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.AssetAccessPolicy{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteAccessPolicy 删除访问策略
func (r *BastionRepository) DeleteAccessPolicy(id uint) error {
	return r.db.Delete(&modelcmdb.AssetAccessPolicy{}, id).Error
}

// FindAccessPolicyByID 根据ID获取访问策略
func (r *BastionRepository) FindAccessPolicyByID(id uint) (*modelcmdb.AssetAccessPolicy, error) {
	var policy modelcmdb.AssetAccessPolicy
	err := r.db.First(&policy, id).Error
	if err != nil {
		return nil, fmt.Errorf("策略不存在: %w", err)
	}
	return &policy, nil
}

// FindAccessPolicyIDByName 根据策略名称查找策略 ID（用于名称唯一性校验）
// excludeID 用于更新场景排除自身，传 0 表示不过滤；未找到时返回 0
func (r *BastionRepository) FindAccessPolicyIDByName(name string, excludeID uint) (uint, error) {
	var id uint
	tx := r.db.Model(&modelcmdb.AssetAccessPolicy{}).
		Select("id").
		Where("name = ?", name)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	if err := tx.Scan(&id).Error; err != nil {
		return 0, err
	}
	return id, nil
}

// ExistsRecord 检查指定模型的表中是否存在给定 ID 的记录（用于外键存在性校验）
func (r *BastionRepository) ExistsRecord(model interface{}, id uint) (bool, error) {
	var count int64
	if err := r.db.Model(model).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
