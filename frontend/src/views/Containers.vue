<template>
    <div class="containers-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">容器管理</h2>
                <p class="page-subtitle">管理和监控您的 Docker 容器实例。</p>
            </div>
            <div class="header-actions">
                <el-button :icon="Refresh" @click="fetchContainers">刷新列表</el-button>
                <el-button type="accent" :icon="Plus">部署容器</el-button>
            </div>
        </header>

        <!-- Search & Filter -->
        <el-card shadow="never" class="filter-card">
            <el-form :model="searchForm" inline class="search-form">
                <el-form-item label="容器信息">
                    <el-input 
                        v-model="searchForm.name" 
                        placeholder="容器名称或镜像" 
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
                        <el-option label="运行中" value="running" />
                        <el-option label="已停止" value="stopped" />
                        <el-option label="重启中" value="restarting" />
                        <el-option label="已暂停" value="paused" />
                    </el-select>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="handleSearch">查询</el-button>
                    <el-button @click="resetForm">重置</el-button>
                </el-form-item>
            </el-form>
        </el-card>

        <!-- Container Grid -->
        <el-card shadow="never" class="table-card">
            <el-table 
                :data="displayContainers" 
                style="width: 100%" 
                v-loading="loading"
                header-cell-class-name="table-header-cell"
            >
                <el-table-column prop="name" label="容器名称" min-width="180">
                    <template #default="{ row }">
                        <div class="container-info-cell">
                            <span class="container-name">{{ row.name }}</span>
                            <span class="container-id">{{ row.id.substring(0, 12) }}</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="image" label="镜像" min-width="180">
                    <template #default="{ row }">
                        <el-tag size="small" effect="plain">{{ row.image }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="120">
                    <template #default="{ row }">
                        <el-tag :type="getStatusType(row.status)" size="small">
                            {{ formatStatus(row.status) }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="ip" label="IP 地址" width="140" />
                <el-table-column prop="ports" label="端口映射" width="180" />
                <el-table-column prop="uptime" label="运行时间" width="120" />
                <el-table-column label="操作" width="180" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleAction(row, 'start')" :disabled="row.status === 'running'">启动</el-button>
                        <el-button link type="primary" @click="handleAction(row, 'stop')" :disabled="row.status === 'stopped'">停止</el-button>
                        <el-dropdown trigger="click" style="margin-left: 12px">
                            <el-button link type="primary">
                                <el-icon><MoreFilled /></el-icon>
                            </el-button>
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item @click="handleAction(row, 'restart')">重启</el-dropdown-item>
                                    <el-dropdown-item>查看日志</el-dropdown-item>
                                    <el-dropdown-item>终端 (Exec)</el-dropdown-item>
                                    <el-dropdown-item divided class="danger-text">删除容器</el-dropdown-item>
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
</template>

<script setup>
    import { ref, computed, onMounted } from 'vue'
    import { Search, Plus, Refresh, MoreFilled } from '@element-plus/icons-vue'
    import { containerApi } from '../api/index.js'
    import { ElMessage } from 'element-plus'

    const containers = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)
    const loading = ref(false)
    const displayContainers = computed(() => {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return containers.value.slice(start, end)
    })

    const searchForm = ref({
        name: '',
        status: ''
    })

    const fetchContainers = async () => {
        loading.value = true
        try {
            const params = {
                name: searchForm.value.name,
                status: searchForm.value.status
            }
            const res = await containerApi.getContainers(params)
            if (res.code === 200) {
                containers.value = res.data
                total.value = res.data.length
            }
        } catch (error) {
            console.error('Error fetching containers:', error)
        } finally {
            loading.value = false
        }
    }

    const handleSearch = () => {
        fetchContainers()
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
            status: ''
        }
        fetchContainers()
    }

    const formatStatus = (status) => {
        const map = {
            running: '运行中',
            stopped: '已停止',
            restarting: '重启中',
            paused: '已暂停'
        }
        return map[status] || status
    }

    const getStatusType = (status) => {
        const map = {
            running: 'success',
            stopped: 'info',
            restarting: 'warning',
            paused: 'danger'
        }
        return map[status] || ''
    }

    const handleAction = (row, action) => {
        ElMessage.success(`${action} 容器 ${row.name} 成功`)
    }

    const handleSizeChange = (val) => {
        pageSize.value = val
        fetchContainers()
    }

    const handleCurrentChange = (val) => {
        currentPage.value = val
        fetchContainers()
    }

    onMounted(fetchContainers)
</script>

<style scoped>
    .containers-container {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .container-info-cell {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .container-name {
        font-weight: 600;
        color: var(--text-primary);
    }

    .container-id {
        font-size: 0.75rem;
        color: var(--text-quaternary);
        font-family: var(--mono);
    }

    .pagination-container {
        margin-top: 24px;
        display: flex;
        justify-content: flex-end;
    }

    .danger-text {
        color: var(--error) !important;
    }

    :deep(.table-header-cell) {
        background-color: var(--bg-secondary) !important;
    }
</style>