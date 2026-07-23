import { request } from '../request';

// 诊断命令
export interface DiagnosticCommand {
  id: string;
  name: string;
  description: string;
  category: string;
  defaultArgs: Record<string, string>;
}

// Java Pod信息
export interface JavaPod {
  podName: string;
  namespace: string;
  ip: string;
  phase: string;
  hasAgent: boolean;
  nodeName?: string;
}

// 诊断结果
export interface DiagnosticResult {
  status: 'success' | 'error';
  output?: string;
  error?: string;
  timestamp: number;
  duration: number;
  method?: string;
}

// 诊断历史
export interface DiagnosticHistory {
  id: number;
  clusterId: string;
  clusterName: string;
  namespace: string;
  podName: string;
  command: string;
  args: string;
  resultStatus: string;
  resultOutput?: string;
  resultError?: string;
  resultMethod?: string;
  resultDuration?: number;
  userId: number;
  username: string;
  timestamp: string;
}

// 获取诊断命令列表
export function fetchDiagnosticCommands() {
  return request<{
    commands: DiagnosticCommand[];
  }>({
    url: '/k8s/diagnostic/commands',
    method: 'get'
  });
}

// 获取Java应用Pod列表
export function fetchJavaPods(clusterId: string | number, namespace: string) {
  return request<JavaPod[]>({
    url: `/k8s/diagnostic/pods/${clusterId}/${namespace}`,
    method: 'get'
  });
}

// 执行诊断
export function executeDiagnostic(data: {
  clusterId: string | number;
  namespace: string;
  podName: string;
  command: string;
  args: Record<string, string>;
  timeout?: number;
}) {
  return request<DiagnosticResult>({
    url: '/k8s/diagnostic/execute',
    method: 'post',
    data
  });
}

// 获取诊断历史
export function fetchDiagnosticHistory(params: {
  clusterId?: string;
  namespace?: string;
  podName?: string;
  page?: number;
  pageSize?: number;
}) {
  return request<{
    list: DiagnosticHistory[];
    total: number;
  }>({
    url: '/k8s/diagnostic/history',
    method: 'get',
    params
  });
}
