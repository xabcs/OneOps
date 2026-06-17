package services

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"oneops/backend/logger"
	"oneops/backend/models"

	"go.uber.org/zap"
)

// RBACService 权限服务
type RBACService struct{}

// NewRBACService 创建权限服务
func NewRBACService() *RBACService {
	return &RBACService{}
}

// GetUserRoles 获取用户的角色列表
func (s *RBACService) GetUserRoles(userID uint) ([]*models.Role, error) {
	var user models.User
	err := db.First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	// 解析 roleIds JSON 数组
	var roleIDs []uint
	err = json.Unmarshal([]byte(user.RoleIDs), &roleIDs)
	if err != nil {
		// 如果解析失败，返回空列表
		return []*models.Role{}, nil
	}

	// 查询角色
	var roles []*models.Role
	if len(roleIDs) > 0 {
		err = db.Where("id IN ? AND status = 1", roleIDs).Find(&roles).Error
	} else {
		roles = []*models.Role{}
	}
	return roles, err
}

// BuildMenuTreeAndPermissions 构建菜单树和权限列表，同时返回角色（避免调用方重复查询）
func (s *RBACService) BuildMenuTreeAndPermissions(userID uint) ([]*models.Menu, []string, []*models.Role, error) {
	startTime := time.Now()
	logger.Debug("[登录调试-RBAC] BuildMenuTreeAndPermissions开始", zap.Uint("userID", userID))

	// 先查缓存
	logger.Debug("[登录调试-RBAC] 检查缓存")
	rbacCache.mu.RLock()
	if entry, ok := rbacCache.entries[userID]; ok && time.Now().Before(entry.expireAt) {
		rbacCache.mu.RUnlock()
		logger.Debug("[登录调试-RBAC] 命中缓存，直接返回")
		return entry.menuTree, entry.permissions, entry.roles, nil
	}
	rbacCache.mu.RUnlock()
	logger.Debug("[登录调试-RBAC] 缓存未命中，开始查询")

	logger.Debug("[登录调试-RBAC] 开始获取用户角色")
	roleStart := time.Now()
	roles, err := s.GetUserRoles(userID)
	logger.Debug("[登录调试-RBAC] 用户角色获取完成",
		zap.Duration("耗时", time.Since(roleStart)),
		zap.Error(err),
		zap.Int("角色数量", len(roles)))

	if err != nil {
		return nil, nil, nil, err
	}

	// 检查是否是管理员
	isAdmin := false
	menuIDs := make(map[uint]bool)
	for _, role := range roles {
		if role.Code == "admin" {
			isAdmin = true
		}

		// 解析 menuIds JSON 数组
		var roleMenuIDs []uint
		if err := json.Unmarshal([]byte(role.MenuIDs), &roleMenuIDs); err == nil {
			for _, id := range roleMenuIDs {
				menuIDs[id] = true
			}
		}
	}

	logger.Debug("[登录调试-RBAC] 开始查询所有菜单")
	menuStart := time.Now()
	// 获取所有菜单
	var allMenus []*models.Menu
	err = db.Where("status = 1").Order("sort ASC").Find(&allMenus).Error
	logger.Debug("[登录调试-RBAC] 所有菜单查询完成",
		zap.Duration("耗时", time.Since(menuStart)),
		zap.Error(err),
		zap.Int("菜单总数", len(allMenus)))

	if err != nil {
		return nil, nil, nil, err
	}

	// 如果是管理员，拥有所有菜单
	if isAdmin {
		menuIDs = make(map[uint]bool)
		for _, menu := range allMenus {
			menuIDs[menu.ID] = true
		}
	}

	logger.Debug("[登录调试-RBAC] 开始构建菜单树")
	treeStart := time.Now()
	// 构建菜单树
	menuTree := s.buildMenuTree(allMenus, menuIDs, 0)
	logger.Debug("[登录调试-RBAC] 菜单树构建完成",
		zap.Duration("耗时", time.Since(treeStart)),
		zap.Int("菜单树节点数", len(menuTree)))

	// 提取权限列表
	permissions := s.extractPermissions(menuTree)
	if isAdmin {
		permissions = append(permissions, "*:*:*")
	}

	logger.Debug("[登录调试-RBAC] 开始写入缓存")
	// 写入缓存
	rbacCache.mu.Lock()
	rbacCache.entries[userID] = &rbacCacheEntry{
		menuTree:    menuTree,
		permissions: permissions,
		roles:       roles,
		expireAt:    time.Now().Add(rbacCacheTTL),
	}
	rbacCache.mu.Unlock()
	logger.Debug("[登录调试-RBAC] 缓存写入完成")

	logger.Debug("[登录调试-RBAC] BuildMenuTreeAndPermissions完成",
		zap.Duration("总耗时", time.Since(startTime)),
		zap.Bool("是管理员", isAdmin),
		zap.Int("权限数量", len(permissions)))

	return menuTree, permissions, roles, nil
}

// buildMenuTree 递归构建菜单树
func (s *RBACService) buildMenuTree(allMenus []*models.Menu, menuIDs map[uint]bool, parentID uint) []*models.Menu {
	var result []*models.Menu

	for _, menu := range allMenus {
		// 如果不是管理员且没有权限，跳过
		if _, hasPermission := menuIDs[menu.ID]; !hasPermission {
			continue
		}

		if menu.ParentID == parentID {
			menuItem := &models.Menu{
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

			// 递归获取子菜单
			children := s.buildMenuTree(allMenus, menuIDs, menu.ID)
			if len(children) > 0 {
				menuItem.Children = children
			}

			result = append(result, menuItem)
		}
	}

	return result
}

// extractPermissions 从菜单树中提取所有权限
func (s *RBACService) extractPermissions(menuTree []*models.Menu) []string {
	permissions := make([]string, 0)
	for _, menu := range menuTree {
		if menu.Permission != "" {
			permissions = append(permissions, menu.Permission)
		}
		if len(menu.Children) > 0 {
			childPermissions := s.extractPermissions(menu.Children)
			permissions = append(permissions, childPermissions...)
		}
	}
	return permissions
}

// JSONArray JSON 数组类型（用于 GORM）
type JSONArray []uint

// Scan 实现 sql.Scanner 接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = []uint{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口
func (j JSONArray) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return json.Marshal(j)
}

// StringArray 字符串数组类型（用于 GORM）
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return nil
	}

	// 去除方括号和空格
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "[")
	str = strings.TrimSuffix(str, "]")

	if str == "" {
		*s = []string{}
		return nil
	}

	// 分割字符串
	*s = strings.Split(str, ",")
	return nil
}

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}
