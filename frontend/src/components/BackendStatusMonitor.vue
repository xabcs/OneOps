<template>
  <div v-if="!isBackendAvailable && showNotification && !isOnLoginPage" class="backend-status-notification">
    <div class="status-content">
      <el-icon class="status-icon warning-icon"><WarningFilled /></el-icon>
      <div class="status-text">
        <h4>后端服务不可用</h4>
        <p>无法连接到服务器，部分功能可能受限。请检查后端服务是否已启动。</p>
      </div>
      <el-button
        type="primary"
        size="small"
        @click="handleRetry"
        :loading="isRetrying"
        class="retry-btn"
      >
        重试连接
      </el-button>
      <el-button
        size="small"
        @click="dismissNotification"
        class="dismiss-btn"
      >
        关闭
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useStore } from 'vuex'
import { WarningFilled } from '@element-plus/icons-vue'
import { loginApi } from '../api/index.js'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const store = useStore()
const isRetrying = ref(false)
const showNotification = ref(true)

const isBackendAvailable = computed(() => store.getters.isBackendAvailable)
const isOnLoginPage = computed(() => route.path === '/login')

// 监听后端状态变化
watch(isBackendAvailable, (newValue) => {
  if (newValue) {
    showNotification.value = false
    ElMessage.success('后端服务已恢复连接')
  } else {
    showNotification.value = true
  }
})

// 重试连接
const handleRetry = async () => {
  isRetrying.value = true
  try {
    await loginApi.getUserInfo()
    // 如果成功，说明后端已恢复
    store.commit('SET_BACKEND_AVAILABLE', true)
    ElMessage.success('已成功连接到后端服务')
  } catch (error) {
    console.error('重试连接失败:', error)
    ElMessage.error('重试连接失败，请检查后端服务状态')
  } finally {
    isRetrying.value = false
  }
}

// 关闭通知
const dismissNotification = () => {
  showNotification.value = false
}
</script>

<style scoped>
.backend-status-notification {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  max-width: 420px;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.status-content {
  background: #fffbeb;
  border: 2px solid #fbbf24;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.status-icon {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  margin-top: 2px;
}

.warning-icon {
  color: #f59e0b;
}

.status-text {
  flex: 1;
  min-width: 0;
}

.status-text h4 {
  margin: 0 0 4px 0;
  font-size: 1rem;
  font-weight: 600;
  color: #92400e;
}

.status-text p {
  margin: 0;
  font-size: 0.875rem;
  color: #b45309;
  line-height: 1.4;
}

.retry-btn,
.dismiss-btn {
  flex-shrink: 0;
  align-self: center;
}

.retry-btn {
  background-color: #f59e0b;
  border-color: #f59e0b;
  color: white;
}

.retry-btn:hover {
  background-color: #d97706;
  border-color: #d97706;
}

.dismiss-btn {
  background-color: transparent;
  border-color: #fbbf24;
  color: #92400e;
}

.dismiss-btn:hover {
  background-color: #fef3c7;
}

/* 暗色主题适配 */
:global(.dark) .status-content {
  background: #451a03;
  border-color: #f59e0b;
}

:global(.dark) .status-text h4 {
  color: #fef3c7;
}

:global(.dark) .status-text p {
  color: #fde68a;
}
</style>
