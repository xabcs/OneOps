# Grafana 无缝集成配置指南

本指南说明如何配置Grafana实现与OneOps的无缝集成，无需用户手动登录。

## 方案概述

我们提供两种Grafana集成方案：

### 方案1：Kiosk模式（推荐，最简单）

在Grafana中启用匿名访问的只读模式，用户无需登录即可查看面板。

### 方案2：API Token认证（更安全）

使用Grafana API Token，通过后端代理访问Grafana。

## 方案1：Kiosk模式配置

### 步骤1：修改Grafana配置文件

编辑Grafana配置文件（通常位于 `/etc/grafana/grafana.ini`）：

```ini
[auth.anonymous]
# 启用匿名访问
enabled = true
# 匿名用户角色（Viewer只能查看，不能编辑）
org_role = Viewer

# 隐藏版本信息
[security]
hide_version = true

# 允许嵌入iframe
[security]
allow_embedding = true
```

### 步骤2：重启Grafana服务

```bash
# Linux系统
sudo systemctl restart grafana-server

# Docker容器
docker restart grafana
```

### 步骤3：验证配置

访问URL：`http://192.168.4.168:3000/d/fdwecevaqo7wge/cloud-dns-record-info?orgId=1&kiosk`

- `kiosk` 参数：进入kiosk模式（全屏显示，无导航栏）
- 无需登录即可查看面板

## 方案2：API Token认证配置

### 步骤1：创建Grafana API Token

1. 登录Grafana（admin账户）
2. 导航到：Configuration → API Keys
3. 点击 "Add API key"
4. 输入名称：`oneops-integration`
5. 角色：`Viewer`（只读）或 `Admin`（完全访问）
6. 过期时间：可选择永不过期或设置过期时间
7. 点击 "Generate" 生成Token
8. **重要**：复制并保存Token（只显示一次）

### 步骤2：配置后端

编辑 `backend/controllers/monitoring.go`：

```go
// GrafanaConfig Grafana配置
type GrafanaConfig struct {
	BaseURL string
	APIKey  string
	OrgID   string
}

config := GrafanaConfig{
	BaseURL: "http://192.168.4.168:3000",
	APIKey:  "YOUR_API_TOKEN_HERE", // 粘贴你的API Token
	OrgID:   "1",
}
```

### 步骤3：重启后端服务

```bash
# 开发环境
npm run dev:backend

# 生产环境
systemctl restart oneops
```

## 前端Kiosk模式说明

前端默认使用kiosk模式URL，特点：

- **全屏显示**：隐藏Grafana顶部和侧边栏
- **只读模式**：用户无法编辑面板
- **无登录提示**：配置匿名访问后无需登录
- **自动刷新**：面板数据自动刷新

## 网络配置要求

### 1. 确保网络连通性

OneOps服务器需要能访问Grafana服务器：

```bash
# 测试网络连通性
curl http://192.168.4.168:3000
```

### 2. 配置CORS（跨域）

如果前端直接访问Grafana，需要在Grafana配置CORS：

```ini
[security]
allow_embedding = true
```

### 3. 防火墙配置

确保防火墙允许访问Grafana端口：

```bash
# UFW防火墙
sudo ufw allow 3000/tcp

# firewalld
sudo firewall-cmd --add-port=3000/tcp --permanent
sudo firewall-cmd --reload
```

## 故障排查

### 问题1：无法显示Grafana面板

**可能原因**：
- 网络不通
- Grafana服务未启动
- 防火墙阻止访问

**解决方法**：
```bash
# 检查Grafana服务状态
sudo systemctl status grafana-server

# 检查网络连通性
curl http://192.168.4.168:3000

# 检查防火墙
sudo ufw status
```

### 问题2：显示登录页面

**可能原因**：
- 未启用匿名访问
- kiosk参数未生效

**解决方法**：
1. 检查Grafana配置是否启用匿名访问
2. 确认URL包含`kiosk`参数
3. 重启Grafana服务

### 问题3：跨域错误

**可能原因**：
- Grafana未允许iframe嵌入
- CORS配置问题

**解决方法**：
```ini
[security]
allow_embedding = true
```

### 问题4：API Token无效

**可能原因**：
- Token过期
- Token权限不足

**解决方法**：
1. 重新生成API Token
2. 确认Token角色为Viewer或Admin
3. 更新后端配置

## 安全建议

1. **使用只读权限**：API Token使用Viewer角色，避免数据被修改
2. **网络隔离**：Grafana服务器应位于内网，不直接暴露到公网
3. **定期更换Token**：建议定期更换API Token（如每90天）
4. **访问日志**：启用Grafana访问日志，监控异常访问
5. **HTTPS加密**：生产环境建议使用HTTPS

## URL参数说明

Grafana支持的URL参数：

| 参数 | 说明 | 示例 |
|------|------|------|
| `kiosk` | kiosk模式（全屏无导航栏） | `&kiosk` |
| `orgId` | 组织ID | `&orgId=1` |
| `refresh` | 自动刷新间隔（秒） | `&refresh=30s` |
| `from` | 起始时间 | `&from=now-24h` |
| `to` | 结束时间 | `&to=now` |
| `theme` | 主题（light/dark） | `&theme=dark` |

完整示例：
```
http://192.168.4.168:3000/d/fdwecevaqo7wge/cloud-dns-record-info?orgId=1&kiosk&refresh=30s&from=now-24h&to=now&theme=dark
```

## 生产环境部署建议

1. **使用反向代理**：通过Nginx等反向代理访问Grafana
2. **启用HTTPS**：使用SSL证书加密传输
3. **限制IP访问**：只允许特定IP访问Grafana
4. **定期备份**：定期备份Grafana配置和面板
5. **监控告警**：配置Grafana服务监控

## 联系支持

如果遇到问题，请联系：
- 技术支持：[support@oneops.com](mailto:support@oneops.com)
- 文档地址：[https://docs.oneops.com](https://docs.oneops.com)
