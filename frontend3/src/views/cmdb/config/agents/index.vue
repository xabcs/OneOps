<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElDropdown, ElDropdownItem, ElDropdownMenu, ElIcon, ElMessageBox, ElNotification } from 'element-plus';
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

defineOptions({ name: 'CmdbConfigAgents' });

// ========== 状态 ==========
const loading = ref(false);
const tableData = ref<CMDB.Server[]>([]);
const total = ref(0);
const selectedRows = ref<CMDB.Server[]>([]);

// 版本相关状态
const latestVersion = ref<CMDB.AgentVersion | null>(null);

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20
});

// 搜索表单
const searchForm = reactive({
  hostname: '',
  ip: '',
  agentStatus: ''
});

// 批量进度
const batchProgress = ref({ current: 0, total: 0, running: false });

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
  type: 'success' | 'danger' | 'info';
} {
  if (status === 'running') return { text: '运行中', type: 'success' };
  if (status === 'offline') return { text: '离线', type: 'danger' };
  return { text: '未安装', type: 'info' };
}

// ========== 版本相关函数 ==========
/**
 * 判断是否为最新版本
 */
function isLatestVersion(currentVersion?: string): boolean {
  if (!currentVersion || !latestVersion.value) return false;
  return currentVersion === latestVersion.value.version;
}

/**
 * 获取版本状态标签
 */
function getVersionStatus(currentVersion?: string): {
  text: string;
  type: 'success' | 'warning' | 'info';
} {
  if (!currentVersion) return { text: '未安装', type: 'info' };
  if (isLatestVersion(currentVersion)) return { text: '最新版', type: 'success' };
  return { text: '可升级', type: 'warning' };
}

/**
 * 获取版本列表
 */
async function getLatestVersion() {
  try {
    const { data } = await fetchGetLatestAgentVersion();
    latestVersion.value = data;
  } catch (err) {
    ElNotification.error('获取最新版本失败');
  }
}

// ========== 数据获取 ==========
async function getAgentList() {
  loading.value = true;
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      pageSize: pagination.pageSize
    };
    if (searchForm.hostname) params.hostname = searchForm.hostname;
    if (searchForm.ip) params.ip = searchForm.ip;
    if (searchForm.agentStatus) params.agentStatus = searchForm.agentStatus;

    const { data } = await fetchGetAgentList(params);
    tableData.value = data?.list || [];
    total.value = data?.total || 0;
  } catch {
    ElNotification.error('获取 Agent 列表失败');
  } finally {
    loading.value = false;
  }
}

// ========== 轮询函数 ==========
function pollAgentStatus(serverId: number, expectedStatus: string, maxTimes = 20) {
  let count = 0;
  const timer = setInterval(async () => {
    count += 1;
    const { data } = await fetchGetAgentStatus(serverId).catch(() => ({
      data: null
    }));
    if (data) {
      const idx = tableData.value.findIndex(r => r.id === serverId);
      if (idx !== -1) {
        tableData.value[idx] = {
          ...tableData.value[idx],
          agentStatus: data.agentStatus as CMDB.Server['agentStatus'],
          agentVersion: data.agentVersion,
          agentPort: data.agentPort,
          lastHeartbeatAt: data.lastHeartbeatAt
        };
      }
    }
    if ((data && data.agentStatus === expectedStatus) || count >= maxTimes) {
      clearInterval(timer);
      getAgentList();
    }
  }, 3000);
}

// ========== 单条操作 ==========
async function handleDeploy(row: CMDB.Server) {
  try {
    await fetchDeployAgent(row.id);
    ElNotification.info(`${row.hostname} Agent 部署任务已提交，正在轮询状态...`);
    pollAgentStatus(row.id, 'running');
  } catch {
    ElNotification.error(`${row.hostname} Agent 部署失败`);
  }
}

async function handleRestart(row: CMDB.Server) {
  try {
    await fetchRestartAgent(row.id);
    ElNotification.info(`${row.hostname} Agent 重启任务已提交...`);
    pollAgentStatus(row.id, 'running');
  } catch {
    ElNotification.error(`${row.hostname} Agent 重启失败`);
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
    ElNotification.info(`${row.hostname} Agent 卸载任务已提交...`);
    pollAgentStatus(row.id, 'uninstalled');
  } catch {
    ElNotification.error(`${row.hostname} Agent 卸载失败`);
  }
}

async function handleDeleteRecord(row: CMDB.Server) {
  try {
    await ElMessageBox.confirm(
      `确认删除主机 "${row.hostname}" 的 Agent 记录？此操作仅清空数据库记录，不会 SSH 到主机执行任何操作。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );
  } catch {
    return;
  }
  try {
    await fetchDeleteAgentRecord(row.id);
    ElNotification.success('Agent 记录已删除');
    getAgentList();
  } catch {
    ElNotification.error('删除记录失败');
  }
}

async function handleUpgrade(row: CMDB.Server) {
  if (!latestVersion.value) {
    ElNotification.warning('无法获取最新版本信息');
    return;
  }

  if (row.agentVersion === latestVersion.value.version) {
    ElNotification.info('当前已是最新版本，无需升级');
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确认将主机 "${row.hostname}" 的 Agent 从 v${
        row.agentVersion || '未知'
      } 升级到 v${latestVersion.value.version}？`,
      '升级确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    );
  } catch {
    return;
  }

  try {
    await fetchUpgradeAgent(row.id, latestVersion.value.version);
    ElNotification.info(`${row.hostname} Agent 升级任务已提交，正在轮询状态...`);
    pollAgentStatus(row.id, 'running');
  } catch (err: unknown) {
    ElNotification.error(`${row.hostname} Agent 升级失败: ${err instanceof Error ? err.message : '未知错误'}`);
  }
}

// ========== 批量操作 ==========
const batchDeployable = computed(() =>
  selectedRows.value.filter(r => !r.agentStatus || r.agentStatus === 'uninstalled')
);

const batchUninstallable = computed(() =>
  selectedRows.value.filter(r => r.agentStatus === 'running' || r.agentStatus === 'offline')
);

async function handleBatchDeploy() {
  const targets = batchDeployable.value;
  if (targets.length === 0) {
    ElNotification.warning('请先勾选状态为"未安装"的主机');
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
  ElNotification.success(`已提交 ${targets.length} 台主机的 Agent 部署任务`);
}

async function handleBatchUninstall() {
  const targets = batchUninstallable.value;
  if (targets.length === 0) {
    ElNotification.warning('请先勾选状态为"运行中"或"离线"的主机');
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
    ElNotification.info(`已提交 ${targets.length} 台主机的 Agent 卸载任务`);
    targets.forEach(s => pollAgentStatus(s.id, 'uninstalled'));
  } catch {
    ElNotification.error('批量卸载失败');
  }
}

// 可升级的主机（运行中且不是最新版本）
const batchUpgradeable = computed(() =>
  selectedRows.value.filter(r => {
    if (r.agentStatus !== 'running') return false;
    if (!r.agentVersion) return false;
    return !isLatestVersion(r.agentVersion);
  })
);

async function handleBatchUpgrade() {
  const targets = batchUpgradeable.value;
  if (targets.length === 0) {
    ElNotification.warning('请先勾选状态为"运行中"且需要升级的主机');
    return;
  }

  if (!latestVersion.value) {
    ElNotification.warning('无法获取最新版本信息');
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确认批量升级 ${targets.length} 台主机的 Agent 到 v${latestVersion.value.version}？`,
      '批量升级确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
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
  ElNotification.success(`已提交 ${targets.length} 台主机的 Agent 升级任务`);
}

// ========== 表格事件 ==========
function handleSelectionChange(rows: CMDB.Server[]) {
  selectedRows.value = rows;
}

// ========== 搜索 & 分页 ==========
function handleSearch() {
  pagination.page = 1;
  getAgentList();
}

function handleReset() {
  searchForm.hostname = '';
  searchForm.ip = '';
  searchForm.agentStatus = '';
  pagination.page = 1;
  getAgentList();
}

function handlePageChange(page: number) {
  pagination.page = page;
  getAgentList();
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  getAgentList();
}

// ========== 初始化 ==========
onMounted(() => {
  getAgentList();
  getLatestVersion();
});
</script>

<template>
  <div class="agent-page">
    <ElCard shadow="never">
      <!-- 工具栏 -->
      <template #header>
        <div class="toolbar">
          <!-- 左侧搜索区 -->
          <ElForm :model="searchForm" inline class="search-form">
            <ElFormItem>
              <ElInput
                v-model="searchForm.hostname"
                placeholder="主机名 / IP"
                clearable
                style="width: 200px"
                @keyup.enter="handleSearch"
              />
            </ElFormItem>
            <ElFormItem>
              <ElSelect v-model="searchForm.agentStatus" placeholder="全部状态" clearable style="width: 140px">
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

          <!-- 右侧操作区 -->
          <ElSpace>
            <!-- 批量部署进度提示 -->
            <span v-if="batchProgress.running" class="batch-progress-text">
              批量部署中：{{ batchProgress.current }} / {{ batchProgress.total }} 台
            </span>

            <ElButton type="primary" plain :disabled="batchDeployable.length === 0" @click="handleBatchDeploy">
              <template #icon><icon-mdi-rocket-launch class="text-icon" /></template>
              批量部署
              <span v-if="batchDeployable.length > 0">（{{ batchDeployable.length }}）</span>
            </ElButton>

            <ElButton type="danger" plain :disabled="batchUninstallable.length === 0" @click="handleBatchUninstall">
              <template #icon><icon-mdi-delete class="text-icon" /></template>
              批量卸载
              <span v-if="batchUninstallable.length > 0">（{{ batchUninstallable.length }}）</span>
            </ElButton>

            <ElButton type="success" plain :disabled="batchUpgradeable.length === 0" @click="handleBatchUpgrade">
              <template #icon><icon-mdi-arrow-up-bold class="text-icon" /></template>
              批量升级
              <span v-if="batchUpgradeable.length > 0">（{{ batchUpgradeable.length }}）</span>
            </ElButton>

            <ElButton plain @click="getAgentList">
              <template #icon>
                <icon-mdi-refresh class="text-icon" :class="{ 'animate-spin': loading }" />
              </template>
              刷新
            </ElButton>
          </ElSpace>
        </div>
      </template>

      <!-- Agent 列表 -->
      <ElTable
        v-loading="loading"
        :data="tableData"
        border
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <ElTableColumn type="selection" width="50" align="center" />

        <ElTableColumn prop="hostname" label="主机名称" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="hostname-text">{{ row.hostname }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn prop="ip" label="IP地址" width="140" show-overflow-tooltip />

        <ElTableColumn prop="innerIp" label="内网IP" width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="text-gray">{{ row.innerIp || '-' }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="版本" width="140" align="center">
          <template #default="{ row }">
            <div v-if="row.agentVersion" class="version-cell">
              <span class="version-text">v{{ row.agentVersion }}</span>
              <ElTag
                v-if="latestVersion && row.agentStatus === 'running'"
                :type="getVersionStatus(row.agentVersion).type"
                size="small"
                class="version-tag"
              >
                {{ getVersionStatus(row.agentVersion).text }}
              </ElTag>
            </div>
            <span v-else class="text-gray">-</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="状态" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="getAgentStatusTag(row.agentStatus).type" size="small">
              {{ getAgentStatusTag(row.agentStatus).text }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn prop="agentPort" label="监听端口" width="100" align="center">
          <template #default="{ row }">
            <span class="text-gray">{{ row.agentPort || '-' }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="最近心跳" width="130" align="center">
          <template #default="{ row }">
            <span class="text-gray">{{ formatRelativeTime(row.lastHeartbeatAt) }}</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <!-- 未安装：直接显示部署按钮 -->
            <template v-if="!row.agentStatus || row.agentStatus === 'uninstalled'">
              <ElButton type="primary" size="small" @click="handleDeploy(row)">部署</ElButton>
            </template>

            <!-- 运行中或离线：显示升级按钮和更多菜单 -->
            <template v-else>
              <span class="actions-wrapper">
                <!-- 运行中且非最新版本：显示升级按钮 -->
                <a
                  v-if="row.agentStatus === 'running' && !isLatestVersion(row.agentVersion)"
                  class="action-link upgrade-link"
                  @click="handleUpgrade(row)"
                >
                  <icon-mdi-arrow-up-bold />
                  <span>升级</span>
                </a>

                <!-- 更多菜单 -->
                <ElDropdown trigger="click">
                  <span class="more-btn">
                    <i class="more-icon">⋮</i>
                  </span>
                  <template #dropdown>
                    <ElDropdownMenu>
                      <!-- 重启 -->
                      <ElDropdownItem @click="handleRestart(row)">
                        <ElIcon class="el-icon--left"><icon-mdi-refresh /></ElIcon>
                        重启
                      </ElDropdownItem>
                      <!-- 卸载 -->
                      <ElDropdownItem @click="handleUninstall(row)">
                        <ElIcon class="el-icon--left"><icon-mdi-delete /></ElIcon>
                        卸载
                      </ElDropdownItem>
                      <!-- 删除记录 -->
                      <ElDropdownItem @click="handleDeleteRecord(row)">
                        <ElIcon class="el-icon--left"><icon-mdi-trash-can /></ElIcon>
                        删除记录
                      </ElDropdownItem>
                    </ElDropdownMenu>
                  </template>
                </ElDropdown>
              </span>
            </template>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div v-if="total > 0" class="pagination-bar">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
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

.hostname-text {
  font-weight: 500;
  color: #303133;
}

.version-text {
  font-size: 12px;
  color: #67c23a;
  font-family: monospace;
}

.version-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.version-tag {
  font-size: 11px;
}

.text-gray {
  color: #909399;
  font-size: 13px;
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
