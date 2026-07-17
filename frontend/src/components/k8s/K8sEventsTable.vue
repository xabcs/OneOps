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

<style scoped>
/* 表格透明样式 - 完全覆盖 Element Plus 默认样式 */
.k8s-events-table,
.k8s-events-table.el-table,
:deep(.el-table),
:deep(.el-table__body),
:deep(.el-table__body-wrapper),
:deep(.el-table__inner-wrapper),
:deep(.el-table__header) {
  background-color: transparent !important;
}

/* 表头样式 */
:deep(.el-table__header-wrapper) {
  background-color: transparent !important;

  th.el-table__cell {
    background-color: #f5f7fa !important;
    color: #303133;
    font-weight: 600;
    text-align: left;
  }
}

/* 移除所有行的背景色 */
:deep(.el-table__body-wrapper) {
  background-color: transparent !important;
}

:deep(.el-table__body) {
  background-color: transparent !important;
}

:deep(.el-table__body tr) {
  background-color: transparent !important;
}

:deep(.el-table__body td.el-table__cell) {
  background-color: transparent !important;
}

/* 去掉斑马纹 */
:deep(.el-table--striped .el-table__body tr.el-table__row--striped) {
  background-color: transparent !important;

  td.el-table__cell {
    background-color: transparent !important;
  }
}

/* 移除 hover 效果的背景色 */
:deep(.el-table__body tr:hover > td.el-table__cell) {
  background-color: transparent !important;
}

/* 移除固定列的背景色 */
:deep(.el-table__fixed),
:deep(.el-table__fixed-body-wrapper) {
  background-color: transparent !important;

  .el-table__body tr {
    background-color: transparent !important;
  }

  .el-table__body td.el-table__cell {
    background-color: transparent !important;
  }
}

/* 移除表格容器的背景色 */
:deep(.el-table__inner-wrapper) {
  background-color: transparent !important;
}
</style>
