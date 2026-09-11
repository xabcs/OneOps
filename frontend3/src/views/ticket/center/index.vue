<script setup lang="tsx">
  import { computed, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import type { TabPaneName } from 'element-plus';
  import dayjs from 'dayjs';
  import { Plus } from '@element-plus/icons-vue';
  import { fetchTicketTypeOptions, fetchTickets } from '@/service/api';
  import { useAuthStore } from '@/store/modules/auth';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';
  import { ticketPriorityMap, ticketStatusMap, ticketStatusOptions } from '../constants';
  import TicketCreateDialog from './modules/ticket-create-dialog.vue';

  defineOptions({ name: 'TicketCenter' });

  const router = useRouter();
  const authStore = useAuthStore();

  interface SearchParams {
    page: number;
    pageSize: number;
    scope: 'todo' | 'created' | 'done' | 'all';
    status: string;
    typeId: number | undefined;
    keyword: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 10, scope: 'todo', status: '', typeId: undefined, keyword: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  // 全部工单 tab 仅对有权限者可见
  const canViewAll = computed(() => authStore.hasPermission('ticket.ticket.list'));

  // 工单类型选项（筛选 + 发起）
  const typeOptions = ref<Api.Ticket.TicketTypeOption[]>([]);
  async function loadTypeOptions() {
    const { data, error } = await fetchTicketTypeOptions();
    if (!error && data) {
      typeOptions.value = data;
    }
  }
  loadTypeOptions();

  const { columns, data, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchTickets(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 10;
    },
    columns: () => [
      { prop: 'ticketNo', label: '工单号', width: 150, align: 'center' },
      { prop: 'title', label: '标题', minWidth: 180, align: 'center', showOverflowTooltip: true },
      { prop: 'typeName', label: '类型', width: 110, align: 'center' },
      {
        prop: 'priority',
        label: '优先级',
        width: 80,
        align: 'center',
        formatter: row => {
          const item = ticketPriorityMap[row.priority] ?? ticketPriorityMap.normal;
          return <ElTag type={item.type}>{item.label}</ElTag>;
        }
      },
      {
        prop: 'status',
        label: '状态',
        width: 90,
        align: 'center',
        formatter: row => {
          const item = ticketStatusMap[row.status] ?? ticketStatusMap.pending;
          return <ElTag type={item.type}>{item.label}</ElTag>;
        }
      },
      {
        prop: 'currentNodeName',
        label: '当前节点',
        minWidth: 140,
        align: 'center',
        formatter: row => (
          <div>
            <div>{row.status === 'pending' ? row.currentNodeName || '-' : '已结束'}</div>
            {row.status === 'pending' && row.currentApprovers ? (
              <div class="text-12px text-gray-400">待审批：{row.currentApprovers}</div>
            ) : null}
          </div>
        )
      },
      { prop: 'creatorName', label: '发起人', width: 100, align: 'center' },
      {
        prop: 'createdAt',
        label: '创建时间',
        width: 170,
        align: 'center',
        formatter: row => dayjs(row.createdAt).format('YYYY-MM-DD HH:mm:ss')
      },
      {
        prop: 'operate',
        label: '操作',
        width: 90,
        fixed: 'right',
        align: 'center',
        className: 'msre-table-actions',
        formatter: row => (
          <ElButton link type="primary" size="small" onClick={() => handleView(row.id)}>
            {row.canApprove ? '去审批' : '查看'}
          </ElButton>
        )
      }
    ]
  });

  function handleView(id: number) {
    // 注意：必须用 path 导航。按 name 导航 vue-router 不会深入空 path 子路由，详情页会空白
    router.push({ path: '/ticket/center/detail', query: { id: String(id) } });
  }

  function handleTabChange(name: TabPaneName) {
    searchParams.value.scope = name as SearchParams['scope'];
    getDataByPage(1);
  }

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    const scope = searchParams.value.scope;
    searchParams.value = getInitSearchParams();
    searchParams.value.scope = scope;
    getDataByPage(1);
  }

  // 发起工单弹窗
  const createVisible = ref(false);
  function handleCreated() {
    // 新工单出现在"我发起的"，也可能立即需要自己审批（发起人节点）
    searchParams.value.scope = 'created';
    getDataByPage(1);
  }
</script>

<template>
  <ListPageLayout
    title="工单中心"
    description="发起与跟进工单，集中处理待我审批的事项"
    :pagination="mobilePagination"
    embedded-search
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 视图切换：页签渲染在表格卡片 header，与内容联动 -->
    <template #tabs>
      <ElTabs :model-value="searchParams.scope" @tab-change="handleTabChange">
        <ElTabPane label="待我审批" name="todo" />
        <ElTabPane label="我发起的" name="created" />
        <ElTabPane label="我已审批" name="done" />
        <ElTabPane v-if="canViewAll" label="全部工单" name="all" />
      </ElTabs>
    </template>

    <!-- 搜索筛选 -->
    <template #search>
      <ElInput
        v-model="searchParams.keyword"
        placeholder="标题/工单号/发起人"
        clearable
        class="w-200px"
        @keyup.enter="handleSearch"
      />
      <ElSelect v-model="searchParams.status" clearable placeholder="全部" class="w-130px">
        <ElOption v-for="opt in ticketStatusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
      </ElSelect>
      <ElSelect v-model="searchParams.typeId" clearable placeholder="全部" class="w-150px">
        <ElOption v-for="t in typeOptions" :key="t.id" :label="t.name" :value="t.id" />
      </ElSelect>
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <ElButton type="primary" :icon="Plus" @click="createVisible = true">发起工单</ElButton>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" height="100%" :data="data" :border="false" row-key="id">
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>
    <!-- 发起工单（teleport 弹层须置于布局内，保持页面单根节点以正常继承 attrs 与 Transition） -->
    <TicketCreateDialog v-model:visible="createVisible" :type-options="typeOptions" @created="handleCreated" />
  </ListPageLayout>
</template>
