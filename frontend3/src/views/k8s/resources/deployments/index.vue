<script setup lang="ts">
  import { onMounted, reactive, ref, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import {
    ElButton,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElPagination,
    ElSelect,
    ElSpace,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus';
  import {
    deleteK8sDeployment,
    fetchK8sClusterNamespaces,
    fetchK8sClusters,
    fetchK8sDeployments,
    restartK8sDeployment
  } from '@/service/api/k8s';
  import ScaleDialog from './modules/ScaleDialog.vue';

  defineOptions({ name: 'K8sDeployments' });

  const router = useRouter();
  const route = useRoute();
  const message = ElMessage;

  const loading = ref(false);
  const dataSource = ref<K8s.Deployment[]>([]);
  const showScaleModal = ref(false);
  const scaleTarget = ref<K8s.Deployment | null>(null);

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
    pageSize: 10,
    itemCount: 0
  });

  // 状态显示
  const getStatusTag = (deployment: K8s.Deployment): { type: 'success' | 'info' | 'danger' | 'warning'; text: string } => {
    const ready = deployment.ready || 0;
    const total = deployment.replicas || 0;

    if (total === 0) {
      return { type: 'info', text: '未就绪' };
    }

    if (ready === total) {
      return { type: 'success', text: `运行中 (${ready}/${total})` };
    } else if (ready > 0) {
      return { type: 'warning', text: `部分就绪 (${ready}/${total})` };
    }
    return { type: 'danger', text: '未就绪' };
  };

  // 加载集群列表
  const loadClusters = async () => {
    try {
      const response = await fetchK8sClusters();
      // 从响应中提取集群数组（flat 请求返回 {data: {list, total}, error}）
      const clusterList = response?.data?.list || [];
      clusters.value = clusterList;

      // 如果有集群，默认选择第一个
      if (clusters.value.length > 0 && !selectedCluster.value) {
        selectedCluster.value = clusters.value[0].id;
        await loadNamespaces();
        await loadDeployments();
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
      const { data, error } = await fetchK8sClusterNamespaces(selectedCluster.value);

      // 提取命名空间数组
      let namespaceList: K8s.Namespace[] = [];

      if (!error && data && Array.isArray(data)) {
        namespaceList = data;
      } else {
        console.error('[调试] 无法从响应中提取数组');
        namespaceList = [];
      }

      // 使用普通数组而不是Vue Proxy对象
      const namespaceNames = namespaceList.map(ns => ns.name);
      namespaces.value = [...namespaceNames]; // 创建新数组避免Proxy问题

      if (namespaces.value.length > 0 && !namespaces.value.includes(filters.namespace)) {
        filters.namespace = namespaces.value[0];
        selectedNamespace.value = namespaces.value[0];
      }
    } catch (error: unknown) {
      console.error('[调试] 加载命名空间失败:', error);
      const err = error as Error;
      message.error(err.message || '加载命名空间列表失败');
    }
  };

  // 加载 Deployment 列表
  const loadDeployments = async () => {
    if (!selectedCluster.value) {
      message.warning('请先选择集群');
      dataSource.value = [];
      pagination.itemCount = 0;
      return;
    }

    loading.value = true;
    try {
      const { data, error } = await fetchK8sDeployments(selectedCluster.value, {
        namespace: filters.namespace,
        page: pagination.page,
        pageSize: pagination.pageSize
      });

      // 从返回数据中提取列表和总数
      const list = !error && data ? data?.list || [] : [];
      const total = !error && data ? data?.total || 0 : 0;

      dataSource.value = list;
      pagination.itemCount = total;

      if (total === 0) {
        console.warn(`[调试] 命名空间 "${filters.namespace}" 下没有 deployments`);
      }
    } catch (error: unknown) {
      console.error('[调试] 请求失败:', error);
      dataSource.value = [];
      pagination.itemCount = 0;
      const err = error as Error;
      message.error(err.message || '加载 Deployment 列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 集群变化
  const handleClusterChange = async () => {
    await loadNamespaces();
    await loadDeployments();
  };

  // 命名空间变化
  const handleNamespaceChange = () => {
    loadDeployments();
  };

  // 缩放 Deployment
  const handleScale = (row: K8s.Deployment) => {
    scaleTarget.value = row;
    showScaleModal.value = true;
  };

  // 重启 Deployment
  const handleRestart = async (row: K8s.Deployment) => {
    try {
      await ElMessageBox.confirm(`确定要重启 Deployment "${row.name}" 吗？`, '确认重启', {
        type: 'warning'
      });

      await restartK8sDeployment(selectedCluster.value!, {
        namespace: filters.namespace,
        name: row.name
      });
      message.success('重启成功');
      loadDeployments();
    } catch (error: unknown) {
      if ((error as string) !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '重启失败');
      }
    }
  };

  // 删除 Deployment
  const handleDelete = async (row: K8s.Deployment) => {
    try {
      await ElMessageBox.confirm(`确定要删除 Deployment "${row.name}" 吗？此操作危险，请谨慎操作。`, '确认删除', {
        type: 'warning',
        confirmButtonText: '危险操作确认',
        cancelButtonText: '取消'
      });

      await deleteK8sDeployment(selectedCluster.value!, {
        namespace: filters.namespace,
        name: row.name
      });
      message.success('删除成功');
      loadDeployments();
    } catch (error: unknown) {
      if ((error as string) !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  };

  // 刷新
  const handleRefresh = () => {
    loadDeployments();
  };

  // 分页变化
  const handlePageChange = (page: number) => {
    pagination.page = page;
    loadDeployments();
  };

  // 每页数量变化
  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.page = 1; // 重置到第一页
    loadDeployments();
  };

  // 查看详情
  const handleViewDetail = (row: K8s.Deployment) => {
    // 保存当前选择到 sessionStorage
    const stateToSave = {
      clusterId: selectedCluster.value,
      namespace: filters.namespace,
      listPath: '/k8s/workloads'
    };

    try {
      sessionStorage.setItem('k8s_deployments_list_state', JSON.stringify(stateToSave));
    } catch (e) {
      console.error('[Deployment列表] ❌ sessionStorage 保存失败:', e);
    }

    router.push({
      path: '/k8s/resources/deployments/detail',
      query: {
        clusterId: selectedCluster.value?.toString(),
        namespace: filters.namespace,
        name: row.name
      }
    });
  };

  // 从 sessionStorage 恢复状态
  const restoreStateFromStorage = () => {
    const savedState = sessionStorage.getItem('k8s_deployments_list_state');

    if (savedState) {
      try {
        const state = JSON.parse(savedState);

        selectedCluster.value = state.clusterId;
        filters.namespace = state.namespace;

        // 清除保存的状态
        sessionStorage.removeItem('k8s_deployments_list_state');

        // 加载命名空间和数据
        if (selectedCluster.value) {
          return loadNamespaces().then(() => {
            return loadDeployments();
          });
        }
      } catch (e) {
        console.error('[Deployment列表] 恢复状态失败:', e);
      }
    }

    // 正常加载流程
    return loadClusters();
  };

  onMounted(() => {
    restoreStateFromStorage();
  });

  // 监听路由变化，当从详情页返回时恢复状态
  watch(
    () => route.path,
    (newPath, oldPath) => {
      // 如果从详情页返回到列表页
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

    <!-- Deployment 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" min-width="200">
        <template #default="{ row }">
          <ElButton link type="primary" @click="handleViewDetail(row)">{{ row.name }}</ElButton>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="namespace" label="命名空间" width="150" />
      <ElTableColumn label="状态" width="150">
        <template #default="{ row }">
          <ElTag :type="getStatusTag(row).type">
            {{ getStatusTag(row).text }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="副本数" width="120" align="center">
        <template #default="{ row }">{{ row.ready || 0 }} / {{ row.replicas || 0 }}</template>
      </ElTableColumn>
      <ElTableColumn prop="age" label="年龄" width="160" />
      <ElTableColumn label="操作" align="center" width="280" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <ElButton link type="primary" size="small" @click="handleViewDetail(row)">详情</ElButton>
          <PermissionButton link type="primary" size="small" code="k8s.resource.update" @click="handleScale(row)">
            缩放
          </PermissionButton>
          <PermissionButton link type="primary" size="small" code="k8s.resource.update" @click="handleRestart(row)">
            重启
          </PermissionButton>
          <PermissionButton link type="danger" size="small" code="k8s.resource.delete" @click="handleDelete(row)">
            删除
          </PermissionButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 分页 -->
    <div class="mt-4 flex justify-end">
      <ElPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.itemCount"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handlePageSizeChange"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 缩放对话框 -->
    <ScaleDialog
      v-model:visible="showScaleModal"
      :cluster-id="selectedCluster"
      :namespace="filters.namespace"
      :deployment="scaleTarget"
      @submitted="loadDeployments"
    />
  </div>
</template>
