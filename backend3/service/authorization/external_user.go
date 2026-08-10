package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

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
