<script setup lang="tsx">
  import { ref } from 'vue';
  import { Delete, Edit, Plus } from '@element-plus/icons-vue';
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
        formatter: row => (
          <ElSpace>
            <PermissionButton
              code="ticket.type.update"
              size="small"
              type="warning"
              icon={Edit}
              onClick={() => handleEdit(row.id)}
            >
              编辑
            </PermissionButton>
            <ElPopconfirm title="确认删除该工单类型？" onConfirm={() => handleDelete(row.id)}>
              {{
                reference: () => (
                  <PermissionButton code="ticket.type.delete" size="small" type="danger" icon={Delete}>
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
  <div class="table-page">
    <ElCard class="card-wrapper sm:flex-1-hidden">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-medium">工单类型管理</span>
          <PermissionButton code="ticket.type.create" type="primary" :icon="Plus" @click="handleAdd">
            新建类型
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
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>

      <!-- 新建/编辑弹框 -->
      <TypeOperateDialog
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getDataByPage"
      />
    </ElCard>
  </div>
</template>
