package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ============================================
// 扩展指标数据结构定义
// ============================================

// ExtendedMetrics 完整的扩展指标结构
type ExtendedMetrics struct {
	Performance  PerformanceMetrics  `json:"performance"`
	SystemInfo   SystemInfo          `json:"systemInfo,omitempty"`
	HardwareInfo HardwareInfo        `json:"hardwareInfo,omitempty"`
	ServiceStatus ServiceStatus      `json:"serviceStatus,omitempty"`
	ProcessInfo  ProcessInfo         `json:"processInfo,omitempty"`
	NetworkConfig NetworkConfig      `json:"networkConfig,omitempty"`
	SecurityInfo SecurityInfo        `json:"securityInfo,omitempty"`
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
}

// CPUMetrics CPU 详细指标
type CPUMetrics struct {
	UsagePercent float64 `json:"usagePercent"`
	User         float64 `json:"user"`
	System       float64 `json:"system"`
	Idle         float64 `json:"idle"`
	Iowait       float64 `json:"iowait"`
	Cores        int     `json:"cores"`
	Mhz          float64 `json:"mhz"`
}

// MemoryMetrics 内存详细指标
type MemoryMetrics struct {
	Total           uint64  `json:"total"`
	Used            uint64  `json:"used"`
	Free            uint64  `json:"free"`
	UsedPercent     float64 `json:"usedPercent"`
	Available       uint64  `json:"available"`
	Buffers         uint64  `json:"buffers"`
	Cached          uint64  `json:"cached"`
	TotalVirtual    uint64  `json:"totalVirtual,omitempty"`
	UsedVirtual     uint64  `json:"usedVirtual,omitempty"`
	UsedVirtualPercent float64 `json:"usedVirtualPercent,omitempty"`
}

// DiskMetrics 磁盘详细指标
type DiskMetrics struct {
	Total           uint64          `json:"total"`
	Used            uint64          `json:"used"`
	Free            uint64          `json:"free"`
	UsedPercent     float64         `json:"usedPercent"`
	InodesTotal     uint64          `json:"inodesTotal,omitempty"`
	InodesUsed      uint64          `json:"inodesUsed,omitempty"`
	InodesUsedPercent float64       `json:"inodesUsedPercent,omitempty"`
	Partitions      []PartitionInfo `json:"partitions,omitempty"`
	IO              *DiskIO         `json:"io,omitempty"`
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

// DiskIO 磁盘 IO 统计
type DiskIO struct {
	ReadBytes  uint64  `json:"readBytes"`
	WriteBytes uint64  `json:"writeBytes"`
	ReadCount  uint64  `json:"readCount"`
	WriteCount uint64  `json:"writeCount"`
	ReadTime   uint64  `json:"readTime"`
	WriteTime  uint64  `json:"writeTime"`
	// P1: 详细 IO 统计
	ReadIOPS        float64 `json:"readIOPS,omitempty"`        // 读 IOPS
	WriteIOPS       float64 `json:"writeIOPS,omitempty"`       // 写 IOPS
	ReadThroughput  float64 `json:"readThroughput,omitempty"`  // 读吞吐量 (MB/s)
	WriteThroughput float64 `json:"writeThroughput,omitempty"` // 写吞吐量 (MB/s)
	Await           float64 `json:"await,omitempty"`           // 平均等待时间 (ms)
	QueueDepth      float64 `json:"queueDepth,omitempty"`      // 队列深度
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	Interfaces []InterfaceStats `json:"interfaces,omitempty"`
	Connections ConnectionStats `json:"connections"`
	TCP        *TCPStats        `json:"tcp,omitempty"`
}

// InterfaceStats 网卡统计
type InterfaceStats struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
	Errin       uint64 `json:"errin"`
	Errout      uint64 `json:"errout"`
	Dropin      uint64 `json:"dropin"`
	Dropout     uint64 `json:"dropout"`
	Speed       uint64 `json:"speed,omitempty"`
}

// ConnectionStats 连接统计
type ConnectionStats struct {
	Established int `json:"established"`
	TimeWait    int `json:"timeWait"`
	CloseWait   int `json:"closeWait"`
	Listen      int `json:"listen"`
}

// TCPStats TCP 统计
type TCPStats struct {
	ActiveOpens  uint64 `json:"activeOpens"`
	PassiveOpens uint64 `json:"passiveOpens"`
	CurrEstab    uint64 `json:"currEstab"`
	InSegs       uint64 `json:"inSegs"`
	OutSegs      uint64 `json:"outSegs"`
}

// LoadMetrics 负载指标
type LoadMetrics struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname string            `json:"hostname"`
	OS       OSInfo            `json:"os"`
	Uptime   uint64            `json:"uptime"`
	BootTime uint64            `json:"bootTime,omitempty"`
	Timezone string            `json:"timezone,omitempty"`
	TimezoneOffsetSec int      `json:"timezoneOffsetSec,omitempty"`
}

// OSInfo 操作系统信息
type OSInfo struct {
	Family            string `json:"family,omitempty"`
	Platform          string `json:"platform,omitempty"`
	PlatformVersion   string `json:"platformVersion,omitempty"`
	KernelVersion     string `json:"kernelVersion,omitempty"`
	KernelArch        string `json:"kernelArch,omitempty"`
	VirtualizationSystem string `json:"virtualizationSystem,omitempty"`
	VirtualizationRole string `json:"virtualizationRole,omitempty"`
}

// HardwareInfo 硬件信息
type HardwareInfo struct {
	CPU     CPUInfo         `json:"cpu"`
	Memory  MemoryHardware  `json:"memory"`
	Disk    []DiskDevice    `json:"disk,omitempty"`
	Network []NetworkDevice `json:"network,omitempty"`
}

// CPUInfo CPU 详细信息
type CPUInfo struct {
	Vendor     string `json:"vendor,omitempty"`
	Model      string `json:"model,omitempty"`
	Cores      int    `json:"cores"`
	Threads    int    `json:"threads"`
	Mhz        float64 `json:"mhz"`
	CacheSize  int    `json:"cacheSize,omitempty"`
	Family     string `json:"family,omitempty"`
	ModelNum   string `json:"modelNum,omitempty"`
	Stepping   string `json:"stepping,omitempty"`
}

// MemoryHardware 内存硬件信息
type MemoryHardware struct {
	Total uint64           `json:"total"`
	Slots []MemorySlot     `json:"slots,omitempty"`
}

// MemorySlot 内存插槽信息
type MemorySlot struct {
	Slot    string `json:"slot,omitempty"`
	Type    string `json:"type,omitempty"`
	Size    uint64 `json:"size,omitempty"`
	Speed   int    `json:"speed,omitempty"`
	Vendor  string `json:"vendor,omitempty"`
	Serial  string `json:"serial,omitempty"`
}

// DiskDevice 磁盘设备信息
type DiskDevice struct {
	Name        string `json:"name,omitempty"`
	Vendor      string `json:"vendor,omitempty"`
	Model       string `json:"model,omitempty"`
	Serial      string `json:"serial,omitempty"`
	Size        uint64 `json:"size,omitempty"`
	Type        string `json:"type,omitempty"`
	Firmware    string `json:"firmware,omitempty"`    // P1: 固件版本
	SmartStatus string `json:"smartStatus,omitempty"` // P1: SMART 状态
	Temperature int    `json:"temperature,omitempty"`
}

// NetworkDevice 网卡设备信息
type NetworkDevice struct {
	Name    string `json:"name,omitempty"`
	MAC     string `json:"mac,omitempty"`
	Vendor  string `json:"vendor,omitempty"`
	Driver  string `json:"driver,omitempty"`
	Speed   uint64 `json:"speed,omitempty"`
	Duplex  string `json:"duplex,omitempty"`
	MTU     int    `json:"mtu,omitempty"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	SystemdServices []SystemdService `json:"systemdServices,omitempty"`
	ListenPorts     []ListenPort     `json:"listenPorts,omitempty"`
}

// SystemdService systemd 服务状态
type SystemdService struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	SubStatus   string `json:"subStatus"`
	Loaded      string `json:"loaded"`
	ActiveState string `json:"activeState"`
	SubState    string `json:"subState"`
	Description string `json:"description,omitempty"`
}

// ListenPort 监听端口
type ListenPort struct {
	Port    int    `json:"port"`
	Protocol string `json:"protocol"`
	Address string `json:"address"`
	Process string `json:"process,omitempty"`
	PID     int    `json:"pid,omitempty"`
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	Total int        `json:"total"`
	Top   []TopProcess `json:"top"`
}

// TopProcess Top 进程信息
type TopProcess struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float32 `json:"memoryPercent"`
	MemoryBytes   uint64  `json:"memoryBytes"`
	Status        string  `json:"status,omitempty"`
	Username      string  `json:"username,omitempty"`
	NumThreads    int     `json:"numThreads,omitempty"`
	CreateTime    int64   `json:"createTime,omitempty"`
	Cmdline       string  `json:"cmdline,omitempty"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
	Routes     []Route           `json:"routes,omitempty"`
	DNS        []string          `json:"dns,omitempty"`
	Hosts      []HostEntry       `json:"hosts,omitempty"`
}

// NetworkInterface 网络接口配置
type NetworkInterface struct {
	Name        string         `json:"name"`
	HardwareAddr string        `json:"hardwareAddr,omitempty"`
	MTU         int            `json:"mtu,omitempty"`
	Flags       []string       `json:"flags,omitempty"`
	Addrs       []InterfaceAddr `json:"addrs,omitempty"`
}

// InterfaceAddr 接口地址
type InterfaceAddr struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Family  string `json:"family"`
}

// Route 路由条目
type Route struct {
	Destination string   `json:"destination"`
	Gateway     string   `json:"gateway,omitempty"`
	Iface       string   `json:"iface"`
	Flags       []string `json:"flags,omitempty"`
}

// HostEntry hosts 文件条目
type HostEntry struct {
	IP    string   `json:"ip"`
	Names []string `json:"names"`
}

// SecurityInfo 安全信息
type SecurityInfo struct {
	SSH       SSHConfig       `json:"ssh,omitempty"`
	Firewall  FirewallConfig  `json:"firewall,omitempty"`
	SELinux   SELinuxConfig   `json:"selinux,omitempty"`
	Users     UserConfig      `json:"users,omitempty"`
	Sudoers   []string        `json:"sudoers,omitempty"`
	LastLogin []LoginEntry    `json:"lastLogin,omitempty"`
}

// SSHConfig SSH 配置
type SSHConfig struct {
	ConfigFile             string `json:"configFile,omitempty"`
	Port                   int    `json:"port,omitempty"`
	PermitRootLogin        string `json:"permitRootLogin,omitempty"`
	PasswordAuthentication string `json:"passwordAuthentication,omitempty"`
	PubkeyAuthentication   string `json:"pubkeyAuthentication,omitempty"`
	MaxAuthTries           int    `json:"maxAuthTries,omitempty"`
}

// FirewallConfig 防火墙配置
type FirewallConfig struct {
	Backend string        `json:"backend,omitempty"`
	Status  string        `json:"status,omitempty"`
	Zones   []FirewallZone `json:"zones,omitempty"`
}

// FirewallZone 防火墙区域
type FirewallZone struct {
	Name     string   `json:"name,omitempty"`
	Services []string `json:"services,omitempty"`
	Ports    []string `json:"ports,omitempty"`
}

// SELinuxConfig SELinux 配置
type SELinuxConfig struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode,omitempty"`
}

// UserConfig 用户配置
type UserConfig struct {
	Total  int         `json:"total"`
	Active int         `json:"active"`
	List   []UserInfo  `json:"list,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	Name   string `json:"name,omitempty"`
	UID    int    `json:"uid,omitempty"`
	GID    int    `json:"gid,omitempty"`
	Home   string `json:"home,omitempty"`
	Shell  string `json:"shell,omitempty"`
}

// LoginEntry 登录记录
type LoginEntry struct {
	User    string `json:"user,omitempty"`
	TTY     string `json:"tty,omitempty"`
	From    string `json:"from,omitempty"`
	Time    string `json:"time,omitempty"`
}

// ConfigChange 配置变更
type ConfigChange struct {
	File       string `json:"file,omitempty"`
	Checksum   string `json:"checksum,omitempty"`
	ChangeTime string `json:"changeTime,omitempty"`
	ChangeType string `json:"changeType,omitempty"`
}

// ============================================
// 扩展指标采集器
// ============================================

// ExtendedMetricsCache 扩展指标缓存
// 注意：变量定义在 main.go 中，此处仅作类型定义
type ExtendedMetricsCache struct {
	mu      sync.RWMutex
	current ExtendedMetrics
}

// collectPerformanceMetrics 采集性能指标
func collectPerformanceMetrics() PerformanceMetrics {
	var metrics PerformanceMetrics

	// CPU 指标
	if percents, err := cpu.Percent(0, false); err == nil && len(percents) > 0 {
		metrics.CPU.UsagePercent = percents[0]
	}
	if times, err := cpu.Times(false); err == nil && len(times) > 0 {
		total := times[0].Total()
		if total > 0 {
			metrics.CPU.User = times[0].User / total * 100
			metrics.CPU.System = times[0].System / total * 100
			metrics.CPU.Idle = times[0].Idle / total * 100
			metrics.CPU.Iowait = times[0].Iowait / total * 100
		}
	}
	if info, err := cpu.Info(); err == nil && len(info) > 0 {
		metrics.CPU.Cores = runtime.NumCPU()
		metrics.CPU.Mhz = info[0].Mhz
	}

	// 内存指标
	if memInfo, err := mem.VirtualMemory(); err == nil {
		metrics.Memory = MemoryMetrics{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			Free:        memInfo.Free,
			UsedPercent: memInfo.UsedPercent,
			Available:   memInfo.Available,
			Buffers:     memInfo.Buffers,
			Cached:      memInfo.Cached,
		}
	}
	if swapInfo, err := mem.SwapMemory(); err == nil {
		metrics.Memory.TotalVirtual = swapInfo.Total
		metrics.Memory.UsedVirtual = swapInfo.Used
		metrics.Memory.UsedVirtualPercent = swapInfo.UsedPercent
	}

	// 磁盘指标
	var maxDiskUsage float64
	var partitions []PartitionInfo
	if parts, err := disk.Partitions(false); err == nil {
		for _, p := range parts {
			// 过滤临时文件系统
			if p.Fstype == "tmpfs" || p.Fstype == "devtmpfs" || p.Fstype == "overlay" || p.Fstype == "squashfs" || p.Fstype == "shm" {
				continue
			}
			if u, err := disk.Usage(p.Mountpoint); err == nil {
				partition := PartitionInfo{
					Device:      p.Device,
					Mountpoint:  p.Mountpoint,
					Fstype:      p.Fstype,
					Total:       u.Total,
					Used:        u.Used,
					Free:        u.Free,
					UsedPercent: u.UsedPercent,
				}
				partitions = append(partitions, partition)
				if u.UsedPercent > maxDiskUsage {
					maxDiskUsage = u.UsedPercent
				}
			}
		}
	}
	metrics.Disk.Partitions = partitions
	metrics.Disk.UsedPercent = maxDiskUsage
	if len(partitions) > 0 {
		metrics.Disk.Total = partitions[0].Total
		metrics.Disk.Used = partitions[0].Used
		metrics.Disk.Free = partitions[0].Free
	}

	// 磁盘 IO
	if ioCounts, err := disk.IOCounters(); err == nil && len(ioCounts) > 0 {
		var totalIO DiskIO
		for _, io := range ioCounts {
			totalIO.ReadBytes += io.ReadBytes
			totalIO.WriteBytes += io.WriteBytes
			totalIO.ReadCount += io.ReadCount
			totalIO.WriteCount += io.WriteCount
			totalIO.ReadTime += io.ReadTime
			totalIO.WriteTime += io.WriteTime
		}
		metrics.Disk.IO = &totalIO
	}

	// 网络指标
	if netStats, err := net.IOCounters(true); err == nil {
		var interfaces []InterfaceStats
		for _, stat := range netStats {
			if stat.BytesRecv == 0 && stat.BytesSent == 0 {
				continue // 跳过无流量的接口
			}
			interfaces = append(interfaces, InterfaceStats{
				Name:        stat.Name,
				BytesSent:   stat.BytesSent,
				BytesRecv:   stat.BytesRecv,
				PacketsSent: stat.PacketsSent,
				PacketsRecv: stat.PacketsRecv,
				Errin:       stat.Errin,
				Errout:      stat.Errout,
				Dropin:      stat.Dropin,
				Dropout:     stat.Dropout,
			})
		}
		metrics.Network.Interfaces = interfaces
	}

	// 网络连接统计
	metrics.Network.Connections = collectConnectionStats()

	// TCP 统计
	metrics.Network.TCP = collectTCPStats()

	// 负载指标
	if avg, err := load.Avg(); err == nil {
		metrics.Load = LoadMetrics{
			Load1:  avg.Load1,
			Load5:  avg.Load5,
			Load15: avg.Load15,
		}
	}

	return metrics
}

// collectConnectionStats 采集网络连接统计
func collectConnectionStats() ConnectionStats {
	var stats ConnectionStats

	// 读取 /proc/net/tcp 统计连接状态
	if file, err := os.Open("/proc/net/tcp"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		estab, tw, cw := 0, 0, 0
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			state := fields[3]
			// 状态码: 01=ESTABLISHED, 06=TIME_WAIT, 09=CLOSE_WAIT
			switch state {
			case "01":
				estab++
			case "06":
				tw++
			case "09":
				cw++
			}
		}
		stats.Established = estab
		stats.TimeWait = tw
		stats.CloseWait = cw
	}

	// 统计监听端口
	if file, err := os.Open("/proc/net/tcp"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		listen := 0
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			if fields[3] == "0A" { // 0A = LISTEN
				listen++
			}
		}
		stats.Listen = listen
	}

	return stats
}

// collectTCPStats 采集 TCP 统计信息
func collectTCPStats() *TCPStats {
	// 读取 /proc/net/snmp 中的 TCP 统计
	if file, err := os.Open("/proc/net/snmp"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		inTcpSection := false
		var tcpStats TCPStats

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Tcp:") {
				inTcpSection = true
				continue
			}
			if inTcpSection {
				if strings.TrimSpace(line) == "" {
					break
				}
				fields := strings.Fields(line)
				if len(fields) >= 12 {
					// Tcp: RtoAlgorithm RtoMin RtoMax MaxConn ActiveOpens PassiveOpens AttemptFails EstabResets CurrEstab InSegs OutSegs RetransSegs InErrs OutRsts
					activeOpens, _ := strconv.ParseUint(fields[5], 10, 64)
					passiveOpens, _ := strconv.ParseUint(fields[6], 10, 64)
					currEstab, _ := strconv.ParseUint(fields[9], 10, 64)
					inSegs, _ := strconv.ParseUint(fields[10], 10, 64)
					outSegs, _ := strconv.ParseUint(fields[11], 10, 64)

					tcpStats.ActiveOpens = activeOpens
					tcpStats.PassiveOpens = passiveOpens
					tcpStats.CurrEstab = currEstab
					tcpStats.InSegs = inSegs
					tcpStats.OutSegs = outSegs

					return &tcpStats
				}
			}
		}
	}

	return nil
}

// collectSystemInfo 采集系统信息
func collectSystemInfo() SystemInfo {
	var info SystemInfo

	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	if hostInfo, err := host.Info(); err == nil {
		info.OS = OSInfo{
			Family:            hostInfo.OS,
			Platform:          hostInfo.Platform,
			PlatformVersion:   hostInfo.PlatformVersion,
			KernelVersion:     hostInfo.KernelVersion,
			KernelArch:        hostInfo.KernelArch,
			VirtualizationSystem: hostInfo.VirtualizationSystem,
			VirtualizationRole:   hostInfo.VirtualizationRole,
		}
		info.Uptime = hostInfo.Uptime
		info.BootTime = hostInfo.BootTime
	}

	return info
}

// collectHardwareInfo 采集硬件信息
func collectHardwareInfo() HardwareInfo {
	var info HardwareInfo

	// CPU 信息
	if cpuInfos, err := cpu.Info(); err == nil && len(cpuInfos) > 0 {
		ci := cpuInfos[0]
		info.CPU = CPUInfo{
			Vendor:    ci.VendorID,
			Model:     ci.ModelName,
			Cores:     runtime.NumCPU(),
			Threads:   runtime.NumCPU(),
			Mhz:       ci.Mhz,
			Family:    fmt.Sprintf("%d", ci.Family),
			ModelNum:  fmt.Sprintf("%d", ci.Model),
			Stepping:  fmt.Sprintf("%d", ci.Stepping),
		}
	}

	// 内存信息
	if memInfo, err := mem.VirtualMemory(); err == nil {
		info.Memory.Total = memInfo.Total
		// 尝试从 dmidecode 读取内存插槽信息（需要 root 权限）
		info.Memory.Slots = collectMemorySlots()
	}

	// 磁盘信息
	info.Disk = collectDiskDevices()

	// 网卡信息
	info.Network = collectNetworkDevices()

	return info
}

// collectMemorySlots 采集内存插槽信息
func collectMemorySlots() []MemorySlot {
	// 使用 dmidecode 命令读取内存信息
	cmd := exec.Command("dmidecode", "-t", "memory")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var slots []MemorySlot
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var currentSlot *MemorySlot

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Memory Device") {
			if currentSlot != nil && currentSlot.Size > 0 {
				slots = append(slots, *currentSlot)
			}
			currentSlot = &MemorySlot{}
		} else if currentSlot != nil {
			if strings.HasPrefix(line, "Locator:") {
				currentSlot.Slot = strings.TrimSpace(stringsPrefix(line, "Locator:"))
			} else if strings.HasPrefix(line, "Type:") {
				currentSlot.Type = strings.TrimSpace(stringsPrefix(line, "Type:"))
			} else if strings.HasPrefix(line, "Size:") {
				sizeStr := strings.TrimSpace(stringsPrefix(line, "Size:"))
				if sizeStr != "No Module Installed" {
					// 解析大小，如 "8192 MB"
					if parts := strings.Fields(sizeStr); len(parts) >= 2 {
						if size, err := strconv.ParseUint(parts[0], 10, 64); err == nil {
							switch strings.ToUpper(parts[1]) {
							case "MB", "MIB":
								currentSlot.Size = size * 1024 * 1024
							case "GB", "GIB":
								currentSlot.Size = size * 1024 * 1024 * 1024
							}
						}
					}
				}
			} else if strings.HasPrefix(line, "Speed:") {
				speedStr := strings.TrimSpace(stringsPrefix(line, "Speed:"))
				if speedStr != "Unknown" {
					if parts := strings.Fields(speedStr); len(parts) >= 1 {
						if speed, err := strconv.Atoi(strings.TrimSuffix(parts[0], "MT/s")); err == nil {
							currentSlot.Speed = speed
						}
					}
				}
			} else if strings.HasPrefix(line, "Manufacturer:") {
				currentSlot.Vendor = strings.TrimSpace(stringsPrefix(line, "Manufacturer:"))
			} else if strings.HasPrefix(line, "Serial Number:") {
				currentSlot.Serial = strings.TrimSpace(stringsPrefix(line, "Serial Number:"))
			}
		}
	}

	if currentSlot != nil && currentSlot.Size > 0 {
		slots = append(slots, *currentSlot)
	}

	return slots
}

// collectDiskDevices 采集磁盘设备信息
func collectDiskDevices() []DiskDevice {
	var devices []DiskDevice

	// 使用 lsblk 读取磁盘信息（包括固件版本）
	cmd := exec.Command("lsblk", "-d", "-o", "NAME,TYPE,SIZE,MODEL,SERIAL,ROTA,FIRMWARE,UUID")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	// 跳过标题行
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		if fields[1] == "disk" {
			deviceName := "/dev/" + fields[0]
			device := DiskDevice{
				Name:   deviceName,
				Model:  fields[3],
				Serial: fields[4],
			}

			// 判断磁盘类型（HDD/SSD）
			if len(fields) > 5 && fields[5] == "1" {
				device.Type = "hdd"
			} else {
				device.Type = "ssd"
			}

			// 获取固件版本
			if len(fields) > 6 {
				device.Firmware = fields[6]
			}

			// 使用 smartctl 获取 SMART 状态
			device.SmartStatus = getSmartStatus(deviceName)

			devices = append(devices, device)
		}
	}

	return devices
}

// getSmartStatus 获取磁盘 SMART 状态
func getSmartStatus(deviceName string) string {
	// 尝试使用 smartctl 获取 SMART 状态
	cmd := exec.Command("smartctl", "-H", deviceName)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// 解析输出，查找 SMART 总体状态
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "SMART overall-health self-assessment test result") {
			if strings.Contains(line, "PASSED") {
				return "ok"
			} else if strings.Contains(line, "FAILED") {
				return "failed"
			}
		}
	}

	// 备用方案：查找其他状态指示
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "SMART Status:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return "unknown"
}



// getDiskQueueDepth 获取磁盘队列深度
func getDiskQueueDepth() float64 {
	// 从 /proc/diskstats 读取队列深度信息
	// 格式:   major minor device reads reads_merged reads_sectors ms_spent_reading writes writes_merged writes_sectors ms_spent_writing ios_in_progress ms_spent_ios ms_weighted_ios_spent
	// ios_in_progress (第 11 列) 就是当前队列深度
	if file, err := os.Open("/proc/diskstats"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		totalQueueDepth := 0.0
		deviceCount := 0

		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) >= 12 {
				// 第 11 列（索引 10）是 ios_in_progress，即队列深度
				if queueDepth, err := strconv.ParseFloat(fields[11], 64); err == nil {
					// 只统计物理磁盘设备，跳过分区
					if fields[1] == "0" {
						totalQueueDepth += queueDepth
						deviceCount++
					}
				}
			}
		}

		if deviceCount > 0 {
			return totalQueueDepth / float64(deviceCount)
		}
	}

	return 0
}

// collectNetworkDevices 采集网卡设备信息
func collectNetworkDevices() []NetworkDevice {
	var devices []NetworkDevice

	// 读取 /sys/class/net 下的网卡信息
	if entries, err := os.ReadDir("/sys/class/net"); err == nil {
		for _, entry := range entries {
			ifaceName := entry.Name()
			if ifaceName == "lo" {
				continue
			}

			device := NetworkDevice{Name: ifaceName}

			// 读取 MAC 地址
			if addr, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/address", ifaceName)); err == nil {
				device.MAC = strings.TrimSpace(string(addr))
			}

			// 读取 MTU
			if mtu, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/mtu", ifaceName)); err == nil {
				if mtuVal, err := strconv.Atoi(strings.TrimSpace(string(mtu))); err == nil {
					device.MTU = mtuVal
				}
			}

			// 读取速度
			if speed, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/speed", ifaceName)); err == nil {
				if speedVal, err := strconv.ParseUint(strings.TrimSpace(string(speed)), 10, 64); err == nil {
					device.Speed = speedVal * 1000 * 1000 // 转换为 bps
				}
			}

			// 读取 duplex
			if duplex, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/duplex", ifaceName)); err == nil {
				device.Duplex = strings.TrimSpace(string(duplex))
			}

			devices = append(devices, device)
		}
	}

	return devices
}

// collectServiceStatus 采集服务状态
func collectServiceStatus() ServiceStatus {
	var status ServiceStatus

	// 采集 systemd 服务状态
	status.SystemdServices = collectSystemdServices()

	// 采集监听端口
	status.ListenPorts = collectListenPorts()

	return status
}

// collectSystemdServices 采集 systemd 服务状态
func collectSystemdServices() []SystemdService {
	// 使用 systemctl 列出所有服务
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var services []SystemdService
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// 跳过标题行
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "UNIT") && strings.Contains(line, "LOAD") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			// 只返回常见服务，避免返回过多
			serviceName := fields[0]
			if strings.Contains(serviceName, "ssh") ||
			   strings.Contains(serviceName, "nginx") ||
			   strings.Contains(serviceName, "mysql") ||
			   strings.Contains(serviceName, "docker") ||
			   strings.Contains(serviceName, "cron") ||
			   strings.Contains(serviceName, "rsyslog") {

				services = append(services, SystemdService{
					Name:        serviceName,
					Status:      fields[1],
					SubStatus:   fields[2],
					Loaded:      fields[1],
					ActiveState: fields[2],
					SubState:    fields[3],
				})
			}
		}
	}

	return services
}

// collectListenPorts 采集监听端口
func collectListenPorts() []ListenPort {
	var ports []ListenPort

	// 使用 ss 命令获取监听端口
	cmd := exec.Command("ss", "-tuln")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	// 跳过标题行
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		protocol := fields[0]
		localAddress := fields[4]

		// 解析地址和端口
		if strings.Contains(localAddress, ":") {
			parts := strings.Split(localAddress, ":")
			if len(parts) >= 2 {
				portStr := parts[len(parts)-1]
				if port, err := strconv.Atoi(portStr); err == nil {
					// 获取监听进程
					var processName string
					var pid int

					// 使用 ss -tulpn 获取进程信息
					cmd2 := exec.Command("ss", "-tulpn")
					if output2, err2 := cmd2.Output(); err2 == nil {
						scanner2 := bufio.NewScanner(bytes.NewReader(output2))
						for scanner2.Scan() {
							line2 := scanner2.Text()
							if strings.Contains(line2, fmt.Sprintf(":%d", port)) {
								// 提取进程名和 PID
								if idx := strings.Index(line2, "pid="); idx > 0 {
									pidStr := ""
									for i := idx + 4; i < len(line2); i++ {
										if line2[i] >= '0' && line2[i] <= '9' {
											pidStr += string(line2[i])
										} else {
											break
										}
									}
									if p, err := strconv.Atoi(pidStr); err == nil {
										pid = p
									}
								}
								if idx := strings.Index(line2, "name=\""); idx > 0 {
									nameStart := idx + 6
									if nameEnd := strings.Index(line2[nameStart:], "\""); nameEnd > 0 {
										processName = line2[nameStart : nameStart+nameEnd]
									}
								}
								break
							}
						}
					}

					ports = append(ports, ListenPort{
						Port:     port,
						Protocol: protocol,
						Address:  localAddress,
						Process:  processName,
						PID:      pid,
					})
				}
			}
		}
	}

	return ports
}

// collectProcessInfo 采集进程信息（Top 20）
func collectProcessInfo() ProcessInfo {
	var info ProcessInfo

	if procs, err := process.Processes(); err == nil {
		info.Total = len(procs)

		// 按 CPU 使用率排序
		type processWithCPU struct {
			pid       int32
			cpuPercent float64
		}
		var procList []processWithCPU

		for _, p := range procs {
			if cpuPercent, err := p.CPUPercent(); err == nil {
				procList = append(procList, processWithCPU{
					pid:         p.Pid,
					cpuPercent:  cpuPercent,
				})
			}
		}

		// 排序并取 Top 20
		for i := 0; i < len(procList)-1; i++ {
			for j := i + 1; j < len(procList); j++ {
				if procList[i].cpuPercent < procList[j].cpuPercent {
					procList[i], procList[j] = procList[j], procList[i]
				}
			}
		}

		topCount := 20
		if len(procList) < topCount {
			topCount = len(procList)
		}

		for i := 0; i < topCount; i++ {
			p, err := process.NewProcess(procList[i].pid)
			if err != nil {
				continue
			}

			name, _ := p.Name()
			memPercent, _ := p.MemoryPercent()
			memInfo, _ := p.MemoryInfo()
			status, _ := p.Status()
			username, _ := p.Username()
			numThreads, _ := p.NumThreads()
			createTime, _ := p.CreateTime()
			cmdline, _ := p.Cmdline()

			info.Top = append(info.Top, TopProcess{
				PID:           procList[i].pid,
				Name:          name,
				CPUPercent:    procList[i].cpuPercent,
				MemoryPercent: memPercent,
				MemoryBytes:   memInfo.RSS,
				Status:        status[0],
				Username:      username,
				NumThreads:    int(numThreads),
				CreateTime:    createTime,
				Cmdline:       cmdline,
			})
		}
	}

	return info
}

// collectNetworkConfig 采集网络配置
func collectNetworkConfig() NetworkConfig {
	var config NetworkConfig

	// 采集网卡信息
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			netIface := NetworkInterface{
				Name: iface.Name,
				MTU:  iface.MTU,
				HardwareAddr: iface.HardwareAddr,
			}

			for _, flag := range iface.Flags {
				netIface.Flags = append(netIface.Flags, string(flag))
			}

			for _, addr := range iface.Addrs {
				netIface.Addrs = append(netIface.Addrs, InterfaceAddr{
					IP:      addr.String(),
					Netmask: "",
					Family:  "",
				})
			}

			config.Interfaces = append(config.Interfaces, netIface)
		}
	}

	// 采集路由信息
	config.Routes = collectRoutes()

	// 采集 DNS 配置
	config.DNS = collectDNS()

	return config
}

// collectRoutes 采集路由表
func collectRoutes() []Route {
	var routes []Route

	// 读取 /proc/net/route
	cmd := exec.Command("ip", "route")
	output, err := cmd.Output()
	if err != nil {
		// 备用方案：读取 /proc/net/route
		return routes
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		route := Route{
			Destination: fields[0],
			Iface:       fields[len(fields)-1],
		}

		if len(fields) >= 2 && fields[1] != "via" {
			route.Gateway = fields[1]
		}

		routes = append(routes, route)
	}

	return routes
}

// collectDNS 采集 DNS 配置
func collectDNS() []string {
	var dns []string

	// 读取 /etc/resolv.conf
	if file, err := os.Open("/etc/resolv.conf"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "nameserver") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					dns = append(dns, fields[1])
				}
			}
		}
	}

	return dns
}

// collectSecurityInfo 采集安全信息
func collectSecurityInfo() SecurityInfo {
	var info SecurityInfo

	// SSH 配置
	info.SSH = collectSSHConfig()

	// 防火墙配置
	info.Firewall = collectFirewallConfig()

	// SELinux 配置
	info.SELinux = collectSELinuxConfig()

	// 用户信息
	info.Users = collectUserInfo()

	// sudoers 配置
	info.Sudoers = collectSudoers()

	// 登录历史
	info.LastLogin = collectLastLogin()

	return info
}

// collectSSHConfig 采集 SSH 配置
func collectSSHConfig() SSHConfig {
	var config SSHConfig
	config.ConfigFile = "/etc/ssh/sshd_config"

	if file, err := os.Open("/etc/ssh/sshd_config"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) >= 2 {
				switch fields[0] {
				case "Port":
					if port, err := strconv.Atoi(fields[1]); err == nil {
						config.Port = port
					}
				case "PermitRootLogin":
					config.PermitRootLogin = fields[1]
				case "PasswordAuthentication":
					config.PasswordAuthentication = fields[1]
				case "PubkeyAuthentication":
					config.PubkeyAuthentication = fields[1]
				case "MaxAuthTries":
					if tries, err := strconv.Atoi(fields[1]); err == nil {
						config.MaxAuthTries = tries
					}
				}
			}
		}
	}

	return config
}

// collectFirewallConfig 采集防火墙配置
func collectFirewallConfig() FirewallConfig {
	var config FirewallConfig

	// 检测防火墙后端
	if _, err := exec.LookPath("firewall-cmd"); err == nil {
		config.Backend = "firewalld"
		config.Status = "active"

		// 获取防火墙区域
		cmd := exec.Command("firewall-cmd", "--list-all-zones")
		if output, err := cmd.Output(); err == nil {
			// 解析区域信息
			scanner := bufio.NewScanner(bytes.NewReader(output))
			var currentZone *FirewallZone

			for scanner.Scan() {
				line := scanner.Text()

				if strings.HasPrefix(line, "") && strings.Contains(line, "(active)") {
					if currentZone != nil {
						config.Zones = append(config.Zones, *currentZone)
					}
					zonename := strings.TrimSpace(strings.TrimSuffix(line, " (active)"))
					currentZone = &FirewallZone{Name: zonename}
				} else if currentZone != nil {
					if strings.HasPrefix(line, "\tservices:") {
						services := strings.TrimSpace(stringsPrefix(line, "\tservices:"))
						currentZone.Services = strings.Fields(services)
					} else if strings.HasPrefix(line, "\tports:") {
						ports := strings.TrimSpace(stringsPrefix(line, "\tports:"))
						currentZone.Ports = strings.Fields(ports)
					}
				}
			}

			if currentZone != nil {
				config.Zones = append(config.Zones, *currentZone)
			}
		}
	} else if _, err := exec.LookPath("ufw"); err == nil {
		config.Backend = "ufw"
		cmd := exec.Command("ufw", "status")
		if output, err := cmd.Output(); err == nil {
			if strings.Contains(string(output), "active") {
				config.Status = "active"
			} else {
				config.Status = "inactive"
			}
		}
	}

	return config
}

// collectSELinuxConfig 采集 SELinux 配置
func collectSELinuxConfig() SELinuxConfig {
	var config SELinuxConfig

	// 读取 /etc/selinux/config
	if file, err := os.Open("/etc/selinux/config"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "SELINUX=") {
				mode := strings.TrimSpace(stringsPrefix(line, "SELINUX="))
				config.Enabled = mode != "disabled"
				config.Mode = mode
			}
		}
	}

	// 读取当前状态
	if cmd := exec.Command("getenforce"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			config.Mode = strings.TrimSpace(string(output))
			config.Enabled = config.Mode != "Disabled"
		}
	}

	return config
}

// collectUserInfo 采集用户信息
func collectUserInfo() UserConfig {
	var config UserConfig

	// 读取 /etc/passwd
	if file, err := os.Open("/etc/passwd"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Split(line, ":")
			if len(fields) >= 7 {
				uid, _ := strconv.Atoi(fields[2])
				gid, _ := strconv.Atoi(fields[3])

				// 只收集 UID >= 1000 的普通用户
				if uid >= 1000 && !strings.HasPrefix(fields[0], "_") {
					config.Total++
					if fields[6] != "/sbin/nologin" && fields[6] != "/bin/false" {
						config.Active++
					}

					config.List = append(config.List, UserInfo{
						Name:  fields[0],
						UID:   uid,
						GID:   gid,
						Home:  fields[5],
						Shell: fields[6],
					})
				}
			}
		}
	}

	return config
}

// collectSudoers 采集 sudoers 配置
func collectSudoers() []string {
	var sudoers []string

	// 读取 /etc/sudoers
	if file, err := os.Open("/etc/sudoers"); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				sudoers = append(sudoers, line)
			}
		}
	}

	return sudoers
}

// collectLastLogin 采集最近登录记录
func collectLastLogin() []LoginEntry {
	var entries []LoginEntry

	// 使用 last 命令获取登录历史
	cmd := exec.Command("last", "-n", "10", "-w", "-F")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "reboot") || strings.HasPrefix(line, "wtmp") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			entries = append(entries, LoginEntry{
				User: fields[0],
				TTY:  fields[1],
				From: fields[2],
				Time: strings.Join(fields[3:], " "),
			})
		}
	}

	return entries
}

// stringsPrefix 移除字符串前缀
func stringsPrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// CollectExtendedMetrics 采集完整的扩展指标
func CollectExtendedMetrics() ExtendedMetrics {
	var metrics ExtendedMetrics

	// 采集性能指标
	metrics.Performance = collectPerformanceMetrics()

	// 采集系统信息
	metrics.SystemInfo = collectSystemInfo()

	// 采集硬件信息
	metrics.HardwareInfo = collectHardwareInfo()

	// 采集服务状态
	metrics.ServiceStatus = collectServiceStatus()

	// 采集进程信息
	metrics.ProcessInfo = collectProcessInfo()

	// 采集网络配置
	metrics.NetworkConfig = collectNetworkConfig()

	// 采集安全信息
	metrics.SecurityInfo = collectSecurityInfo()

	// 设置采集时间和版本
	metrics.CollectedAt = time.Now().Format(time.RFC3339)
	metrics.Version = version

	return metrics
}
