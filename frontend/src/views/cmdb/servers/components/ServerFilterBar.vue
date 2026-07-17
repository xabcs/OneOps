<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElButton, ElForm, ElFormItem, ElIcon, ElInput, ElOption, ElSelect, ElSpace } from 'element-plus';
import { Refresh, RefreshRight, Search } from '@element-plus/icons-vue';
import type { ServerFilters, ServerStats } from '../types/server.types';

interface Props {
  showStats?: boolean;
}

interface Emits {
  (e: 'search', filters: ServerFilters): void;
  (e: 'reset'): void;
  (e: 'refresh'): void;
}

const props = withDefaults(defineProps<Props>(), {
  showStats: true
});

const emit = defineEmits<Emits>();

// 过滤条件
const filters = reactive<ServerFilters>({
  keyword: '',
  env: '',
  agentStatus: '',
  status: undefined,
  groupId: undefined
});

// 统计信息
const stats = reactive<ServerStats>({
  total: 0,
  online: 0,
  offline: 0,
  warning: 0,
  healthy: 0
});

// 刷新状态
const refreshing = ref(false);

// 搜索
const handleSearch = () => {
  emit('search', { ...filters });
};

// 重置
const handleReset = () => {
  filters.keyword = '';
  filters.env = '';
  filters.agentStatus = '';
  filters.status = undefined;
  filters.groupId = undefined;
  emit('reset');
};

// 刷新
const handleRefresh = () => {
  refreshing.value = true;
  emit('refresh');
  setTimeout(() => {
    refreshing.value = false;
  }, 1000);
};

// 更新统计信息
const updateStats = (newStats: ServerStats) => {
  Object.assign(stats, newStats);
};

// 暴露方法给父组件
defineExpose({
  updateStats
});
</script>

<template>
  <div class="server-filter-bar">
    <ElForm inline class="filter-form">
      <!-- 搜索框 -->
      <ElFormItem>
        <ElInput
          v-model="filters.keyword"
          placeholder="搜索主机名或IP"
          clearable
          style="width: 200px"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <ElIcon><Search /></ElIcon>
          </template>
        </ElInput>
      </ElFormItem>

      <!-- 环境筛选 -->
      <ElFormItem>
        <ElSelect v-model="filters.env" placeholder="环境" clearable style="width: 100px" @change="handleSearch">
          <ElOption label="全部" value="" />
          <ElOption label="生产" value="prod" />
          <ElOption label="测试" value="test" />
          <ElOption label="开发" value="dev" />
        </ElSelect>
      </ElFormItem>

      <!-- Agent状态筛选 -->
      <ElFormItem>
        <ElSelect
          v-model="filters.agentStatus"
          placeholder="Agent状态"
          clearable
          style="width: 120px"
          @change="handleSearch"
        >
          <ElOption label="全部" value="" />
          <ElOption label="运行中" value="running" />
          <ElOption label="离线" value="offline" />
          <ElOption label="未安装" value="uninstalled" />
        </ElSelect>
      </ElFormItem>

      <!-- 服务器状态筛选 -->
      <ElFormItem>
        <ElSelect v-model="filters.status" placeholder="状态" clearable style="width: 100px" @change="handleSearch">
          <ElOption label="全部" :value="undefined" />
          <ElOption label="在线" :value="1" />
          <ElOption label="离线" :value="0" />
        </ElSelect>
      </ElFormItem>

      <!-- 操作按钮 -->
      <ElFormItem>
        <ElButton type="primary" @click="handleSearch">
          <ElIcon><Search /></ElIcon>
          搜索
        </ElButton>
        <ElButton @click="handleReset">
          <ElIcon><RefreshRight /></ElIcon>
          重置
        </ElButton>
        <ElButton :loading="refreshing" @click="handleRefresh">
          <ElIcon><Refresh /></ElIcon>
          刷新
        </ElButton>
      </ElFormItem>
    </ElForm>

    <!-- 统计信息 -->
    <div v-if="showStats" class="stats-bar">
      <ElSpace :size="16">
        <div class="stat-item">
          <span class="stat-label">总计:</span>
          <span class="stat-value">{{ stats.total }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">在线:</span>
          <span class="stat-value online">{{ stats.online }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">离线:</span>
          <span class="stat-value offline">{{ stats.offline }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">健康:</span>
          <span class="stat-value healthy">{{ stats.healthy }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">告警:</span>
          <span class="stat-value warning">{{ stats.warning }}</span>
        </div>
      </ElSpace>
    </div>
  </div>
</template>

<style scoped>
.server-filter-bar {
  background: #fff;
  padding: 16px;
  border-radius: 4px;
  margin-bottom: 16px;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.stats-bar {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.stat-label {
  color: #909399;
  font-size: 12px;
}

.stat-value {
  font-weight: 600;
  font-size: 14px;
}

.stat-value.online {
  color: #67c23a;
}

.stat-value.offline {
  color: #909399;
}

.stat-value.healthy {
  color: #409eff;
}

.stat-value.warning {
  color: #e6a23c;
}
</style>
