# 权限代码不匹配问题分析报告

## 🚨 严重问题：权限代码完全不匹配！

### 问题概述

**`assignDefaultPermissions()` 中使用的权限代码与 `permissions.json` 中定义的权限代码完全不匹配！**

## 📊 对比分析

### permissions.json 中定义的权限代码（实际使用）

**模块命名规范：** `module.resource.action`

```
✅ 正确的权限代码：
- system.user.list
- system.user.create
- system.user.update
- system.user.delete
- system.role.list
- system.role.assign_permissions
- cmdb.server.list
- cmdb.server.view
- cmdb.server.connect
- monitor.data.view
- monitor.alert.list
- monitor.task.create
- k8s.cluster.list
- k8s.resource.view
- audit.login_log.list
- audit.operation_log.view
```

**实际定义的模块：**
- `system` - 系统管理
- `cmdb` - CMDB管理
- `monitor` - 监控管理
- `k8s` - K8s管理
- `audit` - 审计管理

### assignDefaultPermissions() 中使用的权限代码（错误）

**错误示例：**

```go
// ❌ 错误：使用了 .query 后缀
"cmdb.server.query"           // 应该是: cmdb.server.list
"cmdb.business.query"         // 应该是: cmdb.business.list
"monitoring.overview.query"   // 应该是: monitor.data.view

// ❌ 错误：使用了 .view 后缀（不一致）
"system.user.view"            // 应该是: system.user.list
"system.role.view"            // 应该是: system.role.list

// ❌ 错误：模块名称错误
"monitoring.*"                // 应该是: monitor.*
"auth.user.query"             // ❌ 根本不存在 auth 模块！

// ❌ 错误：使用了不存在的权限
"cmdb.credentials.access"     // permissions.json 中不存在
"cmdb.policies.query"         // permissions.json 中不存在
"cmdb.dashboard.query"        // permissions.json 中不存在
"cmdb.audit.changes"          // permissions.json 中不存在
```

## 🔍 详细不匹配列表

### ops 角色的权限分配（42个权限）

| assignDefaultPermissions 中的代码 | permissions.json 中的正确代码 | 状态 |
|----------------------------------|------------------------------|------|
| system.user.view | system.user.list | ❌ 不匹配 |
| system.user.create | system.user.create | ✅ 匹配 |
| system.user.update | system.user.update | ✅ 匹配 |
| system.user.delete | system.user.delete | ✅ 匹配 |
| system.role.view | system.role.list | ❌ 不匹配 |
| system.role.update | system.role.update | ✅ 匹配 |
| system.menu.view | system.menu.list | ❌ 不匹配 |
| cmdb.server.query | cmdb.server.list | ❌ 不匹配 |
| cmdb.server.create | cmdb.server.create | ✅ 匹配 |
| cmdb.server.update | cmdb.server.update | ✅ 匹配 |
| cmdb.server.delete | cmdb.server.delete | ✅ 匹配 |
| cmdb.business.query | cmdb.business.list | ❌ 不匹配 |
| cmdb.rooms.query | cmdb.rooms.list | ❌ 不匹配 |
| cmdb.tags.query | cmdb.tags.list | ❌ 不匹配 |
| cmdb.credentials.access | ❌ 不存在 | ❌ 错误 |
| cmdb.credentials.ssh | ❌ 不存在 | ❌ 错误 |
| cmdb.policies.query | ❌ 不存在 | ❌ 错误 |
| cmdb.config.business | ❌ 不存在 | ❌ 错误 |
| cmdb.agents.query | cmdb.agents.list | ❌ 不匹配 |
| cmdb.dashboard.query | ❌ 不存在 | ❌ 错误 |
| cmdb.audit.changes | ❌ 不存在的模块 | ❌ 错误 |
| cmdb.audit.command.history | ❌ 不存在的模块 | ❌ 错误 |
| cmdb.audit.online | ❌ 不存在的模块 | ❌ 错误 |
| cmdb.audit.sessions | ❌ 不存在的模块 | ❌ 错误 |
| cmdb.audit.commands | ❌ 不存在的模块 | ❌ 错误 |
| monitoring.overview.query | monitor.data.view | ❌ 模块名错误 |
| monitoring.servers.query | ❌ 不存在 | ❌ 错误 |
| monitoring.alerts.query | monitor.alert.list | ❌ 模块名错误 |
| monitoring.trends.query | ❌ 不存在 | ❌ 错误 |
| monitoring.reports.query | monitor.task.list | ❌ 模块名错误 |
| monitoring.settings.query | ❌ 不存在 | ❌ 错误 |
| k8s.cluster.query | k8s.cluster.list | ❌ 不匹配 |
| k8s.workload.query | ❌ 不存在 | ❌ 错误 |
| k8s.diagnostic.execute | ❌ 不存在 | ❌ 错误 |
| k8s.network.query | ❌ 不存在 | ❌ 错误 |
| k8s.config.query | ❌ 不存在 | ❌ 错误 |
| audit.login.view | audit.login_log.list | ❌ 不匹配 |
| audit.operation.view | audit.operation_log.list | ❌ 不匹配 |
| audit.system.view | audit.system_event.list | ❌ 不匹配 |
| auth.user.query | ❌ 不存在的模块 | ❌ 错误 |
| auth.role.query | ❌ 不存在的模块 | ❌ 错误 |
| auth.app.query | ❌ 不存在的模块 | ❌ 错误 |
| ... (还有更多 auth.* 权限) | ❌ 全部不存在 | ❌ 错误 |

**统计：ops 角色 42 个权限中，约有 30+ 个权限代码错误或不存在！**

## 🎯 根本原因

### 问题1：命名规范不一致

**permissions.json 使用：**
- ✅ `.list` - 列表查询
- ✅ `.view` - 详情查看
- ✅ `.create` - 创建
- ✅ `.update` - 更新
- ✅ `.delete` - 删除

**assignDefaultPermissions 使用：**
- ❌ `.query` - 查询（不存在于 permissions.json）
- ❌ `.view` - 查看（只用于部分权限）

### 问题2：模块名称错误

**permissions.json 定义的模块：**
- ✅ `monitor` - 监控管理

**assignDefaultPermissions 使用：**
- ❌ `monitoring` - 错误的模块名
- ❌ `auth` - 不存在的模块

### 问题3：缺失的权限定义

**assignDefaultPermissions 引用了 permissions.json 中不存在的权限：**
```
❌ cmdb.credentials.*     - 凭证管理权限
❌ cmdb.policies.*        - 策略管理权限
❌ cmdb.dashboard.*       - 仪表盘权限
❌ cmdb.audit.*           - CMDB审计模块（应该用 audit.*）
❌ auth.*                 - 授权中心模块（完全不存在）
```

## 💥 影响

### 初始化时的后果

```go
// init.go:1103-1109
var permissions []models.Permission
db.Where("code IN ?", permissionCodes).Find(&permissions)
// 由于权限代码不存在，permissions 数组为空或部分为空
// 导致角色没有被分配实际的权限！
```

### 运行时的后果

1. ❌ 权限分配失败 - 权限代码在数据库中不存在
2. ❌ 角色没有实际权限 - 只有少量匹配的权限被分配
3. ❌ 用户操作被拒绝 - "权限不足" 错误
4. ❌ Casbin 规则缺失 - 无法通过权限检查

## ✅ 解决方案

### 方案1：修正 assignDefaultPermissions()（推荐）

**修改 `services/init.go:assignDefaultPermissions()` 使用正确的权限代码：**

```go
rolePermissions := map[string][]string{
    "admin": {"*.*.*"},
    "ops": {
        // 系统管理
        "system.user.list", "system.user.create", "system.user.update", "system.user.delete",
        "system.role.list", "system.role.update",
        "system.menu.list",
        // CMDB
        "cmdb.server.list", "cmdb.server.create", "cmdb.server.update", "cmdb.server.delete",
        "cmdb.business.list",
        "cmdb.rooms.list",
        "cmdb.tags.list",
        "cmdb.agents.list",
        // 监控
        "monitor.data.view",
        "monitor.alert.list",
        "monitor.task.list",
        // K8s
        "k8s.cluster.list",
        "k8s.resource.view",
        // 审计
        "audit.login_log.list",
        "audit.operation_log.list",
    },
    // ...
}
```

### 方案2：补充 permissions.json（不推荐）

在 permissions.json 中添加缺失的权限定义（如 `cmdb.credentials.*`, `auth.*` 等），但这会导致权限定义膨胀。

## 🚨 当前系统状态

**permissions.json：** 83 个操作级权限
**assignDefaultPermissions：** 引用了大量不存在的权限

**结果：** 权限分配大部分失败，角色实际上没有获得应有的权限！

---

**结论：**
- ✅ permissions.json 内容是正确的、完整的（83个权限）
- ❌ assignDefaultPermissions() 中的权限代码是错误的、过时的
- 🔧 需要修正 assignDefaultPermissions() 使其与 permissions.json 一致