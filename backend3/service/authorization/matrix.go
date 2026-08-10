package authorization

import (
	"encoding/json"
	"fmt"
	"time"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/logger"
	repoauth "oneops/backend3/repository/authorization"
	appadapter "oneops/backend3/service/authorization/adapter"

	"go.uber.org/zap"
)

// === 用户身份映射管理 ===

// GetUserIdentityMappings 获取用户身份映射列表
func (s *ApplicationPermissionService) GetUserIdentityMappings(page, pageSize int, username string, appID uint, status string) ([]*modelauth.UserIdentityMapping, int64, error) {
	return s.repo.FindUserIdentityMappings(page, pageSize, username, appID, status)
}

// DeleteUserIdentityMapping 删除用户身份映射
func (s *ApplicationPermissionService) DeleteUserIdentityMapping(id uint) error {
	return s.repo.DeleteUserIdentityMapping(id)
}

// === 用户有效权限查询 ===

// GetUserEffectivePermissions 获取用户有效权限列表
func (s *ApplicationPermissionService) GetUserEffectivePermissions(page, pageSize int, username string, appID uint) ([]map[string]interface{}, int64, error) {
	return s.repo.FindUserEffectivePermissions(page, pageSize, username, appID)
}

// === 矩阵视图 ===

// GetUserEffectivePermissionsMatrix 获取用户有效权限矩阵视图
func (s *ApplicationPermissionService) GetUserEffectivePermissionsMatrix(appID uint) (map[string]interface{}, error) {
	app, err := s.repo.FindApplicationByID(appID)
	if err != nil {
		return nil, fmt.Errorf("应用不存在: %w", err)
	}

	switch app.Type {
	case "jenkins", "gitlab":
		return s.getRoleUserMatrix(appID, app.Name, app.Type)
	case "jumpserver":
		return s.getUserRuleMatrix(appID, app.Name)
	default:
		return s.getRoleUserMatrix(appID, app.Name, app.Type)
	}
}

// getRoleUserMatrix 获取角色-用户矩阵（Jenkins/GitLab）
func (s *ApplicationPermissionService) getRoleUserMatrix(appID uint, appName string, appType string) (map[string]interface{}, error) {
	roles, err := s.repo.FindRolesByAppIDOrdered(appID)
	if err != nil {
		return nil, fmt.Errorf("获取角色列表失败: %w", err)
	}

	users, err := s.repo.FindMatrixUsers(appID)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	permissions, err := s.repo.FindMatrixPermissions(appID)
	if err != nil {
		return nil, fmt.Errorf("获取权限数据失败: %w", err)
	}

	matrix := make(map[uint]map[uint]bool)
	permissionsDetail := make(map[string]map[string]interface{})

	for _, perm := range permissions {
		if matrix[perm.RoleID] == nil {
			matrix[perm.RoleID] = make(map[uint]bool)
		}
		matrix[perm.RoleID][perm.UserID] = true

		key := fmt.Sprintf("%d_%d", perm.RoleID, perm.UserID)
		permissionsDetail[key] = map[string]interface{}{
			"status":            perm.Status,
			"group_name":        perm.GroupName,
			"assigned_at":       perm.AssignedAt.Format(time.RFC3339),
			"external_username": perm.ExternalUsername,
		}
	}

	result := map[string]interface{}{
		"view_type":          "role-user-matrix",
		"app_id":             appID,
		"app_name":           appName,
		"app_type":           appType,
		"roles":              roles,
		"users":              users,
		"matrix":             matrix,
		"permissions_detail": permissionsDetail,
	}

	return result, nil
}

// getUserRuleMatrix 获取用户-授权规则矩阵（JumpServer）
func (s *ApplicationPermissionService) getUserRuleMatrix(appID uint, appName string) (map[string]interface{}, error) {
	// 1. 获取所有授权规则（从 auth_authorization_rules）
	rules, err := s.repo.FindAuthRulesByAppIDOrdered(appID)
	if err != nil {
		return nil, fmt.Errorf("获取授权规则列表失败: %w", err)
	}

	// 如果没有授权规则，返回空数据
	if len(rules) == 0 {
		return map[string]interface{}{
			"view_type":          "user-rule-matrix",
			"app_id":             appID,
			"app_name":           appName,
			"rules":              []modelauth.ApplicationAuthorizationRule{},
			"users":              []interface{}{},
			"matrix":             map[string]map[string]bool{},
			"permissions_detail": map[string]map[string]interface{}{},
			"message":            "暂无授权规则数据，请先同步授权规则",
		}, nil
	}

	// 2. 获取所有有权限的用户（通过 auth_user_identity_mappings）
	users, err := s.repo.FindRuleMatrixUsers(appID)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 确保 users 至少为空数组，避免 null
	if users == nil {
		users = []repoauth.RuleMatrixUserResult{}
	}

	// 3. 获取应用配置（用于调用 Jumpserver API）
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return nil, fmt.Errorf("获取应用配置失败: %w", err)
	}

	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return nil, fmt.Errorf("解析认证配置失败: %w", err)
	}

	// 获取 Jumpserver 适配器
	adapter, err := s.adapterFactory.GetAdapter("jumpserver")
	if err != nil {
		return nil, fmt.Errorf("获取 Jumpserver 适配器失败: %w", err)
	}

	jumpserverAdapter, ok := adapter.(*appadapter.JumpserverAdapter)
	if !ok {
		return nil, fmt.Errorf("适配器类型转换失败")
	}

	// 4. 查询每个授权规则的用户列表
	matrix := make(map[string]map[string]bool) // ruleID -> userID -> hasPermission
	permissionsDetail := make(map[string]map[string]interface{})

	// 构建 userID -> externalUserID 的映射
	userExternalIDMap := make(map[uint]string)
	for _, user := range users {
		if user.ExternalUserID != "" {
			userExternalIDMap[user.ID] = user.ExternalUserID
		}
	}

	// 查询每个授权规则的详情
	for _, rule := range rules {
		ruleDetail, err := jumpserverAdapter.GetAuthorizationRuleDetail(app.BaseURL, authConfig, rule.RuleID)
		if err != nil {
			logger.Warn("获取授权规则详情失败",
				zap.String("rule_id", rule.RuleID),
				zap.String("rule_name", rule.RuleName),
				zap.Error(err))
			continue
		}

		logger.Info("Jumpserver 授权规则用户详情",
			zap.String("rule_id", rule.RuleID),
			zap.String("rule_name", rule.RuleName),
			zap.Int("jumpserver_user_count", len(ruleDetail.Users)))

		// 初始化该规则的矩阵
		if matrix[rule.RuleID] == nil {
			matrix[rule.RuleID] = make(map[string]bool)
		}

		// 遍历规则中的用户，标记权限
		for _, ruleUser := range ruleDetail.Users {
			logger.Debug("Jumpserver 规则用户",
				zap.String("rule_id", rule.RuleID),
				zap.String("jumpserver_user_id", ruleUser.ID),
				zap.String("jumpserver_username", ruleUser.Username))

			// 找到对应的授权中心用户
			for _, user := range users {
				logger.Debug("授权中心用户匹配",
					zap.String("rule_id", rule.RuleID),
					zap.String("jumpserver_user_id", ruleUser.ID),
					zap.Uint("auth_user_id", user.ID),
					zap.String("external_user_id", user.ExternalUserID),
					zap.Bool("is_match", user.ExternalUserID == ruleUser.ID))

				if user.ExternalUserID == ruleUser.ID {
					matrix[rule.RuleID][fmt.Sprintf("%d", user.ID)] = true

					logger.Info("权限匹配成功",
						zap.String("rule_id", rule.RuleID),
						zap.String("rule_name", rule.RuleName),
						zap.Uint("auth_user_id", user.ID),
						zap.String("username", user.Username),
						zap.String("external_user_id", user.ExternalUserID),
						zap.String("jumpserver_user_id", ruleUser.ID))

					// 添加权限详情
					key := fmt.Sprintf("%s_%d", rule.RuleID, user.ID)
					permissionsDetail[key] = map[string]interface{}{
						"status":            "active",
						"external_username": user.ExternalUsername,
						"rule_name":         rule.RuleName,
						"assets_count":      len(ruleDetail.Assets),
						"nodes_count":       len(ruleDetail.Nodes),
						"actions":           ruleDetail.Actions,
					}
				}
			}
		}
	}

	// 5. 为每个规则添加统计信息
	rulesWithStats := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		ruleMap := map[string]interface{}{
			"id":           rule.ID,
			"rule_id":      rule.RuleID,
			"rule_name":    rule.RuleName,
			"rule_type":    rule.RuleType,
			"subject_type": rule.SubjectType,
			"subject_name": rule.SubjectName,
			"object_type":  rule.ObjectType,
			"object_name":  rule.ObjectName,
			"is_enabled":   rule.IsEnabled,
			"is_expired":   rule.IsExpired,
		}

		// 统计该规则的用户数量
		userCount := 0
		if matrix[rule.RuleID] != nil {
			userCount = len(matrix[rule.RuleID])
		}
		ruleMap["user_count"] = userCount

		rulesWithStats = append(rulesWithStats, ruleMap)
	}

	// 6. 构建返回数据
	// 确保 users 不是 null
	usersResult := make([]interface{}, 0, len(users))
	for _, user := range users {
		usersResult = append(usersResult, map[string]interface{}{
			"id":                user.ID,
			"username":          user.Username,
			"nickname":          user.Nickname,
			"external_username": user.ExternalUsername,
			"external_user_id":  user.ExternalUserID,
		})
	}

	result := map[string]interface{}{
		"view_type":          "user-rule-matrix",
		"app_id":             appID,
		"app_name":           appName,
		"app_type":           "jumpserver",
		"rules":              rulesWithStats,
		"users":              usersResult,
		"matrix":             matrix,
		"permissions_detail": permissionsDetail,
	}

	return result, nil
}
