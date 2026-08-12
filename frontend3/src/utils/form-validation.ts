/**
 * 表单验证工具
 * 用于在提交前验证表单数据
 */

import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { ValidationRules, validateObject } from './validation';
import type { ValidationResult, ValidationRule } from './validation';

/**
 * 服务器表单验证规则
 */
export const ServerFormRules = {
  hostname: {
    required: true,
    ...ValidationRules.hostname,
    message: '请输入主机名'
  },
  ip: {
    required: true,
    ...ValidationRules.ip,
    message: '请输入有效的IP地址'
  },
  innerIp: {
    ...ValidationRules.ip,
    message: '请输入有效的内网IP'
  },
  sshPort: {
    required: true,
    ...ValidationRules.port,
    message: '请输入SSH端口（1-65535）'
  },
  env: {
    ...ValidationRules.env,
    message: '请选择环境'
  },
  status: {
    ...ValidationRules.serverStatus,
    message: '请选择状态'
  },
  os: {
    max: 50,
    message: '操作系统名称不能超过50个字符'
  },
  arch: {
    max: 50,
    message: '系统架构不能超过50个字符'
  },
  cpu: {
    type: 'number',
    min: 1,
    message: 'CPU核心数必须大于0'
  },
  memory: {
    type: 'number',
    min: 1,
    message: '内存大小必须大于0'
  },
  disk: {
    type: 'number',
    min: 1,
    message: '磁盘大小必须大于0'
  },
  remarks: {
    max: 500,
    message: '备注不能超过500个字符'
  }
};

/**
 * 用户表单验证规则
 */
export const UserFormRules = {
  username: {
    required: true,
    ...ValidationRules.username,
    message: '请输入用户名'
  },
  password: {
    required: true,
    ...ValidationRules.password,
    message: '请输入密码（至少6位）'
  },
  email: {
    required: true,
    ...ValidationRules.email,
    message: '请输入有效的邮箱地址'
  },
  realName: {
    required: true,
    max: 100,
    message: '真实姓名不能超过100个字符'
  },
  phone: {
    ...ValidationRules.ip,
    message: '请输入有效的手机号码'
  },
  roleIds: {
    required: true,
    type: 'array',
    min: 1,
    message: '请至少选择一个角色'
  }
};

/**
 * 菜单表单验证规则
 */
export const MenuFormRules = {
  name: {
    required: true,
    min: 1,
    max: 50,
    message: '请输入菜单名称'
  },
  path: {
    required: true,
    max: 200,
    message: '请输入菜单路径'
  },
  menuType: {
    required: true,
    enum: ['menu', 'directory'],
    message: '请选择菜单类型'
  },
  permission: {
    max: 100,
    message: '权限标识不能超过100个字符'
  },
  sort: {
    type: 'number',
    min: 0,
    message: '排序值必须大于等于0'
  },
  status: {
    enum: [0, 1],
    message: '请选择状态'
  }
};

/**
 * 表单验证规则类型
 */
type FormRule = ValidationRule & { message?: string };

/**
 * 验证表单数据
 */
export function validateForm<T extends Record<string, unknown>>(
  data: T,
  rules: Record<string, FormRule>
): ValidationResult {
  // 提取验证规则
  const validationRules: Record<string, ValidationRule> = {};

  for (const field in rules) {
    const rule = rules[field];
    if (typeof rule === 'object') {
      validationRules[field] = rule;
      // 提取 message 字段作为错误提示
      if (rule.message) {
        (validationRules[field] as FormRule).message = rule.message;
      }
    }
  }

  return validateObject(data, validationRules);
}

/**
 * 表单验证 Composable
 */
export function useFormValidation<T extends Record<string, unknown>>() {
  const validationErrors = ref<Record<string, string>>({});
  const isValid = ref(true);

  /**
   * 验证表单
   */
  const validate = (data: T, rules: Record<string, FormRule>): boolean => {
    const result = validateForm(data, rules);

    isValid.value = result.valid;
    validationErrors.value = result.errors;

    if (!result.valid) {
      // 显示第一个错误
      const firstField = Object.keys(result.errors)[0];
      ElMessage.error(`${firstField}: ${result.errors[firstField]}`);
    }

    return result.valid;
  };

  /**
   * 清除特定字段的错误
   */
  const clearFieldError = (field: string) => {
    if (validationErrors.value[field]) {
      delete validationErrors.value[field];
    }
  };

  /**
   * 清除所有错误
   */
  const clearErrors = () => {
    validationErrors.value = {};
    isValid.value = true;
  };

  /**
   * 获取字段错误消息
   */
  const getFieldError = (field: string): string | undefined => {
    return validationErrors.value[field];
  };

  /**
   * 设置字段错误
   */
  const setFieldError = (field: string, message: string) => {
    validationErrors.value[field] = message;
    isValid.value = false;
  };

  return {
    isValid,
    validationErrors,
    validate,
    clearFieldError,
    clearErrors,
    getFieldError,
    setFieldError
  };
}
