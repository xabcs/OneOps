/**
 * K8s 数据格式化工具函数
 * 用于统一格式化 Kubernetes 对象的显示数据
 */

/**
 * 格式化标签对象为字符串
 * 输入: {app: "nginx", env: "prod"}
 * 输出: "app:nginx, env:prod"
 */
export function formatLabels(labels: Record<string, string> | undefined): string {
  if (!labels || Object.keys(labels).length === 0) return '-';
  return Object.entries(labels)
    .map(([key, value]) => `${key}:${value}`)
    .join(', ');
}

/**
 * 格式化状态条件数组为字符串
 * 输入: [{type: "Available", status: "True"}, {type: "Progressing", status: "True"}]
 * 输出: "Available:True, Progressing:True"
 */
export function formatConditions(conditions: Array<{type: string, status: string}> | undefined): string {
  if (!conditions || conditions.length === 0) return '-';
  return conditions
    .map(c => `${c.type}:${c.status}`)
    .join(', ');
}

/**
 * 从 Pod 对象提取镜像列表（换行分隔）
 */
export function formatImages(pod: any): string {
  if (!pod?.containers || pod.containers.length === 0) return '-';
  return pod.containers
    .map((c: any) => c.image)
    .join('\n');
}

/**
 * 格式化副本数显示
 */
export function formatReplicas(ready: number | undefined, total: number | undefined): string {
  const r = ready ?? 0;
  const t = total ?? 0;
  return `${r}/${t}`;
}
