<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import type { FlatResponseData } from '@sa/axios';
  import { fetchApplications, fetchOperationLogs } from '@/service/api/application-permission';
  import { createTagMap } from '@/utils/common';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';

  defineOptions({ name: 'AuthCenterOperationLogs' });

  interface SearchParams {
    page: number;
    pageSize: number;
    appId: number | null;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 10, appId: null };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const applications = ref<Api.ApplicationPermission.Application[]>([]);

  async function getApplications() {
    const { data, error } = await fetchApplications({ page: 1, pageSize: 100 });
    if (!error && data) {
      applications.value = data.list || [];
    }
  }

  const { columns, data, loading, mobilePagination, getDataByPage } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => {
      if (!searchParams.value.appId) {
        return Promise.resolve({
          error: null,
          data: { list: [], total: 0 }
        }) as unknown as ReturnType<typeof fetchOperationLogs>;
      }
      return fetchOperationLogs(searchParams.value.appId!, {
        page: searchParams.value.page,
        pageSize: searchParams.value.pageSize
      });
    },
    transform: response =>
      defaultTransform<Api.ApplicationPermission.ApplicationOperationLog>(
        response as unknown as FlatResponseData<
          unknown,
          Api.Common.PaginatingQueryRecord<Api.ApplicationPermission.ApplicationOperationLog>
        >
      ),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 10;
    },
    columns: () => [
      { prop: 'index', type: 'index', label: '序号', width: 60, align: 'center' },
      { prop: 'operation', label: '操作类型', align: 'center', minWidth: 120 },
      { prop: 'target', label: '操作目标', align: 'center', minWidth: 150 },
      {
        prop: 'status',
        label: '状态',
        align: 'center',
        width: 100,
        formatter: row => {
          const t = getStatusTag(row.status);
          return <ElTag type={t.type}>{t.text}</ElTag>;
        }
      },
      { prop: 'operator', label: '操作人', align: 'center', minWidth: 120 },
      { prop: 'createdAt', label: '操作时间', align: 'center', minWidth: 160 }
    ]
  });

  const getStatusTag = createTagMap({
    success: { text: '成功', type: 'success' },
    failed: { text: '失败', type: 'danger' },
    running: { text: '进行中', type: 'warning' }
  });

  function handleAppChange() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
  }

  onMounted(() => {
    getApplications();
  });
</script>

<template>
  <ListPageLayout
    title="操作日志"
    description="按应用查看授权中心操作记录，追踪操作类型、目标与执行状态"
    :pagination="mobilePagination"
    @search="handleAppChange"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElSelect
        v-model="searchParams.appId"
        placeholder="请选择应用查看操作日志"
        class="w-300px"
        @change="handleAppChange"
      >
        <ElOption v-for="app in applications" :key="app.id" :label="app.name" :value="app.id" />
      </ElSelect>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" :data="data" :border="false" height="100%">
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>
  </ListPageLayout>
</template>
