import { createRouter, createWebHistory } from 'vue-router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import Home from '../views/Home.vue'
import store from '../store'

// 配置 NProgress
NProgress.configure({ 
  showSpinner: false,
  easing: 'ease',
  speed: 200,
  trickleSpeed: 100
})

// 预加载核心管理组件，减少“第一次点击”的延迟感
import Menus from '../views/system/Menus.vue'
import Roles from '../views/system/Roles.vue'
import Users from '../views/system/Users.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false, title: '登录' }
  },
  {
    path: '/',
    name: 'Home',
    component: Home,
    meta: { requiresAuth: true, title: '控制台' }
  },
  {
    path: '/servers',
    name: 'Servers',
    component: () => import('../views/Servers.vue'),
    meta: { requiresAuth: true, title: '主机管理', parent: '资产管理' }
  },
  {
    path: '/tasks',
    name: 'Tasks',
    component: () => import('../views/Tasks.vue'),
    meta: { requiresAuth: true, title: '任务列表', parent: '自动化任务' }
  },
  {
    path: '/monitoring',
    name: 'Monitoring',
    component: () => import('../views/Monitoring.vue'),
    meta: { requiresAuth: true, title: '系统监控', parent: '监控中心', permission: 'menu:monitoring' }
  },
  {
    path: '/system/menus',
    name: 'Menus',
    component: Menus,
    meta: { requiresAuth: true, title: '菜单管理', parent: '系统管理', permission: 'menu:system:menus' }
  },
  {
    path: '/system/roles',
    name: 'Roles',
    component: Roles,
    meta: { requiresAuth: true, title: '角色管理', parent: '系统管理', permission: 'menu:system:roles' }
  },
  {
    path: '/system/users',
    name: 'Users',
    component: Users,
    meta: { requiresAuth: true, title: '用户管理', parent: '系统管理', permission: 'menu:system:users' }
  },
  {
    path: '/audit/behavior',
    name: 'BehaviorRecords',
    component: () => import('../views/audit/behavior/Index.vue'),
    redirect: '/audit/behavior/login',
    meta: { requiresAuth: true, title: '行为审计', parent: '审计管理', permission: 'menu:audit:behavior' },
    children: [
      {
        path: 'login',
        name: 'LoginLogs',
        component: () => import('../views/audit/behavior/LoginLogs.vue'),
        meta: { requiresAuth: true, title: '登录日志', parent: '行为审计', permission: 'menu:audit:behavior:login' }
      },
      {
        path: 'operation',
        name: 'OperationLogs',
        component: () => import('../views/audit/behavior/OperationLogs.vue'),
        meta: { requiresAuth: true, title: '操作日志', parent: '行为审计', permission: 'menu:audit:behavior:operation' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  NProgress.start()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const isAuthenticated = !!localStorage.getItem('token')
  
  if (requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && isAuthenticated) {
    next('/')
  } else {
    // 权限校验
    let permissions = []
    try {
      permissions = JSON.parse(localStorage.getItem('permissions') || '[]')
    } catch (e) {
      console.error('Failed to parse permissions from localStorage in router:', e)
    }
    const hasPermission = permissions.includes('*:*:*') || (to.meta.permission && permissions.includes(to.meta.permission))
    
    if (to.meta.permission && !hasPermission) {
      console.warn(`无权访问: ${to.path}`)
      NProgress.done()
      next('/')
    } else {
      next()
    }
  }
})

router.afterEach(() => {
  NProgress.done()
})

export default router