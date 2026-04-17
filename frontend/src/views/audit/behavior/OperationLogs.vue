<template>
    <div class="logs-container">
        <header class="page-header">
            <div class="header-content">
                <div style="display: flex; align-items: center; gap: 12px">
                    <h2 class="page-title">行为日志</h2>
                    <span class="accent-dot"></span>
                </div>
                <p class="page-subtitle">审计系统内所有用户的关键操作行为及接口调用详情。</p>
            </div>
            <div class="header-actions">
                <el-button :icon="Download">导出日志</el-button>
            </div>
        </header>

        <!-- Info Alert -->
        <div class="info-alert">
            <el-icon class="info-icon">
                <InfoFilled />
            </el-icon>
            <span class="info-text">默认免费记录了90天的管控事件可供查询，建议 <el-link type="primary" :underline="false">创建跟踪</el-link> 实现更长时间存储，并通过 <el-link type="primary" :underline="false">事件高级查询</el-link> 实现更灵活的多条件组合查询</span>
        </div>

        <!-- Search & Filter -->
        <div class="compact-filter">
            <el-form :model="searchForm" class="search-form">
                <div class="filter-row">
                    <div class="filter-group">
                        <div class="filter-label">操作人</div>
                        <el-input v-model="searchForm.user" placeholder="请选择" clearable class="filter-input" @clear="handleSearch" @input="handleInput" />
                    </div>
                    <div class="filter-group">
                        <div class="filter-label">所属模块</div>
                        <el-select v-model="searchForm.module" placeholder="请选择" clearable class="filter-select" @change="handleSearch">
                            <el-option label="主机管理" value="主机管理" />
                            <el-option label="自动化任务" value="自动化任务" />
                            <el-option label="监控中心" value="监控中心" />
                            <el-option label="系统设置" value="系统设置" />
                        </el-select>
                    </div>
                    <div class="filter-group">
                        <div class="filter-label">状态</div>
                        <el-select v-model="searchForm.status" placeholder="请选择" clearable class="filter-select" @change="handleSearch">
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
            <el-table :data="displayLogs" style="width: 100%" v-loading="loading" header-cell-class-name="table-header-cell">
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
                <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 15, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" :total="total" @size-change="handleSearch" @current-change="handleSearch" />
            </div>
        </el-card>
    </div>
</template>

<script setup>
    import { ref, computed, onMounted } from 'vue'
    import { Search, User, Download, InfoFilled } from '@element-plus/icons-vue'
    import { auditApi } from '../../../api/index.js'

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
            const res = await auditApi.getOperationLogs(params)
            if (res.code === 200) {
                logs.value = res.data.list
                total.value = res.data.total
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

    .filter-input,
    .filter-select {
      width: 200px;
    }

    :deep(.el-input__wrapper),
    :deep(.el-select .el-input__wrapper) {
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
