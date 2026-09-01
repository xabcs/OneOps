<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { onKeyStroke } from '@vueuse/core';
import {
  ElButton,
  ElDialog,
  ElNotification,
  ElOption,
  ElSelect,
  ElTag
} from 'element-plus';
import { Icon } from '@iconify/vue';
import { fetchConnectServer, fetchGetServerForConnect } from '@/service/api';
import SessionTabs from './components/SessionTabs.vue';
import TerminalArea from './components/TerminalArea.vue';
import AssetTree from './components/AssetTree.vue';
import SessionList from './session_list.vue';

defineOptions({ name: 'TerminalWorkbench' });

const route = useRoute();
const router = useRouter();

// 会话管理
const sessions = ref<any[]>([]);
const activeSession = ref<any>(null);

// 布局状态
const showActivityBar = ref(true);
const showSidebar = ref(true);
const isFullscreen = ref(false);
const sidebarWidth = ref(300);
const isResizing = ref(false);
const activeActivityItem = ref('assets');

// 当前路由hash
const currentHash = ref('');

// 监听路由hash变化
function updateHash() {
  currentHash.value = window.location.hash.substring(1); // 移除 #
  if (currentHash.value === 'sessions') {
    activeActivityItem.value = 'sessions';
  } else if (currentHash.value === 'recent') {
    activeActivityItem.value = 'recent';
  } else if (currentHash.value === '') {
    activeActivityItem.value = 'assets';
  }
}

// 连接对话框
const showConnectDialog = ref(false);
const connectingServer = ref<any>(null);
const selectedCredentialId = ref<number | null>(null);

// 监听 ESC 键退出全屏
onKeyStroke('Escape', () => {
  if (isFullscreen.value) {
    toggleFullscreen();
  }
});

// 开始调整侧边栏大小
function startResize(e: MouseEvent) {
  isResizing.value = true;
  const startX = e.clientX;
  const startWidth = sidebarWidth.value;

  const onMouseMove = (e: MouseEvent) => {
    if (!isResizing.value) return;
    const diff = e.clientX - startX;
    const newWidth = Math.max(200, Math.min(600, startWidth + diff));
    sidebarWidth.value = newWidth;
  };

  const onMouseUp = () => {
    isResizing.value = false;
    window.removeEventListener('mousemove', onMouseMove);
    window.removeEventListener('mouseup', onMouseUp);
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  };

  window.addEventListener('mousemove', onMouseMove);
  window.addEventListener('mouseup', onMouseUp);
  document.body.style.cursor = 'col-resize';
  document.body.style.userSelect = 'none';
}

// 清理
onUnmounted(() => {
  if (isResizing.value) {
    isResizing.value = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  }

  // 移除点击外部监听器
  document.removeEventListener('click', handleClickOutside);

  // 移除hash变化监听器
  window.removeEventListener('hashchange', updateHash);
});

// 当前会话的 serverId 列表
const currentSessionIds = computed(() => sessions.value.map((s: any) => s.serverId));
// 切换侧边栏显示
function toggleSidebar() {
  showSidebar.value = !showSidebar.value;
}

// 服务器菜单折叠状态
const serverMenuExpanded = ref(false);
const serverMenuPosition = ref({ top: 0 });

// 主菜单折叠状态
const mainMenuExpanded = ref(false);
const mainMenuPosition = ref({ top: 0 });

// 切换服务器菜单展开/折叠
function toggleServerMenu(event: MouseEvent) {
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  serverMenuPosition.value = { top: rect.top };

  serverMenuExpanded.value = !serverMenuExpanded.value;

  // 关闭主菜单
  mainMenuExpanded.value = false;
}

// 切换主菜单展开/折叠
function toggleMainMenu(event: MouseEvent) {
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  mainMenuPosition.value = { top: rect.top };

  mainMenuExpanded.value = !mainMenuExpanded.value;

  // 关闭服务器菜单
  serverMenuExpanded.value = false;
}

// 点击外部关闭下拉菜单
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement;
  const activityBar = document.querySelector('.wb-activity-bar');
  const serverDropdown = document.querySelector('.wb-server-menu-dropdown');
  const mainDropdown = document.querySelector('.wb-main-menu-dropdown');

  if (activityBar && !activityBar.contains(target)) {
    if (serverDropdown && !serverDropdown.contains(target)) {
      serverMenuExpanded.value = false;
    }
    if (mainDropdown && !mainDropdown.contains(target)) {
      mainMenuExpanded.value = false;
    }
  }
}

// 切换全屏
function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value;
}

// 取消连接
function cancelConnect() {
  showConnectDialog.value = false;
  connectingServer.value = null;
  // 清除 URL 查询参数
  if (Object.keys(route.query).length > 0) {
    router.replace({ path: route.path, query: {} });
  }
}

// 连接按钮点击
async function handleConnected() {
  if (!connectingServer.value) {
    console.error('connectingServer.value 为空');
    return;
  }

  try {
    console.log('=== 点击连接按钮 ===');
    console.log('服务器信息:', connectingServer.value);
    console.log('选择的凭证ID:', selectedCredentialId.value);
    console.log('sshCredentialId:', connectingServer.value.sshCredentialId);
    console.log('credentials:', connectingServer.value.credentials);

    // 使用选择的凭证ID
    // 确保有选择凭证
    if (!selectedCredentialId.value) {
      ElNotification.error({
        title: '连接失败',
        message: '请先选择连接凭证'
      });
      return;
    }

    // 查找选择的凭证信息，获取实际的用户名
    const selectedCredential = connectingServer.value.credentials.find((c: any) => c.id === selectedCredentialId.value);
    const loginAccount = selectedCredential?.username || 'root';

    console.log('实际使用的凭证ID:', selectedCredentialId.value);
    console.log('登录账号:', loginAccount);

    // 调用后端接口创建 SSH 会话
    const response = await fetchConnectServer(connectingServer.value.id, {
      protocol: 'ssh',
      credentialId: selectedCredentialId.value
    });

    console.log('连接 API 响应:', response);

    if (response.data && response.data.sessionId) {
      const sessionId = response.data.sessionId;
      const websocketUrl = response.data.websocketUrl;

      // 创建会话对象 - loginAccount 从选择的凭证中获取
      const newSession = {
        id: sessionId,
        serverId: connectingServer.value.id,
        serverName: connectingServer.value.hostname || 'Unknown',
        serverIp: connectingServer.value.ip || 'Unknown',
        loginAccount,
        protocol: 'ssh',
        status: 'connected',
        connected: true,
        duration: 0,
        startedAt: new Date().toISOString(),
        websocketUrl
      };

      sessions.value.push(newSession);
      activeSession.value = sessions.value[sessions.value.length - 1];
      showConnectDialog.value = false;
      connectingServer.value = null;

      // 调试信息
      console.log('=== 连接成功 ===');
      console.log('新会话:', newSession);
      console.log('当前所有会话:', sessions.value);
      console.log('活动会话:', activeSession.value);

      // 清除 URL 查询参数
      if (Object.keys(route.query).length > 0) {
        router.replace({ path: route.path, query: {} });
      }
    } else {
      ElNotification.error({
        title: '连接失败',
        message: '未能创建 SSH 会话'
      });
    }
  } catch (error) {
    console.error('连接服务器失败:', error);
    ElNotification.error({
      title: '连接失败',
      message: error instanceof Error ? error.message : '连接服务器时发生错误'
    });
  }
}

// 切换会话
function switchSession(sessionId: number) {
  activeSession.value = sessions.value.find((s: any) => s.id === sessionId) || null;
}

// 移除会话
function removeSession(sessionId: number) {
  const index = sessions.value.findIndex((s: any) => s.id === sessionId);
  if (index > -1) {
    sessions.value.splice(index, 1);
    if (activeSession.value?.id === sessionId) {
      activeSession.value = sessions.value[0] || null;
    }
  }
}

// 处理主机连接（统一入口）
async function handleConnect(server: any) {
  console.log('=== handleConnect 被调用 ===');
  console.log('传入的服务器对象:', server);
  console.log('服务器ID:', server.id);

  // 获取完整的服务器信息（包含 credentials）
  try {
    // 显示加载提示（可选，体验更好）
    // ElMessage.info({
    //     message: "正在连接...",
    //     duration: 500,
    // });

    const response = await fetchGetServerForConnect(server.id);
    console.log('fetchGetServerForConnect 响应:', response);

    if (response.data) {
      connectingServer.value = response.data;
      console.log('获取服务器信息成功:', connectingServer.value);
      console.log('credentials:', connectingServer.value.credentials);

      // 默认选中第一个凭证（如果有）
      if (connectingServer.value.credentials && connectingServer.value.credentials.length > 0) {
        selectedCredentialId.value = connectingServer.value.credentials[0].id;
      } else {
        selectedCredentialId.value = null;
      }
      showConnectDialog.value = true;
    } else {
      console.error('响应中没有 data 字段');
      ElNotification.error({
        title: '获取服务器信息失败',
        message: '服务器不存在或无权访问'
      });
    }
  } catch (error) {
    console.error('获取服务器信息失败:', error);
    ElNotification.error({
      title: '获取服务器信息失败',
      message: error instanceof Error ? error.message : '未知错误'
    });
  }
}
function switchActivityItem(item: string) {
  // 关闭所有下拉菜单
  mainMenuExpanded.value = false;
  serverMenuExpanded.value = false;

  // 特殊处理：会话列表在应用标签页中打开
  if (item === 'sessions') {
    // 创建一个特殊的"会话列表"会话对象
    const sessionListSession = {
      id: 'session-list',
      serverId: 0,
      serverName: '会话列表',
      serverIp: '',
      loginAccount: '',
      protocol: 'view' as any,
      status: 'connected',
      connected: true,
      duration: 0,
      startedAt: new Date().toISOString(),
      isSessionListView: true, // 标记为会话列表视图
      title: '会话列表'
    };

    // 检查是否已存在会话列表标签
    const existingSession = sessions.value.find((s: any) => s.id === 'session-list');
    if (existingSession) {
      // 已存在，切换到该标签
      activeSession.value = existingSession;
    } else {
      // 不存在，添加新标签
      sessions.value.push(sessionListSession);
      activeSession.value = sessions.value[sessions.value.length - 1];
    }
    return;
  }

  activeActivityItem.value = item;
}

// 组件挂载时检查 URL 参数
onMounted(async () => {
  // 初始化hash
  updateHash();

  // 监听hash变化
  window.addEventListener('hashchange', updateHash);

  const query = route.query;
  if (query.serverId && query.hostname) {
    // 从 URL 参数构建基本的服务器对象，然后调用 handleConnect
    const basicServerInfo = {
      id: Number(query.serverId),
      hostname: query.hostname as string,
      ip: query.ip as string,
      env: query.env as string,
      agentStatus: query.agentStatus as string
    };
    await handleConnect(basicServerInfo);
  }

  // 添加点击外部关闭下拉菜单的监听器
  document.addEventListener('click', handleClickOutside);
});
</script>

<template>
  <div class="terminal-workbench" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- 活动栏 (Activity Bar) -->
    <div v-if="showActivityBar && !isFullscreen" class="wb-activity-bar">
      <!-- Logo -->
      <div class="wb-activity-logo">
        <Icon icon="lucide:terminal" class="wb-logo-icon" />
      </div>

      <!-- 菜单按钮 -->
      <div class="wb-activity-item" :class="{ active: mainMenuExpanded }" title="菜单" @click="toggleMainMenu">
        <Icon icon="lucide:menu" class="wb-activity-icon" />
        <!-- 折叠指示器 -->
        <div v-if="mainMenuExpanded" class="wb-menu-indicator"></div>
      </div>

      <!-- 主菜单下拉面板 -->
      <Transition name="wb-menu-dropdown">
        <div v-if="mainMenuExpanded" class="wb-main-menu-dropdown" :style="{ top: mainMenuPosition.top + 'px' }">
          <div class="wb-menu-item" @click="switchActivityItem('recent')">
            <Icon icon="lucide:history" class="wb-menu-item-icon" />
            <span>最近访问</span>
          </div>
          <div class="wb-menu-item" @click="switchActivityItem('sessions')">
            <Icon icon="lucide:list-tree" class="wb-menu-item-icon" />
            <span>会话列表</span>
          </div>
          <div class="wb-menu-item" @click="switchActivityItem('command-audit')">
            <Icon icon="lucide:terminal" class="wb-menu-item-icon" />
            <span>命令行审计</span>
          </div>
          <div class="wb-menu-item" @click="switchActivityItem('audit-list')">
            <Icon icon="lucide:file-text" class="wb-menu-item-icon" />
            <span>审计列表</span>
          </div>
          <div class="wb-menu-divider"></div>
          <div class="wb-menu-item" @click="switchActivityItem('settings')">
            <Icon icon="lucide:settings" class="wb-menu-item-icon" />
            <span>设置</span>
          </div>
        </div>
      </Transition>

      <!-- 新增：服务器管理菜单 -->
      <div
        class="wb-activity-item"
        :class="{ active: activeActivityItem === 'servers' }"
        title="服务器管理"
        @click="toggleServerMenu"
      >
        <Icon icon="lucide:server" class="wb-activity-icon" />
        <!-- 折叠指示器 -->
        <div v-if="serverMenuExpanded" class="wb-menu-indicator"></div>
      </div>

      <!-- 服务器菜单下拉面板 -->
      <Transition name="wb-menu-dropdown">
        <div v-if="serverMenuExpanded" class="wb-server-menu-dropdown" :style="{ top: serverMenuPosition.top + 'px' }">
          <div class="wb-menu-item" @click="switchActivityItem('servers')">
            <Icon icon="lucide:list-tree" class="wb-menu-item-icon" />
            <span>服务器列表</span>
          </div>
          <div class="wb-menu-item" @click="switchActivityItem('monitoring')">
            <Icon icon="lucide:activity" class="wb-menu-item-icon" />
            <span>监控面板</span>
          </div>
          <div class="wb-menu-item" @click="switchActivityItem('alerts')">
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
        @click="toggleSidebar"
      >
        <Icon v-if="showSidebar" icon="lucide:panel-left" class="wb-activity-icon" />
        <Icon v-else icon="lucide:panel-left-open" class="wb-activity-icon" />
      </div>

      <div
        class="wb-activity-item"
        :class="{ active: activeActivityItem === 'sessions' }"
        title="会话管理"
        @click="switchActivityItem('sessions')"
      >
        <Icon icon="lucide:history" class="wb-activity-icon" />
      </div>
    </div>

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
            <TerminalArea :active-session="activeSession" :sessions="sessions" />
          </div>
        </div>
      </template>
    </div>

    <!-- 连接对话框 -->
    <ElDialog
      v-model="showConnectDialog"
      title="连接主机"
      width="520px"
      :close-on-click-modal="true"
      :append-to-body="false"
      class="terminal-connect-dialog"
      @close="cancelConnect"
    >
      <div v-if="connectingServer" class="connect-dialog-content">
        <!-- 主机信息 -->
        <div class="connect-section">
          <div class="connect-section-title">主机信息</div>
          <div class="connect-info-grid">
            <div class="connect-info-item">
              <span class="connect-info-label">主机名称</span>
              <span class="connect-info-value">{{ connectingServer.hostname || '未知' }}</span>
            </div>
            <div class="connect-info-item">
              <span class="connect-info-label">IP地址</span>
              <span class="connect-info-value">{{ connectingServer.ip || '未知' }}</span>
            </div>
            <div class="connect-info-item">
              <span class="connect-info-label">环境</span>
              <span class="connect-info-value">{{ connectingServer.env || 'unknown' }}</span>
            </div>
            <div class="connect-info-item">
              <span class="connect-info-label">Agent状态</span>
              <span class="connect-info-value">
                <ElTag v-if="connectingServer.agentStatus === 'running'" type="success" size="small">在线</ElTag>
                <ElTag v-else-if="connectingServer.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
                <ElTag v-else type="info" size="small">未安装</ElTag>
              </span>
            </div>
          </div>
        </div>

        <!-- 安全上下文 -->
        <div class="connect-section">
          <div class="connect-section-title">安全上下文</div>
          <div class="connect-info-grid">
            <div class="connect-info-item">
              <span class="connect-info-label">认证方式</span>
              <span class="connect-info-value">
                <Icon icon="lucide:shield-check" style="width: 14px; height: 14px; margin-right: 4px; color: #4ec9b0" />
                密钥认证
              </span>
            </div>
            <div class="connect-info-item">
              <span class="connect-info-label">安全策略</span>
              <span class="connect-info-value">
                <ElTag type="success" size="small" effect="plain">允许访问</ElTag>
              </span>
            </div>
            <div class="connect-info-item">
              <span class="connect-info-label">会话记录</span>
              <span class="connect-info-value">
                <Icon
                  icon="lucide:check-circle-2"
                  style="width: 14px; height: 14px; margin-right: 4px; color: #4ec9b0"
                />
                已启用
              </span>
            </div>
          </div>
        </div>

        <!-- 凭证选择 -->
        <div class="connect-section">
          <div class="connect-section-title">选择连接凭证</div>
          <div
            v-if="connectingServer.credentials && connectingServer.credentials.length > 0"
            class="credential-selector"
          >
            <ElSelect
              v-model="selectedCredentialId"
              placeholder="请选择凭证"
              popper-class="terminal-select-dropdown"
              style="width: 100%"
            >
              <ElOption
                v-for="cred in connectingServer.credentials"
                :key="cred.id"
                :label="`${cred.name} - ${cred.username}`"
                :value="cred.id"
              >
                <span style="display: flex; align-items: center; gap: 8px">
                  <Icon icon="lucide:key" style="width: 14px; height: 14px; color: #858585" />
                  <span>{{ cred.name }}</span>
                  <ElTag size="small" effect="plain" style="margin-left: auto">{{ cred.username }}</ElTag>
                </span>
              </ElOption>
            </ElSelect>
            <div class="credential-hint">
              <Icon icon="lucide:info" style="width: 14px; height: 14px; margin-right: 4px; color: #858585" />
              <span>选择的凭证将决定登录账号</span>
            </div>
          </div>
          <div v-else class="credential-empty">
            <Icon icon="lucide:alert-circle" style="width: 16px; height: 16px; color: #f14c4c" />
            <span>该服务器没有可用的连接凭证</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="connect-actions">
          <ElButton @click="cancelConnect">取消</ElButton>
          <ElButton type="primary" :disabled="!selectedCredentialId" @click="handleConnected()">
            <Icon icon="lucide:terminal" style="margin-right: 6px; width: 14px; height: 14px" />
            连接
          </ElButton>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<style lang="scss">
@import '@/styles/scss/terminal-workbench.scss';
@import '@vscode/codicons/dist/codicon.css';

/* 覆盖 Element Plus 树节点的悬停背景色变量。
   注意：本 style 块非 scoped，:deep() 编译后选择器永不匹配（历史写法全部失效），
   必须用全局选择器；悬停色用 VS Code 风格微亮白，与深色工作台协调 */
.terminal-workbench .el-tree {
  --el-tree-node-hover-bg-color: rgba(255, 255, 255, 0.08) !important;
  --el-fill-color-light: rgba(0, 0, 0, 0.2) !important;
  --el-fill-color-lighter: rgba(0, 0, 0, 0.2) !important;
  --el-fill-color-extra-light: rgba(0, 0, 0, 0.15) !important;
  --el-fill-color: rgba(0, 0, 0, 0.2) !important;
  --el-fill-color-dark: rgba(0, 0, 0, 0.3) !important;
}

.terminal-workbench .el-tree-node__content {
  background-color: transparent !important;
}

.terminal-workbench .el-tree-node__content:hover {
  background-color: rgba(255, 255, 255, 0.08) !important;
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

/* Activity Bar 图标样式 - Icon 组件 */
.wb-activity-item svg {
  width: 20px;
  height: 20px;
}

/* 强制覆盖所有按钮和图标的悬停颜色 */
.terminal-workbench :deep(.el-button) {
  color: #858585 !important;
}

.terminal-workbench :deep(.el-button .iconify) {
  color: #858585 !important;
}

.terminal-workbench :deep(.el-button:hover) {
  background: rgba(0, 0, 0, 0.2) !important;
  color: #aaaaaa !important;
}

.terminal-workbench :deep(.el-button:hover .iconify) {
  color: #aaaaaa !important;
}

.terminal-workbench :deep(.el-button.is-link:hover) {
  background-color: rgba(0, 0, 0, 0.2) !important;
}

/* Activity Bar 图标样式 - 参考阿里云设计 */
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

.wb-activity-bottom {
  margin-top: auto;
  flex-shrink: 0;
}

/* 增强 Activity Icon 可见性 */
.wb-activity-icon {
  color: #cccccc !important;
}

.wb-activity-item:hover .wb-activity-icon {
  color: #ffffff !important;
}

.wb-activity-item.active .wb-activity-icon {
  color: #ffffff !important;
}

/* 服务器菜单下拉面板 */
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

/* 折叠指示器 */
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

/* 连接对话框样式 - 终端暗色主题 */
.terminal-workbench :deep(.el-overlay) {
  background-color: rgba(0, 0, 0, 0.7) !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog) {
  background: #252526 !important;
  border: 1px solid #454545 !important;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.6) !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__header) {
  background: #2d2d2d !important;
  border-bottom: 1px solid #000000 !important;
  padding: 12px 16px !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__title) {
  color: #cccccc !important;
  font-size: 13px !important;
  font-weight: 500 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__headerbtn .el-dialog__close) {
  color: #858585 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__headerbtn .el-dialog__close:hover) {
  color: #cccccc !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__body) {
  background: #252526 !important;
  padding: 16px !important;
  color: #cccccc !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-dialog__footer) {
  background: #2d2d2d !important;
  border-top: 1px solid #000000 !important;
  padding: 12px 16px !important;
}

/* 对话框中的标签样式 */
.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag) {
  background: rgba(78, 201, 176, 0.1) !important;
  border-color: transparent !important;
  color: #4ec9b0 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--success) {
  background: rgba(78, 201, 176, 0.1) !important;
  color: #4ec9b0 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--warning) {
  background: rgba(217, 119, 6, 0.1) !important;
  color: #d97706 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--info) {
  background: rgba(107, 114, 128, 0.1) !important;
  color: #6b7280 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-tag.el-tag--plain) {
  background: rgba(78, 201, 176, 0.15) !important;
  border: 1px solid rgba(78, 201, 176, 0.3) !important;
}

/* 对话框中的按钮样式 */
.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button) {
  background: transparent !important;
  color: #cccccc !important;
  border: 1px solid #454545 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button:hover) {
  background: rgba(0, 0, 0, 0.2) !important;
  border-color: #555 !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button--primary) {
  background: #007acc !important;
  border-color: #007acc !important;
  color: #ffffff !important;
}

.terminal-workbench :deep(.el-dialog.terminal-connect-dialog .el-button--primary:hover) {
  background: #0069b4 !important;
  border-color: #0069b4 !important;
}

.connect-dialog-content {
  padding: 0;
}

.connect-section {
  margin-bottom: 20px;
}

.connect-section:last-child {
  margin-bottom: 0;
}

.connect-section-title {
  font-size: 12px;
  font-weight: 500;
  color: #cccccc;
  margin-bottom: 12px;
  padding-left: 12px;
  position: relative;
}

.connect-section-title::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 14px;
  background: #007acc;
  border-radius: 2px;
}

.connect-info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px 16px;
  padding: 0 4px;
}

.connect-info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.connect-info-label {
  font-size: 11px;
  color: #858585;
  line-height: 1.5;
}

.connect-info-value {
  font-size: 12px;
  color: #cccccc;
  line-height: 1.5;
  display: flex;
  align-items: center;
}

/* 对话框按钮样式 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 按钮操作区域 */
.connect-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #333;
}

/* 凭证选择器 */
.credential-selector {
  padding: 0 4px;
}

.credential-hint {
  display: flex;
  align-items: center;
  margin-top: 8px;
  padding: 8px 12px;
  font-size: 11px;
  color: #858585;
  background: rgba(136, 85, 85, 0.1);
  border-radius: 2px;
}

.credential-empty {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  color: #f14c4c;
  font-size: 12px;
  background: rgba(241, 76, 76, 0.1);
  border: 1px dashed #f14c4c;
  border-radius: 2px;
}

/* ==================== Select 下拉弹出层样式（使用 popper-class） ==================== */

/* 下拉框主容器 */
.terminal-select-dropdown {
  background: #252526 !important;
  border: 1px solid #454545 !important;
  border-radius: 0 !important;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5) !important;
}

/* 下拉框内部包装器 */
.terminal-select-dropdown .el-select-dropdown__wrap {
  background: #252526 !important;
}

/* 下拉框列表 */
.terminal-select-dropdown .el-select-dropdown__list {
  background: #252526 !important;
  padding: 4px 0 !important;
}

/* 下拉选项 */
.terminal-select-dropdown .el-select-dropdown__item {
  background: transparent !important;
  color: #cccccc !important;
  padding: 8px 12px !important;
  font-size: 13px !important;
  border-radius: 0 !important;
  display: flex !important;
  align-items: center !important;
  min-height: 36px !important;
}

/* 下拉选项悬停/选中状态 */
.terminal-select-dropdown .el-select-dropdown__item.hover,
.terminal-select-dropdown .el-select-dropdown__item:hover {
  background: rgba(0, 0, 0, 0.3) !important;
  color: #e0e0e0 !important;
}

.terminal-select-dropdown .el-select-dropdown__item.is-selected {
  background: rgba(78, 201, 176, 0.15) !important;
  color: #4ec9b0 !important;
}

.terminal-select-dropdown .el-select-dropdown__item.is-selected:hover {
  background: rgba(78, 201, 176, 0.25) !important;
}

/* 下拉选项中的 Icon 样式 */
.terminal-select-dropdown .el-select-dropdown__item .iconify {
  color: #858585 !important;
}

.terminal-select-dropdown .el-select-dropdown__item:hover .iconify {
  color: #aaaaaa !important;
}

/* 下拉选项中的 Tag 样式 */
.terminal-select-dropdown .el-tag {
  background: rgba(78, 201, 176, 0.12) !important;
  border-color: rgba(78, 201, 176, 0.3) !important;
  color: #4ec9b0 !important;
  border-radius: 0 !important;
}

/* Popper 箭头样式 */
.terminal-select-dropdown .el-popper__arrow::before {
  background: #252526 !important;
  border-color: #454545 !important;
}

/* 下拉滚动条样式 */
.terminal-select-dropdown .el-select-dropdown__listbar::-webkit-scrollbar {
  width: 8px;
}

.terminal-select-dropdown .el-select-dropdown__listbar::-webkit-scrollbar-track {
  background: #1e1e1e !important;
}

.terminal-select-dropdown .el-select-dropdown__listbar::-webkit-scrollbar-thumb {
  background: #555 !important;
  border-radius: 0 !important;
}

.terminal-select-dropdown .el-select-dropdown__listbar::-webkit-scrollbar-thumb:hover {
  background: #666 !important;
}

/* Select 输入框箭头颜色 */
.el-dialog.terminal-connect-dialog .el-select .el-select__caret {
  color: #858585 !important;
}

.el-dialog.terminal-connect-dialog .el-select:hover .el-select__caret {
  color: #cccccc !important;
}

/* ==================== Select 选择框样式 - 输入框本身 ==================== */

/* Select 输入框包装器 - 针对 .el-select__wrapper */
.el-dialog.terminal-connect-dialog .el-select .el-select__wrapper,
.terminal-connect-dialog .el-select .el-select__wrapper {
  background: #1e1e1e !important;
  border: 1px solid #454545 !important;
  border-radius: 0 !important;
  box-shadow: none !important;
}

/* Select 输入框悬停状态 */
.el-dialog.terminal-connect-dialog .el-select:hover .el-select__wrapper,
.terminal-connect-dialog .el-select:hover .el-select__wrapper {
  border-color: #555 !important;
}

/* Select 输入框聚焦状态 */
.el-dialog.terminal-connect-dialog .el-select.is-focused .el-select__wrapper,
.terminal-connect-dialog .el-select.is-focused .el-select__wrapper {
  border-color: #007acc !important;
  box-shadow: none !important;
}

/* Select 内部输入元素 */
.el-dialog.terminal-connect-dialog .el-select .el-input__inner,
.terminal-connect-dialog .el-select .el-input__inner {
  background: #1e1e1e !important;
  color: #cccccc !important;
  border: none !important;
}

/* Select placeholder 样式 */
.el-dialog.terminal-connect-dialog .el-select .el-input__inner::placeholder,
.terminal-connect-dialog .el-select .el-input__inner::placeholder,
.el-dialog.terminal-connect-dialog .el-select__placeholder,
.terminal-connect-dialog .el-select__placeholder {
  color: #6e6e6e !important;
}

/* Select 箭头图标颜色 */
.el-dialog.terminal-connect-dialog .el-select .el-select__caret,
.terminal-connect-dialog .el-select .el-select__caret {
  color: #858585 !important;
}

.el-dialog.terminal-connect-dialog .el-select:hover .el-select__caret,
.terminal-connect-dialog .el-select:hover .el-select__caret {
  color: #cccccc !important;
}

/* ==================== 连接对话框 - 终端暗色主题（无圆角） ==================== */
/* 遮罩层 */
.el-overlay {
  background-color: rgba(0, 0, 0, 0.7) !important;
}

/* 对话框容器 */
.el-dialog.terminal-connect-dialog {
  background: #252526 !important;
  border: 1px solid #454545 !important;
  border-radius: 0 !important;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.6) !important;
  padding: 0 !important;
}

/* 对话框头部 */
.el-dialog.terminal-connect-dialog .el-dialog__header {
  background: #252526 !important;
  border-bottom: 1px solid #333 !important;
  border-radius: 0 !important;
  padding: 8px 16px !important;
  margin: 0 !important;
  height: 36px !important;
  display: flex !important;
  align-items: center !important;
}

.el-dialog.terminal-connect-dialog .el-dialog__title {
  color: #cccccc !important;
  font-size: 13px !important;
  font-weight: 400 !important;
  line-height: 20px !important;
}

.el-dialog.terminal-connect-dialog .el-dialog__headerbtn {
  top: 50% !important;
  transform: translateY(-50%) !important;
  right: 16px !important;
  width: 28px !important;
  height: 28px !important;
}

.el-dialog.terminal-connect-dialog .el-dialog__headerbtn .el-dialog__close {
  color: #858585 !important;
  font-size: 16px !important;
}

.el-dialog.terminal-connect-dialog .el-dialog__headerbtn:hover .el-dialog__close {
  color: #cccccc !important;
}

/* 对话框主体 */
.el-dialog.terminal-connect-dialog .el-dialog__body {
  background: #1e1e1e !important;
  padding: 20px 24px !important;
  color: #cccccc !important;
  border-radius: 0 !important;
}

/* 隐藏对话框底部 */
.el-dialog.terminal-connect-dialog .el-dialog__footer {
  display: none !important;
}

/* 标签样式 - 无圆角 */
.el-dialog.terminal-connect-dialog .el-tag {
  background: rgba(78, 201, 176, 0.12) !important;
  border-color: transparent !important;
  color: #4ec9b0 !important;
  border-radius: 0 !important;
  padding: 2px 8px !important;
  font-size: 11px !important;
  height: 20px !important;
  line-height: 16px !important;
}

.el-dialog.terminal-connect-dialog .el-tag.el-tag--success {
  background: rgba(78, 201, 176, 0.12) !important;
  color: #4ec9b0 !important;
}

.el-dialog.terminal-connect-dialog .el-tag.el-tag--warning {
  background: rgba(217, 119, 6, 0.12) !important;
  color: #d97706 !important;
}

.el-dialog.terminal-connect-dialog .el-tag.el-tag--info {
  background: rgba(107, 114, 128, 0.12) !important;
  color: #9ca3af !important;
}

.el-dialog.terminal-connect-dialog .el-tag.el-tag--plain {
  background: rgba(78, 201, 176, 0.15) !important;
  border: 1px solid rgba(78, 201, 176, 0.4) !important;
  border-radius: 0 !important;
}

/* 按钮样式 - 无圆角 */
.el-dialog.terminal-connect-dialog .el-button {
  background: transparent !important;
  color: #cccccc !important;
  border: 1px solid #555 !important;
  border-radius: 0 !important;
  padding: 6px 16px !important;
  font-size: 13px !important;
  height: 30px !important;
}

.el-dialog.terminal-connect-dialog .el-button:hover {
  background: rgba(255, 255, 255, 0.08) !important;
  border-color: #666 !important;
  color: #e0e0e0 !important;
}

.el-dialog.terminal-connect-dialog .el-button--primary {
  background: #007acc !important;
  border-color: #007acc !important;
  color: #ffffff !important;
}

.el-dialog.terminal-connect-dialog .el-button--primary:hover {
  background: #1c8cd4 !important;
  border-color: #1c8cd4 !important;
  color: #ffffff !important;
}
</style>
