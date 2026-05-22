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
	"sync"
	"time"

	"oneops/backend/logger"
	"oneops/backend/models"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
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
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	DiskUsage   float64 `json:"diskUsage"`
	Load5       float64 `json:"load5"`
	ProcessNum  int     `json:"processNum"`
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
	serviceContent := fmt.Sprintf(`[Unit]
Description=OneOps Agent
After=network.target

[Service]
Type=simple
ExecStart=%s/%s --port %d
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, agentInstallDir, agentBinaryName, agentPort)

	writeCmd := fmt.Sprintf("cat > /etc/systemd/system/%s.service << 'ONEOPS_EOF'\n%sONEOPS_EOF", agentServiceName, serviceContent)
	if err := runSSHCommand2(sshClient, writeCmd); err != nil {
		return fmt.Errorf("写入 systemd 配置失败: %w", err)
	}

	startCmds := fmt.Sprintf("systemctl daemon-reload && systemctl enable %s && systemctl restart %s", agentServiceName, agentServiceName)
	if err := runSSHCommand2(sshClient, startCmds); err != nil {
		return fmt.Errorf("启动 Agent 服务失败: %w", err)
	}

	db.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"agent_status": "running",
		"agent_port":   agentPort,
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

	db.Model(&models.Server{}).Where("id = ?", serverID).Update("agent_status", "running")
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
		return fmt.Errorf("卸载 Agent 失败: %w", err)
	}

	db.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"agent_status":      "uninstalled",
		"agent_version":     "",
		"last_heartbeat_at": nil,
		"cpu_usage":         0,
		"memory_usage":      0,
		"disk_usage":        0,
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

	if server.AgentStatus != "running" {
		return fmt.Errorf("Agent 未运行（当前状态: %s）", server.AgentStatus)
	}

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
	if err := db.Model(&models.Server{}).Where("id = ?", serverID).Updates(map[string]interface{}{
		"cpu_usage":          metrics.CPUUsage,
		"memory_usage":       metrics.MemoryUsage,
		"disk_usage":         metrics.DiskUsage,
		"metrics_updated_at": now,
	}).Error; err != nil {
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
	if data.Version != "" {
		updates["agent_version"] = data.Version
	}
	if data.Port > 0 {
		updates["agent_port"] = data.Port
	}

	return db.Model(&models.Server{}).Where("id = ?", server.ID).Updates(updates).Error
}

// StartAgentMetricsScheduler 启动 Agent 指标采集调度器（每5分钟）并定期检测心跳超时（每1分钟）
func StartAgentMetricsScheduler() {
	metricsTicker := time.NewTicker(5 * time.Minute)
	heartbeatTicker := time.NewTicker(1 * time.Minute)

	logger.Info("Agent 指标采集调度器已启动")

	go collectAllAgentServers()

	go func() {
		for {
			select {
			case <-metricsTicker.C:
				collectAllAgentServers()
			case <-heartbeatTicker.C:
				checkHeartbeatTimeout()
			}
		}
	}()
}

// collectAllAgentServers 对所有 agent_status=running 的主机并发拉取指标
func collectAllAgentServers() {
	var serverIDs []uint
	if err := db.Model(&models.Server{}).
		Where("agent_status = 'running'").
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

// loadServerForAgent 加载主机及其系统运维凭证（credential_type=system）
// 不 fallback 到用户凭证，未配置系统凭证直接返回错误
func (s *AgentService) loadServerForAgent(serverID uint) (*models.Server, error) {
	var server models.Server
	if err := db.Preload("SystemCredential").First(&server, serverID).Error; err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	if server.SystemCredential == nil || server.SystemCredentialID == 0 {
		return nil, fmt.Errorf("主机未配置系统运维凭证，请在主机编辑页绑定 credential_type=system 的凭证")
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
	db.Model(&models.Server{}).Where("id = ?", serverID).
		Update("agent_status", "failed")
}
