<script setup lang="tsx">
import { ref, computed, onMounted } from 'vue';
import { useBoolean } from '@sa/hooks';
import { ElNotification } from 'element-plus';
import { fetchGetRoleList, fetchUpdateRole, fetchDeleteRole, fetchGetUserList } from '@/service/api';
import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import RoleOperateDrawer from './modules/role-operate-drawer.vue';
import RoleSearch from './modules/role-search.vue';
import MenuAuthModal from './modules/menu-auth-modal.vue';

defineOptions({ name: 'RoleManage' });

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
  return map;
});

// 检查角色是否可以删除
function canDeleteRole(role: Api.SystemManage.Role): { canDelete: boolean; reason?: string } {
  // 1. 检查是否是系统内置角色（对应后端的5个内置角色）
  const builtinRoles: Record<string, string> = {
    'admin': '超级管理员',
    'ops': '运维工程师',
    'auditor': '审计员',
    'user': '普通用户',
    'test': '测试角色'
  };

  if (builtinRoles[role.code]) {
    return {
      canDelete: false,
      reason: `"${builtinRoles[role.code]}" 是系统内置角色，不能删除`
    };
  }

  // 2. 检查是否有关联用户
  const users = roleUsersMap.value.get(role.id);
  if (users && users.length > 0) {
    return {
      canDelete: false,
      reason: `该角色已关联 ${users.length} 个用户：${users.slice(0, 3).join('、')}${users.length > 3 ? '...' : ''}。请先解除关联后再删除。`
    };
  }

  return { canDelete: true };
}

// 获取角色删除禁用提示
function getRoleDeleteDisabledReason(role: Api.SystemManage.Role): string {
  const { canDelete, reason } = canDeleteRole(role);
  return canDelete ? '' : (reason || '不可删除');
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

const { columns, columnChecks, data, loading, getData, getDataByPage, mobilePagination } = useUIPaginatedTable({
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
          return <span class="text-gray">-</span>;
        }

        // 最多显示3个用户，超过则显示数量
        const displayUsers = users.slice(0, 3);
        const userCount = users.length;
        const moreText = userCount > 3 ? `等${userCount}人` : '';

        return (
          <div class="flex items-center gap-4px pl-12px">
            {displayUsers.map(username => (
              <ElTag key={username} size="small" type="info">
                {username}
              </ElTag>
            ))}
            {moreText && <span class="text-gray text-12px">{moreText}</span>}
          </div>
        );
      }
    },
    {
      prop: 'status',
      label: $t('page.manage.role.roleStatus'),
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
      width: 280,
      formatter: row => (
        <div class="flex-center gap-8px">
          <ElButton type="primary" plain size="small" onClick={() => handleEdit(row.id)}>
            {$t('common.edit')}
          </ElButton>
          <ElButton type="info" plain size="small" onClick={() => handleAuth(row.id)}>
            权限设置
          </ElButton>
          <ElButton
            type="danger"
            plain
            size="small"
            onClick={() => handleDelete(row.id)}
          >
            {$t('common.delete')}
          </ElButton>
        </div>
      )
    }
  ]
});

const {
  drawerVisible,
  operateType,
  editingData,
  handleAdd,
  handleEdit,
  checkedRowKeys,
  onBatchDeleted,
  onDeleted
} = useTableOperate(data, 'id', getData);

// 处理角色操作完成后的数据刷新
async function handleRoleOperateSubmitted() {
  await getData();
  await getAllUsers();
}

// 权限设置相关
const { bool: authModalVisible, setTrue: openAuthModal } = useBoolean();
const currentRoleId = ref<number>(-1);
const currentRoleData = ref<Api.SystemManage.Role | null>(null);

function handleAuth(id: number) {
  const role = data.value.find(r => r.id === id);
  if (role) {
    currentRoleId.value = id;
    currentRoleData.value = role;
    openAuthModal();
  }
}

async function handleAuthSubmitted() {
  // 权限保存成功后，重新获取角色列表和用户列表数据
  await getData();
  await getAllUsers();
}

async function handleBatchDelete() {
  console.log('🗑️ [角色管理] 开始批量删除角色', { checkedKeys: checkedRowKeys.value });

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

  // 前置检查：检查所有选中的角色是否可以删除
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

  // 如果有不可删除的角色，显示通知
  if (cannotDeleteRoles.length > 0) {
    const message = cannotDeleteRoles
      .map(r => `• ${r.name}: ${r.reason}`)
      .join('\n');

    ElNotification({
      title: `无法删除 ${cannotDeleteRoles.length} 个角色`,
      message: message,
      type: 'warning',
      duration: 5000,
      position: 'top-right'
    });

    console.warn('⚠️ [角色管理] 以下角色不可删除:', cannotDeleteRoles);

    // 如果没有可删除的角色，直接返回
    if (canDeleteIds.length === 0) {
      return;
    }
  }

  // 批量删除（只删除可以删除的角色）
  let successCount = 0;
  let failCount = 0;
  const errorMessages: string[] = [];

  for (const id of canDeleteIds) {
    const { error } = await fetchDeleteRole(id);
    if (!error) {
      successCount++;
      console.log(`✅ [角色管理] 角色删除成功: ${id}`);
    } else {
      failCount++;
      const errorMsg = error.msg || '删除失败';
      errorMessages.push(`角色ID ${id}: ${errorMsg}`);
      console.error(`❌ [角色管理] 角色删除失败: ${id}`, error);
    }
  }

  // 显示批量删除结果通知
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
      message: `${failCount} 个角色删除失败。查看控制台了解详情`,
      type: 'error',
      duration: 5000,
      position: 'top-right'
    });
  }

  onBatchDeleted();
  // 刷新用户列表
  await getAllUsers();
}

async function handleDelete(id: number) {
  console.log('🗑️ [角色管理] 开始删除角色', { id });

  // 前置检查：获取角色信息
  const role = data.value.find(r => r.id === id);
  if (!role) {
    window.$message?.error('角色不存在');
    return;
  }

  // 前置检查：验证是否可以删除
  const { canDelete, reason } = canDeleteRole(role);
  if (!canDelete) {
    console.warn('⚠️ [角色管理] 角色不可删除', { role, reason });

    // 弹出通知提示（右上角）
    ElNotification({
      title: '无法删除角色',
      message: reason || '该角色不可删除',
      type: 'warning',
      duration: 3000,
      position: 'top-right'
    });
    return;
  }

  // 可以删除，直接执行删除（不需要确认框）
  console.log('✅ [角色管理] 前置检查通过，执行删除', { role });

  const { error } = await fetchDeleteRole(id);

  if (!error) {
    console.log('✅ [角色管理] 删除角色成功');
    ElNotification({
      title: '删除成功',
      message: `角色 "${role.name}" 已成功删除`,
      type: 'success',
      duration: 3000,
      position: 'top-right'
    });
    onDeleted();
    // 刷新用户列表
    await getAllUsers();
  } else {
    console.error('❌ [角色管理] 删除角色失败', error);
    ElNotification({
      title: '删除失败',
      message: error.msg || '删除角色失败',
      type: 'error',
      duration: 3000,
      position: 'top-right'
    });
  }
}

async function handleStatusChange(row: Api.SystemManage.Role, val: number) {
  const { error } = await fetchUpdateRole(row.id, { status: val });

  if (!error) {
    window.$message?.success(`${val === 1 ? '启用' : '禁用'}成功`);
  } else {
    // 如果更新失败，恢复原状态
    row.status = val === 1 ? 0 : 1;
    window.$message?.error('状态更新失败');
  }
}

function resetSearchParams() {
  searchParams.value = getInitSearchParams();
}

function edit(id: number) {
  handleEdit(id);
}
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <RoleSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <p>{{ $t('page.manage.role.title') }}</p>
          <TableHeaderOperation
            v-model:columns="columnChecks"
            :disabled-delete="checkedRowKeys.length === 0"
            :loading="loading"
            @add="handleAdd"
            @delete="handleBatchDelete"
            @refresh="getData"
          />
        </div>
      </template>
      <div class="h-[calc(100%-52px)]">
        <ElTable
          v-loading="loading"
          height="100%"
          border
          class="sm:h-full"
          :data="data"
          row-key="id"
          @selection-change="checkedRowKeys = $event"
        >
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
        <div class="mt-20px flex justify-end">
          <ElPagination
            v-if="mobilePagination.total"
            layout="total,prev,pager,next,sizes"
            v-bind="mobilePagination"
            @current-change="mobilePagination['current-change']"
            @size-change="mobilePagination['size-change']"
          />
        </div>
      </div>
      <RoleOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="handleRoleOperateSubmitted"
      />
      <MenuAuthModal v-model:visible="authModalVisible" :role-id="currentRoleId" :role-data="currentRoleData" @submitted="handleAuthSubmitted" />
    </ElCard>
  </div>
</template>

<style lang="scss" scoped>
:deep(.el-card) {
  .ht50 {
    height: calc(100% - 50px);
  }
}
</style>
