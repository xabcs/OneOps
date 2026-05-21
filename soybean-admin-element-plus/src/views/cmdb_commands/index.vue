<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { fetchGetCommands } from '@/service/api/cmdb';

defineOptions({
  name: 'CMDBCommands'
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

    const result = await fetchGetCommands(params);
    commands.value = result.list;
    total.value = result.total;
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
  // 显示命令详情对话框
  console.log('Command detail:', command);
}

onMounted(() => {
  getCommands();
});
</script>

<template>
  <div class="commands-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">命令审计</span>
          <el-button type="primary" :icon="ICON_REGISTRY.refresh" @click="getCommands">
            刷新
          </el-button>
        </div>
      </template>

      <!-- 筛选条件 -->
      <div class="filter-bar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="命令">
            <el-input
              v-model="filters.command"
              placeholder="输入命令关键字"
              clearable
              style="width: 200px"
              @keyup.enter="handleSearch"
            />
          </el-form-item>

          <el-form-item label="风险等级">
            <el-select
              v-model="filters.riskLevel"
              placeholder="选择风险等级"
              clearable
              style="width: 150px"
            >
              <el-option
                v-for="option in riskLevelOptions"
                :key="option.value"
                :value="option.value"
                :label="option.label"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="是否拦截">
            <el-select
              v-model="filters.blocked"
              placeholder="选择"
              clearable
              style="width: 120px"
            >
              <el-option :value="true" label="是" />
              <el-option :value="false" label="否" />
            </el-select>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSearch">
              搜索
            </el-button>
            <el-button @click="handleReset">
              重置
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 命令表格 -->
      <el-table
        v-loading="loading"
        :data="commands"
        stripe
        style="width: 100%; margin-top: 16px"
        :default-sort="{ prop: 'executedAt', order: 'descending' }"
      >
        <el-table-column prop="id" label="ID" width="60" />

        <el-table-column label="会话ID" width="80">
          <template #default="{ row }">
            {{ row.sessionId }}
          </template>
        </el-table-column>

        <el-table-column label="用户" width="100">
          <template #default="{ row }">
            {{ row.session?.user?.username || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="服务器" width="120">
          <template #default="{ row }">
            {{ row.session?.server?.hostname || '-' }}
          </template>
        </el-table-column>

        <el-table-column label="命令" min-width="300">
          <template #default="{ row }">
            <div class="command-cell">
              <code class="command-text" :style="{ color: getRiskLevelColor(row.riskLevel) }">
                {{ row.command }}
              </code>
              <el-tag v-if="row.blocked" type="danger" size="small" style="margin-left: 8px">
                已拦截
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="风险等级" width="100">
          <template #default="{ row }">
            <el-tag :type="getRiskLevelType(row.riskLevel)" size="small">
              {{ getRiskLevelText(row.riskLevel) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="退出码" width="80">
          <template #default="{ row }">
            <span :class="row.exitCode === 0 ? 'text-success' : 'text-danger'">
              {{ row.exitCode ?? '-' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="执行时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.executedAt || '') }}
          </template>
        </el-table-column>

        <el-table-column label="输出摘要" width="200">
          <template #default="{ row }">
            <el-text truncated :title="row.outputSummary">
              {{ row.outputSummary || '-' }}
            </el-text>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleViewDetail(row)"
            >
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handlePageChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
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
