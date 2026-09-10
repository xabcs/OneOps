package cmdb

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	modelcmdb "oneops/backend3/model/cmdb"
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	repocmdb "oneops/backend3/repository/cmdb"
	syssvc "oneops/backend3/service/system"
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

// CheckConnectPermission 检查用户是否有连接指定服务器的权限（主体 + 资产范围命中即有权限）
// 注意：协议/登录账号/时间窗口等更细粒度的约束在 CreateSSHSession 中校验
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

	if syssvc.IsSystemAdmin(s.repo.DB(), user.ID) {
		return true, server.Credentials, nil
	}

	roleIDs := userRoleIDs(user)
	policies, err := s.findApplicablePolicies(userID, roleIDs, serverID, server)
	if err != nil {
		return false, nil, err
	}
	if len(policies) == 0 {
		return false, nil, fmt.Errorf("没有访问权限")
	}

	// 批次三：按命中策略过滤可展示的凭证（LoginAccounts 白名单），
	// 与 CreateSSHSession 的实连校验同口径，不再把全部凭证暴露给无相应权限的用户
	allowedAccounts := map[string]bool{}
	for _, p := range policies {
		for _, account := range p.LoginAccounts {
			allowedAccounts[strings.ToLower(account)] = true
		}
	}
	filtered := make([]modelcmdb.SSHCredential, 0, len(server.Credentials))
	for _, cred := range server.Credentials {
		if allowedAccounts[strings.ToLower(cred.Username)] {
			filtered = append(filtered, cred)
		}
	}
	if len(filtered) == 0 {
		return false, nil, fmt.Errorf("策略未授权任何登录账号")
	}

	return true, filtered, nil
}

// findApplicablePolicies 返回主体（角色/用户）与资产范围都命中的启用策略
func (s *BastionService) findApplicablePolicies(userID uint, roleIDs []uint, serverID uint, server *modelcmdb.Server) ([]modelcmdb.AssetAccessPolicy, error) {
	policies, err := s.repo.FindAccessPoliciesBySubject(userID, roleIDs)
	if err != nil {
		return nil, err
	}

	businessAncestors, err := s.repo.FindBusinessUnitAncestorIDs(server.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("解析业务系统层级失败: %w", err)
	}

	applicable := make([]modelcmdb.AssetAccessPolicy, 0, len(policies))
	for _, policy := range policies {
		matched, err := s.matchesPolicy(policy, serverID, server, businessAncestors)
		if err != nil {
			return nil, err
		}
		if matched {
			applicable = append(applicable, policy)
		}
	}
	return applicable, nil
}

// userRoleIDs 从用户角色关联提取角色 ID 列表
func userRoleIDs(user *modelsystem.User) []uint {
	roleIDs := make([]uint, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleIDs = append(roleIDs, r.ID)
	}
	return roleIDs
}

// matchesPolicy 检查策略资产范围是否覆盖服务器（businessAncestors 为服务器所属业务系统及全部祖先 ID）
func (s *BastionService) matchesPolicy(policy modelcmdb.AssetAccessPolicy, serverID uint, server *modelcmdb.Server, businessAncestors []uint) (bool, error) {
	switch policy.AssetScopeType {
	case "all":
		return true, nil
	case "server":
		return policy.AssetScopeID == serverID, nil
	case "group":
		count, err := s.repo.CountServerGroupRelation(serverID, policy.AssetScopeID)
		if err != nil {
			return false, fmt.Errorf("查询服务器分组关联失败: %w", err)
		}
		return count > 0, nil
	case "tag":
		count, err := s.repo.CountServerTagRelation(serverID, policy.AssetScopeID)
		if err != nil {
			return false, fmt.Errorf("查询服务器标签关联失败: %w", err)
		}
		return count > 0, nil
	case "business":
		// 业务系统子树匹配：命中服务器所属业务或其任意祖先
		for _, id := range businessAncestors {
			if id == policy.AssetScopeID {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

// ========== 会话管理 ==========

// CreateSSHSession 创建SSH会话
func (s *BastionService) CreateSSHSession(userID uint, serverID uint, credentialID uint, clientIP string, protocol string) (*modelcmdb.BastionSession, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	server, err := s.repo.FindServerWithUserCredentials(serverID)
	if err != nil {
		return nil, fmt.Errorf("服务器不存在: %w", err)
	}

	if len(server.Credentials) == 0 {
		return nil, fmt.Errorf("服务器未绑定用户凭证，请先在主机编辑页面绑定 credential_type=user 的 SSH 凭证")
	}

	// 先定位凭证（账号白名单校验需要凭证的登录名）
	var chosenCredential *modelcmdb.SSHCredential
	for i := range server.Credentials {
		if server.Credentials[i].ID == credentialID {
			chosenCredential = &server.Credentials[i]
			break
		}
	}
	if chosenCredential == nil {
		return nil, fmt.Errorf("不允许使用该凭证")
	}

	// 批次三：统一超管判定（code=admin 且启用），不再硬编码 roleID=1
	if !syssvc.IsSystemAdmin(s.repo.DB(), user.ID) {
		roleIDs := userRoleIDs(user)

		policies, err := s.findApplicablePolicies(userID, roleIDs, serverID, server)
		if err != nil {
			return nil, err
		}
		if len(policies) == 0 {
			return nil, fmt.Errorf("没有访问权限")
		}

		// 更细粒度约束：协议、登录账号、时间窗口
		if ok, reason := selectSessionPolicy(policies, protocol, chosenCredential.Username, time.Now()); !ok {
			return nil, fmt.Errorf("%s", reason)
		}
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

// selectSessionPolicy 逐条策略做全维度校验：协议、登录账号、时间窗口必须由**同一条**策略同时允许才放行。
// （历史实现按维度分别扫描策略集合，可组合出单条策略都不允许的授权，属授权绕过，已修复。）
// 拒绝时返回具体原因，便于用户与排障定位
func selectSessionPolicy(policies []modelcmdb.AssetAccessPolicy, protocol string, loginAccount string, now time.Time) (bool, string) {
	protocolOK, accountOK, windowOK := false, false, false
	for _, policy := range policies {
		if !policyAllowsProtocol(policy, protocol) {
			continue
		}
		protocolOK = true
		if !policyAllowsAccount(policy, loginAccount) {
			continue
		}
		accountOK = true
		if !isWithinTimeWindow(policy.TimeWindow, now) {
			continue
		}
		windowOK = true
		return true, ""
	}
	switch {
	case !protocolOK:
		return false, fmt.Sprintf("访问策略不允许使用协议 %s", protocol)
	case !accountOK:
		return false, fmt.Sprintf("访问策略不允许使用登录账号 %s", loginAccount)
	default:
		_ = windowOK
		return false, "当前时间不在策略允许的访问时段内"
	}
}

// policyAllowsProtocol 该策略是否允许指定协议（空列表视为不限，兼容历史数据）
func policyAllowsProtocol(policy modelcmdb.AssetAccessPolicy, protocol string) bool {
	if len(policy.Protocols) == 0 {
		return true
	}
	for _, p := range policy.Protocols {
		if p == protocol {
			return true
		}
	}
	return false
}

// policyAllowsAccount 该策略是否允许指定登录账号（空列表视为不限；大小写不敏感）
func policyAllowsAccount(policy modelcmdb.AssetAccessPolicy, loginAccount string) bool {
	if len(policy.LoginAccounts) == 0 {
		return true
	}
	for _, account := range policy.LoginAccounts {
		if strings.EqualFold(account, loginAccount) {
			return true
		}
	}
	return false
}

// isWithinTimeWindow 检查时间是否落在策略时间窗口内
// 约定：start/end 为空表示不限时；days 非空时星期限制独立生效（1=周一 ... 7=周日）；
// start > end 表示跨天窗口（如 22:00-06:00 值班时段）
func isWithinTimeWindow(tw modelcmdb.TimeWindow, now time.Time) bool {
	// 星期限制独立生效：仅配置 days（时间留空）表示"仅这些天的全天"
	if len(tw.Days) > 0 {
		// Go: Sunday=0...Saturday=6 → 转为 1=周一...7=周日
		goDay := int(now.Weekday())
		if goDay == 0 {
			goDay = 7
		}
		matched := false
		for _, d := range tw.Days {
			if d == goDay {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if tw.Start == "" || tw.End == "" {
		return true
	}

	current := now.Format("15:04")
	if tw.Start <= tw.End {
		return current >= tw.Start && current <= tw.End
	}
	// 跨天窗口：22:00-06:00 → 当前 >= 22:00 或 <= 06:00
	return current >= tw.Start || current <= tw.End
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
	// StartedAt 理论上必填，但历史脏数据可能为 NULL，解引用前先做防护
	duration := 0
	if session.StartedAt != nil {
		duration = int(now.Sub(*session.StartedAt).Seconds())
	}

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
	// 属主校验：仅会话属主或系统管理员可强制断开，防越权终止他人会话
	if err := s.AuthorizeTerminateSession(sessionID, operatorID); err != nil {
		return err
	}
	return s.CloseSession(sessionID, fmt.Sprintf("被用户 %d 强制断开", operatorID), "terminated")
}

// AuthorizeTerminateSession 仅校验终止会话的权限（属主或系统管理员），无任何副作用。
// 调用方必须**先**通过本校验，再拆除内存中的 WS/SSH 连接——避免"先杀连接、后发现无权"的不可回滚副作用
func (s *BastionService) AuthorizeTerminateSession(sessionID uint, operatorID uint) error {
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("会话不存在: %w", err)
	}
	if session.UserID != operatorID && !syssvc.IsSystemAdmin(s.repo.DB(), operatorID) {
		return fmt.Errorf("无权终止他人的会话")
	}
	return nil
}

// CanViewAllSessionAudits 判断用户是否可查看全部会话审计数据（系统管理员）；普通用户只能查看自己的
func (s *BastionService) CanViewAllSessionAudits(userID uint) bool {
	return syssvc.IsSystemAdmin(s.repo.DB(), userID)
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

// RecordCommand 记录命令。blocked 由调用方（WS 转发层）传入拦截判定结果，
// 避免同一条命令重复执行"会话+策略"两次 DB 查询
func (s *BastionService) RecordCommand(sessionID uint, command string, exitCode int, output string, blocked bool) error {
	riskLevel := s.analyzeCommandRisk(command)

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

// IsCommandBlocked 判断命令是否命中拦截（H7：供 WS 转发层在写入 SSH 之前调用，实现先判后执行）
func (s *BastionService) IsCommandBlocked(command string, sessionID uint) bool {
	session, err := s.repo.FindSessionWithServer(sessionID)
	if err != nil {
		return false
	}

	policies, err := s.repo.FindAccessPoliciesWithHighRisk(session.UserID, getRoleIDsFromSession(session))
	if err != nil || len(policies) == 0 {
		return false
	}
	return matchBlockedCommand(policies, command)
}

// IsCommandBlockedForUser 按用户判定命令是否命中高危命令策略（供 K8s 容器终端等
// 非堡垒机会话场景复用，与 SSH 终端同口径：策略黑名单 + 归一化匹配）
func (s *BastionService) IsCommandBlockedForUser(userID uint, command string) bool {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return false
	}
	policies, err := s.repo.FindAccessPoliciesWithHighRisk(userID, userRoleIDs(user))
	if err != nil || len(policies) == 0 {
		return false
	}
	return matchBlockedCommand(policies, command)
}

// matchBlockedCommand 黑名单匹配：原始文本子串 + 归一化文本子串双路匹配
func matchBlockedCommand(policies []modelcmdb.AssetAccessPolicy, command string) bool {
	commandLower := strings.ToLower(command)
	normalizedCommand := normalizeCommandForMatch(command)
	for _, policy := range policies {
		for _, blockedCmd := range policy.HighRiskCommands {
			// 先按原始文本子串匹配（兼容常规场景），再按归一化文本匹配（对抗混淆绕过）
			if strings.Contains(commandLower, strings.ToLower(blockedCmd)) {
				return true
			}
			if nb := normalizeCommandForMatch(blockedCmd); nb != "" && strings.Contains(normalizedCommand, nb) {
				return true
			}
		}
	}
	return false
}

// normalizeCommandForMatch 将命令/策略串归一化后用于黑名单匹配，对抗常见混淆绕过：
// - 剥 ANSI 转义序列与不可见控制字符（含 bracketed paste 标记）
// - ${IFS}/$IFS → 空格（shell 默认分隔符替换）
// - 反斜杠续行（行尾 \ + 换行）→ 直接拼接
// - \t、\r、\n → 空格；连续空白压缩为单个空格
// - 删除引号与反斜杠（对抗 'r'm、r\m、""拼接混淆）
// 归一化以"宁可误拦"为原则，仅用于拦截匹配；审计记录仍保留原始文本
func normalizeCommandForMatch(cmd string) string {
	// 剥 ANSI 转义序列与控制字符（保留可打印 ASCII 与高位 UTF-8 字节）
	var b strings.Builder
	inEscape := false
	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		if c == 0x1b { // ESC
			inEscape = true
			continue
		}
		if inEscape {
			// CSI 序列终止符：字母或 '~'（如 bracketed paste 标记 \x1b[200~）
			if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '~' {
				inEscape = false
			}
			continue
		}
		if c >= 32 && c <= 126 || c == '\t' || c >= 0x80 { // 可打印 + Tab + UTF-8 续字节
			b.WriteByte(c)
		}
	}
	s := b.String()

	s = strings.ReplaceAll(s, "${IFS}", " ")
	s = strings.ReplaceAll(s, "$IFS", " ")
	s = strings.ReplaceAll(s, "\\\r\n", "")
	s = strings.ReplaceAll(s, "\\\n", "")
	s = strings.ReplaceAll(s, "\\\r", "")
	for _, ch := range []string{"\t", "\r", "\n"} {
		s = strings.ReplaceAll(s, ch, " ")
	}
	for _, ch := range []string{"'", "\"", "\\"} {
		s = strings.ReplaceAll(s, ch, "")
	}

	// 压缩连续空白
	var out strings.Builder
	inSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			inSpace = true
			continue
		}
		if inSpace && out.Len() > 0 {
			out.WriteByte(' ')
		}
		inSpace = false
		out.WriteRune(r)
	}
	return strings.ToLower(strings.TrimSpace(out.String()))
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

// 访问策略字段取值约束
var (
	validPolicySubjectTypes    = map[string]bool{"user": true, "role": true, "user_group": true}
	validPolicyAssetScopeTypes = map[string]bool{"server": true, "group": true, "business": true, "tag": true, "all": true}
	validPolicyProtocols       = map[string]bool{"ssh": true, "sftp": true}
)

// CreateAccessPolicy 创建访问策略
func (s *BastionService) CreateAccessPolicy(policy *modelcmdb.AssetAccessPolicy) error {
	if err := s.validateAccessPolicy(policy, 0); err != nil {
		return err
	}
	return s.repo.CreateAccessPolicy(policy)
}

// UpdateAccessPolicy 更新访问策略（字段白名单 + 与创建一致的校验）
func (s *BastionService) UpdateAccessPolicy(id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	existing, err := s.repo.FindAccessPolicyByID(id)
	if err != nil {
		return err
	}

	columns, patched, err := s.sanitizeAccessPolicyUpdates(existing, updates)
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	if err := s.validateAccessPolicy(patched, id); err != nil {
		return err
	}
	return s.repo.UpdateAccessPolicy(id, columns)
}

// validateAccessPolicy 校验访问策略的业务约束（excludeID 用于名称唯一性校验，创建时传 0）
func (s *BastionService) validateAccessPolicy(policy *modelcmdb.AssetAccessPolicy, excludeID uint) error {
	// 名称
	name := strings.TrimSpace(policy.Name)
	if name == "" {
		return fmt.Errorf("策略名称不能为空")
	}
	if len(name) > 100 {
		return fmt.Errorf("策略名称不能超过 100 个字符")
	}
	policy.Name = name

	existingID, err := s.repo.FindAccessPolicyIDByName(name, excludeID)
	if err != nil {
		return fmt.Errorf("校验策略名称唯一性失败: %w", err)
	}
	if existingID > 0 {
		return fmt.Errorf("策略名称 %q 已存在", name)
	}

	// 授权对象
	if !validPolicySubjectTypes[policy.SubjectType] {
		return fmt.Errorf("无效的授权对象类型: %s", policy.SubjectType)
	}
	if policy.SubjectID == 0 {
		return fmt.Errorf("授权对象不能为空")
	}
	if err := s.validateSubjectExists(policy.SubjectType, policy.SubjectID); err != nil {
		return err
	}

	// 资产范围
	if !validPolicyAssetScopeTypes[policy.AssetScopeType] {
		return fmt.Errorf("无效的资产范围类型: %s", policy.AssetScopeType)
	}
	if policy.AssetScopeType != "all" {
		if policy.AssetScopeID == 0 {
			return fmt.Errorf("资产范围不能为空")
		}
		if err := s.validateAssetScopeExists(policy.AssetScopeType, policy.AssetScopeID); err != nil {
			return err
		}
	} else {
		policy.AssetScopeID = 0
	}

	// 登录账号
	accounts := trimPolicyStrings(policy.LoginAccounts)
	if len(accounts) == 0 {
		return fmt.Errorf("至少需要一个登录账号")
	}
	policy.LoginAccounts = accounts

	// 协议
	if len(policy.Protocols) == 0 {
		return fmt.Errorf("至少需要允许一种协议")
	}
	for _, p := range policy.Protocols {
		if !validPolicyProtocols[p] {
			return fmt.Errorf("不支持的协议: %s", p)
		}
	}

	// 高危命令
	policy.HighRiskCommands = trimPolicyStrings(policy.HighRiskCommands)

	// 时间窗口
	if err := validatePolicyTimeWindow(policy.TimeWindow); err != nil {
		return err
	}

	// 状态
	if policy.Status != 0 && policy.Status != 1 {
		return fmt.Errorf("无效的状态值: %d", policy.Status)
	}
	return nil
}

// validateSubjectExists 校验授权对象存在性
func (s *BastionService) validateSubjectExists(subjectType string, subjectID uint) error {
	var model interface{}
	switch subjectType {
	case "user":
		model = &modelsystem.User{}
	case "role":
		model = &modelsystem.Role{}
	case "user_group":
		// 用户组暂无管理功能，跳过存在性校验
		return nil
	}

	exists, err := s.repo.ExistsRecord(model, subjectID)
	if err != nil {
		return fmt.Errorf("校验授权对象失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("授权对象不存在（类型: %s, ID: %d）", subjectType, subjectID)
	}
	return nil
}

// validateAssetScopeExists 校验资产范围对象存在性
func (s *BastionService) validateAssetScopeExists(scopeType string, scopeID uint) error {
	var model interface{}
	switch scopeType {
	case "server":
		model = &modelcmdb.Server{}
	case "group":
		model = &modelcmdb.ServerGroup{}
	case "business":
		model = &modelcmdb.BusinessUnit{}
	case "tag":
		model = &modelcmdb.ServerTag{}
	default:
		return nil
	}

	exists, err := s.repo.ExistsRecord(model, scopeID)
	if err != nil {
		return fmt.Errorf("校验资产范围失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("资产范围对象不存在（类型: %s, ID: %d）", scopeType, scopeID)
	}
	return nil
}

// validatePolicyTimeWindow 校验时间窗口（start/end 为空表示不限制）
func validatePolicyTimeWindow(tw modelcmdb.TimeWindow) error {
	if tw.Start == "" && tw.End == "" {
		return nil
	}
	if tw.Start == "" || tw.End == "" {
		return fmt.Errorf("时间窗口需要同时提供开始和结束时间")
	}
	if _, err := time.Parse("15:04", tw.Start); err != nil {
		return fmt.Errorf("开始时间格式错误，应为 HH:mm")
	}
	if _, err := time.Parse("15:04", tw.End); err != nil {
		return fmt.Errorf("结束时间格式错误，应为 HH:mm")
	}
	if tw.Start == tw.End {
		return fmt.Errorf("开始时间不能等于结束时间")
	}
	// start > end 视为跨天窗口（如 22:00-06:00 值班时段），允许配置
	for _, d := range tw.Days {
		if d < 1 || d > 7 {
			return fmt.Errorf("星期取值必须在 1-7 之间")
		}
	}
	return nil
}

// sanitizeAccessPolicyUpdates 过滤出允许更新的字段并转换为数据库列，
// 同时把补丁应用到 existing 的副本上，供 validateAccessPolicy 统一校验；
// id/createdAt/updatedAt 等只读字段会被忽略
func (s *BastionService) sanitizeAccessPolicyUpdates(
	existing *modelcmdb.AssetAccessPolicy,
	updates map[string]interface{},
) (map[string]interface{}, *modelcmdb.AssetAccessPolicy, error) {
	patched := *existing
	columns := make(map[string]interface{})

	for key, value := range updates {
		switch key {
		case "name":
			v, ok := value.(string)
			if !ok {
				return nil, nil, fmt.Errorf("字段 name 类型错误")
			}
			v = strings.TrimSpace(v)
			patched.Name = v
			columns["name"] = v
		case "subjectType":
			v, ok := value.(string)
			if !ok {
				return nil, nil, fmt.Errorf("字段 subjectType 类型错误")
			}
			patched.SubjectType = v
			columns["subject_type"] = v
		case "subjectId":
			v, err := toPolicyUint(value)
			if err != nil {
				return nil, nil, fmt.Errorf("字段 subjectId %w", err)
			}
			patched.SubjectID = v
			columns["subject_id"] = v
		case "assetScopeType":
			v, ok := value.(string)
			if !ok {
				return nil, nil, fmt.Errorf("字段 assetScopeType 类型错误")
			}
			patched.AssetScopeType = v
			columns["asset_scope_type"] = v
		case "assetScopeId":
			v, err := toPolicyUint(value)
			if err != nil {
				return nil, nil, fmt.Errorf("字段 assetScopeId %w", err)
			}
			patched.AssetScopeID = v
			columns["asset_scope_id"] = v
		case "allowFileTransfer":
			v, ok := value.(bool)
			if !ok {
				return nil, nil, fmt.Errorf("字段 allowFileTransfer 类型错误")
			}
			patched.AllowFileTransfer = v
			columns["allow_file_transfer"] = v
		case "allowSudo":
			v, ok := value.(bool)
			if !ok {
				return nil, nil, fmt.Errorf("字段 allowSudo 类型错误")
			}
			patched.AllowSudo = v
			columns["allow_sudo"] = v
		case "requireApproval":
			v, ok := value.(bool)
			if !ok {
				return nil, nil, fmt.Errorf("字段 requireApproval 类型错误")
			}
			patched.RequireApproval = v
			columns["require_approval"] = v
		case "status":
			v, err := toPolicyInt(value)
			if err != nil {
				return nil, nil, fmt.Errorf("字段 status %w", err)
			}
			patched.Status = v
			columns["status"] = v
		case "loginAccounts":
			v, err := toPolicyStringArray(value)
			if err != nil {
				return nil, nil, fmt.Errorf("字段 loginAccounts %w", err)
			}
			encoded, err := json.Marshal(v)
			if err != nil {
				return nil, nil, fmt.Errorf("序列化字段 loginAccounts 失败: %w", err)
			}
			patched.LoginAccounts = v
			columns["login_accounts"] = string(encoded)
		case "protocols":
			v, err := toPolicyStringArray(value)
			if err != nil {
				return nil, nil, fmt.Errorf("字段 protocols %w", err)
			}
			encoded, err := json.Marshal(v)
			if err != nil {
				return nil, nil, fmt.Errorf("序列化字段 protocols 失败: %w", err)
			}
			patched.Protocols = v
			columns["protocols"] = string(encoded)
		case "highRiskCommands":
			v, err := toPolicyStringArray(value)
			if err != nil {
				return nil, nil, fmt.Errorf("字段 highRiskCommands %w", err)
			}
			encoded, err := json.Marshal(v)
			if err != nil {
				return nil, nil, fmt.Errorf("序列化字段 highRiskCommands 失败: %w", err)
			}
			patched.HighRiskCommands = v
			columns["high_risk_commands"] = string(encoded)
		case "timeWindow":
			switch tw := value.(type) {
			case nil:
				patched.TimeWindow = modelcmdb.TimeWindow{}
				columns["time_window"] = nil
			case map[string]interface{}:
				data, err := json.Marshal(tw)
				if err != nil {
					return nil, nil, fmt.Errorf("字段 timeWindow 格式错误")
				}
				var window modelcmdb.TimeWindow
				if err := json.Unmarshal(data, &window); err != nil {
					return nil, nil, fmt.Errorf("字段 timeWindow 格式错误")
				}
				patched.TimeWindow = window
				columns["time_window"] = string(data)
			default:
				return nil, nil, fmt.Errorf("字段 timeWindow 类型错误")
			}
		default:
			// 忽略未知/只读字段
		}
	}

	return columns, &patched, nil
}

// trimPolicyStrings 去除空白并过滤空字符串
func trimPolicyStrings(arr modelcmdb.StringArray) modelcmdb.StringArray {
	result := make(modelcmdb.StringArray, 0, len(arr))
	for _, s := range arr {
		if s = strings.TrimSpace(s); s != "" {
			result = append(result, s)
		}
	}
	return result
}

// toPolicyUint 将 JSON 数值转换为非负整数
func toPolicyUint(value interface{}) (uint, error) {
	switch v := value.(type) {
	case float64:
		if v < 0 || v != math.Trunc(v) {
			return 0, fmt.Errorf("必须是正整数")
		}
		return uint(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("必须是正整数")
		}
		return uint(v), nil
	default:
		return 0, fmt.Errorf("类型错误")
	}
}

// toPolicyInt 将 JSON 数值转换为整数
func toPolicyInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case float64:
		if v != math.Trunc(v) {
			return 0, fmt.Errorf("必须是整数")
		}
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("类型错误")
	}
}

// toPolicyStringArray 将 JSON 数组转换为去除空白的字符串数组
func toPolicyStringArray(value interface{}) (modelcmdb.StringArray, error) {
	switch v := value.(type) {
	case []interface{}:
		result := make(modelcmdb.StringArray, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("数组元素必须是字符串")
			}
			result = append(result, strings.TrimSpace(s))
		}
		return result, nil
	case []string:
		result := make(modelcmdb.StringArray, 0, len(v))
		for _, s := range v {
			result = append(result, strings.TrimSpace(s))
		}
		return result, nil
	default:
		return nil, fmt.Errorf("必须是字符串数组")
	}
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

// getRoleIDsFromSession 从会话中获取角色ID
func getRoleIDsFromSession(session *modelcmdb.BastionSession) []uint {
	if session.User != nil {
		return userRoleIDs(session.User)
	}
	return []uint{}
}

// truncateString 截断字符串（按 rune 截断，避免多字节中文被从中间截开产生乱码）
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
