package cmdb

import (
	modelcmdb "oneops/backend3/model/cmdb"
)

// ========== 机柜管理 ==========

// GetServerRooms 获取机房列表
func (s *CMDBService) GetServerRooms() ([]modelcmdb.ServerRoom, error) {
	return s.repo.FindServerRooms()
}

// CreateServerRoom 创建机房
func (s *CMDBService) CreateServerRoom(room *modelcmdb.ServerRoom) error {
	return s.repo.CreateServerRoom(room)
}

// UpdateServerRoom 更新机房
func (s *CMDBService) UpdateServerRoom(id uint, updates map[string]interface{}) error {
	return s.repo.UpdateServerRoom(id, updates)
}

// DeleteServerRoom 删除机房
func (s *CMDBService) DeleteServerRoom(id uint) error {
	return s.repo.DeleteServerRoom(id)
}

// GetCabinets 获取机柜列表
func (s *CMDBService) GetCabinets(roomID uint) ([]modelcmdb.Cabinet, error) {
	return s.repo.FindCabinets(roomID)
}
