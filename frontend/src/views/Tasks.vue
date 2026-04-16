<template>
    <div class="tasks-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">自动化任务</h2>
                <p class="page-subtitle">编排、调度及监控系统自动化运维任务。</p>
            </div>
            <div class="header-actions">
                <el-button type="accent" :icon="Plus">新建任务</el-button>
                <el-button :icon="Calendar">调度历史</el-button>
            </div>
        </header>

        <!-- Search & Filter -->
        <el-card shadow="never" class="filter-card">
            <el-form :model="searchForm" inline class="search-form">
                <el-form-item label="任务名称">
                    <el-input 
                        v-model="searchForm.name" 
                        placeholder="任务名称或 ID" 
                        clearable 
                        :prefix-icon="Search"
                        @clear="handleSearch"
                        @input="handleInput"
                    />
                </el-form-item>
                <el-form-item label="任务状态">
                    <el-select 
                        v-model="searchForm.status" 
                        placeholder="全部状态" 
                        clearable 
                        style="width: 140px"
                        @change="handleSearch"
                    >
                        <el-option label="待执行" value="pending" />
                        <el-option label="运行中" value="running" />
                        <el-option label="已完成" value="completed" />
                        <el-option label="已失败" value="failed" />
                    </el-select>
                </el-form-item>
                <el-form-item label="任务类型">
                    <el-select 
                        v-model="searchForm.type" 
                        placeholder="全部类型" 
                        clearable 
                        style="width: 140px"
                        @change="handleSearch"
                    >
                        <el-option label="数据备份" value="backup" />
                        <el-option label="系统更新" value="update" />
                        <el-option label="安全扫描" value="scan" />
                        <el-option label="日志清理" value="cleanup" />
                    </el-select>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="handleSearch">查询</el-button>
                    <el-button @click="resetForm">重置</el-button>
                </el-form-item>
            </el-form>
        </el-card>

        <!-- Task Grid -->
        <el-card shadow="never" class="table-card">
            <el-table 
                :data="displayTasks" 
                style="width: 100%" 
                v-loading="loading"
                header-cell-class-name="table-header-cell"
            >
                <el-table-column prop="name" label="任务信息" min-width="220">
                    <template #default="{ row }">
                        <div class="task-info-cell">
                            <span class="task-name">{{ row.name }}</span>
                            <div class="task-tags">
                                <el-tag size="small" effect="plain">{{ row.type }}</el-tag>
                                <span class="task-id">#{{ row.id }}</span>
                            </div>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="120">
                    <template #default="{ row }">
                        <span class="status-indicator" :class="'status-' + row.status">
                            {{ formatStatus(row.status) }}
                        </span>
                    </template>
                </el-table-column>
                <el-table-column prop="progress" label="执行进度" width="180">
                    <template #default="{ row }">
                        <div v-if="row.status === 'running'" class="progress-cell">
                            <el-progress 
                                :percentage="row.progress || 45" 
                                :stroke-width="6"
                                striped
                                striped-flow
                            />
                        </div>
                        <span v-else class="data-value text-tertiary">-</span>
                    </template>
                </el-table-column>
                <el-table-column prop="createdAt" label="创建时间" width="180">
                    <template #default="{ row }">
                        <span class="data-value">{{ row.createdAt }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="executedAt" label="最后执行" width="180">
                    <template #default="{ row }">
                        <span class="data-value">{{ row.executedAt || '从未执行' }}</span>
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="160" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="handleExecute(row)" :disabled="row.status === 'running'">
                            立即执行
                        </el-button>
                        <el-button link type="primary" @click="handleView(row)">详情</el-button>
                        <el-dropdown trigger="click" style="margin-left: 12px">
                            <el-button link type="primary">
                                <el-icon><MoreFilled /></el-icon>
                            </el-button>
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item>编辑任务</el-dropdown-item>
                                    <el-dropdown-item>复制任务</el-dropdown-item>
                                    <el-dropdown-item>查看日志</el-dropdown-item>
                                    <el-dropdown-item divided class="danger-text" @click="handleCancel(row)">删除任务</el-dropdown-item>
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
    import { Search, Plus, Calendar, MoreFilled } from '@element-plus/icons-vue'
    import { taskApi } from '../api/index.js'

    const tasks = ref([])
    const currentPage = ref(1)
    const pageSize = ref(10)
    const total = ref(0)
    const loading = ref(false)
    const searchForm = ref({
        name: '',
        status: '',
        type: ''
    })
    const displayTasks = computed(() => {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return tasks.value.slice(start, end)
    })

    const formatStatus = (status) => {
        const map = {
            pending: '待执行',
            running: '运行中',
            completed: '已完成',
            failed: '已失败'
        }
        return map[status] || status
    }

    const fetchTasks = async () => {
        loading.value = true
        try {
            const params = {
                name: searchForm.value.name,
                status: searchForm.value.status,
                type: searchForm.value.type
            }
            const res = await taskApi.getTasks(params)
            if (res.code === 200) {
                tasks.value = res.data
                total.value = res.data.length
            } else {
                // Fallback for non-standard response
                tasks.value = res.map(t => ({
                    ...t,
                    status: t.status === '已完成' ? 'completed' : (t.status === '运行中' ? 'running' : 'pending')
                }))
                total.value = tasks.value.length
            }
        } catch (error) {
            console.error('Error fetching tasks:', error)
            // Mock data fallback
            tasks.value = [
                { id: 't-1001', name: '全量数据库备份 (MySQL-Prod)', status: 'completed', type: 'backup', createdAt: '2024-04-10 02:00:00', executedAt: '2024-04-10 02:05:12' },
                { id: 't-1002', name: '系统内核安全更新 (Ubuntu-Group-A)', status: 'running', type: 'update', createdAt: '2024-04-10 10:00:00', executedAt: '2024-04-10 10:00:00', progress: 65 },
                { id: 't-1003', name: '过期日志自动清理 (Nginx-Cluster)', status: 'pending', type: 'cleanup', createdAt: '2024-04-10 11:30:00', executedAt: '' },
                { id: 't-1004', name: 'Web 静态资源同步 (OSS-Sync)', status: 'failed', type: 'sync', createdAt: '2024-04-09 18:00:00', executedAt: '2024-04-09 18:02:45' }
            ]
            total.value = tasks.value.length
        } finally {
            loading.value = false
        }
    }

    onMounted(fetchTasks)

    const handleSearch = () => {
        currentPage.value = 1
        fetchTasks()
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
            type: ''
        }
        fetchTasks()
    }

    const handleView = (row) => {
        // console.log('View task:', row.id)
    }

    const handleExecute = (row) => {
        // console.log('Execute task:', row.id)
    }

    const handleCancel = (row) => {
        // console.log('Cancel task:', row.id)
    }

    const handleCurrentChange = (val) => {
        currentPage.value = val
        fetchTasks()
    }

    const handleSizeChange = (val) => {
        pageSize.value = val
        fetchTasks()
    }
</script>

<style scoped>
    .tasks-container {
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

    .filter-card {
        margin-bottom: 4px;
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
        border: none;
    }

    .task-info-cell {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .task-name {
        font-weight: 500;
        color: var(--text-primary);
    }

    .task-tags {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .task-id {
        font-size: 12px;
        color: var(--text-quaternary);
        font-family: var(--mono);
    }

    .progress-cell {
        width: 100%;
    }

    .pagination-container {
        margin-top: 16px;
        display: flex;
        justify-content: flex-end;
    }

    .danger-text {
        color: var(--error) !important;
    }

    :deep(.table-header-cell) {
        background-color: var(--bg-tertiary) !important;
        color: var(--text-secondary);
        font-weight: 500;
    }

    .status-running::before {
        background: var(--primary);
        box-shadow: 0 0 0 2px rgba(0, 112, 204, 0.2);
        animation: pulse 2s infinite;
    }

    @keyframes pulse {
        0% { opacity: 1; }
        50% { opacity: 0.4; }
        100% { opacity: 1; }
    }
</style>