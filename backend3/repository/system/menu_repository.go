package system

import (
	modelsystem "oneops/backend3/model/system"

	"gorm.io/gorm"
)

// MenuRepository 菜单数据访问层
type MenuRepository struct {
	db *gorm.DB
}

// NewMenuRepository 创建菜单仓库
func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// FindAll 查询所有菜单（按排序字段升序）
func (r *MenuRepository) FindAll() ([]modelsystem.Menu, error) {
	var menus []modelsystem.Menu
	if err := r.db.Order("sort ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// Create 创建菜单
func (r *MenuRepository) Create(menu *modelsystem.Menu) error {
	return r.db.Create(menu).Error
}

// UpdateByID 根据ID更新指定字段
func (r *MenuRepository) UpdateByID(id uint64, updates map[string]interface{}) error {
	return r.db.Model(&modelsystem.Menu{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 根据ID删除菜单
func (r *MenuRepository) Delete(id uint64) error {
	return r.db.Delete(&modelsystem.Menu{}, id).Error
}

// CountByParentID 统计指定父菜单下的子菜单数量
func (r *MenuRepository) CountByParentID(parentID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&modelsystem.Menu{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}
