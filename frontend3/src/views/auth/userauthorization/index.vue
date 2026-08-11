<script setup lang="tsx">
import { onMounted, ref } from 'vue';
import { Delete, Plus } from '@element-plus/icons-vue';
import { assignUserToGroup, deleteUserGroup, fetchUserGroups } from '@/service/api/application-permission';
import { fetchAllAuthGroups, fetchAllUsers } from '@/service/api';

defineOptions({ name: 'AuthCenterUserGroupAssignment' });

const loading = ref(false);
const tableData = ref<Api.ApplicationPermission.AuthUserGroup[]>([]);
const users = ref<Api.ApplicationPermission.AuthUser[]>([]);
const groups = ref<Api.ApplicationPermission.AuthGroup[]>([]);
const selectedUserId = ref<number | null>(null);
const drawerVisible = ref(false);
const resultDialogVisible = ref(false);
const authorizationResults = ref<unknown[]>([]);

const formData = ref({
  groupId: null as number | null
});

async function getUsers() {
  const { data, error } = await fetchAllUsers();
  if (!error && data) {
    users.value = Array.isArray(data) ? data : (data as { list?: Api.ApplicationPermission.AuthUser[] }).list || [];
  }
}

async function getGroups() {
  const { data, error } = await fetchAllAuthGroups();
  if (!error && data) {
    groups.value = Array.isArray(data) ? data : (data as { list?: Api.ApplicationPermission.AuthGroup[] }).list || [];
  }
}

async function getData() {
  if (!selectedUserId.value) return;

  loading.value = true;
  try {
    const { data, error } = await fetchUserGroups(selectedUserId.value);
    if (!error && data) {
      tableData.value = data || [];
    }
  } finally {
    loading.value = false;
  }
}

function handleUserChange() {
  getData();
}

function handleAssignGroup() {
  if (!selectedUserId.value) {
    ElMessage.warning('请先选择用户');
    return;
  }
  formData.value = {
    groupId: null
  };
  drawerVisible.value = true;
}

async function handleSubmit() {
  if (!selectedUserId.value || !formData.value.groupId) {
    ElMessage.warning('请填写完整信息');
    return;
  }

  const { data, error } = await assignUserToGroup({
    userId: selectedUserId.value,
    groupId: formData.value.groupId
  });

  if (!error && data) {
    drawerVisible.value = false;
    getData();

    // 显示授权结果（不显示密码，因为密码在创建用户时已经显示过）
    if (data.results && data.results.length > 0) {
      // 统计成功/失败数量
      const successCount = data.results.filter((r: { success: boolean }) => r.success).length;
      const failCount = data.results.filter((r: { success: boolean }) => !r.success).length;

      if (failCount === 0) {
        ElMessage.success(`授权成功！已在 ${successCount} 个外部系统中授权`);
      } else {
        authorizationResults.value = data.results;
        resultDialogVisible.value = true;
      }
    } else {
      ElMessage.success('分配用户组成功');
    }
  }
}

async function handleDelete(groupId: number) {
  if (!selectedUserId.value) return;

  const { error } = await deleteUserGroup(selectedUserId.value, groupId);
  if (!error) {
    ElMessage.success('删除用户组成员成功');
    getData();
  }
}

/**
 * 复制文本到剪贴板
 */
function copyToClipboard(text: string) {
  navigator.clipboard
    .writeText(text)
    .then(() => {
      ElMessage.success('已复制到剪贴板');
    })
    .catch(() => {
      ElMessage.error('复制失败，请手动复制');
    });
}

onMounted(() => {
  getUsers();
  getGroups();
});
</script>

<template>
    <div class="min-h-500px flex-col gap-4">
        <!-- 用户选择 -->
        <ElCard shadow="never">
            <ElSpace>
                <ElSelect v-model="selectedUserId" placeholder="请选择用户" class="w-300px" @change="handleUserChange">
                    <ElOption v-for="user in users" :key="user.id" :label="`${user.username} (${user.nickname || '-'})`" :value="user.id" />
                </ElSelect>
                <ElButton type="primary" :icon="Plus" :disabled="!selectedUserId" @click="handleAssignGroup">
                    分配用户组
                </ElButton>
            </ElSpace>
        </ElCard>

        <!-- 表格 -->
        <ElCard shadow="never" class="flex-1">
            <template #header>
                <span class="text-lg font-medium">用户所属用户组列表</span>
            </template>

            <ElTable v-loading="loading" :data="tableData" :border="false">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="groupName" label="用户组名称" align="center" min-width="150" />
                <ElTableColumn prop="groupCode" label="用户组代码" align="center" min-width="150" />
                <ElTableColumn prop="grantedBy" label="授权人" align="center" min-width="120" />
                <ElTableColumn prop="grantedAt" label="授权时间" align="center" min-width="160" />
                <ElTableColumn label="操作" align="center" width="100">
                    <template #default="{ row }">
                        <ElButton size="small" type="danger" :icon="Delete" @click="handleDelete(row.groupId)">删除</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </ElCard>

        <!-- 分配用户组抽屉 -->
        <ElDrawer v-model="drawerVisible" title="分配用户组" :width="400">
            <ElForm :model="formData" label-width="80px">
                <ElFormItem label="用户组">
                    <ElSelect v-model="formData.groupId" placeholder="请选择用户组" class="w-full">
                        <ElOption v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
                    </ElSelect>
                </ElFormItem>
            </ElForm>
            <template #footer>
                <ElButton @click="drawerVisible = false">取消</ElButton>
                <ElButton type="primary" @click="handleSubmit">确定</ElButton>
            </template>
        </ElDrawer>

        <!-- 授权结果对话框 -->
        <ElDialog v-model="resultDialogVisible" title="授权结果" width="600px">
            <ElAlert type="info" :closable="false" class="mb-4">
                <template #title>用户已使用统一初始密码（创建用户时已显示），本次授权在外部系统中使用了该密码</template>
            </ElAlert>

            <ElTable :data="authorizationResults" :border="true">
                <ElTableColumn prop="appName" label="应用名称" align="center" min-width="120" />
                <ElTableColumn prop="username" label="用户名" align="center" min-width="100" />
                <ElTableColumn prop="roleName" label="外部角色" align="center" min-width="120" />
                <ElTableColumn label="状态" align="center" width="100">
                    <template #default="{ row }">
                        <ElTag :type="row.success ? 'success' : 'danger'">
                            {{ row.success ? '成功' : '失败' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="备注" align="center" min-width="150">
                    <template #default="{ row }">
                        <div v-if="row.createError" class="text-danger text-xs">{{ row.createError }}</div>
                        <div v-else-if="row.grantError" class="text-danger text-xs">{{ row.grantError }}</div>
                        <div v-else class="text-xs text-success">已授权</div>
                    </template>
                </ElTableColumn>
            </ElTable>

            <template #footer>
                <ElButton type="primary" @click="resultDialogVisible = false">关闭</ElButton>
            </template>
        </ElDialog>
    </div>
</template>
