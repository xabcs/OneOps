import { createStore } from 'vuex'
import { authStorage, backendStatusStorage, themeStorage } from '../utils/storage'

export default createStore({
  state: {
    token: authStorage.getToken() || '',
    user: authStorage.getUser(),
    permissions: authStorage.getPermissions(),
    menuTree: authStorage.getMenuTree(),
    theme: themeStorage.getTheme(),
    isBackendAvailable: !backendStatusStorage.isUnavailable(),
    backendUnavailableTime: backendStatusStorage.getUnavailableTime()
  },
  mutations: {
    SET_TOKEN(state, token) {
      state.token = token
      localStorage.setItem('token', token)
    },
    SET_USER(state, user) {
      state.user = user
      try {
        const seen = new WeakSet()
        const safeUserStr = JSON.stringify(user, (key, value) => {
          if (typeof value === 'object' && value !== null) {
            if (seen.has(value)) return undefined
            seen.add(value)
          }
          if (key === 'component' || key === 'vnode' || (value && value._isVue)) return undefined
          return value
        })
        localStorage.setItem('user', safeUserStr)
      } catch (e) {
        console.warn('Failed to persist user to localStorage:', e)
        try {
          const basicUser = { id: user.id, username: user.username, roleNames: user.roleNames }
          localStorage.setItem('user', JSON.stringify(basicUser))
        } catch (innerE) {
          console.error('Critical failure persisting user:', innerE)
        }
      }
    },
    SET_PERMISSIONS(state, permissions) {
      state.permissions = permissions
      try {
        const seen = new WeakSet()
        const safePermissionsStr = JSON.stringify(permissions, (key, value) => {
          if (typeof value === 'object' && value !== null) {
            if (seen.has(value)) return undefined
            seen.add(value)
          }
          return value
        })
        localStorage.setItem('permissions', safePermissionsStr)
      } catch (e) {
        console.error('Failed to persist permissions to localStorage:', e)
      }
    },
    SET_MENU_TREE(state, menuTree) {
      state.menuTree = menuTree
      try {
        const seen = new WeakSet()
        const safeMenuTreeStr = JSON.stringify(menuTree, (key, value) => {
          if (typeof value === 'object' && value !== null) {
            if (seen.has(value)) return undefined
            seen.add(value)
          }
          if (key === 'component' || key === 'vnode' || (value && value._isVue)) return undefined
          return value
        })
        localStorage.setItem('menuTree', safeMenuTreeStr)
      } catch (e) {
        console.warn('Failed to persist menuTree to localStorage:', e)
      }
    },
    LOGOUT(state) {
      state.token = ''
      state.user = null
      state.permissions = []
      state.menuTree = []
      authStorage.clearAuth()
    },
    SET_THEME(state, theme) {
      state.theme = theme
      themeStorage.saveTheme(theme)
      document.documentElement.setAttribute('data-theme', theme)
      if (theme === 'dark') {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    },
    SET_BACKEND_AVAILABLE(state, isAvailable) {
      state.isBackendAvailable = isAvailable
      if (isAvailable) {
        backendStatusStorage.clearUnavailable()
        state.backendUnavailableTime = 0
      } else {
        backendStatusStorage.setUnavailable()
        state.backendUnavailableTime = Date.now()
      }
    }
  },
  actions: {
    login({ commit }, { token, user }) {
      commit('SET_TOKEN', token)
      commit('SET_USER', user)
    },
    logout({ commit }) {
      commit('LOGOUT')
    },
    async fetchUserInfo({ commit }) {
      try {
        const { loginApi } = await import('../api/index.js')
        const res = await loginApi.getUserInfo()
        if (res.code === 200) {
          commit('SET_USER', res.data)
          commit('SET_PERMISSIONS', res.data.permissions || [])
          commit('SET_MENU_TREE', res.data.menuTree || [])
          return res.data
        }
      } catch (error) {
        console.error('Failed to fetch user info:', error)
      }
    },
    initializeAuth({ commit, state }) {
      const token = localStorage.getItem('token')
      let user = null
      let permissions = []
      let menuTree = []
      
      try {
        user = JSON.parse(localStorage.getItem('user') || 'null')
      } catch (e) {
        console.error('Failed to parse user from localStorage in initializeAuth:', e)
      }
      
      try {
        permissions = JSON.parse(localStorage.getItem('permissions') || '[]')
      } catch (e) {
        console.error('Failed to parse permissions from localStorage in initializeAuth:', e)
      }
      
      try {
        menuTree = JSON.parse(localStorage.getItem('menuTree') || '[]')
      } catch (e) {
        console.error('Failed to parse menuTree from localStorage in initializeAuth:', e)
      }
      
      const theme = localStorage.getItem('theme') || 'light'
      
      document.documentElement.setAttribute('data-theme', theme)
      commit('SET_THEME', theme)

      if (token && user) {
        commit('SET_TOKEN', token)
        commit('SET_USER', user)
        commit('SET_PERMISSIONS', permissions)
        commit('SET_MENU_TREE', menuTree)
      } else {
        commit('LOGOUT')
      }
    }
  },
  getters: {
    isAuthenticated: state => !!state.token,
    user: state => state.user,
    permissions: state => state.permissions,
    menuTree: state => state.menuTree,
    currentTheme: state => state.theme,
    isBackendAvailable: state => state.isBackendAvailable,
    backendUnavailableTime: state => state.backendUnavailableTime,
    hasPermission: state => (permission) => {
      if (!permission) return true
      return state.permissions.includes('*:*:*') || state.permissions.includes(permission)
    }
  }
})
