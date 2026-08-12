import { nextTick } from 'vue';
import type { LocationQueryRaw, RouteLocationNormalized, RouteLocationRaw, Router } from 'vue-router';
import type { RouteKey, RoutePath } from '@elegant-router/types';
import { useAuthStore } from '@/store/modules/auth';
import { useRouteStore } from '@/store/modules/route';
import { localStg } from '@/utils/storage';

/**
 * create route guard
 *
 * @param router router instance
 */
export function createRouteGuard(router: Router) {
  router.beforeEach(async (to, from) => {
    // 🐛 调试信息：路由跳转开始

    const location = await initRoute(to, router);

    if (location) {
      return location;
    }

    const authStore = useAuthStore();

    const rootRoute: RouteKey = 'root';
    const loginRoute: RouteKey = 'login';
    const noAuthorizationRoute: RouteKey = '403';

    const isLogin = Boolean(localStg.get('token'));
    const needLogin = !to.meta.constant;
    const routeRoles = to.meta.roles || [];
    const routePermissions = to.meta.permissions || [];

    // 🔥 修复：同时支持角色名和权限码检查
    const hasRole = routeRoles.length && authStore.userInfo.roleNames.some(role => routeRoles.includes(role));
    const hasPermission = routePermissions.length && routePermissions.some(perm => authStore.hasPermission(perm));
    const hasAuth =
      authStore.isStaticSuper || (!routeRoles.length && !routePermissions.length) || hasRole || hasPermission;

    // if it is login route when logged in, then switch to the root page
    if (to.name === loginRoute && isLogin) {
      return { name: rootRoute };
    }

    // if the route does not need login, then it is allowed to access directly
    if (!needLogin) {
      return handleRouteSwitch(to, from);
    }

    // the route need login but the user is not logged in, then switch to the login page
    if (!isLogin) {
      return { name: loginRoute, query: { redirect: to.fullPath } };
    }

    // if the user is logged in but does not have authorization, then switch to the 403 page
    if (!hasAuth) {
      return { name: noAuthorizationRoute };
    }

    // switch route normally
    return handleRouteSwitch(to, from);
  });
}

/**
 * initialize route
 *
 * @param to to route
 * @param router router instance
 */
async function initRoute(to: RouteLocationNormalized, router: Router): Promise<RouteLocationRaw | undefined> {
  const routeStore = useRouteStore();

  const notFoundRoute: RouteKey = 'not-found';
  const isNotFoundRoute = to.name === notFoundRoute;

  // 🐛 调试信息：initRoute开始

  // 首次初始化常量路由
  if (!routeStore.isInitConstantRoute) {
    await routeStore.initConstantRoute();

    // 只有当路由被not-found捕获时才重定向
    if (isNotFoundRoute) {
      return {
        path: to.fullPath,
        replace: true,
        query: to.query,
        hash: to.hash
      };
    }
    return undefined;
  }

  const isLogin = Boolean(localStg.get('token'));

  if (!isLogin) {
    // 未登录状态：常量路由直接访问，其他路由跳转到登录页
    if (to.meta.constant && !isNotFoundRoute) {
      return undefined;
    }

    return { name: 'login', query: { redirect: to.fullPath } };
  }

  // 已登录状态：首次初始化认证路由
  if (!routeStore.isInitAuthRoute) {
    await routeStore.initAuthRoute();

    // 如果被not-found捕获，检查是否是刚刚初始化的动态路由
    if (isNotFoundRoute) {
      // 等待一个 tick，确保动态路由已完全注册到 Vue Router
      await nextTick();

      // 重新检查目标路由是否已被注册
      const allRoutes = router.getRoutes();
      const matchedRoute = allRoutes.find(r => {
        // 检查路径匹配（支持带参数的路径）
        if (r.path === to.path) return true;

        // 检查路由名称匹配（对于嵌套路由）
        if (r.name && to.name && r.name === to.name) return true;

        // 检查父路由路径匹配（对于嵌套子路由）
        if (r.path && to.path.startsWith(r.path)) return true;

        return false;
      });

      if (matchedRoute) {
        return {
          path: to.fullPath,
          replace: true,
          query: to.query,
          hash: to.hash
        };
      }

      return {
        path: '/',
        replace: true
      };
    }
  }

  routeStore.onRouteSwitchWhenLoggedIn();

  // 正常路由直接访问
  if (!isNotFoundRoute) {
    return undefined;
  }

  // not-found路由：检查是否存在权限路由（有权限但无访问权限）
  const exist = await routeStore.getIsAuthRouteExist(to.path as RoutePath);

  if (exist) {
    return { name: '403' };
  }

  return undefined;
}

function handleRouteSwitch(to: RouteLocationNormalized, from: RouteLocationNormalized) {
  // route with href
  if (to.meta.href) {
    // 如果 href 等于当前路径，说明是直接访问（比如通过新标签页），直接加载路由
    if (to.meta.href === to.path) {
      return undefined; // 直接在当前窗口加载
    }

    window.open(to.meta.href, '_blank');

    return { path: from.fullPath, replace: true, query: from.query, hash: to.hash };
  }
  return undefined;
}
