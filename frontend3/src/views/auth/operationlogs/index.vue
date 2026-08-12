<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { fetchApplications, fetchOperationLogs } from '@/service/api/application-permission';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';
  import type { FlatResponseData } from '@sa/axios';

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
    const { data, error } = await fetchApplications({ page: 1, pageSize: 1000 });
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
          return (
            <ElTag type={t.type}>{t.label}</ElTag>
          );
        }
      },
      { prop: 'operator', label: '操作人', align: 'center', minWidth: 120 },
      { prop: 'createdAt', label: '操作时间', align: 'center', minWidth: 160 }
    ]
  });

  function getStatusTag(status: string): { type: 'primary' | 'success' | 'warning' | 'info' | 'danger'; label: string } {
    const map: Record<string, { type: 'primary' | 'success' | 'warning' | 'info' | 'danger'; label: string }> = {
      success: { type: 'success', label: '成功' },
      failed: { type: 'danger', label: '失败' },
      running: { type: 'warning', label: '进行中' }
    };
    return map[status] || { type: 'primary', label: status };
  }

  function handleAppChange() {
    getDataByPage(1);
  }

  onMounted(() => {
    getApplications();
  });
</script>

<template>
  <div class="min-h-500px flex-col gap-4">
    <!-- 应用选择 -->
    <ElCard shadow="never">
      <ElSelect
        v-model="searchParams.appId"
        placeholder="请选择应用查看操作日志"
        class="w-300px"
        @change="handleAppChange"
      >
        <ElOption v-for="app in applications" :key="app.id" :label="app.name" :value="app.id" />
      </ElSelect>
    </ElCard>

    <!-- 表格 -->
    <ElCard shadow="never" class="flex-1">
      <template #header>
        <span class="text-lg font-medium">操作日志列表</span>
      </template>

      <ElTable v-loading="loading" :data="data" :border="false">
        <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
      </ElTable>

      <div class="mt-4 flex justify-end">
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
