package services

import (
	"encoding/json"
	"fmt"
	"oneops/backend/models"
	"strconv"
)

// CMDBService CMDB服务
type CMDBService struct{}

// NewCMDBService 创建CMDB服务
func NewCMDBService() *CMDBService {
	return &CMDBService{}
}

// ========== 服务器管理 ==========

// GetServersLight 获取服务器列表（轻量级，仅返回显示字段）
func (s *CMDBService) GetServersLight(query map[string]interface{}, page, pageSize int) ([]map[string]interface{}, int64, error) {
	var servers []map[string]interface{}
	var total int64

	tx := db.Model(&models.Server{}).Select("id, hostname, ip, inner_ip, env, status, provider, agent_status, cpu, memory, os, arch, ssh_port")

	// 应用相同的过滤条件（复用逻辑）
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
	if provider, ok := query["provider"].(string); ok && provider != "" {
		tx = tx.Where("provider = ?", provider)
	}
	if agentStatus, ok := query["agentStatus"].(string); ok && agentStatus != "" {
		tx = tx.Where("agent_status = ?", agentStatus)
	}

	var groupIDUint uint
	if groupID, ok := query["groupId"]; ok && groupID != nil {
		switch v := groupID.(type) {
		case uint:
			groupIDUint = v
		case uint64:
			groupIDUint = uint(v)
		case int:
			groupIDUint = uint(v)
		case int64:
			groupIDUint = uint(v)
		case float64:
			groupIDUint = uint(v)
		case float32:
			groupIDUint = uint(v)
		default:
			if strVal, ok := groupID.(string); ok {
				if parsedVal, err := strconv.ParseUint(strVal, 10, 32); err == nil {
					groupIDUint = uint(parsedVal)
				}
			}
		}
		if groupIDUint > 0 {
			tx = tx.Where("id IN (SELECT server_id FROM server_group_relations WHERE group_id = ?)", groupIDUint)
		}
	}

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 单次查询获取列表数据（不含关联）
	err := tx.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&servers).Error

	return servers, total, err
}

// GetServers 获取服务器列表（完整数据，含关联）
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
	if provider, ok := query["provider"].(string); ok && provider != "" {
		tx = tx.Where("provider = ?", provider)
	}
	if agentStatus, ok := query["agentStatus"].(string); ok && agentStatus != "" {
		tx = tx.Where("agent_status = ?", agentStatus)
	}

	// 标记是否需要过滤分组
	var groupIDUint uint
	if groupID, ok := query["groupId"]; ok && groupID != nil {
		// 处理多种数字类型
		switch v := groupID.(type) {
		case uint:
			groupIDUint = v
		case uint64:
			groupIDUint = uint(v)
		case int:
			groupIDUint = uint(v)
		case int64:
			groupIDUint = uint(v)
		case float64:
			groupIDUint = uint(v)
		case float32:
			groupIDUint = uint(v)
		default:
			// 尝试转换字符串
			if strVal, ok := groupID.(string); ok {
				if parsedVal, err := strconv.ParseUint(strVal, 10, 32); err == nil {
					groupIDUint = uint(parsedVal)
				}
			}
		}

		// 使用子查询过滤分组
		if groupIDUint > 0 {
			tx = tx.Where("id IN (SELECT server_id FROM server_group_relations WHERE group_id = ?)", groupIDUint)
		}
	}

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 列表页查询（性能优化：使用冗余字段，避免关联查询）
	err := tx.
		Select(`
			servers.id, servers.hostname, servers.ip, servers.inner_ip, servers.ssh_port, servers.env, servers.status,
			servers.provider, servers.agent_status, servers.agent_version, servers.cpu,
			servers.memory, servers.os, servers.arch, servers.created_at, servers.updated_at,
			servers.group_names, servers.credential_names, servers.system_credential_id
		`).
		Order("servers.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&servers).Error

	// 使用冗余字段填充关联数据（包含ID和名称，便于编辑回显）
	for i := range servers {
		// 从冗余字段解析分组ID和名称
		if servers[i].GroupNames != "[]" && servers[i].GroupNames != "" && servers[i].GroupNames != "null" {
			var groupData []map[string]interface{}
			json.Unmarshal([]byte(servers[i].GroupNames), &groupData)
			for _, item := range groupData {
				servers[i].Groups = append(servers[i].Groups, models.ServerGroup{
					ID:   uint(item["id"].(float64)),
					Name: item["name"].(string),
				})
			}
		}

		// 从冗余字段解析凭证ID、名称和类型
		if servers[i].CredentialNames != "[]" && servers[i].CredentialNames != "" && servers[i].CredentialNames != "null" {
			var credData []map[string]interface{}
			json.Unmarshal([]byte(servers[i].CredentialNames), &credData)
			for _, item := range credData {
				cred := models.SSHCredential{
					ID:   uint(item["id"].(float64)),
					Name: item["name"].(string),
				}
				// 填充凭证类型（如果存在）
				if credentialType, ok := item["credential_type"]; ok {
					cred.CredentialType = models.CredentialType(credentialType.(string))
				} else {
					// 默认为用户凭证（向后兼容）
					cred.CredentialType = models.CredentialTypeUser
				}
				servers[i].Credentials = append(servers[i].Credentials, cred)
			}
		}

			// 处理系统运维凭证（单个凭证，通过ID查询）
			if servers[i].SystemCredentialID > 0 {
			var systemCred models.SSHCredential
			if err := db.Select("id, name, credential_type").First(&systemCred, servers[i].SystemCredentialID).Error; err == nil {
			servers[i].SystemCredential = &systemCred
			}
			}
	}

	return servers, total, err
}

// GetServerByID 根据ID获取服务器
func (s *CMDBService) GetServerByID(id uint) (*models.Server, error) {
	var server models.Server
	err := db.
		Preload("Cabinet").
		Preload("Cabinet.Room").
		Preload("Tags").
		Preload("Credentials").
		Preload("Groups").
		Preload("Attributes").
		First(&server, id).Error
	return &server, err
}

// GetServerForConnect 获取连接所需的服务器信息（轻量级）
func (s *CMDBService) GetServerForConnect(id uint) (*models.Server, error) {
	var server models.Server
	// 只 Preload Credentials，其他不需要的数据不加载
	err := db.Preload("Credentials").First(&server, id).Error
	return &server, err
}

// CreateServer 创建服务器
func (s *CMDBService) CreateServer(server *models.Server, operator string) error {
	// 记录变更
	if err := s.recordAssetChange("server", 0, "create", "", "", operator, "创建服务器", ""); err != nil {
		return err
	}

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建服务器
	if err := tx.Create(server).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 处理分组关联
	if len(server.GroupIDs) > 0 {
		for _, groupID := range server.GroupIDs {
			relation := models.ServerGroupRelation{
				ServerID: server.ID,
				GroupID:  groupID,
			}
			if err := tx.Create(&relation).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("创建分组关联失败: %w", err)
			}
		}
	}

	// 处理凭证关联（多凭证）
	if len(server.CredentialIDs) > 0 {
		for _, credID := range server.CredentialIDs {
			rel := models.ServerCredential{ServerID: server.ID, CredentialID: credID}
			if err := tx.Create(&rel).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("创建凭证关联失败: %w", err)
			}
		}
	} else if server.SSHCredentialID != 0 {
		// 向后兼容：单凭证字段自动迁移到多凭证表
		rel := models.ServerCredential{ServerID: server.ID, CredentialID: server.SSHCredentialID}
		tx.Create(&rel) // 忽略错误（重复时不影响流程）
	} else if server.CredentialID != 0 {
		rel := models.ServerCredential{ServerID: server.ID, CredentialID: server.CredentialID}
		tx.Create(&rel)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 维护冗余字段：group_names 和 credential_names
	return s.updateServerRedundantFields(server.ID)
}

// updateServerRedundantFields 更新服务器的冗余字段
func (s *CMDBService) updateServerRedundantFields(serverID uint) error {
	// 查询分组ID和名称
	var groupData []map[string]interface{}
	db.Table("server_group_relations").
		Select("g.id, g.name").
		Joins("JOIN server_groups g ON g.id = server_group_relations.group_id").
		Where("server_group_relations.server_id = ?", serverID).
		Scan(&groupData)

	// 查询凭证ID、名称和类型
	var credData []map[string]interface{}
	db.Table("server_credentials").
		Select("c.id, c.name, c.credential_type").
		Joins("JOIN ssh_credentials c ON c.id = server_credentials.credential_id").
		Where("server_credentials.server_id = ?", serverID).
		Scan(&credData)

	// 构建JSON数组（包含ID和名称）
	groupJSON, _ := json.Marshal(groupData)
	credJSON, _ := json.Marshal(credData)

	// 更新冗余字段
	return db.Model(&models.Server{}).
		Where("id = ?", serverID).
		Updates(map[string]interface{}{
			"group_names":       string(groupJSON),
			"credential_names":  string(credJSON),
		}).Error
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

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 过滤掉关联对象和虚拟字段，只保留实际数据库列
	columnUpdates := filterServerColumns(updates)

	// 更新服务器基本信息
	if err := tx.Model(&models.Server{}).Where("id = ?", id).Updates(columnUpdates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 处理分组关联更新
	if rawGroupIDs, exists := updates["groupIds"]; exists {
		groupIDs := extractUintSlice(rawGroupIDs)
		// 删除旧的分组关联
		if err := tx.Where("server_id = ?", id).Delete(&models.ServerGroupRelation{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除旧分组关联失败: %w", err)
		}

		// 创建新的分组关联
		for _, groupID := range groupIDs {
			relation := models.ServerGroupRelation{
				ServerID: id,
				GroupID:  groupID,
			}
			if err := tx.Create(&relation).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("创建分组关联失败: %w", err)
			}
		}
	}

	// 处理凭证关联更新（支持 JSON 数组，来自前端）
	if rawCredIDs, exists := updates["credentialIds"]; exists {
		credentialIDs := extractUintSlice(rawCredIDs)
		// 删除旧的凭证关联
		if err := tx.Where("server_id = ?", id).Delete(&models.ServerCredential{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除旧凭证关联失败: %w", err)
		}
		// 创建新的凭证关联
		for _, credID := range credentialIDs {
			rel := models.ServerCredential{ServerID: id, CredentialID: credID}
			if err := tx.Create(&rel).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("创建凭证关联失败: %w", err)
			}
		}
		// 同步 SSHCredentialID（取第一个凭证作为默认，向后兼容）
		if len(credentialIDs) > 0 {
			tx.Model(&models.Server{}).Where("id = ?", id).Update("ssh_credential_id", credentialIDs[0])
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 如果更新了分组或凭证关联，维护冗余字段
	_, hasGroupUpdate := updates["groupIds"]
	_, hasCredUpdate := updates["credentialIds"]
	if hasGroupUpdate || hasCredUpdate {
		if err := s.updateServerRedundantFields(id); err != nil {
			return err
		}
	}

	return nil
}

// filterServerColumns 过滤掉非数据库列字段（关联对象、虚拟字段），只保留可直接更新的列
func filterServerColumns(updates map[string]interface{}) map[string]interface{} {
	// 这些键是关联对象或 gorm:"-" 虚拟字段，不能直接作为 SQL 列名
	skipKeys := map[string]bool{
		"cloudInfo":    true,
		"credential":   true,
		"credentials":  true,
		"credentialIds": true,
		"cabinet":      true,
		"tags":         true,
		"groups":       true,
		"groupIds":     true,
		"sshCredential": true,
		"business":     true,
		"cloudInfoData": true,
		"id":           true, // 主键不允许更新
		"createdAt":    true,
	}
	result := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		if !skipKeys[k] {
			result[k] = v
		}
	}
	return result
}

// extractUintSlice 从 interface{} 中提取 []uint（兼容 JSON 反序列化的 []interface{}/[]float64）
func extractUintSlice(val interface{}) []uint {
	switch v := val.(type) {
	case []uint:
		return v
	case []interface{}:
		result := make([]uint, 0, len(v))
		for _, item := range v {
			if f, ok := item.(float64); ok {
				result = append(result, uint(f))
			}
		}
		return result
	}
	return nil
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

	// 性能优化：使用O(n)算法构建树形结构（原算法为O(n²)）
	return s.buildBusinessTreeOptimized(units), nil
}

// buildBusinessTreeOptimized 构建业务树（O(n)时间复杂度）
func (s *CMDBService) buildBusinessTreeOptimized(units []models.BusinessUnit) []models.BusinessUnit {
	// 创建parent_id -> nodes的映射，用于O(1)查找
	parentMap := make(map[uint][]*models.BusinessUnit)
	idMap := make(map[uint]*models.BusinessUnit)

	// 第一遍：建立映射关系
	for i := range units {
		idMap[units[i].ID] = &units[i]
		parentMap[units[i].ParentID] = append(parentMap[units[i].ParentID], &units[i])
	}

	// 第二遍：构建树形结构
	var roots []models.BusinessUnit
	for _, unit := range parentMap[0] {
		*unit = s.buildTreeNode(unit, parentMap)
		roots = append(roots, *unit)
	}

	return roots
}

// buildTreeNode 递归构建树节点（使用parentMap避免重复扫描）
func (s *CMDBService) buildTreeNode(node *models.BusinessUnit, parentMap map[uint][]*models.BusinessUnit) models.BusinessUnit {
	children := parentMap[node.ID]
	if len(children) == 0 {
		return *node
	}

	node.Children = make([]models.BusinessUnit, len(children))
	for i, child := range children {
		node.Children[i] = s.buildTreeNode(child, parentMap)
	}

	return *node
}

// buildBusinessTree 构建业务树（旧版本，保留用于兼容）
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
	// 先通过 IP 或 hostname 查找 Server
	var server models.Server
	tx := db.Preload("SSHCredential")
	if ip != "" {
		tx = tx.Where("ip = ?", ip)
	} else if hostname != "" {
		tx = tx.Where("hostname = ?", hostname)
	}
	if err := tx.First(&server).Error; err != nil {
		return nil, fmt.Errorf("服务器不存在: %v", err)
	}

	// 从 DB 读取最新配置
	var updated models.Server
	if err := db.First(&updated, server.ID).Error; err != nil {
		return nil, fmt.Errorf("读取服务器配置失败: %v", err)
	}

	result := map[string]interface{}{
		"cpu":       updated.CPU,
		"memory":    updated.Memory,
		"disk":      updated.Disk,
		"os":        updated.OS,
		"osVersion": updated.OSVersion,
		"arch":      updated.Arch,
		"hostname":  updated.Hostname,
	}
	return result, nil
}

// ========== 主机分组管理 ==========

// GetServerGroups 获取主机分组列表（树形结构）
func (s *CMDBService) GetServerGroups() ([]models.ServerGroup, error) {
	var groups []models.ServerGroup
	// 加载所有分组
	err := db.Order("sort_order ASC, id ASC").Find(&groups).Error
	if err != nil {
		return nil, err
	}

	// 为每个分组加载直接关联的主机ID（用于计数）
	groupServerMap := make(map[uint][]uint)
	var relations []models.ServerGroupRelation
	db.Find(&relations)

	for _, relation := range relations {
		groupServerMap[relation.GroupID] = append(groupServerMap[relation.GroupID], relation.ServerID)
	}

	// 为每个分组设置 Servers 字段（仅包含ID用于计数）
	for i := range groups {
		serverIDs := groupServerMap[groups[i].ID]
		var servers []models.Server
		if len(serverIDs) > 0 {
			for _, sid := range serverIDs {
				servers = append(servers, models.Server{ID: sid})
			}
		}
		groups[i].Servers = servers
	}

	// 构建树形结构
	return s.buildGroupTree(groups, 0), nil
}

// buildGroupTree 构建分组树
func (s *CMDBService) buildGroupTree(groups []models.ServerGroup, parentID uint) []models.ServerGroup {
	var result []models.ServerGroup
	for _, group := range groups {
		if group.ParentID == parentID {
			// 递归构建子树
			group.Children = s.buildGroupTree(groups, group.ID)
			result = append(result, group)
		}
	}
	return result
}

// GetAssetTree 获取资产树（轻量级数据，不包含敏感信息和监控数据）
func (s *CMDBService) GetAssetTree() (map[string]interface{}, error) {
	// 1. 查询所有分组
	var groups []models.ServerGroup
	err := db.Order("sort_order ASC, id ASC").Find(&groups).Error
	if err != nil {
		return nil, err
	}

	// 2. 查询所有分组关联关系
	var relations []models.ServerGroupRelation
	err = db.Find(&relations).Error
	if err != nil {
		return nil, err
	}

	// 3. 查询所有服务器（只查询必要字段，不包含 credentials 和 metrics）
	type ServerBasicInfo struct {
		ID           uint    `json:"id"`
		Hostname     string  `json:"hostname"`
		IP           string  `json:"ip"`
		AgentStatus  string  `json:"agentStatus"`
		Env          string  `json:"env"`
	}

	var allServers []ServerBasicInfo
	err = db.Model(&models.Server{}).
		Select("id, hostname, ip, agent_status, env").
		Order("hostname ASC").
		Find(&allServers).Error
	if err != nil {
		return nil, err
	}

	// 4. 构建分组ID到服务器的映射
	groupServerMap := make(map[uint][]ServerBasicInfo)
	groupedServerIDs := make(map[uint]bool)

	for _, relation := range relations {
		groupedServerIDs[relation.ServerID] = true
	}

	for _, server := range allServers {
		// 找到这个服务器所属的所有分组
		for _, relation := range relations {
			if relation.ServerID == server.ID {
				groupServerMap[relation.GroupID] = append(groupServerMap[relation.GroupID], server)
				break
			}
		}
	}

	// 5. 构建带服务器的分组树
	type GroupWithServers struct {
		ID          uint                     `json:"id"`
		ParentID    uint                     `json:"parentId"`
		Name        string                   `json:"name"`
		SortOrder   int                      `json:"sortOrder"`
		Children    []GroupWithServers      `json:"children"`
		Servers     []ServerBasicInfo        `json:"servers"`
		ServerCount int                      `json:"serverCount"`
	}

	// sumServerCounts 递归计算服务器的总数
	sumServerCounts := func(groups []GroupWithServers) int {
		count := 0
		for _, group := range groups {
			count += group.ServerCount
		}
		return count
	}

	// 递归构建树
	var buildTree func(parentID uint) []GroupWithServers
	buildTree = func(parentID uint) []GroupWithServers {
		var result []GroupWithServers
		for _, group := range groups {
			if group.ParentID == parentID {
				servers := groupServerMap[group.ID]
				children := buildTree(group.ID)
				serverCount := len(servers) + sumServerCounts(children)

				result = append(result, GroupWithServers{
					ID:          group.ID,
					ParentID:    group.ParentID,
					Name:        group.Name,
					SortOrder:   group.SortOrder,
					Children:    children,
					Servers:     servers,
					ServerCount: serverCount,
				})
			}
		}
		return result
	}

	tree := buildTree(0)

	// 6. 筛选无分组的服务器
	var ungroupedServers []ServerBasicInfo
	for _, server := range allServers {
		if !groupedServerIDs[server.ID] {
			ungroupedServers = append(ungroupedServers, server)
		}
	}

	return map[string]interface{}{
		"groups":           tree,
		"ungroupedServers": ungroupedServers,
	}, nil
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

// AssignServerToGroup 将服务器分配到单个分组（会清除其他分组）
func (s *CMDBService) AssignServerToGroup(serverID, groupID uint) error {
	// 先删除该服务器的所有分组关联
	db.Where("server_id = ?", serverID).Delete(&models.ServerGroupRelation{})

	// 创建新的分组关联
	if err := db.Create(&models.ServerGroupRelation{
		ServerID: serverID,
		GroupID:   groupID,
	}).Error; err != nil {
		return err
	}

	// 维护冗余字段
	return s.updateServerRedundantFields(serverID)
}

// AssignServerToGroups 将服务器分配到多个分组（会清除其他分组）
func (s *CMDBService) AssignServerToGroups(serverID uint, groupIDs []uint) error {
	if len(groupIDs) == 0 {
		// 如果没有分组，删除所有关联
		if err := db.Where("server_id = ?", serverID).Delete(&models.ServerGroupRelation{}).Error; err != nil {
			return err
		}
		return s.updateServerRedundantFields(serverID)
	}

	// 先删除该服务器的所有分组关联
	db.Where("server_id = ?", serverID).Delete(&models.ServerGroupRelation{})

	// 批量创建新的分组关联
	relations := make([]models.ServerGroupRelation, len(groupIDs))
	for i, groupID := range groupIDs {
		relations[i] = models.ServerGroupRelation{
			ServerID: serverID,
			GroupID:  groupID,
		}
	}
	if err := db.Create(&relations).Error; err != nil {
		return err
	}

	// 维护冗余字段
	return s.updateServerRedundantFields(serverID)
}

// RemoveServerFromGroup 将服务器从分组中移除
func (s *CMDBService) RemoveServerFromGroup(serverID, groupID uint) error {
	if err := db.Where("server_id = ? AND group_id = ?", serverID, groupID).Delete(&models.ServerGroupRelation{}).Error; err != nil {
		return err
	}
	// 维护冗余字段
	return s.updateServerRedundantFields(serverID)
}

// GetServersByGroup 获取指定分组下的服务器列表
func (s *CMDBService) GetServersByGroup(groupID uint) ([]models.Server, error) {
	var servers []models.Server
	err := db.Joins("JOIN server_group_relations ON servers.id = server_group_relations.server_id").
		Preload("Credentials").
		Where("server_group_relations.group_id = ?", groupID).
		Find(&servers).Error
	return servers, err
}

// ========== SSH凭证管理 ==========

// GetSSHCredentials 获取SSH凭证列表，credentialType 为空时返回全部
func (s *CMDBService) GetSSHCredentials(credentialType string) ([]models.SSHCredential, error) {
	var credentials []models.SSHCredential
	tx := db.Order("sort_order ASC, id ASC")
	if credentialType == "user" || credentialType == "system" {
		tx = tx.Where("credential_type = ?", credentialType)
	}
	err := tx.Find(&credentials).Error
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
func (s *CMDBService) TestSSHCredential(id uint, testIP string, testPort int) (map[string]interface{}, error) {
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
	result["test_port"] = testPort

	return result, nil
}

// ClearAgentRecord 清空服务器的 Agent 相关字段（不 SSH，仅数据库操作）
func (s *CMDBService) ClearAgentRecord(id uint) error {
	return db.Model(&models.Server{}).Where("id = ?", id).Updates(map[string]interface{}{
		"agent_status":       "uninstalled",
		"agent_version":      "",
		"agent_port":         0,
		"last_heartbeat_at":  nil,
		"cpu_usage":          0,
		"memory_usage":       0,
		"disk_usage":         0,
		"metrics_updated_at": nil,
	}).Error
}
