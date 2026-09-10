<script setup lang="tsx">
  import { computed, onActivated, onMounted, ref } from 'vue';
  import { useDebounceFn } from '@vueuse/core';
  import { ElMessageBox } from 'element-plus';
  import { Delete, Plus, Refresh, Search, User } from '@element-plus/icons-vue';
  import { fetchDeleteUser, fetchGetAllRoles, fetchGetUserList, fetchUpdateUser } from '@/service/api';
  import { useThemeStore } from '@/store/modules/theme';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import { executeWithPermission } from '@/hooks/business/auth';
  import { createCachedRequest } from '@/utils/request-cache';
  import { $t } from '@/locales';
  import UserOperateDrawer from './modules/user-operate-drawer.vue';
  import ResetPasswordModal from './modules/reset-password-modal.vue';
  import UserStats from './modules/user-stats.vue';
  import { createUserColumns } from './modules/user-columns';

  defineOptions({ name: 'UserManage' });

  const themeStore = useThemeStore();

  // Hero区域显示状态
  const heroVisible = computed(() => themeStore.content.hero.visible !== false);

  // 用户统计数据
  const userStats = ref({
    total: 0,
    active: 0,
    inactive: 0
  });

  // 角色列表和映射
  const allRoles = ref<Api.SystemManage.AllRole[]>([]);
  const roleMap = computed(() => {
    const map = new Map<number, Api.SystemManage.AllRole>();
    if (allRoles.value && Array.isArray(allRoles.value)) {
      allRoles.value.forEach(role => {
        if (role && role.id) {
          map.set(role.id, role);
        }
      });
    }
    return map;
  });

  const roleListCache = createCachedRequest(async () => {
    const { error, data } = await fetchGetAllRoles();

    if (error || !data) {
      throw new Error('Failed to load roles');
    }

    return data.list || [];
  });

  // 获取所有角色；同一时间窗内复用结果，避免反复初始化时重复请求
  async function getAllRoles() {
    try {
      allRoles.value = await roleListCache.load();
    } catch {
      allRoles.value = [];
    }
  }

  // 更新用户统计数据
  async function updateUserStats() {
    const [totalResult, activeResult] = await Promise.all([
      fetchGetUserList({
        page: 1,
        pageSize: 1,
        status: undefined,
        username: undefined,
        nickname: undefined,
        email: undefined
      }),
      fetchGetUserList({
        page: 1,
        pageSize: 1,
        status: 'active',
        username: undefined,
        nickname: undefined,
        email: undefined
      })
    ]);

    const { error, data } = totalResult;
    const { error: activeError, data: activeData } = activeResult;

    if (!error && data) {
      userStats.value.total = data.total || 0;
      userStats.value.active = !activeError && activeData ? activeData.total || 0 : 0;
      userStats.value.inactive = userStats.value.total - userStats.value.active;
    }
  }

  async function refreshAuxiliaryData() {
    await Promise.all([getAllRoles(), updateUserStats()]);
  }

  onMounted(() => {
    refreshAuxiliaryData();
  });

  const searchParams = ref(getInitSearchParams());

  function getInitSearchParams(): Api.SystemManage.UserSearchParams {
    return {
      page: 1,
      pageSize: 10,
      status: undefined,
      username: undefined,
      nickname: undefined,
      email: undefined
    };
  }

  // 重置密码相关
  const resetPasswordVisible = ref(false);
  const resetPasswordUserId = ref(-1);
  const resetPasswordUsername = ref('');

  async function openResetPassword(row: Api.SystemManage.User) {
    await executeWithPermission('system.user.reset_password', async () => {
      resetPasswordUserId.value = row.id;
      resetPasswordUsername.value = row.username || '';
      resetPasswordVisible.value = true;
    });
  }

  function resetSearchParams() {
    searchParams.value = getInitSearchParams();
  }

  async function handleStatusChange(row: Api.SystemManage.User, val: string) {
    // 禁用后该用户无法登录，已登录会话也会立即失效，需二次确认
    if (val === 'inactive') {
      const confirmed = await ElMessageBox.confirm(
        `禁用用户 "${row.username}" 后，该用户将无法登录，已登录的会话也会立即失效，确认禁用吗？`,
        '禁用确认',
        {
          type: 'warning',
          confirmButtonText: '确认禁用',
          cancelButtonText: '取消'
        }
      )
        .then(() => true)
        .catch(() => false);

      if (!confirmed) {
        row.status = 'active'; // 取消时回滚开关状态
        return;
      }
    }

    await executeWithPermission('system.user.update', async () => {
      const { error } = await fetchUpdateUser(row.id, { status: val });

      if (!error) {
        window.$message?.success(`${val === 'active' ? '启用' : '禁用'}成功`);
        await updateUserStats();
      } else {
        row.status = val === 'active' ? 'inactive' : 'active';
      }
    });
  }

  const { columns, data, getData, getDataByPage, loading, pagination, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize
    },
    api: () => fetchGetUserList(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage;
      searchParams.value.pageSize = params.pageSize;
    },
    columns: createUserColumns({
      roleMap: () => roleMap.value,
      edit,
      openResetPassword,
      handleDelete,
      handleStatusChange
    })
  });

  const { drawerVisible, operateType, editingData, handleAdd, handleEdit, checkedRowKeys, onBatchDeleted, onDeleted } =
    useTableOperate(data, 'id', getData);

  async function handleDelete(id: number) {
    await ElMessageBox.confirm('确认删除吗？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });
    await executeWithPermission('system.user.delete', async () => {
      const { error } = await fetchDeleteUser(id);

      if (!error) {
        window.$message?.success($t('common.deleteSuccess'));
        onDeleted();
        await refreshAuxiliaryData();
      }
    });
  }

  function handleResetPasswordSubmitted() {
    getDataByPage();
  }

  async function handleBatchDelete() {
    await executeWithPermission('system.user.batch_delete', async () => {
      if (checkedRowKeys.value.length === 0) {
        window.$message?.warning('请选择要删除的用户');
        return;
      }

      const results = await Promise.all(checkedRowKeys.value.map(id => fetchDeleteUser(Number(id))));
      const successCount = results.filter(({ error }) => !error).length;
      const failCount = results.length - successCount;

      if (successCount > 0) {
        window.$message?.success(`成功删除 ${successCount} 个用户`);
      }

      if (failCount > 0) {
        window.$message?.error(`${failCount} 个用户删除失败`);
      }

      onBatchDeleted();
      await refreshAuxiliaryData();
    });
  }

  async function edit(id: number) {
    await executeWithPermission('system.user.update', async () => {
      handleEdit(id);
    });
  }

  async function handleAddClick() {
    await executeWithPermission('system.user.create', async () => {
      handleAdd();
    });
  }

  // 刷新数据
  async function refreshData() {
    await Promise.all([updateUserStats(), getDataByPage()]);
  }

  let hasFirstActivation = false;

  // 缓存页重新激活时静默刷新，兼顾返回速度与数据新鲜度
  onActivated(() => {
    if (!hasFirstActivation) {
      hasFirstActivation = true;
      return;
    }

    getAllRoles();
    updateUserStats();
    getDataByPage(pagination.currentPage ?? 1);
  });

  // 搜索输入处理（防抖 300ms）
  const handleSearchInput = useDebounceFn(() => {
    searchParams.value.page = 1;
    getDataByPage();
  }, 300);

  // 手动搜索
  function handleSearch() {
    searchParams.value.page = 1;
    getDataByPage();
  }
</script>

<template>
  <div class="table-page">
    <!-- Hero 区域 -->
    <ElCard v-if="heroVisible" shadow="hover" class="card-static msre-hero">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-12px">
          <ElIcon :size="20">
            <User />
          </ElIcon>
          <div class="flex flex-col gap-4px">
            <h2 class="m-0 text-18px font-bold">{{ $t('page.manage.user.title') }}</h2>
            <p class="m-0 text-13px color-[var(--el-text-color-secondary)]">
              统一维护用户、分配角色与权限，支持账号治理与安全策略管理
            </p>
          </div>
        </div>
        <ElButton size="small" :loading="loading" @click="refreshData">
          <template #icon>
            <ElIcon>
              <Refresh />
            </ElIcon>
          </template>
          刷新
        </ElButton>
      </div>
    </ElCard>

    <!-- 统计卡片 -->
    <UserStats :user-stats="userStats" />

    <!-- 搜索卡片 -->
    <ElCard shadow="hover" class="card-static msre-toolbar">
      <ElSpace wrap class="w-full" align="center">
        <span class="whitespace-nowrap text-14px font-semibold">搜索筛选</span>
        <ElInput
          v-model="searchParams.username"
          placeholder="搜索用户名"
          size="small"
          clearable
          class="w-200px"
          @input="handleSearchInput"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>
        <ElInput
          v-model="searchParams.nickname"
          placeholder="搜索昵称"
          size="small"
          clearable
          class="w-200px"
          @input="handleSearchInput"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>
        <ElInput
          v-model="searchParams.email"
          placeholder="搜索邮箱"
          size="small"
          clearable
          class="w-240px"
          @input="handleSearchInput"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>
        <ElButton size="small" @click="resetSearchParams">
          <template #icon>
            <ElIcon><Refresh /></ElIcon>
          </template>
          重置
        </ElButton>
        <ElButton size="small" type="primary" @click="handleSearch">
          <template #icon>
            <ElIcon><Search /></ElIcon>
          </template>
          搜索
        </ElButton>
      </ElSpace>
    </ElCard>

    <!-- 数据表格卡片 -->
    <ElCard shadow="hover">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex flex-col gap-4px">
            <span class="text-16px font-bold">用户列表</span>
            <span class="text-13px color-[var(--el-text-color-secondary)]">管理系统用户账号、角色分配与状态控制</span>
          </div>
          <ElSpace>
            <PermissionButton code="system.user.create" type="primary" size="small" @click="handleAddClick">
              <template #icon>
                <ElIcon><Plus /></ElIcon>
              </template>
              新增用户
            </PermissionButton>
            <PermissionButton
              code="system.user.delete"
              type="danger"
              size="small"
              :disabled="checkedRowKeys.length === 0"
              @click="handleBatchDelete"
            >
              <template #icon>
                <ElIcon><Delete /></ElIcon>
              </template>
              批量删除
            </PermissionButton>
          </ElSpace>
        </div>
      </template>

      <div class="table-scroll-wrap">
        <ElTable
          v-loading="loading"
          :data="data"
          border
          stripe
          row-key="id"
          height="100%"
          @selection-change="checkedRowKeys = $event"
        >
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>

      <div class="mt-16px flex justify-end">
        <ElPagination
          v-if="mobilePagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
    </ElCard>

    <!-- 抽屉和模态框 -->
    <UserOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      :all-roles="allRoles"
      @submitted="refreshData"
    />
    <ResetPasswordModal
      v-model:visible="resetPasswordVisible"
      :user-id="resetPasswordUserId"
      :username="resetPasswordUsername"
      @submitted="handleResetPasswordSubmitted"
    />
  </div>
</template>
