<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElMessage,
  ElMessageBox,
  ElOption,
  ElSelect,
  ElSpace,
  ElTable,
  ElTableColumn,
  ElTag,
  type FormInstance,
  type FormRules
} from 'element-plus';
import {
  deleteK8sService,
  fetchK8sClusters,
  fetchK8sClusterNamespaces,
  fetchK8sServices,
  getK8sService
} from '@/service/api/k8s';

defineOptions({ name: 'K8sServices' });

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

// 服务类型显示
const getServiceTypeTag = (type: string) => {
  switch (type) {
    case 'ClusterIP':
      return { type: 'primary', text: 'ClusterIP' };
    case 'NodePort':
      return { type: 'success', text: 'NodePort' };
    case 'LoadBalancer':
      return { type: 'warning', text: 'LoadBalancer' };
    case 'ExternalName':
      return { type: 'info', text: 'ExternalName' };
    default:
      return { type: 'info', text: type || 'Unknown' };
  }
};

// 格式化端口信息
const formatPorts = (ports: any[]) => {
  if (!ports || ports.length === 0) return '-';

  return ports
    .map(p => {
      const parts = [];
      if (p.name) parts.push(p.name);
      parts.push(`${p.port}/${p.protocol}`);
      if (p.nodePort) parts.push(`NodePort: ${p.nodePort}`);
      return parts.join(' - ');
    })
    .join(', ');
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
      await loadServices();
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

// 加载 Service 列表
const loadServices = async () => {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    dataSource.value = [];
    pagination.itemCount = 0;
    return;
  }

  loading.value = true;
  try {
    const res = await fetchK8sServices(selectedCluster.value, {
      namespace: filters.namespace
    });

    // 处理不同的响应格式
    const services = Array.isArray(res) ? res : (res?.data || []);
    dataSource.value = services;
    pagination.itemCount = services?.length || 0;
  } catch (error: any) {
    dataSource.value = [];
    pagination.itemCount = 0;
    message.error(error.message || '加载 Service 列表失败');
  } finally {
    loading.value = false;
  }
};

// 集群变化
const handleClusterChange = async () => {
  await loadNamespaces();
  await loadServices();
};

// 命名空间变化
const handleNamespaceChange = () => {
  loadServices();
};

// 查看详情
const handleViewDetail = async (row: any) => {
  try {
    const res = await getK8sService(selectedCluster.value!, filters.namespace, row.name);
    currentDetail.value = res;
    showDetail.value = true;
  } catch (error: any) {
    message.error(error.message || '加载详情失败');
  }
};

// 删除 Service
const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Service "${row.name}" 吗？`, '确认删除', {
      type: 'warning'
    });

    await deleteK8sService(selectedCluster.value!, {
      namespace: filters.namespace,
      name: row.name
    });
    message.success('删除成功');
    loadServices();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
};

// 刷新
const handleRefresh = () => {
  loadServices();
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

    <!-- Service 列表 -->
    <el-table v-loading="loading" :data="dataSource" stripe>
      <el-table-column prop="name" label="名称" min-width="180" />
      <el-table-column prop="namespace" label="命名空间" width="150" />
      <el-table-column label="类型" width="140">
        <template #default="{ row }">
          <el-tag :type="getServiceTypeTag(row.type).type">
            {{ getServiceTypeTag(row.type).text }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="clusterIP" label="Cluster IP" width="150" />
      <el-table-column label="端口" min-width="250">
        <template #default="{ row }">
          <div class="text-sm text-gray-600 truncate" :title="formatPorts(row.ports)">
            {{ formatPorts(row.ports) }}
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
    <el-dialog v-model="showDetail" :title="`Service: ${currentDetail?.name}`" width="800px">
      <div v-if="currentDetail" class="space-y-4">
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
            <span class="text-sm font-medium text-gray-700">Cluster IP:</span>
            <span class="ml-2">{{ currentDetail.clusterIP || '-' }}</span>
          </div>
          <div>
            <span class="text-sm font-medium text-gray-700">外部 IP:</span>
            <span class="ml-2">{{ currentDetail.externalIP?.join(', ') || '-' }}</span>
          </div>
          <div>
            <span class="text-sm font-medium text-gray-700">年龄:</span>
            <span class="ml-2">{{ currentDetail.age }}</span>
          </div>
        </div>

        <!-- 端口信息 -->
        <div v-if="currentDetail.ports && currentDetail.ports.length > 0">
          <h4 class="text-sm font-medium text-gray-700 mb-2">端口:</h4>
          <el-table :data="currentDetail.ports" size="small">
            <el-table-column prop="name" label="名称" width="120" />
            <el-table-column prop="protocol" label="协议" width="80" />
            <el-table-column prop="port" label="端口" width="80" />
            <el-table-column prop="targetPort" label="目标端口" width="100" />
            <el-table-column prop="nodePort" label="NodePort" width="100" />
          </el-table>
        </div>

        <!-- Selector -->
        <div v-if="currentDetail.selector">
          <h4 class="text-sm font-medium text-gray-700 mb-2">选择器:</h4>
          <div class="flex flex-wrap gap-2">
            <el-tag
              v-for="(value, key) in currentDetail.selector"
              :key="key"
              type="info"
            >
              {{ key }}: {{ value }}
            </el-tag>
          </div>
        </div>

        <!-- Endpoints -->
        <div v-if="currentDetail.endpoints && currentDetail.endpoints.length > 0">
          <h4 class="text-sm font-medium text-gray-700 mb-2">后端端点:</h4>
          <el-table :data="currentDetail.endpoints" size="small">
            <el-table-column prop="ip" label="IP" width="140" />
            <el-table-column prop="hostname" label="主机名" width="180" />
            <el-table-column label="端口" min-width="200">
              <template #default="{ row }">
                <span v-if="row.ports">
                  {{ row.ports.map((p: any) => `${p.port}/${p.protocol}`).join(', ') }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
          </el-table>
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
