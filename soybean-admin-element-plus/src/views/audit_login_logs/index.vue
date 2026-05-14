<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { IconifyIcon } from '@/components/common/iconify-icon';
import { fetchGetLoginLogs, fetchExportLoginLogs } from '@/service/api';
import type { Audit } from '@/typings/api';
import { ElNotification } from 'element-plus';
import { exportFile } from '@/utils/file';

defineOptions({ name: 'AuditLoginLogs' });

// 查询表单
const searchForm = reactive<Audit.LogQuery>({
  username: '',
  status: '',
  location: '',
  startTime: '',
  endTime: ''
});

// 表格数据
const tableData = ref<Audit.LoginLog[]>([]);
const loading = ref(false);
const total = ref(0);

// 分页信息
const pagination = reactive({
  page: 1,
  pageSize: 20
});

// 状态选项
const statusOptions = [
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' }
];

// 获取登录日志
async function getLoginLogs() {
  loading.value = true;
  try {
    const { data, error } = await fetchGetLoginLogs({
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize
    });

    if (!error && data) {
      tableData.value = data.list || [];
      total.value = data.total || 0;
    }
  } catch (err) {
    console.error('获取登录日志失败:', err);
    ElNotification({
      title: '错误',
      message: '获取登录日志失败',
      type: 'error'
    });
  } finally {
    loading.value = false;
  }
}

// 搜索
function handleSearch() {
  pagination.page = 1;
  getLoginLogs();
}

// 重置
function handleReset() {
  Object.assign(searchForm, {
    username: '',
    status: '',
    location: '',
    startTime: '',
    endTime: ''
  });
  pagination.page = 1;
  getLoginLogs();
}

// 分页变化
function handlePageChange(page: number) {
  pagination.page = page;
  getLoginLogs();
}

// 页面大小变化
function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  getLoginLogs();
}

// 导出日志
async function handleExport() {
  try {
    const blob = await fetchExportLoginLogs(searchForm);
    exportFile(blob, 'login_logs.csv');
    ElNotification({
      title: '成功',
      message: '登录日志导出成功',
      type: 'success'
    });
  } catch (error) {
    console.error('导出失败:', error);
    ElNotification({
      title: '错误',
      message: '导出登录日志失败',
      type: 'error'
    });
  }
}

// 状态标签
function getStatusTag(status: string) {
  const statusMap: Record<string, { text: string; type: any }> = {
    success: { text: '成功', type: 'success' },
    failed: { text: '失败', type: 'danger' }
  };
  return statusMap[status] || { text: status, type: 'info' };
}

// 初始化
onMounted(() => {
  getLoginLogs();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <!-- 搜索表单 -->
      <ElForm :model="searchForm" inline class="search-form">
        <ElFormItem label="用户名">
          <ElInput v-model="searchForm.username" placeholder="请输入用户名" clearable style="width: 200px" />
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
        <ElFormItem label="位置">
          <ElInput v-model="searchForm.location" placeholder="请输入位置" clearable style="width: 200px" />
        </ElFormItem>
        <ElFormItem label="登录时间">
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
          <ElButton type="primary" :icon="IconifyIcon('ri:search-line')" @click="handleSearch">
            搜索
          </ElButton>
          <ElButton :icon="IconifyIcon('ri:refresh-line')" @click="handleReset">
            重置
          </ElButton>
          <ElButton
            type="success"
            :icon="IconifyIcon('ri:download-line')"
            @click="handleExport"
          >
            导出
          </ElButton>
        </ElFormItem>
      </ElForm>

      <!-- 数据表格 -->
      <ElTable
        v-loading="loading"
        :data="tableData"
        border
        stripe
        class="h-full"
        height="calc(100vh - 400px)"
      >
        <ElTableColumn prop="id" label="ID" width="80" align="center" />
        <ElTableColumn prop="username" label="用户名" width="120" align="center" />
        <ElTableColumn prop="nickname" label="昵称" width="120" align="center" />
        <ElTableColumn prop="ip" label="IP地址" width="140" align="center" />
        <ElTableColumn label="位置" width="150" align="center">
          <template #default="{ row }">
            <span class="text-tertiary">{{ row.location || '-' }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusTag(row.status).type">
              {{ getStatusTag(row.status).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="失败原因" width="200" align="center">
          <template #default="{ row }">
            <span v-if="row.status === 'failed'" class="text-error">{{ row.failReason }}</span>
            <span v-else class="text-tertiary">-</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="loginTime" label="登录时间" width="180" align="center" />
        <ElTableColumn label="会话时长" width="100" align="center">
          <template #default="{ row }">
            <span v-if="row.duration > 0">{{ Math.floor(row.duration / 60) }}分钟</span>
            <span v-else class="text-tertiary">-</span>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
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

.text-tertiary {
  color: var(--el-text-color-placeholder);
}

.text-error {
  color: var(--el-color-danger);
}
</style>