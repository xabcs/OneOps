<script setup lang="ts">
  import { computed } from 'vue';
  import { GLOBAL_SIDER_MENU_ID } from '@/constants/app';
  import { useAppStore } from '@/store/modules/app';
  import { useThemeStore } from '@/store/modules/theme';
  import GlobalLogo from '../global-logo/index.vue';

  defineOptions({ name: 'GlobalSider' });

  const appStore = useAppStore();
  const themeStore = useThemeStore();

  const isVerticalMix = computed(() => themeStore.layout.mode === 'vertical-mix');
  const isHorizontalMix = computed(() => themeStore.layout.mode === 'horizontal-mix');

  // 使用UI7.0的逻辑：只在非深色模式、无渐变、且用户开启深色侧边栏时，才使用inverted样式
  const darkMenu = computed(() => {
    // 如果有任意渐变开启，不使用inverted模式
    if (themeStore.sider.useLogoGradient || themeStore.sider.useSiderGradient) {
      return false;
    }
    // 只有在非深色模式且用户开启深色侧边栏时，才使用inverted样式
    return !themeStore.darkMode && !isHorizontalMix.value && themeStore.sider.inverted;
  });

  const showLogo = computed(() => !isVerticalMix.value && !isHorizontalMix.value);
  const menuWrapperClass = computed(() => (showLogo.value ? 'flex-1-hidden' : 'h-full'));

  // 计算Logo区域的自定义背景
  const logoAreaStyle = computed(() => {
    // 如果有Logo区域渐变，直接应用（不需要darkMenu）
    if (themeStore.sider.useLogoGradient) {
      return {
        background: `linear-gradient(180deg, ${themeStore.sider.logoGradientStart || 'rgb(30, 41, 59)'} 0%, ${themeStore.sider.logoGradientEnd || 'rgb(15, 23, 42)'} 100%)`
      };
    }

    // 如果在深色侧边栏模式且有自定义颜色，应用自定义颜色
    if (darkMenu.value && themeStore.sider.useCustomColor && themeStore.sider.customColor) {
      return {
        backgroundColor: themeStore.sider.customColor
      };
    }

    return {};
  });

  // 计算菜单区域的自定义背景
  const menuAreaStyle = computed(() => {
    // 如果有菜单区域渐变，直接应用（不需要darkMenu）
    if (themeStore.sider.useSiderGradient) {
      return {
        background: `linear-gradient(180deg, ${themeStore.sider.siderGradientStart || 'rgb(30, 41, 59)'} 0%, ${themeStore.sider.siderGradientEnd || 'rgb(15, 23, 42)'} 100%)`
      };
    }

    // 如果在深色侧边栏模式且有自定义颜色，应用自定义颜色
    if (darkMenu.value && themeStore.sider.useCustomColor && themeStore.sider.customColor) {
      return {
        backgroundColor: themeStore.sider.customColor
      };
    }

    return {};
  });

  // 计算容器样式类
  const siderContainerClass = computed(() => {
    const classes: string[] = [];

    // 只有在深色侧边栏模式下使用自定义颜色时才添加类
    if (darkMenu.value && themeStore.sider.useCustomColor) {
      classes.push('sider-custom-bg');
    }

    // 渐变模式的样式类（用于确保菜单背景透明）
    if (themeStore.sider.useSiderGradient) {
      classes.push('sider-gradient-bg');
    }

    return classes.join(' ');
  });
</script>

<template>
  <DarkModeContainer class="size-full flex-col-stretch shadow-sider" :inverted="darkMenu">
    <!-- Logo区域：使用独立的样式 -->
    <div
      v-if="showLogo"
      class="logo-area-wrapper"
      :class="{ 'logo-gradient': themeStore.sider.useLogoGradient }"
      :style="logoAreaStyle"
    >
      <GlobalLogo :show-title="!appStore.siderCollapse" :style="{ height: themeStore.header.height + 'px' }" />
    </div>

    <!-- 菜单区域：使用独立的样式 -->
    <div :id="GLOBAL_SIDER_MENU_ID" :class="[menuWrapperClass, siderContainerClass]" :style="menuAreaStyle"></div>
  </DarkModeContainer>
</template>

<style scoped>
  .logo-area-wrapper {
    width: 100%;
    flex-shrink: 0;
  }

  /* 自定义颜色模式 - 确保背景色正确应用 */
  .sider-custom-bg :deep(.el-menu),
  .sider-custom-bg :deep(.el-menu--vertical) {
    background-color: transparent !important;
  }

  /* 渐变模式 - 确保背景色正确应用 */
  .sider-gradient-bg :deep(.el-menu),
  .sider-gradient-bg :deep(.el-menu--vertical) {
    background-color: transparent !important;
  }
</style>
