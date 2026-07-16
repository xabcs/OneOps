<script setup lang="ts">
import { ref, watch } from 'vue';
import { useThemeStore } from '@/store/modules/theme';
import { $t } from '@/locales';
import SettingItem from '../components/setting-item.vue';

defineOptions({ name: 'HeaderColor' });

const themeStore = useThemeStore();
const customColor = ref(themeStore.header.customColor || 'rgb(248, 251, 255)');
const useCustomColor = ref(themeStore.header.useCustomColor || false);

// Header颜色渐变配置
const useHeaderGradient = ref(themeStore.header.useHeaderGradient || false);
const headerGradientStart = ref(themeStore.header.headerGradientStart || 'rgba(232, 237, 255, 0.98)');
const headerGradientEnd = ref(themeStore.header.headerGradientEnd || 'rgba(232, 240, 251, 0.94)');

// Header区域预设颜色（浅色系）
const headerGradientSwatches: string[] = [
  '#f8fbfffa',      // 浅蓝白
  '#f3f7ffe6',      // 淡蓝
  '#f8fafc',        // 极浅蓝
  '#f8f0ffff',      // 淡蓝色
  '#f0f5ffff',      // 浅蓝色
  '#f5f8ffff',      // 浅蓝灰
  '#fafbffff',      // 近白蓝
  '#f0f4ffff',      // 淡紫蓝
  '#f8fcffff',      // 浅蓝白
  '#f5f8ffff'       // 浅灰蓝
];

function handleUseCustomColorChange(value: boolean | string | number) {
  const boolValue = value as boolean;
  useCustomColor.value = boolValue;
  themeStore.setHeaderCustomColor(boolValue, customColor.value);
}

function handleColorChange(color: string | null) {
  if (color !== null) {
    customColor.value = color;
    if (useCustomColor.value) {
      themeStore.setHeaderCustomColor(true, color);
    }
  }
}

// Header颜色渐变配置处理函数
function handleUseHeaderGradientChange(value: boolean | string | number) {
  const boolValue = value as boolean;

  // 互斥逻辑：开启Header渐变时，关闭自定义颜色
  if (boolValue && themeStore.header.useCustomColor) {
    themeStore.setHeaderCustomColor(false);
  }

  useHeaderGradient.value = boolValue;
  themeStore.setHeaderGradient(boolValue, headerGradientStart.value, headerGradientEnd.value);
}

function handleHeaderGradientStartChange(color: string | null) {
  if (color !== null) {
    headerGradientStart.value = color;
    if (useHeaderGradient.value) {
      themeStore.setHeaderGradient(true, color, headerGradientEnd.value);
    }
  }
}

function handleHeaderGradientEndChange(color: string | null) {
  if (color !== null) {
    headerGradientEnd.value = color;
    if (useHeaderGradient.value) {
      themeStore.setHeaderGradient(true, headerGradientStart.value, color);
    }
  }
}

// 监听 store 中的值变化
watch(
  () => themeStore.header,
  (newHeader) => {
    if (newHeader.customColor !== undefined && newHeader.customColor !== customColor.value) {
      customColor.value = newHeader.customColor || 'rgb(248, 251, 255)';
    }
    if (newHeader.useCustomColor !== undefined && newHeader.useCustomColor !== useCustomColor.value) {
      useCustomColor.value = newHeader.useCustomColor;
    }
    // 监听Header颜色渐变配置变化
    if (newHeader.useHeaderGradient !== undefined && newHeader.useHeaderGradient !== useHeaderGradient.value) {
      useHeaderGradient.value = newHeader.useHeaderGradient;
    }
    if (newHeader.headerGradientStart !== undefined && newHeader.headerGradientStart !== headerGradientStart.value) {
      headerGradientStart.value = newHeader.headerGradientStart || 'rgba(232, 237, 255, 0.98)';
    }
    if (newHeader.headerGradientEnd !== undefined && newHeader.headerGradientEnd !== headerGradientEnd.value) {
      headerGradientEnd.value = newHeader.headerGradientEnd || 'rgba(232, 240, 251, 0.94)';
    }
  },
  { deep: true }
);
</script>

<template>
  <ElDivider>{{ $t('theme.header.title') }}</ElDivider>
  <div class="flex-col-stretch gap-12px">
    <!-- Header渐变 -->
    <div class="flex-col-stretch gap-12px">
      <SettingItem :label="$t('theme.header.useHeaderGradient')">
        <ElSwitch v-model="useHeaderGradient" @change="handleUseHeaderGradientChange" />
      </SettingItem>
      <SettingItem v-if="useHeaderGradient" :label="$t('theme.header.headerGradientStart')">
        <ElColorPicker
          v-model="headerGradientStart"
          class="w-40px"
          :show-alpha="true"
          :predefine="headerGradientSwatches"
          @change="handleHeaderGradientStartChange"
        />
      </SettingItem>
      <SettingItem v-if="useHeaderGradient" :label="$t('theme.header.headerGradientEnd')">
        <ElColorPicker
          v-model="headerGradientEnd"
          class="w-40px"
          :show-alpha="true"
          :predefine="headerGradientSwatches"
          @change="handleHeaderGradientEndChange"
        />
      </SettingItem>

      <!-- 自定义单一颜色 -->
      <SettingItem :label="$t('theme.header.useCustomColor')">
        <ElSwitch v-model="useCustomColor" @change="handleUseCustomColorChange" />
      </SettingItem>
      <SettingItem v-if="useCustomColor" :label="$t('theme.header.customColor')">
        <ElColorPicker
          v-model="customColor"
          class="w-40px"
          :show-alpha="false"
          :predefine="headerGradientSwatches"
          @change="handleColorChange"
        />
      </SettingItem>
    </div>
  </div>
</template>

<style scoped></style>
