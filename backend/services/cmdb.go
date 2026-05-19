package services

import (
	"fmt"
	"oneops/backend/models"
)

// CMDBService CMDB服务
type CMDBService struct{}

// NewCMDBService 创建CMDB服务
func NewCMDBService() *CMDBService {
	return &CMDBService{}
}

// ========== 服务器管理 ==========

// GetServers 获取服务器列表
func (s *CMDBService) GetServers(query map[string]interface{}, page, pageSize int) ([]models.Server, int64, error) {
	var servers []models.Server
	var total int64

	tx := db.Model(&models.Server{})

	// 构建查询条件
	if hostname, ok := query["hostname"].(string); ok && hostname != "" {
		tx = tx.Where("hostname LIKE ?", "%"+hostname+"%")
	}
	if ip, ok := query["ip"].(string); ok && ip != "" {
		tx = tx.Where("ip LIKE ?", "%"+ip+"%")
	}
	if env, ok := query["env"].(string); ok && env != "" {
		tx = tx.Where("env = ?", env)
	}
	if status, ok := query["status"].(string); ok && status != "" {
		tx = tx.Where("status = ?", status)
	}
	if businessID, ok := query["businessId"].(uint); ok && businessID > 0 {
		tx = tx.Where("business_id = ?", businessID)
	}
	if provider, ok := query["provider"].(string); ok && provider != "" {
		tx = tx.Where("provider = ?", provider)
	}

	// 标记是否需要JOIN分组表
	needsGroupJoin := false
	if groupID, ok := query["groupId"].(uint); ok && groupID > 0 {
		needsGroupJoin = true
		tx = tx.Joins("JOIN server_group_relations ON servers.id = server_group_relations.server_id").
			Where("server_group_relations.group_id = ?", groupID)
	}

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 预加载关联数据
	queryTx := tx.Preload("Business").
		Preload("Cabinet").
		Preload("Cabinet.Room").
		Preload("Tags").
		Preload("Groups")
	if needsGroupJoin {
		queryTx = queryTx.Distinct("servers.id")
	}
	err := queryTx.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&servers).Error

	return servers, total, err
}

// GetServerByID 根据ID获取服务器
func (s *CMDBService) GetServerByID(id uint) (*models.Server, error) {
	var server models.Server
	err := db.Preload("Business").
		Preload("Cabinet").
		Preload("Cabinet.Room").
		Preload("Tags").
		First(&server, id).Error
	return &server, err
}

// CreateServer 创建服务器
func (s *CMDBService) CreateServer(server *models.Server, operator string) error {
	// 记录变更
	if err := s.recordAssetChange("server", 0, "create", "", "", operator, "创建服务器", ""); err != nil {
		return err
	}

	return db.Create(server).Error
}

// UpdateServer 更新服务器
func (s *CMDBService) UpdateServer(id uint, updates map[string]interface{}, operator string) error {
	var oldServer models.Server
	if err := db.First(&oldServer, id).Error; err != nil {
		return err
	}

	// 记录变更字段
	for field, newValue := range updates {
		oldValue := fmt.Sprintf("%v", getFieldValue(&oldServer, field))
		newValueStr := fmt.Sprintf("%v", newValue)
		if oldValue != newValueStr {
			s.recordAssetChange("server", id, "update", field, oldValue, newValueStr, operator, "")
		}
	}

	return db.Model(&models.Server{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServer 删除服务器
func (s *CMDBService) DeleteServer(id uint, operator string) error {
	var server models.Server
	if err := db.First(&server, id).Error; err != nil {
		return err
	}

	// 记录变更
	if err := s.recordAssetChange("server", id, "delete", "", "", operator, "删除服务器", ""); err != nil {
		return err
	}

	return db.Delete(&server).Error
}

// ========== 业务系统管理 ==========

// GetBusinessUnits 获取业务系统列表（树形结构）
func (s *CMDBService) GetBusinessUnits() ([]models.BusinessUnit, error) {
	var units []models.BusinessUnit
	err := db.Order("sort_order ASC, id ASC").Find(&units).Error
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	return s.buildBusinessTree(units, 0), nil
}

// buildBusinessTree 构建业务树
func (s *CMDBService) buildBusinessTree(units []models.BusinessUnit, parentID uint) []models.BusinessUnit {
	var tree []models.BusinessUnit
	for _, unit := range units {
		if unit.ParentID == parentID {
			unit.Children = s.buildBusinessTree(units, unit.ID)
			tree = append(tree, unit)
		}
	}
	return tree
}

// GetBusinessUnitByID 根据ID获取业务系统
func (s *CMDBService) GetBusinessUnitByID(id uint) (*models.BusinessUnit, error) {
	var unit models.BusinessUnit
	err := db.First(&unit, id).Error
	return &unit, err
}

// CreateBusinessUnit 创建业务系统
func (s *CMDBService) CreateBusinessUnit(unit *models.BusinessUnit, operator string) error {
	// 设置层级
	if unit.ParentID > 0 {
		var parent models.BusinessUnit
		if err := db.First(&parent, unit.ParentID).Error; err != nil {
			return err
		}
		unit.Level = parent.Level + 1
	} else {
		unit.Level = 1
	}

	return db.Create(unit).Error
}

// UpdateBusinessUnit 更新业务系统
func (s *CMDBService) UpdateBusinessUnit(id uint, updates map[string]interface{}) error {
	return db.Model(&models.BusinessUnit{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteBusinessUnit 删除业务系统
func (s *CMDBService) DeleteBusinessUnit(id uint) error {
	// 检查是否有子业务
	var count int64
	if err := db.Model(&models.BusinessUnit{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该业务下有子业务，无法删除")
	}

	// 检查是否有关联服务器
	if err := db.Model(&models.Server{}).Where("business_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("该业务下有关联服务器，无法删除")
	}

	return db.Delete(&models.BusinessUnit{}, id).Error
}

// ========== 机柜管理 ==========

// GetServerRooms 获取机房列表
func (s *CMDBService) GetServerRooms() ([]models.ServerRoom, error) {
	var rooms []models.ServerRoom
	err := db.Preload("Cabinets").Order("id ASC").Find(&rooms).Error
	return rooms, err
}

// CreateServerRoom 创建机房
func (s *CMDBService) CreateServerRoom(room *models.ServerRoom) error {
	return db.Create(room).Error
}

// UpdateServerRoom 更新机房
func (s *CMDBService) UpdateServerRoom(id uint, updates map[string]interface{}) error {
	return db.Model(&models.ServerRoom{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServerRoom 删除机房
func (s *CMDBService) DeleteServerRoom(id uint) error {
	return db.Delete(&models.ServerRoom{}, id).Error
}

// GetCabinets 获取机柜列表
func (s *CMDBService) GetCabinets(roomID uint) ([]models.Cabinet, error) {
	var cabinets []models.Cabinet
	tx := db.Preload("Room")
	if roomID > 0 {
		tx = tx.Where("room_id = ?", roomID)
	}
	err := tx.Order("id ASC").Find(&cabinets).Error
	return cabinets, err
}

// ========== 标签管理 ==========

// GetServerTags 获取服务器标签列表
func (s *CMDBService) GetServerTags() ([]models.ServerTag, error) {
	var tags []models.ServerTag
	err := db.Order("sort_order ASC").Find(&tags).Error
	return tags, err
}

// CreateServerTag 创建服务器标签
func (s *CMDBService) CreateServerTag(tag *models.ServerTag) error {
	return db.Create(tag).Error
}

// UpdateServerTag 更新服务器标签
func (s *CMDBService) UpdateServerTag(id uint, updates map[string]interface{}) error {
	return db.Model(&models.ServerTag{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServerTag 删除服务器标签
func (s *CMDBService) DeleteServerTag(id uint) error {
	return db.Delete(&models.ServerTag{}, id).Error
}

// AssignServerTag 为服务器分配标签
func (s *CMDBService) AssignServerTag(serverID, tagID uint) error {
	// 检查是否已存在
	var count int64
	db.Model(&models.ServerTagRelation{}).Where("server_id = ? AND tag_id = ?", serverID, tagID).Count(&count)
	if count > 0 {
		return nil // 已存在，无需重复添加
	}

	relation := &models.ServerTagRelation{
		ServerID: serverID,
		TagID:    tagID,
	}
	return db.Create(relation).Error
}

// RemoveServerTag 移除服务器标签
func (s *CMDBService) RemoveServerTag(serverID, tagID uint) error {
	return db.Where("server_id = ? AND tag_id = ?", serverID, tagID).Delete(&models.ServerTagRelation{}).Error
}

// ========== 资产变更记录 ==========

// GetAssetChanges 获取资产变更记录
func (s *CMDBService) GetAssetChanges(assetType string, assetID uint, page, pageSize int) ([]models.AssetChange, int64, error) {
	var changes []models.AssetChange
	var total int64

	tx := db.Model(&models.AssetChange{})
	if assetType != "" {
		tx = tx.Where("asset_type = ?", assetType)
	}
	if assetID > 0 {
		tx = tx.Where("asset_id = ?", assetID)
	}

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Order("operate_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&changes).Error

	return changes, total, err
}

// recordAssetChange 记录资产变更
func (s *CMDBService) recordAssetChange(assetType string, assetID uint, changeType, fieldName, oldValue, newValue, operator, remarks string) error {
	change := &models.AssetChange{
		AssetType:  assetType,
		AssetID:    assetID,
		FieldName:  fieldName,
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeType: changeType,
		Operator:   operator,
		Remarks:    remarks,
	}
	return db.Create(change).Error
}

// ========== 辅助函数 ==========

// getFieldValue 获取结构体字段值
func getFieldValue(obj interface{}, field string) interface{} {
	// 简化实现，实际可以使用反射
	return ""
}

// GetServerStats 获取服务器统计信息
func (s *CMDBService) GetServerStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总数
	var total int64
	db.Model(&models.Server{}).Count(&total)
	stats["total"] = total

	// 按环境统计
	var envStats []struct {
		Env   string
		Count int64
	}
	db.Model(&models.Server{}).Select("env, count(*) as count").Group("env").Scan(&envStats)
	envMap := make(map[string]int64)
	for _, stat := range envStats {
		envMap[stat.Env] = stat.Count
	}
	stats["byEnv"] = envMap

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	db.Model(&models.Server{}).Select("status, count(*) as count").Group("status").Scan(&statusStats)
	statusMap := make(map[string]int64)
	for _, stat := range statusStats {
		statusMap[stat.Status] = stat.Count
	}
	stats["byStatus"] = statusMap

	// 按服务商统计
	var providerStats []struct {
		Provider string
		Count    int64
	}
	db.Model(&models.Server{}).Select("provider, count(*) as count").Group("provider").Scan(&providerStats)
	providerMap := make(map[string]int64)
	for _, stat := range providerStats {
		providerMap[stat.Provider] = stat.Count
	}
	stats["byProvider"] = providerMap

	return stats, nil
}

// GetServerConfig 通过SSH获取服务器配置信息
func (s *CMDBService) GetServerConfig(hostname, ip, sshUser string, sshPort int) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	// TODO: 实现SSH连接和命令执行
	// 这里需要使用golang.org/x/crypto/ssh库来连接服务器
	// 然后执行命令获取配置信息

	// 临时返回模拟数据，实际应该通过SSH获取
	config["cpu"] = 4
	config["memory"] = 8  // GB
	config["disk"] = 200  // GB
	config["os"] = "Ubuntu"
	config["osVersion"] = "22.04"
	config["arch"] = "x86_64"
	config["hostname"] = hostname

	return config, nil
}

// ========== 主机分组管理 ==========

// GetServerGroups 获取主机分组列表（树形结构）
func (s *CMDBService) GetServerGroups() ([]models.ServerGroup, error) {
	var groups []models.ServerGroup
	err := db.Preload("Servers").Order("sort_order ASC, id ASC").Find(&groups).Error
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	return s.buildGroupTree(groups, 0), nil
}

// buildGroupTree 构建分组树
func (s *CMDBService) buildGroupTree(groups []models.ServerGroup, parentID uint) []models.ServerGroup {
	var result []models.ServerGroup
	for _, group := range groups {
		if group.ParentID == parentID {
			group.Children = s.buildGroupTree(groups, group.ID)
			result = append(result, group)
		}
	}
	return result
}

// GetServerGroupByID 根据ID获取主机分组
func (s *CMDBService) GetServerGroupByID(id uint) (*models.ServerGroup, error) {
	var group models.ServerGroup
	err := db.Preload("Servers").Preload("Parent").First(&group, id).Error
	return &group, err
}

// CreateServerGroup 创建主机分组
func (s *CMDBService) CreateServerGroup(group *models.ServerGroup) error {
	return db.Create(group).Error
}

// UpdateServerGroup 更新主机分组
func (s *CMDBService) UpdateServerGroup(id uint, updates map[string]interface{}) error {
	return db.Model(&models.ServerGroup{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServerGroup 删除主机分组
func (s *CMDBService) DeleteServerGroup(id uint) error {
	return db.Delete(&models.ServerGroup{}, id).Error
}

// AssignServerToGroup 将服务器分配到分组
func (s *CMDBService) AssignServerToGroup(serverID, groupID uint) error {
	// 先删除该服务器的所有分组关联
	db.Where("server_id = ?", serverID).Delete(&models.ServerGroupRelation{})

	// 创建新的分组关联
	return db.Create(&models.ServerGroupRelation{
		ServerID: serverID,
		GroupID:   groupID,
	}).Error
}

// RemoveServerFromGroup 将服务器从分组中移除
func (s *CMDBService) RemoveServerFromGroup(serverID, groupID uint) error {
	return db.Where("server_id = ? AND group_id = ?", serverID, groupID).Delete(&models.ServerGroupRelation{}).Error
}

// GetServersByGroup 获取指定分组下的服务器列表
func (s *CMDBService) GetServersByGroup(groupID uint) ([]models.Server, error) {
	var servers []models.Server
	err := db.Joins("JOIN server_group_relations ON servers.id = server_group_relations.server_id").
		Where("server_group_relations.group_id = ?", groupID).
		Find(&servers).Error
	return servers, err
}

// ========== SSH凭证管理 ==========

// GetSSHCredentials 获取SSH凭证列表
func (s *CMDBService) GetSSHCredentials() ([]models.SSHCredential, error) {
	var credentials []models.SSHCredential
	err := db.Order("sort_order ASC, id ASC").Find(&credentials).Error
	return credentials, err
}

// GetSSHCredentialByID 根据ID获取SSH凭证
func (s *CMDBService) GetSSHCredentialByID(id uint) (*models.SSHCredential, error) {
	var credential models.SSHCredential
	err := db.First(&credential, id).Error
	return &credential, err
}

// CreateSSHCredential 创建SSH凭证
func (s *CMDBService) CreateSSHCredential(credential *models.SSHCredential) error {
	// 加密密码和私钥
	if credential.Password != "" {
		// TODO: 实际应该使用加密算法加密
		credential.Password = credential.Password
	}
	if credential.PrivateKey != "" {
		// TODO: 实际应该使用加密算法加密
		credential.PrivateKey = credential.PrivateKey
	}
	return db.Create(credential).Error
}

// UpdateSSHCredential 更新SSH凭证
func (s *CMDBService) UpdateSSHCredential(id uint, updates map[string]interface{}) error {
	// 如果有密码或私钥更新，需要加密
	if _, ok := updates["password"]; ok {
		// TODO: 加密处理
	}
	if _, ok := updates["private_key"]; ok {
		// TODO: 加密处理
	}
	return db.Model(&models.SSHCredential{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteSSHCredential 删除SSH凭证
func (s *CMDBService) DeleteSSHCredential(id uint) error {
	return db.Delete(&models.SSHCredential{}, id).Error
}

// TestSSHCredential 测试SSH凭证连接
func (s *CMDBService) TestSSHCredential(id uint, testIP string) (map[string]interface{}, error) {
	credential, err := s.GetSSHCredentialByID(id)
	if err != nil {
		return nil, err
	}

	// TODO: 实现SSH连接测试
	result := make(map[string]interface{})
	result["success"] = true
	result["message"] = "连接测试功能开发中"
	result["credential"] = credential.Username
	result["test_ip"] = testIP

	return result, nil
}
