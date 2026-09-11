<script setup lang="tsx">
  import { computed, onMounted, reactive, ref } from 'vue';
  import { ElNotification } from 'element-plus';
  import { Delete, Plus, Refresh, Search } from '@element-plus/icons-vue';
  import { deleteK8sCluster, fetchK8sClusters, testK8sConnection } from '@/service/api/k8s';
  import ClusterFormDialog from './modules/ClusterFormDialog.vue';
  import ClusterDetail from './ClusterDetail.vue';

  defineOptions({ name: 'ClusterManage' });

  const message = ElNotification;

  // 搜索参数
  const queryParams = reactive({
    name: '',
    clusterType: undefined as string | undefined,
    status: undefined as number | undefined
  });

  // 弹窗状态
  const loading = ref(false);
  const showDialog = ref(false);
  const dialogMode = ref<'create' | 'edit'>('create');
  const currentCluster = ref<K8s.Cluster | null>(null);

  // 列表数据
  const clusterList = ref<K8s.Cluster[]>([]);
  const total = ref(0);
  const currentPage = ref(1);
  const pageSize = ref(10);

  // 选中的行
  const checkedRowKeys = ref<K8s.Cluster[]>([]);

  // 显示列表（分页）
  const displayClusters = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value;
    const end = start + pageSize.value;
    return clusterList.value.slice(start, end);
  });

  // 加载集群列表
  async function loadClusters() {
    loading.value = true;
    try {
      const { data, error } = await fetchK8sClusters(queryParams);
      if (!error && data) {
        clusterList.value = data.list || [];
        total.value = data.total || 0;
      }
    } catch (error: unknown) {
      const err = error as Error;
      message({
        title: '加载失败',
        message: err.message || '加载集群列表失败',
        type: 'error',
        duration: 3000
      });
    } finally {
      loading.value = false;
    }
  }

  // 创建集群
  function handleAdd() {
    dialogMode.value = 'create';
    currentCluster.value = null;
    showDialog.value = true;
  }

  // 编辑集群
  function handleEdit(row: K8s.Cluster) {
    dialogMode.value = 'edit';
    currentCluster.value = row;
    showDialog.value = true;
  }

  // 集群详情（含用户权限/用户组授权 tab）
  const detailVisible = ref(false);
  const detailClusterId = ref(0);

  function handleDetail(row: K8s.Cluster) {
    detailClusterId.value = row.id;
    detailVisible.value = true;
  }

  // 测试连接
  async function handleTestConnection(row: K8s.Cluster) {
    try {
      const { error } = await testK8sConnection(row.id);
      if (!error) {
        message({
          title: '测试成功',
          message: `集群 "${row.name}" 连接测试成功`,
          type: 'success',
          duration: 3000
        });
      } else {
        message({
          title: '测试失败',
          message: '连接测试失败',
          type: 'error',
          duration: 3000
        });
      }
    } catch (error: unknown) {
      const err = error as Error;
      message({
        title: '测试失败',
        message: err.message || '连接测试失败',
        type: 'error',
        duration: 3000
      });
    }
  }

  // 删除集群
  async function handleDelete(row: K8s.Cluster) {
    try {
      const { error } = await deleteK8sCluster(row.id, { confirmName: row.name });
      if (!error) {
        message({
          title: '删除成功',
          message: `集群 "${row.name}" 已成功删除`,
          type: 'success',
          duration: 3000
        });
        loadClusters();
      } else {
        message({
          title: '删除失败',
          message: '删除集群失败',
          type: 'error',
          duration: 3000
        });
      }
    } catch (error: unknown) {
      const err = error as Error;
      message({
        title: '删除失败',
        message: err.message || '删除集群失败',
        type: 'error',
        duration: 3000
      });
    }
  }

  // 批量删除
  async function handleBatchDelete() {
    if (checkedRowKeys.value.length === 0) {
      message({
        title: '提示',
        message: '请选择要删除的集群',
        type: 'warning',
        duration: 3000
      });
      return;
    }

    let successCount = 0;
    let failCount = 0;

    for (const row of checkedRowKeys.value) {
      try {
        const { error } = await deleteK8sCluster(row.id, { confirmName: row.name });
        if (!error) {
          successCount++;
        } else {
          failCount++;
        }
      } catch (error) {
        failCount++;
      }
    }

    if (successCount > 0) {
      message({
        title: '批量删除完成',
        message: `成功删除 ${successCount} 个集群`,
        type: 'success',
        duration: 3000
      });
    }

    if (failCount > 0) {
      message({
        title: '部分删除失败',
        message: `${failCount} 个集群删除失败`,
        type: 'error',
        duration: 3000
      });
    }

    checkedRowKeys.value = [];
    loadClusters();
  }

  // 搜索
  function handleSearch() {
    currentPage.value = 1;
    loadClusters();
  }

  // 分页变化
  function handleSizeChange(val: number) {
    pageSize.value = val;
    loadClusters();
  }

  function handleCurrentChange(val: number) {
    currentPage.value = val;
    loadClusters();
  }

  onMounted(() => {
    loadClusters();
  });
</script>

<template>
  <div class="table-page p-24px">
    <!-- 页头：标题 + 描述（无容器风格，参考工作负载页） -->
    <div class="shrink-0">
      <h1 class="text-28px text-primary font-bold">集群管理</h1>
      <p class="text-tertiary mt-8px text-14px">管理 Kubernetes 集群连接和配置信息，支持多集群统一管理和监控</p>
    </div>

    <!-- 筛选与操作行 -->
    <div class="flex shrink-0 items-center justify-between gap-12px">
      <div class="flex flex-wrap items-center gap-8px">
        <ElSelect
          v-model="queryParams.clusterType"
          placeholder="集群类型"
          clearable
          style="width: 120px"
          @change="handleSearch"
        >
          <ElOption label="标准集群" value="standard" />
          <ElOption label="托管集群" value="managed" />
          <ElOption label="边缘集群" value="edge" />
        </ElSelect>
        <ElSelect
          v-model="queryParams.status"
          placeholder="状态"
          clearable
          style="width: 100px"
          @change="handleSearch"
        >
          <ElOption label="正常" :value="1" />
          <ElOption label="禁用" :value="0" />
        </ElSelect>
        <ElInput
          v-model="queryParams.name"
          placeholder="搜索集群名称"
          clearable
          style="width: 220px"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        >
          <template #prefix>
            <ElIcon>
              <Search />
            </ElIcon>
          </template>
        </ElInput>
      </div>
      <div class="flex items-center gap-8px">
        <ElButton text :loading="loading" @click="loadClusters">
          <ElIcon :class="{ 'animate-spin': loading }">
            <Refresh />
          </ElIcon>
        </ElButton>
        <PermissionButton code="k8s.cluster.create" type="primary" @click="handleAdd">
          <ElIcon>
            <Plus />
          </ElIcon>
          添加集群
        </PermissionButton>
        <PermissionButton
          code="k8s.cluster.delete"
          type="danger"
          plain
          :disabled="checkedRowKeys.length === 0"
          @click="handleBatchDelete"
        >
          <ElIcon>
            <Delete />
          </ElIcon>
          批量删除
        </PermissionButton>
      </div>
    </div>

    <!-- 数据表格（裸放，内部滚动） -->
    <div class="table-scroll-wrap">
      <ElTable
        v-loading="loading"
        :data="displayClusters"
        height="100%"
        border
        stripe
        row-key="id"
        @selection-change="(selection: K8s.Cluster[]) => (checkedRowKeys = selection)"
      >
        <ElTableColumn type="selection" width="48" align="center" />
        <ElTableColumn type="index" label="序号" width="64" align="center" />
        <ElTableColumn prop="name" label="集群名称" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cluster-name-link" @click="handleDetail(row)">{{ row.name }}</span>
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
        <ElTableColumn label="操作" width="280" fixed="right" align="center" class-name="msre-table-actions">
          <template #default="{ row }">
            <div>
              <ElButton link type="primary" size="small" @click="handleDetail(row)">详情</ElButton>
              <PermissionButton link type="primary" size="small" code="k8s.cluster.update" @click="handleEdit(row)">
                编辑
              </PermissionButton>
              <PermissionButton link type="primary" size="small" code="k8s.cluster.connect" @click="handleTestConnection(row)">
                测试
              </PermissionButton>
              <ElPopconfirm title="确定要删除集群吗？" @confirm="handleDelete(row)">
                <template #reference>
                  <PermissionButton link type="danger" size="small" code="k8s.cluster.delete">删除</PermissionButton>
                </template>
              </ElPopconfirm>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>
    </div>

    <!-- 分页 -->
    <div class="mt-16px flex shrink-0 justify-end">
      <ElPagination
        v-if="total > 0"
        layout="total, sizes, prev, pager, next, jumper"
        :current-page="currentPage"
        :page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 创建/编辑集群对话框 -->
    <ClusterFormDialog
      v-model:visible="showDialog"
      :mode="dialogMode"
      :form-data="currentCluster"
      @submitted="loadClusters"
    />

    <!-- 集群详情抽屉（节点/命名空间/用户权限/用户组授权） -->
    <ElDrawer v-model="detailVisible" title="集群详情" size="70%">
      <ClusterDetail v-if="detailVisible" :cluster-id="detailClusterId" />
    </ElDrawer>
  </div>
</template>

<style scoped lang="scss">
  /* 高度模型由 .table-page 提供（h-full + flex-col + gap + overflow-hidden）：整页不滚、表格内部滚动、分页常驻底部 */

  .cluster-name-link {
    color: var(--el-color-primary);
    cursor: pointer;
    transition: color 0.2s;

    &:hover {
      text-decoration: underline;
    }
  }
</style>
