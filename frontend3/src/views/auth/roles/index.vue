<script setup lang="tsx">
  import { ref } from 'vue';
  import { Delete, Edit, Plus } from '@element-plus/icons-vue';
  import { deleteAuthGroup, fetchAuthGroups } from '@/service/api/application-permission';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import RoleSearch from './modules/role-search.vue';
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
        width: 150,
        fixed: 'right',
        formatter: row => (
          <ElSpace>
            <ElButton size="small" type="warning" icon={Edit} onClick={() => handleEdit(row.id)}>
              编辑
            </ElButton>
            <ElPopconfirm title="确认删除该用户组？" onConfirm={() => handleDelete(row.id)}>
              {{
                reference: () => (
                  <ElButton size="small" type="danger" icon={Delete}>
                    删除
                  </ElButton>
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
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 搜索区域 -->
    <RoleSearch v-model:model="searchParams" @reset="handleReset" @search="handleSearch" />

    <!-- 表格卡片 -->
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-medium">授权中心用户组列表</span>
          <ElButton type="primary" :icon="Plus" @click="handleAdd">添加用户组</ElButton>
        </div>
      </template>

      <div class="h-[calc(100%-52px)]">
        <ElTable v-loading="loading" height="100%" :data="data" :border="false" class="sm:h-full" row-key="id">
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
      <RoleOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </ElCard>
  </div>
</template>
