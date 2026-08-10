<script setup lang="ts">
import { computed, ref } from 'vue';
import type { Server } from '@/api/model';
import type { EnvType } from '../types/server.types';

interface Props {
  servers: Server[];
  loading?: boolean;
  selectedServers: Server[];
}

interface Emits {
  (e: 'selection-change', servers: Server[]): void;
  (e: 'connect', server: Server): void;
  (e: 'edit', server: Server): void;
  (e: 'sync-metrics', server: Server): void;
  (e: 'delete', server: Server): void;
  (e: 'deploy-agent', servers: Server[]): void;
  (e: 'uninstall-agent', servers: Server[]): void;
  (e: 'assign-groups', servers: Server[]): void;
  (e: 'batch-delete', servers: Server[]): void;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  selectedServers: () => []
});

const emit = defineEmits<Emits>();

const tableRef = ref();

// 环境显示信息
const getEnvDisplayInfo = (env: string) => {
  const envMap: Record<string, { text: string; type: 'success' | 'warning' | 'danger' | 'info' }> = {
    prod: { text: '生产', type: 'danger' },
    test: { text: '测试', type: 'warning' },
    dev: { text: '开发', type: 'success' }
  };
  return envMap[env] || { text: env || '未知', type: 'info' };
};

// 进度条颜色
const getProgressColor = (value: number) => {
  if (value > 80) return '#ff4d4f';
  if (value > 60) return '#faad14';
  return '#52c41a';
};

// 过滤后的服务器
const filteredServers = computed(() => props.servers);

// 选择变化
const handleSelectionChange = (servers: Server[]) => {
  emit('selection-change', servers);
};

// 过滤变化
const handleFilterChange = () => {
  // 可以在这里添加过滤逻辑
};

// 批量操作
const handleDeployAgent = () => {
  emit('deploy-agent', props.selectedServers);
};

const handleUninstallAgent = () => {
  emit('uninstall-agent', props.selectedServers);
};

const handleAssignGroups = () => {
  emit('assign-groups', props.selectedServers);
};

const handleBatchDelete = () => {
  emit('batch-delete', props.selectedServers);
};
</script>

<template>
  <div class="server-table-container">
    <!-- 批量操作栏 -->
    <div v-if="selectedServers.length > 0" class="batch-actions-bar">
      <ElAlert type="info" :closable="false">
        <template #title>已选择 {{ selectedServers.length }} 台服务器</template>
        <template #default>
          <div class="batch-actions">
            <ElButton size="small" @click="handleDeployAgent">批量部署Agent</ElButton>
            <ElButton size="small" @click="handleUninstallAgent">批量卸载Agent</ElButton>
            <ElButton size="small" @click="handleAssignGroups">分配分组</ElButton>
            <ElButton size="small" type="danger" @click="handleBatchDelete">批量删除</ElButton>
          </div>
        </template>
      </ElAlert>
    </div>

    <!-- 服务器表格 -->
    <ElTable
      ref="tableRef"
      v-loading="loading"
      :data="filteredServers"
      stripe
      border
      size="small"
      row-key="id"
      style="width: 100%"
      @selection-change="handleSelectionChange"
      @filter-change="handleFilterChange"
    >
      <!-- 选择列 -->
      <ElTableColumn type="selection" width="50" align="center" />

      <!-- 主机名 -->
      <ElTableColumn prop="hostname" label="主机名" min-width="120" show-overflow-tooltip />

      <!-- IP地址 -->
      <ElTableColumn prop="ip" label="IP地址" width="140" />

      <!-- 环境 -->
      <ElTableColumn prop="env" label="环境" width="80" align="center">
        <template #default="{ row }">
          <ElTag :type="getEnvDisplayInfo(row.env).type" size="small" effect="plain">
            {{ getEnvDisplayInfo(row.env).text }}
          </ElTag>
        </template>
      </ElTableColumn>

      <!-- Agent状态 -->
      <ElTableColumn label="Agent状态" width="100" align="center">
        <template #default="{ row }">
          <ElTag v-if="row.agentStatus === 'running'" type="success" size="small">运行中</ElTag>
          <ElTag v-else-if="row.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
          <ElTag v-else type="info" size="small">未安装</ElTag>
        </template>
      </ElTableColumn>

      <!-- 采集状态 -->
      <ElTableColumn label="采集状态" width="100" align="center">
        <template #default="{ row }">
          <ElTag v-if="row.metricsUpdatedAt" type="success" size="small">采集中</ElTag>
          <ElTag v-else-if="row.agentStatus === 'running'" type="warning" size="small">未采集</ElTag>
          <ElTag v-else type="info" size="small">-</ElTag>
        </template>
      </ElTableColumn>

      <!-- 状态 -->
      <ElTableColumn prop="status" label="状态" width="80" align="center">
        <template #default="{ row }">
          <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '在线' : '离线' }}
          </ElTag>
        </template>
      </ElTableColumn>

      <!-- CPU使用率 -->
      <ElTableColumn label="CPU" width="120" align="center">
        <template #default="{ row }">
          <ElProgress
            v-if="row.cpuUsage !== undefined"
            :percentage="Math.round(row.cpuUsage)"
            :color="getProgressColor(row.cpuUsage)"
            :stroke-width="4"
            :format="() => (row.cpuUsage ? row.cpuUsage.toFixed(1) + '%' : '0%')"
          />
          <span v-else>-</span>
        </template>
      </ElTableColumn>

      <!-- 内存使用率 -->
      <ElTableColumn label="内存" width="120" align="center">
        <template #default="{ row }">
          <ElProgress
            v-if="row.memoryUsage !== undefined"
            :percentage="Math.round(row.memoryUsage)"
            :color="getProgressColor(row.memoryUsage)"
            :stroke-width="4"
            :format="() => (row.memoryUsage ? row.memoryUsage.toFixed(1) + '%' : '0%')"
          />
          <span v-else>-</span>
        </template>
      </ElTableColumn>

      <!-- 磁盘使用率 -->
      <ElTableColumn label="磁盘" width="120" align="center">
        <template #default="{ row }">
          <ElProgress
            v-if="row.diskUsage !== undefined"
            :percentage="Math.round(row.diskUsage)"
            :color="getProgressColor(row.diskUsage)"
            :stroke-width="4"
            :format="() => (row.diskUsage ? row.diskUsage.toFixed(1) + '%' : '0%')"
          />
          <span v-else>-</span>
        </template>
      </ElTableColumn>

      <!-- 操作列 -->
      <ElTableColumn label="操作" width="200" align="center" fixed="right">
        <template #default="{ row }">
          <div class="action-buttons">
            <ElButton type="primary" link size="small" @click="$emit('connect', row)">连接</ElButton>
            <ElButton type="primary" link size="small" @click="$emit('edit', row)">编辑</ElButton>
            <ElButton type="primary" link size="small" @click="$emit('sync-metrics', row)">同步</ElButton>
            <ElButton type="danger" link size="small" @click="$emit('delete', row)">删除</ElButton>
          </div>
        </template>
      </ElTableColumn>
    </ElTable>
  </div>
</template>

<style scoped>
.server-table-container {
  height: 100%;
}

.batch-actions-bar {
  margin-bottom: 16px;
}

.batch-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}
</style>
