import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElNotification } from 'element-plus';
import { fetchConnectServer, fetchGetServerForConnect } from '@/service/api';

/**
 * 会话管理 composable
 * 管理终端会话的创建、切换、移除以及连接逻辑
 */
export function useSessionManager() {
  const route = useRoute();
  const router = useRouter();

  const sessions = ref<Bastion.TerminalSession[]>([]);
  const activeSession = ref<Bastion.TerminalSession | null>(null);

  // 连接对话框状态
  const showConnectDialog = ref(false);
  const connectingServer = ref<CMDB.Server | null>(null);
  const selectedCredentialId = ref<number | null>(null);

  // 当前会话的 serverId 列表
  const currentSessionIds = computed(() => sessions.value.map(s => s.serverId));

  // 切换会话
  function switchSession(sessionId: number | string) {
    activeSession.value = sessions.value.find(s => s.id === sessionId) || null;
  }

  // 移除会话
  function removeSession(sessionId: number | string) {
    const index = sessions.value.findIndex(s => s.id === sessionId);
    if (index > -1) {
      sessions.value.splice(index, 1);
      if (activeSession.value?.id === sessionId) {
        activeSession.value = sessions.value[0] || null;
      }
    }
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
      // 确保有选择凭证
      if (!selectedCredentialId.value) {
        ElNotification.error({
          title: '连接失败',
          message: '请先选择连接凭证'
        });
        return;
      }

      // 查找选择的凭证信息，获取实际的用户名
      const selectedCredential = connectingServer.value.credentials?.find(c => c.id === selectedCredentialId.value);
      const loginAccount = selectedCredential?.username || 'root';

      // 调用后端接口创建 SSH 会话
      const response = await fetchConnectServer(connectingServer.value.id, {
        protocol: 'ssh',
        credentialId: selectedCredentialId.value
      });

      if (response.data && response.data.sessionId) {
        const sessionId = response.data.sessionId;
        const websocketUrl = response.data.websocketUrl;

        // 创建会话对象
        const newSession: Bastion.TerminalSession = {
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

  // 处理主机连接（统一入口）
  async function handleConnect(server: Bastion.BasicServerInfo) {
    try {
      const response = await fetchGetServerForConnect(server.id);

      if (response.data) {
        connectingServer.value = response.data;

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

  return {
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
    handleConnect
  };
}
