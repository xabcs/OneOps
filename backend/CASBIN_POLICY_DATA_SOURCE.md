# Casbin 策略数据来源详解

## 🎯 简要回答

**Casbin 策略的 127+ 条规则数据主要来自：**

```
/Users/mpm/Desktop/doc/go/OneOpsV2/OneOps/backend/services/permissions_data.go
```

## 📂 相关文件清单

### 1. 权限定义文件（核心数据源）

**`services/permissions_data.go`** - 172行
- 定义了 **184+ 个系统权限**
- 涵盖 5 个主要模块：system、cmdb、monitor、k8s、audit
- 三级权限结构：`module.resource.action`

**权限统计：**
- **系统管理模块**: 28 个权限
- **CMDB模块**: 36 个权限
- **监控模块**: 15 个权限
- **K8s模块**: 17 个权限
- **审计模块**: 12 个权限
- **总计**: **108 个权限**（不含模块级和资源级权限）

### 2. Casbin 模型配置文件

**`config/casbin_model.conf`** - 14行
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

**作用**: 定义 Casbin 策略的数据格式和匹配规则

### 3. Casbin 策略同步服务

**`services/permission_service.go`** - 685行
- `syncRoleToCasbin()` - 同步单个角色到 Casbin
- `SyncAllRolesToCasbin()` - 同步所有角色到 Casbin
- `InitializeCasbinPolicies()` - 初始化所有策略

## 🔄 数据流向图

```
1. 系统启动
   ↓
2. 执行 initPermissions() → 调用 GetAllSystemPermissions()
   ↓
3. 从 permissions_data.go 读取权限定义
   ↓
4. 写入 permissions 表（数据库）
   ↓
5. 执行 assignDefaultPermissions() 
   ↓
6. 为内置角色分配权限 → 写入 role_permissions 表（数据库）
   ↓
7. 执行 SyncAllRolesToCasbin()
   ↓
8. 读取 role_permissions + permissions 表
   ↓
9. 生成 Casbin 策略 → 写入 casbin_rule 表（数据库）
   ↓
10. 权限服务启动 → LoadPolicy() 从 casbin_rule 表加载策略
```

## 📊 权限代码结构

### 权限分层设计

**Level 1: 模块级权限**
```
system      - 系统管理
cmdb        - 资产管理  
monitor     - 监控管理
k8s         - K8s管理
audit       - 审计管理
```

**Level 2: 资源级权限**
```
system.user       - 用户管理
system.role       - 角色管理
system.menu       - 菜单管理
cmdb.server       - 服务器管理
cmdb.group        - 分组管理
monitor.data      - 监控数据
monitor.alert     - 告警管理
```

**Level 3: 操作级权限**
```
system.user.list        - 用户列表
system.user.create      - 创建用户
system.user.update      - 更新用户
system.user.delete      - 删除用户
system.user.reset_password - 重置密码
```

## 🗄️ 数据库表关系

```
permissions (权限定义表)
    ↓ GetAllSystemPermissions()
    ↓ 插入/更新权限定义
    ↓
role_permissions (角色权限关联表)
    ↓ assignDefaultPermissions()
    ↓ 为角色分配权限
    ↓
casbin_rule (Casbin策略表)
    ↓ SyncAllRolesToCasbin()
    ↓ 生成策略规则
    ↓
    格式: p, role_code, permission_code, *
    示例: p, aaa, system.user.list, *
```

## 🚀 运行时权限检查流程

```
1. 用户发起请求 → JWT Token
   ↓
2. Auth 中间件解析 Token → 提取 user_id
   ↓
3. RequirePermission 中间件 → 检查权限代码
   ↓
4. HasPermission(userID, permissionCode)
   ↓
5. 查询 users 表 → 获取 role_ids JSON数组
   ↓
6. 查询 roles 表 → 获取角色信息
   ↓
7. Casbin.Enforce(role.Code, permissionCode, "*")
   ↓
8. 从 casbin_rule 表匹配策略
   ↓
9. 返回 true/false
```

## 📋 权限定义示例

### 系统管理模块 (28个权限)
```go
// 用户管理 (7个)
{Code: "system.user.list", Name: "用户列表", Module: "system", Resource: "user", Action: "list"}
{Code: "system.user.view", Name: "查看用户", Module: "system", Resource: "user", Action: "view"}
{Code: "system.user.create", Name: "创建用户", Module: "system", Resource: "user", Action: "create"}
{Code: "system.user.update", Name: "更新用户", Module: "system", Resource: "user", Action: "update"}
{Code: "system.user.delete", Name: "删除用户", Module: "system", Resource: "user", Action: "delete"}
{Code: "system.user.reset_password", Name: "重置密码", Module: "system", Resource: "user", Action: "reset_password"}

// 角色管理 (6个)
{Code: "system.role.list", Name: "角色列表", Module: "system", Resource: "role", Action: "list"}
{Code: "system.role.assign_permissions", Name: "分配权限", Module: "system", Resource: "role", Action: "assign_permissions"}
...
```

### CMDB模块 (36个权限)
```go
// 服务器管理 (7个)
{Code: "cmdb.server.list", Name: "服务器列表", Module: "cmdb", Resource: "server", Action: "list"}
{Code: "cmdb.server.connect", Name: "连接服务器", Module: "cmdb", Resource: "server", Action: "connect"}

// 分组管理 (6个)
{Code: "cmdb.group.assign", Name: "分配服务器", Module: "cmdb", Resource: "group", Action: "assign"}
...
```

## 🔧 如何修改权限定义

### 添加新权限
1. **编辑** `services/permissions_data.go`
2. **添加权限定义**：
   ```go
   {Code: "new.module.action", Name: "新权限", Module: "new", Resource: "module", Action: "action", Level: 3, Status: 1}
   ```
3. **重启后端** → 自动同步到数据库和 Casbin

### 修改现有权限
1. **编辑** `services/permissions_data.go` 中的权限定义
2. **重启后端** → 权限会自动更新

### 删除权限
1. **从代码中注释**不需要的权限定义
2. **重启后端** → 权限会更新

## 📌 关键特点

1. **单一数据源**: 所有权限定义都在 `permissions_data.go` 中
2. **自动同步**: 权限变更自动同步到 Casbin 和数据库
3. **版本控制**: 权限定义纳入 Git 版本控制
4. **类型安全**: 使用 Go 结构体，编译时检查
5. **可维护性**: 集中管理，易于查找和修改

---

**总结**: Casbin 策略的 127+ 条规则本质上来自 `services/permissions_data.go` 文件中定义的权限数据，通过系统启动时的初始化流程同步到数据库和 Casbin 系统中。