<script setup lang="tsx">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { ElMessageBox } from 'element-plus';
  import { Delete, Plus, Refresh, Search, User } from '@element-plus/icons-vue';
  import { fetchDeleteUser, fetchGetAllRoles, fetchGetUserList } from '@/service/api';
  import { useThemeStore } from '@/store/modules/theme';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import { executeWithPermission } from '@/hooks/business/auth';
  import { $t } from '@/locales';
  import UserOperateDrawer from './modules/user-operate-drawer.vue';
  import ResetPasswordModal from './modules/reset-password-modal.vue';
  import UserStats from './modules/user-stats.vue';
  import { createUserColumns } from './modules/user-columns';

  defineOptions({ name: 'UserManage' });

  const themeStore = useThemeStore();

  // Hero区域显示状态
  const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

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

  // 获取所有角色
  async function getAllRoles() {
    const { error, data } = await fetchGetAllRoles();
    if (!error && data) {
      allRoles.value = data.list || [];
    } else {
      allRoles.value = [];
    }
  }

  // 更新用户统计数据
  async function updateUserStats() {
    const { error, data } = await fetchGetUserList({
      page: 1,
      pageSize: 1,
      status: undefined,
      username: undefined,
      nickname: undefined,
      email: undefined
    });

    if (!error && data) {
      userStats.value.total = data.total || 0;

      const activeResult = await fetchGetUserList({
        page: 1,
        pageSize: 1,
        status: 'active',
        username: undefined,
        nickname: undefined,
        email: undefined
      });

      if (!activeResult.error && activeResult.data) {
        userStats.value.active = activeResult.data.total || 0;
      }

      userStats.value.inactive = userStats.value.total - userStats.value.active;
    }
  }

  onMounted(() => {
    getAllRoles();
    updateUserStats();
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

  function handleResetPasswordSubmitted() {
    getDataByPage();
  }

  async function handleBatchDelete() {
    await executeWithPermission('system.user.batch_delete', async () => {
      if (checkedRowKeys.value.length === 0) {
        window.$message?.warning('请选择要删除的用户');
        return;
      }

      let successCount = 0;
      let failCount = 0;

      for (const id of checkedRowKeys.value) {
        const { error } = await fetchDeleteUser(id as number);
        if (!error) {
          successCount++;
        } else {
          failCount++;
        }
      }

      if (successCount > 0) {
        window.$message?.success(`成功删除 ${successCount} 个用户`);
      }

      if (failCount > 0) {
        window.$message?.error(`${failCount} 个用户删除失败`);
      }

      onBatchDeleted();
      await getAllRoles();
      await updateUserStats();
    });
  }

  async function handleDelete(id: number) {
    await ElMessageBox.confirm('确认删除吗？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });
    await executeWithPermission(
      'system.user.delete',
      async () => {
        const { error } = await fetchDeleteUser(id);

        if (!error) {
          window.$message?.success($t('common.deleteSuccess'));
          onDeleted();
          await getAllRoles();
          await updateUserStats();
        }
      },
      { type: 'error' }
    );
  }

  function resetSearchParams() {
    searchParams.value = getInitSearchParams();
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

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
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
      handleDelete
    })
  });

  const { drawerVisible, operateType, editingData, handleAdd, handleEdit, checkedRowKeys, onBatchDeleted, onDeleted } =
    useTableOperate(data, 'id', getData);

  // 刷新数据
  async function refreshData() {
    await updateUserStats();
    await getDataByPage();
  }

  // 搜索输入处理（防抖）
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  function handleSearchInput() {
    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }
    searchTimeout = setTimeout(() => {
      searchParams.value.page = 1;
      getDataByPage();
    }, 300);
  }

  // 手动搜索
  function handleSearch() {
    searchParams.value.page = 1;
    getDataByPage();
  }

  onUnmounted(() => {
    if (searchTimeout) {
      clearTimeout(searchTimeout);
      searchTimeout = null;
    }
  });
</script>

<template>
  <div class="flex flex-col gap-16px">
    <!-- Hero 区域 -->
    <ElCard v-if="heroVisible" shadow="hover">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-12px">
          <ElIcon :size="24">
            <User />
          </ElIcon>
          <div class="flex flex-col gap-4px">
            <h2 class="text-18px font-bold m-0">{{ $t('page.manage.user.title') }}</h2>
            <p class="text-13px color-[var(--el-text-color-secondary)] m-0">
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
    <ElCard shadow="hover">
      <ElSpace wrap class="w-full" align="center">
        <span class="text-16px font-bold whitespace-nowrap">搜索筛选</span>
        <ElInput
          v-model="searchParams.username"
          placeholder="搜索用户名"
          clearable
          style="width: 200px"
          @input="handleSearchInput"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>
        <ElInput
          v-model="searchParams.nickname"
          placeholder="搜索昵称"
          clearable
          style="width: 200px"
          @input="handleSearchInput"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>
        <ElInput
          v-model="searchParams.email"
          placeholder="搜索邮箱"
          clearable
          style="width: 240px"
          @input="handleSearchInput"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>
        <ElButton @click="resetSearchParams">
          <template #icon>
            <ElIcon><Refresh /></ElIcon>
          </template>
          重置
        </ElButton>
        <ElButton type="primary" @click="handleSearch">
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
            <span class="text-13px color-[var(--el-text-color-secondary)]">
              管理系统用户账号、角色分配与状态控制
            </span>
          </div>
          <ElSpace>
            <ElButton type="primary" size="small" @click="handleAddClick">
              <template #icon>
                <ElIcon><Plus /></ElIcon>
              </template>
              新增用户
            </ElButton>
            <ElButton type="danger" size="small" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
              <template #icon>
                <ElIcon><Delete /></ElIcon>
              </template>
              批量删除
            </ElButton>
          </ElSpace>
        </div>
      </template>

      <ElTable
        v-loading="loading"
        :data="data"
        border
        stripe
        row-key="id"
        @selection-change="checkedRowKeys = $event"
      >
        <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
      </ElTable>

      <div class="flex justify-end mt-16px">
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
      @submitted="getDataByPage"
    />
    <ResetPasswordModal
      v-model:visible="resetPasswordVisible"
      :user-id="resetPasswordUserId"
      :username="resetPasswordUsername"
      @submitted="handleResetPasswordSubmitted"
    />
  </div>
</template>
