/**
 * 统一的存储管理工具
 * 消除重复的 localStorage/sessionStorage 操作逻辑
 */

import { STORAGE_KEYS } from '../constants'

/**
 * 认证存储管理
 */
export const authStorage = {
  /**
   * 保存认证数据
   */
  saveAuth(token, user, permissions = [], menuTree = []) {
    localStorage.setItem(STORAGE_KEYS.TOKEN, token)
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user))
    localStorage.setItem(STORAGE_KEYS.PERMISSIONS, JSON.stringify(permissions))
    localStorage.setItem(STORAGE_KEYS.MENU_TREE, JSON.stringify(menuTree))
  },

  /**
   * 清除认证数据
   */
  clearAuth() {
    localStorage.removeItem(STORAGE_KEYS.TOKEN)
    localStorage.removeItem(STORAGE_KEYS.USER)
    localStorage.removeItem(STORAGE_KEYS.PERMISSIONS)
    localStorage.removeItem(STORAGE_KEYS.MENU_TREE)
  },

  /**
   * 获取认证数据
   */
  getAuth() {
    return {
      token: localStorage.getItem(STORAGE_KEYS.TOKEN),
      user: this.getUser(),
      permissions: this.getPermissions(),
      menuTree: this.getMenuTree()
    }
  },

  /**
   * 获取 token
   */
  getToken() {
    return localStorage.getItem(STORAGE_KEYS.TOKEN)
  },

  /**
   * 获取用户信息
   */
  getUser() {
    try {
      const userStr = localStorage.getItem(STORAGE_KEYS.USER)
      return userStr ? JSON.parse(userStr) : null
    } catch (e) {
      console.error('Failed to parse user from localStorage:', e)
      return null
    }
  },

  /**
   * 获取权限列表
   */
  getPermissions() {
    try {
      const permissionsStr = localStorage.getItem(STORAGE_KEYS.PERMISSIONS)
      return permissionsStr ? JSON.parse(permissionsStr) : []
    } catch (e) {
      console.error('Failed to parse permissions from localStorage:', e)
      return []
    }
  },

  /**
   * 获取菜单树
   */
  getMenuTree() {
    try {
      const menuTreeStr = localStorage.getItem(STORAGE_KEYS.MENU_TREE)
      return menuTreeStr ? JSON.parse(menuTreeStr) : []
    } catch (e) {
      console.error('Failed to parse menuTree from localStorage:', e)
      return []
    }
  }
}

/**
 * 后端状态存储管理
 */
export const backendStatusStorage = {
  /**
   * 设置后端不可用状态
   */
  setUnavailable() {
    sessionStorage.setItem(STORAGE_KEYS.BACKEND_UNAVAILABLE, 'true')
    sessionStorage.setItem(STORAGE_KEYS.BACKEND_UNAVAILABLE_TIME, Date.now().toString())
  },

  /**
   * 清除后端不可用状态
   */
  clearUnavailable() {
    sessionStorage.removeItem(STORAGE_KEYS.BACKEND_UNAVAILABLE)
    sessionStorage.removeItem(STORAGE_KEYS.BACKEND_UNAVAILABLE_TIME)
  },

  /**
   * 检查后端是否不可用
   */
  isUnavailable() {
    return !!sessionStorage.getItem(STORAGE_KEYS.BACKEND_UNAVAILABLE)
  },

  /**
   * 获取不可用时间戳
   */
  getUnavailableTime() {
    const timeStr = sessionStorage.getItem(STORAGE_KEYS.BACKEND_UNAVAILABLE_TIME)
    return timeStr ? parseInt(timeStr) : 0
  }
}

/**
 * 主题设置存储管理
 */
export const themeStorage = {
  /**
   * 保存主题
   */
  saveTheme(theme) {
    localStorage.setItem(STORAGE_KEYS.THEME, theme)
  },

  /**
   * 获取主题
   */
  getTheme() {
    return localStorage.getItem(STORAGE_KEYS.THEME) || 'light'
  }
}

/**
 * 安全的 JSON 解析辅助函数
 */
function safeJSONParse(str, defaultValue = null) {
  try {
    return str ? JSON.parse(str) : defaultValue
  } catch (e) {
    console.error('JSON parse error:', e)
    return defaultValue
  }
}
