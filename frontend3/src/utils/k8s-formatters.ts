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
 * 注解分类
 */
export interface AnnotationItem {
  key: string;
  value: string;
  category: 'user' | 'system';
  isLong: boolean;
}

/**
 * 分类和格式化注解
 * - 系统注解：以 kubernetes.io/ 或 k8s.io/ 开头
 * - 用户注解：其他所有注解
 * - 长内容：value 长度超过 100 字符
 */
export function formatAnnotations(annotations: Record<string, string> | undefined): {
  user: AnnotationItem[];
  system: AnnotationItem[];
  total: number;
} {
  if (!annotations || Object.keys(annotations).length === 0) {
    return { user: [], system: [], total: 0 };
  }

  const result: { user: AnnotationItem[]; system: AnnotationItem[]; total: number } = {
    user: [],
    system: [],
    total: 0
  };

  Object.entries(annotations).forEach(([key, value]) => {
    const item: AnnotationItem = {
      key,
      value: value || '',
      category: key.startsWith('kubernetes.io/') || key.startsWith('k8s.io/') ? 'system' : 'user',
      isLong: (value || '').length > 100
    };

    if (item.category === 'user') {
      result.user.push(item);
    } else {
      result.system.push(item);
    }
    result.total++;
  });

  return result;
}

/**
 * 截断注解显示值（用于基本信息展示）
 */
export function truncateAnnotation(value: string, maxLength: number = 50): string {
  if (!value || value.length <= maxLength) return value;
  return `${value.substring(0, maxLength)}...`;
}

/**
 * 获取简短的注解摘要（用于基本信息预览）
 */
export function getAnnotationSummary(annotations: Record<string, string> | undefined): string {
  const formatted = formatAnnotations(annotations);
  if (formatted.total === 0) return '-';

  const parts: string[] = [];
  if (formatted.user.length > 0) {
    parts.push(`用户: ${formatted.user.length}个`);
  }
  if (formatted.system.length > 0) {
    parts.push(`系统: ${formatted.system.length}个`);
  }
  return parts.join(', ');
}

/**
 * 格式化滚动更新策略
 */
export function formatStrategy(strategy: { type?: string; maxUnavailable?: number; maxSurge?: number; partition?: number } | null | undefined): string {
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
export function formatImages(pod: { containers: { image: string }[] }): string {
  if (!pod?.containers || pod.containers.length === 0) return '-';
  return pod.containers.map(c => c.image).join('\n');
}

/**
 * 格式化副本数显示
 */
export function formatReplicas(ready: number | undefined, total: number | undefined): string {
  const r = ready ?? 0;
  const t = total ?? 0;
  return `${r}/${t}`;
}
