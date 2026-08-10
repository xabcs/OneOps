package authorization

import (
	"encoding/json"
	"fmt"
	"time"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/logger"
	appadapter "oneops/backend3/service/authorization/adapter"

	"go.uber.org/zap"
)

// AssignUserToGroup 为用户分配用户组成员，返回授权结果
func (s *ApplicationPermissionService) AssignUserToGroup(userID, groupID uint, operator string) ([]ExternalUserResult, error) {
	count, _ := s.repo.CountUserGroup(userID, groupID)
	if count > 0 {
		return nil, fmt.Errorf("用户已分配该用户组")
	}

	bindings, err := s.repo.FindGroupBindingsByGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("获取用户组绑定失败: %w", err)
	}

	if len(bindings) == 0 {
		if _, err := s.repo.FindAuthUserByID(userID); err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %w", err)
		}

		if _, err := s.repo.FindAuthGroupByID(groupID); err != nil {
			return nil, fmt.Errorf("获取用户组信息失败: %w", err)
		}

		userGroup := modelauth.AuthUserGroup{
			UserID:    userID,
			GroupID:   groupID,
			GrantedBy: operator,
			GrantedAt: time.Now(),
		}

		if err := s.repo.CreateUserGroup(&userGroup); err != nil {
			return nil, fmt.Errorf("创建用户组成员失败: %w", err)
		}

		return []ExternalUserResult{}, nil
	}

	user, err := s.repo.FindAuthUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	results := make([]ExternalUserResult, 0)

	if user.Password == "" {
		return nil, fmt.Errorf("用户未设置初始密码，请先重置用户密码")
	}

	for _, binding := range bindings {
		result := ExternalUserResult{
			AppName:  binding.AppIDField.Name,
			Username: user.Username,
			RoleCode: binding.ApplicationRole.RoleCode,
			RoleName: binding.ApplicationRole.RoleName,
			Password: user.Password,
			Success:  true,
		}

		if err := s.CreateExternalUser(binding.AppID, user.ID, user.Username, user.Password, user.Email, user.Nickname, operator); err != nil {
			result.CreateError = err.Error()
		}

		if err := s.GrantRoleToUser(binding.AppID, user.ID, binding.ID, user.Username, binding.ApplicationRole.RoleCode, operator); err != nil {
			result.Success = false
			result.GrantError = err.Error()
			results = append(results, result)
			continue
		}

		results = append(results, result)
	}

	userGroup := modelauth.AuthUserGroup{
		UserID:    userID,
		GroupID:   groupID,
		GrantedBy: operator,
		GrantedAt: time.Now(),
	}

	if err := s.repo.CreateUserGroup(&userGroup); err != nil {
		return results, fmt.Errorf("保存用户组成员失败: %w", err)
	}

	return results, nil
}

// GetApplicationRoles 获取应用角色列表
func (s *ApplicationPermissionService) GetApplicationRoles(appID uint) ([]*modelauth.ApplicationRole, error) {
	return s.repo.FindRolesByAppID(appID)
}

// GetApplicationUsers 获取应用用户列表
func (s *ApplicationPermissionService) GetApplicationUsers(appID uint) ([]*modelauth.ApplicationUser, error) {
	return s.repo.FindUsersByAppID(appID)
}

// CreateGroupBinding 创建用户组绑定
func (s *ApplicationPermissionService) CreateGroupBinding(binding *modelauth.GroupBinding) error {
	group, err := s.repo.FindAuthGroupByID(binding.GroupID)
	if err != nil {
		return fmt.Errorf("用户组不存在 (ID: %d): %w", binding.GroupID, err)
	}

	app, err := s.repo.FindApplicationByID(binding.AppID)
	if err != nil {
		return fmt.Errorf("应用不存在 (ID: %d): %w", binding.AppID, err)
	}

	if app.Type == "jumpserver" {
		return s.createJumpserverGroupBinding(binding, group, app)
	}

	role, err := s.repo.FindRoleByAppIDAndID(binding.AppID, binding.ApplicationRoleID)
	if err != nil {
		return fmt.Errorf("应用角色不存在 (RoleID: %d, AppID: %d): %w", binding.ApplicationRoleID, binding.AppID, err)
	}

	if err := s.repo.CreateGroupBinding(binding); err != nil {
		return fmt.Errorf("创建用户组绑定失败: %w", err)
	}

	go s.syncExistingMembersToExternalSystem(binding.GroupID, binding.AppID, binding.ApplicationRoleID, binding.ID, role.RoleCode, role.RoleName, "system")

	return nil
}

// GetGroupBindings 获取用户组的所有绑定
func (s *ApplicationPermissionService) GetGroupBindings(groupID uint) ([]*modelauth.GroupBinding, error) {
	return s.repo.FindGroupBindingsByGroupID(groupID)
}

// DeleteGroupBinding 删除用户组绑定
func (s *ApplicationPermissionService) DeleteGroupBinding(id uint) error {
	return s.repo.DeleteGroupBinding(id)
}

// UserGroupResponse 用户组响应结构
type UserGroupResponse struct {
	ID        uint      `json:"id"`
	GroupID   uint      `json:"groupId"`
	GroupName string    `json:"groupName"`
	GroupCode string    `json:"groupCode"`
	GrantedBy string    `json:"grantedBy"`
	GrantedAt time.Time `json:"grantedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetUserGroups 获取用户的用户组列表
func (s *ApplicationPermissionService) GetUserGroups(userID uint) ([]UserGroupResponse, error) {
	userGroups, err := s.repo.FindUserGroupsByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]UserGroupResponse, 0, len(userGroups))
	for _, ug := range userGroups {
		result = append(result, UserGroupResponse{
			ID:        ug.ID,
			GroupID:   ug.GroupID,
			GroupName: ug.GroupIDField.Name,
			GroupCode: ug.GroupIDField.Code,
			GrantedBy: ug.GrantedBy,
			GrantedAt: ug.GrantedAt,
			CreatedAt: ug.CreatedAt,
		})
	}

	return result, nil
}

// DeleteUserGroup 删除用户组成员
func (s *ApplicationPermissionService) DeleteUserGroup(userID, groupID uint) error {
	return s.repo.DeleteUserGroup(userID, groupID)
}

// GetOperationLogs 获取操作日志
func (s *ApplicationPermissionService) GetOperationLogs(appID uint, page, pageSize int) ([]*modelauth.ApplicationOperationLog, int64, error) {
	return s.repo.FindOperationLogs(appID, page, pageSize)
}

// logOperation 记录操作日志
func (s *ApplicationPermissionService) logOperation(appID uint, operation, target, requestData, status, errorMsg, operator string) {
	log := modelauth.ApplicationOperationLog{
		AppID:       appID,
		Operation:   operation,
		Target:      target,
		RequestData: requestData,
		Status:      status,
		ErrorMsg:    errorMsg,
		Operator:    operator,
	}

	_ = s.repo.CreateOperationLog(&log)
}

// getString 从 map 中获取字符串值
func (s *ApplicationPermissionService) getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// recordGroupBindingExecution 记录角色绑定执行状态
func (s *ApplicationPermissionService) recordGroupBindingExecution(groupBindingID, authUserID uint, externalUsername, actionType, status, message, operator string) {
	execution := modelauth.GroupBindingExecution{
		GroupBindingID:   groupBindingID,
		AuthUserID:       authUserID,
		ExternalUsername: externalUsername,
		ActionType:       actionType,
		Status:           status,
		Message:          message,
		Operator:         operator,
	}

	if err := s.repo.CreateGroupBindingExecution(&execution); err != nil {
		logger.Warn("记录角色绑定执行状态失败",
			zap.Uint("group_binding_id", groupBindingID),
			zap.Uint("auth_user_id", authUserID),
			zap.Error(err))
	}
}

// === 权限执行记录 ===

// GetGroupBindingExecutions 获取权限绑定执行记录
func (s *ApplicationPermissionService) GetGroupBindingExecutions(bindingID uint) ([]modelauth.GroupBindingExecution, error) {
	executions, err := s.repo.FindGroupBindingExecutionsByBindingID(bindingID)
	if err != nil {
		return nil, fmt.Errorf("获取执行记录失败: %w", err)
	}
	return executions, nil
}

// createJumpserverGroupBinding 创建 Jumpserver 用户组绑定
func (s *ApplicationPermissionService) createJumpserverGroupBinding(binding *modelauth.GroupBinding, group *modelauth.AuthGroup, app *modelauth.Application) error {
	// Jumpserver 不使用 application_roles 表
	// binding.ApplicationRoleID 字段用于存储授权规则 ID

	// 1. 检查授权规则是否存在（从 auth_authorization_rules 表）
	var authRule modelauth.ApplicationAuthorizationRule
	ruleID := binding.ApplicationRoleID // 这里的 ApplicationRoleID 实际存储的是授权规则的数据库 ID

	// 尝试从数据库 ID 获取授权规则
	if found, err := s.repo.FindAuthRuleByID(ruleID); err != nil {
		// 如果从数据库 ID 找不到，尝试从 binding 中获取 Jumpserver 规则 ID
		// 可能前端直接传递了 Jumpserver 的规则 UUID
		logger.Warn("从数据库ID获取授权规则失败，尝试从其他字段获取",
			zap.Uint("rule_database_id", ruleID),
			zap.Error(err))

		// 尝试查询是否有匹配的规则
		rules, listErr := s.repo.FindAuthRulesByAppID(app.ID)
		if listErr != nil {
			return fmt.Errorf("获取授权规则列表失败: %w", listErr)
		}

		if len(rules) == 0 {
			return fmt.Errorf("应用 %s 还没有同步授权规则，请先同步授权规则", app.Name)
		}

		// 使用第一个规则作为默认规则（实际应该让用户选择）
		authRule = rules[0]
		logger.Info("使用默认授权规则",
			zap.String("rule_id", authRule.RuleID),
			zap.String("rule_name", authRule.RuleName))
	} else {
		authRule = *found
	}

	// 2. 创建绑定记录
	// 注意：对于 Jumpserver，我们将授权规则的数据库 ID 存储在 ApplicationRoleID 字段
	binding.ApplicationRoleID = authRule.ID
	if err := s.repo.CreateGroupBinding(binding); err != nil {
		return fmt.Errorf("创建用户组绑定失败: %w", err)
	}

	// 3. 为该用户组的所有现有成员在 Jumpserver 中创建用户并添加到授权规则
	go s.syncExistingMembersToJumpserver(binding.GroupID, binding.AppID, binding.ID, authRule.RuleID, authRule.RuleName, "system")

	logger.Info("创建 Jumpserver 用户组绑定成功，开始为现有成员同步权限",
		zap.Uint("group_id", binding.GroupID),
		zap.String("group_name", group.Name),
		zap.Uint("app_id", binding.AppID),
		zap.String("app_name", app.Name),
		zap.String("rule_id", authRule.RuleID),
		zap.String("rule_name", authRule.RuleName))

	return nil
}

// syncExistingMembersToJumpserver 为用户组的现有成员同步 Jumpserver 权限
func (s *ApplicationPermissionService) syncExistingMembersToJumpserver(groupID, appID, groupBindingID uint, jumpserverRuleID, ruleName, operator string) {
	// 获取用户组的所有成员
	userGroups, err := s.repo.FindUserGroupsByGroupID(groupID)
	if err != nil {
		logger.Error("获取用户组成员失败，无法同步 Jumpserver 权限",
			zap.Uint("group_id", groupID),
			zap.Error(err))
		return
	}

	logger.Info("开始为用户组成员同步 Jumpserver 权限",
		zap.Uint("group_id", groupID),
		zap.Int("member_count", len(userGroups)),
		zap.Uint("app_id", appID),
		zap.String("rule_id", jumpserverRuleID))

	// 获取应用配置
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		logger.Error("获取应用配置失败",
			zap.Uint("app_id", appID),
			zap.Error(err))
		return
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		logger.Error("解析认证配置失败",
			zap.Uint("app_id", appID),
			zap.Error(err))
		return
	}

	// 获取 Jumpserver 适配器
	adapter, err := s.adapterFactory.GetAdapter("jumpserver")
	if err != nil {
		logger.Error("获取 Jumpserver 适配器失败",
			zap.Error(err))
		return
	}

	jumpserverAdapter, ok := adapter.(*appadapter.JumpserverAdapter)
	if !ok {
		logger.Error("适配器类型转换失败")
		return
	}

	// 收集需要添加的用户 ID（Jumpserver 的用户 ID）
	var jumpserverUserIDs []string
	userMapping := make(map[string]uint) // jumpserverUserID -> authUserID

	successCount := 0
	failCount := 0

	// 为每个成员创建 Jumpserver 用户（如果不存在）
	for _, userGroup := range userGroups {
		user := userGroup.UserIDField

		// 检查用户是否有密码
		if user.Password == "" {
			logger.Warn("用户未设置初始密码，跳过 Jumpserver 用户创建",
				zap.Uint("user_id", user.ID),
				zap.String("username", user.Username))
			failCount++
			continue
		}

		// 1. 创建 Jumpserver 用户
		if err := s.CreateExternalUser(appID, user.ID, user.Username, user.Password, user.Email, user.Nickname, operator); err != nil {
			logger.Warn("创建 Jumpserver 用户失败",
				zap.Uint("user_id", user.ID),
				zap.String("username", user.Username),
				zap.Error(err))
			// 继续执行，用户可能已存在
		}

		// 2. 获取 Jumpserver 用户 ID
		mapping, err := s.repo.FindUserIdentityMappingByUserAndApp(user.ID, appID)
		if err != nil {
			logger.Warn("获取用户身份映射失败",
				zap.Uint("user_id", user.ID),
				zap.String("username", user.Username),
				zap.Error(err))
			failCount++
			continue
		}

		// 如果有外部用户 ID，添加到列表
		if mapping.ExternalUserID != "" {
			jumpserverUserIDs = append(jumpserverUserIDs, mapping.ExternalUserID)
			userMapping[mapping.ExternalUserID] = user.ID
		} else {
			logger.Warn("用户身份映射中没有 Jumpserver 用户 ID",
				zap.Uint("user_id", user.ID),
				zap.String("username", user.Username))
			failCount++
			continue
		}

		// 记录执行成功
		s.recordGroupBindingExecution(groupBindingID, user.ID, user.Username, "created", "success", "Jumpserver 用户创建成功", operator)
		successCount++
	}

	// 3. 批量添加用户到授权规则
	if len(jumpserverUserIDs) > 0 {
		if err := jumpserverAdapter.AddUsersToAuthorizationRule(app.BaseURL, authConfig, jumpserverRuleID, jumpserverUserIDs); err != nil {
			logger.Error("批量添加用户到授权规则失败",
				zap.String("rule_id", jumpserverRuleID),
				zap.Any("user_i_ds", jumpserverUserIDs),
				zap.Error(err))

			// 记录失败
			for _, jumpserverUserID := range jumpserverUserIDs {
				if authUserID, ok := userMapping[jumpserverUserID]; ok {
					s.recordGroupBindingExecution(groupBindingID, authUserID, "", "granted", "failed", fmt.Sprintf("添加到授权规则失败: %s", err.Error()), operator)
				}
			}
			return
		}

		// 记录成功
		for _, jumpserverUserID := range jumpserverUserIDs {
			if authUserID, ok := userMapping[jumpserverUserID]; ok {
				s.recordGroupBindingExecution(groupBindingID, authUserID, "", "granted", "success", "已添加到授权规则", operator)
			}
		}
	}

	logger.Info("用户组成员 Jumpserver 权限同步完成",
		zap.Uint("group_id", groupID),
		zap.Int("total_members", len(userGroups)),
		zap.Int("success_count", successCount),
		zap.Int("fail_count", failCount),
		zap.Int("added_to_rule", len(jumpserverUserIDs)))
}
