import { createApp } from 'vue';
import './plugins/assets';
import './plugins/ui';
import { setupAppVersionNotification, setupDayjs, setupIconifyOffline, setupLoading, setupNProgress } from './plugins';
import { setupStore } from './store';
import { setupRouter } from './router';
import { setupI18n } from './locales';
import { initContentTheme, initContentTheme2, initHeaderTheme } from './utils/content-theme';
import App from './App.vue';
import permission from './directives/permission';
import PermissionButton from './components/PermissionButton.vue';

async function setupApp() {
  setupLoading();

  setupNProgress();

  setupIconifyOffline();

  setupDayjs();

  // 初始化内容主题和Header主题
  initContentTheme();
  initContentTheme2();
  initHeaderTheme();

  const app = createApp(App);

  setupStore(app);

  await setupRouter(app);

  setupI18n(app);

  // 注册全局权限指令
  app.directive('permission', permission);

  // 注册全局权限按钮组件
  app.component('PermissionButton', PermissionButton);

  setupAppVersionNotification();

  app.mount('#app');
}

setupApp();
