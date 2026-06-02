/** Default theme settings - 腾讯云风格 */
export const themeSettings: App.Theme.ThemeSetting = {
  themeScheme: 'light',
  grayscale: false,
  colourWeakness: false,
  recommendColor: false,
  themeColor: 'rgb(0, 82, 217)', // 腾讯云蓝 #0052D9
  otherColor: {
    info: 'rgb(0, 82, 217)', // 与主色保持一致
    success: 'rgb(38, 187, 23)', // 腾讯云绿
    warning: 'rgb(255, 168, 0)', // 腾讯云橙
    error: 'rgb(245, 34, 46)' // 保持红色
  },
  isInfoFollowPrimary: true,
  layout: {
    mode: 'vertical',
    scrollMode: 'content',
    reverseHorizontalMix: false
  },
  page: {
    animate: true,
    animateMode: 'fade-slide'
  },
  header: {
    height: 56,
    breadcrumb: {
      visible: true,
      showIcon: true
    },
    multilingual: {
      visible: true
    },
    globalSearch: {
      visible: true
    }
  },
  tab: {
    visible: true,
    cache: true,
    height: 44,
    mode: 'chrome'
  },
  fixedHeaderAndTab: true,
  sider: {
    inverted: false,
    width: 180,
    collapsedWidth: 56,
    mixWidth: 80,
    mixCollapsedWidth: 56,
    mixChildMenuWidth: 180
  },
  footer: {
    visible: true,
    fixed: false,
    height: 48,
    right: true
  },
  watermark: {
    visible: false,
    text: 'SoybeanAdmin',
    enableUserName: false
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
