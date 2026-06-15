<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import {
  ElButton,
  ElDialog,
  ElDrawer,
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
  createK8sCluster,
  deleteK8sCluster,
  fetchK8sClusters,
  testK8sConnection,
  updateK8sCluster
} from '@/service/api/k8s';
import ClusterDetail from './ClusterDetail.vue';

defineOptions({ name: 'K8sClusters' });

const message = ElMessage;

const loading = ref(false);
const dataSource = ref<any[]>([]);
const showModal = ref(false);
const showDetail = ref(false);
const modalTitle = ref('添加集群');
const submitting = ref(false);
const currentCluster = ref<any>(null);
const formRef = ref<FormInstance | null>(null);

const filters = reactive({
  name: '',
  clusterType: null,
  status: null
});

const formData = reactive({
  id: undefined,
  name: '',
  description: '',
  endpoint: '',
  kubeconfig: '',
  clusterType: 'standard',
  region: '',
  nodeCount: 0
});

const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0
});

const formRules: FormRules = {
  name: { required: true, message: '请输入集群名称', trigger: 'blur' },
  endpoint: { required: true, message: '请输入 API 地址', trigger: 'blur' },
  kubeconfig: { required: true, message: '请输入 kubeconfig', trigger: 'blur' },
  clusterType: { required: true, message: '请选择集群类型', trigger: 'change' }
};

// 加载集群列表
const loadClusters = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page,
      pageSize: pagination.pageSize
    };
    if (filters.name) params.name = filters.name;
    if (filters.clusterType) params.clusterType = filters.clusterType;
    if (filters.status !== null) params.status = filters.status;

    const res = await fetchK8sClusters(params);
    dataSource.value = res.data || [];
    pagination.itemCount = res.total || 0;
  } catch (error: any) {
    message.error(error.message || '加载集群列表失败');
  } finally {
    loading.value = false;
  }
};

// 创建集群
const handleCreate = () => {
  modalTitle.value = '添加集群';
  Object.assign(formData, {
    id: undefined,
    name: '',
    description: '',
    endpoint: '',
    kubeconfig: '',
    clusterType: 'standard',
    region: '',
    nodeCount: 0
  });
  showModal.value = true;
};

// 编辑集群
const handleEdit = (row: any) => {
  modalTitle.value = '编辑集群';
  Object.assign(formData, row);
  showModal.value = true;
};

// 查看详情
const handleView = (row: any) => {
  currentCluster.value = row;
  showDetail.value = true;
};

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate();
  submitting.value = true;

  try {
    if (modalTitle.value === '添加集群') {
      await createK8sCluster(formData);
      message.success('创建集群成功');
    } else {
      await updateK8sCluster(formData.id, formData);
      message.success('更新集群成功');
    }
    showModal.value = false;
    loadClusters();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  } finally {
    submitting.value = false;
  }
};

// 测试连接
const handleTestConnection = async (row: any) => {
  try {
    await testK8sConnection(row.id);
    message.success('连接测试成功');
  } catch (error: any) {
    message.error(error.message || '连接测试失败');
  }
};

// 删除集群
const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除集群 "${row.name}" 吗？`, '确认删除', {
      type: 'warning'
    });

    const confirmName = await ElMessageBox.prompt('请输入集群名称以确认删除：', '二次确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });

    if (confirmName.value !== row.name) {
      message.warning('集群名称不匹配，取消删除');
      return;
    }

    await deleteK8sCluster(row.id, { confirmName: confirmName.value });
    message.success('删除集群成功');
    loadClusters();
  } catch (error: any) {
    if (error !== 'cancel') {
      message.error(error.message || '删除集群失败');
    }
  }
};

// 搜索
const handleSearch = () => {
  pagination.page = 1;
  loadClusters();
};

// 刷新
const handleRefresh = () => {
  loadClusters();
};

// 分页变化
const handlePageChange = (page: number) => {
  pagination.page = page;
  loadClusters();
};

onMounted(() => {
  loadClusters();
});
</script>

<template>
  <div class="k8s-clusters-page h-full flex flex-col overflow-hidden">
    <!-- 头部：搜索和操作按钮（同一行）-->
    <div class="mb-8px flex items-center justify-between gap-12px px-16px">
      <!-- 左侧：主操作按钮 -->
      <div class="flex items-center gap-8px">
        <ElButton
          type="primary"
          :style="{
            backgroundColor: '#0052D9',
            borderColor: '#0052D9',
            color: '#fff',
            borderRadius: '0',
            fontSize: '12px',
            height: '30px'
          }"
          @click="handleCreate"
        >
          <icon-mdi-plus class="text-icon" />
          添加集群
        </ElButton>
      </div>

      <!-- 右侧：搜索栏和刷新按钮 -->
      <div class="search-inputs flex items-center gap-8px">
        <ElSelect v-model="filters.clusterType" placeholder="集群类型" clearable style="width: 120px">
          <ElOption label="标准集群" value="standard" />
          <ElOption label="托管集群" value="managed" />
          <ElOption label="边缘集群" value="edge" />
        </ElSelect>
        <ElSelect v-model="filters.status" placeholder="状态" clearable style="width: 100px">
          <ElOption label="正常" :value="1" />
          <ElOption label="禁用" :value="0" />
        </ElSelect>
        <ElInput
          v-model="filters.name"
          placeholder="搜索集群名称"
          clearable
          style="width: 200px"
          @keyup.enter="handleSearch"
        >
          <template #suffix>
            <icon-ic-round-search class="cursor-pointer text-icon" @click="handleSearch" />
          </template>
        </ElInput>
        <ElButton text @click="handleRefresh">
          <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
        </ElButton>
      </div>
    </div>

    <!-- 集群列表容器 -->
    <div class="flex-1 overflow-hidden bg-white px-16px pb-16px">
      <!-- 列表内容 -->
      <div class="flex-1 overflow-auto">
        <ElTable
          v-loading="loading"
          height="100%"
          :data="dataSource"
          size="small"
          class="cluster-list-table compact-table"
          :row-style="{ height: '48px' }"
          :cell-style="{ padding: '0', borderRight: 'none' }"
          :header-cell-style="{ backgroundColor: '#f5f7fa', borderRight: 'none' }"
          table-layout="fixed"
          stripe
        >
          <ElTableColumn prop="name" label="集群名称" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="cursor-pointer text-primary hover:underline" @click="handleView(row)">
                {{ row.name }}
              </span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="description" label="描述" min-width="120" show-overflow-tooltip />
          <ElTableColumn prop="endpoint" label="API 地址" min-width="200" show-overflow-tooltip />
          <ElTableColumn prop="clusterType" label="类型" width="100" />
          <ElTableColumn prop="region" label="区域" width="100" />
          <ElTableColumn prop="version" label="版本" width="80" show-overflow-tooltip />
          <ElTableColumn prop="nodeCount" label="节点数" width="80" align="center" />
          <ElTableColumn prop="status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <ElTag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                {{ row.status === 1 ? '正常' : '禁用' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="createdAt" label="创建时间" width="160" />
          <ElTableColumn label="操作" width="280" fixed="right" align="center">
            <template #default="{ row }">
              <ElButton link type="primary" size="small" @click="handleView(row)">详情</ElButton>
              <ElButton link type="primary" size="small" @click="handleEdit(row)">编辑</ElButton>
              <ElButton link size="small" @click="handleTestConnection(row)">测试</ElButton>
              <ElButton link type="danger" size="small" @click="handleDelete(row)">删除</ElButton>
            </template>
          </ElTableColumn>
        </ElTable>

        <!-- 分页 -->
        <div class="mt-12px flex items-center justify-end gap-12px">
          <ElPagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="pagination.itemCount"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="handlePageChange"
            @size-change="handlePageChange"
          />
        </div>
      </div>
    </div>

    <!-- 创建/编辑集群对话框 -->
    <ElDialog v-model="showModal" :title="modalTitle" width="600px" :close-on-click-modal="false">
      <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="100px" class="mt-4">
        <ElFormItem label="集群名称" prop="name">
          <ElInput v-model="formData.name" placeholder="请输入集群名称" />
        </ElFormItem>
        <ElFormItem label="集群描述" prop="description">
          <ElInput v-model="formData.description" type="textarea" placeholder="请输入集群描述" :rows="3" />
        </ElFormItem>
        <ElFormItem label="API 地址" prop="endpoint">
          <ElInput v-model="formData.endpoint" placeholder="https://k8s-api.example.com:6443" />
        </ElFormItem>
        <ElFormItem label="Kubeconfig" prop="kubeconfig">
          <ElInput v-model="formData.kubeconfig" type="textarea" placeholder="粘贴 kubeconfig 内容" :rows="10" />
        </ElFormItem>
        <ElFormItem label="集群类型" prop="clusterType">
          <ElSelect v-model="formData.clusterType" placeholder="选择集群类型">
            <ElOption label="标准集群" value="standard" />
            <ElOption label="托管集群" value="managed" />
            <ElOption label="边缘集群" value="edge" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="区域" prop="region">
          <ElInput v-model="formData.region" placeholder="如：us-west-2" />
        </ElFormItem>
        <ElFormItem label="节点数" prop="nodeCount">
          <ElInputNumber v-model="formData.nodeCount" :min="0" placeholder="自动获取" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElSpace justify="end">
          <ElButton @click="showModal = false">取消</ElButton>
          <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
        </ElSpace>
      </template>
    </ElDialog>

    <!-- 集群详情侧边栏 -->
    <ElDrawer v-model="showDetail" title="集群详情" size="800px">
      <ClusterDetail v-if="currentCluster" :cluster-id="currentCluster.id" />
    </ElDrawer>
  </div>
</template>

<style scoped>
.k8s-clusters-page {
  background-color: #f5f7fa;
}

.text-icon {
  font-size: 18px;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.cluster-list-table :deep(.el-table__header) {
  th {
    font-weight: 600;
  }
}

.cluster-list-table :deep(.el-table__body tr:hover > td) {
  background-color: #f5f7fa;
}

.cluster-list-table :deep(.el-table__body .cell) {
  padding: 0 16px;
}

.cluster-list-table :deep(.compact-table .el-table__cell) {
  padding: 0 16px;
}
</style>
