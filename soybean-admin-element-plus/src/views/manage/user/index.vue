<script setup lang="tsx">
import { ref, computed, onMounted } from 'vue';
import { fetchGetUserList, fetchGetAllRoles, fetchDeleteUser } from '@/service/api';
import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
import { $t } from '@/locales';
import UserOperateDrawer from './modules/user-operate-drawer.vue';
import UserSearch from './modules/user-search.vue';
import ResetPasswordModal from './modules/reset-password-modal.vue';

defineOptions({ name: 'UserManage' });

// 角色列表和映射
const allRoles = ref<Api.SystemManage.AllRole[]>([]);
const roleMap = computed(() => {
  const map = new Map<number, Api.SystemManage.AllRole>();
  allRoles.value.forEach(role => {
    map.set(role.id, role);
  });
  return map;
});

// 预定义的颜色列表（涵盖多种颜色，确保不同角色编码显示不同颜色）
// 包含 Element Plus 主题色 + 自定义颜色
const COLOR_PALETTE: Array<{ type: UI.ThemeColor; customClass?: string }> = [
  { type: 'primary' },     // 蓝色
  { type: 'success' },     // 绿色
  { type: 'warning' },     // 橙色
  { type: 'danger' },      // 红色
  { type: 'info' }         // 灰色
];

// 字符串哈希函数
function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // 转换为32位整数
  }
  return Math.abs(hash);
}

// 根据角色代码获取标签颜色（每个不同编码显示不同颜色）
function getRoleTagType(roleCode: string): UI.ThemeColor {
  // 计算角色编码的哈希值
  const hash = stringHash(roleCode);

  // 根据哈希值选择颜色，确保不同编码大概率显示不同颜色
  const colorIndex = hash % COLOR_PALETTE.length;

  return COLOR_PALETTE[colorIndex].type;
}

// 获取角色颜色的自定义样式类（可选，用于更丰富的颜色）
function getRoleTagClass(roleCode: string): string {
  const hash = stringHash(roleCode);

  // 可以根据哈希值添加自定义样式
  return `role-tag-${hash % 10}`; // 10种自定义样式变体
}

// 获取所有角色
async function getAllRoles() {
  const { error, data } = await fetchGetAllRoles();
  if (!error && data) {
    allRoles.value = data;
  }
}

onMounted(() => {
  getAllRoles();
});

const searchParams = ref(getInitSearchParams());

function getInitSearchParams(): Api.SystemManage.UserSearchParams {
  return {
    current: 1,
    size: 10,
    status: undefined,
    username: undefined,
    nickname: undefined,
    email: undefined
  };
}

const { columns, columnChecks, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
  paginationProps: {
    currentPage: searchParams.value.current,
    pageSize: searchParams.value.size
  },
  api: () => fetchGetUserList(searchParams.value),
  transform: response => {
    return defaultTransform(response);
  },
  onPaginationParamsChange: params => {
    searchParams.value.current = params.currentPage;
    searchParams.value.size = params.pageSize;
  },
  columns: () => [
    { prop: 'selection', type: 'selection', width: 48 },
    { prop: 'index', type: 'index', label: $t('common.index'), width: 64 },
    { prop: 'username', label: $t('page.manage.user.userName'), minWidth: 100 },
    { prop: 'nickname', label: $t('page.manage.user.nickName'), minWidth: 100 },
    { prop: 'email', label: $t('page.manage.user.userEmail'), minWidth: 200 },
    {
      prop: 'roleIds',
      label: '分配角色',
      minWidth: 150,
      formatter: row => {
        if (!row.roleIds || row.roleIds.length === 0) {
          return <span class="text-gray pl-12px">-</span>;
        }

        // 将 roleIds 数组转换为角色标签
        const roleTags = row.roleIds
          .map((id: number) => {
            const role = roleMap.value.get(id);
            if (!role) return null;

            // 根据角色编码（code）获取标签颜色，更严谨
            const tagType = getRoleTagType(role.code);

            return (
              <ElTag key={id} size="small" type={tagType}>
                {role.name}
              </ElTag>
            );
          })
          .filter(Boolean);

        if (roleTags.length === 0) {
          return <span class="text-gray pl-12px">-</span>;
        }

        return <div class="flex items-center gap-4px flex-wrap pl-12px">{roleTags}</div>;
      }
    },
    {
      prop: 'status',
      label: $t('page.manage.user.userStatus'),
      align: 'center',
      width: 100,
      formatter: row => {
        if (row.status === undefined) {
          return '';
        }

        const statusMap: Record<string, UI.ThemeColor> = {
          active: 'success',
          inactive: 'warning'
        };

        const label = row.status === 'active' ? '启用' : '禁用';

        return <ElTag type={statusMap[row.status] || 'info'}>{label}</ElTag>;
      }
    },
    {
      prop: 'homePath',
      label: '家目录',
      align: 'center',
      width: 150,
      formatter: row => {
        const path = row.homePath || '/';
        const pathMap: Record<string, string> = {
          '/': '首页',
          '/servers': '资产管理',
          '/tasks': '自动化任务',
          '/monitoring': '监控中心',
          '/system': '系统管理'
        };
        return <span class="text-primary">{pathMap[path] || path}</span>;
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
          <ElButton type="warning" plain size="small" onClick={() => openResetPassword(row)}>
            重置密码
          </ElButton>
          <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handleDelete(row.id)}>
            {{
              reference: () => (
                <ElButton type="danger" plain size="small" disabled={row.username === 'admin'}>
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

// 重置密码相关
const resetPasswordVisible = ref(false);
const resetPasswordUserId = ref(-1);
const resetPasswordUsername = ref('');

function openResetPassword(row: Api.SystemManage.User) {
  resetPasswordUserId.value = row.id;
  resetPasswordUsername.value = row.username || '';
  resetPasswordVisible.value = true;
}

function handleResetPasswordSubmitted() {
  // 重置密码成功后刷新列表
  getDataByPage();
}

async function handleBatchDelete() {
  if (checkedRowKeys.value.length === 0) {
    window.$message?.warning('请选择要删除的用户');
    return;
  }

  // 批量删除（串行执行）
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
  // 刷新角色列表
  await getAllRoles();
}

async function handleDelete(id: number) {
  const { error } = await fetchDeleteUser(id);

  if (!error) {
    window.$message?.success($t('common.deleteSuccess'));
    onDeleted();
    // 刷新角色列表
    await getAllRoles();
  } else {
    window.$message?.error(error.msg || '删除失败');
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
    <UserSearch v-model:model="searchParams" @reset="resetSearchParams" @search="getDataByPage" />
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <p>{{ $t('page.manage.user.title') }}</p>
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
      </div>
      <div class="mt-20px flex justify-end">
        <ElPagination
          v-if="mobilePagination.total"
          layout="total,prev,pager,next,sizes"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
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
