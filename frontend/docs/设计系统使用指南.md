# OneOps 设计系统快速应用指南

## 🚀 快速开始

### 1. 在现有组件中应用渐变效果

#### 方式一：使用工具类（最简单）

```vue
<template>
  <!-- 直接在模板中使用工具类 -->
  <div class="bg-gradient-diagonal border-gradient-light shadow-gradient-sm p-4 rounded-12px">
    内容区域
  </div>
</template>
```

#### 方式二：使用设计系统组件

```vue
<template>
  <script setup>
  import GradientCard from '@/components/common/gradient-card.vue';
  import GradientButton from '@/components/common/gradient-button.vue';
  import GradientPanel from '@/components/common/gradient-panel.vue';
  </script>

  <!-- 使用预制的渐变组件 -->
  <GradientCard variant="diagonal" :enable-hover="true">
    <h3>卡片标题</h3>
    <p>卡片内容</p>
  </GradientCard>

  <GradientButton variant="light" @click="handleClick">
    操作按钮
  </GradientButton>

  <GradientPanel variant="light" title="面板标题">
    面板内容
  </GradientPanel>
</template>
```

#### 方式三：自定义渐变样式

```vue
<template>
  <div class="custom-gradient-card">
    自定义渐变卡片
  </div>
</template>

<style scoped>
.custom-gradient-card {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
  border-radius: 12px;
  padding: 16px;
  transition: all 0.3s ease;
}

.custom-gradient-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(59, 130, 246, 0.20);
}
</style>
```

### 2. 升级现有 Element Plus 组件

#### 卡片升级

```vue
<template>
  <!-- 原来的卡片 -->
  <el-card class="old-card">
    <template #header>卡片标题</template>
    卡片内容
  </el-card>

  <!-- 升级后的卡片 -->
  <el-card class="gradient-enhanced">
    <template #header>卡片标题</template>
    卡片内容
  </el-card>
</template>

<style scoped>
/* 只需添加这个class即可自动应用渐变效果 */
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

#### 按钮升级

```vue
<template>
  <!-- 添加渐变效果的按钮 -->
  <el-button class="gradient-light" type="primary">
    操作按钮
  </el-button>

  <!-- 或者使用专门的渐变按钮组件 -->
  <GradientButton variant="light">
    操作按钮
  </GradientButton>
</template>

<style scoped>
.el-button.gradient-light {
  background: linear-gradient(135deg, #eff6ffeb, #ecfdf5bd) !important;
  border: 1px solid rgba(96, 165, 250, 0.24) !important;
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14) !important;
}

.el-button.gradient-light:hover {
  background: linear-gradient(135deg, #dbeafefa, #d1fae5d6) !important;
  transform: translateY(-1px) !important;
  box-shadow: 0 12px 24px rgba(59, 130, 246, 0.20) !important;
}
</style>
```

### 3. 在页面布局中应用渐变

#### 页面头部区域

```vue
<template>
  <div class="page-header-gradient">
    <div class="header-content">
      <h1>页面标题</h1>
      <p>页面描述</p>
    </div>
    <div class="header-actions">
      <GradientButton variant="light">操作按钮</GradientButton>
    </div>
  </div>
</template>

<style scoped>
.page-header-gradient {
  background: linear-gradient(135deg, #fbfdff, #f7faff 52%, #f9fbfd);
  border: 1px solid rgba(36, 91, 219, 0.09);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
  padding: 20px;
  margin-bottom: 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-content h1 {
  color: #1e293b;
  margin: 0 0 8px 0;
}

.header-content p {
  color: #64748b;
  margin: 0;
}
</style>
```

#### 工具栏区域

```vue
<template>
  <div class="toolbar-gradient">
    <div class="toolbar-left">
      <GradientButton variant="light" size="small">刷新</GradientButton>
      <GradientButton variant="light" size="small">筛选</GradientButton>
    </div>
    <div class="toolbar-right">
      <GradientButton variant="brand">导出</GradientButton>
    </div>
  </div>
</template>

<style scoped>
.toolbar-gradient {
  background: linear-gradient(180deg, #f8fafceb, #fffffff5);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  padding: 12px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
```

### 4. 在表格和表单中应用

#### 表格容器

```vue
<template>
  <div class="table-container-gradient">
    <el-table :data="tableData">
      <el-table-column prop="name" label="姓名" />
      <el-table-column prop="age" label="年龄" />
      <el-table-column prop="address" label="地址" />
    </el-table>
  </div>
</template>

<style scoped>
.table-container-gradient {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
  border-radius: 12px;
  padding: 16px;
  overflow: hidden;
}

.table-container-gradient .el-table {
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.08);
}
</style>
```

#### 表单区域

```vue
<template>
  <GradientPanel variant="light" title="用户信息">
    <el-form :model="form" label-width="80px">
      <el-form-item label="姓名">
        <el-input v-model="form.name" class="input-gradient" />
      </el-form-item>
      <el-form-item label="邮箱">
        <el-input v-model="form.email" class="input-gradient" />
      </el-form-item>
      <el-form-item>
        <GradientButton variant="light" type="submit">提交</GradientButton>
      </el-form-item>
    </el-form>
  </GradientPanel>
</template>
```

## 🎨 设计原则

### 配色规律

1. **浅色系为主**：使用白色到浅蓝色的渐变
2. **适度透明**：起始色高透明度（fa, eb），结束色中等透明度（e6, bd）
3. **统一阴影**：使用蓝色系阴影保持一致性
4. **圆角统一**：12px（大组件）/8px小组件/4px细微元素

### 交互模式

1. **悬停浮起**：`translateY(-1px)` + 阴影增强
2. **颜色加深**：悬停时渐变颜色略微加深
3. **平滑过渡**：`transition: all 0.3s ease`
4. **渐进反馈**：从hover到active的渐进变化

### 应用场景

#### 适合使用渐变的场景

✅ **推荐使用**：
- 卡片、面板容器
- 按钮、操作元素
- 页面头部、工具栏
- Logo区域、品牌展示
- 数据可视化组件

❌ **不推荐使用**：
- 文本输入框背景（保持纯色）
- 代码编辑器区域
- 大面积背景（避免视觉疲劳）
- 数据表格行（保持简洁）

## 📱 响应式处理

```vue
<style scoped>
/* 移动端优化 */
@media (max-width: 768px) {
  .gradient-card {
    border-radius: 8px;      /* 减小圆角 */
    padding: 12px;           /* 减少内边距 */
  }

  .gradient-button {
    border-radius: 8px;
    padding: 6px 12px;
    font-size: 13px;         /* 减小字体 */
  }
}
</style>
```

## 🌙 深色模式适配

```vue
<style scoped>
/* 自动适配深色模式 */
.gradient-card {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
}

@media (prefers-color-scheme: dark) {
  .gradient-card {
    background: linear-gradient(135deg, #1e293b, #0f172a);
    border-color: rgba(255, 255, 255, 0.1);
  }
}
</style>
```

## 🔧 性能优化建议

1. **避免过度使用**：不要在单个页面使用超过5-6种不同的渐变效果
2. **使用CSS变量**：便于主题切换和批量修改
3. **硬件加速**：优先使用 `transform` 和 `opacity` 动画
4. **合理阴影**：避免过多复杂阴影影响性能
5. **响应式考虑**：移动端简化渐变效果

## 📊 实际应用建议

### 渐进式升级策略

1. **第一阶段**：在新页面中使用渐变组件
2. **第二阶段**：升级高频使用的现有组件
3. **第三阶段**：统一全站视觉风格

### 维护策略

1. **使用CSS变量**：统一管理配色
2. **组件化封装**：复用渐变组件
3. **文档更新**：及时记录新的设计模式
4. **用户反馈**：根据用户使用体验优化