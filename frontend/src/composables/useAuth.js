/**
 * 认证相关 composable
 * 统一管理认证状态和操作
 */

import { computed } from 'vue'
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'
import { loginApi } from '../api'
import { authStorage } from '../utils/storage'
import { ElMessage } from 'element-plus'

export function useAuth() {
  const store = useStore()
  const router = useRouter()

  // 认证状态
  const isAuthenticated = computed(() => store.getters.isAuthenticated)
  const user = computed(() => store.getters.user)
  const permissions = computed(() => store.getters.permissions)
  const menuTree = computed(() => store.getters.menuTree)

  /**
   * 登录
   */
  async function login(credentials) {
    try {
      const response = await loginApi.login(credentials)

      if (response.success) {
        const { token, user } = response.data

        // 保存认证信息
        authStorage.saveAuth(
          token,
          user,
          user.permissions || [],
          user.menuTree || []
        )

        // 更新 Vuex store
        store.commit('SET_TOKEN', token)
        store.commit('SET_USER', user)
        store.commit('SET_PERMISSIONS', user.permissions || [])
        store.commit('SET_MENU_TREE', user.menuTree || [])

        // 清除后端不可用状态
        store.commit('SET_BACKEND_AVAILABLE', true)

        ElMessage.success(`欢迎回来, ${user.username}`)

        return { success: true, user }
      } else {
        ElMessage.error(response.message || '登录失败')
        return { success: false, message: response.message }
      }
    } catch (error) {
      console.error('Login error:', error)

      if (error.response) {
        ElMessage.error(error.response.data.message || '登录失败')
      } else {
        // 网络错误
        ElMessage.error({
          message: '无法连接到服务器，请确认后端服务已启动（端口 8082）',
          duration: 5000,
          showClose: true
        })
        // 标记后端不可用
        store.commit('SET_BACKEND_AVAILABLE', false)
      }

      return { success: false, error }
    }
  }

  /**
   * 登出
   */
  function logout() {
    // 清除存储
    authStorage.clearAuth()

    // 清除 Vuex store
    store.dispatch('logout')

    // 跳转到登录页
    router.push('/login')
  }

  /**
   * 获取用户信息
   */
  async function fetchUserInfo() {
    if (!isAuthenticated.value) return null

    try {
      const res = await loginApi.getUserInfo()

      if (res.code === 200) {
        const oldUserId = user.value?.id

        // 更新 store
        store.commit('SET_USER', res.data)
        store.commit('SET_PERMISSIONS', res.data.permissions)
        store.commit('SET_MENU_TREE', res.data.menuTree)

        // 后端恢复可用
        store.commit('SET_BACKEND_AVAILABLE', true)

        return res.data
      }
    } catch (error) {
      console.error('Error fetching user info:', error)
      // 网络错误时标记后端不可用
      if (error.message?.includes('Network Error') || error.message?.includes('ERR_CONNECTION_REFUSED')) {
        store.commit('SET_BACKEND_AVAILABLE', false)
      }
    }

    return null
  }

  /**
   * 检查权限
   */
  function hasPermission(permission) {
    if (!permission) return true
    return permissions.value.includes('*:*:*') || permissions.value.includes(permission)
  }

  return {
    // 状态
    isAuthenticated,
    user,
    permissions,
    menuTree,

    // 方法
    login,
    logout,
    fetchUserInfo,
    hasPermission
  }
}
