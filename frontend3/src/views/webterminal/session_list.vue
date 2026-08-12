<script setup lang="ts">
  import { onMounted } from 'vue';
  import { useSessionList } from './composables/useSessionList';
  import SessionToolbar from './components/SessionToolbar.vue';
  import SessionTable from './components/SessionTable.vue';

  defineOptions({
    name: 'WebterminalSessionList'
  });

  const {
    activeTab,
    searchKeyword,
    selectedSessionIds,
    loading,
    dataSource,
    pagination,
    loadSessions,
    handleTabChange,
    handleSearch,
    handleClearSearch,
    handleSelectAll,
    handleSelectRow,
    handleBatchTerminate,
    handlePageChange,
    handlePageSizeChange,
    handleRefresh,
    handleTableAction
  } = useSessionList();

  onMounted(() => {
    loadSessions();
  });
</script>

<template>
  <div class="session-list-page">
    <!-- 工具栏 + 标签页 + 搜索栏 -->
    <SessionToolbar
      :active-tab="activeTab"
      :search-keyword="searchKeyword"
      :selected-count="selectedSessionIds.length"
      @tab-change="handleTabChange"
      @search="handleSearch"
      @update:search-keyword="searchKeyword = $event"
      @clear-search="handleClearSearch"
      @batch-terminate="handleBatchTerminate"
      @refresh="handleRefresh"
    />

    <!-- 数据表格 + 分页 -->
    <SessionTable
      :data-source="dataSource"
      :selected-session-ids="selectedSessionIds"
      :loading="loading"
      :pagination="pagination"
      @select-all="handleSelectAll"
      @select-row="handleSelectRow"
      @table-action="handleTableAction"
      @page-change="handlePageChange"
      @page-size-change="handlePageSizeChange"
    />
  </div>
</template>

<style lang="scss" scoped>
  .session-list-page {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #171717;
    color: #fff;
    font-size: 12px;
  }
</style>
