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
    sshUser: string;
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
    // 关联数据
    credential?: SSHCredential;
    cabinet?: Cabinet;
    tags?: ServerTag[];
    groups?: ServerGroup[];
    cloudInfo?: CloudServer;
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
    credentialId: number;
    serverType: ServerType;
    groupIds?: number[];
    sshPort?: number;
    remarks?: string;
    cloudInfo?: CloudServerForm | null;
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
