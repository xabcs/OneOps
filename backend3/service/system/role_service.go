package system

import (
	"errors"
	"fmt"

	modelsystem "oneops/backend3/model/system"
	reposystem "oneops/backend3/repository/system"
)

// 角色业务错误
var (
	ErrRoleNotFound         = errors.New("角色不存在")
	ErrRoleHasUsers         = errors.New("该角色已绑定用户，请先解除绑定")
	ErrAdminRoleProtected   = errors.New("系统管理员角色不能删除")
	ErrBuiltinRoleProtected = errors.New("系统内置角色不能删除")
)

// builtinRoleCodes 内置角色代码列表
var builtinRoleCodes = map[string]bool{
	"admin":   true,
	"ops":     true,
	"auditor": true,
	"viewer":  true,
	"user":    true,
	"test":    true,
}

// RoleService 角色业务逻辑层
type RoleService struct {
	repo *reposystem.RoleRepository
}

// NewRoleService 创建角色服务
func NewRoleService(repo *reposystem.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

// RoleSearchResult 角色搜索结果
type RoleSearchResult struct {
	Records []modelsystem.Role
	Total   int64
}

// RoleSearchQuery 角色搜索条件
type RoleSearchQuery struct {
	Name        string
	Code        string
	Description string
	Status      string
	Offset      int
	Limit       int
}

// Search 分页搜索角色
func (s *RoleService) Search(q RoleSearchQuery) (*RoleSearchResult, error) {
	roles, total, err := s.repo.FindWithPagination(reposystem.RoleQuery{
		Name:        q.Name,
		Code:        q.Code,
		Description: q.Description,
		Status:      q.Status,
		Offset:      q.Offset,
		Limit:       q.Limit,
	})
	if err != nil {
		return nil, err
	}
	return &RoleSearchResult{Records: roles, Total: total}, nil
}

// GetAllRoleOptions 获取所有启用角色的选项列表（不分页，用于选择器）
func (s *RoleService) GetAllRoleOptions() ([]modelsystem.Role, error) {
	return s.repo.FindAllActive()
}

// GetRoleUsers 获取角色绑定的用户列表（角色列表"绑定用户"列 tag 与"查看用户"弹窗共用）
func (s *RoleService) GetRoleUsers(roleID uint) ([]modelsystem.User, error) {
	return s.repo.FindUsersByRoleID(roleID)
}

// Create 创建角色
func (s *RoleService) Create(role *modelsystem.Role) error {
	return s.repo.Create(role)
}

// Update 根据ID更新指定字段。
// code/status 变更会影响 Casbin 策略：改码需清旧码策略（防孤儿策略放行），
// 禁用/启用按角色最新状态重同步
func (s *RoleService) Update(id uint64, updates map[string]interface{}) error {
	oldRole, err := s.repo.FindByID(id)
	if err != nil {
		return ErrRoleNotFound
	}

	if err := s.repo.UpdateByID(id, updates); err != nil {
		return err
	}

	codeChanged := false
	if v, ok := updates["code"].(string); ok && v != "" && v != oldRole.Code {
		codeChanged = true
	}
	statusChanged := false
	if v, ok := updates["status"].(int); ok && v != oldRole.Status {
		statusChanged = true
	}
	if !codeChanged && !statusChanged {
		return nil
	}

	permSvc, err := GetPermissionService()
	if err != nil {
		// 权限服务不可用不回滚角色更新，策略由下次全量同步收敛
		return nil
	}
	if codeChanged {
		if err := permSvc.RemoveRolePolicies(oldRole.Code); err != nil {
			return fmt.Errorf("角色已更新，但清理旧 Casbin 策略失败: %w", err)
		}
	}
	if err := permSvc.SyncRoleToCasbin(uint(id)); err != nil {
		return fmt.Errorf("角色已更新，但同步 Casbin 策略失败: %w", err)
	}
	return nil
}

// Delete 删除角色（含业务校验）
func (s *RoleService) Delete(id uint) error {
	role, err := s.repo.FindByID(uint64(id))
	if err != nil {
		return ErrRoleNotFound
	}

	// 检查是否有用户使用该角色
	count, err := s.repo.CountUsersByRoleID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		users, _ := s.repo.FindUsersByRoleID(id)
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
		return fmt.Errorf("该角色已绑定 %d 个用户：%s。请先在用户管理中解除该角色与用户的绑定关系后再删除", count, userList)
	}

	// 检查是否为内置角色
	if builtinRoleCodes[role.Code] {
		return ErrBuiltinRoleProtected
	}

	// 删除前清理权限绑定与 Casbin 策略，防孤儿数据/孤儿策略继续放行
	if err := s.repo.DeleteRolePermissions(id); err != nil {
		return err
	}
	if permSvc, err := GetPermissionService(); err == nil {
		_ = permSvc.RemoveRolePolicies(role.Code)
	}

	return s.repo.Delete(uint64(id))
}
