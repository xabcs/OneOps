<template>
  <el-tag
    :type="tagType"
    :size="size"
    :effect="effect"
    class="status-tag"
  >
    {{ displayText }}
  </el-tag>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: [String, Boolean, Number],
    required: true
  },
  size: {
    type: String,
    default: 'small',
    validator: (value) => ['large', 'default', 'small'].includes(value)
  },
  effect: {
    type: String,
    default: 'light',
    validator: (value) => ['dark', 'light', 'plain'].includes(value)
  },
  // 自定义状态映射
  statusMap: {
    type: Object,
    default: () => ({})
  }
})

// 状态类型和文本的映射
const defaultStatusMap = {
  // 激活状态
  active: { type: 'success', text: '启用' },
  enabled: { type: 'success', text: '启用' },
  true: { type: 'success', text: '是' },
  1: { type: 'success', text: '启用' },
  '1': { type: 'success', text: '启用' },

  // 禁用状态
  disabled: { type: 'info', text: '禁用' },
  inactive: { type: 'info', text: '禁用' },
  false: { type: 'info', text: '否' },
  0: { type: 'info', text: '禁用' },
  '0': { type: 'info', text: '禁用' },

  // 警告状态
  warning: { type: 'warning', text: '警告' },
  pending: { type: 'warning', text: '待处理' },

  // 错误状态
  error: { type: 'danger', text: '错误' },
  failed: { type: 'danger', text: '失败' },

  // 信息状态
  processing: { type: 'primary', text: '处理中' },
  draft: { type: 'info', text: '草稿' }
}

const tagConfig = computed(() => {
  const statusKey = String(props.status)

  // 优先使用自定义映射
  if (props.statusMap[statusKey]) {
    return props.statusMap[statusKey]
  }

  // 使用默认映射
  if (defaultStatusMap[statusKey]) {
    return defaultStatusMap[statusKey]
  }

  // 未知状态
  return { type: 'info', text: statusKey }
})

const tagType = computed(() => tagConfig.value.type)

const displayText = computed(() => tagConfig.value.text)
</script>

<style scoped>
.status-tag {
  font-weight: 500;
}
</style>
