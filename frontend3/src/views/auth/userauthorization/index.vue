<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Delete, Plus } from '@element-plus/icons-vue';
  import { deleteUserGroup, fetchUserGroups } from '@/service/api/application-permission';
  import { fetchAllAuthGroups, fetchAllUsers } from '@/service/api';
  import AuthorizationOperateDrawer from './modules/authorization-operate-drawer.vue';

  defineOptions({ name: 'AuthCenterUserGroupAssignment' });

  // 这个视图基于用户选择，非分页，保留手动 loading/tableData
  const loading = ref(false);
  const tableData = ref<Api.ApplicationPermission.AuthUserGroup[]>([]);
  const users = ref<Api.ApplicationPermission.AuthUser[]>([]);
  const groups = ref<Api.ApplicationPermission.AuthGroup[]>([]);
  const selectedUserId = ref<number | null>(null);

  // 分配用户组抽屉
  const drawerVisible = ref(false);

  // 授权结果对话框
  const resultDialogVisible = ref(false);
  const authorizationResults = ref<Api.ApplicationPermission.AssignmentResultItem[]>([]);

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
    drawerVisible.value = true;
  }

  function handleShowResults(results: Api.ApplicationPermission.AssignmentResultItem[]) {
    authorizationResults.value = results;
    resultDialogVisible.value = true;
  }

  async function handleDelete(groupId: number) {
    if (!selectedUserId.value) return;

    const { error } = await deleteUserGroup(selectedUserId.value, groupId);
    if (!error) {
      ElMessage.success('删除用户组成员成功');
      getData();
    }
  }

  onMounted(() => {
    getUsers();
    getGroups();
  });
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 用户选择 -->
    <ElCard class="card-wrapper">
      <ElSpace>
        <ElSelect v-model="selectedUserId" placeholder="请选择用户" class="w-300px" @change="handleUserChange">
          <ElOption
            v-for="user in users"
            :key="user.id"
            :label="`${user.username} (${user.nickname || '-'})`"
            :value="user.id"
          />
        </ElSelect>
        <PermissionButton
          code="auth.user.update"
          type="primary"
          :icon="Plus"
          :disabled="!selectedUserId"
          @click="handleAssignGroup"
        >
          分配用户组
        </PermissionButton>
      </ElSpace>
    </ElCard>

    <!-- 表格 -->
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <span class="text-lg font-medium">用户所属用户组列表</span>
      </template>

      <div class="h-[calc(100%-52px)]">
        <ElTable v-loading="loading" height="100%" :data="tableData" :border="false" row-key="groupId">
          <ElTableColumn type="index" label="序号" width="60" align="center" />
          <ElTableColumn prop="groupName" label="用户组名称" align="center" min-width="150" />
          <ElTableColumn prop="groupCode" label="用户组代码" align="center" min-width="150" />
          <ElTableColumn prop="grantedBy" label="授权人" align="center" min-width="120" />
          <ElTableColumn prop="grantedAt" label="授权时间" align="center" min-width="160" />
          <ElTableColumn label="操作" align="center" width="100">
            <template #default="{ row }">
              <PermissionButton
                code="auth.user.update"
                size="small"
                type="danger"
                :icon="Delete"
                @click="handleDelete(row.groupId)"
              >
                删除
              </PermissionButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>

      <!-- 分配用户组抽屉（P0a + P1a + P2a）-->
      <AuthorizationOperateDrawer
        v-model:visible="drawerVisible"
        :user-id="selectedUserId"
        :groups="groups"
        @submitted="getData"
        @show-results="handleShowResults"
      />
    </ElCard>

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
            <ElTag :type="row.success ? 'success' : 'danger'">{{ row.success ? '成功' : '失败' }}</ElTag>
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

<style scoped lang="scss">
  .text-danger {
    color: var(--el-color-danger);
  }

  .text-success {
    color: var(--el-color-success);
  }
</style>
