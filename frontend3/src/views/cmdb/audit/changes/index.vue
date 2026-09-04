<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Refresh, Search } from '@element-plus/icons-vue';
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
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <ElForm :model="searchParams" inline class="mb-16px">
        <ElFormItem label="资产类型">
          <ElSelect v-model="searchParams.assetType" placeholder="请选择资产类型" clearable style="width: 160px">
            <ElOption label="服务器" value="server" />
            <ElOption label="业务系统" value="business" />
            <ElOption label="机房" value="room" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="资产ID">
          <ElInputNumber v-model="searchParams.assetId" :min="1" controls-position="right" style="width: 160px" />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" :icon="Search" @click="handleSearch">搜索</ElButton>
          <ElButton :icon="Refresh" @click="handleReset">重置</ElButton>
        </ElFormItem>
      </ElForm>

      <ElTable v-loading="loading" :data="data" border stripe>
        <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
      </ElTable>

      <div class="mt-16px flex justify-end">
        <ElPagination
          v-if="mobilePagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
    </ElCard>
  </div>
</template>
