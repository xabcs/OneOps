<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue';
  import { Search } from '@element-plus/icons-vue';
  import { type DiagnosticExecution, fetchDiagnosticExecutions } from '@/service/api/diagnostic';
  import { channelLabel, formatDateTime, formatDuration, riskMeta } from '../shared';

  defineOptions({ name: 'DiagnosticHistoryTab' });

  const filters = reactive({
    appName: '',
    username: '',
    command: '',
    riskLevel: ''
  });

  const pagination = reactive({ page: 1, pageSize: 20, total: 0 });
  const loading = ref(false);
  const executions = ref<DiagnosticExecution[]>([]);

  const detailVisible = ref(false);
  const detail = ref<DiagnosticExecution | null>(null);

  async function load() {
    loading.value = true;
    try {
      const { data, error } = await fetchDiagnosticExecutions({
        page: pagination.page,
        pageSize: pagination.pageSize,
        appName: filters.appName || undefined,
        username: filters.username || undefined,
        command: filters.command || undefined,
        riskLevel: filters.riskLevel || undefined
      });
      if (!error && data) {
        executions.value = data.list ?? [];
        pagination.total = data.total ?? 0;
      }
    } finally {
      loading.value = false;
    }
  }

  function search() {
    pagination.page = 1;
    load();
  }

  function showDetail(row: DiagnosticExecution) {
    detail.value = row;
    detailVisible.value = true;
  }

  onMounted(load);
</script>

<template>
  <div class="history-tab">
    <div class="history-filter">
      <ElInput v-model="filters.appName" placeholder="应用名" clearable class="f-input" @keyup.enter="search" />
      <ElInput v-model="filters.username" placeholder="操作者" clearable class="f-input-sm" @keyup.enter="search" />
      <ElInput v-model="filters.command" placeholder="命令关键字" clearable class="f-input" @keyup.enter="search" />
      <ElSelect v-model="filters.riskLevel" placeholder="全部风险级" clearable class="f-input-sm">
        <ElOption v-for="level in ['L0', 'L1', 'L2', 'L3', 'L4', 'L5']" :key="level" :label="level" :value="level" />
      </ElSelect>
      <ElButton type="primary" :icon="Search" :loading="loading" @click="search">查询</ElButton>
    </div>

    <ElTable v-loading="loading" :data="executions" size="small">
      <ElTableColumn label="时间" width="165">
        <template #default="{ row }">{{ formatDateTime(row.timestamp) }}</template>
      </ElTableColumn>
      <ElTableColumn prop="appName" label="应用" min-width="130" show-overflow-tooltip />
      <ElTableColumn prop="podName" label="Pod" min-width="200" show-overflow-tooltip />
      <ElTableColumn label="命令" min-width="240" show-overflow-tooltip>
        <template #default="{ row }">
          <code class="cmd-text">{{ row.command }}</code>
        </template>
      </ElTableColumn>
      <ElTableColumn label="风险" width="80">
        <template #default="{ row }">
          <ElTooltip :content="riskMeta(row.riskLevel).desc" placement="top">
            <ElTag size="small" :type="riskMeta(row.riskLevel).type">{{ row.riskLevel }}</ElTag>
          </ElTooltip>
        </template>
      </ElTableColumn>
      <ElTableColumn label="渠道" width="90">
        <template #default="{ row }">{{ channelLabel(row.channel) }}</template>
      </ElTableColumn>
      <ElTableColumn label="结果" width="80">
        <template #default="{ row }">
          <ElTag size="small" :type="row.resultStatus === 'success' ? 'success' : 'danger'">
            {{ row.resultStatus === 'success' ? '成功' : '失败' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="耗时" width="85">
        <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
      </ElTableColumn>
      <ElTableColumn prop="username" label="操作者" width="100" />
      <ElTableColumn label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <ElButton size="small" link type="primary" @click="showDetail(row)">详情</ElButton>
        </template>
      </ElTableColumn>
      <template #empty>
        <ElEmpty description="暂无执行记录" :image-size="80" />
      </template>
    </ElTable>

    <div class="history-pagination">
      <ElPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="load"
        @size-change="search"
      />
    </div>

    <ElDialog v-model="detailVisible" title="执行详情" width="760px" top="8vh">
      <template v-if="detail">
        <ElDescriptions :column="3" size="small" border class="detail-desc">
          <ElDescriptionsItem label="应用">{{ detail.appName }}</ElDescriptionsItem>
          <ElDescriptionsItem label="Pod">{{ detail.podName || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="命名空间">{{ detail.namespace || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="操作者">{{ detail.username }}</ElDescriptionsItem>
          <ElDescriptionsItem label="渠道">{{ channelLabel(detail.channel) }}</ElDescriptionsItem>
          <ElDescriptionsItem label="耗时">{{ formatDuration(detail.duration) }}</ElDescriptionsItem>
          <ElDescriptionsItem label="命令" :span="3">
            <code class="cmd-text">{{ detail.command }}</code>
          </ElDescriptionsItem>
          <ElDescriptionsItem label="输出" :span="3">
            <pre class="detail-output">{{ detail.resultOutput || detail.resultError || '无输出' }}</pre>
          </ElDescriptionsItem>
        </ElDescriptions>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped lang="scss">
  .history-tab {
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 12px;

    .history-filter {
      display: flex;
      gap: 10px;
      margin-bottom: 10px;

      .f-input {
        width: 200px;
      }

      .f-input-sm {
        width: 140px;
      }
    }

    .cmd-text {
      color: var(--el-color-primary);
      font-size: 12px;
    }

    .history-pagination {
      display: flex;
      justify-content: flex-end;
      margin-top: 10px;
    }

    .detail-desc {
      .detail-output {
        background: #1e1e1e;
        color: #d4d4d4;
        padding: 12px;
        border-radius: 4px;
        font-size: 12px;
        line-height: 1.6;
        max-height: 48vh;
        overflow: auto;
        white-space: pre-wrap;
        word-break: break-all;
        margin: 0;
      }
    }
  }
</style>
