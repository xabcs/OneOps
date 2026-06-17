<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import {
  ElButton,
  ElMessage,
  ElPagination,
  ElSpace,
  ElTag,
  ElTabs,
  ElTabPane,
  type FormInstance
} from 'element-plus';
import {
  deleteK8sDeployment,
  fetchK8sClusters,
  fetchK8sDeployments,
  fetchK8sClusterNamespaces,
  restartK8sDeployment,
  scaleK8sDeployment,
  deleteK8sPod,
  fetchK8sPods,
  fetchK8sServices,
  fetchK8sConfigMaps,
  fetchK8sSecrets
} from '@/service/api/k8s';

defineOptions({ name: 'K8sWorkloads' });

const router = useRouter();
const message = ElMessage;

const loading = ref(false);
const activeTab = ref('deployments');

// 当前选中的集群和命名空间
const selectedCluster = ref<number | null>(null);
const selectedNamespace = ref('default');

// 可用的命名空间列表
const namespaces = ref<string[]>([]);

// 可用的集群列表
const clusters = ref<any[]>([]);

// 各种资源的数据
const deploymentsData = ref<any[]>([]);
const podsData = ref<any[]>([]);
const statefulSetsData = ref<any[]>([]);
const daemonSetsData = ref<any[]>([]);
const jobsData = ref<any[]>([]);
const cronJobsData = ref<any[]>([]);

// 各种资源的分页
const deploymentsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const podsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const statefulSetsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const daemonSetsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const jobsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const cronJobsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

// 根据Tab返回对应的数据和分页
const currentData = computed(() => {
  switch (activeTab.value) {
    case 'pods':
      return podsData.value;
    case 'deployments':
      return deploymentsData.value;
    case 'statefulsets':
      return statefulSetsData.value;
    case 'daemonsets':
      return daemonSetsData.value;
    case 'jobs':
      return jobsData.value;
    case 'cronjobs':
      return cronJobsData.value;
    default:
      return [];
  }
});

const currentPagination = computed(() => {
  switch (activeTab.value) {
    case 'pods':
      return podsPagination;
    case 'deployments':
      return deploymentsPagination;
    case 'statefulsets':
      return statefulSetsPagination;
    case 'daemonsets':
      return daemonSetsPagination;
    case 'jobs':
      return jobsPagination;
    case 'cronjobs':
      return cronJobsPagination;
    default:
      return { page: 1, pageSize: 10, itemCount: 0 };
  }
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

// 加载各种资源数据
async function loadDeployments() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    console.log('[DEBUG] 加载 Deployments:', {
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      page: deploymentsPagination.page,
      pageSize: deploymentsPagination.pageSize
    });

    const response = await fetchK8sDeployments(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: deploymentsPagination.page,
      pageSize: deploymentsPagination.pageSize
    });

    // 处理不同的响应格式
    // 格式1: {list: [...], total: number} (期望格式)
    // 格式2: {data: {list: [...], total: number}}
    // 格式3: {data: {data: {list: [...], total: number}}}
    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    deploymentsData.value = apiData.list || [];
    deploymentsPagination.itemCount = apiData.total || 0;

    console.log('[DEBUG] Deployments 数据赋值后:', {
      dataLength: deploymentsData.value.length,
      total: deploymentsPagination.itemCount,
      firstItem: deploymentsData.value[0]
    });
  } catch (error) {
    console.error('加载 Deployments 失败:', error);
  } finally {
    loading.value = false;
  }
}

async function loadPods() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    console.log('[DEBUG] 加载 Pods:', {
      clusterId: selectedCluster.value,
      namespace: selectedNamespace.value,
      page: podsPagination.page,
      pageSize: podsPagination.pageSize
    });

    const response = await fetchK8sPods(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: podsPagination.page,
      pageSize: podsPagination.pageSize
    });

    // 处理不同的响应格式
    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    podsData.value = apiData.list || [];
    podsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 Pods 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 StatefulSets（占位，需要后端API支持）
async function loadStatefulSets() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    // TODO: 实现StatefulSets API调用
    statefulSetsData.value = [];
    statefulSetsPagination.itemCount = 0;
  } catch (error) {
    console.error('加载 StatefulSets 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 DaemonSets（占位，需要后端API支持）
async function loadDaemonSets() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    // TODO: 实现DaemonSets API调用
    daemonSetsData.value = [];
    daemonSetsPagination.itemCount = 0;
  } catch (error) {
    console.error('加载 DaemonSets 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 Jobs（占位，需要后端API支持）
async function loadJobs() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    // TODO: 实现Jobs API调用
    jobsData.value = [];
    jobsPagination.itemCount = 0;
  } catch (error) {
    console.error('加载 Jobs 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载 CronJobs（占位，需要后端API支持）
async function loadCronJobs() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    // TODO: 实现CronJobs API调用
    cronJobsData.value = [];
    cronJobsPagination.itemCount = 0;
  } catch (error) {
    console.error('加载 CronJobs 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 根据当前Tab加载对应数据
async function loadCurrentData() {
  switch (activeTab.value) {
    case 'pods':
      await loadPods();
      break;
    case 'deployments':
      await loadDeployments();
      break;
    case 'statefulsets':
      await loadStatefulSets();
      break;
    case 'daemonsets':
      await loadDaemonSets();
      break;
    case 'jobs':
      await loadJobs();
      break;
    case 'cronjobs':
      await loadCronJobs();
      break;
  }
}

// 跳转到详情页
function goToDetail(row: any) {
  switch (activeTab.value) {
    case 'deployments':
      router.push({
        path: '/k8s/resources/deployments/detail',
        query: { clusterId: selectedCluster.value, namespace: row.namespace, name: row.name }
      });
      break;
    // TODO: 添加其他资源类型的详情页路由
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
    await loadDeployments();
  }
});
</script>

<template>
  <div class="workloads-page p-24px bg-layout">
    <!-- 页面标题 -->
    <div class="mb-24px">
      <h1 class="text-28px font-bold text-primary">工作负载</h1>
      <p class="text-14px text-tertiary mt-8px">管理 Kubernetes 工作负载资源</p>
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

    <!-- Tab 切换不同资源类型 -->
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 无状态 -->
      <el-tab-pane label="无状态" name="deployments">
        <div v-loading="loading" class="tab-content">
          <el-table :data="deploymentsData" stripe>
            <el-table-column prop="name" label="名称" width="200" />
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column prop="replicas" label="副本数" width="100">
              <template #default="{ row }">
                {{ row.ready }} / {{ row.replicas }}
              </template>
            </el-table-column>
            <el-table-column prop="upToDate" label="最新" width="80" />
            <el-table-column prop="available" label="可用" width="80" />
            <el-table-column prop="age" label="年龄" width="120" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.ready === row.replicas" type="success">运行中</el-tag>
                <el-tag v-else type="warning">更新中</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="goToDetail(row)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="deploymentsPagination.page"
            v-model:page-size="deploymentsPagination.pageSize"
            :total="deploymentsPagination.itemCount"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="mt-16px"
            @current-change="loadDeployments"
            @size-change="handlePageSizeChange"
          />
        </div>
      </el-tab-pane>

      <!-- 容器组 -->
      <el-tab-pane label="容器组" name="pods">
        <div v-loading="loading" class="tab-content">
          <el-table :data="podsData" stripe>
            <el-table-column prop="name" label="名称" width="200" />
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'Running' ? 'success' : 'warning'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ip" label="IP地址" width="140" />
            <el-table-column prop="node" label="节点" width="150" />
            <el-table-column prop="restarts" label="重启次数" width="100" />
            <el-table-column prop="age" label="年龄" width="120" />
          </el-table>

          <el-pagination
            v-model:current-page="podsPagination.page"
            v-model:page-size="podsPagination.pageSize"
            :total="podsPagination.itemCount"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="mt-16px"
            @current-change="loadPods"
            @size-change="handlePageSizeChange"
          />
        </div>
      </el-tab-pane>

      <!-- 有状态 -->
      <el-tab-pane label="有状态" name="statefulsets">
        <div v-loading="loading" class="tab-content">
          <el-empty description="StatefulSets 功能即将推出" />
        </div>
      </el-tab-pane>

      <!-- 守护进程集 -->
      <el-tab-pane label="守护进程集" name="daemonsets">
        <div v-loading="loading" class="tab-content">
          <el-empty description="DaemonSets 功能即将推出" />
        </div>
      </el-tab-pane>

      <!-- 任务 -->
      <el-tab-pane label="任务" name="jobs">
        <div v-loading="loading" class="tab-content">
          <el-empty description="Jobs 功能即将推出" />
        </div>
      </el-tab-pane>

      <!-- 定时任务 -->
      <el-tab-pane label="定时任务" name="cronjobs">
        <div v-loading="loading" class="tab-content">
          <el-empty description="CronJobs 功能即将推出" />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.workloads-page {
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
