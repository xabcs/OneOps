/**
 * 后端状态管理 composable
 * 统一管理后端可用性状态
 */

import { computed } from 'vue'
import { useStore } from 'vuex'
import { backendStatusStorage } from '../utils/storage'

export function useBackendStatus() {
  const store = useStore()

  // 后端可用状态
  const isBackendAvailable = computed(() => store.getters.isBackendAvailable)
  const backendUnavailableTime = computed(() => store.getters.backendUnavailableTime)

  /**
   * 设置后端可用
   */
  function setBackendAvailable() {
    backendStatusStorage.clearUnavailable()
    store.commit('SET_BACKEND_AVAILABLE', true)
  }

  /**
   * 设置后端不可用
   */
  function setBackendUnavailable() {
    backendStatusStorage.setUnavailable()
    store.commit('SET_BACKEND_AVAILABLE', false)
  }

  /**
   * 重试连接
   */
  async function retryConnection() {
    try {
      // 尝试调用 API 来检查连接
      const { loginApi } = await import('../api')
      await loginApi.getUserInfo()

      // 成功则标记为可用
      setBackendAvailable()
      return true
    } catch (error) {
      // 失败则保持不可用状态
      return false
    }
  }

  return {
    isBackendAvailable,
    backendUnavailableTime,
    setBackendAvailable,
    setBackendUnavailable,
    retryConnection
  }
}
