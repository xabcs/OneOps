/**
 * 自定义路由配置
 * 这些路由不会被Elegant Router自动生成或覆盖
 */
import type { RouteRecordRaw } from 'vue-router';
import BaseLayout from '@/layouts/base-layout/index.vue';

// 自定义视图组件导入
const CMDBServerDetail = () => import('@/views/cmdb/server/detail.vue');
const MonitoringServersDetail = () => import('@/views/monitoring/servers/detail/index.vue');
const K8sDeploymentDetail = () => import('@/views/k8s/resources/deployments/detail.vue');
const K8sWorkloads = () => import('@/views/k8s/workloads/index.vue');
const K8sNetwork = () => import('@/views/k8s/network/index.vue');
const K8sConfig = () => import('@/views/k8s/config/index.vue');

/**
 * 自定义路由列表
 * 这些路由会被动态添加到路由器中
 */
export const customRoutes: RouteRecordRaw[] = [
  {
    name: 'cmdb_server_detail',
    path: '/cmdb/server/detail',
    component: BaseLayout,
    meta: {
      title: 'cmdb_server_detail',
      i18nKey: 'route.cmdb_server_detail',
      hideInMenu: true,
      activeMenu: 'cmdb_servers'
    },
    children: [
      {
        name: 'cmdb_server_detail_view',
        path: '',
        component: CMDBServerDetail
      }
    ]
  },
  {
    name: 'monitoring_servers_detail',
    path: '/monitoring/servers/detail',
    component: BaseLayout,
    meta: {
      title: 'monitoring_servers_detail',
      i18nKey: 'route.monitoring_servers_detail',
      hideInMenu: true,
      activeMenu: 'monitoring_servers'
    },
    children: [
      {
        name: 'monitoring_servers_detail_view',
        path: '',
        component: MonitoringServersDetail
      }
    ]
  },
  {
    name: 'k8s_deployment_detail',
    path: '/k8s/resources/deployments/detail',
    component: BaseLayout,
    meta: {
      title: 'Deployment详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_workloads'
    },
    children: [
      {
        name: 'k8s_deployment_detail_view',
        path: '',
        component: K8sDeploymentDetail
      }
    ]
  },
  {
    name: 'k8s_workloads',
    path: '/k8s/workloads',
    component: BaseLayout,
    meta: {
      title: '工作负载',
      i18nKey: null,
      hideInMenu: false,
      activeMenu: 'k8s_workloads'
    },
    children: [
      {
        name: 'k8s_workloads_view',
        path: '',
        component: K8sWorkloads
      }
    ]
  },
  {
    name: 'k8s_network',
    path: '/k8s/network',
    component: BaseLayout,
    meta: {
      title: '网络',
      i18nKey: null,
      hideInMenu: false,
      activeMenu: 'k8s_network'
    },
    children: [
      {
        name: 'k8s_network_view',
        path: '',
        component: K8sNetwork
      }
    ]
  },
  {
    name: 'k8s_config',
    path: '/k8s/config',
    component: BaseLayout,
    meta: {
      title: '配置管理',
      i18nKey: null,
      hideInMenu: false,
      activeMenu: 'k8s_config'
    },
    children: [
      {
        name: 'k8s_config_view',
        path: '',
        component: K8sConfig
      }
    ]
  }
];

/**
 * 将自定义路由添加到路由器
 */
export function addCustomRoutes(router: any) {
  customRoutes.forEach(route => {
    router.addRoute(route);
  });
}
