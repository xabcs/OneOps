declare namespace Monitoring {
  /** 告警级别 */
  type AlertLevel = 'critical' | 'high' | 'medium' | 'low' | 'info';

  /** 告警状态 */
  type AlertStatus = 'pending' | 'acknowledged' | 'resolved';

  /** 告警条件 */
  type AlertCondition = '>' | '<' | '==' | '!=';

  /** 监控指标 */
  type AlertMetric = 'cpu_usage' | 'memory_usage' | 'disk_usage' | 'load1' | 'load5' | 'load15';

  /** 告警规则 */
  type AlertRule = {
    id: string;
    name: string;
    level: AlertLevel;
    metric: AlertMetric;
    condition: AlertCondition;
    threshold: number;
    duration: number;
    description: string;
    enabled: boolean;
    createdAt: string;
    updatedAt: string;
  };

  /** 告警规则表单 */
  type AlertRuleForm = {
    id?: string;
    name: string;
    level: AlertLevel;
    metric: AlertMetric;
    condition: AlertCondition;
    threshold: number;
    duration: number;
    description?: string;
    enabled?: boolean;
  };

  /** 告警 */
  type Alert = {
    id: number;
    serverId: number;
    hostname: string;
    ip: string;
    ruleId: string;
    level: AlertLevel;
    message: string;
    metricValue: number;
    threshold: number;
    firstSeen: string;
    lastSeen: string;
    resolvedAt?: string;
    acknowledgedBy?: string;
    acknowledgedAt?: string;
    acknowledgementComment?: string;
    createdAt: string;
    updatedAt: string;
  };

  /** 告警查询参数 */
  type AlertQuery = {
    serverId?: string;
    level?: AlertLevel;
    acknowledged?: string;
    page?: number;
    pageSize?: number;
  };

  /** 告警统计 */
  type AlertStats = {
    total: number;
    pending: number;
    acknowledged: number;
    resolved: number;
    byLevel: Record<AlertLevel, number>;
  };

  /** 通知渠道类型 */
  type NotificationChannelType = 'email' | 'wechat' | 'dingtalk' | 'feishu';

  /** 通知渠道 */
  type NotificationChannel = {
    id: number;
    channelType: NotificationChannelType;
    channelName: string;
    config: Record<string, unknown>;
    enabled: boolean;
    createdAt: string;
    updatedAt: string;
  };

  /** 通知渠道表单 */
  type NotificationChannelForm = {
    id?: number;
    channelType: NotificationChannelType;
    channelName: string;
    config: Record<string, unknown>;
    enabled?: boolean;
  };

  /** 邮件配置 */
  type EmailConfig = {
    smtpHost: string;
    smtpPort: number;
    username: string;
    password: string;
    from: string;
    fromName: string;
    useTLS: boolean;
    recipients: string[];
  };

  /** 企业微信配置 */
  type WeChatConfig = {
    webhookUrl: string;
    mentionedList?: string[];
  };

  /** 主机扩展指标 */
  type ExtendedMetrics = {
    performance: PerformanceMetrics;
    systemInfo: SystemInfo;
    hardwareInfo: HardwareInfo;
    serviceStatus: ServiceStatus;
    processInfo: ProcessInfo;
    networkConfig: NetworkConfig;
    securityInfo: SecurityInfo;
    collectedAt: string;
    version: string;
  };

  /** 性能指标 */
  type PerformanceMetrics = {
    cpu: CPUMetrics;
    memory: MemoryMetrics;
    disk: DiskMetrics;
    network: NetworkMetrics;
    load: LoadMetrics;
    io?: IOStats;
  };

  /** CPU 指标 */
  type CPUMetrics = {
    usagePercent: number;
    user: number;
    system: number;
    idle: number;
    iowait: number;
    cores: number;
    mhz: number;
  };

  /** 内存指标 */
  type MemoryMetrics = {
    total: number;
    used: number;
    free: number;
    usedPercent: number;
    available: number;
  };

  /** 磁盘指标 */
  type DiskMetrics = {
    total: number;
    used: number;
    free: number;
    usedPercent: number;
    partitions: PartitionInfo[];
  };

  /** 分区信息 */
  type PartitionInfo = {
    device: string;
    mountpoint: string;
    fstype: string;
    total: number;
    used: number;
    free: number;
    usedPercent: number;
  };

  /** IO 统计 */
  type IOStats = {
    readIOPS: number;
    writeIOPS: number;
    readThroughput: number;
    writeThroughput: number;
    await: number;
    queueDepth: number;
  };

  /** 网络指标 */
  type NetworkMetrics = {
    interfaces: InterfaceStats[];
    connections: ConnectionStats;
  };

  /** 网卡统计 */
  type InterfaceStats = {
    name: string;
    bytesSent: number;
    bytesRecv: number;
  };

  /** 连接统计 */
  type ConnectionStats = {
    established: number;
    timeWait: number;
    listen: number;
  };

  /** 负载指标 */
  type LoadMetrics = {
    load1: number;
    load5: number;
    load15: number;
  };

  /** 系统信息 */
  type SystemInfo = {
    hostname: string;
    os: OSInfo;
    uptime: number;
  };

  /** 操作系统信息 */
  type OSInfo = {
    platform: string;
    platformVersion: string;
    kernelVersion: string;
    kernelArch: string;
  };

  /** 硬件信息 */
  type HardwareInfo = {
    cpu: CPUInfo;
    memory: MemoryHardware;
    disk: DiskDevice[];
  };

  /** CPU 信息 */
  type CPUInfo = {
    vendor: string;
    model: string;
    cores: number;
    threads: number;
    mhz: number;
  };

  /** 内存硬件信息 */
  type MemoryHardware = {
    total: number;
    slots: MemorySlot[];
  };

  /** 内存插槽 */
  type MemorySlot = {
    slotNumber: number;
    capacity: number;
    type: string;
    vendor: string;
    speed: string;
    hasECC: boolean;
  };

  /** 磁盘设备 */
  type DiskDevice = {
    name: string;
    model: string;
    serial: string;
    size: number;
    type: string;
    firmware: string;
    smartStatus: string;
  };

  /** 服务状态 */
  type ServiceStatus = {
    systemdServices: SystemdService[];
    listenPorts: ListenPort[];
  };

  /** systemd 服务 */
  type SystemdService = {
    name: string;
    status: string;
    subStatus: string;
    activeState: string;
    subState: string;
    description: string;
  };

  /** 监听端口 */
  type ListenPort = {
    port: number;
    protocol: string;
    address: string;
    process: string;
    pid: number;
  };

  /** 进程信息 */
  type ProcessInfo = {
    total: number;
    top: TopProcess[];
  };

  /** Top 进程 */
  type TopProcess = {
    pid: number;
    name: string;
    cpuPercent: number;
    memoryPercent: number;
    memoryBytes: number;
    status: string;
    username: string;
    numThreads: number;
    cmdline: string;
  };

  /** 网络配置 */
  type NetworkConfig = {
    interfaces: NetworkInterface[];
    routes?: RouteEntry[];
    dns?: DNSConfig;
  };

  /** 网络接口 */
  type NetworkInterface = {
    name: string;
    hardwareAddr: string;
    mtu: number;
    addrs: InterfaceAddr[];
    flags?: string[];
    isUp?: boolean;
  };

  /** 接口地址 */
  type InterfaceAddr = {
    ip: string;
    mask?: string;
    family?: string;
  };

  /** 路由表项 */
  type RouteEntry = {
    destination: string;
    gateway: string;
    interface: string;
    flags?: string;
    metric?: number;
  };

  /** DNS 配置 */
  type DNSConfig = {
    servers: string[];
    searchDomains: string[];
  };

  /** 安全信息 */
  type SecurityInfo = {
    ssh: SSHConfig;
    firewall: FirewallConfig;
  };

  /** SSH 配置 */
  type SSHConfig = {
    port: number;
    permitRootLogin: string;
    passwordAuthentication: string;
  };

  /** 防火墙配置 */
  type FirewallConfig = {
    backend: string;
    status: string;
  };

  /** 监控概览 */
  type Overview = {
    totalServers: number;
    onlineServers: number;
    offlineServers: number;
    totalAlerts: number;
    criticalAlerts: number;
    highAlerts: number;
    avgCpuUsage: number;
    avgMemoryUsage: number;
    avgDiskUsage: number;
    topCpuServers: ServerMetric[];
    topMemoryServers: ServerMetric[];
    topDiskServers: ServerMetric[];
    recentAlerts: Alert[];
  };

  /** 主机指标 */
  type ServerMetric = {
    serverId: number;
    hostname: string;
    ip: string;
    cpuUsage: number;
    memoryUsage: number;
    diskUsage: number;
    load1: number;
  };

  /** 历史指标数据点 */
  type MetricDatapoint = {
    timestamp: string;
    value: number;
  };

  /** 历史指标查询参数 */
  type MetricsHistoryQuery = {
    serverId: number;
    metricType: string;
    startTime: string;
    endTime: string;
    interval?: string;
  };
}
