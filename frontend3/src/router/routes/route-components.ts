import { defineAsyncComponent, defineComponent, h } from 'vue';
import type { RouteComponent, RouteRecordRaw } from 'vue-router';

type AsyncRouteComponent = () => Promise<RouteComponent>;

const namedRouteComponents = new Map<string, RouteComponent>();

function isAsyncRouteComponent(component: unknown): component is AsyncRouteComponent {
  // generated route views are zero-argument dynamic import loaders
  return typeof component === 'function' && component.length === 0;
}

/**
 * KeepAlive 的 include 匹配组件名，路由缓存名单使用路由名。
 * 统一包装可避免每个页面手动维护 defineOptions name。
 */
function createNamedRouteComponent(routeName: string, loader: AsyncRouteComponent) {
  const cachedComponent = namedRouteComponents.get(routeName);
  if (cachedComponent) return cachedComponent;

  const asyncComponent = defineAsyncComponent(loader);
  const namedComponent = defineComponent({
    name: routeName,
    inheritAttrs: false,
    setup(_, { attrs }) {
      return () => h(asyncComponent, attrs);
    }
  });

  namedRouteComponents.set(routeName, namedComponent);
  return namedComponent;
}

export function nameRouteComponents(routes: RouteRecordRaw[]) {
  routes.forEach(route => {
    if (route.name && isAsyncRouteComponent(route.component)) {
      route.component = createNamedRouteComponent(route.name as string, route.component);
    }

    if (route.children?.length) {
      nameRouteComponents(route.children);
    }
  });
}
