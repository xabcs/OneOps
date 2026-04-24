<template>
  <div class="logs-container">
    <!-- 使用新的 AuditFilter 组件 -->
    <AuditFilter
      :form="searchForm"
      @search="handleSearch"
      @reset="resetForm"
      @time-range-change="handleTimeRangeChange"
    >
      <template #filters>
        <el-form-item label="操作人">
          <el-input
            v-model="searchForm.username"
            placeholder="输入用户名"
            clearable
            style="width: 150px"
          />
        </el-form-item>
        <el-form-item label="操作模块">
          <el-select
            v-model="searchForm.module"
            placeholder="全部"
            clearable
            style="width: 140px"
          >
            <el-option
              v-for="module in availableModules"
              :key="module"
              :label="module"
              :value="module"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="请求方法">
          <el-select
            v-model="searchForm.method"
            placeholder="全部"
            clearable
            style="width: 120px"
          >
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部"
            clearable
            style="width: 120px"
          >
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
          </el-select>
        </el-form-item>
      </template>
    </AuditFilter>

    <!-- 使用新的 AuditLogTable 组件 -->
    <AuditLogTable
      :data="logs"
      :loading="loading"
      :total="total"
      :current-page="currentPage"
      :page-size="pageSize"
      :page-sizes="[10, 15, 20, 50, 100]"
      @update:current-page="handlePageChange"
      @update:page-size="handleSizeChange"
      @sort-change="handleSortChange"
    >
      <!-- 操作时间 -->
      <el-table-column prop="operateTime" label="操作时间" width="180" sortable="custom">
        <template #default="{ row }">
          <span class="data-value">{{ row.operate_time || row.time }}</span>
        </template>
      </el-table-column>

      <!-- 操作人 -->
      <el-table-column prop="username" label="操作人" width="110">
        <template #default="{ row }">
          <div class="user-cell">
            <el-avatar :size="24" :src="getAvatarUrl(row.username || row.user)" />
            <span>{{ row.username || row.user }}</span>
          </div>
        </template>
      </el-table-column>

      <!-- 操作行为 -->
      <el-table-column prop="action" label="操作行为" width="120">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.action }}</el-tag>
        </template>
      </el-table-column>

      <!-- 请求路径 -->
      <el-table-column prop="path" label="请求路径" min-width="180">
        <template #default="{ row }">
          <code class="code-text">{{ row.path }}</code>
        </template>
      </el-table-column>

      <!-- 请求方法 -->
      <el-table-column prop="method" label="方法" width="80" align="center">
        <template #default="{ row }">
          <el-tag
            size="small"
            :type="getMethodType(row.method)"
            effect="dark"
          >
            {{ row.method }}
          </el-tag>
        </template>
      </el-table-column>

      <!-- 请求参数 -->
      <el-table-column prop="params" label="请求参数" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="json-text">{{ formatJSON(row.params) }}</span>
        </template>
      </el-table-column>

      <!-- 响应数据 -->
      <el-table-column prop="response" label="响应数据" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="json-text">{{ formatJSON(row.response) }}</span>
        </template>
      </el-table-column>

      <!-- 耗时 -->
      <el-table-column prop="duration" label="耗时(ms)" width="100" align="right">
        <template #default="{ row }">
          <span :class="getDurationClass(row.duration)">{{ row.duration }}ms</span>
        </template>
      </el-table-column>

      <!-- 状态 -->
      <el-table-column prop="status" label="状态" width="80" align="center">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>
    </AuditLogTable>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { AuditFilter, AuditLogTable, StatusTag } from '@/components'
import { auditApi } from '@/api'

const logs = ref([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(15)
const timeRange = ref('7d')
const availableModules = ref([])

const searchForm = ref({
  username: '',
  module: '',
  method: '',
  status: ''
})

// 获取头像URL
const getAvatarUrl = (username) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${username}`
}

// 获取方法类型
const getMethodType = (method) => {
  const types = {
    'GET': 'success',
    'POST': 'warning',
    'PUT': 'primary',
    'DELETE': 'danger'
  }
  return types[method] || ''
}

// 获取耗时样式类
const getDurationClass = (duration) => {
  if (duration < 100) return 'fast'
  if (duration < 500) return 'normal'
  return 'slow'
}

// 格式化JSON
const formatJSON = (data) => {
  if (!data) return '-'
  if (typeof data === 'string') {
    try {
      const parsed = JSON.parse(data)
      return JSON.stringify(parsed, null, 2)
    } catch {
      return data
    }
  }
  return JSON.stringify(data, null, 2)
}

// 获取模块列表
const fetchModules = async () => {
  try {
    const res = await auditApi.getModules()
    if (res.code === 200) {
      availableModules.value = res.data || []
    }
  } catch (error) {
    console.error('获取模块列表失败:', error)
  }
}

// 获取日志数据
const fetchLogs = async () => {
  loading.value = true
  try {
    const params = {
      username: searchForm.value.username,
      module: searchForm.value.module,
      method: searchForm.value.method,
      status: searchForm.value.status,
      page: currentPage.value,
      pageSize: pageSize.value
    }

    // 添加时间范围参数
    const timeParams = getTimeRangeParams()
    Object.assign(params, timeParams)

    console.log('操作日志查询参数:', params)

    const res = await auditApi.getOperationLogs(params)
    if (res.code === 200) {
      logs.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    console.error('获取操作日志失败:', error)
  } finally {
    loading.value = false
  }
}

// 获取时间范围参数
const getTimeRangeParams = () => {
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
      return {}
    default:
      startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
  }

  return {
    startTime: formatDateTime(startTime),
    endTime: formatDateTime(now)
  }
}

// 格式化日期时间
const formatDateTime = (date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// 事件处理
const handleSearch = () => {
  currentPage.value = 1
  fetchLogs()
}

const resetForm = () => {
  searchForm.value = {
    username: '',
    module: '',
    method: '',
    status: ''
  }
  timeRange.value = '7d'
  currentPage.value = 1
  fetchLogs()
}

const handleTimeRangeChange = () => {
  currentPage.value = 1
  fetchLogs()
}

const handlePageChange = (val) => {
  currentPage.value = val
  fetchLogs()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  currentPage.value = 1
  fetchLogs()
}

const handleSortChange = ({ prop, order }) => {
  console.log('排序变化:', prop, order)
}

// 初始化
onMounted(() => {
  fetchModules()
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

.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.code-text {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-secondary, #64748b);
  background: var(--bg-tertiary, #f1f5f9);
  padding: 2px 6px;
  border-radius: 4px;
}

.json-text {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-secondary, #64748b);
}

.fast {
  color: #10b981;
}

.normal {
  color: #f59e0b;
}

.slow {
  color: #ef4444;
}

.data-value {
  color: var(--text-primary, #1e293b);
}
</style>
