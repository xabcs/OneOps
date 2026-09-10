/**
 * OneOps 统一主题配置
 *
 * 内容区提供 standard（标准/扁平）和 modern（现代/渐变）两个预设变体，
 * 每个组件的样式可通过 overrides 覆盖预设默认值。
 */

/** Standard（标准）预设：扁平简洁风格 */
export const standardContent: App.Theme.ThemeSetting['content'] = {
  variant: 'standard',
  hero: {
    visible: false,
    background: '#f8fafc',
    gradientStart: '#f8fafc',
    gradientMiddle: '',
    gradientEnd: '#f1f5f9',
    gradientAngle: 135,
    borderColor: 'rgba(148, 163, 184, 0.12)',
    shadow: '0 1px 3px rgba(0, 0, 0, 0.06)',
    padding: '10px 14px',
    iconGradientStart: '#f8fafc',
    iconGradientEnd: '#f1f5f9',
    iconBorderColor: 'rgba(148, 163, 184, 0.16)',
    iconColor: '#475569'
  },
  statCards: {
    useGradient: false,
    defaultBg: '#ffffff',
    defaultBgStart: '#ffffff',
    defaultBgEnd: '#f8fafc',
    successBg: '#f0fdf4',
    successBgStart: '#f0fdf4',
    successBgEnd: '#ffffff',
    warningBg: '#fffbeb',
    warningBgStart: '#fffbeb',
    warningBgEnd: '#ffffff',
    dangerBg: '#fef2f2',
    dangerBgStart: '#fef2f2',
    dangerBgEnd: '#ffffff',
    border: 'rgba(148, 163, 184, 0.16)',
    shadow: '0 1px 3px rgba(0, 0, 0, 0.06)',
    padding: '10px 14px'
  },
  contentCard: {
    borderVisible: false,
    background: '#ffffff',
    gradientStart: '#ffffff',
    gradientMiddle: '',
    gradientEnd: '#f8fafc',
    gradientAngle: 145,
    borderColor: 'rgba(148, 163, 184, 0.12)',
    shadow: '0 1px 3px rgba(0, 0, 0, 0.06)',
    padding: '20px'
  },
  toolbar: {
    background: '#f8fafc',
    gradientStart: '#f8fafc',
    gradientMiddle: '',
    gradientEnd: '#ffffff',
    gradientAngle: 180,
    borderColor: 'rgba(148, 163, 184, 0.12)',
    padding: '8px 12px',
    shadow: 'none'
  },
  dataTable: {
    tableStyle: 'borderless',
    headerBg: '#f8fafc',
    headerTextColor: '#475569',
    headerBorderColor: 'rgba(148, 163, 184, 0.16)',
    rowHoverBg: '#f8fbff',
    rowBorderColor: 'rgba(148, 163, 184, 0.08)',
    borderColor: 'rgba(148, 163, 184, 0.16)',
    stripedBg: '#f8fafc'
  },
  searchFilters: {
    inputBg: 'rgba(255, 255, 255, 0.92)',
    inputBorder: 'rgba(148, 163, 184, 0.16)',
    inputHoverBorder: 'rgba(59, 130, 246, 0.18)',
    inputFocusBorder: 'rgba(37, 99, 235, 0.22)',
    buttonBg: 'rgba(255, 255, 255, 0.9)',
    buttonTextColor: '#475569',
    buttonHoverBg: '#f8fbff'
  },
  pagination: {
    buttonBg: 'rgba(255, 255, 255, 0.9)',
    buttonTextColor: '#475569',
    buttonHoverBg: '#f8fbff',
    activeBg: 'rgb(99, 102, 241)',
    activeTextColor: '#ffffff'
  },
  tags: {
    defaultBg: '#64748b',
    defaultBorder: '#64748b',
    defaultTextColor: '#ffffff',
    successBg: '#10b981',
    warningBg: '#f59e0b',
    dangerBg: '#ef4444',
    infoBg: '#3b82f6'
  },
  button: {
    defaultBg: 'rgba(255, 255, 255, 0.9)',
    defaultColor: '#475569',
    defaultBorder: 'rgba(148, 163, 184, 0.12)',
    hoverBg: '#f8fbff',
    hoverColor: '#1d4ed8',
    hoverBorder: 'rgba(59, 130, 246, 0.18)'
  },
  input: {
    bg: 'rgba(255, 255, 255, 0.92)',
    borderShadow: 'rgba(148, 163, 184, 0.16)',
    hoverShadow: 'rgba(59, 130, 246, 0.18)',
    focusShadow: 'rgba(37, 99, 235, 0.22)'
  },
  colors: {
    primary: 'rgb(99, 102, 241)',
    primaryLight: 'rgb(129, 140, 248)',
    success: 'rgb(16, 185, 129)',
    warning: 'rgb(245, 158, 11)',
    danger: 'rgb(239, 68, 68)',
    info: 'rgb(59, 130, 246)'
  },
  text: {
    primary: 'rgb(30, 41, 59)',
    secondary: 'rgb(100, 116, 139)',
    muted: 'rgb(148, 163, 184)'
  },
  border: {
    soft: 'rgba(148, 163, 184, 0.12)',
    medium: 'rgba(148, 163, 184, 0.18)'
  }
};

/** Modern（现代）预设：渐变丰富风格 */
export const modernContent: App.Theme.ThemeSetting['content'] = {
  ...standardContent,
  variant: 'modern',
  hero: {
    ...standardContent.hero,
    visible: true,
    background: '#e4e9f2',
    gradientStart: '#e0e6f0',
    gradientMiddle: '#e8e5ee',
    gradientEnd: '#f2efe9',
    gradientAngle: 135,
    borderColor: 'rgba(71, 85, 105, 0.10)',
    shadow: '0 8px 24px rgba(15, 23, 42, 0.05)',
    iconGradientStart: '#f5f7fa',
    iconGradientEnd: '#e7ebf2',
    iconBorderColor: 'rgba(71, 85, 105, 0.14)',
    iconColor: '#3d5675'
  },
  statCards: {
    ...standardContent.statCards,
    useGradient: true,
    defaultBg: 'rgba(255, 255, 255, 0.98)',
    defaultBgStart: 'rgba(255, 255, 255, 0.98)',
    defaultBgEnd: 'rgba(241, 245, 250, 0.96)',
    successBg: '#f4f9f6',
    successBgStart: '#f6faf7',
    successBgEnd: '#ecf4ee',
    warningBg: '#fcf8f0',
    warningBgStart: '#fdf9f2',
    warningBgEnd: '#f7efe3',
    dangerBg: '#fcf4f4',
    dangerBgStart: '#fdf6f6',
    dangerBgEnd: '#f8ebeb',
    shadow: '0 6px 18px rgba(15, 23, 42, 0.05)'
  },
  contentCard: {
    ...standardContent.contentCard,
    background: '#f7f8fa',
    gradientStart: '#f6f7fa',
    gradientMiddle: '#eff2f7',
    gradientEnd: '#fbfcfd',
    gradientAngle: 145,
    shadow: '0 10px 28px rgba(15, 23, 42, 0.06)'
  },
  toolbar: {
    ...standardContent.toolbar,
    background: '#eff2f7',
    gradientStart: '#edf0f5',
    gradientMiddle: '#f1f3f8',
    gradientEnd: '#f7f9fb',
    gradientAngle: 180,
    shadow: '0 4px 12px rgba(15, 23, 42, 0.04)'
  },
  dataTable: {
    ...standardContent.dataTable,
    headerBg: '#f1f4f8',
    rowHoverBg: '#f2f6fb'
  }
};

/** 渐变色系预设（用于 modern 变体快速切换整体色调） */
export const gradientPresets = {
  subtle: {
    label: 'subtle',
    hero: { gradientStart: '#e0e6f0', gradientMiddle: '#e8e5ee', gradientEnd: '#f2efe9', gradientAngle: 135 },
    contentCard: { gradientStart: '#f6f7fa', gradientMiddle: '#eff2f7', gradientEnd: '#fbfcfd', gradientAngle: 145 },
    toolbar: { gradientStart: '#edf0f5', gradientMiddle: '#f1f3f8', gradientEnd: '#f7f9fb', gradientAngle: 180 }
  },
  techBlue: {
    label: 'techBlue',
    hero: { gradientStart: '#dbe6f4', gradientMiddle: '#c9d8eb', gradientEnd: '#e9eff7', gradientAngle: 135 },
    contentCard: { gradientStart: '#f2f6fb', gradientMiddle: '#e9eff8', gradientEnd: '#f8fafd', gradientAngle: 145 },
    toolbar: { gradientStart: '#e7edf6', gradientMiddle: '#edf2f9', gradientEnd: '#f5f8fc', gradientAngle: 180 }
  },
  vibrantPurple: {
    label: 'vibrantPurple',
    hero: { gradientStart: '#e8e5f2', gradientMiddle: '#dedcec', gradientEnd: '#f1eff5', gradientAngle: 135 },
    contentCard: { gradientStart: '#f6f5fa', gradientMiddle: '#eff0f8', gradientEnd: '#fbfbfd', gradientAngle: 145 },
    toolbar: { gradientStart: '#ecebf5', gradientMiddle: '#f1f0f8', gradientEnd: '#f8f7fb', gradientAngle: 180 }
  },
  freshGreen: {
    label: 'freshGreen',
    hero: { gradientStart: '#e2efe9', gradientMiddle: '#d8e8e2', gradientEnd: '#eff5f0', gradientAngle: 135 },
    contentCard: { gradientStart: '#f4f8f5', gradientMiddle: '#ebf3ee', gradientEnd: '#f9fcfa', gradientAngle: 145 },
    toolbar: { gradientStart: '#e6efe9', gradientMiddle: '#edf4f0', gradientEnd: '#f5faf7', gradientAngle: 180 }
  },
  warmOrange: {
    label: 'warmOrange',
    hero: { gradientStart: '#f3eadf', gradientMiddle: '#ecdfd0', gradientEnd: '#f7f1e8', gradientAngle: 135 },
    contentCard: { gradientStart: '#faf5ef', gradientMiddle: '#f3ece3', gradientEnd: '#fcf9f5', gradientAngle: 145 },
    toolbar: { gradientStart: '#f1e9df', gradientMiddle: '#f5efe7', gradientEnd: '#faf5ef', gradientAngle: 180 }
  },
  graphite: {
    label: 'graphite',
    hero: { gradientStart: '#dfe3e9', gradientMiddle: '#d2d7df', gradientEnd: '#eceef1', gradientAngle: 155 },
    contentCard: { gradientStart: '#f4f5f7', gradientMiddle: '#eceef2', gradientEnd: '#fafbfc', gradientAngle: 145 },
    toolbar: { gradientStart: '#e7eaef', gradientMiddle: '#edeff3', gradientEnd: '#f5f7f9', gradientAngle: 180 }
  },
  aurora: {
    label: 'aurora',
    hero: { gradientStart: '#d5eee9', gradientMiddle: '#cfe3f0', gradientEnd: '#e2def2', gradientAngle: 165 },
    contentCard: { gradientStart: '#f1f8f7', gradientMiddle: '#eaf2f8', gradientEnd: '#f8fafd', gradientAngle: 150 },
    toolbar: { gradientStart: '#e2f0ee', gradientMiddle: '#eaf3f6', gradientEnd: '#f3f9fa', gradientAngle: 180 }
  },
  midnight: {
    label: 'midnight',
    hero: { gradientStart: '#ccd7e8', gradientMiddle: '#b6c5dd', gradientEnd: '#dbe4f0', gradientAngle: 160 },
    contentCard: { gradientStart: '#eff3f9', gradientMiddle: '#e7edf7', gradientEnd: '#f8fafd', gradientAngle: 145 },
    toolbar: { gradientStart: '#dde5f1', gradientMiddle: '#e5ebf5', gradientEnd: '#f2f6fb', gradientAngle: 180 }
  },
  orchid: {
    label: 'orchid',
    hero: { gradientStart: '#eee2ee', gradientMiddle: '#e4d8ee', gradientEnd: '#f6ecef', gradientAngle: 150 },
    contentCard: { gradientStart: '#f9f4f9', gradientMiddle: '#f2edf8', gradientEnd: '#fdf9fb', gradientAngle: 145 },
    toolbar: { gradientStart: '#efe5f0', gradientMiddle: '#f5edf6', gradientEnd: '#faf4f9', gradientAngle: 180 }
  },
  jade: {
    label: 'jade',
    hero: { gradientStart: '#d8ece1', gradientMiddle: '#c8e2da', gradientEnd: '#e7f1e9', gradientAngle: 145 },
    contentCard: { gradientStart: '#f1f8f4', gradientMiddle: '#e9f3ee', gradientEnd: '#f9fcfa', gradientAngle: 145 },
    toolbar: { gradientStart: '#e2efe8', gradientMiddle: '#eaf3ef', gradientEnd: '#f3faf6', gradientAngle: 180 }
  },
  champagne: {
    label: 'champagne',
    hero: { gradientStart: '#f2e8d5', gradientMiddle: '#eadcc2', gradientEnd: '#f7f0e2', gradientAngle: 135 },
    contentCard: { gradientStart: '#faf4e9', gradientMiddle: '#f4ecdd', gradientEnd: '#fdfaf3', gradientAngle: 145 },
    toolbar: { gradientStart: '#f0e7d5', gradientMiddle: '#f5eedf', gradientEnd: '#faf5eb', gradientAngle: 180 }
  },
  terracotta: {
    label: 'terracotta',
    hero: { gradientStart: '#f4dfd2', gradientMiddle: '#ecd0bd', gradientEnd: '#f8ece0', gradientAngle: 145 },
    contentCard: { gradientStart: '#faf0e9', gradientMiddle: '#f4e6dc', gradientEnd: '#fdf7f2', gradientAngle: 145 },
    toolbar: { gradientStart: '#f1e0d4', gradientMiddle: '#f6e9de', gradientEnd: '#fbf1ea', gradientAngle: 180 }
  },
  plum: {
    label: 'plum',
    hero: { gradientStart: '#e9dce9', gradientMiddle: '#dcccdf', gradientEnd: '#f2e5ec', gradientAngle: 150 },
    contentCard: { gradientStart: '#f8f1f8', gradientMiddle: '#f1e8f3', gradientEnd: '#fcf7fb', gradientAngle: 145 },
    toolbar: { gradientStart: '#eee1ef', gradientMiddle: '#f4eaf5', gradientEnd: '#f9f1fa', gradientAngle: 180 }
  },
  ocean: {
    label: 'ocean',
    hero: { gradientStart: '#d5e7ef', gradientMiddle: '#c3dced', gradientEnd: '#e1eef4', gradientAngle: 175 },
    contentCard: { gradientStart: '#eff6fa', gradientMiddle: '#e6eff7', gradientEnd: '#f8fcfd', gradientAngle: 145 },
    toolbar: { gradientStart: '#dfeef4', gradientMiddle: '#e7f2f7', gradientEnd: '#f2f9fc', gradientAngle: 180 }
  },
  moss: {
    label: 'moss',
    hero: { gradientStart: '#dfead9', gradientMiddle: '#cfdec9', gradientEnd: '#ecf1e5', gradientAngle: 135 },
    contentCard: { gradientStart: '#f3f8ef', gradientMiddle: '#ecf2e7', gradientEnd: '#fafcf7', gradientAngle: 145 },
    toolbar: { gradientStart: '#e6efe0', gradientMiddle: '#edf3e8', gradientEnd: '#f4f9f1', gradientAngle: 180 }
  },
  dusk: {
    label: 'dusk',
    hero: { gradientStart: '#e2e2f0', gradientMiddle: '#d0d3e8', gradientEnd: '#efe7ef', gradientAngle: 160 },
    contentCard: { gradientStart: '#f5f4fa', gradientMiddle: '#eceef7', gradientEnd: '#fbf9fd', gradientAngle: 145 },
    toolbar: { gradientStart: '#e8e9f4', gradientMiddle: '#eef0f8', gradientEnd: '#f5f6fb', gradientAngle: 180 }
  },
  moonAurora: {
    label: 'moonAurora',
    hero: { gradientStart: '#fbe8ff', gradientMiddle: '#eefff8', gradientEnd: '#e9e2ff', gradientAngle: 145 },
    contentCard: { gradientStart: '#fdf3ff', gradientMiddle: '#f5fff9', gradientEnd: '#f2ecff', gradientAngle: 145 },
    toolbar: { gradientStart: '#fefaff', gradientMiddle: '#fafffc', gradientEnd: '#f7f5ff', gradientAngle: 180 }
  },
  peachMist: {
    label: 'peachMist',
    hero: { gradientStart: '#ffe9ec', gradientMiddle: '#fff4e3', gradientEnd: '#e8f6ff', gradientAngle: 140 },
    contentCard: { gradientStart: '#fff2f3', gradientMiddle: '#fef8ec', gradientEnd: '#eff9ff', gradientAngle: 140 },
    toolbar: { gradientStart: '#fff6f7', gradientMiddle: '#fffbf5', gradientEnd: '#f6fbff', gradientAngle: 180 }
  },
  lilacSpring: {
    label: 'lilacSpring',
    hero: { gradientStart: '#f6e9ff', gradientMiddle: '#e9fbf1', gradientEnd: '#ffeadf', gradientAngle: 145 },
    contentCard: { gradientStart: '#f9f0ff', gradientMiddle: '#f0fcf6', gradientEnd: '#fff1ea', gradientAngle: 145 },
    toolbar: { gradientStart: '#fbf6ff', gradientMiddle: '#f6fdf9', gradientEnd: '#fff6f2', gradientAngle: 180 }
  },
  aquaBlossom: {
    label: 'aquaBlossom',
    hero: { gradientStart: '#e6f8ff', gradientMiddle: '#f0e9ff', gradientEnd: '#ffeaf1', gradientAngle: 150 },
    contentCard: { gradientStart: '#effbff', gradientMiddle: '#f6f0ff', gradientEnd: '#fff1f5', gradientAngle: 150 },
    toolbar: { gradientStart: '#f5fcff', gradientMiddle: '#faf5ff', gradientEnd: '#fff6f9', gradientAngle: 180 }
  },
  lemonLilac: {
    label: 'lemonLilac',
    hero: { gradientStart: '#fff6d9', gradientMiddle: '#edf3ff', gradientEnd: '#f8e8ff', gradientAngle: 140 },
    contentCard: { gradientStart: '#fffbe8', gradientMiddle: '#f4f8ff', gradientEnd: '#faefff', gradientAngle: 140 },
    toolbar: { gradientStart: '#fffcf1', gradientMiddle: '#f9fbff', gradientEnd: '#fcf5ff', gradientAngle: 180 }
  },
  mintPetal: {
    label: 'mintPetal',
    hero: { gradientStart: '#e5fff3', gradientMiddle: '#f6e9ff', gradientEnd: '#ffeef0', gradientAngle: 145 },
    contentCard: { gradientStart: '#edfff8', gradientMiddle: '#faf0ff', gradientEnd: '#fff4f5', gradientAngle: 145 },
    toolbar: { gradientStart: '#f5fffa', gradientMiddle: '#fcf6ff', gradientEnd: '#fff8f8', gradientAngle: 180 }
  },
  skyApricot: {
    label: 'skyApricot',
    hero: { gradientStart: '#e9f2ff', gradientMiddle: '#fff3e4', gradientEnd: '#f3e9ff', gradientAngle: 145 },
    contentCard: { gradientStart: '#f0f7ff', gradientMiddle: '#fff8ec', gradientEnd: '#f8f0ff', gradientAngle: 145 },
    toolbar: { gradientStart: '#f6faff', gradientMiddle: '#fffbf4', gradientEnd: '#fbf6ff', gradientAngle: 180 }
  },
  roseGlacier: {
    label: 'roseGlacier',
    hero: { gradientStart: '#ffe9f2', gradientMiddle: '#e9f8ff', gradientEnd: '#f0eaff', gradientAngle: 150 },
    contentCard: { gradientStart: '#fff1f7', gradientMiddle: '#effbff', gradientEnd: '#f5f0ff', gradientAngle: 150 },
    toolbar: { gradientStart: '#fff6fa', gradientMiddle: '#f6fcff', gradientEnd: '#faf5ff', gradientAngle: 180 }
  },
  lilacFern: {
    label: 'lilacFern',
    hero: { gradientStart: '#efe9ff', gradientMiddle: '#eafbef', gradientEnd: '#fff0e8', gradientAngle: 145 },
    contentCard: { gradientStart: '#f5f0ff', gradientMiddle: '#f0fcf4', gradientEnd: '#fff4ec', gradientAngle: 145 },
    toolbar: { gradientStart: '#faf6ff', gradientMiddle: '#f7fdf9', gradientEnd: '#fff8f3', gradientAngle: 180 }
  },
  cloudMelon: {
    label: 'cloudMelon',
    hero: { gradientStart: '#fff0e9', gradientMiddle: '#e8fbf3', gradientEnd: '#eeeaff', gradientAngle: 145 },
    contentCard: { gradientStart: '#fff4ee', gradientMiddle: '#f0fcf8', gradientEnd: '#f4f0ff', gradientAngle: 145 },
    toolbar: { gradientStart: '#fff8f4', gradientMiddle: '#f7fdfa', gradientEnd: '#f9f6ff', gradientAngle: 180 }
  },
  violetDawn: {
    label: 'violetDawn',
    hero: { gradientStart: '#f2e9ff', gradientMiddle: '#fff0f5', gradientEnd: '#e9f7ff', gradientAngle: 150 },
    contentCard: { gradientStart: '#f7f0ff', gradientMiddle: '#fff4f8', gradientEnd: '#f0faff', gradientAngle: 150 },
    toolbar: { gradientStart: '#fbf6ff', gradientMiddle: '#fff8fa', gradientEnd: '#f6fcff', gradientAngle: 180 }
  }
} as const;

export type GradientPresetKey = keyof typeof gradientPresets;

/** Default theme settings */
export const themeSettings: App.Theme.ThemeSetting = {
  themeScheme: 'light',
  grayscale: false,
  colourWeakness: false,
  recommendColor: false,
  themeColor: '#8b5cf6',
  otherColor: {
    info: 'rgb(99, 102, 241)',
    success: 'rgb(16, 185, 129)',
    warning: 'rgb(245, 158, 11)',
    error: 'rgb(239, 68, 68)'
  },
  isInfoFollowPrimary: false,
  layout: {
    mode: 'vertical',
    scrollMode: 'content',
    reverseHorizontalMix: false
  },
  page: {
    animate: false,
    animateMode: 'fade'
  },
  header: {
    height: 60,
    breadcrumb: {
      visible: true,
      showIcon: false
    },
    multilingual: {
      visible: true
    },
    globalSearch: {
      visible: true
    },
    useCustomColor: false,
    customColor: '#ffffff',
    useHeaderGradient: false,
    headerGradientStart: 'rgba(232, 237, 255, 0.98)',
    headerGradientEnd: 'rgba(232, 240, 251, 0.94)'
  },
  tab: {
    visible: false,
    cache: true,
    height: 44,
    mode: 'chrome'
  },
  fixedHeaderAndTab: true,
  sider: {
    inverted: false,
    width: 200,
    collapsedWidth: 56,
    mixWidth: 80,
    mixCollapsedWidth: 56,
    mixChildMenuWidth: 200,
    useCustomColor: false,
    customColor: 'rgb(241, 246, 255)',
    showIcon: true,
    useSiderGradient: false,
    siderGradientStart: 'rgba(241, 246, 255, 0.98)',
    siderGradientEnd: 'rgba(232, 240, 255, 0.9)',
    useLogoGradient: false,
    logoGradientStart: 'rgba(210, 235, 255, 0.98)',
    logoGradientEnd: 'rgba(221, 222, 222, 0.9)'
  },
  footer: {
    visible: false,
    fixed: false,
    height: 48,
    right: true
  },
  content: standardContent,
  watermark: {
    visible: false,
    text: 'SoybeanAdmin',
    enableUserName: false
  },
  borderRadius: {
    useComponentSpecific: true,
    small: '8px',
    medium: '10px',
    large: '12px',
    components: {
      button: '10px',
      input: '12px',
      select: '12px',
      card: '12px',
      modal: '12px',
      tag: '999px',
      switch: '12px',
      checkbox: '8px',
      radio: '50%',
      menu: '10px'
    }
  },
  tokens: {
    light: {
      colors: {
        container: 'rgb(255, 255, 255)',
        layout: 'rgb(241, 245, 249)',
        inverted: 'rgb(30, 41, 59)',
        'base-text': 'rgb(30, 41, 60)',
        'sider-custom': 'rgb(241, 246, 255)'
      },
      boxShadow: {
        header: '0 1px 3px rgb(15 23 42 / 4%)',
        sider: '8px 0 24px rgb(15 23 42 / 3%)',
        tab: '0 1px 2px rgb(0 0 0 / 6%)'
      },
      borderRadius: {
        small: '8px',
        medium: '10px',
        large: '12px'
      }
    },
    dark: {
      colors: {
        container: 'rgb(28, 28, 28)',
        layout: 'rgb(18, 18, 18)',
        'base-text': 'rgb(224, 224, 224)',
        'sider-custom': 'rgb(28, 28, 28)'
      }
    }
  }
};

/**
 * Override theme settings
 *
 * If publish new version, use `overrideThemeSettings` to override certain theme settings
 */
export const overrideThemeSettings: Partial<App.Theme.ThemeSetting> = {
  // 标签统一为实底胶囊样式（覆盖旧缓存中的浅色标签配置）
  content: {
    tags: {
      defaultBg: '#64748b',
      defaultBorder: '#64748b',
      defaultTextColor: '#ffffff',
      successBg: '#10b981',
      warningBg: '#f59e0b',
      dangerBg: '#ef4444',
      infoBg: '#3b82f6'
    }
  } as App.Theme.ThemeSetting['content'],
  // 标签圆角默认改为胶囊（旧缓存为 8px），后续仍由主题设置控制
  borderRadius: {
    components: { tag: '999px' }
  } as App.Theme.ThemeSetting['borderRadius']
};
