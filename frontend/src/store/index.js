import { createStore } from 'vuex'

export default createStore({
  state: {
    token: localStorage.getItem('token') || '',
    user: (() => {
      try {
        return JSON.parse(localStorage.getItem('user') || 'null')
      } catch (e) {
        console.error('Failed to parse user from localStorage:', e)
        return null
      }
    })(),
    permissions: (() => {
      try {
        return JSON.parse(localStorage.getItem('permissions') || '[]')
      } catch (e) {
        console.error('Failed to parse permissions from localStorage:', e)
        return []
      }
    })(),
    menuTree: (() => {
      try {
        return JSON.parse(localStorage.getItem('menuTree') || '[]')
      } catch (e) {
        console.error('Failed to parse menuTree from localStorage:', e)
        return []
      }
    })(),
    theme: localStorage.getItem('theme') || 'light'
  },
  mutations: {
    SET_TOKEN(state, token) {
      state.token = token
      localStorage.setItem('token', token)
    },
    SET_USER(state, user) {
      state.user = user
      try {
        localStorage.setItem('user', JSON.stringify(user))
      } catch (e) {
        console.error('Failed to persist user to localStorage:', e)
      }
    },
    SET_PERMISSIONS(state, permissions) {
      state.permissions = permissions
      try {
        localStorage.setItem('permissions', JSON.stringify(permissions))
      } catch (e) {
        console.error('Failed to persist permissions to localStorage:', e)
      }
    },
    SET_MENU_TREE(state, menuTree) {
      state.menuTree = menuTree
      try {
        localStorage.setItem('menuTree', JSON.stringify(menuTree))
      } catch (e) {
        console.error('Failed to persist menuTree to localStorage:', e)
      }
    },
    LOGOUT(state) {
      state.token = ''
      state.user = null
      state.permissions = []
      state.menuTree = []
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      localStorage.removeItem('permissions')
      localStorage.removeItem('menuTree')
    },
    SET_THEME(state, theme) {
      state.theme = theme
      localStorage.setItem('theme', theme)
      document.documentElement.setAttribute('data-theme', theme)
      if (theme === 'dark') {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
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
    hasPermission: state => (permission) => {
      if (!permission) return true
      return state.permissions.includes('*:*:*') || state.permissions.includes(permission)
    }
  }
})
