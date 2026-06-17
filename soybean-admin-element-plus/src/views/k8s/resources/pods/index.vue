<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  ElOption,
  ElPagination,
  ElSelect,
  ElSpace,
  ElTable,
  ElTableColumn,
  ElTag,
  type FormInstance
} from 'element-plus';
import {
  deleteK8sPod,
  fetchK8sClusters,
  fetchK8sClusterNamespaces,
  fetchK8sPods,
  fetchK8sPodLogs
} from '@/service/api/k8s';
import PodTerminal from '../../terminal/PodTerminal.vue';

defineOptions({ name: 'K8sPods' });

const message = ElMessage;

const loading = ref(false);
const dataSource = ref<any[]>([]);
const showTerminal = ref(false);
const showLogs = ref(false);
const logContent = ref('');
const submitting = ref(false);

// 当前选中的集群和命名空间
const selectedCluster = ref<number | null>(null);
const selectedNamespace = ref('default');

// 可用的命名空间列表
const namespaces = ref<string[]>([]);

// 可用的集群列表
const clusters = ref<any[]>([]);

const filters = reactive({
  namespace: 'default',
  labelSelector: ''
});

// 终端相关
const terminalProps = ref({
  clusterId: 0,
  namespace: '',
  podName: '',
  containerName: ''
});

// 日志相关
const logPodName = ref('');
const logContainerName = ref('');
const logTailLines = ref(100);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0
});

// 状态显示
const getStatusTag = (pod: any) => {
  const phase = pod.phase || 'Unknown';
  const status = pod.status || phase;

  switch (status) {
    case 'Running':
      return { type: 'success', text: '运行中' };
    case 'Succeeded':
      return { type: 'info', text: '已完成' };
    case 'Failed':
      return { type: 'danger', text: '失败' };
    case 'Pending':
      return { type: 'warning', text: '等待中' };
    case 'Unknown':
    default:
      return { type: 'info', text: '未知' };
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
      await loadPods();
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

// 加载 Pod 列表
const loadPods = async () => {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    dataSource.value = [];
    pagination.itemCount = 0;
    return;
  }

  loading.value = true;
  try {
    const res = await fetchK8sPods(selectedCluster.value, {
      namespace: filters.namespace,
      labelSelector: filters.labelSelector
    });

    // 处理不同的响应格式
    const pods = Array.isArray(res) ? res : (res?.data || []);
    dataSource.value = pods;
    pagination.itemCount = pods?.length || 0;
  } catch (error: any) {
    dataSource.value = [];
    pagination.itemCount = 0;
    message.error(error.message || '加载 Pod 列表失败');
  } finally {
    loading.value = false;
  }
};

// 集群变化
const handleClusterChange = async () => {
  await loadNamespaces();
  await loadPods();
};

// 命名空间变化
const handleNamespaceChange = () => {
  loadPods();
};

// 搜索
const handleSearch = () => {
  pagination.page = 1;
  loadPods();
};

// 进入终端
const handleTerminal = (row: any, containerName?: string) => {
  terminalProps.value = {
    clusterId: selectedCluster.value!,
    namespace: filters.namespace,
    podName: row.name,
    containerName: containerName || ''
  };
  showTerminal.value = true;
};

// 查看日志
const handleLogs = async (row: any, containerName?: string) => {
  logPodName.value = row.name;
  logContainerName.value = containerName || '';
  logContent.value = '加载中...';
  showLogs.value = true;

  try {
    const res = await fetchK8sPodLogs(
      selectedCluster.value!,
      filters.namespace,
      row.name,
      {
        container: containerName || '',
        tailLines: logTailLines.value
      }
    );
    logContent.value = res.logs || '暂无日志';
  } catch (error: any) {
    logContent.value = '日志加载失败: ' + (error.message || '未知错误');
  }
};

// 删除 Pod
const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Pod "${row.name}" 吗？`, '确认删除', {
      type: 'warning'
    });

    await deleteK8sPod(selectedCluster.value!, {
      namespace: filters.namespace,
      name: row.name
    });
    message.success('删除成功');
    loadPods();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
};

// 刷新
const handleRefresh = () => {
  loadPods();
};

// 关闭日志弹窗
const closeLogs = () => {
  showLogs.value = false;
  logContent.value = '';
};

// 关闭终端
const closeTerminal = () => {
  showTerminal.value = false;
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

      <div class="flex items-center gap-2">
        <el-input
          v-model="filters.labelSelector"
          placeholder="标签选择器 (可选)"
          style="width: 200px"
          clearable
          @change="handleSearch"
        />
      </div>

      <div class="flex-1" />

      <el-button type="primary" :disabled="!selectedCluster" @click="handleRefresh">
        刷新
      </el-button>
    </div>

    <!-- Pod 列表 -->
    <el-table v-loading="loading" :data="dataSource" stripe>
      <el-table-column prop="name" label="名称" min-width="200" />
      <el-table-column prop="namespace" label="命名空间" width="150" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="getStatusTag(row).type">
            {{ getStatusTag(row).text }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="IP 地址" width="140" />
      <el-table-column prop="node" label="节点" width="180" />
      <el-table-column label="重启次数" width="100" align="center">
        <template #default="{ row }">
          {{ row.restarts || 0 }}
        </template>
      </el-table-column>
      <el-table-column prop="age" label="年龄" width="160" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-space wrap>
            <el-button size="small" type="primary" @click="handleTerminal(row)">
              进入终端
            </el-button>
            <el-button size="small" @click="handleLogs(row)">
              查看日志
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </el-space>
        </template>
      </el-table-column>
    </el-table>

    <!-- 日志弹窗 -->
    <el-dialog v-model="showLogs" :title="`日志: ${logPodName}`" width="900px" top="5vh">
      <div class="flex items-center gap-4 mb-4">
        <span class="text-sm text-gray-600">行数:</span>
        <el-input-number v-model="logTailLines" :min="10" :max="10000" :step="100" />
        <el-button size="small" @click="handleLogs({ name: logPodName }, logContainerName)">
          刷新
        </el-button>
      </div>
      <div class="bg-black text-green-400 p-4 rounded font-mono text-sm overflow-auto max-h-[600px] whitespace-pre-wrap">{{ logContent }}</div>
    </el-dialog>

    <!-- 终端弹窗 -->
    <el-dialog v-model="showTerminal" title="Pod 终端" width="900px" fullscreen @close="closeTerminal">
      <PodTerminal
        v-if="showTerminal"
        :cluster-id="terminalProps.clusterId"
        :namespace="terminalProps.namespace"
        :pod-name="terminalProps.podName"
        :container-name="terminalProps.containerName"
      />
    </el-dialog>
  </div>
</template>
