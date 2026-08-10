/**
 * 统一的 API 错误码定义
 * 与后端 errors/errors.go 中的错误码保持一致
 */

export enum ErrorCode {
  // 通用错误 (50000-50099)
  INTERNAL_ERROR = 50000,
  BAD_REQUEST = 40002,
  UNAUTHORIZED = 40100,
  FORBIDDEN = 40300,
  NOT_FOUND = 40400,

  // 参数验证错误 (40000-40099)
  INVALID_PARAMS = 40000,
  MISSING_PARAM = 40001,

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

/**
 * 错误码到消息的映射
 */
export const ErrorMessages: Record<ErrorCode, string> = {
  // 通用错误
  [ErrorCode.INTERNAL_ERROR]: '服务器内部错误，请稍后重试',
  [ErrorCode.BAD_REQUEST]: '请求参数错误',
  [ErrorCode.UNAUTHORIZED]: '未授权访问',
  [ErrorCode.FORBIDDEN]: '权限不足',
  [ErrorCode.NOT_FOUND]: '资源不存在',

  // 参数验证错误
  [ErrorCode.INVALID_PARAMS]: '请求参数格式错误',
  [ErrorCode.MISSING_PARAM]: '缺少必填参数',

  // 用户相关错误
  [ErrorCode.USER_NOT_FOUND]: '用户不存在',
  [ErrorCode.INVALID_PASSWORD]: '密码错误',
  [ErrorCode.TOKEN_EXPIRED]: '登录已过期，请重新登录',
  [ErrorCode.TOKEN_INVALID]: '登录无效，请重新登录',

  // 服务器相关错误
  [ErrorCode.SERVER_NOT_FOUND]: '服务器不存在',
  [ErrorCode.DUPLICATE_HOSTNAME]: '主机名已存在，请使用其他主机名',
  [ErrorCode.DUPLICATE_IP]: 'IP地址已存在，请使用其他IP地址',
  [ErrorCode.INVALID_CREDENTIAL]: '无效的SSH凭证',

  // 菜单相关错误
  [ErrorCode.MENU_NOT_FOUND]: '菜单不存在',
  [ErrorCode.MENU_HAS_CHILDREN]: '该菜单下有子菜单，无法删除',
  [ErrorCode.MENU_IN_USE]: '该菜单正在使用中',

  // 角色相关错误
  [ErrorCode.ROLE_NOT_FOUND]: '角色不存在',
  [ErrorCode.ROLE_IN_USE]: '该角色正在使用中',
  [ErrorCode.ROLE_HAS_USERS]: '该角色下有用户，无法删除',

  // 凭证相关错误
  [ErrorCode.CREDENTIAL_NOT_FOUND]: '凭证不存在',
  [ErrorCode.CREDENTIAL_IN_USE]: '该凭证正在使用中',

  // 审计相关错误
  [ErrorCode.AUDIT_NOT_FOUND]: '审计记录不存在'
};

/**
 * 根据错误码获取错误消息
 */
export function getErrorMessage(code: ErrorCode): string {
  return ErrorMessages[code] || '未知错误';
}

/**
 * API 错误接口
 */
export interface ApiError {
  code: ErrorCode;
  message: string;
}

/**
 * API 响应接口
 */
export interface ApiResponse<T = any> {
  code: number;
  success: boolean;
  data?: T;
  message: string;
  errors?: FieldError[];
}

/**
 * 字段级错误
 */
export interface FieldError {
  field: string;
  message: string;
}

/**
 * 错误处理类
 */
export class ApiErrorHandler {
  /**
   * 处理 API 错误
   */
  static handle(error: ApiError): void {
    const message = getErrorMessage(error.code);
    console.error('API Error:', error.code, message);

    switch (error.code) {
      case ErrorCode.TOKEN_EXPIRED:
      case ErrorCode.TOKEN_INVALID:
        ElMessage.warning({
          message,
          duration: 1500,
          onClose: () => {
            // 跳转到登录页
            window.location.href = '/login';
          }
        });
        break;

      case ErrorCode.UNAUTHORIZED:
        ElMessage.error(message);
        break;

      case ErrorCode.FORBIDDEN:
        ElMessage.error('权限不足');
        break;

      default:
        ElMessage.error(message);
    }
  }

  /**
   * 处理字段级验证错误
   */
  static handleFieldErrors(errors: FieldError[], formRef?: any): void {
    if (!errors || errors.length === 0) return;

    // 如果有表单引用，设置字段错误
    if (formRef && formRef.validateFields) {
      errors.forEach(({ field, message }) => {
        formRef.validateField(field, message);
      });
    }

    // 显示错误消息
    ElMessage.error(`${errors[0].field}: ${errors[0].message}`);
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
