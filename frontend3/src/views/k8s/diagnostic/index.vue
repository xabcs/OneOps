<script setup lang="ts">
  import { computed, onMounted, ref, watch, type Component } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { ElMessage } from 'element-plus';
  import { Monitor, Refresh, Setting, SwitchButton, Tickets, Odometer } from '@element-plus/icons-vue';
  import { useAuthStore } from '@/store/modules/auth';
  import { useDiagnosticStore } from '@/store/modules/diagnostic';
  import WorkbenchTab from './modules/WorkbenchTab.vue';
  import OverviewTab from './modules/OverviewTab.vue';
  import SessionsTab from './modules/SessionsTab.vue';
  import HistoryTab from './modules/HistoryTab.vue';
  import GovernanceTab from './modules/GovernanceTab.vue';

  defineOptions({ name: 'K8sDiagnostic' });

  const route = useRoute();
  const router = useRouter();
  const authStore = useAuthStore();
  const diagStore = useDiagnosticStore();

  // ========== 顶部 Tab（页面内导航，query 同步） ==========
  const tabComponents: Record<string, Component> = {
    workbench: WorkbenchTab,
    overview: OverviewTab,
    sessions: SessionsTab,
    history: HistoryTab,
    governance: GovernanceTab
  };

  const allTabs = [
    { key: 'workbench', label: '诊断工作台', icon: Monitor, perm: 'k8s.diagnostic.view' },
    { key: 'overview', label: '总览', icon: Odometer, perm: 'k8s.diagnostic.view' },
    { key: 'sessions', label: '会话管理', icon: SwitchButton, perm: 'k8s.diagnostic.view' },
    { key: 'history', label: '执行历史', icon: Tickets, perm: 'k8s.diagnostic.view' },
    { key: 'governance', label: '接入与治理', icon: Setting, perm: 'k8s.diagnostic.execute' }
  ];

  const visibleTabs = computed(() => allTabs.filter(t => authStore.hasPermission(t.perm)));

  const activeTab = ref('workbench');

  watch(activeTab, tab => {
    if (route.query.tab !== tab) {
      router.replace({ query: { ...route.query, tab } });
    }
  });

  onMounted(async () => {
    const tab = route.query.tab as string;
    if (tab && tabComponents[tab] && visibleTabs.value.some(t => t.key === tab)) {
      activeTab.value = tab;
    }
    // 初始化共享上下文：应用列表 + 命令目录
    diagStore.loadApps();
    diagStore.loadCommandCatalog();
  });

  // ========== 上下文条：目标选择（应用 → 实例） ==========
  const onAppSelect = (appName: string) => {
    const app = diagStore.apps.find(a => a.appName === appName);
    diagStore.selectApp(app ?? null);
  };

  const refreshContext = async () => {
    await Promise.all([diagStore.loadApps(), diagStore.refreshAgents()]);
    ElMessage.success('已刷新诊断目标');
  };
</script>

<template>
  <div class="diagnostic-page">
    <!-- 上下文条：诊断目标选择（跨 Tab 共享） -->
    <div class="context-bar">
      <div class="context-title">应用诊断中心</div>
      <div class="context-selects">
        <ElSelect
          :model-value="diagStore.currentApp?.appName ?? ''"
          placeholder="选择应用"
          filterable
          clearable
          class="app-select"
          :loading="diagStore.appsLoading"
          @change="onAppSelect"
        >
          <ElOption
            v-for="app in diagStore.apps"
            :key="app.appName"
            :label="`${app.appName}（${app.online}/${app.total} 在线）`"
            :value="app.appName"
          />
        </ElSelect>

        <ElSelect
          :model-value="diagStore.currentAgent?.agentId ?? ''"
          placeholder="选择实例（Pod）"
          filterable
          clearable
          class="agent-select"
          :disabled="!diagStore.currentApp"
          :loading="diagStore.agentsLoading"
          @change="(val: string) => diagStore.selectAgent(diagStore.agents.find(a => a.agentId === val) ?? null)"
        >
          <ElOption
            v-for="agent in diagStore.agents"
            :key="agent.agentId"
            :label="agent.podName || agent.agentId"
            :value="agent.agentId"
          >
            <span class="agent-option">
              <span>{{ agent.podName || agent.agentId }}</span>
              <span :class="['agent-dot', { online: agent.online }]">{{ agent.online ? '在线' : '离线' }}</span>
            </span>
          </ElOption>
        </ElSelect>

        <ElTag v-if="diagStore.currentAgent" :type="diagStore.currentAgent.online ? 'success' : 'danger'" size="small">
          {{ diagStore.currentAgent.online ? '可诊断' : 'Agent 离线' }}
        </ElTag>
        <ElTag v-if="diagStore.currentAgent" type="info" size="small">
          {{ diagStore.currentAgent.namespace }}
        </ElTag>
      </div>

      <div class="context-actions">
        <ElButton :icon="Refresh" size="small" @click="refreshContext">刷新</ElButton>
      </div>
    </div>

    <!-- 页面内顶部 Tab -->
    <ElTabs v-model="activeTab" class="diag-tabs">
      <ElTabPane v-for="t in visibleTabs" :key="t.key" :name="t.key">
        <template #label>
          <span class="tab-label">
            <ElIcon :size="14"><component :is="t.icon" /></ElIcon>
            {{ t.label }}
          </span>
        </template>
      </ElTabPane>
    </ElTabs>

    <!-- Tab 内容（keep-alive 缓存，切换不丢状态） -->
    <div class="tab-body">
      <KeepAlive>
        <component :is="tabComponents[activeTab]" />
      </KeepAlive>
    </div>
  </div>
</template>

<style scoped lang="scss">
  .diagnostic-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 12px 16px;
    box-sizing: border-box;
    overflow: hidden;

    .context-bar {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 10px 14px;
      background: var(--el-bg-color);
      border: 1px solid var(--el-border-color-light);
      border-radius: 6px;
      margin-bottom: 10px;
      flex-shrink: 0;

      .context-title {
        font-weight: 600;
        font-size: 15px;
        white-space: nowrap;
      }

      .context-selects {
        display: flex;
        align-items: center;
        gap: 10px;
        flex: 1;
        min-width: 0;

        .app-select {
          width: 280px;
        }

        .agent-select {
          width: 300px;
        }

        .agent-option {
          display: flex;
          justify-content: space-between;
          align-items: center;
          width: 100%;

          .agent-dot {
            font-size: 12px;
            color: var(--el-color-danger);

            &.online {
              color: var(--el-color-success);
            }
          }
        }
      }
    }

    .diag-tabs {
      flex-shrink: 0;

      .tab-label {
        display: inline-flex;
        align-items: center;
        gap: 4px;
      }

      :deep(.el-tabs__header) {
        margin-bottom: 0;
      }

      :deep(.el-tabs__nav-wrap::after) {
        height: 1px;
      }
    }

    .tab-body {
      flex: 1;
      min-height: 0;
      overflow: auto;
      padding: 12px 4px;
    }
  }
</style>
