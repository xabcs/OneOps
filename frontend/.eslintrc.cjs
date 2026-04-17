module.exports = {
  root: true,
  env: {
    node: true,
    browser: true,
    es2021: true,
  },
  extends: [
    'plugin:vue/vue3-recommended',
    'eslint:recommended',
  ],
  parserOptions: {
    ecmaVersion: 2021,
    sourceType: 'module',
  },
  rules: {
    // Vue相关规则
    'vue/multi-word-component-names': 'off', // 允许单词组件名
    'vue/no-v-html': 'warn', // v-html使用警告

    // 通用规则
    'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off', // 生产环境警告console
    'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off', // 生产环境警告debugger
    'no-unused-vars': 'warn', // 未使用的变量警告
    'prefer-const': 'warn', // 建议使用const
    'no-var': 'error', // 禁止使用var
  },
}
