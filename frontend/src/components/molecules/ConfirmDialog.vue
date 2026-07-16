<template>
  <div class="confirm-dialog">
    <ElDialog
      v-model="visible"
      :title="title"
      :width="width"
      :before-close="handleClose"
    >
      <div class="dialog-content">
        <ElIcon :size="40" :color="iconColor">
          <component :is="iconType" />
        </ElIcon>
        <p class="message">{{ message }}</p>
        <p v-if="subMessage" class="sub-message">{{ subMessage }}</p>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <AButton @click="handleCancel">{{ cancelText }}</AButton>
          <AButton type="primary" :loading="loading" @click="handleConfirm">
            {{ confirmText }}
          </AButton>
        </span>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElDialog, ElIcon } from 'element-plus'
import { Warning, Info, CircleCheck } from '@element-plus/icons-vue'
import AButton from '../atoms/AButton.vue'

interface Props {
  modelValue?: boolean
  title?: string
  message?: string
  subMessage?: string
  type?: 'warning' | 'info' | 'success'
  confirmText?: string
  cancelText?: string
  loading?: boolean
  width?: string | number
}

const props = withDefaults(defineProps<Props>(), {
  title: '确认操作',
  message: '您确定要执行此操作吗？',
  type: 'warning',
  confirmText: '确认',
  cancelText: '取消',
  loading: false,
  width: '420px'
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
  cancel: []
}>()

const visible = ref(props.modelValue || false)

// 根据类型选择图标
const iconType = computed(() => {
  switch (props.type) {
    case 'warning':
      return Warning
    case 'info':
      return Info
    case 'success':
      return CircleCheck
    default:
      return Warning
  }
})

// 根据类型选择颜色
const iconColor = computed(() => {
  switch (props.type) {
    case 'warning':
      return '#E6A23C'
    case 'info':
      return '#409EFF'
    case 'success':
      return '#67C23A'
    default:
      return '#E6A23C'
  }
})

// 处理关闭
const handleClose = () => {
  visible.value = false
  emit('update:modelValue', false)
}

// 处理取消
const handleCancel = () => {
  handleClose()
  emit('cancel')
}

// 处理确认
const handleConfirm = () => {
  emit('confirm')
}

// 暴露方法供父组件调用
defineExpose({
  open: () => {
    visible.value = true
    emit('update:modelValue', true)
  },
  close: () => {
    handleClose()
  }
})
</script>

<style scoped>
.dialog-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 0;
  text-align: center;
}

.message {
  margin: 16px 0 8px;
  font-size: 16px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.sub-message {
  margin: 8px 0 0;
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
