<script setup lang="ts">
    import { computed, onMounted, onUnmounted, ref } from 'vue';
    import { useRoute, useRouter } from 'vue-router';
    import { ArrowLeft, Document, Refresh } from '@element-plus/icons-vue';
    import { fetchServerExtendedMetrics } from '@/service/api/monitoring';
    import { useWebSocket } from '@/service/websocket';

    defineOptions({
      name: 'MonitoringServersDetail'
    });

    const route = useRoute();
    const router = useRouter();
    const loading = ref(false);
    const empty = ref(false);

    const serverId = computed(() => Number(route.query.id));
    const activeTab = ref('overview');
    const metrics = ref<Monitoring.ExtendedMetrics | null>(null);
    const refreshTimer = ref<ReturnType<typeof setInterval> | null>(null);

    // WebSocket 客户端
    const { subscribe, unsubscribe } = useWebSocket();

    // 指标更新处理器
    function handleMetricsUpdate(data: { server_id: number }) {
      // 只处理当前主机的数据
      if (data.server_id == serverId.value) {
        // WebSocket 只推送简化数据作为通知，详情页面需要完整的指标数据
        // 收到通知后通过 HTTP API 重新获取完整数据
        getExtendedMetrics();
      } else {
      }
    }

    async function getExtendedMetrics() {
      if (!serverId.value) {
        empty.value = true;
        return;
      }
      loading.value = true;
      try {
        const { data } = await fetchServerExtendedMetrics(serverId.value);
        if (data) {
          metrics.value = data as Monitoring.ExtendedMetrics;
          empty.value = false;
        } else {
          empty.value = true;
        }
      } catch (error) {
        console.error('获取监控指标失败:', error);
        empty.value = true;
      }
      loading.value = false;
    }

    function handleRefresh() {
      getExtendedMetrics();
    }

    function handleBack() {
      router.back();
    }

    function formatBytes(bytes: number): string {
      if (!bytes || bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return `${Number.parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`;
    }

    function formatUptime(seconds: number): string {
      const days = Math.floor(seconds / 86400);
      const hours = Math.floor((seconds % 86400) / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      if (days > 0) return `${days}天 ${hours}小时`;
      if (hours > 0) return `${hours}小时 ${minutes}分钟`;
      return `${minutes}分钟`;
    }

    function getProgressColor(value: number): string {
      if (value > 80) return '#ff4d4f';
      if (value > 60) return '#faad14';
      return '#52c41a';
    }

    function getServiceTagType(status: string): 'success' | 'danger' | 'warning' | 'info' {
      const map: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
        active: 'success',
        inactive: 'info',
        failed: 'danger',
        unknown: 'warning'
      };
      return map[status] ?? 'info';
    }

    function getProcessTagType(status: string): 'success' | 'danger' | 'warning' | 'info' {
      const map: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
        running: 'success',
        sleeping: 'info',
        stopped: 'info',
        zombie: 'danger',
        dead: 'danger',
        unknown: 'warning'
      };
      return map[status] ?? 'info';
    }

    onMounted(() => {
      // 初始加载
      getExtendedMetrics();

      // 订阅 WebSocket 指标更新
      subscribe('metrics_update', handleMetricsUpdate);
    });

    onUnmounted(() => {
      // 取消订阅
      unsubscribe('metrics_update', handleMetricsUpdate);
    });
</script>

<template>
    <div class="space-y-4">
        <ElCard shadow="never">
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                    <ElButton circle @click="handleBack">
                        <ElIcon :size="18">
                            <ArrowLeft />
                        </ElIcon>
                    </ElButton>
                    <span class="text-lg font-semibold">主机监控详情</span>
                </div>
                <ElButton type="primary" :loading="loading" @click="handleRefresh">
                    <ElIcon :size="16">
                        <Refresh />
                    </ElIcon>
                    刷新
                </ElButton>
            </div>
        </ElCard>

        <!-- 内容区域 -->
        <!-- 骨架屏 -->
        <ElSkeleton v-if="loading && !metrics" :rows="8" animated />

        <!-- 空状态 -->
        <ElCard v-else-if="empty && !metrics" shadow="never">
            <ElEmpty description="暂无监控数据">
                <template #image>
                    <ElIcon :size="100" color="#dcdfe6">
                        <Document />
                    </ElIcon>
                </template>
                <p class="mt-4 text-gray-500">该主机可能未安装 Agent 或 Agent 未正常运行</p>
                <ElButton type="primary" class="mt-4" @click="handleRefresh">重新加载</ElButton>
            </ElEmpty>
        </ElCard>

        <!-- 主内容 -->
        <template v-else-if="metrics">
            <!-- 主机基本信息 -->
            <ElCard shadow="never" header="主机信息">
                <ElDescriptions :column="4" border>
                    <ElDescriptionsItem label="主机名">
                        {{ metrics.systemInfo?.hostname || '-' }}
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="操作系统">
                        {{ metrics.systemInfo?.os?.platform || '-' }} {{ metrics.systemInfo?.os?.platformVersion || '' }}
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="内核版本">
                        {{ metrics.systemInfo?.os?.kernelVersion || '-' }}
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="系统架构">
                        {{ metrics.systemInfo?.os?.kernelArch || '-' }}
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="运行时间" :span="4">
                        {{ metrics.systemInfo?.uptime > 0 ? formatUptime(metrics.systemInfo?.uptime) : '-' }}
                    </ElDescriptionsItem>
                </ElDescriptions>
            </ElCard>

            <!-- Tab 内容 -->
            <ElCard shadow="never">
                <ElTabs v-model="activeTab">
                    <!-- 监控概览 -->
                    <ElTabPane label="监控概览" name="overview">
                        <ElRow :gutter="16" class="mb-4">
                            <!-- CPU 指标 -->
                            <ElCol :xs="24" :sm="12" :md="6">
                                <ElCard shadow="hover" class="metric-card">
                                    <div class="metric-header">CPU</div>
                                    <div class="metric-value">{{ metrics.performance?.cpu?.usagePercent || (0).toFixed(1) }}%</div>
                                    <ElProgress :percentage="Math.round(metrics.performance?.cpu?.usagePercent || 0)" :color="getProgressColor(metrics.performance?.cpu?.usagePercent || 0)" :show-text="false" :stroke-width="6" class="mb-2" />
                                    <div class="metric-details">
                                        <div>用户: {{ metrics.performance?.cpu?.user || (0).toFixed(1) }}%</div>
                                        <div>系统: {{ metrics.performance?.cpu?.system || (0).toFixed(1) }}%</div>
                                        <div>空闲: {{ metrics.performance?.cpu?.idle || (0).toFixed(1) }}%</div>
                                        <div>核心数: {{ metrics.performance?.cpu?.cores || 0 }}</div>
                                    </div>
                                </ElCard>
                            </ElCol>

                            <!-- 内存指标 -->
                            <ElCol :xs="24" :sm="12" :md="6">
                                <ElCard shadow="hover" class="metric-card">
                                    <div class="metric-header">内存</div>
                                    <div class="metric-value">{{ metrics.performance?.memory?.usedPercent || (0).toFixed(1) }}%</div>
                                    <ElProgress :percentage="Math.round(metrics.performance?.memory?.usedPercent || 0)" :color="getProgressColor(metrics.performance?.memory?.usedPercent || 0)" :show-text="false" :stroke-width="6" class="mb-2" />
                                    <div class="metric-details">
                                        <div>已用: {{ formatBytes(metrics.performance?.memory?.used || 0) }}</div>
                                        <div>可用: {{ formatBytes(metrics.performance?.memory?.available || 0) }}</div>
                                        <div>总计: {{ formatBytes(metrics.performance?.memory?.total || 0) }}</div>
                                    </div>
                                </ElCard>
                            </ElCol>

                            <!-- 磁盘指标 -->
                            <ElCol :xs="24" :sm="12" :md="6">
                                <ElCard shadow="hover" class="metric-card">
                                    <div class="metric-header">磁盘</div>
                                    <div class="metric-value">{{ metrics.performance?.disk?.usedPercent || (0).toFixed(1) }}%</div>
                                    <ElProgress :percentage="Math.round(metrics.performance?.disk?.usedPercent || 0)" :color="getProgressColor(metrics.performance?.disk?.usedPercent || 0)" :show-text="false" :stroke-width="6" class="mb-2" />
                                    <div class="metric-details">
                                        <div>已用: {{ formatBytes(metrics.performance?.disk?.used || 0) }}</div>
                                        <div>可用: {{ formatBytes(metrics.performance?.disk?.free || 0) }}</div>
                                        <div>总计: {{ formatBytes(metrics.performance?.disk?.total || 0) }}</div>
                                    </div>
                                </ElCard>
                            </ElCol>

                            <!-- 系统负载 -->
                            <ElCol :xs="24" :sm="12" :md="6">
                                <ElCard shadow="hover" class="metric-card">
                                    <div class="metric-header">系统负载</div>
                                    <div class="metric-value">{{ metrics.performance?.load?.load5 || (0).toFixed(2) }}</div>
                                    <div class="metric-details mt-4">
                                        <div>1分钟: {{ metrics.performance?.load?.load1 || (0).toFixed(2) }}</div>
                                        <div>5分钟: {{ metrics.performance?.load?.load5 || (0).toFixed(2) }}</div>
                                        <div>15分钟: {{ metrics.performance?.load?.load15 || (0).toFixed(2) }}</div>
                                    </div>
                                </ElCard>
                            </ElCol>
                        </ElRow>

                        <!-- 磁盘分区详情 -->
                        <div class="mb-4">
                            <div class="section-title">磁盘分区</div>
                            <ElTable :data="metrics.performance?.disk?.partitions || []" border size="small" row-key="mountpoint">
                                <ElTableColumn label="设备" prop="device" width="150" />
                                <ElTableColumn label="挂载点" prop="mountpoint" width="100" />
                                <ElTableColumn label="类型" prop="fstype" width="80" />
                                <ElTableColumn label="总容量" width="120">
                                    <template #default="{ row }">{{ formatBytes(row.total) }}</template>
                                </ElTableColumn>
                                <ElTableColumn label="已用" width="120">
                                    <template #default="{ row }">{{ formatBytes(row.used) }}</template>
                                </ElTableColumn>
                                <ElTableColumn label="可用" width="120">
                                    <template #default="{ row }">{{ formatBytes(row.free) }}</template>
                                </ElTableColumn>
                                <ElTableColumn label="使用率" min-width="150" align="center">
                                    <template #default="{ row }">
                                        <ElProgress :percentage="Math.round(row.usedPercent || 0)" :color="getProgressColor(row.usedPercent || 0)" :stroke-width="8" :format="() => (row.usedPercent || 0).toFixed(1) + '%'" />
                                    </template>
                                </ElTableColumn>
                            </ElTable>
                        </div>

                        <!-- 网络接口 -->
                        <div class="mb-4">
                            <div class="section-title">网络接口</div>
                            <ElTable :data="metrics.performance?.network?.interfaces || []" border size="small" row-key="name">
                                <ElTableColumn label="接口" prop="name" width="150" />
                                <ElTableColumn label="发送字节" width="150">
                                    <template #default="{ row }">{{ formatBytes(row.bytesSent) }}</template>
                                </ElTableColumn>
                                <ElTableColumn label="接收字节" width="150">
                                    <template #default="{ row }">{{ formatBytes(row.bytesRecv) }}</template>
                                </ElTableColumn>
                            </ElTable>
                        </div>

                        <!-- 网络连接 -->
                        <div class="section-title">网络连接状态</div>
                        <ElRow :gutter="16">
                            <ElCol :span="6">
                                <ElCard shadow="never" class="conn-stat">
                                    <div class="conn-value">{{ metrics.performance?.network?.connections?.established || 0 }}</div>
                                    <div class="conn-label">ESTABLISHED</div>
                                </ElCard>
                            </ElCol>
                            <ElCol :span="6">
                                <ElCard shadow="never" class="conn-stat">
                                    <div class="conn-value">{{ metrics.performance?.network?.connections?.timeWait || 0 }}</div>
                                    <div class="conn-label">TIME_WAIT</div>
                                </ElCard>
                            </ElCol>
                            <ElCol :span="6">
                                <ElCard shadow="never" class="conn-stat">
                                    <div class="conn-value">{{ metrics.performance?.network?.connections?.listen || 0 }}</div>
                                    <div class="conn-label">LISTEN</div>
                                </ElCard>
                            </ElCol>
                        </ElRow>
                    </ElTabPane>

                    <!-- 服务状态 -->
                    <ElTabPane label="服务状态" name="services">
                        <div class="mb-4">
                            <div class="section-title">系统服务</div>
                            <ElTable :data="metrics.serviceStatus?.systemdServices || []" border size="small" row-key="name">
                                <ElTableColumn label="服务名" prop="name" width="200" />
                                <ElTableColumn label="状态" width="100" align="center">
                                    <template #default="{ row }">
                                        <ElTag :type="getServiceTagType(row.status)" size="small">{{ row.status }}</ElTag>
                                    </template>
                                </ElTableColumn>
                                <ElTableColumn label="子状态" prop="subStatus" width="100" align="center" />
                                <ElTableColumn label="描述" prop="description" show-overflow-tooltip />
                            </ElTable>
                        </div>

                        <div class="section-title">监听端口</div>
                        <ElTable :data="metrics.serviceStatus?.listenPorts || []" border size="small" row-key="port">
                            <ElTableColumn label="端口" prop="port" width="80" align="center" />
                            <ElTableColumn label="协议" prop="protocol" width="80" align="center" />
                            <ElTableColumn label="地址" prop="address" width="150" />
                            <ElTableColumn label="进程" prop="process" width="150" />
                            <ElTableColumn label="PID" prop="pid" width="100" align="center" />
                        </ElTable>
                    </ElTabPane>

                    <!-- 进程监控 -->
                    <ElTabPane label="进程监控" name="processes">
                        <ElAlert title="提示" description="以下展示 CPU 使用率最高的 Top 20 进程" type="info" :closable="false" class="mb-4" show-icon />
                        <ElTable :data="metrics.processInfo?.top || []" border size="small" row-key="pid" max-height="600">
                            <ElTableColumn label="PID" prop="pid" width="80" align="center" />
                            <ElTableColumn label="进程名" prop="name" width="150" show-overflow-tooltip />
                            <ElTableColumn label="CPU%" prop="cpuPercent" width="90" align="center">
                                <template #default="{ row }">{{ row.cpuPercent?.toFixed(1) || '0.0' }}</template>
                            </ElTableColumn>
                            <ElTableColumn label="内存%" prop="memoryPercent" width="90" align="center">
                                <template #default="{ row }">{{ row.memoryPercent?.toFixed(1) || '0.0' }}</template>
                            </ElTableColumn>
                            <ElTableColumn label="内存" width="120" align="center">
                                <template #default="{ row }">{{ formatBytes(row.memoryBytes || 0) }}</template>
                            </ElTableColumn>
                            <ElTableColumn label="状态" width="90" align="center">
                                <template #default="{ row }">
                                    <ElTag :type="getProcessTagType(row.status)" size="small">{{ row.status }}</ElTag>
                                </template>
                            </ElTableColumn>
                            <ElTableColumn label="用户" prop="username" width="100" />
                            <ElTableColumn label="线程数" prop="numThreads" width="80" align="center" />
                            <ElTableColumn label="命令" prop="cmdline" show-overflow-tooltip />
                        </ElTable>
                    </ElTabPane>

                    <!-- 硬件信息 -->
                    <ElTabPane label="硬件信息" name="hardware">
                        <!-- CPU 详细信息 -->
                        <div class="mb-4">
                            <div class="section-title">CPU 详细信息</div>
                            <ElDescriptions :column="2" border>
                                <ElDescriptionsItem label="厂商">{{ metrics.hardwareInfo?.cpu?.vendor || '-' }}</ElDescriptionsItem>
                                <ElDescriptionsItem label="型号">{{ metrics.hardwareInfo?.cpu?.model || '-' }}</ElDescriptionsItem>
                                <ElDescriptionsItem label="物理核心数">
                                    {{ metrics.hardwareInfo?.cpu?.cores || '-' }}
                                </ElDescriptionsItem>
                                <ElDescriptionsItem label="逻辑线程数">
                                    {{ metrics.hardwareInfo?.cpu?.threads || '-' }}
                                </ElDescriptionsItem>
                                <ElDescriptionsItem label="主频">{{ metrics.hardwareInfo?.cpu?.mhz || '-' }} MHz</ElDescriptionsItem>
                                <ElDescriptionsItem label="架构">{{ metrics.systemInfo?.os?.kernelArch || '-' }}</ElDescriptionsItem>
                            </ElDescriptions>
                        </div>

                        <!-- 内存详细信息 -->
                        <div class="mb-4">
                            <div class="section-title">内存详细信息</div>
                            <ElDescriptions :column="1" border class="mb-4">
                                <ElDescriptionsItem label="总容量">
                                    {{ formatBytes(metrics.hardwareInfo?.memory?.total || 0) }}
                                </ElDescriptionsItem>
                            </ElDescriptions>

                            <!-- 内存插槽详情 -->
                            <ElTable v-if="metrics.hardwareInfo?.memory?.slots && metrics.hardwareInfo.memory.slots.length > 0" :data="metrics.hardwareInfo.memory.slots" border size="small" row-key="slotNumber">
                                <ElTableColumn label="插槽编号" prop="slotNumber" width="100" align="center" />
                                <ElTableColumn label="容量" width="120" align="center">
                                    <template #default="{ row }">{{ formatBytes(row.capacity || 0) }}</template>
                                </ElTableColumn>
                                <ElTableColumn label="类型" prop="type" width="120" />
                                <ElTableColumn label="厂商" prop="vendor" width="150" />
                                <ElTableColumn label="速度" prop="speed" width="120" />
                                <ElTableColumn label="ECC" width="80" align="center">
                                    <template #default="{ row }">
                                        <ElTag :type="row.hasECC ? 'success' : 'info'" size="small">
                                            {{ row.hasECC ? '是' : '否' }}
                                        </ElTag>
                                    </template>
                                </ElTableColumn>
                            </ElTable>
                            <ElAlert v-else title="内存插槽信息" type="info" :closable="false" class="mt-2">暂无详细插槽信息</ElAlert>
                        </div>

                        <!-- 磁盘设备详细信息 -->
                        <div class="section-title">磁盘设备详细信息</div>
                        <ElTable :data="metrics.hardwareInfo?.disk || []" border size="small" row-key="name">
                            <ElTableColumn label="设备" prop="name" width="120" />
                            <ElTableColumn label="型号" prop="model" min-width="200" show-overflow-tooltip />
                            <ElTableColumn label="序列号" prop="serial" width="150" />
                            <ElTableColumn label="固件版本" prop="firmware" width="120" />
                            <ElTableColumn label="类型" prop="type" width="100" />
                            <ElTableColumn label="容量" width="120" align="center">
                                <template #default="{ row }">{{ formatBytes(row.size || 0) }}</template>
                            </ElTableColumn>
                            <ElTableColumn label="S.M.A.R.T. 状态" width="120" align="center">
                                <template #default="{ row }">
                                    <ElTag v-if="row.smartStatus" :type="row.smartStatus === 'PASSED' || row.smartStatus === 'OK' ? 'success' : 'warning'" size="small">
                                        {{ row.smartStatus }}
                                    </ElTag>
                                    <span v-else class="text-gray-400">-</span>
                                </template>
                            </ElTableColumn>
                        </ElTable>
                    </ElTabPane>

                    <!-- 网络配置 -->
                    <ElTabPane label="网络配置" name="network">
                        <!-- 网络接口详情 -->
                        <div class="mb-4">
                            <div class="section-title">网络接口详情</div>
                            <ElTable :data="metrics.networkConfig?.interfaces || []" border size="small" row-key="name">
                                <ElTableColumn label="接口" prop="name" width="120" />
                                <ElTableColumn label="MAC地址" prop="hardwareAddr" width="170" />
                                <ElTableColumn label="MTU" prop="mtu" width="80" align="center" />
                                <ElTableColumn label="状态" width="80" align="center">
                                    <template #default="{ row }">
                                        <ElTag :type="row.isUp ? 'success' : 'info'" size="small">
                                            {{ row.isUp ? 'UP' : 'DOWN' }}
                                        </ElTag>
                                    </template>
                                </ElTableColumn>
                                <ElTableColumn label="IP地址" min-width="200">
                                    <template #default="{ row }">
                                        <div v-for="addr in row.addrs || []" :key="addr.ip" class="mb-1">
                                            <ElTag type="primary" size="small">
                                                {{ addr.ip }}
                                                <span v-if="addr.mask" class="ml-1 text-gray-400">/ {{ addr.mask }}</span>
                                            </ElTag>
                                            <ElTag v-if="addr.family" type="info" size="small" class="ml-1">{{ addr.family }}</ElTag>
                                        </div>
                                    </template>
                                </ElTableColumn>
                                <ElTableColumn label="标志" min-width="150">
                                    <template #default="{ row }">
                                        <template v-if="row.flags && row.flags.length > 0">
                                            <ElTag v-for="flag in row.flags" :key="flag" size="small" class="mr-1">
                                                {{ flag }}
                                            </ElTag>
                                        </template>
                                        <span v-else class="text-gray-400">-</span>
                                    </template>
                                </ElTableColumn>
                            </ElTable>
                        </div>

                        <!-- 路由表 -->
                        <div class="mb-4">
                            <div class="section-title">路由表</div>
                            <ElTable v-if="metrics.networkConfig?.routes && metrics.networkConfig.routes.length > 0" :data="metrics.networkConfig.routes" border size="small" row-key="destination">
                                <ElTableColumn label="目标网络" prop="destination" width="180" />
                                <ElTableColumn label="网关" prop="gateway" width="150" />
                                <ElTableColumn label="接口" prop="interface" width="120" />
                                <ElTableColumn label="标志" prop="flags" width="150" />
                                <ElTableColumn label="度量值" prop="metric" width="80" align="center" />
                            </ElTable>
                            <ElAlert v-else title="路由表" type="info" :closable="false">暂无路由表信息</ElAlert>
                        </div>

                        <!-- DNS 配置 -->
                        <div class="section-title">DNS 配置</div>
                        <ElDescriptions v-if="metrics.networkConfig?.dns" :column="1" border>
                            <ElDescriptionsItem label="DNS 服务器">
                                <div v-if="metrics.networkConfig.dns.servers && metrics.networkConfig.dns.servers.length > 0">
                                    <ElTag v-for="server in metrics.networkConfig.dns.servers" :key="server" type="primary" size="small" class="mr-1">
                                        {{ server }}
                                    </ElTag>
                                </div>
                                <span v-else class="text-gray-400">-</span>
                            </ElDescriptionsItem>
                            <ElDescriptionsItem label="搜索域">
                                <div v-if="metrics.networkConfig.dns.searchDomains && metrics.networkConfig.dns.searchDomains.length > 0">
                                    <ElTag v-for="domain in metrics.networkConfig.dns.searchDomains" :key="domain" type="info" size="small" class="mr-1">
                                        {{ domain }}
                                    </ElTag>
                                </div>
                                <span v-else class="text-gray-400">-</span>
                            </ElDescriptionsItem>
                        </ElDescriptions>
                        <ElAlert v-else title="DNS 配置" type="info" :closable="false">暂无 DNS 配置信息</ElAlert>
                    </ElTabPane>

                    <!-- 安全信息 -->
                    <ElTabPane label="安全信息" name="security">
                        <div class="mb-4">
                            <div class="section-title">SSH 配置</div>
                            <ElDescriptions :column="2" border>
                                <ElDescriptionsItem label="端口">{{ metrics.securityInfo?.ssh?.port || '-' }}</ElDescriptionsItem>
                                <ElDescriptionsItem label="Root登录">
                                    {{ metrics.securityInfo?.ssh?.permitRootLogin || '-' }}
                                </ElDescriptionsItem>
                                <ElDescriptionsItem label="密码认证">
                                    {{ metrics.securityInfo?.ssh?.passwordAuthentication || '-' }}
                                </ElDescriptionsItem>
                            </ElDescriptions>
                        </div>

                        <div class="mb-4">
                            <div class="section-title">防火墙</div>
                            <ElDescriptions :column="2" border>
                                <ElDescriptionsItem label="后端">
                                    {{ metrics.securityInfo?.firewall?.backend || '-' }}
                                </ElDescriptionsItem>
                                <ElDescriptionsItem label="状态">
                                    <ElTag :type="metrics.securityInfo?.firewall?.status === 'active' ? 'success' : 'info'" size="small">
                                        {{ metrics.securityInfo?.firewall?.status || '-' }}
                                    </ElTag>
                                </ElDescriptionsItem>
                            </ElDescriptions>
                        </div>

                        <div class="section-title">SELinux</div>
                        <ElDescriptions :column="2" border>
                            <ElDescriptionsItem label="启用">
                                <ElTag :type="metrics.securityInfo?.selinux?.enabled ? 'success' : 'info'" size="small">
                                    {{ metrics.securityInfo?.selinux?.enabled ? '是' : '否' }}
                                </ElTag>
                            </ElDescriptionsItem>
                            <ElDescriptionsItem label="模式">{{ metrics.securityInfo?.selinux?.mode || '-' }}</ElDescriptionsItem>
                        </ElDescriptions>
                    </ElTabPane>
                </ElTabs>

                <!-- 采集时间 -->
                <div class="px-4 pb-4 text-right">
                    <span class="text-sm text-gray-400">
                        最后更新: {{ metrics.collectedAt ? new Date(metrics.collectedAt).toLocaleString('zh-CN') : '-' }}
                    </span>
                </div>
            </ElCard>
        </template>
    </div>
</template>

<style scoped>
    .metric-card {
      margin-bottom: 0;
      transition: all 0.3s;
    }

    .metric-card:hover {
      transform: translateY(-2px);
    }

    .metric-header {
      font-size: 14px;
      color: #666;
      margin-bottom: 8px;
    }

    .metric-value {
      font-size: 28px;
      font-weight: bold;
      color: #303133;
      margin-bottom: 8px;
    }

    .metric-details {
      font-size: 12px;
      color: #999;
      line-height: 1.8;
    }

    .section-title {
      font-size: 14px;
      font-weight: 600;
      color: #303133;
      margin-bottom: 12px;
      padding-left: 8px;
      border-left: 3px solid #409eff;
    }

    .conn-stat {
      text-align: center;
      padding: 8px;
    }

    .conn-value {
      font-size: 24px;
      font-weight: bold;
      color: #409eff;
    }

    .conn-label {
      font-size: 12px;
      color: #666;
      margin-top: 4px;
    }
</style>
