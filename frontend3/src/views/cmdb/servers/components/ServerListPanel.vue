<script setup lang="ts">
/**
 * 右侧服务器列表面板
 * 从 index.vue 拆分：搜索栏、表格、分页、批量操作下拉
 */

import { MiniTrendChart, ServiceStatusIcon, AlertBadge } from '@/components/features/monitoring/MonitoringComponents';

const props = defineProps<{
  loading: boolean;
  tableData: CMDB.Server[];
  total: number;
  selectedIds: number[];
  pagination: { page: number; pageSize: number };
  searchType: string;
  searchKeyword: string;
  getEnvDisplayInfo: (env: string) => { label: string; type: string };
  getUsageColor: (val: number) => string;
  getMaxDiskPartition: (row: CMDB.Server) => { usage: number; mount: string };
  formatDiskPartitions: (row: CMDB.Server) => string;
}>();

const emit = defineEmits<{
  (e: 'update:searchType', val: string): void;
  (e: 'update:searchKeyword', val: string): void;
  (e: 'search'): void;
  (e: 'refresh'): void;
  (e: 'selection-change', selection: CMDB.Server[]): void;
  (e: 'select-all', selection: CMDB.Server[]): void;
  (e: 'page-change', page: number): void;
  (e: 'page-size-change', size: number): void;
  (e: 'view-detail', row: CMDB.Server): void;
  (e: 'connect', row: CMDB.Server): void;
  (e: 'more-action', cmd: string, row: CMDB.Server): void;
  (e: 'batch-command', cmd: string): void;
  (e: 'create-command', cmd: string): void;
}>();
</script>

<template>
  <div class="min-w-0 flex flex-col flex-1">
    <!-- 头部：搜索和操作按钮 -->
    <div class="mb-8px flex items-center justify-between gap-12px">
      <div class="flex items-center gap-8px">
        <ElDropdown trigger="click" @command="emit('create-command', $event)">
          <ElButton type="primary" :style="{ backgroundColor: '#0052D9', borderColor: '#0052D9', color: '#fff', borderRadius: '0', fontSize: '12px', height: '30px' }">
            <template #icon><icon-ic-round-plus class="text-icon" /></template>
            创建
            <icon-ic-round-keyboard-arrow-down class="ml-4px text-icon" />
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem command="add">新增主机</ElDropdownItem>
              <ElDropdownItem command="import">批量导入</ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
        <ElDropdown trigger="click" :disabled="selectedIds.length === 0" @command="emit('batch-command', $event)">
          <ElButton plain :style="{ borderRadius: '0', fontSize: '12px', height: '30px' }">
            更多操作
            <icon-ic-round-keyboard-arrow-down class="ml-4px text-icon" />
          </ElButton>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem command="test-connection" :disabled="selectedIds.length === 0">测试连接 ({{ selectedIds.length }})</ElDropdownItem>
              <ElDropdownItem command="batch-deploy" :disabled="selectedIds.length === 0">批量部署 ({{ selectedIds.length }})</ElDropdownItem>
              <ElDropdownItem command="batch-uninstall" :disabled="selectedIds.length === 0">批量卸载 ({{ selectedIds.length }})</ElDropdownItem>
              <ElDropdownItem command="batch-delete" :disabled="selectedIds.length === 0" style="color: #f56c6c">批量删除 ({{ selectedIds.length }})</ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </div>

      <div class="search-inputs flex items-center gap-8px">
        <ElSelect :model-value="searchType" placeholder="筛选条件" @update:model-value="emit('update:searchType', $event)">
          <ElOption label="主机名" value="hostname" />
          <ElOption label="IP地址" value="ip" />
          <ElOption label="分组" value="group" />
        </ElSelect>
        <ElInput :model-value="searchKeyword" placeholder="搜索" clearable style="width: 200px" @update:model-value="emit('update:searchKeyword', $event)" @keyup.enter="emit('search')">
          <template #suffix><icon-ic-round-search class="cursor-pointer text-icon" @click="emit('search')" /></template>
        </ElInput>
        <ElButton text @click="emit('refresh')">
          <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
        </ElButton>
      </div>
    </div>

    <!-- 主机列表容器 -->
    <div class="flex flex-col flex-1 overflow-hidden bg-white">
      <div class="flex-1 overflow-auto">
        <ElTable v-loading="loading" height="100%" :data="tableData" size="small" class="server-list-table compact-table" :row-style="{ height: '48px' }" :cell-style="{ padding: '0', borderRight: 'none' }" :header-cell-style="{ backgroundColor: '#f5f7fa', borderRight: 'none' }" table-layout="fixed" @selection-change="emit('selection-change', $event)" @select-all="emit('select-all', $event)">
          <ElTableColumn type="selection" width="50" align="center" />
          <ElTableColumn prop="hostname" label="主机名" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="cursor-pointer text-primary hover:underline" @click="emit('view-detail', row)">{{ row.hostname }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="IP地址" min-width="150" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="ip-list">
                <div class="ip-row">{{ row.ip }}<span class="ip-tag-outer">外</span></div>
                <div v-if="row.innerIp" class="ip-row ip-row-inner">{{ row.innerIp }}<span class="ip-tag-inner">内</span></div>
              </div>
            </template>
          </ElTableColumn>
          <ElTableColumn label="配置" width="110" align="center">
            <template #default="{ row }"><span class="text-12px" style="color: #606266">{{ row.cpu }}C/{{ row.memory }}G</span></template>
          </ElTableColumn>
          <ElTableColumn label="环境" width="75" align="center">
            <template #default="{ row }">
              <ElTag :type="getEnvDisplayInfo(row.env).type" size="small" effect="plain">{{ getEnvDisplayInfo(row.env).label }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="分组" min-width="120" show-overflow-tooltip>
            <template #default="{ row }">
              <ElTag v-if="row.groups && row.groups.length > 0" :style="{ color: row.groups[0].color, borderColor: row.groups[0].color }" size="small">{{ row.groups[0].name }}<span v-if="row.groups.length > 1" class="ml-4px">+{{ row.groups.length - 1 }}</span></ElTag>
              <span v-else class="text-12px text-gray-400">-</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="资源使用率" min-width="200" align="center">
            <template #default="{ row }">
              <template v-if="row.agentStatus === 'offline'"><ElTag type="warning" size="small">Agent离线</ElTag></template>
              <template v-else-if="!['running', 'active'].includes(row.agentStatus) || !row.agentStatus"><ElTag type="info" size="small">未安装</ElTag></template>
              <template v-else-if="row.metricsUpdatedAt">
                <ElTooltip :content="formatDiskPartitions(row)" placement="top">
                  <div class="usage-compact">
                    <span class="usage-item" :style="{ color: getUsageColor(row.cpuUsage) }"><span class="usage-label">CPU</span><span class="usage-value">{{ Math.round(row.cpuUsage || 0) }}%</span></span>
                    <span class="usage-divider">|</span>
                    <span class="usage-item" :style="{ color: getUsageColor(row.memoryUsage) }"><span class="usage-label">内存</span><span class="usage-value">{{ Math.round(row.memoryUsage || 0) }}%</span></span>
                    <span class="usage-divider">|</span>
                    <span class="usage-item disk-usage">
                      <span class="usage-label">磁盘</span>
                      <span class="usage-value" :style="{ color: getUsageColor(getMaxDiskPartition(row).usage) }">{{ Math.round(getMaxDiskPartition(row).usage) }}%</span>
                      <span class="disk-mount">({{ getMaxDiskPartition(row).mount }})</span>
                    </span>
                  </div>
                </ElTooltip>
              </template>
              <template v-else>
                <ElTooltip content="Agent运行中，等待首次采集" placement="top"><ElTag type="success" size="small">采集中</ElTag></ElTooltip>
              </template>
            </template>
          </ElTableColumn>
          <ElTableColumn label="CPU趋势" width="100" align="center">
            <template #default="{ row }">
              <MiniTrendChart v-if="row.cpuTrend && row.cpuTrend.length > 0" :data="row.cpuTrend" :height="30" :color="getUsageColor(row.cpuUsage || 0)" />
              <span v-else class="text-12px text-gray-400">-</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="服务状态" width="100" align="center">
            <template #default="{ row }">
              <ServiceStatusIcon v-if="row.agentStatus === 'running'" :status="row.serviceStatus || 'unknown'" :show-text="true" />
              <ElTag v-else-if="row.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
              <ElTag v-else type="info" size="small">未安装</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="告警" width="80" align="center">
            <template #default="{ row }">
              <AlertBadge v-if="row.agentStatus === 'running'" :count="row.alertCount || 0" :max-count="99" />
              <span v-else class="text-12px text-gray-400">-</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="100" align="center" fixed="right">
            <template #default="{ row }">
              <div style="display: flex; align-items: center; justify-content: center; gap: 12px">
                <a title="连接终端" style="cursor: pointer" @click="emit('connect', row)"><icon-lucide-terminal class="text-14px" style="color: #909399" /></a>
                <ElDropdown trigger="click" @command="emit('more-action', $event, row)">
                  <span style="display: inline-flex; align-items: center; justify-content: center; width: 24px; height: 24px; cursor: pointer; font-size: 16px; color: #909399; font-style: normal; letter-spacing: 1px">⋮</span>
                  <template #dropdown>
                    <ElDropdownMenu>
                      <ElDropdownItem v-if="row.agentStatus === 'running' || row.agentStatus === 'failed'" command="sync-metrics"><icon-mdi-refresh class="mr-8px" />刷新指标</ElDropdownItem>
                      <ElDropdownItem v-if="row.agentStatus === 'running'" command="monitoring"><icon-mdi-chart-line class="mr-8px" />监控详情</ElDropdownItem>
                      <ElDropdownItem command="detail"><icon-ic-round-info class="mr-8px" />主机详情</ElDropdownItem>
                      <ElDropdownItem command="edit"><icon-ic-round-edit class="mr-8px" />编辑</ElDropdownItem>
                      <ElDropdownItem v-if="!row.agentStatus || row.agentStatus === 'uninstalled' || row.agentStatus === 'failed'" command="agent-deploy"><icon-mdi-download class="mr-8px" />部署 Agent</ElDropdownItem>
                      <ElDropdownItem v-if="row.agentStatus === 'running' || row.agentStatus === 'offline'" command="agent-restart"><icon-mdi-restart class="mr-8px" />重启 Agent</ElDropdownItem>
                      <ElDropdownItem v-if="row.agentStatus === 'running' || row.agentStatus === 'offline'" command="agent-uninstall"><icon-mdi-delete-forever class="mr-8px" />卸载 Agent</ElDropdownItem>
                      <ElDropdownItem divided command="delete" style="color: #f56c6c"><icon-ic-round-delete class="mr-8px" />删除主机</ElDropdownItem>
                    </ElDropdownMenu>
                  </template>
                </ElDropdown>
              </div>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>

      <div v-if="tableData.length > 0" class="flex justify-end border-t border-gray-200 bg-white p-12px">
        <ElPagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize" :page-sizes="[10, 20, 50, 100]" :total="total" layout="total, sizes, prev, pager, next" @current-change="emit('page-change', $event)" @size-change="emit('page-size-change', $event)" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-inputs :deep(.el-select__wrapper) { border-radius: 0 !important; }
.search-inputs :deep(.el-input__wrapper) { border-radius: 0 !important; }
.usage-compact { display: inline-flex; align-items: center; justify-content: center; gap: 6px; font-size: 12px; white-space: nowrap; width: 100%; }
.usage-item { display: inline-flex; align-items: center; gap: 4px; }
.usage-label { color: #909399; font-size: 11px; }
.usage-value { font-weight: 600; font-size: 12px; min-width: 32px; text-align: center; }
.usage-divider { color: #dcdfe6; margin: 0 2px; }
.disk-mount { font-size: 10px; color: #909399; margin-left: 2px; }
.ip-list { display: flex; flex-direction: column; gap: 2px; }
.ip-row { font-size: 12px; color: #606266; font-family: ui-monospace, 'SF Mono', Menlo, Monaco, 'Cascadia Code', 'Roboto Mono', 'Consolas', 'Courier New', monospace; font-weight: 400; line-height: 1.3; }
.ip-row-inner { color: #606266; }
.ip-tag-outer, .ip-tag-inner { font-size: 10px; color: #909399; margin-left: 6px; opacity: 0.6; font-weight: normal; letter-spacing: 0.5px; }
</style>
