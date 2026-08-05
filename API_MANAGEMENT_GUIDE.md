# API权限管理使用指南（方案2：直接操作casbin_rule）

## 核心设计

**只使用 `casbin_rule` 表存储所有权限信息**

### casbin_rule 表结构

```sql
CREATE TABLE `casbin_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(255) DEFAULT NULL,  -- 策略类型：p(策略) 或 g(角色继承)
  `v0` varchar(255) DEFAULT NULL,     -- 角色代码
  `v1` varchar(255) DEFAULT NULL,     -- API路径
  `v2` varchar(255) DEFAULT NULL,     -- HTTP方法
  `v3` varchar(255) DEFAULT NULL,     -- API名称（业务信息）
  `v4` varchar(255) DEFAULT NULL,     -- API描述（业务信息）
  `v5` varchar(255) DEFAULT NULL,     -- 模块名称（业务信息）
  PRIMARY KEY (`id`)
);
```

### 策略示例

```sql
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES
('p', 'admin', '/api/system/users', 'GET', '用户列表', '获取用户列表', 'system'),
('p', 'admin', '/api/system/users', 'POST', '创建用户', '创建新用户', 'system'),
('p', 'operator', '/api/system/users', 'GET', '用户列表', '获取用户列表', 'system');
```

## 核心组件

### 1. CasbinAPIManager（服务层）

位置：`backend/services/casbin_api_manager.go`

主要方法：
- `AssignPermissionToRole()` - 添加权限
- `RevokePermissionFromRole()` - 撤销权限
- `GetRolePermissions()` - 获取角色权限
- `BatchAssignPermissions()` - 批量分配
- `CheckPermission()` - 权限检查
- `GetUniqueAPIResources()` - 获取所有API资源
- `SyncCommonAPIs()` - 同步常用API

### 2. CasbinAPIController（控制器层）

位置：`backend/controllers/casbin_api.go`

HTTP接口：
- `GET /api/system/casbin/api-resources` - 获取所有API资源
- `GET /api/system/casbin/policies` - 获取所有策略
- `POST /api/system/casbin/assign` - 分配权限
- `DELETE /api/system/casbin/revoke` - 撤销权限
- `POST /api/system/casbin/batch-assign` - 批量分配
- `GET /api/system/casbin/role-permissions` - 获取角色权限
- `POST /api/system/casbin/sync` - 同步常用API
- `POST /api/system/casbin/check` - 检查权限

## 快速开始

### 1. 初始化

在 `main.go` 中添加：

```go
import "oneops/backend/services"

// 在数据库初始化后添加
if err := services.InitializeAPIManagement(); err != nil {
    logger.Warn("API权限管理初始化失败", zap.Error(err))
}
```

### 2. 权限校验中间件

已创建：`backend/middleware/api_auth.go`

使用方式：

```go
import "oneops/backend/middleware"

// 在需要权限校验的路由组上应用
api.Use(middleware.APIAuthMiddleware())
```

## API使用示例

### 1. 为角色分配权限

```bash
POST /api/system/casbin/assign
Content-Type: application/json

{
  "roleCode": "operator",
  "path": "/api/system/users",
  "method": "GET",
  "name": "用户列表",
  "description": "获取用户列表，支持分页",
  "module": "system"
}
```

### 2. 批量分配权限

```bash
POST /api/system/casbin/batch-assign
Content-Type: application/json

{
  "roleCode": "operator",
  "apis": [
    {
      "path": "/api/system/users",
      "method": "GET",
      "name": "用户列表",
      "description": "获取用户列表",
      "module": "system"
    },
    {
      "path": "/api/system/roles",
      "method": "GET",
      "name": "角色列表",
      "description": "获取角色列表",
      "module": "system"
    }
  ]
}
```

### 3. 获取角色权限

```bash
GET /api/system/casbin/role-permissions?roleCode=operator
```

### 4. 撤销权限

```bash
DELETE /api/system/casbin/revoke
Content-Type: application/json

{
  "roleCode": "operator",
  "path": "/api/system/users",
  "method": "POST"
}
```

### 5. 检查权限

```bash
POST /api/system/casbin/check
Content-Type: application/json

{
  "roleCode": "operator",
  "path": "/api/system/users",
  "method": "GET"
}
```

### 6. 同步常用API

```bash
POST /api/system/casbin/sync
```

## 权限检查流程

### 1. HTTP请求到达

```go
// 用户访问：GET /api/system/users
// 中间件自动执行：
1. 获取用户ID（从JWT token）
2. 获取用户角色列表
3. 对每个角色调用 Casbin.Enforce(role, path, method)
4. 任一角色有权限即放行
```

### 2. Casbin内部流程

```go
// Casbin执行：
1. 从 casbin_rule 表加载策略
2. 匹配策略：role == v0 && path == v1 && method == v2
3. 返回 true/false
```

## 代码示例

### Go代码中使用

```go
package main

import (
    "fmt"
    "oneops/backend/services"
)

func main() {
    // 创建管理器
    manager, err := services.NewCasbinAPIManager()
    if err != nil {
        panic(err)
    }

    // 分配权限
    err = manager.AssignPermissionToRole(
        "operator",           // 角色
        "/api/system/users",  // API路径
        "GET",                // HTTP方法
        "用户列表",            // API名称
        "获取用户列表",        // 描述
        "system",             // 模块
    )

    // 检查权限
    allowed, _ := manager.CheckPermission("operator", "/api/system/users", "GET")
    fmt.Printf("有权限: %v\n", allowed)

    // 批量分配
    apis := []services.APIResource{
        {
            Path:        "/api/system/roles",
            Method:      "GET",
            Name:        "角色列表",
            Description: "获取角色列表",
            Module:      "system",
        },
    }
    manager.BatchAssignPermissions("operator", apis)
}
```

## 数据库操作

### 直接操作 casbin_rule 表

```sql
-- 查看所有权限
SELECT * FROM casbin_rule WHERE ptype = 'p';

-- 查看特定角色的权限
SELECT * FROM casbin_rule WHERE ptype = 'p' AND v0 = 'operator';

-- 查看特定API的权限
SELECT * FROM casbin_rule WHERE v1 = '/api/system/users' AND v2 = 'GET';

-- 手动添加权限
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES
('p', 'operator', '/api/system/users', 'GET', '用户列表', '获取用户列表', 'system');

-- 删除权限
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'operator' AND v1 = '/api/system/users' AND v2 = 'GET';
```

## 优势总结

相比双表方案，这个方案：

✅ **更简洁** - 只需维护一个表
✅ **避免冗余** - 无重复数据
✅ **性能更好** - 无表关联查询
✅ **原生支持** - 充分利用Casbin设计
✅ **易于维护** - 数据结构简单清晰

## 文件清单

```
backend/
├── services/
│   ├── casbin_api_manager.go       # 核心服务
│   ├── init_api_management.go      # 初始化脚本
│   └── api_resource_test.go        # 测试文件
├── controllers/
│   └── casbin_api.go               # HTTP接口
├── middleware/
│   └── api_auth.go                 # 权限校验中间件
└── routes/
    └── routes.go                   # 路由配置
```

## 测试

```bash
# 运行测试
cd backend
go test ./services -v -run TestCasbinAPIManager

# 测试单个功能
go test ./services -v -run TestCasbinAPIManager/AssignPermission
```

## 故障排查

### 权限不生效

```bash
# 1. 检查策略是否存在
SELECT * FROM casbin_rule WHERE ptype = 'p' AND v0 = 'role_code';

# 2. 检查用户角色
SELECT role_ids FROM users WHERE id = 1;

# 3. 检查Casbin是否加载
SELECT COUNT(*) FROM casbin_rule;
```

### 中间件问题

确保中间件顺序正确：

```go
// 正确顺序
api.Use(middleware.Auth())              // 1. 认证
api.Use(middleware.APIAuthMiddleware()) // 2. 权限校验
```

## 总结

这个方案通过直接操作 `casbin_rule` 表，实现了：

- ✅ 简单的数据模型
- ✅ 统一的权限存储
- ✅ 高效的权限检查
- ✅ 灵活的管理接口
- ✅ 完整的CRUD功能

适合中小型项目的API权限管理需求。