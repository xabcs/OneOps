<script setup lang="ts">
interface Props {
  type?: 'success' | 'warning' | 'danger' | 'info' | 'primary';
  text?: string;
  size?: 'small' | 'medium' | 'large';
  showDot?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'info',
  text: '',
  size: 'medium',
  showDot: false
});
</script>

<template>
  <div class="status-badge" :class="[`status-${type}`, size]">
    <span v-if="showDot" class="status-dot" />
    <slot>{{ text }}</slot>
  </div>
</template>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
}

.status-badge.small {
  padding: 2px 8px;
  font-size: 12px;
}

.status-badge.large {
  padding: 6px 16px;
  font-size: 16px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* 状态颜色 */
.status-success {
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.status-success .status-dot {
  background: var(--el-color-success);
}

.status-warning {
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
}

.status-warning .status-dot {
  background: var(--el-color-warning);
}

.status-danger {
  background: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.status-danger .status-dot {
  background: var(--el-color-danger);
}

.status-info {
  background: var(--el-color-info-light-9);
  color: var(--el-color-info);
}

.status-info .status-dot {
  background: var(--el-color-info);
}

.status-primary {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.status-primary .status-dot {
  background: var(--el-color-primary);
}
</style>
