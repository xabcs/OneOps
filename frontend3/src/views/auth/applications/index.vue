<script setup lang="tsx">
import { onMounted, ref, watch } from 'vue';
import { ArrowDown, Plus, Refresh, Search } from '@element-plus/icons-vue';
import {
  createApplication,
  deleteApplication,
  fetchAppTypeConfigTemplate,
  fetchApplicationAuthorizationRules,
  fetchApplicationGroups,
  fetchApplicationRoles,
  fetchApplicationUsers,
  fetchApplications,
  fetchSupportedAppTypes,
  syncApplicationAuthorizationRules,
  syncApplicationGroups,
  syncApplicationRoles,
  syncApplicationUsers,
  updateApplication
} from '@/service/api/application-permission';

defineOptions({ name: 'AuthCenterApplications' });

const loading = ref(false);
const tableData = ref<Api.ApplicationPermission.Application[]>([]);
const total = ref(0);
const searchName = ref('');

const pagination = ref({
  page: 1,
  pageSize: 10
});

const drawerVisible = ref(false);
const isEdit = ref(false);

// 角色和用户查看对话框
const rolesDialogVisible = ref(false);
const usersDialogVisible = ref(false);
const groupsDialogVisible = ref(false);
const rulesDialogVisible = ref(false);
const applicationRoles = ref<Api.ApplicationPermission.ApplicationRole[]>([]);
const applicationUsers = ref<Api.ApplicationPermission.ApplicationUser[]>([]);
const applicationGroups = ref<Api.ApplicationPermission.ApplicationGroup[]>([]);
const authorizationRules = ref<Api.ApplicationPermission.AuthorizationRule[]>([]);
const currentAppName = ref('');

// 支持的应用类型
const supportedAppTypes = ref<Array<{ type: string; displayName: string }>>([]);
const loadingAppTypes = ref(false);

// 当前应用类型的配置模板
const currentConfigTemplate = ref<Record<string, unknown>>({});
const loadingConfigTemplate = ref(false);

// 默认端点配置
const defaultEndpoints: Api.ApplicationPermission.ApplicationEndpoints = {
  createUser: '/api/users',
  listUsers: '/api/users',
  getUser: '/api/users/{username}',
  getRoles: '/api/roles',
  assignRole: '/api/users/{username}/roles',
  revokeRole: '/api/users/{username}/roles/{roleCode}',
  getUserRoles: '/api/users/{username}/roles'
};

const formData = ref<Api.ApplicationPermission.ApplicationFormData>({
  id: 0,
  name: '',
  code: '',
  type: '',
  baseUrl: '',
  endpoints: { ...defaultEndpoints },
  authConfig: {
    type: 'token',
    token: ''
  },
  description: '',
  syncInterval: 60
});

const authTypeOptions = [
  { label: 'API Token', value: 'token' },
  { label: 'Basic Auth', value: 'basic' },
  { label: 'OAuth 2.0', value: 'oauth2' }
];

// 获取支持的应用类型
async function getSupportedAppTypes() {
  loadingAppTypes.value = true;
  try {
    const { data, error } = await fetchSupportedAppTypes();
    if (!error && data) {
      supportedAppTypes.value = data || [];
    }
  } finally {
    loadingAppTypes.value = false;
  }
}

// 监听应用类型变化，加载配置模板
watch(
  () => formData.value.type,
  async newType => {
    if (newType) {
      await loadConfigTemplate(newType);
    }
  }
);

// 加载应用类型的配置模板
async function loadConfigTemplate(appType: string) {
  loadingConfigTemplate.value = true;
  try {
    const { data, error } = await fetchAppTypeConfigTemplate(appType);
    if (!error && data && data.configTemplate) {
      currentConfigTemplate.value = data.configTemplate;

      // 根据模板自动填充配置
      if (data.configTemplate.endpoints) {
        formData.value.endpoints = { ...data.configTemplate.endpoints };
      }
      if (data.configTemplate.token !== undefined) {
        formData.value.authConfig = {
          type: 'token',
          token: data.configTemplate.token || ''
        };
      }
      if (data.configTemplate.username !== undefined) {
        formData.value.authConfig = {
          type: 'basic',
          username: data.configTemplate.username || '',
          password: data.configTemplate.password || ''
        };
      }
      if (data.configTemplate.syncInterval) {
        formData.value.syncInterval = data.configTemplate.syncInterval;
      }
    }
  } finally {
    loadingConfigTemplate.value = false;
  }
}

// 根据应用类型判断是否需要显示 endpoints 配置
function needsEndpointsConfig() {
  const type = formData.value.type;
  return type === 'generic' || type === 'jumpserver' || type === 'gitlab';
}

async function getData() {
  loading.value = true;
  try {
    const { data, error } = await fetchApplications({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      name: searchName.value
    });
    if (!error && data) {
      tableData.value = data.list || [];
      total.value = data.total || 0;
    }
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.value.page = 1;
  getData();
}

function handleReset() {
  searchName.value = '';
  pagination.value.page = 1;
  getData();
}

function handleAdd() {
  isEdit.value = false;
  formData.value = {
    id: 0,
    name: '',
    code: '',
    type: '',
    baseUrl: '',
    endpoints: { ...defaultEndpoints },
    authConfig: {
      type: 'token',
      token: ''
    },
    description: '',
    syncInterval: 60
  };
  currentConfigTemplate.value = {};
  drawerVisible.value = true;
}

function handleEdit(row: Api.ApplicationPermission.Application) {
  isEdit.value = true;
  // 解析JSON字符串为对象
  const endpoints = typeof row.endpoints === 'string' ? JSON.parse(row.endpoints) : row.endpoints;
  const authConfig = typeof row.authConfig === 'string' ? JSON.parse(row.authConfig) : row.authConfig;

  formData.value = {
    ...row,
    endpoints,
    authConfig
  };
  drawerVisible.value = true;
}

async function handleSubmit() {
  // 将对象转为JSON字符串
  const submitData = {
    ...formData.value,
    endpoints: JSON.stringify(formData.value.endpoints),
    authConfig: JSON.stringify(formData.value.authConfig)
  };

  const api = isEdit.value ? updateApplication(submitData.id!, submitData) : createApplication(submitData);

  const { error } = await api;
  if (!error) {
    ElMessage.success(isEdit.value ? '更新成功' : '添加成功');
    drawerVisible.value = false;
    getData();
  }
}

async function handleDelete(id: number) {
  const { error } = await deleteApplication(id);
  if (!error) {
    ElMessage.success('删除成功');
    getData();
  }
}

async function handleSyncRoles(row: Api.ApplicationPermission.Application) {
  const { error } = await syncApplicationRoles(row.id);
  if (!error) {
    ElMessage.success(`同步 ${row.name} 角色成功`);
    getData();
    // 同步成功后自动显示角色列表
    handleViewRoles(row);
  }
}

async function handleSyncUsers(row: Api.ApplicationPermission.Application) {
  const { error } = await syncApplicationUsers(row.id);
  if (!error) {
    ElMessage.success(`同步 ${row.name} 用户成功`);
    getData();
    // 同步成功后自动显示用户列表
    handleViewUsers(row);
  }
}

async function handleViewRoles(row: Api.ApplicationPermission.Application) {
  currentAppName.value = row.name;
  const { data, error } = await fetchApplicationRoles(row.id);
  if (!error && data) {
    applicationRoles.value = data;
    rolesDialogVisible.value = true;
  } else {
    ElMessage.warning('暂无角色数据，请先同步角色');
  }
}

async function handleViewUsers(row: Api.ApplicationPermission.Application) {
  currentAppName.value = row.name;
  const { data, error } = await fetchApplicationUsers(row.id);
  if (!error && data) {
    applicationUsers.value = data;
    usersDialogVisible.value = true;
  } else {
    ElMessage.warning('暂无用户数据，请先同步用户');
  }
}

async function handleSyncGroups(row: Api.ApplicationPermission.Application) {
  const { error } = await syncApplicationGroups(row.id);
  if (!error) {
    ElMessage.success(`同步 ${row.name} 用户组成功`);
    getData();
    // 同步成功后自动显示用户组列表
    handleViewGroups(row);
  }
}

async function handleViewGroups(row: Api.ApplicationPermission.Application) {
  currentAppName.value = row.name;
  const { data, error } = await fetchApplicationGroups(row.id);
  if (!error && data) {
    if (data.length === 0) {
      ElMessage.info('该应用暂无用户组数据，可能该版本不支持用户组功能或端点配置错误');
    }
    applicationGroups.value = data;
    groupsDialogVisible.value = true;
  } else {
    ElMessage.warning('暂无用户组数据，请先同步用户组');
  }
}

async function handleSyncRules(row: Api.ApplicationPermission.Application) {
  const { error } = await syncApplicationAuthorizationRules(row.id);
  if (!error) {
    ElMessage.success(`同步 ${row.name} 授权规则成功`);
    getData();
    // 同步成功后自动显示授权规则列表
    handleViewRules(row);
  }
}

async function handleViewRules(row: Api.ApplicationPermission.Application) {
  currentAppName.value = row.name;
  const { data, error } = await fetchApplicationAuthorizationRules(row.id);
  if (!error && data) {
    if (data.length === 0) {
      ElMessage.info('该应用暂无授权规则数据，可能该应用不支持授权规则或端点配置错误');
    }
    authorizationRules.value = data;
    rulesDialogVisible.value = true;
  } else {
    ElMessage.warning('暂无授权规则数据，请先同步授权规则');
  }
}

function handleCommand(command: string, row: Api.ApplicationPermission.Application) {
  switch (command) {
    case 'viewUsers':
      handleViewUsers(row);
      break;
    case 'viewGroups':
      handleViewGroups(row);
      break;
    case 'viewRules':
      handleViewRules(row);
      break;
    case 'viewRoles':
      handleViewRoles(row);
      break;
    case 'syncRoles':
      handleSyncRoles(row);
      break;
    case 'edit':
      handleEdit(row);
      break;
    case 'delete':
      handleDelete(row.id);
      break;
  }
}

function handlePageChange(page: number) {
  pagination.value.page = page;
  getData();
}

function handleSizeChange(size: number) {
  pagination.value.pageSize = size;
  pagination.value.page = 1;
  getData();
}

// 授权规则操作权限辅助方法
function getActionLabel(action: string): string {
  const actionLabels: Record<string, string> = {
    connect: '连接',
    upload: '上传',
    download: '下载',
    command: '命令',
    all: '全部'
  };
  return actionLabels[action] || action;
}

function getActionType(action: string): string {
  const actionTypes: Record<string, string> = {
    connect: '',
    upload: 'warning',
    download: 'info',
    command: 'success',
    all: 'danger'
  };
  return actionTypes[action] || '';
}

onMounted(() => {
  getData();
  getSupportedAppTypes();
});
</script>

<template>
    <div class="min-h-500px flex-col gap-4">
        <!-- 搜索栏 -->
        <ElCard shadow="never">
            <ElForm :inline="true">
                <ElFormItem label="应用名称">
                    <ElInput v-model="searchName" placeholder="请输入应用名称" clearable class="w-200px" @keyup.enter="handleSearch" />
                </ElFormItem>
                <ElFormItem>
                    <ElButton type="primary" :icon="Search" @click="handleSearch">搜索</ElButton>
                    <ElButton :icon="Refresh" @click="handleReset">重置</ElButton>
                </ElFormItem>
            </ElForm>
        </ElCard>

        <!-- 表格 -->
        <ElCard shadow="never" class="flex-1">
            <template #header>
                <div class="flex items-center justify-between">
                    <span class="text-lg font-medium">应用列表</span>
                    <ElButton type="primary" :icon="Plus" @click="handleAdd">添加应用</ElButton>
                </div>
            </template>

            <ElTable v-loading="loading" :data="tableData" :border="false">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="name" label="应用名称" align="center" min-width="120" />
                <ElTableColumn prop="code" label="应用代码" align="center" min-width="100" />
                <ElTableColumn prop="type" label="应用类型" align="center" min-width="100" />
                <ElTableColumn prop="baseUrl" label="Base URL" align="center" min-width="200" />
                <ElTableColumn prop="status" label="状态" align="center" width="80">
                    <template #default="{ row }">
                        <ElTag :type="row.status === 1 ? 'success' : 'info'">
                            {{ row.status === 1 ? '启用' : '禁用' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="lastSyncTime" label="最后同步时间" align="center" min-width="160" />
                <ElTableColumn label="操作" align="center" width="320" fixed="right">
                    <template #default="{ row }">
                        <ElSpace wrap>
                            <ElButton size="small" type="primary" @click="handleSyncUsers(row)">同步用户</ElButton>
                            <ElButton size="small" type="success" @click="handleSyncGroups(row)">同步用户组</ElButton>
                            <ElButton v-if="row.type === 'jumpserver'" size="small" type="warning" @click="handleSyncRules(row)">
                                同步授权规则
                            </ElButton>
                            <ElDropdown @command="(cmd: string) => handleCommand(cmd, row)">
                                <ElButton size="small">
                                    更多
                                    <ElIcon class="el-icon--right">
                                        <ArrowDown />
                                    </ElIcon>
                                </ElButton>
                                <template #dropdown>
                                    <ElDropdownMenu>
                                        <ElDropdownItem command="viewUsers">查看用户</ElDropdownItem>
                                        <ElDropdownItem command="viewGroups">查看用户组</ElDropdownItem>
                                        <ElDropdownItem v-if="row.type === 'jumpserver'" command="viewRules">查看授权规则</ElDropdownItem>
                                        <ElDropdownItem v-if="row.type !== 'jumpserver'" command="viewRoles">查看角色</ElDropdownItem>
                                        <ElDropdownItem v-if="row.type !== 'jumpserver'" command="syncRoles">同步角色</ElDropdownItem>
                                        <ElDropdownItem divided command="edit">编辑</ElDropdownItem>
                                        <ElDropdownItem command="delete" style="color: #f56c6c">删除</ElDropdownItem>
                                    </ElDropdownMenu>
                                </template>
                            </ElDropdown>
                        </ElSpace>
                    </template>
                </ElTableColumn>
            </ElTable>

            <div class="mt-4 flex justify-end">
                <ElPagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" @size-change="handleSizeChange" @current-change="handlePageChange" />
            </div>
        </ElCard>

        <!-- 添加/编辑抽屉 -->
        <ElDrawer v-model="drawerVisible" :title="isEdit ? '编辑应用' : '添加应用'" :width="600">
            <ElForm :model="formData" label-width="120px">
                <ElFormItem label="应用名称" required>
                    <ElInput v-model="formData.name" placeholder="请输入应用名称" />
                </ElFormItem>
                <ElFormItem label="应用代码" required>
                    <ElInput v-model="formData.code" placeholder="请输入应用代码（唯一标识）" />
                </ElFormItem>
                <ElFormItem label="应用类型" required>
                    <ElSelect v-model="formData.type" placeholder="请选择应用类型" class="w-full" :loading="loadingAppTypes" @change="loadConfigTemplate(formData.type)">
                        <ElOption v-for="item in supportedAppTypes" :key="item.type" :label="item.displayName" :value="item.type" />
                    </ElSelect>
                </ElFormItem>
                <ElFormItem label="Base URL" required>
                    <ElInput v-model="formData.baseUrl" placeholder="请输入应用访问地址，如：https://jenkins.example.com" />
                </ElFormItem>

                <ElDivider content-position="left">认证配置</ElDivider>

                <!-- Jenkins 专用: Basic Auth -->
                <template v-if="formData.type === 'jenkins'">
                    <ElFormItem label="用户名" required>
                        <ElInput v-model="formData.authConfig.username" placeholder="请输入 Jenkins 用户名" />
                    </ElFormItem>
                    <ElFormItem label="密码" required>
                        <ElInput v-model="formData.authConfig.password" type="password" show-password placeholder="请输入 Jenkins 密码或 Token" />
                    </ElFormItem>
                </template>

                <!-- Jumpserver 专用: 支持三种认证方式 -->
                <template v-if="formData.type === 'jumpserver'">
                    <ElFormItem label="认证方式" required>
                        <ElSelect v-model="formData.authConfig.authType" placeholder="请选择认证方式" class="w-full">
                            <ElOption label="AccessKey + Secret（推荐）" value="accessKey" />
                            <ElOption label="Private Token" value="token" />
                            <ElOption label="用户名 + 密码" value="basic" />
                        </ElSelect>
                    </ElFormItem>

                    <!-- AccessKey + Secret 签名认证 -->
                    <template v-if="formData.authConfig.authType === 'accessKey'">
                        <ElAlert type="info" :closable="false" class="mb-4">
                            <template #title>
                                <div class="text-sm">
                                    <strong>AccessKey 签名认证说明：</strong>
                                    <ul class="ml-4 mt-2 list-disc">
                                        <li>需要在 JumpServer 中创建「服务账号」获取密钥</li>
                                        <li>使用 HTTP Signature 签名算法，最安全可靠</li>
                                        <li>长期有效，不会过期</li>
                                    </ul>
                                </div>
                            </template>
                        </ElAlert>
                        <ElFormItem label="Access Key ID" required>
                            <ElInput v-model="formData.authConfig.accessKey" placeholder="请输入 Access Key ID" />
                            <span class="text-xs text-gray-500">在 JumpServer 系统设置 → 账号管理 → 服务账号中获取</span>
                        </ElFormItem>
                        <ElFormItem label="Access Key Secret" required>
                            <ElInput v-model="formData.authConfig.secret" type="password" show-password placeholder="请输入 Access Key Secret" />
                            <span class="text-xs text-gray-500">密钥只显示一次，请妥善保管</span>
                        </ElFormItem>
                        <ElFormItem label="组织ID（可选）">
                            <ElInput v-model="formData.authConfig.orgId" placeholder="默认组织: 00000000-00000000-00000000-000000000002" />
                            <span class="text-xs text-gray-500">不填写则使用默认组织</span>
                        </ElFormItem>
                    </template>

                    <!-- Private Token 认证 -->
                    <template v-if="formData.authConfig.authType === 'token'">
                        <ElFormItem label="Private Token" required>
                            <ElInput v-model="formData.authConfig.token" type="password" show-password placeholder="请输入 Private Token" />
                            <span class="text-xs text-gray-500">在 JumpServer 个人中心 → API Token 获取（长期有效）</span>
                        </ElFormItem>
                    </template>

                    <!-- 用户名密码认证 -->
                    <template v-if="formData.authConfig.authType === 'basic'">
                        <ElFormItem label="用户名" required>
                            <ElInput v-model="formData.authConfig.username" placeholder="请输入用户名" />
                        </ElFormItem>
                        <ElFormItem label="密码" required>
                            <ElInput v-model="formData.authConfig.password" type="password" show-password placeholder="请输入密码" />
                        </ElFormItem>
                    </template>
                </template>

                <!-- GitLab 专用: API Token -->
                <template v-if="formData.type === 'gitlab'">
                    <ElFormItem label="API Token" required>
                        <ElInput v-model="formData.authConfig.token" type="password" show-password placeholder="请输入 API Token" />
                    </ElFormItem>
                </template>

                <!-- 通用应用: 可选择认证方式 -->
                <template v-if="formData.type === 'generic' || formData.type === ''">
                    <ElFormItem label="认证方式" required>
                        <ElSelect v-model="formData.authConfig.type" placeholder="请选择认证方式" class="w-full">
                            <ElOption v-for="item in authTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
                        </ElSelect>
                    </ElFormItem>

                    <ElFormItem v-if="formData.authConfig.type === 'token'" label="API Token" required>
                        <ElInput v-model="formData.authConfig.token" type="password" show-password placeholder="请输入API Token" />
                    </ElFormItem>

                    <template v-if="formData.authConfig.type === 'basic'">
                        <ElFormItem label="用户名" required>
                            <ElInput v-model="formData.authConfig.username" placeholder="请输入用户名" />
                        </ElFormItem>
                        <ElFormItem label="密码" required>
                            <ElInput v-model="formData.authConfig.password" type="password" show-password placeholder="请输入密码" />
                        </ElFormItem>
                    </template>
                </template>

                <!-- API端点配置 - 仅对需要端点配置的应用显示 -->
                <ElDivider v-if="needsEndpointsConfig()" content-position="left">API端点配置</ElDivider>

                <!-- Jumpserver 端点配置 -->
                <template v-if="formData.type === 'jumpserver'">
                    <ElAlert type="info" :closable="false" class="mb-4">
                        <template #title>
                            <div class="text-sm">
                                <strong>JumpServer 授权说明：</strong>
                                <ul class="ml-4 mt-2 list-disc">
                                    <li>JumpServer 不需要同步角色，授权通过授权规则管理</li>
                                    <li>用户授权需在 JumpServer 管理界面中配置授权规则（用户/用户组 → 资产/节点）</li>
                                </ul>
                            </div>
                        </template>
                    </ElAlert>
                    <ElFormItem label="获取用户列表">
                        <ElInput v-model="formData.endpoints.getUsers" placeholder="/api/v1/users/users/" />
                    </ElFormItem>
                    <ElFormItem label="获取用户组列表">
                        <ElInput v-model="formData.endpoints.getGroups" placeholder="/api/v1/users/groups/" />
                    </ElFormItem>
                    <ElFormItem label="创建用户">
                        <ElInput v-model="formData.endpoints.createUser" placeholder="/api/v1/users/users/" />
                    </ElFormItem>
                </template>

                <!-- Generic 应用端点配置 -->
                <template v-if="formData.type === 'generic'">
                    <ElFormItem label="获取用户列表">
                        <ElInput v-model="formData.endpoints.getUsers" placeholder="/api/v1/users/users/" />
                    </ElFormItem>
                    <ElFormItem label="获取角色列表">
                        <ElInput v-model="formData.endpoints.getRoles" placeholder="/api/v1/perms/roles/" />
                    </ElFormItem>
                    <ElFormItem label="创建用户">
                        <ElInput v-model="formData.endpoints.createUser" placeholder="/api/v1/users/users/" />
                    </ElFormItem>
                    <ElFormItem label="分配角色">
                        <ElInput v-model="formData.endpoints.assignRole" placeholder="/api/v1/perms/user-grantings/" />
                    </ElFormItem>

                    <!-- 添加隐藏字段确保所有必需字段都存在 -->
                    <input v-model="formData.endpoints.listUsers" type="hidden" value="/api/v1/users/users/" />
                    <input v-model="formData.endpoints.getUser" type="hidden" value="/api/v1/users/{username}" />
                    <input v-model="formData.endpoints.revokeRole" type="hidden" value="/api/v1/users/{username}/roles/{roleCode}" />
                    <input v-model="formData.endpoints.getUserRoles" type="hidden" value="/api/v1/users/{username}/roles/" />
                </template>

                <!-- GitLab 特殊配置 -->
                <template v-if="formData.type === 'gitlab'">
                    <ElDivider content-position="left">GitLab 特殊配置</ElDivider>
                    <ElFormItem label="项目ID (可选)">
                        <ElInput v-model="formData.endpoints.projectId" placeholder="用于项目级权限管理" />
                        <span class="text-xs text-gray-500">填写项目ID可管理项目成员权限</span>
                    </ElFormItem>
                    <ElFormItem label="群组ID (可选)">
                        <ElInput v-model="formData.endpoints.groupId" placeholder="用于群组级权限管理" />
                        <span class="text-xs text-gray-500">填写群组ID可管理群组成员权限</span>
                    </ElFormItem>
                    <ElFormItem label="获取用户列表">
                        <ElInput v-model="formData.endpoints.getUsers" placeholder="/api/v4/users" />
                    </ElFormItem>
                    <ElFormItem label="创建用户">
                        <ElInput v-model="formData.endpoints.createUser" placeholder="/api/v4/users" />
                    </ElFormItem>
                </template>

                <ElDivider content-position="left">其他配置</ElDivider>

                <ElFormItem label="描述">
                    <ElInput v-model="formData.description" type="textarea" :rows="3" placeholder="请输入描述" />
                </ElFormItem>
                <ElFormItem label="同步间隔">
                    <ElInputNumber v-model="formData.syncInterval" :min="1" placeholder="同步间隔（分钟）" class="w-full" />
                </ElFormItem>
            </ElForm>
            <template #footer>
                <ElButton @click="drawerVisible = false">取消</ElButton>
                <ElButton type="primary" @click="handleSubmit">确定</ElButton>
            </template>
        </ElDrawer>

        <!-- 角色列表对话框 -->
        <ElDialog v-model="rolesDialogVisible" :title="`${currentAppName} - 角色列表`" width="800px">
            <ElTable :data="applicationRoles" :border="true">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="roleCode" label="角色代码" align="center" min-width="120" />
                <ElTableColumn prop="roleName" label="角色名称" align="center" min-width="120" />
                <ElTableColumn prop="roleType" label="角色类型" align="center" min-width="100">
                    <template #default="{ row }">
                        <ElTag>{{ row.roleType || 'global' }}</ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="description" label="描述" align="center" min-width="150" />
                <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
            </ElTable>
            <template #footer>
                <ElButton type="primary" @click="rolesDialogVisible = false">关闭</ElButton>
            </template>
        </ElDialog>

        <!-- 用户列表对话框 -->
        <ElDialog v-model="usersDialogVisible" :title="`${currentAppName} - 用户列表`" width="800px">
            <ElTable :data="applicationUsers" :border="true">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="username" label="用户名" align="center" min-width="120" />
                <ElTableColumn prop="displayName" label="显示名称" align="center" min-width="120" />
                <ElTableColumn prop="email" label="邮箱" align="center" min-width="150" />
                <ElTableColumn prop="status" label="状态" align="center" width="80">
                    <template #default="{ row }">
                        <ElTag :type="row.status === 'active' ? 'success' : 'info'">
                            {{ row.status || 'active' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
            </ElTable>
            <template #footer>
                <ElButton type="primary" @click="usersDialogVisible = false">关闭</ElButton>
            </template>
        </ElDialog>

        <!-- 用户组列表对话框 -->
        <ElDialog v-model="groupsDialogVisible" :title="`${currentAppName} - 用户组列表`" width="800px">
            <ElTable :data="applicationGroups" :border="true">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="groupCode" label="用户组代码" align="center" min-width="120" />
                <ElTableColumn prop="groupName" label="用户组名称" align="center" min-width="120" />
                <ElTableColumn prop="description" label="描述" align="center" min-width="200" />
                <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
            </ElTable>
            <template #footer>
                <ElButton type="primary" @click="groupsDialogVisible = false">关闭</ElButton>
            </template>
        </ElDialog>

        <!-- 授权规则列表对话框 -->
        <ElDialog v-model="rulesDialogVisible" :title="`${currentAppName} - 授权规则列表`" width="1200px">
            <ElAlert title="授权规则说明" type="info" :closable="false" style="margin-bottom: 16px">
                <p>授权规则定义了用户/用户组对资产的访问权限，包括：</p>
                <ul style="margin: 8px 0; padding-left: 20px">
                    <li>
                        <strong>主体</strong>
                        ：谁可以访问（用户或用户组）
                    </li>
                    <li>
                        <strong>对象</strong>
                        ：可以访问什么（具体资产或全部资产）
                    </li>
                    <li>
                        <strong>权限</strong>
                        ：可以执行的操作（连接、上传、下载、命令等）
                    </li>
                </ul>
            </ElAlert>

            <ElTable :data="authorizationRules" :border="true" style="max-height: 500px; overflow-y: auto">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="ruleName" label="规则名称" align="center" min-width="150" />
                <ElTableColumn prop="subjectType" label="主体类型" align="center" width="100">
                    <template #default="{ row }">
                        <ElTag :type="row.subjectType === 'user' ? 'success' : 'warning'" size="small">
                            {{ row.subjectType === 'user' ? '用户' : '用户组' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="subjectName" label="主体名称" align="center" min-width="120" />
                <ElTableColumn prop="objectType" label="对象类型" align="center" width="100">
                    <template #default="{ row }">
                        <ElTag :type="row.objectType === 'system' ? 'danger' : 'primary'" size="small">
                            {{ row.objectType === 'system' ? '全部资产' : '具体资产' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="objectName" label="对象名称" align="center" min-width="150" />
                <ElTableColumn prop="actions" label="操作权限" align="center" min-width="200">
                    <template #default="{ row }">
                        <div v-if="row.actions">
                            <ElTag v-for="action in JSON.parse(row.actions)" :key="action" size="small" style="margin: 2px" :type="getActionType(action)">
                                {{ getActionLabel(action) }}
                            </ElTag>
                        </div>
                        <span v-else style="color: #909399">无权限数据</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="priority" label="优先级" align="center" width="80" />
                <ElTableColumn prop="isEnabled" label="状态" align="center" width="80">
                    <template #default="{ row }">
                        <ElTag :type="row.isEnabled ? 'success' : 'info'" size="small">
                            {{ row.isEnabled ? '启用' : '禁用' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="isExpired" label="是否过期" align="center" width="80">
                    <template #default="{ row }">
                        <ElTag :type="row.isExpired ? 'danger' : 'success'" size="small">
                            {{ row.isExpired ? '已过期' : '有效' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
            </ElTable>
            <template #footer>
                <ElButton type="primary" @click="rulesDialogVisible = false">关闭</ElButton>
            </template>
        </ElDialog>
    </div>
</template>
