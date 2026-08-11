<script setup lang="ts">
import { onMounted, ref } from 'vue';
import {
  fetchGetActiveSessionsFromMemory,
  fetchGetSessionsList,
  fetchTerminateSession
} from '@/service/api/cmdb';

defineOptions({
  name: 'WebterminalSessionList'
});

/** 标签页类型 */
type TabType = 'active' | 'terminated' | 'history';

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

  // 小于1分钟
  if (diff < 60000) {
    return '刚刚';
  }
  // 小于1小时
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}分钟前`;
  }
  // 小于24小时
  if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)}小时前`;
  }
  // 大于24小时，显示具体日期
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
}

/** 标签页切换 */
function handleTabChange(tab: TabType) {
  activeTab.value = tab;
  selectedSessionIds.value = [];
  pagination.value.current = 1;
  // 不清空 dataSource，保留旧数据直到新数据加载完成，避免空状态闪烁
  loadSessions();
}

/** 加载会话列表 */
async function loadSessions() {
  loading.value = true;
  try {
    const params: Record<string, any> = {
      page: pagination.value.current,
      pageSize: pagination.value.pageSize
    };

    // 根据标签页设置状态筛选
    if (activeTab.value === 'active') {
      // 在线会话：从内存获取真正活跃的会话（更准确）
      console.log('=== 开始加载在线会话 ===');
      const response = await fetchGetActiveSessionsFromMemory();
      console.log('API 响应:', response);
      console.log('会话数量:', response.data?.length || 0);
      console.table(response.data || []);
      dataSource.value = response.data || [];
      console.log('赋值后 dataSource.length:', dataSource.value.length);
      pagination.value.total = dataSource.value.length;
      console.log('=== 在线会话加载完成 ===');
      return; // ✅ 直接返回，不要继续执行下面的通用查询
    } else if (activeTab.value === 'terminated') {
      params.status = 'terminated';
    }
    // history 不过滤状态，显示所有

    // 搜索关键词
    if (searchKeyword.value) {
      // 可以根据主机名、IP、用户名等搜索
      // 后端需要支持模糊查询，这里暂时只传参数
    }

    // 使用轻量级接口（已终止会话、会话历史）
    console.log(`=== 加载 ${activeTab.value} 会话 ===`);
    console.log('请求参数:', params);
    const response = await fetchGetSessionsList(params);
    console.log('API 响应:', response);
    console.log('返回数据量:', response.data?.list?.length || 0);
    dataSource.value = response.data?.list || [];
    pagination.value.total = response.data?.total || 0;
    console.log('赋值后 dataSource.length:', dataSource.value.length);
  } catch (error) {
    console.error('加载会话列表失败:', error);
    dataSource.value = [];
    pagination.value.total = 0;
  } finally {
    loading.value = false;
  }
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
    // 逐个终止会话
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

onMounted(() => {
  loadSessions();
});
</script>

<template>
  <div class="session-list-page">
    <!-- 头部工具栏 -->
    <div class="wb-toolbar">
      <div class="wb-toolbar-left">
        <h1 class="wb-page-title">会话列表</h1>
      </div>
      <div class="wb-toolbar-right">
        <button v-if="selectedSessionIds.length > 0" class="wb-button wb-button-danger" @click="handleBatchTerminate">
          终止选中 ({{ selectedSessionIds.length }})
        </button>
        <button class="wb-button wb-button-default" @click="handleRefresh">
          <i class="codicon codicon-refresh"></i>
          刷新
        </button>
      </div>
    </div>

    <!-- 标签页 -->
    <div class="wb-tabs">
      <div class="wb-tab" :class="{ active: activeTab === 'active' }" @click="handleTabChange('active')">在线会话</div>
      <div class="wb-tab" :class="{ active: activeTab === 'terminated' }" @click="handleTabChange('terminated')">
        已终止会话
      </div>
      <div class="wb-tab" :class="{ active: activeTab === 'history' }" @click="handleTabChange('history')">
        会话历史
      </div>
    </div>

    <!-- 搜索栏 -->
    <div class="wb-search-bar">
      <div class="wb-search-input-wrapper">
        <i class="codicon codicon-search wb-search-icon"></i>
        <input
          v-model="searchKeyword"
          type="text"
          class="wb-search-input"
          placeholder="搜索主机名、IP、用户名..."
          autocomplete="off"
          style="border: none !important; outline: none !important; box-shadow: none !important"
          @keyup.enter="handleSearch"
        />
        <button v-if="searchKeyword" class="wb-search-clear" @click="handleClearSearch">
          <i class="codicon codicon-close"></i>
        </button>
      </div>
      <button class="wb-button wb-button-primary" @click="handleSearch">搜索</button>
    </div>

    <!-- 数据表格 -->
    <div class="wb-table-container" @click="handleTableAction">
      <table class="wb-table">
        <colgroup>
          <col class="wb-checkbox-column" style="width: 40px" />
          <col class="wb-index-column" style="width: 50px" />
          <col style="width: 140px" />
          <col style="width: 180px" />
          <col style="width: 80px" />
          <col style="width: 70px" />
          <col style="width: 80px" />
          <col style="width: 90px" />
          <col style="width: 80px" />
          <col style="width: 110px" />
          <col style="width: 60px" />
        </colgroup>
        <thead>
          <tr>
            <th class="wb-checkbox-column">
              <input
                type="checkbox"
                :checked="selectedSessionIds.length > 0 && selectedSessionIds.length === dataSource.length"
                @change="
                  (e: Event) => {
                    const target = e.target as HTMLInputElement;
                    handleSelectAll(target.checked);
                  }
                "
              />
            </th>
            <th class="wb-index-column">序号</th>
            <th>创建/活动时间</th>
            <th>主机/实例</th>
            <th>协议/标识</th>
            <th>状态</th>
            <th>持续时长</th>
            <th>登录账号</th>
            <th>操作用户</th>
            <th>客户端IP</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody v-if="!loading && dataSource.length > 0">
          <tr
            v-for="(session, index) in dataSource"
            :key="session.id"
            :class="{ 'wb-row-selected': selectedSessionIds.includes(session.id) }"
          >
            <td class="wb-checkbox-column">
              <input
                type="checkbox"
                :checked="selectedSessionIds.includes(session.id)"
                @change="
                  (e: Event) => {
                    const target = e.target as HTMLInputElement;
                    handleSelectRow(session.id, target.checked);
                  }
                "
              />
            </td>
            <td class="wb-index-column">{{ (pagination.current - 1) * pagination.pageSize + index + 1 }}</td>
            <td>{{ formatTime(session.startedAt || session.createdAt) }}</td>
            <td>
              <template v-if="session.server">{{ session.server.hostname }} ({{ session.server.ip }})</template>
              <template v-else>服务器ID: {{ session.serverId }}</template>
            </td>
            <td>{{ session.protocol.toUpperCase() }}</td>
            <td>
              <span
                :class="{
                  'wb-status-online': session.status === 'active',
                  'wb-status-offline': session.status !== 'active'
                }"
              >
                {{
                  session.status === 'active'
                    ? '在线'
                    : session.status === 'terminated'
                      ? '已终止'
                      : session.status === 'closed'
                        ? '已关闭'
                        : '异常'
                }}
              </span>
            </td>
            <td>{{ formatDuration(session.duration) }}</td>
            <td>{{ session.loginAccount }}</td>
            <td>{{ session.username }}</td>
            <td>{{ session.clientIp || '-' }}</td>
            <td>
              <button
                v-if="session.status === 'active'"
                class="wb-action-btn"
                data-action="terminate"
                :data-session-id="session.id"
              >
                终止会话
              </button>
              <span v-else class="wb-text-muted">已结束</span>
            </td>
          </tr>
        </tbody>
        <tbody v-else-if="loading">
          <tr>
            <td colspan="11" class="wb-table-loading">
              <i class="codicon codicon-loading wb-spinning"></i>
              加载中...
            </td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr>
            <td colspan="11" class="wb-table-empty">
              <i class="codicon codicon-inbox wb-empty-icon"></i>
              <span>暂无会话记录</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页器 -->
    <div v-if="pagination.total > 0" class="wb-pagination">
      <div class="wb-pagination-info">共 {{ pagination.total }} 条记录</div>
      <div class="wb-pagination-controls">
        <button
          class="wb-pagination-btn"
          :disabled="pagination.current === 1"
          @click="handlePageChange(pagination.current - 1)"
        >
          上一页
        </button>
        <span class="wb-pagination-page">第 {{ pagination.current }} 页</span>
        <button
          class="wb-pagination-btn"
          :disabled="pagination.current * pagination.pageSize >= pagination.total"
          @click="handlePageChange(pagination.current + 1)"
        >
          下一页
        </button>
        <select
          class="wb-pagination-size"
          :value="pagination.pageSize"
          @change="(e: Event) => handlePageSizeChange(Number((e.target as HTMLSelectElement).value))"
        >
          <option value="20">20 条/页</option>
          <option value="50">50 条/页</option>
          <option value="100">100 条/页</option>
        </select>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.session-list-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #171717;
  color: #fff;
  font-size: 12px;
}

/* ========== 工具栏 ========== */
.wb-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-bottom: none;
  background: #171717;
  min-height: 40px;
}

.wb-toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wb-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wb-page-title {
  font-size: 14px;
  font-weight: 500;
  color: #ffffff;
  margin: 0;
}

/* ========== 按钮样式 ========== */
.wb-button {
  height: 28px;
  padding: 0 12px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: 1px solid #3c3c3c;
  border-radius: 2px;
  color: #ffffff;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: #252526;
    border-color: #404040;
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.wb-button-primary {
  background: transparent;
  border: 1px solid #3c3c3c;
  color: #ffffff;

  &:hover {
    background: #252526;
    border-color: #404040;
  }

  &:active {
    background: #2a2d2e;
  }
}

.wb-button-danger {
  background: #f14c4c;
  border-color: #f14c4c;
  color: #ffffff;

  &:hover {
    background: #ff6060;
    border-color: #ff6060;
  }
}

.wb-action-btn {
  padding: 2px 8px;
  background: transparent;
  border: none;
  color: #007acc;
  cursor: pointer;
  font-size: 12px;

  &:hover {
    text-decoration: underline;
  }
}

/* ========== 标签页 ========== */
.wb-tabs {
  display: flex;
  border-bottom: none;
  background: #171717;
  padding-left: 16px; /* 与工具栏、搜索栏对齐 */
}

.wb-tab {
  padding: 8px 16px 8px 0; /* 左侧去掉padding，由父容器统一控制 */
  cursor: pointer;
  color: #ffffff;
  transition: all 0.2s;
  position: relative; /* 为伪元素定位 */
  display: inline-flex; /* flex 布局 */
  align-items: center;

  /* 使用伪元素作为下划线，宽度只覆盖文本内容 */
  &::before {
    content: '';
    position: absolute;
    left: 0;
    right: 16px; /* 排除右边的 padding */
    bottom: 0;
    height: 2px;
    background: transparent;
  }

  &:hover {
    color: #ffffff;
  }

  &.active {
    /* 激活状态下，伪元素显示为蓝色下划线 */
    &::before {
      background: #007acc;
    }
  }
}

/* ========== 搜索栏 ========== */
.wb-search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: none;
  background: #171717;
}

.wb-search-input-wrapper {
  position: relative;
  flex: 1;
  max-width: 400px;
  height: 30px;
  background: #252526;
  border: 1px solid #252526;
  border-radius: 15px;
  transition: all 0.2s;

  // 当内部input获得焦点时，改变wrapper的边框
  &:has(.wb-search-input:focus) {
    border-color: #007acc;
  }
}

.wb-search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: #858585;
  font-size: 12px;
  pointer-events: none;
  z-index: 2;
}

.wb-search-input {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  padding: 0 36px;
  background: transparent !important;
  border: none !important;
  outline: none !important;
  box-shadow: none !important;
  color: #cccccc;
  font-size: 12px;

  &:focus {
    background: transparent !important;
    border: none !important;
    outline: none !important;
    box-shadow: none !important;
  }

  &::placeholder {
    color: #6e6e6e;
  }
}

.wb-search-clear {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: #858585;
  cursor: pointer;
  padding: 4px;
  z-index: 2;

  &:hover {
    color: #cccccc;
  }
}

/* ========== 表格 ========== */
.wb-table-container {
  flex: 1;
  overflow: auto;
  background: #171717;
  padding-left: 16px; /* 与工具栏、标签、搜索框对齐 */
}

.wb-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  table-layout: fixed; // 固定列宽，避免内容变化导致列宽跳动

  thead {
    position: sticky;
    top: 0;
    background: transparent;
    z-index: 10;
  }

  th {
    text-align: left;
    padding: 10px 12px;
    border-bottom: 1px solid #333333;
    font-weight: 500;
    color: #cccccc;
    white-space: nowrap;
    vertical-align: middle;
  }

  td {
    padding: 10px 12px;
    border-bottom: 1px solid #333333;
    vertical-align: middle;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  tbody tr {
    &:hover {
      /* 移除悬停背景色 */

      .wb-index-column {
        /* 移除悬停背景色 */
      }

      .wb-checkbox-column {
        /* 移除悬停背景色 */
      }
    }

    &.wb-row-selected {
      /* 移除选中背景色 */

      .wb-index-column {
        /* 移除选中背景色 */
      }

      .wb-checkbox-column {
        /* 移除选中背景色 */
      }
    }
  }

  tbody tr .wb-index-column,
  tbody tr .wb-checkbox-column {
    /* 移除默认背景色 */
  }
}

/* 选择框列样式 */
.wb-checkbox-column {
  width: 48px;
  text-align: center;
  border-right: 1px solid #333333;
  padding: 10px 12px !important;
  vertical-align: middle;
}

/* 选择框样式 */
.wb-checkbox-column input[type='checkbox'] {
  width: 14px;
  height: 14px;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background: #252526;
  border: 1px solid #555555;
  border-radius: 2px;
  position: relative;
  vertical-align: middle;
  transition: all 0.15s ease;

  &:hover {
    border-color: #007acc;
    background: #2a2d2e;
  }

  &:checked {
    background: #007acc;
    border-color: #007acc;

    &::after {
      content: '';
      position: absolute;
      left: 3px;
      top: 0px;
      width: 4px;
      height: 8px;
      border: solid white;
      border-width: 0 2px 2px 0;
      transform: rotate(45deg);
    }
  }

  &:focus {
    outline: none;
    box-shadow: 0 0 0 1px #007acc;
  }
}

/* 表头选择框列特殊样式 */
thead th.wb-checkbox-column {
  padding: 10px 12px !important;
  text-align: center !important; /* 确保checkbox居中 */
  border-bottom: 1px solid #333333;
  font-weight: 500;
  color: #ffffff;
}

/* 序号列样式 */
.wb-index-column {
  width: 80px;
  text-align: left;
  padding: 10px 12px 10px 16px !important;
  color: #ffffff;
  font-size: 14px;
  font-weight: 400;
  border-right: 1px solid #333333;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  letter-spacing: 0.5px;
  vertical-align: middle;
}

/* 表头序号列特殊样式 */
thead th.wb-index-column {
  padding: 10px 12px 10px 16px !important;
  color: #ffffff;
  font-weight: 500;
  border-bottom: 1px solid #333333;
}

.wb-status-online {
  color: #4ec9b0;
}

.wb-status-offline {
  color: #858585;
}

.wb-text-muted {
  color: #6e6e6e;
}

.wb-table-loading,
.wb-table-empty {
  text-align: center;
  padding: 40px 20px;
  color: #858585;
}

.wb-empty-icon {
  font-size: 32px;
  display: block;
  margin-bottom: 8px;
  opacity: 0.5;
}

/* ========== 分页 ========== */
.wb-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-top: none;
  background: #171717;
}

.wb-pagination-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wb-pagination-btn {
  padding: 4px 12px;
  background: transparent;
  border: 1px solid #3c3c3c;
  border-radius: 2px;
  color: #cccccc;
  cursor: pointer;
  font-size: 12px;

  &:hover:not(:disabled) {
    background: #252526;
  }

  &:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }
}

.wb-pagination-page {
  color: #858585;
}

.wb-pagination-size {
  padding: 4px 8px;
  background: #252526;
  border: 1px solid #252526;
  border-radius: 2px;
  color: #cccccc;
  cursor: pointer;
  font-size: 12px;

  &:focus {
    border-color: #007acc;
  }
}

/* ========== 滚动条 ========== */
.wb-table-container::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.wb-table-container::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.wb-table-container::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 2px;

  &:hover {
    background: #4f4f4f;
  }
}
</style>
