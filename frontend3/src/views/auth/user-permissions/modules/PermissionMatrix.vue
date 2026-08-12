<script setup lang="ts">
  import { computed } from 'vue';

  type MatrixRow = Record<string, unknown>;
  type PermissionsDetail = Record<string, unknown>;

  interface Props {
    rows: MatrixRow[];
    columns: MatrixRow[];
    matrix: Record<number | string, Record<number | string, boolean>>;
    rowKey: string;
    colKey: string;
    rowLabel: string;
    colLabel: string;
    permissionsDetail?: PermissionsDetail;
  }

  const props = withDefaults(defineProps<Props>(), {
    permissionsDetail: () => ({})
  });

  const emit = defineEmits<{
    cellClick: [row: MatrixRow, col: MatrixRow, hasPermission: boolean];
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

  // 模板辅助函数：从行/列对象中安全提取键值
  function getRowId(row: MatrixRow): string | number {
    const value = row[props.rowKey];
    return typeof value === 'number' || typeof value === 'string' ? value : String(value ?? '');
  }

  function getColId(col: MatrixRow): string | number {
    const value = col[props.colKey];
    return typeof value === 'number' || typeof value === 'string' ? value : String(value ?? '');
  }

  function getCellLabel(item: MatrixRow, labelKey: string): string {
    const value = item[labelKey];
    return value === null || value === undefined ? '-' : String(value);
  }

  function handleCellClick(row: MatrixRow, col: MatrixRow) {
    if (!row || !col) return;

    const hasPerm = hasPermission(getRowId(row), getColId(col));
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
            :key="getColId(col)"
            class="min-w-100px border border-gray-200 px-4 py-3 text-center font-medium"
          >
            <div class="flex flex-col items-center gap-1">
              <span class="font-medium">{{ getCellLabel(col, colLabel) }}</span>
              <span v-if="col.roleType" class="text-xs text-gray-500">({{ col.roleType }})</span>
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in safeRows" :key="getRowId(row)" class="hover:bg-blue-50">
          <td class="sticky left-0 z-10 border border-gray-200 bg-white px-4 py-3 font-medium">
            <div class="flex flex-col gap-1">
              <span>{{ getCellLabel(row, rowLabel) }}</span>
              <span v-if="row.nickname && row.nickname !== row[rowLabel]" class="text-xs text-gray-500">
                {{ row.nickname }}
              </span>
            </div>
          </td>
          <td
            v-for="col in safeColumns"
            :key="getColId(col)"
            class="cursor-pointer border border-gray-200 px-4 py-3 text-center transition-colors hover:bg-blue-100"
            @click="handleCellClick(row, col)"
          >
            <ElTag :type="hasPermission(getRowId(row), getColId(col)) ? 'success' : 'info'" size="large">
              {{ hasPermission(getRowId(row), getColId(col)) ? '✓' : '✗' }}
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
