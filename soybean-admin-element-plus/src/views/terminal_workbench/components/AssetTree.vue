<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { Teleport } from 'vue';
import { ElButton, ElDropdown, ElDropdownItem, ElDropdownMenu, ElInput, ElTag, ElTooltip, ElTree } from 'element-plus';
import { fetchGetServerGroups, fetchGetServersByGroup } from '@/service/api';

interface TreeNode {
  id: number;
  name: string;
  code: string;
  parentId: number;
  level: number;
  description?: string;
  color: string;
  icon: string;
  sortOrder: number;
  status: number;
  children?: TreeNode[];
  serverCount?: number;
  isLeaf?: boolean; // 标记是否为叶子节点（实际是服务器）
  server?: CMDB.Server; // 服务器数据
}

interface Props {
  currentSessions: number[]; // 当前已连接的主机ID列表
}

interface Emits {
  (e: 'connect', server: CMDB.Server): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const treeData = ref<TreeNode[]>([]);
const loading = ref(false);
const searchKeyword = ref('');
const searchExpanded = ref(false);
const treeRef = ref<InstanceType<typeof ElTree> | null>(null);

// 右键菜单
const contextMenuVisible = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
const contextMenuServer = ref<CMDB.Server | null>(null);

// 右键菜单项
interface ContextMenuItem {
  icon: string;
  label: string;
  action: () => void;
}

const contextMenuItems = computed<ContextMenuItem[]>(() => [
  {
    icon: 'mdi:console-line',
    label: '连接',
    action: () => {
      if (contextMenuServer.value) {
        handleConnect(contextMenuServer.value);
      }
    }
  },
  {
    icon: 'mdi:window-maximize',
    label: '新窗口打开',
    action: () => {
      if (contextMenuServer.value) {
        const server = contextMenuServer.value;
        const params = new URLSearchParams({
          sessionId: `temp-${server.id}`,
          serverName: server.hostname,
          serverIp: server.ip,
          websocketUrl: '' // 需要建立连接后获取
        });
        window.open(`/cmdb/terminal/${server.id}?${params.toString()}`, '_blank');
        hideContextMenu();
      }
    }
  },
  {
    icon: 'mdi:view-split-vertical',
    label: '分屏连接',
    action: () => {
      window.$message?.info('分屏连接功能开发中');
      hideContextMenu();
    }
  },
  {
    icon: 'mdi:star',
    label: '收藏',
    action: () => {
      if (contextMenuServer.value) {
        window.$message?.info(`已收藏 ${contextMenuServer.value.hostname}`);
      }
      hideContextMenu();
    }
  }
]);

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
      // 服务器节点
      if (
        node.name.toLowerCase().includes(keyword) ||
        (node.server.ip && node.server.ip.toLowerCase().includes(keyword))
      ) {
        result.push(node);
      }
    } else if (node.children) {
      // 分组节点
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

// 构建完整的树（包含所有服务器）
async function loadTreeData() {
  loading.value = true;
  try {
    // 获取分组数据
    const { data: groups } = await fetchGetServerGroups();

    // 构建树结构
    async function buildTree(groups: CMDB.ServerGroup[]): Promise<TreeNode[]> {
      // 为每个分组加载服务器列表
      const groupServerMap = new Map<number, CMDB.Server[]>();

      // 收集所有分组ID（包括子分组）
      function collectGroupIds(groupList: CMDB.ServerGroup[]): number[] {
        const ids: number[] = [];
        groupList.forEach(g => {
          ids.push(g.id);
          if (g.children && g.children.length > 0) {
            ids.push(...collectGroupIds(g.children));
          }
        });
        return ids;
      }

      const allGroupIds = collectGroupIds(groups);

      // 并行获取所有分组的服务器
      await Promise.all(
        allGroupIds.map(async groupId => {
          try {
            const { data } = await fetchGetServersByGroup(groupId, { page: 1, pageSize: 10000 });
            groupServerMap.set(groupId, data?.list || []);
          } catch (error) {
            console.error(`获取分组 ${groupId} 的服务器失败:`, error);
            groupServerMap.set(groupId, []);
          }
        })
      );

      // 构建树节点
      function buildNode(group: CMDB.ServerGroup): TreeNode {
        const childGroups = (group.children || []).map(child => buildNode(child));
        const groupServers = groupServerMap.get(group.id) || [];

        const node: TreeNode = {
          ...group,
          children: [],
          serverCount: groupServers.length + childGroups.reduce((sum, child) => sum + (child.serverCount || 0), 0)
        };

        // 添加子分组节点
        childGroups.forEach(child => {
          node.children!.push(child);
        });

        // 添加该分组下的服务器节点
        groupServers.forEach(server => {
          node.children!.push(createServerNode(server));
        });

        return node;
      }

      // 构建根节点
      const rootGroups = groups.filter(g => g.parentId === 0);
      return rootGroups.map(g => buildNode(g));
    }

    treeData.value = await buildTree(groups || []);
  } catch (error) {
    console.error('加载资产树失败:', error);
  } finally {
    loading.value = false;
  }
}

// 创建服务器节点
function createServerNode(server: CMDB.Server): TreeNode {
  return {
    id: `server-${server.id}` as unknown as number,
    name: `${server.hostname} (${server.ip})`,
    code: server.hostname,
    parentId: server.groups?.[0]?.id || 0,
    level: 2,
    color: '#757575',
    icon: 'mdi:server',
    sortOrder: 0,
    status: 1,
    isLeaf: true,
    server
  };
}

// 节点点击
function handleNodeClick(node: TreeNode) {
  // 可以在这里实现点击分组显示服务器列表的功能
}

// 双击节点
function handleNodeDblClick(node: any) {
  console.log('[AssetTree] 双击节点:', node);
  // 处理两种调用方式：直接传递 data 对象，或传递完整的 node 对象
  const data = (node.data || node) as TreeNode;
  console.log('[AssetTree] 节点数据:', data);
  if (data.isLeaf && data.server) {
    console.log('[AssetTree] 触发连接:', data.server);
    handleConnect(data.server);
  }
}

// 连接服务器
async function handleConnect(server: CMDB.Server) {
  console.log('[AssetTree] handleConnect 被调用:', server);
  hideContextMenu();
  // 直接触发连接，权限检查在对话框中进行
  emit('connect', server);
}

// 处理右键菜单
function handleContextMenu(data: TreeNode, event: MouseEvent) {
  console.log('[AssetTree] 右键菜单触发:', data);
  if (data.isLeaf && data.server) {
    event.preventDefault();
    event.stopPropagation();
    contextMenuServer.value = data.server;
    contextMenuPosition.value = { x: event.clientX, y: event.clientY };
    contextMenuVisible.value = true;
    console.log('[AssetTree] 右键菜单显示');
  }
}

// 隐藏右键菜单
function hideContextMenu() {
  contextMenuVisible.value = false;
  contextMenuServer.value = null;
}

// 点击其他地方关闭右键菜单
function handleClickOutside() {
  hideContextMenu();
}

// 刷新
function handleRefresh() {
  loadTreeData();
}

// 切换搜索框展开状态
function toggleSearch() {
  searchExpanded.value = !searchExpanded.value;
  if (searchExpanded.value) {
    // 展开时自动聚焦
    nextTick(() => {
      const input = document.querySelector('.search-input-box input') as HTMLInputElement;
      if (input) input.focus();
    });
  }
}

// 处理搜索框失焦
function handleSearchBlur() {
  // 延迟关闭，避免点击清除按钮时立即关闭
  setTimeout(() => {
    if (!searchKeyword.value) {
      searchExpanded.value = false;
    }
  }, 150);
}

// 获取节点CSS类
function getNodeClass(node: TreeNode): string {
  const classes = [];
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

// 监听搜索关键词变化
watch(searchKeyword, () => {
  // 搜索逻辑由 computed 处理
});

onMounted(() => {
  loadTreeData();
  // 添加全局点击事件监听，用于关闭右键菜单
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<template>
  <div class="asset-tree">
    <!-- 搜索和操作栏 -->
    <div class="toolbar">
      <span class="title">主机资产</span>
      <div v-if="searchExpanded" class="search-box search-input-box">
        <icon-mdi-magnify class="search-icon" />
        <ElInput v-model="searchKeyword" placeholder="搜索..." size="small" clearable @blur="handleSearchBlur" />
      </div>
      <ElButton v-else size="small" link class="search-toggle-btn" @click="toggleSearch">
        <icon-mdi-magnify />
      </ElButton>
      <ElButton size="small" link class="refresh-btn" @click="handleRefresh">
        <icon-mdi-refresh :class="{ spinning: loading }" />
      </ElButton>
    </div>

    <!-- 树组件容器 -->
    <div class="tree-container">
      <!-- 树组件 -->
      <ElTree
        ref="treeRef"
        :data="filteredTreeData"
        :props="{ label: 'name', children: 'children' }"
        :expand-on-click-node="false"
        node-key="id"
        :default-expand-all="false"
        class="server-tree"
        @node-click="handleNodeClick"
        @node-dblclick="handleNodeDblClick"
      >
        <template #default="{ node, data }">
          <div
            class="custom-node"
            :class="getNodeClass(data)"
            @dblclick="handleNodeDblClick({ data })"
            @contextmenu.prevent="handleContextMenu(data, $event)"
          >
            <div class="node-content">
              <component :is="`icon-${data.icon.replace(':', '-')}`" class="node-icon" :style="{ color: data.color }" />
              <span class="node-label">{{ node.label }}</span>
              <span v-if="data.serverCount !== undefined && !data.isLeaf" class="server-count">({{ data.serverCount }})</span>
              <ElTag
                v-if="data.isLeaf && data.server && currentSessions.includes(Number(data.server.id))"
                size="small"
                type="success"
              >
                已连接
              </ElTag>
            </div>
          </div>
        </template>
      </ElTree>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <icon-mdi-loading class="loading-icon" />
        <span>加载中...</span>
      </div>

      <!-- 空状态 -->
      <div v-else-if="treeData.length === 0" class="empty-state">
        <icon-mdi-server-off class="empty-icon" />
        <p>{{ searchKeyword ? '没有找到匹配的主机' : '暂无主机' }}</p>
      </div>
    </div>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="context-menu"
        :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
        @click.self="handleClickOutside"
      >
        <div
          v-for="item in contextMenuItems"
          :key="item.label"
          class="context-menu-item"
          @click.stop="item.action()"
        >
          <component :is="`icon-${item.icon.replace(':', '-')}`" class="menu-icon" />
          <span>{{ item.label }}</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.asset-tree {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 6px;
  border-bottom: 1px solid #333333;
  flex-shrink: 0;
  min-height: 26px;
}

.toolbar .title {
  font-size: 11px;
  font-weight: 500;
  color: #e0e0e0;
  flex-shrink: 0;
}

.search-box {
  position: relative;
  flex: 1;
  min-width: 0;
}

.search-icon {
  position: absolute;
  left: 6px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 11px;
  color: #666;
  pointer-events: none;
}

.search-box :deep(.el-input__wrapper) {
  padding-left: 22px;
  height: 22px;
  box-shadow: none;
  background: #1a1a1a;
  border: 1px solid #333333;
  border-radius: 11px;
}

.search-box :deep(.el-input__wrapper:hover),
.search-box :deep(.el-input__wrapper.is-focus) {
  border-color: #4a4a4a;
}

.search-box :deep(.el-input__inner) {
  color: #e8e8e8;
  font-size: 11px;
}

.search-box :deep(.el-input__inner::placeholder) {
  color: #777;
}

.search-toggle-btn {
  flex-shrink: 0;
  padding: 2px;
  color: #888;
  font-size: 14px;
}

.search-toggle-btn:hover {
  color: #bbb;
}

.refresh-btn {
  flex-shrink: 0;
  padding: 2px;
  color: #888;
}

.refresh-btn:hover {
  color: #bbb;
}

.tree-container {
  flex: 1;
  overflow-y: auto;
  padding: 2px;
}

.tree-container::-webkit-scrollbar {
  width: 4px;
}

.tree-container::-webkit-scrollbar-track {
  background: #0a0a0a;
}

.tree-container::-webkit-scrollbar-thumb {
  background: #333333;
}

.tree-container::-webkit-scrollbar-thumb:hover {
  background: #4a4a4a;
}

.server-tree {
  background: transparent;
}

.server-tree :deep(.el-tree-node__content) {
  padding: 1px 0;
  min-height: 24px;
  background: transparent;
}

.server-tree :deep(.el-tree-node__content:hover) {
  background: #1a1a1a;
}

.server-tree :deep(.el-tree-node__expand-icon) {
  color: #666;
  font-size: 9px;
  width: 9px;
}

.custom-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  cursor: pointer;
  position: relative;
}

.node-content {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
  min-width: 0;
}

.node-icon {
  font-size: 12px;
  flex-shrink: 0;
  color: #888;
}

.node-label {
  font-size: 11px;
  color: #e8e8e8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.server-count {
  margin-left: 2px;
  font-size: 10px;
  color: #777;
  flex-shrink: 0;
}

.is-connected {
  background: #1a3a2a;
}

.is-offline {
  opacity: 0.5;
}

.dropdown-trigger {
  position: absolute;
  left: 0;
  top: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  z-index: 10;
  pointer-events: auto;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px;
  color: #888;
  font-size: 11px;
  gap: 6px;
}

.loading-icon {
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

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 12px;
  color: #888;
}

.empty-icon {
  font-size: 24px;
  margin-bottom: 6px;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 11px;
}

/* 右键菜单样式 */
.context-menu {
  position: fixed;
  z-index: 9999;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  min-width: 140px;
  padding: 4px 0;
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  cursor: pointer;
  color: #f0f0f0;
  font-size: 12px;
  transition: background 0.15s;
}

.context-menu-item:hover {
  background: #2a2a2a;
  color: #fff;
}

.menu-icon {
  font-size: 16px;
  flex-shrink: 0;
}
</style>
