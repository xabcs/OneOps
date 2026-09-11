<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { fetchGetSessions } from '@/service/api/cmdb';
  // 时间格式化与状态标签映射统一收敛到公共工具
  import { formatDateTime as formatTime, formatDurationHuman as formatDuration } from '@/utils/datetime';
  import { createTagMap } from '@/utils/common';

  defineOptions({ name: 'CmdbAuditSessions' });

  const loading = ref(false);
  const sessions = ref<Bastion.BastionSession[]>([]);
  const total = ref(0);

  const pagination = ref({
    page: 1,
    pageSize: 20
  });

  // ListPageLayout 内联分页适配（替代手写 ElPagination）
  const layoutPagination = computed(() => ({
    total: total.value,
    currentPage: pagination.value.page,
    pageSize: pagination.value.pageSize,
    pageSizes: [10, 20, 50, 100],
    'current-change': handlePageChange,
    'size-change': (size: number) => {
      pagination.value.pageSize = size;
      handleSearch();
    }
  }));

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

  /** 会话状态 → ElTag 标签映射 */
  const getStatusTag = createTagMap({
    active: { text: '活跃', type: 'success' },
    closed: { text: '已关闭', type: 'info' },
    error: { text: '错误', type: 'danger' },
    terminated: { text: '已终止', type: 'warning' }
  });

  onMounted(() => {
    getSessions();
  });
</script>

<template>
  <ListPageLayout
    title="历史会话"
    description="查询堡垒机会话记录，追踪协议、时长与关闭原因"
    :pagination="layoutPagination"
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElSelect v-model="filters.status" placeholder="全部状态" clearable class="w-130px">
        <ElOption v-for="opt in statusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
      </ElSelect>
      <ElSelect v-model="filters.protocol" placeholder="全部协议" clearable class="w-110px">
        <ElOption v-for="opt in protocolOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
      </ElSelect>
      <ElDatePicker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        value-format="YYYY-MM-DD"
        class="w-240px"
      />
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <ElButton type="primary" @click="getSessions">刷新</ElButton>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" :data="sessions" stripe height="100%">
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
          <ElTag :type="getStatusTag(row.status).type" size="small">
            {{ getStatusTag(row.status).text }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="关闭原因" width="150" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.closeReason || '-' }}
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" align="center" width="100" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <ElButton link type="primary" size="small" @click="handleViewDetail(row)">详情</ElButton>
        </template>
      </ElTableColumn>
    </ElTable>
  </ListPageLayout>
</template>
