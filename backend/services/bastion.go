package services

import (
	"encoding/json"
	"fmt"
	"oneops/backend/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BastionService 堡垒机服务
type BastionService struct {
	db *gorm.DB
}

// NewBastionService 创建堡垒机服务实例
func NewBastionService() *BastionService {
	return &BastionService{
		db: db,
	}
}

// ========== 连接权限检查 ==========

// CheckConnectPermission 检查用户是否有连接指定服务器的权限
func (s *BastionService) CheckConnectPermission(userID uint, serverID uint) (bool, []string, error) {
	// 1. 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false, nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 2. 获取用户的角色ID列表
	roleIDs, err := parseRoleIDs(user.RoleIDs)
	if err != nil {
		return false, nil, fmt.Errorf("解析用户角色失败: %w", err)
	}

	// 3. 检查是否是超级管理员（角色ID为1）
	for _, roleID := range roleIDs {
		if roleID == 1 { // 超级管理员角色ID通常是1
			// 超级管理员返回所有可用账号
			return true, s.getAvailableAccounts(serverID), nil
		}
	}

	// 4. 获取服务器信息
	var server models.Server
	if err := s.db.Preload("SSHCredential").First(&server, serverID).Error; err != nil {
		return false, nil, fmt.Errorf("服务器不存在: %w", err)
	}

	// 5. 检查服务器是否绑定了SSH凭证
	if server.SSHCredentialID == 0 {
		return false, nil, fmt.Errorf("服务器未绑定SSH凭证")
	}

	// 6. 查询用户的访问策略
	var policies []models.AssetAccessPolicy
	err = s.db.Where("status = 1 AND subject_type = ? AND subject_id IN (?)",
		"role",
		roleIDs,
	).Find(&policies).Error

	if err != nil {
		return false, nil, err
	}

	// 7. 检查是否有匹配的策略
	allowedAccounts := make([]string, 0)
	hasPermission := false

	for _, policy := range policies {
		if s.matchesPolicy(policy, serverID, &server) {
			hasPermission = true
			// 合并允许的账号
			allowedAccounts = append(allowedAccounts, policy.LoginAccounts...)
		}
	}

	if !hasPermission {
		return false, nil, fmt.Errorf("没有访问权限")
	}

	// 去重
	allowedAccounts = unique(allowedAccounts)

	// 8. 检查是否需要审批
	for _, policy := range policies {
		if s.matchesPolicy(policy, serverID, &server) && policy.RequireApproval {
			// TODO: 检查是否有有效的审批记录
			return true, allowedAccounts, nil
		}
	}

	return true, allowedAccounts, nil
}

// matchesPolicy 检查策略是否匹配服务器
func (s *BastionService) matchesPolicy(policy models.AssetAccessPolicy, serverID uint, server *models.Server) bool {
	switch policy.AssetScopeType {
	case "all":
		return true
	case "server":
		return policy.AssetScopeID == serverID
	case "group":
		// 检查服务器是否属于指定分组
		var count int64
		s.db.Model(&models.ServerGroupRelation{}).
			Where("server_id = ? AND group_id = ?", serverID, policy.AssetScopeID).
			Count(&count)
		return count > 0
	case "tag":
		// 检查服务器是否有指定标签
		var count int64
		s.db.Model(&models.ServerTagRelation{}).
			Where("server_id = ? AND tag_id = ?", serverID, policy.AssetScopeID).
			Count(&count)
		return count > 0
	case "business":
		// TODO: 实现业务系统匹配
		return false
	default:
		return false
	}
}

// getAvailableAccounts 获取服务器可用的登录账号
func (s *BastionService) getAvailableAccounts(serverID uint) []string {
	// 返回常用的系统账号
	return []string{"root", "admin"}
}

// ========== 会话管理 ==========

// CreateSSHSession 创建SSH会话
func (s *BastionService) CreateSSHSession(userID uint, serverID uint, loginAccount string, clientIP string, protocol string) (*models.BastionSession, error) {
	// 1. 检查权限
	hasPermission, allowedAccounts, err := s.CheckConnectPermission(userID, serverID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, fmt.Errorf("没有连接权限")
	}

	// 2. 检查登录账号是否在允许列表中
	if len(allowedAccounts) > 0 {
		accountAllowed := false
		for _, acc := range allowedAccounts {
			if acc == loginAccount {
				accountAllowed = true
				break
			}
		}
		if !accountAllowed {
			return nil, fmt.Errorf("不允许使用账号 %s 连接", loginAccount)
		}
	}

	// 3. 获取用户信息
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 4. 创建会话记录
	now := time.Now()
	session := &models.BastionSession{
		ServerID:     serverID,
		UserID:       userID,
		Username:     user.Username,
		LoginAccount: loginAccount,
		ClientIP:     clientIP,
		Protocol:     protocol,
		StartedAt:    &now,
		Status:       "active",
	}

	// 5. 获取服务器的SSH凭证ID
	var server models.Server
	if err := s.db.First(&server, serverID).Error; err != nil {
		return nil, fmt.Errorf("服务器不存在: %w", err)
	}
	session.SSHCredentialID = server.SSHCredentialID

	// 6. 保存会话
	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	// 7. 更新服务器最后连接时间
	s.db.Model(&models.Server{}).Where("id = ?", serverID).Update("last_connect_time", now)

	return session, nil
}

// CloseSession 关闭会话
func (s *BastionService) CloseSession(sessionID uint, reason string) error {
	var session models.BastionSession
	if err := s.db.First(&session, sessionID).Error; err != nil {
		return fmt.Errorf("会话不存在: %w", err)
	}

	if session.Status != "active" {
		return fmt.Errorf("会话已关闭")
	}

	now := time.Now()
	duration := int(now.Sub(*session.StartedAt).Seconds())

	updates := map[string]interface{}{
		"status":       "closed",
		"ended_at":     now,
		"duration":     duration,
		"close_reason": reason,
	}

	return s.db.Model(&models.BastionSession{}).Where("id = ?", sessionID).Updates(updates).Error
}

// GetActiveSessions 获取活跃会话列表
func (s *BastionService) GetActiveSessions() ([]models.BastionSession, error) {
	var sessions []models.BastionSession
	err := s.db.Preload("Server").
		Preload("User").
		Preload("SSHCredential").
		Where("status = ?", "active").
		Order("started_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// TerminateSession 强制断开会话
func (s *BastionService) TerminateSession(sessionID uint, operatorID uint) error {
	return s.CloseSession(sessionID, fmt.Sprintf("被用户 %d 强制断开", operatorID))
}

// GetSessions 获取会话列表（分页）
func (s *BastionService) GetSessions(filter models.SessionFilter, page int, pageSize int) ([]models.BastionSession, int64, error) {
	var sessions []models.BastionSession
	var total int64

	tx := s.db.Model(&models.BastionSession{})

	// 应用筛选条件
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

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := tx.Preload("Server").
		Preload("User").
		Preload("SSHCredential").
		Order("started_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&sessions).Error

	return sessions, total, err
}

// GetSessionByID 获取会话详情
func (s *BastionService) GetSessionByID(sessionID uint) (*models.BastionSession, error) {
	var session models.BastionSession
	err := s.db.Preload("Server").
		Preload("Server.Cabinet").
		Preload("User").
		Preload("SSHCredential").
		First(&session, sessionID).Error
	return &session, err
}

// ========== 命令审计 ==========

// RecordCommand 记录命令
func (s *BastionService) RecordCommand(sessionID uint, command string, exitCode int, output string) error {
	// 分析命令风险等级
	riskLevel := s.analyzeCommandRisk(command)

	// 检查是否需要拦截命令
	blocked := s.isCommandBlocked(command, sessionID)

	cmd := &models.BastionCommand{
		SessionID:    sessionID,
		Command:      command,
		ExitCode:     &exitCode,
		RiskLevel:    riskLevel,
		Blocked:      blocked,
		OutputSummary: truncateString(output, 1000),
	}

	now := time.Now()
	cmd.ExecutedAt = &now

	return s.db.Create(cmd).Error
}

// analyzeCommandRisk 分析命令风险等级
func (s *BastionService) analyzeCommandRisk(command string) string {
	command = strings.ToLower(strings.TrimSpace(command))

	// 高危命令
	highRiskCommands := []string{
		"rm -rf /", "rm -rf /*", "mkfs", "dd if=/dev/zero",
		"shutdown", "reboot", "halt", "poweroff",
		":(){:|:&};:", // fork bomb
		"chmod 000", "chattr",
	}

	for _, risky := range highRiskCommands {
		if strings.Contains(command, risky) {
			return "critical"
		}
	}

	// 中危命令
	mediumRiskCommands := []string{
		"rm ", "mv ", "cp ",
		"chmod", "chown",
		"iptables", "firewall",
		"userdel", "groupdel",
		"killall", "pkill",
	}

	for _, risky := range mediumRiskCommands {
		if strings.HasPrefix(command, risky) {
			return "medium"
		}
	}

	// 低危命令
	lowRiskPatterns := []string{"yum", "apt", "systemctl", "service"}

	for _, pattern := range lowRiskPatterns {
		if strings.HasPrefix(command, pattern) {
			return "low"
		}
	}

	return "safe"
}

// isCommandBlocked 检查命令是否被拦截
func (s *BastionService) isCommandBlocked(command string, sessionID uint) bool {
	// 1. 获取会话信息
	var session models.BastionSession
	if err := s.db.Preload("Server").First(&session, sessionID).Error; err != nil {
		return false
	}

	// 2. 获取服务器的访问策略
	var policies []models.AssetAccessPolicy
	err := s.db.Where("status = 1 AND subject_type = ? AND subject_id IN (?) AND high_risk_commands IS NOT NULL",
		"role",
		getRoleIDsFromSession(&session),
	).Find(&policies).Error

	if err != nil || len(policies) == 0 {
		return false
	}

	// 3. 检查命令是否在黑名单中
	commandLower := strings.ToLower(command)
	for _, policy := range policies {
		for _, blockedCmd := range policy.HighRiskCommands {
			if strings.Contains(commandLower, strings.ToLower(blockedCmd)) {
				return true
			}
		}
	}

	return false
}

// GetSessionCommands 获取会话的命令列表
func (s *BastionService) GetSessionCommands(sessionID uint) ([]models.BastionCommand, error) {
	var commands []models.BastionCommand
	err := s.db.Where("session_id = ?", sessionID).
		Order("executed_at ASC").
		Find(&commands).Error
	return commands, err
}

// GetCommands 获取命令列表（分页）
func (s *BastionService) GetCommands(filter models.CommandFilter, page int, pageSize int) ([]models.BastionCommand, int64, error) {
	var commands []models.BastionCommand
	var total int64

	tx := s.db.Model(&models.BastionCommand{})

	// 应用筛选条件
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

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
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

// RecordFileTransfer 记录文件传输
func (s *BastionService) RecordFileTransfer(sessionID uint, direction string, remotePath string, localPath string, fileSize int64) (*models.BastionFileTransfer, error) {
	transfer := &models.BastionFileTransfer{
		SessionID:  sessionID,
		Direction:  direction,
		RemotePath: remotePath,
		LocalPath:  localPath,
		FileSize:   fileSize,
		Status:     "pending",
	}

	now := time.Now()
	transfer.StartedAt = &now

	if err := s.db.Create(transfer).Error; err != nil {
		return nil, err
	}

	return transfer, nil
}

// UpdateFileTransferStatus 更新文件传输状态
func (s *BastionService) UpdateFileTransferStatus(transferID uint, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "success" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = now
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return s.db.Model(&models.BastionFileTransfer{}).
		Where("id = ?", transferID).
		Updates(updates).Error
}

// GetSessionFileTransfers 获取会话的文件传输记录
func (s *BastionService) GetSessionFileTransfers(sessionID uint) ([]models.BastionFileTransfer, error) {
	var transfers []models.BastionFileTransfer
	err := s.db.Where("session_id = ?", sessionID).
		Order("started_at DESC").
		Find(&transfers).Error
	return transfers, err
}

// GetFileTransfers 获取文件传输列表（分页）
func (s *BastionService) GetFileTransfers(filter models.FileTransferFilter, page int, pageSize int) ([]models.BastionFileTransfer, int64, error) {
	var transfers []models.BastionFileTransfer
	var total int64

	tx := s.db.Model(&models.BastionFileTransfer{})

	// 应用筛选条件
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

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
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

// GetAccessPolicies 获取访问策略列表（分页）
func (s *BastionService) GetAccessPolicies(page int, pageSize int) ([]models.AssetAccessPolicy, int64, error) {
	var policies []models.AssetAccessPolicy
	var total int64

	tx := s.db.Model(&models.AssetAccessPolicy{})

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err := tx.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&policies).Error

	return policies, total, err
}

// CreateAccessPolicy 创建访问策略
func (s *BastionService) CreateAccessPolicy(policy *models.AssetAccessPolicy) error {
	return s.db.Create(policy).Error
}

// UpdateAccessPolicy 更新访问策略
func (s *BastionService) UpdateAccessPolicy(id uint, updates map[string]interface{}) error {
	return s.db.Model(&models.AssetAccessPolicy{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteAccessPolicy 删除访问策略
func (s *BastionService) DeleteAccessPolicy(id uint) error {
	return s.db.Delete(&models.AssetAccessPolicy{}, id).Error
}

// GetAccessPolicyByID 根据ID获取访问策略
func (s *BastionService) GetAccessPolicyByID(id uint) (*models.AssetAccessPolicy, error) {
	var policy models.AssetAccessPolicy
	err := s.db.First(&policy, id).Error
	return &policy, err
}

// ========== 辅助函数 ==========

// parseRoleIDs 解析用户的角色ID列表
func parseRoleIDs(roleIDsStr string) ([]uint, error) {
	if roleIDsStr == "" || roleIDsStr == "[]" {
		return []uint{}, nil
	}

	var roleIDs []uint
	err := json.Unmarshal([]byte(roleIDsStr), &roleIDs)
	if err != nil {
		return nil, err
	}
	return roleIDs, nil
}

// getRoleIDsFromSession 从会话中获取角色ID
func getRoleIDsFromSession(session *models.BastionSession) []uint {
	if session.User != nil {
		roleIDs, err := parseRoleIDs(session.User.RoleIDs)
		if err == nil {
			return roleIDs
		}
	}
	return []uint{}
}

// unique 去重字符串切片
func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
