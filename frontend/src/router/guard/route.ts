import { nextTick } from 'vue';
import type { RouteLocationNormalized, RouteLocationRaw, Router } from 'vue-router';
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
    console.log('🔍 [路由守卫] 开始处理路由跳转');
    console.log('📍 目标路由:', to.path, 'name:', to.name);
    console.log('📍 来源路由:', from.path, 'name:', from.name);
    console.log('🔍 目标路由meta:', JSON.stringify(to.meta, null, 2));

    const location = await initRoute(to, router);

    if (location) {
      console.log('➡️ [路由守卫] initRoute返回重定向:', location);
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
    const hasAuth = authStore.isStaticSuper || (!routeRoles.length && !routePermissions.length) || hasRole || hasPermission;

    console.log('🔐 [路由守卫] 认证状态检查:');
    console.log('  - isLogin:', isLogin);
    console.log('  - needLogin:', needLogin);
    console.log('  - 用户角色名:', authStore.userInfo.roleNames);
    console.log('  - 路由要求角色:', routeRoles);
    console.log('  - 路由要求权限:', routePermissions);
    console.log('  - hasRole:', hasRole);
    console.log('  - hasPermission:', hasPermission);
    console.log('  - hasAuth:', hasAuth);
    console.log('  - 用户权限码:', authStore.permissions);

    // if it is login route when logged in, then switch to the root page
    if (to.name === loginRoute && isLogin) {
      console.log('➡️ [路由守卫] 已登录访问登录页，重定向到root');
      return { name: rootRoute };
    }

    // if the route does not need login, then it is allowed to access directly
    if (!needLogin) {
      console.log('✅ [路由守卫] 路由不需要登录，直接访问');
      return handleRouteSwitch(to, from);
    }

    // the route need login but the user is not logged in, then switch to the login page
    if (!isLogin) {
      console.log('➡️ [路由守卫] 需要登录但未登录，重定向到登录页');
      return { name: loginRoute, query: { redirect: to.fullPath } };
    }

    // if the user is logged in but does not have authorization, then switch to the 403 page
    if (!hasAuth) {
      console.log('➡️ [路由守卫] 无权限访问，重定向到403页面');
      return { name: noAuthorizationRoute };
    }

    // switch route normally
    console.log('✅ [路由守卫] 路由检查通过，正常处理');
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
  console.log('🚀 [initRoute] 开始初始化路由');
  console.log('  - 目标路径:', to.path);
  console.log('  - 路由名称:', to.name);
  console.log('  - 是否not-found:', isNotFoundRoute);
  console.log('  - isInitConstantRoute:', routeStore.isInitConstantRoute);
  console.log('  - isInitAuthRoute:', routeStore.isInitAuthRoute);

  // 首次初始化常量路由
  if (!routeStore.isInitConstantRoute) {
    await routeStore.initConstantRoute();

    // 只有当路由被not-found捕获时才重定向
    if (isNotFoundRoute) {
      console.log('➡️ [initRoute] not-found路由，重新加载当前路径');
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
      console.log('✅ [initRoute] 未登录但访问常量路由，允许访问');
      return undefined;
    }

    console.log('➡️ [initRoute] 未登录访问需要登录的路由，跳转到登录页');
    return { name: 'login', query: { redirect: to.fullPath } };
  }

  // 已登录状态：首次初始化认证路由
  if (!routeStore.isInitAuthRoute) {
    console.log('🔄 [initRoute] 开始初始化认证路由...');
    await routeStore.initAuthRoute();

    console.log('  - initAuthRoute完成, isInitAuthRoute:', routeStore.isInitAuthRoute);

    // 如果被not-found捕获，检查是否是刚刚初始化的动态路由
    if (isNotFoundRoute) {
      console.log('⚠️ [initRoute] 检测到not-found路由:', to.path);

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
        console.log('✅ [initRoute] 动态路由已注册，重新加载页面');
        return {
          path: to.fullPath,
          replace: true,
          query: to.query,
          hash: to.hash
        };
      }

      console.log('❌ [initRoute] 路由确实不存在，重定向到首页');
      return {
        path: '/',
        replace: true
      };
    }
    console.log('✅ [initRoute] not-found路由处理完成，继续正常流程');
  }

  routeStore.onRouteSwitchWhenLoggedIn();

  // 正常路由直接访问
  if (!isNotFoundRoute) {
    console.log('✅ [initRoute] 正常路由，允许访问');
    return undefined;
  }

  // not-found路由：检查是否存在权限路由（有权限但无访问权限）
  console.log('🔍 [initRoute] 检查not-found路由是否存在权限路由...');
  const exist = await routeStore.getIsAuthRouteExist(to.path as RoutePath);
  console.log('  - 路径存在检查结果:', exist);

  if (exist) {
    console.log('➡️ [initRoute] 路由存在但无权限，重定向到403');
    return { name: '403' };
  }

  console.log('✅ [initRoute] 路由不存在，允许访问（可能是前端自动生成的路由）');
  return undefined;
}

function handleRouteSwitch(to: RouteLocationNormalized, from: RouteLocationNormalized) {
  // route with href
  if (to.meta.href) {
    // 如果 href 等于当前路径，说明是直接访问（比如通过新标签页），直接加载路由
    if (to.meta.href === to.path) {
      console.log('🔗 [路由守卫] href等于当前路径，直接在当前窗口加载');
      return undefined; // 直接在当前窗口加载
    }

    console.log('🔗 [路由守卫] 打开新窗口:', to.meta.href);
    window.open(to.meta.href, '_blank');

    return { path: from.fullPath, replace: true, query: from.query, hash: to.hash };
  }
  return undefined;
}
