<script setup lang="tsx">
  import { ref } from 'vue';
  import { ArrowDown, Plus } from '@element-plus/icons-vue';
  import {
    deleteApplication,
    fetchApplicationAuthorizationRules,
    fetchApplicationGroups,
    fetchApplicationRoles,
    fetchApplicationUsers,
    fetchApplications,
    syncApplicationAuthorizationRules,
    syncApplicationGroups,
    syncApplicationRoles,
    syncApplicationUsers
  } from '@/service/api/application-permission';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import ApplicationSearch from './modules/application-search.vue';
  import ApplicationOperateDrawer from './modules/application-operate-drawer.vue';

  defineOptions({ name: 'AuthCenterApplications' });

  interface SearchParams {
    page: number;
    pageSize: number;
    name?: string;
    code?: string;
    type?: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 10, name: undefined, code: undefined, type: undefined };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () =>
      fetchApplications({
        page: searchParams.value.page,
        pageSize: searchParams.value.pageSize,
        name: searchParams.value.name || undefined,
        code: searchParams.value.code || undefined,
        type: searchParams.value.type || undefined
      }),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 10;
    },
    columns: () => [
      { prop: 'index', type: 'index', label: '序号', width: 60, align: 'center' },
      { prop: 'name', label: '应用名称', align: 'center', minWidth: 120 },
      { prop: 'code', label: '应用代码', align: 'center', minWidth: 100 },
      { prop: 'type', label: '应用类型', align: 'center', minWidth: 100 },
      { prop: 'baseUrl', label: 'Base URL', align: 'center', minWidth: 200 },
      {
        prop: 'status',
        label: '状态',
        align: 'center',
        width: 80,
        formatter: row => (
          <ElTag type={row.status === 1 ? 'success' : 'info'}>{row.status === 1 ? '启用' : '禁用'}</ElTag>
        )
      },
      { prop: 'lastSyncTime', label: '最后同步时间', align: 'center', minWidth: 160 },
      {
        prop: 'operate',
        label: '操作',
        align: 'center',
        width: 320,
        fixed: 'right',
        formatter: row => (
          <ElSpace wrap>
            <ElButton size="small" type="primary" onClick={() => handleSyncUsers(row)}>
              同步用户
            </ElButton>
            <ElButton size="small" type="success" onClick={() => handleSyncGroups(row)}>
              同步用户组
            </ElButton>
            {row.type === 'jumpserver' && (
              <ElButton size="small" type="warning" onClick={() => handleSyncRules(row)}>
                同步授权规则
              </ElButton>
            )}
            <ElDropdown onCommand={(cmd: string) => handleCommand(cmd, row)}>
              <ElButton size="small">
                更多
                <ElIcon class="el-icon--right">
                  <ArrowDown />
                </ElIcon>
              </ElButton>
              {{
                dropdown: () => (
                  <ElDropdownMenu>
                    <ElDropdownItem command="viewUsers">查看用户</ElDropdownItem>
                    <ElDropdownItem command="viewGroups">查看用户组</ElDropdownItem>
                    {row.type === 'jumpserver' && <ElDropdownItem command="viewRules">查看授权规则</ElDropdownItem>}
                    {row.type !== 'jumpserver' && <ElDropdownItem command="viewRoles">查看角色</ElDropdownItem>}
                    {row.type !== 'jumpserver' && <ElDropdownItem command="syncRoles">同步角色</ElDropdownItem>}
                    <ElDropdownItem divided command="edit">
                      编辑
                    </ElDropdownItem>
                    <ElDropdownItem command="delete" style="color: #f56c6c">
                      删除
                    </ElDropdownItem>
                  </ElDropdownMenu>
                )
              }}
            </ElDropdown>
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
    const { error } = await deleteApplication(id);
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

  function handleCommand(command: string, row: Api.ApplicationPermission.Application) {
    switch (command) {
      case 'viewUsers':
        handleViewUsers(row);
        break;
      case 'viewGroups':
        handleViewGroups(row);
        break;
      case 'viewRules':
        handleViewRules(row);
        break;
      case 'viewRoles':
        handleViewRoles(row);
        break;
      case 'syncRoles':
        handleSyncRoles(row);
        break;
      case 'edit':
        handleEdit(row.id);
        break;
      case 'delete':
        handleDelete(row.id);
        break;
    }
  }

  // ===== 同步功能 =====
  async function handleSyncRoles(row: Api.ApplicationPermission.Application) {
    const { error } = await syncApplicationRoles(row.id);
    if (!error) {
      ElMessage.success(`同步 ${row.name} 角色成功`);
      await getData();
      handleViewRoles(row);
    }
  }

  async function handleSyncUsers(row: Api.ApplicationPermission.Application) {
    const { error } = await syncApplicationUsers(row.id);
    if (!error) {
      ElMessage.success(`同步 ${row.name} 用户成功`);
      await getData();
      handleViewUsers(row);
    }
  }

  async function handleSyncGroups(row: Api.ApplicationPermission.Application) {
    const { error } = await syncApplicationGroups(row.id);
    if (!error) {
      ElMessage.success(`同步 ${row.name} 用户组成功`);
      await getData();
      handleViewGroups(row);
    }
  }

  async function handleSyncRules(row: Api.ApplicationPermission.Application) {
    const { error } = await syncApplicationAuthorizationRules(row.id);
    if (!error) {
      ElMessage.success(`同步 ${row.name} 授权规则成功`);
      await getData();
      handleViewRules(row);
    }
  }

  // ===== 查看对话框 =====
  const rolesDialogVisible = ref(false);
  const usersDialogVisible = ref(false);
  const groupsDialogVisible = ref(false);
  const rulesDialogVisible = ref(false);
  const applicationRoles = ref<Api.ApplicationPermission.ApplicationRole[]>([]);
  const applicationUsers = ref<Api.ApplicationPermission.ApplicationUser[]>([]);
  const applicationGroups = ref<Api.ApplicationPermission.ApplicationGroup[]>([]);
  const authorizationRules = ref<Api.ApplicationPermission.AuthorizationRule[]>([]);
  const currentAppName = ref('');

  async function handleViewRoles(row: Api.ApplicationPermission.Application) {
    currentAppName.value = row.name;
    const { data, error } = await fetchApplicationRoles(row.id);
    if (!error && data) {
      applicationRoles.value = data;
      rolesDialogVisible.value = true;
    } else {
      ElMessage.warning('暂无角色数据，请先同步角色');
    }
  }

  async function handleViewUsers(row: Api.ApplicationPermission.Application) {
    currentAppName.value = row.name;
    const { data, error } = await fetchApplicationUsers(row.id);
    if (!error && data) {
      applicationUsers.value = data;
      usersDialogVisible.value = true;
    } else {
      ElMessage.warning('暂无用户数据，请先同步用户');
    }
  }

  async function handleViewGroups(row: Api.ApplicationPermission.Application) {
    currentAppName.value = row.name;
    const { data, error } = await fetchApplicationGroups(row.id);
    if (!error && data) {
      if (data.length === 0) {
        ElMessage.info('该应用暂无用户组数据，可能该版本不支持用户组功能或端点配置错误');
      }
      applicationGroups.value = data;
      groupsDialogVisible.value = true;
    } else {
      ElMessage.warning('暂无用户组数据，请先同步用户组');
    }
  }

  async function handleViewRules(row: Api.ApplicationPermission.Application) {
    currentAppName.value = row.name;
    const { data, error } = await fetchApplicationAuthorizationRules(row.id);
    if (!error && data) {
      if (data.length === 0) {
        ElMessage.info('该应用暂无授权规则数据，可能该应用不支持授权规则或端点配置错误');
      }
      authorizationRules.value = data;
      rulesDialogVisible.value = true;
    } else {
      ElMessage.warning('暂无授权规则数据，请先同步授权规则');
    }
  }

  // 授权规则操作权限辅助方法
  function getActionLabel(action: string): string {
    const actionLabels: Record<string, string> = {
      connect: '连接',
      upload: '上传',
      download: '下载',
      command: '命令',
      all: '全部'
    };
    return actionLabels[action] || action;
  }

  function getActionType(action: string): 'success' | 'warning' | 'info' | 'danger' | 'primary' {
    const actionTypes: Record<string, 'success' | 'warning' | 'info' | 'danger' | 'primary'> = {
      connect: 'primary',
      upload: 'warning',
      download: 'info',
      command: 'success',
      all: 'danger'
    };
    return actionTypes[action] || 'primary';
  }
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 搜索区域 -->
    <ApplicationSearch v-model:model="searchParams" @reset="handleReset" @search="handleSearch" />

    <!-- 表格卡片 -->
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-medium">应用列表</span>
          <ElButton type="primary" :icon="Plus" @click="handleAdd">添加应用</ElButton>
        </div>
      </template>

      <div class="h-[calc(100%-52px)]">
        <ElTable v-loading="loading" height="100%" :data="data" :border="false" row-key="id">
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>

        <div class="mt-20px flex justify-end">
          <ElPagination
            v-if="mobilePagination.total"
            layout="total, sizes, prev, pager, next, jumper"
            v-bind="mobilePagination"
            @current-change="mobilePagination['current-change']"
            @size-change="mobilePagination['size-change']"
          />
        </div>
      </div>

      <!-- 添加/编辑抽屉 -->
      <ApplicationOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </ElCard>

    <!-- 角色列表对话框 -->
    <ElDialog v-model="rolesDialogVisible" :title="`${currentAppName} - 角色列表`" width="800px">
      <ElTable :data="applicationRoles" :border="true">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn prop="roleCode" label="角色代码" align="center" min-width="120" />
        <ElTableColumn prop="roleName" label="角色名称" align="center" min-width="120" />
        <ElTableColumn prop="roleType" label="角色类型" align="center" min-width="100">
          <template #default="{ row }">
            <ElTag>{{ row.roleType || 'global' }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="描述" align="center" min-width="150" />
        <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
      </ElTable>
      <template #footer>
        <ElButton type="primary" @click="rolesDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>

    <!-- 用户列表对话框 -->
    <ElDialog v-model="usersDialogVisible" :title="`${currentAppName} - 用户列表`" width="800px">
      <ElTable :data="applicationUsers" :border="true">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn prop="username" label="用户名" align="center" min-width="120" />
        <ElTableColumn prop="displayName" label="显示名称" align="center" min-width="120" />
        <ElTableColumn prop="email" label="邮箱" align="center" min-width="150" />
        <ElTableColumn prop="status" label="状态" align="center" width="80">
          <template #default="{ row }">
            <ElTag :type="row.status === 'active' ? 'success' : 'info'">{{ row.status || 'active' }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
      </ElTable>
      <template #footer>
        <ElButton type="primary" @click="usersDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>

    <!-- 用户组列表对话框 -->
    <ElDialog v-model="groupsDialogVisible" :title="`${currentAppName} - 用户组列表`" width="800px">
      <ElTable :data="applicationGroups" :border="true">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn prop="groupCode" label="用户组代码" align="center" min-width="120" />
        <ElTableColumn prop="groupName" label="用户组名称" align="center" min-width="120" />
        <ElTableColumn prop="description" label="描述" align="center" min-width="200" />
        <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
      </ElTable>
      <template #footer>
        <ElButton type="primary" @click="groupsDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>

    <!-- 授权规则列表对话框 -->
    <ElDialog v-model="rulesDialogVisible" :title="`${currentAppName} - 授权规则列表`" width="1200px">
      <ElAlert title="授权规则说明" type="info" :closable="false" style="margin-bottom: 16px">
        <p>授权规则定义了用户/用户组对资产的访问权限，包括：</p>
        <ul style="margin: 8px 0; padding-left: 20px">
          <li><strong>主体</strong>：谁可以访问（用户或用户组）</li>
          <li><strong>对象</strong>：可以访问什么（具体资产或全部资产）</li>
          <li><strong>权限</strong>：可以执行的操作（连接、上传、下载、命令等）</li>
        </ul>
      </ElAlert>

      <ElTable :data="authorizationRules" :border="true" style="max-height: 500px; overflow-y: auto">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn prop="ruleName" label="规则名称" align="center" min-width="150" />
        <ElTableColumn prop="subjectType" label="主体类型" align="center" width="100">
          <template #default="{ row }">
            <ElTag :type="row.subjectType === 'user' ? 'success' : 'warning'" size="small">
              {{ row.subjectType === 'user' ? '用户' : '用户组' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="subjectName" label="主体名称" align="center" min-width="120" />
        <ElTableColumn prop="objectType" label="对象类型" align="center" width="100">
          <template #default="{ row }">
            <ElTag :type="row.objectType === 'system' ? 'danger' : 'primary'" size="small">
              {{ row.objectType === 'system' ? '全部资产' : '具体资产' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="objectName" label="对象名称" align="center" min-width="150" />
        <ElTableColumn prop="actions" label="操作权限" align="center" min-width="200">
          <template #default="{ row }">
            <div v-if="row.actions">
              <ElTag
                v-for="action in JSON.parse(row.actions)"
                :key="action"
                size="small"
                style="margin: 2px"
                :type="getActionType(action)"
              >
                {{ getActionLabel(action) }}
              </ElTag>
            </div>
            <span v-else style="color: #909399">无权限数据</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="priority" label="优先级" align="center" width="80" />
        <ElTableColumn prop="isEnabled" label="状态" align="center" width="80">
          <template #default="{ row }">
            <ElTag :type="row.isEnabled ? 'success' : 'info'" size="small">
              {{ row.isEnabled ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="isExpired" label="是否过期" align="center" width="80">
          <template #default="{ row }">
            <ElTag :type="row.isExpired ? 'danger' : 'success'" size="small">
              {{ row.isExpired ? '已过期' : '有效' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="syncTime" label="同步时间" align="center" min-width="160" />
      </ElTable>
      <template #footer>
        <ElButton type="primary" @click="rulesDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>
  </div>
</template>
