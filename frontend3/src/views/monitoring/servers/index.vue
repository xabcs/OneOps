<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { Refresh, Search } from '@element-plus/icons-vue';
  import { fetchGetServers } from '@/service/api/cmdb';
  import { useWebSocket } from '@/service/websocket';

  defineOptions({
    name: 'MonitoringServers'
  });

  const router = useRouter();
  const loading = ref(false);

  const servers = ref<CMDB.Server[]>([]);
  const searchParams = ref({
    page: 1,
    pageSize: 20,
    agentStatus: '',
    keyword: ''
  });

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  });

  async function getServerList() {
    loading.value = true;
    const { data } = await fetchGetServers(searchParams.value);
    if (data) {
      servers.value = data.list || [];
      pagination.value.total = data.total || 0;
    }
    loading.value = false;
  }

  function handleSearch() {
    searchParams.value.page = 1;
    pagination.value.page = 1;
    getServerList();
  }

  function handleReset() {
    searchParams.value = {
      page: 1,
      pageSize: 20,
      agentStatus: '',
      keyword: ''
    };
    pagination.value.page = 1;
    getServerList();
  }

  function handlePageChange(page: number) {
    searchParams.value.page = page;
    pagination.value.page = page;
    getServerList();
  }

  function handlePageSizeChange(pageSize: number) {
    searchParams.value.pageSize = pageSize;
    searchParams.value.page = 1;
    pagination.value.page = 1;
    getServerList();
  }

  // 布局内置分页：托管给 ListPageLayout，复用现有翻页处理器
  const layoutPagination = computed(() => ({
    currentPage: pagination.value.page,
    pageSize: pagination.value.pageSize,
    total: pagination.value.total,
    pageSizes: [20, 50, 100],
    'current-change': handlePageChange,
    'size-change': handlePageSizeChange
  }));

  function handleViewMonitoring(serverId: number) {
    router.push({
      path: '/monitoring/servers/detail',
      query: { id: serverId }
    });
  }

  function getAgentTagType(status: string): 'success' | 'danger' | 'warning' | 'info' {
    const map: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
      running: 'success',
      offline: 'danger',
      failed: 'warning',
      uninstalled: 'info'
    };
    return map[status] ?? 'info';
  }

  function getAgentStatusText(status: string): string {
    const textMap: Record<string, string> = {
      running: '运行中',
      offline: '离线',
      failed: '失败',
      uninstalled: '未安装'
    };
    return textMap[status] || status;
  }

  function getProgressColor(value: number): string {
    if (value > 80) return '#ff4d4f';
    if (value > 60) return '#faad14';
    return '#52c41a';
  }

  function getLoadColor(load: number): string {
    // 系统负载颜色逻辑（考虑 CPU 核心数会有影响，这里简化处理）
    if (load > 5) return '#ff4d4f';
    if (load > 3) return '#faad14';
    return '#52c41a';
  }

  // WebSocket 客户端
  const { subscribe, unsubscribe } = useWebSocket();

  // 处理指标更新
  function handleMetricsUpdate(data: { server_id: number }) {
    // 查找对应的服务器并更新数据
    const server = servers.value.find(s => s.id === data.server_id);
    if (server) {
      server.cpuUsage = data.cpu;
      server.memoryUsage = data.memory;
      server.diskUsage = data.disk;
      server.load5 = data.load5;
      server.metricsUpdatedAt = new Date();
    }
  }

  onMounted(() => {
    getServerList();

    // 订阅 WebSocket 指标更新
    subscribe('metrics_update', handleMetricsUpdate);
  });

  onUnmounted(() => {
    // 取消订阅
    unsubscribe('metrics_update', handleMetricsUpdate);
  });
</script>

<template>
  <ListPageLayout
    title="主机监控"
    description="实时查看主机运行状态与资源使用率"
    :pagination="layoutPagination"
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElInput
        v-model="searchParams.keyword"
        placeholder="搜索主机名或IP"
        clearable
        class="w-200px"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <ElIcon :size="16"><Search /></ElIcon>
        </template>
      </ElInput>
      <ElSelect v-model="searchParams.agentStatus" placeholder="Agent状态" clearable class="w-130px">
        <ElOption label="全部" value="" />
        <ElOption label="运行中" value="running" />
        <ElOption label="离线" value="offline" />
        <ElOption label="失败" value="failed" />
        <ElOption label="未安装" value="uninstalled" />
      </ElSelect>
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <ElButton :loading="loading" @click="getServerList">
        <ElIcon :size="16"><Refresh /></ElIcon>
        刷新
      </ElButton>
    </template>

    <!-- 主机表格 -->
    <ElTable
      v-loading="loading"
      :data="servers"
      border
      stripe
      size="small"
      row-key="id"
      height="100%"
      style="width: 100%"
    >
      <ElTableColumn label="主机名" prop="hostname" min-width="140" show-overflow-tooltip />
      <ElTableColumn label="IP地址" prop="ip" width="130" />
      <ElTableColumn label="Agent状态" width="110" align="center">
        <template #default="{ row }">
          <ElTag :type="getAgentTagType(row.agentStatus)" size="small">
            {{ getAgentStatusText(row.agentStatus) }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="CPU使用率" width="140" align="center">
        <template #default="{ row }">
          <ElProgress
            :percentage="Math.round(row.cpuUsage || 0)"
            :color="getProgressColor(row.cpuUsage || 0)"
            :stroke-width="4"
            :format="(percentage: number) => (row.cpuUsage ? row.cpuUsage.toFixed(1) + '%' : '0%')"
          />
        </template>
      </ElTableColumn>
      <ElTableColumn label="内存使用率" width="140" align="center">
        <template #default="{ row }">
          <ElProgress
            :percentage="Math.round(row.memoryUsage || 0)"
            :color="getProgressColor(row.memoryUsage || 0)"
            :stroke-width="4"
            :format="(percentage: number) => (row.memoryUsage ? row.memoryUsage.toFixed(1) + '%' : '0%')"
          />
        </template>
      </ElTableColumn>
      <ElTableColumn label="磁盘使用率" width="140" align="center">
        <template #default="{ row }">
          <ElProgress
            :percentage="Math.round(row.diskUsage || 0)"
            :color="getProgressColor(row.diskUsage || 0)"
            :stroke-width="4"
            :format="(percentage: number) => (row.diskUsage ? row.diskUsage.toFixed(1) + '%' : '0%')"
          />
        </template>
      </ElTableColumn>
      <ElTableColumn label="系统负载" width="100" align="center">
        <template #default="{ row }">
          <span :style="{ color: getLoadColor(row.load5 || 0), fontSize: '12px', fontWeight: '500' }">
            {{ row.load5 ? row.load5.toFixed(2) : '-' }}
          </span>
        </template>
      </ElTableColumn>
      <ElTableColumn label="最后更新" width="170">
        <template #default="{ row }">
          {{ row.metricsUpdatedAt ? new Date(row.metricsUpdatedAt).toLocaleString('zh-CN') : '-' }}
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" width="110" align="center" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <ElButton
            link
            type="primary"
            size="small"
            :disabled="row.agentStatus !== 'running'"
            @click="handleViewMonitoring(row.id)"
          >
            监控详情
          </ElButton>
        </template>
      </ElTableColumn>
    </ElTable>
  </ListPageLayout>
</template>
