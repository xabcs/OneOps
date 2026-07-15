# Soybean Admin 与后端对接文档

## 概述

本文档记录了 soybean-admin-element-plus 前端项目与 OneOps 后端（Go + Gin）的对接过程。

## 对接状态

✅ **登录功能已成功对接**

## 配置修改

### 1. 环境配置文件

**`.env` 文件：**
```bash
# 后端服务地址
VITE_SERVICE_BASE_URL=http://localhost:8082/api

# 成功响应码
VITE_SERVICE_SUCCESS_CODE=200
```

**`.env.test` 文件：**
```bash
# 后端服务地址（测试环境）
VITE_SERVICE_BASE_URL=http://localhost:8082/api
```

### 2. API 接口修改

**`src/service/api/auth.ts`：**
- 登录接口：`/auth/login` → `/login`
- 用户信息接口：`/auth/getUserInfo` → `/user/info`
- 请求参数：`userName` → `username`

### 3. 类型定义修改

**`src/typings/api/auth.d.ts`：**
```typescript
interface LoginToken {
  token: string;
  user: UserInfo;  // 后端在登录响应中直接返回用户信息
}

interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  avatar?: string;
  email?: string;
  roleIds?: number[];
  status?: string | number;
  homePath?: string;
  createdAt?: string;
  updatedAt?: string;
  roleNames: string[];
  menuTree: any[];
  permissions: string[];
}
```

### 4. 状态管理修改

**`src/store/modules/auth/index.ts`：**
- 更新 userInfo 结构以匹配后端返回格式
- 修改 `loginByToken` 函数，直接使用登录响应中的用户信息
- 更新相关字段引用（userId → id，userName → username，roles → roleNames）

**字段名映射修复：**
- `userId` → `id`
- `userName` → `username`（显示时优先使用 `nickname`）
- `roles` → `roleNames`

修复了以下文件中的字段引用：
- `src/router/guard/route.ts`：路由权限检查
- `src/store/modules/route/index.ts`：路由初始化
- `src/views/function/toggle-auth/index.vue`：角色显示
- `src/App.vue`：水印用户名
- `src/views/home/modules/header-banner.vue`：欢迎消息
- `src/layouts/modules/global-header/components/user-avatar.vue`：用户头像显示

### 5. 登录页面修改

**`src/views/_builtin/login/modules/pwd-login.vue`：**
- 默认账号改为：`admin` / `123456`（匹配后端默认账号）

## 后端接口规范

### 登录接口

**请求：**
```http
POST /api/login
Content-Type: application/json

{
  "username": "admin",
  "password": "123456"
}
```

**响应：**
```json
{
  "code": 200,
  "success": true,
  "data": {
    "token": "eyJhbGci...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "超级管理员",
      "email": "admin@example.com",
      "roleNames": ["超级管理员"],
      "menuTree": [...],
      "permissions": ["*:*:*", ...]
    }
  },
  "message": "登录成功"
}
```

### 获取用户信息接口

**请求：**
```http
GET /api/user/info
Authorization: Bearer {token}
```

**响应：**
```json
{
  "code": 200,
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "nickname": "超级管理员",
    "roleNames": ["超级管理员"],
    "menuTree": [...],
    "permissions": ["*:*:*", ...]
  },
  "message": "success"
}
```

## 运行说明

### 启动后端服务

```bash
# 从项目根目录
npm run dev:backend
# 或
cd backend && air
```

后端服务运行在：`http://localhost:8082`

### 启动前端服务

```bash
# 从项目根目录
pnpm dev
```

前端服务运行在：`http://localhost:9527`

### 访问应用

1. 打开浏览器访问：`http://localhost:9527`
2. 使用默认账号登录：
   - 用户名：`admin`
   - 密码：`123456`

## 代理配置

在开发模式下，前端使用 Vite 代理将 API 请求转发到后端：

- 前端请求路径：`/proxy-default/*`
- 代理到后端：`http://localhost:8082/api/*`

代理配置在 `vite.config.ts` 中自动处理，无需手动配置。

## 测试

### 测试登录 API

```bash
# 直接测试后端
curl -X POST http://localhost:8082/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'

# 通过前端代理测试
curl -X POST http://localhost:9527/proxy-default/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

### 测试用户信息 API

```bash
# 首先登录获取 token
TOKEN=$(curl -s -X POST http://localhost:8082/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}' \
  | jq -r '.data.token')

# 使用 token 获取用户信息
curl -X GET http://localhost:8082/api/user/info \
  -H "Authorization: Bearer $TOKEN"
```

## 后续工作

### 已完成
- ✅ 登录接口对接
- ✅ 用户信息接口对接
- ✅ 环境配置
- ✅ 类型定义
- ✅ 状态管理

### 待对接
- ⏳ 菜单管理接口
- ⏳ 角色管理接口
- ⏳ 用户管理接口
- ⏳ 审计日志接口
- ⏳ 监控管理接口

### 建议优先级
1. **路由菜单**：将后端返回的 `menuTree` 集成到前端路由系统
2. **权限控制**：实现基于 `permissions` 的按钮级权限控制
3. **系统管理**：对接用户、角色、菜单管理功能
4. **审计日志**：对接登录日志和操作日志

## 注意事项

1. **Token 存储**：前端使用 localStorage 存储 token，key 为 `token`
2. **认证方式**：使用 Bearer Token 认证，格式：`Authorization: Bearer {token}`
3. **响应格式**：后端统一响应格式为 `{ code, success, data, message }`
4. **成功码**：使用 `code: 200` 表示成功
5. **环境变量**：确保 `.env.test` 文件中的配置正确

## 故障排查

### 登录后无法跳转
- 检查浏览器控制台是否有错误
- 确认后端服务正常运行
- 检查 token 是否正确存储在 localStorage

### API 请求失败
- 检查代理配置是否正确
- 确认后端服务地址和端口
- 查看浏览器网络面板的请求详情

### 菜单不显示
- 检查 `menuTree` 数据格式
- 确认路由配置是否正确
- 查看前端路由守卫逻辑

## 参考资料

- [Soybean Admin 文档](https://doc.soybeanjs.cn/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [Vue Router 文档](https://router.vuejs.org/)
- [Pinia 文档](https://pinia.vuejs.org/)
