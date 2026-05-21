<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { fetchGetSessions, fetchTerminateSession, fetchGetActiveSessions, fetchGetSessionStats } from '@/service/api/cmdb';
import { $t } from '@/locales';

defineOptions({
  name: 'CMDBSessions'
});

const activeTab = ref<'active' | 'history'>('active');
const loading = ref(false);
const sessions = ref<Bastion.BastionSession[]>([]);
const total = ref(0);
const stats = ref<{
  active: number;
  today: number;
}>({ active: 0, today: 0 });

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20
});

// 筛选条件
const filters = ref<{
  status?: string;
  protocol?: string;
  startDate?: string;
  endDate?: string;
}>({});

// 表格列定义
const columns = computed(() => [
  { prop: 'id', label: '会话ID', width: 80 },
  { prop: 'username', label: '用户', width: 100 },
  { prop: 'server', label: '服务器', width: 150 },
  { prop: 'serverIp', label: 'IP地址', width: 120 },
  { prop: 'loginAccount', label: '登录账号', width: 100 },
  { prop: 'clientIp', label: '客户端IP', width: 120 },
  { prop: 'protocol', label: '协议', width: 80 },
  { prop: 'startedAt', label: '开始时间', width: 160 },
  { prop: 'duration', label: '持续时长', width: 100 },
  { prop: 'status', label: '状态', width: 100 },
  { prop: 'actions', label: '操作', width: 150, fixed: 'right' }
]);

// 获取会话列表
async function getSessions() {
  loading.value = true;
  try {
    const params = {
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      ...filters.value
    };

    if (activeTab.value === 'active') {
      params.status = 'active';
    }

    const result = await fetchGetSessions(params);
    sessions.value = result.list;
    total.value = result.total;
  } catch (error) {
    window.$message?.error('获取会话列表失败');
  } finally {
    loading.value = false;
  }
}

// 获取活跃会话
async function getActiveSessions() {
  loading.value = true;
  try {
    sessions.value = await fetchGetActiveSessions();
    total.value = sessions.value.length;
  } catch (error) {
    window.$message?.error('获取活跃会话失败');
  } finally {
    loading.value = false;
  }
}

// 获取统计数据
async function getStats() {
  try {
    stats.value = await fetchGetSessionStats();
  } catch (error) {
    console.error('获取统计失败:', error);
  }
}

// 终止会话
async function handleTerminate(session: Bastion.BastionSession) {
  const confirmed = await window.$confirm?.(
    `确定要断开会话吗？\n用户: ${session.username}\n服务器: ${session.server?.hostname || session.serverId}`,
    '确认断开',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  );

  if (!confirmed) return;

  try {
    await fetchTerminateSession(session.id);
    window.$message?.success('会话已断开');
    refresh();
  } catch (error: any) {
    window.$message?.error(error.message || '断开会话失败');
  }
}

// 查看会话详情
function handleViewDetail(session: Bastion.BastionSession) {
  // TODO: 打开会话详情抽屉
  window.$message?.info('会话详情功能开发中');
}

// 刷新
function refresh() {
  if (activeTab.value === 'active') {
    getActiveSessions();
  } else {
    getSessions();
  }
  getStats();
}

// 切换标签页
function handleTabChange() {
  pagination.value.page = 1;
  refresh();
}

// 分页变化
function handlePageChange(page: number) {
  pagination.value.page = page;
  if (activeTab.value === 'history') {
    getSessions();
  }
}

// 格式化持续时长
function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}小时${minutes}分钟`;
}

// 格式化时间
function formatTime(time: string): string {
  return time ? new Date(time).toLocaleString('zh-CN') : '-';
}

// 获取状态标签类型
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

// 获取状态文本
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
  refresh();
  getStats();
});
</script>

<template>
  <div class="sessions-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">会话审计</span>
          <el-button type="primary" :icon="ICON_REGISTRY.refresh" @click="refresh">
            刷新
          </el-button>
        </div>
      </template>

      <!-- 统计卡片 -->
      <div class="stats-cards">
        <div class="stat-card">
          <div class="stat-value">{{ stats.active }}</div>
          <div class="stat-label">活跃会话</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats.today }}</div>
          <div class="stat-label">今日会话</div>
        </div>
      </div>

      <!-- 标签页 -->
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="活跃会话" name="active">
          <el-table
            v-loading="loading"
            :data="sessions"
            stripe
            style="width: 100%; margin-top: 16px"
          >
            <el-table-column prop="id" label="会话ID" width="80" />
            <el-table-column prop="username" label="用户" width="100" />
            <el-table-column label="服务器" width="150">
              <template #default="{ row }">
                {{ row.server?.hostname || `ID:${row.serverId}` }}
              </template>
            </el-table-column>
            <el-table-column label="IP地址" width="120">
              <template #default="{ row }">
                {{ row.server?.ip || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="loginAccount" label="登录账号" width="100" />
            <el-table-column prop="clientIp" label="客户端IP" width="120" />
            <el-table-column prop="protocol" label="协议" width="80">
              <template #default="{ row }">
                <el-tag :type="row.protocol === 'ssh' ? 'primary' : 'success'" size="small">
                  {{ row.protocol.toUpperCase() }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="开始时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.startedAt || '') }}
              </template>
            </el-table-column>
            <el-table-column label="持续时长" width="100">
              <template #default="{ row }">
                {{ formatDuration(row.duration) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="danger"
                  size="small"
                  :icon="ICON_REGISTRY.videoStop"
                  @click="handleTerminate(row)"
                >
                  断开
                </el-button>
                <el-button
                  type="primary"
                  size="small"
                  :icon="ICON_REGISTRY.fileDocument"
                  @click="handleViewDetail(row)"
                >
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="历史会话" name="history">
          <el-table
            v-loading="loading"
            :data="sessions"
            stripe
            style="width: 100%; margin-top: 16px"
          >
            <el-table-column prop="id" label="会话ID" width="80" />
            <el-table-column prop="username" label="用户" width="100" />
            <el-table-column label="服务器" width="150">
              <template #default="{ row }">
                {{ row.server?.hostname || `ID:${row.serverId}` }}
              </template>
            </el-table-column>
            <el-table-column label="IP地址" width="120">
              <template #default="{ row }">
                {{ row.server?.ip || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="loginAccount" label="登录账号" width="100" />
            <el-table-column prop="clientIp" label="客户端IP" width="120" />
            <el-table-column prop="protocol" label="协议" width="80">
              <template #default="{ row }">
                <el-tag :type="row.protocol === 'ssh' ? 'primary' : 'success'" size="small">
                  {{ row.protocol.toUpperCase() }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="开始时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.startedAt || '') }}
              </template>
            </el-table-column>
            <el-table-column label="持续时长" width="100">
              <template #default="{ row }">
                {{ formatDuration(row.duration) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="关闭原因" width="150">
              <template #default="{ row }">
                {{ row.closeReason || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="primary"
                  size="small"
                  :icon="ICON_REGISTRY.fileDocument"
                  @click="handleViewDetail(row)"
                >
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="pagination.page"
              v-model:page-size="pagination.pageSize"
              :total="total"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="refresh"
              @current-change="handlePageChange"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
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

.stats-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.stat-card {
  flex: 1;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  padding: 20px;
  color: white;
  text-align: center;
}

.stat-card:nth-child(2) {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
