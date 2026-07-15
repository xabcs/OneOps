# 内容区模块化主题系统指南

## 🎯 概述

本文档介绍新增的**内容区设置2（contentTheme2）**功能，这是一个基于模块化设计的精细主题控制系统，允许为不同页面区域单独设置颜色、渐变和样式。

## 📐 模块分类体系

### 核心设计理念

将内容区划分为 8 个独立模块，每个模块都有完整的主题控制能力：

```
内容区 (contentTheme2)
├── 1. Hero区域 (heroSection)
├── 2. 统计卡片 (statCards) 
├── 3. 工具栏 (toolbar)
├── 4. 内容卡片 (contentCard)
├── 5. 数据表格 (dataTable)
├── 6. 搜索筛选 (searchFilters)
├── 7. 分页组件 (pagination)
└── 8. 标签样式 (tags)
```

## 🎨 模块详细配置

### 1️⃣ Hero区域 (heroSection)

**适用场景**: 页面顶部的标题、描述、操作按钮区域

```typescript
heroSection: {
  // 背景渐变
  useGradient: true,
  gradientStart: 'rgba(251, 253, 255, 0.98)',
  gradientEnd: 'rgba(246, 250, 255, 0.96)',
  
  // 边框和阴影
  borderColor: 'rgba(36, 91, 219, 0.09)',
  borderRadius: '20px',
  shadow: '0 8px 24px rgba(15, 23, 42, 0.04)',
  padding: '14px 22px',
  
  // 图标样式
  iconGradientStart: 'rgba(243, 247, 255, 0.98)',
  iconGradientEnd: 'rgba(235, 242, 255, 0.96)',
  iconBorderColor: 'rgba(36, 91, 219, 0.12)',
  iconColor: '#245bdb'
}
```

**CSS变量映射**:
- `--sx-hero-bg`: Hero背景渐变
- `--sx-hero-border`: 边框颜色
- `--sx-hero-radius`: 圆角大小
- `--sx-hero-shadow`: 阴影效果
- `--sx-hero-padding`: 内边距
- `--sx-hero-icon-bg`: 图标背景渐变
- `--sx-hero-icon-border`: 图标边框
- `--sx-hero-icon-color`: 图标颜色

### 2️⃣ 统计卡片 (statCards)

**适用场景**: 数据统计卡片、状态指标展示

```typescript
statCards: {
  // 默认卡片
  defaultBg: 'rgba(255, 255, 255, 0.98)',
  defaultBorder: 'rgba(148, 163, 184, 0.16)',
  
  // 成功状态卡片渐变
  successBgStart: 'rgba(240, 253, 244, 0.98)',
  successBgEnd: 'rgba(255, 255, 255, 0.94)',
  
  // 警告状态卡片渐变
  warningBgStart: 'rgba(255, 251, 235, 0.98)',
  warningBgEnd: 'rgba(255, 255, 255, 0.94)',
  
  // 危险状态卡片渐变
  dangerBgStart: 'rgba(254, 242, 242, 0.98)',
  dangerBgEnd: 'rgba(255, 255, 255, 0.94)',
  
  // 通用样式
  borderRadius: '12px',
  shadow: '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)'
}
```

**CSS变量映射**:
- `--sx-stat-default-bg`: 默认背景
- `--sx-stat-default-border`: 默认边框
- `--sx-stat-success-bg`: 成功卡片背景渐变
- `--sx-stat-warning-bg`: 警告卡片背景渐变
- `--sx-stat-danger-bg`: 危险卡片背景渐变
- `--sx-stat-radius`: 圆角
- `--sx-stat-shadow`: 阴影

### 3️⃣ 工具栏 (toolbar)

**适用场景**: 顶部工具栏、搜索工具栏

```typescript
toolbar: {
  // 顶部工具栏
  sectionBg: 'transparent',
  sectionBorder: 'rgba(148, 163, 184, 0.12)',
  
  // 搜索工具栏渐变
  searchBg: 'linear-gradient(180deg, rgba(248, 250, 252, 0.92) 0%, rgba(255, 255, 255, 0.96) 100%)',
  searchBorder: 'rgba(148, 163, 184, 0.12)',
  
  // 通用工具栏样式
  gradientStart: 'rgba(248, 250, 252, 0.92)',
  gradientEnd: 'rgba(255, 255, 255, 0.96)',
  borderColor: 'rgba(148, 163, 184, 0.12)',
  borderRadius: '12px',
  padding: '6px 8px',
  shadow: 'inset 0 1px 0 rgba(255, 255, 255, 0.9)'
}
```

**CSS变量映射**:
- `--sx-toolbar-bg`: 工具栏背景
- `--sx-toolbar-border`: 边框颜色
- `--sx-toolbar-radius`: 圆角
- `--sx-toolbar-padding`: 内边距
- `--sx-toolbar-shadow`: 阴影

### 4️⃣ 内容卡片 (contentCard)

**适用场景**: 主要内容容器、数据展示区域

```typescript
contentCard: {
  // 背景设置
  background: '#ffffff',
  bgGradientStart: 'rgba(255, 255, 255, 0.98)',
  bgGradientEnd: 'rgba(248, 250, 252, 0.94)',
  useGradient: false,
  
  // 样式设置
  borderColor: 'rgba(148, 163, 184, 0.16)',
  borderRadius: '12px',
  shadow: '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)',
  padding: '20px'
}
```

**CSS变量映射**:
- `--sx-content-card-bg`: 卡片背景（支持渐变）
- `--sx-content-card-border`: 边框颜色
- `--sx-content-card-radius`: 圆角
- `--sx-content-card-shadow`: 阴影
- `--sx-content-card-padding`: 内边距

### 5️⃣ 数据表格 (dataTable)

**适用场景**: 表格头部、数据行、边框样式

```typescript
dataTable: {
  // 表头样式
  headerBg: '#f8fafc',
  headerTextColor: '#475569',
  headerBorderColor: 'rgba(148, 163, 184, 0.16)',
  
  // 数据行样式
  rowHoverBg: '#f8fbff',
  rowBorderColor: 'rgba(148, 163, 184, 0.16)',
  
  // 表格样式
  tableBorder: 'rgba(148, 163, 184, 0.14)',
  borderRadius: '12px',
  stripedBg: '#f8fafc'
}
```

**CSS变量映射**:
- `--sx-table-header-bg`: 表头背景
- `--sx-table-header-text`: 表头文字颜色
- `--sx-table-header-border`: 表头边框
- `--sx-table-row-hover`: 行悬停背景
- `--sx-table-row-border`: 行边框
- `--sx-table-border`: 表格边框
- `--sx-table-radius`: 表格圆角
- `--sx-table-striped-bg`: 斑纹行背景

### 6️⃣ 搜索筛选 (searchFilters)

**适用场景**: 搜索输入框、筛选器控件

```typescript
searchFilters: {
  // 输入框样式
  inputBg: 'rgba(255, 255, 255, 0.94)',
  inputBorder: 'rgba(148, 163, 184, 0.12)',
  inputHoverBorder: 'rgba(59, 130, 246, 0.16)',
  inputFocusBorder: 'rgba(37, 99, 235, 0.22)',
  inputBorderRadius: '8px',
  
  // 按钮样式
  buttonBg: 'rgba(255, 255, 255, 0.9)',
  buttonTextColor: '#475569',
  buttonHoverBg: '#f8fbff'
}
```

**CSS变量映射**:
- `--sx-search-input-bg`: 输入框背景
- `--sx-search-input-border`: 输入框边框
- `--sx-search-input-hover-border`: 悬停边框
- `--sx-search-input-focus-border`: 聚焦边框
- `--sx-search-input-radius`: 输入框圆角
- `--sx-search-button-bg`: 按钮背景
- `--sx-search-button-text`: 按钮文字颜色
- `--sx-search-button-hover`: 按钮悬停背景

### 7️⃣ 分页组件 (pagination)

**适用场景**: 分页按钮、页码显示

```typescript
pagination: {
  // 按钮样式
  buttonBg: 'rgba(255, 255, 255, 0.9)',
  buttonTextColor: '#475569',
  buttonHoverBg: '#f8fbff',
  
  // 激活状态
  activeButtonBg: '#3b82f6',
  activeButtonTextColor: '#ffffff',
  
  // 通用样式
  borderRadius: '8px'
}
```

**CSS变量映射**:
- `--sx-pagination-button-bg`: 按钮背景
- `--sx-pagination-button-text`: 按钮文字颜色
- `--sx-pagination-button-hover`: 按钮悬停背景
- `--sx-pagination-active-bg`: 激活按钮背景
- `--sx-pagination-active-text`: 激活按钮文字
- `--sx-pagination-radius`: 圆角

### 8️⃣ 标签样式 (tags)

**适用场景**: 状态标签、分类标签、角色标签

```typescript
tags: {
  // 默认标签
  defaultBg: 'rgba(255, 255, 255, 0.9)',
  defaultBorder: 'rgba(148, 163, 184, 0.12)',
  defaultTextColor: '#475569',
  
  // 状态标签背景
  successBg: 'rgba(16, 185, 129, 0.1)',
  warningBg: 'rgba(245, 158, 11, 0.1)',
  dangerBg: 'rgba(239, 68, 68, 0.1)',
  infoBg: 'rgba(59, 130, 246, 0.1)',
  
  // 通用样式
  borderRadius: '8px'
}
```

**CSS变量映射**:
- `--sx-tag-default-bg`: 默认标签背景
- `--sx-tag-default-border`: 默认标签边框
- `--sx-tag-default-text`: 默认标签文字
- `--sx-tag-success-bg`: 成功标签背景
- `--sx-tag-warning-bg`: 警告标签背景
- `--sx-tag-danger-bg`: 危险标签背景
- `--sx-tag-info-bg`: 信息标签背景
- `--sx-tag-radius`: 标签圆角

## 🚀 使用方式

### 1. 主题配置设置

在 `src/theme/settings.ts` 中配置默认值：

```typescript
export const themeSettings: App.Theme.ThemeSetting = {
  // ... 其他配置
  contentTheme2: {
    heroSection: {
      useGradient: true,
      gradientStart: 'rgba(251, 253, 255, 0.98)',
      gradientEnd: 'rgba(246, 250, 255, 0.96)',
      // ... 更多配置
    },
    // ... 更多模块配置
  }
};
```

### 2. 应用主题

```typescript
import { initContentTheme2, applyContentTheme2 } from '@/utils/content-theme';

// 初始化默认主题
initContentTheme2();

// 动态应用自定义主题
applyContentTheme2(customThemeSettings);
```

### 3. 在Vue组件中使用

```vue
<template>
  <div class="user-management-page">
    <!-- Hero区域自动应用主题 -->
    <section class="hero-section">
      <!-- 内容 -->
    </section>
    
    <!-- 统计卡片自动应用主题 -->
    <div class="stats-grid">
      <div class="stat-card content-card success-card">
        <!-- 内容 -->
      </div>
    </div>
    
    <!-- 工具栏自动应用主题 -->
    <div class="workbench-toolbar workbench-toolbar--history">
      <!-- 内容 -->
    </div>
  </div>
</template>

<style scoped lang="scss">
.user-management-page {
  /* 使用CSS变量 */
  .hero-section {
    background: var(--sx-hero-bg);
    border: 1px solid var(--sx-hero-border);
    border-radius: var(--sx-hero-radius);
    box-shadow: var(--sx-hero-shadow);
    padding: var(--sx-hero-padding);
  }
  
  .stat-card.success-card {
    background: var(--sx-stat-success-bg);
    border-radius: var(--sx-stat-radius);
    box-shadow: var(--sx-stat-shadow);
  }
  
  .workbench-toolbar {
    background: var(--sx-toolbar-bg);
    border: 1px solid var(--sx-toolbar-border);
    border-radius: var(--sx-toolbar-radius);
    padding: var(--sx-toolbar-padding);
    box-shadow: var(--sx-toolbar-shadow);
  }
}
</style>
```

## 🎨 实际应用示例：用户管理页面

### Hero区域样式

```vue
<style scoped lang="scss">
.hero-section {
  background: var(--sx-hero-bg);
  border: 1px solid var(--sx-hero-border);
  border-radius: var(--sx-hero-radius);
  box-shadow: var(--sx-hero-shadow);
  padding: var(--sx-hero-padding);
}

.hero-icon {
  background: var(--sx-hero-icon-bg);
  border: 1px solid var(--sx-hero-icon-border);
  color: var(--sx-hero-icon-color);
}
</style>
```

### 统计卡片样式

```vue
<style scoped lang="scss">
.stat-card {
  background: var(--sx-stat-default-bg);
  border: 1px solid var(--sx-stat-default-border);
  border-radius: var(--sx-stat-radius);
  box-shadow: var(--sx-stat-shadow);
}

.stat-card.success-card {
  background: var(--sx-stat-success-bg);
}

.stat-card.warning-card {
  background: var(--sx-stat-warning-bg);
}
</style>
```

### 工具栏样式

```vue
<style scoped lang="scss">
.workbench-toolbar {
  background: var(--sx-toolbar-bg);
  border: 1px solid var(--sx-toolbar-border);
  border-radius: var(--sx-toolbar-radius);
  padding: var(--sx-toolbar-padding);
  box-shadow: var(--sx-toolbar-shadow);
}
</style>
```

### 搜索筛选样式

```vue
<style scoped lang="scss">
.workbench-toolbar {
  .el-input__wrapper {
    background: var(--sx-search-input-bg);
    border-radius: var(--sx-search-input-radius);
    box-shadow: 0 0 0 1px var(--sx-search-input-border) inset;
  }
  
  .el-input__wrapper:hover {
    box-shadow: 0 0 0 1px var(--sx-search-input-hover-border) inset;
  }
  
  .el-input.is-focus .el-input__wrapper {
    box-shadow: 0 0 0 1px var(--sx-search-input-focus-border) inset;
  }
  
  .el-button {
    background: var(--sx-search-button-bg);
    color: var(--sx-search-button-text);
  }
  
  .el-button:hover {
    background: var(--sx-search-button-hover);
  }
}
</style>
```

## 🔧 动态主题切换

### 方式1：通过主题抽屉切换

```typescript
import { useThemeStore } from '@/store/modules/theme';

const themeStore = useThemeStore();

// 切换到预定义主题
themeStore.setContentTheme2('dark');
```

### 方式2：自定义主题配置

```typescript
import { applyContentTheme2 } from '@/utils/content-theme';

// 应用自定义配置
applyContentTheme2({
  heroSection: {
    gradientStart: 'rgba(200, 200, 255, 0.98)',
    gradientEnd: 'rgba(180, 180, 255, 0.96)',
    // ... 其他配置
  },
  // ... 其他模块
});
```

## 🎯 设计优势

### 1. 模块化设计
- 每个模块独立配置，互不影响
- 支持单独调整某个模块而不影响其他模块

### 2. 向后兼容
- 不影响原有的 contentTheme 设置
- 可以同时使用两套主题系统

### 3. 精细控制
- 支持8个核心模块的详细配置
- 每个模块都有独立的颜色、渐变、圆角、阴影设置

### 4. 实时预览
- 主题配置立即生效
- CSS变量自动更新，无需重新编译

### 5. 易于扩展
- 模块化结构便于添加新的配置项
- 清晰的命名规范便于维护

## 📊 与原主题系统对比

| 特性 | contentTheme (原系统) | contentTheme2 (新系统) |
|------|----------------------|------------------------|
| 模块粒度 | 整体配置 | 8个独立模块 |
| 配置项数量 | ~25个 | ~60个 |
| 渐变支持 | 基础渐变 | 多层次渐变 |
| 实时性 | 支持 | 支持 |
| 向后兼容 | - | ✅ 完全兼容 |
| 学习曲线 | 简单 | 适中 |

## 🎨 推荐使用场景

### 使用 contentTheme (原系统)
- 简单主题切换
- 快速原型开发
- 不需要精细控制的页面

### 使用 contentTheme2 (新系统)
- 需要精细设计的管理页面
- 多种状态样式控制
- 品牌化界面定制
- 复杂的数据展示页面

## 📝 总结

**contentTheme2** 提供了一个完整的模块化主题解决方案，让开发者能够：

1. ✅ **精确控制**: 8个核心模块的独立配置
2. ✅ **灵活组合**: 支持多种主题样式组合
3. ✅ **实时预览**: 配置更改立即生效
4. ✅ **向后兼容**: 不影响原有主题系统
5. ✅ **易于维护**: 清晰的模块结构和命名规范

这个系统特别适合需要**精细化设计控制**的管理后台页面，如用户管理、角色管理、系统监控等核心功能页面。