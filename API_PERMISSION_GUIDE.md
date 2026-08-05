# API权限管理系统使用指南

## 概述

本系统提供基于Casbin的API级别权限校验和可视化的API管理功能。

## 架构说明

### 数据表设计

1. **casbin_rule** - Casbin自动创建，存储权限策略
   ```sql
   -- 策略格式：p, role_code, /api/path, METHOD
   INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
   ('p', 'admin', '/api/system/users', 'GET'),
   ('p', 'admin', '/api/system/users', 'POST');
   ```

2. **api_resources** - API资源元数据表
   ```sql
   CREATE TABLE `api_resources` (
     `id` bigint unsigned NOT NULL AUTO_INCREMENT,
     `name` varchar(100) NOT NULL COMMENT 'API名称',
     `path` varchar(200) NOT NULL COMMENT 'API路径',
     `method` varchar(10) NOT NULL COMMENT 'HTTP方法',
     `module` varchar(50) DEFAULT NULL COMMENT '所属模块',
     `category` varchar(50) DEFAULT NULL COMMENT 'API分类',
     `status` int DEFAULT '1' COMMENT '状态:1启用,0禁用',
     PRIMARY KEY (`id`),
     UNIQUE KEY `uk_path_method` (`path`, `method`)
   );
   ```

### 核心组件

- **数据模型**: `models.APIResource`
- **服务层**: `APIResourceManager` - 封装Casbin操作
- **控制器**: `APIResourceController` - 提供HTTP接口
- **中间件**: `APIAuthMiddleware` - API权限校验

## 使用步骤

### 1. 初始化系统

```go
package main

import (
    "fmt"
    "oneops/backend/init"
)

func main() {
    // 初始化API管理系统
    if err := init.InitializeAPIManagement(); err != nil {
        fmt.Printf("初始化失败: %v\n", err)
        return
    }
    fmt.Println("API管理系统初始化成功")
}
```

### 2. 应用API权限中间件

```go
package routes

import (
    "github.com/gin-gonic/gin"
    "oneops/backend/controllers"
    "oneops/backend/middleware"
)

func SetupRoutes(r *gin.Engine) {
    api := r.Group("/api")

    // 应用API权限校验中间件
    system := api.Group("/system")
    system.Use(middleware.Auth()) // 先认证
    system.Use(middleware.APIAuthMiddleware()) // 再校验API权限

    // ... 路由配置
}
```

### 3. API权限校验流程

```go
// 1. 用户访问 API: GET /api/system/users
// 2. 中间件检查：
//    - 获取用户ID（从token）
//    - 获取请求路径和方法
//    - 调用 PermissionService.HasAPIPermission(userID, "/api/system/users", "GET")
//    - Casbin检查：enforcer.Enforce("admin", "/api/system/users", "GET")
// 3. 返回结果：有权限放行，无权限返回403
```

## API接口说明

### API资源管理

#### 1. 获取API资源列表
```bash
GET /api/system/api-resources?current=1&size=10&module=system

Response:
{
  "code": 200,
  "message": "success",
  "data": {
    "records": [
      {
        "id": 1,
        "name": "用户列表",
        "path": "/api/system/users",
        "method": "GET",
        "module": "system",
        "category": "用户管理",
        "status": 1
      }
    ],
    "current": 1,
    "size": 10,
    "total": 50
  }
}
```

#### 2. 创建API资源
```bash
POST /api/system/api-resources
Content-Type: application/json

{
  "name": "用户列表",
  "path": "/api/system/users",
  "method": "GET",
  "module": "system",
  "category": "用户管理",
  "description": "获取用户列表",
  "status": 1
}
```

#### 3. 更新API资源
```bash
PUT /api/system/api-resources/1
Content-Type: application/json

{
  "name": "用户列表（修改）",
  "description": "获取用户列表，支持分页"
}
```

#### 4. 删除API资源
```bash
DELETE /api/system/api-resources/1
```

### API权限分配

#### 1. 分配权限给角色
```bash
POST /api/system/api-permissions/assign
Content-Type: application/json

{
  "roleCode": "operator",
  "apiResourceId": 1
}

Response:
{
  "code": 200,
  "message": "分配成功"
}
```

#### 2. 批量分配权限
```bash
POST /api/system/api-permissions/batch-assign
Content-Type: application/json

{
  "roleCode": "operator",
  "apiResourceIds": [1, 2, 3, 5, 8]
}
```

#### 3. 获取角色的API权限
```bash
GET /api/system/api-permissions/role?roleCode=operator

Response:
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "用户列表",
      "path": "/api/system/users",
      "method": "GET"
    }
  ]
}
```

#### 4. 获取角色权限映射
```bash
GET /api/system/api-permissions/role/map?roleCode=operator

Response:
{
  "code": 200,
  "data": {
    "GET": ["/api/system/users", "/api/system/roles"],
    "POST": ["/api/system/users"]
  }
}
```

#### 5. 撤销权限
```bash
DELETE /api/system/api-permissions/1/revoke?roleCode=operator
```

### API自动发现

#### 1. 自动发现API
```bash
GET /api/system/api-resources/discover

Response:
{
  "code": 200,
  "data": [
    {
      "name": "用户列表",
      "path": "/api/system/users",
      "method": "GET",
      "module": "system",
      "category": "用户管理"
    }
  ]
}
```

#### 2. 同步API到数据库
```bash
POST /api/system/api-resources/sync

Response:
{
  "code": 200,
  "message": "同步成功"
}
```

## Casbin策略说明

### 策略格式
```ini
# config/casbin_model.conf
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

### 策略示例
```sql
-- casbin_rule 表数据
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
-- ptype, role_code, api_path, method
('p', 'admin', '/api/system/users', 'GET'),
('p', 'admin', '/api/system/users', 'POST'),
('p', 'admin', '/api/system/users', 'PUT'),
('p', 'admin', '/api/system/users', 'DELETE'),
('p', 'operator', '/api/system/users', 'GET'),
('p', 'operator', '/api/system/roles', 'GET');
```

## 前端集成示例

### 1. API资源管理页面
```vue
<template>
  <div class="api-management">
    <!-- API资源列表 -->
    <el-table :data="apiResources">
      <el-table-column prop="name" label="API名称" />
      <el-table-column prop="path" label="路径" />
      <el-table-column prop="method" label="方法" />
      <el-table-column prop="module" label="模块" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button @click="editAPI(row)">编辑</el-button>
          <el-button @click="deleteAPI(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 权限分配 -->
    <el-button @click="assignPermissions">分配权限</el-button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const apiResources = ref([])

// 获取API列表
const fetchAPIResources = async () => {
  const response = await fetch('/api/system/api-resources')
  const data = await response.json()
  apiResources.value = data.data.records
}

// 分配权限
const assignPermissions = async (roleId, apiIds) => {
  await fetch('/api/system/api-permissions/batch-assign', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      roleCode: roleId,
      apiResourceIds: apiIds
    })
  })
}

onMounted(() => {
  fetchAPIResources()
})
</script>
```

### 2. 权限检查示例
```typescript
// 检查用户是否有API权限
const checkAPIPermission = async (path, method) => {
  const response = await fetch('/api/system/check-permission', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path, method })
  })
  return response.json()
}

// 按钮权限控制
<el-button
  v-if="hasPermission('/api/system/users', 'POST')"
  @click="createUser"
>
  新建用户
</el-button>
```

## 权限配置最佳实践

### 1. API分组管理
```sql
-- 按模块分类API资源
INSERT INTO api_resources (name, path, method, module, category) VALUES
-- 用户管理模块
('用户列表', '/api/system/users', 'GET', 'system', '用户管理'),
('创建用户', '/api/system/users', 'POST', 'system', '用户管理'),

-- 角色管理模块
('角色列表', '/api/system/roles', 'GET', 'system', '角色管理'),
('创建角色', '/api/system/roles', 'POST', 'system', '角色管理');
```

### 2. 角色权限模板
```sql
-- 管理员角色：拥有所有权限
-- 操作员角色：只读权限
-- 审计员角色：审计日志权限
```

### 3. 权限继承
```ini
# 可以通过角色继承实现权限复用
# casbin_role_definition
g, operator_admin, operator
g, operator_manager, operator_admin
```

## 故障排查

### 1. 权限不生效
```bash
# 检查 Casbin 策略是否加载
SELECT * FROM casbin_rule WHERE ptype = 'p';

# 检查API资源是否正确配置
SELECT * FROM api_resources WHERE path = '/api/system/users';

# 检查用户角色是否正确
SELECT * FROM users WHERE id = 1;
```

### 2. 中间件顺序错误
```go
// 正确的中间件顺序
system.Use(middlewares.Auth())              // 1. 认证
system.Use(middlewares.APIAuthMiddleware()) // 2. 权限校验
```

### 3. 路径匹配问题
```bash
# 确保API路径配置一致
# 前端: /api/system/users
# 后端: /api/system/users
# 数据库: /api/system/users
```

## 性能优化建议

1. **启用Casbin缓存**
2. **使用批量操作**避免频繁的数据库查询
3. **合理设计API路径**避免过度细分权限
4. **定期清理无效的策略**

## 总结

本系统提供了完整的API级别权限管理解决方案：
- ✅ 基于Casbin的标准RBAC模型
- ✅ 可视化的API资源管理
- ✅ 灵活的权限分配机制
- ✅ 自动API发现和同步功能
- ✅ 完善的中间件集成

通过这套系统，你可以轻松实现细粒度的API权限控制，保障系统的安全性。