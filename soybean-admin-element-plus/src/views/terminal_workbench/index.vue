<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { onKeyStroke } from '@vueuse/core';
import { Icon } from '@iconify/vue';
import { ElDialog, ElButton, ElDescriptions, ElDescriptionsItem, ElTag, ElNotification, ElMessage, ElSelect, ElOption } from 'element-plus';
import { fetchConnectServer, fetchGetServerById } from '@/service/api';
import { localStg } from '@/utils/storage';
import SessionTabs from './components/SessionTabs.vue';
import TerminalArea from './components/TerminalArea.vue';
import AssetTree from './components/AssetTree.vue';

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
const sidebarWidth = ref(240);
const isResizing = ref(false);
const activeActivityItem = ref('assets');

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
});

// 当前会话的 serverId 列表
const currentSessionIds = computed(() => sessions.value.map((s: any) => s.serverId));
// 切换侧边栏显示
function toggleSidebar() {
  showSidebar.value = !showSidebar.value;
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
        loginAccount: loginAccount,
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
  console.log('是否包含凭证列表:', server?.credentials?.length);

  // 无论从哪个入口进入，都先获取完整的服务器信息
  try {
    ElMessage.info({
      message: '正在获取服务器信息...',
      duration: 2000
    });

    const response = await fetchGetServerById(server.id);
    console.log('fetchGetServerById 响应:', response);

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
  activeActivityItem.value = item;
}

// 组件挂载时检查 URL 参数
onMounted(async () => {
  const query = route.query;
  if (query.serverId && query.hostname) {
    // 从 URL 参数构建基本的服务器对象，然后调用 handleConnect
    const basicServerInfo = {
      id: Number(query.serverId),
      hostname: query.hostname as string,
      ip: query.ip as string,
      sshUser: query.sshUser as string || 'root',
      env: query.env as string,
      agentStatus: query.agentStatus as string
    };
    await handleConnect(basicServerInfo);
  }
});
</script>

<template>
  <div class="terminal-workbench" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- 活动栏 (Activity Bar) -->
    <div v-if="showActivityBar && !isFullscreen" class="wb-activity-bar">
      <!-- Logo -->
      <div class="wb-activity-logo">
        <Icon icon="lucide:terminal-square" class="wb-logo-icon" />
      </div>

      <!-- 活动项 -->
      <div
        class="wb-activity-item"
        :class="{ active: activeActivityItem === 'assets' }"
        title="主机资产"
        @click="switchActivityItem('assets')"
      >
        <Icon icon="lucide:servers" class="wb-activity-icon" />
      </div>

      <div
        class="wb-activity-item"
        :class="{ active: activeActivityItem === 'sessions' }"
        title="会话管理"
        @click="switchActivityItem('sessions')"
      >
        <Icon icon="lucide:history" class="wb-activity-icon" />
      </div>

      <!-- 底部设置 -->
      <div class="wb-activity-bottom">
        <div class="wb-activity-item" title="设置" @click="switchActivityItem('settings')">
          <Icon icon="lucide:settings" class="wb-activity-icon" />
        </div>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="wb-main-content">
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
                <Icon icon="lucide:shield-check" style="width: 14px; height: 14px; margin-right: 4px; color: #4ec9b0;" />
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
                <Icon icon="lucide:check-circle-2" style="width: 14px; height: 14px; margin-right: 4px; color: #4ec9b0;" />
                已启用
              </span>
            </div>
          </div>
        </div>

        <!-- 凭证选择 -->
        <div class="connect-section">
          <div class="connect-section-title">选择连接凭证</div>
          <div v-if="connectingServer.credentials && connectingServer.credentials.length > 0" class="credential-selector">
            <ElSelect v-model="selectedCredentialId" placeholder="请选择凭证" popper-class="terminal-select-dropdown" style="width: 100%;">
              <ElOption
                v-for="cred in connectingServer.credentials"
                :key="cred.id"
                :label="`${cred.name} - ${cred.username}`"
                :value="cred.id"
              >
                <span style="display: flex; align-items: center; gap: 8px;">
                  <Icon icon="lucide:key" style="width: 14px; height: 14px; color: #858585;" />
                  <span>{{ cred.name }}</span>
                  <ElTag size="small" effect="plain" style="margin-left: auto;">{{ cred.username }}</ElTag>
                </span>
              </ElOption>
            </ElSelect>
            <div class="credential-hint">
              <Icon icon="lucide:info" style="width: 14px; height: 14px; margin-right: 4px; color: #858585;" />
              <span>选择的凭证将决定登录账号</span>
            </div>
          </div>
          <div v-else class="credential-empty">
            <Icon icon="lucide:alert-circle" style="width: 16px; height: 16px; color: #f14c4c;" />
            <span>该服务器没有可用的连接凭证</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="connect-actions">
          <ElButton @click="cancelConnect">取消</ElButton>
          <ElButton type="primary" :disabled="!selectedCredentialId" @click="handleConnected()">
            <Icon icon="lucide:terminal" style="margin-right: 6px; width: 14px; height: 14px;" />
            连接
          </ElButton>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<style lang="scss">
@import '@/styles/scss/terminal-workbench.scss';

/* 覆盖 Element Plus 树节点的悬停背景色变量 */
.terminal-workbench {
  --el-fill-color-light: rgba(0, 0, 0, 0.2) !important;
  --el-fill-color-lighter: rgba(0, 0, 0, 0.2) !important;
  --el-fill-color-extra-light: rgba(0, 0, 0, 0.15) !important;
  --el-fill-color: rgba(0, 0, 0, 0.2) !important;
  --el-fill-color-dark: rgba(0, 0, 0, 0.3) !important;
}

/* 强制覆盖树节点悬停样式 */
.terminal-workbench :deep(.el-tree-node__content:hover) {
  background-color: rgba(0, 0, 0, 0.2) !important;
}

.terminal-workbench :deep(.el-tree-node__content) {
  background-color: transparent !important;
}

.terminal-workbench :deep(.el-tree-node:hover > .el-tree-node__content) {
  background-color: rgba(0, 0, 0, 0.2) !important;
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
}

.wb-terminal-area-wrapper {
  flex: 1;
  overflow: hidden;
  min-height: 0;
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

.wb-activity-item:hover .iconify {
  color: #aaaaaa !important;
}

.wb-activity-bottom {
  margin-top: auto;
  flex-shrink: 0;
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
