<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
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
  type FormInstance,
  type FormRules
} from 'element-plus';
import {
  deleteK8sDeployment,
  fetchK8sClusters,
  fetchK8sDeployments,
  fetchK8sClusterNamespaces,
  restartK8sDeployment,
  scaleK8sDeployment
} from '@/service/api/k8s';

defineOptions({ name: 'K8sDeployments' });

const router = useRouter();
const message = ElMessage;

const loading = ref(false);
const dataSource = ref<any[]>([]);
const showModal = ref(false);
const showScaleModal = ref(false);
const submitting = ref(false);
const formRef = ref<FormInstance | null>(null);
const scaleFormRef = ref<FormInstance | null>(null);

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

const formData = reactive({
  name: '',
  namespace: 'default',
  replicas: 1,
  image: '',
  command: '',
  args: ''
});

const scaleData = reactive({
  replicas: 1
});

const currentDeployment = ref<any>(null);
const currentDeploymentName = ref('');

const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0
});

const formRules: FormRules = {
  name: { required: true, message: '请输入 Deployment 名称', trigger: 'blur' },
  namespace: { required: true, message: '请选择命名空间', trigger: 'change' },
  replicas: { required: true, message: '请输入副本数', trigger: 'blur' },
  image: { required: true, message: '请输入镜像地址', trigger: 'blur' }
};

const scaleRules: FormRules = {
  replicas: [
    { required: true, message: '请输入副本数', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value < 0 || value > 100) {
          callback(new Error('副本数必须在 0-100 之间'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ]
};

// 状态显示
const getStatusTag = (deployment: any) => {
  const ready = deployment.ready || 0;
  const total = deployment.replicas || 0;

  if (total === 0) {
    return { type: 'info', text: '未就绪' };
  }

  if (ready === total) {
    return { type: 'success', text: `运行中 (${ready}/${total})` };
  } else if (ready > 0) {
    return { type: 'warning', text: `部分就绪 (${ready}/${total})` };
  } else {
    return { type: 'danger', text: '未就绪' };
  }
};

// 加载集群列表
const loadClusters = async () => {
  try {
    const response = await fetchK8sClusters();
    // 从响应中提取集群数组
    const clusterList = response?.data || [];
    clusters.value = clusterList;

    // 如果有集群，默认选择第一个
    if (clusters.value.length > 0 && !selectedCluster.value) {
      selectedCluster.value = clusters.value[0].id;
      await loadNamespaces();
      await loadDeployments();
    }
  } catch (error: any) {
    message.error(error.message || '加载集群列表失败');
  }
};

// 加载命名空间列表
const loadNamespaces = async () => {
  if (!selectedCluster.value) return;

  try {
    console.log('[调试] 开始加载命名空间, 集群ID:', selectedCluster.value);
    const response = await fetchK8sClusterNamespaces(selectedCluster.value);
    console.log('[调试] 命名空间完整响应:', response);
    console.log('[调试] 响应类型:', typeof response);
    console.log('[调试] 响应是否为数组:', Array.isArray(response));
    console.log('[调试] 响应.data是否为数组:', response?.data && Array.isArray(response.data));

    // 尝试多种方式提取数组
    let namespaceList: any[] = [];

    if (Array.isArray(response)) {
      console.log('[调试] 响应本身是数组');
      namespaceList = response;
    } else if (response?.data && Array.isArray(response.data)) {
      console.log('[调试] 响应.data是数组');
      namespaceList = response.data;
    } else {
      console.error('[调试] 无法从响应中提取数组');
      namespaceList = [];
    }

    console.log('[调试] 最终命名空间列表:', namespaceList);
    console.log('[调试] 命名空间数量:', namespaceList.length);

    // 使用普通数组而不是Vue Proxy对象
    const namespaceNames = namespaceList.map((ns: any) => ns.name);
    namespaces.value = [...namespaceNames]; // 创建新数组避免Proxy问题
    console.log('[调试] 命名空间名称数组:', namespaces.value);

    if (namespaces.value.length > 0 && !namespaces.value.includes(filters.namespace)) {
      filters.namespace = namespaces.value[0];
      selectedNamespace.value = namespaces.value[0];
    }
  } catch (error: any) {
    console.error('[调试] 加载命名空间失败:', error);
    message.error(error.message || '加载命名空间列表失败');
  }
};

// 加载 Deployment 列表
const loadDeployments = async () => {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    dataSource.value = [];
    pagination.itemCount = 0;
    return;
  }

  loading.value = true;
  try {
    console.log('[调试] 当前命名空间:', filters.namespace);

    const response = await fetchK8sDeployments(selectedCluster.value, {
      namespace: filters.namespace,
      page: pagination.page,
      pageSize: pagination.pageSize
    });

    console.log('[调试] API响应完整结构:', response);
    console.log('[调试] API响应类型:', typeof response);
    console.log('[调试] response.data:', response.data);
    console.log('[调试] response.data?.list:', response.data?.list);
    console.log('[调试] response.data?.total:', response.data?.total);

    // 尝试从不同位置提取数据
    const list = response.data?.list || response.list || [];
    const total = response.data?.total || response.total || 0;

    console.log('[调试] 提取的list:', list);
    console.log('[调试] 提取的total:', total);

    dataSource.value = list;
    pagination.itemCount = total;

    if (total === 0) {
      console.warn(`[调试] 命名空间 "${filters.namespace}" 下没有 deployments`);
    }
  } catch (error: any) {
    console.error('[调试] 请求失败:', error);
    dataSource.value = [];
    pagination.itemCount = 0;
    message.error(error.message || '加载 Deployment 列表失败');
  } finally {
    loading.value = false;
  }
};

// 集群变化
const handleClusterChange = async () => {
  await loadNamespaces();
  await loadDeployments();
};

// 命名空间变化
const handleNamespaceChange = () => {
  loadDeployments();
};

// 创建 Deployment
const handleCreate = () => {
  if (!selectedCluster.value) {
    message.warning('请先选择集群');
    return;
  }

  Object.assign(formData, {
    name: '',
    namespace: filters.namespace,
    replicas: 1,
    image: '',
    command: '',
    args: ''
  });
  showModal.value = true;
};

// 缩放 Deployment
const handleScale = (row: any) => {
  currentDeployment.value = row;
  currentDeploymentName.value = row.name;
  scaleData.replicas = row.replicas || 1;
  showScaleModal.value = true;
};

// 提交缩放
const handleScaleSubmit = async () => {
  if (!scaleFormRef.value) return;

  await scaleFormRef.value.validate();
  submitting.value = true;

  try {
    await scaleK8sDeployment(selectedCluster.value!, {
      namespace: filters.namespace,
      name: currentDeploymentName.value,
      replicas: scaleData.replicas
    });
    message.success('缩放成功');
    showScaleModal.value = false;
    loadDeployments();
  } catch (error: any) {
    message.error(error.message || '缩放失败');
  } finally {
    submitting.value = false;
  }
};

// 重启 Deployment
const handleRestart = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要重启 Deployment "${row.name}" 吗？`, '确认重启', {
      type: 'warning'
    });

    await restartK8sDeployment(selectedCluster.value!, {
      namespace: filters.namespace,
      name: row.name
    });
    message.success('重启成功');
    loadDeployments();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '重启失败');
    }
  }
};

// 删除 Deployment
const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Deployment "${row.name}" 吗？此操作危险，请谨慎操作。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '危险操作确认',
      cancelButtonText: '取消'
    });

    await deleteK8sDeployment(selectedCluster.value!, {
      namespace: filters.namespace,
      name: row.name
    });
    message.success('删除成功');
    loadDeployments();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除失败');
    }
  }
};

// 刷新
const handleRefresh = () => {
  loadDeployments();
};

// 分页变化
const handlePageChange = (page: number) => {
  pagination.page = page;
  loadDeployments();
};

// 每页数量变化
const handlePageSizeChange = (pageSize: number) => {
  pagination.pageSize = pageSize;
  pagination.page = 1; // 重置到第一页
  loadDeployments();
};

// 查看详情
const handleViewDetail = (row: any) => {
  router.push({
    path: '/k8s/resources/deployments/detail',
    query: {
      clusterId: selectedCluster.value?.toString(),
      namespace: filters.namespace,
      name: row.name
    }
  });
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

    <!-- Deployment 列表 -->
    <el-table v-loading="loading" :data="dataSource" stripe>
      <el-table-column prop="name" label="名称" min-width="200" />
      <el-table-column prop="namespace" label="命名空间" width="150" />
      <el-table-column label="状态" width="150">
        <template #default="{ row }">
          <el-tag :type="getStatusTag(row).type">
            {{ getStatusTag(row).text }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="副本数" width="120" align="center">
        <template #default="{ row }">
          {{ row.ready || 0 }} / {{ row.replicas || 0 }}
        </template>
      </el-table-column>
      <el-table-column prop="age" label="年龄" width="160" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-space>
            <el-button size="small" @click="handleViewDetail(row)">
              详情
            </el-button>
            <el-button size="small" type="primary" @click="handleScale(row)">
              缩放
            </el-button>
            <el-button size="small" type="warning" @click="handleRestart(row)">
              重启
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </el-space>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="mt-4 flex justify-end">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.itemCount"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handlePageSizeChange"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 缩放对话框 -->
    <el-dialog v-model="showScaleModal" title="缩放 Deployment" width="500px">
      <el-form ref="scaleFormRef" :model="scaleData" :rules="scaleRules" label-width="100px">
        <el-form-item label="Deployment">
          <el-input :value="currentDeploymentName" disabled />
        </el-form-item>
        <el-form-item label="副本数" prop="replicas">
          <el-input-number
            v-model="scaleData.replicas"
            :min="0"
            :max="100"
            :step="1"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showScaleModal = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleScaleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
