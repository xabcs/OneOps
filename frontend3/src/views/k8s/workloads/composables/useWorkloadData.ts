import { computed, onMounted, reactive, ref, watch } from 'vue';
import {
  fetchK8sClusterNamespaces,
  fetchK8sClusters,
  fetchK8sCronJobs,
  fetchK8sDaemonSets,
  fetchK8sDeployments,
  fetchK8sJobs,
  fetchK8sPods,
  fetchK8sStatefulSets
} from '@/service/api/k8s';
import { useK8sStore } from '@/store/modules/k8s';

export function useWorkloadData() {
  const k8sStore = useK8sStore();

  const loading = ref(false);
  const activeTab = ref(k8sStore.workloadFilterState.activeTab || 'deployments');

  const selectedCluster = ref(k8sStore.workloadFilterState.clusterId);
  const selectedNamespace = ref(k8sStore.workloadFilterState.namespace);

  const namespaces = ref<string[]>([]);
  const clusters = ref<K8s.Cluster[]>([]);

  // 各种资源的数据
  const deploymentsData = ref<K8s.Deployment[]>([]);
  const podsData = ref<K8s.Pod[]>([]);
  const statefulSetsData = ref<K8s.StatefulSet[]>([]);
  const daemonSetsData = ref<K8s.DaemonSet[]>([]);
  const jobsData = ref<K8s.Job[]>([]);
  const cronJobsData = ref<K8s.CronJob[]>([]);

  // 各种资源的分页
  const deploymentsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
  const podsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
  const statefulSetsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
  const daemonSetsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
  const jobsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
  const cronJobsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

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
      namespaces.value = response.data.map((ns: { name: string }) => ns.name);
      if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
        selectedNamespace.value = namespaces.value[0];
      }
    } catch (error) {
      console.error('加载命名空间失败:', error);
    }
  }

  // 通用 API 响应解析
  function parseApiResponse<T>(response: unknown): { list: T[]; total: number } {
    const resp = response as { data?: { data?: { list: T[]; total: number }; list: T[]; total: number }; list: T[]; total: number };
    if (resp?.data?.data?.list) return resp.data.data;
    if (resp?.data?.list) return resp.data;
    return resp;
  }

  // 加载各种资源数据
  async function loadDeployments() {
    if (!selectedCluster.value) return;
    loading.value = true;
    try {
      const response = await fetchK8sDeployments(selectedCluster.value, {
        namespace: selectedNamespace.value,
        page: deploymentsPagination.page,
        pageSize: deploymentsPagination.pageSize
      });
      const apiData = parseApiResponse<K8s.Deployment>(response);
      deploymentsData.value = apiData.list || [];
      deploymentsPagination.itemCount = apiData.total || 0;
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
      const response = await fetchK8sPods(selectedCluster.value, {
        namespace: selectedNamespace.value,
        page: podsPagination.page,
        pageSize: podsPagination.pageSize
      });
      const apiData = parseApiResponse<K8s.Pod>(response);
      podsData.value = apiData.list || [];
      podsPagination.itemCount = apiData.total || 0;
    } catch (error) {
      console.error('加载 Pods 失败:', error);
    } finally {
      loading.value = false;
    }
  }

  async function loadStatefulSets() {
    if (!selectedCluster.value) return;
    loading.value = true;
    try {
      const response = await fetchK8sStatefulSets(selectedCluster.value, {
        namespace: selectedNamespace.value,
        page: statefulSetsPagination.page,
        pageSize: statefulSetsPagination.pageSize
      });
      const apiData = parseApiResponse<K8s.StatefulSet>(response);
      statefulSetsData.value = apiData.list || [];
      statefulSetsPagination.itemCount = apiData.total || 0;
    } catch (error) {
      console.error('加载 StatefulSets 失败:', error);
    } finally {
      loading.value = false;
    }
  }

  async function loadDaemonSets() {
    if (!selectedCluster.value) return;
    loading.value = true;
    try {
      const response = await fetchK8sDaemonSets(selectedCluster.value, {
        namespace: selectedNamespace.value,
        page: daemonSetsPagination.page,
        pageSize: daemonSetsPagination.pageSize
      });
      const apiData = parseApiResponse<K8s.DaemonSet>(response);
      daemonSetsData.value = apiData.list || [];
      daemonSetsPagination.itemCount = apiData.total || 0;
    } catch (error) {
      console.error('加载 DaemonSets 失败:', error);
    } finally {
      loading.value = false;
    }
  }

  async function loadJobs() {
    if (!selectedCluster.value) return;
    loading.value = true;
    try {
      const response = await fetchK8sJobs(selectedCluster.value, {
        namespace: selectedNamespace.value,
        page: jobsPagination.page,
        pageSize: jobsPagination.pageSize
      });
      const apiData = parseApiResponse(response);
      jobsData.value = apiData.list || [];
      jobsPagination.itemCount = apiData.total || 0;
    } catch (error) {
      console.error('加载 Jobs 失败:', error);
    } finally {
      loading.value = false;
    }
  }

  async function loadCronJobs() {
    if (!selectedCluster.value) return;
    loading.value = true;
    try {
      const response = await fetchK8sCronJobs(selectedCluster.value, {
        namespace: selectedNamespace.value,
        page: cronJobsPagination.page,
        pageSize: cronJobsPagination.pageSize
      });
      const apiData = parseApiResponse<K8s.CronJob>(response);
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
    k8sStore.setWorkloadFilterState({
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      activeTab: tabName
    });
    loadCurrentData();
  }

  // 监听集群和命名空间变化，同步到 store
  watch([selectedCluster, selectedNamespace], () => {
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

  // 初始化
  async function init() {
    await loadClusters();

    const savedState = k8sStore.getWorkloadFilterState();

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
    }

    if (selectedCluster.value) {
      await loadNamespaces();

      if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
        selectedNamespace.value = namespaces.value[0];
      }

      await loadCurrentData();
    }
  }

  onMounted(init);

  return {
    loading,
    activeTab,
    selectedCluster,
    selectedNamespace,
    namespaces,
    clusters,
    deploymentsData,
    podsData,
    statefulSetsData,
    daemonSetsData,
    jobsData,
    cronJobsData,
    deploymentsPagination,
    podsPagination,
    statefulSetsPagination,
    daemonSetsPagination,
    jobsPagination,
    cronJobsPagination,
    currentData,
    currentPagination,
    loadClusters,
    loadNamespaces,
    loadDeployments,
    loadPods,
    loadStatefulSets,
    loadDaemonSets,
    loadJobs,
    loadCronJobs,
    loadCurrentData,
    handlePageChange,
    handlePageSizeChange,
    handleTabChange,
    init
  };
}
