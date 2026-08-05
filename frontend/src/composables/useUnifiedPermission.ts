/**
 * 统一权限检查和提醒系统
 * 前端不检查权限，直接执行操作，显示后端返回的错误信息
 */

import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import { useAuthStore } from '@/store/modules/auth'

/**
 * 权限提醒配置
 */
export interface PermissionAlertConfig {
  title?: string
  message?: string
  showContact?: boolean
  contactInfo?: string
  type?: 'warning' | 'error' | 'info'
  duration?: number
  useNotification?: boolean
}

/**
 * 统一权限Hook
 */
export const useUnifiedPermission = () => {
  const authStore = useAuthStore()

  /**
   * 执行操作（不做权限检查，捕获并显示后端错误）
   * @param permission 权限编码（仅用于日志，不用于检查）
   * @param action 要执行的操作
   * @param config 提醒配置（可选）
   */
  const executeWithPermission = async (
    permission: string | string[],
    action: () => void | Promise<void>,
    config: PermissionAlertConfig = {}
  ): Promise<boolean> => {
    try {
      await action()
      return true
    } catch (error: any) {
      // 捕获权限错误（403）
      const statusCode = error?.response?.status || error?.statusCode || error?.code

      if (statusCode === 403) {
        // 显示后端返回的错误信息
        const backendMessage = error?.response?.data?.message
          || error?.response?.data?.error
          || error?.message
          || '权限不足'

        ElMessage.error(backendMessage)
        return false
      }

      // 其他错误也显示
      const errorMessage = error?.response?.data?.message
        || error?.message
        || '操作执行失败，请重试'

      ElMessage.error(errorMessage)
      return false
    }
  }

  /**
   * 显示权限提醒
   */
  const showPermissionAlert = (
    permission: string | string[],
    config: PermissionAlertConfig = {}
  ): boolean => {
    const {
      title = '权限不足',
      message,
      showContact = true,
      contactInfo = 'admin@company.com',
      type = 'warning',
      useNotification = false
    } = config

    const permissionText = Array.isArray(permission)
      ? `以下权限之一：${permission.join('、')}`
      : `【${authStore.getPermissionName(permission)}】权限 (${permission})`

    const defaultMessage = `您需要${permissionText}才能执行此操作`
    const fullMessage = message || defaultMessage

    const contactText = showContact
      ? `\n\n如需使用此功能，请联系管理员申请权限\n📧 联系方式：${contactInfo}`
      : ''

    if (useNotification) {
      ElNotification({
        title,
        message: fullMessage + contactText,
        type,
        duration: 5000,
        showClose: true
      })
    } else {
      ElMessageBox.alert(
        fullMessage + contactText,
        title,
        {
          type,
          confirmButtonText: '我知道了',
          dangerouslyUseHTMLString: false
        }
      )
    }

    return false
  }

  /**
   * 批量权限检查（用于UI显示控制）
   */
  const checkBatchPermissions = (permissions: string[]): Record<string, boolean> => {
    const result: Record<string, boolean> = {}
    permissions.forEach(permission => {
      result[permission] = authStore.hasPermission(permission)
    })
    return result
  }

  /**
   * 显示后端错误信息
   */
  const showBackendError = (error: any) => {
    const backendMessage = error?.response?.data?.message
      || error?.response?.data?.error
      || error?.message
      || '操作失败'

    ElMessage.error(backendMessage)
  }

  return {
    executeWithPermission,
    showPermissionAlert,
    checkBatchPermissions,
    showBackendError
  }
}

export default useUnifiedPermission
