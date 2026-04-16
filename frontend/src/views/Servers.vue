<template>
    <div class="servers-container">
        <div v-if="!detailVisible" class="list-view">
            <header class="page-header">
                <div class="header-content">
                    <div style="display: flex; align-items: center; gap: 12px">
                        <h2 class="page-title">主机管理</h2>
                        <span class="accent-dot"></span>
                    </div>
                    <p class="page-subtitle">管理和监控您的物理机、虚拟机及云主机资源。</p>
                </div>
                <div class="header-actions">
                    <el-button type="accent" :icon="Plus" @click="handleAddServer">新增主机</el-button>
                    <el-button :icon="Download">导出数据</el-button>
                </div>
            </header>

            <div class="main-layout">
            <!-- Left: Asset Grouping -->
            <el-card shadow="never" class="grouping-sidebar">
                <template #header>
                    <div class="sidebar-header">
                        <span class="sidebar-title">资产分组</span>
                        <el-button link type="primary" :icon="Plus" @click="handleAddGroup">新建</el-button>
                    </div>
                </template>
                <el-input
                    v-model="groupSearch"
                    placeholder="搜索分组..."
                    :prefix-icon="Search"
                    size="small"
                    class="group-search"
                    clearable
                />
                <el-tree
                    :data="groupData"
                    :props="defaultProps"
                    @node-click="handleGroupClick"
                    highlight-current
                    default-expand-all
                    class="group-tree"
                >
                    <template #default="{ node, data }">
                        <div class="custom-tree-node">
                            <el-icon v-if="data.children" class="node-icon"><Folder /></el-icon>
                            <el-icon v-else class="node-icon"><Monitor /></el-icon>
                            <span class="node-label">{{ node.label }}</span>
                            <span v-if="data.count" class="node-count">{{ data.count }}</span>
                        </div>
                    </template>
                </el-tree>
            </el-card>

            <!-- Right: Host Assets -->
            <div class="content-area">
                <!-- Search & Filter Bar -->
                <el-card shadow="never" class="filter-card">
                    <el-form :model="searchForm" inline class="search-form">
                        <el-form-item label="资源名称">
                            <el-input 
                                v-model="searchForm.name" 
                                placeholder="输入名称或 IP" 
                                clearable 
                                :prefix-icon="Search"
                                @clear="handleSearch"
                                @input="handleInput"
                            />
                        </el-form-item>
                        <el-form-item label="运行状态">
                            <el-select 
                                v-model="searchForm.status" 
                                placeholder="全部状态" 
                                clearable 
                                style="width: 140px"
                                @change="handleSearch"
                            >
                                <el-option label="运行中" value="online" />
                                <el-option label="已离线" value="offline" />
                                <el-option label="维护中" value="maintenance" />
                            </el-select>
                        </el-form-item>
                        <el-form-item>
                            <el-button type="primary" @click="handleSearch">查询</el-button>
                            <el-button @click="resetForm">重置</el-button>
                        </el-form-item>
                    </el-form>
                </el-card>

                <!-- Data Grid -->
                <el-card shadow="never" class="table-card">
                    <el-table 
                        :data="displayServers" 
                        style="width: 100%" 
                        v-loading="loading"
                        header-cell-class-name="table-header-cell"
                    >
                        <el-table-column prop="name" label="主机名" width="160">
                            <template #default="{ row }">
                                <div class="server-name-cell">
                                    <span class="server-name font-bold">{{ row.name }}</span>
                                    <span class="server-id data-value">{{ row.id }}</span>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column label="IP 地址" min-width="150">
                            <template #default="{ row }">
                                <div class="ip-cell">
                                    <div class="ip-item">
                                        <span class="ip-label">公</span>
                                        <span class="data-value">{{ row.publicIp }}</span>
                                    </div>
                                    <div class="ip-item">
                                        <span class="ip-label">内</span>
                                        <span class="data-value">{{ row.privateIp }}</span>
                                    </div>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column prop="status" label="状态" width="90">
                            <template #default="{ row }">
                                <span class="status-indicator" :class="'status-' + row.status">
                                    {{ formatStatus(row.status) }}
                                </span>
                            </template>
                        </el-table-column>
                        <el-table-column prop="cpu" label="CPU" width="110">
                            <template #default="{ row }">
                                <div class="usage-cell">
                                    <el-progress 
                                        :percentage="row.cpu" 
                                        :stroke-width="6"
                                        :status="row.cpu > 80 ? 'exception' : ''"
                                        :show-text="false"
                                    />
                                    <span class="usage-text">{{ row.cpu }}%</span>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column prop="memory" label="内存" width="110">
                            <template #default="{ row }">
                                <div class="usage-cell">
                                    <el-progress 
                                        :percentage="row.memory" 
                                        :stroke-width="6"
                                        :status="row.memory > 80 ? 'warning' : ''"
                                        :show-text="false"
                                    />
                                    <span class="usage-text">{{ row.memory }}%</span>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column label="操作" width="160" fixed="right">
                            <template #default="{ row }">
                                <el-button link type="primary" @click="handleView(row)">详情</el-button>
                                <el-dropdown trigger="click" style="margin-left: 12px">
                                    <el-button link type="primary">
                                        更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                                    </el-button>
                                    <template #dropdown>
                                        <el-dropdown-menu>
                                            <el-dropdown-item @click="handleRestart(row)">重启服务</el-dropdown-item>
                                            <el-dropdown-item>远程连接 (SSH)</el-dropdown-item>
                                            <el-dropdown-item divided class="danger-text" @click="handleShutdown(row)">关机</el-dropdown-item>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                            </template>
                        </el-table-column>
                    </el-table>

                    <div class="pagination-container">
                        <el-pagination
                            v-model:current-page="currentPage"
                            v-model:page-size="pageSize"
                            :page-sizes="[10, 20, 50, 100]"
                            layout="total, sizes, prev, pager, next, jumper"
                            :total="total"
                            @size-change="handleSizeChange"
                            @current-change="handleCurrentChange"
                        />
                    </div>
                </el-card>
            </div>
        </div>
    </div>

    <!-- Server Detail Full Page View -->
        <div v-else class="detail-view">
            <header class="page-header detail-header">
                <div class="header-content">
                    <div class="header-nav">
                        <el-button link @click="handleBack" class="back-btn">
                            <el-icon><ArrowLeft /></el-icon> 返回列表
                        </el-button>
                        <el-divider direction="vertical" />
                        <div class="detail-title-group">
                            <h2 class="page-title">{{ selectedServer?.name }}</h2>
                            <span class="status-indicator" :class="'status-' + selectedServer?.status">
                                {{ formatStatus(selectedServer?.status) }}
                            </span>
                        </div>
                    </div>
                </div>
                <div class="header-actions">
                    <el-button @click="handleRestart(selectedServer)" :icon="Refresh">重启服务</el-button>
                    <el-button type="danger" plain @click="handleShutdown(selectedServer)" :icon="SwitchButton">关机</el-button>
                    <el-button type="primary" :icon="Monitor">远程连接 (SSH)</el-button>
                </div>
            </header>

            <div class="detail-container">
                <div v-if="selectedServer" class="detail-content">
                    <!-- Quick Stats Cards -->
                    <el-row :gutter="12" class="stats-row">
                        <el-col :span="6">
                            <el-card shadow="never" class="stat-card-new">
                                <div class="stat-label">CPU 使用率</div>
                                <div class="stat-main">
                                    <span class="stat-value">{{ selectedServer.cpu }}%</span>
                                    <el-progress 
                                        type="circle" 
                                        :percentage="selectedServer.cpu" 
                                        :width="40" 
                                        :stroke-width="4"
                                        :status="selectedServer.cpu > 80 ? 'exception' : ''"
                                    />
                                </div>
                            </el-card>
                        </el-col>
                        <el-col :span="6">
                            <el-card shadow="never" class="stat-card-new">
                                <div class="stat-label">内存使用率</div>
                                <div class="stat-main">
                                    <span class="stat-value">{{ selectedServer.memory }}%</span>
                                    <el-progress 
                                        type="circle" 
                                        :percentage="selectedServer.memory" 
                                        :width="40" 
                                        :stroke-width="4"
                                        :status="selectedServer.memory > 80 ? 'warning' : ''"
                                    />
                                </div>
                            </el-card>
                        </el-col>
                        <el-col :span="6">
                            <el-card shadow="never" class="stat-card-new">
                                <div class="stat-label">磁盘占用</div>
                                <div class="stat-main">
                                    <span class="stat-value">45%</span>
                                    <el-progress type="circle" :percentage="45" :width="40" :stroke-width="4" />
                                </div>
                            </el-card>
                        </el-col>
                        <el-col :span="6">
                            <el-card shadow="never" class="stat-card-new">
                                <div class="stat-label">网络流量 (In/Out)</div>
                                <div class="stat-main">
                                    <div class="stat-value-group">
                                        <div class="stat-sub-value"><el-icon><Download /></el-icon> 1.2MB</div>
                                        <div class="stat-sub-value"><el-icon><Plus /></el-icon> 850KB</div>
                                    </div>
                                    <div class="stat-chart-mini" id="mini-net-chart"></div>
                                </div>
                            </el-card>
                        </el-col>
                    </el-row>

                    <el-row :gutter="12" class="detail-main-row">
                        <el-col :span="16">
                            <el-card shadow="never" class="detail-tabs-card">
                                <el-tabs v-model="activeTab" class="detail-tabs">
                                    <el-tab-pane label="实时监控" name="monitoring">
                                        <div class="monitoring-grid">
                                            <div class="chart-item">
                                                <div class="chart-header">
                                                    <span class="chart-title">CPU 负载历史</span>
                                                    <el-tag size="small" type="info" effect="plain">实时</el-tag>
                                                </div>
                                                <div id="cpu-load-chart" class="main-chart"></div>
                                            </div>
                                            <el-divider />
                                            <div class="chart-item">
                                                <div class="chart-header">
                                                    <span class="chart-title">内存占用趋势</span>
                                                    <el-tag size="small" type="info" effect="plain">实时</el-tag>
                                                </div>
                                                <div id="mem-usage-chart" class="main-chart"></div>
                                            </div>
                                        </div>
                                    </el-tab-pane>
                                    <el-tab-pane label="操作审计" name="logs">
                                        <div class="log-container">
                                            <div v-for="(log, index) in mockLogs" :key="index" class="log-line">
                                                <span class="log-time">[{{ log.time }}]</span>
                                                <span class="log-user">{{ log.user }}</span>
                                                <span class="log-action">{{ log.action }}</span>
                                                <span class="log-status" :class="log.status">{{ log.statusText }}</span>
                                            </div>
                                        </div>
                                    </el-tab-pane>
                                </el-tabs>
                            </el-card>
                        </el-col>
                        <el-col :span="8">
                            <el-card shadow="never" header="基础信息" class="info-card">
                                <el-descriptions :column="1" border size="small">
                                    <el-descriptions-item label="主机名称">{{ selectedServer.name }}</el-descriptions-item>
                                    <el-descriptions-item label="实例 ID"><span class="data-value">{{ selectedServer.id }}</span></el-descriptions-item>
                                    <el-descriptions-item label="公网 IP"><span class="data-value">{{ selectedServer.publicIp }}</span></el-descriptions-item>
                                    <el-descriptions-item label="内网 IP"><span class="data-value">{{ selectedServer.privateIp }}</span></el-descriptions-item>
                                    <el-descriptions-item label="操作系统">Ubuntu 22.04 LTS</el-descriptions-item>
                                    <el-descriptions-item label="内核版本">5.15.0-101-generic</el-descriptions-item>
                                    <el-descriptions-item label="所属区域">{{ selectedServer.region === 'shanghai' ? '华东 (上海)' : selectedServer.region }}</el-descriptions-item>
                                    <el-descriptions-item label="创建时间">2024-03-15 10:00:00</el-descriptions-item>
                                </el-descriptions>
                            </el-card>
                        </el-col>
                    </el-row>
                </div>
            </div>
        </div>

        <!-- New Group Dialog -->
        <el-dialog
            v-model="groupDialogVisible"
            title="新建资产分组"
            width="400px"
            destroy-on-close
        >
            <el-form
                ref="groupFormRef"
                :model="groupForm"
                :rules="groupRules"
                label-width="80px"
                label-position="top"
            >
                <el-form-item label="分组名称" prop="name">
                    <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
                </el-form-item>
                <el-form-item label="上级分组" prop="parentId">
                    <el-tree-select
                        v-model="groupForm.parentId"
                        :data="groupData"
                        :props="{ label: 'label', value: 'label' }"
                        placeholder="请选择上级分组 (可选)"
                        check-strictly
                        clearable
                        style="width: 100%"
                    />
                </el-form-item>
            </el-form>
            <template #footer>
                <div class="dialog-footer">
                    <el-button @click="groupDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="submitGroup">确定</el-button>
                </div>
            </template>
        </el-dialog>

        <!-- New Server Dialog -->
        <el-dialog
            v-model="serverDialogVisible"
            title="新增主机"
            width="500px"
            destroy-on-close
        >
            <el-form
                ref="serverFormRef"
                :model="serverForm"
                :rules="serverRules"
                label-width="100px"
                label-position="top"
            >
                <el-form-item label="主机名称" prop="name">
                    <el-input v-model="serverForm.name" placeholder="例如: Web-Server-01" />
                </el-form-item>
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item label="公网 IP" prop="publicIp">
                            <el-input v-model="serverForm.publicIp" placeholder="例如: 1.2.3.4" />
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item label="内网 IP" prop="privateIp">
                            <el-input v-model="serverForm.privateIp" placeholder="例如: 10.0.0.1" />
                        </el-form-item>
                    </el-col>
                </el-row>
                <el-row :gutter="20">
                    <el-col :span="12">
                        <el-form-item label="所属区域" prop="region">
                            <el-select v-model="serverForm.region" placeholder="请选择区域" style="width: 100%">
                                <el-option label="华东 (上海)" value="shanghai" />
                                <el-option label="华北 (北京)" value="beijing" />
                                <el-option label="华南 (广州)" value="guangzhou" />
                            </el-select>
                        </el-form-item>
                    </el-col>
                    <el-col :span="12">
                        <el-form-item label="资产分组" prop="group">
                            <el-tree-select
                                v-model="serverForm.group"
                                :data="groupData"
                                :props="{ label: 'label', value: 'label' }"
                                placeholder="请选择分组"
                                check-strictly
                                style="width: 100%"
                            />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <div class="dialog-footer">
                    <el-button @click="serverDialogVisible = false">取消</el-button>
                    <el-button type="primary" :icon="Plus" :loading="submitting" @click="submitServer">立即创建</el-button>
                </div>
            </template>
        </el-dialog>

    </div>
</template>

<script setup>
    import { ref, computed, onMounted, nextTick, watch } from 'vue'
    import { Search, Plus, Download, ArrowDown, Folder, Monitor, Refresh, SwitchButton, QuestionFilled, ArrowLeft } from '@element-plus/icons-vue'
    import { ElMessage } from 'element-plus'
    import { serverApi } from '../api/index.js'
    import * as d3 from 'd3'

    const servers = ref([])
    const currentPage = ref(1)
    const pageSize = ref(10)
    const total = ref(0)
    const displayServers = computed(() => {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return servers.value.slice(start, end)
    })
    const submitting = ref(false)
    const loading = ref(false)
    const groupSearch = ref('')
    const selectedGroup = ref('全部资产')
    const searchForm = ref({
        name: '',
        status: '',
        region: ''
    })

    // Detail Drawer logic
    const detailVisible = ref(false)
    const selectedServer = ref(null)
    const activeTab = ref('monitoring')
    const mockLogs = ref([
        { time: '2024-04-15 10:30:01', user: 'admin', action: '重启实例', status: 'success', statusText: '成功' },
        { time: '2024-04-15 09:15:22', user: 'system', action: '健康检查', status: 'success', statusText: '正常' },
        { time: '2024-04-14 22:00:00', user: 'admin', action: '修改安全组规则', status: 'success', statusText: '成功' },
        { time: '2024-04-14 18:45:10', user: 'dev_user', action: '登录 SSH', status: 'warning', statusText: '异常登录' },
    ])

    const handleView = (row) => {
        selectedServer.value = row
        detailVisible.value = true
        // Wait for view to switch before drawing charts
        nextTick(() => {
            initCharts()
        })
    }

    const handleBack = () => {
        detailVisible.value = false
        selectedServer.value = null
    }

    const initCharts = () => {
        const drawLineChart = (id, color) => {
            const element = document.getElementById(id)
            if (!element) return
            
            // Clear previous
            d3.select(element).selectAll('*').remove()
            
            const width = element.clientWidth
            const height = element.clientHeight
            const margin = { top: 10, right: 10, bottom: 20, left: 30 }
            
            const svg = d3.select(element)
                .append('svg')
                .attr('width', width)
                .attr('height', height)
                .append('g')
                .attr('transform', `translate(${margin.left},${margin.top})`)
            
            const data = Array.from({ length: 20 }, (_, i) => ({
                x: i,
                y: Math.random() * 100
            }))
            
            const x = d3.scaleLinear()
                .domain([0, 19])
                .range([0, width - margin.left - margin.right])
            
            const y = d3.scaleLinear()
                .domain([0, 100])
                .range([height - margin.top - margin.bottom, 0])
            
            svg.append('g')
                .attr('transform', `translate(0,${height - margin.top - margin.bottom})`)
                .call(d3.axisBottom(x).ticks(5).tickFormat(''))
                .attr('color', '#e2e8f0')
            
            svg.append('g')
                .call(d3.axisLeft(y).ticks(5))
                .attr('color', '#e2e8f0')
            
            const line = d3.line()
                .x(d => x(d.x))
                .y(d => y(d.y))
                .curve(d3.curveMonotoneX)
            
            svg.append('path')
                .datum(data)
                .attr('fill', 'none')
                .attr('stroke', color)
                .attr('stroke-width', 2)
                .attr('d', line)
            
            // Add area
            const area = d3.area()
                .x(d => x(d.x))
                .y0(height - margin.top - margin.bottom)
                .y1(d => y(d.y))
                .curve(d3.curveMonotoneX)
            
            svg.append('path')
                .datum(data)
                .attr('fill', color)
                .attr('fill-opacity', 0.1)
                .attr('d', area)
        }

        drawLineChart('cpu-load-chart', '#0066FF')
        drawLineChart('mem-usage-chart', '#FF6B21')
    }

    // Re-init charts when tab changes to monitoring
    watch(activeTab, (newTab) => {
        if (newTab === 'monitoring') {
            nextTick(() => {
                initCharts()
            })
        }
    })

    const formatStatus = (status) => {
        const map = {
            online: '运行中',
            offline: '已离线',
            maintenance: '维护中'
        }
        return map[status] || status
    }

    // Group dialog logic
    const groupDialogVisible = ref(false)
    const groupFormRef = ref(null)
    const groupForm = ref({
        name: '',
        parentId: null
    })
    const groupRules = {
        name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }]
    }

    // Server dialog logic
    const serverDialogVisible = ref(false)
    const serverFormRef = ref(null)
    const serverForm = ref({
        name: '',
        publicIp: '',
        privateIp: '',
        region: 'shanghai',
        group: '全部资产'
    })
    const serverRules = {
        name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
        publicIp: [{ required: true, message: '请输入公网 IP', trigger: 'blur' }],
        privateIp: [{ required: true, message: '请输入内网 IP', trigger: 'blur' }],
        region: [{ required: true, message: '请选择区域', trigger: 'change' }],
        group: [{ required: true, message: '请选择资产分组', trigger: 'change' }]
    }

    const handleAddServer = () => {
        serverForm.value = {
            name: '',
            publicIp: '',
            privateIp: '',
            region: 'shanghai',
            group: selectedGroup.value !== '全部资产' ? selectedGroup.value : '全部资产'
        }
        serverDialogVisible.value = true
    }

    const submitServer = async () => {
        if (!serverFormRef.value) return
        await serverFormRef.value.validate(async (valid) => {
            if (valid) {
                submitting.value = true
                try {
                    const res = await serverApi.addServer(serverForm.value)
                    if (res.code === 200) {
                        ElMessage.success('主机创建成功')
                        serverDialogVisible.value = false
                        fetchServers()
                    }
                } catch (error) {
                    console.error('Error adding server:', error)
                    ElMessage.error('创建失败，请稍后重试')
                } finally {
                    submitting.value = false
                }
            }
        })
    }

    const handleAddGroup = () => {
        groupForm.value = { name: '', parentId: null }
        groupDialogVisible.value = true
    }

    const submitGroup = async () => {
        if (!groupFormRef.value) return
        await groupFormRef.value.validate((valid) => {
            if (valid) {
                const newGroup = { label: groupForm.value.name, count: 0 }
                if (groupForm.value.parentId) {
                    const findAndAdd = (nodes) => {
                        for (let node of nodes) {
                            if (node.label === groupForm.value.parentId) {
                                if (!node.children) node.children = []
                                node.children.push(newGroup)
                                return true
                            }
                            if (node.children && findAndAdd(node.children)) return true
                        }
                        return false
                    }
                    findAndAdd(groupData.value)
                } else {
                    groupData.value.push(newGroup)
                }
                groupDialogVisible.value = false
                ElMessage.success('分组创建成功')
            }
        })
    }

    const defaultProps = {
        children: 'children',
        label: 'label',
    }

    const groupData = ref([])

    const fetchGroups = async () => {
        try {
            const res = await serverApi.getGroups()
            if (res.code === 200) {
                groupData.value = res.data
            }
        } catch (error) {
            console.error('Error fetching groups:', error)
        }
    }

    const fetchServers = async () => {
        loading.value = true
        try {
            const params = {
                group: selectedGroup.value,
                name: searchForm.value.name,
                status: searchForm.value.status
            }
            const res = await serverApi.getServers(params)
            if (res.code === 200) {
                servers.value = res.data
                total.value = res.data.length
            }
        } catch (error) {
            console.error('Error fetching servers:', error)
            servers.value = []
            total.value = 0
        } finally {
            loading.value = false
        }
    }

    onMounted(() => {
        fetchGroups()
        fetchServers()
    })

    const handleSearch = () => {
        fetchServers()
    }

    let inputTimer = null
    const handleInput = () => {
        if (inputTimer) clearTimeout(inputTimer)
        inputTimer = setTimeout(() => {
            handleSearch()
        }, 500)
    }

    const resetForm = () => {
        searchForm.value = {
            name: '',
            status: '',
            region: ''
        }
        fetchServers()
    }

    const handleGroupClick = (data) => {
        // console.log('Group clicked:', data.label)
        selectedGroup.value = data.label
        fetchServers()
    }

    const handleSizeChange = (val) => {
        pageSize.value = val
        fetchServers()
    }

    const handleCurrentChange = (val) => {
        currentPage.value = val
        fetchServers()
    }
</script>

<style scoped>
    .detail-drawer :deep(.el-drawer__header) {
        margin-bottom: 0;
        padding: 24px;
        border-bottom: 1px solid var(--border);
    }

    .drawer-header {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .header-main {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .detail-title {
        font-family: var(--serif);
        font-style: italic;
        font-weight: 900;
        font-size: 24px;
        color: var(--text-primary);
    }

    .header-sub {
        display: flex;
        align-items: center;
        gap: 12px;
        font-size: 13px;
        color: var(--text-tertiary);
    }

    .detail-content {
        padding: 24px;
    }

    .stats-row {
        margin-bottom: 24px;
    }

    .stat-card {
        background: var(--bg-primary);
        border: 1px solid var(--border);
        border-radius: 12px;
        padding: 16px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .stat-label {
        font-size: 12px;
        color: var(--text-tertiary);
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .stat-value {
        font-family: var(--mono);
        font-size: 20px;
        font-weight: 700;
        color: var(--text-primary);
    }

    .detail-tabs {
        margin-top: 24px;
    }

    .monitoring-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
        margin-top: 12px;
    }

    .chart-card {
        border-radius: 12px;
    }

    .chart-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-size: 14px;
        font-weight: 600;
    }

    .main-chart {
        height: 240px;
        width: 100%;
    }

    .log-container {
        background: #1e293b;
        border-radius: 8px;
        padding: 16px;
        font-family: var(--mono);
        font-size: 12px;
        color: #e2e8f0;
        max-height: 400px;
        overflow-y: auto;
    }

    .log-line {
        margin-bottom: 8px;
        display: flex;
        gap: 12px;
    }

    .log-time { color: #94a3b8; }
    .log-user { color: #38bdf8; }
    .log-action { flex: 1; }
    .log-status.success { color: #4ade80; }
    .log-status.warning { color: #fbbf24; }

    .drawer-footer {
        padding: 16px 24px;
        border-top: 1px solid var(--border);
        display: flex;
        justify-content: flex-end;
        gap: 12px;
    }

    .servers-container {
        display: flex;
        flex-direction: column;
        gap: 16px;
        height: 100%;
    }

    .page-header {
        display: flex;
        align-items: flex-start;
        gap: 24px;
        margin-bottom: 4px;
    }

    .header-actions {
        display: flex;
        gap: 12px;
    }

    .main-layout {
        display: flex;
        gap: 12px;
        flex: 1;
        min-height: 0;
    }

    .grouping-sidebar {
        width: 200px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        border: 1px solid var(--border);
        background: var(--bg-primary);
        border-radius: 4px;
    }

    .sidebar-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 16px;
        border-bottom: 1px solid var(--border);
    }

    .sidebar-title {
        font-weight: 600;
        font-size: 14px;
        color: var(--text-primary);
    }

    .group-search {
        margin-bottom: 12px;
    }

    .group-tree {
        background: transparent;
        padding: 8px;
    }

    .custom-tree-node {
        display: flex;
        align-items: center;
        width: 100%;
        font-size: 13px;
    }

    .node-icon {
        margin-right: 8px;
        font-size: 14px;
        color: var(--text-quaternary);
    }

    .node-label {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .node-count {
        font-size: 11px;
        color: var(--text-quaternary);
        background: var(--bg-secondary);
        padding: 0 6px;
        border-radius: 10px;
        margin-left: 8px;
    }

    .content-area {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 12px;
        min-width: 0;
    }

    .filter-card {
        border: 1px solid var(--border);
        border-radius: 4px;
    }

    .search-form {
        display: flex;
        flex-wrap: wrap;
        gap: 16px;
    }

    :deep(.el-form-item) {
        margin-bottom: 0;
        margin-right: 0;
    }

    .table-card {
        border: 1px solid var(--border);
        flex: 1;
        display: flex;
        flex-direction: column;
        border-radius: 4px;
    }

    :deep(.el-card__body) {
        flex: 1;
        display: flex;
        flex-direction: column;
        padding: 12px;
    }

    .server-name-cell {
        display: flex;
        flex-direction: column;
    }

    .server-name {
        font-weight: 500;
        color: var(--text-primary);
    }

    .server-id {
        font-size: 12px;
        color: var(--text-quaternary);
        font-family: var(--mono);
    }

    .ip-cell {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .ip-item {
        display: flex;
        align-items: center;
        gap: 6px;
    }

    .ip-label {
        font-size: 10px;
        padding: 0 2px;
        border-radius: 2px;
        background: var(--bg-tertiary);
        color: var(--text-tertiary);
        line-height: 14px;
        height: 14px;
        min-width: 14px;
        text-align: center;
    }

    .usage-cell {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .usage-cell :deep(.el-progress) {
        flex: 1;
        margin-right: 0;
    }

    .usage-text {
        font-size: 12px;
        font-weight: 500;
        color: var(--text-secondary);
        width: 36px;
        text-align: right;
        font-family: var(--mono);
    }

    .pagination-container {
        margin-top: auto;
        padding-top: 16px;
        display: flex;
        justify-content: flex-end;
    }

    .danger-text {
        color: var(--error) !important;
    }

    :deep(.table-header-cell) {
        background-color: var(--bg-secondary) !important;
        color: var(--text-secondary);
        font-weight: 500;
    }

    /* Detail View Styles */
    .detail-view {
        display: flex;
        flex-direction: column;
        height: 100%;
        animation: fadeIn 0.3s ease-out;
    }

    @keyframes fadeIn {
        from { opacity: 0; transform: translateY(10px); }
        to { opacity: 1; transform: translateY(0); }
    }

    .detail-header {
        background: var(--bg-primary);
        padding: 12px 20px;
        margin: -12px -12px 16px -12px;
    }

    .header-nav {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .back-btn {
        font-size: 14px;
        color: var(--text-secondary);
        padding: 0;
    }

    .back-btn:hover {
        color: var(--primary);
    }

    .detail-title-group {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .detail-container {
        flex: 1;
        overflow-y: auto;
        padding-bottom: 20px;
    }

    .stat-card-new {
        border: 1px solid var(--border);
        border-radius: 8px;
        transition: all 0.3s;
    }

    .stat-card-new:hover {
        border-color: var(--primary-light);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    }

    .stat-label {
        font-size: 12px;
        color: var(--text-tertiary);
        margin-bottom: 8px;
    }

    .stat-main {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .stat-value {
        font-size: 24px;
        font-weight: 700;
        color: var(--text-primary);
        font-family: var(--mono);
    }

    .stat-value-group {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .stat-sub-value {
        font-size: 13px;
        color: var(--text-secondary);
        display: flex;
        align-items: center;
        gap: 4px;
    }

    .detail-main-row {
        margin-top: 16px;
    }

    .detail-tabs-card {
        border: 1px solid var(--border);
        border-radius: 8px;
        height: 100%;
    }

    .monitoring-grid {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .chart-item {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .chart-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .chart-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
    }

    .main-chart {
        height: 240px;
        width: 100%;
    }

    .info-card {
        border: 1px solid var(--border);
        border-radius: 8px;
    }

    :deep(.el-descriptions__label) {
        width: 100px;
        background-color: var(--bg-tertiary) !important;
        color: var(--text-secondary);
        font-weight: 500;
    }

    .log-container {
        background: var(--bg-tertiary);
        padding: 12px;
        border-radius: 4px;
        font-family: var(--mono);
        font-size: 12px;
        max-height: 500px;
        overflow-y: auto;
    }

    .log-line {
        padding: 4px 0;
        border-bottom: 1px solid var(--border-light);
        display: flex;
        gap: 12px;
    }

    .log-time { color: var(--text-quaternary); }
    .log-user { color: var(--primary); font-weight: 500; }
    .log-action { color: var(--text-primary); flex: 1; }
    .log-status.success { color: #10b981; }
    .log-status.warning { color: #f59e0b; }

</style>