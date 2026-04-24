<template>
  <div class="audit-filter">
    <div class="filter-row">
      <el-form :model="form" class="filter-form" inline>
        <slot name="filters"></slot>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
          <el-button :icon="RefreshRight" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div v-if="showTimeRange" class="filter-row time-range-row">
      <TimeRangeSelector
        v-model="timeRange"
        @change="handleTimeRangeChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Search, RefreshRight } from '@element-plus/icons-vue'
import TimeRangeSelector from './TimeRangeSelector.vue'

const props = defineProps({
  form: {
    type: Object,
    required: true
  },
  showTimeRange: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['search', 'reset', 'time-range-change'])

const timeRange = ref('7d')

const handleSearch = () => {
  emit('search')
}

const handleReset = () => {
  timeRange.value = '7d'
  emit('reset')
}

const handleTimeRangeChange = (data) => {
  emit('time-range-change', data)
}

// 暴露重置方法
defineExpose({
  reset: () => {
    handleReset()
  }
})
</script>

<style scoped>
.audit-filter {
  padding: 16px;
  background-color: var(--bg-primary, #ffffff);
  border-bottom: 1px solid var(--border, #e2e8f0);
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.filter-row:last-child {
  margin-bottom: 0;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

:deep(.el-form-item) {
  margin-bottom: 0;
}

.time-range-row {
  margin-top: 8px;
}
</style>
