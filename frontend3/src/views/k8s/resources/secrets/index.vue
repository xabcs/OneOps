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
  import { deleteK8sSecret, fetchK8sClusterNamespaces, fetchK8sClusters, fetchK8sSecrets } from '@/service/api/k8s';
  import SecretDetailDialog from './modules/SecretDetailDialog.vue';

  defineOptions({ name: 'K8sSecrets' });

  const router = useRouter();
  const route = useRoute();
  const message = ElMessage;

  const loading = ref(false);
  const dataSource = ref<K8s.Secret[]>([]);
  const showDetail = ref(false);
  const currentDetail = ref<K8s.Secret | null>(null);

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

  // Secret 类型显示
  const getSecretTypeTag = (type: string) => {
    switch (type) {
      case 'Opaque':
        return { type: 'info', text: 'Opaque' };
      case 'kubernetes.io/service-account-token':
        return { type: 'success', text: 'Service Account' };
      case 'kubernetes.io/dockercfg':
        return { type: 'warning', text: 'Docker Config' };
      case 'kubernetes.io/dockerconfigjson':
        return { type: 'warning', text: 'Docker Config JSON' };
      case 'kubernetes.io/tls':
        return { type: 'primary', text: 'TLS' };
      case 'bootstrap.kubernetes.io/token':
        return { type: 'success', text: 'Bootstrap Token' };
      default:
        return { type: 'info', text: type || 'Unknown' };
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
        await loadSecrets();
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

  // 加载 Secret 列表
  const loadSecrets = async () => {
    if (!selectedCluster.value) {
      message.warning('请先选择集群');
      dataSource.value = [];
      pagination.itemCount = 0;
      return;
    }

    loading.value = true;
    try {
      const res = await fetchK8sSecrets(selectedCluster.value, {
        namespace: filters.namespace
      });

      // 处理不同的响应格式
      const secrets = Array.isArray(res) ? res : res?.data || [];
      dataSource.value = secrets;
      pagination.itemCount = secrets?.length || 0;
    } catch (error: unknown) {
      dataSource.value = [];
      pagination.itemCount = 0;
      const err = error as Error;
      message.error(err.message || '加载 Secret 列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 集群变化
  const handleClusterChange = async () => {
    await loadNamespaces();
    await loadSecrets();
  };

  // 命名空间变化
  const handleNamespaceChange = () => {
    loadSecrets();
  };

  // 跳转到详情页
  const goToDetail = (row: K8s.Secret) => {
    sessionStorage.setItem(
      'k8s_secrets_list_state',
      JSON.stringify({
        clusterId: selectedCluster.value,
        namespace: filters.namespace,
        listPath: '/k8s/config'
      })
    );

    router.push({
      path: '/k8s/resources/secrets/detail',
      query: {
        clusterId: selectedCluster.value,
        namespace: filters.namespace,
        name: row.name
      }
    });
  };

  // 删除 Secret
  const handleDelete = async (row: K8s.Secret) => {
    try {
      await ElMessageBox.confirm(
        `确定要删除 Secret "${row.name}" 吗？删除后可能导致相关服务无法正常运行。`,
        '确认删除',
        {
          type: 'warning',
          confirmButtonText: '危险操作确认',
          cancelButtonText: '取消'
        }
      );

      await deleteK8sSecret(selectedCluster.value!, {
        namespace: filters.namespace,
        name: row.name
      });
      message.success('删除成功');
      loadSecrets();
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  };

  // 刷新
  const handleRefresh = () => {
    loadSecrets();
  };

  // 从 sessionStorage 恢复状态
  const restoreStateFromStorage = () => {
    const savedState = sessionStorage.getItem('k8s_secrets_list_state');
    if (savedState) {
      try {
        const state = JSON.parse(savedState);
        selectedCluster.value = state.clusterId;
        filters.namespace = state.namespace;

        sessionStorage.removeItem('k8s_secrets_list_state');

        if (selectedCluster.value) {
          return loadNamespaces().then(() => {
            return loadSecrets();
          });
        }
      } catch (e) {
        console.error('Failed to restore state:', e);
      }
    }
    return loadClusters();
  };

  onMounted(() => {
    restoreStateFromStorage();
  });

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

    <!-- Secret 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" min-width="220">
        <template #default="{ row }">
          <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="namespace" label="命名空间" width="150" />
      <ElTableColumn label="类型" width="200">
        <template #default="{ row }">
          <ElTag :type="getSecretTypeTag(row.type).type">
            {{ getSecretTypeTag(row.type).text }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="数据键" min-width="250">
        <template #default="{ row }">
          <div class="flex flex-wrap gap-1">
            <ElTag v-for="key in row.dataKeys" :key="key" size="small" type="warning">
              {{ key }}
            </ElTag>
          </div>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="age" label="年龄" width="160" />
      <ElTableColumn label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <ElSpace>
            <ElButton size="small" type="primary" @click="goToDetail(row)">详情</ElButton>
            <PermissionButton code="k8s.resource.delete" size="small" type="danger" @click="handleDelete(row)">删除</PermissionButton>
          </ElSpace>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 详情弹窗 -->
    <SecretDetailDialog v-model:visible="showDetail" :detail="currentDetail" />
  </div>
</template>
