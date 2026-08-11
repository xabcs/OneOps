<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { fetchGetSessions } from '@/service/api/cmdb';

defineOptions({ name: 'CmdbAuditSessions' });

const loading = ref(false);
const sessions = ref<Bastion.BastionSession[]>([]);
const total = ref(0);

const pagination = ref({
  page: 1,
  pageSize: 20
});

const filters = ref<{
  status?: string;
  protocol?: string;
  startDate?: string;
  endDate?: string;
}>({});

const dateRange = ref<[string, string] | null>(null);

const statusOptions = [
  { label: '全部', value: '' },
  { label: '活跃', value: 'active' },
  { label: '已关闭', value: 'closed' },
  { label: '错误', value: 'error' },
  { label: '已终止', value: 'terminated' }
];

const protocolOptions = [
  { label: '全部', value: '' },
  { label: 'SSH', value: 'ssh' },
  { label: 'SFTP', value: 'sftp' }
];

async function getSessions() {
  loading.value = true;
  try {
    const params: Record<string, unknown> = {
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    };
    if (filters.value.status) params.status = filters.value.status;
    if (filters.value.protocol) params.protocol = filters.value.protocol;
    if (dateRange.value) {
      params.startDate = dateRange.value[0];
      params.endDate = dateRange.value[1];
    }

    const { data } = await fetchGetSessions(params);
    sessions.value = data?.list || [];
    total.value = data?.total || 0;
  } catch (error) {
    window.$message?.error('获取会话列表失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.value.page = 1;
  getSessions();
}

function handleReset() {
  filters.value = {};
  dateRange.value = null;
  pagination.value.page = 1;
  getSessions();
}

function handlePageChange(page: number) {
  pagination.value.page = page;
  getSessions();
}

function handleViewDetail(session: Bastion.BastionSession) {
  window.$message?.info('会话详情功能开发中');
}

function formatDuration(seconds: number): string {
  if (!seconds) return '-';
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}小时${minutes}分钟`;
}

function formatTime(time: string): string {
  return time ? new Date(time).toLocaleString('zh-CN') : '-';
}

function getStatusType(status: string): 'success' | 'info' | 'warning' | 'danger' {
  switch (status) {
    case 'active':
      return 'success';
    case 'closed':
      return 'info';
    case 'error':
      return 'danger';
    case 'terminated':
      return 'warning';
    default:
      return 'info';
  }
}

function getStatusText(status: string): string {
  const map: Record<string, string> = {
    active: '活跃',
    closed: '已关闭',
    error: '错误',
    terminated: '已终止'
  };
  return map[status] || status;
}

onMounted(() => {
  getSessions();
});
</script>

<template>
  <div class="sessions-page">
    <ElCard shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">历史会话</span>
          <ElButton type="primary" @click="getSessions">刷新</ElButton>
        </div>
      </template>

      <div class="filter-bar">
        <ElForm :inline="true" :model="filters">
          <ElFormItem label="状态">
            <ElSelect v-model="filters.status" placeholder="全部状态" clearable style="width: 130px">
              <ElOption v-for="opt in statusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="协议">
            <ElSelect v-model="filters.protocol" placeholder="全部协议" clearable style="width: 110px">
              <ElOption v-for="opt in protocolOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="时间范围">
            <ElDatePicker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 240px"
            />
          </ElFormItem>
          <ElFormItem>
            <ElButton type="primary" @click="handleSearch">搜索</ElButton>
            <ElButton @click="handleReset">重置</ElButton>
          </ElFormItem>
        </ElForm>
      </div>

      <ElTable v-loading="loading" :data="sessions" stripe style="width: 100%; margin-top: 16px">
        <ElTableColumn prop="id" label="ID" width="70" />
        <ElTableColumn prop="username" label="用户名" width="110" />
        <ElTableColumn label="服务器" width="160">
          <template #default="{ row }">
            {{ row.server?.hostname || row.server?.ip || `ID:${row.serverId}` }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="loginAccount" label="登录账号" width="120" />
        <ElTableColumn prop="clientIp" label="客户端IP" width="140" />
        <ElTableColumn prop="protocol" label="协议" width="90">
          <template #default="{ row }">
            <ElTag :type="row.protocol === 'ssh' ? 'primary' : 'success'" size="small">
              {{ row.protocol?.toUpperCase() }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.startedAt || '') }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="时长" width="110">
          <template #default="{ row }">
            {{ formatDuration(row.duration) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="关闭原因" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.closeReason || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleViewDetail(row)">详情</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handlePageChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<style scoped>
.sessions-page {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 16px;
  font-weight: 500;
}

.filter-bar {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
