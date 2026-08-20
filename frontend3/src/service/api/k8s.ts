import { request } from '../request';

// ========== Cluster Management ==========

/**
 * 获取集群列表
 */
export function fetchK8sClusters(params?: K8s.ClusterQuery) {
  return request<{
    list: K8s.Cluster[];
    total: number;
    page: number;
    pageSize: number;
  }>({
    url: '/k8s/clusters',
    method: 'get',
    params
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

// ========== Workloads - Deployments ==========

/**
 * 获取 Deployment 列表（支持分页）
 */
export function fetchK8sDeployments(
  clusterId: number,
  params?: { namespace?: string; page?: number; pageSize?: number }
) {
  return request<K8s.PaginatedResponse<K8s.Deployment>>({
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
 * 获取 Deployment 管理的 Pods
 */
export function getK8sDeploymentPods(clusterId: number, namespace: string, name: string) {
  return request<K8s.Pod[]>({
    url: `/k8s/clusters/${clusterId}/deployments/${namespace}/${name}/pods`,
    method: 'get'
  });
}

/**
 * 创建 Deployment
 */
export function createK8sDeployment(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/deployments`,
    method: 'post',
    data
  });
}

/**
 * 更新 Deployment
 */
export function updateK8sDeployment(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
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
 * 获取 Pod 列表（支持分页）
 */
export function fetchK8sPods(
  clusterId: number,
  params?: { namespace?: string; labelSelector?: string; page?: number; pageSize?: number }
) {
  return request<K8s.PaginatedResponse<K8s.Pod>>({
    url: `/k8s/clusters/${clusterId}/pods`,
    method: 'get',
    params
  });
}

/**
 * 获取 Pod 详情
 */
export function getK8sPod(clusterId: number, namespace: string, name: string) {
  return request<K8s.Pod & { manifest: string; images: string[]; containers: K8s.Container[] }>({
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

/**
 * 更新 Pod
 */
export function updateK8sPod(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/pods`,
    method: 'put',
    data
  });
}

// ========== Services ==========

/**
 * 获取 Service 列表（支持分页）
 */
export function fetchK8sServices(clusterId: number, params?: { namespace?: string; page?: number; pageSize?: number }) {
  return request<K8s.PaginatedResponse<K8s.Service>>({
    url: `/k8s/clusters/${clusterId}/services`,
    method: 'get',
    params
  });
}

/**
 * 获取 Service 详情
 */
export function getK8sService(clusterId: number, namespace: string, name: string) {
  return request<K8s.Service & { manifest: string; endpoints: K8s.ServiceEndpoint[] }>({
    url: `/k8s/clusters/${clusterId}/services/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 更新 Service
 */
export function updateK8sService(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/services`,
    method: 'put',
    data
  });
}

/**
 * 删除 Service
 */
export function deleteK8sService(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/services`,
    method: 'delete',
    data
  });
}

// ========== Ingresses ==========

/**
 * 获取 Ingress 列表（支持分页）
 */
export function fetchK8sIngresses(
  clusterId: number,
  params?: { namespace?: string; page?: number; pageSize?: number }
) {
  return request<K8s.PaginatedResponse<K8s.Ingress>>({
    url: `/k8s/clusters/${clusterId}/ingresses`,
    method: 'get',
    params
  });
}

/**
 * 获取 Ingress 详情
 */
export function getK8sIngress(clusterId: number, namespace: string, name: string) {
  return request<K8s.Ingress & { manifest: string }>({
    url: `/k8s/clusters/${clusterId}/ingresses/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 创建 Ingress
 */
export function createK8sIngress(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/ingresses`,
    method: 'post',
    data
  });
}

/**
 * 更新 Ingress
 */
export function updateK8sIngress(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/ingresses`,
    method: 'put',
    data
  });
}

/**
 * 删除 Ingress
 */
export function deleteK8sIngress(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/ingresses`,
    method: 'delete',
    data
  });
}

// ========== ConfigMaps ==========

/**
 * 获取 ConfigMap 列表（支持分页）
 */
export function fetchK8sConfigMaps(
  clusterId: number,
  params?: { namespace?: string; page?: number; pageSize?: number }
) {
  return request<K8s.PaginatedResponse<K8s.ConfigMap>>({
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

/**
 * 更新 ConfigMap
 */
export function updateK8sConfigMap(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/configmaps`,
    method: 'put',
    data
  });
}

/**
 * 删除 ConfigMap
 */
export function deleteK8sConfigMap(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/configmaps`,
    method: 'delete',
    data
  });
}

// ========== Secrets ==========

/**
 * 获取 Secret 列表（支持分页）
 */
export function fetchK8sSecrets(clusterId: number, params?: { namespace?: string; page?: number; pageSize?: number }) {
  return request<K8s.PaginatedResponse<K8s.Secret>>({
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

/**
 * 更新 Secret
 */
export function updateK8sSecret(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/secrets`,
    method: 'put',
    data
  });
}

/**
 * 删除 Secret
 */
export function deleteK8sSecret(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/secrets`,
    method: 'delete',
    data
  });
}

// ========== Events ==========

/**
 * 获取 Event 列表
 */
export function fetchK8sEvents(clusterId: number, namespace?: string, fieldSelector?: string) {
  return request<K8s.Event[]>({
    url: `/k8s/clusters/${clusterId}/events`,
    method: 'get',
    params: { namespace, fieldSelector }
  });
}

// ========== StatefulSets ==========

/**
 * 获取 StatefulSet 列表（支持分页）
 */
export function fetchK8sStatefulSets(
  clusterId: number,
  params?: { namespace?: string; page?: number; pageSize?: number }
) {
  return request<K8s.PaginatedResponse<K8s.StatefulSet>>({
    url: `/k8s/clusters/${clusterId}/statefulsets`,
    method: 'get',
    params
  });
}

/**
 * 获取 StatefulSet 详情
 */
export function getK8sStatefulSet(clusterId: number, namespace: string, name: string) {
  return request<K8s.StatefulSet & { manifest: string; images: string[] }>({
    url: `/k8s/clusters/${clusterId}/statefulsets/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 重启 StatefulSet
 */
export function restartK8sStatefulSet(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/statefulsets/restart`,
    method: 'post',
    data
  });
}

/**
 * 获取 StatefulSet 管理的 Pods
 */
export function getK8sStatefulSetPods(clusterId: number, namespace: string, name: string) {
  return request<K8s.Pod[]>({
    url: `/k8s/clusters/${clusterId}/statefulsets/${namespace}/${name}/pods`,
    method: 'get'
  });
}

/**
 * 删除 StatefulSet
 */
export function deleteK8sStatefulSet(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/statefulsets`,
    method: 'delete',
    data
  });
}

/**
 * 更新 StatefulSet
 */
export function updateK8sStatefulSet(
  clusterId: number,
  data: { namespace: string; manifest: Record<string, unknown> }
) {
  return request({
    url: `/k8s/clusters/${clusterId}/statefulsets`,
    method: 'put',
    data
  });
}

// ========== DaemonSets ==========

/**
 * 获取 DaemonSet 列表（支持分页）
 */
export function fetchK8sDaemonSets(
  clusterId: number,
  params?: { namespace?: string; page?: number; pageSize?: number }
) {
  return request<K8s.PaginatedResponse<K8s.DaemonSet>>({
    url: `/k8s/clusters/${clusterId}/daemonsets`,
    method: 'get',
    params
  });
}

/**
 * 获取 DaemonSet 详情
 */
export function getK8sDaemonSet(clusterId: number, namespace: string, name: string) {
  return request<K8s.DaemonSet & { manifest: string; images: string[] }>({
    url: `/k8s/clusters/${clusterId}/daemonsets/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 重启 DaemonSet
 */
export function restartK8sDaemonSet(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/daemonsets/restart`,
    method: 'post',
    data
  });
}

/**
 * 获取 DaemonSet 管理的 Pods
 */
export function getK8sDaemonSetPods(clusterId: number, namespace: string, name: string) {
  return request<K8s.Pod[]>({
    url: `/k8s/clusters/${clusterId}/daemonsets/${namespace}/${name}/pods`,
    method: 'get'
  });
}

/**
 * 删除 DaemonSet
 */
export function deleteK8sDaemonSet(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/daemonsets`,
    method: 'delete',
    data
  });
}

/**
 * 更新 DaemonSet
 */
export function updateK8sDaemonSet(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/daemonsets`,
    method: 'put',
    data
  });
}

// ========== Jobs ==========

/**
 * 获取 Job 列表（支持分页）
 */
export function fetchK8sJobs(clusterId: number, params?: { namespace?: string; page?: number; pageSize?: number }) {
  return request<K8s.PaginatedResponse<K8s.Job>>({
    url: `/k8s/clusters/${clusterId}/jobs`,
    method: 'get',
    params
  });
}

/**
 * 获取 Job 详情
 */
export function getK8sJob(clusterId: number, namespace: string, name: string) {
  return request<K8s.Job & { manifest: string; images: string[] }>({
    url: `/k8s/clusters/${clusterId}/jobs/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 获取 Job 管理的 Pods
 */
export function getK8sJobPods(clusterId: number, namespace: string, name: string) {
  return request<K8s.Pod[]>({
    url: `/k8s/clusters/${clusterId}/jobs/${namespace}/${name}/pods`,
    method: 'get'
  });
}

/**
 * 删除 Job
 */
export function deleteK8sJob(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/jobs`,
    method: 'delete',
    data
  });
}

/**
 * 更新 Job
 */
export function updateK8sJob(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/jobs`,
    method: 'put',
    data
  });
}

// ========== CronJobs ==========

/**
 * 获取 CronJob 列表（支持分页）
 */
export function fetchK8sCronJobs(clusterId: number, params?: { namespace?: string; page?: number; pageSize?: number }) {
  return request<K8s.PaginatedResponse<K8s.CronJob>>({
    url: `/k8s/clusters/${clusterId}/cronjobs`,
    method: 'get',
    params
  });
}

/**
 * 获取 CronJob 详情
 */
export function getK8sCronJob(clusterId: number, namespace: string, name: string) {
  return request<K8s.CronJob & { manifest: string; images: string[] }>({
    url: `/k8s/clusters/${clusterId}/cronjobs/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 暂停 CronJob
 */
export function suspendK8sCronJob(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/cronjobs/suspend`,
    method: 'post',
    data
  });
}

/**
 * 获取 CronJob 管理的 Pods
 */
export function getK8sCronJobPods(clusterId: number, namespace: string, name: string) {
  return request<K8s.Pod[]>({
    url: `/k8s/clusters/${clusterId}/cronjobs/${namespace}/${name}/pods`,
    method: 'get'
  });
}

/**
 * 删除 CronJob
 */
export function deleteK8sCronJob(clusterId: number, data: { namespace: string; name: string }) {
  return request({
    url: `/k8s/clusters/${clusterId}/cronjobs`,
    method: 'delete',
    data
  });
}

/**
 * 更新 CronJob
 */
export function updateK8sCronJob(clusterId: number, data: { namespace: string; manifest: Record<string, unknown> }) {
  return request({
    url: `/k8s/clusters/${clusterId}/cronjobs`,
    method: 'put',
    data
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

// ========== Native RBAC（A：集群内原生 RBAC 对象代管） ==========

/**
 * 获取集群内原生 ClusterRole 列表
 */
export function fetchK8sNativeClusterRoles(clusterId: number, search?: string) {
  return request<K8s.NativeRoleItem[]>({
    url: `/k8s/clusters/${clusterId}/rbac/clusterroles`,
    method: 'get',
    params: { search }
  });
}

/**
 * 获取集群内原生 ClusterRole 详情（含完整 rules）
 */
export function getK8sNativeClusterRole(clusterId: number, name: string) {
  return request<K8s.NativeRoleDetail>({
    url: `/k8s/clusters/${clusterId}/rbac/clusterroles/${name}`,
    method: 'get'
  });
}

/**
 * 更新集群内原生 ClusterRole（manifest 整体替换 rules）
 */
export function updateK8sNativeClusterRole(clusterId: number, manifest: Record<string, unknown>) {
  return request({
    url: `/k8s/clusters/${clusterId}/rbac/clusterroles`,
    method: 'put',
    data: manifest
  });
}

/**
 * 获取集群内原生 Role 列表（namespace 为空则全集群）
 */
export function fetchK8sNativeRoles(clusterId: number, params?: { namespace?: string; search?: string }) {
  return request<K8s.NativeRoleItem[]>({
    url: `/k8s/clusters/${clusterId}/rbac/roles`,
    method: 'get',
    params
  });
}

/**
 * 获取集群内原生 Role 详情
 */
export function getK8sNativeRole(clusterId: number, namespace: string, name: string) {
  return request<K8s.NativeRoleDetail>({
    url: `/k8s/clusters/${clusterId}/rbac/roles/${namespace}/${name}`,
    method: 'get'
  });
}

/**
 * 更新命名空间内原生 Role
 */
export function updateK8sNativeRole(
  clusterId: number,
  namespace: string,
  manifest: Record<string, unknown>
) {
  return request({
    url: `/k8s/clusters/${clusterId}/rbac/roles/${namespace}`,
    method: 'put',
    data: manifest
  });
}

/**
 * 获取集群内原生 ClusterRoleBinding 列表
 */
export function fetchK8sNativeClusterRoleBindings(clusterId: number, search?: string) {
  return request<K8s.NativeBindingItem[]>({
    url: `/k8s/clusters/${clusterId}/rbac/clusterrolebindings`,
    method: 'get',
    params: { search }
  });
}

/**
 * 获取集群内原生 ClusterRoleBinding 详情
 */
export function getK8sNativeClusterRoleBinding(clusterId: number, name: string) {
  return request<K8s.NativeBindingItem>({
    url: `/k8s/clusters/${clusterId}/rbac/clusterrolebindings/${name}`,
    method: 'get'
  });
}

/**
 * 获取集群内原生 RoleBinding 列表（namespace 为空则全集群）
 */
export function fetchK8sNativeRoleBindings(clusterId: number, params?: { namespace?: string; search?: string }) {
  return request<K8s.NativeBindingItem[]>({
    url: `/k8s/clusters/${clusterId}/rbac/rolebindings`,
    method: 'get',
    params
  });
}

/**
 * 获取集群内原生 RoleBinding 详情
 */
export function getK8sNativeRoleBinding(clusterId: number, namespace: string, name: string) {
  return request<K8s.NativeBindingItem>({
    url: `/k8s/clusters/${clusterId}/rbac/rolebindings/${namespace}/${name}`,
    method: 'get'
  });
}

// ========== 访问授权（B：OneOps 主体 × K8s 原生角色绑定；入口：K8s管理-安全管理-授权管理） ==========

/**
 * 获取全部集群的访问授权记录（授权管理页列表，可按集群筛选）
 */
export function fetchK8sAllNativeBindings(clusterId?: number) {
  return request<K8s.NativeRoleBinding[]>({
    url: '/k8s/native-bindings',
    method: 'get',
    params: clusterId ? { clusterId } : {}
  });
}

/**
 * 集群候选（授权管理页筛选与表单，权限归属 k8s.permission.list）
 */
export function fetchK8sClusterOptions() {
  return request<{ id: number; name: string; status: string }[]>({
    url: '/k8s/clusters/options',
    method: 'get'
  });
}

/**
 * 获取集群的访问授权绑定列表
 */
export function fetchK8sNativeBindings(clusterId: number) {
  return request<K8s.NativeRoleBinding[]>({
    url: `/k8s/clusters/${clusterId}/native-bindings`,
    method: 'get'
  });
}

/**
 * 创建访问授权（落库 + 集群内创建真实 Binding，主体经 impersonation 生效）
 */
export function assignK8sNativeBinding(clusterId: number, data: K8s.NativeRoleBindingForm) {
  return request({
    url: `/k8s/clusters/${clusterId}/native-bindings`,
    method: 'post',
    data
  });
}

/**
 * 撤销访问授权（删除集群内 Binding 与平台记录）
 */
export function revokeK8sNativeBinding(clusterId: number, bindingId: number) {
  return request({
    url: `/k8s/clusters/${clusterId}/native-bindings/${bindingId}`,
    method: 'delete'
  });
}

/**
 * 授权主体候选（用户+用户组，授权表单专用，权限归属 k8s.permission.list）
 */
export function fetchK8sSubjectOptions(clusterId: number) {
  return request<{
    users: { id: number; username: string; nickname: string; groupIds?: number[] }[];
    groups: { id: number; name: string; code: string }[];
  }>({
    url: `/k8s/clusters/${clusterId}/permission/subject-options`,
    method: 'get'
  });
}

/**
 * 集群内角色候选（授权表单专用，权限归属 k8s.permission.list）
 */
export function fetchK8sRoleOptions(clusterId: number, kind: 'ClusterRole' | 'Role') {
  return request<string[]>({
    url: `/k8s/clusters/${clusterId}/permission/role-options`,
    method: 'get',
    params: { kind }
  });
}
