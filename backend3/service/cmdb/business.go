package cmdb

import (
	"fmt"

	modelcmdb "oneops/backend3/model/cmdb"
)

// ========== 业务系统管理 ==========

// GetBusinessUnits 获取业务系统列表（树形结构）
func (s *CMDBService) GetBusinessUnits() ([]modelcmdb.BusinessUnit, error) {
	units, err := s.repo.FindBusinessUnits()
	if err != nil {
		return nil, err
	}
	return s.buildBusinessTreeOptimized(units), nil
}

// buildBusinessTreeOptimized 构建业务树（O(n)时间复杂度）
func (s *CMDBService) buildBusinessTreeOptimized(units []modelcmdb.BusinessUnit) []modelcmdb.BusinessUnit {
	parentMap := make(map[uint][]*modelcmdb.BusinessUnit)

	for i := range units {
		parentMap[units[i].ParentID] = append(parentMap[units[i].ParentID], &units[i])
	}

	var roots []modelcmdb.BusinessUnit
	for _, unit := range parentMap[0] {
		*unit = s.buildTreeNode(unit, parentMap)
		roots = append(roots, *unit)
	}

	return roots
}

// buildTreeNode 递归构建树节点
func (s *CMDBService) buildTreeNode(node *modelcmdb.BusinessUnit, parentMap map[uint][]*modelcmdb.BusinessUnit) modelcmdb.BusinessUnit {
	children := parentMap[node.ID]
	if len(children) == 0 {
		return *node
	}

	node.Children = make([]modelcmdb.BusinessUnit, len(children))
	for i, child := range children {
		node.Children[i] = s.buildTreeNode(child, parentMap)
	}

	return *node
}

// GetBusinessUnitByID 根据ID获取业务系统
func (s *CMDBService) GetBusinessUnitByID(id uint) (*modelcmdb.BusinessUnit, error) {
	return s.repo.FindBusinessUnitByID(id)
}

// CreateBusinessUnit 创建业务系统
func (s *CMDBService) CreateBusinessUnit(unit *modelcmdb.BusinessUnit, operator string) error {
	if unit.ParentID > 0 {
		parent, err := s.repo.FindBusinessUnitByID(unit.ParentID)
		if err != nil {
			return err
		}
		unit.Level = parent.Level + 1
	} else {
		unit.Level = 1
	}

	return s.repo.CreateBusinessUnit(unit)
}

// UpdateBusinessUnit 更新业务系统
func (s *CMDBService) UpdateBusinessUnit(id uint, updates map[string]interface{}) error {
	return s.repo.UpdateBusinessUnit(id, updates)
}

// DeleteBusinessUnit 删除业务系统
func (s *CMDBService) DeleteBusinessUnit(id uint) error {
	count, err := s.repo.CountBusinessUnitChildren(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该业务下有子业务，无法删除")
	}

	return s.repo.DeleteBusinessUnit(id)
}
