import { DirectiveBinding } from 'vue'
import { useAuthStore } from '@/store/modules/auth'

/**
 * 权限指令 v-permission
 *
 * 使用示例：
 * <button v-permission="'system.user.create'">新增用户</button>
 * <button v-permission="['system.user.update', 'system.user.delete']">操作</button>
 */
export default {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    const authStore = useAuthStore()
    let hasAuth = false
    const { value } = binding

    // 单个权限字符串
    if (typeof value === 'string') {
      hasAuth = authStore.hasPermission(value)
    }
    // 多个权限数组（满足任意一个即可）
    else if (Array.isArray(value)) {
      hasAuth = authStore.hasAnyPermission(value)
    }
    // 对象配置
    else if (typeof value === 'object' && value !== null) {
      const { code, codes, mode = 'any' } = value

      if (code) {
        hasAuth = authStore.hasPermission(code)
      } else if (codes && Array.isArray(codes)) {
        hasAuth = mode === 'all'
          ? authStore.hasAllPermissions(codes)
          : authStore.hasAnyPermission(codes)
      }
    }

    // 无权限时移除元素
    if (!hasAuth) {
      // 方式1：完全移除元素（推荐）
      el.parentNode?.removeChild(el)
    }
  },

  updated(el: HTMLElement, binding: DirectiveBinding) {
    const authStore = useAuthStore()
    let hasAuth = false
    const { value } = binding

    if (typeof value === 'string') {
      hasAuth = authStore.hasPermission(value)
    } else if (Array.isArray(value)) {
      hasAuth = authStore.hasAnyPermission(value)
    } else if (typeof value === 'object' && value !== null) {
      const { code, codes, mode = 'any' } = value

      if (code) {
        hasAuth = authStore.hasPermission(code)
      } else if (codes && Array.isArray(codes)) {
        hasAuth = mode === 'all'
          ? authStore.hasAllPermissions(codes)
          : authStore.hasAnyPermission(codes)
      }
    }

    if (!hasAuth) {
      el.style.display = 'none'
    } else {
      el.style.display = ''
    }
  }
}