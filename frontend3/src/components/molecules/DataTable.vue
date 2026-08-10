<script setup lang="ts">
import { computed } from 'vue';
import { Document } from '@element-plus/icons-vue';
import type { TableColumn } from '@element-plus/components/table';

interface Props {
  data: any[];
  loading?: boolean;
  stripe?: boolean;
  border?: boolean;
  size?: 'large' | 'default' | 'small';
  rowKey?: string | ((row: any) => string);
  height?: string | number;
  maxHeight?: string | number;
  columns?: TableColumn[];
  showSelection?: boolean;
  showIndex?: boolean;
  showActions?: boolean;
  showPagination?: boolean;
  actionWidth?: number;
  actionsFixed?: 'left' | 'right';
  emptyText?: string;
  pagination?: {
    currentPage: number;
    pageSize: number;
    total: number;
    pageSizes?: number[];
  };
  paginationLayout?: string;
  selectable?: (row: any, index: number) => boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  stripe: true,
  border: true,
  size: 'default',
  rowKey: 'id',
  columns: () => [],
  showSelection: false,
  showIndex: false,
  showActions: true,
  showPagination: true,
  actionWidth: 200,
  actionsFixed: 'right',
  emptyText: '暂无数据',
  pagination: () => ({
    currentPage: 1,
    pageSize: 20,
    total: 0
  }),
  paginationLayout: 'total, sizes, prev, pager, next'
});

const emit = defineEmits<{
  'selection-change': [selection: any[]];
  'sort-change': SortConfig;
  'size-change': [pageSize: number];
  'current-change': [currentPage: number];
}>();

interface SortConfig {
  prop: string;
  order: 'ascending' | 'descending';
}

// 获取显示值
const getDisplayValue = (row: any, prop: string) => {
  const value = row[prop];
  if (value === null || value === undefined) {
    return '-';
  }
  return value;
};

// 处理选择变化
const handleSelectionChange = (selection: any[]) => {
  emit('selection-change', selection);
};

// 处理排序变化
const handleSortChange = (sort: SortConfig) => {
  emit('sort-change', sort);
};

// 处理分页大小变化
const handleSizeChange = (pageSize: number) => {
  emit('size-change', pageSize);
};

// 处理当前页变化
const handleCurrentChange = (currentPage: number) => {
  emit('current-change', currentPage);
};
</script>

<template>
  <div class="data-table">
    <ElTable
      v-loading="loading"
      :data="data"
      :stripe="stripe"
      :border="border"
      :size="size"
      :row-key="rowKey"
      :height="height"
      :max-height="maxHeight"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <!-- 选择列 -->
      <ElTableColumn v-if="showSelection" type="selection" width="50" align="center" :selectable="selectable" />

      <!-- 序号列 -->
      <ElTableColumn v-if="showIndex" type="index" label="序号" width="60" align="center" />

      <!-- 动态列 -->
      <template v-for="column in columns" :key="column.prop">
        <ElTableColumn
          :prop="column.prop"
          :label="column.label"
          :width="column.width"
          :min-width="column.minWidth"
          :fixed="column.fixed"
          :align="column.align || 'left'"
          :sortable="column.sortable"
        >
          <template #default="{ row }">
            <slot :name="column.slot || column.prop" :row="row" :column="column">
              {{ getDisplayValue(row, column.prop) }}
            </slot>
          </template>
        </ElTableColumn>
      </template>

      <!-- 操作列 -->
      <ElTableColumn v-if="showActions" label="操作" :width="actionWidth" :fixed="actionsFixed" align="center">
        <template #default="{ row }">
          <slot name="actions" :row="row" />
        </template>
      </ElTableColumn>

      <!-- 空状态 -->
      <template #empty>
        <div class="empty-state">
          <ElIcon :size="40">
            <Document />
          </ElIcon>
          <p>{{ emptyText }}</p>
        </div>
      </template>
    </ElTable>

    <!-- 分页 -->
    <div v-if="showPagination && pagination" class="table-pagination">
      <ElPagination
        v-model:current-page="pagination.currentPage"
        v-model:page-size="pagination.pageSize"
        :page-sizes="pagination.pageSizes || [10, 20, 50, 100]"
        :total="pagination.total"
        :layout="paginationLayout || 'total, sizes, prev, pager, next'"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<style scoped>
.data-table {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.table-pagination {
  display: flex;
  justify-content: flex-end;
  padding: 16px;
  border-top: 1px solid var(--el-border-color);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--el-text-color-placeholder);
}

.empty-state p {
  margin-top: 12px;
  font-size: 14px;
}
</style>
