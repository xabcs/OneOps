package system

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/dto"
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

// GetUsers godoc
// @Summary      获取用户列表
// @Description  分页获取用户列表，支持按用户名、昵称、邮箱、状态搜索
// @Tags         系统管理-用户
// @Produce      json
// @Param        page      query     int     false  "页码"    default(1)
// @Param        pageSize  query     int     false  "每页数量" default(20)
// @Param        username  query     string  false  "用户名"
// @Param        nickname  query     string  false  "昵称"
// @Param        email     query     string  false  "邮箱"
// @Param        status    query     string  false  "状态"
// @Success      200  {object}  utils.Response{data=dto.PageResult}
// @Failure      200  {object}  utils.Response  "请求参数错误 / 获取用户列表失败"
// @Router       /system/users [get]
// @Security     BearerAuth
func (ctrl *UserController) GetUsers(c *gin.Context) {
	var params dto.BasePageQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest(dto.FormatValidationError(err)))
		return
	}

	username := c.Query("username")
	nickname := c.Query("nickname")
	email := c.Query("email")
	status := c.Query("status")

	result, err := ctrl.svc.Search(syssvc.UserSearchQuery{
		Username: username,
		Nickname: nickname,
		Email:    email,
		Status:   status,
		Offset:   params.GetOffset(),
		Limit:    params.GetPageSize(),
	})
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, utils.PageSuccess(dto.NewPageResult(result.Records, result.Total, params)))
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

// CreateUser godoc
// @Summary      创建用户
// @Description  新增用户并分配角色
// @Tags         系统管理-用户
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUserRequest  true  "用户信息"
// @Success      200   {object}  utils.Response{data=object}
// @Failure      200   {object}  utils.Response  "请求参数错误 / 用户名已存在"
// @Router       /system/users [post]
// @Security     BearerAuth
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

// UpdateUser godoc
// @Summary      更新用户
// @Description  按用户 ID 更新指定字段（支持部分更新）
// @Tags         系统管理-用户
// @Accept       json
// @Produce      json
// @Param        id    path      int                true  "用户 ID"
// @Param        body  body      UpdateUserRequest  true  "待更新字段"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的用户ID / 请求参数错误"
// @Router       /system/users/{id} [put]
// @Security     BearerAuth
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

// DeleteUser godoc
// @Summary      删除用户
// @Description  按用户 ID 删除用户（管理员用户受保护不可删除）
// @Tags         系统管理-用户
// @Produce      json
// @Param        id  path      int  true  "用户 ID"
// @Success      200  {object}  utils.Response
// @Failure      200  {object}  utils.Response  "无效的用户ID / 用户不存在 / 不能删除管理员用户"
// @Router       /system/users/{id} [delete]
// @Security     BearerAuth
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

// ResetPassword godoc
// @Summary      重置用户密码
// @Description  管理员按用户 ID 重置其密码
// @Tags         系统管理-用户
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "用户 ID"
// @Param        body  body      ResetPasswordRequest  true  "新密码"
// @Success      200   {object}  utils.Response
// @Failure      200   {object}  utils.Response  "无效的用户ID / 请求参数错误 / 密码重置失败"
// @Router       /system/users/{id}/password [put]
// @Security     BearerAuth
func (ctrl *UserController) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("无效的用户ID"))
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.ErrorBadRequest("请求参数错误"))
		return
	}

	if err := ctrl.svc.ResetPassword(id, req.Password); err != nil {
		if errors.Is(err, syssvc.ErrUserNotFound) {
			c.JSON(http.StatusOK, utils.ErrorInternal("用户不存在"))
			return
		}
		c.JSON(http.StatusOK, utils.ErrorInternal("密码重置失败"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessWithMessage("密码重置成功"))
}

// GetUserOptions godoc
// @Summary      获取用户选项列表
// @Description  不分页返回全部启用用户的精简选项（id/用户名/昵称），用于各业务模块的选人下拉框
// @Tags         系统管理-用户
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]object}  "用户选项列表"
// @Failure      200  {object}  utils.Response  "获取用户选项失败"
// @Router       /system/users/options [get]
// @Security     BearerAuth
func (ctrl *UserController) GetUserOptions(c *gin.Context) {
	users, err := ctrl.svc.GetAllUserOptions()
	if err != nil {
		c.JSON(http.StatusOK, utils.ErrorInternal("获取用户选项失败"))
		return
	}

	// 转换为精简格式
	options := make([]gin.H, len(users))
	for i, u := range users {
		options[i] = gin.H{
			"id":       u.ID,
			"username": u.Username,
			"nickname": u.Nickname,
		}
	}

	c.JSON(http.StatusOK, utils.SuccessWithData(options))
}
