package cmdb

import (
	modelcmdb "oneops/backend3/model/cmdb"
)

// ========== 标签管理 ==========

// GetServerTags 获取服务器标签列表
func (s *CMDBService) GetServerTags() ([]modelcmdb.ServerTag, error) {
	return s.tagRepo.FindServerTags()
}

// CreateServerTag 创建服务器标签
func (s *CMDBService) CreateServerTag(tag *modelcmdb.ServerTag) error {
	return s.tagRepo.CreateServerTag(tag)
}

// UpdateServerTag 更新服务器标签
func (s *CMDBService) UpdateServerTag(id uint, updates map[string]interface{}) error {
	return s.tagRepo.UpdateServerTag(id, updates)
}

// DeleteServerTag 删除服务器标签
func (s *CMDBService) DeleteServerTag(id uint) error {
	return s.tagRepo.DeleteServerTag(id)
}

// AssignServerTag 为服务器分配标签
func (s *CMDBService) AssignServerTag(serverID, tagID uint) error {
	return s.tagRepo.AssignServerTag(serverID, tagID)
}

// RemoveServerTag 移除服务器标签
func (s *CMDBService) RemoveServerTag(serverID, tagID uint) error {
	return s.tagRepo.RemoveServerTag(serverID, tagID)
}
