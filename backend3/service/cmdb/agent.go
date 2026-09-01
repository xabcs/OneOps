package cmdb

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

	modelcmdb "oneops/backend3/model/cmdb"
	"oneops/backend3/pkg/database"
	"oneops/backend3/pkg/logger"
	"oneops/backend3/pkg/utils"
	repocmdb "oneops/backend3/repository/cmdb"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

const (
	agentInstallDir  = "/opt/oneops-agent"
	agentBinaryName  = "oneops-agent"
	agentServiceName = "oneops-agent"
	heartbeatTimeout = 3 * time.Minute
)

// agentBinaryDir 返回 Agent 预编译二进制所在目录
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
	CPUUsage       float64                  `json:"cpuUsage"`
	MemoryUsage    float64                  `json:"memoryUsage"`
	DiskUsage      float64                  `json:"diskUsage"`
	DiskPartitions modelcmdb.DiskPartitions `json:"diskPartitions,omitempty"`
	Load5          float64                  `json:"load5"`
	ProcessNum     int                      `json:"processNum"`
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
type AgentService struct {
	repo *repocmdb.AgentRepository
}

// NewAgentService 创建 AgentService 实例
func NewAgentService(repo *repocmdb.AgentRepository) *AgentService {
	return &AgentService{repo: repo}
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

	arch, _ := runSSHCommand(sshClient, "uname -m")
	binaryArch := "amd64"
	if arch == "aarch64" || arch == "arm64" {
		binaryArch = "arm64"
	}

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

	// TODO: 获取后端外部访问地址（config 模块完成后替换）
	heartbeatURL := fmt.Sprintf("http://localhost:8080/api/cmdb/agent/heartbeat")

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

	s.repo.UpdateServerFields(serverID, map[string]interface{}{
		"agent_status":  "running",
		"agent_port":    agentPort,
		"agent_version": "1.0.0",
	})

	logger.Info("Agent 部署成功", zap.Uint("server_id", serverID), zap.String("arch", binaryArch))
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
		updates["agent_version"] = "1.0.0"
	}
	s.repo.UpdateServerFields(serverID, updates)
	logger.Info("Agent 重启成功", zap.Uint("server_id", serverID))
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
		logger.Error("卸载 Agent SSH命令执行失败", zap.Uint("server_id", serverID), zap.Error(err))
		return fmt.Errorf("卸载 Agent 失败: %w", err)
	}

	s.repo.UpdateServerFields(serverID, map[string]interface{}{
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

	logger.Info("Agent 卸载成功", zap.Uint("server_id", serverID))
	return nil
}

// PullMetrics 对单台主机发起 HTTP 拉取，解析指标写入数据库
func (s *AgentService) PullMetrics(serverID uint) error {
	server, err := s.repo.FindServerByID(serverID)
	if err != nil {
		return fmt.Errorf("主机不存在: %w", err)
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
	updates := map[string]interface{}{
		"cpu_usage":          metrics.CPUUsage,
		"memory_usage":       metrics.MemoryUsage,
		"disk_usage":         metrics.DiskUsage,
		"metrics_updated_at": now,
		"agent_status":       "running",
		"last_heartbeat_at":  now,
	}

	if len(metrics.DiskPartitions) > 0 {
		if partitionsJSON, err := json.Marshal(metrics.DiskPartitions); err == nil {
			updates["disk_partitions"] = string(partitionsJSON)
		}
	}

	if err := s.repo.UpdateServerFields(serverID, updates); err != nil {
		return fmt.Errorf("写入数据库失败: %w", err)
	}

	logger.Debug("Agent 指标拉取成功", zap.Uint("server_id", serverID),
		zap.Float64("cpu", metrics.CPUUsage),
		zap.Float64("mem", metrics.MemoryUsage),
		zap.Float64("disk", metrics.DiskUsage))
	return nil
}

// ReceiveHeartbeat 接收 Agent 心跳，更新主机状态
func (s *AgentService) ReceiveHeartbeat(data AgentHeartbeatData) error {
	server, err := s.repo.FindServerByIPOrHostname(data.IP, data.Hostname)
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

	shouldBackfill := server.AgentStatus != "running" && server.AgentStatus != ""

	if err := s.repo.UpdateServerFields(server.ID, updates); err != nil {
		return err
	}

	if shouldBackfill {
		logger.Info("Agent 从离线恢复，触发数据回填",
			zap.Uint("server_id", server.ID),
			zap.String("hostname", server.Hostname),
			zap.String("previous_status", server.AgentStatus))
		// TODO: ScheduleBackfillOnAgentRecovery — 监控模块完成后实现
	}

	return nil
}

// StartAgentMetricsScheduler 启动 Agent 指标采集调度器
func StartAgentMetricsScheduler() {
	heartbeatTicker := time.NewTicker(1 * time.Minute)

	logger.Info("Agent 指标采集调度器已启动")

	utils.SafeGo("agent-heartbeat-check", func() {
		for range heartbeatTicker.C {
			utils.SafeRun("agent-heartbeat-check", checkHeartbeatTimeout)
		}
	})

	fastMetricsTicker := time.NewTicker(30 * time.Second)
	utils.SafeGo("agent-fast-metrics", func() {
		for range fastMetricsTicker.C {
			utils.SafeRun("agent-fast-metrics", collectAllAgentServers)
		}
	})

	regularMetricsTicker := time.NewTicker(5 * time.Minute)
	utils.SafeGo("agent-regular-metrics", func() {
		for range regularMetricsTicker.C {
			utils.SafeRun("agent-regular-metrics", collectRegularMetrics)
		}
	})

	utils.SafeGo("agent-fast-metrics-init", collectAllAgentServers)
	utils.SafeGo("agent-regular-metrics-init", collectRegularMetrics)
}

// collectAllAgentServers 对所有可能运行Agent的主机并发拉取指标
func collectAllAgentServers() {
	db := database.GetDB()
	agentRepo := repocmdb.NewAgentRepository(db)
	serverIDs, err := agentRepo.PluckAgentServerIDs()
	if err != nil {
		logger.Warn("Agent 指标采集：查询主机失败", zap.Error(err))
		return
	}

	if len(serverIDs) == 0 {
		return
	}

	logger.Info("Agent 指标采集：开始批量采集", zap.Int("count", len(serverIDs)))

	svc := NewAgentService(agentRepo)
	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup
	for _, id := range serverIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sid uint) {
			defer wg.Done()
			defer func() { <-sem }()
			utils.SafeRun("agent-pull-metrics", func() {
				if err := svc.PullMetrics(sid); err != nil {
					logger.Debug("Agent 指标拉取失败", zap.Uint("server_id", sid), zap.Error(err))
				}
			})
		}(id)
	}
	wg.Wait()

	logger.Info("Agent 指标采集：批量采集完成", zap.Int("count", len(serverIDs)))
}

// checkHeartbeatTimeout 将超过3分钟未上报心跳的主机状态置为 offline
func checkHeartbeatTimeout() {
	db := database.GetDB()
	agentRepo := repocmdb.NewAgentRepository(db)
	threshold := time.Now().Add(-heartbeatTimeout)
	count, err := agentRepo.MarkHeartbeatTimeouts(threshold)
	if err != nil {
		logger.Warn("Agent 心跳超时检测失败", zap.Error(err))
		return
	}
	if count > 0 {
		logger.Warn("Agent 心跳超时检测：已标记离线", zap.Int64("count", count))
	}
}

// collectRegularMetrics 常规采集
func collectRegularMetrics() {
	db := database.GetDB()
	agentRepo := repocmdb.NewAgentRepository(db)
	serverIDs, err := agentRepo.PluckAgentServerIDs()
	if err != nil {
		logger.Warn("常规指标采集：查询主机失败", zap.Error(err))
		return
	}

	if len(serverIDs) == 0 {
		return
	}

	logger.Info("常规指标采集：开始批量采集", zap.Int("count", len(serverIDs)))

	svc := NewAgentService(agentRepo)
	sem := make(chan struct{}, 20)
	var wg sync.WaitGroup
	for _, id := range serverIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sid uint) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := svc.PullMetrics(sid); err != nil {
				logger.Debug("常规指标拉取失败", zap.Uint("server_id", sid), zap.Error(err))
			}
		}(id)
	}
	wg.Wait()

	logger.Info("常规指标采集：批量采集完成", zap.Int("count", len(serverIDs)))
}

// loadServerForAgent 加载主机及其系统运维凭证
func (s *AgentService) loadServerForAgent(serverID uint) (*modelcmdb.Server, error) {
	server, err := s.repo.FindServerForAgent(serverID)
	if err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	if server.SystemCredential == nil || server.SystemCredentialID == 0 {
		return nil, fmt.Errorf("主机未配置系统运维凭证，请先在主机编辑页配置")
	}

	// 解密凭证密码
	cred := server.SystemCredential
	if cred.Password != "" {
		if decrypted, err := utils.DecryptString(cred.Password); err == nil {
			cred.Password = decrypted
		} else {
			logger.Warn("系统凭证密码解密失败", zap.Uint("credential_id", cred.ID), zap.Error(err))
		}
	}
	if cred.PrivateKey != "" {
		if decrypted, err := utils.DecryptString(cred.PrivateKey); err == nil {
			cred.PrivateKey = decrypted
		} else {
			logger.Warn("系统凭证私钥解密失败", zap.Uint("credential_id", cred.ID), zap.Error(err))
		}
	}
	if cred.Passphrase != "" {
		if decrypted, err := utils.DecryptString(cred.Passphrase); err == nil {
			cred.Passphrase = decrypted
		}
	}

	server.SSHCredential = cred
	return server, nil
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

// MarkAgentFailed 将指定主机的 agent_status 置为 failed
func (s *AgentService) MarkAgentFailed(serverID uint, reason string) {
	logger.Error("Agent 部署失败",
		zap.Uint("server_id", serverID),
		zap.String("reason", reason))
	s.repo.UpdateServerFields(serverID, map[string]interface{}{
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

	start := time.Now()

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

	session, err := sshClient.NewSession()
	if err != nil {
		return &TestSSHConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("创建SSH会话失败: %v", err),
		}, nil
	}
	defer session.Close()

	output, err := session.Output("uname -s")
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
func (s *AgentService) GetLatestVersion() (*modelcmdb.AgentVersion, error) {
	version, err := s.repo.FindLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("未找到最新版本: %w", err)
	}
	return version, nil
}

// GetAllVersions 获取所有版本列表
func (s *AgentService) GetAllVersions() ([]modelcmdb.AgentVersion, error) {
	versions, err := s.repo.FindAllVersions()
	if err != nil {
		return nil, fmt.Errorf("获取版本列表失败: %w", err)
	}
	return versions, nil
}

// GetVersionByID 根据ID获取版本信息
func (s *AgentService) GetVersionByID(versionID uint) (*modelcmdb.AgentVersion, error) {
	version, err := s.repo.FindVersionByID(versionID)
	if err != nil {
		return nil, fmt.Errorf("版本不存在: %w", err)
	}
	return version, nil
}

// GetVersionByNumber 根据版本号获取版本信息
func (s *AgentService) GetVersionByNumber(versionNumber string) (*modelcmdb.AgentVersion, error) {
	version, err := s.repo.FindVersionByNumber(versionNumber)
	if err != nil {
		return nil, fmt.Errorf("版本不存在: %w", err)
	}
	return version, nil
}

// CreateVersion 创建新版本
func (s *AgentService) CreateVersion(version *modelcmdb.AgentVersion, operator string) error {
	if _, err := s.repo.FindExistingVersion(version.Version); err == nil {
		return fmt.Errorf("版本号 %s 已存在", version.Version)
	}

	if version.IsLatest {
		s.repo.ClearLatestFlags()
	}

	if err := s.repo.CreateVersion(version); err != nil {
		return fmt.Errorf("创建版本失败: %w", err)
	}

	logger.Info("创建 Agent 版本成功",
		zap.String("version", version.Version),
		zap.String("operator", operator))
	return nil
}

// UpdateVersion 更新版本信息
func (s *AgentService) UpdateVersion(versionID uint, updates map[string]interface{}) error {
	if isLatest, ok := updates["is_latest"].(bool); ok && isLatest {
		s.repo.ClearLatestFlagsExcludeID(versionID)
	}

	if err := s.repo.UpdateVersion(versionID, updates); err != nil {
		return fmt.Errorf("更新版本失败: %w", err)
	}

	logger.Info("更新 Agent 版本成功", zap.Uint("version_id", versionID))
	return nil
}

// DeleteVersion 删除版本
func (s *AgentService) DeleteVersion(versionID uint) error {
	version, err := s.repo.FindVersionByID(versionID)
	if err != nil {
		return fmt.Errorf("版本不存在: %w", err)
	}

	if version.IsLatest {
		return fmt.Errorf("不能删除最新版本")
	}

	count, err := s.repo.CountServersByVersion(version.Version)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("仍有 %d 台主机正在使用此版本，无法删除", count)
	}

	if err := s.repo.DeleteVersion(version); err != nil {
		return fmt.Errorf("删除版本失败: %w", err)
	}

	logger.Info("删除 Agent 版本成功",
		zap.Uint("version_id", versionID),
		zap.String("version", version.Version))
	return nil
}

// ========================================
// Agent 版本兼容性检查
// ========================================

// compareVersion 版本号比较函数
func compareVersion(v1, v2, operator string) bool {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for len(v1Parts) < maxLen {
		v1Parts = append(v1Parts, "0")
	}
	for len(v2Parts) < maxLen {
		v2Parts = append(v2Parts, "0")
	}

	for i := 0; i < maxLen; i++ {
		v1Num, _ := strconv.Atoi(v1Parts[i])
		v2Num, _ := strconv.Atoi(v2Parts[i])

		if v1Num > v2Num {
			return (operator == ">" || operator == ">=" || operator == "!=")
		}
		if v1Num < v2Num {
			return (operator == "<" || operator == "<=" || operator == "!=")
		}
	}

	return (operator == "=" || operator == "==" || operator == "<=" || operator == ">=")
}

// CheckVersionCompatibility 检查版本兼容性
func (s *AgentService) CheckVersionCompatibility(currentVersion, targetVersion string) (bool, string) {
	if currentVersion == "" || currentVersion == "unknown" {
		return true, ""
	}

	target, err := s.repo.FindVersionByNumberOrNil(targetVersion)
	if err != nil {
		return false, "目标版本不存在"
	}

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

	compatible, reason := s.CheckVersionCompatibility(server.AgentVersion, targetVersion)
	if !compatible {
		return fmt.Errorf("版本不兼容: %s", reason)
	}

	taskName := fmt.Sprintf("升级 %s 到 v%s", server.Hostname, targetVersion)
	serverIDsJSON, _ := json.Marshal([]uint{serverID})
	now := time.Now()
	task := modelcmdb.AgentUpgradeTask{
		TaskName:        taskName,
		TargetVersion:   targetVersion,
		TargetServerIDs: string(serverIDsJSON),
		Status:          "running",
		TotalCount:      1,
		SuccessCount:    0,
		FailedCount:     0,
		StartedAt:       &now,
	}
	if err := s.repo.CreateUpgradeTask(&task); err != nil {
		logger.Error("创建升级任务记录失败", zap.Error(err))
	}

	sshClient, err := dialSSH(server)
	if err != nil {
		markTaskFailed(task.ID, fmt.Sprintf("SSH 连接失败: %v", err))
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer sshClient.Close()

	targetVersionInfo, err := s.GetVersionByNumber(targetVersion)
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

	utils.SafeGo("agent-upgrade-verify", func() {
		db := database.GetDB()
		agentRepo := repocmdb.NewAgentRepository(db)
		for i := 0; i < 20; i++ {
			time.Sleep(3 * time.Second)
			svr, err := agentRepo.FindServerByID(serverID)
			if err != nil {
				continue
			}
			if svr.AgentVersion == targetVersion && svr.AgentStatus == "running" {
				agentRepo.IncrementDeployCount(targetVersion)
				markTaskSuccess(task.ID)
				logger.Info("Agent 升级成功",
					zap.Uint("server_id", serverID),
					zap.String("version", targetVersion))
				return
			}
		}
		markTaskFailed(task.ID, "升级超时，未收到新版本心跳")
		logger.Error("Agent 升级超时",
			zap.Uint("server_id", serverID),
			zap.String("target_version", targetVersion))
	})

	logger.Info("Agent 升级任务已提交",
		zap.Uint("server_id", serverID),
		zap.String("version", targetVersion),
		zap.Uint("task_id", task.ID))
	return nil
}

func markTaskSuccess(taskID uint) {
	db := database.GetDB()
	agentRepo := repocmdb.NewAgentRepository(db)
	now := time.Now()
	agentRepo.UpdateUpgradeTask(taskID, map[string]interface{}{
		"status":        "completed",
		"success_count": 1,
		"completed_at":  &now,
	})
}

func markTaskFailed(taskID uint, errMsg string) {
	db := database.GetDB()
	agentRepo := repocmdb.NewAgentRepository(db)
	now := time.Now()
	agentRepo.UpdateUpgradeTask(taskID, map[string]interface{}{
		"status":        "failed",
		"failed_count":  1,
		"error_message": errMsg,
		"completed_at":  &now,
	})
}

// GetUpgradeTasks 获取升级任务列表
func (s *AgentService) GetUpgradeTasks(page, pageSize int, status string) ([]modelcmdb.AgentUpgradeTask, int64, error) {
	return s.repo.FindUpgradeTasks(page, pageSize, status)
}

// GetUpgradeTaskByID 获取升级任务详情
func (s *AgentService) GetUpgradeTaskByID(taskID uint) (*modelcmdb.AgentUpgradeTask, error) {
	task, err := s.repo.FindUpgradeTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("任务不存在: %w", err)
	}
	return task, nil
}
