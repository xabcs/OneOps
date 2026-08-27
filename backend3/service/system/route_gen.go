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
		// k8s Pod 终端：新窗口打开的独立工作台（TerminalLayout 无侧边栏），携带 token 直连 ws，免登录
		// 注意：name 用连字符（无下划线）才会被前端 transform 视为单层路由，
		// 从而把 "layout.terminalLayout$view.k8s_terminal" 按 $ 拆分为布局+视图
		{
			"id":        "k8s_terminal",
			"name":      "k8s-terminal",
			"path":      "/k8s/terminal",
			"component": "layout.terminalLayout$view.k8s_terminal",
			"meta": map[string]interface{}{
				"title":      "Pod终端",
				"hideInMenu": true,
				"constant":   true,
			},
		},
	}
}

// GetUserRoutes 获取用户路由（需要登录，根据用户角色返回）
// 返回生成的路由列表
func (s *RouteGenService) GetUserRoutes(userID uint) ([]map[string]interface{}, error) {
	// 清除该用户的缓存，确保路由配置每次都是最新的
	// 因为路由配置（component字段）需要动态生成，缓存会导致代码修改不生效

	logger.Info("[GetUserRoutes] 开始生成路由", zap.Uint("user_id", userID))

	permSvc, err := GetPermissionService()
	if err != nil {
		return nil, err
	}
	menuTree, _, roles, err := permSvc.BuildMenuTreeAndPermissions(userID)
	if err != nil {
		return nil, err
	}

	// 检查是否是超级管理员
	isSuper := permSvc.IsAdmin(roles)

	// 将菜单转换为前端路由格式
	routes := s.convertMenusToRoutes(menuTree, isSuper)

	// web 终端入口与主机管理菜单强耦合（有主机管理菜单才可见终端）
	routes = s.appendTerminalRoute(routes, menuTree)

	// 隐藏详情页路由下发（可见性跟随锚点菜单）
	routes = s.appendHiddenDetailRoutes(routes, menuTree)

	return routes, nil
}

// hiddenDetailRoute 隐藏详情路由登记项
// 组件映射由前端 gen-route 按 views/<...>/detail/index.vue 约定生成，
// 路由本体在此登记并由后端下发（官方动态路由模式，替代前端 custom-routes 手写）
type hiddenDetailRoute struct {
	Name       string // 路由名 = 前端 imports.ts 的视图键
	Path       string
	Title      string
	AnchorPath string // 锚点菜单路径：用户菜单含该路径（能看到列表页）才下发对应详情路由
	ActiveMenu string // 菜单高亮目标（列表页路由名）
}

// hiddenDetailRoutes 隐藏详情路由登记表（新增详情页：前端建 detail/index.vue + gen-route，再在此登记）
var hiddenDetailRoutes = []hiddenDetailRoute{
	{Name: "cmdb_servers_detail", Path: "/cmdb/servers/detail", Title: "主机详情", AnchorPath: "/cmdb/servers", ActiveMenu: "cmdb_servers"},
	{Name: "monitoring_servers_detail", Path: "/monitoring/servers/detail", Title: "主机监控详情", AnchorPath: "/monitoring/servers", ActiveMenu: "monitoring_servers"},
	{Name: "k8s_resources_deployments_detail", Path: "/k8s/resources/deployments/detail", Title: "Deployment详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_statefulsets_detail", Path: "/k8s/resources/statefulsets/detail", Title: "StatefulSet详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_daemonsets_detail", Path: "/k8s/resources/daemonsets/detail", Title: "DaemonSet详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_pods_detail", Path: "/k8s/resources/pods/detail", Title: "Pod详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_jobs_detail", Path: "/k8s/resources/jobs/detail", Title: "Job详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_cronjobs_detail", Path: "/k8s/resources/cronjobs/detail", Title: "CronJob详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_configmaps_detail", Path: "/k8s/resources/configmaps/detail", Title: "ConfigMap详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_secrets_detail", Path: "/k8s/resources/secrets/detail", Title: "Secret详情", AnchorPath: "/k8s/workloads", ActiveMenu: "k8s_workloads"},
	{Name: "k8s_resources_services_detail", Path: "/k8s/resources/services/detail", Title: "Service详情", AnchorPath: "/k8s/network", ActiveMenu: "k8s_network"},
	{Name: "k8s_resources_ingresses_detail", Path: "/k8s/resources/ingresses/detail", Title: "Ingress详情", AnchorPath: "/k8s/network", ActiveMenu: "k8s_network"},
	{Name: "ticket_center_detail", Path: "/ticket/center/detail", Title: "工单详情", AnchorPath: "/ticket/center", ActiveMenu: "ticket_center"},
}

// appendHiddenDetailRoutes 按锚点菜单追加隐藏详情路由
// 详情路由挂在路径前缀匹配的一级路由（layout.base）之下，
// 由前端 transform 嵌入基础布局渲染；hideInMenu 保证不进入菜单
func (s *RouteGenService) appendHiddenDetailRoutes(routes []map[string]interface{}, menuTree []*modelsystem.Menu) []map[string]interface{} {
	for _, spec := range hiddenDetailRoutes {
		// 菜单表已显式配置同名路由时以下发数据为准，避免重复
		if routeTreeContainsName(routes, spec.Name) {
			continue
		}
		// 可见性跟随锚点菜单：无锚点（看不到列表页）则不发放对应详情路由
		if !menuTreeContainsPath(menuTree, spec.AnchorPath) {
			continue
		}
		parent := findFirstLevelRouteByPathPrefix(routes, spec.Path)
		if parent == nil {
			logger.Warn("[appendHiddenDetailRoutes] 未找到一级父路由，跳过", zap.String("path", spec.Path))
			continue
		}
		hidden := map[string]interface{}{
			"name":      spec.Name,
			"path":      spec.Path,
			"component": "view." + spec.Name,
			"meta": map[string]interface{}{
				"title":      spec.Title,
				"hideInMenu": true,
				"activeMenu": spec.ActiveMenu,
			},
		}
		children, _ := parent["children"].([]map[string]interface{})
		parent["children"] = append(children, hidden)
	}
	return routes
}

// findFirstLevelRouteByPathPrefix 在一级路由中查找路径前缀匹配的布局路由
func findFirstLevelRouteByPathPrefix(routes []map[string]interface{}, path string) map[string]interface{} {
	for _, r := range routes {
		p, ok := r["path"].(string)
		if !ok || p == "" || p == "/" {
			continue
		}
		if strings.HasPrefix(path, p+"/") {
			return r
		}
	}
	return nil
}

// routeTreeContainsName 递归检查路由树中是否已存在同名路由
func routeTreeContainsName(routes []map[string]interface{}, name string) bool {
	for _, r := range routes {
		if r["name"] == name {
			return true
		}
		if children, ok := r["children"].([]map[string]interface{}); ok && routeTreeContainsName(children, name) {
			return true
		}
	}
	return false
}

// frameworkRouteNames 前端框架内置路由名（router/routes/builtin.ts 固定生成，非业务配置）
var frameworkRouteNames = map[string]bool{"root": true, "not-found": true}

// derivedRouteNames 派生路由名（不在菜单表，由 GetUserRoutes 按业务规则追加）
// webterminal 随主机管理菜单派生：页面全局存在，无权限直访时应显示 403 而非 404
var derivedRouteNames = map[string]bool{"webterminal": true}

// IsRouteExist 检查路由在系统中是否全局存在（与用户权限无关）
// 用于前端路由守卫区分「路由不存在(404)」与「路由存在但无访问权限(403)」
func (s *RouteGenService) IsRouteExist(userID uint, routeName string) (bool, error) {
	// 常量路由：与 GetConstantRoutes 同源，新增常量路由自动被识别
	for _, r := range s.GetConstantRoutes() {
		if r["name"] == routeName {
			return true, nil
		}
	}
	if frameworkRouteNames[routeName] || derivedRouteNames[routeName] {
		return true, nil
	}

	// 隐藏详情路由：全局登记，存在性以此为准（权限可见性由锚点菜单另行控制）
	for _, spec := range hiddenDetailRoutes {
		if spec.Name == routeName {
			return true, nil
		}
	}

	// 业务路由：以启用菜单为唯一事实来源
	var paths []string
	if err := s.db.Model(&modelsystem.Menu{}).Where("status = 1").Pluck("path", &paths).Error; err != nil {
		return false, err
	}

	for _, p := range paths {
		if s.generateRouteName(p) == routeName {
			return true, nil
		}
	}
	return false, nil
}

// InvalidateCache 清除RBAC缓存并重新同步菜单
func (s *RouteGenService) InvalidateCache() error {
	// 重新同步菜单，确保数据最新
	initializer := NewInitializer()
	return initializer.SyncMenus()
}

// DebugCache 调试当前缓存内容
func (s *RouteGenService) DebugCache(userID uint) (map[string]interface{}, error) {
	permSvc, err := GetPermissionService()
	if err != nil {
		return nil, err
	}
	menuTree, permissions, roles, err := permSvc.BuildMenuTreeAndPermissions(userID)
	if err != nil {
		return nil, err
	}

	// 检查是否包含 webterminal（派生路由：随主机管理菜单出现）
	hasWebTerminal := menuTreeContainsPath(menuTree, "/cmdb/servers")

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

// appendTerminalRoute 追加 web 终端路由（派生入口，不进菜单表）
// 规则：用户菜单树包含主机管理（/cmdb/servers）时自动获得终端工作台入口，
// 使「无主机管理菜单则不可见终端」在结构上成立，管理员无需单独绑定终端菜单
func (s *RouteGenService) appendTerminalRoute(routes []map[string]interface{}, menuTree []*modelsystem.Menu) []map[string]interface{} {
	for _, r := range routes {
		if r["name"] == "webterminal" {
			return routes
		}
	}
	if !menuTreeContainsPath(menuTree, "/cmdb/servers") {
		return routes
	}
	return append(routes, map[string]interface{}{
		"id":        "webterminal",
		"name":      "webterminal",
		"path":      "/webterminal",
		"component": "layout.terminalLayout$view.webterminal",
		"meta": map[string]interface{}{
			"title":   "web终端",
			"i18nKey": "route.webterminal",
			"icon":    "mdi:console",
			"order":   7,
			"href":    "/webterminal", // 新标签页打开；href === path 时在当前窗口渲染
		},
	})
}

// menuTreeContainsPath 递归检查菜单树中是否存在指定路径
func menuTreeContainsPath(menus []*modelsystem.Menu, path string) bool {
	for _, menu := range menus {
		if menu.Path == path {
			return true
		}
		if len(menu.Children) > 0 && menuTreeContainsPath(menu.Children, path) {
			return true
		}
	}
	return false
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
		zap.String("route_name", routeName),
		zap.Uint("parent_id", parentID),
		zap.Bool("has_children", hasChildren))

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
