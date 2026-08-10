package cmdb

import (
	"fmt"

	modelcmdb "oneops/backend3/model/cmdb"
)

// ========== 服务器管理 ==========

// GetServersLight 获取服务器列表（轻量级，仅返回显示字段）
func (s *CMDBService) GetServersLight(query map[string]interface{}, page, pageSize int) ([]map[string]interface{}, int64, error) {
	return s.repo.FindServersLight(query, page, pageSize)
}

// GetServers 获取服务器列表（完整数据，含关联）
func (s *CMDBService) GetServers(query map[string]interface{}, page, pageSize int) ([]modelcmdb.Server, int64, error) {
	return s.repo.FindServers(query, page, pageSize)
}

// GetServerByID 根据ID获取服务器
func (s *CMDBService) GetServerByID(id uint) (*modelcmdb.Server, error) {
	return s.repo.FindServerByID(id)
}

// GetServerForConnect 获取连接所需的服务器信息（轻量级）
func (s *CMDBService) GetServerForConnect(id uint) (*modelcmdb.Server, error) {
	return s.repo.FindServerForConnect(id)
}

// CreateServer 创建服务器
func (s *CMDBService) CreateServer(server *modelcmdb.Server, operator string) error {
	if err := s.recordAssetChange("server", 0, "create", "", "", operator, "创建服务器", ""); err != nil {
		return err
	}

	if err := s.repo.CreateServer(server); err != nil {
		return err
	}

	return s.repo.UpdateServerRedundantFields(server.ID)
}

// UpdateServer 更新服务器
func (s *CMDBService) UpdateServer(id uint, updates map[string]interface{}, operator string) error {
	oldServer, err := s.repo.FindServerForUpdate(id)
	if err != nil {
		return err
	}

	for field, newValue := range updates {
		oldValue := fmt.Sprintf("%v", getFieldValue(&oldServer, field))
		newValueStr := fmt.Sprintf("%v", newValue)
		if oldValue != newValueStr {
			s.recordAssetChange("server", id, "update", field, oldValue, newValueStr, operator, "")
		}
	}

	columnUpdates := filterServerColumns(updates)

	var groupIDs []uint
	if rawGroupIDs, exists := updates["groupIds"]; exists {
		groupIDs = extractUintSlice(rawGroupIDs)
	}

	var credentialIDs []uint
	if rawCredIDs, exists := updates["credentialIds"]; exists {
		credentialIDs = extractUintSlice(rawCredIDs)
	}

	if err := s.repo.UpdateServer(id, columnUpdates, groupIDs, rawExists(updates, "groupIds"), credentialIDs, rawExists(updates, "credentialIds")); err != nil {
		return err
	}

	_, hasGroupUpdate := updates["groupIds"]
	_, hasCredUpdate := updates["credentialIds"]
	if hasGroupUpdate || hasCredUpdate {
		if err := s.repo.UpdateServerRedundantFields(id); err != nil {
			return err
		}
	}

	return nil
}

// DeleteServer 删除服务器
func (s *CMDBService) DeleteServer(id uint, operator string) error {
	if err := s.recordAssetChange("server", id, "delete", "", "", operator, "删除服务器", ""); err != nil {
		return err
	}

	return s.repo.DeleteServer(id)
}
