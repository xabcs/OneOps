/** The storage namespace */
declare namespace StorageType {
  interface Session {
    /** The theme color */
    themeColor: string;
    // /**
    //  * the theme settings
    //  */
    // themeSettings: App.Theme.ThemeSetting;
  }

  interface Local {
    /** The i18n language */
    lang: App.I18n.LangType;
    /** The token */
    token: string;
    /** Fixed sider with mix-menu */
    mixSiderFixed: CommonType.YesOrNo;
    /** The refresh token */
    refreshToken: string;
    /** The theme color */
    themeColor: string;
    /** The dark mode */
    darkMode: boolean;
    /** The theme settings */
    themeSettings: App.Theme.ThemeSetting;
    /** Whether the default content card border has been migrated */
    contentCardBorderMigrated: boolean;
    /**
     * The override theme flags
     *
     * The value is the build time of the project
     */
    overrideThemeFlag: string;
    /** The global tabs */
    globalTabs: App.Global.Tab[];
    /** The backup theme setting before is mobile */
    backupThemeSettingBeforeIsMobile: {
      layout: UnionKey.ThemeLayoutMode;
      siderCollapse: boolean;
    };

    /** The last login user id */
    lastLoginUserId: string;
    /**
     * The table column settings (visibility + order), keyed by page-level unique key
     *
     * Used by `columnSettingKey` of `useUITable` / `useUIPaginatedTable`
     */
    tableColumnSettings: Record<string, { prop: string; checked: boolean }[]>;
  }
}
