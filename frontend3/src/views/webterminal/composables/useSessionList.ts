import { ref } from 'vue';
import { fetchGetActiveSessionsFromMemory, fetchGetSessionsList, fetchTerminateSession } from '@/service/api/cmdb';

/** 标签页类型 */
export type TabType = 'active' | 'terminated' | 'history';

/**
 * 会话列表逻辑 composable
 * 管理会话列表的数据加载、搜索、选择、终止等逻辑
 */
export function useSessionList() {
  /** 当前标签页 */
  const activeTab = ref<TabType>('active');

  /** 搜索关键词 */
  const searchKeyword = ref('');

  /** 选中的会话ID列表 */
  const selectedSessionIds = ref<number[]>([]);

  /** 加载状态 */
  const loading = ref(false);

  /** 数据源 */
  const dataSource = ref<Bastion.BastionSession[]>([]);

  /** 分页信息 */
  const pagination = ref({
    current: 1,
    pageSize: 20,
    total: 0
  });

  /** 格式化持续时长 */
  function formatDuration(seconds: number): string {
    if (seconds < 60) {
      return `${seconds}秒`;
    }
    if (seconds < 3600) {
      const minutes = Math.floor(seconds / 60);
      return `${minutes}分钟`;
    }
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return minutes > 0 ? `${hours}小时${minutes}分` : `${hours}小时`;
  }

  /** 格式化时间 */
  function formatTime(timeStr?: string): string {
    if (!timeStr) return '-';
    const date = new Date(timeStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();

    if (diff < 60000) {
      return '刚刚';
    }
    if (diff < 3600000) {
      return `${Math.floor(diff / 60000)}分钟前`;
    }
    if (diff < 86400000) {
      return `${Math.floor(diff / 3600000)}小时前`;
    }
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  /** 加载会话列表 */
  async function loadSessions() {
    loading.value = true;
    try {
      const params: {
        page: number;
        pageSize: number;
        status?: string;
      } = {
        page: pagination.value.current,
        pageSize: pagination.value.pageSize
      };

      if (activeTab.value === 'active') {
        const { data, error } = await fetchGetActiveSessionsFromMemory();
        if (!error && data) {
          // console.table(data) 在 devtools 打开时渲染大表极慢，是"在线会话"加载卡顿的元凶之一
          dataSource.value = data;
        } else {
          dataSource.value = [];
        }
        pagination.value.total = dataSource.value.length;
        return;
      } else if (activeTab.value === 'terminated') {
        params.status = 'terminated';
      }

      if (searchKeyword.value) {
        // 后端需要支持模糊查询
      }

      const { data: pageData, error: pageError } = await fetchGetSessionsList(params);
      if (!pageError && pageData) {
        dataSource.value = pageData.list || [];
        pagination.value.total = pageData.total || 0;
      } else {
        dataSource.value = [];
        pagination.value.total = 0;
      }
    } catch (error) {
      console.error('加载会话列表失败:', error);
      dataSource.value = [];
      pagination.value.total = 0;
    } finally {
      loading.value = false;
    }
  }

  /** 标签页切换 */
  function handleTabChange(tab: TabType) {
    activeTab.value = tab;
    selectedSessionIds.value = [];
    pagination.value.current = 1;
    loadSessions();
  }

  /** 搜索处理 */
  function handleSearch() {
    pagination.value.current = 1;
    loadSessions();
  }

  /** 清空搜索 */
  function handleClearSearch() {
    searchKeyword.value = '';
    pagination.value.current = 1;
    loadSessions();
  }

  /** 全选/取消全选 */
  function handleSelectAll(checked: boolean) {
    if (checked) {
      selectedSessionIds.value = dataSource.value.map((s: Bastion.BastionSession) => s.id);
    } else {
      selectedSessionIds.value = [];
    }
  }

  /** 单行选择 */
  function handleSelectRow(sessionId: number, checked: boolean) {
    if (checked) {
      if (!selectedSessionIds.value.includes(sessionId)) {
        selectedSessionIds.value.push(sessionId);
      }
    } else {
      selectedSessionIds.value = selectedSessionIds.value.filter(id => id !== sessionId);
    }
  }

  /** 终止单个会话 */
  async function handleTerminateSession(sessionId: number) {
    const session = dataSource.value.find((s: Bastion.BastionSession) => s.id === sessionId);
    if (!session) return;

    const serverName = session.server ? session.server.hostname : `服务器 ${session.serverId}`;
    const confirmMsg = `确定要终止连接到 ${serverName} 的会话吗？\n\n登录账号: ${session.loginAccount}`;

    if (!confirm(confirmMsg)) {
      return;
    }

    try {
      loading.value = true;
      await fetchTerminateSession(sessionId);
      window.$message?.success('会话已终止');
      loadSessions();
    } catch (error) {
      window.$message?.error('终止会话失败');
    } finally {
      loading.value = false;
    }
  }

  /** 批量终止会话 */
  async function handleBatchTerminate() {
    if (selectedSessionIds.value.length === 0) {
      window.$message?.warning('请先选择要终止的会话');
      return;
    }

    const confirmMsg = `确定要终止选中的 ${selectedSessionIds.value.length} 个会话吗？`;
    if (!confirm(confirmMsg)) {
      return;
    }

    try {
      loading.value = true;
      for (const sessionId of selectedSessionIds.value) {
        await fetchTerminateSession(sessionId);
      }
      window.$message?.success(`已终止 ${selectedSessionIds.value.length} 个会话`);
      selectedSessionIds.value = [];
      loadSessions();
    } catch (error) {
      window.$message?.error('批量终止会话失败');
    } finally {
      loading.value = false;
    }
  }

  /** 分页变化 */
  function handlePageChange(page: number) {
    pagination.value.current = page;
    loadSessions();
  }

  /** 每页数量变化 */
  function handlePageSizeChange(pageSize: number) {
    pagination.value.pageSize = pageSize;
    pagination.value.current = 1;
    loadSessions();
  }

  /** 刷新列表 */
  function handleRefresh() {
    loadSessions();
  }

  /** 表格点击事件代理 */
  function handleTableAction(event: Event) {
    const target = event.target as HTMLElement;
    const action = target.dataset.action;
    const sessionId = target.dataset.sessionId;

    if (action === 'terminate' && sessionId) {
      handleTerminateSession(Number(sessionId));
    }
  }

  return {
    activeTab,
    searchKeyword,
    selectedSessionIds,
    loading,
    dataSource,
    pagination,
    formatDuration,
    formatTime,
    loadSessions,
    handleTabChange,
    handleSearch,
    handleClearSearch,
    handleSelectAll,
    handleSelectRow,
    handleTerminateSession,
    handleBatchTerminate,
    handlePageChange,
    handlePageSizeChange,
    handleRefresh,
    handleTableAction
  };
}
