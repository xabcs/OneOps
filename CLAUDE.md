# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

OneOps 是一个全栈运维管理平台，采用前后端分离架构：
- **后端**: Go + Gin + MySQL + JWT + RBAC
- **新前端**: `soybean-admin-element-plus/` - 基于 SoybeanAdmin（Vue 3 + Element Plus + TypeScript + Elegant Router）
- **旧前端**: `frontend/` - 传统 Vue 3 实现（已弃用，仅供参考）

## 常用命令

### 开发启动

```bash
# 同时启动前后端（推荐）
npm run dev

# 仅启动后端（使用Air热加载）
npm run dev:backend
# 等同于：cd backend && air

# 仅启动前端（Vite HMR）
npm run dev:frontend
# 等同于：cd frontend && npm run dev
```

### 后端开发

```bash
cd backend

# 代码检查（需要安装golangci-lint）
make lint

# 代码格式化（需要安装goimports）
make fmt

# 运行应用
make run
# 或
go run main.go

# 构建应用
make build
# 生成 bin/oneops 可执行文件

# 运行测试
make test
```

### 前端开发

**新前端（soybean-admin-element-plus）**：

```bash
cd soybean-admin-element-plus

# 安装依赖（需要 pnpm）
pnpm install

# 开发运行（测试环境）
pnpm dev

# 开发运行（生产环境）
pnpm dev:prod

# 生成路由（Elegant Router）
pnpm gen-route

# 代码检查
pnpm lint

# 类型检查
pnpm typecheck

# 构建生产版本
pnpm build

# 构建测试版本
pnpm build:test
```

**旧前端（frontend）**（仅供参考）：

```bash
cd frontend

# 代码检查
npm run lint

# 代码格式化
npm run format

# 修复lint问题
npm run lint:fix

# 构建生产版本
npm run build
```

## 架构关键点

### 后端架构

**配置管理**：
- 配置文件：`backend/config/config.yaml`（不提交到git）
- 配置模板：`backend/config/config.yaml.example`
- 加载方式：`config.LoadConfig()` 支持 YAML + 环境变量覆盖
- 环境变量可覆盖配置文件中的值（格式：`${ENV_VAR}`）

**日志系统**（Uber Zap）：
- 初始化：`logger.InitLogger(cfg.Log)`
- 使用方式：结构化日志，支持控制台（彩色）和文件（JSON）双输出
- 中间件：`logger.GinLogger()` 自动记录HTTP请求

**分层架构**：
```
routes（路由） → controllers（控制器） → services（业务逻辑） → models（数据模型）
```

**认证与权限**：
- JWT认证中间件：`middlewares.Auth()`
- RBAC权限控制：`services/rbac.go`
- 通配符权限：`*:*:*` 授予所有权限

**审计日志**：
- 操作日志自动记录（`middlewares/audit.go`）
- 三种日志类型：登录日志、操作日志、系统事件日志

### 前端架构

**⚠️ 重要：项目有两个前端目录**

| 目录 | 状态 | 技术栈 | 说明 |
|------|------|--------|------|
| `soybean-admin-element-plus/` | ✅ **当前使用** | SoybeanAdmin + Vue 3 + Element Plus + TypeScript + Elegant Router | 基于 [SoybeanAdmin](https://github.com/soybeanjs/soybean-admin) 模板，路由自动生成 |
| `frontend/` | ❌ 已弃用 | Vue 3 + Element Plus + Tailwind CSS | 旧版本实现，仅供参考 |

**新前端（soybean-admin-element-plus）架构**：

**路由系统**（Elegant Router）：
- 自动生成路由：启动项目后自动生成 `src/router/elegant/` 目录
- 核心类型：`RouteKey`（路由键）、`RoutePath`（路由路径）、`RouteMeta`（路由元信息）
- 新增页面：在 `src/views/` 创建文件，路由自动生成
- 官方文档：https://docs.soybeanjs.cn/zh/guide/router/intro.html

**状态管理**（Pinia）：
- 使用 Pinia 替代 Vuex
- 认证状态：token、user、permissions、menuTree

**路由守卫**：
- 认证检查：`meta.requiresAuth`
- 权限检查：`meta.permission`
- 进度条：NProgress

**环境配置**：
- 开发环境：`soybean-admin-element-plus/.env`
- 生产环境：`soybean-admin-element-plus/.env.prod`
- 测试环境：`soybean-admin-element-plus/.env.test`

**旧前端（frontend）架构**（仅供参考）：

**状态管理**（Vuex）：
- 认证状态：token、user、permissions、menuTree
- 持久化：使用localStorage，包含循环引用保护

**路由守卫**：
- 认证检查：`meta.requiresAuth`
- 权限检查：`meta.permission`
- 进度条：NProgress

**API配置**：
- 基础URL：使用环境变量 `VITE_API_BASE_URL`
- 开发环境：通过Vite代理到 `http://localhost:8082`
- 生产环境：使用相对路径 `/api`

**环境配置**：
- 开发环境：`frontend/.env.development`
- 生产环境：`frontend/.env.production`

## 热加载机制

### 后端（Air）
- 配置文件：`.air.toml`
- 监听文件：`.go`、`.yaml`、`.yml`
- 忽略目录：`tmp/`、`vendor/`、`frontend/`、`soybean-admin-element-plus/`、`node_modules/`、`logs/`
- 工作流程：检测文件变化 → 重新编译 → 重启服务

### 前端（Vite）

**新前端（soybean-admin-element-plus）**：
- 配置文件：`soybean-admin-element-plus/vite.config.ts`
- 使用 rolldown-vite（Vite 7）
- 路由自动生成：Elegant Router 插件
- 环境变量：`pnpm dev`（测试环境）、`pnpm dev:prod`（生产环境）

**旧前端（frontend）**：
- 配置文件：`frontend/vite.config.ts`
- HMR默认启用（可通过 `DISABLE_HMR` 环境变量禁用）
- API代理：开发环境自动代理 `/api` 到 `http://localhost:8082`

## 配置文件管理

### 后端配置优先级
1. 环境变量 `CONFIG_PATH` 指定的路径
2. `backend/config/config.yaml`
3. `/etc/oneops/config.yaml`
4. 默认值

### 环境变量覆盖
配置文件支持环境变量替换：
```yaml
database:
  password: "${DB_PASSWORD}"  # 从环境变量读取
jwt:
  secret: "${JWT_SECRET}"
```

### 敏感信息处理
- `config.yaml` 在 `.gitignore` 中（不提交）
- `config.yaml.example` 包含模板（可提交）
- 生产环境使用环境变量覆盖敏感配置

## 日志最佳实践

```go
// 简单日志
logger.Info("服务器启动成功")
logger.Error("数据库连接失败", zap.Error(err))

// 结构化日志
logger.Info("用户登录",
    zap.String("username", user.Username),
    zap.Uint("user_id", user.ID),
    zap.String("ip", c.ClientIP()))
```

## 默认账号

- 用户名：`admin`
- 密码：`123456`
- ⚠️ 生产环境必须修改默认密码和JWT密钥

## 数据库初始化

首次运行时，`services.InitService.InitDatabase()` 会自动：
1. 创建数据库表（如果不存在）
2. 创建默认管理员用户（admin/123456）
3. 初始化基础数据

## 代码规范

### Go代码
- 使用 `make lint` 检查代码质量
- 使用 `make fmt` 格式化代码
- 遵循 `golangci-lint` 规则（`.golangci.yml`）

### Vue代码
- 使用 `pnpm lint` 检查代码（新前端）
- 使用 `pnpm typecheck` 类型检查（新前端）
- 遵循 ESLint 和 TypeScript 规则（新前端）
- 使用 `npm run lint` 检查代码（旧前端）
- 使用 `npm run format` 格式化代码（旧前端）
- 遵循 ESLint（`.eslintrc.cjs`）和 Prettier（`.prettierrc`）规则（旧前端）

## SoybeanAdmin 核心知识

### 项目特性
- **前沿技术栈**: Vue 3 + Vite 5 + TypeScript + Pinia + UnoCSS
- **Monorepo 架构**: 使用 pnpm workspace 管理多包项目
- **代码规范**: ESLint + Prettier + simple-git-hooks
- **路由系统**: Elegant Router 自动生成路由
- **权限路由**: 支持前端静态路由和后端动态路由
- **国际化**: 内置 i18n 支持
- **主题系统**: 丰富的主题配置，与 UnoCSS 完美结合
- **移动端适配**: 响应式布局
- **命令行工具**: `sa` 命令用于 git 提交、删除文件、发布等

### 技术栈要求

**必须掌握的基础知识**：
- ES6+ JavaScript
- Vue 3（特别是 `<script-setup>` 语法）
- Vite 构建工具
- TypeScript 类型系统
- Vue Router 路由
- Pinia 状态管理
- UnoCSS 原子化 CSS
- VueUse 组合式工具库
- Element Plus 组件库

### 浏览器支持
- **开发推荐**: Chrome 100+
- **生产支持**: Edge/Firefox/Safari 最新 2 个版本
- **不支持**: IE 浏览器

### 目录结构

```
soybean-admin-element-plus/
├── src/
│   ├── router/
│   │   ├── elegant/          # 自动生成的路由文件（勿手动修改）
│   │   │   ├── imports.ts    # 路由导入
│   │   │   ├── routes.ts     # 路由定义
│   │   │   └── transform.ts  # 路由转换
│   │   ├── guard.ts          # 路由守卫
│   │   └── index.ts          # 路由配置
│   ├── views/                # 页面组件（根据此目录生成路由）
│   ├── store/                # Pinia 状态管理
│   ├── components/           # 通用组件
│   ├── layouts/              # 布局组件
│   ├── service/              # API 服务（基于 axios）
│   ├── utils/                # 工具函数
│   ├── locales/              # 国际化文件
│   ├── assets/               # 静态资源
│   ├── typings/              # TypeScript 类型定义
│   │   └── elegant-router.d.ts  # Elegant Router 自动生成的类型
│   └── main.ts               # 应用入口
├── packages/                 # pnpm workspace 包
├── build/                    # 构建配置
├── .env                      # 环境变量（测试）
├── .env.prod                 # 环境变量（生产）
└── vite.config.ts            # Vite 配置
```

### Elegant Router 路由系统

**路由模式**：

SoybeanAdmin 支持两种路由模式：

| 模式 | 路由来源 | 权限控制 | 适用场景 |
|------|----------|----------|----------|
| **静态模式** | 前端 `src/router/elegant/routes.ts` | 前端 `meta.roles` 硬编码 | 权限固定、快速开发 |
| **动态模式** | 后端 API `/api/route/getUserRoutes` | 后端根据数据库角色返回 | 权限灵活、完全由后端控制 |

**当前项目使用：动态路由模式**

**动态模式配置**：

1. **前端配置**（`.env`）：
   ```bash
   VITE_AUTH_ROUTE_MODE=dynamic
   ```

2. **后端接口**：
   - `GET /api/route/getConstantRoutes` - 返回常量路由（无需登录）
   - `GET /api/route/getUserRoutes` - 返回用户路由（根据用户角色）
   - `GET /api/route/isRouteExist` - 检查路由是否存在

3. **权限控制流程**：
   ```
   用户登录 → 获取用户角色 → 从数据库查询菜单权限 → 返回有权访问的路由
   ```

**数据库驱动权限**：
- 菜单数据存储在 `menus` 表
- 角色与菜单关联存储在 `roles.menu_ids` 字段（JSON 数组）
- 超级管理员（角色代码 `admin` 或 `R_SUPER`）自动拥有所有菜单权限

**初始化菜单数据**：
```bash
# 执行菜单初始化脚本
mysql -u root -p ops < backend/scripts/init_dynamic_menus.sql

# 执行角色菜单关联脚本
mysql -u root -p ops < backend/scripts/init_role_menus.sql
```

**核心概念**：

1. **自动生成机制**
   - 启动项目时，Elegant Router 插件会自动扫描 `src/views/` 目录
   - 根据页面文件自动生成路由配置到 `src/router/elegant/`
   - 自动生成 TypeScript 类型定义到 `src/typings/elegant-router.d.ts`

2. **RouteKey 类型**
   - 联合类型，包含所有路由的 key
   - 自动根据 views 目录下的页面文件生成
   - 用于类型安全的路由跳转

3. **RoutePath 类型**
   - 路由路径类型，与 RouteKey 一一对应
   - 确保路径字符串的类型安全

4. **RouteMeta 接口**
   ```typescript
   interface RouteMeta {
     title: string;                    // 路由标题
     i18nKey?: App.I18n.I18nKey;      // 国际化键值（优先于 title）
     roles?: string[];                 // 角色列表（权限控制）
     keepAlive?: boolean;              // 是否缓存路由
     constant?: boolean;               // 是否为常量路由（无需登录）
     icon?: string;                    // Iconify 图标
     localIcon?: string;               // 本地图标（src/assets/svg-icon）
     order?: number;                   // 路由排序
     href?: string;                    // 外部链接
     hideInMenu?: boolean;             // 是否在菜单中隐藏
     activeMenu?: RouteKey;            // 激活指定菜单项
     multiTab?: boolean;               // 是否支持多标签页
     fixedIndexInTab?: number;         // 标签页固定位置
     query?: { key: string; value: string }[];  // 自动携带的查询参数
   }
   ```

**新增页面流程**：

1. 在 `src/views/` 下创建页面文件
2. 启动项目，路由自动生成
3. 如需隐藏菜单，设置 `meta.hideInMenu: true`

**路由跳转示例**：
```typescript
// 使用 RouteKey 实现类型安全的跳转
import { useRouter } from 'vue-router';
import { RouteKey } from '@elegant-router/types';

const router = useRouter();
router.push(RouteKey.home);
```

### Pinia 状态管理

**核心 Store**：
- `authStore`: 认证状态（token、用户信息）
- `routeStore`: 路由状态（菜单、权限）
- `themeStore`: 主题状态
- `appStore`: 应用状态

**使用示例**：
```typescript
import { useAuthStore } from '@/store/modules/auth';

const authStore = useAuthStore();
authStore.setToken('xxx');
```

### 国际化 (i18n)

**语言文件位置**：`src/locales/`

**使用方式**：
```vue
<template>
  <!-- 在模板中使用 -->
  $t('key')
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const title = t('key');
</script>
```

### 主题系统

**主题配置**：`src/store/modules/theme.ts`

**支持的主题模式**：
- 亮色主题
- 暗色主题
- 自动跟随系统

**主题颜色**：通过 UnoCSS 和 CSS 变量配置

### API 服务

**服务位置**：`src/service/`

**请求方式**：基于 axios 的封装

**请求拦截器**：自动添加 token、处理错误

**使用示例**：
```typescript
import { fetchUserList } from '@/service/api';

const users = await fetchUserList();
```

### 常用命令

```bash
# 开发
pnpm dev              # 测试环境
pnpm dev:prod         # 生产环境

# 构建
pnpm build            # 生产构建
pnpm build:test       # 测试构建

# 代码质量
pnpm lint             # ESLint 检查并修复
pnpm typecheck        # TypeScript 类型检查

# 路由
pnpm gen-route        # 手动触发路由生成

# Git 操作（通过 sa 命令）
pnpm commit           # Git 提交
pnpm commit:zh        # Git 提交（中文）
```

### 代码规范

**Vue 3 语法**：
- 统一使用 `<script setup>` 语法
- 使用 Composition API

**命名规范**：
- 组件：PascalCase（如 `UserList.vue`）
- 文件夹：kebab-case（如 `user-management/`）
- 变量/函数：camelCase（如 `getUserList`）
- 常量：UPPER_SNAKE_CASE（如 `API_BASE_URL`）

**提交规范**：
- 使用 `pnpm commit` 遵循 Conventional Commits 规范
- 类型：feat, fix, docs, style, refactor, test, chore

### 组件使用

**Element Plus 组件**：
- 按需导入，自动注册
- 使用中文语言环境

**UnoCSS**：
- 原子化 CSS，类似 Tailwind CSS
- 配置文件：`uno.config.ts`

### 环境变量

```bash
# .env (测试环境)
VITE_SERVICE_BASE_URL=http://localhost:8082

# .env.prod (生产环境)
VITE_SERVICE_BASE_URL=https://api.example.com
```

**使用方式**：
```typescript
const apiUrl = import.meta.env.VITE_SERVICE_BASE_URL;
```

### 常见问题

**路由不生效**：
1. 检查页面文件是否在 `src/views/` 目录下
2. 运行 `pnpm gen-route` 重新生成路由
3. 重启开发服务器

**类型错误**：
1. 运行 `pnpm typecheck` 检查类型
2. 检查 `src/typings/elegant-router.d.ts` 是否已生成

**样式不生效**：
1. 检查 UnoCSS 类名是否正确
2. 查看浏览器开发者工具中的样式

**API 请求失败**：
1. 检查 `.env` 文件中的 `VITE_SERVICE_BASE_URL`
2. 确认后端服务正常运行
3. 查看网络请求的错误信息

## 故障排查

### 后端启动失败
1. 确认 `config.yaml` 存在（从 `config.yaml.example` 复制）
2. 检查MySQL服务运行状态
3. 验证数据库连接配置
4. 查看日志文件：`backend/logs/app.log`

### 前端API请求失败（新前端）
1. 确认后端服务正常运行（端口8082）
2. 检查 `soybean-admin-element-plus/.env` 中的 API 配置
3. 查看浏览器控制台错误信息

### 前端API请求失败（旧前端）
1. 确认后端服务正常运行（端口8082）
2. 检查 `frontend/.env.development` 中的 `VITE_API_BASE_URL`
3. 查看浏览器控制台错误信息

### 热加载不工作
- 后端：确认 Air 已安装（`go install github.com/air-verse/air@latest`）
- 前端：确认 `DISABLE_HMR` 环境变量未设置

## RBAC 权限系统说明

### 用户菜单树权限是什么

"菜单树权限"控制用户能看到哪些导航菜单、能访问哪些功能页面。

**菜单树**是左侧导航栏的完整层级结构，例如：
```
资产管理
├─ 资产总览
├─ 主机资产
├─ 访问控制
│  ├─ 访问策略
│  └─ 凭证库
└─ 会话审计
   ├─ 在线会话
   └─ 历史会话
```

**权限标识（permission string）** 在设计上用于按钮级别的细粒度控制（如 `cmdb:server:query`），但**目前在本项目中完全未被实际使用**：

- **后端**：没有任何中间件读取这些字符串做 API 鉴权，所有接口只校验 JWT 合法性
- **前端**：`useAuth()` / `hasAuth()` 函数存在但在所有业务页面中均未调用，只出现在框架演示页 `function/toggle-auth/index.vue`
- **`*:*:*`**：同样只是生成后存入 store，没有任何代码消费它

权限标识从用户有权访问的菜单记录的 `permission` 字段中提取（见 `rbac.go` 的 `extractPermissions`），随登录接口返回给前端存入 `authStore.userInfo.permissions`，但后续无人使用。

### 数据来源（三张表）

```
menus 表  → 存所有菜单（路径、图标、permission 标识）
roles 表  → 存角色，menu_ids 字段（JSON 数组）记录该角色可访问的菜单 ID
users 表  → 存用户，role_ids 字段（JSON 数组）记录该用户拥有的角色 ID
```

用户每次请求 `/api/route/getUserRoutes` 时，后端做三步：
1. `SELECT users WHERE id=?` → 拿到 `role_ids`
2. `SELECT roles WHERE id IN (...)` → 拿到每个角色的 `menu_ids`
3. `SELECT menus WHERE status=1` → 过滤出有权限的菜单，组装成树形结构

### 关键实现文件

| 文件 | 作用 |
|------|------|
| `backend/services/rbac.go` | `BuildMenuTreeAndPermissions` — 核心查询逻辑，带 5 分钟内存缓存 |
| `backend/services/rbac_cache.go` | RBAC 缓存定义，`InvalidateRBACCache(userID)` 用于失效 |
| `backend/controllers/route.go` | `GetUserRoutes` — 调用 RBAC 服务，将菜单转换为前端路由格式 |
| `backend/services/init.go` | `syncMenus` / `syncRoleMenus` — 启动时自动同步菜单和角色数据 |

### 缓存策略

菜单树结果缓存在内存中，TTL 为 5 分钟（`rbacCacheTTL`）。

- 菜单/角色变更后，调用 `InvalidateRBACCache(0)` 清除所有用户缓存
- 用户登录时（`auth.go`）会预热该用户的缓存，后续 `getUserRoutes` 直接命中（~10ms）
- 不缓存时每次需 3 次远程 DB 查询（~300ms）

## 权限系统设计原则

### 当前采用的方案：两层权限 + 审计

本项目是内部运维平台，权限设计遵循"够用即止"原则：

**第一层：功能权限（菜单 RBAC）**
- 控制"你能看到哪些页面"，由角色绑定菜单 ID 实现
- 已生效，无需额外开发

**第二层：连接权限（访问策略表）**
- 控制"你能 SSH 进哪台服务器"，由 `asset_access_policies` 表实现
- 校验逻辑在 `backend/services/bastion.go` 的 `CheckConnectPermission`
- 管理员（角色 ID=1）直接放行；其他角色需匹配策略表

**审计作为核心安全机制**
- 内部平台的安全依赖"知道自己被审计"的约束，而非复杂权限拦截
- 每次连接、每条命令全部关联用户 + 主机 + 会话 ID

### 审批流为什么暂不规划

审批功能（`bastion_approvals` 表 + 审批记录页 + 审批状态校验）开发成本较高，对当前阶段不必要：

**替代方案：访问策略 `time_window` 字段**
- 需要临时授权时，管理员直接在访问策略表中设置时间窗口（`time_window`），到期自动失效
- 和审批流的效果等价：控制"谁在什么时间段内能连哪台机器"
- 无需额外的审批申请/通知/审批通过/过期清理等状态机

**当前连接权限控制流程（代替审批）：**
```text
用户申请访问 → 管理员在访问策略表中新增或调整策略（设置 time_window）
→ 用户在窗口期内正常连接 → 过期后策略不再命中，自动拒绝
```

**审批流何时再规划（P2+）：**
- 团队规模较大，管理员无法及时手动调整策略
- 有合规要求，需要留存完整的申请/审批/过期审计链路
- 需要申请人自助提交，不依赖管理员操作

### 权限点体系（cmdb:server:connect 等）为什么暂不实现

权限点体系需要**前后端四个环节同时配合**：

```
① 后端维护独立的按钮权限码数据源（区别于菜单 permission 字段）
        ↓
② 登录接口返回 buttons 数组（区别于 permissions 字段）
        ↓
③ 前端 hasAuth('cmdb:server:connect') 控制按钮显示
        ↓
④ 后端接口中间件校验用户是否持有该权限码
```

**当前未实现的原因：**
- 后端没有独立按钮权限数据源，权限字符串全来自菜单 `permission` 字段
- 后端返回 `permissions` 字段，前端 `hasAuth()` 读 `buttons` 字段，字段名不对应
- 所有业务页面均未调用 `hasAuth()`，权限字符串生成后无人消费
- 对内部平台而言，维护这四个环节的同步成本高于收益

**前端连接按钮的正确做法：**
```vue
<!-- 状态由后端 CheckConnectPermission 接口返回值驱动，不用 hasAuth -->
<el-button
  :disabled="!server.credentials?.length"
  @click="handleConnect"
>
  连接
</el-button>
```

### 未来（P2+）才考虑权限点体系的场景

- 平台用户规模超过 500 人
- 有外部合规审查要求（等保、SOC2 等）
- 出现多租户隔离需求
- 需要按按钮级别差异化不同角色的操作能力

届时需要同步改造：后端按钮权限表 + 登录接口 + 前端 store 类型定义 + 各业务页面 `hasAuth` 调用点 + 后端接口校验中间件。
