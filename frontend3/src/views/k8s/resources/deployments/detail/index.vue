<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { ElButton, ElMessage, ElMessageBox, ElTabPane, ElTabs, ElTag } from 'element-plus';
  import * as yaml from 'js-yaml';
  import {
    deleteK8sDeployment,
    fetchK8sEvents,
    getK8sDeployment,
    getK8sDeploymentPods,
    restartK8sDeployment,
    updateK8sDeployment
  } from '@/service/api/k8s';
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
  import ScaleDialog from '../modules/ScaleDialog.vue';

  defineOptions({ name: 'K8sDeploymentDetail' });

  const route = useRoute();
  const router = useRouter();

  const deployment = ref<K8s.Deployment | null>(null);
  const pods = ref<K8s.Pod[]>([]);
  const events = ref<K8s.Event[]>([]);
  const loading = ref(true);
  const podsLoading = ref(false);
  const eventsLoading = ref(false);
  const activeTab = ref('pods');

  const clusterId = ref<number>(0);
  const namespace = ref('');
  const deploymentName = ref('');

  // 缩放相关
  const showScaleModal = ref(false);

  // YAML 编辑相关
  const showYamlEditor = ref(false);
  const yamlContent = ref('');

  // 返回列表页的路径 - 直接返回工作负载汇总页，状态由 store 管理
  const backPath = '/k8s/workloads';

  // 状态显示
  const getStatusTag = computed(() => {
    if (!deployment.value) return { type: 'info', text: '未知' };

    const ready = deployment.value.ready || 0;
    const total = deployment.value.replicas || 0;

    if (total === 0) {
      return { type: 'info', text: '未就绪' };
    }

    if (ready === total) {
      return { type: 'success', text: `运行中 (${ready}/${total})` };
    } else if (ready > 0) {
      return { type: 'warning', text: `部分就绪 (${ready}/${total})` };
    }
    return { type: 'danger', text: '未就绪' };
  });

  // 获取副本状态类型
  const getReplicaStatusType = (ready: number, total: number): 'success' | 'warning' | 'danger' | 'info' => {
    if (total === 0) return 'info';
    if (ready === total) return 'success';
    if (ready > 0) return 'warning';
    return 'danger';
  };

  // 基本信息
  const basicInfoFields = computed(() => {
    if (!deployment.value) return [[]];

    const fields = [
      [
        { label: '名称', value: deployment.value.name || '-' },
        { label: '命名空间', value: deployment.value.namespace || '-' }
      ],
      [{ label: '创建时间', value: deployment.value.age || '-' }]
    ];

    if (deployment.value.labels) {
      fields.push([{ label: '标签', value: formatLabels(deployment.value.labels), fullRow: true, isTags: true }]);
    }

    if (deployment.value.selector && Object.keys(deployment.value.selector).length > 0) {
      fields.push([{ label: '选择器', value: formatSelectors(deployment.value.selector), fullRow: true }]);
    }

    if (deployment.value.strategy) {
      fields.push([
        { label: '更新策略', value: formatStrategy(deployment.value.strategy), fullRow: false },
        { label: '策略类型', value: deployment.value.strategy.type || '-', fullRow: false }
      ]);
    }

    if (deployment.value.annotations && Object.keys(deployment.value.annotations).length > 0) {
      fields.push([{ label: '注解', value: formattedAnnotations.value, fullRow: true, isAnnotations: true }]);
    }

    // 状态条件 + 副本状态信息
    const statusConditions = deployment.value.conditions?.length ? formatConditions(deployment.value.conditions) : [];

    // 状态摘要：聚合显示副本状态
    const ready = deployment.value.ready || 0;
    const total = deployment.value.replicas || 0;
    const statusSummary = [
      { type: '副本', value: `${ready}/${total}`, status: getReplicaStatusType(ready, total) },
      { type: '可用', value: deployment.value.available?.toString() || '0', status: 'info' },
      { type: '已更新', value: deployment.value.upToDate?.toString() || '0', status: 'info' }
    ];

    if (statusConditions.length > 0) {
      fields.push([{ label: '状态条件', value: statusConditions, fullRow: true, isConditions: true }]);
    }

    if (statusSummary.length > 0) {
      fields.push([{ label: '副本状态', value: statusSummary, fullRow: true, isStatusSummary: true }]);
    }

    return fields;
  });

  // 格式化注解数据
  const formattedAnnotations = computed(() => {
    if (!deployment.value?.annotations) return [];
    const result = formatAnnotations(deployment.value.annotations);
    return [...result.user, ...result.system];
  });

  // 加载 Deployment 详情
  async function loadDeploymentDetail() {
    const queryClusterId = route.query.clusterId as string;
    const queryNamespace = route.query.namespace as string;
    const queryName = route.query.name as string;

    if (!queryClusterId || !queryNamespace || !queryName) {
      ElMessage.error('参数不完整');
      router.push('/k8s/workloads');
      return;
    }

    clusterId.value = Number(queryClusterId);
    namespace.value = queryNamespace;
    deploymentName.value = queryName;

    loading.value = true;
    // 容器组参数全部来自 route.query，与详情数据无依赖，与详情请求并行加载
    if (activeTab.value === 'pods') {
      loadPods();
    }
    try {
      const res = await getK8sDeployment(clusterId.value, namespace.value, deploymentName.value);
      deployment.value = res.data || res;
    } catch (error: unknown) {
      console.error('获取 Deployment 详情失败:', error);
      const err = error as Error;
      ElMessage.error(err.message || '获取 Deployment 详情失败');
    } finally {
      loading.value = false;
    }
  }

  // 加载关联的 Pods
  async function loadPods() {
    if (!clusterId.value) return;

    podsLoading.value = true;
    try {
      const res = await getK8sDeploymentPods(clusterId.value, namespace.value, deploymentName.value);
      const podData = res.data || res;
      pods.value = Array.isArray(podData) ? podData : [];
    } catch (error: unknown) {
      console.error('获取 Pods 失败:', error);
      const err = error as Error;
      ElMessage.error(err.message || '获取 Pods 失败');
      pods.value = [];
    } finally {
      podsLoading.value = false;
    }
  }

  // 加载事件
  async function loadEvents() {
    if (!clusterId.value) return;

    eventsLoading.value = true;
    try {
      // 使用 fieldSelector 过滤与当前 Deployment 相关的事件
      const fieldSelector = `involvedObject.name=${deploymentName.value},involvedObject.kind=Deployment`;
      const res = await fetchK8sEvents(clusterId.value, namespace.value, fieldSelector);
      const eventData = res.data || res;
      events.value = Array.isArray(eventData) ? eventData : [];
    } catch (error: unknown) {
      console.error('获取 Events 失败:', error);
      const err = error as Error;
      ElMessage.error(err.message || '获取 Events 失败');
      events.value = [];
    } finally {
      eventsLoading.value = false;
    }
  }

  // 标签页切换
  async function handleTabChange(tab: string) {
    activeTab.value = tab;
    if (tab === 'pods' && pods.value.length === 0) {
      await loadPods();
    } else if (tab === 'events' && events.value.length === 0) {
      await loadEvents();
    }
  }

  // 刷新 Pods
  function handleRefreshPods() {
    loadPods();
  }

  // 缩放 Deployment
  function handleScale() {
    if (!deployment.value) return;
    showScaleModal.value = true;
  }

  // 重启 Deployment
  async function handleRestart() {
    try {
      await ElMessageBox.confirm(`确定要重启 Deployment "${deploymentName.value}" 吗？`, '确认重启', {
        type: 'warning'
      });

      await restartK8sDeployment(clusterId.value, {
        namespace: namespace.value,
        name: deploymentName.value
      });
      ElMessage.success('重启成功');
      loadDeploymentDetail();
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        ElMessage.error(err.message || '重启失败');
      }
    }
  }

  // 删除 Deployment
  async function handleDelete() {
    try {
      await ElMessageBox.confirm(
        `确定要删除 Deployment "${deploymentName.value}" 吗？此操作危险，请谨慎操作。`,
        '确认删除',
        {
          type: 'warning',
          confirmButtonText: '危险操作确认',
          cancelButtonText: '取消'
        }
      );

      await deleteK8sDeployment(clusterId.value, {
        namespace: namespace.value,
        name: deploymentName.value
      });
      ElMessage.success('删除成功');
      router.push('/k8s/workloads');
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        ElMessage.error(err.message || '删除失败');
      }
    }
  }

  // 打开 YAML 编辑器
  function handleEditYaml() {
    try {
      // 后端返回的是 JSON 字符串，需要先解析为对象
      const manifestObj = deployment.value?.manifest ? JSON.parse(deployment.value.manifest) : {};
      // 转换为格式化的 YAML
      yamlContent.value = yaml.dump(manifestObj, {
        indent: 2,
        lineWidth: -1,
        noRefs: true
      });
      showYamlEditor.value = true;
    } catch (error) {
      console.error('解析 manifest 失败:', error);
      ElMessage.error('解析 YAML 失败');
    }
  }

  // YAML 编辑保存：统一走 useYamlEdit（YAML 校验 → 更新 API → 重载详情）
  const { handleYamlApply } = useYamlEdit({
    getClusterId: () => clusterId.value,
    getNamespace: () => namespace.value,
    updateFn: updateK8sDeployment,
    reload: loadDeploymentDetail
  });

  onMounted(() => {
    loadDeploymentDetail();
  });
</script>

<template>
  <div v-loading="loading" class="detail-page">
    <!-- 顶部操作栏 - 使用新组件 -->
    <K8sResourceActionBar
      :name="deployment?.name || '-'"
      :namespace="deployment?.namespace || '-'"
      :status-tag="getStatusTag"
      :back-path="backPath"
      :actions="[
        { label: 'YAML 编辑', handler: handleEditYaml, tooltip: '编辑 YAML 配置' },
        { label: '缩放', type: 'primary', handler: handleScale, tooltip: '调整副本数量' },
        { label: '重启', type: 'warning', handler: handleRestart, tooltip: '滚动重启所有 Pod' },
        { label: '删除', type: 'danger', handler: handleDelete, tooltip: '删除 Deployment（危险操作）' }
      ]"
    />

    <!-- 基本信息区域 -->
    <div class="basic-info-section">
      <div class="info-section">
        <h3 class="section-title">基本信息</h3>
        <K8sBasicInfoGrid :fields="basicInfoFields" />
      </div>
    </div>

    <!-- 标签切换 -->
    <ElTabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <!-- 容器组 -->
      <ElTabPane :label="`容器组 (${pods.length})`" name="pods">
        <template #label>
          <div class="tab-pane-header">
            <span>容器组</span>
            <ElTag size="small" class="count-tag">{{ pods.length }}</ElTag>
            <ElButton size="small" text class="refresh-btn" @click="handleRefreshPods">
              <icon-mdi-refresh class="text-14px" />
            </ElButton>
          </div>
        </template>
        <div class="tab-content">
          <K8sPodsTable :pods="pods" :loading="podsLoading" :cluster-id="clusterId" :namespace="namespace" />
        </div>
      </ElTabPane>

      <!-- 事件 -->
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

    <!-- 缩放对话框 -->
    <ScaleDialog
      v-model:visible="showScaleModal"
      :cluster-id="clusterId"
      :namespace="namespace"
      :deployment="deploymentName"
      @submitted="loadDeploymentDetail"
    />

    <!-- YAML 编辑器 -->
    <YamlEditor
      v-model="showYamlEditor"
      :title="`编辑 ${deploymentName}`"
      :yaml="yamlContent"
      :can-edit="true"
      :on-apply="handleYamlApply"
    />
  </div>
</template>

<style scoped>
  /* 页面容器 */
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

  /* tab工具栏 - 保持原有样式但优化间距 */
  .tab-toolbar {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    padding: 8px 16px;
    background: transparent;
    border-bottom: 1px solid #ebeef5;
  }

  /* 标签内容容器 - 移除多余的padding */
  .tab-content {
    padding: 12px 0;
  }

  /* 基本信息区域 - 透明背景，无边框 */
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

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  /* 描述网格 */
  .desc-grid {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .desc-row {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 24px;
  }

  .desc-item {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .desc-item-full {
    grid-column: 1 / -1;
  }

  .item-label {
    font-size: 12px;
    color: #909399;
  }

  .item-content {
    font-size: 14px;
    color: #303133;
  }

  .value-text {
    color: #303133;
  }

  /* 镜像列换行显示 */
  .whitespace-pre-line {
    white-space: pre-line;
    word-break: break-all;
  }

  .font-mono {
    font-family: 'Courier New', Courier, monospace;
    word-break: break-all;
  }

  /* 标签列表 */
  .tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
