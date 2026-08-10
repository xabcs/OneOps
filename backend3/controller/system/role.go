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

// GetRoles 获取所有角色（支持搜索和分页）
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

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// CreateRole 创建角色
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

// UpdateRole 更新角色
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

// DeleteRole 删除角色
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
