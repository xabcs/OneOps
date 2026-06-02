<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue';
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

function handleConnected(sessionId: number) {
  emit('connected', sessionId);
}

function handleDisconnected(sessionId: number, reason: string) {
  emit('disconnected', sessionId, reason);
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
