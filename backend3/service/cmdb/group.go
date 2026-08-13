package cmdb

import (
	modelcmdb "oneops/backend3/model/cmdb"
)

// ========== 主机分组管理 ==========

// GetServerGroups 获取主机分组列表（树形结构）
func (s *CMDBService) GetServerGroups() ([]modelcmdb.ServerGroup, int64, error) {
	groups, err := s.repo.FindAllServerGroups()
	if err != nil {
		return nil, 0, err
	}

	// 批量统计每个分组直接关联的主机数量
	countMap, err := s.repo.CountServersByGroup()
	if err != nil {
		countMap = make(map[uint]int64)
	}
	for i := range groups {
		groups[i].ServerCount = int(countMap[groups[i].ID])
	}

	ungroupedCount, err := s.repo.CountUngroupedServers()
	if err != nil {
		ungroupedCount = 0
	}

	return s.buildGroupTree(groups, 0), ungroupedCount, nil
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
