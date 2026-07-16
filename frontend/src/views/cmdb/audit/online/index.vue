<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import { fetchGetActiveSessions, fetchGetSessionStats, fetchTerminateSession } from '@/service/api/cmdb';

defineOptions({ name: 'CmdbAuditOnline' });

const loading = ref(false);
const sessions = ref<Bastion.BastionSession[]>([]);
const stats = ref({ active: 0, today: 0 });

let refreshTimer: ReturnType<typeof setInterval> | null = null;

async function getActiveSessions() {
  loading.value = true;
  try {
    const { data } = await fetchGetActiveSessions();
    sessions.value = data || [];
  } catch (error) {
    window.$message?.error('获取活跃会话失败');
  } finally {
    loading.value = false;
  }
}

async function getStats() {
  try {
    const { data } = await fetchGetSessionStats();
    if (data) stats.value = data;
  } catch (error) {
    console.error('获取统计失败:', error);
  }
}

async function handleTerminate(session: Bastion.BastionSession) {
  try {
    await ElMessageBox.confirm(
      `确定要断开会话吗？\n用户: ${session.username}\n服务器: ${session.server?.hostname || session.serverId}`,
      '确认断开',
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
    await fetchTerminateSession(session.id);
    window.$message?.success('会话已断开');
    refresh();
  } catch (error: any) {
    // 处理 401/404 等错误，说明会话已不存在
    if (error?.response?.status === 401 || error?.response?.status === 404) {
      window.$message?.warning('会话已断开或不存在，已从列表移除');
      // 从列表中移除该会话
      sessions.value = sessions.value.filter(s => s.id !== session.id);
      refresh(); // 刷新统计
    } else {
      window.$message?.error(error.message || '断开会话失败');
    }
  }
}

// 批量清理无效会话
async function handleCleanupInvalidSessions() {
  try {
    await ElMessageBox.confirm(
      '这将尝试断开所有显示的在线会话，已断开的会话将自动从列表移除。是否继续？',
      '清理无效会话',
      {
        confirmButtonText: '开始清理',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );
  } catch {
    return;
  }

  loading.value = true;
  let successCount = 0;
  let invalidCount = 0;

  // 并发尝试断开所有会话
  const promises = sessions.value.map(async session => {
    try {
      await fetchTerminateSession(session.id);
      successCount++;
      return { id: session.id, valid: true };
    } catch (error: any) {
      // 401/404 表示会话已不存在（无效）
      if (error?.response?.status === 401 || error?.response?.status === 404) {
        invalidCount++;
        return { id: session.id, valid: false };
      }
      throw error;
    }
  });

  const results = await Promise.allSettled(promises);

  // 从列表中移除无效会话
  sessions.value = sessions.value.filter(session => {
    const result = results.find((r, i) => i === sessions.value.indexOf(session));
    return result?.status === 'fulfilled' && (result.value as any).value.valid;
  });

  loading.value = false;
  refresh(); // 刷新统计

  window.$message?.success(`清理完成！成功断开 ${successCount} 个会话，移除 ${invalidCount} 个无效会话`);
}

function refresh() {
  getActiveSessions();
  getStats();
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}小时${minutes}分钟`;
}

function formatTime(time: string): string {
  return time ? new Date(time).toLocaleString('zh-CN') : '-';
}

onMounted(() => {
  refresh();
  refreshTimer = setInterval(refresh, 30000);
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
});
</script>

<template>
  <div class="online-page">
    <ElCard shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">在线会话</span>
          <div class="header-actions">
            <ElButton type="warning" @click="handleCleanupInvalidSessions">清理无效会话</ElButton>
            <ElButton type="primary" @click="refresh">刷新</ElButton>
          </div>
        </div>
      </template>

      <ElAlert type="info" :closable="false" style="margin-bottom: 16px">
        <template #default>
          在线会话列表可能包含已失效的会话（服务器重启、网络中断等原因）。
          使用「清理无效会话」功能可自动移除已断开的会话。
        </template>
      </ElAlert>

      <div class="stats-cards">
        <div class="stat-card stat-active">
          <div class="stat-value">{{ stats.active }}</div>
          <div class="stat-label">当前活跃会话</div>
        </div>
        <div class="stat-card stat-today">
          <div class="stat-value">{{ stats.today }}</div>
          <div class="stat-label">今日会话总数</div>
        </div>
      </div>

      <ElTable v-loading="loading" :data="sessions" stripe style="width: 100%; margin-top: 16px">
        <ElTableColumn prop="id" label="会话ID" width="80" />
        <ElTableColumn prop="username" label="用户名" width="120" />
        <ElTableColumn label="服务器" width="160">
          <template #default="{ row }">
            {{ row.server?.hostname || `ID:${row.serverId}` }}
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
            {{ formatDuration(row.duration || 0) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag type="success" size="small">活跃</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <ElButton type="danger" size="small" @click="handleTerminate(row)">强制断开</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>

<style scoped>
.online-page {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
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
  border-radius: 8px;
  padding: 20px;
  color: white;
  text-align: center;
}

.stat-active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-today {
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
</style>
