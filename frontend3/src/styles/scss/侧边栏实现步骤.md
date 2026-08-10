# 🎨 OneOps 侧边栏样式实施步骤

## 📋 实施概述

本指南提供了在 OneOps 项目中逐步应用 SxDevOps 侧边栏样式的详细步骤。

## 🚀 分步实施指南

### 第一步：备份现有样式

在开始之前，建议备份当前的样式文件：

```bash
# 备份现有样式文件
cp src/styles/scss/global.scss src/styles/scss/global.scss.backup
cp -r src/styles/scss src/styles/scss.backup
```

### 第二步：集成新的样式系统

#### 1. 更新主样式文件

编辑 `src/styles/scss/global.scss`，添加新的样式导入：

```scss
// 这个文件被 Vite 的 additionalData 自动注入到每个组件中
// 所以不需要在这里 self-import，避免循环依赖
@forward 'element-plus';
@forward 'scrollbar';

// ========== 导入 SxDevOps 配色系统 ==========
@use './sxdevops-theme.scss';
@use './layout-theme.scss';
@use './interaction-states.scss';
@use './sidebar-enhanced.scss';

// ========== 导入项目现有样式 ==========
@use './design-system.scss';
@use './element-plus.scss';
@use './compact-theme.scss';
@use './terminal-workbench.scss';
```

#### 2. 或者使用新的全局样式文件

如果你想使用新的全局样式文件，可以替换为：

```bash
# 备份原文件
mv src/styles/scss/global.scss src/styles/scss/global.scss.old

# 使用新的全局样式
cp src/styles/scss/global-enhanced.scss src/styles/scss/global.scss
```

### 第三步：更新侧边栏组件

#### 1. 编辑 `src/layouts/modules/global-sider/index.vue`

在 `<template>` 中添加新的样式类：

```vue
<template>
  <DarkModeContainer
    class="size-full flex-col-stretch global-sider-enhanced shadow-sider"
    :class="[siderContainerClass, { 'sider-collapsed': appStore.siderCollapse }]"
    :inverted="darkMenu"
    :custom-color="siderCustomColor"
    :style="siderStyle"
  >
    <GlobalLogo
      v-if="showLogo"
      :show-title="!appStore.siderCollapse"
      :style="{ height: themeStore.header.height + 'px' }"
    />
    <div :id="GLOBAL_SIDER_MENU_ID" :class="menuWrapperClass"></div>
  </DarkModeContainer>
</template>
```

#### 2. 更新 `<style>` 部分

移除或注释掉原有的自定义样式，让新的样式系统接管：

```vue
<style scoped>
/* 移除以下样式，使用 sidebar-enhanced.scss 中的样式 */

/* .sider-gradient-bg { ... } */
/* .sider-custom-color { ... } */
/* ... 其他自定义样式 */

/* 保留必要的组件特定样式 */
.shadow-sider {
  /* 如果需要自定义阴影效果 */
}
</style>
```

### 第四步：更新 Logo 组件

#### 编辑 `src/layouts/modules/global-logo/index.vue`

```vue
<template>
  <div class="global-logo">
    <div class="logo-icon">
      <slot name="icon">
        <img src="@/assets/logo.svg" alt="Logo" class="brand-mark" />
      </slot>
    </div>
    <div class="logo-copy" v-show="showTitle">
      <slot name="title">
        <span class="logo-text">{{ title }}</span>
        <span class="logo-subtext">{{ subtitle }}</span>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title?: string;
  subtitle?: string;
  showTitle?: boolean;
}

withDefaults(defineProps<Props>(), {
  title: 'OneOps',
  subtitle: '运维平台',
  showTitle: true
});
</script>

<style scoped>
.global-logo {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 14px;
  transition: all 0.3s ease;
}

.logo-icon {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-mark {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.logo-copy {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 4px;
  margin-left: 10px;
  transition: opacity 0.3s ease;
}

/* 响应式优化 */
@media (max-width: 768px) {
  .global-logo {
    padding: 0 8px;
  }

  .logo-copy {
    margin-left: 8px;
  }

  .logo-text {
    font-size: 15px;
  }

  .logo-subtext {
    font-size: 9px;
  }
}
</style>
```

### 第五步：测试和验证

#### 1. 启动开发服务器

```bash
# 在项目根目录运行
npm run dev
```

#### 2. 检查样式效果

- ✅ 侧边栏是否显示渐变背景
- ✅ 菜单项悬停效果是否正常
- ✅ 激活状态是否使用品牌色渐变
- ✅ 折叠功能是否流畅
- ✅ 深色模式是否正常工作

#### 3. 浏览器开发者工具检查

打开浏览器开发者工具，检查：

```css
/* 检查侧边栏背景 */
.global-sider-enhanced {
  background: linear-gradient(180deg, rgba(241, 246, 255, 0.98) 0%, rgba(232, 240, 255, 0.90) 100%);
}

/* 检查菜单激活状态 */
.el-menu-item.is-active::before {
  background: linear-gradient(180deg, #eaf2ff 0%, #deebff 100%);
}
```

## 🎨 自定义调整

### 调整品牌色

如果你想使用不同的品牌色，可以修改 `sxdevops-theme.scss`：

```scss
:root {
  /* 修改品牌色为你的品牌色 */
  --primary: #your-brand-color;
  --primary-light: #your-brand-light-color;
  --primary-dark: #your-brand-dark-color;

  /* 修改品牌渐变 */
  --brand-gradient: linear-gradient(135deg, 
    #your-color-1 0%, 
    #your-color-2 58%, 
    #your-color-3 100%
  );
}
```

### 调整侧边栏宽度

```scss
:root {
  --sidebar-width: 200px;           /* 默认 188px，可以调整为你的需求 */
  --sidebar-collapsed-width: 70px;   /* 默认 68px */
}
```

### 调整动画时长

```scss
:root {
  --transition: all 0.2s ease;       /* 默认 0.3s，可以加快动画 */
}
```

## 🔧 故障排除

### 问题 1：样式冲突

如果发现样式冲突，可以使用更高优先级的选择器：

```scss
/* 使用 :deep() 提高优先级 */
.global-sider-enhanced :deep(.el-menu-item.is-active) {
  background: linear-gradient(180deg, #eaf2ff 0%, #deebff 100%) !important;
}
```

### 问题 2：渐变不显示

检查浏览器是否支持渐变语法，尝试使用标准语法：

```scss
/* 使用标准渐变语法 */
background: linear-gradient(180deg, 
  rgba(241, 246, 255, 0.98) 0%, 
  rgba(232, 240, 255, 0.90) 100%
);
```

### 问题 3：深色模式不工作

确保媒体查询语法正确，并检查系统深色模式设置：

```scss
@media (prefers-color-scheme: dark) {
  /* 深色模式样式 */
}
```

## 📱 移动端测试

在移动设备上测试侧边栏效果：

1. **折叠状态**：确认图标居中显示
2. **触摸区域**：确保菜单项有足够的触摸区域
3. **滑动操作**：测试侧边栏滑动功能
4. **性能**：检查动画是否流畅

## 🎯 完成检查清单

实施完成后，检查以下项目：

- [ ] 侧边栏显示渐变背景
- [ ] Logo 区域样式正确
- [ ] 菜单项悬停效果正常
- [ ] 激活状态使用品牌色渐变
- [ ] 侧边栏折叠功能流畅
- [ ] 子菜单弹出样式正确
- [ ] 深色模式自动适配
- [ ] 移动端显示正常
- [ ] 浏览器兼容性良好
- [ ] 性能表现良好

## 🚀 下一步优化

完成基础实施后，可以考虑以下优化：

### 1. 添加动画效果

```scss
/* 添加菜单项进入动画 */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.el-menu-item {
  animation: slideIn 0.3s ease forwards;
}
```

### 2. 添加微交互

```scss
/* Logo 悬停效果 */
.logo-icon:hover {
  transform: scale(1.05) rotate(5deg);
}
```

### 3. 优化性能

```scss
/* 使用 will-change 优化动画性能 */
.el-menu-item {
  will-change: background-color, transform;
}
```

## 📚 参考资源

- **sxdevops-theme.scss** - 核心 CSS 变量系统
- **sidebar-enhanced.scss** - 侧边栏增强样式
- **sidebar-apply-guide.md** - 详细应用指南
- **color-system-guide.md** - 配色系统使用指南

## 🎉 开始使用

按照本指南的步骤，你就能成功在 OneOps 项目中应用 SxDevOps 的侧边栏样式！

有任何问题，随时询问！🚀