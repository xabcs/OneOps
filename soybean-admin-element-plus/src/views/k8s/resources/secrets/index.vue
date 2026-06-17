<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
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
  deleteK8sSecret,
  fetchK8sClusters,
  fetchK8sClusterNamespaces,
  fetchK8sSecrets,
  getK8sSecret
} from '@/service/api/k8s';

defineOptions({ name: 'K8sSecrets' });

const message = ElMessage;

const loading = ref(false);
const dataSource = ref<any[]>([]);
const showDetail = ref(false);
const currentDetail = ref<any>(null);

// 当前选中的集群和命名空间
const selectedCluster = ref<number | null>(null);
const selectedNamespace = ref('default');

// 可用的命名空间列表
const namespaces = ref<string[]>([]);

// 可用的集群列表
const clusters = ref<any[]>([]);

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
  } catch (error: any) {
    message.error(error.message || '加载集群列表失败');
  }
};

// 加载命名空间列表
const loadNamespaces = async () => {
  if (!selectedCluster.value) return;

  try {
    const namespaceList = await fetchK8sClusterNamespaces(selectedCluster.value);
    const namespaceNames = namespaceList.map((ns: any) => ns.name);
    namespaces.value = [...namespaceNames];

    if (namespaces.value.length > 0 && !namespaces.value.includes(filters.namespace)) {
      filters.namespace = namespaces.value[0];
      selectedNamespace.value = namespaces.value[0];
    }
  } catch (error: any) {
    message.error(error.message || '加载命名空间列表失败');
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
    const secrets = Array.isArray(res) ? res : (res?.data || []);
    dataSource.value = secrets;
    pagination.itemCount = secrets?.length || 0;
  } catch (error: any) {
    dataSource.value = [];
    pagination.itemCount = 0;
    message.error(error.message || '加载 Secret 列表失败');
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

// 查看详情
const handleViewDetail = async (row: any) => {
  try {
    const res = await getK8sSecret(selectedCluster.value!, filters.namespace, row.name);
    currentDetail.value = res;
    showDetail.value = true;
  } catch (error: any) {
    message.error(error.message || '加载详情失败');
  }
};

// 删除 Secret
const handleDelete = async (row: any) => {
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
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
};

// 刷新
const handleRefresh = () => {
  loadSecrets();
};

onMounted(() => {
  loadClusters();
});
</script>

<template>
  <div class="p-4">
    <!-- 筛选栏 -->
    <div class="mb-4 flex items-center gap-4 bg-card p-4 rounded-lg">
      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">集群:</span>
        <el-select
          v-model="selectedCluster"
          placeholder="请选择集群"
          style="width: 200px"
          @change="handleClusterChange"
        >
          <el-option
            v-for="cluster in clusters"
            :key="cluster.id"
            :label="cluster.name"
            :value="cluster.id"
          />
        </el-select>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-sm font-medium">命名空间:</span>
        <el-select
          v-model="filters.namespace"
          placeholder="请选择命名空间"
          style="width: 180px"
          @change="handleNamespaceChange"
        >
          <el-option
            v-for="ns in namespaces"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>
      </div>

      <div class="flex-1" />

      <el-button type="primary" :disabled="!selectedCluster" @click="handleRefresh">
        刷新
      </el-button>
    </div>

    <!-- Secret 列表 -->
    <el-table v-loading="loading" :data="dataSource" stripe>
      <el-table-column prop="name" label="名称" min-width="220" />
      <el-table-column prop="namespace" label="命名空间" width="150" />
      <el-table-column label="类型" width="200">
        <template #default="{ row }">
          <el-tag :type="getSecretTypeTag(row.type).type">
            {{ getSecretTypeTag(row.type).text }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="数据键" min-width="250">
        <template #default="{ row }">
          <div class="flex flex-wrap gap-1">
            <el-tag
              v-for="key in row.dataKeys"
              :key="key"
              size="small"
              type="warning"
            >
              {{ key }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="age" label="年龄" width="160" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-space>
            <el-button size="small" @click="handleViewDetail(row)">
              详情
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </el-space>
        </template>
      </el-table-column>
    </el-table>

    <!-- 详情弹窗 -->
    <el-dialog v-model="showDetail" :title="`Secret: ${currentDetail?.name}`" width="800px">
      <div v-if="currentDetail" class="space-y-4">
        <!-- 安全警告 -->
        <div class="bg-amber-50 border border-amber-200 rounded p-3 flex items-start gap-2">
          <svg class="w-5 h-5 text-amber-500 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div class="text-sm text-amber-800">
            <strong>敏感数据保护：</strong>为保护敏感信息，Secret 的实际数据值已被隐藏（显示为 ******）。如需查看或编辑，请直接使用 kubectl 命令。
          </div>
        </div>

        <!-- 基本信息 -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <span class="text-sm font-medium text-gray-700">名称:</span>
            <span class="ml-2">{{ currentDetail.name }}</span>
          </div>
          <div>
            <span class="text-sm font-medium text-gray-700">命名空间:</span>
            <span class="ml-2">{{ currentDetail.namespace }}</span>
          </div>
          <div>
            <span class="text-sm font-medium text-gray-700">类型:</span>
            <span class="ml-2">{{ currentDetail.type }}</span>
          </div>
          <div>
            <span class="text-sm font-medium text-gray-700">年龄:</span>
            <span class="ml-2">{{ currentDetail.age }}</span>
          </div>
        </div>

        <!-- Labels -->
        <div v-if="currentDetail.labels && Object.keys(currentDetail.labels).length > 0">
          <h4 class="text-sm font-medium text-gray-700 mb-2">标签:</h4>
          <div class="flex flex-wrap gap-2">
            <el-tag
              v-for="(value, key) in currentDetail.labels"
              :key="key"
              type="info"
            >
              {{ key }}: {{ value }}
            </el-tag>
          </div>
        </div>

        <!-- Data Keys (隐藏实际值) -->
        <div v-if="currentDetail.data && Object.keys(currentDetail.data).length > 0">
          <h4 class="text-sm font-medium text-gray-700 mb-2">数据键:</h4>
          <div class="space-y-2">
            <div
              v-for="(value, key) in currentDetail.data"
              :key="key"
              class="border rounded p-2"
            >
              <div class="text-sm font-medium text-gray-700 mb-1">{{ key }}</div>
              <div class="text-sm bg-gray-50 p-2 rounded font-mono text-amber-600">
                ****** (敏感数据已隐藏)
              </div>
            </div>
          </div>
        </div>

        <!-- YAML Manifest -->
        <div>
          <h4 class="text-sm font-medium text-gray-700 mb-2">YAML 配置:</h4>
          <div class="bg-gray-50 p-3 rounded text-sm font-mono overflow-auto max-h-400 whitespace-pre">
            {{ currentDetail.manifest }}
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>
