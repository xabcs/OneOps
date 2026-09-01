/** Default theme settings - SxDevOps 风格 */
export const themeSettings: App.Theme.ThemeSetting = {
  themeScheme: 'light',
  grayscale: false,
  colourWeakness: false,
  recommendColor: false,
  themeColor: '#8b5cf6', // 主色默认值 rgb(255, 6, 6) - 红色
  otherColor: {
    info: 'rgb(99, 102, 241)', // 与主色保持一致
    success: 'rgb(16, 185, 129)', // 绿色 #10b981
    warning: 'rgb(245, 158, 11)', // 橙色 #f59e0b
    error: 'rgb(239, 68, 68)' // 红色 #ef4444
  },
  isInfoFollowPrimary: false,
  layout: {
    mode: 'vertical',
    scrollMode: 'content',
    reverseHorizontalMix: false
  },
  page: {
    animate: true,
    animateMode: 'fade-bottom'
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
    customColor: '#ffffff',
    useHeaderGradient: false, // 默认不使用渐变
    headerGradientStart: 'rgba(232, 237, 255, 0.98)', // SxDevOps 顶栏渐变起始
    headerGradientEnd: 'rgba(232, 240, 251, 0.94)' // SxDevOps 顶栏渐变结束
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
    useSiderGradient: false, // 默认关闭菜单区域渐变
    siderGradientStart: 'rgba(241, 246, 255, 0.98)', // SxDevOps 浅蓝渐变起始
    siderGradientEnd: 'rgba(232, 240, 255, 0.9)', // SxDevOps 浅蓝渐变结束
    useLogoGradient: false, // 默认关闭Logo区域渐变
    logoGradientStart: 'rgba(210, 235, 255, 0.98)', // SxDevOps Logo渐变起始
    logoGradientEnd: 'rgba(221, 222, 222, 0.9)' // SxDevOps Logo渐变结束
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
    toolbarGradientEnd: 'rgba(255, 255, 255, 0.96)',

    // Hero区域配置 - SxDevOps 风格
    heroGradientStart: 'rgba(251, 253, 255, 0.98)',
    heroGradientEnd: 'rgba(246, 250, 255, 0.96)',

    // 颜色变量 - SxDevOps 风格
    primary: 'rgb(99, 102, 241)',
    primaryLight: 'rgb(129, 140, 248)',
    success: 'rgb(16, 185, 129)',
    warning: 'rgb(245, 158, 11)',
    danger: 'rgb(239, 68, 68)',
    info: 'rgb(59, 130, 246)',

    // 文字颜色 - SxDevOps 风格
    textPrimary: 'rgb(30, 41, 59)',
    textSecondary: 'rgb(100, 116, 139)',
    textMuted: 'rgb(148, 163, 184)',

    // 边框颜色 - SxDevOps 风格
    borderSoft: 'rgba(148, 163, 184, 0.12)',
    borderMedium: 'rgba(148, 163, 184, 0.18)'
  },
  watermark: {
    visible: false,
    text: 'SoybeanAdmin',
    enableUserName: false
  },
  contentTheme2: {
    // Hero 区域配置
    heroSection: {
      visible: false,
      useGradient: true,
      gradientStart: 'rgba(208, 229, 253, 0.98)',
      gradientMiddle: 'rgb(161, 230, 253)',
      gradientEnd: 'rgba(235, 216, 255, 0.96)',
      gradientAngle: 135, // 默认角度
      background: 'rgba(208, 229, 253, 0.98)',
      borderColor: 'rgba(36, 91, 219, 0.09)',
      borderRadius: '20px',
      shadow: '0 8px 24px rgba(15, 23, 42, 0.04)',
      padding: '14px 22px',
      iconGradientStart: 'rgba(243, 247, 255, 0.98)',
      iconGradientEnd: 'rgba(235, 242, 255, 0.96)',
      iconBorderColor: 'rgba(36, 91, 219, 0.12)',
      iconColor: '#245bdb'
    },
    // 统计卡片配置
    statCards: {
      useGradient: true,
      defaultBg: 'rgba(255, 255, 255, 0.98)',
      defaultBgStart: 'rgba(255, 255, 255, 0.98)',
      defaultBgEnd: 'rgba(248, 250, 252, 0.94)',
      defaultBorder: 'rgba(148, 163, 184, 0.16)',
      successBg: 'rgba(240, 253, 244, 0.98)',
      successBgStart: 'rgba(240, 253, 244, 0.98)',
      successBgEnd: 'rgba(255, 255, 255, 0.94)',
      warningBg: 'rgba(255, 251, 235, 0.98)',
      warningBgStart: 'rgba(255, 251, 235, 0.98)',
      warningBgEnd: 'rgba(255, 255, 255, 0.94)',
      dangerBg: 'rgba(254, 242, 242, 0.98)',
      dangerBgStart: 'rgba(254, 242, 242, 0.98)',
      dangerBgEnd: 'rgba(255, 255, 255, 0.94)',
      borderRadius: '12px',
      shadow: '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)'
    },
    // 工具栏配置
    toolbar: {
      sectionBg: 'transparent',
      sectionBorder: 'rgba(148, 163, 184, 0.12)',
      searchBg: 'linear-gradient(180deg, rgba(247, 251, 251, 0.92) 0%, rgba(246, 252, 253, 0.96) 100%)',
      searchBorder: 'rgba(148, 163, 184, 0.12)',
      gradientStart: 'rgba(247, 251, 251, 0.92)',
      gradientMiddle: '', // 可选中间色
      gradientEnd: 'rgba(246, 252, 253, 0.96)',
      gradientAngle: 180, // 默认角度
      borderColor: 'rgba(148, 163, 184, 0.12)',
      borderRadius: '12px',
      padding: '6px 8px',
      shadow: 'inset 0 1px 0 rgba(255, 255, 255, 0.9)'
    },
    // 内容卡片配置
    contentCard: {
      background: '#ffffff',
      bgGradientStart: 'rgba(251, 249, 254, 0.98)',
      bgGradientMiddle: '', // 可选的中间色，用于创建三色渐变
      bgGradientEnd: 'rgba(247, 252, 255, 0.94)',
      useGradient: true,
      gradientAngle: 145, // 默认角度
      borderColor: 'rgba(148, 163, 184, 0.16)',
      borderRadius: '12px',
      shadow: '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04)',
      padding: '20px'
    },
    // 数据表格配置
    dataTable: {
      headerBg: '#fcf2ff',
      headerTextColor: '#000000',
      headerBorderColor: 'rgba(148, 163, 184, 0.16)',
      rowHoverBg: '#f8fbff',
      rowBorderColor: 'rgba(148, 163, 184, 0.16)',
      tableBorder: 'rgba(148, 163, 184, 0.14)',
      borderRadius: '0px',
      stripedBg: '#fbfaff'
    },
    // 搜索筛选配置
    searchFilters: {
      inputBg: 'rgba(255, 255, 255, 0.94)',
      inputBorder: 'rgba(148, 163, 184, 0.12)',
      inputHoverBorder: 'rgba(59, 130, 246, 0.16)',
      inputFocusBorder: 'rgba(37, 99, 235, 0.22)',
      inputBorderRadius: '8px',
      buttonBg: 'rgba(255, 255, 255, 0.9)',
      buttonTextColor: '#475569',
      buttonHoverBg: '#f8fbff'
    },
    // 分页配置
    pagination: {
      buttonBg: 'rgba(255, 255, 255, 0.9)',
      buttonTextColor: '#475569',
      buttonHoverBg: '#f8fbff',
      activeButtonBg: '#3b82f6',
      activeButtonTextColor: '#ffffff',
      borderRadius: '8px'
    },
    // 标签配置
    tags: {
      defaultBg: 'rgba(255, 255, 255, 0.9)',
      defaultBorder: 'rgba(148, 163, 184, 0.12)',
      defaultTextColor: '#475569',
      successBg: 'rgba(16, 185, 129, 0.1)',
      warningBg: 'rgba(245, 158, 11, 0.1)',
      dangerBg: 'rgba(239, 68, 68, 0.1)',
      infoBg: 'rgba(59, 130, 246, 0.1)',
      borderRadius: '8px'
    }
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
        container: 'rgb(28, 28, 28)',
        layout: 'rgb(18, 18, 18)',
        'base-text': 'rgb(224, 224, 224)',
        'sider-custom': 'rgb(28, 28, 28)'
      }
    }
  }
};

/**
 * Preset gradient themes inspired by Aliyun and modern design
 */
export const gradientPresets = {
  /** Classic SxDevOps subtle gradients */
  subtle: {
    heroSection: {
      gradientStart: 'rgba(208, 229, 253, 0.98)',
      gradientMiddle: 'rgb(161, 230, 253)',
      gradientEnd: 'rgba(235, 216, 255, 0.96)',
      gradientAngle: 135
    },
    contentCard: {
      bgGradientStart: 'rgba(251, 249, 254, 0.98)',
      bgGradientEnd: 'rgba(247, 252, 255, 0.94)',
      gradientAngle: 145
    },
    toolbar: {
      gradientStart: 'rgba(247, 251, 251, 0.92)',
      gradientEnd: 'rgba(246, 252, 253, 0.96)',
      gradientAngle: 180
    }
  },
  /** Aliyun Tech Blue - 4-color vibrant gradient */
  techBlue: {
    heroSection: {
      bgGradientStart: '#BAACFF',
      bgGradientMiddle: '#6D54EB',
      bgGradientEnd: '#2B67EE',
      gradientAngle: 75
    },
    contentCard: {
      bgGradientStart: '#5C99FF',
      bgGradientMiddle: '#1366EC',
      bgGradientEnd: '#856EFA',
      gradientAngle: 90
    },
    toolbar: {
      gradientStart: '#EEEBFC',
      gradientEnd: '#DCF1F7',
      gradientAngle: 71
    }
  },
  /** Vibrant Purple - Bold and modern */
  vibrantPurple: {
    heroSection: {
      gradientStart: 'rgba(109, 84, 235, 0.2)',
      gradientMiddle: 'rgba(43, 103, 238, 0.15)',
      gradientEnd: 'rgba(107, 191, 255, 0.1)',
      gradientAngle: 253
    },
    contentCard: {
      bgGradientStart: 'rgba(251, 249, 254, 0.98)',
      bgGradientMiddle: 'rgba(226, 239, 253, 0.96)',
      bgGradientEnd: 'rgba(247, 252, 255, 0.94)',
      gradientAngle: 145
    },
    toolbar: {
      gradientStart: 'rgba(238, 235, 252, 0.95)',
      gradientEnd: 'rgba(220, 241, 247, 0.98)',
      gradientAngle: 71
    }
  },
  /** Fresh Green - Light and airy */
  freshGreen: {
    heroSection: {
      gradientStart: 'rgba(220, 252, 231, 0.98)',
      gradientMiddle: 'rgba(187, 247, 208, 0.95)',
      gradientEnd: 'rgba(134, 239, 172, 0.92)',
      gradientAngle: 135
    },
    contentCard: {
      bgGradientStart: 'rgba(240, 253, 244, 0.98)',
      bgGradientMiddle: 'rgba(220, 252, 231, 0.95)',
      bgGradientEnd: 'rgba(187, 247, 208, 0.92)',
      gradientAngle: 145
    },
    toolbar: {
      gradientStart: 'rgba(240, 253, 244, 0.95)',
      gradientEnd: 'rgba(220, 252, 231, 0.98)',
      gradientAngle: 180
    }
  },
  /** Warm Orange - Energetic and inviting */
  warmOrange: {
    heroSection: {
      gradientStart: 'rgba(255, 237, 213, 0.98)',
      gradientMiddle: 'rgba(254, 215, 170, 0.95)',
      gradientEnd: 'rgba(253, 186, 140, 0.92)',
      gradientAngle: 135
    },
    contentCard: {
      bgGradientStart: 'rgba(255, 247, 237, 0.98)',
      bgGradientMiddle: 'rgba(255, 237, 213, 0.95)',
      bgGradientEnd: 'rgba(254, 215, 170, 0.92)',
      gradientAngle: 145
    },
    toolbar: {
      gradientStart: 'rgba(255, 247, 237, 0.95)',
      gradientEnd: 'rgba(255, 237, 213, 0.98)',
      gradientAngle: 180
    }
  }
};

/**
 * Override theme settings
 *
 * If publish new version, use `overrideThemeSettings` to override certain theme settings
 */
export const overrideThemeSettings: Partial<App.Theme.ThemeSetting> = {};
