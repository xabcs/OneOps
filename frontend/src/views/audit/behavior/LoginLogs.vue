<template>
    <div class="logs-container">
        <!-- Info Alert -->
        <div class="info-alert compact">
            <el-icon class="info-icon"><InfoFilled /></el-icon>
            <span class="info-text">系统默认记录最近90天的登录审计事件。如需长期保存，请前往 <el-link type="primary" :underline="false">审计配置</el-link></span>
        </div>

        <!-- Toolbar -->
        <div class="table-toolbar">
            <el-form :model="searchForm" inline class="search-bar-form">
                <el-form-item label="登录账号">
                    <el-input 
                        v-model="searchForm.username" 
                        placeholder="用户名" 
                        clearable 
                        :prefix-icon="User"
                        style="width: 140px"
                        @clear="handleSearch"
                        @input="handleInput"
                    />
                </el-form-item>
                <el-form-item label="地点">
                    <el-input 
                        v-model="searchForm.location" 
                        placeholder="城市" 
                        clearable 
                        style="width: 120px"
                        @clear="handleSearch"
                        @input="handleInput"
                    />
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
            
            <div v-if="searchForm.username" class="active-filter-tags">
                <el-tag closable @close="searchForm.username=''; handleSearch()" size="small">账号: {{ searchForm.username }}</el-tag>
            </div>
        </div>

        <!-- Log Table -->
        <el-table 
            :data="displayLogs" 
            style="width: 100%" 
            v-loading="loading"
            class="behavior-table"
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
