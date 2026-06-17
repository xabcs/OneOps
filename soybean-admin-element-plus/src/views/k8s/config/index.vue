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
  fetchK8sConfigMaps,
  fetchK8sSecrets
} from '@/service/api/k8s';

defineOptions({ name: 'K8sConfig' });

const message = ElMessage;

const loading = ref(false);
const activeTab = ref('configmaps');

// 当前选中的集群和命名空间
const selectedCluster = ref<number | null>(null);
const selectedNamespace = ref('default');

// 可用的命名空间列表
const namespaces = ref<string[]>([]);

// 可用的集群列表
const clusters = ref<any[]>([]);

// 资源数据
const configmapsData = ref<any[]>([]);
const secretsData = ref<any[]>([]);

// 分页
const configmapsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
const secretsPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

// 当前数据
const currentData = computed(() => {
  return activeTab.value === 'configmaps' ? configmapsData.value : secretsData.value;
});

const currentPagination = computed(() => {
  return activeTab.value === 'configmaps' ? configmapsPagination : secretsPagination;
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

// 加载配置项
async function loadConfigMaps() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    const response = await fetchK8sConfigMaps(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: configmapsPagination.page,
      pageSize: configmapsPagination.pageSize
    });

    // 处理不同的响应格式
    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    configmapsData.value = apiData.list || [];
    configmapsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 ConfigMaps 失败:', error);
  } finally {
    loading.value = false;
  }
}

// 加载保密字典
async function loadSecrets() {
  if (!selectedCluster.value) return;
  loading.value = true;
  try {
    const response = await fetchK8sSecrets(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: secretsPagination.page,
      pageSize: secretsPagination.pageSize
    });

    // 处理不同的响应格式
    let apiData = response;
    if (response?.data?.data?.list) {
      apiData = response.data.data;
    } else if (response?.data?.list) {
      apiData = response.data;
    }

    secretsData.value = apiData.list || [];
    secretsPagination.itemCount = apiData.total || 0;
  } catch (error) {
    console.error('加载 Secrets 失败:', error);
  } finally {
    loading.value = false;
  }
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
    await loadConfigMaps();
  }
});
</script>

<template>
  <div class="config-page p-24px bg-layout">
    <!-- 页面标题 -->
    <div class="mb-24px">
      <h1 class="text-28px font-bold text-primary">配置管理</h1>
      <p class="text-14px text-tertiary mt-8px">管理 Kubernetes 配置资源</p>
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
      <!-- 配置项 -->
      <el-tab-pane label="配置项" name="configmaps">
        <div v-loading="loading" class="tab-content">
          <el-table :data="configmapsData" stripe>
            <el-table-column prop="name" label="名称" width="200" />
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column label="数据键" width="300">
              <template #default="{ row }">
                <el-tag
                  v-for="(key, index) in row.dataKeys"
                  :key="index"
                  size="small"
                  class="mr-4px mb-4px"
                >
                  {{ key }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="age" label="年龄" width="120" />
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button link type="primary" size="small">
                  查看详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="configmapsPagination.page"
            v-model:page-size="configmapsPagination.pageSize"
            :total="configmapsPagination.itemCount"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="mt-16px"
            @current-change="loadConfigMaps"
            @size-change="handlePageSizeChange"
          />
        </div>
      </el-tab-pane>

      <!-- 保密字典 -->
      <el-tab-pane label="保密字典" name="secrets">
        <div v-loading="loading" class="tab-content">
          <el-table :data="secretsData" stripe>
            <el-table-column prop="name" label="名称" width="200" />
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column prop="type" label="类型" width="150">
              <template #default="{ row }">
                <el-tag>{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="数据键" width="300">
              <template #default="{ row }">
                <el-tag
                  v-for="(key, index) in row.dataKeys"
                  :key="index"
                  size="small"
                  class="mr-4px mb-4px"
                  type="warning"
                >
                  ***
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="age" label="年龄" width="120" />
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button link type="primary" size="small">
                  查看详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="secretsPagination.page"
            v-model:page-size="secretsPagination.pageSize"
            :total="secretsPagination.itemCount"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="mt-16px"
            @current-change="loadSecrets"
            @size-change="handlePageSizeChange"
          />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.config-page {
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
