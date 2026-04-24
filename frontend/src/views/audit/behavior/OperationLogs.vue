<template>
    <div class="logs-container">
        <!-- Search & Filter -->
        <div class="compact-filter">
            <div class="filter-row">
                <el-form :model="searchForm" class="search-form-inline" inline>
                    <el-form-item label="操作人">
                        <el-input v-model="searchForm.username" placeholder="输入用户名" clearable style="width: 150px" @clear="handleSearch" @input="handleInput" />
                    </el-form-item>
                    <el-form-item label="操作模块">
                        <el-select v-model="searchForm.module" placeholder="全部" clearable style="width: 140px" @change="handleSearch">
                            <el-option
                                v-for="module in availableModules"
                                :key="module"
                                :label="module"
                                :value="module"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="请求方法">
                        <el-select v-model="searchForm.method" placeholder="全部" clearable style="width: 120px" @change="handleSearch">
                            <el-option label="GET" value="GET" />
                            <el-option label="POST" value="POST" />
                            <el-option label="PUT" value="PUT" />
                            <el-option label="DELETE" value="DELETE" />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="状态">
                        <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px" @change="handleSearch">
                            <el-option label="成功" value="success" />
                            <el-option label="失败" value="failed" />
                        </el-select>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="handleSearch">查询</el-button>
                        <el-button @click="resetForm">重置</el-button>
                    </el-form-item>
                </el-form>
            </div>

            <div class="filter-row time-range-row">
                <span class="time-label">时间范围：</span>
                <el-radio-group v-model="timeRange" size="small" @change="handleSearch">
                    <el-radio-button label="1h">1小时</el-radio-button>
                    <el-radio-button label="1d">今天</el-radio-button>
                    <el-radio-button label="7d">7天</el-radio-button>
                    <el-radio-button label="30d">30天</el-radio-button>
                    <el-radio-button label="custom">自定义</el-radio-button>
                </el-radio-group>
                <el-date-picker
                    v-if="timeRange === 'custom'"
                    v-model="customTimeRange"
                    type="datetimerange"
                    range-separator="至"
                    start-placeholder="开始时间"
                    end-placeholder="结束时间"
                    size="small"
                    style="width: 350px; margin-left: 8px"
                    @change="handleSearch"
                />
            </div>
        </div>

        <!-- Log Table -->
        <el-card shadow="never" class="table-card">
            <el-table :data="displayLogs" style="width: 100%" v-loading="loading" header-cell-class-name="table-header-cell" @sort-change="handleSortChange">
                <el-table-column prop="time" label="操作时间" width="180" sortable="custom">
                    <template #default="{ row }">
                        <span class="data-value nowrap-text">{{ row.time }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="user" label="操作人" width="110">
                    <template #default="{ row }">
                        <div class="user-cell">
                            <el-avatar :size="24" :src="'https://api.dicebear.com/7.x/avataaars/svg?seed=' + (row.username || row.user)" />
                            <span>{{ row.username || row.user }}</span>
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
                <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 15, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" :total="total" @size-change="handleSizeChange" @current-change="handlePageChange" />
            </div>
        </el-card>
    </div>
</template>

<script setup>
    import { ref, computed, onMounted, watch } from 'vue'
    import { Search, User, Download, InfoFilled } from '@element-plus/icons-vue'
    import { auditApi } from '../../../api/index.js'

    const logs = ref([])
    const loading = ref(false)
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(15)
    const sortBy = ref('operate_time') // 排序字段（对应数据库字段）
    const sortOrder = ref('DESC') // 排序方向（ASC/DESC）

    // 计算显示的日志（直接显示后端返回的数据，后端已处理分页和排序）
    const displayLogs = computed(() => {
        return logs.value
    })
    const timeRange = ref('7d')
    const availableModules = ref([]) // 动态获取的模块列表

    const searchForm = ref({
        username: '',
        module: '',
        method: '',
        status: ''
    })
    const customTimeRange = ref([])

    // 格式化本地时间为 YYYY-MM-DD HH:mm:ss
    const formatLocalDateTime = (date) => {
        const year = date.getFullYear()
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const day = String(date.getDate()).padStart(2, '0')
        const hours = String(date.getHours()).padStart(2, '0')
        const minutes = String(date.getMinutes()).padStart(2, '0')
        const seconds = String(date.getSeconds()).padStart(2, '0')
        return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
    }

    // 获取可用的模块列表
    const fetchModules = async () => {
        try {
            const res = await auditApi.getModules()
            if (res.code === 200) {
                availableModules.value = res.data || []
            }
        } catch (error) {
            console.error('Error fetching modules:', error)
        }
    }

    const fetchLogs = async () => {
        loading.value = true

        try {
            const params = {
                username: searchForm.value.username,
                module: searchForm.value.module,
                method: searchForm.value.method,
                status: searchForm.value.status,
                page: currentPage.value,
                pageSize: pageSize.value,
                sortBy: sortBy.value,
                sortOrder: sortOrder.value
            }

            // 添加时间范围参数
            const now = new Date()
            let startTime = new Date()

            switch (timeRange.value) {
                case '1h':
                    startTime = new Date(now.getTime() - 60 * 60 * 1000)
                    break
                case '1d':
                    startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000)
                    break
                case '7d':
                    startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
                    break
                case '30d':
                    startTime = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
                    break
                case 'custom':
                    if (customTimeRange.value && customTimeRange.value.length === 2) {
                        params.startTime = formatLocalDateTime(new Date(customTimeRange.value[0]))
                        params.endTime = formatLocalDateTime(new Date(customTimeRange.value[1]))
                    }
                    break
                default:
                    startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000) // 默认7天
            }

            if (timeRange.value !== 'custom') {
                params.startTime = formatLocalDateTime(startTime)
                params.endTime = formatLocalDateTime(now)
            }

            const res = await auditApi.getOperationLogs(params)
            if (res.code === 200) {
                logs.value = res.data.list || []
                total.value = res.data.total || 0
            }
        } catch (error) {
            console.error('Error fetching logs:', error)
        } finally {
            loading.value = false
        }
    }

    const handleSearch = () => {
        currentPage.value = 1  // 重置到第一页
        fetchLogs()
    }

    const handleSizeChange = (val) => {
        pageSize.value = val
        currentPage.value = 1  // 改变每页大小时，重置到第一页
        fetchLogs()
    }

    const handlePageChange = (val) => {
        currentPage.value = val
        fetchLogs()  // 只改变页码，不重置
    }

    let inputTimer = null
    const handleInput = () => {
        if (inputTimer) clearTimeout(inputTimer)
        inputTimer = setTimeout(() => {
            currentPage.value = 1  // 重置到第一页
            handleSearch()
        }, 500)
    }

    const resetForm = () => {
        searchForm.value = {
            username: '',
            module: '',
            method: '',
            status: ''
        }
        timeRange.value = '7d'
        customTimeRange.value = []
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

    // 处理排序变化
    const handleSortChange = ({ prop, order }) => {
        if (prop) {
            // 将前端的排序字段映射到数据库字段
            const fieldMap = {
                'time': 'operate_time',
                'user': 'username',
                'action': 'action',
                'duration': 'duration',
                'status': 'status'
            }
            sortBy.value = fieldMap[prop] || 'operate_time'
            sortOrder.value = order === 'ascending' ? 'ASC' : 'DESC'

            // 重新获取数据
            fetchLogs()
        }
    }

    onMounted(() => {
        fetchModules() // 先获取模块列表
        fetchLogs()      // 再获取日志数据
    })

    // 使用watch监听timeRange变化
    watch(() => timeRange.value, () => {
        currentPage.value = 1
        fetchLogs()
    })
</script>

<style scoped>
    .logs-container {
      display: flex;
      flex-direction: column;
      height: calc(100vh - 120px);
      background-color: var(--bg-secondary);
    }

    .compact-filter {
      padding: 12px 16px;
      border-bottom: 1px solid var(--border);
      background-color: var(--bg-primary);
      border-radius: 0;
      margin-bottom: 0;
    }

    .table-card {
      border: none;
      border-radius: 0;
      display: flex;
      flex-direction: column;
      flex: 1;
      background-color: transparent;
      box-shadow: none;
    }

    :deep(.el-card__body) {
      padding: 0;
      display: flex;
      flex-direction: column;
      flex: 1;
      background-color: transparent;
    }

    :deep(.el-table) {
      flex: 1;
      border: none;
      background-color: transparent;
    }

    :deep(.el-table__header-wrapper) {
      background-color: var(--bg-primary);
    }

    :deep(.table-header-cell) {
      background-color: var(--bg-primary) !important;
      font-weight: 500;
      color: var(--text-primary);
      border-bottom: 1px solid var(--border);
    }

    :deep(.table-header-cell .caret-wrapper) {
      cursor: pointer;
      margin-left: 4px;
    }

    :deep(.table-header-cell .sort-caret.ascending) {
      border-bottom-color: var(--primary);
    }

    :deep(.table-header-cell .sort-caret.descending) {
      border-top-color: var(--primary);
    }

    :deep(.el-table__body tr) {
      background-color: transparent;
    }

    :deep(.el-table__body tr:hover > td) {
      background-color: var(--bg-tertiary) !important;
    }

    :deep(.el-table td) {
      border: none;
      padding: 12px 0;
    }

    :deep(.el-table__empty-block) {
      background-color: transparent;
    }

    .filter-row {
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 12px;
    }

    .filter-row:last-child {
        margin-bottom: 0;
    }

    .search-form-inline {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        align-items: center;
    }

    :deep(.el-form-item) {
        margin-bottom: 0;
    }

    .time-range-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .time-label {
        font-size: 14px;
        color: #606266;
        white-space: nowrap;
    }

    .user-cell {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .pagination-container {
        margin-top: 24px;
        display: flex;
        justify-content: flex-end;
        padding: 0 16px 16px;
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

    .text-success {
      color: #10b981;
    }
    .text-warning {
      color: #f59e0b;
    }
    .text-danger {
      color: #ef4444;
    }

    .data-value {
      display: inline-block;
    }

    .nowrap-text {
      white-space: nowrap;
    }

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
