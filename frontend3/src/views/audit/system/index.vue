<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Refresh, Search } from '@element-plus/icons-vue';
  import { fetchGetSystemEventLogs } from '@/service/api';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';

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

  const { columns, data, loading, mobilePagination, getData, getDataByPage } = useUIPaginatedTable({
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

  function getLevelTag(level: string): { text: string; type: 'primary' | 'success' | 'warning' | 'danger' | 'info' } {
    const levelMap: Record<string, { text: string; type: 'primary' | 'success' | 'warning' | 'danger' | 'info' }> = {
      info: { text: '信息', type: 'info' },
      warning: { text: '警告', type: 'warning' },
      error: { text: '错误', type: 'danger' },
      critical: { text: '严重', type: 'danger' }
    };
    return levelMap[level] || { text: level, type: 'info' };
  }

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <ElForm :model="searchParams" inline class="search-form">
        <ElFormItem label="级别">
          <ElSelect v-model="searchParams.level" placeholder="请选择级别" clearable style="width: 150px">
            <ElOption v-for="option in levelOptions" :key="option.value" :label="option.label" :value="option.value" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="来源">
          <ElInput v-model="searchParams.source" placeholder="请输入来源" clearable style="width: 150px" />
        </ElFormItem>
        <ElFormItem label="分类">
          <ElInput v-model="searchParams.category" placeholder="请输入分类" clearable style="width: 150px" />
        </ElFormItem>
        <ElFormItem label="事件时间">
          <ElDatePicker
            v-model="searchParams.startTime"
            type="datetime"
            placeholder="开始时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px"
          />
          <span class="mx-2">至</span>
          <ElDatePicker
            v-model="searchParams.endTime"
            type="datetime"
            placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" :icon="Search" @click="handleSearch">搜索</ElButton>
          <ElButton :icon="Refresh" @click="handleReset">重置</ElButton>
        </ElFormItem>
      </ElForm>

      <ElTable v-loading="loading" :data="data" border stripe class="h-full" height="calc(100vh - 400px)">
        <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
      </ElTable>

      <div class="mt-16px flex justify-end">
        <ElPagination
          v-if="mobilePagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
    </ElCard>
  </div>
</template>

<style scoped lang="scss">
  .search-form {
    :deep(.el-form-item) {
      margin-bottom: 12px;
    }
  }
</style>
