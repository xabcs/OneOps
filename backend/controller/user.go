package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

	// 获取所有指定角色的权限
	permissionSet := make(map[string]bool)
	for _, roleID := range roleIDs {
		var rolePermissions []models.RolePermission
		if err := db.Where("role_id = ?", roleID).Preload("Permission").Find(&rolePermissions).Error; err != nil {
			continue
		}

		for _, rp := range rolePermissions {
			if rp.Permission.Code != "" {
				permissionSet[rp.Permission.Code] = true
			}
		}
	}

	// 如果没有任何权限，允许设置任何家目录（向后兼容）
	if len(permissionSet) == 0 {
		return true, ""
	}

	// 查找家目录对应的菜单
	var menu models.Menu
	if err := db.Where("path = ?", homePath).First(&menu).Error; err != nil {
		// 菜单不存在时允许设置（向后兼容）
		return true, ""
	}

	// 检查是否有对应的权限（使用权限码推导）
	if menu.Resource != "" {
		// 检查是否有该资源的任何权限
		hasPermission := false
		for permCode := range permissionSet {
			// 检查权限码是否匹配该资源
			// 权限码格式：module.resource.action
			if strings.Contains(permCode, menu.Resource+".") {
				hasPermission = true
				break
			}
		}
		// 如果没有权限，仍然允许设置（向后兼容）
		_ = hasPermission
	}

	return true, ""
}

// GetUsers 获取所有用户（支持搜索和分页）
func (ctrl *UserController) GetUsers(c *gin.Context) {
	db := services.GetDB()

	// 获取查询参数
	username := c.Query("username")
	nickname := c.Query("nickname")
	email := c.Query("email")
	status := c.Query("status")
	currentStr := c.Query("current")
	sizeStr := c.Query("size")

	// 解析分页参数
	current := 1 // 默认第1页
	size := 10  // 默认每页10条

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
	query := db.Model(&models.User{})

	// 添加搜索条件
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户总数失败"))
		return
	}

	// 分页查询
	var users []models.User
	offset := (current - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&users).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	// 处理用户数据，将 JSON 字符串转换为数组
	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = ctrl.userToMap(user)
	}

	// 构建分页响应数据
	responseData := map[string]interface{}{
		"records": result,
		"current": current,
		"size":   size,
		"total":  total,
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(responseData))
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

	db := services.GetDB()

	// 先查询用户信息，检查是否为 admin
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("用户不存在"))
		return
	}

	// 不能删除 admin 用户（通过用户名判断，更严谨）
	if user.Username == "admin" {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("不能删除管理员用户"))
		return
	}

	if err := db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("删除用户失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("删除成功"))
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// ResetPassword 重置用户密码
func (ctrl *UserController) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    500,
			"message": "无效的用户ID",
			"data":    nil,
		})
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    500,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	db := services.GetDB()

	// 检查用户是否存在
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    500,
			"message": "用户不存在",
			"data":    nil,
		})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
			"data":    nil,
		})
		return
	}

	// 更新密码
	if err := db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码重置失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码重置成功",
		"data":    nil,
	})
}
