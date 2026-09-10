<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { CircleCloseFilled, Refresh } from '@element-plus/icons-vue';
  import {
    type DiagnosticSession,
    fetchDiagnosticSessionDetail,
    fetchDiagnosticSessions,
    terminateDiagnosticSession
  } from '@/service/api/diagnostic';
  import { formatDateTime } from '../shared';

  defineOptions({ name: 'DiagnosticSessionsTab' });

  const loading = ref(false);
  const sessions = ref<DiagnosticSession[]>([]);
  const replayVisible = ref(false);
  const replaySession = ref<DiagnosticSession | null>(null);
  const replayLoading = ref(false);
  const terminatingId = ref<number | null>(null);

  let pollTimer: ReturnType<typeof setInterval> | null = null;

  // 加载会话列表；silent 为 true 时（轮询场景）不置 loading，避免表格遮罩反复闪烁
  async function load(silent = false) {
    if (!silent) {
      loading.value = true;
    }
    try {
      const { data, error } = await fetchDiagnosticSessions();
      if (!error && data) {
        sessions.value = data;
      }
    } finally {
      if (!silent) {
        loading.value = false;
      }
    }
  }

  async function showReplay(session: DiagnosticSession) {
    replayVisible.value = true;
    replaySession.value = session;
    replayLoading.value = true;
    const { data, error } = await fetchDiagnosticSessionDetail(session.id);
    if (!error && data) {
      replaySession.value = data;
    }
    replayLoading.value = false;
  }

  async function terminate(session: DiagnosticSession) {
    try {
      await ElMessageBox.confirm(
        `将强制断开 ${session.username} 在 ${session.agentId} 上的诊断会话，并还原字节码增强。确认熔断？`,
        '管理员熔断',
        { type: 'warning', confirmButtonText: '强制断开', cancelButtonText: '取消' }
      );
    } catch {
      return;
    }
    terminatingId.value = session.id;
    try {
      const { error } = await terminateDiagnosticSession(session.id);
      if (!error) {
        ElMessage.success('会话已终止');
        load();
      }
    } finally {
      terminatingId.value = null;
    }
  }

  const statusMeta: Record<string, { label: string; type: 'success' | 'warning' | 'info' | 'danger' }> = {
    active: { label: '进行中', type: 'success' },
    closed: { label: '已关闭', type: 'info' },
    timeout: { label: '超时断开', type: 'warning' },
    terminated: { label: '被熔断', type: 'danger' }
  };

  onMounted(() => {
    load();
    // 轮询使用静默模式，不触发 loading 遮罩
    pollTimer = setInterval(() => load(true), 10_000);
  });

  onBeforeUnmount(() => {
    if (pollTimer) clearInterval(pollTimer);
  });
</script>

<template>
  <div class="sessions-tab">
    <div class="tab-toolbar">
      <span class="toolbar-title">活跃诊断会话（每实例同时仅允许一个会话，10 秒自动刷新）</span>
      <ElButton :icon="Refresh" size="small" :loading="loading" @click="load()">刷新</ElButton>
    </div>

    <ElTable v-loading="loading" :data="sessions" size="small">
      <ElTableColumn prop="id" label="会话" width="70" />
      <ElTableColumn prop="appName" label="应用" min-width="140" show-overflow-tooltip />
      <ElTableColumn prop="agentId" label="Agent（Pod）" min-width="220" show-overflow-tooltip />
      <ElTableColumn prop="username" label="操作者" width="110" />
      <ElTableColumn label="开始时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.startAt) }}</template>
      </ElTableColumn>
      <ElTableColumn label="状态" width="100">
        <template #default="{ row }">
          <ElTag size="small" :type="statusMeta[row.status]?.type ?? 'info'">
            {{ statusMeta[row.status]?.label ?? row.status }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" align="center" width="180" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <ElButton link type="primary" size="small" @click="showReplay(row)">I/O 回放</ElButton>
          <PermissionButton
            link
            type="danger"
            size="small"
            code="k8s.diagnostic.execute"
            :icon="CircleCloseFilled"
            :loading="terminatingId === row.id"
            @click="terminate(row)"
          >
            强制断开
          </PermissionButton>
        </template>
      </ElTableColumn>
      <template #empty>
        <ElEmpty description="当前无活跃诊断会话" :image-size="80" />
      </template>
    </ElTable>

    <!-- I/O 回放 -->
    <ElDialog v-model="replayVisible" title="会话 I/O 回放（审计录制）" width="760px" top="8vh">
      <div v-if="replaySession" class="replay-meta">
        <ElTag size="small">{{ replaySession.id }}</ElTag>
        <span>{{ replaySession.appName }} / {{ replaySession.agentId }}</span>
        <span>{{ replaySession.username }}</span>
        <span>{{ formatDateTime(replaySession.startAt) }} ~ {{ formatDateTime(replaySession.endAt) }}</span>
      </div>
      <div v-loading="replayLoading" class="replay-body">
        <pre class="replay-output">{{ replaySession?.ioLog || '暂无录制内容' }}</pre>
      </div>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
  .sessions-tab {
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 12px;

    .tab-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 10px;

      .toolbar-title {
        font-size: 13px;
        font-weight: 600;
      }
    }

    .replay-meta {
      display: flex;
      align-items: center;
      gap: 10px;
      font-size: 12px;
      color: var(--el-text-color-secondary);
      margin-bottom: 10px;
      flex-wrap: wrap;
    }

    .replay-body {
      .replay-output {
        background: #1e1e1e;
        color: #d4d4d4;
        padding: 12px;
        border-radius: 4px;
        font-size: 12px;
        line-height: 1.6;
        max-height: 60vh;
        overflow: auto;
        white-space: pre-wrap;
        word-break: break-all;
        margin: 0;
      }
    }
  }
</style>
