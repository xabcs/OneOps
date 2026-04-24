<template>
  <el-card shadow="never" class="data-table-card" v-loading="loading">
    <template #header>
      <slot name="header">
        <PageHeader
          :title="title"
          :subtitle="subtitle"
        >
          <template #actions>
            <slot name="actions"></slot>
          </template>
        </PageHeader>
      </slot>
    </template>

    <!-- 搜索和过滤区域 -->
    <div v-if="$slots.filters" class="filters-section">
      <slot name="filters"></slot>
    </div>

    <!-- 表格主体 -->
    <div class="table-section">
      <slot></slot>
    </div>

    <!-- 分页区域 -->
    <div v-if="$slots.pagination" class="pagination-section">
      <slot name="pagination"></slot>
    </div>
  </el-card>
</template>

<script setup>
import PageHeader from './PageHeader.vue'

defineProps({
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
  }
})
</script>

<style scoped>
.data-table-card {
  border-radius: 12px;
  border: 1px solid var(--border, #e2e8f0);
}

.filters-section {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border, #f1f5f9);
}

.table-section {
  margin-bottom: 16px;
}

.pagination-section {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid var(--border, #f1f5f9);
}
</style>
