<script setup lang="ts">
import { ref, watch } from 'vue';
import { useThemeStore } from '@/store/modules/theme';
import { $t } from '@/locales';
import SettingItem from '../components/setting-item.vue';

defineOptions({ name: 'SiderColor' });

const themeStore = useThemeStore();
const customColor = ref(themeStore.sider.customColor || 'rgb(30, 41, 59)');
const useCustomColor = ref(themeStore.sider.useCustomColor || false);

// 侧边栏颜色渐变配置
const useSiderGradient = ref(themeStore.sider.useSiderGradient || false);
const siderGradientStart = ref(themeStore.sider.siderGradientStart || 'rgba(241, 246, 255, 0.98)');
const siderGradientEnd = ref(themeStore.sider.siderGradientEnd || 'rgba(232, 240, 255, 0.9)');

// Logo区域渐变颜色配置
const useLogoGradient = ref(themeStore.sider.useLogoGradient || false);
const logoGradientStart = ref(themeStore.sider.logoGradientStart || 'rgba(241, 246, 255, 0.98)');
const logoGradientEnd = ref(themeStore.sider.logoGradientEnd || 'rgba(232, 240, 255, 0.9)');

// 深色侧边栏预设颜色（深色系）
const swatches: string[] = [
  'rgb(30, 41, 59)', // 蓝灰
  'rgb(15, 23, 42)', // 深蓝灰
  'rgb(30, 30, 40)', // 深灰蓝
  'rgb(51, 65, 85)', // 中蓝灰
  'rgb(71, 85, 105)', // 灰蓝
  'rgb(20, 20, 30)', // 近黑蓝
  'rgb(26, 32, 44)', // 深岩灰
  'rgb(33, 43, 67)', // 深紫灰
  'rgb(45, 55, 72)' // 蓝灰
];

// Logo区域渐变预设颜色（浅色系）
const logoGradientSwatches: string[] = [
  '#f1f6fffa', // 浅蓝白
  '#e8f0ffe6', // 淡蓝
  '#f5f8ff', // 极浅蓝
  '#e8f0ffff', // 淡蓝色
  '#f0f5ffff', // 浅蓝色
  '#e5ecffff', // 浅蓝灰
  '#f8fbffff', // 近白蓝
  '#e0eaffff', // 淡紫蓝
  '#f0f4ffff', // 浅蓝白
  '#eef2ffff' // 浅灰蓝
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

// 深色侧边栏开关变化：与渐变互斥
function handleInvertedChange(value: boolean | string | number) {
  const inverted = value as boolean;
  themeStore.setSiderInverted(inverted);

  // 互斥逻辑：开启深色侧边栏时，关闭所有渐变
  if (inverted) {
    if (useSiderGradient.value) {
      useSiderGradient.value = false;
      themeStore.setSiderGradient(false);
    }
    if (useLogoGradient.value) {
      useLogoGradient.value = false;
      themeStore.setSiderLogoGradient(false);
    }
  }
}

// 侧边栏颜色渐变配置处理函数
function handleUseSiderGradientChange(value: boolean | string | number) {
  const boolValue = value as boolean;

  // 互斥逻辑：开启菜单区域渐变时，关闭深色侧边栏
  if (boolValue && themeStore.sider.inverted) {
    themeStore.setSiderInverted(false);
  }

  useSiderGradient.value = boolValue;
  themeStore.setSiderGradient(boolValue, siderGradientStart.value, siderGradientEnd.value);
}

function handleSiderGradientStartChange(color: string | null) {
  if (color !== null) {
    siderGradientStart.value = color;
    if (useSiderGradient.value) {
      themeStore.setSiderGradient(true, color, siderGradientEnd.value);
    }
  }
}

function handleSiderGradientEndChange(color: string | null) {
  if (color !== null) {
    siderGradientEnd.value = color;
    if (useSiderGradient.value) {
      themeStore.setSiderGradient(true, siderGradientStart.value, color);
    }
  }
}

// Logo区域渐变配置处理函数
function handleUseLogoGradientChange(value: boolean | string | number) {
  const boolValue = value as boolean;

  // 互斥逻辑：开启Logo区域渐变时，关闭深色侧边栏
  if (boolValue && themeStore.sider.inverted) {
    themeStore.setSiderInverted(false);
  }

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
  newSider => {
    if (newSider.customColor !== customColor.value) {
      customColor.value = newSider.customColor || 'rgb(30, 41, 59)';
    }
    if (newSider.useCustomColor !== undefined && newSider.useCustomColor !== useCustomColor.value) {
      useCustomColor.value = newSider.useCustomColor;
    }
    // 监听侧边栏颜色渐变配置变化
    if (newSider.useSiderGradient !== undefined && newSider.useSiderGradient !== useSiderGradient.value) {
      useSiderGradient.value = newSider.useSiderGradient;
    }
    if (newSider.siderGradientStart !== undefined && newSider.siderGradientStart !== siderGradientStart.value) {
      siderGradientStart.value = newSider.siderGradientStart || 'rgba(241, 246, 255, 0.98)';
    }
    if (newSider.siderGradientEnd !== undefined && newSider.siderGradientEnd !== siderGradientEnd.value) {
      siderGradientEnd.value = newSider.siderGradientEnd || 'rgba(232, 240, 255, 0.9)';
    }
    // 监听logo区域渐变配置变化
    if (newSider.useLogoGradient !== undefined && newSider.useLogoGradient !== useLogoGradient.value) {
      useLogoGradient.value = newSider.useLogoGradient;
    }
    if (newSider.logoGradientStart !== undefined && newSider.logoGradientStart !== logoGradientStart.value) {
      logoGradientStart.value = newSider.logoGradientStart || 'rgba(241, 246, 255, 0.98)';
    }
    if (newSider.logoGradientEnd !== undefined && newSider.logoGradientEnd !== logoGradientEnd.value) {
      logoGradientEnd.value = newSider.logoGradientEnd || 'rgba(232, 240, 255, 0.9)';
    }
  },
  { deep: true }
);
</script>

<template>
  <ElDivider>{{ $t('theme.sider.title') }}</ElDivider>
  <div class="flex-col-stretch gap-12px">
    <SettingItem :label="$t('theme.sider.showIcon')">
      <ElSwitch v-model="themeStore.sider.showIcon" @change="handleShowIconChange" />
    </SettingItem>

    <!-- 模式1：深色侧边栏（单色） -->
    <div class="flex-col-stretch gap-12px">
      <SettingItem :label="$t('theme.sider.inverted')">
        <ElSwitch v-model="themeStore.sider.inverted" @change="handleInvertedChange" />
      </SettingItem>

      <!-- 自定义单一颜色（只在深色侧边栏开启时显示） -->
      <template v-if="themeStore.sider.inverted">
        <SettingItem :label="$t('theme.sider.useCustomColor')">
          <ElSwitch v-model="useCustomColor" @change="handleUseCustomColorChange" />
        </SettingItem>
        <SettingItem v-if="useCustomColor" :label="$t('theme.sider.customColor')">
          <ElColorPicker
            v-model="customColor"
            class="w-40px"
            :show-alpha="false"
            :predefine="swatches"
            @change="handleColorChange"
          />
        </SettingItem>
      </template>
    </div>

    <!-- 模式2：侧边栏渐变（渐变色） -->
    <div class="flex-col-stretch gap-12px">
      <!-- Logo区域渐变 -->
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

      <!-- 菜单区域渐变 -->
      <SettingItem :label="$t('theme.sider.useSiderGradient')">
        <ElSwitch v-model="useSiderGradient" @change="handleUseSiderGradientChange" />
      </SettingItem>
      <SettingItem v-if="useSiderGradient" :label="$t('theme.sider.siderGradientStart')">
        <ElColorPicker
          v-model="siderGradientStart"
          class="w-40px"
          :show-alpha="true"
          :predefine="logoGradientSwatches"
          @change="handleSiderGradientStartChange"
        />
      </SettingItem>
      <SettingItem v-if="useSiderGradient" :label="$t('theme.sider.siderGradientEnd')">
        <ElColorPicker
          v-model="siderGradientEnd"
          class="w-40px"
          :show-alpha="true"
          :predefine="logoGradientSwatches"
          @change="handleSiderGradientEndChange"
        />
      </SettingItem>
    </div>
  </div>
</template>

<style scoped></style>
