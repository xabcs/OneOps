<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { Ref } from 'vue';
import { $t } from '@/locales';
import { fetchGetAssetChanges } from '@/service/api';
import type { AssetChange } from '@/typings/api/cmdb';

defineOptions({
  name: 'CmdbChanges'
});

const loading = ref(false);
const tableData = Ref<AssetChange[]>([]);
const pagination = ref({
  page: 1,
  pageSize: 20
});

async function getTableData() {
  loading.value = true;
  const { data } = await fetchGetAssetChanges({
    page: pagination.value.page,
    pageSize: pagination.value.pageSize
  });
  if (data) {
    tableData.value = data.list || [];
  }
  loading.value = false;
}

onMounted(() => {
  getTableData();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ACard title="变更记录" :bordered="false" class="card-wrapper">
      <template #extra>
        <ASpace>
          <AButton type="primary" @click="getTableData">
            <template #icon>
              <icon-ic-round-refresh class="align-sub text-icon" />
            </template>
            <span class="ml-8px">{{ $t('common.refresh') }}</span>
          </AButton>
        </ASpace>
      </template>
      <ATable :loading="loading" :data="tableData" :pagination="pagination" :bordered="false">
        <ATableColumn title="ID" dataIndex="id" :width="80" />
        <ATableColumn title="资产类型" dataIndex="assetType" :width="120" />
        <ATableColumn title="变更类型" dataIndex="changeType" :width="100">
          <template #default="{ row }">
            <ATag v-if="row.changeType === 'create'" color="success">创建</ATag>
            <ATag v-else-if="row.changeType === 'update'" color="warning">更新</ATag>
            <ATag v-else-if="row.changeType === 'delete'" color="error">删除</ATag>
            <ATag v-else>{{ row.changeType }}</ATag>
          </template>
        </ATableColumn>
        <ATableColumn title="字段名" dataIndex="fieldName" :width="120" />
        <ATableColumn title="旧值" dataIndex="oldValue" :width="150" />
        <ATableColumn title="新值" dataIndex="newValue" :width="150" />
        <ATableColumn title="操作人" dataIndex="operator" :width="100" />
        <ATableColumn title="变更时间" dataIndex="createdAt" :width="180" />
        <ATableColumn title="备注" dataIndex="remarks" :width="200" />
      </ATable>
    </ACard>
  </div>
</template>

<style scoped></style>
