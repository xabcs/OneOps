<template>
    <div class="logs-container">
        <header class="page-header">
            <div class="header-content">
                <h2 class="page-title">操作日志</h2>
                <p class="page-subtitle">审计系统内所有用户的关键操作行为。</p>
            </div>
            <div class="header-actions">
                <el-button :icon="Download">导出日志</el-button>
            </div>
        </header>

        <!-- Search & Filter -->
        <el-card shadow="never" class="filter-card">
            <el-form :model="searchForm" inline class="search-form">
                <el-form-item label="操作人">
                    <el-input 
                        v-model="searchForm.user" 
                        placeholder="用户名" 
                        clearable 
                        :prefix-icon="User"
                        @clear="handleSearch"
                        @input="handleInput"
                    />
                </el-form-item>
                <el-form-item label="所属模块">
                    <el-select 
                        v-model="searchForm.module" 
                        placeholder="全部模块" 
                        clearable 
                        style="width: 140px"
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
                        placeholder="全部状态" 
                        clearable 
                        style="width: 120px"
                        @change="handleSearch"
                    >
                        <el-option label="成功" value="success" />
                        <el-option label="失败" value="failed" />
                    </el-select>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="handleSearch">查询</el-button>
                    <el-button @click="resetForm">重置</el-button>
                </el-form-item>
            </el-form>
        </el-card>

        <!-- Log Table -->
        <el-card shadow="never" class="table-card">
            <el-table 
                :data="logs" 
                style="width: 100%" 
                v-loading="loading"
                header-cell-class-name="table-header-cell"
            >
                <el-table-column prop="time" label="操作时间" width="180">
                    <template #default="{ row }">
                        <span class="data-value">{{ row.time }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="user" label="操作人" width="120">
                    <template #default="{ row }">
                        <div class="user-cell">
                            <el-avatar :size="24" :src="'https://api.dicebear.com/7.x/avataaars/svg?seed=' + row.user" />
                            <span>{{ row.user }}</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column prop="action" label="操作行为" min-width="150" />
                <el-table-column prop="module" label="所属模块" width="120">
                    <template #default="{ row }">
                        <el-tag size="small" effect="plain">{{ row.module }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="ip" label="IP 地址" width="140" />
                <el-table-column prop="status" label="状态" width="100">
                    <template #default="{ row }">
                        <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                            {{ row.status === 'success' ? '成功' : '失败' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="操作" width="100" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary">详情</el-button>
                    </template>
                </el-table-column>
            </el-table>

            <div class="pagination-container">
                <el-pagination
                    v-model:current-page="currentPage"
                    v-model:page-size="pageSize"
                    layout="total, prev, pager, next"
                    :total="total"
                />
            </div>
        </el-card>
    </div>
</template>

<script setup>
    import { ref, onMounted } from 'vue'
    import { Search, User, Download } from '@element-plus/icons-vue'
    import { logApi } from '../api/index.js'

    const logs = ref([])
    const loading = ref(false)
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(15)

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
                status: searchForm.value.status
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
        fetchLogs()
    }

    onMounted(fetchLogs)
</script>

<style scoped>
    .logs-container {
        display: flex;
        flex-direction: column;
        gap: 20px;
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
    }

    :deep(.table-header-cell) {
        background-color: var(--bg-secondary) !important;
    }
</style>