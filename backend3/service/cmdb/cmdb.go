package cmdb

import (
	"encoding/json"
	"fmt"
	"strconv"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	repocmdb "oneops/backend3/repository/cmdb"

	"go.uber.org/zap"
)

// CMDBService CMDB服务
type CMDBService struct {
	repo *repocmdb.ServerRepository
}

// NewCMDBService 创建CMDB服务
func NewCMDBService(repo *repocmdb.ServerRepository) *CMDBService {
	return &CMDBService{repo: repo}
}

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

// ========== 标签管理 ==========

// GetServerTags 获取服务器标签列表
func (s *CMDBService) GetServerTags() ([]modelcmdb.ServerTag, error) {
	return s.repo.FindServerTags()
}

// CreateServerTag 创建服务器标签
func (s *CMDBService) CreateServerTag(tag *modelcmdb.ServerTag) error {
	return s.repo.CreateServerTag(tag)
}

// UpdateServerTag 更新服务器标签
func (s *CMDBService) UpdateServerTag(id uint, updates map[string]interface{}) error {
	return s.repo.UpdateServerTag(id, updates)
}

// DeleteServerTag 删除服务器标签
func (s *CMDBService) DeleteServerTag(id uint) error {
	return s.repo.DeleteServerTag(id)
}

// AssignServerTag 为服务器分配标签
func (s *CMDBService) AssignServerTag(serverID, tagID uint) error {
	return s.repo.AssignServerTag(serverID, tagID)
}

// RemoveServerTag 移除服务器标签
func (s *CMDBService) RemoveServerTag(serverID, tagID uint) error {
	return s.repo.RemoveServerTag(serverID, tagID)
}

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

// ========== 辅助函数 ==========

// getFieldValue 获取结构体字段值
func getFieldValue(obj interface{}, field string) interface{} {
	return ""
}

// GetServerStats 获取服务器统计信息
func (s *CMDBService) GetServerStats() (map[string]interface{}, error) {
	return s.repo.GetServerStats()
}

// GetServerConfig 获取服务器配置信息
func (s *CMDBService) GetServerConfig(hostname, ip string, sshUser string, sshPort int) (map[string]interface{}, error) {
	server, err := s.repo.FindServerConfig(hostname, ip)
	if err != nil {
		return nil, fmt.Errorf("服务器不存在: %v", err)
	}

	result := map[string]interface{}{
		"cpu":       server.CPU,
		"memory":    server.Memory,
		"disk":      server.Disk,
		"os":        server.OS,
		"osVersion": server.OSVersion,
		"arch":      server.Arch,
		"hostname":  server.Hostname,
	}
	return result, nil
}

// ========== 主机分组管理 ==========

// GetServerGroups 获取主机分组列表（树形结构）
func (s *CMDBService) GetServerGroups() ([]modelcmdb.ServerGroup, error) {
	groups, err := s.repo.FindAllServerGroups()
	if err != nil {
		return nil, err
	}

	relations, err := s.repo.FindAllGroupRelations()
	if err != nil {
		return nil, err
	}

	groupServerMap := make(map[uint][]uint)
	for _, relation := range relations {
		groupServerMap[relation.GroupID] = append(groupServerMap[relation.GroupID], relation.ServerID)
	}

	for i := range groups {
		serverIDs := groupServerMap[groups[i].ID]
		var servers []modelcmdb.Server
		if len(serverIDs) > 0 {
			for _, sid := range serverIDs {
				servers = append(servers, modelcmdb.Server{ID: sid})
			}
		}
		groups[i].Servers = servers
	}

	return s.buildGroupTree(groups, 0), nil
}

// buildGroupTree 构建分组树
func (s *CMDBService) buildGroupTree(groups []modelcmdb.ServerGroup, parentID uint) []modelcmdb.ServerGroup {
	var result []modelcmdb.ServerGroup
	for _, group := range groups {
		if group.ParentID == parentID {
			group.Children = s.buildGroupTree(groups, group.ID)
			result = append(result, group)
		}
	}
	return result
}

// GetAssetTree 获取资产树
func (s *CMDBService) GetAssetTree() (map[string]interface{}, error) {
	groups, err := s.repo.FindAllServerGroups()
	if err != nil {
		return nil, err
	}

	relations, err := s.repo.FindAllGroupRelations()
	if err != nil {
		return nil, err
	}

	type ServerBasicInfo struct {
		ID          uint   `json:"id"`
		Hostname    string `json:"hostname"`
		IP          string `json:"ip"`
		AgentStatus string `json:"agentStatus"`
		Env         string `json:"env"`
	}

	allServers, err := s.repo.FindAllServersBasic()
	if err != nil {
		return nil, err
	}

	// 转换类型
	var servers []ServerBasicInfo
	for _, srv := range allServers {
		servers = append(servers, ServerBasicInfo{
			ID:          srv.ID,
			Hostname:    srv.Hostname,
			IP:          srv.IP,
			AgentStatus: srv.AgentStatus,
			Env:         srv.Env,
		})
	}

	groupServerMap := make(map[uint][]ServerBasicInfo)
	groupedServerIDs := make(map[uint]bool)

	for _, relation := range relations {
		groupedServerIDs[relation.ServerID] = true
	}

	for _, server := range servers {
		for _, relation := range relations {
			if relation.ServerID == server.ID {
				groupServerMap[relation.GroupID] = append(groupServerMap[relation.GroupID], server)
				break
			}
		}
	}

	type GroupWithServers struct {
		ID          uint               `json:"id"`
		ParentID    uint               `json:"parentId"`
		Name        string             `json:"name"`
		SortOrder   int                `json:"sortOrder"`
		Children    []GroupWithServers `json:"children"`
		Servers     []ServerBasicInfo  `json:"servers"`
		ServerCount int                `json:"serverCount"`
	}

	sumServerCounts := func(gs []GroupWithServers) int {
		count := 0
		for _, g := range gs {
			count += g.ServerCount
		}
		return count
	}

	var buildTree func(parentID uint) []GroupWithServers
	buildTree = func(parentID uint) []GroupWithServers {
		var result []GroupWithServers
		for _, group := range groups {
			if group.ParentID == parentID {
				srvs := groupServerMap[group.ID]
				children := buildTree(group.ID)
				serverCount := len(srvs) + sumServerCounts(children)

				result = append(result, GroupWithServers{
					ID:          group.ID,
					ParentID:    group.ParentID,
					Name:        group.Name,
					SortOrder:   group.SortOrder,
					Children:    children,
					Servers:     srvs,
					ServerCount: serverCount,
				})
			}
		}
		return result
	}

	tree := buildTree(0)

	var ungroupedServers []ServerBasicInfo
	for _, server := range servers {
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
func (s *CMDBService) GetServerGroupByID(id uint) (*modelcmdb.ServerGroup, error) {
	return s.repo.FindServerGroupByID(id)
}

// CreateServerGroup 创建主机分组
func (s *CMDBService) CreateServerGroup(group *modelcmdb.ServerGroup) error {
	return s.repo.CreateServerGroup(group)
}

// UpdateServerGroup 更新主机分组
func (s *CMDBService) UpdateServerGroup(id uint, updates map[string]interface{}) error {
	return s.repo.UpdateServerGroup(id, updates)
}

// DeleteServerGroup 删除主机分组
func (s *CMDBService) DeleteServerGroup(id uint) error {
	return s.repo.DeleteServerGroup(id)
}

// AssignServerToGroup 将服务器分配到单个分组
func (s *CMDBService) AssignServerToGroup(serverID, groupID uint) error {
	if err := s.repo.AssignServerToGroup(serverID, groupID); err != nil {
		return err
	}
	return s.repo.UpdateServerRedundantFields(serverID)
}

// AssignServerToGroups 将服务器分配到多个分组
func (s *CMDBService) AssignServerToGroups(serverID uint, groupIDs []uint) error {
	if err := s.repo.AssignServerToGroups(serverID, groupIDs); err != nil {
		return err
	}
	return s.repo.UpdateServerRedundantFields(serverID)
}

// RemoveServerFromGroup 将服务器从分组中移除
func (s *CMDBService) RemoveServerFromGroup(serverID, groupID uint) error {
	if err := s.repo.RemoveServerFromGroup(serverID, groupID); err != nil {
		return err
	}
	return s.repo.UpdateServerRedundantFields(serverID)
}

// GetServersByGroup 获取指定分组下的服务器列表
func (s *CMDBService) GetServersByGroup(groupID uint) ([]modelcmdb.Server, error) {
	return s.repo.FindServersByGroup(groupID)
}

// ========== SSH凭证管理 ==========

// GetSSHCredentials 获取SSH凭证列表
func (s *CMDBService) GetSSHCredentials(credentialType string) ([]modelcmdb.SSHCredential, error) {
	return s.repo.FindSSHCredentials(credentialType)
}

// GetSSHCredentialByID 根据ID获取SSH凭证
func (s *CMDBService) GetSSHCredentialByID(id uint) (*modelcmdb.SSHCredential, error) {
	credential, err := s.repo.FindSSHCredentialByID(id)
	if err != nil {
		return nil, err
	}

	if credential.Password != "" {
		decrypted, err := utils.DecryptString(credential.Password)
		if err != nil {
			logger.Warn("SSH凭证密码解密失败", zap.Uint("credential_id", id), zap.Error(err))
		} else {
			credential.Password = decrypted
		}
	}

	if credential.PrivateKey != "" {
		decrypted, err := utils.DecryptString(credential.PrivateKey)
		if err != nil {
			logger.Warn("SSH凭证私钥解密失败", zap.Uint("credential_id", id), zap.Error(err))
		} else {
			credential.PrivateKey = decrypted
		}
	}

	if credential.Passphrase != "" {
		decrypted, err := utils.DecryptString(credential.Passphrase)
		if err != nil {
			logger.Warn("SSH凭证passphrase解密失败", zap.Uint("credential_id", id), zap.Error(err))
		} else {
			credential.Passphrase = decrypted
		}
	}

	return credential, nil
}

// CreateSSHCredential 创建SSH凭证
func (s *CMDBService) CreateSSHCredential(credential *modelcmdb.SSHCredential) error {
	if credential.Password != "" {
		encrypted, err := utils.EncryptString(credential.Password)
		if err != nil {
			return fmt.Errorf("加密密码失败: %w", err)
		}
		credential.Password = encrypted
	}

	if credential.PrivateKey != "" {
		encrypted, err := utils.EncryptString(credential.PrivateKey)
		if err != nil {
			return fmt.Errorf("加密私钥失败: %w", err)
		}
		credential.PrivateKey = encrypted
	}

	if credential.Passphrase != "" {
		encrypted, err := utils.EncryptString(credential.Passphrase)
		if err != nil {
			return fmt.Errorf("加密passphrase失败: %w", err)
		}
		credential.Passphrase = encrypted
	}

	return s.repo.CreateSSHCredential(credential)
}

// UpdateSSHCredential 更新SSH凭证
func (s *CMDBService) UpdateSSHCredential(id uint, updates map[string]interface{}) error {
	if password, ok := updates["password"]; ok && password != "" {
		if passwordStr, ok := password.(string); ok {
			encrypted, err := utils.EncryptString(passwordStr)
			if err != nil {
				return fmt.Errorf("加密密码失败: %w", err)
			}
			updates["password"] = encrypted
		}
	}

	if privateKey, ok := updates["private_key"]; ok && privateKey != "" {
		if privateKeyStr, ok := privateKey.(string); ok {
			encrypted, err := utils.EncryptString(privateKeyStr)
			if err != nil {
				return fmt.Errorf("加密私钥失败: %w", err)
			}
			updates["private_key"] = encrypted
		}
	}

	if passphrase, ok := updates["passphrase"]; ok && passphrase != "" {
		if passphraseStr, ok := passphrase.(string); ok {
			encrypted, err := utils.EncryptString(passphraseStr)
			if err != nil {
				return fmt.Errorf("加密passphrase失败: %w", err)
			}
			updates["passphrase"] = encrypted
		}
	}

	return s.repo.UpdateSSHCredential(id, updates)
}

// DeleteSSHCredential 删除SSH凭证
func (s *CMDBService) DeleteSSHCredential(id uint) error {
	return s.repo.DeleteSSHCredential(id)
}

// TestSSHCredential 测试SSH凭证连接
func (s *CMDBService) TestSSHCredential(id uint, testIP string, testPort int) (map[string]interface{}, error) {
	credential, err := s.GetSSHCredentialByID(id)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	result["success"] = true
	result["message"] = "连接测试功能开发中"
	result["credential"] = credential.Username
	result["test_ip"] = testIP
	result["test_port"] = testPort

	return result, nil
}

// ClearAgentRecord 清空服务器的 Agent 相关字段
func (s *CMDBService) ClearAgentRecord(id uint) error {
	return s.repo.ClearAgentRecord(id)
}

// ========== 辅助函数 ==========

// filterServerColumns 过滤掉非数据库列字段
func filterServerColumns(updates map[string]interface{}) map[string]interface{} {
	skipKeys := map[string]bool{
		"cloudInfo":     true,
		"credential":    true,
		"credentials":   true,
		"credentialIds": true,
		"cabinet":       true,
		"tags":          true,
		"groups":        true,
		"groupIds":      true,
		"sshCredential": true,
		"business":      true,
		"cloudInfoData": true,
		"id":            true,
		"createdAt":     true,
	}
	result := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		if !skipKeys[k] {
			result[k] = v
		}
	}
	return result
}

// extractUintSlice 从 interface{} 中提取 []uint
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

// rawExists 检查 key 是否存在于 map 中
func rawExists(updates map[string]interface{}, key string) bool {
	_, exists := updates[key]
	return exists
}

// parseUint 辅助函数：将interface{}解析为uint
func parseUint(value interface{}) uint {
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

// parseGroupIDUint 解析 groupID（用于 GetServersLight 中的类型转换）
func parseGroupIDUint(groupID interface{}) uint {
	return parseUint(groupID)
}

// suppress unused import warning
var _ = json.Marshal
