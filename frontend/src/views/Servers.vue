<template>
    <div class="servers-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">主机管理</h2>
                <p class="page-subtitle">管理和监控您的物理机、虚拟机及云主机资源。</p>
            </div>
            <div class="header-actions">
                <el-button type="primary" :icon="Plus" @click="handleAddServer">新增主机</el-button>
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
                        :data="servers" 
                        style="width: 100%" 
                        v-loading="loading"
                        header-cell-class-name="table-header-cell"
                    >
                        <el-table-column prop="name" label="服务器名称" min-width="200">
                            <template #default="{ row }">
                                <div class="server-name-cell">
                                    <span class="server-name">{{ row.name }}</span>
                                    <span class="server-id">{{ row.id }}</span>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column label="IP 地址" width="160">
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
                        <el-table-column prop="status" label="状态" width="100">
                            <template #default="{ row }">
                                <span class="status-indicator" :class="'status-' + row.status">
                                    {{ formatStatus(row.status) }}
                                </span>
                            </template>
                        </el-table-column>
                        <el-table-column prop="cpu" label="CPU" width="120">
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
                        <el-table-column prop="memory" label="内存" width="120">
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
                            :page-sizes="[10, 20, 50]"
                            layout="total, prev, pager, next"
                            :total="total"
                            @size-change="handleSizeChange"
                            @current-change="handleCurrentChange"
                        />
                    </div>
                </el-card>
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
    import { ref, onMounted } from 'vue'
    import { Search, Plus, Download, ArrowDown, Folder, Monitor } from '@element-plus/icons-vue'
    import { ElMessage } from 'element-plus'
    import { serverApi } from '../api/index.js'

    const servers = ref([])
    const loading = ref(false)
    const groupSearch = ref('')
    const selectedGroup = ref('全部资产')
    const searchForm = ref({
        name: '',
        status: '',
        region: ''
    })
    const currentPage = ref(1)
    const pageSize = ref(10)
    const total = ref(0)
    const submitting = ref(false)

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

    const formatStatus = (status) => {
        const map = {
            online: '运行中',
            offline: '已离线',
            maintenance: '维护中'
        }
        return map[status] || status
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
        console.log('Group clicked:', data)
        selectedGroup.value = data.label
        fetchServers()
    }

    const handleView = (row) => {
        console.log('View server:', row)
    }

    const handleRestart = (row) => {
        console.log('Restart server:', row)
    }

    const handleShutdown = (row) => {
        console.log('Shutdown server:', row)
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
    .servers-container {
        display: flex;
        flex-direction: column;
        gap: 16px;
        height: 100%;
    }

    .page-header {
        display: flex;
        align-items: flex-start;
        gap: 40px;
        margin-bottom: 4px;
    }

    .header-actions {
        display: flex;
        gap: 12px;
    }

    .main-layout {
        display: flex;
        gap: 16px;
        flex: 1;
        min-height: 0;
    }

    .grouping-sidebar {
        width: 240px;
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
        gap: 20px;
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
        padding: 16px;
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
</style>