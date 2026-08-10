package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func getDB() *gorm.DB {
	return database.GetDB()
}

const (
	// AgentPullTimeout Agent 拉取超时时间
	AgentPullTimeout = 15 * time.Second
	// ExtendedMetricsPath Agent 扩展指标接口路径
	ExtendedMetricsPath = "/extended-metrics"
)

// AgentExtendedMetricsResponse Agent 扩展指标响应结构
type AgentExtendedMetricsResponse struct {
	Performance   PerformanceMetrics `json:"performance"`
	SystemInfo    SystemInfo         `json:"systemInfo,omitempty"`
	HardwareInfo  HardwareInfo       `json:"hardwareInfo,omitempty"`
	ServiceStatus ServiceStatus      `json:"serviceStatus,omitempty"`
	ProcessInfo   ProcessInfo        `json:"processInfo,omitempty"`
	NetworkConfig NetworkConfig      `json:"networkConfig,omitempty"`
	SecurityInfo  SecurityInfo       `json:"securityInfo,omitempty"`
	ConfigChanges []ConfigChange     `json:"configChanges,omitempty"`
	CollectedAt   string             `json:"collectedAt"`
	Version       string             `json:"version"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Disk    DiskMetrics    `json:"disk"`
	Network NetworkMetrics `json:"network"`
	Load    LoadMetrics    `json:"load"`
	IO      IOStats        `json:"io,omitempty"`
}

// CPUMetrics CPU 指标
type CPUMetrics struct {
	UsagePercent float64 `json:"usagePercent"`
	User         float64 `json:"user"`
	System       float64 `json:"system"`
	Idle         float64 `json:"idle"`
	Iowait       float64 `json:"iowait"`
	Cores        int     `json:"cores"`
	Mhz          float64 `json:"mhz"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
	Available   uint64  `json:"available"`
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	Total       uint64          `json:"total"`
	Used        uint64          `json:"used"`
	Free        uint64          `json:"free"`
	UsedPercent float64         `json:"usedPercent"`
	Partitions  []PartitionInfo `json:"partitions,omitempty"`
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	Interfaces  []InterfaceStats `json:"interfaces,omitempty"`
	Connections ConnectionStats  `json:"connections"`
}

// InterfaceStats 网卡统计
type InterfaceStats struct {
	Name      string `json:"name"`
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`
}

// ConnectionStats 连接统计
type ConnectionStats struct {
	Established int `json:"established"`
	TimeWait    int `json:"timeWait"`
	Listen      int `json:"listen"`
}

// LoadMetrics 负载指标
type LoadMetrics struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// IOStats IO 统计指标
type IOStats struct {
	ReadIOPS        float64 `json:"readIOPS,omitempty"`
	WriteIOPS       float64 `json:"writeIOPS,omitempty"`
	ReadThroughput  float64 `json:"readThroughput,omitempty"`
	WriteThroughput float64 `json:"writeThroughput,omitempty"`
	Await           float64 `json:"await,omitempty"`
	QueueDepth      float64 `json:"queueDepth,omitempty"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname string `json:"hostname"`
	OS       OSInfo `json:"os"`
	Uptime   uint64 `json:"uptime"`
}

// OSInfo 操作系统信息
type OSInfo struct {
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platformVersion"`
	KernelVersion   string `json:"kernelVersion"`
	KernelArch      string `json:"kernelArch"`
}

// HardwareInfo 硬件信息
type HardwareInfo struct {
	CPU    CPUInfo        `json:"cpu"`
	Memory MemoryHardware `json:"memory"`
	Disk   []DiskDevice   `json:"disk,omitempty"`
}

// CPUInfo CPU 信息
type CPUInfo struct {
	Vendor  string  `json:"vendor,omitempty"`
	Model   string  `json:"model,omitempty"`
	Cores   int     `json:"cores"`
	Threads int     `json:"threads"`
	Mhz     float64 `json:"mhz"`
}

// MemoryHardware 内存硬件信息
type MemoryHardware struct {
	Total uint64       `json:"total"`
	Slots []MemorySlot `json:"slots,omitempty"`
}

// MemorySlot 内存插槽信息
type MemorySlot struct {
	SlotNumber int    `json:"slotNumber"`
	Capacity   uint64 `json:"capacity"`
	Type       string `json:"type"`
	Vendor     string `json:"vendor"`
	Speed      string `json:"speed"`
	HasECC     bool   `json:"hasECC"`
}

// DiskDevice 磁盘设备
type DiskDevice struct {
	Name   string `json:"name,omitempty"`
	Model  string `json:"model,omitempty"`
	Serial string `json:"serial,omitempty"`
	Size   uint64 `json:"size,omitempty"`
	Type   string `json:"type,omitempty"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	SystemdServices []SystemdService `json:"systemdServices,omitempty"`
	ListenPorts     []ListenPort     `json:"listenPorts,omitempty"`
}

// SystemdService systemd 服务
type SystemdService struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	SubStatus   string `json:"subStatus"`
	ActiveState string `json:"activeState"`
	SubState    string `json:"subState"`
	Description string `json:"description,omitempty"`
}

// ListenPort 监听端口
type ListenPort struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Process  string `json:"process,omitempty"`
	PID      int    `json:"pid,omitempty"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	Total int          `json:"total"`
	Top   []TopProcess `json:"top"`
}

// TopProcess Top 进程
type TopProcess struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float32 `json:"memoryPercent"`
	MemoryBytes   uint64  `json:"memoryBytes"`
	Status        string  `json:"status,omitempty"`
	Username      string  `json:"username,omitempty"`
	NumThreads    int     `json:"numThreads,omitempty"`
	Cmdline       string  `json:"cmdline,omitempty"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
}

// NetworkInterface 网络接口
type NetworkInterface struct {
	Name         string          `json:"name"`
	HardwareAddr string          `json:"hardwareAddr,omitempty"`
	MTU          int             `json:"mtu,omitempty"`
	Addrs        []InterfaceAddr `json:"addrs,omitempty"`
}

// InterfaceAddr 接口地址
type InterfaceAddr struct {
	IP string `json:"ip"`
}

// SecurityInfo 安全信息
type SecurityInfo struct {
	SSH      SSHConfig      `json:"ssh,omitempty"`
	Firewall FirewallConfig `json:"firewall,omitempty"`
}

// SSHConfig SSH 配置
type SSHConfig struct {
	Port                   int    `json:"port,omitempty"`
	PermitRootLogin        string `json:"permitRootLogin,omitempty"`
	PasswordAuthentication string `json:"passwordAuthentication,omitempty"`
}

// FirewallConfig 防火墙配置
type FirewallConfig struct {
	Backend string `json:"backend,omitempty"`
	Status  string `json:"status,omitempty"`
}

// ConfigChange 配置变更
type ConfigChange struct {
	File       string `json:"file,omitempty"`
	Checksum   string `json:"checksum,omitempty"`
	ChangeTime string `json:"changeTime,omitempty"`
	ChangeType string `json:"changeType,omitempty"`
}

// MonitoringService 监控服务
type MonitoringService struct{}

// NewMonitoringService 创建监控服务
func NewMonitoringService() *MonitoringService {
	return &MonitoringService{}
}

// PullExtendedMetrics 从 Agent 拉取扩展指标
func (s *MonitoringService) PullExtendedMetrics(serverID uint) (*AgentExtendedMetricsResponse, error) {
	var server modelcmdb.Server
	if err := getDB().First(&server, serverID).Error; err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	agentPort := server.AgentPort
	if agentPort == 0 {
		agentPort = 9100
	}

	ip := server.InnerIP
	if ip == "" {
		ip = server.IP
	}

	url := fmt.Sprintf("http://%s:%d%s", ip, agentPort, ExtendedMetricsPath)
	ctx, cancel := context.WithTimeout(context.Background(), AgentPullTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Agent 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		logger.Warn("Agent不支持扩展指标端点（可能是旧版本），返回空数据",
			zap.Uint("serverID", serverID),
			zap.String("url", url))
		return &AgentExtendedMetricsResponse{
			Performance: PerformanceMetrics{
				CPU:     CPUMetrics{UsagePercent: 0, Cores: 0},
				Memory:  MemoryMetrics{Total: 0, Used: 0, Free: 0, UsedPercent: 0},
				Disk:    DiskMetrics{Total: 0, Used: 0, Free: 0, UsedPercent: 0},
				Network: NetworkMetrics{Connections: ConnectionStats{}},
				Load:    LoadMetrics{Load1: 0, Load5: 0, Load15: 0},
			},
			ProcessInfo:   ProcessInfo{Total: 0, Top: []TopProcess{}},
			ServiceStatus: ServiceStatus{SystemdServices: []SystemdService{}, ListenPorts: []ListenPort{}},
			CollectedAt:   time.Now().Format(time.RFC3339),
			Version:       "unknown",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent 返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var metrics AgentExtendedMetricsResponse
	if err := json.Unmarshal(body, &metrics); err != nil {
		return nil, fmt.Errorf("解析指标失败: %w", err)
	}

	if err := s.storeMetrics(serverID, &metrics); err != nil {
		logger.Error("存储扩展指标失败", zap.Uint("serverID", serverID), zap.Error(err))
	}

	now := time.Now()
	if err := getDB().Model(&modelcmdb.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"last_heartbeat_at": now,
		"agent_status":      "running",
	}).Error; err != nil {
		logger.Warn("更新主机心跳时间失败", zap.Uint("serverID", serverID), zap.Error(err))
	}

	cache := database.NewRedisCache()
	if err := cache.CacheServerMetricsWithTTL(serverID, metrics, 2*time.Minute); err != nil {
		logger.Warn("缓存主机指标到 Redis 失败", zap.Uint("serverID", serverID), zap.Error(err))
	} else {
		logger.Debug("主机指标已缓存到 Redis (TTL: 2分钟)", zap.Uint("serverID", serverID))
	}

	logger.Debug("Agent 扩展指标拉取成功",
		zap.Uint("serverID", serverID),
		zap.Float64("cpu", metrics.Performance.CPU.UsagePercent),
		zap.Float64("memory", metrics.Performance.Memory.UsedPercent))

	return &metrics, nil
}

// storeMetrics 存储指标到数据库
func (s *MonitoringService) storeMetrics(serverID uint, metrics *AgentExtendedMetricsResponse) error {
	now := time.Now()

	perfData, _ := json.Marshal(metrics.Performance)
	perfQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
	              VALUES (?, 'performance', ?, ?, NOW())`
	if err := getDB().Exec(perfQuery, serverID, string(perfData), now).Error; err != nil {
		logger.Warn("存储性能指标失败", zap.Error(err))
	}

	if metrics.SystemInfo.Hostname != "" {
		sysData, _ := json.Marshal(metrics.SystemInfo)
		sysQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'system', ?, ?, NOW())`
		if err := getDB().Exec(sysQuery, serverID, string(sysData), now).Error; err != nil {
			logger.Warn("存储系统信息失败", zap.Error(err))
		}
	}

	if metrics.HardwareInfo.CPU.Cores > 0 {
		hwData, _ := json.Marshal(metrics.HardwareInfo)
		hwQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		            VALUES (?, 'hardware', ?, ?, NOW())`
		if err := getDB().Exec(hwQuery, serverID, string(hwData), now).Error; err != nil {
			logger.Warn("存储硬件信息失败", zap.Error(err))
		}
	}

	if len(metrics.ServiceStatus.SystemdServices) > 0 || len(metrics.ServiceStatus.ListenPorts) > 0 {
		svcData, _ := json.Marshal(metrics.ServiceStatus)
		svcQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'service', ?, ?, NOW())`
		if err := getDB().Exec(svcQuery, serverID, string(svcData), now).Error; err != nil {
			logger.Warn("存储服务状态失败", zap.Error(err))
		}
	}

	if len(metrics.ProcessInfo.Top) > 0 {
		procData, _ := json.Marshal(metrics.ProcessInfo)
		procQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		              VALUES (?, 'process', ?, ?, NOW())`
		if err := getDB().Exec(procQuery, serverID, string(procData), now).Error; err != nil {
			logger.Warn("存储进程信息失败", zap.Error(err))
		}
	}

	if len(metrics.NetworkConfig.Interfaces) > 0 {
		netData, _ := json.Marshal(metrics.NetworkConfig)
		netQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'network', ?, ?, NOW())`
		if err := getDB().Exec(netQuery, serverID, string(netData), now).Error; err != nil {
			logger.Warn("存储网络配置失败", zap.Error(err))
		}
	}

	if metrics.SecurityInfo.SSH.Port > 0 {
		secData, _ := json.Marshal(metrics.SecurityInfo)
		secQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'security', ?, ?, NOW())`
		if err := getDB().Exec(secQuery, serverID, string(secData), now).Error; err != nil {
			logger.Warn("存储安全信息失败", zap.Error(err))
		}
	}

	updates := map[string]interface{}{
		"cpu_usage":          metrics.Performance.CPU.UsagePercent,
		"memory_usage":       metrics.Performance.Memory.UsedPercent,
		"disk_usage":         metrics.Performance.Disk.UsedPercent,
		"load1":              metrics.Performance.Load.Load1,
		"load5":              metrics.Performance.Load.Load5,
		"load15":             metrics.Performance.Load.Load15,
		"metrics_updated_at": now,
	}

	if metrics.SystemInfo.Hostname != "" {
		updates["hostname"] = metrics.SystemInfo.Hostname
	}

	if err := getDB().Model(&modelcmdb.Server{}).Where("id = ?", serverID).Updates(updates).Error; err != nil {
		logger.Warn("更新服务器基本信息失败", zap.Error(err))
	}

	extendedData, _ := json.Marshal(metrics)
	extendedQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		                  VALUES (?, 'extended', ?, ?, NOW())`
	if err := getDB().Exec(extendedQuery, serverID, string(extendedData), now).Error; err != nil {
		logger.Warn("存储扩展指标失败", zap.Error(err))
	}

	hostname := metrics.SystemInfo.Hostname
	ip := ""
	var srv struct {
		IP string `gorm:"column:ip"`
	}
	if err := getDB().Table("cmdb_servers").Select("ip").Where("id = ?", serverID).First(&srv).Error; err == nil {
		ip = srv.IP
	}
	if hostname == "" {
		hostname = ip
	}
	s.checkAlertThresholds(serverID, hostname, ip, metrics)

	return nil
}
