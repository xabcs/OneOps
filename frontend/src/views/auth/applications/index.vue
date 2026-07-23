<script setup lang="tsx">
import { onMounted, ref } from 'vue';
import { Delete, Edit, Plus, Refresh, Search, View } from '@element-plus/icons-vue';
import {
  createApplication,
  deleteApplication,
  fetchApplications,
  fetchApplicationRoles,
  fetchApplicationUsers,
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
  current: 1,
  size: 10
});

const drawerVisible = ref(false);
const isEdit = ref(false);

// 角色和用户查看对话框
const rolesDialogVisible = ref(false);
const usersDialogVisible = ref(false);
const applicationRoles = ref<Api.ApplicationPermission.ApplicationRole[]>([]);
const applicationUsers = ref<Api.ApplicationPermission.ApplicationUser[]>([]);
const currentAppName = ref('');

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

const appTypeOptions = [
  { label: 'Jenkins', value: 'jenkins' },
  { label: 'Nacos', value: 'nacos' },
  { label: 'XXL-Job', value: 'xxl-job' },
  { label: 'RocketMQ', value: 'rocketmq' },
  { label: '其他', value: 'other' }
];

const authTypeOptions = [
  { label: 'API Token', value: 'token' },
  { label: 'Basic Auth', value: 'basic' },
  { label: 'OAuth 2.0', value: 'oauth2' }
];

async function getData() {
  loading.value = true;
  try {
    const { data, error } = await fetchApplications({
      current: pagination.value.current,
      size: pagination.value.size,
      name: searchName.value
    });
    if (!error && data) {
      tableData.value = data.records || [];
      total.value = data.total || 0;
    }
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.value.current = 1;
  getData();
}

function handleReset() {
  searchName.value = '';
  pagination.value.current = 1;
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
  drawerVisible.value = true;
}

function handleEdit(row: any) {
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

  const { error } = (await api) as any;
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

async function handleSyncRoles(row: any) {
  const { error } = await syncApplicationRoles(row.id);
  if (!error) {
    ElMessage.success(`同步 ${row.name} 角色成功`);
    getData();
    // 同步成功后自动显示角色列表
    handleViewRoles(row);
  }
}

async function handleSyncUsers(row: any) {
  const { error } = await syncApplicationUsers(row.id);
  if (!error) {
    ElMessage.success(`同步 ${row.name} 用户成功`);
    getData();
    // 同步成功后自动显示用户列表
    handleViewUsers(row);
  }
}

async function handleViewRoles(row: any) {
  currentAppName.value = row.name;
  const { data, error } = await fetchApplicationRoles(row.id);
  if (!error && data) {
    applicationRoles.value = data;
    rolesDialogVisible.value = true;
  } else {
    ElMessage.warning('暂无角色数据，请先同步角色');
  }
}

async function handleViewUsers(row: any) {
  currentAppName.value = row.name;
  const { data, error } = await fetchApplicationUsers(row.id);
  if (!error && data) {
    applicationUsers.value = data;
    usersDialogVisible.value = true;
  } else {
    ElMessage.warning('暂无用户数据，请先同步用户');
  }
}

function handlePageChange(page: number) {
  pagination.value.current = page;
  getData();
}

function handleSizeChange(size: number) {
  pagination.value.size = size;
  pagination.value.current = 1;
  getData();
}

onMounted(() => {
  getData();
});
</script>

<template>
  <div class="min-h-500px flex-col gap-4">
    <!-- 搜索栏 -->
    <ElCard shadow="never">
      <ElForm :inline="true">
        <ElFormItem label="应用名称">
          <ElInput
            v-model="searchName"
            placeholder="请输入应用名称"
            clearable
            class="w-200px"
            @keyup.enter="handleSearch"
          />
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
          <span class="text-lg font-medium">外部应用列表</span>
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
        <ElTableColumn label="操作" align="center" width="380" fixed="right">
          <template #default="{ row }">
            <ElSpace>
              <ElButton size="small" type="primary" @click="handleSyncRoles(row)">同步角色</ElButton>
              <ElButton size="small" type="success" @click="handleSyncUsers(row)">同步用户</ElButton>
              <ElButton size="small" @click="handleViewRoles(row)">查看角色</ElButton>
              <ElButton size="small" @click="handleViewUsers(row)">查看用户</ElButton>
              <ElButton size="small" type="warning" @click="handleEdit(row)">编辑</ElButton>
              <ElButton size="small" type="danger" @click="handleDelete(row.id)">删除</ElButton>
            </ElSpace>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="mt-4 flex justify-end">
        <ElPagination
          v-model:current-page="pagination.current"
          v-model:page-size="pagination.size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
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
          <ElSelect v-model="formData.type" placeholder="请选择应用类型" class="w-full">
            <ElOption v-for="item in appTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Base URL" required>
          <ElInput v-model="formData.baseUrl" placeholder="请输入应用访问地址，如：https://jenkins.example.com" />
        </ElFormItem>

        <ElDivider content-position="left">认证配置</ElDivider>

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

        <ElDivider content-position="left">API端点配置</ElDivider>

        <ElFormItem label="创建用户" required>
          <ElInput v-model="formData.endpoints.createUser" placeholder="/api/users" />
        </ElFormItem>
        <ElFormItem label="获取用户列表" required>
          <ElInput v-model="formData.endpoints.listUsers" placeholder="/api/users" />
        </ElFormItem>
        <ElFormItem label="获取单个用户" required>
          <ElInput v-model="formData.endpoints.getUser" placeholder="/api/users/{username}" />
        </ElFormItem>
        <ElFormItem label="获取角色列表" required>
          <ElInput v-model="formData.endpoints.getRoles" placeholder="/api/roles" />
        </ElFormItem>
        <ElFormItem label="分配角色" required>
          <ElInput v-model="formData.endpoints.assignRole" placeholder="/api/users/{username}/roles" />
        </ElFormItem>
        <ElFormItem label="撤销角色" required>
          <ElInput v-model="formData.endpoints.revokeRole" placeholder="/api/users/{username}/roles/{roleCode}" />
        </ElFormItem>
        <ElFormItem label="获取用户角色">
          <ElInput v-model="formData.endpoints.getUserRoles" placeholder="/api/users/{username}/roles" />
        </ElFormItem>

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
  </div>
</template>
