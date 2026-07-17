<script setup lang="ts">
import { ref, watch } from 'vue';
import { RefreshRight, Search } from '@element-plus/icons-vue';
import AInput from '../atoms/AInput.vue';
import AButton from '../atoms/AButton.vue';

interface Props {
  modelValue?: string;
  placeholder?: string;
  clearable?: boolean;
  showFilters?: boolean;
  showSearch?: boolean;
  showReset?: boolean;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '请输入搜索内容',
  clearable: true,
  showFilters: false,
  showSearch: true,
  showReset: true,
  loading: false
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
  search: [keyword: string];
  reset: [];
}>();

const searchKeyword = ref(props.modelValue || '');

// 监听外部值变化
watch(
  () => props.modelValue,
  newValue => {
    searchKeyword.value = newValue || '';
  }
);

// 监听内部值变化
watch(searchKeyword, newValue => {
  emit('update:modelValue', newValue);
});

// 处理搜索输入
const handleSearch = () => {
  emit('search', searchKeyword.value);
};

// 处理搜索按钮点击
const handleSearchClick = () => {
  emit('search', searchKeyword.value);
};

// 处理重置
const handleReset = () => {
  searchKeyword.value = '';
  emit('update:modelValue', '');
  emit('reset');
};
</script>

<template>
  <div class="search-bar">
    <div class="search-input-wrapper">
      <AInput v-model="searchKeyword" :placeholder="placeholder" :clearable="clearable" @input="handleSearch">
        <template #prepend>
          <ElIcon>
            <Search />
          </ElIcon>
        </template>
      </AInput>
    </div>

    <div v-if="showFilters" class="filters-wrapper">
      <slot name="filters" />
    </div>

    <div class="actions-wrapper">
      <AButton v-if="showSearch" type="primary" :loading="loading" @click="handleSearchClick">
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
.search-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--el-bg-color-page);
  border-radius: 4px;
}

.search-input-wrapper {
  flex: 1;
  max-width: 300px;
}

.filters-wrapper {
  flex: 1;
}

.actions-wrapper {
  display: flex;
  gap: 8px;
}
</style>
