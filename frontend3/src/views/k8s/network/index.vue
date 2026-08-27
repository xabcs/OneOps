<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import {
    ElButton,
    ElDropdown,
    ElDropdownItem,
    ElDropdownMenu,
    ElMessage,
    ElMessageBox,
    ElPagination,
    ElSpace,
    ElTabPane,
    ElTable,
    ElTableColumn,
    ElTabs,
    ElTag
  } from 'element-plus';
  import yaml from 'js-yaml';
  import {
    deleteK8sIngress,
    deleteK8sService,
    fetchK8sClusterNamespaces,
    fetchK8sClusters,
    fetchK8sIngresses,
    fetchK8sServices,
    getK8sIngress,
    getK8sService,
    updateK8sIngress,
    updateK8sService
  } from '@/service/api/k8s';
  import YamlEditor from '@/components/k8s/YamlEditor.vue';

  defineOptions({ name: 'K8sNetwork' });

  const router = useRouter();
  const message = ElMessage;

  const loading = ref(false);
  const activeTab = ref('services');

  // 当前选中的集群和命名空间
  const selectedCluster = ref<number | null>(null);
  const selectedNamespace = ref('default');

  // 可用的命名空间列表
  const namespaces = ref<string[]>([]);

  // 可用的集群列表
  const clusters = ref<K8s.Cluster[]>([]);

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

  // 解析资源manifest为YAML格式
  function parseManifest(manifestStr: string): string {
    if (!manifestStr) return '';
    try {
      const obj = JSON.parse(manifestStr);
      delete obj.managedFields;
      return yaml.dump(obj, {
        indent: 2,
        lineWidth: 120,
        noRefs: true,
        sortKeys: false
      });
    } catch (e) {
      console.error('解析manifest失败:', e);
      return manifestStr;
    }
  }

  // 当前数据
  const currentPagination = computed(() => {
    return activeTab.value === 'services' ? servicesPagination : ingressesPagination;
  });

  // 设置选中项
  function setSelectedItems(items: K8s.Service[] | K8s.Ingress[]) {
    if (activeTab.value === 'services') {
      selectedServices.value = items;
    } else {
      selectedIngresses.value = items;
    }
  }

  // 加载集群列表
  async function loadClusters() {
    const { data, error } = await fetchK8sClusters();
    if (!error) {
      clusters.value = data?.list || [];
      if (clusters.value.length > 0 && !selectedCluster.value) {
        selectedCluster.value = clusters.value[0].id;
      }
    } else {
      console.error('加载集群列表失败:', error);
    }
  }

  // 加载命名空间列表
  async function loadNamespaces() {
    if (!selectedCluster.value) return;
    const { data, error } = await fetchK8sClusterNamespaces(selectedCluster.value);
    if (!error) {
      namespaces.value = (data || []).map((ns: K8s.Namespace) => ns.name);
      if (namespaces.value.length > 0 && !namespaces.value.includes(selectedNamespace.value)) {
        selectedNamespace.value = namespaces.value[0];
      }
    } else {
      console.error('加载命名空间失败:', error);
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

  // YAML 应用
  async function handleYamlApply(yamlStr: string) {
    if (!selectedCluster.value || !selectedResource.value) {
      throw new Error('缺少必要参数');
    }

    try {
      if (activeTab.value === 'services') {
        await updateK8sService(selectedCluster.value, {
          namespace: selectedResource.value.namespace,
          manifest: yaml.load(yamlStr)
        });
        await loadServices();
      } else {
        await updateK8sIngress(selectedCluster.value, {
          namespace: selectedResource.value.namespace,
          manifest: yaml.load(yamlStr)
        });
        await loadIngresses();
      }
    } catch (error: unknown) {
      const err = error as Error;
      throw new Error(err.message || 'YAML 应用失败');
    }
  }

  // 操作处理
  function handleCommand(command: string, row: K8s.Service | K8s.Ingress) {
    if (activeTab.value === 'services') {
      if (command === 'edit') handleServiceEdit(row);
      else if (command === 'delete') handleServiceDelete(row);
    } else if (command === 'edit') handleIngressEdit(row);
    else if (command === 'delete') handleIngressDelete(row);
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
  function handleTabChange(tabName: string) {
    activeTab.value = tabName;
    setSelectedItems([]);
    loadCurrentData();
  }

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
      await loadServices();
    }
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
      <div class="filter-inputs flex items-center gap-8px">
        <ElSelect v-model="selectedCluster" style="width: 200px" @change="loadNamespaces">
          <template #prefix><span class="select-fixed-label">选择集群</span></template>
          <ElOption v-for="cluster in clusters" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
        </ElSelect>
        <ElSelect v-model="selectedNamespace" style="width: 180px" @change="loadCurrentData">
          <template #prefix><span class="select-fixed-label">选择命名空间</span></template>
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
            <ElTableColumn label="操作" min-width="180" fixed="right" align="left">
              <template #default="{ row }">
                <ElSpace>
                  <ElButton size="small" type="primary" @click="goToDetail(row)">详情</ElButton>
                  <ElDropdown trigger="click" @command="cmd => handleCommand(cmd, row)">
                    <span class="dropdown-link">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem v-permission="'k8s.resource.update'" command="edit">编辑YAML</ElDropdownItem>
                        <ElDropdownItem v-permission="'k8s.resource.delete'" command="delete">删除</ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </ElSpace>
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
            <ElTableColumn label="操作" min-width="180" fixed="right" align="left">
              <template #default="{ row }">
                <ElSpace>
                  <ElButton size="small" type="primary" @click="goToDetail(row)">详情</ElButton>
                  <ElDropdown trigger="click" @command="cmd => handleCommand(cmd, row)">
                    <span class="dropdown-link">
                      更多
                      <icon-mdi-chevron-down class="dropdown-icon" />
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem v-permission="'k8s.resource.update'" command="edit">编辑YAML</ElDropdownItem>
                        <ElDropdownItem v-permission="'k8s.resource.delete'" command="delete">删除</ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </ElSpace>
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
  .network-page {
    height: calc(100vh - 80px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

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
    :deep(.el-select .el-select__selection),
    :deep(.el-select .el-select__selected-item),
    :deep(.el-select .el-select__placeholder) {
      display: none;
    }
    .select-fixed-label {
      font-size: 12px;
      color: var(--el-text-color-regular);
      line-height: 30px;
      padding-left: 8px;
    }
    :deep(.el-select.has-value .el-select__prefix) {
      position: static;
      flex: none;
    }
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

    :deep(.el-table__header-wrapper th.el-table__cell) {
      background-color: #f5f7fa !important;
      color: #303133;
      font-weight: 600;
      text-align: left;
      position: sticky;
      top: 0;
      z-index: 1;
    }

    :deep(.el-table__body-wrapper),
    :deep(.el-table__body),
    :deep(.el-table__body tr),
    :deep(.el-table__body td.el-table__cell),
    :deep(.el-table__inner-wrapper),
    :deep(.el-table__fixed),
    :deep(.el-table__fixed-body-wrapper) {
      background-color: transparent !important;
    }

    :deep(.el-table__body tr:hover > td.el-table__cell) {
      background-color: transparent !important;
    }

    .dropdown-link {
      display: inline-flex;
      align-items: center;
      gap: 2px;
      cursor: pointer;
      color: var(--el-color-primary);
      font-size: var(--el-font-size-base);
      user-select: none;
      .dropdown-icon {
        font-size: 14px;
      }
      &:hover {
        opacity: 0.8;
      }
    }
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
