<script setup lang="ts">
  import { ElButton, ElPagination, ElTable, ElTableColumn } from 'element-plus';

  defineProps<{
    data: K8s.WorkloadRow[];
    loading: boolean;
    pagination: { page: number; pageSize: number; itemCount: number };
    selectedCount: number;
    batchButtons: Array<{
      label: string;
      type?: string;
      disabled?: boolean;
      handler: () => void;
    }>;
  }>();

  const emit = defineEmits<{
    (e: 'page-change', page: number): void;
    (e: 'size-change', pageSize: number): void;
    (e: 'selection-change', selection: K8s.WorkloadRow[]): void;
    (e: 'select-all', selection: K8s.WorkloadRow[]): void;
    (e: 'clear-selection'): void;
  }>();

  const headerCellStyle = {
    background: '#f5f7fa',
    color: '#303133',
    fontWeight: '600',
    paddingLeft: '16px',
    paddingRight: '16px'
  };

  const rowStyle = { backgroundColor: 'transparent' };
  const cellStyle = { backgroundColor: 'transparent', padding: '8px 16px' };
</script>

<template>
  <div v-loading="loading" class="table-container">
    <ElTable
      :data="data"
      class="workloads-table"
      :header-cell-style="headerCellStyle"
      :row-style="rowStyle"
      :cell-style="cellStyle"
      @selection-change="emit('selection-change', $event)"
      @select-all="emit('select-all', $event)"
    >
      <ElTableColumn type="selection" width="50" align="center" />
      <slot />
    </ElTable>

    <div class="bottom-toolbar">
      <div class="batch-operations-inline">
        <span class="batch-info">已选择 {{ selectedCount }} 项</span>
        <div class="batch-actions">
          <template v-for="btn in batchButtons" :key="btn.label">
            <ElButton
              size="default"
              :type="btn.type as 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'default' | undefined"
              :disabled="btn.disabled"
              @click="btn.handler"
            >
              {{ btn.label }}
            </ElButton>
          </template>
          <ElButton size="default" @click="emit('clear-selection')">取消选择</ElButton>
        </div>
      </div>
      <ElPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.itemCount"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="emit('page-change', $event)"
        @size-change="emit('size-change', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
  .table-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
    min-height: 0;
    background-color: transparent !important;
  }

  .workloads-table {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding-bottom: 60px;

    &,
    &.el-table,
    :deep(.el-table),
    :deep(.el-table__body),
    :deep(.el-table__body-wrapper),
    :deep(.el-table__inner-wrapper),
    :deep(.el-table__header) {
      background-color: transparent !important;
    }

    :deep(.el-table__header-wrapper) {
      background-color: transparent !important;
      th.el-table__cell {
        background-color: #f5f7fa !important;
        color: #303133;
        font-weight: 600;
        text-align: left;
        position: sticky;
        top: 0;
        z-index: 1;
      }
    }

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

    &.el-table--striped {
      :deep(.el-table__body tr.el-table__row--striped) {
        background-color: transparent !important;
        td.el-table__cell {
          background-color: transparent !important;
        }
      }
    }

    :deep(.el-table__body tr:hover > td.el-table__cell) {
      background-color: transparent !important;
    }

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

    :deep(.el-table__inner-wrapper) {
      background-color: transparent !important;
    }

    .operation-buttons {
      display: flex;
      align-items: center;
      gap: 8px;

      .el-button {
        margin: 0;
        padding: 0;
        font-size: var(--el-font-size-base) !important;
        font-weight: 400 !important;
      }

      .dropdown-link {
        display: inline-flex;
        align-items: center;
        gap: 2px;
        cursor: pointer;
        color: var(--el-color-primary);
        font-size: var(--el-font-size-base) !important;
        font-weight: 400;
        user-select: none;

        .dropdown-icon {
          font-size: 14px;
        }

        &:hover {
          opacity: 0.8;
        }
      }
    }
  }

  .bottom-toolbar {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background-color: #fff;
    border-top: 1px solid #ebeef5;
    gap: 16px;
    z-index: 10;

    .batch-operations-inline {
      display: flex;
      align-items: center;
      gap: 16px;

      .batch-info {
        font-size: 14px;
        color: #303133;
        font-weight: 500;
        white-space: nowrap;
      }

      .batch-actions {
        display: flex;
        gap: 8px;
        :deep(.el-button.is-disabled) {
          opacity: 0.5;
        }
      }
    }

    :deep(.el-pagination) {
      margin: 0;
    }
  }
</style>
