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
 * 获取角色选项列表（不分页，用于选择器）
 */
export function fetchRoleOptions() {
  return request<{ id: number; name: string; code: string }[]>({
    url: '/system/roles/options',
    method: 'get'
  });
}

/**
 * 按角色集合查询可见一级菜单（编辑用户时家目录候选）
 */
export function fetchGetRoleMenuPaths(roleIds: number[]) {
  return request<{ id: number; name: string; path: string }[]>({
    url: '/system/roles/menu-paths',
    method: 'get',
    params: { roleIds: roleIds.join(',') }
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
 * 获取用户选项列表（不分页，用于选择器）
 */
export function fetchUserOptions() {
  return request<{ id: number; username: string; nickname: string }[]>({
    url: '/system/users/options',
    method: 'get'
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
export function fetchGetMenuList(params?: { page?: number; pageSize?: number }) {
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
    url: '/cmdb/attributes',
    method: 'get',
    params
  });
}

/**
 * 获取属性定义详情
 */
export function fetchGetAttributeById(id: number) {
  return request<System.AttributeDefinition>({
    url: `/cmdb/attributes/${id}`,
    method: 'get'
  });
}

/**
 * 创建属性定义
 */
export function fetchCreateAttribute(data: System.AttributeDefinitionForm) {
  return request({
    url: '/cmdb/attributes',
    method: 'post',
    data
  });
}

/**
 * 更新属性定义
 */
export function fetchUpdateAttribute(id: number, data: Partial<System.AttributeDefinitionForm>) {
  return request({
    url: `/cmdb/attributes/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除属性定义
 */
export function fetchDeleteAttribute(id: number) {
  return request({
    url: `/cmdb/attributes/${id}`,
    method: 'delete'
  });
}

/**
 * 验证主机属性值
 */
export function fetchValidateServerAttribute(data: { attributeId: number; value: string }) {
  return request({
    url: '/cmdb/attributes/validate',
    method: 'post',
    data
  });
}

/**
 * 获取权限列表
 */
export function fetchGetPermissionList(params?: Api.SystemManage.PermissionSearchParams) {
  return request<Api.Common.PaginatingQueryRecord<Api.SystemManage.Permission>>({
    url: '/system/permissions',
    method: 'get',
    params
  });
}

/**
 * 获取权限选项列表（不分页，用于选择器/权限树）
 */
export function fetchPermissionOptions() {
  return request<{ id: number; name: string; code: string; module: string; resource: string; action: string; parentId: number; level: number; sortOrder: number }[]>({
    url: '/system/permissions/options',
    method: 'get'
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
export function fetchAddPermission(data: Partial<Api.SystemManage.Permission>) {
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
 * 获取权限路由映射列表（可按权限码过滤）
 */
export function fetchGetPermissionRoutes(params?: { permissionCode?: string }) {
  return request<Api.SystemManage.PermissionRoute[]>({
    url: '/system/permissions/routes',
    method: 'get',
    params
  });
}

/**
 * 新增权限路由映射（同一端点只能归属一个权限码，即时生效）
 */
export function fetchCreatePermissionRoute(data: Pick<Api.SystemManage.PermissionRoute, 'permissionCode' | 'method' | 'path'>) {
  return request({
    url: '/system/permissions/routes',
    method: 'post',
    data
  });
}

/**
 * 修改权限路由映射归属的权限码（method/path 不可改）
 */
export function fetchUpdatePermissionRoute(id: number, data: { permissionCode: string }) {
  return request({
    url: `/system/permissions/routes/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除权限路由映射（删除后对应端点 fail-closed 拒绝访问）
 */
export function fetchDeletePermissionRoute(id: number) {
  return request({
    url: `/system/permissions/routes/${id}`,
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

// ==================== 用户组管理相关接口 ====================

/**
 * 获取用户组列表（分页，含成员数）
 */
export function fetchGetUserGroupList(params?: Api.SystemManage.UserGroupSearchParams) {
  return request<Api.SystemManage.UserGroupList>({
    url: '/system/user-groups',
    method: 'get',
    params
  });
}

/**
 * 获取用户组选项（不分页，用于选择器）
 */
export function fetchUserGroupOptions() {
  return request<Api.SystemManage.UserGroupOption[]>({
    url: '/system/user-groups/options',
    method: 'get'
  });
}

/**
 * 创建用户组
 */
export function fetchCreateUserGroup(data: { code: string; name: string; description?: string }) {
  return request({
    url: '/system/user-groups',
    method: 'post',
    data
  });
}

/**
 * 更新用户组（code 不可改；name 必传）
 */
export function fetchUpdateUserGroup(
  id: number,
  data: { name: string; description?: string; status?: number }
) {
  return request({
    url: `/system/user-groups/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除用户组（级联删除组成员与集群组绑定）
 */
export function fetchDeleteUserGroup(id: number) {
  return request({
    url: `/system/user-groups/${id}`,
    method: 'delete'
  });
}

/**
 * 获取组成员列表
 */
export function fetchGetGroupMembers(id: number) {
  return request<Api.SystemManage.GroupMember[]>({
    url: `/system/user-groups/${id}/members`,
    method: 'get'
  });
}

/**
 * 批量添加组成员
 */
export function fetchAddGroupMembers(id: number, userIds: number[]) {
  return request({
    url: `/system/user-groups/${id}/members`,
    method: 'post',
    data: { userIds }
  });
}

/**
 * 移除组成员
 */
export function fetchRemoveGroupMember(id: number, userId: number) {
  return request({
    url: `/system/user-groups/${id}/members/${userId}`,
    method: 'delete'
  });
}
