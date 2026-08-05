# 权限修复步骤

## 问题分析

ops 角色的权限配置错误，包含了几乎所有菜单，导致 ops 用户能看到未授权的页面。

## 修改内容

### 1. 后端代码修改

#### `backend/services/init.go`
- **修改 ops 角色的 menuIDs**（第 617 行）：
  - 修改前：`{1, 2, 3, 4, 5, 13, 14, 20, 21, 22, ...}` - 包含几乎所有菜单
  - 修改后：`{1, 6, 70, 71}` - 仅包含首页和系统管理（用户管理、角色管理）

- **添加公开的 SyncRoleMenus 方法**（第 608-611 行）：
  ```go
  // SyncRoleMenus 公开的角色菜单权限同步方法（用于外部调用）
  func (s *InitService) SyncRoleMenus() error {
      return s.syncRoleMenus()
  }
  ```

#### `backend/controllers/route.go`
- **修改 InvalidateCache 接口**（第 232-248 行）：
  - 添加角色权限同步调用
  - 接口会同时同步菜单和角色权限

## 执行步骤

### 1. 重启后端服务

```bash
# 停止当前后端服务
# 然后重新启动
cd backend
go run main.go
```

### 2. 调用缓存清除接口同步权限

使用管理员账号（admin/admin123）登录后，调用以下接口：

```bash
# 方法1：使用 curl（需要先获取 token）
curl -X POST http://127.0.0.1:9527/proxy-default/route/invalidate-cache \
  -H "Authorization: Bearer YOUR_TOKEN"

# 方法2：通过前端管理界面
# 登录后访问：系统管理 -> 菜单管理 -> 点击"清除缓存"按钮
```

### 3. 验证权限

1. 使用 test 用户登录（密码：123456）
2. 检查左侧菜单，应该只能看到：
   - 首页
   - 系统管理
     - 用户管理
     - 角色管理
3. 不应该看到其他菜单（资产管理、监控中心、K8s管理等）

### 4. 检查数据库（可选）

```sql
-- 查看ops角色的menuIDs
SELECT id, code, name, menu_ids
FROM roles
WHERE code = 'ops';

-- menu_ids 应该是 [1,6,70,71]
```

## 菜单ID映射

| ID  | 名称         | 路径               | 父级ID |
|-----|-------------|-------------------|-------|
| 1   | 首页         | /home             | 0     |
| 6   | 系统管理     | /manage           | 0     |
| 70  | 用户管理     | /manage/user      | 6     |
| 71  | 角色管理     | /manage/role      | 6     |
| 72  | 菜单管理     | /manage/menu      | 6     |

## 权限验证逻辑

后端权限验证流程：
1. 用户登录后，调用 `GetUserInfo` 接口
2. 后端通过 `BuildMenuTreeAndPermissions` 构建菜单树和权限列表
3. `buildMenuTree` 方法根据角色的 `menuIDs` 过滤菜单
4. 前端收到过滤后的菜单树，动态生成路由

前端权限验证：
1. 路由守卫检查 `route.meta.roles` 和 `route.meta.permissions`
2. 如果路由没有 roles 和 permissions 字段，允许所有登录用户访问
3. 如果有，则检查用户角色或权限是否匹配

## 注意事项

1. **数据库已存在的角色不会自动更新**
   - 需要调用 `/api/route/invalidate-cache` 接口触发同步
   - 或者重启后端服务（启动时会自动同步）

2. **前端缓存**
   - 建议清除浏览器缓存或使用隐身模式测试
   - 或者退出登录后重新登录

3. **RBAC 缓存**
   - 系统会缓存用户的菜单树和权限列表（TTL: 1小时）
   - 修改权限后需要清除缓存

## 测试账号

- **admin**: admin/admin123 - 超级管理员，拥有所有权限
- **test**: test/123456 - 绑定 ops 角色，应只有系统管理权限
