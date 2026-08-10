package system

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"oneops/backend3/pkg/utils"
	"oneops/backend3/service/system"
)

// RouteController 路由控制器
type RouteController struct {
	svc *system.RouteGenService
}

// NewRouteController 创建路由控制器
func NewRouteController(svc *system.RouteGenService) *RouteController {
	return &RouteController{svc: svc}
}

// GetConstantRoutes 获取常量路由（无需登录即可访问的路由）
func (ctrl *RouteController) GetConstantRoutes(ctx *gin.Context) {
	constantRoutes := ctrl.svc.GetConstantRoutes()
	ctx.JSON(http.StatusOK, utils.SuccessWithData(constantRoutes))
}

// GetUserRoutes 获取用户路由（需要登录，根据用户角色返回）
func (ctrl *RouteController) GetUserRoutes(ctx *gin.Context) {
	// 从上下文获取用户信息（通过 Auth 中间件设置）
	userIDUint, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return
	}

	routes, err := ctrl.svc.GetUserRoutes(userIDUint)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取菜单失败"))
		return
	}

	// 返回路由和首页
	result := map[string]interface{}{
		"routes": routes,
		"home":   "home", // 默认首页
	}
	ctx.JSON(http.StatusOK, utils.SuccessWithData(result))
}

// IsRouteExist 检查路由是否存在
func (ctrl *RouteController) IsRouteExist(ctx *gin.Context) {
	routeName := ctx.Query("routeName")
	if routeName == "" {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("路由名称不能为空"))
		return
	}

	// 获取用户信息
	userIDUint, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return
	}

	exist, err := ctrl.svc.IsRouteExist(userIDUint, routeName)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取用户权限失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(exist))
}

// InvalidateCache 清除RBAC缓存并重新同步菜单
func (ctrl *RouteController) InvalidateCache(ctx *gin.Context) {
	if err := ctrl.svc.InvalidateCache(); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("菜单同步失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData("缓存已清除并重新同步菜单数据"))
}

// DebugCache 调试当前缓存内容
func (ctrl *RouteController) DebugCache(ctx *gin.Context) {
	userIDUint, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return
	}

	result, err := ctrl.svc.DebugCache(userIDUint)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取缓存失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(result))
}
