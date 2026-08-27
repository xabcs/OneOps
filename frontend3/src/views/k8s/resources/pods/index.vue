<script setup lang="ts">
  import { onMounted, reactive, ref, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import {
    ElButton,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElSelect,
    ElSpace,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus';
  import {
    deleteK8sPod,
    fetchK8sClusterNamespaces,
    fetchK8sClusters,
    fetchK8sPods
  } from '@/service/api/k8s';
  import PodLogDialog from './modules/PodLogDialog.vue';
  import PodTerminalDialog from './modules/PodTerminalDialog.vue';

  defineOptions({ name: 'K8sPods' });

  const router = useRouter();
  const route = useRoute();
  const message = ElMessage;

  const loading = ref(false);
  const dataSource = ref<K8s.Pod[]>([]);
  const showTerminal = ref(false);
  const showLogs = ref(false);

  // 当前选中的集群和命名空间
  const selectedCluster = ref<number | null>(null);
  const selectedNamespace = ref('default');

  // 可用的命名空间列表
  const namespaces = ref<string[]>([]);

  // 可用的集群列表
  const clusters = ref<K8s.Cluster[]>([]);

  const filters = reactive({
    namespace: 'default',
    labelSelector: ''
  });

  // 终端/日志相关
  const terminalProps = ref({
    clusterId: 0,
    namespace: '',
    podName: '',
    containerName: ''
  });

  const logProps = ref({
    clusterId: 0,
    namespace: '',
    podName: '',
    containerName: ''
  });

  const pagination = reactive({
    page: 1,
    pageSize: 20,
    itemCount: 0
  });

  // 状态显示
  const getStatusTag = (pod: K8s.Pod) => {
    const phase = pod.phase || 'Unknown';
    const status = pod.status || phase;

    switch (status) {
      case 'Running':
        return { type: 'success', text: '运行中' };
      case 'Succeeded':
        return { type: 'info', text: '已完成' };
      case 'Failed':
        return { type: 'danger', text: '失败' };
      case 'Pending':
        return { type: 'warning', text: '等待中' };
      case 'Unknown':
      default:
        return { type: 'info', text: '未知' };
    }
  };

  // 加载集群列表
  const loadClusters = async () => {
    try {
      const res = await fetchK8sClusters();
      clusters.value = res.data || [];

      // 如果有集群，默认选择第一个
      if (clusters.value.length > 0 && !selectedCluster.value) {
        selectedCluster.value = clusters.value[0].id;
        await loadNamespaces();
        await loadPods();
      }
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '加载集群列表失败');
    }
  };

  // 加载命名空间列表
  const loadNamespaces = async () => {
    if (!selectedCluster.value) return;
    try {
      const res = await fetchK8sClusterNamespaces(selectedCluster.value);
      const namespaceList = res.data || res || [];
      const namespaceNames = namespaceList.map((ns: K8s.Namespace) => ns.name);
      namespaces.value = [...namespaceNames];

      if (namespaces.value.length > 0 && !namespaces.value.includes(filters.namespace)) {
        filters.namespace = namespaces.value[0];
        selectedNamespace.value = namespaces.value[0];
      }
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '加载命名空间列表失败');
    }
  };

  // 加载 Pod 列表
  const loadPods = async () => {
    if (!selectedCluster.value) {
      message.warning('请先选择集群');
      dataSource.value = [];
      pagination.itemCount = 0;
      return;
    }

    loading.value = true;
    try {
      const res = await fetchK8sPods(selectedCluster.value, {
        namespace: filters.namespace,
        labelSelector: filters.labelSelector
      });

      // 处理不同的响应格式
      const pods = Array.isArray(res) ? res : res?.data || [];
      dataSource.value = pods;
      pagination.itemCount = pods?.length || 0;
    } catch (error: unknown) {
      dataSource.value = [];
      pagination.itemCount = 0;
      const err = error as Error;
      message.error(err.message || '加载 Pod 列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 集群变化
  const handleClusterChange = async () => {
    await loadNamespaces();
    await loadPods();
  };

  // 命名空间变化
  const handleNamespaceChange = () => {
    loadPods();
  };

  // 搜索
  const handleSearch = () => {
    pagination.page = 1;
    loadPods();
  };

  // 进入终端
  const handleTerminal = (row: K8s.Pod, containerName?: string) => {
    terminalProps.value = {
      clusterId: selectedCluster.value!,
      namespace: filters.namespace,
      podName: row.name,
      containerName: containerName || ''
    };
    showTerminal.value = true;
  };

  // 查看日志
  const handleLogs = (row: K8s.Pod, containerName?: string) => {
    logProps.value = {
      clusterId: selectedCluster.value!,
      namespace: filters.namespace,
      podName: row.name,
      containerName: containerName || ''
    };
    showLogs.value = true;
  };

  // 删除 Pod
  const handleDelete = async (row: K8s.Pod) => {
    try {
      await ElMessageBox.confirm(`确定要删除 Pod "${row.name}" 吗？`, '确认删除', {
        type: 'warning'
      });

      await deleteK8sPod(selectedCluster.value!, {
        namespace: filters.namespace,
        name: row.name
      });
      message.success('删除成功');
      loadPods();
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  };

  // 查看详情
  const handleViewDetail = (row: K8s.Pod) => {
    // 保存当前选择到 sessionStorage
    // 保存当前选择到 sessionStorage
    const stateToSave = {
      clusterId: selectedCluster.value,
      namespace: filters.namespace,
      listPath: '/k8s/workloads'
    };
    sessionStorage.setItem('k8s_pods_list_state', JSON.stringify(stateToSave));

    router.push({
      path: '/k8s/resources/pods/detail',
      query: {
        clusterId: selectedCluster.value?.toString(),
        namespace: filters.namespace,
        name: row.name
      }
    });
  };

  // 刷新
  const handleRefresh = () => {
    loadPods();
  };

  // 从 sessionStorage 恢复状态
  const restoreStateFromStorage = () => {
    const savedState = sessionStorage.getItem('k8s_pods_list_state');

    if (savedState) {
      try {
        const state = JSON.parse(savedState);

        selectedCluster.value = state.clusterId;
        filters.namespace = state.namespace;

        // 清除保存的状态
        sessionStorage.removeItem('k8s_pods_list_state');

        // 加载命名空间和数据
        if (selectedCluster.value) {
          return loadNamespaces().then(() => {
            return loadPods();
          });
        }
      } catch (e) {
        console.error('[Pod列表] 恢复状态失败:', e);
      }
    }
    // 正常加载流程
    return loadClusters();
  };

  onMounted(() => {
    restoreStateFromStorage();
  });

  // 监听路由变化
  watch(
    () => route.path,
    (newPath, oldPath) => {
      if (newPath === '/k8s/workloads' && oldPath?.includes('/detail')) {
        restoreStateFromStorage();
      }
    }
  );
</script>

<template>
  <div class="p-4">
    <!-- 筛选栏 -->
    <div class="bg-card mb-4 flex items-center gap-4 rounded-lg p-4">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">集群:</span>
        <ElSelect v-model="selectedCluster" placeholder="请选择集群" style="width: 200px" @change="handleClusterChange">
          <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
        </ElSelect>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">命名空间:</span>
        <ElSelect
          v-model="filters.namespace"
          placeholder="请选择命名空间"
          style="width: 180px"
          @change="handleNamespaceChange"
        >
          <ElOption v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
        </ElSelect>
      </div>

      <div class="flex items-center gap-2">
        <ElInput
          v-model="filters.labelSelector"
          placeholder="标签选择器 (可选)"
          style="width: 200px"
          clearable
          @change="handleSearch"
        />
      </div>

      <div class="flex-1" />

      <ElButton type="primary" :disabled="!selectedCluster" @click="handleRefresh">刷新</ElButton>
    </div>

    <!-- Pod 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" min-width="200">
        <template #default="{ row }">
          <ElButton link type="primary" @click="handleViewDetail(row)">{{ row.name }}</ElButton>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="namespace" label="命名空间" width="150" />
      <ElTableColumn label="状态" width="120">
        <template #default="{ row }">
          <ElTag :type="getStatusTag(row).type">
            {{ getStatusTag(row).text }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="ip" label="IP 地址" width="140" />
      <ElTableColumn prop="node" label="节点" width="180" />
      <ElTableColumn label="重启次数" width="100" align="center">
        <template #default="{ row }">
          {{ row.restarts || 0 }}
        </template>
      </ElTableColumn>
      <ElTableColumn prop="age" label="年龄" width="160" />
      <ElTableColumn label="操作" width="350" fixed="right">
        <template #default="{ row }">
          <ElSpace wrap>
            <ElButton size="small" @click="handleViewDetail(row)">查看详情</ElButton>
            <PermissionButton code="k8s.terminal.connect" size="small" type="primary" @click="handleTerminal(row)">进入终端</PermissionButton>
            <ElButton size="small" @click="handleLogs(row)">查看日志</ElButton>
            <PermissionButton code="k8s.resource.delete" size="small" type="danger" @click="handleDelete(row)">删除</PermissionButton>
          </ElSpace>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 日志弹窗 -->
    <PodLogDialog
      v-model:visible="showLogs"
      :cluster-id="logProps.clusterId"
      :namespace="logProps.namespace"
      :pod-name="logProps.podName"
      :container-name="logProps.containerName"
    />

    <!-- 终端弹窗 -->
    <PodTerminalDialog
      v-model:visible="showTerminal"
      :cluster-id="terminalProps.clusterId"
      :namespace="terminalProps.namespace"
      :pod-name="terminalProps.podName"
      :container-name="terminalProps.containerName"
    />
  </div>
</template>
