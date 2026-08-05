# 权限配置诊断步骤

## 当前问题

test 用户（绑定 ops 角色）登录后，看到的菜单与分配的权限不匹配。

## 根据日志分析

### ops 角色已分配的权限

**前端日志显示：**

```
权限IDs: [1, 2, 7, 12, 3, 8, 13]
过滤后按钮权限IDs: [3, 8, 13]
```

**权限详情：**

| ID  | 名称 | Code | Resource | Level | 说明 |
|-----|------|------|----------|-------|------|
| 1 | 系统管理 | system | '' | 1 | 模块级权限 |
| 2 | 用户管理 | system.user | user | 2 | 页面级权限 |
| 7 | 角色管理 | system.role | role | 2 | 页面级权限 |
| 12 | 菜单管理 | system.menu | menu | 2 | 页面级权限 |
| 3 | 查看用户 | system.user.view | user | 3 | 按钮级权限 |
| 8 | 查看角色 | system.role.view | role | 3 | 按钮级权限 |
| 13 | 查看菜单 | system.menu.view | menu | 3 | 按钮级权限 |

**预期菜单：**
- ✅ 首页
- ✅ 系统管理
  - ✅ 用户管理
  - ✅ 角色管理
  - ✅ 菜单管理

## 诊断步骤

### 1. 检查数据库中的 resource 映射

执行以下 SQL：

```sql
-- 连接数据库
mysql -h 60.191.116.75 -P 38089 -u root -p123456 msre

-- 查看权限码的 resource 字段
SELECT id, code, name, module, resource, action, level
FROM permissions
WHERE module = 'system' AND resource IN ('user', 'role', 'menu')
ORDER BY level, resource, action;

-- 查看菜单的 resource 字段
SELECT id, name, path, resource, parent_id
FROM menus
WHERE parent_id = 6 OR id = 6
ORDER BY parent_id, sort;
```

**预期结果：**

**permissions 表：**
```
id | code                  | resource | level
---|----------------------|----------|-------
2  | system.user          | user     | 2
3  | system.user.view     | user     | 3
7  | system.role          | role     | 2
8  | system.role.view     | role     | 3
12 | system.menu          | menu     | 2
13 | system.menu.view     | menu     | 3
```

**menus 表：**
```
id | name       | path            | resource | parent_id
---|-----------|-----------------|----------|----------
6  | 系统管理   | /manage         | ''       | 0
70 | 用户管理   | /manage/user    | user     | 6
71 | 角色管理   | /manage/role    | role     | 6
72 | 菜单管理   | /manage/menu    | menu     | 6
```

**如果 menus.resource 为空或不匹配，执行：**

```sql
-- 手动更新菜单的 resource 字段
UPDATE menus SET resource = 'user' WHERE id = 70;
UPDATE menus SET resource = 'role' WHERE id = 71;
UPDATE menus SET resource = 'menu' WHERE id = 72;

-- 验证
SELECT id, name, path, resource FROM menus WHERE parent_id = 6;
```

### 2. 检查 role_permissions 表

```sql
-- 查看 ops 角色的权限配置
SELECT rp.role_id, rp.permission_id, p.code, p.name, p.resource, p.level
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE rp.role_id = 2
ORDER BY p.level, p.resource;
```

**预期结果：**

```
role_id | permission_id | code              | resource | level
--------|--------------|-------------------|----------|-------
2       | 1            | system            | ''       | 1
2       | 2            | system.user       | user     | 2
2       | 7            | system.role       | role     | 2
2       | 12           | system.menu       | menu     | 2
2       | 3            | system.user.view  | user     | 3
2       | 8            | system.role.view  | role     | 3
2       | 13           | system.menu.view  | menu     | 3
```

**如果权限不对，执行：**

```sql
-- 清空 ops 角色的现有权限
DELETE FROM role_permissions WHERE role_id = 2;

-- 重新分配权限（假设 ops 角色的 id = 2）
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions
WHERE code IN (
    'system',           -- 系统管理模块
    'system.user',      -- 用户管理页面
    'system.role',      -- 角色管理页面
    'system.menu',      -- 菜单管理页面
    'system.user.view', -- 查看用户
    'system.role.view', -- 查看角色
    'system.menu.view'  -- 查看菜单
);

-- 验证
SELECT COUNT(*) FROM role_permissions WHERE role_id = 2;
-- 应该返回 7
```

### 3. 检查 roles.menu_ids

```sql
SELECT id, code, name, menu_ids
FROM roles
WHERE code = 'ops';
```

**建议：** 清空 menu_ids，让系统完全使用权限码推导：

```sql
-- 清空 ops 角色的 menu_ids
UPDATE roles SET menu_ids = '[]' WHERE code = 'ops';

-- 或者只保留首页
UPDATE roles SET menu_ids = '[1]' WHERE code = 'ops';
```

### 4. 清除 RBAC 缓存

```bash
# 方法1：调用 API（需要 token）
curl -X POST http://127.0.0.1:9527/proxy-default/route/invalidate-cache \
  -H "Authorization: Bearer YOUR_TOKEN"

# 方法2：重启后端
# 在后端目录执行
pkill -f "go run main.go"
go run main.go
```

### 5. 查看 test 用户登录时的后端日志

登录时应该看到类似日志：

```
[登录调试-RBAC] 开始从权限码推导菜单
[登录调试-RBAC] 用户权限码 数量=7 ["system", "system.user", "system.role", ...]
[登录调试-RBAC] 用户有权限的资源 数量=3 ["user", "role", "menu"]
[登录调试-RBAC] 从权限码推导的菜单数量 数量=4
```

**详细日志解读：**

```
用户权限码: ["system", "system.user", "system.role", "system.menu", 
             "system.user.view", "system.role.view", "system.menu.view"]
    ↓
提取 resource: ["user", "role", "menu"]  // system 没有对应的菜单
    ↓
查找菜单:
  - resource="user" → 找到 id=70 (用户管理)
  - resource="role" → 找到 id=71 (角色管理)
  - resource="menu" → 找到 id=72 (菜单管理)
    ↓
标记菜单: [70, 71, 72] + 父菜单 [6]
    ↓
最终菜单: [1(首页), 6(系统管理), 70(用户管理), 71(角色管理), 72(菜单管理)]
```

### 6. 查看前端日志

使用 test 用户登录，查看浏览器控制台：

**期望看到：**

```javascript
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
      { "id": "71", "path": "/manage/role", "name": "manage_role" },
      { "id": "72", "path": "/manage/menu", "name": "manage_menu" }
    ]
  }
]
```

## 常见问题

### 问题1：菜单 resource 字段为空

**原因：** 数据库迁移没有执行

**解决：** 重启后端，迁移脚本会自动执行

**或手动更新：**

```sql
UPDATE menus SET resource = 'user' WHERE path = '/manage/user';
UPDATE menus SET resource = 'role' WHERE path = '/manage/role';
UPDATE menus SET resource = 'menu' WHERE path = '/manage/menu';
```

### 问题2：权限码的 resource 字段错误

**检查：**

```sql
SELECT code, resource FROM permissions WHERE code LIKE 'system.%';
```

**修复：**

```sql
UPDATE permissions SET resource = 'user' WHERE code LIKE 'system.user%';
UPDATE permissions SET resource = 'role' WHERE code LIKE 'system.role%';
UPDATE permissions SET resource = 'menu' WHERE code LIKE 'system.menu%';
```

### 问题3：后端日志显示推导了菜单，但前端没有显示

**可能原因：**
1. menu_ids 字段干扰
2. 缓存未清除

**解决：**

```sql
-- 清空 menu_ids
UPDATE roles SET menu_ids = '[]' WHERE code = 'ops';
```

然后重启后端。

## 完整测试流程

### 1. 数据库准备

```sql
-- 1. 清空 ops 角色的 menu_ids
UPDATE roles SET menu_ids = '[]' WHERE code = 'ops';

-- 2. 清空 ops 角色的现有权限
DELETE FROM role_permissions WHERE role_id = 2;

-- 3. 重新分配权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions
WHERE code IN (
    'system', 'system.user', 'system.role', 'system.menu',
    'system.user.view', 'system.role.view', 'system.menu.view'
);

-- 4. 确保菜单 resource 字段正确
UPDATE menus SET resource = 'user' WHERE id = 70;
UPDATE menus SET resource = 'role' WHERE id = 71;
UPDATE menus SET resource = 'menu' WHERE id = 72;

-- 5. 验证
SELECT id, name, resource FROM menus WHERE parent_id = 6;
SELECT COUNT(*) FROM role_permissions WHERE role_id = 2;
```

### 2. 重启后端

```bash
# 停止后端
pkill -f "go run main.go"

# 启动后端
cd backend
go run main.go
```

### 3. 测试用户登录

1. 清除浏览器缓存或使用隐身模式
2. 使用 test 用户登录（密码：123456）
3. 查看浏览器控制台日志
4. 验证左侧菜单

**预期菜单：**
- ✅ 首页
- ✅ 系统管理
  - ✅ 用户管理
  - ✅ 角色管理
  - ✅ 菜单管理

**不应该出现的菜单：**
- ❌ 资产管理
- ❌ 监控中心
- ❌ K8s管理
- ❌ 审计中心
- ❌ 授权中心
- ❌ 权限管理（因为没有分配 system.permission 权限）

## 验证成功的标志

1. ✅ 后端日志显示从权限码推导菜单
2. ✅ 前端日志显示后端返回正确数量的路由
3. ✅ 左侧菜单只显示授权的页面
4. ✅ 页面上只显示授权的按钮（查看用户、查看角色、查看菜单）
5. ✅ 没有授权的按钮不显示（创建、更新、删除）

## 如果还是不对

请提供以下信息：

1. 后端启动日志（特别是"更新菜单resource字段"部分）
2. test 用户登录时的后端日志（包含"[登录调试-RBAC]"的日志）
3. 前端控制台日志（包含"🐛 [路由Store]"的日志）
4. 执行以下 SQL 的结果：

```sql
SELECT id, name, path, resource FROM menus WHERE parent_id = 6;
SELECT code, resource, level FROM permissions WHERE code LIKE 'system.%';
SELECT rp.role_id, p.code, p.resource FROM role_permissions rp 
JOIN permissions p ON rp.permission_id = p.id 
WHERE rp.role_id = 2;
SELECT menu_ids FROM roles WHERE code = 'ops';
```
