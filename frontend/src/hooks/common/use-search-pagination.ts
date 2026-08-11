/**
 * 搜索分页操作的通用 Hook
 *
 * 用于统一处理各种列表页面的搜索和分页逻辑，减少代码重复
 *
 * @example
 * const {
 *   data,
 *   loading,
 *   searchParams,
 *   pagination,
 *   handleSearch,
 *   handleReset,
 *   handlePageChange
 * } = useSearchPagination(fetchGetServers, {
 *   defaultParams: { agentStatus: '', keyword: '' }
 * });
 */

import { reactive, ref } from 'vue';
import type { Ref } from 'vue';

export interface SearchPaginationOptions<T> {
  /**
   * 数据获取API函数
   * @param params 搜索参数
   * @returns Promise<ApiResponse>
   */
  apiFn: (params: T) => Promise<any>;

  /**
   * 默认搜索参数
   */
  defaultParams?: T;

  /**
   * 默认分页大小
   * @default 20
   */
  defaultPageSize?: number;

  /**
   * 是否在初始化时自动加载数据
   * @default true
   */
  autoLoad?: boolean;
}

export interface SearchPaginationReturn<T, D> {
  /** 列表数据 */
  data: Ref<D[]>;
  /** 加载状态 */
  loading: Ref<boolean>;
  /** 搜索参数 */
  searchParams: Ref<T>;
  /** 分页信息 */
  pagination: {
    page: number;
    pageSize: number;
    total: number;
  };
  /** 获取数据函数 */
  fetchData: () => Promise<void>;
  /** 搜索处理函数 */
  handleSearch: () => void;
  /** 重置处理函数 */
  handleReset: () => void;
  /** 页码变化处理函数 */
  handlePageChange: (page: number) => void;
  /** 每页大小变化处理函数 */
  handlePageSizeChange: (pageSize: number) => void;
}

export function useSearchPagination<T extends { page?: number; pageSize?: number }, D = any>(
  options: SearchPaginationOptions<T>
): SearchPaginationReturn<T, D> {
  const { apiFn, defaultParams, defaultPageSize = 20, autoLoad = true } = options;

  // 状态管理
  const data = ref<D[]>([]) as Ref<D[]>;
  const loading = ref(false);

  // 分页信息
  const pagination = reactive({
    page: 1,
    pageSize: defaultPageSize,
    total: 0
  });

  // 搜索参数
  const searchParams = ref<T>({
    page: 1,
    pageSize: defaultPageSize,
    ...defaultParams
  }) as Ref<T>;

  // 获取数据
  const fetchData = async () => {
    loading.value = true;
    try {
      const result = await apiFn(searchParams.value);
      if (result?.data) {
        data.value = result.data.list || result.data.records || result.data || [];
        pagination.total = result.data.total || 0;
      }
    } catch (error) {
      console.error('获取数据失败:', error);
      data.value = [];
      pagination.total = 0;
    } finally {
      loading.value = false;
    }
  };

  // 搜索处理
  const handleSearch = () => {
    searchParams.value.page = 1;
    pagination.page = 1;
    fetchData();
  };

  // 重置处理
  const handleReset = () => {
    searchParams.value = {
      page: 1,
      pageSize: pagination.pageSize,
      ...defaultParams
    } as T;
    pagination.page = 1;
    fetchData();
  };

  // 页码变化处理
  const handlePageChange = (page: number) => {
    searchParams.value.page = page;
    pagination.page = page;
    fetchData();
  };

  // 每页大小变化处理
  const handlePageSizeChange = (pageSize: number) => {
    searchParams.value.pageSize = pageSize;
    searchParams.value.page = 1;
    pagination.pageSize = pageSize;
    pagination.page = 1;
    fetchData();
  };

  // 自动加载
  if (autoLoad) {
    fetchData();
  }

  return {
    data,
    loading,
    searchParams,
    pagination,
    fetchData,
    handleSearch,
    handleReset,
    handlePageChange,
    handlePageSizeChange
  };
}
