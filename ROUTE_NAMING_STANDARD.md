# 后端路由命名规范

## 当前后端路由分析

### 系统管理模块 (system)

```go
// 单个单词（复数形式）
GET    /api/system/menus
GET    /api/system/roles
GET    /api/system/users

// 多个单词（连字符连接）
GET    /api/system/casbin/api-resources
GET    /api/system/server-attributes/:serverId
GET    /api/system/api-permissions
```

### 后端路由命名规则（标准）

1. **单个单词**：使用复数形式
   - `/menus`, `/roles`, `/users`, `/attributes`

2. **多个单词**：使用连字符连接（kebab-case）
   - `/api-resources`, `/server-attributes`, `/api-permissions`

3. **资源路径**：
   - 列表：`GET /resources`
   - 详情：`GET /resources/:id`
   - 创建：`POST /resources`
   - 更新：`PUT /resources/:id`
   - 删除：`DELETE /resources/:id`

## 前端路由命名规范（应以Backend为准）

### 当前前端路由

```typescript
// 使用下划线连接（不符合后端规范）
manage_menu
manage_role
manage_user
manage_permission

// 使用连字符连接（符合后端规范）
manage_api-permission  ✅
```

### 统一规范（前端遵循后端）

**规则：前端路由名称 = 模块名_资源名（使用下划线）**

转换规则：
- 后端路径 `/api/system/user-roles` → 前端路由 `manage_user-roles`
- 后端路径 `/api/system/api-resources` → 前端路由 `manage_api-resources`

## 需要统一的路由

### 1. API权限管理路由

**后端路由**：
```
GET    /api/system/casbin/api-resources
POST   /api/system/casbin/assign
```

**前端路由（已正确）**：
```typescript
{
  name: 'manage_api-permission',  // ✅ 使用连字符
  path: '/manage/api-permission',
}
```

### 2. 其他管理路由（需要统一）

**建议修改前端路由为**：
```typescript
manage_menu          → 保持不变（单个单词）
manage_role          → 保持不变（单个单词）
manage_user          → 保持不变（单个单词）
manage_permission    → 保持不变（单个单词）
manage_api-permission → ✅ 已符合规范（多个单词用连字符）
manage_user-detail   → ✅ 已符合规范
```

## 总结

### ✅ 后端路由规范（标准）
- 单个单词：复数形式 `/users`, `/roles`, `/menus`
- 多个单词：连字符连接 `/api-resources`, `/user-roles`

### ✅ 前端路由规范（遵循后端）
- 格式：`模块名_资源名`
- 单个单词：`manage_user`, `manage_role`
- 多个单词：`manage_api-permission`, `manage_user-detail`

### 当前问题

数据库菜单表中的 `route_name` 字段错误：
- ❌ 错误：`manage_apipermission`（缺少连字符）
- ✅ 正确：`manage_api-permission`（符合后端规范）

## 修复SQL

```sql
-- 修复数据库菜单路由名称
UPDATE menus
SET route_name = 'manage_api-permission'
WHERE route_name = 'manage_apipermission';
```
