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

// GetConstantRoutes godoc
// @Summary      获取常量路由
// @Description  返回无需登录即可访问的前端常量路由
// @Tags         系统管理-路由
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}
// @Router       /route/getConstantRoutes [get]
func (ctrl *RouteController) GetConstantRoutes(ctx *gin.Context) {
	constantRoutes := ctrl.svc.GetConstantRoutes()
	ctx.JSON(http.StatusOK, utils.SuccessWithData(constantRoutes))
}

// GetUserRoutes godoc
// @Summary      获取用户路由
// @Description  根据当前登录用户的角色，返回其可访问的动态路由及首页配置
// @Tags         系统管理-路由
// @Produce      json
// @Success      200  {object}  utils.Response{data=object{routes=object,home=string}}
// @Failure      200  {object}  utils.Response  "获取菜单失败"
// @Router       /route/getUserRoutes [get]
// @Security     BearerAuth
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

// IsRouteExist godoc
// @Summary      检查路由是否存在
// @Description  检查当前用户可访问的路由中是否存在指定名称的路由
// @Tags         系统管理-路由
// @Produce      json
// @Param        routeName  query     string  true  "路由名称"
// @Success      200  {object}  utils.Response{data=bool}
// @Failure      200  {object}  utils.Response  "路由名称不能为空 / 获取用户权限失败"
// @Router       /route/isRouteExist [get]
// @Security     BearerAuth
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

// InvalidateCache godoc
// @Summary      清除路由缓存
// @Description  清除 RBAC 路由缓存并重新同步菜单数据
// @Tags         系统管理-路由
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "菜单同步失败"
// @Router       /route/invalidateCache [post]
// @Security     BearerAuth
func (ctrl *RouteController) InvalidateCache(ctx *gin.Context) {
	if err := ctrl.svc.InvalidateCache(); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("菜单同步失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("缓存已清除并重新同步菜单数据"))
}

// DebugCache godoc
// @Summary      调试路由缓存
// @Description  返回当前用户路由缓存的详细内容（仅用于调试）
// @Tags         系统管理-路由
// @Produce      json
// @Success      200  {object}  utils.Response{data=object}
// @Failure      200  {object}  utils.Response  "获取缓存失败"
// @Router       /route/debugCache [get]
// @Security     BearerAuth
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
