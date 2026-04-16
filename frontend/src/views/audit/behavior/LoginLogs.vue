<template>
    <div class="logs-container">
        <!-- Info Alert -->
        <div class="info-alert">
            <el-icon class="info-icon"><InfoFilled /></el-icon>
            <span class="info-text">系统默认记录最近90天的登录审计事件。如需长期保存或进行多维分析，请前往 <el-link type="primary" :underline="false">审计配置</el-link> 开启日志投递。</span>
        </div>

        <!-- Search & Filter -->
        <div class="compact-filter">
            <el-form :model="searchForm" class="search-form">
                <div class="filter-row">
                    <div class="filter-group">
                        <div class="filter-label">登录账号</div>
                        <el-input 
                            v-model="searchForm.username" 
                            placeholder="请选择" 
                            clearable 
                            class="filter-input"
                            @clear="handleSearch"
                            @input="handleInput"
                        />
                    </div>
                    <div class="filter-group">
                        <div class="filter-label">登录地点</div>
                        <el-input 
                            v-model="searchForm.location" 
                            placeholder="请选择" 
                            clearable 
                            class="filter-input"
                            @clear="handleSearch"
                            @input="handleInput"
                        />
                    </div>
                    <div class="filter-group">
                        <div class="filter-label">状态</div>
                        <el-select 
                            v-model="searchForm.status" 
                            placeholder="请选择" 
                            clearable 
                            class="filter-select"
                            @change="handleSearch"
                        >
                            <el-option label="成功" value="success" />
                            <el-option label="失败" value="failed" />
                        </el-select>
                    </div>
                    <div class="filter-actions">
                        <el-button type="primary" class="submit-btn" @click="handleSearch">提交查询</el-button>
                        <el-button @click="resetForm">重置</el-button>
                    </div>
                </div>
                
                <div class="time-range-row">
                    <el-radio-group v-model="timeRange" size="small" @change="handleSearch">
                        <el-radio-button label="1h">1h</el-radio-button>
                        <el-radio-button label="12h">12h</el-radio-button>
                        <el-radio-button label="1d">1d</el-radio-button>
                        <el-radio-button label="7d">7d</el-radio-button>
                        <el-radio-button label="30d">30d</el-radio-button>
                        <el-radio-button label="custom">自定义</el-radio-button>
                    </el-radio-group>
                </div>
            </el-form>
        </div>

        <!-- Log Table -->
        <el-card shadow="never" class="table-card">
            <el-table 
                :data="displayLogs" 
                style="width: 100%" 
                v-loading="loading"
                header-cell-class-name="table-header-cell"
            >
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
    import { User, Download, Search, InfoFilled } from '@element-plus/icons-vue'
    import { logApi } from '../../../api/index.js'

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
        location: '',
        status: ''
    })

    const fetchLogs = async () => {
        loading.value = true
        try {
            const params = {
                username: searchForm.value.username,
                location: searchForm.value.location,
                status: searchForm.value.status,
                timeRange: timeRange.value
            }
            const res = await logApi.getLoginLogs(params)
            if (res.code === 200) {
                logs.value = res.data
                total.value = res.data.length
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
            location: '',
            status: ''
        }
        timeRange.value = '7d'
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
        gap: 16px;
    }

    .info-alert {
        background-color: #f0f7ff;
        border: 1px solid #d1e9ff;
        border-radius: 2px;
        padding: 12px 16px;
        display: flex;
        align-items: center;
        gap: 12px;
        margin-bottom: 8px;
    }

    .info-icon {
        color: #0070cc;
        font-size: 18px;
    }

    .info-text {
        font-size: 14px;
        color: #333;
        line-height: 1.5;
    }

    .compact-filter {
        padding: 0 0 16px 0;
    }

    .search-form {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .filter-row {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        align-items: center;
    }

    .filter-group {
        display: flex;
        align-items: center;
        border: 1px solid #dcdfe6;
        border-radius: 2px;
        overflow: hidden;
        transition: border-color 0.2s;
        height: 32px;
    }

    .filter-group:focus-within {
        border-color: var(--primary);
    }

    .filter-label {
        background-color: #f5f7fa;
        padding: 0 12px;
        height: 32px;
        line-height: 32px;
        font-size: 13px;
        color: #606266;
        border-right: 1px solid #dcdfe6;
        white-space: nowrap;
    }

    .filter-input, .filter-select {
        width: 200px;
    }

    :deep(.el-input__wrapper), :deep(.el-select .el-input__wrapper) {
        box-shadow: none !important;
        background-color: transparent !important;
        height: 30px;
        padding: 0 8px;
    }

    .filter-actions {
        display: flex;
        gap: 8px;
        margin-left: auto;
    }

    .time-range-row {
        display: flex;
        align-items: center;
    }

    :deep(.el-radio-button__inner) {
        padding: 6px 12px;
        font-size: 12px;
    }

    .table-card {
        border: 1px solid var(--border);
        flex: 1;
        display: flex;
        flex-direction: column;
        border-radius: 2px;
    }

    .user-cell {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .pagination-container {
        margin-top: 20px;
        display: flex;
        justify-content: flex-end;
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

    :deep(.el-card__body) {
        padding: 12px;
    }
</style>
