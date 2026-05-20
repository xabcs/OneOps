<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue';
import {
  fetchGetServers,
  fetchCreateServer,
  fetchUpdateServer,
  fetchDeleteServer,
  fetchGetServerGroups,
  fetchCreateServerGroup,
  fetchUpdateServerGroup,
  fetchDeleteServerGroup,
  fetchGetSSHCredentials
} from '@/service/api';
import { ElNotification, ElMessageBox, FormInstance, FormRules } from 'element-plus';
import { $t } from '@/locales';

defineOptions({ name: 'CmdbServers' });

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
}

// ========== 分组树相关 ==========
const groupTree = ref<TreeNode[]>([]);
const groupLoading = ref(false);
const treeRef = ref();
const selectedGroupId = ref<number | undefined>(undefined);
const groupSearchKeyword = ref('');

// 编辑状态
const editingNodeId = ref<number | null>(null);
const editingNodeName = ref('');
const editInputRef = ref<HTMLInputElement | null>(null);

// 分组创建对话框
const groupDialogVisible = ref(false);
const groupFormRef = ref<FormInstance>();

// 分组表单
const groupFormData = reactive({
  name: '',
  parentId: 0
});

// 分组表单验证规则
const groupFormRules: FormRules = {
  name: [
    { required: true, message: '请输入分组名称', trigger: 'blur' },
    { min: 2, max: 50, message: '分组名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  parentId: [
    { required: true, message: '请选择父分组', trigger: 'change' }
  ]
};

// 全局点击处理，用于关闭右键菜单
function handleGlobalClick(event: MouseEvent) {
  if (contextMenuVisible.value) {
    contextMenuVisible.value = false;
  }
}

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

// 分组右键菜单
const contextMenuVisible = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
const currentNode = ref<TreeNode | null>(null);

// ========== 主机列表相关 ==========
const tableData = ref<CMDB.Server[]>([]);
const loading = ref(false);
const total = ref(0);
const selectedIds = ref<number[]>([]);

// 分页信息
const pagination = reactive({
  page: 1,
  pageSize: 20
});

// 搜索表单
const searchForm = reactive({
  hostname: '',
  ip: ''
});

// ========== SSH凭证相关 ==========
const sshCredentials = ref<CMDB.SSHCredential[]>([]);

// ========== 对话框相关 ==========
const dialogVisible = ref(false);
const dialogTitle = ref('');
const serverFormRef = ref<FormInstance>();
const serverType = ref<'normal' | 'cloud'>('normal');
const submitError = ref(''); // 提交错误信息

// 主机表单
const serverForm = reactive<CMDB.ServerForm & { groupIds?: number[] }>({
  hostname: '',
  ip: '',
  innerIp: '',
  credentialId: undefined as unknown as number,
  serverType: 'vm',
  groupIds: [],
  sshPort: 22,
  remarks: ''
});

// 云主机表单
const cloudForm = reactive<CMDB.CloudServerForm>({
  provider: 'aliyun',
  instanceName: '',
  instanceType: '',
  region: '',
  zone: '',
  chargeType: 'postpay'
});

// ========== 表单验证规则 ==========
const serverFormRules: FormRules = {
  hostname: [
    { required: true, message: '请输入主机名', trigger: 'blur' },
    { min: 2, max: 100, message: '主机名长度在 2 到 100 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9.-]+$/, message: '主机名只能包含字母、数字、点和连字符', trigger: 'blur' }
  ],
  ip: [
    { required: true, message: '请输入连接IP', trigger: 'blur' },
    { pattern: /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/, message: '请输入有效的IP地址', trigger: 'blur' }
  ],
  credentialId: [
    { required: true, message: '请选择SSH凭证', trigger: 'change' }
  ]
};

// ========== 辅助函数 ==========
// 构建带根节点的完整树（直接使用后端返回的树结构）
function buildFullTree(groups: CMDB.ServerGroup[]): TreeNode[] {
  // 后端已经返回了嵌套的 children 结构，直接使用
  // 计算每个分组的服务器数量（包括子分组的服务器）
  function countServers(group: CMDB.ServerGroup): number {
    let count = group.servers?.length || 0;
    if (group.children && group.children.length > 0) {
      group.children.forEach(child => {
        count += countServers(child);
      });
    }
    return count;
  }

  // 为每个分组添加 serverCount
  function addServerCount(group: CMDB.ServerGroup): TreeNode {
    const node: TreeNode = {
      ...group,
      serverCount: countServers(group),
      children: group.children && group.children.length > 0
        ? group.children.map(child => addServerCount(child))
        : undefined
    };
    return node;
  }

  // 只处理一级分组（parentId === 0）
  const rootGroups = groups.filter(g => g.parentId === 0);
  const rootChildren = rootGroups.map(g => addServerCount(g));

  // 计算总主机数
  const totalServers = rootChildren.reduce((sum, g) => sum + g.serverCount!, 0);

  const root = {
    ...rootNode,
    serverCount: totalServers,
    children: rootChildren
  };

  return [root];
}

// 过滤分组树
function filterGroupTree(nodes: TreeNode[], keyword: string): TreeNode[] {
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
}

// 计算过滤后的分组树
const filteredGroupTree = computed(() => {
  return filterGroupTree(groupTree.value, groupSearchKeyword.value);
});

// 用于表单选择的分组树（不包含根节点）
const groupTreeForSelect = computed(() => {
  if (groupTree.value.length === 0) return [];
  return groupTree.value[0].children || [];
});

// ========== 数据获取 ==========
// 获取分组列表
async function getGroups() {
  groupLoading.value = true;
  try {
    const { data } = await fetchGetServerGroups();
    console.log('后端返回的分组数据:', data);
    if (data && Array.isArray(data)) {
      groupTree.value = buildFullTree(data);
      console.log('构建后的完整树结构:', groupTree.value);
      console.log('根节点子分组数:', groupTree.value[0]?.children?.length);
    }
  } catch (error) {
    console.error('获取分组失败:', error);
    ElNotification.error('获取分组失败');
  } finally {
    groupLoading.value = false;
  }
}

// 获取主机列表
async function getServers() {
  loading.value = true;
  try {
    const params: CMDB.ServerQuery = {
      page: pagination.page,
      pageSize: pagination.pageSize
    };

    // 如果选中了分组，则过滤该分组下的主机
    if (selectedGroupId.value) {
      params.groupId = selectedGroupId.value;
    }

    // 添加搜索条件
    if (searchForm.hostname) {
      params.hostname = searchForm.hostname;
    }
    if (searchForm.ip) {
      params.ip = searchForm.ip;
    }

    const { data } = await fetchGetServers(params);
    tableData.value = data?.list || [];
    total.value = data?.total || 0;
  } catch (error) {
    console.error('获取主机列表失败:', error);
    ElNotification.error('获取主机列表失败');
  } finally {
    loading.value = false;
  }
}

// 获取SSH凭证列表
async function getSSHCredentials() {
  try {
    const { data } = await fetchGetSSHCredentials();
    sshCredentials.value = data || [];
  } catch (error) {
    console.error('获取SSH凭证失败:', error);
  }
}

// ========== 事件处理 ==========
// 点击分组节点
function handleNodeClick(data: TreeNode) {
  // 根节点(id为0)代表所有主机
  if (data.id === 0) {
    selectedGroupId.value = undefined;
  } else if (selectedGroupId.value === data.id) {
    // 如果点击的是已选中的节点，则取消选中
    selectedGroupId.value = undefined;
  } else {
    selectedGroupId.value = data.id;
  }
  pagination.page = 1;
  selectedIds.value = [];
  getServers();
}

// 搜索主机
function handleSearch() {
  pagination.page = 1;
  getServers();
}

// 重置搜索
function handleReset() {
  searchForm.hostname = '';
  searchForm.ip = '';
  pagination.page = 1;
  getServers();
}

// 刷新分组树
function handleRefreshGroups() {
  getGroups();
}

// ========== 分组管理功能 ==========
// 右键菜单处理
function handleNodeContextMenu(event: MouseEvent, data: TreeNode) {
  event.preventDefault();
  event.stopPropagation();

  currentNode.value = data;
  contextMenuPosition.value = {
    x: event.clientX,
    y: event.clientY
  };
  contextMenuVisible.value = true;
}

// 关闭右键菜单
function handleContextMenuClose() {
  contextMenuVisible.value = false;
}

// 添加分组（在当前节点下）
function handleAddGroup() {
  if (!currentNode.value) return;

  contextMenuVisible.value = false;
  // 打开添加子分组对话框
  handleOpenAddChildDialog(currentNode.value.id);
}

// 添加根分组
function handleAddRootGroup() {
  groupFormData.name = '';
  groupFormData.parentId = 0;
  groupDialogVisible.value = true;
}

// 添加子分组（已弃用，使用 handleAddGroup）
function handleAddChildGroup() {
  handleAddGroup();
}

// 打开添加子分组对话框
function handleOpenAddChildDialog(parentId: number) {
  groupFormData.name = '';
  groupFormData.parentId = parentId;
  groupDialogVisible.value = true;
}

// 保存分组
async function handleSaveGroup() {
  if (!groupFormRef.value) return;

  try {
    await groupFormRef.value.validate();

    const newGroup: CMDB.ServerGroupForm = {
      name: groupFormData.name,
      code: groupFormData.name.toLowerCase().replace(/[^a-z0-9]/g, ''),
      parentId: groupFormData.parentId,
      color: '#409EFF',
      icon: 'mdi:folder',
      sortOrder: 0,
      status: 1
    };

    await fetchCreateServerGroup(newGroup);
    ElNotification.success('创建成功');
    groupDialogVisible.value = false;

    // 刷新分组树
    setTimeout(async () => {
      await getGroups();
    }, 300);
  } catch (error) {
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '创建失败');
    }
  }
}

// 编辑分组
function handleEditGroup() {
  if (!currentNode.value) return;

  contextMenuVisible.value = false;

  // 根节点不能重命名
  if (currentNode.value.id === 0) {
    ElNotification.warning('根节点不能重命名');
    return;
  }

  // 设置编辑状态
  editingNodeId.value = currentNode.value.id;
  editingNodeName.value = currentNode.value.name;

  // 聚焦到输入框
  setTimeout(() => {
    if (editInputRef.value) {
      editInputRef.value.focus();
      editInputRef.value.select();
    }
  }, 100);
}

// 保存编辑的分组名称
async function saveEditGroup() {
  if (editingNodeId.value === null) return;

  const newName = editingNodeName.value.trim();
  if (!newName) {
    ElNotification.warning('分组名称不能为空');
    return;
  }

  if (newName.length < 2 || newName.length > 50) {
    ElNotification.warning('分组名称长度在 2 到 50 个字符');
    return;
  }

  try {
    await fetchUpdateServerGroup(editingNodeId.value, { name: newName });
    ElNotification.success('重命名成功');
    editingNodeId.value = null;

    // 刷新分组树
    setTimeout(async () => {
      await getGroups();
    }, 300);
  } catch (error) {
    console.error('重命名失败:', error);
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '重命名失败');
    }
  }
}

// 取消编辑
function cancelEditGroup() {
  editingNodeId.value = null;
  editingNodeName.value = '';
}

// 处理编辑输入的键盘事件
function handleEditKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    event.preventDefault();
    saveEditGroup();
  } else if (event.key === 'Escape') {
    event.preventDefault();
    cancelEditGroup();
  }
}

// 使用指定名称创建分组
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

    console.log('创建分组:', newGroup);
    await fetchCreateServerGroup(newGroup);
    ElNotification.success('创建成功');

    // 延迟刷新，确保后端处理完成
    setTimeout(async () => {
      await getGroups();

      // 找到新创建的分组（通过名称和parentId匹配）
      const findNewGroup = (nodes: TreeNode[], targetParentId: number, targetName: string): TreeNode | null => {
        for (const node of nodes) {
          if (node.id === targetParentId && node.children) {
            // 检查这个父节点的子节点中是否有新创建的分组
            for (const child of node.children) {
              if (child.name === targetName) {
                return child;
              }
            }
          }
          // 递归检查子节点
          if (node.children) {
            const found = findNewGroup(node.children, targetParentId, targetName);
            if (found) return found;
          }
        }
        return null;
      };

      const newGroupNode = findNewGroup(groupTree.value, parentId, name);
      if (newGroupNode) {
        // 自动选中新建的分组
        selectedGroupId.value = newGroupNode.id;
        currentNode.value = newGroupNode;

        // 直接进入编辑状态
        editingNodeId.value = newGroupNode.id;
        editingNodeName.value = newGroupNode.name;

        // 聚焦到输入框
        setTimeout(() => {
          if (editInputRef.value) {
            editInputRef.value.focus();
            editInputRef.value.select();
          }
        }, 100);
      }
    }, 300);
  } catch (error) {
    console.error('创建分组失败:', error);
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '创建失败');
    }
  }
}

// 删除分组
function handleDeleteGroup() {
  if (!currentNode.value) return;

  const hasChildren = currentNode.value.children && currentNode.value.children.length > 0;
  const hasServers = currentNode.value.serverCount && currentNode.value.serverCount > 0;

  let message = `确定要删除分组 "${currentNode.value.name}" 吗？`;
  if (hasChildren) {
    message += '\\n\\n注意：该分组包含子分组，删除后子分组也将被删除！';
  }
  if (hasServers) {
    message += '\\n\\n注意：该分组下还有主机，请先移除主机后再删除！';
  }

  if (hasServers) {
    ElNotification.warning('该分组下还有主机，请先移除主机后再删除！');
    contextMenuVisible.value = false;
    return;
  }

  ElMessageBox.confirm(message, '删除确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
    dangerouslyUseHTMLString: true
  })
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
    .catch(() => {
      contextMenuVisible.value = false;
    });
}

// 分页变化
function handlePageChange(page: number) {
  pagination.page = page;
  getServers();
}

// 页面大小变化
function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  getServers();
}

// 全选/取消全选
function handleSelectAll(selection: CMDB.Server[]) {
  selectedIds.value = selection.map(s => s.id);
}

// 选中行变化
function handleSelectionChange(selection: CMDB.Server[]) {
  selectedIds.value = selection.map(s => s.id);
}

// 批量删除
function handleBatchDelete() {
  if (selectedIds.value.length === 0) {
    ElNotification.warning('请先选择要删除的主机');
    return;
  }

  ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 台主机吗？`, '批量删除', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      try {
        await Promise.all(selectedIds.value.map(id => fetchDeleteServer(id)));
        ElNotification.success('删除成功');
        selectedIds.value = [];
        await getServers();
        await getGroups();
      } catch (error) {
        console.error('批量删除失败:', error);
        ElNotification.error('批量删除失败');
      }
    })
    .catch(() => {});
}

// 打开新增对话框
function handleAdd() {
  dialogTitle.value = '创建主机';
  serverType.value = 'normal';
  submitError.value = ''; // 清除错误信息
  Object.assign(serverForm, {
    hostname: '',
    ip: '',
    innerIp: '',
    credentialId: undefined as unknown as number,
    serverType: 'vm',
    groupIds: selectedGroupId.value ? [selectedGroupId.value] : [],
    sshPort: 22,
    remarks: ''
  });
  Object.assign(cloudForm, {
    provider: 'aliyun',
    instanceId: '',
    instanceName: '',
    instanceType: '',
    region: '',
    zone: '',
    chargeType: 'postpay'
  });
  dialogVisible.value = true;
}

// 打开编辑对话框
function handleEdit(row: CMDB.Server) {
  dialogTitle.value = '编辑主机';
  serverType.value = row.cloudInfo ? 'cloud' : 'normal';
  submitError.value = ''; // 清除错误信息
  Object.assign(serverForm, {
    id: row.id,
    hostname: row.hostname,
    ip: row.ip,
    innerIp: row.innerIp,
    credentialId: row.credentialId || undefined as unknown as number,
    serverType: row.serverType,
    groupIds: row.groups?.map(g => g.id) || [],
    sshPort: row.sshPort,
    remarks: row.remarks
  });
  if (row.cloudInfo) {
    Object.assign(cloudForm, {
      provider: row.provider as any,
      instanceId: row.cloudInfo.instanceId,
      instanceName: row.cloudInfo.instanceName,
      instanceType: row.cloudInfo.instanceType,
      region: row.cloudInfo.region,
      zone: row.cloudInfo.zone,
      chargeType: row.cloudInfo.chargeType
    });
  } else {
    Object.assign(cloudForm, {
      provider: 'aliyun',
      instanceId: '',
      instanceName: '',
      instanceType: '',
      region: '',
      zone: '',
      chargeType: 'postpay'
    });
  }
  dialogVisible.value = true;
}

// 保存
async function handleSave() {
  if (!serverFormRef.value) return;

  // 清除之前的错误
  submitError.value = '';

  console.log('[handleSave] 开始保存');

  try {
    await serverFormRef.value.validate();
    console.log('[handleSave] 表单验证通过');

    const formData: CMDB.ServerForm = {
      ...serverForm,
      serverType: serverType.value === 'cloud' ? 'vm' : serverForm.serverType || 'vm',
      cloudInfo:
        serverType.value === 'cloud'
          ? {
              ...cloudForm,
              provider: cloudForm.provider,
              publicIp: serverForm.ip,
              privateIp: serverForm.innerIp
            }
          : null
    };

    console.log('[handleSave] 准备创建/更新主机');

    if (serverForm.id) {
      const result: any = await fetchUpdateServer(serverForm.id, formData);
      console.log('[handleSave] 更新结果:', result);

      // 检查是否是业务错误
      if (result && result.code !== undefined) {
        throw new Error(result.message || '更新失败');
      }

      console.log('[handleSave] 更新成功');
      ElNotification.success('更新成功');
    } else {
      const result: any = await fetchCreateServer(formData);
      console.log('[handleSave] 创建结果:', result);

      // 检查是否是业务错误
      if (result && result.code !== undefined) {
        // 业务错误
        if (result.code === 40001) {
          submitError.value = result.message || '主机名已存在，请使用其他主机名';
          setTimeout(() => {
            const hostnameInput = document.querySelector('.hostname-input input');
            if (hostnameInput instanceof HTMLInputElement) {
              hostnameInput.focus();
            }
          }, 100);
        } else if (result.code === 40002) {
          submitError.value = result.message || 'IP地址已存在，请使用其他IP地址';
          setTimeout(() => {
            const ipInput = document.querySelector('.ip-input input');
            if (ipInput instanceof HTMLInputElement) {
              ipInput.focus();
            }
          }, 100);
        } else {
          ElNotification.error(result.message || '创建失败');
        }
        return; // 不继续执行
      }

      console.log('[handleSave] 创建成功');
      ElNotification.success('创建成功');
    }

    dialogVisible.value = false;
    await getServers();
    await getGroups();
  } catch (error: any) {
    console.log('[handleSave] 捕获到错误:', error);
    ElNotification.error(error.message || '操作失败');
  }
}

// 删除
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
        await getServers();
        await getGroups();
      } catch (error) {
        console.error('删除失败:', error);
        ElNotification.error('删除失败');
      }
    })
    .catch(() => {});
}

// 获取节点类名
function getNodeClass(node: TreeNode) {
  const classes = ['custom-tree-node'];
  if (node.status === 0) classes.push('disabled');
  return classes.join(' ');
}

// 获取状态标签
function getStatusTag(status: string) {
  const statusMap: Record<string, { text: string; type: 'success' | 'warning' | 'danger' | 'info' }> = {
    online: { text: '在线', type: 'success' },
    offline: { text: '离线', type: 'danger' },
    unknown: { text: '未知', type: 'info' }
  };
  return statusMap[status] || { text: status, type: 'info' };
}

// 初始化
onMounted(() => {
  getGroups();
  getSSHCredentials();
  getServers();

  // 添加全局点击监听器来关闭右键菜单
  document.addEventListener('click', handleGlobalClick);
});

onUnmounted(() => {
  // 移除全局点击监听器
  document.removeEventListener('click', handleGlobalClick);
});
</script>

<template>
  <div class="h-full flex gap-16px overflow-hidden">
    <!-- 左侧分组树 -->
    <div class="w-280px flex-shrink-0 flex flex-col">
      <ElCard class="flex-1 flex flex-col" shadow="never">
        <!-- 树头部 -->
        <div class="flex items-center justify-between mb-12px">
          <span class="text-14px font-bold text-gray-700">资产分组</span>
          <div class="flex items-center gap-4px">
            <ElButton link size="small" @click="handleAddRootGroup">
              <icon-mdi-plus class="text-16px" />
            </ElButton>
            <ElButton link size="small" @click="handleRefreshGroups">
              <icon-mdi-refresh class="text-16px" />
            </ElButton>
          </div>
        </div>
        <ElInput
          v-model="groupSearchKeyword"
          placeholder="搜索分组"
          clearable
          size="small"
          class="mb-12px"
        >
          <template #prefix>
            <icon-mdi-magnify class="text-16px text-gray-400" />
          </template>
        </ElInput>

        <!-- 树内容 -->
        <div class="flex-1 overflow-auto" @click="handleContextMenuClose">
          <ElTree
            ref="treeRef"
            v-loading="groupLoading"
            :data="filteredGroupTree"
            node-key="id"
            :props="{ label: 'name', children: 'children' }"
            :highlight-current="true"
            :current-node-key="selectedGroupId"
            :expand-on-click-node="false"
            :default-expand-all="true"
            @node-click="handleNodeClick"
            @node-contextmenu="handleNodeContextMenu"
          >
            <template #default="{ node, data }">
              <div :class="getNodeClass(data)" class="w-full">
                <div class="flex items-center justify-between w-full pr-8px group-node">
                  <!-- 编辑模式 -->
                  <div v-if="editingNodeId === data.id" class="flex items-center gap-6px flex-1 min-w-0">
                    <component
                      :is="'icon-' + data.icon.replace(':', '-')"
                      class="text-16px flex-shrink-0"
                      :style="{ color: data.color }"
                    />
                    <input
                      ref="editInputRef"
                      v-model="editingNodeName"
                      class="edit-input flex-1 min-w-0"
                      @keydown="handleEditKeydown"
                      @blur="saveEditGroup"
                      @click.stop
                    />
                  </div>
                  <!-- 正常显示模式 -->
                  <div v-else class="flex items-center gap-6px flex-1 min-w-0">
                    <component
                      :is="'icon-' + data.icon.replace(':', '-')"
                      class="text-16px flex-shrink-0"
                      :style="{ color: data.color }"
                    />
                    <span class="text-14px truncate">{{ node.label }}</span>
                  </div>
                  <ElTag
                    v-if="data.serverCount !== undefined && editingNodeId !== data.id"
                    size="small"
                    type="info"
                    class="flex-shrink-0"
                  >
                    {{ data.serverCount }}
                  </ElTag>
                </div>
              </div>
            </template>
          </ElTree>
        </div>
      </ElCard>
    </div>

    <!-- 右侧主机列表 -->
    <div class="flex-1 min-w-0 flex flex-col">
      <!-- 主机列表卡片 -->
      <ElCard class="flex-1" shadow="never">
        <template #header>
          <div class="flex items-center justify-between w-full">
            <div class="flex-1">
              <!-- 搜索表单 -->
              <ElForm :model="searchForm" inline>
                <ElFormItem label="主机名">
                  <ElInput
                    v-model="searchForm.hostname"
                    placeholder="请输入主机名"
                    clearable
                    style="width: 160px"
                    @keyup.enter="handleSearch"
                  />
                </ElFormItem>
                <ElFormItem label="IP地址">
                  <ElInput
                    v-model="searchForm.ip"
                    placeholder="请输入IP地址"
                    clearable
                    style="width: 160px"
                    @keyup.enter="handleSearch"
                  />
                </ElFormItem>
                <ElFormItem>
                  <ElButton type="primary" plain @click="handleSearch">
                    <template #icon>
                      <icon-ic-round-search class="text-icon" />
                    </template>
                    搜索
                  </ElButton>
                  <ElButton @click="handleReset">
                    <template #icon>
                      <icon-ic-round-refresh class="text-icon" />
                    </template>
                    重置
                  </ElButton>
                </ElFormItem>
              </ElForm>
            </div>
            <ElSpace direction="horizontal" wrap justify="end">
              <ElButton
                v-if="selectedIds.length > 0"
                type="danger"
                plain
                @click="handleBatchDelete"
              >
                <template #icon>
                  <icon-ic-round-delete class="text-icon" />
                </template>
                批量删除 ({{ selectedIds.length }})
              </ElButton>
              <ElButton type="primary" plain @click="handleAdd">
                <template #icon>
                  <icon-ic-round-plus class="text-icon" />
                </template>
                新增
              </ElButton>
              <ElButton @click="getServers">
                <template #icon>
                  <icon-mdi-refresh class="text-icon" :class="{ 'animate-spin': loading }" />
                </template>
                刷新
              </ElButton>
            </ElSpace>
          </div>
        </template>

        <!-- 列表内容 -->
        <div class="h-[calc(100%-52px)] overflow-auto">
          <ElTable
            v-loading="loading"
            height="100%"
            border
            :data="tableData"
            @selection-change="handleSelectionChange"
            @select-all="handleSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="hostname" label="主机名" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">
                <div class="flex items-center gap-6px">
                  <icon-mdi-server class="text-16px text-gray-400" />
                  <span class="text-blue-600 cursor-pointer hover-underline">{{ row.hostname }}</span>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="ip" label="IP地址" width="140" show-overflow-tooltip />
            <ElTableColumn prop="innerIp" label="内网IP" width="140" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="text-gray-500">{{ row.innerIp || '-' }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="配置" width="120" align="center">
              <template #default="{ row }">
                <span class="text-gray-600">{{ row.cpu }}C / {{ row.memory }}G</span>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="os" label="操作系统" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="text-gray-600">{{ row.os || '-' }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="所属分组" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">
                <ElTag
                  v-for="group in row.groups"
                  :key="group.id"
                  size="small"
                  :style="{ color: group.color, borderColor: group.color }"
                  class="mr-4px"
                >
                  {{ group.name }}
                </ElTag>
                <span v-if="!row.groups || row.groups.length === 0" class="text-gray-400">-</span>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="sshPort" label="SSH端口" width="90" align="center">
              <template #default="{ row }">
                <span class="text-gray-600">{{ row.sshPort || 22 }}</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="状态" width="80" align="center">
              <template #default="{ row }">
                <ElTag :type="getStatusTag(row.status).type" size="small">
                  {{ getStatusTag(row.status).text }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="160" align="center" fixed="right">
              <template #default="{ row }">
                <ElButton type="primary" plain size="small" @click="handleEdit(row)">编辑</ElButton>
                <ElButton type="danger" plain size="small" @click="handleDelete(row)">删除</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
        </div>

        <!-- 分页 -->
        <div v-if="tableData.length > 0" class="mt-16px flex justify-end">
          <ElPagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="total"
            layout="total, sizes, prev, pager, next"
            @current-change="handlePageChange"
            @size-change="handlePageSizeChange"
          />
        </div>
      </ElCard>
    </div>

    <!-- 主机表单对话框 -->
    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="700px">
      <ElForm ref="serverFormRef" :model="serverForm" :rules="serverFormRules" label-width="120px">
        <ElFormItem label="主机类型">
          <ElRadioGroup v-model="serverType">
            <ElRadio value="normal">普通主机</ElRadio>
            <ElRadio value="cloud">云主机</ElRadio>
          </ElRadioGroup>
        </ElFormItem>

        <ElFormItem label="主机名" prop="hostname" class="hostname-input">
          <ElInput v-model="serverForm.hostname" placeholder="请输入主机名" />
          <div v-if="submitError && submitError.includes('主机名')" class="form-error-text" style="color: #f56c6c; font-size: 12px; line-height: 1; padding-top: 4px;">
            {{ submitError }}
          </div>
        </ElFormItem>

        <ElFormItem label="连接IP" prop="ip" class="ip-input">
          <ElInput v-model="serverForm.ip" placeholder="请输入连接IP" />
          <div v-if="submitError && submitError.includes('IP地址')" class="form-error-text" style="color: #f56c6c; font-size: 12px; line-height: 1; padding-top: 4px;">
            {{ submitError }}
          </div>
        </ElFormItem>

        <ElFormItem label="内网IP">
          <ElInput v-model="serverForm.innerIp" placeholder="请输入内网IP（可选）" />
        </ElFormItem>

        <template v-if="serverType === 'cloud'">
          <ElFormItem label="云服务商">
            <ElSelect v-model="cloudForm.provider" style="width: 100%">
              <ElOption label="阿里云" value="aliyun" />
              <ElOption label="腾讯云" value="tencent" />
              <ElOption label="华为云" value="huawei" />
              <ElOption label="AWS" value="aws" />
              <ElOption label="其他" value="other" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="实例名称">
            <ElInput v-model="cloudForm.instanceName" placeholder="请输入实例名称" />
          </ElFormItem>
          <ElFormItem label="实例规格">
            <ElInput v-model="cloudForm.instanceType" placeholder="如: ecs.t6-c1m2.large" />
          </ElFormItem>
          <ElFormItem label="地域">
            <ElInput v-model="cloudForm.region" placeholder="如: cn-hangzhou" />
          </ElFormItem>
          <ElFormItem label="可用区">
            <ElInput v-model="cloudForm.zone" placeholder="如: cn-hangzhou-i" />
          </ElFormItem>
          <ElFormItem label="计费类型">
            <ElSelect v-model="cloudForm.chargeType" style="width: 100%">
              <ElOption label="按量付费" value="postpay" />
              <ElOption label="包年包月" value="prepay" />
            </ElSelect>
          </ElFormItem>
        </template>

        <ElFormItem label="SSH凭证" prop="credentialId">
          <ElSelect v-model="serverForm.credentialId" placeholder="请选择SSH凭证" style="width: 100%">
            <ElOption
              v-for="cred in sshCredentials"
              :key="cred.id"
              :label="`${cred.name} (${cred.username})`"
              :value="cred.id"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="SSH端口">
          <ElInputNumber v-model="serverForm.sshPort" :min="1" :max="65535" style="width: 100%" />
        </ElFormItem>

        <ElFormItem label="所属分组">
          <ElTreeSelect
            v-model="serverForm.groupIds"
            :data="groupTreeForSelect"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            multiple
            show-checkbox
            check-strictly
            placeholder="请选择分组"
            style="width: 100%"
          />
        </ElFormItem>

        <ElFormItem label="备注">
          <ElInput v-model="serverForm.remarks" type="textarea" :rows="3" placeholder="请输入备注信息" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSave">保存</ElButton>
      </template>
    </ElDialog>

    <!-- 右键菜单 -->
    <teleport to="body">
      <transition name="fade">
        <div
          v-if="contextMenuVisible"
          class="context-menu"
          :style="{
            left: contextMenuPosition.x + 'px',
            top: contextMenuPosition.y + 'px'
          }"
          @click.stop="handleContextMenuClose"
        >
          <div class="context-menu-item" @click.stop="handleAddGroup">
            <icon-mdi-plus class="mr-8px" />
            添加分组
          </div>
          <div class="context-menu-item" @click.stop="handleEditGroup">
            <icon-mdi-pencil class="mr-8px" />
            重命名
          </div>
          <div class="context-menu-item danger" @click.stop="handleDeleteGroup">
            <icon-mdi-delete class="mr-8px" />
            删除分组
          </div>
        </div>
      </transition>
    </teleport>

    <!-- 分组创建对话框 -->
    <ElDialog v-model="groupDialogVisible" title="创建分组" width="500px">
      <ElForm ref="groupFormRef" :model="groupFormData" :rules="groupFormRules" label-width="100px">
        <ElFormItem label="分组名称" prop="name">
          <ElInput v-model="groupFormData.name" placeholder="请输入分组名称（2-50个字符）" />
        </ElFormItem>
        <ElFormItem label="父分组" prop="parentId">
          <ElTreeSelect
            v-model="groupFormData.parentId"
            :data="groupTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            placeholder="请选择父分组"
            check-strictly
            :render-after-expand="false"
            style="width: 100%"
          />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="groupDialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSaveGroup">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
.custom-tree-node {
  .group-node {
    &:hover {
      background-color: #f5f7fa;
    }
  }

  &.disabled {
    opacity: 0.5;
  }
}

.edit-input {
  height: 22px;
  line-height: 22px;
  padding: 0 4px;
  border: 1px solid #409eff;
  border-radius: 3px;
  outline: none;
  font-size: 14px;
  color: #303133;
  background: #fff;
  transition: all 0.2s;

  &:focus {
    border-color: #409eff;
    box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
  }

  &::placeholder {
    color: #c0c4cc;
  }
}

.form-error-text {
  color: #f56c6c;
  font-size: 12px;
  line-height: 1;
  padding-top: 4px;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 0;
    margin-right: 12px;
  }

  :deep(.el-form-item__label) {
    font-size: 13px;
    color: #606266;
  }
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 4px 0;
  min-width: 150px;

  &-item {
    padding: 8px 16px;
    cursor: pointer;
    display: flex;
    align-items: center;
    font-size: 14px;
    color: #606266;
    transition: all 0.2s;

    &:hover {
      background-color: #f5f7fa;
      color: #409eff;
    }

    &.danger:hover {
      background-color: #fef0f0;
      color: #f56c6c;
    }
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
