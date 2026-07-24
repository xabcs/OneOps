package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"oneops/backend/logger"
	"oneops/backend/models"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/twindagger/httpsig.v1"
)

// JumpserverAdapter JumpServer 应用适配器
type JumpserverAdapter struct{}

// ValidateConfig 验证 JumpServer 配置
func (j *JumpserverAdapter) ValidateConfig(config map[string]interface{}) error {
	// 检查是否至少有一种认证方式
	hasAccessKey := false
	hasUsername := false
	hasToken := false

	// AccessKey + Secret 认证
	if accessKey, ok := config["accessKey"].(string); ok && accessKey != "" {
		secret, _ := config["secret"].(string)
		if secret == "" {
			return fmt.Errorf("AccessKey 认证需要提供 secret")
		}
		hasAccessKey = true
	}

	// 用户名密码认证
	if username, ok := config["username"].(string); ok && username != "" {
		password, _ := config["password"].(string)
		if password == "" {
			return fmt.Errorf("用户名密码认证需要提供 password")
		}
		hasUsername = true
	}

	// Token 认证
	if token, ok := config["token"].(string); ok && token != "" {
		hasToken = true
	}

	if !hasAccessKey && !hasUsername && !hasToken {
		return fmt.Errorf("JumpServer 需要认证配置：AccessKey+Secret、Private Token 或 用户名+密码")
	}

	return nil
}

// setAuthentication 设置认证方式
// JumpServer 支持三种认证方式：
// 1. AccessKey + Secret（推荐）: authConfig["accessKey"] + authConfig["secret"] - 使用签名认证
// 2. Private Token: authConfig["token"] - 长期有效
// 3. 用户名 + 密码: authConfig["username"] + authConfig["password"]
func (j *JumpserverAdapter) setAuthentication(req *http.Request, authConfig map[string]interface{}) error {
	// 方式1: AccessKey + Secret（签名认证，推荐）
	if accessKey, ok := authConfig["accessKey"].(string); ok && accessKey != "" {
		secret, _ := authConfig["secret"].(string)
		if secret == "" {
			return fmt.Errorf("AccessKey 认证需要提供 secret")
		}

		// 设置必需的 headers（必须严格按照顺序）
		gmtFmt := "Mon, 02 Jan 2006 15:04:05 GMT"
		req.Header.Set("Date", time.Now().Format(gmtFmt))
		req.Header.Set("Accept", "application/json")

		// 设置组织ID（默认组织）
		orgID := "00000000-0000-0000-0000-000000000002"
		if customOrgID, ok := authConfig["orgId"].(string); ok && customOrgID != "" {
			orgID = customOrgID
		}
		req.Header.Set("X-JMS-ORG", orgID)

		// 记录请求信息（签名前）
		logger.Info("准备签名请求",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.String("date", req.Header.Get("Date")),
			zap.String("x-jms-org", req.Header.Get("X-JMS-ORG")))

		// 执行签名
		if err := j.signRequest(req, accessKey, secret); err != nil {
			return fmt.Errorf("签名失败: %w", err)
		}

		logger.Info("使用 AccessKey 签名认证", zap.String("accessKey", accessKey))
		return nil
	}

	// 方式2: 用户名 + 密码（基础认证）
	if username, ok := authConfig["username"].(string); ok && username != "" {
		password, _ := authConfig["password"].(string)
		if password == "" {
			return fmt.Errorf("用户名密码认证需要提供 password")
		}
		req.SetBasicAuth(username, password)
		logger.Info("使用 Basic Auth 认证", zap.String("username", username))
		return nil
	}

	// 方式3: Private Token（默认）
	if token, ok := authConfig["token"].(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		logger.Info("使用 Private Token 认证")
		return nil
	}

	return fmt.Errorf("JumpServer 需要认证配置：AccessKey+Secret、Private Token 或 用户名+密码")
}

// signRequest 对请求进行 HTTP Signature 签名
// 使用 httpsig 库确保符合 JumpServer 的签名规范
func (j *JumpserverAdapter) signRequest(req *http.Request, keyID, secret string) error {
	// 定义需要签名的 header 列表
	// JumpServer 要求包含 (request-target) 和 date
	headers := []string{"(request-target)", "date"}

	// 创建签名器
	signer, err := httpsig.NewRequestSigner(keyID, secret, "hmac-sha256")
	if err != nil {
		return fmt.Errorf("创建签名器失败: %w", err)
	}

	// 执行签名
	if err := signer.SignRequest(req, headers, nil); err != nil {
		return fmt.Errorf("签名请求失败: %w", err)
	}

	// 调试日志：打印签名结果
	logger.Info("JumpServer 签名完成",
		zap.String("keyId", keyID),
		zap.String("authorization", req.Header.Get("Authorization")))

	return nil
}

// FetchUsers 获取 JumpServer 用户列表
func (j *JumpserverAdapter) FetchUsers(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationUser, error) {
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

	// 如果没有配置，使用默认端点
	if getUsersURL == "" {
		getUsersURL = "/api/v1/users/users/"
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + getUsersURL

	// 创建请求（必须先设置 URL，再进行签名认证）
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置认证（签名依赖于 URL）
	if err := j.setAuthentication(req, authConfig); err != nil {
		return nil, err
	}

	// 记录详细的请求信息
	logger.Info("准备请求 JumpServer 用户列表 API",
		zap.String("baseURL", baseURL),
		zap.String("endpoint", getUsersURL),
		zap.String("fullURL", fullURL))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 记录响应详情
	logger.Info("JumpServer API 响应",
		zap.String("url", fullURL),
		zap.Int("statusCode", resp.StatusCode),
		zap.String("status", resp.Status))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("JumpServer API 请求失败",
			zap.String("url", fullURL),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("获取用户列表失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应 - JumpServer 返回的是直接的数组，不是包装对象
	// 期望格式: [{"id":"...", "username":"...", ...}, ...]
	var apiUsers []struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		IsActive bool   `json:"is_active"`
	}

	body, _ := io.ReadAll(resp.Body)
	logger.Info("JumpServer API 响应体",
		zap.String("url", fullURL),
		zap.Int("bodyLength", len(body)),
		zap.String("responseBody", string(body)))

	if err := json.Unmarshal(body, &apiUsers); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	logger.Info("JumpServer 用户解析结果",
		zap.Int("用户数量", len(apiUsers)))

	// 转换为 ApplicationUser
	users := make([]models.ApplicationUser, 0, len(apiUsers))
	for _, user := range apiUsers {
		status := "active"
		if !user.IsActive {
			status = "inactive"
		}

		users = append(users, models.ApplicationUser{
			Username:    user.Username,
			DisplayName: user.Name,
			Email:       user.Email,
			Status:      status,
		})
	}

	logger.Info("成功获取 JumpServer 用户列表",
		zap.Int("userCount", len(users)))

	return users, nil
}

// FetchGroups 获取 JumpServer 用户组列表
// 注意：JumpServer 不同版本的用户组 API 端点可能不同
func (j *JumpserverAdapter) FetchGroups(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationGroup, error) {
	// 获取端点配置（支持多种类型）
	var getGroupsURL string
	if endpoints, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpoints != nil {
		if url, ok := endpoints["getGroups"].(string); ok && url != "" {
			getGroupsURL = url
		}
	} else if endpoints, ok := authConfig["endpoints"].(map[string]string); ok && endpoints != nil {
		// 兼容旧的字符串类型
		if url, ok := endpoints["getGroups"]; ok && url != "" {
			getGroupsURL = url
		}
	}

	// 如果没有配置，尝试常见的端点
	if getGroupsURL == "" {
		// JumpServer v3 常见的用户组端点
		getGroupsURL = "/api/v1/users/groups/"
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + getGroupsURL

	logger.Info("正在请求 JumpServer 用户组 API",
		zap.String("url", fullURL))

	// 创建请求（必须先设置 URL，再进行签名认证）
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置认证（支持 AccessKey、Token、BasicAuth 等多种方式）
	if err := j.setAuthentication(req, authConfig); err != nil {
		return nil, err
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// JumpServer 某些版本可能没有用户组 API
		// 返回空列表而不是报错，让用户可以继续使用其他功能
		logger.Warn("JumpServer 用户组 API 端点不存在，该版本可能不支持用户组功能或端点配置错误",
			zap.String("url", fullURL),
			zap.String("suggestion", "请在应用配置中设置正确的端点，如: /api/v1/users/groups/ 或 /api/v1/assets/groups/"))

		return []models.ApplicationGroup{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取用户组列表失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应 - JumpServer 可能返回直接的数组或包装对象
	var apiGroups []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Comment string `json:"comment"`
	}

	body, _ := io.ReadAll(resp.Body)
	logger.Info("JumpServer 用户组 API 响应体",
		zap.String("url", fullURL),
		zap.Int("bodyLength", len(body)),
		zap.String("responseBody", string(body)))

	if err := json.Unmarshal(body, &apiGroups); err != nil {
		// 如果直接解析数组失败，尝试解析包装对象
		logger.Warn("尝试解析包装对象格式",
			zap.String("url", fullURL),
			zap.Error(err))

		var wrappedResult struct {
			Data []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Comment string `json:"comment"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &wrappedResult); err != nil {
			return nil, fmt.Errorf("解析响应失败（尝试了数组和包装对象格式）: %w", err)
		}

		apiGroups = make([]struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Comment string `json:"comment"`
		}, 0, len(wrappedResult.Data))

		for _, group := range wrappedResult.Data {
			apiGroups = append(apiGroups, group)
		}
	}

	logger.Info("JumpServer 用户组解析结果",
		zap.Int("groupCount", len(apiGroups)))

	// 转换为 ApplicationGroup
	groups := make([]models.ApplicationGroup, 0, len(apiGroups))
	for _, group := range apiGroups {
		groups = append(groups, models.ApplicationGroup{
			GroupCode:   group.ID,
			GroupName:   group.Name,
			Description: group.Comment,
		})
	}

	logger.Info("成功获取 JumpServer 用户组列表",
		zap.Int("groupCount", len(groups)))

	return groups, nil
}

// FetchRoles 获取 JumpServer 角色列表
// JumpServer 使用内置角色系统，角色是固定的，不需要从 API 同步
// 返回空列表，提示用户 JumpServer 不需要角色同步
func (j *JumpserverAdapter) FetchRoles(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationRole, error) {
	// JumpServer 的授权是通过授权规则（用户/用户组 -> 资产/节点）
	// 不需要同步角色，返回空列表
	logger.Info("JumpServer 不需要角色同步，授权通过授权规则管理")

	return []models.ApplicationRole{}, nil
}

// CreateUser 在 JumpServer 中创建用户
func (j *JumpserverAdapter) CreateUser(baseURL string, authConfig map[string]interface{}, user *UserCreateRequest) error {
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

	// 如果没有配置，使用默认端点
	if createUserURL == "" {
		createUserURL = "/api/v1/users/users/"
	}

	// 构建完整 URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	fullURL := baseURL + createUserURL

	// 构建请求数据
	reqData := map[string]interface{}{
		"username":  user.Username,
		"name":      user.FullName,
		"email":     user.Email,
		"is_active": true,
	}

	// Jumpserver 密码处理（如果 API 支持）
	if user.Password != "" {
		reqData["password"] = user.Password
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	// 创建请求（必须先设置 URL，再进行签名认证）
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 设置认证（支持 AccessKey、Token、BasicAuth 等多种方式）
	if err := j.setAuthentication(req, authConfig); err != nil {
		return err
	}

	// 设置 Content-Type
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

	logger.Info("成功在 JumpServer 创建用户",
		zap.String("username", user.Username))

	return nil
}

// AssignRole 在 JumpServer 中为用户分配角色
// JumpServer 的授权机制与传统 RBAC 不同，不支持通过 API 直接分配角色
// JumpServer 通过授权规则来管理用户对资产/节点的访问权限
func (j *JumpserverAdapter) AssignRole(baseURL string, authConfig map[string]interface{}, username, roleCode string) error {
	// JumpServer 不支持通过 API 分配角色
	// 需要在 JumpServer 管理界面中创建授权规则（用户/用户组 -> 资产/节点）
	return fmt.Errorf("JumpServer 不支持角色分配，请在 JumpServer 管理界面中创建授权规则来管理用户权限")
}

// GetConfigTemplate 获取 JumpServer 配置模板
func (j *JumpserverAdapter) GetConfigTemplate() map[string]interface{} {
	return map[string]interface{}{
		"accessKey":    "",
		"secret":       "",
		"orgId":        "00000000-0000-0000-0000-000000000002",
		"syncInterval": 300,
		"endpoints": map[string]string{
			"getUsers":   "/api/v1/users/users/",
			"getGroups":  "/api/v1/users/groups/",
			"getRules":   "/api/v1/perms/asset-permissions/",
			"createUser": "/api/v1/users/users/",
		},
	}
}

// GetDisplayName 获取 JumpServer 显示名称
func (j *JumpserverAdapter) GetDisplayName() string {
	return "JumpServer"
}

// FetchAuthorizationRules 获取 JumpServer 授权规则列表
func (j *JumpserverAdapter) FetchAuthorizationRules(baseURL string, authConfig map[string]interface{}) ([]models.ApplicationAuthorizationRule, error) {
	// 常见的 JumpServer 授权规则端点（按优先级排序）
	possibleEndpoints := []string{
		"/api/v1/perms/asset-permissions/", // 当前配置
		"/api/v1/perms/permissions/",       // 替代端点1
		"/api/v1/assets/perms/",            // 替代端点2
		"/api/v1/perms/user-permissions/",  // 替代端点3
	}

	// 获取用户配置的端点（如果有）
	var configuredEndpoint string
	if endpoints, ok := authConfig["endpoints"].(map[string]interface{}); ok && endpoints != nil {
		if url, ok := endpoints["getRules"].(string); ok && url != "" {
			configuredEndpoint = url
		}
	} else if endpoints, ok := authConfig["endpoints"].(map[string]string); ok && endpoints != nil {
		if url, ok := endpoints["getRules"]; ok && url != "" {
			configuredEndpoint = url
		}
	}

	// 如果用户配置了端点，优先使用；否则尝试常见端点
	endpointsToTry := []string{}
	if configuredEndpoint != "" {
		endpointsToTry = []string{configuredEndpoint}
	} else {
		endpointsToTry = possibleEndpoints
	}

	baseURL = strings.TrimSuffix(baseURL, "/")

	// 尝试每个端点，直到找到有效的
	for _, getRulesURL := range endpointsToTry {
		fullURL := baseURL + getRulesURL

		logger.Info("正在请求 JumpServer 授权规则 API",
			zap.String("url", fullURL),
			zap.String("endpoint", getRulesURL),
			zap.String("configuredEndpoint", configuredEndpoint))

		// 创建请求（必须先设置 URL，再进行签名认证）
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			logger.Warn("创建请求失败", zap.Error(err))
			continue
		}

		// 设置认证（支持 AccessKey、Token、BasicAuth 等多种方式）
		if err := j.setAuthentication(req, authConfig); err != nil {
			logger.Warn("设置认证失败", zap.Error(err))
			continue
		}

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logger.Warn("请求失败", zap.String("url", fullURL), zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		// 读取响应体（所有情况下都读取，便于调试）
		body, _ := io.ReadAll(resp.Body)

		logger.Info("JumpServer 授权规则 API 响应",
			zap.String("url", fullURL),
			zap.Int("statusCode", resp.StatusCode),
			zap.Int("bodyLength", len(body)),
			zap.String("responseBody", string(body)[:min(len(body), 2000)])) // 只记录前2000字符

		if resp.StatusCode == http.StatusOK {
			// 找到有效的端点，开始解析数据
			rules, err := j.parseAuthorizationRules(body, fullURL)
			if err != nil {
				logger.Warn("解析授权规则失败，尝试下一个端点",
					zap.String("url", fullURL),
					zap.Error(err))
				continue
			}

			logger.Info("成功获取 JumpServer 授权规则",
				zap.String("workingEndpoint", getRulesURL),
				zap.Int("ruleCount", len(rules)))

			return rules, nil
		}

		// 如果不是 404，说明端点可能存在但有问题
		if resp.StatusCode != http.StatusNotFound {
			logger.Warn("端点返回非200状态码",
				zap.String("url", fullURL),
				zap.Int("statusCode", resp.StatusCode),
				zap.String("response", string(body)))
		}
	}

	// 所有端点都失败了
	logger.Warn("所有授权规则 API 端点都失败，JumpServer 可能不支持授权规则同步或 AccessKey 没有相应权限")
	return []models.ApplicationAuthorizationRule{}, nil
}

// parseAuthorizationRules 解析授权规则响应数据
func (j *JumpserverAdapter) parseAuthorizationRules(body []byte, url string) ([]models.ApplicationAuthorizationRule, error) {
	logger.Info("开始解析授权规则响应",
		zap.String("url", url),
		zap.Int("bodyLength", len(body)),
		zap.String("responseBody", string(body)))

	// JumpServer 实际返回的格式（统计数据格式）
	var apiRules []struct {
		ID               string   `json:"id"`
		Name             string   `json:"name"`
		Labels           []string `json:"labels"`
		UsersAmount      int      `json:"users_amount"`
		UserGroupsAmount int      `json:"user_groups_amount"`
		AssetsAmount     int      `json:"assets_amount"`
		NodesAmount      int      `json:"nodes_amount"`
		Accounts         []string `json:"accounts"`
		Protocols        []string `json:"protocols"`
		Actions          []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"actions"`
		OrgID       string `json:"org_id"`
		OrgName     string `json:"org_name"`
		IsActive    bool   `json:"is_active"`
		IsExpired   bool   `json:"is_expired"`
		IsValid     bool   `json:"is_valid"`
		FromTicket  bool   `json:"from_ticket"`
		DateCreated string `json:"date_created"`
		DateStart   string `json:"date_start"`
		DateExpired string `json:"date_expired"`
		CreatedBy   string `json:"created_by"`
		Comment     string `json:"comment"`
	}

	if err := json.Unmarshal(body, &apiRules); err != nil {
		logger.Warn("尝试解析包装对象格式",
			zap.Error(err))

		// 尝试2: 解析为包装对象（使用统计格式）
		var wrappedResult struct {
			Data []struct {
				ID               string   `json:"id"`
				Name             string   `json:"name"`
				Labels           []string `json:"labels"`
				UsersAmount      int      `json:"users_amount"`
				UserGroupsAmount int      `json:"user_groups_amount"`
				AssetsAmount     int      `json:"assets_amount"`
				NodesAmount      int      `json:"nodes_amount"`
				Accounts         []string `json:"accounts"`
				Protocols        []string `json:"protocols"`
				Actions          []struct {
					Value string `json:"value"`
					Label string `json:"label"`
				} `json:"actions"`
				OrgID       string `json:"org_id"`
				OrgName     string `json:"org_name"`
				IsActive    bool   `json:"is_active"`
				IsExpired   bool   `json:"is_expired"`
				IsValid     bool   `json:"is_valid"`
				FromTicket  bool   `json:"from_ticket"`
				DateCreated string `json:"date_created"`
				DateStart   string `json:"date_start"`
				DateExpired string `json:"date_expired"`
				CreatedBy   string `json:"created_by"`
				Comment     string `json:"comment"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &wrappedResult); err != nil {
			// 尝试3: 检查是否是分页响应（使用统计格式）
			var paginatedResult struct {
				Count    int    `json:"count"`
				Next     string `json:"next"`
				Previous string `json:"previous"`
				Results  []struct {
					ID               string   `json:"id"`
					Name             string   `json:"name"`
					Labels           []string `json:"labels"`
					UsersAmount      int      `json:"users_amount"`
					UserGroupsAmount int      `json:"user_groups_amount"`
					AssetsAmount     int      `json:"assets_amount"`
					NodesAmount      int      `json:"nodes_amount"`
					Accounts         []string `json:"accounts"`
					Protocols        []string `json:"protocols"`
					Actions          []struct {
						Value string `json:"value"`
						Label string `json:"label"`
					} `json:"actions"`
					OrgID       string `json:"org_id"`
					OrgName     string `json:"org_name"`
					IsActive    bool   `json:"is_active"`
					IsExpired   bool   `json:"is_expired"`
					IsValid     bool   `json:"is_valid"`
					FromTicket  bool   `json:"from_ticket"`
					DateCreated string `json:"date_created"`
					DateStart   string `json:"date_start"`
					DateExpired string `json:"date_expired"`
					CreatedBy   string `json:"created_by"`
					Comment     string `json:"comment"`
				} `json:"results"`
			}

			if err := json.Unmarshal(body, &paginatedResult); err != nil {
				return nil, fmt.Errorf("解析授权规则响应失败（尝试了数组、包装对象和分页格式）: %w", err)
			}

			logger.Info("检测到分页响应",
				zap.Int("totalCount", paginatedResult.Count))

			apiRules = make([]struct {
				ID               string   `json:"id"`
				Name             string   `json:"name"`
				Labels           []string `json:"labels"`
				UsersAmount      int      `json:"users_amount"`
				UserGroupsAmount int      `json:"user_groups_amount"`
				AssetsAmount     int      `json:"assets_amount"`
				NodesAmount      int      `json:"nodes_amount"`
				Accounts         []string `json:"accounts"`
				Protocols        []string `json:"protocols"`
				Actions          []struct {
					Value string `json:"value"`
					Label string `json:"label"`
				} `json:"actions"`
				OrgID       string `json:"org_id"`
				OrgName     string `json:"org_name"`
				IsActive    bool   `json:"is_active"`
				IsExpired   bool   `json:"is_expired"`
				IsValid     bool   `json:"is_valid"`
				FromTicket  bool   `json:"from_ticket"`
				DateCreated string `json:"date_created"`
				DateStart   string `json:"date_start"`
				DateExpired string `json:"date_expired"`
				CreatedBy   string `json:"created_by"`
				Comment     string `json:"comment"`
			}, 0, len(paginatedResult.Results))

			for _, rule := range paginatedResult.Results {
				apiRules = append(apiRules, rule)
			}
		} else {
			apiRules = make([]struct {
				ID               string   `json:"id"`
				Name             string   `json:"name"`
				Labels           []string `json:"labels"`
				UsersAmount      int      `json:"users_amount"`
				UserGroupsAmount int      `json:"user_groups_amount"`
				AssetsAmount     int      `json:"assets_amount"`
				NodesAmount      int      `json:"nodes_amount"`
				Accounts         []string `json:"accounts"`
				Protocols        []string `json:"protocols"`
				Actions          []struct {
					Value string `json:"value"`
					Label string `json:"label"`
				} `json:"actions"`
				OrgID       string `json:"org_id"`
				OrgName     string `json:"org_name"`
				IsActive    bool   `json:"is_active"`
				IsExpired   bool   `json:"is_expired"`
				IsValid     bool   `json:"is_valid"`
				FromTicket  bool   `json:"from_ticket"`
				DateCreated string `json:"date_created"`
				DateStart   string `json:"date_start"`
				DateExpired string `json:"date_expired"`
				CreatedBy   string `json:"created_by"`
				Comment     string `json:"comment"`
			}, 0, len(wrappedResult.Data))

			for _, rule := range wrappedResult.Data {
				apiRules = append(apiRules, rule)
			}
		}
	}

	logger.Info("JumpServer 授权规则解析结果",
		zap.Int("规则数量", len(apiRules)))

	// 转换为 ApplicationAuthorizationRule
	rules := make([]models.ApplicationAuthorizationRule, 0)
	for i, apiRule := range apiRules {
		logger.Info("开始转换授权规则",
			zap.Int("index", i),
			zap.String("ruleId", apiRule.ID),
			zap.String("ruleName", apiRule.Name),
			zap.Int("usersAmount", apiRule.UsersAmount),
			zap.Int("userGroupsAmount", apiRule.UserGroupsAmount))
		// JumpServer 返回的是统计信息，不是具体的用户-资产映射
		// 因此为每个授权规则创建一条通用记录

		// 提取 actions 的 value 字段
		actions := make([]string, 0, len(apiRule.Actions))
		for _, action := range apiRule.Actions {
			actions = append(actions, action.Value)
		}
		actionsJSON, _ := json.Marshal(actions)

		// 解析过期时间
		var expireTime *time.Time
		if apiRule.DateExpired != "" {
			// JumpServer 日期格式: "2026/01/13 12:20:59 +0800"
			// 去掉时区部分，只保留日期时间
			dateStr := apiRule.DateExpired
			if len(dateStr) > 19 {
				dateStr = dateStr[:19] // 去掉时区信息
			}
			parsedTime, err := time.Parse("2006/01/02 15:04:05", dateStr)
			if err == nil {
				expireTime = &parsedTime
			}
		}

		// 确定规则类型（基于统计信息）
		ruleType := "system"
		subjectType := "system"
		subjectID := "*"
		subjectName := "系统授权"

		if apiRule.UsersAmount > 0 {
			ruleType = "user"
			subjectType = "user"
			subjectID = fmt.Sprintf("%d_users", apiRule.UsersAmount)
			subjectName = fmt.Sprintf("%d 个用户", apiRule.UsersAmount)
		} else if apiRule.UserGroupsAmount > 0 {
			ruleType = "group"
			subjectType = "group"
			subjectID = fmt.Sprintf("%d_groups", apiRule.UserGroupsAmount)
			subjectName = fmt.Sprintf("%d 个用户组", apiRule.UserGroupsAmount)
		}

		// 确定对象类型
		objectType := "system"
		objectID := "*"
		objectName := "所有资产"

		if apiRule.AssetsAmount > 0 && apiRule.NodesAmount > 0 {
			objectType = "mixed"
			objectID = fmt.Sprintf("%d_assets_%d_nodes", apiRule.AssetsAmount, apiRule.NodesAmount)
			objectName = fmt.Sprintf("%d 个资产 + %d 个节点", apiRule.AssetsAmount, apiRule.NodesAmount)
		} else if apiRule.AssetsAmount > 0 {
			objectType = "asset"
			objectID = fmt.Sprintf("%d_assets", apiRule.AssetsAmount)
			objectName = fmt.Sprintf("%d 个资产", apiRule.AssetsAmount)
		} else if apiRule.NodesAmount > 0 {
			objectType = "node"
			objectID = fmt.Sprintf("%d_nodes", apiRule.NodesAmount)
			objectName = fmt.Sprintf("%d 个节点", apiRule.NodesAmount)
		}

		// 构建描述信息（用于调试，不存储到数据库）
		_ = apiRule.Comment // 注释信息暂不存储

		rules = append(rules, models.ApplicationAuthorizationRule{
			RuleID:      apiRule.ID,
			RuleName:    apiRule.Name,
			RuleType:    ruleType,
			SubjectType: subjectType,
			SubjectID:   subjectID,
			SubjectName: subjectName,
			ObjectType:  objectType,
			ObjectID:    objectID,
			ObjectName:  objectName,
			Actions:     string(actionsJSON),
			Priority:    0, // JumpServer 没有提供 priority 字段
			IsEnabled:   apiRule.IsActive,
			IsExpired:   apiRule.IsExpired,
			ExpireTime:  expireTime,
		})
	}

	logger.Info("授权规则解析完成",
		zap.Int("ruleCount", len(rules)))

	return rules, nil
}

// min 辅助函数：返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
