/**
 * OneOps OKLCH颜色系统迁移指南
 *
 * OKLCH是更现代的颜色空间，具有更好的感知均匀性
 *
 * 文档版本: 1.0.0
 * 更新日期: 2026-06-29
 */

// ============ 迁移步骤 ============

/**
 * 第一步：在主题配置中启用OKLCH
 */
// 修改 src/theme/settings.ts
export const themeSettings: App.Theme.ThemeSetting = {
  // ... 现有配置保持不变

  // 新增OKLCH配置
  oklch: {
    enabled: true,           // 启用OKLCH
    fallbackToHex: true,     // 不支持时回退到HEX
    preferOklch: true         // 优先使用OKLCH格式输出
  }
};

/**
 * 第二步：在主题Store中集成OKLCH
 */
// 在 src/store/modules/theme/index.ts 中添加
import { initOklchTheme, applyOklchThemeToDom, getOklchAdapter } from '@sa/color';

// 在 useThemeStore 中添加OKLCH支持
export const useThemeStore = defineStore(SetupStoreId.Theme, () => {
  // ... 现有代码

  /** OKLCH支持 */
  const oklchEnabled = computed(() => settings.value.oklch?.enabled ?? false);
  const oklchAdapter = computed(() => {
    if (!oklchEnabled.value) return null;

    const adapter = initOklchTheme({
      enableOklch: true,
      fallbackToHex: settings.value.oklch?.fallbackToHex ?? true,
      colors: {
        primary: settings.value.themeColor,
        info: settings.value.otherColor.info,
        success: settings.value.otherColor.success,
        warning: settings.value.otherColor.warning,
        error: settings.value.otherColor.error
      }
    });

    return adapter;
  });

  /** 应用OKLCH主题到DOM */
  function applyOklchTheme() {
    if (oklchEnabled.value && oklchAdapter.value) {
      applyOklchThemeToDom();
      console.log('[OKLCH] Theme applied successfully');
    }
  }

  // 监听主题变化，自动应用OKLCH
  watch([themeColors, oklchEnabled], () => {
    applyOklchTheme();
  }, { immediate: true });

  return {
    // ... 现有返回值
    oklchEnabled,
    oklchAdapter,
    applyOklchTheme
  };
});

/**
 * 第三步：修改CSS变量生成逻辑
 */
// 在 src/store/modules/theme/shared.ts 中修改
import { getOklchAdapter } from '@sa/color';

export function addThemeVarsToGlobal(
  tokens: App.Theme.BaseToken,
  darkTokens: App.Theme.BaseToken,
  borderRadius?: App.Theme.ThemeSetting['borderRadius']
) {
  const cssVarStr = getCssVarByTokens(tokens);
  const darkCssVarStr = getCssVarByTokens(darkTokens);
  const borderRadiusStr = borderRadius ? getBorderRadiusCssVars(borderRadius) : '';

  // 检查是否使用OKLCH
  const oklchAdapter = getOklchAdapter();
  let oklchCssVars = '';

  if (oklchAdapter && oklchAdapter.shouldUseOklch()) {
    oklchCssVars = oklchAdapter.generateThemeVars();
    console.log('[ThemeVars] Using OKLCH color format');
  }

  const css = `:root { ${cssVarStr} ${borderRadiusStr} ${oklchCssVars} }`;
  const darkCss = `html.${DARK_CLASS} { ${darkCssVarStr} }`;

  const styleId = 'theme-vars';
  const style = document.querySelector(`#${styleId}`) || document.createElement('style');
  style.id = styleId;
  style.textContent = css + darkCss;
  document.head.appendChild(style);
}

// ============ 使用示例 ============

/**
 * 示例1：在组件中使用OKLCH颜色
 */
<template>
  <div class="oklch-demo">
    <!-- 使用CSS变量 -->
    <button class="btn-primary">主要按钮</button>
    <button class="btn-success">成功按钮</button>

    <!-- 直接使用OKLCH值 -->
    <div class="custom-color" :style="{ color: customColor }">
      自定义颜色
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useThemeStore } from '@/store/modules/theme';
import { hexToOklch, oklchToCssString } from '@sa/color';

const themeStore = useThemeStore();

// 获取OKLCH格式的主题颜色
const primaryColor = computed(() => {
  if (!themeStore.oklchAdapter) {
    return themeStore.themeColors.primary;
  }

  return themeStore.oklchAdapter.getPrimaryColor();
});

// 自定义OKLCH颜色
const customHex = '#e81a65';
const customOklch = hexToOklch(customHex);
const customColor = computed(() => oklchToCssString(customOklch));
</script>

<style scoped>
.oklch-demo {
  /* 使用OKLCH CSS变量 */
  --my-primary: var(--primary-500);
  --my-success: var(--other-success-500);
}

.btn-primary {
  background-color: var(--my-primary);
  color: white;
}

.btn-success {
  background-color: var(--my-success);
  color: white;
}

/* OKLCH的优势：精确控制颜色调整 */
.custom-color:hover {
  /* 轻微增加亮度 */
  color: color-mix(in oklch, var(--my-primary) 90%, white 10%);
}

.custom-color:active {
  /* 降低色度，保持色相 */
  color: color-mix(in oklch, var(--my-primary) 80%, black 20%);
}
</style>

/**
 * 示例2：动态主题切换
 */
const toggleOklchTheme = () => {
  const themeStore = useThemeStore();

  // 切换OKLCH开关
  themeStore.updateThemeSettings({
    oklch: {
      enabled: !themeStore.oklchEnabled,
      fallbackToHex: true,
      preferOklch: true
    }
  });

  // 重新应用主题
  themeStore.applyOklchTheme();
};

/**
 * 示例3：生成品牌色彩方案
 */
import { generateOklchPalette, type PaletteLevel } from '@sa/color';

// 为品牌生成完整色彩方案
const brandColors = generateOklchPalette({
  baseColor: '#e81a65', // OneOps品牌色
  lightnessRange: [0.2, 0.95],
  chromaScale: 1.0,
  preserveHue: true
});

// 输出为CSS变量
const cssVariables = brandColors.map(color =>
  `--brand-${color.level}: ${color.cssString};`
).join('\n');

console.log(cssVariables);

/**
 * 示例4：检测浏览器支持
 */
import { OklchThemeAdapter } from '@sa/color';

const supportsOklch = OklchThemeAdapter.detectOklchSupport();
console.log('Browser OKLCH support:', supportsOklch);

if (!supportsOklch) {
  console.warn('OKLCH is not supported, falling back to HEX');
}

// ============ 迁移检查清单 ============

/**
 * ✅ 迁移前检查
 */
export const MIGRATION_CHECKLIST = {
  compatibility: [
    '检查目标浏览器是否支持OKLCH (Chrome 111+, Firefox 113+, Safari 15.2+)',
    '确认是否需要回退到HEX格式',
    '测试现有颜色在OKLCH下的显示效果'
  ],

  implementation: [
    '更新theme/settings.ts配置',
    '修改theme store集成OKLCH',
    '更新CSS变量生成逻辑',
    '测试组件颜色显示正确性'
  ],

  testing: [
    '测试亮色/暗色主题切换',
    '验证颜色渐变效果',
    '检查浏览器兼容性回退',
    '性能测试（颜色转换开销）'
  ],

  deployment: [
    '渐进式启用OKLCH',
    '监控用户浏览器支持情况',
    '准备回退方案',
    '更新相关文档'
  ]
};

// ============ 常见问题解答 ============

/**
 * Q: OKLCH与现有RGB系统会冲突吗？
 * A: 不会。OKLCH作为附加层存在，现有RGB系统继续工作，
 *    在支持OKLCH的浏览器中会优先使用OKLCH格式。
 *
 * Q: 性能会有影响吗？
 * A: 影响极小。OKLCH转换只在主题初始化时进行一次，
 *    之后使用缓存的CSS变量。
 *
 * Q: 如何确保在不支持的浏览器中正常显示？
 * A: 设置fallbackToHex: true，系统会自动检测浏览器支持
 *    情况并回退到HEX格式。
 *
 * Q: 现有组件需要修改吗？
 * A: 大部分不需要。如果使用了CSS变量，会自动应用OKLCH格式。
 *    只有硬编码颜色的地方才需要修改。
 */

export default migrationGuide;