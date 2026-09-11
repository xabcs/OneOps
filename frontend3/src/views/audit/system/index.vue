<script setup lang="tsx">
  import { ref } from 'vue';
  import { fetchGetSystemEventLogs } from '@/service/api';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';
  import { createTagMap } from '@/utils/common';

  defineOptions({ name: 'AuditSystemEvents' });

  interface SearchParams {
    page: number;
    pageSize: number;
    level: string;
    source: string;
    category: string;
    startTime: string;
    endTime: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 20, level: '', source: '', category: '', startTime: '', endTime: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const levelOptions = [
    { label: '信息', value: 'info' },
    { label: '警告', value: 'warning' },
    { label: '错误', value: 'error' },
    { label: '严重', value: 'critical' }
  ];

  const { columns, data, loading, mobilePagination, getDataByPage } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchGetSystemEventLogs(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 20;
    },
    columns: () => [
      { prop: 'id', label: 'ID', width: 70, align: 'center' },
      {
        prop: 'level',
        label: '级别',
        width: 70,
        align: 'center',
        formatter: row => {
          const t = getLevelTag(row.level);
          return <ElTag type={t.type}>{t.text}</ElTag>;
        }
      },
      { prop: 'source', label: '来源', width: 100, align: 'center' },
      { prop: 'category', label: '分类', width: 100, align: 'center' },
      { prop: 'message', label: '消息', minWidth: 200, align: 'center', showOverflowTooltip: true },
      { prop: 'details', label: '详情', minWidth: 150, align: 'center', showOverflowTooltip: true },
      { prop: 'ip', label: 'IP地址', width: 130, align: 'center' },
      { prop: 'time', label: '事件时间', width: 160, align: 'center' }
    ]
  });

  /** 事件级别 → ElTag 标签映射 */
  const getLevelTag = createTagMap({
    info: { text: '信息', type: 'info' },
    warning: { text: '警告', type: 'warning' },
    error: { text: '错误', type: 'danger' },
    critical: { text: '严重', type: 'danger' }
  });

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
  }
</script>

<template>
  <ListPageLayout
    title="系统日志"
    description="查看系统事件日志，追踪级别、来源与分类信息"
    :pagination="mobilePagination"
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElSelect v-model="searchParams.level" placeholder="请选择级别" clearable class="w-150px">
        <ElOption v-for="option in levelOptions" :key="option.value" :label="option.label" :value="option.value" />
      </ElSelect>
      <ElInput v-model="searchParams.source" placeholder="请输入来源" clearable class="w-150px" />
      <ElInput v-model="searchParams.category" placeholder="请输入分类" clearable class="w-150px" />
      <ElDatePicker
        v-model="searchParams.startTime"
        type="datetime"
        placeholder="开始时间"
        value-format="YYYY-MM-DD HH:mm:ss"
        class="w-200px"
      />
      <span class="self-center text-13px color-[var(--el-text-color-secondary)]">至</span>
      <ElDatePicker
        v-model="searchParams.endTime"
        type="datetime"
        placeholder="结束时间"
        value-format="YYYY-MM-DD HH:mm:ss"
        class="w-200px"
      />
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" :data="data" border stripe height="100%">
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>
  </ListPageLayout>
</template>
