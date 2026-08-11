<script setup lang="tsx">
import { onMounted, ref } from 'vue';
import { fetchApplications, fetchOperationLogs } from '@/service/api/application-permission';

defineOptions({ name: 'AuthCenterOperationLogs' });

const loading = ref(false);
const tableData = ref<Api.ApplicationPermission.ApplicationOperationLog[]>([]);
const applications = ref<Api.ApplicationPermission.Application[]>([]);

const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0,
  appId: null as number | null
});

async function getApplications() {
  const { data, error } = await fetchApplications({ page: 1, pageSize: 1000 });
  if (!error && data) {
    applications.value = data.list || [];
  }
}

async function getData() {
  if (!pagination.value.appId) return;

  loading.value = true;
  try {
    const { data, error } = await fetchOperationLogs(pagination.value.appId, {
      page: pagination.value.page,
      pageSize: pagination.value.pageSize
    });

    if (!error && data) {
      tableData.value = data.list || [];
      pagination.value.total = data.total || 0;
    }
  } finally {
    loading.value = false;
  }
}

function handleAppChange() {
  pagination.value.page = 1;
  getData();
}

function handlePageChange(page: number) {
  pagination.value.page = page;
  getData();
}

function handleSizeChange(size: number) {
  pagination.value.pageSize = size;
  pagination.value.page = 1;
  getData();
}

onMounted(() => {
  getApplications();
});
</script>

<template>
    <div class="min-h-500px flex-col gap-4">
        <!-- 应用选择 -->
        <ElCard shadow="never">
            <ElSelect v-model="pagination.appId" placeholder="请选择应用查看操作日志" class="w-300px" @change="handleAppChange">
                <ElOption v-for="app in applications" :key="app.id" :label="app.name" :value="app.id" />
            </ElSelect>
        </ElCard>

        <!-- 表格 -->
        <ElCard shadow="never" class="flex-1">
            <template #header>
                <span class="text-lg font-medium">操作日志列表</span>
            </template>

            <ElTable v-loading="loading" :data="tableData" :border="false">
                <ElTableColumn type="index" label="序号" width="60" align="center" />
                <ElTableColumn prop="operation" label="操作类型" align="center" min-width="120" />
                <ElTableColumn prop="target" label="操作目标" align="center" min-width="150" />
                <ElTableColumn prop="status" label="状态" align="center" width="100">
                    <template #default="{ row }">
                        <ElTag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'">
                            {{ row.status === 'success' ? '成功' : row.status === 'failed' ? '失败' : '进行中' }}
                        </ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="operator" label="操作人" align="center" min-width="120" />
                <ElTableColumn prop="createdAt" label="操作时间" align="center" min-width="160" />
            </ElTable>

            <div class="mt-4 flex justify-end">
                <ElPagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" @size-change="handleSizeChange" @current-change="handlePageChange" />
            </div>
        </ElCard>
    </div>
</template>
