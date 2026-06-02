<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue';
import SSHTerminal from '@/components/SSHTerminal/index.vue';

// 扩展 Window 接口
declare global {
  interface Window {
    sessionDurationTimers?: Record<number, any>;
  }
}

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
  sessions: Session[]; // 接收所有会话列表
}

interface Emits {
  (e: 'connected', sessionId: number): void;
  (e: 'disconnected', sessionId: number, reason: string): void;
  (e: 'toggleFullscreen'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

// 为每个会话维护独立的时长计时器
const sessionDurations = ref<Record<number, number>>({});

// 初始化所有会话的时长
watch(
  () => props.sessions,
  sessions => {
    sessions.forEach(session => {
      if (!(session.id in sessionDurations.value)) {
        sessionDurations.value[session.id] = session.duration || 0;
      }
    });
  },
  { immediate: true, deep: true }
);

// 监听活跃会话的时长更新
watch(
  () => props.activeSession,
  (newSession, oldSession) => {
    // 清理旧会话的计时器
    if (oldSession?.id && window.sessionDurationTimers?.[oldSession.id]) {
      clearInterval(window.sessionDurationTimers[oldSession.id]);
      delete window.sessionDurationTimers[oldSession.id];
    }

    if (newSession?.connected) {
      // 更新当前活跃会话的时长
      const duration = sessionDurations.value[newSession.id] || 0;
      if (!window.sessionDurationTimers) {
        window.sessionDurationTimers = {};
      }
      window.sessionDurationTimers[newSession.id] = window.setInterval(() => {
        sessionDurations.value[newSession.id] = (sessionDurations.value[newSession.id] || 0) + 1;
      }, 1000);
    }
  },
  { immediate: true }
);

// 清理所有计时器
onUnmounted(() => {
  if (window.sessionDurationTimers) {
    Object.values(window.sessionDurationTimers).forEach(timer => clearInterval(timer));
    window.sessionDurationTimers = {};
  }
});

function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}

function handleConnected(sessionId: number) {
  emit('connected', sessionId);
}

function handleDisconnected(sessionId: number, reason: string) {
  emit('disconnected', sessionId, reason);
}

function toggleFullscreen() {
  emit('toggleFullscreen');
}

// 获取当前显示的时长
const currentDuration = computed(() => {
  return props.activeSession ? sessionDurations.value[props.activeSession.id] || 0 : 0;
});
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
          <span class="time">{{ formatDuration(currentDuration) }}</span>
          <ElButton size="small" link class="fullscreen-btn" @click="toggleFullscreen">
            <icon-mdi-arrow-expand-all />
          </ElButton>
        </div>
      </div>

      <!-- 终端内容 - 为所有会话创建SSHTerminal实例，但只显示活跃会话 -->
      <div class="terminal-content">
        <template v-for="session in sessions" :key="session.id">
          <SSHTerminal
            v-show="activeSession.id === session.id && session.websocketUrl"
            :session-id="session.id"
            :websocket-url="session.websocketUrl || ''"
            :server-name="session.serverName"
            :server-ip="session.serverIp"
            class="terminal-instance"
            @connected="() => handleConnected(session.id)"
            @disconnected="reason => handleDisconnected(session.id, reason)"
          />
        </template>

        <div v-if="!activeSession.websocketUrl" class="disconnected">
          <icon-mdi-connection class="disconnected-icon" />
          <p>会话已断开</p>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.terminal-area {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #050505;
  overflow: hidden;
}

.empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #888;
}

.empty-icon {
  font-size: 40px;
  margin-bottom: 10px;
  opacity: 0.5;
}

.empty h3 {
  margin: 0 0 4px 0;
  font-size: 13px;
  color: #bbb;
  font-weight: 500;
}

.empty p {
  margin: 0;
  font-size: 11px;
}

.terminal-header {
  height: 24px;
  padding: 0 10px;
  background: #1a1a1a;
  border-bottom: 1px solid #333333;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.status {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #666;
  flex-shrink: 0;
}

.status.online {
  background: #bbb;
}

.name {
  font-size: 11px;
  font-weight: 500;
  color: #e8e8e8;
}

.ip {
  font-size: 10px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.time {
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 10px;
  color: #bbb;
}

.fullscreen-btn {
  padding: 2px;
  color: #999;
  font-size: 14px;
}

.fullscreen-btn:hover {
  color: #ccc;
}

.terminal-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.terminal-instance {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.disconnected {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #888;
}

.disconnected-icon {
  font-size: 32px;
  margin-bottom: 6px;
  opacity: 0.5;
}
</style>
