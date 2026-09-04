<script setup lang="tsx">
  import { ref } from 'vue';
  import { Delete, Edit, Plus } from '@element-plus/icons-vue';
  import {
    assignUserToGroup,
    deleteAuthUser,
    deleteUserGroup,
    fetchAllAuthGroups,
    fetchAuthUsers,
    fetchUserGroups,
    getAuthUserPassword
  } from '@/service/api/application-permission';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import PermissionButton from '@/components/common/PermissionButton.vue';
  import UserSearch from './modules/user-search.vue';
  import UserOperateDrawer from './modules/user-operate-drawer.vue';

  defineOptions({ name: 'AuthCenterUsers' });

  interface SearchParams {
    page: number;
    pageSize: number;
    username: string;
    nickname: string;
    email: string;
    phone: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 10, username: '', nickname: '', email: '', phone: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchAuthUsers(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 10;
    },
    columns: () => [
      { prop: 'index', type: 'index', label: '序号', width: 60, align: 'center' },
      { prop: 'username', label: '用户名', align: 'center', minWidth: 120 },
      { prop: 'nickname', label: '昵称', align: 'center', minWidth: 120 },
      { prop: 'email', label: '邮箱', align: 'center', minWidth: 150 },
      { prop: 'phone', label: '电话', align: 'center', minWidth: 120 },
      {
        prop: 'groups',
        label: '用户组',
        align: 'center',
        minWidth: 200,
        formatter: row => {
          const groups = row.groups ?? [];
          return (
            <div class="flex items-center justify-center gap-2">
              {groups.length > 0 ? (
                groups.map(g => (
                  <ElTag key={g.id} type="primary" size="small">
                    {g.name}
                  </ElTag>
                ))
              ) : (
                <span class="text-gray-400">未分配</span>
              )}
              <PermissionButton
                code="auth.user.update"
                size="small"
                type="primary"
                link
                onClick={() => handleManageGroups(row)}
              >
                {groups.length > 0 ? '管理' : '分配'}
              </PermissionButton>
            </div>
          );
        }
      },
      {
        prop: 'password',
        label: '初始密码',
        align: 'center',
        width: 140,
        formatter: row => (
          <PermissionButton
            code="auth.user.update"
            size="small"
            type="primary"
            link
            onClick={() => handleViewPassword(row)}
          >
            查看密码
          </PermissionButton>
        )
      },
      {
        prop: 'status',
        label: '状态',
        align: 'center',
        width: 80,
        formatter: row => (
          <ElTag type={row.status === 1 ? 'success' : 'info'}>{row.status === 1 ? '启用' : '禁用'}</ElTag>
        )
      },
      {
        prop: 'operate',
        label: '操作',
        align: 'center',
        width: 200,
        fixed: 'right',
        formatter: row => (
          <ElSpace>
            <ElButton size="small" type="primary" link onClick={() => handleManageGroups(row)}>
              用户组
            </ElButton>
            <PermissionButton
              code="auth.user.update"
              size="small"
              type="warning"
              link
              icon={Edit}
              onClick={() => handleEdit(row.id)}
            >
              编辑
            </PermissionButton>
            <ElPopconfirm title="确认删除该用户？" onConfirm={() => handleDelete(row.id)}>
              {{
                reference: () => (
                  <PermissionButton code="auth.user.delete" size="small" type="danger" link icon={Delete}>
                    删除
                  </PermissionButton>
                )
              }}
            </ElPopconfirm>
          </ElSpace>
        )
      }
    ]
  });

  const { drawerVisible, operateType, editingData, handleAdd, handleEdit, onDeleted } = useTableOperate(
    data,
    'id',
    getData
  );

  async function handleDelete(id: number) {
    const { error } = await deleteAuthUser(id);
    if (!error) {
      await onDeleted();
    }
  }

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
  }

  // 用户组管理
  const groupDialogVisible = ref(false);
  const selectedUser = ref<Api.ApplicationPermission.AuthUser | null>(null);
  const userGroups = ref<Api.ApplicationPermission.AuthUserGroup[]>([]);
  const allGroups = ref<Api.ApplicationPermission.AuthGroup[]>([]);
  const groupFormData = ref({ groupId: null as number | null });

  async function handleManageGroups(row: Api.ApplicationPermission.AuthUser) {
    selectedUser.value = row;
    groupDialogVisible.value = true;

    const { data: allGroupsData } = await fetchAllAuthGroups();
    if (allGroupsData) {
      allGroups.value = Array.isArray(allGroupsData)
        ? allGroupsData
        : (allGroupsData as { list?: Api.ApplicationPermission.AuthGroup[] }).list || [];
    }

    const { data: currentUserGroups } = await fetchUserGroups(row.id);
    if (currentUserGroups) {
      userGroups.value = currentUserGroups;
    }
  }

  async function handleAssignGroup() {
    if (!selectedUser.value || !groupFormData.value.groupId) {
      ElMessage.warning('请选择用户组');
      return;
    }

    const { error } = await assignUserToGroup({
      userId: selectedUser.value.id,
      groupId: groupFormData.value.groupId
    });

    if (!error) {
      ElMessage.success('分配用户组成功');
      const { data: updatedUserGroups } = await fetchUserGroups(selectedUser.value.id);
      if (updatedUserGroups) {
        userGroups.value = updatedUserGroups;
      }
    }
  }

  async function handleRemoveGroup(groupId: number) {
    if (!selectedUser.value) return;
    const { error } = await deleteUserGroup(selectedUser.value.id, groupId);
    if (!error) {
      ElMessage.success('移除用户组成功');
      const { data: updatedUserGroups } = await fetchUserGroups(selectedUser.value.id);
      if (updatedUserGroups) {
        userGroups.value = updatedUserGroups;
      }
    }
  }

  async function handleViewPassword(row: Api.ApplicationPermission.AuthUser) {
    const { data, error } = await getAuthUserPassword(row.id);
    if (!error && data) {
      // 复用 drawer 内部的密码对话框：临时构造 created password 并通过 drawer 显示
      ElMessageBox.alert(
        `<div style="font-size:16px;line-height:1.6">
          <p><strong>用户名：</strong>${data.username}</p>
          <p><strong>初始密码：</strong><code style="font-size:18px;color:var(--el-color-primary);font-weight:bold">${data.password}</code></p>
          <p style="color:var(--el-color-warning);margin-top:12px">此密码仅显示一次，请妥善保管</p>
        </div>`,
        '用户初始密码',
        { dangerouslyUseHTMLString: true, confirmButtonText: '我已保存，关闭' }
      );
    }
  }
</script>

<template>
  <div class="table-page">
    <!-- 搜索区域 -->
    <UserSearch v-model:model="searchParams" @reset="handleReset" @search="handleSearch" />

    <!-- 表格卡片 -->
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-medium">授权中心用户列表</span>
          <PermissionButton code="auth.user.create" type="primary" :icon="Plus" @click="handleAdd">
            添加用户
          </PermissionButton>
        </div>
      </template>

      <div class="table-scroll-wrap">
        <ElTable v-loading="loading" height="100%" :data="data" :border="false" row-key="id">
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>

      <div class="mt-20px flex justify-end">
        <ElPagination
          v-if="mobilePagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          v-bind="mobilePagination"
          :page-sizes="[10, 20, 50, 100]"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>

      <!-- 添加/编辑抽屉 -->
      <UserOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </ElCard>

    <!-- 用户组管理对话框 -->
    <ElDialog v-model="groupDialogVisible" :title="`${selectedUser?.username || ''} 的用户组管理`" width="600px">
      <div class="mb-4">
        <div class="mb-4 flex items-center gap-2">
          <ElText type="primary">分配用户组：</ElText>
          <ElSelect v-model="groupFormData.groupId" placeholder="选择用户组" style="width: 200px">
            <ElOption
              v-for="group in allGroups"
              :key="group.id"
              :label="`${group.name} (${group.code})`"
              :value="group.id"
            />
          </ElSelect>
          <PermissionButton code="auth.user.update" type="primary" size="small" @click="handleAssignGroup">
            分配
          </PermissionButton>
        </div>
      </div>

      <ElDivider content-position="left">已分配的用户组</ElDivider>

      <ElTable :data="userGroups" :border="false" max-height="400px">
        <ElTableColumn prop="groupName" label="用户组名称" align="center" min-width="150" />
        <ElTableColumn prop="groupCode" label="用户组代码" align="center" min-width="150" />
        <ElTableColumn prop="grantedBy" label="分配人" align="center" min-width="100" />
        <ElTableColumn label="操作" align="center" width="100">
          <template #default="{ row }">
            <PermissionButton code="auth.user.update" size="small" type="danger" @click="handleRemoveGroup(row.id)">
              移除
            </PermissionButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <ElEmpty v-if="userGroups.length === 0" description="暂无用户组" />

      <template #footer>
        <ElButton @click="groupDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
