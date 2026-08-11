/**
 * Vue 组合式函数集合
 * 提供可复用的业务逻辑和状态管理
 */

import type { ComputedRef } from 'vue';
import { computed, ref } from 'vue';

// 分页参数接口
export interface PaginationParams {
  currentPage: number;
  pageSize: number;
  total: number;
}

// 搜索和分页组合式函数
export function useSearchAndPagination(initialPageSize = 20) {
  const searchKeyword = ref('');
  const pagination = ref<PaginationParams>({
    currentPage: 1,
    pageSize: initialPageSize,
    total: 0
  });

  // 重置搜索和分页
  const resetSearch = () => {
    searchKeyword.value = '';
    pagination.value.currentPage = 1;
  };

  // 处理搜索
  const handleSearch = (keyword: string) => {
    searchKeyword.value = keyword;
    pagination.value.currentPage = 1;
  };

  // 处理分页变化
  const handlePageChange = (page: number) => {
    pagination.value.currentPage = page;
  };

  // 处理每页数量变化
  const handleSizeChange = (pageSize: number) => {
    pagination.value.pageSize = pageSize;
    pagination.value.currentPage = 1;
  };

  // 更新总数
  const updateTotal = (total: number) => {
    pagination.value.total = total;
  };

  return {
    searchKeyword,
    pagination,
    resetSearch,
    handleSearch,
    handlePageChange,
    handleSizeChange,
    updateTotal
  };
}

// 批量选择组合式函数
export function useBatchSelection<T = any>() {
  const selectedItems = ref<T[]>([]);
  const currentSelection = ref<T[]>([]);

  // 是否有选中项
  const hasSelection: ComputedRef<boolean> = computed(() => selectedItems.value.length > 0);

  // 清除选择
  const clearSelection = () => {
    selectedItems.value = [];
    currentSelection.value = [];
  };

  // 处理选择变化
  const handleSelectionChange = (selection: T[]) => {
    currentSelection.value = selection;
    selectedItems.value = selection;
  };

  // 获取选中项的ID列表
  const getSelectedIds = (key: keyof T = 'id' as keyof T): unknown[] => {
    return selectedItems.value.map(item => item[key]);
  };

  return {
    selectedItems,
    currentSelection,
    hasSelection,
    clearSelection,
    handleSelectionChange,
    getSelectedIds
  };
}

// 确认对话框组合式函数
export function useConfirmDialog() {
  const visible = ref(false);
  const loading = ref(false);
  const title = ref('');
  const message = ref('');
  const type = ref<'warning' | 'info' | 'success'>('warning');

  // 显示确认对话框
  const showConfirm = (
    confirmTitle: string,
    confirmMessage: string,
    dialogType: 'warning' | 'info' | 'success' = 'warning'
  ) => {
    title.value = confirmTitle;
    message.value = confirmMessage;
    type.value = dialogType;
    visible.value = true;
    loading.value = false;
  };

  // 隐藏对话框
  const hideDialog = () => {
    visible.value = false;
  };

  // 设置加载状态
  const setLoading = (isLoading: boolean) => {
    loading.value = isLoading;
  };

  return {
    visible,
    loading,
    title,
    message,
    type,
    showConfirm,
    hideDialog,
    setLoading
  };
}

// 异步操作处理组合式函数
export function useAsyncOperation<T = unknown>() {
  const loading = ref(false);
  const error = ref<Error | null>(null);
  const data = ref<T | null>(null);

  // 执行异步操作
  const execute = async (operation: () => Promise<T>) => {
    loading.value = true;
    error.value = null;

    try {
      const result = await operation();
      data.value = result;
      return result;
    } catch (err) {
      error.value = err as Error;
      throw err;
    } finally {
      loading.value = false;
    }
  };

  // 重置状态
  const reset = () => {
    loading.value = false;
    error.value = null;
    data.value = null;
  };

  return {
    loading,
    error,
    data,
    execute,
    reset
  };
}

// 表单验证组合式函数
export function useFormValidation<T extends Record<string, any>>(
  initialValues: T,
  validationRules: Record<keyof T, (value: any) => string | null>
) {
  const formData = ref<T>({ ...initialValues });
  const errors = ref<Record<keyof T, string | null>>({} as Record<keyof T, string | null>);

  // 验证表单
  const validate = (): boolean => {
    let isValid = true;

    for (const field in validationRules) {
      const validator = validationRules[field];
      const error = validator(formData.value[field]);

      if (error) {
        errors.value[field] = error;
        isValid = false;
      } else {
        errors.value[field] = null;
      }
    }

    return isValid;
  };

  // 重置表单
  const resetForm = () => {
    formData.value = { ...initialValues };
    errors.value = {} as Record<keyof T, string | null>;
  };

  // 清除错误
  const clearErrors = () => {
    errors.value = {} as Record<keyof T, string | null>;
  };

  return {
    formData,
    errors,
    validate,
    resetForm,
    clearErrors
  };
}
