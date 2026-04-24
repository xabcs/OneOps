/**
 * 存储键名常量
 * 统一管理所有 localStorage 和 sessionStorage 的键名
 */

export const STORAGE_KEYS = {
  // 认证相关 (localStorage)
  TOKEN: 'token',
  USER: 'user',
  PERMISSIONS: 'permissions',
  MENU_TREE: 'menuTree',

  // 设置相关 (localStorage)
  THEME: 'theme',

  // 后端状态相关 (sessionStorage)
  BACKEND_UNAVAILABLE: 'backendUnavailable',
  BACKEND_UNAVAILABLE_TIME: 'backendUnavailableTime'
}

/**
 * 后端状态常量
 */
export const BACKEND_STATUS = {
  CHECK_INTERVAL: 30000, // 30秒
  API_TIMEOUT: 5000,     // 5秒
  AVAILABLE: 'available',
  UNAVAILABLE: 'unavailable'
}

/**
 * 认证相关常量
 */
export const AUTH = {
  DEFAULT_HOME_PATH: '/',
  TOKEN_PREFIX: 'Bearer ',
  SESSION_TIMEOUT: 24 * 60 * 60 * 1000 // 24小时
}

/**
 * 错误类型常量
 */
export const ERROR_TYPES = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT: 'TIMEOUT',
  SERVER_ERROR: 'SERVER_ERROR',
  CLIENT_ERROR: 'CLIENT_ERROR',
  AUTH_ERROR: 'AUTH_ERROR',
  PERMISSION_ERROR: 'PERMISSION_ERROR',
  UNKNOWN_ERROR: 'UNKNOWN_ERROR'
}

/**
 * 路由相关常量
 */
export const ROUTES = {
  LOGIN: '/login',
  HOME: '/',
  DEFAULT_REDIRECT: '/'
}
