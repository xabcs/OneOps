# Grafana 401 未授权错误解决方案

## 问题描述

在证书监控页面中，Grafana显示登录界面，用户交互时出现以下错误：
```
PUT http://192.168.4.168:3000/api/user/password 401 (Unauthorized)
```

## 根本原因

这个问题的根本原因是Grafana的匿名访问配置不完整或未生效，导致：

1. **kiosk模式部分生效** - 面板可以显示，但交互时触发登录检查
2. **匿名访问未启用** - Grafana仍然要求用户认证
3. **浏览器会话冲突** - 浏览器缓存了旧的登录状态

## 🚀 快速解决方案

### 方案1：使用配置脚本（推荐）

在Grafana服务器上运行：

```bash
# 1. 下载配置脚本
wget https://your-server/setup-grafana-anon.sh

# 2. 添加执行权限
chmod +x setup-grafana-anon.sh

# 3. 运行配置脚本
sudo ./setup-grafana-anon.sh
```

### 方案2：手动配置

#### 步骤1：编辑Grafana配置

```bash
sudo vi /etc/grafana/grafana.ini
```

#### 步骤2：添加匿名访问配置

在配置文件中添加或修改以下内容：

```ini
# ===========================================
# Anonymous Access Configuration
# ===========================================
[auth.anonymous]
# 启用匿名访问
enabled = true
# 匿名用户组织角色（Viewer只能查看，Admin可以管理）
org_role = Viewer
# 匿名用户名（可选，用于审计日志）
org_name = Main Org.

# ===========================================
# Security Configuration
# ===========================================
[security]
# 允许在iframe中嵌入
allow_embedding = true
# 隐藏版本信息
hide_version = true
# 允许相同来源的iframe
content_security_policy = false
# 禁用严格传输安全（仅HTTP环境）
strict_transport_security = false

# ===========================================
# Session Configuration
# ===========================================
[session]
# 会话提供商（cookie用于匿名访问）
provider = cookie
# Cookie名称
cookie_name = grafana_session
# Cookie生命周期（秒）
session_life_time = 86400

# ===========================================
# Users Configuration
# ===========================================
[users]
# 默认主题（light/dark）
default_theme = light
# 隐藏"忘记密码"链接
hide_password_login_form = true
```

#### 步骤3：重启Grafana

```bash
sudo systemctl restart grafana-server
```

#### 步骤4：验证配置

```bash
# 检查服务状态
sudo systemctl status grafana-server

# 测试匿名访问（应该返回200）
curl -I http://192.168.4.168:3000/d/fdwecevaqo7wge/cloud-dns-record-info?orgId=1&kiosk=tv
```

## 🔍 故障排查

### 问题1：配置后仍然显示登录界面

**检查方法**：
```bash
# 查看当前配置
sudo cat /etc/grafana/grafana.ini | grep -A 3 '\[auth.anonymous\]'
```

**期望输出**：
```
[auth.anonymous]
enabled = true
org_role = Viewer
```

**解决方法**：
1. 确认配置文件中的enabled=true
2. 重启Grafana服务
3. 清除浏览器缓存和Cookie

### 问题2：配置更改没有生效

**可能原因**：
- Grafana配置文件路径错误
- 多个配置文件冲突
- 服务重启不完整

**解决方法**：
```bash
# 查找Grafana配置文件路径
ps aux | grep grafana

# 检查实际使用的配置文件
sudo grafana-server -config=/etc/grafana/grafana.ini --homepath=/usr/share/grafana

# 完全重启Grafana
sudo systemctl stop grafana-server
sudo systemctl start grafana-server
```

### 问题3：浏览器显示401错误

**解决方法**：
1. **清除浏览器缓存**
   - Chrome: Ctrl+Shift+Delete (Windows) / Cmd+Shift+Delete (Mac)
   - 选择"Cookie和其他网站数据"
   - 时间范围选择"所有时间"
   - 点击"清除数据"

2. **使用隐私模式测试**
   - Chrome: Ctrl+Shift+N (Windows) / Cmd+Shift+N (Mac)
   - 访问证书监控页面

3. **清除Grafana服务端会话**
   ```bash
   # 清除Grafana会话数据
   sudo rm -rf /var/lib/grafana/sessions/*
   sudo systemctl restart grafana-server
   ```

### 问题4：Docker环境配置问题

如果Grafana运行在Docker容器中：

**方法1：使用环境变量**
```yaml
version: '3'
services:
  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer
      - GF_SECURITY_ALLOW_EMBEDDING=true
      - GF_INSTALL_PLUGINS=
    ports:
      - "3000:3000"
```

**方法2：使用配置文件挂载**
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

## 🧪 测试脚本

创建测试脚本验证配置：

```bash
#!/bin/bash
echo "=== Grafana 匿名访问测试 ==="

GRAFANA_URL="http://192.168.4.168:3000"
TEST_URL="${GRAFANA_URL}/d/fdwecevaqo7wge/cloud-dns-record-info?orgId=1&kiosk=tv"

echo "测试URL: $TEST_URL"
echo ""

# 测试基本访问
echo "1. 测试基本访问..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$TEST_URL")
if [ "$HTTP_CODE" -eq 200 ]; then
    echo "   ✓ 访问成功 (HTTP $HTTP_CODE)"
else
    echo "   ✗ 访问失败 (HTTP $HTTP_CODE)"
fi

# 测试API访问
echo "2. 测试API访问..."
API_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${GRAFANA_URL}/api/user")
echo "   API状态码: $API_CODE"

if [ "$API_CODE" -eq 200 ] || [ "$API_CODE" -eq 403 ]; then
    echo "   ✓ 匿名访问正常（API可访问）"
else
    echo "   ✗ 匿名访问可能未启用"
fi

# 测试面板数据
echo "3. 测试面板数据..."
PANEL_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
  "${GRAFANA_URL}/api/datasources" \
  -H "Accept: application/json")

if [ "$PANEL_CODE" -eq 200 ] || [ "$PANEL_CODE" -eq 403 ]; then
    echo "   ✓ 面板数据可访问"
else
    echo "   ✗ 面板数据访问异常 (HTTP $PANEL_CODE)"
fi

echo ""
echo "=== 测试完成 ==="
```

## 📋 配置检查清单

在完成配置后，确认以下所有项都已完成：

- [ ] Grafana配置文件已更新
- [ ] `[auth.anonymous]` 部分的 `enabled = true`
- [ ] `org_role = Viewer`
- [ ] `[security]` 部分的 `allow_embedding = true`
- [ ] Grafana服务已重启
- [ ] 防火墙允许3000端口访问
- [ ] 浏览器缓存已清除
- [ ] 测试脚本验证通过
- [ ] OneOps应用中可以正常访问面板

## 🚨 紧急回退

如果配置后出现严重问题，可以快速回退：

```bash
# 停止Grafana
sudo systemctl stop grafana-server

# 恢复备份配置
sudo cp /etc/grafana/grafana.ini.backup /etc/grafana/grafana.ini

# 重启Grafana
sudo systemctl start grafana-server
```

## 📞 获取帮助

如果问题仍未解决：

1. **检查Grafana日志**
   ```bash
   sudo journalctl -u grafana-server -f
   ```

2. **查看详细配置**
   ```bash
   sudo cat /etc/grafana/grafana.ini
   ```

3. **联系技术支持**
   - 提供Grafana版本：`grafana-server -v`
   - 提供错误日志
   - 提供配置文件内容

## 🔗 相关文档

- [Grafana官方文档 - 认证](https://grafana.com/docs/grafana/latest/authentication/)
- [Grafana官方文档 - 嵌入](https://grafana.com/docs/grafana/latest/http_api/embedding/)
- [Grafana Kiosk模式](https://grafana.com/docs/grafana/latest/whatsnew/whats-new-in-v7-2/)
