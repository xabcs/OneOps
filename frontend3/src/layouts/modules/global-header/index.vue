<script setup lang="ts">
import { computed, ref } from 'vue';
import { useFullscreen } from '@vueuse/core';
import { GLOBAL_HEADER_MENU_ID } from '@/constants/app';
import { useAppStore } from '@/store/modules/app';
import { useThemeStore } from '@/store/modules/theme';
import GlobalLogo from '../global-logo/index.vue';
import GlobalBreadcrumb from '../global-breadcrumb/index.vue';
import GlobalSearch from '../global-search/index.vue';
import ThemeButton from './components/theme-button.vue';
import UserAvatar from './components/user-avatar.vue';

defineOptions({ name: 'GlobalHeader' });

interface Props {
  /** Whether to show the logo */
  showLogo?: App.Global.HeaderProps['showLogo'];
  /** Whether to show the menu toggler */
  showMenuToggler?: App.Global.HeaderProps['showMenuToggler'];
  /** Whether to show the menu */
  showMenu?: App.Global.HeaderProps['showMenu'];
}

defineProps<Props>();

const appStore = useAppStore();
const themeStore = useThemeStore();
const { isFullscreen, toggle } = useFullscreen();

// 计算 Header 的自定义背景样式
const headerStyle = computed(() => {
  // 如果使用渐变
  if (themeStore.header.useHeaderGradient) {
    return {
      background: `linear-gradient(180deg, ${themeStore.header.headerGradientStart || 'rgba(232, 237, 255, 0.98)'} 0%, ${themeStore.header.headerGradientEnd || 'rgba(232, 240, 251, 0.94)'} 100%)`
    };
  }

  // 如果使用自定义单一颜色
  if (themeStore.header.useCustomColor && themeStore.header.customColor) {
    return {
      backgroundColor: themeStore.header.customColor
    };
  }

  // 默认情况，使用主题默认背景
  return {};
});

// 系统状态
const currentTime = ref(new Date());
const systemStatus = ref({
  api: true, // API连接状态
  db: true, // 数据库连接状态
  agents: 15, // 在线Agent数量
  alerts: 3 // 未读告警数量
});

const systemInfo = computed(() => ({
  version: 'v1.0.0',
  environment: import.meta.env.MODE === 'production' ? '生产' : '开发'
}));

// 更新时间
setInterval(() => {
  currentTime.value = new Date();
}, 1000);

// 格式化时间
const formatTime = (date: Date) => {
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

// 快速链接
const quickLinks = [
  {
    name: 'GitHub',
    icon: 'mdi:github-face',
    tooltip: '项目源码',
    action: () => window.open('https://github.com/your-org/oneops', '_blank')
  },
  {
    name: '文档',
    icon: 'mdi:file-document-outline',
    tooltip: '项目文档',
    action: () => window.open('https://github.com/your-org/oneops-docs', '_blank')
  }
];

// 快速操作菜单
const quickActions = [
  {
    name: '刷新数据',
    icon: 'mdi:refresh',
    action: () => location.reload()
  },
  {
    name: '清除缓存',
    icon: 'mdi:cached',
    action: () => {
      localStorage.clear();
      sessionStorage.clear();
      location.reload();
    }
  },
  {
    name: '系统设置',
    icon: 'mdi:cog-outline',
    path: '/system'
  }
];

// 下拉菜单显示状态
const showQuickMenu = ref(false);

// 通知显示状态
const showNotifications = ref(false);

interface AppNotification {
  id: number;
  title: string;
  message: string;
  type: 'warning' | 'error' | 'info';
  time: string;
}

// 模拟通知数据
const notifications = computed<AppNotification[]>(() => [
  {
    id: 1,
    title: '服务器告警',
    message: 'server-01 CPU使用率超过80%',
    type: 'warning',
    time: '2分钟前'
  },
  {
    id: 2,
    title: '安全提醒',
    message: '发现3个待处理的安全漏洞',
    type: 'error',
    time: '15分钟前'
  },
  {
    id: 3,
    title: '系统通知',
    message: '定期维护计划于今晚23:00开始',
    type: 'info',
    time: '1小时前'
  }
]);

// 点击外部关闭下拉菜单
const handleClickOutside = () => {
  showQuickMenu.value = false;
  showNotifications.value = false;
};

// 处理通知点击
const handleNotificationClick = (_notification: AppNotification) => {
  showNotifications.value = false;
};
</script>

<template>
  <DarkModeContainer
    class="global-header relative h-full flex-y-center px-12px shadow-header"
    :style="headerStyle"
    @click="handleClickOutside"
  >
    <GlobalLogo v-if="showLogo" class="h-full" :style="{ width: themeStore.sider.width + 'px' }" />
    <MenuToggler v-if="showMenuToggler" :collapsed="appStore.siderCollapse" @click="appStore.toggleSiderCollapse" />
    <div v-if="showMenu" :id="GLOBAL_HEADER_MENU_ID" class="h-full flex-y-center flex-1-hidden"></div>
    <div v-else class="h-full flex-y-center flex-1-hidden">
      <GlobalBreadcrumb v-if="!appStore.isMobile" class="ml-12px" />
    </div>

    <!-- 左侧系统状态区域 -->
    <div v-if="!appStore.isMobile" class="mx-16px h-full flex-y-center flex-1-hidden">
      <div class="flex items-center gap-12px text-xs">
        <!-- 状态指示器组 -->
        <div class="flex items-center gap-8px">
          <!-- API状态 -->
          <div class="status-item">
            <div class="status-dot" :class="systemStatus.api ? 'status-success' : 'status-error'"></div>
            <span class="status-label">API</span>
          </div>

          <!-- 数据库状态 -->
          <div class="status-item">
            <div class="status-dot" :class="systemStatus.db ? 'status-success' : 'status-error'"></div>
            <span class="status-label">数据库</span>
          </div>

          <!-- Agent状态 -->
          <div class="status-item">
            <Icon icon="mdi:robot" class="status-icon" />
            <span class="status-label">{{ systemStatus.agents }} Agent</span>
          </div>
        </div>

        <div class="separator">|</div>

        <!-- 环境信息 -->
        <div class="flex items-center gap-4px text-gray-500 dark:text-gray-400">
          <span class="rounded bg-gray-100 px-8px py-2px dark:bg-gray-800">{{ systemInfo.environment }}</span>
          <span class="text-gray-400">{{ systemInfo.version }}</span>
        </div>
      </div>
    </div>

    <!-- 右侧操作区域 -->
    <div class="h-full flex-y-center justify-end gap-6px">
      <!-- 通知中心 -->
      <div v-if="!appStore.isMobile" class="header-action-item">
        <div class="relative" @click.stop="showNotifications = !showNotifications">
          <ButtonIcon icon="mdi:bell-outline" tooltip-content="通知中心" />
          <span v-if="systemStatus.alerts > 0" class="notification-badge">{{ systemStatus.alerts }}</span>
          <!-- 通知下拉面板 -->
          <div v-if="showNotifications" class="dropdown-panel notification-dropdown">
            <div class="dropdown-header">
              <span>通知中心</span>
              <span class="text-xs text-gray-500">({{ notifications.length }}条)</span>
            </div>
            <div class="dropdown-content">
              <div
                v-for="notif in notifications"
                :key="notif.id"
                class="notification-item"
                :class="'notification-' + notif.type"
                @click="handleNotificationClick(notif)"
              >
                <div class="notification-icon">
                  <Icon
                    :icon="
                      notif.type === 'warning'
                        ? 'mdi:alert'
                        : notif.type === 'error'
                          ? 'mdi:alert-circle'
                          : 'mdi:information'
                    "
                  />
                </div>
                <div class="notification-content">
                  <div class="notification-title">{{ notif.title }}</div>
                  <div class="notification-message">{{ notif.message }}</div>
                  <div class="notification-time">{{ notif.time }}</div>
                </div>
              </div>
              <div v-if="notifications.length === 0" class="py-12px text-center text-gray-500">暂无通知</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 快速操作菜单 -->
      <div v-if="!appStore.isMobile" class="header-action-item">
        <div class="relative" @click.stop="showQuickMenu = !showQuickMenu">
          <ButtonIcon icon="mdi:dots-horizontal" tooltip-content="快速操作" />
          <div v-if="showQuickMenu" class="dropdown-panel quick-menu-dropdown">
            <div v-for="action in quickActions" :key="action.name" class="dropdown-item" @click="action.action">
              <Icon :icon="action.icon" />
              <span>{{ action.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 快速链接按钮 -->
      <template v-if="!appStore.isMobile">
        <ButtonIcon
          v-for="link in quickLinks"
          :key="link.name"
          :icon="link.icon"
          :tooltip-content="link.tooltip"
          @click="link.action"
        />
      </template>

      <!-- 时间显示 -->
      <div v-if="!appStore.isMobile" class="time-display">
        <Icon icon="mdi:clock-outline" class="time-icon" />
        <span>{{ formatTime(currentTime) }}</span>
      </div>

      <GlobalSearch v-if="themeStore.header.globalSearch.visible" />
      <div>
        <FullScreen v-if="!appStore.isMobile" :full="isFullscreen" @click="toggle" />
      </div>
      <LangSwitch
        v-if="themeStore.header.multilingual.visible"
        :lang="appStore.locale"
        :lang-options="appStore.localeOptions"
        @change-lang="appStore.changeLocale"
      />
      <ThemeSchemaSwitch
        :theme-schema="themeStore.themeScheme"
        :is-dark="themeStore.darkMode"
        @switch="themeStore.toggleThemeScheme"
      />
      <div>
        <ThemeButton />
      </div>
      <UserAvatar />
    </div>
  </DarkModeContainer>
</template>

<style scoped>
.status-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  transition: all 0.2s;
}

.status-item:hover {
  background: var(--el-fill-color);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  transition: all 0.2s;
}

.status-success {
  background-color: #67c23a;
  box-shadow: 0 0 0 2px rgba(103, 194, 58, 0.3);
}

.status-error {
  background-color: #f56c6c;
  box-shadow: 0 0 0 2px rgba(245, 108, 108, 0.3);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.status-label {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.status-icon {
  font-size: 16px;
  color: var(--el-color-success);
}

.separator {
  color: var(--el-border-color);
  margin: 0 8px;
}

.header-action-item {
  position: relative;
}

.notification-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: var(--el-color-danger);
  color: white;
  font-size: 11px;
  font-weight: 600;
  line-height: 16px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(245, 108, 108, 0.3);
}

.dropdown-panel {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  min-width: 200px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  overflow: hidden;
}

.notification-dropdown {
  width: 320px;
}

.dropdown-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-light);
  font-weight: 600;
  font-size: 14px;
}

.dropdown-content {
  max-height: 400px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-light);
  cursor: pointer;
  transition: background 0.2s;
}

.notification-item:hover {
  background: var(--el-fill-color-light);
}

.notification-item:last-child {
  border-bottom: none;
}

.notification-icon {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
}

.notification-warning .notification-icon {
  background: rgba(230, 162, 60, 0.1);
  color: var(--el-color-warning);
}

.notification-error .notification-icon {
  background: rgba(245, 108, 108, 0.1);
  color: var(--el-color-danger);
}

.notification-info .notification-icon {
  background: rgba(64, 158, 255, 0.1);
  color: var(--el-color-primary);
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
}

.notification-message {
  font-size: 12px;
  color: var(--el-text-color-regular);
  line-height: 1.4;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-time {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.quick-menu-dropdown {
  min-width: 140px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  cursor: pointer;
  transition: all 0.2s;
}

.dropdown-item:hover {
  background: var(--el-fill-color-light);
  color: var(--el-color-primary);
}

.time-display {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 6px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-light);
}

.time-icon {
  font-size: 16px;
  color: var(--el-color-primary);
}
</style>
