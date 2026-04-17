<template>
    <div class="logs-container">
        <!-- Search & Filter -->
        <div class="compact-filter">
            <div class="filter-row">
                <el-form :model="searchForm" class="search-form-inline" inline>
                    <el-form-item label="登录账号">
                        <el-input v-model="searchForm.username" placeholder="输入账号" clearable style="width: 180px" @clear="handleSearch" @input="handleInput" />
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
            <el-table :data="displayLogs" style="width: 100%" v-loading="loading" header-cell-class-name="table-header-cell">
                <el-table-column prop="time" label="登录时间" width="180">
                    <template #default="{ row }">
                        <span class="data-value">{{ row.time }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="username" label="登录账号" width="120">
                    <template #default="{ row }">
                        <div class="user-cell">
                            <el-avatar :size="24" :src="'https://api.dicebear.com/7.x/avataaars/svg?seed=' + row.username" />
                            <span>{{ row.username }}</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="ip" label="登录 IP" width="140" />
                <el-table-column prop="location" label="登录地点" width="150" />
                <el-table-column prop="browser" label="浏览器" width="150" />
                <el-table-column prop="os" label="操作系统" width="120" />
                <el-table-column prop="status" label="状态" width="100">
                    <template #default="{ row }">
                        <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                            {{ row.status === 'success' ? '成功' : '失败' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="msg" label="提示信息" min-width="150" />
            </el-table>

            <div class="pagination-container">
                <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" :total="total" @size-change="handleSizeChange" @current-change="handleCurrentChange" />
            </div>
        </el-card>
    </div>
</template>

<script setup>
    import { ref, computed, onMounted } from 'vue'
    import { User, Download, Search, InfoFilled } from '@element-plus/icons-vue'
    import { auditApi } from '../../../api/index.js'

    const loading = ref(false)
    const logs = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(10)
    const displayLogs = computed(() => {
        const start = (currentPage.value - 1) * pageSize.value
        const end = start + pageSize.value
        return logs.value.slice(start, end)
    })
    const timeRange = ref('7d')

    const searchForm = ref({
        username: '',
        status: ''
    })
    const customTimeRange = ref([])

    const fetchLogs = async () => {
        loading.value = true
        try {
            const params = {
                username: searchForm.value.username,
                status: searchForm.value.status,
                page: currentPage.value,
                pageSize: pageSize.value
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
                        params.startTime = customTimeRange.value[0].toISOString().slice(0, 19).replace('T', ' ')
                        params.endTime = customTimeRange.value[1].toISOString().slice(0, 19).replace('T', ' ')
                    }
                    break
                default:
                    startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000) // 默认7天
            }

            if (timeRange.value !== 'custom') {
                params.startTime = startTime.toISOString().slice(0, 19).replace('T', ' ')
                params.endTime = now.toISOString().slice(0, 19).replace('T', ' ')
            }

            const res = await auditApi.getLoginLogs(params)
            if (res.code === 200) {
                logs.value = res.data.list || []
                total.value = res.data.total || 0
            }
        } catch (error) {
            console.error('Error fetching login logs:', error)
        } finally {
            loading.value = false
        }
    }

    const handleSearch = () => {
        currentPage.value = 1
        fetchLogs()
    }

    const handleInput = () => {
        // 可以添加防抖
        handleSearch()
    }

    const resetForm = () => {
        searchForm.value = {
            username: '',
            status: ''
        }
        timeRange.value = '7d'
        customTimeRange.value = []
        handleSearch()
    }

    const handleSizeChange = (val) => {
        pageSize.value = val
        fetchLogs()
    }

    const handleCurrentChange = (val) => {
        currentPage.value = val
        fetchLogs()
    }

    onMounted(() => {
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

    :deep(.el-button),
    :deep(.el-input__wrapper),
    :deep(.el-select .el-input__wrapper),
    :deep(.el-card),
    :deep(.el-tag),
    :deep(.el-radio-button:first-child .el-radio-button__inner),
    :deep(.el-radio-button:last-child .el-radio-button__inner) {
      border-radius: 2px !important;
    }
</style>
