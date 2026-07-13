import { computed, effectScope, onScopeDispose, ref, toRefs, watch } from 'vue';
import type { Ref } from 'vue';
import { usePreferredColorScheme } from '@vueuse/core';
import { defineStore } from 'pinia';
import { getPaletteColorByNumber } from '@sa/color';
import { localStg } from '@/utils/storage';
import { themeSettings } from '@/theme/settings';
import { SetupStoreId } from '@/enum';
import {
  addThemeVarsToGlobal,
  createThemeToken,
  getNaiveTheme,
  initThemeSettings,
  toggleAuxiliaryColorModes,
  toggleCssDarkMode
} from './shared';

/** Theme store */
export const useThemeStore = defineStore(SetupStoreId.Theme, () => {
  const scope = effectScope();
  const osTheme = usePreferredColorScheme();

  /** Theme settings */
  const settings: Ref<App.Theme.ThemeSetting> = ref(initThemeSettings());

  /** Dark mode */
  const darkMode = computed(() => {
    if (settings.value.themeScheme === 'auto') {
      return osTheme.value === 'dark';
    }
    return settings.value.themeScheme === 'dark';
  });

  /** grayscale mode */
  const grayscaleMode = computed(() => settings.value.grayscale);

  /** colourWeakness mode */
  const colourWeaknessMode = computed(() => settings.value.colourWeakness);

  /** Theme colors */
  const themeColors = computed(() => {
    const { themeColor, otherColor, isInfoFollowPrimary } = settings.value;
    const colors: App.Theme.ThemeColor = {
      primary: themeColor,
      ...otherColor,
      info: isInfoFollowPrimary ? themeColor : otherColor.info
    };
    return colors;
  });

  /** UI theme */
  const uiTheme = computed(() => {
    const borderRadius = settings.value.borderRadius.unified
      ? settings.value.borderRadius.global
      : settings.value.borderRadius.medium;
    return getNaiveTheme(themeColors.value, settings.value.recommendColor, borderRadius);
  });

  /**
   * Settings json
   *
   * It is for copy settings
   */
  const settingsJson = computed(() => JSON.stringify(settings.value));

  /** Reset store */
  function resetStore() {
    settings.value = themeSettings;
  }

  /**
   * Set theme scheme
   *
   * @param themeScheme
   */
  function setThemeScheme(themeScheme: UnionKey.ThemeScheme) {
    settings.value.themeScheme = themeScheme;
  }

  /**
   * Set grayscale value
   *
   * @param isGrayscale
   */
  function setGrayscale(isGrayscale: boolean) {
    settings.value.grayscale = isGrayscale;
  }

  /**
   * Set colourWeakness value
   *
   * @param isColourWeakness
   */
  function setColourWeakness(isColourWeakness: boolean) {
    settings.value.colourWeakness = isColourWeakness;
  }

  /** Toggle theme scheme */
  function toggleThemeScheme() {
    const themeSchemes: UnionKey.ThemeScheme[] = ['light', 'dark', 'auto'];

    const index = themeSchemes.findIndex(item => item === settings.value.themeScheme);

    const nextIndex = index === themeSchemes.length - 1 ? 0 : index + 1;

    const nextThemeScheme = themeSchemes[nextIndex];

    setThemeScheme(nextThemeScheme);
  }

  /**
   * Update theme colors
   *
   * @param key Theme color key
   * @param color Theme color
   */
  function updateThemeColors(key: App.Theme.ThemeColorKey, color: string) {
    let colorValue = color;

    if (settings.value.recommendColor) {
      // get a color palette by provided color and color name, and use the suitable color

      colorValue = getPaletteColorByNumber(color, 500, true);
    }

    if (key === 'primary') {
      settings.value.themeColor = colorValue;
    } else {
      settings.value.otherColor[key] = colorValue;
    }
  }

  /**
   * Set theme layout
   *
   * @param mode Theme layout mode
   */
  function setThemeLayout(mode: UnionKey.ThemeLayoutMode) {
    settings.value.layout.mode = mode;
  }

  /** Setup theme vars to global */
  function setupThemeVarsToGlobal() {
    const siderCustomColor = settings.value.sider.useCustomColor ? settings.value.sider.customColor : undefined;
    const { themeTokens, darkThemeTokens } = createThemeToken(
      themeColors.value,
      settings.value.tokens,
      settings.value.recommendColor,
      siderCustomColor
    );
    addThemeVarsToGlobal(themeTokens, darkThemeTokens, settings.value.borderRadius);
  }
  /**
   * Set layout reverse horizontal mix
   *
   * @param reverse Reverse horizontal mix
   */
  function setLayoutReverseHorizontalMix(reverse: boolean) {
    settings.value.layout.reverseHorizontalMix = reverse;
  }

  /**
   * Set border radius
   *
   * @param key Border radius key
   * @param value Border radius value
   */
  function setBorderRadius<K extends keyof App.Theme.ThemeSetting['borderRadius']>(
    key: K,
    value: App.Theme.ThemeSetting['borderRadius'][K]
  ) {
    (settings.value.borderRadius as any)[key] = value;
  }

  /**
   * Set component border radius
   *
   * @param component Component name
   * @param value Border radius value
   */
  function setComponentBorderRadius<K extends keyof App.Theme.ThemeSetting['borderRadius']['components']>(
    component: K,
    value: App.Theme.ThemeSetting['borderRadius']['components'][K]
  ) {
    settings.value.borderRadius.components[component] = value;
  }

  /**
   * Set sider custom color
   *
   * @param useCustomColor Whether to use custom color
   * @param color Custom color
   */
  function setSiderCustomColor(useCustomColor: boolean, color?: string) {
    settings.value.sider.useCustomColor = useCustomColor;
    if (color !== undefined) {
      settings.value.sider.customColor = color;
    }
  }

  /**
   * Set sider inverted
   *
   * @param inverted Inverted
   */
  function setSiderInverted(inverted: boolean) {
    settings.value.sider.inverted = inverted;
  }

  /**
   * Set sider show icon
   *
   * @param showIcon Show icon
   */
  function setSiderShowIcon(showIcon: boolean) {
    settings.value.sider.showIcon = showIcon;
  }

  /**
   * Set sider logo gradient
   *
   * @param useLogoGradient Use logo gradient
   * @param startColor Gradient start color
   * @param endColor Gradient end color
   */
  function setSiderLogoGradient(useLogoGradient: boolean, startColor?: string, endColor?: string) {
    settings.value.sider.useLogoGradient = useLogoGradient;
    if (startColor !== undefined) {
      settings.value.sider.logoGradientStart = startColor;
    }
    if (endColor !== undefined) {
      settings.value.sider.logoGradientEnd = endColor;
    }
  }

  /**
   * Set sider gradient
   *
   * @param useSiderGradient Use sider gradient
   * @param startColor Gradient start color
   * @param endColor Gradient end color
   */
  function setSiderGradient(useSiderGradient: boolean, startColor?: string, endColor?: string) {
    settings.value.sider.useSiderGradient = useSiderGradient;
    if (startColor !== undefined) {
      settings.value.sider.siderGradientStart = startColor;
    }
    if (endColor !== undefined) {
      settings.value.sider.siderGradientEnd = endColor;
    }
  }

  /** Cache theme settings */
  function cacheThemeSettings() {
    const isProd = import.meta.env.MODE === 'prod';

    if (!isProd) return;

    localStg.set('themeSettings', settings.value);
  }

  // cache theme settings when page is closed or refreshed
  // useEventListener(window, 'beforeunload', () => {
  //   cacheThemeSettings();
  // });

  // watch store
  scope.run(() => {
    // watch dark mode
    watch(
      darkMode,
      val => {
        toggleCssDarkMode(val);
        localStg.set('darkMode', val);
      },
      { immediate: true }
    );

    watch(
      [grayscaleMode, colourWeaknessMode],
      val => {
        toggleAuxiliaryColorModes(val[0], val[1]);
      },
      { immediate: true }
    );

    // themeColors change, update css vars and storage theme color
    watch(
      themeColors,
      val => {
        setupThemeVarsToGlobal();
        localStg.set('themeColor', val.primary);
      },
      { immediate: true }
    );

    // watch border radius change
    watch(
      () => settings.value.borderRadius,
      () => {
        setupThemeVarsToGlobal();
      },
      { deep: true }
    );

    // watch sider custom color change
    watch(
      () => [settings.value.sider.useCustomColor, settings.value.sider.customColor],
      () => {
        setupThemeVarsToGlobal();
      },
      { deep: true }
    );

    // watch sider gradient change
    watch(
      () => [
        settings.value.sider.useSiderGradient,
        settings.value.sider.siderGradientStart,
        settings.value.sider.siderGradientEnd,
        settings.value.sider.useLogoGradient,
        settings.value.sider.logoGradientStart,
        settings.value.sider.logoGradientEnd
      ],
      () => {
        // Gradient changes are handled directly by the component
        // No need to regenerate CSS vars
      },
      { deep: true }
    );

    // cache theme settings when settings change
    watch(
      settings,
      () => {
        cacheThemeSettings();
      },
      { deep: true }
    );
  });

  /** On scope dispose */
  onScopeDispose(() => {
    scope.stop();
  });

  return {
    ...toRefs(settings.value),
    darkMode,
    themeColors,
    uiTheme,
    settingsJson,
    setGrayscale,
    setColourWeakness,
    resetStore,
    setThemeScheme,
    toggleThemeScheme,
    updateThemeColors,
    setThemeLayout,
    setLayoutReverseHorizontalMix,
    setBorderRadius,
    setComponentBorderRadius,
    setSiderCustomColor,
    setSiderInverted,
    setSiderShowIcon,
    setSiderLogoGradient,
    setSiderGradient
  };
});
