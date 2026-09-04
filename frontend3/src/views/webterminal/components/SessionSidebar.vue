<script setup lang="ts">
  import { ref } from 'vue';
  import { ElButton, ElDialog, ElForm, ElFormItem, ElOption, ElSelect } from 'element-plus';
  import AssetTree from './AssetTree.vue';

  interface Session {
    id: number;
    serverId: number;
    serverName: string;
    serverIp: string;
    loginAccount: string;
    connected: boolean;
    duration: number;
  }

  interface QuickConnectHost {
    id: number;
    hostname: string;
    ip: string;
  }

  interface Props {
    sessions: Session[];
    activeId: number | null;
    currentSessionIds: number[];
  }

  interface Emits {
    (e: 'select', sessionId: number): void;
    (e: 'remove', sessionId: number): void;
    (e: 'connect', server: CMDB.Server & { loginAccount?: string }): void;
  }

  const props = defineProps<Props>();
  const emit = defineEmits<Emits>();

  const showConnectDialog = ref(false);
  const selectedServer = ref<CMDB.Server | null>(null);
  const loginAccount = ref('');

  // 处理主机连接
  function handleConnect(server: CMDB.Server) {
    selectedServer.value = server;
    loginAccount.value = server.sshCredential?.username || 'root';
    showConnectDialog.value = true;
  }

  // 确认连接
  function handleConfirmConnect() {
    if (selectedServer.value) {
      emit('connect', {
        ...selectedServer.value,
        loginAccount: loginAccount.value
      });
      showConnectDialog.value = false;
    }
  }

  // 取消连接
  function handleCancelConnect() {
    showConnectDialog.value = false;
    selectedServer.value = null;
  }
</script>

<template>
  <div class="sidebar">
    <!-- 会话列表 -->
    <div class="section">
      <div class="section-header">
        <span class="section-title">会话 ({{ sessions.length }})</span>
      </div>

      <div v-if="sessions.length === 0" class="empty">
        <icon-mdi-clipboard-off class="empty-icon" />
        <p>暂无会话</p>
        <p class="hint">双击下方主机或右键菜单连接</p>
      </div>

      <div v-else class="session-list">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: session.id === activeId }"
          @click="$emit('select', session.id)"
        >
          <span class="status" :class="{ online: session.connected }"></span>
          <div class="session-info">
            <div class="name">{{ session.serverName }}</div>
            <div class="meta">{{ session.loginAccount }}@{{ session.serverIp }}</div>
          </div>
          <ElButton size="small" type="danger" link @click.stop="$emit('remove', session.id)">
            <icon-mdi-close />
          </ElButton>
        </div>
      </div>
    </div>

    <!-- 资产树 -->
    <AssetTree :current-sessions="currentSessionIds" @connect="handleConnect" />

    <!-- 连接对话框 -->
    <ElDialog v-model="showConnectDialog" :title="`连接 ${selectedServer?.hostname || ''}`" width="400px">
      <ElForm label-width="80px">
        <ElFormItem label="主机">
          <span>{{ selectedServer?.hostname }} ({{ selectedServer?.ip }})</span>
        </ElFormItem>
        <ElFormItem label="登录账号">
          <ElSelect v-model="loginAccount" placeholder="选择登录账号">
            <ElOption value="root" label="root" />
            <ElOption
              v-if="selectedServer?.sshCredential?.username && selectedServer.sshCredential.username !== 'root'"
              :value="selectedServer.sshCredential.username"
              :label="selectedServer.sshCredential.username"
            />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="handleCancelConnect">取消</ElButton>
        <ElButton type="primary" @click="handleConfirmConnect">连接</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
  .sidebar {
    width: 240px;
    height: 100vh;
    background: #252526;
    border-right: 1px solid #3e3e42;
    display: flex;
    flex-direction: column;
  }

  .section {
    padding: 12px;
  }

  .section:not(:last-child) {
    border-bottom: 1px solid #3e3e42;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .section-title {
    font-size: 13px;
    font-weight: 600;
    color: #ccc;
  }

  .icon {
    margin-right: 4px;
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 32px 12px;
    color: #858585;
  }

  .empty-icon {
    font-size: 40px;
    margin-bottom: 12px;
    opacity: 0.5;
  }

  .empty p {
    margin: 0 0 16px;
    font-size: 12px;
  }

  .session-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .session-item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.2s;
  }

  .session-item:hover {
    background: #2a2d2e;
  }

  .session-item.active {
    background: #37373d;
  }

  .status {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 8px;
    background: #757575;
  }

  .status.online {
    background: #4caf50;
    box-shadow: 0 0 4px #4caf50;
  }

  .session-info {
    flex: 1;
    min-width: 0;
  }

  .name {
    font-size: 13px;
    color: #ccc;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .meta {
    font-size: 11px;
    color: #858585;
    margin-top: 2px;
  }

  .hint {
    font-size: 11px;
    color: #757575;
    margin-top: 4px;
  }
</style>
