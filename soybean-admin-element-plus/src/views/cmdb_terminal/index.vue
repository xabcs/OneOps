<script setup lang="ts">
import { ref, onUnmounted } from 'vue';
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
  timer = setInterval(() => { duration.value++; }, 1000);
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
  if (timer) { clearInterval(timer); timer = null; }
}

function handleError(msg: string) {
  hasError.value = true;
  errorMessage.value = msg;
  connected.value = false;
  if (timer) { clearInterval(timer); timer = null; }
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
        <el-tag type="success" size="small">SSH</el-tag>
        <span class="hostname">{{ serverName }}</span>
        <span class="ip">{{ serverIp }}</span>
        <el-divider direction="vertical" />
        <span class="account">{{ loginAccount }}</span>
      </div>
      <div class="session-info">
        <el-tag v-if="connected" type="success" size="small" effect="plain">已连接</el-tag>
        <el-tag v-else type="info" size="small" effect="plain">连接中...</el-tag>
        <span class="duration">{{ formatDuration(duration) }}</span>
        <el-button type="danger" size="small" plain @click="handleDisconnect">断开连接</el-button>
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
          <el-button type="primary" @click="router.push('/cmdb/servers')">返回主机列表</el-button>
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
  padding: 0 16px;
  height: 48px;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
  flex-shrink: 0;
}
.host-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #cccccc;
}
.hostname {
  font-weight: 600;
  color: #ffffff;
}
.ip {
  font-size: 12px;
  color: #9e9e9e;
}
.account {
  font-size: 12px;
  color: #9cdcfe;
}
.session-info {
  display: flex;
  align-items: center;
  gap: 12px;
}
.duration {
  font-family: 'Courier New', monospace;
  color: #4ec9b0;
  font-size: 14px;
  min-width: 60px;
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
  padding: 32px 40px;
  text-align: center;
  max-width: 420px;
}
.error-icon {
  font-size: 36px;
  color: #f44747;
  margin-bottom: 12px;
}
.error-title {
  font-size: 18px;
  font-weight: 600;
  color: #ffffff;
  margin-bottom: 12px;
}
.error-msg {
  font-size: 14px;
  color: #f14c4c;
  margin-bottom: 8px;
  word-break: break-all;
}
.error-hint {
  font-size: 12px;
  color: #9e9e9e;
  margin-bottom: 20px;
  line-height: 1.6;
}
</style>
