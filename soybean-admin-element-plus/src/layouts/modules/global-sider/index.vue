<script setup lang="ts">
import { computed } from 'vue';
import { GLOBAL_SIDER_MENU_ID } from '@/constants/app';
import { useAppStore } from '@/store/modules/app';
import { useThemeStore } from '@/store/modules/theme';
import { getRgb } from '@sa/color';
import GlobalLogo from '../global-logo/index.vue';

defineOptions({ name: 'GlobalSider' });

const appStore = useAppStore();
const themeStore = useThemeStore();

const isVerticalMix = computed(() => themeStore.layout.mode === 'vertical-mix');
const isHorizontalMix = computed(() => themeStore.layout.mode === 'horizontal-mix');
const darkMenu = computed(() => !themeStore.darkMode && !isHorizontalMix.value && themeStore.sider.inverted);
const showLogo = computed(() => !isVerticalMix.value && !isHorizontalMix.value);
const menuWrapperClass = computed(() => (showLogo.value ? 'flex-1-hidden' : 'h-full'));

// 获取侧边栏自定义颜色
const siderCustomColor = computed(() => {
  if (darkMenu.value && themeStore.sider.useCustomColor && themeStore.sider.customColor) {
    return themeStore.sider.customColor;
  }
  return undefined;
});

// 计算侧边栏渐变背景
const siderGradientBackground = computed(() => {
  // 如果使用了自定义颜色，则不应用渐变
  if (siderCustomColor.value) {
    return {};
  }

  // 根据主题模式应用渐变
  if (themeStore.darkMode) {
    return {
      background: 'linear-gradient(180deg, rgba(30, 41, 59, 0.98) 0%, rgba(15, 23, 42, 0.95) 100%)'
    };
  } else {
    return {
      background: 'linear-gradient(180deg, rgba(241, 246, 255, 0.98) 0%, rgba(232, 240, 255, 0.90) 100%)'
    };
  }
});

// 计算悬停颜色
const hoverBackgroundColor = computed(() => {
  if (siderCustomColor.value) {
    try {
      const { r, g, b } = getRgb(siderCustomColor.value);
      // 计算亮度
      const brightness = (r * 299 + g * 587 + b * 114) / 1000;

      if (brightness < 128) {
        // 深色背景，使用白色悬停
        return `rgba(255, 255, 255, 0.1)`;
      } else {
        // 浅色背景，使用黑色悬停
        return `rgba(0, 0, 0, 0.05)`;
      }
    } catch (e) {
      // 如果颜色解析失败，使用默认值
      return 'rgba(255, 255, 255, 0.1)';
    }
  }
  return undefined;
});

// 计算选中颜色
const selectedBackgroundColor = computed(() => {
  if (siderCustomColor.value) {
    return undefined;
  }

  // 使用主题色渐变作为选中背景
  if (themeStore.darkMode) {
    return 'linear-gradient(90deg, rgba(59, 130, 246, 0.2) 0%, transparent 100%)';
  } else {
    return 'linear-gradient(90deg, rgba(59, 130, 246, 0.15) 0%, transparent 100%)';
  }
});

// 应用自定义颜色到 CSS 变量
const siderStyle = computed(() => {
  const styles: Record<string, string> = {
    ...siderGradientBackground.value
  };

  if (siderCustomColor.value) {
    styles['--sider-custom-bg'] = siderCustomColor.value;
  }

  if (hoverBackgroundColor.value) {
    styles['--sider-custom-hover-bg'] = hoverBackgroundColor.value;
  }

  if (selectedBackgroundColor.value) {
    styles['--sider-selected-bg'] = selectedBackgroundColor.value;
  }

  return styles;
});

// 侧边栏容器样式类
const siderContainerClass = computed(() => {
  return siderCustomColor.value ? 'sider-custom-color' : 'sider-gradient-bg';
});
</script>

<template>
  <DarkModeContainer
    class="size-full flex-col-stretch shadow-sider global-sider-enhanced"
    :class="siderContainerClass"
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

<style scoped>
/* ========== 侧边栏渐变背景 ========== */
.sider-gradient-bg {
  border-right: 1px solid rgba(148, 163, 184, 0.14);
  box-shadow: 8px 0 24px rgba(15, 23, 42, 0.08);
}

/* ========== 自定义颜色模式 ========== */
.sider-custom-color :deep(.el-menu),
.sider-custom-color :deep(.el-menu--vertical),
.sider-custom-color :deep(.el-sub-menu__title),
.sider-custom-color :deep(.el-menu-item),
.sider-custom-color :deep(.el-menu-item.is-active),
.sider-custom-color :deep(.simplebar-content-wrapper) {
  background-color: var(--sider-custom-bg) !important;
}

/* ========== 渐变背景模式 - 菜单透明 ========== */
.sider-gradient-bg :deep(.el-menu),
.sider-gradient-bg :deep(.el-menu--vertical),
.sider-gradient-bg :deep(.el-sub-menu__title),
.sider-gradient-bg :deep(.el-menu-item) {
  background-color: transparent !important;
}

/* ========== 菜单项悬停效果 - 使用渐变 ========== */
.sider-gradient-bg :deep(.el-menu-item:hover::before),
.sider-gradient-bg :deep(.el-sub-menu__title:hover::before) {
  background: linear-gradient(90deg, rgba(64, 73, 101, 0.08) 0%, transparent 50%) !important;
  border-radius: var(--border-radius-medium, 8px) !important;
}

.sider-custom-color :deep(.el-menu-item:hover),
.sider-custom-color :deep(.el-sub-menu__title:hover) {
  background-color: var(--sider-custom-bg) !important;
}

.sider-custom-color :deep(.el-menu-item:hover::before),
.sider-custom-color :deep(.el-sub-menu__title:hover::before) {
  background-color: var(--sider-custom-hover-bg) !important;
}

/* ========== 菜单激活状态 - 使用主题色渐变 ========== */
.sider-gradient-bg :deep(.el-menu-item.is-active::before) {
  background: var(--sider-selected-bg) !important;
  border-radius: var(--border-radius-medium, 8px) !important;
}

.sider-custom-color :deep(.el-menu-item.is-active::before) {
  background-color: var(--el-menu-hover-bg-color) !important;
}

/* ========== 菜单分隔线渐变 ========== */
.sider-gradient-bg :deep(.el-menu-item::after) {
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(191, 219, 255, 0.9) 14%,
    rgba(226, 232, 240, 0.95) 50%,
    rgba(191, 219, 255, 0.9) 86%,
    transparent 100%
  ) !important;
  height: 1px !important;
}

/* ========== 滚动条渐变样式 ========== */
.sider-gradient-bg :deep(.simplebar-track) {
  background: transparent !important;
}

.sider-gradient-bg :deep(.simplebar-scrollbar) {
  background: linear-gradient(180deg, rgba(203, 213, 225, 0.3) 0%, rgba(203, 213, 225, 0.5) 100%) !important;
  border-radius: 10px !important;
}

.sider-gradient-bg :deep(.simplebar-scrollbar:hover) {
  background: linear-gradient(180deg, rgba(59, 130, 246, 0.4) 0%, rgba(59, 130, 246, 0.6) 100%) !important;
}

/* ========== 子菜单背景渐变 ========== */
.sider-gradient-bg :deep(.el-sub-menu__popup) {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.98)) !important;
  border: 1px solid rgba(148, 163, 184, 0.12) !important;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08) !important;
  border-radius: 8px !important;
}

/* ========== 深色模式适配 ========== */
@media (prefers-color-scheme: dark) {
  .sider-gradient-bg :deep(.el-sub-menu__popup) {
    background: linear-gradient(135deg, rgba(30, 41, 59, 0.95), rgba(15, 23, 42, 0.98)) !important;
    border-color: rgba(255, 255, 255, 0.1) !important;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3) !important;
  }

  .sider-gradient-bg :deep(.simplebar-scrollbar) {
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.1) 0%, rgba(255, 255, 255, 0.2) 100%) !important;
  }

  .sider-gradient-bg :deep(.simplebar-scrollbar:hover) {
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.15) 0%, rgba(255, 255, 255, 0.25) 100%) !important;
  }
}

/* ========== 响应式优化 ========== */
@media (max-width: 768px) {
  .sider-gradient-bg {
    box-shadow: 4px 0 16px rgba(15, 23, 42, 0.06);
  }
}
</style>
