package services

import (
	"encoding/json"
	"oneops/backend/models"
	"oneops/backend/utils"
)

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	// 查找用户
	var user models.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return "", nil, result.Error
	}

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		return "", nil, ErrInvalidPassword
	}

	// 生成 token
	token, err := utils.GenerateToken(user.ID, user.Username, 24)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

// GetUserInfo 获取用户信息（包含权限和菜单）
func (s *AuthService) GetUserInfo(userID uint) (*UserInfo, error) {
	var user models.User
	result := db.First(&user, userID)
	if result.Error != nil {
		return nil, result.Error
	}

	// 获取角色信息
	roles, err := NewRBACService().GetUserRoles(user.ID)
	if err != nil {
		return nil, err
	}

	// 构建角色名称列表
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// 获取菜单树和权限
	rbacService := NewRBACService()
	menuTree, permissions, err := rbacService.BuildMenuTreeAndPermissions(user.ID)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		User:        &user,
		RoleNames:   roleNames,
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
