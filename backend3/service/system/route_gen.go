package system

import (
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/logger"
)

// RouteGenService 路由生成服务
type RouteGenService struct {
	db *gorm.DB
}

// NewRouteGenService 创建路由生成服务
func NewRouteGenService(db *gorm.DB) *RouteGenService {
	return &RouteGenService{db: db}
}

// GetConstantRoutes 获取常量路由（无需登录即可访问的路由）
func (s *RouteGenService) GetConstantRoutes() []map[string]interface{} {
	// 常量路由：403、404、500、登录页、iframe-page 等
	return []map[string]interface{}{
		{
			"id":        "403",
			"name":      "403",
			"path":      "/403",
			"component": "layout.blank$view.403",
			"meta": map[string]interface{}{
				"title":      "403",
				"i18nKey":    "route.403",
				"constant":   true,
				"hideInMenu": true,
			},
		},
		{
			"id":        "404",
			"name":      "404",
			"path":      "/404",
			"component": "layout.blank$view.404",
			"meta": map[string]interface{}{
				"title":      "404",
				"i18nKey":    "route.404",
				"constant":   true,
				"hideInMenu": true,
			},
		},
		{
			"id":        "500",
			"name":      "500",
			"path":      "/500",
			"component": "layout.blank$view.500",
			"meta": map[string]interface{}{
				"title":      "500",
				"i18nKey":    "route.500",
				"constant":   true,
				"hideInMenu": true,
			},
		},
		{
			"id":        "login",
			"name":      "login",
			"path":      "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
			"component": "layout.blank$view.login",
			"props":     true,
			"meta": map[string]interface{}{
				"title":      "login",
				"i18nKey":    "route.login",
				"constant":   true,
				"hideInMenu": true,
			},
		},
		{
			"id":        "iframe-page",
			"name":      "iframe-page",
			"path":      "/iframe-page/:url",
			"component": "layout.base$view.iframe-page",
			"props":     true,
			"meta": map[string]interface{}{
				"title":      "iframe-page",
				"i18nKey":    "route.iframe-page",
				"constant":   true,
				"hideInMenu": true,
				"keepAlive":  true,
			},
		},
	}
}

// GetUserRoutes 获取用户路由（需要登录，根据用户角色返回）
// 返回生成的路由列表
func (s *RouteGenService) GetUserRoutes(userID uint) ([]map[string]interface{}, error) {
	// 清除该用户的缓存，确保路由配置每次都是最新的
	// 因为路由配置（component字段）需要动态生成，缓存会导致代码修改不生效

	logger.Info("[GetUserRoutes] 开始生成路由", zap.Uint("userID", userID))

	rbacService := NewRBACService()
	menuTree, _, roles, err := rbacService.BuildMenuTreeAndPermissions(userID)
	if err != nil {
		return nil, err
	}

	// 检查是否是超级管理员
	isSuper := false
	for _, role := range roles {
		if role.Code == "R_SUPER" || role.Code == "admin" {
			isSuper = true
			break
		}
	}

	// 将菜单转换为前端路由格式
	routes := s.convertMenusToRoutes(menuTree, isSuper)

	// 添加隐藏路由到对应的父路由下
	routes = s.appendHiddenRoutesToMenuTree(routes)

	return routes, nil
}

// IsRouteExist 检查路由是否存在于用户权限中
func (s *RouteGenService) IsRouteExist(userID uint, routeName string) (bool, error) {
	rbacService := NewRBACService()

	// 检查是否为管理员
	roles, err := rbacService.GetUserRoles(userID)
	if err == nil {
		for _, role := range roles {
			if role.Code == "admin" {
				// 管理员可以访问所有路由
				return true, nil
			}
		}
	}

	// 获取用户权限列表
	_, permissions, _, err := rbacService.BuildMenuTreeAndPermissions(userID)
	if err != nil {
		return false, err
	}

	// 检查路由是否在用户权限中
	// 对于详情页，检查其父路由权限
	parentRoute := ""
	switch routeName {
	case "k8s_deployment_detail", "k8s_deployment_detail_view":
		parentRoute = "k8s_workload_query"
	case "k8s_statefulset_detail", "k8s_statefulset_detail_view":
		parentRoute = "k8s_workload_query"
	case "k8s_daemonset_detail", "k8s_daemonset_detail_view":
		parentRoute = "k8s_workload_query"
	case "k8s_pod_detail", "k8s_pod_detail_view":
		parentRoute = "k8s_workload_query"
	case "cmdb_server_detail", "cmdb_server_detail_view":
		parentRoute = "cmdb:server:query"
	case "monitoring_servers_detail", "monitoring_servers_detail_view":
		parentRoute = "monitoring:server:query"
	default:
		// 对于其他路由，检查是否直接拥有权限
		for _, perm := range permissions {
			if perm == routeName {
				return true, nil
			}
		}
		return false, nil
	}

	// 检查父路由权限
	if parentRoute != "" {
		for _, perm := range permissions {
			if perm == parentRoute {
				return true, nil
			}
		}
	}

	return false, nil
}

// InvalidateCache 清除RBAC缓存并重新同步菜单
func (s *RouteGenService) InvalidateCache() error {
	// 重新同步菜单，确保数据最新
	initService := NewInitService()
	return initService.SyncMenus()
}

// DebugCache 调试当前缓存内容
func (s *RouteGenService) DebugCache(userID uint) (map[string]interface{}, error) {
	rbacService := NewRBACService()
	menuTree, permissions, roles, err := rbacService.BuildMenuTreeAndPermissions(userID)
	if err != nil {
		return nil, err
	}

	// 检查是否包含 webterminal
	hasWebTerminal := false
	for _, menu := range menuTree {
		if menu.Path == "/webterminal" {
			hasWebTerminal = true
			break
		}
	}

	return map[string]interface{}{
		"has_webterminal":  hasWebTerminal,
		"menu_count":       len(menuTree),
		"permission_count": len(permissions),
		"role_count":       len(roles),
		"menus":            menuTree,
	}, nil
}

// convertMenusToRoutes 将数据库菜单转换为前端路由格式
func (s *RouteGenService) convertMenusToRoutes(menus []*modelsystem.Menu, isSuper bool) []map[string]interface{} {
	routes := make([]map[string]interface{}, 0)

	for _, menu := range menus {
		// 检查是否有子菜单
		hasChildren := len(menu.Children) > 0

		route := s.buildRouteFromMenu(menu, hasChildren, isSuper)

		// 如果有子菜单，递归处理
		if hasChildren {
			childrenRoutes := s.convertMenusToRoutes(menu.Children, isSuper)
			if len(childrenRoutes) > 0 {
				route["children"] = childrenRoutes
			}
		}

		routes = append(routes, route)
	}

	return routes
}

// appendHiddenRoutesToMenuTree 将隐藏路由添加到菜单树中
// 注意：详情页路由由前端 Elegant Router 自动生成，不需要后端硬编码
// 前端文件 src/views/cmdb/server/detail.vue 会自动生成 cmdb_server-detail 路由
func (s *RouteGenService) appendHiddenRoutesToMenuTree(routes []map[string]interface{}) []map[string]interface{} {
	// 不需要添加任何硬编码路由，前端 Elegant Router 会根据文件系统自动生成所有路由
	return routes
}

// buildRouteFromMenu 根据菜单构建路由
func (s *RouteGenService) buildRouteFromMenu(menu *modelsystem.Menu, hasChildren bool, isSuper bool) map[string]interface{} {
	// 路由名称（从路径生成）
	routeName := s.generateRouteName(menu.Path)

	// 组件名称（根据是否有子菜单和层级生成）
	component := s.generateComponent(menu.Path, menu.ParentID, hasChildren)

	route := map[string]interface{}{
		"id":   strconv.FormatUint(uint64(menu.ID), 10),
		"name": routeName,
		"path": menu.Path,
		"meta": map[string]interface{}{
			"title":   menu.Name,
			"i18nKey": "route." + routeName,
			"order":   menu.Sort,
		},
	}

	// 添加图标（如果有）
	if menu.Icon != "" {
		route["meta"].(map[string]interface{})["icon"] = menu.Icon
	}

	// 添加组件（如果有）
	if component != "" {
		route["component"] = component
	}

	// 添加权限标识（如果有）
	if menu.Permission != "" {
		route["meta"].(map[string]interface{})["permission"] = menu.Permission
	}

	// 特殊处理：Web终端 在新窗口打开（基于路径判断，避免ID变化导致的失效）
	if menu.Path == "/webterminal" {
		route["meta"].(map[string]interface{})["href"] = menu.Path
		route["meta"].(map[string]interface{})["hideInMenu"] = false
	}

	return route
}

// generateRouteName 根据路径生成路由名称
func (s *RouteGenService) generateRouteName(path string) string {
	// 特殊处理：保留某些路径的原始格式
	if path == "/user-center" {
		return "user-center"
	}

	// 移除前导斜杠
	name := path
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	// 将路径分隔符 / 替换为下划线，保留连字符 -
	// 例如：/auth/user-identities -> auth_user-identities
	name = strings.ReplaceAll(name, "/", "_")
	return name
}

// generateComponent 根据路径生成组件名称
func (s *RouteGenService) generateComponent(path string, parentID uint, hasChildren bool) string {
	routeName := s.generateRouteName(path)

	logger.Debug("[路由组件生成]",
		zap.String("path", path),
		zap.String("routeName", routeName),
		zap.Uint("parentID", parentID),
		zap.Bool("hasChildren", hasChildren))

	// 特殊处理：web终端使用独立布局（无导航栏）
	if path == "/webterminal" {
		// 使用 terminalLayout 布局，提供纯终端界面体验
		// 注意：命名需与前端 Elegant Router 自动生成的布局名称一致
		return "layout.terminalLayout$view." + routeName
	}

	// 如果是一级菜单（父级为0）
	if parentID == 0 {
		// 如果有子菜单，返回布局容器
		if hasChildren {
			return "layout.base"
		}
		// 如果没有子菜单（单页面），返回完整组件路径
		return "layout.base$view." + routeName
	}
	// 二级及以下菜单：如果有子菜单（目录类型），不设置 component，依靠 redirect 跳转到第一个子路由
	if hasChildren {
		return ""
	}
	// 叶子菜单使用 view 前缀（继承父路由布局）
	return "view." + routeName
}
