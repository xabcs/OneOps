# 权限问题已修复 - 待重启验证

## 问题根源

**ops 角色的 `menu_ids` 字段包含了大量未授权的菜单ID**

这导致即使用户只有用户管理、角色管理、菜单管理的权限，也能看到其他页面（资产管理、监控中心、K8s管理等）。

## 已完成的修复

### 1. 修复菜单 resource 字段

```sql
-- 修复前：权限管理菜单的 resource 为 NULL
-- 修复后：resource = 'permission'
UPDATE menus SET resource = 'permission' WHERE id = 108;
```

**验证结果：**

```
id | name       | path                 | resource
---|-----------|----------------------|----------
70 | 用户管理   | /manage/user         | user
71 | 角色管理   | /manage/role         | role
72 | 菜单管理   | /manage/menu         | menu
108| 权限管理   | /manage/permission   | permission
```

### 2. 清空 ops 角色的 menu_ids

```sql
-- 修复前：包含大量未授权菜单 [1,2,3,4,5,13,14,20,...]
-- 修复后：空数组 []
UPDATE roles SET menu_ids = '[]' WHERE code = 'ops';
```

**验证结果：**

```
id | code | name       | menu_ids
---|------|-----------|---------
2  | ops  | 运维工程师  | []
```

### 3. 权限码配置正确

ops 角色已分配的权限：

| Code | Name | Resource | Level |
|------|------|----------|-------|
| system | 系统管理 | '' | 1 |
| system.user | 用户管理 | user | 2 |
| system.role | 角色管理 | role | 2 |
| system.menu | 菜单管理 | menu | 2 |
| system.user.view | 查看用户 | user | 3 |
| system.role.view | 查看角色 | role | 3 |
| system.menu.view | 查看菜单 | menu | 3 |

## 权限推导逻辑

### 权限码 → 菜单映射

```
用户权限码:
- system.user (resource="user")
- system.role (resource="role")
- system.menu (resource="menu")
    ↓
查找菜单:
- resource="user" → 用户管理 (id=70)
- resource="role" → 角色管理 (id=71)
- resource="menu" → 菜单管理 (id=72)
    ↓
自动标记父菜单:
- parent_id=6 → 系统管理 (id=6)
    ↓
最终菜单:
[首页(1), 系统管理(6), 用户管理(70), 角色管理(71), 菜单管理(72)]
```

## 下一步操作

### 1. 重启后端清除缓存

```bash
# 停止后端
pkill -f "go run main.go"

# 或找到进程并杀掉
ps aux | grep "go run main.go"
kill -9 <PID>

# 启动后端
cd backend
go run main.go
```

### 2. test 用户重新登录

1. 清除浏览器缓存或使用隐身模式
2. 访问登录页面
3. 使用 test/123456 登录

### 3. 验证预期结果

**应该看到的菜单：**
- ✅ 首页
- ✅ 系统管理
  - ✅ 用户管理
  - ✅ 角色管理
  - ✅ 菜单管理

**不应该看到的菜单：**
- ❌ 资产管理
- ❌ 监控中心
- ❌ K8s管理
- ❌ 审计中心
- ❌ 授权中心
- ❌ web终端
- ❌ 权限管理（因为没有分配 system.permission 权限）

### 4. 查看后端日志

登录时应该看到类似日志：

```
[登录调试-RBAC] 开始从权限码推导菜单
[登录调试-RBAC] 用户权限码 数量=7
[登录调试-RBAC] 用户有权限的资源 数量=3 ["user", "role", "menu"]
[登录调试-RBAC] 从权限码推导的菜单数量 数量=4
```

### 5. 查看前端日志

浏览器控制台应该看到：

```javascript
🐛 [路由Store] 后端返回的路由数据: [
  { id: "1", path: "/home", name: "home" },
  {
    id: "6",
    path: "/manage",
    name: "manage",
    children: [
      { id: "70", path: "/manage/user", name: "manage_user" },
      { id: "71", path: "/manage/role", name: "manage_role" },
      { id: "72", path: "/manage/menu", name: "manage_menu" }
    ]
  }
]
```

## 按钮级权限验证

在用户管理页面，应该：
- ✅ 能看到"查看"按钮（有 system.user.view 权限）
- ❌ 不能看到"创建"按钮（没有 system.user.create 权限）
- ❌ 不能看到"编辑"按钮（没有 system.user.update 权限）
- ❌ 不能看到"删除"按钮（没有 system.user.delete 权限）

在角色管理页面，同理：
- ✅ 能看到"查看"按钮
- ❌ 看不到其他操作按钮

## 如果还是不对

### 检查点

1. 后端是否重启？（清除 RBAC 缓存）
2. 浏览器缓存是否清除？
3. 检查后端日志是否有权限码推导的日志
4. 检查前端日志中后端返回的路由数量

### 数据库验证

```sql
-- 确认 ops 角色的 menu_ids 为空
SELECT menu_ids FROM roles WHERE code = 'ops';

-- 确认 ops 角色的权限正确
SELECT p.code, p.resource, p.level
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE rp.role_id = 2
ORDER BY p.level, p.resource;

-- 确认菜单 resource 字段正确
SELECT id, name, path, resource
FROM menus
WHERE parent_id = 6
ORDER BY sort;
```

## 总结

✅ 数据库已修复完成
⏳ 需要重启后端清除缓存
⏳ 需要 test 用户重新登录验证

现在权限系统应该能正常工作：**用户只能看到分配了权限码的页面**。
