<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { fetchGetAssetChanges } from '@/service/api';
import { ElNotification } from 'element-plus';

defineOptions({ name: 'CmdbChanges' });

const loading = ref(false);
const tableData = ref<CMDB.AssetChange[]>([]);
const total = ref(0);

const searchForm = reactive({
  assetType: '',
  assetId: undefined as number | undefined
});

const pagination = reactive({
  page: 1,
  pageSize: 20
});

async function getTableData() {
  loading.value = true;
  try {
    const { data } = await fetchGetAssetChanges({
      assetType: searchForm.assetType || undefined,
      assetId: searchForm.assetId,
      page: pagination.page,
      pageSize: pagination.pageSize
    });
    tableData.value = data?.list || [];
    total.value = data?.total || 0;
  } catch (error) {
    console.error('获取变更记录失败:', error);
    ElNotification.error('获取变更记录失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.page = 1;
  getTableData();
}

function handleReset() {
  Object.assign(searchForm, {
    assetType: '',
    assetId: undefined
  });
  pagination.page = 1;
  getTableData();
}

function handlePageChange(page: number) {
  pagination.page = page;
  getTableData();
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  getTableData();
}

function getChangeTypeTag(type: string) {
  const typeMap: Record<string, { text: string; type: 'success' | 'warning' | 'danger' | 'info' }> = {
    create: { text: '创建', type: 'success' },
    update: { text: '更新', type: 'warning' },
    delete: { text: '删除', type: 'danger' }
  };
  return typeMap[type] || { text: type, type: 'info' };
}

onMounted(() => {
  getTableData();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <ElForm :model="searchForm" inline class="mb-16px">
        <ElFormItem label="资产类型">
          <ElSelect v-model="searchForm.assetType" placeholder="请选择资产类型" clearable style="width: 160px">
            <ElOption label="服务器" value="server" />
            <ElOption label="业务系统" value="business" />
            <ElOption label="机房" value="room" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="资产ID">
          <ElInputNumber v-model="searchForm.assetId" :min="1" controls-position="right" style="width: 160px" />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">搜索</ElButton>
          <ElButton @click="handleReset">重置</ElButton>
        </ElFormItem>
      </ElForm>

      <ElTable v-loading="loading" :data="tableData" border stripe>
        <ElTableColumn prop="id" label="ID" width="80" />
        <ElTableColumn prop="assetType" label="资产类型" width="120" />
        <ElTableColumn prop="assetId" label="资产ID" width="100" />
        <ElTableColumn prop="assetName" label="资产名称" min-width="140" show-overflow-tooltip />
        <ElTableColumn label="变更类型" width="110">
          <template #default="{ row }">
            <ElTag :type="getChangeTypeTag(row.changeType).type" size="small">
              {{ getChangeTypeTag(row.changeType).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="fieldName" label="字段名" width="130" />
        <ElTableColumn prop="oldValue" label="旧值" min-width="160" show-overflow-tooltip />
        <ElTableColumn prop="newValue" label="新值" min-width="160" show-overflow-tooltip />
        <ElTableColumn prop="operator" label="操作人" width="120" />
        <ElTableColumn prop="operateTime" label="变更时间" width="180" />
        <ElTableColumn prop="remarks" label="备注" min-width="180" show-overflow-tooltip />
      </ElTable>

      <div class="flex justify-end mt-16px">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </ElCard>
  </div>
</template>
