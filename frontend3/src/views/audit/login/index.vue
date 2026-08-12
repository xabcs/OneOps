<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Refresh, Search } from '@element-plus/icons-vue';
  import { fetchExportLoginLogs, fetchGetLoginLogs } from '@/service/api';
  import { exportFile } from '@/utils/file';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';

  defineOptions({ name: 'AuditLoginLogs' });

  interface SearchParams {
    page: number;
    pageSize: number;
    username: string;
    status: string;
    location: string;
    startTime: string;
    endTime: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 20, username: '', status: '', location: '', startTime: '', endTime: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const statusOptions = [
    { label: '成功', value: 'success' },
    { label: '失败', value: 'failed' }
  ];

  const { columns, data, loading, mobilePagination, getData, getDataByPage } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchGetLoginLogs(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 20;
    },
    columns: () => [
      { prop: 'id', label: 'ID', width: 70, align: 'center' },
      { prop: 'username', label: '用户名', width: 100, align: 'center' },
      { prop: 'nickname', label: '昵称', width: 100, align: 'center' },
      { prop: 'ip', label: 'IP地址', width: 130, align: 'center' },
      {
        prop: 'location',
        label: '位置',
        minWidth: 120,
        align: 'center',
        showOverflowTooltip: true,
        formatter: row => <span class="text-tertiary">{row.location || '-'}</span>
      },
      {
        prop: 'status',
        label: '状态',
        width: 80,
        align: 'center',
        formatter: row => {
          const t = getStatusTag(row.status);
          return <ElTag type={t.type}>{t.text}</ElTag>;
        }
      },
      {
        prop: 'failReason',
        label: '失败原因',
        minWidth: 150,
        align: 'center',
        showOverflowTooltip: true,
        formatter: row =>
          row.status === 'failed' ? <span class="text-error">{row.failReason}</span> : <span class="text-tertiary">-</span>
      },
      { prop: 'loginTime', label: '登录时间', width: 160, align: 'center' },
      {
        prop: 'duration',
        label: '会话时长',
        width: 100,
        align: 'center',
        formatter: row =>
          row.duration > 0 ? <span>{Math.floor(row.duration / 60)}分钟</span> : <span class="text-tertiary">-</span>
      }
    ]
  });

  function getStatusTag(status: string): { text: string; type: 'primary' | 'success' | 'warning' | 'danger' | 'info' } {
    const statusMap: Record<string, { text: string; type: 'primary' | 'success' | 'warning' | 'danger' | 'info' }> = {
      success: { text: '成功', type: 'success' },
      failed: { text: '失败', type: 'danger' }
    };
    return statusMap[status] || { text: status, type: 'info' };
  }

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
  }

  async function handleExport() {
    try {
      const blob = await fetchExportLoginLogs(searchParams.value);
      exportFile(blob, 'login_logs.csv');
      ElMessage.success('登录日志导出成功');
    } catch (error) {
      console.error('导出失败:', error);
      ElMessage.error('导出登录日志失败');
    }
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <ElForm :model="searchParams" inline class="search-form">
        <ElFormItem label="用户名">
          <ElInput v-model="searchParams.username" placeholder="请输入用户名" clearable style="width: 200px" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="searchParams.status" placeholder="请选择状态" clearable style="width: 150px">
            <ElOption v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="位置">
          <ElInput v-model="searchParams.location" placeholder="请输入位置" clearable style="width: 200px" />
        </ElFormItem>
        <ElFormItem label="登录时间">
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
          <ElButton type="success" @click="handleExport">导出</ElButton>
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

  .text-tertiary {
    color: var(--el-text-color-placeholder);
  }

  .text-error {
    color: var(--el-color-danger);
  }
</style>
