/**
 * 日志工具
 *
 * 提供统一的日志接口，可以根据环境自动控制日志输出
 * 开发环境显示所有日志，生产环境只显示错误和警告
 */

const isDev = import.meta.env.DEV
const isDebug = import.meta.env.VITE_DEBUG === 'true'

class Logger {
  private prefix: string

  constructor(prefix: string = '') {
    this.prefix = prefix
  }

  private log(level: string, ...args: any[]) {
    if (isDev || isDebug) {
      const timestamp = new Date().toISOString()
      const message = `[${timestamp}] [${level}]${this.prefix ? ' [' + this.prefix + ']' : ''}`

      switch (level) {
        case 'debug':
          console.log(message, ...args)
          break
        case 'info':
          console.info(message, ...args)
          break
        case 'warn':
          console.warn(message, ...args)
          break
        case 'error':
          console.error(message, ...args)
          break
        default:
          console.log(message, ...args)
      }
    }
  }

  debug(...args: any[]) {
    if (isDev || isDebug) {
      this.log('debug', ...args)
    }
  }

  info(...args: any[]) {
    this.log('info', ...args)
  }

  warn(...args: any[]) {
    this.log('warn', ...args)
  }

  error(...args: any[]) {
    this.log('error', ...args)
  }
}

// 创建默认日志实例
export const logger = new Logger()

// 创建带前缀的日志实例
export function createLogger(prefix: string): Logger {
  return new Logger(prefix)
}

// 向后兼容的 console.log 替代方案
export const debug = logger.debug.bind(logger)
export const info = logger.info.bind(logger)
export const warn = logger.warn.bind(logger)
export const error = logger.error.bind(logger)