package system

import (
	"strings"
	"time"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"

	"go.uber.org/zap"
)

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, *modelsystem.User, error) {
	startTime := time.Now()
	logger.Debug("[登录调试-服务层] Login方法开始", zap.String("username", username))

	// 查找用户
	logger.Debug("[登录调试-服务层] 开始查询用户")
	var user modelsystem.User
	queryStart := time.Now()
	result := database.GetDB().Preload("Roles").Where("username = ?", username).First(&user)
	logger.Debug("[登录调试-服务层] 用户查询完成",
		zap.Duration("耗时", time.Since(queryStart)),
		zap.Error(result.Error))

	if result.Error != nil {
		logger.Debug("[登录调试-服务层] 用户查询失败", zap.Error(result.Error))
		return "", nil, result.Error
	}

	// 禁用用户禁止登录（H4）
	if user.Status != "active" {
		logger.Warn("禁用用户尝试登录", zap.String("username", username), zap.Uint("user_id", user.ID))
		return "", nil, ErrUserDisabled
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
		zap.Uint("user_id", user.ID))

	return token, &user, nil
}

// GetUserInfo 获取用户信息（包含权限和菜单）
func (s *AuthService) GetUserInfo(userID uint) (*UserInfo, error) {
	startTime := time.Now()
	logger.Debug("[登录调试-服务层] GetUserInfo方法开始", zap.Uint("user_id", userID))

	var user modelsystem.User
	logger.Debug("[登录调试-服务层] 开始查询用户详情")
	queryStart := time.Now()
	result := database.GetDB().Preload("Roles").First(&user, userID)
	logger.Debug("[登录调试-服务层] 用户详情查询完成",
		zap.Duration("耗时", time.Since(queryStart)),
		zap.Error(result.Error))

	if result.Error != nil {
		return nil, result.Error
	}

	// 禁用用户不返回信息（H4：与中间件拦截形成双保险）
	if user.Status != "active" {
		return nil, ErrUserDisabled
	}

	logger.Debug("[登录调试-服务层] 开始获取角色信息")

	// 获取角色信息
	roleStart := time.Now()
	permSvc, err := GetPermissionService()
	if err != nil {
		return nil, err
	}
	roles, err := permSvc.GetUserRoles(user.ID)
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
	menuStart := time.Now()
	menuTree, permissions, _, err := permSvc.BuildMenuTreeAndPermissions(user.ID)
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

	// 查询权限详细信息（用于前端显示权限名称）
	// 批次三：通配符与 BuildMenuTreeAndPermissions 对齐（*:*:*，冒号分隔）。
	// 此前误用 *.*.* 永不命中，导致 admin 每次登录多执行一次无效的 IN 查询
	var permissionInfos []PermissionInfo
	if len(permissions) > 0 && !strings.Contains(permissions[0], "*:*:*") {
		// 非管理员，查询权限详情
		var perms []modelsystem.Permission
		err := database.GetDB().Where("code IN ?", permissions).Find(&perms).Error
		if err == nil {
			for _, perm := range perms {
				permissionInfos = append(permissionInfos, PermissionInfo{
					Code: perm.Code,
					Name: perm.Name,
				})
			}
		}
	}

	return &UserInfo{
		User:           &user,
		RoleNames:      roleCodes,
		MenuTree:       menuTree,
		Permissions:    permissions,
		PermissionInfo: permissionInfos,
	}, nil
}

// PermissionInfo 权限详细信息（用于前端显示权限名称）
type PermissionInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// UserInfo 用户信息（包含权限）
type UserInfo struct {
	User           *modelsystem.User   `json:"-"`
	RoleNames      []string            `json:"roleNames"`
	MenuTree       []*modelsystem.Menu `json:"menuTree"`
	Permissions    []string            `json:"permissions"`
	PermissionInfo []PermissionInfo    `json:"permissionInfo"`
}

// ToMap 转换为 map 格式（用于 JSON 响应）
func (ui *UserInfo) ToMap() map[string]interface{} {
	roleIDs := make([]uint, 0, len(ui.User.Roles))
	for _, r := range ui.User.Roles {
		roleIDs = append(roleIDs, r.ID)
	}

	return map[string]interface{}{
		"id":             ui.User.ID,
		"username":       ui.User.Username,
		"nickname":       ui.User.Nickname,
		"avatar":         ui.User.Avatar,
		"email":          ui.User.Email,
		"roleIds":        roleIDs,
		"status":         ui.User.Status,
		"homePath":       ui.User.HomePath,
		"createdAt":      ui.User.CreatedAt.Format("2006-01-02"),
		"updatedAt":      ui.User.UpdatedAt.Format("2006-01-02"),
		"roleNames":      ui.RoleNames,
		"menuTree":       ui.MenuTree,
		"permissions":    ui.Permissions,
		"permissionInfo": ui.PermissionInfo,
	}
}

// 错误定义
var (
	ErrInvalidPassword = &AuthError{Message: "密码错误"}
	ErrUserDisabled    = &AuthError{Message: "账号已被禁用，请联系管理员"}
)

// AuthError 认证错误
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
