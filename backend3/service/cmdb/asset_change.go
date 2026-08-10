package cmdb

import (
	modelcmdb "oneops/backend3/model/cmdb"
)

// ========== 资产变更记录 ==========

// GetAssetChanges 获取资产变更记录
func (s *CMDBService) GetAssetChanges(assetType string, assetID uint, page, pageSize int) ([]modelcmdb.AssetChange, int64, error) {
	return s.repo.FindAssetChanges(assetType, assetID, page, pageSize)
}

// recordAssetChange 记录资产变更
func (s *CMDBService) recordAssetChange(assetType string, assetID uint, changeType, fieldName, oldValue, newValue, operator, remarks string) error {
	change := &modelcmdb.AssetChange{
		AssetType:  assetType,
		AssetID:    assetID,
		FieldName:  fieldName,
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeType: changeType,
		Operator:   operator,
		Remarks:    remarks,
	}
	return s.repo.CreateAssetChange(change)
}
