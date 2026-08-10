package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"oneops/backend/logger"
	"oneops/backend/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

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
	IO      IOStats        `json:"io,omitempty"` // P1: IO 统计
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

// IOStats IO 统计指标 (P1)
type IOStats struct {
	ReadIOPS        float64 `json:"readIOPS,omitempty"`        // 读 IOPS
	WriteIOPS       float64 `json:"writeIOPS,omitempty"`       // 写 IOPS
	ReadThroughput  float64 `json:"readThroughput,omitempty"`  // 读吞吐量 (MB/s)
	WriteThroughput float64 `json:"writeThroughput,omitempty"` // 写吞吐量 (MB/s)
	Await           float64 `json:"await,omitempty"`           // 平均等待时间 (ms)
	QueueDepth      float64 `json:"queueDepth,omitempty"`      // 队列深度
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
	Slots []MemorySlot `json:"slots,omitempty"` // P1: 内存插槽信息
}

// MemorySlot 内存插槽信息 (P1)
type MemorySlot struct {
	SlotNumber int    `json:"slotNumber"` // 插槽号
	Capacity   uint64 `json:"capacity"`   // 容量 (字节)
	Type       string `json:"type"`       // 类型 (DDR3, DDR4, etc.)
	Vendor     string `json:"vendor"`     // 厂商
	Speed      string `json:"speed"`      // 速度
	HasECC     bool   `json:"hasECC"`     // 是否支持 ECC
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
	var server models.Server
	if err := db.First(&server, serverID).Error; err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	// 移除状态检查，允许通过HTTP拉取来恢复Agent状态

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

	// 如果Agent返回404，说明Agent版本过旧不支持扩展指标，返回空数据
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

	// 存储指标到数据库
	if err := s.storeMetrics(serverID, &metrics); err != nil {
		logger.Error("存储扩展指标失败", zap.Uint("serverID", serverID), zap.Error(err))
		// 不中断流程，继续返回数据
	}

	// 更新主机的最后心跳时间（用于判断Agent在线状态）
	now := time.Now()
	if err := db.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"last_heartbeat_at": now,
		"agent_status":      "running",
	}).Error; err != nil {
		logger.Warn("更新主机心跳时间失败", zap.Uint("serverID", serverID), zap.Error(err))
	}

	// 缓存到 Redis（2 分钟过期）
	cache := NewRedisCache()
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

	// 存储性能指标
	perfData, _ := json.Marshal(metrics.Performance)
	perfQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
	              VALUES (?, 'performance', ?, ?, NOW())`
	if err := db.Exec(perfQuery, serverID, string(perfData), now).Error; err != nil {
		logger.Warn("存储性能指标失败", zap.Error(err))
	}

	// 存储系统信息
	if metrics.SystemInfo.Hostname != "" {
		sysData, _ := json.Marshal(metrics.SystemInfo)
		sysQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'system', ?, ?, NOW())`
		if err := db.Exec(sysQuery, serverID, string(sysData), now).Error; err != nil {
			logger.Warn("存储系统信息失败", zap.Error(err))
		}
	}

	// 存储硬件信息（如果存在）
	if metrics.HardwareInfo.CPU.Cores > 0 {
		hwData, _ := json.Marshal(metrics.HardwareInfo)
		hwQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		            VALUES (?, 'hardware', ?, ?, NOW())`
		if err := db.Exec(hwQuery, serverID, string(hwData), now).Error; err != nil {
			logger.Warn("存储硬件信息失败", zap.Error(err))
		}
	}

	// 存储服务状态
	if len(metrics.ServiceStatus.SystemdServices) > 0 || len(metrics.ServiceStatus.ListenPorts) > 0 {
		svcData, _ := json.Marshal(metrics.ServiceStatus)
		svcQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'service', ?, ?, NOW())`
		if err := db.Exec(svcQuery, serverID, string(svcData), now).Error; err != nil {
			logger.Warn("存储服务状态失败", zap.Error(err))
		}
	}

	// 存储进程信息
	if len(metrics.ProcessInfo.Top) > 0 {
		procData, _ := json.Marshal(metrics.ProcessInfo)
		procQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		              VALUES (?, 'process', ?, ?, NOW())`
		if err := db.Exec(procQuery, serverID, string(procData), now).Error; err != nil {
			logger.Warn("存储进程信息失败", zap.Error(err))
		}
	}

	// 存储网络配置
	if len(metrics.NetworkConfig.Interfaces) > 0 {
		netData, _ := json.Marshal(metrics.NetworkConfig)
		netQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'network', ?, ?, NOW())`
		if err := db.Exec(netQuery, serverID, string(netData), now).Error; err != nil {
			logger.Warn("存储网络配置失败", zap.Error(err))
		}
	}

	// 存储安全信息
	if metrics.SecurityInfo.SSH.Port > 0 {
		secData, _ := json.Marshal(metrics.SecurityInfo)
		secQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		             VALUES (?, 'security', ?, ?, NOW())`
		if err := db.Exec(secQuery, serverID, string(secData), now).Error; err != nil {
			logger.Warn("存储安全信息失败", zap.Error(err))
		}
	}

	// 更新服务器基本信息
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

	if err := db.Model(&models.Server{}).Where("id = ?", serverID).Updates(updates).Error; err != nil {
		logger.Warn("更新服务器基本信息失败", zap.Error(err))
	}

	// 存储完整的 extended 数据（包含所有指标）
	extendedData, _ := json.Marshal(metrics)
	extendedQuery := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
		                  VALUES (?, 'extended', ?, ?, NOW())`
	if err := db.Exec(extendedQuery, serverID, string(extendedData), now).Error; err != nil {
		logger.Warn("存储扩展指标失败", zap.Error(err))
	}

	// 告警阈值检查
	hostname := metrics.SystemInfo.Hostname
	ip := ""
	// 从数据库获取 IP
	var srv struct {
		IP string `gorm:"column:ip"`
	}
	if err := db.Table("cmdb_servers").Select("ip").Where("id = ?", serverID).First(&srv).Error; err == nil {
		ip = srv.IP
	}
	if hostname == "" {
		hostname = ip
	}
	s.checkAlertThresholds(serverID, hostname, ip, metrics)

	return nil
}

// alertRule 告警规则定义
type alertRule struct {
	RuleID    string
	Level     string
	Threshold float64
	Message   string
	Value     float64
	Condition string // 判断条件: >, <, ==, !=
}

// checkAlertThresholds 检查指标是否触发告警规则
func (s *MonitoringService) checkAlertThresholds(serverID uint, hostname, ip string, metrics *AgentExtendedMetricsResponse) {
	// 从数据库加载启用的告警规则
	alertRuleService := NewAlertRuleService()
	dbRules, err := alertRuleService.LoadActiveAlertRules()
	if err != nil {
		logger.Error("加载告警规则失败，使用默认规则", zap.Error(err))
		// 如果加载失败，使用空规则集，不进行告警检测
		return
	}

	// 如果没有配置规则，不进行告警检测
	if len(dbRules) == 0 {
		logger.Debug("没有启用的告警规则，跳过告警检测")
		return
	}

	// 构建指标值映射
	metricValues := map[string]float64{
		"cpu_usage":    metrics.Performance.CPU.UsagePercent,
		"memory_usage": metrics.Performance.Memory.UsedPercent,
		"disk_usage":   metrics.Performance.Disk.UsedPercent,
		"load1":        float64(metrics.Performance.Load.Load1),
		"load5":        float64(metrics.Performance.Load.Load5),
		"load15":       float64(metrics.Performance.Load.Load15),
	}

	now := time.Now()

	for _, dbRule := range dbRules {
		// 获取指标值
		value, exists := metricValues[dbRule.Metric]
		if !exists {
			continue
		}

		// 判断是否触发告警
		triggered := false
		switch dbRule.Condition {
		case ">":
			triggered = value > dbRule.Threshold
		case "<":
			triggered = value < dbRule.Threshold
		case "==":
			triggered = value == dbRule.Threshold
		case "!=":
			triggered = value != dbRule.Threshold
		default:
			logger.Warn("未知的告警条件", zap.String("condition", dbRule.Condition))
			continue
		}

		if triggered {
			// 触发告警：生成告警消息
			message := fmt.Sprintf("%s %s %.1f%% (阈值 %.1f%%)", dbRule.Name, dbRule.Condition, value, dbRule.Threshold)
			if dbRule.Description != "" {
				message = dbRule.Description
			}

			// 告警聚合：检查是否在抑制期内（防止告警风暴）
			cache := NewRedisCache()
			if cache.IsAlertSuppressed(serverID, dbRule.ID) {
				// 在抑制期内，只更新数据库中的 last_seen，不发送新告警通知
				db.Exec(`UPDATE mon_agent_alerts SET last_seen = ?, metric_value = ?, message = ?
					         WHERE server_id = ? AND rule_id = ? AND resolved_at IS NULL`,
					now, value, message, serverID, dbRule.ID)
				logger.Debug("告警已被聚合抑制，仅更新时间戳",
					zap.Uint("serverID", serverID),
					zap.String("ruleId", dbRule.ID))
				continue
			}

			// 检查是否已存在未解决的告警
			var existing struct {
				ID uint `gorm:"column:id"`
			}
			err := db.Table("mon_agent_alerts").
				Select("id").
				Where("server_id = ? AND rule_id = ? AND resolved_at IS NULL", serverID, dbRule.ID).
				First(&existing).Error
			if err != nil {
				// 新增告警
				insertSQL := `INSERT INTO mon_agent_alerts
					(server_id, hostname, ip, rule_id, level, message, metric_value, threshold, first_seen, last_seen)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
				if err := db.Exec(insertSQL, serverID, hostname, ip, dbRule.ID, dbRule.Level,
					message, value, dbRule.Threshold, now, now).Error; err != nil {
					logger.Warn("插入告警失败", zap.String("ruleId", dbRule.ID), zap.Error(err))
				} else {
					logger.Info("触发告警",
						zap.Uint("serverID", serverID),
						zap.String("hostname", hostname),
						zap.String("ruleId", dbRule.ID),
						zap.String("level", dbRule.Level),
						zap.Float64("value", value))

					// 设置告警抑制标记（防止告警风暴）
					suppressDuration := time.Duration(dbRule.Duration) * time.Second
					if suppressDuration == 0 {
						suppressDuration = 5 * time.Minute // 默认抑制5分钟
					}
					cache.SetAlertSuppress(serverID, dbRule.ID, suppressDuration)

					// 发布告警事件到 Redis（用于 WebSocket 推送）
					alertEvent := map[string]interface{}{
						"type":      "new_alert",
						"serverID":  serverID,
						"hostname":  hostname,
						"ip":        ip,
						"ruleID":    dbRule.ID,
						"level":     dbRule.Level,
						"message":   message,
						"value":     value,
						"threshold": dbRule.Threshold,
						"timestamp": now.Format(time.RFC3339),
					}
					cache.PublishAlert(alertEvent)
				}
			} else {
				// 更新 last_seen 和当前值
				db.Exec(`UPDATE mon_agent_alerts SET last_seen = ?, metric_value = ?, message = ? WHERE id = ?`,
					now, value, message, existing.ID)
			}
		} else {
			// 未触发告警：将未解决的告警标记为已恢复
			result := db.Exec(`UPDATE mon_agent_alerts SET resolved_at = ? WHERE server_id = ? AND rule_id = ? AND resolved_at IS NULL`,
				now, serverID, dbRule.ID)
			if result.RowsAffected > 0 {
				logger.Info("告警已恢复",
					zap.Uint("serverID", serverID),
					zap.String("hostname", hostname),
					zap.String("ruleId", dbRule.ID))
			}
		}
	}
}

// GetServerExtendedMetrics 获取主机扩展指标（优先从 Redis 缓存，未命中则从数据库查询）
func (s *MonitoringService) GetServerExtendedMetrics(serverID uint) (*AgentExtendedMetricsResponse, error) {
	// 优先尝试从 Redis 获取
	cache := NewRedisCache()
	var cachedMetrics AgentExtendedMetricsResponse
	if err := cache.GetServerMetrics(serverID, &cachedMetrics); err == nil {
		logger.Debug("从 Redis 缓存获取主机指标", zap.Uint("serverID", serverID))
		return &cachedMetrics, nil
	}

	// Redis 缓存未命中，尝试从数据库获取完整 extended 数据
	var metricData struct {
		MetricData string    `json:"metric_data"`
		ReceivedAt time.Time `json:"received_at"`
	}

	// 查询5分钟内的完整扩展指标
	err := db.Table("mon_agent_metrics").
		Select("metric_data, received_at").
		Where("server_id = ? AND metric_type = 'extended' AND received_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err == nil && metricData.MetricData != "" {
		var metrics AgentExtendedMetricsResponse
		if err := json.Unmarshal([]byte(metricData.MetricData), &metrics); err == nil {
			// 写入 Redis 缓存
			cache.CacheServerMetrics(serverID, metrics)
			return &metrics, nil
		}
	}

	// 数据库也未命中，实时拉取
	return s.PullExtendedMetrics(serverID)
}

// GetServerProcesses 获取主机进程信息
func (s *MonitoringService) GetServerProcesses(serverID uint) (*ProcessInfo, error) {
	var metricData struct {
		MetricData string `json:"metric_data"`
	}

	err := db.Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'process' AND received_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		// 如果缓存未命中，尝试拉取
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

	err := db.Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'service' AND received_at > DATE_SUB(NOW(), INTERVAL 5 MINUTE)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		// 如果缓存未命中，尝试拉取
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

	err := db.Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'hardware' AND received_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		// 如果缓存未命中，尝试拉取
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

	err := db.Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'network' AND received_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		// 如果缓存未命中，尝试拉取
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

	err := db.Table("mon_agent_metrics").
		Select("metric_data").
		Where("server_id = ? AND metric_type = 'security' AND received_at > DATE_SUB(NOW(), INTERVAL 1 HOUR)", serverID).
		Order("received_at DESC").
		First(&metricData).Error

	if err != nil {
		// 如果缓存未命中，尝试拉取
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
	// 根据指标类型确定 JSON 路径
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

	rows, err := db.Raw(query, serverID, startTime, endTime).Rows()
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

// ServerMetric 主机指标（通用，用于排行）
type ServerMetric struct {
	ServerID    uint    `json:"serverId"`
	Hostname    string  `json:"hostname"`
	IP          string  `json:"ip"`
	CPUUsage    float64 `json:"cpuUsage,omitempty"`
	MemoryUsage float64 `json:"memoryUsage,omitempty"`
	DiskUsage   float64 `json:"diskUsage,omitempty"`
}

// AlertItem 告警条目
type AlertItem struct {
	ID             uint64  `json:"id"`
	ServerID       uint    `json:"serverId"`
	Hostname       string  `json:"hostname"`
	IP             string  `json:"ip"`
	RuleID         string  `json:"ruleId"`
	Level          string  `json:"level"`
	Message        string  `json:"message"`
	MetricValue    float64 `json:"metricValue"`
	Threshold      float64 `json:"threshold"`
	FirstSeen      string  `json:"firstSeen"`
	LastSeen       string  `json:"lastSeen"`
	Acknowledged   bool    `json:"acknowledged"`
	AcknowledgedBy string  `json:"acknowledgedBy,omitempty"`
	AcknowledgedAt string  `json:"acknowledgedAt,omitempty"`
	ResolvedAt     string  `json:"resolvedAt,omitempty"`
}

// AlertQueryParams 告警查询参数
type AlertQueryParams struct {
	ServerID     string
	Level        string
	Acknowledged string
	Page         int
	PageSize     int
}

// AlertListResult 告警列表结果
type AlertListResult struct {
	Total int64       `json:"total"`
	Items []AlertItem `json:"items"`
}

// AlertStats 告警统计
type AlertStats struct {
	Total          int64             `json:"total"`
	ByLevel        map[string]int64  `json:"byLevel"`
	Acknowledged   int64             `json:"acknowledged"`
	Unacknowledged int64             `json:"unacknowledged"`
	Resolved       int64             `json:"resolved"`
	Active         int64             `json:"active"`
	Trend          []AlertTrendPoint `json:"trend"`
}

// AlertTrendPoint 告警趋势点
type AlertTrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// GetOverview 获取监控概览数据
func (s *MonitoringService) GetOverview() (*OverviewData, error) {
	var summary OverviewSummary

	// 统计主机总数
	db.Table("cmdb_servers").Count(&summary.TotalServers)

	// 统计在线主机（agent_status = running）
	db.Table("cmdb_servers").Where("agent_status = 'running'").Count(&summary.OnlineServers)

	// 统计离线主机（agent_status = offline 或 uninstalled 且没有心跳）
	db.Table("cmdb_servers").Where("agent_status IN ('offline', 'uninstalled')").Count(&summary.OfflineServers)

	// 告警主机：CPU > 90% 或内存 > 90% 或磁盘 > 90%
	db.Table("cmdb_servers").Where("cpu_usage > 90 OR memory_usage > 90 OR disk_usage > 90").Count(&summary.AlertServers)

	// Top 10 CPU
	var topCPU []struct {
		ID       uint    `gorm:"column:id"`
		Hostname string  `gorm:"column:hostname"`
		IP       string  `gorm:"column:ip"`
		CPUUsage float64 `gorm:"column:cpu_usage"`
	}
	db.Table("cmdb_servers").
		Select("id, hostname, ip, cpu_usage").
		Where("agent_status = 'running' AND cpu_usage > 0").
		Order("cpu_usage DESC").
		Limit(10).
		Scan(&topCPU)

	topCPUList := make([]ServerMetric, 0, len(topCPU))
	for _, s := range topCPU {
		topCPUList = append(topCPUList, ServerMetric{ServerID: s.ID, Hostname: s.Hostname, IP: s.IP, CPUUsage: s.CPUUsage})
	}

	// Top 10 内存
	var topMemory []struct {
		ID          uint    `gorm:"column:id"`
		Hostname    string  `gorm:"column:hostname"`
		IP          string  `gorm:"column:ip"`
		MemoryUsage float64 `gorm:"column:memory_usage"`
	}
	db.Table("cmdb_servers").
		Select("id, hostname, ip, memory_usage").
		Where("agent_status = 'running' AND memory_usage > 0").
		Order("memory_usage DESC").
		Limit(10).
		Scan(&topMemory)

	topMemoryList := make([]ServerMetric, 0, len(topMemory))
	for _, s := range topMemory {
		topMemoryList = append(topMemoryList, ServerMetric{ServerID: s.ID, Hostname: s.Hostname, IP: s.IP, MemoryUsage: s.MemoryUsage})
	}

	// Top 10 磁盘
	var topDisk []struct {
		ID        uint    `gorm:"column:id"`
		Hostname  string  `gorm:"column:hostname"`
		IP        string  `gorm:"column:ip"`
		DiskUsage float64 `gorm:"column:disk_usage"`
	}
	db.Table("cmdb_servers").
		Select("id, hostname, ip, disk_usage").
		Where("agent_status = 'running' AND disk_usage > 0").
		Order("disk_usage DESC").
		Limit(10).
		Scan(&topDisk)

	topDiskList := make([]ServerMetric, 0, len(topDisk))
	for _, s := range topDisk {
		topDiskList = append(topDiskList, ServerMetric{ServerID: s.ID, Hostname: s.Hostname, IP: s.IP, DiskUsage: s.DiskUsage})
	}

	// 最新活跃告警（10条）
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

// getActiveAlerts 获取活跃告警列表（基于 mon_agent_alerts 表）
func (s *MonitoringService) getActiveAlerts(limit int) ([]AlertItem, error) {
	query := `
		SELECT a.id, a.server_id, a.hostname, a.ip, a.rule_id, a.level, a.message,
		       a.metric_value, a.threshold,
		       DATE_FORMAT(a.first_seen, '%Y-%m-%dT%H:%i:%s+08:00') as first_seen,
		       DATE_FORMAT(a.last_seen, '%Y-%m-%dT%H:%i:%s+08:00') as last_seen,
		       a.acknowledged, COALESCE(a.acknowledged_by, '') as acknowledged_by,
		       COALESCE(DATE_FORMAT(a.acknowledged_at, '%Y-%m-%dT%H:%i:%s+08:00'), '') as acknowledged_at,
		       COALESCE(DATE_FORMAT(a.resolved_at, '%Y-%m-%dT%H:%i:%s+08:00'), '') as resolved_at
		FROM mon_agent_alerts a
		WHERE a.resolved_at IS NULL
		ORDER BY
		    CASE a.level
		        WHEN 'critical' THEN 1
		        WHEN 'high' THEN 2
		        WHEN 'medium' THEN 3
		        WHEN 'low' THEN 4
		        ELSE 5
		    END, a.last_seen DESC
		LIMIT ?
	`
	rows, err := db.Raw(query, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []AlertItem
	for rows.Next() {
		var a AlertItem
		if err := rows.Scan(
			&a.ID, &a.ServerID, &a.Hostname, &a.IP, &a.RuleID, &a.Level, &a.Message,
			&a.MetricValue, &a.Threshold, &a.FirstSeen, &a.LastSeen,
			&a.Acknowledged, &a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedAt,
		); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

// GetAlerts 查询告警列表
func (s *MonitoringService) GetAlerts(params AlertQueryParams) (*AlertListResult, error) {
	baseQuery := `FROM mon_agent_alerts WHERE 1=1`
	var args []interface{}

	if params.ServerID != "" {
		baseQuery += " AND server_id = ?"
		args = append(args, params.ServerID)
	}
	if params.Level != "" {
		baseQuery += " AND level = ?"
		args = append(args, params.Level)
	}
	if params.Acknowledged == "true" {
		baseQuery += " AND acknowledged = 1"
	} else if params.Acknowledged == "false" {
		baseQuery += " AND acknowledged = 0"
	}

	// 统计总数
	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("查询告警总数失败: %w", err)
	}

	// 查询列表
	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, server_id, hostname, ip, rule_id, level, message,
		       metric_value, threshold,
		       DATE_FORMAT(first_seen, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00') as first_seen,
		       DATE_FORMAT(last_seen, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00') as last_seen,
		       acknowledged, COALESCE(acknowledged_by, '') as acknowledged_by,
		       COALESCE(DATE_FORMAT(acknowledged_at, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00'), '') as acknowledged_at,
		       COALESCE(DATE_FORMAT(resolved_at, '%%Y-%%m-%%dT%%H:%%i:%%s+08:00'), '') as resolved_at
		%s
		ORDER BY
		    CASE level
		        WHEN 'critical' THEN 1
		        WHEN 'high' THEN 2
		        WHEN 'medium' THEN 3
		        WHEN 'low' THEN 4
		        ELSE 5
		    END, last_seen DESC
		LIMIT %d OFFSET %d
	`, baseQuery, params.PageSize, offset)

	rows, err := db.Raw(dataQuery, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询告警列表失败: %w", err)
	}
	defer rows.Close()

	var items []AlertItem
	for rows.Next() {
		var a AlertItem
		if err := rows.Scan(
			&a.ID, &a.ServerID, &a.Hostname, &a.IP, &a.RuleID, &a.Level, &a.Message,
			&a.MetricValue, &a.Threshold, &a.FirstSeen, &a.LastSeen,
			&a.Acknowledged, &a.AcknowledgedBy, &a.AcknowledgedAt, &a.ResolvedAt,
		); err != nil {
			continue
		}
		items = append(items, a)
	}

	if items == nil {
		items = []AlertItem{}
	}

	return &AlertListResult{Total: total, Items: items}, nil
}

// AcknowledgeAlert 确认告警
func (s *MonitoringService) AcknowledgeAlert(id uint64, acknowledgedBy, comment string) error {
	now := time.Now()
	result := db.Exec(
		`UPDATE mon_agent_alerts SET acknowledged = 1, acknowledged_by = ?, acknowledged_at = ? WHERE id = ?`,
		acknowledgedBy, now, id,
	)
	if result.Error != nil {
		return fmt.Errorf("确认告警失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("告警不存在: %d", id)
	}
	return nil
}

// GetAlertStats 获取告警统计
func (s *MonitoringService) GetAlertStats() (*AlertStats, error) {
	stats := &AlertStats{
		ByLevel: map[string]int64{
			"critical": 0,
			"high":     0,
			"medium":   0,
			"low":      0,
			"info":     0,
		},
		Trend: []AlertTrendPoint{},
	}

	// 各级别统计
	type levelCount struct {
		Level string `gorm:"column:level"`
		Count int64  `gorm:"column:count"`
	}
	var levelCounts []levelCount
	db.Raw("SELECT level, COUNT(*) as count FROM mon_agent_alerts GROUP BY level").Scan(&levelCounts)
	for _, lc := range levelCounts {
		stats.ByLevel[lc.Level] = lc.Count
		stats.Total += lc.Count
	}

	// 已确认/未确认/已解决/活跃
	db.Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE acknowledged = 1").Scan(&stats.Acknowledged)
	db.Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE acknowledged = 0").Scan(&stats.Unacknowledged)
	db.Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE resolved_at IS NOT NULL").Scan(&stats.Resolved)
	db.Raw("SELECT COUNT(*) FROM mon_agent_alerts WHERE resolved_at IS NULL").Scan(&stats.Active)

	// 最近7天趋势
	type trendRow struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:count"`
	}
	var trend []trendRow
	db.Raw(`
		SELECT DATE_FORMAT(first_seen, '%Y-%m-%d') as date, COUNT(*) as count
		FROM mon_agent_alerts
		WHERE first_seen >= DATE_SUB(NOW(), INTERVAL 7 DAY)
		GROUP BY DATE_FORMAT(first_seen, '%Y-%m-%d')
		ORDER BY date ASC
	`).Scan(&trend)

	for _, t := range trend {
		stats.Trend = append(stats.Trend, AlertTrendPoint{Date: t.Date, Count: t.Count})
	}

	return stats, nil
}

// PersistMetric 持久化指标数据结构
type PersistMetric struct {
	ServerID   uint
	MetricType string
	MetricData []byte
	ReportTime time.Time
}

// MetricsPersister 指标数据批量写入器
type MetricsPersister struct {
	persistQueue chan *PersistMetric
	isRunning    bool
	mu           sync.Mutex
}

// 全局批量写入器实例
var metricsPersister *MetricsPersister

// GetMetricsPersister 获取批量写入器实例
func GetMetricsPersister() *MetricsPersister {
	if metricsPersister == nil {
		metricsPersister = &MetricsPersister{
			persistQueue: make(chan *PersistMetric, 5000),
			isRunning:    false,
		}
	}
	return metricsPersister
}

// StartPersistWorker 启动批量写入worker
func (p *MetricsPersister) StartPersistWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return
	}
	p.isRunning = true

	logger.Info("批量指标写入器已启动")

	ticker := time.NewTicker(30 * time.Second)
	batch := make([]*PersistMetric, 0, 1000)

	go func() {
		for {
			select {
			case metric := <-p.persistQueue:
				batch = append(batch, metric)
				// 达到批次大小立即写入
				if len(batch) >= 1000 {
					p.persistBatch(batch)
					batch = make([]*PersistMetric, 0, 1000)
				}

			case <-ticker.C:
				// 定时写入
				if len(batch) > 0 {
					p.persistBatch(batch)
					batch = make([]*PersistMetric, 0, 1000)
				}
			}
		}
	}()
}

// persistBatch 批量写入指标数据
func (p *MetricsPersister) persistBatch(batch []*PersistMetric) {
	if len(batch) == 0 {
		return
	}

	startTime := time.Now()
	logger.Debug("开始批量写入指标", zap.Int("count", len(batch)))

	// 使用事务批量写入
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, metric := range batch {
			query := `INSERT INTO mon_agent_metrics (server_id, metric_type, metric_data, report_time, received_at)
			             VALUES (?, ?, ?, ?, NOW())`
			if err := tx.Exec(query, metric.ServerID, metric.MetricType, string(metric.MetricData), metric.ReportTime).Error; err != nil {
				logger.Warn("批量写入指标失败",
					zap.Uint("serverID", metric.ServerID),
					zap.String("metricType", metric.MetricType),
					zap.Error(err))
				// 继续写入其他记录
			}
		}
		return nil
	})

	duration := time.Since(startTime)
	if err != nil {
		logger.Warn("批量指标写入事务失败", zap.Error(err), zap.Duration("duration", duration))
	} else {
		logger.Info("批量指标写入完成",
			zap.Int("count", len(batch)),
			zap.Duration("duration", duration),
			zap.Float64("avg_latency", float64(duration.Milliseconds())/float64(len(batch))))
	}
}

// EnqueueMetric 将指标数据放入持久化队列
func (p *MetricsPersister) EnqueueMetric(serverID uint, metricType string, metricData []byte, reportTime time.Time) {
	metric := &PersistMetric{
		ServerID:   serverID,
		MetricType: metricType,
		MetricData: metricData,
		ReportTime: reportTime,
	}

	select {
	case p.persistQueue <- metric:
		// 成功放入队列
	default:
		// 队列满，记录警告
		logger.Warn("持久化队列已满，丢弃指标数据",
			zap.Uint("serverID", serverID),
			zap.String("metricType", metricType))
	}
}
