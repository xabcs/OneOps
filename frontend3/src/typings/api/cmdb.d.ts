declare namespace CMDB {
  /** 服务器类型 */
  type ServerType = 'physical' | 'vm' | 'container';

  /** 服务器环境 */
  type ServerEnv = 'prod' | 'test' | 'dev';

  /** 服务器状态 */
  type ServerStatus = 'online' | 'offline' | 'unknown';

  /** 服务器提供商 */
  type ServerProvider = 'aliyun' | 'tencent' | 'aws' | 'self' | 'huawei' | 'other';

  type CloudProvider = Exclude<ServerProvider, 'self'>;

  /** 服务器 */
  type Server = {
    id: number;
    hostname: string;
    ip: string;
    innerIp?: string;
    cpu: number;
    memory: number;
    disk: number;
    os?: string;
    osVersion?: string;
    arch?: string;
    env: ServerEnv;
    status: ServerStatus;
    sshPort: number;
    credentialId?: number;
    cabinetId?: number;
    uPosition?: number;
    sn?: string;
    manufacturer?: string;
    model?: string;
    purchaseDate?: string;
    expireWarranty?: string;
    assetNumber?: string;
    provider: ServerProvider;
    serverType: ServerType;
    businessId?: number;
    business?: BusinessUnit;
    remarks?: string;
    createdAt: string;
    updatedAt: string;
    lastCheckTime?: string;
    connectivityStatus?: 'online' | 'offline' | 'unknown';
    lastConnectTime?: string;
    sshCredentialId?: number;
    cpuUsage?: number;
    memoryUsage?: number;
    diskUsage?: number;
    load1?: number;
    load5?: number;
    load15?: number;
    metricsUpdatedAt?: string;
    agentStatus?: 'uninstalled' | 'running' | 'offline';
    agentPort?: number;
    agentVersion?: string;
    lastHeartbeatAt?: string;
    systemCredentialId?: number;
    // 关联数据
    credential?: SSHCredential;
    credentials?: SSHCredential[];
    systemCredential?: SSHCredential;
    cabinet?: Cabinet;
    tags?: ServerTag[];
    groups?: ServerGroup[];
    cloudInfo?: CloudServer;
    // 监控增强字段
    cpuTrend?: number[];
    memoryTrend?: number[];
    diskTrend?: number[];
    serviceStatus?: 'running' | 'active' | 'dead' | 'failed' | 'offline' | 'unknown';
    alertCount?: number;
    diskPartitions?: Array<{ mount: string; usage: number }>;
  };

  /** 业务系统 */
  type BusinessUnit = {
    id: number;
    name: string;
    code: string;
    parentId: number;
    level: number;
    owner?: string;
    phone?: string;
    email?: string;
    sortOrder: number;
    status: number;
    remarks?: string;
    createdAt: string;
    updatedAt: string;
    children?: BusinessUnit[];
    servers?: Server[];
  };

  /** 机房 */
  type ServerRoom = {
    id: number;
    name: string;
    code: string;
    location?: string;
    address?: string;
    provider?: string;
    contact?: string;
    phone?: string;
    status: number;
    remarks?: string;
    createdAt: string;
    updatedAt: string;
    cabinets?: Cabinet[];
  };

  /** 机柜 */
  type Cabinet = {
    id: number;
    name: string;
    code: string;
    roomId?: number;
    position?: string;
    capacity: number;
    usedU: number;
    powerUsage: number;
    powerCapacity: number;
    status: number;
    remarks?: string;
    createdAt: string;
    updatedAt: string;
    room?: ServerRoom;
    servers?: Server[];
  };

  /** 服务器标签 */
  type ServerTag = {
    id: number;
    name: string;
    color: string;
    description?: string;
    sortOrder: number;
    status: number;
    createdAt: string;
    servers?: Server[];
  };

  /** 主机分组 */
  type ServerGroup = {
    id: number;
    name: string;
    code: string;
    parentId: number;
    level: number;
    description?: string;
    color: string;
    icon: string;
    sortOrder: number;
    status: number;
    createdAt: string;
    updatedAt: string;
    parent?: ServerGroup;
    children?: ServerGroup[];
    servers?: Server[];
  };

  /** 主机分组表单 */
  type ServerGroupForm = {
    id?: number;
    name: string;
    code: string;
    parentId: number;
    description?: string;
    color?: string;
    icon?: string;
    sortOrder: number;
    status: number;
  };

  /** 资产变更记录 */
  type AssetChange = {
    id: number;
    assetType: string;
    assetId: number;
    assetName?: string;
    fieldName: string;
    oldValue?: string;
    newValue?: string;
    changeType: 'create' | 'update' | 'delete';
    operator?: string;
    operatorId?: number;
    operateTime: string;
    remarks?: string;
  };

  /** 服务器查询参数 */
  type ServerQuery = {
    hostname?: string;
    ip?: string;
    env?: ServerEnv;
    status?: ServerStatus;
    businessId?: number;
    provider?: ServerProvider;
    groupId?: number;
    page?: number;
    pageSize?: number;
  };

  /** 服务器表单 */
  type ServerForm = {
    id?: number;
    hostname: string;
    ip: string;
    innerIp?: string;
    credentialIds: number[]; // 用户连接凭证（多选，credential_type=user）
    systemCredentialId?: number; // 系统运维凭证（单选，credential_type=system）
    serverType: ServerType;
    groupIds?: number[];
    tagIds?: number[];
    roomId?: number;
    cabinetId?: number;
    sshPort?: number;
    remarks?: string;
    cloudInfo?: CloudServerForm | null;
    env?: ServerEnv;
    cpu?: number;
    memory?: number;
    disk?: number;
    os?: string;
    osVersion?: string;
    businessId?: number;
  };

  /** 云主机表单 */
  type CloudServerForm = {
    serverId?: number;
    provider: CloudProvider;
    instanceId?: string;
    instanceName?: string;
    instanceType?: string;
    region?: string;
    zone?: string;
    vpcId?: string;
    subnetId?: string;
    publicIp?: string;
    privateIp?: string;
    chargeType?: string;
  };

  /** SSH凭证用途类型 */
  type CredentialType = 'user' | 'system';

  /** SSH凭证表单 */
  type SSHCredentialForm = {
    id?: number;
    name: string;
    description?: string;
    username: string;
    authType: 'password' | 'key';
    password?: string;
    privateKey?: string;
    passphrase?: string;
    credentialType: CredentialType;
    status?: number;
  };

  /** SSH凭证 */
  type SSHCredential = {
    id: number;
    name: string;
    description?: string;
    username: string;
    authType: 'password' | 'key';
    port: number;
    credentialType: CredentialType;
    sortOrder: number;
    status: number;
    createdAt: string;
    updatedAt: string;
  };

  /** 云主机 */
  type CloudServer = {
    id: number;
    serverId: number;
    provider: string;
    instanceId?: string;
    instanceName?: string;
    instanceType?: string;
    region?: string;
    zone?: string;
    vpcId?: string;
    subnetId?: string;
    publicIp?: string;
    privateIp?: string;
    securityGroups?: string;
    chargeType?: string;
    createdAt: string;
    updatedAt: string;
    server?: Server;
  };

  /** 业务系统表单 */
  type BusinessUnitForm = {
    id?: number;
    name: string;
    code: string;
    parentId: number;
    owner?: string;
    phone?: string;
    email?: string;
    sortOrder: number;
    status: number;
    remarks?: string;
  };

  /** 机房表单 */
  type ServerRoomForm = {
    id?: number;
    name: string;
    code: string;
    location?: string;
    address?: string;
    provider?: string;
    contact?: string;
    phone?: string;
    status: number;
    remarks?: string;
  };

  /** 标签表单 */
  type ServerTagForm = {
    id?: number;
    name: string;
    color: string;
    description?: string;
    sortOrder: number;
    status: number;
  };

  /** 服务器统计 */
  type ServerStats = {
    total: number;
    online: number;
    byEnv: Record<string, number>;
    byStatus: Record<string, number>;
    byProvider: Record<string, number>;
  };

  /** 分页响应 */
  type PageResponse<T> = {
    list: T[];
    total: number;
  };
}

/** 堡垒机相关类型 */
declare namespace Bastion {
  /** 会话状态 */
  type SessionStatus = 'active' | 'closed' | 'error' | 'terminated';

  /** 连接协议 */
  type Protocol = 'ssh' | 'sftp';

  /** 风险等级 */
  type RiskLevel = 'safe' | 'low' | 'medium' | 'high' | 'critical';

  /** 传输方向 */
  type TransferDirection = 'upload' | 'download';

  /** 传输状态 */
  type TransferStatus = 'pending' | 'transferring' | 'success' | 'failed';

  /** 审批状态 */
  type ApprovalStatus = 'pending' | 'approved' | 'rejected' | 'expired';

  /** 堡垒机会话 */
  type BastionSession = {
    id: number;
    serverId: number;
    server?: CMDB.Server;
    userId: number;
    user?: {
      id: number;
      username: string;
    };
    username: string;
    loginAccount: string;
    clientIp?: string;
    protocol: Protocol;
    sshCredentialId?: number;
    sshCredential?: CMDB.SSHCredential;
    startedAt?: string;
    endedAt?: string;
    duration: number;
    status: SessionStatus;
    closeReason?: string;
    createdAt: string;
    // 关联数据
    commands?: BastionCommand[];
    fileTransfers?: BastionFileTransfer[];
  };

  /** 命令审计 */
  type BastionCommand = {
    id: number;
    sessionId: number;
    session?: BastionSession;
    command: string;
    executedAt?: string;
    exitCode?: number;
    riskLevel: RiskLevel;
    blocked: boolean;
    outputSummary?: string;
    createdAt: string;
  };

  /** 文件传输审计 */
  type BastionFileTransfer = {
    id: number;
    sessionId: number;
    session?: BastionSession;
    direction: TransferDirection;
    remotePath: string;
    localPath?: string;
    fileSize: number;
    status: TransferStatus;
    errorMessage?: string;
    startedAt?: string;
    completedAt?: string;
    createdAt: string;
  };

  /** 访问策略 */
  type AccessPolicy = {
    id: number;
    name: string;
    subjectType: 'user' | 'role' | 'user_group';
    subjectId: number;
    assetScopeType: 'server' | 'group' | 'business' | 'tag' | 'all';
    assetScopeId: number;
    loginAccounts?: string[];
    protocols?: Protocol[];
    allowFileTransfer: boolean;
    allowSudo: boolean;
    requireApproval: boolean;
    timeWindow?: {
      start: string;
      end: string;
      days: number[];
    };
    highRiskCommands?: string[];
    status: number;
    createdAt: string;
    updatedAt: string;
  };

  /** 访问策略表单 */
  type AccessPolicyForm = {
    id?: number;
    name: string;
    subjectType: 'user' | 'role' | 'user_group';
    subjectId: number | number[];
    assetScopeType: 'server' | 'group' | 'business' | 'tag' | 'all';
    assetScopeId: number | number[];
    loginAccounts?: string[];
    protocols?: Protocol[];
    allowFileTransfer?: boolean;
    allowSudo?: boolean;
    requireApproval?: boolean;
    timeWindow?: {
      start: string;
      end: string;
      days: number[];
    };
    highRiskCommands?: string[];
    status?: number;
  };

  /** 连接请求 */
  type ConnectRequest = {
    protocol: Protocol;
    credentialId: number;
  };

  /** 连接响应 */
  type ConnectResponse = {
    sessionId: number;
    websocketUrl: string;
    serverName: string;
    serverIp: string;
  };
}

/** ======================================== */
/** Agent 版本管理相关类型                */
/** ======================================== */

/** Agent 版本 */
type AgentVersion = {
  id: number;
  version: string;
  releaseNotes?: string;
  changelog?: string;
  releasedAt: string;
  amd64BinaryPath?: string;
  amd64BinaryHash?: string;
  amd64BinarySize?: number;
  arm64BinaryPath?: string;
  arm64BinaryHash?: string;
  arm64BinarySize?: number;
  isLatest: boolean;
  isDeprecated: boolean;
  features?: string;
  minCompatibleVersion?: string;
  maxCompatibleVersion?: string;
  downloadCount: number;
  deployCount: number;
  createdAt: string;
  updatedAt: string;
};

/** Agent 版本表单 */
type AgentVersionForm = {
  id?: number;
  version: string;
  releaseNotes?: string;
  changelog?: string;
  amd64BinaryPath?: string;
  arm64BinaryPath?: string;
  isLatest?: boolean;
  isDeprecated?: boolean;
  features?: Record<string, boolean>;
  minCompatibleVersion?: string;
  maxCompatibleVersion?: string;
};

/** Agent 升级任务 */
type AgentUpgradeTask = {
  id: number;
  taskName?: string;
  targetVersion: string;
  targetServerIds?: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  currentStep: number;
  totalSteps: number;
  totalCount: number;
  successCount: number;
  failedCount: number;
  skippedCount: number;
  startedAt?: string;
  completedAt?: string;
  errorMessage?: string;
  operationLog?: string;
  createdBy?: string;
  createdAt: string;
  updatedAt: string;
};

/** Agent 升级请求 */
type AgentUpgradeRequest = {
  targetVersion: string;
};
