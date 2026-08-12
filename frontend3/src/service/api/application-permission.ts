import { request } from '../request';

// ========== 应用权限管理相关 API ==========

/** 获取应用列表 */
export function fetchApplications(params: Api.ApplicationPermission.ApplicationSearchParams) {
  return request<Api.Common.PaginatingQueryRecord<Api.ApplicationPermission.Application>>({
    url: '/system/applications',
    method: 'get',
    params
  });
}

/** 获取应用选项列表（不分页，用于选择器） */
export function fetchApplicationOptions() {
  return request<{ id: number; name: string; code: string; type: string }[]>({
    url: '/system/applications/options',
    method: 'get'
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
  return request<Api.ApplicationPermission.GroupBinding[]>({
    method: 'get'
  });
}

/** 创建用户组权限绑定 */
export function createGroupBinding(data: { groupId: number; appId: number; applicationRoleId: number }) {
  return request<Api.ApplicationPermission.GroupBindingResponse>({
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
  return request<Api.ApplicationPermission.AuthUserGroup[]>({
    method: 'get'
  });
}

/** 为用户分配用户组成员 */
/** 分配用户到用户组（会触发外部系统授权） */
export function assignUserToGroup(data: { userId: number; groupId: number }) {
  return request<Api.ApplicationPermission.AssignmentResponse>({
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
export function fetchOperationLogs(appId: number, params: { page: number; pageSize: number }) {
  return request<Api.ApplicationPermission.OperationLogList>({
    url: `/system/applications/${appId}/operation-logs`,
    method: 'get',
    params
  });
}

// ========== 授权中心用户管理 API ==========

/** 获取授权中心用户列表 */
export function fetchAuthUsers(params: {
  page: number;
  pageSize: number;
  username?: string;
  nickname?: string;
  email?: string;
  phone?: string;
}) {
  return request<Api.Common.PaginatingQueryRecord<Api.ApplicationPermission.AuthUser>>({
    url: '/system/auth-users/list',
    method: 'get',
    params
  });
}

/** 创建授权中心用户返回的初始密码 */
export type CreatedAuthUser = {
  username: string;
  password: string;
};

/** 创建授权中心用户 */
export function createAuthUser(data: Partial<Api.ApplicationPermission.AuthUser>) {
  return request<CreatedAuthUser>({
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
export function fetchAuthGroups(params: { page: number; pageSize: number; name?: string }) {
  return request<Api.Common.PaginatingQueryRecord<Api.ApplicationPermission.AuthGroup>>({
    url: '/system/auth-groups/list',
    method: 'get',
    params
  });
}

/** 获取用户组选项列表（不分页，用于选择器） */
export function fetchAuthGroupOptions() {
  return request<{ id: number; name: string; code: string }[]>({
    url: '/system/auth-groups/options',
    method: 'get'
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
    configTemplate: Record<string, unknown>;
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

// ========== 用户身份映射管理 API ==========

/** 获取用户身份映射列表 */
export function fetchUserIdentityMappings(params: {
  page: number;
  pageSize: number;
  username?: string;
  appId?: number | null;
  status?: string;
}) {
  return request<Api.ApplicationPermission.IdentityMappingList>({
    url: '/system/user-identity-mappings',
    method: 'get',
    params
  });
}

/** 删除用户身份映射 */
export function deleteUserIdentityMapping(id: number) {
  return request<boolean>({
    url: `/system/user-identity-mappings/${id}`,
    method: 'delete'
  });
}

// ========== 用户有效权限查询 API ==========

/** 获取用户有效权限列表 */
export function fetchUserEffectivePermissions(params: {
  page: number;
  pageSize: number;
  username?: string;
  appId?: number | null;
}) {
  return request<Api.ApplicationPermission.UserPermissionList>({
    url: '/system/user-permissions',
    method: 'get',
    params
  });
}

/** 获取用户有效权限矩阵视图 */
export function fetchUserEffectivePermissionsMatrix(appId: number) {
  return request<Api.ApplicationPermission.UserEffectivePermission[]>({
    url: '/system/user-permissions/matrix',
    method: 'get',
    params: { appId }
  });
}

// ========== 权限执行记录 API ==========

/** 获取权限绑定执行记录 */
export function fetchGroupBindingExecutions(bindingId: number) {
  return request<Api.ApplicationPermission.GroupBindingExecution[]>({
    url: `/system/group-bindings/${bindingId}/executions`,
    method: 'get'
  });
}
