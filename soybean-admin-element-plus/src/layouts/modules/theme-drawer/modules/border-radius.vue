<script setup lang="ts">
import { useThemeStore } from '@/store/modules/theme';
import { $t } from '@/locales';
import SettingItem from '../components/setting-item.vue';

defineOptions({ name: 'BorderRadius' });

const themeStore = useThemeStore();

// 圆角预设选项
const borderRadiusOptions = [
  { label: '0px (方角)', value: '0px' },
  { label: '2px (小圆角)', value: '2px' },
  { label: '4px (适中)', value: '4px' },
  { label: '6px (默认)', value: '6px' },
  { label: '8px (圆润)', value: '8px' },
  { label: '12px (很圆润)', value: '12px' },
  { label: '16px (超圆)', value: '16px' },
  { label: '24px (极圆)', value: '24px' }
];

function handleBorderRadiusChange(key: keyof App.Theme.ThemeSetting['borderRadius'], value: string) {
  themeStore.setBorderRadius(key, value);
}

function handleComponentRadiusChange(key: keyof App.Theme.ThemeSetting['borderRadius']['components'], value: string) {
  themeStore.setComponentBorderRadius(key, value);
}

// 获取选中的预设值
function getSelectedValue(value: string) {
  const option = borderRadiusOptions.find(opt => opt.value === value);
  return option?.value || '';
}

// 组件圆角配置项
const componentConfigs = [
  { key: 'button' as const, label: '按钮', desc: '所有类型的按钮' },
  { key: 'input' as const, label: '输入框', desc: '文本输入、数字输入、文本域' },
  { key: 'select' as const, label: '选择框', desc: '下拉选择、级联选择' },
  { key: 'card' as const, label: '卡片', desc: '卡片容器、面板' },
  { key: 'table' as const, label: '表格', desc: '数据表格容器' },
  { key: 'modal' as const, label: '对话框', desc: '弹窗、抽屉' },
  { key: 'tag' as const, label: '标签', desc: '标签、徽章' },
  { key: 'menu' as const, label: '菜单', desc: '菜单项、导航' },
  { key: 'switch' as const, label: '开关', desc: '开关组件' },
  { key: 'checkbox' as const, label: '复选框', desc: '复选框组件' },
  { key: 'radio' as const, label: '单选框', desc: '单选框组件' }
] as const;

// 基础分类圆角的选中值
function getBasicSelectedValue(type: 'small' | 'medium' | 'large') {
  return getSelectedValue(themeStore.borderRadius[type]);
}
</script>

<template>
  <ElDivider>{{ $t('theme.borderRadius.title') }}</ElDivider>
  <div class="flex-col-stretch gap-12px">
    <!-- 组件级圆角开关 -->
    <SettingItem label="启用组件级圆角">
      <ElSwitch v-model="themeStore.borderRadius.useComponentSpecific" />
      <template #tip>
        <div class="text-xs text-regular mb-1">
          <template v-if="!themeStore.borderRadius.useComponentSpecific">
            <div class="text-gray-500">统一模式：按组件大小分类统一设置</div>
          </template>
          <template v-else>
            <div class="text-gray-500">组件级模式：对每个组件单独精细控制</div>
          </template>
        </div>
      </template>
    </SettingItem>

    <!-- 基础分类圆角模式 -->
    <template v-if="!themeStore.borderRadius.useComponentSpecific">
      <!-- 小组件 -->
      <SettingItem :label="$t('theme.borderRadius.small')">
        <template #label-extra>
          <span class="text-xs text-gray-500 ml-2">{{ $t('theme.borderRadius.smallDesc') }}</span>
        </template>
        <div class="flex-y-center gap-8px">
          <ElSelect
            :model-value="getBasicSelectedValue('small')"
            size="small"
            class="w-120px"
            @change="(val: string) => handleBorderRadiusChange('small', val)"
          >
            <ElOption v-for="opt in borderRadiusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </ElSelect>
          <ElInputNumber
            :model-value="parseInt(themeStore.borderRadius.small)"
            size="small"
            :min="0"
            :max="100"
            class="w-100px"
            controls-position="right"
            @change="(val: number | undefined) => val && handleBorderRadiusChange('small', `${val}px`)"
          />
        </div>
        <template #tip>
          <div class="text-xs text-gray-500">生效范围：分页按钮、小号交互元素</div>
        </template>
      </SettingItem>

      <!-- 中组件 -->
      <SettingItem :label="$t('theme.borderRadius.medium')">
        <template #label-extra>
          <span class="text-xs text-gray-500 ml-2">{{ $t('theme.borderRadius.mediumDesc') }}</span>
        </template>
        <div class="flex-y-center gap-8px">
          <ElSelect
            :model-value="getBasicSelectedValue('medium')"
            size="small"
            class="w-120px"
            @change="(val: string) => handleBorderRadiusChange('medium', val)"
          >
            <ElOption v-for="opt in borderRadiusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </ElSelect>
          <ElInputNumber
            :model-value="parseInt(themeStore.borderRadius.medium)"
            size="small"
            :min="0"
            :max="100"
            class="w-100px"
            controls-position="right"
            @change="(val: number | undefined) => val && handleBorderRadiusChange('medium', `${val}px`)"
          />
        </div>
        <template #tip>
          <div class="text-xs text-gray-500">生效范围：按钮、输入框、选择框、菜单项、标签、通知消息等标准组件</div>
        </template>
      </SettingItem>

      <!-- 大容器 -->
      <SettingItem :label="$t('theme.borderRadius.large')">
        <template #label-extra>
          <span class="text-xs text-gray-500 ml-2">{{ $t('theme.borderRadius.largeDesc') }}</span>
        </template>
        <div class="flex-y-center gap-8px">
          <ElSelect
            :model-value="getBasicSelectedValue('large')"
            size="small"
            class="w-120px"
            @change="(val: string) => handleBorderRadiusChange('large', val)"
          >
            <ElOption v-for="opt in borderRadiusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </ElSelect>
          <ElInputNumber
            :model-value="parseInt(themeStore.borderRadius.large)"
            size="small"
            :min="0"
            :max="100"
            class="w-100px"
            controls-position="right"
            @change="(val: number | undefined) => val && handleBorderRadiusChange('large', `${val}px`)"
          />
        </div>
        <template #tip>
          <div class="text-xs text-gray-500">生效范围：卡片、对话框、面板、表格等大容器组件</div>
        </template>
      </SettingItem>
    </template>

    <!-- 组件级圆角模式 -->
    <template v-else>
      <SettingItem v-for="config in componentConfigs" :key="config.key" :label="config.label">
        <template #label-extra>
          <span class="text-xs text-gray-500 ml-2">{{ config.desc }}</span>
        </template>
        <div class="flex-y-center gap-8px">
          <ElSelect
            :model-value="getSelectedValue(themeStore.borderRadius.components[config.key])"
            size="small"
            class="w-120px"
            @change="(val: string) => handleComponentRadiusChange(config.key, val)"
          >
            <ElOption v-for="opt in borderRadiusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
          </ElSelect>
          <ElInputNumber
            :model-value="parseInt(themeStore.borderRadius.components[config.key])"
            size="small"
            :min="0"
            :max="100"
            class="w-100px"
            controls-position="right"
            @change="(val: number | undefined) => val && handleComponentRadiusChange(config.key, `${val}px`)"
          />
        </div>
      </SettingItem>
    </template>
  </div>
</template>

<style scoped>
.tip-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
