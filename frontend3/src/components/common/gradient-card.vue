<script setup lang="ts">
  interface Props {
    /** 卡片变体 */
    variant?: 'diagonal' | '3d' | 'hover' | 'glass';
    /** 是否启用悬停效果 */
    enableHover?: boolean;
    /** 自定义样式 */
    customStyle?: Record<string, string>;
  }

  defineProps<Props>();
</script>

<template>
  <div
    class="gradient-card"
    :class="[`gradient-card-${variant}`, { 'gradient-card-hover': enableHover }]"
    :style="customStyle"
  >
    <slot />
  </div>
</template>

<style scoped>
  .gradient-card {
    border-radius: 12px;
    transition: all 0.3s ease;
    overflow: hidden;
  }

  .gradient-card-diagonal {
    background: linear-gradient(135deg, #fff, #f7fbff);
    border: 1px solid rgb(148 163 184 / 12%);
    box-shadow: 0 8px 18px rgb(59 130 246 / 14%);
  }

  .gradient-card-3d {
    background: linear-gradient(145deg, #eef5ff, #f6fbff);
    border: 1px solid rgb(148 163 184 / 12%);
    box-shadow:
      inset 0 1px rgb(255 255 255 / 95%),
      0 8px 16px rgb(59 130 246 / 14%);
  }

  .gradient-card-hover:hover {
    background: linear-gradient(180deg, #fff, #f8fbff);
    transform: translateY(-1px);
    box-shadow: 0 12px 24px rgb(59 130 246 / 20%);
  }

  .gradient-card-glass {
    background: linear-gradient(135deg, rgb(255 255 255 / 60%), transparent 52%);
    backdrop-filter: blur(10px);
    backdrop-filter: blur(10px);
    border: 1px solid rgb(148 163 184 / 12%);
  }

  .gradient-card-hover {
    cursor: pointer;
  }

  @media (width <= 768px) {
    .gradient-card {
      border-radius: 8px;
    }
  }
</style>
