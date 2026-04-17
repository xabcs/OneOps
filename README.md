# OneOps

全栈运维管理平台，用于服务器监控、任务自动化和系统管理。

## 技术栈

### 前端
- **框架**: Vue 3（组合式 API）
- **UI 组件**: Element Plus（中文语言环境）
- **样式**: Tailwind CSS
- **构建工具**: Vite
- **状态管理**: Vuex
- **路由**: Vue Router
- **代码规范**: ESLint + Prettier

### 后端
- **语言**: Go
- **框架**: Gin
- **数据库**: MySQL
- **认证**: JWT
- **权限控制**: RBAC（基于角色的访问控制）
- **日志系统**: Uber Zap
- **配置管理**: YAML

## 快速开始

### 前置要求
- Node.js 18+
- Go 1.25+
- MySQL 8.0+
- Air（后端热加载，可选）

### 安装依赖

```bash
# 安装根依赖
npm install

# 安装前端依赖
cd frontend
npm install
```

### 配置

#### 后端配置

1. **配置文件**：
   - 复制 `backend/config/config.yaml.example` 为 `backend/config/config.yaml`
   - 根据需要修改配置

2. **环境变量**（可选）：
   - 复制 `backend/.env.example` 为 `backend/.env`
   - 设置环境变量覆盖配置文件

3. **数据库配置**：
   ```yaml
   database:
     host: "localhost"
     port: "3306"
     user: "root"
     password: "your_password"
     dbname: "oneops"
   ```

#### 前端配置

前端使用环境变量配置，已预配置：
- 开发环境：`frontend/.env.development`
- 生产环境：`frontend/.env.production`

### 运行

```bash
# 同时启动前后端（推荐）
npm run dev

# 或分别启动
npm run dev:backend   # Go 后端（端口 8082）
npm run dev:frontend  # Vue 前端（端口 5173）
```

### 构建

```bash
npm run build
```

## 项目结构

```
OneOps/
├── backend/           # Go 后端
│   ├── config/        # 配置管理
│   │   ├── config.go           # 配置结构
│   │   ├── config.yaml         # 配置文件（本地）
│   │   ├── config.yaml.example # 配置模板
│   │   └── loader.go           # 配置加载器
│   ├── logger/        # 日志系统
│   │   ├── logger.go          # 日志初始化
│   │   └── middleware.go       # Gin日志中间件
│   ├── models/        # 数据模型
│   ├── controllers/   # 控制器
│   ├── services/      # 业务逻辑
│   ├── middlewares/   # 中间件
│   ├── routes/        # 路由配置
│   ├── utils/         # 工具函数
│   ├── main.go        # 入口文件
│   ├── Makefile       # 构建脚本
│   └── .golangci.yml  # 代码检查配置
├── frontend/          # Vue 3 前端
│   ├── src/
│   │   ├── api/       # API 客户端
│   │   ├── router/    # Vue Router
│   │   ├── store/     # Vuex 状态管理
│   │   ├── views/     # 页面组件
│   │   └── mock/      # Mock 数据
│   ├── .eslintrc.cjs  # ESLint配置
│   ├── .prettierrc    # Prettier配置
│   └── vite.config.ts # Vite 配置
├── .air.toml          # Air热加载配置
└── package.json       # 根配置
```

## 配置说明

### 后端配置（config.yaml）

```yaml
server:
  port: "8082"              # 监听端口
  mode: "debug"             # 运行模式: debug, release, test

database:
  host: "localhost"
  port: "3306"
  user: "root"
  password: "your_password"
  dbname: "oneops"

jwt:
  secret: "your_jwt_secret" # 生产环境必须修改
  expire_time: 24           # 过期时间（小时）

log:
  level: "debug"            # 日志级别
  filename: "logs/app.log"  # 日志文件路径

cors:
  allow_origins:
    - "http://localhost:5173"
```

### 环境变量

支持环境变量覆盖配置文件：

```bash
# 服务器
SERVER_PORT=8082
GIN_MODE=debug

# 数据库
DB_HOST=localhost
DB_PASSWORD=your_password

# JWT
JWT_SECRET=your_jwt_secret
```

## 日志系统

### 日志级别
- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息
- `fatal`: 致命错误

### 日志输出
- **控制台**: 彩色格式化输出（开发环境）
- **文件**: JSON格式（生产环境）

### 日志示例

```go
// 简单日志
logger.Info("服务器启动成功")
logger.Error("数据库连接失败", zap.Error(err))

// 结构化日志
logger.Info("用户登录",
    zap.String("username", "admin"),
    zap.Uint("user_id", 1),
    zap.String("ip", c.ClientIP()))
```

## 开发工具

### 后端

```bash
# 代码检查
make lint

# 代码格式化
make fmt

# 运行测试
make test

# 构建应用
make build
```

### 前端

```bash
# 代码检查
npm run lint

# 代码格式化
npm run format

# 修复问题
npm run lint:fix
```

## 默认账号

- 用户名：`admin`
- 密码：`123456`

**⚠️ 重要**: 生产环境请立即修改默认密码！

## 功能特性

- ✅ 用户管理
- ✅ 角色管理
- ✅ 菜单管理
- ✅ 审计日志
- ✅ JWT 认证
- ✅ RBAC 权限控制
- ✅ 结构化日志
- ✅ 配置文件管理
- ✅ 热加载开发
- ✅ 代码规范检查

## 安全建议

1. **修改默认密码**: 生产环境必须修改默认密码
2. **强JWT密钥**: 使用至少32位的随机字符串
3. **配置文件**: 不要将config.yaml提交到版本控制
4. **环境变量**: 敏感信息使用环境变量
5. **CORS配置**: 生产环境限制允许的源

## 故障排查

### 后端启动失败

1. **配置文件不存在**:
   ```bash
   cp backend/config/config.yaml.example backend/config/config.yaml
   ```

2. **数据库连接失败**:
   - 检查MySQL服务是否运行
   - 验证config.yaml中的数据库配置
   - 确保数据库用户有足够权限

3. **端口被占用**:
   ```bash
   # 查看端口占用
   lsof -i :8082
   # 或修改config.yaml中的port
   ```

### 前端问题

1. **API请求失败**:
   - 确认后端服务正常运行
   - 检查前端.env配置中的API_BASE_URL
   - 查看浏览器控制台错误信息

2. **依赖安装失败**:
   ```bash
   # 清除缓存重试
   rm -rf node_modules package-lock.json
   npm install
   ```

## 开发

详细的开发指南请参考 [DEVELOPMENT.md](DEVELOPMENT.md)

## 许可证

MIT
