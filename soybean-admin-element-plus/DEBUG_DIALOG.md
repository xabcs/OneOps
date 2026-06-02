# 连接对话框调试指南

## 问题描述
主机资产中点击"连接"按钮打开终端时，连接信息框中没有显示凭证信息。

## 调试步骤

### 1. 打开浏览器控制台
- 按 `F12` 或右键 → 检查
- 切换到 "Console" 标签

### 2. 清除缓存并刷新
- 按 `Ctrl + Shift + R` (Windows/Linux) 或 `Cmd + Shift + R` (Mac)
- 这将清除缓存并硬刷新页面

### 3. 测试连接功能
1. 登录系统 (admin/123456)
2. 进入 "资产管理" → "主机资产"
3. 点击任意主机的"连接"按钮
4. 观察控制台输出

### 4. 查看调试日志

**正常情况下应该看到以下日志：**

```
[TerminalWorkbench] serverDetail: { id: 41, hostname: "pve5-factory-dev-node1", ... }
[TerminalWorkbench] credentialId from URL: 1
[TerminalWorkbench] 设置 credentialId: 1
[TerminalWorkbench] 准备显示连接对话框, showConnectDialog.value = true
[ServerConnectDialog] onMounted, visible: true, serverId: 41
[ServerConnectDialog] visible changed: true
[ServerConnectDialog] serverId: 41
[ServerConnectDialog] checkPermission 开始，serverId: 41
[ServerConnectDialog] fetchCheckConnectPermission 响应: { data: {...}, error: undefined }
[ServerConnectDialog] hasPermission: true
[ServerConnectDialog] credentials: [{ id: 1, name: "root", ... }]
[ServerConnectDialog] 选中的凭证ID: 1
```

**如果看到错误日志，请记录以下信息：**

1. `[ServerConnectDialog] fetchCheckConnectPermission 响应` 中的 error 字段
2. `[ServerConnectDialog] checkPermission 异常` 中的错误信息
3. Network 标签中的 API 请求状态

### 5. 检查网络请求

在控制台中切换到 "Network" 标签：

**查找以下请求：**
- `GET /api/cmdb/servers/41/permission`

**检查响应数据：**
```json
{
  "code": 200,
  "success": true,
  "data": {
    "credentials": [...],
    "hasPermission": true
  }
}
```

### 6. 常见问题

#### 问题1: 401 Unauthorized
**原因**: Token 过期或无效
**解决**: 重新登录系统

#### 问题2: 凭证列表为空
**原因**: 主机未绑定用户凭证
**解决**: 在主机编辑页面绑定 SSH 凭证

#### 问题3: 对话框不显示
**原因**: 可能有 JavaScript 错误
**解决**: 检查 Console 中的错误信息

## 当前配置

### 主机列表 → 连接按钮
- **URL**: `/terminal/workbench?serverId=xxx&serverName=xxx&serverIp=xxx&serverEnv=xxx&credentialId=xxx`
- **行为**: 在新标签页打开终端工作台

### 连接对话框
- **组件**: `ServerConnectDialog`
- **API**: `/api/cmdb/servers/{serverId}/permission`
- **触发时机**: 
  - `onMounted` 钩子
  - `watch(() => props.visible)` 监听器

## 联系方式

如果以上步骤都无法解决问题，请提供：
1. 浏览器控制台的完整日志
2. Network 标签中 API 请求的响应数据
3. 浏览器和操作系统的版本信息
