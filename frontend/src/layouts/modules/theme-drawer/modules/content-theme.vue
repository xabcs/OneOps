<script setup lang="ts">
import { ref, watch } from 'vue';
import { useThemeStore } from '@/store/modules/theme';
import { $t } from '@/locales';
import SettingItem from '../components/setting-item.vue';
import { applyContentTheme } from '@/utils/content-theme';

defineOptions({ name: 'ContentTheme' });

const themeStore = useThemeStore();

// 卡片设置
const cardBg = ref(themeStore.contentTheme.cardBg || '#ffffff');
const cardRadius = ref(themeStore.contentTheme.cardRadius || '12px');
const useCardGradient = ref(themeStore.contentTheme.useCardGradient || false);
const cardGradientStart = ref(themeStore.contentTheme.cardGradientStart || 'rgba(255, 255, 255, 0.98)');
const cardGradientEnd = ref(themeStore.contentTheme.cardGradientEnd || 'rgba(248, 250, 252, 0.94)');

// 表格设置
const tableBorder = ref(themeStore.contentTheme.tableBorder || 'rgba(148, 163, 184, 0.16)');
const tableHeaderBg = ref(themeStore.contentTheme.tableHeaderBg || '#f8fafc');
const tableHoverBg = ref(themeStore.contentTheme.tableHoverBg || '#f8fbff');
const tableRadius = ref(themeStore.contentTheme.tableRadius || '12px');

// 按钮设置
const buttonRadius = ref(themeStore.contentTheme.buttonRadius || '10px');
const buttonDefaultBg = ref(themeStore.contentTheme.buttonDefaultBg || 'rgba(255, 255, 255, 0.9)');
const buttonDefaultColor = ref(themeStore.contentTheme.buttonDefaultColor || '#475569');

// 输入框设置
const inputRadius = ref(themeStore.contentTheme.inputRadius || '12px');
const inputBg = ref(themeStore.contentTheme.inputBg || 'rgba(255, 255, 255, 0.92)');

// 工具栏渐变设置
const useToolbarGradient = ref(themeStore.contentTheme.useToolbarGradient || false);
const toolbarGradientStart = ref(themeStore.contentTheme.toolbarGradientStart || 'rgba(248, 250, 252, 0.92)');
const toolbarGradientEnd = ref(themeStore.contentTheme.toolbarGradientEnd || 'rgba(255, 255, 255, 0.96)');

// 预设颜色
const lightSwatches: string[] = [
  '#ffffff',
  '#f8fafc',
  '#f1f5f9',
  '#e2e8f0',
  '#cbd5e1',
  '#94a3b8',
  '#64748b',
  '#475569',
  '#334155',
  '#1e293b'
];

// 更新设置
function updateSetting(key: keyof App.Theme.ThemeSetting['contentTheme'], value: any) {
  themeStore.setContentTheme(key, value);
  applyContentTheme(themeStore.contentTheme);
}

// 卡片渐变切换
function handleCardGradientChange(value: boolean | string | number) {
  const boolValue = value as boolean;
  useCardGradient.value = boolValue;
  updateSetting('useCardGradient', boolValue);
}

// 卡片渐变颜色变化
function handleCardGradientStartChange(color: string | null) {
  if (color !== null) {
    cardGradientStart.value = color;
    if (useCardGradient.value) {
      updateSetting('cardGradientStart', color);
    }
  }
}

function handleCardGradientEndChange(color: string | null) {
  if (color !== null) {
    cardGradientEnd.value = color;
    if (useCardGradient.value) {
      updateSetting('cardGradientEnd', color);
    }
  }
}

// 工具栏渐变切换
function handleToolbarGradientChange(value: boolean | string | number) {
  const boolValue = value as boolean;
  useToolbarGradient.value = boolValue;
  updateSetting('useToolbarGradient', boolValue);
}

// 工具栏渐变颜色变化
function handleToolbarGradientStartChange(color: string | null) {
  if (color !== null) {
    toolbarGradientStart.value = color;
    if (useToolbarGradient.value) {
      updateSetting('toolbarGradientStart', color);
    }
  }
}

function handleToolbarGradientEndChange(color: string | null) {
  if (color !== null) {
    toolbarGradientEnd.value = color;
    if (useToolbarGradient.value) {
      updateSetting('toolbarGradientEnd', color);
    }
  }
}

// 监听 store 变化
watch(
  () => themeStore.contentTheme,
  (newTheme) => {
    if (newTheme.cardBg && newTheme.cardBg !== cardBg.value) {
      cardBg.value = newTheme.cardBg;
    }
    if (newTheme.cardRadius && newTheme.cardRadius !== cardRadius.value) {
      cardRadius.value = newTheme.cardRadius;
    }
    if (newTheme.useCardGradient !== undefined && newTheme.useCardGradient !== useCardGradient.value) {
      useCardGradient.value = newTheme.useCardGradient;
    }
    if (newTheme.tableBorder && newTheme.tableBorder !== tableBorder.value) {
      tableBorder.value = newTheme.tableBorder;
    }
    if (newTheme.tableHeaderBg && newTheme.tableHeaderBg !== tableHeaderBg.value) {
      tableHeaderBg.value = newTheme.tableHeaderBg;
    }
    if (newTheme.tableRadius && newTheme.tableRadius !== tableRadius.value) {
      tableRadius.value = newTheme.tableRadius;
    }
    if (newTheme.buttonRadius && newTheme.buttonRadius !== buttonRadius.value) {
      buttonRadius.value = newTheme.buttonRadius;
    }
    if (newTheme.inputRadius && newTheme.inputRadius !== inputRadius.value) {
      inputRadius.value = newTheme.inputRadius;
    }
    if (newTheme.useToolbarGradient !== undefined && newTheme.useToolbarGradient !== useToolbarGradient.value) {
      useToolbarGradient.value = newTheme.useToolbarGradient;
    }
  },
  { deep: true }
);
</script>

<template>
  <ElDivider>{{ $t('theme.content.title') }}</ElDivider>
  <div class="flex-col-stretch gap-12px">
    <!-- 卡片设置 -->
    <div class="flex-col-stretch gap-12px">
      <div class="text-14px font-500 text-gray-700 mb-8px">{{ $t('theme.content.card') }}</div>

      <SettingItem :label="$t('theme.content.cardBg')">
        <ElColorPicker
          v-model="cardBg"
          class="w-40px"
          :show-alpha="false"
          :predefine="lightSwatches"
          @change="(color: string | null) => color && updateSetting('cardBg', color)"
        />
      </SettingItem>

      <SettingItem :label="$t('theme.content.cardRadius')">
        <ElInput v-model="cardRadius" class="w-120px" @change="updateSetting('cardRadius', cardRadius)" />
      </SettingItem>

      <SettingItem :label="$t('theme.content.useCardGradient')">
        <ElSwitch v-model="useCardGradient" @change="handleCardGradientChange" />
      </SettingItem>

      <template v-if="useCardGradient">
        <SettingItem :label="$t('theme.content.cardGradientStart')">
          <ElColorPicker
            v-model="cardGradientStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="lightSwatches"
            @change="handleCardGradientStartChange"
          />
        </SettingItem>
        <SettingItem :label="$t('theme.content.cardGradientEnd')">
          <ElColorPicker
            v-model="cardGradientEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="lightSwatches"
            @change="handleCardGradientEndChange"
          />
        </SettingItem>
      </template>
    </div>

    <!-- 表格设置 -->
    <div class="flex-col-stretch gap-12px">
      <div class="text-14px font-500 text-gray-700 mb-8px">{{ $t('theme.content.table') }}</div>

      <SettingItem :label="$t('theme.content.tableHeaderBg')">
        <ElColorPicker
          v-model="tableHeaderBg"
          class="w-40px"
          :show-alpha="false"
          :predefine="lightSwatches"
          @change="(color: string | null) => color && updateSetting('tableHeaderBg', color)"
        />
      </SettingItem>

      <SettingItem :label="$t('theme.content.tableHoverBg')">
        <ElColorPicker
          v-model="tableHoverBg"
          class="w-40px"
          :show-alpha="false"
          :predefine="lightSwatches"
          @change="(color: string | null) => color && updateSetting('tableHoverBg', color)"
        />
      </SettingItem>

      <SettingItem :label="$t('theme.content.tableRadius')">
        <ElInput v-model="tableRadius" class="w-120px" @change="updateSetting('tableRadius', tableRadius)" />
      </SettingItem>
    </div>

    <!-- 按钮设置 -->
    <div class="flex-col-stretch gap-12px">
      <div class="text-14px font-500 text-gray-700 mb-8px">{{ $t('theme.content.button') }}</div>

      <SettingItem :label="$t('theme.content.buttonRadius')">
        <ElInput v-model="buttonRadius" class="w-120px" @change="updateSetting('buttonRadius', buttonRadius)" />
      </SettingItem>
    </div>

    <!-- 输入框设置 -->
    <div class="flex-col-stretch gap-12px">
      <div class="text-14px font-500 text-gray-700 mb-8px">{{ $t('theme.content.input') }}</div>

      <SettingItem :label="$t('theme.content.inputRadius')">
        <ElInput v-model="inputRadius" class="w-120px" @change="updateSetting('inputRadius', inputRadius)" />
      </SettingItem>
    </div>

    <!-- 工具栏渐变 -->
    <div class="flex-col-stretch gap-12px">
      <div class="text-14px font-500 text-gray-700 mb-8px">{{ $t('theme.content.toolbar') }}</div>

      <SettingItem :label="$t('theme.content.useToolbarGradient')">
        <ElSwitch v-model="useToolbarGradient" @change="handleToolbarGradientChange" />
      </SettingItem>

      <template v-if="useToolbarGradient">
        <SettingItem :label="$t('theme.content.toolbarGradientStart')">
          <ElColorPicker
            v-model="toolbarGradientStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="lightSwatches"
            @change="handleToolbarGradientStartChange"
          />
        </SettingItem>
        <SettingItem :label="$t('theme.content.toolbarGradientEnd')">
          <ElColorPicker
            v-model="toolbarGradientEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="lightSwatches"
            @change="handleToolbarGradientEndChange"
          />
        </SettingItem>
      </template>
    </div>
  </div>
</template>

<style scoped></style>
