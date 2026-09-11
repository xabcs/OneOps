<script setup lang="tsx">
  import { computed, onMounted, ref } from 'vue';
  import {
    fetchApplications,
    fetchUserEffectivePermissions,
    fetchUserEffectivePermissionsMatrix
  } from '@/service/api/application-permission';
  import { createTagMap } from '@/utils/common';
  import AppPermissionMatrix from './modules/AppPermissionMatrix.vue';

  defineOptions({ name: 'AuthUserPermissions' });

  const loading = ref(false);
  const matrixLoading = ref(false);
  // 应用列表初始加载状态：applications 返回前保持 loading，避免闪现"请选择应用"的误导性空态
  const appLoading = ref(true);
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
    appLoading.value = true;
    try {
      const { data, error } = await fetchApplications({ page: 1, pageSize: 100 });
      if (!error && data) {
        applications.value = data.list || [];
        // 默认选择第一个应用
        if (applications.value.length > 0 && !selectedAppId.value) {
          selectedAppId.value = applications.value[0].id;
          handleAppChange();
        }
      }
    } finally {
      appLoading.value = false;
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

  /** 权限状态 → ElTag 标签映射 */
  const getStatusTag = createTagMap({
    active: { text: '有效', type: 'success' },
    inactive: { text: '无效', type: 'info' },
    expired: { text: '已过期', type: 'danger' },
    pending: { text: '待生效', type: 'warning' }
  });

  onMounted(() => {
    getApplications();
  });
</script>

<template>
  <ListPageLayout
    title="用户权限"
    description="查看用户在各应用下的有效权限与角色分配"
    :pagination="
      viewMode === 'list' && pagination.total
        ? {
            total: pagination.total,
            currentPage: pagination.page,
            pageSize: pagination.pageSize,
            pageSizes: [10, 20, 50, 100],
            'current-change': handlePageChange,
            'size-change': handleSizeChange
          }
        : null
    "
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- Tab 切换层 -->
    <template #hero>
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
    </template>

    <!-- 列表视图筛选（矩阵视图不渲染搜索区） -->
    <template v-if="viewMode === 'list'" #search>
      <ElInput v-model="searchParams.username" placeholder="请输入用户名" class="w-200px" clearable />
    </template>

    <!-- 矩阵视图：初始加载阶段（applications 未返回）同样显示 loading，不闪现误导性空态 -->
    <div v-if="viewMode === 'matrix'" v-loading="matrixLoading || appLoading">
      <ElEmpty v-if="!selectedAppId && !appLoading" description="请选择应用" />

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
    <ElTable v-else v-loading="loading" :data="tableData" :border="false" height="100%">
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
                {{ getStatusTag(row.status).text }}
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
  </ListPageLayout>
</template>

<style scoped>
  .permission-tabs-shell {
    display: flex;
    width: 100%;
    padding: 4px;
    border: 1px solid rgb(148 163 184 / 16%);
    border-radius: 12px;
    background: linear-gradient(180deg, rgb(255 255 255 / 96%), rgb(248 250 252 / 90%));
    box-shadow: 0 12px 26px rgb(15 23 42 / 4%);
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
    background: rgb(51 112 255 / 6%);
  }

  .permission-tab-btn.active {
    background: #e8f0ff;
    color: #245bdb;
    box-shadow: inset 0 0 0 1px rgb(51 112 255 / 8%);
  }

  @media (width <= 768px) {
    .permission-tabs-shell {
      flex-wrap: wrap;
    }

    .permission-tab-btn {
      min-width: 0;
      flex: 1;
    }
  }
</style>
