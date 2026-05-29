<script setup lang="ts">
import { onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import SSHTerminal from '@/components/SSHTerminal/index.vue';

defineOptions({ name: 'CmdbTerminal' });

const route = useRoute();
const router = useRouter();

const sessionId = Number(route.params.id);
const websocketUrl = route.query.websocketUrl as string;
const serverName = (route.query.serverName as string) || '';
const serverIp = (route.query.serverIp as string) || '';
const loginAccount = (route.query.loginAccount as string) || '';

const connected = ref(false);
const hasError = ref(false);
const errorMessage = ref('');
const duration = ref(0);
let timer: ReturnType<typeof setInterval> | null = null;

function startTimer() {
  timer = setInterval(() => {
    duration.value++;
  }, 1000);
}

function formatDuration(s: number) {
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
}

function handleConnected() {
  connected.value = true;
  startTimer();
}

function handleDisconnected() {
  connected.value = false;
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}

function handleError(msg: string) {
  hasError.value = true;
  errorMessage.value = msg;
  connected.value = false;
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}

async function handleDisconnect() {
  try {
    await ElMessageBox.confirm('确认断开 SSH 连接并返回主机列表？', '断开连接', {
      confirmButtonText: '确认断开',
      cancelButtonText: '继续连接',
      type: 'warning'
    });
    router.push('/cmdb/servers');
  } catch {
    // 用户取消
  }
}

onUnmounted(() => {
  if (timer) clearInterval(timer);
});
</script>

<template>
  <div class="terminal-page">
    <div class="terminal-header">
      <div class="host-info">
        <span class="ssh-badge">SSH</span>
        <span class="hostname">{{ serverName }}</span>
        <span class="separator">{{ serverIp }}</span>
        <span class="account">{{ loginAccount }}</span>
      </div>
      <div class="session-info">
        <span class="status-dot" :class="{ connected: connected }"></span>
        <span class="duration">{{ formatDuration(duration) }}</span>
        <ElButton size="small" plain @click="handleDisconnect">断开</ElButton>
      </div>
    </div>
    <div class="terminal-body">
      <!-- 连接失败遮罩 -->
      <div v-if="hasError" class="error-overlay">
        <div class="error-card">
          <div class="error-icon">✕</div>
          <div class="error-title">连接失败</div>
          <div class="error-msg">{{ errorMessage }}</div>
          <div class="error-hint">请确认：服务器已绑定 SSH 凭证，且凭证用户名/密码正确</div>
          <ElButton type="primary" size="small" @click="router.push('/cmdb/servers')">返回主机列表</ElButton>
        </div>
      </div>
      <SSHTerminal
        :session-id="sessionId"
        :websocket-url="websocketUrl"
        :server-name="serverName"
        :server-ip="serverIp"
        @connected="handleConnected"
        @disconnected="handleDisconnected"
        @error="handleError"
      />
    </div>
  </div>
</template>

<style scoped>
.terminal-page {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  overflow: hidden;
}
.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 12px;
  height: 36px;
  background: #2d2d2d;
  border-bottom: 1px solid #3c3c3c;
  flex-shrink: 0;
}
.host-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #cccccc;
  font-size: 13px;
}
.ssh-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 18px;
  background: #0dbc79;
  color: #ffffff;
  font-size: 11px;
  font-weight: 600;
  border-radius: 2px;
  line-height: 1;
}
.hostname {
  font-weight: 500;
  color: #ffffff;
  font-size: 13px;
}
.separator {
  font-size: 12px;
  color: #858585;
  margin: 0 2px;
}
.account {
  font-size: 12px;
  color: #9cdcfe;
  font-family: 'Menlo', 'Monaco', monospace;
}
.session-info {
  display: flex;
  align-items: center;
  gap: 10px;
}
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #808080;
  transition: background 0.2s;
}
.status-dot.connected {
  background: #0dbc79;
  box-shadow: 0 0 4px rgba(13, 188, 121, 0.5);
}
.duration {
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  color: #4ec9b0;
  font-size: 12px;
  min-width: 56px;
  letter-spacing: 0.5px;
}
.terminal-body {
  flex: 1;
  overflow: hidden;
  position: relative;
}
.error-overlay {
  position: absolute;
  inset: 0;
  background: rgba(30, 30, 30, 0.92);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}
.error-card {
  background: #252526;
  border: 1px solid #f44747;
  border-radius: 8px;
  padding: 24px 32px;
  text-align: center;
  max-width: 400px;
}
.error-icon {
  font-size: 32px;
  color: #f44747;
  margin-bottom: 10px;
}
.error-title {
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
  margin-bottom: 10px;
}
.error-msg {
  font-size: 13px;
  color: #f14c4c;
  margin-bottom: 6px;
  word-break: break-all;
}
.error-hint {
  font-size: 11px;
  color: #9e9e9e;
  margin-bottom: 16px;
  line-height: 1.5;
}
</style>
