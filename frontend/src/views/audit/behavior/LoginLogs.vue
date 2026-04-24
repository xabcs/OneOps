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
        <el-form-item label="登录账号">
          <el-input
            v-model="searchForm.username"
            placeholder="输入账号"
            clearable
            style="width: 180px"
            @clear="handleSearch"
          />
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
      :data="displayLogs"
      :loading="loading"
      :total="total"
      :current-page="currentPage"
      :page-size="pageSize"
      :page-sizes="[10, 20, 50, 100]"
      @update:current-page="handleCurrentChange"
      @update:page-size="handleSizeChange"
      @sort-change="handleSortChange"
    >
      <!-- 登录时间 -->
      <el-table-column prop="time" label="登录时间" width="180" sortable="custom">
        <template #default="{ row }">
          <span class="data-value">{{ row.time }}</span>
        </template>
      </el-table-column>

      <!-- 登录账号 -->
      <el-table-column prop="username" label="登录账号" width="120">
        <template #default="{ row }">
          <div class="user-cell">
            <el-avatar :size="24" :src="getAvatarUrl(row.username)" />
            <span>{{ row.username }}</span>
          </div>
        </template>
      </el-table-column>

      <!-- 登录 IP -->
      <el-table-column prop="ip" label="登录 IP" width="140" />

      <!-- 登录地点 -->
      <el-table-column prop="location" label="登录地点" width="150" />

      <!-- 浏览器 -->
      <el-table-column prop="browser" label="浏览器" width="150" />

      <!-- 操作系统 -->
      <el-table-column prop="os" label="操作系统" width="120" />

      <!-- 状态 -->
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <StatusTag :status="row.status" />
        </template>
      </el-table-column>

      <!-- 提示信息 -->
      <el-table-column prop="msg" label="提示信息" min-width="150" />
    </AuditLogTable>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { AuditFilter, AuditLogTable, StatusTag } from '@/components'
import { auditApi } from '@/api'

// 状态
const logs = ref([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const timeRange = ref('7d')

// 搜索表单
const searchForm = ref({
  username: '',
  status: ''
})

// 计算显示的日志
const displayLogs = computed(() => {
  // 如果后端处理了分页和排序，直接返回
  // 如果需要前端处理，可以在这里添加逻辑
  return logs.value
})

// 获取头像URL
const getAvatarUrl = (username) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${username}`
}

// 获取日志数据
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
    const timeParams = getTimeRangeParams()
    Object.assign(params, timeParams)

    const res = await auditApi.getLoginLogs(params)
    if (res.code === 200) {
      logs.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    console.error('获取登录日志失败:', error)
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
      // 自定义时间在 handleTimeRangeChange 中处理
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
    status: ''
  }
  timeRange.value = '7d'
  currentPage.value = 1
  fetchLogs()
}

const handleTimeRangeChange = (data) => {
  timeRange.value = data.range
  currentPage.value = 1
  fetchLogs()
}

const handleCurrentChange = (val) => {
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
  // 实现排序逻辑
}

// 初始化
fetchLogs()
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

.data-value {
  color: var(--text-primary, #1e293b);
}

.nowrap-text {
  white-space: nowrap;
}
</style>
