<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { FormInstance, FormRules } from 'element-plus';
import { ElMessageBox, ElNotification } from 'element-plus';
import { ArrowDown, Delete, Document, Edit, Link, More, Plus, Refresh, RefreshRight } from '@element-plus/icons-vue';
import {
  fetchAssignServerToGroups,
  fetchBatchDeployAgent,
  fetchBatchUninstallAgent,
  fetchCheckConnectPermission,
  fetchConnectServer,
  fetchCreateServer,
  fetchCreateServerGroup,
  fetchDeleteServer,
  fetchDeleteServerGroup,
  fetchDeployAgent,
  fetchGetAgentStatus,
  fetchGetAttributes,
  fetchGetBusinessUnits,
  fetchGetCabinets,
  fetchGetSSHCredentials,
  fetchGetServerAttributes,
  fetchGetServerGroups,
  fetchGetServerRooms,
  fetchGetServerTags,
  fetchGetServers,
  fetchGetSessions,
  fetchRestartAgent,
  fetchSaveServerAttributes,
  fetchSyncServerMetrics,
  fetchTestSSHConnection,
  fetchUninstallAgent,
  fetchUpdateServer,
  fetchUpdateServerGroup
} from '@/service/api';
import { views } from '@/router/elegant/imports';
import { $t } from '@/locales';
import { AlertBadge, MiniTrendChart, ServiceStatusIcon } from '@/components/MonitoringComponents';

defineOptions({ name: 'CmdbServers' });

const router = useRouter();
const currentLoginAccount = ref('');

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
    {
      min: 2,
      max: 50,
      message: '分组名称长度在 2 到 50 个字符',
      trigger: 'blur'
    }
  ],
  parentId: [{ required: true, message: '请选择父分组', trigger: 'change' }]
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

// 新搜索方式（Proxmox风格）
const searchType = ref('hostname'); // hostname | ip | group
const searchKeyword = ref('');

// ========== SSH凭证相关 ==========
const userCredentials = ref<CMDB.SSHCredential[]>([]); // credential_type=user
const systemCredentials = ref<CMDB.SSHCredential[]>([]); // credential_type=system

// ========== 业务系统相关 ==========
const businessUnits = ref<CMDB.BusinessUnit[]>([]);

// ========== 机房机柜相关 ==========
const serverRooms = ref<CMDB.ServerRoom[]>([]);
const cabinets = ref<CMDB.Cabinet[]>([]);

// ========== 标签相关 ==========
const serverTags = ref<CMDB.ServerTag[]>([]);

// SSH终端相关
const sshTerminalVisible = ref(false);
const sshSessionId = ref<number>(0);
const sshWebsocketUrl = ref('');
const sshServerName = ref('');
const sshServerIp = ref('');

// ========== 对话框相关 ==========
const dialogVisible = ref(false);
const dialogTitle = ref('');
const serverFormRef = ref<FormInstance>();
const serverType = ref<'normal' | 'cloud'>('normal');
const submitError = ref(''); // 提交错误信息
const activeCollapse = ref<string[]>([]); // 折叠面板激活项（新增时默认折叠）

// 编辑抽屉
const editDrawerVisible = ref(false);
const editDrawerActiveTab = ref('basic'); // basic | advanced | attributes

// 动态属性相关
const attributeDefinitions = ref<Api.SystemManage.AttributeDefinition[]>([]);
const serverAttributes = ref<Api.SystemManage.ServerAttribute[]>([]);
const loadingAttributes = ref(false);

// 主机表单
const serverForm = reactive<
  CMDB.ServerForm & {
    groupIds?: number[];
    tagIds?: number[];
    roomId?: number;
    cabinetId?: number;
  }
>({
  hostname: '',
  ip: '',
  innerIp: '',
  credentialIds: [],
  systemCredentialId: undefined,
  serverType: 'vm',
  groupIds: [],
  tagIds: [],
  roomId: undefined,
  cabinetId: undefined,
  sshPort: 22,
  remarks: '',
  env: 'test' as CMDB.ServerEnv,
  cpu: 0,
  memory: 0,
  disk: 0,
  os: '',
  businessId: undefined as unknown as number
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
    {
      min: 2,
      max: 100,
      message: '主机名长度在 2 到 100 个字符',
      trigger: 'blur'
    },
    {
      pattern: /^[a-zA-Z0-9.-]+$/,
      message: '主机名只能包含字母、数字、点和连字符',
      trigger: 'blur'
    }
  ],
  ip: [
    { required: true, message: '请输入连接IP', trigger: 'blur' },
    {
      pattern: /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/,
      message: '请输入有效的IP地址',
      trigger: 'blur'
    }
  ],
  credentialIds: [
    {
      required: true,
      type: 'array',
      min: 1,
      message: '请至少选择一个SSH凭证',
      trigger: 'change'
    }
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
      children:
        group.children && group.children.length > 0 ? group.children.map(child => addServerCount(child)) : undefined
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
  const children = groupTree.value[0].children || [];
  console.log('groupTreeForSelect 更新:', children);
  return children;
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

    // 旧的搜索方式（保持兼容）
    if (searchForm.hostname) {
      params.hostname = searchForm.hostname;
    }
    if (searchForm.ip) {
      params.ip = searchForm.ip;
    }

    // 新的搜索方式（Proxmox风格）
    if (searchKeyword.value) {
      switch (searchType.value) {
        case 'hostname':
          params.hostname = searchKeyword.value;
          break;
        case 'ip':
          params.ip = searchKeyword.value;
          break;
        case 'group':
          // 分组搜索暂不支持关键字，仅依赖树选择
          break;
      }
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

// 获取SSH凭证列表（分类加载）
async function getSSHCredentials() {
  try {
    const [userRes, systemRes] = await Promise.all([fetchGetSSHCredentials('user'), fetchGetSSHCredentials('system')]);
    userCredentials.value = userRes.data || [];
    systemCredentials.value = systemRes.data || [];
  } catch (error) {
    console.error('获取SSH凭证失败:', error);
  }
}

// 获取机房列表
async function getServerRooms() {
  try {
    const { data } = await fetchGetServerRooms();
    serverRooms.value = data || [];
  } catch (error) {
    console.error('获取机房列表失败:', error);
  }
}

// 获取机柜列表
async function getCabinets(roomId?: number) {
  try {
    const { data } = await fetchGetCabinets(roomId);
    cabinets.value = data || [];
  } catch (error) {
    console.error('获取机柜列表失败:', error);
  }
}

// 获取标签列表
async function getServerTags() {
  try {
    const { data } = await fetchGetServerTags();
    serverTags.value = data || [];
  } catch (error) {
    console.error('获取标签列表失败:', error);
  }
}

// 加载系统选项（从属性定义获取所有属性）
async function getSystemOptions() {
  try {
    // 加载所有属性，不限制 category
    const { data } = await fetchGetAttributes();
    const attributes = data || [];

    // 按 sortOrder 排序所有属性
    const sortedAttrs = attributes.sort(
      (a: Api.SystemManage.AttributeDefinition, b: Api.SystemManage.AttributeDefinition) => a.sortOrder - b.sortOrder
    );

    // 统一存储到 attributeDefinitions
    attributeDefinitions.value = sortedAttrs;
  } catch (error) {
    console.error('加载系统选项失败:', error);
    // 降级使用空对象，前端会显示提示
    attributeDefinitions.value = [];
  }
}

// 机房变化时加载机柜
function handleRoomChange(roomId: number) {
  serverForm.cabinetId = undefined;
  if (roomId) {
    getCabinets(roomId);
  } else {
    cabinets.value = [];
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

// 创建下拉菜单处理
function handleCreateCommand(command: string) {
  switch (command) {
    case 'add':
      handleAdd();
      break;
    case 'import':
      ElNotification.info('批量导入功能开发中...');
      break;
  }
}

// 批量操作下拉菜单处理
function handleBatchCommand(command: string) {
  switch (command) {
    case 'test-connection':
      handleBatchTestConnection();
      break;
    case 'batch-deploy':
      handleBatchDeploy();
      break;
    case 'batch-uninstall':
      handleBatchUninstall();
      break;
    case 'batch-delete':
      handleBatchDelete();
      break;
  }
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

// 添加主机到当前分组
function handleAddServer() {
  if (!currentNode.value) return;

  // 关闭右键菜单
  contextMenuVisible.value = false;

  // 设置选中的分组ID（用于表单初始化时预选）
  if (currentNode.value.id !== 0) {
    selectedGroupId.value = currentNode.value.id;
    console.log('设置 selectedGroupId:', currentNode.value.id, '分组名称:', currentNode.value.name);
  }

  // 确保分组数据已加载
  if (groupTree.value.length === 0) {
    console.log('分组数据未加载，先加载分组数据');
    getGroups().then(() => {
      setTimeout(() => handleAdd(), 100);
    });
  } else {
    // 打开新增对话框（会自动使用 selectedGroupId 预选分组）
    handleAdd();
  }
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

// 批量部署 Agent

// 批量测试SSH连接
async function handleBatchTestConnection() {
  if (selectedIds.value.length === 0) {
    ElNotification.warning('请先选择要测试连接的主机');
    return;
  }

  const results = {
    success: 0,
    failed: 0,
    details: [] as Array<{
      id: number;
      hostname: string;
      success: boolean;
      message: string;
    }>
  };

  ElNotification.info(`正在测试 ${selectedIds.value.length} 台主机的SSH连接...`);

  // 逐个测试连接（串行以避免过多并发连接）
  for (const serverId of selectedIds.value) {
    const server = tableData.value.find(s => s.id === serverId);
    if (!server) continue;

    try {
      const testResult = await fetchTestSSHConnection(serverId);
      results.details.push({
        id: serverId,
        hostname: server.hostname || server.ip,
        success: testResult.data.success,
        message: testResult.data.message
      });

      if (testResult.data.success) {
        results.success++;
      } else {
        results.failed++;
      }
    } catch (error: any) {
      results.failed++;
      results.details.push({
        id: serverId,
        hostname: server.hostname || server.ip,
        success: false,
        message: error?.response?.data?.message || error?.message || '连接测试失败'
      });
    }
  }

  // 显示结果汇总
  const summary = `SSH连接测试完成：
        成功: ${results.success} 台
        失败: ${results.failed} 台`;

  if (results.failed > 0) {
    const failedServers = results.details
      .filter(d => !d.success)
      .map(d => `- ${d.hostname}: ${d.message}`)
      .join('\n');
    await ElMessageBox.alert(
      `${summary}\n\n以下主机连接失败：\n${failedServers}\n\n请检查主机配置和网络连接。`,
      '连接测试结果',
      { type: 'warning', confirmButtonText: '我知道了' }
    );
  } else {
    ElNotification.success(`${summary}\n所有主机连接测试成功！`);
  }
}

async function handleBatchDeploy() {
  if (selectedIds.value.length === 0) {
    ElNotification.warning('请先选择要部署 Agent 的主机');
    return;
  }

  // 检查选中的主机是否都已配置系统运维凭证
  const serversWithoutCred = tableData.value
    .filter(s => selectedIds.value.includes(s.id))
    .filter(s => !s.systemCredentialId || s.systemCredentialId === 0);

  if (serversWithoutCred.length > 0) {
    const serverNames = serversWithoutCred.map(s => s.hostname || s.ip).join('、');
    await ElMessageBox.alert(
      `以下主机未配置系统运维凭证，无法部署：\n\n${serverNames}\n\n请先编辑主机配置系统运维凭证`,
      '凭证未配置',
      { type: 'warning', confirmButtonText: '我知道了' }
    );
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要为选中的 ${selectedIds.value.length} 台主机部署 Agent 吗？部署过程可能需要几分钟，请耐心等待。`,
      '批量部署确认',
      {
        type: 'info',
        confirmButtonText: '确认部署',
        cancelButtonText: '取消'
      }
    );

    await fetchBatchDeployAgent(selectedIds.value);
    ElNotification.success(`批量部署任务已提交，正在后台执行 ${selectedIds.value.length} 台主机的 Agent 部署`);

    // 对每台主机开始轮询状态
    selectedIds.value.forEach(serverId => {
      pollAgentStatus(serverId, 'running');
    });

    selectedIds.value = [];
  } catch (error: any) {
    if (error === 'cancel') return;
    console.error('批量部署失败:', error);
    const errorMsg = error?.response?.data?.message || error?.message || '批量部署失败';
    ElNotification.error(`批量部署失败：${errorMsg}`);
  }
}

// 批量卸载 Agent
async function handleBatchUninstall() {
  if (selectedIds.value.length === 0) {
    ElNotification.warning('请先选择要卸载 Agent 的主机');
    return;
  }

  // 检查选中的主机是否都已配置系统运维凭证
  const serversWithoutCred = tableData.value
    .filter(s => selectedIds.value.includes(s.id))
    .filter(s => !s.systemCredentialId || s.systemCredentialId === 0);

  if (serversWithoutCred.length > 0) {
    const serverNames = serversWithoutCred.map(s => s.hostname || s.ip).join('、');
    await ElMessageBox.alert(
      `以下主机未配置系统运维凭证，无法卸载：\n\n${serverNames}\n\n请先编辑主机配置系统运维凭证`,
      '凭证未配置',
      { type: 'warning', confirmButtonText: '我知道了' }
    );
    return;
  }

  // 检查是否有主机实际运行着 Agent
  const serversNotRunning = tableData.value
    .filter(s => selectedIds.value.includes(s.id))
    .filter(s => s.agentStatus !== 'running' && s.agentStatus !== 'offline');

  if (serversNotRunning.length === selectedIds.value.length) {
    await ElMessageBox.alert('选中的主机中没有运行中的 Agent，无需卸载。', '提示', {
      type: 'info',
      confirmButtonText: '我知道了'
    });
    return;
  }

  try {
    await ElMessageBox.confirm(
      `确定要从选中的 ${selectedIds.value.length} 台主机卸载 Agent 吗？卸载后主机将不再上报监控数据。`,
      '批量卸载确认',
      {
        type: 'warning',
        confirmButtonText: '确认卸载',
        cancelButtonText: '取消'
      }
    );

    await fetchBatchUninstallAgent(selectedIds.value);
    ElNotification.success(`批量卸载任务已提交，正在后台执行 ${selectedIds.value.length} 台主机的 Agent 卸载`);

    // 对每台主机开始轮询状态
    selectedIds.value.forEach(serverId => {
      pollAgentStatus(serverId, 'uninstalled');
    });

    selectedIds.value = [];
  } catch (error: any) {
    if (error === 'cancel') return;
    console.error('批量卸载失败:', error);
    const errorMsg = error?.response?.data?.message || error?.message || '批量卸载失败';
    ElNotification.error(`批量卸载失败：${errorMsg}`);
  }
}

// 打开新增对话框
function handleAdd() {
  dialogTitle.value = '创建主机';
  serverType.value = 'normal';
  submitError.value = ''; // 清除错误信息
  activeCollapse.value = []; // 默认折叠所有面板

  const groupIds = selectedGroupId.value ? [selectedGroupId.value] : [];
  console.log('handleAdd - selectedGroupId:', selectedGroupId.value, '设置 groupIds:', groupIds);
  console.log('handleAdd - groupTreeForSelect:', JSON.stringify(groupTreeForSelect.value));

  Object.assign(serverForm, {
    hostname: '',
    ip: '',
    innerIp: '',
    credentialIds: [],
    systemCredentialId: undefined,
    serverType: 'vm',
    groupIds: selectedGroupId.value ? [selectedGroupId.value] : [],
    tagIds: [],
    roomId: undefined,
    cabinetId: undefined,
    sshPort: 22,
    remarks: '',
    env: 'test' as CMDB.ServerEnv,
    cpu: 0,
    memory: 0,
    disk: 0,
    os: '',
    businessId: undefined as unknown as number
  });
  // 清空机柜列表
  cabinets.value = [];
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
  // 加载属性定义
  loadAttributeDefinitions();
  // 清空属性值
  serverAttributes.value = [];

  // 确保分组选择器正确显示（调试）
  setTimeout(() => {
    console.log('setTimeout检查 - serverForm.groupIds:', serverForm.groupIds);
    console.log('setTimeout检查 - groupTreeForSelect.value.length:', groupTreeForSelect.value.length);
  }, 200);
}

// 打开编辑抽屉
function handleEdit(row: CMDB.Server) {
  console.log('[handleEdit] row 数据:', row);
  console.log('[handleEdit] row.groups:', row.groups);
  console.log('[handleEdit] groupTreeForSelect:', groupTreeForSelect.value);

  dialogTitle.value = '编辑主机';
  serverType.value = row.cloudInfo ? 'cloud' : 'normal';
  submitError.value = ''; // 清除错误信息
  editDrawerActiveTab.value = 'basic'; // 默认显示基础信息tab

  const groupIds = row.groups?.map(g => g.id) || [];
  console.log('[handleEdit] 计算出的 groupIds:', groupIds);

  Object.assign(serverForm, {
    id: row.id,
    hostname: row.hostname,
    ip: row.ip,
    innerIp: row.innerIp,
    credentialIds:
      row.credentials?.filter(c => c.credentialType === 'user').map(c => c.id) ||
      (row.sshCredentialId ? [row.sshCredentialId] : []),
    systemCredentialId: row.systemCredentialId || row.systemCredential?.id || undefined,
    serverType: row.serverType,
    groupIds,
    tagIds: row.tags?.map(t => t.id) || [],
    roomId: row.cabinet?.roomId,
    cabinetId: row.cabinetId,
    sshPort: row.sshPort,
    remarks: row.remarks,
    env: row.env || 'test',
    cpu: row.cpu || 0,
    memory: row.memory || 0,
    disk: row.disk || 0,
    os: row.os || '',
    businessId: row.businessId || (undefined as unknown as number)
  });
  // 如果有机房，加载对应的机柜列表
  if (row.cabinet?.roomId) {
    getCabinets(row.cabinet.roomId);
  }
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
  editDrawerVisible.value = true; // 打开抽屉
  // 加载属性定义
  loadAttributeDefinitions();
  // 加载主机属性
  loadServerAttributes(row.id);
}

// 加载属性定义
async function loadAttributeDefinitions() {
  try {
    const { data } = await fetchGetAttributes();
    attributeDefinitions.value = data || [];
  } catch (error) {
    console.error('加载属性定义失败:', error);
  }
}

// 加载主机属性
async function loadServerAttributes(serverId: number) {
  loadingAttributes.value = true;
  try {
    const { data } = await fetchGetServerAttributes(serverId);
    serverAttributes.value = data || [];
  } catch (error) {
    console.error('加载主机属性失败:', error);
    serverAttributes.value = [];
  } finally {
    loadingAttributes.value = false;
  }
}

// 打开终端工作台（在新标签页中）
function handleConnect(row: CMDB.Server) {
  // 获取用户凭证ID列表
  const userCredentialIds = row.credentials?.filter(c => c.credentialType === 'user').map(c => c.id) || [];
  const defaultCredentialId = userCredentialIds[0] || row.sshCredentialId;

  const params = new URLSearchParams({
    serverId: row.id.toString(),
    serverName: row.hostname,
    serverIp: row.ip,
    serverEnv: row.env || 'unknown'
  });

  // 如果有默认凭证，传递凭证ID
  if (defaultCredentialId) {
    params.append('credentialId', defaultCredentialId.toString());
  }

  // 在新标签页打开终端工作台
  window.open(`/terminal/workbench?${params.toString()}`, '_blank');
}

function throwIfRequestFailed(result: { error: unknown }) {
  if (result.error) {
    throw result.error;
  }
}

function focusFormInput(selector: string) {
  setTimeout(() => {
    const input = document.querySelector(selector);
    if (input instanceof HTMLInputElement) {
      input.focus();
    }
  }, 100);
}

function handleSaveError(error: any) {
  console.log('[handleSave] 捕获到错误:', error);
  console.log('[handleSave] 错误响应:', error?.response);

  const errorCode = error?.response?.data?.code;
  const errorMessage = error?.response?.data?.message || error?.message;

  console.log('[handleSave] 错误码:', errorCode, '错误消息:', errorMessage);

  if (errorCode === 40001) {
    submitError.value = errorMessage || '主机名已存在，请使用其他主机名';
    focusFormInput('.hostname-input input');
    return;
  }

  if (errorCode === 40002) {
    submitError.value = errorMessage || 'IP地址已存在，请使用其他IP地址';
    focusFormInput('.ip-input input');
    return;
  }

  ElNotification.error(errorMessage || '操作失败，请稍后重试');
}

// 保存主机属性
async function saveServerAttributes(serverId: number) {
  try {
    // 只保存有值的属性
    const attributesToSave = serverAttributes.value
      .filter(attr => attr.attributeValue && attr.attributeValue.trim() !== '')
      .map(attr => ({
        attributeId: attr.attributeId,
        attributeKey: attr.attributeKey,
        attributeValue: attr.attributeValue
      }));

    if (attributesToSave.length > 0) {
      await fetchSaveServerAttributes(serverId, attributesToSave);
    }
  } catch (error) {
    console.error('保存主机属性失败:', error);
    // 不阻塞主流程，只记录错误
  }
}

// 保存
async function handleSave() {
  if (!serverFormRef.value) return;

  // 清除之前的错误
  submitError.value = '';

  try {
    await serverFormRef.value.validate();

    // 保存分组ID列表（在删除前保存）
    const groupIds = (serverForm.groupIds || []).map((id: any) => Number(id));

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

    // 移除不需要发送到后端的字段（这些是前端辅助字段，不是数据库字段）
    delete (formData as any).tagIds; // Tags 通过 many2many 关联表处理
    delete (formData as any).groupIds; // Groups 通过 many2many 关联表处理
    delete (formData as any).roomId; // 机房ID，不是服务器字段

    let savedServerId: number;

    if (serverForm.id) {
      const result = await fetchUpdateServer(serverForm.id, formData);
      throwIfRequestFailed(result);
      savedServerId = serverForm.id;
      // 保存主机属性
      await saveServerAttributes(serverForm.id);
      ElNotification.success('更新成功');
    } else {
      const result = await fetchCreateServer(formData);
      throwIfRequestFailed(result);
      // 保存主机属性（新创建的主机）
      const newServerId = (result as any).data?.id;
      savedServerId = newServerId;
      if (newServerId) {
        await saveServerAttributes(newServerId);
      }
      ElNotification.success('创建成功');
    }

    // 保存主机分组关联
    if (groupIds.length > 0) {
      await fetchAssignServerToGroups(savedServerId, groupIds);
    }

    // 关闭对话框和抽屉
    dialogVisible.value = false;
    editDrawerVisible.value = false;
    await getServers();
    await getGroups();
  } catch (error: any) {
    handleSaveError(error);
  }
}

// 获取属性值
function getAttributeValue(attributeId: number): string {
  const attr = serverAttributes.value.find(a => a.attributeId === attributeId);
  if (!attr) {
    return '';
  }
  return attr.attributeValue || '';
}

// 设置属性值
function setAttributeValue(attributeId: number, value: any): void {
  let attr = serverAttributes.value.find(a => a.attributeId === attributeId);
  if (!attr) {
    const definition = attributeDefinitions.value.find(d => d.id === attributeId);
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
  if (attr) {
    attr.attributeValue = String(value);
  }
}

// 获取多选属性值
function getAttributeMultiValue(attributeId: number) {
  const attr = serverAttributes.value.find(a => a.attributeId === attributeId);
  if (!attr) {
    return [];
  }
  try {
    return JSON.parse(attr.attributeValue || '[]');
  } catch {
    return [];
  }
}

// 设置多选属性值
function setAttributeMultiValue(attributeId: number, values: string[]): void {
  let attr = serverAttributes.value.find(a => a.attributeId === attributeId);
  if (!attr) {
    const definition = attributeDefinitions.value.find(d => d.id === attributeId);
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
  if (attr) {
    attr.attributeValue = JSON.stringify(values);
  }
}

// 获取属性显示名称
function getAttributeDisplayName(attr: Api.SystemManage.ServerAttribute): string {
  const definition = attributeDefinitions.value.find(d => d.id === attr.attributeId);
  if (!definition) return attr.attributeValue;

  // 对于select/multiselect类型，尝试获取选项的标签
  if (definition.type === 'select' || definition.type === 'multiselect') {
    try {
      const options = parseAttributeOptions(definition.options);
      if (definition.type === 'select') {
        const option = options.find((o: any) => o.value === attr.attributeValue);
        return option ? `${definition.name}: ${option.label}` : `${definition.name}: ${attr.attributeValue}`;
      }
      const values = JSON.parse(attr.attributeValue || '[]');
      const labels = values
        .map((v: string) => {
          const option = options.find((o: any) => o.value === v);
          return option ? option.label : v;
        })
        .join(', ');
      return `${definition.name}: ${labels}`;
    } catch {
      return `${definition.name}: ${attr.attributeValue}`;
    }
  }

  return `${definition.name}: ${attr.attributeValue}`;
}

// 解析属性选项
function parseAttributeOptions(optionsStr: string) {
  if (!optionsStr) return [];
  try {
    return JSON.parse(optionsStr);
  } catch {
    return [];
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

// 更多操作
function handleMoreAction(cmd: string, row: CMDB.Server) {
  if (cmd === 'delete') handleDelete(row);
  if (cmd === 'edit') handleEdit(row);
  if (cmd === 'sync-metrics') handleSyncMetrics(row);
  if (cmd === 'monitoring') handleViewMonitoring(row);
  if (cmd === 'detail') handleViewDetail(row);
  if (cmd === 'agent-deploy') handleAgentDeploy(row);
  if (cmd === 'agent-restart') handleAgentRestart(row);
  if (cmd === 'agent-uninstall') handleAgentUninstall(row);
}

async function handleSyncMetrics(row: CMDB.Server) {
  try {
    await fetchSyncServerMetrics(row.id);
    ElNotification.info('采集中，3秒后自动刷新...');
    setTimeout(() => {
      getServers();
    }, 3000);
  } catch {
    ElNotification.error('提交采集任务失败');
  }
}

// Agent 操作（部署/重启/卸载），提交后轮询状态最多 20 次
async function handleAgentDeploy(row: CMDB.Server) {
  try {
    // 先测试SSH连接
    const testResult = await fetchTestSSHConnection(row.id);
    if (!testResult.data.success) {
      await ElMessageBox.alert(`${testResult.data.message}`, '连接失败', {
        type: 'error',
        confirmButtonText: '我知道了'
      });
      return;
    }

    // 连接成功，显示提示并继续部署
    const res = await fetchDeployAgent(row.id);
    ElNotification.success({
      message: `Agent 部署任务已提交，正在后台执行...\nSSH连接测试成功（延迟: ${testResult.data.latency}）`,
      duration: 3000
    });
    pollAgentStatus(row.id, 'running');
  } catch (error: any) {
    console.error('Agent 部署失败:', error);
    // 检查是否是凭证未配置的错误
    const errorMsg = error?.response?.data?.message || error?.message || '部署失败';
    if (errorMsg.includes('系统运维凭证') || errorMsg.includes('systemCredential')) {
      ElNotification.error({
        message: '部署失败：主机未配置系统运维凭证，请先在主机编辑页配置',
        duration: 0,
        customClass: 'agent-error-notification'
      });
    } else if (errorMsg.includes('SSH 连接失败')) {
      ElNotification.error({
        message: `部署失败：无法连接到主机 ${row.hostname || row.ip}`,
        duration: 0
      });
    } else {
      ElNotification.error(`Agent 部署失败：${errorMsg}`);
    }
  }
}

async function handleAgentRestart(row: CMDB.Server) {
  try {
    await fetchRestartAgent(row.id);
    ElNotification.info('Agent 重启任务已提交...');
    pollAgentStatus(row.id, 'running');
  } catch (error: any) {
    console.error('Agent 重启失败:', error);
    const errorMsg = error?.response?.data?.message || error?.message || '重启失败';
    if (errorMsg.includes('SSH 连接失败')) {
      ElNotification.error({
        message: `重启失败：无法连接到主机 ${row.hostname || row.ip}`,
        duration: 0
      });
    } else {
      ElNotification.error(`Agent 重启失败：${errorMsg}`);
    }
  }
}

async function handleAgentUninstall(row: CMDB.Server) {
  try {
    await ElMessageBox.confirm(`确认卸载 ${row.hostname} 上的 Agent？卸载后主机将不再上报监控数据。`, '卸载确认', {
      type: 'warning',
      confirmButtonText: '确认卸载',
      cancelButtonText: '取消'
    });

    // 先测试SSH连接
    const testResult = await fetchTestSSHConnection(row.id);
    if (!testResult.data.success) {
      await ElMessageBox.alert(`${testResult.data.message}`, '连接失败', {
        type: 'error',
        confirmButtonText: '我知道了'
      });
      return;
    }

    // 连接成功，继续卸载
    await fetchUninstallAgent(row.id);
    ElNotification.success({
      message: `Agent 卸载任务已提交，正在后台执行...\nSSH连接测试成功（延迟: ${testResult.data.latency}）`,
      duration: 3000
    });
    pollAgentStatus(row.id, 'uninstalled');
  } catch (error: any) {
    // 用户取消操作
    if (error === 'cancel') {
      return;
    }
    console.error('Agent 卸载失败:', error);
    const errorMsg = error?.response?.data?.message || error?.message || '卸载失败';
    if (errorMsg.includes('系统运维凭证') || errorMsg.includes('systemCredential')) {
      ElNotification.error({
        message: '卸载失败：主机未配置系统运维凭证，请先在主机编辑页配置',
        duration: 0,
        customClass: 'agent-error-notification'
      });
    } else if (errorMsg.includes('SSH 连接失败')) {
      ElNotification.error({
        message: `卸载失败：无法连接到主机 ${row.hostname || row.ip}`,
        duration: 0
      });
    } else {
      ElNotification.error(`Agent 卸载失败：${errorMsg}`);
    }
  }
}

// 轮询 Agent 状态（每3秒，最多20次）
// expectedStatus: 达到该状态则提前终止；超时后无论状态都终止
function pollAgentStatus(serverId: number, expectedStatus: CMDB.Server['agentStatus'] = 'running', maxTimes = 20) {
  let count = 0;
  let previousStatus: string | null = null;
  const timer = setInterval(async () => {
    count++;
    try {
      const res = await fetchGetAgentStatus(serverId);
      const status = res.data?.agentStatus as CMDB.Server['agentStatus'];

      // 实时更新表格中对应行
      const idx = tableData.value.findIndex(s => s.id === serverId);
      if (idx !== -1 && res.data) {
        tableData.value[idx] = {
          ...tableData.value[idx],
          agentStatus: status,
          agentPort: res.data.agentPort,
          agentVersion: res.data.agentVersion,
          lastHeartbeatAt: res.data.lastHeartbeatAt
        };

        // 检测状态变化并通知
        if (previousStatus !== null && previousStatus !== status) {
          if (status === 'running') {
            ElNotification.success(`主机 ${tableData.value[idx].hostname} 的 Agent 已成功部署并运行`);
          } else if (status === 'uninstalled') {
            ElNotification.success(`主机 ${tableData.value[idx].hostname} 的 Agent 已成功卸载`);
          } else if (status === 'failed') {
            ElNotification.warning(`主机 ${tableData.value[idx].hostname} 的 Agent 状态异常，请检查日志`);
          }
        }
        previousStatus = status;
      }

      // 达到预期终态
      if (status === expectedStatus) {
        clearInterval(timer);
        getServers(); // 刷新列表以获取完整数据
        return;
      }

      // 超过最大次数，停止轮询
      if (count >= maxTimes) {
        clearInterval(timer);
        const serverName = tableData.value.find(s => s.id === serverId)?.hostname || serverId;
        if (status !== expectedStatus) {
          ElNotification.warning(
            `主机 ${serverName} 的 Agent 状态轮询超时（当前状态: ${status}），请手动刷新页面查看最新状态`
          );
        }
        getServers(); // 刷新列表以获取完整数据
      }
    } catch (error: any) {
      console.error('轮询 Agent 状态失败:', error);
      // 如果连续失败多次，停止轮询
      if (count >= maxTimes) {
        clearInterval(timer);
        ElNotification.error('轮询 Agent 状态失败，请刷新页面查看最新状态');
        getServers();
      }
    }
  }, 3000);
}

function getUsageColor(value: number = 0): string {
  if (value >= 90) return '#f56c6c';
  if (value >= 70) return '#e6a23c';
  return '#67c23a';
}

// 获取指定属性的选项列表
function getAttributeOptions(key: string): Array<{ label: string; value: string }> {
  const attr = attributeDefinitions.value.find(a => a.key === key);
  if (!attr || !attr.options) return [];

  try {
    return JSON.parse(attr.options);
  } catch {
    return [];
  }
}

// 获取统一属性列表（排除特殊字段 env 和 server_type）
function getUnifiedAttributes() {
  return attributeDefinitions.value.filter(attr => attr.key !== 'env' && attr.key !== 'server_type');
}

// 获取指定属性某个值的显示标签
function getAttributeLabel(key: string, value: string): string {
  if (!value) return '-';

  const options = getAttributeOptions(key);
  const option = options.find(opt => opt.value === value);

  // 如果找不到匹配的选项，返回空（数据不一致）
  return option?.label || '-';
}

// 获取环境的显示信息（标签和颜色）
function getEnvDisplayInfo(envValue: string) {
  const options = getAttributeOptions('server_env');
  const option = options.find(opt => opt.value === envValue);
  const label = option?.label || envValue || '-';

  // 根据值确定标签颜色
  let type: 'success' | 'warning' | 'danger' | 'info' = 'info';
  if (envValue === 'prod') type = 'danger';
  else if (envValue === 'test') type = 'warning';
  else if (envValue === 'dev') type = 'success';

  return { label, type };
}

// 获取主机类型的显示信息
function getServerTypeDisplayInfo(typeValue: string) {
  return getAttributeLabel('server_type', typeValue);
}

// 获取磁盘最大使用率分区
function getMaxDiskPartition(row: CMDB.Server) {
  if (!row.diskPartitions || row.diskPartitions.length === 0) {
    return { usage: row.diskUsage || 0, mount: '/' };
  }
  // 找到使用率最大的分区
  const maxPartition = row.diskPartitions.reduce((max, partition) => (partition.usage > max.usage ? partition : max), {
    usage: 0,
    mount: '/'
  });
  return maxPartition;
}

// 格式化磁盘分区信息用于tooltip
function formatDiskPartitions(row: CMDB.Server) {
  if (!row.diskPartitions || row.diskPartitions.length === 0) {
    return '暂无分区信息';
  }
  return row.diskPartitions
    .sort((a, b) => b.usage - a.usage) // 按使用率降序排列
    .map(p => `${p.mount}: ${p.usage}%`)
    .join('\n');
}

// 格式化磁盘分区数据用于详情抽屉
function formatDiskPartitionsForDrawer(server: CMDB.Server | null) {
  if (!server || !server.diskPartitions || server.diskPartitions.length === 0) {
    return [];
  }
  return server.diskPartitions
    .slice() // 创建副本
    .sort((a, b) => b.usage - a.usage) // 按使用率降序排列
    .map(p => ({
      mount: p.mount,
      usage: p.usage
    }));
}

// 格式化时间显示
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

// 详情抽屉
const drawerVisible = ref(false);
const drawerServer = ref<CMDB.Server | null>(null);
const drawerActiveTab = ref('overview');
const drawerSessions = ref<any[]>([]);
const drawerPermission = ref<{ credentials: CMDB.SSHCredential[] }>({
  credentials: []
});
const drawerLoading = ref(false);

async function handleViewDetail(row: CMDB.Server) {
  drawerServer.value = row;
  drawerVisible.value = true;
  drawerActiveTab.value = 'overview';
  drawerLoading.value = true;
  try {
    const [sessionsRes, permRes] = await Promise.allSettled([
      fetchGetSessions({ serverId: row.id, pageSize: 10, page: 1 }),
      fetchCheckConnectPermission(row.id)
    ]);
    if (sessionsRes.status === 'fulfilled') {
      drawerSessions.value = sessionsRes.value.data?.list || [];
    }
    if (permRes.status === 'fulfilled') {
      drawerPermission.value = permRes.value.data || { credentials: [] };
    }
  } finally {
    drawerLoading.value = false;
  }
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
  getServerRooms();
  getServerTags();
  getSystemOptions(); // 加载系统选项（环境、主机类型等）

  // 加载业务系统列表
  fetchGetBusinessUnits()
    .then(res => {
      businessUnits.value = res.data || [];
    })
    .catch(() => {});

  // 添加全局点击监听器来关闭右键菜单
  document.addEventListener('click', handleGlobalClick);
});

onUnmounted(() => {
  // 移除全局点击监听器
  document.removeEventListener('click', handleGlobalClick);
});
</script>

<template>
  <div class="h-full flex gap-12px overflow-hidden">
    <!-- 左侧分组树 -->
    <div class="group-container w-220px flex flex-col flex-shrink-0">
      <ElCard class="group-tree-card flex flex-col flex-1" shadow="never" body-style="padding: 12px; border-radius: 0;">
        <!-- 树头部 -->
        <div class="mb-8px flex items-center justify-between">
          <span class="text-14px text-gray-700 font-bold">资产分组</span>
          <div class="flex items-center gap-4px">
            <ElButton link size="small" @click="handleAddRootGroup">
              <icon-mdi-plus class="text-16px" />
            </ElButton>
            <ElButton link size="small" @click="handleRefreshGroups">
              <icon-mdi-refresh class="text-16px" />
            </ElButton>
          </div>
        </div>
        <ElInput v-model="groupSearchKeyword" placeholder="搜索分组" clearable size="small" class="mb-12px">
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
                <div class="group-node w-full flex items-center justify-between pr-8px">
                  <!-- 编辑模式 -->
                  <div v-if="editingNodeId === data.id" class="min-w-0 flex flex-1 items-center gap-6px">
                    <component
                      :is="'icon-' + data.icon.replace(':', '-')"
                      class="flex-shrink-0 text-16px"
                      :style="{ color: data.color }"
                    />
                    <input
                      ref="editInputRef"
                      v-model="editingNodeName"
                      class="edit-input min-w-0 flex-1"
                      @keydown="handleEditKeydown"
                      @blur="saveEditGroup"
                      @click.stop
                    />
                  </div>
                  <!-- 正常显示模式 -->
                  <div v-else class="min-w-0 flex flex-1 items-center gap-6px">
                    <component
                      :is="'icon-' + data.icon.replace(':', '-')"
                      class="flex-shrink-0 text-16px"
                      :style="{ color: data.color }"
                    />
                    <span class="truncate text-14px">{{ node.label }}</span>
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
    <div class="min-w-0 flex flex-col flex-1">
      <!-- 头部：搜索和操作按钮（同一行）-->
      <div class="mb-8px flex items-center justify-between gap-12px">
        <!-- 左侧：主操作按钮 -->
        <div class="flex items-center gap-8px">
          <ElDropdown trigger="click" @command="handleCreateCommand">
            <ElButton
              type="success"
              :style="{ backgroundColor: '#00A67D', borderColor: '#00A67D', color: '#fff', borderRadius: '0' }"
            >
              <template #icon>
                <icon-ic-round-plus class="text-icon" />
              </template>
              创建
              <icon-ic-round-keyboard-arrow-down class="ml-4px text-icon" />
            </ElButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="add">新增主机</ElDropdownItem>
                <ElDropdownItem command="import">批量导入</ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
          <ElDropdown trigger="click" :disabled="selectedIds.length === 0" @command="handleBatchCommand">
            <ElButton plain :style="{ borderRadius: '0' }">
              更多操作
              <icon-ic-round-keyboard-arrow-down class="ml-4px text-icon" />
            </ElButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem command="test-connection" :disabled="selectedIds.length === 0">
                  测试连接 ({{ selectedIds.length }})
                </ElDropdownItem>
                <ElDropdownItem command="batch-deploy" :disabled="selectedIds.length === 0">
                  批量部署 ({{ selectedIds.length }})
                </ElDropdownItem>
                <ElDropdownItem command="batch-uninstall" :disabled="selectedIds.length === 0">
                  批量卸载 ({{ selectedIds.length }})
                </ElDropdownItem>
                <ElDropdownItem command="batch-delete" :disabled="selectedIds.length === 0" style="color: #f56c6c">
                  批量删除 ({{ selectedIds.length }})
                </ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
        </div>

        <!-- 右侧：搜索栏和刷新按钮 -->
        <div class="search-inputs flex items-center gap-8px">
          <ElSelect v-model="searchType" placeholder="筛选条件">
            <ElOption label="主机名" value="hostname" />
            <ElOption label="IP地址" value="ip" />
            <ElOption label="分组" value="group" />
          </ElSelect>
          <ElInput
            v-model="searchKeyword"
            placeholder="搜索"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          >
            <template #suffix>
              <icon-ic-round-search class="cursor-pointer text-icon" @click="handleSearch" />
            </template>
          </ElInput>
          <ElButton text @click="getServers">
            <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
          </ElButton>
        </div>
      </div>

      <!-- 主机列表容器 -->
      <div class="flex flex-col flex-1 overflow-hidden bg-white">
        <!-- 列表内容 -->
        <div class="flex-1 overflow-auto">
          <ElTable
            v-loading="loading"
            height="100%"
            :data="tableData"
            size="small"
            class="server-list-table"
            :row-style="{ height: '48px' }"
            :cell-style="{ padding: '0', borderRight: 'none' }"
            :header-cell-style="{ backgroundColor: '#f5f7fa', borderRight: 'none' }"
            stripe
            table-layout="fixed"
            @selection-change="handleSelectionChange"
            @select-all="handleSelectAll"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="hostname" label="主机名" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="cursor-pointer text-primary hover:underline" @click="handleViewDetail(row)">
                  {{ row.hostname }}
                </span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="IP地址" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">
                <div class="ip-list">
                  <div class="ip-row">
                    {{ row.ip }}
                    <span class="ip-tag-outer">外</span>
                  </div>
                  <div v-if="row.innerIp" class="ip-row ip-row-inner">
                    {{ row.innerIp }}
                    <span class="ip-tag-inner">内</span>
                  </div>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn label="配置" width="110" align="center">
              <template #default="{ row }">
                <span class="text-12px" style="color: #606266">{{ row.cpu }}C/{{ row.memory }}G</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="环境" width="75" align="center">
              <template #default="{ row }">
                <ElTag :type="getEnvDisplayInfo(row.env).type" size="small" effect="plain">
                  {{ getEnvDisplayInfo(row.env).label }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="分组" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">
                <ElTag
                  v-if="row.groups && row.groups.length > 0"
                  :style="{ color: row.groups[0].color, borderColor: row.groups[0].color }"
                  size="small"
                >
                  {{ row.groups[0].name }}
                  <span v-if="row.groups.length > 1" class="ml-4px">+{{ row.groups.length - 1 }}</span>
                </ElTag>
                <span v-else class="text-12px text-gray-400">-</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="资源使用率" min-width="200" align="center">
              <template #default="{ row }">
                <template v-if="row.agentStatus === 'offline'">
                  <ElTag type="warning" size="small">Agent离线</ElTag>
                </template>
                <template v-else-if="!['running', 'active'].includes(row.agentStatus) || !row.agentStatus">
                  <ElTag type="info" size="small">未安装</ElTag>
                </template>
                <template v-else-if="row.metricsUpdatedAt">
                  <ElTooltip :content="formatDiskPartitions(row)" placement="top">
                    <div class="usage-compact">
                      <span class="usage-item" :style="{ color: getUsageColor(row.cpuUsage) }">
                        <span class="usage-label">CPU</span>
                        <span class="usage-value">{{ Math.round(row.cpuUsage || 0) }}%</span>
                      </span>
                      <span class="usage-divider">|</span>
                      <span class="usage-item" :style="{ color: getUsageColor(row.memoryUsage) }">
                        <span class="usage-label">内存</span>
                        <span class="usage-value">{{ Math.round(row.memoryUsage || 0) }}%</span>
                      </span>
                      <span class="usage-divider">|</span>
                      <span class="usage-item disk-usage">
                        <span class="usage-label">磁盘</span>
                        <span class="usage-value" :style="{ color: getUsageColor(getMaxDiskPartition(row).usage) }">
                          {{ Math.round(getMaxDiskPartition(row).usage) }}%
                        </span>
                        <span class="disk-mount">({{ getMaxDiskPartition(row).mount }})</span>
                      </span>
                    </div>
                  </ElTooltip>
                </template>
                <template v-else>
                  <ElTooltip content="Agent运行中，等待首次采集" placement="top">
                    <ElTag type="success" size="small">采集中</ElTag>
                  </ElTooltip>
                </template>
              </template>
            </ElTableColumn>
            <!-- CPU趋势列 -->
            <ElTableColumn label="CPU趋势" width="100" align="center">
              <template #default="{ row }">
                <MiniTrendChart
                  v-if="row.cpuTrend && row.cpuTrend.length > 0"
                  :data="row.cpuTrend"
                  :height="30"
                  :color="getUsageColor(row.cpuUsage || 0)"
                />
                <span v-else class="text-12px text-gray-400">-</span>
              </template>
            </ElTableColumn>
            <!-- 服务状态列 -->
            <ElTableColumn label="服务状态" width="100" align="center">
              <template #default="{ row }">
                <ServiceStatusIcon
                  v-if="row.agentStatus === 'running'"
                  :status="row.serviceStatus || 'unknown'"
                  :show-text="true"
                />
                <ElTag v-else-if="row.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
                <ElTag v-else type="info" size="small">未安装</ElTag>
              </template>
            </ElTableColumn>
            <!-- 告警徽章列 -->
            <ElTableColumn label="告警" width="80" align="center">
              <template #default="{ row }">
                <AlertBadge v-if="row.agentStatus === 'running'" :count="row.alertCount || 0" :max-count="99" />
                <span v-else class="text-12px text-gray-400">-</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="100" align="center" fixed="right">
              <template #default="{ row }">
                <div style="display: flex; align-items: center; justify-content: center; gap: 12px">
                  <!-- 连接按钮 -->
                  <a title="连接终端" style="cursor: pointer" @click="handleConnect(row)">
                    <icon-mdi-console-line class="text-14px" style="color: #909399" />
                  </a>

                  <!-- 更多菜单 -->
                  <ElDropdown trigger="click" @command="(cmd: string) => handleMoreAction(cmd, row)">
                    <span
                      style="
                        display: inline-flex;
                        align-items: center;
                        justify-content: center;
                        width: 24px;
                        height: 24px;
                        cursor: pointer;
                        font-size: 16px;
                        color: #909399;
                        font-style: normal;
                        letter-spacing: 1px;
                      "
                    >
                      ⋮
                    </span>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem
                          v-if="row.agentStatus === 'running' || row.agentStatus === 'failed'"
                          command="sync-metrics"
                        >
                          <icon-mdi-refresh class="mr-8px" />
                          刷新指标
                        </ElDropdownItem>
                        <ElDropdownItem v-if="row.agentStatus === 'running'" command="monitoring">
                          <icon-mdi-chart-line class="mr-8px" />
                          监控详情
                        </ElDropdownItem>
                        <ElDropdownItem command="detail">
                          <icon-ic-round-info class="mr-8px" />
                          主机详情
                        </ElDropdownItem>
                        <ElDropdownItem command="edit">
                          <icon-ic-round-edit class="mr-8px" />
                          编辑
                        </ElDropdownItem>
                        <ElDropdownItem
                          v-if="!row.agentStatus || row.agentStatus === 'uninstalled' || row.agentStatus === 'failed'"
                          command="agent-deploy"
                        >
                          <icon-mdi-download class="mr-8px" />
                          部署 Agent
                        </ElDropdownItem>
                        <ElDropdownItem
                          v-if="row.agentStatus === 'running' || row.agentStatus === 'offline'"
                          command="agent-restart"
                        >
                          <icon-mdi-restart class="mr-8px" />
                          重启 Agent
                        </ElDropdownItem>
                        <ElDropdownItem
                          v-if="row.agentStatus === 'running' || row.agentStatus === 'offline'"
                          command="agent-uninstall"
                        >
                          <icon-mdi-delete-forever class="mr-8px" />
                          卸载 Agent
                        </ElDropdownItem>
                        <ElDropdownItem divided command="delete" style="color: #f56c6c">
                          <icon-ic-round-delete class="mr-8px" />
                          删除主机
                        </ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </div>
              </template>
            </ElTableColumn>
          </ElTable>
        </div>

        <!-- 分页 -->
        <div v-if="tableData.length > 0" class="flex justify-end border-t border-gray-200 bg-white p-12px">
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
      </div>
    </div>

    <!-- 主机表单对话框 -->
    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <ElForm ref="serverFormRef" :model="serverForm" :rules="serverFormRules" label-width="100px">
        <!-- ========== 新增模式：只显示必要信息 ========== -->
        <template v-if="!serverForm.id">
          <div class="form-section-title">基础信息（必填）</div>

          <ElFormItem label="主机名" prop="hostname" class="hostname-input">
            <ElInput v-model="serverForm.hostname" placeholder="请输入主机名" />
            <div
              v-if="submitError && submitError.includes('主机名')"
              class="form-error-text"
              style="color: #f56c6c; font-size: 12px; line-height: 1; padding-top: 4px"
            >
              {{ submitError }}
            </div>
          </ElFormItem>

          <ElFormItem label="连接IP" prop="ip" class="ip-input">
            <ElInput v-model="serverForm.ip" placeholder="请输入连接IP" />
            <div
              v-if="submitError && submitError.includes('IP地址')"
              class="form-error-text"
              style="color: #f56c6c; font-size: 12px; line-height: 1; padding-top: 4px"
            >
              {{ submitError }}
            </div>
          </ElFormItem>

          <ElFormItem label="SSH凭证" prop="credentialIds">
            <ElSelect
              v-model="serverForm.credentialIds"
              placeholder="请选择用户连接凭证（可多选）"
              style="width: 100%"
              multiple
              collapse-tags
              collapse-tags-tooltip
            >
              <ElOption
                v-for="cred in userCredentials"
                :key="cred.id"
                :label="`${cred.name}（${cred.username}）`"
                :value="cred.id"
              />
            </ElSelect>
            <div class="mt-4px text-12px text-gray-400">用于堡垒机 SSH 连接，受访问策略约束</div>
          </ElFormItem>

          <ElFormItem label="系统运维凭证">
            <ElSelect
              v-model="serverForm.systemCredentialId"
              placeholder="请选择系统运维凭证（可选）"
              style="width: 100%"
              clearable
            >
              <ElOption
                v-for="cred in systemCredentials"
                :key="cred.id"
                :label="`${cred.name}（${cred.username}）`"
                :value="cred.id"
              />
            </ElSelect>
            <div class="mt-4px text-12px text-gray-400">用于 Agent 部署、重启、指标采集，需 root/sudo 权限</div>
          </ElFormItem>

          <ElFormItem label="环境">
            <ElSelect v-model="serverForm.env" placeholder="请选择环境" style="width: 100%">
              <ElOption
                v-for="opt in getAttributeOptions('env')"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </ElSelect>
          </ElFormItem>

          <ElFormItem label="所属分组" prop="groupIds">
            <ElTreeSelect
              v-model="serverForm.groupIds"
              :data="groupTreeForSelect"
              :props="{ label: 'name', value: 'id', children: 'children' }"
              node-key="id"
              value-key="id"
              multiple
              show-checkbox
              check-strictly
              placeholder="请选择分组（可多选）"
              style="width: 100%"
            />
          </ElFormItem>

          <ElCollapse v-model="activeCollapse" class="mt-16px">
            <ElCollapseItem title="更多属性（可选）" name="attributes">
              <div v-if="loadingAttributes" class="py-12px text-center">
                <ElIcon class="is-loading"><icon-mdi-loading /></ElIcon>
                <span class="ml-8px">加载中...</span>
              </div>
              <div v-else-if="getUnifiedAttributes().length === 0" class="py-12px text-center text-gray-400">
                暂无可用属性，请先在"系统管理 → 属性管理"中配置
              </div>
              <div v-else class="attributes-container">
                <div v-for="attr in getUnifiedAttributes()" :key="attr.id" class="attribute-item">
                  <div class="attribute-label">
                    <span v-if="attr.required" class="required-mark">*</span>
                    {{ attr.name }}
                    <span v-if="attr.description" class="attribute-description">{{ attr.description }}</span>
                  </div>
                  <div class="attribute-input">
                    <!-- 文本输入 -->
                    <ElInput
                      v-if="attr.type === 'text'"
                      :model-value="getAttributeValue(attr.id)"
                      :placeholder="attr.defaultValue || `请输入${attr.name}`"
                      style="width: 100%"
                      @change="val => setAttributeValue(attr.id, val)"
                    />
                    <!-- 数字输入 -->
                    <ElInputNumber
                      v-else-if="attr.type === 'number'"
                      :model-value="getAttributeValue(attr.id)"
                      :placeholder="attr.defaultValue || `请输入${attr.name}`"
                      style="width: 100%"
                      @change="val => setAttributeValue(attr.id, val)"
                    />
                    <!-- 日期选择 -->
                    <ElDatePicker
                      v-else-if="attr.type === 'date'"
                      :model-value="getAttributeValue(attr.id)"
                      type="date"
                      :placeholder="attr.defaultValue || `请选择${attr.name}`"
                      style="width: 100%"
                      format="YYYY-MM-DD"
                      value-format="YYYY-MM-DD"
                      @change="val => setAttributeValue(attr.id, val)"
                    />
                    <!-- 布尔值 -->
                    <ElSwitch
                      v-else-if="attr.type === 'boolean'"
                      :model-value="getAttributeValue(attr.id)"
                      active-text="是"
                      inactive-text="否"
                      @change="val => setAttributeValue(attr.id, val)"
                    />
                    <!-- 下拉单选 -->
                    <ElSelect
                      v-else-if="attr.type === 'select'"
                      :model-value="getAttributeValue(attr.id)"
                      :placeholder="`请选择${attr.name}`"
                      style="width: 100%"
                      @change="val => setAttributeValue(attr.id, val)"
                    >
                      <ElOption
                        v-for="opt in parseAttributeOptions(attr.options)"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </ElSelect>
                    <!-- 下拉多选 -->
                    <ElSelect
                      v-else-if="attr.type === 'multiselect'"
                      :model-value="getAttributeMultiValue(attr.id)"
                      :placeholder="`请选择${attr.name}`"
                      style="width: 100%"
                      multiple
                      collapse-tags
                      collapse-tags-tooltip
                      @change="val => setAttributeMultiValue(attr.id, val)"
                    >
                      <ElOption
                        v-for="opt in parseAttributeOptions(attr.options)"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </ElSelect>
                  </div>
                </div>
              </div>
            </ElCollapseItem>
          </ElCollapse>
        </template>
      </ElForm>

      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSave">保存</ElButton>
      </template>
    </ElDialog>

    <!-- 编辑抽屉（铺满屏幕） -->
    <ElDrawer
      v-model="editDrawerVisible"
      :title="`编辑主机 - ${serverForm.hostname}`"
      direction="rtl"
      size="80%"
      destroy-on-close
    >
      <ElTabs v-model="editDrawerActiveTab" type="border-card">
        <!-- 基础信息 Tab -->
        <ElTabPane label="基础信息" name="basic">
          <ElForm
            ref="serverFormRef"
            :model="serverForm"
            :rules="serverFormRules"
            label-width="100px"
            class="edit-form"
          >
            <div class="edit-form-row">
              <div class="edit-form-col">
                <div class="subsection-title">主机信息</div>
                <ElFormItem label="主机名" prop="hostname" class="hostname-input">
                  <ElInput v-model="serverForm.hostname" placeholder="请输入主机名" />
                  <div v-if="submitError && submitError.includes('主机名')" class="form-error-text">
                    {{ submitError }}
                  </div>
                </ElFormItem>

                <ElFormItem label="连接IP" prop="ip" class="ip-input">
                  <ElInput v-model="serverForm.ip" placeholder="请输入连接IP" />
                  <div v-if="submitError && submitError.includes('IP地址')" class="form-error-text">
                    {{ submitError }}
                  </div>
                </ElFormItem>

                <ElFormItem label="内网IP">
                  <ElInput v-model="serverForm.innerIp" placeholder="请输入内网IP" />
                </ElFormItem>

                <ElFormItem label="SSH端口">
                  <ElInputNumber v-model="serverForm.sshPort" :min="1" :max="65535" style="width: 100%" />
                </ElFormItem>
              </div>

              <div class="edit-form-col">
                <div class="subsection-title">环境与归属</div>
                <ElFormItem label="环境">
                  <ElSelect v-model="serverForm.env" placeholder="请选择环境" style="width: 100%">
                    <ElOption
                      v-for="opt in getAttributeOptions('env')"
                      :key="opt.value"
                      :label="opt.label"
                      :value="opt.value"
                    />
                  </ElSelect>
                </ElFormItem>

                <ElFormItem label="业务系统">
                  <ElTreeSelect
                    v-model="serverForm.businessId"
                    :data="businessUnits"
                    :props="{ label: 'name', value: 'id', children: 'children' }"
                    placeholder="请选择业务系统"
                    clearable
                    check-strictly
                    style="width: 100%"
                  />
                </ElFormItem>
              </div>
            </div>

            <div class="edit-form-row">
              <div class="edit-form-col">
                <div class="subsection-title">凭证配置</div>
                <ElFormItem label="用户连接凭证" prop="credentialIds">
                  <ElSelect
                    v-model="serverForm.credentialIds"
                    placeholder="请选择用户连接凭证"
                    style="width: 100%"
                    multiple
                    collapse-tags
                    collapse-tags-tooltip
                  >
                    <ElOption
                      v-for="cred in userCredentials"
                      :key="cred.id"
                      :label="`${cred.name}（${cred.username}）`"
                      :value="cred.id"
                    />
                  </ElSelect>
                </ElFormItem>

                <ElFormItem label="系统运维凭证">
                  <ElSelect
                    v-model="serverForm.systemCredentialId"
                    placeholder="请选择系统运维凭证"
                    style="width: 100%"
                    clearable
                  >
                    <ElOption
                      v-for="cred in systemCredentials"
                      :key="cred.id"
                      :label="`${cred.name}（${cred.username}）`"
                      :value="cred.id"
                    />
                  </ElSelect>
                </ElFormItem>
              </div>

              <div class="edit-form-col">
                <div class="subsection-title">分组与标签</div>
                <ElFormItem label="所属分组" prop="groupIds">
                  <ElTreeSelect
                    v-model="serverForm.groupIds"
                    :data="groupTreeForSelect"
                    :props="{ label: 'name', value: 'id', children: 'children' }"
                    node-key="id"
                    value-key="id"
                    multiple
                    show-checkbox
                    check-strictly
                    placeholder="请选择分组"
                    style="width: 100%"
                  />
                </ElFormItem>

                <ElFormItem label="标签">
                  <ElSelect
                    v-model="serverForm.tagIds"
                    placeholder="请选择标签"
                    multiple
                    collapse-tags
                    collapse-tags-tooltip
                    style="width: 100%"
                  >
                    <ElOption v-for="tag in serverTags" :key="tag.id" :label="tag.name" :value="tag.id">
                      <span>{{ tag.name }}</span>
                      <span
                        :style="{
                          marginLeft: '8px',
                          display: 'inline-block',
                          width: '12px',
                          height: '12px',
                          borderRadius: '2px',
                          backgroundColor: tag.color
                        }"
                      />
                    </ElOption>
                  </ElSelect>
                </ElFormItem>
              </div>
            </div>

            <div class="edit-form-row">
              <div class="edit-form-col">
                <div class="subsection-title">位置信息</div>
                <ElFormItem label="所在机房">
                  <ElSelect
                    v-model="serverForm.roomId"
                    placeholder="请选择机房"
                    clearable
                    style="width: 100%"
                    @change="handleRoomChange"
                  >
                    <ElOption
                      v-for="room in serverRooms"
                      :key="room.id"
                      :label="`${room.name} (${room.code})`"
                      :value="room.id"
                    />
                  </ElSelect>
                </ElFormItem>

                <ElFormItem label="所在机柜">
                  <ElSelect
                    v-model="serverForm.cabinetId"
                    placeholder="请先选择机房"
                    clearable
                    :disabled="!serverForm.roomId"
                    style="width: 100%"
                  >
                    <ElOption
                      v-for="cabinet in cabinets"
                      :key="cabinet.id"
                      :label="`${cabinet.name} (${cabinet.code})`"
                      :value="cabinet.id"
                    />
                  </ElSelect>
                </ElFormItem>
              </div>

              <div class="edit-form-col">
                <div class="subsection-title">硬件配置</div>
                <ElFormItem label="CPU/内存/磁盘">
                  <div style="display: flex; gap: 8px">
                    <ElInputNumber v-model="serverForm.cpu" :min="0" :max="1024" placeholder="CPU" style="flex: 1" />
                    <ElInputNumber
                      v-model="serverForm.memory"
                      :min="0"
                      :max="65536"
                      placeholder="内存GB"
                      style="flex: 1"
                    />
                    <ElInputNumber
                      v-model="serverForm.disk"
                      :min="0"
                      :max="999999"
                      placeholder="磁盘GB"
                      style="flex: 1"
                    />
                  </div>
                </ElFormItem>

                <ElFormItem label="操作系统">
                  <ElInput v-model="serverForm.os" placeholder="如 Ubuntu 22.04" style="width: 100%" />
                </ElFormItem>

                <ElFormItem label="备注">
                  <ElInput v-model="serverForm.remarks" type="textarea" :rows="2" placeholder="请输入备注" />
                </ElFormItem>
              </div>
            </div>
          </ElForm>
        </ElTabPane>

        <!-- 属性配置 Tab -->
        <ElTabPane label="属性配置" name="attributes">
          <div v-if="loadingAttributes" class="py-40px text-center">
            <ElIcon class="is-loading text-32px"><icon-mdi-loading /></ElIcon>
            <div class="mt-16px">加载中...</div>
          </div>
          <div v-else-if="getUnifiedAttributes().length === 0" class="py-40px text-center text-gray-400">
            <icon-mdi-information-outline class="text-48px" />
            <div class="mt-16px">暂无可用属性</div>
            <div class="mt-8px text-12px">请先在"系统管理 → 属性管理"中配置属性</div>
          </div>
          <div v-else class="attributes-container-drawer">
            <div v-for="attr in getUnifiedAttributes()" :key="attr.id" class="attribute-form-item-compact">
              <div class="attribute-label">
                <span v-if="attr.required" class="required-mark">*</span>
                {{ attr.name }}
                <span v-if="attr.description" class="attribute-description">{{ attr.description }}</span>
              </div>
              <div class="attribute-input">
                <!-- 文本输入 -->
                <ElInput
                  v-if="attr.type === 'text'"
                  :model-value="getAttributeValue(attr.id)"
                  :placeholder="attr.defaultValue || `请输入${attr.name}`"
                  @change="val => setAttributeValue(attr.id, val)"
                />
                <!-- 数字输入 -->
                <ElInputNumber
                  v-else-if="attr.type === 'number'"
                  :model-value="getAttributeValue(attr.id)"
                  :placeholder="attr.defaultValue || `请输入${attr.name}`"
                  style="width: 100%"
                  @change="val => setAttributeValue(attr.id, val)"
                />
                <!-- 日期选择 -->
                <ElDatePicker
                  v-else-if="attr.type === 'date'"
                  :model-value="getAttributeValue(attr.id)"
                  type="date"
                  :placeholder="attr.defaultValue || `请选择${attr.name}`"
                  style="width: 100%"
                  format="YYYY-MM-DD"
                  value-format="YYYY-MM-DD"
                  @change="val => setAttributeValue(attr.id, val)"
                />
                <!-- 布尔值 -->
                <ElSwitch
                  v-else-if="attr.type === 'boolean'"
                  :model-value="getAttributeValue(attr.id)"
                  active-text="是"
                  inactive-text="否"
                  @change="val => setAttributeValue(attr.id, val)"
                />
                <!-- 下拉单选 -->
                <ElSelect
                  v-else-if="attr.type === 'select'"
                  :model-value="getAttributeValue(attr.id)"
                  :placeholder="`请选择${attr.name}`"
                  style="width: 100%"
                  @change="val => setAttributeValue(attr.id, val)"
                >
                  <ElOption
                    v-for="opt in parseAttributeOptions(attr.options)"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </ElSelect>
                <!-- 下拉多选 -->
                <ElSelect
                  v-else-if="attr.type === 'multiselect'"
                  :model-value="getAttributeMultiValue(attr.id)"
                  :placeholder="`请选择${attr.name}`"
                  style="width: 100%"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  @change="val => setAttributeMultiValue(attr.id, val)"
                >
                  <ElOption
                    v-for="opt in parseAttributeOptions(attr.options)"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </ElSelect>
              </div>
            </div>
          </div>
        </ElTabPane>

        <!-- 云主机配置 Tab -->
        <ElTabPane label="云主机配置" name="cloud" :disabled="serverType !== 'cloud'">
          <div v-if="serverType === 'cloud'" class="edit-form">
            <ElForm label-width="100px">
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
            </ElForm>
          </div>
          <div v-else class="py-40px text-center text-gray-400">
            <icon-mdi-cloud-off-outline class="text-48px" />
            <div class="mt-16px">当前主机不是云主机，无需配置云服务信息</div>
          </div>
        </ElTabPane>
      </ElTabs>

      <template #footer>
        <div style="flex: 1"></div>
        <ElButton size="large" @click="editDrawerVisible = false">取消</ElButton>
        <ElButton type="primary" size="large" @click="handleSave">保存更改</ElButton>
      </template>
    </ElDrawer>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <Transition name="fade">
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
          <div class="context-menu-item primary" @click.stop="handleAddServer">
            <icon-mdi-server class="mr-8px" />
            添加主机
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
      </Transition>
    </Teleport>

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

    <!-- SSH终端已改为路由跳转，保留连接对话框 -->

    <!-- 主机详情抽屉 -->
    <ElDrawer
      v-model="drawerVisible"
      direction="rtl"
      size="820px"
      :title="drawerServer?.hostname || '主机详情'"
      destroy-on-close
    >
      <div v-if="drawerServer">
        <ElTabs v-model="drawerActiveTab">
          <!-- 概览 Tab -->
          <ElTabPane label="概览" name="overview">
            <!-- 基础信息 -->
            <div class="overview-section">
              <div class="section-title">基础信息</div>
              <ElDescriptions :column="2" border size="small">
                <ElDescriptionsItem label="主机名">{{ drawerServer.hostname }}</ElDescriptionsItem>
                <ElDescriptionsItem label="连接IP">{{ drawerServer.ip }}</ElDescriptionsItem>
                <ElDescriptionsItem label="内网IP">{{ drawerServer.innerIp || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="SSH端口">{{ drawerServer.sshPort || 22 }}</ElDescriptionsItem>
                <ElDescriptionsItem label="操作系统">{{ drawerServer.os || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="系统架构">{{ drawerServer.arch || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="环境">
                  <ElTag :type="getEnvDisplayInfo(drawerServer.env).type" size="small">
                    {{ getEnvDisplayInfo(drawerServer.env).label }}
                  </ElTag>
                </ElDescriptionsItem>
                <ElDescriptionsItem label="主机状态">
                  <ElTag :type="drawerServer.status === 1 ? 'success' : 'info'" size="small">
                    {{ drawerServer.status === 1 ? '正常' : '停用' }}
                  </ElTag>
                </ElDescriptionsItem>
              </ElDescriptions>
            </div>

            <!-- 硬件配置 -->
            <div class="overview-section">
              <div class="section-title">硬件配置</div>
              <ElDescriptions :column="3" border size="small">
                <ElDescriptionsItem label="CPU">
                  {{ drawerServer.cpu ? `${drawerServer.cpu} 核` : '-' }}
                </ElDescriptionsItem>
                <ElDescriptionsItem label="内存">
                  {{ drawerServer.memory ? `${drawerServer.memory} GB` : '-' }}
                </ElDescriptionsItem>
                <ElDescriptionsItem label="磁盘">
                  {{ drawerServer.disk ? `${drawerServer.disk} GB` : '-' }}
                </ElDescriptionsItem>
              </ElDescriptions>
            </div>

            <!-- 云主机信息（仅云主机显示） -->
            <div v-if="drawerServer.serverType === 'cloud' && drawerServer.cloudInfo" class="overview-section">
              <div class="section-title">云主机信息</div>
              <ElDescriptions :column="2" border size="small">
                <ElDescriptionsItem label="服务商">
                  <ElTag size="small">{{ drawerServer.provider || '-' }}</ElTag>
                </ElDescriptionsItem>
                <ElDescriptionsItem label="实例类型">
                  {{ drawerServer.cloudInfo.instanceType || '-' }}
                </ElDescriptionsItem>
                <ElDescriptionsItem label="区域">{{ drawerServer.region || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="可用区">{{ drawerServer.zone || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="实例ID" :span="2">
                  {{ drawerServer.cloudInfo.instanceId || '-' }}
                </ElDescriptionsItem>
              </ElDescriptions>
            </div>

            <!-- 归属信息 -->
            <div class="overview-section">
              <div class="section-title">归属信息</div>
              <ElDescriptions :column="2" border size="small">
                <ElDescriptionsItem label="业务系统">{{ drawerServer.business?.name || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="所属分组">
                  <ElTag v-for="group in drawerServer.groups" :key="group.id" size="small" style="margin-right: 4px">
                    {{ group.name }}
                  </ElTag>
                  <span v-if="!drawerServer.groups?.length" class="text-gray-400">未分组</span>
                </ElDescriptionsItem>
                <ElDescriptionsItem label="标签" :span="2">
                  <ElTag
                    v-for="tag in drawerServer.tags"
                    :key="tag.id"
                    size="small"
                    :color="tag.color"
                    style="margin-right: 4px"
                  >
                    {{ tag.name }}
                  </ElTag>
                  <span v-if="!drawerServer.tags?.length" class="text-gray-400">无标签</span>
                </ElDescriptionsItem>
              </ElDescriptions>
            </div>

            <!-- 状态信息 -->
            <div class="overview-section">
              <div class="section-title">状态信息</div>
              <ElDescriptions :column="2" border size="small">
                <ElDescriptionsItem label="Agent 状态">
                  <ElTag v-if="drawerServer.agentStatus === 'running'" type="success" size="small">
                    运行中 (v{{ drawerServer.agentVersion || '-' }})
                  </ElTag>
                  <ElTag v-else-if="drawerServer.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
                  <ElTag v-else type="info" size="small">未安装</ElTag>
                </ElDescriptionsItem>
                <ElDescriptionsItem label="最近连接">{{ drawerServer.lastConnectTime || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="最后连通性检查">
                  {{ drawerServer.lastCheckTime || '-' }}
                </ElDescriptionsItem>
                <ElDescriptionsItem label="连通性状态">
                  <ElTag
                    :type="
                      drawerServer.connectivityStatus === 'online'
                        ? 'success'
                        : drawerServer.connectivityStatus === 'offline'
                          ? 'danger'
                          : 'info'
                    "
                    size="small"
                  >
                    {{
                      drawerServer.connectivityStatus === 'online'
                        ? '在线'
                        : drawerServer.connectivityStatus === 'offline'
                          ? '离线'
                          : '未知'
                    }}
                  </ElTag>
                </ElDescriptionsItem>
                <ElDescriptionsItem label="创建时间" :span="2">{{ drawerServer.createdAt || '-' }}</ElDescriptionsItem>
                <ElDescriptionsItem label="备注" :span="2">{{ drawerServer.remarks || '-' }}</ElDescriptionsItem>
              </ElDescriptions>
            </div>
          </ElTabPane>

          <!-- 连接 Tab -->
          <ElTabPane label="连接信息" name="connect">
            <ElDescriptions :column="1" border>
              <ElDescriptionsItem label="SSH端口">{{ drawerServer.sshPort || 22 }}</ElDescriptionsItem>
              <ElDescriptionsItem label="绑定凭证">
                <div v-if="drawerServer.credentials?.length">
                  <ElTag
                    v-for="cred in drawerServer.credentials"
                    :key="cred.id"
                    type="success"
                    size="small"
                    style="margin-right: 4px; margin-bottom: 4px"
                  >
                    {{ cred.name }}（{{ cred.username }}）
                  </ElTag>
                </div>
                <ElTag v-else type="warning" size="small">未配置</ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="可用凭证">
                <div>
                  <ElTag
                    v-for="cred in drawerPermission.credentials"
                    :key="cred.id"
                    size="small"
                    style="margin-right: 4px; margin-bottom: 4px"
                  >
                    {{ cred.name }}（{{ cred.username }}）
                  </ElTag>
                  <span v-if="!drawerPermission.credentials?.length" class="text-gray-400">暂无数据</span>
                </div>
              </ElDescriptionsItem>
            </ElDescriptions>
            <div style="margin-top: 16px">
              <ElButton type="success" @click="handleConnect(drawerServer)">
                <icon-mdi-console-line style="margin-right: 4px" />
                连接此主机
              </ElButton>
            </div>
          </ElTabPane>

          <!-- 会话记录 Tab -->
          <ElTabPane label="会话记录" name="sessions">
            <ElTable v-loading="drawerLoading" :data="drawerSessions" size="small">
              <ElTableColumn prop="username" label="用户" width="100" />
              <ElTableColumn prop="loginAccount" label="登录账号" width="100" />
              <ElTableColumn prop="protocol" label="协议" width="70">
                <template #default="{ row }">
                  <ElTag size="small">{{ row.protocol?.toUpperCase() }}</ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn prop="startedAt" label="开始时间" min-width="150" />
              <ElTableColumn prop="status" label="状态" width="80">
                <template #default="{ row }">
                  <ElTag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status }}</ElTag>
                </template>
              </ElTableColumn>
            </ElTable>
          </ElTabPane>

          <!-- 监控信息 Tab -->
          <ElTabPane label="监控信息" name="monitor">
            <!-- 资源使用率卡片 -->
            <div class="monitor-section">
              <div class="section-title">资源使用率</div>
              <div class="usage-cards">
                <div class="usage-card">
                  <div class="card-label">CPU</div>
                  <div class="card-value" :style="{ color: getUsageColor(drawerServer.cpuUsage || 0) }">
                    {{ Math.round(drawerServer.cpuUsage || 0) }}%
                  </div>
                  <ElProgress
                    :percentage="Math.round(drawerServer.cpuUsage || 0)"
                    :color="getUsageColor(drawerServer.cpuUsage || 0)"
                    :show-text="false"
                  />
                  <div class="card-detail">{{ drawerServer.cpu || '-' }} 核</div>
                </div>
                <div class="usage-card">
                  <div class="card-label">内存</div>
                  <div class="card-value" :style="{ color: getUsageColor(drawerServer.memoryUsage || 0) }">
                    {{ Math.round(drawerServer.memoryUsage || 0) }}%
                  </div>
                  <ElProgress
                    :percentage="Math.round(drawerServer.memoryUsage || 0)"
                    :color="getUsageColor(drawerServer.memoryUsage || 0)"
                    :show-text="false"
                  />
                  <div class="card-detail">
                    {{
                      drawerServer.memory
                        ? ((drawerServer.memory * (drawerServer.memoryUsage || 0)) / 100).toFixed(1) +
                          ' / ' +
                          drawerServer.memory +
                          ' GB'
                        : '-'
                    }}
                  </div>
                </div>
                <div class="usage-card">
                  <div class="card-label">磁盘</div>
                  <div class="card-value" :style="{ color: getUsageColor(drawerServer.diskUsage || 0) }">
                    {{ Math.round(drawerServer.diskUsage || 0) }}%
                  </div>
                  <ElProgress
                    :percentage="Math.round(drawerServer.diskUsage || 0)"
                    :color="getUsageColor(drawerServer.diskUsage || 0)"
                    :show-text="false"
                  />
                  <div class="card-detail">{{ drawerServer.disk || '-' }} GB</div>
                </div>
              </div>
            </div>

            <!-- 磁盘分区详情 -->
            <div class="monitor-section">
              <div class="section-title">磁盘分区详情</div>
              <ElTable :data="formatDiskPartitionsForDrawer(drawerServer)" size="small" border>
                <ElTableColumn prop="mount" label="挂载点" width="120" />
                <ElTableColumn label="使用率" width="150">
                  <template #default="{ row }">
                    <ElProgress :percentage="Math.round(row.usage)" :color="getUsageColor(row.usage)" />
                  </template>
                </ElTableColumn>
                <ElTableColumn prop="usage" label="使用率" width="80" align="center">
                  <template #default="{ row }">
                    <span :style="{ color: getUsageColor(row.usage) }">{{ Math.round(row.usage) }}%</span>
                  </template>
                </ElTableColumn>
              </ElTable>
              <div
                v-if="!drawerServer.diskPartitions || drawerServer.diskPartitions.length === 0"
                class="py-12px text-center text-gray-400"
              >
                {{ drawerServer.agentStatus === 'running' ? '正在采集...' : '暂无数据，请先部署 Agent' }}
              </div>
            </div>

            <!-- Agent 状态 -->
            <div class="monitor-section">
              <div class="section-title">Agent 状态</div>
              <ElDescriptions :column="2" border size="small">
                <ElDescriptionsItem label="状态">
                  <template #default>
                    <div v-if="drawerServer.agentStatus === 'running'" class="flex items-center gap-2">
                      <ElTag type="success" size="small">运行中</ElTag>
                      <span v-if="drawerServer.agentVersion" class="text-12px text-gray-500">
                        v{{ drawerServer.agentVersion }}
                      </span>
                    </div>
                    <ElTag v-else-if="drawerServer.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
                    <ElTag v-else type="info" size="small">未安装</ElTag>
                  </template>
                </ElDescriptionsItem>
                <ElDescriptionsItem label="监听端口">{{ drawerServer.agentPort || 9100 }}</ElDescriptionsItem>
                <ElDescriptionsItem label="最后心跳">
                  <template #default>
                    <span v-if="drawerServer.lastHeartbeatAt">{{ formatTime(drawerServer.lastHeartbeatAt) }}</span>
                    <span v-else class="text-gray-400">-</span>
                  </template>
                </ElDescriptionsItem>
                <ElDescriptionsItem label="指标更新">
                  <template #default>
                    <span v-if="drawerServer.metricsUpdatedAt">{{ formatTime(drawerServer.metricsUpdatedAt) }}</span>
                    <span v-else class="text-gray-400">-</span>
                  </template>
                </ElDescriptionsItem>
              </ElDescriptions>
            </div>
          </ElTabPane>
        </ElTabs>
      </div>
    </ElDrawer>
  </div>
</template>

<style scoped lang="scss">
/* 去掉资产分组容器的圆角 */
.group-container {
  border-radius: 0 !important;

  * {
    border-radius: 0 !important;
  }
}

/* 去掉资产分组Card的所有圆角 */
.group-tree-card {
  :deep(.el-card) {
    border-radius: 0 !important;
  }

  :deep(.el-card__header) {
    border-radius: 0 !important;
  }

  :deep(.el-card__body) {
    border-radius: 0 !important;
  }

  :deep(.el-tree) {
    border-radius: 0 !important;
  }

  /* 强制去掉所有子元素的圆角 */
  * {
    border-radius: 0 !important;
  }
}

/* 去掉搜索框和下拉框的圆角 */
.search-inputs {
  :deep(.el-select__wrapper) {
    border-radius: 0 !important;
  }

  :deep(.el-input__wrapper) {
    border-radius: 0 !important;
  }
}

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

.search-form-compact {
  @extend .search-form;

  :deep(.el-form-item) {
    margin-right: 8px;
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

/* 动态属性样式 */
.attributes-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.attribute-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.attribute-label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 4px;
}

.required-mark {
  color: #f56c6c;
  font-size: 14px;
}

.attribute-description {
  font-size: 12px;
  color: #909399;
  font-weight: normal;
  margin-left: 8px;
}

.attribute-input {
  width: 100%;
}

/* ========== 表单优化样式 ========== */
.form-section-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 2px solid #e4e7ed;
}

.hostname-input,
.ip-input {
  margin-bottom: 8px;
}

.form-error-text {
  animation: shake 0.5s;
}

@keyframes shake {
  0%,
  100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-4px);
  }
  75% {
    transform: translateX(4px);
  }
}

.advanced-options {
  padding: 16px;
  background-color: #f9fafb;
  border-radius: 8px;
}

/* ========== 编辑抽屉样式 ========== */
.edit-form {
  padding: 16px;
}

.edit-form-row {
  display: flex;
  gap: 20px;
  margin-bottom: 16px;
}

.edit-form-col {
  flex: 1;
  min-width: 0;
}

.subsection-title {
  font-size: 13px;
  font-weight: 600;
  color: #409eff;
  margin-bottom: 10px;
  padding-bottom: 4px;
  border-bottom: 1px solid #dcdfe6;
}

.form-section {
  margin-bottom: 32px;
  padding: 20px;
  background-color: #fafafa;
  border-radius: 8px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 2px solid #e4e7ed;
}

.field-hint {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}

.attribute-form-item {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  margin-bottom: 16px;
  background-color: #fafafa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.attribute-form-item .attribute-label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 6px;
}

.attribute-form-item .attribute-input {
  width: 100%;
}

/* 属性网格布局 */
.attributes-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px 20px;
  padding: 12px;
}

/* 属性配置标签页容器 */
.attributes-container-drawer {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px 20px;
  padding: 20px;
}

.attributes-container-drawer .attribute-form-item-compact {
  margin: 0;
}

.attribute-form-item-compact {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background-color: #fafafa;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

.attribute-form-item-compact .attribute-label {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 4px;
}

.attribute-form-item-compact .attribute-input {
  width: 100%;
}

/* 概览区块样式 */
.overview-section {
  margin-bottom: 24px;
}

.overview-section .section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e4e7ed;
}

/* 资源使用率紧凑显示 */
.usage-compact {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  white-space: nowrap;
  width: 100%;
}

.usage-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.usage-label {
  color: #909399;
  font-size: 11px;
}

.usage-value {
  font-weight: 600;
  font-size: 12px;
  min-width: 32px;
  text-align: center;
}

.usage-divider {
  color: #dcdfe6;
  margin: 0 2px;
}

.disk-mount {
  font-size: 10px;
  color: #909399;
  margin-left: 2px;
}

/* 操作列分隔符 */
.action-divider {
  color: #dcdfe6;
  margin: 0 8px;
  font-size: 12px;
}

/* IP地址显示样式 - 简洁云平台风格 */
.ip-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.ip-row {
  font-size: 12px;
  color: #606266;
  font-family:
    ui-monospace, 'SF Mono', Menlo, Monaco, 'Cascadia Code', 'Roboto Mono', 'Consolas', 'Courier New', monospace;
  font-weight: 400;
  line-height: 1.3;
}

.ip-row-inner {
  color: #606266;
}

.ip-tag-outer,
.ip-tag-inner {
  font-size: 10px;
  color: #909399;
  margin-left: 6px;
  opacity: 0.6;
  font-weight: normal;
  letter-spacing: 0.5px;
}

/* 统一表格字体大小 */
.server-list-table {
  :deep(.el-table__cell) {
    font-size: 12px;
  }
  :deep(.el-table__header .cell) {
    font-size: 12px;
    font-weight: 500;
  }
  :deep(.el-tag) {
    font-size: 12px;
  }
}

/* 监控信息包装器 */
.monitor-info-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.agent-version-tag {
  font-size: 11px;
}

/* 监控信息 Tab 样式 */
.monitor-section {
  margin-bottom: 24px;
}

.monitor-section .section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e4e7ed;
}

.usage-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.usage-card {
  padding: 16px;
  background-color: #fafafa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  text-align: center;
}

.usage-card .card-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.usage-card .card-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 12px;
}

.usage-card .card-detail {
  font-size: 12px;
  color: #606266;
  margin-top: 8px;
}

/* 操作列样式 */
.actions-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.action-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  font-size: 12px;
  color: #409eff;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
  text-decoration: none;

  &:hover {
    background-color: #ecf5ff;
    color: #66b1ff;
  }
}

.connect-link {
  color: #909399;

  &:hover {
    background-color: #f5f7fa;
    color: #606266;
  }
}

.more-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;

  &:hover {
    background-color: #f5f7fa;
  }
}

.more-icon {
  font-size: 16px;
  color: #909399;
  font-style: normal;
  letter-spacing: 1px;
}
</style>
