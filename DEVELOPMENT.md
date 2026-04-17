# OneOps 开发指南

本文档介绍 OneOps 项目的开发环境搭建、代码规范和最佳实践。

## 目录

- [开发环境搭建](#开发环境搭建)
- [代码规范](#代码规范)
- [Git工作流](#git工作流)
- [调试技巧](#调试技巧)
- [常见问题](#常见问题)

## 开发环境搭建

### 必需工具

1. **Go 1.25+**
   ```bash
   # macOS
   brew install go

   # 验证安装
   go version
   ```

2. **Node.js 18+**
   ```bash
   # macOS
   brew install node

   # 验证安装
   node --version
   npm --version
   ```

3. **MySQL 8.0+**
   ```bash
   # macOS
   brew install mysql
   brew services start mysql

   # 创建数据库
   mysql -u root -p
   CREATE DATABASE oneops CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

### 推荐工具

1. **Air - Go热加载**
   ```bash
   go install github.com/air-verse/air@latest
   ```

2. **golangci-lint - 代码检查**
   ```bash
   # macOS
   brew install golangci-lint
   ```

3. **goimports - import排序**
   ```bash
   go install golang.org/x/tools/cmd/goimports@latest
   ```

4. **VSCode插件**
   - Vue - Official
   - ESLint
   - Prettier
   - Go

### 初始化项目

```bash
# 1. 克隆项目
git clone <repository-url>
cd OneOps

# 2. 安装依赖
npm install
cd frontend && npm install && cd ..

# 3. 配置后端
cp backend/config/config.yaml.example backend/config/config.yaml
# 编辑config.yaml设置数据库连接

# 4. 启动开发服务器
npm run dev
```

## 代码规范

### Go代码规范

#### 命名规范

```go
// 包名：小写单词
package controllers

// 常量：大驼峰或全大写
const MaxRetry = 3
const API_KEY = "xxx"

// 变量/函数：小驼峰
var userCount int
func getUserInfo() {}

// 私有变量/函数：小驼峰（首字母小写）
var internalCount int
func internalHelper() {}

// 公开变量/函数：大驼峰（首字母大写）
var PublicCount int
func PublicHelper() {}
```

#### 错误处理

```go
// ✅ 好的做法
result, err := someFunction()
if err != nil {
    logger.Error("操作失败", zap.Error(err))
    return nil, err
}

// ❌ 不好的做法
result, _ := someFunction() // 忽略错误
```

#### 日志规范

```go
// 使用结构化日志
logger.Info("用户登录",
    zap.String("username", user.Username),
    zap.Uint("user_id", user.ID),
    zap.String("ip", c.ClientIP()))

// 记录错误
logger.Error("数据库查询失败",
    zap.Error(err),
    zap.String("query", query))
```

### Vue代码规范

#### 组件命名

```vue
<!-- ✅ 好的做法：多词组件名 -->
<template>
  <UserList />
  <UserProfile />
</template>

<!-- ❌ 不好的做法：单词组件名 -->
<template>
  <User />
</template>
```

#### 代码风格

```javascript
// 使用ES6+语法
const fetchData = async () => {
  try {
    const response = await api.getData()
    return response.data
  } catch (error) {
    console.error('获取数据失败:', error)
  }
}

// 解构赋值
const { name, email } = user

// 模板字符串
const message = `你好，${name}！`
```

## Git工作流

### 分支策略

- `main`: 生产分支
- `develop`: 开发分支
- `feature/*`: 功能分支
- `bugfix/*`: 修复分支

### 提交规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型（type）:**
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具更新

**示例:**

```bash
feat(auth): 添加JWT认证中间件

- 实现JWT token生成和验证
- 添加认证中间件
- 更新路由以使用新中间件

Closes #123
```

### 开发流程

1. 从develop创建功能分支
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/user-management
   ```

2. 开发并提交代码
   ```bash
   git add .
   git commit -m "feat(user): 添加用户管理功能"
   ```

3. 推送并创建PR
   ```bash
   git push origin feature/user-management
   ```

4. 代码审查后合并到develop

## 调试技巧

### 后端调试

#### 使用Delve（Go调试器）

```bash
# 安装Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试应用
dlv debug main.go

# 常用命令
(dlv) break main.main:1   # 设置断点
(dlv) breakpoints        # 查看断点
(dlv) continue           # 继续执行
(dlv) next               # 下一步
(dlv) print variable     # 打印变量
(dlv) quit               # 退出
```

#### 日志调试

```go
// 使用debug级别日志
logger.Debug("变量值",
    zap.String("key", key),
    zap.Any("value", value))

// 记录函数调用
logger.Debug("函数调用",
    zap.String("function", "GetUser"),
    zap.Uint("user_id", userID))
```

### 前端调试

#### Vue DevTools

1. 安装Vue DevTools浏览器插件
2. 查看组件树、Vuex状态、路由信息
3. 性能分析和事件追踪

#### Console调试

```javascript
// 基础日志
console.log('普通日志')
console.warn('警告信息')
console.error('错误信息')

// 分组日志
console.group('用户信息')
console.log('姓名:', user.name)
console.log('邮箱:', user.email)
console.groupEnd()

// 表格显示
console.table(users)

// 性能测量
console.time('API请求')
await fetchUser()
console.timeEnd('API请求')
```

## 常见问题

### 后端问题

#### Q: 编译错误：cannot find package

```bash
# 解决方案
go mod tidy
go mod download
```

#### Q: 配置文件加载失败

```bash
# 确认配置文件存在
ls -la backend/config/config.yaml

# 检查文件权限
chmod 644 backend/config/config.yaml

# 验证YAML语法
# 使用在线工具：https://www.yamllint.com/
```

#### Q: 热加载不工作

```bash
# 确认air已安装
which air

# 如果未安装
go install github.com/air-verse/air@latest

# 检查.air.toml配置
cat .air.toml
```

### 前端问题

#### Q: npm install失败

```bash
# 清除缓存
npm cache clean --force

# 删除node_modules重试
rm -rf node_modules package-lock.json
npm install

# 或使用cnpm
npm install -g cnpm --registry=https://registry.npmmirror.com
cnpm install
```

#### Q: Vite HMR不工作

```javascript
// 检查vite.config.ts中的HMR配置
server: {
  hmr: process.env.DISABLE_HMR !== 'true',
}

// 确保没有设置DISABLE_HMR环境变量
```

#### Q: API请求跨域

```javascript
// 开发环境使用Vite代理
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8082',
      changeOrigin: true
    }
  }
}
```

## 性能优化建议

### 后端优化

1. **数据库查询**
   - 使用索引
   - 避免N+1查询
   - 使用预加载（Gorm Preload）

2. **缓存策略**
   - 使用Redis缓存热点数据
   - 实现查询结果缓存

3. **并发处理**
   - 使用Goroutine处理并发请求
   - 使用Channel进行通信

### 前端优化

1. **代码分割**
   - 路由懒加载
   - 组件异步加载

2. **资源优化**
   - 图片压缩
   - 启用Gzip压缩
   - 使用CDN

3. **性能监控**
   - 使用Lighthouse分析
   - 监控Core Web Vitals

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交代码
4. 确保通过所有检查
5. 创建Pull Request

## 联系方式

- 项目Issues: [GitHub Issues](https://github.com/your-repo/issues)
- 邮箱: support@oneops.com
