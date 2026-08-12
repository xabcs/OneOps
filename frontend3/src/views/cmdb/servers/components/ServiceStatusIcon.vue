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
  <div
    class="service-status-icon"
    :class="[`status-${status}`, size, { 'with-text': showText }]"
    :title="getStatusText()"
  >
    <span class="status-indicator" :style="{ color: getStatusColor() }">{{ getStatusIcon() }}</span>
    <span v-if="showText" class="status-text">{{ getStatusText() }}</span>
  </div>
</template>

<style scoped>
  .service-status-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }

  .status-indicator {
    font-size: 12px;
  }

  .status-text {
    font-size: 12px;
  }

  .service-status-icon.small .status-indicator {
    font-size: 10px;
  }

  .service-status-icon.small .status-text {
    font-size: 10px;
  }

  .service-status-icon.default .status-indicator {
    font-size: 14px;
  }

  .service-status-icon.default .status-text {
    font-size: 12px;
  }

  .status-running,
  .status-active {
    color: #52c41a;
  }

  .status-dead,
  .status-failed {
    color: #ff4d4f;
  }

  .status-offline,
  .status-unknown {
    color: #d9d9d9;
  }
</style>
