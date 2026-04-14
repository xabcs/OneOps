import axios from 'axios'

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 响应拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API 请求错误:', error)
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

// 日志管理 API
export const logApi = {
  getOperationLogs: (params) => api.get('/logs/operation', { params }),
  getEventRecords: (params) => api.get('/logs/event', { params })
}

// 监控管理 API
export const monitoringApi = {
  getMonitoring: () => api.get('/monitoring')
}

export default api