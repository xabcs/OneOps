<script setup lang="ts">
  interface Props {
    type?: 'primary' | 'success' | 'warning' | 'danger' | 'info';
    size?: number;
  }

  withDefaults(defineProps<Props>(), {
    type: 'primary',
    size: 18
  });

  const colorMap: Record<string, string> = {
    primary: 'var(--el-color-primary)',
    success: 'var(--el-color-success)',
    warning: 'var(--el-color-warning)',
    danger: 'var(--el-color-danger)',
    info: 'var(--el-color-info)'
  };
</script>

<template>
  <span
    :style="{
      display: 'inline-flex',
      width: `${size}px`,
      height: `${size}px`,
      borderRadius: '50%',
      backgroundColor: colorMap[type],
      color: '#fff',
      fontSize: `${Math.round(size * 0.6)}px`,
      alignItems: 'center',
      justifyContent: 'center',
      flexShrink: 0,
      lineHeight: 1
    }"
  >
    <slot />
  </span>
</template>
