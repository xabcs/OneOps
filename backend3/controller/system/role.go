package system

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/dto"
	"oneops/backend3/pkg/utils"
	syssvc "oneops/backend3/service/system"
)

// RoleController 角色控制器
type RoleController struct {
	svc *syssvc.RoleService
}

// NewRoleController 创建角色控制器
func NewRoleController(svc *syssvc.RoleService) *RoleController {
	return &RoleController{svc: svc}
}

// GetRoles godoc
// @Summary      获取角色列表
// @Description  分页获取角色列表，支持按名称、编码、状态、描述搜索
// @Tags         系统管理-角色
// @Produce      json
// @Param        page        query     int     false  "页码"    default(1)
// @Param        pageSize    query     int     false  "每页数量" default(20)
// @Param        name        query     string  false  "角色名称"
// @Param        code        query     string  false  "角色编码"
// @Param        status      query     string  false  "状态"
// @Param        description query     string  false  "描述"
// @Success      200  {object}  utils.Response{data=dto.PageResult{list=[]modelsystem.Role}}
// @Failure      200  {object}  utils.Response  "获取角色列表失败"
// @Router       /system/roles [get]
// @Security     BearerAuth
func (ctrl *RoleController) GetRoles(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	name := c.Query("name")
	code := c.Query("code")
	status := c.Query("status")
	description := c.Query("description")

	result, err := ctrl.svc.Search(syssvc.RoleSearchQuery{
		Name:        name,
		Code:        code,
		Description: description,
		Status:      status,
		Offset:      params.GetOffset(),
		Limit:       params.GetPageSize(),
	})
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(result.Records, result.Total, params)))
}

// GetRoleOptions godoc
// @Summary      获取角色选项列表
// @Description  获取所有启用角色的精简选项（不分页，用于下拉选择器）
// @Tags         系统管理-角色
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]gin.H}
// @Failure      200  {object}  utils.Response  "获取角色选项失败"
// @Router       /system/roles/options [get]
// @Security     BearerAuth
func (ctrl *RoleController) GetRoleOptions(c *gin.Context) {
	roles, err := ctrl.svc.GetAllRoleOptions()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色选项失败"))
		return
	}

	// 转换为精简格式
	options := make([]gin.H, len(roles))
	for i, r := range roles {
		options[i] = gin.H{
			"id":   r.ID,
			"name": r.Name,
			"code": r.Code,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(options))
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// CreateRole godoc
// @Summary      创建角色
// @Description  新增一个角色
// @Tags         系统管理-角色
// @Accept       json
// @Produce      json
// @Param        body  body      CreateRoleRequest  true  "角色信息"
// @Success      200   {object}  utils.Response{data=modelsystem.Role}
// @Failure      200   {object}  utils.Response  "请求参数错误 / 创建角色失败"
// @Router       /system/roles [post]
// @Security     BearerAuth
func (ctrl *RoleController) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	role := modelsystem.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := ctrl.svc.Create(&role); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建角色失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(role))
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// UpdateRole godoc
// @Summary      更新角色
// @Description  按角色 ID 更新指定字段（支持部分更新）
// @Tags         系统管理-角色
// @Accept       json
// @Produce      json
// @Param        id    path      int                true  "角色 ID"
// @Param        body  body      UpdateRoleRequest  true  "待更新字段"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的角色ID / 请求参数错误 / 更新角色失败"
// @Router       /system/roles/{id} [put]
// @Security     BearerAuth
func (ctrl *RoleController) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if err := ctrl.svc.Update(id, updates); err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新角色失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteRole godoc
// @Summary      删除角色
// @Description  按角色 ID 删除角色（受保护的角色禁止删除）
// @Tags         系统管理-角色
// @Produce      json
// @Param        id  path      int  true  "角色 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "无效的角色ID / 角色不存在 / 删除失败"
// @Router       /system/roles/{id} [delete]
// @Security     BearerAuth
func (ctrl *RoleController) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的角色ID"))
		return
	}

	if err := ctrl.svc.Delete(uint(id)); err != nil {
		switch {
		case errors.Is(err, syssvc.ErrRoleNotFound):
			c.JSON(http.StatusOK, utils.ErrorBadRequest("角色不存在"))
		case errors.Is(err, syssvc.ErrAdminRoleProtected):
			c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		default:
			c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
