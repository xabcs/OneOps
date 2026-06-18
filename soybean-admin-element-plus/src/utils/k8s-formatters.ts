/**
 * K8s 数据格式化工具函数
 * 用于统一格式化 Kubernetes 对象的显示数据
 */

/**
 * 格式化标签对象为 Tag 数组
 * 输入: {app: "nginx", env: "prod"}
 * 输出: [{key: "app", value: "nginx"}, {key: "env", value: "prod"}]
 */
export function formatLabels(labels: Record<string, string> | undefined): Array<{ key: string; value: string }> {
  if (!labels || Object.keys(labels).length === 0) return [];
  return Object.entries(labels).map(([key, value]) => ({ key, value }));
}

/**
 * 格式化选择器对象为字符串
 * 输入: {app: "nginx"}
 * 输出: "app=nginx"
 */
export function formatSelectors(selectors: Record<string, string> | undefined): string {
  if (!selectors || Object.keys(selectors).length === 0) return '-';
  return Object.entries(selectors)
    .map(([key, value]) => `${key}=${value}`)
    .join(', ');
}

/**
 * 格式化注解对象为字符串（只显示常用字段）
 */
export function formatAnnotations(annotations: Record<string, string> | undefined): string {
  if (!annotations || Object.keys(annotations).length === 0) return '-';
  // 过滤掉 kubernetes 系统注解，只显示用户添加的
  const userAnnotations = Object.entries(annotations).filter(
    ([key]) => !key.startsWith('kubernetes.io/') && !key.startsWith('k8s.io/')
  );
  if (userAnnotations.length === 0) return '-';
  return userAnnotations
    .map(([key, value]) => `${key}: ${value}`)
    .join(', ');
}

/**
 * 格式化滚动更新策略
 */
export function formatStrategy(strategy: any): string {
  if (!strategy) return '-';
  const type = strategy.type || 'RollingUpdate';
  const params: string[] = [];
  if (strategy.maxUnavailable !== undefined) params.push(`不可用: ${strategy.maxUnavailable}`);
  if (strategy.maxSurge !== undefined) params.push(`超出: ${strategy.maxSurge}`);
  if (strategy.partition !== undefined) params.push(`分区: ${strategy.partition}`);
  return params.length > 0 ? `${type} (${params.join(', ')})` : type;
}

/**
 * 格式化状态条件数组为条件数组（包含 reason）
 * 输入: [{type: "Available", status: "True", reason: "MinimumReplicasAvailable"}, {type: "Progressing", status: "True"}]
 * 输出: [{type: "Available", status: "True", reason: "MinimumReplicasAvailable"}, ...]
 */
export function formatConditions(
  conditions: Array<{ type: string; status: string; reason?: string; message?: string }> | undefined
): Array<{ type: string; status: string; reason?: string; message?: string }> {
  if (!conditions || conditions.length === 0) return [];
  return conditions;
}

/**
 * 从 Pod 对象提取镜像列表（换行分隔）
 */
export function formatImages(pod: any): string {
  if (!pod?.containers || pod.containers.length === 0) return '-';
  return pod.containers.map((c: any) => c.image).join('\n');
}

/**
 * 格式化副本数显示
 */
export function formatReplicas(ready: number | undefined, total: number | undefined): string {
  const r = ready ?? 0;
  const t = total ?? 0;
  return `${r}/${t}`;
}
