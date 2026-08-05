# 权限问题诊断步骤

## 问题描述

- ops角色应该只有系统管理（用户管理、角色管理）的权限
- test用户绑定ops角色
- test用户登录后能看到其他页面（不符合授权）

## 诊断步骤

### 1. 检查数据库中ops角色的权限配置

```sql
-- 连接数据库
mysql -h 60.191.116.75 -P 38089 -u root -p123456 msre

-- 查询ops角色的menu_ids
SELECT id, code, name, menu_ids
FROM roles
WHERE code = 'ops';

-- 正确的menu_ids应该是：[1, 6, 70, 71]
-- 1 = 首页
-- 6 = 系统管理
-- 70 = 用户管理
-- 71 = 角色管理
```

如果 `menu_ids` 不正确，说明 `syncRoleMenus` 覆盖了你的配置。需要手动更新：

```sql
-- 手动更新ops角色权限（仅首页和系统管理）
UPDATE roles
SET menu_ids = '[1, 6, 70, 71]'
WHERE code = 'ops';
```

### 2. 检查test用户的角色绑定

```sql
-- 查询test用户的角色
SELECT id, username, role_ids
FROM users
WHERE username = 'test';

-- role_ids应该是包含ops角色ID的数组，例如 [2]
```

### 3. 清除RBAC缓存

更新数据库后，需要清除缓存：

```bash
# 方法1：调用缓存清除接口
curl -X POST http://127.0.0.1:9527/proxy-default/route/invalidate-cache \
  -H "Authorization: Bearer YOUR_TOKEN"

# 方法2：重启后端服务
```

### 4. 查看后端日志确认

使用test用户登录，查看后端日志：

```
🐛 [GetUserRoutes] 开始生成路由 userID=?
🐛 [登录调试-RBAC] BuildMenuTreeAndPermissions开始
```

确认返回的菜单树只包含：
- 首页
- 系统管理
  - 用户管理
  - 角色管理

### 5. 查看前端日志确认

打开浏览器控制台（F12），查看：

```
🐛 [路由Store] 后端返回的路由数据: [...]
```

确认前端收到的路由数据只包含授权的页面。

### 6. 检查前端路由守卫

如果前端收到了正确的路由，但还能看到其他页面，检查：

```
🔍 [路由守卫] 开始处理路由跳转
🔐 [路由守卫] 认证状态检查:
  - 用户角色名: ['ops']
  - 路由要求角色: []
  - 路由要求权限: []
  - hasAuth: true  <-- 如果路由没有配置roles/permissions，hasAuth会为true
```

**这是问题所在！** 参见 `frontend/src/router/guard/route.ts:43`

## 根本原因

### 问题1：syncRoleMenus 覆盖用户配置

**文件：** `backend/services/init.go:642-648`

```go
if err == nil {
    // 角色存在，更新权限
    db.Model(&existingRole).Updates(map[string]interface{}{
        "name":        builtinRole.name,
        "description": builtinRole.description,
        "menu_ids":    string(builtinRole.menuIDsJSON),  // ❌ 覆盖用户配置
    })
}
```

**解决方案：** 不更新 `menu_ids`，保留用户通过界面配置的权限。

### 问题2：前端路由守卫逻辑错误

**文件：** `frontend/src/router/guard/route.ts:43`

```typescript
const hasAuth = authStore.isStaticSuper || (!routeRoles.length && !routePermissions.length) || hasRole || hasPermission;
```

**问题：** `(!routeRoles.length && !routePermissions.length)` 允许所有登录用户访问没有配置roles/permissions的路由。

**正确的逻辑：**
- 后端已经通过 menuTree 控制了用户能访问的路由
- 前端动态注册后端返回的路由
- **路由守卫应该检查路由是否在已注册的路由中，而不是检查 meta.roles**

## 修复方案

### 方案A：修改路由守卫逻辑（推荐）

既然后端已经返回了用户有权限的路由，前端应该：

1. **检查路由是否存在：** 如果路由不存在（not-found），说明用户没有权限
2. **不再检查 roles/permissions：** 因为后端已经过滤了

修改 `frontend/src/router/guard/route.ts`:

```typescript
// 简化的权限检查：只要路由存在且已登录就有权限
const hasAuth = isLogin && to.name !== 'not-found';
```

### 方案B：保持现有逻辑，但为所有路由添加 roles 或 permissions

这需要在 `backend/controllers/route.go` 的 `convertMenusToRoutes` 方法中，为每个路由添加 `meta.roles` 或 `meta.permissions`。

## 当前临时解决方案

1. **手动更新数据库：**
   ```sql
   UPDATE roles SET menu_ids = '[1, 6, 70, 71]' WHERE code = 'ops';
   ```

2. **清除缓存：**
   - 调用 `/api/route/invalidate-cache` 接口
   - 或重启后端

3. **重新登录 test 用户**

## 验证

登录test用户后，应该只能看到：
- ✅ 首页
- ✅ 系统管理
  - ✅ 用户管理
  - ✅ 角色管理
- ❌ 其他菜单不应该出现
