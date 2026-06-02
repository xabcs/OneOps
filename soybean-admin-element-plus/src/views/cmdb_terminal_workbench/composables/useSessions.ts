import { computed, ref } from 'vue';

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

  function removeSession(sessionId: number): void {
    const index = sessions.value.findIndex(s => s.id === sessionId);
    if (index !== -1) {
      sessions.value.splice(index, 1);
      if (activeSessionId.value === sessionId) {
        activeSessionId.value = sessions.value.length > 0 ? sessions.value[0].id : null;
      }
      saveToStorage();
    }
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
      localStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(sessions.value));
    } catch (error) {
      console.error('保存会话列表失败:', error);
    }
  }

  function loadFromStorage(): void {
    try {
      const stored = localStorage.getItem(SESSIONS_STORAGE_KEY);
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

  // 初始化时加载
  loadFromStorage();

  return {
    sessions,
    activeSession,
    activeSessionId,
    addSession,
    removeSession,
    switchSession,
    updateSession
  };
}
