<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
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
  fetchK8sClusterNamespaces,
  fetchK8sClusters,
  fetchK8sDeployments,
  restartK8sDeployment,
  scaleK8sDeployment
} from '@/service/api/k8s';

defineOptions({ name: 'K8sDeployments' });

const router = useRouter();
const route = useRoute();
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
  }
  return { type: 'danger', text: '未就绪' };
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
  console.log('[Deployment列表] ========== handleViewDetail 被调用 ==========');
  console.log('[Deployment列表] 当前选择的集群:', selectedCluster.value);
  console.log('[Deployment列表] 当前选择的命名空间:', filters.namespace);
  console.log('[Deployment列表] 点击的行数据:', row);

  // 保存当前选择到 sessionStorage
  const stateToSave = {
    clusterId: selectedCluster.value,
    namespace: filters.namespace,
    listPath: '/k8s/workloads'
  };
  console.log('[Deployment列表] 准备保存状态到 sessionStorage:', stateToSave);

  try {
    sessionStorage.setItem('k8s_deployments_list_state', JSON.stringify(stateToSave));
    console.log('[Deployment列表] ✅ sessionStorage 保存成功');

    // 验证保存是否成功
    const saved = sessionStorage.getItem('k8s_deployments_list_state');
    console.log('[Deployment列表] 验证保存结果:', saved);
  } catch (e) {
    console.error('[Deployment列表] ❌ sessionStorage 保存失败:', e);
  }

  router.push({
    path: '/k8s/resources/deployments/detail',
    query: {
      clusterId: selectedCluster.value?.toString(),
      namespace: filters.namespace,
      name: row.name
    }
  });
};

// 从 sessionStorage 恢复状态
const restoreStateFromStorage = () => {
  console.log('[Deployment列表] 开始恢复状态...');
  const savedState = sessionStorage.getItem('k8s_deployments_list_state');
  console.log('[Deployment列表] sessionStorage 内容:', savedState);

  if (savedState) {
    try {
      const state = JSON.parse(savedState);
      console.log('[Deployment列表] 解析后的状态:', state);

      selectedCluster.value = state.clusterId;
      filters.namespace = state.namespace;
      console.log('[Deployment列表] 已设置 clusterId:', selectedCluster.value, 'namespace:', filters.namespace);

      // 清除保存的状态
      sessionStorage.removeItem('k8s_deployments_list_state');
      console.log('[Deployment列表] 已清除 sessionStorage');

      // 加载命名空间和数据
      if (selectedCluster.value) {
        console.log('[Deployment列表] 开始加载命名空间和数据...');
        return loadNamespaces().then(() => {
          console.log('[Deployment列表] 命名空间加载完成，开始加载数据...');
          return loadDeployments();
        });
      }
    } catch (e) {
      console.error('[Deployment列表] 恢复状态失败:', e);
    }
  }

  // 正常加载流程
  console.log('[Deployment列表] 无保存状态，执行正常加载流程');
  return loadClusters();
};

onMounted(() => {
  restoreStateFromStorage();
});

// 监听路由变化，当从详情页返回时恢复状态
watch(
  () => route.path,
  (newPath, oldPath) => {
    console.log('[Deployment列表] 路由变化:', { oldPath, newPath });
    // 如果从详情页返回到列表页
    if (newPath === '/k8s/workloads' && oldPath?.includes('/detail')) {
      console.log('[Deployment列表] 检测到从详情页返回，触发状态恢复');
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
        <ElSelect
          v-model="filters.namespace"
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

    <!-- Deployment 列表 -->
    <ElTable v-loading="loading" :data="dataSource" stripe>
      <ElTableColumn prop="name" label="名称" min-width="200">
        <template #default="{ row }">
          <ElButton link type="primary" @click="handleViewDetail(row)">{{ row.name }}</ElButton>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="namespace" label="命名空间" width="150" />
      <ElTableColumn label="状态" width="150">
        <template #default="{ row }">
          <ElTag :type="getStatusTag(row).type">
            {{ getStatusTag(row).text }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="副本数" width="120" align="center">
        <template #default="{ row }">{{ row.ready || 0 }} / {{ row.replicas || 0 }}</template>
      </ElTableColumn>
      <ElTableColumn prop="age" label="年龄" width="160" />
      <ElTableColumn label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <ElSpace>
            <ElButton size="small" @click="handleViewDetail(row)">详情</ElButton>
            <ElButton size="small" type="primary" @click="handleScale(row)">缩放</ElButton>
            <ElButton size="small" type="warning" @click="handleRestart(row)">重启</ElButton>
            <ElButton size="small" type="danger" @click="handleDelete(row)">删除</ElButton>
          </ElSpace>
        </template>
      </ElTableColumn>
    </ElTable>

    <!-- 分页 -->
    <div class="mt-4 flex justify-end">
      <ElPagination
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
    <ElDialog v-model="showScaleModal" title="缩放 Deployment" width="500px">
      <ElForm ref="scaleFormRef" :model="scaleData" :rules="scaleRules" label-width="100px">
        <ElFormItem label="Deployment">
          <ElInput :value="currentDeploymentName" disabled />
        </ElFormItem>
        <ElFormItem label="副本数" prop="replicas">
          <ElInputNumber v-model="scaleData.replicas" :min="0" :max="100" :step="1" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showScaleModal = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleScaleSubmit">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
