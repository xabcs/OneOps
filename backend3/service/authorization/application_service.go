package authorization

import (
	"encoding/json"
	"fmt"
	"time"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// CreateApplication 创建应用
func (s *ApplicationPermissionService) CreateApplication(app *modelauth.Application) error {
	return s.repo.CreateApplication(app)
}

// GetApplications 获取应用列表
func (s *ApplicationPermissionService) GetApplications(page, pageSize int, name string) ([]*modelauth.Application, int64, error) {
	return s.repo.FindApplications(page, pageSize, name)
}

// GetApplicationByID 根据ID获取应用
func (s *ApplicationPermissionService) GetApplicationByID(id uint) (*modelauth.Application, error) {
	return s.repo.FindApplicationByID(id)
}

// UpdateApplication 更新应用
func (s *ApplicationPermissionService) UpdateApplication(app *modelauth.Application) error {
	return s.repo.SaveApplication(app)
}

// DeleteApplication 删除应用
func (s *ApplicationPermissionService) DeleteApplication(id uint) error {
	return s.repo.DeleteApplication(id)
}

// GetSupportedAppTypes 获取支持的应用类型列表
func (s *ApplicationPermissionService) GetSupportedAppTypes() []map[string]interface{} {
	types := s.adapterFactory.GetSupportedTypes()
	result := make([]map[string]interface{}, 0, len(types))

	for _, appType := range types {
		adapter, _ := s.adapterFactory.GetAdapter(appType)
		result = append(result, map[string]interface{}{
			"type":        appType,
			"displayName": adapter.GetDisplayName(),
		})
	}

	return result
}

// GetAppTypeConfigTemplate 获取应用类型的配置模板
func (s *ApplicationPermissionService) GetAppTypeConfigTemplate(appType string) (map[string]interface{}, error) {
	adapter, err := s.adapterFactory.GetAdapter(appType)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"appType":        appType,
		"displayName":    adapter.GetDisplayName(),
		"configTemplate": adapter.GetConfigTemplate(),
	}, nil
}

// SyncRoles 同步应用角色（使用适配器模式）
func (s *ApplicationPermissionService) SyncRoles(appID uint, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	// 解析端点配置并合并到 authConfig 中
	if app.Endpoints != "" && app.Endpoints != "null" {
		var endpoints map[string]interface{}
		if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err == nil {
			authConfig["endpoints"] = endpoints
			logger.Info("从数据库加载端点配置",
				zap.Uint("app_id", appID),
				zap.Any("endpoints", endpoints))
		} else {
			logger.Warn("解析端点配置失败，将使用适配器默认端点",
				zap.Uint("app_id", appID),
				zap.String("endpoints", app.Endpoints),
				zap.Error(err))
		}
	} else {
		logger.Info("数据库中无端点配置，将使用适配器默认端点",
			zap.Uint("app_id", appID),
			zap.String("endpoints", app.Endpoints))
	}

	// 获取对应类型的适配器
	adapter, err := s.adapterFactory.GetAdapter(app.Type)
	if err != nil {
		errorMsg := fmt.Sprintf("获取应用适配器失败: %v", err)
		s.logOperation(appID, "sync_roles", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("%v (应用: %s, 类型: %s)", err, app.Name, app.Type)
	}

	// 验证配置
	if err := adapter.ValidateConfig(authConfig); err != nil {
		errorMsg := fmt.Sprintf("配置验证失败: %v", err)
		s.logOperation(appID, "sync_roles", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 使用适配器获取角色
	roles, err := adapter.FetchRoles(app.BaseURL, authConfig)
	if err != nil {
		errorMsg := fmt.Sprintf("同步角色失败: %v", err)
		s.logOperation(appID, "sync_roles", "", "", "failed", errorMsg, operator)
		return err
	}

	// 删除旧的同步数据
	s.repo.DeleteRolesByAppID(appID)

	// 保存新的角色数据
	for _, role := range roles {
		role.AppID = appID
		role.SyncTime = time.Now()
		s.repo.CreateRole(&role)
	}

	// 更新应用最后同步时间
	now := time.Now()
	app.LastSyncTime = &now
	s.repo.SaveApplication(app)

	s.logOperation(appID, "sync_roles", "", fmt.Sprintf("同步了 %d 个角色", len(roles)), "success", "", operator)

	logger.Info("同步应用角色成功",
		zap.Uint("app_id", appID),
		zap.String("app_name", app.Name),
		zap.String("app_type", app.Type),
		zap.Int("role_count", len(roles)))

	return nil
}

// SyncUsers 同步应用用户（使用适配器模式）
func (s *ApplicationPermissionService) SyncUsers(appID uint, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	// 解析端点配置并合并到 authConfig 中
	if app.Endpoints != "" && app.Endpoints != "null" {
		var endpoints map[string]interface{}
		if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err == nil {
			authConfig["endpoints"] = endpoints
			logger.Info("从数据库加载端点配置",
				zap.Uint("app_id", appID),
				zap.Any("endpoints", endpoints))
		} else {
			logger.Warn("解析端点配置失败，将使用适配器默认端点",
				zap.Uint("app_id", appID),
				zap.String("endpoints", app.Endpoints),
				zap.Error(err))
		}
	} else {
		logger.Info("数据库中无端点配置，将使用适配器默认端点",
			zap.Uint("app_id", appID),
			zap.String("endpoints", app.Endpoints))
	}

	// 获取对应类型的适配器
	adapter, err := s.adapterFactory.GetAdapter(app.Type)
	if err != nil {
		errorMsg := fmt.Sprintf("获取应用适配器失败: %v", err)
		s.logOperation(appID, "sync_users", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("%v (应用: %s, 类型: %s)", err, app.Name, app.Type)
	}

	// 验证配置
	if err := adapter.ValidateConfig(authConfig); err != nil {
		errorMsg := fmt.Sprintf("配置验证失败: %v", err)
		s.logOperation(appID, "sync_users", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 使用适配器获取用户
	users, err := adapter.FetchUsers(app.BaseURL, authConfig)
	if err != nil {
		errorMsg := fmt.Sprintf("同步用户失败: %v", err)
		s.logOperation(appID, "sync_users", "", "", "failed", errorMsg, operator)
		return err
	}

	logger.Info("从外部应用获取到用户数据",
		zap.Uint("app_id", appID),
		zap.String("app_name", app.Name),
		zap.Int("user_count", len(users)))

	// 删除旧的同步数据
	deletedCount, err := s.repo.DeleteUsersByAppID(appID)
	if err != nil {
		logger.Error("删除旧用户数据失败",
			zap.Uint("app_id", appID),
			zap.Error(err))
		return fmt.Errorf("删除旧用户数据失败: %w", err)
	}

	logger.Info("已删除旧用户数据",
		zap.Uint("app_id", appID),
		zap.Int64("deleted_count", deletedCount))

	// 保存新的用户数据
	successCount := 0
	for _, user := range users {
		user.AppID = appID
		user.SyncTime = time.Now()
		if err := s.repo.CreateUser(&user); err != nil {
			logger.Error("保存用户数据失败",
				zap.Uint("app_id", appID),
				zap.String("username", user.Username),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	s.logOperation(appID, "sync_users", "", fmt.Sprintf("同步了 %d 个用户（成功保存 %d 个）", len(users), successCount), "success", "", operator)

	logger.Info("同步应用用户完成",
		zap.Uint("app_id", appID),
		zap.String("app_name", app.Name),
		zap.String("app_type", app.Type),
		zap.Int("total_fetched", len(users)),
		zap.Int("success_saved", successCount))

	return nil
}

// SyncGroups 同步应用用户组（使用适配器模式）
func (s *ApplicationPermissionService) SyncGroups(appID uint, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	if app.Endpoints != "" && app.Endpoints != "null" {
		var endpoints map[string]interface{}
		if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err == nil {
			authConfig["endpoints"] = endpoints
		}
	}

	adapter, err := s.adapterFactory.GetAdapter(app.Type)
	if err != nil {
		errorMsg := fmt.Sprintf("获取应用适配器失败: %v", err)
		s.logOperation(appID, "sync_groups", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("%v (应用: %s, 类型: %s)", err, app.Name, app.Type)
	}

	if err := adapter.ValidateConfig(authConfig); err != nil {
		errorMsg := fmt.Sprintf("配置验证失败: %v", err)
		s.logOperation(appID, "sync_groups", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("配置验证失败: %w", err)
	}

	groups, err := adapter.FetchGroups(app.BaseURL, authConfig)
	if err != nil {
		errorMsg := fmt.Sprintf("同步用户组失败: %v", err)
		s.logOperation(appID, "sync_groups", "", "", "failed", errorMsg, operator)
		return err
	}

	s.repo.DeleteGroupsByAppID(appID)

	for _, group := range groups {
		group.AppID = appID
		group.SyncTime = time.Now()
		if err := s.repo.CreateGroup(&group); err != nil {
			logger.Error("保存用户组数据失败",
				zap.String("group_code", group.GroupCode),
				zap.Error(err))
		}
	}

	now := time.Now()
	app.LastSyncTime = &now
	s.repo.SaveApplication(app)

	s.logOperation(appID, "sync_groups", "", "", "success",
		fmt.Sprintf("成功同步 %d 个用户组", len(groups)), operator)

	return nil
}

// GetApplicationGroups 获取应用的用户组列表
func (s *ApplicationPermissionService) GetApplicationGroups(appID uint) ([]modelauth.ApplicationGroup, error) {
	groups, err := s.repo.FindGroupsByAppID(appID)
	if err != nil {
		return nil, fmt.Errorf("获取用户组列表失败: %w", err)
	}
	return groups, nil
}

// SyncAuthorizationRules 同步应用授权规则
func (s *ApplicationPermissionService) SyncAuthorizationRules(appID uint, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	if app.Endpoints != "" && app.Endpoints != "null" {
		var endpoints map[string]interface{}
		if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err == nil {
			authConfig["endpoints"] = endpoints
		}
	}

	adapter, err := s.adapterFactory.GetAdapter(app.Type)
	if err != nil {
		errorMsg := fmt.Sprintf("获取应用适配器失败: %v", err)
		s.logOperation(appID, "sync_rules", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("%v (应用: %s, 类型: %s)", err, app.Name, app.Type)
	}

	rulesAdapter, ok := adapter.(interface {
		FetchAuthorizationRules(baseURL string, authConfig map[string]interface{}) ([]modelauth.ApplicationAuthorizationRule, error)
	})
	if !ok {
		warningMsg := fmt.Sprintf("应用类型 %s 不支持授权规则同步", app.Type)
		s.logOperation(appID, "sync_rules", "", "", "failed", warningMsg, operator)
		return fmt.Errorf("%s", warningMsg)
	}

	if err := adapter.ValidateConfig(authConfig); err != nil {
		errorMsg := fmt.Sprintf("配置验证失败: %v", err)
		s.logOperation(appID, "sync_rules", "", "", "failed", errorMsg, operator)
		return fmt.Errorf("配置验证失败: %w", err)
	}

	rules, err := rulesAdapter.FetchAuthorizationRules(app.BaseURL, authConfig)
	if err != nil {
		errorMsg := fmt.Sprintf("同步授权规则失败: %v", err)
		s.logOperation(appID, "sync_rules", "", "", "failed", errorMsg, operator)
		return err
	}

	if _, err := s.repo.DeleteAuthRulesByAppID(appID); err != nil {
		return fmt.Errorf("删除旧授权规则数据失败: %w", err)
	}

	successCount := 0
	for _, rule := range rules {
		rule.AppID = appID
		rule.SyncTime = time.Now()
		if err := s.repo.CreateAuthRule(&rule); err != nil {
			logger.Error("保存授权规则数据失败",
				zap.Uint("app_id", appID),
				zap.String("rule_id", rule.RuleID),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	now := time.Now()
	app.LastSyncTime = &now
	s.repo.SaveApplication(app)

	s.logOperation(appID, "sync_rules", "", "", "success",
		fmt.Sprintf("成功同步 %d 条授权规则（成功保存 %d 条）", len(rules), successCount), operator)

	return nil
}

// GetApplicationAuthorizationRules 获取应用的授权规则列表
func (s *ApplicationPermissionService) GetApplicationAuthorizationRules(appID uint) ([]modelauth.ApplicationAuthorizationRule, error) {
	rules, err := s.repo.FindAuthRulesByAppID(appID)
	if err != nil {
		return nil, fmt.Errorf("获取授权规则列表失败: %w", err)
	}
	return rules, nil
}
