<script setup lang="tsx">
  import { computed, onUnmounted, ref } from 'vue';
  import { ElMessage, ElMessageBox, ElNotification } from 'element-plus';
  import { useBoolean } from '@sa/hooks';
  import { Delete, Management, Plus, Refresh, Search } from '@element-plus/icons-vue';
  import { fetchDeleteRole, fetchGetRoleList, fetchUpdateRole } from '@/service/api';
  import { useThemeStore } from '@/store/modules/theme';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import { executeWithPermission } from '@/hooks/business/auth';
  import { $t } from '@/locales';
  import RoleOperateDrawer from './modules/role-operate-drawer.vue';
  import PermissionAssignModal from './modules/permission-assign-modal.vue';
  import RoleDetailDrawer from './modules/role-detail-drawer.vue';
  import { createRoleColumns } from './modules/role-columns';

  defineOptions({ name: 'RoleManage' });

  const themeStore = useThemeStore();

  // Hero区域显示状态
  const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

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
  }

  // 统一权限分配
  const { bool: permissionModalVisible, setTrue: openPermissionModal } = useBoolean();
  const currentRoleId = ref<number>(-1);
  const currentRoleData = ref<Api.SystemManage.Role | null>(null);
  function handleAssignPermissions(id: number) {
    const role = data.value.find(r => r.id === id);
    if (role) {
      currentRoleId.value = id;
      currentRoleData.value = role;
      openPermissionModal();
    }
  }

  // 角色详情抽屉（点击角色名打开）
  const { bool: detailDrawerVisible, setTrue: openDetailDrawer } = useBoolean();
  const currentDetailRole = ref<Api.SystemManage.Role | null>(null);

  function handleViewDetail(row: Api.SystemManage.Role) {
    currentDetailRole.value = row;
    openDetailDrawer();
  }

  async function handlePermissionSubmitted() {
    await getData();
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
      let successCount = 0;
      let failCount = 0;

      for (const id of checkedRowKeys.value) {
        const role = data.value.find(r => r.id === id);
        const roleName = role ? role.name : `ID:${id}`;

        const { error } = await fetchDeleteRole(id as number);
        if (!error) {
          successCount++;
        } else {
          failCount++;
          cannotDeleteRoles.push({
            id: id as number,
            name: roleName,
            reason: error?.response?.data?.message || error?.message || '删除失败'
          });
        }
      }

      if (cannotDeleteRoles.length > 0) {
        const message = cannotDeleteRoles.map(r => `• ${r.name}: ${r.reason}`).join('\n');
        ElNotification({
          title: `${cannotDeleteRoles.length} 个角色删除失败`,
          message,
          type: 'warning',
          duration: 5000,
          position: 'top-right'
        });
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
    });
  }

  async function handleDelete(id: number) {
    await ElMessageBox.confirm('确认删除吗？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });
    await executeWithPermission('system.role.delete', async () => {
      const role = data.value.find(r => r.id === id);
      if (!role) {
        window.$message?.error('角色不存在');
        return;
      }

      const { error } = await fetchDeleteRole(id);

      if (!error) {
        ElMessage.success(`角色 "${role.name}" 已成功删除`);
        onDeleted();
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
      edit,
      handleViewDetail,
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
                        <Management />
                    </ElIcon>
                    <div class="flex flex-col gap-2px">
                        <h2 class="text-18px font-bold m-0">{{ $t('page.manage.role.title') }}</h2>
                        <p class="text-13px opacity-70 m-0">统一管理角色权限，支持角色创建、编辑与权限分配</p>
                    </div>
                </div>
                <ElButton size="small" :loading="loading" @click="refreshData">
                    <ElIcon>
                        <Refresh />
                    </ElIcon>
                    刷新
                </ElButton>
            </div>
        </ElCard>

        <!-- 内容卡片 -->
        <ElCard shadow="hover">
            <template #header>
                <div class="flex items-center justify-between">
                    <div class="flex flex-col gap-2px">
                        <span class="text-16px font-bold">角色列表</span>
                        <span class="text-13px opacity-70">管理系统角色、分配菜单权限与用户关联</span>
                    </div>
                    <div class="flex items-center gap-8px">
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
            </template>

            <!-- 搜索工具栏 -->
            <ElSpace wrap class="mb-16px">
                <ElInput v-model="searchParams.name" placeholder="搜索角色名称" clearable style="width: 200px" :prefix-icon="Search" @input="handleSearchInput" />
                <ElInput v-model="searchParams.code" placeholder="搜索角色编码" clearable style="width: 200px" :prefix-icon="Search" @input="handleSearchInput" />
                <ElButton @click="resetSearchParams">
                    <ElIcon>
                        <Refresh />
                    </ElIcon>
                    重置
                </ElButton>
                <ElButton type="primary" @click="handleSearch">
                    <ElIcon>
                        <Search />
                    </ElIcon>
                    搜索
                </ElButton>
            </ElSpace>

            <!-- 数据表格 -->
            <ElTable v-loading="loading" :data="data" border stripe row-key="id" @selection-change="checkedRowKeys = $event">
                <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
            </ElTable>

            <!-- 分页 -->
            <div v-if="mobilePagination.total" class="flex justify-end mt-16px">
                <ElPagination layout="total, sizes, prev, pager, next, jumper" v-bind="mobilePagination" @current-change="mobilePagination['current-change']" @size-change="mobilePagination['size-change']" />
            </div>
        </ElCard>

        <!-- 抽屉和模态框 -->
        <RoleOperateDrawer v-model:visible="drawerVisible" :operate-type="operateType" :row-data="editingData" @submitted="handleRoleOperateSubmitted" />
        <PermissionAssignModal v-model:visible="permissionModalVisible" :role-id="currentRoleId" :role-data="currentRoleData" @submitted="handlePermissionSubmitted" />
        <RoleDetailDrawer v-model:visible="detailDrawerVisible" :role="currentDetailRole" />
    </div>
</template>
