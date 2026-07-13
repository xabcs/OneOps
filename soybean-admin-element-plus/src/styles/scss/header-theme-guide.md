# Header 内部元素配色跟随机制

## 🎯 核心功能

Header 内部的所有元素（按钮、输入框、面包屑、状态指示器等）会**自动跟随** header 的背景配色，确保视觉协调统一。

## 🔧 工作原理

### 1️⃣ **智能颜色计算**

系统会根据 header 的背景色自动计算内部元素的颜色：

```typescript
// 渐变模式
当 header 使用渐变背景时：
→ 按钮/输入框自动使用半透明背景
→ 透明度基于渐变的起始色 alpha 值
→ 确保"玻璃态"视觉效果

// 纯色模式
当 header 使用纯色背景时：
→ 根据背景亮度自动调整
→ 浅色背景 → 深色文字
→ 深色背景 → 浅色文字

// 默认模式
当 header 没有自定义颜色时：
→ 使用 SxDevOps 默认浅色渐变
→ 内部元素使用协调的半透明样式
```

### 2️⃣ **自适应圆角系统**

```css
header 按钮: 8px (略小于内容区按钮)
header 输入框: 8px
header 状态标签: 6px
```

圆角比内容区略小，适应 header 的紧凑布局。

### 3️⃣ **玻璃态效果**

所有内部元素都使用 `backdrop-filter: blur(8px)` 实现毛玻璃效果：

```css
按钮背景: rgba(255, 255, 255, 0.85) + blur
输入框背景: rgba(255, 255, 255, 0.75) + blur
状态标签背景: rgba(255, 255, 255, 0.6) + blur
```

## 🎨 支持的 Header 元素

### 完全支持的元素：

✅ **按钮** (`.el-button`)
- 主按钮、默认按钮、图标按钮
- 自动半透明背景
- 悬停时透明度提升

✅ **输入框** (`.el-input`, `.el-input-number`)
- 文本输入框、数字输入框
- 自动半透明背景
- 聚焦时边框高亮

✅ **选择器** (`.el-select`)
- 下拉选择框
- 与输入框相同的玻璃态效果

✅ **面包屑** (`.el-breadcrumb`)
- 自动颜色调整
- 悬停时高亮

✅ **状态指示器** (`.status-item`, `.time-display`)
- 系统状态、时间显示
- 半透明背景

✅ **下拉面板** (`.dropdown-panel`)
- 通知面板、快速菜单
- 玻璃态模糊背景

✅ **徽章** (`.notification-badge`)
- 通知数量徽章
- 自动配色

## 📋 使用示例

### 自动适配（无需额外代码）：

```vue
<template>
  <!-- Header 组件内部所有元素自动跟随配色 -->
  <div class="global-header" :style="headerStyle">
    <GlobalBreadcrumb />
    <el-input v-model="search" placeholder="搜索..." />
    <el-button>搜索</el-button>
    <div class="status-item">
      <div class="status-dot status-success"></div>
      <span class="status-label">API</span>
    </div>
  </div>
</template>
```

### 手动触发颜色更新：

```typescript
import { applyHeaderTheme } from '@/utils/content-theme';

// 当用户在主题设置中调整 header 颜色时
const updateHeaderColor = () => {
  themeStore.setHeaderGradient(true, '#f8fbff', '#f0f9ff');
  // 自动调用 applyHeaderTheme 更新内部元素颜色
};
```

## 🎭 CSS 变量系统

Header 内部元素使用独立的 CSS 变量：

```css
/* 按钮变量 */
--header-button-default-bg: 按钮默认背景
--header-button-default-color: 按钮默认文字颜色
--header-button-default-border: 按钮默认边框
--header-button-hover-bg: 按钮悬停背景
--header-button-hover-color: 按钮悬停文字颜色
--header-button-hover-border: 按钮悬停边框
--header-button-radius: 按钮圆角

/* 输入框变量 */
--header-input-bg: 输入框背景
--header-input-border: 输入框边框
--header-input-hover-border: 输入框悬停边框
--header-input-focus-border: 输入框聚焦边框
--header-input-radius: 输入框圆角
--header-input-color: 输入框文字颜色

/* 其他元素变量 */
--header-breadcrumb-color: 面包屑颜色
--header-breadcrumb-hover: 面包屑悬停颜色
--header-status-bg: 状态标签背景
--header-status-radius: 状态标签圆角
```

## 🌈 配色方案

### 浅色 Header（默认）：
```css
按钮背景: rgba(255, 255, 255, 0.85)  /* 85% 不透明白 */
输入框背景: rgba(255, 255, 255, 0.75)  /* 75% 不透明白 */
状态标签: rgba(255, 255, 255, 0.6)    /* 60% 不透明白 */
文字颜色: #475569                        /* 深灰 */
```

### 深色 Header：
```css
按钮背景: rgba(255, 255, 255, 0.15)  /* 15% 不透明白 */
输入框背景: rgba(255, 255, 255, 0.12)  /* 12% 不透明白 */
状态标签: rgba(255, 255, 255, 0.12)    /* 12% 不透明白 */
文字颜色: rgba(255, 255, 255, 0.9)  /* 90% 不透明白 */
```

## 🔬 自动亮度检测

系统使用以下公式计算颜色亮度：

```typescript
brightness = (r × 299 + g × 587 + b × 114) / 1000

brightness > 128 → 浅色 → 深色文字
brightness ≤ 128 → 深色 → 浅色文字
```

这确保无论 header 使用什么背景色，文字始终清晰可读。

## ⚙️ 主题设置集成

在主题设置的"顶栏设置"中调整 header 颜色时：

1. **开启顶栏渐变** → 内部元素自动使用玻璃态效果
2. **使用自定义颜色** → 内部元素自动根据亮度调整
3. **关闭自定义** → 使用默认 SxDevOps 浅色渐变

所有调整都是**实时生效**，无需刷新页面！

## 🎯 设计优势

1. **视觉统一** - header 内部元素永远不会与背景冲突
2. **自动适应** - 用户修改 header 颜色，内部元素自动调整
3. **保持对比** - 确保文字在所有背景下清晰可读
4. **玻璃态美学** - 半透明元素创造现代感视觉效果
5. **无需手动** - 开发者无需单独处理每个元素的样式

## 📝 注意事项

1. **类名要求** - header 容器必须有 `.global-header` 类名
2. **CSS 优先级** - 使用 `!important` 确保样式覆盖 Element Plus 默认
3. **性能考虑** - `backdrop-filter` 可能影响性能，大面积使用时注意
4. **浏览器兼容** - 玻璃态效果需要现代浏览器支持

## 🚀 完整示例

```vue
<template>
  <!-- Header 自动应用内部元素配色 -->
  <div class="global-header" :style="headerStyle">
    <!-- 面包屑 - 自动配色 -->
    <GlobalBreadcrumb />

    <!-- 搜索输入框 - 玻璃态 -->
    <el-input v-model="searchText" placeholder="搜索..." style="width: 200px" />

    <!-- 按钮 - 半透明背景 -->
    <el-button type="primary">搜索</el-button>
    <el-button>重置</el-button>

    <!-- 状态指示器 - 自动配色 -->
    <div class="status-item">
      <div class="status-dot status-success"></div>
      <span class="status-label">API 正常</span>
    </div>

    <!-- 下拉菜单 - 玻璃态背景 -->
    <el-dropdown>
      <div class="header-icon-button">
        <Icon icon="mdi:dots-horizontal" />
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="handleRefresh">刷新</el-dropdown-item>
          <el-dropdown-item @click="handleExport">导出</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const searchText = ref('');

const handleRefresh = () => {
  console.log('刷新');
};

const handleExport = () => {
  console.log('导出');
};
</script>
```

所有内部元素都会自动根据 header 的背景色调整自己的样式，无需任何额外配置！
