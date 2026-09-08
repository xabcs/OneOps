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
  import { deleteK8sService, fetchK8sClusterNamespaces, fetchK8sClusters, fetchK8sServices } from '@/service/api/k8s';
  import ServiceDetailDialog from './modules/ServiceDetailDialog.vue';

  defineOptions({ name: 'K8sServices' });

  const router = useRouter();
  const route = useRoute();
  const message = ElMessage;

  const loading = ref(false);
  const dataSource = ref<K8s.Service[]>([]);
  const showDetail = ref(false);
  const currentDetail = ref<K8s.Service | null>(null);

  // 当前选中的集群和命名空间
  const selectedCluster = ref<number | null>(null);
  const selectedNamespace = ref('default');

  // 可用的命名空间列表
  const namespaces = ref<string[]>([]);

  // 可用的集群列表
  const clusters = ref<K8s.Cluster[]>([]);

  const filters = reactive({
    namespace: 'default'
  });

  const pagination = reactive({
    page: 1,
    pageSize: 20,
    itemCount: 0
  });

  // 服务类型显示
  const getServiceTypeTag = (type: string) => {
    switch (type) {
      case 'ClusterIP':
        return { type: 'primary', text: 'ClusterIP' };
      case 'NodePort':
        return { type: 'success', text: 'NodePort' };
      case 'LoadBalancer':
        return { type: 'warning', text: 'LoadBalancer' };
      case 'ExternalName':
        return { type: 'info', text: 'ExternalName' };
      default:
        return { type: 'info', text: type || 'Unknown' };
    }
  };

  // 格式化端口信息
  const formatPorts = (ports: K8s.ServicePort[]) => {
    if (!ports || ports.length === 0) return '-';

    return ports
      .map(p => {
        const parts = [];
        if (p.name) parts.push(p.name);
        parts.push(`${p.port}/${p.protocol}`);
        if (p.nodePort) parts.push(`NodePort: ${p.nodePort}`);
        return parts.join(' - ');
      })
      .join(', ');
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
        await loadServices();
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
      const namespaceList = await fetchK8sClusterNamespaces(selectedCluster.value);
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

  // 加载 Service 列表
  const loadServices = async () => {
    if (!selectedCluster.value) {
      message.warning('请先选择集群');
      dataSource.value = [];
      pagination.itemCount = 0;
      return;
    }

    loading.value = true;
    try {
      const res = await fetchK8sServices(selectedCluster.value, {
        namespace: filters.namespace
      });

      // 处理不同的响应格式
      const services = Array.isArray(res) ? res : res?.data || [];
      dataSource.value = services;
      pagination.itemCount = services?.length || 0;
    } catch (error: unknown) {
      dataSource.value = [];
      pagination.itemCount = 0;
      const err = error as Error;
      message.error(err.message || '加载 Service 列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 集群变化
  const handleClusterChange = async () => {
    await loadNamespaces();
    await loadServices();
  };

  // 命名空间变化
  const handleNamespaceChange = () => {
    loadServices();
  };

  // 跳转到详情页
  const goToDetail = (row: K8s.Service) => {
    // 保存当前选择到 sessionStorage
    sessionStorage.setItem(
      'k8s_services_list_state',
      JSON.stringify({
        clusterId: selectedCluster.value,
        namespace: filters.namespace,
        listPath: '/k8s/services'
      })
    );

    router.push({
      path: '/k8s/resources/services/detail',
      query: {
        clusterId: selectedCluster.value,
        namespace: filters.namespace,
        name: row.name
      }
    });
  };

  // 删除 Service
  const handleDelete = async (row: K8s.Service) => {
    try {
      await ElMessageBox.confirm(`确定要删除 Service "${row.name}" 吗？`, '确认删除', {
        type: 'warning'
      });

      await deleteK8sService(selectedCluster.value!, {
        namespace: filters.namespace,
        name: row.name
      });
      message.success('删除成功');
      loadServices();
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  };

  // 刷新
  const handleRefresh = () => {
    loadServices();
  };

  // 从 sessionStorage 恢复状态
  const restoreStateFromStorage = () => {
    const savedState = sessionStorage.getItem('k8s_services_list_state');
    if (savedState) {
      try {
        const state = JSON.parse(savedState);
        selectedCluster.value = state.clusterId;
        filters.namespace = state.namespace;

        // 清除保存的状态
        sessionStorage.removeItem('k8s_services_list_state');

        // 加载命名空间和数据
        if (selectedCluster.value) {
          return loadNamespaces().then(() => {
            return loadServices();
          });
        }
      } catch (e) {
        console.error('Failed to restore state:', e);
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

      <div class="flex-1" />

      <ElButton type="primary" :disabled="!selectedCluster" @click="handleRefresh">刷新</ElButton>
    </div>

    <!-- Service 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" min-width="180">
        <template #default="{ row }">
          <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="namespace" label="命名空间" width="150" />
      <ElTableColumn label="类型" width="140">
        <template #default="{ row }">
          <ElTag :type="getServiceTypeTag(row.type).type">
            {{ getServiceTypeTag(row.type).text }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="clusterIP" label="Cluster IP" width="150" />
      <ElTableColumn label="端口" min-width="250">
        <template #default="{ row }">
          <div class="truncate text-sm text-gray-600" :title="formatPorts(row.ports)">
            {{ formatPorts(row.ports) }}
          </div>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="age" label="年龄" width="160" />
      <ElTableColumn label="操作" align="center" width="180" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <ElButton link type="primary" size="small" @click="goToDetail(row)">详情</ElButton>
          <PermissionButton link type="danger" size="small" code="k8s.resource.delete" @click="handleDelete(row)">
            删除
          </PermissionButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 详情弹窗 -->
    <ServiceDetailDialog v-model:visible="showDetail" :detail="currentDetail" />
  </div>
</template>
