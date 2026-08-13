<script setup lang="ts">
  /**
   * 主机详情抽屉
   * 从 index.vue 拆分：查看主机详情的抽屉（概览/连接信息/会话记录/监控信息）
   */

  import { ref, watch } from 'vue';
  import { fetchCheckConnectPermission, fetchGetSessions, fetchGetServerAttributes, fetchGetAttributes } from '@/service/api';
  import AttributeFormItems from './AttributeFormItems.vue';

  const props = defineProps<{
    visible: boolean;
    server: CMDB.Server | null;
    getUsageColor: (val: number) => string;
    getEnvDisplayInfo: (env?: string) => { label: string; type: string };
    formatTime: (time: string) => string;
    formatDiskPartitionsForDrawer: (server: CMDB.Server | null) => Array<{ mount: string; usage: number }>;
    handleConnect: (server: CMDB.Server) => void;
  }>();

  const emit = defineEmits<{
    (e: 'update:visible', val: boolean): void;
  }>();

  const drawerActiveTab = ref('overview');
  const drawerSessions = ref<Bastion.BastionSession[]>([]);
  const drawerPermission = ref<{ credentials: CMDB.SSHCredential[] }>({ credentials: [] });
  const drawerLoading = ref(false);

  // 扩展属性
  const detailAttributes = ref<Api.SystemManage.AttributeDefinition[]>([]);
  const detailServerAttributes = ref<System.ServerAttribute[]>([]);
  const detailAttrLoading = ref(false);

  function getDetailAttributeValue(id: number): string {
    const attr = detailServerAttributes.value.find(a => a.attributeId === id);
    return attr?.attributeValue || '';
  }

  function parseDetailAttributeOptions(optionsStr: string) {
    if (!optionsStr) return [];
    try {
      return JSON.parse(optionsStr);
    } catch {
      return [];
    }
  }

  function getDetailUnifiedAttributes() {
    return detailAttributes.value;
  }

  watch(
    () => props.visible,
    async val => {
      if (val && props.server) {
        drawerActiveTab.value = 'overview';
        await loadDrawerData();
        await loadDetailAttributes();
      }
    }
  );

  async function loadDrawerData() {
    if (!props.server) return;
    drawerLoading.value = true;
    try {
      const [sessionRes, permRes] = await Promise.all([
        fetchGetSessions(props.server.id).catch(() => ({ data: [] })),
        fetchCheckConnectPermission(props.server.id).catch(() => ({ data: { credentials: [] } }))
      ]);
      drawerSessions.value = sessionRes.data || [];
      drawerPermission.value = permRes.data || { credentials: [] };
    } catch {
      /* ignore */
    } finally {
      drawerLoading.value = false;
    }
  }

  async function loadDetailAttributes() {
    if (!props.server) return;
    detailAttrLoading.value = true;
    try {
      const [attrRes, serverAttrRes] = await Promise.all([
        fetchGetAttributes().catch(() => ({ data: [] })),
        fetchGetServerAttributes(props.server.id).catch(() => ({ data: [] }))
      ]);
      detailAttributes.value = (attrRes.data || []) as Api.SystemManage.AttributeDefinition[];
      detailServerAttributes.value = (serverAttrRes.data || []) as System.ServerAttribute[];
    } catch {
      /* ignore */
    } finally {
      detailAttrLoading.value = false;
    }
  }
</script>

<template>
  <ElDrawer
    :model-value="visible"
    direction="rtl"
    size="820px"
    :title="server?.hostname || '主机详情'"
    destroy-on-close
    @close="emit('update:visible', false)"
  >
    <div v-if="server">
      <ElTabs v-model="drawerActiveTab">
        <!-- 概览 Tab -->
        <ElTabPane label="概览" name="overview">
          <div class="overview-section">
            <div class="section-title">基础信息</div>
            <ElDescriptions :column="2" border size="small">
              <ElDescriptionsItem label="主机名">{{ server.hostname }}</ElDescriptionsItem>
              <ElDescriptionsItem label="连接IP">{{ server.ip }}</ElDescriptionsItem>
              <ElDescriptionsItem label="内网IP">{{ server.innerIp || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="SSH端口">{{ server.sshPort || 22 }}</ElDescriptionsItem>
              <ElDescriptionsItem label="操作系统">{{ server.os || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="系统架构">{{ server.arch || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="环境">
                <ElTag :type="getEnvDisplayInfo(server.attributeValues?.env).type" size="small">
                  {{ getEnvDisplayInfo(server.attributeValues?.env).label }}
                </ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="主机状态">
                <ElTag :type="server.status === 1 ? 'success' : 'info'" size="small">
                  {{ server.status === 1 ? '正常' : '停用' }}
                </ElTag>
              </ElDescriptionsItem>
            </ElDescriptions>
          </div>
          <div class="overview-section">
            <div class="section-title">硬件配置</div>
            <ElDescriptions :column="3" border size="small">
              <ElDescriptionsItem label="CPU">{{ server.cpu ? `${server.cpu} 核` : '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="内存">{{ server.memory ? `${server.memory} GB` : '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="磁盘">{{ server.disk ? `${server.disk} GB` : '-' }}</ElDescriptionsItem>
            </ElDescriptions>
          </div>
          <div v-if="server.serverType === 'cloud' && server.cloudInfo" class="overview-section">
            <div class="section-title">云主机信息</div>
            <ElDescriptions :column="2" border size="small">
              <ElDescriptionsItem label="服务商">
                <ElTag size="small">{{ server.provider || '-' }}</ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="实例类型">{{ server.cloudInfo.instanceType || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="区域">{{ server.region || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="可用区">{{ server.zone || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="实例ID" :span="2">{{ server.cloudInfo.instanceId || '-' }}</ElDescriptionsItem>
            </ElDescriptions>
          </div>
          <div class="overview-section">
            <div class="section-title">归属信息</div>
            <ElDescriptions :column="2" border size="small">
              <ElDescriptionsItem label="业务系统">{{ server.business?.name || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="所属分组">
                <ElTag v-for="group in server.groups" :key="group.id" size="small" style="margin-right: 4px">
                  {{ group.name }}
                </ElTag>
                <span v-if="!server.groups?.length" class="text-gray-400">未分组</span>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="标签" :span="2">
                <ElTag
                  v-for="tag in server.tags"
                  :key="tag.id"
                  size="small"
                  :color="tag.color"
                  style="margin-right: 4px"
                >
                  {{ tag.name }}
                </ElTag>
                <span v-if="!server.tags?.length" class="text-gray-400">无标签</span>
              </ElDescriptionsItem>
            </ElDescriptions>
          </div>
          <div class="overview-section">
            <div class="section-title">状态信息</div>
            <ElDescriptions :column="2" border size="small">
              <ElDescriptionsItem label="Agent 状态">
                <ElTag v-if="server.agentStatus === 'running'" type="success" size="small">
                  运行中 (v{{ server.agentVersion || '-' }})
                </ElTag>
                <ElTag v-else-if="server.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
                <ElTag v-else type="info" size="small">未安装</ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="最近连接">{{ server.lastConnectTime || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="最后连通性检查">{{ server.lastCheckTime || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="连通性状态">
                <ElTag
                  :type="
                    server.connectivityStatus === 'online'
                      ? 'success'
                      : server.connectivityStatus === 'offline'
                        ? 'danger'
                        : 'info'
                  "
                  size="small"
                >
                  {{
                    server.connectivityStatus === 'online'
                      ? '在线'
                      : server.connectivityStatus === 'offline'
                        ? '离线'
                        : '未知'
                  }}
                </ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="创建时间" :span="2">{{ server.createdAt || '-' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="备注" :span="2">{{ server.remarks || '-' }}</ElDescriptionsItem>
            </ElDescriptions>
          </div>
          <div class="overview-section">
            <div class="section-title">扩展属性</div>
            <AttributeFormItems
              :attributes="getDetailUnifiedAttributes()"
              :loading-attributes="detailAttrLoading"
              :get-attribute-value="getDetailAttributeValue"
              :parse-attribute-options="parseDetailAttributeOptions"
              readonly
            />
          </div>
        </ElTabPane>

        <!-- 连接 Tab -->
        <ElTabPane label="连接信息" name="connect">
          <ElDescriptions :column="1" border>
            <ElDescriptionsItem label="SSH端口">{{ server.sshPort || 22 }}</ElDescriptionsItem>
            <ElDescriptionsItem label="绑定凭证">
              <div v-if="server.credentials?.length">
                <ElTag
                  v-for="cred in server.credentials"
                  :key="cred.id"
                  type="success"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 4px"
                >
                  {{ cred.name }}（{{ cred.username }}）
                </ElTag>
              </div>
              <ElTag v-else type="warning" size="small">未配置</ElTag>
            </ElDescriptionsItem>
            <ElDescriptionsItem label="可用凭证">
              <div>
                <ElTag
                  v-for="cred in drawerPermission.credentials"
                  :key="cred.id"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 4px"
                >
                  {{ cred.name }}（{{ cred.username }}）
                </ElTag>
                <span v-if="!drawerPermission.credentials?.length" class="text-gray-400">暂无数据</span>
              </div>
            </ElDescriptionsItem>
          </ElDescriptions>
          <div style="margin-top: 16px">
            <ElButton type="success" @click="handleConnect(server)">
              <icon-mdi-console-line style="margin-right: 4px" />
              连接此主机
            </ElButton>
          </div>
        </ElTabPane>

        <!-- 会话记录 Tab -->
        <ElTabPane label="会话记录" name="sessions">
          <ElTable v-loading="drawerLoading" :data="drawerSessions" size="small">
            <ElTableColumn prop="username" label="用户" width="100" />
            <ElTableColumn prop="loginAccount" label="登录账号" width="100" />
            <ElTableColumn prop="protocol" label="协议" width="70">
              <template #default="{ row }">
                <ElTag size="small">{{ row.protocol?.toUpperCase() }}</ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="startedAt" label="开始时间" min-width="150" />
            <ElTableColumn prop="status" label="状态" width="80">
              <template #default="{ row }">
                <ElTag :type="row.status === 'active' ? 'success' : 'info'" size="small">{{ row.status }}</ElTag>
              </template>
            </ElTableColumn>
          </ElTable>
        </ElTabPane>

        <!-- 监控信息 Tab -->
        <ElTabPane label="监控信息" name="monitor">
          <div class="monitor-section">
            <div class="section-title">资源使用率</div>
            <div class="usage-cards">
              <div class="usage-card">
                <div class="card-label">CPU</div>
                <div class="card-value" :style="{ color: getUsageColor(server.cpuUsage || 0) }">
                  {{ Math.round(server.cpuUsage || 0) }}%
                </div>
                <ElProgress
                  :percentage="Math.round(server.cpuUsage || 0)"
                  :color="getUsageColor(server.cpuUsage || 0)"
                  :show-text="false"
                />
                <div class="card-detail">{{ server.cpu || '-' }} 核</div>
              </div>
              <div class="usage-card">
                <div class="card-label">内存</div>
                <div class="card-value" :style="{ color: getUsageColor(server.memoryUsage || 0) }">
                  {{ Math.round(server.memoryUsage || 0) }}%
                </div>
                <ElProgress
                  :percentage="Math.round(server.memoryUsage || 0)"
                  :color="getUsageColor(server.memoryUsage || 0)"
                  :show-text="false"
                />
                <div class="card-detail">
                  {{
                    server.memory
                      ? ((server.memory * (server.memoryUsage || 0)) / 100).toFixed(1) + ' / ' + server.memory + ' GB'
                      : '-'
                  }}
                </div>
              </div>
              <div class="usage-card">
                <div class="card-label">磁盘</div>
                <div class="card-value" :style="{ color: getUsageColor(server.diskUsage || 0) }">
                  {{ Math.round(server.diskUsage || 0) }}%
                </div>
                <ElProgress
                  :percentage="Math.round(server.diskUsage || 0)"
                  :color="getUsageColor(server.diskUsage || 0)"
                  :show-text="false"
                />
                <div class="card-detail">{{ server.disk || '-' }} GB</div>
              </div>
            </div>
          </div>
          <div class="monitor-section">
            <div class="section-title">磁盘分区详情</div>
            <ElTable :data="formatDiskPartitionsForDrawer(server)" size="small" border>
              <ElTableColumn prop="mount" label="挂载点" width="120" />
              <ElTableColumn label="使用率" width="150">
                <template #default="{ row }">
                  <ElProgress :percentage="Math.round(row.usage)" :color="getUsageColor(row.usage)" />
                </template>
              </ElTableColumn>
              <ElTableColumn prop="usage" label="使用率" width="80" align="center">
                <template #default="{ row }">
                  <span :style="{ color: getUsageColor(row.usage) }">{{ Math.round(row.usage) }}%</span>
                </template>
              </ElTableColumn>
            </ElTable>
            <div
              v-if="!server.diskPartitions || server.diskPartitions.length === 0"
              class="py-12px text-center text-gray-400"
            >
              {{ server.agentStatus === 'running' ? '正在采集...' : '暂无数据，请先部署 Agent' }}
            </div>
          </div>
          <div class="monitor-section">
            <div class="section-title">Agent 状态</div>
            <ElDescriptions :column="2" border size="small">
              <ElDescriptionsItem label="状态">
                <div v-if="server.agentStatus === 'running'" class="flex items-center gap-2">
                  <ElTag type="success" size="small">运行中</ElTag>
                  <span v-if="server.agentVersion" class="text-12px text-gray-500">v{{ server.agentVersion }}</span>
                </div>
                <ElTag v-else-if="server.agentStatus === 'offline'" type="warning" size="small">离线</ElTag>
                <ElTag v-else type="info" size="small">未安装</ElTag>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="监听端口">{{ server.agentPort || 9100 }}</ElDescriptionsItem>
              <ElDescriptionsItem label="最后心跳">
                <span v-if="server.lastHeartbeatAt">{{ formatTime(server.lastHeartbeatAt) }}</span>
                <span v-else class="text-gray-400">-</span>
              </ElDescriptionsItem>
              <ElDescriptionsItem label="指标更新">
                <span v-if="server.metricsUpdatedAt">{{ formatTime(server.metricsUpdatedAt) }}</span>
                <span v-else class="text-gray-400">-</span>
              </ElDescriptionsItem>
            </ElDescriptions>
          </div>
        </ElTabPane>
      </ElTabs>
    </div>
  </ElDrawer>
</template>

<style scoped>
  .overview-section {
    margin-bottom: 24px;
  }
  .overview-section .section-title {
    font-size: 14px;
    font-weight: 500;
    color: #303133;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #e4e7ed;
  }
  .monitor-section {
    margin-bottom: 24px;
  }
  .monitor-section .section-title {
    font-size: 14px;
    font-weight: 500;
    color: #303133;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #e4e7ed;
  }
  .usage-cards {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
  }
  .usage-card {
    padding: 16px;
    background-color: #fafafa;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
    text-align: center;
  }
  .usage-card .card-label {
    font-size: 12px;
    color: #909399;
    margin-bottom: 8px;
  }
  .usage-card .card-value {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 12px;
  }
  .usage-card .card-detail {
    font-size: 12px;
    color: #606266;
    margin-top: 8px;
  }
</style>
