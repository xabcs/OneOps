import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/modules/auth'

const TOKEN_KEY = 'oneops_token'

const request = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

/**
 * 提取错误信息
 */
async function extractErrorMessage(error: any): Promise<string> {
  if (error.code === 'ECONNABORTED' || String(error.message || '').toLowerCase().includes('timeout')) {
    return '请求超时，请稍后重试'
  }

  const response = error.response
  const data = response?.data

  if (typeof Blob !== 'undefined' && data instanceof Blob) {
    try {
      const text = await data.text()
      if (!text) return error.message || '请求失败'
      try {
        const parsed = JSON.parse(text)
        return parsed.error || parsed.message || text
      } catch {
        return text
      }
    } catch {
      return error.message || '请求失败'
    }
  }

  return data?.error || data?.message || error.message || '请求失败'
}

/**
 * 重定向到登录页
 */
function redirectToLogin() {
  if (window.location.pathname.startsWith('/login')) return
  const redirect = encodeURIComponent(window.location.pathname + window.location.search)
  window.location.href = `/login?redirect=${redirect}`
}

/**
 * 处理会话过期
 */
let isHandlingSessionExpired = false
function handleSessionExpired() {
  const authStore = useAuthStore()

  if (isHandlingSessionExpired) return
  isHandlingSessionExpired = true

  authStore.clearSession()
  ElMessage.error('登录状态已过期，请重新登录')

  window.setTimeout(() => {
    redirectToLogin()
    isHandlingSessionExpired = false
  }, 700)
}

/**
 * 请求拦截器：添加 token
 */
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore()
    const token = authStore.token || localStorage.getItem(TOKEN_KEY)

    // 添加认证 token
    if (token) {
      config.headers = config.headers || {}
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => Promise.reject(error)
)

/**
 * 响应拦截器：处理权限错误
 */
request.interceptors.response.use(
  (response) => {
    // 检查自定义权限标记
    if (response.config._permissionDenied) {
      return Promise.reject(new Error('权限不足'))
    }
    return response.data
  },
  async (error: AxiosError) => {
    const status = error.response?.status
    const config = error.config as any

    // 处理权限不足
    if (status === 403) {
      const errorMsg = error.response?.data?.error || '权限不足'
      ElMessage.error(errorMsg)
      return Promise.reject(error)
    }

    // 处理未授权
    if (status === 401) {
      const authStore = useAuthStore() // 在需要时获取 store
      const isProfileRequest = String(config?.url || '').includes('/auth/me')

      if (!isProfileRequest) {
        authStore.clearSession()
      }
    }

    // 处理其他错误
    if (!(status === 401) && !config?.skipErrorMessage) {
      const errorMsg = await extractErrorMessage(error)
      ElMessage.error(errorMsg)
    }

    return Promise.reject(error)
  }
)

export default request