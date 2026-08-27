/**
 * 工单模块公共常量：状态/优先级/节点状态的文案与标签颜色映射。
 * 与后端 model/ticket/consts.go 的枚举值保持一致，各页面统一引用。
 */

/** 工单状态 → 文案/标签颜色 */
export const ticketStatusMap: Record<string, { label: string; type: 'primary' | 'success' | 'danger' | 'info' }> = {
  pending: { label: '审批中', type: 'primary' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' },
  canceled: { label: '已撤销', type: 'info' }
};

/** 工单状态选项（筛选下拉用） */
export const ticketStatusOptions = Object.entries(ticketStatusMap).map(([value, { label }]) => ({ value, label }));

/** 优先级 → 文案/标签颜色 */
export const ticketPriorityMap: Record<string, { label: string; type: 'info' | 'primary' | 'warning' | 'danger' }> = {
  low: { label: '低', type: 'info' },
  normal: { label: '中', type: 'primary' },
  high: { label: '高', type: 'warning' },
  urgent: { label: '紧急', type: 'danger' }
};

/** 优先级选项（发起/重提弹窗单选用） */
export const ticketPriorityOptions = Object.entries(ticketPriorityMap).map(([value, { label }]) => ({
  value,
  label
}));

/** 节点审批状态 → 文案/steps 类型 */
export const nodeStatusMap: Record<string, { label: string; type: 'wait' | 'process' | 'finish' | 'error' }> = {
  waiting: { label: '未到达', type: 'wait' },
  pending: { label: '审批中', type: 'process' },
  approved: { label: '已通过', type: 'finish' },
  rejected: { label: '已驳回', type: 'error' },
  skipped: { label: '已跳过', type: 'wait' },
  canceled: { label: '已终止', type: 'wait' }
};

/** 节点审批状态 → 标签颜色 */
export const nodeTagType: Record<string, 'info' | 'primary' | 'success' | 'danger'> = {
  waiting: 'info',
  pending: 'primary',
  approved: 'success',
  rejected: 'danger',
  skipped: 'info',
  canceled: 'info'
};

/** 催办冷却时长（毫秒），与后端 modelticket.UrgeCooldown 保持一致 */
export const URGE_COOLDOWN_MS = 30 * 60 * 1000;
