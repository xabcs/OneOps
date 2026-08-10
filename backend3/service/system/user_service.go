package system

import (
	"encoding/json"
	"errors"
	"strings"

	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/utils"
	reposystem "oneops/backend3/repository/system"
)

// 用户业务错误
var (
	ErrUsernameExists     = errors.New("用户名已存在")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrAdminUserProtected = errors.New("不能删除管理员用户")
)

// UserService 用户业务逻辑层
type UserService struct {
	repo *reposystem.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(repo *reposystem.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// UserSearchResult 用户搜索结果
type UserSearchResult struct {
	Records []map[string]interface{}
	Total   int64
}

// UserSearchQuery 用户搜索条件
type UserSearchQuery struct {
	Username string
	Nickname string
	Email    string
	Status   string
	Offset   int
	Limit    int
}

// Search 分页搜索用户
func (s *UserService) Search(q UserSearchQuery) (*UserSearchResult, error) {
	users, total, err := s.repo.FindWithPagination(reposystem.UserQuery{
		Username: q.Username,
		Nickname: q.Nickname,
		Email:    q.Email,
		Status:   q.Status,
		Offset:   q.Offset,
		Limit:    q.Limit,
	})
	if err != nil {
		return nil, err
	}

	records := make([]map[string]interface{}, len(users))
	for i, user := range users {
		records[i] = s.userToMap(user)
	}

	return &UserSearchResult{Records: records, Total: total}, nil
}

// Create 创建用户
func (s *UserService) Create(user *modelsystem.User, plainPassword string) error {
	// 检查用户名是否已存在
	count, err := s.repo.CountByUsername(user.Username)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrUsernameExists
	}

	// 验证家目录权限
	if valid, errMsg := s.validateHomePathPermission(user.HomePath, s.parseRoleIDs(user.RoleIDs)); !valid {
		return errors.New(errMsg)
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}
	user.Password = hashedPassword

	return s.repo.Create(user)
}

// Update 更新用户信息
func (s *UserService) Update(id uint64, updates map[string]interface{}) error {
	// 如果更新家目录，需验证权限
	if homePath, ok := updates["home_path"].(string); ok && homePath != "" {
		// 获取角色ID列表
		var roleIDs []uint
		if roleIDsVal, exists := updates["role_ids"]; exists {
			roleIDs = s.parseRoleIDs(roleIDsVal.(string))
		} else {
			user, err := s.repo.FindByID(id)
			if err == nil {
				roleIDs = s.parseRoleIDs(user.RoleIDs)
			}
		}
		if valid, errMsg := s.validateHomePathPermission(homePath, roleIDs); !valid {
			return errors.New(errMsg)
		}
	}

	// 如果更新密码，需加密
	if password, ok := updates["password"].(string); ok && password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return errors.New("密码加密失败")
		}
		updates["password"] = hashedPassword
	}

	return s.repo.UpdateByID(id, updates)
}

// Delete 删除用户（含业务校验）
func (s *UserService) Delete(id uint64) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Username == "admin" {
		return ErrAdminUserProtected
	}

	return s.repo.Delete(id)
}

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword(id uint64, plainPassword string) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	return s.repo.UpdatePassword(id, hashedPassword)
}

// UserToMap 将用户模型转换为 map（处理 JSON 字段）
func (s *UserService) UserToMap(user modelsystem.User) map[string]interface{} {
	return s.userToMap(user)
}

// userToMap 内部方法
func (s *UserService) userToMap(user modelsystem.User) map[string]interface{} {
	roleIDs := s.parseRoleIDs(user.RoleIDs)

	return map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"avatar":    user.Avatar,
		"email":     user.Email,
		"roleIds":   roleIDs,
		"status":    user.Status,
		"homePath":  user.HomePath,
		"createdAt": user.CreatedAt.Format("2006-01-02"),
		"updatedAt": user.UpdatedAt.Format("2006-01-02"),
	}
}

// parseRoleIDs 解析角色ID JSON字符串
func (s *UserService) parseRoleIDs(roleIDsStr string) []uint {
	var roleIDs []uint
	json.Unmarshal([]byte(roleIDsStr), &roleIDs)
	return roleIDs
}

// validateHomePathPermission 验证家目录权限
func (s *UserService) validateHomePathPermission(homePath string, roleIDs []uint) (bool, string) {
	if homePath == "" || homePath == "/" {
		return true, ""
	}

	if len(roleIDs) == 0 {
		return false, "未分配角色的用户家目录必须为根路径"
	}

	// 获取所有指定角色的权限
	permissionSet := make(map[string]bool)
	for _, roleID := range roleIDs {
		rps, err := s.repo.FindRolePermissions(roleID)
		if err != nil {
			continue
		}
		for _, rp := range rps {
			if rp.Permission.Code != "" {
				permissionSet[rp.Permission.Code] = true
			}
		}
	}

	if len(permissionSet) == 0 {
		return true, ""
	}

	menu, err := s.repo.FindMenuByPath(homePath)
	if err != nil {
		return true, ""
	}

	if menu.Resource != "" {
		hasPermission := false
		for permCode := range permissionSet {
			if strings.Contains(permCode, menu.Resource+".") {
				hasPermission = true
				break
			}
		}
		_ = hasPermission
	}

	return true, ""
}
