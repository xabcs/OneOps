# 权限问题诊断和修复 SQL

## 1. 查看当前 ops 角色配置

```sql
-- 连接数据库
mysql -h 60.191.116.75 -P 38089 -u root -p123456 msre

-- 查询ops角色的menu_ids
SELECT id, code, name, menu_ids
FROM roles
WHERE code = 'ops';
```

**预期结果：** `menu_ids` 应该是 `[1,6,70,71]`

**实际情况（根据日志推测）：** `menu_ids` 可能是类似这样的：
```json
[1,2,3,4,5,13,14,20,21,22,23,24,60,80,87,88,89,25,26,27,28,29,30,31,32,33,34,35,36,37,38,40,41,42,43,44,45,46,50,51,52,90,100,101,102,103,104,105,106,107]
```

这包含了几乎所有菜单！

## 2. 查看菜单ID对应关系

```sql
-- 查看所有菜单
SELECT id, name, path, parent_id
FROM menus
WHERE status = 1
ORDER BY parent_id, sort;
```

**关键菜单ID：**
- 1 = 首页
- 2 = 资产管理
- 3 = 监控中心
- 4 = 审计中心
- 5 = K8s管理
- **6 = 系统管理** ✓
- 7 = 授权中心
- **70 = 用户管理** ✓（parent_id=6）
- **71 = 角色管理** ✓（parent_id=6）
- **72 = 菜单管理**（parent_id=6）

## 3. 修复 ops 角色权限

```sql
-- 正确的ops角色权限：仅首页和系统管理（用户管理、角色管理）
UPDATE roles
SET menu_ids = '[1, 6, 70, 71]'
WHERE code = 'ops';

-- 验证更新
SELECT id, code, name, menu_ids
FROM roles
WHERE code = 'ops';
```

## 4. 检查其他角色的权限

```sql
-- 查看所有角色的权限配置
SELECT code, name, menu_ids
FROM roles
ORDER BY code;

-- 如果admin角色的menu_ids也是错的，应该设置为所有菜单
UPDATE roles
SET menu_ids = '[1,2,3,4,5,6,7,13,14,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,40,41,42,43,44,45,46,50,51,52,60,70,71,72,80,87,88,89,90,100,101,102,103,104,105,106,107]'
WHERE code = 'admin';
```

## 5. 清除缓存

更新数据库后，**必须清除RBAC缓存**：

### 方法1：调用API（推荐）

```bash
# 获取token（使用admin账号登录）
TOKEN=$(curl -X POST http://127.0.0.1:9527/proxy-default/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

# 清除缓存
curl -X POST http://127.0.0.1:9527/proxy-default/route/invalidate-cache \
  -H "Authorization: Bearer $TOKEN"
```

### 方法2：重启后端

```bash
# 重启后端服务
# 具体命令取决于你的部署方式
```

## 6. 验证修复结果

### 6.1 重新登录test用户

```bash
# 清除浏览器缓存或使用隐身模式
# 访问登录页面
# 使用test/123456登录
```

### 6.2 查看后端日志

登录时应该看到类似的日志：

```
🐛 [登录调试-RBAC] BuildMenuTreeAndPermissions开始
🐛 [登录调试-RBAC] 菜单树构建完成, 菜单树节点数: 3
```

**菜单树应该只包含：**
- 首页
- 系统管理
  - 用户管理
  - 角色管理

### 6.3 查看前端日志

浏览器控制台应该看到：

```javascript
🐛 [路由Store] 后端返回的路由数据: [
  {
    "component": "layout.base$view.home",
    "id": "1",
    "path": "/home",
    "name": "home"
  },
  {
    "component": "layout.base",
    "id": "6",
    "path": "/manage",
    "name": "manage",
    "children": [
      {
        "component": "view.manage_user",
        "id": "70",
        "path": "/manage/user",
        "name": "manage_user"
      },
      {
        "component": "view.manage_role",
        "id": "71",
        "path": "/manage/role",
        "name": "manage_role"
      }
    ]
  }
]
```

**只有这3个路由！**

### 6.4 查看用户权限码

```javascript
用户权限码: Proxy(Array) {
  0: 'system:user:query',
  1: 'system:role:query'
}
```

**只有系统管理的权限！**

## 7. 检查 syncRoleMenus 问题

查看后端日志，确认是否有类似信息：

```
更新内置角色权限 name=运维工程师 code=ops menus=XX
```

如果看到这个日志，说明 `syncRoleMenus` 在启动时**覆盖了你的配置**。

### 修复方法

编辑 `backend/services/init.go` 第642-648行，删除 `menu_ids` 更新：

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

然后重新编译和部署后端。

## 8. 通过前端界面修改权限（替代方案）

如果不想直接操作数据库，可以：

1. 使用admin账号登录
2. 进入"系统管理 → 角色管理"
3. 找到ops角色，点击"分配权限"
4. 只勾选：
   - ✅ 首页
   - ✅ 系统管理
     - ✅ 用户管理
     - ✅ 角色管理
5. 点击确定
6. 清除缓存（调用API或重启后端）

## 预期结果

修复后，test用户登录应该只能看到：

**左侧菜单：**
- 首页
- 系统管理
  - 用户管理
  - 角色管理

**不应该出现：**
- ❌ 资产管理
- ❌ 监控中心
- ❌ K8s管理
- ❌ 审计中心
- ❌ 授权中心
- ❌ web终端
