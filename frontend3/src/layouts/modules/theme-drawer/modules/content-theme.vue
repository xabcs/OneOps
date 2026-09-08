<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useThemeStore } from '@/store/modules/theme';
  import { gradientPresets, type GradientPresetKey } from '@/theme/settings';
  import SettingItem from '../components/setting-item.vue';

  defineOptions({ name: 'ContentTheme' });

  type ContentSettings = App.Theme.ThemeSetting['content'];
  type TableStyle = ContentSettings['dataTable']['tableStyle'];
  type ModuleKey = Exclude<keyof ContentSettings, 'variant' | 'colors' | 'text' | 'border'>;

  interface FieldDef {
    key: string;
    label: string;
    type: 'switch' | 'color' | 'alphaColor' | 'text' | 'number';
    show?: () => boolean;
  }

  interface ModuleDef {
    key: ModuleKey;
    titleKey: string;
    fields: FieldDef[];
  }

  const themeStore = useThemeStore();
  const content = computed(() => themeStore.content);
  const activeSections = ref<ModuleKey[]>(['hero', 'contentCard', 'toolbar']);
  const gradientPreset = ref<GradientPresetKey | ''>('');
  const gradientPresetOptions = Object.keys(gradientPresets) as GradientPresetKey[];
  const tableStyleOptions: Array<{ value: TableStyle; labelKey: string }> = [
    { value: 'borderless', labelKey: 'theme.content.tableStyles.borderless' },
    { value: 'zebra', labelKey: 'theme.content.tableStyles.zebra' },
    { value: 'soft', labelKey: 'theme.content.tableStyles.soft' },
    { value: 'grid', labelKey: 'theme.content.tableStyles.grid' }
  ];

  const commonFields: Record<string, FieldDef> = {
    background: { key: 'background', label: 'theme.content.fields.background', type: 'color' },
    gradientStart: { key: 'gradientStart', label: 'theme.content.fields.gradientStart', type: 'alphaColor' },
    gradientMiddle: { key: 'gradientMiddle', label: 'theme.content.fields.gradientMiddle', type: 'alphaColor' },
    gradientEnd: { key: 'gradientEnd', label: 'theme.content.fields.gradientEnd', type: 'alphaColor' },
    gradientAngle: { key: 'gradientAngle', label: 'theme.content.fields.gradientAngle', type: 'number' },
    border: { key: 'borderColor', label: 'theme.content.fields.border', type: 'alphaColor' },
    shadow: { key: 'shadow', label: 'theme.content.fields.shadow', type: 'text' },
    padding: { key: 'padding', label: 'theme.content.fields.padding', type: 'text' }
  };

  const gradientFields = (showAngle = true): FieldDef[] => [
    commonFields.gradientStart,
    commonFields.gradientMiddle,
    commonFields.gradientEnd,
    ...(showAngle ? [commonFields.gradientAngle] : [])
  ];

  const modules: ModuleDef[] = [
    {
      key: 'hero',
      titleKey: 'theme.content.sections.hero',
      fields: [{ key: 'visible', label: 'theme.content.fields.visible', type: 'switch' }, ...gradientFields()]
    },
    {
      key: 'statCards',
      titleKey: 'theme.content.sections.statCards',
      fields: [
        { key: 'useGradient', label: 'theme.content.fields.useGradient', type: 'switch' },
        {
          key: 'defaultBg',
          label: 'theme.content.fields.background',
          type: 'color',
          show: () => !content.value.statCards.useGradient
        },
        {
          key: 'defaultBgStart',
          label: 'theme.content.fields.defaultStart',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        {
          key: 'defaultBgEnd',
          label: 'theme.content.fields.defaultEnd',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        {
          key: 'successBg',
          label: 'theme.content.fields.successBg',
          type: 'color',
          show: () => !content.value.statCards.useGradient
        },
        {
          key: 'successBgStart',
          label: 'theme.content.fields.successStart',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        {
          key: 'successBgEnd',
          label: 'theme.content.fields.successEnd',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        {
          key: 'warningBg',
          label: 'theme.content.fields.warningBg',
          type: 'color',
          show: () => !content.value.statCards.useGradient
        },
        {
          key: 'warningBgStart',
          label: 'theme.content.fields.warningStart',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        {
          key: 'warningBgEnd',
          label: 'theme.content.fields.warningEnd',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        {
          key: 'dangerBg',
          label: 'theme.content.fields.dangerBg',
          type: 'color',
          show: () => !content.value.statCards.useGradient
        },
        {
          key: 'dangerBgStart',
          label: 'theme.content.fields.dangerStart',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        {
          key: 'dangerBgEnd',
          label: 'theme.content.fields.dangerEnd',
          type: 'alphaColor',
          show: () => content.value.statCards.useGradient
        },
        commonFields.border,
        commonFields.shadow,
        commonFields.padding
      ]
    },
    {
      key: 'contentCard',
      titleKey: 'theme.content.sections.contentCard',
      fields: [
        { key: 'borderVisible', label: 'theme.content.fields.borderVisible', type: 'switch' },
        commonFields.background,
        ...gradientFields(),
        commonFields.border,
        commonFields.shadow,
        commonFields.padding
      ]
    },
    {
      key: 'toolbar',
      titleKey: 'theme.content.sections.toolbar',
      fields: [
        commonFields.background,
        ...gradientFields(),
        commonFields.border,
        commonFields.padding,
        commonFields.shadow
      ]
    },
    {
      key: 'dataTable',
      titleKey: 'theme.content.sections.dataTable',
      fields: [
        { key: 'headerBg', label: 'theme.content.fields.headerBg', type: 'color' },
        { key: 'headerTextColor', label: 'theme.content.fields.headerText', type: 'color' },
        { key: 'headerBorderColor', label: 'theme.content.fields.headerBorder', type: 'alphaColor' },
        { key: 'rowHoverBg', label: 'theme.content.fields.rowHover', type: 'color' },
        { key: 'rowBorderColor', label: 'theme.content.fields.rowBorder', type: 'alphaColor' },
        { key: 'borderColor', label: 'theme.content.fields.border', type: 'alphaColor' },
        { key: 'stripedBg', label: 'theme.content.fields.stripedBg', type: 'color' }
      ]
    },
    {
      key: 'searchFilters',
      titleKey: 'theme.content.sections.searchFilters',
      fields: [
        { key: 'inputBg', label: 'theme.content.fields.inputBg', type: 'alphaColor' },
        { key: 'inputBorder', label: 'theme.content.fields.inputBorder', type: 'alphaColor' },
        { key: 'inputHoverBorder', label: 'theme.content.fields.inputHoverBorder', type: 'alphaColor' },
        { key: 'inputFocusBorder', label: 'theme.content.fields.inputFocusBorder', type: 'alphaColor' },
        { key: 'buttonBg', label: 'theme.content.fields.buttonBg', type: 'alphaColor' },
        { key: 'buttonTextColor', label: 'theme.content.fields.buttonText', type: 'color' },
        { key: 'buttonHoverBg', label: 'theme.content.fields.buttonHover', type: 'color' }
      ]
    },
    {
      key: 'pagination',
      titleKey: 'theme.content.sections.pagination',
      fields: [
        { key: 'buttonBg', label: 'theme.content.fields.buttonBg', type: 'alphaColor' },
        { key: 'buttonTextColor', label: 'theme.content.fields.buttonText', type: 'color' },
        { key: 'buttonHoverBg', label: 'theme.content.fields.buttonHover', type: 'color' },
        { key: 'activeBg', label: 'theme.content.fields.activeBg', type: 'color' },
        { key: 'activeTextColor', label: 'theme.content.fields.activeText', type: 'color' }
      ]
    },
    {
      key: 'tags',
      titleKey: 'theme.content.sections.tags',
      fields: [
        { key: 'defaultBg', label: 'theme.content.fields.defaultBg', type: 'color' },
        { key: 'defaultBorder', label: 'theme.content.fields.defaultBorder', type: 'alphaColor' },
        { key: 'defaultTextColor', label: 'theme.content.fields.defaultText', type: 'color' },
        { key: 'successBg', label: 'theme.content.fields.successBg', type: 'color' },
        { key: 'warningBg', label: 'theme.content.fields.warningBg', type: 'color' },
        { key: 'dangerBg', label: 'theme.content.fields.dangerBg', type: 'color' },
        { key: 'infoBg', label: 'theme.content.fields.infoBg', type: 'color' }
      ]
    },
    {
      key: 'button',
      titleKey: 'theme.content.sections.button',
      fields: [
        { key: 'defaultBg', label: 'theme.content.fields.buttonBg', type: 'alphaColor' },
        { key: 'defaultColor', label: 'theme.content.fields.buttonText', type: 'color' },
        { key: 'defaultBorder', label: 'theme.content.fields.buttonBorder', type: 'alphaColor' },
        { key: 'hoverBg', label: 'theme.content.fields.buttonHover', type: 'color' },
        { key: 'hoverColor', label: 'theme.content.fields.hoverText', type: 'color' },
        { key: 'hoverBorder', label: 'theme.content.fields.hoverBorder', type: 'alphaColor' }
      ]
    },
    {
      key: 'input',
      titleKey: 'theme.content.sections.input',
      fields: [
        { key: 'bg', label: 'theme.content.fields.inputBg', type: 'alphaColor' },
        { key: 'borderShadow', label: 'theme.content.fields.inputBorder', type: 'alphaColor' },
        { key: 'hoverShadow', label: 'theme.content.fields.inputHoverBorder', type: 'alphaColor' },
        { key: 'focusShadow', label: 'theme.content.fields.inputFocusBorder', type: 'alphaColor' }
      ]
    }
  ];

  const colorFields: FieldDef[] = [
    { key: 'primary', label: 'theme.content.colorGroups.primary', type: 'color' },
    { key: 'primaryLight', label: 'theme.content.colorGroups.primaryLight', type: 'color' },
    { key: 'success', label: 'theme.content.colorGroups.success', type: 'color' },
    { key: 'warning', label: 'theme.content.colorGroups.warning', type: 'color' },
    { key: 'danger', label: 'theme.content.colorGroups.danger', type: 'color' },
    { key: 'info', label: 'theme.content.colorGroups.info', type: 'color' }
  ];

  const textFields: FieldDef[] = [
    { key: 'primary', label: 'theme.content.textGroups.primary', type: 'color' },
    { key: 'secondary', label: 'theme.content.textGroups.secondary', type: 'color' },
    { key: 'muted', label: 'theme.content.textGroups.muted', type: 'color' }
  ];

  const borderFields: FieldDef[] = [
    { key: 'soft', label: 'theme.content.borderGroups.soft', type: 'alphaColor' },
    { key: 'medium', label: 'theme.content.borderGroups.medium', type: 'alphaColor' }
  ];

  const colorSwatches = [
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
    '#8b5cf6',
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ef4444'
  ];

  function getModuleValue(moduleKey: ModuleKey, fieldKey: string) {
    return content.value[moduleKey][fieldKey as keyof ContentSettings[ModuleKey]];
  }

  function setModuleValue(moduleKey: ModuleKey, fieldKey: string, value?: string | number | boolean) {
    if (value === undefined || value === null) return;
    themeStore.setContentModule(moduleKey, { [fieldKey]: value } as never);

    if (moduleKey !== 'statCards' && fieldKey.startsWith('gradient')) {
      gradientPreset.value = '';
    }
  }

  function setColorValue(fieldKey: string, value?: string | number | boolean) {
    if (typeof value === 'string') themeStore.setContentColors({ [fieldKey]: value });
  }

  function setTextValue(fieldKey: string, value?: string | number | boolean) {
    if (typeof value === 'string') themeStore.setContentText({ [fieldKey]: value });
  }

  function setBorderValue(fieldKey: string, value?: string | number | boolean) {
    if (typeof value === 'string') themeStore.setContentBorder({ [fieldKey]: value });
  }

  function setVariant(variant?: string | number | boolean) {
    if (variant === 'standard' || variant === 'modern') {
      themeStore.setContentVariant(variant);
    }
  }

  function applyGradientPreset(key?: GradientPresetKey) {
    if (!key) return;
    const preset = gradientPresets[key];
    themeStore.setContentModule('hero', { ...preset.hero });
    themeStore.setContentModule('contentCard', { ...preset.contentCard });
    themeStore.setContentModule('toolbar', { ...preset.toolbar });
    gradientPreset.value = key;
  }

  function setTableStyle(value?: string | number | boolean) {
    if (value === 'borderless' || value === 'zebra' || value === 'soft' || value === 'grid') {
      themeStore.setContentModule('dataTable', { tableStyle: value });
    }
  }
</script>

<template>
  <ElDivider>{{ $t('theme.content.title') }}</ElDivider>

  <section class="content-theme">
    <SettingItem :label="$t('theme.content.variant.title')">
      <ElRadioGroup :model-value="content.variant" size="small" @change="setVariant">
        <ElRadioButton value="standard">{{ $t('theme.content.variant.standard') }}</ElRadioButton>
        <ElRadioButton value="modern">{{ $t('theme.content.variant.modern') }}</ElRadioButton>
      </ElRadioGroup>
    </SettingItem>

    <SettingItem :label="$t('theme.content.gradientPreset')">
      <ElSelect
        :model-value="gradientPreset"
        size="small"
        clearable
        filterable
        :placeholder="$t('theme.content.customGradient')"
        @change="applyGradientPreset"
      >
        <ElOption
          v-for="key in gradientPresetOptions"
          :key="key"
          :value="key"
          :label="$t(`theme.content.presets.${key}`)"
        />
      </ElSelect>
    </SettingItem>

    <div class="preview" :class="`preview--${content.variant}`" />

    <ElCollapse v-model="activeSections" class="content-collapse">
      <ElCollapseItem v-for="module in modules" :key="module.key" :name="module.key" :title="$t(module.titleKey)">
        <SettingItem v-if="module.key === 'dataTable'" :label="$t('theme.content.fields.tableStyle')">
          <ElSelect :model-value="content.dataTable.tableStyle" size="small" class="w-150px" @change="setTableStyle">
            <ElOption
              v-for="option in tableStyleOptions"
              :key="option.value"
              :value="option.value"
              :label="$t(option.labelKey)"
            />
          </ElSelect>
        </SettingItem>
        <SettingItem
          v-for="field in module.fields.filter(field => !field.show || field.show())"
          :key="field.key"
          :label="$t(field.label)"
        >
          <ElSwitch
            v-if="field.type === 'switch'"
            :model-value="Boolean(getModuleValue(module.key, field.key))"
            @change="value => setModuleValue(module.key, field.key, value as boolean)"
          />
          <ElColorPicker
            v-else-if="field.type === 'color' || field.type === 'alphaColor'"
            class="w-40px"
            :model-value="String(getModuleValue(module.key, field.key) || '')"
            :show-alpha="field.type === 'alphaColor'"
            :predefine="colorSwatches"
            @change="value => setModuleValue(module.key, field.key, value || undefined)"
          />
          <ElInputNumber
            v-else-if="field.type === 'number'"
            class="w-110px"
            :model-value="Number(getModuleValue(module.key, field.key))"
            :min="0"
            :max="360"
            :step="5"
            controls-position="right"
            @change="value => setModuleValue(module.key, field.key, value || 0)"
          />
          <ElInput
            v-else
            class="w-130px"
            :model-value="String(getModuleValue(module.key, field.key) || '')"
            @change="value => setModuleValue(module.key, field.key, value)"
          />
        </SettingItem>
      </ElCollapseItem>

      <ElCollapseItem name="colors" :title="$t('theme.content.sections.colors')">
        <SettingItem v-for="field in colorFields" :key="field.key" :label="$t(field.label)">
          <ElColorPicker
            class="w-40px"
            :model-value="String(content.colors[field.key as keyof typeof content.colors])"
            :predefine="colorSwatches"
            @change="value => setColorValue(field.key, value || undefined)"
          />
        </SettingItem>
      </ElCollapseItem>

      <ElCollapseItem name="text" :title="$t('theme.content.sections.text')">
        <SettingItem v-for="field in textFields" :key="field.key" :label="$t(field.label)">
          <ElColorPicker
            class="w-40px"
            :model-value="String(content.text[field.key as keyof typeof content.text])"
            :predefine="colorSwatches"
            @change="value => setTextValue(field.key, value || undefined)"
          />
        </SettingItem>
      </ElCollapseItem>

      <ElCollapseItem name="border" :title="$t('theme.content.sections.border')">
        <SettingItem v-for="field in borderFields" :key="field.key" :label="$t(field.label)">
          <ElColorPicker
            class="w-40px"
            :model-value="String(content.border[field.key as keyof typeof content.border])"
            :predefine="colorSwatches"
            @change="value => setBorderValue(field.key, value || undefined)"
          />
        </SettingItem>
      </ElCollapseItem>
    </ElCollapse>
  </section>
</template>

<style scoped lang="scss">
  .content-theme {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .preview {
    height: 28px;
    border: 1px solid var(--msre-border-soft);
    border-radius: var(--border-radius-large, 8px);
  }

  .preview--standard {
    background: var(--msre-content-card-bg);
  }

  .preview--modern {
    background: var(--msre-content-card-bg);
  }

  .content-collapse {
    border-top: none;
    border-bottom: none;

    :deep(.el-collapse-item__header) {
      font-weight: 500;
    }

    :deep(.el-collapse-item__content) {
      padding-bottom: 12px;
    }
  }
</style>
