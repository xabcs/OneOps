# OneOps Frontend

一个现代化、专业的运维管理平台前端项目，基于 Vue 3 和 TypeScript 构建。

## ✨ 特性

- 🚀 **现代化技术栈** - Vue 3.5 + TypeScript 5.9 + Vite 7
- 📦 **组件设计** - 采用原子设计方法论（Atomic Design）
- 🎨 **主题系统** - 灵活的主题配置和定制能力
- 🌍 **国际化** - 完整的多语言支持
- 📱 **响应式** - 适配各种屏幕尺寸
- 🧪 **测试完善** - Vitest 单元测试 + E2E 测试
- 🔧 **工程化** - ESLint + Prettier + TypeScript 严格模式

## 🛠️ 技术栈

### 核心框架
- **Vue 3.5.31** - 渐进式 JavaScript 框架
- **TypeScript 5.9.3** - 类型安全的 JavaScript 超集
- **Vite 7** - 下一代前端构建工具
- **Pinia 3.0.4** - Vue 状态管理
- **Vue Router 5.0.4** - Vue 路由管理

### UI 框架
- **Element Plus 2.13.6** - Vue 3 组件库
- **UnoCSS** - 即时原子化 CSS 引擎
- **Iconify** - 统一的图标框架

### 工具链
- **ESLint** - 代码检查
- **Prettier** - 代码格式化
- **Vitest** - 单元测试框架
- **pnpm** - 快速、节省磁盘空间的包管理器

## 📁 项目结构

```
frontend/
├── src/                  # 源代码
│   ├── components/      # 组件（原子设计）
│   ├── views/           # 页面视图
│   ├── store/           # 状态管理
│   ├── service/         # API 服务
│   ├── router/          # 路由配置
│   └── utils/           # 工具函数
├── tests/                # 测试文件
├── docs/                 # 项目文档
└── public/               # 静态资源
```

详细结构请查看 [项目结构说明](./docs/STRUCTURE.md)

## 🚀 快速开始

### 环境要求
- Node.js >= 20.19.0
- pnpm >= 8.7.0

### 安装依赖
```bash
pnpm install
```

### 开发模式
```bash
pnpm dev
```

访问 http://localhost:9528

### 构建生产版本
```bash
pnpm build
```

### 预览生产构建
```bash
pnpm preview
```

## 🧪 测试

```bash
# 运行测试
pnpm test

# 生成覆盖率报告
pnpm test:coverage

# 测试 UI
pnpm test:ui
```

## 📝 开发规范

### 代码风格
- 使用 ESLint 和 Prettier 统一代码风格
- 提交前运行 `pnpm lint` 检查代码
- TypeScript 严格模式

### 命名规范
- 组件：PascalCase（如 `UserSearch.vue`）
- 文件：kebab-case（如 `user-search.ts`）
- 变量：camelCase（如 `searchParams`）

### Git 提交
遵循 Conventional Commits 规范：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

## 📚 文档

- [项目结构说明](./docs/STRUCTURE.md)
- [优化总结](./docs/OPTIMIZATION_SUMMARY.md)
- [环境变量配置](./.env.example)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request。

## 📄 许可证

[MIT License](./LICENSE)

## 🙏 致谢

本项目基于 [Soybean Admin](https://github.com/soybeanjs/soybean-admin) 开发。
