<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onMounted, onUnmounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { onKeyStroke } from '@vueuse/core';
import { fetchGetServerById } from '@/service/api';
import { localStg } from '@/utils/storage';
import SessionTabs from './components/SessionTabs.vue';
import TerminalArea from './components/TerminalArea.vue';
import AssetTree from './components/AssetTree.vue';
import { useSessions } from './composables/useSessions';
import { useBroadcast } from './composables/useBroadcast';

// 懒加载连接对话框组件
const ServerConnectDialog = defineAsyncComponent(() =>
  import('@/components/ServerConnectDialog/index.vue')
);

defineOptions({ name: 'TerminalWorkbench' });

const route = useRoute();
const { sessions, activeSession, addSession, removeSession, switchSession, updateSession, terminateAllSessions } = useSessions();
const { broadcast, onMessage, dispose } = useBroadcast('oneops-workbench');

const showAssetTree = ref(true);
const isFullscreen = ref(false);
const sidebarWidth = ref(240);
const isResizing = ref(false);

// 连接对话框
const showConnectDialog = ref(false);
const connectingServer = ref<CMDB.Server | null>(null);

// 监听 ESC 键退出全屏
onKeyStroke('Escape', () => {
  if (isFullscreen.value) {
    toggleFullscreen();
  }
});

// 开始调整大小
function startResize(e: MouseEvent) {
  isResizing.value = true;
  const startX = e.clientX;
  const startWidth = sidebarWidth.value;

  const onMouseMove = (e: MouseEvent) => {
    if (!isResizing.value) return;
    const diff = e.clientX - startX;
    const newWidth = Math.max(100, Math.min(600, startWidth + diff));
    sidebarWidth.value = newWidth;
  };

  const onMouseUp = () => {
    isResizing.value = false;
    window.removeEventListener('mousemove', onMouseMove);
    window.removeEventListener('mouseup', onMouseUp);
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  };

  window.addEventListener('mousemove', onMouseMove);
  window.addEventListener('mouseup', onMouseUp);
  document.body.style.cursor = 'col-resize';
  document.body.style.userSelect = 'none';
}

// 清理
onUnmounted(() => {
  if (isResizing.value) {
    isResizing.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  }
  // 移除页面关闭监听
  window.removeEventListener('beforeunload', handleBeforeUnload);
});

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

onMounted(async () => {
  broadcast('workbench-opened', { workbenchId: 'oneops-workbench' });

  // 从 URL 参数加载服务器信息并显示连接对话框
  const serverId = route.query.serverId as string;
  const serverName = route.query.serverName as string;
  const serverIp = route.query.serverIp as string;
  const serverEnv = route.query.serverEnv as string;
  const credentialId = route.query.credentialId as string;

  if (serverId) {
    try {
      // 获取服务器详情
      const serverRes = await fetchGetServerById(Number(serverId));
      const serverDetail = serverRes.data;

      if (serverDetail) {
        // 设置凭证ID
        if (credentialId) {
          (serverDetail as any).credentialId = Number(credentialId);
        }
        connectingServer.value = serverDetail;
        showConnectDialog.value = true;
      }
    } catch (error) {
      console.error('获取服务器详情失败:', error);
      window.$message?.error('获取服务器详情失败');
    }
  }

  // 从 URL 参数加载初始会话（用于已经建立的连接）
  const sessionId = route.query.sessionId as string;
  const websocketUrl = route.query.websocketUrl as string;
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

// 页面关闭前终止所有会话
function handleBeforeUnload() {
  if (sessions.value.length > 0) {
    const token = localStg.get('token') || '';

    // 使用 fetch + keepalive 替代 sendBeacon，支持 Authorization header
    sessions.value.forEach(session => {
      fetch(`/api/cmdb/sessions/${session.id}/terminate`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        keepalive: true  // 允许页面关闭后继续发送
      }).catch(err => {
        // 静默处理错误，页面关闭时无法显示错误
        console.error('终止会话失败:', err);
      });
    });
  }
}

onMounted(async () => {
  broadcast('workbench-opened', { workbenchId: 'oneops-workbench' });

  // 添加页面关闭监听
  window.addEventListener('beforeunload', handleBeforeUnload);

  // 从 URL 参数加载服务器信息并显示连接对话框
  const serverId = route.query.serverId as string;
  const serverName = route.query.serverName as string;
  const serverIp = route.query.serverIp as string;
  const serverEnv = route.query.serverEnv as string;
  const credentialId = route.query.credentialId as string;

  if (serverId) {
    try {
      // 获取服务器详情
      const serverRes = await fetchGetServerById(Number(serverId));
      const serverDetail = serverRes.data;

      if (serverDetail) {
        // 设置凭证ID
        if (credentialId) {
          (serverDetail as any).credentialId = Number(credentialId);
        }
        connectingServer.value = serverDetail;
        showConnectDialog.value = true;
      }
    } catch (error) {
      console.error('获取服务器详情失败:', error);
      window.$message?.error('获取服务器详情失败');
    }
  }

  // 从 URL 参数加载初始会话（用于已经建立的连接）
  const sessionId = route.query.sessionId as string;
  const websocketUrl = route.query.websocketUrl as string;
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

// 处理主机连接（优化：直接使用传入的 server 对象，避免重复请求）
function handleConnect(server: CMDB.Server & { loginAccount?: string }) {
  console.log('[index.vue] handleConnect 被调用:', server);
  // 直接使用传入的 server 对象，不再重复请求服务器详情
  connectingServer.value = server;
  showConnectDialog.value = true;
}

// 连接成功回调
function handleConnected(sessionId: number, websocketUrl: string, loginAccount: string) {
  console.log('[handleConnected] 参数:', { sessionId, websocketUrl, loginAccount });
  console.log('[handleConnected] connectingServer:', connectingServer.value);

  if (!connectingServer.value) {
    console.error('[handleConnected] connectingServer 为空');
    return;
  }

  const serverName = connectingServer.value.hostname || '未知主机';
  const serverIp = connectingServer.value.ip || '';

  console.log('[handleConnected] 准备添加会话:', {
    id: sessionId,
    serverId: connectingServer.value.id,
    serverName,
    serverIp,
    loginAccount
  });

  try {
    // 在当前工作台添加会话
    addSession({
      id: sessionId,
      serverId: connectingServer.value.id,
      serverName,
      serverIp,
      loginAccount: loginAccount || connectingServer.value.sshUser || 'root',
      protocol: 'ssh',
      status: 'connected',
      connected: true,
      duration: 0,
      startedAt: new Date().toISOString(),
      websocketUrl
    });

    showConnectDialog.value = false;
    const hostname = connectingServer.value.hostname;
    connectingServer.value = null;
    window.$message?.success(`正在连接到 ${hostname}...`);
  } catch (error) {
    console.error('[handleConnected] 添加会话失败:', error);
    window.$message?.error('连接失败，请重试');
  }
}

// 切换资产树显示
function toggleAssetTree() {
  showAssetTree.value = !showAssetTree.value;
}

// 切换全屏
function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value;
}
</script>

<template>
  <div class="workbench" :class="{ fullscreen: isFullscreen }">
    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 左侧资产树 -->
      <div v-if="showAssetTree && !isFullscreen" class="asset-sidebar" :style="{ width: sidebarWidth + 'px' }">
        <AssetTree :current-sessions="currentSessionIds" @connect="handleConnect" />
      </div>

      <!-- 拖动分隔条 -->
      <div v-if="showAssetTree && !isFullscreen" class="resize-handle" @mousedown="startResize" />

      <!-- 右侧终端区域 -->
      <div class="terminal-area-wrapper">
        <SessionTabs
          :sessions="sessions"
          :active-id="activeSession?.id || null"
          :current-session-ids="currentSessionIds"
          :show-asset-tree="showAssetTree && !isFullscreen"
          @select="switchSession"
          @remove="removeSession"
          @connect="handleConnect"
          @toggle-asset-tree="toggleAssetTree"
          @toggle-fullscreen="toggleFullscreen"
        />
        <TerminalArea
          :active-session="activeSession"
          :sessions="sessions"
          @connected="handleSessionConnected"
          @disconnected="handleSessionDisconnected"
          @toggle-fullscreen="toggleFullscreen"
        />
      </div>
    </div>

    <!-- 连接对话框（懒加载 + 预加载 + 缓存） -->
    <Suspense>
      <template #default>
        <KeepAlive>
          <ServerConnectDialog
            v-if="connectingServer"
            v-model:visible="showConnectDialog"
            :server-id="connectingServer.id"
            :server-name="connectingServer.hostname"
            :server-ip="connectingServer.ip"
            :server-env="connectingServer.env"
            :credential-id="(connectingServer as any).credentialId"
            @connected="handleConnected"
          />
        </KeepAlive>
      </template>
      <template #fallback>
        <!-- 加载状态 -->
        <ElDialog
          :model-value="showConnectDialog"
          title="连接主机"
          width="600px"
          :close-on-click-modal="false"
        >
          <div class="dialog-loading">
            <icon-mdi-loading class="loading-icon" />
            <span>加载中...</span>
          </div>
        </ElDialog>
      </template>
    </Suspense>
  </div>
</template>

<style scoped>
.workbench {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100vh;
  background: #050505;
  overflow: hidden;
}

.workbench.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

.asset-sidebar {
  background: #0a0a0a;
  border-right: 1px solid #333333;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
  transition: width 0.15s;
}

.resize-handle {
  width: 3px;
  background: transparent;
  cursor: col-resize;
  flex-shrink: 0;
  transition: background 0.1s;
}

.resize-handle:hover {
  background: #3a3a3a;
}

.resize-handle:active {
  background: #4a4a4a;
}

.terminal-area-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #050505;
}

/* 对话框加载状态 */
.dialog-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 12px;
  color: #888;
}

.dialog-loading .loading-icon {
  font-size: 24px;
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
</style>
