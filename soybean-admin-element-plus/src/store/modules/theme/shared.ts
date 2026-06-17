import { defu } from 'defu';
import { addColorAlpha, getColorPalette, getPaletteColorByNumber, getRgb } from '@sa/color';
import { DARK_CLASS } from '@/constants/app';
import { toggleHtmlClass } from '@/utils/common';
import { localStg } from '@/utils/storage';
import { overrideThemeSettings, themeSettings } from '@/theme/settings';
import { themeVars } from '@/theme/vars';

/** Init theme settings */
export function initThemeSettings() {
  const isProd = import.meta.env.MODE === 'prod';

  // if it is development mode, the theme settings will not be cached, by update `themeSettings` in `src/theme/settings.ts` to update theme settings
  if (!isProd) {
    // 调试：打印初始化的设置
    console.log('[ThemeInit] themeSettings.borderRadius:', themeSettings.borderRadius);
    return themeSettings;
  }

  // if it is production mode, the theme settings will be cached in localStorage
  // if want to update theme settings when publish new version, please update `overrideThemeSettings` in `src/theme/settings.ts`

  const localSettings = localStg.get('themeSettings');

  let settings = defu(localSettings, themeSettings);

  const isOverride = localStg.get('overrideThemeFlag') === BUILD_TIME;

  if (!isOverride) {
    settings = defu(overrideThemeSettings, settings);
    localStg.set('overrideThemeFlag', BUILD_TIME);
  }

  return settings;
}

/**
 * create theme token css vars value by theme settings
 *
 * @param colors Theme colors
 * @param tokens Theme setting tokens
 * @param [recommended=false] Use recommended color. Default is `false`
 */
export function createThemeToken(
  colors: App.Theme.ThemeColor,
  tokens?: App.Theme.ThemeSetting['tokens'],
  recommended = false
) {
  const paletteColors = createThemePaletteColors(colors, recommended);

  const { light, dark } = tokens || themeSettings.tokens;

  const themeTokens: App.Theme.ThemeTokenCSSVars = {
    colors: {
      ...paletteColors,
      nprogress: paletteColors.primary,
      ...light.colors
    },
    boxShadow: {
      ...light.boxShadow
    },
    borderRadius: {
      ...light.borderRadius
    }
  };

  const darkThemeTokens: App.Theme.ThemeTokenCSSVars = {
    colors: {
      ...themeTokens.colors,
      ...dark?.colors
    },
    boxShadow: {
      ...themeTokens.boxShadow,
      ...dark?.boxShadow
    },
    borderRadius: {
      ...themeTokens.borderRadius,
      ...dark?.borderRadius
    }
  };

  return {
    themeTokens,
    darkThemeTokens
  };
}

/**
 * Create theme palette colors
 *
 * @param colors Theme colors
 * @param [recommended=false] Use recommended color. Default is `false`
 */
function createThemePaletteColors(colors: App.Theme.ThemeColor, recommended = false) {
  const colorKeys = Object.keys(colors) as App.Theme.ThemeColorKey[];
  const colorPaletteVar = {} as App.Theme.ThemePaletteColor;

  colorKeys.forEach(key => {
    const colorMap = getColorPalette(colors[key], recommended);

    colorPaletteVar[key] = colorMap.get(500)!;

    colorMap.forEach((hex, number) => {
      colorPaletteVar[`${key}-${number}`] = hex;
    });
  });

  return colorPaletteVar;
}

/**
 * Get css var by tokens
 *
 * @param tokens Theme base tokens
 */
function getCssVarByTokens(tokens: App.Theme.BaseToken) {
  const styles: string[] = [];

  function removeVarPrefix(value: string) {
    return value.replace('var(', '').replace(')', '');
  }

  function removeRgbPrefix(value: string) {
    return value.replace('rgb(', '').replace(')', '');
  }

  for (const [key, tokenValues] of Object.entries(themeVars)) {
    for (const [tokenKey, tokenValue] of Object.entries(tokenValues)) {
      let cssVarsKey = removeVarPrefix(tokenValue);
      let cssValue = tokens[key][tokenKey];

      if (key === 'colors') {
        cssVarsKey = removeRgbPrefix(cssVarsKey);
        const { r, g, b } = getRgb(cssValue);
        cssValue = `${r} ${g} ${b}`;
      }

      styles.push(`${cssVarsKey}: ${cssValue}`);
    }
  }

  const styleStr = styles.join(';');

  return styleStr;
}

/**
 * Get border radius CSS vars
 *
 * @param borderRadius Border radius settings
 */
export function getBorderRadiusCssVars(borderRadius: App.Theme.ThemeSetting['borderRadius']) {
  const { useComponentSpecific, small, medium, large, components } = borderRadius;

  if (useComponentSpecific) {
    // 使用组件级圆角
    // 同时覆盖 Element Plus 的默认变量和自定义组件变量
    return `
      --border-radius-button: ${components.button};
      --border-radius-input: ${components.input};
      --border-radius-select: ${components.select};
      --border-radius-card: ${components.card};
      --border-radius-table: ${components.table};
      --border-radius-modal: ${components.modal};
      --border-radius-tag: ${components.tag};
      --border-radius-switch: ${components.switch};
      --border-radius-checkbox: ${components.checkbox};
      --border-radius-radio: ${components.radio};
      --border-radius-menu: ${components.menu};
      --border-radius-small: ${small};
      --border-radius-medium: ${medium};
      --border-radius-large: ${large};
      --el-border-radius-base: ${components.button};
      --el-border-radius-small: ${components.input};
      --el-border-radius-round: ${components.tag};
      --el-border-radius-circle: ${components.radio};
    `.replace(/\s+/g, ' ').trim();
  }

  // 使用统一分类圆角
  return `
    --border-radius-small: ${small};
    --border-radius-medium: ${medium};
    --border-radius-large: ${large};
    --el-border-radius-base: ${medium};
    --el-border-radius-small: ${small};
    --border-radius-card: ${large};
    --border-radius-modal: ${large};
    --border-radius-table: ${medium};
    --border-radius-menu: ${medium};
    --border-radius-tag: ${medium};
    --border-radius-switch: 12px;
    --border-radius-checkbox: 4px;
    --border-radius-radio: 50%;
  `.replace(/\s+/g, ' ').trim();
}

/**
 * Add theme vars to global
 *
 * @param tokens
 * @param borderRadius Border radius settings
 */
export function addThemeVarsToGlobal(tokens: App.Theme.BaseToken, darkTokens: App.Theme.BaseToken, borderRadius?: App.Theme.ThemeSetting['borderRadius']) {
  // 调试：入口日志
  console.log('[ThemeVars] ========== addThemeVarsToGlobal called ==========');

  const cssVarStr = getCssVarByTokens(tokens);
  const darkCssVarStr = getCssVarByTokens(darkTokens);
  const borderRadiusStr = borderRadius ? getBorderRadiusCssVars(borderRadius) : '';

  // 调试：打印参数
  console.log('[ThemeVars] borderRadius:', borderRadius);
  console.log('[ThemeVars] borderRadiusStr:', borderRadiusStr);
  console.log('[ThemeVars] full CSS:', cssVarStr + ' ' + borderRadiusStr);

  const css = `:root { ${cssVarStr} ${borderRadiusStr} }`;

  const darkCss = `html.${DARK_CLASS} { ${darkCssVarStr} }`;

  const styleId = 'theme-vars';

  const style = document.querySelector(`#${styleId}`) || document.createElement('style');

  style.id = styleId;

  style.textContent = css + darkCss;

  document.head.appendChild(style);

  console.log('[ThemeVars] ========== addThemeVarsToGlobal completed ==========');
}

/**
 * Toggle css dark mode
 *
 * @param darkMode Is dark mode
 */
export function toggleCssDarkMode(darkMode = false) {
  const { add, remove } = toggleHtmlClass(DARK_CLASS);

  if (darkMode) {
    add();
  } else {
    remove();
  }
}

/**
 * Toggle auxiliary color modes
 *
 * @param grayscaleMode
 * @param colourWeakness
 */
export function toggleAuxiliaryColorModes(grayscaleMode = false, colourWeakness = false) {
  const htmlElement = document.documentElement;
  htmlElement.style.filter = [grayscaleMode ? 'grayscale(100%)' : '', colourWeakness ? 'invert(80%)' : '']
    .filter(Boolean)
    .join(' ');
}

type NaiveColorScene = '' | 'Suppl' | 'Hover' | 'Pressed' | 'Active';
type NaiveColorKey = `${App.Theme.ThemeColorKey}Color${NaiveColorScene}`;
type NaiveThemeColor = Partial<Record<NaiveColorKey, string>>;
interface NaiveColorAction {
  scene: NaiveColorScene;
  handler: (color: string) => string;
}

/**
 * Get naive theme colors
 *
 * @param colors Theme colors
 * @param [recommended=false] Use recommended color. Default is `false`
 */
function getNaiveThemeColors(colors: App.Theme.ThemeColor, recommended = false) {
  const colorActions: NaiveColorAction[] = [
    { scene: '', handler: color => color },
    { scene: 'Suppl', handler: color => color },
    { scene: 'Hover', handler: color => getPaletteColorByNumber(color, 500, recommended) },
    { scene: 'Pressed', handler: color => getPaletteColorByNumber(color, 700, recommended) },
    { scene: 'Active', handler: color => addColorAlpha(color, 0.1) }
  ];

  const themeColors: NaiveThemeColor = {};

  const colorEntries = Object.entries(colors) as [App.Theme.ThemeColorKey, string][];

  colorEntries.forEach(color => {
    colorActions.forEach(action => {
      const [colorType, colorValue] = color;
      const colorKey: NaiveColorKey = `${colorType}Color${action.scene}`;
      themeColors[colorKey] = action.handler(colorValue);
    });
  });

  return themeColors;
}

/**
 * Get naive theme
 *
 * @param colors Theme colors
 * @param [recommended=false] Use recommended color. Default is `false`
 * @param [borderRadius='6px'] Border radius value. Default is `'6px'`
 */
export function getNaiveTheme(colors: App.Theme.ThemeColor, recommended = false, borderRadius = '6px') {
  const { primary: colorLoading } = colors;

  const theme = {
    common: {
      ...getNaiveThemeColors(colors, recommended),
      borderRadius
    },
    LoadingBar: {
      colorLoading
    },
    Tag: {
      borderRadius
    }
  };

  return theme;
}
