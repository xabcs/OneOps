package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/logger"
	repoauth "oneops/backend3/repository/authorization"
	appadapter "oneops/backend3/service/authorization/adapter"

	"go.uber.org/zap"
)

// ApplicationPermissionService 应用权限服务
type ApplicationPermissionService struct {
	repo           *repoauth.AuthorizationRepository
	adapterFactory *AdapterFactory
}

// NewApplicationPermissionService 创建服务实例
func NewApplicationPermissionService(repo *repoauth.AuthorizationRepository) *ApplicationPermissionService {
	return &ApplicationPermissionService{
		repo:           repo,
		adapterFactory: NewAdapterFactory(),
	}
}

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
				zap.Uint("appID", appID),
				zap.Any("endpoints", endpoints))
		} else {
			logger.Warn("解析端点配置失败，将使用适配器默认端点",
				zap.Uint("appID", appID),
				zap.String("endpoints", app.Endpoints),
				zap.Error(err))
		}
	} else {
		logger.Info("数据库中无端点配置，将使用适配器默认端点",
			zap.Uint("appID", appID),
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
		zap.Uint("appID", appID),
		zap.String("appName", app.Name),
		zap.String("appType", app.Type),
		zap.Int("roleCount", len(roles)))

	return nil
}

// fetchRolesFromAPI 从外部 API 获取角色
func (s *ApplicationPermissionService) fetchRolesFromAPI(baseURL string, endpoints map[string]string, authConfig map[string]interface{}) ([]modelauth.ApplicationRole, error) {
	// 解析 getRoles 端点
	getRolesURL, ok := endpoints["getRoles"]
	if !ok {
		return nil, fmt.Errorf("缺少 getRoles 配置，请在应用配置中添加 getRoles 端点")
	}

	// 构建完整 URL（处理 BaseURL 末尾斜杠）
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + getRolesURL

	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加认证头
	if token, ok := authConfig["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 支持 Basic Auth 认证
	if authConfig["type"] == "basic" {
		username, _ := authConfig["username"].(string)
		password, _ := authConfig["password"].(string)
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w (URL: %s)", err, fullURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("认证失败(403)，请检查 Token 是否正确或是否有权限访问 API: %s (响应: %s)", fullURL, string(body))
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("未授权(401)，请检查认证配置是否正确")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误: %d (响应: %s)", resp.StatusCode, string(body))
	}

	// 解析响应
	var apiResp struct {
		Code    int                      `json:"code"`
		Data    []map[string]interface{} `json:"data"`
		Message string                   `json:"message"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为 ApplicationRole
	roles := make([]modelauth.ApplicationRole, 0)
	for _, item := range apiResp.Data {
		role := modelauth.ApplicationRole{
			RoleCode:    s.getString(item, "code"),
			RoleName:    s.getString(item, "name"),
			RoleType:    s.getString(item, "type"),
			Description: s.getString(item, "description"),
		}
		roles = append(roles, role)
	}

	return roles, nil
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
				zap.Uint("appID", appID),
				zap.Any("endpoints", endpoints))
		} else {
			logger.Warn("解析端点配置失败，将使用适配器默认端点",
				zap.Uint("appID", appID),
				zap.String("endpoints", app.Endpoints),
				zap.Error(err))
		}
	} else {
		logger.Info("数据库中无端点配置，将使用适配器默认端点",
			zap.Uint("appID", appID),
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
		zap.Uint("appID", appID),
		zap.String("appName", app.Name),
		zap.Int("userCount", len(users)))

	// 删除旧的同步数据
	deletedCount, err := s.repo.DeleteUsersByAppID(appID)
	if err != nil {
		logger.Error("删除旧用户数据失败",
			zap.Uint("appID", appID),
			zap.Error(err))
		return fmt.Errorf("删除旧用户数据失败: %w", err)
	}

	logger.Info("已删除旧用户数据",
		zap.Uint("appID", appID),
		zap.Int64("deletedCount", deletedCount))

	// 保存新的用户数据
	successCount := 0
	for _, user := range users {
		user.AppID = appID
		user.SyncTime = time.Now()
		if err := s.repo.CreateUser(&user); err != nil {
			logger.Error("保存用户数据失败",
				zap.Uint("appID", appID),
				zap.String("username", user.Username),
				zap.Error(err))
		} else {
			successCount++
		}
	}

	s.logOperation(appID, "sync_users", "", fmt.Sprintf("同步了 %d 个用户（成功保存 %d 个）", len(users), successCount), "success", "", operator)

	logger.Info("同步应用用户完成",
		zap.Uint("appID", appID),
		zap.String("appName", app.Name),
		zap.String("appType", app.Type),
		zap.Int("totalFetched", len(users)),
		zap.Int("successSaved", successCount))

	return nil
}

// fetchUsersFromAPI 从外部 API 获取用户
func (s *ApplicationPermissionService) fetchUsersFromAPI(baseURL string, endpoints map[string]string, authConfig map[string]interface{}) ([]modelauth.ApplicationUser, error) {
	listUsersURL, ok := endpoints["listUsers"]
	if !ok || listUsersURL == "" {
		return nil, fmt.Errorf("缺少 listUsers 配置，请在应用配置中添加 listUsers 端点")
	}

	// 构建完整 URL（处理 BaseURL 末尾斜杠）
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + listUsersURL

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加认证头
	if token, ok := authConfig["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 支持 Basic Auth 认证
	if authConfig["type"] == "basic" {
		username, _ := authConfig["username"].(string)
		password, _ := authConfig["password"].(string)
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w (URL: %s)", err, fullURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("认证失败(403)，请检查 Token 是否正确或是否有权限访问 API: %s (响应: %s)", fullURL, string(body))
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("未授权(401)，请检查认证配置是否正确")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误: %d (响应: %s)", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Code    int                      `json:"code"`
		Data    []map[string]interface{} `json:"data"`
		Message string                   `json:"message"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	users := make([]modelauth.ApplicationUser, 0)
	for _, item := range apiResp.Data {
		user := modelauth.ApplicationUser{
			Username:    s.getString(item, "username"),
			DisplayName: s.getString(item, "displayName"),
			Email:       s.getString(item, "email"),
			Status:      s.getString(item, "status"),
		}
		users = append(users, user)
	}

	return users, nil
}

// ExternalUserResult 外部用户创建结果
type ExternalUserResult struct {
	AppName     string `json:"appName"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	RoleCode    string `json:"roleCode"`
	RoleName    string `json:"roleName"`
	CreateError string `json:"createError"`
	GrantError  string `json:"grantError"`
	Success     bool   `json:"success"`
}

// createJenkinsSession 创建 Jenkins 会话（带 Cookie 支持）
func (s *ApplicationPermissionService) createJenkinsSession(app *modelauth.Application, authConfig map[string]interface{}) (*JenkinsSession, error) {
	// 创建带 Cookie Jar 的 HTTP Client
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建 cookie jar 失败: %w", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	session := &JenkinsSession{
		Client:  client,
		BaseURL: strings.TrimSuffix(app.BaseURL, "/"),
	}

	// 获取 crumb
	crumbURL := session.BaseURL + "/crumbIssuer/api/json"
	logger.Info("尝试获取 Jenkins crumb", zap.String("crumbURL", crumbURL))

	req, err := http.NewRequest("GET", crumbURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 crumb 请求失败: %w", err)
	}

	// 设置认证
	if authConfig["type"] == "basic" {
		username, _ := authConfig["username"].(string)
		password, _ := authConfig["password"].(string)
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取 crumb 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Info("Jenkins crumb 响应", zap.Int("statusCode", resp.StatusCode), zap.String("body", string(body)))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 crumb 失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var crumbResponse struct {
		CrumbRequestField string `json:"crumbRequestField"`
		Crumb             string `json:"crumb"`
	}
	if err := json.Unmarshal(body, &crumbResponse); err != nil {
		logger.Warn("解析 Jenkins crumb JSON 失败", zap.Error(err), zap.String("body", string(body)))
		return nil, fmt.Errorf("解析 crumb 响应失败: %w", err)
	}

	if crumbResponse.CrumbRequestField == "" || crumbResponse.Crumb == "" {
		logger.Warn("Jenkins crumb 字段为空，可能 CSRF 保护已禁用",
			zap.String("crumbRequestField", crumbResponse.CrumbRequestField),
			zap.String("crumb", crumbResponse.Crumb))
		session.Crumb = nil
	} else {
		session.Crumb = map[string]string{
			crumbResponse.CrumbRequestField: crumbResponse.Crumb,
		}
		logger.Info("成功获取 Jenkins crumb",
			zap.String("crumbField", crumbResponse.CrumbRequestField),
			zap.String("crumbValue", crumbResponse.Crumb))
	}

	return session, nil
}

// getJenkinsCrumb 获取 Jenkins crumb token (已废弃，使用 createJenkinsSession 替代)
func (s *ApplicationPermissionService) getJenkinsCrumb(app *modelauth.Application, authConfig map[string]interface{}) (map[string]string, error) {
	baseURL := strings.TrimSuffix(app.BaseURL, "/")
	crumbURL := baseURL + "/crumbIssuer/api/json"

	logger.Info("尝试获取 Jenkins crumb", zap.String("crumbURL", crumbURL))

	req, err := http.NewRequest("GET", crumbURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 crumb 请求失败: %w", err)
	}

	if authConfig["type"] == "basic" {
		username, _ := authConfig["username"].(string)
		password, _ := authConfig["password"].(string)
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取 crumb 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Info("Jenkins crumb 响应", zap.Int("statusCode", resp.StatusCode), zap.String("body", string(body)))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 crumb 失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	var crumbResponse struct {
		CrumbRequestField string `json:"crumbRequestField"`
		Crumb             string `json:"crumb"`
	}
	if err := json.Unmarshal(body, &crumbResponse); err != nil {
		logger.Warn("解析 Jenkins crumb JSON 失败", zap.Error(err), zap.String("body", string(body)))
		return nil, fmt.Errorf("解析 crumb 响应失败: %w", err)
	}

	if crumbResponse.CrumbRequestField == "" || crumbResponse.Crumb == "" {
		return nil, fmt.Errorf("crumb 字段为空，可能 CSRF 保护已禁用")
	}

	result := map[string]string{
		crumbResponse.CrumbRequestField: crumbResponse.Crumb,
	}
	return result, nil
}

// CreateExternalUser 在外部系统创建用户（使用适配器模式）
func (s *ApplicationPermissionService) CreateExternalUser(appID uint, authUserID uint, username, password, email, fullName string, operator string) error {
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
		} else {
			logger.Warn("解析端点配置失败，将使用适配器默认端点",
				zap.Uint("appID", appID),
				zap.Error(err))
		}
	}

	// 获取对应类型的适配器
	adapter, err := s.adapterFactory.GetAdapter(app.Type)
	if err != nil {
		return fmt.Errorf("获取应用适配器失败: %w", err)
	}

	// 验证配置
	if err := adapter.ValidateConfig(authConfig); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 构建用户创建请求
	displayName := fullName
	if displayName == "" {
		displayName = username
	}

	userEmail := email
	if userEmail == "" {
		userEmail = fmt.Sprintf("%s@oneops.local", username)
	}

	userReq := &UserCreateRequest{
		Username:    username,
		Password:    password,
		FullName:    displayName,
		Email:       userEmail,
		Description: "",
	}

	// 使用适配器创建用户，获取返回的用户ID
	externalUserID, err := adapter.CreateUser(app.BaseURL, authConfig, userReq)
	if err != nil {
		s.logOperation(appID, "create_user", username, "", "failed", err.Error(), operator)
		return err
	}

	s.logOperation(appID, "create_user", username, "", "success", "", operator)
	logger.Info("在外部系统创建用户成功",
		zap.Uint("appID", appID),
		zap.String("appName", app.Name),
		zap.String("appType", app.Type),
		zap.String("username", username),
		zap.String("externalUserID", externalUserID))

	// 记录用户身份映射
	now := time.Now()
	mapping := modelauth.UserIdentityMapping{
		AuthUserID:       authUserID,
		AppID:            appID,
		ExternalUsername: username,
		ExternalUserID:   externalUserID,
		MappingType:      "auto",
		MappingStatus:    "active",
		LastSyncAt:       &now,
	}

	if err := s.repo.FirstOrCreateUserIdentityMapping(&mapping, authUserID, appID); err != nil {
		logger.Warn("记录用户身份映射失败",
			zap.Uint("authUserID", authUserID),
			zap.Uint("appID", appID),
			zap.String("username", username),
			zap.Error(err))
	} else {
		logger.Info("记录用户身份映射成功",
			zap.Uint("authUserID", authUserID),
			zap.Uint("appID", appID),
			zap.String("externalUsername", username),
			zap.String("externalUserID", externalUserID))
	}

	return nil
}

// createJenkinsUser 在 Jenkins 中创建用户（使用会话保持 Cookie）
func (s *ApplicationPermissionService) createJenkinsUser(app *modelauth.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, operator string) error {
	session, err := s.createJenkinsSession(app, authConfig)
	if err != nil {
		logger.Warn("创建 Jenkins 会话失败，尝试不使用会话进行请求", zap.Error(err))
		return s.createOtherSystemUser(app, authConfig, endpoints, jsonData, username, operator)
	}

	var jsonParams map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonParams); err != nil {
		return fmt.Errorf("解析用户数据失败: %w", err)
	}

	formData := url.Values{}
	formData.Set("username", username)
	if pwd, ok := jsonParams["password"].(string); ok {
		formData.Set("password1", pwd)
		formData.Set("password2", pwd)
	}

	email := ""
	if e, ok := jsonParams["email"].(string); ok && e != "" {
		if strings.Contains(e, "@") && strings.Contains(e, ".") {
			email = e
		} else {
			email = username + "@example.com"
		}
	} else {
		email = username + "@example.com"
	}
	formData.Set("email", email)

	fullName := ""
	if fn, ok := jsonParams["fullName"].(string); ok && fn != "" {
		fullName = fn
	} else if dn, ok := jsonParams["displayName"].(string); ok && dn != "" {
		fullName = dn
	}
	if fullName == "" {
		fullName = username
	}
	formData.Set("fullname", fullName)

	if session.Crumb != nil {
		for key, value := range session.Crumb {
			formData.Set(key, value)
		}
	}

	createUserURL := session.BaseURL + "/securityRealm/createAccountByAdmin"
	req, err := http.NewRequest("POST", createUserURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if authConfig["type"] == "basic" {
		authUsername, _ := authConfig["username"].(string)
		authPassword, _ := authConfig["password"].(string)
		if authUsername != "" && authPassword != "" {
			req.SetBasicAuth(authUsername, authPassword)
		}
	}

	resp, err := session.Client.Do(req)
	if err != nil {
		s.logOperation(app.ID, "create_user", username, formData.Encode(), "failed", err.Error(), operator)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if len(bodyStr) > 500 {
		bodyStr = bodyStr[:500]
	}

	if strings.Contains(strings.ToLower(bodyStr), "error") || strings.Contains(bodyStr, "Oops") {
		s.logOperation(app.ID, "create_user", username, formData.Encode(), "failed", bodyStr, operator)
		return fmt.Errorf("创建用户失败，Jenkins 返回错误信息")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusFound {
		s.logOperation(app.ID, "create_user", username, formData.Encode(), "failed", bodyStr, operator)
		return fmt.Errorf("创建用户失败 (状态码 %d)", resp.StatusCode)
	}

	s.logOperation(app.ID, "create_user", username, formData.Encode(), "success", "", operator)
	return nil
}

// createOtherSystemUser 在其他外部系统创建用户
func (s *ApplicationPermissionService) createOtherSystemUser(app *modelauth.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, operator string) error {
	baseURL := strings.TrimSuffix(app.BaseURL, "/")
	createUserURL := baseURL + endpoints["createUser"]
	req, err := http.NewRequest("POST", createUserURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token, ok := authConfig["token"].(string); ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if authConfig["type"] == "basic" {
		authUsername, _ := authConfig["username"].(string)
		authPassword, _ := authConfig["password"].(string)
		if authUsername != "" && authPassword != "" {
			req.SetBasicAuth(authUsername, authPassword)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logOperation(app.ID, "create_user", username, string(jsonData), "failed", err.Error(), operator)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logOperation(app.ID, "create_user", username, string(jsonData), "failed", string(body), operator)
		return fmt.Errorf("创建用户失败: %s", string(body))
	}

	s.logOperation(app.ID, "create_user", username, string(jsonData), "success", "", operator)
	return nil
}

// GrantRoleToUser 为用户授予应用角色（使用适配器模式）
func (s *ApplicationPermissionService) GrantRoleToUser(appID uint, authUserID uint, groupBindingID uint, username, roleCode string, operator string) error {
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
		return fmt.Errorf("获取应用适配器失败: %w", err)
	}

	if err := adapter.ValidateConfig(authConfig); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	if err := adapter.AssignRole(app.BaseURL, authConfig, username, roleCode); err != nil {
		s.logOperation(appID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), "", "failed", err.Error(), operator)
		s.recordGroupBindingExecution(groupBindingID, authUserID, username, "granted", "failed", err.Error(), operator)
		return err
	}

	s.logOperation(appID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), "", "success", "", operator)

	s.recordGroupBindingExecution(groupBindingID, authUserID, username, "granted", "success", "角色分配成功", operator)

	return nil
}

// grantJenkinsRole 在 Jenkins 中授予角色（使用会话保持 Cookie）
func (s *ApplicationPermissionService) grantJenkinsRole(app *modelauth.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, roleCode, operator string) error {
	session, err := s.createJenkinsSession(app, authConfig)
	if err != nil {
		return fmt.Errorf("创建 Jenkins 会话失败: %w", err)
	}

	formData := url.Values{}
	roleType := "globalRoles"
	if strings.Contains(strings.ToLower(roleCode), "item") ||
		strings.Contains(strings.ToLower(roleCode), "project") ||
		strings.Contains(roleCode, "-") ||
		strings.Contains(roleCode, "_") {
		roleType = "projectRoles"
	}

	formData.Set("type", roleType)
	formData.Set("roleName", roleCode)
	formData.Set("sid", username)

	if session.Crumb != nil {
		for key, value := range session.Crumb {
			formData.Set(key, value)
		}
	}

	assignRoleURL := session.BaseURL + "/role-strategy/strategy/assignRole"
	req, err := http.NewRequest("POST", assignRoleURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if authConfig["type"] == "basic" {
		authUsername, _ := authConfig["username"].(string)
		authPassword, _ := authConfig["password"].(string)
		if authUsername != "" && authPassword != "" {
			req.SetBasicAuth(authUsername, authPassword)
		}
	}

	resp, err := session.Client.Do(req)
	if err != nil {
		s.logOperation(app.ID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), formData.Encode(), "failed", err.Error(), operator)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
		s.logOperation(app.ID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), formData.Encode(), "success", "", operator)
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	errorMsg := string(body)
	s.logOperation(app.ID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), formData.Encode(), "failed", errorMsg, operator)
	return fmt.Errorf("授权失败 (状态码 %d): %s", resp.StatusCode, errorMsg)
}

// grantOtherSystemRole 在其他外部系统授予角色
func (s *ApplicationPermissionService) grantOtherSystemRole(app *modelauth.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, roleCode, operator string) error {
	baseURL := strings.TrimSuffix(app.BaseURL, "/")
	assignRoleURL := baseURL + endpoints["assignRole"]
	req, err := http.NewRequest("POST", assignRoleURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token, ok := authConfig["token"].(string); ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if authConfig["type"] == "basic" {
		authUsername, _ := authConfig["username"].(string)
		authPassword, _ := authConfig["password"].(string)
		if authUsername != "" && authPassword != "" {
			req.SetBasicAuth(authUsername, authPassword)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logOperation(app.ID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), string(jsonData), "failed", err.Error(), operator)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logOperation(app.ID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), string(jsonData), "failed", string(body), operator)
		return fmt.Errorf("授权失败: %s", string(body))
	}

	s.logOperation(app.ID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), string(jsonData), "success", "", operator)
	return nil
}

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

// ========== 授权中心用户管理 ==========

// CreateAuthUser 创建授权中心用户，返回生成的密码
func (s *ApplicationPermissionService) CreateAuthUser(user *modelauth.AuthUser) (string, error) {
	password, err := GenerateRandomPassword(DefaultPasswordConfig)
	if err != nil {
		return "", fmt.Errorf("生成密码失败: %w", err)
	}

	user.Password = password

	if err := s.repo.CreateAuthUser(user); err != nil {
		return "", err
	}

	return password, nil
}

// GetAuthUsers 获取授权中心用户列表
func (s *ApplicationPermissionService) GetAuthUsers(page, pageSize int, username string) ([]*modelauth.AuthUser, int64, error) {
	return s.repo.FindAuthUsers(page, pageSize, username)
}

// GetAuthUserByID 根据ID获取授权中心用户
func (s *ApplicationPermissionService) GetAuthUserByID(id uint) (*modelauth.AuthUser, error) {
	return s.repo.FindAuthUserByID(id)
}

// UpdateAuthUser 更新授权中心用户
func (s *ApplicationPermissionService) UpdateAuthUser(user *modelauth.AuthUser) error {
	return s.repo.SaveAuthUser(user)
}

// DeleteAuthUser 删除授权中心用户
func (s *ApplicationPermissionService) DeleteAuthUser(id uint) error {
	return s.repo.DeleteAuthUser(id)
}

// GetAllAuthUsers 获取所有授权中心用户
func (s *ApplicationPermissionService) GetAllAuthUsers() ([]*modelauth.AuthUser, error) {
	return s.repo.FindAllActiveAuthUsers()
}

// ========== 授权中心角色管理 ==========

// CreateAuthGroup 创建授权中心用户组
func (s *ApplicationPermissionService) CreateAuthGroup(group *modelauth.AuthGroup) error {
	return s.repo.CreateAuthGroup(group)
}

// GetAuthGroups 获取授权中心用户组列表
func (s *ApplicationPermissionService) GetAuthGroups(page, pageSize int, name string) ([]*modelauth.AuthGroup, int64, error) {
	return s.repo.FindAuthGroups(page, pageSize, name)
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
			zap.Uint("groupBindingID", groupBindingID),
			zap.Uint("authUserID", authUserID),
			zap.Error(err))
	}
}

// GetAuthGroupByID 根据ID获取授权中心用户组
func (s *ApplicationPermissionService) GetAuthGroupByID(id uint) (*modelauth.AuthGroup, error) {
	return s.repo.FindAuthGroupByID(id)
}

// UpdateAuthGroup 更新授权中心用户组
func (s *ApplicationPermissionService) UpdateAuthGroup(group *modelauth.AuthGroup) error {
	return s.repo.SaveAuthGroup(group)
}

// DeleteAuthGroup 删除授权中心用户组
func (s *ApplicationPermissionService) DeleteAuthGroup(id uint) error {
	return s.repo.DeleteAuthGroup(id)
}

// GetAllAuthGroups 获取所有授权中心用户组
func (s *ApplicationPermissionService) GetAllAuthGroups() ([]*modelauth.AuthGroup, error) {
	return s.repo.FindAllActiveAuthGroups()
}

// ========== Jenkins Script Console API 支持 ==========

// executeJenkinsScript 执行 Jenkins Groovy 脚本
func (s *ApplicationPermissionService) executeJenkinsScript(baseURL, username, password, script string) (string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("创建 Cookie Jar 失败: %w", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	crumbURL := strings.TrimSuffix(baseURL, "/") + "/crumbIssuer/api/json"
	crumbReq, err := http.NewRequest("GET", crumbURL, nil)
	if err != nil {
		return "", err
	}
	crumbReq.SetBasicAuth(username, password)

	crumbResp, err := client.Do(crumbReq)
	if err != nil {
		return "", fmt.Errorf("获取 CSRF Token 失败: %w", err)
	}

	if crumbResp.StatusCode != http.StatusOK {
		crumbResp.Body.Close()
		return "", fmt.Errorf("获取 CSRF Token 失败，状态码: %d", crumbResp.StatusCode)
	}

	var crumb JenkinsCrumb
	crumbBody, _ := io.ReadAll(crumbResp.Body)
	crumbResp.Body.Close()

	if err := json.Unmarshal(crumbBody, &crumb); err != nil {
		return "", fmt.Errorf("解析 CSRF Token 失败: %w", err)
	}

	scriptURL := strings.TrimSuffix(baseURL, "/") + "/scriptText"
	data := "script=" + script
	req, err := http.NewRequest("POST", scriptURL, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(crumb.CrumbRequestField, crumb.Crumb)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("执行脚本失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("执行脚本失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

// fetchJenkinsUsers 通过 Script Console 获取 Jenkins 用户列表
func (s *ApplicationPermissionService) fetchJenkinsUsers(baseURL string, authConfig map[string]interface{}) ([]modelauth.ApplicationUser, error) {
	username, _ := authConfig["username"].(string)
	password, _ := authConfig["password"].(string)

	if username == "" || password == "" {
		return nil, fmt.Errorf("Jenkins 需要 Basic Auth 认证，请配置用户名和密码")
	}

	script := `
import jenkins.model.*
import hudson.tasks.Mailer

def users = Jenkins.instance.securityRealm.getAllUsers()
users.each { user ->
  def email = user.getProperty(Mailer.UserProperty)?.address ?: ""
  println "${user.id}|${user.fullName}|${email}|active"
}
`

	result, err := s.executeJenkinsScript(baseURL, username, password, script)
	if err != nil {
		return nil, err
	}

	users := make([]modelauth.ApplicationUser, 0)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Result:") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			user := modelauth.ApplicationUser{
				Username:    parts[0],
				DisplayName: parts[1],
				Status:      "active",
			}
			if len(parts) >= 3 {
				user.Email = parts[2]
			}
			users = append(users, user)
		}
	}

	return users, nil
}

// fetchJenkinsRoles 通过 Script Console 获取 Jenkins 角色列表
func (s *ApplicationPermissionService) fetchJenkinsRoles(baseURL string, authConfig map[string]interface{}) ([]modelauth.ApplicationRole, error) {
	username, _ := authConfig["username"].(string)
	password, _ := authConfig["password"].(string)

	if username == "" || password == "" {
		return nil, fmt.Errorf("Jenkins 需要 Basic Auth 认证，请配置用户名和密码")
	}

	script := `
import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*
import jenkins.model.*

def rbas = Jenkins.instance.getAuthorizationStrategy()
def roles = []

if (rbas instanceof RoleBasedAuthorizationStrategy) {
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
    globalRoles.each { role ->
        println "${role.name}|${role.name}|global|Global Role"
    }

    def projectRoles = rbas.getRoleMap(RoleType.Project).getRoles()
    projectRoles.each { role ->
        println "${role.name}|${role.name}|project|Project Role"
    }
} else {
    println "Warning: Role Strategy plugin not configured"
}
`

	result, err := s.executeJenkinsScript(baseURL, username, password, script)
	if err != nil {
		return nil, err
	}

	roles := make([]modelauth.ApplicationRole, 0)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Warning:") || strings.HasPrefix(line, "Result:") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			role := modelauth.ApplicationRole{
				RoleCode:    parts[0],
				RoleName:    parts[1],
				RoleType:    "global",
				Description: "Jenkins Role",
			}
			if len(parts) >= 3 {
				role.RoleType = parts[2]
			}
			if len(parts) >= 4 {
				role.Description = parts[3]
			}
			roles = append(roles, role)
		}
	}

	if len(roles) == 0 {
		roles = append(roles, modelauth.ApplicationRole{
			RoleCode:    "admin",
			RoleName:    "Administrator",
			RoleType:    "global",
			Description: "Jenkins Administrator (请在 Jenkins 中配置角色)",
		})
		roles = append(roles, modelauth.ApplicationRole{
			RoleCode:    "developer",
			RoleName:    "Developer",
			RoleType:    "global",
			Description: "Jenkins Developer (请在 Jenkins 中配置角色)",
		})
	}

	return roles, nil
}

// syncExistingMembersToExternalSystem 为用户组的现有成员同步外部系统权限
func (s *ApplicationPermissionService) syncExistingMembersToExternalSystem(groupID, appID, roleID, groupBindingID uint, roleCode, roleName, operator string) {
	userGroups, err := s.repo.FindUserGroupsByGroupID(groupID)
	if err != nil {
		logger.Error("获取用户组成员失败，无法同步外部系统权限",
			zap.Uint("groupID", groupID),
			zap.Error(err))
		return
	}

	successCount := 0
	failCount := 0

	for _, userGroup := range userGroups {
		user := userGroup.UserIDField

		if user.Password == "" {
			failCount++
			continue
		}

		if err := s.CreateExternalUser(appID, user.ID, user.Username, user.Password, user.Email, user.Nickname, operator); err != nil {
			failCount++
			continue
		}

		if err := s.GrantRoleToUser(appID, user.ID, groupBindingID, user.Username, roleCode, operator); err != nil {
			failCount++
			continue
		}

		successCount++
	}

	logger.Info("用户组成员外部系统权限同步完成",
		zap.Uint("groupID", groupID),
		zap.Int("totalMembers", len(userGroups)),
		zap.Int("successCount", successCount),
		zap.Int("failCount", failCount))
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
				zap.String("groupCode", group.GroupCode),
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
				zap.Uint("appID", appID),
				zap.String("ruleId", rule.RuleID),
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
				zap.String("ruleID", rule.RuleID),
				zap.String("ruleName", rule.RuleName),
				zap.Error(err))
			continue
		}

		logger.Info("Jumpserver 授权规则用户详情",
			zap.String("ruleID", rule.RuleID),
			zap.String("ruleName", rule.RuleName),
			zap.Int("jumpserverUserCount", len(ruleDetail.Users)))

		// 初始化该规则的矩阵
		if matrix[rule.RuleID] == nil {
			matrix[rule.RuleID] = make(map[string]bool)
		}

		// 遍历规则中的用户，标记权限
		for _, ruleUser := range ruleDetail.Users {
			logger.Debug("Jumpserver 规则用户",
				zap.String("ruleID", rule.RuleID),
				zap.String("jumpserverUserID", ruleUser.ID),
				zap.String("jumpserverUsername", ruleUser.Username))

			// 找到对应的授权中心用户
			for _, user := range users {
				logger.Debug("授权中心用户匹配",
					zap.String("ruleID", rule.RuleID),
					zap.String("jumpserverUserID", ruleUser.ID),
					zap.Uint("authUserID", user.ID),
					zap.String("externalUserID", user.ExternalUserID),
					zap.Bool("isMatch", user.ExternalUserID == ruleUser.ID))

				if user.ExternalUserID == ruleUser.ID {
					matrix[rule.RuleID][fmt.Sprintf("%d", user.ID)] = true

					logger.Info("权限匹配成功",
						zap.String("ruleID", rule.RuleID),
						zap.String("ruleName", rule.RuleName),
						zap.Uint("authUserID", user.ID),
						zap.String("username", user.Username),
						zap.String("externalUserID", user.ExternalUserID),
						zap.String("jumpserverUserID", ruleUser.ID))

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
			zap.Uint("ruleDatabaseID", ruleID),
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
			zap.String("ruleID", authRule.RuleID),
			zap.String("ruleName", authRule.RuleName))
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
		zap.Uint("groupID", binding.GroupID),
		zap.String("groupName", group.Name),
		zap.Uint("appID", binding.AppID),
		zap.String("appName", app.Name),
		zap.String("ruleID", authRule.RuleID),
		zap.String("ruleName", authRule.RuleName))

	return nil
}

// syncExistingMembersToJumpserver 为用户组的现有成员同步 Jumpserver 权限
func (s *ApplicationPermissionService) syncExistingMembersToJumpserver(groupID, appID, groupBindingID uint, jumpserverRuleID, ruleName, operator string) {
	// 获取用户组的所有成员
	userGroups, err := s.repo.FindUserGroupsByGroupID(groupID)
	if err != nil {
		logger.Error("获取用户组成员失败，无法同步 Jumpserver 权限",
			zap.Uint("groupID", groupID),
			zap.Error(err))
		return
	}

	logger.Info("开始为用户组成员同步 Jumpserver 权限",
		zap.Uint("groupID", groupID),
		zap.Int("memberCount", len(userGroups)),
		zap.Uint("appID", appID),
		zap.String("ruleID", jumpserverRuleID))

	// 获取应用配置
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		logger.Error("获取应用配置失败",
			zap.Uint("appID", appID),
			zap.Error(err))
		return
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		logger.Error("解析认证配置失败",
			zap.Uint("appID", appID),
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
				zap.Uint("userID", user.ID),
				zap.String("username", user.Username))
			failCount++
			continue
		}

		// 1. 创建 Jumpserver 用户
		if err := s.CreateExternalUser(appID, user.ID, user.Username, user.Password, user.Email, user.Nickname, operator); err != nil {
			logger.Warn("创建 Jumpserver 用户失败",
				zap.Uint("userID", user.ID),
				zap.String("username", user.Username),
				zap.Error(err))
			// 继续执行，用户可能已存在
		}

		// 2. 获取 Jumpserver 用户 ID
		mapping, err := s.repo.FindUserIdentityMappingByUserAndApp(user.ID, appID)
		if err != nil {
			logger.Warn("获取用户身份映射失败",
				zap.Uint("userID", user.ID),
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
				zap.Uint("userID", user.ID),
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
				zap.String("ruleID", jumpserverRuleID),
				zap.Any("userIDs", jumpserverUserIDs),
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
		zap.Uint("groupID", groupID),
		zap.Int("totalMembers", len(userGroups)),
		zap.Int("successCount", successCount),
		zap.Int("failCount", failCount),
		zap.Int("addedToRule", len(jumpserverUserIDs)))
}
