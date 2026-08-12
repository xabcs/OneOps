import { reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import yaml from 'js-yaml';
import {
  deleteK8sCronJob,
  deleteK8sDaemonSet,
  deleteK8sDeployment,
  deleteK8sJob,
  deleteK8sPod,
  deleteK8sStatefulSet,
  fetchK8sPodLogs,
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

function parseManifest(manifestStr: string): string {
  if (!manifestStr) return '';
  try {
    const obj = JSON.parse(manifestStr);
    delete obj.managedFields;
    return yaml.dump(obj, { indent: 2, lineWidth: 120, noRefs: true, sortKeys: false });
  } catch (e) {
    console.error('解析manifest失败:', e);
    return manifestStr;
  }
}

// 资源操作配置映射
const resourceConfig: Record<
  string,
  {
    label: string;
    getFn: (clusterId: string, ns: string, name: string) => Promise<unknown>;
    deleteFn: (clusterId: string, params: K8s.ResourceDeleteParams) => Promise<unknown>;
    restartFn?: (clusterId: string, params: K8s.ResourceDeleteParams) => Promise<unknown>;
    updateFn: ((clusterId: string, params: K8s.ResourceUpdateParams) => Promise<unknown>) | null;
  }
> = {
  deployments: {
    label: 'Deployment',
    getFn: getK8sDeployment,
    deleteFn: deleteK8sDeployment,
    restartFn: restartK8sDeployment,
    updateFn: updateK8sDeployment
  },
  statefulsets: {
    label: 'StatefulSet',
    getFn: getK8sStatefulSet,
    deleteFn: deleteK8sStatefulSet,
    restartFn: restartK8sStatefulSet,
    updateFn: updateK8sStatefulSet
  },
  daemonsets: {
    label: 'DaemonSet',
    getFn: getK8sDaemonSet,
    deleteFn: deleteK8sDaemonSet,
    restartFn: restartK8sDaemonSet,
    updateFn: updateK8sDaemonSet
  },
  jobs: { label: 'Job', getFn: getK8sJob, deleteFn: deleteK8sJob, updateFn: updateK8sJob },
  cronjobs: { label: 'CronJob', getFn: getK8sCronJob, deleteFn: deleteK8sCronJob, updateFn: updateK8sCronJob },
  pods: { label: 'YAML', getFn: getK8sPod, deleteFn: deleteK8sPod, updateFn: null }
};

export function useWorkloadActions(getSelectedCluster: () => string | undefined, getActiveTab: () => string) {
  const message = ElMessage;

  const showScaleDialog = ref(false);
  const scaleData = reactive({ replicas: 1 });
  const selectedDeployment = ref<K8s.Deployment | null>(null);

  const showYamlDialog = ref(false);
  const yamlDialogTitle = ref('');
  const yamlContent = ref('');
  const selectedResource = ref<K8s.WorkloadRow | null>(null);
  const yamlLoading = ref(false);

  const showLogDialog = ref(false);
  const logDialogContent = ref('');

  const selectedDeployments = ref<K8s.Deployment[]>([]);
  const selectedPods = ref<K8s.Pod[]>([]);
  const selectedStatefulSets = ref<K8s.StatefulSet[]>([]);
  const selectedDaemonSets = ref<K8s.DaemonSet[]>([]);
  const selectedJobs = ref<K8s.Job[]>([]);
  const selectedCronJobs = ref<K8s.CronJob[]>([]);
  const batchOperationsVisible = ref(false);

  // 加载函数引用
  const loadFns: Record<string, () => Promise<void>> = {};

  function setLoadFunctions(fns: Record<string, () => Promise<void>>) {
    Object.assign(loadFns, fns);
  }

  // ========== 通用编辑操作 ==========
  async function handleResourceEdit(resourceType: string, row: K8s.WorkloadRow) {
    if (!getSelectedCluster()) {
      message.warning('请先选择集群');
      return;
    }
    const config = resourceConfig[resourceType];
    if (!config) return;
    selectedResource.value = row;
    yamlDialogTitle.value = `编辑 ${config.label}: ${row.name}`;
    yamlLoading.value = true;
    try {
      const response = (await config.getFn(getSelectedCluster()!, row.namespace, row.name)) as {
        data?: { manifest?: string };
        manifest?: string;
      };
      yamlContent.value = parseManifest(response.data?.manifest || response.manifest || '');
      showYamlDialog.value = true;
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '获取 YAML 失败');
    } finally {
      yamlLoading.value = false;
    }
  }

  // ========== YAML 应用 ==========
  async function handleYamlApply(yamlStr: string) {
    if (!getSelectedCluster() || !selectedResource.value) throw new Error('缺少必要参数');
    const resourceType = getActiveTab();
    const config = resourceConfig[resourceType];
    if (!config?.updateFn) throw new Error('不支持的资源类型');
    try {
      const manifestObj = yaml.load(yamlStr);
      await config.updateFn(getSelectedCluster()!, {
        namespace: selectedResource.value.namespace,
        manifest: manifestObj
      });
      await loadFns[resourceType]?.();
    } catch (error: unknown) {
      const err = error as Error;
      throw new Error(err.message || 'YAML 应用失败');
    }
  }

  // ========== 通用删除/重启操作 ==========
  async function handleResourceDelete(resourceType: string, row: K8s.WorkloadRow) {
    const config = resourceConfig[resourceType];
    if (!config) return;
    try {
      await ElMessageBox.confirm(`确定要删除 ${config.label} "${row.name}" 吗？此操作不可恢复！`, '确认删除', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      });
      if (!getSelectedCluster()) return;
      await config.deleteFn(getSelectedCluster()!, { namespace: row.namespace, name: row.name });
      message.success('删除成功');
      await loadFns[resourceType]?.();
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || '删除失败');
    }
  }

  async function handleResourceRestart(resourceType: string, row: K8s.WorkloadRow) {
    const config = resourceConfig[resourceType];
    if (!config?.restartFn) return;
    try {
      await ElMessageBox.confirm(`确定要重启 ${config.label} "${row.name}" 吗？`, '确认重启', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      });
      if (!getSelectedCluster()) return;
      await config.restartFn(getSelectedCluster()!, { namespace: row.namespace, name: row.name });
      message.success('重启成功');
      await loadFns[resourceType]?.();
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || '重启失败');
    }
  }

  // ========== Deployment 专属操作 ==========
  function handleScale(row: K8s.Deployment) {
    selectedDeployment.value = row;
    scaleData.replicas = row.replicas || 1;
    showScaleDialog.value = true;
  }

  async function handleScaleSubmit() {
    if (!selectedDeployment.value || !getSelectedCluster()) return;
    try {
      await scaleK8sDeployment(getSelectedCluster()!, {
        namespace: selectedDeployment.value.namespace,
        name: selectedDeployment.value.name,
        replicas: scaleData.replicas
      });
      message.success('缩放成功');
      showScaleDialog.value = false;
      await loadFns.deployments?.();
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '缩放失败');
    }
  }

  function handleMoreCommand(command: string, row: K8s.Deployment) {
    if (command === 'edit') handleResourceEdit('deployments', row);
    else if (command === 'restart') handleResourceRestart('deployments', row);
    else if (command === 'delete') handleResourceDelete('deployments', row);
  }

  // 保持向后兼容的独立函数（供外部直接调用）
  const handleRestart = (row: K8s.Deployment) => handleResourceRestart('deployments', row);
  const handleDelete = (row: K8s.Deployment) => handleResourceDelete('deployments', row);
  const handleStatefulSetRestart = (row: K8s.StatefulSet) => handleResourceRestart('statefulsets', row);
  const handleStatefulSetDelete = (row: K8s.StatefulSet) => handleResourceDelete('statefulsets', row);
  const handleDaemonSetRestart = (row: K8s.DaemonSet) => handleResourceRestart('daemonsets', row);
  const handleDaemonSetDelete = (row: K8s.DaemonSet) => handleResourceDelete('daemonsets', row);
  const handleJobDelete = (row: K8s.Job) => handleResourceDelete('jobs', row);
  const handleCronJobDelete = (row: K8s.CronJob) => handleResourceDelete('cronjobs', row);

  // ========== Pod 操作 ==========
  async function handlePodLogs(row: K8s.Pod) {
    if (!getSelectedCluster()) {
      message.warning('请先选择集群');
      return;
    }
    const containerName = row.containers?.[0]?.name || '';
    selectedResource.value = row;
    logDialogContent.value = '加载中...';
    showLogDialog.value = true;
    try {
      const response = (await fetchK8sPodLogs(getSelectedCluster()!, row.namespace, row.name, {
        container: containerName,
        tailLines: 100
      })) as { data?: { logs?: string } };
      logDialogContent.value = response.data?.logs || '暂无日志';
    } catch (error: unknown) {
      const err = error as Error;
      logDialogContent.value = `日志加载失败: ${err.message || '未知错误'}`;
      message.error(`日志加载失败: ${err.message}`);
    }
  }

  function handlePodMoreCommand(command: string, row: K8s.Pod) {
    if (command === 'edit') handleResourceEdit('pods', row);
    else if (command === 'terminal') {
      if (!getSelectedCluster()) {
        message.warning('请先选择集群');
        return;
      }
      const containerName = row.containers?.[0]?.name || '';
      const terminalUrl = `${window.location.origin}/k8s/terminal?clusterId=${getSelectedCluster()}&namespace=${row.namespace}&podName=${row.name}&containerName=${containerName}`;
      const newWindow = window.open(terminalUrl, '_blank');
      if (!newWindow) message.warning('浏览器阻止了新标签页打开，请检查浏览器设置允许弹窗');
      else newWindow.focus();
    } else if (command === 'delete') handleResourceDelete('pods', row);
  }

  const handlePodDelete = (row: K8s.Pod) => handleResourceDelete('pods', row);

  // ========== CronJob 暂停/恢复 ==========
  async function handleCronJobSuspend(row: K8s.CronJob) {
    const suspend = !row.suspend;
    const action = suspend ? '暂停' : '恢复';
    try {
      await ElMessageBox.confirm(`确定要${action} CronJob "${row.name}" 吗？`, `确认${action}`, {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      });
      if (!getSelectedCluster()) return;
      await suspendK8sCronJob(getSelectedCluster()!, { namespace: row.namespace, name: row.name, suspend });
      message.success(`${action}成功`);
      await loadFns.cronjobs?.();
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || `${action}失败`);
    }
  }

  // ========== 通用操作处理（Tab Dropdown） ==========
  function handleWorkloadCommand(resourceType: string, command: string, row: K8s.WorkloadRow) {
    const typeMap: Record<string, string> = {
      statefulset: 'statefulsets',
      daemonset: 'daemonsets',
      job: 'jobs',
      cronjob: 'cronjobs'
    };
    const tabType = typeMap[resourceType];
    if (command === 'edit') handleResourceEdit(tabType, row);
    else if (command === 'restart') handleResourceRestart(tabType, row);
    else if (command === 'delete') handleResourceDelete(tabType, row);
    else if (command === 'suspend') handleCronJobSuspend(row as K8s.CronJob);
  }

  // ========== 批量操作 ==========
  function handleDeploymentSelectionChange(selection: K8s.Deployment[]) {
    selectedDeployments.value = selection;
    batchOperationsVisible.value = selection.length > 0;
  }
  function handleDeploymentSelectAll(selection: K8s.Deployment[]) {
    selectedDeployments.value = selection;
  }
  function handlePodSelectionChange(selection: K8s.Pod[]) {
    selectedPods.value = selection;
  }
  function handlePodSelectAll(selection: K8s.Pod[]) {
    selectedPods.value = selection;
  }
  function handleStatefulSetSelectionChange(selection: K8s.StatefulSet[]) {
    selectedStatefulSets.value = selection;
  }
  function handleStatefulSetSelectAll(selection: K8s.StatefulSet[]) {
    selectedStatefulSets.value = selection;
  }
  function handleDaemonSetSelectionChange(selection: K8s.DaemonSet[]) {
    selectedDaemonSets.value = selection;
  }
  function handleDaemonSetSelectAll(selection: K8s.DaemonSet[]) {
    selectedDaemonSets.value = selection;
  }
  function handleJobSelectionChange(selection: K8s.Job[]) {
    selectedJobs.value = selection;
  }
  function handleJobSelectAll(selection: K8s.Job[]) {
    selectedJobs.value = selection;
  }
  function handleCronJobSelectionChange(selection: K8s.CronJob[]) {
    selectedCronJobs.value = selection;
  }
  function handleCronJobSelectAll(selection: K8s.CronJob[]) {
    selectedCronJobs.value = selection;
  }

  // 通用批量确认+执行
  async function executeBatchAction<T extends K8s.WorkloadRow>(
    items: T[],
    actionFn: (row: T) => Promise<unknown>,
    successMsg: string,
    afterAction: () => void
  ) {
    try {
      if (!getSelectedCluster()) return;
      await Promise.all(items.map(actionFn));
      message.success(successMsg);
      afterAction();
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || '批量操作失败');
    }
  }

  async function handleBatchRestart() {
    try {
      await ElMessageBox.confirm(
        `确定要重启选中的 ${selectedDeployments.value.length} 个 Deployment 吗？`,
        '确认批量重启',
        { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
      );
      await executeBatchAction(
        selectedDeployments.value,
        row => restartK8sDeployment(getSelectedCluster()!, { namespace: row.namespace, name: row.name }),
        `成功重启 ${selectedDeployments.value.length} 个 Deployment`,
        async () => {
          selectedDeployments.value = [];
          batchOperationsVisible.value = false;
          await loadFns.deployments?.();
        }
      );
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || '批量重启失败');
    }
  }

  async function handleBatchDelete() {
    try {
      await ElMessageBox.confirm(
        `确定要删除选中的 ${selectedDeployments.value.length} 个 Deployment 吗？此操作不可恢复！`,
        '确认批量删除',
        { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
      );
      await executeBatchAction(
        selectedDeployments.value,
        row => deleteK8sDeployment(getSelectedCluster()!, { namespace: row.namespace, name: row.name }),
        `成功删除 ${selectedDeployments.value.length} 个 Deployment`,
        async () => {
          selectedDeployments.value = [];
          batchOperationsVisible.value = false;
          await loadFns.deployments?.();
        }
      );
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || '批量删除失败');
    }
  }

  async function handleBatchDeletePods() {
    try {
      await ElMessageBox.confirm(`确定要删除选中的 ${selectedPods.value.length} 个 Pod 吗？`, '确认批量删除', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      });
      await executeBatchAction(
        selectedPods.value,
        row => deleteK8sPod(getSelectedCluster()!, { namespace: row.namespace, name: row.name }),
        `成功删除 ${selectedPods.value.length} 个 Pod`,
        async () => {
          selectedPods.value = [];
          batchOperationsVisible.value = false;
          await loadFns.pods?.();
        }
      );
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || '批量删除失败');
    }
  }

  async function handleBatchDeleteByType(resourceType: string) {
    const batchConfig: Record<
      string,
      {
        items: K8s.WorkloadRow[];
        deleteFn: (clusterId: string, params: K8s.ResourceDeleteParams) => Promise<unknown>;
        label: string;
      }
    > = {
      statefulset: { items: selectedStatefulSets.value, deleteFn: deleteK8sStatefulSet, label: 'StatefulSet' },
      daemonset: { items: selectedDaemonSets.value, deleteFn: deleteK8sDaemonSet, label: 'DaemonSet' },
      job: { items: selectedJobs.value, deleteFn: deleteK8sJob, label: 'Job' },
      cronjob: { items: selectedCronJobs.value, deleteFn: deleteK8sCronJob, label: 'CronJob' }
    };
    const cfg = batchConfig[resourceType];
    if (!cfg) return;

    try {
      await ElMessageBox.confirm(
        `确定要删除选中的 ${cfg.items.length} 个 ${cfg.label} 吗？此操作不可恢复！`,
        '确认批量删除',
        { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
      );
      await executeBatchAction(
        cfg.items,
        row => cfg.deleteFn(getSelectedCluster()!, { namespace: row.namespace, name: row.name }),
        `成功删除 ${cfg.items.length} 个 ${cfg.label}`,
        async () => {
          const clearMap: Record<string, { value: K8s.WorkloadRow[] }> = {
            statefulset: selectedStatefulSets,
            daemonset: selectedDaemonSets,
            job: selectedJobs,
            cronjob: selectedCronJobs
          };
          clearMap[resourceType].value = [];
          await loadFns[
            resourceType === 'statefulset'
              ? 'statefulsets'
              : resourceType === 'daemonset'
                ? 'daemonsets'
                : resourceType === 'job'
                  ? 'jobs'
                  : 'cronjobs'
          ]?.();
        }
      );
    } catch (error: unknown) {
      const err = error as Error;
      if (error !== 'cancel') message.error(err.message || '批量删除失败');
    }
  }

  function clearSelection() {
    selectedDeployments.value = [];
    selectedPods.value = [];
  }

  return {
    showScaleDialog,
    scaleData,
    selectedDeployment,
    showYamlDialog,
    yamlDialogTitle,
    yamlContent,
    selectedResource,
    yamlLoading,
    showLogDialog,
    logDialogContent,
    selectedDeployments,
    selectedPods,
    selectedStatefulSets,
    selectedDaemonSets,
    selectedJobs,
    selectedCronJobs,
    batchOperationsVisible,
    setLoadFunctions,
    handleResourceEdit,
    handleYamlApply,
    handleScale,
    handleScaleSubmit,
    handleMoreCommand,
    handleRestart,
    handleDelete,
    handlePodLogs,
    handlePodMoreCommand,
    handlePodDelete,
    handleStatefulSetRestart,
    handleStatefulSetDelete,
    handleDaemonSetRestart,
    handleDaemonSetDelete,
    handleJobDelete,
    handleCronJobDelete,
    handleCronJobSuspend,
    handleWorkloadCommand,
    handleDeploymentSelectionChange,
    handleDeploymentSelectAll,
    handlePodSelectionChange,
    handlePodSelectAll,
    handleBatchRestart,
    handleBatchDelete,
    handleBatchDeletePods,
    handleStatefulSetSelectionChange,
    handleStatefulSetSelectAll,
    handleDaemonSetSelectionChange,
    handleDaemonSetSelectAll,
    handleJobSelectionChange,
    handleJobSelectAll,
    handleCronJobSelectionChange,
    handleCronJobSelectAll,
    handleBatchDeleteByType,
    clearSelection
  };
}
