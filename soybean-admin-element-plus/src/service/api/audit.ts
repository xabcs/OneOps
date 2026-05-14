import { request } from '../request';

/**
 * 获取审计统计信息
 */
export async function fetchGetAuditStats() {
  return request<Audit.Stats>({
    url: '/audit/stats',
    method: 'get'
  });
}

/**
 * 获取登录日志列表
 */
export async function fetchGetLoginLogs(params: Audit.LogQuery) {
  return request<Api.PaginatingQueryRecord<Audit.LoginLog>>({
    url: '/audit/login-logs',
    method: 'get',
    params
  });
}

/**
 * 导出登录日志
 */
export async function fetchExportLoginLogs(params: Audit.LogQuery) {
  return request<Blob>({
    url: '/audit/login-logs/export',
    method: 'get',
    params,
    responseType: 'blob'
  });
}

/**
 * 获取操作日志列表
 */
export async function fetchGetOperationLogs(params: Audit.LogQuery) {
  return request<Api.PaginatingQueryRecord<Audit.OperationLog>>({
    url: '/audit/operation-logs',
    method: 'get',
    params
  });
}

/**
 * 导出操作日志
 */
export async function fetchExportOperationLogs(params: Audit.LogQuery) {
  return request<Blob>({
    url: '/audit/operation-logs/export',
    method: 'get',
    params,
    responseType: 'blob'
  });
}

/**
 * 获取系统事件日志列表
 */
export async function fetchGetSystemEventLogs(params: Audit.LogQuery) {
  return request<Api.PaginatingQueryRecord<Audit.SystemEventLog>>({
    url: '/audit/system-event-logs',
    method: 'get',
    params
  });
}

/**
 * 获取可用模块列表
 */
export async function fetchGetModules() {
  return request<string[]>({
    url: '/audit/modules',
    method: 'get'
  });
}