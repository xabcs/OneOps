<script setup lang="ts">
    import { onMounted, reactive, ref, watch } from 'vue';
    import { useRoute, useRouter } from 'vue-router';
    import {
      ElButton,
      ElDialog,
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
      deleteK8sConfigMap,
      fetchK8sClusterNamespaces,
      fetchK8sClusters,
      fetchK8sConfigMaps
    } from '@/service/api/k8s';

    defineOptions({ name: 'K8sConfigMaps' });

    const router = useRouter();
    const route = useRoute();
    const message = ElMessage;

    const loading = ref(false);
    const dataSource = ref<K8s.ConfigMap[]>([]);
    const showDetail = ref(false);
    const currentDetail = ref<K8s.ConfigMap | null>(null);

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

    // 加载集群列表
    const loadClusters = async () => {
      try {
        const res = await fetchK8sClusters();
        clusters.value = res.data || [];

        // 如果有集群，默认选择第一个
        if (clusters.value.length > 0 && !selectedCluster.value) {
          selectedCluster.value = clusters.value[0].id;
          await loadNamespaces();
          await loadConfigMaps();
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
          namespace: filters.namespace
        });

        // 处理不同的响应格式
        const configmaps = Array.isArray(res) ? res : res?.data || [];
        dataSource.value = configmaps;
        pagination.itemCount = configmaps?.length || 0;
      } catch (error: unknown) {
        dataSource.value = [];
        pagination.itemCount = 0;
        const err = error as Error;
        message.error(err.message || '加载 ConfigMap 列表失败');
      } finally {
        loading.value = false;
      }
    };

    // 集群变化
    const handleClusterChange = async () => {
      await loadNamespaces();
      await loadConfigMaps();
    };

    // 命名空间变化
    const handleNamespaceChange = () => {
      loadConfigMaps();
    };

    // 跳转到详情页
    const goToDetail = (row: K8s.ConfigMap) => {
      sessionStorage.setItem(
        'k8s_configmaps_list_state',
        JSON.stringify({
          clusterId: selectedCluster.value,
          namespace: filters.namespace,
          listPath: '/k8s/config'
        })
      );

      router.push({
        path: '/k8s/resources/configmaps/detail',
        query: {
          clusterId: selectedCluster.value,
          namespace: filters.namespace,
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
          namespace: filters.namespace,
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

    // 从 sessionStorage 恢复状态
    const restoreStateFromStorage = () => {
      const savedState = sessionStorage.getItem('k8s_configmaps_list_state');
      if (savedState) {
        try {
          const state = JSON.parse(savedState);
          selectedCluster.value = state.clusterId;
          filters.namespace = state.namespace;

          sessionStorage.removeItem('k8s_configmaps_list_state');

          if (selectedCluster.value) {
            return loadNamespaces().then(() => {
              return loadConfigMaps();
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
                <ElSelect v-model="filters.namespace" placeholder="请选择命名空间" style="width: 180px" @change="handleNamespaceChange">
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
            <ElTableColumn label="操作" width="180" fixed="right">
                <template #default="{ row }">
                    <ElSpace>
                        <ElButton size="small" type="primary" @click="goToDetail(row)">详情</ElButton>
                        <ElButton size="small" type="danger" @click="handleDelete(row)">删除</ElButton>
                    </ElSpace>
                </template>
            </ElTableColumn>
        </ElTable>

        <!-- 详情弹窗 -->
        <ElDialog v-model="showDetail" :title="`ConfigMap: ${currentDetail?.name}`" width="800px">
            <div v-if="currentDetail" class="space-y-4">
                <!-- 基本信息 -->
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <span class="text-sm text-gray-700 font-medium">名称:</span>
                        <span class="ml-2">{{ currentDetail.name }}</span>
                    </div>
                    <div>
                        <span class="text-sm text-gray-700 font-medium">命名空间:</span>
                        <span class="ml-2">{{ currentDetail.namespace }}</span>
                    </div>
                    <div>
                        <span class="text-sm text-gray-700 font-medium">年龄:</span>
                        <span class="ml-2">{{ currentDetail.age }}</span>
                    </div>
                </div>

                <!-- Labels -->
                <div v-if="currentDetail.labels && Object.keys(currentDetail.labels).length > 0">
                    <h4 class="mb-2 text-sm text-gray-700 font-medium">标签:</h4>
                    <div class="flex flex-wrap gap-2">
                        <ElTag v-for="(value, key) in currentDetail.labels" :key="key" type="info">{{ key }}: {{ value }}</ElTag>
                    </div>
                </div>

                <!-- Data -->
                <div v-if="currentDetail.data && Object.keys(currentDetail.data).length > 0">
                    <h4 class="mb-2 text-sm text-gray-700 font-medium">数据:</h4>
                    <div class="space-y-2">
                        <div v-for="(value, key) in currentDetail.data" :key="key" class="border rounded p-2">
                            <div class="mb-1 text-sm text-gray-700 font-medium">{{ key }}</div>
                            <div class="whitespace-pre-wrap break-all rounded bg-gray-50 p-2 text-sm font-mono">
                                {{ value }}
                            </div>
                        </div>
                    </div>
                </div>

                <!-- YAML Manifest -->
                <div>
                    <h4 class="mb-2 text-sm text-gray-700 font-medium">YAML 配置:</h4>
                    <div class="max-h-400 overflow-auto whitespace-pre rounded bg-gray-50 p-3 text-sm font-mono">
                        {{ currentDetail.manifest }}
                    </div>
                </div>
            </div>
        </ElDialog>
    </div>
</template>
