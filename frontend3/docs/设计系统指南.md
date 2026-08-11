# OneOps 设计系统使用指南

## 🎨 渐变体系应用指南

### 核心设计思想

1. **渐变方向规律**：
   - `135deg` - 对角线渐变（卡片、按钮常用）
   - `145deg` - 立体感渐变（Logo、图标背景）
   - `180deg` - 上下渐变（Header、侧边栏）
   - `90deg` - 左右渐变（分隔线、进度条）

2. **颜色透明度策略**：
   - 起始色高透明度（`fa`, `eb`）
   - 结束色中等透明度（`e6`, `bd`, `d6`）
   - 创造层次感和深度

3. **阴影配合系统**：
   - `inset` 阴影创造内凹效果
   - 渐进阴影强度（sm → md → lg）
   - 悬停时阴影增强

4. **交互反馈模式**：
   - 悬停时颜色加深
   - 配合 `translateY(-1px)` 浮起效果
   - 阴影同步增强

### 组件应用示例

#### 1. 卡片组件

```vue
<template>
  <div class="gradient-card-diagonal">
    <div class="p-4">
      <h3>卡片标题</h3>
      <p>卡片内容</p>
    </div>
  </div>
</template>

<style scoped>
.gradient-card-diagonal {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.gradient-card-diagonal:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(59, 130, 246, 0.20);
}
</style>
```

#### 2. 按钮组件

```vue
<template>
  <button class="gradient-button-light">
    <span class="button-icon">✨</span>
    <span>操作按钮</span>
  </button>
</template>

<style scoped>
.gradient-button-light {
  background: linear-gradient(135deg, #eff6ffeb, #ecfdf5bd);
  border: 1px solid rgba(96, 165, 250, 0.24);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
  color: #3b82f6;
  border-radius: 10px;
  padding: 8px 16px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.gradient-button-light:hover {
  background: linear-gradient(135deg, #dbeafefa, #d1fae5d6);
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(59, 130, 246, 0.20);
}
</style>
```

#### 3. 面板组件

```vue
<template>
  <div class="panel-gradient-light">
    <div class="panel-header">
      <h3>面板标题</h3>
    </div>
    <div class="panel-content">
      面板内容
    </div>
  </div>
</template>

<style scoped>
.panel-gradient-light {
  background: linear-gradient(135deg, #fbfdff, #f7faff 52%, #f9fbfd);
  border: 1px solid rgba(36, 91, 219, 0.09);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
  padding: 16px;
}

.panel-header {
  border-bottom: 1px solid rgba(148, 163, 184, 0.12);
  padding-bottom: 12px;
  margin-bottom: 16px;
}
</style>
```

### Element Plus 组件增强

#### 1. 卡片增强

```vue
<template>
  <el-card class="gradient-enhanced">
    <template #header>
      <span>卡片标题</span>
    </template>
    <p>卡片内容</p>
  </el-card>
</template>

<style scoped>
.el-card.gradient-enhanced {
  background: linear-gradient(135deg, #ffffff, #f7fbff) !important;
  border: 1px solid rgba(148, 163, 184, 0.12) !important;
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14) !important;
  border-radius: 12px !important;
}

.el-card.gradient-enhanced:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(59, 130, 246, 0.20) !important;
}
</style>
```

#### 2. 按钮增强

```vue
<template>
  <el-button class="gradient-light" type="primary">
    操作按钮
  </el-button>
</template>

<style scoped>
.el-button.gradient-light {
  background: linear-gradient(135deg, #eff6ffeb, #ecfdf5bd) !important;
  border: 1px solid rgba(96, 165, 250, 0.24) !important;
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14) !important;
  transition: all 0.3s ease;
}

.el-button.gradient-light:hover {
  background: linear-gradient(135deg, #dbeafefa, #d1fae5d6) !important;
  transform: translateY(-1px) !important;
  box-shadow: 0 12px 24px rgba(59, 130, 246, 0.20) !important;
}
</style>
```

### 配色变量使用

#### 在组件中使用CSS变量

```vue
<template>
  <div class="custom-card">
    <div class="custom-button">按钮</div>
  </div>
</template>

<style scoped>
.custom-card {
  background: linear-gradient(135deg, var(--oneops-light-white), var(--oneops-light-bg-2));
  border: 1px solid var(--oneops-border-light);
  box-shadow: var(--oneops-shadow-sm);
  border-radius: 12px;
  padding: 16px;
}

.custom-button {
  background: linear-gradient(135deg, var(--oneops-gradient-start-2), var(--oneops-gradient-end-2));
  border: 1px solid var(--oneops-border-brand);
  box-shadow: var(--oneops-shadow-sm);
  color: var(--oneops-brand-accent);
  padding: 8px 16px;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.custom-button:hover {
  transform: translateY(-1px);
  box-shadow: var(--oneops-shadow-md);
}
</style>
```

### 工具类应用

#### 快速应用渐变效果

```vue
<template>
  <!-- 对角线渐变卡片 -->
  <div class="bg-gradient-diagonal border-gradient-light shadow-gradient-sm p-4 rounded-12px">
    内容区域
  </div>

  <!-- Header渐变 -->
  <div class="bg-gradient-header border-b border-gradient-light p-4">
    头部区域
  </div>

  <!-- 按钮渐变 -->
  <button class="bg-gradient-diagonal border-gradient-brand shadow-gradient-sm p-2 rounded-8px">
    按钮文本
  </button>
</template>
```

### 响应式处理

#### 移动端优化

```vue
<template>
  <div class="responsive-card">
    卡片内容
  </div>
</template>

<style scoped>
.responsive-card {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
  border-radius: 12px;
  padding: 16px;
  transition: all 0.3s ease;
}

@media (max-width: 768px) {
  .responsive-card {
    border-radius: 8px;
    padding: 12px;
  }
}
</style>
```

### 深色模式适配

#### 自动主题切换

```vue
<template>
  <div class="theme-card">
    <h3 :style="{ color: 'var(--oneops-text-primary)' }">标题</h3>
    <p :style="{ color: 'var(--oneops-text-secondary)' }">内容</p>
  </div>
</template>

<style scoped>
.theme-card {
  background: linear-gradient(135deg, var(--oneops-light-white), var(--oneops-light-bg-2));
  border: 1px solid var(--oneops-border-light);
  box-shadow: var(--oneops-shadow-sm);
  border-radius: 12px;
  padding: 16px;
}

@media (prefers-color-scheme: dark) {
  .theme-card {
    background: linear-gradient(135deg, #1e293b, #0f172a);
    border-color: rgba(255, 255, 255, 0.1);
  }
}
</style>
```

### 渐变动画效果

#### 微妙动画

```vue
<template>
  <div class="animated-card">
    卡片内容
  </div>
</template>

<style scoped>
.animated-card {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
  border-radius: 12px;
  padding: 16px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.animated-card:hover {
  transform: translateY(-2px) scale(1.01);
  box-shadow: 0 16px 32px rgba(59, 130, 246, 0.25);
  background: linear-gradient(135deg, #fafbff, #f2f7ff);
}
</style>
```

### 性能优化建议

1. **避免过度渐变**：不要在同一页面使用过多不同的渐变效果
2. **使用CSS变量**：便于主题切换和维护
3. **硬件加速**：使用 `transform` 和 `opacity` 进行动画
4. **合理使用阴影**：过多的阴影会影响性能
5. **响应式考虑**：移动端减少复杂渐变效果

### 最佳实践

1. **一致性**：在整个应用中保持相同的渐变风格
2. **可读性**：确保文本在渐变背景上清晰可读
3. **层次感**：通过透明度变化创造深度
4. **交互反馈**：悬停状态要有明显的视觉变化
5. **品牌一致性**：渐变配色要与品牌色保持协调