/**
 * UnoCSS 渐变设计系统扩展
 * 为 OneOps 项目提供快捷的渐变工具类
 */

import type { Rule } from 'unocss';

export const gradientDesignSystem: Rule[] = [
  // 渐变背景工具类
  [
    /^bg-gradient-diagonal$/,
    () => ({
      'background': 'linear-gradient(135deg, #ffffff, #f7fbff)',
      'border': '1px solid rgba(148, 163, 184, 0.12)',
      'box-shadow': '0 8px 18px rgba(59, 130, 246, 0.14)',
    }),
  ],
  [
    /^bg-gradient-card$/,
    () => ({
      'background': 'linear-gradient(180deg, #ffffff, #f8fbff)',
      'border': '1px solid rgba(148, 163, 184, 0.12)',
    }),
  ],
  [
    /^bg-gradient-header$/,
    () => ({
      'background': 'linear-gradient(180deg, rgba(248, 251, 255, 0.98), rgba(243, 247, 255, 0.94))',
      'border-bottom': '1px solid rgba(148, 163, 184, 0.12)',
    }),
  ],

  // 渐变边框工具类
  [
    /^border-gradient-light$/,
    () => ({
      'border': '1px solid rgba(148, 163, 184, 0.12) !important',
    }),
  ],
  [
    /^border-gradient-brand$/,
    () => ({
      'border': '1px solid rgba(96, 165, 250, 0.24) !important',
    }),
  ],

  // 渐变阴影工具类
  [
    /^shadow-gradient-sm$/,
    () => ({
      'box-shadow': '0 8px 18px rgba(59, 130, 246, 0.14) !important',
    }),
  ],
  [
    /^shadow-gradient-md$/,
    () => ({
      'box-shadow': '0 12px 24px rgba(59, 130, 246, 0.20) !important',
    }),
  ],
  [
    /^shadow-gradient-lg$/,
    () => ({
      'box-shadow': '0 16px 32px rgba(59, 130, 246, 0.25) !important',
    }),
  ],

  // 渐变圆角工具类
  [
    /^rounded-gradient$/,
    () => ({
      'border-radius': '12px !important',
    }),
  ],
  [
    /^rounded-gradient-sm$/,
    () => ({
      'border-radius': '8px !important',
    }),
  ],

  // 渐变文本工具类
  [
    /^text-gradient-brand$/,
    () => ({
      'background': 'linear-gradient(135deg, #3b82f6, #38bdf8)',
      '-webkit-background-clip': 'text',
      '-webkit-text-fill-color': 'transparent',
      'background-clip': 'text',
    }),
  ],

  // 渐变按钮基础样式
  [
    /^btn-gradient$/,
    () => ({
      'background': 'linear-gradient(135deg, #eff6ffeb, #ecfdf5bd)',
      'border': '1px solid rgba(96, 165, 250, 0.24)',
      'box-shadow': '0 8px 18px rgba(59, 130, 246, 0.14)',
      'color': '#3b82f6',
      'border-radius': '10px',
      'transition': 'all 0.3s ease',
    }),
  ],

  // 渐变卡片基础样式
  [
    /^card-gradient$/,
    () => ({
      'background': 'linear-gradient(135deg, #ffffff, #f7fbff)',
      'border': '1px solid rgba(148, 163, 184, 0.12)',
      'box-shadow': '0 8px 18px rgba(59, 130, 246, 0.14)',
      'border-radius': '12px',
      'transition': 'all 0.3s ease',
    }),
  ],
];

export default gradientDesignSystem;