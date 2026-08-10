/**
 * Content theme utilities
 * Applies content theme settings to CSS variables
 */

import { themeSettings } from '@/theme/settings';

export type ContentThemeSettings = App.Theme.ThemeSetting['contentTheme'];
export type ContentTheme2Settings = App.Theme.ThemeSetting['contentTheme2'];

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

  // Hero gradient
  if (theme.heroGradientStart && theme.heroGradientEnd) {
    root.style.setProperty(
      '--sx-gradient-hero',
      `linear-gradient(135deg, ${theme.heroGradientStart} 0%, ${theme.heroGradientEnd} 100%)`
    );
  }

  // Color variables
  if (theme.primary) {
    root.style.setProperty('--sx-primary', theme.primary);
  }
  if (theme.primaryLight) {
    root.style.setProperty('--sx-primary-light', theme.primaryLight);
  }
  if (theme.success) {
    root.style.setProperty('--sx-success', theme.success);
  }
  if (theme.warning) {
    root.style.setProperty('--sx-warning', theme.warning);
  }
  if (theme.danger) {
    root.style.setProperty('--sx-danger', theme.danger);
  }
  if (theme.info) {
    root.style.setProperty('--sx-info', theme.info);
  }

  // Text colors
  if (theme.textPrimary) {
    root.style.setProperty('--sx-text-primary', theme.textPrimary);
  }
  if (theme.textSecondary) {
    root.style.setProperty('--sx-text-secondary', theme.textSecondary);
  }
  if (theme.textMuted) {
    root.style.setProperty('--sx-text-muted', theme.textMuted);
  }

  // Border colors
  if (theme.borderSoft) {
    root.style.setProperty('--sx-border-soft', theme.borderSoft);
  }
  if (theme.borderMedium) {
    root.style.setProperty('--sx-border-medium', theme.borderMedium);
  }
}

/**
 * Apply content theme 2 to CSS variables
 * Enhanced modular theme system for fine-grained control
 */
export function applyContentTheme2(theme: ContentTheme2Settings) {
  const root = document.documentElement;

  // Hero section styles
  if (theme.heroSection) {
    const hero = theme.heroSection;

    // 设置 Hero 区域的显示状态 CSS 变量
    const isVisible = hero.visible !== false;
    root.style.setProperty('--sx-hero-visible', isVisible ? '1' : '0');

    if (hero.useGradient && hero.gradientStart && hero.gradientEnd) {
      const angle = hero.gradientAngle || 135;
      if (hero.gradientMiddle) {
        // 三色渐变
        root.style.setProperty(
          '--sx-hero-bg',
          `linear-gradient(${angle}deg, ${hero.gradientStart} 0%, ${hero.gradientMiddle} 50%, ${hero.gradientEnd} 100%)`
        );
      } else {
        // 两色渐变
        root.style.setProperty(
          '--sx-hero-bg',
          `linear-gradient(${angle}deg, ${hero.gradientStart} 0%, ${hero.gradientEnd} 100%)`
        );
      }
    } else if (hero.background) {
      root.style.setProperty('--sx-hero-bg', hero.background);
    }
    if (hero.borderColor) {
      root.style.setProperty('--sx-hero-border', hero.borderColor);
    }
    if (hero.borderRadius) {
      root.style.setProperty('--sx-hero-radius', hero.borderRadius);
    }
    if (hero.shadow) {
      root.style.setProperty('--sx-hero-shadow', hero.shadow);
    }
    if (hero.padding) {
      root.style.setProperty('--sx-hero-padding', hero.padding);
    }
    if (hero.iconGradientStart && hero.iconGradientEnd) {
      root.style.setProperty(
        '--sx-hero-icon-bg',
        `linear-gradient(180deg, ${hero.iconGradientStart} 0%, ${hero.iconGradientEnd} 100%)`
      );
    }
    if (hero.iconBorderColor) {
      root.style.setProperty('--sx-hero-icon-border', hero.iconBorderColor);
    }
    if (hero.iconColor) {
      root.style.setProperty('--sx-hero-icon-color', hero.iconColor);
    }
  }

  // Statistics cards styles
  if (theme.statCards) {
    const stat = theme.statCards;

    if (stat.useGradient) {
      // 渐变模式：所有卡片使用渐变
      if (stat.defaultBgStart && stat.defaultBgEnd) {
        root.style.setProperty(
          '--sx-stat-default-bg',
          `linear-gradient(145deg, ${stat.defaultBgStart} 0%, ${stat.defaultBgEnd} 100%)`
        );
      }
      if (stat.successBgStart && stat.successBgEnd) {
        root.style.setProperty(
          '--sx-stat-success-bg',
          `linear-gradient(145deg, ${stat.successBgStart} 0%, ${stat.successBgEnd} 100%)`
        );
      }
      if (stat.warningBgStart && stat.warningBgEnd) {
        root.style.setProperty(
          '--sx-stat-warning-bg',
          `linear-gradient(145deg, ${stat.warningBgStart} 0%, ${stat.warningBgEnd} 100%)`
        );
      }
      if (stat.dangerBgStart && stat.dangerBgEnd) {
        root.style.setProperty(
          '--sx-stat-danger-bg',
          `linear-gradient(145deg, ${stat.dangerBgStart} 0%, ${stat.dangerBgEnd} 100%)`
        );
      }
    } else {
      // 纯色模式：所有卡片使用纯色
      if (stat.defaultBg) {
        root.style.setProperty('--sx-stat-default-bg', stat.defaultBg);
      }
      if (stat.successBg) {
        root.style.setProperty('--sx-stat-success-bg', stat.successBg);
      }
      if (stat.warningBg) {
        root.style.setProperty('--sx-stat-warning-bg', stat.warningBg);
      }
      if (stat.dangerBg) {
        root.style.setProperty('--sx-stat-danger-bg', stat.dangerBg);
      }
    }

    // 通用样式（不受渐变开关影响）
    if (stat.defaultBorder) {
      root.style.setProperty('--sx-stat-default-border', stat.defaultBorder);
    }
    if (stat.borderRadius) {
      root.style.setProperty('--sx-stat-radius', stat.borderRadius);
    }
    if (stat.shadow) {
      root.style.setProperty('--sx-stat-shadow', stat.shadow);
    }
  }

  // Toolbar styles
  if (theme.toolbar) {
    const toolbar = theme.toolbar;
    if (toolbar.gradientStart && toolbar.gradientEnd) {
      const angle = toolbar.gradientAngle || 180;
      if (toolbar.gradientMiddle) {
        // 三色渐变
        root.style.setProperty(
          '--sx-toolbar-bg',
          `linear-gradient(${angle}deg, ${toolbar.gradientStart} 0%, ${toolbar.gradientMiddle} 50%, ${toolbar.gradientEnd} 100%)`
        );
      } else {
        // 两色渐变
        root.style.setProperty(
          '--sx-toolbar-bg',
          `linear-gradient(${angle}deg, ${toolbar.gradientStart} 0%, ${toolbar.gradientEnd} 100%)`
        );
      }
    }
    if (toolbar.borderColor) {
      root.style.setProperty('--sx-toolbar-border', toolbar.borderColor);
    }
    if (toolbar.borderRadius) {
      root.style.setProperty('--sx-toolbar-radius', toolbar.borderRadius);
    }
    if (toolbar.padding) {
      root.style.setProperty('--sx-toolbar-padding', toolbar.padding);
    }
    if (toolbar.shadow) {
      root.style.setProperty('--sx-toolbar-shadow', toolbar.shadow);
    }
  }

  // Content card styles
  if (theme.contentCard) {
    const card = theme.contentCard;
    if (card.useGradient && card.bgGradientStart && card.bgGradientEnd) {
      const angle = card.gradientAngle || 145;
      if (card.bgGradientMiddle) {
        // 三色渐变
        root.style.setProperty(
          '--sx-content-card-bg',
          `linear-gradient(${angle}deg, ${card.bgGradientStart} 0%, ${card.bgGradientMiddle} 50%, ${card.bgGradientEnd} 100%)`
        );
      } else {
        // 两色渐变
        root.style.setProperty(
          '--sx-content-card-bg',
          `linear-gradient(${angle}deg, ${card.bgGradientStart} 0%, ${card.bgGradientEnd} 100%)`
        );
      }
    } else if (card.background) {
      root.style.setProperty('--sx-content-card-bg', card.background);
    }
    if (card.borderColor) {
      root.style.setProperty('--sx-content-card-border', card.borderColor);
    }
    if (card.borderRadius) {
      root.style.setProperty('--sx-content-card-radius', card.borderRadius);
    }
    if (card.shadow) {
      root.style.setProperty('--sx-content-card-shadow', card.shadow);
    }
    if (card.padding) {
      root.style.setProperty('--sx-content-card-padding', card.padding);
    }
  }

  // Data table styles
  if (theme.dataTable) {
    const table = theme.dataTable;
    if (table.headerBg) {
      root.style.setProperty('--sx-table-header-bg', table.headerBg);
    }
    if (table.headerTextColor) {
      root.style.setProperty('--sx-table-header-text', table.headerTextColor);
    }
    if (table.headerBorderColor) {
      root.style.setProperty('--sx-table-header-border', table.headerBorderColor);
    }
    if (table.rowHoverBg) {
      root.style.setProperty('--sx-table-row-hover', table.rowHoverBg);
    }
    if (table.rowBorderColor) {
      root.style.setProperty('--sx-table-row-border', table.rowBorderColor);
    }
    if (table.tableBorder) {
      root.style.setProperty('--sx-table-border', table.tableBorder);
    }
    if (table.borderRadius) {
      root.style.setProperty('--sx-table-radius', table.borderRadius);
    }
    if (table.stripedBg) {
      root.style.setProperty('--sx-table-striped-bg', table.stripedBg);
    }
  }

  // Search filters styles
  if (theme.searchFilters) {
    const search = theme.searchFilters;
    if (search.inputBg) {
      root.style.setProperty('--sx-search-input-bg', search.inputBg);
    }
    if (search.inputBorder) {
      root.style.setProperty('--sx-search-input-border', search.inputBorder);
    }
    if (search.inputHoverBorder) {
      root.style.setProperty('--sx-search-input-hover-border', search.inputHoverBorder);
    }
    if (search.inputFocusBorder) {
      root.style.setProperty('--sx-search-input-focus-border', search.inputFocusBorder);
    }
    if (search.inputBorderRadius) {
      root.style.setProperty('--sx-search-input-radius', search.inputBorderRadius);
    }
    if (search.buttonBg) {
      root.style.setProperty('--sx-search-button-bg', search.buttonBg);
    }
    if (search.buttonTextColor) {
      root.style.setProperty('--sx-search-button-text', search.buttonTextColor);
    }
    if (search.buttonHoverBg) {
      root.style.setProperty('--sx-search-button-hover', search.buttonHoverBg);
    }
  }

  // Pagination styles
  if (theme.pagination) {
    const pagination = theme.pagination;
    if (pagination.buttonBg) {
      root.style.setProperty('--sx-pagination-button-bg', pagination.buttonBg);
    }
    if (pagination.buttonTextColor) {
      root.style.setProperty('--sx-pagination-button-text', pagination.buttonTextColor);
    }
    if (pagination.buttonHoverBg) {
      root.style.setProperty('--sx-pagination-button-hover', pagination.buttonHoverBg);
    }
    if (pagination.activeButtonBg) {
      root.style.setProperty('--sx-pagination-active-bg', pagination.activeButtonBg);
    }
    if (pagination.activeButtonTextColor) {
      root.style.setProperty('--sx-pagination-active-text', pagination.activeButtonTextColor);
    }
    if (pagination.borderRadius) {
      root.style.setProperty('--sx-pagination-radius', pagination.borderRadius);
    }
  }

  // Tags styles
  if (theme.tags) {
    const tags = theme.tags;
    if (tags.defaultBg) {
      root.style.setProperty('--sx-tag-default-bg', tags.defaultBg);
    }
    if (tags.defaultBorder) {
      root.style.setProperty('--sx-tag-default-border', tags.defaultBorder);
    }
    if (tags.defaultTextColor) {
      root.style.setProperty('--sx-tag-default-text', tags.defaultTextColor);
    }
    if (tags.successBg) {
      root.style.setProperty('--sx-tag-success-bg', tags.successBg);
    }
    if (tags.warningBg) {
      root.style.setProperty('--sx-tag-warning-bg', tags.warningBg);
    }
    if (tags.dangerBg) {
      root.style.setProperty('--sx-tag-danger-bg', tags.dangerBg);
    }
    if (tags.infoBg) {
      root.style.setProperty('--sx-tag-info-bg', tags.infoBg);
    }
    if (tags.borderRadius) {
      root.style.setProperty('--sx-tag-radius', tags.borderRadius);
    }
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
    root.style.setProperty('--header-input-bg', `rgba(255, 255, 255, ${Math.max(startAlpha - 0.23, 0.62).toFixed(2)})`);
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
    return Number.parseFloat(match[4]);
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
    const r = Number.parseInt(rgbMatch[1], 10);
    const g = Number.parseInt(rgbMatch[2], 10);
    const b = Number.parseInt(rgbMatch[3], 10);
    return (r * 299 + g * 587 + b * 114) / 1000;
  }

  // Handle hex
  const hexMatch = color.match(/#?([a-f0-9]{2})([a-f0-9]{2})([a-f0-9]{2})?/i);
  if (hexMatch) {
    const r = Number.parseInt(hexMatch[1], 16);
    const g = Number.parseInt(hexMatch[2], 16);
    const b = Number.parseInt(hexMatch[3] || '', 16);
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
 * Initialize hero section visibility
 * @deprecated 使用组件内的 v-if="heroVisible" 条件渲染
 */
export function initHeroVisibility(theme?: ContentTheme2Settings) {
  // 不再需要 DOM 操作，由组件自行处理显示逻辑
}

/**
 * Initialize content theme 2 with default settings
 */
export function initContentTheme2() {
  applyContentTheme2(themeSettings.contentTheme2);
  initHeroVisibility(themeSettings.contentTheme2);
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

/**
 * Reset content theme 2 to defaults
 */
export function resetContentTheme2() {
  applyContentTheme2(themeSettings.contentTheme2);
}
