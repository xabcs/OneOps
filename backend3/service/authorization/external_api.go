package authorization

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

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
