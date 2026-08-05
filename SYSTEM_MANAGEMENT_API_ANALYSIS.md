# 系统管理模块API接口清单与问题分析

## 📋 当前API接口清单

### 1. 用户管理API (9个接口)

| HTTP方法 | 路径 | 权限码 | 权限检查 | 前端函数 | 命名问题 |
|---------|------|--------|---------|---------|----------|---------|
| GET | `/api/system/users` | system.user.view | ✅ 有 | fetchGetUserList | ✅ 正确 |
| POST | `/api/system/users` | system.user.create | ✅ 有 | fetchCreateUser | ✅ 正确 |
| PUT | `/api/system/users/:id` | system.user.update | ✅ 有 | fetchUpdateUser | ✅ 正确 |
| DELETE | `/api/system/users/:id` | system.user.delete | ✅ 有 | fetchDeleteUser | ✅ 正确 |
| PUT | `/api/system/users/:id/password` | system.user.reset_password | ✅ 有 | fetchResetUserPassword | ✅ 正确 |
| GET | `/api/system/users/:id` | system.user.view | ✅ 有 | fetchGetUserById | ✅ 正确 |
| GET | `/api/system/users/:id/roles` | 未定义权限码 | ❌ 无 | 未实现 | ❌ 缺失 |
| GET | `/api/system/users/:id/permissions` | 未定义权限码 | ❌ 无 | 未实现 | ❌ 缺失 |
| GET | `/api/system/users/:id/groups` | 未定义权限码 | ❌ 无 | 未实现 | ❌ 缺失 |

### 2. 角色管理API (8个接口)

| HTTP方法 | 路径 | 权限码 | 权限检查 | 前端函数 | 命名问题 |
|---------|------|--------|---------|---------|----------|---------|
| GET | `/api/system/roles` | ❌ 无权限检查 | ❌ 无 | fetchGetAllRoles | ⚠️ 不安全 |
| GET | `/api/system/roles/:id` | ❌ 无权限检查 | ❌ 无 | fetchGetRoleById | ⚠️ 不安全 |
| POST | `/api/system/roles` | system.role.create | ✅ 有 | fetchCreateRole | ✅ 正确 |
| PUT | `/api/system/roles/:id` | system.role.update | ✅ 有 | fetchUpdateRole | ✅ 正确 |
| DELETE | `/api/system/roles/:id` | system.role.delete | ✅ 有 | fetchDeleteRole | ✅ 正确 |
| GET | `/api/system/roles/:id/permissions` | 未定义权限码 | ❌ 无 | fetchGetRolePermissions | ⚠️ 查询接口 |
| POST | `/api/system/roles/:id/permissions` | 未定义权限码 | ❌ 无 | fetchAssignRolePermissions | ⚠️ 操作接口 |

### 3. 菜单管理API (6个接口)

| HTTP方法 | 路径 | 权限码 | 权限检查 | 前端函数 | 命名问题 |
|---------|------|--------|---------|---------|----------|---------|
| GET | `/api/system/menus` | ❌ 无权限检查 | ❌ 无 | fetchGetMenuList | ⚠️ 不安全 |
| GET | `/api/system/menus/tree` | ❌ 无权限检查 | ❌ 无 | fetchGetMenuTree | ⚠️ 不安全 |
| POST | `/api/system/menus` | system.menu.create | ✅ 有 | fetchCreateMenu | ✅ 正确 |
| PUT | `/api/system/menus/:id` | system.menu.update | ✅ 有 | fetchUpdateMenu | ✅ 正确 |
| DELETE | `/api/system/menus/:id` | system.menu.delete | ✅ 有 | fetchDeleteMenu | ✅ 正确 |
| GET | `/api/system/menus/:id` | 未定义权限码 | ❌ 无 | 未实现 | ❌ 缺失 |

### 4. 权限管理API (7个接口)

| HTTP方法 | 路径 | 权限码 | 权限检查 | 前端函数 | 命名问题 |
|---------|------|--------|---------|---------|----------|---------|
| GET | `/api/system/permissions` | ❌ 无权限检查 | ❌ 无 | fetchGetPermissionList | ⚠️ 不安全 |
| GET | `/api/system/permissions/:id` | ❌ 无权限检查 | ❌ 无 | fetchGetPermissionById | ⚠️ 不安全 |
| POST | `/api/system/permissions` | system.permission.create | ✅ 有 | fetchAddPermission | ✅ 正确 |
| PUT | `/api/system/permissions/:id` | system.permission.update | ✅ 有 | fetchUpdatePermission | ✅ 正确 |
| DELETE | `/api/system/permissions/:id` | system.permission.delete | ✅ 有 | fetchDeletePermission | ✅ 正确 |
| GET | `/api/system/roles/:id/permissions` | 查询接口 | ❌ 无 | fetchGetRolePermissions | ✅ 合理 |
| POST | `/api/system/roles/:id/permissions` | 操作接口 | ❌ 无 | fetchAssignRolePermissions | ✅ 合理 |

---

## 🔍 发现的问题

### 问题1: 权限检查不一致

**❌ GET接口缺少权限保护**：
```go
// 问题代码
system.GET("/roles", roleController.GetRoles)  // ❌ 没有权限检查
system.GET("/menus", menuController.GetMenus)  // ❌ 没有权限检查
system.GET("/menus/tree", menuController.GetMenuTree)  // ❌ 没有权限检查
```

**影响**：
- 任何人都可以查看角色列表、菜单列表
- 权限控制形同虚设
- 数据泄露风险

**修复方案**：
```go
// 正确做法
system.GET("/roles",
    middlewares.PermissionMiddleware("/api/system/roles", "GET"),
    roleController.GetRoles)

system.GET("/menus",
    middlewares.PermissionMiddleware("/api/system/menus", "GET"),
    menuController.GetMenus)

system.GET("/menus/tree",
    middlewares.PermissionMiddleware("/api/system/menus/tree", "GET"),
    menuController.GetMenuTree)
```

### 问题2: 命名规范不统一

**后端命名**：
```go
// 路由定义
system.GET("/users", userController.GetUsers)        // 使用复数
system.GET("/menus/tree", menuController.GetMenuTree) // 使用树形结构
```

**前端命名**：
```typescript
// API函数命名
fetchGetUserList()     // "Get" 前缀
fetchCreateUser()    // 没有 "Get" 前缀
fetchGetAllRoles()    // "GetAll" 前缀
```

**问题**：
- ❌ 前端命名不统一（有的有Get前缀，有的没有）
- ❌ "GetAll" vs "Get" 用法不一致
- ❌ 权限码和API路径不匹配

**统一命名规范**：
```typescript
// 推荐命名规范
fetchUserList()        // 获取列表（复数）
fetchUserById(id)     // 获取详情
fetchCreateUser()     // 创建
fetchUpdateUser(id)  // 更新
fetchDeleteUser(id)   // 删除
```

### 问题3: 权限码与API路径不匹配

**当前问题**：
```go
// 权限码使用点号分隔
PermissionMiddleware("system.user.create")

// 但API路径使用斜杠分隔
// 实际Casbin检查可能是：
enforcer.Enforce("ops", "system.user.create", "*")
// 而不是API路径
```

**Level 4权限策略应该**：
```go
// 正确的API级权限检查
system.GET("/users",
    middlewares.APIMiddleware("/api/system/users", "GET"),
    userController.GetUsers)

// 对应的Casbin策略
p, ops, /api/system/users, GET
p, ops, /api/system/users, POST
```

### 问题4: 代码组织混乱

**routes.go 文件问题**：
- ❌ 518行代码过长
- ❌ 所有路由混在一起
- ❌ 没有按功能模块分组
- ❌ 难以维护和查找

**推荐重构方案**：
```
routes/
├── routes.go              # 主路由入口
├── system_routes.go      # 系统管理路由
├── cmdb_routes.go        # CMDB路由
├── k8s_routes.go         # K8s路由
└── audit_routes.go      # 审计路由
```

### 问题5: 数据库命名不统一

**发现的问题**：
```sql
-- 不统一的命名
role_ids    -- 使用下划线
menu_ids    -- 使用下划线（已废弃）
home_path   -- 使用下划线
password    -- 无下划线
```

**统一规范**：
```sql
-- 推荐使用下划线命名
role_ids
menu_ids (废弃后需要清理)
home_path
created_at
updated_at
```

---

## 🎯 Level 4 API级权限实现方案

### 权限策略设计

#### Casbin策略表结构
```csv
p_type, v0 (角色), v1 (API路径), v2 (HTTP方法)
p, admin, *, *                               ← 超级管理员
p, ops, /api/system/users, GET
p, ops, /api/system/users, POST
p, ops, /api/system/users/:id, PUT
p, ops, /api/system/users/:id, DELETE
p, ops, /api/system/users/:id/password, PUT
p, ops, /api/system/roles, GET
p, ops, /api/system/roles, POST
p, ops, /api/system/roles/:id, PUT
p, ops, /api/system/roles/:id, DELETE
p, ops, /api/system/menus, GET
p, ops, /api/system/menus/tree, GET
p, ops, /api/system/menus, POST
p, ops, /api/system/menus/:id, PUT
p, ops, /api/system/menus/:id, DELETE
```

#### 修改后的路由配置
```go
// 用户管理路由
system.GET("/users",
    middlewares.APIMiddleware("/api/system/users", "GET"),
    userController.GetUsers)
system.POST("/users",
    middlewares.APIMiddleware("/api/system/users", "POST"),
    userController.CreateUser)
system.PUT("/users/:id",
    middlewares.APIMiddleware("/api/system/users/:id", "PUT"),
    userController.UpdateUser)
system.DELETE("/users/:id",
    middlewares.APIMiddleware("/api/system/users/:id", "DELETE"),
    userController.DeleteUser)
system.PUT("/users/:id/password",
    middlewares.APIMiddleware("/api/system/users/:id/password", "PUT"),
    userController.ResetPassword)
```

---

## 📝 前端API调用逻辑分析

### 当前前端调用逻辑

#### 用户创建流程
```typescript
async function handleCreateUser(userData) {
  // 1. 创建用户
  await fetchCreateUser(userData)
  // ↑ 调用 POST /api/system/users
  // ↑ 应该检查 ops, /api/system/users, POST 权限

  // 2. 刷新用户列表
  await fetchGetUserList()
  // ↑ 调用 GET /api/system/users
  // ↑ 应该检查 ops, /api/system/users, GET 权限

  // 3. 获取角色列表（用于下次创建）
  await fetchGetAllRoles()
  // ↑ 调用 GET /api/system/roles
  // ↑ ❌ 当前没有权限保护！
}
```

#### 问题：级联API调用

**前端一次操作可能触发多个API调用**：
1. 主操作API
2. 辅助数据API（角色列表、菜单树等）
3. 刷新列表API

**当前的权限检查缺陷**：
- ❌ 只检查主操作API的权限
- ❌ 不检查辅助数据API的权限
- ❌ 可能导致前端看到"无权限"错误，但不知道是哪个API

### 解决方案

#### 前端统一错误处理
```typescript
// 统一的API调用包装器
async function callWithPermission<T>(
  apiCall: () => Promise<T>,
  errorMessage: string
): Promise<T> {
  try {
    return await apiCall()
  } catch (error: any) {
    if (error.code === 403) {
      showPermissionError(errorMessage)
    }
    throw error
  }
}

// 使用示例
await callWithPermission(
  () => fetchCreateUser(userData),
  '创建用户'
)
```

#### 后端权限检查统一
```go
// 所有API都用API级别权限检查
// 包括查询接口，不仅是修改接口
system.GET("/roles",
    middlewares.APIMiddleware("/api/system/roles", "GET"),
    roleController.GetRoles)
```

---

## 🔧 推荐重构方案

### 1. 代码文件重构

#### 拆分routes.go
```
routes/
├── main.go                 # 主路由入口
├── system/
│   ├── routes.go          # 系统管理主路由
│   ├── user_routes.go     # 用户路由
│   ├── role_routes.go     # 角色路由
│   ├── menu_routes.go     # 菜单路由
│   └── permission_routes.go  # 权限路由
└── other/
    ├── cmdb_routes.go
    ├── k8s_routes.go
    └── audit_routes.go
```

### 2. API接口标准化

#### 统一API函数命名
```typescript
// 当前（不统一）
fetchGetUserList()
fetchCreateUser()
fetchGetAllRoles()

// 推荐（统一）
fetchUserList()
fetchUserById(id)
fetchCreateUser(data)
fetchUpdateUser(id, data)
fetchDeleteUser(id)
```

#### 统一权限码命名
```go
// 当前（操作级）
system.user.create

// Level 4（API级）
/api/system/users:POST:write
/api/system/users:GET:read
```

### 3. 数据库字段统一

#### 统一使用下划线命名
```sql
-- 统一后
role_ids (保持)
menu_ids (废弃后删除)
home_path (保持)
api_path (新增，用于权限映射)
```

---

## 📋 实施计划

### Phase 1: 权限检查完善
1. 为所有GET接口添加权限保护
2. 实现API级权限中间件
3. 更新Casbin策略

### Phase 2: 代码重构
1. 拆分routes.go文件
2. 统一API函数命名
3. 优化前端调用逻辑

### Phase 3: 数据库优化
1. 统一字段命名规范
2. 清理废弃字段
3. 添加必要的索引

### Phase 4: 测试验证
1. 权限检查测试
2. 前端调用测试
3. 错误处理测试

---

## 🎯 总结

### 当前主要问题
1. ❌ **权限检查不完整**：GET接口缺少权限保护
2. ❌ **命名规范不统一**：前后端API命名不一致
3. ❌ **代码组织混乱**：路由文件过长，功能混杂
4. ❌ **前端调用逻辑复杂**：错误处理不统一

### Level 4 API级权限的优势
1. ✅ **最细粒度控制**：每个API单独控制
2. ✅ **安全级别最高**：精确控制访问权限
3. ✅ **易于调试**：权限问题容易定位
4. ✅ **灵活性最强**：支持复杂权限组合

### 建议优先级
1. 🔴 高优先级：完善权限检查，确保所有API都有保护
2. 🟡 中优先级：代码重构，提升可维护性
3. 🟢 低优先级：命名规范统一，提升一致性

需要我开始实施这些修复吗？
