<template>
  <el-space class="action-buttons">
    <el-button
      v-for="(action, index) in normalizedActions"
      :key="index"
      :type="action.type || 'default'"
      :size="action.size || 'small'"
      :icon="action.icon"
      :disabled="action.disabled"
      :loading="action.loading"
      link
      @click="handleAction(action)"
    >
      {{ action.label }}
    </el-button>
  </el-space>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  actions: {
    type: Array,
    default: () => []
  },
  row: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['action'])

// 规范化操作配置
const normalizedActions = computed(() => {
  return props.actions.map(action => {
    if (typeof action === 'string') {
      // 简写字符串形式
      return { label: action }
    }
    return action
  })
})

const handleAction = (action) => {
  if (action.handler) {
    // 如果有handler函数，直接调用
    if (props.row) {
      action.handler(props.row, action)
    } else {
      action.handler(action)
    }
  } else {
    // 否则触发action事件
    emit('action', action.action || action.label, props.row)
  }
}
</script>

<style scoped>
.action-buttons {
  display: flex;
  align-items: center;
}

.action-buttons :deep(.el-button) {
  padding: 4px 8px;
}
</style>
