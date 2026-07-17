<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  ElPagination,
  ElTabPane,
  ElTable,
  ElTableColumn,
  ElTabs,
  ElTag,
  type FormInstance
} from 'element-plus';
import yaml from 'js-yaml';
import {
  deleteK8sCronJob,
  deleteK8sDaemonSet,
  deleteK8sDeployment,
  deleteK8sJob,
  deleteK8sPod,
  deleteK8sStatefulSet,
  fetchK8sClusterNamespaces,
  fetchK8sClusters,
  fetchK8sCronJobs,
  fetchK8sDaemonSets,
  fetchK8sDeployments,
  fetchK8sJobs,
  fetchK8sPodLogs,
  fetchK8sPods,
  fetchK8sStatefulSets,
  getK8sCronJob,
  getK8sDaemonSet,
  getK8sDeployment,
  getK8sJob,
  getK8sPod,
  getK8sStatefulSet,
  restartK8sDaemonSet,
  restartK8sDeployment,
  restartK8sStatefulSet,
  scaleK8sDeployment,
  suspendK8sCronJob,
  updateK8sCronJob,
  updateK8sDaemonSet,
  updateK8sDeployment,
  updateK8sJob,
  updateK8sStatefulSet
} from '@/service/api/k8s';
import { useK8sStore } from '@/store/modules/k8s';
import YamlEditor from '@/components/YamlEditor.vue';

// 解析资源manifest为YAML格式
function parseManifest(manifestStr: string): string {
  if (!manifestStr) return '';
  try {
    const obj = JSON.parse(manifestStr);
    // 移除managedFields等不需要显示的字段
    delete obj.managedFields;
    return yaml.dump(obj, {
      indent: 2,
      lineWidth: 120,
      noRefs: true,
      sortKeys: false
    });
  } catch (e) {
    console.error('解析manifest失败:', e);
    return manifestStr;
  }
}

defineOptions({ name: 'K8sWorkloads' });

const router = useRouter();
const message = ElMessage;
const k8sStore = useK8sStore();

const loading = ref(false);
const activeTab = ref(k8sStore.workloadFilterState.activeTab || 'deployments');

// 当前选中的集群和命名空间 - 从 store 读取
const selectedCluster = ref(k8sStore.workloadFilterState.clusterId);
const selectedNamespace = ref(k8sStore.workloadFilterState.namespace);

// 可用的命名空间列表
const namespaces = ref<string[]>([]);

// 可用的集群列表
const clusters = ref<any[]>([]);

// 各种资源的数据
const deploymentsData = ref<any[]>([]);
const podsData = ref<any[]>([]);
const statefulSetsData = ref<any[]>([]);
const daemonSetsData = ref<any[]>([]);
const jobsData = ref<any[]>([]);
const cronJobsData = ref<any[]>([]);

// 各种资源的分页
const deploymentsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const podsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const statefulSetsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const daemonSetsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const jobsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const cronJobsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

// 弹窗状态
const showScaleDialog = ref(false);
const scaleFormRef = ref<FormInstance>();
const scaleData = reactive({ replicas: 1 });
const selectedDeployment = ref<any>(null);

// YAML 编辑器状态
const showYamlDialog = ref(false);
const yamlDialogTitle = ref('');
const yamlContent = ref('');
const selectedResource = ref<any>(null);
const yamlLoading = ref(false);

// 日志对话框状态
const showLogDialog = ref(false);
const logDialogContent = ref('');

// 批量操作
const selectedDeployments = ref<any[]>([]);
const selectedPods = ref<any[]>([]);
const selectedStatefulSets = ref<any[]>([]);
const selectedDaemonSets = ref<any[]>([]);
const selectedJobs = ref<any[]>([]);
const selectedCronJobs = ref<any[]>([]);
const batchOperationsVisible = ref(false);

// 计算当前Tab的批量操作是否可见
const showBatchOperations = computed(() => {
  if (activeTab.value === 'deployments') {
    return selectedDeployments.value.length > 0;
  } else if (activeTab.value === 'pods') {
    return selectedPods.value.length > 0;
  } else if (activeTab.value === 'statefulsets') {
    return selectedStatefulSets.value.length > 0;
  } else if (activeTab.value === 'daemonsets') {
    return selectedDaemonSets.value.length > 0;
  } else if (activeTab.value === 'jobs') {
    return selectedJobs.value.length > 0;
  } else if (activeTab.value === 'cronjobs') {
    return selectedCronJobs.value.length > 0;
  }
  return false;
});

// 根据Tab返回对应的数据和分页
const currentData = computed(() => {
  switch (activeTab.value) {
    case 'pods':
      return podsData.value;
    case 'deployments':
      return deploymentsData.value;
    case 'statefulsets':
      return statefulSetsData.value;
    case 'daemonsets':
      return daemonSetsData.value;
    case 'jobs':
      return jobsData.value;
    case 'cronjobs':
      return cronJobsData.value;
    default:
      return [];
  }
});

const currentPagination = computed(() => {
  switch (activeTab.value) {
    case 'pods':
      return podsPagination;
    case 'deployments':
      return deploymentsPagination;
    case 'statefulsets':
      return statefulSetsPagination;
    case 'daemonsets':
      return daemonSetsPagination;
    case 'jobs':
      return jobsPagination;
    case 'cronjobs':
      return cronJobsPagination;
    default:
      return { page: 1, pageSize: 10, itemCount: 0 };
  }
});

// 加载集群列表
async function loadClusters() {
  try {
    const response = await fetchK8sClusters();
    clusters.value = response.data || [];
    // 自动选中第一个集群
    if (clusters.value.length > 0 && !selectedCluster.value) {
      selectedCluster.value = clusters.value[0].id;
    }
  } catch (error) {
    console.error('加载集群列表失败:', error);
  }
}

// 加载命名空间列表
async function loadNamespaces() {
  if (!selectedCluster.value) return;
  try {
    const response = await fetchK8sClusterNamespaces(selectedCluster.value);
    namespaces.value = response.data.map((ns: any) => ns.name);
    // 自动选中第一个命名空间
    if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
      selectedNamespace.value = namespaces.value[0];
    }
  } catch (error) {
    console.error('加载命名空间失败:', error);
  }
}

// 加载各种资源数据
async function loadDeployments() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    console.log('[DEBUG] 加载 Deployments:', {
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      page: deploymentsPagination.page,
      pageSize: deploymentsPagination.pageSize
    });

    const response = await fetchK8sDeployments(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: deploymentsPagination.page,
      pageSize: deploymentsPagination.pageSize
    });

    // 处理不同的响应格式
    // 格式1: {list: [...], total: number} (期望格式)
    // 格式2: {data: {list: [...], total: number}}
    // 格式3: {data: {data: {list: [...], total: number}}}
    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    deploymentsData.value = apiData.list || [];
    deploymentsPagination.itemCount = apiData.total || 0;

    console.log('[DEBUG] Deployments 数据赋值后:', {
      dataLength: deploymentsData.value.length,
      total: deploymentsPagination.itemCount,
      firstItem: deploymentsData.value[0]
    });
  } catch (error) {
    console.error('加载 Deployments 失败:', error);
  } finally {
    loading.value = false;
  }
}

async function loadPods() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    console.log('[DEBUG] 加载 Pods:', {
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      page: podsPagination.page,
      pageSize: podsPagination.pageSize
    });

    const response = await fetchK8sPods(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: podsPagination.page,
      pageSize: podsPagination.pageSize
    });

    // 处理不同的响应格式
    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    podsData.value = apiData.list || [];
    podsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 Pods 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 StatefulSets
async function loadStatefulSets() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    const response = await fetchK8sStatefulSets(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: statefulSetsPagination.page,
      pageSize: statefulSetsPagination.pageSize
    });

    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    statefulSetsData.value = apiData.list || [];
    statefulSetsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 StatefulSets 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 DaemonSets
async function loadDaemonSets() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    const response = await fetchK8sDaemonSets(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: daemonSetsPagination.page,
      pageSize: daemonSetsPagination.pageSize
    });

    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    daemonSetsData.value = apiData.list || [];
    daemonSetsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 DaemonSets 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 Jobs
async function loadJobs() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    const response = await fetchK8sJobs(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: jobsPagination.page,
      pageSize: jobsPagination.pageSize
    });

    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    jobsData.value = apiData.list || [];
    jobsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 Jobs 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 CronJobs
async function loadCronJobs() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    const response = await fetchK8sCronJobs(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: cronJobsPagination.page,
      pageSize: cronJobsPagination.pageSize
    });

    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    cronJobsData.value = apiData.list || [];
    cronJobsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 CronJobs 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 根据当前Tab加载对应数据
async function loadCurrentData() {
  switch (activeTab.value) {
    case 'pods':
      await loadPods();
      break;
    case 'deployments':
      await loadDeployments();
      break;
    case 'statefulsets':
      await loadStatefulSets();
      break;
    case 'daemonsets':
      await loadDaemonSets();
      break;
    case 'jobs':
      await loadJobs();
      break;
    case 'cronjobs':
      await loadCronJobs();
      break;
  }
}

// 跳转到详情页
function goToDetail(row: any) {
  console.log('[工作负载汇总] goToDetail 被调用');
  console.log('[工作负载汇总] 当前标签页:', activeTab.value);
  console.log('[工作负载汇总] 点击的行数据:', row);

  // 状态已经通过 watch 自动同步到 store，无需手动保存
  const baseQuery = { clusterId: selectedCluster.value, namespace: row.namespace, name: row.name };

  // 跳转到详情页
  switch (activeTab.value) {
    case 'deployments':
      router.push({ path: '/k8s/resources/deployments/detail', query: baseQuery });
      break;
    case 'statefulsets':
      router.push({ path: '/k8s/resources/statefulsets/detail', query: baseQuery });
      break;
    case 'daemonsets':
      router.push({ path: '/k8s/resources/daemonsets/detail', query: baseQuery });
      break;
    case 'pods':
      router.push({ path: '/k8s/resources/pods/detail', query: baseQuery });
      break;
    case 'jobs':
      router.push({ path: '/k8s/resources/jobs/detail', query: baseQuery });
      break;
    case 'cronjobs':
      router.push({ path: '/k8s/resources/cronjobs/detail', query: baseQuery });
      break;
  }
}

// Deployment 操作
async function handleEdit(row: any) {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  selectedResource.value = row;
  yamlDialogTitle.value = `编辑 Deployment: ${row.name}`;
  yamlLoading.value = true;

  try {
    const response = await getK8sDeployment(selectedCluster.value, row.namespace, row.name);
    const manifestStr = response.data?.manifest || response.manifest || '';
    // 先设置 YAML 内容，再打开弹窗
    yamlContent.value = parseManifest(manifestStr);
    showYamlDialog.value = true;
  } catch (error: any) {
    message.error(error.message || '获取 YAML 失败');
  } finally {
    yamlLoading.value = false;
  }
}

// StatefulSet 操作
async function handleStatefulSetEdit(row: any) {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  selectedResource.value = row;
  yamlDialogTitle.value = `编辑 StatefulSet: ${row.name}`;
  yamlLoading.value = true;

  try {
    const response = await getK8sStatefulSet(selectedCluster.value, row.namespace, row.name);
    const manifestStr = response.data?.manifest || response.manifest || '';
    // 先设置 YAML 内容，再打开弹窗
    yamlContent.value = parseManifest(manifestStr);
    showYamlDialog.value = true;
  } catch (error: any) {
    message.error(error.message || '获取 YAML 失败');
  } finally {
    yamlLoading.value = false;
  }
}

// DaemonSet 操作
async function handleDaemonSetEdit(row: any) {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  selectedResource.value = row;
  yamlDialogTitle.value = `编辑 DaemonSet: ${row.name}`;
  yamlLoading.value = true;

  try {
    const response = await getK8sDaemonSet(selectedCluster.value, row.namespace, row.name);
    const manifestStr = response.data?.manifest || response.manifest || '';
    // 先设置 YAML 内容，再打开弹窗
    yamlContent.value = parseManifest(manifestStr);
    showYamlDialog.value = true;
  } catch (error: any) {
    message.error(error.message || '获取 YAML 失败');
  } finally {
    yamlLoading.value = false;
  }
}

// Job 操作
async function handleJobEdit(row: any) {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  selectedResource.value = row;
  yamlDialogTitle.value = `编辑 Job: ${row.name}`;
  yamlLoading.value = true;

  try {
    const response = await getK8sJob(selectedCluster.value, row.namespace, row.name);
    const manifestStr = response.data?.manifest || response.manifest || '';
    // 先设置 YAML 内容，再打开弹窗
    yamlContent.value = parseManifest(manifestStr);
    showYamlDialog.value = true;
  } catch (error: any) {
    message.error(error.message || '获取 YAML 失败');
  } finally {
    yamlLoading.value = false;
  }
}

// CronJob 操作
async function handleCronJobEdit(row: any) {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  selectedResource.value = row;
  yamlDialogTitle.value = `编辑 CronJob: ${row.name}`;
  yamlLoading.value = true;

  try {
    const response = await getK8sCronJob(selectedCluster.value, row.namespace, row.name);
    const manifestStr = response.data?.manifest || response.manifest || '';
    // 先设置 YAML 内容，再打开弹窗
    yamlContent.value = parseManifest(manifestStr);
    showYamlDialog.value = true;
  } catch (error: any) {
    message.error(error.message || '获取 YAML 失败');
  } finally {
    yamlLoading.value = false;
  }
}

// YAML 应用
async function handleYamlApply(yamlStr: string) {
  if (!selectedCluster.value || !selectedResource.value) {
    throw new Error('缺少必要参数');
  }

  const resourceType = activeTab.value;

  try {
    // 将 YAML 字符串解析为对象
    const manifestObj = yaml.load(yamlStr);

    switch (resourceType) {
      case 'deployments':
        await updateK8sDeployment(selectedCluster.value, {
          namespace: selectedResource.value.namespace,
          manifest: manifestObj
        });
        await loadDeployments();
        break;
      case 'statefulsets':
        await updateK8sStatefulSet(selectedCluster.value, {
          namespace: selectedResource.value.namespace,
          manifest: manifestObj
        });
        await loadStatefulSets();
        break;
      case 'daemonsets':
        await updateK8sDaemonSet(selectedCluster.value, {
          namespace: selectedResource.value.namespace,
          manifest: manifestObj
        });
        await loadDaemonSets();
        break;
      case 'jobs':
        await updateK8sJob(selectedCluster.value, {
          namespace: selectedResource.value.namespace,
          manifest: manifestObj
        });
        await loadJobs();
        break;
      case 'cronjobs':
        await updateK8sCronJob(selectedCluster.value, {
          namespace: selectedResource.value.namespace,
          manifest: manifestObj
        });
        await loadCronJobs();
        break;
      default:
        throw new Error('不支持的资源类型');
    }
  } catch (error: any) {
    throw new Error(error.message || 'YAML 应用失败');
  }
}

function handleScale(row: any) {
  selectedDeployment.value = row;
  scaleData.replicas = row.replicas || 1;
  showScaleDialog.value = true;
}

async function handleScaleSubmit() {
  if (!selectedDeployment.value || !selectedCluster.value) return;

  try {
    await scaleK8sDeployment(selectedCluster.value, {
      namespace: selectedDeployment.value.namespace,
      name: selectedDeployment.value.name,
      replicas: scaleData.replicas
    });
    message.success('缩放成功');
    showScaleDialog.value = false;
    await loadDeployments();
  } catch (error: any) {
    message.error(error.message || '缩放失败');
  }
}

function handleMoreCommand(command: string, row: any) {
  switch (command) {
    case 'edit':
      handleEdit(row);
      break;
    case 'restart':
      handleRestart(row);
      break;
    case 'delete':
      handleDelete(row);
      break;
  }
}

async function handleRestart(row: any) {
  try {
    await ElMessageBox.confirm(`确定要重启 Deployment "${row.name}" 吗？`, '确认重启', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await restartK8sDeployment(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('重启成功');
    await loadDeployments();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '重启失败');
    }
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除 Deployment "${row.name}" 吗？此操作不可恢复！`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await deleteK8sDeployment(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('删除成功');
    await loadDeployments();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
}

// Pod 操作
async function handlePodLogs(row: any) {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  const containerName = row.containers?.[0]?.name || '';
  selectedResource.value = row;
  logDialogContent.value = '加载中...';
  showLogDialog.value = true;

  try {
    const response = await fetchK8sPodLogs(selectedCluster.value, row.namespace, row.name, {
      container: containerName,
      tailLines: 100
    });
    logDialogContent.value = response.data?.logs || '暂无日志';
  } catch (error: any) {
    logDialogContent.value = `日志加载失败: ${error.message || '未知错误'}`;
    message.error(`日志加载失败: ${error.message}`);
  }
}

// Pod YAML编辑
async function handlePodEdit(row: any) {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  selectedResource.value = row;
  yamlDialogTitle.value = `编辑 YAML: ${row.name}`;
  yamlLoading.value = true;

  try {
    const response = await getK8sPod(selectedCluster.value, row.namespace, row.name);
    const manifestStr = response.data?.manifest || response.manifest || '';
    // 先设置 YAML 内容，再打开弹窗
    yamlContent.value = parseManifest(manifestStr);
    showYamlDialog.value = true;
  } catch (error: any) {
    message.error(error.message || '获取 YAML 失败');
  } finally {
    yamlLoading.value = false;
  }
}

function handlePodMoreCommand(command: string, row: any) {
  switch (command) {
    case 'edit':
      handlePodEdit(row);
      break;
    case 'terminal':
      // 打开 Pod 终端
      if (!selectedCluster.value) {
        message.warning('请先选择集群');
        return;
      }
      const containerName = row.containers?.[0]?.name || '';
      const baseUrl = window.location.origin;
      const terminalUrl = `${baseUrl}/k8s/terminal?clusterId=${selectedCluster.value}&namespace=${row.namespace}&podName=${row.name}&containerName=${containerName}`;
      const newWindow = window.open(terminalUrl, '_blank');
      if (!newWindow) {
        message.warning('浏览器阻止了新标签页打开，请检查浏览器设置允许弹窗');
      } else {
        newWindow.focus();
      }
      break;
    case 'delete':
      handlePodDelete(row);
      break;
  }
}

async function handlePodDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除 Pod "${row.name}" 吗？`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await deleteK8sPod(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('删除成功');
    await loadPods();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
}

// ========== StatefulSet 操作 ==========
async function handleStatefulSetRestart(row: any) {
  try {
    await ElMessageBox.confirm(`确定要重启 StatefulSet "row.name}" 吗？`, '确认重启', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await restartK8sStatefulSet(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('重启成功');
    await loadStatefulSets();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '重启失败');
    }
  }
}

async function handleStatefulSetDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除 StatefulSet "row.name}" 吗？此操作不可恢复！`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await deleteK8sStatefulSet(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('删除成功');
    await loadStatefulSets();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
}

// ========== DaemonSet 操作 ==========
async function handleDaemonSetRestart(row: any) {
  try {
    await ElMessageBox.confirm(`确定要重启 DaemonSet "row.name}" 吗？`, '确认重启', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await restartK8sDaemonSet(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('重启成功');
    await loadDaemonSets();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '重启失败');
    }
  }
}

async function handleDaemonSetDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除 DaemonSet "row.name}" 吗？此操作不可恢复！`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await deleteK8sDaemonSet(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('删除成功');
    await loadDaemonSets();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
}

// ========== Job 操作 ==========
async function handleJobDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除 Job "row.name}" 吗？`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await deleteK8sJob(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('删除成功');
    await loadJobs();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
}

// ========== CronJob 操作 ==========
async function handleCronJobDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除 CronJob "row.name}" 吗？此操作不可恢复！`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await deleteK8sCronJob(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name
    });
    message.success('删除成功');
    await loadCronJobs();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
}

async function handleCronJobSuspend(row: any) {
  const suspend = !row.suspend;
  const action = suspend ? '暂停' : '恢复';

  try {
    await ElMessageBox.confirm(`确定要${action} CronJob "${row.name}" 吗？`, `确认${action}`, {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    await suspendK8sCronJob(selectedCluster.value, {
      namespace: row.namespace,
      name: row.name,
      suspend
    });
    message.success(`${action}成功`);
    await loadCronJobs();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || `${action}失败`);
    }
  }
}

// ========== 通用操作处理 ==========
function handleWorkloadCommand(resourceType: string, command: string, row: any) {
  switch (resourceType) {
    case 'statefulset':
      if (command === 'edit') handleStatefulSetEdit(row);
      else if (command === 'restart') handleStatefulSetRestart(row);
      else if (command === 'delete') handleStatefulSetDelete(row);
      break;
    case 'daemonset':
      if (command === 'edit') handleDaemonSetEdit(row);
      else if (command === 'restart') handleDaemonSetRestart(row);
      else if (command === 'delete') handleDaemonSetDelete(row);
      break;
    case 'job':
      if (command === 'edit') handleJobEdit(row);
      else if (command === 'delete') handleJobDelete(row);
      break;
    case 'cronjob':
      if (command === 'edit') handleCronJobEdit(row);
      else if (command === 'delete') handleCronJobDelete(row);
      else if (command === 'suspend') handleCronJobSuspend(row);
      break;
  }
}

// ========== 批量操作 ==========

// Deployment 批量选择
function handleDeploymentSelectionChange(selection: any[]) {
  selectedDeployments.value = selection;
  batchOperationsVisible.value = selection.length > 0;
}

// Deployment 全选
function handleDeploymentSelectAll(selection: any[]) {
  selectedDeployments.value = selection;
}

// Pod 批量选择
function handlePodSelectionChange(selection: any[]) {
  selectedPods.value = selection;
}

// Pod 全选
function handlePodSelectAll(selection: any[]) {
  selectedPods.value = selection;
}

// 批量重启 Deployment
async function handleBatchRestart() {
  try {
    await ElMessageBox.confirm(
      `确定要重启选中的 ${selectedDeployments.value.length} 个 Deployment 吗？`,
      '确认批量重启',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    );

    if (!selectedCluster.value) return;

    loading.value = true;
    const promises = selectedDeployments.value.map(row =>
      restartK8sDeployment(selectedCluster.value, {
        namespace: row.namespace,
        name: row.name
      })
    );

    await Promise.all(promises);
    message.success(`成功重启 ${selectedDeployments.value.length} 个 Deployment`);
    selectedDeployments.value = [];
    batchOperationsVisible.value = false;
    await loadDeployments();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '批量重启失败');
    }
  } finally {
    loading.value = false;
  }
}

// 批量删除 Deployment
async function handleBatchDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedDeployments.value.length} 个 Deployment 吗？此操作不可恢复！`,
      '确认批量删除',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    );

    if (!selectedCluster.value) return;

    loading.value = true;
    const promises = selectedDeployments.value.map(row =>
      deleteK8sDeployment(selectedCluster.value, {
        namespace: row.namespace,
        name: row.name
      })
    );

    await Promise.all(promises);
    message.success(`成功删除 ${selectedDeployments.value.length} 个 Deployment`);
    selectedDeployments.value = [];
    batchOperationsVisible.value = false;
    await loadDeployments();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '批量删除失败');
    }
  } finally {
    loading.value = false;
  }
}

// 批量删除 Pod
async function handleBatchDeletePods() {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedPods.value.length} 个 Pod 吗？`, '确认批量删除', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (!selectedCluster.value) return;

    loading.value = true;
    const promises = selectedPods.value.map(row =>
      deleteK8sPod(selectedCluster.value, {
        namespace: row.namespace,
        name: row.name
      })
    );

    await Promise.all(promises);
    message.success(`成功删除 ${selectedPods.value.length} 个 Pod`);
    selectedPods.value = [];
    batchOperationsVisible.value = false;
    await loadPods();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '批量删除失败');
    }
  } finally {
    loading.value = false;
  }
}

// StatefulSet 批量选择
function handleStatefulSetSelectionChange(selection: any[]) {
  selectedStatefulSets.value = selection;
}

// StatefulSet 全选
function handleStatefulSetSelectAll(selection: any[]) {
  selectedStatefulSets.value = selection;
}

// DaemonSet 批量选择
function handleDaemonSetSelectionChange(selection: any[]) {
  selectedDaemonSets.value = selection;
}

// DaemonSet 全选
function handleDaemonSetSelectAll(selection: any[]) {
  selectedDaemonSets.value = selection;
}

// Job 批量选择
function handleJobSelectionChange(selection: any[]) {
  selectedJobs.value = selection;
}

// Job 全选
function handleJobSelectAll(selection: any[]) {
  selectedJobs.value = selection;
}

// CronJob 批量选择
function handleCronJobSelectionChange(selection: any[]) {
  selectedCronJobs.value = selection;
}

// CronJob 全选
function handleCronJobSelectAll(selection: any[]) {
  selectedCronJobs.value = selection;
}

// 通用批量删除处理器
async function handleBatchDeleteByType(resourceType: string) {
  let selectedItems: any[] = [];
  let deleteFunc: any;
  let loadFunc: any;
  let resourceName: string;

  switch (resourceType) {
    case 'statefulset':
      selectedItems = selectedStatefulSets.value;
      deleteFunc = deleteK8sStatefulSet;
      loadFunc = loadStatefulSets;
      resourceName = 'StatefulSet';
      break;
    case 'daemonset':
      selectedItems = selectedDaemonSets.value;
      deleteFunc = deleteK8sDaemonSet;
      loadFunc = loadDaemonSets;
      resourceName = 'DaemonSet';
      break;
    case 'job':
      selectedItems = selectedJobs.value;
      deleteFunc = deleteK8sJob;
      loadFunc = loadJobs;
      resourceName = 'Job';
      break;
    case 'cronjob':
      selectedItems = selectedCronJobs.value;
      deleteFunc = deleteK8sCronJob;
      loadFunc = loadCronJobs;
      resourceName = 'CronJob';
      break;
    default:
      return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedItems.length} 个 ${resourceName} 吗？此操作不可恢复！`,
      '确认批量删除',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    );

    if (!selectedCluster.value) return;

    loading.value = true;
    const promises = selectedItems.map(row =>
      deleteFunc(selectedCluster.value, {
        namespace: row.namespace,
        name: row.name
      })
    );

    await Promise.all(promises);
    message.success(`成功删除 ${selectedItems.length} 个 ${resourceName}`);

    // 清空选择
    switch (resourceType) {
      case 'statefulset':
        selectedStatefulSets.value = [];
        break;
      case 'daemonset':
        selectedDaemonSets.value = [];
        break;
      case 'job':
        selectedJobs.value = [];
        break;
      case 'cronjob':
        selectedCronJobs.value = [];
        break;
    }

    await loadFunc();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '批量删除失败');
    }
  } finally {
    loading.value = false;
  }
}

// 清空选择
function clearSelection() {
  selectedDeployments.value = [];
  selectedPods.value = [];
}

// 分页变化
function handlePageChange(page: number) {
  currentPagination.value.page = page;
  loadCurrentData();
}

function handlePageSizeChange(pageSize: number) {
  currentPagination.value.pageSize = pageSize;
  currentPagination.value.page = 1;
  loadCurrentData();
}

// Tab切换
function handleTabChange(tabName: string) {
  activeTab.value = tabName;
  // 切换Tab时清空选择并同步到 store
  clearSelection();
  k8sStore.setWorkloadFilterState({
    clusterId: selectedCluster.value,
    namespace: selectedNamespace.value,
    activeTab: tabName
  });
  loadCurrentData();
}

// 监听集群和命名空间变化，同步到 store
watch([selectedCluster, selectedNamespace], () => {
  // 同步状态到 store
  k8sStore.setWorkloadFilterState({
    clusterId: selectedCluster.value,
    namespace: selectedNamespace.value,
    activeTab: activeTab.value
  });

  if (selectedCluster.value) {
    loadNamespaces();
    loadCurrentData();
  }
});

onMounted(async () => {
  console.log('[工作负载汇总] ========== onMounted 开始 ==========');

  await loadClusters();

  // 从 store 恢复状态
  const savedState = k8sStore.getWorkloadFilterState();
  console.log('[工作负载汇总] 从 store 恢复状态:', savedState);

  if (savedState.clusterId) {
    selectedCluster.value = savedState.clusterId;
  } else if (clusters.value.length > 0) {
    selectedCluster.value = clusters.value[0].id;
  }

  if (savedState.namespace) {
    selectedNamespace.value = savedState.namespace;
  }

  if (savedState.activeTab) {
    activeTab.value = savedState.activeTab;
    console.log('[工作负载汇总] 恢复标签页:', activeTab.value);
  }

  console.log(
    '[工作负载汇总] 恢复后状态 - 集群:',
    selectedCluster.value,
    '命名空间:',
    selectedNamespace.value,
    '标签页:',
    activeTab.value
  );

  // 手动加载数据
  if (selectedCluster.value) {
    await loadNamespaces();

    // 如果命名空间列表中没有当前选择的命名空间，选择第一个
    if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
      selectedNamespace.value = namespaces.value[0];
    }

    await loadCurrentData();
  }

  console.log('[工作负载汇总] ========== onMounted 结束 ==========');
});
</script>

<template>
  <div class="workloads-page p-24px">
    <!-- 页面标题 -->
    <div class="mb-24px">
      <h1 class="text-28px text-primary font-bold">工作负载</h1>
      <p class="text-tertiary mt-8px text-14px">管理 Kubernetes 工作负载资源</p>
    </div>

    <!-- 头部：集群/命名空间选择和刷新按钮（同一行）-->
    <div class="mb-16px flex items-center justify-between gap-12px">
      <!-- 左侧：集群和命名空间选择 -->
      <div class="filter-inputs flex items-center gap-8px">
        <ElSelect v-model="selectedCluster" style="width: 200px" @change="loadNamespaces">
          <template #prefix>
            <span class="select-fixed-label">选择集群</span>
          </template>
          <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
        </ElSelect>
        <ElSelect v-model="selectedNamespace" style="width: 180px" @change="loadCurrentData">
          <template #prefix>
            <span class="select-fixed-label">选择命名空间</span>
          </template>
          <ElOption v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
        </ElSelect>
      </div>

      <!-- 右侧：刷新按钮 -->
      <ElButton text @click="loadCurrentData">
        <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
      </ElButton>
    </div>

    <!-- Tab 切换不同资源类型 -->
    <ElTabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 容器组 -->
      <ElTabPane label="容器组" name="pods">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="podsData"
            class="workloads-table"
            :header-cell-style="{
              background: '#f5f7fa',
              color: '#303133',
              fontWeight: '600',
              paddingLeft: '16px',
              paddingRight: '16px'
            }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
            @selection-change="handlePodSelectionChange"
            @select-all="handlePodSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="200" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn label="状态" min-width="100" align="left">
              <template #default="{ row }">
                <ElTag :type="row.status === 'Running' ? 'success' : 'warning'">
                  {{ row.status }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="ip" label="IP地址" min-width="140" align="left" />
            <ElTableColumn prop="node" label="节点" min-width="150" align="left" />
            <ElTableColumn prop="restarts" label="重启次数" min-width="100" align="left" />
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="150" fixed="right" align="left">
              <template #default="{ row }">
                <span class="operation-buttons">
                  <ElButton link type="primary" size="default" @click="goToDetail(row)">详情</ElButton>
                  <ElButton link type="primary" size="default" @click="handlePodLogs(row)">日志</ElButton>
                  <ElDropdown trigger="click" @command="cmd => handlePodMoreCommand(cmd, row)">
                    <span class="dropdown-link">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem command="edit">编辑YAML</ElDropdownItem>
                        <ElDropdownItem command="terminal">终端</ElDropdownItem>
                        <ElDropdownItem command="delete">删除</ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </span>
              </template>
            </ElTableColumn>
          </ElTable>

          <!-- 底部工具栏：批量操作 + 分页 -->
          <div class="bottom-toolbar">
            <div class="batch-operations-inline">
              <span class="batch-info">已选择 {{ selectedPods.length }} 项</span>
              <div class="batch-actions">
                <ElButton
                  size="default"
                  type="danger"
                  :disabled="selectedPods.length === 0"
                  @click="handleBatchDeletePods"
                >
                  批量删除
                </ElButton>
                <ElButton size="default" @click="clearSelection">取消选择</ElButton>
              </div>
            </div>
            <ElPagination
              v-model:current-page="podsPagination.page"
              v-model:page-size="podsPagination.pageSize"
              :total="podsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadPods"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 无状态 -->
      <ElTabPane label="无状态" name="deployments">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="deploymentsData"
            class="workloads-table"
            :header-cell-style="{
              background: '#f5f7fa',
              color: '#303133',
              fontWeight: '600',
              paddingLeft: '16px',
              paddingRight: '16px'
            }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
            @selection-change="handleDeploymentSelectionChange"
            @select-all="handleDeploymentSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="180" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn prop="replicas" label="副本数" min-width="100" align="left">
              <template #default="{ row }">{{ row.ready }} / {{ row.replicas }}</template>
            </ElTableColumn>
            <ElTableColumn prop="upToDate" label="最新" min-width="80" align="left" />
            <ElTableColumn prop="available" label="可用" min-width="80" align="left" />
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="状态" min-width="100" align="left">
              <template #default="{ row }">
                <ElTag v-if="row.ready === row.replicas" type="success">运行中</ElTag>
                <ElTag v-else type="warning">更新中</ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" min-width="200" fixed="right" align="left">
              <template #default="{ row }">
                <span class="operation-buttons">
                  <ElButton link type="primary" size="default" @click="goToDetail(row)">详情</ElButton>
                  <ElButton link type="primary" size="default" @click="handleScale(row)">伸缩</ElButton>
                  <ElDropdown trigger="click" @command="cmd => handleMoreCommand(cmd, row)">
                    <span class="dropdown-link">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem command="edit">编辑YAML</ElDropdownItem>
                        <ElDropdownItem command="restart">重启</ElDropdownItem>
                        <ElDropdownItem command="delete">删除</ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </span>
              </template>
            </ElTableColumn>
          </ElTable>

          <!-- 底部工具栏：批量操作 + 分页 -->
          <div class="bottom-toolbar">
            <div class="batch-operations-inline">
              <span class="batch-info">已选择 {{ selectedDeployments.length }} 项</span>
              <div class="batch-actions">
                <ElButton size="default" :disabled="selectedDeployments.length === 0" @click="handleBatchRestart">
                  批量重启
                </ElButton>
                <ElButton
                  size="default"
                  type="danger"
                  :disabled="selectedDeployments.length === 0"
                  @click="handleBatchDelete"
                >
                  批量删除
                </ElButton>
                <ElButton size="default" @click="clearSelection">取消选择</ElButton>
              </div>
            </div>
            <ElPagination
              v-model:current-page="deploymentsPagination.page"
              v-model:page-size="deploymentsPagination.pageSize"
              :total="deploymentsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadDeployments"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 有状态 -->
      <ElTabPane label="有状态" name="statefulsets">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="statefulSetsData"
            class="workloads-table"
            :header-cell-style="{
              background: '#f5f7fa',
              color: '#303133',
              fontWeight: '600',
              paddingLeft: '16px',
              paddingRight: '16px'
            }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
            @selection-change="handleStatefulSetSelectionChange"
            @select-all="handleStatefulSetSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="180" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn prop="replicas" label="副本数" min-width="100" align="left">
              <template #default="{ row }">{{ row.ready }} / {{ row.replicas }}</template>
            </ElTableColumn>
            <ElTableColumn prop="current" label="当前" min-width="80" align="left" />
            <ElTableColumn prop="updated" label="已更新" min-width="80" align="left" />
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="200" fixed="right" align="left">
              <template #default="{ row }">
                <span class="operation-buttons">
                  <ElButton link type="primary" size="default" @click="goToDetail(row)">详情</ElButton>
                  <ElDropdown trigger="click" @command="cmd => handleWorkloadCommand('statefulset', cmd, row)">
                    <span class="dropdown-link">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem command="edit">编辑YAML</ElDropdownItem>
                        <ElDropdownItem command="restart">重启</ElDropdownItem>
                        <ElDropdownItem command="delete">删除</ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </span>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <div class="batch-operations-inline">
              <span class="batch-info">已选择 {{ selectedStatefulSets.length }} 项</span>
              <ElButton
                size="small"
                type="danger"
                :disabled="selectedStatefulSets.length === 0"
                @click="handleBatchDeleteByType('statefulset')"
              >
                批量删除
              </ElButton>
            </div>
            <ElPagination
              v-model:current-page="statefulSetsPagination.page"
              v-model:page-size="statefulSetsPagination.pageSize"
              :total="statefulSetsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadStatefulSets"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 守护进程集 -->
      <ElTabPane label="守护进程集" name="daemonsets">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="daemonSetsData"
            class="workloads-table"
            :header-cell-style="{
              background: '#f5f7fa',
              color: '#303133',
              fontWeight: '600',
              paddingLeft: '16px',
              paddingRight: '16px'
            }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
            @selection-change="handleDaemonSetSelectionChange"
            @select-all="handleDaemonSetSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="180" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn label="节点数" min-width="120" align="left">
              <template #default="{ row }">{{ row.current }} / {{ row.desired }}</template>
            </ElTableColumn>
            <ElTableColumn prop="ready" label="就绪" min-width="80" align="left" />
            <ElTableColumn prop="available" label="可用" min-width="80" align="left" />
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="200" fixed="right" align="left">
              <template #default="{ row }">
                <span class="operation-buttons">
                  <ElButton link type="primary" size="default" @click="goToDetail(row)">详情</ElButton>
                  <ElDropdown trigger="click" @command="cmd => handleWorkloadCommand('daemonset', cmd, row)">
                    <span class="dropdown-link">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem command="edit">编辑YAML</ElDropdownItem>
                        <ElDropdownItem command="restart">重启</ElDropdownItem>
                        <ElDropdownItem command="delete">删除</ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </span>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <div class="batch-operations-inline">
              <span class="batch-info">已选择 {{ selectedDaemonSets.length }} 项</span>
              <ElButton
                size="small"
                type="danger"
                :disabled="selectedDaemonSets.length === 0"
                @click="handleBatchDeleteByType('daemonset')"
              >
                批量删除
              </ElButton>
            </div>
            <ElPagination
              v-model:current-page="daemonSetsPagination.page"
              v-model:page-size="daemonSetsPagination.pageSize"
              :total="daemonSetsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadDaemonSets"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 任务 -->
      <ElTabPane label="任务" name="jobs">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="jobsData"
            class="workloads-table"
            :header-cell-style="{
              background: '#f5f7fa',
              color: '#303133',
              fontWeight: '600',
              paddingLeft: '16px',
              paddingRight: '16px'
            }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
            @selection-change="handleJobSelectionChange"
            @select-all="handleJobSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="180" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn label="完成数" min-width="100" align="left">
              <template #default="{ row }">{{ row.succeeded || 0 }} / {{ row.completions || '-' }}</template>
            </ElTableColumn>
            <ElTableColumn prop="duration" label="时长" min-width="100" align="left" />
            <ElTableColumn label="状态" min-width="100" align="left">
              <template #default="{ row }">
                <ElTag
                  :type="
                    row.status === '完成'
                      ? 'success'
                      : row.status === '失败'
                        ? 'danger'
                        : row.status === '运行中'
                          ? 'primary'
                          : 'info'
                  "
                >
                  {{ row.status }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="180" fixed="right" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" size="default" @click="goToDetail(row)">详情</ElButton>
                <ElDropdown trigger="click" @command="cmd => handleWorkloadCommand('job', cmd, row)">
                  <span class="dropdown-link">
                    更多
                    <icon-mdi-chevron-down class="dropdown-icon" />
                  </span>
                  <template #dropdown>
                    <ElDropdownMenu>
                      <ElDropdownItem command="edit">编辑YAML</ElDropdownItem>
                      <ElDropdownItem command="delete">删除</ElDropdownItem>
                    </ElDropdownMenu>
                  </template>
                </ElDropdown>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <div class="batch-operations-inline">
              <span class="batch-info">已选择 {{ selectedJobs.length }} 项</span>
              <ElButton
                size="small"
                type="danger"
                :disabled="selectedJobs.length === 0"
                @click="handleBatchDeleteByType('job')"
              >
                批量删除
              </ElButton>
            </div>
            <ElPagination
              v-model:current-page="jobsPagination.page"
              v-model:page-size="jobsPagination.pageSize"
              :total="jobsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadJobs"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 定时任务 -->
      <ElTabPane label="定时任务" name="cronjobs">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="cronJobsData"
            class="workloads-table"
            :header-cell-style="{
              background: '#f5f7fa',
              color: '#303133',
              fontWeight: '600',
              paddingLeft: '16px',
              paddingRight: '16px'
            }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 16px' }"
            @selection-change="handleCronJobSelectionChange"
            @select-all="handleCronJobSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="180" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn prop="schedule" label="调度规则" min-width="160" align="left" />
            <ElTableColumn label="挂起" min-width="80" align="left">
              <template #default="{ row }">
                <ElTag :type="row.suspend ? 'warning' : 'success'">
                  {{ row.suspend ? '是' : '否' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="上次执行" min-width="160" align="left">
              <template #default="{ row }">
                {{ row.lastSchedule || '-' }}
              </template>
            </ElTableColumn>
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="200" fixed="right" align="left">
              <template #default="{ row }">
                <span class="operation-buttons">
                  <ElButton link type="primary" size="default" @click="goToDetail(row)">详情</ElButton>
                  <ElDropdown trigger="click" @command="cmd => handleWorkloadCommand('cronjob', cmd, row)">
                    <span class="dropdown-link">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem command="edit">编辑YAML</ElDropdownItem>
                        <ElDropdownItem command="suspend">{{ row.suspend ? '恢复' : '暂停' }}</ElDropdownItem>
                        <ElDropdownItem command="delete">删除</ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </span>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <div class="batch-operations-inline">
              <span class="batch-info">已选择 {{ selectedCronJobs.length }} 项</span>
              <ElButton
                size="small"
                type="danger"
                :disabled="selectedCronJobs.length === 0"
                @click="handleBatchDeleteByType('cronjob')"
              >
                批量删除
              </ElButton>
            </div>
            <ElPagination
              v-model:current-page="cronJobsPagination.page"
              v-model:page-size="cronJobsPagination.pageSize"
              :total="cronJobsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadCronJobs"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- 缩放弹窗 -->
    <ElDialog v-model="showScaleDialog" title="缩放 Deployment" width="500px">
      <ElForm ref="scaleFormRef" :model="scaleData" label-width="100px">
        <ElFormItem label="Deployment">
          <ElInput :value="selectedDeployment?.name" disabled />
        </ElFormItem>
        <ElFormItem label="副本数" prop="replicas">
          <ElInputNumber v-model="scaleData.replicas" :min="0" :max="100" :step="1" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showScaleDialog = false">取消</ElButton>
        <ElButton type="primary" @click="handleScaleSubmit">确定</ElButton>
      </template>
    </ElDialog>

    <!-- 日志对话框 -->
    <ElDialog
      v-model="showLogDialog"
      :title="`日志: ${selectedResource?.name}${selectedResource?.containers?.[0]?.name ? ' (' + selectedResource.containers[0].name + ')' : ''}`"
      width="900px"
      top="5vh"
    >
      <div class="log-content">{{ logDialogContent }}</div>
      <template #footer>
        <ElButton @click="showLogDialog = false">关闭</ElButton>
      </template>
    </ElDialog>

    <!-- YAML 编辑器 -->
    <YamlEditor
      v-model="showYamlDialog"
      :title="yamlDialogTitle"
      :yaml="yamlContent"
      :can-edit="true"
      :on-apply="handleYamlApply"
    />
  </div>
</template>

<style scoped>
.workloads-page {
  height: calc(100vh - 80px); /* 减去顶部导航和标题区域的高度 */
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Tab 容器自动填充剩余空间 */
:deep(.el-tabs) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Tab 内容区域自动填充 */
:deep(.el-tabs__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 每个 Tab 页面填充空间 */
:deep(.el-tab-pane) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 100%;
}

/* 筛选输入框样式 - 参考主机资产页面 */
.filter-inputs {
  :deep(.el-select__wrapper) {
    border-radius: 0 !important;
    height: 30px;
    font-size: 12px;
    line-height: 30px;
  }

  :deep(.el-input__wrapper) {
    border-radius: 0 !important;
    height: 30px;
    font-size: 12px;
  }

  :deep(.el-select) {
    height: 30px;
    font-size: 12px;
  }

  /* 隐藏选中项的显示，因为我们要显示固定的标签文案 */
  :deep(.el-select .el-select__selection) {
    display: none;
  }

  :deep(.el-select .el-select__selected-item) {
    display: none;
  }

  /* placeholder 样式 - 隐藏，因为我们用 prefix 替代 */
  :deep(.el-select .el-select__placeholder) {
    display: none;
  }

  /* 固定标签文案样式 */
  .select-fixed-label {
    font-size: 12px;
    color: var(--el-text-color-regular);
    line-height: 30px;
    padding-left: 8px;
  }

  /* 当有选中值时，调整 prefix 的位置 */
  :deep(.el-select.has-value .el-select__prefix) {
    position: static;
    flex: none;
  }
}

/* 底部工具栏：批量操作 + 分页 - 固定在容器底部 */
.bottom-toolbar {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background-color: #fff;
  border-top: 1px solid #ebeef5;
  gap: 16px;
  z-index: 10;

  /* 批量操作区域 */
  .batch-operations-inline {
    display: flex;
    align-items: center;
    gap: 16px;

    .batch-info {
      font-size: 14px;
      color: #303133;
      font-weight: 500;
      white-space: nowrap;
    }

    .batch-actions {
      display: flex;
      gap: 8px;

      /* 禁用状态的按钮样式 */
      :deep(.el-button.is-disabled) {
        opacity: 0.5;
      }
    }
  }

  /* 分页区域 */
  :deep(.el-pagination) {
    margin: 0;
  }
}

/* 批量操作按钮栏（旧版，保留备用） */
.batch-operations {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background-color: #e6f7ff;
  border: 1px solid #91d5ff;
  border-radius: 4px;
  margin-bottom: 16px;

  .batch-info {
    font-size: 14px;
    color: #303133;
    font-weight: 500;
  }

  .batch-actions {
    display: flex;
    gap: 8px;
  }
}

/* 表格容器样式 */
.table-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  min-height: 0; /* 让flex自动计算高度，不设置固定最小值 */
  background-color: transparent !important;
}

/* 表格样式优化 */
.workloads-table {
  flex: 1;
  min-height: 0; /* 确保表格能够正确伸缩 */
  overflow: auto;
  padding-bottom: 60px; /* 为底部工具栏留出空间 */

  /* 强制移除表格所有背景色 - 使用更多选择器 */
  &,
  &.el-table,
  :deep(.el-table),
  :deep(.el-table__body),
  :deep(.el-table__body-wrapper),
  :deep(.el-table__inner-wrapper),
  :deep(.el-table__header) {
    background-color: transparent !important;
  }

  /* 表头样式 - 使用稍深颜色区分 */
  :deep(.el-table__header-wrapper) {
    background-color: transparent !important;

    th.el-table__cell {
      background-color: #f5f7fa !important;
      color: #303133;
      font-weight: 600;
      text-align: left;
      position: sticky;
      top: 0;
      z-index: 1;
    }
  }

  /* 移除所有行的背景色 - 数据行透明 */
  :deep(.el-table__body-wrapper) {
    background-color: transparent !important;
  }

  :deep(.el-table__body) {
    background-color: transparent !important;
  }

  :deep(.el-table__body tr) {
    background-color: transparent !important;
  }

  :deep(.el-table__body td.el-table__cell) {
    background-color: transparent !important;
  }

  /* 去掉斑马纹 */
  &.el-table--striped {
    :deep(.el-table__body tr.el-table__row--striped) {
      background-color: transparent !important;

      td.el-table__cell {
        background-color: transparent !important;
      }
    }
  }

  /* 移除 hover 效果的背景色 */
  :deep(.el-table__body tr:hover > td.el-table__cell) {
    background-color: transparent !important;
  }

  /* 移除固定列的背景色 */
  :deep(.el-table__fixed),
  :deep(.el-table__fixed-body-wrapper) {
    background-color: transparent !important;

    .el-table__body tr {
      background-color: transparent !important;
    }

    .el-table__body td.el-table__cell {
      background-color: transparent !important;
    }
  }

  /* 移除表格容器的背景色 */
  :deep(.el-table__inner-wrapper) {
    background-color: transparent !important;
  }

  /* 操作按钮容器 */
  .operation-buttons {
    display: flex;
    align-items: center;
    gap: 8px;

    .el-button {
      margin: 0;
      padding: 0;
      font-size: var(--el-font-size-base) !important;
      font-weight: 400 !important;
    }

    /* 下拉链接样式 */
    .dropdown-link {
      display: inline-flex;
      align-items: center;
      gap: 2px;
      cursor: pointer;
      color: var(--el-color-primary);
      font-size: var(--el-font-size-base) !important;
      font-weight: 400;
      user-select: none;

      .dropdown-icon {
        font-size: 14px;
      }

      &:hover {
        opacity: 0.8;
      }
    }
  }
}

/* 日志内容样式 */
.log-content {
  background: #1e1e1e;
  color: #4ec9b0;
  padding: 16px;
  border-radius: 4px;
  font-family: "Courier New", Courier, monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 600px;
  overflow: auto;
}
</style>
