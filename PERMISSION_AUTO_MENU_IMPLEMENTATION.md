# 权限码自动关联菜单功能 - 实现完成

## 已完成的修改

### 1. 数据库模型修改

**文件：** `backend/models/menu.go`

添加了 `Resource` 字段，用于关联权限码的资源：

```go
type Menu struct {
    ID         uint
    Name       string
    Icon       string
    Path       string
    Permission string
    Resource   string    // 新增：对应的资源名称，用于权限码自动关联
    MenuType   string
    ParentID   uint
    Sort       int
    Status     int
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Children   []*Menu
}
```

### 2. 数据库迁移

**文件：** `backend/services/init.go`

添加了迁移脚本：
- 添加 `resource` 字段到 `menus` 表
- 创建索引 `idx_menus_resource`
- 自动更新现有菜单的 resource 字段

### 3. 权限推导逻辑

**文件：** `backend/services/rbac.go`

在 `BuildMenuTreeAndPermissions` 方法中添加了权限码到菜单的自动推导：

```go
// 从权限码自动推导菜单权限
if !isAdmin {
    // 1. 获取用户的所有权限码
    var userPermissions []string
    for _, role := range roles {
        var rolePerms []models.RolePermission
        db.Where("role_id = ?", role.ID).Preload("Permission").Find(&rolePerms)
        for _, rp := range rolePerms {
            if rp.Permission.Code != "" {
                userPermissions = append(userPermissions, rp.Permission.Code)
            }
        }
    }

    // 2. 从权限码中提取有权限的资源列表
    // 权限码格式: module:resource:action
    allowedResources := make(map[string]bool)
    for _, permCode := range userPermissions {
        parts := strings.Split(permCode, ":")
        if len(parts) >= 2 {
            resource := parts[1]
            allowedResources[resource] = true
        }
    }

    // 3. 查找 resource 匹配的菜单
    for _, menu := range allMenus {
        if menu.Resource != "" && allowedResources[menu.Resource] {
            menuIDs[menu.ID] = true

            // 同时标记父菜单（目录）
            for _, m := range allMenus {
                if m.ID == menu.ParentID {
                    menuIDs[m.ID] = true
                }
            }
        }
    }
}
```

### 4. 修复了 syncRoleMenus 覆盖问题

**文件：** `backend/services/init.go`

修改后不再覆盖用户配置的 `menu_ids`：

```go
if err == nil {
    // 角色已存在：只更新名称和描述，不覆盖用户配置的权限
    db.Model(&existingRole).Updates(map[string]interface{}{
        "name":        builtinRole.name,
        "description": builtinRole.description,
        // 注意：不更新 menu_ids，保留用户通过界面配置的权限
    })
}
```

## 工作原理

### 权限码格式

```
module:resource:action
  ↓       ↓       ↓
系统    用户    查看

示例：
- system:user:view     → resource = "user"
- system:role:create   → resource = "role"
- cmdb:server:query    → resource = "server"
```

### 菜单 resource 映射

```
菜单表:
- id: 70, path: /manage/user, resource: "user"
- id: 71, path: /manage/role, resource: "role"
- id: 20, path: /cmdb/servers, resource: "server"

权限码表:
- system:user:view  → resource = "user" → 匹配菜单 id=70
- system:role:view  → resource = "role" → 匹配菜单 id=71
```

### 自动推导流程

```
1. 用户分配权限码: system:user:view
   ↓
2. 提取 resource: "user"
   ↓
3. 查找菜单: resource = "user" → 找到 id=70 (用户管理)
   ↓
4. 自动显示菜单: 系统管理 → 用户管理
   ↓
5. 同时显示父菜单: 系统管理 (id=6)
```

## 兼容性

### 两套机制并存

当前实现保留了两种方式：

1. **原有方式：** 通过 `roles.menu_ids` 配置菜单权限
2. **新增方式：** 通过权限码自动推导菜单权限

**逻辑：** 两种方式的菜单权限会合并，用户能访问的菜单 = menu_ids + 权限码推导的菜单

### 未来计划

建议逐步废弃 `menu_ids` 机制，完全使用权限码控制。

## 测试步骤

### 1. 启动后端

```bash
cd backend
go run main.go
```

### 2. 查看日志

启动时会看到：

```
开始执行 SQL 迁移...
添加 resource 字段（可能已存在）
开始更新菜单resource字段...
更新菜单resource字段 path=/manage/user resource=user
更新菜单resource字段 path=/manage/role resource=role
...
菜单resource字段更新完成
```

### 3. 准备权限数据

检查 `permissions` 表中是否有权限码：

```sql
SELECT code, module, resource, action, level
FROM permissions
WHERE module = 'system'
ORDER BY resource, action;
```

预期结果：

```
system:user:view     | system | user    | view     | 3
system:user:create   | system | user    | create   | 3
system:user:update   | system | user    | update   | 3
system:user:delete   | system | user    | delete   | 3
system:role:view     | system | role    | view     | 3
...
```

### 4. 为 ops 角色分配权限码

```sql
-- 假设 ops 角色的 id = 2
-- 分配用户管理和角色管理的查看权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions
WHERE code IN ('system:user:view', 'system:role:view');
```

### 5. 清除缓存

```bash
curl -X POST http://127.0.0.1:9527/proxy-default/route/invalidate-cache \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 6. 测试用户登录

使用 test 用户（密码 123456）登录，查看浏览器控制台：

**预期日志：**

```javascript
[登录调试-RBAC] 开始从权限码推导菜单
[登录调试-RBAC] 用户权限码 数量=2 ["system:user:view", "system:role:view"]
[登录调试-RBAC] 用户有权限的资源 数量=2 ["user", "role"]
[登录调试-RBAC] 从权限码推导的菜单数量 数量=3  // 用户管理(70) + 角色管理(71) + 系统管理(6)

🐛 [路由Store] 后端返回的路由数据: [
  {
    "id": "1",
    "path": "/home",
    "name": "home"
  },
  {
    "id": "6",
    "path": "/manage",
    "name": "manage",
    "children": [
      { "id": "70", "path": "/manage/user", "name": "manage_user" },
      { "id": "71", "path": "/manage/role", "name": "manage_role" }
    ]
  }
]
```

**预期菜单：**

- ✅ 首页
- ✅ 系统管理
  - ✅ 用户管理
  - ✅ 角色管理

**不应该出现的菜单：**
- ❌ 资产管理
- ❌ 监控中心
- ❌ K8s管理
- ❌ 审计中心
- ❌ 授权中心

## 验证要点

### 1. 权限码匹配

检查 `permissions.resource` 和 `menus.resource` 是否一致：

```sql
-- 检查权限码的 resource
SELECT code, resource FROM permissions WHERE module='system';

-- 检查菜单的 resource
SELECT id, name, path, resource FROM menus WHERE parent_id=6;

-- 应该匹配：
-- permissions: system:user:* → resource = "user"
-- menus: 用户管理 → resource = "user"
```

### 2. 父菜单自动显示

如果用户有子菜单的权限，父菜单会自动显示：

```
有权限码: system:user:view
    ↓
匹配菜单: 用户管理 (id=70, parent_id=6)
    ↓
自动显示: 系统管理 (id=6)
```

### 3. 管理员权限

admin 角色会忽略权限码推导，直接显示所有菜单：

```go
if isAdmin {
    menuIDs = make(map[uint]bool)
    for _, menu := range allMenus {
        menuIDs[menu.ID] = true
    }
}
```

## 下一步

### 可选优化

1. **完全废弃 menu_ids**
   - 删除 `roles.menu_ids` 字段
   - 删除相关代码

2. **优化权限分配界面**
   - 在 `permission-assign-modal.vue` 中显示对应的菜单路径
   - 添加提示："分配此权限将自动显示 XXX 菜单"

3. **支持通配符权限**
   - `system:*:view` → 显示 system 模块所有页面
   - `*:*:*` → 显示所有菜单

### 前端调整（可选）

可以移除 `menu-auth-modal.vue` 组件，因为现在不再需要单独分配菜单权限。

## 总结

✅ 实现了权限码自动关联菜单的功能
✅ 解决了 syncRoleMenus 覆盖用户配置的问题
✅ 两套机制并存，保证兼容性
✅ 支持父菜单自动显示
✅ 管理员权限保持不变

现在用户只需要通过前端"分配权限"功能分配权限码，对应的菜单就会自动显示，不再需要单独配置菜单权限。
