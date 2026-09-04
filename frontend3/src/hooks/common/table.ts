import { computed, effectScope, nextTick, onScopeDispose, reactive, shallowRef, watch } from 'vue';
import type { Ref } from 'vue';
import { useRoute } from 'vue-router';
import type { PaginationEmits, PaginationProps } from 'element-plus';
import { useBoolean, useTable } from '@sa/hooks';
import type { PaginationData, TableColumnCheck, UseTableOptions } from '@sa/hooks';
import type { FlatResponseData } from '@sa/axios';
import { jsonClone } from '@sa/utils';
import { useAppStore } from '@/store/modules/app';
import { localStg } from '@/utils/storage';
import { $t } from '@/locales';

type RemoveReadonly<T> = {
  -readonly [key in keyof T]: T[key];
};

export type UseUITableOptions<ResponseData, ApiData, Pagination extends boolean> = Omit<
  UseTableOptions<ResponseData, ApiData, UI.TableColumn<ApiData>, Pagination>,
  'pagination' | 'getColumnChecks' | 'getColumns'
> & {
  /**
   * get column visible
   *
   * @param column
   *
   * @default true
   *
   * @returns true if the column is visible, false otherwise
   */
  getColumnVisible?: (column: UI.TableColumn<ApiData>) => boolean;
  /**
   * 列配置（勾选状态 + 顺序）持久化 key
   * - 字符串：以该 key 持久化，需全局唯一（如 'cmdb-servers'）
   * - true：自动使用当前路由名作为 key
   * - 不传：不持久化（默认，保持原有行为）
   */
  columnSettingKey?: string | true;
};

const SELECTION_KEY = '__selection__';

const EXPAND_KEY = '__expand__';

const INDEX_KEY = '__index__';

export function useUITable<ResponseData, ApiData>(options: UseUITableOptions<ResponseData, ApiData, false>) {
  const scope = effectScope();
  const appStore = useAppStore();

  const result = useTable<ResponseData, ApiData, UI.TableColumn<ApiData>, false>({
    ...options,
    getColumnChecks: cols => getColumnChecks(cols, options.getColumnVisible),
    getColumns
  });

  // 用户级列配置持久化（勾选状态 + 顺序）
  applyColumnCheckPersistence(result.columnChecks, options.columnSettingKey);

  // calculate the total width of the table this is used for horizontal scrolling
  const scrollX = computed(() => {
    return result.columns.value.reduce((acc, column) => {
      return acc + Number(column.width ?? column.minWidth ?? 120);
    }, 0);
  });

  scope.run(() => {
    watch(
      () => appStore.locale,
      () => {
        result.reloadColumns();
      }
    );
  });

  onScopeDispose(() => {
    scope.stop();
  });

  return {
    ...result,
    scrollX
  };
}

type PaginationParams = Pick<PaginationProps, 'currentPage' | 'pageSize'>;

type UseUIPaginatedTableOptions<ResponseData, ApiData> = UseUITableOptions<ResponseData, ApiData, true> & {
  paginationProps?: Partial<Omit<PaginationProps, 'total'>>;
  /**
   * whether to show the total count of the table
   *
   * @default true
   */
  showTotal?: boolean;
  onPaginationParamsChange?: (params: PaginationParams) => void | Promise<void>;
};

export function useUIPaginatedTable<ResponseData, ApiData>(options: UseUIPaginatedTableOptions<ResponseData, ApiData>) {
  const scope = effectScope();
  const appStore = useAppStore();

  const isMobile = computed(() => appStore.isMobile);

  const pagination: Partial<RemoveReadonly<PaginationProps & PaginationEmits>> = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
    pageSizes: [10, 15, 20, 25, 30],
    'current-change': (page: number) => {
      pagination.currentPage = page;

      return true;
    },
    'size-change': (pageSize: number) => {
      pagination.currentPage = 1;
      pagination.pageSize = pageSize;

      return true;
    },
    ...options.paginationProps
  }) as PaginationProps;

  // this is for mobile, if the system does not support mobile, you can use `pagination` directly
  const mobilePagination = computed(() => {
    const p: Partial<RemoveReadonly<PaginationProps & PaginationEmits>> = {
      ...pagination,
      pagerCount: isMobile.value ? 3 : 9
    };

    return p;
  });

  const paginationParams = computed(() => {
    const { currentPage, pageSize } = pagination;

    return {
      currentPage,
      pageSize
    };
  });

  const result = useTable<ResponseData, ApiData, UI.TableColumn<ApiData>, true>({
    ...options,
    pagination: true,
    getColumnChecks: cols => getColumnChecks(cols, options.getColumnVisible),
    getColumns,
    onFetched: data => {
      pagination.total = data.total;
    }
  });

  // 用户级列配置持久化（勾选状态 + 顺序）
  applyColumnCheckPersistence(result.columnChecks, options.columnSettingKey);

  async function getDataByPage(page: number = 1) {
    if (page !== pagination.currentPage) {
      pagination.currentPage = page;
    }

    await result.getData();
  }

  scope.run(() => {
    watch(
      () => appStore.locale,
      () => {
        result.reloadColumns();
      }
    );

    watch(paginationParams, async newVal => {
      await options.onPaginationParamsChange?.(newVal);

      await result.getData();
    });
  });

  onScopeDispose(() => {
    scope.stop();
  });

  return {
    ...result,
    getDataByPage,
    pagination,
    mobilePagination
  };
}

export function useTableOperate<TableData>(
  data: Ref<TableData[]>,
  idKey: keyof TableData,
  getData: () => Promise<void>
) {
  const { bool: drawerVisible, setTrue: openDrawer, setFalse: closeDrawer } = useBoolean();

  const operateType = shallowRef<UI.TableOperateType>('add');

  /** the editing row data */
  const editingData = shallowRef<TableData | null>(null);

  function handleAdd() {
    operateType.value = 'add';
    editingData.value = null;

    // 等待 Vue 响应式更新后再打开抽屉
    nextTick(() => {
      openDrawer();
    });
  }

  function handleEdit(id: TableData[keyof TableData]) {
    operateType.value = 'edit';
    const findItem = data.value.find(item => item[idKey] === id) || null;
    editingData.value = jsonClone(findItem);

    // 等待 Vue 响应式更新后再打开抽屉
    nextTick(() => {
      openDrawer();
    });
  }

  /** the checked row keys of table */
  const checkedRowKeys = shallowRef<string[]>([]);

  /** the hook after the batch delete operation is completed */
  async function onBatchDeleted() {
    window.$message?.success($t('common.deleteSuccess'));

    checkedRowKeys.value = [];

    await getData();
  }

  /** the hook after the delete operation is completed */
  async function onDeleted() {
    window.$message?.success($t('common.deleteSuccess'));

    await getData();
  }

  return {
    drawerVisible,
    openDrawer,
    closeDrawer,
    operateType,
    handleAdd,
    editingData,
    handleEdit,
    checkedRowKeys,
    onBatchDeleted,
    onDeleted
  };
}

export function defaultTransform<ApiData>(
  response: FlatResponseData<unknown, Api.Common.PaginatingQueryRecord<ApiData> | ApiData[]>
): PaginationData<ApiData> {
  const { data, error } = response;

  if (!error) {
    // 检查是否是简单的数组格式（后端返回格式）
    if (Array.isArray(data)) {
      return {
        data,
        pageNum: 1,
        pageSize: data.length,
        total: data.length
      };
    }

    // 检查是否是统一分页格式（list/page/pageSize/total）
    if (data && typeof data === 'object' && 'list' in data) {
      const paginated = data as Api.Common.PaginatingQueryRecord<ApiData>;
      return {
        data: paginated.list,
        pageNum: paginated.page,
        pageSize: paginated.pageSize,
        total: paginated.total
      };
    }

    // 兼容旧格式（records/current/size/total）
    if (data && typeof data === 'object' && 'records' in data) {
      const { records, current, size, total } = data as {
        records: ApiData[];
        current: number;
        size: number;
        total: number;
      };
      return {
        data: records,
        pageNum: current,
        pageSize: size,
        total
      };
    }

    // 如果是其他格式，返回空数据
    return {
      data: [],
      pageNum: 1,
      pageSize: 10,
      total: 0
    };
  }

  return {
    data: [],
    pageNum: 1,
    pageSize: 10,
    total: 0
  };
}

type PersistedColumnCheck = { prop: string; checked: boolean };

/**
 * 为 columnChecks 接入用户级持久化：
 * - 初始化时按保存的顺序与勾选状态恢复（代码中新增的列追加到末尾，已删除的列自动忽略）
 * - 用户在列设置抽屉中勾选/拖拽后自动写回 localStorage
 */
function applyColumnCheckPersistence(columnChecks: Ref<TableColumnCheck[]>, keyOption?: string | true) {
  if (!keyOption) return;

  let key: string;
  if (keyOption === true) {
    const route = useRoute();
    key = String(route.name ?? '');
  } else {
    key = keyOption;
  }
  if (!key) return;

  // 恢复：以保存的顺序为准，合并勾选状态
  const saved = localStg.get('tableColumnSettings')?.[key];
  if (saved && saved.length > 0) {
    columnChecks.value = mergeSavedChecks(columnChecks.value, saved);
  }

  // 保存：勾选或拖拽排序都会触发
  watch(
    columnChecks,
    val => {
      const all = localStg.get('tableColumnSettings') ?? {};
      all[key] = val.map(check => ({ prop: check.prop, checked: check.checked }));
      localStg.set('tableColumnSettings', all);
    },
    { deep: true }
  );
}

function mergeSavedChecks(defaults: TableColumnCheck[], saved: PersistedColumnCheck[]): TableColumnCheck[] {
  const checkMap = new Map(defaults.map(check => [check.prop, check]));
  const merged: TableColumnCheck[] = [];

  for (const item of saved) {
    const match = checkMap.get(item.prop);
    if (match) {
      merged.push({ ...match, checked: item.checked });
      checkMap.delete(item.prop);
    }
  }
  // 代码中新增的列（保存时不存在的）追加到末尾，保持默认可见性
  merged.push(...checkMap.values());

  return merged;
}

function getColumnChecks<Column extends UI.TableColumn<unknown>>(
  cols: Column[],
  getColumnVisible?: (column: Column) => boolean
) {
  const checks: TableColumnCheck[] = [];

  cols.forEach(column => {
    if (column.type === 'selection') {
      checks.push({
        prop: SELECTION_KEY,
        label: $t('common.check'),
        checked: true,
        visible: getColumnVisible?.(column) ?? false
      });
    } else if (column.type === 'expand') {
      checks.push({
        prop: EXPAND_KEY,
        label: $t('common.expandColumn'),
        checked: true,
        visible: getColumnVisible?.(column) ?? false
      });
    } else if (column.type === 'index') {
      checks.push({
        prop: INDEX_KEY,
        label: $t('common.index'),
        checked: true,
        visible: getColumnVisible?.(column) ?? false
      });
    } else {
      checks.push({
        prop: column.prop as string,
        label: column.label as string,
        checked: true,
        visible: getColumnVisible?.(column) ?? true
      });
    }
  });

  return checks;
}

function getColumns<Column extends UI.TableColumn<Record<string, unknown>>>(
  cols: Column[],
  checks: TableColumnCheck[]
) {
  const columnMap = new Map<string, Column>();

  cols.forEach(column => {
    if (column.type === 'selection') {
      columnMap.set(SELECTION_KEY, column);
    } else if (column.type === 'expand') {
      columnMap.set(EXPAND_KEY, column);
    } else if (column.type === 'index') {
      columnMap.set(INDEX_KEY, column);
    } else {
      columnMap.set(column.prop as string, column);
    }
  });

  const filteredColumns = checks.filter(item => item.checked).map(check => columnMap.get(check.prop) as Column);

  return filteredColumns;
}
