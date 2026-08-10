package system

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// RBACService 权限服务
type RBACService struct{}

// NewRBACService 创建权限服务
func NewRBACService() *RBACService {
	return &RBACService{}
}

// GetUserRoles 获取用户的角色列表
func (s *RBACService) GetUserRoles(userID uint) ([]*modelsystem.Role, error) {
	var user modelsystem.User
	err := database.GetDB().First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	// 解析 roleIds JSON 数组
	var roleIDs []uint
	err = json.Unmarshal([]byte(user.RoleIDs), &roleIDs)
	if err != nil {
		// 如果解析失败，返回空列表
		return []*modelsystem.Role{}, nil
	}

	// 查询角色
	var roles []*modelsystem.Role
	if len(roleIDs) > 0 {
		err = database.GetDB().Where("id IN ? AND status = 1", roleIDs).Find(&roles).Error
	} else {
		roles = []*modelsystem.Role{}
	}
	return roles, err
}

// BuildMenuTreeAndPermissions 构建菜单树和权限列表，同时返回角色（避免调用方重复查询）
func (s *RBACService) BuildMenuTreeAndPermissions(userID uint) ([]*modelsystem.Menu, []string, []*modelsystem.Role, error) {
	startTime := time.Now()
	logger.Debug("[登录调试-RBAC] BuildMenuTreeAndPermissions开始", zap.Uint("userID", userID))

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
	for _, role := range roles {
		if role.Code == "admin" {
			isAdmin = true
			break
		}
	}

	logger.Debug("[登录调试-RBAC] 开始查询所有菜单")
	menuStart := time.Now()
	// 获取所有菜单
	var allMenus []*modelsystem.Menu
	err = database.GetDB().Where("status = 1").Order("sort ASC").Find(&allMenus).Error
	logger.Debug("[登录调试-RBAC] 所有菜单查询完成",
		zap.Duration("耗时", time.Since(menuStart)),
		zap.Error(err),
		zap.Int("菜单总数", len(allMenus)))

	if err != nil {
		return nil, nil, nil, err
	}

	// 权限推导策略：统一从权限码推导菜单权限
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
		logger.Debug("[登录调试-RBAC] 从权限码推导菜单")

		// 1. 获取用户的所有权限码（从 role_permissions 表）
		for _, role := range roles {
			var rolePerms []modelsystem.RolePermission
			database.GetDB().Where("role_id = ?", role.ID).Preload("Permission").Find(&rolePerms)
			for _, rp := range rolePerms {
				if rp.Permission.Code != "" {
					permissions = append(permissions, rp.Permission.Code)
				}
			}
		}

		logger.Debug("[登录调试-RBAC] 获取到的权限码",
			zap.Int("权限码数量", len(permissions)),
			zap.Any("权限码列表", permissions))

		// 2. 从权限码提取 resource 列表
		allowedResources := make(map[string]bool)
		for _, permCode := range permissions {
			// 权限码格式：module.resource.action (使用点号分隔)
			parts := strings.Split(permCode, ".")
			logger.Debug("[登录调试-RBAC] 解析权限码",
				zap.String("权限码", permCode),
				zap.Int("段数", len(parts)),
				zap.Any("分段", parts))

			if len(parts) >= 2 {
				resource := parts[1]
				allowedResources[resource] = true
				logger.Debug("[登录调试-RBAC] 提取资源",
					zap.String("权限码", permCode),
					zap.String("资源", resource))
			}
		}

		// 转换为切片以便日志显示
		resourceList := make([]string, 0, len(allowedResources))
		for resource := range allowedResources {
			resourceList = append(resourceList, resource)
		}

		logger.Debug("[登录调试-RBAC] 权限码推导",
			zap.Int("权限码数量", len(permissions)),
			zap.Int("资源数量", len(allowedResources)),
			zap.Any("资源列表", resourceList))

		// 3. 根据 resource 匹配菜单
		for _, menu := range allMenus {
			if menu.Resource != "" && allowedResources[menu.Resource] {
				menuIDs[menu.ID] = true
				logger.Debug("[登录调试-RBAC] 菜单匹配",
					zap.String("菜单", menu.Name),
					zap.String("资源", menu.Resource))
				// 标记父菜单
				for _, m := range allMenus {
					if m.ID == menu.ParentID {
						menuIDs[m.ID] = true
					}
				}
			}
		}

		logger.Debug("[登录调试-RBAC] 推导出的菜单数量", zap.Int("数量", len(menuIDs)))
	}

	logger.Debug("[登录调试-RBAC] 开始构建菜单树")
	treeStart := time.Now()
	// 构建菜单树
	menuTree := s.buildMenuTree(allMenus, menuIDs, 0)
	logger.Debug("[登录调试-RBAC] 菜单树构建完成",
		zap.Duration("耗时", time.Since(treeStart)),
		zap.Int("菜单树节点数", len(menuTree)))

	logger.Debug("[登录调试-RBAC] BuildMenuTreeAndPermissions完成",
		zap.Duration("总耗时", time.Since(startTime)),
		zap.Bool("是管理员", isAdmin),
		zap.Int("权限数量", len(permissions)))

	return menuTree, permissions, roles, nil
}

// buildMenuTree 递归构建菜单树
func (s *RBACService) buildMenuTree(allMenus []*modelsystem.Menu, menuIDs map[uint]bool, parentID uint) []*modelsystem.Menu {
	var result []*modelsystem.Menu

	for _, menu := range allMenus {
		// 如果不是管理员且没有权限，跳过
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
func (s *RBACService) extractPermissions(menuTree []*modelsystem.Menu) []string {
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
