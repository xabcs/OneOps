# OneOps Agent 监控增强 - API 接口定义

**版本**: v1.0
**创建日期**: 2026-05-26
**说明**: 定义 Agent、后端服务、前端之间的完整接口契约

---

## 1. Agent 扩展指标 API

### 1.1 获取扩展指标

**接口**: `GET /extended-metrics`

**说明**: Agent 采集并返回所有扩展指标数据，由后端服务定时拉取

**请求参数**: 无

**响应格式**:
```json
{
  "performance": {
    "cpu": {
      "usagePercent": 45.2,
      "user": 35.1,
      "system": 8.3,
      "idle": 56.7,
      "iowait": 1.5,
      "cores": 4,
      "mhz": 2400.0
    },
    "memory": {
      "total": 16765271040,
      "used": 8349756928,
      "free": 8415514112,
      "usedPercent": 49.8,
      "available": 8415514112,
      "buffers": 134217728,
      "cached": 2147483648
    },
    "disk": {
      "total": 1099511627776,
      "used": 548685096960,
      "free": 550826530816,
      "usedPercent": 49.9,
      "inodesTotal": 67108864,
      "inodesUsed": 33554432,
      "inodesUsedPercent": 50.0,
      "partitions": [
        {
          "device": "/dev/sda1",
          "mountpoint": "/",
          "fstype": "ext4",
          "total": 1099511627776,
          "used": 548685096960,
          "free": 550826530816,
          "usedPercent": 49.9
        },
        {
          "device": "/dev/sdb1",
          "mountpoint": "/data",
          "fstype": "xfs",
          "total": 2199023255552,
          "used": 1099511627776,
          "free": 1099511627776,
          "usedPercent": 50.0
        }
      ],
      "io": {
        "readBytes": 1073741824,
        "writeBytes": 536870912,
        "readCount": 1024,
        "writeCount": 2048,
        "readTime": 150,
        "writeTime": 200
      }
    },
    "network": {
      "interfaces": [
        {
          "name": "eth0",
          "bytesSent": 10737418240,
          "bytesRecv": 21474836480,
          "packetsSent": 1000000,
          "packetsRecv": 2000000,
          "errin": 0,
          "errout": 0,
          "dropin": 0,
          "dropout": 0,
          "speed": 1000000000
        }
      ],
      "connections": {
        "established": 150,
        "timeWait": 20,
        "closeWait": 5,
        "listen": 10
      },
      "tcp": {
        "activeOpens": 5000,
        "passiveOpens": 3000,
        "currEstab": 150,
        "inSegs": 1000000,
        "outSegs": 800000
      }
    },
    "load": {
      "load1": 1.5,
      "load5": 1.8,
      "load15": 1.6
    }
  },
  "systemInfo": {
    "hostname": "web-server-01",
    "os": {
      "family": "linux",
      "platform": "centos",
      "platformVersion": "7.9.2009",
      "kernelVersion": "3.10.0-1160.el7.x86_64",
      "kernelArch": "x86_64",
      "virtualizationSystem": "VMware",
      "virtualizationRole": "guest"
    },
    "uptime": 3456789,
    "bootTime": 1716789456,
    "timezone": "CST",
    "timezoneOffsetSec": 28800
  },
  "hardwareInfo": {
    "cpu": {
      "vendor": "GenuineIntel",
      "model": "Intel(R) Xeon(R) CPU E5-2680 v4 @ 2.40GHz",
      "cores": 4,
      "threads": 8,
      "mhz": 2400.0,
      "cacheSize": 25600,
      "family": "6",
      "modelNum": "79",
      "stepping": "1"
    },
    "memory": {
      "total": 16765271040,
      "slots": [
        {
          "slot": "DIMM0",
          "type": "DDR4",
          "size": 8589934592,
          "speed": 2666,
          "vendor": "Samsung",
          "serial": "12345678"
        },
        {
          "slot": "DIMM1",
          "type": "DDR4",
          "size": 8589934592,
          "speed": 2666,
          "vendor": "Samsung",
          "serial": "87654321"
        }
      ]
    },
    "disk": {
      "devices": [
        {
          "name": "/dev/sda",
          "vendor": "DELL",
          "model": "PERC H730",
          "serial": "1234567890",
          "size": 1099511627776,
          "type": "ssd",
          "smartStatus": "healthy",
          "temperature": 35
        }
      ]
    },
    "network": [
      {
        "name": "eth0",
        "mac": "00:50:56:xx:xx:xx",
        "vendor": "VMware",
        "driver": "e1000e",
        "speed": 1000000000,
        "duplex": "full",
        "mtu": 1500
      }
    ]
  },
  "serviceStatus": {
    "systemdServices": [
      {
        "name": "nginx.service",
        "status": "active",
        "subStatus": "running",
        "loaded": "loaded",
        "activeState": "active",
        "subState": "running",
        "description": "The nginx HTTP and reverse proxy server"
      },
      {
        "name": "mysql.service",
        "status": "active",
        "subStatus": "running",
        "loaded": "loaded",
        "activeState": "active",
        "subState": "running",
        "description": "MySQL Community Server"
      },
      {
        "name": "docker.service",
        "status": "inactive",
        "subStatus": "dead",
        "loaded": "loaded",
        "activeState": "inactive",
        "subState": "dead",
        "description": "Docker Application Container Engine"
      }
    ],
    "listenPorts": [
      {
        "port": 22,
        "protocol": "tcp",
        "address": "0.0.0.0",
        "process": "sshd",
        "pid": 1234
      },
      {
        "port": 80,
        "protocol": "tcp",
        "address": "0.0.0.0",
        "process": "nginx",
        "pid": 5678
      },
      {
        "port": 3306,
        "protocol": "tcp",
        "address": "0.0.0.0",
        "process": "mysqld",
        "pid": 9012
      },
      {
        "port": 9100,
        "protocol": "tcp",
        "address": "0.0.0.0",
        "process": "oneops-agent",
        "pid": 4567
      }
    ]
  },
  "processInfo": {
    "total": 156,
    "top": [
      {
        "pid": 1234,
        "name": "mysqld",
        "cpuPercent": 25.5,
        "memoryPercent": 15.2,
        "memoryBytes": 2550136832,
        "status": "sleeping",
        "username": "mysql",
        "numThreads": 45,
        "createTime": 1716789456
      },
      {
        "pid": 5678,
        "name": "nginx",
        "cpuPercent": 12.3,
        "memoryPercent": 3.5,
        "memoryBytes": 586201600,
        "status": "sleeping",
        "username": "nginx",
        "numThreads": 8,
        "createTime": 1716789500
      }
      // ... Top 20 进程
    ]
  },
  "networkConfig": {
    "interfaces": [
      {
        "name": "eth0",
        "hardwareAddr": "00:50:56:xx:xx:xx",
        "mtu": 1500,
        "flags": ["up", "broadcast", "multicast"],
        "addrs": [
          {
            "ip": "192.168.1.100",
            "netmask": "255.255.255.0",
            "family": "ipv4"
          },
          {
            "ip": "fe80::250:56ff:fe00:1",
            "netmask": "ffff:ffff:ffff:ffff::",
            "family": "ipv6"
          }
        ]
      },
      {
        "name": "lo",
        "hardwareAddr": "00:00:00:00:00:00",
        "mtu": 65536,
        "flags": ["up", "loopback"],
        "addrs": [
          {
            "ip": "127.0.0.1",
            "netmask": "255.0.0.0",
            "family": "ipv4"
          }
        ]
      }
    ],
    "routes": [
      {
        "destination": "0.0.0.0/0",
        "gateway": "192.168.1.1",
        "iface": "eth0",
        "flags": ["up", "gateway"]
      },
      {
        "destination": "192.168.1.0/24",
        "gateway": "",
        "iface": "eth0",
        "flags": ["up"]
      }
    ],
    "dns": [
      "114.114.114.114",
      "8.8.8.8"
    ],
    "hosts": [
      {
        "ip": "127.0.0.1",
        "names": ["localhost"]
      },
      {
        "ip": "::1",
        "names": ["localhost", "ip6-localhost", "ip6-loopback"]
      }
    ]
  },
  "securityInfo": {
    "ssh": {
      "configFile": "/etc/ssh/sshd_config",
      "port": 22,
      "permitRootLogin": "no",
      "passwordAuthentication": "no",
      "pubkeyAuthentication": "yes",
      "maxAuthTries": 3
    },
    "firewall": {
      "backend": "firewalld",
      "status": "active",
      "zones": [
        {
          "name": "public",
          "services": ["ssh", "dhcpv6-client"],
          "ports": ["80/tcp", "443/tcp"]
        }
      ]
    },
    "selinux": {
      "enabled": true,
      "mode": "enforcing"
    },
    "users": {
      "total": 10,
      "active": 5,
      "list": [
        {
          "name": "root",
          "uid": 0,
          "gid": 0,
          "home": "/root",
          "shell": "/bin/bash"
        },
        {
          "name": "admin",
          "uid": 1000,
          "gid": 1000,
          "home": "/home/admin",
          "shell": "/bin/bash"
        }
      ]
    },
    "sudoers": [
      "root ALL=(ALL:ALL) ALL",
      "admin ALL=(ALL:ALL) ALL"
    ],
    "lastLogin": [
      {
        "user": "admin",
        "tty": "pts/0",
        "from": "192.168.1.50",
        "time": "2026-05-26T10:30:00+08:00"
      }
    ]
  },
  "configChanges": [
    {
      "file": "/etc/nginx/nginx.conf",
      "checksum": "abc123def456",
      "changeTime": "2026-05-26T09:15:00+08:00",
      "changeType": "modified"
    }
  ],
  "collectedAt": "2026-05-26T12:00:00+08:00",
  "version": "2.0.0"
}
```

**响应状态码**:
- `200 OK`: 成功获取指标
- `503 Service Unavailable`: Agent 服务异常

---

## 2. 后端服务 API

### 2.1 监控概览

**接口**: `GET /api/monitoring/overview`

**说明**: 获取监控中心总览数据

**请求参数**: 无

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "summary": {
      "totalServers": 156,
      "onlineServers": 148,
      "offlineServers": 5,
      "alertServers": 23
    },
    "topCpu": [
      {
        "serverId": 1,
        "hostname": "web-server-01",
        "ip": "192.168.1.100",
        "cpuUsage": 95.2,
        "trend": [80, 82, 85, 88, 90, 92, 95]
      }
      // ... Top 10
    ],
    "topMemory": [
      {
        "serverId": 2,
        "hostname": "db-server-01",
        "ip": "192.168.1.101",
        "memoryUsage": 92.5,
        "trend": [75, 78, 82, 85, 88, 90, 92]
      }
      // ... Top 10
    ],
    "topDisk": [
      {
        "serverId": 3,
        "hostname": "storage-server-01",
        "ip": "192.168.1.102",
        "diskUsage": 88.0,
        "trend": [70, 75, 78, 82, 85, 87, 88]
      }
      // ... Top 10
    ],
    "activeAlerts": [
      {
        "id": 1001,
        "serverId": 1,
        "hostname": "web-server-01",
        "level": "critical",
        "message": "CPU使用率持续超过95%",
        "firstSeen": "2026-05-26T11:00:00+08:00",
        "lastSeen": "2026-05-26T12:00:00+08:00"
      }
      // ... 最新告警
    ],
    "refreshTime": "2026-05-26T12:00:00+08:00"
  }
}
```

### 2.2 主机扩展指标

**接口**: `GET /api/servers/:id/extended-metrics`

**说明**: 获取指定主机的实时扩展指标

**路径参数**:
- `id`: 主机ID

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "performance": { /* 性能指标 */ },
    "systemInfo": { /* 系统信息 */ },
    "hardwareInfo": { /* 硬件信息 */ },
    "serviceStatus": { /* 服务状态 */ },
    "processInfo": { /* 进程信息 */ },
    "networkConfig": { /* 网络配置 */ },
    "securityInfo": { /* 安全信息 */ },
    "collectedAt": "2026-05-26T12:00:00+08:00"
  }
}
```

### 2.3 历史指标查询

**接口**: `GET /api/servers/:id/metrics/history`

**说明**: 查询主机历史指标数据

**路径参数**:
- `id`: 主机ID

**查询参数**:
- `metricType`: 指标类型 (cpu/memory/disk/network/load)
- `startTime`: 开始时间 (RFC3339)
- `endTime`: 结束时间 (RFC3339)
- `interval`: 聚合间隔 (5m/15m/1h/6h/24h)，默认 5m

**示例请求**:
```
GET /api/servers/1/metrics/history?metricType=cpu&startTime=2026-05-26T00:00:00+08:00&endTime=2026-05-26T12:00:00+08:00&interval=5m
```

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "metricType": "cpu",
    "interval": "5m",
    "datapoints": [
      {
        "timestamp": "2026-05-26T00:00:00+08:00",
        "value": 25.5
      },
      {
        "timestamp": "2026-05-26T00:05:00+08:00",
        "value": 26.2
      }
      // ... 更多数据点
    ]
  }
}
```

### 2.4 硬件资产信息

**接口**: `GET /api/servers/:id/hardware`

**说明**: 获取主机硬件资产详细信息

**路径参数**:
- `id`: 主机ID

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "cpu": { /* CPU 详细信息 */ },
    "memory": { /* 内存详细信息 */ },
    "disk": { /* 磁盘详细信息 */ },
    "network": [ /* 网卡详细信息 */ ],
    "collectedAt": "2026-05-26T12:00:00+08:00"
  }
}
```

### 2.5 进程信息

**接口**: `GET /api/servers/:id/processes`

**说明**: 获取主机 Top 20 进程信息

**路径参数**:
- `id`: 主机ID

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 156,
    "top": [
      {
        "pid": 1234,
        "name": "mysqld",
        "cpuPercent": 25.5,
        "memoryPercent": 15.2,
        "memoryBytes": 2550136832,
        "status": "sleeping",
        "username": "mysql",
        "numThreads": 45,
        "cmdline": "mysqld --defaults-file=/etc/my.cnf"
      }
      // ... Top 20
    ],
    "collectedAt": "2026-05-26T12:00:00+08:00"
  }
}
```

### 2.6 服务状态

**接口**: `GET /api/servers/:id/services`

**说明**: 获取主机服务状态信息

**路径参数**:
- `id`: 主机ID

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "systemdServices": [
      {
        "name": "nginx.service",
        "status": "active",
        "subStatus": "running",
        "description": "The nginx HTTP and reverse proxy server"
      }
      // ... 更多服务
    ],
    "listenPorts": [
      {
        "port": 80,
        "protocol": "tcp",
        "address": "0.0.0.0",
        "process": "nginx",
        "pid": 5678
      }
      // ... 更多端口
    ],
    "collectedAt": "2026-05-26T12:00:00+08:00"
  }
}
```

### 2.7 网络配置

**接口**: `GET /api/servers/:id/network`

**说明**: 获取主机网络配置信息

**路径参数**:
- `id`: 主机ID

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "interfaces": [ /* 网卡列表 */ ],
    "routes": [ /* 路由表 */ ],
    "dns": [ /* DNS 配置 */ ],
    "collectedAt": "2026-05-26T12:00:00+08:00"
  }
}
```

### 2.8 安全信息

**接口**: `GET /api/servers/:id/security`

**说明**: 获取主机安全信息

**路径参数**:
- `id`: 主机ID

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "ssh": { /* SSH 配置 */ },
    "firewall": { /* 防火墙状态 */ },
    "selinux": { /* SELinux 状态 */ },
    "users": { /* 用户信息 */ },
    "sudoers": [ /* sudo 配置 */ ],
    "lastLogin": [ /* 登录历史 */ ],
    "collectedAt": "2026-05-26T12:00:00+08:00"
  }
}
```

### 2.9 告警列表

**接口**: `GET /api/monitoring/alerts`

**说明**: 查询告警列表

**查询参数**:
- `serverId`: 主机ID (可选)
- `level`: 告警级别 (critical/high/medium/low/info)
- `acknowledged`: 是否已确认 (true/false)
- `resolved`: 是否已解决 (true/false，不传则查全部)
- `page`: 页码，默认 1
- `pageSize`: 每页数量，默认 20

**示例请求**:
```
GET /api/monitoring/alerts?level=critical&acknowledged=false&page=1&pageSize=20
```

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 45,
    "items": [
      {
        "id": 1001,
        "serverId": 1,
        "hostname": "web-server-01",
        "ip": "192.168.1.100",
        "ruleId": "cpu_critical",
        "level": "critical",
        "message": "CPU使用率持续超过95%",
        "metricValue": 96.5,
        "threshold": 95.0,
        "firstSeen": "2026-05-26T11:00:00+08:00",
        "lastSeen": "2026-05-26T12:00:00+08:00",
        "acknowledged": false,
        "acknowledgedBy": null,
        "acknowledgedAt": null,
        "resolvedAt": null
      }
      // ... 更多告警
    ]
  }
}
```

### 2.10 确认告警

**接口**: `POST /api/monitoring/alerts/:id/acknowledge`

**说明**: 确认告警

**路径参数**:
- `id`: 告警ID

**请求体**:
```json
{
  "comment": "已知悉，正在处理"
}
```

**响应格式**:
```json
{
  "code": 0,
  "message": "success"
}
```

### 2.11 告警统计

**接口**: `GET /api/monitoring/alerts/stats`

**说明**: 获取告警统计数据

**查询参数**:
- `startTime`: 开始时间 (可选)
- `endTime`: 结束时间 (可选)

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 150,
    "byLevel": {
      "critical": 5,
      "high": 15,
      "medium": 30,
      "low": 50,
      "info": 50
    },
    "acknowledged": 80,
    "unacknowledged": 70,
    "resolved": 60,
    "active": 90,
    "trend": [
      {
        "date": "2026-05-20",
        "count": 20
      },
      {
        "date": "2026-05-21",
        "count": 25
      }
      // ... 最近7天
    ]
  }
}
```

### 2.12 告警规则

**接口**: `GET /api/monitoring/alert-rules`

**说明**: 获取告警规则列表

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "cpu_critical",
      "name": "CPU严重告警",
      "level": "critical",
      "metric": "cpu_usage",
      "condition": ">",
      "threshold": 95.0,
      "duration": 600,
      "description": "CPU使用率持续超过95%，可能影响业务性能",
      "enabled": true
    }
    // ... 更多规则
  ]
}
```

### 2.13 更新告警规则

**接口**: `PUT /api/monitoring/alert-rules/:id`

**说明**: 更新告警规则

**路径参数**:
- `id`: 规则ID

**请求体**:
```json
{
  "name": "CPU严重告警",
  "threshold": 98.0,
  "duration": 300,
  "enabled": true
}
```

**响应格式**:
```json
{
  "code": 0,
  "message": "success"
}
```

### 2.14 通知渠道

**接口**: `GET /api/monitoring/notification-channels`

**说明**: 获取通知渠道列表

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "channelType": "wechat",
      "channelName": "企业微信群",
      "config": {
        "webhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
      },
      "enabled": true
    },
    {
      "id": 2,
      "channelType": "dingtalk",
      "channelName": "钉钉群",
      "config": {
        "webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
        "secret": "SEC***"
      },
      "enabled": true
    }
  ]
}
```

### 2.15 创建通知渠道

**接口**: `POST /api/monitoring/notification-channels`

**说明**: 创建通知渠道

**请求体**:
```json
{
  "channelType": "wechat",
  "channelName": "企业微信群",
  "config": {
    "webhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
  },
  "enabled": true
}
```

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1
  }
}
```

### 2.16 测试通知

**接口**: `POST /api/monitoring/notification-channels/:id/test`

**说明**: 测试通知渠道

**路径参数**:
- `id`: 渠道ID

**请求体**:
```json
{
  "message": "这是一条测试通知"
}
```

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "response": "发送成功"
  }
}
```

### 2.17 通知历史

**接口**: `GET /api/monitoring/notification-history`

**说明**: 查询通知历史

**查询参数**:
- `channelId`: 渠道ID (可选)
- `status`: 发送状态 (success/failed/pending)
- `page`: 页码
- `pageSize`: 每页数量

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 150,
    "items": [
      {
        "id": 1001,
        "alertId": 500,
        "channelId": 1,
        "channelName": "企业微信群",
        "status": "success",
        "message": "告警：CPU使用率持续超过95%",
        "errorMsg": null,
        "sentAt": "2026-05-26T12:00:00+08:00"
      }
    ]
  }
}
```

### 2.18 巡检报告

**接口**: `GET /api/monitoring/inspection-reports`

**说明**: 查询巡检报告列表

**查询参数**:
- `reportType`: 报告类型 (host/security/capacity)
- `status`: 状态 (pending/generating/completed/failed)
- `page`: 页码
- `pageSize`: 每页数量

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 20,
    "items": [
      {
        "id": 1,
        "reportType": "host",
        "title": "主机巡检报告 - 2026-05-26",
        "serverIds": [1, 2, 3],
        "status": "completed",
        "createdBy": "admin",
        "createdAt": "2026-05-26T10:00:00+08:00",
        "completedAt": "2026-05-26T10:05:00+08:00"
      }
    ]
  }
}
```

### 2.19 创建巡检报告

**接口**: `POST /api/monitoring/inspection-reports`

**说明**: 创建巡检报告任务

**请求体**:
```json
{
  "reportType": "host",
  "title": "主机巡检报告 - 2026-05-26",
  "serverIds": [1, 2, 3]
}
```

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1
  }
}
```

### 2.20 导出巡检报告

**接口**: `GET /api/monitoring/inspection-reports/:id/export`

**说明**: 导出巡检报告 (PDF/Excel)

**路径参数**:
- `id`: 报告ID

**查询参数**:
- `format`: 导出格式 (pdf/excel)

**响应**:
- 成功时返回文件流
- Content-Type: application/pdf 或 application/vnd.ms-excel

### 2.21 配置变更记录

**接口**: `GET /api/servers/:id/config-changes`

**说明**: 查询主机配置变更记录

**路径参数**:
- `id`: 主机ID

**查询参数**:
- `startTime`: 开始时间
- `endTime`: 结束时间
- `page`: 页码
- `pageSize`: 每页数量

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 30,
    "items": [
      {
        "id": 1,
        "filePath": "/etc/nginx/nginx.conf",
        "oldChecksum": "abc123",
        "newChecksum": "def456",
        "changeType": "modified",
        "detectedAt": "2026-05-26T09:15:00+08:00"
      }
    ]
  }
}
```

---

## 3. 前端 API 调用约定

### 3.1 请求头

所有 API 请求需携带认证 Token：

```typescript
headers: {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
}
```

### 3.2 错误处理

统一错误响应格式：

```json
{
  "code": -1,
  "message": "错误描述",
  "data": null
}
```

**常见错误码**:
- `0`: 成功
- `-1`: 通用错误
- `401`: 未认证
- `403`: 无权限
- `404`: 资源不存在
- `500`: 服务器错误

### 3.3 分页参数

```typescript
{
  page: 1,       // 页码，从 1 开始
  pageSize: 20   // 每页数量
}
```

### 3.4 时间格式

所有时间字段使用 RFC3339 格式（北京时间）：

```
2026-05-26T12:00:00+08:00
```

### 3.5 前端服务函数示例

```typescript
// src/service/api/monitoring.ts

import { request } from '../request';

/**
 * 获取监控概览
 */
export function fetchMonitoringOverview() {
  return request<Api.MonitoringOverview>({
    url: '/monitoring/overview',
    method: 'get'
  });
}

/**
 * 获取主机扩展指标
 */
export function fetchServerExtendedMetrics(id: number) {
  return request<Api.ExtendedMetrics>({
    url: `/servers/${id}/extended-metrics`,
    method: 'get'
  });
}

/**
 * 查询历史指标
 */
export function fetchServerMetricsHistory(
  id: number,
  params: {
    metricType: string;
    startTime: string;
    endTime: string;
    interval?: string;
  }
) {
  return request<Api.MetricsHistory>({
    url: `/servers/${id}/metrics/history`,
    method: 'get',
    params
  });
}

/**
 * 获取告警列表
 */
export function fetchAlerts(params: {
  serverId?: number;
  level?: string;
  acknowledged?: boolean;
  resolved?: boolean;
  page: number;
  pageSize: number;
}) {
  return request<Api.AlertList>({
    url: '/monitoring/alerts',
    method: 'get',
    params
  });
}

/**
 * 确认告警
 */
export function acknowledgeAlert(id: number, comment?: string) {
  return request({
    url: `/monitoring/alerts/${id}/acknowledge`,
    method: 'post',
    data: { comment }
  });
}
```

---

## 4. 数据采集流程

### 4.1 定时采集调度

```
后端调度器 (每1分钟)
    ↓
遍历所有 agent_status=running 的主机
    ↓
并发拉取 Agent /extended-metrics 接口 (并发度20)
    ↓
解析并存储到 agent_metrics 表 (按 metric_type 分类)
    ↓
触发告警检测引擎
    ↓
发现告警 → 写入 agent_alerts 表
    ↓
查询启用的通知渠道
    ↓
发送告警通知 (企业微信/钉钉/邮件)
    ↓
记录通知历史到 notification_history 表
```

### 4.2 数据生命周期

```
新数据 → agent_metrics (热数据，30天)
    ↓ (每天归档)
agent_metrics_archive (温数据，90天)
    ↓ (每周清理)
删除超过365天的冷数据
```

---

## 5. WebSocket 推送 (可选)

### 5.1 实时告警推送

**连接**: `ws://localhost:8082/ws/alerts`

**消息格式**:
```json
{
  "type": "new_alert",
  "data": {
    "id": 1001,
    "serverId": 1,
    "hostname": "web-server-01",
    "level": "critical",
    "message": "CPU使用率持续超过95%",
    "timestamp": "2026-05-26T12:00:00+08:00"
  }
}
```

### 5.2 实时指标推送 (订阅模式)

**订阅**: `ws://localhost:8082/ws/metrics`

**消息格式**:
```json
{
  "action": "subscribe",
  "serverIds": [1, 2, 3]
}
```

**推送格式**:
```json
{
  "type": "metrics_update",
  "serverId": 1,
  "data": {
    "cpu": 85.5,
    "memory": 72.3,
    "disk": 88.0
  },
  "timestamp": "2026-05-26T12:00:00+08:00"
}
```

---

## 6. 性能要求

### 6.1 响应时间要求

| 接口类型 | P95 响应时间 | 说明 |
|---------|-------------|------|
| 实时指标查询 | < 100ms | 从缓存读取 |
| 历史数据查询 | < 500ms | 聚合查询 |
| 告警列表查询 | < 200ms | 带过滤条件 |
| 概览数据 | < 300ms | 聚合多表 |

### 6.2 并发要求

- 支持 100+ 并发 Agent 指标拉取
- 支持 50+ 并发前端 API 请求
- 告警通知发送延迟 < 5 秒

---

**文档版本**: v1.0
**最后更新**: 2026-05-26
