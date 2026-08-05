import { request } from '../request';

// ==================== 用户管理相关接口 ====================

/**
 * 获取所有角色
 */
export function fetchGetAllRoles() {
  return request<Api.SystemManage.RoleList>({
    url: '/system/roles',
    method: 'get'
  });
}

/**
 * 获取用户列表
 */
export function fetchGetUserList(params?: Api.SystemManage.UserSearchParams) {
  return request<Api.SystemManage.UserList>({
    url: '/system/users',
    method: 'get',
    params
  });
}

/**
 * 获取用户详情
 */
export function fetchGetUserById(id: number) {
  return request<Api.SystemManage.User>({
    url: `/system/users/${id}`,
    method: 'get'
  });
}

/**
 * 创建用户
 */
export function fetchCreateUser(data: Api.SystemManage.User) {
  return request({
    url: '/system/users',
    method: 'post',
    data
  });
}

/**
 * 更新用户
 */
export function fetchUpdateUser(id: number, data: Partial<Api.SystemManage.User>) {
  return request({
    url: `/system/users/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除用户
 */
export function fetchDeleteUser(id: number) {
  return request({
    url: `/system/users/${id}`,
    method: 'delete'
  });
}

/**
 * 重置用户密码
 */
export function fetchResetUserPassword(userId: number, newPassword: string) {
  return request({
    url: `/system/users/${userId}/password`,
    method: 'put',
    data: { password: newPassword }
  });
}

/**
 * 获取角色列表
 */
export function fetchGetRoleList(params?: Api.SystemManage.RoleSearchParams) {
  return request<Api.SystemManage.RoleList>({
    url: '/system/roles',
    method: 'get',
    params
  });
}

/**
 * 获取角色详情
 */
export function fetchGetRoleById(id: number) {
  return request<Api.SystemManage.Role>({
    url: `/system/roles/${id}`,
    method: 'get'
  });
}

/**
 * 创建角色
 */
export function fetchCreateRole(data: Api.SystemManage.Role) {
  return request({
    url: '/system/roles',
    method: 'post',
    data
  });
}

/**
 * 更新角色
 */
export function fetchUpdateRole(id: number, data: Partial<Api.SystemManage.Role>) {
  return request({
    url: `/system/roles/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除角色
 */
export function fetchDeleteRole(id: number) {
  return request({
    url: `/system/roles/${id}`,
    method: 'delete'
  });
}

/**
 * 获取菜单列表
 */
export function fetchGetMenuList(params?: { current?: number; size?: number }) {
  return request<Api.SystemManage.MenuList>({
    url: '/system/menus',
    method: 'get',
    params
  });
}

/**
 * 获取菜单树
 */
export function fetchGetMenuTree() {
  return request<Api.SystemManage.MenuTree[]>({
    url: '/system/menus/tree',
    method: 'get'
  });
}

/**
 * 创建菜单
 */
export function fetchCreateMenu(data: Api.SystemManage.Menu) {
  return request({
    url: '/system/menus',
    method: 'post',
    data
  });
}

/**
 * 更新菜单
 */
export function fetchUpdateMenu(id: number, data: Partial<Api.SystemManage.Menu>) {
  return request({
    url: `/system/menus/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除菜单
 */
export function fetchDeleteMenu(id: number) {
  return request({
    url: `/system/menus/${id}`,
    method: 'delete'
  });
}

/**
 * 获取属性定义列表
 */
export function fetchGetAttributes(params?: { category?: string }) {
  return request<System.AttributeDefinition[]>({
    url: '/system/attributes',
    method: 'get',
    params
  });
}

/**
 * 获取属性定义详情
 */
export function fetchGetAttributeById(id: number) {
  return request<System.AttributeDefinition>({
    url: `/system/attributes/${id}`,
    method: 'get'
  });
}

/**
 * 创建属性定义
 */
export function fetchCreateAttribute(data: System.AttributeDefinitionForm) {
  return request({
    url: '/system/attributes',
    method: 'post',
    data
  });
}

/**
 * 更新属性定义
 */
export function fetchUpdateAttribute(id: number, data: Partial<System.AttributeDefinitionForm>) {
  return request({
    url: `/system/attributes/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除属性定义
 */
export function fetchDeleteAttribute(id: number) {
  return request({
    url: `/system/attributes/${id}`,
    method: 'delete'
  });
}

/**
 * 验证主机属性值
 */
export function fetchValidateServerAttribute(data: { attributeId: number; value: string }) {
  return request({
    url: '/system/attributes/validate',
    method: 'post',
    data
  });
}

/**
 * 获取权限列表
 */
export function fetchGetPermissionList(params?: Api.SystemManage.PermissionSearchParams) {
  return request<{
    records: Api.SystemManage.Permission[];
    total: number;
    current: number;
    size: number;
  }>({
    url: '/system/permissions',
    method: 'get',
    params
  });
}

/**
 * 获取权限详情
 */
export function fetchGetPermissionById(id: number) {
  return request<Api.SystemManage.Permission>({
    url: `/system/permissions/${id}`,
    method: 'get'
  });
}

/**
 * 新增权限
 */
export function fetchAddPermission(data: Api.SystemManage.Permission) {
  return request({
    url: '/system/permissions',
    method: 'post',
    data
  });
}

/**
 * 更新权限
 */
export function fetchUpdatePermission(id: number, data: Partial<Api.SystemManage.Permission>) {
  return request({
    url: `/system/permissions/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除权限
 */
export function fetchDeletePermission(id: number) {
  return request({
    url: `/system/permissions/${id}`,
    method: 'delete'
  });
}

/**
 * 获取角色权限列表
 */
export function fetchGetRolePermissions(roleId: number) {
  return request<number[]>({
    url: `/system/roles/${roleId}/permissions`,
    method: 'get'
  });
}

/**
 * 为角色分配权限
 */
export function fetchAssignRolePermissions(roleId: number, data: { permissionIds: number[] }) {
  return request({
    url: `/system/roles/${roleId}/permissions`,
    method: 'post',
    data
  });
}

