<script setup lang="tsx">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { ElNotification } from 'element-plus';
import { useBoolean } from '@sa/hooks';
import { Delete, Management, Plus, Refresh, Search } from '@element-plus/icons-vue';
import { fetchDeleteRole, fetchGetRoleList, fetchGetUserList, fetchUpdateRole } from '@/service/api';
import { useThemeStore } from '@/store/modules/theme';
import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import { useUnifiedPermission } from '@/composables/useUnifiedPermission';
import RoleOperateDrawer from './modules/role-operate-drawer.vue';
import PermissionAssignModal from './modules/permission-assign-modal.vue';

defineOptions({ name: 'RoleManage' });

const themeStore = useThemeStore();

// 使用统一权限检查
const { executeWithPermission } = useUnifiedPermission();

// Hero区域显示状态
const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

// 用户列表和角色-用户映射
const allUsers = ref<Api.SystemManage.User[]>([]);
const roleUsersMap = computed(() => {
  const map = new Map<number, string[]>();
  allUsers.value.forEach(user => {
    if (user.roleIds && user.roleIds.length > 0) {
      user.roleIds.forEach(roleId => {
        if (!map.has(roleId)) {
          map.set(roleId, []);
        }
        map.get(roleId)?.push(user.username);
      });
    }
  });
  // 添加安全检查，避免在路由切换时出错
  if (allUsers.value && Array.isArray(allUsers.value)) {
    allUsers.value.forEach(user => {
      if (user && user.roleIds && user.roleIds.length > 0) {
        user.roleIds.forEach(roleId => {
          if (!map.has(roleId)) {
            map.set(roleId, []);
          }
          map.get(roleId)?.push(user.username);
        });
      }
    });
  }
  return map;
});

// 检查角色是否可以删除
function canDeleteRole(role: Api.SystemManage.Role): { canDelete: boolean; reason?: string } {
  const builtinRoles: Record<string, string> = {
    admin: '超级管理员',
    ops: '运维工程师',
    auditor: '审计员',
    user: '普通用户',
    test: '测试角色'
  };

  if (builtinRoles[role.code]) {
    return {
      canDelete: false,
      reason: `"${builtinRoles[role.code]}" 是系统内置角色，不能删除`
    };
  }

  const users = roleUsersMap.value.get(role.id);
  if (users && users.length > 0) {
    return {
      canDelete: false,
      reason: `该角色已关联 ${users.length} 个用户：${users.slice(0, 3).join('、')}${users.length > 3 ? '...' : ''}。请先解除关联后再删除。`
    };
  }

  return { canDelete: true };
}

// 获取所有用户
async function getAllUsers() {
  const { error, data } = await fetchGetUserList({ current: 1, size: 1000 });
  if (!error && data) {
    allUsers.value = data.records || data || [];
  }
}

onMounted(() => {
  getAllUsers();
});

const searchParams = ref(getInitSearchParams());

function getInitSearchParams(): Api.SystemManage.RoleSearchParams {
  return {
    current: 1,
    size: 10,
    status: undefined,
    name: undefined,
    code: undefined
  };
}

const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
  paginationProps: {
    currentPage: searchParams.value.current,
    pageSize: searchParams.value.size
  },
  api: () => fetchGetRoleList(searchParams.value),
  transform: response => defaultTransform(response),
  onPaginationParamsChange: params => {
    searchParams.value.current = params.currentPage;
    searchParams.value.size = params.pageSize;
  },
  columns: () => [
    { prop: 'selection', type: 'selection', width: 48 },
    { prop: 'index', type: 'index', label: $t('common.index'), width: 64 },
    { prop: 'name', label: $t('page.manage.role.roleName'), minWidth: 120 },
    { prop: 'code', label: $t('page.manage.role.roleCode'), minWidth: 120 },
    { prop: 'description', label: $t('page.manage.role.roleDesc'), minWidth: 120 },
    {
      prop: 'users',
      label: '关联用户',
      minWidth: 150,
      formatter: row => {
        const users = roleUsersMap.value.get(row.id);
        if (!users || users.length === 0) {
          return <span class="pl-12px text-gray">-</span>;
        }

        const displayUsers = users.slice(0, 3);
        const userCount = users.length;
        const moreText = userCount > 3 ? `等${userCount}人` : '';

        return (
          <div class="flex flex-wrap items-center gap-4px pl-12px">
            {displayUsers.map(username => (
              <ElTag key={username} size="small" type="info">
                {username}
              </ElTag>
            ))}
            {moreText && <span class="text-12px text-gray">{moreText}</span>}
          </div>
        );
      }
    },
    {
      prop: 'status',
      label: $t('page.manage.role.roleStatus'),
      align: 'center',
      width: 100,
      formatter: row => {
        if (row.status === undefined) {
          return '';
        }

        return (
          <ElSwitch
            v-model={row.status}
            activeValue={1}
            inactiveValue={0}
            disabled={row.code === 'admin'}
            onChange={(val: number) => handleStatusChange(row, val)}
          />
        );
      }
    },
    {
      prop: 'canDelete',
      label: '状态',
      align: 'center',
      width: 120,
      formatter: row => {
        const { canDelete, reason } = canDeleteRole(row);

        if (!canDelete) {
          let tagType: UI.ThemeColor = 'danger';
          let statusText = '不可删除';

          if (reason?.includes('系统内置')) {
            statusText = '系统内置';
            tagType = 'warning';
          } else if (reason?.includes('已关联')) {
            statusText = '已关联用户';
            tagType = 'info';
          }

          return (
            <ElTag size="small" type={tagType}>
              {statusText}
            </ElTag>
          );
        }

        return (
          <ElTag size="small" type="success">
            可删除
          </ElTag>
        );
      }
    },
    {
      prop: 'operate',
      label: $t('common.operate'),
      align: 'center',
      width: 260,
      formatter: row => (
        <div class="flex-center gap-8px">
          <ElButton type="primary" plain size="small" onClick={() => edit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElButton type="info" plain size="small" onClick={() => handleViewPermissions(row.id)}>
            查看权限
          </ElButton>
          <ElButton type="success" plain size="small" onClick={() => handleAssignPermissions(row.id)}>
            分配权限
          </ElButton>
          <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handleDelete(row.id)}>
            {{
              reference: () => (
                <ElButton type="danger" plain size="small">
                  {$t('common.delete')}
                </ElButton>
              )
            }}
          </ElPopconfirm>
        </div>
      )
    }
  ]
});

const { drawerVisible, operateType, editingData, handleAdd, handleEdit, checkedRowKeys, onBatchDeleted, onDeleted } =
  useTableOperate(data, 'id', getData);

// 角色操作完成后的数据刷新
async function handleRoleOperateSubmitted() {
  await getData();
  await getAllUsers();
}

// 统一权限分配
const { bool: permissionModalVisible, setTrue: openPermissionModal } = useBoolean();
const currentRoleId = ref<number>(-1);
const currentRoleData = ref<Api.SystemManage.Role | null>(null);
const isViewMode = ref(false); // 是否为只读查看模式

function handleAssignPermissions(id: number) {
  const role = data.value.find(r => r.id === id);
  if (role) {
    currentRoleId.value = id;
    currentRoleData.value = role;
    isViewMode.value = false; // 编辑模式
    openPermissionModal();
  }
}

function handleViewPermissions(id: number) {
  const role = data.value.find(r => r.id === id);
  if (role) {
    currentRoleId.value = id;
    currentRoleData.value = role;
    isViewMode.value = true; // 只读查看模式
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

      const { canDelete, reason } = canDeleteRole(role);
      if (!canDelete) {
        cannotDeleteRoles.push({
          id: role.id,
          name: role.name,
          reason: reason || '不可删除'
        });
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

    const { canDelete, reason } = canDeleteRole(role);
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
        message: error.msg || '删除角色失败',
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
    searchParams.value.current = 1;
    getDataByPage();
  }, 300);
}

// 手动搜索（点击搜索按钮）
function handleSearch() {
  searchParams.value.current = 1;
  getDataByPage();
}

// 组件卸载时清理资源
onUnmounted(() => {
  // 清理搜索定时器，避免内存泄漏
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
            <ElIcon><Management /></ElIcon>
          </span>
          <h2>{{ $t('page.manage.role.title') }}</h2>
          <p class="hero-desc">统一管理角色权限，支持角色创建、编辑与权限分配</p>
        </div>
      </div>
      <div class="hero-actions">
        <ElButton size="small" :loading="loading" @click="refreshData">
          <ElIcon><Refresh /></ElIcon>
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
            <ElIcon><Plus /></ElIcon>
            新增角色
          </ElButton>
          <ElButton type="danger" size="small" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
            <ElIcon><Delete /></ElIcon>
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
            <ElIcon><Refresh /></ElIcon>
            重置
          </ElButton>
          <ElButton class="filter-refresh-btn" type="primary" @click="handleSearch">
            <ElIcon><Search /></ElIcon>
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
/* ============================================
	   1. 页面容器与布局
	   ============================================ */
.role-management-page {
  padding: 16px 20px;
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* ============================================
	   2. Hero 区域
	   ============================================ */
.hero-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.hero-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.hero-title-row {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.hero-icon {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--sx-hero-icon-color);
  background: var(--sx-hero-icon-bg);
  border: 1px solid var(--sx-hero-icon-border);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.hero-title-row h2 {
  color: var(--sx-text-primary);
  font-size: 23px;
  font-weight: 700;
  margin: 0;
}

.hero-desc {
  color: var(--sx-text-secondary);
  font-size: 13px;
  line-height: 1.45;
  margin: 0;
  max-width: 600px;
}

.hero-actions {
  display: flex;
  gap: 4px;
}

/* ============================================
	   3. 内容卡片
	   ============================================ */
.content-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--sx-border-soft);
}

.toolbar-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.toolbar-title {
  color: var(--sx-text-primary);
  font-size: 16px;
  font-weight: 700;
}

.toolbar-desc {
  color: var(--sx-text-secondary);
  font-size: 12px;
}

.toolbar-actions {
  display: flex;
  gap: 4px;
  align-items: center;
}

/* ============================================
	   4. 搜索工具栏
	   ============================================ */
.workbench-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin: 6px 0 8px;
}

.workbench-toolbar-left,
.workbench-toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.workbench-toolbar.workbench-toolbar--history {
  background: var(--sx-toolbar-bg);
  border: 1px solid var(--sx-toolbar-border);
  border-radius: var(--sx-toolbar-radius);
  padding: var(--sx-toolbar-padding);
  box-shadow: var(--sx-toolbar-shadow);
  gap: 8px;
  margin: 6px 0;

  .workbench-toolbar-left,
  .workbench-toolbar-right {
    gap: 5px;
  }

  .el-input__wrapper {
    background: var(--sx-search-input-bg);
    border-radius: var(--sx-search-input-radius);
    box-shadow: 0 0 0 1px var(--sx-search-input-border) inset;
    transition: all 0.3s ease;

    &:hover {
      box-shadow: 0 0 0 1px var(--sx-search-input-hover-border) inset;
    }
  }

  .el-input.is-focus .el-input__wrapper {
    box-shadow: 0 0 0 1px var(--sx-search-input-focus-border) inset;
  }

  .el-button:not(.is-link) {
    background: var(--sx-search-button-bg);
    color: var(--sx-search-button-text);
    transition: all 0.3s ease;

    &:hover {
      background: var(--sx-search-button-hover);
    }
  }
}

/* ============================================
	   5. 数据表格
	   ============================================ */
.table-section {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.data-table {
  width: 100%;
  border-radius: var(--sx-table-radius);
  overflow: hidden;
  border: 1px solid var(--sx-table-border);
  border-collapse: separate;
  border-spacing: 0;

  --el-table-border-color: var(--sx-table-border);
  --el-table-header-bg-color: var(--sx-table-header-bg);
  --el-table-row-hover-bg-color: var(--sx-table-row-hover);

  :deep(th.el-table__cell) {
    background-color: var(--sx-table-header-bg);
    color: var(--sx-table-header-text);
    font-weight: 600;
    border-radius: 0;
  }

  :deep(td.el-table__cell) {
    border-radius: 0;
  }

  :deep(.el-checkbox__inner),
  :deep(.el-checkbox__inner::before),
  :deep(.el-checkbox__inner::after) {
    border-radius: 0;
  }

  &--striped :deep(.el-table__body tr.el-table__row--striped td) {
    background-color: var(--sx-table-striped-bg);
  }
}

/* ============================================
	   6. 分页组件
	   ============================================ */
.table-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;

  .el-pagination {
    .btn-next,
    .btn-prev,
    .el-pager li {
      background: var(--sx-pagination-button-bg);
      color: var(--sx-pagination-button-text);
      border-radius: var(--sx-pagination-radius);
      transition: all 0.3s ease;

      &:hover {
        background: var(--sx-pagination-button-hover);
      }

      &.is-active {
        background: var(--sx-pagination-active-bg);
        color: var(--sx-pagination-active-text);
      }
    }
  }
}

/* ============================================
	   7. 响应式设计
	   ============================================ */
@media (max-width: 768px) {
  .hero-section {
    flex-direction: column;
    align-items: flex-start;
  }

  .card-toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
}
</style>
