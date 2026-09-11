<script setup lang="tsx">
  import { ref } from 'vue';
  import { Plus } from '@element-plus/icons-vue';
  import { deleteAuthGroup, fetchAuthGroups } from '@/service/api/application-permission';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import PermissionButton from '@/components/common/PermissionButton.vue';
  import RoleOperateDrawer from './modules/role-operate-drawer.vue';

  defineOptions({ name: 'AuthCenterGroups' });

  interface SearchParams {
    page: number;
    pageSize: number;
    name: string;
    code: string;
    description: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 10, name: '', code: '', description: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchAuthGroups(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 10;
    },
    columns: () => [
      { prop: 'index', type: 'index', label: '序号', width: 60, align: 'center' },
      { prop: 'name', label: '用户组名称', align: 'center', minWidth: 120 },
      { prop: 'code', label: '用户组代码', align: 'center', minWidth: 120 },
      { prop: 'description', label: '描述', align: 'center', minWidth: 200 },
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
        className: 'msre-table-actions',
        width: 150,
        fixed: 'right',
        formatter: row => (
          <div>
            <PermissionButton
              code="auth.group.update"
              size="small"
              link
              type="primary"
              onClick={() => handleEdit(row.id)}
            >
              编辑
            </PermissionButton>
            <ElPopconfirm title="确认删除该用户组？" onConfirm={() => handleDelete(row.id)}>
              {{
                reference: () => (
                  <PermissionButton code="auth.group.delete" link type="danger" size="small">
                    删除
                  </PermissionButton>
                )
              }}
            </ElPopconfirm>
          </div>
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
    const { error } = await deleteAuthGroup(id);
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
</script>

<template>
  <ListPageLayout
    title="授权中心用户组列表"
    description="管理授权中心用户组的基本信息与启用状态"
    :pagination="mobilePagination"
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElInput v-model="searchParams.name" placeholder="请输入角色名称" clearable class="w-200px" />
      <ElInput v-model="searchParams.code" placeholder="请输入角色编码" clearable class="w-200px" />
      <ElInput v-model="searchParams.description" placeholder="请输入描述" clearable class="w-200px" />
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <PermissionButton code="auth.group.create" type="primary" :icon="Plus" @click="handleAdd">
        添加用户组
      </PermissionButton>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" height="100%" :data="data" :border="false" row-key="id">
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>
    <!-- 添加/编辑抽屉（teleport 弹层须置于布局内，保持页面单根节点以正常继承 attrs 与 Transition） -->
    <RoleOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="getDataByPage"
    />
  </ListPageLayout>
</template>
