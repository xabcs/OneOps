import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/modules/auth'

const TOKEN_KEY = 'oneops_token'

// 权限白名单（不需要权限检查的接口）
const PERMISSION_WHITE_LIST = [
  '/api/v1/auth/login',
  '/api/v1/auth/logout',
  '/api/v1/auth/me',
  '/api/v1/public',
  '/api/v1/menus' // 获取菜单列表
]

// 接口权限映射（可选：用于接口级权限控制）
const API_PERMISSION_MAP: Record<string, string> = {
  // 用户管理
  'GET:/api/v1/users': 'system.user.view',
  'POST:/api/v1/users': 'system.user.create',
  'PUT:/api/v1/users': 'system.user.update',
  'DELETE:/api/v1/users': 'system.user.delete',

  // 角色管理
  'GET:/api/v1/roles': 'system.role.view',
  'POST:/api/v1/roles': 'system.role.create',
  'PUT:/api/v1/roles': 'system.role.update',
  'DELETE:/api/v1/roles': 'system.role.delete',

  // 权限管理
  'GET:/api/v1/permissions': 'system.permission.view',
  'POST:/api/v1/permissions': 'system.permission.create',
  'PUT:/api/v1/permissions': 'system.permission.update',
  'DELETE:/api/v1/permissions': 'system.permission.delete',
}

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
 * 请求拦截器：添加 token 并检查权限
 */
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore() // 在每次请求时获取 store
    const token = authStore.token || localStorage.getItem(TOKEN_KEY)

    // 添加认证 token
    if (token) {
      config.headers = config.headers || {}
      config.headers.Authorization = `Bearer ${token}`
    }

    // 权限检查逻辑
    if (shouldCheckPermission(config)) {
      const requiredPermission = getRequiredPermission(config)

      if (requiredPermission && !authStore.hasPermission(requiredPermission)) {
        const errorMsg = `权限不足：需要 ${requiredPermission} 权限`

        // 方式1：直接拦截请求（推荐）
        return Promise.reject(new Error(errorMsg))

        // 方式2：静默失败，返回特殊标记
        // config._permissionDenied = true
        // return config
      }
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

/**
 * 判断是否需要检查权限
 */
function shouldCheckPermission(config: InternalAxiosRequestConfig): boolean {
  const url = config.url || ''

  // 检查白名单
  if (PERMISSION_WHITE_LIST.some(path => url.startsWith(path))) {
    return false
  }

  // 检查是否标记为跳过权限检查
  if (config.skipPermissionCheck) {
    return false
  }

  return true
}

/**
 * 获取接口所需的权限
 */
function getRequiredPermission(config: InternalAxiosRequestConfig): string | null {
  const method = config.method?.toUpperCase() || 'GET'
  const url = config.url || ''
  const key = `${method}:${url}`

  // 如果有明确映射，使用映射的权限
  if (API_PERMISSION_MAP[key]) {
    return API_PERMISSION_MAP[key]
  }

  // 自动推断权限（基于 URL 路径）
  const match = url.match(/\/api\/v1\/(\w+)/)
  if (match) {
    const resource = match[1]
    const actionMap: Record<string, string> = {
      'GET': 'view',
      'POST': 'create',
      'PUT': 'update',
      'DELETE': 'delete'
    }
    const action = actionMap[method] || 'view'
    return `system.${resource}.${action}`
  }

  return null
}

export default request