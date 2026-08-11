declare namespace K8s {
  /** 集群 */
  type Cluster = {
    id: number;
    name: string;
    description: string;
    endpoint: string;
    clusterType: string;
    region: string;
    version: string;
    nodeCount: number;
    status: number;
    createdAt: string;
    updatedAt: string;
  };

  /** 集群表单 */
  type ClusterForm = {
    name: string;
    description?: string;
    endpoint: string;
    kubeconfig: string;
    clusterType?: string;
    region?: string;
    nodeCount?: number;
  };

  /** 集群查询参数 */
  type ClusterQuery = {
    page?: number;
    pageSize?: number;
    name?: string;
    clusterType?: string;
    status?: number;
  };

  /** 节点 */
  type Node = {
    name: string;
    status: string;
    roles: string[];
    version: string;
    created: string;
    capacity?: {
      cpu: string;
      memory: string;
    };
  };

  /** 命名空间 */
  type Namespace = {
    name: string;
    status: string;
    created: string;
    labels?: Record<string, string>;
  };

  /** 集群用户 */
  type ClusterUser = {
    user_id: number;
    username: string;
    nickname: string;
    role_id: number;
    role_name: string;
    created_at: string;
  };

  /** 工作负载条件 */
  type WorkloadCondition = {
    type: string;
    status: string;
    reason: string;
    message: string;
  };

  /** Deployment */
  type Deployment = {
    name: string;
    namespace: string;
    replicas: number;
    ready: number;
    upToDate: number;
    available: number;
    age: string;
    labels?: Record<string, string>;
    conditions?: WorkloadCondition[];
  };

  /** Pod */
  type Pod = {
    name: string;
    namespace: string;
    status: string;
    phase: string;
    ip: string;
    node: string;
    age: string;
    labels?: Record<string, string>;
    restarts: number;
    containers?: Container[];
  };

  /** Pod 容器 */
  type Container = {
    name: string;
    image: string;
    status: string;
    ready: boolean;
    restartCount: number;
  };

  /** Service 端点 */
  type ServiceEndpoint = {
    ip: string;
    hostname?: string;
    nodeName?: string;
    ready?: boolean;
    port?: number;
  };

  /** Service 端口 */
  type ServicePort = {
    name: string;
    protocol: string;
    port: number;
    targetPort: string;
    nodePort?: number;
  };

  /** Service */
  type Service = {
    name: string;
    namespace: string;
    type: string;
    clusterIP: string;
    externalIP: string[];
    ports: ServicePort[];
    age: string;
    selector?: Record<string, string>;
  };

  /** Ingress */
  type Ingress = {
    name: string;
    namespace: string;
    hosts: string[];
    addresses: string[];
    ports: string[];
    age: string;
    annotations?: Record<string, string>;
    ingressClassName?: string;
    rules?: IngressRule[];
    tls?: IngressTLS[];
  };

  /** Ingress 路径 */
  type IngressPath = {
    path: string;
    serviceName: string;
    servicePort: string | number;
  };

  /** Ingress 规则 */
  type IngressRule = {
    host: string;
    http?: {
      paths: IngressPath[];
    };
  };

  /** Ingress TLS */
  type IngressTLS = {
    hosts: string[];
    secretName: string;
  };

  /** ConfigMap */
  type ConfigMap = {
    name: string;
    namespace: string;
    age: string;
    labels?: Record<string, string>;
    dataKeys: string[];
  };

  /** Secret */
  type Secret = {
    name: string;
    namespace: string;
    type: string;
    age: string;
    labels?: Record<string, string>;
    dataKeys: string[];
  };

  /** StatefulSet */
  type StatefulSet = {
    name: string;
    namespace: string;
    replicas: number;
    ready: number;
    upToDate: number;
    available: number;
    age: string;
    labels?: Record<string, string>;
    conditions?: WorkloadCondition[];
  };

  /** DaemonSet */
  type DaemonSet = {
    name: string;
    namespace: string;
    desired: number;
    current: number;
    ready: number;
    available: number;
    age: string;
    labels?: Record<string, string>;
    conditions?: WorkloadCondition[];
  };

  /** Job */
  type Job = {
    name: string;
    namespace: string;
    completions: number;
    duration: string;
    age: string;
    labels?: Record<string, string>;
    conditions?: WorkloadCondition[];
  };

  /** CronJob */
  type CronJob = {
    name: string;
    namespace: string;
    schedule: string;
    suspend: boolean;
    active: number;
    lastSchedule: string;
    age: string;
    labels?: Record<string, string>;
  };

  /** 终端会话 */
  type TerminalSession = {
    session_id: number;
    cluster_id: number;
    namespace: string;
    pod_name: string;
    container: string;
    last_activity: string;
  };

  /** 事件 */
  type Event = {
    type: string;
    reason: string;
    message: string;
    source: string;
    count: number;
    firstTimestamp: string;
    lastTimestamp: string;
  };

  /** 分页响应 */
  type PaginatedResponse<T> = {
    list: T[];
    total: number;
  };

  /** 资源删除参数 */
  type ResourceDeleteParams = {
    namespace: string;
    name: string;
  };

  /** 资源更新参数 */
  type ResourceUpdateParams = {
    namespace: string;
    manifest: unknown;
  };

  /** Deployment 缩放参数 */
  type DeploymentScaleParams = {
    namespace: string;
    name: string;
    replicas: number;
  };

  /** CronJob 暂停/恢复参数 */
  type CronJobSuspendParams = {
    namespace: string;
    name: string;
    suspend: boolean;
  };

  /** 工作负载行类型（表格行共用） */
  type WorkloadRow = Deployment | StatefulSet | DaemonSet | Job | CronJob | Pod;
}
