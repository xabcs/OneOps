# Grafana 集成配置指南

## 快速配置步骤

### 方法1：自动配置（推荐）

如果您的Grafana安装在Linux服务器上，可以使用自动配置脚本：

```bash
# 1. 下载配置脚本到Grafana服务器
scp setup-grafana-anon.sh user@grafana-server:/tmp/

# 2. 在Grafana服务器上执行
ssh user@grafana-server
cd /tmp
chmod +x setup-grafana-anon.sh
sudo ./setup-grafana-anon.sh
```

### 方法2：手动配置

如果自动脚本不适用，可以手动配置：

#### 步骤1：编辑Grafana配置

```bash
sudo vi /etc/grafana/grafana.ini
```

#### 步骤2：添加以下配置

```ini
# ===========================================
# Anonymous Access Configuration
# ===========================================
[auth.anonymous]
# 启用匿名访问
enabled = true
# 匿名用户组织角色（Viewer只能查看）
org_role = Viewer

# ===========================================
# Security Configuration
# ===========================================
[security]
# 允许在iframe中嵌入
allow_embedding = true
# 隐藏版本信息
hide_version = true
```

#### 步骤3：重启Grafana

```bash
sudo systemctl restart grafana-server
```

## 验证配置

### 1. 检查配置是否生效

```bash
# 查看当前配置
sudo cat /etc/grafana/grafana.ini | grep -A 3 '\[auth.anonymous\]'
```

期望输出：
```
[auth.anonymous]
enabled = true
org_role = Viewer
```

### 2. 测试匿名访问

```bash
# 测试URL访问（应该返回200）
curl -I http://192.168.4.168:3000/d/fdwecevaqo7wge/cloud-dns-record-info?orgId=1&kiosk
```

### 3. 在OneOps中验证

1. 打开OneOps应用
2. 导航到"监控中心"
3. 选择"证书监控"
4. 确认Grafana面板正常显示，无需登录

## Docker环境配置

如果Grafana运行在Docker容器中：

### 使用环境变量

```yaml
version: '3'
services:
  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer
      - GF_SECURITY_ALLOW_EMBEDDING=true
      - GF_SECURITY_HIDE_VERSION=true
    ports:
      - "3000:3000"
```

### 使用配置文件

```yaml
version: '3'
services:
  grafana:
    image: grafana/grafana:latest
    volumes:
      - ./grafana.ini:/etc/grafana/grafana.ini
      - grafana-storage:/var/lib/grafana
    ports:
      - "3000:3000"

volumes:
  grafana-storage:
```

## 故障排查

### 问题1：仍然显示登录界面

**解决方法：**

1. 清除浏览器缓存和Cookie
2. 使用隐私模式测试
3. 检查Grafana日志：

```bash
sudo journalctl -u grafana-server -f
```

### 问题2：跨域错误

**解决方法：**

确保Grafana配置中包含：
```ini
[security]
allow_embedding = true
```

### 问题3：配置更改未生效

**解决方法：**

```bash
# 完全重启Grafana
sudo systemctl stop grafana-server
sudo systemctl start grafana-server

# 检查服务状态
sudo systemctl status grafana-server
```

## 网络配置

### 防火墙配置

确保防火墙允许访问Grafana端口：

```bash
# UFW防火墙
sudo ufw allow 3000/tcp

# firewalld
sudo firewall-cmd --add-port=3000/tcp --permanent
sudo firewall-cmd --reload
```

### 网络连通性测试

```bash
# 从OneOps服务器测试Grafana连通性
curl http://192.168.4.168:3000
```

## 安全建议

1. **网络隔离**：Grafana服务器应位于内网
2. **只读权限**：匿名用户使用Viewer角色
3. **HTTPS加密**：生产环境使用HTTPS
4. **访问日志**：启用访问日志监控
5. **定期备份**：定期备份Grafana配置

## URL参数说明

可以在Grafana URL中添加以下参数：

| 参数 | 说明 | 示例 |
|------|------|------|
| `kiosk` | 全屏模式（无导航栏） | `&kiosk` |
| `orgId` | 组织ID | `&orgId=1` |
| `refresh` | 自动刷新间隔（秒） | `&refresh=30s` |
| `from` | 起始时间 | `&from=now-24h` |
| `to` | 结束时间 | `&to=now` |
| `theme` | 主题 | `&theme=dark` |

完整示例：
```
http://192.168.4.168:3000/d/fdwecevaqo7wge/cloud-dns-record-info?orgId=1&kiosk&refresh=30s&from=now-24h&to=now&theme=dark
```

## 联系支持

如果遇到问题：
1. 检查Grafana日志：`sudo journalctl -u grafana-server -f`
2. 查看Grafana文档：https://grafana.com/docs/grafana/latest/
3. 联系技术支持

## 配置检查清单

完成配置后，确认以下所有项：

- [ ] Grafana配置文件已更新
- [ ] `[auth.anonymous]` 部分的 `enabled = true`
- [ ] `org_role = Viewer`
- [ ] `[security]` 部分的 `allow_embedding = true`
- [ ] Grafana服务已重启
- [ ] 防火墙允许3000端口访问
- [ ] 浏览器缓存已清除
- [ ] 测试URL返回200状态码
- [ ] OneOps应用中可以正常访问面板
