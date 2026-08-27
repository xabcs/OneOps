package system

import (
	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// RoleRepository 角色数据访问层
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// RoleQuery 角色查询条件
type RoleQuery struct {
	Name        string
	Code        string
	Description string
	Status      string
	Offset      int
	Limit       int
}

// FindWithPagination 分页查询角色
func (r *RoleRepository) FindWithPagination(q RoleQuery) ([]modelsystem.Role, int64, error) {
	query := r.db.Model(&modelsystem.Role{})

	if q.Name != "" {
		query = query.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Code != "" {
		query = query.Where("code LIKE ?", "%"+q.Code+"%")
	}
	if q.Description != "" {
		query = query.Where("description LIKE ?", "%"+q.Description+"%")
	}
	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var roles []modelsystem.Role
	if err := query.Preload("Users").Offset(q.Offset).Limit(q.Limit).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// FindAllActive 查询所有启用状态的角色（不分页，用于选择器选项）
func (r *RoleRepository) FindAllActive() ([]modelsystem.Role, error) {
	var roles []modelsystem.Role
	if err := r.db.Where("status = ?", 1).Order("id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Create 创建角色
func (r *RoleRepository) Create(role *modelsystem.Role) error {
	return r.db.Create(role).Error
}

// UpdateByID 根据ID更新指定字段
func (r *RoleRepository) UpdateByID(id uint64, updates map[string]interface{}) error {
	return r.db.Model(&modelsystem.Role{}).Where("id = ?", id).Updates(updates).Error
}

// FindByID 根据ID查询角色
func (r *RoleRepository) FindByID(id uint64) (*modelsystem.Role, error) {
	var role modelsystem.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// Delete 根据ID删除角色
func (r *RoleRepository) Delete(id uint64) error {
	return r.db.Delete(&modelsystem.Role{}, id).Error
}

// DeleteRolePermissions 清理角色的权限绑定（删除角色前调用，防孤儿绑定）
func (r *RoleRepository) DeleteRolePermissions(roleID uint) error {
	return r.db.Where("role_id = ?", roleID).Delete(&modelsystem.RolePermission{}).Error
}

// CountUsersByRoleID 统计使用指定角色的用户数量
func (r *RoleRepository) CountUsersByRoleID(roleID uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelsystem.User{}).
		Joins("JOIN sys_user_roles ur ON ur.user_id = sys_users.id").
		Where("ur.role_id = ?", roleID).
		Count(&count).Error
	return count, err
}

// FindUsersByRoleID 查询使用指定角色的用户列表（含邮箱/状态，供角色管理页查看绑定用户）
func (r *RoleRepository) FindUsersByRoleID(roleID uint) ([]modelsystem.User, error) {
	var users []modelsystem.User
	err := r.db.Select("sys_users.id, sys_users.username, sys_users.nickname, sys_users.email, sys_users.status").
		Joins("JOIN sys_user_roles ur ON ur.user_id = sys_users.id").
		Where("ur.role_id = ?", roleID).
		Order("sys_users.id ASC").
		Find(&users).Error
	return users, err
}
