package authorization

import (
	"fmt"

	modelauth "oneops/backend3/model/authorization"
)

// ========== 授权中心用户管理 ==========

// CreateAuthUser 创建授权中心用户，返回生成的密码
func (s *ApplicationPermissionService) CreateAuthUser(user *modelauth.AuthUser) (string, error) {
	password, err := GenerateRandomPassword(DefaultPasswordConfig)
	if err != nil {
		return "", fmt.Errorf("生成密码失败: %w", err)
	}

	user.Password = password

	if err := s.repo.CreateAuthUser(user); err != nil {
		return "", err
	}

	return password, nil
}

// GetAuthUsers 获取授权中心用户列表
func (s *ApplicationPermissionService) GetAuthUsers(page, pageSize int, username string) ([]*modelauth.AuthUser, int64, error) {
	return s.repo.FindAuthUsers(page, pageSize, username)
}

// GetAuthUserByID 根据ID获取授权中心用户
func (s *ApplicationPermissionService) GetAuthUserByID(id uint) (*modelauth.AuthUser, error) {
	return s.repo.FindAuthUserByID(id)
}

// UpdateAuthUser 更新授权中心用户
func (s *ApplicationPermissionService) UpdateAuthUser(user *modelauth.AuthUser) error {
	return s.repo.SaveAuthUser(user)
}

// DeleteAuthUser 删除授权中心用户
func (s *ApplicationPermissionService) DeleteAuthUser(id uint) error {
	return s.repo.DeleteAuthUser(id)
}

// GetAllAuthUsers 获取所有授权中心用户
func (s *ApplicationPermissionService) GetAllAuthUsers() ([]*modelauth.AuthUser, error) {
	return s.repo.FindAllActiveAuthUsers()
}

// ========== 授权中心角色管理 ==========

// CreateAuthGroup 创建授权中心用户组
func (s *ApplicationPermissionService) CreateAuthGroup(group *modelauth.AuthGroup) error {
	return s.repo.CreateAuthGroup(group)
}

// GetAuthGroups 获取授权中心用户组列表
func (s *ApplicationPermissionService) GetAuthGroups(page, pageSize int, name string) ([]*modelauth.AuthGroup, int64, error) {
	return s.repo.FindAuthGroups(page, pageSize, name)
}

// GetAuthGroupByID 根据ID获取授权中心用户组
func (s *ApplicationPermissionService) GetAuthGroupByID(id uint) (*modelauth.AuthGroup, error) {
	return s.repo.FindAuthGroupByID(id)
}

// UpdateAuthGroup 更新授权中心用户组
func (s *ApplicationPermissionService) UpdateAuthGroup(group *modelauth.AuthGroup) error {
	return s.repo.SaveAuthGroup(group)
}

// DeleteAuthGroup 删除授权中心用户组
func (s *ApplicationPermissionService) DeleteAuthGroup(id uint) error {
	return s.repo.DeleteAuthGroup(id)
}

// GetAllAuthGroups 获取所有授权中心用户组
func (s *ApplicationPermissionService) GetAllAuthGroups() ([]*modelauth.AuthGroup, error) {
	return s.repo.FindAllActiveAuthGroups()
}
