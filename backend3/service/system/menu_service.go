package system

import (
	"errors"

	modelsystem "oneops/backend3/model/system"
	reposystem "oneops/backend3/repository/system"
)

// ErrMenuHasChildren 菜单下有子菜单
var ErrMenuHasChildren = errors.New("该菜单下有子菜单，无法删除")

// MenuService 菜单业务逻辑层
type MenuService struct {
	repo *reposystem.MenuRepository
}

// NewMenuService 创建菜单服务
func NewMenuService(repo *reposystem.MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

// GetAllAsTree 获取所有菜单并构建树形结构
func (s *MenuService) GetAllAsTree() ([]*modelsystem.Menu, error) {
	menus, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(menus, 0), nil
}

// Create 创建菜单
func (s *MenuService) Create(menu *modelsystem.Menu) error {
	return s.repo.Create(menu)
}

// Update 根据ID更新指定字段
func (s *MenuService) Update(id uint64, updates map[string]interface{}) error {
	return s.repo.UpdateByID(id, updates)
}

// Delete 删除菜单（先检查子菜单）
func (s *MenuService) Delete(id uint64) error {
	count, err := s.repo.CountByParentID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrMenuHasChildren
	}
	return s.repo.Delete(id)
}

// buildMenuTree 递归构建菜单树
func (s *MenuService) buildMenuTree(menus []modelsystem.Menu, parentID uint) []*modelsystem.Menu {
	var result []*modelsystem.Menu
	for _, menu := range menus {
		if menu.ParentID == parentID {
			children := s.buildMenuTree(menus, menu.ID)

			menuItem := &modelsystem.Menu{
				ID:         menu.ID,
				Name:       menu.Name,
				Icon:       menu.Icon,
				Path:       menu.Path,
				Permission: menu.Permission,
				MenuType:   menu.MenuType,
				ParentID:   menu.ParentID,
				Sort:       menu.Sort,
				Status:     menu.Status,
				CreatedAt:  menu.CreatedAt,
				UpdatedAt:  menu.UpdatedAt,
			}

			menuItem.Children = make([]*modelsystem.Menu, 0)
			if len(children) > 0 {
				menuItem.Children = children
			}

			result = append(result, menuItem)
		}
	}
	return result
}
