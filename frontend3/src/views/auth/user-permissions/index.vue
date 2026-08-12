<script setup lang="tsx">
  import { computed, onMounted, ref } from 'vue';
  import {
    fetchApplications,
    fetchUserEffectivePermissions,
    fetchUserEffectivePermissionsMatrix
  } from '@/service/api/application-permission';
  import AppPermissionMatrix from './modules/AppPermissionMatrix.vue';

  defineOptions({ name: 'AuthUserPermissions' });

  const loading = ref(false);
  const matrixLoading = ref(false);
  const tableData = ref<Api.ApplicationPermission.UserEffectivePermission[]>([]);
  const applications = ref<Api.ApplicationPermission.Application[]>([]);
  const matrixData = ref<unknown>(null);
  const viewMode = ref<'list' | 'matrix'>('matrix');

  const selectedAppId = ref<number | null>(null);
  const selectedRoleType = ref<string>('all'); // 支持所有应用类型：all, global, project, user, group

  // Jenkins/GitLab 的角色类型
  const roleTypes = [
    { value: 'all', label: '全部' },
    { value: 'global', label: 'Global 角色' },
    { value: 'project', label: 'Project 角色' }
  ];

  const searchParams = ref({
    username: '',
    appId: null as number | null
  });

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  });

  const selectedApp = computed(() => {
    if (!selectedAppId.value) return null;
    return applications.value.find(app => app.id === selectedAppId.value);
  });

  // 数据适配函数：将不同应用的数据结构统一化
  interface AdaptedMatrixData {
    columns: Array<Record<string, unknown>>;
    users: Array<Record<string, unknown>>;
    matrix: Record<string, Record<string, boolean>>;
    permissions_detail: Record<string, unknown>;
    message?: string;
  }

  const adaptedMatrixData = computed<AdaptedMatrixData | null>(() => {
    if (!matrixData.value) return null;

    const appType = selectedApp.value?.type || 'jenkins';
    const data = matrixData.value as Record<string, unknown>;

    // 根据应用类型适配数据结构
    switch (appType) {
      case 'jenkins':
      case 'gitlab':
        // Jenkins/GitLab: roles -> columns
        // 添加 name 字段以兼容 PermissionMatrix 组件
        const adaptedRoles = ((data.roles as Record<string, unknown>[]) || []).map((role: Record<string, unknown>) => ({
          ...role,
          name: role.roleName || role.name // 兼容不同字段名
        }));

        return {
          columns: adaptedRoles,
          users: (data.users as Array<Record<string, unknown>>) || [],
          matrix: (data.matrix as Record<string, Record<string, boolean>>) || {},
          permissions_detail: (data.permissions_detail as Record<string, unknown>) || {},
          message: (data.message as string) || ''
        };

      case 'jumpserver':
        // Jumpserver: rules -> columns
        const adaptedRules = ((data.rules as Record<string, unknown>[]) || []).map((rule: Record<string, unknown>) => ({
          ...rule, // 先展开 rule，这样后续的字段会覆盖它
          id: rule.rule_id, // 覆盖 id 为 rule_id (UUID)，用于矩阵匹配
          name: rule.rule_name,
          type: rule.subject_type,
          ruleType: rule.subject_type // 保持兼容性
        }));

        // 确保 users 不是 null
        const adaptedUsers = (data.users as Array<Record<string, unknown>>) || [];

        return {
          columns: adaptedRules,
          users: adaptedUsers,
          matrix: (data.matrix as Record<string, Record<string, boolean>>) || {},
          permissions_detail: (data.permissions_detail as Record<string, unknown>) || {},
          message: (data.message as string) || ''
        };

      default:
        // 默认尝试直接使用
        return data as unknown as AdaptedMatrixData;
    }
  });

  async function getApplications() {
    const { data, error } = await fetchApplications({ page: 1, pageSize: 1000 });
    if (!error && data) {
      applications.value = data.list || [];
      // 默认选择第一个应用
      if (applications.value.length > 0 && !selectedAppId.value) {
        selectedAppId.value = applications.value[0].id;
        handleAppChange();
      }
    }
  }

  async function getData() {
    loading.value = true;
    try {
      const { data, error } = await fetchUserEffectivePermissions({
        page: pagination.value.page,
        pageSize: pagination.value.pageSize,
        ...searchParams.value
      });
      if (!error && data) {
        tableData.value = data.list || [];
        pagination.value.total = data.total || 0;
      }
    } finally {
      loading.value = false;
    }
  }

  async function getMatrixData() {
    if (!selectedAppId.value) return;

    matrixLoading.value = true;
    try {
      const { data, error } = await fetchUserEffectivePermissionsMatrix(selectedAppId.value);
      if (!error && data) {
        matrixData.value = data;
      }
    } finally {
      matrixLoading.value = false;
    }
  }

  function handleAppChange() {
    selectedRoleType.value = 'all';
    if (viewMode.value === 'matrix') {
      getMatrixData();
    } else {
      searchParams.value.appId = selectedAppId.value;
      getData();
    }
  }

  function handleRoleTypeChange() {
    // 角色类型切换时重新加载矩阵数据
    getMatrixData();
  }

  function handleSearch() {
    pagination.value.page = 1;
    if (viewMode.value === 'matrix') {
      getMatrixData();
    } else {
      getData();
    }
  }

  function handleReset() {
    searchParams.value = {
      username: '',
      appId: selectedAppId.value
    };
    handleSearch();
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

  function handleViewModeChange() {
    if (viewMode.value === 'matrix') {
      getMatrixData();
    } else {
      searchParams.value.appId = selectedAppId.value;
      getData();
    }
  }

  function getStatusTag(status: string) {
    const statusMap: Record<
      string,
      { type: 'primary' | 'success' | 'warning' | 'info' | 'danger'; label: string }
    > = {
      active: { type: 'success', label: '有效' },
      inactive: { type: 'info', label: '无效' },
      expired: { type: 'danger', label: '已过期' },
      pending: { type: 'warning', label: '待生效' }
    };
    return statusMap[status] || { type: 'primary', label: status };
  }

  onMounted(() => {
    getApplications();
  });
</script>

<template>
  <div class="min-h-500px flex-col gap-4">
    <!-- Tab 切换层 -->
    <div class="space-y-3">
      <!-- 第一层：应用切换 + 视图切换 -->
      <div class="flex items-center justify-between gap-4">
        <div class="permission-tabs-shell flex-1">
          <button
            v-for="app in applications"
            :key="app.id"
            type="button"
            class="permission-tab-btn"
            :class="{ active: selectedAppId === app.id }"
            @click="
              selectedAppId = app.id;
              handleAppChange();
            "
          >
            {{ app.name }}
          </button>
        </div>

        <div class="permission-tabs-shell">
          <button
            type="button"
            class="permission-tab-btn"
            :class="{ active: viewMode === 'matrix' }"
            @click="
              viewMode = 'matrix';
              handleViewModeChange();
            "
          >
            矩阵视图
          </button>
          <button
            type="button"
            class="permission-tab-btn"
            :class="{ active: viewMode === 'list' }"
            @click="
              viewMode = 'list';
              handleViewModeChange();
            "
          >
            列表视图
          </button>
        </div>
      </div>

      <!-- 第二层：类型切换（根据应用类型显示不同选项） -->
      <div v-if="viewMode === 'matrix'" class="permission-tabs-shell">
        <!-- Jenkins/GitLab 的角色类型切换 -->
        <template v-if="selectedApp?.type === 'jenkins' || selectedApp?.type === 'gitlab'">
          <button
            v-for="type in roleTypes"
            :key="type.value"
            type="button"
            class="permission-tab-btn"
            :class="{ active: selectedRoleType === type.value }"
            @click="
              selectedRoleType = type.value;
              handleRoleTypeChange();
            "
          >
            {{ type.label }}
          </button>
        </template>

        <!-- Jumpserver 的规则类型切换 -->
        <template v-else-if="selectedApp?.type === 'jumpserver'">
          <button
            v-for="type in [
              { value: 'all', label: '全部' },
              { value: 'user', label: '用户规则' },
              { value: 'group', label: '用户组规则' }
            ]"
            :key="type.value"
            type="button"
            class="permission-tab-btn"
            :class="{ active: selectedRoleType === type.value }"
            @click="
              selectedRoleType = type.value;
              handleRoleTypeChange();
            "
          >
            {{ type.label }}
          </button>
        </template>
      </div>
    </div>

    <!-- 数据展示 -->
    <ElCard shadow="never" class="flex-1">
      <!-- 矩阵视图 -->
      <div v-if="viewMode === 'matrix'" v-loading="matrixLoading">
        <ElEmpty v-if="!selectedAppId" description="请选择应用" />

        <!-- 使用统一的矩阵组件 -->
        <AppPermissionMatrix
          v-else-if="adaptedMatrixData && adaptedMatrixData.columns && adaptedMatrixData.columns.length > 0"
          :data="adaptedMatrixData as any"
          :item-type="selectedRoleType"
          :app-type="selectedApp?.type || 'jenkins'"
          @refresh="getMatrixData"
        />

        <ElEmpty v-else-if="selectedAppId && !matrixLoading" description="暂无数据" />
      </div>

      <!-- 列表视图 -->
      <div v-else>
        <!-- 搜索栏 -->
        <div class="mb-4">
          <ElForm :model="searchParams" inline>
            <ElFormItem label="用户名">
              <ElInput v-model="searchParams.username" placeholder="请输入用户名" class="w-200px" clearable />
            </ElFormItem>
            <ElFormItem>
              <ElButton type="primary" @click="handleSearch">查询</ElButton>
              <ElButton @click="handleReset">重置</ElButton>
            </ElFormItem>
          </ElForm>
        </div>

        <ElTable v-loading="loading" :data="tableData" :border="false">
          <ElTableColumn type="index" label="序号" width="60" align="center" />
          <ElTableColumn label="授权中心用户" align="center" min-width="120">
            <template #default="{ row }">
              <div>
                <div class="font-medium">{{ row.username || '-' }}</div>
                <div class="text-xs text-gray-500">{{ row.nickname || '' }}</div>
              </div>
            </template>
          </ElTableColumn>
          <ElTableColumn label="应用" align="center" min-width="120">
            <template #default="{ row }">
              {{ row.app_name || '-' }}
            </template>
          </ElTableColumn>
          <ElTableColumn label="角色名称" align="center" min-width="150">
            <template #default="{ row }">
              <ElTag type="primary">{{ row.role_name }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="角色代码" align="center" min-width="150">
            <template #default="{ row }">
              {{ row.role_code || '-' }}
            </template>
          </ElTableColumn>
          <ElTableColumn label="角色类型" align="center" width="100">
            <template #default="{ row }">
              <ElTag>{{ row.role_type || 'global' }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="权限状态" align="center" width="100">
            <template #default="{ row }">
              <ElTag :type="getStatusTag(row.status).type">
                {{ getStatusTag(row.status).label }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="外部用户名" align="center" min-width="120">
            <template #default="{ row }">
              {{ row.external_username || '-' }}
            </template>
          </ElTableColumn>
          <ElTableColumn label="来源用户组" align="center" min-width="150">
            <template #default="{ row }">
              <div v-if="row.group_name">
                <div>{{ row.group_name }}</div>
                <div class="text-xs text-gray-500">{{ row.group_code }}</div>
              </div>
              <span v-else>-</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="分配时间" align="center" min-width="160">
            <template #default="{ row }">
              {{ row.assigned_at || '-' }}
            </template>
          </ElTableColumn>
        </ElTable>

        <div class="mt-4 flex justify-end">
          <ElPagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :total="pagination.total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="handlePageChange"
            @size-change="handleSizeChange"
          />
        </div>
      </div>
    </ElCard>
  </div>
</template>

<style scoped>
  .permission-tabs-shell {
    display: flex;
    width: 100%;
    padding: 4px;
    border: 1px solid rgba(148, 163, 184, 0.16);
    border-radius: 12px;
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.9));
    box-shadow: 0 12px 26px rgba(15, 23, 42, 0.04);
  }

  .permission-tab-btn {
    min-height: 38px;
    padding: 0 20px;
    border: 0;
    border-radius: 8px;
    background: transparent;
    color: #4e5969;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 13px;
    font-weight: 700;
    line-height: 1.2;
    transition: all 0.2s;
  }

  .permission-tab-btn:hover {
    background: rgba(51, 112, 255, 0.06);
  }

  .permission-tab-btn.active {
    background: #e8f0ff;
    color: #245bdb;
    box-shadow: inset 0 0 0 1px rgba(51, 112, 255, 0.08);
  }

  @media (max-width: 768px) {
    .permission-tabs-shell {
      flex-wrap: wrap;
    }

    .permission-tab-btn {
      min-width: 0;
      flex: 1;
    }
  }
</style>
