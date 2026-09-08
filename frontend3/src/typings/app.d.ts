/** The global namespace for the app */
declare namespace App {
  /** Theme namespace */
  namespace Theme {
    type ColorPaletteNumber = import('@sa/color').ColorPaletteNumber;

    /** Theme setting */
    interface ThemeSetting {
      /** Theme scheme */
      themeScheme: UnionKey.ThemeScheme;
      /** grayscale mode */
      grayscale: boolean;
      /** colour weakness mode */
      colourWeakness: boolean;
      /** Whether to recommend color */
      recommendColor: boolean;
      /** Theme color */
      themeColor: string;
      /** Other color */
      otherColor: OtherColor;
      /** Whether info color is followed by the primary color */
      isInfoFollowPrimary: boolean;
      /** Layout */
      layout: {
        /** Layout mode */
        mode: UnionKey.ThemeLayoutMode;
        /** Scroll mode */
        scrollMode: UnionKey.ThemeScrollMode;
        /**
         * Whether to reverse the horizontal mix
         *
         * if true, the vertical child level menus in left and horizontal first level menus in top
         */
        reverseHorizontalMix: boolean;
      };
      /** Page */
      page: {
        /** Whether to show the page transition */
        animate: boolean;
        /** Page animate mode */
        animateMode: UnionKey.ThemePageAnimateMode;
      };
      /** Header */
      header: {
        /** Header height */
        height: number;
        /** Header breadcrumb */
        breadcrumb: {
          /** Whether to show the breadcrumb */
          visible: boolean;
          /** Whether to show the breadcrumb icon */
          showIcon: boolean;
        };
        /** Multilingual */
        multilingual: {
          /** Whether to show the multilingual */
          visible: boolean;
        };
        /** Global search */
        globalSearch: {
          /** Whether to show the global search */
          visible: boolean;
        };
        /** Use custom header color */
        useCustomColor?: boolean;
        /** Custom header background color (single color) */
        customColor?: string;
        /** Use header gradient background */
        useHeaderGradient?: boolean;
        /** Header gradient start color */
        headerGradientStart?: string;
        /** Header gradient end color */
        headerGradientEnd?: string;
      };
      /** Tab */
      tab: {
        /** Whether to show the tab */
        visible: boolean;
        /**
         * Whether to cache the tab
         *
         * If cache, the tabs will get from the local storage when the page is refreshed
         */
        cache: boolean;
        /** Tab height */
        height: number;
        /** Tab mode */
        mode: UnionKey.ThemeTabMode;
      };
      /** Fixed header and tab */
      fixedHeaderAndTab: boolean;
      /** Sider */
      sider: {
        /** Inverted sider */
        inverted: boolean;
        /** Sider width */
        width: number;
        /** Collapsed sider width */
        collapsedWidth: number;
        /** Sider width when the layout is 'vertical-mix' or 'horizontal-mix' */
        mixWidth: number;
        /** Collapsed sider width when the layout is 'vertical-mix' or 'horizontal-mix' */
        mixCollapsedWidth: number;
        /** Child menu width when the layout is 'vertical-mix' or 'horizontal-mix' */
        mixChildMenuWidth: number;
        /** Custom background color for sider (when inverted is true) */
        customColor?: string;
        /** Use custom color for sider background */
        useCustomColor?: boolean;
        /** Show sider menu icon */
        showIcon?: boolean;
        /** Enable sider gradient background */
        useSiderGradient?: boolean;
        /** Sider gradient start color */
        siderGradientStart?: string;
        /** Sider gradient end color */
        siderGradientEnd?: string;
        /** Enable logo area gradient background */
        useLogoGradient?: boolean;
        /** Logo gradient start color */
        logoGradientStart?: string;
        /** Logo gradient end color */
        logoGradientEnd?: string;
      };
      /** Footer */
      footer: {
        /** Whether to show the footer */
        visible: boolean;
        /** Whether fixed the footer */
        fixed: boolean;
        /** Footer height */
        height: number;
        /** Whether float the footer to the right when the layout is 'horizontal-mix' */
        right: boolean;
      };
      /** Watermark */
      watermark: {
        /** Whether to show the watermark */
        visible: boolean;
        /** Watermark text */
        text: string;
        /** Whether to use user name as watermark text */
        enableUserName: boolean;
      };
      /** Border radius settings */
      borderRadius: {
        /** Whether to use component-specific border radius */
        useComponentSpecific: boolean;
        /** Small components border radius: buttons, inputs, selects, etc. */
        small: string;
        /** Medium components border radius: menu items, tabs, popups, etc. */
        medium: string;
        /** Large containers border radius: cards, panels, etc. */
        large: string;
        /** Component-specific border radius values */
        components: {
          /** Button border radius */
          button: string;
          /** Input border radius */
          input: string;
          /** Select border radius */
          select: string;
          /** Card border radius */
          card: string;
          /** Modal/Dialog border radius */
          modal: string;
          /** Tag border radius */
          tag: string;
          /** Switch border radius */
          switch: string;
          /** Checkbox border radius */
          checkbox: string;
          /** Radio border radius */
          radio: string;
          /** Menu border radius */
          menu: string;
        };
      };
      /** Content area theme settings */
      /** Content area unified theme configuration */
      content: {
        /** Content variant: standard (flat) or modern (gradient) */
        variant: 'standard' | 'modern';

        /** Hero section (top banner area) */
        hero: {
          /** Whether hero section is visible */
          visible: boolean;
          /** Background color when gradient is off */
          background: string;
          /** Gradient start color */
          gradientStart: string;
          /** Gradient middle color (optional, for 3-color gradients) */
          gradientMiddle: string;
          /** Gradient end color */
          gradientEnd: string;
          /** Gradient angle in degrees */
          gradientAngle: number;
          /** Border color */
          borderColor: string;
          /** Box shadow */
          shadow: string;
          /** Padding */
          padding: string;
          /** Icon background gradient start */
          iconGradientStart: string;
          /** Icon background gradient end */
          iconGradientEnd: string;
          /** Icon border color */
          iconBorderColor: string;
          /** Icon text color */
          iconColor: string;
        };

        /** Statistics cards */
        statCards: {
          /** Use gradient for all stat cards */
          useGradient: boolean;
          /** Default card solid background (when gradient off) */
          defaultBg: string;
          /** Default card gradient start */
          defaultBgStart: string;
          /** Default card gradient end */
          defaultBgEnd: string;
          /** Success card solid background */
          successBg: string;
          /** Success card gradient start */
          successBgStart: string;
          /** Success card gradient end */
          successBgEnd: string;
          /** Warning card solid background */
          warningBg: string;
          /** Warning card gradient start */
          warningBgStart: string;
          /** Warning card gradient end */
          warningBgEnd: string;
          /** Danger card solid background */
          dangerBg: string;
          /** Danger card gradient start */
          dangerBgStart: string;
          /** Danger card gradient end */
          dangerBgEnd: string;
          /** Card border color */
          border: string;
          /** Card shadow */
          shadow: string;
          /** Card padding */
          padding: string;
        };

        /** Content card (page-level card container) */
        contentCard: {
          /** Background color when gradient is off */
          background: string;
          /** Gradient start color */
          gradientStart: string;
          /** Gradient middle color (optional) */
          gradientMiddle: string;
          /** Gradient end color */
          gradientEnd: string;
          /** Gradient angle */
          gradientAngle: number;
          /** Border color */
          borderColor: string;
          /** Box shadow */
          shadow: string;
          /** Padding */
          padding: string;
        };

        /** Toolbar (search/filter area) */
        toolbar: {
          /** Background color when gradient is off */
          background: string;
          /** Gradient start color */
          gradientStart: string;
          /** Gradient middle color (optional) */
          gradientMiddle: string;
          /** Gradient end color */
          gradientEnd: string;
          /** Gradient angle */
          gradientAngle: number;
          /** Border color */
          borderColor: string;
          /** Padding */
          padding: string;
          /** Box shadow */
          shadow: string;
        };

        /** Data table */
        dataTable: {
          /** Table presentation style */
          tableStyle: 'borderless' | 'zebra' | 'soft' | 'grid';
          /** Header background color */
          headerBg: string;
          /** Header text color */
          headerTextColor: string;
          /** Header border color */
          headerBorderColor: string;
          /** Row hover background */
          rowHoverBg: string;
          /** Row border color */
          rowBorderColor: string;
          /** Table border color */
          borderColor: string;
          /** Striped row background */
          stripedBg: string;
        };

        /** Search filters (input area within toolbar) */
        searchFilters: {
          /** Input background color */
          inputBg: string;
          /** Input border color */
          inputBorder: string;
          /** Input hover border color */
          inputHoverBorder: string;
          /** Input focus border color */
          inputFocusBorder: string;
          /** Button background color */
          buttonBg: string;
          /** Button text color */
          buttonTextColor: string;
          /** Button hover background */
          buttonHoverBg: string;
        };

        /** Pagination */
        pagination: {
          /** Button background color */
          buttonBg: string;
          /** Button text color */
          buttonTextColor: string;
          /** Button hover background */
          buttonHoverBg: string;
          /** Active page background */
          activeBg: string;
          /** Active page text color */
          activeTextColor: string;
        };

        /** Tags */
        tags: {
          /** Default tag background */
          defaultBg: string;
          /** Default tag border */
          defaultBorder: string;
          /** Default tag text color */
          defaultTextColor: string;
          /** Success tag background */
          successBg: string;
          /** Warning tag background */
          warningBg: string;
          /** Danger tag background */
          dangerBg: string;
          /** Info tag background */
          infoBg: string;
        };

        /** Shared button style for content area */
        button: {
          /** Default background color */
          defaultBg: string;
          /** Default text color */
          defaultColor: string;
          /** Default border color */
          defaultBorder: string;
          /** Hover background color */
          hoverBg: string;
          /** Hover text color */
          hoverColor: string;
          /** Hover border color */
          hoverBorder: string;
        };

        /** Shared input style for content area */
        input: {
          /** Input background color */
          bg: string;
          /** Inner border shadow color */
          borderShadow: string;
          /** Hover inner shadow color */
          hoverShadow: string;
          /** Focus inner shadow color */
          focusShadow: string;
        };

        /** Shared card style for content area */
        card: {
          /** Card background color */
          bg: string;
          /** Card border color */
          border: string;
          /** Card shadow */
          shadow: string;
          /** Card padding */
          padding: string;
        };

        /** Content area color palette */
        colors: {
          primary: string;
          primaryLight: string;
          success: string;
          warning: string;
          danger: string;
          info: string;
        };

        /** Content area text colors */
        text: {
          primary: string;
          secondary: string;
          muted: string;
        };

        /** Content area border colors */
        border: {
          soft: string;
          medium: string;
        };
      };
      tokens: {
        light: ThemeSettingToken;
        dark?: {
          [K in keyof ThemeSettingToken]?: Partial<ThemeSettingToken[K]>;
        };
      };
    }

    interface OtherColor {
      info: string;
      success: string;
      warning: string;
      error: string;
    }

    interface ThemeColor extends OtherColor {
      primary: string;
    }

    type ThemeColorKey = keyof ThemeColor;

    type ThemePaletteColor = {
      [key in ThemeColorKey | `${ThemeColorKey}-${ColorPaletteNumber}`]: string;
    };

    type BaseToken = Record<string, Record<string, string>>;

    interface ThemeSettingTokenColor {
      /** the progress bar color, if not set, will use the primary color */
      nprogress?: string;
      container: string;
      layout: string;
      inverted: string;
      'base-text': string;
      /** Custom sider background color */
      'sider-custom'?: string;
    }

    interface ThemeSettingTokenBoxShadow {
      header: string;
      sider: string;
      tab: string;
    }

    interface ThemeSettingTokenBorderRadius {
      small: string;
      medium: string;
      large: string;
    }

    interface ThemeSettingToken {
      colors: ThemeSettingTokenColor;
      boxShadow: ThemeSettingTokenBoxShadow;
      borderRadius: ThemeSettingTokenBorderRadius;
    }

    type ThemeTokenColor = ThemePaletteColor & ThemeSettingTokenColor;

    /** Theme token CSS variables */
    type ThemeTokenCSSVars = {
      colors: ThemeTokenColor & { [key: string]: string };
      boxShadow: ThemeSettingTokenBoxShadow & { [key: string]: string };
      borderRadius: ThemeSettingTokenBorderRadius & { [key: string]: string };
    };
  }

  /** Global namespace */
  namespace Global {
    type VNode = import('vue').VNode;
    type RouteLocationNormalizedLoaded = import('vue-router').RouteLocationNormalizedLoaded;
    type RouteKey = import('@elegant-router/types').RouteKey;
    type RouteMap = import('@elegant-router/types').RouteMap;
    type RoutePath = import('@elegant-router/types').RoutePath;
    type LastLevelRouteKey = import('@elegant-router/types').LastLevelRouteKey;

    /** The global header props */
    interface HeaderProps {
      /** Whether to show the logo */
      showLogo?: boolean;
      /** Whether to show the menu toggler */
      showMenuToggler?: boolean;
      /** Whether to show the menu */
      showMenu?: boolean;
    }

    /** The global menu */
    type Menu = {
      /**
       * The menu key
       *
       * Equal to the route key
       */
      key: string;
      /** The menu label */
      label: string;
      /** The menu i18n key */
      i18nKey?: I18n.I18nKey | null;
      /** The route key */
      routeKey: RouteKey;
      /** The route path */
      routePath: RoutePath;
      /** The menu icon */
      icon?: () => VNode;
      /** The menu children */
      children?: Menu[];
    };

    type Breadcrumb = Omit<Menu, 'children'> & {
      options?: Breadcrumb[];
    };

    /** Tab route */
    type TabRoute = Pick<RouteLocationNormalizedLoaded, 'name' | 'path' | 'meta'> &
      Partial<Pick<RouteLocationNormalizedLoaded, 'fullPath' | 'query' | 'matched'>>;

    /** The global tab */
    type Tab = {
      /** The tab id */
      id: string;
      /** The tab label */
      label: string;
      /**
       * The new tab label
       *
       * If set, the tab label will be replaced by this value
       */
      newLabel?: string;
      /**
       * The old tab label
       *
       * when reset the tab label, the tab label will be replaced by this value
       */
      oldLabel?: string;
      /** The tab route key */
      routeKey: LastLevelRouteKey;
      /** The tab route path */
      routePath: RouteMap[LastLevelRouteKey];
      /** The tab route full path */
      fullPath: string;
      /** The tab fixed index */
      fixedIndex?: number | null;
      /**
       * Tab icon
       *
       * Iconify icon
       */
      icon?: string;
      /**
       * Tab local icon
       *
       * Local icon
       */
      localIcon?: string;
      /** I18n key */
      i18nKey?: I18n.I18nKey | null;
    };

    /** Form rule */
    type FormRule = import('element-plus').FormItemRule;

    /** The global dropdown key */
    type DropdownKey = 'closeCurrent' | 'closeOther' | 'closeLeft' | 'closeRight' | 'closeAll';
  }

  /**
   * I18n namespace
   *
   * Locales type
   */
  namespace I18n {
    type RouteKey = import('@elegant-router/types').RouteKey;

    type LangType = 'en-US' | 'zh-CN';

    type LangOption = {
      label: string;
      key: LangType;
    };

    type I18nRouteKey = Exclude<RouteKey, 'root' | 'not-found'>;

    type FormMsg = {
      required: string;
      invalid: string;
    };

    type Schema = {
      system: {
        title: string;
        updateTitle: string;
        updateContent: string;
        updateConfirm: string;
        updateCancel: string;
      };
      common: {
        action: string;
        add: string;
        addSuccess: string;
        backToHome: string;
        batchDelete: string;
        cancel: string;
        close: string;
        check: string;
        expandColumn: string;
        columnSetting: string;
        config: string;
        confirm: string;
        delete: string;
        deleteSuccess: string;
        confirmDelete: string;
        edit: string;
        warning: string;
        error: string;
        index: string;
        keywordSearch: string;
        logout: string;
        logoutConfirm: string;
        lookForward: string;
        modify: string;
        modifySuccess: string;
        noData: string;
        operate: string;
        pleaseCheckValue: string;
        refresh: string;
        reset: string;
        search: string;
        switch: string;
        tip: string;
        trigger: string;
        update: string;
        updateSuccess: string;
        userCenter: string;
        yesOrNo: {
          yes: string;
          no: string;
        };
      };
      request: {
        logout: string;
        logoutMsg: string;
        logoutWithModal: string;
        logoutWithModalMsg: string;
        refreshToken: string;
        tokenExpired: string;
      };
      theme: {
        themeSchema: { title: string } & Record<UnionKey.ThemeScheme, string>;
        grayscale: string;
        colourWeakness: string;
        layoutMode: { title: string; reverseHorizontalMix: string } & Record<UnionKey.ThemeLayoutMode, string>;
        recommendColor: string;
        recommendColorDesc: string;
        themeColor: {
          title: string;
          followPrimary: string;
        } & Record<Theme.ThemeColorKey, string>;
        scrollMode: { title: string } & Record<UnionKey.ThemeScrollMode, string>;
        page: {
          animate: string;
          mode: { title: string } & Record<UnionKey.ThemePageAnimateMode, string>;
        };
        fixedHeaderAndTab: string;
        header: {
          height: string;
          breadcrumb: {
            visible: string;
            showIcon: string;
          };
          multilingual: {
            visible: string;
          };
          globalSearch: {
            visible: string;
          };
        };
        tab: {
          visible: string;
          cache: string;
          height: string;
          mode: { title: string } & Record<UnionKey.ThemeTabMode, string>;
        };
        sider: {
          inverted: string;
          width: string;
          collapsedWidth: string;
          mixWidth: string;
          mixCollapsedWidth: string;
          mixChildMenuWidth: string;
        };
        footer: {
          visible: string;
          fixed: string;
          height: string;
          right: string;
        };
        watermark: {
          visible: string;
          text: string;
          enableUserName: string;
        };
        themeDrawerTitle: string;
        pageFunTitle: string;
        configOperation: {
          copyConfig: string;
          copySuccessMsg: string;
          resetConfig: string;
          resetSuccessMsg: string;
        };
      };
      route: Record<I18nRouteKey, string>;
      page: {
        login: {
          common: {
            loginOrRegister: string;
            userNamePlaceholder: string;
            phonePlaceholder: string;
            codePlaceholder: string;
            passwordPlaceholder: string;
            confirmPasswordPlaceholder: string;
            codeLogin: string;
            confirm: string;
            back: string;
            validateSuccess: string;
            loginSuccess: string;
            welcomeBack: string;
          };
          pwdLogin: {
            title: string;
            rememberMe: string;
            forgetPassword: string;
            register: string;
            otherAccountLogin: string;
            otherLoginMode: string;
            superAdmin: string;
            admin: string;
            user: string;
          };
          codeLogin: {
            title: string;
            getCode: string;
            reGetCode: string;
            sendCodeSuccess: string;
            imageCodePlaceholder: string;
          };
          register: {
            title: string;
            agreement: string;
            protocol: string;
            policy: string;
          };
          resetPwd: {
            title: string;
          };
          bindWeChat: {
            title: string;
          };
        };
        about: {
          title: string;
          introduction: string;
          projectInfo: {
            title: string;
            version: string;
            latestBuildTime: string;
            githubLink: string;
            previewLink: string;
          };
          prdDep: string;
          devDep: string;
        };
        home: {
          branchDesc: string;
          greeting: string;
          weatherDesc: string;
          projectCount: string;
          todo: string;
          message: string;
          downloadCount: string;
          registerCount: string;
          schedule: string;
          study: string;
          work: string;
          rest: string;
          entertainment: string;
          visitCount: string;
          turnover: string;
          dealCount: string;
          projectNews: {
            title: string;
            moreNews: string;
            desc1: string;
            desc2: string;
            desc3: string;
            desc4: string;
            desc5: string;
          };
          creativity: string;
        };
        function: {
          tab: {
            tabOperate: {
              title: string;
              addTab: string;
              addTabDesc: string;
              closeTab: string;
              closeCurrentTab: string;
              closeAboutTab: string;
              addMultiTab: string;
              addMultiTabDesc1: string;
              addMultiTabDesc2: string;
            };
            tabTitle: {
              title: string;
              changeTitle: string;
              change: string;
              resetTitle: string;
              reset: string;
            };
          };
          multiTab: {
            routeParam: string;
            backTab: string;
          };
          toggleAuth: {
            toggleAccount: string;
            authHook: string;
            superAdminVisible: string;
            adminVisible: string;
            adminOrUserVisible: string;
          };
          request: {
            repeatedErrorOccurOnce: string;
            repeatedError: string;
            repeatedErrorMsg1: string;
            repeatedErrorMsg2: string;
          };
        };
        alova: {
          scenes: {
            captchaSend: string;
            autoRequest: string;
            visibilityRequestTips: string;
            pollingRequestTips: string;
            networkRequestTips: string;
            refreshTime: string;
            startRequest: string;
            stopRequest: string;
            requestCrossComponent: string;
            triggerAllRequest: string;
          };
        };
        manage: {
          common: {
            status: {
              enable: string;
              disable: string;
            };
          };
          role: {
            title: string;
            roleName: string;
            roleCode: string;
            roleStatus: string;
            roleDesc: string;
            form: {
              roleName: string;
              roleCode: string;
              roleStatus: string;
              roleDesc: string;
            };
            addRole: string;
            editRole: string;
            menuAuth: string;
            buttonAuth: string;
          };
          user: {
            title: string;
            userName: string;
            userGender: string;
            nickName: string;
            userPhone: string;
            userEmail: string;
            userStatus: string;
            userRole: string;
            form: {
              userName: string;
              userGender: string;
              nickName: string;
              userPhone: string;
              userEmail: string;
              userStatus: string;
              userRole: string;
            };
            addUser: string;
            editUser: string;
            gender: {
              male: string;
              female: string;
            };
          };
          menu: {
            home: string;
            title: string;
            id: string;
            parentId: string;
            menuType: string;
            menuName: string;
            routeName: string;
            routePath: string;
            pathParam: string;
            layout: string;
            page: string;
            i18nKey: string;
            icon: string;
            localIcon: string;
            iconTypeTitle: string;
            order: string;
            constant: string;
            keepAlive: string;
            href: string;
            hideInMenu: string;
            activeMenu: string;
            multiTab: string;
            fixedIndexInTab: string;
            query: string;
            button: string;
            buttonCode: string;
            buttonDesc: string;
            menuStatus: string;
            form: {
              home: string;
              menuType: string;
              menuName: string;
              routeName: string;
              routePath: string;
              pathParam: string;
              layout: string;
              page: string;
              i18nKey: string;
              icon: string;
              localIcon: string;
              order: string;
              keepAlive: string;
              href: string;
              hideInMenu: string;
              activeMenu: string;
              multiTab: string;
              fixedInTab: string;
              fixedIndexInTab: string;
              queryKey: string;
              queryValue: string;
              button: string;
              buttonCode: string;
              buttonDesc: string;
              menuStatus: string;
            };
            addMenu: string;
            editMenu: string;
            addChildMenu: string;
            type: {
              directory: string;
              menu: string;
            };
            iconType: {
              iconify: string;
              local: string;
            };
          };
        };
        audit: {
          title: string;
          overview: string;
          statistics: string;
          loginLogs: string;
          operationLogs: string;
          systemEvents: string;
          username: string;
          nickname: string;
          module: string;
          action: string;
          level: string;
          source: string;
          category: string;
          message: string;
          details: string;
          location: string;
          status: string;
          success: string;
          failed: string;
          failReason: string;
          duration: string;
          sessionDuration: string;
          loginTime: string;
          logoutTime: string;
          operateTime: string;
          eventTime: string;
          export: string;
          search: string;
          reset: string;
          info: string;
          warning: string;
          error: string;
          critical: string;
        };
      };
      form: {
        required: string;
        userName: FormMsg;
        phone: FormMsg;
        pwd: FormMsg;
        confirmPwd: FormMsg;
        code: FormMsg;
        email: FormMsg;
      };
      dropdown: Record<Global.DropdownKey, string>;
      icon: {
        themeConfig: string;
        themeSchema: string;
        lang: string;
        fullscreen: string;
        fullscreenExit: string;
        reload: string;
        collapse: string;
        expand: string;
        pin: string;
        unpin: string;
      };
      datatable: {
        itemCount: string;
      };
    };

    type GetI18nKey<T extends Record<string, unknown>, K extends keyof T = keyof T> = K extends string
      ? T[K] extends Record<string, unknown>
        ? `${K}.${GetI18nKey<T[K]>}`
        : K
      : never;

    type I18nKey = GetI18nKey<Schema>;

    type TranslateOptions<Locales extends string> = import('vue-i18n').TranslateOptions<Locales>;

    interface $T {
      (key: I18nKey): string;
      (key: I18nKey, plural: number, options?: TranslateOptions<LangType>): string;
      (key: I18nKey, defaultMsg: string, options?: TranslateOptions<I18nKey>): string;
      (key: I18nKey, list: unknown[], options?: TranslateOptions<I18nKey>): string;
      (key: I18nKey, list: unknown[], plural: number): string;
      (key: I18nKey, list: unknown[], defaultMsg: string): string;
      (key: I18nKey, named: Record<string, unknown>, options?: TranslateOptions<LangType>): string;
      (key: I18nKey, named: Record<string, unknown>, plural: number): string;
      (key: I18nKey, named: Record<string, unknown>, defaultMsg: string): string;
    }
  }

  /** Service namespace */
  namespace Service {
    /** Other baseURL key */
    type OtherBaseURLKey = 'demo';

    interface ServiceConfigItem {
      /** The backend service base url */
      baseURL: string;
      /** The proxy pattern of the backend service base url */
      proxyPattern: string;
    }

    interface OtherServiceConfigItem extends ServiceConfigItem {
      key: OtherBaseURLKey;
    }

    /** The backend service config */
    interface ServiceConfig extends ServiceConfigItem {
      /** Other backend service config */
      other: OtherServiceConfigItem[];
    }

    interface SimpleServiceConfig extends Pick<ServiceConfigItem, 'baseURL'> {
      other: Record<OtherBaseURLKey, string>;
    }

    /** The backend service response data */
    type Response<T = unknown> = {
      /** The backend service response code */
      code: string;
      /** The backend service response message */
      msg: string;
      /** The backend service response data */
      data: T;
    };

    /** The demo backend service response data */
    type DemoResponse<T = unknown> = {
      /** The backend service response code */
      status: string;
      /** The backend service response message */
      message: string;
      /** The backend service response data */
      result: T;
    };
  }
}
