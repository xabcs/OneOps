<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useThemeStore } from '@/store/modules/theme';
import { applyContentTheme2 } from '@/utils/content-theme';
import SettingItem from '../components/setting-item.vue';

defineOptions({ name: 'ContentTheme2' });

const themeStore = useThemeStore();

// Hero区域设置 - 直接绑定到 store，不需要本地状态
const heroSection = computed(() => themeStore.contentTheme2.heroSection);

// 统计卡片设置
const statUseGradient = ref(themeStore.contentTheme2.statCards.useGradient || false);
const statDefaultBg = ref(themeStore.contentTheme2.statCards.defaultBg || 'rgba(255, 255, 255, 0.98)');
const statDefaultBgStart = ref(themeStore.contentTheme2.statCards.defaultBgStart || 'rgba(255, 255, 255, 0.98)');
const statDefaultBgEnd = ref(themeStore.contentTheme2.statCards.defaultBgEnd || 'rgba(248, 250, 252, 0.94)');
const statDefaultBorder = ref(themeStore.contentTheme2.statCards.defaultBorder || 'rgba(148, 163, 184, 0.16)');
const statSuccessBg = ref(themeStore.contentTheme2.statCards.successBg || 'rgba(240, 253, 244, 0.98)');
const statSuccessBgStart = ref(themeStore.contentTheme2.statCards.successBgStart || 'rgba(240, 253, 244, 0.98)');
const statSuccessBgEnd = ref(themeStore.contentTheme2.statCards.successBgEnd || 'rgba(255, 255, 255, 0.94)');
const statWarningBg = ref(themeStore.contentTheme2.statCards.warningBg || 'rgba(255, 251, 235, 0.98)');
const statWarningBgStart = ref(themeStore.contentTheme2.statCards.warningBgStart || 'rgba(255, 251, 235, 0.98)');
const statWarningBgEnd = ref(themeStore.contentTheme2.statCards.warningBgEnd || 'rgba(255, 255, 255, 0.94)');
const statDangerBg = ref(themeStore.contentTheme2.statCards.dangerBg || 'rgba(254, 242, 242, 0.98)');
const statDangerBgStart = ref(themeStore.contentTheme2.statCards.dangerBgStart || 'rgba(254, 242, 242, 0.98)');
const statDangerBgEnd = ref(themeStore.contentTheme2.statCards.dangerBgEnd || 'rgba(255, 255, 255, 0.94)');
const statBorderRadius = ref(themeStore.contentTheme2.statCards.borderRadius || '12px');
const statShadow = ref(
  themeStore.contentTheme2.statCards.shadow || '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)'
);

// 工具栏设置
const toolbarGradientStart = ref(themeStore.contentTheme2.toolbar.gradientStart || 'rgba(248, 250, 252, 0.92)');
const toolbarGradientMiddle = ref(themeStore.contentTheme2.toolbar.gradientMiddle || '');
const toolbarGradientEnd = ref(themeStore.contentTheme2.toolbar.gradientEnd || 'rgba(255, 255, 255, 0.96)');
const toolbarGradientAngle = ref(themeStore.contentTheme2.toolbar.gradientAngle || 180);
const toolbarBorderColor = ref(themeStore.contentTheme2.toolbar.borderColor || 'rgba(148, 163, 184, 0.12)');
const toolbarBorderRadius = ref(themeStore.contentTheme2.toolbar.borderRadius || '12px');

// 内容卡片设置
const contentCardBg = ref(themeStore.contentTheme2.contentCard.background || '#ffffff');
const contentCardBgGradientStart = ref(
  themeStore.contentTheme2.contentCard.bgGradientStart || 'rgba(255, 255, 255, 0.98)'
);
const contentCardBgGradientMiddle = ref(themeStore.contentTheme2.contentCard.bgGradientMiddle || '');
const contentCardBgGradientEnd = ref(themeStore.contentTheme2.contentCard.bgGradientEnd || 'rgba(248, 250, 252, 0.94)');
const contentCardUseGradient = ref(themeStore.contentTheme2.contentCard.useGradient || false);
const contentCardGradientAngle = ref(themeStore.contentTheme2.contentCard.gradientAngle || 145);
const contentCardBorderColor = ref(themeStore.contentTheme2.contentCard.borderColor || 'rgba(148, 163, 184, 0.16)');
const contentCardBorderRadius = ref(themeStore.contentTheme2.contentCard.borderRadius || '12px');

// 数据表格设置
const tableHeaderBg = ref(themeStore.contentTheme2.dataTable.headerBg || '#7f748c');
const tableHeaderTextColor = ref(themeStore.contentTheme2.dataTable.headerTextColor || '#ffffff');
const tableRowHoverBg = ref(themeStore.contentTheme2.dataTable.rowHoverBg || '#f8fbff');
const tableStripedBg = ref(themeStore.contentTheme2.dataTable.stripedBg || '#fbfaff');
const tableBorderColor = ref(themeStore.contentTheme2.dataTable.tableBorder || 'rgba(148, 163, 184, 0.14)');
const tableBorderRadius = ref(themeStore.contentTheme2.dataTable.borderRadius || '0px');

// 搜索筛选设置
const searchInputBg = ref(themeStore.contentTheme2.searchFilters.inputBg || 'rgba(255, 255, 255, 0.94)');
const searchInputBorder = ref(themeStore.contentTheme2.searchFilters.inputBorder || 'rgba(148, 163, 184, 0.12)');
const searchInputHoverBorder = ref(
  themeStore.contentTheme2.searchFilters.inputHoverBorder || 'rgba(59, 130, 246, 0.16)'
);
const searchInputFocusBorder = ref(
  themeStore.contentTheme2.searchFilters.inputFocusBorder || 'rgba(37, 99, 235, 0.22)'
);
const searchInputBorderRadius = ref(themeStore.contentTheme2.searchFilters.inputBorderRadius || '8px');

// 分页设置
const paginationButtonBg = ref(themeStore.contentTheme2.pagination.buttonBg || 'rgba(255, 255, 255, 0.9)');
const paginationButtonTextColor = ref(themeStore.contentTheme2.pagination.buttonTextColor || '#475569');
const paginationButtonHoverBg = ref(themeStore.contentTheme2.pagination.buttonHoverBg || '#f8fbff');
const paginationActiveButtonBg = ref(themeStore.contentTheme2.pagination.activeButtonBg || '#3b82f6');
const paginationActiveButtonTextColor = ref(themeStore.contentTheme2.pagination.activeButtonTextColor || '#ffffff');
const paginationBorderRadius = ref(themeStore.contentTheme2.pagination.borderRadius || '8px');

// 标签设置
const tagDefaultBg = ref(themeStore.contentTheme2.tags.defaultBg || 'rgba(255, 255, 255, 0.9)');
const tagSuccessBg = ref(themeStore.contentTheme2.tags.successBg || 'rgba(16, 185, 129, 0.1)');
const tagWarningBg = ref(themeStore.contentTheme2.tags.warningBg || 'rgba(245, 158, 11, 0.1)');
const tagDangerBg = ref(themeStore.contentTheme2.tags.dangerBg || 'rgba(239, 68, 68, 0.1)');
const tagInfoBg = ref(themeStore.contentTheme2.tags.infoBg || 'rgba(59, 130, 246, 0.1)');
const tagBorderRadius = ref(themeStore.contentTheme2.tags.borderRadius || '8px');

// 预设颜色
const colorSwatches: string[] = [
  '#ffffff',
  '#f8fafc',
  '#f1f5f9',
  '#e2e8f0',
  '#cbd5e1',
  '#94a3b8',
  '#64748b',
  '#475569',
  '#334155',
  '#1e293b',
  '#3b82f6',
  '#10b981',
  '#f59e0b',
  '#ef4444'
];

// 更新设置
function updateSetting(module: keyof App.Theme.ThemeSetting['contentTheme2'], key: string, value: unknown) {
  const currentTheme = themeStore.contentTheme2;
  const updatedTheme = {
    ...currentTheme,
    [module]: {
      ...currentTheme[module],
      [key]: value
    }
  };
  themeStore.setContentTheme2(updatedTheme);
  applyContentTheme2(updatedTheme);
}

// Hero区域更新
function updateHeroSetting(key: string, value: unknown) {
  updateSetting('heroSection', key, value);
}

// 统计卡片更新
function updateStatSetting(key: string, value: unknown) {
  updateSetting('statCards', key, value);
}

// 工具栏更新
function updateToolbarSetting(key: string, value: unknown) {
  updateSetting('toolbar', key, value);
}

// 内容卡片更新
function updateContentCardSetting(key: string, value: unknown) {
  updateSetting('contentCard', key, value);
}

// 数据表格更新
function updateTableSetting(key: string, value: unknown) {
  updateSetting('dataTable', key, value);
}

// 搜索筛选更新
function updateSearchSetting(key: string, value: unknown) {
  updateSetting('searchFilters', key, value);
}

// 分页更新
function updatePaginationSetting(key: string, value: unknown) {
  updateSetting('pagination', key, value);
}

// 标签更新
function updateTagSetting(key: string, value: unknown) {
  updateSetting('tags', key, value);
}

// 监听 store 变化 - 简化为只监听必要的变化
watch(
  () => themeStore.contentTheme2,
  newTheme => {
    // 当主题设置变更时应用新主题
    applyContentTheme2(newTheme);
  },
  { deep: true }
);
</script>

<template>
  <ElDivider>内容区设置2 (模块化主题)</ElDivider>
  <div class="flex-col-stretch gap-16px">
    <!-- Hero区域设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">🎯</span>
        <span class="section-text">Hero 区域</span>
      </div>

      <SettingItem label="显示 Hero 区域">
        <ElSwitch :model-value="heroSection.visible !== false" @change="updateHeroSetting('visible', $event)" />
      </SettingItem>

      <template v-if="heroSection.visible !== false">
        <SettingItem label="自定义渐变">
          <ElSwitch :model-value="heroSection.useGradient" @change="updateHeroSetting('useGradient', $event)" />
        </SettingItem>

        <template v-if="heroSection.useGradient">
          <SettingItem label="渐变起始颜色">
            <ElColorPicker
              :model-value="heroSection.gradientStart"
              class="w-40px"
              :show-alpha="true"
              :predefine="colorSwatches"
              @change="(color: string | null) => color && updateHeroSetting('gradientStart', color)"
            />
          </SettingItem>
          <SettingItem label="渐变结束颜色">
            <ElColorPicker
              :model-value="heroSection.gradientEnd"
              class="w-40px"
              :show-alpha="true"
              :predefine="colorSwatches"
              @change="(color: string | null) => color && updateHeroSetting('gradientEnd', color)"
            />
          </SettingItem>
          <SettingItem label="渐变中间颜色 (可选)">
            <ElColorPicker
              :model-value="heroSection.gradientMiddle"
              class="w-40px"
              :show-alpha="true"
              :predefine="colorSwatches"
              @change="(color: string | null) => color && updateHeroSetting('gradientMiddle', color)"
            />
          </SettingItem>
          <SettingItem label="渐变角度">
            <ElInputNumber
              :model-value="heroSection.gradientAngle"
              :min="0"
              :max="360"
              :step="5"
              controls-position="right"
              class="w-120px"
              @change="updateHeroSetting('gradientAngle', $event)"
            />
          </SettingItem>
        </template>

        <SettingItem v-if="!heroSection.useGradient" label="背景颜色">
          <ElColorPicker
            :model-value="heroSection.background"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateHeroSetting('background', color)"
          />
        </SettingItem>

        <SettingItem label="边框颜色">
          <ElColorPicker
            :model-value="heroSection.borderColor"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateHeroSetting('borderColor', color)"
          />
        </SettingItem>

        <SettingItem label="圆角大小">
          <ElInput :model-value="heroSection.borderRadius" class="w-120px" @change="updateHeroSetting('borderRadius', $event)" />
        </SettingItem>

        <div class="sub-section-title">图标样式</div>

        <SettingItem label="图标背景渐变起始">
          <ElColorPicker
            :model-value="heroSection.iconGradientStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateHeroSetting('iconGradientStart', color)"
          />
        </SettingItem>
        <SettingItem label="图标背景渐变结束">
          <ElColorPicker
            :model-value="heroSection.iconGradientEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateHeroSetting('iconGradientEnd', color)"
          />
        </SettingItem>
        <SettingItem label="图标边框颜色">
          <ElColorPicker
            :model-value="heroSection.iconBorderColor"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateHeroSetting('iconBorderColor', color)"
          />
        </SettingItem>
        <SettingItem label="图标文字颜色">
          <ElColorPicker
            :model-value="heroSection.iconColor"
            class="w-40px"
            :show-alpha="false"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateHeroSetting('iconColor', color)"
          />
        </SettingItem>
      </template>
    </div>

    <!-- 统计卡片设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">📊</span>
        <span class="section-text">统计卡片</span>
      </div>

      <SettingItem label="自定义渐变">
        <ElSwitch v-model="statUseGradient" @change="updateStatSetting('useGradient', $event)" />
      </SettingItem>

      <div class="sub-section-title">默认卡片</div>

      <template v-if="statUseGradient">
        <SettingItem label="渐变起始颜色">
          <ElColorPicker
            v-model="statDefaultBgStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('defaultBgStart', color)"
          />
        </SettingItem>
        <SettingItem label="渐变结束颜色">
          <ElColorPicker
            v-model="statDefaultBgEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('defaultBgEnd', color)"
          />
        </SettingItem>
      </template>

      <SettingItem v-if="!statUseGradient" label="背景颜色">
        <ElColorPicker
          v-model="statDefaultBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateStatSetting('defaultBg', color)"
        />
      </SettingItem>

      <SettingItem label="默认边框颜色">
        <ElColorPicker
          v-model="statDefaultBorder"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateStatSetting('defaultBorder', color)"
        />
      </SettingItem>

      <div class="sub-section-title">成功状态卡片</div>

      <template v-if="statUseGradient">
        <SettingItem label="渐变起始颜色">
          <ElColorPicker
            v-model="statSuccessBgStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('successBgStart', color)"
          />
        </SettingItem>
        <SettingItem label="渐变结束颜色">
          <ElColorPicker
            v-model="statSuccessBgEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('successBgEnd', color)"
          />
        </SettingItem>
      </template>

      <SettingItem v-if="!statUseGradient" label="背景颜色">
        <ElColorPicker
          v-model="statSuccessBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateStatSetting('successBg', color)"
        />
      </SettingItem>

      <div class="sub-section-title">警告状态卡片</div>

      <template v-if="statUseGradient">
        <SettingItem label="渐变起始颜色">
          <ElColorPicker
            v-model="statWarningBgStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('warningBgStart', color)"
          />
        </SettingItem>
        <SettingItem label="渐变结束颜色">
          <ElColorPicker
            v-model="statWarningBgEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('warningBgEnd', color)"
          />
        </SettingItem>
      </template>

      <SettingItem v-if="!statUseGradient" label="背景颜色">
        <ElColorPicker
          v-model="statWarningBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateStatSetting('warningBg', color)"
        />
      </SettingItem>

      <div class="sub-section-title">危险状态卡片</div>

      <template v-if="statUseGradient">
        <SettingItem label="渐变起始颜色">
          <ElColorPicker
            v-model="statDangerBgStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('dangerBgStart', color)"
          />
        </SettingItem>
        <SettingItem label="渐变结束颜色">
          <ElColorPicker
            v-model="statDangerBgEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateStatSetting('dangerBgEnd', color)"
          />
        </SettingItem>
      </template>

      <SettingItem v-if="!statUseGradient" label="背景颜色">
        <ElColorPicker
          v-model="statDangerBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateStatSetting('dangerBg', color)"
        />
      </SettingItem>

      <div class="sub-section-title">通用样式</div>

      <SettingItem label="阴影效果">
        <ElInput v-model="statShadow" class="w-200px" @change="updateStatSetting('shadow', $event)" />
      </SettingItem>

      <SettingItem label="圆角大小">
        <ElInput v-model="statBorderRadius" class="w-120px" @change="updateStatSetting('borderRadius', $event)" />
      </SettingItem>
    </div>

    <!-- 工具栏设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">🔧</span>
        <span class="section-text">工具栏</span>
      </div>

      <SettingItem label="背景渐变起始颜色">
        <ElColorPicker
          v-model="toolbarGradientStart"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateToolbarSetting('gradientStart', color)"
        />
      </SettingItem>
      <SettingItem label="背景渐变结束颜色">
        <ElColorPicker
          v-model="toolbarGradientEnd"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateToolbarSetting('gradientEnd', color)"
        />
      </SettingItem>
      <SettingItem label="背景渐变中间颜色 (可选)">
        <ElColorPicker
          v-model="toolbarGradientMiddle"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateToolbarSetting('gradientMiddle', color)"
        />
      </SettingItem>
      <SettingItem label="渐变角度">
        <ElInputNumber
          v-model="toolbarGradientAngle"
          :min="0"
          :max="360"
          :step="5"
          controls-position="right"
          class="w-120px"
          @change="updateToolbarSetting('gradientAngle', $event)"
        />
      </SettingItem>
      <SettingItem label="边框颜色">
        <ElColorPicker
          v-model="toolbarBorderColor"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateToolbarSetting('borderColor', color)"
        />
      </SettingItem>
      <SettingItem label="圆角大小">
        <ElInput v-model="toolbarBorderRadius" class="w-120px" @change="updateToolbarSetting('borderRadius', $event)" />
      </SettingItem>
    </div>

    <!-- 内容卡片设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">📦</span>
        <span class="section-text">内容卡片</span>
      </div>

      <SettingItem label="自定义渐变">
        <ElSwitch v-model="contentCardUseGradient" @change="updateContentCardSetting('useGradient', $event)" />
      </SettingItem>

      <template v-if="!contentCardUseGradient">
        <SettingItem label="背景颜色">
          <ElColorPicker
            v-model="contentCardBg"
            class="w-40px"
            :show-alpha="false"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateContentCardSetting('background', color)"
          />
        </SettingItem>
      </template>

      <template v-if="contentCardUseGradient">
        <SettingItem label="背景渐变起始颜色">
          <ElColorPicker
            v-model="contentCardBgGradientStart"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateContentCardSetting('bgGradientStart', color)"
          />
        </SettingItem>
        <SettingItem label="背景渐变结束颜色">
          <ElColorPicker
            v-model="contentCardBgGradientEnd"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateContentCardSetting('bgGradientEnd', color)"
          />
        </SettingItem>
        <SettingItem label="背景渐变中间颜色 (可选)">
          <ElColorPicker
            v-model="contentCardBgGradientMiddle"
            class="w-40px"
            :show-alpha="true"
            :predefine="colorSwatches"
            @change="(color: string | null) => color && updateContentCardSetting('bgGradientMiddle', color)"
          />
        </SettingItem>
        <SettingItem label="渐变角度">
          <ElInputNumber
            v-model="contentCardGradientAngle"
            :min="0"
            :max="360"
            :step="5"
            controls-position="right"
            class="w-120px"
            @change="updateContentCardSetting('gradientAngle', $event)"
          />
        </SettingItem>
      </template>

      <SettingItem label="边框颜色">
        <ElColorPicker
          v-model="contentCardBorderColor"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateContentCardSetting('borderColor', color)"
        />
      </SettingItem>
      <SettingItem label="圆角大小">
        <ElInput
          v-model="contentCardBorderRadius"
          class="w-120px"
          @change="updateContentCardSetting('borderRadius', $event)"
        />
      </SettingItem>
    </div>

    <!-- 数据表格设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">📋</span>
        <span class="section-text">数据表格</span>
      </div>

      <SettingItem label="表头背景颜色">
        <ElColorPicker
          v-model="tableHeaderBg"
          class="w-40px"
          :show-alpha="false"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTableSetting('headerBg', color)"
        />
      </SettingItem>
      <SettingItem label="表头文字颜色">
        <ElColorPicker
          v-model="tableHeaderTextColor"
          class="w-40px"
          :show-alpha="false"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTableSetting('headerTextColor', color)"
        />
      </SettingItem>
      <SettingItem label="行悬停背景">
        <ElColorPicker
          v-model="tableRowHoverBg"
          class="w-40px"
          :show-alpha="false"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTableSetting('rowHoverBg', color)"
        />
      </SettingItem>
      <SettingItem label="斑马纹背景">
        <ElColorPicker
          v-model="tableStripedBg"
          class="w-40px"
          :show-alpha="false"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTableSetting('stripedBg', color)"
        />
      </SettingItem>
      <SettingItem label="边框颜色">
        <ElColorPicker
          v-model="tableBorderColor"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTableSetting('tableBorder', color)"
        />
      </SettingItem>
      <SettingItem label="圆角大小">
        <ElInput v-model="tableBorderRadius" class="w-120px" @change="updateTableSetting('borderRadius', $event)" />
      </SettingItem>
    </div>

    <!-- 搜索筛选设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">🔍</span>
        <span class="section-text">搜索筛选</span>
      </div>

      <SettingItem label="输入框背景">
        <ElColorPicker
          v-model="searchInputBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateSearchSetting('inputBg', color)"
        />
      </SettingItem>
      <SettingItem label="输入框边框">
        <ElColorPicker
          v-model="searchInputBorder"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateSearchSetting('inputBorder', color)"
        />
      </SettingItem>
      <SettingItem label="输入框悬停边框">
        <ElColorPicker
          v-model="searchInputHoverBorder"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateSearchSetting('inputHoverBorder', color)"
        />
      </SettingItem>
      <SettingItem label="输入框聚焦边框">
        <ElColorPicker
          v-model="searchInputFocusBorder"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateSearchSetting('inputFocusBorder', color)"
        />
      </SettingItem>
      <SettingItem label="输入框圆角">
        <ElInput
          v-model="searchInputBorderRadius"
          class="w-120px"
          @change="updateSearchSetting('inputBorderRadius', $event)"
        />
      </SettingItem>
    </div>

    <!-- 分页设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">📄</span>
        <span class="section-text">分页组件</span>
      </div>

      <SettingItem label="按钮背景">
        <ElColorPicker
          v-model="paginationButtonBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updatePaginationSetting('buttonBg', color)"
        />
      </SettingItem>
      <SettingItem label="按钮文字颜色">
        <ElColorPicker
          v-model="paginationButtonTextColor"
          class="w-40px"
          :show-alpha="false"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updatePaginationSetting('buttonTextColor', color)"
        />
      </SettingItem>
      <SettingItem label="按钮悬停背景">
        <ElColorPicker
          v-model="paginationButtonHoverBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updatePaginationSetting('buttonHoverBg', color)"
        />
      </SettingItem>
      <SettingItem label="激活按钮背景">
        <ElColorPicker
          v-model="paginationActiveButtonBg"
          class="w-40px"
          :show-alpha="false"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updatePaginationSetting('activeButtonBg', color)"
        />
      </SettingItem>
      <SettingItem label="激活按钮文字">
        <ElColorPicker
          v-model="paginationActiveButtonTextColor"
          class="w-40px"
          :show-alpha="false"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updatePaginationSetting('activeButtonTextColor', color)"
        />
      </SettingItem>
      <SettingItem label="圆角大小">
        <ElInput
          v-model="paginationBorderRadius"
          class="w-120px"
          @change="updatePaginationSetting('borderRadius', $event)"
        />
      </SettingItem>
    </div>

    <!-- 标签样式设置 -->
    <div class="theme-section">
      <div class="section-title">
        <span class="section-icon">🏷️</span>
        <span class="section-text">标签样式</span>
      </div>

      <SettingItem label="默认标签背景">
        <ElColorPicker
          v-model="tagDefaultBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTagSetting('defaultBg', color)"
        />
      </SettingItem>
      <SettingItem label="成功标签背景">
        <ElColorPicker
          v-model="tagSuccessBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTagSetting('successBg', color)"
        />
      </SettingItem>
      <SettingItem label="警告标签背景">
        <ElColorPicker
          v-model="tagWarningBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTagSetting('warningBg', color)"
        />
      </SettingItem>
      <SettingItem label="危险标签背景">
        <ElColorPicker
          v-model="tagDangerBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTagSetting('dangerBg', color)"
        />
      </SettingItem>
      <SettingItem label="信息标签背景">
        <ElColorPicker
          v-model="tagInfoBg"
          class="w-40px"
          :show-alpha="true"
          :predefine="colorSwatches"
          @change="(color: string | null) => color && updateTagSetting('infoBg', color)"
        />
      </SettingItem>
      <SettingItem label="圆角大小">
        <ElInput v-model="tagBorderRadius" class="w-120px" @change="updateTagSetting('borderRadius', $event)" />
      </SettingItem>
    </div>
  </div>
</template>

<style scoped lang="scss">
.theme-section {
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);

  &:last-child {
    border-bottom: none;
  }
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}

.section-icon {
  font-size: 18px;
}

.section-text {
  font-size: 14px;
  font-weight: 500;
  color: #475569;
}

.sub-section-title {
  margin: 12px 0 8px;
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  border-left: 2px solid rgba(59, 130, 246, 0.2);
  padding-left: 8px;
}

.w-40px {
  width: 40px;
}

.w-120px {
  width: 120px;
}
</style>
