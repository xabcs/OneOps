import { request } from '../request';

/**
 * 获取监控概览
 */
export function fetchMonitoringOverview() {
  return request<Monitoring.MonitoringOverview>({
    url: '/monitoring/overview',
    method: 'get'
  });
}

/**
 * 获取主机扩展指标
 */
export function fetchServerExtendedMetrics(serverId: number) {
  return request<Monitoring.ExtendedMetrics>({
    url: `/cmdb/servers/${serverId}/extended-metrics`,
    method: 'get'
  });
}

/**
 * 查询主机历史指标
 */
export function fetchServerMetricsHistory(
  serverId: number,
  params: {
    metricType: string;
    startTime: string;
    endTime: string;
    interval?: string;
  }
) {
  return request<Monitoring.MetricsHistoryResponse>({
    url: `/cmdb/servers/${serverId}/metrics/history`,
    method: 'get',
    params
  });
}

/**
 * 获取主机硬件信息
 */
export function fetchServerHardware(serverId: number) {
  return request<Pick<Monitoring.ExtendedMetrics, 'hardwareInfo'>>({
    url: `/cmdb/servers/${serverId}/hardware`,
    method: 'get'
  });
}

/**
 * 获取主机进程信息
 */
export function fetchServerProcesses(serverId: number) {
  return request<Pick<Monitoring.ExtendedMetrics, 'processInfo'>>({
    url: `/cmdb/servers/${serverId}/processes`,
    method: 'get'
  });
}

/**
 * 获取主机服务状态
 */
export function fetchServerServices(serverId: number) {
  return request<Pick<Monitoring.ExtendedMetrics, 'serviceStatus'>>({
    url: `/cmdb/servers/${serverId}/services`,
    method: 'get'
  });
}

/**
 * 获取主机网络配置
 */
export function fetchServerNetwork(serverId: number) {
  return request<Pick<Monitoring.ExtendedMetrics, 'networkConfig'>>({
    url: `/cmdb/servers/${serverId}/network`,
    method: 'get'
  });
}

/**
 * 获取主机安全信息
 */
export function fetchServerSecurity(serverId: number) {
  return request<Pick<Monitoring.ExtendedMetrics, 'securityInfo'>>({
    url: `/cmdb/servers/${serverId}/security`,
    method: 'get'
  });
}

/**
 * 获取告警列表
 */
export function fetchAlerts(params: {
  serverId?: number;
  level?: string;
  acknowledged?: boolean;
  resolved?: boolean;
  page: number;
  pageSize: number;
}) {
  return request<Monitoring.AlertListResponse>({
    url: '/monitoring/alerts',
    method: 'get',
    params
  });
}

/**
 * 确认告警
 */
export function acknowledgeAlert(alertId: number, comment?: string) {
  return request({
    url: `/monitoring/alerts/${alertId}/acknowledge`,
    method: 'post',
    data: { comment }
  });
}

/**
 * 获取告警统计
 */
export function fetchAlertStats(params?: { startTime?: string; endTime?: string }) {
  return request<Monitoring.AlertStats>({
    url: '/monitoring/alerts/stats',
    method: 'get',
    params
  });
}

/**
 * 获取告警规则列表
 */
export function fetchAlertRules() {
  return request<Monitoring.AlertRule[]>({
    url: '/monitoring/alerts/rules',
    method: 'get'
  });
}

/**
 * 创建告警规则
 */
export function createAlertRule(data: Monitoring.AlertRuleForm) {
  return request({
    url: '/monitoring/alerts/rules',
    method: 'post',
    data
  });
}

/**
 * 更新告警规则
 */
export function updateAlertRule(id: string, data: Monitoring.AlertRuleForm) {
  return request({
    url: `/monitoring/alerts/rules/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除告警规则
 */
export function deleteAlertRule(id: string) {
  return request({
    url: `/monitoring/alerts/rules/${id}`,
    method: 'delete'
  });
}

/**
 * 更新告警规则状态
 */
export function updateAlertRuleStatus(id: string, enabled: boolean) {
  return request({
    url: `/monitoring/alerts/rules/${id}/status`,
    method: 'put',
    data: { enabled }
  });
}

/**
 * 获取通知渠道列表
 */
export function fetchNotificationChannels() {
  return request<Monitoring.NotificationChannel[]>({
    url: '/monitoring/notifications/channels',
    method: 'get'
  });
}

/**
 * 创建通知渠道
 */
export function createNotificationChannel(data: Monitoring.NotificationChannelForm) {
  return request<{ id: number }>({
    url: '/monitoring/notifications/channels',
    method: 'post',
    data
  });
}

/**
 * 更新通知渠道
 */
export function updateNotificationChannel(id: number, data: Monitoring.NotificationChannelForm) {
  return request({
    url: `/monitoring/notifications/channels/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除通知渠道
 */
export function deleteNotificationChannel(id: number) {
  return request({
    url: `/monitoring/notifications/channels/${id}`,
    method: 'delete'
  });
}

/**
 * 测试通知渠道
 */
export function testNotificationChannel(id: number) {
  return request({
    url: `/monitoring/notifications/channels/${id}/test`,
    method: 'post'
  });
}
