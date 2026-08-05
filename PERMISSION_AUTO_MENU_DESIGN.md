# 权限系统改进方案：权限码自动关联菜单

## 当前问题

系统使用两套分离的权限系统：

### 系统1：菜单权限（当前使用）
- **存储：** `roles.menu_ids` 字段
- **逻辑：** 直接指定角色能访问的菜单ID
- **问题：**
  - ❌ 需要单独配置菜单权限
  - ❌ 与功能权限分离，容易不一致
  - ❌ 前端缺少配置界面（menu-auth-modal.vue 未使用）

### 系统2：功能权限（已实现）
- **存储：** `permissions` 表 + `role_permissions` 关联表
- **逻辑：** 细粒度的操作权限（按钮级别）
- **字段：**
  - `module`: 模块（system, cmdb, monitoring）
  - `resource`: 资源（user, role, server）
  - `action`: 操作（view, create, update, delete）
  - `level`: 级别（1-模块，2-页面，3-按钮）

## 理想的权限逻辑

**权限码 → 自动显示菜单**

```
用户分配权限码: system:user:view
    ↓
权限码的 resource = user
    ↓
查找 menu: path = /manage/user (用户管理页面)
    ↓
自动显示菜单: 系统管理 → 用户管理
```

## 实现方案

### 方案概述

1. **菜单添加 resource 字段**
2. **权限码自动关联菜单**
3. **废弃 menu_ids 字段**

### 步骤1：菜单添加 resource 字段

**修改 Menu 模型：**

```go
type Menu struct {
    ID         uint      `json:"id" gorm:"primaryKey"`
    Name       string    `json:"name" gorm:"size:50;not null"`
    Icon       string    `json:"icon" gorm:"size:50"`
    Path       string    `json:"path" gorm:"size:200"`
    Permission string    `json:"permission" gorm:"size:100"`
    Resource   string    `json:"resource" gorm:"size:30;index"` // 新增：对应的资源名称
    MenuType   string    `json:"menuType" gorm:"size:20;default:menu"`
    ParentID   uint      `json:"parentId" gorm:"default:0"`
    Sort       int       `json:"sort" gorm:"default:0"`
    Status     int       `json:"status" gorm:"default:1"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
    Children   []*Menu   `json:"children,omitempty" gorm:"-"`
}
```

**数据库迁移：**

```sql
ALTER TABLE menus ADD COLUMN resource VARCHAR(30) DEFAULT '' COMMENT '对应的资源名称' AFTER permission;
CREATE INDEX idx_menus_resource ON menus(resource);

-- 更新现有菜单的 resource 字段
UPDATE menus SET resource = 'user' WHERE path = '/manage/user';
UPDATE menus SET resource = 'role' WHERE path = '/manage/role';
UPDATE menus SET resource = 'menu' WHERE path = '/manage/menu';
UPDATE menus SET resource = 'server' WHERE path = '/cmdb/servers';
UPDATE menus SET resource = 'business' WHERE path = '/cmdb/business';
-- ... 其他菜单
```

### 步骤2：修改 BuildMenuTreeAndPermissions 逻辑

**新的权限推导逻辑：**

```go
func (s *RBACService) BuildMenuTreeAndPermissions(userID uint) ([]*models.Menu, []string, []*models.Role, error) {
    // 1. 获取用户的角色
    roles, err := s.GetUserRoles(userID)
    if err != nil {
        return nil, nil, nil, err
    }

    // 2. 获取用户的权限码（从 role_permissions 表）
    var permissionCodes []string
    for _, role := range roles {
        var rolePerms []models.RolePermission
        db.Where("role_id = ?", role.ID).Preload("Permission").Find(&rolePerms)
        for _, rp := range rolePerms {
            permissionCodes = append(permissionCodes, rp.Permission.Code)
        }
    }

    // 3. 从权限码中提取有权限的资源列表
    allowedResources := make(map[string]bool)
    for _, code := range permissionCodes {
        // 解析权限码: system:user:view → resource = user
        parts := strings.Split(code, ":")
        if len(parts) >= 2 {
            resource := parts[1]
            allowedResources[resource] = true
        }
    }

    // 4. 查询所有菜单
    var allMenus []*models.Menu
    db.Where("status = 1").Order("sort ASC").Find(&allMenus)

    // 5. 过滤菜单：如果菜单的 resource 在允许列表中，就显示
    menuIDs := make(map[uint]bool)
    for _, menu := range allMenus {
        // 一级菜单（目录）：如果子菜单有权限，父菜单也显示
        if menu.Resource == "" {
            continue
        }

        if allowedResources[menu.Resource] {
            menuIDs[menu.ID] = true

            // 同时标记父菜单
            for _, m := range allMenus {
                if m.ID == menu.ParentID {
                    menuIDs[m.ID] = true
                }
            }
        }
    }

    // 6. 构建菜单树
    menuTree := s.buildMenuTree(allMenus, menuIDs, 0)

    return menuTree, permissionCodes, roles, nil
}
```

### 步骤3：废弃 menu_ids 字段

**可选方案：**
1. **保留但忽略** - 不删除字段，但不再使用
2. **完全移除** - 删除 `roles.menu_ids` 字段和所有相关代码

**建议：** 先保留但不使用，验证新逻辑稳定后再移除。

## 权限码与菜单映射规则

### 规则定义

```
权限码格式: module:resource:action
           ↓       ↓
        系统管理  用户管理

示例:
- system:user:view    → 显示"系统管理 → 用户管理"
- system:role:view    → 显示"系统管理 → 角色管理"
- cmdb:server:view    → 显示"资产管理 → 主机管理"
- monitoring:*:view   → 显示"监控中心"所有子菜单
```

### 特殊规则

**1. 通配符权限：**
```
system:*:view  → 显示 system 模块下所有页面
*:*:*          → 显示所有菜单（超级管理员）
```

**2. 目录菜单自动显示：**
```
如果用户有 system:user:view 和 system:role:view
自动显示"系统管理"目录（即使没有单独授权）
```

**3. 至少一个权限原则：**
```
只要用户有某个资源的任意一个权限码
就显示该页面的菜单

例如：
有 system:user:view 或 system:user:create 或 system:user:update
都会显示"用户管理"菜单
```

## 前端调整

### 1. 废弃 menu-auth-modal.vue

不再需要单独分配菜单权限。

### 2. 优化 permission-assign-modal.vue

分配权限码时，可以显示对应的菜单路径：

```
权限: system:user:view
对应菜单: 系统管理 → 用户管理
操作: 查看用户列表
```

### 3. 菜单显示提示

在权限分配界面显示：
```
💡 提示：分配功能权限后，用户将自动获得对应菜单的访问权限
```

## 优势

### 1. 权限一致性
✅ 有操作权限必然有菜单权限
✅ 不会出现"能看到菜单但没有操作权限"的情况

### 2. 配置简化
✅ 只需配置一套权限系统
✅ 减少管理员操作步骤

### 3. 维护性强
✅ 新增页面只需配置权限码
✅ 菜单权限自动推导

### 4. 审计友好
✅ 权限变更记录清晰
✅ 易于追踪权限来源

## 迁移步骤

### 阶段1：兼容期（当前）
- ✅ 保留 menu_ids 机制
- ✅ 添加 resource 字段到菜单
- ✅ 实现权限码 → 菜单推导逻辑
- ✅ 两套机制并存

### 阶段2：验证期
- ✅ 逐步将现有角色转换为权限码配置
- ✅ 验证新逻辑的正确性
- ✅ 收集用户反馈

### 阶段3：正式期
- ✅ 废弃 menu_ids 机制
- ✅ 完全使用权限码控制
- ✅ 移除旧代码

## 数据库迁移示例

### menus 表

```sql
-- 添加 resource 字段
ALTER TABLE menus ADD COLUMN resource VARCHAR(30) DEFAULT '' COMMENT '对应的资源名称' AFTER permission;
CREATE INDEX idx_menus_resource ON menus(resource);

-- 更新数据
UPDATE menus SET resource = 'user' WHERE id = 70;      -- 用户管理
UPDATE menus SET resource = 'role' WHERE id = 71;      -- 角色管理
UPDATE menus SET resource = 'menu' WHERE id = 72;      -- 菜单管理
UPDATE menus SET resource = 'server' WHERE id = 20;    -- 主机管理
UPDATE menus SET resource = 'business' WHERE id = 21;  -- 业务管理
```

### permissions 表（已有）

```sql
-- 确保权限码的 resource 字段与菜单一致
UPDATE permissions SET resource = 'user' WHERE code LIKE 'system:user:%';
UPDATE permissions SET resource = 'role' WHERE code LIKE 'system:role:%';
UPDATE permissions SET resource = 'server' WHERE code LIKE 'cmdb:server:%';
```

## 测试验证

### 测试用例

**1. ops 角色测试：**
```sql
-- 分配权限码
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE code IN ('system:user:view', 'system:role:view');

-- 预期结果：
-- ✅ 显示菜单：首页、系统管理（用户管理、角色管理）
-- ❌ 不显示：资产管理、监控中心、K8s管理等
```

**2. 权限变更测试：**
```sql
-- 添加资产管理权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE code LIKE 'cmdb:server:%';

-- 预期结果：
-- ✅ 自动显示：资产管理 → 主机管理
```

## 总结

这个方案将菜单权限和功能权限统一为一套系统，通过权限码自动推导菜单显示，避免了当前两套系统分离导致的不一致问题。
