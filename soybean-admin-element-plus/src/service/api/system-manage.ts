import { request } from '../request';

/** ================= Role ================= */

/** get role list */
export function fetchGetRoleList(params?: Api.SystemManage.RoleSearchParams) {
  return request<Api.SystemManage.RoleList>({
    url: '/system/roles',
    method: 'get',
    params
  });
}

/** create role */
export function fetchCreateRole(data: Omit<Api.SystemManage.Role, 'id' | 'createdAt' | 'updatedAt'>) {
  return request({
    url: '/system/roles',
    method: 'post',
    data
  });
}

/** update role */
export function fetchUpdateRole(id: number, data: Partial<Api.SystemManage.Role>) {
  return request({
    url: `/system/roles/${id}`,
    method: 'put',
    data
  });
}

/** delete role */
export function fetchDeleteRole(id: number) {
  return request({
    url: `/system/roles/${id}`,
    method: 'delete'
  });
}

/**
 * get all roles
 *
 * these roles are all enabled
 */
export function fetchGetAllRoles() {
  return request<Api.SystemManage.AllRole[]>({
    url: '/system/roles',
    method: 'get'
  });
}

/** ================= User ================= */

/** get user list */
export function fetchGetUserList(params?: Api.SystemManage.UserSearchParams) {
  return request<Api.SystemManage.UserList>({
    url: '/system/users',
    method: 'get',
    params
  });
}

/** create user */
export function fetchCreateUser(data: Omit<Api.SystemManage.User, 'id' | 'createdAt' | 'updatedAt'>) {
  return request({
    url: '/system/users',
    method: 'post',
    data
  });
}

/** update user */
export function fetchUpdateUser(id: number, data: Partial<Api.SystemManage.User>) {
  return request({
    url: `/system/users/${id}`,
    method: 'put',
    data
  });
}

/** delete user */
export function fetchDeleteUser(id: number) {
  return request({
    url: `/system/users/${id}`,
    method: 'delete'
  });
}

/** reset user password */
export function fetchResetUserPassword(id: number, password: string) {
  return request({
    url: `/system/users/${id}`,
    method: 'put',
    data: { password }
  });
}

/** ================= Menu ================= */

/** get menu list */
export function fetchGetMenuList() {
  return request<Api.SystemManage.MenuList>({
    url: '/system/menus',
    method: 'get'
  });
}

/** get menu tree */
export function fetchGetMenuTree() {
  return request<Api.SystemManage.MenuTree[]>({
    url: '/system/menus',
    method: 'get'
  });
}

/** create menu */
export function fetchCreateMenu(data: Omit<Api.SystemManage.Menu, 'id' | 'createdAt' | 'updatedAt' | 'children'>) {
  return request({
    url: '/system/menus',
    method: 'post',
    data
  });
}

/** update menu */
export function fetchUpdateMenu(id: number, data: Partial<Api.SystemManage.Menu>) {
  return request({
    url: `/system/menus/${id}`,
    method: 'put',
    data
  });
}

/** delete menu */
export function fetchDeleteMenu(id: number) {
  return request({
    url: `/system/menus/${id}`,
    method: 'delete'
  });
}
