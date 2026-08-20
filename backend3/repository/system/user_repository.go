package system

import (
	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// UserQuery 用户查询条件
type UserQuery struct {
	Username string
	Nickname string
	Email    string
	Status   string
	Offset   int
	Limit    int
}

// FindWithPagination 分页查询用户
func (r *UserRepository) FindWithPagination(q UserQuery) ([]modelsystem.User, int64, error) {
	query := r.db.Model(&modelsystem.User{})
	if q.Username != "" {
		query = query.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+q.Nickname+"%")
	}
	if q.Email != "" {
		query = query.Where("email LIKE ?", "%"+q.Email+"%")
	}
	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []modelsystem.User
	if err := query.Preload("Roles").Offset(q.Offset).Limit(q.Limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *modelsystem.User) error {
	return r.db.Create(user).Error
}

// UpdateByID 根据ID更新指定字段
func (r *UserRepository) UpdateByID(id uint64, updates map[string]interface{}) error {
	return r.db.Model(&modelsystem.User{}).Where("id = ?", id).Updates(updates).Error
}

// FindByID 根据ID查询用户（含角色关联）
func (r *UserRepository) FindByID(id uint64) (*modelsystem.User, error) {
	var user modelsystem.User
	if err := r.db.Preload("Roles").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ReplaceRoles 全量替换用户的角色绑定（many2many）；仅保留实际存在的角色 ID，过滤悬空 ID
func (r *UserRepository) ReplaceRoles(id uint64, roleIDs []uint) error {
	var user modelsystem.User
	if err := r.db.First(&user, id).Error; err != nil {
		return err
	}

	var roles []modelsystem.Role
	if len(roleIDs) > 0 {
		if err := r.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return err
		}
	}

	return r.db.Model(&user).Association("Roles").Replace(&roles)
}

// Delete 根据ID删除用户（级联清理 sys_user_roles 关联行）
func (r *UserRepository) Delete(id uint64) error {
	return r.db.Select("Roles").Delete(&modelsystem.User{}, id).Error
}

// CountByUsername 根据用户名统计数量（检查重复）
func (r *UserRepository) CountByUsername(username string) (int64, error) {
	var count int64
	err := r.db.Model(&modelsystem.User{}).Where("username = ?", username).Count(&count).Error
	return count, err
}

// FindRolePermissions 查询角色关联的权限（含预加载权限详情）
func (r *UserRepository) FindRolePermissions(roleID uint) ([]modelsystem.RolePermission, error) {
	var rps []modelsystem.RolePermission
	err := r.db.Where("role_id = ?", roleID).Preload("Permission").Find(&rps).Error
	return rps, err
}

// FindMenuByPath 根据路径查询菜单
func (r *UserRepository) FindMenuByPath(path string) (*modelsystem.Menu, error) {
	var menu modelsystem.Menu
	if err := r.db.Where("path = ?", path).First(&menu).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// UpdatePassword 更新用户密码
func (r *UserRepository) UpdatePassword(id uint64, password string) error {
	return r.db.Model(&modelsystem.User{}).Where("id = ?", id).Update("password", password).Error
}

// FindAllOptions 获取所有用户选项（仅 id, username, nickname）
func (r *UserRepository) FindAllOptions() ([]modelsystem.User, error) {
	var users []modelsystem.User
	err := r.db.Select("id, username, nickname").Where("status = ?", "active").Find(&users).Error
	return users, err
}
