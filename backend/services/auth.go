package services

import (
	"encoding/json"
	"time"

	"oneops/backend/logger"
	"oneops/backend/models"
	"oneops/backend/utils"

	"go.uber.org/zap"
)

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	startTime := time.Now()
	logger.Debug("[登录调试-服务层] Login方法开始", zap.String("username", username))

	// 查找用户
	logger.Debug("[登录调试-服务层] 开始查询用户")
	var user models.User
	queryStart := time.Now()
	result := db.Where("username = ?", username).First(&user)
	logger.Debug("[登录调试-服务层] 用户查询完成",
		zap.Duration("耗时", time.Since(queryStart)),
		zap.Error(result.Error))

	if result.Error != nil {
		logger.Debug("[登录调试-服务层] 用户查询失败", zap.Error(result.Error))
		return "", nil, result.Error
	}

	logger.Debug("[登录调试-服务层] 开始验证密码")

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		logger.Debug("[登录调试-服务层] 密码验证失败")
		return "", nil, ErrInvalidPassword
	}

	logger.Debug("[登录调试-服务层] 密码验证成功，开始生成Token")

	// 生成 token
	token, err := utils.GenerateToken(user.ID, user.Username, 24)
	if err != nil {
		logger.Debug("[登录调试-服务层] Token生成失败", zap.Error(err))
		return "", nil, err
	}

	logger.Debug("[登录调试-服务层] Login方法完成",
		zap.Duration("总耗时", time.Since(startTime)),
		zap.Uint("userID", user.ID))

	return token, &user, nil
}

// GetUserInfo 获取用户信息（包含权限和菜单）
func (s *AuthService) GetUserInfo(userID uint) (*UserInfo, error) {
	startTime := time.Now()
	logger.Debug("[登录调试-服务层] GetUserInfo方法开始", zap.Uint("userID", userID))

	var user models.User
	logger.Debug("[登录调试-服务层] 开始查询用户详情")
	queryStart := time.Now()
	result := db.First(&user, userID)
	logger.Debug("[登录调试-服务层] 用户详情查询完成",
		zap.Duration("耗时", time.Since(queryStart)),
		zap.Error(result.Error))

	if result.Error != nil {
		return nil, result.Error
	}

	logger.Debug("[登录调试-服务层] 开始获取角色信息")

	// 获取角色信息
	roleStart := time.Now()
	roles, err := NewRBACService().GetUserRoles(user.ID)
	logger.Debug("[登录调试-服务层] 角色信息获取完成",
		zap.Duration("耗时", time.Since(roleStart)),
		zap.Error(err),
		zap.Int("角色数量", len(roles)))

	if err != nil {
		return nil, err
	}

	// 构建角色代码列表（使用 Code 而不是 Name，方便前端进行权限匹配）
	roleCodes := make([]string, len(roles))
	for i, role := range roles {
		roleCodes[i] = role.Code
	}

	logger.Debug("[登录调试-服务层] 开始构建菜单树和权限")

	// 获取菜单树和权限
	rbacService := NewRBACService()
	menuStart := time.Now()
	menuTree, permissions, _, err := rbacService.BuildMenuTreeAndPermissions(user.ID)
	logger.Debug("[登录调试-服务层] 菜单树和权限构建完成",
		zap.Duration("耗时", time.Since(menuStart)),
		zap.Error(err),
		zap.Int("菜单数量", len(menuTree)),
		zap.Int("权限数量", len(permissions)))

	if err != nil {
		return nil, err
	}

	logger.Debug("[登录调试-服务层] GetUserInfo方法完成",
		zap.Duration("总耗时", time.Since(startTime)))

	return &UserInfo{
		User:        &user,
		RoleNames:   roleCodes,
		MenuTree:    menuTree,
		Permissions: permissions,
	}, nil
}

// UserInfo 用户信息（包含权限）
type UserInfo struct {
	User        *models.User `json:"-"`
	RoleNames   []string     `json:"roleNames"`
	MenuTree    []*models.Menu `json:"menuTree"`
	Permissions []string     `json:"permissions"`
}

// ToMap 转换为 map 格式（用于 JSON 响应）
func (ui *UserInfo) ToMap() map[string]interface{} {
	// 解析 roleIds JSON 字符串为数组
	var roleIDs []uint
	json.Unmarshal([]byte(ui.User.RoleIDs), &roleIDs)

	return map[string]interface{}{
		"id":         ui.User.ID,
		"username":   ui.User.Username,
		"nickname":   ui.User.Nickname,
		"avatar":     ui.User.Avatar,
		"email":      ui.User.Email,
		"roleIds":    roleIDs,
		"status":     ui.User.Status,
		"homePath":   ui.User.HomePath,
		"createdAt":  ui.User.CreatedAt.Format("2006-01-02"),
		"updatedAt":  ui.User.UpdatedAt.Format("2006-01-02"),
		"roleNames":  ui.RoleNames,
		"menuTree":   ui.MenuTree,
		"permissions": ui.Permissions,
	}
}

// 错误定义
var (
	ErrInvalidPassword = &AuthError{Message: "密码错误"}
)

// AuthError 认证错误
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
