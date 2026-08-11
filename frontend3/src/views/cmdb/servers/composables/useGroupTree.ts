/**
 * 分组树管理逻辑
 * 从 index.vue 拆分：分组树数据、CRUD、右键菜单、编辑重命名
 */

import { computed, ref } from 'vue';
import { ElMessageBox, ElNotification } from 'element-plus';
import {
  fetchCreateServerGroup,
  fetchDeleteServerGroup,
  fetchGetServerGroups,
  fetchUpdateServerGroup
} from '@/service/api';
import type { TreeNode } from '../types/server.types';

export function useGroupTree() {
  // ===== 状态 =====
  const groupTree = ref<TreeNode[]>([]);
  const groupLoading = ref(false);
  const selectedGroupId = ref<number | undefined>(undefined);
  const groupSearchKeyword = ref('');

  // 编辑状态
  const editingNodeId = ref<number | null>(null);
  const editingNodeName = ref('');

  // 分组创建对话框
  const groupDialogVisible = ref(false);
  const groupFormData = ref({ name: '', parentId: 0 });

  // 右键菜单
  const contextMenuVisible = ref(false);
  const contextMenuPosition = ref({ x: 0, y: 0 });
  const currentNode = ref<TreeNode | null>(null);

  // 根节点
  const rootNode: TreeNode = {
    id: 0,
    name: '资产树',
    code: 'root',
    parentId: -1,
    level: 0,
    color: '#409EFF',
    icon: 'mdi:folder',
    sortOrder: 0,
    status: 1,
    serverCount: 0
  };

  // ===== 辅助函数 =====
  function buildFullTree(groups: CMDB.ServerGroup[]): TreeNode[] {
    function countServers(group: CMDB.ServerGroup): number {
      let count = group.servers?.length || 0;
      if (group.children && group.children.length > 0) {
        group.children.forEach(child => {
          count += countServers(child);
        });
      }
      return count;
    }

    function addServerCount(group: CMDB.ServerGroup): TreeNode {
      return {
        ...group,
        serverCount: countServers(group),
        children: group.children && group.children.length > 0 ? group.children.map(child => addServerCount(child)) : undefined
      };
    }

    const rootGroups = groups.filter(g => g.parentId === 0);
    const rootChildren = rootGroups.map(g => addServerCount(g));
    const totalServers = rootChildren.reduce((sum, g) => sum + g.serverCount!, 0);

    return [{ ...rootNode, serverCount: totalServers, children: rootChildren }];
  }

  function filterGroupTree(nodes: TreeNode[], keyword: string): TreeNode[] {
    if (!keyword) return nodes;
    const result: TreeNode[] = [];
    nodes.forEach(node => {
      const matches = node.name.toLowerCase().includes(keyword.toLowerCase());
      const filteredChildren = node.children ? filterGroupTree(node.children, keyword) : [];
      if (matches || filteredChildren.length > 0) {
        result.push({ ...node, children: filteredChildren.length > 0 ? filteredChildren : node.children });
      }
    });
    return result;
  }

  // ===== 计算属性 =====
  const filteredGroupTree = computed(() => filterGroupTree(groupTree.value, groupSearchKeyword.value));

  const groupTreeForSelect = computed(() => {
    if (groupTree.value.length === 0) return [];
    return groupTree.value[0].children || [];
  });

  // ===== 数据获取 =====
  async function getGroups() {
    groupLoading.value = true;
    try {
      const { data } = await fetchGetServerGroups();
      if (data && Array.isArray(data)) {
        groupTree.value = buildFullTree(data);
      }
    } catch (error) {
      console.error('获取分组失败:', error);
      ElNotification.error('获取分组失败');
    } finally {
      groupLoading.value = false;
    }
  }

  // ===== 节点交互 =====
  function handleNodeClick(data: TreeNode) {
    if (data.id === 0) {
      selectedGroupId.value = undefined;
    } else if (selectedGroupId.value === data.id) {
      selectedGroupId.value = undefined;
    } else {
      selectedGroupId.value = data.id;
    }
  }

  function handleNodeContextMenu(event: MouseEvent, data: TreeNode) {
    event.preventDefault();
    event.stopPropagation();
    currentNode.value = data;
    contextMenuPosition.value = { x: event.clientX, y: event.clientY };
    contextMenuVisible.value = true;
  }

  function handleContextMenuClose() {
    contextMenuVisible.value = false;
  }

  function handleGlobalClick() {
    if (contextMenuVisible.value) {
      contextMenuVisible.value = false;
    }
  }

  // ===== 分组 CRUD =====
  function handleAddRootGroup() {
    groupFormData.value.name = '';
    groupFormData.value.parentId = 0;
    groupDialogVisible.value = true;
  }

  function handleOpenAddChildDialog(parentId: number) {
    groupFormData.value.name = '';
    groupFormData.value.parentId = parentId;
    groupDialogVisible.value = true;
  }

  function handleAddGroup() {
    if (!currentNode.value) return;
    contextMenuVisible.value = false;
    handleOpenAddChildDialog(currentNode.value.id);
  }

  async function handleSaveGroup() {
    const newGroup: CMDB.ServerGroupForm = {
      name: groupFormData.value.name,
      code: groupFormData.value.name.toLowerCase().replace(/[^a-z0-9]/g, ''),
      parentId: groupFormData.value.parentId,
      color: '#409EFF',
      icon: 'mdi:folder',
      sortOrder: 0,
      status: 1
    };

    try {
      await fetchCreateServerGroup(newGroup);
      ElNotification.success('创建成功');
      groupDialogVisible.value = false;
      setTimeout(async () => { await getGroups(); }, 300);
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'message' in error) {
        ElNotification.error(typeof error.message === 'string' ? error.message : '创建失败');
      }
    }
  }

  function handleEditGroup() {
    if (!currentNode.value) return;
    contextMenuVisible.value = false;
    if (currentNode.value.id === 0) {
      ElNotification.warning('根节点不能重命名');
      return;
    }
    editingNodeId.value = currentNode.value.id;
    editingNodeName.value = currentNode.value.name;
  }

  async function saveEditGroup() {
    if (editingNodeId.value === null) return;
    const newName = editingNodeName.value.trim();
    if (!newName) { ElNotification.warning('分组名称不能为空'); return; }
    if (newName.length < 2 || newName.length > 50) { ElNotification.warning('分组名称长度在 2 到 50 个字符'); return; }

    try {
      await fetchUpdateServerGroup(editingNodeId.value, { name: newName });
      ElNotification.success('重命名成功');
      editingNodeId.value = null;
      setTimeout(async () => { await getGroups(); }, 300);
    } catch (error: unknown) {
      console.error('重命名失败:', error);
      if (error && typeof error === 'object' && 'message' in error) {
        ElNotification.error(typeof error.message === 'string' ? error.message : '重命名失败');
      }
    }
  }

  function cancelEditGroup() {
    editingNodeId.value = null;
    editingNodeName.value = '';
  }

  function handleEditKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') { event.preventDefault(); saveEditGroup(); }
    else if (event.key === 'Escape') { event.preventDefault(); cancelEditGroup(); }
  }

  async function createGroupWithName(parentId: number, name: string) {
    try {
      const newGroup: CMDB.ServerGroupForm = {
        name,
        code: name.toLowerCase().replace(/[^a-z0-9]/g, ''),
        parentId,
        color: '#409EFF',
        icon: 'mdi:folder',
        sortOrder: 0,
        status: 1
      };
      await fetchCreateServerGroup(newGroup);
      ElNotification.success('创建成功');

      setTimeout(async () => {
        await getGroups();
        const findNewGroup = (nodes: TreeNode[], targetParentId: number, targetName: string): TreeNode | null => {
          for (const node of nodes) {
            if (node.id === targetParentId && node.children) {
              for (const child of node.children) {
                if (child.name === targetName) return child;
              }
            }
            if (node.children) {
              const found = findNewGroup(node.children, targetParentId, targetName);
              if (found) return found;
            }
          }
          return null;
        };
        const newGroupNode = findNewGroup(groupTree.value, parentId, name);
        if (newGroupNode) {
          selectedGroupId.value = newGroupNode.id;
          currentNode.value = newGroupNode;
          editingNodeId.value = newGroupNode.id;
          editingNodeName.value = newGroupNode.name;
        }
      }, 300);
    } catch (error: unknown) {
      console.error('创建分组失败:', error);
      if (error && typeof error === 'object' && 'message' in error) {
        ElNotification.error(typeof error.message === 'string' ? error.message : '创建失败');
      }
    }
  }

  function handleDeleteGroup() {
    if (!currentNode.value) return;
    const hasChildren = currentNode.value.children && currentNode.value.children.length > 0;
    const hasServers = currentNode.value.serverCount && currentNode.value.serverCount > 0;

    if (hasServers) {
      ElNotification.warning('该分组下还有主机，请先移除主机后再删除！');
      contextMenuVisible.value = false;
      return;
    }

    let message = `确定要删除分组 "${currentNode.value.name}" 吗？`;
    if (hasChildren) message += '\n\n注意：该分组包含子分组，删除后子分组也将被删除！';

    ElMessageBox.confirm(message, '删除确认', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning', dangerouslyUseHTMLString: true })
      .then(async () => {
        try {
          await fetchDeleteServerGroup(currentNode.value!.id);
          ElNotification.success('删除成功');
          contextMenuVisible.value = false;
          await getGroups();
        } catch (error) {
          console.error('删除失败:', error);
          ElNotification.error('删除失败');
        }
      })
      .catch(() => { contextMenuVisible.value = false; });
  }

  function handleAddServer() {
    if (!currentNode.value) return;
    contextMenuVisible.value = false;
    if (currentNode.value.id !== 0) {
      selectedGroupId.value = currentNode.value.id;
    }
  }

  function getNodeClass(node: TreeNode) {
    const classes = ['custom-tree-node'];
    if (node.status === 0) classes.push('disabled');
    return classes.join(' ');
  }

  return {
    // 状态
    groupTree,
    groupLoading,
    selectedGroupId,
    groupSearchKeyword,
    editingNodeId,
    editingNodeName,
    groupDialogVisible,
    groupFormData,
    contextMenuVisible,
    contextMenuPosition,
    currentNode,
    // 计算属性
    filteredGroupTree,
    groupTreeForSelect,
    // 方法
    getGroups,
    handleNodeClick,
    handleNodeContextMenu,
    handleContextMenuClose,
    handleGlobalClick,
    handleAddRootGroup,
    handleOpenAddChildDialog,
    handleAddGroup,
    handleSaveGroup,
    handleEditGroup,
    saveEditGroup,
    cancelEditGroup,
    handleEditKeydown,
    createGroupWithName,
    handleDeleteGroup,
    handleAddServer,
    getNodeClass
  };
}
