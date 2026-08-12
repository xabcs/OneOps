<script setup lang="tsx">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { ElNotification } from 'element-plus';
  import { useBoolean } from '@sa/hooks';
  import { Delete, Management, Plus, Refresh, Search } from '@element-plus/icons-vue';
  import { fetchDeleteRole, fetchGetRoleList, fetchUpdateRole } from '@/service/api';
  import { useThemeStore } from '@/store/modules/theme';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import { executeWithPermission } from '@/hooks/business/auth';
  import { $t } from '@/locales';
  import RoleOperateDrawer from './modules/role-operate-drawer.vue';
  import PermissionAssignModal from './modules/permission-assign-modal.vue';
  import { canDeleteRole, useRoleUsersMap } from './modules/role-helper';
  import { createRoleColumns } from './modules/role-columns';

  defineOptions({ name: 'RoleManage' });

  const themeStore = useThemeStore();

  // Hero区域显示状态
  const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

  // 用户列表和角色-用户映射
  const { roleUsersMap, getAllUsers } = useRoleUsersMap();

  onMounted(() => {
    getAllUsers();
  });

  const searchParams = ref(getInitSearchParams());

  function getInitSearchParams(): Api.SystemManage.RoleSearchParams {
    return {
      page: 1,
      pageSize: 10,
      status: undefined,
      name: undefined,
      code: undefined
    };
  }

  // 角色操作完成后的数据刷新
  async function handleRoleOperateSubmitted() {
    await getData();
    await getAllUsers();
  }

  // 统一权限分配
  const { bool: permissionModalVisible, setTrue: openPermissionModal } = useBoolean();
  const currentRoleId = ref<number>(-1);
  const currentRoleData = ref<Api.SystemManage.Role | null>(null);
  const isViewMode = ref(false);

  function handleAssignPermissions(id: number) {
    const role = data.value.find(r => r.id === id);
    if (role) {
      currentRoleId.value = id;
      currentRoleData.value = role;
      isViewMode.value = false;
      openPermissionModal();
    }
  }

  function handleViewPermissions(id: number) {
    const role = data.value.find(r => r.id === id);
    if (role) {
      currentRoleId.value = id;
      currentRoleData.value = role;
      isViewMode.value = true;
      openPermissionModal();
    }
  }

  async function handlePermissionSubmitted() {
    await getData();
    await getAllUsers();
  }

  async function handleBatchDelete() {
    await executeWithPermission('system.role.batch_delete', async () => {
      if (checkedRowKeys.value.length === 0) {
        ElNotification({
          title: '提示',
          message: '请选择要删除的角色',
          type: 'warning',
          duration: 3000,
          position: 'top-right'
        });
        return;
      }

      const cannotDeleteRoles: Array<{ id: number; name: string; reason: string }> = [];
      const canDeleteIds: number[] = [];

      for (const id of checkedRowKeys.value) {
        const role = data.value.find(r => r.id === id);
        if (!role) {
          cannotDeleteRoles.push({ id: id as number, name: `ID:${id}`, reason: '角色不存在' });
          continue;
        }

        const { canDelete, reason } = canDeleteRole(role, roleUsersMap.value);
        if (!canDelete) {
          cannotDeleteRoles.push({ id: role.id, name: role.name, reason: reason || '不可删除' });
        } else {
          canDeleteIds.push(role.id);
        }
      }

      if (cannotDeleteRoles.length > 0) {
        const message = cannotDeleteRoles.map(r => `• ${r.name}: ${r.reason}`).join('\n');
        ElNotification({
          title: `无法删除 ${cannotDeleteRoles.length} 个角色`,
          message,
          type: 'warning',
          duration: 5000,
          position: 'top-right'
        });

        if (canDeleteIds.length === 0) {
          return;
        }
      }

      let successCount = 0;
      let failCount = 0;

      for (const id of canDeleteIds) {
        const { error } = await fetchDeleteRole(id);
        if (!error) {
          successCount++;
        } else {
          failCount++;
        }
      }

      if (successCount > 0) {
        ElNotification({
          title: '批量删除完成',
          message: `成功删除 ${successCount} 个角色`,
          type: 'success',
          duration: 3000,
          position: 'top-right'
        });
      }

      if (failCount > 0) {
        ElNotification({
          title: '部分删除失败',
          message: `${failCount} 个角色删除失败`,
          type: 'error',
          duration: 5000,
          position: 'top-right'
        });
      }

      onBatchDeleted();
      await getAllUsers();
    });
  }

  async function handleDelete(id: number) {
    await executeWithPermission('system.role.delete', async () => {
      const role = data.value.find(r => r.id === id);
      if (!role) {
        window.$message?.error('角色不存在');
        return;
      }

      const { canDelete, reason } = canDeleteRole(role, roleUsersMap.value);
      if (!canDelete) {
        ElNotification({
          title: '无法删除角色',
          message: reason || '该角色不可删除',
          type: 'warning',
          duration: 3000,
          position: 'top-right'
        });
        return;
      }

      const { error } = await fetchDeleteRole(id);

      if (!error) {
        ElNotification({
          title: '删除成功',
          message: `角色 "${role.name}" 已成功删除`,
          type: 'success',
          duration: 3000,
          position: 'top-right'
        });
        onDeleted();
        await getAllUsers();
      } else {
        ElNotification({
          title: '删除失败',
          message: error?.response?.data?.message || error?.message || '删除角色失败',
          type: 'error',
          duration: 3000,
          position: 'top-right'
        });
      }
    });
  }

  async function handleStatusChange(row: Api.SystemManage.Role, val: number) {
    await executeWithPermission('system.role.update', async () => {
      const { error } = await fetchUpdateRole(row.id, { status: val });

      if (!error) {
        window.$message?.success(`${val === 1 ? '启用' : '禁用'}成功`);
      } else {
        row.status = val === 1 ? 0 : 1;
        window.$message?.error('状态更新失败');
      }
    });
  }

  function resetSearchParams() {
    searchParams.value = getInitSearchParams();
  }

  async function edit(id: number) {
    await executeWithPermission('system.role.update', async () => {
      handleEdit(id);
    });
  }

  async function handleAddClick() {
    await executeWithPermission('system.role.create', async () => {
      handleAdd();
    });
  }

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize
    },
    api: () => fetchGetRoleList(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage;
      searchParams.value.pageSize = params.pageSize;
    },
    columns: createRoleColumns({
      roleUsersMap: () => roleUsersMap.value,
      edit,
      handleViewPermissions,
      handleAssignPermissions,
      handleDelete,
      handleStatusChange
    })
  });

  const { drawerVisible, operateType, editingData, handleAdd, handleEdit, checkedRowKeys, onBatchDeleted, onDeleted } =
    useTableOperate(data, 'id', getData);

  // 刷新数据
  async function refreshData() {
    await getDataByPage();
    await getAllUsers();
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
  <div class="page-container role-management-page">
    <!-- Hero 区域 -->
    <section
      v-if="heroVisible"
      class="hero-section"
      :style="{
        background: 'var(--sx-hero-bg)',
        border: '1px solid var(--sx-hero-border)',
        borderRadius: 'var(--sx-hero-radius)',
        boxShadow: 'var(--sx-hero-shadow)',
        padding: 'var(--sx-hero-padding)'
      }"
    >
      <div class="hero-content">
        <div class="hero-title-row">
          <span class="hero-icon">
            <ElIcon>
              <Management />
            </ElIcon>
          </span>
          <h2>{{ $t('page.manage.role.title') }}</h2>
          <p class="hero-desc">统一管理角色权限，支持角色创建、编辑与权限分配</p>
        </div>
      </div>
      <div class="hero-actions">
        <ElButton size="small" :loading="loading" @click="refreshData">
          <ElIcon>
            <Refresh />
          </ElIcon>
          刷新
        </ElButton>
      </div>
    </section>

    <!-- 内容卡片 -->
    <div
      class="content-card role-content-card"
      :style="{
        background: 'var(--sx-content-card-bg)',
        border: '1px solid var(--sx-content-card-border)',
        borderRadius: 'var(--sx-content-card-radius)',
        boxShadow: 'var(--sx-content-card-shadow)',
        padding: 'var(--sx-content-card-padding)'
      }"
    >
      <!-- 工具栏 -->
      <div class="card-toolbar">
        <div class="toolbar-head">
          <span class="toolbar-title">角色列表</span>
          <span class="toolbar-desc">管理系统角色、分配菜单权限与用户关联</span>
        </div>
        <div class="toolbar-actions">
          <ElButton type="primary" size="small" @click="handleAddClick">
            <ElIcon>
              <Plus />
            </ElIcon>
            新增角色
          </ElButton>
          <ElButton type="danger" size="small" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
            <ElIcon>
              <Delete />
            </ElIcon>
            批量删除
          </ElButton>
        </div>
      </div>

      <!-- 搜索工具栏 -->
      <div class="workbench-toolbar workbench-toolbar--history roles-toolbar">
        <div class="workbench-toolbar-left">
          <ElInput
            v-model="searchParams.name"
            placeholder="搜索角色名称"
            clearable
            style="width: 200px"
            @input="handleSearchInput"
          />
          <ElInput
            v-model="searchParams.code"
            placeholder="搜索角色编码"
            clearable
            style="width: 200px"
            @input="handleSearchInput"
          />
        </div>
        <div class="workbench-toolbar-right">
          <ElButton class="filter-refresh-btn" @click="resetSearchParams">
            <ElIcon>
              <Refresh />
            </ElIcon>
            重置
          </ElButton>
          <ElButton class="filter-refresh-btn" type="primary" @click="handleSearch">
            <ElIcon>
              <Search />
            </ElIcon>
            搜索
          </ElButton>
        </div>
      </div>

      <!-- 数据表格 -->
      <div class="table-section">
        <ElTable
          v-loading="loading"
          :data="data"
          border
          stripe
          class="data-table"
          row-key="id"
          @selection-change="checkedRowKeys = $event"
        >
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>

      <!-- 分页 -->
      <div class="table-pagination">
        <ElPagination
          v-if="mobilePagination.total"
          layout="total, sizes, prev, pager, next"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
    </div>

    <!-- 抽屉和模态框 -->
    <RoleOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="handleRoleOperateSubmitted"
    />
    <PermissionAssignModal
      v-model:visible="permissionModalVisible"
      :role-id="currentRoleId"
      :role-data="currentRoleData"
      :view-only="isViewMode"
      @submitted="handlePermissionSubmitted"
    />
  </div>
</template>

<style scoped lang="scss">
  @use './modules/role-page.scss';
</style>
