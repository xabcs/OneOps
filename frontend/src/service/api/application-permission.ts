import { request } from '../request';

// ========== 应用权限管理相关 API ==========

/** 获取应用列表 */
export function fetchApplications(params: Api.ApplicationPermission.ApplicationSearchParams) {
  return request<Api.ApplicationPermission.ApplicationList>({
    url: '/system/applications',
    method: 'get',
    params
  });
}

/** 创建应用 */
export function createApplication(data: Partial<Api.ApplicationPermission.Application>) {
  return request<boolean>({
    url: '/system/applications',
    method: 'post',
    data
  });
}

/** 更新应用 */
export function updateApplication(id: number, data: Partial<Api.ApplicationPermission.Application>) {
  return request<boolean>({
    url: `/system/applications/${id}`,
    method: 'put',
    data
  });
}

/** 删除应用 */
export function deleteApplication(id: number) {
  return request<boolean>({
    url: `/system/applications/${id}`,
    method: 'delete'
  });
}

/** 同步应用角色 */
export function syncApplicationRoles(id: number) {
  return request<boolean>({
    url: `/system/applications/${id}/sync-roles`,
    method: 'post'
  });
}

/** 获取应用角色列表 */
export function fetchApplicationRoles(id: number) {
  return request<Api.ApplicationPermission.ApplicationRole[]>({
    url: `/system/applications/${id}/roles`,
    method: 'get'
  });
}

/** 同步应用用户 */
export function syncApplicationUsers(id: number) {
  return request<boolean>({
    url: `/system/applications/${id}/sync-users`,
    method: 'post'
  });
}

/** 获取应用用户列表 */
export function fetchApplicationUsers(id: number) {
  return request<Api.ApplicationPermission.ApplicationUser[]>({
    url: `/system/applications/${id}/users`,
    method: 'get'
  });
}

/** 获取用户组权限绑定列表 */
export function fetchGroupBindings(groupId: number) {
  return request<any[]>({
    url: `/system/groups/${groupId}/bindings`,
    method: 'get'
  });
}

/** 创建用户组权限绑定 */
export function createGroupBinding(data: { groupId: number; appId: number; applicationRoleId: number }) {
  return request<boolean>({
    url: `/system/groups/${data.groupId}/bindings`,
    method: 'post',
    data
  });
}

/** 删除用户组权限绑定 */
export function deleteGroupBinding(id: number) {
  return request<boolean>({
    url: `/system/groups/bindings/${id}`,
    method: 'delete'
  });
}

/** 获取用户所属用户组列表 */
export function fetchUserGroups(userId: number) {
  return request<any[]>({
    url: `/system/users/${userId}/groups`,
    method: 'get'
  });
}

/** 为用户分配用户组成员 */
export function assignUserToGroup(data: { userId: number; groupId: number }) {
  return request<boolean>({
    url: '/system/users/assign-group',
    method: 'post',
    data
  });
}

/** 删除用户组成员 */
export function deleteUserGroup(userId: number, groupId: number) {
  return request<boolean>({
    url: `/system/users/${userId}/groups/${groupId}`,
    method: 'delete'
  });
}

/** 获取操作日志 */
export function fetchOperationLogs(appId: number, params: { current: number; size: number }) {
  return request<Api.ApplicationPermission.OperationLogList>({
    url: `/system/applications/${appId}/operation-logs`,
    method: 'get',
    params
  });
}

// ========== 授权中心用户管理 API ==========

/** 获取授权中心用户列表 */
export function fetchAuthUsers(params: { current: number; size: number; username?: string }) {
  return request<Api.ApplicationPermission.ApplicationList>({
    url: '/system/auth-users/list',
    method: 'get',
    params
  });
}

/** 创建授权中心用户 */
export function createAuthUser(data: Partial<Api.ApplicationPermission.AuthUser>) {
  return request<boolean>({
    url: '/system/auth-users',
    method: 'post',
    data
  });
}

/** 更新授权中心用户 */
export function updateAuthUser(id: number, data: Partial<Api.ApplicationPermission.AuthUser>) {
  return request<boolean>({
    url: `/system/auth-users/${id}`,
    method: 'put',
    data
  });
}

/** 删除授权中心用户 */
export function deleteAuthUser(id: number) {
  return request<boolean>({
    url: `/system/auth-users/${id}`,
    method: 'delete'
  });
}

/** 获取授权中心用户密码 */
export function getAuthUserPassword(id: number) {
  return request<{ username: string; password: string }>({
    url: `/system/auth-users/${id}/password`,
    method: 'get'
  });
}

/** 获取所有授权中心用户 */
export function fetchAllAuthUsers() {
  return request<Api.ApplicationPermission.AuthUser[]>({
    url: '/system/auth-users',
    method: 'get'
  });
}

// ========== 授权中心用户组管理 API ==========

/** 获取授权中心用户组列表 */
export function fetchAuthGroups(params: { current: number; size: number; name?: string }) {
  return request<Api.ApplicationPermission.ApplicationList>({
    url: '/system/auth-groups/list',
    method: 'get',
    params
  });
}

/** 创建授权中心用户组 */
export function createAuthGroup(data: Partial<Api.ApplicationPermission.AuthGroup>) {
  return request<boolean>({
    url: '/system/auth-groups',
    method: 'post',
    data
  });
}

/** 更新授权中心用户组 */
export function updateAuthGroup(id: number, data: Partial<Api.ApplicationPermission.AuthGroup>) {
  return request<boolean>({
    url: `/system/auth-groups/${id}`,
    method: 'put',
    data
  });
}

/** 删除授权中心用户组 */
export function deleteAuthGroup(id: number) {
  return request<boolean>({
    url: `/system/auth-groups/${id}`,
    method: 'delete'
  });
}

/** 获取所有授权中心用户组 */
export function fetchAllAuthGroups() {
  return request<Api.ApplicationPermission.AuthGroup[]>({
    url: '/system/auth-groups',
    method: 'get'
  });
}

// ========== 用户和用户组列表 API (兼容旧接口，调用授权中心接口) ==========

/** 获取所有用户列表 (调用授权中心用户) */
export function fetchAllUsers() {
  return fetchAllAuthUsers();
}

/** 获取所有用户组列表 (调用授权中心用户组) */
export function fetchAllRoles() {
  return fetchAllAuthGroups();
}

// ========== 应用类型配置 API ==========

/** 获取支持的应用类型列表 */
export function fetchSupportedAppTypes() {
  return request<Array<{ type: string; displayName: string }>>({
    url: '/system/applications/types',
    method: 'get'
  });
}

/** 获取应用类型的配置模板 */
export function fetchAppTypeConfigTemplate(appType: string) {
  return request<{
    appType: string;
    displayName: string;
    configTemplate: Record<string, any>;
  }>({
    url: `/system/applications/types/${appType}/config`,
    method: 'get'
  });
}

/** 同步应用用户组 */
export function syncApplicationGroups(id: number) {
  return request<boolean>({
    url: `/system/applications/${id}/sync-groups`,
    method: 'post'
  });
}

/** 获取应用用户组列表 */
export function fetchApplicationGroups(id: number) {
  return request<Api.ApplicationPermission.ApplicationGroup[]>({
    url: `/system/applications/${id}/groups`,
    method: 'get'
  });
}

/** 同步应用授权规则 */
export function syncApplicationAuthorizationRules(id: number) {
  return request<boolean>({
    url: `/system/applications/${id}/sync-rules`,
    method: 'post'
  });
}

/** 获取应用授权规则列表 */
export function fetchApplicationAuthorizationRules(id: number) {
  return request<Api.ApplicationPermission.AuthorizationRule[]>({
    url: `/system/applications/${id}/rules`,
    method: 'get'
  });
}
