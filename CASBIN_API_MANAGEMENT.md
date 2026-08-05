# 基于Casbin的API权限管理系统

## 核心设计理念

**直接操作 `casbin_rule` 表，充分利用Casbin的存储能力**

### 为什么直接操作 casbin_rule？

1. **避免数据冗余**：不需要额外的 `api_resources` 表
2. **统一数据源**：所有权限信息集中在 `casbin_rule` 表中
3. **简化维护**：减少表间关联，降低复杂度
4. **原生支持**：充分利用Casbin的存储机制

## Casbin策略格式扩展

### 标准策略格式
```sql
-- 简单格式（v0-v2）
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'admin', '/api/users', 'GET');
```

### 扩展策略格式（包含业务信息）
```sql
-- 扩展格式（v0-v5）
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES
('p', 'admin', '/api/users', 'GET', '用户列表', '获取用户列表', 'system');

-- 字段说明：
-- ptype: 策略类型（p = 策略, g = 角色继承）
-- v0:   角色代码
-- v1:   API路径
-- v2:   HTTP方法
-- v3:   API名称（业务信息）
-- v4:   API描述（业务信息）
-- v5:   模块名称（业务信息）
```

## 核心组件

### 1. CasbinAPIManager（服务层）

```go
// 直接基于casbin_rule的API管理器
type CasbinAPIManager struct {
    enforcer *casbin.Enforcer
}

// 主要功能：
- GetUniqueAPIResources()     // 从策略中提取唯一API列表
- GetAllPolicies()            // 获取所有策略
- AssignPermissionToRole()    // 添加权限（直接写casbin_rule）
- RevokePermissionFromRole()  // 删除权限（直接操作casbin_rule）
- BatchAssignPermissions()    // 批量操作
```

### 2. CasbinAPIController（控制器层）

```go
// 提供HTTP接口
type CasbinAPIController struct{}

// 主要接口：
- GET  /api/casbin/api-resources    // 获取所有API资源
- GET  /api/casbin/policies         // 获取所有策略
- POST /api/casbin/assign           // 分配权限
- DELETE /api/casbin/revoke         // 撤销权限
- POST /api/casbin/batch-assign     // 批量分配
- POST /api/casbin/sync             // 同步常用API
```

## 使用示例

### 1. 添加API资源（带业务信息）

```bash
POST /api/casbin/assign
Content-Type: application/json

{
  "roleCode": "operator",
  "path": "/api/system/users",
  "method": "GET",
  "name": "用户列表",
  "description": "获取用户列表，支持分页和搜索",
  "module": "system"
}

# 对应的casbin_rule记录：
# ptype | v0      | v1                | v2   | v3      | v4                    | v5
# p     | operator| /api/system/users | GET  | 用户列表| 获取用户列表，支持... | system
```

### 2. 获取所有API资源

```bash
GET /api/casbin/api-resources

Response:
{
  "code": 200,
  "data": [
    {
      "path": "/api/system/users",
      "method": "GET",
      "name": "用户列表",
      "description": "获取用户列表",
      "module": "system"
    },
    {
      "path": "/api/system/users",
      "method": "POST",
      "name": "创建用户",
      "description": "创建新用户",
      "module": "system"
    }
  ]
}
```

### 3. 获取角色权限

```bash
GET /api/casbin/role-permissions?roleCode=operator

Response:
{
  "code": 200,
  "data": [
    {
      "path": "/api/system/users",
      "method": "GET",
      "name": "用户列表",
      "description": "获取用户列表",
      "module": "system"
    }
  ]
}
```

### 4. 批量分配权限

```bash
POST /api/casbin/batch-assign
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

### 5. 撤销权限

```bash
DELETE /api/casbin/revoke
Content-Type: application/json

{
  "roleCode": "operator",
  "path": "/api/system/users",
  "method": "POST"
}
```

## 权限检查流程

### 1. 中间件检查
```go
// middleware/api_auth.go
func APIAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.Get("userID")
        path := c.Request.URL.Path
        method := c.Request.Method

        // 获取用户角色
        roles := getUserRoles(userID)

        // 检查任意角色是否有权限
        for _, role := range roles {
            allowed, _ := enforcer.Enforce(role.Code, path, method)
            if allowed {
                c.Next()
                return
            }
        }

        c.JSON(403, gin.H{"error": "无权限"})
        c.Abort()
    }
}
```

### 2. Casbin内部检查
```go
// Casbin检查逻辑：
// 1. 从casbin_rule表加载策略
// 2. 匹配：g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
// 3. 返回是否有权限

// 示例：
enforcer.Enforce("operator", "/api/system/users", "GET")
// 检查 casbin_rule 中是否存在：
// p, operator, /api/system/users, GET
```

## 数据库操作示例

### 直接操作 casbin_rule 表

```sql
-- 1. 查看所有策略
SELECT * FROM casbin_rule WHERE ptype = 'p';

-- 2. 查看特定角色的权限
SELECT * FROM casbin_rule WHERE ptype = 'p' AND v0 = 'admin';

-- 3. 查看特定API的权限配置
SELECT * FROM casbin_rule WHERE ptype = 'p' AND v1 = '/api/users' AND v2 = 'GET';

-- 4. 添加权限策略
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5) VALUES
('p', 'operator', '/api/users', 'GET', '用户列表', '获取用户列表', 'system');

-- 5. 删除权限策略
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'operator' AND v1 = '/api/users' AND v2 = 'GET';

-- 6. 批量操作（删除角色所有权限）
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'operator';
```

## 优势总结

### 相比之前的双表方案

**旧方案（api_resources + casbin_rule）：**
```sql
-- 需要维护两个表
SELECT * FROM api_resources WHERE path = '/api/users';
SELECT * FROM casbin_rule WHERE v1 = '/api/users';
-- 需要同步两个表的数据
```

**新方案（仅使用casbin_rule）：**
```sql
-- 只需操作一个表
SELECT * FROM casbin_rule WHERE v1 = '/api/users' AND v2 = 'GET';
-- 所有信息都在一个表中
```

### 实际好处

1. **简化架构**：只需关注 `casbin_rule` 一个表
2. **减少同步**：不需要在多个表间同步数据
3. **原生支持**：充分利用Casbin的设计
4. **易于维护**：数据结构简单，查询直接
5. **性能更好**：减少表关联查询

## 路由配置

```go
// 在 routes/routes.go 中添加
casbinAPIController := controllers.NewCasbinAPIController()

casbin := api.Group("/casbin")
casbin.Use(middlewares.Auth())
{
    casbin.GET("/api-resources", casbinAPIController.GetAllAPIResources)
    casbin.GET("/policies", casbinAPIController.GetAllPolicies)
    casbin.POST("/assign", casbinAPIController.AssignPermissionToRole)
    casbin.DELETE("/revoke", casbinAPIController.RevokePermissionFromRole)
    casbin.POST("/batch-assign", casbinAPIController.BatchAssignPermissions)
    casbin.GET("/role-permissions", casbinAPIController.GetRolePermissions)
    casbin.POST("/sync", casbinAPIController.SyncCommonAPIs)
}
```

## 初始化数据

```go
// 在 init_api_management.go 中
func SeedDefaultPolicies() error {
    manager, _ := services.NewCasbinAPIManager()

    // 为admin角色添加默认权限
    defaultAPIs := []struct{
        path, method, name, description, module string
    }{
        {"/api/system/users", "GET", "用户列表", "获取用户列表", "system"},
        {"/api/system/users", "POST", "创建用户", "创建新用户", "system"},
        // ... 更多API
    }

    for _, api := range defaultAPIs {
        manager.AddAPIWithMetadata("admin", api.path, api.method, api.name, api.description, api.module)
    }

    return manager.enforcer.SavePolicy()
}
```

## 总结

通过直接操作 `casbin_rule` 表，我们：

1. ✅ **简化了数据模型**：只需维护一个表
2. ✅ **统一了数据源**：所有权限信息集中存储
3. ✅ **提高了性能**：减少表关联查询
4. ✅ **增强了可维护性**：数据结构简单清晰
5. ✅ **保持了灵活性**：充分利用Casbin的功能

这种方案更加符合Casbin的设计理念，是API权限管理的最佳实践。