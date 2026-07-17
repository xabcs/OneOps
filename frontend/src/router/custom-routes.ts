/**
 * 自定义路由配置
 * 这些路由不会被Elegant Router自动生成或覆盖
 */
import type { RouteRecordRaw } from 'vue-router';
import BaseLayout from '@/layouts/base-layout/index.vue';
import TerminalLayout from '@/layouts/terminal-layout/index.vue';

// 自定义视图组件导入
const CMDBServerDetail = () => import('@/views/cmdb/server/detail.vue');
const MonitoringServersDetail = () => import('@/views/monitoring/servers/detail/index.vue');
const K8sDeploymentDetail = () => import('@/views/k8s/resources/deployments/detail.vue');
const K8sStatefulSetDetail = () => import('@/views/k8s/resources/statefulsets/detail.vue');
const K8sDaemonSetDetail = () => import('@/views/k8s/resources/daemonsets/detail.vue');
const K8sPodDetail = () => import('@/views/k8s/resources/pods/detail.vue');
const K8sJobDetail = () => import('@/views/k8s/resources/jobs/detail.vue');
const K8sCronJobDetail = () => import('@/views/k8s/resources/cronjobs/detail.vue');
const K8sConfigMapDetail = () => import('@/views/k8s/resources/configmaps/detail.vue');
const K8sSecretDetail = () => import('@/views/k8s/resources/secrets/detail.vue');
const K8sServiceDetail = () => import('@/views/k8s/resources/services/detail.vue');
const K8sIngressDetail = () => import('@/views/k8s/resources/ingresses/detail.vue');
const K8sWorkloads = () => import('@/views/k8s/workloads/index.vue');
const K8sNetwork = () => import('@/views/k8s/network/index.vue');
const K8sConfig = () => import('@/views/k8s/config/index.vue');
const K8sTerminal = () => import('@/views/k8s/terminal/index.vue');

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
    name: 'k8s_statefulset_detail',
    path: '/k8s/resources/statefulsets/detail',
    component: BaseLayout,
    meta: {
      title: 'StatefulSet详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_workloads'
    },
    children: [
      {
        name: 'k8s_statefulset_detail_view',
        path: '',
        component: K8sStatefulSetDetail
      }
    ]
  },
  {
    name: 'k8s_daemonset_detail',
    path: '/k8s/resources/daemonsets/detail',
    component: BaseLayout,
    meta: {
      title: 'DaemonSet详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_workloads'
    },
    children: [
      {
        name: 'k8s_daemonset_detail_view',
        path: '',
        component: K8sDaemonSetDetail
      }
    ]
  },
  {
    name: 'k8s_pod_detail',
    path: '/k8s/resources/pods/detail',
    component: BaseLayout,
    meta: {
      title: 'Pod详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_workloads'
    },
    children: [
      {
        name: 'k8s_pod_detail_view',
        path: '',
        component: K8sPodDetail
      }
    ]
  },
  {
    name: 'k8s_job_detail',
    path: '/k8s/resources/jobs/detail',
    component: BaseLayout,
    meta: {
      title: 'Job详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_workloads'
    },
    children: [
      {
        name: 'k8s_job_detail_view',
        path: '',
        component: K8sJobDetail
      }
    ]
  },
  {
    name: 'k8s_cronjob_detail',
    path: '/k8s/resources/cronjobs/detail',
    component: BaseLayout,
    meta: {
      title: 'CronJob详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_workloads'
    },
    children: [
      {
        name: 'k8s_cronjob_detail_view',
        path: '',
        component: K8sCronJobDetail
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
  // ConfigMap 详情页
  {
    name: 'k8s_configmap_detail',
    path: '/k8s/resources/configmaps/detail',
    component: BaseLayout,
    meta: {
      title: 'ConfigMap详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_resources_configmaps'
    },
    children: [
      {
        name: 'k8s_configmap_detail_view',
        path: '',
        component: K8sConfigMapDetail
      }
    ]
  },

  // Secret 详情页
  {
    name: 'k8s_secret_detail',
    path: '/k8s/resources/secrets/detail',
    component: BaseLayout,
    meta: {
      title: 'Secret详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_resources_secrets'
    },
    children: [
      {
        name: 'k8s_secret_detail_view',
        path: '',
        component: K8sSecretDetail
      }
    ]
  },

  // Service 详情页
  {
    name: 'k8s_service_detail',
    path: '/k8s/resources/services/detail',
    component: BaseLayout,
    meta: {
      title: 'Service详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_network'
    },
    children: [
      {
        name: 'k8s_service_detail_view',
        path: '',
        component: K8sServiceDetail
      }
    ]
  },

  // Ingress 详情页
  {
    name: 'k8s_ingress_detail',
    path: '/k8s/resources/ingresses/detail',
    component: BaseLayout,
    meta: {
      title: 'Ingress详情',
      i18nKey: null,
      hideInMenu: true,
      activeMenu: 'k8s_network'
    },
    children: [
      {
        name: 'k8s_ingress_detail_view',
        path: '',
        component: K8sIngressDetail
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
  },
  // K8s 终端页面 - 使用 TerminalLayout，无菜单栏
  {
    name: 'k8s_terminal',
    path: '/k8s/terminal',
    component: TerminalLayout,
    meta: {
      title: 'Pod终端',
      i18nKey: null,
      hideInMenu: true,
      constant: true // 标记为常量路由，无需认证
    },
    children: [
      {
        name: 'k8s_terminal_view',
        path: '',
        component: K8sTerminal
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
