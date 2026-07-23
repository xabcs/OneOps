package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"oneops/backend/logger"
	"oneops/backend/models"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ApplicationPermissionService 外部应用权限服务
type ApplicationPermissionService struct{}

// NewApplicationPermissionService 创建服务实例
func NewApplicationPermissionService() *ApplicationPermissionService {
	return &ApplicationPermissionService{}
}

// CreateApplication 创建外部应用
func (s *ApplicationPermissionService) CreateApplication(app *models.Application) error {
	return db.Create(app).Error
}

// GetApplications 获取应用列表
func (s *ApplicationPermissionService) GetApplications(page, pageSize int, name string) ([]*models.Application, int64, error) {
	var apps []*models.Application
	var total int64

	query := db.Model(&models.Application{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

// GetApplicationByID 根据ID获取应用
func (s *ApplicationPermissionService) GetApplicationByID(id uint) (*models.Application, error) {
	var app models.Application
	err := db.First(&app, id).Error
	return &app, err
}

// UpdateApplication 更新应用
func (s *ApplicationPermissionService) UpdateApplication(app *models.Application) error {
	return db.Save(app).Error
}

// DeleteApplication 删除应用
func (s *ApplicationPermissionService) DeleteApplication(id uint) error {
	return db.Delete(&models.Application{}, id).Error
}

// SyncRoles 同步应用角色
func (s *ApplicationPermissionService) SyncRoles(appID uint, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	// 解析 API 配置
	var endpoints map[string]string
	if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err != nil {
		return fmt.Errorf("解析 API 配置失败: %w", err)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	var roles []models.ApplicationRole

	// 检查是否是 Jenkins 类型应用，使用 Script Console API
	if app.Type == "jenkins" {
		roles, err = s.fetchJenkinsRoles(app.BaseURL, authConfig)
		if err != nil {
			errorMsg := fmt.Sprintf("同步 Jenkins 角色失败: %w", err)
			s.logOperation(appID, "sync_roles", "", "", "failed", errorMsg, operator)
			return err
		}
	} else {
		// 调用外部 API 获取角色
		roles, err = s.fetchRolesFromAPI(app.BaseURL, endpoints, authConfig)
		if err != nil {
			errorMsg := fmt.Sprintf("同步角色失败: %w", err)
			s.logOperation(appID, "sync_roles", "", "", "failed", errorMsg, operator)
			return fmt.Errorf("%s (应用: %s, BaseURL: %s)", err, app.Name, app.BaseURL)
		}
	}

	// 删除旧的同步数据
	db.Where("app_id = ?", appID).Delete(&models.ApplicationRole{})

	// 保存新的角色数据
	for _, role := range roles {
		role.AppID = appID
		role.SyncTime = time.Now()
		db.Create(&role)
	}

	// 更新应用最后同步时间
	now := time.Now()
	app.LastSyncTime = &now
	db.Save(app)

	s.logOperation(appID, "sync_roles", "", fmt.Sprintf("同步了 %d 个角色", len(roles)), "success", "", operator)

	logger.Info("同步应用角色成功",
		zap.Uint("appID", appID),
		zap.String("appName", app.Name),
		zap.Int("roleCount", len(roles)))

	return nil
}

// fetchRolesFromAPI 从外部 API 获取角色
func (s *ApplicationPermissionService) fetchRolesFromAPI(baseURL string, endpoints map[string]string, authConfig map[string]interface{}) ([]models.ApplicationRole, error) {
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
		Code    int                       `json:"code"`
		Data    []map[string]interface{} `json:"data"`
		Message string                    `json:"message"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为 ApplicationRole
	roles := make([]models.ApplicationRole, 0)
	for _, item := range apiResp.Data {
		role := models.ApplicationRole{
			RoleCode:    s.getString(item, "code"),
			RoleName:    s.getString(item, "name"),
			RoleType:    s.getString(item, "type"),
			Description: s.getString(item, "description"),
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// SyncUsers 同步应用用户
func (s *ApplicationPermissionService) SyncUsers(appID uint, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	// 解析 API 配置
	var endpoints map[string]string
	if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err != nil {
		return fmt.Errorf("解析 API 配置失败: %w", err)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	var users []models.ApplicationUser

	// 检查是否是 Jenkins 类型应用，使用 Script Console API
	if app.Type == "jenkins" {
		users, err = s.fetchJenkinsUsers(app.BaseURL, authConfig)
		if err != nil {
			errorMsg := fmt.Sprintf("同步 Jenkins 用户失败: %w", err)
			s.logOperation(appID, "sync_users", "", "", "failed", errorMsg, operator)
			return err
		}
	} else {
		// 调用外部 API 获取用户
		users, err = s.fetchUsersFromAPI(app.BaseURL, endpoints, authConfig)
		if err != nil {
			errorMsg := fmt.Sprintf("同步用户失败: %w", err)
			s.logOperation(appID, "sync_users", "", "", "failed", errorMsg, operator)
			return fmt.Errorf("%s", err)
		}
	}

	// 删除旧的同步数据
	db.Where("app_id = ?", appID).Delete(&models.ApplicationUser{})

	// 保存新的用户数据
	for _, user := range users {
		user.AppID = appID
		user.SyncTime = time.Now()
		db.Create(&user)
	}

	s.logOperation(appID, "sync_users", "", fmt.Sprintf("同步了 %d 个用户", len(users)), "success", "", operator)

	logger.Info("同步应用用户成功",
		zap.Uint("appID", appID),
		zap.String("appName", app.Name),
		zap.Int("userCount", len(users)))

	return nil
}

// fetchUsersFromAPI 从外部 API 获取用户
func (s *ApplicationPermissionService) fetchUsersFromAPI(baseURL string, endpoints map[string]string, authConfig map[string]interface{}) ([]models.ApplicationUser, error) {
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
		Code    int                       `json:"code"`
		Data    []map[string]interface{} `json:"data"`
		Message string                    `json:"message"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	users := make([]models.ApplicationUser, 0)
	for _, item := range apiResp.Data {
		user := models.ApplicationUser{
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
	AppName     string `json:"appName"`     // 应用名称
	Username    string `json:"username"`     // 用户名
	Password    string `json:"password"`     // 初始密码
	RoleCode    string `json:"roleCode"`     // 分配的角色代码
	RoleName    string `json:"roleName"`     // 角色名称
	CreateError string `json:"createError"`  // 创建用户错误信息（如果失败）
	GrantError  string `json:"grantError"`   // 授权错误信息（如果失败）
	Success     bool   `json:"success"`      // 是否成功
}

// JenkinsSession Jenkins 会话信息
type JenkinsSession struct {
	Client   *http.Client
	Crumb    map[string]string
	BaseURL  string
}

// createJenkinsSession 创建 Jenkins 会话（带 Cookie 支持）
func (s *ApplicationPermissionService) createJenkinsSession(app *models.Application, authConfig map[string]interface{}) (*JenkinsSession, error) {
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
		Crumb              string `json:"crumb"`
	}
	if err := json.Unmarshal(body, &crumbResponse); err != nil {
		logger.Warn("解析 Jenkins crumb JSON 失败", zap.Error(err), zap.String("body", string(body)))
		return nil, fmt.Errorf("解析 crumb 响应失败: %w", err)
	}

	if crumbResponse.CrumbRequestField == "" || crumbResponse.Crumb == "" {
		logger.Warn("Jenkins crumb 字段为空，可能 CSRF 保护已禁用",
			zap.String("crumbRequestField", crumbResponse.CrumbRequestField),
			zap.String("crumb", crumbResponse.Crumb))
		// CSRF 保护可能已禁用，返回空 crumb
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
func (s *ApplicationPermissionService) getJenkinsCrumb(app *models.Application, authConfig map[string]interface{}) (map[string]string, error) {
	baseURL := strings.TrimSuffix(app.BaseURL, "/")
	crumbURL := baseURL + "/crumbIssuer/api/json"

	logger.Info("尝试获取 Jenkins crumb", zap.String("crumbURL", crumbURL))

	// 创建请求
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

	// 发送请求
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

	// 解析响应
	var crumbResponse struct {
		CrumbRequestField string `json:"crumbRequestField"`
		Crumb              string `json:"crumb"`
	}
	if err := json.Unmarshal(body, &crumbResponse); err != nil {
		logger.Warn("解析 Jenkins crumb JSON 失败", zap.Error(err), zap.String("body", string(body)))
		return nil, fmt.Errorf("解析 crumb 响应失败: %w", err)
	}

	if crumbResponse.CrumbRequestField == "" || crumbResponse.Crumb == "" {
		logger.Warn("Jenkins crumb 字段为空，可能 CSRF 保护已禁用",
			zap.String("crumbRequestField", crumbResponse.CrumbRequestField),
			zap.String("crumb", crumbResponse.Crumb))
		return nil, fmt.Errorf("crumb 字段为空，可能 CSRF 保护已禁用")
	}

	result := map[string]string{
		crumbResponse.CrumbRequestField: crumbResponse.Crumb,
	}

	logger.Info("成功获取 Jenkins crumb",
		zap.String("crumbField", crumbResponse.CrumbRequestField),
		zap.String("crumbValue", crumbResponse.Crumb))

	return result, nil
}

// CreateExternalUser 在外部系统创建用户
func (s *ApplicationPermissionService) CreateExternalUser(appID uint, username, password, email, fullName string, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	// 解析 API 配置
	var endpoints map[string]string
	if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err != nil {
		return fmt.Errorf("解析 API 配置失败: %w", err)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	// 构建请求数据
	reqData := map[string]interface{}{
		"username": username,
		"password": password,
	}

	// 添加可选字段
	if email != "" {
		reqData["email"] = email
	}
	if fullName != "" {
		reqData["fullName"] = fullName
		// 某些系统可能使用displayName
		reqData["displayName"] = fullName
	}

	jsonData, _ := json.Marshal(reqData)

	// 如果是 Jenkins，使用带会话的请求方式
	if app.Type == "jenkins" {
		return s.createJenkinsUser(app, authConfig, endpoints, jsonData, username, operator)
	}

	// 其他系统使用原有逻辑
	return s.createOtherSystemUser(app, authConfig, endpoints, jsonData, username, operator)
}

// createJenkinsUser 在 Jenkins 中创建用户（使用会话保持 Cookie）
func (s *ApplicationPermissionService) createJenkinsUser(app *models.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, operator string) error {
	// 创建 Jenkins 会话
	session, err := s.createJenkinsSession(app, authConfig)
	if err != nil {
		logger.Warn("创建 Jenkins 会话失败，尝试不使用会话进行请求", zap.Error(err))
		// 降级到无会话模式
		return s.createOtherSystemUser(app, authConfig, endpoints, jsonData, username, operator)
	}

	// Jenkins 使用 form-urlencoded 格式，需要转换数据
	var jsonParams map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonParams); err != nil {
		return fmt.Errorf("解析用户数据失败: %w", err)
	}

	// 构建 form 表单数据
	formData := url.Values{}
	formData.Set("username", username)
	if pwd, ok := jsonParams["password"].(string); ok {
		formData.Set("password1", pwd)  // Jenkins 表单字段名
		formData.Set("password2", pwd)  // 确认密码
	}

	// email 字段（Jenkins 会验证邮箱格式）
	email := ""
	if e, ok := jsonParams["email"].(string); ok && e != "" {
		// 简单的邮箱格式验证
		if strings.Contains(e, "@") && strings.Contains(e, ".") {
			email = e
		} else {
			// 邮箱格式无效，使用默认值
			logger.Warn("邮箱格式无效，使用默认邮箱", zap.String("originalEmail", e))
			email = username + "@example.com"
		}
	} else {
		// 如果没有 email，生成一个默认的
		email = username + "@example.com"
	}
	formData.Set("email", email)

	// fullname 字段 - Jenkins 可能需要这个字段
	fullName := ""
	if fn, ok := jsonParams["fullName"].(string); ok && fn != "" {
		fullName = fn
	} else if dn, ok := jsonParams["displayName"].(string); ok && dn != "" {
		fullName = dn
	}

	// 如果 fullname 为空，使用 username 作为默认值
	if fullName == "" {
		fullName = username
	}
	formData.Set("fullname", fullName)

	// ⭐ 重要：Jenkins-Crumb 必须作为表单字段提交，而不是 HTTP Header
	if session.Crumb != nil {
		for key, value := range session.Crumb {
			formData.Set(key, value)
			logger.Info("已添加 crumb 到表单字段", zap.String("fieldName", key), zap.String("fieldValue", value))
		}
	}

	// 使用正确的 Jenkins API 端点
	createUserURL := session.BaseURL + "/securityRealm/createAccountByAdmin"
	logger.Info("发送创建用户请求到 Jenkins",
		zap.String("url", createUserURL),
		zap.String("username", username),
		zap.String("fullname", fullName),
		zap.String("email", email))

	req, err := http.NewRequest("POST", createUserURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	// 设置表单内容类型
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 设置认证
	if authConfig["type"] == "basic" {
		authUsername, _ := authConfig["username"].(string)
		authPassword, _ := authConfig["password"].(string)
		if authUsername != "" && authPassword != "" {
			req.SetBasicAuth(authUsername, authPassword)
		}
	}

	// 使用会话客户端发送请求（会自动携带 Cookie）
	resp, err := session.Client.Do(req)
	if err != nil {
		s.logOperation(app.ID, "create_user", username, formData.Encode(), "failed", err.Error(), operator)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 记录详细响应信息
	bodyStr := string(body)
	if len(bodyStr) > 500 {
		bodyStr = bodyStr[:500]
	}
	logger.Info("Jenkins 创建用户响应",
		zap.Int("statusCode", resp.StatusCode),
		zap.String("location", resp.Header.Get("Location")),
		zap.String("body", bodyStr))

	// 检查响应中是否包含错误信息（即使状态码是 200）
	if strings.Contains(strings.ToLower(bodyStr), "error") || strings.Contains(bodyStr, "Oops") {
		s.logOperation(app.ID, "create_user", username, formData.Encode(), "failed", bodyStr, operator)
		return fmt.Errorf("创建用户失败，Jenkins 返回错误信息")
	}

	// Jenkins 成功创建用户后会重定向到用户列表页，接受 302 状态码
	// 或者返回 200 状态码（空响应或成功页面）
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusFound {
		s.logOperation(app.ID, "create_user", username, formData.Encode(), "failed", bodyStr, operator)
		return fmt.Errorf("创建用户失败 (状态码 %d)", resp.StatusCode)
	}

	s.logOperation(app.ID, "create_user", username, formData.Encode(), "success", "", operator)

	logger.Info("在 Jenkins 创建用户成功",
		zap.Uint("appID", app.ID),
		zap.String("appName", app.Name),
		zap.String("username", username))

	return nil
}

// createOtherSystemUser 在其他外部系统创建用户
func (s *ApplicationPermissionService) createOtherSystemUser(app *models.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, operator string) error {
	// 调用创建用户 API
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

	// 支持 Basic Auth 认证
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

	logger.Info("在外部系统创建用户成功",
		zap.Uint("appID", app.ID),
		zap.String("appName", app.Name),
		zap.String("username", username))

	return nil
}

// GrantRoleToUser 为用户授予外部系统角色
func (s *ApplicationPermissionService) GrantRoleToUser(appID uint, username, roleCode string, operator string) error {
	app, err := s.GetApplicationByID(appID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}

	// 解析 API 配置
	var endpoints map[string]string
	if err := json.Unmarshal([]byte(app.Endpoints), &endpoints); err != nil {
		return fmt.Errorf("解析 API 配置失败: %w", err)
	}

	// 解析认证配置
	var authConfig map[string]interface{}
	if err := json.Unmarshal([]byte(app.AuthConfig), &authConfig); err != nil {
		return fmt.Errorf("解析认证配置失败: %w", err)
	}

	// 构建请求数据
	reqData := map[string]interface{}{
		"username": username,
		"roleCode": roleCode,
	}

	jsonData, _ := json.Marshal(reqData)

	// 如果是 Jenkins，使用带会话的请求方式
	if app.Type == "jenkins" {
		return s.grantJenkinsRole(app, authConfig, endpoints, jsonData, username, roleCode, operator)
	}

	// 其他系统使用原有逻辑
	return s.grantOtherSystemRole(app, authConfig, endpoints, jsonData, username, roleCode, operator)
}

// grantJenkinsRole 在 Jenkins 中授予角色（使用会话保持 Cookie）
func (s *ApplicationPermissionService) grantJenkinsRole(app *models.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, roleCode, operator string) error {
	// 创建 Jenkins 会话
	session, err := s.createJenkinsSession(app, authConfig)
	if err != nil {
		return fmt.Errorf("创建 Jenkins 会话失败: %w", err)
	}

	// Jenkins 角色分配使用表单格式 (Role-Based Strategy Plugin)
	// Jenkins 角色类型：globalRoles（全局角色）或 projectRoles（项目角色/Item角色）
	formData := url.Values{}

	// 根据角色名称判断角色类型
	// 通常项目角色会包含项目名称或特定的命名模式
	// 如果角色名称包含特定前缀或后缀，视为项目角色
	roleType := "globalRoles"  // 注意：是复数形式
	if strings.Contains(strings.ToLower(roleCode), "item") ||
	   strings.Contains(strings.ToLower(roleCode), "project") ||
	   strings.Contains(roleCode, "-") || // 带连字符的通常是项目角色
	   strings.Contains(roleCode, "_") {  // 带下划线的通常是项目角色
		roleType = "projectRoles"  // 注意：是复数形式
	}

	formData.Set("type", roleType)
	formData.Set("roleName", roleCode)
	formData.Set("sid", username)

	// ⭐ 重要：Jenkins-Crumb 必须作为表单字段提交
	if session.Crumb != nil {
		for key, value := range session.Crumb {
			formData.Set(key, value)
			logger.Info("已添加 crumb 到表单字段", zap.String("fieldName", key))
		}
	}

	assignRoleURL := session.BaseURL + "/role-strategy/strategy/assignRole"
	logger.Info("发送角色分配请求到 Jenkins",
		zap.String("url", assignRoleURL),
		zap.String("username", username),
		zap.String("roleCode", roleCode),
		zap.String("roleType", roleType))

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
		logger.Info("在 Jenkins 为用户授予角色成功",
			zap.Uint("appID", app.ID),
			zap.String("appName", app.Name),
			zap.String("username", username),
			zap.String("roleCode", roleCode),
			zap.String("roleType", roleType))
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	errorMsg := string(body)

	s.logOperation(app.ID, "grant_role", fmt.Sprintf("%s -> %s", username, roleCode), formData.Encode(), "failed", errorMsg, operator)
	return fmt.Errorf("授权失败 (状态码 %d): %s", resp.StatusCode, errorMsg)
}

// grantOtherSystemRole 在其他外部系统授予角色
func (s *ApplicationPermissionService) grantOtherSystemRole(app *models.Application, authConfig map[string]interface{}, endpoints map[string]string, jsonData []byte, username, roleCode, operator string) error {
	// 调用授权 API
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

	// 支持 Basic Auth 认证
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

	logger.Info("为用户授予角色成功",
		zap.Uint("appID", app.ID),
		zap.String("appName", app.Name),
		zap.String("username", username),
		zap.String("roleCode", roleCode))

	return nil
}

// AssignUserToGroup 为用户分配用户组成员，返回授权结果
func (s *ApplicationPermissionService) AssignUserToGroup(userID, groupID uint, operator string) ([]ExternalUserResult, error) {
	// 检查是否已分配
	var count int64
	db.Model(&models.AuthUserGroup{}).Where("user_id = ? AND group_id = ?", userID, groupID).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("用户已分配该用户组")
	}

	// 获取用户组绑定的所有外部角色
	var bindings []models.GroupBinding
	if err := db.Where("group_id = ?", groupID).Preload("ApplicationRole").Preload("AppIDField").Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("获取用户组绑定失败: %w", err)
	}

	// 如果用户组没有绑定任何外部角色，直接创建用户组成员关系
	if len(bindings) == 0 {
		// 检查用户和用户组是否存在
		var user models.AuthUser
		if err := db.First(&user, userID).Error; err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %w", err)
		}

		var group models.AuthGroup
		if err := db.First(&group, groupID).Error; err != nil {
			return nil, fmt.Errorf("获取用户组信息失败: %w", err)
		}

		// 创建用户组成员关系
		userGroup := models.AuthUserGroup{
			UserID:    userID,
			GroupID:   groupID,
			GrantedBy: operator,
			GrantedAt: time.Now(),
		}

		if err := db.Create(&userGroup).Error; err != nil {
			return nil, fmt.Errorf("创建用户组成员失败: %w", err)
		}

		logger.Info("为用户分配用户组成功（无外部角色）",
			zap.Uint("userID", userID),
			zap.Uint("groupID", groupID),
			zap.String("operator", operator))

		// 返回空结果，表示没有外部系统需要授权
		return []ExternalUserResult{}, nil
	}

	// 获取授权中心用户信息
	var user models.AuthUser
	if err := db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 收集授权结果
	results := make([]ExternalUserResult, 0)

	// 检查用户是否有密码
	if user.Password == "" {
		return nil, fmt.Errorf("用户未设置初始密码，请先重置用户密码")
	}

	// 为每个绑定创建外部用户并授权
	for _, binding := range bindings {
		result := ExternalUserResult{
			AppName:  binding.AppIDField.Name,
			Username: user.Username,
			RoleCode: binding.ApplicationRole.RoleCode,
			RoleName: binding.ApplicationRole.RoleName,
			Password: user.Password, // 使用用户的统一密码
			Success:  true,
		}

		// 1. 创建外部用户（如果不存在）
		if err := s.CreateExternalUser(binding.AppID, user.Username, user.Password, user.Email, user.Nickname, operator); err != nil {
			logger.Warn("创建外部用户失败，继续授权", zap.Error(err))
			result.CreateError = err.Error()
			// 继续执行，用户可能已存在
		}

		// 2. 授予角色
		if err := s.GrantRoleToUser(binding.AppID, user.Username, binding.ApplicationRole.RoleCode, operator); err != nil {
			result.Success = false
			result.GrantError = err.Error()
			results = append(results, result)
			continue
		}

		results = append(results, result)
	}

	// 保存用户组成员分配
	userGroup := models.AuthUserGroup{
		UserID:    userID,
		GroupID:   groupID,
		GrantedBy: operator,
		GrantedAt: time.Now(),
	}

	if err := db.Create(&userGroup).Error; err != nil {
		return results, fmt.Errorf("保存用户组成员失败: %w", err)
	}

	logger.Info("为用户分配用户组成功",
		zap.Uint("userID", userID),
		zap.Uint("groupID", groupID),
		zap.String("operator", operator))

	return results, nil
}

// GetApplicationRoles 获取应用角色列表
func (s *ApplicationPermissionService) GetApplicationRoles(appID uint) ([]*models.ApplicationRole, error) {
	var roles []*models.ApplicationRole
	err := db.Where("app_id = ?", appID).Find(&roles).Error
	return roles, err
}

// GetApplicationUsers 获取应用用户列表
func (s *ApplicationPermissionService) GetApplicationUsers(appID uint) ([]*models.ApplicationUser, error) {
	var users []*models.ApplicationUser
	err := db.Where("app_id = ?", appID).Find(&users).Error
	return users, err
}

// CreateGroupBinding 创建用户组绑定
func (s *ApplicationPermissionService) CreateGroupBinding(binding *models.GroupBinding) error {
	// 验证关联的实体是否存在

	// 1. 检查用户组是否存在
	var group models.AuthGroup
	if err := db.First(&group, binding.GroupID).Error; err != nil {
		return fmt.Errorf("用户组不存在 (ID: %d): %w", binding.GroupID, err)
	}

	// 2. 检查应用是否存在
	var app models.Application
	if err := db.First(&app, binding.AppID).Error; err != nil {
		return fmt.Errorf("外部应用不存在 (ID: %d): %w", binding.AppID, err)
	}

	// 3. 检查应用角色是否存在
	var role models.ApplicationRole
	if err := db.Where("id = ? AND app_id = ?", binding.ApplicationRoleID, binding.AppID).First(&role).Error; err != nil {
		return fmt.Errorf("应用角色不存在 (RoleID: %d, AppID: %d): %w", binding.ApplicationRoleID, binding.AppID, err)
	}

	// 4. 创建绑定记录
	if err := db.Create(binding).Error; err != nil {
		return fmt.Errorf("创建用户组绑定失败: %w", err)
	}

	// 5. 为该用户组的所有现有成员在外部系统中创建用户并授权
	go s.syncExistingMembersToExternalSystem(binding.GroupID, binding.AppID, binding.ApplicationRoleID, role.RoleCode, role.RoleName, "system")

	logger.Info("创建用户组绑定成功，开始为现有成员同步外部系统权限",
		zap.Uint("groupID", binding.GroupID),
		zap.String("groupName", group.Name),
		zap.Uint("appID", binding.AppID),
		zap.String("appName", app.Name),
		zap.String("roleCode", role.RoleCode))

	return nil
}

// GetGroupBindings 获取用户组的所有绑定
func (s *ApplicationPermissionService) GetGroupBindings(groupID uint) ([]*models.GroupBinding, error) {
	var bindings []*models.GroupBinding
	err := db.Where("group_id = ?", groupID).
		Preload("ApplicationRole").
		Preload("AppIDField").
		Find(&bindings).Error
	return bindings, err
}

// DeleteGroupBinding 删除用户组绑定
func (s *ApplicationPermissionService) DeleteGroupBinding(id uint) error {
	return db.Delete(&models.GroupBinding{}, id).Error
}

// UserGroupResponse 用户组响应结构
type UserGroupResponse struct {
	ID         uint      `json:"id"`
	GroupID    uint      `json:"groupId"`
	GroupName  string    `json:"groupName"`
	GroupCode  string    `json:"groupCode"`
	GrantedBy  string    `json:"grantedBy"`
	GrantedAt  time.Time `json:"grantedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

// GetUserGroups 获取用户的用户组列表
func (s *ApplicationPermissionService) GetUserGroups(userID uint) ([]UserGroupResponse, error) {
	var userGroups []*models.AuthUserGroup
	err := db.Where("user_id = ?", userID).Preload("GroupIDField").Find(&userGroups).Error
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	result := make([]UserGroupResponse, 0, len(userGroups))
	for _, ug := range userGroups {
		result = append(result, UserGroupResponse{
			ID:         ug.ID,
			GroupID:    ug.GroupID,
			GroupName:  ug.GroupIDField.Name,
			GroupCode:  ug.GroupIDField.Code,
			GrantedBy:  ug.GrantedBy,
			GrantedAt:  ug.GrantedAt,
			CreatedAt:  ug.CreatedAt,
		})
	}

	return result, nil
}

// DeleteUserGroup 删除用户组成员
func (s *ApplicationPermissionService) DeleteUserGroup(userID, groupID uint) error {
	return db.Where("user_id = ? AND group_id = ?", userID, groupID).Delete(&models.AuthUserGroup{}).Error
}

// GetOperationLogs 获取操作日志
func (s *ApplicationPermissionService) GetOperationLogs(appID uint, page, pageSize int) ([]*models.ApplicationOperationLog, int64, error) {
	var logs []*models.ApplicationOperationLog
	var total int64

	query := db.Model(&models.ApplicationOperationLog{}).Where("app_id = ?", appID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// logOperation 记录操作日志
func (s *ApplicationPermissionService) logOperation(appID uint, operation, target, requestData, status, errorMsg, operator string) {
	log := models.ApplicationOperationLog{
		AppID:        appID,
		Operation:    operation,
		Target:       target,
		RequestData:  requestData,
		Status:       status,
		ErrorMsg:     errorMsg,
		Operator:     operator,
	}

	db.Create(&log)
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
func (s *ApplicationPermissionService) CreateAuthUser(user *models.AuthUser) (string, error) {
	// 生成随机密码
	password, err := GenerateRandomPassword(DefaultPasswordConfig)
	if err != nil {
		return "", fmt.Errorf("生成密码失败: %w", err)
	}

	// 加密存储密码（这里使用简单的处理，实际应使用 bcrypt）
	// 注意：这里暂不加密，直接存储，后续可改进
	user.Password = password

	if err := db.Create(user).Error; err != nil {
		return "", err
	}

	// 返回明文密码，用于显示给管理员
	return password, nil
}

// GetAuthUsers 获取授权中心用户列表
func (s *ApplicationPermissionService) GetAuthUsers(page, pageSize int, username string) ([]*models.AuthUser, int64, error) {
	var users []*models.AuthUser
	var total int64

	query := db.Model(&models.AuthUser{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetAuthUserByID 根据ID获取授权中心用户
func (s *ApplicationPermissionService) GetAuthUserByID(id uint) (*models.AuthUser, error) {
	var user models.AuthUser
	err := db.First(&user, id).Error
	return &user, err
}

// UpdateAuthUser 更新授权中心用户
func (s *ApplicationPermissionService) UpdateAuthUser(user *models.AuthUser) error {
	return db.Save(user).Error
}

// DeleteAuthUser 删除授权中心用户
func (s *ApplicationPermissionService) DeleteAuthUser(id uint) error {
	return db.Delete(&models.AuthUser{}, id).Error
}

// GetAllAuthUsers 获取所有授权中心用户
func (s *ApplicationPermissionService) GetAllAuthUsers() ([]*models.AuthUser, error) {
	var users []*models.AuthUser
	err := db.Where("status = ?", 1).Find(&users).Error
	return users, err
}

// ========== 授权中心角色管理 ==========

// CreateAuthRole 创建授权中心角色
func (s *ApplicationPermissionService) CreateAuthGroup(group *models.AuthGroup) error {
	return db.Create(group).Error
}

// GetAuthGroups 获取授权中心用户组列表
func (s *ApplicationPermissionService) GetAuthGroups(page, pageSize int, name string) ([]*models.AuthGroup, int64, error) {
	var groups []*models.AuthGroup
	var total int64

	query := db.Model(&models.AuthGroup{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// GetAuthGroupByID 根据ID获取授权中心用户组
func (s *ApplicationPermissionService) GetAuthGroupByID(id uint) (*models.AuthGroup, error) {
	var group models.AuthGroup
	err := db.First(&group, id).Error
	return &group, err
}

// UpdateAuthGroup 更新授权中心用户组
func (s *ApplicationPermissionService) UpdateAuthGroup(group *models.AuthGroup) error {
	return db.Save(group).Error
}

// DeleteAuthGroup 删除授权中心用户组
func (s *ApplicationPermissionService) DeleteAuthGroup(id uint) error {
	return db.Delete(&models.AuthGroup{}, id).Error
}

// GetAllAuthGroups 获取所有授权中心用户组
func (s *ApplicationPermissionService) GetAllAuthGroups() ([]*models.AuthGroup, error) {
	var groups []*models.AuthGroup
	err := db.Where("status = ?", 1).Find(&groups).Error
	return groups, err
}

// ========== Jenkins Script Console API 支持 ==========

// JenkinsCrumb CSRF Token 结构
type JenkinsCrumb struct {
	Crumb             string `json:"crumb"`
	CrumbRequestField string `json:"crumbRequestField"`
}

// executeJenkinsScript 执行 Jenkins Groovy 脚本
func (s *ApplicationPermissionService) executeJenkinsScript(baseURL, username, password, script string) (string, error) {
	// 创建 Cookie Jar 来保存会话信息
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("创建 Cookie Jar 失败: %w", err)
	}

	// 创建共享 Cookie 的 HTTP Client
	client := &http.Client{
		Jar: jar,
	}

	// 1. 获取 CSRF Token（同时保存 Cookie）
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
	defer crumbResp.Body.Close()

	if crumbResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取 CSRF Token 失败，状态码: %d", crumbResp.StatusCode)
	}

	var crumb JenkinsCrumb
	crumbBody, _ := io.ReadAll(crumbResp.Body)
	if err := json.Unmarshal(crumbBody, &crumb); err != nil {
		return "", fmt.Errorf("解析 CSRF Token 失败: %w", err)
	}

	// 2. 执行脚本（使用相同的 Cookie）
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
func (s *ApplicationPermissionService) fetchJenkinsUsers(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationUser, error) {
	// 获取认证信息
	username, _ := authConfig["username"].(string)
	password, _ := authConfig["password"].(string)

	if username == "" || password == "" {
		return nil, fmt.Errorf("Jenkins 需要 Basic Auth 认证，请配置用户名和密码")
	}

	// Groovy 脚本：获取所有用户
	script := `
import jenkins.model.*
import hudson.tasks.Mailer

def users = Jenkins.instance.securityRealm.getAllUsers()
users.each { user ->
  def email = user.getProperty(Mailer.UserProperty)?.address ?: ""
  println "${user.id}|${user.fullName}|${email}|active"
}
`

	// 执行脚本
	result, err := s.executeJenkinsScript(baseURL, username, password, script)
	if err != nil {
		return nil, err
	}

	// 解析结果
	users := make([]models.ApplicationUser, 0)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Result:") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			user := models.ApplicationUser{
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
func (s *ApplicationPermissionService) fetchJenkinsRoles(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationRole, error) {
	// 获取认证信息
	username, _ := authConfig["username"].(string)
	password, _ := authConfig["password"].(string)

	if username == "" || password == "" {
		return nil, fmt.Errorf("Jenkins 需要 Basic Auth 认证，请配置用户名和密码")
	}

	// Groovy 脚本：获取所有角色
	// 注意：需要安装 Role Strategy 插件
	script := `
import com.michelin.cio.hudson.plugins.rolestrategy.*
import com.synopsys.arc.jenkins.plugins.rolestrategy.*
import jenkins.model.*

def rbas = Jenkins.instance.getAuthorizationStrategy()
def roles = []

if (rbas instanceof RoleBasedAuthorizationStrategy) {
    // 获取全局角色
    def globalRoles = rbas.getRoleMap(RoleType.Global).getRoles()
    globalRoles.each { role ->
        println "${role.name}|${role.name}|global|Global Role"
    }

    // 获取项目角色（注意：使用 Project 而不是 Item）
    def projectRoles = rbas.getRoleMap(RoleType.Project).getRoles()
    projectRoles.each { role ->
        println "${role.name}|${role.name}|project|Project Role"
    }
} else {
    println "Warning: Role Strategy plugin not configured"
}
`

	// 执行脚本
	result, err := s.executeJenkinsScript(baseURL, username, password, script)
	if err != nil {
		return nil, err
	}

	// 解析结果
	roles := make([]models.ApplicationRole, 0)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Warning:") || strings.HasPrefix(line, "Result:") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			role := models.ApplicationRole{
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

	// 如果没有角色，添加一个默认提示
	if len(roles) == 0 {
		roles = append(roles, models.ApplicationRole{
			RoleCode:    "admin",
			RoleName:    "Administrator",
			RoleType:    "global",
			Description: "Jenkins Administrator (请在 Jenkins 中配置角色)",
		})
		roles = append(roles, models.ApplicationRole{
			RoleCode:    "developer",
			RoleName:    "Developer",
			RoleType:    "global",
			Description: "Jenkins Developer (请在 Jenkins 中配置角色)",
		})
	}

	return roles, nil
}



// syncExistingMembersToExternalSystem 为用户组的现有成员同步外部系统权限
func (s *ApplicationPermissionService) syncExistingMembersToExternalSystem(groupID, appID, roleID uint, roleCode, roleName, operator string) {
	// 获取用户组的所有成员
	var userGroups []*models.AuthUserGroup
	if err := db.Where("group_id = ?", groupID).Preload("UserIDField").Find(&userGroups).Error; err != nil {
		logger.Error("获取用户组成员失败，无法同步外部系统权限",
			zap.Uint("groupID", groupID),
			zap.Error(err))
		return
	}

	logger.Info("开始为用户组成员同步外部系统权限",
		zap.Uint("groupID", groupID),
		zap.Int("memberCount", len(userGroups)),
		zap.Uint("appID", appID))

	successCount := 0
	failCount := 0

	// 为每个成员在外部系统中创建用户并授权
	for _, userGroup := range userGroups {
		user := userGroup.UserIDField

		// 检查用户是否有密码
		if user.Password == "" {
			logger.Warn("用户未设置初始密码，跳过外部系统授权",
				zap.Uint("userID", user.ID),
				zap.String("username", user.Username))
			failCount++
			continue
		}

		// 1. 创建外部用户（如果不存在）
		if err := s.CreateExternalUser(appID, user.Username, user.Password, user.Email, user.Nickname, operator); err != nil {
			logger.Warn("创建外部用户失败",
				zap.Uint("userID", user.ID),
				zap.String("username", user.Username),
				zap.Uint("appID", appID),
				zap.Error(err))
			failCount++
			continue
		}

		// 2. 授予角色
		if err := s.GrantRoleToUser(appID, user.Username, roleCode, operator); err != nil {
			logger.Warn("授予外部角色失败",
				zap.Uint("userID", user.ID),
				zap.String("username", user.Username),
				zap.String("roleCode", roleCode),
				zap.Uint("appID", appID),
				zap.Error(err))
			failCount++
			continue
		}

		successCount++
		logger.Info("为用户组成员同步外部系统权限成功",
			zap.Uint("userID", user.ID),
			zap.String("username", user.Username),
			zap.String("roleCode", roleCode),
			zap.Uint("appID", appID))
	}

	logger.Info("用户组成员外部系统权限同步完成",
		zap.Uint("groupID", groupID),
		zap.Int("totalMembers", len(userGroups)),
		zap.Int("successCount", successCount),
		zap.Int("failCount", failCount))
}
