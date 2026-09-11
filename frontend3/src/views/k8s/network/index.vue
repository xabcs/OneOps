<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import {
    ElButton,
    ElDropdown,
    ElDropdownItem,
    ElDropdownMenu,
    ElMessage,
    ElMessageBox,
    ElPagination,
    ElTabPane,
    ElTable,
    ElTableColumn,
    ElTabs,
    ElTag
  } from 'element-plus';
  import {
    deleteK8sIngress,
    deleteK8sService,
    fetchK8sIngresses,
    fetchK8sServices,
    getK8sIngress,
    getK8sService,
    updateK8sIngress,
    updateK8sService
  } from '@/service/api/k8s';
  import { parseManifest } from '@/views/k8s/shared/k8s-formatters';
  import { useClusterNamespace } from '@/views/k8s/composables/useClusterNamespace';
  import { useYamlEdit } from '@/views/k8s/composables/useYamlEdit';
  import YamlEditor from '@/components/k8s/YamlEditor.vue';

  defineOptions({ name: 'K8sNetwork' });

  const router = useRouter();
  const message = ElMessage;

  const loading = ref(false);
  const activeTab = ref('services');

  // 集群/命名空间初始化与联动统一走 useClusterNamespace；就绪后加载当前 Tab 数据
  const { clusters, namespaces, selectedCluster, selectedNamespace, loadAll, handleClusterChange, handleNamespaceChange } =
    useClusterNamespace({
      onDataReady: () => loadCurrentData()
    });

  // 资源数据
  const servicesData = ref<K8s.Service[]>([]);
  const ingressesData = ref<K8s.Ingress[]>([]);

  // 选中状态
  const selectedServices = ref<K8s.Service[]>([]);
  const selectedIngresses = ref<K8s.Ingress[]>([]);

  // 分页
  const servicesPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });
  const ingressesPagination = reactive({ page: 1, pageSize: 10, itemCount: 0 });

  // YAML 编辑器状态
  const showYamlDialog = ref(false);
  const yamlDialogTitle = ref('');
  const yamlContent = ref('');
  const selectedResource = ref<K8s.Service | K8s.Ingress | null>(null);
  const yamlLoading = ref(false);

  // 当前数据
  const currentPagination = computed(() => {
    return activeTab.value === 'services' ? servicesPagination : ingressesPagination;
  });

  // 设置选中项
  function setSelectedItems(items: K8s.Service[] | K8s.Ingress[]) {
    if (activeTab.value === 'services') {
      // services 表格勾选行，按当前 Tab 收窄类型
      selectedServices.value = items as K8s.Service[];
    } else {
      selectedIngresses.value = items as K8s.Ingress[];
    }
  }

  // 加载服务
  async function loadServices() {
    if (!selectedCluster.value) return;
    loading.value = true;
    const { data, error } = await fetchK8sServices(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: servicesPagination.page,
      pageSize: servicesPagination.pageSize
    });

    if (!error) {
      servicesData.value = data?.list || [];
      servicesPagination.itemCount = data?.total || 0;
    } else {
      console.error('加载 Services 失败:', error);
    }
    loading.value = false;
  }

  // 加载路由
  async function loadIngresses() {
    if (!selectedCluster.value) return;
    loading.value = true;
    const { data, error } = await fetchK8sIngresses(selectedCluster.value, {
      namespace: selectedNamespace.value,
      page: ingressesPagination.page,
      pageSize: ingressesPagination.pageSize
    });

    if (!error) {
      ingressesData.value = data?.list || [];
      ingressesPagination.itemCount = data?.total || 0;
    } else {
      console.error('加载 Ingresses 失败:', error);
    }
    loading.value = false;
  }

  // 加载当前Tab数据
  async function loadCurrentData() {
    if (activeTab.value === 'services') {
      await loadServices();
    } else {
      await loadIngresses();
    }
  }

  // Service 类型显示
  const getServiceTypeTag = (
    type: string
  ): { type: 'primary' | 'success' | 'warning' | 'info'; text: string } => {
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

  // 跳转到详情页
  function goToDetail(row: K8s.Service | K8s.Ingress) {
    if (activeTab.value === 'services') {
      router.push({
        path: '/k8s/resources/services/detail',
        query: { clusterId: selectedCluster.value, namespace: row.namespace, name: row.name }
      });
    } else {
      router.push({
        path: '/k8s/resources/ingresses/detail',
        query: { clusterId: selectedCluster.value, namespace: row.namespace, name: row.name }
      });
    }
  }

  // Service YAML 编辑
  async function handleServiceEdit(row: K8s.Service) {
    if (!selectedCluster.value) {
      message.warning('请先选择集群');
      return;
    }

    selectedResource.value = row;
    yamlDialogTitle.value = `编辑 Service: ${row.name}`;
    yamlLoading.value = true;

    try {
      const { data, error } = await getK8sService(selectedCluster.value, row.namespace, row.name);
      if (error) {
        message.error('获取 YAML 失败');
        return;
      }
      const manifestStr = data?.manifest || '';
      yamlContent.value = parseManifest(manifestStr);
      showYamlDialog.value = true;
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '获取 YAML 失败');
    } finally {
      yamlLoading.value = false;
    }
  }

  // Ingress YAML 编辑
  async function handleIngressEdit(row: K8s.Ingress) {
    if (!selectedCluster.value) {
      message.warning('请先选择集群');
      return;
    }

    selectedResource.value = row;
    yamlDialogTitle.value = `编辑 Ingress: ${row.name}`;
    yamlLoading.value = true;

    try {
      const { data, error } = await getK8sIngress(selectedCluster.value, row.namespace, row.name);
      if (error) {
        message.error('获取 YAML 失败');
        return;
      }
      const manifestStr = data?.manifest || '';
      yamlContent.value = parseManifest(manifestStr);
      showYamlDialog.value = true;
    } catch (error: unknown) {
      const err = error as Error;
      message.error(err.message || '获取 YAML 失败');
    } finally {
      yamlLoading.value = false;
    }
  }

  // 删除 Service
  async function handleServiceDelete(row: K8s.Service) {
    try {
      await ElMessageBox.confirm(`确定要删除 Service "${row.name}" 吗？`, '确认删除', {
        type: 'warning'
      });

      await deleteK8sService(selectedCluster.value!, {
        namespace: row.namespace,
        name: row.name
      });
      message.success('删除成功');
      loadServices();
    } catch (error: unknown) {
      if ((error as string) !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  }

  // 删除 Ingress
  async function handleIngressDelete(row: K8s.Ingress) {
    try {
      await ElMessageBox.confirm(`确定要删除 Ingress "${row.name}" 吗？`, '确认删除', {
        type: 'warning'
      });

      await deleteK8sIngress(selectedCluster.value!, {
        namespace: row.namespace,
        name: row.name
      });
      message.success('删除成功');
      loadIngresses();
    } catch (error: unknown) {
      if ((error as string) !== 'cancel') {
        const err = error as Error;
        message.error(err.message || '删除失败');
      }
    }
  }

  // YAML 应用：按当前选中 Tab 分发到 Service/Ingress 的更新 API
  const { handleYamlApply } = useYamlEdit({
    getClusterId: () => selectedCluster.value!,
    getNamespace: () => selectedResource.value?.namespace || selectedNamespace.value,
    updateFn: (clusterId, params) =>
      activeTab.value === 'services' ? updateK8sService(clusterId, params) : updateK8sIngress(clusterId, params),
    reload: () => (activeTab.value === 'services' ? loadServices() : loadIngresses())
  });

  // 操作处理
  function handleCommand(command: string, row: K8s.Service | K8s.Ingress) {
    if (activeTab.value === 'services') {
      // services 表格的操作行，按当前 Tab 收窄类型
      const serviceRow = row as K8s.Service;
      if (command === 'edit') handleServiceEdit(serviceRow);
      else if (command === 'delete') handleServiceDelete(serviceRow);
    } else {
      const ingressRow = row as K8s.Ingress;
      if (command === 'edit') handleIngressEdit(ingressRow);
      else if (command === 'delete') handleIngressDelete(ingressRow);
    }
  }

  // 选择变化
  function handleSelectionChange(selection: K8s.Service[] | K8s.Ingress[]) {
    setSelectedItems(selection);
  }

  // 全选变化
  function handleSelectAll(selection: K8s.Service[] | K8s.Ingress[]) {
    setSelectedItems(selection);
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
  function handleTabChange(tabName: string | number) {
    activeTab.value = String(tabName);
    setSelectedItems([]);
    loadCurrentData();
  }

  onMounted(() => {
    // 集群 → 命名空间 → 首次加载当前 Tab 数据（只查一次，联动由 handle*Change 处理）
    loadAll();
  });
</script>

<template>
  <div class="network-page p-24px">
    <!-- 页面标题 -->
    <div class="mb-24px">
      <h1 class="text-28px text-primary font-bold">网络</h1>
      <p class="text-tertiary mt-8px text-14px">管理 Kubernetes 网络资源</p>
    </div>

    <!-- 头部：集群/命名空间选择和刷新按钮 -->
    <div class="mb-16px flex items-center justify-between gap-12px">
      <div class="flex items-center gap-8px">
        <ElSelect v-model="selectedCluster" placeholder="选择集群" style="width: 200px" @change="handleClusterChange">
          <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
        </ElSelect>
        <ElSelect
          v-model="selectedNamespace"
          placeholder="选择命名空间"
          style="width: 180px"
          @change="handleNamespaceChange"
        >
          <ElOption v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
        </ElSelect>
      </div>
      <ElButton text @click="loadCurrentData">
        <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
      </ElButton>
    </div>

    <!-- Tab 切换 -->
    <ElTabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 服务 -->
      <ElTabPane label="服务" name="services">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="servicesData"
            class="network-table"
            :header-cell-style="{ background: '#f5f7fa', color: '#303133', fontWeight: '600' }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 0' }"
            @selection-change="handleSelectionChange"
            @select-all="handleSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="180" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn prop="type" label="类型" min-width="100" align="left">
              <template #default="{ row }">
                <ElTag :type="getServiceTypeTag(row.type).type">{{ getServiceTypeTag(row.type).text }}</ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="clusterIP" label="集群IP" min-width="130" align="left" />
            <ElTableColumn label="外部IP" min-width="150" align="left">
              <template #default="{ row }">{{ row.externalIP?.join(', ') || '-' }}</template>
            </ElTableColumn>
            <ElTableColumn label="端口" min-width="180" align="left">
              <template #default="{ row }">
                <ElTag v-for="(port, index) in row.ports" :key="index" size="small" class="mr-4px">
                  {{ port.port }}/{{ port.protocol }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="180" fixed="right" align="center" class-name="msre-table-actions">
              <template #default="{ row }">
                <ElButton link type="primary" size="small" @click="goToDetail(row)">详情</ElButton>
                <ElDropdown trigger="click" @command="cmd => handleCommand(cmd, row)">
                  <ElButton link type="primary" size="small" class="table-dropdown-trigger">
                    更多
                    <icon-mdi-chevron-down class="dropdown-icon" />
                  </ElButton>
                  <template #dropdown>
                    <ElDropdownMenu>
                      <ElDropdownItem v-permission="'k8s.resource.update'" command="edit">编辑YAML</ElDropdownItem>
                      <ElDropdownItem v-permission="'k8s.resource.delete'" command="delete">删除</ElDropdownItem>
                    </ElDropdownMenu>
                  </template>
                </ElDropdown>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <div class="batch-info">已选择 {{ selectedServices.length }} 项</div>
            <ElPagination
              v-model:current-page="servicesPagination.page"
              v-model:page-size="servicesPagination.pageSize"
              :total="servicesPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="handlePageChange"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 路由 -->
      <ElTabPane label="路由" name="ingresses">
        <div v-loading="loading" class="table-container">
          <ElTable
            :data="ingressesData"
            class="network-table"
            :header-cell-style="{ background: '#f5f7fa', color: '#303133', fontWeight: '600' }"
            :row-style="{ backgroundColor: 'transparent' }"
            :cell-style="{ backgroundColor: 'transparent', padding: '8px 0' }"
            @selection-change="handleSelectionChange"
            @select-all="handleSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="name" label="名称" min-width="180" align="left">
              <template #default="{ row }">
                <ElButton link type="primary" @click="goToDetail(row)">{{ row.name }}</ElButton>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="namespace" label="命名空间" min-width="120" align="left" />
            <ElTableColumn label="主机" min-width="180" align="left">
              <template #default="{ row }">
                <ElTag v-for="(host, index) in row.hosts" :key="index" size="small" class="mr-4px">
                  {{ host }}
                </ElTag>
                <span v-if="!row.hosts?.length" class="text-gray-400">-</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="地址" min-width="150" align="left">
              <template #default="{ row }">{{ row.addresses?.join(', ') || '-' }}</template>
            </ElTableColumn>
            <ElTableColumn prop="ingressClassName" label="Ingress类" min-width="120" align="left">
              <template #default="{ row }">
                <ElTag v-if="row.ingressClassName" size="small">{{ row.ingressClassName }}</ElTag>
                <span v-else class="text-gray-400">-</span>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="age" label="年龄" min-width="120" align="left" />
            <ElTableColumn label="操作" min-width="180" fixed="right" align="center" class-name="msre-table-actions">
              <template #default="{ row }">
                <ElButton link type="primary" size="small" @click="goToDetail(row)">详情</ElButton>
                <ElDropdown trigger="click" @command="cmd => handleCommand(cmd, row)">
                  <ElButton link type="primary" size="small" class="table-dropdown-trigger">
                    更多
                    <icon-mdi-chevron-down class="dropdown-icon" />
                  </ElButton>
                  <template #dropdown>
                    <ElDropdownMenu>
                      <ElDropdownItem v-permission="'k8s.resource.update'" command="edit">编辑YAML</ElDropdownItem>
                      <ElDropdownItem v-permission="'k8s.resource.delete'" command="delete">删除</ElDropdownItem>
                    </ElDropdownMenu>
                  </template>
                </ElDropdown>
              </template>
            </ElTableColumn>
          </ElTable>

          <div class="bottom-toolbar">
            <div class="batch-info">已选择 {{ selectedIngresses.length }} 项</div>
            <ElPagination
              v-model:current-page="ingressesPagination.page"
              v-model:page-size="ingressesPagination.pageSize"
              :total="ingressesPagination.itemCount"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="handlePageChange"
              @size-change="handlePageSizeChange"
            />
          </div>
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- YAML 编辑器 -->
    <YamlEditor
      v-model="showYamlDialog"
      :title="yamlDialogTitle"
      :yaml="yamlContent"
      :can-edit="true"
      :on-apply="handleYamlApply"
    />
  </div>
</template>

<style scoped>
  /* 高度模型由 .table-page 提供（h-full + flex-col + gap + overflow-hidden） */

  :deep(.el-tabs) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  :deep(.el-tabs__content) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  :deep(.el-tab-pane) {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    height: 100%;
  }

  .table-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
    min-height: 0;
  }

  .network-table {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding-bottom: 60px;
  }

  .bottom-toolbar {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background-color: #fff;
    border-top: 1px solid #ebeef5;
    z-index: 10;

    .batch-info {
      font-size: 14px;
      color: #303133;
      font-weight: 500;
    }

    :deep(.el-pagination) {
      margin: 0;
    }
  }
</style>
