<script setup lang="ts">
import { ref } from 'vue';
import PageHeader from '../organisms/PageHeader.vue';
import SearchBar from '../molecules/SearchBar.vue';
import DataTable from '../molecules/DataTable.vue';

interface Column {
  prop: string;
  label: string;
  slot?: string;
  width?: number;
  minWidth?: number;
  align?: 'left' | 'center' | 'right';
  sortable?: boolean;
}

type TableRow = Record<string, unknown>;

interface Props {
  title?: string;
  searchPlaceholder?: string;
  columns: Column[];
  tableData: TableRow[];
  loading?: boolean;
  showSelection?: boolean;
  showIndex?: boolean;
  showPagination?: boolean;
  emptyText?: string;
  pagination?: {
    currentPage: number;
    pageSize: number;
    total: number;
    pageSizes?: number[];
  };
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  searchPlaceholder: '请输入搜索内容',
  loading: false,
  showSelection: false,
  showIndex: true,
  showPagination: true,
  emptyText: '暂无数据',
  pagination: () => ({
    currentPage: 1,
    pageSize: 20,
    total: 0
  })
});

const emit = defineEmits<{
  search: [keyword: string];
  reset: [];
  'selection-change': [selection: TableRow[]];
  'page-change': [page: number];
  'size-change': [pageSize: number];
}>();

const searchKeyword = ref('');

// 处理搜索
const handleSearch = (keyword: string) => {
  emit('search', keyword);
};

// 处理重置
const handleReset = () => {
  searchKeyword.value = '';
  emit('reset');
};

// 处理选择变化
const handleSelectionChange = (selection: TableRow[]) => {
  emit('selection-change', selection);
};

// 处理页面变化
const handlePageChange = (page: number) => {
  emit('page-change', page);
};

// 处理每页数量变化
const handleSizeChange = (pageSize: number) => {
  emit('size-change', pageSize);
};
</script>

<template>
  <div class="list-view">
    <!-- 页面头部 -->
    <PageHeader>
      <template #title>
        <slot name="title">
          <h2>{{ title }}</h2>
        </slot>
      </template>

      <template #actions>
        <slot name="header-actions" />
      </template>
    </PageHeader>

    <!-- 搜索和筛选 -->
    <div class="search-section">
      <SearchBar
        v-model="searchKeyword"
        :placeholder="searchPlaceholder"
        :loading="loading"
        @search="handleSearch"
        @reset="handleReset"
      >
        <template #filters>
          <slot name="filters" />
        </template>

        <template #actions>
          <slot name="search-actions" />
        </template>
      </SearchBar>
    </div>

    <!-- 数据表格 -->
    <div class="table-section">
      <DataTable
        :data="tableData"
        :columns="columns"
        :loading="loading"
        :show-selection="showSelection"
        :show-index="showIndex"
        :show-pagination="showPagination"
        :pagination="pagination"
        :empty-text="emptyText"
        @selection-change="handleSelectionChange"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      >
        <template v-for="column in columns" :key="column.prop" #[column.slot]="slotProps">
          <slot :name="column.slot" v-bind="slotProps" />
        </template>

        <template #actions="slotProps">
          <slot name="row-actions" v-bind="slotProps" />
        </template>
      </DataTable>
    </div>
  </div>
</template>

<style scoped>
.list-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
}

.search-section {
  flex-shrink: 0;
}

.table-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
</style>
