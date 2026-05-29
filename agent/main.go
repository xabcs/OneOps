package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

const version = "1.0.0"

var (
	port         = flag.Int("port", 9100, "Agent 监听端口")
	heartbeatURL = flag.String("heartbeat-url", "", "OneOps 后端心跳接收地址")
	heartbeatTok = flag.String("heartbeat-token", "", "心跳鉴权 token")
)

// Metrics 指标结构
type Metrics struct {
	CPUUsage       float64         `json:"cpuUsage"`
	MemoryUsage    float64         `json:"memoryUsage"`
	DiskUsage      float64         `json:"diskUsage"`
	DiskPartitions []DiskPartition `json:"diskPartitions,omitempty"`
	Load5          float64         `json:"load5"`
	ProcessNum     int             `json:"processNum"`
	Version        string          `json:"version"`
}

// DiskPartition 磁盘分区信息
type DiskPartition struct {
	Mount string  `json:"mount"`
	Usage float64 `json:"usage"`
}

// HeartbeatPayload 心跳上报结构
type HeartbeatPayload struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	PID      int    `json:"pid"`
	Port     int    `json:"port"`
	Version  string `json:"version"`
	Token    string `json:"token"`
}

// metricsCache 后台持续采样并缓存最新指标
type metricsCache struct {
	mu      sync.RWMutex
	current Metrics
}

var cache = &metricsCache{}

// extMetricsCache 扩展指标缓存
type extMetricsCache struct {
	mu      sync.RWMutex
	current ExtendedMetrics
}

var extCache = &extMetricsCache{}

// startSampler 每 60 秒采样一次 CPU/内存/磁盘/load，cpu.Percent(0,...) 计算
// 两次调用之间的增量，精度远高于单次快照。
// 启动时先做一次"热身"调用（结果丢弃），再等 3 秒后才正式采样，
// 避免第一次读数恒为 0。
func startSampler() {
	// 热身：建立基准值
	cpu.Percent(0, false) //nolint
	time.Sleep(3 * time.Second)

	sample := func() {
		m := Metrics{}

		// CPU：interval=0 表示取上次调用以来的增量，需后台持续调用才准确
		if percents, err := cpu.Percent(0, false); err == nil && len(percents) > 0 {
			m.CPUUsage = percents[0]
		}

		if memInfo, err := mem.VirtualMemory(); err == nil {
			m.MemoryUsage = memInfo.UsedPercent
		}

		// 采集所有磁盘分区信息
		var maxDiskUsage float64
		if parts, err := disk.Partitions(false); err == nil {
			for _, p := range parts {
				// 过滤临时文件系统
				if p.Fstype == "tmpfs" || p.Fstype == "devtmpfs" || p.Fstype == "overlay" || p.Fstype == "squashfs" || p.Fstype == "shm" {
					continue
				}
				if u, err := disk.Usage(p.Mountpoint); err == nil {
					partition := DiskPartition{
						Mount: p.Mountpoint,
						Usage: u.UsedPercent,
					}
					m.DiskPartitions = append(m.DiskPartitions, partition)
					if u.UsedPercent > maxDiskUsage {
						maxDiskUsage = u.UsedPercent
					}
				}
			}
			// 使用最大使用率作为整体磁盘使用率
			m.DiskUsage = maxDiskUsage
		}

		if avg, err := load.Avg(); err == nil {
			m.Load5 = avg.Load5
		}

		if procs, err := process.Processes(); err == nil {
			m.ProcessNum = len(procs)
		}

		// 添加版本号
		m.Version = version

		cache.mu.Lock()
		cache.current = m
		cache.mu.Unlock()
	}

	// 立刻采一次，填充缓存
	sample()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		sample()
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

func sendHeartbeat() {
	if *heartbeatURL == "" {
		return
	}
	hostname, _ := os.Hostname()
	payload := HeartbeatPayload{
		Hostname: hostname,
		IP:       getLocalIP(),
		PID:      os.Getpid(),
		Port:     *port,
		Version:  version,
		Token:    *heartbeatTok,
	}
	data, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(*heartbeatURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("心跳发送失败: %v", err)
		return
	}
	defer resp.Body.Close()
}

func main() {
	flag.Parse()

	// 后台持续采样（必须早于 HTTP 服务启动）
	go startSampler()

	// 后台采集扩展指标（每2分钟采集一次）
	go func() {
		// 立即采一次
		extMetrics := CollectExtendedMetrics()
		extCache.mu.Lock()
		extCache.current = extMetrics
		extCache.mu.Unlock()

		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			extMetrics := CollectExtendedMetrics()
			extCache.mu.Lock()
			extCache.current = extMetrics
			extCache.mu.Unlock()
		}
	}()

	// 心跳
	go func() {
		sendHeartbeat()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			sendHeartbeat()
		}
	}()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		cache.mu.RLock()
		m := cache.current
		cache.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	})

	http.HandleFunc("/extended-metrics", func(w http.ResponseWriter, r *http.Request) {
		extCache.mu.RLock()
		m := extCache.current
		extCache.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("OneOps Agent %s 启动，监听 %s", version, addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Agent 启动失败: %v", err)
	}
}
