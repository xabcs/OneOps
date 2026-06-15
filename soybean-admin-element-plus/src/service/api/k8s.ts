import { request } from '../request';

// ========== Types ==========

export namespace K8s {
  export interface Cluster {
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
  }

  export interface ClusterForm {
    name: string;
    description?: string;
    endpoint: string;
    kubeconfig: string;
    clusterType?: string;
    region?: string;
    nodeCount?: number;
  }

  export interface ClusterQuery {
    page?: number;
    pageSize?: number;
    name?: string;
    clusterType?: string;
    status?: number;
  }

  export interface Node {
    name: string;
    status: string;
    roles: string[];
    version: string;
    created: string;
    capacity?: {
      cpu: string;
      memory: string;
    };
  }

  export interface Namespace {
    name: string;
    status: string;
    created: string;
    labels?: Record<string, string>;
  }

  export interface ClusterUser {
    user_id: number;
    username: string;
    nickname: string;
    role_id: number;
    role_name: string;
    created_at: string;
  }

  export interface Deployment {
    name: string;
    namespace: string;
    replicas: number;
    ready: number;
    upToDate: number;
    available: number;
    age: string;
    labels?: Record<string, string>;
    conditions?: Array<{
      type: string;
      status: string;
      reason: string;
      message: string;
    }>;
  }

  export interface Pod {
    name: string;
    namespace: string;
    status: string;
    phase: string;
    ip: string;
    node: string;
    age: string;
    labels?: Record<string, string>;
    restarts: number;
  }

  export interface Service {
    name: string;
    namespace: string;
    type: string;
    clusterIP: string;
    externalIP: string[];
    ports: Array<{
      name: string;
      protocol: string;
      port: number;
      targetPort: string;
      nodePort?: number;
    }>;
    age: string;
    selector?: Record<string, string>;
  }

  export interface ConfigMap {
    name: string;
    namespace: string;
    age: string;
    labels?: Record<string, string>;
    dataKeys: string[];
  }

  export interface Secret {
    name: string;
    namespace: string;
    type: string;
    age: string;
    labels?: Record<string, string>;
    dataKeys: string[];
  }

  export interface TerminalSession {
    session_id: number;
    cluster_id: number;
    namespace: string;
    pod_name: string;
    container: string;
    last_activity: string;
  }
}

// ========== Cluster Management ==========

/**
 * 获取集群列表
 */
export function fetchK8sClusters(params?: K8s.ClusterQuery) {
  return request<K8s.Cluster[]>({
    url: '/k8s/clusters',
    method: 'get',
    params,
    transform: (response: any) => {
      // 返回完整的响应对象，包含 data 和 total
      return response.data;
    }
  });
}

/**
 * 获取集群详情
 */
export function getK8sClusterDetail(id: number) {
  return request<K8s.Cluster>({
    url: `/k8s/clusters/${id}`,
    method: 'get'
  });
}

/**
 * 创建集群
 */
export function createK8sCluster(data: K8s.ClusterForm) {
  return request({
    url: '/k8s/clusters',
    method: 'post',
    data
  });
}

/**
 * 更新集群
 */
export function updateK8sCluster(id: number, data: Partial<K8s.ClusterForm>) {
  return request({
    url: `/k8s/clusters/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除集群
 */
export function deleteK8sCluster(id: number, data: { confirmName: string }) {
  return request({
    url: `/k8s/clusters/${id}`,
    method: 'delete',
    data
  });
}

/**
 * 测试集群连接
 */
export function testK8sConnection(id: number) {
  return request({
    url: `/k8s/clusters/${id}/test`,
    method: 'post'
  });
}

/**
 * 获取集群节点列表
 */
export function fetchK8sClusterNodes(clusterId: number) {
  return request<K8s.Node[]>({
    url: `/k8s/clusters/${clusterId}/nodes`,
    method: 'get'
  });
}

/**
 * 获取集群命名空间列表
 */
export function fetchK8sClusterNamespaces(clusterId: number) {
  return request<K8s.Namespace[]>({
    url: `/k8s/clusters/${clusterId}/namespaces`,
    method: 'get'
  });
}

/**
 * 获取集群用户列表
 */
export function fetchK8sClusterUsers(clusterId: number) {
  return request<K8s.ClusterUser[]>({
    url: `/k8s/clusters/${clusterId}/users`,
    method: 'get'
  });
}

// ========== Permission Management ==========

/**
 * 分配集群角色
 */
export function assignK8sClusterRole(clusterId: number, data: { userId: number; roleId: number }) {
  return request({
    url: `/k8s/clusters/${clusterId}/permissions`,
    method: 'post',
    data
  });
}

/**
 * 撤销集群角色
 */
export function revokeK8sClusterRole(clusterId: number, userId: number) {
  return request({
    url: `/k8s/clusters/${clusterId}/permissions/${userId}`,
    method: 'delete'
  });
}

// ========== Workloads - Deployments ==========

/**
 * 获取 Deployment 列表
 */
export function fetchK8sDeployments(clusterId: number, params?: { namespace?: string }) {
  return request<K8s.Deployment[]>({
    url: `/k8s/clusters/${clusterId}/deployments`,
    method: 'get',
    params
  });
}

/**
 * 获取 Deployment 详情
 */
export function getK8sDeployment(clusterId: number, namespace: string, name: string) {
  return request<K8s.Deployment & { manifest: string; images: string[] }>({
    url: `/k8s/clusters/${clusterId}/deployments/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 创建 Deployment
 */
export function createK8sDeployment(clusterId: number, data: { namespace: string; manifest: Record<string, any> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/deployments`,
    method: 'post',
    data
  });
}

/**
 * 更新 Deployment
 */
export function updateK8sDeployment(clusterId: number, data: { namespace: string; manifest: Record<string, any> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/deployments`,
    method: 'put',
    data
  });
}

/**
 * 删除 Deployment
 */
export function deleteK8sDeployment(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/deployments`,
    method: 'delete',
    data
  });
}

/**
 * 扩缩容 Deployment
 */
export function scaleK8sDeployment(clusterId: number, data: { namespace: string; name: string; replicas: number }) {
  return request({
    url: `/k8s/clusters/${clusterId}/deployments/scale`,
    method: 'post',
    data
  });
}

/**
 * 重启 Deployment
 */
export function restartK8sDeployment(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/deployments/restart`,
    method: 'post',
    data
  });
}

// ========== Pods ==========

/**
 * 获取 Pod 列表
 */
export function fetchK8sPods(clusterId: number, params?: { namespace?: string; labelSelector?: string }) {
  return request<K8s.Pod[]>({
    url: `/k8s/clusters/${clusterId}/pods`,
    method: 'get',
    params
  });
}

/**
 * 获取 Pod 详情
 */
export function getK8sPod(clusterId: number, namespace: string, name: string) {
  return request<K8s.Pod & { manifest: string; images: string[]; containers: any[] }>({
    url: `/k8s/clusters/${clusterId}/pods/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 获取 Pod 日志
 */
export function fetchK8sPodLogs(
  clusterId: number,
  namespace: string,
  name: string,
  params?: { container?: string; tailLines?: number }
) {
  return request<{ logs: string }>({
    url: `/k8s/clusters/${clusterId}/pods/${namespace}/${name}/logs`,
    method: 'get',
    params
  });
}

/**
 * 删除 Pod
 */
export function deleteK8sPod(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/pods`,
    method: 'delete',
    data
  });
}

// ========== Services ==========

/**
 * 获取 Service 列表
 */
export function fetchK8sServices(clusterId: number, params?: { namespace?: string }) {
  return request<K8s.Service[]>({
    url: `/k8s/clusters/${clusterId}/services`,
    method: 'get',
    params
  });
}

/**
 * 获取 Service 详情
 */
export function getK8sService(clusterId: number, namespace: string, name: string) {
  return request<K8s.Service & { manifest: string; endpoints: any[] }>({
    url: `/k8s/clusters/${clusterId}/services/${namespace}/${name}`,
    method: 'get'
  });
}

// ========== ConfigMaps ==========

/**
 * 获取 ConfigMap 列表
 */
export function fetchK8sConfigMaps(clusterId: number, params?: { namespace?: string }) {
  return request<K8s.ConfigMap[]>({
    url: `/k8s/clusters/${clusterId}/configmaps`,
    method: 'get',
    params
  });
}

/**
 * 获取 ConfigMap 详情
 */
export function getK8sConfigMap(clusterId: number, namespace: string, name: string) {
  return request<K8s.ConfigMap & { manifest: string; data: Record<string, string> }>({
    url: `/k8s/clusters/${clusterId}/configmaps/${namespace}/${name}`,
    method: 'get'
  });
}

// ========== Secrets ==========

/**
 * 获取 Secret 列表
 */
export function fetchK8sSecrets(clusterId: number, params?: { namespace?: string }) {
  return request<K8s.Secret[]>({
    url: `/k8s/clusters/${clusterId}/secrets`,
    method: 'get',
    params
  });
}

/**
 * 获取 Secret 详情
 */
export function getK8sSecret(clusterId: number, namespace: string, name: string) {
  return request<K8s.Secret & { manifest: string; data: Record<string, string> }>({
    url: `/k8s/clusters/${clusterId}/secrets/${namespace}/${name}`,
    method: 'get'
  });
}

// ========== Terminal ==========

/**
 * 获取活跃的终端会话
 */
export function fetchK8sActiveTerminalSessions() {
  return request<K8s.TerminalSession[]>({
    url: '/k8s/terminal/active',
    method: 'get'
  });
}

/**
 * 终止终端会话
 */
export function terminateK8sTerminalSession(sessionId: number) {
  return request({
    url: `/k8s/terminal/sessions/${sessionId}/terminate`,
    method: 'post'
  });
}
