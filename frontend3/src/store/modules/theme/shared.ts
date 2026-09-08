import { defu } from 'defu';
import { addColorAlpha, getColorPalette, getHex, getPaletteColorByNumber, getRgb, mixColor } from '@sa/color';
import { DARK_CLASS } from '@/constants/app';
import { toggleHtmlClass } from '@/utils/common';
import { localStg } from '@/utils/storage';
import { overrideThemeSettings, themeSettings } from '@/theme/settings';
import { themeVars } from '@/theme/vars';

const CONTENT_CARD_BORDER_MIGRATED_KEY = 'contentCardBorderMigrated';

/** Init theme settings */
export function initThemeSettings() {
  const isProd = import.meta.env.MODE === 'prod';

  // if it is development mode, the theme settings will not be cached, by update `themeSettings` in `src/theme/settings.ts` to update theme settings
  if (!isProd) {
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

  // 内容卡片边框默认值曾短暂为 true；这里只做一次迁移，避免旧缓存覆盖新默认值
  if (localStg.get(CONTENT_CARD_BORDER_MIGRATED_KEY) !== true) {
    settings.content.contentCard.borderVisible = false;
    localStg.set(CONTENT_CARD_BORDER_MIGRATED_KEY, true);
  }

  return settings;
}

/**
 * create theme token css vars value by theme settings
 *
 * @param colors Theme colors
 * @param tokens Theme setting tokens
 * @param [recommended=false] Use recommended color. Default is `false`
 * @param siderCustomColor Custom sider color
 */
interface CreateThemeTokenOptions {
  tokens?: App.Theme.ThemeSetting['tokens'];
  recommended?: boolean;
  siderCustomColor?: string;
}

export function createThemeToken(colors: App.Theme.ThemeColor, options: CreateThemeTokenOptions = {}) {
  const { tokens, recommended = false, siderCustomColor } = options;
  const paletteColors = createThemePaletteColors(colors, recommended);

  const { light, dark } = tokens || themeSettings.tokens;

  const themeTokens: App.Theme.ThemeTokenCSSVars = {
    colors: {
      ...paletteColors,
      nprogress: paletteColors.primary,
      ...light.colors,
      ...(siderCustomColor && { 'sider-custom': siderCustomColor })
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
      ...dark?.colors,
      ...(siderCustomColor && { 'sider-custom': siderCustomColor })
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
      --el-input-border-radius: ${components.input};
      --el-checkbox-border-radius: ${components.checkbox};
      --el-radio-input-border-radius: ${components.radio};
      --el-dialog-border-radius: ${components.modal};
      --el-messagebox-border-radius: ${components.modal};
      --el-notification-radius: ${medium};
    `
      .replace(/\s+/g, ' ')
      .trim();
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
    --border-radius-menu: ${medium};
    --border-radius-tag: ${medium};
    --border-radius-switch: 12px;
    --border-radius-checkbox: 4px;
    --border-radius-radio: 50%;
    --el-input-border-radius: ${small};
    --el-checkbox-border-radius: 4px;
    --el-radio-input-border-radius: 50%;
    --el-dialog-border-radius: ${large};
    --el-messagebox-border-radius: ${large};
    --el-notification-radius: ${medium};
  `
    .replace(/\s+/g, ' ')
    .trim();
}

/** 生成 Element Plus 主色变量，保证 link 按钮等组件跟随主题主色 */
function getElementPlusPrimaryVars(primary: string, dark: boolean) {
  const baseColor = dark ? '#141414' : '#ffffff';

  return `
    --el-color-primary: ${getHex(primary)};
    --el-color-primary-light-3: ${mixColor(primary, baseColor, 0.3)};
    --el-color-primary-light-5: ${mixColor(primary, baseColor, 0.5)};
    --el-color-primary-light-7: ${mixColor(primary, baseColor, 0.7)};
    --el-color-primary-light-8: ${mixColor(primary, baseColor, 0.8)};
    --el-color-primary-light-9: ${mixColor(primary, baseColor, 0.9)};
    --el-color-primary-dark-2: ${mixColor(primary, '#000000', 0.2)};
  `
    .replace(/\s+/g, ' ')
    .trim();
}

/**
 * Add theme vars to global
 *
 * @param tokens
 * @param borderRadius Border radius settings
 */
interface AddThemeVarsOptions {
  borderRadius?: App.Theme.ThemeSetting['borderRadius'];
  themeColors?: App.Theme.ThemeColor;
}

export function addThemeVarsToGlobal(
  tokens: App.Theme.BaseToken,
  darkTokens: App.Theme.BaseToken,
  options: AddThemeVarsOptions = {}
) {
  const cssVarStr = getCssVarByTokens(tokens);
  const darkCssVarStr = getCssVarByTokens(darkTokens);
  const { borderRadius, themeColors } = options;
  const borderRadiusStr = borderRadius ? getBorderRadiusCssVars(borderRadius) : '';

  const primaryVars = themeColors ? ` ${getElementPlusPrimaryVars(themeColors.primary, false)}` : '';
  const darkPrimaryVars = themeColors ? ` ${getElementPlusPrimaryVars(themeColors.primary, true)}` : '';
  const css = `:root { ${cssVarStr} ${borderRadiusStr}${primaryVars} }`;

  const darkCss = `html.${DARK_CLASS} { ${darkCssVarStr}${darkPrimaryVars} }`;

  const styleId = 'theme-vars';

  const style = document.querySelector(`#${styleId}`) || document.createElement('style');

  style.id = styleId;

  style.textContent = css + darkCss;

  document.head.appendChild(style);
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
