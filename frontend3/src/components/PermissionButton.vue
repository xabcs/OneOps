<template>
  <div class="smart-permission-button">
    <!-- 有权限：显示按钮 -->
    <el-button
      v-if="hasPermission"
      v-bind="$attrs"
      @click="handleClick">
      <slot name="icon">
        <el-icon v-if="icon"><component :is="icon" /></el-icon>
      </slot>
      <slot></slot>
    </el-button>

    <!-- 无权限：根据模式显示不同内容 -->
    <template v-else>
      <!-- 模式1：完全隐藏 -->
      <div v-if="mode === 'hidden'" class="hidden-content"></div>

      <!-- 模式2：禁用 + 提示 -->
      <el-tooltip
        v-else-if="mode === 'disabled'"
        :content="tooltip || '您没有此操作权限'"
        placement="top">
        <el-button
          v-bind="$attrs"
          disabled
          class="permission-disabled">
          <slot name="icon">
            <el-icon v-if="icon"><component :is="icon" /></el-icon>
          </slot>
          <slot></slot>
        </el-button>
      </el-tooltip>

      <!-- 模式3：显示申请按钮 -->
      <div v-else-if="mode === 'request'" class="permission-request">
        <el-text type="info" size="small">
          <el-icon><Lock /></el-icon>
          需要权限
        </el-text>
        <el-button
          type="text"
          size="small"
          @click="handleRequestPermission">
          申请
        </el-button>
      </div>

      <!-- 模式4：显示提示信息 -->
      <div v-else-if="mode === 'placeholder'" class="permission-placeholder">
        <el-text type="info" size="small">
          <el-icon><Lock /></el-icon>
          {{ placeholder || '暂无权限' }}
        </el-text>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/store/modules/auth'

interface Props {
  permission: string | string[]
  mode?: 'hidden' | 'disabled' | 'request' | 'placeholder'
  icon?: string
  tooltip?: string
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'hidden'
})

const emit = defineEmits<{
  click: [event: MouseEvent]
  requestPermission: []
}>()

const authStore = useAuthStore()

const hasPermission = computed(() => {
  if (typeof props.permission === 'string') {
    return authStore.hasPermission(props.permission)
  }
  return authStore.hasAnyPermission(props.permission)
})

const handleClick = (event: MouseEvent) => {
  if (hasPermission.value) {
    emit('click', event)
  }
}

const handleRequestPermission = () => {
  emit('requestPermission')
}
</script>

<style scoped>
.permission-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.permission-request {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.permission-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 32px;
  color: #909399;
}

.hidden-content {
  display: none;
}
</style>