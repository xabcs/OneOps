<script setup lang="ts">
import type { TabType } from '../composables/useSessionList';

interface Props {
  activeTab: TabType;
  searchKeyword: string;
  selectedCount: number;
}

interface Emits {
  (e: 'tab-change', tab: TabType): void;
  (e: 'search'): void;
  (e: 'update:searchKeyword', value: string): void;
  (e: 'clear-search'): void;
  (e: 'batch-terminate'): void;
  (e: 'refresh'): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();
</script>

<template>
  <!-- 头部工具栏 -->
  <div class="wb-toolbar">
    <div class="wb-toolbar-left">
      <h1 class="wb-page-title">会话列表</h1>
    </div>
    <div class="wb-toolbar-right">
      <button v-if="selectedCount > 0" class="wb-button wb-button-danger" @click="emit('batch-terminate')">
        终止选中 ({{ selectedCount }})
      </button>
      <button class="wb-button wb-button-default" @click="emit('refresh')">
        <i class="codicon codicon-refresh"></i>
        刷新
      </button>
    </div>
  </div>

  <!-- 标签页 -->
  <div class="wb-tabs">
    <div class="wb-tab" :class="{ active: activeTab === 'active' }" @click="emit('tab-change', 'active')">在线会话</div>
    <div class="wb-tab" :class="{ active: activeTab === 'terminated' }" @click="emit('tab-change', 'terminated')">
      已终止会话
    </div>
    <div class="wb-tab" :class="{ active: activeTab === 'history' }" @click="emit('tab-change', 'history')">
      会话历史
    </div>
  </div>

  <!-- 搜索栏 -->
  <div class="wb-search-bar">
    <div class="wb-search-input-wrapper">
      <i class="codicon codicon-search wb-search-icon"></i>
      <input
        :value="searchKeyword"
        type="text"
        class="wb-search-input"
        placeholder="搜索主机名、IP、用户名..."
        autocomplete="off"
        style="border: none !important; outline: none !important; box-shadow: none !important"
        @input="emit('update:searchKeyword', ($event.target as HTMLInputElement).value)"
        @keyup.enter="emit('search')"
      />
      <button v-if="searchKeyword" class="wb-search-clear" @click="emit('clear-search')">
        <i class="codicon codicon-close"></i>
      </button>
    </div>
    <button class="wb-button wb-button-primary" @click="emit('search')">搜索</button>
  </div>
</template>

<style lang="scss" scoped>
/* ========== 工具栏 ========== */
.wb-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-bottom: none;
  background: #171717;
  min-height: 40px;
}

.wb-toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wb-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wb-page-title {
  font-size: 14px;
  font-weight: 500;
  color: #ffffff;
  margin: 0;
}

/* ========== 按钮样式 ========== */
.wb-button {
  height: 28px;
  padding: 0 12px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: 1px solid #3c3c3c;
  border-radius: 2px;
  color: #ffffff;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: #252526;
    border-color: #404040;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.wb-button-primary {
  background: transparent;
  border: 1px solid #3c3c3c;
  color: #ffffff;

  &:hover {
    background: #252526;
    border-color: #404040;
  }

  &:active {
    background: #2a2d2e;
  }
}

.wb-button-danger {
  background: #f14c4c;
  border-color: #f14c4c;
  color: #ffffff;

  &:hover {
    background: #ff6060;
    border-color: #ff6060;
  }
}

/* ========== 标签页 ========== */
.wb-tabs {
  display: flex;
  border-bottom: none;
  background: #171717;
  padding-left: 16px;
}

.wb-tab {
  padding: 8px 16px 8px 0;
  cursor: pointer;
  color: #ffffff;
  transition: all 0.2s;
  position: relative;
  display: inline-flex;
  align-items: center;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    right: 16px;
    bottom: 0;
    height: 2px;
    background: transparent;
  }

  &:hover {
    color: #ffffff;
  }

  &.active {
    &::before {
      background: #007acc;
    }
  }
}

/* ========== 搜索栏 ========== */
.wb-search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: none;
  background: #171717;
}

.wb-search-input-wrapper {
  position: relative;
  flex: 1;
  max-width: 400px;
  height: 30px;
  background: #252526;
  border: 1px solid #252526;
  border-radius: 15px;
  transition: all 0.2s;

  &:has(.wb-search-input:focus) {
    border-color: #007acc;
  }
}

.wb-search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: #858585;
  font-size: 12px;
  pointer-events: none;
  z-index: 2;
}

.wb-search-input {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  padding: 0 36px;
  background: transparent !important;
  border: none !important;
  outline: none !important;
  box-shadow: none !important;
  color: #cccccc;
  font-size: 12px;

  &:focus {
    background: transparent !important;
    border: none !important;
    outline: none !important;
    box-shadow: none !important;
  }

  &::placeholder {
    color: #6e6e6e;
  }
}

.wb-search-clear {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: #858585;
  cursor: pointer;
  padding: 4px;
  z-index: 2;

  &:hover {
    color: #cccccc;
  }
}
</style>
