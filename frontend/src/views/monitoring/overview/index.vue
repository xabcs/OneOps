<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { fetchMonitoringOverview } from '@/service/api/monitoring';
import { useWebSocket } from '@/service/websocket';
import type { MonitoringOverview } from '@/service/api/monitoring';

defineOptions({
  name: 'MonitoringOverview'
});

const loading = ref(false);
const overview = ref<MonitoringOverview | null>(null);

// WebSocket 客户端
const { subscribe, unsubscribe, client: wsClient } = useWebSocket();

async function getOverview() {
  loading.value = true;
  try {
    const { data } = await fetchMonitoringOverview();
    if (data) {
      overview.value = data;
    }
  } catch (error) {
    console.error('获取监控概览失败:', error);
  }
  loading.value = false;
}

function handleRefresh() {
  getOverview();
}

// WebSocket 实时更新处理
function handleOverviewUpdate(data: any) {
  if (data && overview.value) {
    overview.value = { ...overview.value, ...data };
  }
}

function handleAlertUpdate(alert: any) {
  if (overview.value && alert) {
    const exists = overview.value.activeAlerts?.find(a => a.id === alert.id);
    if (!exists) {
      overview.value.activeAlerts = [alert, ...(overview.value.activeAlerts || [])].slice(0, 10);
    }
  }
}

function formatTime(time: string): string {
  return new Date(time).toLocaleString('zh-CN');
}

function getAlertTagType(level: string): 'danger' | 'warning' | 'primary' | 'info' | 'success' {
  const map: Record<string, 'danger' | 'warning' | 'primary' | 'info' | 'success'> = {
    critical: 'danger',
    high: 'warning',
    medium: 'primary',
    low: 'info',
    info: 'info'
  };
  return map[level] ?? 'info';
}

function getAlertLevelText(level: string): string {
  const textMap: Record<string, string> = {
    critical: '严重',
    high: '高',
    medium: '中',
    low: '低',
    info: '信息'
  };
  return textMap[level] || level;
}

function getProgressColor(value: number): string {
  if (value > 80) return '#ff4d4f';
  if (value > 60) return '#faad14';
  return '#52c41a';
}

onMounted(() => {
  // 初始加载
  getOverview();

  // 订阅 WebSocket 实时更新
  subscribe('overview_update', handleOverviewUpdate);
  subscribe('alert', handleAlertUpdate);
});

onUnmounted(() => {
  // 取消订阅
  unsubscribe('overview_update', handleOverviewUpdate);
  unsubscribe('alert', handleAlertUpdate);
});
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- 标题栏 -->
    <ElCard shadow="never">
      <div class="flex items-center justify-between">
        <span class="text-lg font-semibold">监控概览</span>
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-400">
            <span
              class="mr-1 inline-block h-2 w-2 rounded-full"
              :class="wsClient?.isConnected ? 'bg-green-500' : 'bg-red-500'"
            ></span>
            {{ wsClient?.isConnected ? '已连接' : '未连接' }}
          </span>
          <ElButton type="primary" :loading="loading" @click="handleRefresh">
            <ElIcon :size="16"><Refresh /></ElIcon>
            刷新
          </ElButton>
        </div>
      </div>
    </ElCard>

    <ElSkeleton :loading="loading && !overview" animated :rows="10">
      <template #default>
        <div v-if="overview" class="space-y-4">
          <!-- 统计卡片 -->
          <ElRow :gutter="16">
            <ElCol :xs="24" :sm="12" :md="6">
              <ElCard shadow="hover" class="stat-card">
                <div class="stat-content">
                  <div class="stat-value" style="color: rgb(0, 82, 217)">{{ overview.summary?.totalServers || 0 }}</div>
                  <div class="stat-label">总主机数</div>
                </div>
              </ElCard>
            </ElCol>
            <ElCol :xs="24" :sm="12" :md="6">
              <ElCard shadow="hover" class="stat-card">
                <div class="stat-content">
                  <div class="stat-value" style="color: rgb(38, 187, 23)">
                    {{ overview.summary?.onlineServers || 0 }}
                  </div>
                  <div class="stat-label">在线主机</div>
                </div>
              </ElCard>
            </ElCol>
            <ElCol :xs="24" :sm="12" :md="6">
              <ElCard shadow="hover" class="stat-card">
                <div class="stat-content">
                  <div class="stat-value" style="color: rgb(245, 34, 46)">
                    {{ overview.summary?.offlineServers || 0 }}
                  </div>
                  <div class="stat-label">离线主机</div>
                </div>
              </ElCard>
            </ElCol>
            <ElCol :xs="24" :sm="12" :md="6">
              <ElCard shadow="hover" class="stat-card">
                <div class="stat-content">
                  <div class="stat-value" style="color: rgb(255, 168, 0)">
                    {{ overview.summary?.alertServers || 0 }}
                  </div>
                  <div class="stat-label">告警主机</div>
                </div>
              </ElCard>
            </ElCol>
          </ElRow>

          <!-- Top 10 资源使用 -->
          <ElRow :gutter="16">
            <ElCol :xs="24" :lg="8">
              <ElCard shadow="never" header="Top 10 CPU 使用率">
                <div class="top-list">
                  <div
                    v-for="(item, index) in (overview.topCpu || []).slice(0, 10)"
                    :key="item.serverId"
                    class="top-list-item"
                  >
                    <div class="top-list-rank" :class="`rank-${index + 1}`">{{ index + 1 }}</div>
                    <div class="top-list-content">
                      <div class="top-list-title">{{ item.hostname }}</div>
                      <div class="top-list-subtitle">{{ item.ip }}</div>
                    </div>
                    <div class="top-list-value">
                      <ElProgress
                        :percentage="Math.round(item.cpuUsage || 0)"
                        :color="getProgressColor(item.cpuUsage || 0)"
                        :stroke-width="8"
                        :show-text="true"
                        :format="() => (item.cpuUsage || 0).toFixed(1) + '%'"
                      />
                    </div>
                  </div>
                </div>
              </ElCard>
            </ElCol>

            <ElCol :xs="24" :lg="8">
              <ElCard shadow="never" header="Top 10 内存使用率">
                <div class="top-list">
                  <div
                    v-for="(item, index) in (overview.topMemory || []).slice(0, 10)"
                    :key="item.serverId"
                    class="top-list-item"
                  >
                    <div class="top-list-rank" :class="`rank-${index + 1}`">{{ index + 1 }}</div>
                    <div class="top-list-content">
                      <div class="top-list-title">{{ item.hostname }}</div>
                      <div class="top-list-subtitle">{{ item.ip }}</div>
                    </div>
                    <div class="top-list-value">
                      <ElProgress
                        :percentage="Math.round(item.memoryUsage || 0)"
                        :color="getProgressColor(item.memoryUsage || 0)"
                        :stroke-width="8"
                        :show-text="true"
                        :format="() => (item.memoryUsage || 0).toFixed(1) + '%'"
                      />
                    </div>
                  </div>
                </div>
              </ElCard>
            </ElCol>

            <ElCol :xs="24" :lg="8">
              <ElCard shadow="never" header="Top 10 磁盘使用率">
                <div class="top-list">
                  <div
                    v-for="(item, index) in (overview.topDisk || []).slice(0, 10)"
                    :key="item.serverId"
                    class="top-list-item"
                  >
                    <div class="top-list-rank" :class="`rank-${index + 1}`">{{ index + 1 }}</div>
                    <div class="top-list-content">
                      <div class="top-list-title">{{ item.hostname }}</div>
                      <div class="top-list-subtitle">{{ item.ip }}</div>
                    </div>
                    <div class="top-list-value">
                      <ElProgress
                        :percentage="Math.round(item.diskUsage || 0)"
                        :color="getProgressColor(item.diskUsage || 0)"
                        :stroke-width="8"
                        :show-text="true"
                        :format="() => (item.diskUsage || 0).toFixed(1) + '%'"
                      />
                    </div>
                  </div>
                </div>
              </ElCard>
            </ElCol>
          </ElRow>

          <!-- 实时告警 -->
          <ElCard shadow="never" header="实时告警">
            <ElTable :data="overview.activeAlerts || []" border stripe size="small" max-height="300">
              <ElTableColumn label="级别" width="80" align="center">
                <template #default="{ row }">
                  <ElTag :type="getAlertTagType(row.level)" size="small">
                    {{ getAlertLevelText(row.level) }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn label="主机" prop="hostname" min-width="140" />
              <ElTableColumn label="告警消息" prop="message" min-width="260" show-overflow-tooltip />
              <ElTableColumn label="首次发生" width="180">
                <template #default="{ row }">{{ formatTime(row.firstSeen) }}</template>
              </ElTableColumn>
              <ElTableColumn label="最后发生" width="180">
                <template #default="{ row }">{{ formatTime(row.lastSeen) }}</template>
              </ElTableColumn>
            </ElTable>
          </ElCard>

          <!-- 刷新时间 -->
          <div class="text-right">
            <span class="text-sm text-gray-400">
              最后更新: {{ formatTime(overview.refreshTime || new Date().toISOString()) }}
            </span>
          </div>
        </div>
      </template>
    </ElSkeleton>
  </div>
</template>

<style scoped>
.stat-card {
  margin-bottom: 0;
  transition: all 0.3s;
}

.stat-content {
  text-align: center;
  padding: 8px 0;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 8px;
}

.top-list {
  max-height: 400px;
  overflow-y: auto;
}

.top-list-item {
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.top-list-item:last-child {
  border-bottom: none;
}

.top-list-rank {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  line-height: 28px;
  text-align: center;
  background-color: #f0f0f0;
  border-radius: 50%;
  font-weight: bold;
  font-size: 13px;
  margin-right: 10px;
}

.rank-1 {
  background-color: #ffd700;
  color: #fff;
}

.rank-2 {
  background-color: #c0c0c0;
  color: #fff;
}

.rank-3 {
  background-color: #cd7f32;
  color: #fff;
}

.top-list-content {
  flex: 1;
  min-width: 0;
  margin-right: 8px;
}

.top-list-title {
  font-weight: 500;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.top-list-subtitle {
  font-size: 12px;
  color: #999;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.top-list-value {
  flex-shrink: 0;
  width: 130px;
}
</style>
