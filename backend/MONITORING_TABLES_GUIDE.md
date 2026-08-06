# 监控表 agent_metrics 和 agent_alerts 详解

## 📋 表结构定义

### 1. agent_metrics 表（指标存储）

**在系统初始化时创建：**
```sql
CREATE TABLE IF NOT EXISTS agent_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
    metric_type VARCHAR(50) NOT NULL COMMENT '指标类型: performance/system/hardware/service/process/network/security',
    metric_data LONGTEXT NOT NULL COMMENT 'JSON格式指标数据',
    report_time DATETIME NOT NULL COMMENT '采集时间',
    received_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '接收时间',
    INDEX idx_server_type_time (server_id, metric_type, received_at),
    INDEX idx_report_time (report_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent指标数据表';
```

**存储内容：**
- **performance**: 性能指标（CPU、内存、磁盘、网络）
- **system**: 系统信息（内核版本、系统架构）
- **hardware**: 硬件信息（CPU型号、内存条、磁盘）
- **service**: 服务状态（系统服务、进程）
- **process**: 进程信息（Top进程、线程数）
- **network**: 网络配置（网卡配置、连接数）
- **security**: 安全信息（SSH配置、防火墙）

### 2. agent_alerts 表（告警记录）

**在系统初始化时创建：**
```sql
CREATE TABLE IF NOT EXISTS agent_alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
    hostname VARCHAR(100) NOT NULL COMMENT '主机名',
    ip VARCHAR(50) NOT NULL COMMENT 'IP地址',
    rule_id VARCHAR(50) NOT NULL COMMENT '规则ID',
    level VARCHAR(20) NOT NULL COMMENT '告警级别: critical/high/medium/low/info',
    message VARCHAR(500) NOT NULL COMMENT '告警内容',
    metric_value DECIMAL(10,2) COMMENT '当前值',
    threshold DECIMAL(10,2) COMMENT '阈值',
    first_seen DATETIME NOT NULL COMMENT '首次触发时间',
    last_seen DATETIME NOT NULL COMMENT '最近触发时间',
    acknowledged TINYINT(1) DEFAULT 0 COMMENT '是否已确认',
    acknowledged_by VARCHAR(100) COMMENT '确认人',
    acknowledged_at DATETIME COMMENT '确认时间',
    resolved_at DATETIME COMMENT '恢复时间',
    INDEX idx_server_level (server_id, level),
    INDEX idx_acknowledged (acknowledged, first_seen),
    INDEX idx_first_seen (first_seen)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警记录表';
```

**告警级别：**
- **critical**: 严重告警（系统不可用）
- **high**: 高级告警（严重影响）
- **medium**: 中级告警（需要关注）
- **low**: 低级告警（轻微影响）
- **info**: 信息提示

## 🔄 系统启动时的操作

### 阶段2：SQL迁移脚本 (runMigrations)

```go
// 在 services/init.go:76-118
func (s *InitService) runMigrations() error {
    // 1. 创建 agent_metrics 表
    createMetricsTable := `...`
    db.Exec(createMetricsTable)

    // 2. 创建 agent_alerts 表
    createAlertsTable := `...`
    db.Exec(createAlertsTable)

    return nil
}
```

**特点：**
- 使用 `CREATE TABLE IF NOT EXISTS` 确保幂等性
- 创建复合索引优化查询性能
- 支持时间范围查询和状态过滤
- 不会删除已有数据，安全增量更新

## 🚀 运行时工作机制

### Agent 指标采集调度器

**启动时机：** main.go:102
```go
go services.StartAgentMetricsScheduler()
```

**调度周期：**
```go
// 在 services/agent.go:351-400
func StartAgentMetricsScheduler() {
    // 心跳检测：每1分钟
    heartbeatTicker := time.NewTicker(1 * time.Minute)
    
    // 性能指标采集：每30秒
    fastMetricsTicker := time.NewTicker(30 * time.Second)
    
    // 系统信息采集：每5分钟
    slowMetricsTicker := time.NewTicker(5 * time.Minute)
}
```

### 数据流向

```
Agent (部署在目标主机)
   ↓ HTTP 心跳 (每分钟)
POST /api/v1/cmdb/agent/heartbeat
   ↓ 更新主机状态
servers.agent_status = 'running'
servers.last_heartbeat_at = NOW()

后端调度器 (主动拉取)
   ↓ 性能指标 (每30秒)
GET http://{agent_ip}:9100/extended-metrics
   ↓ 存储到 agent_metrics
INSERT INTO agent_metrics (metric_type='performance', ...)

   ↓ 系统信息 (每5分钟)
GET http://{agent_ip}:9100/extended-metrics
   ↓ 存储到 agent_metrics
INSERT INTO agent_metrics (metric_type='system', ...)

告警规则引擎
   ↓ 检查阈值
SELECT * FROM agent_alert_rules WHERE enabled = true
   ↓ 对比当前指标
if cpu_usage > 80% → INSERT INTO agent_alerts
```

## 📊 数据存储示例

### agent_metrics 记录示例

```json
{
  "id": 12345,
  "server_id": 101,
  "metric_type": "performance",
  "metric_data": {
    "cpu": {
      "usagePercent": 45.2,
      "cores": 4,
      "mhz": 2400
    },
    "memory": {
      "total": 8589934592,
      "used": 4294967296,
      "usedPercent": 50.0
    },
    "disk": {
      "total": 1099511627776,
      "used": 549755813888,
      "usedPercent": 50.0
    }
  },
  "report_time": "2026-08-06 10:30:00",
  "received_at": "2026-08-06 10:30:02"
}
```

### agent_alerts 记录示例

```json
{
  "id": 23456,
  "server_id": 101,
  "hostname": "web-server-01",
  "ip": "192.168.1.100",
  "rule_id": "cpu_high",
  "level": "high",
  "message": "CPU使用率超过80%",
  "metric_value": 85.5,
  "threshold": 80.0,
  "first_seen": "2026-08-06 10:25:00",
  "last_seen": "2026-08-06 10:30:00",
  "acknowledged": false,
  "acknowledged_by": null,
  "acknowledged_at": null,
  "resolved_at": null
}
```

## 🎯 主要功能

### 1. 性能监控
- **CPU使用率**: 实时监控CPU负载
- **内存使用**: 跟踪内存使用情况
- **磁盘空间**: 监控磁盘使用率和分区
- **网络流量**: 统计网络接口流量
- **系统负载**: 1/5/15分钟平均负载

### 2. 系统信息收集
- **操作系统**: 版本、内核、架构
- **硬件信息**: CPU型号、内存配置、磁盘型号
- **网络配置**: 网卡、IP地址、路由
- **服务状态**: 系统服务运行状态
- **进程信息**: Top进程列表

### 3. 告警管理
- **阈值告警**: 基于规则自动触发
- **告警级别**: 分级告警处理
- **告警确认**: 支持人工确认和跟踪
- **告警恢复**: 自动检测恢复时间
- **告警历史**: 保留历史告警记录

## 🔧 告警规则配置

### 规则定义
```sql
CREATE TABLE IF NOT EXISTS agent_alert_rules (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    level VARCHAR(20) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    condition VARCHAR(10) NOT NULL,
    threshold DECIMAL(10,2) NOT NULL,
    duration INT DEFAULT 300,
    description VARCHAR(500),
    enabled BOOLEAN DEFAULT TRUE
);
```

### 规则示例
```sql
-- CPU使用率告警
INSERT INTO agent_alert_rules VALUES (
    'cpu_high', 'CPU使用率过高', 'high',
    'cpu_usage', '>', 80.0, 300,
    'CPU使用率连续5分钟超过80%'
);

-- 内存使用率告警
INSERT INTO agent_alert_rules VALUES (
    'memory_high', '内存使用率过高', 'medium',
    'memory_usage', '>', 85.0, 600,
    '内存使用率连续10分钟超过85%'
);
```

## 📈 查询使用场景

### 1. 主机性能趋势
```sql
-- 查询最近24小时的CPU使用率
SELECT 
    server_id,
    JSON_EXTRACT(metric_data, '$.cpu.usagePercent') as cpu_usage,
    report_time
FROM agent_metrics
WHERE metric_type = 'performance'
  AND server_id = 101
  AND report_time > NOW() - INTERVAL 24 HOUR
ORDER BY report_time;
```

### 2. 活跃告警查询
```sql
-- 查询未确认的严重告警
SELECT 
    a.id,
    a.hostname,
    a.ip,
    a.level,
    a.message,
    a.metric_value,
    a.first_seen
FROM agent_alerts a
WHERE a.level IN ('critical', 'high')
  AND a.acknowledged = 0
  AND a.resolved_at IS NULL
ORDER BY a.level, a.first_seen;
```

### 3. 主机健康状态
```sql
-- 统计各主机告警数量
SELECT 
    server_id,
    hostname,
    SUM(CASE WHEN level = 'critical' THEN 1 ELSE 0 END) as critical_count,
    SUM(CASE WHEN level = 'high' THEN 1 ELSE 0 END) as high_count,
    SUM(CASE WHEN acknowledged = 0 THEN 1 ELSE 0 END) as unacknowledged_count
FROM agent_alerts
WHERE resolved_at IS NULL
GROUP BY server_id, hostname;
```

## 🚨 告警生命周期

```
正常状态
   ↓ 指标超过阈值 + 持续时间满足
触发告警 (INSERT INTO agent_alerts)
   ↓ 状态: unresolved, unacknowledged
   ↓ 管理员确认
确认告警 (UPDATE acknowledged = 1)
   ↓ 状态: unresolved, acknowledged
   ↓ 指标恢复正常
告警恢复 (UPDATE resolved_at = NOW())
   ↓ 状态: resolved, archived
```

## 💡 设计特点

1. **高性能存储**: 使用 JSON 存储复杂指标数据，避免表结构频繁变更
2. **灵活扩展**: 新增指标类型只需添加 metric_type，无需修改表结构
3. **时间序列**: 按时间索引，支持高效的时间范围查询
4. **告警去重**: 同一规则相同主机只保留最新告警记录
5. **状态管理**: 支持告警确认、恢复等状态转换
6. **历史追溯**: 保留历史数据用于性能分析和故障排查

## 🎛️ 管理接口

- **告警规则管理**: `/api/v1/monitoring/alerts/rules`
- **告警查询**: `/api/v1/monitoring/alerts`
- **性能趋势**: `/api/v1/monitoring/performance/trend`
- **主机状态**: `/api/v1/cmdb/servers/:id/extended-metrics`

---

**总结**: 这两个表支撑了 OneOps 的完整监控告警体系，实现了从主机状态监控、性能指标采集、告警规则触发到告警管理的全流程自动化。