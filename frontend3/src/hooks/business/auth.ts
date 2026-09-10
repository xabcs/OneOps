import { ElMessage } from 'element-plus';
import { useAuthStore } from '@/store/modules/auth';

export function useAuth() {
  const authStore = useAuthStore();

  function hasAuth(codes: string | string[]) {
    if (!authStore.isLogin) {
      return false;
    }

    if (typeof codes === 'string') {
      return authStore.userInfo.buttons.includes(codes);
    }

    return codes.some(code => authStore.userInfo.buttons.includes(code));
  }

  return {
    hasAuth
  };
}

/**
 * 权限工具函数
 *
 * 提供简单的权限检查函数
 * 可以在任何地方使用，不局限于 Vue 组件
 */

/**
 * 检查单个权限
 * @param permission 权限编码
 * @returns 是否有权限
 */
export function can(permission: string): boolean {
  return useAuthStore().hasPermission(permission);
}

/**
 * 检查是否拥有任意一个权限
 * @param permissions 权限编码数组
 * @returns 是否有任意一个权限
 */
export function canAny(permissions: string[]): boolean {
  return useAuthStore().hasAnyPermission(permissions);
}

/**
 * 检查是否拥有所有权限
 * @param permissions 权限编码数组
 * @returns 是否拥有所有权限
 */
export function canAll(permissions: string[]): boolean {
  return useAuthStore().hasAllPermissions(permissions);
}

/**
 * 执行操作（前端不做权限检查，捕获并显示后端返回的错误信息）
 *
 * 该函数不进行本地权限校验，直接执行操作；
 * 当后端返回 403 或其他错误时，统一提取并展示后端错误信息。
 *
 * @param permission 权限编码（仅用于语义标注，不参与检查）
 * @param action 要执行的操作
 * @returns 操作是否成功
 */
export async function executeWithPermission(
  permission: string | string[],
  action: () => void | Promise<void>
): Promise<boolean> {
  try {
    await action();
    return true;
  } catch (error: unknown) {
    const err = error as Record<string, unknown>;
    const statusCode = (err?.response as Record<string, unknown>)?.status || err?.statusCode || err?.code;

    if (statusCode === 403) {
      const backendMessage =
        error?.response?.data?.message || error?.response?.data?.error || error?.message || '权限不足';

      ElMessage.error(backendMessage);
      return false;
    }

    const errorMessage = error?.response?.data?.message || error?.message || '操作执行失败，请重试';

    ElMessage.error(errorMessage);
    return false;
  }
}
