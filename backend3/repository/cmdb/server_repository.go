package cmdb

import (
	"encoding/json"
	"fmt"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	modelsystem "oneops/backend3/model/system"
	"oneops/backend3/pkg/utils"

	"gorm.io/gorm"
)

// ServerRepository CMDB服务器数据访问层
type ServerRepository struct {
	db *gorm.DB
}

// NewServerRepository 创建服务器仓库
func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

// DB 返回底层 *gorm.DB（供 service 层事务使用）
func (r *ServerRepository) DB() *gorm.DB {
	return r.db
}

// ========== 服务器查询 ==========

// FindServersLight 获取服务器列表（轻量级，仅返回显示字段）
func (r *ServerRepository) FindServersLight(query map[string]interface{}, page, pageSize int) ([]map[string]interface{}, int64, error) {
	var servers []map[string]interface{}
	var total int64

	tx := r.db.Model(&modelcmdb.Server{}).Select("id, hostname, ip, inner_ip, status, provider, agent_status, cpu, memory, os, arch, ssh_port")

	if hostname, ok := query["hostname"].(string); ok && hostname != "" {
		tx = tx.Where("hostname LIKE ?", "%"+utils.SanitizeLikeInput(hostname)+"%")
	}
	if ip, ok := query["ip"].(string); ok && ip != "" {
		tx = tx.Where("ip LIKE ?", "%"+utils.SanitizeLikeInput(ip)+"%")
	}
	if env, ok := query["env"].(string); ok && env != "" {
		tx = tx.Where("id IN (SELECT server_id FROM cmdb_server_attributes WHERE attribute_key = 'env' AND attribute_value = ?)", env)
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

	if groupID, ok := query["groupId"]; ok && groupID != nil {
		groupIDUint := parseUintVal(groupID)
		if groupIDUint > 0 {
			tx = tx.Where("id IN (SELECT server_id FROM cmdb_server_group_relations WHERE group_id = ?)", groupIDUint)
		}
	}

	if ungrouped, ok := query["ungrouped"].(bool); ok && ungrouped {
		tx = tx.Where("id NOT IN (SELECT server_id FROM cmdb_server_group_relations)")
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&servers).Error
	if err != nil {
		return nil, 0, err
	}

	// 从属性表填充属性值到 attributeValues
	if len(servers) > 0 {
		serverIDs := make([]uint, len(servers))
		for i, s := range servers {
			if id, ok := s["id"]; ok {
				switch v := id.(type) {
				case float64:
					serverIDs[i] = uint(v)
				case uint:
					serverIDs[i] = v
				case int:
					serverIDs[i] = uint(v)
				}
			}
		}
		type attrRow struct {
			ServerID       uint   `gorm:"column:server_id"`
			AttributeKey   string `gorm:"column:attribute_key"`
			AttributeValue string `gorm:"column:attribute_value"`
		}
		var rows []attrRow
		r.db.Table("cmdb_server_attributes").
			Select("server_id, attribute_key, attribute_value").
			Where("server_id IN ?", serverIDs).
			Scan(&rows)
		attrMap := make(map[uint]map[string]string)
		for _, row := range rows {
			if attrMap[row.ServerID] == nil {
				attrMap[row.ServerID] = make(map[string]string)
			}
			attrMap[row.ServerID][row.AttributeKey] = row.AttributeValue
		}
		for i, s := range servers {
			if id, ok := s["id"]; ok {
				var sid uint
				switch v := id.(type) {
				case float64:
					sid = uint(v)
				case uint:
					sid = v
				case int:
					sid = uint(v)
				}
				if attrs, ok := attrMap[sid]; ok {
					servers[i]["attributeValues"] = attrs
				}
			}
		}
	}

	return servers, total, err
}

// FindServers 获取服务器列表（完整数据，含关联）
func (r *ServerRepository) FindServers(query map[string]interface{}, page, pageSize int) ([]modelcmdb.Server, int64, error) {
	var servers []modelcmdb.Server

	tx := r.db.Model(&modelcmdb.Server{})

	if hostname, ok := query["hostname"].(string); ok && hostname != "" {
		tx = tx.Where("hostname LIKE ?", "%"+utils.SanitizeLikeInput(hostname)+"%")
	}
	if ip, ok := query["ip"].(string); ok && ip != "" {
		tx = tx.Where("ip LIKE ?", "%"+utils.SanitizeLikeInput(ip)+"%")
	}
	if innerIp, ok := query["innerIp"].(string); ok && innerIp != "" {
		tx = tx.Where("inner_ip LIKE ?", "%"+utils.SanitizeLikeInput(innerIp)+"%")
	}
	if env, ok := query["env"].(string); ok && env != "" {
		tx = tx.Where("id IN (SELECT server_id FROM cmdb_server_attributes WHERE attribute_key = 'env' AND attribute_value = ?)", env)
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

	if groupID, ok := query["groupId"]; ok && groupID != nil {
		groupIDUint := parseUintVal(groupID)
		if groupIDUint > 0 {
			tx = tx.Where("id IN (SELECT server_id FROM cmdb_server_group_relations WHERE group_id = ?)", groupIDUint)
		}
	}

	if ungrouped, ok := query["ungrouped"].(bool); ok && ungrouped {
		tx = tx.Where("id NOT IN (SELECT server_id FROM cmdb_server_group_relations)")
	}

	if businessUnitID, ok := query["businessUnitId"]; ok && businessUnitID != nil {
		businessUnitIDUint := parseUintVal(businessUnitID)
		if businessUnitIDUint > 0 {
			tx = tx.Where("business_id = ?", businessUnitIDUint)
		}
	}

	if tagID, ok := query["tagId"]; ok && tagID != nil {
		tagIDUint := parseUintVal(tagID)
		if tagIDUint > 0 {
			tx = tx.Where("id IN (SELECT server_id FROM cmdb_server_tag_relations WHERE tag_id = ?)", tagIDUint)
		}
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.
		Preload("Tags").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&servers).Error
	if err != nil {
		return nil, 0, err
	}

	// 从属性表填充所有属性值
	r.populateAttributes(servers)

	for i := range servers {
		if servers[i].GroupNames != "[]" && servers[i].GroupNames != "" && servers[i].GroupNames != "null" {
			var groupData []struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal([]byte(servers[i].GroupNames), &groupData); err == nil {
				for _, item := range groupData {
					servers[i].Groups = append(servers[i].Groups, modelcmdb.ServerGroup{
						ID:   item.ID,
						Name: item.Name,
					})
				}
			}
		}

		if servers[i].CredentialNames != "[]" && servers[i].CredentialNames != "" && servers[i].CredentialNames != "null" {
			var credData []struct {
				ID             uint   `json:"id"`
				Name           string `json:"name"`
				CredentialType string `json:"credential_type"`
			}
			if err := json.Unmarshal([]byte(servers[i].CredentialNames), &credData); err == nil {
				for _, item := range credData {
					cred := modelcmdb.SSHCredential{
						ID:   item.ID,
						Name: item.Name,
					}
					if item.CredentialType != "" {
						cred.CredentialType = modelcmdb.CredentialType(item.CredentialType)
					} else {
						cred.CredentialType = modelcmdb.CredentialTypeUser
					}
					servers[i].Credentials = append(servers[i].Credentials, cred)
				}
			}
		}

		if servers[i].SystemCredentialID > 0 {
			var systemCred modelcmdb.SSHCredential
			if err := r.db.Select("id, name, credential_type").First(&systemCred, servers[i].SystemCredentialID).Error; err == nil {
				servers[i].SystemCredential = &systemCred
			}
		}
	}

	return servers, total, err
}

// FindServerByID 根据ID获取服务器（含关联）
func (r *ServerRepository) FindServerByID(id uint) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	err := r.db.
		Preload("Cabinet").
		Preload("Cabinet.Room").
		Preload("Tags").
		Preload("Credentials").
		Preload("Groups").
		Preload("Attributes").
		Preload("SSHCredential").
		First(&server, id).Error
	if err != nil {
		return &server, err
	}
	// 从 Attributes 填充属性值 map
	if len(server.Attributes) > 0 {
		server.AttributeValues = make(map[string]string, len(server.Attributes))
		for _, attr := range server.Attributes {
			server.AttributeValues[attr.AttributeKey] = attr.AttributeValue
		}
	}
	return &server, nil
}

// FindServerForConnect 获取连接所需的服务器信息（轻量级）
func (r *ServerRepository) FindServerForConnect(id uint) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	err := r.db.
		Preload("SSHCredential").
		Preload("Credentials").
		First(&server, id).Error
	if err != nil {
		return &server, err
	}
	// 从属性表填充属性值 map
	r.populateAttributes([]modelcmdb.Server{server})
	return &server, nil
}

// FindServerWithUserCredentials 获取服务器（含用户凭证，用于堡垒连接权限检查）
func (r *ServerRepository) FindServerWithUserCredentials(serverID uint) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	if err := r.db.Preload("Credentials", "credential_type = ?", modelcmdb.CredentialTypeUser).First(&server, serverID).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// FindServerConfig 获取服务器配置信息
func (r *ServerRepository) FindServerConfig(hostname, ip string) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	tx := r.db.Preload("SSHCredential")
	if ip != "" {
		tx = tx.Where("ip = ?", ip)
	} else if hostname != "" {
		tx = tx.Where("hostname = ?", hostname)
	}
	if err := tx.First(&server).Error; err != nil {
		return nil, err
	}

	var updated modelcmdb.Server
	if err := r.db.First(&updated, server.ID).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}

// ========== 服务器增删改 ==========

// CreateServer 创建服务器（含分组和凭证关联）
func (r *ServerRepository) CreateServer(server *modelcmdb.Server) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(server).Error; err != nil {
			return err
		}

		if len(server.GroupIDs) > 0 {
			for _, groupID := range server.GroupIDs {
				relation := modelcmdb.ServerGroupRelation{
					ServerID: server.ID,
					GroupID:  groupID,
				}
				if err := tx.Create(&relation).Error; err != nil {
					return fmt.Errorf("创建分组关联失败: %w", err)
				}
			}
		}

		if len(server.CredentialIDs) > 0 {
			for _, credID := range server.CredentialIDs {
				rel := modelcmdb.ServerCredential{ServerID: server.ID, CredentialID: credID}
				if err := tx.Create(&rel).Error; err != nil {
					return fmt.Errorf("创建凭证关联失败: %w", err)
				}
			}
		} else if server.SSHCredentialID != 0 {
			rel := modelcmdb.ServerCredential{ServerID: server.ID, CredentialID: server.SSHCredentialID}
			tx.Create(&rel)
		} else if server.CredentialID != 0 {
			rel := modelcmdb.ServerCredential{ServerID: server.ID, CredentialID: server.CredentialID}
			tx.Create(&rel)
		}

		return nil
	})
}

// UpdateServerRedundantFields 更新服务器的冗余字段
func (r *ServerRepository) UpdateServerRedundantFields(serverID uint) error {
	var groupData []map[string]interface{}
	r.db.Table("cmdb_server_group_relations").
		Select("g.id, g.name").
		Joins("JOIN cmdb_server_groups g ON g.id = cmdb_server_group_relations.group_id").
		Where("cmdb_server_group_relations.server_id = ?", serverID).
		Scan(&groupData)

	var credData []map[string]interface{}
	r.db.Table("cmdb_server_credentials").
		Select("c.id, c.name, c.credential_type").
		Joins("JOIN cmdb_ssh_credentials c ON c.id = cmdb_server_credentials.credential_id").
		Where("cmdb_server_credentials.server_id = ?", serverID).
		Scan(&credData)

	groupJSON, _ := json.Marshal(groupData)
	credJSON, _ := json.Marshal(credData)

	return r.db.Model(&modelcmdb.Server{}).
		Where("id = ?", serverID).
		Updates(map[string]interface{}{
			"group_names":      string(groupJSON),
			"credential_names": string(credJSON),
		}).Error
}

// UpdateServer 更新服务器（含分组和凭证关联）
func (r *ServerRepository) UpdateServer(id uint, columnUpdates map[string]interface{}, groupIDs []uint, hasGroupUpdate bool, credentialIDs []uint, hasCredUpdate bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&modelcmdb.Server{}).Where("id = ?", id).Updates(columnUpdates).Error; err != nil {
			return err
		}

		if hasGroupUpdate {
			if err := tx.Where("server_id = ?", id).Delete(&modelcmdb.ServerGroupRelation{}).Error; err != nil {
				return fmt.Errorf("删除旧分组关联失败: %w", err)
			}
			for _, groupID := range groupIDs {
				relation := modelcmdb.ServerGroupRelation{
					ServerID: id,
					GroupID:  groupID,
				}
				if err := tx.Create(&relation).Error; err != nil {
					return fmt.Errorf("创建分组关联失败: %w", err)
				}
			}
		}

		if hasCredUpdate {
			if err := tx.Where("server_id = ?", id).Delete(&modelcmdb.ServerCredential{}).Error; err != nil {
				return fmt.Errorf("删除旧凭证关联失败: %w", err)
			}
			for _, credID := range credentialIDs {
				rel := modelcmdb.ServerCredential{ServerID: id, CredentialID: credID}
				if err := tx.Create(&rel).Error; err != nil {
					return fmt.Errorf("创建凭证关联失败: %w", err)
				}
			}
			if len(credentialIDs) > 0 {
				tx.Model(&modelcmdb.Server{}).Where("id = ?", id).Update("ssh_credential_id", credentialIDs[0])
			}
		}

		return nil
	})
}

// FindServerForUpdate 根据ID获取服务器（用于更新前获取旧值）
func (r *ServerRepository) FindServerForUpdate(id uint) (*modelcmdb.Server, error) {
	var server modelcmdb.Server
	if err := r.db.First(&server, id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

// DeleteServer 删除服务器
func (r *ServerRepository) DeleteServer(id uint) error {
	var server modelcmdb.Server
	if err := r.db.First(&server, id).Error; err != nil {
		return err
	}
	return r.db.Delete(&server).Error
}

// UpdateServerFields 更新服务器指定字段
func (r *ServerRepository) UpdateServerFields(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.Server{}).Where("id = ?", id).Updates(updates).Error
}

// ClearAgentRecord 清空服务器的 Agent 相关字段
func (r *ServerRepository) ClearAgentRecord(id uint) error {
	return r.db.Model(&modelcmdb.Server{}).Where("id = ?", id).Updates(map[string]interface{}{
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

// ========== 服务器统计 ==========

// GetServerStats 获取服务器统计信息
func (r *ServerRepository) GetServerStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var total int64
	r.db.Model(&modelcmdb.Server{}).Count(&total)
	stats["total"] = total

	var envStats []struct {
		Env   string
		Count int64
	}
	r.db.Table("cmdb_server_attributes").
		Select("attribute_value as env, count(*) as count").
		Where("attribute_key = 'env'").
		Group("attribute_value").
		Scan(&envStats)

	// 获取 env 属性定义的 options，解析 label
	var envDef modelsystem.AttributeDefinition
	if err := r.db.Where("key = 'env'").First(&envDef).Error; err == nil && envDef.Options != "" {
		var options []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		}
		if json.Unmarshal([]byte(envDef.Options), &options) == nil {
			envLabelMap := make(map[string]string)
			for _, opt := range options {
				envLabelMap[opt.Value] = opt.Label
			}
			envList := make([]map[string]interface{}, 0, len(envStats))
			for _, stat := range envStats {
				label := stat.Env
				if l, ok := envLabelMap[stat.Env]; ok {
					label = l
				}
				envList = append(envList, map[string]interface{}{
					"value": stat.Env,
					"label": label,
					"count": stat.Count,
				})
			}
			stats["byEnv"] = envList
		} else {
			envMap := make(map[string]int64)
			for _, stat := range envStats {
				envMap[stat.Env] = stat.Count
			}
			stats["byEnv"] = envMap
		}
	} else {
		envMap := make(map[string]int64)
		for _, stat := range envStats {
			envMap[stat.Env] = stat.Count
		}
		stats["byEnv"] = envMap
	}

	var statusStats []struct {
		Status string
		Count  int64
	}
	r.db.Model(&modelcmdb.Server{}).Select("status, count(*) as count").Group("status").Scan(&statusStats)
	statusMap := make(map[string]int64)
	for _, stat := range statusStats {
		statusMap[stat.Status] = stat.Count
	}
	stats["byStatus"] = statusMap

	var providerStats []struct {
		Provider string
		Count    int64
	}
	r.db.Model(&modelcmdb.Server{}).Select("provider, count(*) as count").Group("provider").Scan(&providerStats)
	providerMap := make(map[string]int64)
	for _, stat := range providerStats {
		providerMap[stat.Provider] = stat.Count
	}
	stats["byProvider"] = providerMap

	return stats, nil
}

// ========== 资产树 ==========

// FindAllServerGroups 查询所有主机分组
func (r *ServerRepository) FindAllServerGroups() ([]modelcmdb.ServerGroup, error) {
	var groups []modelcmdb.ServerGroup
	err := r.db.Order("sort_order ASC, id ASC").Find(&groups).Error
	return groups, err
}

// FindAllGroupRelations 查询所有分组关联
func (r *ServerRepository) FindAllGroupRelations() ([]modelcmdb.ServerGroupRelation, error) {
	var relations []modelcmdb.ServerGroupRelation
	err := r.db.Find(&relations).Error
	return relations, err
}

// CountUngroupedServers 统计未分组主机数量
func (r *ServerRepository) CountUngroupedServers() (int64, error) {
	var count int64
	err := r.db.Model(&modelcmdb.Server{}).
		Where("id NOT IN (SELECT server_id FROM cmdb_server_group_relations)").
		Count(&count).Error
	return count, err
}

// CountServersByGroup 批量统计每个分组直接关联的主机数量
func (r *ServerRepository) CountServersByGroup() (map[uint]int64, error) {
	type result struct {
		GroupID uint  `json:"group_id"`
		Count   int64 `json:"count"`
	}
	var results []result
	err := r.db.Table("cmdb_server_group_relations").
		Select("group_id, COUNT(*) as count").
		Group("group_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint]int64, len(results))
	for _, r := range results {
		m[r.GroupID] = r.Count
	}
	return m, nil
}

// FindAllServersBasic 查询所有服务器基本信息（用于资产树）
func (r *ServerRepository) FindAllServersBasic() ([]struct {
	ID          uint   `json:"id"`
	Hostname    string `json:"hostname"`
	IP          string `json:"ip"`
	AgentStatus string `json:"agentStatus"`
	Env         string `json:"env"`
}, error) {
	var servers []struct {
		ID          uint   `json:"id"`
		Hostname    string `json:"hostname"`
		IP          string `json:"ip"`
		AgentStatus string `json:"agentStatus"`
		Env         string `json:"env"`
	}
	err := r.db.Model(&modelcmdb.Server{}).
		Select("id, hostname, ip, agent_status").
		Order("hostname ASC").
		Find(&servers).Error
	if err != nil {
		return servers, err
	}

	// 从属性表填充 env
	if len(servers) > 0 {
		serverIDs := make([]uint, len(servers))
		for i := range servers {
			serverIDs[i] = servers[i].ID
		}
		type envAttr struct {
			ServerID       uint   `gorm:"column:server_id"`
			AttributeValue string `gorm:"column:attribute_value"`
		}
		var envs []envAttr
		r.db.Table("cmdb_server_attributes").
			Select("server_id, attribute_value").
			Where("server_id IN ? AND attribute_key = 'env'", serverIDs).
			Scan(&envs)
		envMap := make(map[uint]string)
		for _, e := range envs {
			envMap[e.ServerID] = e.AttributeValue
		}
		for i := range servers {
			if env, ok := envMap[servers[i].ID]; ok {
				servers[i].Env = env
			}
		}
	}
	return servers, err
}

// populateAttributes 从 cmdb_server_attributes 表批量填充 Server.AttributeValues
func (r *ServerRepository) populateAttributes(servers []modelcmdb.Server) {
	if len(servers) == 0 {
		return
	}
	serverIDs := make([]uint, len(servers))
	for i := range servers {
		serverIDs[i] = servers[i].ID
	}

	type attrRow struct {
		ServerID       uint   `gorm:"column:server_id"`
		AttributeKey   string `gorm:"column:attribute_key"`
		AttributeValue string `gorm:"column:attribute_value"`
	}
	var rows []attrRow
	r.db.Table("cmdb_server_attributes").
		Select("server_id, attribute_key, attribute_value").
		Where("server_id IN ?", serverIDs).
		Scan(&rows)

	// 按 serverID 分组
	attrMap := make(map[uint]map[string]string)
	for _, row := range rows {
		if attrMap[row.ServerID] == nil {
			attrMap[row.ServerID] = make(map[string]string)
		}
		attrMap[row.ServerID][row.AttributeKey] = row.AttributeValue
	}
	for i := range servers {
		if attrs, ok := attrMap[servers[i].ID]; ok {
			servers[i].AttributeValues = attrs
		}
	}
}

// ========== 业务系统管理 ==========

// FindBusinessUnits 获取业务系统列表
func (r *ServerRepository) FindBusinessUnits() ([]modelcmdb.BusinessUnit, error) {
	var units []modelcmdb.BusinessUnit
	err := r.db.Order("sort_order ASC, id ASC").Find(&units).Error
	return units, err
}

// FindBusinessUnitByID 根据ID获取业务系统
func (r *ServerRepository) FindBusinessUnitByID(id uint) (*modelcmdb.BusinessUnit, error) {
	var unit modelcmdb.BusinessUnit
	err := r.db.First(&unit, id).Error
	return &unit, err
}

// CreateBusinessUnit 创建业务系统
func (r *ServerRepository) CreateBusinessUnit(unit *modelcmdb.BusinessUnit) error {
	return r.db.Create(unit).Error
}

// UpdateBusinessUnit 更新业务系统
func (r *ServerRepository) UpdateBusinessUnit(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.BusinessUnit{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteBusinessUnit 删除业务系统
func (r *ServerRepository) DeleteBusinessUnit(id uint) error {
	return r.db.Delete(&modelcmdb.BusinessUnit{}, id).Error
}

// CountBusinessUnitChildren 统计子业务数量
func (r *ServerRepository) CountBusinessUnitChildren(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&modelcmdb.BusinessUnit{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

// ========== 机房机柜管理 ==========

// FindServerRooms 获取机房列表（含机柜）
func (r *ServerRepository) FindServerRooms() ([]modelcmdb.ServerRoom, error) {
	var rooms []modelcmdb.ServerRoom
	err := r.db.Preload("Cabinets").Order("id ASC").Find(&rooms).Error
	return rooms, err
}

// CreateServerRoom 创建机房
func (r *ServerRepository) CreateServerRoom(room *modelcmdb.ServerRoom) error {
	return r.db.Create(room).Error
}

// UpdateServerRoom 更新机房
func (r *ServerRepository) UpdateServerRoom(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.ServerRoom{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServerRoom 删除机房
func (r *ServerRepository) DeleteServerRoom(id uint) error {
	return r.db.Delete(&modelcmdb.ServerRoom{}, id).Error
}

// FindCabinets 获取机柜列表
func (r *ServerRepository) FindCabinets(roomID uint) ([]modelcmdb.Cabinet, error) {
	var cabinets []modelcmdb.Cabinet
	tx := r.db.Preload("Room")
	if roomID > 0 {
		tx = tx.Where("room_id = ?", roomID)
	}
	err := tx.Order("id ASC").Find(&cabinets).Error
	return cabinets, err
}

// ========== 标签管理 ==========

// FindServerTags 获取服务器标签列表
func (r *ServerRepository) FindServerTags() ([]modelcmdb.ServerTag, error) {
	var tags []modelcmdb.ServerTag
	err := r.db.Order("sort_order ASC").Find(&tags).Error
	return tags, err
}

// CreateServerTag 创建服务器标签
func (r *ServerRepository) CreateServerTag(tag *modelcmdb.ServerTag) error {
	return r.db.Create(tag).Error
}

// UpdateServerTag 更新服务器标签
func (r *ServerRepository) UpdateServerTag(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.ServerTag{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServerTag 删除服务器标签
func (r *ServerRepository) DeleteServerTag(id uint) error {
	return r.db.Delete(&modelcmdb.ServerTag{}, id).Error
}

// AssignServerTag 为服务器分配标签
func (r *ServerRepository) AssignServerTag(serverID, tagID uint) error {
	var count int64
	r.db.Model(&modelcmdb.ServerTagRelation{}).Where("server_id = ? AND tag_id = ?", serverID, tagID).Count(&count)
	if count > 0 {
		return nil
	}
	relation := &modelcmdb.ServerTagRelation{
		ServerID: serverID,
		TagID:    tagID,
	}
	return r.db.Create(relation).Error
}

// RemoveServerTag 移除服务器标签
func (r *ServerRepository) RemoveServerTag(serverID, tagID uint) error {
	return r.db.Where("server_id = ? AND tag_id = ?", serverID, tagID).Delete(&modelcmdb.ServerTagRelation{}).Error
}

// ========== 主机分组管理 ==========

// FindServerGroupByID 根据ID获取主机分组
func (r *ServerRepository) FindServerGroupByID(id uint) (*modelcmdb.ServerGroup, error) {
	var group modelcmdb.ServerGroup
	err := r.db.Preload("Servers").Preload("Parent").First(&group, id).Error
	return &group, err
}

// CreateServerGroup 创建主机分组
func (r *ServerRepository) CreateServerGroup(group *modelcmdb.ServerGroup) error {
	return r.db.Create(group).Error
}

// UpdateServerGroup 更新主机分组
func (r *ServerRepository) UpdateServerGroup(id uint, updates map[string]interface{}) error {
	return r.db.Model(&modelcmdb.ServerGroup{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteServerGroup 删除主机分组
func (r *ServerRepository) DeleteServerGroup(id uint) error {
	return r.db.Delete(&modelcmdb.ServerGroup{}, id).Error
}

// AssignServerToGroup 将服务器分配到单个分组
func (r *ServerRepository) AssignServerToGroup(serverID, groupID uint) error {
	r.db.Where("server_id = ?", serverID).Delete(&modelcmdb.ServerGroupRelation{})
	return r.db.Create(&modelcmdb.ServerGroupRelation{
		ServerID: serverID,
		GroupID:  groupID,
	}).Error
}

// AssignServerToGroups 将服务器分配到多个分组
func (r *ServerRepository) AssignServerToGroups(serverID uint, groupIDs []uint) error {
	if len(groupIDs) == 0 {
		if err := r.db.Where("server_id = ?", serverID).Delete(&modelcmdb.ServerGroupRelation{}).Error; err != nil {
			return err
		}
		return nil
	}

	r.db.Where("server_id = ?", serverID).Delete(&modelcmdb.ServerGroupRelation{})

	relations := make([]modelcmdb.ServerGroupRelation, len(groupIDs))
	for i, groupID := range groupIDs {
		relations[i] = modelcmdb.ServerGroupRelation{
			ServerID: serverID,
			GroupID:  groupID,
		}
	}
	return r.db.Create(&relations).Error
}

// RemoveServerFromGroup 将服务器从分组中移除
func (r *ServerRepository) RemoveServerFromGroup(serverID, groupID uint) error {
	return r.db.Where("server_id = ? AND group_id = ?", serverID, groupID).Delete(&modelcmdb.ServerGroupRelation{}).Error
}

// FindServersByGroup 获取指定分组下的服务器列表
func (r *ServerRepository) FindServersByGroup(groupID uint) ([]modelcmdb.Server, error) {
	var servers []modelcmdb.Server
	err := r.db.Joins("JOIN cmdb_server_group_relations ON cmdb_servers.id = cmdb_server_group_relations.server_id").
		Preload("Credentials").
		Where("cmdb_server_group_relations.group_id = ?", groupID).
		Find(&servers).Error
	return servers, err
}

// ========== 资产变更记录 ==========
//
// 注意：SSH凭证管理（FindSSHCredentials/FindSSHCredentialByID/CreateSSHCredential/
// UpdateSSHCredential/DeleteSSHCredential）已迁移至 CredentialRepository（credential_repository.go）。

// FindAssetChanges 获取资产变更记录
func (r *ServerRepository) FindAssetChanges(assetType string, assetID uint, page, pageSize int) ([]modelcmdb.AssetChange, int64, error) {
	var changes []modelcmdb.AssetChange
	var total int64

	tx := r.db.Model(&modelcmdb.AssetChange{})
	if assetType != "" {
		tx = tx.Where("asset_type = ?", assetType)
	}
	if assetID > 0 {
		tx = tx.Where("asset_id = ?", assetID)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Order("operate_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&changes).Error

	return changes, total, err
}

// CreateAssetChange 记录资产变更
func (r *ServerRepository) CreateAssetChange(change *modelcmdb.AssetChange) error {
	return r.db.Create(change).Error
}

// ========== 辅助函数 ==========

func parseUintVal(value interface{}) uint {
	switch v := value.(type) {
	case uint:
		return v
	case uint64:
		return uint(v)
	case int:
		return uint(v)
	case int64:
		return uint(v)
	case float64:
		return uint(v)
	case float32:
		return uint(v)
	case string:
		if parsedVal, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(parsedVal)
		}
	}
	return 0
}
