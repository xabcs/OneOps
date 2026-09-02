import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import { SetupStoreId } from '@/enum';
import type {
  DiagnosticAgent,
  DiagnosticAppSummary,
  DiagnosticCommandCatalogItem
} from '@/service/api/diagnostic';
import { fetchDiagnosticAppAgents, fetchDiagnosticApps, fetchDiagnosticCommands } from '@/service/api/diagnostic';

/**
 * 诊断中心上下文 Store
 * 跨顶部 Tab（工作台/会话/历史/总览/治理）共享的诊断上下文：
 * 目标选择（应用 → 实例）与命令目录缓存
 */
export const useDiagnosticStore = defineStore(SetupStoreId.Diagnostic, () => {
  // ========== 命令目录（全局缓存，治理 Tab 修改后需刷新） ==========
  const commandCatalog = ref<DiagnosticCommandCatalogItem[]>([]);
  const catalogLoading = ref(false);
  const catalogLoaded = ref(false);

  // ========== 目标选择上下文 ==========
  /** 选中的应用 */
  const currentApp = ref<DiagnosticAppSummary | null>(null);
  /** 选中应用下的实例列表 */
  const agents = ref<DiagnosticAgent[]>([]);
  /** 选中的实例（agent）——诊断执行目标 */
  const currentAgent = ref<DiagnosticAgent | null>(null);
  const agentsLoading = ref(false);

  /** 应用列表（工作台左侧 + 总览共用） */
  const apps = ref<DiagnosticAppSummary[]>([]);
  const appsLoading = ref(false);

  // ========== 计算属性 ==========
  const onlineAgents = computed(() => agents.value.filter(a => a.online));
  /** 当前目标是否可诊断（已选且在线） */
  const hasTarget = computed(() => Boolean(currentAgent.value?.online));

  // ========== 动作 ==========

  /** 加载命令目录（默认带缓存，force 用于治理变更后刷新） */
  async function loadCommandCatalog(force = false) {
    if (catalogLoaded.value && !force) return;
    catalogLoading.value = true;
    try {
      const { data, error } = await fetchDiagnosticCommands();
      if (!error && data) {
        commandCatalog.value = data;
        catalogLoaded.value = true;
      }
    } finally {
      catalogLoading.value = false;
    }
  }

  /** 加载应用列表（同步 tunnel agent 注册表） */
  async function loadApps() {
    appsLoading.value = true;
    try {
      const { data, error } = await fetchDiagnosticApps();
      if (!error && data) {
        apps.value = data;
        return data;
      }
      return null;
    } finally {
      appsLoading.value = false;
    }
  }

  /** 选择应用并加载其实例列表 */
  async function selectApp(app: DiagnosticAppSummary | null) {
    currentApp.value = app;
    agents.value = [];
    currentAgent.value = null;
    if (!app) return;
    agentsLoading.value = true;
    try {
      const { data, error } = await fetchDiagnosticAppAgents(app.appName);
      if (!error && data) {
        agents.value = data;
      }
    } finally {
      agentsLoading.value = false;
    }
  }

  /** 选择实例（诊断目标） */
  function selectAgent(agent: DiagnosticAgent | null) {
    currentAgent.value = agent;
  }

  /** 刷新当前应用的实例列表（保留选择，若已不在线则清空） */
  async function refreshAgents() {
    if (!currentApp.value) return;
    agentsLoading.value = true;
    try {
      const { data, error } = await fetchDiagnosticAppAgents(currentApp.value.appName);
      if (!error && data) {
        agents.value = data;
        if (currentAgent.value) {
          const fresh = data.find(a => a.agentId === currentAgent.value?.agentId);
          currentAgent.value = fresh ?? null;
        }
      }
    } finally {
      agentsLoading.value = false;
    }
  }

  /** 重置整个上下文 */
  function resetContext() {
    currentApp.value = null;
    agents.value = [];
    currentAgent.value = null;
  }

  return {
    // 命令目录
    commandCatalog,
    catalogLoading,
    catalogLoaded,
    loadCommandCatalog,
    // 应用
    apps,
    appsLoading,
    loadApps,
    currentApp,
    // 实例
    agents,
    agentsLoading,
    onlineAgents,
    currentAgent,
    hasTarget,
    selectApp,
    selectAgent,
    refreshAgents,
    resetContext
  };
});
