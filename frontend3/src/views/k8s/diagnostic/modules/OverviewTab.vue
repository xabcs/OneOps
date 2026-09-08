<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { Refresh } from '@element-plus/icons-vue';
  import { type DiagnosticOverviewData, fetchDiagnosticOverview } from '@/service/api/diagnostic';
  import { useDiagnosticStore } from '@/store/modules/diagnostic';

  defineOptions({ name: 'DiagnosticOverviewTab' });

  const route = useRoute();
  const router = useRouter();
  const diagStore = useDiagnosticStore();

  const loading = ref(false);
  const overview = ref<DiagnosticOverviewData | null>(null);

  async function load() {
    loading.value = true;
    try {
      const { data, error } = await fetchDiagnosticOverview();
      if (!error && data) {
        overview.value = data;
      }
    } finally {
      loading.value = false;
    }
  }

  /** 跳工作台并选中该应用 */
  function gotoWorkbench(appName: string) {
    const app = diagStore.apps.find(a => a.appName === appName);
    diagStore.selectApp(app ?? null);
    router.replace({ query: { ...route.query, tab: 'workbench' } });
  }

  const cards = [
    { key: 'apps', label: '接入应用', type: 'primary' },
    { key: 'agentOnline', label: '在线 Agent', type: 'success' },
    { key: 'agentTotal', label: 'Agent 总数', type: 'info' },
    { key: 'total', label: '累计执行', type: 'info' },
    { key: 'highRisk', label: '高危执行', type: 'warning' },
    { key: 'errorCount', label: '执行失败', type: 'danger' },
    { key: 'activeSess', label: '活跃会话', type: 'success' }
  ] as const;

  onMounted(load);
</script>

<template>
  <div v-loading="loading" class="overview-tab">
    <!-- 统计卡片 -->
    <div class="stat-cards">
      <div v-for="card in cards" :key="card.key" class="stat-card" :class="[card.type]">
        <div class="stat-value">
          {{
            card.key === 'apps'
              ? (overview?.apps ?? 0)
              : card.key === 'agentOnline'
                ? (overview?.agentOnline ?? 0)
                : card.key === 'agentTotal'
                  ? (overview?.agentTotal ?? 0)
                  : (overview?.execStats?.[card.key] ?? 0)
          }}
        </div>
        <div class="stat-label">{{ card.label }}</div>
      </div>
    </div>

    <!-- 应用接入表 -->
    <div class="app-section">
      <div class="section-head">
        <span class="section-title">应用接入情况</span>
        <ElButton :icon="Refresh" size="small" @click="load">刷新</ElButton>
      </div>
      <ElTable :data="overview?.appList ?? []" size="small">
        <ElTableColumn prop="appName" label="应用" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="实例在线" width="200">
          <template #default="{ row }">
            <ElProgress
              :percentage="row.total ? Math.round((row.online / row.total) * 100) : 0"
              :status="row.online === row.total ? 'success' : row.online === 0 ? 'exception' : 'warning'"
            />
            <span class="progress-text">{{ row.online }}/{{ row.total }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="agentVersion" label="Agent 版本" width="140">
          <template #default="{ row }">{{ row.agentVersion || '-' }}</template>
        </ElTableColumn>
        <ElTableColumn label="操作" align="center" width="120" fixed="right" class-name="msre-table-actions">
          <template #default="{ row }">
            <ElButton link type="primary" size="small" @click="gotoWorkbench(row.appName)">去诊断</ElButton>
          </template>
        </ElTableColumn>
        <template #empty>
          <ElEmpty description="暂无应用接入（请参考「接入与治理」为 Java 服务注入 Arthas agent）" :image-size="80" />
        </template>
      </ElTable>
    </div>
  </div>
</template>

<style scoped lang="scss">
  .overview-tab {
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 14px;

    .stat-cards {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
      gap: 10px;
      margin-bottom: 14px;

      .stat-card {
        border-radius: 6px;
        padding: 14px 16px;
        border: 1px solid var(--el-border-color-light);
        background: var(--el-fill-color-extra-light);

        &.primary {
          border-left: 3px solid var(--el-color-primary);
        }

        &.success {
          border-left: 3px solid var(--el-color-success);
        }

        &.warning {
          border-left: 3px solid var(--el-color-warning);
        }

        &.danger {
          border-left: 3px solid var(--el-color-danger);
        }

        &.info {
          border-left: 3px solid var(--el-color-info);
        }

        .stat-value {
          font-size: 26px;
          font-weight: 700;
          line-height: 1.2;
        }

        .stat-label {
          font-size: 12px;
          color: var(--el-text-color-secondary);
          margin-top: 4px;
        }
      }
    }

    .app-section {
      .section-head {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .section-title {
          font-size: 13px;
          font-weight: 600;
        }
      }

      .progress-text {
        font-size: 12px;
        color: var(--el-text-color-secondary);
        margin-left: 8px;
      }
    }
  }
</style>
