<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { fetchGetCommands } from '@/service/api/cmdb';

defineOptions({
  name: 'CmdbAuditCommands'
});

const loading = ref(false);
const commands = ref<Bastion.BastionCommand[]>([]);
const total = ref(0);

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20
});

// 筛选条件
const filters = ref<{
  sessionId?: number;
  riskLevel?: string;
  blocked?: boolean;
  command?: string;
  startDate?: string;
  endDate?: string;
}>({});

// 风险等级选项
const riskLevelOptions = [
  { label: '全部', value: '' },
  { label: '安全', value: 'safe' },
  { label: '低危', value: 'low' },
  { label: '中危', value: 'medium' },
  { label: '高危', value: 'high' },
  { label: '严重', value: 'critical' }
];

// 获取命令列表
async function getCommands() {
  loading.value = true;
  try {
    const params = {
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      ...filters.value
    };

    const { data } = await fetchGetCommands(params);
    commands.value = data?.list || [];
    total.value = data?.total || 0;
  } catch (error) {
    window.$message?.error('获取命令列表失败');
  } finally {
    loading.value = false;
  }
}

// 搜索
function handleSearch() {
  pagination.value.page = 1;
  getCommands();
}

// 重置筛选
function handleReset() {
  filters.value = {
    sessionId: undefined,
    riskLevel: undefined,
    blocked: undefined,
    command: undefined,
    startDate: undefined,
    endDate: undefined
  };
  handleSearch();
}

// 分页变化
function handlePageChange(page: number) {
  pagination.value.page = page;
  getCommands();
}

// 格式化时间
function formatTime(time: string): string {
  return time ? new Date(time).toLocaleString('zh-CN') : '-';
}

// 获取风险等级标签类型
function getRiskLevelType(level: string): 'success' | 'info' | 'warning' | 'danger' {
  switch (level) {
    case 'safe':
      return 'success';
    case 'low':
      return 'info';
    case 'medium':
      return 'warning';
    case 'high':
    case 'critical':
      return 'danger';
    default:
      return 'info';
  }
}

// 获取风险等级文本
function getRiskLevelText(level: string): string {
  const map: Record<string, string> = {
    safe: '安全',
    low: '低危',
    medium: '中危',
    high: '高危',
    critical: '严重'
  };
  return map[level] || level;
}

// 获取风险等级颜色
function getRiskLevelColor(level: string): string {
  const map: Record<string, string> = {
    safe: '#67c23a',
    low: '#409eff',
    medium: '#e6a23c',
    high: '#f56c6c',
    critical: '#ff0000'
  };
  return map[level] || '#909399';
}

// 查看命令详情
function handleViewDetail(command: Bastion.BastionCommand) {
}

onMounted(() => {
  getCommands();
});
</script>

<template>
  <div class="commands-page">
    <ElCard shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">命令审计</span>
          <ElButton type="primary" @click="getCommands">刷新</ElButton>
        </div>
      </template>

      <!-- 筛选条件 -->
      <div class="filter-bar">
        <ElForm :inline="true" :model="filters">
          <ElFormItem label="命令">
            <ElInput
              v-model="filters.command"
              placeholder="输入命令关键字"
              clearable
              style="width: 200px"
              @keyup.enter="handleSearch"
            />
          </ElFormItem>

          <ElFormItem label="风险等级">
            <ElSelect v-model="filters.riskLevel" placeholder="选择风险等级" clearable style="width: 150px">
              <ElOption
                v-for="option in riskLevelOptions"
                :key="option.value"
                :value="option.value"
                :label="option.label"
              />
            </ElSelect>
          </ElFormItem>

          <ElFormItem label="是否拦截">
            <ElSelect v-model="filters.blocked" placeholder="选择" clearable style="width: 120px">
              <ElOption :value="true" label="是" />
              <ElOption :value="false" label="否" />
            </ElSelect>
          </ElFormItem>

          <ElFormItem>
            <ElButton type="primary" @click="handleSearch">搜索</ElButton>
            <ElButton @click="handleReset">重置</ElButton>
          </ElFormItem>
        </ElForm>
      </div>

      <!-- 命令表格 -->
      <ElTable
        v-loading="loading"
        :data="commands"
        stripe
        style="width: 100%; margin-top: 16px"
        :default-sort="{ prop: 'executedAt', order: 'descending' }"
      >
        <ElTableColumn prop="id" label="ID" width="60" />

        <ElTableColumn label="会话ID" width="80">
          <template #default="{ row }">
            {{ row.sessionId }}
          </template>
        </ElTableColumn>

        <ElTableColumn label="用户" width="100">
          <template #default="{ row }">
            {{ row.session?.user?.username || '-' }}
          </template>
        </ElTableColumn>

        <ElTableColumn label="服务器" width="120">
          <template #default="{ row }">
            {{ row.session?.server?.hostname || '-' }}
          </template>
        </ElTableColumn>

        <ElTableColumn label="命令" min-width="300">
          <template #default="{ row }">
            <div class="command-cell">
              <code class="command-text" :style="{ color: getRiskLevelColor(row.riskLevel) }">
                {{ row.command }}
              </code>
              <ElTag v-if="row.blocked" type="danger" size="small" style="margin-left: 8px">已拦截</ElTag>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="风险等级" width="100">
          <template #default="{ row }">
            <ElTag :type="getRiskLevelType(row.riskLevel)" size="small">
              {{ getRiskLevelText(row.riskLevel) }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="退出码" width="80">
          <template #default="{ row }">
            <span :class="row.exitCode === 0 ? 'text-success' : 'text-danger'">
              {{ row.exitCode ?? '-' }}
            </span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="执行时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.executedAt || '') }}
          </template>
        </ElTableColumn>

        <ElTableColumn label="输出摘要" width="200">
          <template #default="{ row }">
            <ElText truncated :title="row.outputSummary">
              {{ row.outputSummary || '-' }}
            </ElText>
          </template>
        </ElTableColumn>

        <ElTableColumn label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="handleViewDetail(row)">详情</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handlePageChange"
          @current-change="handlePageChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<style scoped>
.commands-page {
  padding: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 16px;
  font-weight: 500;
}

.filter-bar {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
}

.command-cell {
  display: flex;
  align-items: center;
}

.command-text {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 13px;
  word-break: break-all;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}

:deep(.el-table__cell) {
  padding: 8px 0;
}
</style>
