# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

OneOps 是一个全栈运维管理平台，采用前后端分离架构：
- **后端**: Go + Gin + MySQL + JWT + RBAC
- **前端**: Vue 3 + Element Plus + Tailwind CSS + Vite

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
- 忽略目录：`tmp/`、`vendor/`、`frontend/`、`node_modules/`、`logs/`
- 工作流程：检测文件变化 → 重新编译 → 重启服务

### 前端（Vite）
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
- 使用 `npm run lint` 检查代码
- 使用 `npm run format` 格式化代码
- 遵循 ESLint（`.eslintrc.cjs`）和 Prettier（`.prettierrc`）规则

## 故障排查

### 后端启动失败
1. 确认 `config.yaml` 存在（从 `config.yaml.example` 复制）
2. 检查MySQL服务运行状态
3. 验证数据库连接配置
4. 查看日志文件：`backend/logs/app.log`

### 前端API请求失败
1. 确认后端服务正常运行（端口8082）
2. 检查 `frontend/.env.development` 中的 `VITE_API_BASE_URL`
3. 查看浏览器控制台错误信息

### 热加载不工作
- 后端：确认 Air 已安装（`go install github.com/air-verse/air@latest`）
- 前端：确认 `DISABLE_HMR` 环境变量未设置
