<script setup lang="tsx">
import { onMounted, ref } from 'vue';
import { Delete, Edit, Plus, Refresh, Search, View } from '@element-plus/icons-vue';
import {
  assignUserToGroup,
  createAuthUser,
  deleteAuthUser,
  deleteUserGroup,
  fetchAllAuthGroups,
  fetchAuthUsers,
  fetchUserGroups,
  getAuthUserPassword,
  updateAuthUser
} from '@/service/api/application-permission';

defineOptions({ name: 'AuthCenterUsers' });

const loading = ref(false);
const tableData = ref<Api.ApplicationPermission.AuthUser[]>([]);
const total = ref(0);
const searchUsername = ref('');

const pagination = ref({
  current: 1,
  size: 10
});

const drawerVisible = ref(false);
const isEdit = ref(false);
const passwordDialogVisible = ref(false);
const createdUserPassword = ref({
  username: '',
  password: ''
});

const formData = ref<Api.ApplicationPermission.AuthUser>({
  username: '',
  nickname: '',
  email: '',
  phone: '',
  description: '',
  status: 1
});

// 用户组管理相关状态
const groupDialogVisible = ref(false);
const selectedUser = ref<Api.ApplicationPermission.AuthUser | null>(null);
const userGroups = ref<any[]>([]);
const allGroups = ref<any[]>([]);
const groupFormData = ref({
  groupId: null as number | null
});

async function getData() {
  loading.value = true;
  try {
    const { data, error } = await fetchAuthUsers({
      current: pagination.value.current,
      size: pagination.value.size,
      username: searchUsername.value
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
  searchUsername.value = '';
  pagination.value.current = 1;
  getData();
}

// 用户组管理相关函数
async function handleManageGroups(row: Api.ApplicationPermission.AuthUser) {
  selectedUser.value = row;
  groupDialogVisible.value = true;

  // 获取所有用户组
  const { data: allGroupsData } = await fetchAllAuthGroups();
  if (allGroupsData) {
    allGroups.value = Array.isArray(allGroupsData) ? allGroupsData : (allGroupsData as any).records || [];
  }

  // 获取该用户所属的用户组
  const { data: currentUserGroups } = await fetchUserGroups(row.id);
  if (currentUserGroups) {
    userGroups.value = currentUserGroups;
  }
}

async function handleAssignGroup() {
  if (!selectedUser.value || !groupFormData.value.groupId) {
    ElMessage.warning('请选择用户组');
    return;
  }

  const { error } = await assignUserToGroup({
    userId: selectedUser.value.id,
    groupId: groupFormData.value.groupId
  });

  if (!error) {
    ElMessage.success('分配用户组成功');
    // 重新获取用户组列表
    if (selectedUser.value) {
      const { data: updatedUserGroups } = await fetchUserGroups(selectedUser.value.id);
      if (updatedUserGroups) {
        userGroups.value = updatedUserGroups;
      }
    }
  }
}

async function handleRemoveGroup(groupId: number) {
  if (!selectedUser.value) return;

  const { error } = await deleteUserGroup(selectedUser.value.id, groupId);
  if (!error) {
    ElMessage.success('移除用户组成功');
    // 重新获取用户组列表
    if (selectedUser.value) {
      const { data: updatedUserGroups } = await fetchUserGroups(selectedUser.value.id);
      if (updatedUserGroups) {
        userGroups.value = updatedUserGroups;
      }
    }
  }
}

function handleAdd() {
  isEdit.value = false;
  formData.value = {
    username: '',
    nickname: '',
    email: '',
    phone: '',
    description: '',
    status: 1
  };
  drawerVisible.value = true;
}

function handleEdit(row: any) {
  isEdit.value = true;
  formData.value = { ...row };
  drawerVisible.value = true;
}

async function handleSubmit() {
  if (isEdit.value) {
    // 编辑模式
    const { error } = (await updateAuthUser(formData.value.id!, formData.value)) as any;
    if (!error) {
      ElMessage.success('更新成功');
      drawerVisible.value = false;
      getData();
    }
  } else {
    // 新增模式
    const { data, error } = (await createAuthUser(formData.value)) as any;
    if (!error && data) {
      drawerVisible.value = false;
      getData();

      // 显示初始密码
      if (data.password) {
        createdUserPassword.value = {
          username: data.username,
          password: data.password
        };
        passwordDialogVisible.value = true;
      }
    }
  }
}

async function handleDelete(id: number) {
  const { error } = await deleteAuthUser(id);
  if (!error) {
    ElMessage.success('删除成功');
    getData();
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

async function handleViewPassword(row: any) {
  const { data, error } = await getAuthUserPassword(row.id);
  if (!error && data) {
    createdUserPassword.value = {
      username: data.username,
      password: data.password
    };
    passwordDialogVisible.value = true;
  }
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
        <ElFormItem label="用户名">
          <ElInput
            v-model="searchUsername"
            placeholder="请输入用户名"
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
          <span class="text-lg font-medium">授权中心用户列表</span>
          <ElButton type="primary" :icon="Plus" @click="handleAdd">添加用户</ElButton>
        </div>
      </template>

      <ElTable v-loading="loading" :data="tableData" :border="false">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn prop="username" label="用户名" align="center" min-width="120" />
        <ElTableColumn prop="nickname" label="昵称" align="center" min-width="120" />
        <ElTableColumn prop="email" label="邮箱" align="center" min-width="150" />
        <ElTableColumn prop="phone" label="电话" align="center" min-width="120" />
        <ElTableColumn label="用户组" align="center" min-width="200">
          <template #default="{ row }">
            <div class="flex items-center justify-center gap-2">
              <ElTag v-if="row.groups && row.groups.length > 0"
                     v-for="group in row.groups"
                     :key="group.id"
                     type="primary"
                     size="small">
                {{ group.groupName }}
              </ElTag>
              <span v-else class="text-gray-400">未分配</span>
              <ElButton size="small" type="primary" link @click="handleManageGroups(row)">
                {{ row.groups && row.groups.length > 0 ? '管理' : '分配' }}
              </ElButton>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="初始密码" align="center" min-width="140">
          <template #default="{ row }">
            <ElButton size="small" type="primary" link @click="handleViewPassword(row)">查看密码</ElButton>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="status" label="状态" align="center" width="80">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" align="center" width="200" fixed="right">
          <template #default="{ row }">
            <ElSpace>
              <ElButton size="small" type="primary" link @click="handleManageGroups(row)">用户组</ElButton>
              <ElButton size="small" type="warning" link :icon="Edit" @click="handleEdit(row)">编辑</ElButton>
              <ElButton size="small" type="danger" link :icon="Delete" @click="handleDelete(row.id)">删除</ElButton>
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
    <ElDrawer v-model="drawerVisible" :title="isEdit ? '编辑用户' : '添加用户'" :width="500">
      <ElForm :model="formData" label-width="100px">
        <ElFormItem label="用户名" required>
          <ElInput v-model="formData.username" placeholder="请输入用户名" :disabled="isEdit" />
        </ElFormItem>
        <ElFormItem label="昵称">
          <ElInput v-model="formData.nickname" placeholder="请输入昵称" />
        </ElFormItem>
        <ElFormItem label="邮箱">
          <ElInput v-model="formData.email" placeholder="请输入邮箱" />
        </ElFormItem>
        <ElFormItem label="电话">
          <ElInput v-model="formData.phone" placeholder="请输入电话" />
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="formData.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElRadioGroup v-model="formData.status">
            <ElRadio :value="1">启用</ElRadio>
            <ElRadio :value="0">禁用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="drawerVisible = false">取消</ElButton>
        <ElButton type="primary" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDrawer>

    <!-- 初始密码显示对话框 -->
    <ElDialog v-model="passwordDialogVisible" title="用户创建成功" width="500px">
      <ElAlert type="warning" :closable="false" class="mb-4">
        <template #title>
          <strong>重要提示：</strong>
          以下初始密码仅显示一次，请妥善保管或立即通知用户修改密码
        </template>
      </ElAlert>

      <ElDescriptions :column="1" border>
        <ElDescriptionsItem label="用户名">{{ createdUserPassword.username }}</ElDescriptionsItem>
        <ElDescriptionsItem label="初始密码">
          <div class="flex items-center gap-2">
            <code class="text-lg text-primary font-bold">{{ createdUserPassword.password }}</code>
            <ElButton
              size="small"
              type="primary"
              @click="
                () => {
                  navigator.clipboard.writeText(createdUserPassword.password);
                  ElMessage.success('密码已复制到剪贴板');
                }
              "
            >
              复制密码
            </ElButton>
          </div>
        </ElDescriptionsItem>
      </ElDescriptions>

      <div class="mt-4 text-sm text-gray-500">
        <p>• 该密码将用于所有外部系统的初始登录</p>
        <p>• 请通知用户在首次登录后立即修改密码</p>
        <p>• 此密码仅显示一次，关闭后将无法再次查看</p>
      </div>

      <template #footer>
        <ElButton type="primary" @click="passwordDialogVisible = false">我已保存，关闭</ElButton>
      </template>
    </ElDialog>

    <!-- 用户组管理对话框 -->
    <ElDialog v-model="groupDialogVisible" :title="`${selectedUser?.username || ''} 的用户组管理`" width="600px">
      <div class="mb-4">
        <div class="flex items-center gap-2 mb-4">
          <ElText type="primary">分配用户组：</ElText>
          <ElSelect v-model="groupFormData.groupId" placeholder="选择用户组" style="width: 200px">
            <ElOption
              v-for="group in allGroups"
              :key="group.id"
              :label="`${group.name} (${group.code})`"
              :value="group.id"
            />
          </ElSelect>
          <ElButton type="primary" size="small" @click="handleAssignGroup">分配</ElButton>
        </div>
      </div>

      <ElDivider content-position="left">已分配的用户组</ElDivider>

      <ElTable :data="userGroups" :border="false" max-height="400px">
        <ElTableColumn prop="groupName" label="用户组名称" align="center" min-width="150" />
        <ElTableColumn prop="groupCode" label="用户组代码" align="center" min-width="150" />
        <ElTableColumn prop="grantedBy" label="分配人" align="center" min-width="100" />
        <ElTableColumn label="操作" align="center" width="100">
          <template #default="{ row }">
            <ElButton size="small" type="danger" @click="handleRemoveGroup(row.id)">移除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <ElEmpty v-if="userGroups.length === 0" description="暂无用户组" />

      <template #footer>
        <ElButton @click="groupDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
