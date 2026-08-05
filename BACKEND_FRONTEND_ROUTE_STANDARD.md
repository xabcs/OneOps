# 后端路由命名规范与前端对齐指南

## 一、后端路由命名规范（标准）

### 1. URL路径命名规则

**规则：使用kebab-case（连字符命名法）**

```
✅ 正确示例：
GET /api/system/users                    # 单个单词
GET /api/system/api-resources            # 多个单词用连字符
GET /api/system/server-attributes/:id    # 多个单词用连字符
GET /api/system/user-roles               # 多个单词用连字符

❌ 错误示例：
GET /api/system/apiResources             # camelCase
GET /api/system/api_resources            # snake_case
GET /api/system/APIRESOURCES             # 全大写
```

### 2. 后端路由配置示例

```go
// backend/routes/routes.go

// ✅ 正确：使用连字符
system.GET("/casbin/api-resources", ...)
system.GET("/server-attributes/:serverId", ...)
system.POST("/api-permissions/assign", ...)

// ✅ 正确：单个单词用复数
system.GET("/users", ...)
system.GET("/roles", ...)
system.GET("/menus", ...)
```

## 二、前端路由命名规范（遵循后端）

### 1. 路由名称生成规则

**规则：从后端path转换为前端route name**

转换公式：
```
path: /manage/api-permission
↓ 去掉开头的 /
↓ 将 / 替换为 _
route name: manage_api-permission
```

### 2. 前端路由配置示例

```typescript
// frontend/src/router/elegant/routes.ts

{
  // ✅ 正确：从path转换而来
  name: 'manage_api-permission',  // /manage/api-permission → manage_api-permission
  path: '/manage/api-permission',
  component: 'view.manage_api-permission',
}

{
  // ✅ 正确：单个单词
  name: 'manage_user',  // /manage/user → manage_user
  path: '/manage/user',
}
```

## 三、当前问题分析

### 问题描述

前端报错找不到路由：`manage_apipermission`

### 原因

这个路由名称 `manage_apipermission` 不符合规范：
- ❌ 错误：`manage_apipermission`（没有连字符）
- ✅ 正确：`manage_api-permission`（保留连字符）

### 可能的来源

1. **浏览器缓存**：缓存了旧的路由状态
2. **手动配置错误**：某处硬编码了错误的路由名称
3. **动态生成错误**：代码从path生成route name时错误地移除了连字符

## 四、解决方案

### 立即修复步骤

#### 步骤1：检查数据库菜单配置

```bash
# 执行SQL检查
mysql -h 60.191.116.75 -P 38089 -u root -p123456 msre < check_and_fix_menu.sql
```

#### 步骤2：清除浏览器缓存

```
用户操作：
1. 清除浏览器缓存
2. 或使用无痕模式
3. 或重新登录
```

#### 步骤3：验证前端路由配置

```bash
# 确认transform.ts中的路由映射正确
grep "manage.*api.*permission" frontend/src/router/elegant/transform.ts

# 应该输出：
# "manage_api-permission": "/manage/api-permission",
```

### 长期规范

#### 1. 后端开发规范

创建新的API路由时：
```go
// ✅ 正确
r.GET("/api/system/user-profiles", handler)     // 多个单词用连字符
r.GET("/api/system/audit-logs", handler)        // 多个单词用连字符
r.GET("/api/users", handler)                    // 单个单词用复数

// ❌ 错误
r.GET("/api/system/userProfiles", handler)      // camelCase
r.GET("/api/system/user_profiles", handler)     // snake_case
```

#### 2. 前端开发规范

创建新的页面时：
```typescript
// 1. 创建页面组件
// frontend/src/views/manage/user-profile/index.vue

// 2. 在 routes.ts 中添加路由（自动生成或手动配置）
{
  name: 'manage_user-profile',  // 从 path /manage/user-profile 转换
  path: '/manage/user-profile',
  component: 'view.manage_user-profile',
}

// 3. 在 transform.ts 中确认映射
"user-profile": "/manage/user-profile",
"manage_user-profile": "/manage/user-profile",
```

## 五、路由命名对照表

| 后端path | 前端route name | 说明 |
|---------|---------------|------|
| `/users` | `manage_user` | 单个单词 |
| `/roles` | `manage_role` | 单个单词 |
| `/menus` | `manage_menu` | 单个单词 |
| `/api-permission` | `manage_api-permission` | 多个单词保留连字符 |
| `/user-detail/:id` | `manage_user-detail` | 多个单词保留连字符 |
| `/server-attributes` | `manage_server-attributes` | 多个单词保留连字符 |

## 六、检查清单

开发新的管理页面时，确认：

- [ ] 后端API路径使用连字符：`/api/system/my-resource`
- [ ] 前端页面路径使用连字符：`/manage/my-resource`
- [ ] 前端路由名称保留连字符：`manage_my-resource`
- [ ] 数据库菜单path字段使用连字符：`/manage/my-resource`
- [ ] 测试菜单跳转是否正常

## 七、故障排查

如果菜单跳转失败：

```bash
# 1. 检查前端路由配置
cat frontend/src/router/elegant/transform.ts | grep "路由名称"

# 2. 检查数据库菜单配置
mysql -e "SELECT name, path FROM menus WHERE name LIKE '%页面名%'"

# 3. 检查浏览器控制台报错
# 打开开发者工具 → Console

# 4. 清除所有缓存并重新登录
```

## 八、总结

### 核心原则

**后端定义规范，前端遵循后端**

- 后端API路径使用 **kebab-case**（连字符）
- 前端路由名称从path转换，**保留连字符**
- 前端不创建新的命名规则，只做转换

### 当前状态

- ✅ 后端路由规范：正确使用连字符
- ✅ 前端路由配置：正确使用 `manage_api-permission`
- ✅ 数据库菜单：路径正确 `/manage/api-permission`
- ❌ 报错路由：`manage_apipermission`（来源不明，可能是缓存）

### 下一步

1. 清除浏览器缓存
2. 重新登录测试
3. 如果仍有问题，检查是否有代码硬编码了错误的路由名称
