import { request } from '../request';

/**
 * K8s 应用诊断中心 API（Arthas Tunnel 架构）
 * 后端骨干：tunnel-server agent 注册表 + WS 双桥 + 执行审计
 */

/** 风险级标签映射（与后端 L0-L5 / disabled 对齐） */
export type DiagnosticRiskLevel = 'L0' | 'L1' | 'L2' | 'L3' | 'L4' | 'L5' | 'disabled';

/** 命令目录项（内置 + 管理员覆盖合并后的结果） */
export interface DiagnosticCommandCatalogItem {
  command: string;
  category: string;
  description: string;
  riskLevel: string;
  streaming: boolean;
  usage?: string;
}

/** 应用分组概要 */
export interface DiagnosticAppSummary {
  appName: string;
  total: number;
  online: number;
  agentVersion?: string;
}

/** tunnel agent 实例（约定 agentId = POD_NAME-NAMESPACE） */
export interface DiagnosticAgent {
  id: number;
  clusterId: string;
  appName: string;
  agentId: string;
  namespace: string;
  podName: string;
  agentVersion: string;
  online: boolean;
  lastSeen: string;
}

/** 一次性命令执行结果 */
export interface DiagnosticOneShotResult {
  output: string;
  duration: number;
  riskLevel: string;
  status: 'success' | 'error';
  error?: string;
}

/** 诊断执行记录（审计/历史） */
export interface DiagnosticExecution {
  id: number;
  sessionId: number;
  clusterId: string;
  appName: string;
  namespace: string;
  podName: string;
  agentId: string;
  command: string;
  riskLevel: string;
  channel: 'oneshot' | 'terminal';
  resultStatus: 'success' | 'error';
  resultOutput?: string;
  resultError?: string;
  duration: number;
  userId: number;
  username: string;
  timestamp: string;
}

/** 执行历史查询参数 */
export interface DiagnosticExecutionQuery {
  page?: number;
  pageSize?: number;
  clusterId?: string;
  appName?: string;
  agentId?: string;
  username?: string;
  command?: string;
  riskLevel?: string;
}

/** 执行统计 */
export interface DiagnosticExecutionStats {
  total: number;
  highRisk: number;
  errorCount: number;
  activeSess: number;
}

/** 诊断总览数据 */
export interface DiagnosticOverviewData {
  apps: number;
  agentTotal: number;
  agentOnline: number;
  execStats: DiagnosticExecutionStats;
  appList: DiagnosticAppSummary[];
}

/** 诊断会话（专家终端） */
export interface DiagnosticSession {
  id: number;
  sessionKey: string;
  clusterId: string;
  appName: string;
  agentId: string;
  status: 'active' | 'closed' | 'timeout' | 'terminated';
  userId: number;
  username: string;
  ioLog?: string;
  startAt: string;
  endAt?: string;
}

// ========== 命令目录 ==========

/** 获取 Arthas 全量命令目录（含风险分级） */
export function fetchDiagnosticCommands() {
  return request<DiagnosticCommandCatalogItem[]>({
    url: '/k8s/diagnostic/commands',
    method: 'get'
  });
}

// ========== 目标发现 ==========

/** 获取已接入诊断的应用列表（自动同步 tunnel agent 注册表） */
export function fetchDiagnosticApps() {
  return request<DiagnosticAppSummary[]>({
    url: '/k8s/diagnostic/apps',
    method: 'get'
  });
}

/** 获取应用下的实例（agent）列表 */
export function fetchDiagnosticAppAgents(appName: string) {
  return request<DiagnosticAgent[]>({
    url: `/k8s/diagnostic/apps/${encodeURIComponent(appName)}/agents`,
    method: 'get'
  });
}

// ========== 一次性执行 ==========

/** 经 tunnel 执行单条诊断命令（适合只读命令；流式命令请使用专家终端） */
export function executeDiagnosticOneShot(data: {
  clusterId?: string;
  agentId: string;
  command: string;
  timeout?: number;
}) {
  return request<DiagnosticOneShotResult>({
    url: '/k8s/diagnostic/execute',
    method: 'post',
    data
  });
}

// ========== 执行历史 ==========

/** 分页查询诊断执行记录 */
export function fetchDiagnosticExecutions(params: DiagnosticExecutionQuery) {
  return request<{
    list: DiagnosticExecution[];
    total: number;
    page: number;
    pageSize: number;
  }>({
    url: '/k8s/diagnostic/executions',
    method: 'get',
    params
  });
}

// ========== 总览 ==========

/** 诊断中心总览数据 */
export function fetchDiagnosticOverview() {
  return request<DiagnosticOverviewData>({
    url: '/k8s/diagnostic/overview',
    method: 'get'
  });
}

// ========== 会话治理 ==========

/** 获取活跃诊断会话列表 */
export function fetchDiagnosticSessions() {
  return request<DiagnosticSession[]>({
    url: '/k8s/diagnostic/sessions',
    method: 'get'
  });
}

/** 获取会话详情（含 I/O 回放数据） */
export function fetchDiagnosticSessionDetail(sessionId: number) {
  return request<DiagnosticSession>({
    url: `/k8s/diagnostic/sessions/${sessionId}`,
    method: 'get'
  });
}

/** 强制断开诊断会话（管理员熔断） */
export function terminateDiagnosticSession(sessionId: number) {
  return request({
    url: `/k8s/diagnostic/sessions/${sessionId}/terminate`,
    method: 'post'
  });
}

// ========== 接入与治理（管理员） ==========

/** 命令风险覆盖记录 */
export interface DiagnosticCommandOverride {
  id: number;
  command: string;
  riskLevel: string; // L0-L5 或 disabled
  description: string;
  updatedBy: string;
  updatedAt: string;
}

/** 获取命令风险覆盖表 */
export function fetchDiagnosticCommandOverrides() {
  return request<DiagnosticCommandOverride[]>({
    url: '/k8s/diagnostic/command-overrides',
    method: 'get'
  });
}

/** 新增/更新命令风险覆盖（返回合并后的最新命令目录） */
export function saveDiagnosticCommandOverride(data: { command: string; riskLevel: string; description?: string }) {
  return request<DiagnosticCommandCatalogItem[]>({
    url: '/k8s/diagnostic/command-overrides',
    method: 'post',
    data
  });
}

/** 删除命令风险覆盖（恢复内置默认分级），返回最新目录 */
export function deleteDiagnosticCommandOverride(id: number) {
  return request<DiagnosticCommandCatalogItem[]>({
    url: `/k8s/diagnostic/command-overrides/${id}`,
    method: 'delete'
  });
}

// ========== Webhook 自动注入（Pod label 触发） ==========

/** Webhook 注入状态 */
export interface DiagnosticWebhookStatus {
  enabled: boolean;
  /** apiserver 回调地址（全局配置） */
  webhookURL: string;
  /** agent 反连地址（集群内） */
  tunnelWS: string;
  /** 平台 HTTP 探测地址 */
  tunnelURL: string;
  /** 平台 WS 会话地址 */
  tunnelSess: string;
  initImage: string;
  objectRule: string;
}

/** 查询集群注入状态（MWC 是否存在、当前配置） */
export function fetchDiagnosticWebhookStatus(clusterId: number) {
  return request<DiagnosticWebhookStatus>({
    url: '/k8s/diagnostic/webhook',
    method: 'get',
    params: { clusterId }
  });
}

/** 启用集群自动注入（创建/更新 MutatingWebhookConfiguration） */
export function enableDiagnosticWebhook(clusterId: number) {
  return request({
    url: '/k8s/diagnostic/webhook/enable',
    method: 'post',
    data: { clusterId }
  });
}

/** 禁用集群自动注入（删除 MWC，存量已注入 Pod 不受影响） */
export function disableDiagnosticWebhook(clusterId: number) {
  return request({
    url: '/k8s/diagnostic/webhook/disable',
    method: 'post',
    data: { clusterId }
  });
}

/** 保存注入配置 */
export function saveDiagnosticWebhookConfig(data: {
  clusterId: number;
  tunnelWS?: string;
  initImage?: string;
  webhookURL?: string;
  tunnelURL?: string;
  tunnelSess?: string;
}) {
  return request({
    url: '/k8s/diagnostic/webhook/config',
    method: 'post',
    data
  });
}
