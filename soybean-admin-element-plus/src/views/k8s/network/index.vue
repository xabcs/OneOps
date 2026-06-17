<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import {
  ElButton,
  ElMessage,
  ElPagination,
  ElSpace,
  ElTag,
  ElTabs,
  ElTabPane
} from 'element-plus';
import {
  fetchK8sClusters,
  fetchK8sClusterNamespaces,
  fetchK8sServices
} from '@/service/api/k8s';

defineOptions({ name: 'K8sNetwork' });

const message = ElMessage;

const loading = ref(false);
const activeTab = ref('services');

// 当前选中的集群和命名空间
const selectedCluster = ref<number | null>(null);
const selectedNamespace = ref('default');

// 可用的命名空间列表
const namespaces = ref<string[]>([]);

// 可用的集群列表
const clusters = ref<any[]>([]);

// 资源数据
const servicesData = ref<any[]>([]);
const ingressesData = ref<any[]>([]);

// 分页
const servicesPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const ingressesPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

// 当前数据
const currentData = computed(() => {
  return activeTab.value === 'services' ? servicesData.value : ingressesData.value;
});

const currentPagination = computed(() => {
  return activeTab.value === 'services' ? servicesPagination : ingressesPagination;
});

// 加载集群列表
async function loadClusters() {
  try {
    const response = await fetchK8sClusters();
    clusters.value = response.data || [];
    // 自动选中第一个集群
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
    namespaces.value = response.data.map((ns: any) => ns.name);
    // 自动选中第一个命名空间
    if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
      selectedNamespace.value = namespaces.value[0];
    }
  } catch (error) {
    console.error('加载命名空间失败:', error);
  }
}

// 加载服务
async function loadServices() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    const response = await fetchK8sServices(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: servicesPagination.page,
      pageSize: servicesPagination.pageSize
    });

    // 处理不同的响应格式
    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    servicesData.value = apiData.list || [];
    servicesPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 Services 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载路由（占位）
async function loadIngresses() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    // TODO: 实现Ingress API调用
    ingressesData.value = [];
    ingressesPagination.itemCount = 0;
  } catch (error) {
    console.error('加载 Ingress 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载当前Tab数据
async function loadCurrentData() {
  if (activeTab.value === 'services') {
    await loadServices();
  } else {
    await loadIngresses();
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
    await loadServices();
  }
});
</script>

<template>
  <div class="network-page p-24px bg-layout">
    <!-- 页面标题 -->
    <div class="mb-24px">
      <h1 class="text-28px font-bold text-primary">网络</h1>
      <p class="text-14px text-tertiary mt-8px">管理 Kubernetes 网络资源</p>
    </div>

    <!-- 头部：集群/命名空间选择和刷新按钮（同一行）-->
    <div class="mb-16px flex items-center justify-between gap-12px">
      <!-- 左侧：集群和命名空间选择 -->
      <div class="filter-inputs flex items-center gap-8px">
        <el-select
          v-model="selectedCluster"
          style="width: 200px"
          @change="loadNamespaces"
        >
          <template #prefix>
            <span class="select-fixed-label">选择集群</span>
          </template>
          <el-option
            v-for="cluster in clusters"
            :key="cluster.id"
            :label="cluster.name"
            :value="cluster.id"
          />
        </el-select>
        <el-select
          v-model="selectedNamespace"
          style="width: 180px"
          @change="loadCurrentData"
        >
          <template #prefix>
            <span class="select-fixed-label">选择命名空间</span>
          </template>
          <el-option
            v-for="ns in namespaces"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>
      </div>

      <!-- 右侧：刷新按钮 -->
      <el-button text @click="loadCurrentData">
        <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
      </el-button>
    </div>

    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 服务 -->
      <el-tab-pane label="服务" name="services">
        <div v-loading="loading" class="tab-content">
          <el-table :data="servicesData" stripe>
            <el-table-column prop="name" label="名称" width="200" />
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column prop="type" label="类型" width="120">
              <template #default="{ row }">
                <el-tag>{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="clusterIP" label="集群IP" width="140" />
            <el-table-column label="外部IP" width="150">
              <template #default="{ row }">
                {{ row.externalIP?.join(', ') || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="端口" width="200">
              <template #default="{ row }">
                <el-tag
                  v-for="(port, index) in row.ports"
                  :key="index"
                  size="small"
                  class="mr-4px"
                >
                  {{ port.port }}/{{ port.protocol }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="age" label="年龄" width="120" />
          </el-table>

          <el-pagination
            v-model:current-page="servicesPagination.page"
            v-model:page-size="servicesPagination.pageSize"
            :total="servicesPagination.itemCount"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="mt-16px"
            @current-change="loadServices"
            @size-change="handlePageSizeChange"
          />
        </div>
      </el-tab-pane>

      <!-- 路由 -->
      <el-tab-pane label="路由" name="ingresses">
        <div v-loading="loading" class="tab-content">
          <el-empty description="Ingress 功能即将推出" />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.network-page {
  min-height: 100vh;
}

/* 筛选输入框样式 - 参考主机资产页面 */
.filter-inputs {
  :deep(.el-select__wrapper) {
    border-radius: 0 !important;
    height: 30px;
    font-size: 12px;
    line-height: 30px;
  }

  :deep(.el-input__wrapper) {
    border-radius: 0 !important;
    height: 30px;
    font-size: 12px;
  }

  :deep(.el-select) {
    height: 30px;
    font-size: 12px;
  }

  /* 隐藏选中项的显示，因为我们要显示固定的标签文案 */
  :deep(.el-select .el-select__selection) {
    display: none;
  }

  :deep(.el-select .el-select__selected-item) {
    display: none;
  }

  /* placeholder 样式 - 隐藏，因为我们用 prefix 替代 */
  :deep(.el-select .el-select__placeholder) {
    display: none;
  }

  /* 固定标签文案样式 */
  .select-fixed-label {
    font-size: 12px;
    color: var(--el-text-color-regular);
    line-height: 30px;
    padding-left: 8px;
  }

  /* 当有选中值时，调整 prefix 的位置 */
  :deep(.el-select.has-value .el-select__prefix) {
    position: static;
    flex: none;
  }
}

.tab-content {
  min-height: 400px;
}
</style>
