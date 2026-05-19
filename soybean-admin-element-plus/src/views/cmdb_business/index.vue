<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { Ref } from 'vue';
import { $t } from '@/locales';
import { fetchGetBusinessUnits } from '@/service/api';
import type { BusinessUnit } from '@/typings/api/cmdb';

defineOptions({
  name: 'CmdbBusiness'
});

const loading = ref(false);
const tableData = Ref<BusinessUnit[]>([]);

async function getTableData() {
  loading.value = true;
  const { data } = await fetchGetBusinessUnits();
  if (data) {
    tableData.value = data || [];
  }
  loading.value = false;
}

onMounted(() => {
  getTableData();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ACard title="业务管理" :bordered="false" class="card-wrapper">
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
      <ATable :loading="loading" :data="tableData" :pagination="false" :bordered="false">
        <ATableColumn title="ID" dataIndex="id" :width="80" />
        <ATableColumn title="业务名称" dataIndex="name" :width="150" />
        <ATableColumn title="业务代码" dataIndex="code" :width="120" />
        <ATableColumn title="负责人" dataIndex="owner" :width="100" />
        <ATableColumn title="层级" dataIndex="level" :width="80" />
        <ATableColumn title="排序" dataIndex="sortOrder" :width="80" />
        <ATableColumn title="状态" :width="80">
          <template #default="{ row }">
            <ATag v-if="row.status === 1" color="success">启用</ATag>
            <ATag v-else color="error">禁用</ATag>
          </template>
        </ATableColumn>
        <ATableColumn title="备注" dataIndex="remarks" :width="200" />
      </ATable>
    </ACard>
  </div>
</template>

<style scoped></style>
