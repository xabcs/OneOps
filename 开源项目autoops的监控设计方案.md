# AutoOps 平台监控架构完整总结

该项目采用**自定义 Agent + Prometheus 架构**实现主机监控，支持物理主机和云主机两种场景。**agent.go:844-919**

---

## 核心架构

```
目标主机 Agent (gopsutil)
  ├─ 每30秒 PUSH → Pushgateway :9091
  └─ 被动暴露 /metrics :9100

Prometheus :9090 ──scrape──→ Pushgateway ──存储时序数据
OneOps 后端 ──PromQL──→ Prometheus → 返回给前端
```

---

## 1. Agent 实现方式

### 自定义 Agent 特点

- **动态生成代码**：通过 `CreateAgentMainFile` 函数动态生成完整的 Go 代码，不是 node-exporter **agent.go:844-919**
- **使用 gopsutil 库**：采集 CPU、内存、磁盘、网络等系统指标 **agent.go:19-25**
- **主动推送**：将指标推送到 Prometheus Pushgateway，而非被动拉取 **agent.go:1515-1526**
- **心跳机制**：定期向后端发送心跳报告运行状态 **agent.go:1127-1159**
- **监听 9100 端口**：提供 `/metrics` 端点 **agent.go:800-813**

### Agent 代码生成示例

```go
// CreateAgentMainFile 创建独立可执行的 agent main 文件
func CreateAgentMainFile(heartbeatURL, heartbeatToken, pushgatewayURL string) string {
    // 设置默认值
    if heartbeatURL == "" {
        heartbeatURL = "http://127.0.0.1:8000/api/v1/monitor/agent/heartbeat"
    }
    if heartbeatToken == "" {
        heartbeatToken = "agent-heartbeat-token-2024"
    }
    if pushgatewayURL == "" {
        pushgatewayURL = "http://8.130.14.34:9091"
    }

    // 直接构建完整的 agent 代码
    mainCode := fmt.Sprintf(`package main
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/mem"
    // ... 其他依赖
)
// ... 完整的 Agent 代码
`, pushgatewayURL, heartbeatURL, heartbeatToken)

    return mainCode
}
```

### 扩展监控能力

- 进程监控（nginx、mysql、redis 等）**agent.go:660-665**
- 端点监控（HTTP 响应时间和状态码）**agent.go:660-665**
- 网络连通性监控（Ping）**agent.go:660-665**
- TCP 端口监控 **agent.go:660-665**

---

## 2. 数据采集流程

### Agent 端采集

```go
func monitorSystemMetrics(intervalSeconds float64) {
    // 采集 CPU 指标
    if percents, err := cpu.Percent(0, false); err == nil {
        if len(percents) > 0 {
            cpuTotalUsage.WithLabelValues(hostname).Set(percents[0])
        }
    }

    // 采集内存指标
    if memInfo, err := mem.VirtualMemory(); err == nil {
        memUsage.WithLabelValues(hostname).Set(memInfo.UsedPercent)
    }

    // 采集磁盘使用率（只采集根分区）
    if partitions, err := disk.Partitions(false); err == nil {
        for _, part := range partitions {
            if part.Mountpoint == "/" {
                if usage, err := disk.Usage(part.Mountpoint); err == nil {
                    diskUsage.WithLabelValues(hostname, part.Device, part.Mountpoint).Set(usage.UsedPercent)
                }
            }
        }
    }
}
```

采集指标列表：
- **CPU 使用率**：`cpu.Percent(0, false)` **agent.go:336-340**
- **内存使用率**：`mem.VirtualMemory()` **agent.go:349-352**
- **磁盘使用率**：`disk.Partitions(false)` 和 `disk.Usage()` **agent.go:354-367**
- **磁盘 IO 速率**：`disk.IOCounters()` **agent.go:369-383**
- **网络速率**：`psnet.IOCounters(true)` **agent.go:390-445**
- **进程总数**：`process.Processes()` **agent.go:385-388**

### Agent 启动和推送逻辑

```go
func StartAgentWithConfig(config *Config) error {
    // 创建自定义注册表
    registry := prometheus.NewRegistry()
    registerMetrics(registry)
    http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

    // 启动 HTTP 服务
    go func() {
        log.Println("Agent starting on :9100")
        http.ListenAndServe(":9100", nil)
    }()

    // 定期推送到 Pushgateway
    pushTicker := time.NewTicker(30 * time.Second)
    defer pushTicker.Stop()

    for {
        select {
        case <-pushTicker.C:
            if config.Pushgateway.URL != "" {
                pusher := push.New(config.Pushgateway.URL, config.Pushgateway.Job)
                if err := pusher.Gatherer(registry).Push(); err != nil {
                    log.Printf("Could not push to gateway: %v", err)
                }
            }
        }
    }
}
```

### 后端查询

- 使用 PromQL 从 Prometheus 查询数据，如 `system_cpu_usage_percent{instance="主机名"}`
- 监控中心和主机管理模块共享相同的数据源和 API

---

## 3. 物理主机 vs 云主机

| 维度     | 物理主机         | 云主机          |
| -------- | ---------------- | --------------- |
| 配置信息 | 手动录入         | 云厂商 API 同步 |
| 实时监控 | Agent 采集       | Agent 采集      |
| 数据来源 | 统一 Prometheus  | 统一 Prometheus |

### 云主机同步

- 支持阿里云、腾讯云、百度云、华为云、AWS（部分已实现）
- 支持定时任务配置

---

## 4. Agent 部署方式

### 部署流程

```go
// 1. 生成代码
mainGoContent := agent.CreateAgentMainFile(heartbeatURL, heartbeatToken, pushgatewayURL)
mainGoPath := filepath.Join(tempDir, "main.go")
os.WriteFile(mainGoPath, []byte(mainGoContent), 0644)

// 2. 创建 go.mod
goModContent := `module agent
go 1.24
require (
    github.com/prometheus/client_golang v1.17.0
    github.com/shirou/gopsutil/v3 v3.23.10
)`
goModPath := filepath.Join(tempDir, "go.mod")
os.WriteFile(goModPath, []byte(goModContent), 0644)

// 3. 编译二进制
tidyCmd := exec.Command("go", "mod", "tidy")
tidyCmd.Dir = tempDir
tidyCmd.Run()

buildCmd := exec.Command("go", "build", "-o", binaryName)
buildCmd.Dir = tempDir
buildCmd.Run()
```

步骤说明：
1. **生成代码**：后端动态生成 Agent 的 Go 源代码（main.go 和 go.mod）**agent.go:580-607**
2. **编译二进制**：使用 Go 编译器编译 **agent.go:627-643**
3. **SSH 传输**：拷贝到目标主机 **agent.go:734-790**
4. **启动服务**：systemd 或 nohup 后台运行 **agent.go:836-878**

### SSH 配置和传输

```go
// 获取 SSH 配置
sshConfig = util.SSHConfig{
    IP:       host.SSHIP,
    Port:     host.SSHPort,
    Username: host.SSHName,
    Type:     sshKey.Type,
    Timeout:  30 * time.Second,
}

switch sshKey.Type {
case 1:
    sshConfig.Password = sshKey.Password
case 2:
    sshConfig.PublicKey = sshKey.PublicKey
case 3:
    // 公钥免认证
}

// 创建远程目录
remotePath := "/opt/agent"
createDirCmd := fmt.Sprintf("mkdir -p %s", remotePath)
ExecuteSSHCommand(sshConfig, createDirCmd)

// 拷贝二进制文件
remoteBinaryPath := filepath.Join(remotePath, binaryName)
CopyFileViaSSH(sshConfig, binaryPath, remoteBinaryPath)
```

### systemd 服务配置

```go
serviceContent := `[Unit]
Description=DevOps Monitoring Agent
After=network.target

[Service]
Type=simple
ExecStart=/opt/agent/dodevops-agent
Restart=always
RestartSec=5
User=root
WorkingDirectory=/opt/agent
StandardOutput=append:/var/log/dodevops-agent.log
StandardError=append:/var/log/dodevops-agent.log

[Install]
WantedBy=multi-user.target`

createServiceCmd := fmt.Sprintf("sudo bash -c 'cat > /etc/systemd/system/agent.service << \"EOF\"\n%s\nEOF'", serviceContent)
ExecuteSSHCommand(sshConfig, createServiceCmd)
ExecuteSSHCommand(sshConfig, "sudo systemctl daemon-reload")
ExecuteSSHCommand(sshConfig, "sudo systemctl enable agent.service")
ExecuteSSHCommand(sshConfig, "sudo systemctl start agent.service")
```

### 安装路径

- **Linux 主机**：默认安装路径为 `/opt/agent` **agent.go:147-148**
- **Windows 主机**：默认安装路径为 `C:\dodevops-agent` **agent.go:772-775**

---

## 5. Prometheus 和 Pushgateway 部署方式

### Docker Compose 配置

```yaml
# Pushgateway 服务
pushgateway:
  image: prom/pushgateway:v1.9.0
  container_name: devops-pushgateway
  restart: always
  ports:
    - "${PUSHGATEWAY_PORT:-9091}:9091"
  volumes:
    - ./pushgateway/data:/pushgateway
  networks:
    - devops-network
  healthcheck:
    test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9091/-/healthy"]

# Prometheus 服务
prometheus:
  image: prom/prometheus:v2.47.0
  container_name: devops-prometheus
  restart: always
  ports:
    - "${PROMETHEUS_PORT:-9090}:9090"
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    - ./prometheus/data:/prometheus
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--storage.tsdb.retention.time=15d'
  depends_on:
    pushgateway:
      condition: service_healthy
```

- **Prometheus 服务**：使用镜像 `prometheus:v2.47.0`，监听 9090 端口 **docker-compose.yml:77-108**
- **Pushgateway 服务**：用于接收 Agent 推送的指标，监听 9091 端口 **docker-compose.yml:56-75**
- **数据持久化**：通过 volume 挂载保存 Prometheus 数据 **docker-compose.yml:87-91**
- **健康检查**：两个服务都配置了健康检查机制 **docker-compose.yml:71-75** **docker-compose.yml:104-108**

### 环境变量配置

```bash
# docker/.env
PROMETHEUS_PORT=9090
PUSHGATEWAY_PORT=9091
```

### 配置文件

```yaml
# docker/api/config.yaml
monitor:
  prometheus:
    url: "http://prometheus:9090"
  pushgateway:
    url: "http://pushgateway:9091"
  agent:
    heartbeat_server_url: "http://devops-api:8000/api/v1/monitor/agent/heartbeat"
    heartbeat_token: "agent-heartbeat-token-2024"
```

- Prometheus 和 Pushgateway 的 URL 在 `docker/api/config.yaml` 中配置 **config.yaml:44-51**
- Agent 在编译时会将这些配置写入二进制文件中 **agent.go:911-926**

---

## 6. Agent 前端管理

### Agent 管理页面

前端提供完整的 Agent 管理界面，位于 `web/src/views/Tools/Agent.vue` **Agent.vue:1-120**。

### 主要功能

#### 搜索和筛选

```vue
<el-form :model="queryParams" :inline="true">
  <el-form-item label="主机名称">
    <el-input v-model="queryParams.hostName" placeholder="请输入主机名称" />
  </el-form-item>
  <el-form-item label="Agent状态">
    <el-select v-model="queryParams.status">
      <el-option label="部署中" :value="1" />
      <el-option label="部署失败" :value="2" />
      <el-option label="运行中" :value="3" />
      <el-option label="启动异常" :value="4" />
    </el-select>
  </el-form-item>
</el-form>
```

支持按主机名称、Agent 状态、版本进行筛选 **Agent.vue:10-50**。

#### Agent 列表展示

```vue
<el-table :data="agents" @selection-change="handleSelectionChange">
  <el-table-column type="selection" width="55" />
  <el-table-column prop="hostName" label="主机名称" />
  <el-table-column prop="sshIp" label="IP地址" />
  <el-table-column prop="version" label="版本" />
  <el-table-column prop="status" label="状态" />
  <el-table-column prop="port" label="监听端口" />
</el-table>
```

展示 Agent 的主机名称、IP 地址、版本、状态、监听端口等信息 **Agent.vue:74-120**。

#### 部署操作

```js
const handleDeployToHosts = () => {
  showDeployHostDialog.value = true
}

const handleHostsSelected = async (selectedHosts) => {
  const hostIds = selectedHosts.map(host => host.id)
  await cmdbAPI.deployAgent(hostIds, '1.0.0')
  ElMessage.success(`已开始在 ${selectedHosts.length} 台主机上部署Agent`)
  fetchAgents()
  startPolling()
}
```

点击"部署 Agent"按钮，弹出主机选择对话框，选择主机后调用部署 API **Agent.vue:412-443**。

#### 卸载操作

```js
const handleUninstall = async (row) => {
  await ElMessageBox.confirm(`确定要卸载主机 ${row.hostName} 的Agent吗？`)
  await cmdbAPI.uninstallAgent([row.hostId])
  ElMessage.success('Agent卸载已启动')
  fetchAgents()
  startPolling()
}
```

支持单个卸载和批量卸载 **Agent.vue:445-473** **Agent.vue:523-554**。

#### 重启操作

```js
const handleRestart = async (row) => {
  await ElMessageBox.confirm(`确定要重启主机 ${row.hostName} 的Agent吗？`)
  await cmdbAPI.restartAgent(row.hostId)
  ElMessage.success('Agent重启指令已发送')
  fetchAgents()
  startPolling()
}
```

对启动异常的 Agent 可以执行重启操作 **Agent.vue:497-520**。

#### 删除操作

```js
const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确定要删除主机 ${row.hostName} 的Agent数据吗？`)
  await cmdbAPI.deleteAgent(row.id)
  ElMessage.success('Agent数据删除成功')
  fetchAgents()
}
```

删除 Agent 的数据库记录（用于离线服务器）**Agent.vue:475-495**。

#### 状态轮询

```js
const startPolling = () => {
  pollingTimer = setInterval(() => {
    const hasActiveOperations = agents.value.some(agent => agent.status === 1)
    if (hasActiveOperations) {
      pollingCounter = 0
      fetchAgents()
    } else {
      pollingCounter++
      if (pollingCounter <= MAX_POLLING_COUNT) {
        fetchAgents()
      } else {
        stopPolling()
      }
    }
  }, 3000)
}
```

部署、卸载、重启操作会自动启动轮询，实时更新 Agent 状态 **Agent.vue:691-726**。

---

## 7. 心跳机制

```go
func sendHeartbeat(serverURL, token string) {
    pid := os.Getpid()
    localIP := getLocalIP()

    heartbeat := HeartbeatData{
        PID:      pid,
        IP:       localIP,
        Hostname: hostname,
        Port:     9100,
        Token:    token,
    }

    jsonData, _ := json.Marshal(heartbeat)
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
    if err == nil && resp.StatusCode == 200 {
        log.Printf("Heartbeat sent successfully - PID: %d, IP: %s", pid, localIP)
    }
}
```

Agent 定期向后端发送心跳，报告运行状态 **agent.go:1127-1159**。

---

## 8. Pushgateway 数据清理

```go
func cleanupPushgatewayMetrics(jobName string) {
    pushgatewayURL := "http://8.130.14.34:9091"
    if config.Config != nil && config.Config.Monitor.Pushgateway.URL != "" {
        pushgatewayURL = config.Config.Monitor.Pushgateway.URL
    }

    // Pushgateway API: DELETE /metrics/job/<job_name>
    deleteURL := fmt.Sprintf("%s/metrics/job/%s", pushgatewayURL, jobName)

    req, _ := http.NewRequest("DELETE", deleteURL, nil)
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)

    if resp.StatusCode == 202 || resp.StatusCode == 200 {
        log.Printf("成功清理Pushgateway数据: job=%s", jobName)
    }
}
```

支持清理离线主机的 Pushgateway 数据 **agent.go:404-444**。

---

## 9. 架构设计理由

### 为什么选择自定义 Agent + Pushgateway

| 特性     | 自定义 Agent + Pushgateway    | node-exporter              |
| -------- | ----------------------------- | -------------------------- |
| 网络要求 | Agent 可访问 Pushgateway 即可 | Prometheus 需访问每个主机  |
| 监控范围 | 系统指标 + 业务监控           | 仅系统指标                 |
| 部署方式 | 平台统一部署管理              | 手动部署或配置管理工具     |
| 动态性   | 支持动态增减主机              | 需手动更新 Prometheus 配置 |
| 心跳机制 | 内置心跳报告状态              | 无心跳机制                 |

### 核心优势

1. **网络环境适应性**：解决内网主机、防火墙后主机无法被 Prometheus 直接拉取的问题
2. **扩展监控能力**：提供进程监控、端点监控、Ping 监控等业务监控功能
3. **动态部署与管理**：平台统一管理 Agent 的部署、卸载、重启
4. **云主机场景适配**：更适合云主机频繁增减的动态环境

---

## Notes

- 前端每 10 秒自动刷新监控数据
- Agent 部署需要目标主机配置 SSH 认证信息
