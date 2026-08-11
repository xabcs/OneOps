/**
 * 性能优化工具集合
 *
 * 提供防抖、节流、缓存等性能优化相关的工具函数
 */

import { type ComputedRef, type Ref, ref, watch } from 'vue';

/**
 * 防抖 Hook
 * @param value 要防抖的值
 * @param delay 延迟时间（毫秒）
 * @returns 防抖后的值
 *
 * @example
 * const searchValue = useDebounce(ref(''), 300)
 */
export function useDebounce<T>(value: Ref<T> | ComputedRef<T>, delay: number = 300): Ref<T> {
  const debouncedValue = ref<T>(value.value) as Ref<T>;
  let timeout: ReturnType<typeof setTimeout> | null = null;

  watch(
    value,
    newValue => {
      if (timeout !== null) {
        clearTimeout(timeout);
      }

      timeout = setTimeout(() => {
        debouncedValue.value = newValue;
      }, delay);
    },
    { immediate: true }
  );

  return debouncedValue;
}

/**
 * 节流 Hook
 * @param fn 要节流的函数
 * @param delay 节流间隔（毫秒）
 * @returns 节流后的函数
 *
 * @example
 * const throttledScroll = useThrottle(() => console.log('scroll'), 200)
 */
export function useThrottle<T extends (...args: unknown[]) => unknown>(fn: T, delay: number = 200): T {
  let lastCall = 0;
  let timeout: ReturnType<typeof setTimeout> | null = null;

  return ((...args: Parameters<T>) => {
    const now = Date.now();
    const remaining = delay - (now - lastCall);

    if (remaining <= 0) {
      if (timeout !== null) {
        clearTimeout(timeout);
        timeout = null;
      }
      lastCall = now;
      fn(...args);
    } else if (timeout === null) {
      timeout = setTimeout(() => {
        lastCall = Date.now();
        timeout = null;
        fn(...args);
      }, remaining);
    }
  }) as T;
}

/**
 * 内存缓存 Hook
 * @param fetchFn 数据获取函数
 * @param key 缓存键
 * @param ttl 缓存时间（毫秒）
 * @returns 缓存数据和获取函数
 *
 * @example
 * const { data, fetch } = useMemoryCache(
 *   async (id) => await fetchUser(id),
 *   'user',
 *   5000
 * )
 */
export function useMemoryCache<T, K extends unknown[]>(
  fetchFn: (...args: K) => Promise<T>,
  key: string,
  ttl: number = 5000
) {
  const cache = new Map<string, { data: T; expires: number }>();
  const data = ref<T | null>(null);
  const loading = ref(false);

  const generateCacheKey = (...args: K) => {
    return `${key}:${args.join(':')}`;
  };

  const fetch = async (...args: K) => {
    const cacheKey = generateCacheKey(...args);
    const cached = cache.get(cacheKey);

    // 检查缓存是否有效
    if (cached && Date.now() < cached.expires) {
      data.value = cached.data;
      return cached.data;
    }

    // 获取新数据
    loading.value = true;
    try {
      const result = await fetchFn(...args);
      cache.set(cacheKey, {
        data: result,
        expires: Date.now() + ttl
      });
      data.value = result;
      return result;
    } finally {
      loading.value = false;
    }
  };

  const clearCache = () => {
    cache.clear();
  };

  return {
    data,
    loading,
    fetch,
    clearCache
  };
}

/**
 * 懒加载 Hook
 * @param factory 数据创建函数
 * @returns 懒加载数据和加载函数
 *
 * @example
 * const { data, load } = useLazyLoad(async () => {
 *   return await fetchExpensiveData()
 * })
 */
export function useLazyLoad<T>(factory: () => Promise<T>) {
  const data = ref<T | null>(null);
  const loading = ref(false);
  const loaded = ref(false);

  const load = async () => {
    if (loaded.value || loading.value) {
      return;
    }

    loading.value = true;
    try {
      data.value = await factory();
      loaded.value = true;
    } finally {
      loading.value = false;
    }
  };

  return {
    data,
    loading,
    loaded,
    load
  };
}
