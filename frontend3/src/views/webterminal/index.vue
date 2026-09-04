<script setup lang="ts">
  import { onMounted, onUnmounted, ref } from 'vue';
  import { useRoute } from 'vue-router';
  import { useSessionManager } from './composables/useSessionManager';
  import { useLayout } from './composables/useLayout';
  import SessionTabs from './components/SessionTabs.vue';
  import TerminalArea from './components/TerminalArea.vue';
  import AssetTree from './components/AssetTree.vue';
  import ActivityBar from './components/ActivityBar.vue';
  import ConnectDialog from './components/ConnectDialog.vue';
  import SessionList from './session_list.vue';

  defineOptions({ name: 'TerminalWorkbench' });

  const route = useRoute();

  // 会话管理
  const {
    sessions,
    activeSession,
    showConnectDialog,
    connectingServer,
    selectedCredentialId,
    currentSessionIds,
    switchSession,
    removeSession,
    cancelConnect,
    handleConnected,
    handleConnect,
    handleSessionDisconnected
  } = useSessionManager();

  // 布局管理
  const {
    showActivityBar,
    showSidebar,
    isFullscreen,
    sidebarWidth,
    isResizing,
    activeActivityItem,
    serverMenuExpanded,
    serverMenuPosition,
    mainMenuExpanded,
    mainMenuPosition,
    startResize,
    toggleSidebar,
    toggleFullscreen,
    toggleServerMenu,
    toggleMainMenu,
    handleClickOutside
  } = useLayout();

  // 当前路由hash
  const currentHash = ref('');

  function updateHash() {
    currentHash.value = window.location.hash.substring(1);
    if (currentHash.value === 'sessions') {
      activeActivityItem.value = 'sessions';
    } else if (currentHash.value === 'recent') {
      activeActivityItem.value = 'recent';
    } else if (currentHash.value === '') {
      activeActivityItem.value = 'assets';
    }
  }

  function switchActivityItem(item: string) {
    mainMenuExpanded.value = false;
    serverMenuExpanded.value = false;

    // 特殊处理：会话列表在应用标签页中打开
    if (item === 'sessions') {
      const sessionListSession: Bastion.TerminalSession = {
        id: 'session-list',
        serverId: 0,
        serverName: '会话列表',
        serverIp: '',
        loginAccount: '',
        protocol: 'view',
        status: 'connected',
        connected: true,
        duration: 0,
        startedAt: new Date().toISOString(),
        isSessionListView: true,
        title: '会话列表'
      };

      const existingSession = sessions.value.find(s => s.id === 'session-list');
      if (existingSession) {
        activeSession.value = existingSession;
      } else {
        sessions.value.push(sessionListSession);
        activeSession.value = sessions.value[sessions.value.length - 1];
      }
      return;
    }

    activeActivityItem.value = item;
  }

  // 清理
  onUnmounted(() => {
    window.removeEventListener('hashchange', updateHash);
  });

  // 组件挂载时检查 URL 参数
  onMounted(async () => {
    updateHash();
    window.addEventListener('hashchange', updateHash);

    const query = route.query;
    if (query.serverId && query.hostname) {
      const basicServerInfo: Bastion.BasicServerInfo = {
        id: Number(query.serverId),
        hostname: query.hostname as string,
        ip: query.ip as string,
        env: query.env as string,
        agentStatus: query.agentStatus as string
      };
      await handleConnect(basicServerInfo);
    }

    document.addEventListener('click', handleClickOutside);
  });
</script>

<template>
  <div class="terminal-workbench" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- 活动栏 (Activity Bar) -->
    <ActivityBar
      v-if="showActivityBar && !isFullscreen"
      :show-sidebar="showSidebar && !isFullscreen"
      :is-fullscreen="isFullscreen"
      :active-activity-item="activeActivityItem"
      :server-menu-expanded="serverMenuExpanded"
      :server-menu-position="serverMenuPosition"
      :main-menu-expanded="mainMenuExpanded"
      :main-menu-position="mainMenuPosition"
      @toggle-sidebar="toggleSidebar"
      @toggle-fullscreen="toggleFullscreen"
      @toggle-server-menu="toggleServerMenu"
      @toggle-main-menu="toggleMainMenu"
      @switch-activity-item="switchActivityItem"
    />

    <!-- 主内容区 -->
    <div class="wb-main-content">
      <!-- 会话列表页面 -->
      <div v-if="activeActivityItem === 'sessions'" class="wb-full-page">
        <SessionList />
      </div>

      <!-- 主机资产和终端页面 -->
      <template v-else>
        <!-- 侧边栏 -->
        <div v-if="showSidebar && !isFullscreen" class="wb-sidebar" :style="{ width: sidebarWidth + 'px' }">
          <AssetTree :current-sessions="currentSessionIds" @connect="handleConnect" />
        </div>

        <!-- 分割线 -->
        <div
          v-if="showSidebar && !isFullscreen"
          class="wb-sash"
          :class="{ 'wb-sash-resizing': isResizing }"
          @mousedown="startResize"
        >
          <div class="wb-sash-line"></div>
        </div>

        <!-- 右侧内容区 -->
        <div class="wb-content-area">
          <SessionTabs
            :sessions="sessions"
            :active-id="activeSession?.id || null"
            :current-session-ids="currentSessionIds"
            :show-sidebar="showSidebar && !isFullscreen"
            @select="switchSession"
            @remove="removeSession"
            @connect="handleConnect"
            @toggle-sidebar="toggleSidebar"
            @toggle-fullscreen="toggleFullscreen"
          />
          <div class="wb-terminal-area-wrapper">
            <TerminalArea
              :active-session="activeSession"
              :sessions="sessions"
              @session-disconnected="handleSessionDisconnected"
            />
          </div>
        </div>
      </template>
    </div>

    <!-- 连接对话框 -->
    <ConnectDialog
      v-model="showConnectDialog"
      :connecting-server="connectingServer"
      :selected-credential-id="selectedCredentialId"
      @update:selected-credential-id="selectedCredentialId = $event"
      @connect="handleConnected"
      @cancel="cancelConnect"
    />
  </div>
</template>

<style lang="scss">
  @use '@/styles/scss/terminal-workbench.scss' as *;
  @import '@vscode/codicons/dist/codicon.css';

  /* 覆盖 Element Plus 树节点的悬停背景色变量。
     注意：本 style 块非 scoped，:deep() 编译后选择器永不匹配（历史写法全部失效），
     必须用全局选择器；悬停色用 VS Code 风格微亮白，与深色工作台协调 */
  .terminal-workbench .el-tree {
    --el-tree-node-hover-bg-color: rgb(255 255 255 / 8%) !important;
    --el-fill-color-light: rgb(0 0 0 / 20%) !important;
    --el-fill-color-lighter: rgb(0 0 0 / 20%) !important;
    --el-fill-color-extra-light: rgb(0 0 0 / 15%) !important;
    --el-fill-color: rgb(0 0 0 / 20%) !important;
    --el-fill-color-dark: rgb(0 0 0 / 30%) !important;
  }

  .terminal-workbench .el-tree-node__content {
    background-color: transparent !important;
  }

  .terminal-workbench .el-tree-node__content:hover {
    background-color: rgb(255 255 255 / 8%) !important;
  }

  .terminal-workbench.is-fullscreen {
    position: fixed;
    inset: 0;
    z-index: 9999;
  }

  .wb-main-content {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .wb-content-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
    background: #171717;
  }

  .wb-terminal-area-wrapper {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-height: 0;
    background: #171717;
  }

  .wb-full-page {
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  /* 强制覆盖所有按钮和图标的悬停颜色 */
  .terminal-workbench :deep(.el-button) {
    color: #858585 !important;
  }

  .terminal-workbench :deep(.el-button .iconify) {
    color: #858585 !important;
  }

  .terminal-workbench :deep(.el-button:hover) {
    background: rgb(0 0 0 / 20%) !important;
    color: #aaa !important;
  }

  .terminal-workbench :deep(.el-button:hover .iconify) {
    color: #aaa !important;
  }

  .terminal-workbench :deep(.el-button.is-link:hover) {
    background-color: rgb(0 0 0 / 20%) !important;
  }

  /* 对话框和 Select 组件的暗色主题覆盖样式已在 ConnectDialog.vue 中定义 */
</style>
