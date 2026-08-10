import { request } from '../request';

/**
 * 监控概览数据
 */
export interface MonitoringOverview {
  summary: {
    totalServers: number;
    onlineServers: number;
    offlineServers: number;
    alertServers: number;
  };
  topCpu: Array<{
    serverId: number;
    hostname: string;
    ip: string;
    cpuUsage: number;
    trend: number[];
  }>;
  topMemory: Array<{
    serverId: number;
    hostname: string;
    ip: string;
    memoryUsage: number;
    trend: number[];
  }>;
  topDisk: Array<{
    serverId: number;
    hostname: string;
    ip: string;
    diskUsage: number;
    trend: number[];
  }>;
  activeAlerts: Array<{
    id: number;
    serverId: number;
    hostname: string;
    level: string;
    message: string;
    firstSeen: string;
    lastSeen: string;
  }>;
  refreshTime: string;
}

/**
 * 扩展指标数据
 */
export interface ExtendedMetrics {
  performance: {
    cpu: {
      usagePercent: number;
      user: number;
      system: number;
      idle: number;
      iowait: number;
      cores: number;
      mhz: number;
    };
    memory: {
      total: number;
      used: number;
      free: number;
      usedPercent: number;
      available: number;
    };
    disk: {
      total: number;
      used: number;
      free: number;
      usedPercent: number;
      partitions: Array<{
        device: string;
        mountpoint: string;
        fstype: string;
        total: number;
        used: number;
        free: number;
        usedPercent: number;
      }>;
    };
    network: {
      interfaces: Array<{
        name: string;
        bytesSent: number;
        bytesRecv: number;
      }>;
      connections: {
        established: number;
        timeWait: number;
        listen: number;
      };
    };
    load: {
      load1: number;
      load5: number;
      load15: number;
    };
  };
  systemInfo: {
    hostname: string;
    os: {
      platform: string;
      platformVersion: string;
      kernelVersion: string;
      kernelArch: string;
    };
    uptime: number;
  };
  hardwareInfo: {
    cpu: {
      vendor: string;
      model: string;
      cores: number;
      threads: number;
      mhz: number;
    };
    memory: {
      total: number;
    };
    disk: Array<{
      name: string;
      model: string;
      serial: string;
      size: number;
      type: string;
    }>;
  };
  serviceStatus: {
    systemdServices: Array<{
      name: string;
      status: string;
      subStatus: string;
      description: string;
    }>;
    listenPorts: Array<{
      port: number;
      protocol: string;
      address: string;
      process: string;
      pid: number;
    }>;
  };
  processInfo: {
    total: number;
    top: Array<{
      pid: number;
      name: string;
      cpuPercent: number;
      memoryPercent: number;
      memoryBytes: number;
      status: string;
      username: string;
      numThreads: number;
      cmdline: string;
    }>;
  };
  networkConfig: {
    interfaces: Array<{
      name: string;
      hardwareAddr: string;
      mtu: number;
      addrs: Array<{
        ip: string;
      }>;
    }>;
  };
  securityInfo: {
    ssh: {
      port: number;
      permitRootLogin: string;
      passwordAuthentication: string;
    };
    firewall: {
      backend: string;
      status: string;
    };
    selinux: {
      enabled: boolean;
      mode: string;
    };
  };
  collectedAt: string;
  version: string;
}

/**
 * 历史指标数据点
 */
export interface MetricsDatapoint {
  timestamp: string;
  value: number | string;
}

/**
 * 历史指标响应
 */
export interface MetricsHistoryResponse {
  metricType: string;
  interval: string;
  datapoints: MetricsDatapoint[];
}

/**
 * 告警列表项
 */
export interface Alert {
  id: number;
  serverId: number;
  hostname: string;
  ip: string;
  ruleId: string;
  level: string;
  message: string;
  metricValue: number;
  threshold: number;
  firstSeen: string;
  lastSeen: string;
  acknowledged: boolean;
  acknowledgedBy: string | null;
  acknowledgedAt: string | null;
  resolvedAt: string | null;
}

/**
 * 告警列表响应
 */
export interface AlertListResponse {
  total: number;
  items: Alert[];
}

/**
 * 告警统计
 */
export interface AlertStats {
  total: number;
  byLevel: {
    critical: number;
    high: number;
    medium: number;
    low: number;
    info: number;
  };
  acknowledged: number;
  unacknowledged: number;
  resolved: number;
  active: number;
  trend: Array<{
    date: string;
    count: number;
  }>;
}

/**
 * 获取监控概览
 */
export function fetchMonitoringOverview() {
  return request<MonitoringOverview>({
    url: '/monitoring/overview',
    method: 'get'
  });
}

/**
 * 获取主机扩展指标
 */
export function fetchServerExtendedMetrics(serverId: number) {
  return request<ExtendedMetrics>({
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
  return request<MetricsHistoryResponse>({
    url: `/cmdb/servers/${serverId}/metrics/history`,
    method: 'get',
    params
  });
}

/**
 * 获取主机硬件信息
 */
export function fetchServerHardware(serverId: number) {
  return request<Pick<ExtendedMetrics, 'hardwareInfo'>>({
    url: `/cmdb/servers/${serverId}/hardware`,
    method: 'get'
  });
}

/**
 * 获取主机进程信息
 */
export function fetchServerProcesses(serverId: number) {
  return request<Pick<ExtendedMetrics, 'processInfo'>>({
    url: `/cmdb/servers/${serverId}/processes`,
    method: 'get'
  });
}

/**
 * 获取主机服务状态
 */
export function fetchServerServices(serverId: number) {
  return request<Pick<ExtendedMetrics, 'serviceStatus'>>({
    url: `/cmdb/servers/${serverId}/services`,
    method: 'get'
  });
}

/**
 * 获取主机网络配置
 */
export function fetchServerNetwork(serverId: number) {
  return request<Pick<ExtendedMetrics, 'networkConfig'>>({
    url: `/cmdb/servers/${serverId}/network`,
    method: 'get'
  });
}

/**
 * 获取主机安全信息
 */
export function fetchServerSecurity(serverId: number) {
  return request<Pick<ExtendedMetrics, 'securityInfo'>>({
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
  return request<AlertListResponse>({
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
  return request<AlertStats>({
    url: '/monitoring/alerts/stats',
    method: 'get',
    params
  });
}

/**
 * 告警规则
 */
export interface AlertRule {
  id: string;
  name: string;
  level: 'critical' | 'high' | 'medium' | 'low' | 'info';
  metric: 'cpu_usage' | 'memory_usage' | 'disk_usage' | 'load1' | 'load5' | 'load15';
  condition: '>' | '<' | '==' | '!=';
  threshold: number;
  duration: number;
  description: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * 告警规则表单
 */
export interface AlertRuleForm {
  id?: string;
  name: string;
  level: 'critical' | 'high' | 'medium' | 'low' | 'info';
  metric: 'cpu_usage' | 'memory_usage' | 'disk_usage' | 'load1' | 'load5' | 'load15';
  condition: '>' | '<' | '==' | '!=';
  threshold: number;
  duration: number;
  description?: string;
  enabled?: boolean;
}

/**
 * 获取告警规则列表
 */
export function fetchAlertRules() {
  return request<AlertRule[]>({
    url: '/monitoring/alerts/rules',
    method: 'get'
  });
}

/**
 * 创建告警规则
 */
export function createAlertRule(data: AlertRuleForm) {
  return request({
    url: '/monitoring/alerts/rules',
    method: 'post',
    data
  });
}

/**
 * 更新告警规则
 */
export function updateAlertRule(id: string, data: AlertRuleForm) {
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
 * 通知渠道
 */
export interface NotificationChannel {
  id: number;
  channelType: 'email' | 'wechat' | 'dingtalk' | 'feishu';
  channelName: string;
  config: Record<string, unknown>;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * 通知渠道表单
 */
export interface NotificationChannelForm {
  id?: number;
  channelType: 'email' | 'wechat' | 'dingtalk' | 'feishu';
  channelName: string;
  config: Record<string, unknown>;
  enabled?: boolean;
}

/**
 * 获取通知渠道列表
 */
export function fetchNotificationChannels() {
  return request<NotificationChannel[]>({
    url: '/monitoring/notifications/channels',
    method: 'get'
  });
}

/**
 * 创建通知渠道
 */
export function createNotificationChannel(data: NotificationChannelForm) {
  return request<{ id: number }>({
    url: '/monitoring/notifications/channels',
    method: 'post',
    data
  });
}

/**
 * 更新通知渠道
 */
export function updateNotificationChannel(id: number, data: NotificationChannelForm) {
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
