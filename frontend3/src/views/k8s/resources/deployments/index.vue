<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue';
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
  import { deleteK8sDeployment, fetchK8sDeployments, restartK8sDeployment } from '@/service/api/k8s';
  import { useClusterNamespace, useListStateRestore } from '@/views/k8s/composables/useClusterNamespace';
  import ScaleDialog from './modules/ScaleDialog.vue';

  defineOptions({ name: 'K8sDeployments' });

  const router = useRouter();
  const route = useRoute();
  const message = ElMessage;

  const loading = ref(false);
  const dataSource = ref<K8s.Deployment[]>([]);
  const showScaleModal = ref(false);
  const scaleTarget = ref<K8s.Deployment | null>(null);

  // 列表状态保存 key（跳转详情后返回时恢复集群/命名空间）
  const LIST_STATE_KEY = 'k8s_deployments_list_state';

  // 集群/命名空间初始化与联动统一走 useClusterNamespace
  const clusterNs = useClusterNamespace({
    storageKey: LIST_STATE_KEY,
    // 集群/命名空间就绪后加载 Deployment 列表
    onDataReady: () => loadDeployments()
  });
  const { clusters, namespaces, selectedCluster, selectedNamespace, handleClusterChange, handleNamespaceChange } =
    clusterNs;

  // sessionStorage 状态恢复（返回列表时）与保存（跳转详情前）
  const { restoreState, saveState } = useListStateRestore(clusterNs, {
    onLoadList: () => loadDeployments()
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
        namespace: selectedNamespace.value,
        page: pagination.page,
        pageSize: pagination.pageSize
      });

      // 从返回数据中提取列表和总数
      const list = !error && data ? data?.list || [] : [];
      const total = !error && data ? data?.total || 0 : 0;

      dataSource.value = list;
      pagination.itemCount = total;

      if (total === 0) {
        console.warn(`[调试] 命名空间 "${selectedNamespace.value}" 下没有 deployments`);
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
        namespace: selectedNamespace.value,
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
        namespace: selectedNamespace.value,
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

  // 布局分页适配对象：受控状态 + 变更处理器（size-change 内显式回写 pageSize 保证受控同步）
  const layoutPagination = computed(() => ({
    currentPage: pagination.page,
    pageSize: pagination.pageSize,
    pageSizes: [10, 20, 50, 100],
    total: pagination.itemCount,
    'current-change': handlePageChange,
    'size-change': handlePageSizeChange
  }));

  // 查看详情
  const handleViewDetail = (row: K8s.Deployment) => {
    // 保存当前选择到 sessionStorage，返回列表时恢复
    saveState({
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      listPath: '/k8s/workloads'
    });

    router.push({
      path: '/k8s/resources/deployments/detail',
      query: {
        clusterId: selectedCluster.value?.toString(),
        namespace: selectedNamespace.value,
        name: row.name
      }
    });
  };

  onMounted(() => {
    restoreState();
  });

  // 监听路由变化，当从详情页返回时恢复状态
  watch(
    () => route.path,
    (newPath, oldPath) => {
      // 如果从详情页返回到列表页
      if (newPath === '/k8s/workloads' && oldPath?.includes('/detail')) {
        restoreState();
      }
    }
  );
</script>

<template>
  <ListPageLayout
    title="部署（Deployment）"
    description="管理集群 Deployment 工作负载，支持缩放、重启与删除"
    :pagination="layoutPagination"
    @search="handleRefresh"
  >
    <!-- 搜索筛选：集群 / 命名空间 -->
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
    </template>

    <!-- Deployment 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe height="100%">
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

    <!-- 缩放对话框 -->
    <ScaleDialog
      v-model:visible="showScaleModal"
      :cluster-id="selectedCluster"
      :namespace="selectedNamespace"
      :deployment="scaleTarget"
      @submitted="loadDeployments"
    />
  </ListPageLayout>
</template>
