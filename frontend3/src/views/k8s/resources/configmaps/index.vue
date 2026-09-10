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
  import { deleteK8sConfigMap, fetchK8sConfigMaps } from '@/service/api/k8s';
  import { useClusterNamespace, useListStateRestore } from '@/views/k8s/composables/useClusterNamespace';
  import ConfigMapDetailDialog from './modules/ConfigMapDetailDialog.vue';

  defineOptions({ name: 'K8sConfigMaps' });

  const router = useRouter();
  const route = useRoute();
  const message = ElMessage;

  const loading = ref(false);
  const dataSource = ref<K8s.ConfigMap[]>([]);
  const showDetail = ref(false);
  const currentDetail = ref<K8s.ConfigMap | null>(null);

  // 列表状态保存 key（跳转详情后返回时恢复集群/命名空间）
  const LIST_STATE_KEY = 'k8s_configmaps_list_state';

  // 集群/命名空间初始化与联动统一走 useClusterNamespace
  const clusterNs = useClusterNamespace({
    storageKey: LIST_STATE_KEY,
    // 集群/命名空间就绪后加载 ConfigMap 列表
    onDataReady: () => loadConfigMaps()
  });
  const { clusters, namespaces, selectedCluster, selectedNamespace, handleClusterChange, handleNamespaceChange } =
    clusterNs;

  // sessionStorage 状态恢复（返回列表时）与保存（跳转详情前）
  const { restoreState, saveState } = useListStateRestore(clusterNs, {
    onLoadList: () => loadConfigMaps()
  });

  const pagination = reactive({
    page: 1,
    pageSize: 20,
    itemCount: 0
  });

  // 加载 ConfigMap 列表
  const loadConfigMaps = async () => {
    if (!selectedCluster.value) {
      message.warning('请先选择集群');
      dataSource.value = [];
      pagination.itemCount = 0;
      return;
    }

    loading.value = true;
    try {
      const res = await fetchK8sConfigMaps(selectedCluster.value, {
        namespace: selectedNamespace.value
      });

      // flat 请求返回 {data: {list, total}, error}，列表与总数取自分页结构
      const configmaps = res.data?.list || [];
      dataSource.value = configmaps;
      pagination.itemCount = res.data?.total ?? configmaps.length;
    } catch (error: unknown) {
      dataSource.value = [];
      pagination.itemCount = 0;
      const err = error as Error;
      message.error(err.message || '加载 ConfigMap 列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 跳转到详情页
  const goToDetail = (row: K8s.ConfigMap) => {
    // 保存当前选择到 sessionStorage，返回列表时恢复
    saveState({
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      listPath: '/k8s/config'
    });

    router.push({
      path: '/k8s/resources/configmaps/detail',
      query: {
        clusterId: selectedCluster.value,
        namespace: selectedNamespace.value,
        name: row.name
      }
    });
  };

  // 删除 ConfigMap
  const handleDelete = async (row: K8s.ConfigMap) => {
    try {
      await ElMessageBox.confirm(`确定要删除 ConfigMap "${row.name}" 吗？`, '确认删除', {
        type: 'warning'
      });

      await deleteK8sConfigMap(selectedCluster.value!, {
        namespace: selectedNamespace.value,
        name: row.name
      });
      message.success('删除成功');
      loadConfigMaps();
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  };

  // 刷新
  const handleRefresh = () => {
    loadConfigMaps();
  };

  onMounted(() => {
    // 有存档则恢复集群/命名空间后加载列表，否则走 loadAll 正常初始化
    restoreState();
  });

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
          v-model="selectedNamespace"
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

    <!-- ConfigMap 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" min-width="200">
        <template #default="{ row }">
          <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="namespace" label="命名空间" width="150" />
      <ElTableColumn label="数据键" min-width="300">
        <template #default="{ row }">
          <div class="flex flex-wrap gap-1">
            <ElTag v-for="key in row.dataKeys" :key="key" size="small" type="info">
              {{ key }}
            </ElTag>
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
    <ConfigMapDetailDialog v-model:visible="showDetail" :detail="currentDetail" />
  </div>
</template>
