import { computed, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import { defineStore } from 'pinia';
import { useLoading } from '@sa/hooks';
import { fetchGetUserInfo, fetchLogin } from '@/service/api';
import { useRouterPush } from '@/hooks/common/router';
import { localStg } from '@/utils/storage';
import { SetupStoreId } from '@/enum';
import { $t } from '@/locales';
import { useRouteStore } from '../route';
import { useTabStore } from '../tab';
import { clearAuthStorage, getToken } from './shared';

export const useAuthStore = defineStore(SetupStoreId.Auth, () => {
  const route = useRoute();
  const authStore = useAuthStore();
  const routeStore = useRouteStore();
  const tabStore = useTabStore();
  const { toLogin, redirectFromLogin } = useRouterPush(false);
  const { loading: loginLoading, startLoading, endLoading } = useLoading();

  const token = ref(getToken());

  const userInfo: Api.Auth.UserInfo = reactive({
    id: 0,
    username: '',
    nickname: '',
    avatar: '',
    email: '',
    roleIds: [],
    status: 1,
    homePath: '/',
    createdAt: '',
    updatedAt: '',
    roleNames: [],
    menuTree: [],
    permissions: [],
    permissionInfo: []
  });

  /** is super role in static route */
  const isStaticSuper = computed(() => {
    const { VITE_AUTH_ROUTE_MODE, VITE_STATIC_SUPER_ROLE } = import.meta.env;

    return VITE_AUTH_ROUTE_MODE === 'static' && userInfo.roleNames.includes(VITE_STATIC_SUPER_ROLE);
  });

  /** Is login */
  const isLogin = computed(() => Boolean(token.value));

  /** User permissions */
  const permissions = computed(() => userInfo.permissions || []);

  /** Check if user is super admin */
  const isSuperAdmin = computed(() => {
    // 用户名是 admin 直接认定为超级管理员
    if (userInfo.username === 'admin') return true;
    // 或者拥有通配符权限
    return permissions.value.includes('*.*.*') || permissions.value.includes('admin');
  });

  /** Permission code to name mapping */

  /**
   * Get permission name by code
   * @param code Permission code (e.g., 'system.user.create')
   * @returns Permission name (e.g., '创建用户') or the code itself if not found
   */
  function getPermissionName(code: string): string {
		// 从登录返回的 permissionInfo 中查找权限名称
		if (userInfo.permissionInfo && userInfo.permissionInfo.length > 0) {
			const perm = userInfo.permissionInfo.find((p: any) => p.code === code)
			if (perm) return perm.name
		}
		// 回退到显示权限码
		return code
  }

  /**
   * Check if user has specific permission
   * @param permissionCode Permission code (e.g., 'system.user.create')
   */
  function hasPermission(permissionCode: string): boolean {
    if (!permissionCode) return true;
    if (isSuperAdmin.value) return true;
    return permissions.value.includes(permissionCode);
  }

  /**
   * Check if user has any of the specified permissions
   * @param permissionCodes Array of permission codes
   */
  function hasAnyPermission(permissionCodes: string[]): boolean {
    if (!permissionCodes || permissionCodes.length === 0) return true;
    if (isSuperAdmin.value) return true;
    return permissionCodes.some(code => hasPermission(code));
  }

  /**
   * Check if user has all of the specified permissions
   * @param permissionCodes Array of permission codes
   */
  function hasAllPermissions(permissionCodes: string[]): boolean {
    if (!permissionCodes || permissionCodes.length === 0) return true;
    if (isSuperAdmin.value) return true;
    return permissionCodes.every(code => hasPermission(code));
  }

  /** Reset auth store */
  async function resetStore() {
    recordUserId();

    clearAuthStorage();

    authStore.$reset();

    if (!route.meta.constant) {
      await toLogin();
    }

    tabStore.cacheTabs();
    routeStore.resetStore();
  }

  /** Record the user ID of the previous login session Used to compare with the current user ID on next login */
  function recordUserId() {
    if (!userInfo.id) {
      return;
    }

    // Store current user ID locally for next login comparison
    localStg.set('lastLoginUserId', String(userInfo.id));
  }

  /**
   * Check if current login user is different from previous login user If different, clear all tabs
   *
   * @returns {boolean} Whether to clear all tabs
   */
  function checkTabClear(): boolean {
    if (!userInfo.id) {
      return false;
    }

    const lastLoginUserId = localStg.get('lastLoginUserId');

    // Clear all tabs if current user is different from previous user
    if (lastLoginUserId !== String(userInfo.id)) {
      localStg.remove('globalTabs');
      tabStore.clearTabs();

      return true;
    }

    return false;
  }

  /**
   * Login
   *
   * @param userName User name
   * @param password Password
   * @param [redirect=true] Whether to redirect after login. Default is `true`
   */
  async function login(userName: string, password: string, redirect = true) {
    startLoading();

    const { data: loginToken, error } = await fetchLogin(userName, password);

    if (!error) {
      const pass = await loginByToken(loginToken);

      if (pass) {
        // Check if the tab needs to be cleared
        const isClear = checkTabClear();
        let needRedirect = redirect;

        if (isClear) {
          // If the tab needs to be cleared,it means we don't need to redirect.
          needRedirect = false;
        }
        await redirectFromLogin(needRedirect);

        window.$notification?.success({
          title: $t('page.login.common.loginSuccess'),
          message: $t('page.login.common.welcomeBack', { userName: userInfo.nickname || userInfo.username }),
          duration: 4500
        });
      }
    } else {
      resetStore();
    }

    endLoading();
  }

  async function loginByToken(loginToken: Api.Auth.LoginToken) {
    // 1. stored in the localStorage, the later requests need it in headers
    localStg.set('token', loginToken.token);

    // 2. get user info from login response
    const pass = await handleUserInfo(loginToken.user);

    if (pass) {
      token.value = loginToken.token;

      return true;
    }

    return false;
  }

  async function handleUserInfo(info: Api.Auth.UserInfo) {
    // update store - 需要深度更新以触发响应式
    Object.keys(info).forEach(key => {
      if (key === 'menuTree' || key === 'permissions' || key === 'permissionInfo') {
        // 对于数组类型，需要特殊处理
        (userInfo as any)[key] = info[key as keyof Api.Auth.UserInfo];
      } else {
        (userInfo as any)[key] = info[key as keyof Api.Auth.UserInfo];
      }
    });

    return true;
  }

  async function getUserInfo() {
    const { data: info, error } = await fetchGetUserInfo();

    if (!error) {
      // update store - 需要深度更新以触发响应式
      Object.keys(info).forEach(key => {
        if (key === 'menuTree' || key === 'permissions') {
          // 对于数组类型，需要特殊处理
          (userInfo as any)[key] = info[key as keyof Api.Auth.UserInfo];
        } else {
          (userInfo as any)[key] = info[key as keyof Api.Auth.UserInfo];
        }
      });

      return true;
    }

    return false;
  }

  async function initUserInfo() {
    const hasToken = getToken();

    if (hasToken) {
      const pass = await getUserInfo();

      if (!pass) {
        resetStore();
      }
    }
  }

  return {
    token,
    userInfo,
    isStaticSuper,
    isLogin,
    permissions,
    isSuperAdmin,
    loginLoading,
    resetStore,
    login,
    getUserInfo,
    initUserInfo,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    getPermissionName
  };
});
