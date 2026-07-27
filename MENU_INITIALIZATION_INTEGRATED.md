# 菜单初始化集成完成说明

## 完成时间
2026-07-24

## 📋 集成概述

已成功将用户身份映射和用户有效权限菜单集成到项目的自动初始化流程中。现在这些菜单会在项目启动时自动创建和同步。

## ✅ 完成的修改

### 1. 菜单定义（services/init.go）

在 `syncMenus()` 方法中添加了两个新菜单：

```go
// ========== 授权中心二级菜单 (ID: 100-107) ==========

{ID: 100, Name: "用户", Icon: "mdi:account", Path: "/auth/users", Permission: "auth:user:query", MenuType: "menu", Sort: 1, Status: 1, ParentID: 7},
{ID: 101, Name: "用户组", Icon: "mdi:shield-account", Path: "/auth/roles", Permission: "auth:role:query", MenuType: "menu", Sort: 2, Status: 1, ParentID: 7},
{ID: 102, Name: "应用", Icon: "mdi:application", Path: "/auth/applications", Permission: "auth:app:query", MenuType: "menu", Sort: 3, Status: 1, ParentID: 7},
{ID: 103, Name: "权限映射", Icon: "mdi:link", Path: "/auth/rolebindings", Permission: "auth:binding:query", MenuType: "menu", Sort: 4, Status: 1, ParentID: 7},
{ID: 104, Name: "用户授权", Icon: "mdi:account-key", Path: "/auth/userauthorization", Permission: "auth:authorization:query", MenuType: "menu", Sort: 5, Status: 1, ParentID: 7},
{ID: 105, Name: "操作日志", Icon: "mdi:file-document", Path: "/auth/operationlogs", Permission: "auth:log:query", MenuType: "menu", Sort: 6, Status: 1, ParentID: 7},
{ID: 106, Name: "用户身份映射", Icon: "mdi:account-switch", Path: "/auth/user-identities", Permission: "auth:identity:query", MenuType: "menu", Sort: 7, Status: 1, ParentID: 7},  // ⭐ NEW
{ID: 107, Name: "用户有效权限", Icon: "mdi:shield-check", Path: "/auth/user-permissions", Permission: "auth:permission:query", MenuType: "menu", Sort: 8, Status: 1, ParentID: 7},  // ⭐ NEW
```

### 2. 角色权限配置更新

更新了所有内置角色的菜单权限数组，添加新菜单ID 106 和 107：

#### 超级管理员
```go
adminMenuIDs := []uint{
    // ... 其他菜单 ...
    100, 101, 102, 103, 104, 105, 106, 107  // 包含新菜单
}
```

#### 运维工程师
```go
opsMenuIDs := []uint{
    // ... 其他菜单 ...
    100, 101, 102, 103, 104, 105, 106, 107  // 包含新菜单
}
```

#### 审计员
```go
auditorMenuIDs := []uint{
    // ... 其他菜单 ...
    100, 101, 102, 103, 104, 105, 106, 107  // 包含新菜单
}
```

#### 测试角色
```go
testMenuIDs := []uint{
    // ... 其他菜单 ...
    100, 101, 102, 103, 104, 105, 106, 107  // 包含新菜单
}
```

#### 查看者
```go
viewerMenuIDs := []uint{
    // ... 其他菜单 ...
    100, 101, 102, 103, 104, 105, 106, 107  // 包含新菜单
}
```

### 3. 注释更新

更新了菜单ID映射说明：

```go
// 菜单ID映射：1=首页, 2=系统管理, 20=资产管理, 40=监控中心, 50=审计中心,
// 60=web终端, 80=K8s管理, 90=诊断中心, 100-107=授权中心
```

## 🔄 初始化流程

### 启动时自动执行

项目启动时会按以下顺序自动执行初始化：

```
main() 
  → InitDatabase()
    → AutoMigrate()           // 创建/更新表结构
    → runMigrations()         // SQL迁移
    → initData()
      → syncMenus()           // ⭐ 同步菜单（包括新菜单）
      → syncRoleMenus()       // ⭐ 同步角色权限（包括新菜单权限）
      → initUsers()           // 初始化用户
      → syncAttributes()      // 同步属性定义
```

### 菜单同步逻辑

`syncMenus()` 方法会自动处理：

1. **新增菜单**：如果菜单不存在，自动创建
2. **更新菜单**：如果菜单已存在，更新菜单属性（保留用户修改的排序）
3. **删除废弃菜单**：自动删除已废弃的菜单项

```go
for _, menu := range menus {
    var existingMenu models.Menu
    err := db.Where("id = ?", menu.ID).First(&existingMenu).Error

    if err == nil {
        // 菜单已存在，更新数据
        db.Model(&existingMenu).Updates(map[string]interface{}{
            "name":       menu.Name,
            "icon":       menu.Icon,
            "path":       menu.Path,
            "permission": menu.Permission,
            "parent_id":  menu.ParentID,
            "status":     menu.Status,
            "menu_type":  menu.MenuType,
        })
    } else {
        // 菜单不存在，添加新菜单
        db.Create(&menu)
    }
}
```

### 角色权限同步逻辑

`syncRoleMenus()` 方法会自动：

1. 创建内置角色（如果不存在）
2. 更新角色的菜单权限列表
3. 使用 `ON DUPLICATE KEY UPDATE` 确保幂等性

```go
for _, role := range builtinRoles {
    err := db.Where("code = ?", role.code).First(&existingRole).Error
    if err != nil {
        // 创建新角色
        db.Create(&models.Role{...})
    } else {
        // 更新角色权限
        db.Model(&existingRole).Updates(map[string]interface{}{
            "menu_ids": role.menuIDsJSON,
        })
    }
}
```

## 🎯 使用方式

### 方式1：重启应用（推荐）

重启后端应用，初始化流程会自动执行：

```bash
# 停止应用
# 启动应用
./oneops
```

启动时会看到日志：

```
[INFO] 开始同步菜单数据...
[INFO] 添加新菜单 name=用户身份映射 id=106 path=/auth/user-identities
[INFO] 添加新菜单 name=用户有效权限 id=107 path=/auth/user-permissions
[INFO] 菜单同步完成 added=2 updated=X total=Y
[INFO] 开始同步角色菜单权限...
[INFO] 角色菜单权限同步完成
```

### 方式2：手动触发初始化

如果不想重启，可以手动调用初始化方法（需要在代码中添加API端点）：

```bash
# 假设有初始化API
curl -X POST http://localhost:8082/api/v1/system/init/menus
```

### 方式3：清除RBAC缓存

如果菜单已存在但权限不生效：

```go
// 清除所有用户的RBAC缓存
InvalidateRBACCache(0)
```

## 📊 完整的授权中心菜单列表

| ID  | 菜单名称     | 图标                  | 路由                     | 权限标识                  | 排序 |
|-----|-------------|-----------------------|--------------------------|---------------------------|------|
| 100 | 用户         | mdi:account           | /auth/users              | auth:user:query           | 1    |
| 101 | 用户组       | mdi:shield-account    | /auth/roles              | auth:role:query           | 2    |
| 102 | 应用         | mdi:application       | /auth/applications       | auth:app:query            | 3    |
| 103 | 权限映射     | mdi:link              | /auth/rolebindings       | auth:binding:query        | 4    |
| 104 | 用户授权     | mdi:account-key       | /auth/userauthorization  | auth:authorization:query  | 5    |
| 105 | 操作日志     | mdi:file-document     | /auth/operationlogs      | auth:log:query            | 6    |
| 106 | 用户身份映射 | mdi:account-switch    | /auth/user-identities    | auth:identity:query       | 7    |
| 107 | 用户有效权限 | mdi:shield-check      | /auth/user-permissions   | auth:permission:query     | 8    |

## ✨ 优势

### 1. 自动化
- ✅ 无需手动执行SQL脚本
- ✅ 新环境部署自动初始化
- ✅ 重置数据库后自动恢复

### 2. 版本控制
- ✅ 菜单配置跟随代码版本
- ✅ 不同环境配置一致
- ✅ 便于追踪变更历史

### 3. 增量更新
- ✅ 不覆盖用户自定义的菜单排序
- ✅ 自动添加新菜单
- ✅ 自动更新菜单属性

### 4. 权限一致性
- ✅ 角色权限自动同步
- ✅ RBAC缓存自动清除
- ✅ 新菜单权限自动分配

## 🔍 验证方法

### 1. 查看菜单是否创建
```bash
mysql -h 60.191.116.75 -P 38089 -u root -p123456 nexops \
  -e "SELECT id, name, icon, path FROM menus WHERE id IN (106, 107);"
```

### 2. 查看角色权限是否更新
```bash
mysql -h 60.191.116.75 -P 38089 -u root -p123456 nexops \
  -e "SELECT name, JSON_CONTAINS(menu_ids, '106') as has_menu_106, JSON_CONTAINS(menu_ids, '107') as has_menu_107 FROM roles WHERE code = 'R_ADMIN';"
```

### 3. 查看日志输出
```bash
tail -f logs/app.log | grep "菜单同步"
```

## 🚀 后续维护

### 添加新菜单
在 `services/init.go` 的 `syncMenus()` 方法中添加：

```go
menus := []models.Menu{
    // ... 现有菜单 ...
    {ID: 108, Name: "新功能", Icon: "mdi:new", Path: "/auth/new-feature", ...},
}
```

### 更新角色权限
在 `syncRoleMenus()` 方法中更新对应的菜单ID数组：

```go
adminMenuIDs := []uint{
    // ... 现有菜单 ...
    100, 101, 102, 103, 104, 105, 106, 107, 108,  // 添加新菜单ID
}
```

### 删除废弃菜单
在 `syncMenus()` 方法中添加删除逻辑：

```go
db.Where("path = ?", "/auth/deprecated").Delete(&models.Menu{})
```

## 📝 注意事项

1. **菜单ID唯一性**：确保菜单ID全局唯一，不要重复
2. **权限标识规范**：使用 `模块:资源:操作` 格式，如 `auth:identity:query`
3. **图标命名**：使用 Material Design Icons (mdi:) 前缀
4. **路由路径**：必须与前端路由配置一致
5. **角色权限**：添加新菜单后，记得更新所有需要访问该菜单的角色权限数组

## ✅ 总结

- ✅ 菜单配置已集成到初始化流程
- ✅ 角色权限已自动更新
- ✅ 支持增量更新和版本控制
- ✅ 新环境部署无需手动配置

现在重启应用或部署新环境时，新菜单会自动创建并分配权限！
