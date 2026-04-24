/**
 * 统一的后端健康检查服务
 * 替代之前重复的 BackendHealthChecker 和 HealthCheckService
 */

import { loginApi } from '../api'
import { isNetworkError } from '../utils/errorHandler'
import { backendStatusStorage } from '../utils/storage'
import { BACKEND_STATUS } from '../constants'

class UnifiedHealthChecker {
  constructor() {
    this.isHealthy = true
    this.isChecking = false
    this.lastCheckTime = null
    this.checkTimer = null
    this.listeners = new Set()
  }

  /**
   * 订阅健康状态变化
   * @param {Function} callback - 回调函数 (isHealthy) => {}
   * @returns {Function} 取消订阅函数
   */
  subscribe(callback) {
    this.listeners.add(callback)
    return () => this.listeners.delete(callback)
  }

  /**
   * 通知所有订阅者
   */
  notify(isHealthy) {
    if (this.isHealthy !== isHealthy) {
      this.isHealthy = isHealthy
      this.listeners.forEach(callback => {
        try {
          callback(isHealthy)
        } catch (e) {
          console.error('Health check callback error:', e)
        }
      })
    }
  }

  /**
   * 执行健康检查
   */
  async check() {
    if (this.isChecking) {
      return this.isHealthy
    }

    // 如果没有认证，跳过检查
    const hasToken = localStorage.getItem('token')
    if (!hasToken) {
      this.notify(true)
      return true
    }

    this.isChecking = true
    try {
      // 使用 Promise.race 实现超时
      await Promise.race([
        loginApi.getUserInfo(),
        new Promise((_, reject) =>
          setTimeout(() => reject(new Error('Health check timeout')), BACKEND_STATUS.API_TIMEOUT)
        )
      ])

      // 检查成功，清除不可用标记
      backendStatusStorage.clearUnavailable()
      this.notify(true)
      return true
    } catch (error) {
      const isNetworkErr = isNetworkError(error)

      // 只有网络错误才标记为不可用
      if (isNetworkErr) {
        backendStatusStorage.setUnavailable()
      }

      this.notify(!isNetworkErr)
      return !isNetworkErr
    } finally {
      this.isChecking = false
      this.lastCheckTime = Date.now()
    }
  }

  /**
   * 开始定期检查
   */
  start() {
    this.stop()
    // 立即执行一次检查
    this.check()
    // 启动定期检查
    this.checkTimer = setInterval(() => {
      this.check()
    }, BACKEND_STATUS.CHECK_INTERVAL)
  }

  /**
   * 停止定期检查
   */
  stop() {
    if (this.checkTimer) {
      clearInterval(this.checkTimer)
      this.checkTimer = null
    }
  }

  /**
   * 手动设置为可用（外部调用）
   */
  setAvailable() {
    backendStatusStorage.clearUnavailable()
    this.notify(true)
  }

  /**
   * 手动设置为不可用（外部调用）
   */
  setUnavailable() {
    backendStatusStorage.setUnavailable()
    this.notify(false)
  }

  /**
   * 获取当前状态
   */
  getStatus() {
    return {
      isHealthy: this.isHealthy,
      lastCheckTime: this.lastCheckTime,
      isUnavailable: backendStatusStorage.isUnavailable()
    }
  }
}

// 导出单例实例
export const healthChecker = new UnifiedHealthChecker()
export default healthChecker
