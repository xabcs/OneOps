<template>
    <div class="dashboard-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">仪表盘概览</h2>
                <p class="page-subtitle">欢迎回来, {{ user?.username }}。这是您今天的系统运行概况。</p>
            </div>
        </header>

        <!-- KPI Stats -->
        <el-row :gutter="20" class="stats-row">
            <el-col :span="6" v-for="stat in statsConfig" :key="stat.title">
                <el-card shadow="hover" class="stat-card">
                    <div class="stat-content">
                        <div class="stat-main">
                            <span class="stat-label">{{ stat.title }}</span>
                            <div class="stat-value-group">
                                <span class="stat-value">{{ stat.value }}</span>
                                <span class="stat-unit" v-if="stat.unit">{{ stat.unit }}</span>
                            </div>
                        </div>
                        <div class="stat-icon" :class="stat.type">
                            <el-icon><component :is="stat.icon" /></el-icon>
                        </div>
                    </div>
                    <div class="stat-footer">
                        <span class="stat-trend" :class="stat.trendType">
                            <el-icon v-if="stat.trendType === 'up'"><CaretTop /></el-icon>
                            <el-icon v-if="stat.trendType === 'down'"><CaretBottom /></el-icon>
                            {{ stat.trend }}
                        </span>
                        <span class="stat-period">较昨日同期</span>
                    </div>
                </el-card>
            </el-col>
        </el-row>

        <!-- Main Content Area -->
        <el-row :gutter="20" class="content-row">
            <!-- System Load Chart -->
            <el-col :span="16">
                <el-card shadow="never" class="chart-card">
                    <template #header>
                        <div class="card-header">
                            <div class="card-title-group">
                                <span class="card-title">系统负载趋势</span>
                                <span class="card-desc">实时监控全网服务器 CPU 平均负载</span>
                            </div>
                            <el-radio-group v-model="timeRange" size="small">
                                <el-radio-button label="1h">1小时</el-radio-button>
                                <el-radio-button label="6h">6小时</el-radio-button>
                                <el-radio-button label="24h">24小时</el-radio-button>
                            </el-radio-group>
                        </div>
                    </template>
                    <div class="chart-container">
                        <div ref="chartRef" class="d3-chart"></div>
                    </div>
                </el-card>
            </el-col>

            <!-- Recent Activity -->
            <el-col :span="8">
                <el-card shadow="never" class="activity-card">
                    <template #header>
                        <div class="card-header">
                            <span class="card-title">最近活动日志</span>
                            <el-button link type="primary">查看全部</el-button>
                        </div>
                    </template>
                    <el-scrollbar height="340px">
                        <div class="activity-list">
                            <div v-for="item in recentActivities" :key="item.id" class="activity-item">
                                <div class="activity-dot" :class="item.type"></div>
                                <div class="activity-info">
                                    <div class="activity-msg">{{ item.message }}</div>
                                    <div class="activity-time">{{ item.time }}</div>
                                </div>
                            </div>
                        </div>
                    </el-scrollbar>
                </el-card>
            </el-col>
        </el-row>

        <!-- Server Status Table -->
        <el-row class="table-row">
            <el-col :span="24">
                <el-card shadow="never">
                    <template #header>
                        <div class="card-header">
                            <span class="card-title">关键服务器状态</span>
                            <el-button link type="primary">管理资产</el-button>
                        </div>
                    </template>
                    <el-table :data="serverStatus" style="width: 100%">
                        <el-table-column prop="name" label="服务器名称" min-width="150" />
                        <el-table-column prop="ip" label="IP 地址" width="140">
                            <template #default="{ row }">
                                <span class="data-value">{{ row.ip }}</span>
                            </template>
                        </el-table-column>
                        <el-table-column prop="status" label="运行状态" width="120">
                            <template #default="{ row }">
                                <span class="status-indicator" :class="'status-' + row.status">
                                    {{ row.status === 'online' ? '运行中' : '已离线' }}
                                </span>
                            </template>
                        </el-table-column>
                        <el-table-column prop="cpu" label="CPU 使用率" width="150">
                            <template #default="{ row }">
                                <el-progress :percentage="row.cpu" :status="row.cpu > 80 ? 'exception' : ''" />
                            </template>
                        </el-table-column>
                        <el-table-column prop="mem" label="内存使用率" width="150">
                            <template #default="{ row }">
                                <el-progress :percentage="row.mem" :status="row.mem > 80 ? 'warning' : ''" />
                            </template>
                        </el-table-column>
                        <el-table-column label="操作" width="100" fixed="right">
                            <template #default>
                                <el-button link type="primary">详情</el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script setup>
    import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
    import { useStore } from 'vuex'
    import * as d3 from 'd3'
    import { 
        Monitor, Timer, DataLine, Warning, Refresh, Plus,
        CaretTop, CaretBottom
    } from '@element-plus/icons-vue'
    import { taskApi } from '../api/index.js'

    const store = useStore()
    const user = computed(() => store.getters.user)
    const currentTheme = computed(() => store.getters.currentTheme)
    const timeRange = ref('1h')
    const chartRef = ref(null)
    let refreshTimer = null

    const statsConfig = ref([
        { title: '在线节点', value: '128', icon: Monitor, type: 'primary', trend: '12%', trendType: 'up' },
        { title: '待处理任务', value: '42', icon: Timer, type: 'success', trend: '5%', trendType: 'down' },
        { title: '系统平均负载', value: '0.85', icon: DataLine, type: 'warning', trend: '2%', trendType: 'up' },
        { title: '活动告警', value: '14', icon: Warning, type: 'danger', trend: '0%', trendType: 'neutral' }
    ])

    const recentActivities = ref([
        { id: 1, time: '2024-04-10 10:30:05', message: '生产环境 K8s 集群扩容完成', type: 'success' },
        { id: 2, time: '2024-04-10 09:45:12', message: '数据库主从同步延迟告警已恢复', type: 'primary' },
        { id: 3, time: '2024-04-10 09:12:33', message: '检测到异常登录尝试: 192.168.45.12', type: 'danger' },
        { id: 4, time: '2024-04-10 08:30:00', message: '每日全量备份任务启动', type: 'info' },
        { id: 5, time: '2024-04-10 07:20:15', message: '服务器 Web-Node-04 自动重启成功', type: 'warning' }
    ])

    const serverStatus = ref([
        { name: 'Core-API-Gateway', ip: '10.0.1.45', status: 'online', cpu: 45, mem: 62 },
        { name: 'DB-Master-Cluster', ip: '10.0.2.10', status: 'online', cpu: 78, mem: 85 },
        { name: 'Redis-Cache-01', ip: '10.0.2.22', status: 'online', cpu: 12, mem: 45 },
        { name: 'Web-Frontend-01', ip: '10.0.1.102', status: 'offline', cpu: 0, mem: 0 },
        { name: 'Log-Collector-Svc', ip: '10.0.5.8', status: 'online', cpu: 34, mem: 56 }
    ])

    const refreshData = async () => {
        try {
            // Simulate data fluctuation
            statsConfig.value[2].value = (0.7 + Math.random() * 0.3).toFixed(2)
            serverStatus.value.forEach(s => {
                if (s.status === 'online') {
                    s.cpu = Math.min(100, Math.max(0, s.cpu + (Math.random() * 10 - 5)))
                    s.mem = Math.min(100, Math.max(0, s.mem + (Math.random() * 6 - 3)))
                }
            })
            
            await taskApi.getTasks()
            renderChart()
        } catch (error) {
            console.error('Refresh failed:', error)
            renderChart()
        }
    }

    const renderChart = () => {
        if (!chartRef.value) return
        
        // Clear previous chart
        d3.select(chartRef.value).selectAll("*").remove()

        const width = chartRef.value.clientWidth
        const height = 300
        const margin = { top: 20, right: 30, bottom: 30, left: 40 }

        // Get colors from CSS variables
        const style = getComputedStyle(document.documentElement)
        const primaryColor = style.getPropertyValue('--primary').trim() || '#0070cc'
        const textColor = style.getPropertyValue('--text-quaternary').trim() || '#bfbfbf'
        const gridColor = style.getPropertyValue('--border').trim() || '#f0f0f0'

        const svg = d3.select(chartRef.value)
            .append("svg")
            .attr("width", width)
            .attr("height", height)
            .append("g")
            .attr("transform", `translate(${margin.left},${margin.top})`)

        // Generate mock data
        const data = d3.range(24).map(i => ({
            time: i,
            value: 30 + Math.random() * 40
        }))

        const x = d3.scaleLinear()
            .domain([0, 23])
            .range([0, width - margin.left - margin.right])

        const y = d3.scaleLinear()
            .domain([0, 100])
            .range([height - margin.top - margin.bottom, 0])

        // Add X axis
        svg.append("g")
            .attr("transform", `translate(0,${height - margin.top - margin.bottom})`)
            .call(d3.axisBottom(x).ticks(12).tickFormat(d => `${d}:00`))
            .attr("color", textColor)
            .selectAll("text")
            .style("font-size", "11px")

        // Add Y axis
        svg.append("g")
            .call(d3.axisLeft(y).ticks(5))
            .attr("color", textColor)
            .selectAll("text")
            .style("font-size", "11px")

        // Add grid lines
        svg.append("g")
            .attr("class", "grid")
            .attr("opacity", 0.05)
            .call(d3.axisLeft(y).ticks(5).tickSize(-width + margin.left + margin.right).tickFormat(""))

        // Add area
        const area = d3.area()
            .x(d => x(d.time))
            .y0(height - margin.top - margin.bottom)
            .y1(d => y(d.value))
            .curve(d3.curveMonotoneX)

        svg.append("path")
            .datum(data)
            .attr("fill", "url(#gradient)")
            .attr("d", area)

        // Add line
        const line = d3.line()
            .x(d => x(d.time))
            .y(d => y(d.value))
            .curve(d3.curveMonotoneX)

        svg.append("path")
            .datum(data)
            .attr("fill", "none")
            .attr("stroke", primaryColor)
            .attr("stroke-width", 2)
            .attr("d", line)

        // Add gradient
        const defs = svg.append("defs")
        const gradient = defs.append("linearGradient")
            .attr("id", "gradient")
            .attr("x1", "0%").attr("y1", "0%")
            .attr("x2", "0%").attr("y2", "100%")

        gradient.append("stop").attr("offset", "0%").attr("stop-color", primaryColor).attr("stop-opacity", 0.15)
        gradient.append("stop").attr("offset", "100%").attr("stop-color", primaryColor).attr("stop-opacity", 0)
    }

    onMounted(() => {
        renderChart()
        window.addEventListener('resize', renderChart)
        
        // Auto refresh every 5 seconds
        refreshTimer = setInterval(refreshData, 5000)
    })

    onUnmounted(() => {
        window.removeEventListener('resize', renderChart)
        if (refreshTimer) clearInterval(refreshTimer)
    })

    watch([timeRange, currentTheme], () => {
        renderChart()
    })
</script>

<style scoped>
    .dashboard-container {
        display: flex;
        flex-direction: column;
        gap: 32px;
        padding: 32px;
        max-width: 1440px;
        margin: 0 auto;
    }

    .stats-row {
        margin-bottom: 4px;
    }

    .page-header {
        margin-bottom: 8px;
    }

    .page-title {
        font-size: 28px;
        font-weight: 700;
        color: var(--text-primary);
        letter-spacing: -0.02em;
        margin: 0 0 8px 0;
    }

    .page-subtitle {
        font-size: 14px;
        color: var(--text-secondary);
        margin: 0;
    }

    .stat-card {
        border: 1px solid var(--border);
        background: var(--bg-primary);
        border-radius: 12px;
        transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
    
    .stat-card:hover {
        transform: translateY(-4px);
        box-shadow: var(--shadow-lg);
        border-color: var(--primary-border);
    }

    .stat-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .stat-main {
        display: flex;
        flex-direction: column;
    }

    .stat-label {
        font-size: 12px;
        color: var(--text-secondary);
        margin-bottom: 4px;
    }

    .stat-value-group {
        display: flex;
        align-items: baseline;
        gap: 4px;
    }

    .stat-value {
        font-size: 24px;
        font-weight: 500;
        color: var(--text-primary);
        font-family: var(--sans);
    }

    .stat-unit {
        font-size: 12px;
        color: var(--text-tertiary);
    }

    .stat-icon {
        width: 40px;
        height: 40px;
        border-radius: 10px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 20px;
    }

    .stat-icon.primary { background: var(--info-light); color: var(--info); }
    .stat-icon.success { background: var(--success-light); color: var(--success); }
    .stat-icon.warning { background: var(--warning-light); color: var(--warning); }
    .stat-icon.danger { background: var(--error-light); color: var(--error); }

    .stat-footer {
        margin-top: 12px;
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 12px;
        padding-top: 12px;
        border-top: 1px solid var(--border-light);
    }

    .stat-trend {
        display: inline-flex;
        align-items: center;
        gap: 2px;
        font-weight: 500;
    }

    .stat-trend.up { color: var(--success); }
    .stat-trend.down { color: var(--error); }
    .stat-trend.neutral { color: var(--text-tertiary); }

    .stat-period {
        color: var(--text-quaternary);
    }

    .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        height: 32px;
    }

    .card-title-group {
        display: flex;
        flex-direction: column;
    }

    .card-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
    }

    .card-desc {
        font-size: 12px;
        color: var(--text-tertiary);
        margin-top: 0;
    }

    .chart-container {
        height: 300px;
        width: 100%;
        padding-top: 20px;
    }

    .d3-chart {
        width: 100%;
        height: 100%;
    }

    .activity-list {
        display: flex;
        flex-direction: column;
        gap: 16px;
        padding: 8px 0;
    }

    .activity-item {
        display: flex;
        gap: 12px;
        position: relative;
    }

    .activity-item:not(:last-child)::after {
        content: '';
        position: absolute;
        left: 4px;
        top: 16px;
        bottom: -16px;
        width: 1px;
        background-color: var(--border);
    }

    .activity-dot {
        width: 9px;
        height: 9px;
        border-radius: 50%;
        margin-top: 4px;
        flex-shrink: 0;
        background: var(--border-strong);
        z-index: 1;
    }

    .activity-dot.success { background: var(--success); }
    .activity-dot.primary { background: var(--primary); }
    .activity-dot.danger { background: var(--error); }
    .activity-dot.warning { background: var(--warning); }
    .activity-dot.info { background: var(--text-tertiary); }

    .activity-info {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .activity-msg {
        font-size: 13px;
        color: var(--text-secondary);
        line-height: 1.4;
    }

    .activity-time {
        font-size: 11px;
        color: var(--text-quaternary);
        font-family: var(--mono);
    }

    .table-row {
        margin-top: 20px;
    }
</style>
