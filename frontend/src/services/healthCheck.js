/**
 * 后端健康检查服务
 * 用于定期检查后端服务的可用性
 */

import { loginApi } from '../api/index.js'
import { isNetworkError } from '../utils/errorHandler'

class HealthCheckService {
  constructor() {
    this.checkInterval = 30000 // 30秒检查一次
    this.checkTimer = null
    this.isChecking = false
    this.lastCheckResult = null
    this.listeners = new Set()
  }

  /**
   * 订阅健康状态变化
   * @param {Function} callback - 回调函数，接收 (isHealthy, error) 参数
   * @returns {Function} 取消订阅函数
   */
  subscribe(callback) {
    this.listeners.add(callback)
    // 立即通知当前状态
    if (this.lastCheckResult !== null) {
      callback(this.lastCheckResult.isHealthy, this.lastCheckResult.error)
    }
    // 返回取消订阅函数
    return () => this.listeners.delete(callback)
  }

  /**
   * 通知所有订阅者状态变化
   * @param {boolean} isHealthy - 是否健康
   * @param {Error} error - 错误对象（可选）
   */
  notify(isHealthy, error = null) {
    this.lastCheckResult = { isHealthy, error }
    this.listeners.forEach(callback => {
      try {
        callback(isHealthy, error)
      } catch (e) {
        console.error('Health check callback error:', e)
      }
    })
  }

  /**
   * 执行一次健康检查
   * @returns {Promise<boolean>} 是否健康
   */
  async check() {
    if (this.isChecking) {
      return this.lastCheckResult?.isHealthy ?? true
    }

    // 如果没有认证，直接跳过检查（假设后端可用）
    const hasToken = !!localStorage.getItem('token')
    if (!hasToken) {
      // 如果之前标记为不可用，现在尝试清除标记
      const wasUnavailable = !!sessionStorage.getItem('backendUnavailable')
      if (wasUnavailable) {
        this.notify(true) // 标记为可用
      }
      return true
    }

    this.isChecking = true
    try {
      // 尝试获取用户信息来检查后端是否可用
      // 使用轻量级的请求，设置较短的超时时间
      await Promise.race([
        loginApi.getUserInfo(),
        new Promise((_, reject) =>
          setTimeout(() => reject(new Error('Health check timeout')), 5000)
        )
      ])

      // 检查成功，后端可用，清除不可用标记
      if (sessionStorage.getItem('backendUnavailable')) {
        sessionStorage.removeItem('backendUnavailable')
        sessionStorage.removeItem('backendUnavailableTime')
      }
      this.notify(true)
      return true
    } catch (error) {
      // 检查失败
      const isNetworkErr = isNetworkError(error)
      this.notify(false, error)

      // 标记后端不可用
      if (isNetworkErr) {
        sessionStorage.setItem('backendUnavailable', 'true')
        sessionStorage.setItem('backendUnavailableTime', Date.now().toString())
      }

      if (!isNetworkErr) {
        console.warn('Health check failed (non-network error):', error.message)
      }

      return false
    } finally {
      this.isChecking = false
    }
  }

  /**
   * 开始定期健康检查
   */
  start() {
    this.stop() // 确保没有重复的定时器

    // 立即执行一次检查
    this.check()

    // 启动定期检查
    this.checkTimer = setInterval(() => {
      this.check()
    }, this.checkInterval)
  }

  /**
   * 停止定期健康检查
   */
  stop() {
    if (this.checkTimer) {
      clearInterval(this.checkTimer)
      this.checkTimer = null
    }
  }

  /**
   * 获取当前健康状态
   * @returns {{ isHealthy: boolean|null, lastCheckTime: number, error: Error|null }}
   */
  getStatus() {
    return {
      isHealthy: this.lastCheckResult?.isHealthy ?? null,
      lastCheckTime: this.lastCheckResult ? Date.now() : null,
      error: this.lastCheckResult?.error ?? null
    }
  }

  /**
   * 设置检查间隔
   * @param {number} interval - 检查间隔（毫秒）
   */
  setCheckInterval(interval) {
    this.checkInterval = interval
    // 如果正在运行，重新启动以应用新的间隔
    if (this.checkTimer) {
      this.start()
    }
  }
}

// 导出全局单例实例
export const healthCheckService = new HealthCheckService()

export default healthCheckService
