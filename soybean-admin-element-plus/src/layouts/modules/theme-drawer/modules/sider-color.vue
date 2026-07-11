<script setup lang="ts">
import { ref, watch } from 'vue';
import { useThemeStore } from '@/store/modules/theme';
import { $t } from '@/locales';
import SettingItem from '../components/setting-item.vue';

defineOptions({ name: 'SiderColor' });

const themeStore = useThemeStore();
const customColor = ref(themeStore.sider.customColor);
const useCustomColor = ref(themeStore.sider.useCustomColor || false);

// Logo区域渐变颜色配置
const useLogoGradient = ref(themeStore.sider.useLogoGradient || false);
const logoGradientStart = ref(themeStore.sider.logoGradientStart || '#f1f6fffa');
const logoGradientEnd = ref(themeStore.sider.logoGradientEnd || '#e8f0ffe6');

// 预设颜色
const swatches: string[] = [
  'rgb(30, 30, 40)',   // 深灰蓝
  'rgb(20, 20, 30)',   // 深蓝黑
  'rgb(40, 40, 50)',   // 浅深灰
  'rgb(15, 23, 42)',   // 深蓝灰
  'rgb(30, 41, 59)',   // 蓝灰
  'rgb(51, 65, 85)',   // 中蓝灰
  'rgb(15, 15, 20)',   // 近黑
  'rgb(10, 10, 15)',   // 纯黑
  'rgb(45, 55, 72)',   // 蓝灰
  'rgb(26, 32, 44)'    // 深岩灰
];

// Logo区域渐变预设颜色（浅色系）
const logoGradientSwatches: string[] = [
  '#f1f6fffa',      // 浅蓝白
  '#e8f0ffe6',      // 淡蓝
  '#f5f8ff',        // 极浅蓝
  '#e8f0ffff',      // 淡蓝色
  '#f0f5ffff',      // 浅蓝色
  '#e5ecffff',      // 浅蓝灰
  '#f8fbffff',      // 近白蓝
  '#e0eaffff',      // 淡紫蓝
  '#f0f4ffff',      // 浅蓝白
  '#eef2ffff'       // 浅灰蓝
];

function handleUseCustomColorChange(value: boolean | string | number) {
  const boolValue = value as boolean;
  useCustomColor.value = boolValue;
  themeStore.setSiderCustomColor(boolValue, customColor.value);
}

function handleColorChange(color: string | null) {
  if (color !== null) {
    customColor.value = color;
    if (useCustomColor.value) {
      themeStore.setSiderCustomColor(true, color);
    }
  }
}

function handleShowIconChange(value: boolean | string | number) {
  themeStore.setSiderShowIcon(value as boolean);
}

// Logo区域渐变配置处理函数
function handleUseLogoGradientChange(value: boolean | string | number) {
  const boolValue = value as boolean;
  useLogoGradient.value = boolValue;
  themeStore.setSiderLogoGradient(boolValue, logoGradientStart.value, logoGradientEnd.value);
}

function handleLogoGradientStartChange(color: string | null) {
  if (color !== null) {
    logoGradientStart.value = color;
    if (useLogoGradient.value) {
      themeStore.setSiderLogoGradient(true, color, logoGradientEnd.value);
    }
  }
}

function handleLogoGradientEndChange(color: string | null) {
  if (color !== null) {
    logoGradientEnd.value = color;
    if (useLogoGradient.value) {
      themeStore.setSiderLogoGradient(true, logoGradientStart.value, color);
    }
  }
}

// 监听 store 中的值变化
watch(
  () => themeStore.sider,
  (newSider) => {
    if (newSider.customColor !== customColor.value) {
      customColor.value = newSider.customColor || 'rgb(30, 30, 40)';
    }
    if (newSider.useCustomColor !== undefined && newSider.useCustomColor !== useCustomColor.value) {
      useCustomColor.value = newSider.useCustomColor;
    }
    // 监听logo区域渐变配置变化
    if (newSider.useLogoGradient !== undefined && newSider.useLogoGradient !== useLogoGradient.value) {
      useLogoGradient.value = newSider.useLogoGradient;
    }
    if (newSider.logoGradientStart !== undefined && newSider.logoGradientStart !== logoGradientStart.value) {
      logoGradientStart.value = newSider.logoGradientStart;
    }
    if (newSider.logoGradientEnd !== undefined && newSider.logoGradientEnd !== logoGradientEnd.value) {
      logoGradientEnd.value = newSider.logoGradientEnd;
    }
  },
  { deep: true }
);
</script>

<template>
  <ElDivider>{{ $t('theme.sider.title') }}</ElDivider>
  <div class="flex-col-stretch gap-12px">
    <SettingItem :label="$t('theme.sider.inverted')">
      <ElSwitch v-model="themeStore.sider.inverted" @change="themeStore.setSiderInverted" />
    </SettingItem>
    <SettingItem :label="$t('theme.sider.showIcon')">
      <ElSwitch v-model="themeStore.sider.showIcon" @change="handleShowIconChange" />
    </SettingItem>
    <SettingItem v-if="themeStore.sider.inverted" :label="$t('theme.sider.useCustomColor')">
      <ElSwitch v-model="useCustomColor" @change="handleUseCustomColorChange" />
    </SettingItem>
    <SettingItem v-if="themeStore.sider.inverted && useCustomColor" :label="$t('theme.sider.customColor')">
      <ElColorPicker
        v-model="customColor"
        class="w-40px"
        :show-alpha="false"
        :predefine="swatches"
        @change="handleColorChange"
      />
    </SettingItem>

    <!-- Logo区域渐变配置 -->
    <SettingItem :label="$t('theme.sider.useLogoGradient')">
      <ElSwitch v-model="useLogoGradient" @change="handleUseLogoGradientChange" />
    </SettingItem>
    <SettingItem v-if="useLogoGradient" :label="$t('theme.sider.logoGradientStart')">
      <ElColorPicker
        v-model="logoGradientStart"
        class="w-40px"
        :show-alpha="true"
        :predefine="logoGradientSwatches"
        @change="handleLogoGradientStartChange"
      />
    </SettingItem>
    <SettingItem v-if="useLogoGradient" :label="$t('theme.sider.logoGradientEnd')">
      <ElColorPicker
        v-model="logoGradientEnd"
        class="w-40px"
        :show-alpha="true"
        :predefine="logoGradientSwatches"
        @change="handleLogoGradientEndChange"
      />
    </SettingItem>
  </div>
</template>

<style scoped></style>
