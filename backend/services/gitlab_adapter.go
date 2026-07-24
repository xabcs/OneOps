package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oneops/backend/logger"
	"oneops/backend/models"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// GitLabAdapter GitLab 应用适配器
type GitLabAdapter struct{}

// ValidateConfig 验证 GitLab 配置
func (g *GitLabAdapter) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["token"]; !ok {
		return fmt.Errorf("GitLab 需要 token 配置")
	}
	return nil
}

// FetchUsers 获取 GitLab 用户列表
func (g *GitLabAdapter) FetchUsers(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationUser, error) {
	// 获取认证信息
	token, ok := authConfig["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("GitLab 需要认证 token")
	}

	// 获取端点配置（支持多种类型）
	var getUsersURL string
	if endpoints, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpoints != nil {
		if url, ok := endpoints["getUsers"].(string); ok && url != "" {
			getUsersURL = url
		}
	} else if endpoints, ok := authConfig["endpoints"].(map[string]string); ok && endpoints != nil {
		// 兼容旧的字符串类型
		if url, ok := endpoints["getUsers"]; ok && url != "" {
			getUsersURL = url
		}
	}

	// 使用默认端点
	if getUsersURL == "" {
		getUsersURL = "/api/v4/users"
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + getUsersURL

	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置私有 token 认证
	req.Header.Set("PRIVATE-TOKEN", token)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户列表失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var usersData []map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &usersData); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为 ApplicationUser
	users := make([]models.ApplicationUser, 0, len(usersData))
	for _, userData := range usersData {
		username := ""
		if val, ok := userData["username"].(string); ok {
			username = val
		}

		name := ""
		if val, ok := userData["name"].(string); ok {
			name = val
		}

		email := ""
		if val, ok := userData["email"].(string); ok {
			email = val
		}

		state := "active"
		if val, ok := userData["state"].(string); ok {
			state = val
		}

		users = append(users, models.ApplicationUser{
			Username:    username,
			DisplayName: name,
			Email:       email,
			Status:      state,
		})
	}

	logger.Info("成功获取 GitLab 用户列表",
		zap.Int("userCount", len(users)))

	return users, nil
}

// FetchRoles 获取 GitLab 角色列表
// GitLab 使用访问级别（Access Levels）而不是传统角色
func (g *GitLabAdapter) FetchRoles(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationRole, error) {
	// GitLab 的访问级别是预定义的
	roles := []models.ApplicationRole{
		{
			RoleCode:    "10",
			RoleName:    "Guest",
			RoleType:    "project",
			Description: "访客权限 - 只读访问",
		},
		{
			RoleCode:    "20",
			RoleName:    "Reporter",
			RoleType:    "project",
			Description: "报告者权限 - 可以查看问题和拉取请求",
		},
		{
			RoleCode:    "30",
			RoleName:    "Developer",
			RoleType:    "project",
			Description: "开发者权限 - 可以推送到分支和标签",
		},
		{
			RoleCode:    "40",
			RoleName:    "Maintainer",
			RoleType:    "project",
			Description: "维护者权限 - 可以管理项目资源",
		},
		{
			RoleCode:    "50",
			RoleName:    "Owner",
			RoleType:    "project",
			Description: "所有者权限 - 完全控制项目",
		},
		// 群组级别的角色
		{
			RoleCode:    "10",
			RoleName:    "Guest",
			RoleType:    "group",
			Description: "群组访客权限",
		},
		{
			RoleCode:    "30",
			RoleName:    "Developer",
			RoleType:    "group",
			Description: "群组开发者权限",
		},
		{
			RoleCode:    "40",
			RoleName:    "Maintainer",
			RoleType:    "group",
			Description: "群组维护者权限",
		},
		{
			RoleCode:    "50",
			RoleName:    "Owner",
			RoleType:    "group",
			Description: "群组所有者权限",
		},
	}

	logger.Info("成功获取 GitLab 角色列表",
		zap.Int("roleCount", len(roles)))

	return roles, nil
}

// CreateUser 在 GitLab 中创建用户
func (g *GitLabAdapter) CreateUser(baseURL string, authConfig map[string]interface{}, user *UserCreateRequest) error {
	// 获取认证信息
	token, ok := authConfig["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("GitLab 需要认证 token")
	}

	// 获取端点配置（支持多种类型）
	var createUserURL string
	if endpoints, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpoints != nil {
		if url, ok := endpoints["createUser"].(string); ok && url != "" {
			createUserURL = url
		}
	} else if endpoints, ok := authConfig["endpoints"].(map[string]string); ok && endpoints != nil {
		// 兼容旧的字符串类型
		if url, ok := endpoints["createUser"]; ok && url != "" {
			createUserURL = url
		}
	}

	// 使用默认端点
	if createUserURL == "" {
		createUserURL = "/api/v4/users"
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + createUserURL

	// 构建请求数据
	reqData := map[string]interface{}{
		"email":            user.Email,
		"username":         user.Username,
		"name":             user.FullName,
		"password":         user.Password,
		"skip_confirmation": true, // 跳过邮件确认
	}

	// 可选字段
	if user.Description != "" {
		reqData["bio"] = user.Description
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	// 创建请求
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 设置认证头
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建用户失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	logger.Info("成功在 GitLab 创建用户",
		zap.String("username", user.Username),
		zap.String("email", user.Email))

	return nil
}

// AssignRole 在 GitLab 中为用户分配角色
// GitLab 需要知道是项目还是群组，以及对应的项目/群组 ID
func (g *GitLabAdapter) AssignRole(baseURL string, authConfig map[string]interface{}, username, roleCode string) error {
	// 获取认证信息
	token, ok := authConfig["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("GitLab 需要认证 token")
	}

	// 获取端点配置（支持多种类型）
	var projectID, groupID string
	var hasProject, hasGroup bool
	var endpoints map[string]string

	if endpointsInterface, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpointsInterface != nil {
		// 将 map[string]interface{} 转换为 map[string]string
		endpoints = make(map[string]string)
		for k, v := range endpointsInterface {
			if strVal, ok := v.(string); ok {
				endpoints[k] = strVal
			}
		}
	} else if endpointsStr, ok := authConfig["endpoints"].(map[string]string); ok && endpointsStr != nil {
		// 兼容旧的字符串类型
		endpoints = endpointsStr
	}

	// 从配置中获取项目/群组信息
	if projectID = endpoints["projectId"]; projectID != "" {
		hasProject = true
	}
	if groupID = endpoints["groupId"]; groupID != "" {
		hasGroup = true
	}

	var fullURL string
	var reqData map[string]interface{}

	if hasProject {
		// 项目级别的权限分配
		assignRoleURL, ok := endpoints["assignProjectRole"]
		if !ok || assignRoleURL == "" {
			assignRoleURL = "/api/v4/projects/:id/members"
		}

		baseURL = strings.TrimSuffix(baseURL, "/")
		assignRoleURL = strings.ReplaceAll(assignRoleURL, ":id", projectID)
		fullURL = baseURL + assignRoleURL

		// 获取用户 ID（GitLab 需要用户 ID 而不是用户名）
		userID, err := g.getUserIDByUsername(baseURL, token, username)
		if err != nil {
			return fmt.Errorf("获取用户 ID 失败: %w", err)
		}

		// 解析访问级别
		accessLevel, err := strconv.Atoi(roleCode)
		if err != nil {
			return fmt.Errorf("无效的角色代码: %s", roleCode)
		}

		reqData = map[string]interface{}{
			"user_id":      userID,
			"access_level": accessLevel,
		}

	} else if hasGroup {
		// 群组级别的权限分配
		assignRoleURL, ok := endpoints["assignGroupRole"]
		if !ok || assignRoleURL == "" {
			assignRoleURL = "/api/v4/groups/:id/members"
		}

		baseURL = strings.TrimSuffix(baseURL, "/")
		assignRoleURL = strings.ReplaceAll(assignRoleURL, ":id", groupID)
		fullURL = baseURL + assignRoleURL

		// 获取用户 ID
		userID, err := g.getUserIDByUsername(baseURL, token, username)
		if err != nil {
			return fmt.Errorf("获取用户 ID 失败: %w", err)
		}

		// 解析访问级别
		accessLevel, err := strconv.Atoi(roleCode)
		if err != nil {
			return fmt.Errorf("无效的角色代码: %s", roleCode)
		}

		reqData = map[string]interface{}{
			"user_id":      userID,
			"access_level": accessLevel,
		}
	} else {
		return fmt.Errorf("需要指定项目 ID (projectId) 或群组 ID (groupId) 配置")
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	// 创建请求
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 设置认证头
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("分配角色失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	logger.Info("成功在 GitLab 为用户分配角色",
		zap.String("username", username),
		zap.String("roleCode", roleCode))

	return nil
}

// GetConfigTemplate 获取 GitLab 配置模板
func (g *GitLabAdapter) GetConfigTemplate() map[string]interface{} {
	return map[string]interface{}{
		"token": "",
		"syncInterval": 300,
		"endpoints": map[string]string{
			"getUsers":        "/api/v4/users",
			"getGroups":       "/api/v4/groups",
			"getRoles":        "/api/v4/users", // GitLab 角色是预定义的，不需要从 API 获取
			"createUser":      "/api/v4/users",
			"assignProjectRole": "/api/v4/projects/:id/members",
			"assignGroupRole":   "/api/v4/groups/:id/members",
			"projectId": "", // 需要用户配置项目 ID
			"groupId":   "", // 或群组 ID
		},
	}
}

// GetDisplayName 获取 GitLab 显示名称
func (g *GitLabAdapter) GetDisplayName() string {
	return "GitLab"
}

// getUserIDByUsername 根据用户名获取 GitLab 用户 ID
func (g *GitLabAdapter) getUserIDByUsername(baseURL, token, username string) (int, error) {
	fullURL := baseURL + "/api/v4/users?username=" + username

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("获取用户失败 (状态码 %d)", resp.StatusCode)
	}

	var usersData []map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &usersData); err != nil {
		return 0, err
	}

	if len(usersData) == 0 {
		return 0, fmt.Errorf("用户不存在: %s", username)
	}

	userID := int(usersData[0]["id"].(float64))
	return userID, nil
}

// FetchGroups 获取 GitLab 用户组列表
// GitLab 有 Group 概念，可以通过 API 获取
func (g *GitLabAdapter) FetchGroups(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationGroup, error) {
	// 获取认证信息
	token, ok := authConfig["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("GitLab 需要认证 token")
	}

	// 获取端点配置（支持多种类型）
	var getGroupsURL string
	if endpoints, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpoints != nil {
		if url, ok := endpoints["getGroups"].(string); ok && url != "" {
			getGroupsURL = url
		}
	} else if endpoints, ok := authConfig["endpoints"].(map[string]string); ok && endpoints != nil {
		if url, ok := endpoints["getGroups"]; ok && url != "" {
			getGroupsURL = url
		}
	}

	// 使用默认端点
	if getGroupsURL == "" {
		getGroupsURL = "/api/v4/groups"
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + getGroupsURL

	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置认证头
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户组列表失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var groups []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Path        string `json:"path"`
		Description string `json:"description"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为 ApplicationGroup
	result := make([]models.ApplicationGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, models.ApplicationGroup{
			GroupCode:   fmt.Sprintf("%d", group.ID),
			GroupName:   group.Name,
			Description: group.Description,
		})
	}

	logger.Info("成功获取 GitLab 用户组列表",
		zap.Int("groupCount", len(result)))

	return result, nil
}
