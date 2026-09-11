<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import type { FlatResponseData } from '@sa/axios';
  import { fetchGetAssetChanges } from '@/service/api';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';

  defineOptions({ name: 'CmdbChanges' });

  interface SearchParams {
    page: number;
    pageSize: number;
    assetType: string;
    assetId: number | undefined;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 20, assetType: '', assetId: undefined };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const { columns, data, loading, mobilePagination, getData, getDataByPage } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () =>
      fetchGetAssetChanges({
        assetType: searchParams.value.assetType || undefined,
        assetId: searchParams.value.assetId,
        page: searchParams.value.page,
        pageSize: searchParams.value.pageSize
      }),
    transform: response =>
      defaultTransform<CMDB.AssetChange>(
        response as unknown as FlatResponseData<unknown, Api.Common.PaginatingQueryRecord<CMDB.AssetChange>>
      ),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 20;
    },
    columns: () => [
      { prop: 'id', label: 'ID', width: 80 },
      { prop: 'assetType', label: '资产类型', width: 120 },
      { prop: 'assetId', label: '资产ID', width: 100 },
      { prop: 'assetName', label: '资产名称', minWidth: 140, showOverflowTooltip: true },
      {
        prop: 'changeType',
        label: '变更类型',
        width: 110,
        formatter: row => {
          const t = getChangeTypeTag(row.changeType);
          return (
            <ElTag type={t.type} size="small">
              {t.text}
            </ElTag>
          );
        }
      },
      { prop: 'fieldName', label: '字段名', width: 130 },
      { prop: 'oldValue', label: '旧值', minWidth: 160, showOverflowTooltip: true },
      { prop: 'newValue', label: '新值', minWidth: 160, showOverflowTooltip: true },
      { prop: 'operator', label: '操作人', width: 120 },
      { prop: 'operateTime', label: '变更时间', width: 180 },
      { prop: 'remarks', label: '备注', minWidth: 180, showOverflowTooltip: true }
    ]
  });

  function getChangeTypeTag(type: string): {
    text: string;
    type: 'success' | 'warning' | 'danger' | 'info' | 'primary';
  } {
    const typeMap: Record<string, { text: string; type: 'success' | 'warning' | 'danger' | 'info' | 'primary' }> = {
      create: { text: '创建', type: 'success' },
      update: { text: '更新', type: 'warning' },
      delete: { text: '删除', type: 'danger' }
    };
    return typeMap[type] || { text: type, type: 'info' };
  }

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <ListPageLayout
    title="资产变更记录"
    description="追踪资产创建、更新与删除的字段级变更历史"
    :pagination="mobilePagination"
    @search="handleSearch"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElSelect v-model="searchParams.assetType" placeholder="请选择资产类型" clearable class="w-160px">
        <ElOption label="服务器" value="server" />
        <ElOption label="业务系统" value="business" />
        <ElOption label="机房" value="room" />
      </ElSelect>
      <ElInputNumber
        v-model="searchParams.assetId"
        :min="1"
        controls-position="right"
        placeholder="资产ID"
        class="w-160px"
      />
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" :data="data" border stripe height="100%">
      <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
    </ElTable>
  </ListPageLayout>
</template>
