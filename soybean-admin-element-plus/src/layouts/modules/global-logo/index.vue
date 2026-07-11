<script setup lang="ts">
import { computed } from 'vue';
import { useThemeStore } from '@/store/modules/theme';

defineOptions({ name: 'GlobalLogo' });

const themeStore = useThemeStore();

// 计算logo区域的渐变背景 - 仅根据整体深色模式判断
const logoBackgroundClass = computed(() => {
  // 只根据全局深色模式判断，忽略侧边栏反转设置
  return themeStore.darkMode ? 'logo-area-dark' : 'logo-area-light';
});

// 获取logo区域渐变颜色配置
const logoGradientEnabled = computed(() => themeStore.sider.useLogoGradient);
const logoGradientStart = computed(() => themeStore.sider.logoGradientStart || '#f1f6fffa');
const logoGradientEnd = computed(() => themeStore.sider.logoGradientEnd || '#e8f0ffe6');

// 计算渐变背景样式
const logoGradientStyle = computed(() => {
  if (!logoGradientEnabled.value) {
    return {};
  }

  return {
    background: `linear-gradient(180deg, ${logoGradientStart.value} 0%, ${logoGradientEnd.value} 100%)`
  };
});
</script>

<template>
  <RouterLink
    to="/"
    :class="['w-full', 'flex-center', 'flex-row', 'nowrap-hidden', logoBackgroundClass]"
    :style="logoGradientStyle"
  >
    <!-- Logo图标 -->
    <img
      src="/src/assets/svg-icon/msre-free-01-aurora.svg"
      alt="OneOps Logo"
      class="logo-icon"
    />
    <!-- 系统Logo文字 -->
    <SystemLogo class="system-logo-text" />
  </RouterLink>
</template>

<style scoped>
/* Logo区域基础样式 */
RouterLink {
  padding: 8px 8px 4px 8px;
  margin: 6px 8px 0 8px;
  position: relative;
  border-bottom: 1px solid transparent;
  transition: all 0.3s ease;
  border-radius: var(--border-radius-medium, 8px);
}

/* 浅色模式Logo区域基础样式 */
.logo-area-light {
  border-bottom: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.06);
}

/* 深色模式Logo区域渐变 */
.logo-area-dark {
  background: linear-gradient(180deg,
    rgba(30, 41, 59, 0.95) 0%,
    rgba(15, 23, 42, 0.98) 50%,
    rgba(30, 41, 59, 0.95) 100%
  );
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

/* Logo区域悬停效果 - 浅色模式 */
.logo-area-light:hover {
  box-shadow: 0 2px 12px rgba(59, 130, 246, 0.1);
}

/* Logo区域悬停效果 - 深色模式 */
.logo-area-dark:hover {
  background: linear-gradient(180deg,
    rgba(35, 45, 65, 0.98) 0%,
    rgba(20, 28, 47, 1) 50%,
    rgba(35, 45, 65, 0.98) 100%
  );
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.4);
}

/* Logo图标样式 */
.logo-icon {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  object-fit: contain;
  transition: transform 0.3s ease;
}

.logo-icon:hover {
  transform: scale(1.05) rotate(5deg);
}

/* 系统Logo文字样式 */
.system-logo-text {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
  line-height: 1.2;
  white-space: nowrap;
}

/* 浅色模式文字颜色 */
.logo-area-light .system-logo-text {
  color: #1e293b;
}

/* 浅色模式悬停文字颜色 */
.logo-area-light:hover .system-logo-text {
  color: #2563eb;
}

/* 深色模式文字颜色 */
.logo-area-dark .system-logo-text {
  color: #f1f5f9;
}

/* 深色模式悬停文字颜色 */
.logo-area-dark:hover .system-logo-text {
  color: #60a5fa;
}

/* 调整RouterLink布局为横向，与菜单项精确对齐 */
RouterLink {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-start;
  padding: 8px 8px 4px 8px;
  margin: 6px 8px 0 8px;
  position: relative;
  border-bottom: 1px solid transparent;
  transition: all 0.3s ease;
  border-radius: var(--border-radius-medium, 8px);
}

/* 响应式优化 */
@media (max-width: 768px) {
  RouterLink {
    padding: 6px 6px 3px 6px;
    gap: 8px;
    margin: 4px 6px 0 6px;
  }

  .logo-icon {
    width: 32px;
    height: 32px;
  }

  .system-logo-text {
    font-size: 12px;
  }
}

/* 当侧边栏折叠时隐藏文字，只显示图标 */
@media (max-width: 56px) {
  .system-logo-text {
    display: none;
  }

  RouterLink {
    padding: 8px 6px;
    margin: 6px 4px 0 4px;
    justify-content: center;
  }
}

/* Logo图标背景渐变效果 - 参考HTML文件 */
.logo-area-light :deep(.system-logo),
.logo-area-dark :deep(.system-logo) {
  background: linear-gradient(145deg, #eef4ff, #f8fbff) !important;
  border: 1px solid rgba(96, 165, 250, 0.24) !important;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.14) !important;
  border-radius: 11px !important;
}
</style>