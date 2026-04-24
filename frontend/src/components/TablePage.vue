<template>
  <el-card shadow="never" class="table-card" v-loading="loading">
    <template #header>
      <div class="card-header-content">
        <div class="header-left">
          <PageHeader
            :title="title"
            :subtitle="subtitle"
          />
        </div>
        <div class="header-right" v-if="$slots.actions">
          <slot name="actions" />
        </div>
      </div>
    </template>

    <slot name="default" />

    <div class="pagination-container" v-if="showPagination && total > 0">
      <el-pagination
        :current-page="currentPage"
        :page-size="pageSize"
        :page-sizes="pageSizes"
        :layout="paginationLayout"
        :total="total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </el-card>
</template>

<script setup>
import { computed } from 'vue'
import PageHeader from './PageHeader.vue'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: ''
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
    default: 10
  },
  pageSizes: {
    type: Array,
    default: () => [10, 20, 50, 100]
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

const emit = defineEmits(['update:currentPage', 'update:pageSize', 'size-change', 'current-change'])

const handleSizeChange = (val) => {
  emit('update:pageSize', val)
  emit('size-change', val)
}

const handleCurrentChange = (val) => {
  emit('update:currentPage', val)
  emit('current-change', val)
}
</script>

<style scoped>
.table-card {
  border: 1px solid var(--border);
  border-radius: 8px;
}

:deep(.el-card__header) {
  padding: 12px 20px;
  border-bottom: 1px solid var(--border);
  background-color: var(--bg-primary);
}

.card-header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right :deep(.el-space) {
  display: flex;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  padding: 0 20px 20px;
}
</style>
