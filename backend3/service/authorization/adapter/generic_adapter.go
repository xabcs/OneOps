package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	modelauth "oneops/backend3/model/authorization"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Username    string
	Password    string
	FullName    string
	Email       string
	Description string
}

// GenericAdapter 通用 REST API 应用适配器
type GenericAdapter struct{}

// ValidateConfig 验证通用应用配置
func (g *GenericAdapter) ValidateConfig(config map[string]interface{}) error {
	// 通用适配器配置灵活，不强制要求特定字段
	return nil
}

// FetchUsers 从通用 API 获取用户列表
func (g *GenericAdapter) FetchUsers(baseURL string, authConfig map[string]interface{}) ([]modelauth.ApplicationUser, error) {
	// 获取 endpoints 配置（支持多种类型）
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

	if getUsersURL == "" {
		return nil, fmt.Errorf("缺少 getUsers 端点配置")
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + getUsersURL

	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加认证头
	if token, ok := authConfig["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 支持 Basic Auth
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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户列表失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Data []modelauth.ApplicationUser `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析用户列表失败: %w", err)
	}

	return result.Data, nil
}

// FetchRoles 从通用 API 获取角色列表
func (g *GenericAdapter) FetchRoles(baseURL string, authConfig map[string]interface{}) ([]modelauth.ApplicationRole, error) {
	// 获取 endpoints 配置（支持多种类型）
	var getRolesURL string
	if endpoints, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpoints != nil {
		if url, ok := endpoints["getRoles"].(string); ok && url != "" {
			getRolesURL = url
		}
	} else if endpoints, ok := authConfig["endpoints"].(map[string]string); ok && endpoints != nil {
		// 兼容旧的字符串类型
		if url, ok := endpoints["getRoles"]; ok && url != "" {
			getRolesURL = url
		}
	}

	if getRolesURL == "" {
		return nil, fmt.Errorf("缺少 getRoles 端点配置")
	}

	// 构建完整 URL
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

	// 支持 Basic Auth
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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取角色列表失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Data []modelauth.ApplicationRole `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析角色列表失败: %w", err)
	}

	return result.Data, nil
}

// CreateUser 在通用 API 中创建用户
func (g *GenericAdapter) CreateUser(baseURL string, authConfig map[string]interface{}, user *UserCreateRequest) (string, error) {
	// 获取 endpoints 配置（支持多种类型）
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

	if createUserURL == "" {
		return "", fmt.Errorf("缺少 createUser 端点配置")
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + createUserURL

	// 构建请求数据
	reqData := map[string]interface{}{
		"username": user.Username,
		"password": user.Password,
	}

	if user.Email != "" {
		reqData["email"] = user.Email
	}
	if user.FullName != "" {
		reqData["fullName"] = user.FullName
		reqData["displayName"] = user.FullName
	}
	if user.Description != "" {
		reqData["description"] = user.Description
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	// 创建请求
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加认证头
	if token, ok := authConfig["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 支持 Basic Auth
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
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("创建用户失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 尝试解析响应以获取用户 ID
	var result struct {
		Data struct {
			ID       interface{} `json:"id"`
			Username string      `json:"username"`
		} `json:"data"`
	}

	userID := ""
	if err := json.Unmarshal(body, &result); err == nil {
		if result.Data.ID != nil {
			userID = fmt.Sprintf("%v", result.Data.ID)
		}
	}

	logger.Info("在通用系统创建用户成功",
		zap.String("username", user.Username),
		zap.String("user_id", userID))

	return userID, nil
}

// AssignRole 在通用 API 中为用户分配角色
func (g *GenericAdapter) AssignRole(baseURL string, authConfig map[string]interface{}, username, roleCode string) error {
	// 获取 endpoints 配置（支持多种类型）
	var assignRoleURL string
	if endpoints, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpoints != nil {
		if url, ok := endpoints["assignRole"].(string); ok && url != "" {
			assignRoleURL = url
		}
	} else if endpoints, ok := authConfig["endpoints"].(map[string]string); ok && endpoints != nil {
		// 兼容旧的字符串类型
		if url, ok := endpoints["assignRole"]; ok && url != "" {
			assignRoleURL = url
		}
	}

	if assignRoleURL == "" {
		return fmt.Errorf("缺少 assignRole 端点配置")
	}

	// 构建完整 URL，替换 {username} 占位符
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + strings.ReplaceAll(assignRoleURL, "{username}", username)

	// 构建请求数据
	reqData := map[string]interface{}{
		"username": username,
		"roleCode": roleCode,
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

	req.Header.Set("Content-Type", "application/json")

	// 添加认证头
	if token, ok := authConfig["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 支持 Basic Auth
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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("分配角色失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	logger.Info("在通用系统为用户分配角色成功",
		zap.String("username", username),
		zap.String("role_code", roleCode))

	return nil
}

// GetConfigTemplate 获取通用应用配置模板
func (g *GenericAdapter) GetConfigTemplate() map[string]interface{} {
	return map[string]interface{}{
		"authType":     "token",
		"token":        "",
		"syncInterval": 300,
		"endpoints": map[string]string{
			"getUsers":   "/api/users",
			"getRoles":   "/api/roles",
			"createUser": "/api/users",
			"assignRole": "/api/users/{username}/roles",
		},
	}
}

// GetDisplayName 获取通用应用显示名称
func (g *GenericAdapter) GetDisplayName() string {
	return "通用应用 (REST API)"
}

// FetchGroups 通用应用不支持用户组，返回空列表
func (g *GenericAdapter) FetchGroups(baseURL string, authConfig map[string]interface{}) ([]modelauth.ApplicationGroup, error) {
	return []modelauth.ApplicationGroup{}, nil
}
