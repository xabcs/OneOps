<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElButton, ElTooltip } from 'element-plus';

interface Session {
  id: number;
  serverId: number;
  serverName: string;
  serverIp: string;
  loginAccount: string;
  connected: boolean;
  duration: number;
  startedAt?: string;
}

interface Props {
  sessions: Session[];
  activeId: number | null;
  currentSessionIds: number[];
  showAssetTree?: boolean;
}

interface Emits {
  (e: 'select', sessionId: number): void;
  (e: 'remove', sessionId: number): void;
  (e: 'connect', server: CMDB.Server & { loginAccount?: string }): void;
  (e: 'toggleAssetTree'): void;
  (e: 'toggleFullscreen'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

// 格式化时长
function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) {
    const mins = Math.floor(seconds / 60);
    return `${mins}m`;
  }
  const hours = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  return `${hours}h ${mins}m`;
}

// 计算会话时长（实时）
const sessionDurations = computed<Record<number, string>>(() => {
  const result: Record<number, string> = {};
  props.sessions.forEach(session => {
    if (session.startedAt) {
      const started = new Date(session.startedAt).getTime();
      const now = Date.now();
      const seconds = Math.floor((now - started) / 1000);
      result[session.id] = formatDuration(seconds);
    } else {
      result[session.id] = formatDuration(session.duration || 0);
    }
  });
  return result;
});

// 获取标签页标题（优化：显示完整信息，格式：hostname (user@ip)）
function getTabTitle(session: Session): string {
  // 为了方便截断，分开处理
  return session.serverName;
}

// 获取标签副标题（user@ip）
function getTabSubtitle(session: Session): string {
  return `${session.loginAccount}@${session.serverIp}`;
}
</script>

<template>
  <div class="session-tabs-container">
    <!-- 标签页栏 -->
    <div class="tabs-header">
      <!-- 左侧工具栏 -->
      <div class="toolbar-left">
        <ElTooltip :content="showAssetTree ? '隐藏资产' : '显示资产'" placement="bottom">
          <ElButton size="small" @click="$emit('toggleAssetTree')">
            <icon-mdi-chevron-left v-if="showAssetTree" class="toolbar-icon" />
            <icon-mdi-chevron-right v-else class="toolbar-icon" />
          </ElButton>
        </ElTooltip>
      </div>

      <!-- 会话标签 -->
      <div class="tabs-wrapper">
        <div v-if="sessions.length === 0" class="empty-tabs">
          <icon-mdi-clipboard-off class="empty-icon" />
          <span>暂无会话</span>
        </div>

        <div v-else class="tabs-list">
          <div
            v-for="session in sessions"
            :key="session.id"
            class="tab-item"
            :class="{ active: session.id === activeId }"
            @click="$emit('select', session.id)"
          >
            <span class="status" :class="{ online: session.connected }"></span>
            <div class="tab-content">
              <span class="tab-title">{{ getTabTitle(session) }}</span>
              <span class="tab-subtitle">{{ getTabSubtitle(session) }}</span>
            </div>
            <span class="tab-duration">{{ sessionDurations[session.id] || '0s' }}</span>
            <ElButton size="small" type="danger" link class="close-btn" @click.stop="$emit('remove', session.id)">
              <icon-mdi-close />
            </ElButton>
          </div>
        </div>
      </div>

      <!-- 右侧工具栏 -->
      <div class="toolbar-right">
        <ElTooltip content="全屏" placement="bottom">
          <ElButton size="small" @click="$emit('toggleFullscreen')">
            <icon-mdi-arrow-expand-all class="toolbar-icon" />
          </ElButton>
        </ElTooltip>
      </div>
    </div>
  </div>
</template>

<style scoped>
.session-tabs-container {
  display: flex;
  min-height: 30px;
  background: #121212;
  flex-shrink: 0;
}

.tabs-header {
  display: flex;
  align-items: center;
  min-height: 28px;
  flex: 1;
}

.toolbar-left {
  display: flex;
  align-items: center;
  height: 28px;
  flex-shrink: 0;
}

.toolbar-left .el-button {
  padding: 0 6px;
  height: 28px;
  background: transparent;
  border: none;
  color: #999;
}

.toolbar-left .el-button:hover {
  color: #ccc;
}

.toolbar-right {
  display: flex;
  align-items: center;
  height: 28px;
  flex-shrink: 0;
}

.toolbar-right .el-button {
  padding: 0 6px;
  height: 28px;
  background: transparent;
  border: none;
  color: #999;
}

.toolbar-right .el-button:hover {
  color: #ccc;
}

.toolbar-icon {
  font-size: 14px;
}

.tabs-wrapper {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  min-width: 0;
  min-height: 28px;
}

.tabs-wrapper::-webkit-scrollbar {
  height: 0;
}

.empty-tabs {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 28px;
  color: #888;
  font-size: 11px;
}

.empty-icon {
  font-size: 12px;
  opacity: 0.5;
}

.tabs-list {
  display: flex;
  align-items: center;
  min-height: 28px;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 8px;
  background: #1a1a1a;
  border-right: 1px solid #333333;
  cursor: pointer;
  white-space: nowrap;
  min-width: 0;
  max-width: 240px;
  transition: all 0.15s;
}

.tab-item:hover {
  background: #252525;
}

.tab-item.active {
  background: #2a2a2a;
  border-bottom: 2px solid #0dbc79;
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

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 1px;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.tab-title {
  font-size: 11px;
  color: #f0f0f0;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-subtitle {
  font-size: 9px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-duration {
  font-size: 10px;
  color: #c0c0c0;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  flex-shrink: 0;
}

.close-btn {
  flex-shrink: 0;
  padding: 2px;
  margin-left: 2px;
  opacity: 0;
  transition: opacity 0.1s;
  color: #999;
  font-size: 12px;
}

.tab-item:hover .close-btn {
  opacity: 1;
}

.close-btn:hover {
  color: #ccc;
}
</style>
