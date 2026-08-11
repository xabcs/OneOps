/**
 * 搜索/筛选逻辑
 * 从 index.vue 拆分：搜索类型、搜索关键词、搜索与重置
 */

import { reactive, ref } from 'vue';
import type { SearchType } from '../types/server.types';

export function useServerSearch() {
  const searchType = ref<SearchType>('hostname');
  const searchKeyword = ref('');

  // 旧的搜索方式（保持兼容）
  const searchForm = reactive({ hostname: '', ip: '' });

  function handleSearch() {
    return { searchType: searchType.value, searchKeyword: searchKeyword.value };
  }

  function handleReset() {
    searchForm.hostname = '';
    searchForm.ip = '';
    searchKeyword.value = '';
  }

  return {
    searchType,
    searchKeyword,
    searchForm,
    handleSearch,
    handleReset
  };
}
