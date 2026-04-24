import axios from 'axios'
import qs from 'qs'
import { ElMessage } from 'element-plus'
import { handleApiError, isNetworkError, isAuthError } from '../utils/errorHandler'
import { authStorage, backendStatusStorage } from '../utils/storage'
import { ROUTES } from '../constants'

// 创建 axios 实例
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  },
  paramsSerializer: params => {
    return qs.stringify(params, { arrayFormat: 'repeat', format: 'RFC1738' })
  }
})

// 请求拦截器 - 添加认证令牌
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    handleApiError(error, {
      showMessage: false,
      onAuthError: () => {
        // 认证错误时清除认证数据
        authStorage.clearAuth()

        // 只在没有已经在登录页时才重定向
        if (window.location.pathname !== ROUTES.LOGIN) {
          window.location.href = ROUTES.LOGIN
        }

        ElMessage.error('登录已过期，请重新登录')
      }
    })

    // 处理后端可用性状态
    if (isNetworkError(error)) {
      // 网络错误时标记为不可用（只设置一次）
      if (!backendStatusStorage.isUnavailable()) {
        backendStatusStorage.setUnavailable()
      }
    } else if (!isAuthError(error)) {
      // 其他错误说明后端可用
      backendStatusStorage.clearUnavailable()
    }

    return Promise.reject(error)
  }
)

// 认证 API
export const loginApi = {
  login: (data) => api.post('/login', data),
  register: (data) => api.post('/register', data),
  getUserInfo: () => api.get('/user/info')
}

// 系统管理 API
export const systemApi = {
  getMenus: () => api.get('/system/menus'),
  addMenu: (data) => api.post('/system/menus', data),
  updateMenu: (id, data) => api.put(`/system/menus/${id}`, data),
  deleteMenu: (id) => api.delete(`/system/menus/${id}`),
  getRoles: () => api.get('/system/roles'),
  addRole: (data) => api.post('/system/roles', data),
  updateRole: (id, data) => api.put(`/system/roles/${id}`, data),
  deleteRole: (id) => api.delete(`/system/roles/${id}`),
  getUsers: () => api.get('/system/users'),
  addUser: (data) => api.post('/system/users', data),
  updateUser: (id, data) => api.put(`/system/users/${id}`, data),
  deleteUser: (id) => api.delete(`/system/users/${id}`)
}

// 服务器管理 API
export const serverApi = {
  getServers: (params) => api.get('/servers', { params }),
  getGroups: () => api.get('/groups'),
  addServer: (data) => api.post('/servers', data)
}

// 任务管理 API
export const taskApi = {
  getTasks: (params) => api.get('/tasks', { params })
}

// 容器管理 API
export const containerApi = {
  getContainers: (params) => api.get('/containers', { params }),
  getImages: (params) => api.get('/images', { params })
}

// 审计管理 API
export const auditApi = {
  // 登录日志
  getLoginLogs: (params) => api.get('/audit/login-logs', { params }),
  exportLoginLogs: (params) => api.get('/audit/login-logs/export', { params, responseType: 'blob' }),
  
  // 操作日志
  getOperationLogs: (params) => api.get('/audit/operation-logs', { params }),
  exportOperationLogs: (params) => api.get('/audit/operation-logs/export', { params, responseType: 'blob' }),
  
  // 系统事件日志
  getSystemEventLogs: (params) => api.get('/audit/system-event-logs', { params }),

  // 审计统计
  getAuditStats: () => api.get('/audit/stats'),

  // 可用模块列表
  getModules: () => api.get('/audit/modules')
}

// 监控管理 API
export const monitoringApi = {
  getMonitoring: () => api.get('/monitoring'),
  getGrafanaUrl: () => api.get('/monitoring/grafana/url'),
  getStats: () => api.get('/monitoring/stats'),
  refresh: () => api.post('/monitoring/refresh'),
  handleAlert: (data) => api.post('/monitoring/alert/handle', data)
}

// 证书监控 API
export const certificateMonitoringApi = {
  getGrafanaUrl: () => api.get('/monitoring/grafana/url')
}

export default api