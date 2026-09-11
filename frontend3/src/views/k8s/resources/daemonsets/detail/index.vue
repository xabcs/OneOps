<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { ElButton, ElMessage, ElTabPane, ElTabs, ElTag } from 'element-plus';
  import yaml from 'js-yaml';
  import { fetchK8sEvents, getK8sDaemonSet, getK8sDaemonSetPods, updateK8sDaemonSet } from '@/service/api/k8s';
  import {
    formatAnnotations,
    formatConditions,
    formatLabels,
    formatSelectors,
    formatStrategy
  } from '@/views/k8s/shared/k8s-formatters';
  import { useYamlEdit } from '@/views/k8s/composables/useYamlEdit';
  import K8sBasicInfoGrid from '@/components/k8s/K8sBasicInfoGrid.vue';
  import K8sPodsTable from '@/components/k8s/K8sPodsTable.vue';
  import K8sEventsTable from '@/components/k8s/K8sEventsTable.vue';
  import K8sResourceActionBar from '@/components/k8s/K8sResourceActionBar.vue';
  import YamlEditor from '@/components/k8s/YamlEditor.vue';

  defineOptions({ name: 'K8sDaemonSetDetail' });

  const route = useRoute();
  const router = useRouter();

  const resource = ref<K8s.DaemonSet | null>(null);
  const pods = ref<K8s.Pod[]>([]);
  const events = ref<K8s.Event[]>([]);
  const loading = ref(true);
  const podsLoading = ref(false);
  const eventsLoading = ref(false);
  const activeTab = ref('pods');

  const clusterId = ref<number>(0);
  const namespace = ref('');
  const resourceName = ref('');

  // YAML 编辑相关
  const showYamlEditor = ref(false);
  const yamlContent = ref('');

  const getStatusTag = computed(() => {
    if (!resource.value) return { type: 'info', text: '未知' };
    const ready = resource.value.ready || 0;
    const desired = resource.value.desired || 0;
    if (desired === 0) return { type: 'info', text: '未就绪' };
    if (ready === desired) return { type: 'success', text: `运行中 (${ready}/${desired})` };
    if (ready > 0) return { type: 'warning', text: `部分就绪 (${ready}/${desired})` };
    return { type: 'danger', text: '未就绪' };
  });

  // 返回列表页的路径 - 直接返回工作负载汇总页，状态由 store 管理
  const backPath = '/k8s/workloads';

  // 获取副本状态类型
  const getReplicaStatusType = (ready: number, total: number): 'success' | 'warning' | 'danger' | 'info' => {
    if (total === 0) return 'info';
    if (ready === total) return 'success';
    if (ready > 0) return 'warning';
    return 'danger';
  };

  // 基本信息
  const basicInfoFields = computed(() => {
    if (!resource.value) return [[]];

    const fields = [
      [
        { label: '名称', value: resource.value.name || '-' },
        { label: '命名空间', value: resource.value.namespace || '-' }
      ],
      [{ label: '创建时间', value: resource.value.age || '-' }]
    ];

    if (resource.value.labels) {
      fields.push([{ label: '标签', value: formatLabels(resource.value.labels), fullRow: true, isTags: true }]);
    }

    if (resource.value.selector && Object.keys(resource.value.selector).length > 0) {
      fields.push([{ label: '选择器', value: formatSelectors(resource.value.selector), fullRow: true }]);
    }

    if (resource.value.strategy) {
      fields.push([
        { label: '更新策略', value: formatStrategy(resource.value.strategy), fullRow: false },
        { label: '策略类型', value: resource.value.strategy.type || '-', fullRow: false }
      ]);
    }

    if (resource.value.annotations && Object.keys(resource.value.annotations).length > 0) {
      fields.push([{ label: '注解', value: formattedAnnotations.value, fullRow: true, isAnnotations: true }]);
    }

    // 状态条件 + 节点状态信息
    const statusConditions = resource.value.conditions?.length ? formatConditions(resource.value.conditions) : [];

    // 状态摘要：聚合显示节点状态
    const ready = resource.value.ready || 0;
    const desired = resource.value.desired || 0;
    const statusSummary = [
      { type: '就绪/期望', value: `${ready}/${desired}`, status: getReplicaStatusType(ready, desired) },
      { type: '当前节点', value: resource.value.current?.toString() || '0', status: 'info' },
      { type: '已更新', value: resource.value.updated?.toString() || '0', status: 'info' },
      { type: '可用', value: resource.value.available?.toString() || '0', status: 'info' }
    ];

    if (statusConditions.length > 0) {
      fields.push([{ label: '状态条件', value: statusConditions, fullRow: true, isConditions: true }]);
    }

    if (statusSummary.length > 0) {
      fields.push([{ label: '节点状态', value: statusSummary, fullRow: true, isStatusSummary: true }]);
    }

    return fields;
  });

  // 格式化注解数据
  const formattedAnnotations = computed(() => {
    if (!resource.value?.annotations) return [];
    const result = formatAnnotations(resource.value.annotations);
    return [...result.user, ...result.system];
  });

  async function loadDetail() {
    const qClusterId = route.query.clusterId as string;
    const qNamespace = route.query.namespace as string;
    const qName = route.query.name as string;

    if (!qClusterId || !qNamespace || !qName) {
      ElMessage.error('参数不完整');
      router.push('/k8s/workloads');
      return;
    }

    clusterId.value = Number(qClusterId);
    namespace.value = qNamespace;
    resourceName.value = qName;

    loading.value = true;
    // 容器组参数全部来自 route.query，与详情数据无依赖，与详情请求并行加载
    if (activeTab.value === 'pods') {
      loadPods();
    }
    try {
      const res = await getK8sDaemonSet(clusterId.value, namespace.value, resourceName.value);
      resource.value = res.data || res;
      // 解析 manifest 为 YAML
      if (resource.value?.manifest) {
        try {
          const manifestObj = JSON.parse(resource.value.manifest);
          yamlContent.value = yaml.dump(manifestObj, { indent: 2, lineWidth: 120, noRefs: true });
        } catch (e) {
          yamlContent.value = resource.value.manifest;
        }
      }
    } catch (error: unknown) {
      const err = error as Error;
      ElMessage.error(err.message || '获取详情失败');
    } finally {
      loading.value = false;
    }
  }

  async function loadPods() {
    if (!clusterId.value) return;
    podsLoading.value = true;
    try {
      const res = await getK8sDaemonSetPods(clusterId.value, namespace.value, resourceName.value);
      const podData = res.data || res;
      pods.value = Array.isArray(podData) ? podData : [];
    } catch (error: unknown) {
      const err = error as Error;
      ElMessage.error(err.message || '获取 Pods 失败');
      pods.value = [];
    } finally {
      podsLoading.value = false;
    }
  }

  async function loadEvents() {
    if (!clusterId.value) return;
    eventsLoading.value = true;
    try {
      const fieldSelector = `involvedObject.name=${resourceName.value},involvedObject.kind=DaemonSet`;
      const res = await fetchK8sEvents(clusterId.value, namespace.value, fieldSelector);
      const eventData = res.data || res;
      events.value = Array.isArray(eventData) ? eventData : [];
    } catch (error: unknown) {
      const err = error as Error;
      ElMessage.error(err.message || '获取 Events 失败');
      events.value = [];
    } finally {
      eventsLoading.value = false;
    }
  }

  async function handleTabChange(tab: string) {
    activeTab.value = tab;
    if (tab === 'pods' && pods.value.length === 0) await loadPods();
    if (tab === 'events' && events.value.length === 0) await loadEvents();
  }

  // 打开 YAML 编辑器
  function handleEditYaml() {
    showYamlEditor.value = true;
  }

  // YAML 编辑保存：统一走 useYamlEdit（YAML 校验 → 更新 API → 重载详情）
  const { handleYamlApply } = useYamlEdit({
    getClusterId: () => clusterId.value,
    getNamespace: () => namespace.value,
    updateFn: updateK8sDaemonSet,
    reload: loadDetail,
    successMessage: '更新成功'
  });

  onMounted(() => {
    loadDetail();
  });
</script>

<template>
  <div v-loading="loading" class="detail-page">
    <!-- 顶部操作栏 - 使用新组件 -->
    <K8sResourceActionBar
      :name="resource?.name || '-'"
      :namespace="resource?.namespace || '-'"
      :status-tag="getStatusTag"
      :back-path="backPath"
      :actions="[{ label: 'YAML 编辑', handler: handleEditYaml, tooltip: '编辑 YAML 配置' }]"
    />

    <!-- 基本信息 -->
    <div class="basic-info-section">
      <div class="info-section">
        <h3 class="section-title">基本信息</h3>
        <K8sBasicInfoGrid :fields="basicInfoFields" />
      </div>
    </div>

    <!-- 标签切换 -->
    <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <ElTabPane :label="`容器组 (${pods.length})`" name="pods">
        <template #label>
          <div class="tab-pane-header">
            <span>容器组</span>
            <ElTag size="small" class="count-tag">{{ pods.length }}</ElTag>
            <ElButton size="small" text class="refresh-btn" @click="loadPods">
              <icon-mdi-refresh class="text-14px" />
            </ElButton>
          </div>
        </template>
        <div class="tab-content">
          <K8sPodsTable :pods="pods" :loading="podsLoading" :cluster-id="clusterId" :namespace="namespace" />
        </div>
      </ElTabPane>

      <ElTabPane :label="`事件 (${events.length})`" name="events">
        <template #label>
          <div class="tab-pane-header">
            <span>事件</span>
            <ElTag size="small" class="count-tag">{{ events.length }}</ElTag>
            <ElButton size="small" text class="refresh-btn" @click="loadEvents">
              <icon-mdi-refresh class="text-14px" />
            </ElButton>
          </div>
        </template>
        <div class="tab-content">
          <K8sEventsTable :events="events" :loading="eventsLoading" />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- YAML 编辑器 -->
    <YamlEditor
      v-model="showYamlEditor"
      :title="`编辑 ${resourceName}`"
      :yaml="yamlContent"
      :can-edit="true"
      :on-apply="handleYamlApply"
    />
  </div>
</template>

<style scoped>
  .detail-page {
    min-height: 100vh;
    padding: 24px;
  }

  /* 标签页样式 - 优化与tab的衔接 */
  .detail-tabs {
    margin-top: 16px;
  }

  /* 移除 tab内容默认的padding，让内容更紧凑 */
  .detail-tabs :deep(.el-tabs__content) {
    padding: 0;
  }

  /* tab标题栏 - 自定义样式 */
  .tab-pane-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 8px;
    line-height: 1;
  }

  /* 计数标签 */
  .tab-pane-header .count-tag {
    background: #f0f2f5;
    color: #606266;
    border: none;
    font-size: 12px;
    height: 18px;
    line-height: 18px;
    padding: 0 6px;
    font-weight: 500;
  }

  /* 刷新按钮 - 内联样式 */
  .tab-pane-header .refresh-btn {
    padding: 4px;
    margin-left: auto;
  }

  .tab-pane-header .refresh-btn:hover {
    background: transparent;
  }

  .tab-toolbar {
    display: flex;
    justify-content: flex-end;
    padding: 0 16px 8px;
  }

  .basic-info-section {
    background: transparent;
    margin-bottom: 24px;
  }

  .info-section {
    background: transparent;
    margin-bottom: 16px;
  }

  .section-title {
    font-size: 14px;
    font-weight: 500;
    color: #303133;
    margin-bottom: 12px;
  }

  .tab-content {
    padding: 16px;
  }

  .detail-tabs {
    margin-top: 16px;
  }
</style>
