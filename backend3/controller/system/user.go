package system

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/utils"
	syssvc "oneops/backend3/service/system"
)

// UserController 用户控制器
type UserController struct {
	svc *syssvc.UserService
}

// NewUserController 创建用户控制器
func NewUserController(svc *syssvc.UserService) *UserController {
	return &UserController{svc: svc}
}

// GetUsers 获取所有用户（支持搜索和分页）
func (ctrl *UserController) GetUsers(c *gin.Context) {
	username := c.Query("username")
	nickname := c.Query("nickname")
	email := c.Query("email")
	status := c.Query("status")
	currentStr := c.Query("current")
	sizeStr := c.Query("size")

	current := 1
	size := 10
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

	result, err := ctrl.svc.Search(syssvc.UserSearchQuery{
		Username: username,
		Nickname: nickname,
		Email:    email,
		Status:   status,
		Offset:   (current - 1) * size,
		Limit:    size,
	})
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	responseData := map[string]interface{}{
		"records": result.Records,
		"current": current,
		"size":    size,
		"total":   result.Total,
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(responseData))
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	RoleIDs  []uint `json:"roleIds"`
	Status   string `json:"status"`
	HomePath string `json:"homePath"`
}

// CreateUser 创建用户
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	roleIDsJSON, err := json.Marshal(req.RoleIDs)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("JSON 序列化失败"))
		return
	}

	user := modelsystem.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Email:    req.Email,
		RoleIDs:  string(roleIDsJSON),
		Status:   req.Status,
		HomePath: req.HomePath,
	}

	if err := ctrl.svc.Create(&user, req.Password); err != nil {
		if errors.Is(err, syssvc.ErrUsernameExists) {
			c.JSON(http.StatusOK, utils.ErrorBadRequest("用户名已存在"))
			return
		}
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(ctrl.svc.UserToMap(user)))
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	RoleIDs  []uint `json:"roleIds"`
	Status   string `json:"status"`
	HomePath string `json:"homePath"`
	Password string `json:"password"`
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
		updates["home_path"] = req.HomePath
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}

	if err := ctrl.svc.Update(id, updates); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(err.Error()))
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

	if err := ctrl.svc.Delete(id); err != nil {
		switch {
		case errors.Is(err, syssvc.ErrUserNotFound):
			c.JSON(http.StatusOK, utils.ErrorInternal("用户不存在"))
		case errors.Is(err, syssvc.ErrAdminUserProtected):
			c.JSON(http.StatusOK, utils.ErrorBadRequest("不能删除管理员用户"))
		default:
			c.JSON(http.StatusOK, utils.ErrorInternal("删除用户失败"))
		}
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

	if err := ctrl.svc.ResetPassword(id, req.Password); err != nil {
		if errors.Is(err, syssvc.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    500,
				"message": "用户不存在",
				"data":    nil,
			})
			return
		}
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
