<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { fetchGetOperationLogs, fetchExportOperationLogs, fetchGetModules } from '@/service/api';
import { ElNotification } from 'element-plus';
import { exportFile } from '@/utils/file';

defineOptions({ name: 'AuditOperationLogs' });

const searchForm = reactive<Audit.LogQuery>({
  username: '',
  module: '',
  action: '',
  status: '',
  startTime: '',
  endTime: ''
});

const tableData = ref<Audit.OperationLog[]>([]);
const loading = ref(false);
const total = ref(0);
const modules = ref<string[]>([]);

const pagination = reactive({
  page: 1,
  pageSize: 20
});

const statusOptions = [
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' }
];

async function getOperationLogs() {
  loading.value = true;
  try {
    const { data, error } = await fetchGetOperationLogs({
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize
    });

    if (!error && data) {
      tableData.value = data.list || [];
      total.value = data.total || 0;
    }
  } catch (err) {
    console.error('获取操作日志失败:', err);
    ElNotification({
      title: '错误',
      message: '获取操作日志失败',
      type: 'error'
    });
  } finally {
    loading.value = false;
  }
}

async function getModules() {
  try {
    const { data } = await fetchGetModules();
    if (data) {
      modules.value = data || [];
    }
  } catch (err) {
    console.error('获取模块列表失败:', err);
  }
}

function handleSearch() {
  pagination.page = 1;
  getOperationLogs();
}

function handleReset() {
  Object.assign(searchForm, {
    username: '',
    module: '',
    action: '',
    status: '',
    startTime: '',
    endTime: ''
  });
  pagination.page = 1;
  getOperationLogs();
}

function handlePageChange(page: number) {
  pagination.page = page;
  getOperationLogs();
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  getOperationLogs();
}

async function handleExport() {
  try {
    const blob = await fetchExportOperationLogs(searchForm);
    exportFile(blob, 'operation_logs.csv');
    ElNotification({
      title: '成功',
      message: '操作日志导出成功',
      type: 'success'
    });
  } catch (error) {
    console.error('导出失败:', error);
    ElNotification({
      title: '错误',
      message: '导出操作日志失败',
      type: 'error'
    });
  }
}

function getStatusTag(status: string) {
  const statusMap: Record<string, { text: string; type: any }> = {
    success: { text: '成功', type: 'success' },
    failed: { text: '失败', type: 'danger' }
  };
  return statusMap[status] || { text: status, type: 'info' };
}

function getMethodTag(method: string) {
  const methodMap: Record<string, { text: string; type: any }> = {
    GET: { text: 'GET', type: '' },
    POST: { text: 'POST', type: 'success' },
    PUT: { text: 'PUT', type: 'warning' },
    DELETE: { text: 'DELETE', type: 'danger' }
  };
  return methodMap[method] || { text: method, type: 'info' };
}

onMounted(() => {
  getModules();
  getOperationLogs();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <ElForm :model="searchForm" inline class="search-form">
        <ElFormItem label="用户名">
          <ElInput v-model="searchForm.username" placeholder="请输入用户名" clearable style="width: 200px" />
        </ElFormItem>
        <ElFormItem label="模块">
          <ElSelect v-model="searchForm.module" placeholder="请选择模块" clearable style="width: 150px">
            <ElOption v-for="module in modules" :key="module" :label="module" :value="module" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="操作">
          <ElInput v-model="searchForm.action" placeholder="请输入操作" clearable style="width: 150px" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="searchForm.status" placeholder="请选择状态" clearable style="width: 150px">
            <ElOption
              v-for="option in statusOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="操作时间">
          <ElDatePicker
            v-model="searchForm.startTime"
            type="datetime"
            placeholder="开始时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px"
          />
          <span class="mx-2">至</span>
          <ElDatePicker
            v-model="searchForm.endTime"
            type="datetime"
            placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">
            搜索
          </ElButton>
          <ElButton @click="handleReset">
            重置
          </ElButton>
          <ElButton
            type="success"
            @click="handleExport"
          >
            导出
          </ElButton>
        </ElFormItem>
      </ElForm>

      <ElTable
        v-loading="loading"
        :data="tableData"
        border
        stripe
        class="h-full"
        height="calc(100vh - 400px)"
      >
        <ElTableColumn prop="id" label="ID" width="70" align="center" />
        <ElTableColumn prop="username" label="用户名" width="90" align="center" />
        <ElTableColumn prop="module" label="模块" width="100" align="center" />
        <ElTableColumn prop="action" label="操作" width="100" align="center" />
        <ElTableColumn prop="description" label="描述" min-width="150" align="center" show-overflow-tooltip />
        <ElTableColumn label="方法" width="70" align="center">
          <template #default="{ row }">
            <ElTag :type="getMethodTag(row.method).type" size="small">
              {{ getMethodTag(row.method).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="path" label="路径" min-width="180" align="center" show-overflow-tooltip />
        <ElTableColumn prop="statusCode" label="状态码" width="80" align="center" />
        <ElTableColumn prop="ip" label="IP地址" width="130" align="center" />
        <ElTableColumn label="耗时(ms)" width="80" align="center">
          <template #default="{ row }">
            <span :class="{ 'text-warning': row.duration > 1000 }">{{ row.duration }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusTag(row.status).type">
              {{ getStatusTag(row.status).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="time" label="操作时间" width="160" align="center" />
      </ElTable>

      <div class="flex justify-end mt-16px">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
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

.text-warning {
  color: var(--el-color-warning);
}
</style>
