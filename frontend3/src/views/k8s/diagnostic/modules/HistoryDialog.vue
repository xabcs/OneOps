<script setup lang="ts">
  defineOptions({ name: 'DiagnosticHistoryDialog' });

  const visible = defineModel<boolean>('visible', { default: false });

  defineProps<{
    history: any[];
  }>();

  // 辅助方法
  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleString('zh-CN');
  };
</script>

<template>
  <ElDialog v-model="visible" title="诊断历史" width="800px">
    <ElTimeline>
      <ElTimelineItem
        v-for="item in history"
        :key="item.id"
        :timestamp="formatTime(item.timestamp)"
        placement="top"
      >
        <ElCard>
          <div class="history-item">
            <div class="history-header">
              <strong>{{ item.command }}</strong>
              <ElTag :type="item.status === 'success' ? 'success' : 'danger'" size="small">
                {{ item.status }}
              </ElTag>
            </div>
            <div v-if="item.args && item.args.length" class="history-args">参数: {{ item.args.join(', ') }}</div>
            <div class="history-user">执行人: {{ item.user }}</div>
          </div>
        </ElCard>
      </ElTimelineItem>
    </ElTimeline>
  </ElDialog>
</template>

<style scoped lang="scss">
  .history-item {
    .history-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
    }

    .history-args {
      color: #606266;
      font-size: 14px;
      margin-bottom: 4px;
    }

    .history-user {
      color: #909399;
      font-size: 12px;
    }
  }
</style>
