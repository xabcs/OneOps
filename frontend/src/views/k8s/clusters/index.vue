<script setup lang="tsx">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElNotification } from 'element-plus';
import { Delete, Platform, Plus, Refresh, RefreshRight, Search } from '@element-plus/icons-vue';
import {
  createK8sCluster,
  deleteK8sCluster,
  fetchK8sClusters,
  testK8sConnection,
  updateK8sCluster
} from '@/service/api/k8s';
import { useThemeStore } from '@/store/modules/theme';

defineOptions({ name: 'ClusterManage' });

const themeStore = useThemeStore();

// Hero区域显示状态
const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

const message = ElNotification;

// 搜索参数
const queryParams = reactive({
  name: '',
  clusterType: null,
  status: null
});

// 表单数据
const form = ref({
  id: null,
  name: '',
  description: '',
  endpoint: '',
  kubeconfig: '',
  clusterType: 'standard',
  region: '',
  nodeCount: 0
});

// 弹窗状态
const dialogVisible = ref(false);
const loading = ref(false);
const submitting = ref(false);

// 列表数据
const clusterList = ref<any[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);

// 选中的行
const checkedRowKeys = ref<any[]>([]);

// 显示列表（分页）
const displayClusters = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return clusterList.value.slice(start, end);
});

const modalTitle = computed(() => (form.value.id ? '编辑集群' : '添加集群'));

// 加载集群列表
async function loadClusters() {
  loading.value = true;
  try {
    const res = await fetchK8sClusters(queryParams);
    if (res.data) {
      clusterList.value = res.data || [];
      total.value = res.total || 0;
    }
  } catch (error: any) {
    message({
      title: '加载失败',
      message: error.message || '加载集群列表失败',
      type: 'error',
      duration: 3000
    });
  } finally {
    loading.value = false;
  }
}

// 创建集群
function handleAdd() {
  form.value = {
    id: null,
    name: '',
    description: '',
    endpoint: '',
    kubeconfig: '',
    clusterType: 'standard',
    region: '',
    nodeCount: 0
  };
  dialogVisible.value = true;
}

// 编辑集群
function handleEdit(row: any) {
  form.value = { ...row };
  dialogVisible.value = true;
}

// 提交表单
async function handleSubmit() {
  submitting.value = true;
  try {
    let res;
    if (form.value.id) {
      res = await updateK8sCluster(form.value.id, form.value);
    } else {
      res = await createK8sCluster(form.value);
    }

    if (res.data || res.code === 200) {
      message({
        title: '操作成功',
        message: form.value.id ? '更新集群成功' : '创建集群成功',
        type: 'success',
        duration: 3000
      });
      dialogVisible.value = false;
      loadClusters();
    } else {
      message({
        title: '操作失败',
        message: res.message || '操作失败',
        type: 'error',
        duration: 3000
      });
    }
  } catch (error: any) {
    message({
      title: '操作失败',
      message: error.message || '操作失败',
      type: 'error',
      duration: 3000
    });
  } finally {
    submitting.value = false;
  }
}

// 测试连接
async function handleTestConnection(row: any) {
  try {
    const res = await testK8sConnection(row.id);
    if (res.data || res.code === 200) {
      message({
        title: '测试成功',
        message: `集群 "${row.name}" 连接测试成功`,
        type: 'success',
        duration: 3000
      });
    } else {
      message({
        title: '测试失败',
        message: res.message || '连接测试失败',
        type: 'error',
        duration: 3000
      });
    }
  } catch (error: any) {
    message({
      title: '测试失败',
      message: error.message || '连接测试失败',
      type: 'error',
      duration: 3000
    });
  }
}

// 删除集群
async function handleDelete(row: any) {
  try {
    const res = await deleteK8sCluster(row.id, { confirmName: row.name });
    if (res.data || res.code === 200) {
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
        message: res.message || '删除集群失败',
        type: 'error',
        duration: 3000
      });
    }
  } catch (error: any) {
    message({
      title: '删除失败',
      message: error.message || '删除集群失败',
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
      const res = await deleteK8sCluster(row.id, { confirmName: row.name });
      if (res.data || res.code === 200) {
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

// 重置搜索
function handleResetSearch() {
  queryParams.name = '';
  queryParams.clusterType = null;
  queryParams.status = null;
  currentPage.value = 1;
  loadClusters();
}

// 刷新
function handleRefresh() {
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
  <div class="page-container cluster-management-page">
    <!-- Hero 区域 -->
    <section
      v-if="heroVisible"
      class="hero-section"
      :style="{
        background: 'var(--sx-hero-bg)',
        border: '1px solid var(--sx-hero-border)',
        borderRadius: 'var(--sx-hero-radius)',
        boxShadow: 'var(--sx-hero-shadow)',
        padding: 'var(--sx-hero-padding)'
      }"
    >
      <div class="hero-content">
        <div class="hero-title-row">
          <span class="hero-icon">
            <ElIcon><Platform /></ElIcon>
          </span>
          <h2>集群管理</h2>
          <p class="hero-desc">管理 Kubernetes 集群连接和配置信息，支持多集群统一管理和监控</p>
        </div>
      </div>
      <div class="hero-actions">
        <ElButton size="small" :loading="loading" @click="handleRefresh">
          <ElIcon><Refresh /></ElIcon>
          刷新
        </ElButton>
      </div>
    </section>

    <!-- 内容卡片 -->
    <div
      class="content-card cluster-content-card"
      :style="{
        background: 'var(--sx-content-card-bg)',
        border: '1px solid var(--sx-content-card-border)',
        borderRadius: 'var(--sx-content-card-radius)',
        boxShadow: 'var(--sx-content-card-shadow)',
        padding: 'var(--sx-content-card-padding)'
      }"
    >
      <!-- 工具栏 -->
      <div class="card-toolbar">
        <div class="toolbar-head">
          <span class="toolbar-title">集群列表</span>
          <span class="toolbar-desc">共 {{ total }} 个集群</span>
        </div>
        <div class="toolbar-actions">
          <ElButton type="primary" size="small" @click="handleAdd">
            <ElIcon><Plus /></ElIcon>
            添加集群
          </ElButton>
          <ElButton type="danger" size="small" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
            <ElIcon><Delete /></ElIcon>
            批量删除
          </ElButton>
        </div>
      </div>

      <!-- 搜索工具栏 -->
      <div class="workbench-toolbar workbench-toolbar--history clusters-toolbar">
        <div class="workbench-toolbar-left">
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
            style="width: 200px"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <ElIcon><Search /></ElIcon>
            </template>
          </ElInput>
        </div>
        <div class="workbench-toolbar-right">
          <ElButton class="filter-refresh-btn" @click="handleResetSearch">
            <ElIcon><RefreshRight /></ElIcon>
            重置
          </ElButton>
          <ElButton class="filter-refresh-btn" type="primary" @click="handleSearch">
            <ElIcon><Search /></ElIcon>
            搜索
          </ElButton>
        </div>
      </div>

      <!-- 数据表格 -->
      <div class="table-section">
        <ElTable
          v-loading="loading"
          :data="displayClusters"
          border
          stripe
          class="data-table"
          row-key="id"
          @selection-change="(selection: any[]) => (checkedRowKeys = selection)"
        >
          <ElTableColumn type="selection" width="48" align="center" />
          <ElTableColumn type="index" label="序号" width="64" align="center" />
          <ElTableColumn prop="name" label="集群名称" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="cluster-name-link">{{ row.name }}</span>
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
              <div class="flex-center gap-8px">
                <ElButton link type="primary" @click="handleEdit(row)">编辑</ElButton>
                <ElButton link class="text-accent" @click="handleTestConnection(row)">测试</ElButton>
                <ElPopconfirm title="确定要删除集群吗？" @confirm="handleDelete(row)">
                  <template #reference>
                    <ElButton link type="danger">删除</ElButton>
                  </template>
                </ElPopconfirm>
              </div>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>

      <!-- 分页 -->
      <div class="table-pagination">
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
    </div>

    <!-- 创建/编辑集群对话框 -->
    <ElDialog v-model="dialogVisible" :title="modalTitle" width="600px" :close-on-click-modal="false">
      <ElForm :model="form" label-width="100px">
        <ElFormItem label="集群名称" required>
          <ElInput v-model="form.name" placeholder="请输入集群名称" />
        </ElFormItem>
        <ElFormItem label="集群描述">
          <ElInput v-model="form.description" type="textarea" placeholder="请输入集群描述" :rows="3" />
        </ElFormItem>
        <ElFormItem label="API 地址" required>
          <ElInput v-model="form.endpoint" placeholder="https://k8s-api.example.com:6443" />
        </ElFormItem>
        <ElFormItem label="Kubeconfig" required>
          <ElInput v-model="form.kubeconfig" type="textarea" placeholder="粘贴 kubeconfig 内容" :rows="10" />
        </ElFormItem>
        <ElFormItem label="集群类型" required>
          <ElSelect v-model="form.clusterType" placeholder="选择集群类型">
            <ElOption label="标准集群" value="standard" />
            <ElOption label="托管集群" value="managed" />
            <ElOption label="边缘集群" value="edge" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="区域">
          <ElInput v-model="form.region" placeholder="如：us-west-2" />
        </ElFormItem>
        <ElFormItem label="节点数">
          <ElInputNumber v-model="form.nodeCount" :min="0" placeholder="自动获取" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
/* ============================================
	   1. 页面容器与布局
	   ============================================ */
.cluster-management-page {
  padding: 16px 20px;
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* ============================================
	   2. Hero 区域
	   ============================================ */
.hero-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.hero-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.hero-title-row {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.hero-icon {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--sx-hero-icon-color);
  background: var(--sx-hero-icon-bg);
  border: 1px solid var(--sx-hero-icon-border);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.hero-title-row h2 {
  color: var(--sx-text-primary);
  font-size: 23px;
  font-weight: 700;
  margin: 0;
}

.hero-desc {
  color: var(--sx-text-secondary);
  font-size: 13px;
  line-height: 1.45;
  margin: 0;
  max-width: 600px;
}

.hero-actions {
  display: flex;
  gap: 4px;
}

/* ============================================
	   3. 内容卡片
	   ============================================ */
.content-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--sx-border-soft);
}

.toolbar-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.toolbar-title {
  color: var(--sx-text-primary);
  font-size: 16px;
  font-weight: 700;
}

.toolbar-desc {
  color: var(--sx-text-secondary);
  font-size: 12px;
}

.toolbar-actions {
  display: flex;
  gap: 4px;
  align-items: center;
}

/* ============================================
	   4. 搜索工具栏
	   ============================================ */
.workbench-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin: 6px 0 8px;
}

.workbench-toolbar-left,
.workbench-toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.workbench-toolbar.workbench-toolbar--history {
  background: var(--sx-toolbar-bg);
  border: 1px solid var(--sx-toolbar-border);
  border-radius: var(--sx-toolbar-radius);
  padding: var(--sx-toolbar-padding);
  box-shadow: var(--sx-toolbar-shadow);
  gap: 8px;
  margin: 6px 0;

  .workbench-toolbar-left,
  .workbench-toolbar-right {
    gap: 5px;
    flex-wrap: nowrap; /* 防止元素换行 */
  }

  .el-input__wrapper,
  .el-select__wrapper {
    background: var(--sx-search-input-bg);
    border-radius: var(--sx-search-input-radius);
    box-shadow: 0 0 0 1px var(--sx-search-input-border) inset;
    transition: all 0.3s ease;

    &:hover {
      box-shadow: 0 0 0 1px var(--sx-search-input-hover-border) inset;
    }
  }

  .el-input.is-focus .el-input__wrapper,
  .el-select.is-focus .el-select__wrapper {
    box-shadow: 0 0 0 1px var(--sx-search-input-focus-border) inset;
  }

  .el-button:not(.is-link) {
    background: var(--sx-search-button-bg);
    color: var(--sx-search-button-text);
    transition: all 0.3s ease;

    &:hover {
      background: var(--sx-search-button-hover);
    }
  }
}

/* ============================================
	   5. 数据表格
	   ============================================ */
.table-section {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.data-table {
  width: 100%;
  border-radius: var(--sx-table-radius);
  overflow: hidden;
  border: 1px solid var(--sx-table-border);
  border-collapse: separate;
  border-spacing: 0;

  --el-table-border-color: var(--sx-table-border);
  --el-table-header-bg-color: var(--sx-table-header-bg);
  --el-table-row-hover-bg-color: var(--sx-table-row-hover);

  :deep(th.el-table__cell) {
    background-color: var(--sx-table-header-bg);
    color: var(--sx-table-header-text);
    font-weight: 600;
    border-radius: 0;
  }

  :deep(td.el-table__cell) {
    border-radius: 0;
  }

  :deep(.el-checkbox__inner),
  :deep(.el-checkbox__inner::before),
  :deep(.el-checkbox__inner::after) {
    border-radius: 0;
  }

  &--striped :deep(.el-table__body tr.el-table__row--striped td) {
    background-color: var(--sx-table-striped-bg);
  }
}

.cluster-name-link {
  color: var(--accent, #245bdb);
  cursor: pointer;
  transition: color 0.2s;

  &:hover {
    color: var(--accent-hover, #1d4ed8);
    text-decoration: underline;
  }
}

.text-accent {
  color: var(--accent, #245bdb) !important;

  &:hover {
    color: var(--accent-hover, #1d4ed8) !important;
    text-decoration: underline;
  }
}

/* ============================================
	   6. 分页组件
	   ============================================ */
.table-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;

  .el-pagination {
    .btn-next,
    .btn-prev,
    .el-pager li {
      background: var(--sx-pagination-button-bg);
      color: var(--sx-pagination-button-text);
      border-radius: var(--sx-pagination-radius);
      transition: all 0.3s ease;

      &:hover {
        background: var(--sx-pagination-button-hover);
      }

      &.is-active {
        background: var(--sx-pagination-active-bg);
        color: var(--sx-pagination-active-text);
      }
    }
  }
}

/* ============================================
	   7. 响应式设计
	   ============================================ */
@media (max-width: 768px) {
  .hero-section {
    flex-direction: column;
    align-items: flex-start;
  }

  .card-toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
}
</style>
