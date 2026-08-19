package system

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	permissionService     *PermissionService
	permissionServiceOnce sync.Once
	permissionServiceMu   sync.RWMutex
)

// PermissionService 权限服务
type PermissionService struct {
	db       *gorm.DB
	enforcer *casbin.Enforcer
}

// GetPermissionService 获取权限服务单例
func GetPermissionService() (*PermissionService, error) {
	permissionServiceMu.RLock()
	if permissionService != nil {
		permissionServiceMu.RUnlock()
		return permissionService, nil
	}
	permissionServiceMu.RUnlock()

	var initErr error
	permissionServiceOnce.Do(func() {
		service, err := NewPermissionService()
		if err != nil {
			initErr = err
			return
		}
		permissionServiceMu.Lock()
		permissionService = service
		permissionServiceMu.Unlock()
	})

	if initErr != nil {
		return nil, initErr
	}

	return permissionService, nil
}

// NewPermissionService 创建权限服务实例
func NewPermissionService() (*PermissionService, error) {
	db := database.GetDB()

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

	return &PermissionService{
		db:       db,
		enforcer: enforcer,
	}, nil
}

// ======================================
// 权限检查相关方法
// ======================================

// HasPermission 检查用户是否拥有指定权限（统一的权限检查方法）
// 权限代码格式：模块.资源.操作，例如 system.user.list
func (s *PermissionService) HasPermission(userID uint, permissionCode string) (bool, error) {
	// 获取用户角色
	roles, err := s.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	// 超级管理员角色检查（统一通过角色码判断）
	if s.IsAdmin(roles) {
		return true, nil
	}

	// 检查任意一个角色是否有权限
	// Casbin 策略格式：p, role_code, system.user.list, *
	for _, role := range roles {
		allowed, err := s.enforcer.Enforce(role.Code, permissionCode, "*")
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission 检查用户是否拥有任意一个指定权限
func (s *PermissionService) HasAnyPermission(userID uint, permissionCodes []string) (bool, error) {
	for _, code := range permissionCodes {
		allowed, err := s.HasPermission(userID, code)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions 检查用户是否拥有所有指定权限
func (s *PermissionService) HasAllPermissions(userID uint, permissionCodes []string) (bool, error) {
	for _, code := range permissionCodes {
		allowed, err := s.HasPermission(userID, code)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// ======================================
// 用户权限获取相关方法
// ======================================

// GetUserPermissions 获取用户的所有权限编码列表
func (s *PermissionService) GetUserPermissions(userID uint) ([]string, error) {
	// 获取用户角色
	roles, err := s.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	// 超级管理员返回所有权限
	if s.IsAdmin(roles) {
		return s.getAllPermissionCodes(), nil
	}

	// 获取角色权限
	permissionSet := make(map[string]bool)
	for _, role := range roles {
		permissions, err := s.GetRolePermissions(role.ID)
		if err != nil {
			continue
		}
		for _, perm := range permissions {
			permissionSet[perm.Code] = true
		}
	}

	// 转换为切片
	permissions := make([]string, 0, len(permissionSet))
	for code := range permissionSet {
		permissions = append(permissions, code)
	}

	return permissions, nil
}

// GetUserRoles 获取用户角色列表（统一方法）
func (s *PermissionService) GetUserRoles(userID uint) ([]*modelsystem.Role, error) {
	var user modelsystem.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var roleIDs []uint
	if err := json.Unmarshal([]byte(user.RoleIDs), &roleIDs); err != nil {
		return []*modelsystem.Role{}, nil
	}

	var roles []*modelsystem.Role
	if len(roleIDs) > 0 {
		if err := s.db.Where("id IN ? AND status = 1", roleIDs).Find(&roles).Error; err != nil {
			return nil, err
		}
	}
	return roles, nil
}

// IsAdmin 检查是否为超级管理员（统一通过角色码 "admin" 判断）
func (s *PermissionService) IsAdmin(roles []*modelsystem.Role) bool {
	for _, role := range roles {
		if role.Code == "admin" {
			return true
		}
	}
	return false
}

// getAllPermissionCodes 获取所有权限编码（用于超级管理员）
func (s *PermissionService) getAllPermissionCodes() []string {
	var permissions []modelsystem.Permission
	if err := s.db.Where("status = 1").Pluck("code", &permissions).Error; err != nil {
		return []string{"*.*.*"}
	}

	codes := make([]string, len(permissions))
	for i, perm := range permissions {
		codes[i] = perm.Code
	}
	return codes
}

// ======================================
// 角色权限管理相关方法
// ======================================

// GetRolePermissions 获取角色的所有权限
func (s *PermissionService) GetRolePermissions(roleID uint) ([]modelsystem.Permission, error) {
	var permissions []modelsystem.Permission

	err := s.db.Table("sys_permissions").
		Joins("INNER JOIN sys_role_permissions ON sys_permissions.id = sys_role_permissions.permission_id").
		Where("sys_role_permissions.role_id = ? AND sys_permissions.status = 1", roleID).
		Order("sys_permissions.level ASC, sys_permissions.sort_order ASC").
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// AssignPermissionToRole 为角色分配权限
func (s *PermissionService) AssignPermissionToRole(roleID uint, permissionID uint) error {
	// 检查角色和权限是否存在
	var role modelsystem.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	var permission modelsystem.Permission
	if err := s.db.First(&permission, permissionID).Error; err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// 创建关联
	rolePerm := modelsystem.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	if err := s.db.FirstOrCreate(&rolePerm, "role_id = ? AND permission_id = ?", roleID, permissionID).Error; err != nil {
		return err
	}

	// 同步到 Casbin
	if err := s.syncRoleToCasbin(&role); err != nil {
		return err
	}

	return nil
}

// RevokePermissionFromRole 撤销角色权限
func (s *PermissionService) RevokePermissionFromRole(roleID uint, permissionID uint) error {
	if err := s.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).Delete(&modelsystem.RolePermission{}).Error; err != nil {
		return err
	}

	// 重新同步角色到 Casbin
	var role modelsystem.Role
	if err := s.db.First(&role, roleID).Error; err == nil {
		s.syncRoleToCasbin(&role)
	}

	return nil
}

// BatchAssignPermissions 批量分配权限
func (s *PermissionService) BatchAssignPermissions(roleID uint, permissionIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有权限
		if err := tx.Where("role_id = ?", roleID).Delete(&modelsystem.RolePermission{}).Error; err != nil {
			return err
		}

		// 添加新权限
		for _, permID := range permissionIDs {
			rolePerm := modelsystem.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			if err := tx.Create(&rolePerm).Error; err != nil {
				return err
			}
		}

		// 同步到 Casbin
		var role modelsystem.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return err
		}

		if err := s.syncRoleToCasbin(&role); err != nil {
			return err
		}

		return nil
	})
}

// ======================================
// 权限管理相关方法
// ======================================

// GetPermissionList 获取权限列表（分页）
func (s *PermissionService) GetPermissionList(page, size int, query map[string]interface{}) ([]modelsystem.Permission, int64, error) {
	var permissions []modelsystem.Permission
	var total int64

	// 构建查询
	db := s.db.Model(&modelsystem.Permission{})

	// 添加搜索条件
	if name, ok := query["name"].(string); ok && name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if code, ok := query["code"].(string); ok && code != "" {
		db = db.Where("code LIKE ?", "%"+code+"%")
	}
	if module, ok := query["module"].(string); ok && module != "" {
		db = db.Where("module = ?", module)
	}
	if status, ok := query["status"].(string); ok && status != "" {
		db = db.Where("status = ?", status)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	if err := db.Order("level ASC, sort_order ASC, id ASC").
		Offset(offset).Limit(size).
		Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// CreatePermission 创建权限
func (s *PermissionService) CreatePermission(permission *modelsystem.Permission) error {
	if err := s.db.Create(permission).Error; err != nil {
		return err
	}
	return nil
}

// UpdatePermission 更新权限
func (s *PermissionService) UpdatePermission(permission *modelsystem.Permission) error {
	if err := s.db.Save(permission).Error; err != nil {
		return err
	}
	return nil
}

// DeletePermission 删除权限
func (s *PermissionService) DeletePermission(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除角色权限关联
		if err := tx.Where("permission_id = ?", id).Delete(&modelsystem.RolePermission{}).Error; err != nil {
			return err
		}

		// 删除用户权限关联
		if err := tx.Where("permission_id = ?", id).Delete(&modelsystem.UserPermission{}).Error; err != nil {
			return err
		}

		// 删除权限
		return tx.Delete(&modelsystem.Permission{}, id).Error
	})
}

// GetAllPermissionOptions 获取所有权限选项（不分页，用于权限树/选择器）
func (s *PermissionService) GetAllPermissionOptions() ([]modelsystem.Permission, error) {
	var permissions []modelsystem.Permission
	err := s.db.Where("status = ?", 1).Select("id, name, code, module, resource, action, parent_id, level, sort_order").Order("sort_order ASC").Find(&permissions).Error
	return permissions, err
}

// GetPermissionTree 获取权限树
func (s *PermissionService) GetPermissionTree() ([]modelsystem.Permission, error) {
	var permissions []modelsystem.Permission

	err := s.db.Where("status = 1").
		Order("level ASC, sort_order ASC, id ASC").
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return s.buildPermissionTree(permissions, nil), nil
}

// buildPermissionTree 构建权限树
func (s *PermissionService) buildPermissionTree(permissions []modelsystem.Permission, parentID *uint) []modelsystem.Permission {
	var tree []modelsystem.Permission

	for _, perm := range permissions {
		if (parentID == nil && perm.ParentID == nil) ||
			(parentID != nil && perm.ParentID != nil && *perm.ParentID == *parentID) {
			perm.Children = s.buildPermissionTree(permissions, &perm.ID)
			tree = append(tree, perm)
		}
	}

	return tree
}

// ======================================
// API权限映射表（操作级权限 -> API级权限）
// ======================================

// SystemAPIEndpoint 系统API端点定义
type SystemAPIEndpoint struct {
	ID          string   // API唯一标识
	Path        string   // API路径
	Methods     []string // 支持的HTTP方法
	Name        string   // API名称（中文）
	Description string   // API描述
	Category    string   // API分类（用户管理、角色管理等）
}

// systemAPIEndpoints 系统所有API端点定义
var systemAPIEndpoints = []SystemAPIEndpoint{
	// 用户管理API
	{ID: "user.list", Path: "/api/system/users", Methods: []string{"GET"}, Name: "查看用户列表", Description: "获取系统用户列表（分页）", Category: "用户管理"},
	{ID: "user.view", Path: "/api/system/users/:id", Methods: []string{"GET"}, Name: "查看用户详情", Description: "获取单个用户的详细信息", Category: "用户管理"},
	{ID: "user.create", Path: "/api/system/users", Methods: []string{"POST"}, Name: "创建用户", Description: "创建新用户", Category: "用户管理"},
	{ID: "user.update", Path: "/api/system/users/:id", Methods: []string{"PUT"}, Name: "更新用户", Description: "更新用户信息", Category: "用户管理"},
	{ID: "user.delete", Path: "/api/system/users/:id", Methods: []string{"DELETE"}, Name: "删除用户", Description: "删除用户", Category: "用户管理"},
	{ID: "user.reset_password", Path: "/api/system/users/:id/password", Methods: []string{"PUT"}, Name: "重置密码", Description: "重置用户密码", Category: "用户管理"},
	{ID: "user.view_roles", Path: "/api/system/users/:id/roles", Methods: []string{"GET"}, Name: "查看用户角色", Description: "查看用户的角色分配", Category: "用户管理"},

	// 角色管理API
	{ID: "role.list", Path: "/api/system/roles", Methods: []string{"GET"}, Name: "查看角色列表", Description: "获取系统角色列表", Category: "角色管理"},
	{ID: "role.view", Path: "/api/system/roles/:id", Methods: []string{"GET"}, Name: "查看角色详情", Description: "获取单个角色的详细信息", Category: "角色管理"},
	{ID: "role.create", Path: "/api/system/roles", Methods: []string{"POST"}, Name: "创建角色", Description: "创建新角色", Category: "角色管理"},
	{ID: "role.update", Path: "/api/system/roles/:id", Methods: []string{"PUT"}, Name: "更新角色", Description: "更新角色信息", Category: "角色管理"},
	{ID: "role.delete", Path: "/api/system/roles/:id", Methods: []string{"DELETE"}, Name: "删除角色", Description: "删除角色", Category: "角色管理"},
	{ID: "role.view_permissions", Path: "/api/system/roles/:id/permissions", Methods: []string{"GET"}, Name: "查看角色权限", Description: "查看角色的API权限分配", Category: "角色管理"},
	{ID: "role.assign_permissions", Path: "/api/system/roles/:id/permissions", Methods: []string{"POST"}, Name: "分配权限", Description: "为角色分配API权限", Category: "角色管理"},

	// 菜单管理API
	{ID: "menu.list", Path: "/api/system/menus", Methods: []string{"GET"}, Name: "查看菜单列表", Description: "获取系统菜单列表", Category: "菜单管理"},
	{ID: "menu.tree", Path: "/api/system/menus/tree", Methods: []string{"GET"}, Name: "查看菜单树", Description: "获取菜单树形结构", Category: "菜单管理"},
	{ID: "menu.view", Path: "/api/system/menus/:id", Methods: []string{"GET"}, Name: "查看菜单详情", Description: "获取单个菜单的详细信息", Category: "菜单管理"},
	{ID: "menu.create", Path: "/api/system/menus", Methods: []string{"POST"}, Name: "创建菜单", Description: "创建新菜单", Category: "菜单管理"},
	{ID: "menu.update", Path: "/api/system/menus/:id", Methods: []string{"PUT"}, Name: "更新菜单", Description: "更新菜单信息", Category: "菜单管理"},
	{ID: "menu.delete", Path: "/api/system/menus/:id", Methods: []string{"DELETE"}, Name: "删除菜单", Description: "删除菜单", Category: "菜单管理"},

	// 权限管理API
	{ID: "permission.list", Path: "/api/system/permissions", Methods: []string{"GET"}, Name: "查看权限列表", Description: "获取系统权限列表（分页）", Category: "权限管理"},
	{ID: "permission.tree", Path: "/api/system/permissions/tree", Methods: []string{"GET"}, Name: "查看权限树", Description: "获取权限树形结构", Category: "权限管理"},
	{ID: "permission.view", Path: "/api/system/permissions/:id", Methods: []string{"GET"}, Name: "查看权限详情", Description: "获取单个权限的详细信息", Category: "权限管理"},
	{ID: "permission.create", Path: "/api/system/permissions", Methods: []string{"POST"}, Name: "创建权限", Description: "创建新权限", Category: "权限管理"},
	{ID: "permission.update", Path: "/api/system/permissions/:id", Methods: []string{"PUT"}, Name: "更新权限", Description: "更新权限信息", Category: "权限管理"},
	{ID: "permission.delete", Path: "/api/system/permissions/:id", Methods: []string{"DELETE"}, Name: "删除权限", Description: "删除权限", Category: "权限管理"},
}

// GetSystemAPIEndpoints 获取系统所有API端点定义
func GetSystemAPIEndpoints() []SystemAPIEndpoint {
	return systemAPIEndpoints
}

// GetAPIEndpointsByCategory 按分类获取API端点
func GetAPIEndpointsByCategory(category string) []SystemAPIEndpoint {
	result := []SystemAPIEndpoint{}
	for _, endpoint := range systemAPIEndpoints {
		if endpoint.Category == category {
			result = append(result, endpoint)
		}
	}
	return result
}

// GetAllCategories 获取所有API分类
func GetAllCategories() []string {
	categoryMap := make(map[string]bool)
	for _, endpoint := range systemAPIEndpoints {
		categoryMap[endpoint.Category] = true
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}
	return categories
}

// ======================================
// 真正的 Level 4 API级权限管理方法
// ======================================

// AssignAPIPermission 为角色分配API权限（直接操作API端点）
func (s *PermissionService) AssignAPIPermission(roleCode string, apiPath string, httpMethod string) error {
	_, err := s.enforcer.AddPolicy(roleCode, apiPath, httpMethod)
	if err != nil {
		return err
	}
	return s.enforcer.SavePolicy()
}

// RevokeAPIPermission 撤销角色的API权限
func (s *PermissionService) RevokeAPIPermission(roleCode string, apiPath string, httpMethod string) error {
	_, err := s.enforcer.RemovePolicy(roleCode, apiPath, httpMethod)
	if err != nil {
		return err
	}
	return s.enforcer.SavePolicy()
}

// BatchAssignAPIPermissions 批量为角色分配API权限
func (s *PermissionService) BatchAssignAPIPermissions(roleCode string, permissions []struct {
	Path   string
	Method string
}) error {
	for _, perm := range permissions {
		_, _ = s.enforcer.AddPolicy(roleCode, perm.Path, perm.Method)
	}
	return s.enforcer.SavePolicy()
}

// GetRoleAPIPermissions 获取角色的所有API权限
func (s *PermissionService) GetRoleAPIPermissions(roleCode string) ([]struct {
	Path   string
	Method string
}, error) {
	policies := s.enforcer.GetFilteredPolicy(0, roleCode)

	result := make([]struct {
		Path   string
		Method string
	}, 0, len(policies))

	for _, policy := range policies {
		if len(policy) >= 3 {
			result = append(result, struct {
				Path   string
				Method string
			}{
				Path:   policy[1],
				Method: policy[2],
			})
		}
	}

	return result, nil
}

// ClearAllPolicies 清除所有Casbin策略
func (s *PermissionService) ClearAllPolicies() error {
	s.enforcer.ClearPolicy()
	return s.enforcer.SavePolicy()
}

// GetAllPolicies 获取所有Casbin策略
func (s *PermissionService) GetAllPolicies() [][]string {
	return s.enforcer.GetPolicy()
}

// ======================================
// Casbin 同步相关方法
// ======================================

// syncRoleToCasbin 同步角色权限到 Casbin
func (s *PermissionService) syncRoleToCasbin(role *modelsystem.Role) error {
	// 删除角色的所有旧策略
	s.enforcer.RemoveFilteredPolicy(0, role.Code)

	// 获取角色的权限
	permissions, err := s.GetRolePermissions(role.ID)
	if err != nil {
		return err
	}

	// 添加新策略
	for _, perm := range permissions {
		if _, err := s.enforcer.AddPolicy(role.Code, perm.Code, "*"); err != nil {
			return err
		}
	}

	// 保存策略
	return s.enforcer.SavePolicy()
}

// SyncAllRolesToCasbin 同步所有角色到 Casbin
func (s *PermissionService) SyncAllRolesToCasbin() error {
	var roles []modelsystem.Role
	if err := s.db.Where("status = 1").Find(&roles).Error; err != nil {
		return err
	}

	for _, role := range roles {
		if err := s.syncRoleToCasbin(&role); err != nil {
			return err
		}
	}

	return nil
}

// InitializeCasbinPolicies 初始化 Casbin 策略
// 清除所有旧的策略（包括API路径格式），并从权限表重新同步
func (s *PermissionService) InitializeCasbinPolicies() error {
	// 1. 清除所有现有策略
	if err := s.ClearAllPolicies(); err != nil {
		return fmt.Errorf("清除旧策略失败: %w", err)
	}

	// 2. 同步所有角色的权限到 Casbin
	if err := s.SyncAllRolesToCasbin(); err != nil {
		return fmt.Errorf("同步角色权限失败: %w", err)
	}

	return nil
}

// ======================================
// 菜单与路由构建（从 RBACService 合并）
// ======================================

// BuildMenuTreeAndPermissions 构建菜单树和权限列表，同时返回角色
func (s *PermissionService) BuildMenuTreeAndPermissions(userID uint) ([]*modelsystem.Menu, []string, []*modelsystem.Role, error) {
	roles, err := s.GetUserRoles(userID)
	if err != nil {
		return nil, nil, nil, err
	}

	isAdmin := s.IsAdmin(roles)

	// 获取所有菜单
	var allMenus []*modelsystem.Menu
	err = s.db.Where("status = 1").Order("sort ASC").Find(&allMenus).Error
	if err != nil {
		return nil, nil, nil, err
	}

	menuIDs := make(map[uint]bool)
	permissions := make([]string, 0)

	if isAdmin {
		// 管理员：拥有所有菜单和通配符权限
		for _, menu := range allMenus {
			menuIDs[menu.ID] = true
		}
		permissions = append(permissions, "*:*:*")
	} else {
		// 非管理员：从权限码推导菜单
		// 1. 获取用户的所有权限码（从 role_permissions 表）
		for _, role := range roles {
			var rolePerms []modelsystem.RolePermission
			s.db.Where("role_id = ?", role.ID).Preload("Permission").Find(&rolePerms)
			for _, rp := range rolePerms {
				if rp.Permission.Code != "" {
					permissions = append(permissions, rp.Permission.Code)
				}
			}
		}

		// 2. 从权限码提取 resource 列表
		allowedResources := make(map[string]bool)
		for _, permCode := range permissions {
			parts := strings.Split(permCode, ".")
			if len(parts) >= 2 {
				resource := parts[1]
				allowedResources[resource] = true
			}
		}

		// 3. 根据 resource 匹配菜单
		for _, menu := range allMenus {
			if menu.Resource != "" && allowedResources[menu.Resource] {
				menuIDs[menu.ID] = true
				// 标记父菜单
				for _, m := range allMenus {
					if m.ID == menu.ParentID {
						menuIDs[m.ID] = true
					}
				}
			}
		}
	}

	// 构建菜单树
	menuTree := s.buildMenuTree(allMenus, menuIDs, 0)

	logger.Debug("[BuildMenuTreeAndPermissions]",
		zap.Uint("user_id", userID),
		zap.Bool("isAdmin", isAdmin),
		zap.Int("menu_count", len(menuTree)),
		zap.Int("permission_count", len(permissions)))

	return menuTree, permissions, roles, nil
}

// GetMenuPathsByRoleIDs 按角色集合推导可见一级菜单（编辑用户时家目录候选的权威来源）
// 推导链路与 BuildMenuTreeAndPermissions 一致：角色 → 权限码 → resource → 菜单.Resource 匹配，
// admin 角色返回全部启用菜单。仅返回一级菜单（parentId = 0）且配置了 path 的项。
func (s *PermissionService) GetMenuPathsByRoleIDs(roleIDs []uint) ([]*modelsystem.Menu, error) {
	result := make([]*modelsystem.Menu, 0)
	if len(roleIDs) == 0 {
		return result, nil
	}

	var roles []*modelsystem.Role
	if err := s.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}

	var allMenus []*modelsystem.Menu
	if err := s.db.Where("status = 1").Order("sort ASC").Find(&allMenus).Error; err != nil {
		return nil, err
	}

	menuByID := make(map[uint]*modelsystem.Menu, len(allMenus))
	for _, m := range allMenus {
		menuByID[m.ID] = m
	}

	visible := make(map[uint]bool)
	if s.IsAdmin(roles) {
		// 管理员：拥有所有菜单
		for _, m := range allMenus {
			visible[m.ID] = true
		}
	} else {
		// 角色权限码 → resource
		var rolePerms []modelsystem.RolePermission
		if err := s.db.Where("role_id IN ?", roleIDs).Preload("Permission").Find(&rolePerms).Error; err != nil {
			return nil, err
		}
		allowedResources := make(map[string]bool)
		for _, rp := range rolePerms {
			parts := strings.Split(rp.Permission.Code, ".")
			if len(parts) >= 2 {
				allowedResources[parts[1]] = true
			}
		}

		// resource 匹配菜单，并标记父菜单
		for _, m := range allMenus {
			if m.Resource != "" && allowedResources[m.Resource] {
				visible[m.ID] = true
				if parent, ok := menuByID[m.ParentID]; ok {
					visible[parent.ID] = true
				}
			}
		}
	}

	for _, m := range allMenus {
		if m.ParentID == 0 && visible[m.ID] && m.Path != "" {
			result = append(result, m)
		}
	}
	return result, nil
}

// buildMenuTree 递归构建菜单树
func (s *PermissionService) buildMenuTree(allMenus []*modelsystem.Menu, menuIDs map[uint]bool, parentID uint) []*modelsystem.Menu {
	var result []*modelsystem.Menu

	for _, menu := range allMenus {
		if _, hasPermission := menuIDs[menu.ID]; !hasPermission {
			continue
		}

		if menu.ParentID == parentID {
			menuItem := &modelsystem.Menu{
				ID:         menu.ID,
				Name:       menu.Name,
				Icon:       menu.Icon,
				Path:       menu.Path,
				Permission: menu.Permission,
				MenuType:   menu.MenuType,
				ParentID:   menu.ParentID,
				Sort:       menu.Sort,
				Status:     menu.Status,
			}

			children := s.buildMenuTree(allMenus, menuIDs, menu.ID)
			if len(children) > 0 {
				menuItem.Children = children
			}

			result = append(result, menuItem)
		}
	}

	return result
}

// ======================================
// 辅助方法
// ======================================

// LogPermissionOperation 记录权限操作日志
func (s *PermissionService) LogPermissionOperation(userID uint, permissionCode string, action string, result bool, ipAddress string, userAgent string) error {
	log := modelsystem.PermissionLog{
		UserID:         userID,
		PermissionCode: permissionCode,
		Action:         action,
		Result:         "allowed",
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}

	if !result {
		log.Result = "denied"
	}

	return s.db.Create(&log).Error
}
