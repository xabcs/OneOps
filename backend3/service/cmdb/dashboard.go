package cmdb

import (
	"fmt"
)

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
