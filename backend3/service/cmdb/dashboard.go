package cmdb

import (
	"fmt"
	"reflect"
	"strings"
)

// ========== 辅助函数 ==========

// getFieldValue 获取结构体字段值：优先按 JSON tag 名匹配（与更新请求体的字段名一致），
// 未配置 JSON tag 时按 Go 字段名匹配；未命中时返回 nil
func getFieldValue(obj interface{}, field string) interface{} {
	if obj == nil {
		return nil
	}
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if tag := t.Field(i).Tag.Get("json"); tag != "" {
			if idx := strings.Index(tag, ","); idx >= 0 {
				tag = tag[:idx]
			}
			if tag != "" {
				name = tag
			}
		}
		if name == field || t.Field(i).Name == field {
			return v.Field(i).Interface()
		}
	}
	return nil
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
