type CachedRequest<T> = {
  load: (force?: boolean) => Promise<T>;
  invalidate: () => void;
};

/**
 * 创建带 TTL 和请求去重的基础数据加载器，避免同一页面重复拉取不常变化的数据。
 */
export function createCachedRequest<T>(loader: () => Promise<T>, ttl = 60_000): CachedRequest<T> {
  let pending: Promise<T> | null = null;
  let value: T | null = null;
  let expiresAt = 0;

  function load(force = false) {
    const now = Date.now();

    if (!force && value !== null && now < expiresAt) {
      return Promise.resolve(value);
    }

    if (!force && pending && now < expiresAt) {
      return pending;
    }

    pending = loader().then(result => {
      value = result;
      expiresAt = Date.now() + ttl;
      return result;
    });

    pending.finally(() => {
      pending = null;
    });

    return pending;
  }

  function invalidate() {
    pending = null;
    value = null;
    expiresAt = 0;
  }

  return { load, invalidate };
}
