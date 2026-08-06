import { useAuthStore } from '@/store/modules/auth'

/**
 * 权限工具函数
 *
 * 提供简单的权限检查函数，类似于 sxdevops 项目的实现
 * 可以在任何地方使用，不局限于 Vue 组件
 */

/**
 * 检查单个权限
 * @param permission 权限编码
 * @returns 是否有权限
 */
export function can(permission: string): boolean {
  return useAuthStore().hasPermission(permission)
}

/**
 * 检查是否拥有任意一个权限
 * @param permissions 权限编码数组
 * @returns 是否有任意一个权限
 */
export function canAny(permissions: string[]): boolean {
  return useAuthStore().hasAnyPermission(permissions)
}

/**
 * 检查是否拥有所有权限
 * @param permissions 权限编码数组
 * @returns 是否拥有所有权限
 */
export function canAll(permissions: string[]): boolean {
  return useAuthStore().hasAllPermissions(permissions)
}

/**
 * 权限编码常量
 * 方便在代码中使用，避免硬编码权限字符串
 */
export const PERMISSIONS = {
  // 用户管理
  USER_LIST: 'system.user.list',
  USER_CREATE: 'system.user.create',
  USER_UPDATE: 'system.user.update',
  USER_DELETE: 'system.user.delete',
  USER_RESET_PASSWORD: 'system.user.reset_password',
  USER_UNLOCK: 'system.user.unlock',
  USER_ASSIGN_ROLE: 'system.user.assign_role',
  USER_EXPORT: 'system.user.export',
  USER_IMPORT: 'system.user.import',

  // 角色管理
  ROLE_LIST: 'system.role.list',
  ROLE_CREATE: 'system.role.create',
  ROLE_UPDATE: 'system.role.update',
  ROLE_DELETE: 'system.role.delete',
  ROLE_ASSIGN_PERMISSION: 'system.role.assign_permission',
  ROLE_VIEW_USERS: 'system.role.view_users',

  // 权限管理
  PERMISSION_LIST: 'system.permission.list',
  PERMISSION_CREATE: 'system.permission.create',
  PERMISSION_UPDATE: 'system.permission.update',
  PERMISSION_DELETE: 'system.permission.delete',
  PERMISSION_EXPORT: 'system.permission.export',
  PERMISSION_IMPORT: 'system.permission.import',

  // 菜单管理
  MENU_LIST: 'system.menu.list',
  MENU_CREATE: 'system.menu.create',
  MENU_UPDATE: 'system.menu.update',
  MENU_DELETE: 'system.menu.delete',
  MENU_SORT: 'system.menu.sort',

  // 登录日志
  LOGIN_LOG_VIEW: 'audit.login_log.view',
  LOGIN_LOG_EXPORT: 'audit.login_log.export',

  // 操作日志
  OPERATION_LOG_VIEW: 'audit.operation_log.view',
  OPERATION_LOG_EXPORT: 'audit.operation_log.export',

  // 系统设置
  SETTINGS_VIEW: 'system.settings.view',
  SETTINGS_UPDATE: 'system.settings.update',
  SETTINGS_RESTART: 'system.settings.restart'
}

/**
 * 权限分组常量
 * 用于权限选择器等场景
 */
export const PERMISSION_GROUPS = {
  USER_MANAGEMENT: {
    name: '用户管理',
    permissions: [
      PERMISSIONS.USER_LIST,
      PERMISSIONS.USER_CREATE,
      PERMISSIONS.USER_UPDATE,
      PERMISSIONS.USER_DELETE,
      PERMISSIONS.USER_RESET_PASSWORD,
      PERMISSIONS.USER_UNLOCK,
      PERMISSIONS.USER_ASSIGN_ROLE,
      PERMISSIONS.USER_EXPORT,
      PERMISSIONS.USER_IMPORT
    ]
  },
  ROLE_MANAGEMENT: {
    name: '角色管理',
    permissions: [
      PERMISSIONS.ROLE_LIST,
      PERMISSIONS.ROLE_CREATE,
      PERMISSIONS.ROLE_UPDATE,
      PERMISSIONS.ROLE_DELETE,
      PERMISSIONS.ROLE_ASSIGN_PERMISSION,
      PERMISSIONS.ROLE_VIEW_USERS
    ]
  },
  PERMISSION_MANAGEMENT: {
    name: '权限管理',
    permissions: [
      PERMISSIONS.PERMISSION_LIST,
      PERMISSIONS.PERMISSION_CREATE,
      PERMISSIONS.PERMISSION_UPDATE,
      PERMISSIONS.PERMISSION_DELETE,
      PERMISSIONS.PERMISSION_EXPORT,
      PERMISSIONS.PERMISSION_IMPORT
    ]
  }
}