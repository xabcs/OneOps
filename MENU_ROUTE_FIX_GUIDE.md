# API权限管理菜单路由错误修复

## 问题描述

前端报错：
```
Error: No match for {"name":"manage_apipermission","params":{}}
```

## 问题原因

- **前端路由名称**：`manage_api-permission`（带连字符）
- **数据库菜单配置**：`manage_apipermission`（无连字符）

两者不匹配导致路由跳转失败。

## 修复方法

### 方法1：直接执行SQL（推荐）

```sql
-- 连接数据库
mysql -h 60.191.116.75 -P 38089 -u root -p ops

-- 执行修复SQL
UPDATE menus
SET route_name = 'manage_api-permission'
WHERE route_name = 'manage_apipermission';

-- 验证结果
SELECT id, name, route_name, path
FROM menus
WHERE route_name = 'manage_api-permission';
```

### 方法2：使用提供的SQL文件

```bash
mysql -h 60.191.116.75 -P 38089 -u root -p ops < fix_api_permission_menu.sql
```

### 方法3：运行Go修复程序

```bash
# 1. 修改数据库密码
vim backend/fix_menu_route.go
# 修改 dsn 中的 YourPassword 为实际密码

# 2. 运行修复程序
cd backend
go run fix_menu_route.go
```

## 验证修复

修复完成后：

1. **清除浏览器缓存** 或 **重新登录**
2. 点击菜单中的"API权限管理"
3. 应该可以正常跳转到 `/manage/api-permission` 页面

## 前端路由配置

正确的路由配置（已在项目中）：

```typescript
// frontend/src/router/elegant/routes.ts
{
  name: 'manage_api-permission',
  path: '/manage/api-permission',
  component: 'view.manage_api-permission',
  meta: {
    title: 'manage_api-permission',
    i18nKey: 'route.manage_api-permission'
  }
}
```

## 数据库正确配置

菜单表应该包含：

| 字段 | 值 |
|------|-----|
| name | API权限管理 |
| route_name | **manage_api-permission** |
| path | /manage/api-permission |
| icon | api |
| menu_type | menu |
| resource | api-permission |

## 预防措施

为避免类似问题：

1. **统一命名规范**：前端路由和数据库菜单使用相同的命名格式
2. **菜单管理界面**：在创建菜单时添加路由名称验证
3. **测试检查**：新菜单创建后立即测试跳转功能

## 相关文件

- SQL修复脚本：`fix_api_permission_menu.sql`
- Shell修复脚本：`fix_menu.sh`
- Go修复程序：`backend/fix_menu_route.go`
- 前端页面：`frontend/src/views/manage/api-permission/index.vue`
- 前端路由：`frontend/src/router/elegant/routes.ts`
