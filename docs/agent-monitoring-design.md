# OneOps Agent 监控增强设计方案

## 文档信息

| 项目     | 内容                                 |
| -------- | ------------------------------------ |
| 文档版本 | v2.0                                 |
| 创建日期 | 2026-05-26                           |
| 更新日期 | 2026-05-26                           |
| 作者     | OneOps 团队                          |
| 状态     | 设计阶段（已更新为独立监控中心架构） |
| 更新内容 | 重新设计为独立的监控中心菜单模块     |

---

## 一、设计背景与目标

### 1.1 背景

当前 OneOps Agent 采集的指标较为简单，仅包含基础的 CPU、内存、磁盘、网络使用率，无法满足企业级运维监控的需求。为了提供更强大的运维支撑能力，需要对 Agent 进行全面增强。

### 1.2 设计目标

构建一个**企业级主机监控与巡检系统**，通过 Agent 采集全方位的主机信息，支撑以下运维场景：

1. **资产自动发现与更新** - CMDB 数据准确性
2. **容量规划与预测** - 资源使用趋势分析
3. **故障快速定位** - 关键指标异常检测
4. **安全合规检查** - 配置基线审计
5. **自动化运维** - 配置变更检测、巡检报表

### 1.3 架构定位

本设计是在现有 Agent 基础上的**增强扩展**，保持向后兼容：

```
现有架构：
┌─────────────────┐         ┌────────────────┐
│   Agent (9MB)   │────────>│  后端服务      │
│  - 基础性能指标  │         │  - 拉取指标    │
│  - 心跳上报     │         │  - 存储       │
└─────────────────┘         └────────────────┘

增强架构：
┌─────────────────────────┐         ┌────────────────────────┐
│  Agent (增强版 ~12MB)    │────────>│   后端服务（增强）      │
│  ✅ 基础性能指标          │         │  📊 多维度指标存储      │
│  ✅ 详细硬件信息          │         │  📈 数据聚合分析        │
│  ✅ 服务与端口监控        │         │  🎯 智能告警引擎        │
│  ✅ 进程Top 20           │         │  📋 告警通知集成        │
│  ✅ 网络配置详情          │         │  📁 归档管理            │
│  ✅ 安全基线检查          │         └────────────────────────┘
│  ✅ 配置变更检测          │                    ↓
└─────────────────────────┘         ┌────────────────────────┐
                                      │    前端（增强）          │
                                      │  🎯 独立监控中心菜单     │
                                      │  📊 监控仪表板           │
                                      │  🔔 告警管理             │
                                      │  📈 趋势分析             │
                                      │  📋 巡检报告             │
                                      │  🔧 监控配置             │
                                      └────────────────────────┘
```

### 1.4 菜单结构设计

采用**独立监控中心**的菜单设计，与资产管理、系统管理平级：

OneOps 菜单结构
│
├── 🏠 资产管理 [现有]
│   ├── 资产总览
│   ├── 主机资产  ←─── 新增"查看监控"快捷入口
│   ├── 访问控制
│   └── 会话审计
│
├── 📊 监控中心 [新增 - 一级菜单]
│   ├── 📈 监控概览
│   │   ├── 总览仪表板
│   │   ├── 主机状态
│   │   └── 资源排行
│   │
│   ├── 🔔 告警管理
│   │   ├── 实时告警
│   │   ├── 告警历史
│   │   ├── 告警规则
│   │   └── 通知渠道
│   │
│   ├── 📉 趋势分析
│   │   ├── 性能趋势
│   │   ├── 容量预测
│   │   └── 对比分析
│   │
│   ├── 🔍 主机监控
│   │   ├── 主机列表（增强版）
│   │   ├── 监控详情
│   │   ├── 进程监控
│   │   └── 服务状态
│   │
│   ├── 🛡️ 安全巡检
│   │   ├── 巡检报告
│   │   ├── 基线检查
│   │   └── 风险分析
│   │
│   └── ⚙️ 监控配置
│       ├── 采集配置
│       ├── 告警配置
│       └── 通知配置
│
├── 🔐 系统管理 [现有]
│   ├── 用户管理
│   ├── 角色管理
│   ├── 菜单管理
│   └── 系统属性
│
└── 📋 系统日志 [现有]
    ├── 登录日志
    └── 操作日志

---

## 二、指标采集优先级

### 2.1 优先级分类

基于运维价值 × 实现复杂度，将采集指标分为四个优先级：

#### 🥇 P0 - 核心功能（第1周，必须实现）

| 指标类别     | 具体指标                     | 运营价值   | 复杂度  | 优先级理由                   |
| ------------ | ---------------------------- | ---------- | ------- | ---------------------------- |
| 性能指标增强 | CPU/内存/磁盘IO/网络流量详情 | ⭐⭐⭐⭐⭐ | 🟢 简单 | 故障诊断基础，实时性要求高   |
| 系统基础信息 | OS版本/内核/主机名/运行时间  | ⭐⭐⭐⭐⭐ | 🟢 简单 | CMDB自动化，资产准确性的基础 |
| 服务状态     | systemd服务状态/监听端口     | ⭐⭐⭐⭐⭐ | 🟢 简单 | 快速判断服务可用性           |
| 网络配置     | IP/MAC/网关/DNS/路由表       | ⭐⭐⭐⭐   | 🟢 简单 | 网络故障排查必需             |

#### 🥈 P1 - 重要功能（第2周，强烈推荐）

| 指标类别     | 具体指标                         | 运营价值 | 复杂度  | 优先级理由             |
| ------------ | -------------------------------- | -------- | ------- | ---------------------- |
| 硬件资产信息 | CPU型号/内存插槽/磁盘型号/序列号 | ⭐⭐⭐⭐ | 🟡 中等 | 资产盘点、硬件故障预警 |
| 进程监控     | Top进程列表(CPU/内存)/进程数统计 | ⭐⭐⭐⭐ | 🟢 简单 | 快速定位资源占用异常   |
| 磁盘分区详情 | 每个分区使用率/inode使用率       | ⭐⭐⭐⭐ | 🟢 简单 | 避免单分区满导致故障   |
| 负载趋势     | Load Average趋势/历史基线对比    | ⭐⭐⭐⭐ | 🟡 中等 | 容量规划、性能瓶颈分析 |

#### 🥉 P2 - 增值功能（第3周，推荐实现）

| 指标类别     | 具体指标                        | 运营价值 | 复杂度  | 优先级理由         |
| ------------ | ------------------------------- | -------- | ------- | ------------------ |
| 安全基线检查 | SSH配置/防火墙状态/登录失败次数 | ⭐⭐⭐   | 🟡 中等 | 安全巡检自动化     |
| 运行时环境   | Python/Java/Go/Docker版本       | ⭐⭐⭐   | 🟡 中等 | 应用部署环境一致性 |
| 系统日志摘要 | 内核错误/OOM Kill/服务异常      | ⭐⭐⭐   | 🔴 复杂 | 故障根因分析       |
| 网络连接状态 | ESTABLISHED/TIME_WAIT连接数     | ⭐⭐⭐   | 🟢 简单 | 网络连接泄漏检测   |

#### 🏅 P3 - 高级功能（第4周，可选实现）

| 指标类别     | 具体指标                   | 运营价值 | 复杂度  | 备注                 |
| ------------ | -------------------------- | -------- | ------- | -------------------- |
| 配置变更检测 | 关键配置文件变更告警       | ⭐⭐⭐⭐ | 🔴 复杂 | 需维护文件清单和基线 |
| 容器监控     | Docker容器状态/资源占用    | ⭐⭐⭐   | 🟡 中等 | 仅在有容器环境时启用 |
| 用户活动审计 | 登录记录/Sudo操作/用户变更 | ⭐⭐⭐   | 🔴 复杂 | 合规要求高时实现     |
| 定时任务监控 | Crontab任务执行状态        | ⭐⭐     | 🟡 中等 | 运维自动化辅助       |

---

## 三、详细指标设计

### 3.1 性能指标（高频采集）

```yaml
performance_metrics:
  # CPU 详细信息
  cpu:
    usage_percent: 45.2              # 总体使用率
    iowait: 2.1                      # IO等待时间
    steal: 0.0                       # 被窃取时间（虚拟机）
    load_avg: [2.5, 3.2, 3.8]       # 1分钟、5分钟、15分钟负载
    per_cpu_usage: [40.5, 48.2, ...] # 每个CPU核心使用率
  
  # 内存详细信息
  memory:
    total: 67308912KB                # 总内存
    used: 45223424KB                 # 已使用
    free: 22085488KB                  # 空闲
    cached: 8364032KB                 # 缓存
    buffers: 1048576KB                # 缓冲区
    swap_total: 8589934592            # 交换空间总量
    swap_used: 0                      # 交换空间使用
    swap_free: 8589934592             # 交换空间空闲
    pages_in: 120                     # 换入页面数
    pages_out: 45                     # 换出页面数
  
  # 磁盘详细信息
  disk:
    partitions:
      - mount: "/"
        usage_percent: 65.2
        inodes_used_percent: 45.3
        inodes_total: 1000000
        inodes_used: 450000
        io_stats:
          read_iops: 150
          write_iops: 80
          read_throughput_mb: 25.5
          write_throughput_mb: 12.3
          await_ms: 8.5
          queue_depth: 2.3
  
  # 网络详细信息
  network:
    interfaces:
      - name: "eth0"
        rx_bytes_mb: 1024
        tx_bytes_mb: 2048
        rx_packets: 1000000
        tx_packets: 800000
        rx_errors: 0
        tx_errors: 0
        rx_dropped: 5
        tx_dropped: 0
        speed: "10000 Mbps"
```

### 3.2 系统基础信息

```yaml
system_info:
  hostname: "web-server-01"
  os:
    family: "redhat"
    name: "CentOS Linux"
    version: "7.9.2009"
    kernel: "3.10.0-1160.el7.x86_64"
    arch: "x86_64"
  uptime: 1234567                  # 运行时长（秒）
  boot_time: "2024-01-15T10:30:00Z"
  timezone: "Asia/Shanghai"
  locale: "zh_CN.UTF-8"
```

### 3.3 硬件资产信息

```yaml
hardware_info:
  cpu:
    model: "Intel(R) Xeon(R) CPU E5-2680 v4"
    cores: 16                        # 物理核心数
    threads: 32                      # 逻辑线程数
    sockets: 2                       # CPU 插槽数
    mhz: 2400                        # 频率
    cache_size: "35840 KB"           # 缓存大小
    vendor_id: "GenuineIntel"
    architecture: "x86_64"
    flags: ["fpu", "vmx", "aes"]     # CPU 特性标志
  
  memory:
    total: 67308912KB
    slots:
      - slot: "DIMM0"
        size: "16GB"
        type: "DDR4"
        speed: "2666MT/s"
        vendor: "Samsung"
      - slot: "DIMM1"
        size: "16GB"
        type: "DDR4"
        speed: "2666MT/s"
        vendor: "Samsung"
    ecc_enabled: true                # 是否启用 ECC
  
  disk:
    devices:
      - name: "/dev/sda"
        model: "Samsung SSD 850"
        size: "500G"
        type: "SSD"
        serial: "S1ZWNBHC123456"
        firmware: "2B6Q"
        smart_status: "healthy"
  
  system:
    manufacturer: "Dell Inc."
    product_name: "PowerEdge R730"
    serial_number: "ABC123"
    chassis_type: "Rack Mount Chassis"
    bios_version: "2.8.1"
    bios_release_date: "2020-03-15"
```

### 3.4 服务与端口监控

```yaml
service_status:
  systemd_services:
    - name: "nginx"
      state: "active"
      status: "running"
    - name: "mysql"
      state: "active"
      status: "running"
    - name: "docker"
      state: "active"
      status: "running"
    - name: "cron"
      state: "inactive"
      enabled: true
  
  listening_ports:
    - port: 22
      protocol: "tcp"
      process: "sshd"
      address: "0.0.0.0"
    - port: 80
      protocol: "tcp"
      process: "nginx"
      address: "0.0.0.0"
    - port: 3306
      protocol: "tcp"
      process: "mysqld"
      address: "0.0.0.0"
```

### 3.5 进程监控（Top 20）

```yaml
process_info:
  total: 245                        # 总进程数
  running: 3                        # 运行中
  sleeping: 240                     # 睡眠中
  
  top_cpu:
    - pid: 1234
      name: "java"
      user: "app"
      cpu_percent: 25.5
      memory_percent: 15.2
      memory_mb: 10240
      cmdline: "/usr/bin/java -jar app.jar"
      start_time: "2024-05-26T08:00:00Z"
    - pid: 5678
      name: "nginx"
      user: "nginx"
      cpu_percent: 5.2
      memory_percent: 2.1
      memory_mb: 512
      cmdline: "nginx: worker process"
  
  top_memory:
    - pid: 1234
      name: "java"
      cpu_percent: 25.5
      memory_percent: 15.2
      memory_mb: 10240
```

### 3.6 网络配置

```yaml
network_config:
  interfaces:
    - name: "eth0"
      ipv4: ["192.168.1.100/24"]
      ipv6: ["fe80::xxxx/64"]
      mac: "00:16:3e:xx:xx:xx"
      mtu: 1500
      speed: "10000 Mbps"
      duplex: "full"
      carrier: "on"
      driver: "ixgbe"
      gateway: "192.168.1.1"
      dns: ["8.8.8.8", "114.114.114.114"]
  
  routes:
    - destination: "0.0.0.0/0"
      gateway: "192.168.1.1"
      iface: "eth0"
    - destination: "10.0.0.0/8"
      gateway: "192.168.1.254"
      iface: "eth0"
  
  arp_table:
    - ip: "192.168.1.1"
      mac: "aa:bb:cc:dd:ee:ff"
      iface: "eth0"
```

### 3.7 安全基线检查

```yaml
security_info:
  ssh_config:
    password_auth: false             # 是否允许密码登录
    root_login: false                 # 是否允许 root 登录
    port: 22
    key_algorithms: ["rsa-sha2-256", "ecdsa-sha2-nistp256"]
  
  firewall:
    backend: "firewalld"
    enabled: true
    zones: ["public", "trusted"]
    rules_count: 25
  
  selinux:
    enabled: true
    mode: "enforcing"
  
  users:
    total: 10
    with_sudo: 2                     # 有 sudo 权限的用户数
  
  auth_failures: 15                  # 最近1小时认证失败次数
```

---

## 四、数据存储与归档方案

### 4.1 存储策略

```yaml
storage_policy:
  # 热数据保留期（在线查询）
  hot_retention_days: 30
  
  # 温数据保留期（归档存储，可快速恢复）
  warm_retention_days: 90
  
  # 冷数据保留期（压缩存储，长期保留）
  cold_retention_days: 365
  
  # 数据清理间隔
  cleanup_interval: 24h
```

### 4.2 数据库设计

```sql
-- 指标数据表（热数据）
CREATE TABLE agent_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL,
    metric_type VARCHAR(50) NOT NULL,  -- performance, system, asset, security, process
    metric_data JSON NOT NULL,
    report_time DATETIME NOT NULL,
    received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_server_time (server_id, report_time),
    INDEX idx_type_time (metric_type, report_time),
    INDEX idx_report_time (report_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
PARTITION BY RANGE (TO_DAYS(report_time)) (
    PARTITION p_2024_05 VALUES LESS THAN (TO_DAYS('2024-06-01')),
    PARTITION p_2024_06 VALUES LESS THAN (TO_DAYS('2024-07-01')),
    PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 归档表（冷数据）
CREATE TABLE agent_metrics_archive (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    metric_data JSON NOT NULL,
    report_time DATETIME NOT NULL,
    archived_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_server_time (server_id, report_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
PARTITION BY RANGE (TO_DAYS(report_time)) (
    PARTITION p_2024_01 VALUES LESS THAN (TO_DAYS('2024-02-01')),
    PARTITION p_future VALUES LESS THAN MAXVALUE
);

-- 告警表
CREATE TABLE agent_alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL,
    rule_id VARCHAR(50) NOT NULL,
    level VARCHAR(20) NOT NULL,       -- critical, high, medium, low, info
    message VARCHAR(500) NOT NULL,
    metric_value DECIMAL(10,2),
    threshold DECIMAL(10,2),
    first_seen DATETIME NOT NULL,
    last_seen DATETIME NOT NULL,
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by VARCHAR(100),
    acknowledged_at DATETIME,
    resolved_at DATETIME,
    INDEX idx_server_level (server_id, level),
    INDEX idx_acknowledged (acknowledged, first_seen),
    INDEX idx_first_seen (first_seen)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 告警规则配置表
CREATE TABLE agent_alert_rules (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    level VARCHAR(20) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    condition VARCHAR(10) NOT NULL,   -- >, <, ==, !=
    threshold DECIMAL(10,2) NOT NULL,
    duration INT NOT NULL,            -- 持续时间（秒）
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 通知渠道配置表
CREATE TABLE notification_channels (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    channel_type VARCHAR(20) NOT NULL, -- wechat, dingtalk, email, sms
    channel_name VARCHAR(100) NOT NULL,
    config JSON NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 4.3 数据归档实现

```go
// 数据生命周期管理
type DataLifecycleManager struct{}

func (m *DataLifecycleManager) Start() {
    ticker := time.NewTicker(24 * time.Hour)
    go func() {
        for range ticker.C {
            m.ArchiveOldMetrics()
            m.CleanupExpiredMetrics()
        }
    }()
}

// 归档30天前的数据到归档表
func (m *DataLifecycleManager) ArchiveOldMetrics() {
    cutoffDate := time.Now().AddDate(0, 0, -HotRetentionDays)
  
    // 将热数据迁移到归档表
    db.Exec(`
        INSERT INTO agent_metrics_archive 
        (server_id, metric_type, metric_data, report_time)
        SELECT server_id, metric_type, metric_data, report_time
        FROM agent_metrics
        WHERE report_time < ?
    `, cutoffDate)
  
    // 删除已归档的热数据
    db.Exec("DELETE FROM agent_metrics WHERE report_time < ?", cutoffDate)
  
    logger.Info("指标数据归档完成", zap.Time("cutoffDate", cutoffDate))
}

// 清理过期数据（>365天）
func (m *DataLifecycleManager) CleanupExpiredMetrics() {
    cutoffDate := time.Now().AddDate(0, 0, -ColdRetentionDays)
  
    result := db.Exec("DELETE FROM agent_metrics_archive WHERE report_time < ?", cutoffDate)
    logger.Info("过期指标数据清理完成", 
        zap.Int64("deleted", result.RowsAffected),
        zap.Time("cutoffDate", cutoffDate))
}
```

---

## 五、告警规则设计

### 5.1 告警级别定义

```go
const (
    AlertLevelCritical = "critical"  // 严重：立即处理，影响业务
    AlertLevelHigh     = "high"      // 高级：1小时内处理
    AlertLevelMedium   = "medium"    // 中级：当天处理
    AlertLevelLow      = "low"       // 低级：知晓即可
    AlertLevelInfo     = "info"      // 信息：仅供参考
)
```

### 5.2 核心告警规则

#### 🔴 Critical 级别（立即通知）

| 告警名称     | 触发条件             | 持续时间 | 通知方式       | 恢复条件       |
| ------------ | -------------------- | -------- | -------------- | -------------- |
| 主机宕机     | Agent心跳丢失        | 3分钟    | 电话+短信+邮件 | Agent恢复心跳  |
| CPU持续100%  | CPU使用率 > 95%      | 10分钟   | 电话+短信+邮件 | < 80%          |
| 内存不足     | 内存使用率 > 95%     | 5分钟    | 电话+短信+邮件 | < 85%          |
| 磁盘满       | 任意分区使用率 > 95% | 立即     | 电话+短信+邮件 | < 90%          |
| OOM Kill     | 检测到OOM事件        | 立即     | 电话+短信+邮件 | -              |
| 系统负载过高 | Load5 > CPU核心数×4 | 15分钟   | 短信+邮件      | < CPU核心数×2 |

#### 🟠 High 级别（1小时内处理）

| 告警名称       | 触发条件                      | 持续时间 | 通知方式  | 恢复条件     |
| -------------- | ----------------------------- | -------- | --------- | ------------ |
| CPU持续高负载  | CPU使用率 > 80%               | 30分钟   | 短信+邮件 | < 70%        |
| 内存使用率告警 | 内存使用率 > 85%              | 30分钟   | 短信+邮件 | < 75%        |
| 磁盘空间不足   | 任意分区使用率 > 85%          | 立即     | 短信+邮件 | < 80%        |
| Inode耗尽      | Inode使用率 > 90%             | 立即     | 短信+邮件 | < 85%        |
| 核心服务停止   | Nginx/MySQL/Redis停止         | 5分钟    | 短信+邮件 | 服务恢复     |
| 磁盘IO异常     | await > 50ms 或 %iowait > 20% | 20分钟   | 短信+邮件 | await < 20ms |

#### 🟡 Medium 级别（当天处理）

| 告警名称       | 触发条件             | 持续时间 | 通知方式 | 恢复条件 |
| -------------- | -------------------- | -------- | -------- | -------- |
| CPU使用率偏高  | CPU使用率 > 60%      | 1小时    | 邮件     | < 50%    |
| 内存使用率偏高 | 内存使用率 > 75%     | 1小时    | 邮件     | < 65%    |
| 磁盘增长过快   | 日增长率 > 10%       | -        | 邮件     | -        |
| 网络流量异常   | 流量超过基线3倍      | 30分钟   | 邮件     | 恢复正常 |
| 登录失败频繁   | 1小时内失败 > 50次   | -        | 邮件     | -        |
| 进程数异常     | 进程数 > 1000        | 30分钟   | 邮件     | < 800    |
| TIME_WAIT过多  | TIME_WAIT连接 > 5000 | 30分钟   | 邮件     | < 3000   |

### 5.3 告警规则配置示例

```go
type AlertRule struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Level       string            `json:"level"`
    Metric      string            `json:"metric"`
    Condition   string            `json:"condition"`
    Threshold   float64           `json:"threshold"`
    Duration    time.Duration     `json:"duration"`
    Description string            `json:"description"`
    Actions     []AlertAction     `json:"actions"`
}

var defaultAlertRules = []AlertRule{
    {
        ID:          "cpu_critical",
        Name:        "CPU严重告警",
        Level:       AlertLevelCritical,
        Metric:      "cpu_usage",
        Condition:   ">",
        Threshold:   95.0,
        Duration:    10 * time.Minute,
        Description: "CPU使用率持续超过95%，可能影响业务性能",
    },
    {
        ID:          "memory_high",
        Name:        "内存使用率过高",
        Level:       AlertLevelHigh,
        Metric:      "memory_usage",
        Condition:   ">",
        Threshold:   85.0,
        Duration:    30 * time.Minute,
        Description: "内存使用率超过85%，需关注",
    },
    {
        ID:          "disk_full",
        Name:        "磁盘空间告警",
        Level:       AlertLevelCritical,
        Metric:      "disk_usage",
        Condition:   ">",
        Threshold:   95.0,
        Duration:    0,
        Description: "磁盘空间不足95%，可能导致服务异常",
    },
}
```

---

## 六、告警通知集成

### 6.1 通知渠道配置

```yaml
notification:
  enabled: true
  silent_hours:
    - "22:00-08:00"  # 夜间免打扰（严重告警除外）
  
  # 企业微信配置
  wechat:
    corp_id: "ww1234567890abcdef"
    corp_secret: "xxxxxxxxxxxxxxxxxx"
    agent_id: 1000001
    webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
  
  # 钉钉配置
  dingtalk:
    webhook_url: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
    secret: "SECxxxxxxxxxxxxxxxxxx"
  
  # 邮件配置
  email:
    smtp_host: "smtp.example.com"
    smtp_port: 587
    username: "alerts@oneops.com"
    password: "xxxxxxxx"
    from: "OneOps告警系统 <alerts@oneops.com>"
    to:
      - "ops-team@example.com"
  
  # 短信配置
  sms:
    provider: "aliyun"
    access_key: "xxxxxxxx"
    secret_key: "xxxxxxxx"
    sign_name: "OneOps运维"
    template_code: "SMS_123456789"
    phones:
      - "13800138000"
  
  # 告警路由
  routing:
    critical: ["wechat", "dingtalk", "email", "sms"]
    high: ["wechat", "dingtalk", "email"]
    medium: ["wechat", "email"]
    low: ["email"]
    info: ["email"]
```

### 6.2 消息格式设计

#### 企业微信消息格式

```
🔴 【OneOps告警】
告警级别：严重
主机名称：web-server-01
主机IP：192.168.1.100
告警内容：CPU使用率持续超过95%，可能影响业务性能
当前值：98.50
阈值：95.00
发生时间：2026-05-26 12:00:00
请及时处理！
```

#### 钉钉消息格式

```markdown
### <font color=#FF0000>严重告警</font>

#### 主机信息
> 主机名称：web-server-01
> 主机IP：192.168.1.100

#### 告警详情
> 告警内容：CPU使用率持续超过95%，可能影响业务性能
> 当前值：`98.50`
> 阈值：`95.00`

#### 发生时间
> 2026-05-26 12:00:00
```

---

## 七、前端展示设计（监控中心架构）

### 7.1 监控中心菜单结构

```
📊 监控中心（一级菜单）
│
├── 📈 监控概览 (monitoring/overview) [默认页]
│   ├── 总览仪表板
│   │   ├── 主机状态卡片（总数/在线/离线/告警）
│   │   ├── 资源使用率卡片（CPU/内存/磁盘平均值）
│   │   └── 告警统计卡片（各级别告警数量）
│   ├── 资源使用 Top 10
│   │   ├── CPU 使用率排行
│   │   ├── 内存使用率排行
│   │   └── 磁盘使用率排行
│   └── 实时告警列表（最新10条）
│
├── 🔔 告警管理 (monitoring/alerts)
│   ├── 实时告警
│   │   ├── 告警列表（可筛选级别/主机/时间）
│   │   ├── 告警详情（查看主机监控）
│   │   └── 批量操作（确认/导出）
│   ├── 告警历史
│   │   ├── 历史告警查询
│   │   ├── 告警趋势图
│   │   └── 告警统计分析
│   ├── 告警规则
│   │   ├── 规则列表
│   │   ├── 规则配置（新增/编辑）
│   │   └── 规则启用/禁用
│   └── 通知渠道
│       ├── 渠道列表（微信/钉钉/邮件/短信）
│       ├── 渠道配置
│       └── 测试通知
│
├── 📉 趋势分析 (monitoring/trends)
│   ├── 性能趋势
│   │   ├── CPU/内存/磁盘/网络趋势图
│   │   ├── 时间范围选择（1h/6h/24h/7d/30d）
│   │   └── 多主机对比
│   ├── 容量预测
│   │   ├── 磁盘空间预测
│   │   ├── 内存使用预测
│   │   └── 容量告警阈值建议
│   └── 对比分析
│       ├── 环比分析（与上周对比）
│       └── 同比分析（与去年同期对比）
│
├── 🔍 主机监控 (monitoring/servers)
│   ├── 主机列表（增强版）
│   │   ├── 表格列：主机名/IP/状态/CPU/内存/磁盘/告警
│   │   ├── 迷你趋势图（每台主机最近1小时）
│   │   ├── 服务状态图标
│   │   ├── 告警徽章
│   │   └── 快捷操作（查看监控/连接/编辑）
│   ├── 监控详情
│   │   ├── 实时性能仪表盘
│   │   ├── 服务状态监控
│   │   ├── 进程监控（Top 20）
│   │   └── 网络连接状态
│   ├── 硬件资产
│   │   ├── CPU/内存/磁盘/网卡详情
│   │   ├── 序列号/固件版本
│   │   └── 厂商/型号信息
│   ├── 网络配置
│   │   ├── 网卡列表
│   │   ├── IP配置
│   │   ├── 路由表
│   │   └── DNS配置
│   └── 安全巡检
│       ├── 安全基线检查结果
│       ├── SSH配置审计
│       ├── 防火墙状态
│       └── 登录审计
│
├── 🛡️ 巡检报告 (monitoring/reports)
│   ├── 主机巡检报告
│   │   ├── 巡检任务配置
│   │   ├── 巡检执行记录
│   │   ├── 巡检结果详情
│   │   └── 导出报告（PDF/Excel）
│   ├── 安全合规报告
│   │   ├── 合规检查项配置
│   │   ├── 检查结果汇总
│   │   └── 风险等级分析
│   └── 容量分析报告
│       ├── 资源使用趋势
│       ├── 容量预测
│       └── 扩容建议
│
└── ⚙️ 监控配置 (monitoring/settings)
    ├── 采集配置
    │   ├── 采集频率设置
    │   ├── 指标启用/禁用
    │   └── 数据保留策略
    ├── 告警配置
    │   ├── 全局告警设置
    │   ├── 免打扰时间配置
    │   └── 告警聚合策略
    └── 通知配置
        ├── 通知渠道配置
        ├── 通知模板管理
        └── 通知路由规则
```

### 7.3 监控概览页面设计

#### 页面结构

```vue
<template>
  <div class="monitoring-overview">
    <!-- 总览卡片 -->
    <div class="overview-cards">
      <div class="stat-card online">
        <div class="stat-icon">
          <icon-mdi-server-network />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.onlineServers }}</div>
          <div class="stat-label">在线主机</div>
          <div class="stat-rate">/{{ stats.totalServers }}</div>
        </div>
      </div>
    
      <div class="stat-card offline">
        <div class="stat-icon">
          <icon-mdi-server-off />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.offlineServers }}</div>
          <div class="stat-label">离线主机</div>
        </div>
      </div>
    
      <div class="stat-card alert-critical">
        <div class="stat-icon">
          <icon-mdi-alert-circle />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.criticalAlerts }}</div>
          <div class="stat-label">严重告警</div>
        </div>
      </div>
    
      <div class="stat-card avg-cpu">
        <div class="stat-icon">
          <icon-mdi-cpu-64-bit />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.avgCpuUsage.toFixed(1) }}%</div>
          <div class="stat-label">平均CPU</div>
        </div>
      </div>
    
      <div class="stat-card avg-memory">
        <div class="stat-icon">
          <icon-mdi-memory />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.avgMemoryUsage.toFixed(1) }}%</div>
          <div class="stat-label">平均内存</div>
        </div>
      </div>
    
      <div class="stat-card avg-disk">
        <div class="stat-icon">
          <icon-mdi-harddisk />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.avgDiskUsage.toFixed(1) }}%</div>
          <div class="stat-label">平均磁盘</div>
        </div>
      </div>
    </div>
  
    <!-- Top 10 资源使用 -->
    <div class="top-resources">
      <div class="chart-panel">
        <div class="panel-header">
          <h3>CPU 使用率 Top 10</h3>
          <ElButton size="small" @click="refreshTopCpu">
            <icon-mdi-refresh />
          </ElButton>
        </div>
        <TopBarChart :data="topCpuServers" unit="%" :height="300" />
      </div>
    
      <div class="chart-panel">
        <div class="panel-header">
          <h3>内存使用率 Top 10</h3>
          <ElButton size="small" @click="refreshTopMemory">
            <icon-mdi-refresh />
          </ElButton>
        </div>
        <TopBarChart :data="topMemoryServers" unit="%" :height="300" />
      </div>
    
      <div class="chart-panel">
        <div class="panel-header">
          <h3>磁盘使用率 Top 10</h3>
          <ElButton size="small" @click="refreshTopDisk">
            <icon-mdi-refresh />
          </ElButton>
        </div>
        <TopBarChart :data="topDiskServers" unit="%" :height="300" />
      </div>
    </div>
  
    <!-- 实时告警列表 -->
    <div class="recent-alerts">
      <div class="panel-header">
        <h3>实时告警</h3>
        <div class="header-actions">
          <ElButton size="small" @click="refreshAlerts">
            <icon-mdi-refresh />
            刷新
          </ElButton>
          <ElButton size="small" @click="viewAllAlerts">
            查看全部
            <icon-mdi-arrow-right />
          </ElButton>
        </div>
      </div>
    
      <ElTable :data="recentAlerts" v-loading="alertsLoading" size="small">
        <ElTableColumn label="级别" width="80">
          <template #default="{ row }">
            <ElTag :type="getAlertTagType(row.level)" size="small">
              {{ getAlertLevelLabel(row.level) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="hostname" label="主机" width="150" />
        <ElTableColumn prop="message" label="告警内容" min-width="200" />
        <ElTableColumn prop="metricValue" label="当前值" width="100" />
        <ElTableColumn prop="timestamp" label="发生时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.timestamp) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <ElButton size="small" @click="viewServerDetail(row)">
              查看主机
            </ElButton>
            <ElButton size="small" type="primary" @click="acknowledgeAlert(row)">
              确认
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </div>
  </div>
</template>
```

### 7.4 主机监控页面设计

#### 主机列表（监控视图）

```vue
<template>
  <div class="monitoring-servers">
    <!-- 筛选栏 -->
    <div class="filter-bar">
      <ElForm inline>
        <ElFormItem label="主机状态">
          <ElSelect v-model="filters.status" placeholder="全部">
            <ElOption label="全部" value="" />
            <ElOption label="在线" value="online" />
            <ElOption label="离线" value="offline" />
            <ElOption label="告警中" value="alert" />
          </ElSelect>
        </ElFormItem>
      
        <ElFormItem label="告警级别">
          <ElSelect v-model="filters.alertLevel" placeholder="全部">
            <ElOption label="全部" value="" />
            <ElOption label="严重" value="critical" />
            <ElOption label="高级" value="high" />
            <ElOption label="中级" value="medium" />
          </ElSelect>
        </ElFormItem>
      
        <ElFormItem label="资源使用率">
          <ElInput v-model="filters.minUsage" placeholder="最小使用率" style="width: 120px" />
          <span>-</span>
          <ElInput v-model="filters.maxUsage" placeholder="最大使用率" style="width: 120px" />
          <ElSelect v-model="filters.metricType" style="width: 100px">
            <ElOption label="CPU" value="cpu" />
            <ElOption label="内存" value="memory" />
            <ElOption label="磁盘" value="disk" />
          </ElSelect>
        </ElFormItem>
      
        <ElFormItem>
          <ElButton type="primary" @click="searchServers">查询</ElButton>
          <ElButton @click="resetFilters">重置</ElButton>
          <ElButton @click="exportData">导出</ElButton>
        </ElFormItem>
      </ElForm>
    </div>
  
    <!-- 主机列表表格 -->
    <ElTable :data="servers" v-loading="loading" stripe>
      <ElTableColumn type="selection" width="55" />
    
      <ElTableColumn prop="hostname" label="主机名" width="150">
        <template #default="{ row }">
          <div class="hostname-cell">
            <span>{{ row.hostname }}</span>
            <ElTag v-if="row.agentStatus === 'offline'" type="danger" size="small">
              离线
            </ElTag>
          </div>
        </template>
      </ElTableColumn>
    
      <ElTableColumn prop="ip" label="IP地址" width="140" />
    
      <ElTableColumn label="性能趋势" width="200">
        <template #default="{ row }">
          <div class="mini-trends">
            <div class="trend-item">
              <span class="trend-label">CPU</span>
              <MiniSparkLine :data="row.cpuHistory" :height="24" />
              <span class="trend-value">{{ row.cpuUsage?.toFixed(1) || 0 }}%</span>
            </div>
            <div class="trend-item">
              <span class="trend-label">MEM</span>
              <MiniSparkLine :data="row.memoryHistory" :height="24" />
              <span class="trend-value">{{ row.memoryUsage?.toFixed(1) || 0 }}%</span>
            </div>
          </div>
        </template>
      </ElTableColumn>
    
      <ElTableColumn label="服务状态" width="150">
        <template #default="{ row }">
          <div class="services-cell">
            <ElTag 
              v-for="service in row.criticalServices" 
              :key="service.name"
              :type="service.status === 'running' ? 'success' : 'danger'"
              size="small"
              class="service-tag"
            >
              {{ service.name }}
            </ElTag>
            <ElTooltip v-if="row.serviceCount > 3" :content="`共${row.serviceCount}个服务`">
              <ElTag size="small" type="info">+{{ row.serviceCount - 3 }}</ElTag>
            </ElTooltip>
          </div>
        </template>
      </ElTableColumn>
    
      <ElTableColumn label="告警" width="80" align="center">
        <template #default="{ row }">
          <ElBadge 
            v-if="row.alertCount > 0" 
            :value="row.alertCount" 
            :max="99"
            :type="getAlertBadgeType(row.maxAlertLevel)"
          />
          <span v-else>-</span>
        </template>
      </ElTableColumn>
    
      <ElTableColumn label="操作" width="180" align="center" fixed="right">
        <template #default="{ row }">
          <ElButton size="small" type="primary" @click="viewMonitoring(row)">
            监控详情
          </ElButton>
          <ElDropdown @command="handleMoreAction">
            <ElButton size="small">
              更多
              <icon-mdi-chevron-down />
            </ElButton>
            <template #dropdown>
              <ElDropdownItem command="connect">连接主机</ElDropdownItem>
              <ElDropdownItem command="edit">编辑主机</ElDropdownItem>
              <ElDropdownItem command="viewHistory">查看历史</ElDropdownItem>
              <ElDropdownItem command="viewAlerts">告警历史</ElDropdownItem>
            </template>
          </ElDropdown>
        </template>
      </ElTableColumn>
    </ElTable>
  </div>
</template>
```

#### 主机监控详情页

```vue
<template>
  <div class="server-monitoring-detail">
    <!-- 主机基本信息头 -->
    <div class="server-header">
      <div class="server-info">
        <icon-mdi-server class="server-icon" :class="getStatusClass(server.status)" />
        <div class="server-text">
          <h2>{{ server.hostname }}</h2>
          <div class="server-meta">
            <span>{{ server.ip }}</span>
            <span>•</span>
            <span>{{ server.os.name }} {{ server.os.version }}</span>
            <span>•</span>
            <span>运行时间: {{ formatUptime(server.uptime) }}</span>
          </div>
        </div>
      </div>
    
      <div class="header-actions">
        <ElButton @click="refreshData">
          <icon-mdi-refresh />
          刷新
        </ElButton>
        <ElButton @click="connectServer">
          <icon-mdi-console />
          连接
        </ElButton>
        <ElButton @click="viewAlerts">
          <icon-mdi-bell />
          告警历史
        </ElButton>
      </div>
    </div>
  
    <!-- 标签页 -->
    <ElTabs v-model="activeTab" type="card">
      <!-- 监控概览 Tab -->
      <ElTabPane label="监控概览" name="overview">
        <div class="monitor-overview-content">
          <!-- 实时性能卡片 -->
          <div class="performance-cards">
            <div class="perf-card cpu">
              <div class="card-header">
                <span class="card-title">CPU使用率</span>
                <ElTag :type="getCpuTagType(currentMetrics.cpuUsage)">
                  {{ currentMetrics.cpuUsage?.toFixed(1) || 0 }}%
                </ElTag>
              </div>
              <div class="card-chart">
                <GaugeChart :value="currentMetrics.cpuUsage" :height="120" />
              </div>
              <div class="card-details">
                <div class="detail-item">
                  <span class="label">Load Average:</span>
                  <span class="value">{{ currentMetrics.load1?.toFixed(2) || 0 }} / {{ currentMetrics.load5?.toFixed(2) || 0 }} / {{ currentMetrics.load15?.toFixed(2) || 0 }}</span>
                </div>
              </div>
            </div>
          
            <div class="perf-card memory">
              <div class="card-header">
                <span class="card-title">内存使用率</span>
                <ElTag :type="getMemoryTagType(currentMetrics.memoryUsage)">
                  {{ currentMetrics.memoryUsage?.toFixed(1) || 0 }}%
                </ElTag>
              </div>
              <div class="card-chart">
                <GaugeChart :value="currentMetrics.memoryUsage" :height="120" />
              </div>
              <div class="card-details">
                <div class="detail-item">
                  <span class="label">已用:</span>
                  <span class="value">{{ formatBytes(currentMetrics.memoryUsed) }} / {{ formatBytes(currentMetrics.memoryTotal) }}</span>
                </div>
              </div>
            </div>
          
            <div class="perf-card disk">
              <div class="card-header">
                <span class="card-title">磁盘使用率</span>
                <ElTag :type="getDiskTagType(currentMetrics.diskUsage)">
                  {{ currentMetrics.diskUsage?.toFixed(1) || 0 }}%
                </ElTag>
              </div>
              <div class="card-chart">
                <DonutChart :data="currentMetrics.diskPartitions" :height="120" />
              </div>
            </div>
          
            <div class="perf-card network">
              <div class="card-header">
                <span class="card-title">网络流量</span>
              </div>
              <div class="card-details">
                <div class="detail-item">
                  <span class="label">↓ 接收:</span>
                  <span class="value">{{ formatBytes(currentMetrics.networkRx) }}/s</span>
                </div>
                <div class="detail-item">
                  <span class="label">↑ 发送:</span>
                  <span class="value">{{ formatBytes(currentMetrics.networkTx) }}/s</span>
                </div>
              </div>
            </div>
          </div>
        
          <!-- Top 20 进程 -->
          <div class="top-processes">
            <div class="section-header">
              <h3>Top 20 进程</h3>
              <ElRadioGroup v-model="processSortBy" size="small">
                <ElRadioButton label="cpu">按CPU</ElRadioButton>
                <ElRadioButton label="memory">按内存</ElRadioButton>
              </ElRadioGroup>
            </div>
          
            <ElTable :data="topProcesses" size="small" max-height="400px">
              <ElTableColumn prop="pid" label="PID" width="80" />
              <ElTableColumn prop="name" label="进程名" width="150" />
              <ElTableColumn prop="user" label="用户" width="100" />
              <ElTableColumn label="CPU" width="120">
                <template #default="{ row }">
                  <ElProgress 
                    :percentage="row.cpuPercent" 
                    :color="getProcessColor(row.cpuPercent)"
                    :show-text="true"
                  />
                </template>
              </ElTableColumn>
              <ElTableColumn label="内存" width="120">
                <template #default="{ row }">
                  {{ formatBytes(row.memoryMB * 1024 * 1024) }}
                  <span class="process-percent">({{ row.memoryPercent.toFixed(1) }}%)</span>
                </template>
              </ElTableColumn>
              <ElTableColumn prop="cmdline" label="命令行" min-width="200" show-overflow-tooltip />
            </ElTable>
          </div>
        </div>
      </ElTabPane>
    
      <!-- 服务状态 Tab -->
      <ElTabPane label="服务状态" name="services">
        <div class="services-content">
          <!-- systemd 服务列表 -->
          <div class="systemd-services">
            <h3>系统服务</h3>
            <ElTable :data="systemdServices" size="small">
              <ElTableColumn prop="name" label="服务名" width="200" />
              <ElTableColumn prop="description" label="描述" min-width="200" />
              <ElTableColumn label="状态" width="100">
                <template #default="{ row }">
                  <ElTag :type="getServiceTagType(row.state)" size="small">
                    {{ row.status }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn label="启用" width="80">
                <template #default="{ row }">
                  <ElTag :type="row.enabled ? 'success' : 'info'" size="small">
                    {{ row.enabled ? '是' : '否' }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn label="操作" width="150">
                <template #default="{ row }">
                  <ElButton size="small" @click="startService(row)" :disabled="row.status === 'running'">
                    启动
                  </ElButton>
                  <ElButton size="small" @click="stopService(row)" :disabled="row.status !== 'running'">
                    停止
                  </ElButton>
                  <ElButton size="small" @click="restartService(row)">
                    重启
                  </ElButton>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        
          <!-- 监听端口列表 -->
          <div class="listening-ports">
            <h3>监听端口</h3>
            <ElTable :data="listeningPorts" size="small">
              <ElTableColumn prop="port" label="端口" width="100" />
              <ElTableColumn prop="protocol" label="协议" width="80" />
              <ElTableColumn prop="address" label="监听地址" width="150" />
              <ElTableColumn prop="process" label="进程" width="150" />
              <ElTableColumn label="操作" width="100">
                <template #default="{ row }">
                  <ElButton size="small" @click="viewProcessInfo(row)">
                    详情
                  </ElButton>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        </div>
      </ElTabPane>
    
      <!-- 硬件资产 Tab -->
      <ElTabPane label="硬件资产" name="hardware">
        <div class="hardware-content">
          <!-- CPU 详情 -->
          <div class="hardware-section">
            <h3>CPU 信息</h3>
            <ElDescriptions :column="2" border>
              <ElDescriptionsItem label="型号">{{ hardwareInfo.cpu.model }}</ElDescriptionsItem>
              <ElDescriptionsItem label="架构">{{ hardwareInfo.cpu.architecture }}</ElDescriptionsItem>
              <ElDescriptionsItem label="物理核心">{{ hardwareInfo.cpu.cores }} 核</ElDescriptionsItem>
              <ElDescriptionsItem label="逻辑线程">{{ hardwareInfo.cpu.threads }} 线程</ElDescriptionsItem>
              <ElDescriptionsItem label="插槽">{{ hardwareInfo.cpu.sockets }}</ElDescriptionsItem>
              <ElDescriptionsItem label="频率">{{ hardwareInfo.cpu.mhz }} MHz</ElDescriptionsItem>
              <ElDescriptionsItem label="缓存" :span="2">{{ hardwareInfo.cpu.cacheSize }}</ElDescriptionsItem>
            </ElDescriptions>
          </div>
        
          <!-- 内存详情 -->
          <div class="hardware-section">
            <h3>内存信息</h3>
            <ElDescriptions :column="2" border>
              <ElDescriptionsItem label="总内存" :span="2">{{ formatBytes(hardwareInfo.memory.total) }}</ElDescriptionsItem>
              <ElDescriptionsItem label="插槽数">{{ hardwareInfo.memory.slots.length }}</ElDescriptionsItem>
              <ElDescriptionsItem label="ECC">{{ hardwareInfo.memory.eccEnabled ? '启用' : '未启用' }}</ElDescriptionsItem>
            </ElDescriptions>
          
            <ElTable :data="hardwareInfo.memory.slots" size="small" class="mt-16px">
              <ElTableColumn prop="slot" label="插槽" width="100" />
              <ElTableColumn prop="size" label="容量" width="100" />
              <ElTableColumn prop="type" label="类型" width="100" />
              <ElTableColumn prop="speed" label="速度" width="120" />
              <ElTableColumn prop="vendor" label="厂商" />
            </ElTable>
          </div>
        </div>
      </ElTabPane>
    
      <!-- 网络配置 Tab -->
      <ElTabPane label="网络配置" name="network">
        <div class="network-content">
          <h3>网卡配置</h3>
          <div v-for="iface in networkConfig.interfaces" :key="iface.name" class="network-interface">
            <h4>{{ iface.name }}</h4>
            <ElDescriptions :column="2" border>
              <ElDescriptionsItem label="MAC地址">{{ iface.mac }}</ElDescriptionsItem>
              <ElDescriptionsItem label="MTU">{{ iface.mtu }}</ElDescriptionsItem>
              <ElDescriptionsItem label="速率">{{ iface.speed }}</ElDescriptionsItem>
              <ElDescriptionsItem label=" duplex">{{ iface.duplex }}</ElDescriptionsItem>
              <ElDescriptionsItem label="状态">
                <ElTag :type="iface.carrier === 'on' ? 'success' : 'danger'" size="small">
                  {{ iface.carrier === 'on' ? '连接' : '未连接' }}
                </ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="驱动">{{ iface.driver }}</ElDescriptionsItem>
              <ElDescriptionsItem label="IPv4" :span="2">
                <ElTag v-for="ip in iface.ipv4" :key="ip" size="small" class="mr-4px">
                  {{ ip }}
                </ElTag>
              </ElDescriptionsItem>
            </ElDescriptions>
          </div>
        
          <h3>路由表</h3>
          <ElTable :data="networkConfig.routes" size="small">
            <ElTableColumn prop="destination" label="目标网络" width="150" />
            <ElTableColumn prop="gateway" label="网关" width="150" />
            <ElTableColumn prop="iface" label="接口" width="100" />
          </ElTable>
        </div>
      </ElTabPane>
    
      <!-- 安全巡检 Tab -->
      <ElTabPane label="安全巡检" name="security">
        <div class="security-content">
          <div class="security-section">
            <h3>安全基线检查</h3>
            <ElTable :data="securityInfo.baselineChecks" size="small">
              <ElTableColumn prop="item" label="检查项" width="200" />
              <ElTableColumn prop="result" label="结果" width="100">
                <template #default="{ row }">
                  <ElTag :type="row.passed ? 'success' : 'danger'" size="small">
                    {{ row.passed ? '通过' : '失败' }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn prop="current" label="当前配置" width="150" />
              <ElTableColumn prop="expected" label="期望配置" width="150" />
              <ElTableColumn prop="risk" label="风险等级" width="100">
                <template #default="{ row }">
                  <ElTag :type="getRiskTagType(row.risk)" size="small">
                    {{ row.risk }}
                  </ElTag>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        
          <div class="security-section">
            <h3>SSH 配置审计</h3>
            <ElDescriptions :column="2" border>
              <ElDescriptionsItem label="密码登录">
                <ElTag :type="securityInfo.sshConfig.passwordAuth ? 'danger' : 'success'" size="small">
                  {{ securityInfo.sshConfig.passwordAuth ? '允许' : '禁止' }}
                </ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="Root登录">
                <ElTag :type="securityInfo.sshConfig.rootLogin ? 'danger' : 'success'" size="small">
                  {{ securityInfo.sshConfig.rootLogin ? '允许' : '禁止' }}
                </ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="端口" :span="2">{{ securityInfo.sshConfig.port }}</ElDescriptionsItem>
            </ElDescriptions>
          </div>
        
          <div class="security-section">
            <h3>登录审计</h3>
            <ElTable :data="securityInfo.loginHistory" size="small">
              <ElTableColumn prop="user" label="用户" width="100" />
              <ElTableColumn prop="from" label="来源IP" width="150" />
              <ElTableColumn prop="time" label="登录时间" width="180" />
              <ElTableColumn prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <ElTag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                    {{ row.status === 'success' ? '成功' : '失败' }}
                  </ElTag>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        </div>
      </ElTabPane>
    
      <!-- 历史趋势 Tab -->
      <ElTabPane label="历史趋势" name="trends">
        <div class="trends-content">
          <!-- 时间范围选择 -->
          <div class="time-range-selector">
            <ElRadioGroup v-model="timeRange" @change="loadTrendData">
              <ElRadioButton label="1h">近1小时</ElRadioButton>
              <ElRadioButton label="6h">近6小时</ElRadioButton>
              <ElRadioButton label="24h">近24小时</ElRadioButton>
              <ElRadioButton label="7d">近7天</ElRadioButton>
              <ElRadioButton label="30d">近30天</ElRadioButton>
            </ElRadioGroup>
          
            <ElDatePicker
              v-model="customDateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              @change="loadCustomRangeData"
            />
          </div>
        
          <!-- 趋势图表 -->
          <div class="trend-charts">
            <div class="chart-container">
              <h3>CPU 使用率趋势</h3>
              <TrendLineChart
                :data="cpuTrendData"
                :height="300"
                unit="%"
                :thresholds="[60, 80, 95]"
              />
            </div>
          
            <div class="chart-container">
              <h3>内存使用率趋势</h3>
              <TrendLineChart
                :data="memoryTrendData"
                :height="300"
                unit="%"
                :thresholds="[75, 85, 95]"
              />
            </div>
          
            <div class="chart-container">
              <h3>磁盘使用率趋势</h3>
              <TrendLineChart
                :data="diskTrendData"
                :height="300"
                unit="%"
                :thresholds="[80, 90, 95]"
                :multiple="true"
              />
            </div>
          </div>
        
          <!-- 统计摘要 -->
          <div class="trend-summary">
            <h3>统计摘要</h3>
            <ElDescriptions :column="4" border>
              <ElDescriptionsItem label="CPU平均使用率">
                {{ trendsSummary.cpuAvg.toFixed(2) }}%
              </ElDescriptionsItem>
              <ElDescriptionsItem label="CPU峰值">
                {{ trendsSummary.cpuMax.toFixed(2) }}%
              </ElDescriptionsItem>
              <ElDescriptionsItem label="内存平均使用率">
                {{ trendsSummary.memoryAvg.toFixed(2) }}%
              </ElDescriptionsItem>
              <ElDescriptionsItem label="内存峰值">
                {{ trendsSummary.memoryMax.toFixed(2) }}%
              </ElDescriptionsItem>
            </ElDescriptions>
          </div>
        </div>
      </ElTabPane>
    </ElTabs>
  </div>
</template>
```

### 7.5 与主机资产页面的联动

在主机资产页面中增加"查看监控"快捷入口，并在操作菜单中添加监控相关操作：

```vue
<!-- 主机资产页面联动 -->
<template>
  <!-- 在现有表格的操作列中 -->
  <ElTableColumn label="操作" width="200" align="center">
    <template #default="{ row }">
      <ElButton size="small" @click="handleEdit(row)">编辑</ElButton>
      <ElButton size="small" type="primary" @click="viewMonitoring(row)">
        <icon-mdi-chart-line class="mr-4px" />
        查看监控
      </ElButton>
    </template>
  </ElTableColumn>
</template>

<script setup lang="ts">
// 跳转到监控中心的主机详情页
function viewMonitoring(server: Server) {
  router.push({
    name: 'monitoring_server_detail',
    params: { id: server.id }
  });
}
</script>
```

### 7.6 权限控制设计

```typescript
// 监控中心权限定义
interface MonitoringPermissions {
  // 监控概览权限
  'monitoring:overview:view': boolean;
  
  // 告警管理权限
  'monitoring:alerts:view': boolean;
  'monitoring:alerts:acknowledge': boolean;
  'monitoring:alerts:configure': boolean;
  
  // 趋势分析权限
  'monitoring:trends:view': boolean;
  'monitoring:trends:export': boolean;
  
  // 主机监控权限
  'monitoring:servers:view': boolean;
  'monitoring:servers:control': boolean;  // 服务启停
  
  // 巡检报告权限
  'monitoring:reports:view': boolean;
  'monitoring:reports:generate': boolean;
  'monitoring:reports:export': boolean;
  
  // 监控配置权限
  'monitoring:settings:view': boolean;
  'monitoring:settings:modify': boolean;
}

// 路由守卫中添加权限检查
router.beforeEach((to, from, next) => {
  if (to.path.startsWith('/monitoring')) {
    const permission = `monitoring:${to.name}:view`;
    if (!hasPermission(permission)) {
      ElMessage.error('您没有访问监控中心的权限');
      next(false);
    }
  }
  next();
});
```

### 7.7 图标设计

使用 `icon-mdi` 图标库，保持与现有菜单风格一致：

```typescript
// 监控中心图标映射
const monitoringIcons = {
  // 一级菜单
  'monitoring': 'icon-mdi-chart-line',
  
  // 二级菜单
  'monitoring_overview': 'icon-mdi-gauge',
  'monitoring_alerts': 'icon-mdi-bell',
  'monitoring_trends': 'icon-mdi-chart-line',
  'monitoring_servers': 'icon-mdi-server',
  'monitoring_reports': 'icon-mdi-clipboard-check',
  'monitoring_settings': 'icon-mdi-cog',
};
```

---

## 八、API 设计（更新）

    name: 'monitoring_trends',
        component: () => import('@/views/monitoring/trends/index.vue'),
        meta: {
          title: '趋势分析',
          i18nKey: 'route.monitoring_trends._value',
          icon: 'icon-mdi-chart-line',
          order: 3
        }
      },
      {
        path: 'servers',
        name: 'monitoring_servers',
        component: () => import('@/views/monitoring/servers/index.vue'),
        meta: {
          title: '主机监控',
          i18nKey: 'route.monitoring_servers._value',
          icon: 'icon-mdi-server',
          order: 4
        }
      },
      {
        path: 'reports',
        name: 'monitoring_reports',
        component: () => import('@/views/monitoring/reports/index.vue'),
        meta: {
          title: '巡检报告',
          i18nKey: 'route.monitoring_reports._value',
          icon: 'icon-mdi-clipboard-check',
          order: 5
        }
      },
      {
        path: 'settings',
        name: 'monitoring_settings',
        component: () => import('@/views/monitoring/settings/index.vue'),
        meta: {
          title: '监控配置',
          i18nKey: 'route.monitoring_settings._value',
          icon: 'icon-mdi-cog',
          order: 6
        }
      }
    ]
  }
];

```

### 7.3 主机资产页面的增强

在现有的"主机资产"页面中，增加快捷入口到监控中心：

```vue
<!-- cmdb_servers/index.vue 表格增强 -->
<template>
  <div class="servers-page">
    <!-- 现有筛选栏保持不变 -->
  
    <!-- 主机列表表格增强 -->
    <ElTable :data="servers">
      <!-- 现有列保持不变 -->
    
      <!-- 新增：迷你趋势图列 -->
      <ElTableColumn label="性能趋势" width="150">
        <template #default="{ row }">
          <div class="mini-trend">
            <MiniSparkLine 
              :data="row.cpuHistory" 
              :height="30" 
              :color="getTrendColor(row.cpuTrend)"
            />
          </div>
        </template>
      </ElTableColumn>
    
      <!-- 新增：服务状态列 -->
      <ElTableColumn label="服务" width="100">
        <template #default="{ row }">
          <ElTag 
            v-for="service in row.criticalServices" 
            :key="service.name"
            :type="service.status === 'running' ? 'success' : 'danger'"
            size="small"
          >
            {{ service.name }}
          </ElTag>
        </template>
      </ElTableColumn>
    
      <!-- 新增：告警徽章列 -->
      <ElTableColumn label="告警" width="80" align="center">
        <template #default="{ row }">
          <ElBadge 
            v-if="row.alertCount > 0" 
            :value="row.alertCount" 
            :max="99"
            type="danger"
          />
          <span v-else class="text-gray-400">-</span>
        </template>
      </ElTableColumn>
    
      <!-- 操作列增强 -->
      <ElTableColumn label="操作" width="200" align="center" fixed="right">
        <template #default="{ row }">
          <ElButton 
            size="small" 
            type="primary"
            @click="viewMonitoring(row)"
          >
            <icon-mdi-chart-line class="mr-4px" />
            查看监控
          </ElButton>
          <ElButton size="small" @click="handleEdit(row)">
            编辑
          </ElButton>
        </template>
      </ElTableColumn>
    </ElTable>
  </div>
</template>

<script setup lang="ts">
// 跳转到监控中心的主机详情页
function viewMonitoring(server: Server) {
  router.push({
    name: 'monitoring_server_detail',
    params: { id: server.id },
    query: { tab: 'overview' }  // 默认显示监控概览
  });
}
</script>
```

### 7.2 主机详情页增强

#### 监控概览 Tab

```vue
<!-- 实时性能仪表盘 -->
<div class="monitor-dashboard">
  <div class="metric-cards">
    <!-- CPU 卡片 -->
    <div class="metric-card">
      <div class="metric-header">
        <span class="metric-title">CPU使用率</span>
        <span class="metric-value">45.2%</span>
      </div>
      <div class="metric-chart">
        <MiniSparkLine :data="cpuHistory" height="60px" />
      </div>
      <div class="metric-details">
        <span>Load: 2.50 / 3.20 / 3.80</span>
      </div>
    </div>
  
    <!-- 内存卡片 -->
    <div class="metric-card">
      <div class="metric-header">
        <span class="metric-title">内存使用率</span>
        <span class="metric-value">65.8%</span>
      </div>
      <div class="metric-chart">
        <MiniSparkLine :data="memoryHistory" height="60px" />
      </div>
    </div>
  
    <!-- 磁盘卡片 -->
    <div class="metric-card">
      <div class="metric-header">
        <span class="metric-title">磁盘使用率</span>
        <span class="metric-value">72.3%</span>
      </div>
      <div class="metric-chart">
        <MiniDonutChart :data="diskPartitionData" height="60px" />
      </div>
    </div>
  </div>
  
  <!-- Top 20 进程 -->
  <div class="process-monitor">
    <h3>Top 20 进程</h3>
    <ElTable :data="topProcesses" size="small">
      <ElTableColumn prop="pid" label="PID" width="80" />
      <ElTableColumn prop="name" label="进程名" width="150" />
      <ElTableColumn prop="user" label="用户" width="100" />
      <ElTableColumn label="CPU" width="100">
        <template #default="{ row }">
          <ElProgress :percentage="row.cpuPercent" />
        </template>
      </ElTableColumn>
      <ElTableColumn label="内存" width="120">
        <template #default="{ row }">
          {{ formatBytes(row.memoryMB * 1024 * 1024) }}
        </template>
      </ElTableColumn>
    </ElTable>
  </div>
</div>
```

### 7.3 监控仪表板

```vue
<template>
  <div class="monitor-dashboard">
    <!-- 总览卡片 -->
    <div class="overview-cards">
      <div class="stat-card">
        <div class="stat-icon">
          <icon-mdi-server />
        </div>
        <div class="stat-content">
          <div class="stat-value">152</div>
          <div class="stat-label">主机总数</div>
        </div>
      </div>
    
      <div class="stat-card">
        <div class="stat-icon">
          <icon-mdi-check-circle />
        </div>
        <div class="stat-content">
          <div class="stat-value">148</div>
          <div class="stat-label">在线主机</div>
        </div>
      </div>
    
      <div class="stat-card">
        <div class="stat-icon">
          <icon-mdi-alert />
        </div>
        <div class="stat-content">
          <div class="stat-value">5</div>
          <div class="stat-label">活跃告警</div>
        </div>
      </div>
    </div>
  
    <!-- Top 10 资源使用 -->
    <div class="top-resources">
      <div class="chart-panel">
        <h3>CPU 使用率 Top 10</h3>
        <TopBarChart :data="topCpuServers" />
      </div>
    
      <div class="chart-panel">
        <h3>内存使用率 Top 10</h3>
        <TopBarChart :data="topMemoryServers" />
      </div>
    
      <div class="chart-panel">
        <h3>磁盘使用率 Top 10</h3>
        <TopBarChart :data="topDiskServers" />
      </div>
    </div>
  </div>
</template>
```

---

## 八、API 设计

### 8.1 Agent API（向后兼容）

```go
// 现有 API（保持不变）
GET /metrics          // 基础性能指标
GET /health           // 健康检查

// 新增 API
GET /extended-metrics // 扩展指标端点
POST /metrics/batch   // 批量上报
```

### 8.2 后端 API

```go
// 指标查询
GET /api/servers/:id/metrics           // 获取主机指标
GET /api/servers/:id/metrics/history   // 获取历史指标
GET /api/servers/:id/processes         // 获取进程信息
GET /api/servers/:id/services          // 获取服务状态
GET /api/servers/:id/hardware          // 获取硬件信息

// 告警管理
GET /api/alerts                         // 获取告警列表
POST /api/alerts/:id/acknowledge         // 确认告警
GET /api/alerts/statistics              // 告警统计
GET /api/alerts/rules                   // 获取告警规则
POST /api/alerts/rules                  // 创建告警规则

// 通知配置
GET /api/notifications/channels          // 获取通知渠道
POST /api/notifications/channels         // 创建通知渠道
POST /api/notifications/test             // 测试通知
```

---

## 九、实施计划

### 9.1 时间安排

```
第1步：后端开发（3-5天）
├── Agent 增强开发
│   ├── 扩展指标采集模块
│   ├── 进程 Top 20 采集
│   └── 新增 /extended-metrics API
├── 后端服务增强
│   ├── 指标存储优化（分区表）
│   ├── 告警引擎实现
│   ├── 通知渠道集成（微信+钉钉）
│   └── 数据归档任务
└── 数据库迁移

第2步：前端开发（5-7天）
├── 监控仪表板页面（新增）
├── 主机详情页增强（新增5个Tab）
├── 告警中心页面（新增）
└── 主机列表页增强

第3步：Agent 部署（1天）
├── 编译新版本 Agent
├── 批量部署到所有主机
└── 验证采集数据

第4步：测试与优化（2-3天）
├── 功能测试
├── 性能测试
├── 告警测试
└── 用户体验优化

第5步：上线与培训（1天）
├── 正式上线
├── 用户培训
└── 文档编写
```

### 9.2 风险控制

| 风险类型   | 风险描述          | 应对措施                            |
| ---------- | ----------------- | ----------------------------------- |
| 性能风险   | Agent资源消耗过高 | 限制CPU<5%、内存<50MB，支持动态降级 |
| 数据风险   | 存储空间不足      | 分区表+自动归档+定期清理            |
| 可用性风险 | 新Agent不兼容     | 保持向后兼容，增量更新              |
| 用户体验   | 功能复杂难上手    | 分级展示+操作引导+帮助文档          |

---

## 十、预期效果

### 10.1 功能对比

| 功能         | 改进前     | 改进后             |
| ------------ | ---------- | ------------------ |
| 采集指标数量 | 4个        | 100+个             |
| 告警能力     | 无         | 多级告警+实时通知  |
| 历史数据     | 无         | 30天详细+365天归档 |
| 故障定位     | 需登录主机 | 一键查看所有指标   |
| 容量规划     | 无数据支撑 | 完整趋势分析       |
| 安全巡检     | 手动执行   | 自动化检查         |

### 10.2 用户体验提升

1. **故障排查效率**：从平均30分钟 → 5分钟
2. **告警响应速度**：从被动发现 → 主动通知
3. **资源利用率**：可视化展示，容量规划有依据
4. **运维自动化**：巡检自动化，减少重复工作

---

## 十一、附录

### 11.1 技术栈

**后端**

- Go 1.21+
- Gin Web Framework
- GORM
- MySQL 8.0+
- Zap (日志)

**前端**

- Vue 3 + Element Plus
- ECharts (图表)
- WebSocket (实时推送)

**Agent**

- Go 1.21+
- gopsutil (系统信息采集)
- 标准库网络模块

### 11.2 参考文档

- [gopsutil 文档](https://github.com/shirou/gopsutil)
- [Prometheus 采集规范](https://prometheus.io/docs/instrumenting/exposition_formats/)
- [企业微信 API](https://developer.work.weixin.qq.com/document/path/90665)
- [钉钉机器人 API](https://open.dingtalk.com/document/robots/custom-bot-access)

---

**文档更新记录**

| 版本 | 日期       | 更新内容 | 作者        |
| ---- | ---------- | -------- | ----------- |
| v1.0 | 2026-05-26 | 初始版本 | OneOps 团队 |
