<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue';
import {
  fetchGetServerGroups,
  fetchCreateServerGroup,
  fetchUpdateServerGroup,
  fetchDeleteServerGroup,
  fetchGetServers,
  fetchCreateServer,
  fetchUpdateServer,
  fetchDeleteServer,
  fetchGetSSHCredentials
} from '@/service/api';
import { ElNotification, ElMessageBox, FormInstance, FormRules } from 'element-plus';

defineOptions({ name: 'CmdbGroups' });

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

// 分组树数据
const groupTree = ref<TreeNode[]>([]);
const groupLoading = ref(false);

// Tree 组件引用
const treeRef = ref();

// 当前选中的分组
const selectedGroupId = ref<number | undefined>(undefined);

// 右键菜单状态
const contextMenuVisible = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
const currentNode = ref<TreeNode | null>(null);

// ========== 分组管理 ==========
// 对话框状态
const groupDialogVisible = ref(false);
const groupDialogTitle = ref('');
const groupDialogMode = ref<'create' | 'edit' | 'createChild'>('create');
const groupFormRef = ref<FormInstance>();

// 分组表单数据
const groupFormData = reactive<CMDB.ServerGroupForm>({
  name: '',
  code: '',
  parentId: 0,
  description: '',
  color: '#409EFF',
  icon: 'mdi:folder',
  sortOrder: 0,
  status: 1
});

// 分组表单验证规则
const groupFormRules: FormRules = {
  name: [
    { required: true, message: '请输入分组名称', trigger: 'blur' },
    { min: 2, max: 50, message: '分组名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入分组代码', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '分组代码只能包含字母、数字、下划线和连字符', trigger: 'blur' }
  ]
};

// 颜色选项
const colorOptions = [
  { label: '蓝色', value: '#409EFF' },
  { label: '绿色', value: '#67C23A' },
  { label: '橙色', value: '#E6A23C' },
  { label: '红色', value: '#F56C6C' },
  { label: '紫色', value: '#9C27B0' },
  { label: '青色', value: '#00BCD4' },
  { label: '灰色', value: '#909399' }
];

// 图标选项
const iconOptions = [
  { label: '文件夹', value: 'mdi:folder' },
  { label: '服务器', value: 'mdi:server' },
  { label: '云服务器', value: 'mdi:cloud' },
  { label: '数据库', value: 'mdi:database' },
  { label: '网络', value: 'mdi:lan' },
  { label: '容器', value: 'mdi:docker' },
  { label: '应用', value: 'mdi:application' }
];

// ========== 主机管理 ==========
// 主机列表数据
const serverList = ref<CMDB.Server[]>([]);
const serverLoading = ref(false);
const serverTotal = ref(0);

// 分页信息
const serverPagination = reactive({
  page: 1,
  pageSize: 20
});

// 搜索表单
const searchForm = reactive({
  hostname: '',
  ip: ''
});

// 主机对话框状态
const serverDialogVisible = ref(false);
const serverDialogTitle = ref('');
const serverDialogMode = ref<'create' | 'edit'>('create');
const serverFormRef = ref<FormInstance>();

// 主机类型
const serverType = ref<'normal' | 'cloud'>('normal');

// 主机表单数据
const serverFormData = reactive<CMDB.ServerForm & { groupIds?: number[] }>({
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

// SSH凭证列表
const sshCredentials = ref<CMDB.SSHCredential[]>([]);

// 环境选项
const envOptions = [
  { label: '生产', value: 'prod' },
  { label: '测试', value: 'test' },
  { label: '开发', value: 'dev' }
] as const;

// 主机表单验证规则
const serverFormRules: FormRules = {
  hostname: [
    { required: true, message: '请输入主机名', trigger: 'blur' },
    { min: 2, max: 100, message: '主机名长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  ip: [
    { required: true, message: '请输入连接IP', trigger: 'blur' },
    { pattern: /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/, message: '请输入有效的IP地址', trigger: 'blur' }
  ],
  credentialId: [
    { required: true, message: '请选择SSH凭证', trigger: 'change' }
  ]
};

// ========== 分组树功能 ==========
// 构建分组树
function buildGroupTree(groups: CMDB.ServerGroup[], parentId: number): TreeNode[] {
  const result: TreeNode[] = [];
  groups.forEach(group => {
    if (group.parentId === parentId) {
      const children = buildGroupTree(groups, group.id);
      result.push({
        ...group,
        children: children.length > 0 ? children : undefined,
        serverCount: group.servers?.length || 0
      });
    }
  });
  return result;
}

// 获取分组列表
async function getGroups() {
  groupLoading.value = true;
  try {
    const { data } = await fetchGetServerGroups();
    if (data && Array.isArray(data)) {
      groupTree.value = buildGroupTree(data, 0);
    }
  } catch (error) {
    console.error('获取分组失败:', error);
    ElNotification.error('获取分组失败');
  } finally {
    groupLoading.value = false;
  }
}

// 点击分组节点
function handleNodeClick(data: TreeNode) {
  selectedGroupId.value = data.id;
  serverPagination.page = 1;
  getServers();
}

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

// 添加根分组
function handleAddRootGroup() {
  groupDialogMode.value = 'create';
  groupDialogTitle.value = '添加根分组';
  Object.assign(groupFormData, {
    name: '',
    code: '',
    parentId: 0,
    description: '',
    color: '#409EFF',
    icon: 'mdi:folder',
    sortOrder: 0,
    status: 1
  });
  groupDialogVisible.value = true;
}

// 添加子分组
function handleAddChildGroup() {
  if (!currentNode.value) return;

  groupDialogMode.value = 'createChild';
  groupDialogTitle.value = '添加子分组';
  Object.assign(groupFormData, {
    name: '',
    code: '',
    parentId: currentNode.value.id,
    description: '',
    color: '#409EFF',
    icon: 'mdi:folder',
    sortOrder: 0,
    status: 1
  });
  contextMenuVisible.value = false;
  groupDialogVisible.value = true;
}

// 编辑分组
function handleEditGroup() {
  if (!currentNode.value) return;

  groupDialogMode.value = 'edit';
  groupDialogTitle.value = '编辑分组';
  Object.assign(groupFormData, {
    id: currentNode.value.id,
    name: currentNode.value.name,
    code: currentNode.value.code,
    parentId: currentNode.value.parentId,
    description: currentNode.value.description,
    color: currentNode.value.color,
    icon: currentNode.value.icon,
    sortOrder: currentNode.value.sortOrder,
    status: currentNode.value.status
  });
  contextMenuVisible.value = false;
  groupDialogVisible.value = true;
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

// 保存分组
async function handleSaveGroup() {
  if (!groupFormRef.value) return;

  try {
    await groupFormRef.value.validate();

    if (groupDialogMode.value === 'edit') {
      await fetchUpdateServerGroup(groupFormData.id!, groupFormData);
      ElNotification.success('更新成功');
    } else {
      await fetchCreateServerGroup(groupFormData);
      ElNotification.success('创建成功');
    }

    groupDialogVisible.value = false;
    await getGroups();
  } catch (error) {
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '操作失败');
    }
  }
}

// 生成代码
function handleGenerateCode() {
  if (groupFormData.name) {
    groupFormData.code = groupFormData.name
      .toLowerCase()
      .replace(/[\u4e00-\u9fa5]/g, '')
      .replace(/\s+/g, '-')
      .replace(/[^a-zA-Z0-9_-]/g, '');
  }
}

// ========== 主机管理功能 ==========
// 获取主机列表
async function getServers() {
  serverLoading.value = true;
  try {
    const { data } = await fetchGetServers({
      groupId: selectedGroupId.value,
      hostname: searchForm.hostname || undefined,
      ip: searchForm.ip || undefined,
      page: serverPagination.page,
      pageSize: serverPagination.pageSize
    });
    serverList.value = data?.list || [];
    serverTotal.value = data?.total || 0;
  } catch (error) {
    console.error('获取主机列表失败:', error);
    ElNotification.error('获取主机列表失败');
  } finally {
    serverLoading.value = false;
  }
}

// 获取SSH凭证列表
async function getSSHCredentials() {
  try {
    const { data } = await fetchGetSSHCredentials();
    sshCredentials.value = data || [];
  } catch (err) {
    console.error('获取SSH凭证失败:', err);
  }
}

// 搜索主机
function handleSearch() {
  serverPagination.page = 1;
  getServers();
}

// 重置搜索
function handleReset() {
  searchForm.hostname = '';
  searchForm.ip = '';
  serverPagination.page = 1;
  getServers();
}

// 分页变化
function handlePageChange(page: number) {
  serverPagination.page = page;
  getServers();
}

// 页面大小变化
function handlePageSizeChange(pageSize: number) {
  serverPagination.pageSize = pageSize;
  serverPagination.page = 1;
  getServers();
}

// 添加主机
function handleAddServer() {
  if (!selectedGroupId.value) {
    ElNotification.warning('请先选择一个分组');
    return;
  }

  serverDialogMode.value = 'create';
  serverDialogTitle.value = '添加主机';
  serverType.value = 'normal';
  Object.assign(serverFormData, {
    hostname: '',
    ip: '',
    innerIp: '',
    credentialId: undefined as unknown as number,
    serverType: 'vm',
    groupIds: [selectedGroupId.value],
    sshPort: 22,
    remarks: ''
  });
  Object.assign(cloudForm, {
    provider: 'aliyun',
    instanceName: '',
    instanceType: '',
    region: '',
    zone: '',
    chargeType: 'postpay'
  });
  serverDialogVisible.value = true;
}

// 编辑主机
function handleEditServer(row: CMDB.Server) {
  serverDialogMode.value = 'edit';
  serverDialogTitle.value = '编辑主机';
  serverType.value = row.cloudInfo ? 'cloud' : 'normal';
  Object.assign(serverFormData, {
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
  serverDialogVisible.value = true;
}

// 保存主机
async function handleSaveServer() {
  if (!serverFormRef.value) return;

  try {
    await serverFormRef.value.validate();

    const formData: CMDB.ServerForm = {
      ...serverFormData,
      serverType: serverType.value === 'cloud' ? 'vm' : serverFormData.serverType || 'vm',
      cloudInfo:
        serverType.value === 'cloud'
          ? {
              ...cloudForm,
              provider: cloudForm.provider,
              publicIp: serverFormData.ip,
              privateIp: serverFormData.innerIp
            }
          : null
    };

    if (serverDialogMode.value === 'edit') {
      await fetchUpdateServer(serverFormData.id!, formData);
      ElNotification.success('更新成功');
    } else {
      await fetchCreateServer(formData);
      ElNotification.success('创建成功');
    }

    serverDialogVisible.value = false;
    await getServers();
    await getGroups(); // 更新分组下的主机数量
  } catch (error) {
    if (error && typeof error === 'object' && 'message' in error) {
      ElNotification.error(typeof error.message === 'string' ? error.message : '保存失败');
    }
  }
}

// 删除主机
function handleDeleteServer(row: CMDB.Server) {
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
        await getGroups(); // 更新分组下的主机数量
      } catch (error) {
        console.error('删除失败:', error);
        ElNotification.error('删除失败');
      }
    })
    .catch(() => {});
}

// 获取环境标签
function getEnvTag(env: string) {
  const envMap: Record<string, { text: string; type: any }> = {
    prod: { text: '生产', type: 'danger' },
    test: { text: '测试', type: 'warning' },
    dev: { text: '开发', type: 'info' }
  };
  return envMap[env] || { text: env, type: 'info' };
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

// 获取服务器类型文本
function getServerTypeText(type: string) {
  const typeMap: Record<string, string> = {
    physical: '物理机',
    vm: '虚拟机',
    container: '容器'
  };
  return typeMap[type] || type;
}

// 获取服务器所属分组名称
function getGroupNames(groups?: CMDB.ServerGroup[]) {
  if (!groups || groups.length === 0) return '-';
  return groups.map(g => g.name).join(', ');
}

// 初始化
onMounted(() => {
  getGroups();
  getSSHCredentials();
});
</script>

<template>
  <div class="h-full p-16px">
    <div class="flex gap-16px h-full">
      <!-- 左侧分组树 -->
      <div class="w-320px flex-shrink-0">
        <ElCard class="h-full" shadow="never">
          <template #header>
            <div class="flex justify-between items-center">
              <span class="text-16px font-bold">主机分组</span>
              <ElButton type="primary" size="small" @click="handleAddRootGroup">
                <icon-mdi-plus class="mr-4px" />
                添加
              </ElButton>
            </div>
          </template>

          <div class="h-full overflow-auto" @click="handleContextMenuClose">
            <ElTree
              ref="treeRef"
              v-loading="groupLoading"
              :data="groupTree"
              node-key="id"
              default-expand-all
              :highlight-current="true"
              :current-node-key="selectedGroupId"
              :props="{ label: 'name', children: 'children' }"
              :expand-on-click-node="false"
              @node-click="handleNodeClick"
              @node-contextmenu="handleNodeContextMenu"
            >
              <template #default="{ node, data }">
                <div
                  class="flex items-center justify-between w-full pr-8px group-node"
                  :class="{ 'opacity-50': data.status === 0 }"
                >
                  <div class="flex items-center gap-8px flex-1 min-w-0">
                    <component
                      :is="'icon-' + data.icon.replace(':', '-')"
                      class="text-18px flex-shrink-0"
                      :style="{ color: data.color }"
                    />
                    <span class="truncate">{{ node.label }}</span>
                    <ElTag v-if="data.serverCount" size="small" type="info" class="flex-shrink-0">
                      {{ data.serverCount }}
                    </ElTag>
                  </div>
                </div>
              </template>
            </ElTree>

            <!-- 空状态 -->
            <div
              v-if="!groupLoading && groupTree.length === 0"
              class="flex flex-col items-center justify-center h-400px text-gray-400"
            >
              <icon-mdi-folder-open class="text-64px mb-16px" />
              <p class="text-16px">暂无分组</p>
              <p class="text-14px mt-8px">点击右上角"添加"按钮创建第一个分组</p>
            </div>
          </div>
        </ElCard>
      </div>

      <!-- 右侧主机列表 -->
      <div class="flex-1 min-w-0">
        <ElCard class="h-full" shadow="never">
          <template #header>
            <div class="flex justify-between items-center">
              <span class="text-16px font-bold">
                {{ selectedGroupId ? '分组主机' : '全部主机' }}
              </span>
              <ElButton
                type="primary"
                size="small"
                :disabled="!selectedGroupId"
                @click="handleAddServer"
              >
                <icon-mdi-plus class="mr-4px" />
                添加主机
              </ElButton>
            </div>
          </template>

          <!-- 搜索表单 -->
          <ElForm :model="searchForm" inline class="search-form mb-16px">
            <ElFormItem label="主机名">
              <ElInput v-model="searchForm.hostname" placeholder="请输入主机名" clearable style="width: 200px" />
            </ElFormItem>
            <ElFormItem label="IP地址">
              <ElInput v-model="searchForm.ip" placeholder="请输入IP地址" clearable style="width: 200px" />
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="handleSearch">搜索</ElButton>
              <ElButton @click="handleReset">重置</ElButton>
            </ElFormItem>
          </ElForm>

          <!-- 主机表格 -->
          <ElTable v-loading="serverLoading" :data="serverList" border stripe height="calc(100vh - 480px)">
            <ElTableColumn prop="id" label="ID" width="70" align="center" />
            <ElTableColumn prop="hostname" label="主机名" min-width="150" align="center" show-overflow-tooltip />
            <ElTableColumn prop="ip" label="IP地址" width="140" align="center" />
            <ElTableColumn label="配置" width="140" align="center">
              <template #default="{ row }">
                <span>{{ row.cpu }}C / {{ row.memory }}G / {{ row.disk }}G</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="环境" width="80" align="center">
              <template #default="{ row }">
                <ElTag :type="getEnvTag(row.env).type" size="small">
                  {{ getEnvTag(row.env).text }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="状态" width="80" align="center">
              <template #default="{ row }">
                <ElTag :type="getStatusTag(row.status).type" size="small">
                  {{ getStatusTag(row.status).text }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="类型" width="90" align="center">
              <template #default="{ row }">
                <ElTag size="small">{{ getServerTypeText(row.serverType) }}</ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="os" label="操作系统" min-width="120" align="center" show-overflow-tooltip />
            <ElTableColumn label="所属分组" min-width="120" align="center" show-overflow-tooltip>
              <template #default="{ row }">{{ getGroupNames(row.groups) }}</template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="150" align="center" fixed="right">
              <template #default="{ row }">
                <ElButton type="primary" size="small" @click="handleEditServer(row)">编辑</ElButton>
                <ElButton type="danger" size="small" @click="handleDeleteServer(row)">删除</ElButton>
              </template>
            </ElTableColumn>
          </ElTable>

          <!-- 分页 -->
          <div class="flex justify-end mt-16px">
            <ElPagination
              v-model:current-page="serverPagination.page"
              v-model:page-size="serverPagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="serverTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @current-change="handlePageChange"
              @size-change="handlePageSizeChange"
            />
          </div>

          <!-- 空状态 -->
          <div
            v-if="!serverLoading && serverList.length === 0"
            class="flex flex-col items-center justify-center py-80px text-gray-400"
          >
            <icon-mdi-server-network class="text-64px mb-16px" />
            <p class="text-16px">
              {{ selectedGroupId ? '该分组下暂无主机' : '请选择左侧分组查看主机' }}
            </p>
            <p v-if="selectedGroupId" class="text-14px mt-8px">点击"添加主机"按钮添加第一台主机</p>
          </div>
        </ElCard>
      </div>
    </div>

    <!-- 右键菜单 -->
    <transition>
      <div
        v-if="contextMenuVisible"
        class="context-menu"
        :style="{
          left: contextMenuPosition.x + 'px',
          top: contextMenuPosition.y + 'px'
        }"
        @click.stop
      >
        <div class="context-menu-item" @click="handleAddChildGroup">
          <icon-mdi-plus class="mr-8px" />
          添加子分组
        </div>
        <div class="context-menu-item" @click="handleEditGroup">
          <icon-mdi-pencil class="mr-8px" />
          重命名
        </div>
        <div class="context-menu-item danger" @click="handleDeleteGroup">
          <icon-mdi-delete class="mr-8px" />
          删除分组
        </div>
      </div>
    </transition>

    <!-- 分组表单对话框 -->
    <ElDialog v-model="groupDialogVisible" :title="groupDialogTitle" width="600px">
      <ElForm ref="groupFormRef" :model="groupFormData" :rules="groupFormRules" label-width="100px">
        <ElFormItem label="分组名称" prop="name">
          <ElInput v-model="groupFormData.name" placeholder="请输入分组名称" @blur="handleGenerateCode" />
        </ElFormItem>

        <ElFormItem label="分组代码" prop="code">
          <ElInput v-model="groupFormData.code" placeholder="请输入分组代码（英文）" />
        </ElFormItem>

        <ElFormItem label="分组颜色">
          <div class="flex gap-8px">
            <div
              v-for="color in colorOptions"
              :key="color.value"
              class="color-option"
              :class="{ active: groupFormData.color === color.value }"
              :style="{ backgroundColor: color.value }"
              @click="groupFormData.color = color.value"
            />
          </div>
        </ElFormItem>

        <ElFormItem label="分组图标">
          <div class="flex gap-8px flex-wrap">
            <div
              v-for="icon in iconOptions"
              :key="icon.value"
              class="icon-option"
              :class="{ active: groupFormData.icon === icon.value }"
              @click="groupFormData.icon = icon.value"
            >
              <component :is="'icon-' + icon.value.replace(':', '-')" />
            </div>
          </div>
        </ElFormItem>

        <ElFormItem label="排序">
          <ElInputNumber v-model="groupFormData.sortOrder" :min="0" :max="9999" />
        </ElFormItem>

        <ElFormItem label="状态">
          <ElRadioGroup v-model="groupFormData.status">
            <ElRadio :value="1">启用</ElRadio>
            <ElRadio :value="0">禁用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>

        <ElFormItem label="描述">
          <ElInput v-model="groupFormData.description" type="textarea" :rows="3" placeholder="请输入分组描述" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="groupDialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSaveGroup">保存</ElButton>
      </template>
    </ElDialog>

    <!-- 主机表单对话框 -->
    <ElDialog v-model="serverDialogVisible" :title="serverDialogTitle" width="700px">
      <ElForm ref="serverFormRef" :model="serverFormData" :rules="serverFormRules" label-width="120px">
        <ElFormItem label="主机类型">
          <ElRadioGroup v-model="serverType">
            <ElRadio value="normal">普通主机</ElRadio>
            <ElRadio value="cloud">云主机</ElRadio>
          </ElRadioGroup>
        </ElFormItem>

        <ElFormItem label="主机名" prop="hostname">
          <ElInput v-model="serverFormData.hostname" placeholder="请输入主机名" />
        </ElFormItem>

        <ElFormItem label="连接IP" prop="ip">
          <ElInput v-model="serverFormData.ip" placeholder="请输入连接IP" />
        </ElFormItem>

        <ElFormItem label="内网IP">
          <ElInput v-model="serverFormData.innerIp" placeholder="请输入内网IP（可选）" />
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
          <ElSelect v-model="serverFormData.credentialId" placeholder="请选择SSH凭证" style="width: 100%">
            <ElOption
              v-for="cred in sshCredentials"
              :key="cred.id"
              :label="`${cred.name} (${cred.username})`"
              :value="cred.id"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="SSH端口">
          <ElInputNumber v-model="serverFormData.sshPort" :min="1" :max="65535" style="width: 100%" />
        </ElFormItem>

        <ElFormItem label="所属分组">
          <ElTreeSelect
            v-model="serverFormData.groupIds"
            :data="groupTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            multiple
            show-checkbox
            check-strictly
            placeholder="请选择分组"
            style="width: 100%"
          />
        </ElFormItem>

        <ElFormItem label="备注">
          <ElInput v-model="serverFormData.remarks" type="textarea" :rows="3" placeholder="请输入备注信息" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="serverDialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSaveServer">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
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

.color-option {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s;

  &:hover {
    transform: scale(1.1);
  }

  &.active {
    border-color: #409eff;
    box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
  }
}

.icon-option {
  width: 36px;
  height: 36px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;

  &:hover {
    border-color: #409eff;
    background-color: #ecf5ff;
  }

  &.active {
    border-color: #409eff;
    background-color: #ecf5ff;
    color: #409eff;
  }
}

.group-node {
  &:hover {
    background-color: #f5f7fa;
  }
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
