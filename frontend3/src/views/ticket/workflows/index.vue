<script setup lang="tsx">
  import { ref } from 'vue';
  import { Delete, Edit, Plus } from '@element-plus/icons-vue';
  import { deleteWorkflow, fetchTicketTypeOptions, fetchWorkflows } from '@/service/api';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import PermissionButton from '@/components/common/PermissionButton.vue';
  import WorkflowOperateDialog from './modules/workflow-operate-dialog.vue';

  defineOptions({ name: 'TicketWorkflows' });

  interface SearchParams {
    page: number;
    pageSize: number;
    keyword: string;
    status: number;
    typeId: number;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 10, keyword: '', status: -1, typeId: -1 };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  // 场景选项：筛选与列表展示场景名
  const typeOptions = ref<Api.Ticket.TicketTypeOption[]>([]);
  const typeNameMap = ref<Record<number, string>>({});
  fetchTicketTypeOptions().then(({ data, error }) => {
    if (!error && data) {
      typeOptions.value = data;
      const map: Record<number, string> = {};
      data.forEach(t => {
        map[t.id] = t.name;
      });
      typeNameMap.value = map;
    }
  });

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () =>
      fetchWorkflows({
        ...searchParams.value,
        typeId: searchParams.value.typeId > 0 ? searchParams.value.typeId : undefined
      }),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 10;
    },
    columns: () => [
      { prop: 'name', label: '流程名称', minWidth: 150, align: 'center' },
      {
        prop: 'typeId',
        label: '工单类型',
        width: 120,
        align: 'center',
        formatter: row => typeNameMap.value[row.typeId] || '-'
      },
      { prop: 'nodeCount', label: '审批节点数', width: 100, align: 'center' },
      { prop: 'code', label: '流程编码', width: 140, align: 'center' },
      { prop: 'description', label: '描述', minWidth: 200, align: 'center', showOverflowTooltip: true },
      { prop: 'version', label: '版本', width: 70, align: 'center' },
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
        formatter: row => (
          <ElSpace>
            <PermissionButton
              code="ticket.workflow.update"
              size="small"
              type="warning"
              icon={Edit}
              onClick={() => handleEdit(row.id)}
            >
              编辑
            </PermissionButton>
            <ElPopconfirm title="确认删除该流程？已有工单记录时无法删除" onConfirm={() => handleDelete(row.id)}>
              {{
                reference: () => (
                  <PermissionButton code="ticket.workflow.delete" size="small" type="danger" icon={Delete}>
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
    const { error } = await deleteWorkflow(id);
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
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-medium">审批流程</span>
          <PermissionButton code="ticket.workflow.create" type="primary" :icon="Plus" @click="handleAdd">
            新建流程
          </PermissionButton>
        </div>
      </template>

      <!-- 搜索区 -->
      <ElForm inline class="mb-12px" @submit.prevent>
        <ElFormItem label="关键词">
          <ElInput
            v-model="searchParams.keyword"
            placeholder="名称/编码"
            clearable
            class="w-200px"
            @keyup.enter="handleSearch"
          />
        </ElFormItem>
        <ElFormItem label="场景">
          <ElSelect v-model="searchParams.typeId" class="w-160px">
            <ElOption label="全部" :value="-1" />
            <ElOption v-for="t in typeOptions" :key="t.id" :label="t.name" :value="t.id" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="searchParams.status" class="w-120px">
            <ElOption label="全部" :value="-1" />
            <ElOption label="启用" :value="1" />
            <ElOption label="禁用" :value="0" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">搜索</ElButton>
          <ElButton @click="handleReset">重置</ElButton>
        </ElFormItem>
      </ElForm>

      <div class="h-[calc(100%-100px)]">
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

      <!-- 新建/编辑弹框 -->
      <WorkflowOperateDialog
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </ElCard>
  </div>
</template>
