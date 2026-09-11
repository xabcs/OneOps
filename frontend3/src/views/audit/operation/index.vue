<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { fetchExportOperationLogs, fetchGetModules, fetchGetOperationLogs } from '@/service/api';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';
  import { exportFile } from '@/utils/file';
  import { createTagMap } from '@/utils/common';

  defineOptions({ name: 'AuditOperationLogs' });

  interface SearchParams {
    page: number;
    pageSize: number;
    username: string;
    module: string;
    action: string;
    status: string;
    startTime: string;
    endTime: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 20, username: '', module: '', action: '', status: '', startTime: '', endTime: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const statusOptions = [
    { label: '成功', value: 'success' },
    { label: '失败', value: 'failed' }
  ];

  const modules = ref<string[]>([]);

  async function getModules() {
    const { data } = await fetchGetModules();
    if (data) {
      modules.value = data || [];
    }
  }

  const { columns, columnChecks, data, loading, mobilePagination, getDataByPage } = useUIPaginatedTable({
    // 列配置（勾选 + 顺序）持久化到 localStorage，刷新页面后保留
    columnSettingKey: 'audit-operation-logs',
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchGetOperationLogs(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 20;
    },
    columns: () => [
      { prop: 'id', label: 'ID', width: 70, align: 'center' },
      { prop: 'username', label: '用户名', width: 90, align: 'center' },
      { prop: 'module', label: '模块', width: 100, align: 'center' },
      { prop: 'action', label: '操作', width: 100, align: 'center' },
      { prop: 'description', label: '描述', minWidth: 150, align: 'center', showOverflowTooltip: true },
      {
        prop: 'method',
        label: '方法',
        width: 70,
        align: 'center',
        formatter: row => {
          const t = getMethodTag(row.method);
          return (
            <ElTag type={t.type} size="small">
              {t.text}
            </ElTag>
          );
        }
      },
      { prop: 'path', label: '路径', minWidth: 180, align: 'center', showOverflowTooltip: true },
      { prop: 'statusCode', label: '状态码', width: 80, align: 'center' },
      { prop: 'ip', label: 'IP地址', width: 130, align: 'center' },
      {
        prop: 'duration',
        label: '耗时(ms)',
        width: 80,
        align: 'center',
        formatter: row => <span class={{ 'text-warning': row.duration > 1000 }}>{row.duration}</span>
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
      { prop: 'time', label: '操作时间', width: 160, align: 'center' }
    ]
  });

  /** 操作状态 → ElTag 标签映射 */
  const getStatusTag = createTagMap({
    success: { text: '成功', type: 'success' },
    failed: { text: '失败', type: 'danger' }
  });

  /** 请求方法 → ElTag 标签映射 */
  const getMethodTag = createTagMap({
    GET: { text: 'GET', type: 'primary' },
    POST: { text: 'POST', type: 'success' },
    PUT: { text: 'PUT', type: 'warning' },
    DELETE: { text: 'DELETE', type: 'danger' }
  });

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
  }

  async function handleExport() {
    try {
      const { data, error } = await fetchExportOperationLogs(searchParams.value);
      if (!error && data) {
        exportFile(data, 'operation_logs.csv');
        ElMessage.success('操作日志导出成功');
      } else {
        ElMessage.error('导出操作日志失败');
      }
    } catch (error) {
      console.error('导出失败:', error);
      ElMessage.error('导出操作日志失败');
    }
  }

  onMounted(() => {
    getModules();
  });
</script>

<template>
  <ListPageLayout
    title="操作日志"
    description="查看系统操作记录，追踪用户请求与执行结果"
    :pagination="mobilePagination"
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElInput v-model="searchParams.username" placeholder="请输入用户名" clearable class="w-200px" />
      <ElSelect v-model="searchParams.module" placeholder="请选择模块" clearable class="w-150px">
        <ElOption v-for="m in modules" :key="m" :label="m" :value="m" />
      </ElSelect>
      <ElInput v-model="searchParams.action" placeholder="请输入操作" clearable class="w-150px" />
      <ElSelect v-model="searchParams.status" placeholder="请选择状态" clearable class="w-150px">
        <ElOption v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
      </ElSelect>
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

    <!-- 工具栏 -->
    <template #toolbar>
      <TableColumnSetting v-model:columns="columnChecks" />
      <PermissionButton code="audit.operation_log.export" type="success" @click="handleExport">
        导出
      </PermissionButton>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" :data="data" border stripe height="100%">
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>
  </ListPageLayout>
</template>

<style scoped lang="scss">
  .text-warning {
    color: var(--el-color-warning);
  }
</style>
