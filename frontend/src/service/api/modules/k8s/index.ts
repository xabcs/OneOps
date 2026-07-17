/**
 * K8s 模块 API 服务
 * 统一管理所有 Kubernetes 相关的 API 调用
 */

import { request } from '@/service/request';

/**
 * 集群相关 API
 */

// 获取 K8s 集群列表
export function fetchGetClusters(params?: {
  page?: number;
  pageSize?: number;
  name?: string;
  status?: string;
  clusterType?: string;
}) {
  return request<K8s.ClusterPageResult>({
    url: '/api/v1/k8s/clusters',
    method: 'get',
    params
  });
}

// 获取集群详情
export function fetchGetClusterById(id: number) {
  return request<K8s.Cluster>({
    url: `/api/v1/k8s/clusters/${id}`,
    method: 'get'
  });
}

// 创建集群
export function fetchCreateCluster(data: K8s.ClusterFormData) {
  return request<{ id: number }>({
    url: '/api/v1/k8s/clusters',
    method: 'post',
    data
  });
}

// 更新集群
export function fetchUpdateCluster(id: number, data: Partial<K8s.ClusterFormData>) {
  return request({
    url: `/api/v1/k8s/clusters/${id}`,
    method: 'put',
    data
  });
}

// 删除集群
export function fetchDeleteCluster(id: number, confirmName: string) {
  return request({
    url: `/api/v1/k8s/clusters/${id}`,
    method: 'delete',
    data: { confirmName }
  });
}

/**
 * Deployments 相关 API
 */

// 获取 Deployment 列表
export function fetchListDeployments(params: {
  clusterId: number;
  namespace?: string;
  page?: number;
  pageSize?: number;
}) {
  return request<K8s.DeploymentPageResult>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/deployments`,
    method: 'get',
    params: {
      namespace: params.namespace,
      page: params.page,
      pageSize: params.pageSize
    }
  });
}

// 获取 Deployment 详情
export function fetchGetDeployment(params: { clusterId: number; namespace: string; name: string }) {
  return request<K8s.Deployment>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/deployments/${params.namespace}/${params.name}`,
    method: 'get'
  });
}

// 获取 Deployment 的 Pods
export function fetchGetDeploymentPods(params: { clusterId: number; namespace: string; name: string }) {
  return request<K8s.Pod[]>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/deployments/${params.namespace}/${params.name}/pods`,
    method: 'get'
  });
}

// 创建 Deployment
export function fetchCreateDeployment(data: {
  clusterId: number;
  namespace: string;
  name: string;
  replicas: number;
  image: string;
}) {
  return request<K8s.Deployment>({
    url: `/api/v1/k8s/clusters/${data.clusterId}/deployments`,
    method: 'post',
    data: {
      namespace: data.namespace,
      name: data.name,
      replicas: data.replicas,
      image: data.image
    }
  });
}

// 更新 Deployment
export function fetchUpdateDeployment(params: {
  clusterId: number;
  namespace: string;
  name: string;
  replicas?: number;
  image?: string;
}) {
  return request<K8s.Deployment>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/deployments/${params.namespace}/${params.name}`,
    method: 'put',
    data: {
      replicas: params.replicas,
      image: params.image
    }
  });
}

// 删除 Deployment
export function fetchDeleteDeployment(params: { clusterId: number; namespace: string; name: string }) {
  return request({
    url: `/api/v1/k8s/clusters/${params.clusterId}/deployments/${params.namespace}/${params.name}`,
    method: 'delete'
  });
}

// 扩缩容 Deployment
export function fetchScaleDeployment(data: { clusterId: number; namespace: string; name: string; replicas: number }) {
  return request({
    url: `/api/v1/k8s/clusters/${data.clusterId}/deployments/${data.namespace}/${data.name}/scale`,
    method: 'post',
    data: { replicas: data.replicas }
  });
}

// 重启 Deployment
export function fetchRestartDeployment(params: { clusterId: number; namespace: string; name: string }) {
  return request({
    url: `/api/v1/k8s/clusters/${params.clusterId}/deployments/${params.namespace}/${params.name}/restart`,
    method: 'post'
  });
}

/**
 * Pods 相关 API
 */

// 获取 Pod 列表
export function fetchListPods(params: { clusterId: number; namespace?: string; page?: number; pageSize?: number }) {
  return request<K8s.PodPageResult>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/pods`,
    method: 'get',
    params: {
      namespace: params.namespace,
      page: params.page,
      pageSize: params.pageSize
    }
  });
}

// 获取 Pod 详情
export function fetchGetPod(params: { clusterId: number; namespace: string; name: string }) {
  return request<K8s.Pod>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/pods/${params.namespace}/${params.name}`,
    method: 'get'
  });
}

// 获取 Pod 日志
export function fetchGetPodLogs(params: { clusterId: number; namespace: string; name: string; tailLines?: string }) {
  return request<string>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/pods/${params.namespace}/${params.name}/logs`,
    method: 'get',
    params: { tailLines: params.tailLines || '100' }
  });
}

// 删除 Pod
export function fetchDeletePod(params: { clusterId: number; namespace: string; name: string }) {
  return request({
    url: `/api/v1/k8s/clusters/${params.clusterId}/pods/${params.namespace}/${params.name}`,
    method: 'delete'
  });
}

/**
 * Services 相关 API
 */

// 获取 Service 列表
export function fetchListServices(params: { clusterId: number; namespace?: string; page?: number; pageSize?: number }) {
  return request<K8s.ServicePageResult>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/services`,
    method: 'get',
    params: {
      namespace: params.namespace,
      page: params.page,
      pageSize: params.pageSize
    }
  });
}

// 获取 Service 详情
export function fetchGetService(params: { clusterId: number; namespace: string; name: string }) {
  return request<K8s.Service>({
    url: `/api/v1/k8s/clusters/${params.clusterId}/services/${params.namespace}/${params.name}`,
    method: 'get'
  });
}

// 创建 Service
export function fetchCreateService(data: {
  clusterId: number;
  namespace: string;
  name: string;
  type: string;
  ports: Array<{
    name: string;
    protocol: string;
    port: number;
    targetPort: number;
  }>;
}) {
  return request<K8s.Service>({
    url: `/api/v1/k8s/clusters/${data.clusterId}/services`,
    method: 'post',
    data: {
      namespace: data.namespace,
      name: data.name,
      type: data.type,
      ports: data.ports
    }
  });
}

// 删除 Service
export function fetchDeleteService(params: { clusterId: number; namespace: string; name: string }) {
  return request({
    url: `/api/v1/k8s/clusters/${params.clusterId}/services/${params.namespace}/${params.name}`,
    method: 'delete'
  });
}

/**
 * 权限相关 API
 */

// 获取集群权限列表
export function fetchGetClusterPermissions() {
  return request<K8s.ClusterPermission[]>({
    url: '/api/v1/k8s/permissions/clusters',
    method: 'get'
  });
}

// 获取用户集群权限
export function fetchGetUserClusters(userId: number) {
  return request<K8s.Cluster[]>({
    url: `/api/v1/k8s/permissions/users/${userId}/clusters`,
    method: 'get'
  });
}

// 分配用户集群权限
export function fetchAssignUserClusters(userId: number, clusterIds: number[]) {
  return request({
    url: `/api/v1/k8s/permissions/users/${userId}/clusters`,
    method: 'post',
    data: { clusterIds }
  });
}

// 移除用户集群权限
export function fetchRemoveUserCluster(userId: number, clusterId: number) {
  return request({
    url: `/api/v1/k8s/permissions/users/${userId}/clusters/${clusterId}`,
    method: 'delete'
  });
}

// 导出索引
export default {
  // 集群
  fetchGetClusters,
  fetchGetClusterById,
  fetchCreateCluster,
  fetchUpdateCluster,
  fetchDeleteCluster,

  // Deployments
  fetchListDeployments,
  fetchGetDeployment,
  fetchGetDeploymentPods,
  fetchCreateDeployment,
  fetchUpdateDeployment,
  fetchDeleteDeployment,
  fetchScaleDeployment,
  fetchRestartDeployment,

  // Pods
  fetchListPods,
  fetchGetPod,
  fetchGetPodLogs,
  fetchDeletePod,

  // Services
  fetchListServices,
  fetchGetService,
  fetchCreateService,
  fetchDeleteService,

  // 权限
  fetchGetClusterPermissions,
  fetchGetUserClusters,
  fetchAssignUserClusters,
  fetchRemoveUserCluster
};
