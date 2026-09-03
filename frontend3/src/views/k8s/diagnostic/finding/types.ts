/**
 * 诊断结果结构化 Finding 模型（五型）
 *
 * 管线：Arthas 输出 → parsers(文本→结构) → rules(判定) → FindingResult 渲染
 * 设计对齐 JMC 规则四段式（id/severity/summary/explanation）+ Watchdog 三分（症状/根因/影响）。
 */

/** 判定级别（数据不足时 unknown，不硬判——宁可未判定不可误报） */
export type FindingLevel = 'ok' | 'info' | 'warning' | 'critical' | 'unknown';

/** 结果渲染分型：判定型/信息型/导出型/变更型（观察型流式场景走终端，不经过此模块） */
export type FindingKind = 'finding' | 'describe' | 'export' | 'mutate';

/** 单条指标（信息型/水位展示） */
export interface FindingMetric {
  label: string;
  value: string;
  level?: FindingLevel;
  /** 0~100，有值时渲染水位进度条 */
  percent?: number;
}

/** 表格行（行级动作联动终端） */
export interface FindingRow {
  cells: string[];
  /** 异常行红标（Dynatrace：引导注意力） */
  highlight?: boolean;
  /** 行备注（如：空闲 worker / IO 等待） */
  note?: string;
  /** 行级动作：送专家终端 */
  actions?: FindingAction[];
}

/** 下一步动作（联动专家终端执行） */
export interface FindingAction {
  label: string;
  command: string;
}

/** 结果分区 */
export interface FindingSection {
  title?: string;
  type: 'metrics' | 'table' | 'kv' | 'code';
  columns?: string[];
  metrics?: FindingMetric[];
  rows?: FindingRow[];
  kv?: { key: string; value: string }[];
  code?: string;
}

/** Finding 主体 */
export interface DiagnosticFinding {
  kind: FindingKind;
  level: FindingLevel;
  /** 症状：直接可观测的事实 */
  symptom?: string;
  /** 根因候选（启发式提取，界面须标注"候选"，由人确认） */
  rootCause?: string;
  /** 影响面（保守推断，无法推断则留空） */
  impact?: string;
  sections: FindingSection[];
  nextActions?: FindingAction[];
  /** 解析失败：降级为仅展示原始输出 */
  parseFailed?: boolean;
}

/** 构建上下文 */
export interface FindingContext {
  scenarioId: string;
  /** 用户实际执行的命令（含参数，动作生成时复用） */
  command: string;
  output: string;
  error?: string;
}
