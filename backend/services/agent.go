package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"oneops/backend/config"
	"oneops/backend/logger"
	"oneops/backend/models"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

const (
	agentInstallDir  = "/opt/oneops-agent"
	agentBinaryName  = "oneops-agent"
	agentServiceName = "oneops-agent"
	heartbeatTimeout = 3 * time.Minute
)

// agentBinaryDir 返回 Agent 预编译二进制所在目录，按优先级依次探测：
//  1. 可执行文件同级的 agent-binaries/（生产部署）
//  2. 工作目录下的 agent-binaries/（Air 开发模式，CWD = 项目根）
//  3. 可执行文件向上两级的 agent-binaries/（backend/tmp/main → 项目根）
func agentBinaryDir() string {
	candidates := []string{}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "agent-binaries"),
			filepath.Clean(filepath.Join(exeDir, "..", "..", "agent-binaries")),
		)
	}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "agent-binaries"))
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return "agent-binaries"
}

// agentMetricsResponse Agent /metrics 接口响应
type agentMetricsResponse struct {
	CPUUsage       float64               `json:"cpuUsage"`
	MemoryUsage    float64               `json:"memoryUsage"`
	DiskUsage      float64               `json:"diskUsage"`
	DiskPartitions models.DiskPartitions `json:"diskPartitions,omitempty"`
	Load5          float64               `json:"load5"`
	ProcessNum     int                   `json:"processNum"`
}

// AgentHeartbeatData Agent 心跳上报数据
type AgentHeartbeatData struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	PID      int    `json:"pid"`
	Port     int    `json:"port"`
	Version  string `json:"version"`
	Token    string `json:"token"`
}

// AgentService Agent 管理服务
type AgentService struct{}

// NewAgentService 创建 AgentService 实例
func NewAgentService() *AgentService {
	return &AgentService{}
}

// DeployAgent 通过 SSH 部署 Agent 到目标主机
func (s *AgentService) DeployAgent(serverID uint) error {
	server, err := s.loadServerForAgent(serverID)
	if err != nil {
		return err
	}

	sshClient, err := dialSSH(server)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer sshClient.Close()

	// 检测目标主机架构
	arch, _ := runSSHCommand(sshClient, "uname -m")
	binaryArch := "amd64"
	if arch == "aarch64" || arch == "arm64" {
		binaryArch = "arm64"
	}

	// 上传预编译二进制
	localBinary := fmt.Sprintf("%s/oneops-agent-linux-%s", agentBinaryDir(), binaryArch)
	remoteBinary := fmt.Sprintf("%s/%s", agentInstallDir, agentBinaryName)

	if err := runSSHCommand2(sshClient, fmt.Sprintf("mkdir -p %s", agentInstallDir)); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := scpFile(sshClient, localBinary, remoteBinary); err != nil {
		return fmt.Errorf("上传 Agent 二进制失败: %w", err)
	}

	if err := runSSHCommand2(sshClient, fmt.Sprintf("chmod +x %s", remoteBinary)); err != nil {
		return fmt.Errorf("设置执行权限失败: %w", err)
	}

	agentPort := server.AgentPort
	if agentPort == 0 {
		agentPort = 9100
	}

	// 获取后端外部访问地址（用于Agent心跳上报）
	cfg := config.GetConfig()
	heartbeatURL := fmt.Sprintf("%s/api/cmdb/agent/heartbeat", cfg.Server.GetExternalURL())

	serviceContent := fmt.Sprintf(`[Unit]
Description=OneOps Agent
After=network.target

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s/%s --port %d --heartbeat-url %s
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, agentInstallDir, agentInstallDir, agentBinaryName, agentPort, heartbeatURL)

	writeCmd := fmt.Sprintf("cat > /etc/systemd/system/%s.service << 'ONEOPS_EOF'\n%sONEOPS_EOF", agentServiceName, serviceContent)
	if err := runSSHCommand2(sshClient, writeCmd); err != nil {
		return fmt.Errorf("写入 systemd 配置失败: %w", err)
	}

	startCmds := fmt.Sprintf("systemctl daemon-reload && systemctl enable %s && systemctl restart %s", agentServiceName, agentServiceName)
	if err := runSSHCommand2(sshClient, startCmds); err != nil {
		return fmt.Errorf("启动 Agent 服务失败: %w", err)
	}

	db.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"agent_status":  "running",
		"agent_port":    agentPort,
		"agent_version": "1.0.0", // Agent 默认版本号
	})

	logger.Info("Agent 部署成功", zap.Uint("serverID", serverID), zap.String("arch", binaryArch))
	return nil
}

// RestartAgent 通过 SSH 重启 Agent 服务
func (s *AgentService) RestartAgent(serverID uint) error {
	server, err := s.loadServerForAgent(serverID)
	if err != nil {
		return err
	}

	sshClient, err := dialSSH(server)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer sshClient.Close()

	if err := runSSHCommand2(sshClient, fmt.Sprintf("systemctl restart %s", agentServiceName)); err != nil {
		return fmt.Errorf("重启 Agent 失败: %w", err)
	}

	updates := map[string]interface{}{"agent_status": "running"}
	if server.AgentVersion == "" {
		updates["agent_version"] = "1.0.0" // 如果版本为空，设置默认版本
	}
	db.Model(&models.Server{}).Where("id = ?", serverID).Updates(updates)
	logger.Info("Agent 重启成功", zap.Uint("serverID", serverID))
	return nil
}

// UninstallAgent 通过 SSH 卸载 Agent
func (s *AgentService) UninstallAgent(serverID uint) error {
	server, err := s.loadServerForAgent(serverID)
	if err != nil {
		return err
	}

	sshClient, err := dialSSH(server)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer sshClient.Close()

	cmds := fmt.Sprintf(
		"systemctl stop %s; systemctl disable %s; rm -f /etc/systemd/system/%s.service; rm -rf %s; systemctl daemon-reload",
		agentServiceName, agentServiceName, agentServiceName, agentInstallDir,
	)
	if err := runSSHCommand2(sshClient, cmds); err != nil {
		logger.Error("卸载 Agent SSH命令执行失败", zap.Uint("serverID", serverID), zap.Error(err))
		return fmt.Errorf("卸载 Agent 失败: %w", err)
	}

	db.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"agent_status":       "uninstalled",
		"agent_version":      "",
		"last_heartbeat_at":  nil,
		"metrics_updated_at": nil,
		"cpu_usage":          0,
		"memory_usage":       0,
		"disk_usage":         0,
		"load1":              0,
		"load5":              0,
		"load15":             0,
	})

	logger.Info("Agent 卸载成功", zap.Uint("serverID", serverID))
	return nil
}

// PullMetrics 对单台主机发起 HTTP 拉取，解析指标写入数据库
func (s *AgentService) PullMetrics(serverID uint) error {
	var server models.Server
	if err := db.First(&server, serverID).Error; err != nil {
		return fmt.Errorf("主机不存在: %w", err)
	}

	// 移除状态检查，允许手动触发拉取（即使数据库状态不是 running）

	agentPort := server.AgentPort
	if agentPort == 0 {
		agentPort = 9100
	}

	ip := server.InnerIP
	if ip == "" {
		ip = server.IP
	}

	url := fmt.Sprintf("http://%s:%d/metrics", ip, agentPort)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求 Agent 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var metrics agentMetricsResponse
	if err := json.Unmarshal(body, &metrics); err != nil {
		return fmt.Errorf("解析指标失败: %w", err)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"cpu_usage":          metrics.CPUUsage,
		"memory_usage":       metrics.MemoryUsage,
		"disk_usage":         metrics.DiskUsage,
		"metrics_updated_at": now,
		"agent_status":       "running", // 成功拉取指标，更新状态为 running
		"last_heartbeat_at":  now,
	}

	// 如果 Agent 返回了分区信息，保存到数据库
	if len(metrics.DiskPartitions) > 0 {
		if partitionsJSON, err := json.Marshal(metrics.DiskPartitions); err == nil {
			updates["disk_partitions"] = string(partitionsJSON)
		}
	}

	if err := db.Model(&models.Server{}).Where("id = ?", serverID).Updates(updates).Error; err != nil {
		return fmt.Errorf("写入数据库失败: %w", err)
	}

	logger.Debug("Agent 指标拉取成功", zap.Uint("serverID", serverID),
		zap.Float64("cpu", metrics.CPUUsage),
		zap.Float64("mem", metrics.MemoryUsage),
		zap.Float64("disk", metrics.DiskUsage))
	return nil
}

// ReceiveHeartbeat 接收 Agent 心跳，更新主机状态
func (s *AgentService) ReceiveHeartbeat(data AgentHeartbeatData) error {
	var server models.Server
	err := db.Where("ip = ? OR hostname = ?", data.IP, data.Hostname).First(&server).Error
	if err != nil {
		return fmt.Errorf("未找到对应主机 IP=%s hostname=%s: %w", data.IP, data.Hostname, err)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"agent_status":      "running",
		"last_heartbeat_at": now,
	}
	if data.Version != "" && (server.AgentVersion == "" || data.Version != server.AgentVersion) {
		updates["agent_version"] = data.Version
	} else if server.AgentVersion == "" && data.Version == "" {
		updates["agent_version"] = "1.0.0"
	}
	if data.Port > 0 {
		updates["agent_port"] = data.Port
	}

	// 检查 Agent 状态是否从离线变为在线（触发数据回填）
	shouldBackfill := server.AgentStatus != "running" && server.AgentStatus != ""

	if err := db.Model(&models.Server{}).Where("id = ?", server.ID).Updates(updates).Error; err != nil {
		return err
	}

	// 如果 Agent 从离线恢复，触发数据回填
	if shouldBackfill {
		logger.Info("Agent 从离线恢复，触发数据回填",
			zap.Uint("serverID", server.ID),
			zap.String("hostname", server.Hostname),
			zap.String("previousStatus", server.AgentStatus))
		ScheduleBackfillOnAgentRecovery(server.ID)
	}

	return nil
}

// StartAgentMetricsScheduler 启动 Agent 指标采集调度器
// - 快速采集（性能指标）：每30秒
// - 常规采集（系统信息）：每5分钟
// - 心跳检测：每1分钟
func StartAgentMetricsScheduler() {
	heartbeatTicker := time.NewTicker(1 * time.Minute)

	logger.Info("Agent 指标采集调度器已启动")

	monitoringService := NewMonitoringService()

	// 启动心跳超时检测（每1分钟）
	go func() {
		for range heartbeatTicker.C {
			checkHeartbeatTimeout()
		}
	}()

	// 启动 WebSocket 广播器
	StartMetricsBroadcaster()

	// 快速采集：性能指标（每30秒）
	fastMetricsTicker := time.NewTicker(30 * time.Second)
	go func() {
		for range fastMetricsTicker.C {
			collectFastMetrics(monitoringService)
		}
	}()

	// 常规采集：基础指标 + 系统信息（每5分钟）
	regularMetricsTicker := time.NewTicker(5 * time.Minute)
	go func() {
		for range regularMetricsTicker.C {
			collectRegularMetrics(monitoringService)
		}
	}()

	// 立即执行一次采集
	go collectFastMetrics(monitoringService)
	go collectRegularMetrics(monitoringService)
}

// collectAllAgentServers 对所有可能运行Agent的主机并发拉取指标
func collectAllAgentServers() {
	var serverIDs []uint
	// 采集所有可能运行Agent的主机（running、offline、failed）
	if err := db.Model(&models.Server{}).
		Where("agent_status IN ('running', 'offline', 'failed')").
		Pluck("id", &serverIDs).Error; err != nil {
		logger.Warn("Agent 指标采集：查询主机失败", zap.Error(err))
		return
	}

	if len(serverIDs) == 0 {
		return
	}

	logger.Info("Agent 指标采集：开始批量采集", zap.Int("count", len(serverIDs)))

	svc := NewAgentService()
	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup
	for _, id := range serverIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sid uint) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := svc.PullMetrics(sid); err != nil {
				logger.Debug("Agent 指标拉取失败", zap.Uint("serverID", sid), zap.Error(err))
			}
		}(id)
	}
	wg.Wait()

	logger.Info("Agent 指标采集：批量采集完成", zap.Int("count", len(serverIDs)))
}

// checkHeartbeatTimeout 将超过3分钟未上报心跳的主机状态置为 offline
func checkHeartbeatTimeout() {
	threshold := time.Now().Add(-heartbeatTimeout)
	result := db.Model(&models.Server{}).
		Where("agent_status = 'running' AND last_heartbeat_at < ?", threshold).
		Update("agent_status", "offline")
	if result.RowsAffected > 0 {
		logger.Warn("Agent 心跳超时检测：已标记离线", zap.Int64("count", result.RowsAffected))
	}
}

// collectFastMetrics 快速采集：对所有 agent_status=running 的主机并发拉取性能指标（每30秒）
func collectFastMetrics(monitoringService *MonitoringService) {
	var serverIDs []uint
	// 采集所有可能运行Agent的主机（running、offline、failed）
	// 这允许从离线状态恢复，并更新last_heartbeat_at
	if err := db.Model(&models.Server{}).
		Where("agent_status IN ('running', 'offline', 'failed')").
		Pluck("id", &serverIDs).Error; err != nil {
		logger.Warn("快速指标采集：查询主机失败", zap.Error(err))
		return
	}

	if len(serverIDs) == 0 {
		return
	}

	logger.Debug("快速指标采集：开始批量采集", zap.Int("count", len(serverIDs)))

	sem := make(chan struct{}, 20) // 并发度限制为20
	var wg sync.WaitGroup
	for _, id := range serverIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sid uint) {
			defer wg.Done()
			defer func() { <-sem }()
			if metrics, err := monitoringService.PullExtendedMetrics(sid); err != nil {
				logger.Debug("快速指标拉取失败", zap.Uint("serverID", sid), zap.Error(err))
			} else {
				logger.Debug("快速指标采集成功",
					zap.Uint("serverID", sid),
					zap.Float64("cpu", metrics.Performance.CPU.UsagePercent),
					zap.Float64("memory", metrics.Performance.Memory.UsedPercent))
				// 触发 WebSocket 广播
				broadcastMetricsUpdate(sid, metrics)

				// 存储监控数据到数据库
				now := time.Now()
				updates := map[string]interface{}{
					"cpu_usage":        metrics.Performance.CPU.UsagePercent,
					"memory_usage":     metrics.Performance.Memory.UsedPercent,
					"disk_usage":       metrics.Performance.Disk.UsedPercent,
					"load1":            metrics.Performance.Load.Load1,
					"load5":            metrics.Performance.Load.Load5,
					"load15":           metrics.Performance.Load.Load15,
					"disk_partitions":  nil,
					"metrics_updated_at": &now,
				}

				if len(metrics.Performance.Disk.Partitions) > 0 {
					partitionsJSON, _ := json.Marshal(metrics.Performance.Disk.Partitions)
					updates["disk_partitions"] = string(partitionsJSON)
				}

				if err := db.Model(&models.Server{}).Where("id = ?", sid).Updates(updates).Error; err != nil {
					logger.Warn("存储监控数据失败", zap.Uint("serverID", sid), zap.Error(err))
				}

			}
		}(id)
	}
	wg.Wait()

	logger.Debug("快速指标采集：批量采集完成", zap.Int("count", len(serverIDs)))
}

// collectRegularMetrics 常规采集：对所有可能运行Agent的主机并发拉取基础指标（每5分钟）
func collectRegularMetrics(monitoringService *MonitoringService) {
	var serverIDs []uint
	// 采集所有可能运行Agent的主机（running、offline、failed）
	if err := db.Model(&models.Server{}).
		Where("agent_status IN ('running', 'offline', 'failed')").
		Pluck("id", &serverIDs).Error; err != nil {
		logger.Warn("常规指标采集：查询主机失败", zap.Error(err))
		return
	}

	if len(serverIDs) == 0 {
		return
	}

	logger.Info("常规指标采集：开始批量采集", zap.Int("count", len(serverIDs)))

	svc := NewAgentService()
	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup
	for _, id := range serverIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sid uint) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := svc.PullMetrics(sid); err != nil {
				logger.Debug("常规指标拉取失败", zap.Uint("serverID", sid), zap.Error(err))
			}
		}(id)
	}
	wg.Wait()

	logger.Info("常规指标采集：批量采集完成", zap.Int("count", len(serverIDs)))
}

// collectAllExtendedMetrics 对所有可能运行Agent的主机并发拉取扩展指标
func collectAllExtendedMetrics(monitoringService *MonitoringService) {
	var serverIDs []uint
	// 采集所有可能运行Agent的主机（running、offline、failed）
	if err := db.Model(&models.Server{}).
		Where("agent_status IN ('running', 'offline', 'failed')").
		Pluck("id", &serverIDs).Error; err != nil {
		logger.Warn("扩展指标采集：查询主机失败", zap.Error(err))
		return
	}

	if len(serverIDs) == 0 {
		return
	}

	logger.Debug("扩展指标采集：开始批量采集", zap.Int("count", len(serverIDs)))

	sem := make(chan struct{}, 10) // 并发度限制为10
	var wg sync.WaitGroup
	for _, id := range serverIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sid uint) {
			defer wg.Done()
			defer func() { <-sem }()
			if _, err := monitoringService.PullExtendedMetrics(sid); err != nil {
				logger.Debug("扩展指标拉取失败", zap.Uint("serverID", sid), zap.Error(err))
			}
		}(id)
	}
	wg.Wait()

	logger.Debug("扩展指标采集：批量采集完成", zap.Int("count", len(serverIDs)))
}

// loadServerForAgent 加载主机及其系统运维凭证（credential_type=system）
// 不 fallback 到用户凭证，未配置系统凭证直接返回错误
func (s *AgentService) loadServerForAgent(serverID uint) (*models.Server, error) {
	var server models.Server
	if err := db.Preload("SystemCredential").First(&server, serverID).Error; err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	if server.SystemCredential == nil || server.SystemCredentialID == 0 {
		return nil, fmt.Errorf("主机未配置系统运维凭证，请先在主机编辑页配置")
	}

	// dialSSH 从 server.SSHCredential 读取凭证，将系统凭证放入该字段以复用 dialSSH
	server.SSHCredential = server.SystemCredential
	return &server, nil
}

// scpFile 通过 SSH 将本地文件内容传输到远端路径
func scpFile(client *ssh.Client, localPath, remotePath string) error {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("读取本地文件失败 %s: %w", localPath, err)
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = bytes.NewReader(content)
	return session.Run(fmt.Sprintf("cat > %s", remotePath))
}

// runSSHCommand2 执行 SSH 命令，忽略输出只返回错误
func runSSHCommand2(client *ssh.Client, cmd string) error {
	_, err := runSSHCommand(client, cmd)
	return err
}

// MarkAgentFailed 将指定主机的 agent_status 置为 failed 并记录错误日志
func MarkAgentFailed(serverID uint, reason string) {
	logger.Error("Agent 部署失败",
		zap.Uint("serverID", serverID),
		zap.String("reason", reason))
	db.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"agent_status":       "failed",
		"agent_version":      "",
		"last_heartbeat_at":  nil,
		"metrics_updated_at": nil,
		"cpu_usage":          0,
		"memory_usage":       0,
		"disk_usage":         0,
		"load1":              0,
		"load5":              0,
		"load15":             0,
	})
}

// StartMetricsBroadcaster 启动指标广播器，每30秒广播监控概览数据
func StartMetricsBroadcaster() {
	hub := GetMonitoringHub()
	ticker := time.NewTicker(30 * time.Second)

	logger.Info("监控指标广播器已启动")

	go func() {
		for range ticker.C {
			broadcastOverviewUpdate(hub)
		}
	}()
}

// broadcastOverviewUpdate 广播监控概览更新
func broadcastOverviewUpdate(hub *WebSocketHub) {
	// 获取所有在线主机的最新指标
	var servers []models.Server
	if err := db.Where("agent_status = 'running'").Find(&servers).Error; err != nil {
		logger.Warn("获取在线主机列表失败", zap.Error(err))
		return
	}

	if len(servers) == 0 {
		return
	}

	// 构建概览数据
	serversData := make([]map[string]interface{}, 0, len(servers))
	for _, server := range servers {
		serverData := map[string]interface{}{
			"id":       server.ID,
			"hostname": server.Hostname,
			"ip":       server.IP,
			"cpu":      server.CPUUsage,
			"memory":   server.MemoryUsage,
			"disk":     server.DiskUsage,
			"status":   server.AgentStatus,
		}
		serversData = append(serversData, serverData)
	}

	// 广播消息
	message := &WebSocketMessage{
		Type: "overview_update",
		Data: map[string]interface{}{
			"total_servers": len(servers),
			"online":        len(servers),
			"offline":       0,
			"servers":       serversData,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	hub.Broadcast(message)
	logger.Debug("监控概览已广播", zap.Int("clients", hub.GetClientCount()))
}

// broadcastMetricsUpdate 广播单个主机指标更新
func broadcastMetricsUpdate(serverID uint, metrics *AgentExtendedMetricsResponse) {
	hub := GetMonitoringHub()

	message := &WebSocketMessage{
		Type: "metrics_update",
		Data: map[string]interface{}{
			"server_id":   serverID,
			"hostname":    metrics.SystemInfo.Hostname,
			"cpu":         metrics.Performance.CPU.UsagePercent,
			"memory":      metrics.Performance.Memory.UsedPercent,
			"disk":        metrics.Performance.Disk.UsedPercent,
			"load5":       metrics.Performance.Load.Load5,
			"process_num": len(metrics.ProcessInfo.Top),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	hub.Broadcast(message)
	logger.Debug("主机指标更新已广播",
		zap.Uint("serverID", serverID),
		zap.String("hostname", metrics.SystemInfo.Hostname),
		zap.Int("clients", hub.GetClientCount()))
}

// TestSSHConnectionResponse SSH连接测试响应
type TestSSHConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Host    string `json:"host,omitempty"`
	Port    int    `json:"port,omitempty"`
	User    string `json:"user,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// TestSSHConnection 测试SSH连接是否成功
func (s *AgentService) TestSSHConnection(serverID uint) (*TestSSHConnectionResponse, error) {
	server, err := s.loadServerForAgent(serverID)
	if err != nil {
		return &TestSSHConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("主机不存在: %v", err),
		}, nil
	}

	// 记录开始时间用于计算延迟
	start := time.Now()

	// 尝试建立SSH连接
	sshClient, err := dialSSH(server)
	if err != nil {
		return &TestSSHConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("SSH连接失败: %v", err),
			Host:    server.IP,
			Port:    server.SSHPort,
		}, nil
	}
	defer sshClient.Close()

	// 执行一个简单的命令验证连接
	cmd := "uname -s"
	session, err := sshClient.NewSession()
	if err != nil {
		return &TestSSHConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("创建SSH会话失败: %v", err),
		}, nil
	}
	defer session.Close()

	output, err := session.Output(cmd)
	if err != nil {
		return &TestSSHConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("执行命令失败: %v", err),
		}, nil
	}

	latency := time.Since(start)

	return &TestSSHConnectionResponse{
		Success: true,
		Message: fmt.Sprintf("SSH连接成功，系统: %s", string(output)),
		Host:    server.IP,
		Port:    server.SSHPort,
		User:    server.SystemCredential.Username,
		Latency: latency.String(),
	}, nil
}

// ========================================
// Agent 版本管理服务
// ========================================

// GetLatestVersion 获取最新版本信息
func GetLatestVersion() (*models.AgentVersion, error) {
	var version models.AgentVersion
	err := db.Where("is_latest = ?", true).First(&version).Error
	if err != nil {
		return nil, fmt.Errorf("未找到最新版本: %w", err)
	}
	return &version, nil
}

// GetAllVersions 获取所有版本列表
func GetAllVersions() ([]models.AgentVersion, error) {
	var versions []models.AgentVersion
	err := db.Order("released_at DESC").Find(&versions).Error
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	return versions, nil
}

// GetVersionByID 根据ID获取版本信息
func GetVersionByID(versionID uint) (*models.AgentVersion, error) {
	var version models.AgentVersion
	err := db.First(&version, versionID).Error
	if err != nil {
		return nil, fmt.Errorf("版本不存在: %w", err)
	}
	return &version, nil
}

// GetVersionByNumber 根据版本号获取版本信息
func GetVersionByNumber(versionNumber string) (*models.AgentVersion, error) {
	var version models.AgentVersion
	err := db.Where("version = ?", versionNumber).First(&version).Error
	if err != nil {
		return nil, fmt.Errorf("版本不存在: %w", err)
	}
	return &version, nil
}

// CreateVersion 创建新版本
func CreateVersion(version *models.AgentVersion, operator string) error {

	// 检查版本号是否已存在
	var existingVersion models.AgentVersion
	if err := db.Where("version = ?", version.Version).First(&existingVersion).Error; err == nil {
		return fmt.Errorf("版本号 %s 已存在", version.Version)
	}

	// 如果设置为最新版本，先将其他版本的 is_latest 设为 false
	if version.IsLatest {
		db.Model(&models.AgentVersion{}).Where("is_latest = ?", true).Update("is_latest", false)
	}

	if err := db.Create(version).Error; err != nil {
		return fmt.Errorf("创建版本失败: %w", err)
	}

	logger.Info("创建 Agent 版本成功",
		zap.String("version", version.Version),
		zap.String("operator", operator))
	return nil
}

// UpdateVersion 更新版本信息
func UpdateVersion(versionID uint, updates map[string]interface{}) error {

	// 如果更新 is_latest，需要处理其他版本
	if isLatest, ok := updates["is_latest"].(bool); ok && isLatest {
		db.Model(&models.AgentVersion{}).Where("is_latest = ? AND id != ?", true, versionID).Update("is_latest", false)
	}

	if err := db.Model(&models.AgentVersion{}).Where("id = ?", versionID).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新版本失败: %w", err)
	}

	logger.Info("更新 Agent 版本成功", zap.Uint("versionID", versionID))
	return nil
}

// DeleteVersion 删除版本
func DeleteVersion(versionID uint) error {
	var version models.AgentVersion
	if err := db.First(&version, versionID).Error; err != nil {
		return fmt.Errorf("版本不存在: %w", err)
	}

	// 不能删除最新版本
	if version.IsLatest {
		return fmt.Errorf("不能删除最新版本")
	}

	// 检查是否有主机正在使用此版本
	var count int64
	db.Model(&models.Server{}).Where("agent_version = ?", version.Version).Count(&count)
	if count > 0 {
		return fmt.Errorf("仍有 %d 台主机正在使用此版本，无法删除", count)
	}

	if err := db.Delete(&version).Error; err != nil {
		return fmt.Errorf("删除版本失败: %w", err)
	}

	logger.Info("删除 Agent 版本成功",
		zap.Uint("versionID", versionID),
		zap.String("version", version.Version))
	return nil
}

// ========================================
// Agent 版本兼容性检查
// ========================================

// compareVersion 版本号比较函数
// 返回: v1 operator v2 的结果
// operator 支持: >, >=, <, <=, =, ==
func compareVersion(v1, v2, operator string) bool {
	// 将版本号字符串分割为数字数组
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	// 确保两个版本号段数相同
	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	// 补齐短版本号
	for len(v1Parts) < maxLen {
		v1Parts = append(v1Parts, "0")
	}
	for len(v2Parts) < maxLen {
		v2Parts = append(v2Parts, "0")
	}

	// 逐段比较
	for i := 0; i < maxLen; i++ {
		v1Num, _ := strconv.Atoi(v1Parts[i])
		v2Num, _ := strconv.Atoi(v2Parts[i])

		if v1Num > v2Num {
			result := (operator == ">" || operator == ">=" || operator == "!=")
			return result
		}
		if v1Num < v2Num {
			result := (operator == "<" || operator == "<=" || operator == "!=")
			return result
		}
	}

	// 版本号相等
	return (operator == "=" || operator == "==" || operator == "<=" || operator == ">=")
}

// CheckVersionCompatibility 检查版本兼容性
func CheckVersionCompatibility(currentVersion, targetVersion string) (bool, string) {
	// 如果当前版本为空，允许升级
	if currentVersion == "" || currentVersion == "unknown" {
		return true, ""
	}

	var target models.AgentVersion
	if err := db.Where("version = ?", targetVersion).First(&target).Error; err != nil {
		return false, "目标版本不存在"
	}

	// 检查版本兼容性范围
	if target.MinCompatibleVersion != "" {
		if !compareVersion(currentVersion, target.MinCompatibleVersion, ">=") {
			return false, fmt.Sprintf("当前版本 %s 低于最小兼容版本 %s", currentVersion, target.MinCompatibleVersion)
		}
	}

	if target.MaxCompatibleVersion != "" {
		if !compareVersion(currentVersion, target.MaxCompatibleVersion, "<=") {
			return false, fmt.Sprintf("当前版本 %s 高于最大兼容版本 %s", currentVersion, target.MaxCompatibleVersion)
		}
	}

	return true, ""
}

// ========================================
// Agent 升级功能
// ========================================

// UpgradeAgent 升级单台主机的 Agent
func (s *AgentService) UpgradeAgent(serverID uint, targetVersion string) error {
	server, err := s.loadServerForAgent(serverID)
	if err != nil {
		return err
	}

	if server.AgentVersion == targetVersion {
		return fmt.Errorf("主机当前版本已是目标版本 %s", targetVersion)
	}

	compatible, reason := CheckVersionCompatibility(server.AgentVersion, targetVersion)
	if !compatible {
		return fmt.Errorf("版本不兼容: %s", reason)
	}

	taskName := fmt.Sprintf("升级 %s 到 v%s", server.Hostname, targetVersion)
	serverIDsJSON, _ := json.Marshal([]uint{serverID})
	now := time.Now()
	task := models.AgentUpgradeTask{
		TaskName:        taskName,
		TargetVersion:   targetVersion,
		TargetServerIDs: string(serverIDsJSON),
		Status:          "running",
		TotalCount:      1,
		SuccessCount:    0,
		FailedCount:     0,
		StartedAt:       &now,
	}
	if err := db.Create(&task).Error; err != nil {
		logger.Error("创建升级任务记录失败", zap.Error(err))
	}

	sshClient, err := dialSSH(server)
	if err != nil {
		markTaskFailed(task.ID, fmt.Sprintf("SSH 连接失败: %v", err))
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer sshClient.Close()

	targetVersionInfo, err := GetVersionByNumber(targetVersion)
	if err != nil {
		markTaskFailed(task.ID, fmt.Sprintf("获取目标版本信息失败: %v", err))
		return fmt.Errorf("获取目标版本信息失败: %w", err)
	}

	arch, _ := runSSHCommand(sshClient, "uname -m")
	binaryArch := "amd64"
	if arch == "aarch64" || arch == "arm64" {
		binaryArch = "arm64"
	}

	var binaryPath string
	if binaryArch == "amd64" {
		binaryPath = targetVersionInfo.AMD64BinaryPath
		if binaryPath == "" {
			binaryPath = fmt.Sprintf("%s/oneops-agent-linux-amd64", agentBinaryDir())
		}
	} else {
		binaryPath = targetVersionInfo.ARM64BinaryPath
		if binaryPath == "" {
			binaryPath = fmt.Sprintf("%s/oneops-agent-linux-arm64", agentBinaryDir())
		}
	}

	if binaryPath == "" {
		markTaskFailed(task.ID, fmt.Sprintf("目标版本不支持 %s 架构", binaryArch))
		return fmt.Errorf("目标版本不支持 %s 架构", binaryArch)
	}

	if err := runSSHCommand2(sshClient, fmt.Sprintf("systemctl stop %s", agentServiceName)); err != nil {
		markTaskFailed(task.ID, fmt.Sprintf("停止 Agent 失败: %v", err))
		return fmt.Errorf("停止 Agent 失败: %w", err)
	}

	remoteBinary := fmt.Sprintf("%s/%s", agentInstallDir, agentBinaryName)
	if err := scpFile(sshClient, binaryPath, remoteBinary); err != nil {
		runSSHCommand2(sshClient, fmt.Sprintf("systemctl start %s", agentServiceName))
		markTaskFailed(task.ID, fmt.Sprintf("上传新版本失败: %v", err))
		return fmt.Errorf("上传新版本失败: %w", err)
	}

	if err := runSSHCommand2(sshClient, fmt.Sprintf("chmod +x %s", remoteBinary)); err != nil {
		markTaskFailed(task.ID, fmt.Sprintf("设置执行权限失败: %v", err))
		return fmt.Errorf("设置执行权限失败: %w", err)
	}

	if err := runSSHCommand2(sshClient, fmt.Sprintf("systemctl start %s", agentServiceName)); err != nil {
		markTaskFailed(task.ID, fmt.Sprintf("启动新版本失败: %v", err))
		return fmt.Errorf("启动新版本失败: %w", err)
	}

	go func() {
		for i := 0; i < 20; i++ {
			time.Sleep(3 * time.Second)
			var svr models.Server
			if err := db.First(&svr, serverID).Error; err != nil {
				continue
			}
			if svr.AgentVersion == targetVersion && svr.AgentStatus == "running" {
				db.Model(&models.AgentVersion{}).Where("version = ?", targetVersion).
					UpdateColumn("deploy_count", gorm.Expr("deploy_count + ?", 1))
				markTaskSuccess(task.ID)
				logger.Info("Agent 升级成功",
					zap.Uint("serverID", serverID),
					zap.String("version", targetVersion))
				return
			}
		}
		markTaskFailed(task.ID, "升级超时，未收到新版本心跳")
		logger.Error("Agent 升级超时",
			zap.Uint("serverID", serverID),
			zap.String("targetVersion", targetVersion))
	}()

	logger.Info("Agent 升级任务已提交",
		zap.Uint("serverID", serverID),
		zap.String("version", targetVersion),
		zap.Uint("taskID", task.ID))
	return nil
}

func markTaskSuccess(taskID uint) {
	now := time.Now()
	db.Model(&models.AgentUpgradeTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":       "completed",
		"success_count": 1,
		"completed_at": &now,
	})
}

func markTaskFailed(taskID uint, errMsg string) {
	now := time.Now()
	db.Model(&models.AgentUpgradeTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":        "failed",
		"failed_count":  1,
		"error_message": errMsg,
		"completed_at":  &now,
	})
}

// GetUpgradeTasks 获取升级任务列表
func GetUpgradeTasks(page, pageSize int, status string) ([]models.AgentUpgradeTask, int64, error) {
	var tasks []models.AgentUpgradeTask
	var total int64

	query := db.Model(&models.AgentUpgradeTask{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询任务总数失败: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("查询任务列表失败: %w", err)
	}

	return tasks, total, nil
}

// GetUpgradeTaskByID 获取升级任务详情
func GetUpgradeTaskByID(taskID uint) (*models.AgentUpgradeTask, error) {
	var task models.AgentUpgradeTask
	if err := db.First(&task, taskID).Error; err != nil {
		return nil, fmt.Errorf("任务不存在: %w", err)
	}
	return &task, nil
}
