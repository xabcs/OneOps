package monitoring

import (
	"encoding/json"
	"fmt"
	"time"

	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
)

// GetServerExtendedMetrics 获取主机扩展指标
func (s *MonitoringService) GetServerExtendedMetrics(serverID uint) (*AgentExtendedMetricsResponse, error) {
	cache := database.NewRedisCache()
	var cachedMetrics AgentExtendedMetricsResponse
	if err := cache.GetServerMetrics(serverID, &cachedMetrics); err == nil {
		logger.Debug("从 Redis 缓存获取主机指标", zap.Uint("serverID", serverID))
		return &cachedMetrics, nil
	}

	var metricData struct {
		MetricData string    `json:"metric_data"`
		ReceivedAt time.Time `json:"received_at"`
	}

	err := getDB().Table("mon_agent_metrics").
		Select("metric_data, received_at").
		Where("server_id = ? AND metric_type = 'extended' AND received_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err == nil && metricData.MetricData != "" {
		var metrics AgentExtendedMetricsResponse
		if err := json.Unmarshal([]byte(metricData.MetricData), &metrics); err == nil {
			cache.CacheServerMetrics(serverID, metrics)
			return &metrics, nil
		}
	}

	return s.PullExtendedMetrics(serverID)
}

// GetServerProcesses 获取主机进程信息
func (s *MonitoringService) GetServerProcesses(serverID uint) (*ProcessInfo, error) {
	var metricData struct {
		MetricData string `json:"metric_data"`
	}

	err := getDB().Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'process' AND received_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		metrics, err := s.PullExtendedMetrics(serverID)
		if err != nil {
			return nil, err
		}
		return &metrics.ProcessInfo, nil
	}

	var processInfo ProcessInfo
	if err := json.Unmarshal([]byte(metricData.MetricData), &processInfo); err != nil {
		return nil, fmt.Errorf("解析进程信息失败: %w", err)
	}

	return &processInfo, nil
}

// GetServerServices 获取主机服务状态
func (s *MonitoringService) GetServerServices(serverID uint) (*ServiceStatus, error) {
	var metricData struct {
		MetricData string `json:"metric_data"`
	}

	err := getDB().Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'service' AND received_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		metrics, err := s.PullExtendedMetrics(serverID)
		if err != nil {
			return nil, err
		}
		return &metrics.ServiceStatus, nil
	}

	var serviceStatus ServiceStatus
	if err := json.Unmarshal([]byte(metricData.MetricData), &serviceStatus); err != nil {
		return nil, fmt.Errorf("解析服务状态失败: %w", err)
	}

	return &serviceStatus, nil
}

// GetServerHardware 获取主机硬件信息
func (s *MonitoringService) GetServerHardware(serverID uint) (*HardwareInfo, error) {
	var metricData struct {
		MetricData string `json:"metric_data"`
	}

	err := getDB().Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'hardware' AND received_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		metrics, err := s.PullExtendedMetrics(serverID)
		if err != nil {
			return nil, err
		}
		return &metrics.HardwareInfo, nil
	}

	var hardwareInfo HardwareInfo
	if err := json.Unmarshal([]byte(metricData.MetricData), &hardwareInfo); err != nil {
		return nil, fmt.Errorf("解析硬件信息失败: %w", err)
	}

	return &hardwareInfo, nil
}

// GetServerNetwork 获取主机网络配置
func (s *MonitoringService) GetServerNetwork(serverID uint) (*NetworkConfig, error) {
	var metricData struct {
		MetricData string `json:"metric_data"`
	}

	err := getDB().Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'network' AND received_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		metrics, err := s.PullExtendedMetrics(serverID)
		if err != nil {
			return nil, err
		}
		return &metrics.NetworkConfig, nil
	}

	var networkConfig NetworkConfig
	if err := json.Unmarshal([]byte(metricData.MetricData), &networkConfig); err != nil {
		return nil, fmt.Errorf("解析网络配置失败: %w", err)
	}

	return &networkConfig, nil
}

// GetServerSecurity 获取主机安全信息
func (s *MonitoringService) GetServerSecurity(serverID uint) (*SecurityInfo, error) {
	var metricData struct {
		MetricData string `json:"metric_data"`
	}

	err := getDB().Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'security' AND received_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		metrics, err := s.PullExtendedMetrics(serverID)
		if err != nil {
			return nil, err
		}
		return &metrics.SecurityInfo, nil
	}

	var securityInfo SecurityInfo
	if err := json.Unmarshal([]byte(metricData.MetricData), &securityInfo); err != nil {
		return nil, fmt.Errorf("解析安全信息失败: %w", err)
	}

	return &securityInfo, nil
}

// GetMetricsHistory 查询历史指标数据
func (s *MonitoringService) GetMetricsHistory(serverID uint, metricType string, startTime, endTime time.Time, interval string) ([]map[string]interface{}, error) {
	jsonPath := metricTypeToJSONPath(metricType)

	query := fmt.Sprintf(`
		SELECT
			DATE_FORMAT(report_time, '%%Y-%%m-%%d %%H:%%i:%%s') as timestamp,
			CAST(JSON_UNQUOTE(JSON_EXTRACT(metric_data, '%s')) AS DECIMAL(10,2)) as value
		FROM mon_agent_metrics
		WHERE server_id = ?
			AND metric_type = 'performance'
			AND report_time BETWEEN ? AND ?
		ORDER BY report_time ASC
		LIMIT 1000
	`, jsonPath)

	rows, err := getDB().Raw(query, serverID, startTime, endTime).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询历史指标失败: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var timestamp string
		var value interface{}
		if err := rows.Scan(&timestamp, &value); err != nil {
			continue
		}
		results = append(results, map[string]interface{}{
			"timestamp": timestamp,
			"value":     value,
		})
	}

	return results, nil
}

// metricTypeToJSONPath 将指标类型映射到 JSON 路径
func metricTypeToJSONPath(metricType string) string {
	switch metricType {
	case "cpu":
		return "$.cpu.usagePercent"
	case "memory":
		return "$.memory.usedPercent"
	case "disk":
		return "$.disk.usedPercent"
	case "load1":
		return "$.load.load1"
	case "load5":
		return "$.load.load5"
	default:
		return "$.cpu.usagePercent"
	}
}
