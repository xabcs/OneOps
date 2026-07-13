/** Default theme settings - SxDevOps 风格 */
export const themeSettings: App.Theme.ThemeSetting = {
  themeScheme: 'light',
  grayscale: false,
  colourWeakness: false,
  recommendColor: false,
  themeColor: 'rgb(99, 102, 241)', // SxDevOps 主色 #6366f1 靛蓝色
  otherColor: {
    info: 'rgb(99, 102, 241)', // 与主色保持一致
    success: 'rgb(16, 185, 129)', // 绿色 #10b981
    warning: 'rgb(245, 158, 11)', // 橙色 #f59e0b
    error: 'rgb(239, 68, 68)' // 红色 #ef4444
  },
  isInfoFollowPrimary: true,
  layout: {
    mode: 'vertical',
    scrollMode: 'content',
    reverseHorizontalMix: false
  },
  page: {
    animate: false,
    animateMode: 'fade-slide'
  },
  header: {
    height: 60, // SxDevOps 顶栏高度
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
    customColor: 'rgb(248, 251, 255)',
    useHeaderGradient: true, // SxDevOps 默认使用渐变
    headerGradientStart: 'rgba(248, 251, 255, 0.98)', // SxDevOps 顶栏渐变起始
    headerGradientEnd: 'rgba(243, 247, 255, 0.94)' // SxDevOps 顶栏渐变结束
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
    useSiderGradient: true, // SxDevOps 默认开启菜单区域渐变
    siderGradientStart: 'rgba(241, 246, 255, 0.98)', // SxDevOps 浅蓝渐变起始
    siderGradientEnd: 'rgba(232, 240, 255, 0.9)', // SxDevOps 浅蓝渐变结束
    useLogoGradient: true, // SxDevOps 默认开启Logo区域渐变
    logoGradientStart: 'rgba(241, 246, 255, 0.98)', // SxDevOps Logo渐变起始
    logoGradientEnd: 'rgba(232, 240, 255, 0.9)' // SxDevOps Logo渐变结束
  },
  footer: {
    visible: false,
    fixed: false,
    height: 48,
    right: true
  },
  contentTheme: {
    // 卡片配置 - SxDevOps 风格
    cardBg: '#ffffff',
    cardRadius: '12px',
    cardShadow: '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)',
    useCardGradient: false, // 默认不使用卡片渐变
    cardGradientStart: 'rgba(255, 255, 255, 0.98)',
    cardGradientEnd: 'rgba(248, 250, 252, 0.94)',

    // 表格配置 - SxDevOps 风格
    tableBorder: 'rgba(148, 163, 184, 0.16)',
    tableHeaderBg: '#f8fafc',
    tableHoverBg: '#f8fbff',
    tableRadius: '12px',

    // 按钮配置 - SxDevOps 风格
    buttonRadius: '10px',
    buttonDefaultBg: 'rgba(255, 255, 255, 0.9)',
    buttonDefaultColor: '#475569',
    buttonDefaultBorder: 'rgba(148, 163, 184, 0.12)',
    buttonHoverBg: '#f8fbff',
    buttonHoverColor: '#1d4ed8',
    buttonHoverBorder: 'rgba(59, 130, 246, 0.18)',

    // 输入框配置 - SxDevOps 风格
    inputRadius: '12px',
    inputBg: 'rgba(255, 255, 255, 0.92)',
    inputBorderShadow: 'rgba(148, 163, 184, 0.16)',
    inputHoverShadow: 'rgba(59, 130, 246, 0.18)',
    inputFocusShadow: 'rgba(37, 99, 235, 0.22)',

    // 工具栏渐变 - SxDevOps 风格
    useToolbarGradient: false,
    toolbarGradientStart: 'rgba(248, 250, 252, 0.92)',
    toolbarGradientEnd: 'rgba(255, 255, 255, 0.96)'
  },
  watermark: {
    visible: false,
    text: 'SoybeanAdmin',
    enableUserName: false
  },
  borderRadius: {
    useComponentSpecific: false,
    small: '8px',
    medium: '10px', // 略微增大到SxDevOps风格
    large: '12px', // 卡片使用更大的圆角
    components: {
      button: '8px',
      input: '8px',
      select: '8px',
      card: '12px', // 卡片使用SxDevOps的12px
      table: '8px',
      modal: '12px',
      tag: '8px',
      switch: '12px',
      checkbox: '8px',
      radio: '50%',
      menu: '10px' // 菜单项使用SxDevOps的10px
    }
  },
  tokens: {
    light: {
      colors: {
        container: 'rgb(255, 255, 255)',
        layout: 'rgb(241, 245, 249)', // SxDevOps 内容背景 #f1f5f9
        inverted: 'rgb(30, 41, 59)',
        'base-text': 'rgb(30, 41, 60)', // SxDevOps 深色文字
        'sider-custom': 'rgb(241, 246, 255)' // SxDevOps 侧边栏颜色
      },
      boxShadow: {
        header: '0 1px 3px rgb(15 23 42 / 4%)', // SxDevOps 极简阴影
        sider: '8px 0 24px rgb(15 23 42 / 3%)', // SxDevOps 侧边栏阴影
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
        container: 'rgb(30, 41, 59)',
        layout: 'rgb(15, 23, 42)',
        'base-text': 'rgb(226, 232, 240)',
        'sider-custom': 'rgb(30, 41, 59)'
      }
    }
  }
};

/**
 * Override theme settings
 *
 * If publish new version, use `overrideThemeSettings` to override certain theme settings
 */
export const overrideThemeSettings: Partial<App.Theme.ThemeSetting> = {};
