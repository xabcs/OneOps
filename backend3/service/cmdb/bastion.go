package cmdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/database"
	repocmdb "oneops/backend3/repository/cmdb"
)

// BastionService 堡垒机服务
type BastionService struct {
	repo *repocmdb.BastionRepository
}

// NewBastionService 创建堡垒机服务实例
func NewBastionService(repo *repocmdb.BastionRepository) *BastionService {
	return &BastionService{
		repo: repo,
	}
}

// ========== 连接权限检查 ==========

// CheckConnectPermission 检查用户是否有连接指定服务器的权限
func (s *BastionService) CheckConnectPermission(userID uint, serverID uint) (bool, []modelcmdb.SSHCredential, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return false, nil, fmt.Errorf("用户不存在: %w", err)
	}

	server, err := s.repo.FindServerWithUserCredentials(serverID)
	if err != nil {
		return false, nil, fmt.Errorf("服务器不存在: %w", err)
	}

	if len(server.Credentials) == 0 {
		return false, nil, fmt.Errorf("服务器未绑定用户凭证，请先在主机编辑页面绑定 credential_type=user 的 SSH 凭证")
	}

	roleIDs, err := parseRoleIDs(user.RoleIDs)
	if err != nil {
		return false, nil, fmt.Errorf("解析用户角色失败: %w", err)
	}

	for _, roleID := range roleIDs {
		if roleID == 1 {
			return true, server.Credentials, nil
		}
	}

	policies, err := s.repo.FindAccessPoliciesByRoleIDs(roleIDs)
	if err != nil {
		return false, nil, err
	}

	hasAccess := false
	for _, policy := range policies {
		if s.matchesPolicy(policy, serverID, server) {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return false, nil, fmt.Errorf("没有访问权限")
	}

	return true, server.Credentials, nil
}

// matchesPolicy 检查策略是否匹配服务器
func (s *BastionService) matchesPolicy(policy modelcmdb.AssetAccessPolicy, serverID uint, server *modelcmdb.Server) bool {
	switch policy.AssetScopeType {
	case "all":
		return true
	case "server":
		return policy.AssetScopeID == serverID
	case "group":
		count, _ := s.repo.CountServerGroupRelation(serverID, policy.AssetScopeID)
		return count > 0
	case "tag":
		count, _ := s.repo.CountServerTagRelation(serverID, policy.AssetScopeID)
		return count > 0
	case "business":
		return false
	default:
		return false
	}
}

// ========== 会话管理 ==========

// CreateSSHSession 创建SSH会话
func (s *BastionService) CreateSSHSession(userID uint, serverID uint, credentialID uint, clientIP string, protocol string) (*modelcmdb.BastionSession, error) {
	hasPermission, allowedCredentials, err := s.CheckConnectPermission(userID, serverID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, fmt.Errorf("没有连接权限")
	}

	var chosenCredential *modelcmdb.SSHCredential
	for i := range allowedCredentials {
		if allowedCredentials[i].ID == credentialID {
			chosenCredential = &allowedCredentials[i]
			break
		}
	}
	if chosenCredential == nil {
		return nil, fmt.Errorf("不允许使用该凭证")
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	now := time.Now()
	session := &modelcmdb.BastionSession{
		ServerID:        serverID,
		UserID:          userID,
		Username:        user.Username,
		LoginAccount:    chosenCredential.Username,
		ClientIP:        clientIP,
		Protocol:        protocol,
		SSHCredentialID: credentialID,
		StartedAt:       &now,
		Status:          "active",
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	_ = s.repo.UpdateServerLastConnectTime(serverID, now)

	return session, nil
}

// CloseSession 关闭会话
func (s *BastionService) CloseSession(sessionID uint, reason string, status ...string) error {
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("会话不存在: %w", err)
	}

	if session.Status != "active" {
		return nil
	}

	now := time.Now()
	duration := int(now.Sub(*session.StartedAt).Seconds())

	targetStatus := "closed"
	if len(status) > 0 && status[0] != "" {
		targetStatus = status[0]
	}

	updates := map[string]interface{}{
		"status":       targetStatus,
		"ended_at":     now,
		"duration":     duration,
		"close_reason": reason,
	}

	return s.repo.UpdateSession(sessionID, updates)
}

// CleanupOrphanedSessions 清理孤儿会话
func CleanupOrphanedSessions() {
	repo := repocmdb.NewBastionRepository(database.GetDB())
	rowsAffected, err := repo.CleanupOrphanedSessions(time.Now())
	if err != nil {
		return
	}
	if rowsAffected > 0 {
		fmt.Printf("[startup] 清理孤儿会话 %d 条\n", rowsAffected)
	}
}

// GetActiveSessions 获取活跃会话列表
func (s *BastionService) GetActiveSessions() ([]modelcmdb.BastionSession, error) {
	sessions, err := s.repo.FindActiveSessions()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for i := range sessions {
		if sessions[i].StartedAt != nil {
			sessions[i].Duration = int(now.Sub(*sessions[i].StartedAt).Seconds())
		}
	}

	return sessions, nil
}

// TerminateSession 强制断开会话
func (s *BastionService) TerminateSession(sessionID uint, operatorID uint) error {
	return s.CloseSession(sessionID, fmt.Sprintf("被用户 %d 强制断开", operatorID), "terminated")
}

// GetSessionsList 获取会话列表（轻量级）
func (s *BastionService) GetSessionsList(filter modelcmdb.SessionFilter, page int, pageSize int) ([]repocmdb.SessionListItem, int64, error) {
	return s.repo.FindSessionsList(filter, page, pageSize)
}

// GetSessions 获取会话列表（完整信息）
func (s *BastionService) GetSessions(filter modelcmdb.SessionFilter, page int, pageSize int) ([]modelcmdb.BastionSession, int64, error) {
	sessions, total, err := s.repo.FindSessions(filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	for i := range sessions {
		if sessions[i].Status == "active" && sessions[i].StartedAt != nil {
			sessions[i].Duration = int(now.Sub(*sessions[i].StartedAt).Seconds())
		}
	}

	return sessions, total, nil
}

// GetSessionByID 获取会话详情
func (s *BastionService) GetSessionByID(sessionID uint) (*modelcmdb.BastionSession, error) {
	return s.repo.FindSessionDetailByID(sessionID)
}

// ========== 命令审计 ==========

// RecordCommand 记录命令
func (s *BastionService) RecordCommand(sessionID uint, command string, exitCode int, output string) error {
	riskLevel := s.analyzeCommandRisk(command)
	blocked := s.isCommandBlocked(command, sessionID)

	cmd := &modelcmdb.BastionCommand{
		SessionID:     sessionID,
		Command:       command,
		ExitCode:      &exitCode,
		RiskLevel:     riskLevel,
		Blocked:       blocked,
		OutputSummary: truncateString(output, 1000),
	}

	now := time.Now()
	cmd.ExecutedAt = &now

	return s.repo.CreateCommand(cmd)
}

// analyzeCommandRisk 分析命令风险等级
func (s *BastionService) analyzeCommandRisk(command string) string {
	command = strings.ToLower(strings.TrimSpace(command))

	highRiskCommands := []string{
		"rm -rf /", "rm -rf /*", "mkfs", "dd if=/dev/zero",
		"shutdown", "reboot", "halt", "poweroff",
		":(){:|:&};:",
		"chmod 000", "chattr",
	}

	for _, risky := range highRiskCommands {
		if strings.Contains(command, risky) {
			return "critical"
		}
	}

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
	session, err := s.repo.FindSessionWithServer(sessionID)
	if err != nil {
		return false
	}

	policies, err := s.repo.FindAccessPoliciesWithHighRiskByRoleIDs(getRoleIDsFromSession(session))
	if err != nil || len(policies) == 0 {
		return false
	}

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
func (s *BastionService) GetSessionCommands(sessionID uint) ([]modelcmdb.BastionCommand, error) {
	return s.repo.FindSessionCommands(sessionID)
}

// GetCommands 获取命令列表（分页）
func (s *BastionService) GetCommands(filter modelcmdb.CommandFilter, page int, pageSize int) ([]modelcmdb.BastionCommand, int64, error) {
	return s.repo.FindCommands(filter, page, pageSize)
}

// ========== 文件传输审计 ==========

// RecordFileTransfer 记录文件传输
func (s *BastionService) RecordFileTransfer(sessionID uint, direction string, remotePath string, localPath string, fileSize int64) (*modelcmdb.BastionFileTransfer, error) {
	transfer := &modelcmdb.BastionFileTransfer{
		SessionID:  sessionID,
		Direction:  direction,
		RemotePath: remotePath,
		LocalPath:  localPath,
		FileSize:   fileSize,
		Status:     "pending",
	}

	now := time.Now()
	transfer.StartedAt = &now

	if err := s.repo.CreateFileTransfer(transfer); err != nil {
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

	return s.repo.UpdateFileTransferStatus(transferID, updates)
}

// GetSessionFileTransfers 获取会话的文件传输记录
func (s *BastionService) GetSessionFileTransfers(sessionID uint) ([]modelcmdb.BastionFileTransfer, error) {
	return s.repo.FindSessionFileTransfers(sessionID)
}

// GetFileTransfers 获取文件传输列表（分页）
func (s *BastionService) GetFileTransfers(filter modelcmdb.FileTransferFilter, page int, pageSize int) ([]modelcmdb.BastionFileTransfer, int64, error) {
	return s.repo.FindFileTransfers(filter, page, pageSize)
}

// ========== 访问策略管理 ==========

// GetAccessPolicies 获取访问策略列表
func (s *BastionService) GetAccessPolicies(page int, pageSize int) ([]modelcmdb.AssetAccessPolicy, int64, error) {
	return s.repo.FindAccessPolicies(page, pageSize)
}

// CreateAccessPolicy 创建访问策略
func (s *BastionService) CreateAccessPolicy(policy *modelcmdb.AssetAccessPolicy) error {
	return s.repo.CreateAccessPolicy(policy)
}

// UpdateAccessPolicy 更新访问策略
func (s *BastionService) UpdateAccessPolicy(id uint, updates map[string]interface{}) error {
	return s.repo.UpdateAccessPolicy(id, updates)
}

// DeleteAccessPolicy 删除访问策略
func (s *BastionService) DeleteAccessPolicy(id uint) error {
	return s.repo.DeleteAccessPolicy(id)
}

// GetAccessPolicyByID 根据ID获取访问策略
func (s *BastionService) GetAccessPolicyByID(id uint) (*modelcmdb.AssetAccessPolicy, error) {
	return s.repo.FindAccessPolicyByID(id)
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
func getRoleIDsFromSession(session *modelcmdb.BastionSession) []uint {
	if session.User != nil {
		roleIDs, err := parseRoleIDs(session.User.RoleIDs)
		if err == nil {
			return roleIDs
		}
	}
	return []uint{}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
