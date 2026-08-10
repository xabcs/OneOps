# 内容主题使用指南

## 🎯 设计原则

**卡片内部元素统一配色** - 所有卡片内部的组件（按钮、输入框、表格等）自动使用协调的配色方案，确保视觉一致性。

## 📦 使用方式

### 1️⃣ **卡片容器类**

在页面中使用 `.content-card` 或 `.sx-card` 类来创建统一配色的卡片：

```vue
<template>
  <div class="content-card">
    <!-- 卡片内容 -->
    <div class="toolbar">
      <el-button>按钮</el-button>
      <el-input placeholder="搜索..." />
    </div>

    <el-table :data="tableData">
      <!-- 表格内容 -->
    </el-table>
  </div>
</template>
```

### 2️⃣ **工具栏样式**

使用 `.toolbar` 或 `.filter-bar` 类创建统一配色的工具栏：

```vue
<template>
  <div class="content-card">
    <div class="toolbar">
      <el-button>搜索</el-button>
      <el-button>重置</el-button>
      <el-input v-model="searchText" placeholder="关键词" />
    </div>

    <!-- 表格内容 -->
  </div>
</template>
```

### 3️⃣ **状态卡片**

使用状态卡片类创建不同状态的卡片：

```vue
<template>
  <!-- 成功状态卡片 -->
  <div class="content-card status-card success-card">
    <h3>操作成功</h3>
  </div>

  <!-- 警告状态卡片 -->
  <div class="content-card status-card warning-card">
    <h3>注意</h3>
  </div>

  <!-- 危险状态卡片 -->
  <div class="content-card status-card danger-card">
    <h3>错误</h3>
  </div>
</template>
```

## 🎨 CSS 变量系统

所有样式都通过 CSS 变量控制，可在主题设置中实时调整：

```css
/* 卡片变量 */
--sx-card-bg: 卡片背景色
--sx-card-radius: 卡片圆角
--sx-card-shadow: 卡片阴影
--sx-card-padding: 卡片内边距

/* 表格变量 */
--sx-table-border: 表格边框
--sx-table-header-bg: 表头背景
--sx-table-hover-bg: 悬停背景
--sx-table-radius: 表格圆角

/* 按钮变量 */
--sx-button-radius: 按钮圆角
--sx-button-default-bg: 默认按钮背景
--sx-button-hover-bg: 悬停背景

/* 输入框变量 */
--sx-input-radius: 输入框圆角
--sx-input-bg: 输入框背景
--sx-input-border-shadow: 边框阴影

/* 工具栏变量 */
--sx-toolbar-bg: 工具栏背景
--sx-toolbar-radius: 工具栏圆角
--sx-toolbar-padding: 工具栏内边距
```

## 🔧 主题设置

在主题设置的"内容区设置"部分可以实时调整：

1. **卡片设置**
   - 卡片背景颜色
   - 卡片圆角大小
   - 可选的卡片渐变背景

2. **表格设置**
   - 表头背景颜色
   - 行悬停背景颜色
   - 表格圆角大小

3. **按钮和输入框**
   - 圆角大小统一调整

4. **工具栏**
   - 可选的渐变背景

## 🎭 设计协调性

### 圆角统一
- 卡片: 12px
- 表格: 12px
- 按钮: 10px
- 输入框: 12px
- 工具栏: 12px

### 颜色协调
- 卡片背景: `#ffffff` 纯白或渐变
- 表头: `#f8fafc` 浅灰蓝
- 悬停: `#f8fbff` 浅蓝白
- 边框: `rgba(148, 163, 184, 0.16)` 半透明灰

### 阴影层次
- 卡片: `0 1px 3px rgba(0, 0, 0, 0.06)` 微弱阴影
- 悬停: `0 8px 25px rgba(0, 0, 0, 0.1)` 提升阴影
- 按钮: `0 4px 12px rgba(0, 0, 0, 0.1)` 交互阴影

## 📋 完整示例

```vue
<template>
  <div class="content-card">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-input v-model="searchText" placeholder="搜索..." style="width: 200px" />
      <el-button type="primary">搜索</el-button>
      <el-button>重置</el-button>
      <el-button @click="handleExport">导出</el-button>
    </div>

    <!-- 状态卡片 -->
    <div class="status-cards" style="display: flex; gap: 16px; margin-bottom: 20px;">
      <div class="status-card success-card" style="flex: 1; padding: 16px;">
        <div class="stat">成功: {{ successCount }}</div>
      </div>
      <div class="status-card warning-card" style="flex: 1; padding: 16px;">
        <div class="stat">警告: {{ warningCount }}</div>
      </div>
    </div>

    <!-- 表格 -->
    <el-table :data="tableData" style="width: 100%">
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="date" label="日期" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 20px; justify-content: flex-end;"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const searchText = ref('');
const successCount = ref(10);
const warningCount = ref(5);
const currentPage = ref(1);
const pageSize = ref(10);
const total = ref(100);

const tableData = ref([
  { name: '示例1', status: '正常', date: '2024-01-01' },
  { name: '示例2', status: '异常', date: '2024-01-02' },
]);

const handleEdit = (row: any) => {
  console.log('编辑:', row);
};

const handleDelete = (row: any) => {
  console.log('删除:', row);
};

const handleExport = () => {
  console.log('导出数据');
};
</script>
```

## 🚀 最佳实践

1. **始终使用卡片容器** - 将页面主要内容包裹在 `.content-card` 中
2. **工具栏独立** - 使用 `.toolbar` 类创建操作区域
3. **状态区分** - 使用状态卡片类传递视觉信息
4. **保持简洁** - 不要过度嵌套，保持结构清晰
5. **响应式设计** - 确保在不同屏幕尺寸下都能正常显示

## 🎨 预设配色方案

主题设置中提供了多种预设配色方案：

- **SxDevOps 浅色** - 默认方案，清新简洁
- **SxDevOps 暖色** - 使用暖色调渐变
- **SxDevOps 冷色** - 使用冷色调渐变
- **自定义** - 完全自定义所有颜色

所有方案都确保卡片内部元素的配色协调统一！
