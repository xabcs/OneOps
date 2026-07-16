# OneOps 配色系统使用指南

## 📋 概述

本指南帮助您将 SxDevOps 配色系统完整集成到 OneOps 项目中。我们已经创建了三个核心样式文件：

1. **sxdevops-theme.scss** - 核心 CSS 变量系统
2. **layout-theme.scss** - 布局组件配色样式
3. **interaction-states.scss** - 交互状态配色样式

## 🎯 配色系统特点

- **现代化渐变设计**：基于 135deg、145deg、180deg 等多角度渐变系统
- **玻璃态设计**：半透明背景 + 模糊效果
- **完整状态管理**：悬停、激活、禁用等交互状态
- **响应式适配**：支持移动端和深色模式
- **工业风格组件**：K8s、Docker 等工具栏配色

## 📦 安装步骤

### 1. 在主样式文件中导入

在您的 `src/styles/scss/global.scss` 或主样式文件中添加以下导入：

```scss
// 导入 SxDevOps 配色系统
@import './sxdevops-theme.scss';
@import './layout-theme.scss';
@import './interaction-states.scss';

// 导入其他现有样式
@import './design-system.scss';
@import './element-plus.scss';
// ... 其他样式文件
```

### 2. 在 main.ts 中初始化

确保在 Vue 应用入口文件中导入样式：

```typescript
// src/main.ts
import './styles/scss/global.scss'; // 或您的主样式文件路径
```

### 3. 在组件中使用

您现在可以在组件中使用这些配色变量和样式类：

#### 使用 CSS 变量

```vue
<template>
  <div class="my-component">内容区域</div>
</template>

<style scoped>
.my-component {
  background: var(--sidebar-bg);
  color: var(--text-primary);
  border: 1px solid var(--border-light);
  border-radius: var(--card-radius);
  transition: var(--transition);
}

.my-component:hover {
  background: var(--sidebar-hover);
  box-shadow: var(--brand-shadow-md);
}
</style>
```

#### 使用预定义样式类

```vue
<template>
  <!-- 卡片组件 -->
  <div class="card-interactive">
    <h3>卡片标题</h3>
    <p>卡片内容</p>
  </div>

  <!-- 按钮组件 -->
  <button class="btn-interactive">点击我</button>

  <!-- 输入框组件 -->
  <input type="text" class="input-interactive" placeholder="请输入内容">

  <!-- 状态指示器 -->
  <span class="status-dot success"></span>
  <span class="status-dot warning"></span>
  <span class="status-dot danger"></span>
</template>
```

## 🎨 主要配色变量

### 布局变量

```css
--sidebar-width: 188px;              /* 侧边栏宽度 */
--sidebar-collapsed-width: 68px;     /* 侧边栏折叠宽度 */
--header-height: 60px;                /* 顶栏高度 */
--content-bg: #f1f5f9;                /* 内容区背景 */
```

### 品牌配色

```css
--primary: #6366f1;                   /* 主色调 */
--primary-light: #818cf8;             /* 浅主色 */
--primary-dark: #4f46e5;              /* 深主色 */
--brand-gradient: linear-gradient(...); /* 品牌渐变 */
```

### 侧边栏配色

```css
--sidebar-bg: linear-gradient(...);   /* 侧边栏背景 */
--sidebar-hover: #e8f0ff;             /* 悬停背景 */
--sidebar-active: linear-gradient(...); /* 激活背景 */
--sidebar-text: #5f6b7a;              /* 文字颜色 */
--sidebar-text-active: #2563eb;       /* 激活文字颜色 */
```

### 文字层级

```css
--text-primary: #1e293b;              /* 主要文字 */
--text-secondary: #64748b;            /* 次要文字 */
--text-muted: #94a3b8;                /* 弱化文字 */
```

### 状态色

```css
--success: #10b981;                   /* 成功 */
--warning: #f59e0b;                   /* 警告 */
--danger: #ef4444;                    /* 危险 */
--info: #3b82f6;                      /* 信息 */
```

## 🧩 布局组件使用

### 侧边栏样式

```vue
<template>
  <aside class="sidebar" :class="{ collapsed: isCollapsed }">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <img src="@/assets/logo.svg" alt="Logo" />
      </div>
      <div class="logo-copy">
        <span class="logo-text">OneOps</span>
        <span class="logo-subtext">AI Agent</span>
      </div>
    </div>

    <nav class="sidebar-nav">
      <el-menu :collapse="isCollapsed">
        <el-menu-item index="/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
      </el-menu>
    </nav>
  </aside>
</template>
```

### 顶栏样式

```vue
<template>
  <header class="header">
    <div class="header-left">
      <button class="collapse-btn" @click="toggleSidebar">
        <el-icon><Fold /></el-icon>
      </button>
      <span class="breadcrumb">{{ currentPageTitle }}</span>
    </div>

    <div class="header-right">
      <button class="promo-trigger">
        <el-icon><Promotion /></el-icon>
        <span>产品介绍</span>
      </button>

      <button class="assistant-trigger">
        <el-icon><Service /></el-icon>
      </button>

      <div class="user-trigger">
        <el-avatar :size="36" class="user-avatar">
          {{ userInitials }}
        </el-avatar>
        <div class="user-meta">
          <span class="user-name">{{ userName }}</span>
          <span class="user-role">{{ userRole }}</span>
        </div>
      </div>
    </div>
  </header>
</template>
```

## 🔧 组件交互状态

### 按钮组件

```vue
<template>
  <button class="btn-interactive">
    点击我
  </button>

  <button class="btn-interactive" disabled>
    禁用状态
  </button>
</template>
```

### 卡片组件

```vue
<template>
  <div class="card-interactive">
    <h3>卡片标题</h3>
    <p>卡片内容</p>
  </div>
</template>
```

### 输入框组件

```vue
<template>
  <input
    type="text"
    class="input-interactive"
    placeholder="请输入内容"
  />
</template>
```

### 表格组件

```vue
<template>
  <el-table class="table-interactive" :data="tableData">
    <el-table-column prop="name" label="姓名" />
    <el-table-column prop="age" label="年龄" />
  </el-table>
</template>
```

## 🎭 高级样式效果

### 品牌渐变效果

```vue
<template>
  <div class="gradient-brand">
    <h1>品牌渐变标题</h1>
  </div>
</template>
```

### 玻璃态效果

```vue
<template>
  <div class="gradient-glass">
    <p>玻璃态效果内容</p>
  </div>
</template>
```

### 脉冲动画

```vue
<template>
  <div>
    <span class="state-pulse running"></span>
    <span>运行中</span>
  </div>
</template>
```

## 🌙 深色模式支持

配色系统自动支持深色模式，无需额外配置。当用户系统设置为深色模式时，会自动应用相应的配色。

```css
@media (prefers-color-scheme: dark) {
  /* 自动应用深色模式配色 */
}
```

## 📱 响应式适配

配色系统包含完整的响应式支持，在移动端会自动调整侧边栏和布局。

```css
@media (max-width: 768px) {
  /* 移动端自动优化 */
}
```

## 🎯 最佳实践

### 1. 使用 CSS 变量而非硬编码

```css
/* ✅ 推荐 */
.my-element {
  background: var(--sidebar-bg);
  color: var(--text-primary);
}

/* ❌ 避免 */
.my-element {
  background: linear-gradient(180deg, rgba(241, 246, 255, 0.98), rgba(232, 240, 255, 0.9));
  color: #1e293b;
}
```

### 2. 使用预定义样式类

```vue
<!-- ✅ 推荐 -->
<div class="card-interactive">内容</div>

<!-- ❌ 避免 -->
<div class="custom-card-class">内容</div>
```

### 3. 保持配色一致性

在整个项目中保持使用相同的配色变量，确保视觉一致性。

## 🔄 迁移现有组件

如果您有现有的组件需要迁移到新的配色系统：

### 1. 逐步替换硬编码颜色

```css
/* 替换前 */
.old-button {
  background: #3b82f6;
  color: white;
  border-radius: 8px;
}

/* 替换后 */
.new-button {
  background: var(--primary);
  color: white;
  border-radius: var(--card-radius);
}
```

### 2. 添加交互状态

```css
/* 添加悬停效果 */
.new-button:hover {
  background: var(--primary-light);
  transform: translateY(-1px);
  box-shadow: var(--brand-shadow-md);
}
```

### 3. 使用预定义样式类

```vue
<!-- 最终简化版本 -->
<button class="btn-interactive">
  点击我
</button>
```

## 🧪 测试和验证

### 1. 视觉测试

在不同页面和组件中测试配色效果，确保：

- 渐变效果正常显示
- 交互状态符合预期
- 响应式布局正常工作

### 2. 兼容性测试

测试不同浏览器的兼容性：

- Chrome、Firefox、Safari、Edge
- 移动端浏览器
- 深色模式切换

### 3. 性能测试

确保：

- 页面加载性能正常
- 动画流畅不卡顿
- 内存使用合理

## 📚 参考资源

### 完整配色变量列表

查看 `sxdevops-theme.scss` 文件获取完整的 CSS 变量列表。

### 布局组件样式

查看 `layout-theme.scss` 文件了解详细的布局组件样式。

### 交互状态样式

查看 `interaction-states.scss` 文件了解所有交互状态样式。

## 🆘 故障排除

### 问题：配色没有生效

**解决方案**：
1. 确保正确导入了所有样式文件
2. 检查 CSS 变量名称是否正确
3. 清除浏览器缓存

### 问题：渐变效果显示异常

**解决方案**：
1. 检查浏览器兼容性
2. 确认渐变语法正确
3. 测试不同的渐变角度

### 问题：移动端显示异常

**解决方案**：
1. 检查媒体查询是否正常工作
2. 测试不同移动设备的显示效果
3. 确认 viewport 设置正确

## 🚀 下一步

1. **逐步迁移**：从新组件开始使用新的配色系统
2. **统一风格**：逐步替换现有组件的硬编码颜色
3. **优化性能**：根据实际使用情况优化配色系统
4. **扩展功能**：根据项目需求添加新的配色变量和样式类

希望这个指南能帮助您成功集成 SxDevOps 配色系统到您的 OneOps 项目中！