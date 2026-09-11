<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Delete, Plus, View } from '@element-plus/icons-vue';
  import {
    deleteGroupBinding,
    fetchAllAuthGroups,
    fetchGroupBindingExecutions,
    fetchGroupBindings
  } from '@/service/api/application-permission';
  import { createTagMap } from '@/utils/common';
  import BindingOperateDrawer from './modules/binding-operate-drawer.vue';

  defineOptions({ name: 'AuthCenterGroupBindings' });

  // 这个视图基于 group 选择，非分页，保留手动 loading/tableData
  const loading = ref(false);
  const tableData = ref<Api.ApplicationPermission.GroupBinding[]>([]);
  const groups = ref<Api.ApplicationPermission.AuthGroup[]>([]);
  const selectedGroupId = ref<number | null>(null);

  // 添加映射抽屉
  const drawerVisible = ref(false);

  // 执行详情抽屉
  const executionDetailVisible = ref(false);
  const executionDetails = ref<Api.ApplicationPermission.GroupBindingExecution[]>([]);

  async function getRoles() {
    const { data, error } = await fetchAllAuthGroups();
    if (!error && data) {
      groups.value = Array.isArray(data) ? data : (data as { list?: Api.ApplicationPermission.AuthGroup[] }).list || [];
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

  // 布局重置按钮：清空用户组筛选
  function handleReset() {
    selectedGroupId.value = null;
  }

  function handleAdd() {
    if (!selectedGroupId.value) {
      ElMessage.warning('请先选择用户组');
      return;
    }
    drawerVisible.value = true;
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

  /** 映射执行状态 → ElTag 标签映射 */
  const getStatusTag = createTagMap({
    success: { text: '成功', type: 'success' },
    pending: { text: '处理中', type: 'warning' },
    partial: { text: '部分成功', type: 'info' },
    failed: { text: '失败', type: 'danger' }
  });

  onMounted(() => {
    getRoles();
  });
</script>

<template>
  <ListPageLayout title="权限映射列表" @search="getData" @reset="handleReset">
    <!-- 用户组选择 -->
    <template #search>
      <ElSelect v-model="selectedGroupId" placeholder="请选择用户组" class="w-300px" @change="handleGroupChange">
        <ElOption v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
      </ElSelect>
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <PermissionButton
        code="auth.group.create"
        type="primary"
        :icon="Plus"
        :disabled="!selectedGroupId"
        @click="handleAdd"
      >
        添加映射
      </PermissionButton>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" height="100%" :data="tableData" :border="false" row-key="id">
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
          <ElTableColumn label="操作" align="center" width="180" class-name="msre-table-actions">
            <template #default="{ row }">
              <ElButton link type="primary" size="small" :icon="View" @click="handleViewExecutionDetail(row.id)">
                执行详情
              </ElButton>
              <PermissionButton
                link
                type="danger"
                size="small"
                code="auth.group.delete"
                :icon="Delete"
                @click="handleDelete(row.id)"
              >
                删除
              </PermissionButton>
            </template>
          </ElTableColumn>
    </ElTable>

    <!-- 添加映射抽屉（P0a + P1a + P2a）-->
    <BindingOperateDrawer v-model:visible="drawerVisible" :group-id="selectedGroupId" @submitted="getData" />

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
              {{ getStatusTag(row.status).text }}
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
  </ListPageLayout>
</template>
