<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { fetchGetSystemEventLogs } from '@/service/api';
import type { Audit } from '@/typings/api';
import { ElNotification } from 'element-plus';

defineOptions({ name: 'AuditSystemEvents' });

const searchForm = reactive<Audit.LogQuery>({
  level: '',
  source: '',
  category: '',
  startTime: '',
  endTime: ''
});

const tableData = ref<Audit.SystemEventLog[]>([]);
const loading = ref(false);
const total = ref(0);

const pagination = reactive({
  page: 1,
  pageSize: 20
});

const levelOptions = [
  { label: '信息', value: 'info' },
  { label: '警告', value: 'warning' },
  { label: '错误', value: 'error' },
  { label: '严重', value: 'critical' }
];

async function getSystemEventLogs() {
  loading.value = true;
  try {
    const { data, error } = await fetchGetSystemEventLogs({
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize
    });

    if (!error && data) {
      tableData.value = data.list || [];
      total.value = data.total || 0;
    }
  } catch (err) {
    console.error('获取系统事件日志失败:', err);
    ElNotification({
      title: '错误',
      message: '获取系统事件日志失败',
      type: 'error'
    });
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.page = 1;
  getSystemEventLogs();
}

function handleReset() {
  Object.assign(searchForm, {
    level: '',
    source: '',
    category: '',
    startTime: '',
    endTime: ''
  });
  pagination.page = 1;
  getSystemEventLogs();
}

function handlePageChange(page: number) {
  pagination.page = page;
  getSystemEventLogs();
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  getSystemEventLogs();
}

function getLevelTag(level: string) {
  const levelMap: Record<string, { text: string; type: any }> = {
    info: { text: '信息', type: 'info' },
    warning: { text: '警告', type: 'warning' },
    error: { text: '错误', type: 'danger' },
    critical: { text: '严重', type: 'danger' }
  };
  return levelMap[level] || { text: level, type: 'info' };
}

onMounted(() => {
  getSystemEventLogs();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <ElForm :model="searchForm" inline class="search-form">
        <ElFormItem label="级别">
          <ElSelect v-model="searchForm.level" placeholder="请选择级别" clearable style="width: 150px">
            <ElOption
              v-for="option in levelOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="来源">
          <ElInput v-model="searchForm.source" placeholder="请输入来源" clearable style="width: 150px" />
        </ElFormItem>
        <ElFormItem label="分类">
          <ElInput v-model="searchForm.category" placeholder="请输入分类" clearable style="width: 150px" />
        </ElFormItem>
        <ElFormItem label="事件时间">
          <ElDatePicker
            v-model="searchForm.startTime"
            type="datetime"
            placeholder="开始时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px"
          />
          <span class="mx-2">至</span>
          <ElDatePicker
            v-model="searchForm.endTime"
            type="datetime"
            placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 200px"
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">
            搜索
          </ElButton>
          <ElButton @click="handleReset">
            重置
          </ElButton>
        </ElFormItem>
      </ElForm>

      <ElTable
        v-loading="loading"
        :data="tableData"
        border
        stripe
        class="h-full"
        height="calc(100vh - 400px)"
      >
        <ElTableColumn prop="id" label="ID" width="70" align="center" />
        <ElTableColumn label="级别" width="70" align="center">
          <template #default="{ row }">
            <ElTag :type="getLevelTag(row.level).type">
              {{ getLevelTag(row.level).text }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="source" label="来源" width="100" align="center" />
        <ElTableColumn prop="category" label="分类" width="100" align="center" />
        <ElTableColumn prop="message" label="消息" min-width="200" align="center" show-overflow-tooltip />
        <ElTableColumn prop="details" label="详情" min-width="150" align="center" show-overflow-tooltip />
        <ElTableColumn prop="ip" label="IP地址" width="130" align="center" />
        <ElTableColumn prop="time" label="事件时间" width="160" align="center" />
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

<style scoped lang="scss">
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>