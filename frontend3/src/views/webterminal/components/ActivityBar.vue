<script setup lang="ts">
  import { Icon } from '@iconify/vue';

  interface Props {
    showSidebar?: boolean;
    isFullscreen?: boolean;
    activeActivityItem: string;
    serverMenuExpanded: boolean;
    serverMenuPosition: { top: number };
    mainMenuExpanded: boolean;
    mainMenuPosition: { top: number };
  }

  interface Emits {
    (e: 'toggle-sidebar'): void;
    (e: 'toggle-fullscreen'): void;
    (e: 'toggle-server-menu', event: MouseEvent): void;
    (e: 'toggle-main-menu', event: MouseEvent): void;
    (e: 'switch-activity-item', item: string): void;
  }

  defineProps<Props>();
  const emit = defineEmits<Emits>();
</script>

<template>
  <div class="wb-activity-bar">
    <!-- Logo -->
    <div class="wb-activity-logo">
      <Icon icon="lucide:terminal" class="wb-logo-icon" />
    </div>

    <!-- 菜单按钮 -->
    <div
      class="wb-activity-item"
      :class="{ active: mainMenuExpanded }"
      title="菜单"
      @click="emit('toggle-main-menu', $event)"
    >
      <Icon icon="lucide:menu" class="wb-activity-icon" />
      <div v-if="mainMenuExpanded" class="wb-menu-indicator"></div>
    </div>

    <!-- 主菜单下拉面板 -->
    <Transition name="wb-menu-dropdown">
      <div v-if="mainMenuExpanded" class="wb-main-menu-dropdown" :style="{ top: mainMenuPosition.top + 'px' }">
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'recent')">
          <Icon icon="lucide:history" class="wb-menu-item-icon" />
          <span>最近访问</span>
        </div>
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'sessions')">
          <Icon icon="lucide:list-tree" class="wb-menu-item-icon" />
          <span>会话列表</span>
        </div>
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'command-audit')">
          <Icon icon="lucide:terminal" class="wb-menu-item-icon" />
          <span>命令行审计</span>
        </div>
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'audit-list')">
          <Icon icon="lucide:file-text" class="wb-menu-item-icon" />
          <span>审计列表</span>
        </div>
        <div class="wb-menu-divider"></div>
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'settings')">
          <Icon icon="lucide:settings" class="wb-menu-item-icon" />
          <span>设置</span>
        </div>
      </div>
    </Transition>

    <!-- 服务器管理菜单 -->
    <div
      class="wb-activity-item"
      :class="{ active: activeActivityItem === 'servers' }"
      title="服务器管理"
      @click="emit('toggle-server-menu', $event)"
    >
      <Icon icon="lucide:server" class="wb-activity-icon" />
      <div v-if="serverMenuExpanded" class="wb-menu-indicator"></div>
    </div>

    <!-- 服务器菜单下拉面板 -->
    <Transition name="wb-menu-dropdown">
      <div v-if="serverMenuExpanded" class="wb-server-menu-dropdown" :style="{ top: serverMenuPosition.top + 'px' }">
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'servers')">
          <Icon icon="lucide:list-tree" class="wb-menu-item-icon" />
          <span>服务器列表</span>
        </div>
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'monitoring')">
          <Icon icon="lucide:activity" class="wb-menu-item-icon" />
          <span>监控面板</span>
        </div>
        <div class="wb-menu-item" @click="emit('switch-activity-item', 'alerts')">
          <Icon icon="lucide:bell" class="wb-menu-item-icon" />
          <span>告警管理</span>
        </div>
      </div>
    </Transition>

    <!-- 资产树折叠/展开按钮 -->
    <div
      class="wb-activity-item"
      :class="{ active: showSidebar }"
      :title="showSidebar ? '折叠资产树' : '展开资产树'"
      @click="emit('toggle-sidebar')"
    >
      <Icon v-if="showSidebar" icon="lucide:panel-left" class="wb-activity-icon" />
      <Icon v-else icon="lucide:panel-left-open" class="wb-activity-icon" />
    </div>

    <div
      class="wb-activity-item"
      :class="{ active: activeActivityItem === 'sessions' }"
      title="会话管理"
      @click="emit('switch-activity-item', 'sessions')"
    >
      <Icon icon="lucide:history" class="wb-activity-icon" />
    </div>
  </div>
</template>

<style lang="scss">
  .wb-activity-logo {
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 8px 0;
    flex-shrink: 0;
  }

  .wb-logo-icon {
    width: 24px;
    height: 24px;
    color: #007acc;
  }

  /* Activity Bar 图标样式 */
  .wb-activity-item svg {
    width: 20px;
    height: 20px;
  }

  .wb-activity-item svg,
  .wb-activity-item .iconify {
    color: #c5c5c5 !important;
    transition: color 0.15s ease;
  }

  .wb-activity-item:hover svg,
  .wb-activity-item:hover .iconify {
    color: #ffffff !important;
  }

  .wb-activity-item.active svg,
  .wb-activity-item.active .iconify {
    color: #ffffff !important;
  }

  .wb-activity-icon {
    color: #cccccc !important;
  }

  .wb-activity-item:hover .wb-activity-icon {
    color: #ffffff !important;
  }

  .wb-activity-item.active .wb-activity-icon {
    color: #ffffff !important;
  }

  /* 下拉菜单 */
  .wb-server-menu-dropdown,
  .wb-main-menu-dropdown {
    position: fixed;
    left: 48px;
    min-width: 140px;
    background: #212121;
    border: none;
    border-radius: 0;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    padding: 4px 0;
    z-index: 100;
  }

  .wb-menu-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    cursor: pointer;
    color: #cccccc;
    font-size: 12px;
    transition: background 0.15s;
  }

  .wb-menu-item:hover {
    background: #2d2d2d;
  }

  .wb-menu-divider {
    height: 1px;
    background: #3c3c3c;
    margin: 4px 0;
  }

  .wb-menu-item-icon {
    width: 14px;
    height: 14px;
    color: #858585;
  }

  .wb-menu-item:hover .wb-menu-item-icon {
    color: #cccccc;
  }

  .wb-menu-indicator {
    position: absolute;
    left: 0;
    top: 12px;
    bottom: 12px;
    width: 2px;
    background: #007acc;
  }

  /* 下拉菜单过渡动画 */
  .wb-menu-dropdown-enter-active,
  .wb-menu-dropdown-leave-active {
    transition: all 0.2s ease;
  }

  .wb-menu-dropdown-enter-from {
    opacity: 0;
    transform: translateX(-8px);
  }

  .wb-menu-dropdown-leave-to {
    opacity: 0;
    transform: translateX(-8px);
  }

  .wb-menu-dropdown-enter-to,
  .wb-menu-dropdown-leave-from {
    opacity: 1;
    transform: translateX(0);
  }
</style>
