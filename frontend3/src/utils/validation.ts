/**
 * 参数验证工具
 * 不依赖外部库的轻量级验证
 */

/**
 * 验证结果
 */
export interface ValidationResult {
  valid: boolean;
  errors: Record<string, string>;
}

/**
 * 验证规则
 */
export interface ValidationRule {
  required?: boolean;
  min?: number;
  max?: number;
  pattern?: RegExp;
  enum?: string[];
  type?: 'string' | 'number' | 'boolean' | 'array' | 'email' | 'ip' | 'url';
}

/**
 * 验证器类
 */
export class Validator {
  private errors: Record<string, string> = {};

  /**
   * 添加验证错误
   */
  addError(field: string, message: string): void {
    this.errors[field] = message;
  }

  /**
   * 验证字符串字段
   */
  validateString(value: unknown, rules: ValidationRule): boolean {
    if (value === undefined || value === null) {
      if (rules.required) {
        this.addError('field', '此字段为必填项');
        return false;
      }
      return true;
    }

    const str = String(value);

    // min 长度验证
    if (rules.min !== undefined && str.length < rules.min) {
      this.addError('field', `长度不能少于${rules.min}个字符`);
      return false;
    }

    // max 长度验证
    if (rules.max !== undefined && str.length > rules.max) {
      this.addError('field', `长度不能超过${rules.max}个字符`);
      return false;
    }

    // pattern 正则验证
    if (rules.pattern && !rules.pattern.test(str)) {
      this.addError('field', '格式不正确');
      return false;
    }

    // enum 枚举验证
    if (rules.enum && !rules.enum.includes(str)) {
      this.addError('field', `值必须是以下之一：${rules.enum.join(', ')}`);
      return false;
    }

    // type 类型验证
    switch (rules.type) {
      case 'email':
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(str)) {
          this.addError('field', '请输入有效的邮箱地址');
          return false;
        }
        break;
      case 'ip':
        const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/;
        if (!ipRegex.test(str)) {
          this.addError('field', '请输入有效的IP地址');
          return false;
        }
        break;
      case 'url':
        try {
          new URL(str);
        } catch {
          this.addError('field', '请输入有效的URL');
          return false;
        }
        break;
    }

    return true;
  }

  /**
   * 验证数字字段
   */
  validateNumber(value: unknown, rules: ValidationRule): boolean {
    if (value === undefined || value === null) {
      if (rules.required) {
        this.addError('field', '此字段为必填项');
        return false;
      }
      return true;
    }

    const num = Number(value);

    // min 范围验证
    if (rules.min !== undefined && num < rules.min) {
      this.addError('field', `不能小于${rules.min}`);
      return false;
    }

    // max 范围验证
    if (rules.max !== undefined && num > rules.max) {
      this.addError('field', `不能大于${rules.max}`);
      return false;
    }

    return true;
  }

  /**
   * 验证数组字段
   */
  validateArray(value: unknown, rules: ValidationRule): boolean {
    if (value === undefined || value === null) {
      if (rules.required) {
        this.addError('field', '此字段为必填项');
        return false;
      }
      return true;
    }

    if (!Array.isArray(value)) {
      this.addError('field', '必须是数组');
      return false;
    }

    // min 长度验证
    if (rules.min !== undefined && value.length < rules.min) {
      this.addError('field', `至少选择${rules.min}项`);
      return false;
    }

    return true;
  }

  /**
   * 获取验证结果
   */
  getResult(): ValidationResult {
    return {
      valid: Object.keys(this.errors).length === 0,
      errors: this.errors
    };
  }

  /**
   * 重置验证器
   */
  reset(): void {
    this.errors = {};
  }
}

/**
 * 验证对象
 */
export function validateObject(obj: Record<string, unknown>, rules: Record<string, ValidationRule>): ValidationResult {
  const validator = new Validator();

  for (const field in rules) {
    const value = obj[field];
    const rule = rules[field];

    if (rule.type === 'number') {
      validator.validateNumber(value, { ...rule, type: undefined });
    } else if (rule.type === 'array') {
      validator.validateArray(value, { ...rule, type: undefined });
    } else {
      validator.validateString(value, rule);
    }
  }

  return validator.getResult();
}

/**
 * 常用验证规则
 */
export const ValidationRules = {
  // 主机名
  hostname: {
    pattern: /^[a-zA-Z0-9-]+$/,
    max: 100
  },

  // IP 地址
  ip: {
    pattern: /^(\d{1,3}\.){3}\d{1,3}$/,
    max: 15
  },

  // 端口
  port: {
    type: 'number',
    min: 1,
    max: 65535
  },

  // 邮箱
  email: {
    type: 'email'
  },

  // 分页
  page: {
    type: 'number',
    min: 1,
    default: 1
  },
  pageSize: {
    type: 'number',
    min: 1,
    max: 100,
    default: 20
  },

  // 状态
  serverStatus: {
    enum: ['active', 'inactive', 'maintenance'],
    default: 'active'
  },

  // 环境
  env: {
    enum: ['dev', 'test', 'prod'],
    default: 'prod'
  },

  // 用户名
  username: {
    pattern: /^[a-zA-Z0-9_]+$/,
    min: 3,
    max: 50
  },

  // 密码
  password: {
    min: 6,
    max: 100
  },

  // 数组
  roleIds: {
    type: 'array',
    min: 1
  },
  groupIds: {
    type: 'array',
    min: 1
  }
};
