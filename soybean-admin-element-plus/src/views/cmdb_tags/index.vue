<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { Ref } from 'vue';
import { $t } from '@/locales';
import { fetchGetServerTags } from '@/service/api';
import type { ServerTag } from '@/typings/api/cmdb';

defineOptions({
  name: 'CmdbTags'
});

const loading = ref(false);
const tableData = Ref<ServerTag[]>([]);

async function getTableData() {
  loading.value = true;
  const { data } = await fetchGetServerTags();
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
    <ACard title="标签管理" :bordered="false" class="card-wrapper">
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
        <ATableColumn title="标签名称" dataIndex="name" :width="150" />
        <ATableColumn title="标签代码" dataIndex="code" :width="120" />
        <ATableColumn title="颜色" dataIndex="color" :width="100">
          <template #default="{ row }">
            <span v-if="row.color" :style="{ backgroundColor: row.color }" class="inline-block w-40px h-24px rounded-4px" />
          </template>
        </ATableColumn>
        <ATableColumn title="描述" dataIndex="description" :width="200" />
        <ATableColumn title="排序" dataIndex="sortOrder" :width="80" />
        <ATableColumn title="状态" :width="80">
          <template #default="{ row }">
            <ATag v-if="row.status === 1" color="success">启用</ATag>
            <ATag v-else color="error">禁用</ATag>
          </template>
        </ATableColumn>
      </ATable>
    </ACard>
  </div>
</template>

<style scoped></style>
