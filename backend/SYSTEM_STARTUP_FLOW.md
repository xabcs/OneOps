# OneOps 系统启动流程详解

## 📋 完整启动流程概览

```
1. 加载配置文件 (config.yaml)
   ↓
2. 初始化日志系统
   ↓
3. 初始化数据库连接
   ↓
4. 初始化加密模块
   ↓
5. 初始化服务容器
   ↓
6. 设置 JWT 密钥
   ↓
7. 初始化 Redis (可选)
   ↓
8. 初始化数据库表和数据 (重点)
   ↓
9. 清理孤儿会话
   ↓
10. 初始化 SessionManager
   ↓
11. 启动 Agent 指标采集调度器
    ↓
12. 注册路由并启动 HTTP 服务器
```

## 🔍 核心数据库初始化流程

### 阶段1：数据库模式迁移 (migrateSchema)

**执行内容：**
```sql
-- 关闭外键检查
SET FOREIGN_KEY_CHECKS=0;

-- 清理历史遗留外键
ALTER TABLE cabinets DROP FOREIGN KEY cabinets_ibfk_1;
ALTER TABLE cabinets DROP FOREIGN KEY fk_server_rooms_cabinets;

-- 基础表自动迁移（按依赖顺序）
User, Role, Menu, Permission, LoginLog, OperationLog, SystemEventLog...
ServerRoom, Cabinet, Server, ServerTag, ServerGroup...
K8sCluster, ClusterRoleBinding, K8sSession...
```

**特点：**
- 使用 GORM AutoMigrate 自动创建/更新表结构
- 按依赖顺序分批迁移（基础表 → 依赖表 → K8s表）
- 处理历史遗留的外键约束问题

### 阶段2：执行 SQL 迁移脚本 (runMigrations)

**关键 DDL 操作：**
```sql
-- 1. 添加 menu_ids 字段到 roles 表
ALTER TABLE roles ADD COLUMN IF NOT EXISTS menu_ids JSON NULL;

-- 2. 添加 resource 字段到 menus 表
ALTER TABLE menus ADD COLUMN IF NOT EXISTS resource VARCHAR(30) DEFAULT '';

-- 3. 创建索引
CREATE INDEX IF NOT EXISTS idx_menus_resource ON menus(resource);

-- 4. 添加 disk_partitions 字段到 servers 表
ALTER TABLE servers ADD COLUMN IF NOT EXISTS disk_partitions JSON NULL;

-- 5. 创建监控相关表
CREATE TABLE IF NOT EXISTS agent_metrics (...);
CREATE TABLE IF NOT EXISTS agent_alerts (...);

-- 6. 修复 group_bindings 表字段
ALTER TABLE group_bindings ADD COLUMN IF NOT EXISTS application_permission_id BIGINT UNSIGNED NULL;
```

### 阶段3：基础数据初始化 (initSeedData)

**3.1 同步菜单 (syncMenus)**
- 创建/更新 60+ 个菜单项
- 结构化菜单树（一级菜单 → 二级菜单 → 三级菜单）
- 支持动态路由模式
- 清理已废弃的菜单

**3.2 同步内置角色 (syncBuiltinRoles)**
```go
builtinRoles := []struct {
    code        string
    name        string
    description string
}{
    {"admin", "超级管理员", "拥有系统所有权限"},
    {"ops", "运维工程师", "负责主机和任务管理"},
    {"auditor", "审计员", "仅拥有查看权限"},
    {"viewer", "查看者", "仅拥有查看权限"},
    {"user", "普通用户", "系统普通用户"},
    {"test", "测试角色", "用于测试"},
}
```

**3.3 初始化管理员用户**
- 检查用户表是否为空
- 如果为空，创建默认管理员：`admin / 123456`

**3.4 同步属性定义**
- 服务器预置属性（CPU、内存、磁盘等）
- 业务属性定义

### 阶段4：模块数据初始化 (initModuleData)

**4.1 初始化权限数据 (initPermissions)**
- 从 `permissions_data.go` 加载 184+ 系统权限
- 按模块组织：system、cmdb、monitor、audit、k8s
- 三级权限结构：`module.resource.action`
- 增量更新，已存在权限自动更新

**4.2 分配默认权限 (assignDefaultPermissions)**
```sql
-- 为不同角色分配对应权限
INSERT INTO role_permissions (role_id, permission_id) VALUES (...);
```

**4.3 初始化诊断数据 (initDiagnosticData)**
- 诊断权限配置
- K8s 诊断参数配置

**4.4 初始化 Agent 版本**
- 创建默认 Agent 版本记录

**4.5 初始化 API 权限 (initAPIPermissions)**
- Level 4 API 级权限初始化
- 支持细粒度的 API 访问控制

## 🚀 Casbin 权限系统初始化

### Casbin 策略同步机制

**自动触发时机：**
1. 系统启动时调用 `InitializeCasbinPolicies()`
2. 角色权限分配时调用 `syncRoleToCasbin()`
3. 批量权限更新时事务中自动同步

**同步流程：**
```go
// 1. 清除旧的 Casbin 策略
enforcer.ClearPolicy();

// 2. 从 role_permissions 表读取所有角色权限
SELECT * FROM role_permissions 
JOIN permissions ON role_permissions.permission_id = permissions.id

// 3. 为每个角色生成 Casbin 策略
for _, role := range roles {
    permissions := GetRolePermissions(role.ID)
    for _, perm := range permissions {
        enforcer.AddPolicy(role.Code, perm.Code, "*")
    }
}

// 4. 保存策略到 casbin_rule 表
enforcer.SavePolicy();
```

**Casbin 策略格式：**
```
p, role_code, permission_code, action
例如：p, aaa, system.user.list, *
```

## 📊 数据库表结构概览

### 核心权限表
```sql
-- 用户表
users (id, username, password, role_ids JSON, ...)

-- 角色表  
roles (id, code, name, menu_ids JSON, ...)

-- 权限表
permissions (id, code, name, module, resource, action, level, ...)

-- 角色权限关联表
role_permissions (role_id, permission_id)

-- 用户角色关联（通过 users.role_ids JSON 字段）
-- 菜单表
menus (id, name, path, permission, parent_id, ...)

-- Casbin 策略表
casbin_rule (p_type, v0, v1, v2, ...)
-- v0: role_code, v1: permission_code, v2: action
```

## 🔧 为什么"重启后端"能解决权限问题？

### 问题根源
1. **权限服务单例模式**：`permissionServiceOnce sync.Once`
   - 只初始化一次，启动后不重新加载策略

2. **Casbin 策略加载时机**：仅在权限服务初始化时调用 `LoadPolicy()`
   - 从 `casbin_rule` 表读取策略
   - 如果策略在运行时通过 SQL 修改，权限服务不知道

### 解决方案
**重启后端：**
```
启动 → 权限服务初始化 → 重新加载 Casbin 策略 → 获取最新的 role_permissions 数据
```

**更好的方式（已实现）：**
```go
// 通过 API 分配权限会自动同步
func BatchAssignPermissions(roleID uint, permissionIDs []uint) error {
    // 更新数据库
    tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{})
    // ...添加新权限
    
    // ✅ 自动同步到 Casbin
    syncRoleToCasbin(&role)
}
```

## 🎯 启动性能优化要点

### 已实现的优化
1. **分阶段初始化**：按依赖顺序分4个阶段
2. **增量数据同步**：检测已存在记录，只更新必要字段
3. **事务保护**：权限分配使用事务确保一致性
4. **并发启动**：Agent 调度器异步启动

### 启动时数据库操作统计
- **表结构迁移**：50+ 表
- **菜单同步**：60+ 菜单项
- **角色同步**：6 个内置角色
- **权限初始化**：184+ 权限定义
- **Casbin 策略**：127+ 条规则
- **总操作数**：500+ 次 SQL 操作

## 📌 关键设计原则

1. **幂等性**：多次启动不会重复创建数据
2. **增量更新**：检测已存在记录并更新，而非报错
3. **容错性**：单个初始化失败不影响其他模块
4. **自动同步**：权限变更自动同步到 Casbin
5. **分层初始化**：按依赖顺序分阶段执行

## 🚨 常见问题排查

**问题：权限检查失败但数据库有数据**
- 检查 `casbin_rule` 表是否包含最新策略
- 检查 `role_permissions` 表关联是否正确
- 重启后端让权限服务重新加载策略

**问题：菜单显示不正确**
- 检查 `menus` 表数据完整性
- 检查 `roles.menu_ids` 字段是否正确
- 清理前端缓存重新登录

**问题：初始化缓慢**
- 检查数据库连接性能
- 检查外键约束是否过多
- 考虑禁用不必要的模块初始化

---

**最后更新：2026-08-06**
**维护者：OneOps 开发团队**
