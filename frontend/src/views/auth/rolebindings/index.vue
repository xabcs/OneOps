<script setup lang="tsx">
import { onMounted, ref } from 'vue';
import { Delete, Plus } from '@element-plus/icons-vue';
import {
  createGroupBinding,
  deleteGroupBinding,
  fetchApplicationRoles,
  fetchApplications,
  fetchGroupBindings
} from '@/service/api/application-permission';
import { fetchAllAuthGroups } from '@/service/api';

defineOptions({ name: 'AuthCenterGroupBindings' });

const loading = ref(false);
const tableData = ref<any[]>([]);
const groups = ref<any[]>([]);
const selectedGroupId = ref<number | null>(null);
const drawerVisible = ref(false);

const formData = ref({
  appId: null as number | null,
  applicationRoleId: null as number | null
});

const applications = ref<Api.ApplicationPermission.Application[]>([]);
const applicationRoles = ref<Api.ApplicationPermission.ApplicationRole[]>([]);

async function getRoles() {
  const { data, error } = await fetchAllAuthGroups();
  if (!error && data) {
    groups.value = Array.isArray(data) ? data : (data as any).records || [];
  }
}

async function getApplications() {
  const { data, error } = await fetchApplications({ current: 1, size: 1000 });
  if (!error && data) {
    applications.value = data.records || [];
  }
}

async function getData() {
  if (!selectedGroupId.value) return;

  loading.value = true;
  try {
    const { data, error } = await fetchGroupBindings(selectedGroupId.value);
    if (!error && data) {
      tableData.value = data || [];
    }
  } finally {
    loading.value = false;
  }
}

function handleGroupChange() {
  getData();
}

async function handleAdd() {
  if (!selectedGroupId.value) {
    ElMessage.warning('请先选择用户组');
    return;
  }

  formData.value = {
    appId: null,
    applicationRoleId: null
  };
  drawerVisible.value = true;
}

async function handleAppChange(appId: number) {
  if (appId) {
    const { data, error } = await fetchApplicationRoles(appId);
    if (!error && data) {
      applicationRoles.value = data;
    }
  }
}

async function handleSubmit() {
  if (!selectedGroupId.value || !formData.value.applicationRoleId) {
    ElMessage.warning('请填写完整信息');
    return;
  }

  const { error } = await createGroupBinding({
    groupId: selectedGroupId.value,
    appId: formData.value.appId!,
    applicationRoleId: formData.value.applicationRoleId
  } as any);

  if (!error) {
    ElMessage.success('添加映射成功');
    drawerVisible.value = false;
    getData();
  }
}

async function handleDelete(id: number) {
  const { error } = await deleteGroupBinding(id);
  if (!error) {
    ElMessage.success('删除映射成功');
    getData();
  }
}

onMounted(() => {
  getRoles();
  getApplications();
});
</script>

<template>
  <div class="min-h-500px flex-col gap-4">
    <!-- 用户组选择 -->
    <ElCard shadow="never">
      <ElSpace>
        <ElSelect v-model="selectedGroupId" placeholder="请选择用户组" class="w-300px" @change="handleGroupChange">
          <ElOption v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
        </ElSelect>
        <ElButton type="primary" :icon="Plus" :disabled="!selectedGroupId" @click="handleAdd">添加映射</ElButton>
      </ElSpace>
    </ElCard>

    <!-- 表格 -->
    <ElCard shadow="never" class="flex-1">
      <template #header>
        <span class="text-lg font-medium">权限映射列表</span>
      </template>

      <ElTable v-loading="loading" :data="tableData" :border="false">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn label="应用名称" align="center" min-width="120">
          <template #default="{ row }">
            {{ row.appIDField?.name || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="外部角色名称" align="center" min-width="150">
          <template #default="{ row }">
            {{ row.applicationRole?.roleName || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="外部角色代码" align="center" min-width="150">
          <template #default="{ row }">
            {{ row.applicationRole?.roleCode || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="角色类型" align="center" min-width="100">
          <template #default="{ row }">
            <ElTag>{{ row.applicationRole?.roleType || 'global' }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="createdAt" label="映射时间" align="center" min-width="160" />
        <ElTableColumn label="操作" align="center" width="100">
          <template #default="{ row }">
            <ElButton size="small" type="danger" :icon="Delete" @click="handleDelete(row.id)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <!-- 添加映射抽屉 -->
    <ElDrawer v-model="drawerVisible" title="添加权限映射" :width="400">
      <ElForm :model="formData" label-width="80px">
        <ElFormItem label="应用">
          <ElSelect v-model="formData.appId" placeholder="请选择应用" class="w-full" @change="handleAppChange">
            <ElOption v-for="app in applications" :key="app.id" :label="app.name" :value="app.id" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="外部角色">
          <ElSelect v-model="formData.applicationRoleId" placeholder="请选择外部角色" class="w-full">
            <ElOption v-for="role in applicationRoles" :key="role.id" :label="role.roleName" :value="role.id" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="drawerVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDrawer>
  </div>
</template>
