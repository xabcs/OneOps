<script setup lang="ts">
  import { onMounted, onUnmounted, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { ElMessageBox, ElNotification } from 'element-plus';
  import { fetchDeleteServer, fetchGetServersByAttributes } from '@/service/api';
  import type { TreeNode } from './types/server.types';

  // 子组件
  import GroupTreePanel from './modules/GroupTreePanel.vue';
  import ServerListPanel from './modules/ServerListPanel.vue';
  import ServerFormDialog from './modules/ServerFormDialog.vue';
  import ServerEditDrawer from './modules/ServerEditDrawer.vue';
  import ServerDetailDrawer from './modules/ServerDetailDrawer.vue';
  import ServerConnectDialog from './modules/ServerConnectDialog.vue';

  // Composables
  import { useGroupTree } from './composables/useGroupTree';
  import { useServerData } from './composables/useServerData';
  import { useServerSearch } from './composables/useServerSearch';
  import { useServerForm } from './composables/useServerForm';
  import { useServerAgent } from './composables/useServerAgent';

  defineOptions({ name: 'CmdbServers' });

  const router = useRouter();

  // ===== Composables =====
  const groupTreeCtx = useGroupTree();
  const serverDataCtx = useServerData();
  const searchCtx = useServerSearch();
  const formCtx = useServerForm();
  const agentCtx = useServerAgent();

  // ===== 属性相关状态 =====
  const serverAttributes = ref<Api.SystemManage.ServerAttribute[]>([]);
  const loadingAttributes = ref(false);

  // ===== 详情抽屉 =====
  const drawerVisible = ref(false);
  const drawerServer = ref<CMDB.Server | null>(null);

  // ===== 连接对话框 =====
  const connectDialogVisible = ref(false);
  const connectingServer = ref<CMDB.Server | null>(null);

  // ===== 辅助函数 =====
  function getUsageColor(value: number = 0): string {
    if (value >= 90) return '#f56c6c';
    if (value >= 70) return '#e6a23c';
    return '#67c23a';
  }

  function getAttributeOptions(key: string): Array<{ label: string; value: string }> {
    const attr = serverDataCtx.attributeDefinitions.value.find(a => a.key === key);
    if (!attr || !attr.options) return [];
    try {
      return JSON.parse(attr.options);
    } catch {
      return [];
    }
  }

  function getEnvDisplayInfo(envValue?: string) {
    const options = getAttributeOptions('env');
    const option = options.find(opt => opt.value === envValue);
    const label = option?.label || envValue || '-';
    let type: 'success' | 'warning' | 'danger' | 'info' = 'info';
    if (envValue === 'prod') type = 'danger';
    else if (envValue === 'test') type = 'warning';
    else if (envValue === 'dev') type = 'success';
    return { label, type };
  }

  function getMaxDiskPartition(row: CMDB.Server) {
    if (!row.diskPartitions || row.diskPartitions.length === 0) return { usage: row.diskUsage || 0, mount: '/' };
    return row.diskPartitions.reduce((max, partition) => (partition.usage > max.usage ? partition : max), {
      usage: 0,
      mount: '/'
    });
  }

  function formatDiskPartitions(row: CMDB.Server) {
    if (!row.diskPartitions || row.diskPartitions.length === 0) return '暂无分区信息';
    return row.diskPartitions
      .sort((a, b) => b.usage - a.usage)
      .map(p => `${p.mount}: ${p.usage}%`)
      .join('\n');
  }

  function formatDiskPartitionsForDrawer(server: CMDB.Server | null) {
    if (!server || !server.diskPartitions || server.diskPartitions.length === 0) return [];
    return server.diskPartitions
      .slice()
      .sort((a, b) => b.usage - a.usage)
      .map(p => ({ mount: p.mount, usage: p.usage }));
  }

  function formatTime(timeStr: string): string {
    if (!timeStr) return '-';
    const date = new Date(timeStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    if (minutes < 1) return '刚刚';
    if (minutes < 60) return `${minutes} 分钟前`;
    if (hours < 24) return `${hours} 小时前`;
    if (days < 7) return `${days} 天前`;
    return date.toLocaleDateString('zh-CN');
  }

  function getUnifiedAttributes() {
    return serverDataCtx.attributeDefinitions.value;
  }

  function parseAttributeOptions(optionsStr: string) {
    if (!optionsStr) return [];
    try {
      return JSON.parse(optionsStr);
    } catch {
      return [];
    }
  }

  function getAttributeValue(attributeId: number): string {
    const attr = serverAttributes.value.find(a => a.attributeId === attributeId);
    return attr?.attributeValue || '';
  }

  function setAttributeValue(attributeId: number, value: string | number | boolean): void {
    let attr = serverAttributes.value.find(a => a.attributeId === attributeId);
    if (!attr) {
      const definition = serverDataCtx.attributeDefinitions.value.find(d => d.id === attributeId);
      attr = {
        id: 0,
        serverId: 0,
        attributeId,
        attributeKey: definition?.key || '',
        attributeValue: String(value),
        valueType: 'string',
        category: definition?.category || '',
        createdAt: '',
        updatedAt: ''
      };
      serverAttributes.value.push(attr);
    }
    if (attr) attr.attributeValue = String(value);
  }

  function getAttributeMultiValue(attributeId: number) {
    const attr = serverAttributes.value.find(a => a.attributeId === attributeId);
    if (!attr) return [];
    try {
      return JSON.parse(attr.attributeValue || '[]');
    } catch {
      return [];
    }
  }

  function setAttributeMultiValue(attributeId: number, values: string[]): void {
    let attr = serverAttributes.value.find(a => a.attributeId === attributeId);
    if (!attr) {
      const definition = serverDataCtx.attributeDefinitions.value.find(d => d.id === attributeId);
      attr = {
        id: 0,
        serverId: 0,
        attributeId,
        attributeKey: definition?.key || '',
        attributeValue: '[]',
        valueType: 'string',
        category: definition?.category || '',
        createdAt: '',
        updatedAt: ''
      };
      serverAttributes.value.push(attr);
    }
    if (attr) attr.attributeValue = JSON.stringify(values);
  }

  // ===== 搜索相关 =====
  function buildSearchParams() {
    return {
      searchType: searchCtx.searchType.value,
      searchKeyword: searchCtx.searchKeyword.value,
      groupId: groupTreeCtx.selectedGroupId.value,
      ungrouped: groupTreeCtx.selectedUngrouped.value,
      tagId: tagFilterId.value
    };
  }

  function refreshServers() {
    serverDataCtx.getServers(buildSearchParams());
  }

  function handleSearch() {
    serverDataCtx.pagination.page = 1;
    refreshServers();
  }

  // ===== 按属性筛选 =====
  const attrFilterActive = ref(false);

  // ===== 按标签筛选 =====
  const tagFilterId = ref<number | undefined>(undefined);

  async function handleAttributeFilter(filters: Record<string, string>) {
    serverDataCtx.pagination.page = 1;
    if (Object.keys(filters).length === 0) {
      attrFilterActive.value = false;
      refreshServers();
      return;
    }
    attrFilterActive.value = true;
    serverDataCtx.loading.value = true;
    try {
      const res = await fetchGetServersByAttributes(filters, serverDataCtx.pagination.page, serverDataCtx.pagination.pageSize);
      if (res.data) {
        serverDataCtx.tableData.value = res.data.list || [];
        serverDataCtx.total.value = res.data.total || 0;
      }
    } catch {
      /* 全局拦截器处理 */
    } finally {
      serverDataCtx.loading.value = false;
    }
  }

  function handleTagFilter(tagId: number | undefined) {
    tagFilterId.value = tagId;
    serverDataCtx.pagination.page = 1;
    refreshServers();
  }

  // ===== 分组树事件转发 =====
  function onNodeClick(data: TreeNode) {
    groupTreeCtx.handleNodeClick(data);
    serverDataCtx.pagination.page = 1;
    serverDataCtx.selectedIds.value = [];
    refreshServers();
  }

  function onRefreshGroups() {
    groupTreeCtx.getGroups();
  }

  function onAddServer() {
    groupTreeCtx.handleAddServer();
    // 确保分组数据已加载
    if (groupTreeCtx.groupTree.value.length === 0) {
      groupTreeCtx.getGroups().then(() => {
        setTimeout(() => onAdd(), 100);
      });
    } else {
      onAdd();
    }
  }

  // ===== 创建/编辑 =====
  function onAdd() {
    formCtx.openAddDialog(groupTreeCtx.selectedGroupId.value);
    serverDataCtx.loadAttributeDefinitions();
    serverAttributes.value = [];
  }

  function onEdit(row: CMDB.Server) {
    formCtx.openEditDrawer(row);
    serverDataCtx.loadAttributeDefinitions();
    if (row.cabinet?.roomId) serverDataCtx.getCabinets(row.cabinet.roomId);
    // 加载主机属性
    loadingAttributes.value = true;
    serverDataCtx.loadServerAttributes(row.id).then(data => {
      serverAttributes.value = data;
      loadingAttributes.value = false;
    });
  }

  // ===== 创建/编辑提交（子组件验证通过后触发）=====
  async function onSubmitted() {

    const success = await formCtx.submitForm(async (serverId: number) => {
      await serverDataCtx.saveServerAttributes(serverId, serverAttributes.value);
    });
    if (success) {
      refreshServers();
      groupTreeCtx.getGroups();
    }
  }

  // ===== 删除 =====
  function handleDelete(row: CMDB.Server) {
    ElMessageBox.confirm(`确定要删除主机 "${row.hostname}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        try {
          await fetchDeleteServer(row.id);
          ElNotification.success('删除成功');
          refreshServers();
          groupTreeCtx.getGroups();
        } catch (error) {
          console.error('删除失败:', error);
          ElNotification.error('删除失败');
        }
      })
      .catch(() => {});
  }

  function handleBatchDelete() {
    if (serverDataCtx.selectedIds.value.length === 0) {
      ElNotification.warning('请先选择要删除的主机');
      return;
    }
    ElMessageBox.confirm(`确定要删除选中的 ${serverDataCtx.selectedIds.value.length} 台主机吗？`, '批量删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        try {
          await Promise.all(serverDataCtx.selectedIds.value.map(id => fetchDeleteServer(id)));
          ElNotification.success('删除成功');
          serverDataCtx.selectedIds.value = [];
          refreshServers();
          groupTreeCtx.getGroups();
        } catch (error) {
          console.error('批量删除失败:', error);
          ElNotification.error('批量删除失败');
        }
      })
      .catch(() => {});
  }

  // ===== 连接 =====
  function handleConnect(row: CMDB.Server) {
    const params = new URLSearchParams({
      serverId: row.id.toString(),
      hostname: row.hostname,
      ip: row.ip,
      sshUser: row.sshCredential?.username || 'root',
      env: row.attributeValues?.env || 'unknown',
      agentStatus: row.agentStatus || 'unknown'
    });
    window.open(`/webterminal?${params.toString()}`, '_blank');
  }

  async function confirmConnect() {
    if (!connectingServer.value) return;
    try {
      const userCredentialIds =
        connectingServer.value.credentials?.filter(c => c.credentialType === 'user').map(c => c.id) || [];
      const defaultCredentialId = userCredentialIds[0] || connectingServer.value.sshCredentialId;
      const params = new URLSearchParams({
        serverId: connectingServer.value.id.toString(),
        serverName: connectingServer.value.hostname,
        serverIp: connectingServer.value.ip,
        serverEnv: connectingServer.value.attributeValues?.env || 'unknown'
      });
      if (defaultCredentialId) params.append('credentialId', defaultCredentialId.toString());
      window.open(`/webterminal?${params.toString()}`, '_blank');
      connectDialogVisible.value = false;
      connectingServer.value = null;
      ElNotification({
        title: '连接成功',
        message: `正在连接到 ${connectingServer.value?.hostname}...`,
        type: 'success',
        duration: 2000
      });
    } catch (error) {
      ElNotification({ title: '连接失败', message: '连接过程中发生错误', type: 'error', duration: 3000 });
    }
  }

  // ===== 详情 =====
  function handleViewDetail(row: CMDB.Server) {
    router.push({ path: '/cmdb/servers/detail', query: { id: row.id.toString() } });
  }

  function handleViewMonitoring(row: CMDB.Server) {
    // 跳转监控详情
  }

  // ===== 更多操作 =====
  function handleMoreAction(cmd: string, row: CMDB.Server) {
    if (cmd === 'delete') handleDelete(row);
    if (cmd === 'edit') onEdit(row);
    if (cmd === 'sync-metrics') {
      serverDataCtx.handleSyncMetrics(row);
      setTimeout(() => refreshServers(), 3000);
    }
    if (cmd === 'monitoring') handleViewMonitoring(row);
    if (cmd === 'detail') handleViewDetail(row);
    if (cmd === 'agent-deploy') agentCtx.handleAgentDeploy(row, serverDataCtx.tableData.value);
    if (cmd === 'agent-restart') agentCtx.handleAgentRestart(row, serverDataCtx.tableData.value);
    if (cmd === 'agent-uninstall') agentCtx.handleAgentUninstall(row, serverDataCtx.tableData.value);
  }

  // ===== 批量操作 =====
  function handleBatchCommand(cmd: string) {
    if (cmd === 'test-connection')
      agentCtx.handleBatchTestConnection(serverDataCtx.selectedIds.value, serverDataCtx.tableData.value);
    if (cmd === 'batch-deploy')
      agentCtx.handleBatchDeploy(serverDataCtx.selectedIds.value, serverDataCtx.tableData.value, (id, status) =>
        agentCtx.pollAgentStatus(id, status, serverDataCtx.tableData.value)
      );
    if (cmd === 'batch-uninstall')
      agentCtx.handleBatchUninstall(serverDataCtx.selectedIds.value, serverDataCtx.tableData.value, (id, status) =>
        agentCtx.pollAgentStatus(id, status, serverDataCtx.tableData.value)
      );
    if (cmd === 'batch-delete') handleBatchDelete();
  }

  function handleCreateCommand(cmd: string) {
    if (cmd === 'add') onAdd();
    if (cmd === 'import') ElNotification.info('批量导入功能开发中...');
  }

  // ===== 初始化 =====
  onMounted(() => {
    groupTreeCtx.getGroups();
    serverDataCtx.getSSHCredentials();
    refreshServers();
    serverDataCtx.getServerRooms();
    serverDataCtx.getServerTags();
    serverDataCtx.getSystemOptions();
    serverDataCtx.getBusinessUnits();
    document.addEventListener('click', groupTreeCtx.handleGlobalClick);
  });

  onUnmounted(() => {
    document.removeEventListener('click', groupTreeCtx.handleGlobalClick);
  });
</script>

<template>
  <el-container style="height: 100%">
    <!-- 左侧分组树 -->
    <el-aside width="220px" style="overflow: hidden;">
      <GroupTreePanel
        :group-loading="groupTreeCtx.groupLoading.value"
        :filtered-group-tree="groupTreeCtx.filteredGroupTree.value"
        :group-search-keyword="groupTreeCtx.groupSearchKeyword.value"
        :selected-group-id="groupTreeCtx.selectedGroupId.value"
        :selected-ungrouped="groupTreeCtx.selectedUngrouped.value"
        :editing-node-id="groupTreeCtx.editingNodeId.value"
        :editing-node-name="groupTreeCtx.editingNodeName.value"
        :context-menu-visible="groupTreeCtx.contextMenuVisible.value"
        :context-menu-position="groupTreeCtx.contextMenuPosition.value"
        :context-menu-node-id="groupTreeCtx.contextMenuNodeId.value"
        :group-dialog-visible="groupTreeCtx.groupDialogVisible.value"
        :group-form-data="groupTreeCtx.groupFormData.value"
        :group-tree="groupTreeCtx.groupTree.value"
        :get-node-class="groupTreeCtx.getNodeClass"
        @update:group-search-keyword="groupTreeCtx.groupSearchKeyword.value = $event"
        @update:editing-node-name="groupTreeCtx.editingNodeName.value = $event"
        @node-click="onNodeClick"
        @node-contextmenu="groupTreeCtx.handleNodeContextMenu"
        @context-menu-close="groupTreeCtx.handleContextMenuClose"
        @add-root-group="groupTreeCtx.handleAddRootGroup"
        @refresh-groups="onRefreshGroups"
        @edit-keydown="groupTreeCtx.handleEditKeydown"
        @save-edit-group="groupTreeCtx.saveEditGroup"
        @add-group="groupTreeCtx.handleAddGroup"
        @add-server="onAddServer"
        @edit-group="groupTreeCtx.handleEditGroup"
        @delete-group="groupTreeCtx.handleDeleteGroup"
        @update:group-dialog-visible="groupTreeCtx.groupDialogVisible.value = $event"
        @save-group="groupTreeCtx.handleSaveGroup"
      />
    </el-aside>

    <!-- 右侧主机列表 -->
    <el-main style="padding: 0; margin-left: 8px;">
      <ServerListPanel
        :loading="serverDataCtx.loading.value"
        :table-data="serverDataCtx.tableData.value"
        :total="serverDataCtx.total.value"
        :selected-ids="serverDataCtx.selectedIds.value"
        :pagination="serverDataCtx.pagination"
        :search-type="searchCtx.searchType.value"
        :search-keyword="searchCtx.searchKeyword.value"
        :attribute-definitions="serverDataCtx.attributeDefinitions.value"
        :server-tags="serverDataCtx.serverTags.value"
        :get-env-display-info="getEnvDisplayInfo"
        :get-usage-color="getUsageColor"
        :get-max-disk-partition="getMaxDiskPartition"
        :format-disk-partitions="formatDiskPartitions"
        @update:search-type="searchCtx.searchType.value = $event"
        @update:search-keyword="searchCtx.searchKeyword.value = $event"
        @search="handleSearch"
        @refresh="refreshServers"
        @selection-change="serverDataCtx.handleSelectionChange"
        @select-all="serverDataCtx.handleSelectAll"
        @page-change="
          serverDataCtx.handlePageChange($event);
          refreshServers();
        "
        @page-size-change="
          serverDataCtx.handlePageSizeChange($event);
          refreshServers();
        "
        @view-detail="handleViewDetail"
        @connect="handleConnect"
        @more-action="handleMoreAction"
        @batch-command="handleBatchCommand"
        @create-command="handleCreateCommand"
        @attribute-filter="handleAttributeFilter"
        @tag-filter="handleTagFilter"
      />

      <!-- 创建主机对话框 -->
      <ServerFormDialog
        v-model:visible="formCtx.dialogVisible.value"
        v-model:active-collapse="formCtx.activeCollapse.value"
        :server-form="formCtx.serverForm"
        :server-form-rules="formCtx.serverFormRules"
        :submit-error="formCtx.submitError.value"
        :user-credentials="serverDataCtx.userCredentials.value"
        :system-credentials="serverDataCtx.systemCredentials.value"
        :group-tree-for-select="groupTreeCtx.groupTreeForSelect.value"
        :loading-attributes="loadingAttributes"
        :attribute-definitions="serverDataCtx.attributeDefinitions.value"
        :server-attributes="serverAttributes"
        :get-attribute-value="getAttributeValue"
        :set-attribute-value="setAttributeValue"
        :get-attribute-multi-value="getAttributeMultiValue"
        :set-attribute-multi-value="setAttributeMultiValue"
        :get-attribute-options="getAttributeOptions"
        :parse-attribute-options="parseAttributeOptions"
        :get-unified-attributes="getUnifiedAttributes"
        :server-tags="serverDataCtx.serverTags.value"
        @submitted="onSubmitted"
      />

      <!-- 编辑主机抽屉 -->
      <ServerEditDrawer
        v-model:visible="formCtx.editDrawerVisible.value"
        v-model:active-tab="formCtx.editDrawerActiveTab.value"
        :server-form="formCtx.serverForm"
        :server-form-rules="formCtx.serverFormRules"
        :submit-error="formCtx.submitError.value"
        :server-type="formCtx.serverType.value"
        :cloud-form="formCtx.cloudForm"
        :user-credentials="serverDataCtx.userCredentials.value"
        :system-credentials="serverDataCtx.systemCredentials.value"
        :server-tags="serverDataCtx.serverTags.value"
        :server-rooms="serverDataCtx.serverRooms.value"
        :cabinets="serverDataCtx.cabinets.value"
        :business-units="serverDataCtx.businessUnits.value"
        :group-tree-for-select="groupTreeCtx.groupTreeForSelect.value"
        :loading-attributes="loadingAttributes"
        :attribute-definitions="serverDataCtx.attributeDefinitions.value"
        :get-attribute-options="getAttributeOptions"
        :get-unified-attributes="getUnifiedAttributes"
        :get-attribute-value="getAttributeValue"
        :set-attribute-value="setAttributeValue"
        :get-attribute-multi-value="getAttributeMultiValue"
        :set-attribute-multi-value="setAttributeMultiValue"
        :parse-attribute-options="parseAttributeOptions"
        :handle-room-change="serverDataCtx.handleRoomChange"
        @submitted="onSubmitted"
      />

      <!-- 主机详情抽屉 -->
      <ServerDetailDrawer
        :visible="drawerVisible"
        :server="drawerServer"
        :get-usage-color="getUsageColor"
        :get-env-display-info="getEnvDisplayInfo"
        :format-time="formatTime"
        :format-disk-partitions-for-drawer="formatDiskPartitionsForDrawer"
        :handle-connect="handleConnect"
        @update:visible="drawerVisible = $event"
      />

      <!-- 连接确认对话框 -->
      <ServerConnectDialog
        :visible="connectDialogVisible"
        :connecting-server="connectingServer"
        @update:visible="connectDialogVisible = $event"
        @confirm="confirmConnect"
        @cancel="connectingServer = null"
      />
    </el-main>
  </el-container>
</template>
