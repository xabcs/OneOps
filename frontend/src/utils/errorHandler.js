/**
 * 网络错误处理工具类
 * 用于检测后端服务可用性并提供友好的错误提示
 */

import { ElMessage, ElNotification } from 'element-plus'

// 错误类型枚举
export const ErrorType = {
  NETWORK_ERROR: 'NETWORK_ERROR',           // 网络错误（连接被拒绝）
  TIMEOUT: 'TIMEOUT',                       // 请求超时
  SERVER_ERROR: 'SERVER_ERROR',             // 服务器错误（5xx）
  CLIENT_ERROR: 'CLIENT_ERROR',             // 客户端错误（4xx）
  AUTH_ERROR: 'AUTH_ERROR',                 // 认证错误（401）
  PERMISSION_ERROR: 'PERMISSION_ERROR',     // 权限错误（403）
  UNKNOWN_ERROR: 'UNKNOWN_ERROR'            // 未知错误
}

// 错误提示消息
const ErrorMessages = {
  [ErrorType.NETWORK_ERROR]: '无法连接到服务器，请确认后端服务已启动',
  [ErrorType.TIMEOUT]: '请求超时，请检查网络连接或稍后重试',
  [ErrorType.SERVER_ERROR]: '服务器内部错误，请联系管理员',
  [ErrorType.AUTH_ERROR]: '登录已过期，请重新登录',
  [ErrorType.PERMISSION_ERROR]: '您没有权限执行此操作',
  [ErrorType.UNKNOWN_ERROR]: '发生未知错误，请稍后重试'
}

/**
 * 解析错误类型
 * @param {Error} error - Axios 错误对象
 * @returns {string} 错误类型
 */
export function parseErrorType(error) {
  if (!error.response) {
    // 没有响应，可能是网络错误或超时
    if (error.code === 'ECONNABORTED' || error.message?.includes('timeout')) {
      return ErrorType.TIMEOUT
    }
    if (error.message?.includes('Network Error') || error.message?.includes('ERR_CONNECTION_REFUSED')) {
      return ErrorType.NETWORK_ERROR
    }
    return ErrorType.NETWORK_ERROR // 默认为网络错误
  }

  const status = error.response.status
  if (status === 401) return ErrorType.AUTH_ERROR
  if (status === 403) return ErrorType.PERMISSION_ERROR
  if (status >= 500) return ErrorType.SERVER_ERROR
  if (status >= 400) return ErrorType.CLIENT_ERROR

  return ErrorType.UNKNOWN_ERROR
}

/**
 * 显示错误提示
 * @param {string} errorType - 错误类型
 * @param {string} customMessage - 自定义消息（可选）
 * @param {boolean} useNotification - 是否使用通知（默认使用消息）
 */
export function showError(errorType, customMessage = null, useNotification = false) {
  const message = customMessage || ErrorMessages[errorType] || ErrorMessages[ErrorType.UNKNOWN_ERROR]

  if (useNotification) {
    ElNotification({
      title: '错误',
      message,
      type: 'error',
      duration: 5000,
      showClose: true
    })
  } else {
    ElMessage({
      message,
      type: 'error',
      duration: 5000,
      showClose: true
    })
  }
}

/**
 * 处理API错误
 * @param {Error} error - Axios 错误对象
 * @param {Object} options - 配置选项
 * @returns {string} 错误类型
 */
export function handleApiError(error, options = {}) {
  const {
    showMessage = true,
    useNotification = false,
    customMessage = null,
    onNetworkError = null,
    onAuthError = null
  } = options

  const errorType = parseErrorType(error)

  if (showMessage) {
    showError(errorType, customMessage, useNotification)
  }

  // 特定错误类型的回调
  if (errorType === ErrorType.NETWORK_ERROR && onNetworkError) {
    onNetworkError()
  }

  if (errorType === ErrorType.AUTH_ERROR && onAuthError) {
    onAuthError()
  }

  return errorType
}

/**
 * 检查是否为网络错误（连接被拒绝）
 * @param {Error} error - Axios 错误对象
 * @returns {boolean}
 */
export function isNetworkError(error) {
  const errorType = parseErrorType(error)
  return errorType === ErrorType.NETWORK_ERROR
}

/**
 * 检查是否为认证错误
 * @param {Error} error - Axios 错误对象
 * @returns {boolean}
 */
export function isAuthError(error) {
  const errorType = parseErrorType(error)
  return errorType === ErrorType.AUTH_ERROR
}

/**
 * 创建后端健康检查器
 */
export class BackendHealthChecker {
  constructor() {
    this.isHealthy = true
    this.lastCheckTime = null
    this.checkInterval = 30000 // 30秒
    this.checkTimer = null
    this.listeners = new Set()
  }

  /**
   * 订阅健康状态变化
   * @param {Function} callback - 回调函数
   */
  subscribe(callback) {
    this.listeners.add(callback)
    // 返回取消订阅函数
    return () => this.listeners.delete(callback)
  }

  /**
   * 通知所有订阅者
   */
  notify(isHealthy) {
    if (this.isHealthy !== isHealthy) {
      this.isHealthy = isHealthy
      this.listeners.forEach(callback => callback(isHealthy))
    }
  }

  /**
   * 执行健康检查
   * @param {Function} apiCheck - API检查函数
   */
  async check(apiCheck) {
    this.lastCheckTime = Date.now()
    try {
      await apiCheck()
      if (!this.isHealthy) {
        this.notify(true)
      }
      return true
    } catch (error) {
      if (this.isHealthy && isNetworkError(error)) {
        this.notify(false)
      }
      return false
    }
  }

  /**
   * 开始定期检查
   * @param {Function} apiCheck - API检查函数
   */
  startPeriodicCheck(apiCheck) {
    this.stopPeriodicCheck()
    this.checkTimer = setInterval(() => {
      this.check(apiCheck)
    }, this.checkInterval)
  }

  /**
   * 停止定期检查
   */
  stopPeriodicCheck() {
    if (this.checkTimer) {
      clearInterval(this.checkTimer)
      this.checkTimer = null
    }
  }

  /**
   * 获取当前健康状态
   */
  getHealthStatus() {
    return {
      isHealthy: this.isHealthy,
      lastCheckTime: this.lastCheckTime
    }
  }
}

// 导出全局健康检查器实例
export const backendHealthChecker = new BackendHealthChecker()

export default {
  ErrorType,
  parseErrorType,
  showError,
  handleApiError,
  isNetworkError,
  isAuthError,
  BackendHealthChecker,
  backendHealthChecker
}
