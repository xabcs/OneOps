<template>
    <div class="logs-container">
        <!-- Info Alert -->
        <div class="info-alert compact">
            <el-icon class="info-icon"><InfoFilled /></el-icon>
            <span class="info-text">默认免费记录了90天的管控事件可供查询，建议 <el-link type="primary" :underline="false">创建跟踪</el-link> 实现更长时间存储</span>
        </div>

        <!-- Toolbar -->
        <div class="table-toolbar">
            <div class="toolbar-content">
                <el-form :model="searchForm" inline class="search-bar-form">
                    <el-form-item label="操作人">
                        <el-input 
                            v-model="searchForm.user" 
                            placeholder="用户名" 
                            clearable 
                            :prefix-icon="User"
                            style="width: 140px"
                            @clear="handleSearch"
                            @input="handleInput"
                        />
                    </el-form-item>
                    <el-form-item label="模块">
                        <el-select 
                            v-model="searchForm.module" 
                            placeholder="全部" 
                            clearable 
                            style="width: 120px"
                            @change="handleSearch"
                        >
                            <el-option label="主机管理" value="主机管理" />
                            <el-option label="自动化任务" value="自动化任务" />
                            <el-option label="监控中心" value="监控中心" />
                            <el-option label="系统设置" value="系统设置" />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="状态">
                        <el-select 
                            v-model="searchForm.status" 
                            placeholder="全部" 
                            clearable 
                            style="width: 90px"
                            @change="handleSearch"
                        >
                            <el-option label="成功" value="success" />
                            <el-option label="失败" value="failed" />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="范围">
                        <el-radio-group v-model="timeRange" size="small" class="compact-time-group" @change="handleSearch">
                            <el-radio-button label="1h">1h</el-radio-button>
                            <el-radio-button label="12h">12h</el-radio-button>
                            <el-radio-button label="1d">1d</el-radio-button>
                            <el-radio-button label="7d">7d</el-radio-button>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" :icon="Search" @click="handleSearch">查询</el-button>
                        <el-button @click="resetForm">重置</el-button>
                    </el-form-item>
                </el-form>
                <div class="toolbar-actions">
                    <el-button :icon="Download">导出</el-button>
                </div>
            </div>
            
            <div v-if="searchForm.user || searchForm.module" class="active-filter-tags">
                <el-tag v-if="searchForm.user" closable @close="searchForm.user=''; handleSearch()" size="small" class="mr-2">操作人: {{ searchForm.user }}</el-tag>
                <el-tag v-if="searchForm.module" closable @close="searchForm.module=''; handleSearch()" size="small">模块: {{ searchForm.module }}</el-tag>
            </div>
        </div>

        <!-- Log Table -->
        <el-table 
            :data="displayLogs" 
            style="width: 100%" 
            v-loading="loading"
            class="behavior-table"
        >
                <el-table-column prop="time" label="操作时间" width="170">
                    <template #default="{ row }">
                        <span class="data-value">{{ row.time }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="user" label="操作人" width="110">
                    <template #default="{ row }">
                        <div class="user-cell">
                            <el-avatar :size="24" :src="'https://api.dicebear.com/7.x/avataaars/svg?seed=' + row.user" />
                            <span>{{ row.user }}</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="action" label="操作行为" width="120" />
                <el-table-column prop="path" label="请求路径" min-width="180">
                    <template #default="{ row }">
                        <code class="code-text">{{ row.path }}</code>
                    </template>
                </el-table-column>
                <el-table-column prop="method" label="方法" width="80" align="center">
                    <template #default="{ row }">
                        <el-tag size="small" :type="getMethodType(row.method)" effect="dark">{{ row.method }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="params" label="请求参数" min-width="150" show-overflow-tooltip>
                    <template #default="{ row }">
                        <span class="json-text">{{ row.params }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="response" label="响应数据" min-width="150" show-overflow-tooltip>
                    <template #default="{ row }">
                        <span class="json-text">{{ row.response }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="duration" label="耗时(ms)" width="100" align="right">
                    <template #default="{ row }">
                        <span :class="getDurationClass(row.duration)">{{ row.duration }}ms</span>
                    </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="80" align="center">
                    <template #default="{ row }">
                        <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                            {{ row.status === 'success' ? '成功' : '失败' }}
                        </el-tag>
                    </template>
                </el-table-column>
            </el-table>

        <div class="pagination-container">
            <el-pagination
                v-model:current-page="currentPage"
                v-model:page-size="pageSize"
                :page-sizes="[10, 15, 20, 50, 100]"
                layout="total, sizes, prev, pager, next, jumper"
                :total="total"
                @size-change="handleSearch"
                @current-change="handleSearch"
            />
        </div>
    </div>
</template>

<script setup>
    import { ref, computed, onMounted } from 'vue'
    import { Search, User, Download, InfoFilled } from '@element-plus/icons-vue'
    import { logApi } from '../../../api/index.js'

    const logs = ref([])
    const loading = ref(false)
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(15)
    const displayLogs = computed(() => {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return logs.value.slice(start, end)
    })
    const timeRange = ref('7d')

    const searchForm = ref({
        user: '',
        module: '',
        status: ''
    })

    const fetchLogs = async () => {
        loading.value = true
        try {
            const params = {
                user: searchForm.value.user,
                module: searchForm.value.module,
                status: searchForm.value.status,
                timeRange: timeRange.value
            }
            const res = await logApi.getOperationLogs(params)
            if (res.code === 200) {
                logs.value = res.data
                total.value = res.data.length
            }
        } catch (error) {
            console.error('Error fetching logs:', error)
        } finally {
            loading.value = false
        }
    }

    const handleSearch = () => {
        currentPage.value = 1
        fetchLogs()
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
            user: '',
            module: '',
            status: ''
        }
        timeRange.value = '7d'
        fetchLogs()
    }

    const getMethodType = (method) => {
        const map = {
            'GET': 'info',
            'POST': 'success',
            'PUT': 'warning',
            'DELETE': 'danger'
        }
        return map[method] || 'info'
    }

    const getDurationClass = (duration) => {
        if (duration > 1000) return 'text-danger'
        if (duration > 500) return 'text-warning'
        return 'text-success'
    }

    onMounted(fetchLogs)
</script>

<style scoped>
    .logs-container {
        display: flex;
        flex-direction: column;
        gap: 0;
    }

    .info-alert {
        background-color: var(--el-color-primary-light-9);
        border-bottom: 1px solid var(--el-color-primary-light-8);
        padding: 8px 24px;
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .info-alert.compact {
        padding: 6px 24px;
    }

    .table-toolbar {
        padding: 16px 24px;
        background: var(--bg-primary);
        border-bottom: 1px solid var(--border);
    }

    .toolbar-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
        width: 100%;
    }

    .search-bar-form :deep(.el-form-item) {
        margin-bottom: 0;
        margin-right: 16px;
    }

    .search-bar-form :deep(.el-form-item__label) {
        font-weight: 500;
        color: var(--text-secondary);
        font-size: 13px;
    }

    .compact-time-group :deep(.el-radio-button__inner) {
        padding: 7px 12px;
        font-size: 12px;
    }

    .active-filter-tags {
        margin-top: 12px;
        padding-top: 12px;
        border-top: 1px dashed var(--border);
        display: flex;
        gap: 8px;
    }

    .mr-2 { margin-right: 8px; }

    .behavior-table {
        --el-table-header-bg-color: var(--bg-primary);
    }

    :deep(.el-table) {
        border-radius: 0;
    }

    :deep(.el-table__inner-wrapper::before) {
        display: none;
    }

    .pagination-container {
        padding: 16px 24px;
        display: flex;
        justify-content: flex-end;
        background: var(--bg-primary);
        border-top: 1px solid var(--border);
    }

    .user-cell {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .code-text {
        font-family: var(--mono);
        background: var(--bg-tertiary);
        padding: 2px 6px;
        border-radius: 2px;
        font-size: 12px;
        color: var(--primary);
    }

    .json-text {
        font-family: var(--mono);
        font-size: 11px;
        color: var(--text-secondary);
        white-space: nowrap;
    }

    .text-success { color: #10b981; }
    .text-warning { color: #f59e0b; }
    .text-danger { color: #ef4444; }

    :deep(.el-button),
    :deep(.el-input__wrapper),
    :deep(.el-select .el-input__wrapper),
    :deep(.el-card),
    :deep(.el-tag),
    :deep(.el-radio-button:first-child .el-radio-button__inner),
    :deep(.el-radio-button:last-child .el-radio-button__inner) {
        border-radius: 2px !important;
    }

    :deep(.table-header-cell) {
        background-color: var(--bg-secondary) !important;
    }
</style>
