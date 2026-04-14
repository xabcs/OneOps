<template>
    <div class="monitoring-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">监控中心</h2>
                <p class="page-subtitle">实时监控系统核心指标，快速响应异常告警。</p>
            </div>
            <div class="header-actions">
                <el-button type="primary" :icon="Refresh" @click="refreshData">刷新状态</el-button>
                <el-button :icon="Setting">告警配置</el-button>
            </div>
        </header>

        <!-- Real-time Metrics -->
        <el-row :gutter="20" class="metrics-row">
            <el-col :span="6" v-for="metric in metrics" :key="metric.label">
                <el-card shadow="never" class="metric-card">
                    <div class="metric-header">
                        <span class="metric-label">{{ metric.label }}</span>
                        <el-icon :class="metric.status"><component :is="metric.icon" /></el-icon>
                    </div>
                    <div class="metric-body">
                        <div class="metric-value-group">
                            <span class="metric-value">{{ metric.value }}</span>
                            <span class="metric-unit">{{ metric.unit }}</span>
                        </div>
                        <el-progress 
                            :percentage="metric.percentage" 
                            :status="metric.percentage > 85 ? 'exception' : (metric.percentage > 70 ? 'warning' : '')"
                            :show-text="false"
                            :stroke-width="6"
                        />
                    </div>
                    <div class="metric-footer">
                        <span class="metric-trend" :class="metric.trendType">
                            {{ metric.trend }}
                        </span>
                        <span class="metric-period">较 5 分钟前</span>
                    </div>
                </el-card>
            </el-col>
        </el-row>

        <!-- Alerts Management -->
        <el-card shadow="never" class="alerts-card">
            <template #header>
                <div class="card-header">
                    <div class="card-title-group">
                        <span class="card-title">活动告警列表</span>
                        <span class="card-desc">当前系统中所有未关闭的异常通知</span>
                    </div>
                    <div class="card-actions">
                        <el-radio-group v-model="alertFilter" size="small">
                            <el-radio-button label="all">全部</el-radio-button>
                            <el-radio-button label="critical">紧急</el-radio-button>
                            <el-radio-button label="warning">警告</el-radio-button>
                        </el-radio-group>
                    </div>
                </div>
            </template>

            <el-table :data="filteredAlerts" style="width: 100%" v-loading="loading">
                <el-table-column prop="level" label="级别" width="100">
                    <template #default="{ row }">
                        <el-tag :type="getLevelType(row.level)" effect="dark" size="small">
                            {{ formatLevel(row.level) }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="time" label="发生时间" width="180">
                    <template #default="{ row }">
                        <span class="data-value">{{ row.time }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="source" label="告警源" width="180">
                    <template #default="{ row }">
                        <span class="data-value text-primary">{{ row.source }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="message" label="告警详情" min-width="300" />
                <el-table-column prop="status" label="状态" width="120">
                    <template #default="{ row }">
                        <span class="status-indicator" :class="'status-' + row.status">
                            {{ row.status === 'unhandled' ? '待处理' : '处理中' }}
                        </span>
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="160" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleHandle(row)">处理</el-button>
                        <el-button link type="primary" @click="handleIgnore(row)">忽略</el-button>
                        <el-button link type="primary">详情</el-button>
                    </template>
                </el-table-column>
            </el-table>

            <div class="pagination-container">
                <el-pagination layout="total, prev, pager, next" :total="alerts.length" :page-size="10" />
            </div>
        </el-card>
    </div>
</template>

<script setup>
    import { ref, computed, onMounted } from 'vue'
    import { 
        Cpu, Connection, PieChart, Bell, Refresh, Setting,
        Warning, CircleClose, InfoFilled
    } from '@element-plus/icons-vue'
    import { monitoringApi } from '../api/index.js'

    const loading = ref(false)
    const alertFilter = ref('all')
    const metrics = ref([
        { label: 'CPU 平均负载', value: '45.2', unit: '%', percentage: 45.2, icon: Cpu, status: 'primary', trend: '+2.4%', trendType: 'up' },
        { label: '内存使用率', value: '62.8', unit: '%', percentage: 62.8, icon: PieChart, status: 'warning', trend: '-1.2%', trendType: 'down' },
        { label: '网络吞吐量', value: '128', unit: 'MB/s', percentage: 35, icon: Connection, status: 'success', trend: '+12%', trendType: 'up' },
        { label: '活动告警数', value: '14', unit: '个', percentage: 14, icon: Bell, status: 'danger', trend: '0', trendType: 'neutral' }
    ])

    const alerts = ref([])
    const filteredAlerts = computed(() => {
        if (alertFilter.value === 'all') return alerts.value
        return alerts.value.filter(a => a.level === alertFilter.value)
    })

    const formatLevel = (level) => {
        const map = { critical: '紧急', warning: '警告', info: '信息' }
        return map[level] || level
    }

    const getLevelType = (level) => {
        const map = { critical: 'danger', warning: 'warning', info: 'info' }
        return map[level] || ''
    }

    const fetchMonitoringData = async () => {
        loading.value = true
        try {
            const res = await monitoringApi.getMonitoring()
            if (res.code === 200) {
                const data = res.data
                // Update metrics
                metrics.value[0].value = data.cpu
                metrics.value[0].percentage = data.cpu
                metrics.value[1].value = data.memory
                metrics.value[1].percentage = data.memory
                metrics.value[2].value = data.network
                metrics.value[2].percentage = data.network
                
                alerts.value = data.alerts.map(a => ({
                    ...a,
                    status: 'unhandled'
                }))
            }
        } catch (error) {
            console.error('Error fetching monitoring data:', error)
            // Mock data
            alerts.value = [
                { id: 1, time: '2024-04-10 10:30:00', level: 'warning', source: 'Web-Server-01', message: 'CPU 使用率持续超过 80% (当前 85.4%)', status: 'unhandled' },
                { id: 2, time: '2024-04-10 10:15:22', level: 'critical', source: 'DB-Master-01', message: '检测到数据库连接数异常增长，触发限流策略', status: 'unhandled' },
                { id: 3, time: '2024-04-10 09:45:10', level: 'info', source: 'Log-Svc', message: '日志存储空间占用超过 70%，建议清理', status: 'unhandled' },
                { id: 4, time: '2024-04-10 09:12:33', level: 'critical', source: 'Gateway-02', message: '节点心跳丢失，服务已自动切换至备用节点', status: 'unhandled' }
            ]
        } finally {
            loading.value = false
        }
    }

    const refreshData = () => {
        fetchMonitoringData()
    }

    onMounted(fetchMonitoringData)

    const handleHandle = (row) => {
        console.log('Handle alert:', row)
        row.status = 'handled'
    }

    const handleIgnore = (row) => {
        console.log('Ignore alert:', row)
        alerts.value = alerts.value.filter(a => a.id !== row.id)
    }
</script>

<style scoped>
    .monitoring-container {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .page-header {
        display: flex;
        align-items: flex-start;
        gap: 40px;
        margin-bottom: 8px;
    }

    .header-actions {
        display: flex;
        gap: 12px;
    }

    .metrics-row {
        margin-bottom: 4px;
    }

    .metric-card {
        border: none;
    }

    .metric-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
    }

    .metric-label {
        font-size: 0.875rem;
        font-weight: 600;
        color: var(--text-tertiary);
    }

    .metric-header .el-icon {
        font-size: 20px;
    }

    .metric-header .primary { color: var(--primary); }
    .metric-header .warning { color: var(--warning); }
    .metric-header .success { color: var(--success); }
    .metric-header .danger { color: var(--error); }

    .metric-body {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .metric-value-group {
        display: flex;
        align-items: baseline;
        gap: 4px;
    }

    .metric-value {
        font-size: 1.5rem;
        font-weight: 800;
        color: var(--text-primary);
        font-family: var(--mono);
    }

    .metric-unit {
        font-size: 0.875rem;
        color: var(--text-tertiary);
        font-weight: 600;
    }

    .metric-footer {
        margin-top: 16px;
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 0.75rem;
    }

    .metric-trend {
        font-weight: 700;
    }

    .metric-trend.up { color: var(--error); }
    .metric-trend.down { color: var(--success); }
    .metric-trend.neutral { color: var(--text-tertiary); }

    .metric-period {
        color: var(--text-quaternary);
    }

    .alerts-card {
        border: none;
    }

    .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .card-title-group {
        display: flex;
        flex-direction: column;
    }

    .card-title {
        font-size: 1rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    .card-desc {
        font-size: 0.75rem;
        color: var(--text-tertiary);
        margin-top: 2px;
    }

    .pagination-container {
        margin-top: 24px;
        display: flex;
        justify-content: flex-end;
    }

    :deep(.el-table) {
        --el-table-header-bg-color: var(--bg-secondary);
    }
</style>