package controllers

import (
	"net/http"
	"oneops/backend/logger"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteController 路由控制器
type RouteController struct{}

// NewRouteController 创建路由控制器
func NewRouteController() *RouteController {
	return &RouteController{}
}

// GetConstantRoutes 获取常量路由（无需登录即可访问的路由）
func (c *RouteController) GetConstantRoutes(ctx *gin.Context) {
	// 常量路由：403、404、500、登录页、iframe-page 等
	constantRoutes := []map[string]interface{}{
		{
			"id":   "403",
			"name": "403",
			"path": "/403",
			"component": "layout.blank$view.403",
			"meta": map[string]interface{}{
				"title":       "403",
				"i18nKey":     "route.403",
				"constant":    true,
				"hideInMenu":  true,
			},
		},
		{
			"id":   "404",
			"name": "404",
			"path": "/404",
			"component": "layout.blank$view.404",
			"meta": map[string]interface{}{
				"title":       "404",
				"i18nKey":     "route.404",
				"constant":    true,
				"hideInMenu":  true,
			},
		},
		{
			"id":   "500",
			"name": "500",
			"path": "/500",
			"component": "layout.blank$view.500",
			"meta": map[string]interface{}{
				"title":       "500",
				"i18nKey":     "route.500",
				"constant":    true,
				"hideInMenu":  true,
			},
		},
		{
			"id":   "login",
			"name": "login",
			"path": "/login/:module(pwd-login|code-login|register|reset-pwd|bind-wechat)?",
			"component": "layout.blank$view.login",
			"props": true,
			"meta": map[string]interface{}{
				"title":       "login",
				"i18nKey":     "route.login",
				"constant":    true,
				"hideInMenu":  true,
			},
		},
		{
			"id":   "iframe-page",
			"name": "iframe-page",
			"path": "/iframe-page/:url",
			"component": "layout.base$view.iframe-page",
			"props": true,
			"meta": map[string]interface{}{
				"title":       "iframe-page",
				"i18nKey":     "route.iframe-page",
				"constant":    true,
				"hideInMenu":  true,
				"keepAlive":   true,
			},
		},
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(constantRoutes))
}

// GetUserRoutes 获取用户路由（需要登录，根据用户角色返回）
func (c *RouteController) GetUserRoutes(ctx *gin.Context) {
	// 从上下文获取用户信息（通过 Auth 中间件设置）
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("未授权"))
		return
	}

	userIDUint := userID.(uint)

	// 🐛 调试：显示路由生成开始
	logger.Info("🐛 [GetUserRoutes] 开始生成路由", zap.Uint("userID", userIDUint))

	// 清除该用户的缓存，确保路由配置每次都是最新的
	// 因为路由配置（component字段）需要动态生成，缓存会导致代码修改不生效

	rbacService := services.NewRBACService()
	menuTree, _, roles, err := rbacService.BuildMenuTreeAndPermissions(userIDUint)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取菜单失败"))
		return
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
	routes := c.convertMenusToRoutes(menuTree, isSuper)

		// 添加隐藏路由到对应的父路由下
		routes = c.appendHiddenRoutesToMenuTree(routes)

		// 🐛 调试：显示webterminal路由的最终配置
		for _, route := range routes {
			if path, ok := route["path"].(string); ok && path == "/webterminal" {
				component, _ := route["component"].(string)
				logger.Info("🐛 [GetUserRoutes] 返回给前端的webterminal路由",
					zap.String("path", path),
					zap.String("component", component),
					zap.Any("full_route", route))
				break
			}
		}

	// 返回路由和首页
	result := map[string]interface{}{
		"routes": routes,
		"home":   "home", // 默认首页
	}
	ctx.JSON(http.StatusOK, utils.SuccessWithData(result))
}

// IsRouteExist 检查路由是否存在
func (c *RouteController) IsRouteExist(ctx *gin.Context) {
	routeName := ctx.Query("routeName")
	if routeName == "" {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("路由名称不能为空"))
		return
	}

	// 获取用户信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	// 检查是否为管理员
	rbacService := services.NewRBACService()
	roles, err := rbacService.GetUserRoles(userID.(uint))
	if err == nil {
		for _, role := range roles {
			if role.Code == "admin" {
				// 管理员可以访问所有路由
				ctx.JSON(http.StatusOK, utils.SuccessWithData(true))
				return
			}
		}
	}

	// 获取用户权限列表
	_, permissions, _, err := rbacService.BuildMenuTreeAndPermissions(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取用户权限失败: " + err.Error()))
		return
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
				ctx.JSON(http.StatusOK, utils.SuccessWithData(true))
				return
			}
		}
		ctx.JSON(http.StatusOK, utils.SuccessWithData(false))
		return
	}

	// 检查父路由权限
	if parentRoute != "" {
		for _, perm := range permissions {
			if perm == parentRoute {
				ctx.JSON(http.StatusOK, utils.SuccessWithData(true))
				return
			}
		}
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(false))
}

// InvalidateCache 清除RBAC缓存
func (c *RouteController) InvalidateCache(ctx *gin.Context) {
	// 先清除缓存

	// 然后重新同步菜单，确保数据最新
	initService := services.NewInitService()
	if err := initService.SyncMenus(); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("菜单同步失败: "+err.Error()))
		return
	}

	// 再次清除缓存，确保新菜单权限立即生效

	ctx.JSON(http.StatusOK, utils.SuccessWithData("缓存已清除并重新同步菜单数据"))
}

// DebugCache 调试当前缓存内容
func (c *RouteController) DebugCache(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusOK, utils.ErrorUnauthorized("用户未登录"))
		return
	}

	rbacService := services.NewRBACService()
	menuTree, permissions, roles, err := rbacService.BuildMenuTreeAndPermissions(userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取缓存失败: "+err.Error()))
		return
	}

	// 检查是否包含 webterminal
	hasWebTerminal := false
	for _, menu := range menuTree {
		if menu.Path == "/webterminal" {
			hasWebTerminal = true
			break
		}
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(map[string]interface{}{
		"has_webterminal":     hasWebTerminal,
		"menu_count":          len(menuTree),
		"permission_count":    len(permissions),
		"role_count":          len(roles),
		"menus":               menuTree,
	}))
}

// convertMenusToRoutes 将数据库菜单转换为前端路由格式
func (c *RouteController) convertMenusToRoutes(menus []*models.Menu, isSuper bool) []map[string]interface{} {
	routes := make([]map[string]interface{}, 0)

	for _, menu := range menus {
		// 检查是否有子菜单
		hasChildren := len(menu.Children) > 0

		route := c.buildRouteFromMenu(menu, hasChildren, isSuper)

		// 如果有子菜单，递归处理
		if hasChildren {
			childrenRoutes := c.convertMenusToRoutes(menu.Children, isSuper)
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
func (c *RouteController) appendHiddenRoutesToMenuTree(routes []map[string]interface{}) []map[string]interface{} {

	// 不需要添加任何硬编码路由，前端 Elegant Router 会根据文件系统自动生成所有路由
	return routes
}

// buildRouteFromMenu 根据菜单构建路由
func (c *RouteController) buildRouteFromMenu(menu *models.Menu, hasChildren bool, isSuper bool) map[string]interface{} {
	// 路由名称（从路径生成）
	routeName := c.generateRouteName(menu.Path)

	// 组件名称（根据是否有子菜单和层级生成）
	component := c.generateComponent(menu.Path, menu.ParentID, hasChildren)

	route := map[string]interface{}{
		"id":    strconv.FormatUint(uint64(menu.ID), 10),
		"name":  routeName,
		"path":  menu.Path,
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

	// 🐛 调试信息：检查webterminal路由处理
	if menu.Path == "/webterminal" {
		logger.Info("🐛 [Web终端路由] 开始构建",
			zap.String("path", menu.Path),
			zap.Uint("id", menu.ID),
			zap.String("name", menu.Name),
			zap.String("component", component))
	}

	// 特殊处理：Web终端 在新窗口打开（基于路径判断，避免ID变化导致的失效）
	if menu.Path == "/webterminal" {
		logger.Info("🐛 [Web终端路由] 设置href属性", zap.String("href", menu.Path))
		route["meta"].(map[string]interface{})["href"] = menu.Path
		route["meta"].(map[string]interface{})["hideInMenu"] = false

		// 🐛 调试：显示最终的路由配置
		logger.Info("🐛 [Web终端路由] 最终配置",
			zap.String("component", component),
			zap.Any("route", route))
	}

	return route
}

// generateRouteName 根据路径生成路由名称
func (c *RouteController) generateRouteName(path string) string {
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
func (c *RouteController) generateComponent(path string, parentID uint, hasChildren bool) string {
	routeName := c.generateRouteName(path)

	// 🐛 调试信息：显示路由组件生成过程
	logger.Debug("🐛 [路由组件生成]",
		zap.String("path", path),
		zap.String("routeName", routeName),
		zap.Uint("parentID", parentID),
		zap.Bool("hasChildren", hasChildren))

	// 特殊处理：web终端使用独立布局（无导航栏）
	if path == "/webterminal" {
		// 使用 terminalLayout 布局，提供纯终端界面体验
		// 注意：命名需与前端 Elegant Router 自动生成的布局名称一致
		component := "layout.terminalLayout$view." + routeName
		logger.Info("🐛 [Web终端] 使用terminalLayout布局",
			zap.String("path", path),
			zap.String("component", component))
		return component
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
