/**
 * Content theme utilities
 * Applies content theme settings to CSS variables
 */

import { themeSettings } from '@/theme/settings';

export type ContentThemeSettings = App.Theme.ThemeSetting['contentTheme'];

/**
 * Apply content theme to CSS variables
 */
export function applyContentTheme(theme: ContentThemeSettings) {
  const root = document.documentElement;

  // Card styles
  if (theme.cardBg) {
    root.style.setProperty('--sx-card-bg', theme.cardBg);
  }
  if (theme.cardRadius) {
    root.style.setProperty('--sx-card-radius', theme.cardRadius);
  }
  if (theme.cardShadow) {
    root.style.setProperty('--sx-card-shadow', theme.cardShadow);
  }

  // Card gradient
  if (theme.useCardGradient && theme.cardGradientStart && theme.cardGradientEnd) {
    root.style.setProperty(
      '--sx-card-bg',
      `linear-gradient(145deg, ${theme.cardGradientStart} 0%, ${theme.cardGradientEnd} 100%)`
    );
  }

  // Table styles
  if (theme.tableBorder) {
    root.style.setProperty('--sx-table-border', theme.tableBorder);
  }
  if (theme.tableHeaderBg) {
    root.style.setProperty('--sx-table-header-bg', theme.tableHeaderBg);
  }
  if (theme.tableHoverBg) {
    root.style.setProperty('--sx-table-hover-bg', theme.tableHoverBg);
  }
  if (theme.tableRadius) {
    root.style.setProperty('--sx-table-radius', theme.tableRadius);
  }

  // Button styles
  if (theme.buttonRadius) {
    root.style.setProperty('--sx-button-radius', theme.buttonRadius);
  }
  if (theme.buttonDefaultBg) {
    root.style.setProperty('--sx-button-default-bg', theme.buttonDefaultBg);
  }
  if (theme.buttonDefaultColor) {
    root.style.setProperty('--sx-button-default-color', theme.buttonDefaultColor);
  }
  if (theme.buttonDefaultBorder) {
    root.style.setProperty('--sx-button-default-border', theme.buttonDefaultBorder);
  }
  if (theme.buttonHoverBg) {
    root.style.setProperty('--sx-button-hover-bg', theme.buttonHoverBg);
  }
  if (theme.buttonHoverColor) {
    root.style.setProperty('--sx-button-hover-color', theme.buttonHoverColor);
  }
  if (theme.buttonHoverBorder) {
    root.style.setProperty('--sx-button-hover-border', theme.buttonHoverBorder);
  }

  // Input styles
  if (theme.inputRadius) {
    root.style.setProperty('--sx-input-radius', theme.inputRadius);
  }
  if (theme.inputBg) {
    root.style.setProperty('--sx-input-bg', theme.inputBg);
  }
  if (theme.inputBorderShadow) {
    root.style.setProperty('--sx-input-border-shadow', theme.inputBorderShadow);
  }
  if (theme.inputHoverShadow) {
    root.style.setProperty('--sx-input-hover-shadow', theme.inputHoverShadow);
  }
  if (theme.inputFocusShadow) {
    root.style.setProperty('--sx-input-focus-shadow', theme.inputFocusShadow);
  }

  // Toolbar gradient
  if (theme.useToolbarGradient && theme.toolbarGradientStart && theme.toolbarGradientEnd) {
    root.style.setProperty(
      '--sx-toolbar-bg',
      `linear-gradient(180deg, ${theme.toolbarGradientStart} 0%, ${theme.toolbarGradientEnd} 100%)`
    );
  } else {
    root.style.removeProperty('--sx-toolbar-bg');
  }
}

/**
 * Apply header theme to CSS variables
 * This ensures header internal elements follow header color scheme
 */
export function applyHeaderTheme(headerTheme: App.Theme.ThemeSetting['header']) {
  const root = document.documentElement;

  // Extract background color from header style
  let headerBg = '#ffffff';
  let isGradient = false;

  if (headerTheme.useHeaderGradient && headerTheme.headerGradientStart && headerTheme.headerGradientEnd) {
    // Header uses gradient - calculate appropriate colors for internal elements
    isGradient = true;
    // For gradient headers, use slightly darker colors for internal elements
    const gradientStart = headerTheme.headerGradientStart;
    const gradientEnd = headerTheme.headerGradientEnd;

    // Parse alpha values and create adjusted colors
    const startAlpha = parseAlpha(gradientStart) || 0.98;
    const endAlpha = parseAlpha(gradientEnd) || 0.94;

    root.style.setProperty(
      '--header-button-default-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.13, 0.72).toFixed(2)})`
    );
    root.style.setProperty(
      '--header-button-hover-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.03, 0.85).toFixed(2)})`
    );
    root.style.setProperty(
      '--header-input-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.23, 0.62).toFixed(2)})`
    );
    root.style.setProperty(
      '--header-status-bg',
      `rgba(255, 255, 255, ${Math.max(startAlpha - 0.38, 0.47).toFixed(2)})`
    );
  } else if (headerTheme.useCustomColor && headerTheme.customColor) {
    // Header uses custom solid color
    headerBg = headerTheme.customColor;
    const brightness = getColorBrightness(headerBg);

    // Adjust internal element colors based on header background brightness
    if (brightness > 128) {
      // Light header
      root.style.setProperty('--header-button-default-bg', 'rgba(255, 255, 255, 0.85)');
      root.style.setProperty('--header-button-hover-bg', 'rgba(255, 255, 255, 0.95)');
      root.style.setProperty('--header-input-bg', 'rgba(255, 255, 255, 0.75)');
      root.style.setProperty('--header-status-bg', 'rgba(255, 255, 255, 0.6)');
    } else {
      // Dark header
      root.style.setProperty('--header-button-default-bg', 'rgba(255, 255, 255, 0.15)');
      root.style.setProperty('--header-button-hover-bg', 'rgba(255, 255, 255, 0.25)');
      root.style.setProperty('--header-input-bg', 'rgba(255, 255, 255, 0.12)');
      root.style.setProperty('--header-status-bg', 'rgba(255, 255, 255, 0.12)');
    }
  } else {
    // Default header (no custom color)
    root.style.setProperty('--header-button-default-bg', 'rgba(255, 255, 255, 0.85)');
    root.style.setProperty('--header-button-hover-bg', 'rgba(255, 255, 255, 0.95)');
    root.style.setProperty('--header-input-bg', 'rgba(255, 255, 255, 0.75)');
    root.style.setProperty('--header-status-bg', 'rgba(255, 255, 255, 0.6)');
  }

  // Set text colors based on header brightness
  if (isGradient || getColorBrightness(headerBg) > 128) {
    root.style.setProperty('--header-button-default-color', '#475569');
    root.style.setProperty('--header-button-hover-color', '#1d4ed8');
    root.style.setProperty('--header-input-color', '#475569');
    root.style.setProperty('--header-breadcrumb-color', '#475569');
    root.style.setProperty('--header-breadcrumb-hover', '#1d4ed8');
    root.style.setProperty('--header-status-text', '#475569');
  } else {
    root.style.setProperty('--header-button-default-color', 'rgba(255, 255, 255, 0.9)');
    root.style.setProperty('--header-button-hover-color', '#ffffff');
    root.style.setProperty('--header-input-color', 'rgba(255, 255, 255, 0.9)');
    root.style.setProperty('--header-breadcrumb-color', 'rgba(255, 255, 255, 0.8)');
    root.style.setProperty('--header-breadcrumb-hover', '#ffffff');
    root.style.setProperty('--header-status-text', 'rgba(255, 255, 255, 0.8)');
  }

  // Border colors
  if (isGradient || getColorBrightness(headerBg) > 128) {
    root.style.setProperty('--header-button-default-border', 'rgba(148, 163, 184, 0.1)');
    root.style.setProperty('--header-button-hover-border', 'rgba(59, 130, 246, 0.16)');
    root.style.setProperty('--header-input-border', 'rgba(148, 163, 184, 0.1)');
    root.style.setProperty('--header-input-hover-border', 'rgba(59, 130, 246, 0.14)');
    root.style.setProperty('--header-input-focus-border', 'rgba(37, 99, 235, 0.18)');
  } else {
    root.style.setProperty('--header-button-default-border', 'rgba(255, 255, 255, 0.2)');
    root.style.setProperty('--header-button-hover-border', 'rgba(255, 255, 255, 0.3)');
    root.style.setProperty('--header-input-border', 'rgba(255, 255, 255, 0.2)');
    root.style.setProperty('--header-input-hover-border', 'rgba(255, 255, 255, 0.3)');
    root.style.setProperty('--header-input-focus-border', 'rgba(255, 255, 255, 0.4)');
  }
}

/**
 * Parse alpha value from color string
 */
function parseAlpha(color: string): number | null {
  const match = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)/);
  if (match && match[4]) {
    return parseFloat(match[4]);
  }

  // Handle hex colors
  const hexMatch = color.match(/#([a-f0-9]{6})([a-f0-9]{2})?/i);
  if (hexMatch) {
    return 1; // Hex colors are always opaque
  }

  return null;
}

/**
 * Get perceived brightness of a color (0-255)
 */
function getColorBrightness(color: string): number {
  // Remove whitespace
  color = color.replace(/\s/g, '');

  // Handle rgb/rgba
  const rgbMatch = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
  if (rgbMatch) {
    const r = parseInt(rgbMatch[1], 10);
    const g = parseInt(rgbMatch[2], 10);
    const b = parseInt(rgbMatch[3], 10);
    return (r * 299 + g * 587 + b * 114) / 1000;
  }

  // Handle hex
  const hexMatch = color.match(/#?([a-f0-9]{2})([a-f0-9]{2})([a-f0-9]{2})?/i);
  if (hexMatch) {
    const r = parseInt(hexMatch[1], 16);
    const g = parseInt(hexMatch[2], 16);
    const b = parseInt(hexMatch[3] || '', 16);
    return (r * 299 + g * 587 + b * 114) / 1000;
  }

  // Default to light
  return 255;
}

/**
 * Initialize content theme with default settings
 */
export function initContentTheme() {
  applyContentTheme(themeSettings.contentTheme);
}

/**
 * Initialize header theme with default settings
 */
export function initHeaderTheme() {
  applyHeaderTheme(themeSettings.header);
}

/**
 * Reset content theme to defaults
 */
export function resetContentTheme() {
  applyContentTheme(themeSettings.contentTheme);
}
