// frontend/src/router/index.ts
// 前端路由配置 - 添加诊断页面

import { createRouter, createWebHistory } from 'vue-router';

const routes = [
  // ... 现有路由配置 ...

  // K8s集群管理路由
  {
    path: '/k8s',
    component: () => import('@/views/k8s/layout.vue'),
    children: [
      {
        path: 'clusters',
        name: 'K8sClusters',
        component: () => import('@/views/k8s/clusters/index.vue'),
        meta: {
          title: 'K8s集群管理',
          requiresAuth: true,
          permission: 'k8s.cluster.view'
        }
      },
      {
        path: 'pods',
        name: 'K8sPods',
        component: () => import('@/views/k8s/pods/index.vue'),
        meta: {
          title: 'Pod管理',
          requiresAuth: true,
          permission: 'k8s.pod.view'
        }
      },
      {
        path: 'diagnostic',
        name: 'K8sDiagnostic',
        component: () => import('@/views/k8s/diagnostic/index.vue'),
        meta: {
          title: 'K8s诊断中心',
          requiresAuth: true,
          permission: 'k8s.diagnostic.execute'
        }
      }
    ]
  }
  // ... 其他路由 ...
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

export default router;
