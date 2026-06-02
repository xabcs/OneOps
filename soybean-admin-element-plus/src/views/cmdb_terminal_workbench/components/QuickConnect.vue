<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { ElButton, ElInput, ElTag, ElTooltip } from 'element-plus';
import { fetchCheckConnectPermission, fetchGetServers } from '@/service/api';
import type { ApiServer } from '@/typings/api';

interface QuickConnectHost {
  id: number;
  hostname: string;
  ip: string;
  agentStatus: string;
  credentials?: string[];
  canConnect: boolean;
}

interface Props {
  currentSessions: number[]; // 当前已连接的主机ID列表
}

interface Emits {
  (e: 'connect', host: QuickConnectHost): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const hosts = ref<QuickConnectHost[]>([]);
const loading = ref(false);
const searchKeyword = ref('');
const expanded = ref(false);

// 过滤后的主机列表
const filteredHosts = computed(() => {
  const keyword = searchKeyword.value.toLowerCase().trim();
  if (!keyword) return hosts.value;

  return hosts.value.filter(
    host => host.hostname.toLowerCase().includes(keyword) || host.ip.toLowerCase().includes(keyword)
  );
});

// 已连接的主机
const connectedHosts = computed(() => hosts.value.filter(h => props.currentSessions.includes(h.id)));

// 未连接的主机
const availableHosts = computed(() => hosts.value.filter(h => !props.currentSessions.includes(h.id) && h.canConnect));

// 加载可连接的主机列表
async function loadHosts() {
  loading.value = true;
  try {
    const { data } = await fetchGetServers({
      page: 1,
      pageSize: 100,
      agentStatus: 'online'
    });

    if (data?.records) {
      // 并行检查每个主机的连接权限
      const hostsWithPermission = await Promise.all(
        data.records.map(async (server: ApiServer) => {
          try {
            const permission = await fetchCheckConnectPermission(server.id);
            return {
              id: server.id,
              hostname: server.hostname,
              ip: server.ip,
              agentStatus: server.agentStatus,
              credentials: server.credentials,
              canConnect: permission.data?.allowed || false
            };
          } catch {
            return {
              id: server.id,
              hostname: server.hostname,
              ip: server.ip,
              agentStatus: server.agentStatus,
              credentials: server.credentials,
              canConnect: false
            };
          }
        })
      );

      hosts.value = hostsWithPermission.filter(h => h.agentStatus === 'online');
    }
  } catch (error) {
    console.error('加载主机列表失败:', error);
  } finally {
    loading.value = false;
  }
}

// 连接主机
function handleConnect(host: QuickConnectHost) {
  emit('connect', host);
}

// 切换展开/折叠
function toggleExpand() {
  expanded.value = !expanded.value;
  if (expanded.value && hosts.value.length === 0) {
    loadHosts();
  }
}

// 格式化主机标签
function getHostTags(host: QuickConnectHost) {
  const tags = [];
  if (props.currentSessions.includes(host.id)) {
    tags.push({ text: '已连接', type: 'success' });
  } else if (!host.canConnect) {
    tags.push({ text: '无权限', type: 'info' });
  } else if (!host.credentials?.length) {
    tags.push({ text: '无凭证', type: 'warning' });
  }
  return tags;
}
</script>

<template>
  <div class="quick-connect">
    <div class="section-header" @click="toggleExpand">
      <span class="section-title">
        <icon-mdi-server class="icon" />
        可连接主机
      </span>
      <div class="header-actions">
        <ElButton v-if="expanded && hosts.length > 0" size="small" link @click.stop="loadHosts">
          <icon-mdi-refresh :class="{ spinning: loading }" />
        </ElButton>
        <icon-mdi-chevron-down :class="{ rotated: expanded }" class="chevron" />
      </div>
    </div>

    <div v-if="expanded" class="host-list">
      <!-- 搜索框 -->
      <div class="search-box">
        <icon-mdi-magnify class="search-icon" />
        <ElInput v-model="searchKeyword" placeholder="搜索主机..." size="small" clearable />
      </div>

      <!-- 主机列表 -->
      <div v-if="loading" class="loading-state">
        <icon-mdi-loading class="loading-icon" />
        <span>加载中...</span>
      </div>

      <div v-else-if="filteredHosts.length === 0" class="empty-state">
        <icon-mdi-server-off class="empty-icon" />
        <p>{{ searchKeyword ? '没有找到匹配的主机' : '暂无可连接主机' }}</p>
      </div>

      <div v-else class="hosts">
        <div
          v-for="host in filteredHosts"
          :key="host.id"
          class="host-item"
          :class="{
            disabled: !host.canConnect || !host.credentials?.length,
            connected: currentSessions.includes(host.id)
          }"
          @click="handleConnect(host)"
        >
          <div class="host-info">
            <div class="host-header">
              <span class="host-name">{{ host.hostname }}</span>
              <ElTag v-if="currentSessions.includes(host.id)" size="small" type="success">已连接</ElTag>
              <ElTag v-else-if="!host.canConnect" size="small" type="info">无权限</ElTag>
              <ElTag v-else-if="!host.credentials?.length" size="small" type="warning">无凭证</ElTag>
            </div>
            <div class="host-meta">
              <span>{{ host.ip }}</span>
            </div>
          </div>
          <icon-mdi-chevron-right class="arrow" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.quick-connect {
  border-top: 1px solid #3e3e42;
  margin-top: 8px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;
}

.section-header:hover {
  background: #2a2d2e;
}

.section-title {
  display: flex;
  align-items: center;
  font-size: 12px;
  font-weight: 600;
  color: #cccccc;
}

.icon {
  margin-right: 6px;
  font-size: 14px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.chevron {
  font-size: 14px;
  transition: transform 0.3s;
  color: #858585;
}

.chevron.rotated {
  transform: rotate(180deg);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.host-list {
  padding: 8px;
}

.search-box {
  position: relative;
  margin-bottom: 8px;
}

.search-icon {
  position: absolute;
  left: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 14px;
  color: #858585;
}

.search-box :deep(.el-input__wrapper) {
  padding-left: 32px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  color: #858585;
  font-size: 12px;
  gap: 8px;
}

.loading-icon {
  animation: spin 1s linear infinite;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 12px;
  color: #858585;
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 12px;
}

.hosts {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.host-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.host-item:hover:not(.disabled) {
  background: #2a2d2e;
}

.host-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.host-item.connected {
  background: #1e3a2e;
}

.host-info {
  flex: 1;
  min-width: 0;
}

.host-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 2px;
}

.host-name {
  font-size: 13px;
  font-weight: 500;
  color: #cccccc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.host-meta {
  font-size: 11px;
  color: #858585;
}

.arrow {
  font-size: 14px;
  color: #858585;
  margin-left: 8px;
}
</style>
