<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElButton } from 'element-plus';

interface Session {
  id: number;
  serverId: number;
  serverName: string;
  serverIp: string;
  loginAccount: string;
  connected: boolean;
  duration: number;
}

interface QuickConnectHost {
  id: number;
  hostname: string;
  ip: string;
}

interface Props {
  sessions: Session[];
  activeId: number | null;
  currentSessionIds: number[];
}

interface Emits {
  (e: 'select', sessionId: number): void;
  (e: 'remove', sessionId: number): void;
  (e: 'quickConnect', host: QuickConnectHost): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const showQuickConnect = ref(false);
const searchKeyword = ref('');
const hosts = ref<QuickConnectHost[]>([]);

// 格式化时长
function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) {
    const mins = Math.floor(seconds / 60);
    return `${mins}分`;
  }
  const hours = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  return `${hours}小时${mins}分`;
}

// 处理快速连接
function handleQuickConnect(host: QuickConnectHost) {
  emit('quickConnect', host);
}
</script>

<template>
  <div class="sidebar">
    <!-- 会话列表 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">会话 ({{ sessions.length }})</span>
        <ElButton size="small" type="primary" @click="showQuickConnect = !showQuickConnect">
          <icon-mdi-plus class="icon" />
          {{ showQuickConnect ? '收起' : '主机' }}
        </ElButton>
      </div>

      <div v-if="sessions.length === 0" class="empty">
        <icon-mdi-clipboard-off class="empty-icon" />
        <p>暂无会话</p>
        <ElButton type="primary" size="small" @click="showQuickConnect = true">
          <icon-mdi-plus class="icon" />
          连接主机
        </ElButton>
      </div>

      <div v-else class="session-list">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: session.id === activeId }"
          @click="$emit('select', session.id)"
        >
          <span class="status" :class="{ online: session.connected }"></span>
          <div class="session-info">
            <div class="name">{{ session.serverName }}</div>
            <div class="meta">{{ session.loginAccount }}@{{ session.serverIp }}</div>
          </div>
          <ElButton size="small" type="danger" link @click.stop="$emit('remove', session.id)">
            <icon-mdi-close />
          </ElButton>
        </div>
      </div>
    </div>

    <!-- 快速连接面板 -->
    <div v-if="showQuickConnect" class="section quick-connect">
      <div class="section-header">
        <span class="section-title">可连接主机</span>
        <ElButton size="small" link @click="showQuickConnect = false">
          <icon-mdi-chevron-up />
        </ElButton>
      </div>
      <div class="quick-connect-info">
        <p>请从主机列表点击"连接"按钮</p>
        <ElButton type="success" size="small" @click="window.open('/cmdb/servers', '_blank')">
          <icon-mdi-server class="icon" />
          打开主机列表
        </ElButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  width: 240px;
  height: 100vh;
  background: #252526;
  border-right: 1px solid #3e3e42;
  display: flex;
  flex-direction: column;
}

.section {
  padding: 12px;
}

.section:not(:last-child) {
  border-bottom: 1px solid #3e3e42;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #cccccc;
}

.icon {
  margin-right: 4px;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 12px;
  color: #858585;
}

.empty-icon {
  font-size: 40px;
  margin-bottom: 12px;
  opacity: 0.5;
}

.empty p {
  margin: 0 0 16px 0;
  font-size: 12px;
}

.session-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.session-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.session-item:hover {
  background: #2a2d2e;
}

.session-item.active {
  background: #37373d;
}

.status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
  background: #757575;
}

.status.online {
  background: #4caf50;
  box-shadow: 0 0 4px #4caf50;
}

.session-info {
  flex: 1;
  min-width: 0;
}

.name {
  font-size: 13px;
  color: #cccccc;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.meta {
  font-size: 11px;
  color: #858585;
  margin-top: 2px;
}

.quick-connect {
  flex: 1;
}

.quick-connect-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 12px;
  color: #858585;
  text-align: center;
}

.quick-connect-info p {
  margin: 0 0 16px 0;
  font-size: 12px;
}
</style>
