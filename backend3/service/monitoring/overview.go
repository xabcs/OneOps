package monitoring

import "time"

// OverviewData 监控概览数据
type OverviewData struct {
	Summary      OverviewSummary `json:"summary"`
	TopCPU       []ServerMetric  `json:"topCpu"`
	TopMemory    []ServerMetric  `json:"topMemory"`
	TopDisk      []ServerMetric  `json:"topDisk"`
	ActiveAlerts []AlertItem     `json:"activeAlerts"`
	RefreshTime  string          `json:"refreshTime"`
}

// OverviewSummary 概览统计
type OverviewSummary struct {
	TotalServers   int64 `json:"totalServers"`
	OnlineServers  int64 `json:"onlineServers"`
	OfflineServers int64 `json:"offlineServers"`
	AlertServers   int64 `json:"alertServers"`
}

// ServerMetric 主机指标
type ServerMetric struct {
	ServerID    uint    `json:"serverId"`
	Hostname    string  `json:"hostname"`
	IP          string  `json:"ip"`
	CPUUsage    float64 `json:"cpuUsage,omitempty"`
	MemoryUsage float64 `json:"memoryUsage,omitempty"`
	DiskUsage   float64 `json:"diskUsage,omitempty"`
}

// GetOverview 获取监控概览数据
func (s *MonitoringService) GetOverview() (*OverviewData, error) {
	var summary OverviewSummary

	getDB().Table("cmdb_servers").Count(&summary.TotalServers)
	getDB().Table("cmdb_servers").Where("agent_status = 'running'").Count(&summary.OnlineServers)
	getDB().Table("cmdb_servers").Where("agent_status IN ('offline', 'uninstalled')").Count(&summary.OfflineServers)
	getDB().Table("cmdb_servers").Where("cpu_usage > 90 OR memory_usage > 90 OR disk_usage > 90").Count(&summary.AlertServers)

	var topCPU []struct {
		ID       uint    `gorm:"column:id"`
		Hostname string  `gorm:"column:hostname"`
		IP       string  `gorm:"column:ip"`
		CPUUsage float64 `gorm:"column:cpu_usage"`
	}
	getDB().Table("cmdb_servers").
		Select("id, hostname, ip, cpu_usage").
		Where("agent_status = 'running' AND cpu_usage > 0").
		Order("cpu_usage DESC").
		Limit(10).
		Scan(&topCPU)

	topCPUList := make([]ServerMetric, 0, len(topCPU))
	for _, s := range topCPU {
		topCPUList = append(topCPUList, ServerMetric{ServerID: s.ID, Hostname: s.Hostname, IP: s.IP, CPUUsage: s.CPUUsage})
	}

	var topMemory []struct {
		ID          uint    `gorm:"column:id"`
		Hostname    string  `gorm:"column:hostname"`
		IP          string  `gorm:"column:ip"`
		MemoryUsage float64 `gorm:"column:memory_usage"`
	}
	getDB().Table("cmdb_servers").
		Select("id, hostname, ip, memory_usage").
		Where("agent_status = 'running' AND memory_usage > 0").
		Order("memory_usage DESC").
		Limit(10).
		Scan(&topMemory)

	topMemoryList := make([]ServerMetric, 0, len(topMemory))
	for _, s := range topMemory {
		topMemoryList = append(topMemoryList, ServerMetric{ServerID: s.ID, Hostname: s.Hostname, IP: s.IP, MemoryUsage: s.MemoryUsage})
	}

	var topDisk []struct {
		ID        uint    `gorm:"column:id"`
		Hostname  string  `gorm:"column:hostname"`
		IP        string  `gorm:"column:ip"`
		DiskUsage float64 `gorm:"column:disk_usage"`
	}
	getDB().Table("cmdb_servers").
		Select("id, hostname, ip, disk_usage").
		Where("agent_status = 'running' AND disk_usage > 0").
		Order("disk_usage DESC").
		Limit(10).
		Scan(&topDisk)

	topDiskList := make([]ServerMetric, 0, len(topDisk))
	for _, s := range topDisk {
		topDiskList = append(topDiskList, ServerMetric{ServerID: s.ID, Hostname: s.Hostname, IP: s.IP, DiskUsage: s.DiskUsage})
	}

	activeAlerts, _ := s.getActiveAlerts(10)

	return &OverviewData{
		Summary:      summary,
		TopCPU:       topCPUList,
		TopMemory:    topMemoryList,
		TopDisk:      topDiskList,
		ActiveAlerts: activeAlerts,
		RefreshTime:  time.Now().Format(time.RFC3339),
	}, nil
}
