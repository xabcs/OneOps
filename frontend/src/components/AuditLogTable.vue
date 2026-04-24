<template>
  <el-card shadow="never" class="audit-table-card" :body-style="{ padding: 0 }">
    <el-table
      :data="data"
      v-loading="loading"
      :border="border"
      style="width: 100%"
      header-cell-class-name="audit-table-header"
      @sort-change="handleSortChange"
    >
      <slot></slot>
    </el-table>

    <div v-if="showPagination" class="pagination-container">
      <el-pagination
        v-model:current-page="internalCurrentPage"
        v-model:page-size="internalPageSize"
        :page-sizes="pageSizes"
        :total="total"
        :layout="paginationLayout"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </el-card>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  data: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  total: {
    type: Number,
    default: 0
  },
  currentPage: {
    type: Number,
    default: 1
  },
  pageSize: {
    type: Number,
    default: 15
  },
  pageSizes: {
    type: Array,
    default: () => [10, 15, 20, 50, 100]
  },
  border: {
    type: Boolean,
    default: false
  },
  showPagination: {
    type: Boolean,
    default: true
  },
  paginationLayout: {
    type: String,
    default: 'total, sizes, prev, pager, next, jumper'
  }
})

const emit = defineEmits(['update:currentPage', 'update:pageSize', 'sort-change'])

const internalCurrentPage = ref(props.currentPage)
const internalPageSize = ref(props.pageSize)

watch(() => props.currentPage, (newVal) => {
  internalCurrentPage.value = newVal
})

watch(() => props.pageSize, (newVal) => {
  internalPageSize.value = newVal
})

const handleSizeChange = (val) => {
  emit('update:pageSize', val)
}

const handleCurrentChange = (val) => {
  emit('update:currentPage', val)
}

const handleSortChange = (...args) => {
  emit('sort-change', ...args)
}
</script>

<style scoped>
.audit-table-card {
  border-radius: 8px;
  border: 1px solid var(--border, #e2e8f0);
  overflow: hidden;
}

:deep(.audit-table-header) {
  background-color: var(--bg-primary, #ffffff) !important;
  font-weight: 500;
  color: var(--text-primary, #1e293b);
  border-bottom: 1px solid var(--border, #e2e8f0);
}

:deep(.audit-table-header .sort-caret.ascending) {
  border-bottom-color: var(--primary, #2563eb);
}

:deep(.audit-table-header .sort-caret.descending) {
  border-top-color: var(--primary, #2563eb);
}

:deep(.el-table) {
  border: none;
}

:deep(.el-table td) {
  border: none;
  padding: 12px 16px;
}

:deep(.el-table tr:hover > td) {
  background-color: var(--bg-tertiary, #f8fafc);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  padding: 16px;
  background-color: var(--bg-primary, #ffffff);
  border-top: 1px solid var(--border, #e2e8f0);
}
</style>
