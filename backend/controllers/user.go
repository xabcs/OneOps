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

// UserController 用户控制器
type UserController struct{}

// NewUserController 创建用户控制器
func NewUserController() *UserController {
	return &UserController{}
}

// validateHomePathPermission 验证家目录权限
// 返回: (是否有权限, 错误信息)
func (ctrl *UserController) validateHomePathPermission(homePath string, roleIDs []uint) (bool, string) {
	// 根路径不需要验证
	if homePath == "" || homePath == "/" {
		return true, ""
	}

	// 没有分配角色时，家目录必须是根路径
	if len(roleIDs) == 0 {
		return false, "未分配角色的用户家目录必须为根路径"
	}

	db := services.GetDB()

	// 获取所有指定角色的菜单ID集合
	menuIDSet := make(map[uint]bool)
	for _, roleID := range roleIDs {
		var role models.Role
		if err := db.First(&role, roleID).Error; err != nil {
			return false, "角色不存在"
		}

		// 解析角色的菜单ID列表
		var roleMenuIDs []uint
		if err := json.Unmarshal([]byte(role.MenuIDs), &roleMenuIDs); err != nil {
			continue
		}

		// 添加到集合中
		for _, menuID := range roleMenuIDs {
			menuIDSet[menuID] = true
		}
	}

	// 如果没有任何菜单权限，拒绝任何非根路径
	if len(menuIDSet) == 0 {
		return false, "角色没有分配任何菜单权限，无法设置家目录"
	}

	// 查找家目录对应的菜单
	var menu models.Menu
	if err := db.Where("path = ?", homePath).First(&menu).Error; err != nil {
		return false, "家目录路径对应的菜单不存在"
	}

	// 检查该菜单是否在角色的权限范围内
	if !menuIDSet[menu.ID] {
		return false, "家目录不在用户角色权限范围内"
	}

	return true, ""
}

// GetUsers 获取所有用户
func (ctrl *UserController) GetUsers(c *gin.Context) {
	db := services.GetDB()

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	// 处理用户数据，将 JSON 字符串转换为数组
	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = ctrl.userToMap(user)
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(result))
}

// userToMap 将用户模型转换为 map，处理 JSON 字段
func (ctrl *UserController) userToMap(user models.User) map[string]interface{} {
	var roleIDs []uint
	json.Unmarshal([]byte(user.RoleIDs), &roleIDs)

	return map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"avatar":     user.Avatar,
		"email":      user.Email,
		"roleIds":    roleIDs,
		"status":     user.Status,
		"homePath":   user.HomePath,
		"createdAt":  user.CreatedAt.Format("2006-01-02"),
		"updatedAt":  user.UpdatedAt.Format("2006-01-02"),
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	RoleIDs   []uint `json:"roleIds"`
	Status    string `json:"status"`
	HomePath  string `json:"homePath"`
}

// CreateUser 创建用户
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	// 检查用户名是否已存在
	db := services.GetDB()
	var count int64
	db.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("用户名已存在"))
		return
	}

	// 验证家目录权限
	if valid, errMsg := ctrl.validateHomePathPermission(req.HomePath, req.RoleIDs); !valid {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(errMsg))
		return
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("密码加密失败"))
		return
	}

	roleIDsJSON, err := json.Marshal(req.RoleIDs)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("JSON 序列化失败"))
		return
	}

	user := models.User{
		Username:  req.Username,
		Password:  hashedPassword,
		Nickname:  req.Nickname,
		Avatar:    req.Avatar,
		Email:     req.Email,
		RoleIDs:   string(roleIDsJSON),
		Status:    req.Status,
		HomePath:  req.HomePath,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("创建用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(ctrl.userToMap(user)))
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	RoleIDs   []uint `json:"roleIds"`
	Status    string `json:"status"`
	HomePath  string `json:"homePath"`
	Password  string `json:"password"`
}

// UpdateUser 更新用户
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户ID"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	db := services.GetDB()
	updates := map[string]interface{}{}

	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.RoleIDs != nil {
		roleIDsJSON, err := json.Marshal(req.RoleIDs)
		if err != nil {
			c.JSON(http.StatusOK, utils.ErrorInternal("JSON 序列化失败"))
			return
		}
		updates["role_ids"] = string(roleIDsJSON)
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.HomePath != "" {
		// 获取用户的角色ID列表（使用请求中的或现有的）
		roleIDs := req.RoleIDs
		if roleIDs == nil {
			// 如果请求中没有更新角色，获取现有的角色
			var user models.User
			if err := db.First(&user, id).Error; err == nil {
				json.Unmarshal([]byte(user.RoleIDs), &roleIDs)
			}
		}

		// 验证家目录权限
		if valid, errMsg := ctrl.validateHomePathPermission(req.HomePath, roleIDs); !valid {
			c.JSON(http.StatusOK, utils.ErrorBadRequest(errMsg))
			return
		}

		updates["home_path"] = req.HomePath
	}
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusOK, utils.ErrorInternal("密码加密失败"))
			return
		}
		updates["password"] = hashedPassword
	}

	if err := db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("更新用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("更新成功"))
}

// DeleteUser 删除用户
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户ID"))
		return
	}

	// 不能删除 admin 用户
	if id == 1 {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("不能删除管理员用户"))
		return
	}

	db := services.GetDB()
	if err := db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}
