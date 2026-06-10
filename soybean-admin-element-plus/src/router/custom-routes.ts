/**
 * 自定义路由配置
 * 这些路由不会被Elegant Router自动生成或覆盖
 */
import type { RouteRecordRaw } from 'vue-router';

// 自定义视图组件导入
const CMDBServerDetail = () => import('@/views/cmdb/server/detail.vue');

/**
 * 自定义路由列表
 * 这些路由会被动态添加到路由器中
 */
export const customRoutes: RouteRecordRaw[] = [
  {
    name: 'cmdb_server_detail',
    path: '/cmdb/server/detail',
    component: CMDBServerDetail,
    meta: {
      title: 'cmdb_server_detail',
      i18nKey: 'route.cmdb_server_detail',
      hideInMenu: true,
      activeMenu: 'cmdb_servers'
    }
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
