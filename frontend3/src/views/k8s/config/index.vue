<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import { ElButton, ElPagination, ElTabPane, ElTabs, ElTag } from 'element-plus';
  import { fetchK8sClusterNamespaces, fetchK8sClusters, fetchK8sConfigMaps, fetchK8sSecrets } from '@/service/api/k8s';

  defineOptions({ name: 'K8sConfig' });

  const loading = ref(false);
  const activeTab = ref('configmaps');

  // 当前选中的集群和命名空间
  const selectedCluster = ref<number | null>(null);
  const selectedNamespace = ref('default');

  // 可用的命名空间列表
  const namespaces = ref<string[]>([]);

  // 可用的集群列表
  const clusters = ref<K8s.Cluster[]>([]);

  // 资源数据
  const configmapsData = ref<K8s.ConfigMap[]>([]);
  const secretsData = ref<K8s.Secret[]>([]);

  // 分页
  const configmapsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
  const secretsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

  // 当前数据
  const currentPagination = computed(() => {
    return activeTab.value === 'configmaps' ? configmapsPagination : secretsPagination;
  });

  // 加载集群列表
  async function loadClusters() {
    const { data, error } = await fetchK8sClusters();
    if (!error) {
      clusters.value = data?.list || [];
      // 自动选中第一个集群
      if (clusters.value.length > 0 && !selectedCluster.value) {
        selectedCluster.value = clusters.value[0].id;
      }
    } else {
      console.error('加载集群列表失败:', error);
    }
  }

  // 加载命名空间列表
  async function loadNamespaces() {
    if (!selectedCluster.value) return;
    const { data, error } = await fetchK8sClusterNamespaces(selectedCluster.value);
    if (!error) {
      namespaces.value = (data || []).map((ns: K8s.Namespace) => ns.name);
      // 自动选中第一个命名空间
      if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
        selectedNamespace.value = namespaces.value[0];
      }
    } else {
      console.error('加载命名空间失败:', error);
    }
  }

  // 加载配置项
  async function loadConfigMaps() {
    if (!selectedCluster.value) return;
    loading.value = true;
    const { data, error } = await fetchK8sConfigMaps(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: configmapsPagination.page,
      pageSize: configmapsPagination.pageSize
    });

    if (!error) {
      configmapsData.value = data?.list || [];
      configmapsPagination.itemCount = data?.total || 0;
    } else {
      console.error('加载 ConfigMaps 失败:', error);
    }
    loading.value = false;
  }

  // 加载保密字典
  async function loadSecrets() {
    if (!selectedCluster.value) return;
    loading.value = true;
    const { data, error } = await fetchK8sSecrets(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: secretsPagination.page,
      pageSize: secretsPagination.pageSize
    });

    if (!error) {
      secretsData.value = data?.list || [];
      secretsPagination.itemCount = data?.total || 0;
    } else {
      console.error('加载 Secrets 失败:', error);
    }
    loading.value = false;
  }

  // 加载当前Tab数据
  async function loadCurrentData() {
    if (activeTab.value === 'configmaps') {
      await loadConfigMaps();
    } else {
      await loadSecrets();
    }
  }

  // 分页变化
  function handlePageSizeChange(pageSize: number) {
    currentPagination.value.pageSize = pageSize;
    currentPagination.value.page = 1;
    loadCurrentData();
  }

  // Tab切换
  function handleTabChange(tabName: string) {
    activeTab.value = tabName;
    loadCurrentData();
  }

  // 监听集群和命名空间变化
  // 监听集群和命名空间变化
  watch([selectedCluster, selectedNamespace], () => {
    if (selectedCluster.value) {
      loadNamespaces();
      loadCurrentData();
    }
  });

  onMounted(async () => {
    await loadClusters();
    if (selectedCluster.value) {
      await loadNamespaces();
      await loadConfigMaps();
    }
  });
</script>

<template>
  <div class="config-page p-24px">
    <!-- 页面标题 -->
    <div class="mb-24px">
      <h1 class="text-28px text-primary font-bold">配置管理</h1>
      <p class="text-tertiary mt-8px text-14px">管理 Kubernetes 配置资源</p>
    </div>

    <!-- 头部：集群/命名空间选择和刷新按钮（同一行）-->
    <div class="mb-16px flex items-center justify-between gap-12px">
      <!-- 左侧：集群和命名空间选择 -->
      <div class="flex items-center gap-8px">
        <ElSelect v-model="selectedCluster" placeholder="选择集群" style="width: 200px" @change="loadNamespaces">
          <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
        </ElSelect>
        <ElSelect v-model="selectedNamespace" placeholder="选择命名空间" style="width: 180px" @change="loadCurrentData">
          <ElOption v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
        </ElSelect>
      </div>

      <!-- 右侧：刷新按钮 -->
      <ElButton text @click="loadCurrentData">
        <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
      </ElButton>
    </div>

    <!-- Tab 切换 -->
    <ElTabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 配置项 -->
      <ElTabPane label="配置项" name="configmaps">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="configmapsData"
            class="config-table"
            :header-cell-style="{ background: '#f5f7fa', color: '#303133', fontWeight: '600' }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 0' }"
          >
            <ElTableColumn prop="name" label="名称" min-width="200" align="left" />
            <ElTableColumn prop="namespace" label="命名空间" min-width="150" align="left" />
            <ElTableColumn label="数据键" min-width="300" align="left">
              <template #default="{ row }">
                <ElTag v-for="(key, index) in row.dataKeys" :key="index" size="small" class="mb-4px mr-4px">
                  {{ key }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="120" align="center" class-name="msre-table-actions">
              <template #default>
                <ElButton link type="primary" size="small">查看详情</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <ElPagination
              v-model:current-page="configmapsPagination.page"
              v-model:page-size="configmapsPagination.pageSize"
              :total="configmapsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadConfigMaps"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 保密字典 -->
      <ElTabPane label="保密字典" name="secrets">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="secretsData"
            class="config-table"
            :header-cell-style="{ background: '#f5f7fa', color: '#303133', fontWeight: '600' }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 0' }"
          >
            <ElTableColumn prop="name" label="名称" min-width="200" align="left" />
            <ElTableColumn prop="namespace" label="命名空间" min-width="150" align="left" />
            <ElTableColumn prop="type" label="类型" min-width="150" align="left">
              <template #default="{ row }">
                <ElTag>{{ row.type }}</ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="数据键" min-width="300" align="left">
              <template #default="{ row }">
                <ElTag
                  v-for="(_key, index) in row.dataKeys"
                  :key="index"
                  size="small"
                  class="mb-4px mr-4px"
                  type="warning"
                >
                  ***
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="120" align="center" class-name="msre-table-actions">
              <template #default>
                <ElButton link type="primary" size="small">查看详情</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <ElPagination
              v-model:current-page="secretsPagination.page"
              v-model:page-size="secretsPagination.pageSize"
              :total="secretsPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="loadSecrets"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<style scoped>
  .config-page {
    height: calc(100vh - 80px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Tab 容器自动填充剩余空间 */
  :deep(.el-tabs) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Tab 内容区域自动填充 */
  :deep(.el-tabs__content) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* 每个 Tab 页面填充空间 */
  :deep(.el-tab-pane) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    height: 100%;
  }

  /* 表格容器 */
  .table-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
    min-height: 0;
  }

  /* 表格样式 */
  .config-table {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding-bottom: 60px;
  }

  /* 底部工具栏固定在底部 */
  .bottom-toolbar {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 12px 16px;
    background-color: #fff;
    border-top: 1px solid #ebeef5;
    z-index: 10;

    :deep(.el-pagination) {
      margin: 0;
    }
  }
</style>
