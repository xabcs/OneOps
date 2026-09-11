<script setup lang="tsx">
  import { ref } from 'vue';
  import { Plus } from '@element-plus/icons-vue';
  import { deleteTicketType, fetchTicketTypes } from '@/service/api';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import PermissionButton from '@/components/common/PermissionButton.vue';
  import TypeOperateDialog from './modules/type-operate-dialog.vue';

  defineOptions({ name: 'TicketTypes' });

  interface SearchParams {
    page: number;
    pageSize: number;
    keyword: string;
    status: number;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 10, keyword: '', status: -1 };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchTicketTypes(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 10;
    },
    columns: () => [
      {
        prop: 'name',
        label: '类型名称',
        minWidth: 140,
        align: 'center',
        formatter: row => (
          <div class="flex items-center justify-center gap-6px">
            {row.icon ? <SvgIcon icon={row.icon} class="h-18px w-18px text-primary" /> : null}
            <span>{row.name}</span>
          </div>
        )
      },
      { prop: 'code', label: '类型编码', width: 130, align: 'center' },
      { prop: 'description', label: '描述', minWidth: 180, align: 'center', showOverflowTooltip: true },
      {
        prop: 'workflowCount',
        label: '审批流程',
        width: 110,
        align: 'center',
        formatter: row =>
          row.status !== 1 ? (
            ''
          ) : row.workflowCount > 0 ? (
            <ElTag type="success">{row.workflowCount} 个</ElTag>
          ) : (
            <ElTooltip content="该场景未配置审批流程，发起工单时将无流程可选，请前往「审批流程」新建并选择所属场景">
              {{
                default: () => <ElTag type="danger">未配置</ElTag>
              }}
            </ElTooltip>
          )
      },
      {
        prop: 'status',
        label: '状态',
        width: 80,
        align: 'center',
        formatter: row => (
          <ElTag type={row.status === 1 ? 'success' : 'info'}>{row.status === 1 ? '启用' : '禁用'}</ElTag>
        )
      },
      { prop: 'createdAt', label: '创建时间', width: 170, align: 'center' },
      {
        prop: 'operate',
        label: '操作',
        width: 150,
        fixed: 'right',
        align: 'center',
        className: 'msre-table-actions',
        formatter: row => (
          <div>
            <PermissionButton
              code="ticket.type.update"
              size="small"
              link
              type="primary"
              onClick={() => handleEdit(row.id)}
            >
              编辑
            </PermissionButton>
            <ElPopconfirm title="确认删除该工单类型？" onConfirm={() => handleDelete(row.id)}>
              {{
                reference: () => (
                  <PermissionButton code="ticket.type.delete" link type="danger" size="small">
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
    const { error } = await deleteTicketType(id);
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
    title="工单类型管理"
    description="维护工单场景类型与启停状态，支撑审批流程绑定"
    :pagination="mobilePagination"
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElInput
        v-model="searchParams.keyword"
        placeholder="名称/编码"
        clearable
        class="w-200px"
        @keyup.enter="handleSearch"
      />
      <ElSelect v-model="searchParams.status" class="w-120px">
        <ElOption label="全部" :value="-1" />
        <ElOption label="启用" :value="1" />
        <ElOption label="禁用" :value="0" />
      </ElSelect>
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <PermissionButton code="ticket.type.create" type="primary" :icon="Plus" @click="handleAdd">
        新建类型
      </PermissionButton>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" height="100%" :data="data" :border="false" row-key="id">
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>
    <!-- 新建/编辑弹框（teleport 弹层须置于布局内，保持页面单根节点以正常继承 attrs 与 Transition） -->
    <TypeOperateDialog
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="getDataByPage"
    />
  </ListPageLayout>
</template>
