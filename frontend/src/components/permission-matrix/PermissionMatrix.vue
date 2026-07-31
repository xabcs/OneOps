<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  rows: any[];
  columns: any[];
  matrix: Record<number | string, Record<number | string, boolean>>;
  rowKey: string;
  colKey: string;
  rowLabel: string;
  colLabel: string;
  permissionsDetail?: Record<string, any>;
}

const props = withDefaults(defineProps<Props>(), {
  permissionsDetail: () => ({})
});

const emit = defineEmits<{
  cellClick: [row: any, col: any, hasPermission: boolean];
}>();

// 安全的数据访问
const safeRows = computed(() => props.rows || []);
const safeColumns = computed(() => props.columns || []);
const safeMatrix = computed(() => props.matrix || {});

function hasPermission(rowId: number | string, colId: number | string): boolean {
  // 确保 matrix 存在
  if (!safeMatrix.value || Object.keys(safeMatrix.value).length === 0) {
    return false;
  }

  // 矩阵结构：matrix[colId][rowId]
  // 支持数字和字符串类型的键
  const colIdStr = String(colId);
  const rowIdStr = String(rowId);

  const result = safeMatrix.value[colIdStr]?.[rowIdStr] || false;
  return result;
}

function getCellKey(rowId: number | string, colId: number | string): string {
  return `${colId}_${rowId}`;
}

function handleCellClick(row: any, col: any) {
  if (!row || !col) return;

  const hasPerm = hasPermission(row[props.rowKey], col[props.colKey]);
  emit('cellClick', row, col, hasPerm);
}
</script>

<template>
  <div class="permission-matrix overflow-x-auto">
    <table class="w-full border-collapse">
      <thead>
        <tr class="bg-gray-50">
          <th class="sticky left-0 z-10 min-w-150px border border-gray-200 bg-gray-50 px-4 py-3 text-left font-medium">
            {{ rowLabel }}
          </th>
          <th
            v-for="col in safeColumns"
            :key="col[colKey]"
            class="min-w-100px border border-gray-200 px-4 py-3 text-center font-medium"
          >
            <div class="flex flex-col items-center gap-1">
              <span class="font-medium">{{ col[colLabel] || '-' }}</span>
              <span v-if="col.roleType" class="text-xs text-gray-500">({{ col.roleType }})</span>
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in safeRows" :key="row[rowKey]" class="hover:bg-blue-50">
          <td class="sticky left-0 z-10 border border-gray-200 bg-white px-4 py-3 font-medium">
            <div class="flex flex-col gap-1">
              <span>{{ row[rowLabel] || '-' }}</span>
              <span v-if="row.nickname && row.nickname !== row[rowLabel]" class="text-xs text-gray-500">
                {{ row.nickname }}
              </span>
            </div>
          </td>
          <td
            v-for="col in safeColumns"
            :key="col[colKey]"
            class="cursor-pointer border border-gray-200 px-4 py-3 text-center transition-colors hover:bg-blue-100"
            @click="handleCellClick(row, col)"
          >
            <ElTag :type="hasPermission(row[rowKey], col[colKey]) ? 'success' : 'info'" size="large">
              {{ hasPermission(row[rowKey], col[colKey]) ? '✓' : '✗' }}
            </ElTag>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.permission-matrix {
  max-height: 600px;
  overflow-y: auto;
}

.permission-matrix table {
  border-spacing: 0;
}

.permission-matrix th,
.permission-matrix td {
  border: 1px solid #e5e7eb;
}

.permission-matrix th.sticky,
.permission-matrix td.sticky {
  position: sticky;
  left: 0;
  z-index: 10;
}
</style>
