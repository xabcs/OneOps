<script setup lang="ts">
import { RefreshRight, Search } from '@element-plus/icons-vue';
import AButton from '../atoms/AButton.vue';

interface Props {
  showActions?: boolean;
  showSearch?: boolean;
  showReset?: boolean;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showActions: true,
  showSearch: true,
  showReset: true,
  loading: false
});

const emit = defineEmits<{
  search: [];
  reset: [];
}>();

// 处理搜索
const handleSearch = () => {
  emit('search');
};

// 处理重置
const handleReset = () => {
  emit('reset');
};
</script>

<template>
  <div class="filter-bar">
    <div class="filter-section">
      <slot name="filters" />
    </div>

    <div v-if="showActions" class="actions-section">
      <AButton v-if="showSearch" type="primary" :loading="loading" @click="handleSearch">
        <ElIcon><Search /></ElIcon>
        搜索
      </AButton>

      <AButton v-if="showReset" @click="handleReset">
        <ElIcon><RefreshRight /></ElIcon>
        重置
      </AButton>

      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--el-bg-color-page);
  border-radius: 4px;
}

.filter-section {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.actions-section {
  display: flex;
  gap: 8px;
}
</style>
