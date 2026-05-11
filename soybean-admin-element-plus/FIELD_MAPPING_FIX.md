# 字段名修复说明

## 问题

在对接后端 API 后，前端应用出现了路由导航错误：

```
TypeError: Cannot read properties of undefined (reading 'some')
    at route.ts:31:46
```

## 原因

后端返回的用户信息结构与前端模板中的字段名不匹配：

| 前端模板字段名 | 后端返回字段名 | 说明 |
|--------------|--------------|------|
| `userId`     | `id`         | 用户ID |
| `userName`   | `username`   | 用户名 |
| `roles`      | `roleNames`  | 角色名称列表 |

## 修复内容

### 1. 路由守卫 (`src/router/guard/route.ts`)

```typescript
// 修复前
const hasRole = authStore.userInfo.roles.some(role => routeRoles.includes(role));

// 修复后
const hasRole = authStore.userInfo.roleNames.some(role => routeRoles.includes(role));
```

### 2. 路由状态管理 (`src/store/modules/route/index.ts`)

```typescript
// 修复前
if (!authStore.userInfo.userId) {

// 修复后
if (!authStore.userInfo.id) {
```

以及角色过滤：
```typescript
// 修复前
const filteredAuthRoutes = filterAuthRoutesByRoles(staticAuthRoutes, authStore.userInfo.roles);

// 修复后
const filteredAuthRoutes = filterAuthRoutesByRoles(staticAuthRoutes, authStore.userInfo.roleNames);
```

### 3. 用户界面组件

修复了以下文件中的用户名显示，优先使用 `nickname`，回退到 `username`：

- `src/App.vue`：水印用户名
- `src/views/home/modules/header-banner.vue`：欢迎消息
- `src/layouts/modules/global-header/components/user-avatar.vue`：用户头像显示

```typescript
// 修复前
authStore.userInfo.userName

// 修复后
authStore.userInfo.nickname || authStore.userInfo.username
```

### 4. 角色显示 (`src/views/function/toggle-auth/index.vue`)

```typescript
// 修复前
<ElTag v-for="role in authStore.userInfo.roles" :key="role">{{ role }}</ElTag>

// 修复后
<ElTag v-for="role in authStore.userInfo.roleNames" :key="role">{{ role }}</ElTag>
```

## 用户信息结构对比

### 修复前（前端模板）
```typescript
interface UserInfo {
  userId: string;
  userName: string;
  roles: string[];
  buttons: string[];
}
```

### 修复后（匹配后端）
```typescript
interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  roleNames: string[];
  // ... 其他字段
}
```

## 测试结果

修复后，前端应用应该能够：
1. ✅ 正常加载页面
2. ✅ 登录后正确跳转
3. ✅ 显示用户信息和角色
4. ✅ 路由权限检查正常工作

## 注意事项

在后续开发中，如果需要添加新的用户相关功能，请使用正确的字段名：
- 用户ID：`userInfo.id`
- 用户名：`userInfo.username`
- 昵称：`userInfo.nickname`（用于显示）
- 角色列表：`userInfo.roleNames`
- 权限列表：`userInfo.permissions`
