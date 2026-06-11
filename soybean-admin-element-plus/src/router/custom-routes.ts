/**
 * 自定义路由配置
 * 这些路由不会被Elegant Router自动生成或覆盖
 */
import type { RouteRecordRaw } from 'vue-router';

import BaseLayout from '@/layouts/base-layout/index.vue';

// 自定义视图组件导入
const CMDBServerDetail = () => import('@/views/cmdb/server/detail.vue');
const MonitoringServersDetail = () => import('@/views/monitoring/servers/detail/index.vue');

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
