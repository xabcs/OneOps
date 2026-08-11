<script setup lang="ts">
interface Props {
  dataSource: Bastion.BastionSession[];
  selectedSessionIds: number[];
  loading: boolean;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
  };
}

interface Emits {
  (e: 'select-all', checked: boolean): void;
  (e: 'select-row', sessionId: number, checked: boolean): void;
  (e: 'table-action', event: Event): void;
  (e: 'page-change', page: number): void;
  (e: 'page-size-change', pageSize: number): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分钟`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return minutes > 0 ? `${hours}小时${minutes}分` : `${hours}小时`;
}

function formatTime(timeStr?: string): string {
  if (!timeStr) return '-';
  const date = new Date(timeStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  if (diff < 60000) return '刚刚';
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
  return date.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
}
</script>

<template>
  <div class="wb-table-container" @click="emit('table-action', $event)">
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
              @change="(e: Event) => { const target = e.target as HTMLInputElement; emit('select-all', target.checked); }"
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
              @change="(e: Event) => { const target = e.target as HTMLInputElement; emit('select-row', session.id, target.checked); }"
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
        @click="emit('page-change', pagination.current - 1)"
      >
        上一页
      </button>
      <span class="wb-pagination-page">第 {{ pagination.current }} 页</span>
      <button
        class="wb-pagination-btn"
        :disabled="pagination.current * pagination.pageSize >= pagination.total"
        @click="emit('page-change', pagination.current + 1)"
      >
        下一页
      </button>
      <select
        class="wb-pagination-size"
        :value="pagination.pageSize"
        @change="(e: Event) => emit('page-size-change', Number((e.target as HTMLSelectElement).value))"
      >
        <option value="20">20 条/页</option>
        <option value="50">50 条/页</option>
        <option value="100">100 条/页</option>
      </select>
    </div>
  </div>
</template>

<style lang="scss" scoped>
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

/* ========== 表格 ========== */
.wb-table-container {
  flex: 1;
  overflow: auto;
  background: #171717;
  padding-left: 16px;
}

.wb-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  table-layout: fixed;

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
    &.wb-row-selected {
      /* 选中行样式 */
    }
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
  text-align: center !important;
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
