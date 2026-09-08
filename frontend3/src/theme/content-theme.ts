/**
 * Content theme utilities
 * Unified content area CSS variable injection based on variant
 */

type ContentThemeSettings = App.Theme.ThemeSetting['content'];
import { themeSettings } from './settings';

export type { ContentThemeSettings };

type CssVarMap = Record<string, string>;

/** 圆角统一由主题设置控制，内容区只负责映射到对应的组件变量 */
function radiusVar(token: string, fallback: string): string {
  return `var(--border-radius-${token}, ${fallback})`;
}

/** Build a gradient or solid background value */
function buildBg(
  useGradient: boolean,
  solid: string,
  start: string,
  middle: string,
  end: string,
  angle: number
): string {
  if (useGradient && start && end) {
    if (middle) {
      return `linear-gradient(${angle}deg, ${start} 0%, ${middle} 50%, ${end} 100%)`;
    }
    return `linear-gradient(${angle}deg, ${start} 0%, ${end} 100%)`;
  }
  return solid || start;
}

/** Map content theme config to CSS custom properties */
export function buildContentVars(theme: ContentThemeSettings): CssVarMap {
  const vars: CssVarMap = {};

  // Hero section
  vars['--msre-hero-visible'] = theme.hero.visible ? '1' : '0';
  vars['--msre-hero-bg'] = buildBg(
    theme.variant === 'modern',
    theme.hero.background,
    theme.hero.gradientStart,
    theme.hero.gradientMiddle,
    theme.hero.gradientEnd,
    theme.hero.gradientAngle
  );
  vars['--msre-hero-border'] = theme.hero.borderColor;
  vars['--msre-hero-radius'] = radiusVar('card', '12px');
  vars['--msre-hero-shadow'] = theme.hero.shadow;
  vars['--msre-hero-padding'] = theme.hero.padding;
  vars['--msre-hero-icon-bg'] =
    `linear-gradient(180deg, ${theme.hero.iconGradientStart} 0%, ${theme.hero.iconGradientEnd} 100%)`;
  vars['--msre-hero-icon-border'] = theme.hero.iconBorderColor;
  vars['--msre-hero-icon-color'] = theme.hero.iconColor;

  // Statistics cards
  const stat = theme.statCards;
  if (stat.useGradient) {
    vars['--msre-stat-default-bg'] = `linear-gradient(145deg, ${stat.defaultBgStart} 0%, ${stat.defaultBgEnd} 100%)`;
    vars['--msre-stat-success-bg'] = `linear-gradient(145deg, ${stat.successBgStart} 0%, ${stat.successBgEnd} 100%)`;
    vars['--msre-stat-warning-bg'] = `linear-gradient(145deg, ${stat.warningBgStart} 0%, ${stat.warningBgEnd} 100%)`;
    vars['--msre-stat-danger-bg'] = `linear-gradient(145deg, ${stat.dangerBgStart} 0%, ${stat.dangerBgEnd} 100%)`;
  } else {
    vars['--msre-stat-default-bg'] = stat.defaultBg;
    vars['--msre-stat-success-bg'] = stat.successBg;
    vars['--msre-stat-warning-bg'] = stat.warningBg;
    vars['--msre-stat-danger-bg'] = stat.dangerBg;
  }
  vars['--msre-stat-default-border'] = stat.border;
  vars['--msre-stat-radius'] = radiusVar('card', '12px');
  vars['--msre-stat-shadow'] = stat.shadow;
  vars['--msre-stat-padding'] = stat.padding;

  // Content card
  vars['--msre-content-card-bg'] = buildBg(
    theme.variant === 'modern',
    theme.contentCard.background,
    theme.contentCard.gradientStart,
    theme.contentCard.gradientMiddle,
    theme.contentCard.gradientEnd,
    theme.contentCard.gradientAngle
  );
  vars['--msre-content-card-border'] = theme.contentCard.borderColor;
  vars['--msre-content-card-radius'] = radiusVar('card', '12px');
  vars['--msre-content-card-shadow'] = theme.contentCard.shadow;
  vars['--msre-content-card-padding'] = theme.contentCard.padding;

  // Toolbar
  vars['--msre-toolbar-bg'] = buildBg(
    theme.variant === 'modern',
    theme.toolbar.background,
    theme.toolbar.gradientStart,
    theme.toolbar.gradientMiddle,
    theme.toolbar.gradientEnd,
    theme.toolbar.gradientAngle
  );
  vars['--msre-toolbar-border'] = theme.toolbar.borderColor;
  vars['--msre-toolbar-radius'] = radiusVar('card', '12px');
  vars['--msre-toolbar-padding'] = theme.toolbar.padding;
  vars['--msre-toolbar-shadow'] = theme.toolbar.shadow;

  // Data table
  vars['--msre-table-header-bg'] = theme.dataTable.headerBg;
  vars['--msre-table-header-text'] = theme.dataTable.headerTextColor;
  vars['--msre-table-header-border'] = theme.dataTable.headerBorderColor;
  vars['--msre-table-hover-bg'] = theme.dataTable.rowHoverBg;
  vars['--msre-table-row-border'] = theme.dataTable.rowBorderColor;
  vars['--msre-table-border'] = theme.dataTable.borderColor;
  vars['--msre-table-striped-bg'] = theme.dataTable.stripedBg;

  // Search filters
  vars['--msre-search-input-bg'] = theme.searchFilters.inputBg;
  vars['--msre-search-input-border'] = theme.searchFilters.inputBorder;
  vars['--msre-search-input-hover-border'] = theme.searchFilters.inputHoverBorder;
  vars['--msre-search-input-focus-border'] = theme.searchFilters.inputFocusBorder;
  vars['--msre-search-input-radius'] = radiusVar('input', '12px');
  vars['--msre-search-button-bg'] = theme.searchFilters.buttonBg;
  vars['--msre-search-button-text'] = theme.searchFilters.buttonTextColor;
  vars['--msre-search-button-hover'] = theme.searchFilters.buttonHoverBg;

  // Pagination
  vars['--msre-pagination-button-bg'] = theme.pagination.buttonBg;
  vars['--msre-pagination-button-text'] = theme.pagination.buttonTextColor;
  vars['--msre-pagination-button-hover'] = theme.pagination.buttonHoverBg;
  vars['--msre-pagination-active-bg'] = theme.pagination.activeBg;
  vars['--msre-pagination-active-text'] = theme.pagination.activeTextColor;
  vars['--msre-pagination-radius'] = radiusVar('small', '8px');

  // Tags
  vars['--msre-tag-default-bg'] = theme.tags.defaultBg;
  vars['--msre-tag-default-border'] = theme.tags.defaultBorder;
  vars['--msre-tag-default-text'] = theme.tags.defaultTextColor;
  vars['--msre-tag-success-bg'] = theme.tags.successBg;
  vars['--msre-tag-warning-bg'] = theme.tags.warningBg;
  vars['--msre-tag-danger-bg'] = theme.tags.dangerBg;
  vars['--msre-tag-info-bg'] = theme.tags.infoBg;
  vars['--msre-tag-radius'] = radiusVar('tag', '8px');

  // Shared button style
  vars['--msre-button-radius'] = radiusVar('button', '10px');
  vars['--msre-button-default-bg'] = theme.button.defaultBg;
  vars['--msre-button-default-color'] = theme.button.defaultColor;
  vars['--msre-button-default-border'] = theme.button.defaultBorder;
  vars['--msre-button-hover-bg'] = theme.button.hoverBg;
  vars['--msre-button-hover-color'] = theme.button.hoverColor;
  vars['--msre-button-hover-border'] = theme.button.hoverBorder;

  // Shared input style
  vars['--msre-input-radius'] = radiusVar('input', '12px');
  vars['--msre-input-bg'] = theme.input.bg;
  vars['--msre-input-border-shadow'] = theme.input.borderShadow;
  vars['--msre-input-hover-shadow'] = theme.input.hoverShadow;
  vars['--msre-input-focus-shadow'] = theme.input.focusShadow;

  // Shared card style
  vars['--msre-card-bg'] = theme.card.bg;
  vars['--msre-card-border'] = theme.card.border;
  vars['--msre-card-radius'] = radiusVar('card', '12px');
  vars['--msre-card-shadow'] = theme.card.shadow;
  vars['--msre-card-padding'] = theme.card.padding;

  // Colors
  vars['--msre-primary'] = theme.colors.primary;
  vars['--msre-primary-light'] = theme.colors.primaryLight;
  vars['--msre-success'] = theme.colors.success;
  vars['--msre-warning'] = theme.colors.warning;
  vars['--msre-danger'] = theme.colors.danger;
  vars['--msre-info'] = theme.colors.info;

  // Text
  vars['--msre-text-primary'] = theme.text.primary;
  vars['--msre-text-secondary'] = theme.text.secondary;
  vars['--msre-text-muted'] = theme.text.muted;

  // Borders
  vars['--msre-border-soft'] = theme.border.soft;
  vars['--msre-border-medium'] = theme.border.medium;

  return vars;
}

/** Apply content theme CSS variables to document root */
export function applyContentTheme(theme: ContentThemeSettings) {
  const root = document.documentElement;
  const vars = buildContentVars(theme);
  root.setAttribute('data-msre-table-style', theme.dataTable.tableStyle);
  for (const [key, value] of Object.entries(vars)) {
    root.style.setProperty(key, value);
  }
}

/**
 * Apply header theme to CSS variables
 * Ensures header internal elements follow header color scheme
 */
export function applyHeaderTheme(headerTheme: App.Theme.ThemeSetting['header']) {
  const root = document.documentElement;

  let headerBg = '#ffffff';
  let isGradient = false;

  if (headerTheme.useHeaderGradient && headerTheme.headerGradientStart && headerTheme.headerGradientEnd) {
    isGradient = true;
    const gradientStart = headerTheme.headerGradientStart;
    const startAlpha = parseAlpha(gradientStart) || 0.98;

    root.style.setProperty(
      '--msre-header-button-default-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.13, 0.72).toFixed(2)})`
    );
    root.style.setProperty(
      '--msre-header-button-hover-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.03, 0.85).toFixed(2)})`
    );
    root.style.setProperty(
      '--msre-header-input-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.23, 0.62).toFixed(2)})`
    );
    root.style.setProperty(
      '--msre-header-status-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.38, 0.47).toFixed(2)})`
    );
  } else if (headerTheme.useCustomColor && headerTheme.customColor) {
    headerBg = headerTheme.customColor;
    const brightness = getColorBrightness(headerBg);

    if (brightness > 128) {
      root.style.setProperty('--msre-header-button-default-bg', 'rgba(255, 255, 255, 0.85)');
      root.style.setProperty('--msre-header-button-hover-bg', 'rgba(255, 255, 255, 0.95)');
      root.style.setProperty('--msre-header-input-bg', 'rgba(255, 255, 255, 0.75)');
      root.style.setProperty('--msre-header-status-bg', 'rgba(255, 255, 255, 0.6)');
    } else {
      root.style.setProperty('--msre-header-button-default-bg', 'rgba(255, 255, 255, 0.15)');
      root.style.setProperty('--msre-header-button-hover-bg', 'rgba(255, 255, 255, 0.25)');
      root.style.setProperty('--msre-header-input-bg', 'rgba(255, 255, 255, 0.12)');
      root.style.setProperty('--msre-header-status-bg', 'rgba(255, 255, 255, 0.12)');
    }
  } else {
    root.style.setProperty('--msre-header-button-default-bg', 'rgba(255, 255, 255, 0.85)');
    root.style.setProperty('--msre-header-button-hover-bg', 'rgba(255, 255, 255, 0.95)');
    root.style.setProperty('--msre-header-input-bg', 'rgba(255, 255, 255, 0.75)');
    root.style.setProperty('--msre-header-status-bg', 'rgba(255, 255, 255, 0.6)');
  }

  if (isGradient || getColorBrightness(headerBg) > 128) {
    root.style.setProperty('--msre-header-button-default-color', '#475569');
    root.style.setProperty('--msre-header-button-hover-color', '#1d4ed8');
    root.style.setProperty('--msre-header-input-color', '#475569');
    root.style.setProperty('--msre-header-breadcrumb-color', '#475569');
    root.style.setProperty('--msre-header-breadcrumb-hover', '#1d4ed8');
    root.style.setProperty('--msre-header-status-text', '#475569');
  } else {
    root.style.setProperty('--msre-header-button-default-color', 'rgba(255, 255, 255, 0.9)');
    root.style.setProperty('--msre-header-button-hover-color', '#ffffff');
    root.style.setProperty('--msre-header-input-color', 'rgba(255, 255, 255, 0.9)');
    root.style.setProperty('--msre-header-breadcrumb-color', 'rgba(255, 255, 255, 0.8)');
    root.style.setProperty('--msre-header-breadcrumb-hover', '#ffffff');
    root.style.setProperty('--msre-header-status-text', 'rgba(255, 255, 255, 0.8)');
  }

  if (isGradient || getColorBrightness(headerBg) > 128) {
    root.style.setProperty('--msre-header-button-default-border', 'rgba(148, 163, 184, 0.1)');
    root.style.setProperty('--msre-header-button-hover-border', 'rgba(59, 130, 246, 0.16)');
    root.style.setProperty('--msre-header-input-border', 'rgba(148, 163, 184, 0.1)');
    root.style.setProperty('--msre-header-input-hover-border', 'rgba(59, 130, 246, 0.14)');
    root.style.setProperty('--msre-header-input-focus-border', 'rgba(37, 99, 235, 0.18)');
  } else {
    root.style.setProperty('--msre-header-button-default-border', 'rgba(255, 255, 255, 0.2)');
    root.style.setProperty('--msre-header-button-hover-border', 'rgba(255, 255, 255, 0.3)');
    root.style.setProperty('--msre-header-input-border', 'rgba(255, 255, 255, 0.2)');
    root.style.setProperty('--msre-header-input-hover-border', 'rgba(255, 255, 255, 0.3)');
    root.style.setProperty('--msre-header-input-focus-border', 'rgba(255, 255, 255, 0.4)');
  }
}

/** Parse alpha value from color string */
function parseAlpha(color: string): number | null {
  const match = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)/);
  if (match && match[4]) {
    return Number.parseFloat(match[4]);
  }
  const hexMatch = color.match(/#([a-f0-9]{6})([a-f0-9]{2})?/i);
  if (hexMatch) {
    return 1;
  }
  return null;
}

/** Get perceived brightness of a color (0-255) */
function getColorBrightness(color: string): number {
  color = color.replace(/\s/g, '');
  const rgbMatch = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
  if (rgbMatch) {
    const r = Number.parseInt(rgbMatch[1], 10);
    const g = Number.parseInt(rgbMatch[2], 10);
    const b = Number.parseInt(rgbMatch[3], 10);
    return (r * 299 + g * 587 + b * 114) / 1000;
  }
  const hexMatch = color.match(/#?([a-f0-9]{2})([a-f0-9]{2})([a-f0-9]{2})?/i);
  if (hexMatch) {
    const r = Number.parseInt(hexMatch[1], 16);
    const g = Number.parseInt(hexMatch[2], 16);
    const b = Number.parseInt(hexMatch[3] || '', 16);
    return (r * 299 + g * 587 + b * 114) / 1000;
  }
  return 255;
}

/** Initialize content theme with default settings */
export function initContentTheme() {
  applyContentTheme(themeSettings.content);
}

/** Initialize header theme with default settings */
export function initHeaderTheme() {
  applyHeaderTheme(themeSettings.header);
}
