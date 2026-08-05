# 角色权限保存不一致问题 - 完整诊断和修复方案

## 问题现象

用户在前端"角色管理"中为ops角色勾选了菜单权限（仅系统管理的用户管理、角色管理），但：
1. 数据库中保存的是错误的权限配置 `[1,2,3,4,5,13,14,20,21,22,...]`
2. test用户（绑定ops角色）登录后能看到几乎所有页面

## 根本原因

### 问题1：syncRoleMenus 覆盖用户配置 ✅ 已修复

**文件：** `backend/services/init.go:642-648`

系统每次启动时，`syncRoleMenus` 会执行，把所有内置角色的权限**强制覆盖**回代码中写死的配置。

**修复：** 已修改代码，角色存在时只更新名称和描述，**不更新 menu_ids**

```go
// 修复前：
if err == nil {
    db.Model(&existingRole).Updates(map[string]interface{}{
        "name":        builtinRole.name,
        "description": builtinRole.description,
        "menu_ids":    string(builtinRole.menuIDsJSON),  // ❌ 覆盖用户配置
    })
}

// 修复后：
if err == nil {
    db.Model(&existingRole).Updates(map[string]interface{}{
        "name":        builtinRole.name,
        "description": builtinRole.description,
        // ✅ 不更新 menu_ids，保留用户通过界面配置的权限
    })
}
```

### 问题2：前端保存可能有问题

**文件：** `frontend/src/views/manage/role/modules/menu-auth-modal.vue:108-129`

前端逻辑看起来是正确的：
```typescript
async function handleSubmit() {
  const checkedKeys = treeRef.value.getCheckedKeys() || [];
  const halfCheckedKeys = treeRef.value.getHalfCheckedKeys() || [];
  const allKeys = [...checkedKeys, ...halfCheckedKeys];

  await fetchUpdateRole(props.roleId, {
    menuIds: allKeys  // 传递选中的菜单ID
  });
}
```

## 验证步骤

### 步骤1：重新编译后端

```bash
cd backend
go build
```

### 步骤2：手动更新数据库（清除错误的权限配置）

```sql
-- 连接数据库
mysql -h 60.191.116.75 -P 38089 -u root -p123456 msre

-- 正确的ops角色权限：首页 + 系统管理（用户管理、角色管理）
UPDATE roles
SET menu_ids = '[1, 6, 70, 71]'
WHERE code = 'ops';

-- 验证
SELECT id, code, name, menu_ids FROM roles WHERE code='ops';
```

**预期结果：** `menu_ids = [1, 6, 70, 71]`

**菜单ID说明：**
- 1 = 首页
- 6 = 系统管理（一级目录）
- 70 = 用户管理（系统管理的子菜单）
- 71 = 角色管理（系统管理的子菜单）

### 步骤3：清除RBAC缓存

```bash
# 方法1：调用API（需要token）
curl -X POST http://127.0.0.1:9527/proxy-default/route/invalidate-cache \
  -H "Authorization: Bearer YOUR_TOKEN"

# 方法2：重启后端
```

### 步骤4：验证权限

使用 test 用户登录（密码：123456），检查：

**浏览器控制台日志：**
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
      { "id": "71", "path": "/manage/role", "name": "manage_role" }
    ]
  }
]
```

**用户权限码：**
```javascript
用户权限码: Proxy(Array) {
  0: 'system:user:query',
  1: 'system:role:query'
}
```

**左侧菜单：**
- ✅ 首页
- ✅ 系统管理
  - ✅ 用户管理
  - ✅ 角色管理

**不应该出现：**
- ❌ 资产管理
- ❌ 监控中心
- ❌ K8s管理
- ❌ 审计中心
- ❌ 授权中心
- ❌ web终端

### 步骤5：验证系统重启不覆盖配置

```bash
# 1. 重启后端
# 2. 查看后端日志，应该看到：
# ✅ "更新内置角色信息（不更新权限）" code=ops

# 3. 再次查询数据库
SELECT id, code, name, menu_ids FROM roles WHERE code='ops';

# 预期：menu_ids 仍然是 [1, 6, 70, 71]
```

## 前端调试

如果权限还是不对，打开浏览器控制台查看：

### 1. 检查菜单树数据

```javascript
// 在 menu-auth-modal.vue 的 handleSubmit 中添加日志
async function handleSubmit() {
  const checkedKeys = treeRef.value.getCheckedKeys() || [];
  const halfCheckedKeys = treeRef.value.getHalfCheckedKeys() || [];
  const allKeys = [...checkedKeys, ...halfCheckedKeys];

  console.log('🔍 [分配菜单权限] 选中的菜单ID:', {
    checkedKeys,
    halfCheckedKeys,
    allKeys
  });

  // 调用后端API
  const { error } = await fetchUpdateRole(props.roleId, {
    menuIds: allKeys
  });

  console.log('📤 [分配菜单权限] 保存结果:', error ? '失败' : '成功');
}
```

### 2. 检查网络请求

打开浏览器开发者工具 → Network 标签页，查看：
- 请求URL: `PUT /api/system/roles/2`
- 请求Payload:
  ```json
  {
    "menuIds": [1, 6, 70, 71]
  }
  ```
- 响应:
  ```json
  {
    "code": 200,
    "success": true,
    "message": "更新成功"
  }
  ```

## 后端调试

如果前端传递的数据正确但数据库保存失败，检查后端日志：

### 1. 添加调试日志

编辑 `backend/controllers/role.go:143` 的 `UpdateRole` 方法：

```go
func (ctrl *RoleController) UpdateRole(c *gin.Context) {
    // ... 解析参数 ...

    logger.Info("🔥 [UpdateRole] 更新角色",
        zap.Uint("id", uint(id)),
        zap.Any("menuIds", req.MenuIDs))

    // ... 更新数据库 ...

    if err := db.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
        logger.Error("❌ [UpdateRole] 更新失败", zap.Error(err))
    } else {
        logger.Info("✅ [UpdateRole] 更新成功", zap.Any("updates", updates))
    }
}
```

### 2. 查看日志输出

```
🔥 [UpdateRole] 更新角色 id=2 menuIds=[1,6,70,71]
✅ [UpdateRole] 更新成功 updates={"menu_ids":"[1,6,70,71]"}
```

## 问题排查清单

- [ ] 后端代码已修复（`syncRoleMenus` 不覆盖用户配置）
- [ ] 后端已重新编译
- [ ] 数据库中ops角色的menu_ids已手动更正为 `[1, 6, 70, 71]`
- [ ] RBAC缓存已清除
- [ ] 浏览器缓存已清除或使用隐身模式
- [ ] test用户重新登录
- [ ] 验证前端日志：后端返回的路由只包含首页和系统管理
- [ ] 验证左侧菜单：只显示首页和系统管理
- [ ] 系统重启后，数据库权限配置没有被覆盖

## 预期结果

修复完成后：
1. ✅ 用户通过前端分配的菜单权限能正确保存到数据库
2. ✅ 系统重启不会覆盖用户配置的权限
3. ✅ test用户登录后只能看到授权的页面
