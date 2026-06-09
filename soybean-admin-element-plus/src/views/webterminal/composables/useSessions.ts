import { computed, ref } from 'vue';
import { fetchTerminateSession } from '@/service/api/cmdb';

export interface WorkbenchSession {
  id: number;
  serverId: number;
  serverName: string;
  serverIp: string;
  loginAccount: string;
  protocol: 'ssh' | 'sftp';
  status: 'connecting' | 'connected' | 'disconnected' | 'error';
  connected: boolean;
  duration: number;
  startedAt?: string;
  errorMessage?: string;
  websocketUrl?: string;
}

const SESSIONS_STORAGE_KEY = 'oneops_workbench_sessions';

export function useSessions() {
  const sessions = ref<WorkbenchSession[]>([]);
  const activeSessionId = ref<number | null>(null);

  const activeSession = computed(() => {
    if (!activeSessionId.value) return null;
    return sessions.value.find(s => s.id === activeSessionId.value) || null;
  });

  function addSession(session: WorkbenchSession): void {
    const existing = sessions.value.find(s => s.id === session.id);
    if (existing) {
      Object.assign(existing, session);
    } else {
      sessions.value.push({
        ...session,
        startedAt: new Date().toISOString(),
        duration: 0
      });
    }
    activeSessionId.value = session.id;
    saveToStorage();
  }

  async function removeSession(sessionId: number): Promise<void> {
    const index = sessions.value.findIndex(s => s.id === sessionId);
    if (index !== -1) {
      // 先调用后端 API 终止会话
      try {
        await fetchTerminateSession(sessionId);
      } catch (error) {
        console.error('终止会话失败:', error);
        // 即使 API 调用失败，也删除前端记录
      }

      // 删除前端记录
      sessions.value.splice(index, 1);
      if (activeSessionId.value === sessionId) {
        activeSessionId.value = sessions.value.length > 0 ? sessions.value[0].id : null;
      }
      saveToStorage();
    }
  }

  // 批量终止所有会话（用于页面关闭前）
  async function terminateAllSessions(): Promise<void> {
    const sessionIds = sessions.value.map(s => s.id);
    await Promise.allSettled(sessionIds.map(id => fetchTerminateSession(id)));
  }

  // 清理所有会话（不调用 API，只清理前端）
  function clearAllSessions(): void {
    sessions.value = [];
    activeSessionId.value = null;
    saveToStorage();
  }

  function switchSession(sessionId: number): void {
    const session = sessions.value.find(s => s.id === sessionId);
    if (session) {
      activeSessionId.value = sessionId;
    }
  }

  function updateSession(sessionId: number, updates: Partial<WorkbenchSession>): void {
    const session = sessions.value.find(s => s.id === sessionId);
    if (session) {
      Object.assign(session, updates);
      saveToStorage();
    }
  }

  function saveToStorage(): void {
    try {
      // 使用 sessionStorage 替代 localStorage，会话只在当前标签页有效
      sessionStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(sessions.value));
    } catch (error) {
      console.error('保存会话列表失败:', error);
    }
  }

  function loadFromStorage(): void {
    try {
      const stored = sessionStorage.getItem(SESSIONS_STORAGE_KEY);
      if (stored) {
        sessions.value = JSON.parse(stored);
        if (sessions.value.length > 0) {
          activeSessionId.value = sessions.value[0].id;
        }
      }
    } catch (error) {
      console.error('加载会话列表失败:', error);
    }
  }

  function clearStorage(): void {
    try {
      sessionStorage.removeItem(SESSIONS_STORAGE_KEY);
      localStorage.removeItem(SESSIONS_STORAGE_KEY);
    } catch (error) {
      console.error('清除会话列表失败:', error);
    }
  }

  // 不自动恢复会话，让用户手动连接
  // 如需恢复，可以调用 loadFromStorage()
  // loadFromStorage();

  return {
    sessions,
    activeSession,
    activeSessionId,
    addSession,
    removeSession,
    switchSession,
    updateSession,
    terminateAllSessions,
    clearAllSessions
  };
}
