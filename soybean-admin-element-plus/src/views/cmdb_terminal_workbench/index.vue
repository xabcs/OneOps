<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import SessionSidebar from './components/SessionSidebar.vue';
import TerminalArea from './components/TerminalArea.vue';
import { useSessions } from './composables/useSessions';
import { useBroadcast } from './composables/useBroadcast';

defineOptions({ name: 'CmdbTerminalWorkbench' });

const route = useRoute();
const { sessions, activeSession, addSession, removeSession, switchSession, updateSession } = useSessions();
const { broadcast, onMessage, dispose } = useBroadcast('oneops-workbench');

// 监听其他标签页的消息
onMessage(message => {
  switch (message.type) {
    case 'session-added':
      addSession(message.data.session);
      break;
    case 'session-removed':
      removeSession(message.data.sessionId);
      break;
    case 'session-updated':
      updateSession(message.data.sessionId, message.data.updates);
      break;
    case 'session-activated':
      switchSession(message.data.sessionId);
      break;
  }
});

onMounted(() => {
  broadcast('workbench-opened', { workbenchId: 'oneops-workbench' });

  // 从 URL 参数加载初始会话
  const sessionId = route.query.sessionId as string;
  const websocketUrl = route.query.websocketUrl as string;
  const serverName = route.query.serverName as string;
  const serverIp = route.query.serverIp as string;
  const loginAccount = route.query.loginAccount as string;

  if (sessionId && websocketUrl) {
    addSession({
      id: Number(sessionId),
      serverId: 0,
      serverName: serverName || '未知主机',
      serverIp: serverIp || '',
      loginAccount: loginAccount || '',
      protocol: 'ssh',
      status: 'connected',
      connected: true,
      duration: 0,
      startedAt: new Date().toISOString(),
      websocketUrl
    });
  }
});

onUnmounted(() => {
  dispose();
});

// 会话事件处理
function handleSessionConnected(sessionId: number) {
  updateSession(sessionId, {
    status: 'connected',
    connected: true,
    startedAt: new Date().toISOString()
  });
  broadcast('session-updated', {
    sessionId,
    updates: { status: 'connected', connected: true }
  });
}

function handleSessionDisconnected(sessionId: number, reason: string) {
  updateSession(sessionId, {
    status: 'disconnected',
    connected: false
  });
  broadcast('session-updated', {
    sessionId,
    updates: { status: 'disconnected', connected: false }
  });
}

// 当前会话的 serverId 列表
const currentSessionIds = computed(() => sessions.value.map(s => s.serverId));

// 处理快速连接
function handleQuickConnect(host: any) {
  // 打开主机列表并选中该主机
  window.open(`/cmdb/servers?id=${host.id}`, '_blank');
}
</script>

<template>
  <div class="workbench">
    <SessionSidebar
      :sessions="sessions"
      :active-id="activeSession?.id || null"
      :current-session-ids="currentSessionIds"
      @select="switchSession"
      @remove="removeSession"
      @quick-connect="handleQuickConnect"
    />
    <TerminalArea
      :active-session="activeSession"
      @connected="handleSessionConnected"
      @disconnected="handleSessionDisconnected"
    />
  </div>
</template>

<style scoped>
.workbench {
  display: flex;
  width: 100%;
  height: 100vh;
  background: #1e1e1e;
  overflow: hidden;
}
</style>
