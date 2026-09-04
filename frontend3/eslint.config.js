import { defineConfig } from '@soybeanjs/eslint-config';
// oxlint 先行快速检查 correctness 类规则，eslint 侧关闭重复规则避免双报
import oxlint from 'eslint-plugin-oxlint';

export default defineConfig(
  { vue: true, unocss: true },
  {
    rules: {
      'vue/multi-word-component-names': [
        'warn',
        {
          ignores: ['index', 'App', 'Register', '[id]', '[url]']
        }
      ],
      'vue/component-name-in-template-casing': [
        'warn',
        'PascalCase',
        {
          registeredComponentsOnly: false,
          ignores: ['/^icon-/']
        }
      ],
      'unocss/order-attributify': 'off'
    }
  },
  ...oxlint.configs['flat/recommended']
);
