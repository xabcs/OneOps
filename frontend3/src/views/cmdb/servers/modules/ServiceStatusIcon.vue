<script setup lang="ts">
  interface Props {
    status?: string;
    size?: 'small' | 'default';
    showText?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    status: 'unknown',
    size: 'small',
    showText: false
  });

  function getStatusColor(): string {
    const colorMap: Record<string, string> = {
      running: '#52c41a',
      active: '#52c41a',
      dead: '#ff4d4f',
      failed: '#ff4d4f',
      offline: '#d9d9d9',
      unknown: '#d9d9d9'
    };
    return colorMap[props.status] || '#d9d9d9';
  }

  function getStatusIcon(): string {
    const iconMap: Record<string, string> = {
      running: '●',
      active: '●',
      dead: '●',
      failed: '●',
      offline: '○',
      unknown: '○'
    };
    return iconMap[props.status] || '○';
  }

  function getStatusText(): string {
    const textMap: Record<string, string> = {
      running: '运行中',
      active: '活动',
      dead: '停止',
      failed: '失败',
      offline: '离线',
      unknown: '未知'
    };
    return textMap[props.status] || props.status;
  }
</script>

<template>
  <ElTag
    :size="size"
    effect="light"
    round
    :title="getStatusText()"
    :style="{
      color: getStatusColor(),
      borderColor: getStatusColor(),
      display: 'inline-flex',
      alignItems: 'center',
      gap: '4px'
    }"
  >
    <span>{{ getStatusIcon() }}</span>
    <span v-if="showText">{{ getStatusText() }}</span>
  </ElTag>
</template>
