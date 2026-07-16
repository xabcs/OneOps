<template>
  <div class="cmdb-servers-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h2>服务器管理</h2>
      <ElSpace>
        <ElButton
          type="primary"
          :icon="Plus"
          @click="handleCreateServer"
        >
          添加服务器
        </ElButton>
        <ElButton
          :icon="Refresh"
          @click="refreshAll"
        >
          刷新
        </ElButton>
      </ElSpace>
    </div>

    <!-- 主内容区域 -->
    <div class="main-content">
      <!-- 左侧分组树 -->
      <div class="left-panel">
        <GroupTree
          :tree-data="groupTree"
          :loading="groupLoading"
          :selected-group-id="selectedGroupId"
          @node-click="handleGroupClick"
          @update-group="handleUpdateGroup"
          @delete-group="handleDeleteGroup"
        />
      </div>

      <!-- 右侧服务器列表 -->
      <div class="right-panel">
        <!-- 过滤器 -->
        <ServerFilterBar
          ref="filterBarRef"
          @search="handleSearch"
          @reset="handleReset"
          @refresh="refreshServers"
        />

        <!-- 服务器表格 -->
        <ServerTable
          :servers="servers"
          :loading="loading"
          :selected-servers="selectedServers"
          @selection-change="handleSelectionChange"
          @connect="handleConnect"
          @edit="handleEditServer"
          @sync-metrics="handleSyncMetrics"
          @delete="handleDeleteServer"
          @deploy-agent="handleDeployAgent"
          @uninstall-agent="handleUninstallAgent"
          @assign-groups="handleAssignGroups"
          @batch-delete="handleBatchDelete"
        />
      </div>
    </div>

    <!-- 服务器详情抽屉 -->
    <ServerDetailDrawer
      v-if="showDetailDrawer"
      :server-id="currentServerId"
      :server-attributes="serverAttributes"
      :user-credentials="userCredentials"
      :system-credentials="systemCredentials"
      @close="showDetailDrawer = false"
      @save="handleSaveServer"
    />

    <!-- 创建/编辑服务器对话框 -->
    <ServerFormDialog
      v-if="showFormDialog"
      :server-id="editingServerId"
      :mode="dialogMode"
      @close="showFormDialog = false"
      @saved="handleServerSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElButton, ElSpace } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

// 组件导入
import GroupTree from './components/GroupTree.vue'
import ServerTable from './components/ServerTable.vue'
import ServerFilterBar from './components/ServerFilterBar.vue'
import ServerDetailDrawer from './components/ServerDetailDrawer.vue'
import ServerFormDialog from './components/ServerFormDialog.vue'

// Composables导入
import { useServerData } from './composables/useServerData'
import { useGroupTree } from './composables/useGroupTree'

// 类型导入
import type { ServerFilters, TreeNode } from './types/server.types'

defineOptions({
  name: 'CmdbServersPage'
})

const router = useRouter()

// 使用服务器数据管理
const {
  servers,
  serverAttributes,
  userCredentials,
  systemCredentials,
  loading,
  serverStats,
  getServers,
  getServerDetail,
  createServer,
  updateServer,
  deleteServer,
  batchDeployAgent,
  batchUninstallAgent,
  syncServerMetrics
} = useServerData()

// 使用分组树管理
const {
  groupTree,
  groupLoading,
  selectedGroupId,
  filteredGroupTree,
  getGroups,
  handleNodeClick: handleGroupNodeClick,
  createGroup,
  updateGroupName,
  deleteGroup,
  assignServersToGroups
} = useGroupTree()

// UI状态
const selectedServers = ref<CMDB.Server[]>([])
const showDetailDrawer = ref(false)
const showFormDialog = ref(false)
const currentServerId = ref<number>()
const editingServerId = ref<number>()
const dialogMode = ref<'create' | 'edit'>('create')
const filterBarRef = ref()

// 初始化
onMounted(async () => {
  await Promise.all([
    getServers(),
    getGroups()
  ])
})

// 分组操作
const handleGroupClick = (node: TreeNode) => {
  selectedGroupId.value = node.id === 0 ? undefined : node.id
  handleSearch({ groupId: selectedGroupId.value })
}

const handleUpdateGroup = async (nodeId: number, newName: string) => {
  const result = await updateGroupName(nodeId, newName)
  if (result.success) {
    // 显示成功消息
  }
}

const handleDeleteGroup = async (nodeId: number) => {
  await deleteGroup(nodeId)
}

// 过滤操作
const handleSearch = (filters: ServerFilters) => {
  getServers(filters)
}

const handleReset = () => {
  selectedGroupId.value = undefined
  getServers({})
}

const refreshServers = () => {
  getServers()
}

const refreshAll = () => {
  Promise.all([
    getServers(),
    getGroups()
  ])
}

// 服务器操作
const handleSelectionChange = (selections: CMDB.Server[]) => {
  selectedServers.value = selections
}

const handleCreateServer = () => {
  editingServerId.value = undefined
  dialogMode.value = 'create'
  showFormDialog.value = true
}

const handleEditServer = (server: CMDB.Server) => {
  currentServerId.value = server.id
  editingServerId.value = server.id
  dialogMode.value = 'edit'
  getServerDetail(server.id)
  showDetailDrawer.value = true
}

const handleDeleteServer = async (server: CMDB.Server) => {
  const result = await deleteServer(server.id)
  if (result.success) {
    // 显示成功消息
  }
}

const handleConnect = (server: CMDB.Server) => {
  router.push({
    path: '/webterminal/connect',
    query: { serverId: server.id }
  })
}

const handleSyncMetrics = async (server: CMDB.Server) => {
  await syncServerMetrics(server.id)
}

// 批量操作
const handleDeployAgent = async () => {
  const serverIds = selectedServers.value.map(s => s.id)
  await batchDeployAgent(serverIds)
}

const handleUninstallAgent = async () => {
  const serverIds = selectedServers.value.map(s => s.id)
  await batchUninstallAgent(serverIds)
}

const handleAssignGroups = async () => {
  // 实现分配分组逻辑
}

const handleBatchDelete = async () => {
  // 实现批量删除逻辑
}

// 表单保存
const handleSaveServer = async (formData: any) => {
  if (dialogMode.value === 'create') {
    await createServer(formData)
  } else {
    await updateServer(currentServerId.value!, formData)
  }
  showDetailDrawer.value = false
  showFormDialog.value = false
}

const handleServerSaved = () => {
  refreshServers()
  showFormDialog.value = false
}
</script>

<style scoped>
.cmdb-servers-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.main-content {
  display: flex;
  gap: 16px;
  flex: 1;
  min-height: 0;
}

.left-panel {
  width: 280px;
  background: #fff;
  border-radius: 4px;
  padding: 16px;
}

.right-panel {
  flex: 1;
  background: #fff;
  border-radius: 4px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
</style>