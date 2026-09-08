<script setup lang="ts">
  import { ElTable, ElTableColumn, ElTag } from 'element-plus';

  interface Event {
    type: string;
    reason: string;
    message: string;
    source?: string;
    count?: number;
    lastTimestamp?: string;
  }

  defineProps<{
    events: Event[];
    loading?: boolean;
  }>();
</script>

<template>
  <ElTable
    :data="events"
    :loading="loading"
    stripe
    size="small"
    class="k8s-events-table"
    :header-cell-style="{
      background: '#f5f7fa',
      color: '#303133',
      fontWeight: '600',
      paddingLeft: '16px',
      paddingRight: '16px'
    }"
    :row-style="{ backgroundColor: 'transparent' }"
    :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
  >
    <ElTableColumn prop="type" label="类型" width="120" align="left">
      <template #default="{ row }">
        <ElTag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">
          {{ row.type }}
        </ElTag>
      </template>
    </ElTableColumn>
    <ElTableColumn prop="reason" label="原因" width="150" show-overflow-tooltip align="left" />
    <ElTableColumn prop="message" label="消息" min-width="300" show-overflow-tooltip align="left" />
    <ElTableColumn prop="source" label="来源" width="150" show-overflow-tooltip align="left" />
    <ElTableColumn prop="count" label="次数" width="80" align="center" />
    <ElTableColumn prop="lastTimestamp" label="最后时间" width="160" align="left" />
  </ElTable>
</template>
