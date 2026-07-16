<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { ElButton, ElInput, ElMessage, ElTooltip, ElTree } from 'element-plus';
import { Icon } from '@iconify/vue';
import { fetchGetAssetTree } from '@/service/api';

interface TreeNode {
  id: number | string;
  name: string;
  children?: TreeNode[];
  serverCount?: number;
  isLeaf?: boolean;
  server?: any;
}

interface Props {
  currentSessions: number[];
}

interface Emits {
  (e: 'connect', server: any): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const treeData = ref<TreeNode[]>([]);
const loading = ref(false);
const searchKeyword = ref('');
const searchExpanded = ref(false);
const treeRef = ref<InstanceType<typeof ElTree> | null>(null);
const searchInputRef = ref<InstanceType<typeof ElInput> | null>(null);

// 右键菜单
const contextMenuVisible = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
const contextMenuServer = ref<any>(null);

// 右键菜单项
const contextMenuItems = [
  {
    icon: 'vscode-icons:default-terminal',
    label: '连接',
    action: () => {
      if (contextMenuServer.value) {
        handleConnect(contextMenuServer.value);
      }
    }
  },
  {
    icon: 'lucide:external-link',
    label: '新窗口打开',
    action: () => {
      if (contextMenuServer.value) {
        window.open(`/cmdb/terminal/${contextMenuServer.value.id}`, '_blank');
      }
    }
  },
  {
    icon: 'lucide:star',
    label: '收藏',
    action: () => {
      if (contextMenuServer.value) {
        ElMessage.info(`已收藏 ${contextMenuServer.value.hostname}`);
      }
    }
  }
];

// 过滤后的树数据
const filteredTreeData = computed(() => {
  if (!searchKeyword.value) return treeData.value;
  return filterTree(treeData.value, searchKeyword.value.toLowerCase());
});

// 递归过滤树
function filterTree(nodes: TreeNode[], keyword: string): TreeNode[] {
  const result: TreeNode[] = [];
  for (const node of nodes) {
    if (node.isLeaf && node.server) {
      if (
        node.name.toLowerCase().includes(keyword) ||
        (node.server.ip && node.server.ip.toLowerCase().includes(keyword))
      ) {
        result.push(node);
      }
    } else if (node.children) {
      const filteredChildren = filterTree(node.children, keyword);
      if (filteredChildren.length > 0 || node.name.toLowerCase().includes(keyword)) {
        result.push({
          ...node,
          children: filteredChildren
        });
      }
    } else if (node.name.toLowerCase().includes(keyword)) {
      result.push(node);
    }
  }
  return result;
}

// 构建完整的树（使用优化后的API）
async function loadTreeData() {
  loading.value = true;
  try {
    // 一次性获取所有数据：分组树 + 每个分组的服务器 + 无分组服务器
    const { data } = await fetchGetAssetTree();

    const groups = data.groups || [];
    const ungroupedServers = data.ungroupedServers || [];

    // 构建树节点
    const buildNode = (group: any): TreeNode => {
      const childGroups = (group.children || []).map((child: any) => buildNode(child));
      const groupServers = (group.servers || []).map((server: any) => ({
        id: `server-${server.id}` as unknown as number,
        name: `${server.hostname} (${server.ip})`,
        children: undefined,
        serverCount: undefined,
        isLeaf: true,
        server
      }));

      const node: TreeNode = {
        ...group,
        children: []
      };

      // 先添加子分组
      childGroups.forEach((child: TreeNode) => {
        node.children!.push(child);
      });

      // 再添加该分组的服务器
      groupServers.forEach((serverNode: TreeNode) => {
        node.children!.push(serverNode);
      });

      return node;
    };

    const result = groups.map((g: any) => buildNode(g));

    // 添加"Default"节点（如果有未分组的服务器）
    if (ungroupedServers.length > 0) {
      result.unshift({
        id: 'ungrouped' as unknown as number,
        name: 'Default',
        children: ungroupedServers.map((server: any) => ({
          id: `server-${server.id}` as unknown as number,
          name: `${server.hostname} (${server.ip})`,
          children: undefined,
          serverCount: undefined,
          isLeaf: true,
          server
        })),
        serverCount: ungroupedServers.length
      });
    }

    treeData.value = result;
  } catch (error) {
    console.error('加载资产树失败:', error);
    ElMessage.error('加载资产树失败');
  } finally {
    loading.value = false;
  }
}

// 节点点击
function handleNodeClick(node: TreeNode) {
  console.log('节点点击:', node);
}

// 双击节点
function handleNodeDblClick(node: any) {
  const data = (node.data || node) as TreeNode;
  if (data.isLeaf && data.server) {
    handleConnect(data.server);
  }
}

// 连接服务器
function handleConnect(server: any) {
  hideContextMenu();
  emit('connect', server);
}

// 处理右键菜单
function handleContextMenu(data: TreeNode, event: MouseEvent) {
  if (data.isLeaf && data.server) {
    event.preventDefault();
    event.stopPropagation();
    contextMenuServer.value = data.server;
    contextMenuPosition.value = { x: event.clientX, y: event.clientY };
    contextMenuVisible.value = true;
  }
}

// 隐藏右键菜单
function hideContextMenu() {
  contextMenuVisible.value = false;
  contextMenuServer.value = null;
}

// 点击其他地方关闭右键菜单和搜索框
function handleClickOutside(event: MouseEvent) {
  hideContextMenu();
  // 检查点击是否在搜索区域外
  const target = event.target as HTMLElement;
  const searchArea = document.querySelector('.wb-search-container');
  if (searchArea && !searchArea.contains(target) && searchExpanded.value) {
    searchExpanded.value = false;
    searchKeyword.value = '';
  }
}

// 刷新
function handleRefresh() {
  loadTreeData();
}

// 切换搜索框展开状态
function toggleSearch(event?: MouseEvent) {
  if (event) {
    event.stopPropagation();
  }
  console.log('toggleSearch called, current searchExpanded:', searchExpanded.value);
  searchExpanded.value = !searchExpanded.value;
  console.log('searchExpanded after toggle:', searchExpanded.value);

  if (searchExpanded.value) {
    // 自动聚焦搜索框
    nextTick(() => {
      const input = document.querySelector('.wb-search-input-inner input') as HTMLInputElement;
      console.log('Input element:', input);
      input?.focus();

      // 调试：检查输入框样式
      nextTick(() => {
        const wrapper = document.querySelector('.wb-search-input-inner .el-input__wrapper');
        console.log('Input wrapper:', wrapper);
        if (wrapper) {
          const styles = getComputedStyle(wrapper);
          console.log('Wrapper border-radius:', styles.borderRadius);
          console.log('Wrapper height:', styles.height);
          console.log('Wrapper background-color:', styles.backgroundColor);
        }
      });
    });
  } else {
    searchKeyword.value = '';
  }
}

// 搜索框点击时阻止冒泡
function handleSearchClick(event: MouseEvent) {
  event.stopPropagation();
}

// 获取节点CSS类
function getNodeClass(node: TreeNode): string {
  const classes: string[] = [];
  if (node.isLeaf && node.server) {
    if (props.currentSessions.includes(Number(node.server.id))) {
      classes.push('is-connected');
    }
    if (node.server.agentStatus === 'offline' || node.server.agentStatus === 'uninstalled') {
      classes.push('is-offline');
    }
  }
  return classes.join(' ');
}

// 获取节点图标
function getNodeIcon(node: TreeNode): string {
  if (node.isLeaf) {
    return 'lucide:server';
  }
  return 'lucide:folder';
}

// 监听搜索关键词变化，自动展开
watch(searchKeyword, newVal => {
  if (newVal && !searchExpanded.value) {
    searchExpanded.value = true;
  }
});

onMounted(() => {
  loadTreeData();
  // 添加全局点击监听
  document.addEventListener('click', handleClickOutside);
});

// 清理监听
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<template>
  <div class="wb-sidebar" @click="handleClickOutside">
    <!-- 侧边栏头部 -->
    <div class="wb-sidebar-header">
      <div class="wb-header-left">
        <Transition name="wb-search-expand">
          <span v-if="!searchExpanded" key="title" class="wb-sidebar-title">主机资产</span>
          <div v-else key="search" class="wb-search-container">
            <div class="wb-search-input-wrapper">
              <Icon icon="lucide:search" class="wb-search-icon-inline" />
              <ElInput
                ref="searchInputRef"
                v-model="searchKeyword"
                placeholder="搜索..."
                size="small"
                clearable
                class="wb-search-input-inner"
                @click="handleSearchClick"
              />
            </div>
          </div>
        </Transition>
      </div>
      <div class="wb-sidebar-actions">
        <ElTooltip v-if="!searchExpanded" content="搜索" placement="bottom">
          <ElButton size="small" link @click="toggleSearch">
            <Icon icon="lucide:search" class="wb-icon" />
          </ElButton>
        </ElTooltip>
        <ElTooltip content="刷新" placement="bottom">
          <ElButton size="small" link @click="handleRefresh">
            <Icon
              :icon="loading ? 'lucide:loader-2' : 'lucide:refresh-cw'"
              :class="{ 'wb-spinning': loading }"
              class="wb-icon"
            />
          </ElButton>
        </ElTooltip>
      </div>
    </div>

    <!-- 树内容区 -->
    <div class="wb-sidebar-content">
      <ElTree
        ref="treeRef"
        :data="filteredTreeData"
        :props="{ label: 'name', children: 'children' }"
        :expand-on-click-node="false"
        node-key="id"
        :default-expand-all="false"
        class="wb-tree"
        @node-click="handleNodeClick"
        @node-dblclick="handleNodeDblClick"
      >
        <template #default="{ node, data }">
          <div
            class="wb-tree-node"
            :class="[getNodeClass(data), { selected: node.selected }]"
            @dblclick="handleNodeDblClick({ data })"
            @contextmenu.prevent="handleContextMenu(data, $event)"
          >
            <Icon :icon="getNodeIcon(data)" class="wb-tree-node-icon" />
            <span class="wb-tree-node-label">{{ node.label }}</span>
            <span v-if="data.serverCount !== undefined && !data.isLeaf" class="wb-tree-node-count">
              {{ data.serverCount }}
            </span>
            <span
              v-if="data.isLeaf && data.server && currentSessions.includes(Number(data.server.id))"
              class="wb-tag wb-tag-success"
            >
              已连接
            </span>
          </div>
        </template>
      </ElTree>

      <!-- 加载状态 -->
      <div v-if="loading" class="wb-loading-state">
        <Icon icon="lucide:loader-2" class="wb-loading-icon wb-spinning" />
        <span>加载中...</span>
      </div>

      <!-- 空状态 -->
      <div v-else-if="treeData.length === 0" class="wb-empty-state">
        <Icon icon="lucide:folder-open" class="wb-empty-icon" />
        <p>{{ searchKeyword ? '没有找到匹配的主机' : '暂无主机' }}</p>
      </div>
    </div>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="wb-dropdown-menu"
        :style="{
          left: contextMenuPosition.x + 'px',
          top: contextMenuPosition.y + 'px'
        }"
        @click="handleClickOutside"
      >
        <div v-for="item in contextMenuItems" :key="item.label" class="wb-dropdown-item" @click.stop="item.action()">
          <Icon :icon="item.icon" class="wb-dropdown-item-icon" />
          <span>{{ item.label }}</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style lang="scss">
.wb-sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #252526;
}

.wb-sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  height: 35px;
  flex-shrink: 0;
  gap: 8px;
}

.wb-header-left {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow: visible;
  position: relative;
}

.wb-sidebar-title {
  font-size: 12px;
  font-weight: 500;
  color: #cccccc;
  white-space: nowrap;
  flex-shrink: 0;
}

/* 搜索框容器 */
.wb-search-container {
  flex: 1;
  max-width: 180px;
  display: flex;
  align-items: center;
}

.wb-search-input-wrapper {
  display: flex;
  align-items: center;
  width: 100%;
  height: 24px;
  background: #3c3c3c;
  border-radius: 12px;
  padding: 0 8px;
  gap: 6px;
}

.wb-search-icon-inline {
  width: 16px;
  height: 16px;
  color: #858585;
  flex-shrink: 0;
}

.wb-search-input-inner {
  flex: 1;
  min-width: 0;
}

/* 强制覆盖 Element Plus 输入框样式 */
.wb-search-input-inner {
  --el-input-bg-color: transparent !important;
  --el-input-border-color: transparent !important;
  --el-input-hover-border-color: transparent !important;
  --el-input-focus-border-color: transparent !important;
  --el-input-clear-bg-color: transparent !important;
}

.wb-search-input-inner :deep(.el-input__wrapper) {
  height: 24px !important;
  background: transparent !important;
  border: none !important;
  box-shadow: none !important;
  padding: 0 !important;
}

.wb-search-input-inner :deep(.el-input__inner) {
  height: 24px !important;
  line-height: 24px !important;
  background: transparent !important;
  border: none !important;
  padding: 0 4px !important;
  color: #cccccc;
  font-size: 12px;
}

.wb-search-input-inner :deep(.el-input__wrapper::before),
.wb-search-input-inner :deep(.el-input__wrapper::after),
.wb-search-input-inner :deep(.el-input__wrapper:hover),
.wb-search-input-inner :deep(.el-input__wrapper.is-focus) {
  border: none !important;
  box-shadow: none !important;
  background: transparent !important;
}

.wb-search-input-inner :deep(.el-input__prefix),
.wb-search-input-inner :deep(.el-input__suffix) {
  display: none !important;
}

.wb-search-input-inner :deep(.el-input__clear) {
  background: transparent !important;
  color: #858585 !important;
}

.wb-search-input-inner :deep(.el-input__clear:hover) {
  color: #cccccc !important;
}

/* 搜索框展开动画 */
.wb-search-expand-enter-active,
.wb-search-expand-leave-active {
  transition: all 0.2s ease;
}

.wb-search-expand-enter-from {
  opacity: 0;
  width: 0;
}

.wb-search-expand-leave-to {
  opacity: 0;
  width: 0;
}

.wb-search-expand-enter-to,
.wb-search-expand-leave-from {
  opacity: 1;
  width: 180px;
}

.wb-sidebar-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.wb-sidebar-actions :deep(.el-button) {
  background: transparent !important;
  border: none !important;
}

.wb-sidebar-actions :deep(.el-button:hover) {
  background: rgba(0, 0, 0, 0.2) !important;
}

.wb-sidebar-actions :deep(.el-button:hover .iconify) {
  color: #aaaaaa !important;
}

.wb-sidebar-actions :deep(.el-button.is-link:hover) {
  background-color: rgba(0, 0, 0, 0.2) !important;
}

.wb-icon {
  width: 16px;
  height: 16px;
  color: #858585;
  transition: color 0.15s;
}

.wb-sidebar-actions :deep(.el-button:hover) .wb-icon {
  color: #aaaaaa;
}

.wb-spinning {
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

.wb-sidebar-content {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.wb-tree {
  background: transparent;
}

.wb-tree :deep(.el-tree-node__content) {
  height: 24px;
  background: transparent;
  border-radius: 2px;
  margin: 1px 4px;
}

.wb-tree :deep(.el-tree-node__content:hover) {
  background: rgba(0, 0, 0, 0.2) !important;
}

.wb-tree :deep(.el-tree-node__expand-icon) {
  color: #6e6e6e;
  font-size: 12px;
  width: 16px;
}

.wb-tree-node {
  display: flex;
  align-items: center;
  cursor: pointer;
  width: 100%;
}

.wb-tree-node-icon {
  width: 14px;
  height: 14px;
  color: #858585;
  flex-shrink: 0;
  margin-right: 6px;
}

.wb-tree-node-label {
  font-size: 12px;
  color: #cccccc;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 0 1 auto;
}

.wb-tree-node-count {
  font-size: 10px;
  color: #6e6e6e;
  flex-shrink: 0;
  margin-left: 0;
}

.is-connected {
  background: rgba(78, 201, 176, 0.1);
}

.is-connected .wb-tree-node-icon {
  color: #4ec9b0;
}

.is-offline {
  opacity: 0.5;
}

.wb-loading-state,
.wb-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 12px;
  color: #858585;
  font-size: 12px;
  gap: 8px;
}

.wb-loading-icon,
.wb-empty-icon {
  width: 24px;
  height: 24px;
  opacity: 0.5;
}

.wb-empty-state p {
  margin: 0;
}

/* 右键菜单 */
.wb-dropdown-menu {
  position: fixed;
  z-index: 9999;
  min-width: 140px;
  padding: 4px 0;
  background: #252526;
  border: 1px solid #454545;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
}

.wb-dropdown-item {
  padding: 6px 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #cccccc;
  font-size: 12px;
  transition: background 0.1s;
}

.wb-dropdown-item:hover {
  background: #2a2d2e;
}

.wb-dropdown-item-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  color: #858585;
}

.wb-dropdown-item:hover .wb-dropdown-item-icon {
  color: #cccccc;
}

.wb-tag {
  display: inline-flex;
  align-items: center;
  padding: 0 6px;
  height: 18px;
  font-size: 11px;
  border-radius: 2px;
  white-space: nowrap;
  flex-shrink: 0;
}

.wb-tag-success {
  background: #1a3a2a;
  color: #4ec9b0;
}
</style>
