package services

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
	"oneops/backend/models"
)

// CasbinAPIManager 基于Casbin的API管理器
type CasbinAPIManager struct {
	db       *gorm.DB
	enforcer *casbin.Enforcer
}

// NewCasbinAPIManager 创建基于Casbin的API管理器
func NewCasbinAPIManager() (*CasbinAPIManager, error) {
	db := GetDB()

	// 初始化 Casbin GORM 适配器
	adapter, err := gormadapter.NewAdapterByDBUsePrefix(db, "sys_")
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// 创建 Casbin enforcer
	enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	return &CasbinAPIManager{
		db:       db,
		enforcer: enforcer,
	}, nil
}

// APIResource API资源定义（基于casbin_rule）
type APIResource struct {
	Name        string `json:"name"`        // 存储在v3字段
	Path        string `json:"path"`        // v1字段
	Method      string `json:"method"`      // v2字段
	Description string `json:"description"` // v4字段
	Module      string `json:"module"`      // v5字段
}

// PolicyEntry Casbin策略条目
type PolicyEntry struct {
	PType       string `json:"ptype"`        // p
	Subject     string `json:"subject"`      // v0: 角色
	Object      string `json:"object"`       // v1: API路径
	Action      string `json:"action"`       // v2: HTTP方法
	Name        string `json:"name"`        // v3: API名称
	Description string `json:"description"`  // v4: 描述
	Module      string `json:"module"`      // v5: 模块
}

// ======================================
// API资源发现和管理（基于casbin_rule）
// ======================================

// GetAllPolicies 获取所有策略
func (m *CasbinAPIManager) GetAllPolicies() ([]PolicyEntry, error) {
	policies := m.enforcer.GetPolicy()

	entries := make([]PolicyEntry, 0, len(policies))
	for _, policy := range policies {
		if len(policy) >= 3 && policy[0] == "p" {
			entry := PolicyEntry{
				PType:   policy[0],
				Subject: policy[1], // v0: 角色
				Object:  policy[2], // v1: API路径
				Action:  policy[3], // v2: HTTP方法
			}

			// 扩展字段
			if len(policy) > 4 {
				entry.Name = policy[4] // v3: API名称
			}
			if len(policy) > 5 {
				entry.Description = policy[5] // v4: 描述
			}
			if len(policy) > 6 {
				entry.Module = policy[6] // v5: 模块
			}

			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// GetUniqueAPIResources 获取唯一的API资源列表（去重）
func (m *CasbinAPIManager) GetUniqueAPIResources() ([]APIResource, error) {
	policies := m.enforcer.GetPolicy()

	// 使用map去重：key = path:method
	uniqueAPIs := make(map[string]APIResource)

	for _, policy := range policies {
		if len(policy) >= 4 && policy[0] == "p" {
			path := policy[2] // v1
			method := policy[3] // v2
			key := fmt.Sprintf("%s:%s", path, method)

			// 如果已存在，跳过
			if _, exists := uniqueAPIs[key]; exists {
				continue
			}

			api := APIResource{
				Path:   path,
				Method: method,
			}

			// 扩展字段
			if len(policy) > 4 {
				api.Name = policy[4]
			}
			if len(policy) > 5 {
				api.Description = policy[5]
			}
			if len(policy) > 6 {
				api.Module = policy[6]
			}

			uniqueAPIs[key] = api
		}
	}

	// 转换为切片
	result := make([]APIResource, 0, len(uniqueAPIs))
	for _, api := range uniqueAPIs {
		result = append(result, api)
	}

	return result, nil
}

// AddAPIWithMetadata 添加API资源（包含元数据）
func (m *CasbinAPIManager) AddAPIWithMetadata(roleCode, path, method, name, description, module string) error {
	// 策略格式：p, role, path, method, name, description, module
	policy := []string{roleCode, path, method, name, description, module}

	added, err := m.enforcer.AddPolicy("p", policy)
	if err != nil {
		return fmt.Errorf("添加策略失败: %w", err)
	}

	if !added {
		return fmt.Errorf("策略已存在")
	}

	// 保存策略
	return m.enforcer.SavePolicy()
}

// ======================================
// 角色权限管理
// ======================================

// GetRolePermissions 获取角色的所有权限
func (m *CasbinAPIManager) GetRolePermissions(roleCode string) ([]APIResource, error) {
	policies := m.enforcer.GetPermissionsForUser(roleCode)

	resources := make([]APIResource, 0, len(policies))
	for _, policy := range policies {
		if len(policy) >= 3 {
			api := APIResource{
				Path:  policy[1], // v1: 路径
				Method: policy[2], // v2: 方法
			}

			// 扩展字段
			if len(policy) > 3 {
				api.Name = policy[3]
			}
			if len(policy) > 4 {
				api.Description = policy[4]
			}
			if len(policy) > 5 {
				api.Module = policy[5]
			}

			resources = append(resources, api)
		}
	}

	return resources, nil
}

// AssignPermissionToRole 为角色分配权限
func (m *CasbinAPIManager) AssignPermissionToRole(roleCode, path, method, name, description, module string) error {
	// 添加策略：p, role, path, method, name, description, module
	_, err := m.enforcer.AddPolicy("p", roleCode, path, method, name, description, module)
	if err != nil {
		return fmt.Errorf("添加策略失败: %w", err)
	}

	return m.enforcer.SavePolicy()
}

// RevokePermissionFromRole 撤销角色权限
func (m *CasbinAPIManager) RevokePermissionFromRole(roleCode, path, method string) error {
	// 删除策略
	removed, err := m.enforcer.RemovePolicy("p", roleCode, path, method)
	if err != nil {
		return fmt.Errorf("删除策略失败: %w", err)
	}

	if !removed {
		return fmt.Errorf("策略不存在")
	}

	return m.enforcer.SavePolicy()
}

// BatchAssignPermissions 批量分配权限
func (m *CasbinAPIManager) BatchAssignPermissions(roleCode string, apis []APIResource) error {
	// 先删除现有策略
	existingPolicies := m.enforcer.GetPermissionsForUser(roleCode)
	for _, policy := range existingPolicies {
		if len(policy) >= 3 {
			m.enforcer.RemovePolicy("p", roleCode, policy[1], policy[2])
		}
	}

	// 添加新策略
	for _, api := range apis {
		_, err := m.enforcer.AddPolicy("p", roleCode, api.Path, api.Method, api.Name, api.Description, api.Module)
		if err != nil {
			return fmt.Errorf("添加策略失败: %w", err)
		}
	}

	return m.enforcer.SavePolicy()
}

// ======================================
// API发现和同步
// ======================================

// SyncCommonAPIs 同步常用API到casbin_rule
func (m *CasbinAPIManager) SyncCommonAPIs() error {
	// 定义常用API列表
	commonAPIs := []struct {
		Path        string
		Method      string
		Name        string
		Description string
		Module      string
	}{
		// 用户管理
		{"/api/system/users", "GET", "用户列表", "获取用户列表", "system"},
		{"/api/system/users", "POST", "创建用户", "创建新用户", "system"},
		{"/api/system/users/:id", "PUT", "更新用户", "更新用户信息", "system"},
		{"/api/system/users/:id", "DELETE", "删除用户", "删除用户", "system"},

		// 角色管理
		{"/api/system/roles", "GET", "角色列表", "获取角色列表", "system"},
		{"/api/system/roles", "POST", "创建角色", "创建新角色", "system"},

		// 菜单管理
		{"/api/system/menus", "GET", "菜单列表", "获取菜单列表", "system"},
		{"/api/system/menus", "POST", "创建菜单", "创建新菜单", "system"},
	}

	// 为admin角色添加这些API（默认权限）
	for _, api := range commonAPIs {
		// 检查策略是否已存在
		exists := m.enforcer.HasPolicy("p", "admin", api.Path, api.Method)

		if !exists {
			// 添加策略
			_, err := m.enforcer.AddPolicy("p", "admin", api.Path, api.Method, api.Name, api.Description, api.Module)
			if err != nil {
				return fmt.Errorf("添加策略失败: %w", err)
			}
		}
	}

	return m.enforcer.SavePolicy()
}

// GetAPIByPathAndMethod 根据路径和方法获取API信息
func (m *CasbinAPIManager) GetAPIByPathAndMethod(path, method string) (*APIResource, error) {
	// 从所有策略中查找匹配的API
	policies := m.enforcer.GetPolicy()

	for _, policy := range policies {
		if len(policy) >= 4 && policy[0] == "p" {
			policyPath := policy[2]
			policyMethod := policy[3]

			// 匹配路径（支持通配符）
			if matchPath(policyPath, path) && policyMethod == method {
				api := &APIResource{
					Path:   policyPath,
					Method: policyMethod,
				}

				if len(policy) > 4 {
					api.Name = policy[4]
				}
				if len(policy) > 5 {
					api.Description = policy[5]
				}
				if len(policy) > 6 {
					api.Module = policy[6]
				}

				return api, nil
			}
		}
	}

	return nil, fmt.Errorf("API not found")
}

// matchPath 路径匹配（支持通配符）
func matchPath(policyPath, requestPath string) bool {
	// 精确匹配
	if policyPath == requestPath {
		return true
	}

	// 通配符匹配：/api/users/* 匹配 /api/users/123
	if strings.Contains(policyPath, "*") {
		return strings.HasPrefix(requestPath, strings.TrimSuffix(policyPath, "*"))
	}

	return false
}

// ======================================
// 权限检查
// ======================================

// CheckPermission 检查权限
func (m *CasbinAPIManager) CheckPermission(roleCode, path, method string) (bool, error) {
	return m.enforcer.Enforce(roleCode, path, method)
}

// GetUserRoles 获取用户角色
func (m *CasbinAPIManager) GetUserRoles(userID uint) ([]models.Role, error) {
	var user models.User
	if err := m.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// 解析角色ID
	var roleIDs []uint
	if user.RoleIDs != "" {
		// 假设 RoleIDs 是JSON数组字符串
		// 这里需要根据实际实现调整
	}

	var roles []models.Role
	if len(roleIDs) > 0 {
		if err := m.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return nil, err
		}
	}

	return roles, nil
}