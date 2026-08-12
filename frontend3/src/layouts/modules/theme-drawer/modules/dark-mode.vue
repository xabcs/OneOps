<script setup lang="ts">
  import { themeSchemaRecord } from '@/constants/app';
  import { useThemeStore } from '@/store/modules/theme';
  import { $t } from '@/locales';
  import SettingItem from '../components/setting-item.vue';

  defineOptions({ name: 'DarkMode' });

  const themeStore = useThemeStore();

  const icons: Record<UnionKey.ThemeScheme, string> = {
    light: 'material-symbols:sunny',
    dark: 'material-symbols:nightlight-rounded',
    auto: 'material-symbols:hdr-auto'
  };

  function handleSegmentChange(value: string | number) {
    themeStore.setThemeScheme(value as UnionKey.ThemeScheme);
  }

  function handleGrayscaleChange(value: boolean) {
    themeStore.setGrayscale(value);
  }

  function handleColourWeaknessChange(value: boolean) {
    themeStore.setColourWeakness(value);
  }
</script>

<template>
  <ElDivider>{{ $t('theme.themeSchema.title') }}</ElDivider>
  <div class="flex-col-stretch gap-16px">
    <div class="i-flex-center">
      <ElTabs v-model="themeStore.themeScheme" type="border-card" class="segment" @tab-change="handleSegmentChange">
        <ElTabPane v-for="(_, key) in themeSchemaRecord" :key="key" :name="key">
          <template #label>
            <SvgIcon :icon="icons[key]" class="h-23px text-icon-small" />
          </template>
        </ElTabPane>
      </ElTabs>
    </div>
    <SettingItem :label="$t('theme.grayscale')">
      <ElSwitch v-model:model-value="themeStore.grayscale" :update:model-value="handleGrayscaleChange" />
    </SettingItem>
    <SettingItem :label="$t('theme.colourWeakness')">
      <ElSwitch v-model:model-value="themeStore.colourWeakness" :update:model-value="handleColourWeaknessChange" />
    </SettingItem>
  </div>
</template>

<style scoped></style>
