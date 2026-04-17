package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"oneops/backend/models"
	"oneops/backend/services"
	"oneops/backend/utils"
)

// RoleController 角色控制器
type RoleController struct{}

// NewRoleController 创建角色控制器
func NewRoleController() *RoleController {
	return &RoleController{}
}

// GetRoles 获取所有角色
func (ctrl *RoleController) GetRoles(c *gin.Context) {
	db := services.GetDB()

	var roles []models.Role
	if err := db.Find(&roles).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(roles))
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	MenuIDs     []uint `json:"menuIds"`
	Status      int    `json:"status"`
}

// CreateRole 创建角色
func (ctrl *RoleController) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	menuIDsJSON, err := json.Marshal(req.MenuIDs)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("JSON 序列化失败"))
		return
	}

	db := services.GetDB()
	role := models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		MenuIDs:     string(menuIDsJSON),
		Status:      req.Status,
	}

	if err := db.Create(&role).Error; err != nil {
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
	MenuIDs     []uint `json:"menuIds"`
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

	db := services.GetDB()
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
	if req.MenuIDs != nil {
		menuIDsJSON, err := json.Marshal(req.MenuIDs)
		if err != nil {
			c.JSON(http.StatusOK, utils.ErrorInternal("JSON 序列化失败"))
			return
		}
		updates["menu_ids"] = string(menuIDsJSON)
	}
	// 只有明确设置了 status 才更新（避免零值覆盖）
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if err := db.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
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

	db := services.GetDB()

	// 检查是否有用户使用该角色
	var count int64
	db.Model(&models.User{}).Where("role_ids LIKE ?", "%\""+idStr+"\"%").Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("该角色正在被使用，无法删除"))
		return
	}

	if err := db.Delete(&models.Role{}, id).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除角色失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
