<script setup lang="tsx">
  import { computed, onMounted, ref, resolveDirective, withDirectives } from 'vue';
  import type { FlatResponseData } from '@sa/axios';
  import {
    fetchBatchUninstallAgent,
    fetchDeleteAgentRecord,
    fetchDeployAgent,
    fetchGetAgentList,
    fetchGetAgentStatus,
    fetchGetLatestAgentVersion,
    fetchRestartAgent,
    fetchUninstallAgent,
    fetchUpgradeAgent
  } from '@/service/api/cmdb';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';
  import PermissionButton from '@/components/common/PermissionButton.vue';

  defineOptions({ name: 'CmdbConfigAgents' });

  // JSX 中自定义指令不生效（v-permission 会被当作普通 prop），用 withDirectives 手动挂载
  const vPermission = resolveDirective('permission')!;

  // ========== 搜索参数 ==========
  interface SearchParams {
    page: number;
    pageSize: number;
    hostname: string;
    ip: string;
    agentStatus: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 20, hostname: '', ip: '', agentStatus: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const selectedRows = ref<CMDB.Server[]>([]);
  const latestVersion = ref<{ version: string; releaseNotes?: string } | null>(null);
  const batchProgress = ref({ current: 0, total: 0, running: false });

  // ========== 表格 ==========
  const { columns, data, loading, mobilePagination, getData, getDataByPage } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => {
      const params: Record<string, unknown> = {
        page: searchParams.value.page,
        pageSize: searchParams.value.pageSize
      };
      if (searchParams.value.hostname) params.hostname = searchParams.value.hostname;
      if (searchParams.value.ip) params.ip = searchParams.value.ip;
      if (searchParams.value.agentStatus) params.agentStatus = searchParams.value.agentStatus;
      return fetchGetAgentList(params);
    },
    transform: response =>
      defaultTransform<CMDB.Server>(
        response as unknown as FlatResponseData<unknown, Api.Common.PaginatingQueryRecord<CMDB.Server>>
      ),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 20;
    },
    columns: () => [
      { prop: 'selection', type: 'selection', width: 50, align: 'center' },
      { prop: 'hostname', label: '主机名称', minWidth: 150, showOverflowTooltip: true },
      { prop: 'ip', label: 'IP地址', width: 140, showOverflowTooltip: true },
      {
        prop: 'innerIp',
        label: '内网IP',
        width: 140,
        showOverflowTooltip: true,
        formatter: row => <span class="text-gray">{row.innerIp || '-'}</span>
      },
      {
        prop: 'agentVersion',
        label: '版本',
        width: 140,
        align: 'center',
        formatter: row =>
          row.agentVersion ? (
            <div class="version-cell">
              <span class="version-text">v{row.agentVersion}</span>
              {latestVersion.value && row.agentStatus === 'running' && (
                <ElTag type={getVersionStatus(row.agentVersion).type} size="small" class="version-tag">
                  {getVersionStatus(row.agentVersion).text}
                </ElTag>
              )}
            </div>
          ) : (
            <span class="text-gray">-</span>
          )
      },
      {
        prop: 'agentStatus',
        label: '状态',
        width: 100,
        align: 'center',
        formatter: row => {
          const t = getAgentStatusTag(row.agentStatus);
          return (
            <ElTag type={t.type} size="small">
              {t.text}
            </ElTag>
          );
        }
      },
      {
        prop: 'agentPort',
        label: '监听端口',
        width: 100,
        align: 'center',
        formatter: row => <span class="text-gray">{row.agentPort || '-'}</span>
      },
      {
        prop: 'lastHeartbeatAt',
        label: '最近心跳',
        width: 130,
        align: 'center',
        formatter: row => <span class="text-gray">{formatRelativeTime(row.lastHeartbeatAt)}</span>
      },
      {
        prop: 'operate',
        label: '操作',
        width: 160,
        align: 'center',
        className: 'msre-table-actions',
        fixed: 'right',
        formatter: row => {
          if (!row.agentStatus || row.agentStatus === 'uninstalled') {
            return (
              <PermissionButton
                code="cmdb.agents.deploy"
                link
                type="primary"
                size="small"
                onClick={() => handleDeploy(row)}
              >
                部署
              </PermissionButton>
            );
          }
          return (
            <div>
              {row.agentStatus === 'running' && !isLatestVersion(row.agentVersion) && (
                <PermissionButton
                  code="cmdb.agents.upgrade"
                  link
                  type="primary"
                  size="small"
                  onClick={() => handleUpgrade(row)}
                >
                  升级
                </PermissionButton>
              )}
              <ElDropdown trigger="click">
                {{
                  default: () => (
                    <ElButton link type="primary" size="small" class="table-dropdown-trigger">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </ElButton>
                  ),
                  dropdown: () => (
                    <ElDropdownMenu>
                      {withDirectives(
                        <ElDropdownItem onClick={() => handleRestart(row)}>
                          <ElIcon class="el-icon--left">
                            <icon-mdi-refresh />
                          </ElIcon>
                          重启
                        </ElDropdownItem>,
                        [[vPermission, 'cmdb.agents.restart']]
                      )}
                      {withDirectives(
                        <ElDropdownItem onClick={() => handleUninstall(row)}>
                          <ElIcon class="el-icon--left">
                            <icon-mdi-delete />
                          </ElIcon>
                          卸载
                        </ElDropdownItem>,
                        [[vPermission, 'cmdb.agents.uninstall']]
                      )}
                      {withDirectives(
                        <ElDropdownItem onClick={() => handleDeleteRecord(row)}>
                          <ElIcon class="el-icon--left">
                            <icon-mdi-trash-can />
                          </ElIcon>
                          删除记录
                        </ElDropdownItem>,
                        [[vPermission, 'cmdb.agents.delete']]
                      )}
                    </ElDropdownMenu>
                  )
                }}
              </ElDropdown>
            </div>
          );
        }
      }
    ]
  });

  // ========== 辅助函数 ==========
  function formatRelativeTime(time?: string) {
    if (!time) return '-';
    const diff = Date.now() - new Date(time).getTime();
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}小时前`;
    return `${Math.floor(hours / 24)}天前`;
  }

  function getAgentStatusTag(status?: string): {
    text: string;
    type: 'success' | 'danger' | 'info' | 'primary' | 'warning';
  } {
    if (status === 'running') return { text: '运行中', type: 'success' };
    if (status === 'offline') return { text: '离线', type: 'danger' };
    return { text: '未安装', type: 'info' };
  }

  function isLatestVersion(currentVersion?: string): boolean {
    if (!currentVersion || !latestVersion.value) return false;
    return currentVersion === latestVersion.value.version;
  }

  function getVersionStatus(currentVersion?: string): {
    text: string;
    type: 'success' | 'warning' | 'info' | 'primary' | 'danger';
  } {
    if (!currentVersion) return { text: '未安装', type: 'info' };
    if (isLatestVersion(currentVersion)) return { text: '最新版', type: 'success' };
    return { text: '可升级', type: 'warning' };
  }

  async function getLatestVersion() {
    try {
      const { data } = await fetchGetLatestAgentVersion();
      latestVersion.value = data;
    } catch {
      ElMessage.error('获取最新版本失败');
    }
  }

  // ========== 轮询 ==========
  function pollAgentStatus(serverId: number, expectedStatus: string, maxTimes = 20) {
    let count = 0;
    const timer = setInterval(async () => {
      count += 1;
      const { data: statusData } = await fetchGetAgentStatus(serverId).catch(() => ({ data: null }));
      if (statusData) {
        const idx = data.value.findIndex(r => r.id === serverId);
        if (idx !== -1) {
          data.value[idx] = {
            ...data.value[idx],
            agentStatus: statusData.agentStatus as CMDB.Server['agentStatus'],
            agentVersion: statusData.agentVersion,
            agentPort: statusData.agentPort,
            lastHeartbeatAt: statusData.lastHeartbeatAt
          };
        }
      }
      if ((statusData && statusData.agentStatus === expectedStatus) || count >= maxTimes) {
        clearInterval(timer);
        getData();
      }
    }, 3000);
  }

  // ========== 单条操作 ==========
  async function handleDeploy(row: CMDB.Server) {
    try {
      await fetchDeployAgent(row.id);
      ElMessage.info(`${row.hostname} Agent 部署任务已提交，正在轮询状态...`);
      pollAgentStatus(row.id, 'running');
    } catch {
      ElMessage.error(`${row.hostname} Agent 部署失败`);
    }
  }

  async function handleRestart(row: CMDB.Server) {
    try {
      await fetchRestartAgent(row.id);
      ElMessage.info(`${row.hostname} Agent 重启任务已提交...`);
      pollAgentStatus(row.id, 'running');
    } catch {
      ElMessage.error(`${row.hostname} Agent 重启失败`);
    }
  }

  async function handleUninstall(row: CMDB.Server) {
    try {
      await ElMessageBox.confirm(`确认卸载主机 "${row.hostname}" 上的 Agent？`, '卸载确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });
    } catch {
      return;
    }
    try {
      await fetchUninstallAgent(row.id);
      ElMessage.info(`${row.hostname} Agent 卸载任务已提交...`);
      pollAgentStatus(row.id, 'uninstalled');
    } catch {
      ElMessage.error(`${row.hostname} Agent 卸载失败`);
    }
  }

  async function handleDeleteRecord(row: CMDB.Server) {
    try {
      await ElMessageBox.confirm(
        `确认删除主机 "${row.hostname}" 的 Agent 记录？此操作仅清空数据库记录，不会 SSH 到主机执行任何操作。`,
        '删除确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      );
    } catch {
      return;
    }
    try {
      await fetchDeleteAgentRecord(row.id);
      ElMessage.success('Agent 记录已删除');
      getData();
    } catch {
      ElMessage.error('删除记录失败');
    }
  }

  async function handleUpgrade(row: CMDB.Server) {
    if (!latestVersion.value) {
      ElMessage.warning('无法获取最新版本信息');
      return;
    }
    if (row.agentVersion === latestVersion.value.version) {
      ElMessage.info('当前已是最新版本，无需升级');
      return;
    }
    try {
      await ElMessageBox.confirm(
        `确认将主机 "${row.hostname}" 的 Agent 从 v${row.agentVersion || '未知'} 升级到 v${latestVersion.value.version}？`,
        '升级确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
      );
    } catch {
      return;
    }
    try {
      await fetchUpgradeAgent(row.id, latestVersion.value.version);
      ElMessage.info(`${row.hostname} Agent 升级任务已提交，正在轮询状态...`);
      pollAgentStatus(row.id, 'running');
    } catch (err: unknown) {
      ElMessage.error(`${row.hostname} Agent 升级失败: ${err instanceof Error ? err.message : '未知错误'}`);
    }
  }

  // ========== 批量操作 ==========
  const batchDeployable = computed(() =>
    selectedRows.value.filter(r => !r.agentStatus || r.agentStatus === 'uninstalled')
  );
  const batchUninstallable = computed(() =>
    selectedRows.value.filter(r => r.agentStatus === 'running' || r.agentStatus === 'offline')
  );
  const batchUpgradeable = computed(() =>
    selectedRows.value.filter(r => {
      if (r.agentStatus !== 'running') return false;
      if (!r.agentVersion) return false;
      return !isLatestVersion(r.agentVersion);
    })
  );

  async function handleBatchDeploy() {
    const targets = batchDeployable.value;
    if (targets.length === 0) {
      ElMessage.warning('请先勾选状态为"未安装"的主机');
      return;
    }
    try {
      await ElMessageBox.confirm(`确认批量部署 ${targets.length} 台主机的 Agent？`, '批量部署确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      });
    } catch {
      return;
    }
    batchProgress.value = { current: 0, total: targets.length, running: true };
    for (const server of targets) {
      await fetchDeployAgent(server.id).catch(() => {});
      batchProgress.value.current++;
      pollAgentStatus(server.id, 'running');
    }
    batchProgress.value.running = false;
    ElMessage.success(`已提交 ${targets.length} 台主机的 Agent 部署任务`);
  }

  async function handleBatchUninstall() {
    const targets = batchUninstallable.value;
    if (targets.length === 0) {
      ElMessage.warning('请先勾选状态为"运行中"或"离线"的主机');
      return;
    }
    try {
      await ElMessageBox.confirm(`确认批量卸载 ${targets.length} 台主机的 Agent？`, '批量卸载确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });
    } catch {
      return;
    }
    try {
      await fetchBatchUninstallAgent(targets.map(r => r.id));
      ElMessage.info(`已提交 ${targets.length} 台主机的 Agent 卸载任务`);
      targets.forEach(s => pollAgentStatus(s.id, 'uninstalled'));
    } catch {
      ElMessage.error('批量卸载失败');
    }
  }

  async function handleBatchUpgrade() {
    const targets = batchUpgradeable.value;
    if (targets.length === 0) {
      ElMessage.warning('请先勾选状态为"运行中"且需要升级的主机');
      return;
    }
    if (!latestVersion.value) {
      ElMessage.warning('无法获取最新版本信息');
      return;
    }
    try {
      await ElMessageBox.confirm(
        `确认批量升级 ${targets.length} 台主机的 Agent 到 v${latestVersion.value.version}？`,
        '批量升级确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
      );
    } catch {
      return;
    }
    batchProgress.value = { current: 0, total: targets.length, running: true };
    for (const server of targets) {
      await fetchUpgradeAgent(server.id, latestVersion.value.version).catch(() => {});
      batchProgress.value.current++;
      pollAgentStatus(server.id, 'running');
    }
    batchProgress.value.running = false;
    ElMessage.success(`已提交 ${targets.length} 台主机的 Agent 升级任务`);
  }

  function handleSelectionChange(rows: CMDB.Server[]) {
    selectedRows.value = rows;
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
    getLatestVersion();
  });
</script>

<template>
  <div class="table-page">
    <ElCard shadow="never">
      <template #header>
        <div class="toolbar">
          <ElForm :model="searchParams" inline class="search-form">
            <ElFormItem>
              <ElInput
                v-model="searchParams.hostname"
                placeholder="主机名 / IP"
                clearable
                style="width: 200px"
                @keyup.enter="handleSearch"
              />
            </ElFormItem>
            <ElFormItem>
              <ElSelect v-model="searchParams.agentStatus" placeholder="全部状态" clearable style="width: 140px">
                <ElOption label="运行中" value="running" />
                <ElOption label="离线" value="offline" />
                <ElOption label="未安装" value="uninstalled" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" plain @click="handleSearch">
                <template #icon><icon-ic-round-search class="text-icon" /></template>
                搜索
              </ElButton>
              <ElButton plain @click="handleReset">
                <template #icon><icon-ic-round-refresh class="text-icon" /></template>
                重置
              </ElButton>
            </ElFormItem>
          </ElForm>

          <ElSpace>
            <span v-if="batchProgress.running" class="batch-progress-text">
              批量部署中：{{ batchProgress.current }} / {{ batchProgress.total }} 台
            </span>
            <PermissionButton
              code="cmdb.agents.deploy"
              type="primary"
              plain
              :disabled="batchDeployable.length === 0"
              @click="handleBatchDeploy"
            >
              <template #icon><icon-mdi-rocket-launch class="text-icon" /></template>
              批量部署
              <span v-if="batchDeployable.length > 0">（{{ batchDeployable.length }}）</span>
            </PermissionButton>
            <PermissionButton
              code="cmdb.agents.uninstall"
              type="danger"
              plain
              :disabled="batchUninstallable.length === 0"
              @click="handleBatchUninstall"
            >
              <template #icon><icon-mdi-delete class="text-icon" /></template>
              批量卸载
              <span v-if="batchUninstallable.length > 0">（{{ batchUninstallable.length }}）</span>
            </PermissionButton>
            <PermissionButton
              code="cmdb.agents.upgrade"
              type="success"
              plain
              :disabled="batchUpgradeable.length === 0"
              @click="handleBatchUpgrade"
            >
              <template #icon><icon-mdi-arrow-up-bold class="text-icon" /></template>
              批量升级
              <span v-if="batchUpgradeable.length > 0">（{{ batchUpgradeable.length }}）</span>
            </PermissionButton>
            <ElButton plain @click="getData">
              <template #icon>
                <icon-mdi-refresh class="text-icon" :class="{ 'animate-spin': loading }" />
              </template>
              刷新
            </ElButton>
          </ElSpace>
        </div>
      </template>

      <div class="table-scroll-wrap">
        <ElTable v-loading="loading" :data="data" border stripe height="100%" @selection-change="handleSelectionChange">
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>

      <div v-if="mobilePagination.total" class="pagination-bar">
        <ElPagination
          layout="total, sizes, prev, pager, next"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
    </ElCard>
  </div>
</template>

<style scoped>
  .agent-page {
    padding: 16px;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .search-form :deep(.el-form-item) {
    margin-bottom: 0;
    margin-right: 8px;
  }

  .text-gray {
    color: #909399;
    font-size: 13px;
  }

  .version-cell {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
  }

  .version-text {
    font-size: 12px;
    color: #67c23a;
    font-family: monospace;
  }

  .version-tag {
    font-size: 11px;
  }

  .batch-progress-text {
    font-size: 13px;
    color: #409eff;
    font-weight: 500;
  }

  .pagination-bar {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  .actions-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
  }

  .action-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    font-size: 12px;
    color: #409eff;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s;
    text-decoration: none;

    &:hover {
      background-color: #ecf5ff;
      color: #66b1ff;
    }
  }

  .upgrade-link {
    color: #67c23a;

    &:hover {
      background-color: #f0f9eb;
      color: #85ce61;
    }
  }

  .more-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s;

    &:hover {
      background-color: #f5f7fa;
    }
  }

  .more-icon {
    font-size: 16px;
    color: #909399;
    font-style: normal;
    letter-spacing: 1px;
  }
</style>
