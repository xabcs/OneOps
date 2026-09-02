/**
 * 诊断中心共享工具：风险分级元数据、渠道、耗时格式化
 */

/** 风险级元数据（与后端 L0-L5 / disabled 对齐） */
export const RISK_LEVEL_META: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'danger'; desc: string }> = {
  L0: { label: 'L0 环境操作', type: 'info', desc: '会话/环境类操作，无 JVM 影响' },
  L1: { label: 'L1 只读诊断', type: 'success', desc: '只读查看，不改变任何状态' },
  L2: { label: 'L2 观测增强', type: 'success', desc: '字节码增强类观测，默认带 -n 限量' },
  L3: { label: 'L3 资源敏感', type: 'warning', desc: '可能产生大对象/CPU 开销（如 heapdump）' },
  L4: { label: 'L4 变更操作', type: 'warning', desc: '修改运行时行为（如 logger 级别、sysprop）' },
  L5: { label: 'L5 任意执行', type: 'danger', desc: '可执行任意代码/表达式（ognl、vmtool、redefine）' },
  disabled: { label: '已禁用', type: 'danger', desc: '管理员已拉黑该命令' }
};

/** 高风险级（执行前需强确认） */
export const HIGH_RISK_LEVELS = ['L3', 'L4', 'L5'];

export function riskMeta(level: string) {
  return RISK_LEVEL_META[level] ?? RISK_LEVEL_META.L2;
}

/** 是否高风险（需二次确认） */
export function isHighRisk(level: string) {
  return HIGH_RISK_LEVELS.includes(level);
}

/** 执行渠道标签 */
export function channelLabel(channel: string) {
  return channel === 'terminal' ? '专家终端' : '快捷执行';
}

/** 耗时格式化 */
export function formatDuration(ms: number) {
  if (ms == null) return '-';
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

/** 时间格式化 */
export function formatDateTime(value?: string) {
  if (!value) return '-';
  return new Date(value).toLocaleString('zh-CN', { hour12: false });
}
