<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Delete, Plus, View } from '@element-plus/icons-vue';
  import {
    createGroupBinding,
    deleteGroupBinding,
    fetchApplicationAuthorizationRules,
    fetchApplicationRoles,
    fetchApplications,
    fetchGroupBindingExecutions,
    fetchGroupBindings
  } from '@/service/api/application-permission';
  import { fetchAllAuthGroups } from '@/service/api';

  defineOptions({ name: 'AuthCenterGroupBindings' });

  const loading = ref(false);
  const tableData = ref<Api.ApplicationPermission.GroupBinding[]>([]);
  const groups = ref<Api.ApplicationPermission.AuthGroup[]>([]);
  const selectedGroupId = ref<number | null>(null);
  const drawerVisible = ref(false);
  const executionDetailVisible = ref(false);
  const executionDetails = ref<Api.ApplicationPermission.GroupBindingExecution[]>([]);

  const formData = ref({
    appId: null as number | null,
    applicationRoleId: null as number | null,
    authorizationRuleId: null as number | null
  });

  const applications = ref<Api.ApplicationPermission.Application[]>([]);
  const applicationRoles = ref<Api.ApplicationPermission.ApplicationRole[]>([]);
  const authorizationRules = ref<Api.ApplicationPermission.AuthorizationRule[]>([]);
  const selectedApplication = ref<Api.ApplicationPermission.Application | null>(null);

  // 权限分配结果
  const assignmentResult = ref<unknown>(null);
  const showAssignmentResult = ref(false);

  async function getRoles() {
    const { data, error } = await fetchAllAuthGroups();
    if (!error && data) {
      groups.value = Array.isArray(data) ? data : (data as { list?: Api.ApplicationPermission.AuthGroup[] }).list || [];
    }
  }

  async function getApplications() {
    const { data, error } = await fetchApplications({ page: 1, pageSize: 1000 });
    if (!error && data) {
      applications.value = data.list || [];
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
      applicationRoleId: null,
      authorizationRuleId: null
    };
    selectedApplication.value = null;
    applicationRoles.value = [];
    authorizationRules.value = [];
    drawerVisible.value = true;
  }

  async function handleAppChange(appId: number) {
    if (appId) {
      // 查找选中的应用
      const app = applications.value.find(a => a.id === appId);
      selectedApplication.value = app || null;

      if (app?.type === 'jumpserver') {
        // Jumpserver 应用：加载授权规则
        const { data, error } = await fetchApplicationAuthorizationRules(appId);
        if (!error && data) {
          authorizationRules.value = data;
        }
      } else {
        // 其他应用：加载角色
        const { data, error } = await fetchApplicationRoles(appId);
        if (!error && data) {
          applicationRoles.value = data;
        }
      }
    } else {
      selectedApplication.value = null;
      applicationRoles.value = [];
      authorizationRules.value = [];
    }
  }

  async function handleSubmit() {
    if (!selectedGroupId.value || !formData.value.appId) {
      ElMessage.warning('请填写完整信息');
      return;
    }

    const app = applications.value.find(a => a.id === formData.value.appId);

    // 根据应用类型检查必填项
    if (app?.type === 'jumpserver') {
      if (!formData.value.authorizationRuleId) {
        ElMessage.warning('请选择授权规则');
        return;
      }
    } else if (!formData.value.applicationRoleId) {
      ElMessage.warning('请选择角色');
      return;
    }

    loading.value = true;
    try {
      // 对于 Jumpserver，使用 authorizationRuleId 作为 applicationRoleId
      const bindingData = {
        groupId: selectedGroupId.value,
        appId: formData.value.appId!,
        applicationRoleId:
          app?.type === 'jumpserver' ? formData.value.authorizationRuleId! : formData.value.applicationRoleId!
      };

      const { data, error } = await createGroupBinding(bindingData);

      if (!error) {
        drawerVisible.value = false;

        // 显示权限分配结果
        if (data && data.result) {
          assignmentResult.value = data.result;
          showAssignmentResult.value = true;
          showDetailedResult();
        } else {
          ElMessage.success('添加映射成功');
        }

        getData();
      }
    } finally {
      loading.value = false;
    }
  }

  async function handleDelete(id: number) {
    const { error } = await deleteGroupBinding(id);
    if (!error) {
      ElMessage.success('删除映射成功');
      getData();
    }
  }

  async function handleViewExecutionDetail(bindingId: number) {
    const { data, error } = await fetchGroupBindingExecutions(bindingId);
    if (!error && data) {
      executionDetails.value = data || [];
      executionDetailVisible.value = true;
    }
  }

  function showDetailedResult() {
    if (!assignmentResult.value) return;

    const result = assignmentResult.value;
    const messages = [];

    if (result.successCount > 0) {
      messages.push(`成功分配权限: ${result.successCount} 个成员`);
    }

    if (result.createdIdentities > 0) {
      messages.push(`创建外部账号: ${result.createdIdentities} 个`);
    }

    if (result.failedCount > 0) {
      messages.push(`失败: ${result.failedCount} 个成员`);
    }

    if (result.pendingMembers && result.pendingMembers.length > 0) {
      messages.push(`待处理: ${result.pendingMembers.length} 个成员`);
    }

    ElMessageBox.alert(
      `
    <div style="max-height: 400px; overflow-y: auto;">
      <h4 style="margin-bottom: 10px;">权限分配详细结果</h4>
      <div style="margin-bottom: 15px;">
        ${messages.map(m => `<p style="margin: 5px 0;">${m}</p>`).join('')}
      </div>

      ${
        result.createdIdentities && result.createdIdentities.length > 0
          ? `
        <div style="margin-bottom: 15px;">
          <h5 style="margin-bottom: 5px; color: #67C23A;">✓ 已创建的外部账号</h5>
          <ul style="margin: 0; padding-left: 20px;">
            ${result.createdIdentities
              .map(
                (item: { username: string; appName: string; status: string }) => `
              <li style="margin: 3px 0;">${item.username} - ${item.appName} - ${item.status}</li>
            `
              )
              .join('')}
          </ul>
        </div>
      `
          : ''
      }

      ${
        result.pendingMembers && result.pendingMembers.length > 0
          ? `
        <div style="margin-bottom: 15px;">
          <h5 style="margin-bottom: 5px; color: #E6A23C;">⚠ 待处理成员</h5>
          <ul style="margin: 0; padding-left: 20px;">
            ${result.pendingMembers
              .map(
                (item: { username: string; reason: string }) => `
              <li style="margin: 3px 0;">${item.username} - ${item.reason}</li>
            `
              )
              .join('')}
          </ul>
          <p style="margin-top: 10px; color: #909399; font-size: 12px;">
            建议：可以为这些成员手动创建外部账号后重新分配
          </p>
        </div>
      `
          : ''
      }

      ${
        result.failedMembers && result.failedMembers.length > 0
          ? `
        <div>
          <h5 style="margin-bottom: 5px; color: #F56C6C;">✗ 失败的成员</h5>
          <ul style="margin: 0; padding-left: 20px;">
            ${result.failedMembers
              .map(
                (item: { username: string; error: string }) => `
              <li style="margin: 3px 0;">${item.username} - ${item.error}</li>
            `
              )
              .join('')}
          </ul>
        </div>
      `
          : ''
      }
    </div>
    `,
      '权限分配结果',
      {
        dangerouslyUseHTMLString: true,
        confirmButtonText: '确定'
      }
    );
  }

  function getStatusTag(status: string) {
    const statusMap: Record<string, { type: '' | 'success' | 'warning' | 'info' | 'danger'; label: string }> = {
      success: { type: 'success', label: '成功' },
      pending: { type: 'warning', label: '处理中' },
      partial: { type: 'info', label: '部分成功' },
      failed: { type: 'danger', label: '失败' }
    };
    return statusMap[status] || { type: '', label: status };
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
        <ElTableColumn label="操作" align="center" width="180">
          <template #default="{ row }">
            <ElButton size="small" type="primary" :icon="View" @click="handleViewExecutionDetail(row.id)">
              执行详情
            </ElButton>
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
            <ElOption v-for="app in applications" :key="app.id" :label="`${app.name} (${app.type})`" :value="app.id" />
          </ElSelect>
        </ElFormItem>

        <!-- Jumpserver 应用：选择授权规则 -->
        <ElFormItem v-if="selectedApplication?.type === 'jumpserver'" label="授权规则">
          <ElSelect v-model="formData.authorizationRuleId" placeholder="请选择授权规则" class="w-full">
            <ElOption
              v-for="rule in authorizationRules"
              :key="rule.id"
              :label="`${rule.ruleName} - ${rule.subjectType}`"
              :value="rule.id"
            />
          </ElSelect>
        </ElFormItem>

        <!-- 其他应用：选择角色 -->
        <ElFormItem v-else label="外部角色">
          <ElSelect v-model="formData.applicationRoleId" placeholder="请选择外部角色" class="w-full">
            <ElOption v-for="role in applicationRoles" :key="role.id" :label="role.roleName" :value="role.id" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="drawerVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="loading" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDrawer>

    <!-- 执行详情抽屉 -->
    <ElDrawer v-model="executionDetailVisible" title="权限分配执行详情" :width="600">
      <ElTable :data="executionDetails" :border="false">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn label="用户名" align="center" min-width="120">
          <template #default="{ row }">
            {{ row.authUser?.username || row.externalUsername || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="外部用户名" align="center" min-width="120">
          <template #default="{ row }">
            {{ row.externalUsername || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作类型" align="center" width="100">
          <template #default="{ row }">
            <ElTag>{{ row.actionType }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" align="center" width="80">
          <template #default="{ row }">
            <ElTag :type="getStatusTag(row.status).type">
              {{ getStatusTag(row.status).label }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="详细信息" align="center" min-width="200">
          <template #default="{ row }">
            {{ row.message || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn prop="createdAt" label="执行时间" align="center" min-width="160" />
      </ElTable>
    </ElDrawer>
  </div>
</template>

<style scoped></style>
