<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import {
    fetchGetCommands,
    fetchGetServerStats,
    fetchGetServers,
    fetchGetSessionStats,
    fetchGetSessions
  } from '@/service/api/cmdb';

  defineOptions({ name: 'CmdbDashboard' });

  const router = useRouter();

  const serverStats = ref<CMDB.ServerStats | null>(null);
  const sessionStats = ref<{ active: number; today: number } | null>(null);
  const recentSessions = ref<Bastion.BastionSession[]>([]);
  const recentCommands = ref<Bastion.BastionCommand[]>([]);

  const envCounts = ref({ prod: 0, test: 0, dev: 0 });

  const loading = ref({
    serverStats: false,
    sessionStats: false,
    sessions: false,
    commands: false,
    env: false
  });

  async function loadServerStats() {
    loading.value.serverStats = true;
    try {
      const { data } = await fetchGetServerStats();
      serverStats.value = data || null;
    } catch (error) {
      console.error('获取服务器统计失败:', error);
    } finally {
      loading.value.serverStats = false;
    }
  }

  async function loadSessionStats() {
    loading.value.sessionStats = true;
    try {
      const { data } = await fetchGetSessionStats();
      sessionStats.value = data || null;
    } catch (error) {
      console.error('获取会话统计失败:', error);
    } finally {
      loading.value.sessionStats = false;
    }
  }

  async function loadEnvCounts() {
    loading.value.env = true;
    try {
      const [prodRes, testRes, devRes] = await Promise.all([
        fetchGetServers({ env: 'prod', pageSize: 1 } as CMDB.ServerQuery),
        fetchGetServers({ env: 'test', pageSize: 1 } as CMDB.ServerQuery),
        fetchGetServers({ env: 'dev', pageSize: 1 } as CMDB.ServerQuery)
      ]);
      envCounts.value = {
        prod: prodRes.data?.total || 0,
        test: testRes.data?.total || 0,
        dev: devRes.data?.total || 0
      };
    } catch (error) {
      console.error('获取环境分布失败:', error);
    } finally {
      loading.value.env = false;
    }
  }

  async function loadRecentSessions() {
    loading.value.sessions = true;
    try {
      const { data } = await fetchGetSessions({ page: 1, pageSize: 5 });
      recentSessions.value = data?.list || [];
    } catch (error) {
      console.error('获取最近会话失败:', error);
    } finally {
      loading.value.sessions = false;
    }
  }

  async function loadRecentCommands() {
    loading.value.commands = true;
    try {
      const { data } = await fetchGetCommands({ page: 1, pageSize: 5 });
      recentCommands.value = data?.list || [];
    } catch (error) {
      console.error('获取最近命令失败:', error);
    } finally {
      loading.value.commands = false;
    }
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
    loadServerStats();
    loadSessionStats();
    loadEnvCounts();
    loadRecentSessions();
    loadRecentCommands();
  });
</script>

<template>
  <div class="dashboard-page">
    <!-- 顶部统计卡片 -->
    <ElRow :gutter="16" class="stats-row">
      <ElCol :span="6">
        <ElCard shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon stat-icon--servers">
              <icon-mdi-server class="icon" />
            </div>
            <div class="stat-info">
              <div v-loading="loading.serverStats" class="stat-value">
                {{ serverStats?.total ?? '-' }}
              </div>
              <div class="stat-label">主机总数</div>
            </div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon stat-icon--online">
              <icon-mdi-server-network class="icon" />
            </div>
            <div class="stat-info">
              <div v-loading="loading.serverStats" class="stat-value">
                {{ serverStats?.online ?? '-' }}
              </div>
              <div class="stat-label">在线主机</div>
            </div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon stat-icon--today">
              <icon-mdi-calendar-today class="icon" />
            </div>
            <div class="stat-info">
              <div v-loading="loading.sessionStats" class="stat-value">
                {{ sessionStats?.today ?? '-' }}
              </div>
              <div class="stat-label">今日会话</div>
            </div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon stat-icon--active">
              <icon-mdi-connection class="icon" />
            </div>
            <div class="stat-info">
              <div v-loading="loading.sessionStats" class="stat-value">
                {{ sessionStats?.active ?? '-' }}
              </div>
              <div class="stat-label">活跃会话</div>
            </div>
          </div>
        </ElCard>
      </ElCol>
    </ElRow>

    <!-- 中部：环境分布 + 快速入口 -->
    <ElRow :gutter="16" class="middle-row">
      <ElCol :span="12">
        <ElCard shadow="never" header="按环境分布">
          <div v-loading="loading.env" class="env-list">
            <div class="env-item">
              <span class="env-dot env-dot--prod" />
              <span class="env-name">生产环境</span>
              <span class="env-count">{{ envCounts.prod }} 台</span>
            </div>
            <div class="env-item">
              <span class="env-dot env-dot--test" />
              <span class="env-name">测试环境</span>
              <span class="env-count">{{ envCounts.test }} 台</span>
            </div>
            <div class="env-item">
              <span class="env-dot env-dot--dev" />
              <span class="env-name">开发环境</span>
              <span class="env-count">{{ envCounts.dev }} 台</span>
            </div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="12">
        <ElCard shadow="never" header="快速入口">
          <div class="quick-links">
            <ElButton type="primary" plain size="large" @click="router.push('/cmdb/servers')">主机资产</ElButton>
            <ElButton type="success" plain size="large" @click="router.push('/cmdb/audit/online')">在线会话</ElButton>
            <ElButton type="warning" plain size="large" @click="router.push('/cmdb/audit/commands')">命令审计</ElButton>
            <ElButton type="info" plain size="large" @click="router.push('/cmdb/access/credentials')">凭证库</ElButton>
          </div>
        </ElCard>
      </ElCol>
    </ElRow>

    <!-- 底部：最近会话 + 最近命令 -->
    <ElRow :gutter="16" class="bottom-row">
      <ElCol :span="12">
        <ElCard shadow="never" header="最近会话">
          <div v-loading="loading.sessions">
            <div v-for="item in recentSessions" :key="item.id" class="list-item">
              <div class="list-item-main">
                <span class="list-item-user">{{ item.username }}</span>
                <span class="list-item-sep">→</span>
                <span class="list-item-server">{{ item.server?.hostname || `ID:${item.serverId}` }}</span>
              </div>
              <div class="list-item-meta">
                <span class="list-item-time">{{ formatTime(item.startedAt || '') }}</span>
                <ElTag :type="getStatusType(item.status)" size="small">{{ getStatusText(item.status) }}</ElTag>
              </div>
            </div>
            <ElEmpty v-if="!recentSessions.length && !loading.sessions" description="暂无数据" :image-size="60" />
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="12">
        <ElCard shadow="never" header="最近命令">
          <div v-loading="loading.commands">
            <div v-for="item in recentCommands" :key="item.id" class="list-item">
              <div class="list-item-main">
                <code class="list-item-command">{{ item.command }}</code>
              </div>
              <div class="list-item-meta">
                <span class="list-item-server">{{ item.session?.server?.hostname || '-' }}</span>
                <span class="list-item-time">{{ formatTime(item.executedAt || '') }}</span>
              </div>
            </div>
            <ElEmpty v-if="!recentCommands.length && !loading.commands" description="暂无数据" :image-size="60" />
          </div>
        </ElCard>
      </ElCol>
    </ElRow>
  </div>
</template>

<style scoped>
  .dashboard-page {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .stat-card {
    height: 100%;
  }

  .stat-content {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .stat-icon .icon {
    font-size: 28px;
    color: white;
  }

  .stat-icon--servers {
    background: linear-gradient(135deg, #667eea, #764ba2);
  }

  .stat-icon--online {
    background: linear-gradient(135deg, #11998e, #38ef7d);
  }

  .stat-icon--today {
    background: linear-gradient(135deg, #f7971e, #ffd200);
  }

  .stat-icon--active {
    background: linear-gradient(135deg, #f093fb, #f5576c);
  }

  .stat-info {
    flex: 1;
  }

  .stat-value {
    font-size: 28px;
    font-weight: bold;
    line-height: 1.2;
    color: #303133;
  }

  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-top: 4px;
  }

  .env-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .env-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: #f5f7fa;
    border-radius: 8px;
  }

  .env-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .env-dot--prod {
    background: #f56c6c;
  }

  .env-dot--test {
    background: #e6a23c;
  }

  .env-dot--dev {
    background: #67c23a;
  }

  .env-name {
    flex: 1;
    font-size: 14px;
    color: #303133;
  }

  .env-count {
    font-size: 16px;
    font-weight: 600;
    color: #409eff;
  }

  .quick-links {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .quick-links .el-button {
    width: 100%;
  }

  .list-item {
    padding: 10px 0;
    border-bottom: 1px solid #f0f0f0;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
  }

  .list-item:last-child {
    border-bottom: none;
  }

  .list-item-main {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    min-width: 0;
  }

  .list-item-user {
    font-weight: 500;
    color: #303133;
  }

  .list-item-sep {
    color: #c0c4cc;
    font-size: 12px;
  }

  .list-item-server {
    color: #606266;
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .list-item-command {
    font-family: 'Courier New', monospace;
    font-size: 13px;
    color: #409eff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 240px;
  }

  .list-item-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .list-item-time {
    font-size: 12px;
    color: #909399;
    white-space: nowrap;
  }
</style>
