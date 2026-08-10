package authorization

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"oneops/backend2/pkg/logger"
	"strings"
	"time"

	"go.uber.org/zap"
)

// JenkinsAdapter Jenkins 应用适配器
type JenkinsAdapter struct{}

// ValidateConfig 验证 Jenkins 配置
func (j *JenkinsAdapter) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["username"]; !ok {
		return fmt.Errorf("Jenkins 需要 username 配置")
	}
	if _, ok := config["password"]; !ok {
		return fmt.Errorf("Jenkins 需要 password 配置")
	}
	return nil
}

// FetchUsers 获取 Jenkins 用户列表
func (j *JenkinsAdapter) FetchUsers(baseURL string, authConfig map[string]interface{}) ([]ApplicationUser, error) {
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
	result, err := j.executeScript(baseURL, username, password, script)
	if err != nil {
		return nil, err
	}

	// 解析结果
	users := make([]ApplicationUser, 0)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Result:") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			user := ApplicationUser{
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

// FetchRoles 获取 Jenkins 角色列表
func (j *JenkinsAdapter) FetchRoles(baseURL string, authConfig map[string]interface{}) ([]ApplicationRole, error) {
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
	result, err := j.executeScript(baseURL, username, password, script)
	if err != nil {
		return nil, err
	}

	// 解析结果
	roles := make([]ApplicationRole, 0)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Warning:") || strings.HasPrefix(line, "Result:") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			role := ApplicationRole{
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

	return roles, nil
}

// CreateUser 在 Jenkins 中创建用户
func (j *JenkinsAdapter) CreateUser(baseURL string, authConfig map[string]interface{}, user *UserCreateRequest) (string, error) {
	// 创建 Jenkins 会话
	session, err := j.createSession(baseURL, authConfig)
	if err != nil {
		return "", fmt.Errorf("创建 Jenkins 会话失败: %w", err)
	}

	// 构建 form 表单数据
	formData := url.Values{}
	formData.Set("username", user.Username)
	formData.Set("password1", user.Password) // Jenkins 表单字段名
	formData.Set("password2", user.Password) // 确认密码
	formData.Set("fullname", user.FullName)

	// email 字段（Jenkins 会验证邮箱格式）
	email := user.Email
	if email == "" || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		logger.Warn("邮箱格式无效或为空，使用默认邮箱",
			zap.String("originalEmail", user.Email),
			zap.String("defaultEmail", user.Username+"@example.com"))
		email = user.Username + "@example.com"
	}
	formData.Set("email", email)

	// ⭐ 重要：Jenkins-Crumb 必须作为表单字段提交
	if session.Crumb != nil {
		for key, value := range session.Crumb {
			formData.Set(key, value)
			logger.Info("已添加 crumb 到表单字段", zap.String("fieldName", key))
		}
	}

	createUserURL := session.BaseURL + "/securityRealm/createAccountByAdmin"
	logger.Info("发送用户创建请求到 Jenkins",
		zap.String("url", createUserURL),
		zap.String("username", user.Username))

	req, err := http.NewRequest("POST", createUserURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 设置 Basic Auth
	if username, ok := authConfig["username"].(string); ok {
		if password, ok := authConfig["password"].(string); ok {
			req.SetBasicAuth(username, password)
		}
	}

	resp, err := session.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Jenkins 成功创建用户时返回 302 重定向或 200
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
		logger.Info("在 Jenkins 创建用户成功",
			zap.String("username", user.Username),
			zap.String("email", email))
		// Jenkins 使用用户名作为用户 ID
		return user.Username, nil
	}

	// 失败时解析错误信息
	body, _ := io.ReadAll(resp.Body)
	errorMsg := string(body)

	// 尝试从 HTML 中提取错误信息
	if strings.Contains(errorMsg, "无效的 e-mail 地址") {
		return "", fmt.Errorf("无效的 e-mail 地址: %s", email)
	}

	return "", fmt.Errorf("创建用户失败 (状态码 %d): %s", resp.StatusCode, errorMsg)
}

// AssignRole 在 Jenkins 中为用户分配角色
func (j *JenkinsAdapter) AssignRole(baseURL string, authConfig map[string]interface{}, username, roleCode string) error {
	// 创建 Jenkins 会话
	session, err := j.createSession(baseURL, authConfig)
	if err != nil {
		return fmt.Errorf("创建 Jenkins 会话失败: %w", err)
	}

	// Jenkins 角色分配使用表单格式 (Role-Based Strategy Plugin)
	// Jenkins 角色类型：globalRoles（全局角色）或 projectRoles（项目角色/Item角色）
	formData := url.Values{}

	// 根据角色名称判断角色类型
	roleType := "globalRoles" // 注意：是复数形式
	if strings.Contains(strings.ToLower(roleCode), "item") ||
		strings.Contains(strings.ToLower(roleCode), "project") ||
		strings.Contains(roleCode, "-") || // 带连字符的通常是项目角色
		strings.Contains(roleCode, "_") { // 带下划线的通常是项目角色
		roleType = "projectRoles" // 注意：是复数形式
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

	// 设置 Basic Auth
	if username, ok := authConfig["username"].(string); ok {
		if password, ok := authConfig["password"].(string); ok {
			req.SetBasicAuth(username, password)
		}
	}

	resp, err := session.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
		logger.Info("在 Jenkins 为用户授予角色成功",
			zap.String("username", username),
			zap.String("roleCode", roleCode),
			zap.String("roleType", roleType))
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	errorMsg := string(body)
	return fmt.Errorf("授权失败 (状态码 %d): %s", resp.StatusCode, errorMsg)
}

// GetConfigTemplate 获取 Jenkins 配置模板
func (j *JenkinsAdapter) GetConfigTemplate() map[string]interface{} {
	return map[string]interface{}{
		"username":     "",
		"password":     "",
		"syncInterval": 300,
	}
}

// GetDisplayName 获取 Jenkins 显示名称
func (j *JenkinsAdapter) GetDisplayName() string {
	return "Jenkins"
}

// === 私有辅助方法 ===

// createSession 创建 Jenkins 会话（带 Cookie 支持）
func (j *JenkinsAdapter) createSession(baseURL string, authConfig map[string]interface{}) (*JenkinsSession, error) {
	// 创建带 Cookie Jar 的 HTTP Client
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建 cookie jar 失败: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	session := &JenkinsSession{
		Client:  client,
		BaseURL: strings.TrimSuffix(baseURL, "/"),
	}

	// 获取 crumb
	crumbURL := session.BaseURL + "/crumbIssuer/api/json"
	logger.Info("尝试获取 Jenkins crumb", zap.String("crumbURL", crumbURL))

	req, err := http.NewRequest("GET", crumbURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 crumb 请求失败: %w", err)
	}

	// 设置 Basic Auth
	if username, ok := authConfig["username"].(string); ok {
		if password, ok := authConfig["password"].(string); ok {
			req.SetBasicAuth(username, password)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取 crumb 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 crumb 失败，状态码: %d", resp.StatusCode)
	}

	var crumbResponse JenkinsCrumb
	crumbBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(crumbBody, &crumbResponse); err != nil {
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

// executeScript 执行 Jenkins Groovy 脚本
func (j *JenkinsAdapter) executeScript(baseURL, username, password, script string) (string, error) {
	// 创建 Cookie Jar 来保存会话信息
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("创建 Cookie Jar 失败: %w", err)
	}

	// 创建共享 Cookie 的 HTTP Client
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
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
	data := "script=" + url.QueryEscape(script)
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

// FetchGroups Jenkins 不支持用户组，返回空列表
func (j *JenkinsAdapter) FetchGroups(baseURL string, authConfig map[string]interface{}) ([]ApplicationGroup, error) {
	return []ApplicationGroup{}, nil
}
