/** Default theme settings - OneOps 风格 */
export const themeSettings: App.Theme.ThemeSetting = {
  themeScheme: 'light',
  grayscale: false,
  colourWeakness: false,
  recommendColor: false,
  themeColor: 'rgb(232, 26, 101)', // OneOps #e81a65
  otherColor: {
    info: 'rgb(232, 26, 101)', // 与主色保持一致
    success: 'rgb(38, 187, 23)', // 绿色
    warning: 'rgb(255, 168, 0)', // 橙色
    error: 'rgb(245, 34, 46)' // 保持红色
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
    height: 56,
    breadcrumb: {
      visible: true,
      showIcon: false
    },
    multilingual: {
      visible: true
    },
    globalSearch: {
      visible: true
    }
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
    mixChildMenuWidth: 200
  },
  footer: {
    visible: false,
    fixed: false,
    height: 48,
    right: true
  },
  watermark: {
    visible: false,
    text: 'SoybeanAdmin',
    enableUserName: false
  },
  borderRadius: {
    useComponentSpecific: false,
    small: '0px',
    medium: '0px',
    large: '0px',
    components: {
      button: '0px',
      input: '0px',
      select: '0px',
      card: '0px',
      table: '0px',
      modal: '0px',
      tag: '0px',
      switch: '12px',
      checkbox: '0px',
      radio: '50%',
      menu: '0px'
    }
  },
  tokens: {
    light: {
      colors: {
        container: 'rgb(255, 255, 255)',
        layout: 'rgb(247, 250, 252)',
        inverted: 'rgb(0, 20, 40)',
        'base-text': 'rgb(31, 31, 31)'
      },
      boxShadow: {
        header: '0 1px 3px rgb(0 0 0 / 6%), 0 1px 2px rgb(0 0 0 / 4%)', // 腾讯云风格：更轻的阴影
        sider: '2px 0 6px 0 rgb(0 0 0 / 4%), 1px 0 2px 0 rgb(0 0 0 / 2%)', // 腾讯云风格：更柔和的侧边栏阴影
        tab: '0 1px 2px rgb(0 0 0 / 6%)' // 腾讯云风格：标签页阴影
      },
      borderRadius: {
        small: '0px',
        medium: '0px',
        large: '0px'
      }
    },
    dark: {
      colors: {
        container: 'rgb(28, 28, 28)',
        layout: 'rgb(18, 18, 18)',
        'base-text': 'rgb(224, 224, 224)'
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
