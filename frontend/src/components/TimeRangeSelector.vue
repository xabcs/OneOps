<template>
  <div class="time-range-selector">
    <span class="time-label">时间范围：</span>
    <el-radio-group v-model="internalRange" size="small" @change="handleRangeChange">
      <el-radio-button label="1h">1小时</el-radio-button>
      <el-radio-button label="1d">今天</el-radio-button>
      <el-radio-button label="7d">7天</el-radio-button>
      <el-radio-button label="30d">30天</el-radio-button>
      <el-radio-button label="custom">自定义</el-radio-button>
    </el-radio-group>
    <el-date-picker
      v-if="internalRange === 'custom'"
      v-model="customTime"
      type="datetimerange"
      range-separator="至"
      start-placeholder="开始时间"
      end-placeholder="结束时间"
      size="small"
      style="width: 350px; margin-left: 8px"
      @change="handleCustomTimeChange"
    />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: '7d'
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const internalRange = ref(props.modelValue)
const customTime = ref([])

watch(internalRange, (newVal) => {
  emit('update:modelValue', newVal)
})

const handleRangeChange = (value) => {
  emit('change', { range: value, customTime: customTime.value })
}

const handleCustomTimeChange = (value) => {
  emit('change', { range: 'custom', customTime: value })
}

// 暴露方法给父组件重置
defineExpose({
  reset: () => {
    internalRange.value = '7d'
    customTime.value = []
  }
})
</script>

<style scoped>
.time-range-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.time-label {
  font-size: 14px;
  color: var(--text-secondary, #606266);
  white-space: nowrap;
}
</style>
