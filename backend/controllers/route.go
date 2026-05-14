package controllers

import (
	"net/http"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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

	// 获取用户信息和角色
	userIDUint, _ := strconv.ParseUint(strconv.FormatUint(uint64(userID.(uint)), 10), 10, 32)
	rbacService := services.NewRBACService()
	roles, err := rbacService.GetUserRoles(uint(userIDUint))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取用户角色失败"))
		return
	}

	// 构建角色代码列表
	roleCodes := make([]string, len(roles))
	for i, role := range roles {
		roleCodes[i] = role.Code
	}

	// 检查是否是超级管理员
	isSuper := false
	for _, code := range roleCodes {
		if code == "R_SUPER" || code == "admin" {
			isSuper = true
			break
		}
	}

	// 获取用户菜单
	menuTree, _, err := rbacService.BuildMenuTreeAndPermissions(uint(userIDUint))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取菜单失败"))
		return
	}

	// 将菜单转换为前端路由格式
	routes := c.convertMenusToRoutes(menuTree, isSuper)

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

	// 这里可以根据实际需求实现路由存在性检查
	// 目前简单返回 true
	ctx.JSON(http.StatusOK, utils.SuccessWithData(true))
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
	// 使用下划线替换路径分隔符 / 和连字符 -，与前端路由命名保持一致
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// generateComponent 根据路径生成组件名称
func (c *RouteController) generateComponent(path string, parentID uint, hasChildren bool) string {
	routeName := c.generateRouteName(path)

	// 如果是一级菜单（父级为0）
	if parentID == 0 {
		// 如果有子菜单，返回布局容器
		if hasChildren {
			return "layout.base"
		}
		// 如果没有子菜单（单页面），返回完整组件路径
		return "layout.base$view." + routeName
	}
	// 二级及以下菜单使用 view 前缀
	return "view." + routeName
}
