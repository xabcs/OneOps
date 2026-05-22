<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessageBox, ElNotification } from 'element-plus';
import {
  fetchGetAgentList,
  fetchDeployAgent,
  fetchRestartAgent,
  fetchUninstallAgent,
  fetchGetAgentStatus,
  fetchBatchDeployAgent,
  fetchBatchUninstallAgent,
  fetchDeleteAgentRecord
} from '@/service/api/cmdb';

defineOptions({ name: 'CmdbConfigAgents' });

// ========== 状态 ==========
const loading = ref(false);
const tableData = ref<CMDB.Server[]>([]);
const total = ref(0);
const selectedRows = ref<CMDB.Server[]>([]);

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

function getAgentStatusTag(status?: string): { text: string; type: 'success' | 'danger' | 'info' } {
  if (status === 'running') return { text: '运行中', type: 'success' };
  if (status === 'offline') return { text: '离线', type: 'danger' };
  return { text: '未安装', type: 'info' };
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
    count++;
    const { data } = await fetchGetAgentStatus(serverId).catch(() => ({ data: null }));
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
    await ElMessageBox.confirm(
      `确认卸载主机 "${row.hostname}" 上的 Agent？`,
      '卸载确认',
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
    await ElMessageBox.confirm(
      `确认批量部署 ${targets.length} 台主机的 Agent？`,
      '批量部署确认',
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
    await ElMessageBox.confirm(
      `确认批量卸载 ${targets.length} 台主机的 Agent？`,
      '批量卸载确认',
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
    await fetchBatchUninstallAgent(targets.map(r => r.id));
    ElNotification.info(`已提交 ${targets.length} 台主机的 Agent 卸载任务`);
    targets.forEach(s => pollAgentStatus(s.id, 'uninstalled'));
  } catch {
    ElNotification.error('批量卸载失败');
  }
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
});
</script>

<template>
  <div class="agent-page">
    <el-card shadow="never">
      <!-- 工具栏 -->
      <template #header>
        <div class="toolbar">
          <!-- 左侧搜索区 -->
          <el-form :model="searchForm" inline class="search-form">
            <el-form-item>
              <el-input
                v-model="searchForm.hostname"
                placeholder="主机名 / IP"
                clearable
                style="width: 200px"
                @keyup.enter="handleSearch"
              />
            </el-form-item>
            <el-form-item>
              <el-select
                v-model="searchForm.agentStatus"
                placeholder="全部状态"
                clearable
                style="width: 140px"
              >
                <el-option label="运行中" value="running" />
                <el-option label="离线" value="offline" />
                <el-option label="未安装" value="uninstalled" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" plain @click="handleSearch">
                <template #icon><icon-ic-round-search class="text-icon" /></template>
                搜索
              </el-button>
              <el-button plain @click="handleReset">
                <template #icon><icon-ic-round-refresh class="text-icon" /></template>
                重置
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 右侧操作区 -->
          <el-space>
            <!-- 批量部署进度提示 -->
            <span v-if="batchProgress.running" class="batch-progress-text">
              批量部署中：{{ batchProgress.current }} / {{ batchProgress.total }} 台
            </span>

            <el-button
              type="primary"
              plain
              :disabled="batchDeployable.length === 0"
              @click="handleBatchDeploy"
            >
              <template #icon><icon-mdi-rocket-launch class="text-icon" /></template>
              批量部署
              <span v-if="batchDeployable.length > 0">（{{ batchDeployable.length }}）</span>
            </el-button>

            <el-button
              type="danger"
              plain
              :disabled="batchUninstallable.length === 0"
              @click="handleBatchUninstall"
            >
              <template #icon><icon-mdi-delete class="text-icon" /></template>
              批量卸载
              <span v-if="batchUninstallable.length > 0">（{{ batchUninstallable.length }}）</span>
            </el-button>

            <el-button plain @click="getAgentList">
              <template #icon>
                <icon-mdi-refresh class="text-icon" :class="{ 'animate-spin': loading }" />
              </template>
              刷新
            </el-button>
          </el-space>
        </div>
      </template>

      <!-- Agent 列表 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" align="center" />

        <el-table-column prop="hostname" label="主机名称" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="hostname-text">{{ row.hostname }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="ip" label="IP地址" width="140" show-overflow-tooltip />

        <el-table-column prop="innerIp" label="内网IP" width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="text-gray">{{ row.innerIp || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="agentVersion" label="版本" width="100" align="center">
          <template #default="{ row }">
            <span v-if="row.agentVersion" class="version-text">v{{ row.agentVersion }}</span>
            <span v-else class="text-gray">-</span>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="getAgentStatusTag(row.agentStatus).type"
              size="small"
            >
              {{ getAgentStatusTag(row.agentStatus).text }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="agentPort" label="监听端口" width="100" align="center">
          <template #default="{ row }">
            <span class="text-gray">{{ row.agentPort || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="最近心跳" width="130" align="center">
          <template #default="{ row }">
            <span class="text-gray">{{ formatRelativeTime(row.lastHeartbeatAt) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="230" align="center" fixed="right">
          <template #default="{ row }">
            <!-- 未安装：只显示部署 -->
            <template v-if="!row.agentStatus || row.agentStatus === 'uninstalled'">
              <el-button type="primary" size="small" @click="handleDeploy(row)">部署</el-button>
            </template>

            <!-- 运行中：重启 + 卸载 + 删除记录 -->
            <template v-else-if="row.agentStatus === 'running'">
              <el-button type="warning" size="small" @click="handleRestart(row)">重启</el-button>
              <el-button type="danger" size="small" plain @click="handleUninstall(row)">卸载</el-button>
              <el-button type="info" size="small" plain @click="handleDeleteRecord(row)">删除记录</el-button>
            </template>

            <!-- 离线：重启 + 卸载 + 删除记录 -->
            <template v-else-if="row.agentStatus === 'offline'">
              <el-button type="warning" size="small" @click="handleRestart(row)">重启</el-button>
              <el-button type="danger" size="small" plain @click="handleUninstall(row)">卸载</el-button>
              <el-button type="info" size="small" plain @click="handleDeleteRecord(row)">删除记录</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div v-if="total > 0" class="pagination-bar">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </el-card>
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
</style>
