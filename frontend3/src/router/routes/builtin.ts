import type { CustomRoute } from '@elegant-router/types';
import { getRoutePath, transformElegantRoutesToVueRoutes } from '../elegant/transform';
import { layouts, views } from '../elegant/imports';

export const ROOT_ROUTE: CustomRoute = {
  name: 'root',
  path: '/',
  redirect: getRoutePath(import.meta.env.VITE_ROUTE_HOME) || '/home',
  meta: {
    title: 'root',
    constant: true
  }
};

const NOT_FOUND_ROUTE: CustomRoute = {
  name: 'not-found',
  path: '/:pathMatch(.*)*',
  component: 'layout.blank$view.404',
  meta: {
    title: 'not-found',
    constant: true
  }
};

/** web终端旧路由重定向：/terminal/workbench → /webterminal */
const WEBTERMINAL_REDIRECT: CustomRoute = {
  name: 'webterminal_redirect',
  path: '/terminal/workbench',
  redirect: '/webterminal',
  meta: {
    title: 'webterminal_redirect',
    constant: true
  }
};

/** builtin routes, it must be constant and setup in vue-router */
export const builtinRoutes: CustomRoute[] = [ROOT_ROUTE, NOT_FOUND_ROUTE, WEBTERMINAL_REDIRECT];

/** create builtin vue routes */
export function createBuiltinVueRoutes() {
  return transformElegantRoutesToVueRoutes(builtinRoutes, layouts, views);
}
