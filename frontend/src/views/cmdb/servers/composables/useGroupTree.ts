/**
 * 分组树管理逻辑
 */

import { computed, ref } from 'vue';
import { ElMessageBox, ElNotification } from 'element-plus';
import {
  type ServerGroup,
  fetchAssignServerToGroups,
  fetchCreateServerGroup,
  fetchDeleteServerGroup,
  fetchGetServerGroups,
  fetchUpdateServerGroup
} from '@/service/api';
import type { TreeNode } from '../types/server.types';

export function useGroupTree() {
  // 树形数据
  const groupTree = ref<TreeNode[]>([]);
  const groupLoading = ref(false);

  // 编辑状态
  const editingNodeId = ref<number | null>(null);
  const editingNodeName = ref('');
  const selectedGroupId = ref<number | undefined>(undefined);

  // 搜索
  const groupSearchKeyword = ref('');

  // 对话框状态
  const groupDialogVisible = ref(false);
  const groupForm = ref({
    name: '',
    code: '',
    parentId: 0,
    owner: '',
    description: ''
  });

  // 根节点定义
  const rootNode: TreeNode = {
    id: 0,
    name: '全部分组',
    code: 'root',
    parentId: 0,
    level: 0,
    color: '#409EFF',
    icon: 'el-icon-folder',
    sortOrder: 0,
    status: 1
  };

  /**
   * 构建带根节点的完整树
   */
  const buildFullTree = (groups: ServerGroup[]): TreeNode[] => {
    // 计算每个分组的服务器数量（包括子分组的服务器）
    const countServers = (group: ServerGroup): number => {
      let count = group.servers?.length || 0;
      if (group.children && group.children.length > 0) {
        group.children.forEach(child => {
          count += countServers(child);
        });
      }
      return count;
    };

    // 为每个分组添加 serverCount
    const addServerCount = (group: ServerGroup): TreeNode => {
      const node: TreeNode = {
        ...group,
        serverCount: countServers(group),
        children:
          group.children && group.children.length > 0 ? group.children.map(child => addServerCount(child)) : undefined
      };
      return node;
    };

    // 只处理一级分组（parentId === 0）
    const rootGroups = groups.filter(g => g.parentId === 0);
    const rootChildren = rootGroups.map(g => addServerCount(g));

    // 计算总主机数
    const totalServers = rootChildren.reduce((sum, g) => sum + (g.serverCount || 0), 0);

    const root = {
      ...rootNode,
      serverCount: totalServers,
      children: rootChildren
    };

    return [root];
  };

  /**
   * 过滤分组树
   */
  const filterGroupTree = (nodes: TreeNode[], keyword: string): TreeNode[] => {
    if (!keyword) return nodes;

    const result: TreeNode[] = [];

    nodes.forEach(node => {
      const matches = node.name.toLowerCase().includes(keyword.toLowerCase());
      const filteredChildren = node.children ? filterGroupTree(node.children, keyword) : [];

      if (matches || filteredChildren.length > 0) {
        result.push({
          ...node,
          children: filteredChildren.length > 0 ? filteredChildren : node.children
        });
      }
    });

    return result;
  };

  /**
   * 计算的属性
   */
  const filteredGroupTree = computed(() => {
    if (!groupSearchKeyword.value) return groupTree.value;
    return filterGroupTree(groupTree.value, groupSearchKeyword.value);
  });

  const groupTreeForSelect = computed(() => {
    const tree = filteredGroupTree.value;

    const flatten = (nodes: TreeNode[]): TreeNode[] => {
      const result: TreeNode[] = [];

      nodes.forEach(node => {
        result.push(node);
        if (node.children) {
          result.push(...flatten(node.children));
        }
      });

      return result;
    };

    return flatten(tree);
  });

  /**
   * 获取分组列表
   */
  const getGroups = async () => {
    groupLoading.value = true;
    try {
      const { data } = await fetchGetServerGroups();
      if (data) {
        groupTree.value = buildFullTree(data);
      }
    } finally {
      groupLoading.value = false;
    }
  };

  /**
   * 处理节点点击
   */
  const handleNodeClick = (data: TreeNode) => {
    selectedGroupId.value = data.id === 0 ? undefined : data.id;
  };

  /**
   * 处理全局点击（关闭右键菜单）
   */
  const handleGlobalClick = () => {
    // 这里可以添加关闭右键菜单的逻辑
  };

  /**
   * 创建分组
   */
  const createGroup = async () => {
    try {
      const response = await fetchCreateServerGroup({
        name: groupForm.value.name,
        code: groupForm.value.code,
        parentId: groupForm.value.parentId,
        owner: groupForm.value.owner,
        description: groupForm.value.description
      });

      if (response.data) {
        ElNotification.success('分组创建成功');
        groupDialogVisible.value = false;
        await getGroups();
        return { success: true };
      }
      return { success: false, message: '创建失败' };
    } catch (error) {
      const message =
        error && typeof error === 'object' && 'message' in error
          ? typeof error.message === 'string'
            ? error.message
            : '创建失败'
          : '创建失败';
      return { success: false, message };
    }
  };

  /**
   * 更新分组名称
   */
  const updateGroupName = async (nodeId: number, newName: string) => {
    try {
      await fetchUpdateServerGroup(nodeId, { name: newName });
      await getGroups();
      return { success: true };
    } catch (error) {
      const message =
        error && typeof error === 'object' && 'message' in error
          ? typeof error.message === 'string'
            ? error.message
            : '重命名失败'
          : '重命名失败';
      return { success: false, message };
    }
  };

  /**
   * 删除分组
   */
  const deleteGroup = async (nodeId: number) => {
    try {
      await ElMessageBox.confirm('确定要删除这个分组吗？分组下的服务器将变为未分组状态。', '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });

      await fetchDeleteServerGroup(nodeId);
      ElNotification.success('分组删除成功');
      await getGroups();
      return { success: true };
    } catch (error) {
      if (error === 'cancel') {
        return { success: false, message: '用户取消' };
      }

      const message =
        error && typeof error === 'object' && 'message' in error
          ? typeof error.message === 'string'
            ? error.message
            : '删除失败'
          : '删除失败';
      return { success: false, message };
    }
  };

  /**
   * 分配服务器到分组
   */
  const assignServersToGroups = async (serverIds: number[], groupIds: number[]) => {
    try {
      await fetchAssignServerToGroups({
        serverIds,
        groupIds
      });
      await getGroups();
      return { success: true };
    } catch (error) {
      return { success: false, message: '分配失败' };
    }
  };

  return {
    // 状态
    groupTree,
    groupLoading,
    editingNodeId,
    editingNodeName,
    selectedGroupId,
    groupSearchKeyword,
    groupDialogVisible,
    groupForm,

    // 计算属性
    filteredGroupTree,
    groupTreeForSelect,

    // 方法
    getGroups,
    handleNodeClick,
    handleGlobalClick,
    createGroup,
    updateGroupName,
    deleteGroup,
    assignServersToGroups
  };
}
