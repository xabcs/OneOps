import { computed, ref } from 'vue';
import { fetchTerminateSession } from '@/service/api/cmdb';

/**
 * 工作台会话类型 - 复用全局 TerminalSession 定义
 * @deprecated 直接使用 `Bastion.TerminalSession` 即可
 */
export type WorkbenchSession = Bastion.TerminalSession;

const SESSIONS_STORAGE_KEY = 'oneops_workbench_sessions';

export function useSessions() {
  const sessions = ref<WorkbenchSession[]>([]);
  const activeSessionId = ref<number | string | null>(null);

  const activeSession = computed(() => {
    if (activeSessionId.value === null) return null;
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

  async function removeSession(sessionId: number | string): Promise<void> {
    const index = sessions.value.findIndex(s => s.id === sessionId);
    if (index !== -1) {
      // 先调用后端 API 终止会话（仅数字 ID 调用）
      if (typeof sessionId === 'number') {
        try {
          await fetchTerminateSession(sessionId);
        } catch (error) {
          console.error('终止会话失败:', error);
          // 即使 API 调用失败，也删除前端记录
        }
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
    const sessionIds = sessions.value
      .map(s => s.id)
      .filter((id): id is number => typeof id === 'number');
    await Promise.allSettled(sessionIds.map(id => fetchTerminateSession(id)));
  }

  // 清理所有会话（不调用 API，只清理前端）
  function clearAllSessions(): void {
    sessions.value = [];
    activeSessionId.value = null;
    saveToStorage();
  }

  function switchSession(sessionId: number | string): void {
    const session = sessions.value.find(s => s.id === sessionId);
    if (session) {
      activeSessionId.value = sessionId;
    }
  }

  function updateSession(sessionId: number | string, updates: Partial<WorkbenchSession>): void {
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

  // 不自动恢复会话，让用户手动连接

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
