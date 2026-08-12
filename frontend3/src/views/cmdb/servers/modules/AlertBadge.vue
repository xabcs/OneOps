<script setup lang="ts">
  interface Props {
    count?: number;
    maxCount?: number;
    size?: 'small' | 'default';
  }

  const props = withDefaults(defineProps<Props>(), {
    count: 0,
    maxCount: 99,
    size: 'small'
  });

  function getBadgeType(): 'danger' | 'warning' | 'info' {
    if (props.count >= 10) return 'danger';
    if (props.count >= 5) return 'warning';
    return 'info';
  }

  function formatCount(): string {
    if (props.count > props.maxCount) {
      return `${props.maxCount}+`;
    }
    return props.count.toString();
  }
</script>

<template>
  <ElBadge v-if="count > 0" :type="getBadgeType()" :value="formatCount()" :size="size" class="alert-badge">
    <slot />
  </ElBadge>
  <div v-else class="alert-badge-placeholder">
    <slot />
  </div>
</template>

<style scoped>
  .alert-badge-placeholder {
    display: inline-block;
  }
</style>
