# OneOps 前端项目结构说明

## 📁 目录结构

```
frontend/
├── .github/              # GitHub 配置
│   └── workflows/       # CI/CD 工作流
├── build/                # 构建配置
├── docs/                 # 📄 项目文档
│   ├── design/          # 设计文档
│   ├── api/             # API 文档
│   └── guides/          # 使用指南
├── packages/             # Monorepo 子包
├── public/               # 静态资源
│   └── icons/           # SVG 图标
├── src/                  # 源代码
│   ├── App.vue          # 根组件
│   ├── main.ts          # 入口文件
│   ├── assets/          # 静态资源
│   ├── components/      # 组件库（原子设计）
│   │   ├── atoms/      # 原子组件
│   │   ├── molecules/  # 分子组件
│   │   ├── organisms/  # 有机体组件
│   │   ├── common/     # 通用组件
│   │   ├── custom/     # 自定义组件
│   │   └── features/   # 功能组件
│   ├── views/          # 页面视图
│   │   ├── auth/       # 授权中心
│   │   ├── cmdb/       # 资产管理
│   │   ├── k8s/        # 容器管理
│   │   └── manage/     # 系统管理
│   ├── store/          # 状态管理
│   ├── service/        # API 服务层
│   ├── router/         # 路由配置
│   ├── utils/          # 工具函数
│   ├── hooks/          # 自定义 Hooks
│   ├── constants/      # 常量定义
│   ├── locales/        # 国际化
│   ├── theme/          # 主题配置
│   ├── typings/        # TypeScript 类型定义
│   ├── plugins/        # 插件
│   └── unocss/         # UnoCSS 配置
├── tests/                # 🧪 测试文件
│   ├── unit/            # 单元测试
│   │   ├── components/ # 组件测试
│   │   └── utils/      # 工具测试
│   ├── e2e/            # E2E 测试
│   └── html/           # HTML 测试
├── .env.example          # 环境变量示例
├── .gitignore           # Git 忽略规则
├── .prettierrc          # Prettier 配置
├── eslint.config.js     # ESLint 配置
├── package.json         # 项目依赖
├── tsconfig.json        # TypeScript 配置
├── vite.config.ts       # Vite 配置
├── vitest.config.ts     # Vitest 测试配置
└── README.md            # 项目说明
```

## 🎯 命名规范

### 文件命名
- **组件**: PascalCase (如 `UserSearch.vue`)
- **工具函数**: kebab-case (如 `auth-permission.ts`)
- **常量**: UPPER_SNAKE_CASE (如 `API_BASE_URL`)

### 代码命名
- **变量**: camelCase (如 `searchParams`)
- **函数**: camelCase (如 `handleSearch`)
- **组件**: PascalCase (如 `UserSearch`)
- **类型/接口**: PascalCase (如 `UserSearchParams`)

## 📦 技术栈

- **框架**: Vue 3.5.31
- **语言**: TypeScript 5.9.3
- **构建工具**: Vite 7 (rolldown-vite)
- **UI 框架**: Element Plus 2.13.6
- **状态管理**: Pinia 3.0.4
- **路由**: Vue Router 5.0.4
- **样式**: UnoCSS
- **测试**: Vitest + @vue/test-utils
- **包管理**: pnpm workspace

## 🚀 快速开始

### 安装依赖
```bash
pnpm install
```

### 运行开发服务器
```bash
pnpm dev
```

### 运行测试
```bash
# 运行所有测试
pnpm test

# 运行测试（监听模式）
pnpm test:run

# 生成测试覆盖率
pnpm test:coverage

# 打开测试 UI
pnpm test:ui
```

### 构建生产版本
```bash
pnpm build
```

## 📝 开发规范

### 组件开发
1. 遵循原子设计方法论
2. 单一职责原则
3. 组件命名清晰明确
4. 添加必要的 TypeScript 类型
5. 编写单元测试

### Git 提交
- 遵循 Conventional Commits 规范
- 提交前运行 `pnpm lint` 检查代码
- 提交前运行 `pnpm typecheck` 检查类型

### 测试要求
- 组件测试覆盖率 > 80%
- 工具函数测试覆盖率 > 90%
- 关键业务逻辑必须有测试

## 🔗 相关文档

- [API 文档](./api/README.md)
- [设计系统](./design/README.md)
- [开发指南](./guides/README.md)
