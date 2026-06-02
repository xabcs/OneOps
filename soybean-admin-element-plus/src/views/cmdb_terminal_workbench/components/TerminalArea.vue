<script setup lang="ts">
import { ref, watch } from 'vue';
import SSHTerminal from '@/components/SSHTerminal/index.vue';

interface Session {
  id: number;
  serverName: string;
  serverIp: string;
  loginAccount: string;
  connected: boolean;
  duration: number;
  websocketUrl?: string;
}

interface Props {
  activeSession: Session | null;
}

interface Emits {
  (e: 'connected', sessionId: number): void;
  (e: 'disconnected', sessionId: number, reason: string): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const duration = ref(0);

watch(
  () => props.activeSession,
  newSession => {
    if (newSession?.connected) {
      duration.value = newSession.duration || 0;
    }
  },
  { immediate: true }
);

// 更新时长
let timer: number | null = null;
if (!timer) {
  timer = window.setInterval(() => {
    if (props.activeSession?.connected) {
      duration.value++;
    }
  }, 1000);
}

function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}

function handleConnected() {
  if (props.activeSession) {
    emit('connected', props.activeSession.id);
  }
}

function handleDisconnected(reason: string) {
  if (props.activeSession) {
    emit('disconnected', props.activeSession.id, reason);
  }
}
</script>

<template>
  <div class="terminal-area">
    <div v-if="!activeSession" class="empty">
      <icon-mdi-monitor-off class="empty-icon" />
      <h3>未选择会话</h3>
      <p>请选择一个会话或连接新主机</p>
    </div>

    <template v-else>
      <!-- 终端头部 -->
      <div class="terminal-header">
        <div class="header-left">
          <span class="status" :class="{ online: activeSession.connected }"></span>
          <span class="name">{{ activeSession.serverName }}</span>
          <span class="ip">{{ activeSession.loginAccount }}@{{ activeSession.serverIp }}</span>
        </div>
        <div class="header-right">
          <span class="time">{{ formatDuration(duration) }}</span>
        </div>
      </div>

      <!-- 终端内容 -->
      <div class="terminal-content">
        <SSHTerminal
          v-if="activeSession.websocketUrl"
          :session-id="activeSession.id"
          :websocket-url="activeSession.websocketUrl"
          :server-name="activeSession.serverName"
          :server-ip="activeSession.serverIp"
          @connected="handleConnected"
          @disconnected="handleDisconnected"
        />
        <div v-else class="disconnected">
          <icon-mdi-connection class="disconnected-icon" />
          <p>会话已断开</p>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.terminal-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  overflow: hidden;
}

.empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #858585;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #cccccc;
}

.empty p {
  margin: 0;
  font-size: 13px;
}

.terminal-header {
  height: 36px;
  padding: 0 16px;
  background: #2d2d2d;
  border-bottom: 1px solid #3e3e42;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #757575;
}

.status.online {
  background: #4caf50;
  box-shadow: 0 0 4px #4caf50;
}

.name {
  font-size: 13px;
  font-weight: 600;
  color: #cccccc;
}

.ip {
  font-size: 12px;
  color: #858585;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.time {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: #4caf50;
}

.terminal-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.disconnected {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #858585;
}

.disconnected-icon {
  font-size: 48px;
  margin-bottom: 12px;
  opacity: 0.5;
}
</style>
