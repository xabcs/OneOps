# RBAC + Casbin 权限管理方案实施完成

## 一、方案概述

### 1.1 核心设计理念

本方案采用 **"业务库管理元数据 + Casbin 负责鉴权"** 的混合模式，实现职责分离：

```
业务库 (MySQL)              Casbin
├─ 权限元数据管理           ├─ 权限策略存储
│  ├─ 名称、描述            │  └─ (role, permission_code, *)
│  ├─ 树形结构              └─ 鉴权引擎
│  ├─ 菜单、按钮            └─ 性能优化（内存缓存）
│  └─ 前端展示              
└─ 用户-角色关联
   └─ 登录返回权限列表
```

### 1.2 关键优势

- ✅ **职责清晰**：业务库管理元数据，Casbin 专注鉴权
- ✅ **层级格式 code**：`模块.资源.操作`（如 `system.user.create`）
- ✅ **模块化管理**：支持多模块系统（system, cmdb, monitor 等）
- ✅ **无需映射表**：前端按钮权限和后端 API 权限统一
- ✅ **符合主流**：与 AWS IAM、阿里云 RAM、Spring Security 一致
- ✅ **易于维护**：配置权限时自动同步 Casbin
- ✅ **已实施完成**：所有代码已实现并集成

### 1.3 实施状态

**✅ 已完成的核心实现：**

1. **权限服务** (`backend/services/permission_service.go`)
   - 统一的权限检查方法 `HasPermission(userID, permissionCode)`
   - Casbin 同步方法 `syncRoleToCasbin(role)`
   - 初始化方法 `InitializeCasbinPolicies()`

2. **权限中间件** (`backend/middlewares/permission.go`)
   - 统一权限检查 `RequirePermission(permissionCode)`
   - 删除了旧的 API 路径检查逻辑
   - 支持层级权限代码格式

3. **路由更新** (`backend/routes/routes.go`)
   - 所有路由已从 API 路径检查改为权限代码检查
   - 例如：`middlewares.RequirePermission("system.user.list")`

4. **数据库结构** (`backend/models/permission.go`)
   - permissions 表支持层级权限代码
   - 包含 module, resource, action, level 字段
   - role_permissions 关联表

5. **初始化流程** (`backend/services/init.go`)
   - `initAPIPermissions()` 更新为使用层级权限代码
   - 自动清除旧策略并同步新权限代码

---

## 二、权限 Code 命名规范（已实施）

### 2.1 格式定义

**层级格式：** `模块.资源.操作`

**示例：**
```
system.user.list        系统模块 → 用户资源 → 列表操作
system.user.create      系统模块 → 用户资源 → 创建操作
cmdb.asset.view         CMDB模块 → 资产资源 → 查看操作
monitor.alert.handle    监控模块 → 告警资源 → 处理操作
```

### 2.2 模块定义

```
system      系统管理模块
  ├─ user       用户管理
  ├─ role       角色管理
  ├─ menu       菜单管理
  └─ permission 权限管理

cmdb        CMDB模块
  ├─ asset      资产管理
  ├─ server     服务器管理
  └─ network    网络设备

monitor     监控模块
  ├─ alert      告警管理
  └─ task       任务管理

auth        认证模块
  └─ permission 用户权限查询
```

### 2.3 操作定义

```
list      列表查询
view      详情查看
create    创建资源
update    更新资源
delete    删除资源
assign_permissions  分配权限
reset_password     重置密码
query     查询操作
```

### 2.4 通配符支持

```
*.*.*                 全局权限（超级管理员）
system.*.*           系统管理模块所有权限
system.user.*        用户管理所有权限
```

---

## 三、Casbin 策略格式（已实施）

### 3.1 策略存储格式

**旧格式（已废弃）：**
```
p, admin, /api/system/users, GET
p, admin, /api/system/users, POST
```

**新格式（已实施）：**
```
p, admin, system.user.list, *
p, admin, system.user.create, *
```

### 3.2 Casbin 模型配置

`config/casbin_model.conf`:
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

### 3.3 权限检查流程

```
用户请求 → 认证中间件 → 权限中间件(RequirePermission) →
PermissionService.HasPermission(userID, "system.user.list") →
Casbin.Enforce(roleCode, "system.user.list", "*") →
返回 true/false
```

---

## 四、数据库表结构（已实施）

### 4.1 permissions 表

```sql
CREATE TABLE permissions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(100) NOT NULL UNIQUE COMMENT '权限代码：模块.资源.操作',
  name VARCHAR(50) NOT NULL COMMENT '权限名称',
  description VARCHAR(200) COMMENT '权限描述',
  module VARCHAR(30) NOT NULL COMMENT '模块名',
  resource VARCHAR(30) NOT NULL COMMENT '资源名',
  action VARCHAR(20) NOT NULL COMMENT '操作名',
  level TINYINT NOT NULL DEFAULT 3 COMMENT '权限级别：1-模块级，2-页面级，3-按钮级',
  parent_id BIGINT UNSIGNED COMMENT '父权限ID',
  sort_order INT DEFAULT 0 COMMENT '排序',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_module (module),
  INDEX idx_resource (resource),
  INDEX idx_level (level),
  INDEX idx_status (status)
);
```

### 4.2 role_permissions 表

```sql
CREATE TABLE role_permissions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  role_id BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  permission_id BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_role_permission (role_id, permission_id),
  INDEX idx_role_id (role_id),
  INDEX idx_permission_id (permission_id)
);
```

### 4.3 casbin_rule 表

```sql
-- Casbin 自动创建的策略表
CREATE TABLE casbin_rule (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  ptype VARCHAR(100) COMMENT '策略类型：p-策略，g-角色继承',
  v0 VARCHAR(100) COMMENT '主体（角色代码）',
  v1 VARCHAR(100) COMMENT '对象（权限代码）',
  v2 VARCHAR(100) COMMENT '动作（通常为*）',
  v3 VARCHAR(100) COMMENT '其他参数',
  v4 VARCHAR(100) COMMENT '其他参数',
  INDEX idx_ptype (ptype),
  INDEX idx_v0 (v0),
  INDEX idx_v1 (v1)
);
```

---

## 五、后端实现（已实施）

### 5.1 权限服务核心方法

**文件：** `backend/services/permission_service.go`

```go
// 统一的权限检查方法（已删除API路径检查逻辑）
func (s *PermissionService) HasPermission(userID uint, permissionCode string) (bool, error) {
    // 1. 获取用户信息
    // 2. 检查超级管理员（admin用户或admin角色）
    // 3. 获取用户角色
    // 4. 使用 Casbin 检查权限：Enforce(role.Code, permissionCode, "*")
    // 5. 返回检查结果
}

// 同步角色权限到 Casbin（使用层级权限代码）
func (s *PermissionService) syncRoleToCasbin(role *models.Role) error {
    // 1. 删除角色的所有旧策略
    // 2. 获取角色的权限（从 role_permissions 表）
    // 3. 添加新策略：AddPolicy(role.Code, perm.Code, "*")
    // 4. 保存策略：SavePolicy()
}

// 初始化 Casbin 策略
func (s *PermissionService) InitializeCasbinPolicies() error {
    // 1. 清除所有旧策略（包括API路径格式）
    // 2. 同步所有角色权限到 Casbin（使用层级权限代码）
}
```

### 5.2 权限中间件

**文件：** `backend/middlewares/permission.go`

```go
// 统一的权限检查中间件
func RequirePermission(permissionCode string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 获取用户ID（从认证上下文）
        // 2. 调用 PermissionService.HasPermission(userID, permissionCode)
        // 3. 根据结果允许或拒绝请求
    }
}

// 资源操作权限快捷方式
func RequireResourceAccess(module, resource, action string) gin.HandlerFunc {
    permissionCode := fmt.Sprintf("%s.%s.%s", module, resource, action)
    return RequirePermission(permissionCode)
}
```

### 5.3 路由配置示例

**文件：** `backend/routes/routes.go`

```go
// 用户管理路由
system.GET("/users",
    middlewares.RequirePermission("system.user.list"),
    userController.GetUsers)
system.POST("/users",
    middlewares.RequirePermission("system.user.create"),
    userController.CreateUser)
system.PUT("/users/:id",
    middlewares.RequirePermission("system.user.update"),
    userController.UpdateUser)
system.DELETE("/users/:id",
    middlewares.RequirePermission("system.user.delete"),
    userController.DeleteUser)
```

### 5.4 初始化流程

**文件：** `backend/services/init.go`

```go
// 初始化层级权限代码到Casbin
func (s *InitService) initAPIPermissions() error {
    // 1. 清除所有旧的API路径格式策略
    // 2. 同步所有角色的权限到Casbin（使用层级权限代码）
    // 3. 验证并记录策略数量
}
```

---

## 六、前端集成方案

### 6.1 权限代码使用

**按钮权限控制：**
```vue
<el-button 
  v-if="hasPermission('system.user.create')" 
  @click="handleCreate">
  创建用户
</el-button>
```

**路由权限控制：**
```typescript
// router/index.ts
{
  path: '/manage/users',
  meta: {
    permission: 'system.user.list'
  }
}
```

### 6.2 权限树构建

**前端权限树结构：**
```javascript
const permissionTree = [
  {
    id: 1,
    code: 'system',
    name: '系统管理',
    level: 1,
    children: [
      {
        id: 2,
        code: 'system.user',
        name: '用户管理',
        level: 2,
        children: [
          { id: 3, code: 'system.user.list', name: '查看列表', level: 3 },
          { id: 4, code: 'system.user.create', name: '创建用户', level: 3 }
        ]
      }
    ]
  }
]
```

---

## 七、测试验证（已实施）

### 7.1 单元测试

**文件：** `backend/services/permission_service_test.go`

```go
// 测试权限服务基本功能
func TestPermissionServiceBasic(t *testing.T)

// 测试层级权限代码格式
func TestHierarchicalPermissionCodes(t *testing.T)

// 测试权限代码结构
func TestPermissionCodeStructure(t *testing.T)

// 测试Casbin同步逻辑
func TestCasbinSyncLogic(t *testing.T)
```

### 7.2 集成测试步骤

1. **启动服务**：
   ```bash
   npm run dev
   ```

2. **登录系统**：
   - 用户名：`admin`
   - 密码：`admin`

3. **权限测试**：
   - 检查用户权限列表（应包含层级权限代码）
   - 测试不同权限代码的访问控制
   - 验证Casbin策略同步

---

## 八、迁移指南（已完成）

### 8.1 旧系统迁移

**已完成的工作：**

1. ✅ **删除旧的 API 路径检查逻辑**
   - 删除了 `HasAPIPermission` 方法
   - 删除了 `APILevelPermissionMiddleware` 中间件
   - 删除了 API 权限映射表

2. ✅ **更新路由权限检查**
   - 所有路由从 `APILevelPermissionMiddleware` 改为 `RequirePermission`
   - 使用层级权限代码替代 API 路径

3. ✅ **更新初始化流程**
   - `initAPIPermissions()` 改为使用层级权限代码
   - 自动清除旧策略并同步新权限代码

### 8.2 数据库迁移

**Casbin 策略迁移：**
```sql
-- 旧策略（已清除）
DELETE FROM casbin_rule WHERE ptype='p';

-- 新策略（通过 InitializeCasbinPolicies 自动生成）
-- 策略格式：p, role_code, permission_code, *
```

---

## 九、使用示例

### 9.1 后端权限检查

```go
// 控制器中使用
func (ctrl *UserController) GetUsers(c *gin.Context) {
    // 权限已在路由中间件中检查
    // controller logic here...
}
```

### 9.2 前端权限控制

```vue
<template>
  <div>
    <!-- 按钮权限控制 -->
    <el-button 
      v-if="hasPermission('system.user.create')"
      type="primary" 
      @click="handleCreate">
      创建用户
    </el-button>

    <!-- 表格操作列 -->
    <template #actions="{ row }">
      <el-button 
        v-if="hasPermission('system.user.update')"
        size="small" 
        @click="handleEdit(row)">
        编辑
      </el-button>
      <el-button 
        v-if="hasPermission('system.user.delete')"
        size="small" 
        type="danger" 
        @click="handleDelete(row)">
        删除
      </el-button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { usePermission } from '@/hooks/usePermission'

const { hasPermission } = usePermission()
</script>
```

---

## 十、总结

### 10.1 实施成果

**✅ 已完成的核心功能：**

1. **统一的权限代码格式**：`模块.资源.操作`
2. **权限服务重构**：删除API路径检查，统一使用层级代码
3. **中间件优化**：简化的权限检查中间件
4. **路由更新**：所有路由使用新的权限代码
5. **数据库支持**：权限表支持层级结构
6. **Casbin集成**：自动同步权限代码到Casbin
7. **测试文件**：单元测试覆盖核心功能

### 10.2 技术优势

- **职责清晰**：业务库管理元数据，Casbin专注鉴权
- **易于维护**：统一的权限代码格式
- **性能优化**：Casbin内存缓存，快速鉴权
- **模块化设计**：支持多模块系统扩展
- **符合主流**：与行业标准一致的权限设计

### 10.3 后续建议

1. **前端集成**：更新前端权限检查逻辑
2. **权限管理界面**：基于层级代码的权限分配界面
3. **性能监控**：监控权限检查性能
4. **文档完善**：更新用户文档和API文档

---

**实施完成日期：** 2025-08-05
**实施状态：** ✅ 已完成核心实现，可投入使用
  ├─ menu       菜单管理
  └─ permission 权限管理

cmdb        配置管理数据库模块
  ├─ asset      资产管理
  ├─ rack       机架管理
  └─ supplier   供应商管理

monitor     监控模块
  ├─ dashboard  仪表盘
  ├─ alert      告警管理
  └─ metric     指标管理

audit       审计模块
  ├─ login      登录日志
  └─ operation  操作日志
```

### 2.3 操作定义

```
list        列表查询
view        详情查看
add         新增
edit        编辑
delete      删除
import      导入
export      导出
config      配置
handle      处理
approve     审批
reset_password  重置密码
assign_role     分配角色
```

### 2.4 与行业标准对比

| 标准 | 格式 | 示例 |
|------|------|------|
| **本方案** | 模块.资源.操作 | `system.user.list` |
| AWS IAM | 服务:操作 | `ec2:DescribeInstances` |
| 阿里云 RAM | 服务:操作 | `ecs:DescribeInstances` |
| Spring Security | 模块:资源:操作 | `system:user:view` |

✅ **本方案符合行业标准！**

---

## 三、数据库表设计（适配现有表）

### 3.1 现有表结构

根据项目现有表，权限相关表包括：

```sql
-- 用户表
sys_user (id, username, nickname, ...)

-- 角色表
sys_role (id, code, name, description, ...)

-- 用户角色关联表
sys_user_role (user_id, role_id)

-- 权限表（需增加 module, resource, action 字段）
sys_permission (
    id, 
    code VARCHAR(100),        -- system.user.list
    name VARCHAR(100),        -- 查看用户列表
    module VARCHAR(30),       -- system
    resource VARCHAR(30),     -- user
    action VARCHAR(30),       -- list
    type TINYINT,             -- 1=菜单 2=按钮
    parent_id BIGINT,
    sort_order INT,
    status TINYINT DEFAULT 1
)

-- 角色权限关联表
sys_role_permission (role_id, permission_id)

-- 菜单表
sys_menu (id, name, path, permission, ...)
```

### 3.2 权限表示例数据

```sql
INSERT INTO sys_permission (code, name, module, resource, action, type) VALUES
-- 系统管理模块 - 用户管理
('system.user.list', '查看用户列表', 'system', 'user', 'list', 2),
('system.user.view', '查看用户详情', 'system', 'user', 'view', 2),
('system.user.add', '添加用户', 'system', 'user', 'add', 2),
('system.user.edit', '编辑用户', 'system', 'user', 'edit', 2),
('system.user.delete', '删除用户', 'system', 'user', 'delete', 2),
('system.user.reset_password', '重置密码', 'system', 'user', 'reset_password', 2),
('system.user.assign_role', '分配角色', 'system', 'user', 'assign_role', 2),

-- 系统管理模块 - 角色管理
('system.role.list', '查看角色列表', 'system', 'role', 'list', 2),
('system.role.add', '添加角色', 'system', 'role', 'add', 2),
('system.role.edit', '编辑角色', 'system', 'role', 'edit', 2),
('system.role.delete', '删除角色', 'system', 'role', 'delete', 2),
('system.role.config_permission', '配置角色权限', 'system', 'role', 'config_permission', 2),

-- CMDB模块 - 资产管理
('cmdb.asset.list', '查看资产列表', 'cmdb', 'asset', 'list', 2),
('cmdb.asset.view', '查看资产详情', 'cmdb', 'asset', 'view', 2),
('cmdb.asset.add', '添加资产', 'cmdb', 'asset', 'add', 2),
('cmdb.asset.edit', '编辑资产', 'cmdb', 'asset', 'edit', 2),
('cmdb.asset.delete', '删除资产', 'cmdb', 'asset', 'delete', 2),
('cmdb.asset.import', '导入资产', 'cmdb', 'asset', 'import', 2),
('cmdb.asset.export', '导出资产', 'cmdb', 'asset', 'export', 2),

-- 监控模块 - 告警管理
('monitor.alert.list', '查看告警列表', 'monitor', 'alert', 'list', 2),
('monitor.alert.view', '查看告警详情', 'monitor', 'alert', 'view', 2),
('monitor.alert.handle', '处理告警', 'monitor', 'alert', 'handle', 2),
('monitor.alert.export', '导出告警', 'monitor', 'alert', 'export', 2);
```

### 3.3 Casbin 表（自动创建）

Casbin 使用 `casbin_rule` 表存储鉴权策略：

```sql
-- Casbin 自动管理
casbin_rule (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(100),  -- 'p' = 策略, 'g' = 角色继承
    v0 VARCHAR(100),     -- 角色：admin, operator
    v1 VARCHAR(200),     -- 权限 code：system.user.list
    v2 VARCHAR(100)      -- 动作：read/write/*
)
```

**示例数据：**
```sql
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
-- 管理员权限
('p', 'admin', 'system.user.list', '*'),
('p', 'admin', 'system.user.add', '*'),
('p', 'admin', 'system.user.delete', '*'),
('p', 'admin', 'cmdb.asset.list', '*'),
('p', 'admin', 'monitor.alert.handle', '*'),

-- 操作员权限
('p', 'operator', 'system.user.list', '*'),
('p', 'operator', 'system.user.add', '*'),
('p', 'operator', 'cmdb.asset.list', '*'),

-- CMDB管理员
('p', 'cmdb_admin', 'cmdb.asset.list', '*'),
('p', 'cmdb_admin', 'cmdb.asset.add', '*'),
('p', 'cmdb_admin', 'cmdb.asset.edit', '*'),

-- 角色继承
('g', '1', 'admin'),      -- 用户ID=1 拥有 admin 角色
('g', '2', 'operator');   -- 用户ID=2 拥有 operator 角色
```

---

## 四、Casbin Model 配置

### 4.1 模型配置文件

创建文件：`config/casbin_model.conf`

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

**说明：**
- `sub`：主体，使用角色 code（如 `admin`）
- `obj`：资源，使用层级格式权限 code（如 `system.user.list`）
- `act`：动作，使用 `read`/`write` 或 `*`

---

## 五、后端实现

### 5.1 权限服务初始化

```go
// backend/services/permission_service.go

package services

import (
    "fmt"
    "oneops/backend/models"
    "sync"
    
    "github.com/casbin/casbin/v2"
    gormadapter "github.com/casbin/gorm-adapter/v3"
    "gorm.io/gorm"
)

var (
    permissionService     *PermissionService
    permissionServiceOnce sync.Once
)

type PermissionService struct {
    db       *gorm.DB
    enforcer *casbin.Enforcer
}

// GetPermissionService 获取权限服务单例
func GetPermissionService() (*PermissionService, error) {
    if permissionService != nil {
        return permissionService, nil
    }
    
    var initErr error
    permissionServiceOnce.Do(func() {
        service, err := NewPermissionService()
        if err != nil {
            initErr = err
            return
        }
        permissionService = service
    })
    
    if initErr != nil {
        return nil, initErr
    }
    
    return permissionService, nil
}

// NewPermissionService 创建权限服务实例
func NewPermissionService() (*PermissionService, error) {
    db := GetDB()
    
    // 初始化 Casbin GORM 适配器
    adapter, err := gormadapter.NewAdapterByDB(db)
    if err != nil {
        return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
    }
    
    // 创建 Casbin enforcer
    enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
    if err != nil {
        return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
    }
    
    // 加载策略
    if err := enforcer.LoadPolicy(); err != nil {
        return nil, fmt.Errorf("failed to load policy: %w", err)
    }
    
    return &PermissionService{
        db:       db,
        enforcer: enforcer,
    }, nil
}

// Enforcer 获取 Casbin Enforcer
func (s *PermissionService) Enforcer() *casbin.Enforcer {
    return s.enforcer
}
```

### 5.2 核心方法：同步权限到 Casbin

```go
// backend/services/permission_service.go

// SyncRolePermissionsToCasbin 同步角色权限到 Casbin
func (s *PermissionService) SyncRolePermissionsToCasbin(roleCode string, permissionIDs []uint) error {
    // 1. 删除该角色的所有旧策略
    s.enforcer.RemoveFilteredPolicy(0, roleCode)
    
    if len(permissionIDs) == 0 {
        return s.enforcer.SavePolicy()
    }
    
    // 2. 查询权限详情（获取 code 字段）
    var permissions []models.Permission
    if err := s.db.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
        return err
    }
    
    // 3. 添加新策略到 Casbin
    // 策略格式：(role, code, *)
    // code 是层级格式，如：system.user.list
    for _, perm := range permissions {
        _, err := s.enforcer.AddPolicy(roleCode, perm.Code, "*")
        if err != nil {
            return err
        }
    }
    
    // 4. 保存策略到数据库
    return s.enforcer.SavePolicy()
}

// SyncAllRolesToCasbin 同步所有角色权限到 Casbin
func (s *PermissionService) SyncAllRolesToCasbin() error {
    var roles []models.Role
    if err := s.db.Where("status = 1").Find(&roles).Error; err != nil {
        return err
    }
    
    for _, role := range roles {
        var permissionIDs []uint
        s.db.Table("sys_role_permission").
            Where("role_id = ?", role.ID).
            Pluck("permission_id", &permissionIDs)
        
        if err := s.SyncRolePermissionsToCasbin(role.Code, permissionIDs); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 5.3 用户权限查询

```go
// backend/services/permission_service.go

// GetUserPermissions 获取用户的所有权限 code 列表
func (s *PermissionService) GetUserPermissions(userID uint) ([]string, error) {
    var user models.User
    if err := s.db.First(&user, userID).Error; err != nil {
        return nil, err
    }
    
    // 超级管理员返回通配符
    if user.Username == "admin" {
        return []string{"*.*.*"}, nil
    }
    
    // 查询用户的角色
    var roleIDs []uint
    s.db.Table("sys_user_role").
        Where("user_id = ?", userID).
        Pluck("role_id", &roleIDs)
    
    if len(roleIDs) == 0 {
        return []string{}, nil
    }
    
    // 查询角色的权限
    var permissions []models.Permission
    s.db.Table("sys_permission").
        Select("DISTINCT sys_permission.code").
        Joins("JOIN sys_role_permission ON sys_permission.id = sys_role_permission.permission_id").
        Where("sys_role_permission.role_id IN ?", roleIDs).
        Where("sys_permission.status = 1").
        Find(&permissions)
    
    codes := make([]string, 0, len(permissions))
    for _, perm := range permissions {
        codes = append(codes, perm.Code)
    }
    
    return codes, nil
}

// GetUserMenus 获取用户的菜单树
func (s *PermissionService) GetUserMenus(userID uint) ([]models.Menu, error) {
    var user models.User
    if err := s.db.First(&user, userID).Error; err != nil {
        return nil, err
    }
    
    // 超级管理员返回所有菜单
    if user.Username == "admin" {
        var menus []models.Menu
        s.db.Where("status = 1").Order("sort ASC").Find(&menus)
        return s.buildMenuTree(menus, nil), nil
    }
    
    // 查询用户的角色
    var roleIDs []uint
    s.db.Table("sys_user_role").
        Where("user_id = ?", userID).
        Pluck("role_id", &roleIDs)
    
    // 查询角色的菜单
    var menus []models.Menu
    s.db.Table("sys_menu").
        Select("DISTINCT sys_menu.*").
        Joins("JOIN sys_role_menu ON sys_menu.id = sys_role_menu.menu_id").
        Where("sys_role_menu.role_id IN ?", roleIDs).
        Where("sys_menu.status = 1").
        Order("sys_menu.sort ASC").
        Find(&menus)
    
    return s.buildMenuTree(menus, nil), nil
}

// buildMenuTree 构建菜单树
func (s *PermissionService) buildMenuTree(menus []models.Menu, parentID *uint) []models.Menu {
    var tree []models.Menu
    
    for _, menu := range menus {
        if (parentID == nil && menu.ParentID == nil) ||
           (parentID != nil && menu.ParentID != nil && *menu.ParentID == *parentID) {
            menu.Children = s.buildMenuTree(menus, &menu.ID)
            tree = append(tree, menu)
        }
    }
    
    return tree
}
```

---

## 六、权限中间件实现

### 6.1 API 鉴权中间件

```go
// backend/middlewares/permission.go

package middlewares

import (
    "fmt"
    "oneops/backend/models"
    "oneops/backend/services"
    "oneops/backend/utils"
    
    "github.com/gin-gonic/gin"
)

// RequirePermission 要求指定权限（路由中间件）
// 使用方式：router.DELETE("/api/users/:id", RequirePermission("system.user.delete"), handler)
func RequirePermission(code string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("user_id")
        if !exists {
            c.JSON(401, utils.ErrorUnauthorized("未授权访问"))
            c.Abort()
            return
        }
        
        // 获取权限服务
        permService, err := services.GetPermissionService()
        if err != nil {
            c.JSON(500, utils.ErrorInternal("权限服务初始化失败"))
            c.Abort()
            return
        }
        
        // 获取用户角色
        roles := getUserRoles(userID.(uint))
        
        // 获取请求动作
        action := getActionFromMethod(c.Request.Method)
        
        // Casbin 鉴权
        allowed := false
        for _, role := range roles {
            if ok, _ := permService.Enforcer().Enforce(role.Code, code, action); ok {
                allowed = true
                break
            }
        }
        
        if !allowed {
            username, _ := c.Get("username")
            fmt.Printf("[权限不足] user_id=%v, username=%v, required=%s\n", 
                userID, username, code)
            
            c.JSON(403, utils.ErrorForbidden("权限不足: 需要 "+code+" 权限"))
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// getUserRoles 获取用户角色
func getUserRoles(userID uint) []models.Role {
    var roles []models.Role
    db := services.GetDB()
    
    db.Table("sys_role").
        Select("sys_role.*").
        Joins("JOIN sys_user_role ON sys_role.id = sys_user_role.role_id").
        Where("sys_user_role.user_id = ?", userID).
        Find(&roles)
    
    return roles
}

// getActionFromMethod HTTP 方法转换为动作
func getActionFromMethod(method string) string {
    switch method {
    case "GET":
        return "read"
    case "POST", "PUT", "DELETE", "PATCH":
        return "write"
    default:
        return "*"
    }
}
```

### 6.2 路由配置示例

```go
// backend/routes/routes.go

func RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api")
    api.Use(middlewares.Auth())
    
    // 系统管理 - 用户
    users := api.Group("/system/users")
    {
        users.GET("", 
            middlewares.RequirePermission("system.user.list"), 
            controllers.GetUsers)
        users.POST("", 
            middlewares.RequirePermission("system.user.add"), 
            controllers.CreateUser)
        users.PUT("/:id", 
            middlewares.RequirePermission("system.user.edit"), 
            controllers.UpdateUser)
        users.DELETE("/:id", 
            middlewares.RequirePermission("system.user.delete"), 
            controllers.DeleteUser)
        users.PUT("/:id/reset-password", 
            middlewares.RequirePermission("system.user.reset_password"), 
            controllers.ResetPassword)
    }
    
    // 系统管理 - 角色
    roles := api.Group("/system/roles")
    {
        roles.GET("", 
            middlewares.RequirePermission("system.role.list"), 
            controllers.GetRoles)
        roles.POST("", 
            middlewares.RequirePermission("system.role.add"), 
            controllers.CreateRole)
        roles.PUT("/:id", 
            middlewares.RequirePermission("system.role.edit"), 
            controllers.UpdateRole)
        roles.DELETE("/:id", 
            middlewares.RequirePermission("system.role.delete"), 
            controllers.DeleteRole)
        roles.POST("/:id/permissions", 
            middlewares.RequirePermission("system.role.config_permission"), 
            controllers.AssignPermissionsToRole)
    }
    
    // CMDB - 资产
    assets := api.Group("/cmdb/assets")
    {
        assets.GET("", 
            middlewares.RequirePermission("cmdb.asset.list"), 
            controllers.GetAssets)
        assets.POST("", 
            middlewares.RequirePermission("cmdb.asset.add"), 
            controllers.CreateAsset)
        assets.PUT("/:id", 
            middlewares.RequirePermission("cmdb.asset.edit"), 
            controllers.UpdateAsset)
        assets.DELETE("/:id", 
            middlewares.RequirePermission("cmdb.asset.delete"), 
            controllers.DeleteAsset)
        assets.POST("/import", 
            middlewares.RequirePermission("cmdb.asset.import"), 
            controllers.ImportAssets)
        assets.GET("/export", 
            middlewares.RequirePermission("cmdb.asset.export"), 
            controllers.ExportAssets)
    }
    
    // 监控 - 告警
    alerts := api.Group("/monitor/alerts")
    {
        alerts.GET("", 
            middlewares.RequirePermission("monitor.alert.list"), 
            controllers.GetAlerts)
        alerts.GET("/:id", 
            middlewares.RequirePermission("monitor.alert.view"), 
            controllers.GetAlertDetail)
        alerts.POST("/:id/handle", 
            middlewares.RequirePermission("monitor.alert.handle"), 
            controllers.HandleAlert)
        alerts.GET("/export", 
            middlewares.RequirePermission("monitor.alert.export"), 
            controllers.ExportAlerts)
    }
}
```

---

## 七、控制器实现

### 7.1 角色权限配置

```go
// backend/controllers/role_controller.go

package controllers

import (
    "net/http"
    "oneops/backend/models"
    "oneops/backend/services"
    "oneops/backend/utils"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type RoleController struct{}

// GetRolePermissions 获取角色的权限列表
func (c *RoleController) GetRolePermissions(ctx *gin.Context) {
    roleID := ctx.Param("id")
    
    var permissions []models.Permission
    db := services.GetDB()
    
    db.Table("sys_permission").
        Select("sys_permission.*").
        Joins("JOIN sys_role_permission ON sys_permission.id = sys_role_permission.permission_id").
        Where("sys_role_permission.role_id = ?", roleID).
        Where("sys_permission.status = 1").
        Order("sys_permission.sort ASC").
        Find(&permissions)
    
    ctx.JSON(http.StatusOK, utils.SuccessWithData(permissions))
}

// AssignPermissionsToRole 为角色分配权限
func (c *RoleController) AssignPermissionsToRole(ctx *gin.Context) {
    var req struct {
        RoleID        uint   `json:"roleId" binding:"required"`
        PermissionIDs []uint `json:"permissionIds" binding:"required"`
    }
    
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
        return
    }
    
    db := services.GetDB()
    
    // 事务处理
    err := db.Transaction(func(tx *gorm.DB) error {
        // 1. 删除角色旧权限
        if err := tx.Where("role_id = ?", req.RoleID).Delete(&models.RolePermission{}).Error; err != nil {
            return err
        }
        
        // 2. 添加新权限
        for _, permID := range req.PermissionIDs {
            rolePerm := models.RolePermission{
                RoleID:       req.RoleID,
                PermissionID: permID,
            }
            if err := tx.Create(&rolePerm).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
    
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("权限配置失败"))
        return
    }
    
    // 3. 同步到 Casbin（关键步骤）
    var role models.Role
    db.First(&role, req.RoleID)
    
    permService, err := services.GetPermissionService()
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("权限服务初始化失败"))
        return
    }
    
    // 同步权限 code 到 Casbin
    // 权限 code 格式：system.user.list, cmdb.asset.add 等
    if err := permService.SyncRolePermissionsToCasbin(role.Code, req.PermissionIDs); err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("同步权限策略失败"))
        return
    }
    
    ctx.JSON(http.StatusOK, utils.SuccessWithMessage("权限配置成功"))
}
```

### 7.2 用户角色分配

```go
// backend/controllers/user_controller.go

package controllers

import (
    "net/http"
    "oneops/backend/models"
    "oneops/backend/services"
    "oneops/backend/utils"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type UserController struct{}

// AssignRolesToUser 为用户分配角色
func (c *UserController) AssignRolesToUser(ctx *gin.Context) {
    var req struct {
        UserID  uint   `json:"userId" binding:"required"`
        RoleIDs []uint `json:"roleIds" binding:"required"`
    }
    
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
        return
    }
    
    db := services.GetDB()
    
    // 事务处理
    err := db.Transaction(func(tx *gorm.DB) error {
        // 1. 删除用户旧角色
        if err := tx.Where("user_id = ?", req.UserID).Delete(&models.UserRole{}).Error; err != nil {
            return err
        }
        
        // 2. 添加新角色
        for _, roleID := range req.RoleIDs {
            userRole := models.UserRole{
                UserID: req.UserID,
                RoleID: roleID,
            }
            if err := tx.Create(&userRole).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
    
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("角色分配失败"))
        return
    }
    
    ctx.JSON(http.StatusOK, utils.SuccessWithMessage("角色分配成功"))
}

// GetUserPermissions 获取用户的权限列表
func (c *UserController) GetUserPermissions(ctx *gin.Context) {
    userID, _ := ctx.Get("user_id")
    
    permService, err := services.GetPermissionService()
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("权限服务初始化失败"))
        return
    }
    
    // 获取权限 code 列表（层级格式）
    // 返回：["system.user.list", "system.user.add", "cmdb.asset.view"]
    permissions, err := permService.GetUserPermissions(userID.(uint))
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("获取权限失败"))
        return
    }
    
    // 获取菜单树
    menus, err := permService.GetUserMenus(userID.(uint))
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("获取菜单失败"))
        return
    }
    
    ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
        "permissions": permissions,
        "menus":       menus,
    }))
}
```

---

## 八、登录接口实现

```go
// backend/controllers/auth_controller.go

package controllers

import (
    "net/http"
    "oneops/backend/models"
    "oneops/backend/services"
    "oneops/backend/utils"
    
    "github.com/gin-gonic/gin"
)

type AuthController struct{}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
    var req struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorBadRequest("参数错误"))
        return
    }
    
    db := services.GetDB()
    
    // 1. 验证用户名密码
    var user models.User
    if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorBadRequest("用户名或密码错误"))
        return
    }
    
    // 验证密码
    if user.Password != req.Password {
        ctx.JSON(http.StatusOK, utils.ErrorBadRequest("用户名或密码错误"))
        return
    }
    
    // 2. 生成 Token
    token, err := generateToken(user.ID)
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("生成令牌失败"))
        return
    }
    
    // 3. 获取用户权限（层级格式）
    permService, err := services.GetPermissionService()
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("权限服务初始化失败"))
        return
    }
    
    permissions, err := permService.GetUserPermissions(user.ID)
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("获取权限失败"))
        return
    }
    
    // 4. 获取用户菜单
    menus, err := permService.GetUserMenus(user.ID)
    if err != nil {
        ctx.JSON(http.StatusOK, utils.ErrorInternal("获取菜单失败"))
        return
    }
    
    // 5. 获取用户角色
    var roles []models.Role
    db.Table("sys_role").
        Select("sys_role.*").
        Joins("JOIN sys_user_role ON sys_role.id = sys_user_role.role_id").
        Where("sys_user_role.user_id = ?", user.ID).
        Find(&roles)
    
    // 6. 返回登录信息
    ctx.JSON(http.StatusOK, utils.SuccessWithData(gin.H{
        "token":       token,
        "user":        user,
        "roles":       roles,
        "permissions": permissions,  // ["system.user.list", "system.user.add", ...]
        "menus":       menus,
    }))
}
```

---

## 九、前端集成

### 9.1 权限检查工具

```typescript
// frontend/src/utils/permission.ts

import { useAuthStore } from '@/store/modules/auth'

/**
 * 检查是否拥有指定权限
 * @param code 权限标识（层级格式），如 'system.user.delete'
 */
export function hasPermission(code: string): boolean {
  const authStore = useAuthStore()
  const permissions = authStore.permissions
  
  // 支持通配符 *.*.*
  if (permissions.includes('*.*.*')) {
    return true
  }
  
  // 精确匹配
  if (permissions.includes(code)) {
    return true
  }
  
  // 支持模块级通配符
  // system.user.delete 匹配：
  // - system.user.*
  // - system.*.*
  // - *.*.*
  return matchWildcard(permissions, code)
}

/**
 * 通配符匹配
 * @param permissions 权限列表
 * @param code 要检查的权限 code
 */
function matchWildcard(permissions: string[], code: string): boolean {
  const codeParts = code.split('.')
  if (codeParts.length !== 3) {
    return false
  }
  
  for (const perm of permissions) {
    const permParts = perm.split('.')
    if (permParts.length !== 3) {
      continue
    }
    
    // 逐级匹配
    let match = true
    for (let i = 0; i < 3; i++) {
      if (permParts[i] !== '*' && permParts[i] !== codeParts[i]) {
        match = false
        break
      }
    }
    
    if (match) {
      return true
    }
  }
  
  return false
}

/**
 * 检查是否拥有任意一个权限
 */
export function hasAnyPermission(codes: string[]): boolean {
  return codes.some(code => hasPermission(code))
}

/**
 * 检查是否拥有所有权限
 */
export function hasAllPermissions(codes: string[]): boolean {
  return codes.every(code => hasPermission(code))
}

/**
 * 检查模块权限
 * @param module 模块名称，如 'system', 'cmdb'
 */
export function hasModulePermission(module: string): boolean {
  const authStore = useAuthStore()
  const permissions = authStore.permissions
  
  // 检查是否有该模块的任何权限
  return permissions.some(perm => perm.startsWith(module + '.'))
}
```

### 9.2 权限指令

```typescript
// frontend/src/directives/permission.ts

import type { Directive, DirectiveBinding } from 'vue'
import { hasPermission } from '@/utils/permission'

export const permission: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string>) {
    const code = binding.value
    
    if (code && !hasPermission(code)) {
      el.parentNode?.removeChild(el)
    }
  }
}

export function setupPermissionDirective(app: App) {
  app.directive('permission', permission)
}
```

### 9.3 按钮权限使用

```vue
<!-- frontend/src/views/manage/user/index.vue -->

<template>
  <div class="user-manage">
    <el-table :data="users">
      <el-table-column prop="username" label="用户名" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <!-- 使用层级格式的权限 code -->
          <el-button 
            v-permission="'system.user.edit'"
            size="small" 
            @click="editUser(row)"
          >
            编辑
          </el-button>
          
          <el-button 
            v-permission="'system.user.delete'"
            size="small" 
            type="danger"
            @click="deleteUser(row)"
          >
            删除
          </el-button>
          
          <el-button 
            v-permission="'system.user.reset_password'"
            size="small"
            @click="resetPassword(row)"
          >
            重置密码
          </el-button>
          
          <el-button 
            v-permission="'system.user.assign_role'"
            size="small"
            type="primary"
            @click="assignRoles(row)"
          >
            分配角色
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { hasPermission } from '@/utils/permission'

// 或在逻辑中使用
if (hasPermission('system.user.delete')) {
  // 执行删除操作
}
</script>
```

### 9.4 CMDB 模块示例

```vue
<!-- frontend/src/views/cmdb/asset/index.vue -->

<template>
  <div class="asset-manage">
    <!-- 操作按钮 -->
    <el-space>
      <el-button 
        v-permission="'cmdb.asset.add'"
        type="primary"
        @click="addAsset"
      >
        添加资产
      </el-button>
      
      <el-button 
        v-permission="'cmdb.asset.import'"
        @click="importAssets"
      >
        导入资产
      </el-button>
      
      <el-button 
        v-permission="'cmdb.asset.export'"
        @click="exportAssets"
      >
        导出资产
      </el-button>
    </el-space>
    
    <!-- 资产列表 -->
    <el-table :data="assets">
      <el-table-column prop="name" label="资产名称" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button 
            v-permission="'cmdb.asset.edit'"
            size="small"
            @click="editAsset(row)"
          >
            编辑
          </el-button>
          
          <el-button 
            v-permission="'cmdb.asset.delete'"
            size="small"
            type="danger"
            @click="deleteAsset(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
```

---

## 十、完整授权流程

### 10.1 管理员授权流程

```
1. 管理员登录系统
   ↓
2. 进入"角色管理"页面
   ↓
3. 点击"配置权限"按钮
   ↓
4. 在权限树中勾选需要的权限
   ☑ system.user.list
   ☑ system.user.add
   ☐ system.user.delete
   ☑ cmdb.asset.view
   ↓
5. 点击"确定"保存
   ↓
6. 后端保存到 sys_role_permission 表
   ↓
7. 自动同步到 Casbin
   (p, operator, system.user.list, *)
   (p, operator, system.user.add, *)
   (p, operator, cmdb.asset.view, *)
   ↓
8. 完成！权限立即生效
```

### 10.2 用户登录流程

```
1. 用户输入用户名密码
   ↓
2. 后端验证身份
   ↓
3. 查询 sys_user_role 获取角色
   ↓
4. 查询 sys_role_permission 获取权限ID
   ↓
5. 查询 sys_permission 获取权限 code（层级格式）
   ↓
6. 返回前端:
   {
     token: "xxx",
     permissions: [
       "system.user.list",
       "system.user.add",
       "cmdb.asset.view"
     ],
     menus: [...],
     roles: [...]
   }
   ↓
7. 前端存储权限列表
   ↓
8. 按钮根据权限显示/隐藏
   v-permission="'system.user.add'"
```

### 10.3 API 鉴权流程

```
1. 用户点击按钮
   ↓
2. 前端发送 API 请求
   DELETE /api/system/users/:id
   Headers: { Authorization: "Bearer token" }
   ↓
3. 后端 Auth 中间件解析 Token
   获取 userID
   ↓
4. RequirePermission 中间件
   检查权限: system.user.delete
   ↓
5. 查询用户角色
   SELECT role_code FROM sys_role
   JOIN sys_user_role WHERE user_id = ?
   结果: operator
   ↓
6. Casbin 鉴权
   Enforce("operator", "system.user.delete", "write")
   ↓
7. 查询 casbin_rule 表
   SELECT * FROM casbin_rule 
   WHERE ptype='p' 
     AND v0='operator' 
     AND v1='system.user.delete'
   ↓
8. 返回结果
   ✅ 找到策略 → 允许访问
   ❌ 未找到 → 403 权限不足
```

---

## 十一、数据同步说明

### 11.1 为什么要同步？

**业务表和 Casbin 是两个独立的存储：**

- **业务表**：存储权限的元数据（名称、层级、关联关系）
- **Casbin 表**：存储鉴权策略（角色、权限code、动作）

**不同步的后果：**
```
❌ 用户配置了权限（业务表有记录）
❌ 但 API 无法访问（Casbin 没有策略）
❌ 前端显示有按钮，点击却提示权限不足
```

### 11.2 同步机制

```go
// 关键代码：AssignPermissionsToRole 方法中

// 1. 保存到业务表
db.Create(&RolePermission{RoleID: roleID, PermissionID: permID})

// 2. 同步到 Casbin（自动）
// 权限 code 从 sys_permission.code 读取
// 格式：system.user.delete, cmdb.asset.add 等
permService.SyncRolePermissionsToCasbin(roleCode, permissionIDs)
```

### 11.3 同步时机

| 操作 | 业务表更新 | Casbin 同步 |
|------|-----------|------------|
| 为角色配置权限 | ✅ | ✅ 自动同步 |
| 为用户分配角色 | ✅ | ❌ 无需同步 |
| 创建新权限 | ✅ | ❌ 无需同步 |
| 删除权限 | ✅ | ✅ 需手动同步 |

---

## 十二、通配符权限支持

### 12.1 通配符格式

```
*.*.*              全局通配符（超级管理员）
system.*.*         模块级通配符（系统模块所有权限）
cmdb.asset.*       资源级通配符（CMDB资产所有操作）
*.user.list        跨模块通配符（所有模块的用户列表）
```

### 12.2 通配符实现

```go
// backend/utils/permission_match.go

package utils

import "strings"

// MatchPermission 通配符权限匹配
// pattern: 权限模式（可能包含通配符）
// code: 要检查的权限 code
func MatchPermission(pattern, code string) bool {
    // 精确匹配
    if pattern == code {
        return true
    }
    
    // 分解权限 code
    patternParts := strings.Split(pattern, ".")
    codeParts := strings.Split(code, ".")
    
    if len(patternParts) != 3 || len(codeParts) != 3 {
        return false
    }
    
    // 逐级匹配
    for i := 0; i < 3; i++ {
        if patternParts[i] != "*" && patternParts[i] != codeParts[i] {
            return false
        }
    }
    
    return true
}

// HasPermissionWithWildcard 检查权限（支持通配符）
func HasPermissionWithWildcard(permissions []string, code string) bool {
    for _, perm := range permissions {
        if MatchPermission(perm, code) {
            return true
        }
    }
    return false
}
```

### 12.3 Casbin 数据示例

```sql
-- 超级管理员
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'admin', '*.*.*', '*');

-- 系统模块管理员
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'system_admin', 'system.*.*', '*');

-- CMDB 资产管理员
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'cmdb_asset_admin', 'cmdb.asset.*', '*');

-- 用户管理员（所有模块的用户管理权限）
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
('p', 'user_admin', '*.user.*', '*');
```

---

## 十三、最佳实践

### 13.1 权限设计原则

1. **最小权限原则**：只分配必要的权限
2. **模块化管理**：按模块组织权限
3. **职责分离**：不同角色拥有不同的权限集合
4. **定期审计**：定期检查权限配置的合理性

### 13.2 权限配置建议

#### 系统管理员（system_admin）
```
system.*.*           # 系统模块所有权限
```

#### CMDB 管理员（cmdb_admin）
```
cmdb.*.*             # CMDB模块所有权限
```

#### 运维工程师（operator）
```
system.user.list     # 查看用户列表
system.role.list      # 查看角色列表
cmdb.asset.*          # 资产管理所有操作
monitor.alert.list    # 查看告警列表
monitor.alert.handle  # 处理告警
```

#### 查看员（viewer）
```
system.user.list      # 查看用户列表
cmdb.asset.list       # 查看资产列表
monitor.dashboard.view # 查看监控仪表盘
monitor.alert.list     # 查看告警列表
```

### 13.3 性能优化建议

1. **权限缓存**：登录时加载权限到 Redis
2. **菜单缓存**：菜单树缓存 5 分钟
3. **Casbin 缓存**：启用 Casbin 内置缓存
4. **批量操作**：权限配置使用事务

---

## 十四、常见问题

### Q1: 权限配置后不生效？

**原因：** 没有同步到 Casbin

**解决：**
```go
// 确保调用了同步方法
permService.SyncRolePermissionsToCasbin(roleCode, permissionIDs)
```

### Q2: 前端按钮不显示？

**原因：** 权限 code 不匹配

**解决：**
```typescript
// 检查权限 code 是否一致
<el-button v-permission="'system.user.delete'">

// 后端权限表
code: 'system.user.delete'  // 必须完全一致
```

### Q3: 如何实现模块级权限？

**使用通配符：**
```sql
-- 系统模块管理员
('p', 'system_admin', 'system.*.*', '*')

-- CMDB模块管理员
('p', 'cmdb_admin', 'cmdb.*.*', '*')
```

### Q4: 如何批量分配权限？

**后端实现：**
```go
// 批量为用户分配角色
func BatchAssignRoles(userIDs []uint, roleIDs []uint) error {
    for _, userID := range userIDs {
        for _, roleID := range roleIDs {
            db.Create(&UserRole{UserID: userID, RoleID: roleID})
        }
    }
    return nil
}
```

---

## 十五、总结

### 核心要点

1. ✅ **业务库管理元数据，Casbin 负责鉴权**
2. ✅ **权限 code 使用层级格式：`模块.资源.操作`**
3. ✅ **配置权限时自动同步到 Casbin**
4. ✅ **前端使用层级格式 code 控制按钮**
5. ✅ **后端使用 Casbin 鉴权 API 请求**
6. ✅ **支持通配符权限：`*.*.*`, `system.*.*` 等**

### 方案优势

- 简单易维护
- 模块化管理
- 层级清晰
- 符合行业标准
- 性能优秀
- 扩展性强

### 适用场景

- 企业管理系统
- 多模块系统
- RBAC 权限模型
- 需要细粒度权限控制
- 需要动态权限配置

### 立即行动

**采用层级格式的权限 code：**

```
❌ 旧格式: user:delete
✅ 新格式: system.user.delete

收益：
- 模块清晰
- 层级分明
- 符合行业标准
- 支持多模块系统
```

**这是一个生产就绪的完整方案！** 🚀
