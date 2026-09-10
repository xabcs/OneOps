<script setup lang="ts">
  /**
   * 右侧服务器列表面板
   * 从 index.vue 拆分：搜索栏、表格、分页、批量操作下拉
   */

  import { computed, ref, watch } from 'vue';
  import CircleBadge from '@/components/common/CircleBadge.vue';
  import AlertBadge from './AlertBadge.vue';
  import MiniTrendChart from './MiniTrendChart.vue';
  import ServiceStatusIcon from './ServiceStatusIcon.vue';

  const props = defineProps<{
    loading: boolean;
    tableData: CMDB.Server[];
    total: number;
    selectedIds: number[];
    pagination: { page: number; pageSize: number };
    searchType: string;
    searchKeyword: string;
    serverTags: CMDB.ServerTag[];
    attributeDefinitions: Api.SystemManage.AttributeDefinition[];
    getEnvDisplayInfo: (env?: string) => { label: string; type: string };
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
    (e: 'attribute-filter', filters: Record<string, string>): void;
    (e: 'tag-filter', tagId: number | undefined): void;
  }>();

  // 属性筛选
  const attrFilterVisible = ref(false);
  const selectedAttrKey = ref('');
  const selectedAttrValue = ref('');

  // 标签筛选
  const selectedTagId = ref<number | undefined>(undefined);

  function getSelectableAttrs() {
    return props.attributeDefinitions;
  }

  function parseOptions(optionsStr: string) {
    if (!optionsStr) return [];
    try {
      return JSON.parse(optionsStr);
    } catch {
      return [];
    }
  }

  const selectedAttrDef = computed(() => {
    return getSelectableAttrs().find(a => a.key === selectedAttrKey.value);
  });

  function applyAttrFilter() {
    if (selectedAttrKey.value && selectedAttrValue.value) {
      emit('attribute-filter', { [selectedAttrKey.value]: selectedAttrValue.value });
    }
    attrFilterVisible.value = false;
  }

  function clearAttrFilter() {
    selectedAttrKey.value = '';
    selectedAttrValue.value = '';
    emit('attribute-filter', {});
    attrFilterVisible.value = false;
  }

  // 标签筛选变化时通知父组件
  watch(selectedTagId, val => {
    emit('tag-filter', val);
  });

  const groupTypes = ['primary', 'success', 'warning', 'danger'] as const;
  function groupType(id: number) {
    return groupTypes[id % groupTypes.length];
  }
</script>

<template>
  <div style="min-width: 0; height: 100%; display: flex; flex-direction: column; flex: 1; overflow: hidden">
    <ElCard
      shadow="never"
      body-style="padding: 12px; border-radius: 0; flex: 1; display: flex; flex-direction: column; overflow: hidden;"
      style="border-radius: 0; flex: 1; display: flex; flex-direction: column; overflow: hidden"
    >
      <!-- 头部：搜索和操作按钮 -->
      <div class="mb-8px flex items-center justify-between gap-12px">
        <div class="flex items-center gap-8px">
          <ElDropdown trigger="click" @command="emit('create-command', $event)">
            <PermissionButton code="cmdb.server.create" type="primary">
              <template #icon><icon-ic-round-plus class="text-icon" /></template>
              创建
              <icon-ic-round-keyboard-arrow-down class="ml-4px text-icon" />
            </PermissionButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem v-permission="'cmdb.server.create'" command="add">新增主机</ElDropdownItem>
                <ElDropdownItem command="import">批量导入</ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
          <ElDropdown trigger="click" :disabled="selectedIds.length === 0" @command="emit('batch-command', $event)">
            <ElButton plain>
              更多操作
              <icon-ic-round-keyboard-arrow-down class="ml-4px text-icon" />
            </ElButton>
            <template #dropdown>
              <ElDropdownMenu>
                <ElDropdownItem
                  v-permission="'cmdb.server.create'"
                  command="test-connection"
                  :disabled="selectedIds.length === 0"
                >
                  测试连接 ({{ selectedIds.length }})
                </ElDropdownItem>
                <ElDropdownItem
                  v-permission="'cmdb.agents.deploy'"
                  command="batch-deploy"
                  :disabled="selectedIds.length === 0"
                >
                  批量部署 ({{ selectedIds.length }})
                </ElDropdownItem>
                <ElDropdownItem
                  v-permission="'cmdb.agents.uninstall'"
                  command="batch-uninstall"
                  :disabled="selectedIds.length === 0"
                >
                  批量卸载 ({{ selectedIds.length }})
                </ElDropdownItem>
                <ElDropdownItem
                  v-permission="'cmdb.server.delete'"
                  command="batch-delete"
                  :disabled="selectedIds.length === 0"
                  style="color: #f56c6c"
                >
                  批量删除 ({{ selectedIds.length }})
                </ElDropdownItem>
              </ElDropdownMenu>
            </template>
          </ElDropdown>
        </div>

        <div :style="{ display: 'flex', alignItems: 'center', gap: '8px' }">
          <ElSelect
            :model-value="searchType"
            placeholder="筛选条件"
            style="width: 120px"
            @update:model-value="emit('update:searchType', $event)"
          >
            <ElOption label="主机名" value="hostname" />
            <ElOption label="IP地址" value="ip" />
            <ElOption label="分组" value="group" />
          </ElSelect>
          <ElInput
            :model-value="searchKeyword"
            placeholder="搜索"
            clearable
            style="width: 200px"
            @update:model-value="emit('update:searchKeyword', $event)"
            @keyup.enter="emit('search')"
          >
            <template #suffix>
              <icon-ic-round-search class="cursor-pointer text-icon" @click="emit('search')" />
            </template>
          </ElInput>
          <ElButton text @click="emit('refresh')">
            <icon-mdi-refresh class="text-18px" :class="{ 'animate-spin': loading }" />
          </ElButton>
          <ElPopover v-model:visible="attrFilterVisible" placement="bottom" :width="280" trigger="click">
            <template #reference>
              <ElButton text>
                <icon-mdi-filter-variant class="text-18px" :class="{ 'text-primary': selectedAttrKey }" />
              </ElButton>
            </template>
            <div class="flex flex-col gap-12px">
              <div class="text-14px font-500">按属性筛选</div>
              <ElSelect v-model="selectedAttrKey" placeholder="选择属性" clearable style="width: 100%">
                <ElOption v-for="attr in getSelectableAttrs()" :key="attr.id" :label="attr.name" :value="attr.key" />
              </ElSelect>
              <ElSelect
                v-if="selectedAttrDef?.type === 'select'"
                v-model="selectedAttrValue"
                placeholder="选择值"
                clearable
                style="width: 100%"
              >
                <ElOption
                  v-for="opt in parseOptions(selectedAttrDef.options)"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </ElSelect>
              <ElInput v-else v-model="selectedAttrValue" placeholder="输入属性值" clearable />
              <div class="flex justify-end gap-8px">
                <ElButton size="small" @click="clearAttrFilter">清除</ElButton>
                <ElButton size="small" type="primary" @click="applyAttrFilter">筛选</ElButton>
              </div>
            </div>
          </ElPopover>
          <ElSelect v-model="selectedTagId" placeholder="按标签筛选" clearable size="small" style="width: 140px">
            <ElOption v-for="tag in serverTags" :key="tag.id" :label="tag.name" :value="tag.id">
              <span>{{ tag.name }}</span>
              <span
                :style="{
                  display: 'inline-block',
                  width: '8px',
                  height: '8px',
                  borderRadius: '50%',
                  backgroundColor: tag.color,
                  marginLeft: '8px'
                }"
              />
            </ElOption>
          </ElSelect>
        </div>
      </div>

      <!-- 主机列表容器 -->
      <div class="flex flex-col flex-1 overflow-hidden bg-white">
        <div class="flex-1 overflow-auto">
          <ElTable
            v-loading="loading"
            height="100%"
            :data="tableData"
            size="small"
            :row-style="{ height: '48px' }"
            :cell-style="{ padding: '0', borderRight: 'none' }"
            :header-cell-style="{ backgroundColor: '#f5f7fa', borderRight: 'none' }"
            table-layout="fixed"
            @selection-change="emit('selection-change', $event)"
            @select-all="emit('select-all', $event)"
          >
            <ElTableColumn type="selection" width="50" align="center" />
            <ElTableColumn prop="hostname" label="主机名" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">
                <span class="cursor-pointer text-primary hover:underline" @click="emit('view-detail', row)">
                  {{ row.hostname }}
                </span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="IP地址" min-width="150" show-overflow-tooltip>
              <template #default="{ row }">
                <div :style="{ display: 'flex', flexDirection: 'column', gap: '2px' }">
                  <div
                    :style="{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '6px',
                      fontSize: '12px',
                      color: '#606266',
                      fontFamily: 'ui-monospace, SF Mono, Menlo, Monaco, Consolas, Courier New, monospace'
                    }"
                  >
                    <span>{{ row.ip }}</span>
                    <CircleBadge type="primary">外</CircleBadge>
                  </div>
                  <div
                    v-if="row.innerIp"
                    :style="{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '6px',
                      fontSize: '12px',
                      color: '#606266',
                      fontFamily: 'ui-monospace, SF Mono, Menlo, Monaco, Consolas, Courier New, monospace'
                    }"
                  >
                    <span>{{ row.innerIp }}</span>
                    <CircleBadge type="success">内</CircleBadge>
                  </div>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn label="配置" width="110" align="center">
              <template #default="{ row }">
                <span class="text-12px" style="color: #606266">{{ row.cpu }}C/{{ row.memory }}G</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="环境" width="75" align="center">
              <template #default="{ row }">
                <ElTag
                  v-if="row.attributeValues?.env"
                  :type="getEnvDisplayInfo(row.attributeValues?.env).type"
                  size="small"
                  effect="dark"
                  round
                >
                  {{ getEnvDisplayInfo(row.attributeValues?.env).label }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="分组" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">
                <ElTag
                  v-if="row.groups && row.groups.length > 0"
                  :type="groupType(row.groups[0].id)"
                  size="small"
                  effect="dark"
                >
                  {{ row.groups[0].name }}
                  <span v-if="row.groups.length > 1" class="ml-4px">+{{ row.groups.length - 1 }}</span>
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="标签" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">
                <div v-if="row.tags && row.tags.length > 0" class="flex flex-wrap gap-2px">
                  <ElTag
                    v-for="tag in row.tags.slice(0, 3)"
                    :key="tag.id"
                    class="custom-color-tag"
                    size="small"
                    effect="dark"
                    round
                    :style="{ backgroundColor: tag.color, borderColor: tag.color }"
                  >
                    {{ tag.name }}
                  </ElTag>
                  <span v-if="row.tags.length > 3" class="text-12px text-gray-400">+{{ row.tags.length - 3 }}</span>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn label="资源使用率" min-width="200" align="center">
              <template #default="{ row }">
                <template v-if="row.agentStatus === 'offline'">
                  <ElTag type="warning" size="small">Agent离线</ElTag>
                </template>
                <template v-else-if="!['running', 'active'].includes(row.agentStatus) || !row.agentStatus">
                  <ElTag type="info" size="small">未安装</ElTag>
                </template>
                <template v-else-if="row.metricsUpdatedAt">
                  <ElTooltip :content="formatDiskPartitions(row)" placement="top">
                    <div
                      :style="{
                        display: 'inline-flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '6px',
                        fontSize: '12px',
                        whiteSpace: 'nowrap',
                        width: '100%'
                      }"
                    >
                      <span :style="{ display: 'inline-flex', alignItems: 'center', gap: '4px' }">
                        <span :style="{ color: '#909399', fontSize: '11px' }">CPU</span>
                        <span
                          :style="{
                            color: getUsageColor(row.cpuUsage),
                            fontWeight: 600,
                            fontSize: '12px',
                            minWidth: '32px',
                            textAlign: 'center'
                          }"
                        >
                          {{ Math.round(row.cpuUsage || 0) }}%
                        </span>
                      </span>
                      <span :style="{ color: '#dcdfe6', margin: '0 2px' }">|</span>
                      <span :style="{ display: 'inline-flex', alignItems: 'center', gap: '4px' }">
                        <span :style="{ color: '#909399', fontSize: '11px' }">内存</span>
                        <span
                          :style="{
                            color: getUsageColor(row.memoryUsage),
                            fontWeight: 600,
                            fontSize: '12px',
                            minWidth: '32px',
                            textAlign: 'center'
                          }"
                        >
                          {{ Math.round(row.memoryUsage || 0) }}%
                        </span>
                      </span>
                      <span :style="{ color: '#dcdfe6', margin: '0 2px' }">|</span>
                      <span :style="{ display: 'inline-flex', alignItems: 'center', gap: '4px' }">
                        <span :style="{ color: '#909399', fontSize: '11px' }">磁盘</span>
                        <span
                          :style="{
                            color: getUsageColor(getMaxDiskPartition(row).usage),
                            fontWeight: 600,
                            fontSize: '12px',
                            minWidth: '32px',
                            textAlign: 'center'
                          }"
                        >
                          {{ Math.round(getMaxDiskPartition(row).usage) }}%
                        </span>
                        <span :style="{ fontSize: '10px', color: '#909399', marginLeft: '2px' }">
                          ({{ getMaxDiskPartition(row).mount }})
                        </span>
                      </span>
                    </div>
                  </ElTooltip>
                </template>
                <template v-else>
                  <ElTooltip content="Agent运行中，等待首次采集" placement="top">
                    <ElTag type="success" size="small">采集中</ElTag>
                  </ElTooltip>
                </template>
              </template>
            </ElTableColumn>
            <ElTableColumn label="CPU趋势" width="100" align="center">
              <template #default="{ row }">
                <MiniTrendChart
                  v-if="row.cpuTrend && row.cpuTrend.length > 0"
                  :data="row.cpuTrend"
                  :height="30"
                  :color="getUsageColor(row.cpuUsage || 0)"
                />
                <span v-else class="text-12px text-gray-400">-</span>
              </template>
            </ElTableColumn>
            <ElTableColumn label="服务状态" width="100" align="center">
              <template #default="{ row }">
                <ServiceStatusIcon
                  v-if="row.agentStatus === 'running'"
                  :status="row.serviceStatus || 'unknown'"
                  :show-text="true"
                />
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
            <ElTableColumn label="操作" width="100" align="center" fixed="right" class-name="msre-table-actions">
              <template #default="{ row }">
                <PermissionButton
                  code="cmdb.server.connect"
                  link
                  type="primary"
                  size="small"
                  title="连接终端"
                  @click="emit('connect', row)"
                >
                  <icon-lucide-terminal class="text-14px" />
                </PermissionButton>
                <ElDropdown trigger="click" @command="emit('more-action', $event, row)">
                  <ElButton link type="primary" size="small" title="更多操作">
                    <icon-lucide-ellipsis-vertical class="text-14px" />
                  </ElButton>
                  <template #dropdown>
                    <ElDropdownMenu>
                      <ElDropdownItem
                        v-if="row.agentStatus === 'running' || row.agentStatus === 'failed'"
                        command="sync-metrics"
                      >
                        <icon-mdi-refresh class="mr-8px" />
                        刷新指标
                      </ElDropdownItem>
                      <ElDropdownItem v-if="row.agentStatus === 'running'" command="monitoring">
                        <icon-mdi-chart-line class="mr-8px" />
                        监控详情
                      </ElDropdownItem>
                      <ElDropdownItem command="detail">
                        <icon-ic-round-info class="mr-8px" />
                        主机详情
                      </ElDropdownItem>
                      <ElDropdownItem v-permission="'cmdb.server.update'" command="edit">
                        <icon-ic-round-edit class="mr-8px" />
                        编辑
                      </ElDropdownItem>
                      <ElDropdownItem
                        v-if="!row.agentStatus || row.agentStatus === 'uninstalled' || row.agentStatus === 'failed'"
                        v-permission="'cmdb.agents.deploy'"
                        command="agent-deploy"
                      >
                        <icon-mdi-download class="mr-8px" />
                        部署 Agent
                      </ElDropdownItem>
                      <ElDropdownItem
                        v-if="row.agentStatus === 'running' || row.agentStatus === 'offline'"
                        v-permission="'cmdb.agents.restart'"
                        command="agent-restart"
                      >
                        <icon-mdi-restart class="mr-8px" />
                        重启 Agent
                      </ElDropdownItem>
                      <ElDropdownItem
                        v-if="row.agentStatus === 'running' || row.agentStatus === 'offline'"
                        v-permission="'cmdb.agents.uninstall'"
                        command="agent-uninstall"
                      >
                        <icon-mdi-delete-forever class="mr-8px" />
                        卸载 Agent
                      </ElDropdownItem>
                      <ElDropdownItem
                        v-permission="'cmdb.server.delete'"
                        divided
                        command="delete"
                        style="color: #f56c6c"
                      >
                        <icon-ic-round-delete class="mr-8px" />
                        删除主机
                      </ElDropdownItem>
                    </ElDropdownMenu>
                  </template>
                </ElDropdown>
              </template>
            </ElTableColumn>
          </ElTable>
        </div>

        <div v-if="tableData.length > 0" class="flex justify-end border-t border-gray-200 bg-white p-12px">
          <ElPagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="total"
            layout="total, sizes, prev, pager, next"
            @current-change="emit('page-change', $event)"
            @size-change="emit('page-size-change', $event)"
          />
        </div>
      </div>
    </ElCard>
  </div>
</template>
