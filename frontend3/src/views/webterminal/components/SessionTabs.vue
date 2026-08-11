<script setup lang="ts">
import { Icon } from '@iconify/vue';
import type { WorkbenchSession } from '../composables/useSessions';

interface Props {
  sessions: WorkbenchSession[];
  activeId: number | string | null;
  currentSessionIds: number[];
  showSidebar?: boolean;
}

interface Emits {
  (e: 'select', sessionId: number | string): void;
  (e: 'remove', sessionId: number | string): void;
  (e: 'connect', server: WorkbenchSession): void;
  (e: 'toggleSidebar'): void;
  (e: 'toggleFullscreen'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

function handleClose(session: WorkbenchSession) {
  emit('remove', session.id);
}
</script>

<template>
  <div class="wb-tab-bar">
    <!-- 左侧工具栏 -->
    <div class="wb-toolbar-group">
      <button type="button" class="wb-toolbar-item" @click="$emit('toggleSidebar')">
        <Icon icon="lucide:panel-left" class="wb-icon" />
      </button>
    </div>

    <!-- 标签列表 -->
    <div class="wb-tab-list">
      <div v-if="sessions.length === 0" class="wb-tab-empty">
        <Icon icon="lucide:terminal" class="wb-empty-icon" />
        <span>暂无会话，请从左侧选择主机连接</span>
      </div>
      <div
        v-for="session in sessions"
        :key="session.id"
        class="wb-tab-item"
        :class="{ active: session.id === activeId }"
        @click="$emit('select', session.id)"
      >
        <Icon :icon="session.isSessionListView ? 'lucide:list-tree' : 'lucide:terminal'" class="wb-tab-icon" />
        <span class="wb-tab-label">{{ session.title || session.serverName }}</span>
        <button type="button" class="wb-tab-close" @click.stop="handleClose(session)">
          <Icon icon="lucide:x" class="wb-close-icon" />
        </button>
      </div>
    </div>

    <!-- 右侧工具栏 -->
    <div class="wb-toolbar-group">
      <button type="button" class="wb-toolbar-item" @click="$emit('toggleFullscreen')">
        <Icon icon="lucide:expand" class="wb-icon" />
      </button>
    </div>
  </div>
</template>

<style lang="scss">
.wb-tab-bar {
  display: flex;
  align-items: center;
  height: 35px;
  background: #252526;
  flex-shrink: 0;
}

.wb-toolbar-group {
  display: flex;
  align-items: center;
  gap: 0;
  flex-shrink: 0;
  height: 100%;
}

.wb-toolbar-item {
  width: 35px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent !important;
  border: none !important;
  cursor: pointer;
  position: relative;
  transition: background 0.1s;
}

.wb-toolbar-item:hover {
  background: rgba(0, 0, 0, 0.2) !important;
}

.wb-toolbar-item:hover .iconify,
.wb-toolbar-item:hover .wb-icon {
  color: #aaaaaa !important;
}

.wb-icon {
  width: 16px;
  height: 16px;
  color: #858585;
  transition: color 0.15s;
}

.wb-tab-list {
  flex: 1;
  display: flex;
  align-items: center;
  overflow-x: auto;
  overflow-y: hidden;
  min-width: 0;
  height: 100%;
  padding-left: 16px;
  gap: 8px;
}

.wb-tab-list::-webkit-scrollbar {
  height: 3px;
}

.wb-tab-list::-webkit-scrollbar-track {
  background: transparent;
}

.wb-tab-list::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 2px;
}

.wb-tab-list::-webkit-scrollbar-thumb:hover {
  background: #4f4f4f;
}

.wb-tab-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  height: 100%;
  color: #6e6e6e;
  font-size: 12px;
  padding: 0 16px;
}

.wb-empty-icon {
  width: 14px;
  height: 14px;
  opacity: 0.5;
}

.wb-tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 8px;
  height: 100%;
  background: #2d2d2d;
  cursor: pointer;
  white-space: nowrap;
  min-width: 0;
  transition: background 0.1s;
}

.wb-tab-item:hover {
  background: #2a2d2e;
}

.wb-tab-item.active {
  background: #1e1e1e;
}

.wb-tab-icon {
  width: 14px;
  height: 14px;
  color: #858585;
  flex-shrink: 0;
}

.wb-tab-item.active .wb-tab-icon {
  color: #cccccc;
}

.wb-tab-label {
  font-size: 12px;
  color: #ffffff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.wb-tab-close {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  cursor: pointer;
  border-radius: 2px;
  flex-shrink: 0;
  opacity: 0;
  transition: all 0.1s;
}

.wb-tab-item:hover .wb-tab-close {
  opacity: 1;
}

.wb-tab-close:hover {
  background: rgba(0, 0, 0, 0.3);
}

.wb-close-icon {
  width: 12px;
  height: 12px;
  color: #858585;
}

.wb-tab-close:hover .wb-close-icon {
  color: #aaaaaa;
}
</style>
