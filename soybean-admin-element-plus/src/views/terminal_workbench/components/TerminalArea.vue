<script setup lang="ts">
import { Icon } from '@iconify/vue';
import XTermTerminal from './XTermTerminal.vue';
import { watch } from 'vue';

interface Props {
  activeSession: any;
  sessions: any[];
}

interface Emits {
  (e: 'toggleFullscreen'): void;
}

const props = defineProps<Props>();
defineEmits<Emits>();

// 调试：监听活动会话变化
watch(() => props.activeSession, (newVal) => {
  console.log('=== TerminalArea 活动会话变化 ===');
  console.log('活动会话:', newVal);
  console.log('所有会话:', props.sessions);
}, { deep: true });
</script>

<template>
  <div class="wb-terminal-area">
    <!-- 终端工具栏 -->
    <div class="wb-terminal-toolbar">
      <div class="wb-toolbar-section">
        <button type="button" class="wb-toolbar-btn" title="监控">
          <Icon icon="lucide:line-chart" class="wb-btn-icon" />
        </button>
      </div>
      <div class="wb-toolbar-section">
        <div class="wb-terminal-mode">
          <button type="button" class="wb-mode-btn">Shell</button>
          <button type="button" class="wb-mode-btn wb-mode-active">Agent</button>
        </div>
        <button type="button" class="wb-toolbar-btn" title="帮助">
          <Icon icon="lucide:circle-help" class="wb-btn-icon" />
        </button>
      </div>
      <div class="wb-toolbar-section">
        <button type="button" class="wb-toolbar-btn" title="安全连接">
          <Icon icon="lucide:shield-check" class="wb-btn-icon" />
        </button>
      </div>
    </div>

    <!-- 终端内容 -->
    <div class="wb-terminal-content">
      <div v-if="!activeSession" class="wb-terminal-placeholder">
        <Icon icon="lucide:terminal" class="wb-placeholder-icon" />
        <p class="wb-placeholder-title">终端工作台</p>
        <p class="wb-placeholder-desc">请从左侧主机资产中选择主机进行连接</p>
        <p v-if="sessions.length === 0" class="wb-placeholder-hint">当前会话数: 0</p>
        <p v-else class="wb-placeholder-hint">当前会话数: {{ sessions.length }}</p>
      </div>
      <div v-else class="wb-terminal-session">
        <XTermTerminal
          :key="activeSession.id"
          :session-id="activeSession.id"
          :server-id="activeSession.serverId"
          :server-name="activeSession.serverName"
          :server-ip="activeSession.serverIp"
          :login-account="activeSession.loginAccount"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss">
.wb-terminal-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  overflow: hidden;
  min-width: 0;
}

.wb-terminal-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 12px;
  height: 32px;
  background: #252526;
  border-bottom: 1px solid #000000;
  flex-shrink: 0;
}

.wb-toolbar-section {
  display: flex;
  align-items: center;
  gap: 4px;
}

.wb-toolbar-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent !important;
  border: none !important;
  cursor: pointer;
  border-radius: 2px;
  transition: background 0.1s;
}

.wb-toolbar-btn:hover {
  background: rgba(0, 0, 0, 0.2) !important;
}

.wb-toolbar-btn:hover .iconify,
.wb-toolbar-btn:hover .wb-btn-icon {
  color: #aaaaaa !important;
}

.wb-btn-icon {
  width: 14px;
  height: 14px;
  color: #858585;
}

.wb-terminal-mode {
  display: flex;
  background: #1e1e1e;
  border-radius: 2px;
  padding: 1px;
}

.wb-mode-btn {
  padding: 2px 8px;
  font-size: 11px;
  background: transparent;
  border: none;
  color: #858585;
  cursor: pointer;
  border-radius: 1px;
  transition: all 0.1s;
}

.wb-mode-btn:hover {
  color: #cccccc;
}

.wb-mode-active {
  background: #007acc;
  color: #ffffff;
}

.wb-terminal-content {
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.wb-terminal-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #858585;
  gap: 8px;
}

.wb-placeholder-icon {
  width: 48px;
  height: 48px;
  opacity: 0.3;
}

.wb-placeholder-title {
  font-size: 14px;
  font-weight: 500;
  color: #cccccc;
  margin: 0;
}

.wb-placeholder-desc {
  font-size: 12px;
  color: #6e6e6e;
  margin: 0;
}

.wb-placeholder-hint {
  font-size: 11px;
  color: #4a4a4a;
  margin: 0;
}

.wb-terminal-session {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
</style>
