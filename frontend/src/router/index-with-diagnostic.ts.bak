// frontend/src/router/index.ts
// 完整的路由配置，包含诊断功能

import { createRouter, createWebHistory } from 'vue-router';
import { useUserStore } from '@/store/user';
import { usePermissionStore } from '@/store/permission';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      name: 'Home',
      component: () => import('@/views/home/index.vue'),
      meta: { requiresAuth: true }
    },
    // 系统管理
    {
      path: '/system',
      name: 'System',
      component: () => import('@/views/system/layout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: 'users',
          name: 'SystemUsers',
          component: () => import('@/views/system/users/index.vue'),
          meta: {
            title: '用户管理',
            requiresAuth: true,
            permission: 'system:users:view'
          }
        },
        {
          path: 'roles',
          name: 'SystemRoles',
          component: () => import('@/views/system/roles/index.vue'),
          meta: {
            title: '角色管理',
            requiresAuth: true,
            permission: 'system:roles:view'
          }
        },
        {
          path: 'menus',
          name: 'SystemMenus',
          component: () => import('@/views/system/menus/index.vue'),
          meta: {
            title: '菜单管理',
            requiresAuth: true,
            permission: 'system:menus:view'
          }
        }
      ]
    },
    // 审计日志
    {
      path: '/audit',
      name: 'Audit',
      component: () => import('@/views/audit/layout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: 'login',
          name: 'AuditLogin',
          component: () => import('@/views/audit/login.vue'),
          meta: {
            title: '登录日志',
            requiresAuth: true,
            permission: 'audit:login:view'
          }
        },
        {
          path: 'operation',
          name: 'AuditOperation',
          component: () => import('@/views/audit/operation.vue'),
          meta: {
            title: '操作日志',
            requiresAuth: true,
            permission: 'audit:operation:view'
          }
        }
      ]
    },
    // K8s集群管理
    {
      path: '/k8s',
      name: 'K8s',
      component: () => import('@/views/k8s/layout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: 'clusters',
          name: 'K8sClusters',
          component: () => import('@/views/k8s/clusters/index.vue'),
          meta: {
            title: '集群管理',
            requiresAuth: true,
            permission: 'k8s:clusters:view'
          }
        },
        {
          path: 'nodes',
          name: 'K8sNodes',
          component: () => import('@/views/k8s/nodes/index.vue'),
          meta: {
            title: '节点管理',
            requiresAuth: true,
            permission: 'k8s:nodes:view'
          }
        },
        {
          path: 'namespaces',
          name: 'K8sNamespaces',
          component: () => import('@/views/k8s/namespaces/index.vue'),
          meta: {
            title: '命名空间',
            requiresAuth: true,
            permission: 'k8s:namespaces:view'
          }
        },
        {
          path: 'pods',
          name: 'K8sPods',
          component: () => import('@/views/k8s/pods/index.vue'),
          meta: {
            title: 'Pod管理',
            requiresAuth: true,
            permission: 'k8s:pods:view'
          }
        },
        {
          path: 'services',
          name: 'K8sServices',
          component: () => import('@/views/k8s/services/index.vue'),
          meta: {
            title: '服务管理',
            requiresAuth: true,
            permission: 'k8s:services:view'
          }
        },
        {
          path: 'deployments',
          name: 'K8sDeployments',
          component: () => import('@/views/k8s/deployments/index.vue'),
          meta: {
            title: '部署管理',
            requiresAuth: true,
            permission: 'k8s:deployments:view'
          }
        },
        // 监控子菜单
        {
          path: 'monitoring/pods',
          name: 'K8sMonitoringPods',
          component: () => import('@/views/k8s/monitoring/pods.vue'),
          meta: {
            title: 'Pod监控',
            requiresAuth: true,
            permission: 'k8s:monitoring:view'
          }
        },
        {
          path: 'monitoring/nodes',
          name: 'K8sMonitoringNodes',
          component: () => import('@/views/k8s/monitoring/nodes.vue'),
          meta: {
            title: '节点监控',
            requiresAuth: true,
            permission: 'k8s:monitoring:view'
          }
        },
        // 新增：诊断中心
        {
          path: 'diagnostic',
          name: 'K8sDiagnostic',
          component: () => import('@/views/k8s/diagnostic/index.vue'),
          meta: {
            title: '诊断中心',
            requiresAuth: true,
            permission: 'k8s:diagnostic:execute'
          }
        },
        // 工具子菜单
        {
          path: 'logs',
          name: 'K8sLogs',
          component: () => import('@/views/k8s/tools/logs.vue'),
          meta: {
            title: '日志查询',
            requiresAuth: true,
            permission: 'k8s:tools:logs'
          }
        },
        {
          path: 'events',
          name: 'K8sEvents',
          component: () => import('@/views/k8s/tools/events.vue'),
          meta: {
            title: '事件查询',
            requiresAuth: true,
            permission: 'k8s:tools:events'
          }
        },
        {
          path: 'yaml',
          name: 'K8sYaml',
          component: () => import('@/views/k8s/tools/yaml.vue'),
          meta: {
            title: 'YAML编辑器',
            requiresAuth: true,
            permission: 'k8s:tools:yaml'
          }
        },
        {
          path: 'settings',
          name: 'K8sSettings',
          component: () => import('@/views/k8s/settings/index.vue'),
          meta: {
            title: 'K8s设置',
            requiresAuth: true,
            permission: 'k8s:settings:view'
          }
        }
      ]
    },
    // CMDB
    {
      path: '/cmdb',
      name: 'CMDB',
      component: () => import('@/views/cmdb/layout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: 'servers',
          name: 'CMDBServers',
          component: () => import('@/views/cmdb/servers/index.vue'),
          meta: {
            title: '服务器管理',
            requiresAuth: true,
            permission: 'cmdb:servers:view'
          }
        }
      ]
    },
    // 404页面
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/404.vue'),
      meta: { requiresAuth: false }
    }
  ]
});

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();
  const permissionStore = usePermissionStore();

  // 检查是否需要认证
  if (to.meta.requiresAuth !== false && !userStore.isLoggedIn) {
    next({
      name: 'Login',
      query: { redirect: to.fullPath }
    });
    return;
  }

  // 检查权限
  if (to.meta.permission) {
    const hasPermission = await permissionStore.hasPermission(to.meta.permission);
    if (!hasPermission) {
      // 显示权限不足提示
      next({ name: 'Home' });
      return;
    }
  }

  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - OneOps`;
  }

  next();
});

export default router;
