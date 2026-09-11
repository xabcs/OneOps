<script setup lang="ts">
  import { onMounted, reactive, ref, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import {
    ElButton,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElSelect,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus';
  import { deleteK8sPod, fetchK8sPods } from '@/service/api/k8s';
  import { useClusterNamespace, useListStateRestore } from '@/views/k8s/composables/useClusterNamespace';
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

  // 列表状态保存 key（跳转详情后返回时恢复集群/命名空间）
  const LIST_STATE_KEY = 'k8s_pods_list_state';

  // 集群/命名空间初始化与联动统一走 useClusterNamespace
  const clusterNs = useClusterNamespace({
    storageKey: LIST_STATE_KEY,
    // 集群/命名空间就绪后加载 Pod 列表
    onDataReady: () => loadPods()
  });
  const { clusters, namespaces, selectedCluster, selectedNamespace, handleClusterChange, handleNamespaceChange } =
    clusterNs;

  // sessionStorage 状态恢复（返回列表时）与保存（跳转详情前）
  const { restoreState, saveState } = useListStateRestore(clusterNs, {
    onLoadList: () => loadPods()
  });

  const filters = reactive({
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
  const getStatusTag = (pod: K8s.Pod): { type: 'success' | 'info' | 'danger' | 'warning'; text: string } => {
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
        namespace: selectedNamespace.value,
        labelSelector: filters.labelSelector
      });

      // flat 请求返回 {data: {list, total}, error}，列表与总数取自分页结构
      const pods = res.data?.list || [];
      dataSource.value = pods;
      pagination.itemCount = res.data?.total ?? pods.length;
    } catch (error: unknown) {
      dataSource.value = [];
      pagination.itemCount = 0;
      const err = error as Error;
      message.error(err.message || '加载 Pod 列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 搜索
  const handleSearch = () => {
    pagination.page = 1;
    loadPods();
  };

  // 重置筛选（集群/命名空间为作用域，仅重置标签选择器）
  const handleReset = () => {
    filters.labelSelector = '';
    pagination.page = 1;
    loadPods();
  };

  // 进入终端
  const handleTerminal = (row: K8s.Pod, containerName?: string) => {
    terminalProps.value = {
      clusterId: selectedCluster.value!,
      namespace: selectedNamespace.value,
      podName: row.name,
      containerName: containerName || ''
    };
    showTerminal.value = true;
  };

  // 查看日志
  const handleLogs = (row: K8s.Pod, containerName?: string) => {
    logProps.value = {
      clusterId: selectedCluster.value!,
      namespace: selectedNamespace.value,
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
        namespace: selectedNamespace.value,
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
    // 保存当前选择到 sessionStorage，返回列表时恢复
    saveState({
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      listPath: '/k8s/workloads'
    });

    router.push({
      path: '/k8s/resources/pods/detail',
      query: {
        clusterId: selectedCluster.value?.toString(),
        namespace: selectedNamespace.value,
        name: row.name
      }
    });
  };

  // 刷新
  const handleRefresh = () => {
    loadPods();
  };

  onMounted(() => {
    restoreState();
  });

  // 监听路由变化
  watch(
    () => route.path,
    (newPath, oldPath) => {
      if (newPath === '/k8s/workloads' && oldPath?.includes('/detail')) {
        restoreState();
      }
    }
  );
</script>

<template>
  <ListPageLayout
    title="容器组（Pod）"
    description="查看集群内容器组运行状态，支持日志查看、终端连接与删除"
    @search="handleRefresh"
    @reset="handleReset"
  >
    <!-- 搜索筛选：集群 / 命名空间 / 标签选择器 -->
    <template #search>
      <ElSelect v-model="selectedCluster" placeholder="请选择集群" style="width: 200px" @change="handleClusterChange">
        <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
      </ElSelect>
      <ElSelect
        v-model="selectedNamespace"
        placeholder="请选择命名空间"
        style="width: 180px"
        @change="handleNamespaceChange"
      >
        <ElOption v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
      </ElSelect>
      <ElInput
        v-model="filters.labelSelector"
        placeholder="标签选择器 (可选)"
        style="width: 200px"
        clearable
        @change="handleSearch"
      />
    </template>

    <!-- Pod 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe height="100%">
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
      <ElTableColumn label="操作" align="center" width="350" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <ElButton link type="primary" size="small" @click="handleViewDetail(row)">查看详情</ElButton>
          <PermissionButton link type="primary" size="small" code="k8s.terminal.connect" @click="handleTerminal(row)">
            进入终端
          </PermissionButton>
          <ElButton link type="primary" size="small" @click="handleLogs(row)">查看日志</ElButton>
          <PermissionButton link type="danger" size="small" code="k8s.resource.delete" @click="handleDelete(row)">
            删除
          </PermissionButton>
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
  </ListPageLayout>
</template>
