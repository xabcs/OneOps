package system

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	"oneops/backend3/service/system"
)

// PermissionController 权限控制器
type PermissionController struct {
	svc *system.PermissionService
}

// NewPermissionController 创建权限控制器实例
func NewPermissionController(svc *system.PermissionService) *PermissionController {
	return &PermissionController{svc: svc}
}

// GetPermissionOptions godoc
// @Summary      获取权限选项列表
// @Description  不分页返回全部权限的精简选项（id/名称/编码/模块/资源/动作/层级），用于角色授权选择器与权限树
// @Tags         系统管理-权限
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]object}  "权限选项列表"
// @Failure      200  {object}  utils.Response  "获取权限选项失败"
// @Router       /system/permissions/options [get]
// @Security     BearerAuth
func (ctrl *PermissionController) GetPermissionOptions(ctx *gin.Context) {
	permissions, err := ctrl.svc.GetAllPermissionOptions()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取权限选项失败"))
		return
	}

	// 转换为精简格式
	options := make([]gin.H, len(permissions))
	for i, p := range permissions {
		options[i] = gin.H{
			"id":        p.ID,
			"name":      p.Name,
			"code":      p.Code,
			"module":    p.Module,
			"resource":  p.Resource,
			"action":    p.Action,
			"parentId":  p.ParentID,
			"level":     p.Level,
			"sortOrder": p.SortOrder,
		}
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(options))
}

// GetPermissionTree godoc
// @Summary      获取权限树
// @Description  返回全部权限的树形结构
// @Tags         系统管理-权限
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]modelsystem.PermissionTreeNode}
// @Failure      200  {object}  utils.Response  "获取权限树失败"
// @Router       /system/permissions/tree [get]
// @Security     BearerAuth
func (ctrl *PermissionController) GetPermissionTree(ctx *gin.Context) {
	tree, err := ctrl.svc.GetPermissionTree()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取权限树失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(tree))
}

// GetPermissionList godoc
// @Summary      获取权限列表
// @Description  分页获取权限列表，支持按名称、编码、模块、状态搜索；兼容 current/size 和 page/pageSize 两种分页参数
// @Tags         系统管理-权限
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(10)
// @Param        name      query     string  false  "权限名称"
// @Param        code      query     string  false  "权限编码"
// @Param        module    query     string  false  "所属模块"
// @Param        status    query     string  false  "状态"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "获取权限列表失败"
// @Router       /system/permissions [get]
// @Security     BearerAuth
func (ctrl *PermissionController) GetPermissionList(ctx *gin.Context) {
	// 获取分页参数（兼容 current/size 和 page/pageSize 两种参数名）
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", ctx.DefaultQuery("current", "1")))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", ctx.DefaultQuery("size", "10")))

	// 获取搜索参数
	name := ctx.Query("name")
	code := ctx.Query("code")
	module := ctx.Query("module")
	status := ctx.Query("status")

	// 构建查询条件
	query := map[string]interface{}{}
	if name != "" {
		query["name"] = name
	}
	if code != "" {
		query["code"] = code
	}
	if module != "" {
		query["module"] = module
	}
	if status != "" {
		query["status"] = status
	}

	// 获取权限列表
	permissions, total, err := ctrl.svc.GetPermissionList(page, pageSize, query)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取权限列表失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(permissions, total, dto.BasePageQuery{Page: page, PageSize: pageSize})))
}

// CreatePermission godoc
// @Summary      创建权限
// @Description  新增一个权限节点
// @Tags         系统管理-权限
// @Accept       json
// @Produce      json
// @Param        body  body      modelsystem.CreatePermissionRequest  true  "权限信息"
// @Success      200   {object}  utils.Response{data=modelsystem.Permission}
// @Failure      200   {object}  utils.Response  "参数错误 / 创建权限失败"
// @Router       /system/permissions [post]
// @Security     BearerAuth
func (ctrl *PermissionController) CreatePermission(ctx *gin.Context) {
	var req modelsystem.CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	permission := &modelsystem.Permission{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Module:      req.Module,
		Resource:    req.Resource,
		Action:      req.Action,
		Level:       req.Level,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		Status:      1, // 默认启用
	}

	if err := ctrl.svc.CreatePermission(permission); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("创建权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permission))
}

// UpdatePermission godoc
// @Summary      更新权限
// @Description  按权限 ID 更新权限信息
// @Tags         系统管理-权限
// @Accept       json
// @Produce      json
// @Param        id    path      int                              true  "权限 ID"
// @Param        body  body      modelsystem.UpdatePermissionRequest  true  "待更新字段"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的权限ID / 参数错误 / 更新权限失败"
// @Router       /system/permissions/{id} [put]
// @Security     BearerAuth
func (ctrl *PermissionController) UpdatePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的权限ID"))
		return
	}

	var req modelsystem.UpdatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}

	permission := &modelsystem.Permission{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Level:       req.Level,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}

	if err := ctrl.svc.UpdatePermission(permission); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("更新权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeletePermission godoc
// @Summary      删除权限
// @Description  按权限 ID 删除权限
// @Tags         系统管理-权限
// @Produce      json
// @Param        id  path      int  true  "权限 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "无效的权限ID / 删除权限失败"
// @Router       /system/permissions/{id} [delete]
// @Security     BearerAuth
func (ctrl *PermissionController) DeletePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的权限ID"))
		return
	}

	if err := ctrl.svc.DeletePermission(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("删除权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// GetRolePermissions godoc
// @Summary      获取角色权限
// @Description  返回指定角色已分配的权限 ID 列表
// @Tags         系统管理-权限
// @Produce      json
// @Param        roleId  path      int  true  "角色 ID"
// @Success      200  {object}  utils.Response{data=[]uint}
// @Failure      200  {object}  utils.Response  "无效的角色ID / 获取角色权限失败"
// @Router       /system/roles/{roleId}/permissions [get]
// @Security     BearerAuth
func (ctrl *PermissionController) GetRolePermissions(ctx *gin.Context) {
	roleIDStr := ctx.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	permissions, err := ctrl.svc.GetRolePermissions(uint(roleID))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取角色权限失败"))
		return
	}

	// 提取权限ID数组
	permissionIDs := make([]uint, len(permissions))
	for i, perm := range permissions {
		permissionIDs[i] = perm.ID
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(permissionIDs))
}

// AssignRolePermissions godoc
// @Summary      分配角色权限
// @Description  为指定角色批量分配权限（覆盖原有权限）
// @Tags         系统管理-权限
// @Accept       json
// @Produce      json
// @Param        roleId  path      int                                        true  "角色 ID"
// @Param        body    body      modelsystem.AssignRolePermissionsRequest  true  "权限 ID 列表"
// @Success      200     {object}  utils.Response
// @Failure      200     {object}  utils.Response  "无效的角色ID / 请求参数错误 / 分配权限失败"
// @Router       /system/roles/{roleId}/permissions [post]
// @Security     BearerAuth
func (ctrl *PermissionController) AssignRolePermissions(ctx *gin.Context) {
	// 从URL路径参数获取角色ID（符合项目规范）
	roleIDStr := ctx.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	// 从请求体绑定参数（符合项目规范）
	var req modelsystem.AssignRolePermissionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	// 调用服务层
	if err := ctrl.svc.BatchAssignPermissions(uint(roleID), req.PermissionIDs); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("分配权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("权限分配成功"))
}

// GetUserPermissions godoc
// @Summary      获取当前用户权限
// @Description  返回当前登录用户拥有的权限列表
// @Tags         系统管理-权限
// @Produce      json
// @Success      200  {object}  utils.Response{data=object{permissions=[]string}}
// @Failure      200  {object}  utils.Response  "获取用户权限失败"
// @Router       /system/user/permissions [get]
// @Security     BearerAuth
func (ctrl *PermissionController) GetUserPermissions(ctx *gin.Context) {
	// 从上下文获取用户ID（假设由认证中间件设置）
	userID, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return
	}

	permissions, err := ctrl.svc.GetUserPermissions(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取用户权限失败"))
		return
	}

	responseData := gin.H{
		"permissions": permissions,
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(responseData))
}

// CheckPermission godoc
// @Summary      检查权限
// @Description  检查指定用户是否拥有某项权限
// @Tags         系统管理-权限
// @Accept       json
// @Produce      json
// @Param        body  body      modelsystem.CheckPermissionRequest  true  "权限检查请求"
// @Success      200   {object}  utils.Response{data=modelsystem.CheckPermissionResponse}
// @Failure      200   {object}  utils.Response  "参数错误 / 权限检查失败"
// @Router       /system/permissions/check [post]
// @Security     BearerAuth
func (ctrl *PermissionController) CheckPermission(ctx *gin.Context) {
	var req modelsystem.CheckPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
		return
	}

	allowed, err := ctrl.svc.HasPermission(req.UserID, req.Permission)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("权限检查失败"))
		return
	}

	response := modelsystem.CheckPermissionResponse{
		Allowed: allowed,
	}

	if !allowed {
		response.Reason = "权限不足"
	}

	ctx.JSON(http.StatusOK, utils.SuccessWithData(response))
}

// GetPermissionRoutes godoc
// @Summary      获取权限路由映射
// @Description  查询权限码与 API 路由的映射列表，可按权限码过滤
// @Tags         系统管理-权限
// @Produce      json
// @Param        permissionCode  query  string  false  "权限编码（为空返回全部）"
// @Success      200  {object}  utils.Response{data=[]modelsystem.PermissionRoute}
// @Failure      200  {object}  utils.Response  "获取权限路由映射失败"
// @Router       /system/permissions/routes [get]
// @Security     BearerAuth
func (ctrl *PermissionController) GetPermissionRoutes(ctx *gin.Context) {
	routes, err := ctrl.svc.GetPermissionRoutes(ctx.Query("permissionCode"))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("获取权限路由映射失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.SuccessWithData(routes))
}

// CreatePermissionRoute godoc
// @Summary      新增权限路由映射
// @Description  为一个 API 路由配置权限码（同一端点只能归属一个权限码），即时生效
// @Tags         系统管理-权限
// @Accept       json
// @Produce      json
// @Param        body  body  modelsystem.CreatePermissionRouteRequest  true  "映射信息"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "参数错误 / 权限码不存在 / 路由已存在映射"
// @Router       /system/permissions/routes [post]
// @Security     BearerAuth
func (ctrl *PermissionController) CreatePermissionRoute(ctx *gin.Context) {
	var req modelsystem.CreatePermissionRouteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}
	if err := ctrl.svc.CreatePermissionRoute(req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("创建成功"))
}

// UpdatePermissionRoute godoc
// @Summary      修改权限路由映射
// @Description  修改映射归属的权限码（端点 method/path 不可改，删除后重建）
// @Tags         系统管理-权限
// @Accept       json
// @Produce      json
// @Param        id    path  int                                         true  "映射ID"
// @Param        body  body  modelsystem.UpdatePermissionRouteRequest  true  "目标权限码"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "参数错误 / 权限码不存在 / 映射不存在"
// @Router       /system/permissions/routes/{id} [put]
// @Security     BearerAuth
func (ctrl *PermissionController) UpdatePermissionRoute(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的映射ID"))
		return
	}
	var req modelsystem.UpdatePermissionRouteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误: "+err.Error()))
		return
	}
	if err := ctrl.svc.UpdatePermissionRoute(uint(id), req); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeletePermissionRoute godoc
// @Summary      删除权限路由映射
// @Description  删除后对应端点将被拒绝访问（fail-closed）
// @Tags         系统管理-权限
// @Produce      json
// @Param        id  path  int  true  "映射ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "映射不存在"
// @Router       /system/permissions/routes/{id} [delete]
// @Security     BearerAuth
func (ctrl *PermissionController) DeletePermissionRoute(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorBadRequest("无效的映射ID"))
		return
	}
	if err := ctrl.svc.DeletePermissionRoute(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.ErrorInternal("删除权限路由映射失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
