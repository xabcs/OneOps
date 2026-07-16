<template>
  <div
    :class="['gradient-card', `gradient-card-${variant}`, { 'gradient-card-hover': enableHover }]"
    :style="customStyle"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  /** 卡片变体 */
  variant?: 'diagonal' | '3d' | 'hover' | 'glass';
  /** 是否启用悬停效果 */
  enableHover?: boolean;
  /** 自定义样式 */
  customStyle?: Record<string, string>;
}

defineProps<Props>();

const variantClasses = computed(() => ({
  'gradient-card-diagonal': 'gradient-card-diagonal',
  'gradient-card-3d': 'gradient-card-3d',
  'gradient-card-hover': 'gradient-card-hover',
  'gradient-card-glass': 'gradient-card-glass',
}));
</script>

<style scoped>
.gradient-card {
  border-radius: 12px;
  transition: all 0.3s ease;
  overflow: hidden;
}

.gradient-card-diagonal {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
}

.gradient-card-3d {
  background: linear-gradient(145deg, #eef5ff, #f6fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.95), 0 8px 16px rgba(59, 130, 246, 0.14);
}

.gradient-card-hover:hover {
  background: linear-gradient(180deg, #ffffff, #f8fbff);
  transform: translateY(-1px);
  box-shadow: 0 12px 24px rgba(59, 130, 246, 0.20);
}

.gradient-card-glass {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.6), transparent 52%);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1px solid rgba(148, 163, 184, 0.12);
}

.gradient-card-hover {
  cursor: pointer;
}

@media (max-width: 768px) {
  .gradient-card {
    border-radius: 8px;
  }
}
</style>