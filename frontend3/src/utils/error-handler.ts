/**
 * API 错误处理工具
 */

// 业务错误码定义
export enum ErrorCode {
  // 通用错误
  INTERNAL_ERROR = 50000,
  BAD_REQUEST = 40002,
  UNAUTHORIZED = 40100,
  FORBIDDEN = 40300,
  NOT_FOUND = 40400,

  // 参数验证错误
  INVALID_PARAMS = 40000,

  // 用户相关错误 (41000-41099)
  USER_NOT_FOUND = 41000,
  INVALID_PASSWORD = 41001,
  TOKEN_EXPIRED = 41002,
  TOKEN_INVALID = 41003,

  // 服务器相关错误 (42000-42099)
  SERVER_NOT_FOUND = 42000,
  DUPLICATE_HOSTNAME = 42001,
  DUPLICATE_IP = 42002,
  INVALID_CREDENTIAL = 42003,

  // 菜单相关错误 (43000-43099)
  MENU_NOT_FOUND = 43000,
  MENU_HAS_CHILDREN = 43001,
  MENU_IN_USE = 43002,

  // 角色相关错误 (44000-44099)
  ROLE_NOT_FOUND = 44000,
  ROLE_IN_USE = 44001,
  ROLE_HAS_USERS = 44002,

  // 凭证相关错误 (45000-45099)
  CREDENTIAL_NOT_FOUND = 45000,
  CREDENTIAL_IN_USE = 45001,

  // 审计相关错误 (46000-46099)
  AUDIT_NOT_FOUND = 46000
}

// API 错误接口
export interface ApiError {
  code: ErrorCode;
  message: string;
}

// API 响应接口
export interface ApiResponse<T = unknown> {
  code: number;
  success: boolean;
  data?: T;
  message: string;
  errors?: FieldError[];
}

// 字段级错误
export interface FieldError {
  field: string;
  message: string;
}

/**
 * 统一错误处理
 */
export class ErrorHandler {
  /**
   * 处理 API 错误
   */
  static handleApiError(error: ApiError): void {
    console.error('API Error:', error);

    switch (error.code) {
      case ErrorCode.DUPLICATE_HOSTNAME:
        ElMessage.error('主机名已存在，请使用其他主机名');
        break;
      case ErrorCode.DUPLICATE_IP:
        ElMessage.error('IP地址已存在，请使用其他IP地址');
        break;
      case ErrorCode.SERVER_NOT_FOUND:
        ElMessage.error('服务器不存在');
        break;
      case ErrorCode.INVALID_CREDENTIAL:
        ElMessage.error('无效的SSH凭证');
        break;
      case ErrorCode.USER_NOT_FOUND:
        ElMessage.error('用户不存在');
        break;
      case ErrorCode.INVALID_PASSWORD:
        ElMessage.error('密码错误');
        break;
      case ErrorCode.TOKEN_EXPIRED:
        ElMessage.warning('登录已过期，请重新登录');
        // 跳转到登录页
        setTimeout(() => {
          window.location.href = '/login';
        }, 1500);
        break;
      case ErrorCode.TOKEN_INVALID:
        ElMessage.warning('登录无效，请重新登录');
        setTimeout(() => {
          window.location.href = '/login';
        }, 1500);
        break;
      case ErrorCode.UNAUTHORIZED:
        ElMessage.error('未授权访问');
        break;
      case ErrorCode.FORBIDDEN:
        ElMessage.error('权限不足');
        break;
      case ErrorCode.MENU_HAS_CHILDREN:
        ElMessage.error('该菜单下有子菜单，无法删除');
        break;
      case ErrorCode.ROLE_IN_USE:
        ElMessage.error('该角色正在使用中，无法删除');
        break;
      case ErrorCode.CREDENTIAL_IN_USE:
        ElMessage.error('该凭证正在使用中，无法删除');
        break;
      case ErrorCode.INVALID_PARAMS:
        ElMessage.error(`请求参数错误：${error.message}`);
        break;
      case ErrorCode.INTERNAL_ERROR:
        ElMessage.error('服务器内部错误，请稍后重试');
        break;
      default:
        ElMessage.error(error.message || '未知错误');
    }
  }

  /**
   * 处理字段级验证错误
   */
  static handleFieldErrors(errors: FieldError[], formRef?: { validateField: (field: string, message: string) => void }): void {
    if (!errors || errors.length === 0) return;

    // 如果有表单引用，设置字段错误
    if (formRef && formRef.validateFields) {
      errors.forEach(({ field, message }) => {
        formRef.validateField(field, message);
      });
    }

    // 显示第一个错误消息
    const firstError = errors[0];
    ElMessage.error(`${firstError.field}: ${firstError.message}`);
  }

  /**
   * 从响应中提取错误信息
   */
  static extractError(response: ApiResponse): ApiError | null {
    if (!response.success) {
      return {
        code: response.code as ErrorCode,
        message: response.message
      };
    }
    return null;
  }
}

/**
 * 请求错误处理 Hook
 */
export function useErrorHandler() {
  const handleError = (error: { response?: { data?: unknown } }): void => {
    if (error.response) {
      const data = error.response.data as ApiResponse;
      const apiError = ErrorHandler.extractError(data);
      if (apiError) {
        ErrorHandler.handleApiError(apiError);
      } else {
        ElMessage.error(data.message || '请求失败');
      }
    } else if (error.message) {
      ElMessage.error(error.message);
    } else {
      ElMessage.error('未知错误');
    }
  };

  return { handleError };
}
