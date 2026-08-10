package system

import (
	"fmt"
	"net/http"
	"strconv"

	"oneops/backend2/pkg/database"
	"oneops/backend2/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RoleController 角色控制器
type RoleController struct{}

// NewRoleController 创建角色控制器
func NewRoleController() *RoleController {
	return &RoleController{}
}

// GetRoles 获取所有角色（支持搜索和分页）
func (ctrl *RoleController) GetRoles(c *gin.Context) {
	db := database.GetDB()

	// 获取查询参数
	name := c.Query("name")
	code := c.Query("code")
	status := c.Query("status")
	description := c.Query("description")
	currentStr := c.Query("current")
	sizeStr := c.Query("size")

	// 解析分页参数
	current := 1 // 默认第1页
	size := 10   // 默认每页10条

	if currentStr != "" {
		if page, err := strconv.Atoi(currentStr); err == nil && page > 0 {
			current = page
		}
	}
	if sizeStr != "" {
		if limit, err := strconv.Atoi(sizeStr); err == nil && limit > 0 {
			size = limit
		}
	}

	// 构建查询
	query := db.Model(&Role{})

	// 添加搜索条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色总数失败"))
		return
	}

	// 分页查询
	var roles []Role
	offset := (current - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&roles).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取角色列表失败"))
		return
	}

	// 构建分页响应数据
	responseData := map[string]interface{}{
		"records": roles,
		"current": current,
		"size":    size,
		"total":   total,
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(responseData))
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

	db := database.GetDB()
	role := Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
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

	db := database.GetDB()
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
	// 只有明确设置了 status 才更新（避免零值覆盖）
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if err := db.Model(&Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
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

	db := database.GetDB()

	// 检查角色是否存在
	var role Role
	if err := db.First(&role, id).Error; err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusOK, utils.ErrorBadRequest("角色不存在"))
			return
		}
		c.JSON(http.StatusOK, utils.ErrorInternal("查询角色失败"))
		return
	}

	// 检查是否有用户使用该角色（role_ids是JSON字符串）
	var count int64
	// 使用CAST将JSON字符串转换为JSON类型，然后使用JSON_CONTAINS
	db.Model(&User{}).Where("JSON_CONTAINS(CAST(role_ids AS JSON), ?)", fmt.Sprintf("[%d]", id)).Count(&count)

	if count > 0 {
		// 查询使用该角色的用户列表
		var users []User
		db.Select("id, username, nickname").Where("JSON_CONTAINS(CAST(role_ids AS JSON), ?)", fmt.Sprintf("[%d]", id)).Find(&users)

		// 构建用户列表字符串
		userList := ""
		for i, user := range users {
			if i > 0 {
				userList += "、"
			}
			userList += user.Username
			if user.Nickname != "" {
				userList += "(" + user.Nickname + ")"
			}
		}

		c.JSON(http.StatusOK, utils.ErrorBadRequest(
			fmt.Sprintf("该角色已绑定 %d 个用户：%s。请先在用户管理中解除该角色与用户的绑定关系后再删除。", count, userList),
		))
		return
	}

	// 检查是否为默认角色，防止删除系统必需角色
	if role.Code == "admin" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("系统管理员角色不能删除"))
		return
	}

	if err := db.Delete(&Role{}, id).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除角色失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
